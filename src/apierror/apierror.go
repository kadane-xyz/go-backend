package apierror

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents the outer structure of the error response
type ErrorResponse struct {
	Error *ErrorDetails `json:"error"`
}

// ErrorDetails contains the detailed error information
type ErrorDetails struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

// NewError creates a new ErrorDetails instance
func NewError(statusCode int, message string) *ErrorDetails {
	return &ErrorDetails{
		StatusCode: statusCode,
		Message:    message,
	}
}

// Error method implements the error interface
func (e *ErrorDetails) Error() string {
	return e.Message
}

// WriteJSON writes the error as JSON to an http.ResponseWriter
func WriteJSON(w http.ResponseWriter, err *ErrorDetails) {
	response := ErrorResponse{Error: err}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(response)
}

// SendError sends an error response using the ErrorResponse structure
func SendError(w http.ResponseWriter, statusCode int, message string) {
	err := NewError(statusCode, message)
	WriteJSON(w, err)
}
