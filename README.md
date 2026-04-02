# 📹 rtsp-recorder

[![Go Version](https://img.shields.io/badge/go-1.26-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

> A lightweight CLI tool for recording RTSP video streams to MP4 files with style ✨

## 🚀 Features

- 📹 **RTSP Recording** — Capture any RTSP stream to MP4 format
- ⚡ **Timelapse Mode** — Condense hours of footage into seconds (e.g., 1 hour → 10 seconds)
- 🎛️ **Flexible Stop Conditions** — Stop by duration, file size, or manual interrupt (Ctrl+C)
- 📊 **Real-time Progress** — Live display of recording stats
- 🔄 **Auto-retry** — Automatically reconnects on network errors
- ⚙️ **Configuration** — YAML config, CLI flags, or environment variables
- 🪶 **Lightweight** — Minimal CPU usage with stream copy (no re-encoding for normal recordings)

## 📦 Installation

### From Source

```bash
git clone https://github.com/fieu/rtsp-recorder.git
cd rtsp-recorder
go build -o rtsp-recorder .
```

### Prerequisites

- [Go](https://golang.org/dl/) 1.26 or later
- [FFmpeg](https://ffmpeg.org/download.html) 7.1+ or 8.x

## 🎮 Quick Start

### Basic Recording

```bash
# Record for 30 minutes
./rtsp-recorder record --duration 30m rtsp://camera.local/stream

# Output: recording_2025-04-02-14-30-00.mp4
```

### Timelapse Recording ⏱️

```bash
# Record 1 hour, output 10-second timelapse at 360x speed
./rtsp-recorder record --duration 1h --timelapse 10s rtsp://camera.local/stream

[INFO] Timelapse: 360x speed (1h -> 10s)
Recording: 30m elapsed | Output: ~5s | 360x
```

### With Configuration File

Create `rtsp-recorder.yml`:

```yaml
url: rtsp://192.168.1.100:554/stream
duration: 1h
max_file_size: 1024
retry_attempts: 3
```

Then simply run:

```bash
./rtsp-recorder record
```

## 📋 Usage

```
rtsp-recorder record [RTSP_URL] [flags]

Flags:
  -u, --url string              RTSP stream URL
  -d, --duration duration       Recording duration (e.g., 30m, 1h, 2h30m)
  -s, --max-file-size int       Maximum file size in MB before stopping
  -l, --timelapse duration      Target output duration for timelapse (e.g., 10s)
  -r, --retry-attempts int      Number of retry attempts on connection failure (default 3)
  -f, --ffmpeg-path string      Path to ffmpeg binary (optional)
  -t, --filename-template string Output filename template (optional)
  -h, --help                    Help for record

Examples:
  # Basic recording
  rtsp-recorder record rtsp://camera.local/stream

  # With duration and file size limits
  rtsp-recorder record --duration 1h --max-file-size 500 rtsp://192.168.1.100:554/stream

  # Timelapse (1 hour → 10 seconds)
  rtsp-recorder record --duration 1h --timelapse 10s rtsp://camera.local/stream

  # Short flags
  rtsp-recorder record -d 30m -s 256 rtsp://camera.local/stream
```

## ⚙️ Configuration

Configuration precedence (highest to lowest):
1. CLI flags
2. Environment variables (`RTSP_RECORDER_*`)
3. Config file (`rtsp-recorder.yml`)
4. Defaults

### Example Config File

```yaml
# rtsp-recorder.yml
url: rtsp://192.168.1.100:554/stream
duration: 30m
max_file_size: 1024
retry_attempts: 3
```

### Environment Variables

```bash
export RTSP_RECORDER_URL=rtsp://camera.local/stream
export RTSP_RECORDER_DURATION=1h
export RTSP_RECORDER_RETRY_ATTEMPTS=5
```

## 🛠️ Other Commands

### Validate Setup

```bash
./rtsp-recorder validate

[INFO] FFmpeg found: /usr/bin/ffmpeg (version 8.1)
[INFO] Configuration valid ✓
```

## 🎯 Use Cases

- 🏠 **Home Security** — Record IP camera footage with scheduled duration
- 🌅 **Time-lapse Videos** — Create accelerated videos of construction, weather, or nature
- 📹 **Stream Archiving** — Backup live streams for later review
- 🔬 **Monitoring** — Continuous recording with automatic rotation

## 🏗️ Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│   CLI (Cobra)   │────▶│   Recorder   │────▶│   FFmpeg    │
└─────────────────┘     └──────────────┘     └─────────────┘
        │
        ▼
┌─────────────────┐
│  Config (Viper) │
└─────────────────┘
```

## 🧪 Testing

```bash
go test ./...
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) 🐍
- Configuration powered by [Viper](https://github.com/spf13/viper) 🐍
- Video processing by [FFmpeg](https://ffmpeg.org/) 🎬

---

Made with ❤️ for the RTSP recording community
