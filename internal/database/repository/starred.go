package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type StarredRepository interface {
	GetStarredProblems(ctx context.Context, params sql.GetStarredProblemsParams) ([]sql.GetStarredProblemsRow, error)
}

type SQLStarredRepository struct {
	queries *sql.Queries
}

func NewSQLStarredRepository(queries *sql.Queries) *SQLStarredRepository {
	return &SQLStarredRepository{queries: queries}
}

func (r *SQLStarredRepository) GetStarredProblems(ctx context.Context, params sql.GetStarredProblemsParams) ([]sql.GetStarredProblemsRow, error) {
	q, err := r.queries.GetStarredProblems(ctx, params)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLStarredRepository) StarProblem(ctx context.Context, params sql.StarProblemParams) error {
	q, err := r.queries.StarProblem(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLStarredRepository) UnstarProblem(ctx context.Context, params sql.UnstarProblemParams) error {
	q, err := r.queries.UnstarProblem(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLStarredRepository) GetStarredProblems(ctx context.Context, params sql.GetStarredProblemsParams) ([]sql.GetStarredProblemsRow, error) {
	q, err := r.queries.GetStarredProblems(ctx, params)
	if err != nil {
		return nil, err
	}
	return q, nil
}
