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
	fmt.Println("[INFO] Validating rtsp-recorder configuration...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("[ERROR] Configuration: %w", err)
	}

	fmt.Println("[INFO] Configuration loaded successfully")

	// Display configuration summary
	fmt.Println("\n[INFO] Current configuration:")
	fmt.Printf("  URL: %s\n", valueOrDefault(cfg.URL, "(not set - will require --url flag)"))
	fmt.Printf("  Duration: %v\n", cfg.Duration)
	fmt.Printf("  Max File Size: %d MB\n", cfg.MaxFileSize)
	fmt.Printf("  Retry Attempts: %d\n", cfg.RetryAttempts)
	fmt.Printf("  FFmpeg Path: %s\n", cfg.GetFFmpegPath())

	// Validate FFmpeg
	fmt.Println("\n[INFO] Checking FFmpeg installation...")
	version, path, err := validator.ValidateFFmpeg()
	if err != nil {
		// err already has [ERROR] prefix from validator
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return fmt.Errorf("[ERROR] Validation failed: FFmpeg not available")
	}

	fmt.Printf("[INFO] FFmpeg found: %s (version %s)\n", path, version)

	fmt.Println("\n[INFO] Validation completed successfully!")
	return nil
}

// valueOrDefault returns value if non-empty, otherwise returns defaultStr
func valueOrDefault(value, defaultStr string) string {
	if value == "" {
		return defaultStr
	}
	return value
}
