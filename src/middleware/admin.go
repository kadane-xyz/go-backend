package middleware

import (
	"context"
	"net/http"
	"strings"

	"kadane.xyz/go-backend/v2/src/apierror"
)

// AdminAuth is a middleware that enforces admin authentication for routes that include /admin after the versioning prefix.
func (h *Handler) AdminAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Unconditionally trim the version prefix "/v1" from the URL path.
			path := strings.TrimPrefix(r.URL.Path, "/v1")

			if !strings.HasPrefix(path, "/admin") {
				next.ServeHTTP(w, r)
				return
			}

			// Optionally allow certain admin endpoints (e.g., /admin/validate) to bypass the auth check.
			if path == "/admin/validate" {
				next.ServeHTTP(w, r)
				return
			}

			// Retrieve the client token from the context.
			claimsRaw := r.Context().Value(ClientTokenKey)

			claims, ok := claimsRaw.(ClientContext)
			if !ok {
				apierror.SendError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			// Validate the admin using your database query.
			adminCheck, err := h.PostgresQueries.ValidateAdmin(r.Context(), claims.UserID)
			if err != nil {
				apierror.SendError(w, http.StatusForbidden, "Forbidden")
				return
			}
			if !adminCheck.Bool {
				apierror.SendError(w, http.StatusForbidden, "Forbidden")
				return
			}

			// Update the claims with the admin flag.
			newClaims := claims
			newClaims.Admin = true

			// Add the updated claims to the context.
			ctx := context.WithValue(r.Context(), ClientTokenKey, newClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
