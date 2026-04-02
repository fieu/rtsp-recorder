/*
Copyright © 2026 rtsp-recorder contributors

Record subcommand - records an RTSP stream to MP4.
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"rtsp-recorder/config"
	"rtsp-recorder/internal/validator"
	"rtsp-recorder/recorder"
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
}

func runRecord(cmd *cobra.Command, args []string) error {
	fmt.Println("[INFO] Starting rtsp-recorder...")

	// Load configuration (flags have already been bound to viper)
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to load configuration: %w", err)
	}

	// If URL provided as positional argument, override config
	if len(args) > 0 {
		cfg.URL = args[0]
		fmt.Printf("[INFO] Using URL from command line: %s\n", cfg.URL)
	}

	// Validate URL is present
	if cfg.URL == "" {
		return fmt.Errorf("[ERROR] No RTSP URL provided. Either:\n" +
			"  - Provide URL as argument: rtsp-recorder record rtsp://camera.local/stream\n" +
			"  - Set URL in config file: url: rtsp://camera.local/stream\n" +
			"  - Use --url flag: rtsp-recorder record --url rtsp://camera.local/stream")
	}

	// Validate ffmpeg availability
	fmt.Println("[INFO] Checking FFmpeg installation...")
	version, path, err := validator.ValidateFFmpeg()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return fmt.Errorf("[ERROR] Cannot start recording: FFmpeg not available")
	}
	fmt.Printf("[INFO] FFmpeg found: %s (version %s)\n", path, version)

	// Validate RTSP stream is accessible before starting (ERR-02)
	fmt.Println("[INFO] Validating RTSP stream...")
	if err := validator.ValidateRTSP(cfg.URL, 10*time.Second); err != nil {
		return err // Already formatted with [ERROR] prefix
	}
	fmt.Println("[INFO] RTSP stream validated successfully")

	// Display configuration being used
	fmt.Println("\n[INFO] Recording configuration:")
	fmt.Printf("  URL: %s\n", cfg.URL)
	fmt.Printf("  Duration: %v\n", cfg.Duration)
	if cfg.MaxFileSize > 0 {
		fmt.Printf("  Max File Size: %d MB\n", cfg.MaxFileSize)
	} else {
		fmt.Println("  Max File Size: unlimited")
	}
	if cfg.RetryAttempts > 0 {
		fmt.Printf("  Retry Attempts: %d\n", cfg.RetryAttempts)
	}

	fmt.Println()
	fmt.Println("[INFO] Starting recording...")
	fmt.Println("[INFO] Press Ctrl+C to stop")

	// Create and run recorder
	rec := recorder.New(cfg)
	if err := rec.Record(cfg.URL); err != nil {
		return err
	}

	return nil
}
