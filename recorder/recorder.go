/*
Copyright © 2026 rtsp-recorder contributors

Recording orchestration for rtsp-recorder.
Coordinates FFmpeg process management, stop conditions, and real-time progress display.
*/
package recorder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"rtsp-recorder/config"
	"rtsp-recorder/ffmpeg"
	"rtsp-recorder/internal/utils"
)

// Recorder coordinates recording sessions with FFmpeg, stop conditions,
// and progress display. It provides a high-level API for recording RTSP streams.
type Recorder struct {
	config        *config.Config
	ffmpeg        *ffmpeg.Cmd
	outputPath    string
	startTime     time.Time
	bytesRecorded int64 // atomic
}

// New creates a new Recorder with the given configuration.
func New(cfg *config.Config) *Recorder {
	return &Recorder{
		config: cfg,
	}
}

// Record starts recording from the given RTSP URL.
// It validates the URL, generates an output filename, and starts the FFmpeg process.
// Recording continues until a stop condition is triggered (signal, duration, or file size).
func (r *Recorder) Record(url string) error {
	// Validate URL
	if url == "" {
		return fmt.Errorf("[ERROR] No RTSP URL provided")
	}

	// Generate output filename
	var filename string
	if r.config.FilenameTemplate != "" {
		filename = utils.GenerateFilenameFromTemplate(r.config.FilenameTemplate)
	} else {
		filename = utils.GenerateTimestampFilename() // REC-03
	}

	// Get current directory for output (REC-04)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to get current directory: %w", err)
	}

	r.outputPath = filepath.Join(cwd, filename)

	fmt.Printf("[INFO] Output file: %s\n", r.outputPath)

	// Create ffmpeg command
	r.ffmpeg = ffmpeg.New(r.config)

	// Create parent context
	ctx := context.Background()

	// Start recording
	if err := r.ffmpeg.Start(ctx, url, r.outputPath); err != nil {
		return fmt.Errorf("[ERROR] Failed to start ffmpeg: %w", err)
	}

	r.startTime = time.Now()

	// Recording will continue with stop conditions
	return r.runWithStopConditions(ctx)
}

// runWithStopConditions sets up all stop condition monitors and runs the progress loop.
// It waits for the first stop condition to trigger, then gracefully stops recording.
func (r *Recorder) runWithStopConditions(ctx context.Context) error {
	// Create stop manager
	sm := NewStopManager()

	// Add signal monitor (STOP-01)
	sm.AddMonitor(NewSignalMonitor())

	// Add duration monitor if configured (STOP-02)
	if r.config.Duration > 0 {
		sm.AddMonitor(NewDurationMonitor(r.config.Duration))
	}

	// Add file size monitor if configured (STOP-03)
	if r.config.MaxFileSize > 0 {
		sm.AddMonitor(NewFileSizeMonitor(r.config.MaxFileSize, r.outputPath))
	}

	// Start all monitors
	sm.Start()

	// Start progress display goroutine (REC-05, D-20, D-21)
	progressDone := make(chan struct{})
	go r.displayProgress(progressDone)

	// Wait for stop condition (STOP-04: first trigger wins)
	reason := <-sm.Wait()

	// Stop progress display
	close(progressDone)
	time.Sleep(100 * time.Millisecond) // Let final display flush

	fmt.Println() // New line after progress display
	fmt.Printf("[INFO] Stopping recording: %s\n", reason.Desc)

	// Stop ffmpeg gracefully
	if err := r.ffmpeg.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "[WARNING] FFmpeg stop error: %v\n", err)
	}

	// Print final summary (D-23)
	r.printFinalSummary()

	return nil
}

// displayProgress shows recording progress every 1 second.
// Updates display with file size, elapsed time, and bitrate.
func (r *Recorder) displayProgress(done <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second) // D-21
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			elapsed := time.Since(r.startTime)

			// Get current file size
			var size int64
			if info, err := os.Stat(r.outputPath); err == nil {
				size = info.Size()
				atomic.StoreInt64(&r.bytesRecorded, size)
			}

			// Calculate bitrate (bytes/seconds * 8 = bits per second)
			var bitrate string
			elapsedSecs := elapsed.Seconds()
			if elapsedSecs > 0 && size > 0 {
				bps := float64(size) * 8 / elapsedSecs
				bitrate = formatBitrate(bps)
			} else {
				bitrate = "0 Mbps"
			}

			// Format output per D-22: "Recording: 1.2GB | 00:05:30 | 4.5Mbps"
			fmt.Printf("\rRecording: %s | %s | %s",
				formatBytes(size),
				formatDuration(elapsed),
				bitrate,
			)
		}
	}
}

// printFinalSummary displays a formatted summary after recording completes.
func (r *Recorder) printFinalSummary() {
	// Get final file info
	info, err := os.Stat(r.outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARNING] Could not read output file: %v\n", err)
		return
	}

	duration := time.Since(r.startTime).Round(time.Second)
	size := info.Size()

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("  Recording Complete")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("  File:      %s\n", r.outputPath)
	fmt.Printf("  Size:      %s (%d bytes)\n", formatBytes(size), size)
	fmt.Printf("  Duration:  %s\n", formatDuration(duration))

	if duration.Seconds() > 0 {
		avgBitrate := float64(size) * 8 / duration.Seconds()
		fmt.Printf("  Avg Rate:  %s\n", formatBitrate(avgBitrate))
	}

	// Check file is valid (basic check - non-zero size)
	if size == 0 {
		fmt.Println("  Status:    [WARNING] Output file is empty")
	} else {
		fmt.Println("  Status:    [OK] Recording saved")
	}

	fmt.Println(strings.Repeat("=", 52))
}

// formatBytes converts bytes to human-readable format.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	switch exp {
	case 0:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(div))
	case 1:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(div))
	case 2:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(div))
	default:
		return fmt.Sprintf("%.1f TB", float64(bytes)/float64(div))
	}
}

// formatDuration formats a duration as HH:MM:SS.
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// formatBitrate formats bits per second as human-readable string.
func formatBitrate(bps float64) string {
	if bps < 1000 {
		return fmt.Sprintf("%.0f bps", bps)
	} else if bps < 1000000 {
		return fmt.Sprintf("%.1f Kbps", bps/1000)
	} else {
		return fmt.Sprintf("%.1f Mbps", bps/1000000)
	}
}
