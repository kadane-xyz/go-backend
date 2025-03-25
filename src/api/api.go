package api

import (
	"encoding/json"
	"net/http"

	"kadane.xyz/go-backend/v2/src/apierror"
)

// helper functions

// DecodeJSONRequest is a generic function to validate and decode HTTP request bodies
func DecodeJSONRequest[T any](r *http.Request) (T, *apierror.APIError) {
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		var empty T
		return empty, apierror.NewError(http.StatusBadRequest, "Invalid request body")
	}
	return request, nil
}

func SendJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
