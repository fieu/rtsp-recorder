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

// FileSizeMonitor polls file size periodically and triggers stop when limit is reached.
// Polls every 1 second per D-27. Updates currentSize for progress reporting (REC-05).
type FileSizeMonitor struct {
	mu          sync.Mutex
	maxBytes    int64 // 0 = unlimited
	filePath    string
	ticker      *time.Ticker
	stop        chan struct{}
	started     bool
	currentSize int64 // for progress reporting
}

// NewFileSizeMonitor creates a new FileSizeMonitor.
// maxSizeMB is converted to bytes (maxSizeMB * 1024 * 1024).
// If maxSizeMB is 0, the monitor triggers immediately (unlimited).
func NewFileSizeMonitor(maxSizeMB int64, filePath string) *FileSizeMonitor {
	var maxBytes int64
	if maxSizeMB > 0 {
		maxBytes = maxSizeMB * 1024 * 1024
	}
	return &FileSizeMonitor{
		maxBytes: maxBytes,
		filePath: filePath,
		stop:     make(chan struct{}),
	}
}

// Start begins polling file size in a goroutine.
func (m *FileSizeMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return
	}

	if m.maxBytes <= 0 {
		// Unlimited file size - close channel immediately
		close(m.stop)
	} else {
		// Start polling goroutine
		m.ticker = time.NewTicker(1 * time.Second) // D-27: poll every 1 second
		go func() {
			defer close(m.stop)
			defer m.ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-m.ticker.C:
					size, err := m.checkSize()
					if err != nil {
						// File may not exist yet, continue polling
						continue
					}
					m.mu.Lock()
					m.currentSize = size
					m.mu.Unlock()

					if m.maxBytes > 0 && size >= m.maxBytes {
						return // Channel will close via defer
					}
				}
			}
		}()
	}

	m.started = true
}

// checkSize returns the current file size.
func (m *FileSizeMonitor) checkSize() (int64, error) {
	stat, err := os.Stat(m.filePath)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// Wait returns a channel that closes when file size reaches limit.
func (m *FileSizeMonitor) Wait() <-chan struct{} {
	return m.stop
}

// Name returns "file_size".
func (m *FileSizeMonitor) Name() string {
	return "file_size"
}

// CurrentSize returns the last polled file size (for progress reporting).
func (m *FileSizeMonitor) CurrentSize() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentSize
}

// MaxBytes returns the configured max file size in bytes.
func (m *FileSizeMonitor) MaxBytes() int64 {
	return m.maxBytes
}

// StopReason describes why recording was stopped.
type StopReason struct {
	Name string // Which monitor triggered: "signal", "duration", "file_size"
	Desc string // Human-readable description
}

// StopManager coordinates multiple stop condition monitors.
// Implements "first trigger wins" logic per D-17: any one stopping condition
// causes all others to stop. Uses Go context for cancellation propagation.
type StopManager struct {
	mu       sync.Mutex
	monitors []Monitor
	stop     chan StopReason
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewStopManager creates a new StopManager with background context.
func NewStopManager() *StopManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &StopManager{
		stop:   make(chan StopReason, 1), // Buffered to prevent blocking
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddMonitor adds a monitor to be coordinated.
func (sm *StopManager) AddMonitor(m Monitor) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.monitors = append(sm.monitors, m)
}

// Start begins all monitors and the coordination goroutine.
// First monitor to trigger wins per D-17.
func (sm *StopManager) Start() {
	var wg sync.WaitGroup

	// Start all monitors
	for _, m := range sm.monitors {
		m.Start(sm.ctx)
	}

	// Start a goroutine for each monitor to watch for triggers
	for _, m := range sm.monitors {
		wg.Add(1)
		go func(mon Monitor) {
			defer wg.Done()
			select {
			case <-mon.Wait():
				// Try to send stop reason (non-blocking)
				select {
				case sm.stop <- StopReason{Name: mon.Name(), Desc: sm.describeReason(mon.Name())}:
					sm.cancel() // Cancel context to signal other monitors
				default:
					// Another monitor already triggered, ignore
				}
			case <-sm.ctx.Done():
				return
			}
		}(m)
	}

	// Cleanup goroutine to close stop channel when all monitors are done
	go func() {
		wg.Wait()
		close(sm.stop)
	}()
}

// describeReason returns a human-readable description for a monitor name.
func (sm *StopManager) describeReason(name string) string {
	switch name {
	case "signal":
		return "Interrupted by user (Ctrl+C)"
	case "duration":
		return "Duration limit reached"
	case "file_size":
		return "File size limit reached"
	default:
		return "Stop condition triggered"
	}
}

// Wait returns a channel that receives the StopReason when any monitor triggers.
func (sm *StopManager) Wait() <-chan StopReason {
	return sm.stop
}

// Stop manually cancels the context to stop all monitors.
func (sm *StopManager) Stop() {
	sm.cancel()
}

// Context returns the internal context for use with ffmpeg CommandContext.
func (sm *StopManager) Context() context.Context {
	return sm.ctx
}
