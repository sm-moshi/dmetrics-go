/*
Package dmetrics provides a comprehensive system metrics collection library for macOS.

It offers real-time monitoring of various system metrics including:
  - CPU usage, frequency, and load averages
  - Memory usage and virtual memory statistics
  - GPU metrics and VRAM usage
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
	    fmt.Printf("Frequency: %d MHz\n", stats.FrequencyMHz)
	}

The library uses cgo to interface with macOS system calls and frameworks:
  - sysctl for CPU and memory statistics
  - IOKit for GPU and power information
  - SMC for temperature and fan data
  - libproc for process monitoring

All packages are designed to be thread-safe and efficient, with proper resource cleanup
and error handling. The library supports both Intel and Apple Silicon Macs.

Package Structure:
  - api/metrics: Public interfaces for metrics collection
  - internal/cpu: Platform-specific CPU metrics implementation
  - internal/gpu: Platform-specific GPU metrics implementation
  - internal/power: Platform-specific power metrics implementation
  - internal/temperature: Platform-specific temperature metrics implementation
  - internal/memory: Platform-specific memory metrics implementation
  - internal/network: Platform-specific network metrics implementation
  - internal/process: Platform-specific process metrics implementation
  - pkg/metrics/types: Common type definitions for all metrics
*/
package dmetrics
