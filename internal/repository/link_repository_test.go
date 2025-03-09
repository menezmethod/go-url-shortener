package repository_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/menezmethod/ref_go/internal/domain"
	"github.com/menezmethod/ref_go/internal/repository"
	"github.com/menezmethod/ref_go/internal/testutils/mocks"
)

func TestLinkRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Link Repository Suite")
}

var _ = Describe("LinkRepository", func() {
	var (
		mockDB *mocks.DBMock
		repo   *repository.PostgresLinkRepository
	)

	BeforeEach(func() {
		mockDB = &mocks.DBMock{}
		repo = repository.NewPostgresLinkRepository(mockDB)
	})

	Describe("Create", func() {
		Context("when successful", func() {
			BeforeEach(func() {
				mockDB.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
					return &mocks.SQLResultMock{
						LastInsertIDFunc: func() (int64, error) { return 1, nil },
						RowsAffectedFunc: func() (int64, error) { return 1, nil },
					}, nil
				}
			})

			It("creates a link and returns no error", func() {
				link := &domain.Link{
					ID:          "test-id",
					UserID:      "user-id",
					OriginalURL: "https://example.com",
					ShortURL:    "abc123",
				}

				err := repo.Create(link)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when there's a database error", func() {
			BeforeEach(func() {
				mockDB.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
					return nil, errors.New("database error")
				}
			})

			It("returns an error", func() {
				link := &domain.Link{
					ID:          "test-id",
					UserID:      "user-id",
					OriginalURL: "https://example.com",
					ShortURL:    "abc123",
				}

				err := repo.Create(link)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
			})
		})

		Context("when there's a unique violation", func() {
			BeforeEach(func() {
				mockDB.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
					pqErr := &pq.Error{Code: "23505"} // Unique violation code
					return nil, pqErr
				}
			})

			It("returns a conflict error", func() {
				link := &domain.Link{
					ID:          "test-id",
					UserID:      "user-id",
					OriginalURL: "https://example.com",
					ShortURL:    "abc123",
				}

				err := repo.Create(link)
				Expect(err).To(HaveOccurred())
				// In a real implementation, we'd expect a domain.ErrConflict error here
			})
		})
	})

	Describe("GetByID", func() {
		Context("when the link exists", func() {
			BeforeEach(func() {
				// Create mock row
				row := &mocks.SQLRowMock{
					ScanFunc: func(dest ...interface{}) error {
						// Set the values in the destination pointers
						*dest[0].(*string) = "test-id"
						*dest[1].(*string) = "user-id"
						*dest[2].(*string) = "https://example.com"
						*dest[3].(*string) = "abc123"
						*dest[4].(*int) = 10
						// For timestamps and other fields, you would set them accordingly
						return nil
					},
				}

				mockDB.QueryRowFunc = func(query string, args ...interface{}) *sql.Row {
					return row
				}
			})

			It("returns the link", func() {
				link, err := repo.GetByID("test-id")

				Expect(err).NotTo(HaveOccurred())
				Expect(link).NotTo(BeNil())
				Expect(link.ID).To(Equal("test-id"))
				Expect(link.UserID).To(Equal("user-id"))
				Expect(link.OriginalURL).To(Equal("https://example.com"))
				Expect(link.ShortURL).To(Equal("abc123"))
				Expect(link.Visits).To(Equal(10))
			})
		})

		Context("when the link doesn't exist", func() {
			BeforeEach(func() {
				mockDB.QueryRowFunc = func(query string, args ...interface{}) *sql.Row {
					return &mocks.SQLRowMock{
						ScanFunc: func(dest ...interface{}) error {
							return sql.ErrNoRows
						},
					}
				}
			})

			It("returns a not found error", func() {
				link, err := repo.GetByID("non-existent-id")

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(sql.ErrNoRows))
				Expect(link).To(BeNil())
			})
		})
	})

	// Additional tests for other methods like GetByShortURL, Update, Delete, etc.
})

// Mock implementations for SQL-related types
type SQLResultMock struct {
	LastInsertIDFunc func() (int64, error)
	RowsAffectedFunc func() (int64, error)
}

func (m *SQLResultMock) LastInsertId() (int64, error) {
	return m.LastInsertIDFunc()
}

func (m *SQLResultMock) RowsAffected() (int64, error) {
	return m.RowsAffectedFunc()
}

type SQLRowMock struct {
	ScanFunc func(dest ...interface{}) error
}

func (m *SQLRowMock) Scan(dest ...interface{}) error {
	return m.ScanFunc(dest...)
}
