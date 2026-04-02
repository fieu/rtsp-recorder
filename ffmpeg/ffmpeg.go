/*
Copyright © 2026 rtsp-recorder contributors

FFmpeg process wrapper package.
Manages ffmpeg subprocess lifecycle with proper signal handling for graceful shutdown
and MP4 finalization.

This package follows PITFALLS.md guidance:
- Pitfall 1: Prevents zombie processes via proper signal handling and Wait()
- Pitfall 2: Ensures MP4 moov atom is written via graceful shutdown sequence
- Pitfall 3: Sets connection timeouts to prevent indefinite hangs
- Pitfall 6: Parses stderr for meaningful error classification
- Pitfall 7: Uses Setpgid to prevent process group leaks
*/
package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"rtsp-recorder/config"
	rrerrors "rtsp-recorder/internal/errors"
)

// CalculateFrameInterval calculates the frame selection interval for timelapse.
// Returns the N value for select='not(mod(n,N))' - keep every Nth frame.
// Per D-54: N = total_expected_frames / target_frames
func CalculateFrameInterval(recordDuration, timelapseDuration time.Duration, frameRate float64) int {
	if timelapseDuration <= 0 || recordDuration <= 0 {
		return 1 // No timelapse, keep all frames
	}

	// Per D-53: Speedup factor = record_duration / timelapse_duration
	speedup := float64(recordDuration) / float64(timelapseDuration)

	// Per D-54: Keep every Nth frame where N = speedup (at default frame rate)
	// At 30fps: if speedup is 360x, we keep every 360th frame
	interval := int(speedup)
	if interval < 1 {
		interval = 1
	}

	return interval
}

// Cmd wraps an ffmpeg exec.Cmd with lifecycle management.
// It provides graceful shutdown via signal escalation and proper
// process cleanup to prevent zombies and ensure MP4 finalization.
type Cmd struct {
	cmd               *exec.Cmd      // The underlying ffmpeg command
	config            *config.Config // Reference to configuration settings
	stderr            bytes.Buffer   // Captured stderr for error analysis
	mu                sync.Mutex     // Thread-safe state access
	started           bool           // Track if process has been started
	outputPath        string         // Track output path for cleanup
	timelapseInterval int            // Frame selection interval (1 = keep all, N = keep every Nth)
}

// New creates a new Cmd instance with the given configuration.
// The config provides settings like ffmpeg path, timeout values, etc.
// Per D-56: Interval is calculated at initialization and used in buildArgs().
func New(cfg *config.Config) *Cmd {
	interval := 1 // Default: keep all frames
	if cfg.TimelapseDuration > 0 && cfg.Duration > 0 {
		interval = CalculateFrameInterval(cfg.Duration, cfg.TimelapseDuration, 30.0)
	}

	return &Cmd{
		config:            cfg,
		stderr:            bytes.Buffer{},
		started:           false,
		outputPath:        "",
		timelapseInterval: interval,
	}
}

// Start begins the ffmpeg recording process.
// It builds the command-line arguments, creates the process with proper
// process group settings, and starts the recording.
//
// Arguments are built according to locked decisions:
//   - D-13: TCP transport (-rtsp_transport tcp)
//   - D-14: Stream copy mode (-c copy)
//   - D-16: MP4 output with faststart (-f mp4 -movflags +faststart)
//   - PITFALLS.md §Pitfall 3: Connection timeouts and reconnection
//   - PITFALLS.md §Pitfall 7: Process group for cleanup
func (c *Cmd) Start(ctx context.Context, url, outputPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate ffmpeg path exists
	ffmpegPath := c.config.GetFFmpegPath()
	if ffmpegPath == "" {
		return fmt.Errorf("ffmpeg path not configured")
	}

	// Store output path for potential cleanup
	c.outputPath = outputPath

	// Build ffmpeg arguments per locked decisions
	args := c.buildArgs(url, outputPath)

	// Create command with context for cancellation support
	c.cmd = exec.CommandContext(ctx, ffmpegPath, args...)

	// Set up process group to prevent zombies and enable group kill
	// Per PITFALLS.md §Pitfall 7
	c.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
	}

	// Capture stderr for error analysis
	// Per PITFALLS.md §Pitfall 6
	c.cmd.Stderr = io.MultiWriter(&c.stderr, os.Stderr)

	// Start the process
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	c.started = true
	return nil
}

// Stop implements graceful shutdown with signal escalation per D-18:
// 1. Send SIGINT for graceful shutdown (allows MP4 finalization)
// 2. Wait up to 10 seconds for graceful exit
// 3. Escalate to SIGTERM if still running
// 4. Wait up to 5 more seconds
// 5. Force kill entire process group with SIGKILL if still running
//
// This follows PITFALLS.md §Pitfall 1 (zombie prevention),
// §Pitfall 2 (MP4 finalization), and §Pitfall 7 (process group cleanup).
func (c *Cmd) Stop() error {
	c.mu.Lock()

	// Check if process is running
	if !c.started || c.cmd == nil || c.cmd.Process == nil {
		c.mu.Unlock()
		return nil // Already stopped or never started
	}

	c.started = false
	process := c.cmd.Process
	c.mu.Unlock()

	// Step 1: Send SIGINT for graceful shutdown (allows ffmpeg to finalize MP4)
	if err := process.Signal(syscall.SIGINT); err != nil {
		// Process may have already exited, continue to wait
		return c.waitAndParseError()
	}

	// Create done channel for goroutine to report exit
	done := make(chan error, 1)
	go func() {
		c.mu.Lock()
		cmd := c.cmd
		c.mu.Unlock()
		if cmd != nil {
			done <- cmd.Wait()
		} else {
			done <- nil
		}
	}()

	// Step 2: Wait up to 10 seconds for graceful exit per D-18
	select {
	case err := <-done:
		return c.parseExitError(err)
	case <-time.After(10 * time.Second):
		// Graceful timeout, escalate to SIGTERM
	}

	// Step 3: Escalate to SIGTERM
	c.mu.Lock()
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Signal(syscall.SIGTERM)
	}
	c.mu.Unlock()

	// Step 4: Wait up to 5 more seconds for SIGTERM per D-18
	select {
	case err := <-done:
		return c.parseExitError(err)
	case <-time.After(5 * time.Second):
		// SIGTERM timeout, force kill
	}

	// Step 5: Force kill entire process group with SIGKILL
	// Negative PID sends signal to entire process group per PITFALLS.md §Pitfall 7
	c.mu.Lock()
	if c.cmd != nil && c.cmd.Process != nil {
		syscall.Kill(-c.cmd.Process.Pid, syscall.SIGKILL)
	}
	c.mu.Unlock()

	// Wait for final exit (should be immediate after SIGKILL)
	err := <-done
	return c.parseExitError(err)
}

