/*
Copyright © 2026 rtsp-recorder contributors

Stop condition monitors for rtsp-recorder.
Provides signal, duration, and file size monitoring with context-based coordination.
*/
package recorder

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Monitor defines the interface for stop condition monitors.
// All monitors must be startable and provide a wait channel that closes
// when the stop condition is triggered.
type Monitor interface {
	// Start begins monitoring in a goroutine. Must be called before Wait().
	Start(ctx context.Context)
	// Wait returns a channel that closes when stop condition is triggered.
	// Safe to call before Start() - will block until Start() is called.
	Wait() <-chan struct{}
	// Name returns the monitor type name for logging and reporting.
	Name() string
}

// SignalMonitor watches for OS signals (SIGINT, SIGTERM) to trigger stop.
// Uses signal.NotifyContext per D-22 (buffered internally) to avoid
// the unbuffered channel pitfall described in PITFALLS.md §Pitfall 4.
type SignalMonitor struct {
	mu       sync.Mutex
	stop     chan struct{}
	stopFunc context.CancelFunc // from signal.NotifyContext
	started  bool
}

// NewSignalMonitor creates a new SignalMonitor.
func NewSignalMonitor() *SignalMonitor {
	return &SignalMonitor{
		stop: make(chan struct{}),
	}
}

// Start begins monitoring for signals in a goroutine.
// Uses signal.NotifyContext for proper signal handling per D-22.
func (m *SignalMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return
	}

	// Use signal.NotifyContext (buffered internally) per D-22 and PITFALLS.md §Pitfall 4
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	m.stopFunc = stop

	go func() {
		defer close(m.stop)
		select {
		case <-sigCtx.Done():
			// Signal received - channel will close
		case <-ctx.Done():
			// Parent context cancelled - clean exit
		}
	}()

	m.started = true
}

// Wait returns a channel that closes when a signal is received.
func (m *SignalMonitor) Wait() <-chan struct{} {
	return m.stop
}

// Name returns "signal".
func (m *SignalMonitor) Name() string {
	return "signal"
}
