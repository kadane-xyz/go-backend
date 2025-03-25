package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kadane.xyz/go-backend/v2/internal/config"
	"kadane.xyz/go-backend/v2/internal/container"
	"kadane.xyz/go-backend/v2/internal/server"
)

func main() {
	// Create the root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load the application configuration to be able to initialize Sentry as soon as possible
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}

	// Set up signal handling for graceful shutdowns
	signalCtx, stopSignalHandler := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignalHandler()

	// Create the service container
	container, err := container.NewContainer(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to create the service container: %v", err)
	}
	defer func() {
		container.Close()
	}()

	// Create the server using the service container
	httpServer := server.New(container)
	if err != nil {
		log.Fatal("Failed to create the server: %v", err)
	}

	// Start the server with the signal-aware context
	err = httpServer.ListenAndServe(signalCtx)

	// Determine if this was a clean shutdown or an error
	if err != nil && err != context.Canceled {
		log.Fatal("Server error: %v", err)
	}

	log.Println("Server has shut down")
}
