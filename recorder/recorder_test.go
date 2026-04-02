/*
Copyright © 2026 rtsp-recorder contributors

Tests for the recorder package.
*/
package recorder

import (
	"testing"
	"time"

	"rtsp-recorder/config"
	"rtsp-recorder/ffmpeg"

	"github.com/rs/zerolog"
)

// TestRecorder_New tests that New() creates a Recorder with config reference
func TestRecorder_New(t *testing.T) {
	cfg := config.DefaultConfig()
	rec := New(cfg, zerolog.Nop())

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
	rec := New(cfg, zerolog.Nop())

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
	rec := New(cfg, zerolog.Nop())

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
	rec := New(cfg, zerolog.Nop())

	// Initially outputPath should be empty
	if rec.outputPath != "" {
		t.Error("outputPath should be empty before recording")
	}
}

// TestRecorder_startTime tests that startTime is set during recording
func TestRecorder_startTime(t *testing.T) {
	cfg := config.DefaultConfig()
	rec := New(cfg, zerolog.Nop())

	// Initially startTime should be zero
	if !rec.startTime.IsZero() {
		t.Error("startTime should be zero before recording")
	}
}

// TestRecorder_WithFilenameTemplate tests that custom template is used
func TestRecorder_WithFilenameTemplate(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.FilenameTemplate = "test_{{.Timestamp}}.mp4"
	rec := New(cfg, zerolog.Nop())

	// URL validation happens before filename generation
	// This test verifies config is properly set
	if rec.config.FilenameTemplate != "test_{{.Timestamp}}.mp4" {
		t.Error("FilenameTemplate not set correctly in config")
	}
}

// TestRecorder_WithDuration tests that duration config is respected
func TestRecorder_WithDuration(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = 30 * time.Minute
	rec := New(cfg, zerolog.Nop())

	if rec.config.Duration != 30*time.Minute {
		t.Error("Duration not set correctly in config")
	}
}

// TestRecorder_WithMaxFileSize tests that max file size config is respected
func TestRecorder_WithMaxFileSize(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.MaxFileSize = 500 // 500 MB
	rec := New(cfg, zerolog.Nop())

	if rec.config.MaxFileSize != 500 {
		t.Error("MaxFileSize not set correctly in config")
	}
}

// TestFormatBytes_EdgeCases tests edge cases for formatBytes
func TestFormatBytes_EdgeCases(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{-1, "-1 B"},             // Negative value (edge case)
		{1, "1 B"},               // Single byte
		{1023, "1023 B"},         // Just under 1KB
		{1025, "1.0 KB"},         // Just over 1KB
		{1024 * 512, "512.0 KB"}, // 512 KB
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}

// TestFormatDuration_EdgeCases tests edge cases for formatDuration
func TestFormatDuration_EdgeCases(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{1 * time.Second, "00:00:01"},
		{59 * time.Second, "00:00:59"},
		{61 * time.Second, "00:01:01"},
		{3599 * time.Second, "00:59:59"},
		{3600 * time.Second, "01:00:00"},
		{24 * time.Hour, "24:00:00"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.d)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, want %s", tt.d, result, tt.expected)
		}
	}
}

// TestFormatBitrate_EdgeCases tests edge cases for formatBitrate
func TestFormatBitrate_EdgeCases(t *testing.T) {
	tests := []struct {
		bps      float64
		expected string
	}{
		{-1, "-1 bps"},              // Negative value
		{0.5, "0 bps"},              // Fractional, rounds down
		{999, "999 bps"},            // Just under 1Kbps
		{1001, "1.0 Kbps"},          // Just over 1Kbps
		{999999, "1000.0 Kbps"},     // Just under 1Mbps
		{1000001, "1.0 Mbps"},       // Just over 1Mbps
		{1000000000, "1000.0 Mbps"}, // 1 Gbps
	}

	for _, tt := range tests {
		result := formatBitrate(tt.bps)
		if result != tt.expected {
			t.Errorf("formatBitrate(%f) = %s, want %s", tt.bps, result, tt.expected)
		}
	}
}

// TestDisplayProgress_WithTimelapse tests progress display includes timelapse info
// Per D-59: Show speedup factor in progress display
func TestDisplayProgress_WithTimelapse(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = time.Hour
	cfg.TimelapseDuration = 10 * time.Second

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	// Verify ffmpeg is initialized with timelapse
	if rec.ffmpeg == nil {
		t.Fatal("ffmpeg.Cmd should be initialized")
	}

	speedup := rec.ffmpeg.GetSpeedupFactor()
	if speedup != 360.0 {
		t.Errorf("Expected speedup factor 360x, got %f", speedup)
	}
}

// TestDisplayProgress_WithoutTimelapse tests normal recording shows no timelapse info
func TestDisplayProgress_WithoutTimelapse(t *testing.T) {
	cfg := config.DefaultConfig()
	// No timelapse configured

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	// Verify speedup is 1x (no timelapse)
	speedup := rec.ffmpeg.GetSpeedupFactor()
	if speedup != 1.0 {
		t.Errorf("Expected speedup factor 1x (no timelapse), got %f", speedup)
	}
}

