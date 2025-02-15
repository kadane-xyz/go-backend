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
		apierror.SendError(w, http.StatusBadRequest, "Missing client user ID context")
		return "", errors.New("missing client user ID context")
	}
	return value, nil
}

func GetClientPlan(w http.ResponseWriter, r *http.Request) (sql.AccountPlan, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Plan
	if value == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing client plan context")
		return "", errors.New("missing client plan context")
	}
	return value, nil
}

func GetClientAdmin(w http.ResponseWriter, r *http.Request) (bool, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Admin
	if !value {
		apierror.SendError(w, http.StatusBadRequest, "Missing client admin context")
		return false, errors.New("missing client admin context")
	}
	return value, nil
}

func GetClientEmail(w http.ResponseWriter, r *http.Request) (string, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Email
	if value == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing client email context")
		return "", errors.New("missing client email context")
	}
	return value, nil
}

func GetClientName(w http.ResponseWriter, r *http.Request) (string, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).Name
	if value == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing client name context")
		return "", errors.New("missing client name context")
	}
	return value, nil
}

func GetClientFullContext(w http.ResponseWriter, r *http.Request) (middleware.ClientContext, error) {
	value := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext)
	if value == (middleware.ClientContext{}) {
		apierror.SendError(w, http.StatusBadRequest, "Missing full client context")
		return middleware.ClientContext{}, errors.New("missing full client context")
	}
	return value, nil
}
