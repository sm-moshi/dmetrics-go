// Package metrics provides interfaces for collecting system metrics.
package metrics

import (
	"errors"
	"time"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

// Common errors returned by CPU metrics collection.
var (
	// ErrInvalidInterval is returned when a non-positive interval is provided for monitoring.
	ErrInvalidInterval = errors.New("interval must be positive")

	// ErrUnsupportedPlatform is returned when attempting to use platform-specific features
	// that are not available on the current system (e.g., Apple Silicon features on Intel).
	ErrUnsupportedPlatform = errors.New("operation not supported on this platform")

	// ErrHardwareAccess is returned when hardware information cannot be accessed.
	ErrHardwareAccess = errors.New("failed to access hardware information")
)

// CPUMetrics provides an interface for collecting CPU metrics.
type CPUMetrics interface {
	// GetFrequency returns the current CPU frequency in MHz.
	// For Apple Silicon Macs, this returns the highest frequency among all cores.
	// Returns ErrHardwareAccess if the frequency cannot be determined.
	GetFrequency() (uint64, error)

	// GetPerformanceFrequency returns the current frequency of performance cores in MHz.
	// This method is only applicable to Apple Silicon Macs and will return 0 on Intel Macs.
	// Returns ErrUnsupportedPlatform on Intel Macs.
	// Returns ErrHardwareAccess if the frequency cannot be determined.
	GetPerformanceFrequency() (uint64, error)

	// GetEfficiencyFrequency returns the current frequency of efficiency cores in MHz.
	// This method is only applicable to Apple Silicon Macs and will return 0 on Intel Macs.
	// Returns ErrUnsupportedPlatform on Intel Macs.
	// Returns ErrHardwareAccess if the frequency cannot be determined.
	GetEfficiencyFrequency() (uint64, error)

	// GetCoreCount returns the number of physical CPU cores.
	// Returns ErrHardwareAccess if the core count cannot be determined.
	GetCoreCount() (int, error)

	// GetPerformanceCoreCount returns the number of performance cores.
	// This method is only applicable to Apple Silicon Macs and will return 0 on Intel Macs.
	// Returns ErrUnsupportedPlatform on Intel Macs.
	// Returns ErrHardwareAccess if the core count cannot be determined.
	GetPerformanceCoreCount() (int, error)

	// GetEfficiencyCoreCount returns the number of efficiency cores.
	// This method is only applicable to Apple Silicon Macs and will return 0 on Intel Macs.
	// Returns ErrUnsupportedPlatform on Intel Macs.
	// Returns ErrHardwareAccess if the core count cannot be determined.
	GetEfficiencyCoreCount() (int, error)

	// GetStats returns current CPU statistics.
	// Returns ErrHardwareAccess if the statistics cannot be collected.
	GetStats() (*types.CPUStats, error)

	// GetPlatform returns information about the CPU platform.
	// Returns ErrHardwareAccess if the platform information cannot be determined.
	GetPlatform() (*types.CPUPlatform, error)

	// Watch starts monitoring CPU metrics and sends updates to the provided channel.
	// The channel will be closed when monitoring stops or an error occurs.
	// The interval parameter specifies how often to collect metrics.
	// Returns ErrInvalidInterval if interval is not positive.
	Watch(interval time.Duration) (<-chan *types.CPUStats, error)

	// Shutdown cleans up any resources used by the provider.
	// This should be called when the provider is no longer needed.
	// After calling Shutdown, other methods may return errors.
	Shutdown() error
}
