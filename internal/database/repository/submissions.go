package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type SubmissionsRepository interface {
	GetSubmission(ctx context.Context, params *domain.SubmissionGetParams) (*domain.Submission, error)
	GetSubmissions(ctx context.Context, ids []uuid.UUID) ([]*domain.Submission, error)
	GetSubmissionsByUsername(ctx context.Context, params *domain.SubmissionsGetByUsernameParams) ([]*domain.Submission, error)
	CreateSubmission(ctx context.Context, params *domain.SubmissionCreateParams) error
}

type SQLSubmissionsRepository struct {
	queries *sql.Queries
}

func NewSQLSubmissionsRepository(queries *sql.Queries) *SQLSubmissionsRepository {
	return &SQLSubmissionsRepository{queries: queries}
}

func (r *SQLSubmissionsRepository) GetSubmissions(ctx context.Context, ids []uuid.UUID) ([]*domain.Submission, error) {
	queryIDs := []pgtype.UUID{}
	for _, id := range ids {
		queryIDs = append(queryIDs, pgtype.UUID{Bytes: id, Valid: true})
	}

	q, err := r.queries.GetSubmissions(ctx, queryIDs)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLSubmissions(q)
}

func (r *SQLSubmissionsRepository) GetSubmission(ctx context.Context, params *domain.SubmissionGetParams) (*domain.Submission, error) {
	q, err := r.queries.GetSubmission(ctx, sql.GetSubmissionParams{
		UserID:       params.UserID,
		SubmissionID: pgtype.UUID{Bytes: params.SubmissionID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetSubmissionRow(q)
}

func (r *SQLSubmissionsRepository) GetSubmissionsByUsername(ctx context.Context, params *domain.SubmissionsGetByUsernameParams) ([]*domain.Submission, error) {
	q, err := r.queries.GetSubmissionsByUsername(ctx, sql.GetSubmissionsByUsernameParams{
		Username:      params.Username,
		ProblemID:     params.ProblemID,
		Status:        params.Status,
		Sort:          params.Sort,
		SortDirection: params.Order,
		Page:          params.Page,
		PerPage:       params.PerPage,
	})
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetSubmissionByUsernameRows(q)
}

func (r *SQLSubmissionsRepository) CreateSubmission(ctx context.Context, params *domain.SubmissionCreateParams) error {
	_, err := r.queries.CreateSubmission(ctx, sql.CreateSubmissionParams{
		ID:              pgtype.UUID{Bytes: params.ID, Valid: true},
		Stdout:          params.Stdout,
		Time:            params.Time,
		Memory:          params.Memory,
		Stderr:          params.Stdout,
		CompileOutput:   params.CompileOutput,
		Message:         params.Message,
		Status:          sql.SubmissionStatus(params.Status),
		LanguageID:      params.LanguageID,
		LanguageName:    params.LanguageName,
		AccountID:       params.AccountID,
		ProblemID:       params.ProblemID,
		SubmittedCode:   params.SubmittedCode,
		SubmittedStdin:  &params.SubmittedStdin,
		FailedTestCase:  params.FailedTestCase,
		PassedTestCases: params.PassedTestCases,
		TotalTestCases:  params.TotalTestCases,
	})
	if err != nil {
		return err
	}

	return nil
}
