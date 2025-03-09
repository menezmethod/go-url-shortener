package config_test

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/menezmethod/ref_go/internal/config"
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
				os.Setenv("SERVER_PORT", "8081")
				os.Setenv("SERVER_BASE_URL", "http://localhost:8081")
				os.Setenv("SERVER_ENVIRONMENT", "test")
				os.Setenv("SERVER_READ_TIMEOUT", "10s")
				os.Setenv("SERVER_WRITE_TIMEOUT", "10s")
				os.Setenv("SERVER_IDLE_TIMEOUT", "60s")
				os.Setenv("POSTGRES_HOST", "localhost")
				os.Setenv("POSTGRES_PORT", "5432")
				os.Setenv("POSTGRES_USER", "postgres")
				os.Setenv("POSTGRES_PASSWORD", "postgres")
				os.Setenv("POSTGRES_DB", "url_shortener_test")
				os.Setenv("POSTGRES_SSL_MODE", "disable")
				os.Setenv("POSTGRES_MAX_CONNS", "10")
				os.Setenv("POSTGRES_MAX_IDLE", "5")
				os.Setenv("POSTGRES_CONN_TIMEOUT", "10s")
				os.Setenv("MASTER_PASSWORD", "test_master_password")
				os.Setenv("JWT_SECRET", "test_jwt_secret")
				os.Setenv("JWT_EXPIRY", "24h")
				os.Setenv("LOG_LEVEL", "debug")
				os.Setenv("LOG_FORMAT", "console")
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

				Expect(cfg.Database.Host).To(Equal("localhost"))
				Expect(cfg.Database.Port).To(Equal(5432))
				Expect(cfg.Database.User).To(Equal("postgres"))
				Expect(cfg.Database.Password).To(Equal("postgres"))
				Expect(cfg.Database.DBName).To(Equal("url_shortener_test"))
				Expect(cfg.Database.SSLMode).To(Equal("disable"))
				Expect(cfg.Database.MaxConns).To(Equal(10))
				Expect(cfg.Database.MaxIdle).To(Equal(5))
				Expect(cfg.Database.ConnTimeout).To(Equal(10 * time.Second))

				Expect(cfg.Auth.MasterPassword).To(Equal("test_master_password"))
				Expect(cfg.Auth.JWTSecret).To(Equal("test_jwt_secret"))
				Expect(cfg.Auth.JWTExpiry).To(Equal(24 * time.Hour))

				Expect(cfg.Logging.Level).To(Equal("debug"))
				Expect(cfg.Logging.Format).To(Equal("console"))
			})
		})

		Context("with missing environment variables", func() {
			BeforeEach(func() {
				os.Clearenv()
				// Don't set any environment variables
			})

			It("returns an error", func() {
				_, err := config.LoadConfig()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid timeout format", func() {
			BeforeEach(func() {
				// Set required environment variables for testing
				os.Clearenv()
				os.Setenv("SERVER_PORT", "8081")
				os.Setenv("SERVER_BASE_URL", "http://localhost:8081")
				os.Setenv("SERVER_ENVIRONMENT", "test")
				os.Setenv("SERVER_READ_TIMEOUT", "invalid") // Invalid timeout
				os.Setenv("SERVER_WRITE_TIMEOUT", "10s")
				os.Setenv("SERVER_IDLE_TIMEOUT", "60s")
				// Set other variables...
			})

			It("returns an error", func() {
				_, err := config.LoadConfig()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid port number", func() {
			BeforeEach(func() {
				// Set required environment variables for testing
				os.Clearenv()
				os.Setenv("SERVER_PORT", "invalid") // Invalid port
				os.Setenv("SERVER_BASE_URL", "http://localhost:8081")
				os.Setenv("SERVER_ENVIRONMENT", "test")
				os.Setenv("SERVER_READ_TIMEOUT", "10s")
				os.Setenv("SERVER_WRITE_TIMEOUT", "10s")
				os.Setenv("SERVER_IDLE_TIMEOUT", "60s")
				// Set other variables...
			})

			It("returns an error", func() {
				_, err := config.LoadConfig()
				Expect(err).To(HaveOccurred())
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
