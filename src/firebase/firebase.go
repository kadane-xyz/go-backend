package firebase

import (
	"context"
	"encoding/base64"
	"log"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"kadane.xyz/go-backend/v2/src/config"
)

func NewFirebaseApp(config *config.Config) (*firebase.App, error) {
	ctx := context.Background()
	// Load Firebase credentials from environment variable
	// decode the JSON data from base64
	cred, err := base64.StdEncoding.DecodeString(config.FirebaseCred)
	if err != nil {
		log.Printf("error decoding base64 encoded FIREBASE_CRED: %v\n", err)
		return nil, err
	}
	opt := option.WithCredentialsJSON(cred)

	// Initialize Firebase App
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Printf("error initializing app: %v\n", err)
		return nil, err
	}

	return app, nil
}
