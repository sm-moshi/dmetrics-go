//go:build darwin
// +build darwin

// Package cpu provides a platform-agnostic interface for retrieving CPU-related
// metrics from the system. For macOS (Darwin), it uses the host_statistics64 API
// to gather CPU usage metrics including system, user, and idle times.
//
// The package is designed with the following principles:
// - Platform independence through clear interface boundaries
// - Efficient sampling of CPU metrics with minimal overhead
// - Thread-safe access to CPU information
// - Context-aware operations for proper cancellation
//
// Performance Characteristics:
//   - Initial stats collection takes ~500ms to establish baseline
//   - Subsequent calls take ~1-2ms
//   - Memory usage is approximately 4KB per CPU core
//   - Thread-safe with minimal lock contention
//   - No background goroutines when idle
//
// Example usage:
//
//	// Create a new provider
//	provider := cpu.NewProvider()
//	defer provider.Shutdown()
//
//	// Get a single CPU stats snapshot
//	stats, err := provider.GetStats(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("CPU Usage: %.1f%%\n", stats.TotalUsage)
//
//	// Watch CPU stats with updates every second
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	statsCh, err := provider.Watch(ctx, time.Second)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for stats := range statsCh {
//	    fmt.Printf("Total: %.1f%%, User: %.1f%%, System: %.1f%%\n",
//	        stats.TotalUsage, stats.User, stats.System)
//	}
package cpu

import (
	"github.com/sm-moshi/dmetrics-go/api/metrics"
	"github.com/sm-moshi/dmetrics-go/internal/cpu/darwin"
)

// NewProvider creates a new CPU metrics provider for the current platform.
// On Darwin systems, this returns a provider that uses host_statistics64 for CPU metrics.
// The returned provider is thread-safe and can be used concurrently.
// Remember to call Shutdown() when done to release resources.
func NewProvider() metrics.CPUMetrics {
	return darwin.NewProvider()
}
