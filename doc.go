/*
Package dmetrics provides a comprehensive system metrics collection library for macOS.

The library enables real-time monitoring of various system metrics:
  - CPU utilisation, frequency (where available), and load averages
  - Memory usage and virtual memory statistics
  - GPU metrics and VRAM utilisation
  - Temperature sensors and fan speeds
  - Power source and battery information
  - Process statistics
  - Network interface metrics

Example usage:

	package main

	import (
	    "context"
	    "fmt"
	    "log"
	    "github.com/sm-moshi/dmetrics-go/internal/cpu"
	    "github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
	)

	func main() {
	    provider := cpu.NewProvider()
	    stats, err := provider.GetStats(context.Background())
	    if err != nil {
	        log.Fatal(err)
	    }
	    fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)
	    if stats.FrequencyMHz > 0 {
	        fmt.Printf("Frequency: %d MHz\n", stats.FrequencyMHz)
	    } else {
	        fmt.Println("CPU frequency detection not available")
	    }
	}

The library interfaces with macOS system calls and frameworks through cgo:
  - sysctl: Core system statistics and configuration
  - IOKit: GPU metrics and power management
  - SMC: Temperature monitoring and fan control
  - libproc: Process monitoring and statistics

All packages are designed for thread safety and efficiency, with proper resource
management and error handling. The implementation supports both Intel and
Apple Silicon Macs, with architecture-specific optimisations where available.

Package Organisation:
  - api/metrics: Public interfaces for metrics collection
  - internal/cpu: Platform-specific CPU metrics implementation
  - internal/gpu: Platform-specific GPU metrics implementation (TODO)
  - internal/power: Platform-specific power metrics implementation (TODO)
  - internal/temperature: Platform-specific temperature metrics implementation (TODO)
  - internal/memory: Platform-specific memory metrics implementation (TODO)
  - internal/network: Platform-specific network metrics implementation (TODO)
  - internal/process: Platform-specific process metrics implementation (TODO)
  - pkg/metrics/types: Common type definitions for all metrics
*/
package dmetrics
