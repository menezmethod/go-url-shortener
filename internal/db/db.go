package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/menezmethod/ref_go/internal/config"
)

// DB represents a database connection
type DB struct {
	*sql.DB
}

// SQLOpen is a variable that allows us to mock sql.Open in tests
var SQLOpen = func(driverName, dataSourceName string) (interface{}, error) {
	return sql.Open(driverName, dataSourceName)
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	// Construct connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
	)

	// Open connection
	db, err := SQLOpen("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening database connection: %w", err)
	}

	// Convert back to *sql.DB
	sqlDB := db.(*sql.DB)

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.Database.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &DB{sqlDB}, nil
}

// HealthCheck checks database connectivity
func (db *DB) HealthCheck(ctx context.Context) error {
	return db.PingContext(ctx)
}
