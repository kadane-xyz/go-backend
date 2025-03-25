package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"kadane.xyz/go-backend/v2/src/apierror"
)

func (h *Handler) FirebaseAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check endpoint
			if r.URL.Path == "/health" || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Get the Firebase ID token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("Missing Authorization header: %s\n", r.URL.Path)
				apierror.SendError(w, http.StatusUnauthorized, "Missing Authorization header")
				return
			}

			const bearerPrefix = "Bearer "
			if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
				log.Printf("Invalid Authorization header: %s\n", r.URL.Path)
				apierror.SendError(w, http.StatusUnauthorized, "Invalid Authorization header")
				return
			}

			// Extract the ID token from the Authorization header
			idToken := authHeader[len(bearerPrefix):]
			if idToken == "" {
				log.Printf("Missing ID token: %s\n", r.URL.Path)
				apierror.SendError(w, http.StatusUnauthorized, "Missing ID token")
				return
			}

			client, err := h.FirebaseApp.Auth(r.Context())
			if err != nil {
				log.Fatalf("error getting Auth client: %v\n", err)
			}

			// Verify the ID token
			token, err := client.VerifyIDToken(r.Context(), idToken)
			if err != nil {
				log.Println("Error verifying ID token: ", err)
				SendError(w, http.StatusUnauthorized, "Invalid Firebase ID token")
				return
			}
			claims := ClientContext{
				UserID: token.UID,
				Email:  getStringClaim(token.Claims, "email"),
				Name:   getStringClaim(token.Claims, "name"),
			}

			// Pass the claims to the next handler via the context
			ctx := context.WithValue(r.Context(), ClientTokenKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *Handler) FirebaseDebugAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check endpoint
			if r.URL.Path == "/health" || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			token := "123abc"

			claims := ClientContext{
				UserID: token,
				Email:  "john@example.com",
				Name:   "John Doe",
			}

			// Pass the claims to the next handler via the context
			ctx := context.WithValue(r.Context(), ClientTokenKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CustomLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		middleware.Logger(next).ServeHTTP(w, r)
	})
}
