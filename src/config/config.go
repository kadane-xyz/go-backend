package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
}

// Fetch environment variables
func LoadConfig() (*Config, error) {
	log.Println("Loading configuration")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// load .env after fetching environment variables
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalln("Error loading .env")
	}

	// Return the configuration by fetching environment variables
	config := &Config{
		Port: port,
	}

	log.Println("Configuration loaded")

	return config, nil
}
