package db_test

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/menezmethod/ref_go/internal/config"
	"github.com/menezmethod/ref_go/internal/db"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migrations", func() {
	var (
		testConfig *config.Config
	)

	BeforeEach(func() {
		testConfig = &config.Config{
			Database: config.DatabaseConfig{},
		}
	})

	Describe("MigrateDatabase", func() {
		Context("when errors occur", func() {
			It("should handle the error when creating postgres driver fails", func() {
				// Monkey patch the WithInstance function
				originalWithInstance := db.WithInstance
				defer func() { db.WithInstance = originalWithInstance }()

				expectedErr := fmt.Errorf("driver error")
				db.WithInstance = func(instance interface{}, config map[string]string) (interface{}, error) {
					return nil, expectedErr
				}

				// Call the function under test with nil since we're mocking everything
				err := db.MigrateDatabase(nil, testConfig)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("creating postgres driver"))
			})

			It("should handle ErrNoChange gracefully", func() {
				// Monkey patch the WithInstance function
				originalWithInstance := db.WithInstance
				defer func() { db.WithInstance = originalWithInstance }()

				// Make it return a successful driver instance
				db.WithInstance = func(instance interface{}, config map[string]string) (interface{}, error) {
					return "mock-driver", nil
				}

				// Monkey patch the NewWithDatabaseInstance function
				originalNewWithDB := db.NewWithDatabaseInstance
				defer func() { db.NewWithDatabaseInstance = originalNewWithDB }()

				// Make it return a migrate instance that returns ErrNoChange on Up()
				db.NewWithDatabaseInstance = func(sourceName string, databaseName string, driverInstance interface{}) (interface{}, error) {
					return &mockMigrate{
						upFunc: func() error {
							return migrate.ErrNoChange
						},
					}, nil
				}

				// Call the function under test
				err := db.MigrateDatabase(nil, testConfig)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CheckMigrations", func() {
		Context("when migrations are in a dirty state", func() {
			It("should return an error", func() {
				// Monkey patch the WithInstance function
				originalWithInstance := db.WithInstance
				defer func() { db.WithInstance = originalWithInstance }()

				// Make it return a successful driver instance
				db.WithInstance = func(instance interface{}, config map[string]string) (interface{}, error) {
					return "mock-driver", nil
				}

				// Monkey patch the NewWithDatabaseInstance function
				originalNewWithDB := db.NewWithDatabaseInstance
				defer func() { db.NewWithDatabaseInstance = originalNewWithDB }()

				// Make it return a migrate instance that returns a dirty state
				db.NewWithDatabaseInstance = func(sourceName string, databaseName string, driverInstance interface{}) (interface{}, error) {
					return &mockMigrate{
						versionFunc: func() (uint, bool, error) {
							return 5, true, nil // Version 5, dirty
						},
					}, nil
				}

				// Call the function under test
				upToDate, err := db.CheckMigrations(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("dirty state"))
				Expect(err.Error()).To(ContainSubstring("5"))
				Expect(upToDate).To(BeFalse())
			})

			It("should handle ErrNilVersion correctly", func() {
				// Monkey patch the WithInstance function
				originalWithInstance := db.WithInstance
				defer func() { db.WithInstance = originalWithInstance }()

				// Make it return a successful driver instance
				db.WithInstance = func(instance interface{}, config map[string]string) (interface{}, error) {
					return "mock-driver", nil
				}

				// Monkey patch the NewWithDatabaseInstance function
				originalNewWithDB := db.NewWithDatabaseInstance
				defer func() { db.NewWithDatabaseInstance = originalNewWithDB }()

				// Make it return a migrate instance that returns ErrNilVersion
				db.NewWithDatabaseInstance = func(sourceName string, databaseName string, driverInstance interface{}) (interface{}, error) {
					return &mockMigrate{
						versionFunc: func() (uint, bool, error) {
							return 0, false, migrate.ErrNilVersion
						},
					}, nil
				}

				// Call the function under test
				upToDate, err := db.CheckMigrations(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(upToDate).To(BeFalse())
			})
		})
	})
})

// Simple mockMigrate implementation for our tests
type mockMigrate struct {
	upFunc      func() error
	versionFunc func() (uint, bool, error)
}

func (m *mockMigrate) Up() error {
	if m.upFunc != nil {
		return m.upFunc()
	}
	return nil
}

func (m *mockMigrate) Version() (uint, bool, error) {
	if m.versionFunc != nil {
		return m.versionFunc()
	}
	return 0, false, nil
}

// Other methods not used in our tests
func (m *mockMigrate) Close() error {
	return nil
}

func (m *mockMigrate) Down() error {
	return nil
}

func (m *mockMigrate) Force(version int) error {
	return errors.New("Not implemented")
}

func (m *mockMigrate) Step(n int) error {
	return errors.New("Not implemented")
}

func (m *mockMigrate) Drop() error {
	return errors.New("Not implemented")
}

func (m *mockMigrate) GotoVersion(version uint) error {
	return errors.New("Not implemented")
}
