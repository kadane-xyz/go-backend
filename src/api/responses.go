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
