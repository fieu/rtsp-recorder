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

	"github.com/rs/zerolog"
)

// Recorder coordinates recording sessions with FFmpeg, stop conditions,
// and progress display. It provides a high-level API for recording RTSP streams.
type Recorder struct {
	config        *config.Config
	ffmpeg        *ffmpeg.Cmd
	outputPath    string
	startTime     time.Time
	bytesRecorded int64 // atomic
	logger        zerolog.Logger
}

// New creates a new Recorder with the given configuration.
func New(cfg *config.Config, logger zerolog.Logger) *Recorder {
	return &Recorder{
		config: cfg,
		logger: logger,
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

	r.logger.Info().Str("path", r.outputPath).Msg("Output file")

	// Create ffmpeg command
	r.ffmpeg = ffmpeg.New(r.config)

	// Debug logging for ffmpeg configuration
	r.logger.Debug().
		Str("ffmpeg_path", r.config.GetFFmpegPath()).
		Bool("timelapse_enabled", r.config.TimelapseDuration > 0).
		Msg("FFmpeg command created")

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
	// Ticker stopped via defer in displayProgress()

	r.logger.Info().Str("reason", reason.Desc).Msg("Stopping recording")

	// Stop ffmpeg gracefully
	if err := r.ffmpeg.Stop(); err != nil {
		r.logger.Warn().Err(err).Msg("FFmpeg stop error")
	}

	// Print final summary (D-23)
	r.printFinalSummary()

	return nil
}

// displayProgress logs recording progress periodically using structured logging.
// Per D-96, D-97, D-104: Uses zerolog instead of \r overwrite.
// Per D-108: Uses time.Ticker with configurable interval.
// Per D-110: Logs immediately at start (0 seconds).
// Per D-102: Skips entirely if ProgressInterval is 0.
func (r *Recorder) displayProgress(done <-chan struct{}) {
	// Per D-102: Skip entirely if interval is 0
	if r.config.ProgressInterval <= 0 {
		return
	}

	ticker := time.NewTicker(r.config.ProgressInterval) // Per D-108: configurable interval
	defer ticker.Stop()

	// Per D-110: Log immediately at start (0 seconds)
	r.logProgress()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			r.logProgress()
		}
	}
}

// logProgress outputs a single structured progress log entry.
// Per D-103: Includes elapsed time, bytes, file size, bitrate.
// Per D-104: Uses structured zerolog fields.
func (r *Recorder) logProgress() {
	elapsed := time.Since(r.startTime)

	// Get current file size
	var size int64
	if info, err := os.Stat(r.outputPath); err == nil {
		size = info.Size()
		atomic.StoreInt64(&r.bytesRecorded, size)
	}

	bytesStr := formatBytes(size)

	// Build log event with structured fields per D-104
	event := r.logger.Info().
		Dur("elapsed", elapsed).
		Int64("bytes", size).
		Str("size", bytesStr)

	// Add bitrate if calculable per D-103
	elapsedSecs := elapsed.Seconds()
	if elapsedSecs > 0 && size > 0 {
		bps := float64(size) * 8 / elapsedSecs
		event = event.
			Float64("bitrate_bps", bps).
			Str("bitrate", formatBitrate(bps))
	}

	// Add timelapse fields when active per D-103
	if r.ffmpeg != nil {
		speedup := r.ffmpeg.GetSpeedupFactor()
		if speedup > 1 {
			estimatedOutput := time.Duration(float64(elapsed) / speedup)
			event = event.
				Float64("speedup", speedup).
				Dur("output_duration", estimatedOutput)
		}
	}

	event.Msg("Recording progress")
}

// printFinalSummary displays a formatted summary after recording completes.
// Per D-59: Includes timelapse info when enabled (real duration, output duration, speedup).
func (r *Recorder) printFinalSummary() {
	// Get final file info
	info, err := os.Stat(r.outputPath)
	if err != nil {
		r.logger.Warn().Err(err).Msg("Could not read output file")
		return
	}

	realDuration := time.Since(r.startTime).Round(time.Second)
	size := info.Size()

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("  Recording Complete")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("  File:      %s\n", r.outputPath)
	fmt.Printf("  Size:      %s (%d bytes)\n", formatBytes(size), size)
	fmt.Printf("  Duration:  %s (real)\n", formatDuration(realDuration))

	// Per D-59: Show timelapse summary when enabled
	if r.ffmpeg != nil && r.ffmpeg.GetSpeedupFactor() > 1 {
		speedup := r.ffmpeg.GetSpeedupFactor()
		outputDuration := time.Duration(float64(realDuration) / speedup).Round(time.Second)
		fmt.Printf("  Output:    %s (timelapse)\n", formatDuration(outputDuration))
		fmt.Printf("  Speedup:   %.0fx\n", speedup)
	}

	if realDuration.Seconds() > 0 {
		avgBitrate := float64(size) * 8 / realDuration.Seconds()
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
	switch {
	case bps < 1000:
		return fmt.Sprintf("%.0f bps", bps)
	case bps < 1000000:
		return fmt.Sprintf("%.1f Kbps", bps/1000)
	default:
		return fmt.Sprintf("%.1f Mbps", bps/1000000)
	}
}
