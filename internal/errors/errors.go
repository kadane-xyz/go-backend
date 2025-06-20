package errors

import (
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
	ErrPayloadTooLarge     = errors.New("payload too large")     // 413
	ErrUnprocessableEntity = errors.New("unprocessable entity")  // 422
	ErrInternalServer      = errors.New("internal server error") // 500
)

type AppError struct {
	Err        error  `json:"error"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

type ValidationError struct {
	AppError
}

type ApplicationError struct {
	AppError
}

type DatabaseError struct {
	AppError
}

// ApiError is the error response for the API
// { error: { statusCode: 400, message: "Bad Request" } } }
type ApiError struct {
	Err AppError `json:"error"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *ApiError) Error() string {
	return e.Err.Message
}

func (e *ApiError) Unwrap() error {
	return e.Err.Err
}

func NewAppError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

func NewApiError(err error, message string, statusCode int) *ApiError {
	return &ApiError{
		Err: AppError{
			Err:        err,
			Message:    message,
			StatusCode: statusCode,
		},
	}
}

// 400 Bad Request
func NewBadRequestError(message string) *ApiError {
	return &ApiError{
		Err: AppError{
			Message:    message,
			StatusCode: http.StatusBadRequest,
		},
	}
}

// 401 Unauthorized
func NewUnauthorizedError(message string) *ApiError {
	return &ApiError{
		Err: AppError{
			Message:    message,
			StatusCode: http.StatusUnauthorized,
		},
	}
}

// 403 Forbidden
func NewForbiddenError(message string) *ApiError {
	return &ApiError{
		Err: AppError{
			Message:    message,
			StatusCode: http.StatusForbidden,
		},
	}
}

// 404 Not Found
func NewNotFoundError(message string) *ApiError {
	return &ApiError{
		Err: AppError{
			Message:    message,
			StatusCode: http.StatusNotFound,
		},
	}
}

// 405 Method Not Allowed
func NewMethodNotAllowedError(message string) *ApiError {
	return &ApiError{
		Err: AppError{
			Message:    message,
			StatusCode: http.StatusMethodNotAllowed,
		},
	}
}

// 409 Conflict
func NewConflictError(message string) *ApiError {
	return &ApiError{
		Err: AppError{
			Message:    message,
			StatusCode: http.StatusConflict,
		},
	}
}

// 422 Unprocessable Entity
func NewUnprocessableEntityError(message string) *ApiError {
	return &ApiError{
		Err: AppError{
			Message:    message,
			StatusCode: http.StatusUnprocessableEntity,
		},
	}
}

// 500 Internal Server Error
func NewInternalServerError(message string) *ApiError {
	return &ApiError{
		Err: AppError{
			Message:    message,
			StatusCode: http.StatusInternalServerError,
		},
	}
}
