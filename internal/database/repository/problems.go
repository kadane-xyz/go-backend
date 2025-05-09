package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type ProblemsRepository interface {
	GetProblem(ctx context.Context, params *domain.ProblemGetParams) (domain.Problem, error)
	GetProblemsFilteredPaginated(ctx context.Context, params *domain.ProblemsGetParams) ([]sql.GetProblemsFilteredPaginatedRow, error)
	CreateProblem(ctx context.Context, params *domain.ProblemCreateParams) (*domain.ProblemCreate, error)
}

type SQLProblemsRepository struct {
	queries *sql.Queries
}

func NewSQLProblemsRepository(queries *sql.Queries) *SQLProblemsRepository {
	return &SQLProblemsRepository{queries: queries}
}

func (r *SQLProblemsRepository) GetProblem(ctx context.Context, params *domain.ProblemGetParams) (sql.GetProblemRow, error) {
	q, err := r.queries.GetProblem(ctx, sql.GetProblemParams{
		UserID:    params.UserId,
		ProblemID: int32(params.ProblemId),
	})
	if err != nil {
		return sql.GetProblemRow{}, err
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

func (r *SQLProblemsRepository) CreateProblem(ctx context.Context, params *domain.ProblemCreateParams) (*domain.ProblemCreate, error) {
	q, err := r.queries.CreateProblem(ctx, sql.CreateProblemParams{
		Title:        params.Title,
		Description:  params.Description,
		FunctionName: params.FunctionName,
		Points:       params.Points,
		Tags:         params.Tags,
		Difficulty:   sql.ProblemDifficulty(params.Description),
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLCreateProblemRow(q)
}
