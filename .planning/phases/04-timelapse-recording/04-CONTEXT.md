# Phase 4: Timelapse Recording - Context

**Gathered:** 2025-04-02
**Status:** Ready for planning

<domain>
## Phase Boundary

Add timelapse recording capability to condense long recordings into shorter videos. This extends the record command with a `--timelapse` flag that drops frames in real-time during recording to produce accelerated output directly.

Depends on: Phase 2 (Core Recording Engine) — FFmpeg wrapper, recording orchestration

Extends: cmd/record.go — Add timelapse flag and processing

</domain>

<decisions>
## Implementation Decisions

### Timelapse Approach
- **D-45:** Real-time frame dropping during recording (not post-processing)
- **D-46:** Use FFmpeg select filter to drop frames at calculated interval
- **D-47:** Single output file (timelapse only, no separate full-speed recording)
- **D-48:** Frame selection: `select='not(mod(n\,X))'` where X = total_frames / output_frames

### Flag Interaction Design
- **D-49:** Independent flags:
  - `--duration` = how long to record (input duration)
  - `--timelapse` = target output duration
- **D-50:** Both flags can be used together: `--duration 1h --timelapse 10s`
- **D-51:** If only `--timelapse` provided without `--duration`, error with helpful message

### Speed Calculation Method
- **D-52:** `--timelapse` value represents target OUTPUT duration (e.g., `10s` = 10 second video)
- **D-53:** Speedup factor calculated automatically: `speedup = record_duration / timelapse_duration`
- **D-54:** Frame interval: keep every Nth frame where `N = total_expected_frames / target_frames`
- **D-55:** Minimum timelapse duration: 1 second (prevent invalid ultra-short outputs)

### FFmpeg Implementation
- **D-56:** Use `-vf "select='not(mod(n\,X))',setpts=N/(FRAME_RATE*TB)"` filter chain
- **D-57:** Calculate X based on frame rate and duration ratio
- **D-58:** Audio handling: either drop audio entirely (timelapse usually silent) or pitch-correct to match

### User Feedback
- **D-59:** Show calculated speedup factor: `[INFO] Timelapse: 360x speed (1h -> 10s)`
- **D-60:** Progress shows both real elapsed and estimated output time

### the agent's Discretion
- Exact FFmpeg filter syntax details
- Whether to include audio (and how to handle it)
- Exact progress display format for timelapse mode
- Frame calculation precision (round up/down handling)

</decisions>

<canonical_refs>
## Canonical References

### Project Requirements
- `.planning/ROADMAP.md` §Phase 4 — Timelapse Recording requirements

### Prior Phase Context
- `.planning/phases/02-core-recording-engine/02-CONTEXT.md` — Phase 2 decisions (D-13 through D-29)
- `.planning/phases/03-resilience-feedback/03-CONTEXT.md` — Phase 3 decisions (D-30 through D-44)

### Reusable Components
- `cmd/record.go` — Record command (add timelapse flag)
- `ffmpeg/ffmpeg.go` — FFmpeg wrapper (add timelapse filter option)
- `config/config.go` — Config struct (add TimelapseDuration field)

### FFmpeg Documentation
- `select` filter: frame selection based on expression
- `setpts` filter: adjust presentation timestamps

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `cmd/record.go` — Record command with flag registration (add `--timelapse` to init())
- `config/config.go` — Config struct (add `TimelapseDuration time.Duration` field)
- `config.BindFlags()` — Add timelapse flag binding
- `recorder/recorder.go` — Recording orchestrator (pass timelapse to ffmpeg)
- `ffmpeg/ffmpeg.go` — FFmpeg wrapper (add timelapse filter generation)

### Established Patterns
- Flag naming: both short and long (`-tl`, `--timelapse`)
- Duration parsing: `time.ParseDuration` compatible strings (e.g., "10s", "5m")
- FFmpeg args: built in `buildArgs()` method

### Integration Points
- Timelapse flag registered in `cmd/record.go` init()
- Config field added to `config.Config` struct
- FFmpeg wrapper checks if timelapse enabled, adds `-vf select` filter
- Progress display updated to show timelapse info

</code_context>

<specifics>
## Specific Ideas

- Example usage: `rtsp-recorder record --duration 1h --timelapse 10s rtsp://camera/stream`
- Output: 10-second video showing 1 hour of real-time activity at 360x speed
- Show calculated speed in progress: `Recording: 1h elapsed | Output: ~10s | 360x`
- Consider dropping audio for timelapse (simpler, more typical for timelapse videos)

</specifics>

<deferred>
## Deferred Ideas

- Post-processing timelapse (generate from existing recording) — out of scope
- Variable speed timelapse (speed up/slow down sections) — v2 feature
- Timelapse preview during recording — too complex for v1.x

</deferred>

---

*Phase: 04-timelapse-recording*
*Context gathered: 2025-04-02*
