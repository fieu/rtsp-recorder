/*
Copyright © 2026 rtsp-recorder contributors

Error classification and message formatting for ffmpeg failures.
Provides meaningful, actionable error messages for common failure modes.

This package follows D-39 through D-44 from Phase 3 context:
- Error pattern detection from ffmpeg stderr
- Classification into categories: NetworkError, AuthenticationError, StreamError, etc.
- Actionable error messages that guide users to fix the problem
*/
package errors

import (
	"fmt"
	"regexp"
	"strings"
)

// ErrorCategory represents the classification of an error.
type ErrorCategory string

const (
	// NetworkError — Connection, timeout, route issues (retryable per D-30)
	NetworkError ErrorCategory = "network"

	// AuthenticationError — 401, 403 (not retryable)
	AuthenticationError ErrorCategory = "auth"

	// StreamError — Invalid data, codec issues, 404 not found (not retryable)
	StreamError ErrorCategory = "stream"

	// ConfigurationError — Invalid URL, missing fields (not retryable)
	ConfigurationError ErrorCategory = "config"

	// FFmpegError — Internal ffmpeg failures (not retryable)
	FFmpegError ErrorCategory = "ffmpeg"
)

// ClassifiedError contains the classification and formatted message for an error.
type ClassifiedError struct {
	Category  ErrorCategory
	Message   string
	Retryable bool
	Original  error
	Stderr    string
	ExitCode  int
}

// Error implements the error interface for ClassifiedError.
func (e *ClassifiedError) Error() string {
	return e.Message
}

// ClassifyError analyzes ffmpeg stderr output and exit code to classify the error.
// Returns a ClassifiedError with category, actionable message, and retryability.
//
// Per D-39 through D-41, D-44:
//   - Detects common error patterns from stderr
//   - Maps patterns to specific error categories
//   - Provides actionable messages per D-40
func ClassifyError(stderr string, exitCode int) *ClassifiedError {
	stderrLower := strings.ToLower(stderr)

	// Define error patterns and their classifications
	// Patterns are checked in order of specificity

	// 1. Authentication errors (401/403) - not retryable
	if matchesAny(stderrLower,
		"401 unauthorized",
		"403 forbidden",
		"authentication required",
		"invalid credentials",
		"unauthorized",
		"forbidden",
	) {
		return &ClassifiedError{
			Category:  AuthenticationError,
			Message:   "[ERROR] Authentication required. Check username/password in URL.",
			Retryable: false,
			Stderr:    stderr,
			ExitCode:  exitCode,
		}
	}

	// 2. 404 Not Found - stream path error - not retryable
	if matchesAny(stderrLower,
		"404 not found",
		"not found",
	) {
		return &ClassifiedError{
			Category:  StreamError,
			Message:   "[ERROR] Stream path not found. Verify the RTSP URL path.",
			Retryable: false,
			Stderr:    stderr,
			ExitCode:  exitCode,
		}
	}

	// 3. Invalid data / stream format errors - not retryable
	if matchesAny(stderrLower,
		"invalid data found when processing input",
		"invalid data",
		"corrupt input",
		"unsupported codec",
		"codec not currently supported in container",
	) {
		return &ClassifiedError{
			Category:  StreamError,
			Message:   "[ERROR] Stream data invalid. Camera may be offline or incompatible.",
			Retryable: false,
			Stderr:    stderr,
			ExitCode:  exitCode,
		}
	}

	// 4. Network unreachable / no route to host - retryable
	if matchesAny(stderrLower,
		"no route to host",
		"network is unreachable",
		"host unreachable",
	) {
		return &ClassifiedError{
			Category:  NetworkError,
			Message:   "[ERROR] Network unreachable. Check network connectivity.",
			Retryable: true,
			Stderr:    stderr,
			ExitCode:  exitCode,
		}
	}

	// 5. Connection refused - retryable
	if matchesAny(stderrLower,
		"connection refused",
		"refused",
	) {
		return &ClassifiedError{
			Category:  NetworkError,
			Message:   "[ERROR] Cannot connect to camera. Check IP address and port.",
			Retryable: true,
			Stderr:    stderr,
			ExitCode:  exitCode,
		}
	}

	// 6. Operation timed out - retryable
	if matchesAny(stderrLower,
		"operation timed out",
		"timeout",
		"time out",
		"i/o timeout",
		"connection timed out",
	) {
		return &ClassifiedError{
			Category:  NetworkError,
			Message:   "[ERROR] Connection timeout. Camera may be offline or behind firewall.",
			Retryable: true,
			Stderr:    stderr,
			ExitCode:  exitCode,
		}
	}

	// 7. Broken pipe / connection reset - retryable
	if matchesAny(stderrLower,
		"broken pipe",
		"connection reset",
		"connection closed",
	) {
		return &ClassifiedError{
			Category:  NetworkError,
			Message:   "[ERROR] Connection interrupted. Camera may have disconnected.",
			Retryable: true,
			Stderr:    stderr,
			ExitCode:  exitCode,
		}
	}

	// 8. Configuration errors - invalid URL - not retryable
	if matchesAny(stderrLower,
		"invalid argument",
		"no such file or directory",
		"protocol not found",
		"unknown protocol",
	) {
		return &ClassifiedError{
			Category:  ConfigurationError,
			Message:   "[ERROR] Invalid configuration. Check RTSP URL and output path.",
			Retryable: false,
			Stderr:    stderr,
			ExitCode:  exitCode,
		}
	}

	// Exit code 255 often indicates network/connection issues - retryable
	if exitCode == 255 {
		return &ClassifiedError{
			Category:  NetworkError,
			Message:   "[ERROR] Network connection failed. Check camera availability.",
			Retryable: true,
			Stderr:    stderr,
			ExitCode:  exitCode,
		}
	}

	// Default: generic ffmpeg error - not retryable
	return &ClassifiedError{
		Category:  FFmpegError,
		Message:   fmt.Sprintf("[ERROR] FFmpeg failed (exit %d): %s", exitCode, truncateStderr(stderr, 100)),
		Retryable: false,
		Stderr:    stderr,
		ExitCode:  exitCode,
	}
}

