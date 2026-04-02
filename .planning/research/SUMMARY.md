# Research Summary: RTSP Recorder CLI

**Project:** rtsp-recorder  
**Domain:** RTSP Stream Recording CLI Tool  
**Synthesized:** 2026-04-02  
**Confidence:** HIGH

---

## Executive Summary

The RTSP Recorder CLI is a Go-based command-line tool that wraps FFmpeg to provide reliable RTSP stream recording with configurable stop conditions. Based on research across technology stacks, feature landscapes, architectural patterns, and domain pitfalls, the recommended approach is a **layered architecture using Go 1.26.x with Cobra CLI framework, Viper configuration management, and FFmpeg as the external encoding engine**.

This is a well-understood domain with established patterns. FFmpeg is the undisputed industry standard for RTSP handling and MP4 encoding—attempting native Go implementation would be an anti-pattern that introduces significant complexity and reliability issues. The CLI should follow Unix philosophy: do one thing well (reliably record single streams) with clear configuration options and graceful shutdown handling. Key risks center around **process management** (preventing zombie FFmpeg processes), **file integrity** (ensuring clean MP4 finalization), and **signal handling** (proper Ctrl+C behavior).

The research indicates this is a **HIGH confidence** project with mature, well-documented technologies and clear patterns from successful tools like Streamlink and scrcpy. Phase 1 should focus on core recording with robust stop conditions while avoiding the 10+ critical pitfalls identified around process management and FFmpeg integration.

---

## Key Findings

### From Technology Stack Research

**Core Technologies (All HIGH confidence):**

| Technology | Version | Rationale |
|------------|---------|-----------|
| **Go** | 1.26.x | Current stable with generics support; 1.26.0 already installed on system |
| **Cobra** | v1.10.2 | Standard Go CLI framework; provides command structure, completions, help generation |
| **Viper** | v1.21.0 | Industry-standard configuration with automatic precedence (flags > env > config > defaults) |
| **FFmpeg** | 7.1+ or 8.x | Industry standard for video processing; mature RTSP support with 20+ years battle-testing |

**Critical Integration Pattern:** FFmpeg process management via `os/exec.CommandContext()` with proper signal handling and `WaitDelay` for graceful shutdown. Using `sh -c` or shell execution is an anti-pattern that introduces injection risks.

**What NOT to Use:**
- Native Go video encoding (too complex, unreliable)
- Cobra/Viper versions older than specified (compatibility issues)
- FFmpeg v5.x or older (missing security fixes and RTSP improvements)
- Direct shell execution for FFmpeg (injection risks)

### From Feature Research

**MVP Must-Have (P1):**
- URL-based stream input (CLI argument)
- FFmpeg integration with proper wrapper
- MP4 output format (-movflags +faststart for web compatibility)
- Timestamp-based automatic filenames (ISO 8601 format)
- Manual stop via Ctrl+C with graceful shutdown
- Duration limit stop condition (Go context-based, NOT ffmpeg -t)
- File size limit stop condition
- YAML configuration via Viper
- Early FFmpeg availability validation
- Single stream focus (deliberate limitation for v1)

**v1.x Additions (P2):**
- Segmented recording (time-based file rotation)
- Progress indication (ffmpeg stderr parsing)
- Automatic retry with reconnection
- Custom filename templates (strftime-style)
- Codec passthrough option (-c copy)
- Stream validation before recording (RTSP DESCRIBE check)
- Output directory configuration

**Anti-Features to Avoid:**
- Native Go encoding (scope explosion, unreliable)
- Real-time stream health monitoring (complex, ill-defined)
- Concurrent multi-stream recording (significant architectural change)
- Motion detection (different product category—NVRs like Frigate)
- Built-in web interface (not a CLI tool anymore)
- Cloud upload integration (authentication complexity, scope creep)

**Industry Pattern:** Simplicity wins for CLI tools. Streamlink and scrcpy succeed with minimal surface area. Segmentation is expected for serious use—users don't want 100GB files.

### From Architecture Research

**Recommended Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Interface                          │
│                    (cmd/root.go, cmd/record.go)               │
└──────────────────────────┬──────────────────────────────────┘
                           │ uses
┌──────────────────────────▼──────────────────────────────────┐
│                   Configuration                             │
│                      (config/config.go)                       │
└──────────────────────────┬──────────────────────────────────┘
                           │ uses
┌──────────────────────────▼──────────────────────────────────┐
│                 Recording Service                           │
│              (recorder/recorder.go, stop_conditions.go)       │
└──────────────────────────┬──────────────────────────────────┘
                           │ uses
