/*
Copyright © 2026 rtsp-recorder contributors

Tests for error classifier package.
*/
package errors

import (
	"strings"
	"testing"
)

// TestClassifyError tests the error classification function
func TestClassifyError(t *testing.T) {
	tests := []struct {
		name           string
		stderr         string
		exitCode       int
		wantCategory   ErrorCategory
		wantRetryable  bool
		wantMsgContain string
	}{
		// Authentication errors
		{
			name:           "401 Unauthorized",
			stderr:         "rtsp://user:pass@cam:401 Unauthorized",
			exitCode:       1,
			wantCategory:   AuthenticationError,
			wantRetryable:  false,
			wantMsgContain: "Authentication required",
		},
		{
			name:           "403 Forbidden",
			stderr:         "403 Forbidden: access denied",
			exitCode:       1,
			wantCategory:   AuthenticationError,
			wantRetryable:  false,
			wantMsgContain: "Authentication required",
		},

		// Stream/path errors (404)
		{
			name:           "404 Not Found",
			stderr:         "method DESCRIBE failed: 404 Not Found",
			exitCode:       1,
			wantCategory:   StreamError,
			wantRetryable:  false,
			wantMsgContain: "Stream path not found",
		},

		// Invalid data / stream errors
		{
			name:           "Invalid data found",
			stderr:         "[rtsp @ 0x7f8b3c0] Invalid data found when processing input",
			exitCode:       1,
			wantCategory:   StreamError,
			wantRetryable:  false,
			wantMsgContain: "Stream data invalid",
		},
		{
			name:           "Unsupported codec",
			stderr:         "codec not currently supported in container",
			exitCode:       1,
			wantCategory:   StreamError,
			wantRetryable:  false,
			wantMsgContain: "Stream data invalid",
		},

		// Network errors - retryable
		{
			name:           "No route to host",
			stderr:         "rtsp://192.168.1.100:554/stream: No route to host",
			exitCode:       1,
			wantCategory:   NetworkError,
			wantRetryable:  true,
			wantMsgContain: "Network unreachable",
		},
		{
			name:           "Network is unreachable",
			stderr:         "Network is unreachable",
			exitCode:       1,
			wantCategory:   NetworkError,
			wantRetryable:  true,
			wantMsgContain: "Network unreachable",
		},
		{
			name:           "Connection refused",
			stderr:         "Connection refused",
			exitCode:       1,
			wantCategory:   NetworkError,
			wantRetryable:  true,
			wantMsgContain: "Cannot connect to camera",
		},
		{
			name:           "Operation timed out",
			stderr:         "Operation timed out",
			exitCode:       1,
			wantCategory:   NetworkError,
			wantRetryable:  true,
			wantMsgContain: "Connection timeout",
		},
		{
			name:           "I/O timeout",
			stderr:         "i/o timeout",
			exitCode:       1,
			wantCategory:   NetworkError,
			wantRetryable:  true,
			wantMsgContain: "Connection timeout",
		},
		{
			name:           "Broken pipe",
			stderr:         "write tcp: broken pipe",
			exitCode:       1,
			wantCategory:   NetworkError,
			wantRetryable:  true,
			wantMsgContain: "Connection interrupted",
		},
		{
			name:           "Connection reset",
			stderr:         "connection reset by peer",
			exitCode:       1,
			wantCategory:   NetworkError,
			wantRetryable:  true,
			wantMsgContain: "Connection interrupted",
		},

		// Configuration errors
		{
			name:           "Invalid argument",
			stderr:         "Invalid argument",
			exitCode:       1,
			wantCategory:   ConfigurationError,
			wantRetryable:  false,
			wantMsgContain: "Invalid configuration",
		},
		{
			name:           "No such file or directory",
			stderr:         "No such file or directory",
			exitCode:       1,
			wantCategory:   ConfigurationError,
			wantRetryable:  false,
			wantMsgContain: "Invalid configuration",
		},
		{
			name:           "Protocol not found",
			stderr:         "Unknown protocol 'rstp'",
			exitCode:       1,
			wantCategory:   ConfigurationError,
			wantRetryable:  false,
			wantMsgContain: "Invalid configuration",
		},

		// Exit code 255
		{
			name:           "Exit code 255",
			stderr:         "some error occurred",
			exitCode:       255,
			wantCategory:   NetworkError,
			wantRetryable:  true,
			wantMsgContain: "Network connection failed",
		},

		// Default FFmpegError
		{
			name:           "Unknown ffmpeg error",
			stderr:         "Some random ffmpeg error message",
			exitCode:       1,
			wantCategory:   FFmpegError,
			wantRetryable:  false,
			wantMsgContain: "FFmpeg failed",
		},
		{
			name:           "Empty stderr with exit code",
			stderr:         "",
			exitCode:       1,
			wantCategory:   FFmpegError,
			wantRetryable:  false,
			wantMsgContain: "FFmpeg failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classified := ClassifyError(tt.stderr, tt.exitCode)

			if classified.Category != tt.wantCategory {
				t.Errorf("ClassifyError() category = %v, want %v", classified.Category, tt.wantCategory)
			}

			if classified.Retryable != tt.wantRetryable {
				t.Errorf("ClassifyError() retryable = %v, want %v", classified.Retryable, tt.wantRetryable)
			}

			if !strings.Contains(classified.Message, tt.wantMsgContain) {
				t.Errorf("ClassifyError() message = %v, should contain %v", classified.Message, tt.wantMsgContain)
			}

			// Verify [ERROR] prefix
			if !strings.HasPrefix(classified.Message, "[ERROR]") {
				t.Errorf("ClassifyError() message should start with [ERROR]: %v", classified.Message)
			}

			// Verify ExitCode is preserved
			if classified.ExitCode != tt.exitCode {
				t.Errorf("ClassifyError() exitCode = %v, want %v", classified.ExitCode, tt.exitCode)
			}

			// Verify Stderr is preserved
			if classified.Stderr != tt.stderr {
				t.Errorf("ClassifyError() stderr not preserved correctly")
			}
		})
	}
}

