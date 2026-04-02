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

	"github.com/spf13/viper"
)

// Config holds all configuration settings for rtsp-recorder.
// Fields are tagged with mapstructure for Viper unmarshaling.
type Config struct {
	// URL is the RTSP stream URL to record
	URL string `mapstructure:"url"`

	// Duration is the maximum recording duration (0 = unlimited)
	Duration time.Duration `mapstructure:"duration"`

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
		Duration:         60 * time.Minute,
		MaxFileSize:      1024,
		RetryAttempts:    3,
		FFmpegPath:       "ffmpeg",
		FilenameTemplate: "recording_{{.Timestamp}}.mp4",
	}
}
