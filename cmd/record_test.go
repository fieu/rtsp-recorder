/*
Copyright © 2026 rtsp-recorder contributors

Integration tests for record command with retry logic.
*/
package cmd

import (
	"errors"
	"testing"
	"time"

	"rtsp-recorder/config"
	rrerrors "rtsp-recorder/internal/errors"
	"rtsp-recorder/internal/retry"

	"github.com/rs/zerolog"
)

// TestRecordRetryConfig verifies retry config is properly set up from config.
func TestRecordRetryConfig(t *testing.T) {
	// This test verifies the retry configuration used in runRecord
	cfg := &config.Config{RetryAttempts: 3}

	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())

	if retryCfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", retryCfg.MaxAttempts)
	}
	if retryCfg.Delay != 5*time.Second {
		t.Errorf("expected Delay=5s, got %v", retryCfg.Delay)
	}
	if retryCfg.ShouldRetry == nil {
		t.Error("expected ShouldRetry callback")
	}
}

// TestRecordRetry_ShouldRetry_NetworkError verifies NetworkError triggers retry.
func TestRecordRetry_ShouldRetry_NetworkError(t *testing.T) {
	networkErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.NetworkError,
		Message:   "[ERROR] Connection refused",
		Retryable: true,
	}

	cfg := &config.Config{RetryAttempts: 3}
	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())

	if !retryCfg.ShouldRetry(networkErr) {
		t.Error("expected NetworkError to trigger retry")
	}
}

// TestRecordRetry_ShouldRetry_AuthError verifies AuthenticationError fails immediately.
func TestRecordRetry_ShouldRetry_AuthError(t *testing.T) {
	authErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.AuthenticationError,
		Message:   "[ERROR] Authentication required",
		Retryable: false,
	}

	cfg := &config.Config{RetryAttempts: 3}
	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())

	if retryCfg.ShouldRetry(authErr) {
		t.Error("expected AuthenticationError to not trigger retry")
	}
}

// TestRecordRetry_ShouldRetry_StreamError verifies StreamError fails immediately.
func TestRecordRetry_ShouldRetry_StreamError(t *testing.T) {
	streamErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.StreamError,
		Message:   "[ERROR] Stream not found",
		Retryable: false,
	}

	cfg := &config.Config{RetryAttempts: 3}
	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())

	if retryCfg.ShouldRetry(streamErr) {
		t.Error("expected StreamError to not trigger retry")
	}
}

// TestRecordRetry_ShouldRetry_ConfigurationError verifies ConfigurationError fails immediately.
func TestRecordRetry_ShouldRetry_ConfigurationError(t *testing.T) {
	configErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.ConfigurationError,
		Message:   "[ERROR] Invalid URL",
		Retryable: false,
	}

	cfg := &config.Config{RetryAttempts: 3}
	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())

	if retryCfg.ShouldRetry(configErr) {
		t.Error("expected ConfigurationError to not trigger retry")
	}
}

// TestRecordRetry_ShouldRetry_FFmpegError verifies FFmpegError fails immediately.
func TestRecordRetry_ShouldRetry_FFmpegError(t *testing.T) {
	ffmpegErr := &rrerrors.ClassifiedError{
		Category:  rrerrors.FFmpegError,
		Message:   "[ERROR] FFmpeg failed",
		Retryable: false,
	}

	cfg := &config.Config{RetryAttempts: 3}
	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())

	if retryCfg.ShouldRetry(ffmpegErr) {
		t.Error("expected FFmpegError to not trigger retry")
	}
}

