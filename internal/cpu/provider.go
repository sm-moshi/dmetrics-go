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
// Example usage:
//
//	provider := cpu.NewProvider()
//	stats, err := provider.GetStats(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("CPU Usage: User=%.1f%%, System=%.1f%%\n", stats.User, stats.System)
package cpu

import (
	"github.com/sm-moshi/dmetrics-go/api/metrics"
	"github.com/sm-moshi/dmetrics-go/internal/cpu/darwin"
)

// NewProvider creates a new CPU metrics provider for the current platform.
// On Darwin systems, this returns a provider that uses host_statistics64 for CPU metrics.
func NewProvider() metrics.CPUMetrics {
	return darwin.NewProvider()
}
