# Roadmap: rtsp-recorder

**Project:** rtsp-recorder  
**Created:** 2025-04-02  
**Granularity:** Coarse (3 phases)  
**Total Requirements:** 18 v1 requirements

---

## Phases

- [x] **Phase 1: Foundation & Configuration** — CLI scaffolding, config system, and pre-flight validation
- [x] **Phase 2: Core Recording Engine** — Record RTSP streams with stop conditions and graceful shutdown **(Complete)**
- [x] **Phase 3: Resilience & Feedback** — Retry logic, progress display, and edge case handling **(Complete)**

---

## Phase Overview

| Phase | Name | Goal | Requirements | Success Criteria |
|-------|------|------|--------------|------------------|
| 1 | Foundation & Configuration | Tool can be configured and validated before recording | Complete    | 2026-04-02 |
| 2 | Core Recording Engine | User can record RTSP streams with flexible stop conditions | Complete    | 2026-04-02 |
| 3 | Resilience & Feedback | Recording is robust with retry, progress visibility, and clean error handling | Complete    | 2026-04-02 |

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

**Plans:** 2/2 plans complete

Plans:
- [x] 01-01-PLAN.md — CLI scaffolding, config system, ffmpeg validation (Wave 1)
- [x] 01-02-PLAN.md — Record command with flag support, completing config precedence (Wave 2)

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

**Plans:** 3/3 plans complete

Plans:
- [x] 02-01-PLAN.md — FFmpeg process wrapper with graceful shutdown (Wave 1)
- [x] 02-02-PLAN.md — Stop conditions: signal, duration, file size monitors (Wave 1)
- [x] 02-03-PLAN.md — Recording orchestration and progress display (Wave 2)

**Wave Structure:**
```
Wave 1 (Parallel):
  02-01 (FFmpeg wrapper) ──┐
                           ├──→ Wave 2
  02-02 (Stop conditions) ──┘

Wave 2:
  02-03 (Orchestrator - depends on 02-01 and 02-02)
```

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

**Plans:** 2/2 plans complete

Plans:
- [x] 03-01-PLAN.md — RTSP validation & error classification (Wave 1)
- [x] 03-02-PLAN.md — Retry logic integration (Wave 2)

**Wave Structure:**
```
Wave 1 (Independent):
  03-01 (RTSP validation & error classification)

Wave 2 (Depends on Wave 1):
  03-02 (Retry logic integration - uses error classifier from 03-01)
```

---

### Phase 4: Timelapse Recording

**Goal:** User can create timelapse videos by recording for a duration and condensing to a shorter target duration

**Depends on:** Phase 2 (Core Recording Engine)

**Requirements:** TIMELAPSE-01, TIMELAPSE-02, TIMELAPSE-03

**Success Criteria** (what must be TRUE):

1. User can specify `--timelapse` or `-tl` flag with target duration (e.g., `--timelapse 10s` for 10 second output)
2. Tool records for the full configured duration, then condenses video to target timelapse duration
3. Output video plays at accelerated speed showing the condensed timeline
4. Timelapse works with all existing stop conditions (Ctrl+C, duration, file size)

**Plans:** 3 plans in 2 waves

**Wave Structure:**
```
Wave 1 (Parallel):
  04-01 (Config & flags) ──┐
                           ├──→ Wave 2
  04-02 (FFmpeg filter) ───┘

Wave 2:
  04-03 (Progress display & integration)
```

Plans:
- [ ] 04-01-PLAN.md — Timelapse config field and flag registration (Wave 1)
- [ ] 04-02-PLAN.md — FFmpeg timelapse filter implementation (Wave 1)
- [ ] 04-03-PLAN.md — Progress display and stop condition integration (Wave 2)

---

## Progress Tracking

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Configuration | 2/2 | **Complete** | 2025-04-02 |
| 2. Core Recording Engine | 3/3 | **Complete** | 2026-04-02 |
| 3. Resilience & Feedback | 2/2 | **Complete** | 2026-04-02 |
| 4. Timelapse Recording | 0/3 | Planned | — |

---

## Coverage Validation

| Requirement | Phase | Status |
|-------------|-------|--------|
| CONF-01 | Phase 1 | Complete (Plan 01-01) |
| CONF-02 | Phase 1 | Complete (Plan 01-02) |
| CONF-03 | Phase 1 | Complete (Plan 01-01) |
| CONF-04 | Phase 1 | Complete (Plan 01-02) |
| REC-01 | Phase 2 | Complete (Plan 02-03) |
| REC-02 | Phase 2 | Complete (Plan 02-03) |
| REC-03 | Phase 2 | Complete (Plan 02-03) |
| REC-04 | Phase 2 | Complete (Plan 02-03) |
| REC-05 | Phase 2 | Complete (Plan 02-03) |
| REC-06 | Phase 3 | Complete (Plan 03-02) |
| REC-07 | Phase 1 | Complete (Plan 01-01) |
| STOP-01 | Phase 2 | Complete (Plan 02-02) |
| STOP-02 | Phase 2 | Complete (Plan 02-02) |
| STOP-03 | Phase 2 | Complete (Plan 02-02) |
| STOP-04 | Phase 2 | Complete (Plan 02-02) |
| ERR-01 | Phase 1 | Complete (Plan 01-01) |
| ERR-02 | Phase 3 | Complete (Plan 03-01) |
| ERR-03 | Phase 2 | Complete (Plan 02-01) |
| ERR-04 | Phase 3 | Complete (Plan 03-01) |

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
