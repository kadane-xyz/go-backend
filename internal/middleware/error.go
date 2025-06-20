package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/martian/v3/log"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	appError "kadane.xyz/go-backend/v2/internal/errors"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func ErrorMiddleware(next HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)
		if err == nil {
			return
		}

		fmt.Println("test")

		// Handle application errors
		var appErr *appError.AppError
		if errors.As(err, &appErr) {
			log.Errorf("application error: %v", appErr.Err, appErr.StatusCode)
			httputils.SendJSONResponse(w, appErr.StatusCode, appErr.Message)
			return
		}

		// Handle api errors
		var apiErr *appError.ApiError
		if errors.As(err, &apiErr) {
			log.Errorf("api error: %v", apiErr.Err.Err, apiErr.Err.StatusCode)
			httputils.SendJSONResponse(w, apiErr.Err.StatusCode, apiErr.Err.Message)
			return
		}

		// Handle database errors
		var dbErr *appError.DatabaseError
		if errors.As(err, &dbErr) {
			log.Errorf("database error: %w", dbErr.Err, dbErr.StatusCode)
			httputils.SendJSONResponse(w, dbErr.StatusCode, dbErr.Message)
			return
		}

		// Handle validation errors
		var validErr *appError.ValidationError
		if errors.As(err, &validErr) {
			log.Errorf("validation error: %v", validErr.Err, validErr.StatusCode)
			httputils.SendJSONResponse(w, validErr.StatusCode, validErr.Message)
			return
		}

		fmt.Println("error", err)

		// Handle unknown errors
		log.Errorf("unknown error: %v", err)
		httputils.SendJSONResponse(w, http.StatusInternalServerError, "unexpected error")
	})
}
