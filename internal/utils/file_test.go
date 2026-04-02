/*
Copyright © 2026 rtsp-recorder contributors

Tests for file utilities.
*/
package utils

import (
	"regexp"
	"strings"
	"testing"
)

func TestGenerateTimestampFilename(t *testing.T) {
	filename := GenerateTimestampFilename()

	// Should end with .mp4
	if !strings.HasSuffix(filename, ".mp4") {
		t.Errorf("GenerateTimestampFilename() = %v, want suffix .mp4", filename)
	}

	// Should match expected pattern: YYYY-MM-DD-HH-MM-SS.mp4
	// Example: 2025-04-02-14-30-45.mp4
	pattern := `^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}\.mp4$`
	matched, err := regexp.MatchString(pattern, filename)
	if err != nil {
		t.Fatalf("Regex error: %v", err)
	}
	if !matched {
		t.Errorf("GenerateTimestampFilename() = %v, want format YYYY-MM-DD-HH-MM-SS.mp4", filename)
	}
}

func TestGenerateFilenameFromTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantExt  string
		wantPart string
	}{
		{
			name:     "empty template falls back to timestamp",
			template: "",
			wantExt:  ".mp4",
			wantPart: "",
		},
		{
			name:     "template with timestamp placeholder",
			template: "recording_{{.Timestamp}}.mp4",
			wantExt:  ".mp4",
			wantPart: "recording_",
		},
		{
			name:     "template with custom prefix",
			template: "camera_{{.Timestamp}}_feed.mp4",
			wantExt:  ".mp4",
			wantPart: "camera_",
		},
		{
			name:     "template without placeholder",
			template: "static_name.mp4",
			wantExt:  ".mp4",
			wantPart: "static_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFilenameFromTemplate(tt.template)

			// Check extension
			if !strings.HasSuffix(got, tt.wantExt) {
				t.Errorf("GenerateFilenameFromTemplate(%q) = %v, want suffix %v", tt.template, got, tt.wantExt)
			}

			// Check for expected part
			if tt.wantPart != "" && !strings.HasPrefix(got, tt.wantPart) {
				t.Errorf("GenerateFilenameFromTemplate(%q) = %v, want prefix %v", tt.template, got, tt.wantPart)
			}

			// If template has placeholder, verify it was replaced (not literal {{.Timestamp}})
			if strings.Contains(tt.template, "{{.Timestamp}}") {
				if strings.Contains(got, "{{.Timestamp}}") {
					t.Errorf("GenerateFilenameFromTemplate(%q) = %v, placeholder not replaced", tt.template, got)
				}
				// Verify timestamp format was injected (should have numbers in timestamp position)
				if !strings.Contains(got, "-") {
					t.Errorf("GenerateFilenameFromTemplate(%q) = %v, timestamp format incorrect", tt.template, got)
				}
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid filename unchanged",
			input:    "recording_2025-04-02.mp4",
			expected: "recording_2025-04-02.mp4",
		},
		{
			name:     "removes path traversal",
			input:    "../../etc/passwd.mp4",
			expected: "passwd.mp4",
		},
		{
			name:     "removes invalid windows chars",
			input:    "recording<>:\"/\\|?*.mp4",
			expected: "mp4", // underscores before .mp4 get trimmed
		},
		{
			name:     "replaces multiple invalid chars",
			input:    "file<name>with:invalid.mp4",
			expected: "file_name_with_invalid.mp4",
		},
		{
			name:     "trims leading dots and spaces",
			input:    "  ...hidden_file.mp4",
			expected: "hidden_file.mp4",
		},
		{
			name:     "trims trailing dots and spaces",
			input:    "file.mp4...  ",
			expected: "file.mp4",
		},
		{
			name:     "collapses multiple underscores",
			input:    "file___name____test.mp4",
			expected: "file_name_test.mp4",
		},
		{
			name:     "empty string returns default",
			input:    "",
			expected: "",
		},
		{
			name:     "only invalid chars returns recording",
			input:    "<>|",
			expected: "recording",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidMP4Extension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"mp4 lowercase", "video.mp4", true},
		{"mp4 uppercase", "video.MP4", true},
		{"mp4 mixed case", "video.Mp4", true},
		{"m4v extension", "video.m4v", true},
		{"m4v uppercase", "video.M4V", true},
		{"avi extension", "video.avi", false},
		{"mov extension", "video.mov", false},
		{"mkv extension", "video.mkv", false},
		{"no extension", "video", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidMP4Extension(tt.filename)
			if got != tt.want {
				t.Errorf("IsValidMP4Extension(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestEnsureMP4Extension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"already has mp4", "video.mp4", "video.mp4"},
		{"already has m4v", "video.m4v", "video.m4v"},
		{"no extension", "video", "video.mp4"},
		{"other extension", "video.avi", "video.avi.mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsureMP4Extension(tt.filename)
			if got != tt.want {
				t.Errorf("EnsureMP4Extension(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
