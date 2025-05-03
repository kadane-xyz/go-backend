package middleware

import (
	"errors"
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

		// Handle application errors
		var appErr *appError.AppError
		if errors.As(err, &appErr) {
			log.Errorf("application error: %v", appErr.Err, appErr.StatusCode)
			httputils.SendJSONResponse(w, appErr.StatusCode, appErr.Message)
			return
		}

		// Handle unknown errors
		log.Errorf("unknown error: %v", err)
		httputils.SendJSONResponse(w, http.StatusInternalServerError, "unexpected error")
	})
}
