/*
Copyright © 2026 rtsp-recorder contributors

RTSP stream validation utilities.
Provides pre-flight checks for RTSP stream accessibility using DESCRIBE request.

This package follows D-35 through D-38 from Phase 3 context:
- DESCRIBE request with 10-second timeout
- "Accessible" criteria: 200 OK with valid SDP response
- Fail fast with clear error messages
*/
package validator

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ValidateRTSP performs RTSP DESCRIBE request validation on the given URL.
// It sends a DESCRIBE request with the specified timeout and checks for
// 200 OK response with valid SDP content.
//
// Returns an error with [ERROR] prefix and actionable message if validation fails.
//
// Per D-35 through D-38:
//   - DESCRIBE request with configurable timeout (D-36: default 10s)
//   - "Accessible" criteria: 200 OK with valid SDP response (D-37)
//   - Fail fast with descriptive error if DESCRIBE fails (D-38)
func ValidateRTSP(rtspURL string, timeout time.Duration) error {
	// Validate URL format and extract host:port
	host, port, path, err := parseRTSPURL(rtspURL)
	if err != nil {
		return fmt.Errorf("[ERROR] Invalid RTSP URL: %w", err)
	}

	// Build address string
	address := net.JoinHostPort(host, port)

	// Establish TCP connection with timeout
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("[ERROR] Cannot connect to RTSP server. Check IP address and port.")
		}
		if strings.Contains(err.Error(), "no route to host") || strings.Contains(err.Error(), "network is unreachable") {
			return fmt.Errorf("[ERROR] Network unreachable. Check network connectivity.")
		}
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "i/o timeout") {
			return fmt.Errorf("[ERROR] Connection timeout. Camera may be offline or behind firewall.")
		}
		return fmt.Errorf("[ERROR] Cannot connect to RTSP server. Check IP address and port.")
	}
	defer conn.Close()

	// Set read/write timeouts on connection
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("[ERROR] Failed to set connection timeout: %w", err)
	}

	// Build and send DESCRIBE request
	describeReq := buildDESCRIBERequest(rtspURL, host)
	if _, err := conn.Write([]byte(describeReq)); err != nil {
		return fmt.Errorf("[ERROR] Failed to send DESCRIBE request. Camera may be offline.")
	}

	// Read and parse response
	reader := bufio.NewReader(conn)
	response, err := readResponse(reader)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to read RTSP response. Camera may be incompatible.")
	}

	// Parse status line
	statusCode, statusText, err := parseStatusLine(response)
	if err != nil {
		return fmt.Errorf("[ERROR] Invalid RTSP response from server. Camera may be incompatible.")
	}

	// Handle different response codes
	switch statusCode {
	case 200:
		// Success - check for SDP content
		if !hasSDPContent(response) {
			return fmt.Errorf("[ERROR] Stream found but no valid SDP data. Camera may be misconfigured.")
		}
		return nil
	case 401, 403:
		return fmt.Errorf("[ERROR] Authentication required. Check username/password in URL.")
	case 404:
		return fmt.Errorf("[ERROR] Stream not found. Verify the RTSP URL path (got %s).", path)
	case 400:
		return fmt.Errorf("[ERROR] Bad request. Verify the RTSP URL format.")
	case 503:
		return fmt.Errorf("[ERROR] Service unavailable. Camera may be overloaded or in maintenance mode.")
	default:
		return fmt.Errorf("[ERROR] RTSP server returned %d %s. Check camera status.", statusCode, statusText)
	}
}

// IsStreamAccessible is a quick check wrapper around ValidateRTSP.
// Returns true if the stream is accessible, false otherwise.
func IsStreamAccessible(rtspURL string) bool {
	err := ValidateRTSP(rtspURL, 10*time.Second)
	return err == nil
}

// parseRTSPURL extracts host, port, and path from an RTSP URL.
// Returns error if URL is invalid or missing scheme.
func parseRTSPURL(rtspURL string) (host, port, path string, err error) {
	// Handle empty URL
	if rtspURL == "" {
		return "", "", "", fmt.Errorf("URL must use rtsp:// scheme")
	}

	// Check for scheme
	if !strings.HasPrefix(rtspURL, "rtsp://") {
		// Try to add scheme if missing
		if !strings.Contains(rtspURL, "://") {
			rtspURL = "rtsp://" + rtspURL
		} else {
			return "", "", "", fmt.Errorf("URL must use rtsp:// scheme")
		}
	}

	u, err := url.Parse(rtspURL)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Extract host
	host = u.Hostname()
	if host == "" {
		return "", "", "", fmt.Errorf("URL missing host")
	}

	// Extract port (default to 554 if not specified)
	port = u.Port()
	if port == "" {
		port = "554"
	}

	// Extract path (default to "/" if empty)
	path = u.Path
	if path == "" {
		path = "/"
	}

	return host, port, path, nil
}

// buildDESCRIBERequest constructs an RTSP DESCRIBE request string.
func buildDESCRIBERequest(rtspURL, host string) string {
	// Generate a simple CSeq number (can use time-based or random)
	// For simplicity, we'll use a fixed CSeq of 1
	return fmt.Sprintf(
		"DESCRIBE %s RTSP/1.0\r\n"+
			"CSeq: 1\r\n"+
			"Host: %s\r\n"+
			"Accept: application/sdp\r\n"+
			"User-Agent: rtsp-recorder/1.0\r\n"+
			"\r\n",
		rtspURL, host,
	)
}

// readResponse reads the RTSP response from the connection.
// It reads until headers end (double CRLF) or buffer limit.
func readResponse(reader *bufio.Reader) (string, error) {
	var response strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// If we've read some content, return what we have
			if response.Len() > 0 {
				return response.String(), nil
			}
			return "", err
		}
		response.WriteString(line)

		// Check for end of headers (empty line)
		if line == "\r\n" || line == "\n" {
			// Headers end here, we've got what we need
			return response.String(), nil
		}
	}
}

// parseStatusLine extracts the status code and text from the RTSP response.
// Expected format: "RTSP/1.0 200 OK"
func parseStatusLine(response string) (int, string, error) {
	lines := strings.Split(response, "\n")
	if len(lines) == 0 {
		return 0, "", fmt.Errorf("empty response")
	}

	// Parse first line
	firstLine := strings.TrimSpace(lines[0])

	// RTSP status line pattern: "RTSP/1.0 200 OK"
	re := regexp.MustCompile(`^RTSP/\d+\.\d+\s+(\d+)\s+(.+)$`)
	matches := re.FindStringSubmatch(firstLine)

	if len(matches) < 3 {
		return 0, "", fmt.Errorf("invalid status line: %s", firstLine)
	}

	var statusCode int
	if _, err := fmt.Sscanf(matches[1], "%d", &statusCode); err != nil {
		return 0, "", fmt.Errorf("failed to parse status code: %w", err)
	}
	statusText := matches[2]

	return statusCode, statusText, nil
}

// hasSDPContent checks if the response contains SDP content.
// It looks for Content-Type: application/sdp header.
func hasSDPContent(response string) bool {
	// Check for Content-Type header indicating SDP
	lowerResponse := strings.ToLower(response)
	return strings.Contains(lowerResponse, "content-type: application/sdp") ||
		strings.Contains(lowerResponse, "content-type:application/sdp")
}
