/*
Copyright © 2026 rtsp-recorder contributors

Validate subcommand - checks configuration and dependencies.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"rtsp-recorder/config"
	"rtsp-recorder/internal/validator"
	"rtsp-recorder/logger"
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
	logger.Logger.Info().Msg("Validating rtsp-recorder configuration")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("[ERROR] Configuration: %w", err)
	}

	logger.Logger.Info().Msg("Configuration loaded successfully")

	// Display configuration summary with structured logging
	logger.Logger.Info().
		Str("url", valueOrDefault(cfg.URL, "(not set - will require --url flag)")).
		Dur("duration", cfg.Duration).
		Int64("max_file_size_mb", cfg.MaxFileSize).
		Int("retry_attempts", cfg.RetryAttempts).
		Str("ffmpeg_path", cfg.GetFFmpegPath()).
		Msg("Current configuration")

	// Validate FFmpeg
	logger.Logger.Info().Msg("Checking FFmpeg installation")
	version, path, err := validator.ValidateFFmpeg()
	if err != nil {
		// err already has [ERROR] prefix from validator
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return fmt.Errorf("[ERROR] Validation failed: FFmpeg not available")
	}

	logger.Logger.Info().Str("path", path).Str("version", version).Msg("FFmpeg found")

	logger.Logger.Info().Msg("Validation completed successfully")
	return nil
}

// valueOrDefault returns value if non-empty, otherwise returns defaultStr
func valueOrDefault(value, defaultStr string) string {
	if value == "" {
		return defaultStr
	}
	return value
}
