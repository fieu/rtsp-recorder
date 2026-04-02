/*
Copyright © 2026 rtsp-recorder contributors

A CLI tool for recording RTSP video streams to MP4 files.
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rtsp-recorder",
	Short: "Record RTSP video streams to MP4 files",
	Long: `rtsp-recorder is a CLI tool that records RTSP video streams to MP4 files.

Uses ffmpeg for encoding and supports flexible stop conditions including:
- Manual interruption (Ctrl+C)
- Time limits (--duration)
- File size limits (--max-file-size)

Configuration can be provided via:
1. Command-line flags (highest priority)
2. Environment variables (RTSP_RECORDER_*)
3. Config file: rtsp-recorder.yml in current directory
4. Default values (lowest priority)

Example rtsp-recorder.yml:
  url: rtsp://192.168.1.100:554/stream
  duration: 60m
  max_file_size: 1024
  retry_attempts: 3
  ffmpeg_path: ffmpeg
  filename_template: "recording_{{.Timestamp}}.mp4"`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags available to all subcommands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default: rtsp-recorder.yml in current directory)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Explicitly look for rtsp-recorder.yml to avoid finding the binary
		viper.SetConfigFile("./rtsp-recorder.yml")
	}

	// Set environment variable prefix
	viper.SetEnvPrefix("RTSP_RECORDER")
	viper.AutomaticEnv()

	// Set conservative defaults (per D-05)
	viper.SetDefault("duration", 60*time.Minute)
	viper.SetDefault("max_file_size", 1024) // MB
	viper.SetDefault("retry_attempts", 3)
	viper.SetDefault("ffmpeg_path", "ffmpeg")
	viper.SetDefault("filename_template", "recording_{{.Timestamp}}.mp4")

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional (per D-06), only report if it's not a "not found" error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Check if it's a file not found error by checking the error string
			if os.IsNotExist(err) || isFileNotFoundError(err) {
				// Config file doesn't exist, which is OK
				return
			}
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to read config file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "[INFO] Using config file: %s\n", viper.ConfigFileUsed())
	}
}

// isFileNotFoundError checks if an error is related to file not being found
func isFileNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "no such file") ||
		contains(errStr, "cannot find") ||
		contains(errStr, "not found")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) <= len(s) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
