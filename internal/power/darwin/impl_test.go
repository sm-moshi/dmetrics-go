//go:build darwin
// +build darwin

package darwin

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

func TestGetStatsImpl(t *testing.T) {
	stats, err := getStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Validate struct fields
	assert.GreaterOrEqual(t, stats.Voltage, 0.0)
	assert.GreaterOrEqual(t, stats.Amperage, -100.0) // Can be negative when discharging
	assert.GreaterOrEqual(t, stats.Power, 0.0)
	assert.GreaterOrEqual(t, stats.DesignCapacity, 0.0)
	assert.GreaterOrEqual(t, stats.MaxCapacity, 0.0)
	assert.GreaterOrEqual(t, stats.CycleCount, 0)
	assert.NotZero(t, stats.Timestamp)
}

func TestGetPowerSourceImpl(t *testing.T) {
	source, err := getPowerSource()
	require.NoError(t, err)
	assert.NotEqual(t, types.PowerSourceUnknown, source, "Power source should be determinable")
}

func TestGetBatteryPercentageImpl(t *testing.T) {
	percentage, err := getBatteryPercentage()
	if errors.Is(err, types.ErrNoBattery) {
		t.Skip("No battery present, skipping test")
	}
	require.NoError(t, err)
	assert.GreaterOrEqual(t, percentage, 0.0)
	assert.LessOrEqual(t, percentage, 100.0)
}

func TestGetBatteryStateImpl(t *testing.T) {
	state, err := getBatteryState()
	require.NoError(t, err)

	// State should be one of the defined states
	validStates := []types.BatteryState{
		types.BatteryStateCharging,
		types.BatteryStateDischarging,
		types.BatteryStateFull,
		types.BatteryStateUnknown,
	}
	assert.Contains(t, validStates, state)
}

func TestGetBatteryHealthImpl(t *testing.T) {
	health, err := getBatteryHealth()
	if errors.Is(err, types.ErrNoBattery) {
		t.Skip("No battery present, skipping test")
	}
	require.NoError(t, err)

	// Validate health calculation
	validHealth := []types.BatteryHealth{
		types.BatteryHealthGood,
		types.BatteryHealthFair,
		types.BatteryHealthPoor,
		types.BatteryHealthUnknown,
	}
	assert.Contains(t, validHealth, health)
}

func TestGetTimeRemainingImpl(t *testing.T) {
	stats, err := getStats()
	if errors.Is(err, types.ErrNoBattery) {
		t.Skip("No battery present, skipping test")
	}
	require.NoError(t, err)

	// Only verify time remaining if we have a battery
	if stats.IsPresent {
		timeRemaining, err := getTimeRemaining()
		require.NoError(t, err)

		// Time remaining can be negative when charging
		if stats.State == types.BatteryStateCharging {
			assert.LessOrEqual(t, timeRemaining, time.Duration(0),
				"charging time should be negative")
		} else {
			assert.GreaterOrEqual(t, timeRemaining, time.Duration(0),
				"discharging time should be non-negative")
		}
	}
}

func TestGetPowerConsumptionImpl(t *testing.T) {
	power, err := getPowerConsumption()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, power, 0.0)
}

func TestSMCInitialization(t *testing.T) {
	// Test SMC initialization (already done in init())
	// Just verify we can get power info
	stats, err := getStats()
	require.NoError(t, err)
	assert.NotNil(t, stats)

	// Verify power consumption data is available
	assert.GreaterOrEqual(t, stats.CPUPower, 0.0)
	assert.GreaterOrEqual(t, stats.GPUPower, 0.0)
	assert.GreaterOrEqual(t, stats.TotalPower, 0.0)
}

func TestBatteryHealthCalculation(t *testing.T) {
	stats, err := getStats()
	if errors.Is(err, types.ErrNoBattery) {
		t.Skip("No battery present, skipping test")
	}
	require.NoError(t, err)

	// Only verify capacity values if we have a battery
	if stats.IsPresent {
		// Verify that we have valid capacity values
		assert.GreaterOrEqual(t, stats.DesignCapacity, 0.0, "design capacity should be non-negative")
		assert.GreaterOrEqual(t, stats.MaxCapacity, 0.0, "max capacity should be non-negative")

		// Skip health calculation if we don't have valid capacity values
		if stats.DesignCapacity <= 0 || stats.MaxCapacity <= 0 {
			t.Skip("Invalid capacity values, skipping health calculation")
		}

		// Test health calculation
		health, err := getBatteryHealth()
		require.NoError(t, err)

		// Health should correlate with capacity ratio
		healthPercent := (stats.MaxCapacity / stats.DesignCapacity) * 100
		t.Logf("Health percent: %.2f%%", healthPercent)

		// Verify that we got a valid health status
		assert.NotEqual(t, types.BatteryHealthUnknown, health, "health status should not be unknown")

		// Verify that the health status matches the percentage
		switch {
		case healthPercent >= 80:
			assert.Equal(t, types.BatteryHealthGood, health)
		case healthPercent >= 50:
			assert.Equal(t, types.BatteryHealthFair, health)
		default:
			assert.Equal(t, types.BatteryHealthPoor, health)
		}
	}
}
