package server

import (
	"net/http"

	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

func GetClientUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).UserID
	if value == "" {
		errors.NewNotFoundError("Missing client user ID context")
	}
	return value, nil
}

func GetClientPlan(w http.ResponseWriter, r *http.Request) (sql.AccountPlan, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Plan
	if value == "" {
		errors.NewNotFoundError("Missing client plan context")
	}
	return value, nil
}

func GetClientAdmin(w http.ResponseWriter, r *http.Request) bool {
	return r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Admin
}

func GetClientEmail(w http.ResponseWriter, r *http.Request) (string, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Email
	if value == "" {
		errors.NewNotFoundError("Missing client email context")
	}
	return value, nil
}

func GetClientName(w http.ResponseWriter, r *http.Request) (string, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Name
	if value == "" {
		errors.NewNotFoundError("Missing client name context")
	}
	return value, nil
}

func GetClientFullContext(w http.ResponseWriter, r *http.Request) (middleware.ClientContext, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext)
	if value == (middleware.ClientContext{}) {
		errors.NewNotFoundError("Missing full client context")
	}
	return value, nil
}
