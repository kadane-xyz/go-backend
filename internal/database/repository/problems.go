package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type ProblemsRepository interface {
	GetProblem(ctx context.Context, params *domain.ProblemGetParams) (*domain.Problem, error)
	GetProblems(ctx context.Context, params *domain.ProblemsGetParams) ([]*domain.Problem, error)
	CreateProblem(ctx context.Context, params *domain.ProblemCreateParams) (*domain.ProblemCreate, error)
	GetProblemTestCases(ctx context.Context, params *domain.ProblemTestCasesGetParams) ([]*domain.TestCase, error)
}

type SQLProblemsRepository struct {
	queries *sql.Queries
}

func NewSQLProblemsRepository(queries *sql.Queries) *SQLProblemsRepository {
	return &SQLProblemsRepository{queries: queries}
}

func (r *SQLProblemsRepository) GetProblem(ctx context.Context, params *domain.ProblemGetParams) (*domain.Problem, error) {
	q, err := r.queries.GetProblem(ctx, sql.GetProblemParams{
		UserID:    params.UserID,
		ProblemID: int32(params.ProblemID),
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetProblemRow(q)
}

func (r *SQLProblemsRepository) GetProblems(ctx context.Context, params *domain.ProblemsGetParams) ([]*domain.Problem, error) {
	q, err := r.queries.GetProblems(ctx, sql.GetProblemsParams{
		UserID:        params.UserID,
		Title:         params.Title,
		Difficulty:    string(params.Difficulty),
		Sort:          params.Sort,
		SortDirection: params.Order,
		Page:          params.Page,
		PerPage:       params.PerPage,
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetProblemsRow(q)
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

func (r *SQLProblemsRepository) GetProblemTestCases(ctx context.Context, params *domain.ProblemTestCasesGetParams) ([]*domain.TestCase, error) {
	q, err := r.queries.GetProblemTestCases(ctx, sql.GetProblemTestCasesParams{
		ProblemID:  params.ProblemID,
		Visibility: string(params.Visibility),
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetProblemTestCases(q)
}
