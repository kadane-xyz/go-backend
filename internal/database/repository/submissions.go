package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type SubmissionsRepository interface {
	GetSubmissions(ctx context.Context, params sql.GetSubmissionsParams) ([]sql.GetSubmissionsRow, error)
	GetSubmission(ctx context.Context, id string) (*domain.Submission, error)
}

type SQLSubmissionsRepository struct {
	queries *sql.Queries
}

func NewSQLSubmissionsRepository(queries *sql.Queries) *SQLSubmissionsRepository {
	return &SQLSubmissionsRepository{queries: queries}
}

func (r *SQLSubmissionsRepository) GetSubmissions(ctx context.Context, params sql.GetSubmissionsParams) ([]sql.GetSubmissionsRow, error) {
	q, err := r.queries.GetSubmissions(ctx, params)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLSubmissionsRepository) GetSubmission(ctx context.Context, id string) (sql.GetSubmissionByIdRow, error) {
	q, err := r.queries.GetSubmission(ctx, id)
	if err != nil {
		return nil, err
	}
	return q, nil
}
