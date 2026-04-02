# Roadmap: rtsp-recorder

**Project:** rtsp-recorder  
**Created:** 2025-04-02  
**Granularity:** Coarse (3 phases)  
**Total Requirements:** 18 v1 requirements

---

## Phases

- [ ] **Phase 1: Foundation & Configuration** — CLI scaffolding, config system, and pre-flight validation
- [ ] **Phase 2: Core Recording Engine** — Record RTSP streams with stop conditions and graceful shutdown
- [ ] **Phase 3: Resilience & Feedback** — Retry logic, progress display, and edge case handling

---

## Phase Overview

| Phase | Name | Goal | Requirements | Success Criteria |
|-------|------|------|--------------|------------------|
| 1 | Foundation & Configuration | Tool can be configured and validated before recording | 6 | 4 |
| 2 | Core Recording Engine | User can record RTSP streams with flexible stop conditions | 7 | 5 |
| 3 | Resilience & Feedback | Recording is robust with retry, progress visibility, and clean error handling | 5 | 4 |

---

## Phase Details

### Phase 1: Foundation & Configuration

**Goal:** User can install, configure, and validate the tool before attempting recording

**Depends on:** Nothing (first phase)

**Requirements:** CONF-01, CONF-02, CONF-03, CONF-04, REC-07, ERR-01

**Success Criteria** (what must be TRUE):

1. User can install the CLI binary and run `rtsp-recorder --help` to see available commands and flags
2. User can create a `rtsp-recorder.yml` config file with default settings that the tool recognizes
3. User can override any config value via CLI flags or environment variables (following precedence: flags > env > config > defaults)
4. Tool fails immediately with a clear error message if ffmpeg is not found in PATH, before attempting any recording

**Plans:** 2 plans

Plans:
- [ ] 01-01-PLAN.md — CLI scaffolding, config system, ffmpeg validation (Wave 1)
- [ ] 01-02-PLAN.md — Record command with flag support, completing config precedence (Wave 2)

---

### Phase 2: Core Recording Engine

**Goal:** User can successfully record RTSP streams with multiple stop conditions and clean output files

**Depends on:** Phase 1

**Requirements:** REC-01, REC-02, REC-03, REC-04, REC-05, STOP-01, STOP-02, STOP-03, STOP-04, ERR-03

**Success Criteria** (what must be TRUE):

1. User can start recording by providing an RTSP URL via CLI flag or config file
2. Tool records the stream to an MP4 file with an auto-generated timestamp-based filename (YYYY-MM-DD-HH-MM-SS.mp4) in the current directory
3. Recording stops gracefully when user presses Ctrl+C, with the MP4 file properly finalized and playable
4. Recording stops automatically when either the configured duration limit or file size limit is reached (whichever comes first)
5. Tool displays real-time progress showing bytes recorded, elapsed duration, and current file size during active recording
6. MP4 file remains valid and playable even if recording ends unexpectedly (unclean shutdown protection)

**Plans:** TBD

---

### Phase 3: Resilience & Feedback

**Goal:** Recording is reliable with automatic recovery from transient failures and clear error messages

**Depends on:** Phase 2

**Requirements:** REC-06, ERR-02, ERR-04

**Success Criteria** (what must be TRUE):

1. When network errors occur, tool automatically retries connection up to the configured number of attempts before giving up
2. Tool provides a descriptive, actionable error message when an invalid RTSP URL is provided (not just a generic failure)
3. Tool provides meaningful, specific error messages for common ffmpeg failures (connection refused, 404, invalid stream data) instead of just exit codes
4. Tool validates the RTSP stream is accessible (via DESCRIBE request) before starting recording to fail fast on bad URLs

**Plans:** TBD

---

## Progress Tracking

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Configuration | 0/2 | Not started | — |
| 2. Core Recording Engine | 0/3 | Not started | — |
| 3. Resilience & Feedback | 0/2 | Not started | — |

---

## Coverage Validation

| Requirement | Phase | Status |
|-------------|-------|--------|
| CONF-01 | Phase 1 | Pending |
| CONF-02 | Phase 1 | Pending |
| CONF-03 | Phase 1 | Pending |
| CONF-04 | Phase 1 | Pending |
| REC-01 | Phase 2 | Pending |
| REC-02 | Phase 2 | Pending |
| REC-03 | Phase 2 | Pending |
| REC-04 | Phase 2 | Pending |
| REC-05 | Phase 2 | Pending |
| REC-06 | Phase 3 | Pending |
| REC-07 | Phase 1 | Pending |
| STOP-01 | Phase 2 | Pending |
| STOP-02 | Phase 2 | Pending |
| STOP-03 | Phase 2 | Pending |
| STOP-04 | Phase 2 | Pending |
| ERR-01 | Phase 1 | Pending |
| ERR-02 | Phase 3 | Pending |
| ERR-03 | Phase 2 | Pending |
| ERR-04 | Phase 3 | Pending |

**Coverage Summary:**
- v1 requirements: 18 total
- Mapped to phases: 18 ✓
- Unmapped: 0 ✓

---

## Dependencies

```
Phase 1 (Foundation)
    ↓
Phase 2 (Core Recording)
    ↓
Phase 3 (Resilience)
```

**Build Order Logic:**
1. Configuration must exist before recording can be configured
2. FFmpeg validation must happen before recording attempts
3. Core recording must work before adding retry/resilience layers
4. Basic error handling (ffmpeg not found) needed before complex error cases

---

## Goal-Backward Summary

**When all phases complete, the user can:**

1. Install and configure the tool via YAML, CLI flags, or environment variables
2. Record any RTSP stream to timestamped MP4 files in the current directory
3. Stop recording via Ctrl+C, duration limit, or file size limit
4. See real-time progress during recording
5. Recover automatically from transient network failures via retry logic
6. Receive clear, actionable error messages for common failure modes
7. Trust that MP4 files will be valid even on unexpected shutdown

**Core Value Achieved:** Reliably capture RTSP streams to timestamped MP4 files with minimal setup and predictable behavior.

---

*Roadmap created: 2025-04-02*
