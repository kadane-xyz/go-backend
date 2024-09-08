package middleware

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Middleware(m *Handler, r chi.Router) {
	r.Use(middleware.AllowContentEncoding("deflate", "gzip")) // AllowContentEncoding middleware will allow the client to request compressed content
	r.Use(middleware.CleanPath)                               // CleanPath middleware will clean up the request URL path
	r.Use(middleware.Heartbeat("/api"))                       // Add a heartbeat endpoint
	r.Use(middleware.Recoverer)
	//r.Use(middleware.Throttle(100)) // Throttle middleware will limit the number of requests per second
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
}
