package dbaccessors

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type AccountAccessor interface {
	GetAccountWithAttributes(ctx context.Context, id string) (*sql.GetAccountWithAttributesRow, error)
	ListAccountsWithAttributes(ctx context.Context, params sql.ListAccountsWithAttributesParams) ([]sql.ListAccountsWithAttributesRow, error)

	GetAccountAttributesWithAccount(ctx context.Context, id string) (*sql.GetAccountAttributesWithAccountRow, error)
	GetAccountAvatarUrl(ctx context.Context, id string) (pgtype.Text, error)
	GetAccountExists(ctx context.Context, id string) (bool, error)
	GetAccountIDByUsername(ctx context.Context, username string) (string, error)
	GetAccountLevel(ctx context.Context, id string) (int32, error)
	GetAccountPlan(ctx context.Context, id string) (sql.AccountPlan, error)
	GetAccountUsername(ctx context.Context, id string) (string, error)
	GetAccountByUsername(ctx context.Context, username string) (*sql.GetAccountByUsernameRow, error)
	GetAccountValidation(ctx context.Context, id string) (*sql.Account, error)
	GetAccounts(ctx context.Context, params sql.GetAccountsParams) ([]sql.GetAccountsRow, error)
	CreateAccount(ctx context.Context, params sql.CreateAccountParams) error
	CreateAccountAttributes(ctx context.Context, params sql.CreateAccountAttributesParams) (sql.AccountAttribute, error)
	UpdateAccountAttributes(ctx context.Context, params sql.UpdateAccountAttributesParams) (sql.AccountAttribute, error)
	UpdateAccountAvatar(ctx context.Context, params sql.UpdateAvatarParams) error
	DeleteAccount(ctx context.Context, id string) error
}

type SQLAccountsAccessor struct {
	queries *sql.Queries
}

func NewSQLAccountsAccessor(queries *sql.Queries) *SQLAccountsAccessor {
	return &SQLAccountsAccessor{queries: queries}
}

func (a *SQLAccountsAccessor) GetAccount(ctx context.Context, id string) (*sql.Account, error) {
	return a.queries.GetAccount(ctx, id)
}

func (a *SQLAccountsAccessor) GetAccountByUsername(ctx context.Context, username string) (*sql.Account, error) {
	return a.queries.GetAccountByUsername(ctx, username)
}

func (a *SQLAccountsAccessor) GetAccountValidation(ctx context.Context, id string) (*sql.AccountValidation, error) {
	return a.queries.GetAccountValidation(ctx, id)
}

func (a *SQLAccountsAccessor) GetAccounts(ctx context.Context, params sql.GetAccountsParams) ([]sql.Account, error) {
	return a.queries.GetAccounts(ctx, params)
}

func (a *SQLAccountsAccessor) CreateAccount(ctx context.Context, params sql.CreateAccountParams) (*sql.Account, error) {
	return a.queries.CreateAccount(ctx, params)
}

func (a *SQLAccountsAccessor) UpdateAccount(ctx context.Context, params sql.UpdateAccountParams) (*sql.Account, error) {
	return a.queries.UpdateAccount(ctx, params)
}

func (a *SQLAccountsAccessor) DeleteAccount(ctx context.Context, id string) error {
	return a.queries.DeleteAccount(ctx, id)
}
