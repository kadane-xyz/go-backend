// main_test.go
package api

import (
	"log"
	"os"
	"testing"

	"kadane.xyz/go-backend/v2/src/db"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

var firebaseToken middleware.FirebaseTokenInfo = middleware.FirebaseTokenInfo{
	UserID: "test-user-id",
	Email:  "test-email",
	Name:   "test-name",
}

var handler Handler

// TestMain runs once for the entire package.
// It sets up the container, runs all tests, and tears down afterward.
func TestMain(m *testing.M) {
	if err := db.SetupTestContainer(); err != nil {
		// If setup fails, exit immediately.
		panic(err)
	}

	log.Println("firebaseToken: ", firebaseToken)

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
