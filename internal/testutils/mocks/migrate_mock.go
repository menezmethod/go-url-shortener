package mocks

import (
	"github.com/golang-migrate/migrate/v4"
)

// MockMigrate is a mock implementation of the migrate.Migrate type
type MockMigrate struct {
	UpFunc      func() error
	DownFunc    func() error
	StepFunc    func(n int) error
	VersionFunc func() (uint, bool, error)
	CloseFunc   func() error
}

// Up is a mock implementation of migrate.Migrate.Up
func (m *MockMigrate) Up() error {
	if m.UpFunc != nil {
		return m.UpFunc()
	}
	return nil
}

// Down is a mock implementation of migrate.Migrate.Down
func (m *MockMigrate) Down() error {
	if m.DownFunc != nil {
		return m.DownFunc()
	}
	return nil
}

// Step is a mock implementation of migrate.Migrate.Step
func (m *MockMigrate) Step(n int) error {
	if m.StepFunc != nil {
		return m.StepFunc(n)
	}
	return nil
}

// Version is a mock implementation of migrate.Migrate.Version
func (m *MockMigrate) Version() (uint, bool, error) {
	if m.VersionFunc != nil {
		return m.VersionFunc()
	}
	return 0, false, nil
}

// Close is a mock implementation of migrate.Migrate.Close
func (m *MockMigrate) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Drop is a mock implementation of migrate.Migrate.Drop
func (m *MockMigrate) Drop() error {
	return nil
}

// Force is a mock implementation of migrate.Migrate.Force
func (m *MockMigrate) Force(version int) error {
	return nil
}

// SetGlobalLogger is a mock implementation of migrate.SetGlobalLogger
func SetGlobalLogger(logger migrate.Logger) {
	// No-op for testing
}

// GotoVersion is a mock implementation of migrate.Migrate.GotoVersion
func (m *MockMigrate) GotoVersion(version uint) error {
	return nil
}
