package config

import (
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
}

// Fetch environment variables
func LoadConfig() (*Config, error) {
	log.Println("Loading configuration")

	// load .env after fetching environment variables
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalln("Error loading .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	postgresUrl := os.Getenv("POSTGRES_URL")
	if postgresUrl == "" {
		log.Fatalln("POSTGRES_URL is not set")
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		log.Fatalln("POSTGRES_USER is not set")
	}

	postgresPass := os.Getenv("POSTGRES_PASSWORD")
	if postgresPass == "" {
		log.Fatalln("POSTGRES_PASS is not set")
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		log.Fatalln("POSTGRES_DB is not set")
	}

	// Return the configuration by fetching environment variables
	config := &Config{
		//Postgres
		PostgresUrl:  postgresUrl,
		PostgresUser: postgresUser,
		PostgresPass: postgresPass,
		PostgresDB:   postgresDB,
		//Server
		Port: port,
	}

	log.Println("Configuration loaded")

	return config, nil
}
