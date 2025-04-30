package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type StarredRepository interface {
	GetStarredProblems(ctx context.Context, id string) ([]*domain.StarredProblem, error)
	StarProblem(ctx context.Context, params domain.StarProblemParams) error
}

type SQLStarredRepository struct {
	queries *sql.Queries
}

func NewSQLStarredRepository(queries *sql.Queries) *SQLStarredRepository {
	return &SQLStarredRepository{queries: queries}
}

func (r *SQLStarredRepository) GetStarredProblems(ctx context.Context, id string) ([]*domain.Problem, error) {
	q, err := r.queries.GetStarredProblems(ctx, id)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetStarredProblemsRow(q), nil
}

func (r *SQLStarredRepository) StarProblem(ctx context.Context, params domain.StarProblemParams) (bool, error) {
	q, err := r.queries.PutStarredProblem(ctx, sql.PutStarredProblemParams{
		UserID:    params.UserId,
		ProblemID: params.ProblemId,
	})
	if err != nil {
		return false, err
	}
	return q, nil
}
