# Architecture Patterns

**Domain:** Go CLI tool for RTSP stream recording  
**Researched:** 2026-04-02  
**Confidence:** HIGH

## Recommended Architecture

For the RTSP recorder CLI, we recommend a **layered architecture** following Go CLI best practices with clear separation between:

1. **CLI Interface Layer** (Cobra commands, flags, help)
2. **Configuration Layer** (Viper for YAML/env/flags)
3. **Application Logic Layer** (recording orchestration)
4. **External Integration Layer** (ffmpeg wrapper)

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Interface                          │
│  (cmd/root.go, cmd/record.go - Cobra commands)               │
└──────────────────────────┬──────────────────────────────────┘
                           │ uses
┌──────────────────────────▼──────────────────────────────────┐
│                   Configuration                             │
│  (config/config.go - Viper wrapper)                        │
└──────────────────────────┬──────────────────────────────────┘
                           │ uses
┌──────────────────────────▼──────────────────────────────────┐
│                 Recording Service                           │
│  (recorder/recorder.go - business logic)                   │
│  • Stop condition coordination                             │
│  • File naming & path management                             │
│  • Progress/status output                                    │
└──────────────────────────┬──────────────────────────────────┘
                           │ uses
┌──────────────────────────▼──────────────────────────────────┐
│                  FFmpeg Wrapper                             │
│  (ffmpeg/ffmpeg.go - external process management)            │
│  • Process lifecycle (start, monitor, stop)                  │
│  • Signal handling & graceful shutdown                      │
│  • Output capture & error handling                          │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

Following standard Go CLI patterns (Cobra generator conventions):

```
rtsp-recorder/
├── main.go              # Entry point: just calls cmd.Execute()
├── cmd/
│   ├── root.go          # Root command + global flags + Viper init
│   └── record.go        # Record subcommand (main functionality)
├── config/
│   └── config.go        # Configuration struct + Viper wrapper
├── recorder/
│   ├── recorder.go      # Recording orchestration logic
│   └── stop_conditions.go # Stop condition monitors (time, size, signals)
├── ffmpeg/
│   └── ffmpeg.go        # FFmpeg process wrapper
├── internal/
│   └── utils/           # Internal helpers (file naming, validation)
└── rtsp-recorder.yml    # Default config file
```

### Why this structure?

**Cobra conventions**: The `cmd/` package pattern is the standard Cobra structure. Each command gets its own file, commands are registered in `init()` functions, and `main.go` stays minimal.

**Separation of concerns**: 
- `config/` handles all configuration concerns (parsing, validation, defaults)
- `recorder/` contains the business logic independent of CLI or config details
- `ffmpeg/` isolates external process complexity
- `internal/` for utilities that shouldn't be exposed

## Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| **cmd/root.go** | CLI entry, flag parsing, help text | config (initializes), cmd/record |
| **cmd/record.go** | Record command implementation | config (reads), recorder (orchestrates) |
| **config/config.go** | Config loading, validation, defaults | Viper library only |
| **recorder/recorder.go** | Recording session lifecycle | config (reads), ffmpeg (controls), stop_conditions |
| **recorder/stop_conditions.go** | Signal, timer, file size monitors | recorder (signals), os/signal |
| **ffmpeg/ffmpeg.go** | FFmpeg process management | recorder (commands), os/exec |

### Boundary Rules

1. **cmd/ never imports os/exec directly** - all external process logic goes through ffmpeg/
2. **recorder/ never imports Viper** - it receives a config struct
3. **ffmpeg/ is the only package that imports os/exec**
4. **config/ is the only package that imports Viper**

## Data Flow

