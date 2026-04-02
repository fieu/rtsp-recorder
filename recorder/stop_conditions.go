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
	"time"
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

// DurationMonitor watches for a time duration to elapse before triggering stop.
// Uses Go time.Timer per D-27/D-29 and PITFALLS.md §Pitfall 8 (avoiding ffmpeg -t inaccuracy).
type DurationMonitor struct {
	mu       sync.Mutex
	duration time.Duration // 0 = unlimited
	timer    *time.Timer
	stop     chan struct{}
	started  bool
}

// NewDurationMonitor creates a new DurationMonitor with the specified duration.
// If duration is 0 or negative, the monitor triggers immediately (unlimited).
func NewDurationMonitor(duration time.Duration) *DurationMonitor {
	return &DurationMonitor{
		duration: duration,
		stop:     make(chan struct{}),
	}
}

// Start begins the duration timer in a goroutine.
func (m *DurationMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return
	}

	if m.duration <= 0 {
		// Unlimited duration - close channel immediately
		close(m.stop)
	} else {
		// Start timer that will close stop channel when duration elapses
		m.timer = time.AfterFunc(m.duration, func() {
			close(m.stop)
		})

		// Cleanup goroutine for context cancellation
		go func() {
			<-ctx.Done()
			m.mu.Lock()
			defer m.mu.Unlock()
			if m.timer != nil {
				m.timer.Stop()
			}
			// If stop not already closed, close it
			select {
			case <-m.stop:
			default:
				close(m.stop)
			}
		}()
	}

	m.started = true
}

// Wait returns a channel that closes when the duration elapses.
func (m *DurationMonitor) Wait() <-chan struct{} {
	return m.stop
}

// Name returns "duration".
func (m *DurationMonitor) Name() string {
	return "duration"
}

// Duration returns the configured duration.
func (m *DurationMonitor) Duration() time.Duration {
	return m.duration
}
