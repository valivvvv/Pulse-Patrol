package main

import (
	"errors"
	"strings"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrInvalidStatus    = errors.New("status must be APPROVED or REJECTED")
)

type DocumentService struct {
	store DocumentStore
}

func NewDocumentService(store DocumentStore) *DocumentService {
	return &DocumentService{store: store}
}

type CreateDocumentInput struct {
	PatientId  string
	HospitalId string
	Title      string
	Category   string
	Notes      string
}

func (s *DocumentService) CreateDocument(input CreateDocumentInput) (Document, error) {
	now := nowIso8601Utc()
	document := Document{
		DocumentId:             "",
		PatientId:              input.PatientId,
		HospitalId:             input.HospitalId,
		Title:                  input.Title,
		Category:               input.Category,
		Notes:                  input.Notes,
		Status:                 DocumentStatusPendingReview,
		ReviewNote:             "",
		LinkedMedicalRecordIds: []string{},
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	return s.store.CreateDocument(document)
}

func (s *DocumentService) GetDocument(documentId string) (Document, error) {
	document, found := s.store.GetDocumentById(documentId)
	if !found {
		return Document{}, ErrDocumentNotFound
	}
	return document, nil
}

type ListDocumentsFilter struct {
	HospitalId   string // If non-empty, only include documents from this hospital
	StatusFilter string // If non-empty, only include documents with this status
}

func (s *DocumentService) ListDocumentsForPatient(patientId string, filter ListDocumentsFilter) []Document {
	allDocuments := s.store.ListDocumentsByPatientId(patientId)

	var filtered []Document
	for _, doc := range allDocuments {
		if filter.HospitalId != "" && doc.HospitalId != filter.HospitalId {
			continue
		}
		if filter.StatusFilter != "" && string(doc.Status) != filter.StatusFilter {
			continue
		}
		filtered = append(filtered, doc)
	}

	return filtered
}

type ReviewDocumentInput struct {
	Status     string
	ReviewNote string
}

func (s *DocumentService) ReviewDocument(documentId string, input ReviewDocumentInput) (Document, error) {
	document, found := s.store.GetDocumentById(documentId)
	if !found {
		return Document{}, ErrDocumentNotFound
	}

	newStatus := strings.TrimSpace(input.Status)
	if newStatus != string(DocumentStatusApproved) && newStatus != string(DocumentStatusRejected) {
		return Document{}, ErrInvalidStatus
	}

	document.Status = DocumentStatus(newStatus)
	document.ReviewNote = input.ReviewNote
	document.UpdatedAt = nowIso8601Utc()

	_ = s.store.UpdateDocument(document)
	return document, nil
}

func (s *DocumentService) LinkDocumentToMedicalRecord(documentId string, medicalRecordId string) (Document, error) {
	document, found := s.store.GetDocumentById(documentId)
	if !found {
		return Document{}, ErrDocumentNotFound
	}

	if !containsString(document.LinkedMedicalRecordIds, medicalRecordId) {
		document.LinkedMedicalRecordIds = append(document.LinkedMedicalRecordIds, medicalRecordId)
		document.UpdatedAt = nowIso8601Utc()
		_ = s.store.UpdateDocument(document)
	}

	return document, nil
}