```
User Input
    │
    ▼
┌─────────────┐    ┌──────────────┐    ┌────────────────┐
│ CLI Flags   │───▶│ Viper Merge  │───▶│ Config Struct  │
│ (Cobra)     │    │ (env→file→flag)│   │ (validated)    │
└─────────────┘    └──────────────┘    └────────────────┘
                                                │
                                                ▼
                                        ┌────────────────┐
                                        │ Recorder       │
                                        │ (orchestrates) │
                                        └────────────────┘
                                                │
                    ┌──────────────────────────┼──────────────────────────┐
                    │                          │                          │
                    ▼                          ▼                          ▼
            ┌──────────────┐          ┌──────────────┐          ┌──────────────┐
            │ Signal       │          │ Timer        │          │ File Watcher │
            │ Monitor      │          │ (duration)   │          │ (max size)   │
            └──────────────┘          └──────────────┘          └──────────────┘
                    │                          │                          │
                    └──────────────────────────┼──────────────────────────┘
                                               │
                                               ▼ (any stop condition)
                                        ┌──────────────┐
                                        │ FFmpeg.Stop()│
                                        │ (SIGTERM →   │
                                        │  SIGKILL)    │
                                        └──────────────┘
```

### Configuration Precedence (Viper)

Following 12-Factor App patterns:

1. **Explicit `Set()` calls** (highest priority)
2. **Command-line flags** `--duration 5m`
3. **Environment variables** `RTSP_DURATION=5m`
4. **Config file** `rtsp-recorder.yml`
5. **Defaults** (lowest priority)

### Stop Condition Coordination

Multiple stop conditions run concurrently. First one triggers stops the recording:

- **Manual**: `os.Interrupt` (Ctrl+C) → SIGTERM to ffmpeg
- **Duration**: Timer → SIGTERM after elapsed
- **File size**: Periodic stat checks → SIGTERM when exceeded

All conditions share a `context.Context` for cancellation propagation.

## Patterns to Follow

### Pattern 1: Cobra Command Structure
**What:** Each command in its own file with `init()` registration  
**When:** All CLI commands  
**Example:**
```go
// cmd/record.go
var recordCmd = &cobra.Command{
    Use:   "record [RTSP_URL]",
    Short: "Record an RTSP stream to MP4",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg := config.FromViper()
        rec := recorder.New(cfg)
        return rec.Record(args[0])
    },
}

func init() {
    rootCmd.AddCommand(recordCmd)
    recordCmd.Flags().Duration("duration", 0, "Max recording duration")
    viper.BindPFlag("duration", recordCmd.Flags().Lookup("duration"))
}
```

### Pattern 2: Context-Based Cancellation
**What:** Pass `context.Context` for cancellation propagation  
**When:** Long-running operations, goroutine coordination  
**Example:**
```go
// recorder/recorder.go
func (r *Recorder) Record(ctx context.Context, url string) error {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    // Signal handler
    sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt)
    defer stop()
    
    // Start ffmpeg
    ffmpegCmd := r.ffmpeg.Start(url, r.outputPath)
    
    // Wait for stop condition
    select {
    case <-sigCtx.Done():
        log.Println("Interrupted by signal")
    case <-r.durationTimer.C:
        log.Println("Duration limit reached")
    case <-r.fileSizeWatcher.Done():
        log.Println("File size limit reached")
    }
    
    return r.ffmpeg.Stop()
}
```

### Pattern 3: FFmpeg Process Wrapper
**What:** Encapsulate ffmpeg lifecycle with proper signal handling  
**When:** Managing external processes  
**Example:**
```go
// ffmpeg/ffmpeg.go
type Cmd struct {
    cmd    *exec.Cmd
    cancel context.CancelFunc
}

func (c *Cmd) Start(ctx context.Context, args []string) error {
    ctx, cancel := context.WithCancel(ctx)
    c.cancel = cancel
    
    c.cmd = exec.CommandContext(ctx, "ffmpeg", args...)
    c.cmd.Stdout = os.Stdout
    c.cmd.Stderr = os.Stderr
    
    return c.cmd.Start()
}

func (c *Cmd) Stop() error {
    // Try graceful shutdown first
    if c.cmd.Process != nil {
        c.cmd.Process.Signal(syscall.SIGTERM)
        
        // Wait with timeout
        done := make(chan error, 1)
        go func() { done <- c.cmd.Wait() }()
        
        select {
        case err := <-done:
            return err
        case <-time.After(5 * time.Second):
            c.cmd.Process.Kill()
            return c.cmd.Wait()
        }
    }
    return nil
}
```

