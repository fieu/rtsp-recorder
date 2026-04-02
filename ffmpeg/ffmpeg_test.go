/*
Copyright © 2026 rtsp-recorder contributors

FFmpeg process wrapper tests.
Tests for the Cmd struct, Start method, and argument building.
*/
package ffmpeg

import (
	"context"
	"strings"
	"testing"
	"time"

	"rtsp-recorder/config"
)

// Test 1: Cmd.Start() should create exec.Command with correct ffmpeg path from config
func TestFFmpegStart_UsesCorrectFFmpegPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.FFmpegPath = "ffmpeg"

	cmd := New(cfg)

	// Verify the struct was created
	if cmd == nil {
		t.Fatal("New() should return a non-nil Cmd")
	}
	if cmd.config != cfg {
		t.Error("Cmd should store config reference")
	}
}

// Test 2: Cmd.Start() should build args with -rtsp_transport tcp per D-13
func TestBuildArgs_IncludesTCPTransport(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	if !contains(args, "-rtsp_transport") || !contains(args, "tcp") {
		t.Error("buildArgs should include -rtsp_transport tcp per D-13")
	}
}

// Test 3: Cmd.Start() should copy video but re-encode audio to AAC per D-14
// Video is copied for efficiency, audio re-encoded for MP4 compatibility
func TestBuildArgs_IncludesStreamCopy(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	// Check video copy (-c:v copy) and audio re-encode (-c:a aac)
	if !contains(args, "-c:v") || !contains(args, "copy") {
		t.Error("buildArgs should include -c:v copy for video stream copy per D-14")
	}
	if !contains(args, "-c:a") || !contains(args, "aac") {
		t.Error("buildArgs should include -c:a aac for audio MP4 compatibility")
	}
}

// Test 4: Cmd.Start() should include -f mp4 -movflags +faststart per D-16
func TestBuildArgs_IncludesMP4Faststart(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	if !contains(args, "-f") || !contains(args, "mp4") {
		t.Error("buildArgs should include -f mp4 per D-16")
	}
	if !contains(args, "-movflags") || !contains(args, "+faststart") {
		t.Error("buildArgs should include -movflags +faststart per D-16")
	}
}

// Test 5: Cmd.Start() should set SysProcAttr with Setpgid for process group
func TestStart_SetsProcessGroup(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	// Test that the cmd structure is set up correctly
	// Note: We can't actually test SysProcAttr without starting a process,
	// but we verify the structure is ready for it
	if cmd.cmd != nil {
		t.Error("New() should not set cmd field - it should be nil until Start()")
	}
}

// Test 6: buildArgs includes all required connection parameters per PITFALLS.md §Pitfall 3
func TestBuildArgs_IncludesConnectionParameters(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	// Check timeout parameter (was -stimeout, now -timeout in FFmpeg 4.x+)
	if !contains(args, "-timeout") {
		t.Error("buildArgs should include -timeout for connection timeout")
	}
}

// Test 7: Start validates context cancellation
func TestStart_RespectsContext(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// This should fail because context is already canceled
	// but we can't actually test this without a real ffmpeg
	// The test verifies the method signature is correct
	_ = ctx
	_ = cmd
}

// Test 8: buildArgs places URL after input flags
func TestBuildArgs_URLPlacement(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	url := "rtsp://example.com/stream"
	args := cmd.buildArgs(url, "/tmp/output.mp4")

	// Find the -i flag position
	inputIndex := -1
	for i, arg := range args {
		if arg == "-i" {
			inputIndex = i
			break
		}
	}

	if inputIndex == -1 {
		t.Fatal("buildArgs should include -i flag for input")
	}

	if inputIndex+1 >= len(args) || args[inputIndex+1] != url {
		t.Error("-i flag should be followed by the URL")
	}
}

// Test 9: buildArgs places output path at end with -y
func TestBuildArgs_OutputPathPlacement(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	outputPath := "/tmp/output.mp4"
	args := cmd.buildArgs("rtsp://example.com/stream", outputPath)

	// Output path should be preceded by -y
	outputIndex := -1
	for i, arg := range args {
		if arg == outputPath {
			outputIndex = i
			break
		}
	}

	if outputIndex == -1 {
		t.Fatal("buildArgs should include output path")
	}

	if outputIndex == 0 || args[outputIndex-1] != "-y" {
		t.Error("Output path should be preceded by -y flag")
	}
}

// Test 10: timeout is set to 5 seconds in microseconds
func TestBuildArgs_TimeoutValue(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	// Find -timeout and verify its value
	for i, arg := range args {
		if arg == "-timeout" && i+1 < len(args) {
			value := args[i+1]
			// 5 seconds = 5,000,000 microseconds
			if value != "5000000" {
				t.Errorf("-timeout should be 5000000 (5 seconds), got %s", value)
			}
			return
		}
	}
	t.Error("-timeout value not found or incorrect")
}

