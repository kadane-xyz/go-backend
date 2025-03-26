package httputils

import (
	"encoding/json"
	"net/http"

	"kadane.xyz/go-backend/v2/internal/errors"
)

type ApiResponse struct {
	Data any `json:"data"`
}

func DecodeJSONRequest[T any](r *http.Request) (T, *errors.ApiError) {
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		var empty T
		return empty, errors.NewApiError(err, http.StatusBadRequest, "Invalid request body")
	}
	return request, nil
}

// { data: {} }
func EmptyDataResponse(w http.ResponseWriter) {
	response := map[string]any{
		"data": map[string]any{},
	}
	json.NewEncoder(w).Encode(response)
}

// { data: [] }
func EmptyDataArrayResponse(w http.ResponseWriter) {
	response := map[string]any{
		"data": []any{},
	}
	json.NewEncoder(w).Encode(response)
}

// Send a JSON response with the given status code and data
func SendJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Send a JSON response with the given status code and data
// { data: {} }
func SendJSONDataResponse(w http.ResponseWriter, statusCode int, data any) {
	SendJSONResponse(w, statusCode, ApiResponse{
		Data: data,
	})
}
