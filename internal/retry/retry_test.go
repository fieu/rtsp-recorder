/*
Copyright © 2026 rtsp-recorder contributors

Unit tests for retry package.
*/
package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"rtsp-recorder/config"
	rrerrors "rtsp-recorder/internal/errors"
)

// TestRetry_SuccessFirstAttempt verifies no retries when operation succeeds.
func TestRetry_SuccessFirstAttempt(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}

	cfg := RetryConfig{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
	}

	ctx := context.Background()
	err := Retry(ctx, cfg, operation)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

// TestRetry_SuccessAfterFailures verifies retries work and eventually succeed.
func TestRetry_SuccessAfterFailures(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("network error")
		}
		return nil
	}

	cfg := RetryConfig{
		MaxAttempts: 3,
		Delay:       50 * time.Millisecond,
		ShouldRetry: func(err error) bool { return true },
	}

	ctx := context.Background()
	err := Retry(ctx, cfg, operation)

	if err != nil {
		t.Errorf("expected no error after retries, got: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

// TestRetry_ExhaustedAttempts verifies all retries fail returns last error.
func TestRetry_ExhaustedAttempts(t *testing.T) {
	callCount := 0
	expectedErr := errors.New("persistent network error")
	operation := func() error {
		callCount++
		return expectedErr
	}

	cfg := RetryConfig{
		MaxAttempts: 3,
		Delay:       50 * time.Millisecond,
		ShouldRetry: func(err error) bool { return true },
	}

	ctx := context.Background()
	err := Retry(ctx, cfg, operation)

	if err == nil {
		t.Error("expected error after exhausted attempts, got nil")
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got: %v", expectedErr, err)
	}
}

// TestRetry_NonRetryableError verifies non-retryable errors fail immediately.
func TestRetry_NonRetryableError(t *testing.T) {
	callCount := 0
	authErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.AuthenticationError,
		Message:   "[ERROR] Authentication required",
		Retryable: false,
	}

	operation := func() error {
		callCount++
		return authErr
	}

	cfg := RetryConfig{
		MaxAttempts: 3,
		Delay:       50 * time.Millisecond,
		ShouldRetry: func(err error) bool {
			if classified, ok := err.(*rrerrors.ClassifiedError); ok {
				return classified.Retryable
			}
			return false
		},
	}

	ctx := context.Background()
	err := Retry(ctx, cfg, operation)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("expected immediate fail (1 call), got %d calls", callCount)
	}
}

// TestRetry_ContextCancellation verifies retry stops on context cancellation.
func TestRetry_ContextCancellation(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		return errors.New("network error")
	}

	cfg := RetryConfig{
		MaxAttempts: 10, // Would normally take forever
		Delay:       500 * time.Millisecond,
		ShouldRetry: func(err error) bool { return true },
	}

	// Cancel context after 100ms
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := Retry(ctx, cfg, operation)

	if err == nil {
		t.Error("expected error from context cancellation, got nil")
	}
	if callCount < 1 {
		t.Errorf("expected at least 1 call, got %d", callCount)
	}
	if ctx.Err() == nil {
		t.Error("expected context to be cancelled")
	}
}

// TestRetry_CallbacksInvoked verifies OnRetry and OnFailure are called correctly.
func TestRetry_CallbacksInvoked(t *testing.T) {
	var onRetryCalls []int
	var onFailureCalled bool

	operation := func() error {
		return errors.New("always fails")
	}

	cfg := RetryConfig{
		MaxAttempts: 2,
		Delay:       50 * time.Millisecond,
		ShouldRetry: func(err error) bool { return true },
		OnRetry: func(attempt, maxAttempts int, delay time.Duration) {
			onRetryCalls = append(onRetryCalls, attempt)
		},
		OnFailure: func(attempts int, lastErr error) error {
			onFailureCalled = true
			return fmt.Errorf("custom failure after %d attempts", attempts)
		},
	}

	ctx := context.Background()
	err := Retry(ctx, cfg, operation)

	if err == nil {
		t.Error("expected error")
	}
	if len(onRetryCalls) != 1 { // Only 1 retry call for 2 attempts
		t.Errorf("expected 1 OnRetry call, got %d", len(onRetryCalls))
	}
	if onRetryCalls[0] != 1 {
		t.Errorf("expected attempt=1 in OnRetry, got %d", onRetryCalls[0])
	}
	if !onFailureCalled {
		t.Error("expected OnFailure to be called")
	}
	if err.Error() != "custom failure after 2 attempts" {
		t.Errorf("expected custom failure message, got: %v", err)
	}
}

