package db_test

import (
	"fmt"
	"testing"

	"github.com/menezmethod/ref_go/internal/config"
	"github.com/menezmethod/ref_go/internal/db"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DB Suite")
}

var _ = Describe("DB", func() {
	var (
		cfg *config.Config
	)

	BeforeEach(func() {
		// Create a mock config for database tests
		cfg = &config.Config{
			Database: config.DatabaseConfig{
				Host:            "localhost",
				Port:            5432,
				User:            "testuser",
				Password:        "testpassword",
				Database:        "testdb",
				MaxConnections:  10,
				MaxIdle:         5,
				ConnMaxLifetime: 3600,
			},
		}
	})

	Describe("New", func() {
		Context("when SQL.Open fails", func() {
			It("should return an error", func() {
				// Replace the sql.Open function with our mock that returns an error
				originalSQLOpen := db.SQLOpen
				defer func() { db.SQLOpen = originalSQLOpen }()

				expectedErr := fmt.Errorf("open error")
				db.SQLOpen = func(driverName, dataSourceName string) (interface{}, error) {
					Expect(driverName).To(Equal("postgres"), "Driver name should be postgres")

					// Verify the connection string contains our config values
					Expect(dataSourceName).To(ContainSubstring("host=localhost"))
					Expect(dataSourceName).To(ContainSubstring("port=5432"))
					Expect(dataSourceName).To(ContainSubstring("user=testuser"))
					Expect(dataSourceName).To(ContainSubstring("password=testpassword"))
					Expect(dataSourceName).To(ContainSubstring("dbname=testdb"))

					return nil, expectedErr
				}

				// Call the function under test
				database, err := db.New(cfg)

				// Assertions
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("opening database connection"))
				Expect(database).To(BeNil())
			})
		})
	})
})
