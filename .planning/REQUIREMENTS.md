# Requirements: rtsp-recorder

**Defined:** 2025-04-02
**Core Value:** Reliably capture RTSP streams to timestamped MP4 files with minimal setup and predictable behavior

## v1 Requirements

### Configuration (CONF)

- [x] **CONF-01**: User can define settings in `rtsp-recorder.yml` file
- [x] **CONF-02**: User can override config values via CLI flags
- [x] **CONF-03**: User can set values via environment variables
- [x] **CONF-04**: Configuration is loaded using Viper library with proper precedence (flags > env > config > defaults)

### Recording Core (REC)

- [x] **REC-01**: User can specify RTSP stream URL via flag or config file
- [x] **REC-02**: Tool records stream to MP4 using ffmpeg subprocess
- [x] **REC-03**: Output filename uses timestamp format (YYYY-MM-DD-HH-MM-SS.mp4)
- [x] **REC-04**: Output files are saved to current working directory
- [x] **REC-05**: Tool displays progress (bytes recorded, duration, current file size)
- [x] **REC-06**: Tool automatically retries connection on network errors (configurable attempts)
- [x] **REC-07**: Tool validates ffmpeg is installed before starting recording

### Stop Conditions (STOP)

- [x] **STOP-01**: Recording stops gracefully when user presses Ctrl+C (SIGINT)
- [x] **STOP-02**: User can specify maximum recording duration in minutes via flag or config
- [x] **STOP-03**: Tool stops recording when file size reaches configured maximum (in MB)
- [x] **STOP-04**: Multiple stop conditions can be active simultaneously (first one triggered wins)

### Error Handling (ERR)

- [x] **ERR-01**: Tool fails early with clear message if ffmpeg is not found in PATH
- [x] **ERR-02**: Tool handles invalid RTSP URLs with descriptive error message
- [x] **ERR-03**: Tool ensures MP4 file is properly finalized even on unclean shutdown
- [x] **ERR-04**: Tool provides meaningful error messages for common ffmpeg failures

## v2 Requirements

### Recording Enhancements

- **REC-08**: Segmented recording — automatically split files by time or size intervals
- **REC-09**: Custom filename templates (support date/time placeholders)
- **REC-10**: Output directory configuration

### Advanced Features

- **ADV-01**: Concurrent multiple stream recording (architecture change)
- **ADV-02**: HTTP API for remote control
- **ADV-03**: Prometheus metrics export

## Out of Scope

| Feature | Reason |
|---------|--------|
| Motion detection | Scope explosion — full NVR territory, different product category |
| Native Go encoding without ffmpeg | Anti-pattern — ffmpeg's 20+ years of RTSP/MP4 handling is unmatched |
| Web interface | Violates CLI tool focus — would require additional architecture |
| Real-time stream health monitoring | Deferred to focus on recording reliability first |
| Custom output directories per recording | Current directory only for v1 simplicity |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CONF-01 | Phase 1 | Complete |
| CONF-02 | Phase 1 | Complete |
| CONF-03 | Phase 1 | Complete |
| CONF-04 | Phase 1 | Complete |
| REC-01 | Phase 2 | Complete |
| REC-02 | Phase 2 | Complete |
| REC-03 | Phase 2 | Complete |
| REC-04 | Phase 2 | Complete |
| REC-05 | Phase 2 | Complete |
| REC-06 | Phase 3 | Complete |
| REC-07 | Phase 1 | Complete |
| STOP-01 | Phase 2 | Complete |
| STOP-02 | Phase 2 | Complete |
| STOP-03 | Phase 2 | Complete |
| STOP-04 | Phase 2 | Complete |
| ERR-01 | Phase 1 | Complete |
| ERR-02 | Phase 3 | Complete |
| ERR-03 | Phase 2 | Complete |
| ERR-04 | Phase 3 | Complete |

**Coverage:**
- v1 requirements: 18 total
- Mapped to phases: 18
- Unmapped: 0 ✓

---
*Requirements defined: 2025-04-02*
*Last updated: 2025-04-02 after roadmap creation*
