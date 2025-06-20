package integration_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"kadane.xyz/go-backend/v2/internal/cli/migration"
	"kadane.xyz/go-backend/v2/internal/config"
	"kadane.xyz/go-backend/v2/internal/container"
	"kadane.xyz/go-backend/v2/internal/database"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/middleware"
	"kadane.xyz/go-backend/v2/internal/server"
)

type txCtxKey string

const txKey txCtxKey = "test_tx"

type TestContainer struct {
	*container.Container
}

var (
	PGContainer      *postgres.PostgresContainer // PGContainer is the global PostgreSQL container used by all tests
	TestingContainer *TestContainer              // MockContainer holds a service locator instance for use by the integration tests
	TestingServer    *server.Server              // MockServer is a global reference to the mock server used by all tests
	teardownFunc     func()                      // teardownFunc holds the function to shut down the container

	//errFailedTestAssertion = errors.New("failed test assertion")
)

var clientToken middleware.ClientContext = middleware.ClientContext{
	UserID: "123abc",
	Email:  "john@example.com",
	Name:   "John Doe",
	Plan:   sql.AccountPlanPro, // Set the account type to pro
	Admin:  true,               // Set the admin flag
}

func TestMain(m *testing.M) {
	var err error

	testingContainer, err := setupTestEnv()
	if err != nil {
		log.Fatalf("failed to setup testing environment: %v", err)
	}

	TestingContainer = &TestContainer{
		Container: testingContainer,
	}

	TestingServer = server.New(TestingContainer.Container)
	if err != nil {
		log.Fatalf("failed to create testing server: %v", err)
	}

	exitCode := m.Run()

	teardownContainer()

	os.Exit(exitCode)
}

func setupTestEnv() (*container.Container, error) {
	ctx := context.Background()

	// Start the PostgreSQL container.
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:17",
		postgres.WithDatabase("kadane-test"),
		postgres.WithUsername("kadane"),
		postgres.WithPassword("kadane"),
		postgres.WithSQLDriver("pgx"),
		// Wait strategy for container log.
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second),
			wait.ForListeningPort("5432/tcp"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	PGContainer = pgContainer

	// Build the connection string.
	testDBConn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get the Postgres container connection string: %w", err)
	}

	dbPoolCfg, err := pgxpool.ParseConfig(testDBConn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, dbPoolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Handle database migrations
	migrationsPath := filepath.Join("../..", "database", "migrations")
	if err := migration.MigrateUp(testDBConn, migrationsPath); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Handle database seed data
	seedDataPath := filepath.Join("../..", "database", "seed", "seed.sql")

	seed, err := os.ReadFile(seedDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed file: %w", err)
	}

	_, err = dbPool.Exec(ctx, string(seed))
	if err != nil {
		return nil, fmt.Errorf("failed to execute seed data: %w", err)
	}

	// Create services
	queries := sql.New(dbPool)

	// database transaction manager
	txManager := database.NewTransactionManager(dbPool)

	// Create repositories
	accountsRepo := repository.NewAccountRepository(queries, txManager)
	adminRepo := repository.NewSQLAdminRepository(queries)
	problemsRepo := repository.NewSQLProblemsRepository(queries)
	commentsRepo := repository.NewSQLCommentsRepository(queries)
	friendsRepo := repository.NewSQLFriendRepository(queries)
	submissionsRepo := repository.NewSQLSubmissionsRepository(queries)
	solutionsRepo := repository.NewSQLSolutionsRepository(queries)
	starredRepo := repository.NewSQLStarredRepository(queries)

	container := &container.Container{
		Config: &config.Config{
			Debug:       false,
			Environment: config.EnvTest,
			Port:        "3000",
			DatabaseURL: testDBConn,
			Firebase: config.FirebaseConfig{
				Cred: "",
			},
			AWS: config.AWSConfig{
				Key:           "",
				Secret:        "",
				BucketAvatar:  "",
				Region:        "",
				CloudFrontURL: "",
			},
			Judge0: config.Judge0Config{
				Url:   "",
				Token: "",
			},
		},
		DB:         dbPool,
		SqlQueries: queries,
		Repositories: &container.Repositories{
			AccountRepo:    accountsRepo,
			AdminRepo:      adminRepo,
			CommentRepo:    commentsRepo,
			ProblemRepo:    problemsRepo,
			FriendRepo:     friendsRepo,
			SubmissionRepo: submissionsRepo,
			SolutionRepo:   solutionsRepo,
			StarredRepo:    starredRepo,
		},
	}

	teardownFunc = func() {
		if TestingContainer != nil {
			TestingContainer.Close()
		}

		_ = pgContainer.Terminate(ctx)
	}

	return container, nil
}

// teardownContainer calls the saved teardown function.
func teardownContainer() {
	if teardownFunc != nil {
		teardownFunc()
		teardownFunc = nil
	}
}

// contextWithTimeout adds a timeout to the test context and applies its cancel function to the
// test cleanup.
func contextWithTimeout(tb testing.TB) context.Context {
	tb.Helper()
	ctx, cancel := context.WithTimeout(tb.Context(), 10*time.Second)
	tb.Cleanup(func() {
		cancel()
	})
	return ctx
}

// withTestTransaction creates a transaction for test isolation and returns a context with the transaction.
func withTestTransaction(t *testing.T) context.Context {
	t.Helper()
	ctx := contextWithTimeout(t)

	tx, err := TestingContainer.DB.Begin(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	return context.WithValue(ctx, txKey, tx)
}
