package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type HttpHandler struct {
	service *DocumentService
}

func NewHttpHandler(service *DocumentService) http.Handler {
	handler := &HttpHandler{service: service}

	mux := http.NewServeMux()
	mux.HandleFunc("/documents", handler.handleDocumentsRoot)
	mux.HandleFunc("/documents/", handler.handleDocumentsById)
	mux.HandleFunc("/patients/", handler.handlePatientsNamespace)

	return mux
}

func (h *HttpHandler) handleDocumentsRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/documents" {
		h.createDocument(w, r)
		return
	}
	writeError(w, http.StatusNotFound, "Not Found", "Route not found.")
}

func (h *HttpHandler) handlePatientsNamespace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusNotFound, "Not Found", "Route not found.")
		return
	}

	pathParts := splitPath(r.URL.Path)
	if len(pathParts) != 3 || pathParts[0] != "patients" || pathParts[2] != "documents" {
		writeError(w, http.StatusNotFound, "Not Found", "Route not found.")
		return
	}

	patientId := pathParts[1]
	h.listDocumentsForPatient(w, r, patientId)
}

func (h *HttpHandler) handleDocumentsById(w http.ResponseWriter, r *http.Request) {
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 2 || pathParts[0] != "documents" {
		writeError(w, http.StatusNotFound, "Not Found", "Route not found.")
		return
	}

	documentId := pathParts[1]

	if r.Method == http.MethodGet && len(pathParts) == 2 {
		h.getDocument(w, r, documentId)
		return
	}

	if r.Method == http.MethodPatch && len(pathParts) == 3 && pathParts[2] == "review" {
		h.reviewDocument(w, r, documentId)
		return
	}

	if r.Method == http.MethodPost && len(pathParts) == 5 && pathParts[2] == "links" && pathParts[3] == "medical-records" {
		medicalRecordId := pathParts[4]
		h.linkDocumentToMedicalRecord(w, r, documentId, medicalRecordId)
		return
	}

	writeError(w, http.StatusNotFound, "Not Found", "Route not found.")
}

func (h *HttpHandler) createDocument(w http.ResponseWriter, r *http.Request) {
	role, patientId, _, ok := readIdentityHeaders(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized", "Missing required identity headers.")
		return
	}
	if role != "PATIENT" {
		writeError(w, http.StatusForbidden, "Forbidden", "Only PATIENT can create documents.")
		return
	}

	var req CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body.")
		return
	}

	if strings.TrimSpace(req.HospitalId) == "" {
		writeError(w, http.StatusBadRequest, "Bad Request", "hospitalId is required.")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "Bad Request", "title is required.")
		return
	}
	if strings.TrimSpace(req.Category) == "" {
		writeError(w, http.StatusBadRequest, "Bad Request", "category is required.")
		return
	}

	doc, err := h.service.CreateDocument(CreateDocumentInput{
		PatientId:  patientId,
		HospitalId: req.HospitalId,
		Title:      req.Title,
		Category:   req.Category,
		Notes:      req.Notes,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error", "Could not create document.")
		return
	}

	writeJson(w, http.StatusCreated, doc)
}

func (h *HttpHandler) listDocumentsForPatient(w http.ResponseWriter, r *http.Request, patientId string) {
	role, callerPatientId, staffHospitalId, ok := readIdentityHeaders(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized", "Missing required identity headers.")
		return
	}
	if role == "PATIENT" && callerPatientId != patientId {
		writeError(w, http.StatusForbidden, "Forbidden", "Patient can only list their own documents.")
		return
	}

	filter := ListDocumentsFilter{
		StatusFilter: strings.TrimSpace(r.URL.Query().Get("status")),
	}
	if role == "STAFF" {
		filter.HospitalId = staffHospitalId
	}

	docs := h.service.ListDocumentsForPatient(patientId, filter)
	writeJson(w, http.StatusOK, docs)
}

func (h *HttpHandler) getDocument(w http.ResponseWriter, r *http.Request, documentId string) {
	role, callerPatientId, staffHospitalId, ok := readIdentityHeaders(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized", "Missing required identity headers.")
		return
	}

	doc, err := h.service.GetDocument(documentId)
	if errors.Is(err, ErrDocumentNotFound) {
		writeError(w, http.StatusNotFound, "Not Found", "Document not found.")
		return
	}

	if role == "PATIENT" && doc.PatientId != callerPatientId {
		writeError(w, http.StatusForbidden, "Forbidden", "Patient cannot access another patient's document.")
		return
	}
	if role == "STAFF" && doc.HospitalId != staffHospitalId {
		writeError(w, http.StatusForbidden, "Forbidden", "Staff cannot access documents outside their hospital.")
		return
	}

	writeJson(w, http.StatusOK, doc)
}

func (h *HttpHandler) reviewDocument(w http.ResponseWriter, r *http.Request, documentId string) {
	role, _, staffHospitalId, ok := readIdentityHeaders(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized", "Missing required identity headers.")
		return
	}
	if role != "STAFF" {
		writeError(w, http.StatusForbidden, "Forbidden", "Only STAFF can review documents.")
		return
	}

	doc, err := h.service.GetDocument(documentId)
	if errors.Is(err, ErrDocumentNotFound) {
		writeError(w, http.StatusNotFound, "Not Found", "Document not found.")
		return
	}
	if doc.HospitalId != staffHospitalId {
		writeError(w, http.StatusForbidden, "Forbidden", "Staff cannot review documents outside their hospital.")
		return
	}

	var req ReviewDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body.")
		return
	}

	doc, err = h.service.ReviewDocument(documentId, ReviewDocumentInput{
		Status:     req.Status,
		ReviewNote: req.ReviewNote,
	})
	if errors.Is(err, ErrInvalidStatus) {
		writeError(w, http.StatusBadRequest, "Bad Request", "status must be APPROVED or REJECTED.")
		return
	}

	writeJson(w, http.StatusOK, doc)
}

func (h *HttpHandler) linkDocumentToMedicalRecord(w http.ResponseWriter, r *http.Request, documentId string, medicalRecordId string) {
	role, _, staffHospitalId, ok := readIdentityHeaders(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized", "Missing required identity headers.")
		return
	}
	if role != "STAFF" {
		writeError(w, http.StatusForbidden, "Forbidden", "Only STAFF can link documents.")
		return
	}

	doc, err := h.service.GetDocument(documentId)
	if errors.Is(err, ErrDocumentNotFound) {
		writeError(w, http.StatusNotFound, "Not Found", "Document not found.")
		return
	}
	if doc.HospitalId != staffHospitalId {
		writeError(w, http.StatusForbidden, "Forbidden", "Staff cannot link documents outside their hospital.")
		return
	}

	doc, _ = h.service.LinkDocumentToMedicalRecord(documentId, medicalRecordId)
	writeJson(w, http.StatusOK, doc)
}
