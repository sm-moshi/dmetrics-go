# Power Metrics

The power metrics module provides system power and battery information for Darwin systems.

## Features

- Battery status monitoring (percentage, state, health)
- Power source detection (AC vs Battery)
- System power consumption metrics (CPU, GPU, total)
- Real-time monitoring with configurable intervals
- Thread-safe operations

### Usage

```go
import (
    "context"
    "fmt"
    "time"
    "github.com/sm-moshi/dmetrics-go/internal/power"
)

// Create a new power metrics provider
provider := power.NewProvider()

// Get current power stats
stats, err := provider.GetStats(context.Background())
if err != nil {
    log.Fatal(err)
}

// Print battery information
if stats.Source == types.PowerSourceBattery {
    fmt.Printf("Battery: %.1f%% (%v)\n", stats.Percentage, stats.State)
    fmt.Printf("Health: %v\n", stats.Health)
    fmt.Printf("Time Remaining: %v\n", stats.TimeRemaining.Round(time.Minute))
}

// Monitor power metrics
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
defer cancel()

ch, err := provider.Watch(ctx, 5*time.Second)
if err != nil {
    log.Fatal(err)
}

for stats := range ch {
    fmt.Printf("CPU: %.1fW, GPU: %.1fW, Total: %.1fW\n",
        stats.CPUPower, stats.GPUPower, stats.TotalPower)
}
```

### System Requirements

- macOS 10.15 or later
- IOKit framework
- System Management Controller (SMC) access

### Performance

Based on benchmark results:

- GetStats: ~1-2ms per call
- Memory allocation: ~256 bytes per call
- Concurrent access: Thread-safe with minimal contention

### Limitations

- Battery metrics only available on systems with a battery
- Some metrics may require administrator privileges
- Power consumption accuracy depends on hardware support

### Platform Support

- ✅ Darwin (macOS)
- ❌ Linux (planned)
- ❌ Windows (planned)
