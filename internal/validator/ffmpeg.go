/*
Copyright © 2026 rtsp-recorder contributors

FFmpeg validation utilities.
Provides pre-flight checks for ffmpeg availability.
*/
package validator

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// CheckFFmpeg verifies that ffmpeg is available in PATH.
// Returns the full path to ffmpeg if found, or an error with [ERROR] prefix if not.
// The error message includes actionable installation instructions.
func CheckFFmpeg() (string, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("[ERROR] ffmpeg: not found in PATH. Install: apt install ffmpeg (Debian/Ubuntu), brew install ffmpeg (macOS), or download from https://ffmpeg.org/download.html")
	}
	return path, nil
}

// CheckFFmpegVersion runs ffmpeg and parses its version string.
// Returns the version string (e.g., "7.1") and an error if the check fails.
// On success, returns version and nil error.
// On failure, returns empty string and error with [ERROR] prefix.
func CheckFFmpegVersion() (string, error) {
	// Run ffmpeg -version to get version info
	cmd := exec.Command("ffmpeg", "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("[ERROR] ffmpeg: failed to run ffmpeg -version: %w", err)
	}

	// Parse version from first line: "ffmpeg version 7.1 Copyright..."
	version, err := parseFFmpegVersion(string(output))
	if err != nil {
		return "", fmt.Errorf("[ERROR] ffmpeg: %w", err)
	}

	return version, nil
}

// parseFFmpegVersion extracts the version string from ffmpeg -version output.
// Expected format: "ffmpeg version X.Y.Z ..." or "ffmpeg version X.Y ..."
func parseFFmpegVersion(output string) (string, error) {
	// Get first line
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("empty ffmpeg -version output")
	}
	firstLine := lines[0]

	// Extract version using regex: "ffmpeg version (\\d+\\.\\d+(?:\\.\\d+)?)"
	re := regexp.MustCompile(`ffmpeg version (\d+\.\d+(?:\.\d+)?)`)
	matches := re.FindStringSubmatch(firstLine)
	if len(matches) < 2 {
		return "", fmt.Errorf("unable to parse version from: %s", firstLine)
	}

	return matches[1], nil
}

// ValidateFFmpeg performs all ffmpeg validation checks.
// Returns the version string and full path to ffmpeg on success.
// Returns empty strings and error on failure (with [ERROR] prefix).
func ValidateFFmpeg() (string, string, error) {
	// First check if ffmpeg exists
	path, err := CheckFFmpeg()
	if err != nil {
		return "", "", err
	}

	// Then check version
	version, err := CheckFFmpegVersion()
	if err != nil {
		return "", "", err
	}

	return version, path, nil
}
