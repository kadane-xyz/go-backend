package dbaccessors

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type AdminAccessor interface {
	GetAdminProblems(ctx context.Context) ([]sql.GetAdminProblemsRow, error)
	ValidateAdmin(ctx context.Context, id string) (bool, error)
}

type SQLAdminAccessor struct {
	queries *sql.Queries
}

func NewSQLAdminAccessor(queries *sql.Queries) AdminAccessor {
	return &SQLAdminAccessor{queries: queries}
}

func (a *SQLAdminAccessor) GetAdminProblems(ctx context.Context) ([]sql.GetAdminProblemsRow, error) {
	return a.queries.GetAdminProblems(ctx)
}

func (a *SQLAdminAccessor) ValidateAdmin(ctx context.Context, id string) (bool, error) {
	return a.queries.ValidateAdmin(ctx, id)
}
