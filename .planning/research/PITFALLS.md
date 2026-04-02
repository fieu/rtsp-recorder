# Domain Pitfalls: RTSP Recorder CLI

**Domain:** Video streaming/RTSP recording with Go CLI
**Researched:** 2026-04-02
**Overall confidence:** HIGH (based on official Go docs, ffmpeg docs, and verified patterns)

## Critical Pitfalls

### Pitfall 1: Zombie FFmpeg Processes on Interrupt

**What goes wrong:** When the user presses Ctrl+C, the Go process exits but leaves ffmpeg running as a zombie process, consuming system resources and potentially corrupting the output file.

**Why it happens:** 
- Go's default signal handling doesn't propagate signals to child processes
- Using `cmd.Run()` blocks until completion, preventing signal interception
- Not calling `cmd.Wait()` after `cmd.Start()` leaves process state unreleased

**Consequences:**
- Orphaned ffmpeg processes consuming CPU/memory
- Corrupted/incomplete MP4 files (moov atom not written)
- Port conflicts on subsequent runs

**Prevention:**
```go
// Use signal.NotifyContext for graceful shutdown
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

// Use CommandContext for automatic cancellation
cmd := exec.CommandContext(ctx, "ffmpeg", args...)

// Or implement custom Cancel function for graceful ffmpeg shutdown
cmd := exec.Command("ffmpeg", args...)
cmd.Cancel = func() error {
    return cmd.Process.Signal(syscall.SIGINT) // Send SIGINT first for clean shutdown
}
cmd.WaitDelay = 5 * time.Second // Give ffmpeg time to finalize MP4
```

**Warning signs:** 
- Process list shows multiple ffmpeg instances after stopping CLI
- Output MP4 files are unplayable or show incorrect duration

**Phase to address:** Phase 1 (Core Recording)

---

### Pitfall 2: MP4 File Corruption on Unclean Shutdown

**What goes wrong:** MP4 files require the moov atom (metadata) to be written at the end of recording. If ffmpeg is killed abruptly (SIGKILL), the file is unplayable.

**Why it happens:**
- MP4 container format stores metadata at the end
- FFmpeg needs time to "finalize" the file on shutdown
- Using `-movflags +faststart` helps but doesn't eliminate the need for clean shutdown

**Consequences:**
- Unplayable output files
- Lost recording data
- User perceives the tool as unreliable

**Prevention:**
```go
// Always use SIGINT/SIGTERM first, wait, then SIGKILL if needed
cmd := exec.Command("ffmpeg", 
    "-i", rtspURL,
    "-c", "copy",           // Stream copy (no re-encode) for lower CPU
    "-movflags", "+faststart", // Move moov to start for web playback
    "-y", outputPath,
)

// Implement graceful shutdown with timeout
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

go func() {
    <-quit
    if cmd.Process != nil {
        cmd.Process.Signal(syscall.SIGINT)
        // Wait up to 5 seconds for clean exit
        time.AfterFunc(5*time.Second, func() {
            cmd.Process.Kill()
        })
    }
}()
```

**Warning signs:**
- Files play in VLC but not in QuickTime/Windows Media Player
- FFmpeg reports "moov atom not found" when inspecting files

**Phase to address:** Phase 1 (Core Recording)

---

### Pitfall 3: RTSP Connection Timeout Not Handled

**What goes wrong:** RTSP streams can hang during connection establishment or mid-stream. Without timeout handling, the recorder hangs indefinitely.

**Why it happens:**
- Default ffmpeg RTSP timeout is -1 (infinite) for some protocols
- Network issues, camera reboots, or stream unavailability cause hangs
- Go's `exec.Command` doesn't timeout by default

**Consequences:**
- Recorder appears to hang with no feedback
- No way to recover without killing the process
- Missed recording windows

**Prevention:**
```go
// Set explicit timeouts in ffmpeg args
args := []string{
    "-rtsp_transport", "tcp",        // More reliable than UDP
    "-stimeout", "5000000",          // Socket timeout: 5 seconds (microseconds)
    "-reconnect", "1",               // Auto-reconnect on disconnect
    "-reconnect_at_eof", "1",
    "-reconnect_streamed", "1",
    "-reconnect_delay_max", "5",
    "-i", rtspURL,
    // ... rest of args
}

// Also use Go context for overall timeout
ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
defer cancel()

cmd := exec.CommandContext(ctx, "ffmpeg", args...)
```

**Warning signs:**
- Recorder hangs with no output
- Process shows as sleeping in `ps` indefinitely

**Phase to address:** Phase 1 (Core Recording)

---

### Pitfall 4: Signal.Notify with Unbuffered Channel

**What goes wrong:** Using an unbuffered channel with `signal.Notify` can cause missed signals during high load or rapid signal delivery.

