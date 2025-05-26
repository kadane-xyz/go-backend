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

type AccountAvatarParams struct {
	AvatarUrl string
	ID        string
}

type AccountAttributesCreateParams = sql.CreateAccountAttributesParams

type AccountGetParams struct {
	ID                string            `json:"id"`
	IncludeAttributes bool              `json:"includeAttributes"`
	UsernamesFilter   []string          `json:"usernamesFilter"`
	LocationsFilter   []string          `json:"locationsFilter"`
	Sort              string            `json:"sort"`
	SortDirection     sql.SortDirection `json:"sortDirection"`
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

func FromSQLAccountAttributes(row sql.AccountAttribute) *AccountAttributes {
	return &AccountAttributes{
		ID:           row.ID,
		Bio:          nullHandler(row.Bio),
		ContactEmail: nullHandler(row.ContactEmail),
		Location:     nullHandler(row.Location),
		RealName:     nullHandler(row.RealName),
		GithubUrl:    nullHandler(row.GithubUrl),
		LinkedinUrl:  nullHandler(row.LinkedinUrl),
		FacebookUrl:  nullHandler(row.FacebookUrl),
		InstagramUrl: nullHandler(row.InstagramUrl),
		TwitterUrl:   nullHandler(row.TwitterUrl),
		School:       nullHandler(row.School),
		WebsiteUrl:   nullHandler(row.WebsiteUrl),
	}
}
