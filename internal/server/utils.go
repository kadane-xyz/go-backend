package server

import (
	"encoding/json"
	"net/http"
)

// helper functions

// DecodeJSONRequest is a generic function to validate and decode HTTP request bodies
func DecodeJSONRequest[T any](r *http.Request) (T, *server.APIError) {
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		var empty T
		return empty, NewError(http.StatusBadRequest, "Invalid request body")
	}
	return request, nil
}
