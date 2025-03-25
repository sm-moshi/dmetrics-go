//go:build darwin
// +build darwin

// Package darwin provides Darwin-specific CPU metrics implementation.
package darwin

import (
	"context"
	"time"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

// Provider implements the CPU metrics collection for Darwin systems.
type Provider struct {
	// Add any necessary fields for the provider
}

// NewProvider creates a new Darwin CPU metrics provider.
func NewProvider() *Provider {
	initCleanup()
	return &Provider{}
}

// GetStats returns current CPU statistics.
func (p *Provider) GetStats(ctx context.Context) (*types.CPUStats, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	stats, err := getStats()
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetUsage returns the current total CPU usage percentage (0-100).
// The interval parameter determines the sampling period for calculating usage.
func (p *Provider) GetUsage(ctx context.Context, interval time.Duration) (float64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	// Create a timer for the interval
	timer := time.NewTimer(interval)
	defer timer.Stop()

	// Get initial usage
	initial, err := usage()
	if err != nil {
		return 0, err
	}

	// Wait for either context cancellation or interval completion
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-timer.C:
		var final float64
		final, err = usage()
		if err != nil {
			return 0, err
		}
		// Return the difference in usage over the interval
		return final - initial, nil
	}
}

// GetFrequency returns the current CPU frequency in MHz.
func (p *Provider) GetFrequency(ctx context.Context) (uint64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	return getFrequency()
}

// GetCoreCount returns the number of CPU cores.
func (p *Provider) GetCoreCount(ctx context.Context) (int, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	return stats.PhysicalCores, nil
}

// GetEfficiencyCoreCount returns the number of efficiency cores on Apple Silicon.
// Returns 0 on Intel processors.
func (p *Provider) GetEfficiencyCoreCount(ctx context.Context) (int, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	return stats.EfficiencyCores, nil
}

// GetPerformanceCoreCount returns the number of performance cores on Apple Silicon.
// Returns 0 on Intel processors.
func (p *Provider) GetPerformanceCoreCount(ctx context.Context) (int, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	return stats.PerformanceCores, nil
}

// Watch monitors CPU statistics and sends updates through the returned channel.
// The interval parameter determines how often updates are sent.
// The returned channel will be closed when the context is cancelled or an error occurs.
func (p *Provider) Watch(ctx context.Context, interval time.Duration) (<-chan types.CPUStats, error) {
	if interval <= 0 {
		return nil, types.ErrInvalidInterval
	}

	ch := make(chan types.CPUStats)

	go func() {
		defer close(ch)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats, err := p.GetStats(ctx)
				if err != nil {
					// Log error if needed
					return
				}

				// Try to send stats, but respect context cancellation
				select {
				case ch <- *stats:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// Shutdown cleans up resources used by the provider.
func (p *Provider) Shutdown() error {
	cleanup()
	return nil
}
