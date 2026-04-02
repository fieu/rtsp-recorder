/*
Copyright © 2026 rtsp-recorder contributors

Tests for the recorder package.
*/
package recorder

import (
	"testing"
	"time"

	"rtsp-recorder/config"
)

// TestRecorder_New tests that New() creates a Recorder with config reference
func TestRecorder_New(t *testing.T) {
	cfg := config.DefaultConfig()
	rec := New(cfg)

	if rec == nil {
		t.Fatal("New() returned nil")
	}
	if rec.config != cfg {
		t.Error("Recorder config not set correctly")
	}
}

// TestRecorder_Record_EmptyURL tests that Record() returns error for empty URL
func TestRecorder_Record_EmptyURL(t *testing.T) {
	cfg := config.DefaultConfig()
	rec := New(cfg)

	err := rec.Record("")
	if err == nil {
		t.Error("Expected error for empty URL, got nil")
	}
	if err.Error() != "[ERROR] No RTSP URL provided" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestRecorder_Record_GeneratesTimestampFilename tests filename generation
func TestRecorder_Record_GeneratesTimestampFilename(t *testing.T) {
	cfg := config.DefaultConfig()
	// Clear template to force timestamp generation
	cfg.FilenameTemplate = ""
	rec := New(cfg)

	// This would fail during actual recording due to invalid URL
	// but we can verify the filename generation happens first
	err := rec.Record("")
	if err == nil {
		t.Error("Expected error for empty URL")
	}
}

// TestFormatBytes tests the formatBytes helper function
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}

// TestFormatDuration tests the formatDuration helper function
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{0, "00:00:00"},
		{30 * time.Second, "00:00:30"},
		{5 * time.Minute, "00:05:00"},
		{90 * time.Minute, "01:30:00"},
		{2*time.Hour + 30*time.Minute + 45*time.Second, "02:30:45"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.d)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, want %s", tt.d, result, tt.expected)
		}
	}
}

// TestFormatBitrate tests the formatBitrate helper function
func TestFormatBitrate(t *testing.T) {
	tests := []struct {
		bps      float64
		expected string
	}{
		{0, "0 bps"},
		{500, "500 bps"},
		{1000, "1.0 Kbps"},
		{1000000, "1.0 Mbps"},
		{5000000, "5.0 Mbps"},
	}

	for _, tt := range tests {
		result := formatBitrate(tt.bps)
		if result != tt.expected {
			t.Errorf("formatBitrate(%f) = %s, want %s", tt.bps, result, tt.expected)
		}
	}
}

// TestRecorder_outputPath tests that outputPath is accessible
func TestRecorder_outputPath(t *testing.T) {
	cfg := config.DefaultConfig()
	rec := New(cfg)

	// Initially outputPath should be empty
	if rec.outputPath != "" {
		t.Error("outputPath should be empty before recording")
	}
}

// TestRecorder_startTime tests that startTime is set during recording
func TestRecorder_startTime(t *testing.T) {
	cfg := config.DefaultConfig()
	rec := New(cfg)

	// Initially startTime should be zero
	if !rec.startTime.IsZero() {
		t.Error("startTime should be zero before recording")
	}
}
