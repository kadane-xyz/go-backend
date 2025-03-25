package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"kadane.xyz/go-backend/v2/internal/config"
)

// Create new database pool
func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	connCfg, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s/%s", cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Url, cfg.Postgres.DB))
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