// buildArgs constructs the ffmpeg command-line arguments according to
// locked decisions and best practices from PITFALLS.md.
func (c *Cmd) buildArgs(url, outputPath string) []string {
	args := []string{
		// D-13: Use TCP transport - more reliable than UDP
		"-rtsp_transport", "tcp",

		// PITFALLS.md §Pitfall 3: Connection timeout (5 seconds in microseconds)
		// Note: -stimeout is deprecated in FFmpeg 4.x+, use -timeout instead
		"-timeout", "5000000",

		// Fix blank/black start: discard corrupted frames and generate proper timestamps
		"-fflags", "+discardcorrupt+genpts",
		"-use_wallclock_as_timestamps", "1",

		// Input URL
		"-i", url,
	}

	// Per D-56, D-57: Add timelapse filter if enabled
	if c.timelapseInterval > 1 {
		filter := fmt.Sprintf("select='not(mod(n,%d))',setpts=N/(FRAME_RATE*TB)", c.timelapseInterval)
		args = append(args, "-vf", filter)
		// Video filter requires re-encoding, cannot use -c:v copy
		args = append(args, "-c:v", "libx264", "-preset", "fast", "-crf", "23")
	} else {
		// Normal recording: copy video without re-encoding (efficient)
		args = append(args, "-c:v", "copy")
	}

	// Per D-58: Audio handling for timelapse
	if c.timelapseInterval > 1 {
		// Drop audio for timelapse (simpler, more typical)
		args = append(args, "-an")
	} else {
		// Normal recording: encode audio to AAC
		args = append(args, "-c:a", "aac", "-b:a", "128k")
	}

	// Remaining args
	args = append(args,
		// Force CFR (constant frame rate) for better seeking
		"-fps_mode", "cfr",

		// D-16: Output format MP4
		"-f", "mp4",

		// D-16, D-26: Faststart - move moov atom to start for web playback
		// and ensure proper MP4 finalization on graceful shutdown
		"-movflags", "+faststart",

		// Overwrite output file without prompting
		"-y",

		// Output path
		outputPath,
	)

	return args
}

// GetStderr returns the captured stderr content from the ffmpeg process.
// This is useful for error analysis and debugging.
func (c *Cmd) GetStderr() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stderr.String()
}

// IsRunning returns true if the ffmpeg process is currently active.
func (c *Cmd) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.started && c.cmd != nil && c.cmd.Process != nil
}

// GetExitCode returns the exit code of the process after it exits.
// Returns -1 if the process hasn't exited yet or was never started.
func (c *Cmd) GetExitCode() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd == nil || c.cmd.ProcessState == nil {
		return -1
	}

	return c.cmd.ProcessState.ExitCode()
}

// GetTimelapseInterval returns the frame selection interval.
// Returns 1 if timelapse is disabled (all frames kept).
func (c *Cmd) GetTimelapseInterval() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.timelapseInterval
}

// GetSpeedupFactor returns the calculated speedup factor for timelapse.
// Returns 1 if timelapse is disabled.
// Per D-59: Enables progress display to show "[INFO] Timelapse: 360x speed (1h -> 10s)"
func (c *Cmd) GetSpeedupFactor() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.timelapseInterval <= 1 {
		return 1.0
	}

	// Speedup is approximately the interval (at default frame rate assumption)
	// More precise: use config durations
	if c.config.TimelapseDuration > 0 && c.config.Duration > 0 {
		return float64(c.config.Duration) / float64(c.config.TimelapseDuration)
	}
	return float64(c.timelapseInterval)
}

// waitAndParseError waits for the process to exit and parses the error.
func (c *Cmd) waitAndParseError() error {
	c.mu.Lock()
	cmd := c.cmd
	c.mu.Unlock()

	if cmd == nil {
		return nil
	}

	err := cmd.Wait()
	return c.parseExitError(err)
}

// parseExitError analyzes the exit error and stderr content to provide
// meaningful error messages per PITFALLS.md §Pitfall 6.
// Uses internal/errors.ClassifyError for consistent error classification.
func (c *Cmd) parseExitError(err error) error {
	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		stderr := c.stderr.String()
		exitCode := exitErr.ExitCode()

		// Use error classifier for consistent, actionable messages
		classified := rrerrors.ClassifyError(stderr, exitCode)
		classified.Original = err

		return classified
	}

	return fmt.Errorf("[ERROR] FFmpeg process error: %w", err)
}
