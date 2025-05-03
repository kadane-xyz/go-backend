package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type SubmissionsRepository interface {
	GetSubmissions(ctx context.Context, ids []uuid.UUID) ([]*domain.Submission, error)
	GetSubmission(ctx context.Context, id string) (*domain.Submission, error)
	GetSubmissionByUserName(ctx context.Context, params domain.SubmissionGetParams) ([]*domain.Submission, error)
}

type SQLSubmissionsRepository struct {
	queries *sql.Queries
}

func NewSQLSubmissionsRepository(queries *sql.Queries) *SQLSubmissionsRepository {
	return &SQLSubmissionsRepository{queries: queries}
}

func (r *SQLSubmissionsRepository) GetSubmissions(ctx context.Context, ids []uuid.UUID) ([]domain.Submission, error) {
	q, err := r.queries.GetSubmissions(ctx, []pgtype.UUID{
		Bytes:  ids,
		Status: Valid,
	})
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLSubmissionsRepository) GetSubmission(ctx context.Context, id string) (*domain.Submission, error) {
	q, err := r.queries.GetSubmission(ctx, id)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLSubmissionsRepository) GetSubmissionByUsername(ctx context.Context, params domain.SubmissionGetParams) ([]*domain.Submission, error) {
	q, err := r.queries.GetSubmissionsByUsername(ctx, sql.GetSubmissionsByUsernameParams{
		Sort:          params.Sort,
		SortDirection: params.SortDirection,
		UserID:        params.UserId,
		Username:      params.Username,
		ProblemID:     params.ProblemID,
		Status:        params.Status,
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetSubmissionByUsernameRows(q)
}
