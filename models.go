package main

import "time"

type DocumentStatus string

const (
	DocumentStatusPendingReview DocumentStatus = "PENDING_REVIEW"
	DocumentStatusApproved      DocumentStatus = "APPROVED"
	DocumentStatusRejected      DocumentStatus = "REJECTED"
)

type Document struct {
	DocumentId            string         `json:"documentId"`
	PatientId             string         `json:"patientId"`
	HospitalId            string         `json:"hospitalId"`
	Title                 string         `json:"title"`
	Category              string         `json:"category"`
	Notes                 string         `json:"notes"`
	Status                DocumentStatus `json:"status"`
	ReviewNote            string         `json:"reviewNote"`
	LinkedMedicalRecordIds []string      `json:"linkedMedicalRecordIds"`
	CreatedAt             string         `json:"createdAt"`
	UpdatedAt             string         `json:"updatedAt"`
}

func nowIso8601Utc() string {
	return time.Now().UTC().Format(time.RFC3339)
}

type CreateDocumentRequest struct {
	HospitalId string `json:"hospitalId"`
	Title      string `json:"title"`
	Category   string `json:"category"`
	Notes      string `json:"notes"`
}

type ReviewDocumentRequest struct {
	Status     string `json:"status"`
	ReviewNote string `json:"reviewNote"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
