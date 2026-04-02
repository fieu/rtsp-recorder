---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
current_phase: 4 (Timelapse Recording)
current_plan: 04-03 complete
status: executing
last_updated: "2026-04-02T10:45:00.000Z"
progress:
  total_phases: 4
  completed_phases: 4
  total_plans: 10
  completed_plans: 10
  percent: 100
---

# State: rtsp-recorder

**Project:** rtsp-recorder  
**Created:** 2025-04-02  
**Mode:** yolo 

---

## Project Reference

**Core Value:** Reliably capture RTSP streams to timestamped MP4 files with minimal setup and predictable behavior.

**Tech Stack:** Go, Cobra CLI, Viper, ffmpeg (external dependency)

**Constraints:** 

- ffmpeg must be installed and available in PATH
- Cross-platform Go binary (Linux, macOS, Windows)
- Single MP4 file per recording session (single stream for v1)

---

## Current Position

**Current Phase:** 4 (Timelapse Recording)

**Current Plan:** 04-03 complete

**Status:** In Progress

**Progress:** [██████████] 100%

---

## Phase Tracking

| Phase | Name | Status | Req Complete | Plans Done |
|-------|------|--------|--------------|------------|
| 1 | Foundation & Configuration | **Complete** | 6/6 | 2/2 |
| 2 | Core Recording Engine | **Complete** | 7/7 | 3/3 |
| 3 | Resilience & Feedback | **Complete** | 5/5 | 2/2 |
| 4 | Timelapse Recording | **In Progress** | 3/3 | 3/3 |

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Requirements completed | 17/21 |
| Phases completed | 3/4 |
| Plans completed | 10/10 |
| Success criteria verified | 18/18 |
| Defects found | 2 |
| Defects fixed | 2 |

| Plan | Duration | Tasks |
|------|----------|-------|
| 03-01 | 148s | 3 |
| 03-02 | 180s | 3 |
| 04-01 | ~300s | 3 |
| 04-02 | 15m | 4 |
| 04-03 | 15m | 4 |

---
| Phase 04 P03 | 15m | 4 tasks | 4 files |

## Accumulated Context

### Decisions Made

| Decision | Rationale | Date |
|----------|-----------|------|
| Use ffmpeg over native Go | ffmpeg handles RTSP/MP4 encoding reliably, well-tested | 2025-04-02 |
| Single stream for v1 | Keeps initial scope focused, concurrent adds complexity | 2025-04-02 |
| Timestamp-based filenames | Automatic organization, no naming decisions needed | 2025-04-02 |
| YAML config with Viper | Standard Go config pattern, supports env override | 2025-04-02 |
| Explicit config file path | Avoid Viper finding binary "rtsp-recorder" and parsing as YAML | 2025-04-02 |
| Conservative defaults | 60m duration, 1024MB max, 3 retries for safe operation | 2025-04-02 |
| All flags have short forms | Short flags are convenient for frequent use (e.g., -d 30m) | 2025-04-02 |
| Positional URL argument | More intuitive than --url flag for primary input | 2025-04-02 |
| First trigger wins for stop conditions | Any one stopping condition causes all to stop (Ctrl+C OR duration OR file size) | 2025-04-02 |
| Use signal.NotifyContext | Go 1.16+ best practice, buffered internally avoids signal drops | 2025-04-02 |
| Go timer instead of ffmpeg -t | Avoids ffmpeg startup time inaccuracy, more precise | 2025-04-02 |
| Poll file size every 1 second | Balance between accuracy and system load | 2025-04-02 |
| Display all metrics in progress | Users need visibility into bytes, time, bitrate | 2026-04-02 |
| 1 second progress updates | Frequent enough for feedback, not too spammy | 2026-04-02 |
| Single line progress with \r | Clean terminal output without scrolling | 2026-04-02 |
| Final summary after recording | Users need confirmation of what was recorded | 2026-04-02 |
| RTSP validation before recording | Fail fast on bad URLs, save user time | 2026-04-02 |
| Use net.DialTimeout for DESCRIBE | Minimal dependencies, no external RTSP library needed | 2026-04-02 |
| Error classification by pattern | Enables retry logic and actionable messages | 2026-04-02 |
| ClassifiedError implements error | Seamless integration with Go error handling | 2026-04-02 |
| Fixed 5-second retry delay | Simple, predictable backoff per D-32 | 2026-04-02 |
| Retry only NetworkError category | Auth/Stream/Config errors fail immediately per D-33 | 2026-04-02 |
| Full Record() re-attempt on retry | Fresh ffmpeg process per attempt per D-34 | 2026-04-02 |
| RTSP validation inside retry loop | Fresh connectivity check each attempt per D-34 | 2026-04-02 |
| Signal context for graceful shutdown | Allows cancellation during retry delays | 2026-04-02 |
| Real-time frame dropping for timelapse | More efficient than post-processing per D-45 | 2026-04-02 |
| --timelapse value is target OUTPUT | User specifies desired output duration per D-52 | 2026-04-02 |
| Timelapse requires --duration | Cannot calculate speedup without recording duration per D-51 | 2026-04-02 |
| Minimum 1s timelapse duration | Prevent invalid ultra-short outputs per D-55 | 2026-04-02 |
| Timelapse filter: select+setpts | FFmpeg filter chain for frame dropping per D-56 | 2026-04-02 |
| Timelapse drops audio | Simpler approach, typical for timelapse videos per D-58 | 2026-04-02 |

### Open Questions

(None yet)

### Known Issues

| Issue | Status | Resolution |
|-------|--------|------------|
| Viper config discovery conflict | Fixed | Use SetConfigFile("./rtsp-recorder.yml") instead of SetConfigName to avoid binary conflict |

### Technical Debt

(None yet)

---

## Session Continuity

**Last action:** Completed Plan 04-03 (Timelapse progress display and stop condition integration)

**Next action:** Phase 4 completion review / Transition to Phase 5

**Blockers:** None

**Working Notes:**

- Plan 04-03 complete: Timelapse progress display and stop condition integration
- displayProgress(): Shows "[INFO] Recording: X elapsed | Output: ~Y | Zx speed | bytes | bitrate"
- runRecord(): Shows "Timelapse: 360x speed (1h -> 10s)" at recording start
- printFinalSummary(): Shows real duration, output duration, and speedup when timelapse enabled
- Stop conditions (Ctrl+C, duration, file size) verified compatible with timelapse
- 14 new tests added, all passing (58 recorder tests, 22 cmd tests)
- Phase 4 Timelapse Recording feature is complete
- All 3 plans in Phase 4 complete: 04-01 (config), 04-02 (ffmpeg filter), 04-03 (progress/stop)

---

*State updated: 2026-04-02 after Plan 04-03 completion*
