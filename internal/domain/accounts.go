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

type AccountAttributesUpdateParams struct {
	ID           string
	Bio          *string
	ContactEmail *string
	Location     *string
	RealName     *string
	GithubUrl    *string
	LinkedinUrl  *string
	FacebookUrl  *string
	InstagramUrl *string
	TwitterUrl   *string
	School       *string
	WebsiteUrl   *string
}

type AccountValidation struct {
	Plan sql.AccountPlan `json:"plan"`
}

type AccountCreateRequest struct {
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

// Use for calling repo
type AccountUpdateParams struct {
	ID string
	AccountUpdateRequest
}

func FromSQLAccountRow(row sql.GetAccountRow) Account {
	account := Account{
		ID:         row.ID,
		Username:   row.Username,
		Email:      row.Email,
		Level:      row.Level,
		CreatedAt:  row.CreatedAt.Time,
		Plan:       row.Plan,
		Attributes: row.Attributes.(AccountAttributes),
	}
	if row.AvatarUrl != nil {
		account.AvatarUrl = *row.AvatarUrl
	}

	return account
}

func FromSQLAccountByUsernameRow(row sql.GetAccountByUsernameRow) Account {
	account := Account{
		ID:         row.ID,
		Username:   row.Username,
		Email:      row.Email,
		Level:      row.Level,
		CreatedAt:  row.CreatedAt.Time,
		Plan:       row.Plan,
		Attributes: row.Attributes.(AccountAttributes),
	}
	if row.AvatarUrl != nil {
		account.AvatarUrl = *row.AvatarUrl
	}

	return account
}

func FromSQLListAccountsRow(rows []sql.ListAccountsRow) []Account {
	accounts := make([]Account, len(rows))
	for i, row := range rows {
		accounts[i] = Account{
			ID:         row.ID,
			Username:   row.Username,
			Email:      row.Email,
			Level:      row.Level,
			CreatedAt:  row.CreatedAt.Time,
			Plan:       row.Plan,
			Attributes: row.Attributes.(AccountAttributes),
		}
		if row.AvatarUrl != nil {
			accounts[i].AvatarUrl = *row.AvatarUrl
		}
	}
	return accounts
}

func FromSQLGetAccountAttributes(row sql.AccountAttribute) *AccountAttributes {
	accountAttributes := AccountAttributes{
		ID: row.ID,
	}

	if row.Bio != nil {
		accountAttributes.Bio = *row.Bio
	}

	if row.ContactEmail != nil {
		accountAttributes.ContactEmail = *row.ContactEmail
	}

	if row.Location != nil {
		accountAttributes.Location = *row.Location
	}

	if row.RealName != nil {
		accountAttributes.RealName = *row.RealName
	}

	if row.GithubUrl != nil {
		accountAttributes.GithubUrl = *row.GithubUrl
	}

	if row.LinkedinUrl != nil {
		accountAttributes.LinkedinUrl = *row.LinkedinUrl
	}

	if row.FacebookUrl != nil {
		accountAttributes.FacebookUrl = *row.FacebookUrl
	}

	if row.InstagramUrl != nil {
		accountAttributes.InstagramUrl = *row.InstagramUrl
	}

	if row.TwitterUrl != nil {
		accountAttributes.TwitterUrl = *row.TwitterUrl
	}

	if row.School != nil {
		accountAttributes.School = *row.School
	}

	if row.WebsiteUrl != nil {
		accountAttributes.WebsiteUrl = *row.WebsiteUrl
	}

	return &accountAttributes
}