// TestRetry_NetworkErrorRetryable verifies NetworkError triggers retry.
func TestRetry_NetworkErrorRetryable(t *testing.T) {
	callCount := 0
	networkErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.NetworkError,
		Message:   "[ERROR] Connection refused",
		Retryable: true,
	}

	operation := func() error {
		callCount++
		if callCount == 1 {
			return networkErr
		}
		return nil
	}

	cfg := RetryConfig{
		MaxAttempts: 3,
		Delay:       50 * time.Millisecond,
		ShouldRetry: func(err error) bool {
			if classified, ok := err.(*rrerrors.ClassifiedError); ok {
				return classified.Retryable
			}
			return false
		},
	}

	ctx := context.Background()
	err := Retry(ctx, cfg, operation)

	if err != nil {
		t.Errorf("expected success after retry, got: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (1 fail + 1 success), got %d", callCount)
	}
}

// TestRetry_ZeroMaxAttempts verifies 0 max attempts defaults to 1 try.
func TestRetry_ZeroMaxAttempts(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}

	cfg := RetryConfig{
		MaxAttempts: 0, // Should still attempt at least once
		Delay:       50 * time.Millisecond,
	}

	ctx := context.Background()
	Retry(ctx, cfg, operation)

	if callCount != 1 {
		t.Errorf("expected 1 call even with MaxAttempts=0, got %d", callCount)
	}
}

// TestRetry_AttemptCounting verifies attempt counting is correct (1-indexed).
func TestRetry_AttemptCounting(t *testing.T) {
	var attempts []int

	operation := func() error {
		return errors.New("fail")
	}

	cfg := RetryConfig{
		MaxAttempts: 3,
		Delay:       10 * time.Millisecond,
		ShouldRetry: func(err error) bool { return true },
		OnRetry: func(attempt, max int, delay time.Duration) {
			attempts = append(attempts, attempt)
		},
	}

	ctx := context.Background()
	Retry(ctx, cfg, operation)

	if len(attempts) != 2 {
		t.Errorf("expected 2 retry callbacks (for attempts 1 and 2), got %d", len(attempts))
	}
	if attempts[0] != 1 || attempts[1] != 2 {
		t.Errorf("expected attempts [1, 2], got %v", attempts)
	}
}

// TestDefaultRetryConfig verifies DefaultRetryConfig uses config values.
func TestDefaultRetryConfig(t *testing.T) {
	cfg := &config.Config{
		RetryAttempts: 5,
	}

	rc := DefaultRetryConfig(cfg, nil)

	if rc.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", rc.MaxAttempts)
	}
	if rc.Delay != 5*time.Second {
		t.Errorf("expected Delay=5s, got %v", rc.Delay)
	}
	if rc.ShouldRetry == nil {
		t.Error("expected ShouldRetry callback to be set")
	}
	if rc.OnRetry == nil {
		t.Error("expected OnRetry callback to be set")
	}
	if rc.OnFailure == nil {
		t.Error("expected OnFailure callback to be set")
	}
}

// TestDefaultRetryConfig_DefaultValue verifies default when RetryAttempts is 0.
func TestDefaultRetryConfig_DefaultValue(t *testing.T) {
	cfg := &config.Config{
		RetryAttempts: 0,
	}

	rc := DefaultRetryConfig(cfg, nil)

	if rc.MaxAttempts != 3 {
		t.Errorf("expected default MaxAttempts=3, got %d", rc.MaxAttempts)
	}
}

// TestDefaultShouldRetry_NetworkPatterns verifies network patterns trigger retry.
func TestDefaultShouldRetry_NetworkPatterns(t *testing.T) {
	patterns := []string{
		"connection refused",
		"timeout occurred",
		"no route to host",
		"network is unreachable",
		"broken pipe",
		"connection reset by peer",
		"connection closed",
		"i/o timeout",
	}

	for _, pattern := range patterns {
		err := errors.New("Error: " + pattern)
		if !defaultShouldRetry(err) {
			t.Errorf("expected retry for pattern '%s'", pattern)
		}
	}
}

// TestDefaultShouldRetry_NonNetworkErrors verifies non-network errors don't retry.
func TestDefaultShouldRetry_NonNetworkErrors(t *testing.T) {
	errStrings := []string{
		"file not found",
		"permission denied",
		"invalid argument",
		"authentication failed",
	}

	for _, errStr := range errStrings {
		err := errors.New(errStr)
		if defaultShouldRetry(err) {
			t.Errorf("expected no retry for '%s'", errStr)
		}
	}
}

// TestDefaultShouldRetry_ClassifiedErrors verifies classified errors use Retryable flag.
func TestDefaultShouldRetry_ClassifiedErrors(t *testing.T) {
	// NetworkError should retry
	networkErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.NetworkError,
		Retryable: true,
	}
	if !defaultShouldRetry(networkErr) {
		t.Error("expected NetworkError to be retryable")
	}

	// AuthenticationError should not retry
	authErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.AuthenticationError,
		Retryable: false,
	}
	if defaultShouldRetry(authErr) {
		t.Error("expected AuthenticationError to not be retryable")
	}

	// StreamError should not retry
	streamErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.StreamError,
		Retryable: false,
	}
	if defaultShouldRetry(streamErr) {
		t.Error("expected StreamError to not be retryable")
	}
}
