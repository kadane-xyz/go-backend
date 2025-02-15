package middleware

import (
	"context"
	"net/http"
	"strings"

	"kadane.xyz/go-backend/v2/src/apierror"
)

func (h *Handler) AdminAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/admin") {
				next.ServeHTTP(w, r)
				return
			}
			claims := r.Context().Value(ClientTokenKey).(ClientContext)
			adminCheck, err := h.PostgresQueries.ValidateAdmin(r.Context(), claims.UserID)
			if err != nil {
				apierror.SendError(w, http.StatusForbidden, "Forbidden")
				return
			}
			if !adminCheck.Bool {
				apierror.SendError(w, http.StatusForbidden, "Forbidden")
				return
			}

			claims.Admin = true // Set the admin flag

			ctx := context.WithValue(r.Context(), ClientTokenKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
