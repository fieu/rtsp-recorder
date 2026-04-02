---
phase: 06-structured-logging-zap
plan: 02
subsystem: logging
tags: [zap, structured-logging, migration, fmt-to-zap]

# Dependency graph
requires:
  - phase: 06-structured-logging-zap
    plan: 01
    provides: "Zap logger foundation with global Logger variable"
provides:
  - "All [INFO] logging replaced with zap structured logging"
  - "Retry warnings use Logger.Warn() with structured fields"
  - "Recorder accepts *zap.Logger for internal logging"
affects:
  - "cmd/record.go - uses Logger.Info for all status messages"
  - "cmd/validate.go - uses Logger.Info for validation messages"
  - "recorder/recorder.go - uses logger.Info/Debug/Warn"
  - "internal/retry/retry.go - uses Logger.Warn for retry messages"

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Package-level logger field in structs (recorder.Logger)"
    - "Global cmd.Logger for command-level logging"
    - "Logger injection via constructor (recorder.New(cfg, logger))"
    - "Structured fields with zap.String, zap.Int, zap.Duration"

key-files:
  created: []
  modified:
    - "cmd/record.go - Replaced 8 fmt.Println calls with Logger.Info()"
    - "cmd/validate.go - Replaced 6 fmt.Println calls with Logger.Info()"
    - "recorder/recorder.go - Added logger field, replaced fmt.Printf with logger.Info/Warn/Debug"
    - "internal/retry/retry.go - Added Logger to RetryConfig, replaced fmt.Printf with logger.Warn"
    - "cmd/root.go - Already had Logger.Info for config file (from 06-01)"

key-decisions:
  - "Progress display (fmt.Printf with \\r) stays on stdout per D-76"
  - "Retry messages use Warn level per D-71 (recoverable issues)"
  - "Recorder struct accepts logger via constructor for dependency injection"
  - "Tests pass nil logger for simplicity"

patterns-established:
  - "Logger injection: New(cfg, logger) pattern for packages needing logging"
  - "Structured logging: Use zap fields for all configuration values"
  - "Log level appropriateness: Info for status, Warn for retries, Debug for details"

requirements-completed: [LOG-02]

# Metrics
duration: 20min
completed: 2026-04-02
---

# Phase 6 Plan 02: Replace fmt.Println with Zap Logging Summary

**Complete migration from fmt.Println/fmt.Printf logging to structured zap logging across the entire codebase**

## Performance

- **Duration:** 20 min
- **Started:** 2026-04-02
- **Completed:** 2026-04-02
- **Tasks:** 5
- **Files modified:** 6 (4 code files + 2 test files)

## Accomplishments

- Replaced all `fmt.Println [INFO]` calls with `Logger.Info()` using structured fields
- Added zap import to cmd/record.go, cmd/validate.go, recorder/recorder.go, internal/retry/retry.go
- Added `logger *zap.Logger` field to Recorder struct with constructor injection
- Updated `recorder.New()` to accept `*zap.Logger` parameter
- Updated `retry.DefaultRetryConfig()` to accept `*zap.Logger` parameter
- Replaced retry `fmt.Printf [INFO]` with `logger.Warn()` (per D-71 - retries are warnings)
- Added `Logger` field to `RetryConfig` struct for callback access
- Replaced recorder `fmt.Printf [INFO]` and `[WARNING]` with `logger.Info()` and `logger.Warn()`
- Added `logger.Debug()` for ffmpeg configuration details (per D-69)
- Preserved progress display on stdout using `fmt.Printf` with `\r` (per D-76)
- Updated all test files to pass `nil` logger for simplicity

## Task Commits

Each task was committed atomically:

1. **Task 1: Replace logging in cmd/record.go** - `b41466e` (feat)
2. **Task 2: Replace logging in cmd/validate.go** - `5339614` (feat)
3. **Task 3: Replace logging in recorder/recorder.go** - `fd00eaa` (feat)
4. **Task 4: Replace logging in internal/retry/retry.go** - `ff8aad8` (feat)
5. **Task 5: Update test files for logger parameter** - `1234b4c` (test)

