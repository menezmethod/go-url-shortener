package testutils

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/menezmethod/ref_go/internal/config"
)

// TestConfig returns a config used for testing
func TestConfig() (*config.Config, error) {
	// Save original environment and restore it after the test
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, envVar := range originalEnv {
			key, value, _ := splitEnvVar(envVar)
			os.Setenv(key, value)
		}
	}()

	// Set test environment variables
	// SECURITY NOTE: These are non-sensitive default values for testing only.
	// For local development, override these by setting environment variables before running tests.
	// In CI/CD environments, these should be set as secure environment variables.
	testEnvVars := map[string]string{
		"SERVER_PORT":           "8081",
		"SERVER_BASE_URL":       "http://localhost:8081",
		"SERVER_ENVIRONMENT":    "test",
		"SERVER_READ_TIMEOUT":   "10s",
		"SERVER_WRITE_TIMEOUT":  "10s",
		"SERVER_IDLE_TIMEOUT":   "60s",
		"POSTGRES_HOST":         "localhost",
		"POSTGRES_PORT":         "5432",
		"POSTGRES_USER":         "postgres",
		"POSTGRES_PASSWORD":     getEnvWithFallback("TEST_POSTGRES_PASSWORD", "dummy_password"),
		"POSTGRES_DB":           "url_shortener_test",
		"POSTGRES_SSL_MODE":     "disable",
		"POSTGRES_MAX_CONNS":    "10",
		"POSTGRES_MAX_IDLE":     "5",
		"POSTGRES_CONN_TIMEOUT": "10s",
		"MASTER_PASSWORD":       getEnvWithFallback("TEST_MASTER_PASSWORD", "dummy_master_password"),
		"JWT_SECRET":            getEnvWithFallback("TEST_JWT_SECRET", "dummy_jwt_secret"),
		"JWT_EXPIRY":            "24h",
		"LOG_LEVEL":             "debug",
		"LOG_FORMAT":            "console",
	}

	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}

	return config.LoadConfig()
}

// getEnvWithFallback returns the value of an environment variable or a fallback value if not set
func getEnvWithFallback(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// Helper function to split environment variables
func splitEnvVar(envVar string) (string, string, bool) {
	for i := 0; i < len(envVar); i++ {
		if envVar[i] == '=' {
			return envVar[:i], envVar[i+1:], true
		}
	}
	return envVar, "", false
}

// GetTestDataPath returns the absolute path to the test data directory
func GetTestDataPath() string {
	_, currentFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(currentFile), "..", "..", "testdata")
	return testdataDir
}

// EnsureTestDataDir ensures the test data directory exists
func EnsureTestDataDir() string {
	testdataDir := GetTestDataPath()
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		os.MkdirAll(testdataDir, 0755)
	}
	return testdataDir
}
