package api

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Handler struct {
	PostgresClient  *pgxpool.Pool
	PostgresQueries *sql.Queries
	AWSClient       *s3.Client
	AWSBucketAvatar string
}
