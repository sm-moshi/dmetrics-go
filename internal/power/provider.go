//go:build darwin
// +build darwin

// Package power provides a platform-agnostic interface for retrieving power-related
// information from the system. For macOS (Darwin), it uses the IOKit framework's
// IOPowerSources API to gather basic power metrics like battery presence, charging
// state, and capacity percentage.
//
// The package is designed with the following principles:
// - Platform independence through clear interface boundaries
// - Minimal dependencies on low-level system APIs
// - Thread-safe access to power information
// - Context-aware operations for proper cancellation
//
// Example usage:
//
//	provider := power.NewProvider()
//	stats, err := provider.GetStats(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Power source: %s, Battery: %.1f%%\n", stats.Source, stats.Percentage)
package power

import (
	"github.com/sm-moshi/dmetrics-go/api/metrics"
	"github.com/sm-moshi/dmetrics-go/internal/power/darwin"
)

// NewProvider creates a new power metrics provider for the current platform.
// On Darwin systems, this returns a provider that uses IOKit for power metrics.
func NewProvider() metrics.PowerMetrics {
	return darwin.NewProvider()
}
