package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func writeJson(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, errorValue string, message string) {
	writeJson(w, statusCode, ErrorResponse{
		Error:   errorValue,
		Message: message,
	})
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func readIdentityHeaders(r *http.Request) (role string, patientId string, hospitalId string, ok bool) {
	role = strings.TrimSpace(r.Header.Get("X-Role"))
	if role == "" {
		return "", "", "", false
	}

	if role == "PATIENT" {
		patientId = strings.TrimSpace(r.Header.Get("X-Patient-Id"))
		if patientId == "" {
			return "", "", "", false
		}
		return role, patientId, "", true
	}

	if role == "STAFF" {
		hospitalId = strings.TrimSpace(r.Header.Get("X-Hospital-Id"))
		if hospitalId == "" {
			return "", "", "", false
		}
		return role, "", hospitalId, true
	}

	return "", "", "", false
}
