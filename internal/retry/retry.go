/*
Copyright © 2026 rtsp-recorder contributors

Retry orchestration logic for network errors.
Provides configurable retry with backoff for transient failures.

This package follows D-30 through D-34 from Phase 3 context:
- Retry on NetworkError category errors
- Fixed 5-second delay between attempts (D-32)
- Uses config.RetryAttempts for max attempts (D-31)
- Re-attempts full Record() call with fresh ffmpeg process (D-34)
*/
package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"rtsp-recorder/config"
	rrerrors "rtsp-recorder/internal/errors"
)

// RetryConfig configures the retry behavior.
type RetryConfig struct {
	MaxAttempts int           // Maximum number of retry attempts
	Delay       time.Duration // Fixed delay between attempts
	ShouldRetry func(error) bool
	OnRetry     func(attempt int, maxAttempts int, delay time.Duration)
	OnFailure   func(attempts int, lastErr error) error
	Logger      zerolog.Logger // Logger for structured retry logging
}

// Retry executes the operation with configured retry logic.
// It attempts the operation up to MaxAttempts times, with the specified
// delay between attempts. If ShouldRetry returns false for an error,
// retry stops immediately and the error is returned.
//
// Context cancellation is checked between attempts.
// The last error is preserved and returned when all attempts are exhausted.
func Retry(ctx context.Context, cfg RetryConfig, operation func() error) error {
	// Ensure at least one attempt
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	// Use background context if none provided
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Execute the operation
		if err := operation(); err != nil {
			lastErr = err

			// Check if we should retry this error
			if cfg.ShouldRetry != nil && !cfg.ShouldRetry(err) {
				// Non-retryable error - fail immediately
				if cfg.OnFailure != nil {
					return cfg.OnFailure(attempt, lastErr)
				}
				return lastErr
			}

			// If this was the last attempt, we're done
			if attempt >= maxAttempts {
				break
			}

			// Notify about retry
			if cfg.OnRetry != nil {
				cfg.OnRetry(attempt, maxAttempts, cfg.Delay)
			}

			// Wait for delay, checking context cancellation
			select {
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled: %w", ctx.Err())
			case <-time.After(cfg.Delay):
				// Continue to next attempt
			}
		} else {
			// Success!
			return nil
		}
	}

	// All attempts exhausted
	if cfg.OnFailure != nil {
		return cfg.OnFailure(maxAttempts, lastErr)
	}

	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// DefaultRetryConfig creates a RetryConfig from the application Config.
// It uses cfg.RetryAttempts for MaxAttempts (default 3 per D-05).
// Fixed 5-second delay per D-32.
func DefaultRetryConfig(cfg *config.Config, logger zerolog.Logger) RetryConfig {
	maxAttempts := cfg.RetryAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3 // Default per D-05
	}

	return RetryConfig{
		MaxAttempts: maxAttempts,
		Delay:       5 * time.Second, // Fixed 5s delay per D-32
		ShouldRetry: defaultShouldRetry,
		OnRetry:     defaultOnRetry(logger),
		OnFailure:   defaultOnFailure,
		Logger:      logger,
	}
}

// defaultShouldRetry checks if an error should trigger a retry.
// Returns true only for NetworkError category (D-30, D-33).
func defaultShouldRetry(err error) bool {
	// Check if this is a classified error
	if classified, ok := err.(*rrerrors.ClassifiedError); ok {
		return rrerrors.IsRetryable(classified.Category)
	}

	// For unclassified errors, check the error message for network patterns
	errStr := err.Error()
	networkPatterns := []string{
		"connection refused",
		"timeout",
		"no route to host",
		"network is unreachable",
		"broken pipe",
		"connection reset",
		"connection closed",
		"i/o timeout",
	}

	for _, pattern := range networkPatterns {
		// Simple case-insensitive check
		if containsIgnoreCase(errStr, pattern) {
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if str contains substr (case-insensitive).
func containsIgnoreCase(str, substr string) bool {
	if len(str) < len(substr) {
		return false
	}
	// Simple case folding: convert both to lower case
	strLower := toLower(str)
	substrLower := toLower(substr)
	return contains(strLower, substrLower)
}

// toLower converts ASCII string to lowercase.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// contains checks if str contains substr.
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// defaultOnRetry provides the default retry notification.
// Per D-40: "[INFO] Retry 1/3 after 5s..."
// Per D-71: Use Warn level for retries (recoverable issues)
func defaultOnRetry(logger zerolog.Logger) func(attempt, maxAttempts int, delay time.Duration) {
	return func(attempt, maxAttempts int, delay time.Duration) {
		logger.Warn().
			Int("attempt", attempt).
			Int("max_attempts", maxAttempts).
			Dur("delay", delay).
			Msg("Retrying after error")
	}
}

// defaultOnFailure provides the default failure handler.
// Returns an error with attempt count and root cause.
func defaultOnFailure(attempts int, lastErr error) error {
	return fmt.Errorf("[ERROR] Recording failed after %d attempts: %w", attempts, lastErr)
}
