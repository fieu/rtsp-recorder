/*
Copyright © 2026 rtsp-recorder contributors

Configuration management for rtsp-recorder.
Provides Config struct and loading utilities.
*/
package config

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds all configuration settings for rtsp-recorder.
// Fields are tagged with mapstructure for Viper unmarshaling.
type Config struct {
	// URL is the RTSP stream URL to record
	URL string `mapstructure:"url"`

	// Duration is the maximum recording duration (0 = unlimited)
	Duration time.Duration `mapstructure:"duration"`

	// TimelapseDuration is the target output duration for timelapse (0 = no timelapse)
	TimelapseDuration time.Duration `mapstructure:"timelapse_duration"`

	// MaxFileSize is the maximum file size in MB before stopping (0 = unlimited)
	MaxFileSize int64 `mapstructure:"max_file_size"`

	// RetryAttempts is the number of retry attempts on connection failure
	RetryAttempts int `mapstructure:"retry_attempts"`

	// FFmpegPath is the path to the ffmpeg binary (empty = search PATH)
	FFmpegPath string `mapstructure:"ffmpeg_path"`

	// FilenameTemplate is the template for output filenames
	// Supports {{.Timestamp}} placeholder
	FilenameTemplate string `mapstructure:"filename_template"`
}

// Load reads configuration from Viper and returns a Config struct.
// This function:
//   - Uses Viper's already-loaded values (set in cmd/root.go initConfig)
//   - Returns ConfigFileNotFoundError as non-fatal (config file is optional per D-06)
//   - Returns other errors (parse errors) as fatal
//
// The caller should handle the returned error appropriately.
func Load() (*Config, error) {
	var cfg Config

	// Unmarshal Viper's configuration into the struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// GetFFmpegPath returns the path to use for ffmpeg.
// If FFmpegPath is set in config, it returns that value.
// Otherwise, it returns "ffmpeg" for PATH lookup.
func (c *Config) GetFFmpegPath() string {
	if c.FFmpegPath != "" {
		return c.FFmpegPath
	}
	return "ffmpeg"
}

// FindFFmpeg attempts to locate the ffmpeg binary in PATH.
// Returns the full path if found, or an error if not found.
func FindFFmpeg() (string, error) {
	return exec.LookPath("ffmpeg")
}

// DefaultConfig returns a Config with the default values.
// This is useful for testing or when you need defaults without Viper.
func DefaultConfig() *Config {
	return &Config{
		Duration:          60 * time.Minute,
		TimelapseDuration: 0, // Disabled by default
		MaxFileSize:       1024,
		RetryAttempts:     3,
		FFmpegPath:        "ffmpeg",
		FilenameTemplate:  "recording_{{.Timestamp}}.mp4",
	}
}

// BindFlags registers all configuration flags for the given command.
// It creates both long and short flag forms and binds them to Viper.
// This enables the configuration precedence: flags > env > config > defaults.
//
// Flags registered:
//   - --url, -u: RTSP stream URL
//   - --duration, -d: Maximum recording duration
//   - --max-file-size, -s: Maximum file size in MB
//   - --retry-attempts, -r: Number of retry attempts
//   - --ffmpeg-path, -f: Path to ffmpeg binary
//   - --filename-template, -t: Output filename template
func BindFlags(cmd *cobra.Command) {
	// URL flag
	cmd.Flags().StringP("url", "u", "", "RTSP stream URL to record (required if not in config file)")
	viper.BindPFlag("url", cmd.Flags().Lookup("url"))

	// Duration flag (60m default)
	cmd.Flags().DurationP("duration", "d", 60*time.Minute, "Maximum recording duration (e.g., 30m, 1h, 0=unlimited)")
	viper.BindPFlag("duration", cmd.Flags().Lookup("duration"))

	// Max file size flag (1024MB default)
	cmd.Flags().Int64P("max-file-size", "s", 1024, "Maximum file size in MB before stopping (0=unlimited)")
	viper.BindPFlag("max_file_size", cmd.Flags().Lookup("max-file-size"))

	// Retry attempts flag (3 default)
	cmd.Flags().IntP("retry-attempts", "r", 3, "Number of retry attempts on connection failure")
	viper.BindPFlag("retry_attempts", cmd.Flags().Lookup("retry-attempts"))

	// FFmpeg path flag
	cmd.Flags().StringP("ffmpeg-path", "f", "", "Path to ffmpeg binary (default: search PATH)")
	viper.BindPFlag("ffmpeg_path", cmd.Flags().Lookup("ffmpeg-path"))

	// Filename template flag
	cmd.Flags().StringP("filename-template", "t", "", "Output filename template (default: recording_{{.Timestamp}}.mp4)")
	viper.BindPFlag("filename_template", cmd.Flags().Lookup("filename-template"))
}
