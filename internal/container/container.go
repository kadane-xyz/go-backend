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
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

// APIHandlers is an inner container holding all API handlers
type APIHandlers struct {
	AccountHandler    *handlers.AccountHandler
	AdminHandler      *handlers.AdminHandler
	ProblemHandler    *handlers.ProblemHandler
	CommentHandler    *handlers.CommentHandler
	FriendHandler     *handlers.FriendHandler
	RunHandler        *handlers.RunHandler
	SubmissionHandler *handlers.SubmissionHandler
	SolutionsHandler  *handlers.SolutionsHandler
	StarredHandler    *handlers.StarredHandler
}

type Repositories struct {
	AccountRepo    repository.AccountRepository
	AdminRepo      repository.AdminRepository
	ProblemRepo    repository.ProblemsRepository
	CommentRepo    repository.CommentsRepository
	FriendRepo     repository.FriendRepository
	SubmissionRepo repository.SubmissionsRepository
	SolutionRepo   repository.SolutionsRepository
	StarredRepo    repository.StarredRepository
}

// Container is a service locator holding all application dependencies
type Container struct {
	Config       *config.Config
	DB           *pgxpool.Pool
	AWSClient    *s3.Client
	Judge0       *judge0.Judge0Client
	SqlQueries   *sql.Queries
	APIHandlers  *APIHandlers
	Repositories *Repositories
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
	accountsRepo := repository.NewSQLAccountsRepository(queries)
	adminRepo := repository.NewSQLAdminRepository(queries)
	problemsRepo := repository.NewSQLProblemsRepository(queries)
	commentsRepo := repository.NewSQLCommentsRepository(queries)
	friendsRepo := repository.NewSQLFriendRepository(queries)
	submissionsRepo := repository.NewSQLSubmissionsRepository(queries)
	solutionsRepo := repository.NewSQLSolutionsRepository(queries)
	starredRepo := repository.NewSQLStarredRepository(queries)

	// Create API handlers
	apiHandlers := handlers.NewAccountHandler(accountsRepo, awsClient, cfg)
	problemHandler := handlers.NewProblemHandler(problemsRepo)
	adminHandler := handlers.NewAdminHandler(adminRepo, judge0Client, problemHandler)
	commentHandler := handlers.NewCommentHandler(commentsRepo, solutionsRepo)
	friendHandler := handlers.NewFriendHandler(friendsRepo, accountsRepo)
	runHandler := handlers.NewRunHandler(problemsRepo, judge0Client)
	submissionHandler := handlers.NewSubmissionHandler(submissionsRepo)
	solutionsHandler := handlers.NewSolutionsHandler(solutionsRepo)
	starredHandler := handlers.NewStarredHandler(starredRepo)

	return &Container{
		Config:     cfg,
		DB:         dbPool,
		SqlQueries: queries,
		AWSClient:  awsClient,
		Judge0:     judge0Client,
		APIHandlers: &APIHandlers{
			AccountHandler:    apiHandlers,
			AdminHandler:      adminHandler,
			ProblemHandler:    problemHandler,
			CommentHandler:    commentHandler,
			FriendHandler:     friendHandler,
			RunHandler:        runHandler,
			SubmissionHandler: submissionHandler,
			SolutionsHandler:  solutionsHandler,
			StarredHandler:    starredHandler,
		},
		Repositories: &Repositories{
			AccountRepo:    accountsRepo,
			AdminRepo:      adminRepo,
			CommentRepo:    commentsRepo,
			ProblemRepo:    problemsRepo,
			FriendRepo:     friendsRepo,
			SubmissionRepo: submissionsRepo,
			SolutionRepo:   solutionsRepo,
			StarredRepo:    starredRepo,
		},
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
