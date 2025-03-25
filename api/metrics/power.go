// Package metrics defines interfaces for system metrics collection.
package metrics

import (
	"context"
	"time"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

// PowerMetrics defines the interface for power and battery metrics collection.
type PowerMetrics interface {
	// GetStats returns the current power and battery statistics.
	// Returns an error if the metrics cannot be collected.
	GetStats(ctx context.Context) (*types.PowerStats, error)

	// GetPowerSource returns the current power source (AC or Battery).
	// Returns PowerSourceUnknown and an error if the source cannot be determined.
	GetPowerSource(ctx context.Context) (types.PowerSource, error)

	// GetBatteryPercentage returns the current battery charge percentage (0-100).
	// Returns types.ErrNoBattery if no battery is present.
	// Returns an error if the percentage cannot be determined.
	GetBatteryPercentage(ctx context.Context) (float64, error)

	// GetBatteryState returns the current battery charging state.
	// Returns BatteryStateUnknown and types.ErrNoBattery if no battery is present.
	// Returns BatteryStateUnknown and an error if the state cannot be determined.
	GetBatteryState(ctx context.Context) (types.BatteryState, error)

	// GetBatteryPresent returns whether a battery is present in the system.
	// Returns false if no battery is present or if the status cannot be determined.
	GetBatteryPresent(ctx context.Context) (bool, error)

	// GetBatteryCharging returns whether the battery is currently charging.
	// Returns false and types.ErrNoBattery if no battery is present.
	// Returns false and an error if the charging state cannot be determined.
	GetBatteryCharging(ctx context.Context) (bool, error)

	// GetPowerConsumption returns the current system power consumption in Watts.
	// This includes CPU, GPU, and total system power if available.
	// Returns an error if the power consumption cannot be determined.
	GetPowerConsumption(ctx context.Context) (float64, error)

	// GetCPUPower returns the current CPU power consumption in Watts.
	// Returns an error if the power consumption cannot be determined.
	GetCPUPower(ctx context.Context) (float64, error)

	// GetGPUPower returns the current GPU power consumption in Watts.
	// Returns an error if the power consumption cannot be determined.
	GetGPUPower(ctx context.Context) (float64, error)

	// Watch starts monitoring power metrics and sends updates through the returned channel.
	// The channel will be closed when:
	// - The context is cancelled
	// - An unrecoverable error occurs
	// - The monitoring is stopped
	//
	// The interval parameter determines how often updates are sent.
	// A zero or negative interval will result in an error.
	Watch(ctx context.Context, interval time.Duration) (<-chan types.PowerStats, error)

	// Shutdown cleans up resources used by the provider.
	// This method should be called when the provider is no longer needed.
	// After calling Shutdown, other methods may return errors.
	Shutdown() error
}
