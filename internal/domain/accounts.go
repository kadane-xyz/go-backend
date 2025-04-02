package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Account struct {
	ID           string
	Username     string
	Email        string
	AvatarUrl    string
	Level        int32
	CreatedAt    time.Time
	FriendStatus sql.FriendshipStatus
	Plan         sql.AccountPlan
	IsAdmin      bool
	Attributes   AccountAttributes
}

type AccountAttributes struct {
	ID                 string
	Bio                string
	ContactEmail       string
	Location           string
	RealName           string
	GithubUrl          string
	LinkedinUrl        string
	FacebookUrl        string
	InstagramUrl       string
	TwitterUrl         string
	School             string
	WebsiteUrl         string
	FriendCount        int64
	BlockedCount       int64
	FriendRequestCount int64
}

type AccountValidation struct {
	Plan sql.AccountPlan `json:"plan"`
}
