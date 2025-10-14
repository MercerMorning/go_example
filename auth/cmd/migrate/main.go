package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/MercerMorning/go_example/auth/internal/client/db/migrate"
	"github.com/MercerMorning/go_example/auth/internal/config"
)

func main() {
	var (
		command = flag.String("command", "", "Migration command: up, down, force, version")
		version = flag.Int("version", 0, "Version for force command")
	)
	flag.Parse()

	if *command == "" {
		log.Fatal("command is required. Use: up, down, force, version")
	}

	// Get database configuration
	pgConfig, err := config.NewPGConfig()
	if err != nil {
		log.Fatalf("Failed to create PG config: %v", err)
	}

	// Get migrations path
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	migrationsPath := filepath.Join(wd, "..", "..", "migrations")

	// Create migrator
	migrator, err := migrate.NewMigrator(pgConfig.DSN(), migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer migrator.Close()

	// Execute command
	switch *command {
	case "up":
		if err := migrator.Up(); err != nil {
			log.Fatalf("Failed to run migrations up: %v", err)
		}
	case "down":
		if err := migrator.Down(); err != nil {
			log.Fatalf("Failed to run migrations down: %v", err)
		}
	case "force":
		if *version == 0 {
			log.Fatal("version is required for force command")
		}
		if err := migrator.Force(*version); err != nil {
			log.Fatalf("Failed to force migration version: %v", err)
		}
	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		if dirty {
			fmt.Printf("Current version: %d (dirty)\n", version)
		} else {
			fmt.Printf("Current version: %d\n", version)
		}
	default:
		log.Fatalf("Unknown command: %s. Use: up, down, force, version", *command)
	}
}
