# Phase 3: Resilience & Feedback - Context

**Gathered:** 2025-04-02
**Status:** Ready for planning
**Mode:** Agent discretion (user deferred all areas)

<domain>
## Phase Boundary

Retry logic on network failures, RTSP stream accessibility validation via DESCRIBE request, and descriptive error messages for common failure modes. This phase enhances reliability and user feedback without changing core recording behavior.

Depends on: Phase 2 (Core Recording Engine) — FFmpeg wrapper, stop conditions, recording orchestration

</domain>

<decisions>
## Implementation Decisions

### Retry Logic
- **D-30:** Retry triggered by: connection refused, connection timeout, broken pipe, network unreachable
- **D-31:** Retry attempts: use config `retry-attempts` (default: 3, per D-05)
- **D-32:** Retry delay: fixed 5-second delay between attempts (simple, predictable)
- **D-33:** No retry for: authentication failures (401/403), invalid URLs, permanent errors
- **D-34:** Retry happens at recorder level — re-attempts full Record() call with fresh ffmpeg process

### RTSP Pre-Validation
- **D-35:** Use RTSP DESCRIBE request before starting ffmpeg recording
- **D-36:** Timeout: 10 seconds for DESCRIBE response
- **D-37:** "Accessible" criteria: 200 OK response with valid SDP (Session Description Protocol)
- **D-38:** Early failure: fail fast with clear error if DESCRIBE fails (don't start ffmpeg)

### Error Specificity
- **D-39:** Parse ffmpeg stderr for common error patterns
- **D-40:** Map patterns to actionable messages:
  - `Connection refused` → "[ERROR] Cannot connect to camera. Check IP address and port."
  - `404 Not Found` → "[ERROR] Stream path not found. Verify the RTSP URL path."
  - `Invalid data` → "[ERROR] Stream data invalid. Camera may be offline or incompatible."
  - `No route to host` → "[ERROR] Network unreachable. Check network connectivity."
  - `401 Unauthorized` / `403 Forbidden` → "[ERROR] Authentication required. Check username/password in URL."
  - `Operation timed out` → "[ERROR] Connection timeout. Camera may be offline or behind firewall."

### Failure Classification
- **D-41:** Error categories:
  - `NetworkError` — Connection, timeout, route issues (retryable per D-30)
  - `AuthenticationError` — 401, 403 (not retryable)
  - `StreamError` — Invalid data, codec issues (not retryable)
  - `ConfigurationError` — Invalid URL, missing required fields (not retryable)
  - `FFmpegError` — Internal ffmpeg failures (not retryable)

### FFmpeg Stderr Parsing
- **D-42:** Parse ffmpeg stderr for error detection and bitrate estimation
- **D-43:** Update progress display with actual bitrate from ffmpeg output (improves accuracy)
- **D-44:** Buffer stderr lines, scan for error patterns while recording

### the agent's Discretion
- Exact regex patterns for error detection
- Additional error patterns beyond the 6 listed in D-40
- Exact retry delay timing (can vary ±1 second)
- DESCRIBE request implementation details (use net.Dial or external library)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Requirements
- `.planning/REQUIREMENTS.md` §Recording Core (REC-06) — Retry attempts requirement
- `.planning/REQUIREMENTS.md` §Error Handling (ERR-02) — Invalid URL error messages
- `.planning/REQUIREMENTS.md` §Error Handling (ERR-04) — Meaningful ffmpeg error messages

### Prior Phase Context
- `.planning/phases/01-foundation-configuration/01-CONTEXT.md` — Phase 1 decisions (D-01 through D-12)
- `.planning/phases/02-core-recording-engine/02-CONTEXT.md` — Phase 2 decisions (D-13 through D-29)

### Research Insights
- `.planning/research/PITFALLS.md` §RTSP validation — DESCRIBE request patterns
- `.planning/research/PITFALLS.md` §Error handling — FFmpeg stderr parsing
- `.planning/research/PITFALLS.md` §Retry logic — Network error detection

### Go Libraries
- `net` — RTSP DESCRIBE request (net.Dial, net.Conn)
- `bufio` — Stderr line buffering and scanning
- `regexp` — Error pattern matching

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `ffmpeg/ffmpeg.go` — FFmpeg wrapper (add stderr parsing hooks)
- `recorder/recorder.go` — Recording orchestrator (add retry loop)
- `recorder/stop_conditions.go` — Monitor pattern (reuse for error detection)
- `config/config.go` — Config struct has `RetryAttempts` field
- `internal/validator/ffmpeg.go` — Validation pattern

### Established Patterns
- Error format: `[ERROR]`, `[WARNING]`, `[INFO]` tags (from Phase 1)
- Progress display: `Recording: X | Y | Z` (from Phase 2)
- Config loading: `config.Load()` returns Config struct
- Signal handling: `signal.NotifyContext` pattern (from Phase 2)

### Integration Points
- RTSP validation happens before `recorder.Record()` call in `cmd/record.go`
- Retry loop wraps `recorder.Record()` in `cmd/record.go`
- Stderr parsing integrates with `ffmpeg/ffmpeg.go` process management
- Error classification feeds into existing error display in `recorder/recorder.go`

</code_context>

<specifics>
## Specific Ideas

- Error messages should guide user to fix the problem, not just state what happened
- Retry attempts should be logged: `[INFO] Retry 1/3 after 5s...`
- DESCRIBE validation should show: `[INFO] Validating RTSP stream...`
- Keep error classification simple — don't over-engineer category system

</specifics>

<deferred>
## Deferred Ideas

None — Phase 3 scope is v1 final phase

</deferred>

---

*Phase: 03-resilience-feedback*
*Context gathered: 2025-04-02*
