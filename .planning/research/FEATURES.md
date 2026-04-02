# Feature Landscape: RTSP Recording CLI Tools

**Domain:** RTSP Stream Recording CLI Tools
**Researched:** 2025-04-02
**Confidence:** HIGH (based on analysis of ffmpeg docs, MediaMTX, Motion, node-rtsp-recorder, Frigate, and industry patterns)

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **URL-based stream input** | Fundamental to RTSP; users expect to paste RTSP URL | LOW | Must support standard rtsp:// format with optional auth |
| **Manual stop (Ctrl+C)** | Unix standard for CLI processes; expected graceful shutdown | LOW | Critical for ffmpeg wrapper - must send SIGINT not SIGKILL |
| **Timestamp-based filenames** | Avoids overwrites, provides automatic organization | LOW | Default should be ISO 8601 or similar; users shouldn't think about naming |
| **MP4 output format** | Industry standard for recorded video; universal compatibility | LOW | ffmpeg handles this natively with -c copy or transcoding |
| **FFmpeg integration** | ffmpeg is the de facto standard for video processing | LOW | Users expect it; our project wraps ffmpeg rather than reimplementing |
| **Basic error handling** | Users need to know if stream is unreachable or ffmpeg missing | LOW | Clear error messages for common failure modes |
| **Single-stream recording** | Simplest use case; one URL = one output file | LOW | Concurrent multi-stream is advanced feature (v2) |
| **Stop conditions** | Users need control over when recording ends | MEDIUM | Duration limit, file size limit, and manual interrupt are standard |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Segmented recording** | Automatic file rotation prevents huge files; enables time-based organization | MEDIUM | node-rtsp-recorder uses `timeLimit` per segment; MediaMTX uses `recordSegmentDuration` |
| **YAML configuration** | Power users want repeatable, version-controlled setups | LOW | Viper library provides this; environment variable override is bonus |
| **Automatic retry/reconnection** | Network streams drop; automatic recovery is valuable | MEDIUM | MediaMTX supports this; requires careful state management |
| **Progress indication** | Users want visibility into recording status | LOW | Simple console output: bytes written, duration, frames |
| **Custom filename templates** | Power users want control over naming patterns | LOW | strftime-style templates (%Y-%m-%d_%H-%M-%S) |
| **Codec passthrough option** | Avoids transcoding overhead; preserves original quality | LOW | ffmpeg `-c copy` vs `-c:v libx264` |
| **Stream validation before recording** | Fail fast if stream is unreachable | LOW | Quick RTSP DESCRIBE check before starting ffmpeg |
| **Output directory configuration** | Users want organized storage structures | LOW | PROJECT.md defers this to v2, but it's common in other tools |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Native Go encoding (no ffmpeg)** | "Remove external dependency" | RTSP/MP4 handling is complex; ffmpeg is battle-tested | Keep ffmpeg as external dependency; validate on startup |
| **Real-time stream health monitoring** | "Show if stream is healthy" | Adds complexity; difficult to define "healthy" | Basic retry logic + clear error on failure |
| **Concurrent multi-stream recording** | "Record multiple cameras at once" | Significant complexity: resource management, multiple ffmpeg processes | Single-stream CLI; users can run multiple instances |
| **Motion detection / smart recording** | "Only record when something happens" | Entire product category (NVRs like Frigate); scope creep | Focus on reliable 24/7 recording; integrate with NVR if needed |
| **Built-in web interface** | "View streams and recordings" | Full web stack complexity; not a CLI tool anymore | Keep it CLI; users can use VLC, MediaMTX, or other tools for viewing |
| **Cloud upload integration** | "Automatically upload to S3/Drive" | Authentication complexity, bandwidth concerns, failure modes | Local recording only; separate tool for sync |
| **Transcoding to multiple formats** | "Create MP4, AVI, and MKV simultaneously" | Massive CPU overhead; niche use case | Single high-quality output; user can transcode post-recording |

## Feature Dependencies

