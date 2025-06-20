package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type AccountRepository interface {
	GetAccount(ctx context.Context, params *domain.AccountGetParams) (*domain.Account, error)
	ListAccounts(ctx context.Context, params *domain.AccountGetParams) ([]*domain.Account, error)
	CreateAccount(ctx context.Context, params *domain.AccountCreateRequest) error
	CreateAccountAttributes(ctx context.Context, params *domain.AccountAttributesCreateParams) (*domain.AccountAttributes, error)
	UploadAccountAvatar(ctx context.Context, params *domain.AccountAvatarParams) error
	GetAccountAttributes(ctx context.Context, id string) (*domain.AccountAttributes, error)
	DeleteAccount(ctx context.Context, id string) error
	GetAccountByUsername(ctx context.Context, params sql.GetAccountByUsernameParams) (*domain.Account, error)
	UpdateAccountAttributes(ctx context.Context, params *domain.AccountUpdateParams) (*domain.AccountAttributes, error)
}

type accountRepository struct {
	*DatabaseRepository
}

func NewAccountRepository(queries *sql.Queries, txManager *database.TransactionManager) *accountRepository {
	return &accountRepository{
		DatabaseRepository: NewDatabaseRepository(queries, txManager),
	}

}

func (r *accountRepository) GetAccount(ctx context.Context, params *domain.AccountGetParams) (*domain.Account, error) {
	q := r.getQueries(ctx)

	account, err := q.GetAccount(ctx, sql.GetAccountParams{
		ID:                params.ID,
		IncludeAttributes: params.IncludeAttributes,
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLAccountRow(account), nil
}

func (r *accountRepository) ListAccounts(ctx context.Context, params *domain.AccountGetParams) ([]*domain.Account, error) {
	q := r.getQueries(ctx)

	accounts, err := q.ListAccounts(ctx, sql.ListAccountsParams{
		IncludeAttributes: params.IncludeAttributes,
		UsernamesFilter:   params.UsernamesFilter,
		LocationsFilter:   params.LocationsFilter,
		Sort:              params.Sort,
		SortDirection:     params.SortDirection,
	})
	if err != nil {
		return nil, err
	}

	return domain.FromSQLListAccountsRow(accounts), nil
}

func (r *accountRepository) CreateAccount(ctx context.Context, params *domain.AccountCreateRequest) error {
	q := r.getQueries(ctx)

	err := q.CreateAccount(ctx, sql.CreateAccountParams{
		ID:       params.ID,
		Email:    params.Email,
		Username: params.Username,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *accountRepository) CreateAccountAttributes(ctx context.Context, params *domain.AccountAttributesCreateParams) (*domain.AccountAttributes, error) {
	q := r.getQueries(ctx)

	attributes, err := q.CreateAccountAttributes(ctx, *params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLAccountAttributes(attributes), nil
}

func (r *accountRepository) UploadAccountAvatar(ctx context.Context, params *domain.AccountAvatarParams) error {
	q := r.getQueries(ctx)

	err := q.UpdateAccountAvatar(ctx, sql.UpdateAccountAvatarParams{
		AvatarUrl: pgtype.Text{String: params.AvatarUrl, Valid: true},
		ID:        params.ID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *accountRepository) GetAccountAttributes(ctx context.Context, id string) (*domain.AccountAttributes, error) {
	q := r.getQueries(ctx)

	attributes, err := q.GetAccountAttributes(ctx, id)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLAccountAttributes(attributes), nil
}

func (r *accountRepository) DeleteAccount(ctx context.Context, id string) error {
	q := r.getQueries(ctx)

	err := q.DeleteAccount(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *accountRepository) GetAccountByUsername(ctx context.Context, params sql.GetAccountByUsernameParams) (*domain.Account, error) {
	q := r.getQueries(ctx)

	account, err := q.GetAccountByUsername(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLAccountByUsernameRow(account), nil
}

func (r *accountRepository) UpdateAccountAttributes(ctx context.Context, params *domain.AccountUpdateParams) (*domain.AccountAttributes, error) {
	q := r.getQueries(ctx)

	attributes, err := q.UpdateAccountAttributes(ctx, sql.UpdateAccountAttributesParams{
		Bio:          *params.Bio,
		ContactEmail: *params.ContactEmail,
		Location:     *params.Location,
		RealName:     *params.RealName,
		GithubUrl:    *params.GithubUrl,
		LinkedinUrl:  *params.LinkedinUrl,
		FacebookUrl:  *params.FacebookUrl,
		InstagramUrl: *params.InstagramUrl,
		TwitterUrl:   *params.TwitterUrl,
		School:       *params.School,
		WebsiteUrl:   *params.WebsiteUrl,
		ID:           params.ID,
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLAccountAttributes(attributes), nil
}
