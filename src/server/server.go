package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/src/api"
	"kadane.xyz/go-backend/v2/src/cache"
	"kadane.xyz/go-backend/v2/src/config"
	"kadane.xyz/go-backend/v2/src/middleware"
)

var cacheInstance *cache.Cache // CacheInstance is the global cache instance for middleware/pages

type Server struct {
	config *config.Config
	//api clients
}

func NewServer(config *config.Config) *Server {
	//initialize api clients
	//ctx := context.Background()

	return &Server{
		config: config,
	}
}

func (s *Server) Run() error {
	// Initialize cache
	cacheInstance = cache.NewCache()

	//middleware handler
	middlewareHandler := &middleware.Handler{
		CacheInstance: cacheInstance,
	}

	//api handler
	ApiHandler := &api.Handler{
		CacheInstance: cacheInstance,
	}

	// HTTP router
	r := chi.NewRouter()

	// Middleware
	middleware.Middleware(middlewareHandler, r)

	// HTTP routes
	api.RegisterApiRoutes(ApiHandler, r) // Api routes

	// Start server
	log.Println("Starting server on :" + s.config.Port)
	if err := http.ListenAndServe(":"+s.config.Port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	return nil
}
