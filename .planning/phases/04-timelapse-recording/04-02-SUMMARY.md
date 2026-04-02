---
phase: "04"
plan: "02"
name: "timelapse-recording"
subsystem: "ffmpeg"
tags: ["timelapse", "filter", "frame-selection"]
dependencies:
  requires: ["04-01"]
  provides: ["timelapse-filter", "speedup-calculation"]
  affects: ["ffmpeg.Cmd", "ffmpeg.buildArgs"]
tech-stack:
  added: []
  patterns: ["TDD", "real-time-frame-dropping"]
key-files:
  created: []
  modified:
    - "ffmpeg/ffmpeg.go"
    - "ffmpeg/ffmpeg_test.go"
decisions:
  - "D-56: Use FFmpeg select filter with setpts for timelapse"
  - "D-58: Drop audio for timelapse recordings"
metrics:
  duration: "15m"
  completed: "2026-04-02T10:35:00Z"
---

# Phase 04 Plan 02: FFmpeg Timelapse Filter Implementation

**One-liner:** Implemented FFmpeg timelapse filter generation using select/setpts filters and frame interval calculation based on duration ratio.

---

## What Was Built

### Core Functionality

Implemented timelapse support in the FFmpeg wrapper:

1. **CalculateFrameInterval()** - Utility function that calculates the frame selection interval for timelapse based on the speedup factor (record_duration / timelapse_duration). Returns N for `select='not(mod(n,N))'`.

2. **ffmpeg.Cmd.timelapseInterval** - Added field to store the calculated frame interval (1 = keep all frames, N = keep every Nth frame).

3. **buildArgs() integration** - Modified to conditionally add timelapse filter when timelapseInterval > 1:
   - Filter format: `select='not(mod(n,X))',setpts=N/(FRAME_RATE*TB)`
   - Per D-56: Uses FFmpeg select and setpts filters for real-time frame dropping
   - Drops audio entirely (-an) for timelapse recordings per D-58

4. **Getter methods** - Added GetTimelapseInterval() and GetSpeedupFactor() for progress display integration per D-59.

### Technical Decisions

- **Filter approach:** Used select + setpts filter chain per D-56 rather than post-processing
- **Audio handling:** Disabled audio (-an) for timelapse recordings as it's simpler and more typical for timelapse videos
- **Calculation method:** Frame interval based on speedup factor (duration ratio), not frame rate per D-52, D-53, D-54

---

## Key Files Modified

| File | Lines | Purpose |
|------|-------|---------|
| `ffmpeg/ffmpeg.go` | +85/-16 | CalculateFrameInterval function, timelapseInterval field, buildArgs integration, getter methods |
| `ffmpeg/ffmpeg_test.go` | +175/-0 | 16 new test functions for timelapse functionality |

---

## Commits

| Hash | Type | Description |
|------|------|-------------|
| 264e65d | test | CalculateFrameInterval function with tests |
| 628cb20 | feat | Add timelapseInterval field to Cmd, New() calculation |
| d0ae178 | feat | Integrate timelapse filter into buildArgs() |
| 79431fb | feat | Add GetSpeedupFactor() and GetTimelapseInterval() methods |

---

## Deviations from Plan

None - plan executed exactly as written.

---

## Known Stubs

None - all functionality fully implemented.

---

## Test Coverage

| Test Category | Count | Status |
|---------------|-------|--------|
| CalculateFrameInterval | 6 | All passing |
| TimelapseInterval field | 4 | All passing |
| buildArgs with timelapse | 5 | All passing |
| SpeedupFactor methods | 3 | All passing |
| **Total** | **18** | **All passing** |

Total ffmpeg package tests: 46 (all passing)

---

## Self-Check: PASSED

- [x] CalculateFrameInterval function exists with proper calculation logic
- [x] Cmd struct has timelapseInterval field, initialized in New()
- [x] buildArgs includes timelapse filter when interval > 1
- [x] buildArgs excludes -vf when timelapse disabled
- [x] buildArgs includes -an when timelapse enabled
- [x] buildArgs excludes -an and includes -c:a aac when timelapse disabled
- [x] GetSpeedupFactor() returns correct values
- [x] GetTimelapseInterval() returns correct values
- [x] All tests pass
- [x] Build succeeds
- [x] No breaking changes to existing functionality

---

## Integration Notes

This plan provides the core timelapse FFmpeg infrastructure. The next plan (04-03) will integrate this with:
- Progress display to show timelapse info per D-59
- Stop condition handling for timelapse mode
- Command-line error handling for timelapse validation

The timelapse configuration (04-01) added the --timelapse flag and config field. This plan (04-02) provides the FFmpeg filter implementation. Both are Wave 1 and can proceed independently.

---

*Summary created: 2026-04-02*
*Phase: 04-timelapse-recording*
