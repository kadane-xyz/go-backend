package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresClient(ctx context.Context, PostgresUrl, PostgresUser, PostgresPass, PostgresDB string) (*pgxpool.Pool, func(), error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s", PostgresUser, PostgresPass, PostgresUrl, PostgresDB)
	connCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing connection string: %v", err)
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, connCfg)
	if err != nil {
		log.Printf("Error database pool: %v", err)
		return nil, nil, fmt.Errorf("error creating database pool: %v", err)
	}

	closeFunc := func() {
		dbpool.Close()
		log.Println("Database pool closed")
	}

	return dbpool, closeFunc, nil
}
