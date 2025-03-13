package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/menezmethod/ref_go/internal/config"
)

// WithInstance is a variable that allows us to mock postgres.WithInstance in tests
var WithInstance = func(instance interface{}, config map[string]string) (interface{}, error) {
	if sqlDB, ok := instance.(*sql.DB); ok {
		return postgres.WithInstance(sqlDB, &postgres.Config{})
	}
	return nil, fmt.Errorf("invalid database instance")
}

// NewWithDatabaseInstance is a variable that allows us to mock migrate.NewWithDatabaseInstance in tests
var NewWithDatabaseInstance = func(sourceName string, databaseName string, driverInstance interface{}) (interface{}, error) {
	if driver, ok := driverInstance.(database.Driver); ok {
		return migrate.NewWithDatabaseInstance(sourceName, databaseName, driver)
	}
	return nil, fmt.Errorf("invalid driver instance")
}

// MigrateDatabase applies all migrations to the database
func MigrateDatabase(db *sql.DB, cfg *config.Config) error {
	// Create a new migrate instance
	driver, err := WithInstance(db, nil)
	if err != nil {
		return fmt.Errorf("creating postgres driver: %w", err)
	}

	// Provide the path to migrations
	m, err := NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", "./migrations/postgres"),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	// Cast to the actual migrate.Migrate type or use our interface
	var migrateInstance interface {
		Up() error
	}

	if migrator, ok := m.(*migrate.Migrate); ok {
		migrateInstance = migrator
	} else {
		migrateInstance = m.(interface {
			Up() error
		})
	}

	// Apply migrations
	if err := migrateInstance.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("applying migrations: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}

// CheckMigrations verifies that migrations are up to date
func CheckMigrations(db *sql.DB) (bool, error) {
	driver, err := WithInstance(db, nil)
	if err != nil {
		return false, fmt.Errorf("creating postgres driver: %w", err)
	}

	m, err := NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", "./migrations/postgres"),
		"postgres", driver)
	if err != nil {
		return false, fmt.Errorf("creating migrate instance: %w", err)
	}

	// Cast to the actual migrate.Migrate type or use our interface
	var migrateInstance interface {
		Version() (uint, bool, error)
	}

	if migrator, ok := m.(*migrate.Migrate); ok {
		migrateInstance = migrator
	} else {
		migrateInstance = m.(interface {
			Version() (uint, bool, error)
		})
	}

	version, dirty, err := migrateInstance.Version()
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
