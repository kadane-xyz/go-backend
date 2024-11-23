package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"kadane.xyz/go-backend/v2/src/api"
	"kadane.xyz/go-backend/v2/src/aws"
	"kadane.xyz/go-backend/v2/src/config"
	"kadane.xyz/go-backend/v2/src/db"
	"kadane.xyz/go-backend/v2/src/judge0"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"

	firebase "firebase.google.com/go/v4"
	firebaseApp "kadane.xyz/go-backend/v2/src/firebase"
)

type Server struct {
	config          *config.Config
	postgresClient  *pgxpool.Pool
	closeFunc       func()
	PostgresQueries *sql.Queries
	firebaseApp     *firebase.App
	awsClient       *s3.Client
	judge0Client    *judge0.Judge0Client
}

func NewServer(config *config.Config) (*Server, error) {
	ctx := context.Background()

	// Connect to Postgres
	postgresClient, closeFunc, err := db.NewPostgresClient(ctx, config.PostgresUrl, config.PostgresUser, config.PostgresPass, config.PostgresDB)
	if err != nil {
		return nil, fmt.Errorf("error connecting to postgres: %v", err)
	}

	_, err = postgresClient.Acquire(ctx)
	if err != nil {
		closeFunc()
		return nil, fmt.Errorf("error acquiring connection: %v", err)
	}
	log.Println("Database connection established")

	// Initialize Firebase client
	firebaseApp, err := firebaseApp.NewFirebaseApp(config)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Firebase: %v", err)
	}
	log.Println("Firebase connection established")

	// Initialize AWS client
	awsClient, err := aws.NewAWSClient(config)
	if err != nil {
		return nil, fmt.Errorf("error connecting to AWS: %v", err)
	}
	log.Println("AWS connection established")

	// Initialize Judge0 client
	judge0Client := judge0.NewJudge0Client(config)
	if judge0Client == nil {
		return nil, fmt.Errorf("failed to initialize Judge0 client")
	}
	log.Println("Judge0 initialized")

	server := &Server{
		config:         config,
		postgresClient: postgresClient,
		closeFunc:      closeFunc,
		firebaseApp:    firebaseApp,
		awsClient:      awsClient,
		judge0Client:   judge0Client,
	}

	return server, nil
}

func (s *Server) Run() error {
	//defer closing postgres connection
	defer s.closeFunc()

	//sql queries
	queries := sql.New(s.postgresClient)
	s.PostgresQueries = queries // Set queries to server

	//middleware handler
	middlewareHandler := &middleware.Handler{
		Config:      s.config,
		FirebaseApp: s.firebaseApp,
	}

	//api handler
	ApiHandler := &api.Handler{
		PostgresClient:  s.postgresClient,
		PostgresQueries: s.PostgresQueries,
		AWSClient:       s.awsClient,
		AWSBucketAvatar: s.config.AWSBucketAvatar,
		AWSRegion:       s.config.AWSRegion,
		CloudFrontUrl:   s.config.CloudFrontUrl,
		Judge0Client:    s.judge0Client,
	}

	// HTTP router
	r := chi.NewRouter()

	// Middleware
	middleware.Middleware(middlewareHandler, r)

	// HTTP routes
	api.RegisterApiRoutes(ApiHandler, r) // Api routes

	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: h2c.NewHandler(r, h2s),
	}

	// Start server
	log.Println("Starting server on :" + s.config.Port)
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("error starting server: %v", err)
	}

	return nil
}
