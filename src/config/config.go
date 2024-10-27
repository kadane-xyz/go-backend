package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Postgres
	PostgresUrl  string
	PostgresUser string
	PostgresPass string
	PostgresDB   string
	// Server
	Port string
	// Firebase
	FirebaseCred string
	// AWS
	AWSKey          string
	AWSSecret       string
	AWSBucketAvatar string
	AWSRegion       string
	// Judge0
	Judge0Url   string
	Judge0Token string
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

	awsKey := os.Getenv("AWS_KEY")
	if awsKey == "" {
		return nil, fmt.Errorf("AWS_KEY is not set")
	}

	awsSecret := os.Getenv("AWS_SECRET")
	if awsSecret == "" {
		return nil, fmt.Errorf("AWS_SECRET is not set")
	}

	awsBucketAvatar := os.Getenv("AWS_BUCKET_AVATAR")
	if awsBucketAvatar == "" {
		return nil, fmt.Errorf("AWS_BUCKET_AVATAR is not set")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		return nil, fmt.Errorf("AWS_REGION is not set")
	}

	judge0Url := os.Getenv("JUDGE0_URL")
	if judge0Url == "" {
		return nil, fmt.Errorf("JUDGE0_URL is not set")
	}

	judge0Token := os.Getenv("JUDGE0_TOKEN")
	if judge0Token == "" {
		return nil, fmt.Errorf("JUDGE0_TOKEN is not set")
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
		//AWS
		AWSKey:          awsKey,
		AWSSecret:       awsSecret,
		AWSBucketAvatar: awsBucketAvatar,
		AWSRegion:       awsRegion,
		//Judge0
		Judge0Url:   judge0Url,
		Judge0Token: judge0Token,
	}

	log.Println("Configuration loaded")

	return config, nil
}
