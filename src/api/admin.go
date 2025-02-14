package api

import (
	"encoding/json"
	"net/http"

	"kadane.xyz/go-backend/v2/src/apierror"
)

type AdminValidationResponse struct {
	Data bool `json:"data"`
}

// GET: /admin/validate
func (h *Handler) GetAdminValidation(w http.ResponseWriter, r *http.Request) {
	admin, err := GetClientAdmin(w, r)
	if err != nil {
		return
	}

	if !admin {
		apierror.SendError(w, http.StatusForbidden, "You are not authorized as admin")
		return
	}

	response := AdminValidationResponse{
		Data: admin,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
