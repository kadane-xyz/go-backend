package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type CommentsRepository interface {
	GetComment(ctx context.Context, params sql.GetCommentParams) (*domain.Comment, error)
	GetCommentByID(ctx context.Context, id int64) (*bool, error)
	GetComments(ctx context.Context, params sql.GetCommentsParams) ([]*domain.Comment, error)
	GetCommentsSorted(ctx context.Context, params sql.GetCommentsSortedParams) ([]*domain.Comment, error)
	CreateComment(ctx context.Context, params *domain.CommentCreateParams) (*domain.Comment, error)
	UpdateComment(ctx context.Context, params sql.UpdateCommentParams) (*domain.Comment, error)
	DeleteComment(ctx context.Context, params sql.DeleteCommentParams) error
	VoteComment(ctx context.Context, params sql.VoteCommentParams) error
}

type SQLCommentsRepository struct {
	queries *sql.Queries
}

func NewSQLCommentsRepository(queries *sql.Queries) *SQLCommentsRepository {
	return &SQLCommentsRepository{queries: queries}
}

func (r *SQLCommentsRepository) GetComment(ctx context.Context, params sql.GetCommentParams) (*domain.Comment, error) {
	q, err := r.queries.GetComment(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetCommentRow(q)
}

func (r *SQLCommentsRepository) GetCommentByID(ctx context.Context, id int64) (*bool, error) {
	q, err := r.queries.GetCommentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *SQLCommentsRepository) GetComments(ctx context.Context, params sql.GetCommentsParams) ([]*domain.Comment, error) {
	q, err := r.queries.GetComments(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetCommentsRow(q), nil
}

func (r *SQLCommentsRepository) GetCommentsSorted(ctx context.Context, params sql.GetCommentsSortedParams) ([]*domain.Comment, error) {
	q, err := r.queries.GetCommentsSorted(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetCommentsSorted(q), nil
}

func (r *SQLCommentsRepository) CreateComment(ctx context.Context, params *domain.CommentCreateParams) (*domain.Comment, error) {
	q, err := r.queries.CreateComment(ctx, sql.CreateCommentParams{
		UserID:     params.UserID,
		SolutionID: params.SolutionID,
		ParentID:   pgtype.Int8{Int64: *params.ParentID, Valid: true},
		Body:       params.Body,
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLComment(q), nil
}

func (r *SQLCommentsRepository) UpdateComment(ctx context.Context, params sql.UpdateCommentParams) (*domain.Comment, error) {
	q, err := r.queries.UpdateComment(ctx, params)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLComment(q), nil
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
