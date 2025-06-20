package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"kadane.xyz/go-backend/v2/internal/config"
)

// Create new database pool
func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	connCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, connCfg)
	if err != nil {
		return nil, err
	}

	// Verify connectivity with a ping
	if err := dbpool.Ping(ctx); err != nil {
		dbpool.Close()
		log.Println("Database pool not connected", err)
		return nil, err
	}

	log.Println("Database pool connected")

	return dbpool, nil
}
