package common

import (
	"database/sql"
)

// Scanner is an interface for the Scan method that both sql.Row and sql.Rows implement
type Scanner interface {
	Scan(dest ...interface{}) error
}

// DB is an interface for database operations
type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) Scanner
	Begin() (*sql.Tx, error)
	Prepare(query string) (*sql.Stmt, error)
	Ping() error
	Close() error
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)
	SetConnMaxLifetime(d interface{})
}
