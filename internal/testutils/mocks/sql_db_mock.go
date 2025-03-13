package mocks

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

// MockSQLDB is a mock implementation of *sql.DB
type MockSQLDB struct {
	*sql.DB // Embed the actual type to implement the interface

	// Function replacements
	PingContextFunc     func(ctx context.Context) error
	ExecContextFunc     func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContextFunc    func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContextFunc func(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContextFunc  func(ctx context.Context, query string) (*sql.Stmt, error)
	BeginTxFunc         func(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	ConnFunc            func(ctx context.Context) (*sql.Conn, error)

	// Tracking for method calls
	SetMaxOpenConnsCallCount    int
	SetMaxOpenConnsArgs         []int
	SetMaxIdleConnsCallCount    int
	SetMaxIdleConnsArgs         []int
	SetConnMaxLifetimeCallCount int
	SetConnMaxLifetimeArgs      []time.Duration
}

// NewMockSQLDB creates a new MockSQLDB instance
func NewMockSQLDB() *MockSQLDB {
	return &MockSQLDB{
		SetMaxOpenConnsArgs:    make([]int, 0),
		SetMaxIdleConnsArgs:    make([]int, 0),
		SetConnMaxLifetimeArgs: make([]time.Duration, 0),
	}
}

// PingContext mocks the PingContext method
func (m *MockSQLDB) PingContext(ctx context.Context) error {
	if m.PingContextFunc != nil {
		return m.PingContextFunc(ctx)
	}
	return nil
}

// ExecContext mocks the ExecContext method
func (m *MockSQLDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.ExecContextFunc != nil {
		return m.ExecContextFunc(ctx, query, args...)
	}
	return &SQLResultMock{}, nil
}

// QueryContext mocks the QueryContext method
func (m *MockSQLDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.QueryContextFunc != nil {
		return m.QueryContextFunc(ctx, query, args...)
	}
	return nil, nil
}

// QueryRowContext mocks the QueryRowContext method
func (m *MockSQLDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.QueryRowContextFunc != nil {
		return m.QueryRowContextFunc(ctx, query, args...)
	}
	return nil
}

// PrepareContext mocks the PrepareContext method
func (m *MockSQLDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if m.PrepareContextFunc != nil {
		return m.PrepareContextFunc(ctx, query)
	}
	return nil, nil
}

// BeginTx mocks the BeginTx method
func (m *MockSQLDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx, opts)
	}
	return nil, nil
}

// Conn mocks the Conn method
func (m *MockSQLDB) Conn(ctx context.Context) (*sql.Conn, error) {
	if m.ConnFunc != nil {
		return m.ConnFunc(ctx)
	}
	return nil, nil
}

// SetMaxOpenConns tracks calls to SetMaxOpenConns
func (m *MockSQLDB) SetMaxOpenConns(n int) {
	m.SetMaxOpenConnsCallCount++
	m.SetMaxOpenConnsArgs = append(m.SetMaxOpenConnsArgs, n)
}

// SetMaxIdleConns tracks calls to SetMaxIdleConns
func (m *MockSQLDB) SetMaxIdleConns(n int) {
	m.SetMaxIdleConnsCallCount++
	m.SetMaxIdleConnsArgs = append(m.SetMaxIdleConnsArgs, n)
}

// SetConnMaxLifetime tracks calls to SetConnMaxLifetime
func (m *MockSQLDB) SetConnMaxLifetime(d time.Duration) {
	m.SetConnMaxLifetimeCallCount++
	m.SetConnMaxLifetimeArgs = append(m.SetConnMaxLifetimeArgs, d)
}

// Driver returns nil to satisfy the required interface
func (m *MockSQLDB) Driver() driver.Driver {
	return nil
}
