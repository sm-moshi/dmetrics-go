/*
Package dmetrics provides a comprehensive system metrics collection library for macOS.

Current Status (v0.1):
  - CPU metrics: ✓ Implemented and tested
  - Memory metrics: Planned
  - GPU metrics: Planned
  - Temperature sensors: Planned
  - Power metrics: Planned
  - Process metrics: Planned
  - Network metrics: Planned

The library enables real-time monitoring of various system metrics, with the CPU
package being the first fully implemented component:

CPU Metrics (Available Now):
  - Total and per-core CPU utilisation
  - CPU frequency detection (may require elevated permissions)
  - Load averages and core counts
  - Apple Silicon specific metrics (P-core/E-core frequencies)
  - Thread-safe concurrent monitoring

Example usage:

	package main

	import (
	    "fmt"
	    "log"
	    "time"
	    "github.com/sm-moshi/dmetrics-go/internal/cpu"
	)

	func main() {
	    provider := cpu.NewProvider()
	    defer provider.Shutdown()

	    // Get current CPU stats
	    stats, err := provider.GetStats()
	    if err != nil {
	        log.Fatal(err)
	    }
	    fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)

	    // Monitor CPU metrics
	    ch, err := provider.Watch(time.Second)
	    if err != nil {
	        log.Fatal(err)
	    }
	    for stats := range ch {
	        fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)
	        if stats.FrequencyMHz > 0 {
	            fmt.Printf("Frequency: %d MHz\n", stats.FrequencyMHz)
	        }
	    }
	}

The library interfaces with macOS system calls and frameworks through cgo:
  - sysctl: Core system statistics and configuration
  - host_processor_info: CPU utilisation tracking
  - mach_host: System-wide metrics collection
  - IOKit: Hardware information and power management (planned)
  - SMC: Temperature and fan control (planned)
  - libproc: Process monitoring (planned)

Known Limitations:
  - CPU frequency detection may require elevated permissions (sudo)
  - Some sysctl operations may fail without proper permissions
  - Apple Silicon specific features require macOS 11.0 or later

Package Organisation:
  - api/metrics: Public interfaces for metrics collection
  - internal/cpu: ✓ CPU metrics implementation (complete)
  - internal/gpu: Platform-specific GPU metrics (planned)
  - internal/power: Platform-specific power metrics (planned)
  - internal/temperature: Platform-specific temperature metrics (planned)
  - internal/memory: Platform-specific memory metrics (planned)
  - internal/network: Platform-specific network metrics (planned)
  - internal/process: Platform-specific process metrics (planned)
  - pkg/metrics/types: Common type definitions for all metrics

Thread Safety:
All implemented packages are designed for thread safety and efficiency, with proper
resource management and error handling. The implementation supports both Intel and
Apple Silicon Macs, with architecture-specific optimisations where available.
*/
package dmetrics
