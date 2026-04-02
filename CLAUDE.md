<!-- GSD:project-start source:PROJECT.md -->
## Project

**rtsp-recorder**

A CLI tool that records RTSP video streams to MP4 files. Uses ffmpeg for encoding and supports flexible stop conditions including manual interruption, time limits, and file size limits. Configuration is managed via YAML file using the Viper library.

**Core Value:** Reliably capture RTSP streams to timestamped MP4 files with minimal setup and predictable behavior.

### Constraints

- **Tech stack**: Go, Cobra CLI, Viper, ffmpeg (external dependency)
- **Dependencies**: ffmpeg must be installed and available in PATH
- **Platform**: Cross-platform Go binary (Linux, macOS, Windows)
- **Output**: Single MP4 file per recording session
<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->
## Technology Stack

## Recommended Stack
### Core Technologies
| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| **Go** | 1.26.x (latest 1.26.1) | Primary language | Current stable with full support for generics, improved performance, and standard library enhancements. Go 1.26.0 is already installed on this system. |
| **Cobra** | v1.10.2 | CLI framework | Standard choice for Go CLI applications. Provides command structure, flag parsing, shell completions, and help generation. Used by Kubernetes, Hugo, and Docker. |
| **Viper** | v1.21.0 | Configuration management | The de facto standard for Go configuration. Supports YAML, JSON, env vars, and flags with automatic precedence. Seamlessly integrates with Cobra. |
| **FFmpeg** | 7.1+ (LTS) or 8.x (current) | Video encoding/RTSP handling | Industry standard for video processing. RTSP support is mature and battle-tested. Version 7.1.3 and 8.0+ are current stable releases as of March 2025. |
### Supporting Libraries
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `os/exec` | Standard library | Execute FFmpeg process | Always — the idiomatic way to run external commands in Go. Use `CommandContext` for timeout support. |
| `context` | Standard library | Cancellation/timeout handling | Always — essential for graceful shutdown and stop conditions (duration limits, Ctrl+C handling). |
| `signal` | Standard library | OS signal handling | For graceful interruption handling (SIGINT/SIGTERM). Needed for Ctrl+C stop condition. |
| `path/filepath` | Standard library | Cross-platform file paths | For timestamp-based filename generation that works on Linux, macOS, and Windows. |
| `time` | Standard library | Timestamp generation | For creating timestamp-based output filenames. |
| `fmt`, `log` | Standard library | Output and logging | Basic output and error logging. Consider `slog` (Go 1.21+) for structured logging if needed later. |
### Development Tools
| Tool | Purpose | Notes |
|------|---------|-------|
| `cobra-cli` | Project scaffolding | Install with `go install github.com/spf13/cobra-cli@latest`. Generates standard project structure with `cobra-cli init --viper`. |
| `golangci-lint` | Linting | Standard Go linter. Configuration can be added as `.golangci.yml`. |
| `go mod` | Dependency management | Use Go modules for reproducible builds. |
## Installation
# Initialize Go module (if not already done)
# Install cobra-cli generator
# Initialize Cobra project with Viper
# Core dependencies
### FFmpeg Installation (External Dependency)
## Project Structure
## Alternatives Considered
| Category | Recommended | Alternative | When to Use Alternative |
|----------|-------------|-------------|-------------------------|
| CLI Framework | Cobra | urfave/cli | urfave/cli v2 is simpler for very basic CLIs with <3 commands. Cobra wins for extensibility and standard patterns. |
| Config Library | Viper | koanf | koanf is lighter weight if only file-based config is needed. Viper provides better env/flag integration out-of-the-box. |
| Video Encoding | FFmpeg (external) | gstreamer (external) | GStreamer has better pipeline flexibility but FFmpeg is more ubiquitous and has simpler CLI interface for basic recording. |
| Native Go RTSP | FFmpeg wrapper | go-rtsp libraries | Native Go RTSP libraries exist but FFmpeg handles codec negotiation, reconnection, and format conversion reliably without custom code. |
## What NOT to Use
| Avoid | Why | Use Instead |
|-------|-----|-------------|
| **Native Go video encoding** | Implementing RTSP/MP4 handling natively requires complex codec knowledge, reconnection logic, and format handling. FFmpeg has solved these problems over 20+ years. | FFmpeg wrapper via `os/exec` |
| **Cobra v1.9.x or older** | v1.10.0+ includes important fixes and context propagation improvements. Older versions may have compatibility issues with newer Go versions. | Cobra v1.10.2 |
| **Viper v1.19.x or older** | v1.20.0+ includes major improvements to the encoding layer and drops deprecated HCL/Java properties support. Latest versions have better performance. | Viper v1.21.0 |
| **FFmpeg v5.x or older** | Older versions lack current codec optimizations and security fixes. RTSP handling has improved significantly in 6.x/7.x series. | FFmpeg 7.1+ or 8.x |
| **Direct shell execution** | Using `sh -c "ffmpeg ..."` introduces shell injection risks and portability issues. | `exec.Command()` with explicit arguments |
| **FFmpeg static builds for production** | Static builds may lack hardware acceleration support and specific codecs needed for your streams. | Use distribution packages when possible, or build with needed codecs. |
## Version Compatibility
| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| Go 1.26.x | Cobra v1.10.x, Viper v1.21.x | Full compatibility. Go 1.26.0+ tested in both Cobra and Viper CI. |
| Cobra v1.10.2 | Viper v1.21.0 | Both use spf13/pflag v1.0.10. No conflicts. |
| FFmpeg 7.1+ | All Go versions | External binary, no direct dependency. Version detection done at runtime. |
## Stack Patterns for This Domain
### Pattern 1: FFmpeg Process Management
### Pattern 2: Configuration Hierarchy
### Pattern 3: Signal Handling for Stop Conditions
### Pattern 4: FFmpeg Version Detection
## FFmpeg Command Pattern for RTSP Recording
- `-i <url>`: Input RTSP stream
- `-c copy`: Stream copy (no re-encoding, low CPU usage)
- `-f mp4`: Force MP4 output format
- `-movflags +faststart`: Enable web-optimized MP4 (moov atom at front, allows streaming playback)
- `-t <seconds>`: Duration limit (alternative to manual stop)
- `-fs <bytes>`: File size limit
- `-rtsp_transport tcp`: Force TCP transport (more reliable than UDP for some networks)
- `-stimeout <microseconds>`: RTSP socket timeout for connection failures
## Confidence Assessment
| Area | Level | Reason |
|------|-------|--------|
| **Go Version** | HIGH | Official releases page confirms 1.26.1 is current. System has 1.26.0 installed (compatible). |
| **Cobra Version** | HIGH | GitHub releases page confirms v1.10.2 is latest stable (Dec 2024). Verified with `go list`. |
| **Viper Version** | HIGH | GitHub releases page confirms v1.21.0 is latest stable (Sep 2024). Verified with `go list`. |
| **FFmpeg Integration** | HIGH | `os/exec` is standard library, well-documented. FFmpeg RTSP support is mature industry standard. |
| **CLI Patterns** | HIGH | Cobra + Viper combination is ubiquitous in Go CLI ecosystem with extensive documentation. |
## Sources
- https://go.dev/dl/ — Go 1.26.1 current stable release
- https://github.com/spf13/cobra/releases — Cobra v1.10.2 (Latest, Dec 2024)
- https://github.com/spf13/viper/releases — Viper v1.21.0 (Latest, Sep 2024)
- https://github.com/spf13/cobra-cli — Official CLI generator tool
- https://pkg.go.dev/os/exec — Standard library documentation for external process execution
- https://ffmpeg.org/download.html — FFmpeg 8.1, 7.1.3, 6.1.4 current stable releases
- https://trac.ffmpeg.org/wiki/Encode/H.264 — FFmpeg encoding best practices
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd:quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd:debug` for investigation and bug fixing
- `/gsd:execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd:profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
