package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type SolutionsRepository interface {
	GetSolutions(ctx context.Context, params sql.GetSolutionsParams) ([]sql.GetSolutionsRow, error)
	GetSolutionById(ctx context.Context, id int64) (int32, error)
}

type SQLSolutionsRepository struct {
	queries *sql.Queries
}

func NewSQLSolutionsRepository(queries *sql.Queries) *SQLSolutionsRepository {
	return &SQLSolutionsRepository{queries: queries}
}

func (r *SQLSolutionsRepository) GetSolutions(ctx context.Context, params sql.GetSolutionsParams) ([]sql.GetSolutionsRow, error) {
	q, err := r.queries.GetSolutions(ctx, params)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLSolutionsRepository) GetSolutionById(ctx context.Context, id int64) (int32, error) {
	q, err := r.queries.GetSolutionById(ctx, id)
	if err != nil {
		return 0, err
	}
	return q, nil
}
