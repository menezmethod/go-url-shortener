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

		BeforeEach(func() {
			// Set up environment variables
			os.Setenv("LOG_LEVEL", "debug")

			// Create a minimal config for testing
			cfg = &config.Config{
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

				zapLogger, err := logger.NewLogger(cfg)

				Expect(err).NotTo(HaveOccurred())
				Expect(zapLogger).NotTo(BeNil())
			})
		})

		Context("with different environments", func() {
			It("defaults to info level for other environments", func() {
				cfg.Server.Environment = "staging"

				zapLogger, err := logger.NewLogger(cfg)

				Expect(err).NotTo(HaveOccurred())
				Expect(zapLogger).NotTo(BeNil())
			})
		})
	})
})