**Why it happens:**
- Signal delivery is not blocking; if the channel is full, signals are dropped
- Go docs explicitly warn: "the caller must ensure that c has sufficient buffer space"

**Consequences:**
- Missed Ctrl+C signals
- Unresponsive CLI during shutdown
- Process requires SIGKILL to terminate

**Prevention:**
```go
// WRONG - can miss signals
c := make(chan os.Signal) // Unbuffered!
signal.Notify(c, os.Interrupt)

// CORRECT - buffer of 1 is sufficient for single signals
c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)

// Even better: use NotifyContext (Go 1.16+)
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
```

**Warning signs:**
- Intermittent failure to respond to Ctrl+C
- Tests for signal handling pass sometimes, fail others

**Phase to address:** Phase 1 (Core Recording)

---

### Pitfall 5: Viper Configuration Case Sensitivity Issues

**What goes wrong:** Configuration values don't load from environment variables or files as expected due to Viper's case-insensitive key handling.

**Why it happens:**
- Viper lowercases all keys internally (documented behavior)
- Environment variables are typically UPPERCASE
- Struct tags need to match Viper's expectations

**Consequences:**
- Config values ignored silently
- Environment variable overrides don't work
- Difficult debugging

**Prevention:**
```go
// Use consistent lowercase keys in Viper
viper.SetDefault("rtsp_url", "")
viper.SetDefault("output_dir", "./recordings")
viper.SetDefault("max_duration", "1h")

// For environment variables, use SetEnvPrefix and bind explicitly
viper.SetEnvPrefix("RTSP_RECORDER") // Will look for RTSP_RECORDER_RTSP_URL
viper.BindEnv("rtsp_url") // Map to RTSP_RECORDER_RTSP_URL

// Or use AutomaticEnv with key replacer for snake_case
viper.AutomaticEnv()
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
// Now RTSP_RECORDER.RTSP_URL becomes RTSP_RECORDER_RTSP_URL

// For unmarshaling, use mapstructure tags
type Config struct {
    RTSPURL      string        `mapstructure:"rtsp_url"`
    OutputDir    string        `mapstructure:"output_dir"`
    MaxDuration  time.Duration `mapstructure:"max_duration"`
}
```

**Warning signs:**
- `viper.Get()` returns zero values
- Config file loads but values don't apply

**Phase to address:** Phase 1 (CLI & Config Setup)

---

### Pitfall 6: Not Checking FFmpeg Exit Error Properly

**What goes wrong:** FFmpeg can fail for many reasons (bad URL, codec issues, network errors) but the Go code doesn't distinguish between exit errors and other failures.

**Why it happens:**
- `cmd.Run()` returns `*exec.ExitError` for non-zero exits
- Exit code 1 from ffmpeg is common and not always an error for RTSP (e.g., timeout reached)
- Stderr contains the actual error details

**Consequences:**
- Users get unhelpful "exit status 1" errors
- No way to distinguish "stream ended" from "connection failed"
- Silent failures in automated deployments

**Prevention:**
```go
var stderr bytes.Buffer
cmd.Stderr = &stderr

if err := cmd.Run(); err != nil {
    if exitErr, ok := err.(*exec.ExitError); ok {
        // Check stderr for ffmpeg-specific error details
        stderrStr := stderr.String()
        
        // Common patterns to detect
        if strings.Contains(stderrStr, "Connection refused") {
            return fmt.Errorf("RTSP connection refused: %w", err)
        }
        if strings.Contains(stderrStr, "404 Not Found") {
            return fmt.Errorf("RTSP stream not found: %w", err)
        }
        if strings.Contains(stderrStr, "Invalid data") {
            return fmt.Errorf("stream format not supported: %w", err)
        }
        
        // Check exit code
        if exitErr.ExitCode() == 255 {
            // Often indicates network/connection issues
            return fmt.Errorf("ffmpeg network error (exit 255): %s", stderrStr)
        }
    }
    return fmt.Errorf("ffmpeg failed: %w\nstderr: %s", err, stderr.String())
}
```

**Warning signs:**
- Generic "exit status 1" errors
- No actionable error messages for users

**Phase to address:** Phase 1 (Core Recording)

---

### Pitfall 7: ProcessGroup Leak on Unix Systems

**What goes wrong:** On Linux/macOS, ffmpeg may spawn child processes. Killing the parent Go process doesn't kill the entire process group, leaving orphaned ffmpeg processes.

**Why it happens:**
- `cmd.Process.Kill()` only kills the immediate process
- FFmpeg can spawn threads or sub-processes
- Signal doesn't propagate to process group by default

**Consequences:**
- Accumulating zombie processes
- Resource exhaustion over time
- Port binding conflicts

