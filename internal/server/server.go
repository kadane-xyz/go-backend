package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"kadane.xyz/go-backend/v2/internal/container"
)

type Server struct {
	container *container.Container
	mux       *chi.Mux
	srv       *http.Server
}

func New(container *container.Container) *Server {
	server := &Server{
		container: container,
		mux:       chi.NewRouter(),
	}

	h2s := &http2.Server{}
	mux := http.NewServeMux()
	server.RegisterApiRoutes()

	handler := h2c.NewHandler(mux, h2s)

	server.srv = &http.Server{
		Addr:    ":" + server.container.Config.Port,
		Handler: handler,
	}

	return server
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	address := ":" + s.container.Config.Port

	s.srv = &http.Server{
		Addr:         address,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to capture server errors
	serverError := make(chan error, 1)

	// Start the server in a goroutine so it's non-blocking
	go func() {
		log.Println("Starting server on " + address)

		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverError <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for the context to be cancelled or an error to occur
	select {
	case <-ctx.Done():
		log.Println("Server is shutting down gracefully...")
		return s.gracefulShutdown()

	case err := <-serverError:
		log.Println("Server error: " + err.Error())

		// Try to gracefully shut the server down
		s.gracefulShutdown()

		return err
	}
}

// ServeHTTP implements the http.Handler interface for testing
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// gracefulShutdown attempts to shut down the server gracefully with a timeout
func (s *Server) gracefulShutdown() error {
	// Create a timeout context for the shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)

		// Force close if graceful shutdown fails
		if err := s.srv.Close(); err != nil {
			log.Printf("Server close error: %v", err)
		}

		return fmt.Errorf("server failed to shutdown cleanly: %w", err)
	}

	log.Println("Server shut down successfully")
	return nil
}
