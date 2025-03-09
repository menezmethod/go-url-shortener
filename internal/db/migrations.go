package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/menezmethod/ref_go/internal/config"
)

// MigrateDatabase applies all migrations to the database
func MigrateDatabase(db *sql.DB, cfg *config.Config) error {
	// Create a new migrate instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating postgres driver: %w", err)
	}

	// Provide the path to migrations
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", "./migrations/postgres"),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	// Apply migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("applying migrations: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}

// CheckMigrations verifies that migrations are up to date
func CheckMigrations(db *sql.DB) (bool, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return false, fmt.Errorf("creating postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", "./migrations/postgres"),
		"postgres", driver)
	if err != nil {
		return false, fmt.Errorf("creating migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		// If no migration has been applied yet, that's okay
		if errors.Is(err, migrate.ErrNilVersion) {
			return false, nil
		}
		return false, fmt.Errorf("checking migration version: %w", err)
	}

	// If the schema is dirty, migrations need to be fixed
	if dirty {
		return false, fmt.Errorf("database schema is in dirty state at version %d", version)
	}

	return true, nil
}
