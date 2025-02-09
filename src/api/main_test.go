// main_test.go
package api

import (
	"os"
	"os/exec"
	"testing"

	"kadane.xyz/go-backend/v2/src/db"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

var firebaseToken middleware.FirebaseTokenInfo = middleware.FirebaseTokenInfo{
	UserID: "123abc",
	Email:  "john@example.com",
	Name:   "John Doe",
}

var handler Handler

// TestMain runs once for the entire package.
// It sets up the container, runs all tests, and tears down afterward.
func TestMain(m *testing.M) {
	// Create init.sql file
	wd, err := os.Getwd() // Get the working directory
	if err != nil {
		panic(err)
	}
	os.Chdir(wd + "/src/sql")     // Change to the sql directory
	exec.Command("./init-sql.sh") // Run the init-db.sh script
	os.Chdir(wd)                  // Reset the working directory

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
