package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"kadane.xyz/go-backend/v2/src/api"
	"kadane.xyz/go-backend/v2/src/config"
	"kadane.xyz/go-backend/v2/src/db"
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
		closeFunc()
		return nil, fmt.Errorf("error connecting to Firebase: %v", err)
	}
	log.Println("Firebase connection established")

	return &Server{
		config:         config,
		postgresClient: postgresClient,
		closeFunc:      closeFunc,
		firebaseApp:    firebaseApp,
	}, nil
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
	}

	// HTTP router
	r := chi.NewRouter()

	// Middleware
	middleware.Middleware(middlewareHandler, r)

	// HTTP routes
	api.RegisterApiRoutes(ApiHandler, r) // Api routes

	// Start server
	log.Println("Starting server on :" + s.config.Port)
	if err := http.ListenAndServe(":"+s.config.Port, r); err != nil {
		return fmt.Errorf("error starting server: %v", err)
	}

	return nil
}
