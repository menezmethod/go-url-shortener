package mocks

import (
	"database/sql"
)

// SQLResultMock provides a mock implementation of sql.Result
type SQLResultMock struct {
	LastInsertIDFunc func() (int64, error)
	RowsAffectedFunc func() (int64, error)
}

// LastInsertId mocks the LastInsertId method
func (m *SQLResultMock) LastInsertId() (int64, error) {
	if m.LastInsertIDFunc != nil {
		return m.LastInsertIDFunc()
	}
	return 0, nil
}

// RowsAffected mocks the RowsAffected method
func (m *SQLResultMock) RowsAffected() (int64, error) {
	if m.RowsAffectedFunc != nil {
		return m.RowsAffectedFunc()
	}
	return 0, nil
}

// SQLRowMock provides a mock implementation of common.Scanner
type SQLRowMock struct {
	ScanFunc func(dest ...interface{}) error
}

// Scan mocks the Scan method
func (m *SQLRowMock) Scan(dest ...interface{}) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}

// SQLRowsMock provides a mock implementation of sql.Rows
type SQLRowsMock struct {
	CloseFunc       func() error
	ColumnTypesFunc func() ([]*sql.ColumnType, error)
	ColumnsFunc     func() ([]string, error)
	ErrFunc         func() error
	NextFunc        func() bool
	ScanFunc        func(dest ...interface{}) error
}

// Close mocks the Close method
func (m *SQLRowsMock) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// ColumnTypes mocks the ColumnTypes method
func (m *SQLRowsMock) ColumnTypes() ([]*sql.ColumnType, error) {
	if m.ColumnTypesFunc != nil {
		return m.ColumnTypesFunc()
	}
	return nil, nil
}

// Columns mocks the Columns method
func (m *SQLRowsMock) Columns() ([]string, error) {
	if m.ColumnsFunc != nil {
		return m.ColumnsFunc()
	}
	return nil, nil
}

// Err mocks the Err method
func (m *SQLRowsMock) Err() error {
	if m.ErrFunc != nil {
		return m.ErrFunc()
	}
	return nil
}

// Next mocks the Next method
func (m *SQLRowsMock) Next() bool {
	if m.NextFunc != nil {
		return m.NextFunc()
	}
	return false
}

// Scan mocks the Scan method
func (m *SQLRowsMock) Scan(dest ...interface{}) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}

// Make SQLRowWrapperMock implement the same behavior as sql.Row but use our mock
type SQLRowWrapperMock struct {
	Mock *SQLRowMock
}

// Scan delegates to the mock's Scan method
func (w *SQLRowWrapperMock) Scan(dest ...interface{}) error {
	return w.Mock.Scan(dest...)
}
