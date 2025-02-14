package api

import (
	"errors"
	"net/http"

	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

func GetClientUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).UserID
	if value == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID context")
		return "", errors.New("missing user ID context")
	}
	return value, nil
}

func GetClientType(w http.ResponseWriter, r *http.Request) (sql.AccountType, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Type
	if value == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing account type context")
		return "", errors.New("missing account type context")
	}
	return value, nil
}

func GetClientAdmin(w http.ResponseWriter, r *http.Request) (bool, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Admin
	if !value {
		apierror.SendError(w, http.StatusBadRequest, "Missing admin context")
		return false, errors.New("missing admin context")
	}
	return value, nil
}

func GetClientEmail(w http.ResponseWriter, r *http.Request) (string, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Email
	if value == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing email context")
		return "", errors.New("missing email context")
	}
	return value, nil
}

func GetClientName(w http.ResponseWriter, r *http.Request) (string, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Name
	if value == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing name context")
		return "", errors.New("missing name context")
	}
	return value, nil
}

func GetClientFullContext(w http.ResponseWriter, r *http.Request) (middleware.ClientContext, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext)
	if value == (middleware.ClientContext{}) {
		apierror.SendError(w, http.StatusBadRequest, "Missing full context")
		return middleware.ClientContext{}, errors.New("missing full context")
	}
	return value, nil
}
