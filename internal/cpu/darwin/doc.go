//go:build darwin
// +build darwin

/*
Package darwin provides the Darwin-specific implementation of CPU metrics collection.

This package implements the metrics.CPUMetrics interface by interfacing with various
Darwin-specific system calls. Each call serves a specific purpose:

  - sysctl: Provides core count and frequency data for system-wide CPU information
  - host_processor_info: Enables per-core CPU utilisation tracking
  - host_statistics: Delivers system-wide CPU metrics for overall health monitoring
  - mach_absolute_time: Ensures high-precision timing for accurate measurements

The implementation supports both Intel and Apple Silicon Macs, offering specialised
metrics for each architecture:

  - Total and per-core CPU utilisation
  - Current CPU frequency (may return 0 if detection fails)
  - Load averages for system activity assessment
  - Physical and logical core counts
  - Platform-specific optimisations (Apple Silicon vs Intel)

CPU Frequency Detection:
The package implements a multi-stage frequency detection strategy:
 1. Performance core frequency (Apple Silicon) - Primary method
 2. Efficiency core frequency (Apple Silicon) - Secondary method
 3. Traditional sysctl method (Intel) - Fallback method

If all detection methods fail, the function returns 0 and an error explaining
the failure. This behaviour allows applications to handle missing frequency
data gracefully.

Example usage:

	provider := cpu.NewProvider()
	stats, err := provider.GetStats(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)

Thread Safety:
All functions are designed to be thread-safe and can be called concurrently.
The implementation carefully manages system resources and ensures proper cleanup
through finalisation.
*/
package darwin
