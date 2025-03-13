package mocks

import (
	"io"
)

// MockDatabaseDriver is a mock implementation of migrate/database.Driver
type MockDatabaseDriver struct {
	OpenFunc       func() error
	CloseFunc      func() error
	LockFunc       func() error
	UnlockFunc     func() error
	RunFunc        func(migration io.Reader) error
	SetVersionFunc func(version int, dirty bool) error
	VersionFunc    func() (version int, dirty bool, err error)
	DropFunc       func() error
}

// Open is a mock implementation of Driver.Open
func (m *MockDatabaseDriver) Open(url string) (io.ReadCloser, error) {
	if m.OpenFunc != nil {
		return nil, m.OpenFunc()
	}
	return nil, nil
}

// Close is a mock implementation of Driver.Close
func (m *MockDatabaseDriver) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Lock is a mock implementation of Driver.Lock
func (m *MockDatabaseDriver) Lock() error {
	if m.LockFunc != nil {
		return m.LockFunc()
	}
	return nil
}

// Unlock is a mock implementation of Driver.Unlock
func (m *MockDatabaseDriver) Unlock() error {
	if m.UnlockFunc != nil {
		return m.UnlockFunc()
	}
	return nil
}

// Run is a mock implementation of Driver.Run
func (m *MockDatabaseDriver) Run(migration io.Reader) error {
	if m.RunFunc != nil {
		return m.RunFunc(migration)
	}
	return nil
}

// SetVersion is a mock implementation of Driver.SetVersion
func (m *MockDatabaseDriver) SetVersion(version int, dirty bool) error {
	if m.SetVersionFunc != nil {
		return m.SetVersionFunc(version, dirty)
	}
	return nil
}

// Version is a mock implementation of Driver.Version
func (m *MockDatabaseDriver) Version() (version int, dirty bool, err error) {
	if m.VersionFunc != nil {
		return m.VersionFunc()
	}
	return 0, false, nil
}

// Drop is a mock implementation of Driver.Drop
func (m *MockDatabaseDriver) Drop() error {
	if m.DropFunc != nil {
		return m.DropFunc()
	}
	return nil
}

// Scan is a mock implementation of sql.Scanner.Scan
func (m *MockDatabaseDriver) Scan(src interface{}) error {
	return nil
}
