package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type ProblemsRepository interface {
	GetProblem(ctx context.Context, params sql.GetProblemParams) (sql.GetProblemRow, error)
	CreateProblem(ctx context.Context, params sql.CreateProblemParams) (sql.CreateProblemRow, error)
	GetProblemsFilteredPaginated(ctx context.Context, params sql.GetProblemsFilteredPaginatedParams) ([]sql.GetProblemsFilteredPaginatedRow, error)
}

type SQLProblemsRepository struct {
	queries *sql.Queries
}

func NewSQLProblemsRepository(queries *sql.Queries) *SQLProblemsRepository {
	return &SQLProblemsRepository{queries: queries}
}

func (r *SQLProblemsRepository) GetProblem(ctx context.Context, params sql.GetProblemParams) (sql.GetProblemRow, error) {
	q, err := r.queries.GetProblem(ctx, params)
	if err != nil {
		return sql.GetProblemRow{}, err
	}
	return q, nil
}

func (r *SQLProblemsRepository) CreateProblem(ctx context.Context, params sql.CreateProblemParams) (sql.CreateProblemRow, error) {
	q, err := r.queries.CreateProblem(ctx, params)
	if err != nil {
		return sql.CreateProblemRow{}, err
	}
	return q, nil
}

func (r *SQLProblemsRepository) GetProblemsFilteredPaginated(ctx context.Context, params sql.GetProblemsFilteredPaginatedParams) ([]sql.GetProblemsFilteredPaginatedRow, error) {
	q, err := r.queries.GetProblemsFilteredPaginated(ctx, params)
	if err != nil {
		return nil, err
	}
	return q, nil
}
