package mocks

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

// SQLRowMock provides a mock implementation of sql.Row
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