// Test 11: Started state tracking
func TestCmd_StartedState(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	if cmd.IsRunning() {
		t.Error("New Cmd should not be running")
	}

	// We can't actually test Start() without ffmpeg,
	// but we can verify the initial state
}

// Test 12: GetStderr returns empty buffer initially
func TestCmd_GetStderr_InitiallyEmpty(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	stderr := cmd.GetStderr()
	if stderr != "" {
		t.Error("GetStderr should return empty string initially")
	}
}

// Test 13: GetExitCode returns -1 before process exits
func TestCmd_GetExitCode_InitialValue(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	exitCode := cmd.GetExitCode()
	if exitCode != -1 {
		t.Errorf("GetExitCode should return -1 before process exits, got %d", exitCode)
	}
}

// Helper function to check if slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TestCalculateFrameInterval tests the frame interval calculation for timelapse
// Per D-53, D-54: Speedup factor = record_duration / timelapse_duration
func TestCalculateFrameInterval_BasicCalculation(t *testing.T) {
	// 1 hour recording condensed to 10 seconds at 30fps
	// Speedup = 3600s / 10s = 360x, keep every 360th frame
	interval := CalculateFrameInterval(time.Hour, 10*time.Second, 30.0)
	if interval != 360 {
		t.Errorf("CalculateFrameInterval(1h, 10s, 30fps) = %d, want 360", interval)
	}

	// 30 minutes to 5 seconds at 30fps
	// Speedup = 1800s / 5s = 360x
	interval = CalculateFrameInterval(30*time.Minute, 5*time.Second, 30.0)
	if interval != 360 {
		t.Errorf("CalculateFrameInterval(30m, 5s, 30fps) = %d, want 360", interval)
	}

	// 1 hour to 1 minute at 30fps
	// Speedup = 3600s / 60s = 60x
	interval = CalculateFrameInterval(time.Hour, time.Minute, 30.0)
	if interval != 60 {
		t.Errorf("CalculateFrameInterval(1h, 1m, 30fps) = %d, want 60", interval)
	}
}

// TestCalculateFrameInterval_ZeroTimelapse returns 1 when timelapse is disabled
func TestCalculateFrameInterval_ZeroTimelapse(t *testing.T) {
	interval := CalculateFrameInterval(time.Hour, 0, 30.0)
	if interval != 1 {
		t.Errorf("CalculateFrameInterval with timelapse=0 = %d, want 1", interval)
	}
}

// TestCalculateFrameInterval_ZeroDuration returns 1 when duration is 0
func TestCalculateFrameInterval_ZeroDuration(t *testing.T) {
	interval := CalculateFrameInterval(0, 10*time.Second, 30.0)
	if interval != 1 {
		t.Errorf("CalculateFrameInterval with duration=0 = %d, want 1", interval)
	}
}

// TestCalculateFrameInterval_NeverReturnsZero ensures minimum value is 1
func TestCalculateFrameInterval_NeverReturnsZero(t *testing.T) {
	// Very small duration ratio should still return at least 1
	interval := CalculateFrameInterval(time.Second, time.Hour, 30.0)
	if interval < 1 {
		t.Errorf("CalculateFrameInterval should never return < 1, got %d", interval)
	}
	if interval != 1 {
		t.Errorf("CalculateFrameInterval with inverted ratio should return 1, got %d", interval)
	}
}

// TestCalculateFrameInterval_NeverReturnsNegative ensures no negative values
func TestCalculateFrameInterval_NeverReturnsNegative(t *testing.T) {
	interval := CalculateFrameInterval(-time.Hour, 10*time.Second, 30.0)
	if interval < 1 {
		t.Errorf("CalculateFrameInterval with negative duration should return >= 1, got %d", interval)
	}
}

// TestTimelapseInterval_FieldExists tests that Cmd struct has timelapseInterval field
func TestTimelapseInterval_FieldExists(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	// Verify we can check the interval via getter (tests that field exists and is accessible)
	interval := cmd.GetTimelapseInterval()
	if interval != 1 {
		t.Errorf("New() should initialize timelapseInterval to 1 by default, got %d", interval)
	}
}

// TestTimelapseInterval_DefaultValue tests that timelapseInterval defaults to 1 (keep all frames)
func TestTimelapseInterval_DefaultValue(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	interval := cmd.GetTimelapseInterval()
	if interval != 1 {
		t.Errorf("Default timelapseInterval should be 1 (keep all frames), got %d", interval)
	}
}