### Pattern 4: Config Struct with Viper
**What:** Viper for loading, typed struct for runtime use  
**When:** All configuration needs  
**Example:**
```go
// config/config.go
type Config struct {
    Duration    time.Duration `mapstructure:"duration"`
    MaxFileSize int64         `mapstructure:"max_file_size"`
    OutputDir   string        `mapstructure:"output_dir"`
    RTSPTimeout time.Duration `mapstructure:"rtsp_timeout"`
}

func Load() (*Config, error) {
    viper.SetConfigName("rtsp-recorder")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("$HOME/.config/rtsp-recorder")
    
    viper.SetDefault("duration", 0)
    viper.SetDefault("output_dir", ".")
    viper.SetDefault("rtsp_timeout", "30s")
    
    viper.SetEnvPrefix("RTSP")
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        // Config file optional - only fail if explicitly set
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }
    
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Global Viper Usage
**What:** Using `viper.Get*` directly throughout the codebase  
**Why bad:** Hard to test, couples code to global state, configuration source unclear  
**Instead:** Load config once, pass struct to components

### Anti-Pattern 2: Unchecked ffmpeg Process Leaks
**What:** Starting ffmpeg without guaranteed cleanup  
**Why bad:** Zombie processes, resource leaks, partial files  
**Instead:** Use `defer`, context cancellation, and `Process.Kill()` fallback

### Anti-Pattern 3: Synchronous Signal Handling
**What:** Blocking on signals in the main flow  
**Why bad:** Can't handle multiple stop conditions concurrently  
**Instead:** Use `signal.NotifyContext` + `select` with multiple channels

### Anti-Pattern 4: String Typing for Paths
**What:** Using plain strings for file paths  
**Why bad:** Platform incompatibilities, path traversal issues  
**Instead:** Use `filepath` package, validate paths, sanitize inputs

## Scalability Considerations

| Concern | v1 (Single Stream) | Future (Multi-Stream) |
|---------|-------------------|----------------------|
| **Concurrency** | Single goroutine + signal monitor | Goroutine per stream + sync.WaitGroup |
| **Process mgmt** | One ffmpeg process | Process pool, resource limits |
| **Configuration** | Single config file | Per-stream config + global defaults |
| **Output** | Timestamped files | Organized directory structure |
| **Monitoring** | Console output | Structured logging, metrics |

## Build Order Implications

Based on component dependencies:

1. **Phase 1: Core Infrastructure**
   - `config/config.go` - Configuration loading
   - `internal/utils/file.go` - File naming utilities
   
2. **Phase 2: External Integration**
   - `ffmpeg/ffmpeg.go` - FFmpeg wrapper (depends on config for validation)
   
3. **Phase 3: Business Logic**
   - `recorder/stop_conditions.go` - Stop condition monitors
   - `recorder/recorder.go` - Recording orchestration
   
4. **Phase 4: CLI Interface**
   - `cmd/root.go` - Root command + config init
   - `cmd/record.go` - Record command
   - `main.go` - Entry point

**Dependency rationale:**
- Config has no internal dependencies → build first
- FFmpeg depends on config for validation → build second  
- Recorder depends on both → build third
- CLI depends on everything → build last

## File Naming Convention

For automatic organization without user decisions:

```go
// internal/utils/file.go
func GenerateFilename(timestamp time.Time, streamID string) string {
    // Format: rtsp_YYYY-MM-DD_HH-MM-SS_[streamID].mp4
    base := timestamp.Format("2006-01-02_15-04-05")
    if streamID != "" {
        return fmt.Sprintf("rtsp_%s_%s.mp4", base, streamID)
    }
    return fmt.Sprintf("rtsp_%s.mp4", base)
}
```

**Why:** ISO 8601-ish format sorts chronologically, stream ID prevents collisions if same stream recorded multiple times.

## Sources

- **HIGH confidence:**
  - Cobra user guide: https://github.com/spf13/cobra/blob/main/site/content/user_guide.md
  - Viper README: https://github.com/spf13/viper/blob/master/README.md
  - Go os/exec documentation: https://pkg.go.dev/os/exec
  - Go os/signal documentation: https://pkg.go.dev/os/signal
  - Go context documentation: https://pkg.go.dev/context
  - Effective Go: https://go.dev/doc/effective_go.html

- **Pattern conventions:**
  - Standard Go CLI patterns established by Hugo, kubectl, Docker CLI
  - 12-Factor App configuration principles
