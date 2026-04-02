/*
Copyright © 2026 rtsp-recorder contributors

Package logger provides structured logging using zerolog library.
*/
package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

// Logger is the global zerolog logger instance.
// Initialized by New() and accessible throughout the application.
var Logger zerolog.Logger

// New initializes the global Logger with the specified log level and color settings.
// Per D-78: Uses zerolog instead of zap.
// Per D-79/D-80: ConsoleWriter for TTY, JSON for non-TTY.
// Per D-81: Auto-detects TTY using go-isatty.
// Per D-89: Respects NO_COLOR environment variable.
// Per D-88: noColor parameter can force disable colors.
//
// Valid log levels: "debug", "info", "warn", "error", "fatal", "panic"
func New(logLevel string, noColor bool) error {
	level, err := ParseLevel(logLevel)
	if err != nil {
		return err
	}

	// Check NO_COLOR environment variable (per D-89)
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}

	// Detect if stdout is a TTY (per D-81)
	isTTY := isatty.IsTerminal(os.Stdout.Fd())

	if isTTY && !noColor {
		// Console output with colors (per D-79)
		// Use ConsoleWriter for human-readable colored output
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05", // Human-readable time (agent's discretion per D-83-D-85)
			NoColor:    false,
		}
		Logger = zerolog.New(output).Level(level).With().Timestamp().Logger()
	} else {
		// JSON output for non-TTY or when colors disabled (per D-80)
		Logger = zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
	}

	return nil
}

// ParseLevel converts a log level string to zerolog.Level.
// Valid levels: "debug", "info", "warn", "error", "fatal", "panic"
// Returns an error for invalid levels.
func ParseLevel(level string) (zerolog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("invalid log level: %q (valid: debug, info, warn, error, fatal, panic)", level)
	}
}
