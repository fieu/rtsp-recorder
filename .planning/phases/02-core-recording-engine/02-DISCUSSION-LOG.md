# Phase 2: Core Recording Engine - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2025-04-02
**Phase:** 2-Core Recording Engine
**Areas discussed:** FFmpeg command options, Stop condition coordination, Progress display format, MP4 finalization strategy, File size monitoring

---

## FFmpeg Command Options

| Option | Description | Selected |
|--------|-------------|----------|
| TCP transport | More reliable, handles packet loss better | ✓ |
| UDP transport | Lower latency, may drop frames | |
| Auto transport | Let ffmpeg choose | |

| Option | Description | Selected |
|--------|-------------|----------|
| Copy stream | No re-encoding, fastest, low CPU | ✓ |
| H.264 encoding | Re-encode for compatibility | |
| You decide | Based on stream format | |

| Option | Description | Selected |
|--------|-------------|----------|
| 5 second buffer | Balance latency and smoothness | ✓ |
| 10 second buffer | More buffer for unstable networks | |
| No buffer | Minimal latency | |

**User's choices:** TCP transport, stream copy mode, 5-second buffer
**Notes:** Lowest CPU usage, reliable transport, reasonable buffer

---

## Stop Condition Coordination

| Option | Description | Selected |
|--------|-------------|----------|
| First trigger wins | Stop as soon as any condition met | ✓ |
| Only allow one | Tool rejects multiple conditions | |
| Complex logic | AND/OR based on flags | |

| Option | Description | Selected |
|--------|-------------|----------|
| SIGINT then SIGTERM | Graceful shutdown with timeout | ✓ |
| SIGKILL only | Immediate termination | |
| You decide | Based on context | |

**User's choices:** First trigger wins, graceful shutdown with timeout
**Notes:** Multiple conditions can be active, whichever happens first stops recording

---

## Progress Display Format

| Option | Description | Selected |
|--------|-------------|----------|
| All metrics | Bytes, time, size, bitrate | ✓ |
| Time and size only | Minimal display | |
| You decide | What's most useful | |

| Option | Description | Selected |
|--------|-------------|----------|
| Every 1 second | Live feel | ✓ |
| Every 5 seconds | Less spam | |
| On change only | Event-driven | |

| Option | Description | Selected |
|--------|-------------|----------|
| Single line (overwrite) | Carriage return style | ✓ |
| New line per update | Traditional log style | |
| Progress bar | Visual bar | |

**User's choices:** All metrics, every 1 second, single line with carriage return
**Notes:** Professional looking output: `Recording: 1.2GB | 00:05:30 | 4.5Mbps`

---

## MP4 Finalization Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Graceful 10s timeout | SIGINT → wait 10s → SIGKILL | ✓ |
| Graceful 5s timeout | Faster fallback | |
| Immediate SIGTERM | Quick stop | |

| Option | Description | Selected |
|--------|-------------|----------|
| Always save partial | Partial file better than lost data | ✓ |
| Delete on error | Avoid confusion | |
| Atomic write with temp | Rename on success | |

**User's choices:** 10-second graceful timeout, always save partial
**Notes:** Gives ffmpeg time to write moov atom for valid MP4

---

## File Size Monitoring

| Option | Description | Selected |
|--------|-------------|----------|
| Poll every 1 second | Simple, reliable | ✓ |
| Poll every 5 seconds | Less IO | |
| Filesystem events | Efficient, platform-specific | |

**User's choice:** Poll every 1 second
**Notes:** Matches progress update frequency, simple implementation

---

## the agent's Discretion

- Exact ffmpeg flag ordering and additional optimization flags
- Progress output formatting details (precision, units)
- Exact polling interval timing (can vary ±200ms)
- Error recovery on partial write failures

## Deferred Ideas

- FFmpeg stderr parsing for accurate bitrate — Phase 3
- RTSP stream pre-check (DESCRIBE) — Phase 3
- Segmented recording — v2
