package logger_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/menezmethod/ref_go/internal/config"
	"github.com/menezmethod/ref_go/internal/logger"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Suite")
}

var _ = Describe("Logger", func() {
	Describe("NewLogger", func() {
		var cfg *config.Config
		var err error

		BeforeEach(func() {
			// Set up environment variables
			os.Setenv("LOG_LEVEL", "debug")
			os.Setenv("LOG_FORMAT", "console")
			os.Setenv("SERVER_ENVIRONMENT", "development")

			// Create a minimal config for testing
			cfg = &config.Config{
				Logging: config.LoggingConfig{
					Level:  "debug",
					Format: "console",
				},
				Server: config.ServerConfig{
					Environment: "development",
				},
			}
		})

		Context("with valid configuration", func() {
			It("creates a logger successfully", func() {
				zapLogger, err := logger.NewLogger(cfg)

				Expect(err).NotTo(HaveOccurred())
				Expect(zapLogger).NotTo(BeNil())
			})

			It("creates a production logger when environment is production", func() {
				cfg.Server.Environment = "production"
				cfg.Logging.Format = "json"

				zapLogger, err := logger.NewLogger(cfg)

				Expect(err).NotTo(HaveOccurred())
				Expect(zapLogger).NotTo(BeNil())
			})
		})

		Context("with invalid log level", func() {
			It("defaults to info level for invalid level", func() {
				cfg.Logging.Level = "invalid_level"

				zapLogger, err := logger.NewLogger(cfg)

				Expect(err).NotTo(HaveOccurred())
				Expect(zapLogger).NotTo(BeNil())
			})
		})

		Context("with invalid log format", func() {
			It("defaults to console format for invalid format", func() {
				cfg.Logging.Format = "invalid_format"

				zapLogger, err := logger.NewLogger(cfg)

				Expect(err).NotTo(HaveOccurred())
				Expect(zapLogger).NotTo(BeNil())
			})
		})
	})
})
