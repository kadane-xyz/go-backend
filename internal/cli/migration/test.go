package migration

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Imports the postgres database driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Imports the file source driver
)

// used for integration testing
func MigrateUp(databaseURL string, migrationsPath string) error {
	mi, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}
	defer mi.Close()

	return Up(mi, -1)
}
