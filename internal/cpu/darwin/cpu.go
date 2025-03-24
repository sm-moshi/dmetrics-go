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
	return &Provider{}
}

// GetStats returns current CPU statistics.
func (p *Provider) GetStats(ctx context.Context) (*types.CPUStats, error) {
	stats, err := getStats()
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetUsage returns the current total CPU usage percentage (0-100).
func (p *Provider) GetUsage(ctx context.Context, interval time.Duration) (float64, error) {
	return usage()
}

// GetFrequency returns the current CPU frequency in MHz.
func (p *Provider) GetFrequency(ctx context.Context) (uint64, error) {
	return getFrequency()
}

// GetCoreCount returns the number of CPU cores.
func (p *Provider) GetCoreCount(ctx context.Context) (int, error) {
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	return stats.PhysicalCores, nil
}

// Watch starts monitoring CPU metrics and sends updates through the returned channel.
func (p *Provider) Watch(ctx context.Context, interval time.Duration) (<-chan types.CPUStats, error) {
	ch := make(chan types.CPUStats)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats, err := p.GetStats(ctx)
				if err != nil {
					continue
				}
				ch <- *stats
			}
		}
	}()

	return ch, nil
}
