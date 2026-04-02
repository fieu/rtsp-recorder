/*
Copyright © 2026 rtsp-recorder contributors

Tests for stop condition monitors.
*/
package recorder

import (
	"context"
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
