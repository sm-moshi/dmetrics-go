//go:build darwin
// +build darwin

package cpu_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sm-moshi/dmetrics-go/internal/cpu"
)

func TestNewProvider(t *testing.T) {
	provider := cpu.NewProvider()
	require.NotNil(t, provider, "provider should not be nil")

	ctx := t.Context()

	t.Run("GetFrequency", func(t *testing.T) {
		freq, err := provider.GetFrequency(ctx)
		require.NoError(t, err)
		assert.Greater(t, freq, uint64(0), "frequency should be > 0 MHz")
	})

	t.Run("GetUsage", func(t *testing.T) {
		usage, err := provider.GetUsage(ctx, 100*time.Millisecond)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, usage, 0.0, "usage should be >= 0%")
		assert.LessOrEqual(t, usage, 100.0, "usage should be <= 100%")
	})

	t.Run("GetCoreCount", func(t *testing.T) {
		cores, err := provider.GetCoreCount(ctx)
		require.NoError(t, err)
		assert.Greater(t, cores, 0, "should have at least one core")
	})

	t.Run("GetStats", func(t *testing.T) {
		stats, err := provider.GetStats(ctx)
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Greater(t, stats.PhysicalCores, 0, "should have at least one physical core")
		assert.Greater(t, stats.FrequencyMHz, uint64(0), "CPU frequency should be > 0 MHz")
		assert.Len(t, stats.CoreUsage, stats.PhysicalCores, "should have usage values for each core")

		for i, usage := range stats.CoreUsage {
			assert.GreaterOrEqual(t, usage, 0.0, "core %d usage should be >= 0%%", i)
			assert.LessOrEqual(t, usage, 100.0, "core %d usage should be <= 100%%", i)
		}

		assert.GreaterOrEqual(t, stats.TotalUsage, 0.0, "total usage should be >= 0%")
		assert.LessOrEqual(t, stats.TotalUsage, 100.0, "total usage should be <= 100%")

		for i, load := range stats.LoadAvg {
			assert.GreaterOrEqual(t, load, 0.0, "load average %d should be >= 0", i)
		}

		assert.WithinDuration(t, time.Now(), stats.Timestamp, 2*time.Second, "timestamp should be recent")
	})

	t.Run("Watch", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		ch, err := provider.Watch(ctx, 100*time.Millisecond)
		require.NoError(t, err)

		var updates int
		for stats := range ch {
			assert.GreaterOrEqual(t, stats.TotalUsage, 0.0, "usage should be >= 0%")
			assert.LessOrEqual(t, stats.TotalUsage, 100.0, "usage should be <= 100%")
			updates++
		}

		assert.Greater(t, updates, 0, "should receive at least one update")
	})
}
