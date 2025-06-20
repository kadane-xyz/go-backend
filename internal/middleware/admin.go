package middleware

import (
	"context"
	"net/http"
	"strings"

	"kadane.xyz/go-backend/v2/internal/errors"
)

// AdminAuth is a middleware that enforces admin authentication for routes that include /admin after the versioning prefix.
func (h *Handler) AdminAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Retrieve the client token from the context.
			claimsRaw := r.Context().Value(ClientTokenKey)

			claims, ok := claimsRaw.(ClientContext)
			if !ok {
				errors.NewUnauthorizedError("Unauthorized")
				return
			}

			// Check user admin status
			isAdmin, err := h.PostgresQueries.ValidateAdmin(r.Context(), claims.UserID)
			if err != nil {
				errors.NewForbiddenError("Forbidden")
				return
			}

			// Update the claims with the admin status.
			newClaims := claims
			newClaims.Admin = isAdmin

			// Add the updated claims to the context.
			ctx := context.WithValue(r.Context(), ClientTokenKey, newClaims)

			// Unconditionally trim the version prefix "/v1" from the URL path.
			path := strings.TrimPrefix(r.URL.Path, "/v1")

			if !strings.HasPrefix(path, "/admin") {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Optionally allow certain admin endpoints (e.g., /admin/validate) to bypass the auth check.
			if path == "/admin/validate" {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
