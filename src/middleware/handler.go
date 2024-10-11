package middleware

import (
	firebase "firebase.google.com/go/v4"
	"kadane.xyz/go-backend/v2/src/config"
)

type Handler struct {
	Config      *config.Config
	FirebaseApp *firebase.App
}
