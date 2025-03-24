//go:build darwin
// +build darwin

// Package darwin provides Darwin-specific power metrics implementation.
package darwin

import (
	"context"
	"sync"
	"time"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

// Provider implements the power metrics collection for Darwin systems.
type Provider struct {
	mu sync.RWMutex
}

// NewProvider creates a new Darwin power metrics provider.
func NewProvider() *Provider {
	return &Provider{}
}

// GetStats returns current power and battery statistics.
func (p *Provider) GetStats(ctx context.Context) (*types.PowerStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getStats()
}

// GetPowerSource returns the current power source.
func (p *Provider) GetPowerSource(ctx context.Context) (types.PowerSource, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getPowerSource()
}

// GetBatteryPercentage returns the current battery charge percentage.
func (p *Provider) GetBatteryPercentage(ctx context.Context) (float64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getBatteryPercentage()
}

// GetBatteryState returns the current battery charging state.
func (p *Provider) GetBatteryState(ctx context.Context) (types.BatteryState, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getBatteryState()
}

// GetBatteryHealth returns the current battery health status.
func (p *Provider) GetBatteryHealth(ctx context.Context) (types.BatteryHealth, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getBatteryHealth()
}

// GetTimeRemaining returns the estimated time remaining on battery power.
func (p *Provider) GetTimeRemaining(ctx context.Context) (time.Duration, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getTimeRemaining()
}

// GetPowerConsumption returns the current system power consumption.
func (p *Provider) GetPowerConsumption(ctx context.Context) (float64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getPowerConsumption()
}

// Watch starts monitoring power metrics and sends updates through the returned channel.
func (p *Provider) Watch(ctx context.Context, interval time.Duration) (<-chan types.PowerStats, error) {
	if interval <= 0 {
		return nil, types.ErrInvalidInterval
	}

	ch := make(chan types.PowerStats)

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
					// Skip sending on error
					continue
				}
				ch <- *stats
			}
		}
	}()

	return ch, nil
}
