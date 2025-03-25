# dmetrics-go Examples

This directory contains example applications demonstrating the usage of dmetrics-go library for collecting system metrics on macOS.

## Requirements

- macOS 10.15 (Catalina) or later
- Go 1.21 or later
- Root privileges may be required for some metrics (e.g., CPU frequency)

## Examples

### CPU Statistics (`cpu/cpu_stats.go`)

Demonstrates real-time CPU monitoring including:

- Core usage visualization
- CPU frequency
- Load averages
- Physical core count

```bash
go run cpu/cpu_stats.go
```

### Power Statistics (`power/power_stats.go`)

Shows power-related metrics including:

- Power source information
- Battery status and health
- Power consumption metrics
- Charging state

```bash
go run power/power_stats.go
```

### Combined System Monitor (`combined_stats.go`)

Provides a unified view of system metrics including:

- CPU usage and frequency
- Power consumption
- Battery status
- Real-time updates

```bash
go run combined_stats.go
```

## Notes

1. Some metrics may require root privileges. Run with `sudo` if needed:

   ```bash
   sudo go run cpu/cpu_stats.go
   ```

2. Power metrics are only available on battery-powered devices (MacBooks).

3. Update intervals can be adjusted by modifying the `updateInterval` constant in each example.

## Error Handling

The examples demonstrate proper error handling including:

- Graceful shutdown on Ctrl+C
- Context cancellation
- Permission-related errors
- Hardware availability checks

## Contributing

Feel free to submit improvements or additional examples via pull requests.
