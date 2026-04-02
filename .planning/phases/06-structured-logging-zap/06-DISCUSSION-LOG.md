# Phase 6: Structured Logging with Zap - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2025-04-02
**Phase:** 6-Structured Logging with Zap
**Areas discussed:** Zap integration, configuration, log level usage, backward compatibility

---

## Discussion Summary

All areas deferred to agent discretion based on standard practices and prior phase patterns.

---

## Decisions Made (Agent Discretion)

### Zap Integration
- **Decision:** Use Uber's zap library with structured JSON in production, human-readable in dev
- **Decision:** 4 log levels: debug, info, warn, error
- **Rationale:** Industry standard, battle-tested in production

### Configuration
- **Decision:** `--log-level` flag (no shorthand), `RTSP_RECORDER_LOG_LEVEL` env var, `log_level` in YAML
- **Decision:** Default level: "info"
- **Rationale:** Standard viper precedence, follows Phase 1 config patterns

### Log Level Usage
- **Decision:** Debug = internal details, Info = user status, Warn = recoverable issues, Error = fatal
- **Rationale:** Standard logging practices, clear separation of concerns

### Progress Display
- **Decision:** Progress bar stays on stdout (not logged)
- **Rationale:** Real-time feedback should not be mixed with structured logs

---

## the agent's Discretion

- Exact zap configuration details
- Logger field naming conventions
- Additional log context fields
- Whether to support log file output

## Deferred Ideas

- Log rotation and file output — defer to later
- Remote log aggregation — out of scope
- Request tracing — not needed for CLI

