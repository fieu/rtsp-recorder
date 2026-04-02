# Phase 4: Timelapse Recording - Research

**Researched:** 2025-04-02
**Status:** Ready for planning

---

## Executive Summary

Timelapse recording uses FFmpeg's `select` and `setpts` filters to drop frames in real-time during recording, producing accelerated output directly. This approach is more efficient than post-processing because it avoids writing then re-reading the full video.

## FFmpeg Timelapse Techniques

### Method 1: Frame Selection with `select` Filter (RECOMMENDED)

The `select` filter evaluates an expression for each frame and keeps only frames where the expression is non-zero.

**Basic syntax:**
```bash
ffmpeg -i input.mp4 -vf "select='not(mod(n,10))',setpts=N/(FRAME_RATE*TB)" output.mp4
```

**Key components:**
- `select='not(mod(n,X))'` — Keep every Xth frame (drop X-1 of every X frames)
- `setpts=N/(FRAME_RATE*TB)` — Adjust presentation timestamps to maintain smooth playback
  - `N` = frame number in output
  - `FRAME_RATE` = original frame rate
  - `TB` = timebase (1/frame_rate)

**Example calculations:**
| Input Duration | Target Output | Speedup | Frame Interval (X) |
|----------------|---------------|---------|-------------------|
| 1 hour (3600s) | 10 seconds | 360x | Every 360th frame |
| 30 minutes | 5 seconds | 360x | Every 360th frame |
| 1 hour | 1 minute | 60x | Every 60th frame |

### Method 2: Speed Filter (Alternative)

The `setpts` filter can also be used directly with a speed factor:

```bash
ffmpeg -i input.mp4 -vf "setpts=PTS/360" output.mp4
```

**Tradeoffs:**
- Simpler syntax
- Works with stream copy mode (`-c:v copy`) when combined with `-r` output frame rate
- Less precise control over output duration

## Audio Handling Options

### Option A: Drop Audio (Recommended for Timelapse)

Most timelapse videos don't need audio, and keeping audio requires pitch correction or sounds strange when sped up.

```bash
ffmpeg -i rtsp://... -vf "select='...',setpts=..." -an output.mp4
```

**Pros:**
- Simpler implementation
- Smaller output files
- No audio quality concerns
- Matches user expectations for timelapse

### Option B: Keep Audio with Pitch Correction

If audio is needed, use the `atempo` filter (limited to 0.5-2.0x range, chain multiple for higher speeds):

```bash
ffmpeg -i input.mp4 -vf "select='...',setpts=..." -af "atempo=2.0,atempo=2.0,atempo=2.0" output.mp4
```

**Cons:**
- Complex chaining for high speedups (>8x requires multiple atempo filters)
- Audio quality degrades at extreme speeds
- Not commonly expected in timelapse videos

## Integration with RTSP Streams

### Complete FFmpeg Command for RTSP Timelapse

```bash
ffmpeg \
  -rtsp_transport tcp \
  -timeout 5000000 \
  -fflags +discardcorrupt+genpts \
  -use_wallclock_as_timestamps 1 \
  -i rtsp://camera.local/stream \
  -vf "select='not(mod(n,360))',setpts=N/(FRAME_RATE*TB)" \
  -c:v copy \
  -an \
  -f mp4 \
  -movflags +faststart \
  -y \
  output.mp4
```

**Important considerations:**

1. **Stream copy compatibility:** `-c:v copy` with video filters requires careful handling. When using filters, ffmpeg may need to re-encode. For timelapse, re-encoding is acceptable since we're already processing frames.

2. **Frame rate calculation:** The speedup factor depends on knowing the input frame rate. RTSP streams may have variable frame rates, so we should:
   - Use a default assumption (e.g., 30fps)
   - Or detect from stream metadata
   - Or calculate based on elapsed time rather than frame count

