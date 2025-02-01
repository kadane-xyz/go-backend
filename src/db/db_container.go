// db_container.go
package db

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// dbPool is the global connection pool used by all tests.
var DBPool *pgxpool.Pool

// teardownFunc holds the function to shut down the container.
var teardownFunc func()

// SetupTestContainer starts the PostgreSQL container and creates the connection pool.
// It returns an error if the setup fails.
func SetupTestContainer() error {
	ctx := context.Background()

	// Start the container.
	pgContainer, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithInitScripts(filepath.Join("../sql", "init.sql")),
		postgres.WithDatabase("kadane-test"),
		postgres.WithUsername("kadane"),
		postgres.WithPassword("kadane"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container host: %v", err)
	}
	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return fmt.Errorf("failed to get container port: %v", err)
	}

	// Build the connection string.
	testDBConn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		"kadane",
		"kadane",
		host,
		port.Int(),
		"kadane-test",
	)

	// Parse and create a new connection pool.
	config, err := pgxpool.ParseConfig(testDBConn)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}
	testPool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	DBPool = testPool

	// Save the teardown function.
	teardownFunc = func() {
		if DBPool != nil {
			DBPool.Close()
		}
		_ = pgContainer.Terminate(ctx)
	}

	return nil
}

// TeardownContainer calls the saved teardown function.
func TeardownContainer() {
	if teardownFunc != nil {
		teardownFunc()
		teardownFunc = nil
	}
}
