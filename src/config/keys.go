// Desc: Load Ed25519 keys from files
package config

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

var (
	Ed25519PrivateKey ed25519.PrivateKey
	Ed25519PublicKey  ed25519.PublicKey
)

func LoadKeys() {
	log.Println("Loading signing keys")
	//private key
	privateKeyData, err := os.ReadFile("ed25519-private.pem")
	if err != nil {
		log.Fatalf("Failed to read private key: %v", err)
	}
	block, _ := pem.Decode(privateKeyData)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Fatalf("Failed to decode private key")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes) //parse private key
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	ed25519PKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		log.Fatalf("Invalid private key")
	}

	Ed25519PrivateKey = ed25519PKey

	//public key
	publicKeyData, err := os.ReadFile("ed25519-public.pem")
	if err != nil {
		log.Fatalf("Failed to read public key: %v", err)
	}
	block, _ = pem.Decode(publicKeyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Fatalf("Failed to decode public key")
	}
	key, err = x509.ParsePKIXPublicKey(block.Bytes) //parse public key
	if err != nil {
		log.Fatalf("Failed to parse public key: %v", err)
	}
	ed25519PubKey, ok := key.(ed25519.PublicKey)
	if !ok {
		log.Fatalf("Invalid public key")
	}

	Ed25519PublicKey = ed25519PubKey

	log.Println("Signing keys loaded")
}