**Prevention:**
```go
import "syscall"

// Create process group on start
cmd := exec.Command("ffmpeg", args...)
cmd.SysProcAttr = &syscall.SysProcAttr{
    Setpgid: true, // Create new process group
}

// Kill entire process group on cleanup
func cleanup(cmd *exec.Cmd) {
    if cmd.Process != nil {
        // Negative PID sends signal to entire process group
        syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
        
        // Give it time to exit gracefully
        time.Sleep(2 * time.Second)
        
        // Force kill if still running
        syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
    }
}
```

**Warning signs:**
- `ps aux | grep ffmpeg` shows processes not associated with any parent
- System load increases over time

**Phase to address:** Phase 1 (Core Recording)

---

### Pitfall 8: Using -t (duration) Without Accounting for FFmpeg Startup Time

**What goes wrong:** Using ffmpeg's `-t` flag for duration limiting measures from when ffmpeg starts, not when the stream actually begins recording.

**Why it happens:**
- `-t 60` stops after 60 seconds of ffmpeg runtime
- RTSP connection and stream setup can take 5-30 seconds
- Actual recorded content is less than requested duration

**Consequences:**
- Shorter recordings than specified
- Inaccurate duration limits
- User confusion

**Prevention:**
```go
// Don't use ffmpeg's -t for precise duration limiting
// Instead, manage duration in Go and send interrupt at the right time

duration := viper.GetDuration("duration")
if duration > 0 {
    ctx, cancel := context.WithTimeout(context.Background(), duration)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, "ffmpeg", args...)
    // Context cancellation will terminate ffmpeg after exact duration
}

// Alternative: use -t with a buffer for connection time
// But this is less accurate than Go context
```

**Warning signs:**
- 60-second duration setting results in 45-second recordings
- Inconsistent recording lengths

**Phase to address:** Phase 1 (Core Recording - Stop Conditions)

---

### Pitfall 9: Not Validating FFmpeg Availability Early

**What goes wrong:** The CLI accepts all arguments and tries to start recording, only to fail with "executable not found" after all setup is done.

**Why it happens:**
- No early check for ffmpeg in PATH
- Wasted time on config parsing and validation
- Poor user experience

**Consequences:**
- Delayed failure feedback
- User confusion about requirements
- Potentially misleading error messages

**Prevention:**
```go
// Check ffmpeg availability at startup
func init() {
    if _, err := exec.LookPath("ffmpeg"); err != nil {
        fmt.Fprintln(os.Stderr, "Error: ffmpeg is required but not found in PATH")
        fmt.Fprintln(os.Stderr, "Please install ffmpeg: https://ffmpeg.org/download.html")
        os.Exit(1)
    }
    
    // Optional: verify ffmpeg version meets minimum requirements
    cmd := exec.Command("ffmpeg", "-version")
    output, _ := cmd.Output()
    version := parseVersion(string(output))
    if version.Before(minRequiredVersion) {
        fmt.Fprintf(os.Stderr, "Warning: ffmpeg version %s may be too old. Recommended: %s+\n", 
            version, minRequiredVersion)
    }
}
```

**Warning signs:**
- "executable file not found in $PATH" errors after argument parsing
- Users unaware of ffmpeg dependency

**Phase to address:** Phase 1 (CLI Setup)

---

### Pitfall 10: Cobra/Viper Flag Binding Order Issues

**What goes wrong:** Flags bound to Viper don't reflect the values users set on the command line because binding happens in the wrong order.

**Why it happens:**
- Viper binding must happen after flags are parsed
- `init()` runs before command execution
- Flag values aren't set when `viper.BindPFlag()` is called

**Consequences:**
- CLI flags are ignored
- Config file values override command-line arguments
- User confusion about precedence

**Prevention:**
```go
// WRONG - in init()
func init() {
    rootCmd.Flags().StringP("url", "u", "", "RTSP URL")
    viper.BindPFlag("url", rootCmd.Flags().Lookup("url")) // Value not set yet!
}

// CORRECT - use PersistentPreRun or PreRun
var rootCmd = &cobra.Command{
    Use:   "rtsp-recorder",
    Short: "Record RTSP streams",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Bind flags after they're parsed
        viper.BindPFlag("url", cmd.Flags().Lookup("url"))
        viper.BindPFlag("duration", cmd.Flags().Lookup("duration"))
        viper.BindPFlag("output", cmd.Flags().Lookup("output"))
        return nil
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        // Now viper.GetString("url") will have the correct value
        url := viper.GetString("url")
        // ...
    },
}
```

**Warning signs:**
- `--url` flag is passed but config file value is used
- Flags show correct values in help but don't work

**Phase to address:** Phase 1 (CLI Setup)

---

## Moderate Pitfalls

### Pitfall 11: UDP vs TCP Transport Selection

**What goes wrong:** Using UDP (default for RTSP) causes packet loss and corruption on lossy networks.

