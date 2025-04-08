package container

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"kadane.xyz/go-backend/v2/internal/api/handlers"
	"kadane.xyz/go-backend/v2/internal/aws"
	"kadane.xyz/go-backend/v2/internal/config"
	"kadane.xyz/go-backend/v2/internal/database"
	"kadane.xyz/go-backend/v2/internal/database/dbaccessors"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/judge0"
	"kadane.xyz/go-backend/v2/internal/services"
)

// APIHandlers is an inner container holding all API handlers
type APIHandlers struct {
	AccountHandler *handlers.AccountHandler
	AdminHandler   *handlers.AdminHandler
	ProblemHandler *handlers.ProblemHandler
}

// Container is a service locator holding all application dependencies
type Container struct {
	Config      *config.Config
	DB          *pgxpool.Pool
	AWSClient   *s3.Client
	Judge0      *judge0.Judge0Client
	SqlQueries  *sql.Queries
	APIHandlers *APIHandlers
	Services    services.ServiceAccessor
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

	// Create the AWS client
	awsClient, err := aws.NewAWSClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create the S3 client: %w", err)
	}

	// Create the Judge0 client
	judge0Client := judge0.NewJudge0Client(cfg)

	// Create queries client
	queries := sql.New(dbPool)

	// Create database accessors
	accountsAccessor := dbaccessors.NewSQLAccountsAccessor(queries)
	adminAccessor := dbaccessors.NewSQLAdminAccessor(queries)
	problemsAccessor := dbaccessors.NewSQLProblemsAccessor(queries)

	// Create services1
	accountService := services.NewAccountService(accountsAccessor)
	adminService := services.NewAdminService(adminAccessor)
	problemService := services.NewProblemService(problemsAccessor)

	// Create API handlers
	apiHandlers := handlers.NewAccountHandler(accountService, awsClient, cfg)
	adminHandler := handlers.NewAdminHandler(adminService, problemService, judge0Client)
	problemHandler := handlers.NewProblemHandler(problemService)

	services := services.NewServiceAccessor(accountsAccessor)

	return &Container{
		Config:     cfg,
		DB:         dbPool,
		SqlQueries: queries,
		AWSClient:  awsClient,
		Judge0:     judge0Client,
		APIHandlers: &APIHandlers{
			AccountHandler: apiHandlers,
			AdminHandler:   adminHandler,
			ProblemHandler: problemHandler,
		},
		Services: services,
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
