package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type StarredRepository interface {
	// Starred problems
	GetStarredProblems(ctx context.Context, id string) ([]*domain.StarredProblem, error)
	StarProblem(ctx context.Context, params domain.StarProblemParams) error
	// Starred solutions
	GetStarredSolutions(ctx context.Context, id string) ([]*domain.StarredSolution, error)
	StarSolution(ctx context.Context, params domain.StarSolutionParams) error
	// Starred submissions
	GetStarredSubmissions(ctx context.Context, id string) ([]*domain.StarredSubmission, error)
	StarSubmission(ctx context.Context, params domain.StarSubmissionParams) error
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

func (r *SQLStarredRepository) GetStarredSolutions(ctx context.Context, id string) ([]*domain.StarredSolution, error) {
	q, err := r.queries.GetStarredSolutions(ctx, id)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetStarredSolutionsRows(q)
}

func (r *SQLStarredRepository) StarSolution(ctx context.Context, params domain.StarSolutionParams) (bool, error) {
	q, err := r.queries.PutStarredSolution(ctx, sql.PutStarredSolutionParams{
		UserID:     params.UserId,
		SolutionID: params.SolutionId,
	})
	if err != nil {
		return false, err
	}
	return q, nil
}

func (r *SQLStarredRepository) GetStarredSubmissions(ctx context.Context, id string) ([]*domain.StarredSubmission, error) {
	q, err := r.queries.GetStarredSubmissions(ctx, id)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetStarredSubmissionRow(q), nil
}

func (r *SQLStarredRepository) StarSubmission(ctx context.Context, params domain.StarSubmissionParams) (bool, error) {
	q, err := r.queries.PutStarredSubmission(ctx, sql.PutStarredSubmissionParams{
		UserID:       params.UserId,
		SubmissionID: pgtype.UUID{Bytes: params.SubmissionId, Valid: true},
	})
	if err != nil {
		return false, err
	}
	return q, nil
}
