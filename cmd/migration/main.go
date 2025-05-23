package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	"kadane.xyz/go-backend/v2/internal/config"
)

func main() {
	Run()
}

func handleCommandHelp(help bool, usage string, flagSet *flag.FlagSet) {
	if help {
		fmt.Fprintln(os.Stderr, usage)
		flagSet.PrintDefaults()
		os.Exit(0)
	}
}

func newFlagSetWithHelp(name string) (*flag.FlagSet, *bool) {
	flagSet := flag.NewFlagSet(name, flag.ExitOnError)
	helpPtr := flagSet.Bool("help", false, "Print help information")

	return flagSet, helpPtr
}

func Run() {
	helpFlag := flag.Bool("help", false, "")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr,
			`Usage: migrate OPTIONS COMMAND [arg...]
       migrate [ -help ]

Options:
  -help            Print usage

Commands:
  up               Apply up migrations
  down             Apply down migrations
  seed             Seed the database with testing data
  version          Print current migration version`+"\n",
		)
	}

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Migrations path
	migrationsPath := filepath.Join("database", "migrations")

	migrator, migratorErr := migrate.New(
		"file://"+migrationsPath,
		cfg.DatabaseURL,
	)

	defer func() {
		if migratorErr == nil {
			if _, err := migrator.Close(); err != nil {
				log.Println(err)
			}
		}
	}()

	if migratorErr == nil {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT)

		go func() {
			for range signals {
				log.Println("Stopping after the current migration ...")
				migrator.GracefulStop <- true

				return
			}
		}()
	}

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	args := flag.Args()[1:]

	// Parse command
	switch flag.Arg(0) {
	case "up":
		upFlagSet, helpPtr := newFlagSetWithHelp("up")

		if err := upFlagSet.Parse(args); err != nil {
			log.Fatal(err)
		}

		handleCommandHelp(*helpPtr, "up [N]  Apply all or N up migrations", upFlagSet)

		if migratorErr != nil {
			log.Fatal(migratorErr)
		}

		limit := -1

		if upFlagSet.NArg() > 0 {
			limitStr := upFlagSet.Arg(0)

			limitUint64, err := strconv.ParseUint(limitStr, 10, 0)
			if err != nil {
				log.Fatal(err)
			}

			if limitUint64 > uint64(math.MaxInt) {
				log.Fatalf("error: limit argument %v exceeds maximum integer value", limitUint64)
			}

			limit = int(limitUint64)
		}

		if err := Up(migrator, limit); err != nil {
			log.Fatal(err)
		}

	case "down":
		downFlagSet, helpPtr := newFlagSetWithHelp("down")
		applyAll := downFlagSet.Bool("all", false, "Apply all down migrations")

		if err := downFlagSet.Parse(args); err != nil {
			log.Fatal(err)
		}

		downUsage := `down [N] [-all]  Apply all or N down migrations
                 Use -all to apply all down migrations`

		handleCommandHelp(*helpPtr, downUsage, downFlagSet)

		if migratorErr != nil {
			log.Fatal(migratorErr)
		}

		downArgs := downFlagSet.Args()

		num, needsConfirm, err := numDownMigrationsFromArgs(*applyAll, downArgs)
		if err != nil {
			log.Fatal(err)
		}

		if needsConfirm {
			log.Println("Are you sure you want to apply all down migrations? [y/n]")

			var response string
			_, _ = fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response == "y" {
				log.Println("Applying all down migrations")
			} else {
				log.Fatal("Not applying all down migrations")
			}
		}

		if err := Down(migrator, num); err != nil {
			log.Fatal(err)
		}

	case "seed":
		if cfg.Environment != config.EnvDevelopment && cfg.Environment != config.EnvTest {
			log.Fatal("error: the migrate application's seed command can only be used in development or testing")
		}

		seedFlagSet, helpPtr := newFlagSetWithHelp("seed")
		pathPtr := seedFlagSet.String("path", "", "Path for the seed file to load (default: DATABASE_SEED_FILE_PATH env var or db/seed/seed.sql)")

		if err := seedFlagSet.Parse(args); err != nil {
			log.Fatal(err)
		}

		// Default seed path
		defaultSeedPath := os.Getenv("DATABASE_SEED_FILE_PATH")
		if defaultSeedPath == "" {
			defaultSeedPath = filepath.Join("db", "seed", "seed.sql")
		}

		seedUsage := `seed [-path]  Seed the database with testing data
              Specify an optional path with the -path option or the DATABASE_SEED_FILE_PATH environment variable`

		handleCommandHelp(*helpPtr, seedUsage, seedFlagSet)

		seedPath := defaultSeedPath
		if *pathPtr != "" {
			seedPath = *pathPtr
		}

		if _, err := os.Stat(seedPath); os.IsNotExist(err) {
			log.Fatal("error: seed file does not exist", seedPath)
		}

		if err := seedDatabase(cfg.DatabaseURL, seedPath); err != nil {
			log.Fatal(err)
		}

	case "version":
		if migratorErr != nil {
			log.Fatal(migratorErr)
		}

		if err := Version(migrator); err != nil {
			log.Fatal(err)
		}

	default:
		flag.Usage()
		os.Exit(2)
	}
}
