package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ctxKey string

const txKey ctxKey = "tx"

type TransactionManager struct {
	Pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{
		Pool: pool,
	}
}

func (m *TransactionManager) WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	if tx := m.GetTx(ctx); tx != nil {
		return fn(ctx)
	}

	tx, err := m.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txKey, tx)

	err = fn(txCtx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, rbErr)
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (m *TransactionManager) GetTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}

	return nil
}
