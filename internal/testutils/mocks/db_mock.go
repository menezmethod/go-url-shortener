package mocks

import (
	"database/sql"

	"github.com/menezmethod/ref_go/internal/common"
)

// DBMock provides a mock implementation of database operations for testing
type DBMock struct {
	ExecFunc               func(query string, args ...interface{}) (sql.Result, error)
	QueryFunc              func(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowFunc           func(query string, args ...interface{}) common.Scanner
	BeginFunc              func() (*sql.Tx, error)
	PrepareFunc            func(query string) (*sql.Stmt, error)
	PingFunc               func() error
	CloseFunc              func() error
	SetMaxOpenConnsFunc    func(n int)
	SetMaxIdleConnsFunc    func(n int)
	SetConnMaxLifetimeFunc func(d interface{})
}

// Exec mocks the database Exec function
func (m *DBMock) Exec(query string, args ...interface{}) (sql.Result, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(query, args...)
	}
	return nil, nil
}

// Query mocks the database Query function
func (m *DBMock) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(query, args...)
	}
	return nil, nil
}

// QueryRow mocks the database QueryRow function
func (m *DBMock) QueryRow(query string, args ...interface{}) common.Scanner {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(query, args...)
	}
	return nil
}

// Begin mocks the database Begin function
func (m *DBMock) Begin() (*sql.Tx, error) {
	if m.BeginFunc != nil {
		return m.BeginFunc()
	}
	return nil, nil
}

// Prepare mocks the database Prepare function
func (m *DBMock) Prepare(query string) (*sql.Stmt, error) {
	if m.PrepareFunc != nil {
		return m.PrepareFunc(query)
	}
	return nil, nil
}

// Ping mocks the database Ping function
func (m *DBMock) Ping() error {
	if m.PingFunc != nil {
		return m.PingFunc()
	}
	return nil
}

// Close mocks the database Close function
func (m *DBMock) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// SetMaxOpenConns mocks the database SetMaxOpenConns function
func (m *DBMock) SetMaxOpenConns(n int) {
	if m.SetMaxOpenConnsFunc != nil {
		m.SetMaxOpenConnsFunc(n)
	}
}

// SetMaxIdleConns mocks the database SetMaxIdleConns function
func (m *DBMock) SetMaxIdleConns(n int) {
	if m.SetMaxIdleConnsFunc != nil {
		m.SetMaxIdleConnsFunc(n)
	}
}

// SetConnMaxLifetime mocks the database SetConnMaxLifetime function
func (m *DBMock) SetConnMaxLifetime(d interface{}) {
	if m.SetConnMaxLifetimeFunc != nil {
		m.SetConnMaxLifetimeFunc(d)
	}
}
