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
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getStats()
}

// GetPowerSource returns the current power source.
func (p *Provider) GetPowerSource(ctx context.Context) (types.PowerSource, error) {
	if ctx.Err() != nil {
		return types.PowerSourceUnknown, ctx.Err()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getPowerSource()
}

// GetBatteryPercentage returns the current battery charge percentage.
func (p *Provider) GetBatteryPercentage(ctx context.Context) (float64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getBatteryPercentage()
}

// GetBatteryPresent returns whether a battery is present in the system.
func (p *Provider) GetBatteryPresent(context.Context) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats, err := getStats()
	if err != nil {
		return false, err
	}
	return stats.IsPresent, nil
}

// GetBatteryState returns the current battery charging state.
func (p *Provider) GetBatteryState(ctx context.Context) (types.BatteryState, error) {
	if ctx.Err() != nil {
		return types.BatteryStateUnknown, ctx.Err()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getBatteryState()
}

// GetBatteryHealth returns the current battery health status.
func (p *Provider) GetBatteryHealth(ctx context.Context) (types.BatteryHealth, error) {
	if ctx.Err() != nil {
		return types.BatteryHealthUnknown, ctx.Err()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getBatteryHealth()
}

// GetTimeRemaining returns the estimated time remaining on battery power.
func (p *Provider) GetTimeRemaining(ctx context.Context) (time.Duration, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getTimeRemaining()
}

// GetPowerConsumption returns the current system power consumption.
func (p *Provider) GetPowerConsumption(ctx context.Context) (float64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return getPowerConsumption()
}

// GetBatteryCharging returns whether the battery is currently charging.
func (p *Provider) GetBatteryCharging(ctx context.Context) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats, err := getStats()
	if err != nil {
		return false, err
	}
	if !stats.IsPresent {
		return false, types.ErrNoBattery
	}
	return stats.State == types.BatteryStateCharging, nil
}

// Watch monitors power metrics and sends updates through the returned channel.
// The interval parameter determines how often updates are sent.
// The returned channel will be closed when the context is cancelled or an error occurs.
func (p *Provider) Watch(ctx context.Context, interval time.Duration) (<-chan *types.PowerStats, error) {
	if interval <= 0 {
		return nil, types.ErrInvalidInterval
	}

	ch := make(chan *types.PowerStats)

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
				case ch <- stats:
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
	return nil
}