// TestIsRetryable tests the retryability function
func TestIsRetryable(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		want     bool
	}{
		{NetworkError, true},
		{AuthenticationError, false},
		{StreamError, false},
		{ConfigurationError, false},
		{FFmpegError, false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			got := IsRetryable(tt.category)
			if got != tt.want {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

// TestFormatErrorMessage tests the message formatting
func TestFormatErrorMessage(t *testing.T) {
	// Test with valid classified error
	classified := &ClassifiedError{
		Category:  NetworkError,
		Message:   "[ERROR] Test message",
		Retryable: true,
		ExitCode:  1,
	}

	msg := FormatErrorMessage(classified)
	if msg != "[ERROR] Test message" {
		t.Errorf("FormatErrorMessage() = %v, want %v", msg, "[ERROR] Test message")
	}

	// Test with nil
	msg = FormatErrorMessage(nil)
	if msg != "" {
		t.Errorf("FormatErrorMessage(nil) = %v, want empty string", msg)
	}
}

// TestGetCategoryDescription tests category descriptions
func TestGetCategoryDescription(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		contains string
	}{
		{NetworkError, "Network connectivity"},
		{AuthenticationError, "Authentication"},
		{StreamError, "Stream data"},
		{ConfigurationError, "Configuration"},
		{FFmpegError, "FFmpeg error"},
		{"unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			desc := GetCategoryDescription(tt.category)
			if !strings.Contains(desc, tt.contains) {
				t.Errorf("GetCategoryDescription(%v) = %v, should contain %v", tt.category, desc, tt.contains)
			}
		})
	}
}

// TestTruncateStderr tests stderr truncation
func TestTruncateStderr(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		maxLen int
		want   string
	}{
		{
			name:   "short string not truncated",
			stderr: "short",
			maxLen: 10,
			want:   "short",
		},
		{
			name:   "long string truncated",
			stderr: "this is a very long error message",
			maxLen: 15,
			want:   "this is a ve...",
		},
		{
			name:   "exactly at limit",
			stderr: "exactly10!",
			maxLen: 10,
			want:   "exactly10!",
		},
		{
			name:   "one over limit",
			stderr: "exactly11!!",
			maxLen: 10,
			want:   "exactly...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateStderr(tt.stderr, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateStderr(%q, %d) = %q, want %q", tt.stderr, tt.maxLen, got, tt.want)
			}
		})
	}
}

// TestExtractBitrate tests bitrate extraction from ffmpeg output
func TestExtractBitrate(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   string
	}{
		{
			name:   "bitrate with colon",
			stderr: "bitrate: 1024.5kbits/s",
			want:   "1024.5Kbit/s",
		},
		{
			name:   "bitrate with equals",
			stderr: "bitrate= 2048kbit/s",
			want:   "2048Kbit/s",
		},
		{
			name:   "bitrate mbits",
			stderr: "bitrate: 5.2mbits/s",
			want:   "5.2Mbit/s",
		},
		{
			name:   "no bitrate",
			stderr: "some other output",
			want:   "",
		},
		{
			name:   "empty stderr",
			stderr: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractBitrate(tt.stderr)
			if got != tt.want {
				t.Errorf("ExtractBitrate(%q) = %q, want %q", tt.stderr, got, tt.want)
			}
		})
	}
}

// TestClassifiedErrorError tests the error interface implementation
func TestClassifiedErrorError(t *testing.T) {
	ce := &ClassifiedError{
		Category: NetworkError,
		Message:  "[ERROR] Test error",
	}

	// Should implement error interface
	var _ error = ce

	if ce.Error() != "[ERROR] Test error" {
		t.Errorf("ClassifiedError.Error() = %v, want %v", ce.Error(), "[ERROR] Test error")
	}
}

// TestParseFFmpegErrors tests multi-error parsing
func TestParseFFmpegErrors(t *testing.T) {
	stderr := `Connecting to rtsp://192.168.1.100:554/stream...
Connection refused
Retrying...
Operation timed out`

	errors := ParseFFmpegErrors(stderr)

	// Should find both connection refused and timeout
	foundRefused := false
	foundTimeout := false

	for _, err := range errors {
		if strings.Contains(err.Message, "Cannot connect to camera") {
			foundRefused = true
		}
		if strings.Contains(err.Message, "Connection timeout") {
			foundTimeout = true
		}
	}

	if !foundRefused {
		t.Error("ParseFFmpegErrors should find connection refused error")
	}
	if !foundTimeout {
		t.Error("ParseFFmpegErrors should find timeout error")
	}
}

// TestMatchesAny tests the pattern matching helper
func TestMatchesAny(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		patterns []string
		want     bool
	}{
		{
			name:     "matches first pattern",
			text:     "connection refused by server",
			patterns: []string{"connection refused", "timeout"},
			want:     true,
		},
		{
			name:     "matches second pattern",
			text:     "operation timeout occurred",
			patterns: []string{"connection refused", "timeout"},
			want:     true,
		},
		{
			name:     "no match",
			text:     "some other error",
			patterns: []string{"connection refused", "timeout"},
			want:     false,
		},
		{
			name:     "empty patterns",
			text:     "any text",
			patterns: []string{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesAny(tt.text, tt.patterns...)
			if got != tt.want {
				t.Errorf("matchesAny(%q, %v) = %v, want %v", tt.text, tt.patterns, got, tt.want)
			}
		})
	}
}
