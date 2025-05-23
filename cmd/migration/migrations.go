package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Up(mi *migrate.Migrate, limit int) error {
	if limit >= 0 {
		if err := mi.Steps(limit); err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			log.Println(err)
		}
	} else {
		if err := mi.Up(); err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			log.Println(err)
		}
	}

	log.Println("Applied all up migrations")

	return nil
}

func Down(mi *migrate.Migrate, limit int) error {
	if limit >= 0 {
		if err := mi.Steps(-limit); err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			log.Println(err)
		}
	} else {
		if err := mi.Down(); err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			log.Println(err)
		}
	}

	return nil
}

func Version(mi *migrate.Migrate) error {
	version, dirty, err := mi.Version()
	if err != nil {
		return err
	}

	if dirty {
		log.Printf("%v (dirty)\n", version)
	} else {
		log.Println(version)
	}

	return nil
}

func numDownMigrationsFromArgs(applyAll bool, args []string) (int, bool, error) {
	if applyAll {
		if len(args) > 0 {
			return 0, false, errors.New("-all cannot be used with other arguments")
		}

		return -1, false, nil
	}

	switch len(args) {
	case 0:
		return -1, true, nil
	case 1:
		downStr := args[0]

		downUint64, err := strconv.ParseUint(downStr, 10, 0)
		if err != nil {
			log.Fatal(err)
		}

		if downUint64 > uint64(math.MaxInt) {
			log.Fatalf("error: number argument %v exceeds maximum integer value", downUint64)
		}

		return int(downUint64), false, nil
	default:
		return 0, false, errors.New("too many arguments")
	}
}

func seedDatabase(databaseURL, seedPath string) error {
	seedSQL, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	_, err = db.Exec(context.Background(), string(seedSQL))
	if err != nil {
		return fmt.Errorf("failed to execute seed SQL: %w", err)
	}

	log.Println("Test database seeded successfully")

	return nil
}
