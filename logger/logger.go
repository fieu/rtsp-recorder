/*
Copyright © 2026 rtsp-recorder contributors

Package logger provides structured logging using Uber's zap library.
*/
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap logger with the specified log level.
// The logger uses development configuration for human-readable console output.
// Valid log levels: "debug", "info", "warn", "error"
func New(logLevel string) (*zap.Logger, error) {
	level, err := ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	// Use development config for human-readable console output (per D-62)
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(level)

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

// ParseLevel converts a log level string to a zapcore.Level.
// Valid levels: "debug", "info", "warn", "error"
// Returns an error for invalid levels.
func ParseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("invalid log level: %q (valid: debug, info, warn, error)", level)
	}
}