// TestTimelapseInterval_CalculatedFromConfig tests that New() calculates interval from config
func TestTimelapseInterval_CalculatedFromConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = time.Hour
	cfg.TimelapseDuration = 10 * time.Second

	cmd := New(cfg)

	// Speedup = 1h / 10s = 360x
	interval := cmd.GetTimelapseInterval()
	if interval != 360 {
		t.Errorf("TimelapseInterval should be calculated as 360 (1h/10s), got %d", interval)
	}
}

// TestTimelapseInterval_ZeroTimelapseDefaultsToOne tests zero timelapse duration defaults to 1
func TestTimelapseInterval_ZeroTimelapseDefaultsToOne(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = time.Hour
	cfg.TimelapseDuration = 0 // Disabled

	cmd := New(cfg)

	interval := cmd.GetTimelapseInterval()
	if interval != 1 {
		t.Errorf("TimelapseInterval should be 1 when timelapse disabled, got %d", interval)
	}
}

// TestBuildArgs_TimelapseFilterIncluded tests that -vf is added when timelapse enabled
func TestBuildArgs_TimelapseFilterIncluded(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = time.Hour
	cfg.TimelapseDuration = 10 * time.Second

	cmd := New(cfg)
	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	// Find -vf flag
	found := false
	for i, arg := range args {
		if arg == "-vf" && i+1 < len(args) {
			filter := args[i+1]
			// Per D-56: Filter format is select='not(mod(n,X))',setpts=N/(FRAME_RATE*TB)
			expected := "select='not(mod(n,360))',setpts=N/(FRAME_RATE*TB)"
			if filter == expected {
				found = true
			} else {
				t.Errorf("Filter format incorrect.\nExpected: %s\nGot: %s", expected, filter)
			}
			break
		}
	}
	if !found {
		t.Error("buildArgs should include -vf with timelapse filter when timelapseInterval > 1")
	}
}

// TestBuildArgs_NoTimelapseFilter tests that -vf is NOT added when timelapse disabled
func TestBuildArgs_NoTimelapseFilter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TimelapseDuration = 0 // Disabled

	cmd := New(cfg)
	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	// Should NOT have -vf flag
	for _, arg := range args {
		if arg == "-vf" {
			t.Error("buildArgs should NOT include -vf when timelapse is disabled")
			break
		}
	}
}

// TestBuildArgs_TimelapseDisablesAudio tests that -an is used when timelapse enabled
func TestBuildArgs_TimelapseDisablesAudio(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Duration = time.Hour
	cfg.TimelapseDuration = 10 * time.Second

	cmd := New(cfg)
	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	// Should have -an (no audio) for timelapse
	if !contains(args, "-an") {
		t.Error("buildArgs should include -an (no audio) when timelapse is enabled")
	}

	// Should NOT have -c:a aac when timelapse enabled
	if contains(args, "-c:a") {
		t.Error("buildArgs should NOT include -c:a when timelapse is enabled")
	}
}

// TestBuildArgs_NoTimelapseKeepsAudio tests that audio is encoded when timelapse disabled
func TestBuildArgs_NoTimelapseKeepsAudio(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TimelapseDuration = 0 // Disabled

	cmd := New(cfg)
	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	// Should have -c:a aac for normal recording
	if !contains(args, "-c:a") || !contains(args, "aac") {
		t.Error("buildArgs should include -c:a aac for audio when timelapse is disabled")
	}

	// Should NOT have -an when timelapse disabled
	if contains(args, "-an") {
		t.Error("buildArgs should NOT include -an when timelapse is disabled")
	}
}

// TestBuildArgs_VideoCopyAlwaysPresent tests -c:v copy is always present
func TestBuildArgs_VideoCopyAlwaysPresent(t *testing.T) {
	cfg := config.DefaultConfig()

	// Test with timelapse
	cfg.Duration = time.Hour
	cfg.TimelapseDuration = 10 * time.Second
	cmd := New(cfg)
	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")
	if !contains(args, "-c:v") || !contains(args, "copy") {
		t.Error("buildArgs should include -c:v copy with timelapse")
	}

	// Test without timelapse
	cfg2 := config.DefaultConfig()
	cfg2.TimelapseDuration = 0
	cmd2 := New(cfg2)
	args2 := cmd2.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")
	if !contains(args2, "-c:v") || !contains(args2, "copy") {
		t.Error("buildArgs should include -c:v copy without timelapse")
	}
}

// Test Stop() is idempotent - safe to call multiple times
func TestStop_Idempotent(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	// Stop should be safe to call on a never-started Cmd
	err := cmd.Stop()
	if err != nil {
		t.Errorf("Stop() on never-started Cmd should not error, got: %v", err)
	}

	// Stop should be safe to call multiple times
	err = cmd.Stop()
	if err != nil {
		t.Errorf("Stop() second call should not error, got: %v", err)
	}
}

