package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type CommentsRepository interface {
	GetComment(ctx context.Context, params sql.GetCommentParams) (domain.CommentRelation, error)
	GetCommentById(ctx context.Context, id int64) (domain.Comment, error)
	GetComments(ctx context.Context, params sql.GetCommentsParams) ([]domain.Comment, error)
	GetCommentsSorted(ctx context.Context, params sql.GetCommentsSortedParams) ([]domain.Comment, error)
	CreateComment(ctx context.Context, params sql.CreateCommentParams) (domain.Comment, error)
	UpdateComment(ctx context.Context, params sql.UpdateCommentParams) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, params sql.DeleteCommentParams) error
	VoteComment(ctx context.Context, params sql.VoteCommentParams) error
}

type SQLCommentsRepository struct {
	queries *sql.Queries
}

func NewSQLCommentsRepository(queries *sql.Queries) *SQLCommentsRepository {
	return &SQLCommentsRepository{queries: queries}
}

func (r *SQLCommentsRepository) GetComment(ctx context.Context, params sql.GetCommentParams) (*domain.CommentRelation, error) {
	q, err := r.queries.GetComment(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetCommentRow(q)
}

func (r *SQLCommentsRepository) GetCommentById(ctx context.Context, id int64) (*int32, error) {
	q, err := r.queries.GetCommentById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *SQLCommentsRepository) GetComments(ctx context.Context, params sql.GetCommentsParams) ([]*domain.CommentRelation, error) {
	q, err := r.queries.GetComments(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetCommentsRow(q)
}

func (r *SQLCommentsRepository) GetCommentsSorted(ctx context.Context, params sql.GetCommentsSortedParams) ([]*domain.CommentRelation, error) {
	q, err := r.queries.GetCommentsSorted(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetCommentsSorted(q)
}

func (r *SQLCommentsRepository) CreateComment(ctx context.Context, params sql.CreateCommentParams) (*domain.Comment, error) {
	q, err := r.queries.CreateComment(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLComment(q)
}

func (r *SQLCommentsRepository) UpdateComment(ctx context.Context, params sql.UpdateCommentParams) (*domain.Comment, error) {
	q, err := r.queries.UpdateComment(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLComment(q)
}

func (r *SQLCommentsRepository) DeleteComment(ctx context.Context, params sql.DeleteCommentParams) error {
	err := r.queries.DeleteComment(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLCommentsRepository) VoteComment(ctx context.Context, params sql.VoteCommentParams) error {
	err := r.queries.VoteComment(ctx, params)
	if err != nil {
		return err
	}

	return nil
}
