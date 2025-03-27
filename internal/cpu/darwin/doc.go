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
  - Current CPU frequency (may require elevated permissions)
  - Load averages for system activity assessment
  - Physical and logical core counts
  - Platform-specific optimisations (Apple Silicon vs Intel)

Known Limitations:
  - CPU frequency detection may return 0 without elevated permissions (sudo)
  - Some sysctl operations may fail without proper permissions
  - Apple Silicon frequency detection requires macOS 11.0 or later
  - Performance impact of frequent sampling should be considered

CPU Frequency Detection:
The package implements a multi-stage frequency detection strategy:
 1. Performance core frequency (Apple Silicon) - Primary method
 2. Efficiency core frequency (Apple Silicon) - Secondary method
 3. Traditional sysctl method (Intel) - Fallback method
 4. Various other fallback methods if the above fail

If all detection methods fail, the function returns 0 and logs appropriate warnings.
This behaviour allows applications to handle missing frequency data gracefully.

Example usage:

	provider := cpu.NewProvider()
	defer func() {
		if err := provider.Shutdown(); err != nil {
			log.Printf("error during shutdown: %v", err)
		}
	}()

	// Get current CPU stats
	stats, err := provider.GetStats(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)

	// Monitor CPU metrics
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := provider.Watch(ctx, time.Second)
	if err != nil {
		log.Fatal(err)
	}
	for stats := range ch {
		fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)
	}

Thread Safety:
All functions are designed to be thread-safe and can be called concurrently.
The implementation uses mutexes to protect shared resources and ensures proper
cleanup through the Shutdown method.

Watch Implementation:
The Watch function is implemented with several components for maintainability and reliability:
  - Input validation ensures proper interval values
  - Buffered channel (size 1) prevents blocking on slow consumers
  - Automatic old value dropping when channel is full
  - Clean shutdown through context cancellation
  - Proper resource cleanup with deferred operations
  - Thread-safe stats collection and sending

Resource Management:
The package implements proper resource management:
  - Automatic cleanup of system resources
  - Memory leak prevention in C code
  - Proper handling of host_processor_info data
  - Thread-safe access to shared state
  - Graceful shutdown with error handling
*/
package darwin

// Package cpu provides CPU metrics collection for Darwin systems.
//
// # API Stability
//
// This package follows semantic versioning. For v0.1:
// - All exported types and functions are considered stable
// - Method signatures will not change in incompatible ways
// - New methods may be added
// - Internal implementation details may change
//
// # Performance Characteristics
//
// CPU Stats Collection:
// - Initial call takes ~500ms to establish baseline
// - Subsequent calls take ~1-2ms
// - Memory: ~4KB per core for stats
// - Thread-safe with minimal lock contention
//
// Watch Operation:
// - Uses buffered channel (size 1) to prevent blocking
// - Updates every interval (minimum 100ms recommended)
// - Memory: ~8KB fixed + ~4KB per core
// - One active Watch call at a time recommended
// - Efficient value dropping for slow consumers
// - Clean shutdown through context cancellation
//
// Resource Usage:
// - No background goroutines when idle
// - Minimal syscall overhead
// - Automatic cleanup on shutdown
// - Proper error handling in cleanup paths
//
// Version Information
const (
	// Version represents the current package version.
	Version = "v0.1.0"

	// MinimumDarwinVersion is the minimum supported Darwin version.
	MinimumDarwinVersion = "10.15.0"
)
