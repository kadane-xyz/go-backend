package server

import (
	"encoding/json"
	"net/http"

	"kadane.xyz/go-backend/v2/internal/errors"
)

// helper functions

// DecodeJSONRequest is a generic function to validate and decode HTTP request bodies
func DecodeJSONRequest[T any](r *http.Request) (T, *errors.ApiError) {
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		var empty T
		return empty, errors.NewApiError(err, http.StatusBadRequest, "Invalid request body")
	}
	return request, nil
}
