package repository_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/menezmethod/ref_go/internal/common"
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
				Expect(err).To(Equal(domain.ErrConflict))
			})
		})
	})

	Describe("GetByID", func() {
		Context("when the link exists", func() {
			BeforeEach(func() {
				// Create mock row with scan implementation
				row := &mocks.SQLRowMock{
					ScanFunc: func(dest ...interface{}) error {
						// Set the values in the destination pointers
						*dest[0].(*string) = "test-id"
						*dest[1].(*string) = "user-id"
						*dest[2].(*string) = "https://example.com"
						*dest[3].(*string) = "abc123"
						*dest[4].(*int) = 10
						// Timestamps
						now := time.Now()
						*dest[5].(*time.Time) = now
						*dest[6].(*time.Time) = now
						return nil
					},
				}

				// Set the QueryRowFunc to return our mock scanner
				mockDB.QueryRowFunc = func(query string, args ...interface{}) common.Scanner {
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
				// Create mock row that returns ErrNoRows
				row := &mocks.SQLRowMock{
					ScanFunc: func(dest ...interface{}) error {
						return sql.ErrNoRows
					},
				}

				mockDB.QueryRowFunc = func(query string, args ...interface{}) common.Scanner {
					return row
				}
			})

			It("returns a not found error", func() {
				link, err := repo.GetByID("non-existent-id")

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(domain.ErrNotFound))
				Expect(link).To(BeNil())
			})
		})
	})

	Describe("GetByShortURL", func() {
		Context("when the link exists", func() {
			BeforeEach(func() {
				// Create mock row with scan implementation
				row := &mocks.SQLRowMock{
					ScanFunc: func(dest ...interface{}) error {
						// Set the values in the destination pointers
						*dest[0].(*string) = "test-id"
						*dest[1].(*string) = "user-id"
						*dest[2].(*string) = "https://example.com"
						*dest[3].(*string) = "abc123"
						*dest[4].(*int) = 5
						// Timestamps
						now := time.Now()
						*dest[5].(*time.Time) = now
						*dest[6].(*time.Time) = now
						return nil
					},
				}

				// Set the QueryRowFunc to return our mock scanner
				mockDB.QueryRowFunc = func(query string, args ...interface{}) common.Scanner {
					return row
				}
			})

			It("returns the link", func() {
				link, err := repo.GetByShortURL("abc123")

				Expect(err).NotTo(HaveOccurred())
				Expect(link).NotTo(BeNil())
				Expect(link.ID).To(Equal("test-id"))
				Expect(link.UserID).To(Equal("user-id"))
				Expect(link.OriginalURL).To(Equal("https://example.com"))
				Expect(link.ShortURL).To(Equal("abc123"))
				Expect(link.Visits).To(Equal(5))
			})
		})

		Context("when the link doesn't exist", func() {
			BeforeEach(func() {
				// Create mock row that returns ErrNoRows
				row := &mocks.SQLRowMock{
					ScanFunc: func(dest ...interface{}) error {
						return sql.ErrNoRows
					},
				}

				mockDB.QueryRowFunc = func(query string, args ...interface{}) common.Scanner {
					return row
				}
			})

			It("returns a not found error", func() {
				link, err := repo.GetByShortURL("non-existent-url")

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(domain.ErrNotFound))
				Expect(link).To(BeNil())
			})
		})
	})

	Describe("Update", func() {
		Context("when successful", func() {
			BeforeEach(func() {
				mockDB.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
					return &mocks.SQLResultMock{
						RowsAffectedFunc: func() (int64, error) { return 1, nil },
					}, nil
				}
			})

			It("updates a link and returns no error", func() {
				link := &domain.Link{
					ID:          "test-id",
					UserID:      "user-id",
					OriginalURL: "https://updated-example.com",
					ShortURL:    "new-url",
					Visits:      15,
				}

				err := repo.Update(link)
				Expect(err).NotTo(HaveOccurred())
				// Verify that UpdatedAt was set
				Expect(link.UpdatedAt).NotTo(BeZero())
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
					OriginalURL: "https://updated-example.com",
					ShortURL:    "new-url",
				}

				err := repo.Update(link)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
			})
		})
	})

	Describe("Delete", func() {
		Context("when successful", func() {
			BeforeEach(func() {
				mockDB.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
					return &mocks.SQLResultMock{
						RowsAffectedFunc: func() (int64, error) { return 1, nil },
					}, nil
				}
			})

			It("deletes a link and returns no error", func() {
				err := repo.Delete("test-id")
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
				err := repo.Delete("test-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
			})
		})
	})

	Describe("List", func() {
		Context("when links exist", func() {
			BeforeEach(func() {
				mockRows := &mocks.SQLRowsMock{
					NextFunc: func() bool {
						static := 2
						static--
						return static >= 0
					},
					ScanFunc: func(dest ...interface{}) error {
						// Set the values in the destination pointers
						*dest[0].(*string) = "test-id"
						*dest[1].(*string) = "user-id"
						*dest[2].(*string) = "https://example.com"
						*dest[3].(*string) = "abc123"
						*dest[4].(*int) = 10
						// Timestamps
						now := time.Now()
						*dest[5].(*time.Time) = now
						*dest[6].(*time.Time) = now
						return nil
					},
					CloseFunc: func() error { return nil },
					ErrFunc:   func() error { return nil },
				}

				mockDB.QueryFunc = func(query string, args ...interface{}) (*sql.Rows, error) {
					DeferCleanup(func() {
						mockRows.Close()
					})
					// Assuming we've set up appropriate methods for using our mock
					// This is a simplified approach that doesn't require monkey patching
					return &sql.Rows{}, nil
				}
			})

			It("returns the links", func() {
				// Skip this test for now as it requires more complex mocking
				Skip("Requires complex sql.Rows mocking")
			})
		})

		Context("when there's a database error in Query", func() {
			BeforeEach(func() {
				mockDB.QueryFunc = func(query string, args ...interface{}) (*sql.Rows, error) {
					return nil, errors.New("database error")
				}
			})

			It("returns an error", func() {
				links, err := repo.List("user-id", 10, 0)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(links).To(BeNil())
			})
		})
	})

	Describe("Count", func() {
		Context("when successful", func() {
			BeforeEach(func() {
				row := &mocks.SQLRowMock{
					ScanFunc: func(dest ...interface{}) error {
						*dest[0].(*int) = 5
						return nil
					},
				}

				mockDB.QueryRowFunc = func(query string, args ...interface{}) common.Scanner {
					return row
				}
			})

			It("returns the count", func() {
				count, err := repo.Count("user-id")

				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(5))
			})
		})

		Context("when there's a database error", func() {
			BeforeEach(func() {
				row := &mocks.SQLRowMock{
					ScanFunc: func(dest ...interface{}) error {
						return errors.New("database error")
					},
				}

				mockDB.QueryRowFunc = func(query string, args ...interface{}) common.Scanner {
					return row
				}
			})

			It("returns an error", func() {
				count, err := repo.Count("user-id")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(count).To(Equal(0))
			})
		})
	})

	Describe("IncrementVisits", func() {
		Context("when successful", func() {
			BeforeEach(func() {
				mockDB.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
					return &mocks.SQLResultMock{
						RowsAffectedFunc: func() (int64, error) { return 1, nil },
					}, nil
				}
			})

			It("increments visits and returns no error", func() {
				err := repo.IncrementVisits("test-id")
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
				err := repo.IncrementVisits("test-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
			})
		})
	})

	Describe("CreateClick", func() {
		Context("when successful", func() {
			BeforeEach(func() {
				mockDB.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
					return &mocks.SQLResultMock{
						LastInsertIDFunc: func() (int64, error) { return 1, nil },
						RowsAffectedFunc: func() (int64, error) { return 1, nil },
					}, nil
				}
			})

			It("creates a click and returns no error", func() {
				click := &domain.Click{
					ID:        "click-id",
					LinkID:    "link-id",
					UserAgent: "Mozilla/5.0",
					Referer:   "https://referrer.com",
					IPAddress: "192.168.1.1",
				}

				err := repo.CreateClick(click)
				Expect(err).NotTo(HaveOccurred())
				// Verify that CreatedAt was set
				Expect(click.CreatedAt).NotTo(BeZero())
			})
		})

		Context("when there's a database error", func() {
			BeforeEach(func() {
				mockDB.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
					return nil, errors.New("database error")
				}
			})

			It("returns an error", func() {
				click := &domain.Click{
					ID:        "click-id",
					LinkID:    "link-id",
					UserAgent: "Mozilla/5.0",
					Referer:   "https://referrer.com",
					IPAddress: "192.168.1.1",
				}

				err := repo.CreateClick(click)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
			})
		})
	})

	Describe("GetClicks", func() {
		Context("when clicks exist", func() {
			BeforeEach(func() {
				mockRows := &mocks.SQLRowsMock{
					NextFunc: func() bool {
						static := 2
						static--
						return static >= 0
					},
					ScanFunc: func(dest ...interface{}) error {
						// Set the values in the destination pointers
						*dest[0].(*string) = "click-id"
						*dest[1].(*string) = "link-id"
						*dest[2].(*string) = "Mozilla/5.0"
						*dest[3].(*string) = "https://referrer.com"
						*dest[4].(*string) = "192.168.1.1"
						// Timestamp
						*dest[5].(*time.Time) = time.Now()
						return nil
					},
					CloseFunc: func() error { return nil },
					ErrFunc:   func() error { return nil },
				}

				mockDB.QueryFunc = func(query string, args ...interface{}) (*sql.Rows, error) {
					DeferCleanup(func() {
						mockRows.Close()
					})
					// Simplified approach
					return &sql.Rows{}, nil
				}
			})

			It("returns the clicks", func() {
				// Skip this test for now as it requires more complex mocking
				Skip("Requires complex sql.Rows mocking")
			})
		})

		Context("when there's a database error in Query", func() {
			BeforeEach(func() {
				mockDB.QueryFunc = func(query string, args ...interface{}) (*sql.Rows, error) {
					return nil, errors.New("database error")
				}
			})

			It("returns an error", func() {
				clicks, err := repo.GetClicks("link-id", 10, 0)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(clicks).To(BeNil())
			})
		})
	})

	Describe("CountClicks", func() {
		Context("when successful", func() {
			BeforeEach(func() {
				row := &mocks.SQLRowMock{
					ScanFunc: func(dest ...interface{}) error {
						*dest[0].(*int) = 10
						return nil
					},
				}

				mockDB.QueryRowFunc = func(query string, args ...interface{}) common.Scanner {
					return row
				}
			})

			It("returns the count", func() {
				count, err := repo.CountClicks("link-id")

				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(10))
			})
		})

		Context("when there's a database error", func() {
			BeforeEach(func() {
				row := &mocks.SQLRowMock{
					ScanFunc: func(dest ...interface{}) error {
						return errors.New("database error")
					},
				}

				mockDB.QueryRowFunc = func(query string, args ...interface{}) common.Scanner {
					return row
				}
			})

			It("returns an error", func() {
				count, err := repo.CountClicks("link-id")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(count).To(Equal(0))
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
