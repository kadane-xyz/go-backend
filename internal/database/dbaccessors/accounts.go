package dbaccessors

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type AccountAccessor interface {
	GetAccount(ctx context.Context, params sql.GetAccountParams) (sql.GetAccountRow, error)
	ListAccountsWithAttributesFiltered(ctx context.Context, params sql.ListAccountsWithAttributesFilteredParams) ([]sql.ListAccountsWithAttributesFilteredRow, error)
	CreateAccount(ctx context.Context, params sql.CreateAccountParams) error
	CreateAccountAttributes(ctx context.Context, params sql.CreateAccountAttributesParams) (sql.AccountAttribute, error)
	UploadAccountAvatar(ctx context.Context, params sql.UpdateAccountAvatarParams) error
	GetAccountAttributes(ctx context.Context, id string) (sql.AccountAttribute, error)
	DeleteAccount(ctx context.Context, id string) error
	GetAccountByUsername(ctx context.Context, params sql.GetAccountByUsernameParams) (sql.GetAccountByUsernameRow, error)
	UpdateAccountAttributes(ctx context.Context, params sql.UpdateAccountAttributesParams) (sql.AccountAttribute, error)
}

type SQLAccountsAccessor struct {
	queries *sql.Queries
}

func NewSQLAccountsAccessor(queries *sql.Queries) AccountAccessor {
	return &SQLAccountsAccessor{queries: queries}
}

func (a *SQLAccountsAccessor) GetAccount(ctx context.Context, params sql.GetAccountParams) (sql.GetAccountRow, error) {
	return a.queries.GetAccount(ctx, params)
}

func (a *SQLAccountsAccessor) ListAccountsWithAttributesFiltered(ctx context.Context, params sql.ListAccountsWithAttributesFilteredParams) ([]sql.ListAccountsWithAttributesFilteredRow, error) {
	return a.queries.ListAccountsWithAttributesFiltered(ctx, params)
}

func (a *SQLAccountsAccessor) CreateAccount(ctx context.Context, params sql.CreateAccountParams) error {
	return a.queries.CreateAccount(ctx, params)
}

func (a *SQLAccountsAccessor) CreateAccountAttributes(ctx context.Context, params sql.CreateAccountAttributesParams) (sql.AccountAttribute, error) {
	return a.queries.CreateAccountAttributes(ctx, params)
}

func (a *SQLAccountsAccessor) UploadAccountAvatar(ctx context.Context, params sql.UpdateAccountAvatarParams) error {
	return a.queries.UpdateAccountAvatar(ctx, params)
}

func (a *SQLAccountsAccessor) GetAccountAttributes(ctx context.Context, id string) (sql.AccountAttribute, error) {
	return a.queries.GetAccountAttributes(ctx, id)
}

func (a *SQLAccountsAccessor) DeleteAccount(ctx context.Context, id string) error {
	return a.queries.DeleteAccount(ctx, id)
}

func (a *SQLAccountsAccessor) GetAccountByUsername(ctx context.Context, params sql.GetAccountByUsernameParams) (sql.GetAccountByUsernameRow, error) {
	return a.queries.GetAccountByUsername(ctx, params)
}

func (a *SQLAccountsAccessor) UpdateAccountAttributes(ctx context.Context, params sql.UpdateAccountAttributesParams) (sql.AccountAttribute, error) {
	return a.queries.UpdateAccountAttributes(ctx, params)
}
