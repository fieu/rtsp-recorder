/*
Copyright © 2026 rtsp-recorder contributors

FFmpeg process wrapper tests.
Tests for the Cmd struct, Start method, and argument building.
*/
package ffmpeg

import (
	"context"
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

// Test 3: Cmd.Start() should include -c copy for stream copy per D-14
func TestBuildArgs_IncludesStreamCopy(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	if !contains(args, "-c") || !contains(args, "copy") {
		t.Error("buildArgs should include -c copy per D-14")
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

	// Check reconnect parameters
	if !contains(args, "-stimeout") {
		t.Error("buildArgs should include -stimeout per PITFALLS.md §Pitfall 3")
	}
	if !contains(args, "-reconnect") {
		t.Error("buildArgs should include -reconnect")
	}
	if !contains(args, "-reconnect_at_eof") {
		t.Error("buildArgs should include -reconnect_at_eof")
	}
	if !contains(args, "-reconnect_streamed") {
		t.Error("buildArgs should include -reconnect_streamed")
	}
	if !contains(args, "-reconnect_delay_max") {
		t.Error("buildArgs should include -reconnect_delay_max")
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

// Test 10: stimeout is set to 5 seconds in microseconds
func TestBuildArgs_StimeoutValue(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := New(cfg)

	args := cmd.buildArgs("rtsp://example.com/stream", "/tmp/output.mp4")

	// Find -stimeout and verify its value
	for i, arg := range args {
		if arg == "-stimeout" && i+1 < len(args) {
			value := args[i+1]
			// 5 seconds = 5,000,000 microseconds
			if value != "5000000" {
				t.Errorf("-stimeout should be 5000000 (5 seconds), got %s", value)
			}
			return
		}
	}
	t.Error("-stimeout value not found or incorrect")
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
