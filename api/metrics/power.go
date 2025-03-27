// Package metrics provides interfaces for collecting system metrics.
package metrics

import (
	"context"
	"time"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

// PowerMetrics provides an interface for collecting power and battery metrics.
type PowerMetrics interface {
	// GetStats returns current power and battery statistics.
	// Returns types.ErrNoBattery if no battery is present.
	GetStats(ctx context.Context) (*types.PowerStats, error)

	// GetPowerSource returns the current power source (AC or Battery).
	GetPowerSource(ctx context.Context) (types.PowerSource, error)

	// GetBatteryPercentage returns the current battery charge percentage (0-100).
	// Returns types.ErrNoBattery if no battery is present.
	GetBatteryPercentage(ctx context.Context) (float64, error)

	// GetBatteryPresent returns whether a battery is present in the system.
	GetBatteryPresent(ctx context.Context) (bool, error)

	// Watch starts monitoring power metrics and sends updates to the provided channel.
	// The channel will be closed when monitoring stops or an error occurs.
	// The interval parameter specifies how often to collect metrics.
	// Returns types.ErrInvalidInterval if interval is not positive.
	Watch(ctx context.Context, interval time.Duration) (<-chan *types.PowerStats, error)

	// Shutdown cleans up any resources used by the provider.
	// This should be called when the provider is no longer needed.
	Shutdown() error
}
