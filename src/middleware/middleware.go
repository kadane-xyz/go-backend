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
	"kadane.xyz/go-backend/v2/src/apierror"
)

type ContextKey string

const FirebaseTokenKey ContextKey = "firebaseToken"

type FirebaseTokenInfo struct {
	UserID string
	Email  string
	Name   string
}

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
			apierror.SendError(w, http.StatusMethodNotAllowed, "CONNECT method is not allowed")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getStringClaim(claims map[string]interface{}, key string) string {
	if val, ok := claims[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return "" // or handle missing/invalid claims as appropriate
}

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
				apierror.SendError(w, http.StatusUnauthorized, "Invalid Firebase ID token")
				return
			}
			claims := FirebaseTokenInfo{
				UserID: token.UID,
				Email:  getStringClaim(token.Claims, "email"),
				Name:   getStringClaim(token.Claims, "name"),
			}

			// Pass the claims to the next handler via the context
			ctx := context.WithValue(r.Context(), FirebaseTokenKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *Handler) DebugAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check endpoint
			if r.URL.Path == "/health" || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			token := "123abc"

			claims := FirebaseTokenInfo{
				UserID: token,
				Email:  "john@example.com",
				Name:   "John Doe",
			}

			// Pass the claims to the next handler via the context
			ctx := context.WithValue(r.Context(), FirebaseTokenKey, claims)
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

func Middleware(m *Handler, r chi.Router) {
	r.Use(middleware.RealIP)
	//r.Use(routeValidator)
	r.Use(httprate.LimitByIP(10, 1*time.Second)) // LimitByIP middleware will limit the number of requests per IP address
	r.Use(middleware.Heartbeat("/health"))       // Heartbeat middleware will create a simple health check endpoint
	r.Use(CustomLogger)
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
	// DEBUG bypass firebase auth
	if m.Config.Debug {
		r.Use(m.DebugAuth())
	} else {
		r.Use(m.FirebaseAuth()) // Firebase Auth middleware
	}

	log.Println("Middleware initialized")
}
