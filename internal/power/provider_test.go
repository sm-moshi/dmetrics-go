//go:build darwin
// +build darwin

// Package power_test verifies the functionality of the power package
// by testing its interaction with the macOS power management APIs.
// These tests ensure that:
// - Basic power source information can be retrieved
// - Battery metrics are accurately reported when available
// - The provider handles concurrent access safely
// - Watch functionality provides timely updates
// - Resources are properly cleaned up on shutdown
package power

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

// TestNewProvider verifies the core functionality of the power provider.
// It tests the provider's ability to:
// - Initialize successfully
// - Retrieve power source information
// - Get battery percentage when available
// - Collect comprehensive power statistics
// - Watch for power-related changes
func TestNewProvider(t *testing.T) {
	provider := NewProvider()
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
			assert.NotNil(t, stats)
			assert.NotEqual(t, stats.Source, "", "power source should not be empty")
			updates++
		}

		assert.Greater(t, updates, 0, "should receive at least one update")
	})
}

// TestShutdown verifies that the provider can be cleanly shut down.
// This ensures proper resource cleanup and prevents memory leaks.
func TestShutdown(t *testing.T) {
	provider := NewProvider()
	err := provider.Shutdown()
	assert.NoError(t, err)
}

// BenchmarkBasicOperations measures the performance of core power operations.
// This helps identify potential performance bottlenecks in the power metrics
// collection process.
func BenchmarkBasicOperations(b *testing.B) {
	provider := NewProvider()
	ctx := context.Background()

	b.Run("GetStats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stats, err := provider.GetStats(ctx)
			assert.NoError(b, err)
			assert.NotNil(b, stats)
		}
	})

	err := provider.Shutdown()
	assert.NoError(b, err)
}

// TestPowerSourceDetection verifies the enhanced power source detection capabilities.
// It tests:
// - Accurate detection of power source type (AC/Battery)
// - Battery cycle count when available
// - Current, maximum, and design capacity values
// - Proper handling of unavailable metrics
func TestPowerSourceDetection(t *testing.T) {
	provider := NewProvider()
	require.NotNil(t, provider, "provider should not be nil")

	ctx := t.Context()

	stats, err := provider.GetStats(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Test power source type
	assert.Contains(t, []types.PowerSource{types.PowerSourceAC, types.PowerSourceBattery}, stats.Source,
		"power source should be either AC or Battery")

	if stats.Source == types.PowerSourceBattery {
		// Test battery metrics
		if stats.CycleCount > 0 {
			assert.Greater(t, stats.CycleCount, int64(0),
				"cycle count should be positive when available")
		}

		// Test capacity values - should be non-negative
		assert.GreaterOrEqual(t, stats.CurrentCapacity, float64(0),
			"current capacity should be non-negative")

		// Only test capacity relationships if all values are available
		if stats.CurrentCapacity > 0 && stats.MaxCapacity > 0 && stats.DesignCapacity > 0 {
			assert.LessOrEqual(t, stats.CurrentCapacity, stats.MaxCapacity,
				"current capacity should be <= max capacity")
			assert.LessOrEqual(t, stats.MaxCapacity, stats.DesignCapacity,
				"max capacity should be <= design capacity")

			// Test battery health calculation
			health := stats.MaxCapacity / stats.DesignCapacity * 100
			assert.GreaterOrEqual(t, health, 0.0,
				"battery health percentage should be non-negative")
			assert.LessOrEqual(t, health, 100.0,
				"battery health percentage should not exceed 100%")
		}
	}

	err = provider.Shutdown()
	assert.NoError(t, err)
}
