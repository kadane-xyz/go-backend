package apierror

import (
	"encoding/json"
	"net/http"
)

// ErrorDetails contains the detailed error information
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

// NewError creates a new APIError instance
func NewError(statusCode int, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// SendError sends an error response using the APIError structure
func SendError(w http.ResponseWriter, statusCode int, message string) {
	response := APIError{
		StatusCode: statusCode,
		Message:    message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
