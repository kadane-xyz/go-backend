// main_test.go
package server_test

import (
	"os"
	"testing"

	"kadane.xyz/go-backend/v2/src/db"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

var clientToken middleware.ClientContext = middleware.ClientContext{
	UserID: "123abc",
	Email:  "john@example.com",
	Name:   "John Doe",
	Plan:   sql.AccountPlanPro, // Set the account type to pro
	Admin:  true,               // Set the admin flag
}

var handler Handler

// TestMain runs once for the entire package.
// It sets up the container, runs all tests, and tears down afterward.
func TestMain(m *testing.M) {
	// Create init.sql file

	if err := db.SetupTestContainer(); err != nil {
		// If setup fails, exit immediately.
		panic(err)
	}

	// Setup the handler
	db := db.DBPool
	queries := sql.New(db)
	handler = Handler{
		PostgresClient:  db,
		PostgresQueries: queries,
	}

	// Run all tests in the package.
	exitCode := m.Run()

	// Clean up the container after all tests are done.
	db.Close()

	os.Exit(exitCode)
}
