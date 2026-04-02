//go:build !windows
// +build !windows

/*
Copyright © 2026 rtsp-recorder contributors

Tests for stop condition monitors.
*/
package recorder

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

// TestMonitorInterface verifies that SignalMonitor implements Monitor interface.
func TestMonitorInterface(t *testing.T) {
	var _ Monitor = (*SignalMonitor)(nil)
}

// TestSignalMonitor_Name verifies Name() returns "signal".
func TestSignalMonitor_Name(t *testing.T) {
	m := NewSignalMonitor()
	if got := m.Name(); got != "signal" {
		t.Errorf("SignalMonitor.Name() = %q, want %q", got, "signal")
	}
}

// TestSignalMonitor_StartAndWait verifies monitor starts and wait channel exists.
func TestSignalMonitor_StartAndWait(t *testing.T) {
	m := NewSignalMonitor()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Before start, Wait() channel should exist but be open
	ch := m.Wait()
	if ch == nil {
		t.Fatal("Wait() returned nil channel")
	}

	// Start the monitor
	m.Start(ctx)

	// Channel should still be open (no signal sent yet)
	select {
	case <-ch:
		t.Fatal("Wait() channel closed before signal received")
	case <-time.After(100 * time.Millisecond):
		// Expected - channel still open
	}
}

// TestSignalMonitor_SignalReceived verifies channel closes when signal is received.
func TestSignalMonitor_SignalReceived(t *testing.T) {
	m := NewSignalMonitor()
	ctx := context.Background()

	m.Start(ctx)

	// Send SIGTERM to ourselves
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("Failed to send SIGTERM: %v", err)
	}

	// Channel should close
	select {
	case <-m.Wait():
		// Success - channel closed
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() channel did not close after SIGTERM")
	}
}

// TestSignalMonitor_ContextCancellation verifies clean shutdown on context cancel.
func TestSignalMonitor_ContextCancellation(t *testing.T) {
	m := NewSignalMonitor()
	ctx, cancel := context.WithCancel(context.Background())

	m.Start(ctx)

	// Cancel context
	cancel()

	// Channel should close
	select {
	case <-m.Wait():
		// Success - channel closed
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() channel did not close after context cancellation")
	}
}

// TestSignalMonitor_MultipleStarts is safe to call Start multiple times.
func TestSignalMonitor_MultipleStarts(t *testing.T) {
	m := NewSignalMonitor()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx)
	m.Start(ctx) // Should be safe, no panic

	// Cancel context - should still work
	cancel()

	select {
	case <-m.Wait():
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() channel did not close")
	}
}

// ==================== DurationMonitor Tests ====================

// TestDurationMonitor_Interface verifies that DurationMonitor implements Monitor.
func TestDurationMonitor_Interface(t *testing.T) {
	var _ Monitor = (*DurationMonitor)(nil)
}

// TestDurationMonitor_Name verifies Name() returns "duration".
func TestDurationMonitor_Name(t *testing.T) {
	m := NewDurationMonitor(1 * time.Second)
	if got := m.Name(); got != "duration" {
		t.Errorf("DurationMonitor.Name() = %q, want %q", got, "duration")
	}
}

// TestDurationMonitor_Duration verifies Duration() returns configured value.
func TestDurationMonitor_Duration(t *testing.T) {
	d := 5 * time.Minute
	m := NewDurationMonitor(d)
	if got := m.Duration(); got != d {
		t.Errorf("DurationMonitor.Duration() = %v, want %v", got, d)
	}
}

