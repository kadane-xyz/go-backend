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

type CreateAccountRequest struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AccountUpdateRequest struct {
	Bio          *string `json:"bio,omitempty"`
	ContactEmail *string `json:"contactEmail,omitempty"`
	Location     *string `json:"location,omitempty"`
	RealName     *string `json:"realName,omitempty"`
	GithubUrl    *string `json:"githubUrl,omitempty"`
	LinkedinUrl  *string `json:"linkedinUrl,omitempty"`
	FacebookUrl  *string `json:"facebookUrl,omitempty"`
	InstagramUrl *string `json:"instagramUrl,omitempty"`
	TwitterUrl   *string `json:"twitterUrl,omitempty"`
	School       *string `json:"school,omitempty"`
	WebsiteUrl   *string `json:"websiteUrl,omitempty"`
}
