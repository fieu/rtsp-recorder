---
phase: 01-foundation-configuration
plan: 01
subsystem: cli

tags: [cobra, viper, yaml, ffmpeg, go, config]

# Dependency graph
requires: []
provides:
  - CLI framework with Cobra
  - Viper configuration system (YAML + env vars)
  - Config struct with mapstructure tags
  - FFmpeg validation utilities
  - validate subcommand
affects: [01-foundation-configuration]

# Tech tracking
tech-stack:
  added: [cobra v1.10.2, viper v1.21.0]
  patterns:
    - "Config package isolated from CLI (receives struct, not Viper directly)"
    - "Structured error messages with [ERROR]/[INFO] prefixes"
    - "Explicit config file path to avoid binary name conflict"

key-files:
  created:
    - cmd/root.go - Root command with Viper initialization
    - cmd/validate.go - Validate subcommand
    - config/config.go - Config struct and loading
    - internal/validator/ffmpeg.go - FFmpeg validation
  modified:
    - main.go - Entry point calling cmd.Execute()
    - go.mod - Dependencies (cobra, viper)

key-decisions:
  - "Use explicit config file path (./rtsp-recorder.yml) to avoid Viper finding binary"
  - "Conservative defaults: 60m duration, 1024MB max size, 3 retries"
  - "Config file is optional - tool works with flags/env vars alone"

patterns-established:
  - "initConfig uses PersistentPreRun for global config initialization"
  - "Validator package provides CheckFFmpeg() with actionable error messages"
  - "Config package returns struct with helper methods (GetFFmpegPath)"

requirements-completed: [CONF-01, CONF-03, REC-07, ERR-01]

# Metrics
duration: 15min
completed: 2025-04-02
---

# Phase 1 Plan 1: Foundation & Configuration Summary

**CLI foundation with Cobra scaffolding, Viper configuration (YAML + RTSP_RECORDER_* env vars), and ffmpeg pre-flight validation with actionable error messages**

## Performance

- **Duration:** 15 min
- **Started:** 2025-04-02T11:30:00Z
- **Completed:** 2025-04-02T11:45:00Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Go module initialized with Cobra CLI framework (v1.10.2)
- Viper configuration system supporting YAML config files and environment variables
- Config struct with mapstructure tags for all settings (URL, duration, max_file_size, retry_attempts, ffmpeg_path, filename_template)
- FFmpeg validation with version detection and actionable installation instructions
- validate subcommand that checks configuration and dependencies

## Task Commits

Each task was committed atomically:

1. **Task 1: Initialize Go project with Cobra CLI structure** - `438e9a6` (feat)
2. **Task 2: Create config package with Config struct and defaults** - `252089e` (feat)
3. **Task 3: Create ffmpeg validator and validate command** - `43dc9fe` (feat)

**Plan metadata:** `TBD` (docs: complete plan)

## Files Created/Modified

- `main.go` - Entry point calling cmd.Execute()
- `go.mod` / `go.sum` - Go module with cobra v1.10.2 and viper v1.21.0
- `cmd/root.go` - Root command with initConfig(), Viper setup, defaults, and comprehensive help
- `cmd/validate.go` - Validate subcommand for checking config and ffmpeg
- `config/config.go` - Config struct with Load() and helper methods
- `internal/validator/ffmpeg.go` - FFmpeg availability and version checking

## Decisions Made

- **Explicit config file path:** Changed from `SetConfigName("rtsp-recorder")` to `SetConfigFile("./rtsp-recorder.yml")` to prevent Viper from finding the `rtsp-recorder` binary and attempting to parse it as YAML.
- **Conservative defaults:** Set 60m duration, 1024MB max file size, 3 retry attempts as per D-05.
- **Structured output format:** Using [INFO] and [ERROR] prefixes for all user-facing messages per D-07/D-08.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed Viper config file discovery conflict**
- **Found during:** Task 1 (initConfig implementation)
- **Issue:** Viper's `SetConfigName("rtsp-recorder")` with `AddConfigPath(".")` caused it to find the `rtsp-recorder` binary (no extension) and attempt to parse it as YAML, resulting in "invalid trailing UTF-8 octet" error
- **Fix:** Changed to `SetConfigFile("./rtsp-recorder.yml")` for explicit file path, avoiding the binary name conflict
- **Files modified:** cmd/root.go
- **Verification:** Tested with and without config file - both scenarios work correctly
- **Committed in:** 43dc9fe (Task 3 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Fix was essential for correct operation. No scope creep.

## Issues Encountered

None - plan executed successfully after addressing the Viper config file discovery issue.

## User Setup Required

None - no external service configuration required.

FFmpeg is an external dependency that must be installed separately:
- **macOS:** `brew install ffmpeg`
- **Debian/Ubuntu:** `apt install ffmpeg`
- **Other:** Download from https://ffmpeg.org/download.html

The `validate` command will check for ffmpeg and provide installation instructions if not found.

## Verification Results

All success criteria verified:

1. ✓ `./rtsp-recorder --help` displays commands and config file example
2. ✓ `rtsp-recorder.yml` is recognized and loaded (shows [INFO] Using config file)
3. ✓ Environment variables work (`RTSP_RECORDER_DURATION=15m ./rtsp-recorder validate` shows correct duration)
4. ✓ `./rtsp-recorder validate` shows [INFO] lines for successful checks including ffmpeg version

## Next Phase Readiness

- CLI foundation complete with configuration system
- Ready for Phase 2: Core Recording Engine
- validate command provides early feedback on ffmpeg availability

---
*Phase: 01-foundation-configuration*
*Plan: 01*
*Completed: 2025-04-02*