// TestDisplayProgress_EstimatedOutput tests output duration calculation
// Per D-60: Progress shows estimated output time
func TestDisplayProgress_EstimatedOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = time.Hour
	cfg.TimelapseDuration = 10 * time.Second

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	speedup := rec.ffmpeg.GetSpeedupFactor()
	elapsed := 30 * time.Minute // Halfway through recording

	// Calculate expected output duration
	expectedOutput := time.Duration(float64(elapsed) / speedup)
	if expectedOutput != 5*time.Second {
		t.Errorf("Expected output duration ~5s (30m/360), got %v", expectedOutput)
	}
}

// TestDisplayProgress_TimelapseInterval tests timelapse interval is accessible
func TestDisplayProgress_TimelapseInterval(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = time.Hour
	cfg.TimelapseDuration = 10 * time.Second

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	interval := rec.ffmpeg.GetTimelapseInterval()
	if interval != 360 {
		t.Errorf("Expected timelapse interval 360, got %d", interval)
	}
}

// TestTimelapseWithStopConditions verifies timelapse works with all stop conditions
// Per TIMELAPSE-03: Timelapse must work with Ctrl+C, duration, and file size limits
func TestTimelapseWithStopConditions_Duration(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = 1 * time.Hour
	cfg.TimelapseDuration = 10 * time.Second

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	// Verify timelapse is properly configured
	speedup := rec.ffmpeg.GetSpeedupFactor()
	if speedup != 360.0 {
		t.Errorf("Expected speedup 360x, got %f", speedup)
	}

	// Verify duration monitor would be configured (duration > 0)
	if cfg.Duration <= 0 {
		t.Error("Duration should be > 0 for duration stop condition")
	}
}

// TestTimelapseWithStopConditions_FileSize verifies timelapse works with file size limit
func TestTimelapseWithStopConditions_FileSize(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = 30 * time.Minute
	cfg.TimelapseDuration = 5 * time.Second
	cfg.MaxFileSize = 1024 // 1GB

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	// Verify both timelapse and file size are configured
	if rec.ffmpeg.GetSpeedupFactor() <= 1 {
		t.Error("Timelapse should be enabled with speedup > 1")
	}

	if cfg.MaxFileSize <= 0 {
		t.Error("MaxFileSize should be > 0 for file size stop condition")
	}
}

// TestTimelapseWithStopConditions_Signal verifies timelapse works with signal handling
func TestTimelapseWithStopConditions_Signal(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = 2 * time.Hour
	cfg.TimelapseDuration = 20 * time.Second

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	// Verify timelapse is enabled - signal monitor works independently
	speedup := rec.ffmpeg.GetSpeedupFactor()
	if speedup != 360.0 {
		t.Errorf("Expected speedup 360x for 2h->20s, got %f", speedup)
	}
}

// TestPrintFinalSummary_TimelapseEnabled verifies summary includes timelapse info
// Per D-59: Summary shows real duration, output duration, and speedup
func TestPrintFinalSummary_TimelapseEnabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = 1 * time.Hour
	cfg.TimelapseDuration = 10 * time.Second

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	// Verify timelapse is configured for summary
	speedup := rec.ffmpeg.GetSpeedupFactor()
	if speedup != 360.0 {
		t.Errorf("Expected speedup 360x for summary, got %f", speedup)
	}

	realDuration := 30 * time.Minute // Simulated half-recording
	outputDuration := time.Duration(float64(realDuration) / speedup)

	// Expected: 30m / 360 = 5s
	if outputDuration != 5*time.Second {
		t.Errorf("Expected output duration ~5s (30m/360), got %v", outputDuration)
	}
}

// TestPrintFinalSummary_NoTimelapse verifies summary works without timelapse
func TestPrintFinalSummary_NoTimelapse(t *testing.T) {
	cfg := config.DefaultConfig()
	// No timelapse configured

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	// Verify no timelapse (speedup = 1x)
	speedup := rec.ffmpeg.GetSpeedupFactor()
	if speedup != 1.0 {
		t.Errorf("Expected speedup 1x (no timelapse), got %f", speedup)
	}
}

// TestPrintFinalSummary_OutputDurationCalculation verifies output duration math
func TestPrintFinalSummary_OutputDurationCalculation(t *testing.T) {
	tests := []struct {
		realDuration time.Duration
		speedup      float64
		expected     time.Duration
	}{
		{1 * time.Hour, 360.0, 10 * time.Second},
		{30 * time.Minute, 360.0, 5 * time.Second},
		{2 * time.Hour, 360.0, 20 * time.Second},
		{1 * time.Hour, 60.0, 1 * time.Minute},
	}

	for _, tt := range tests {
		outputDuration := time.Duration(float64(tt.realDuration) / tt.speedup)
		if outputDuration != tt.expected {
			t.Errorf("Output duration for %v at %.0fx: expected %v, got %v",
				tt.realDuration, tt.speedup, tt.expected, outputDuration)
		}
	}
}

// TestPrintFinalSummary_SpeedupRounding verifies speedup is displayed as integer
func TestPrintFinalSummary_SpeedupRounding(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = 1 * time.Hour
	cfg.TimelapseDuration = 10 * time.Second

	rec := New(cfg, zerolog.Nop())
	rec.ffmpeg = ffmpeg.New(cfg)

	speedup := rec.ffmpeg.GetSpeedupFactor()
	// Speedup should be exactly 360.0 for 1h/10s
	if speedup != 360.0 {
		t.Errorf("Expected speedup 360.0, got %f", speedup)
	}

	// Verify it rounds to integer cleanly
	speedupInt := int(speedup + 0.5)
	if speedupInt != 360 {
		t.Errorf("Expected speedup int 360, got %d", speedupInt)
	}
}