┌──────────────────────────▼──────────────────────────────────┐
│                  FFmpeg Wrapper                             │
│                     (ffmpeg/ffmpeg.go)                        │
└─────────────────────────────────────────────────────────────┘
```

**Component Boundary Rules:**
1. `cmd/` never imports `os/exec` directly—all external process logic through `ffmpeg/`
2. `recorder/` never imports Viper—receives config struct only
3. `ffmpeg/` is the only package importing `os/exec`
4. `config/` is the only package importing Viper

**Data Flow:** User Input → CLI Flags (Cobra) → Viper Merge (env→file→flag) → Config Struct → Recorder Orchestration → Stop Condition Monitors (Signal/Timer/File Watcher) → FFmpeg.Stop()

**Build Order:** Config → FFmpeg wrapper → Recorder (stop conditions + orchestration) → CLI commands → Entry point

### From Pitfalls Research

**Critical Pitfalls (Must Address in Phase 1):**

1. **Zombie FFmpeg Processes on Interrupt** — Use `signal.NotifyContext`, `CommandContext`, and implement custom `Cancel` function with `WaitDelay` for graceful shutdown
2. **MP4 File Corruption on Unclean Shutdown** — Always SIGINT/SIGTERM first, wait up to 5 seconds, then SIGKILL; use `-movflags +faststart`
3. **RTSP Connection Timeout Not Handled** — Set `-stimeout 5000000` (5s), use TCP transport (`-rtsp_transport tcp`), implement Go context timeout
4. **Signal.Notify with Unbuffered Channel** — Use buffered channel (size 1) or `signal.NotifyContext`
5. **Viper Configuration Case Sensitivity** — Use lowercase keys internally, `SetEnvPrefix`, explicit `BindEnv` calls
6. **Not Checking FFmpeg Exit Error Properly** — Parse stderr for specific patterns ("Connection refused", "404 Not Found", "Invalid data")
7. **ProcessGroup Leak on Unix** — Set `Setpgid: true` in `SysProcAttr`, kill negative PID to terminate entire process group
8. **Using -t Without Accounting for Startup Time** — Use Go `context.WithTimeout` instead of ffmpeg `-t` for accurate duration limits
9. **Not Validating FFmpeg Availability Early** — Check `exec.LookPath` in `init()`, fail fast with clear message
10. **Cobra/Viper Flag Binding Order** — Bind in `PersistentPreRun`, not `init()`, after flags are parsed

**Moderate Pitfalls:**
- UDP vs TCP transport selection (always use TCP for reliability)
- Filename timestamp collisions (use nanosecond precision)
- No progress feedback for long recordings (parse ffmpeg stderr)

---

## Implications for Roadmap

### Suggested Phase Structure

**Phase 1: Foundation & Core Recording (Weeks 1-2)**

*Rationale:* Configuration and FFmpeg wrapper are prerequisites for everything else. Core recording must solve the critical pitfalls (zombie processes, MP4 corruption, signal handling) from day one.

*Delivers:*
- Project scaffolding with Cobra CLI generator
- Configuration layer (Viper integration with YAML support)
- FFmpeg wrapper with proper process lifecycle management
- Signal handling for graceful shutdown
- Single URL recording with timestamped filenames
- Duration and file size stop conditions
- Early FFmpeg validation

*Features from FEATURES.md:* URL input, FFmpeg integration, MP4 output, timestamps, manual stop, duration limit, file size limit, YAML config, early validation

*Pitfalls to avoid:* #1 (zombie processes), #2 (MP4 corruption), #3 (RTSP timeout), #4 (signal channel buffering), #5 (Viper case sensitivity), #7 (process group leak), #8 (ffmpeg -t inaccuracy), #9 (late FFmpeg validation), #10 (flag binding order)

**Phase 2: Enhanced Recording (Week 3)**

*Rationale:* Segmented recording is an industry-expected feature (users don't want huge files). Progress indication addresses user anxiety during long recordings. Retry logic improves reliability.

*Delivers:*
- Segmented recording with configurable duration
- Progress output (bytes written, duration, frames)
- Automatic retry with reconnection logic
- Custom filename templates
- Codec passthrough option
- Stream validation (RTSP DESCRIBE check)
- Output directory configuration

*Features from FEATURES.md:* Segmented recording, progress indication, automatic retry, custom templates, codec passthrough, stream validation, output directory

*Pitfalls to avoid:* #6 (FFmpeg error parsing), #11 (UDP transport), #12 (filename collisions), #13 (no progress feedback)

**Phase 3: Polish & Future Prep (Week 4)**

*Rationale:* Testing, documentation, and architectural preparation for v2 features. Multi-stream support is a major architectural change deferred to v2.

*Delivers:*
- Comprehensive error handling and edge cases
- Cross-platform testing (Linux, macOS, Windows)
- Documentation and usage examples
- Architecture validation for future multi-stream support

*Features from FEATURES.md:* Deferred v2 features (multi-stream, scheduler, retention) require architectural changes and are out of scope for v1

*Pitfalls to avoid:* #14 (config file handling), #15 (concurrent config access for v2), #16 (file overwrites)

### Research Flags

**Needs Deeper Research:** None—all phases have HIGH confidence patterns from research. Standard Go/FFmpeg patterns are well-documented.

**Standard Patterns (Skip Additional Research):**
- Phase 1: Cobra/Viper patterns are ubiquitous in Go ecosystem
- Phase 1: FFmpeg process management is standard `os/exec` usage
- Phase 1: Signal handling follows Go official docs exactly
- Phase 2: Segmented recording is straightforward time-based logic
- Phase 2: Progress indication via ffmpeg stderr is documented pattern

**No Research Needed:** All technologies and patterns are mature with extensive documentation. Proceed directly to requirements definition.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| **Stack** | HIGH | Go, Cobra, Viper, FFmpeg are all mature with stable APIs. Version compatibility verified. |
| **Features** | HIGH | Clear feature categories based on industry tools (MediaMTX, Motion, node-rtsp-recorder, Frigate). MVP is well-defined. |
| **Architecture** | HIGH | Standard Go CLI patterns from Cobra conventions. Layered architecture follows kubectl, Hugo, Docker CLI models. |
| **Pitfalls** | HIGH | 10+ critical pitfalls identified with specific prevention strategies from official Go docs and FFmpeg documentation. |

### Gaps to Address

1. **None identified** — All areas have HIGH confidence with clear patterns and prevention strategies.
2. **FFmpeg version compatibility** — Should validate FFmpeg version at startup (suggest 7.1+ or 8.x) and warn on older versions.
3. **Cross-platform signal handling** — Test Windows vs Unix process group management (Windows doesn't support `Setpgid`).

### Research-to-Action Mapping

| Research Finding | Roadmap Action |
|-----------------|----------------|
| Cobra + Viper are standard | Use cobra-cli generator for project scaffolding |
| FFmpeg is required external dep | Check `exec.LookPath` at startup, fail fast with install instructions |
| Zombie processes are #1 pitfall | Implement `signal.NotifyContext` + custom `Cancel` + `WaitDelay` in Phase 1 |
| MP4 corruption on unclean shutdown | Always SIGTERM first, wait 5s, then SIGKILL; use `-movflags +faststart` |
| Segmented recording is expected | Include in Phase 2 as key differentiator |
| Multi-stream is v2 feature | Explicitly defer to maintain focus and avoid architecture complexity |

---

## Sources

### Technology Stack
- https://go.dev/dl/ — Go 1.26.1 current stable
- https://github.com/spf13/cobra/releases — Cobra v1.10.2
- https://github.com/spf13/viper/releases — Viper v1.21.0
- https://ffmpeg.org/download.html — FFmpeg stable releases
- https://pkg.go.dev/os/exec — Go process execution

### Feature Landscape
- https://ffmpeg.org/ffmpeg-protocols.html — RTSP options
- https://raw.githubusercontent.com/bluenviron/mediamtx/main/mediamtx.yml — MediaMTX configuration
- https://github.com/sahilchaddha/node-rtsp-recorder — JavaScript recorder patterns
- https://motion-project.github.io/motion_config.html — Motion NVR features
- https://github.com/blakeblackshear/frigate — Modern AI NVR
- https://github.com/streamlink/streamlink — CLI tool patterns
- https://github.com/Genymobile/scrcpy — Recording CLI patterns

### Architecture
- https://github.com/spf13/cobra/blob/main/site/content/user_guide.md — Cobra patterns
- https://github.com/spf13/viper/blob/master/README.md — Viper configuration
- https://pkg.go.dev/os/signal — Go signal handling
- https://pkg.go.dev/context — Context patterns
- https://go.dev/doc/effective_go.html — Go best practices

### Pitfalls
- https://pkg.go.dev/os/exec — Process management
- https://pkg.go.dev/os/signal — Signal handling
- https://github.com/spf13/viper/blob/master/TROUBLESHOOTING.md — Viper issues
- https://ffmpeg.org/ffmpeg-protocols.html — FFmpeg RTSP handling

---

*Research synthesis complete for: rtsp-recorder*  
*Synthesized: 2026-04-02*  
*Ready for: Requirements definition phase*
