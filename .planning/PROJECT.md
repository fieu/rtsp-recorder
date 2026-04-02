# rtsp-recorder

## What This Is

A CLI tool that records RTSP video streams to MP4 files. Uses ffmpeg for encoding and supports flexible stop conditions including manual interruption, time limits, and file size limits. Configuration is managed via YAML file using the Viper library.

## Core Value

Reliably capture RTSP streams to timestamped MP4 files with minimal setup and predictable behavior.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] CLI accepts RTSP URL as argument or from config
- [ ] Records stream to MP4 using ffmpeg
- [ ] Generates timestamp-based output filenames
- [ ] Supports stop conditions: Ctrl+C, duration limit, max file size
- [ ] Configuration via rtsp-recorder.yml (Viper)
- [ ] Fails early with clear error if ffmpeg not installed
- [ ] Single stream recording (concurrent streams deferred)
- [ ] Output to current working directory

### Out of Scope

- Concurrent multiple stream recording — defer to v2, adds complexity
- Native Go encoding without ffmpeg — ffmpeg is the requirement
- Custom output directories per recording — current directory only for v1
- Real-time stream health monitoring — focus on recording reliability first

## Context

Built with Go 1.25.0 using Cobra CLI framework. The tool wraps ffmpeg for robust video encoding rather than implementing RTSP/MP4 handling natively. Viper provides flexible configuration management supporting both file-based and environment-based config.

## Constraints

- **Tech stack**: Go, Cobra CLI, Viper, ffmpeg (external dependency)
- **Dependencies**: ffmpeg must be installed and available in PATH
- **Platform**: Cross-platform Go binary (Linux, macOS, Windows)
- **Output**: Single MP4 file per recording session

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use ffmpeg over native Go | ffmpeg handles RTSP/MP4 encoding reliably, well-tested | — Pending |
| Single stream for v1 | Keeps initial scope focused, concurrent adds complexity | — Pending |
| Timestamp-based filenames | Automatic organization, no naming decisions needed | — Pending |
| YAML config with Viper | Standard Go config pattern, supports env override | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2025-04-02 after initialization*
