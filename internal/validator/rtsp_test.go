/*
Copyright © 2026 rtsp-recorder contributors

Tests for RTSP validator package.
*/
package validator

import (
	"testing"
	"time"
)

// TestParseRTSPURL tests the URL parsing functionality
func TestParseRTSPURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantHost    string
		wantPort    string
		wantPath    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid URL with port",
			url:      "rtsp://192.168.1.100:554/stream",
			wantHost: "192.168.1.100",
			wantPort: "554",
			wantPath: "/stream",
			wantErr:  false,
		},
		{
			name:     "valid URL without port",
			url:      "rtsp://192.168.1.100/stream",
			wantHost: "192.168.1.100",
			wantPort: "554", // default
			wantPath: "/stream",
			wantErr:  false,
		},
		{
			name:     "valid URL with credentials",
			url:      "rtsp://admin:password@camera.local:554/live/ch00_0",
			wantHost: "camera.local",
			wantPort: "554",
			wantPath: "/live/ch00_0",
			wantErr:  false,
		},
		{
			name:     "URL missing scheme - auto-add",
			url:      "192.168.1.100:554/stream",
			wantHost: "192.168.1.100",
			wantPort: "554",
			wantPath: "/stream",
			wantErr:  false,
		},
		{
			name:        "empty URL",
			url:         "",
			wantErr:     true,
			errContains: "URL must use rtsp:// scheme",
		},
		{
			name:        "invalid scheme",
			url:         "http://192.168.1.100/stream",
			wantErr:     true,
			errContains: "URL must use rtsp:// scheme",
		},
		{
			name:        "URL missing host",
			url:         "rtsp:///stream",
			wantErr:     true,
			errContains: "URL missing host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, path, err := parseRTSPURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseRTSPURL() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !containsStr(err.Error(), tt.errContains) {
					t.Errorf("parseRTSPURL() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("parseRTSPURL() unexpected error = %v", err)
				return
			}

			if host != tt.wantHost {
				t.Errorf("parseRTSPURL() host = %v, want %v", host, tt.wantHost)
			}
			if port != tt.wantPort {
				t.Errorf("parseRTSPURL() port = %v, want %v", port, tt.wantPort)
			}
			if path != tt.wantPath {
				t.Errorf("parseRTSPURL() path = %v, want %v", path, tt.wantPath)
			}
		})
	}
}

// TestParseStatusLine tests the status line parsing
func TestParseStatusLine(t *testing.T) {
	tests := []struct {
		name           string
		response       string
		wantCode       int
		wantStatusText string
		wantErr        bool
	}{
		{
			name:           "200 OK",
			response:       "RTSP/1.0 200 OK\r\n",
			wantCode:       200,
			wantStatusText: "OK",
			wantErr:        false,
		},
		{
			name:           "404 Not Found",
			response:       "RTSP/1.0 404 Not Found\r\n",
			wantCode:       404,
			wantStatusText: "Not Found",
			wantErr:        false,
		},
		{
			name:           "401 Unauthorized",
			response:       "RTSP/1.0 401 Unauthorized\r\n",
			wantCode:       401,
			wantStatusText: "Unauthorized",
			wantErr:        false,
		},
		{
			name:     "invalid status line",
			response: "Not a valid status line\r\n",
			wantErr:  true,
		},
		{
			name:     "empty response",
			response: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, statusText, err := parseStatusLine(tt.response)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseStatusLine() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("parseStatusLine() unexpected error = %v", err)
				return
			}

			if code != tt.wantCode {
				t.Errorf("parseStatusLine() code = %v, want %v", code, tt.wantCode)
			}
			if statusText != tt.wantStatusText {
				t.Errorf("parseStatusLine() statusText = %v, want %v", statusText, tt.wantStatusText)
			}
		})
	}
}

// TestHasSDPContent tests the SDP content detection
func TestHasSDPContent(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     bool
	}{
		{
			name:     "has SDP content type",
			response: "Content-Type: application/sdp\r\n",
			want:     true,
		},
		{
			name:     "has lowercase SDP content type",
			response: "content-type: application/sdp\r\n",
			want:     true,
		},
		{
			name:     "no space in header",
			response: "Content-Type:application/sdp\r\n",
			want:     true,
		},
		{
			name:     "wrong content type",
			response: "Content-Type: text/html\r\n",
			want:     false,
		},
		{
			name:     "no content type header",
			response: "Server: test\r\n",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasSDPContent(tt.response)
			if got != tt.want {
				t.Errorf("hasSDPContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBuildDESCRIBERequest tests the request building
func TestBuildDESCRIBERequest(t *testing.T) {
	url := "rtsp://192.168.1.100:554/stream"
	host := "192.168.1.100"

	req := buildDESCRIBERequest(url, host)

	// Check that the request contains the expected components
	if !containsStr(req, "DESCRIBE") {
		t.Error("DESCRIBE request missing 'DESCRIBE' method")
	}
	if !containsStr(req, url) {
		t.Errorf("DESCRIBE request missing URL %s", url)
	}
	if !containsStr(req, "RTSP/1.0") {
		t.Error("DESCRIBE request missing 'RTSP/1.0' version")
	}
	if !containsStr(req, "CSeq: 1") {
		t.Error("DESCRIBE request missing 'CSeq: 1'")
	}
	if !containsStr(req, "Accept: application/sdp") {
		t.Error("DESCRIBE request missing 'Accept: application/sdp'")
	}
	if !containsStr(req, "User-Agent: rtsp-recorder") {
		t.Error("DESCRIBE request missing User-Agent header")
	}
}

// TestValidateRTSPInvalidURL tests validation with invalid URLs
func TestValidateRTSPInvalidURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		errContains string
	}{
		{
			name:        "empty URL",
			url:         "",
			errContains: "Invalid RTSP URL",
		},
		{
			name:        "wrong scheme",
			url:         "http://192.168.1.100/stream",
			errContains: "Invalid RTSP URL",
		},
		{
			name:        "missing host",
			url:         "rtsp:///stream",
			errContains: "Invalid RTSP URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRTSP(tt.url, 2*time.Second)
			if err == nil {
				t.Errorf("ValidateRTSP() error = nil, want error containing %q", tt.errContains)
				return
			}
			if !containsStr(err.Error(), tt.errContains) {
				t.Errorf("ValidateRTSP() error = %v, want containing %q", err, tt.errContains)
			}
			if !containsStr(err.Error(), "[ERROR]") {
				t.Errorf("ValidateRTSP() error should have [ERROR] prefix: %v", err)
			}
		})
	}
}

// TestIsStreamAccessible tests the quick check wrapper
func TestIsStreamAccessible(t *testing.T) {
	// Test with invalid URL should return false
	if IsStreamAccessible("invalid-url") {
		t.Error("IsStreamAccessible() with invalid URL should return false")
	}

	// Test with empty URL should return false
	if IsStreamAccessible("") {
		t.Error("IsStreamAccessible() with empty URL should return false")
	}
}

// containsStr checks if a string contains a substring (case-insensitive)
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) &&
		(stringContainsFold(s, substr) ||
			stringContainsFold(s, substr))
}

// stringContainsFold case-insensitive substring check
func stringContainsFold(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	// Simple case-insensitive comparison
	lowerS := toLower(s)
	lowerSubstr := toLower(substr)
	return contains(lowerS, lowerSubstr)
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
