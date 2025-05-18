package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvProduction  Environment = "prod"
	EnvStaging     Environment = "stage"
	EnvDevelopment Environment = "dev"
	EnvTest        Environment = "test"
)

type FirebaseConfig struct {
	Cred string
}

type AWSConfig struct {
	Key           string
	Secret        string
	BucketAvatar  string
	Region        string
	CloudFrontURL string
}

type Judge0Config struct {
	Url   string
	Token string
}

type Config struct {
	// DEBUG
	Debug bool
	// Server
	Port        string
	Environment Environment
	// Database
	DatabaseURL string
	// Firebase
	Firebase FirebaseConfig
	// AWS
	AWS AWSConfig
	// Judge0
	Judge0 Judge0Config
}

// IsValid reports whether e is one of the known Environments.
func (e Environment) IsValid() bool {
	switch e {
	case EnvProduction, EnvStaging, EnvDevelopment, EnvTest:
		return true
	}
	return false
}

// ParseEnvironment turns a raw string into an Environment, or returns an error.
func ParseEnvironment(s string) (Environment, error) {
	e := Environment(s)
	if !e.IsValid() {
		return "", fmt.Errorf("%q is not a valid environment (allowed: prod, stage, dev, test)", s)
	}
	return e, nil
}

// LoadEnvironment reads ENV from the OS, parses it, and returns a valid Environment.
func LoadEnvironment() (Environment, error) {
	raw := os.Getenv("ENV")
	if raw == "" {
		return "", fmt.Errorf("ENV is not set")
	}
	return ParseEnvironment(raw)
}

// Fetch environment variables
func LoadConfig() (*Config, error) {
	log.Println("Loading configuration")

	// load .env after fetching environment variables
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	environment, err := LoadEnvironment()
	if err != nil {
		return nil, err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	debugStr := os.Getenv("DEBUG")
	debug := debugStr == "true"
	if debug {
		log.Println("Debug mode enabled")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
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

	cloudFrontUrl := os.Getenv("CLOUD_FRONT_URL")
	if cloudFrontUrl == "" {
		return nil, fmt.Errorf("CLOUD_FRONT_URL is not set")
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
		Debug:       debug,
		Environment: environment,
		//Database
		DatabaseURL: databaseURL,
		//Firebase
		Firebase: FirebaseConfig{
			Cred: firebaseCred,
		},
		//Server
		Port: port,
		//AWS
		AWS: AWSConfig{
			Key:           awsKey,
			Secret:        awsSecret,
			BucketAvatar:  awsBucketAvatar,
			Region:        awsRegion,
			CloudFrontURL: cloudFrontUrl,
		},
		//Judge0
		Judge0: Judge0Config{
			Url:   judge0Url,
			Token: judge0Token,
		},
	}

	log.Println("Configuration loaded")

	return config, nil
}
