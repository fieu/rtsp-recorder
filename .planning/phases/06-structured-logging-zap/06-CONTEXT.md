# Phase 6: Structured Logging with Zap - Context

**Gathered:** 2025-04-02
**Status:** Ready for planning

<domain>
## Phase Boundary

Integrate Uber's zap library for structured logging with configurable log levels. Replace existing fmt.Println logging with proper structured logging throughout the codebase. Log level can be configured via YAML config, environment variable, or CLI flag.

Depends on: Phase 1 (Foundation & Configuration) — Config system, CLI flags

</domain>

<decisions>
## Implementation Decisions

### Zap Integration
- **D-61:** Use Uber's zap library for structured logging
- **D-62:** Log format: structured JSON for production, human-readable console for development
- **D-63:** Log levels: debug, info, warn, error (standard zap levels)
- **D-64:** Default log level: "info" for production use

### Configuration
- **D-65:** Config field: `log_level` in YAML (string: "debug", "info", "warn", "error")
- **D-66:** Environment variable: `RTSP_RECORDER_LOG_LEVEL`
- **D-67:** CLI flag: `--log-level` (no shorthand to avoid confusion with timelapse `-l`)
- **D-68:** Config precedence: flag > env > config > default (standard viper precedence)

### Log Level Usage
- **D-69:** Debug: detailed ffmpeg args, frame processing info, internal state
- **D-70:** Info: user-facing status messages (recording start/stop, progress summary)
- **D-71:** Warn: recoverable issues (retries, partial failures, deprecated features)
- **D-72:** Error: fatal errors (connection failures, invalid config, ffmpeg crashes)

### Logger Initialization
- **D-73:** Initialize logger early in main.go before any logging
- **D-74:** Store logger in context or global variable accessible throughout app
- **D-75:** Replace all fmt.Println/fmt.Printf calls with appropriate zap level

### Backward Compatibility
- **D-76:** Progress display stays on stdout (not logged) for real-time user feedback
- **D-77:** Error messages to stderr remain visible even at "error" log level

### the agent's Discretion
- Exact zap configuration (development vs production mode)
- Logger field naming conventions (camelCase vs snake_case)
- Whether to add additional log fields (version, build info, etc.)
- Log file output option (in addition to console)

</decisions>

<canonical_refs>
## Canonical References

### Project Requirements
- `.planning/ROADMAP.md` §Phase 6 — Structured Logging requirements

### Prior Phase Context
- `.planning/phases/01-foundation-configuration/01-CONTEXT.md` — Config system decisions
- `.planning/REQUIREMENTS.md` §Configuration — Config precedence patterns

### Uber Zap Documentation
- `go.uber.org/zap` — Structured logging library
- Zap supports both development (human-readable) and production (JSON) modes

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `config/config.go` — Config struct (add LogLevel field)
- `config.BindFlags()` — Add log-level flag binding
- `cmd/root.go` — Root command (initialize logger here)
- `cmd/record.go` — Record command (replace fmt.Println with zap)
- `recorder/recorder.go` — Progress display (keep on stdout)

### Current Logging Patterns
- fmt.Println("[INFO] ...") — 20+ occurrences across codebase
- fmt.Printf for progress display (should stay on stdout)
- Error messages via fmt.Errorf wrapped with [ERROR] prefix

### Integration Points
- Logger initialized in main.go or cmd/root.go init()
- Passed to recorder, ffmpeg, validator packages
- Config loading happens before logger setup (chicken-egg problem)

</code_context>

<specifics>
## Specific Ideas

- Example: zap.L().Info("Starting recording", zap.String("url", cfg.URL), zap.Duration("duration", cfg.Duration))
- Example: zap.L().Error("Recording failed", zap.Error(err), zap.Int("attempt", attempt))
- Progress display stays unchanged: fmt.Printf("\rRecording: %s | %s | %s", ...)

</specifics>

<deferred>
## Deferred Ideas

- Log rotation and file output — defer to later phase
- Remote log aggregation (e.g., to cloud) — out of scope for v1.x
- Request tracing/correlation IDs — not needed for CLI tool

</deferred>

---

*Phase: 06-structured-logging-zap*
*Context gathered: 2025-04-02*