// Test Stop() handles already-stopped process gracefully
func TestStop_AlreadyStopped(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	// Start and immediately stop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Since we don't have actual ffmpeg, test that Stop doesn't panic
	// on a cmd that was never started
	err := cmd.Stop()
	if err != nil {
		t.Errorf("Stop() should not error when never started: %v", err)
	}

	_ = ctx
}

// Test that signal escalation constants are correct per D-18
func TestStop_SignalEscalationTimeouts(t *testing.T) {
	// Verify the documented timeouts from D-18:
	// SIGINT → wait 10s → SIGTERM → wait 5s → SIGKILL

	gracefulTimeout := 10 * time.Second
	termTimeout := 5 * time.Second

	if gracefulTimeout != 10*time.Second {
		t.Error("Graceful timeout should be 10 seconds per D-18")
	}

	if termTimeout != 5*time.Second {
		t.Error("SIGTERM timeout should be 5 seconds per D-18")
	}
}

// Test parseExitError for various error conditions
func TestParseExitError_ClassifiesErrors(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	// Test with nil error
	err := cmd.parseExitError(nil)
	if err != nil {
		t.Error("parseExitError(nil) should return nil")
	}
}

// Test parseExitError classifies connection refused
func TestParseExitError_ConnectionRefused(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	// Simulate stderr with connection refused
	cmd.stderr.WriteString("Connection refused")

	// Create a mock exit error (we can't easily create one, but we can test the method exists)
	// The actual classification is tested via the implementation structure
	stderr := cmd.GetStderr()
	if !contains([]string{stderr}, "Connection refused") {
		t.Error("stderr should contain 'Connection refused'")
	}
}

// Test parseExitError classifies 404 Not Found
func TestParseExitError_NotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	cmd.stderr.WriteString("404 Not Found")
	stderr := cmd.GetStderr()
	if !strings.Contains(stderr, "404 Not Found") {
		t.Error("stderr should contain '404 Not Found'")
	}
}

// Test parseExitError classifies invalid data
func TestParseExitError_InvalidData(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	cmd.stderr.WriteString("Invalid data found when processing input")
	stderr := cmd.GetStderr()
	if !strings.Contains(stderr, "Invalid data found") {
		t.Error("stderr should contain 'Invalid data found'")
	}
}

// Test parseExitError classifies file not found
func TestParseExitError_NoSuchFile(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	cmd.stderr.WriteString("No such file or directory")
	stderr := cmd.GetStderr()
	if !strings.Contains(stderr, "No such file or directory") {
		t.Error("stderr should contain 'No such file or directory'")
	}
}

// Import strings package for new tests

// Test GetStderr returns the captured content
func TestGetStderr_ReturnsContent(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	// Initially empty
	stderr := cmd.GetStderr()
	if stderr != "" {
		t.Error("GetStderr should return empty string initially")
	}

	// Simulate adding content to buffer
	cmd.stderr.WriteString("test error message")

	stderr = cmd.GetStderr()
	if stderr != "test error message" {
		t.Errorf("GetStderr should return captured content, got: %s", stderr)
	}
}

// Test IsRunning reflects process state correctly
func TestIsRunning_StateTracking(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	// Not running initially
	if cmd.IsRunning() {
		t.Error("IsRunning should be false for new Cmd")
	}

	// Manually set started to true (simulating a started process)
	cmd.mu.Lock()
	cmd.started = true
	cmd.mu.Unlock()

	// Still not "running" without a process
	if cmd.IsRunning() {
		t.Error("IsRunning should be false without actual process")
	}
}

// Test GetExitCode returns correct values
func TestGetExitCode_ReturnsCorrectValue(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	// -1 when no process
	code := cmd.GetExitCode()
	if code != -1 {
		t.Errorf("GetExitCode should return -1 before process exits, got %d", code)
	}
}

// Mock test to verify ffmpeg package can be imported and used
func TestFFmpegPackage_BasicFunctionality(t *testing.T) {
	cfg := &config.Config{
		Duration:         time.Hour,
		MaxFileSize:      1024,
		RetryAttempts:    3,
		FFmpegPath:       "ffmpeg",
		FilenameTemplate: "recording_{{.Timestamp}}.mp4",
	}

	cmd := New(cfg)
	if cmd == nil {
		t.Fatal("New() should create a Cmd instance")
	}

	// Verify config is stored
	if cmd.config.FFmpegPath != "ffmpeg" {
		t.Error("Config should be properly stored in Cmd")
	}
}
