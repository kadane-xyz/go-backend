package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

type ContextKey string

const FirebaseTokenKey ContextKey = "firebaseToken"

// BlockConnectMethod blocks any request using the CONNECT method
func BlockConnectMethod(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			// Optionally log the blocked request with minimal info to reduce CPU usage
			log.Printf("[BLOCKED] Method: %s - Path: %s - IP: %s",
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
			)
			http.Error(w, "CONNECT method is not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) FirebaseAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Firebase ID token from the Authorization header
			idToken := r.Header.Get("Authorization")

			if idToken == "" {
				http.Error(w, "Authorization header with Firebase ID token is required", http.StatusUnauthorized)
				return
			}

			client, err := h.FirebaseApp.Auth(r.Context())
			if err != nil {
				log.Fatalf("error getting Auth client: %v\n", err)
			}

			// Verify the ID token
			token, err := client.VerifyIDToken(r.Context(), idToken)
			if err != nil {
				http.Error(w, "Invalid Firebase ID token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), FirebaseTokenKey, token)

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Middleware(m *Handler, r chi.Router) {
	r.Use(middleware.RealIP)
	//r.Use(routeValidator)
	r.Use(httprate.LimitByIP(10, 1*time.Second)) // LimitByIP middleware will limit the number of requests per IP address
	r.Use(middleware.Logger)
	r.Use(BlockConnectMethod)                                 // BlockConnectMethod middleware will block any request using the CONNECT method
	r.Use(middleware.AllowContentEncoding("deflate", "gzip")) // AllowContentEncoding middleware will allow the client to request compressed content
	r.Use(middleware.NoCache)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.CleanPath) // CleanPath middleware will clean up the request URL path
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://kadane.xyz", "https://www.kadane.xyz", "https://api.kadane.xyz", "http://localhost:5173"}, // Define allowed origins (wildcard "*" can be used but it's not recommended for security reasons)
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},                                                // HTTP methods that are allowed
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"}, // Headers that browsers can expose to frontend JavaScript
		AllowCredentials: true,             // Allow credentials (cookies, authentication) to be shared
		MaxAge:           300,              // Max age for preflight requests
	}))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health")) // Heartbeat middleware will create a simple health check endpoint
	r.Use(m.FirebaseAuth())                // Firebase Auth middleware
}