// FormatErrorMessage returns a formatted error message for the classified error.
// This is a convenience wrapper that returns the Message field with [ERROR] prefix.
func FormatErrorMessage(classified *ClassifiedError) string {
	if classified == nil {
		return ""
	}
	return classified.Message
}

// IsRetryable returns true if the error category allows retry attempts.
// Per D-30: Only NetworkError is retryable
func IsRetryable(category ErrorCategory) bool {
	return category == NetworkError
}

// GetCategoryDescription returns a human-readable description of the error category.
func GetCategoryDescription(category ErrorCategory) string {
	switch category {
	case NetworkError:
		return "Network connectivity issue (may be temporary)"
	case AuthenticationError:
		return "Authentication failure (check credentials)"
	case StreamError:
		return "Stream data or path issue (not accessible)"
	case ConfigurationError:
		return "Configuration or URL issue (invalid setup)"
	case FFmpegError:
		return "Internal FFmpeg error (unexpected failure)"
	default:
		return "Unknown error category"
	}
}

// matchesAny checks if the text contains any of the given patterns.
// Patterns are matched as substrings (case-insensitive via pre-lowercased text).
func matchesAny(text string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// truncateStderr limits stderr output to maxLen characters.
// If truncated, it adds "..." at the end.
func truncateStderr(stderr string, maxLen int) string {
	if len(stderr) <= maxLen {
		return stderr
	}
	if maxLen <= 3 {
		return stderr[:maxLen]
	}
	return stderr[:maxLen-3] + "..."
}

// ExtractBitrate attempts to extract the current bitrate from ffmpeg stderr.
// Returns the bitrate string if found (e.g., "1024kbits/s"), or empty string.
// This implements D-43 for progress display accuracy.
func ExtractBitrate(stderr string) string {
	// Look for bitrate pattern: "bitrate: 1234.5kbits/s" or "rate: 1234 kbits/s"
	// Common patterns in ffmpeg output:
	// - "bitrate= 1234.5kbits/s"
	// - "rate= 1024 kbits/s"
	// - Size/time lines often contain bitrate info

	re := regexp.MustCompile(`bitrate[=:]\s*(\d+\.?\d*)\s*([kmg]?)bits?/s`)
	matches := re.FindStringSubmatch(stderr)

	if len(matches) >= 3 {
		value := matches[1]
		unit := matches[2]
		return fmt.Sprintf("%s%sbit/s", value, strings.ToUpper(unit))
	}

	return ""
}

// ParseFFmpegErrors scans ffmpeg stderr for multiple error patterns.
// Returns a slice of ClassifiedErrors found in the stderr.
// Useful for analyzing long recordings with multiple issues.
func ParseFFmpegErrors(stderr string) []*ClassifiedError {
	var errors []*ClassifiedError

	// Split by common error delimiters
	lines := strings.Split(stderr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this line indicates an error
		if classified := ClassifyError(line, -1); classified.Category != FFmpegError {
			// Found a specific error pattern
			errors = append(errors, classified)
		}
	}

	return errors
}