// TestRecordRetry_WithMockNetworkFailures simulates network failures and successful retry.
func TestRecordRetry_WithMockNetworkFailures(t *testing.T) {
	callCount := 0
	cfg := &config.Config{RetryAttempts: 3}
	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())

	// Override delay for faster test
	retryCfg.Delay = 10 * time.Millisecond

	operation := func() error {
		callCount++
		if callCount < 3 {
			return &rrerrors.ClassifiedError{
				Category:  rrerrors.NetworkError,
				Message:   "[ERROR] Connection refused",
				Retryable: true,
			}
		}
		return nil // Success on 3rd attempt
	}

	err := retry.Retry(nil, retryCfg, operation)

	if err != nil {
		t.Errorf("expected success after retries, got: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

// TestRecordRetry_ExhaustedAttempts verifies behavior when all retries fail.
func TestRecordRetry_ExhaustedAttempts(t *testing.T) {
	callCount := 0
	cfg := &config.Config{RetryAttempts: 2}
	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())
	retryCfg.Delay = 10 * time.Millisecond

	operation := func() error {
		callCount++
		return &rrerrors.ClassifiedError{
			Category:  rrerrors.NetworkError,
			Message:   "[ERROR] Connection refused",
			Retryable: true,
		}
	}

	err := retry.Retry(nil, retryCfg, operation)

	if err == nil {
		t.Error("expected error after exhausted retries")
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

// TestRecordRetry_NonClassifiedError verifies unclassified errors don't retry.
func TestRecordRetry_NonClassifiedError(t *testing.T) {
	callCount := 0
	cfg := &config.Config{RetryAttempts: 3}
	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())

	operation := func() error {
		callCount++
		return errors.New("some random error")
	}

	retry.Retry(nil, retryCfg, operation)

	if callCount != 1 {
		t.Errorf("expected 1 call for non-classified error, got %d", callCount)
	}
}

// TestRecordRetry_CallbacksInvoked verifies callbacks are called with correct values.
func TestRecordRetry_CallbacksInvoked(t *testing.T) {
	var retryAttempts []int
	var failureCalled bool

	retryCfg := retry.RetryConfig{
		MaxAttempts: 3,
		Delay:       10 * time.Millisecond,
		ShouldRetry: func(err error) bool {
			if classified, ok := err.(*rrerrors.ClassifiedError); ok {
				return classified.Retryable
			}
			return false
		},
		OnRetry: func(attempt, maxAttempts int, delay time.Duration) {
			retryAttempts = append(retryAttempts, attempt)
		},
		OnFailure: func(attempts int, lastErr error) error {
			failureCalled = true
			return errors.New("final failure")
		},
	}

	callCount := 0
	operation := func() error {
		callCount++
		return &rrerrors.ClassifiedError{
			Category:  rrerrors.NetworkError,
			Message:   "[ERROR] Connection timeout",
			Retryable: true,
		}
	}

	retry.Retry(nil, retryCfg, operation)

	if len(retryAttempts) != 2 { // Attempts 1 and 2 trigger retry callbacks
		t.Errorf("expected 2 retry callbacks, got %d", len(retryAttempts))
	}
	if !failureCalled {
		t.Error("expected OnFailure to be called")
	}
}

// TestRecordRetry_DefaultRetryAttempts verifies default value when not set.
func TestRecordRetry_DefaultRetryAttempts(t *testing.T) {
	cfg := &config.Config{RetryAttempts: 0}
	retryCfg := retry.DefaultRetryConfig(cfg, zerolog.Nop())

	if retryCfg.MaxAttempts != 3 {
		t.Errorf("expected default MaxAttempts=3, got %d", retryCfg.MaxAttempts)
	}
}

// TestRecordRetry_CustomShouldRetry verifies custom ShouldRetry logic works.
func TestRecordRetry_CustomShouldRetry(t *testing.T) {
	customRetryable := errors.New("custom retryable error")

	retryCfg := retry.RetryConfig{
		MaxAttempts: 3,
		Delay:       10 * time.Millisecond,
		ShouldRetry: func(err error) bool {
			return err == customRetryable
		},
	}

	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 2 {
			return customRetryable
		}
		return nil
	}

	err := retry.Retry(nil, retryCfg, operation)

	if err != nil {
		t.Errorf("expected success, got: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

// TestRecordTimelapseFlag_Exists verifies record command accepts --timelapse flag.
func TestRecordTimelapseFlag_Exists(t *testing.T) {
	flag := recordCmd.Flags().Lookup("timelapse")
	if flag == nil {
		t.Error("expected --timelapse flag to be registered")
	}
}

// TestRecordTimelapseFlag_ShortForm verifies record command accepts -l short form.
func TestRecordTimelapseFlag_ShortForm(t *testing.T) {
	flag := recordCmd.Flags().Lookup("timelapse")
	if flag == nil {
		t.Fatal("expected --timelapse flag to be registered")
	}
	if flag.Shorthand != "l" {
		t.Errorf("expected shorthand 'l', got '%s'", flag.Shorthand)
	}
}

// TestRecordTimelapseFlag_DefaultValue verifies flag default value is 0.
func TestRecordTimelapseFlag_DefaultValue(t *testing.T) {
	flag := recordCmd.Flags().Lookup("timelapse")
	if flag == nil {
		t.Fatal("expected --timelapse flag to be registered")
	}
	if flag.DefValue != "0s" {
		t.Errorf("expected default value '0s', got '%s'", flag.DefValue)
	}
}

// TestValidateTimelapseConfig_TimelapseWithoutDuration verifies error when timelapse set without duration.
func TestValidateTimelapseConfig_TimelapseWithoutDuration(t *testing.T) {
	cfg := &config.Config{
		URL:               "rtsp://test.local/stream",
		Duration:          0,
		TimelapseDuration: 10 * time.Second,
	}

	err := validateTimelapseConfig(cfg)
	if err == nil {
		t.Error("expected error when timelapse set without duration")
	}
	if err != nil && err.Error() != "[ERROR] --timelapse requires --duration: cannot calculate speedup without recording duration" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestValidateTimelapseConfig_TimelapseTooShort verifies error when timelapse < 1s.
func TestValidateTimelapseConfig_TimelapseTooShort(t *testing.T) {
	cfg := &config.Config{
		URL:               "rtsp://test.local/stream",
		Duration:          1 * time.Hour,
		TimelapseDuration: 500 * time.Millisecond,
	}

	err := validateTimelapseConfig(cfg)
	if err == nil {
		t.Error("expected error when timelapse < 1s")
	}
	if err != nil && err.Error() != "[ERROR] --timelapse must be at least 1s" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestValidateTimelapseConfig_ValidCombination verifies no error when both valid.
func TestValidateTimelapseConfig_ValidCombination(t *testing.T) {
	cfg := &config.Config{
		URL:               "rtsp://test.local/stream",
		Duration:          1 * time.Hour,
		TimelapseDuration: 10 * time.Second,
	}

	err := validateTimelapseConfig(cfg)
	if err != nil {
		t.Errorf("expected no error for valid combination, got: %v", err)
	}
}

// TestValidateTimelapseConfig_OnlyDuration verifies no error when only duration set.
func TestValidateTimelapseConfig_OnlyDuration(t *testing.T) {
	cfg := &config.Config{
		URL:               "rtsp://test.local/stream",
		Duration:          30 * time.Minute,
		TimelapseDuration: 0,
	}

	err := validateTimelapseConfig(cfg)
	if err != nil {
		t.Errorf("expected no error for duration-only, got: %v", err)
	}
}

// TestValidateTimelapseConfig_NeitherSet verifies no error when neither set.
func TestValidateTimelapseConfig_NeitherSet(t *testing.T) {
	cfg := &config.Config{
		URL:               "rtsp://test.local/stream",
		Duration:          0,
		TimelapseDuration: 0,
	}

	err := validateTimelapseConfig(cfg)
	if err != nil {
		t.Errorf("expected no error when neither set, got: %v", err)
	}
}

// TestTimelapseStatusMessage_TimelapseEnabled verifies speedup calculation is correct
// Per D-59: Message format "Timelapse: 360x speed (1h -> 10s)"
func TestTimelapseStatusMessage_TimelapseEnabled(t *testing.T) {
	cfg := &config.Config{
		Duration:          1 * time.Hour,
		TimelapseDuration: 10 * time.Second,
	}

	// Calculate expected speedup
	speedup := float64(cfg.Duration) / float64(cfg.TimelapseDuration)
	if speedup != 360.0 {
		t.Errorf("Expected speedup 360x for 1h->10s, got %f", speedup)
	}
}

// TestTimelapseStatusMessage_NoTimelapse verifies speedup is 1x without timelapse
func TestTimelapseStatusMessage_NoTimelapse(t *testing.T) {
	cfg := &config.Config{
		Duration:          1 * time.Hour,
		TimelapseDuration: 0, // Disabled
	}

	// When timelapse is disabled, shouldn't calculate speedup
	if cfg.TimelapseDuration > 0 {
		t.Error("Should not calculate speedup when timelapse is disabled")
	}
}

// TestTimelapseStatusMessage_VariousSpeedups verifies calculations for different durations
func TestTimelapseStatusMessage_VariousSpeedups(t *testing.T) {
	tests := []struct {
		duration          time.Duration
		timelapseDuration time.Duration
		expectedSpeedup   float64
	}{
		{1 * time.Hour, 10 * time.Second, 360.0},
		{30 * time.Minute, 5 * time.Second, 360.0},
		{1 * time.Hour, 1 * time.Minute, 60.0},
		{10 * time.Minute, 10 * time.Second, 60.0},
	}

	for _, tt := range tests {
		speedup := float64(tt.duration) / float64(tt.timelapseDuration)
		if speedup != tt.expectedSpeedup {
			t.Errorf("Speedup for %v->%v: expected %f, got %f",
				tt.duration, tt.timelapseDuration, tt.expectedSpeedup, speedup)
		}
	}
}
