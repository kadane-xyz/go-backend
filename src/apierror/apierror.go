package apierror

import (
	"encoding/json"
	"net/http"
)

// { data: { error: { statusCode: 400, message: "Bad Request" } } }

// ErrorDetails contains the detailed error information
type APIErrorData struct {
	Error struct {
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
	} `json:"error"`
}

type APIErrorError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type APIError struct {
	Data APIErrorData `json:"data"`
}

// NewError creates a new APIError instance
func NewError(statusCode int, message string) *APIError {
	return &APIError{
		Data: APIErrorData{
			Error: APIErrorError{
				StatusCode: statusCode,
				Message:    message,
			},
		},
	}
}

// SendError sends an error response using the APIError structure
func SendError(w http.ResponseWriter, statusCode int, message string) {
	response := APIError{
		Data: APIErrorData{
			Error: APIErrorError{
				StatusCode: statusCode,
				Message:    message,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (e *APIError) StatusCode() int {
	return e.Data.Error.StatusCode
}

func (e *APIError) Message() string {
	return e.Data.Error.Message
}
