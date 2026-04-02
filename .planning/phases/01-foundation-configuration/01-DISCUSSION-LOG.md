# Phase 1: Foundation & Configuration - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2025-04-02
**Phase:** 1-Foundation & Configuration
**Areas discussed:** CLI structure, Config file structure, Error message style, Help format, Config precedence

---

## CLI Structure

| Option | Description | Selected |
|--------|-------------|----------|
| Single command | Simple `rtsp-recorder` with flags only | |
| Subcommands | Separate commands for record, validate, config | ✓ |
| Other | Custom organization | |

**User's choice:** Subcommands pattern
**Notes:** Phase 1 includes `record`, `validate`, and `config` subcommands. Support both short and long flags.

---

## Config File Structure

| Setting | Description | Selected |
|---------|-------------|----------|
| url | RTSP stream URL to record from | ✓ |
| duration | Maximum recording duration in minutes | ✓ |
| max-file-size | Maximum file size in MB before stopping | ✓ |
| retry-attempts | Number of retry attempts on connection failure | ✓ |
| ffmpeg-path | Path to ffmpeg binary (optional, uses PATH by default) | ✓ |
| filename-template | Output filename template | ✓ |

**User's choice:** All six settings included in v1 config
**Notes:** Conservative defaults: 60 min duration, 1024 MB max size, 3 retries. Config file is optional.

---

## Error Message Style

| Option | Description | Selected |
|--------|-------------|----------|
| Plain text | Simple text like "Error: ffmpeg not found" | |
| With emoji | Emoji indicators like "❌ Error:" | |
| Structured with level tags | Format like "[ERROR] ffmpeg: not found" | ✓ |
| You decide | Let the agent choose | |

**User's choice:** Structured with level tags
**Notes:** Use [ERROR], [WARNING], [INFO] prefixes with context.

---

## Help Format

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal | Command, flags list, brief description | |
| With examples | Includes usage examples and config file example | ✓ |
| You decide | Let the agent choose | |

**User's choice:** With examples
**Notes:** Help text should show example config file and usage patterns.

---

## Config Precedence

| Option | Description | Selected |
|--------|-------------|----------|
| Strict hierarchy | CLI flags > env vars > config file > defaults | ✓ |
| Merge with last-wins | Combine all sources, last one wins | |
| You decide | Let the agent choose | |

**User's choice:** Strict hierarchy
**Notes:** Environment variables use RTSP_RECORDER_ prefix.

---

## the agent's Discretion

- Exact error message wording and formatting details
- Help text layout and organization
- Config file parsing error handling strategy

## Deferred Ideas

None — discussion stayed within phase scope.