**Why it happens:**
- FFmpeg defaults to UDP for RTSP
- UDP doesn't guarantee packet delivery
- Lost packets = video artifacts

**Prevention:**
- Always use `-rtsp_transport tcp` for reliable recording
- Only use UDP if latency is critical and quality is secondary

**Phase to address:** Phase 1 (Core Recording)

---

### Pitfall 12: Filename Timestamp Collisions

**What goes wrong:** Rapid successive recordings or concurrent runs can produce the same filename, causing overwrites.

**Why it happens:**
- Timestamp granularity is seconds
- Multiple recordings in same second overwrite each other

**Prevention:**
```go
// Use nanosecond precision or add random component
timestamp := time.Now().Format("20060102_150405") // seconds
// Better: timestamp := time.Now().Format("20060102_150405.000000000") // nanoseconds
// Or: add counter/nanosecond to filename
```

**Phase to address:** Phase 1 (File Naming)

---

### Pitfall 13: No Progress/Status Feedback

**What goes wrong:** Long recordings (>1 hour) provide no feedback, leaving users unsure if recording is working.

**Why it happens:**
- No output parsing from ffmpeg
- Silent operation during long runs

**Prevention:**
```go
// Parse ffmpeg stderr for progress updates
// Add -progress pipe:1 to output progress to stdout
// Or use file-based progress and poll
```

**Phase to address:** Phase 2 (Nice-to-haves)

---

### Pitfall 14: Config File Not Found Handling

**What goes wrong:** Missing config file causes panic or confusing error instead of graceful fallback to defaults.

**Prevention:**
```go
if err := viper.ReadInConfig(); err != nil {
    var fileNotFound viper.ConfigFileNotFoundError
    if !errors.As(err, &fileNotFound) {
        // Config file was found but another error occurred
        return fmt.Errorf("config error: %w", err)
    }
    // Config file not found; use defaults
    log.Println("No config file found, using defaults")
}
```

**Phase to address:** Phase 1 (CLI Setup)

---

## Minor Pitfalls

### Pitfall 15: Not Handling Concurrent Config Access

**What goes wrong:** Using global Viper instance from multiple goroutines causes race conditions.

**Prevention:**
```go
// Create instance per goroutine or use sync.Mutex
v := viper.New()
// Use v instead of global viper package
```

**Phase to address:** Phase 3 (Concurrent Streams)

---

### Pitfall 16: FFmpeg Output Overwriting

**What goes wrong:** Using `-y` flag blindly overwrites existing files without warning.

**Prevention:**
- Check for file existence before starting
- Or use `-n` to fail on existing files
- Make overwrite behavior configurable

**Phase to address:** Phase 1 (File Handling)

---

## Phase-Specific Warnings

| Phase | Topic | Likely Pitfall | Mitigation |
|-------|-------|----------------|------------|
| Phase 1 | Signal handling | Zombie processes | Use signal.NotifyContext + proper WaitDelay |
| Phase 1 | Process management | Process group leaks | Setpgid + kill negative PID |
| Phase 1 | Config binding | Flag precedence issues | Bind in PersistentPreRun |
| Phase 1 | Error handling | Generic exit errors | Parse stderr for specific error patterns |
| Phase 1 | FFmpeg args | UDP packet loss | Always use `-rtsp_transport tcp` |
| Phase 1 | Duration limits | FFmpeg -t inaccuracy | Use Go context for precise timing |
| Phase 3 | Concurrent streams | Resource exhaustion | Rate limiting + process quotas |
| Phase 3 | Multiple configs | Viper race conditions | Use viper.New() instances |

---

## Confidence Assessment

| Pitfall Area | Confidence | Notes |
|--------------|------------|-------|
| Go signal handling | HIGH | Official Go docs, widely documented patterns |
| Process management | HIGH | Standard practice in Go community |
| FFmpeg integration | HIGH | FFmpeg docs + common Go patterns |
| Viper pitfalls | HIGH | From official troubleshooting guide |
| RTSP specifics | MEDIUM | Verified against FFmpeg RTSP docs |
| Cobra patterns | HIGH | Official user guide + common usage |

## Sources

1. **Go os/exec docs**: https://pkg.go.dev/os/exec - HIGH confidence
2. **Go os/signal docs**: https://pkg.go.dev/os/signal - HIGH confidence
3. **Cobra user guide**: https://github.com/spf13/cobra/blob/main/site/content/user_guide.md - HIGH confidence
4. **Viper README**: https://github.com/spf13/viper - HIGH confidence
5. **Viper Troubleshooting**: https://github.com/spf13/viper/blob/master/TROUBLESHOOTING.md - HIGH confidence
6. **FFmpeg Protocols**: https://ffmpeg.org/ffmpeg-protocols.html - HIGH confidence
7. **Go exec tests**: Go source code for patterns - HIGH confidence
