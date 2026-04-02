/*
Copyright © 2026 rtsp-recorder contributors

Validate subcommand - checks configuration and dependencies.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"rtsp-recorder/config"
	"rtsp-recorder/internal/validator"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and dependencies",
	Long: `Validate that rtsp-recorder is properly configured and all dependencies are available.

This command checks:
- Configuration file (rtsp-recorder.yml) if present
- Environment variables
- FFmpeg installation and version

Examples:
  # Validate with default config file
  rtsp-recorder validate

  # Validate with specific config file
  rtsp-recorder validate --config /path/to/config.yml

  # Validate with environment variables
  RTSP_RECORDER_DURATION=30m rtsp-recorder validate`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	Logger.Info("Validating rtsp-recorder configuration")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("[ERROR] Configuration: %w", err)
	}

	Logger.Info("Configuration loaded successfully")

	// Display configuration summary with structured logging
	Logger.Info("Current configuration",
		zap.String("url", valueOrDefault(cfg.URL, "(not set - will require --url flag)")),
		zap.Duration("duration", cfg.Duration),
		zap.Int64("max_file_size_mb", cfg.MaxFileSize),
		zap.Int("retry_attempts", cfg.RetryAttempts),
		zap.String("ffmpeg_path", cfg.GetFFmpegPath()),
	)

	// Validate FFmpeg
	Logger.Info("Checking FFmpeg installation")
	version, path, err := validator.ValidateFFmpeg()
	if err != nil {
		// err already has [ERROR] prefix from validator
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return fmt.Errorf("[ERROR] Validation failed: FFmpeg not available")
	}

	Logger.Info("FFmpeg found", zap.String("path", path), zap.String("version", version))

	Logger.Info("Validation completed successfully")
	return nil
}

// valueOrDefault returns value if non-empty, otherwise returns defaultStr
func valueOrDefault(value, defaultStr string) string {
	if value == "" {
		return defaultStr
	}
	return value
}
