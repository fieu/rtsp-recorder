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

**Current Phase:** Phase 1 — Foundation & Configuration

**Current Plan:** 01-02 (ready to execute)

**Status:** Planned

**Progress:** 4/18 requirements complete, 1/8 plans complete

```
[██                  ] 22%
```

---

## Phase Tracking

| Phase | Name | Status | Req Complete | Plans Done |
|-------|------|--------|--------------|------------|
| 1 | Foundation & Configuration | In Progress | 4/6 | 1/2 |
| 2 | Core Recording Engine | Not started | 0/7 | 0/3 |
| 3 | Resilience & Feedback | Not started | 0/5 | 0/2 |

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Requirements completed | 4/18 |
| Phases completed | 0/3 |
| Plans completed | 1/8 |
| Success criteria verified | 4/4 |
| Defects found | 1 |
| Defects fixed | 1 |

---

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

**Last action:** Completed Plan 01-01 (Foundation & Configuration - CLI scaffolding, Viper config, ffmpeg validation)

**Next action:** Execute Plan 01-02

**Blockers:** None

**Working Notes:**

- Plan 01-01 complete: CLI foundation with Cobra, Viper config system, ffmpeg validation
- All 4 success criteria verified (--help, config file, env vars, validate command)
- Deviation: Fixed Viper config discovery to avoid binary name conflict
- Ready for Phase 1 Plan 2: Core recording command implementation

---

## Milestone Status

| Milestone | Target Phase | Status |
|-----------|--------------|--------|
| v1.0 MVP | Phase 3 | Not started |

---

*State initialized: 2025-04-02*
*Update this file after every phase transition and significant decision*
