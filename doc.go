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
	    "fmt"
	    "log"
	    "github.com/sm-moshi/dmetrics-go/cpu"
	)

	func main() {
	    stats, err := cpu.Get()
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

For detailed documentation of each subsystem, see the respective packages:
  - cpu: CPU metrics and processor information
  - gpu: GPU statistics and VRAM usage
  - memory: Physical and virtual memory metrics
  - power: Battery status and power source info
  - temp: Temperature sensors and fan speeds
  - process: Process statistics and monitoring
  - network: Network interface metrics
*/
package dmetrics
