package dbaccessors

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type ProblemsAccessor interface {
	GetProblem(ctx context.Context, params sql.GetProblemParams) (sql.GetProblemRow, error)
	CreateProblem(ctx context.Context, params sql.CreateProblemParams) (sql.CreateProblemRow, error)
	GetProblemsFilteredPaginated(ctx context.Context, params sql.GetProblemsFilteredPaginatedParams) ([]sql.GetProblemsFilteredPaginatedRow, error)
}

type SQLProblemsAccessor struct {
	queries *sql.Queries
}

func NewSQLProblemsAccessor(queries *sql.Queries) ProblemsAccessor {
	return &SQLProblemsAccessor{queries: queries}
}

func (a *SQLProblemsAccessor) GetProblem(ctx context.Context, params sql.GetProblemParams) (sql.GetProblemRow, error) {
	return a.queries.GetProblem(ctx, params)
}

func (a *SQLProblemsAccessor) CreateProblem(ctx context.Context, params sql.CreateProblemParams) (sql.CreateProblemRow, error) {
	return a.queries.CreateProblem(ctx, params)
}

func (a *SQLProblemsAccessor) GetProblemsFilteredPaginated(ctx context.Context, params sql.GetProblemsFilteredPaginatedParams) ([]sql.GetProblemsFilteredPaginatedRow, error) {
	return a.queries.GetProblemsFilteredPaginated(ctx, params)
}
