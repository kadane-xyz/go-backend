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
	CreateSubmission(ctx context.Context, params domain.SubmissionCreateParams) error
}

type SQLSubmissionsRepository struct {
	queries *sql.Queries
}

func NewSQLSubmissionsRepository(queries *sql.Queries) *SQLSubmissionsRepository {
	return &SQLSubmissionsRepository{queries: queries}
}

func (r *SQLSubmissionsRepository) GetSubmissions(ctx context.Context, ids []uuid.UUID) ([]*domain.Submission, error) {
	queryIds := []pgtype.UUID{}
	for _, id := range ids {
		queryIds = append(queryIds, pgtype.UUID{Bytes: id, Valid: true})
	}

	q, err := r.queries.GetSubmissions(ctx, queryIds)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLSubmissions(q)
}

func (r *SQLSubmissionsRepository) GetSubmission(ctx context.Context, id string) (*domain.Submission, error) {
	q, err := r.queries.GetSubmission(ctx, id)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetSubmissionRow(q)
}

func (r *SQLSubmissionsRepository) GetSubmissionByUsername(ctx context.Context, params *domain.SubmissionsGetByUsernameParams) ([]*domain.Submission, error) {
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

func (r *SQLSubmissionsRepository) CreateSubmission(ctx context.Context, params domain.SubmissionCreateParams) error {
	_, err := r.queries.CreateSubmission(ctx, sql.CreateSubmissionParams{
		ID:              pgtype.UUID{Bytes: params.Id, Valid: true},
		Stdout:          params.Stdout,
		Time:            pgtype.Time{Microseconds: params.Time.Unix(), Valid: true},
		Memory:          params.Memory,
		Stderr:          params.Stdout,
		CompileOutput:   params.CompileOutput,
		Message:         params.Message,
		Status:          sql.SubmissionStatus(params.Status),
		LanguageID:      params.LanguageId,
		LanguageName:    params.LanguageName,
		AccountID:       params.AccountId,
		ProblemID:       params.ProblemId,
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
