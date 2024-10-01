package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

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
		AllowedOrigins:   []string{"https://kadane.xyz", "https://www.kadane.xyz", "https://api.kadane.xyz", "http://localhost"}, // Define allowed origins (wildcard "*" can be used but it's not recommended for security reasons)
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},                                                    // HTTP methods that are allowed
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"}, // Headers that browsers can expose to frontend JavaScript
		AllowCredentials: true,             // Allow credentials (cookies, authentication) to be shared
		MaxAge:           300,              // Max age for preflight requests
	}))
	r.Use(middleware.Recoverer)
}
