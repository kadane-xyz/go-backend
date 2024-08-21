package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

var middlewareHandler *Handler

// fix for Google
func ReferrerPolicyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func CustomLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		log.Printf("Started %s %s %s", r.Method, r.RequestURI, r.Proto)
		log.Printf("Headers: %v", r.Header)

		next.ServeHTTP(ww, r)

		log.Printf("Completed %v %s in %v", ww.Status(), http.StatusText(ww.Status()), time.Since(start))
		log.Printf("Response Headers: %v", ww.Header())
	})
}

func Middleware(m *Handler, r chi.Router) {
	middlewareHandler = m
	//CORS
	corsOptions := cors.Options{
		AllowedOrigins:   []string{"https://www.sandbox.paypal.com"}, // Use specific origins in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}

	// Middleware
	r.Use(cors.Handler(corsOptions))
	r.Use(middleware.AllowContentEncoding("deflate", "gzip")) // AllowContentEncoding middleware will allow the client to request compressed content
	r.Use(middleware.CleanPath)                               // CleanPath middleware will clean up the request URL path
	r.Use(middleware.Heartbeat("/api"))                       // Add a heartbeat endpoint
	r.Use(middleware.Recoverer)
	r.Use(middleware.Throttle(100)) // Throttle middleware will limit the number of requests per second
	//r.Use(CustomLogger)
	r.Use(middleware.Logger)
	r.Use(ReferrerPolicyMiddleware)
	//r.Use(JWTDecodeMiddleware) // JWTDecodeMiddleware will decode the JWT token and set the claims in the context
}
