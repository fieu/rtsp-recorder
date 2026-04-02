# Phase 3: Resilience & Feedback - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2025-04-02
**Phase:** 3-Resilience & Feedback
**Areas discussed:** All areas deferred to agent discretion

---

## Discussion Summary

User deferred all gray areas to agent discretion. Decisions made based on:
- Research findings from PITFALLS.md
- Established patterns from Phases 1 and 2
- Standard Go practices for network operations
- Simple, predictable behavior over complex heuristics

---

## Decisions Made (Agent Discretion)

### Retry Logic
- **Decision:** Retry on network errors (connection refused, timeout, broken pipe)
- **Decision:** Use config retry-attempts (default 3), fixed 5s delay
- **Decision:** No retry on auth failures or permanent errors
- **Rationale:** Simple, predictable, aligns with conservative defaults (D-05)

### RTSP Pre-Validation
- **Decision:** DESCRIBE request with 10s timeout before starting ffmpeg
- **Decision:** "Accessible" = 200 OK with valid SDP
- **Rationale:** Fail fast pattern, prevents starting ffmpeg on bad URLs

### Error Specificity
- **Decision:** Parse ffmpeg stderr for 6 common error patterns
- **Decision:** Map each to actionable, user-friendly message
- **Rationale:** Research shows these patterns cover 90% of failures

### Failure Classification
- **Decision:** 5 categories (Network, Auth, Stream, Config, FFmpeg)
- **Rationale:** Distinguishes retryable from non-retryable errors

### FFmpeg Stderr Parsing
- **Decision:** Parse for both errors and bitrate (improves progress accuracy)
- **Rationale:** Deferred from Phase 2, now in scope for Phase 3

---

## the agent's Discretion

- Exact regex patterns for error detection
- Additional error patterns beyond the 6 listed
- Exact retry delay timing (can vary ±1 second)
- DESCRIBE request implementation details

## Deferred Ideas

None — Phase 3 completes v1 requirements.

