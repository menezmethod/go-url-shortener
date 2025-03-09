package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Security  SecurityConfig
	RateLimit RateLimitConfig
	ShortLink ShortLinkConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         int
	BaseURL      string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxConnections  int
	MaxIdle         int
	ConnMaxLifetime time.Duration
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	MasterPassword string
	TokenExpiry    time.Duration
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// ShortLinkConfig holds URL shortener configuration
type ShortLinkConfig struct {
	DefaultExpiry time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Server config
	port, err := strconv.Atoi(getEnvOrDefault("PORT", "8081"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	cfg.Server = ServerConfig{
		Port:         port,
		BaseURL:      getEnvOrDefault("BASE_URL", fmt.Sprintf("http://localhost:%d", port)),
		Environment:  getEnvOrDefault("ENVIRONMENT", "development"),
		ReadTimeout:  parseDuration(getEnvOrDefault("READ_TIMEOUT", "30s")),
		WriteTimeout: parseDuration(getEnvOrDefault("WRITE_TIMEOUT", "30s")),
		IdleTimeout:  parseDuration(getEnvOrDefault("IDLE_TIMEOUT", "120s")),
	}

	// Database config
	dbPort, err := strconv.Atoi(getEnvOrDefault("POSTGRES_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid POSTGRES_PORT: %w", err)
	}

	maxConns, err := strconv.Atoi(getEnvOrDefault("POSTGRES_MAX_CONNECTIONS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid POSTGRES_MAX_CONNECTIONS: %w", err)
	}

	maxIdle, err := strconv.Atoi(getEnvOrDefault("POSTGRES_MAX_IDLE_CONNECTIONS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid POSTGRES_MAX_IDLE_CONNECTIONS: %w", err)
	}

	cfg.Database = DatabaseConfig{
		Host:            getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            dbPort,
		User:            getEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:        getEnv("POSTGRES_PASSWORD"),
		Database:        getEnvOrDefault("POSTGRES_DB", "url_shortener"),
		MaxConnections:  maxConns,
		MaxIdle:         maxIdle,
		ConnMaxLifetime: parseDuration(getEnvOrDefault("POSTGRES_CONN_MAX_LIFETIME", "15m")),
	}

	// Security config
	cfg.Security = SecurityConfig{
		MasterPassword: getEnv("MASTER_PASSWORD"),
		TokenExpiry:    parseDuration(getEnvOrDefault("TOKEN_EXPIRY", "24h")),
	}

	// Rate limit config
	requests, err := strconv.Atoi(getEnvOrDefault("RATE_LIMIT_REQUESTS", "60"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_REQUESTS: %w", err)
	}

	cfg.RateLimit = RateLimitConfig{
		Requests: requests,
		Window:   parseDuration(getEnvOrDefault("RATE_LIMIT_WINDOW", "60s")),
	}

	// Short link config
	cfg.ShortLink = ShortLinkConfig{
		DefaultExpiry: parseDuration(getEnvOrDefault("SHORTLINK_DEFAULT_EXPIRY", "30d")),
	}

	// Validate required configurations
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// parseDuration safely parses a duration string with a fallback
func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		// Log this in a real implementation
		return 30 * time.Second // Default fallback
	}
	return duration
}

// validateConfig ensures required fields are present
func validateConfig(cfg *Config) error {
	if cfg.Security.MasterPassword == "" {
		return fmt.Errorf("MASTER_PASSWORD is required")
	}

	return nil
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnv gets an environment variable or returns empty string
func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return ""
}
