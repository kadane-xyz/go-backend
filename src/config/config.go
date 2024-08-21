package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
}

func LoadConfig() (*Config, error) {
	log.Println("Loading configuration")
	// Attempt to load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Fetch environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Return the configuration by fetching environment variables
	config := &Config{
		Port: port,
	}

	LoadKeys() // Load Ed25519 keys

	log.Println("Configuration loaded")

	return config, nil
}
