package api

import (
	"encoding/json"
	"net/http"
)

// { data: {} }
func EmptyDataResponse(w http.ResponseWriter) {
	response := map[string]interface{}{
		"data": map[string]interface{}{},
	}
	json.NewEncoder(w).Encode(response)
}

// { data: [] }
func EmptyDataArrayResponse(w http.ResponseWriter) {
	response := map[string]interface{}{
		"data": []interface{}{},
	}
	json.NewEncoder(w).Encode(response)
}

func SendJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
