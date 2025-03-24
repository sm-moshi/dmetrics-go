//go:build darwin
// +build darwin

/*
Package darwin provides the Darwin-specific implementation of CPU metrics collection.

It implements the metrics.CPUMetrics interface defined in api/metrics/cpu.go and uses
various Darwin-specific system calls to gather CPU statistics:

  - sysctl for CPU core count and frequency
  - host_processor_info for CPU usage statistics
  - host_statistics for system-wide CPU metrics
  - mach_absolute_time for high-precision timing

The package supports both Intel and Apple Silicon Macs, providing detailed CPU metrics
including:

  - Total and per-core CPU usage
  - Current CPU frequency
  - Load averages
  - Physical and logical core counts
  - Platform-specific information (Apple Silicon vs Intel)

Example usage:

	provider := cpu.NewProvider()
	stats, err := provider.GetStats(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)

Thread Safety:
All functions in this package are thread-safe and can be called concurrently.
The implementation properly manages system resources and handles cleanup.
*/
package darwin
