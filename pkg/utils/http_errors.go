package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// HTTPError represents a standardized error response for HTTP APIs
type HTTPError struct {
	Status  int         `json:"status"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// WriteErrorResponse writes a standardized error response to the HTTP response writer
func WriteErrorResponse(w http.ResponseWriter, status int, code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errResponse := HTTPError{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	}
	if err := json.NewEncoder(w).Encode(errResponse); err != nil {
		logrus.WithError(err).Error("Failed to write error response")
	}
}

// Common HTTP error codes
const (
	A1ErrorUnauthorized = "A1_UNAUTHORIZED"
	A1ErrorInvalidInput = "A1_INVALID_INPUT"
	A1ErrorNotFound     = "A1_NOT_FOUND"
	A1ErrorInternal     = "A1_INTERNAL_ERROR"
)
