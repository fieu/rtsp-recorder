# Phase 4: Timelapse Recording - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2025-04-02
**Phase:** 4-Timelapse Recording
**Areas discussed:** Timelapse approach, Flag interaction design, Speed calculation method

---

## Timelapse Approach

| Option | Description | Selected |
|--------|-------------|----------|
| Real-time frame dropping | Drop frames during recording for direct timelapse output | ✓ |
| Post-recording processing | Record normally, process after | |
| Hybrid with user choice | Record full, generate timelapse, optionally delete | |

**User's choice:** Real-time frame dropping during recording
**Rationale:** More efficient, single pass, direct output

---

## Flag Interaction Design

| Option | Description | Selected |
|--------|-------------|----------|
| Independent flags | --duration = recording time, --timelapse = output duration | ✓ |
| Timelapse alone sets both | --timelapse X means record X*10, output X | |
| Timelapse as speed multiplier | --timelapse 60 means 60x speed | |

**User's choice:** Independent flags
**Rationale:** Clear separation of concerns, explicit control

**Example:** `--duration 1h --timelapse 10s` means record for 1 hour, output 10 seconds

---

## Speed Calculation Method

| Option | Description | Selected |
|--------|-------------|----------|
| Target output duration | --timelapse 10s means "output should be 10s long" | ✓ |
| Speed multiplier | --timelapse 10 means "10x speed" | |

**User's choice:** Target output duration
**Rationale:** More intuitive — user thinks "I want a 10 second video" not "I want 360x speed"

**Calculation:** speedup = record_duration / timelapse_duration

---

## the agent's Discretion

- Exact FFmpeg filter syntax for frame selection
- Whether to include audio and how to handle it
- Exact progress display format
- Frame calculation precision

## Deferred Ideas

- Post-processing timelapse from existing recordings
- Variable speed timelapse
- Timelapse preview during recording

