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
	ErrInvalidRequest   = errors.New("invalid request")
	ErrInternalServer   = errors.New("internal server error")
	ErrNotFound         = errors.New("not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrConflict         = errors.New("conflict")
	ErrMethodNotAllowed = errors.New("method not allowed")
)

type Error struct {
	Err        error
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

type ApiErrorData struct {
	Error Error `json:"error"`
}

// ApiError is the error response for the API
// { data: { error: { statusCode: 400, message: "Bad Request" } } }
type ApiError struct {
	Data ApiErrorData `json:"data"`
}

// NewError creates a new Error instance
func NewError(err error, statusCode int, message string) *Error {
	return &Error{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

// SendError sends an error response using the APIError structure
func SendError(w http.ResponseWriter, err error, statusCode int, message string) {
	error := ApiError{
		Data: ApiErrorData{
			Error: Error{
				Err:        err,
				Message:    message,
				StatusCode: statusCode,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(error)
}

func (e *ApiError) Send(w http.ResponseWriter) {
	SendError(w, e.Data.Error.Err, e.Data.Error.StatusCode, e.Data.Error.Message)
}

func NewApiError(err error, statusCode int, message string) *ApiError {
	return &ApiError{
		Data: ApiErrorData{
			Error: Error{
				Err:        err,
				Message:    message,
				StatusCode: statusCode,
			},
		},
	}
}

func NewInternalServerError(message string) *ApiError {
	return &ApiError{
		Data: ApiErrorData{
			Error: Error{
				Err:        ErrInternalServer,
				Message:    message,
				StatusCode: http.StatusInternalServerError,
			},
		},
	}
}

func NewNotFoundError(message string) *ApiError {
	return &ApiError{
		Data: ApiErrorData{
			Error: Error{
				Err:        ErrNotFound,
				Message:    message,
				StatusCode: http.StatusNotFound,
			},
		},
	}
}

func NewUnauthorizedError(message string) *ApiError {
	return &ApiError{
		Data: ApiErrorData{
			Error: Error{
				Err:        ErrUnauthorized,
				Message:    message,
				StatusCode: http.StatusUnauthorized,
			},
		},
	}
}

func NewForbiddenError(message string) *ApiError {
	return &ApiError{
		Data: ApiErrorData{
			Error: Error{
				Err:        ErrForbidden,
				Message:    message,
				StatusCode: http.StatusForbidden,
			},
		},
	}
}

func NewConflictError(message string) *ApiError {
	return &ApiError{
		Data: ApiErrorData{
			Error: Error{
				Err:        ErrConflict,
				Message:    message,
				StatusCode: http.StatusConflict,
			},
		},
	}
}

func NewMethodNotAllowedError(message string) *ApiError {
	return &ApiError{
		Data: ApiErrorData{
			Error: Error{
				Err:        ErrMethodNotAllowed,
				Message:    message,
				StatusCode: http.StatusMethodNotAllowed,
			},
		},
	}
}
