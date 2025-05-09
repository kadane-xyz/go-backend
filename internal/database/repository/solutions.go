package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type SolutionsRepository interface {
	GetSolutions(ctx context.Context, params *domain.SolutionsGetParams) ([]*domain.Solution, error)
	GetSolutionById(ctx context.Context, id int32) (int32, error)
}

type SQLSolutionsRepository struct {
	queries *sql.Queries
}

func NewSQLSolutionsRepository(queries *sql.Queries) *SQLSolutionsRepository {
	return &SQLSolutionsRepository{queries: queries}
}

func (r *SQLSolutionsRepository) GetSolutions(ctx context.Context, params *domain.SolutionsGetParams) ([]*domain.SolutionRelations, error) {
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
	return FromGetSolutionsRow(q), nil
}

func (r *SQLSolutionsRepository) GetSolutionById(ctx context.Context, id int32) (int32, error) {
	q, err := r.queries.GetSolutionById(ctx, id)
	if err != nil {
		return 0, err
	}
	return q, nil
}
