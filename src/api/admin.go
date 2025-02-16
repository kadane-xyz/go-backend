package api

import (
	"encoding/json"
	"net/http"
)

type AdminValidation struct {
	IsAdmin bool `json:"isAdmin"`
}

type AdminValidationResponse struct {
	Data AdminValidation `json:"data"`
}

// GET: /admin/validate
func (h *Handler) GetAdminValidation(w http.ResponseWriter, r *http.Request) {
	admin := GetClientAdmin(w, r)

	response := AdminValidationResponse{
		Data: AdminValidation{
			IsAdmin: admin,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
