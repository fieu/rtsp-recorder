# Phase 8: Periodic Progress Logging - Context

**Gathered:** 2025-04-02
**Status:** Ready for planning

<domain>
## Phase Boundary

Replace the live progress bar (which uses \r to overwrite the same line) with periodic log messages that appear every X seconds. This provides better logging for systems that capture logs (like systemd, containers, or CI/CD) where the \r overwrite doesn't work well.

Depends on: Phase 2 (Core Recording Engine) — Progress display in recorder/recorder.go

</domain>

<decisions>
## Implementation Decisions

### Progress Display Change
- **D-96:** Remove live progress bar with \r carriage return
- **D-97:** Replace with periodic log messages every X seconds
- **D-98:** Log level for progress: "info" (visible by default)

### Configuration
- **D-99:** Config field: `progress_interval` (duration string like "10s", "30s", "1m")
- **D-100:** Default value: 10 seconds
- **D-101:** Minimum value: 5 seconds (prevent spam)
- **D-102:** Set to 0 to disable progress logging entirely

### Log Message Format
- **D-103:** Include: elapsed time, bytes recorded, file size, current bitrate
- **D-104:** Structured logging with zerolog (fields not string concatenation)
- **D-105:** Example: `logger.Info().Dur("elapsed", elapsed).Int64("bytes", bytes).Msg("Recording progress")`

### Backward Compatibility
- **D-106:** Old behavior removed entirely (no flag to keep it)
- **D-107:** Final summary still printed at end (that worked well)

### Implementation Details
- **D-108:** Use time.Ticker for periodic logging
- **D-109:** Stop ticker when recording stops
- **D-110:** Log immediately at start (0 seconds) and then every interval

### the agent's Discretion
- Exact log field names
- Whether to include human-readable formatting alongside structured fields
- Log message text content

</decisions>

<canonical_refs>
## Canonical References

### Project Requirements
- `.planning/ROADMAP.md` §Phase 8 — Periodic Progress Logging requirements

### Prior Phase Context
- `.planning/phases/02-core-recording-engine/02-CONTEXT.md` — Recorder implementation

### Code to Modify
- `recorder/recorder.go` — displayProgress() method
- `config/config.go` — add ProgressInterval field

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `recorder/recorder.go` — displayProgress() with \r logic
- `config/config.go` — Config struct for new field
- `cmd/root.go` — Flag binding if needed

### Current Progress Implementation
- Uses `fmt.Printf("\rRecording: %s | %s | %s", ...)`
- Overwrites same line every second
- Located in recorder.go displayProgress() method

### Integration Points
- Replace ticker in displayProgress()
- Use zerolog for output instead of fmt.Printf
- Keep stats calculation logic

</code_context>

<specifics>
## Specific Ideas

- Current: `\rRecording: 1.2MB | 00:05:30 | 650 Kbps`
- New: `{"level":"info","elapsed":"5m30s","bytes":1258291,"bitrate":650000,"message":"Recording progress"}`
- Or human-friendly console: `INF Recording progress elapsed=5m30s bytes=1.2MB bitrate=650Kbps`

</specifics>

<deferred>
## Deferred Ideas

- Real-time web dashboard — out of scope
- Progress bar as optional flag — simpler to just remove it

</deferred>

---

*Phase: 08-periodic-progress-logging*
*Context gathered: 2025-04-02*