## Files Created/Modified

### Code Files Modified
- `cmd/record.go` - Replaced 8 fmt.Println calls with Logger.Info() and structured fields
- `cmd/validate.go` - Replaced 6 fmt.Println calls with Logger.Info() and structured fields
- `recorder/recorder.go` - Added logger field, replaced fmt.Printf with logger.Info/Warn/Debug
- `internal/retry/retry.go` - Added Logger to RetryConfig, replaced fmt.Printf with closure-based logger.Warn
- `cmd/root.go` - Already complete from 06-01 (Logger.Info for config file)

### Test Files Modified
- `cmd/record_test.go` - Updated DefaultRetryConfig() calls to pass nil logger
- `recorder/recorder_test.go` - Updated recorder.New() calls to pass nil logger
- `internal/retry/retry_test.go` - Updated DefaultRetryConfig() calls to pass nil logger

## Decisions Made

- Progress display (fmt.Printf with `\r`) stays on stdout per D-76 - NOT logged
- Retry messages use Warn level per D-71 (recoverable issues should be warnings)
- Recorder struct uses logger injection pattern via constructor for flexibility
- Test files pass nil logger - production code passes actual Logger
- Error messages before logger initialization (in root.go) stay as fmt.Fprintf to stderr

## Deviations from Plan

### None - plan executed exactly as written.

All tasks completed as specified in the plan:
- ✓ Task 1: cmd/record.go - All fmt.Println replaced with zap logging
- ✓ Task 2: cmd/validate.go - All fmt.Println replaced with zap logging
- ✓ Task 3: recorder/recorder.go - Added logger field and replaced fmt.Printf
- ✓ Task 4: internal/retry/retry.go - Retry messages use Logger.Warn with structured fields
- ✓ Task 5: root.go config file logging - Already complete from 06-01

## Issues Encountered

1. **Test file compilation errors** - After changing function signatures (recorder.New and DefaultRetryConfig), test files needed updates to pass the new logger parameter. Fixed by passing nil for tests.

2. **sed command over-replacement** - Initial sed command replaced both `recorder.New(cfg)` and `ffmpeg.New(cfg)` with the nil parameter. Fixed by reverting the ffmpeg.New calls.

## User Setup Required

None - no external service configuration required.

## Verification Results

All success criteria verified:
- [x] `go build` compiles successfully
- [x] `go test ./...` passes all tests
- [x] No fmt.Println calls with [INFO] prefix remain in modified files
- [x] All zap imports are present in modified files
- [x] Progress display still uses fmt.Printf with \r (verified 3 occurrences in recorder.go)
- [x] Error messages remain visible
- [x] Log levels appropriate:
  - Info: Starting rtsp-recorder, Using URL, FFmpeg found, Recording configuration, Timelapse enabled, Starting recording, Press Ctrl+C, Output file, Stopping recording
  - Warn: Retry messages (per D-71), FFmpeg stop errors, Could not read output file
  - Debug: FFmpeg command created details
  - Error: Config file errors (before logger init, stays as fmt.Fprintf)

## Self-Check: PASSED

- [x] cmd/record.go uses Logger.Info with structured fields (zap.String, zap.Duration, zap.Int64, zap.Int)
- [x] cmd/validate.go uses Logger.Info with structured fields
- [x] recorder/recorder.go has logger field and uses logger.Info/Warn/Debug
- [x] internal/retry/retry.go has Logger field and uses logger.Warn for retries
- [x] cmd/root.go already has Logger.Info for config file
- [x] All commits exist (b41466e, 5339614, fd00eaa, ff8aad8, 1234b4c)
- [x] Build succeeds
- [x] All tests pass

## Known Stubs

None. All logging is fully wired with actual logger instances.

## Next Phase Readiness

- All fmt.Println [INFO] logging replaced with zap structured logging
- Logger injection pattern established for future packages
- Log levels consistently applied: Info (status), Warn (retries/recoverable), Debug (details), Error (fatal)
- Ready for Phase 7 and beyond - all infrastructure in place

---
*Phase: 06-structured-logging-zap*
*Completed: 2026-04-02*
