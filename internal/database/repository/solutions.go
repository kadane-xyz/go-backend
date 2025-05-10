package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type SolutionsRepository interface {
	GetSolution(ctx context.Context, params *domain.SolutionGetParams) (*domain.SolutionGetParams, error)
	GetSolutions(ctx context.Context, params *domain.SolutionsGetParams) ([]*domain.SolutionsRelations, error)
	GetSolutionById(ctx context.Context, id int32) (int32, error)
	CreateSolution(ctx context.Context, params *domain.SolutionsCreateParams) (*domain.Solution, error)
	UpdateSolution(ctx context.Context, params *domain.SolutionsUpdateParams) (*domain.Solution, error)
	DeleteSolution(ctx context.Context, userid string, id int32) error
	VoteSolution(ctx context.Context, params *domain.VoteSolutionsParams) error
}

type SQLSolutionsRepository struct {
	queries *sql.Queries
}

func NewSQLSolutionsRepository(queries *sql.Queries) *SQLSolutionsRepository {
	return &SQLSolutionsRepository{queries: queries}
}

func (r *SQLSolutionsRepository) GetSolution(ctx context.Context, params *domain.SolutionGetParams) (*domain.SolutionRelations, error) {
	q, err := r.queries.GetSolution(ctx, sql.GetSolutionParams{
		UserID: params.UserId,
		ID:     params.Id,
	})
	if err != nil {
		return nil, err
	}

	return domain.FromSQLGetSolutionRow(q), err
}

func (r *SQLSolutionsRepository) GetSolutions(ctx context.Context, params *domain.SolutionsGetParams) ([]*domain.SolutionsRelations, error) {
	q, err := r.queries.GetSolutions(ctx, sql.GetSolutionsParams{
		UserID:        params.UserId,
		ProblemID:     &params.ProblemId,
		Page:          params.Page,
		PerPage:       params.PerPage,
		Tags:          params.Tags,
		Title:         params.Title,
		Sort:          params.Sort,
		SortDirection: params.SortDirection,
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetSolutionsRow(q), nil
}

func (r *SQLSolutionsRepository) GetSolutionById(ctx context.Context, id int32) (int32, error) {
	q, err := r.queries.GetSolutionById(ctx, id)
	if err != nil {
		return 0, err
	}
	return q, nil
}

func (r *SQLSolutionsRepository) DeleteSolution(ctx context.Context, userId string, id int32) error {
	return r.queries.DeleteSolution(ctx, sql.DeleteSolutionParams{
		UserID: &userId,
		ID:     id,
	})
}

func (r *SQLSolutionsRepository) VoteSolution(ctx context.Context, params *domain.VoteSolutionsParams) error {
	return r.queries.VoteSolution(ctx, sql.VoteSolutionParams{
		UserID:     params.UserId,
		SolutionID: params.SolutionId,
		Vote:       params.Vote,
	})
}