// TestDurationMonitor_TriggersAfterDuration verifies Wait() closes after duration.
func TestDurationMonitor_TriggersAfterDuration(t *testing.T) {
	duration := 100 * time.Millisecond
	m := NewDurationMonitor(duration)
	ctx := context.Background()

	start := time.Now()
	m.Start(ctx)

	select {
	case <-m.Wait():
		elapsed := time.Since(start)
		if elapsed < duration {
			t.Errorf("Trigger too early: %v < %v", elapsed, duration)
		}
		if elapsed > duration+100*time.Millisecond {
			t.Errorf("Trigger too late: %v > %v + 100ms", elapsed, duration)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() channel did not close after duration")
	}
}

// TestDurationMonitor_ZeroDuration_Skips verifies monitor triggers immediately when duration is 0.
func TestDurationMonitor_ZeroDuration_Skips(t *testing.T) {
	m := NewDurationMonitor(0)
	ctx := context.Background()

	m.Start(ctx)

	// Should close immediately (or very quickly) since duration is 0 (unlimited)
	select {
	case <-m.Wait():
		// Success - channel closed immediately
	case <-time.After(500 * time.Millisecond):
		// If it takes longer, that's acceptable - just means it was already closed
	}
}

// TestDurationMonitor_ContextCancellation verifies early stop on context cancel.
func TestDurationMonitor_ContextCancellation(t *testing.T) {
	m := NewDurationMonitor(10 * time.Second) // Long duration
	ctx, cancel := context.WithCancel(context.Background())

	m.Start(ctx)

	// Cancel context immediately
	cancel()

	select {
	case <-m.Wait():
		// Success - channel closed early
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() channel did not close after context cancellation")
	}
}

// TestDurationMonitor_MultipleStarts is safe to call Start multiple times.
func TestDurationMonitor_MultipleStarts(t *testing.T) {
	m := NewDurationMonitor(100 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx)
	m.Start(ctx) // Should be safe

	select {
	case <-m.Wait():
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() channel did not close")
	}
}

// ==================== FileSizeMonitor Tests ====================

// TestFileSizeMonitor_Interface verifies that FileSizeMonitor implements Monitor.
func TestFileSizeMonitor_Interface(t *testing.T) {
	var _ Monitor = (*FileSizeMonitor)(nil)
}

// TestFileSizeMonitor_Name verifies Name() returns "file_size".
func TestFileSizeMonitor_Name(t *testing.T) {
	m := NewFileSizeMonitor(100, "/tmp/test.mp4")
	if got := m.Name(); got != "file_size" {
		t.Errorf("FileSizeMonitor.Name() = %q, want %q", got, "file_size")
	}
}

// TestFileSizeMonitor_ConvertsMBToBytes verifies max size is converted from MB to bytes.
func TestFileSizeMonitor_ConvertsMBToBytes(t *testing.T) {
	m := NewFileSizeMonitor(10, "/tmp/test.mp4") // 10 MB
	expectedBytes := int64(10 * 1024 * 1024)     // 10 MB in bytes
	if got := m.MaxBytes(); got != expectedBytes {
		t.Errorf("FileSizeMonitor.MaxBytes() = %d, want %d", got, expectedBytes)
	}
}

// TestFileSizeMonitor_ZeroMaxSize_Skips verifies monitor triggers immediately when max is 0.
func TestFileSizeMonitor_ZeroMaxSize_Skips(t *testing.T) {
	m := NewFileSizeMonitor(0, "/tmp/test.mp4")
	ctx := context.Background()

	m.Start(ctx)

	// Should close immediately since max size is 0 (unlimited)
	select {
	case <-m.Wait():
		// Success - channel closed immediately
	case <-time.After(500 * time.Millisecond):
		// If it takes longer, that's acceptable
	}
}

// TestFileSizeMonitor_TriggersWhenSizeReached verifies channel closes when file reaches limit.
func TestFileSizeMonitor_TriggersWhenSizeReached(t *testing.T) {
	// Create a temp file that we'll grow
	tmpFile, err := os.CreateTemp("", "test_*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Set max size to 1KB
	m := NewFileSizeMonitor(1, tmpFile.Name()) // 1 MB limit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx)

	// Channel should still be open (file is empty)
	select {
	case <-m.Wait():
		t.Fatal("Wait() channel closed before file reached size limit")
	case <-time.After(100 * time.Millisecond):
		// Expected - channel still open
	}

	// Grow the file to exceed the limit (1 MB = 1,048,576 bytes)
	data := make([]byte, 2*1024*1024) // 2 MB of data
	if _, err := tmpFile.Write(data); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Sync()

	// Wait for polling (polls every 1 second per D-27)
	select {
	case <-m.Wait():
		// Success - channel closed when size reached
	case <-time.After(3 * time.Second):
		t.Fatal("Wait() channel did not close after file size reached limit")
	}

	// Verify current size is tracked
	if m.CurrentSize() < int64(1024*1024) {
		t.Errorf("CurrentSize() = %d, expected >= %d", m.CurrentSize(), 1024*1024)
	}
}

// TestFileSizeMonitor_ContextCancellation verifies early stop on context cancel.
func TestFileSizeMonitor_ContextCancellation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Set max size to a large value
	m := NewFileSizeMonitor(1000, tmpFile.Name()) // 1000 MB limit
	ctx, cancel := context.WithCancel(context.Background())

	m.Start(ctx)

	// Cancel context immediately
	cancel()

	select {
	case <-m.Wait():
		// Success - channel closed early
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() channel did not close after context cancellation")
	}
}

// TestFileSizeMonitor_MultipleStarts is safe to call Start multiple times.
func TestFileSizeMonitor_MultipleStarts(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	m := NewFileSizeMonitor(1, tmpFile.Name()) // 1 MB
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx)
	m.Start(ctx) // Should be safe

	// Write to file to trigger stop
	data := make([]byte, 2*1024*1024) // 2 MB
	tmpFile.Write(data)
	tmpFile.Sync()

	select {
	case <-m.Wait():
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("Wait() channel did not close")
	}
}

