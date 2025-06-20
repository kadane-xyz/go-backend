package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type SolutionsRepository interface {
	GetSolution(ctx context.Context, params *domain.SolutionGetParams) (*domain.Solution, error)
	GetSolutions(ctx context.Context, params *domain.SolutionsGetParams) ([]*domain.Solution, error)
	GetSolutionById(ctx context.Context, id int32) (int32, error)
	CreateSolution(ctx context.Context, params *domain.SolutionsCreateParams) error
	UpdateSolution(ctx context.Context, params *domain.SolutionsUpdateParams) error
	DeleteSolution(ctx context.Context, userid string, id int32) error
	VoteSolution(ctx context.Context, params *domain.VoteSolutionsParams) error
}

type SQLSolutionsRepository struct {
	queries *sql.Queries
}

func NewSQLSolutionsRepository(queries *sql.Queries) *SQLSolutionsRepository {
	return &SQLSolutionsRepository{queries: queries}
}

func (r *SQLSolutionsRepository) GetSolution(ctx context.Context, params *domain.SolutionGetParams) (*domain.Solution, error) {
	q, err := r.queries.GetSolution(ctx, sql.GetSolutionParams{
		UserID: params.UserID,
		ID:     params.ID,
	})
	if err != nil {
		return nil, err
	}

	return domain.FromSQLGetSolutionRow(q), err
}

func (r *SQLSolutionsRepository) GetSolutions(ctx context.Context, params *domain.SolutionsGetParams) ([]*domain.Solution, error) {
	q, err := r.queries.GetSolutions(ctx, sql.GetSolutionsParams{
		UserID:        params.UserId,
		ProblemID:     pgtype.Int4{Int32: params.ProblemID, Valid: true},
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

func (r *SQLSolutionsRepository) CreateSolution(ctx context.Context, params *domain.SolutionsCreateParams) error {
	return r.queries.CreateSolution(ctx, sql.CreateSolutionParams{
		UserID:    pgtype.Text{String: params.UserID, Valid: true},
		Title:     params.Title,
		Tags:      params.Tags,
		Body:      params.Body,
		ProblemID: pgtype.Int4{Int32: *params.ProblemID, Valid: true},
	})
}

func (r *SQLSolutionsRepository) UpdateSolution(ctx context.Context, params *domain.SolutionsUpdateParams) error {
	return r.queries.UpdateSolution(ctx, sql.UpdateSolutionParams{
		UserID: pgtype.Text{String: params.UserID, Valid: true},
		Title:  params.Title,
		Body:   params.Body,
		Tags:   params.Tags,
		ID:     *params.SolutionID,
	})
}

func (r *SQLSolutionsRepository) DeleteSolution(ctx context.Context, userId string, id int32) error {
	return r.queries.DeleteSolution(ctx, sql.DeleteSolutionParams{
		UserID: pgtype.Text{String: userId, Valid: true},
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
