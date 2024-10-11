package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	//Postgres
	PostgresUrl  string
	PostgresUser string
	PostgresPass string
	PostgresDB   string
	Port         string
	FirebaseCred string
}

// Fetch environment variables
func LoadConfig() (*Config, error) {
	log.Println("Loading configuration")

	// load .env after fetching environment variables
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	postgresUrl := os.Getenv("POSTGRES_URL")
	if postgresUrl == "" {
		return nil, fmt.Errorf("POSTGRES_URL is not set")
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		return nil, fmt.Errorf("POSTGRES_USER is not set")
	}

	postgresPass := os.Getenv("POSTGRES_PASSWORD")
	if postgresPass == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD is not set")
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		return nil, fmt.Errorf("POSTGRES_DB is not set")
	}

	firebaseCred := os.Getenv("FIREBASE_CRED")
	if firebaseCred == "" {
		return nil, fmt.Errorf("FIREBASE_CRED is not set")
	}

	// Return the configuration by fetching environment variables
	config := &Config{
		//Postgres
		PostgresUrl:  postgresUrl,
		PostgresUser: postgresUser,
		PostgresPass: postgresPass,
		PostgresDB:   postgresDB,
		//Firebase
		FirebaseCred: firebaseCred,
		//Server
		Port: port,
	}

	log.Println("Configuration loaded")

	return config, nil
}
