---
phase: 08-periodic-progress-logging
plan: 01
type: execute
subsystem: recorder
requirements:
  - P8-01
  - P8-02
  - P8-03
  - P8-04
tags: [progress, logging, zerolog, config]
dependencies:
  requires: [02-core-recording-engine]
  provides: [periodic-progress-logging]
  affects: [recorder, config]
tech-stack:
  added: []
  patterns:
    - time.Ticker for periodic logging (per D-108)
    - zerolog structured logging (per D-104)
    - Configurable interval via Viper (per D-99)
key-files:
  created: []
  modified:
    - config/config.go
    - cmd/root.go
    - recorder/recorder.go
decisions:
  - D-96: Remove \r progress bar (no option to keep it)
  - D-97: Periodic log messages every X seconds
  - D-99: Config field `progress_interval` (duration string)
  - D-100: Default 10 seconds
  - D-102: Set to 0 to disable
  - D-103: Include elapsed time, bytes, file size, bitrate
  - D-104: Structured logging with zerolog
  - D-108: Use time.Ticker
  - D-109: Stop ticker when recording stops
  - D-110: Log immediately at start
metrics:
  duration: ~5m
  completed_date: 2026-04-02
---

# Phase 08 Plan 01: Periodic Progress Logging Summary

**One-liner:** Replaced live \r progress bar with periodic structured zerolog messages at configurable intervals (default 10s, 0=disabled).

## What Was Built

### 1. Config Field and Flag (Task 1)

Added `ProgressInterval` configuration option to control how often progress is logged:

**config/config.go:**
- Added `ProgressInterval time.Duration` field with `mapstructure:"progress_interval"` tag
- Set default to `10 * time.Second` in `DefaultConfig()` (per D-100)
- Added `--progress-interval` (`-p`) flag binding in `BindFlags()` with 10s default

**cmd/root.go:**
- Added `viper.SetDefault("progress_interval", 10*time.Second)` in `initConfig()`

### 2. Periodic Structured Logging (Task 2)

Rewrote `displayProgress()` to use zerolog instead of fmt.Printf with carriage returns:

**recorder/recorder.go:**
- Modified `displayProgress()` to use `r.config.ProgressInterval` instead of hardcoded 1s
- Added early return if `ProgressInterval <= 0` (per D-102: disables progress logging)
- Log immediately at start via `r.logProgress()` call before ticker loop (per D-110)
- Created new `logProgress()` helper with structured zerolog fields:
  - `elapsed`: Duration since start
  - `bytes`: Raw byte count
  - `size`: Human-readable size string
  - `bitrate_bps`: Bitrate in bits per second
  - `bitrate`: Human-readable bitrate string
  - `speedup`: Timelapse speedup factor (when > 1)
  - `output_duration`: Estimated timelapse output duration (when speedup > 1)
- Message: `"Recording progress"` with all fields as structured data
- Removed `fmt.Println()` after progress display (no longer needed without \r overwrite)
- Ticker stopped via `defer ticker.Stop()` as before (per D-109)

## Key Files Modified

| File | Changes |
|------|---------|
| `config/config.go` | +4 lines: ProgressInterval field, default, flag binding |
| `cmd/root.go` | +1 line: viper.SetDefault for progress_interval |
| `recorder/recorder.go` | ~60 lines changed: Rewrote displayProgress(), added logProgress() |

## Design Decisions Applied

| Decision | Implementation |
|----------|---------------|
| D-96: Remove \r progress bar | All `fmt.Printf("\r...")` calls replaced with zerolog |
| D-97: Periodic messages | `time.Ticker` with configurable interval |
| D-99: Config field | `progress_interval` duration string (e.g., "10s", "1m") |
| D-100: Default 10s | `10 * time.Second` in DefaultConfig() and flag |
| D-102: 0 = disabled | Early return in displayProgress() if <= 0 |
| D-103: Include metrics | All metrics logged: elapsed, bytes, size, bitrate, timelapse fields |
| D-104: Structured zerolog | `r.logger.Info().Dur().Int64().Str().Float64().Msg()` |
| D-108: time.Ticker | `time.NewTicker(r.config.ProgressInterval)` |
| D-109: Stop ticker | `defer ticker.Stop()` in displayProgress() |
| D-110: Log at start | `r.logProgress()` called before entering ticker loop |

## Verification Results

```bash
# Build succeeds
$ go build ./...
# (no errors)

# Config has ProgressInterval field
$ grep -n "ProgressInterval" config/config.go
46:  // ProgressInterval is the interval for progress log messages (0 = disabled)
47:  ProgressInterval time.Duration `mapstructure:"progress_interval"`
95:    ProgressInterval:  10 * time.Second, // Default 10s per D-100

# Flag binding exists
$ grep -n "progress_interval" config/config.go cmd/root.go
config/config.go:47:    mapstructure:"progress_interval"`
config/config.go:137:  viper.BindPFlag("progress_interval", cmd.Flags().Lookup("progress-interval"))
cmd/root.go:97:       viper.SetDefault("progress_interval", 10*time.Second)

# No \r fmt.Printf remains
$ grep -n 'fmt.Printf.*\\r' recorder/recorder.go
# (no output)

# Structured logging present
$ grep -n "Recording progress" recorder/recorder.go
207:    event.Msg("Recording progress")

# ProgressInterval used in ticker
$ grep -n "ProgressInterval" recorder/recorder.go
143:    // Per D-102: Skips entirely if ProgressInterval is 0.
146:    if r.config.ProgressInterval <= 0 {
150:    ticker := time.NewTicker(r.config.ProgressInterval)
```

## Self-Check: PASSED

- [x] Config has ProgressInterval field with 10s default (per D-100)
- [x] Flag --progress-interval registered and bound (per D-99)
- [x] recorder.displayProgress() uses time.Ticker with configurable interval (per D-108)
- [x] Progress logs use structured zerolog fields (per D-104)
- [x] Setting interval to 0 disables progress logging (per D-102)
- [x] First log appears at start (per D-110)
- [x] No \r carriage return progress bar remains (per D-96)
- [x] Final summary still prints at end (per D-107) - unchanged in printFinalSummary()

## Deviations from Plan

**None** - Plan executed exactly as written.

## Commits

| Commit | Message | Files |
|--------|---------|-------|
| `1b7528e` | feat(08-01): add ProgressInterval config field and flag | config/config.go, cmd/root.go |
| `0285f89` | feat(08-01): replace progress bar with periodic zerolog logging | recorder/recorder.go |

## Example Output

With default 10s interval:
```
INF Recording progress elapsed=0s bytes=0 size="0 B"
INF Recording progress elapsed=10s bytes=1048576 size="1.0 MB" bitrate_bps=838860.8 bitrate="838.9 Kbps"
INF Recording progress elapsed=20s bytes=2097152 size="2.0 MB" bitrate_bps=838860.8 bitrate="838.9 Kbps"
```

With timelapse enabled:
```
INF Recording progress elapsed=30s bytes=5242880 size="5.0 MB" bitrate_bps=1398101.3 bitrate="1.4 Mbps" speedup=30 output_duration=1s
```

## Notes

- The progress interval accepts any valid Go duration string: "10s", "30s", "1m", "0" (disabled)
- Log output adapts to TTY (human-readable) vs non-TTY (JSON) via zerolog ConsoleWriter
- Timelapse fields only appear when speedup > 1 (i.e., timelapse mode is active)
