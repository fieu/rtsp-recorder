# Phase 1: Foundation & Configuration - Context

**Gathered:** 2025-04-02
**Status:** Ready for planning

<domain>
## Phase Boundary

CLI scaffolding, configuration system using Viper with YAML support, and pre-flight validation including ffmpeg installation check. This phase establishes the CLI structure and config loading before any recording functionality.

</domain>

<decisions>
## Implementation Decisions

### CLI Structure
- **D-01:** Use subcommands pattern, not single command
- **D-02:** Phase 1 includes three subcommands: `record`, `validate`, `config`
- **D-03:** Support both short and long flags (e.g., `-u` and `--url`)

### Config File Structure
- **D-04:** Config file `rtsp-recorder.yml` supports these settings:
  - `url`: RTSP stream URL to record from
  - `duration`: Maximum recording duration in minutes
  - `max-file-size`: Maximum file size in MB before stopping
  - `retry-attempts`: Number of retry attempts on connection failure
  - `ffmpeg-path`: Path to ffmpeg binary (optional, uses PATH by default)
  - `filename-template`: Output filename template
- **D-05:** Conservative defaults: 60 min duration, 1024 MB max size, 3 retries
- **D-06:** Config file is optional — tool works with flags alone

### Error Message Style
- **D-07:** Use structured format with level tags: `[ERROR]`, `[WARNING]`, `[INFO]`
- **D-08:** Error messages include context (e.g., `[ERROR] ffmpeg: not found in PATH`)

### Help Format
- **D-09:** Help text includes detailed usage examples
- **D-10:** Show example config file in help output

### Config Precedence
- **D-11:** Strict hierarchy: CLI flags > environment variables > config file > defaults
- **D-12:** Environment variables use `RTSP_RECORDER_` prefix

### the agent's Discretion
- Exact error message wording and formatting details
- Help text layout and organization
- Config file parsing error handling strategy

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Requirements
- `.planning/REQUIREMENTS.md` §Configuration (CONF-01 through CONF-04) — Config requirements
- `.planning/REQUIREMENTS.md` §Recording Core (REC-07) — FFmpeg validation requirement
- `.planning/REQUIREMENTS.md` §Error Handling (ERR-01) — Early failure requirement

### Research Insights
- `.planning/research/STACK.md` — Technology stack recommendations
- `.planning/research/ARCHITECTURE.md` — Go CLI patterns and Viper integration
- `.planning/research/PITFALLS.md` §Viper configuration pitfalls — Critical config loading gotchas

### Go Libraries
- Cobra v1.10.2 — CLI framework documentation
- Viper v1.21.0 — Configuration library documentation

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- (None — greenfield project)

### Established Patterns
- (None — greenfield project, follow standard Go CLI conventions)

### Integration Points
- FFmpeg must be available in system PATH (validated at startup)
- Config file in current directory or user home directory

</code_context>

<specifics>
## Specific Ideas

- Tool should feel like standard Unix CLI tools (simple, composable)
- Config file should be self-documenting with comments showing available options
- Error messages should guide user to fix the problem (e.g., "Install ffmpeg: apt install ffmpeg")

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-foundation-configuration*
*Context gathered: 2025-04-02*