package main

import (
	"log"

	"kadane.xyz/go-backend/v2/src/config"
	"kadane.xyz/go-backend/v2/src/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv := server.NewServer(cfg)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
