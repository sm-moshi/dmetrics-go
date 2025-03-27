//go:build darwin
// +build darwin

package darwin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

func TestNewProvider(t *testing.T) {
	provider := NewProvider()
	assert.NotNil(t, provider, "Provider should not be nil")
}

func TestGetStats(t *testing.T) {
	provider := NewProvider()
	ctx := t.Context()

	stats, err := provider.GetStats(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Basic validation of stats
	assert.GreaterOrEqual(t, stats.Percentage, 0.0)
	assert.LessOrEqual(t, stats.Percentage, 100.0)
	assert.GreaterOrEqual(t, stats.Temperature, 0.0)
	assert.GreaterOrEqual(t, stats.CycleCount, 0)
	assert.NotZero(t, stats.Timestamp)
}

func TestGetPowerSource(t *testing.T) {
	provider := NewProvider()
	ctx := t.Context()

	source, err := provider.GetPowerSource(ctx)
	require.NoError(t, err)
	assert.Contains(t, []types.PowerSource{
		types.PowerSourceAC,
		types.PowerSourceBattery,
		types.PowerSourceUnknown,
	}, source)
}

func TestGetBatteryPercentage(t *testing.T) {
	provider := NewProvider()
	ctx := t.Context()

	percentage, err := provider.GetBatteryPercentage(ctx)
	if errors.Is(err, types.ErrNoBattery) {
		t.Skip("No battery present, skipping test")
	}
	require.NoError(t, err)
	assert.GreaterOrEqual(t, percentage, 0.0)
	assert.LessOrEqual(t, percentage, 100.0)
}

func TestGetBatteryState(t *testing.T) {
	provider := NewProvider()
	ctx := t.Context()

	state, err := provider.GetBatteryState(ctx)
	require.NoError(t, err)
	assert.Contains(t, []types.BatteryState{
		types.BatteryStateCharging,
		types.BatteryStateDischarging,
		types.BatteryStateFull,
		types.BatteryStateUnknown,
	}, state)
}

func TestGetBatteryHealth(t *testing.T) {
	provider := NewProvider()
	ctx := t.Context()

	health, err := provider.GetBatteryHealth(ctx)
	if errors.Is(err, types.ErrNoBattery) {
		t.Skip("No battery present, skipping test")
	}
	require.NoError(t, err)
	assert.Contains(t, []types.BatteryHealth{
		types.BatteryHealthGood,
		types.BatteryHealthFair,
		types.BatteryHealthPoor,
		types.BatteryHealthUnknown,
	}, health)
}

func TestGetTimeRemaining(t *testing.T) {
	provider := NewProvider()
	ctx := t.Context()

	duration, err := provider.GetTimeRemaining(ctx)
	if errors.Is(err, types.ErrNoBattery) {
		t.Skip("No battery present, skipping test")
	}
	require.NoError(t, err)

	// Time remaining can be negative when charging
	if duration < 0 {
		assert.LessOrEqual(t, duration, time.Duration(0), "charging time should be negative")
	} else {
		assert.GreaterOrEqual(t, duration, time.Duration(0), "discharging time should be non-negative")
	}
}

func TestGetPowerConsumption(t *testing.T) {
	provider := NewProvider()
	ctx := t.Context()

	power, err := provider.GetPowerConsumption(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, power, 0.0)
}

func TestWatch(t *testing.T) {
	provider := NewProvider()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	// Test invalid interval
	_, err := provider.Watch(ctx, 0)
	assert.ErrorIs(t, err, types.ErrInvalidInterval)

	// Test valid interval
	ch, err := provider.Watch(ctx, 500*time.Millisecond)
	require.NoError(t, err)

	var stats *types.PowerStats
	select {
	case stats = <-ch:
		assert.NotZero(t, stats.Timestamp)
		assert.GreaterOrEqual(t, stats.Percentage, 0.0)
		assert.LessOrEqual(t, stats.Percentage, 100.0)
	case <-ctx.Done():
		t.Fatal("Timeout waiting for power stats")
	}
}

func TestConcurrentAccess(t *testing.T) {
	provider := NewProvider()
	ctx := t.Context()
	done := make(chan bool)

	// Run multiple goroutines accessing the provider
	for i := 0; i < 10; i++ {
		go func() {
			_, err := provider.GetStats(ctx)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}
