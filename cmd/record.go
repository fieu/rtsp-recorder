/*
Copyright © 2026 rtsp-recorder contributors

Record subcommand - records an RTSP stream to MP4.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rtsp-recorder/config"
	rrerrors "rtsp-recorder/internal/errors"
	"rtsp-recorder/internal/retry"
	"rtsp-recorder/internal/validator"
	"rtsp-recorder/logger"
	"rtsp-recorder/recorder"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// recordCmd represents the record command
var recordCmd = &cobra.Command{
	Use:   "record [RTSP_URL]",
	Short: "Record an RTSP stream to MP4",
	Long: `Record an RTSP stream to a timestamped MP4 file.

Uses ffmpeg for encoding and supports flexible stop conditions including:
- Manual interruption (Ctrl+C)
- Time limits (--duration)
- File size limits (--max-file-size)

If no URL is provided as argument, the URL from config file or --url flag is used.

Examples:
  # Record with URL as argument
  rtsp-recorder record rtsp://camera.local/stream

  # Record with flags
  rtsp-recorder record --duration 30m --max-file-size 500 rtsp://192.168.1.100:554/stream

  # Record with URL from config file
  rtsp-recorder record

  # Record with short flags
  rtsp-recorder record -d 15m -s 256 rtsp://camera.local/stream`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRecord,
}

func init() {
	rootCmd.AddCommand(recordCmd)

	// Register all configuration flags for the record command
	config.BindFlags(recordCmd)

	// Timelapse flag per D-49, D-52
	recordCmd.Flags().DurationP("timelapse", "l", 0, "Target output duration for timelapse (e.g., 10s, 1m). Requires --duration.")
	_ = viper.BindPFlag("timelapse_duration", recordCmd.Flags().Lookup("timelapse"))
}

func runRecord(cmd *cobra.Command, args []string) error {
	logger.Logger.Info().Msg("Starting rtsp-recorder")

	// Load configuration (flags have already been bound to viper)
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to load configuration: %w", err)
	}

	// If URL provided as positional argument, override config
	if len(args) > 0 {
		cfg.URL = args[0]
		logger.Logger.Info().Str("url", cfg.URL).Msg("Using URL from command line")
	}

	// Validate URL is present
	if cfg.URL == "" {
		return fmt.Errorf("[ERROR] No RTSP URL provided. Either:\n" +
			"  - Provide URL as argument: rtsp-recorder record rtsp://camera.local/stream\n" +
			"  - Set URL in config file: url: rtsp://camera.local/stream\n" +
			"  - Use --url flag: rtsp-recorder record --url rtsp://camera.local/stream")
	}

	// Validate timelapse configuration per D-51, D-55
	if err := validateTimelapseConfig(cfg); err != nil {
		return err
	}

	// Validate ffmpeg availability
	logger.Logger.Info().Msg("Checking FFmpeg installation")
	version, path, err := validator.ValidateFFmpeg()
	if err != nil {
		// err already has [ERROR] prefix from validator
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return fmt.Errorf("[ERROR] Cannot start recording: FFmpeg not available")
	}
	logger.Logger.Info().Str("path", path).Str("version", version).Msg("FFmpeg found")

	// Note: RTSP validation is now done inside the retry loop for fresh checks
	// Display configuration being used with structured logging
	logger.Logger.Info().
		Str("url", cfg.URL).
		Dur("duration", cfg.Duration).
		Int64("max_file_size_mb", cfg.MaxFileSize).
		Int("retry_attempts", cfg.RetryAttempts).
		Msg("Recording configuration")

	// Per D-59: Display timelapse info when enabled
	if cfg.TimelapseDuration > 0 {
		speedup := float64(cfg.Duration) / float64(cfg.TimelapseDuration)
		logger.Logger.Info().
			Float64("speedup", speedup).
			Dur("input_duration", cfg.Duration).
			Dur("output_duration", cfg.TimelapseDuration).
			Msg("Timelapse enabled")
		logger.Logger.Info().Msg("Audio disabled (timelapse mode)")
	}

	logger.Logger.Info().Msg("Starting recording")
	logger.Logger.Info().Msg("Press Ctrl+C to stop")

	// Create signal context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create and run recorder with retry logic
	rec := recorder.New(cfg, logger.Logger)

	// Create retry configuration
	retryCfg := retry.DefaultRetryConfig(cfg, logger.Logger)
	retryCfg.ShouldRetry = func(err error) bool {
		// Check if error is classified and retryable
		if classified, ok := err.(*rrerrors.ClassifiedError); ok {
			return rrerrors.IsRetryable(classified.Category)
		}
		return false // Non-classified errors fail immediately
	}

	// Execute recording with retry
	if err := retry.Retry(ctx, retryCfg, func() error {
		// Validate RTSP before each attempt (fresh check per D-34)
		if err := validator.ValidateRTSP(cfg.URL, 10*time.Second); err != nil {
			return err
		}
		return rec.Record(cfg.URL)
	}); err != nil {
		return err // Error already formatted by retry.OnFailure
	}

	return nil
}

// validateTimelapseConfig validates timelapse configuration settings.
// Per D-51: Timelapse requires duration to be set.
// Per D-55: Timelapse must be at least 1 second.
func validateTimelapseConfig(cfg *config.Config) error {
	if cfg.TimelapseDuration > 0 {
		// D-51: Timelapse requires duration
		if cfg.Duration == 0 {
			return fmt.Errorf("[ERROR] --timelapse requires --duration: cannot calculate speedup without recording duration")
		}
		// D-55: Minimum timelapse duration
		if cfg.TimelapseDuration < time.Second {
			return fmt.Errorf("[ERROR] --timelapse must be at least 1s")
		}
	}
	return nil
}
