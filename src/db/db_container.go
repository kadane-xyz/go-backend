// db_container.go
package db

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// DBPool is the global connection pool used by all tests.
var DBPool *pgxpool.Pool

// PGContainer is the global container used by all tests.
var PGContainer testcontainers.Container

// teardownFunc holds the function to shut down the container.
var teardownFunc func()

// SetupTestContainer starts the PostgreSQL container and creates the connection pool.
// It also initializes the database by executing the init-sql.sh script.
func SetupTestContainer() error {
	ctx := context.Background()

	// Get current working directory.
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Printf("SetupTestContainer: current working directory: %s\n", wd)

	// Calculate the SQL directory relative to the current working directory.
	// If tests are run from /repository-root/src/db then "../sql" points to /repository-root/src/sql.
	sqlDir := filepath.Join("..", "sql")
	fullSQLDir := filepath.Join(wd, sqlDir)
	fmt.Printf("Changing directory to SQL directory: %s\n", fullSQLDir)
	if err := os.Chdir(fullSQLDir); err != nil {
		panic(fmt.Sprintf("Failed to change directory to SQL directory %s: %v", fullSQLDir, err))
	}

	// Execute the initialization script using bash and capture combined output.
	cmd := exec.Command("bash", "./init-sql.sh")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("Error executing init-sql.sh: %v\nOutput: %s", err, output))
	}
	fmt.Printf("init-sql.sh output:\n%s\n", output)

	// Return to the original working directory.
	if err := os.Chdir(wd); err != nil {
		panic(fmt.Sprintf("Failed to return to working directory %s: %v", wd, err))
	}

	// Start the PostgreSQL container.
	pgContainer, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithInitScripts(filepath.Join(sqlDir, "init.sql")),
		postgres.WithDatabase("kadane-test"),
		postgres.WithUsername("kadane"),
		postgres.WithPassword("kadane"),
		postgres.WithSQLDriver("pgx"),
		// Wait strategy for container log.
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	PGContainer = pgContainer

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

	// Parse configuration and initialize connection pool.
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
