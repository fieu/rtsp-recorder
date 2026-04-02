# Phase 2: Core Recording Engine - Context

**Gathered:** 2025-04-02
**Status:** Ready for planning

<domain>
## Phase Boundary

FFmpeg subprocess integration for actual RTSP recording, stop condition orchestration (Ctrl+C, duration, file size), real-time progress display, and MP4 finalization on shutdown. This phase makes the record command actually record streams to files.

Depends on: Phase 1 (Foundation & Configuration) — CLI structure, config loading, ffmpeg validation

</domain>

<decisions>
## Implementation Decisions

### FFmpeg Command Options
- **D-13:** Use TCP transport for RTSP (`-rtsp_transport tcp`) — more reliable than UDP
- **D-14:** Stream copy mode (`-c copy`) — no re-encoding, lowest CPU usage
- **D-15:** 5-second RTSP buffer (`-rtsp_transport tcp -buffer_size 65536` or similar)
- **D-16:** Output format MP4 with faststart (`-f mp4 -movflags +faststart`)

### Stop Condition Coordination
- **D-17:** Multiple stop conditions: first trigger wins (Ctrl+C OR duration limit OR file size limit)
- **D-18:** Graceful shutdown sequence: SIGINT → wait 10s → SIGTERM → wait 5s → SIGKILL (if needed)
- **D-19:** Use Go context.Context for coordinating cancellation across goroutines

### Progress Display Format
- **D-20:** Display all metrics: bytes recorded, elapsed time, file size, current bitrate
- **D-21:** Update every 1 second during active recording
- **D-22:** Single line format with carriage return (`\r`) to overwrite: `Recording: 1.2GB | 00:05:30 | 4.5Mbps`
- **D-23:** Final summary on new line when recording stops

### MP4 Finalization Strategy
- **D-24:** 10-second graceful shutdown timeout before escalating to SIGKILL
- **D-25:** Always save partial MP4 on interruption — partial file is better than lost data
- **D-26:** Ensure ffmpeg writes moov atom (MP4 metadata) by using `-movflags +faststart`

### File Size Monitoring
- **D-27:** Poll file size every 1 second using `os.Stat()`
- **D-28:** Stop recording when file size reaches configured max (in MB)
- **D-29:** File size check happens in parallel goroutine alongside duration timer

### the agent's Discretion
- Exact ffmpeg flag ordering and additional optimization flags
- Progress output formatting details (precision, units)
- Exact polling interval timing (can vary ±200ms)
- Error recovery on partial write failures

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Requirements
- `.planning/REQUIREMENTS.md` §Recording Core (REC-01 through REC-05) — Recording requirements
- `.planning/REQUIREMENTS.md` §Stop Conditions (STOP-01 through STOP-04) — Stop condition requirements
- `.planning/REQUIREMENTS.md` §Error Handling (ERR-03) — MP4 finalization requirement

### Prior Phase Context
- `.planning/phases/01-foundation-configuration/01-CONTEXT.md` — Phase 1 decisions (carry forward D-01 through D-12)

### Research Insights
- `.planning/research/PITFALLS.md` §Process management — Critical: zombie processes, signal handling
- `.planning/research/PITFALLS.md` §MP4 corruption — Critical: unclean shutdown prevention
- `.planning/research/PITFALLS.md` §Duration handling — Go context vs ffmpeg `-t`
- `.planning/research/ARCHITECTURE.md` — Process management patterns with `os/exec`

### Go Libraries
- `os/exec` — Process execution and management
- `context` — Cancellation propagation
- `os/signal` — Signal handling (SIGINT, SIGTERM)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `cmd/record.go` — Record command stub (needs actual recording implementation in RunE)
- `config/config.go` — Config struct with all settings (Duration, MaxFileSize, etc.)
- `internal/validator/ffmpeg.go` — FFmpeg path validation (reuse CheckFFmpeg())
- `internal/utils/file.go` — GenerateTimestampFilename() for output naming

### Established Patterns
- Error format: `[ERROR]`, `[WARNING]`, `[INFO]` tags (from Phase 1)
- Config loading: `config.Load()` returns Config struct with user settings
- Filename format: `YYYY-MM-DD-HH-MM-SS.mp4` via `utils.GenerateTimestampFilename()`

### Integration Points
- Record command connects to: ffmpeg subprocess via `os/exec.Command()`
- Config feeds into: ffmpeg command-line argument generation
- Signal handling: needs to hook into existing Cobra command lifecycle

</code_context>

<specifics>
## Specific Ideas

- Progress output should look professional: `Recording: 1.23 GB | 00:12:34 | 5.1 Mbps`
- On Ctrl+C, show: `[INFO] Stopping recording...` then wait for graceful shutdown
- If graceful shutdown fails, show: `[WARNING] Forcing stop, file may be incomplete`
- FFmpeg stderr parsing optional for Phase 2 (can defer to Phase 3 for progress accuracy)

</specifics>

<deferred>
## Deferred Ideas

- FFmpeg stderr parsing for accurate bitrate/progress — Phase 3 (Resilience & Feedback)
- RTSP stream accessibility pre-check (DESCRIBE request) — Phase 3
- Segmented recording (auto-split files) — v2 requirement

</deferred>

---

*Phase: 02-core-recording-engine*
*Context gathered: 2025-04-02*
