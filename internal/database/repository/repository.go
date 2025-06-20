package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database"
	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type DatabaseRepository struct {
	queries   *sql.Queries
	txManager *database.TransactionManager
}

func NewDatabaseRepository(queries *sql.Queries, txManager *database.TransactionManager) *DatabaseRepository {
	return &DatabaseRepository{
		queries:   queries,
		txManager: txManager,
	}
}

func (r *DatabaseRepository) getQueries(ctx context.Context) *sql.Queries {
	if tx := r.txManager.GetTx(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}

	return r.queries
}