```
[YAML Config]
    └──requires──> [Viper Integration]

[Automatic Retry]
    └──requires──> [Stream Validation]
        └──requires──> [FFmpeg Integration]

[Segmented Recording]
    └──requires──> [Timestamp-based Filenames]
        └──enhances──> [File Size Limits]

[Progress Indication]
    └──requires──> [FFmpeg stderr parsing]

[Custom Filename Templates]
    └──requires──> [Timestamp-based Filenames]

[Codec Passthrough]
    └──conflicts──> [Automatic Retry with transcoding]
        (passthrough requires different recovery strategy)

[Concurrent Multi-stream] ──conflicts──> [Simple CLI Design]
    (v2 feature; fundamentally changes architecture)
```

### Dependency Notes

- **Segmented Recording requires Timestamp-based Filenames:** Each segment needs unique, organized naming
- **Automatic Retry requires Stream Validation:** Must know if initial failure is recoverable
- **Codec Passthrough conflicts with complex retry:** If using `-c copy`, stream recovery is harder (key frame alignment issues)
- **Progress Indication requires FFmpeg stderr parsing:** Need to capture and interpret ffmpeg's progress output

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [x] **URL-based stream input** — CLI argument or config; fundamental feature
- [x] **FFmpeg integration** — Wraps ffmpeg for encoding; core of the tool
- [x] **MP4 output format** — Standard container; works everywhere
- [x] **Timestamp-based filenames** — Automatic organization; no naming decisions
- [x] **Manual stop (Ctrl+C)** — Standard Unix pattern; graceful shutdown
- [x] **Duration limit stop condition** — Users want time-bounded recordings
- [x] **File size limit stop condition** — Prevents disk exhaustion
- [x] **YAML configuration** — Viper integration for power users
- [x] **Early validation** — Check ffmpeg exists before starting
- [x] **Single stream recording** — Keep v1 focused and reliable

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] **Segmented recording** — Time-based file rotation (popular in node-rtsp-recorder, MediaMTX)
- [ ] **Progress indication** — Simple console output for active recording
- [ ] **Automatic retry with reconnection** — Handle temporary network drops
- [ ] **Custom filename templates** — strftime-style patterns
- [ ] **Codec passthrough option** — `-c copy` for zero CPU overhead
- [ ] **Stream validation before recording** — RTSP DESCRIBE check
- [ ] **Output directory configuration** — Allow custom paths beyond current directory

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] **Concurrent multi-stream recording** — Run multiple streams from one process
- [ ] **Recording scheduler** — Start/stop based on time rules
- [ ] **Integration hooks** — Run scripts on segment complete (like MediaMTX's `runOnRecordSegmentComplete`)
- [ ] **Format options** — MKV, AVI, segmented MPEG-TS (MediaMTX supports fMP4 and MPEG-TS)
- [ ] **RTSP server/proxy mode** — Become a media server (MediaMTX scope)
- [ ] **Retention policies** — Automatic deletion of old recordings (Motion, MediaMTX have this)

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| URL-based stream input | HIGH | LOW | P1 |
| FFmpeg integration | HIGH | LOW | P1 |
| MP4 output | HIGH | LOW | P1 |
| Timestamp-based filenames | HIGH | LOW | P1 |
| Manual stop (Ctrl+C) | HIGH | LOW | P1 |
| Duration limit | MEDIUM | LOW | P1 |
| File size limit | MEDIUM | LOW | P1 |
| YAML configuration | MEDIUM | LOW | P1 |
| Early ffmpeg validation | MEDIUM | LOW | P1 |
| Single stream focus | HIGH | LOW | P1 |
| Segmented recording | HIGH | MEDIUM | P2 |
| Progress indication | MEDIUM | LOW | P2 |
| Automatic retry | MEDIUM | MEDIUM | P2 |
| Custom filename templates | LOW | LOW | P2 |
| Codec passthrough | MEDIUM | LOW | P2 |
| Stream validation | MEDIUM | LOW | P2 |
| Output directory config | MEDIUM | LOW | P2 |
| Concurrent multi-stream | LOW | HIGH | P3 |
| Recording scheduler | LOW | HIGH | P3 |
| Integration hooks | LOW | MEDIUM | P3 |
| Multiple format options | LOW | MEDIUM | P3 |
| Retention policies | LOW | MEDIUM | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | MediaMTX | node-rtsp-recorder | Motion | Frigate | Our Approach (v1) |
|---------|----------|-------------------|--------|---------|-------------------|
| **Stream Input** | RTSP, RTMP, HLS, WebRTC, SRT | RTSP only | RTSP, v4l2, files | RTSP | RTSP only (focused) |
| **Recording Trigger** | Automatic, on-demand, hooks | Programmatic API | Motion detection, 24/7 | Motion detection, 24/7 | Manual CLI start, stop conditions |
| **Output Format** | fMP4, MPEG-TS | MP4 | MP4, AVI, etc. | MP4 | MP4 (focused) |
| **Segmentation** | Configurable duration | `timeLimit` option | `movie_max_time` | Configurable | Duration/file size limits |
| **Configuration** | YAML | JavaScript object | Text file | YAML | YAML (Viper) |
| **Multi-stream** | Yes (server) | Single | Yes | Yes | Single (by design) |
| **Motion Detection** | No | No | Yes (core feature) | Yes (AI-based) | No (out of scope) |
| **Retry/Reconnect** | Yes | No | Yes | Yes | Manual retry (v1) |
| **Progress Output** | Logs | None | Motion areas shown | UI dashboard | Console progress (v2) |
| **Web Interface** | Playback API | No | Web control | Full UI | No (CLI only) |

## Industry Patterns Observed

### From ffmpeg Documentation
- RTSP transport options: UDP, TCP (interleaved), HTTP tunneling, multicast
- Timeout configuration for network operations (`rw_timeout`, `timeout`)
- Reordering queue for UDP packet handling
- Connection retry options for live streams

### From MediaMTX
- Recording to fMP4 or MPEG-TS with configurable segment duration
- Time-based retention policies (`recordDeleteAfter`)
- Hooks for recording events (`runOnRecordSegmentCreate`, `runOnRecordSegmentComplete`)
- Playback server for accessing recordings via API

### From node-rtsp-recorder
- Simple programmatic API: `startRecording()`, `stopRecording()`
- `timeLimit` for automatic file rotation
- Custom filename format using moment.js patterns
- Audio-only recording mode
- Image capture mode (snapshot)

### From Motion
- Extensive configuration (150+ parameters)
- Motion detection as core differentiator
- Event-based recording with pre/post capture
- Multiple output formats and codecs
- Web control interface for configuration

### From Streamlink (CLI Patterns)
- Simple command: `streamlink URL quality`
- Focus on "pipe to player" or "write to file"
- Plugin system for different services
- Minimal flags for common operations

### From scrcpy (Recording Patterns)
- Simple recording: `--record=file.mp4`
- Codec selection: `--video-codec=h265`
- Resolution/framerate limiting for performance
- No transcoding by default (passthrough)

## Key Insights

1. **Simplicity wins for CLI tools**: Streamlink and scrcpy succeed with minimal surface area
2. **Configuration files matter for repeatability**: MediaMTX and Motion use config files; node-rtsp-recorder's programmatic API limits adoption
3. **Segmentation is expected**: Users don't want 100GB files; automatic rotation is table stakes for serious use
4. **Motion detection is a different product**: Frigate and Motion are full NVRs; trying to add this to a simple recorder creates scope explosion
5. **FFmpeg is the right foundation**: Attempting native encoding is an anti-pattern; ffmpeg's RTSP and MP4 handling is unmatched
6. **Progress visibility reduces anxiety**: Users want to know something is happening during long recordings

## Sources

- [FFmpeg Protocols Documentation](https://ffmpeg.org/ffmpeg-protocols.html) - RTSP options and transport methods
- [MediaMTX Configuration](https://raw.githubusercontent.com/bluenviron/mediamtx/main/mediamtx.yml) - Professional RTSP server features
- [node-rtsp-recorder](https://github.com/sahilchaddha/node-rtsp-recorder) - Simple JavaScript recorder implementation
- [Motion Configuration Guide](https://motion-project.github.io/motion_config.html) - Comprehensive NVR feature set
- [Frigate NVR](https://github.com/blakeblackshear/frigate) - Modern AI-powered NVR
- [Streamlink](https://github.com/streamlink/streamlink) - CLI streaming tool patterns
- [scrcpy](https://github.com/Genymobile/scrcpy) - Screen recording CLI patterns

---
*Feature research for: RTSP Recording CLI Tools*
*Researched: 2025-04-02*
