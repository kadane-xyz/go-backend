package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type AdminRepository interface {
	GetAdminProblems(ctx context.Context) ([]sql.GetAdminProblemsRow, error)
	ValidateAdmin(ctx context.Context, id string) (bool, error)
}

type SQLAdminRepository struct {
	queries *sql.Queries
}

func NewSQLAdminRepository(queries *sql.Queries) *SQLAdminRepository {
	return &SQLAdminRepository{queries: queries}
}

func (r *SQLAdminRepository) GetAdminProblems(ctx context.Context) ([]sql.GetAdminProblemsRow, error) {
	q, err := r.queries.GetAdminProblems(ctx)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLAdminRepository) ValidateAdmin(ctx context.Context, id string) (bool, error) {
	q, err := r.queries.ValidateAdmin(ctx, id)
	if err != nil {
		return false, err
	}
	return q, nil
}
