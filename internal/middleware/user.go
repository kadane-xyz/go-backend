package middleware

import (
	"context"
	"net/http"

	"kadane.xyz/go-backend/v2/internal/errors"
)

func (h *Handler) UserAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawClaims := r.Context().Value(ClientTokenKey)
			claims, ok := rawClaims.(ClientContext)
			if !ok {
				errors.NewUnauthorizedError("Unauthorized")
				return
			}
			accountPlan, err := h.PostgresQueries.GetAccountPlan(r.Context(), claims.UserID)
			if err != nil {
				errors.NewForbiddenError("Forbidden")
				return
			}

			newClaims := claims
			newClaims.Plan = accountPlan

			ctx := context.WithValue(r.Context(), ClientTokenKey, newClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
