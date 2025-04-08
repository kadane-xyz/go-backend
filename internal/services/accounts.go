package services

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/dbaccessors"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type AccountService interface {
	ListAccounts(ctx context.Context, params sql.ListAccountsWithAttributesFilteredParams) ([]domain.Account, error)
}

type AccountServiceAccessor struct {
	dbaccessor dbaccessors.AccountAccessor
}

func NewAccountService(dbaccessor dbaccessors.AccountAccessor) AccountService {
	return &AccountServiceAccessor{dbaccessor: dbaccessor}
}

func (s *AccountServiceAccessor) ListAccounts(ctx context.Context, params sql.ListAccountsWithAttributesFilteredParams) ([]domain.Account, error) {
	accounts, err := s.dbaccessor.ListAccountsWithAttributesFiltered(ctx, params)
	if err != nil {
		return nil, err
	}
	// accounts response
	response := []domain.Account{}
	for _, account := range accounts {
		if account.Attributes == nil {
			account.Attributes = domain.AccountAttributes{}
		}

		response = append(response, domain.Account{
			ID:         account.ID,
			Username:   account.Username,
			Email:      account.Email,
			CreatedAt:  account.CreatedAt.Time,
			AvatarUrl:  account.AvatarUrl.String,
			Level:      account.Level,
			Plan:       account.Plan,
			IsAdmin:    account.Admin,
			Attributes: account.Attributes.(domain.AccountAttributes),
		})
	}

	return response, nil
}
