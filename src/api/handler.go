package api

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Handler struct {
	PostgresClient  *pgxpool.Pool
	PostgresQueries *sql.Queries
}
