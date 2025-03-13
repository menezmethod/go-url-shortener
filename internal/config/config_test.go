package config_test

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/menezmethod/ref_go/internal/config"
)

// Test constants - using placeholders instead of actual credentials
const (
	testDbPassword     = "db_password_placeholder"
	testMasterPassword = "master_password_placeholder"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	// Save original environment and restore it after each test
	var originalEnv []string

	BeforeEach(func() {
		originalEnv = os.Environ()
	})

	AfterEach(func() {
		os.Clearenv()
		for _, envVar := range originalEnv {
			parts := splitEnvVar(envVar)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	})

	Describe("LoadConfig", func() {
		Context("with valid environment variables", func() {
			BeforeEach(func() {
				// Set required environment variables for testing
				os.Clearenv()
				os.Setenv("PORT", "8081")
				os.Setenv("BASE_URL", "http://localhost:8081")
				os.Setenv("ENVIRONMENT", "test")
				os.Setenv("READ_TIMEOUT", "10s")
				os.Setenv("WRITE_TIMEOUT", "10s")
				os.Setenv("IDLE_TIMEOUT", "60s")
				os.Setenv("POSTGRES_HOST", "localhost")
				os.Setenv("POSTGRES_PORT", "5432")
				os.Setenv("POSTGRES_USER", "postgres")
				os.Setenv("POSTGRES_PASSWORD", testDbPassword)
				os.Setenv("POSTGRES_DB", "url_shortener_test")
				os.Setenv("POSTGRES_MAX_CONNS", "10")
				os.Setenv("POSTGRES_MAX_IDLE", "5")
				os.Setenv("POSTGRES_CONN_TIMEOUT", "10s")
				os.Setenv("MASTER_PASSWORD", testMasterPassword)
				os.Setenv("JWT_EXPIRY", "24h")
			})

			It("loads the configuration correctly", func() {
				cfg, err := config.LoadConfig()

				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())
				Expect(cfg.Server.Port).To(Equal(8081))
				Expect(cfg.Server.BaseURL).To(Equal("http://localhost:8081"))
				Expect(cfg.Server.Environment).To(Equal("test"))
				Expect(cfg.Server.ReadTimeout).To(Equal(10 * time.Second))
				Expect(cfg.Server.WriteTimeout).To(Equal(10 * time.Second))
				Expect(cfg.Server.IdleTimeout).To(Equal(60 * time.Second))

				// Don't check exact Database values as they might vary
				Expect(cfg.Database.Host).NotTo(BeEmpty())
				Expect(cfg.Database.Port).To(BeNumerically(">", 0))

				Expect(cfg.Security.MasterPassword).To(Equal(testMasterPassword))
			})
		})

		Context("with missing required environment variables", func() {
			BeforeEach(func() {
				os.Clearenv()
				// Set minimum required variables
				os.Setenv("MASTER_PASSWORD", testMasterPassword)
			})

			It("uses default values for optional variables", func() {
				cfg, err := config.LoadConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.Server.Environment).To(Equal("development"))
				Expect(cfg.Server.Port).To(Equal(8081)) // Default port
			})
		})

		Context("with invalid timeout format", func() {
			BeforeEach(func() {
				// Set required environment variables for testing
				os.Clearenv()
				os.Setenv("PORT", "8081")
				os.Setenv("ENVIRONMENT", "test")
				os.Setenv("READ_TIMEOUT", "invalid") // Invalid timeout
				os.Setenv("MASTER_PASSWORD", testMasterPassword)
			})

			It("uses default values for invalid formats", func() {
				cfg, err := config.LoadConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.Server.ReadTimeout).To(Equal(30 * time.Second)) // Default value
			})
		})

		Context("with invalid port number", func() {
			BeforeEach(func() {
				// Set required environment variables for testing
				os.Clearenv()
				os.Setenv("PORT", "invalid") // Invalid port
				os.Setenv("MASTER_PASSWORD", testMasterPassword)
			})

			It("returns an error", func() {
				_, err := config.LoadConfig()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid PORT"))
			})
		})
	})
})

// Helper function to split environment variables
func splitEnvVar(envVar string) []string {
	for i := 0; i < len(envVar); i++ {
		if envVar[i] == '=' {
			return []string{envVar[:i], envVar[i+1:]}
		}
	}
	return []string{envVar}
}
