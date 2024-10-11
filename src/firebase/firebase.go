package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"kadane.xyz/go-backend/v2/src/config"
)

func NewFirebaseApp(config *config.Config) (*firebase.App, error) {
	ctx := context.Background()
	// Load Firebase credentials from environment variable
	cred := []byte(config.FirebaseCred)
	opt := option.WithCredentialsJSON(cred)

	// Initialize Firebase App
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
		return nil, err
	}

	return app, nil
}