3. **Real-time vs post-processing:** The `--timelapse` flag approach records for a duration, dropping frames in real-time. This differs from:
   - Recording full-speed then processing (post-processing)
   - True "slow motion capture" where camera captures at intervals

## Implementation Patterns

### Frame Interval Calculation

```go
// Calculate frame interval based on durations
func CalculateFrameInterval(recordDuration, timelapseDuration time.Duration, frameRate float64) int {
    if timelapseDuration <= 0 || recordDuration <= 0 {
        return 1 // No timelapse, keep all frames
    }
    
    // Speedup factor
    speedup := float64(recordDuration) / float64(timelapseDuration)
    
    // Frames to keep: every Nth frame
    interval := int(speedup)
    if interval < 1 {
        interval = 1
    }
    
    return interval
}
```

### FFmpeg Filter String Generation

```go
func (c *Cmd) buildTimelapseFilter(interval int) string {
    // select='not(mod(n,X))' - keep every Xth frame
    // setpts=N/(FRAME_RATE*TB) - adjust timestamps
    return fmt.Sprintf("select='not(mod(n,%d))',setpts=N/(FRAME_RATE*TB)", interval)
}
```

### Config Integration

Add to `config.Config`:
```go
// TimelapseDuration is the target output duration (0 = no timelapse)
TimelapseDuration time.Duration `mapstructure:"timelapse_duration"`
```

Add flag in `cmd/record.go`:
```go
cmd.Flags().DurationP("timelapse", "tl", 0, "Target output duration for timelapse (e.g., 10s, 1m)")
viper.BindPFlag("timelapse_duration", cmd.Flags().Lookup("timelapse"))
```

## Error Handling Considerations

### Validation Requirements

1. **Timelapse without duration:** Error if `--timelapse` provided without `--duration`
   - Reason: Cannot calculate speedup without knowing recording duration
   - Message: "--timelapse requires --duration: cannot calculate speedup without recording duration"

2. **Minimum timelapse duration:** Reject values less than 1 second
   - Reason: Would produce essentially no output
   - Message: "--timelapse must be at least 1s"

3. **Maximum speedup sanity check:** Warn on extreme speedups (>10000x)
   - Reason: May produce choppy output
   - Not an error, just a warning

### Edge Cases

| Scenario | Handling |
|----------|----------|
| Recording stops early (Ctrl+C) | Output contains frames captured up to that point, plays at calculated speed |
| Network interruption during timelapse | Retry logic applies same as normal recording |
| Very short recordings (< timelapse target) | Output shows all captured frames, may be shorter than target |
| Variable frame rate streams | Use elapsed time for calculation, not frame count |

## Performance Considerations

### CPU Impact

- **Without timelapse:** `-c:v copy` uses minimal CPU
- **With timelapse:** Frame selection still uses copy mode, but filter chain adds overhead
- **Expected overhead:** Low - single frame evaluation per input frame

### Memory Impact

- Negligible increase - filters operate frame-by-frame
- No buffering of large frame sequences required

### File Size

- Timelapse output is proportionally smaller (by speedup factor)
- Example: 1 hour recording at 360x speedup = ~10 seconds output

## Testing Strategy

### Unit Tests

1. **Frame interval calculation:** Test various duration combinations
2. **Filter string generation:** Verify correct FFmpeg syntax
3. **Validation logic:** Test error cases

### Integration Tests

1. **Short test recording:** 10-second recording with 1-second timelapse target
2. **Verify output duration:** Check actual output matches expected (within tolerance)
3. **Verify playback:** Ensure output plays without errors

## References

- FFmpeg `select` filter: https://ffmpeg.org/ffmpeg-filters.html#select_002c-aselect
- FFmpeg `setpts` filter: https://ffmpeg.org/ffmpeg-filters.html#setpts_002c-asetpts
- FFmpeg mod function: https://ffmpeg.org/ffmpeg-utils.html#Expression-Evaluation

---

*Research completed: 2025-04-02*
