/*
Copyright © 2026 rtsp-recorder contributors

File utilities for rtsp-recorder.
Provides filename generation and sanitization utilities.
*/
package utils

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// GenerateTimestampFilename returns a filename with the current timestamp.
// Format: "2006-01-02-15-04-05.mp4" (YYYY-MM-DD-HH-MM-SS.mp4)
// This format provides automatic chronological sorting and is human-readable.
//
// Example output: "2025-04-02-14-30-45.mp4"
func GenerateTimestampFilename() string {
	return time.Now().Format("2006-01-02-15-04-05") + ".mp4"
}

// GenerateFilenameFromTemplate creates a filename using the provided template.
// The template supports the {{.Timestamp}} placeholder which is replaced
// with the current timestamp in YYYY-MM-DD-HH-MM-SS format.
//
// If template is empty, falls back to GenerateTimestampFilename().
//
// Example:
//
//	template: "camera_recording_{{.Timestamp}}.mp4"
//	result:   "camera_recording_2025-04-02-14-30-45.mp4"
func GenerateFilenameFromTemplate(template string) string {
	if template == "" {
		return GenerateTimestampFilename()
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := strings.ReplaceAll(template, "{{.Timestamp}}", timestamp)

	// Sanitize the resulting filename
	return SanitizeFilename(filename)
}

// SanitizeFilename removes or replaces characters that are invalid in filenames.
// It handles:
//   - Path traversal attempts (../, ./)
//   - Invalid filename characters on common filesystems
//   - Multiple consecutive separators
//   - Leading/trailing spaces and separators
//
// The function ensures the filename is safe for cross-platform use.
func SanitizeFilename(name string) string {
	if name == "" {
		return ""
	}

	// Remove path traversal attempts
	name = filepath.Base(name)

	// Replace invalid characters with underscore
	// Invalid chars on Windows: < > : " / \ | ? *
	// Invalid chars on Unix: / and null
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	name = invalidChars.ReplaceAllString(name, "_")

	// Replace multiple consecutive underscores with single
	multipleUnderscores := regexp.MustCompile(`_+`)
	name = multipleUnderscores.ReplaceAllString(name, "_")

	// Trim leading/trailing spaces, dots, and underscores
	name = strings.Trim(name, " ._")

	// Ensure we have something left
	if name == "" {
		return "recording"
	}

	return name
}

// IsValidMP4Extension checks if the filename has a valid MP4 extension.
// Valid extensions: .mp4, .m4v (case-insensitive)
func IsValidMP4Extension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".mp4" || ext == ".m4v"
}

// EnsureMP4Extension adds .mp4 extension if the filename doesn't have
// a valid video extension already.
func EnsureMP4Extension(filename string) string {
	if IsValidMP4Extension(filename) {
		return filename
	}
	return filename + ".mp4"
}
