package logger

import (
	"github.com/menezmethod/ref_go/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new logger instance based on the environment
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	var logLevel zapcore.Level

	// Set log level based on environment
	switch cfg.Server.Environment {
	case "production":
		logLevel = zapcore.InfoLevel
	case "development":
		logLevel = zapcore.DebugLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	// Create appropriate zap configuration
	var zapConfig zap.Config
	if cfg.Server.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapConfig.Level = zap.NewAtomicLevelAt(logLevel)

	return zapConfig.Build()
}

// RequestLogger creates a logger with request details
func RequestLogger(baseLogger *zap.Logger, requestID string) *zap.Logger {
	return baseLogger.With(
		zap.String("request_id", requestID),
	)
}
