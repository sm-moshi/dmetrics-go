// Package metrics provides interfaces for collecting system metrics.
//
// Example usage:
//
//	provider := cpu.NewProvider()
//	defer provider.Shutdown()
//
//	// Get current CPU stats
//	stats, err := provider.GetStats()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("CPU Usage: %.2f%%\n", stats.Usage)
//
//	// Monitor CPU metrics
//	ch, err := provider.Watch(time.Second)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for stats := range ch {
//	    fmt.Printf("CPU Usage: %.2f%%\n", stats.Usage)
//	}
package metrics

import (
	"context"
	"errors"
	"time"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

// Common errors that may be returned by CPU metrics collection.
var (
	// ErrInvalidInterval is returned when a non-positive interval is provided for monitoring.
	ErrInvalidInterval = errors.New("interval must be positive")

	// ErrUnsupportedPlatform is returned when attempting to use platform-specific features
	// that are not available on the current system (e.g., Apple Silicon features on Intel).
	ErrUnsupportedPlatform = errors.New("operation not supported on this platform")

	// ErrHardwareAccess is returned when hardware information cannot be accessed.
	ErrHardwareAccess = errors.New("failed to access hardware information")

	// ErrShutdown is returned when attempting to use a provider that has been shut down.
	ErrShutdown = errors.New("provider has been shut down")
)

// CPUMetrics provides an interface for collecting CPU metrics.
// All methods are safe for concurrent use unless otherwise noted.
//
// Example usage of key methods:
//
//	// Get CPU frequency
//	freq, err := cpu.GetFrequency()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("CPU Frequency: %d MHz\n", freq)
//
//	// Get platform information
//	platform, err := cpu.GetPlatform()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if platform.IsAppleSilicon {
//	    pFreq, err := cpu.GetPerformanceFrequency()
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Printf("Performance Core Frequency: %d MHz\n", pFreq)
//	}
type CPUMetrics interface {
	// GetFrequency returns the current CPU frequency in MHz.
	// For Apple Silicon Macs, this returns the highest frequency among all cores.
	// Returns:
	//   - Current frequency in MHz
	//   - ErrHardwareAccess if the frequency cannot be determined
	//   - ErrShutdown if the provider has been shut down
	GetFrequency() (uint64, error)

	// GetPerformanceFrequency returns the current frequency of performance cores in MHz.
	// This method is only applicable to Apple Silicon Macs.
	// Returns:
	//   - Current performance core frequency in MHz
	//   - ErrUnsupportedPlatform on Intel Macs
	//   - ErrHardwareAccess if the frequency cannot be determined
	//   - ErrShutdown if the provider has been shut down
	GetPerformanceFrequency() (uint64, error)

	// GetEfficiencyFrequency returns the current frequency of efficiency cores in MHz.
	// This method is only applicable to Apple Silicon Macs.
	// Returns:
	//   - Current efficiency core frequency in MHz
	//   - ErrUnsupportedPlatform on Intel Macs
	//   - ErrHardwareAccess if the frequency cannot be determined
	//   - ErrShutdown if the provider has been shut down
	GetEfficiencyFrequency() (uint64, error)

	// GetCoreCount returns the number of physical CPU cores.
	// Returns:
	//   - Number of physical cores
	//   - ErrHardwareAccess if the core count cannot be determined
	//   - ErrShutdown if the provider has been shut down
	GetCoreCount() (int, error)

	// GetPerformanceCoreCount returns the number of performance cores.
	// This method is only applicable to Apple Silicon Macs.
	// Returns:
	//   - Number of performance cores
	//   - ErrUnsupportedPlatform on Intel Macs
	//   - ErrHardwareAccess if the core count cannot be determined
	//   - ErrShutdown if the provider has been shut down
	GetPerformanceCoreCount() (int, error)

	// GetEfficiencyCoreCount returns the number of efficiency cores.
	// This method is only applicable to Apple Silicon Macs.
	// Returns:
	//   - Number of efficiency cores
	//   - ErrUnsupportedPlatform on Intel Macs
	//   - ErrHardwareAccess if the core count cannot be determined
	//   - ErrShutdown if the provider has been shut down
	GetEfficiencyCoreCount() (int, error)

	// GetStats returns current CPU statistics.
	// Returns:
	//   - CPU statistics including usage, frequency, and core information
	//   - ErrHardwareAccess if the statistics cannot be collected
	//   - ErrShutdown if the provider has been shut down
	GetStats() (*types.CPUStats, error)

	// GetPlatform returns information about the CPU platform.
	// Returns:
	//   - CPU platform information including processor type and capabilities
	//   - ErrHardwareAccess if the platform information cannot be determined
	//   - ErrShutdown if the provider has been shut down
	GetPlatform() (*types.CPUPlatform, error)

	// Watch starts monitoring CPU metrics and sends updates to the provided channel.
	// The context can be used to stop monitoring. When the context is cancelled,
	// the channel will be closed after any pending updates are sent.
	//
	// The returned channel is buffered with a capacity of 1 to prevent blocking
	// when the consumer is slower than the update interval. If a consumer cannot
	// keep up with updates, the oldest update will be dropped in favor of the newest.
	//
	// This method is safe for concurrent use, but only one Watch call should be active
	// at a time. Additional calls will return an error.
	// Returns:
	//   - Channel receiving CPU statistics updates
	//   - ErrInvalidInterval if interval is not positive
	//   - ErrShutdown if the provider has been shut down
	Watch(ctx context.Context, interval time.Duration) (<-chan *types.CPUStats, error)

	// Shutdown cleans up any resources used by the provider.
	// This method is safe for concurrent use but should only be called once.
	// After calling Shutdown, all other methods will return ErrShutdown.
	Shutdown() error
}
