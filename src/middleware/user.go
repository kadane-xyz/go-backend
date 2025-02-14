package middleware

import (
	"context"
	"net/http"

	"kadane.xyz/go-backend/v2/src/apierror"
)

func (h *Handler) UserAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value(ClientTokenKey).(ClientContext)
			accountType, err := h.PostgresQueries.GetAccountType(r.Context(), claims.UserID)
			if err != nil {
				apierror.SendError(w, http.StatusForbidden, "Forbidden")
				return
			}

			claims.Type = accountType.AccountType // Set the account type

			ctx := context.WithValue(r.Context(), ClientTokenKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
