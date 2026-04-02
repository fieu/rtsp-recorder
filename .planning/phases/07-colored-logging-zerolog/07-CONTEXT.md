# Phase 7: Colored Logging with Zerolog - Context

**Gathered:** 2025-04-02
**Status:** Ready for planning

<domain>
## Phase Boundary

Replace Uber's zap library with zerolog for better terminal experience with colored output when running in a TTY. Zerolog provides better performance and prettier console output with automatic color detection.

Depends on: Phase 6 (Structured Logging with Zap) — Existing logging infrastructure

</domain>

<decisions>
## Implementation Decisions

### Zerolog Integration
- **D-78:** Replace zap with rs/zerolog library
- **D-79:** Use zerolog.ConsoleWriter for TTY output with colors
- **D-80:** Use zerolog.JSON output for non-TTY (pipes, files)
- **D-81:** Auto-detect TTY using go.isatty or similar
- **D-82:** Log levels: debug, info, warn, error, fatal, panic (zerolog standard)

### Colored Output
- **D-83:** Color scheme: Debug (gray), Info (green), Warn (yellow), Error (red), Fatal (red+bold)
- **D-84:** Include timestamp, level, message, and fields in colored output
- **D-85:** Pretty print fields in console mode

### Configuration
- **D-86:** Keep existing log_level config field (no breaking changes)
- **D-87:** Keep --log-level flag (no breaking changes)
- **D-88:** Add optional --no-color flag to disable colors even in TTY
- **D-89:** Environment variable NO_COLOR respected (standard convention)

### API Compatibility
- **D-90:** Maintain similar API to minimize code changes
- **D-91:** Global logger accessible as logger.Logger (same as before)
- **D-92:** Structured fields using zap-like API (String, Int, Duration, etc.)

### Migration Strategy
- **D-93:** Replace logger/logger.go implementation
- **D-94:** Update all log calls to use zerolog API (similar structure)
- **D-95:** Remove zap dependency from go.mod

### the agent's Discretion
- Exact color palette for each log level
- Field formatting in console mode (key=value vs JSON)
- Timestamp format (RFC3339 vs human-readable)
- Whether to add caller info (file:line) in development mode

</decisions>

<canonical_refs>
## Canonical References

### Project Requirements
- `.planning/ROADMAP.md` §Phase 7 — Colored Logging requirements

### Prior Phase Context
- `.planning/phases/06-structured-logging-zap/06-CONTEXT.md` — Zap logging decisions

### Zerolog Documentation
- `github.com/rs/zerolog` — High-performance structured logging
- ConsoleWriter for human-readable output
- Automatic TTY detection for color output

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `logger/logger.go` — Replace zap implementation with zerolog
- `config/config.go` — LogLevel field stays the same
- `cmd/root.go` — Logger initialization stays in same location
- All files with logging — Update import and API calls

### Current Logging Patterns
- `logger.Logger.Info(...)` — Replace with zerolog equivalent
- Field chaining: `zap.String(), zap.Int()` — Replace with `Str(), Int()`
- ~50+ log calls across codebase

### Integration Points
- Logger initialized in cmd/root.go initConfig()
- Global logger variable: cmd.Logger
- Passed to recorder via constructor

</code_context>

<specifics>
## Specific Ideas

- Console output example:
  ```
  18:16:09 INF Starting rtsp-recorder
  18:16:09 INF FFmpeg found path=/opt/homebrew/bin/ffmpeg version=8.1
  ```
- Error output in red:
  ```
  18:16:12 ERR Recording failed error="connection refused"
  ```

</specifics>

<deferred>
## Deferred Ideas

- Log sampling (for high-throughput) — not needed for CLI tool
- Hook system for custom outputs — defer if needed
- Log rotation — not applicable for CLI

</deferred>

---

*Phase: 07-colored-logging-zerolog*
*Context gathered: 2025-04-02*
