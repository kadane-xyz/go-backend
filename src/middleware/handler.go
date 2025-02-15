package middleware

import (
	firebase "firebase.google.com/go/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"kadane.xyz/go-backend/v2/src/config"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Handler struct {
	Config          *config.Config
	FirebaseApp     *firebase.App
	PostgresClient  *pgxpool.Pool
	PostgresQueries *sql.Queries
}
