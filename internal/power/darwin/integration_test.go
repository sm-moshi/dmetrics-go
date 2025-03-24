//go:build darwin && integration
// +build darwin,integration

package darwin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

func TestPowerMetricsIntegration(t *testing.T) {
	provider := NewProvider()
	ctx := context.Background()

	// Test continuous monitoring
	t.Run("ContinuousMonitoring", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		ch, err := provider.Watch(ctx, time.Second)
		require.NoError(t, err)

		var readings []types.PowerStats
		for reading := range ch {
			readings = append(readings, reading)
		}

		require.GreaterOrEqual(t, len(readings), 4, "Should have at least 4 readings in 5 seconds")

		// Verify readings are sequential
		for i := 1; i < len(readings); i++ {
			assert.Greater(t, readings[i].Timestamp.UnixNano(),
				readings[i-1].Timestamp.UnixNano(),
				"Timestamps should be sequential")
		}
	})

	// Test power source transition detection
	t.Run("PowerSourceDetection", func(t *testing.T) {
		// Get initial power source
		initial, err := provider.GetPowerSource(ctx)
		require.NoError(t, err)

		t.Logf("Current power source: %v", initial)
		t.Log("Please change power source (connect/disconnect AC) within 30 seconds")

		// Monitor for power source changes
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		ch, err := provider.Watch(ctx, time.Second)
		require.NoError(t, err)

		var sourceChanged bool
		for stats := range ch {
			if stats.Source != initial {
				sourceChanged = true
				t.Logf("Power source changed to: %v", stats.Source)
				break
			}
		}

		if !sourceChanged {
			t.Skip("Power source was not changed during test")
		}
	})

	// Test battery metrics consistency
	t.Run("BatteryMetricsConsistency", func(t *testing.T) {
		// Take multiple readings and verify consistency
		var lastPercentage float64
		var lastState types.BatteryState

		for i := 0; i < 5; i++ {
			stats, err := provider.GetStats(ctx)
			require.NoError(t, err)

			if i == 0 {
				lastPercentage = stats.Percentage
				lastState = stats.State
				continue
			}

			// Battery percentage shouldn't change dramatically in short time
			assert.InDelta(t, lastPercentage, stats.Percentage, 1.0,
				"Battery percentage changed too rapidly")

			// State should be consistent unless explicitly changed
			assert.Equal(t, lastState, stats.State,
				"Battery state changed unexpectedly")

			lastPercentage = stats.Percentage
			lastState = stats.State

			time.Sleep(time.Second)
		}
	})

	// Test power consumption metrics
	t.Run("PowerConsumptionMetrics", func(t *testing.T) {
		stats, err := provider.GetStats(ctx)
		require.NoError(t, err)

		// Verify power consumption values
		assert.GreaterOrEqual(t, stats.CPUPower, 0.0)
		assert.GreaterOrEqual(t, stats.GPUPower, 0.0)
		assert.GreaterOrEqual(t, stats.TotalPower, stats.CPUPower+stats.GPUPower)

		// Monitor power consumption changes
		var readings []float64
		for i := 0; i < 5; i++ {
			power, err := provider.GetPowerConsumption(ctx)
			require.NoError(t, err)
			readings = append(readings, power)
			time.Sleep(time.Second)
		}

		// Verify we get varying power readings
		var hasVariation bool
		for i := 1; i < len(readings); i++ {
			if readings[i] != readings[0] {
				hasVariation = true
				break
			}
		}
		assert.True(t, hasVariation, "Power consumption should vary over time")
	})
}
