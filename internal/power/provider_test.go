//go:build darwin
// +build darwin

package power_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sm-moshi/dmetrics-go/internal/power"
	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

func TestNewProvider(t *testing.T) {
	provider := power.NewProvider()
	require.NotNil(t, provider, "provider should not be nil")

	ctx := t.Context()

	t.Run("GetPowerSource", func(t *testing.T) {
		source, err := provider.GetPowerSource(ctx)
		require.NoError(t, err)
		assert.NotEqual(t, source, "", "power source should not be empty")
	})

	t.Run("GetBatteryPercentage", func(t *testing.T) {
		percentage, err := provider.GetBatteryPercentage(ctx)
		if errors.Is(err, types.ErrNoBattery) {
			t.Skip("No battery present")
		}
		require.NoError(t, err)
		assert.GreaterOrEqual(t, percentage, 0.0, "percentage should be >= 0%")
		assert.LessOrEqual(t, percentage, 100.0, "percentage should be <= 100%")
	})

	t.Run("GetStats", func(t *testing.T) {
		stats, err := provider.GetStats(ctx)
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.NotEqual(t, stats.Source, "", "power source should not be empty")
		if stats.Source == types.PowerSourceBattery {
			assert.GreaterOrEqual(t, stats.Percentage, 0.0, "battery percentage should be >= 0%")
			assert.LessOrEqual(t, stats.Percentage, 100.0, "battery percentage should be <= 100%")
			assert.NotEqual(t, stats.State, "", "battery state should not be empty")
			assert.NotEqual(t, stats.Health, "", "battery health should not be empty")
		}

		assert.GreaterOrEqual(t, stats.CPUPower, 0.0, "CPU power should be >= 0W")
		assert.GreaterOrEqual(t, stats.GPUPower, 0.0, "GPU power should be >= 0W")
		assert.GreaterOrEqual(t, stats.TotalPower, 0.0, "total power should be >= 0W")
		assert.WithinDuration(t, time.Now(), stats.Timestamp, 2*time.Second, "timestamp should be recent")
	})

	t.Run("Watch", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		ch, err := provider.Watch(ctx, 100*time.Millisecond)
		require.NoError(t, err)

		var updates int
		for stats := range ch {
			assert.NotNil(t, stats)
			assert.NotEqual(t, stats.Source, "", "power source should not be empty")
			updates++
		}

		assert.Greater(t, updates, 0, "should receive at least one update")
	})
}
