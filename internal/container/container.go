package container

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"kadane.xyz/go-backend/v2/internal/config"
	"kadane.xyz/go-backend/v2/internal/database"
	"kadane.xyz/go-backend/v2/internal/database/sql"
)

// APIHandlers is an inner container holding all API handlers
type APIHandlers struct {
}

// Container is a service locator holding all application dependencies
type Container struct {
	Config      *config.Config
	DB          *pgxpool.Pool
	SqlQueries  *sql.Queries
	APIHandlers *APIHandlers
}

// NewContainer creates the application's container
func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	// Create the Redis client
	/*redisClient, err := redisClient.NewRedisClient(ctx, log, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create the Redis client: %w", err)
	}*/

	// Create the database pool
	dbPool, err := database.NewPool(ctx, *cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create the database connection pool: %w", err)
	}

	// Create the queries client
	queries := sql.New(dbPool)

	// Create the database repositories

	// Create the API handlers

	return &Container{
		Config:      cfg,
		DB:          dbPool,
		SqlQueries:  queries,
		APIHandlers: &APIHandlers{},
	}, nil
}

// Close releases all resources held by the container
func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
	/*if c.RedisClient != nil {
		c.RedisClient.Close()
	}*/
}
