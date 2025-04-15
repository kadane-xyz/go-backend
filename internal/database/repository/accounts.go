package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type AccountRepository interface {
	GetAccount(ctx context.Context, params sql.GetAccountParams) (domain.Account, error)
	ListAccounts(ctx context.Context, params sql.ListAccountsParams) ([]domain.Account, error)
	CreateAccount(ctx context.Context, params sql.CreateAccountParams) error
	CreateAccountAttributes(ctx context.Context, params sql.CreateAccountAttributesParams) (sql.AccountAttribute, error)
	UploadAccountAvatar(ctx context.Context, params sql.UpdateAccountAvatarParams) error
	GetAccountAttributes(ctx context.Context, id string) (sql.AccountAttribute, error)
	DeleteAccount(ctx context.Context, id string) error
	GetAccountByUsername(ctx context.Context, params sql.GetAccountByUsernameParams) (domain.Account, error)
	UpdateAccountAttributes(ctx context.Context, params sql.UpdateAccountAttributesParams) (sql.AccountAttribute, error)
}

type SQLAccountsRepository struct {
	queries *sql.Queries
}

func NewSQLAccountsRepository(queries *sql.Queries) *SQLAccountsRepository {
	return &SQLAccountsRepository{queries: queries}
}

func (r *SQLAccountsRepository) GetAccount(ctx context.Context, params sql.GetAccountParams) (domain.Account, error) {
	q, err := r.queries.GetAccount(ctx, params)
	if err != nil {
		return domain.Account{}, err
	}
	return domain.FromSQLAccountRow(q), nil
}

func (r *SQLAccountsRepository) ListAccounts(ctx context.Context, params sql.ListAccountsParams) ([]domain.Account, error) {
	q, err := r.queries.ListAccounts(ctx, params)
	if err != nil {
		return nil, err
	}

	return domain.FromSQLListAccountsRow(q), nil
}

func (r *SQLAccountsRepository) CreateAccount(ctx context.Context, params sql.CreateAccountParams) error {
	err := r.queries.CreateAccount(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLAccountsRepository) CreateAccountAttributes(ctx context.Context, params sql.CreateAccountAttributesParams) (sql.AccountAttribute, error) {
	q, err := r.queries.CreateAccountAttributes(ctx, params)
	if err != nil {
		return sql.AccountAttribute{}, err
	}
	return q, nil
}

func (r *SQLAccountsRepository) UploadAccountAvatar(ctx context.Context, params sql.UpdateAccountAvatarParams) error {
	err := r.queries.UpdateAccountAvatar(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLAccountsRepository) GetAccountAttributes(ctx context.Context, id string) (sql.AccountAttribute, error) {
	q, err := r.queries.GetAccountAttributes(ctx, id)
	if err != nil {
		return sql.AccountAttribute{}, err
	}
	return q, nil
}

func (r *SQLAccountsRepository) DeleteAccount(ctx context.Context, id string) error {
	err := r.queries.DeleteAccount(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLAccountsRepository) GetAccountByUsername(ctx context.Context, params sql.GetAccountByUsernameParams) (domain.Account, error) {
	q, err := r.queries.GetAccountByUsername(ctx, params)
	if err != nil {
		return domain.Account{}, err
	}
	return domain.FromSQLAccountByUsernameRow(q), nil
}

func (r *SQLAccountsRepository) UpdateAccountAttributes(ctx context.Context, params sql.UpdateAccountAttributesParams) (sql.AccountAttribute, error) {
	q, err := r.queries.UpdateAccountAttributes(ctx, params)
	if err != nil {
		return sql.AccountAttribute{}, err
	}
	return q, nil
}