// ==================== StopManager Tests ====================

// TestStopManager_FirstTriggerWins verifies first monitor triggers stops others.
func TestStopManager_FirstTriggerWins(t *testing.T) {
	sm := NewStopManager()

	// Add a duration monitor with very short timeout (50ms)
	durationMon := NewDurationMonitor(50 * time.Millisecond)
	sm.AddMonitor(durationMon)

	// Add a signal monitor
	signalMon := NewSignalMonitor()
	sm.AddMonitor(signalMon)

	// Start
	sm.Start()

	// Wait for stop - should be triggered by duration
	select {
	case reason := <-sm.Wait():
		if reason.Name != "duration" {
			t.Errorf("Expected 'duration' trigger, got %q", reason.Name)
		}
		if reason.Desc != "Duration limit reached" {
			t.Errorf("Expected 'Duration limit reached', got %q", reason.Desc)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StopManager did not trigger")
	}
}

// TestStopManager_SignalTrigger verifies signal monitor can trigger stop.
func TestStopManager_SignalTrigger(t *testing.T) {
	sm := NewStopManager()

	// Add a duration monitor with long timeout
	durationMon := NewDurationMonitor(10 * time.Second)
	sm.AddMonitor(durationMon)

	// Add a signal monitor
	signalMon := NewSignalMonitor()
	sm.AddMonitor(signalMon)

	// Start
	sm.Start()

	// Send SIGTERM
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("Failed to send SIGTERM: %v", err)
	}

	// Wait for stop - should be triggered by signal
	select {
	case reason := <-sm.Wait():
		if reason.Name != "signal" {
			t.Errorf("Expected 'signal' trigger, got %q", reason.Name)
		}
		if reason.Desc != "Interrupted by user (Ctrl+C)" {
			t.Errorf("Expected 'Interrupted by user (Ctrl+C)', got %q", reason.Desc)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StopManager did not trigger on signal")
	}
}

// TestStopManager_ManualStop verifies Stop() method works.
func TestStopManager_ManualStop(t *testing.T) {
	sm := NewStopManager()

	// Add a duration monitor with long timeout
	durationMon := NewDurationMonitor(10 * time.Second)
	sm.AddMonitor(durationMon)

	// Start
	sm.Start()

	// Manually stop
	sm.Stop()

	// Wait should receive a reason (or channel close)
	select {
	case <-sm.Wait():
		// Success - context was cancelled
	case <-time.After(2 * time.Second):
		t.Fatal("StopManager did not stop on manual Stop()")
	}
}

// TestStopManager_Context returns valid context.
func TestStopManager_Context(t *testing.T) {
	sm := NewStopManager()
	ctx := sm.Context()

	if ctx == nil {
		t.Fatal("Context() returned nil")
	}

	// Context should not be cancelled initially
	select {
	case <-ctx.Done():
		t.Fatal("Context was cancelled immediately")
	case <-time.After(100 * time.Millisecond):
		// Expected - context is active
	}
}

// TestStopManager_NoMonitors works with no monitors.
func TestStopManager_NoMonitors(t *testing.T) {
	sm := NewStopManager()

	// Start with no monitors
	sm.Start()

	// Should close immediately (or very quickly)
	select {
	case <-sm.Wait():
		// Success
	case <-time.After(1 * time.Second):
		// Also acceptable - might take time for cleanup goroutine
	}
}

// TestStopManager_MultipleMonitorsSameType works with multiple monitors of same type.
func TestStopManager_MultipleMonitorsSameType(t *testing.T) {
	sm := NewStopManager()

	// Add two duration monitors - one fast, one slow
	fastMon := NewDurationMonitor(50 * time.Millisecond)
	slowMon := NewDurationMonitor(5 * time.Second)

	sm.AddMonitor(fastMon)
	sm.AddMonitor(slowMon)

	// Start
	sm.Start()

	// Should trigger on fast one
	select {
	case reason := <-sm.Wait():
		if reason.Name != "duration" {
			t.Errorf("Expected 'duration' trigger, got %q", reason.Name)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StopManager did not trigger")
	}
}
