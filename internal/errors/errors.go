package errors

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Types of errors
// Application errors
// Validation errors
// Database errors
// External service errors
// Internal server errors

var (
	ErrInvalidRequest      = errors.New("invalid request")       // 400
	ErrUnauthorized        = errors.New("unauthorized")          // 401
	ErrForbidden           = errors.New("forbidden")             // 403
	ErrNotFound            = errors.New("not found")             // 404
	ErrMethodNotAllowed    = errors.New("method not allowed")    // 405
	ErrConflict            = errors.New("conflict")              // 409
	ErrUnprocessableEntity = errors.New("unprocessable entity")  // 422
	ErrInternalServer      = errors.New("internal server error") // 500
)

type Error struct {
	Error      error
	Message    string
	StatusCode int
}

type ApiErrorData struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

// ApiError is the error response for the API
// { error: { statusCode: 400, message: "Bad Request" } } }
type ApiError struct {
	Error ApiErrorData `json:"error"`
}

// NewError creates a new Error instance
func NewError(err error, statusCode int, message string) *Error {
	return &Error{
		Error:      err,
		Message:    message,
		StatusCode: statusCode,
	}
}

// SendError sends an error response using the APIError structure
func SendError(w http.ResponseWriter, statusCode int, message string) {
	error := ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: statusCode,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(error)
}

func (e *ApiError) Send(w http.ResponseWriter) {
	SendError(w, e.Error.StatusCode, e.Error.Message)
}

func NewApiError(err error, statusCode int, message string) *ApiError {
	return &ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: statusCode,
		},
	}
}

// 400 Bad Request
func NewBadRequestError(message string) *ApiError {
	return &ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: http.StatusBadRequest,
		},
	}
}

// 401 Unauthorized
func NewUnauthorizedError(message string) *ApiError {
	return &ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: http.StatusUnauthorized,
		},
	}
}

// 403 Forbidden
func NewForbiddenError(message string) *ApiError {
	return &ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: http.StatusForbidden,
		},
	}
}

// 404 Not Found
func NewNotFoundError(message string) *ApiError {
	return &ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: http.StatusNotFound,
		},
	}
}

// 405 Method Not Allowed
func NewMethodNotAllowedError(message string) *ApiError {
	return &ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: http.StatusMethodNotAllowed,
		},
	}
}

// 409 Conflict
func NewConflictError(message string) *ApiError {
	return &ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: http.StatusConflict,
		},
	}
}

// 422 Unprocessable Entity
func NewUnprocessableEntityError(message string) *ApiError {
	return &ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: http.StatusUnprocessableEntity,
		},
	}
}

// 500 Internal Server Error
func NewInternalServerError(message string) *ApiError {
	return &ApiError{
		Error: ApiErrorData{
			Message:    message,
			StatusCode: http.StatusInternalServerError,
		},
	}
}
