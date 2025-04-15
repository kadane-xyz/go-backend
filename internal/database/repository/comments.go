package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type CommentsRepository interface {
	GetComment(ctx context.Context, params sql.GetCommentParams) (domain.Comment, error)
	GetCommentById(ctx context.Context, id string) (domain.Comment, error)
	GetComments(ctx context.Context, params sql.GetCommentsParams) ([]domain.Comment, error)
	GetCommentsSorted(ctx context.Context, params sql.GetCommentsSortedParams) ([]domain.Comment, error)
}

type SQLCommentsRepository struct {
	queries *sql.Queries
}

func NewSQLCommentsRepository(queries *sql.Queries) *SQLCommentsRepository {
	return &SQLCommentsRepository{queries: queries}
}

func (r *SQLCommentsRepository) GetComment(ctx context.Context, params sql.GetCommentParams) (domain.Comment, error) {
	q, err := r.queries.GetComment(ctx, params)
	if err != nil {
		return domain.Comment{}, err
	}
	return domain.FromSQLCommentRow(q), nil
}

func (r *SQLCommentsRepository) GetCommentById(ctx context.Context, id string) (domain.Comment, error) {
	q, err := r.queries.GetCommentById(ctx, id)
	if err != nil {
		return domain.Comment{}, err
	}
	return domain.FromSQLCommentRow(q), nil
}

func (r *SQLCommentsRepository) GetComments(ctx context.Context, params sql.GetCommentsParams) ([]domain.Comment, error) {
	q, err := r.queries.GetComments(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLCommentsRow(q), nil
}
