//go:build darwin
// +build darwin

// Package darwin provides power management functionality for Darwin/macOS systems.
package darwin

/*
#cgo CFLAGS: -x objective-c -I${SRCDIR}
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include "power.h"
*/
import "C"

import (
	"errors"
	"time"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

// batteryHealthPercentMultiplier converts capacity ratios to percentages.
const batteryHealthPercentMultiplier = 100.0

// batteryHealthGoodThreshold defines the minimum percentage for good health status.
const batteryHealthGoodThreshold = 80.0

// batteryHealthFairThreshold defines the minimum percentage for fair health status.
const batteryHealthFairThreshold = 50.0

// getStats retrieves basic power and battery statistics.
// For v0.1, we focus on essential metrics using IOPowerSources API.
func getStats() (*types.PowerStats, error) {
	var cStats C.power_stats_t
	if ok := C.get_power_source_info(&cStats); !ok {
		return nil, errors.New("failed to get power source info")
	}

	stats := &types.PowerStats{
		IsPresent:  bool(cStats.is_present),
		Percentage: float64(cStats.percentage),
		Timestamp:  time.Now(),
		Health:     types.BatteryHealthUnknown, // Health calculation requires more data
	}

	// Set battery state
	if bool(cStats.is_charging) {
		stats.State = types.BatteryStateCharging
	} else if bool(cStats.is_present) {
		stats.State = types.BatteryStateDischarging
	} else {
		stats.State = types.BatteryStateUnknown
	}

	// Set power source
	if !cStats.is_present {
		stats.Source = types.PowerSourceAC
	} else {
		stats.Source = types.PowerSourceBattery
	}

	// Initialize other fields with zero values
	// These will be implemented in future versions
	stats.TimeRemaining = 0
	stats.TimeToFull = 0
	stats.CycleCount = 0
	stats.Temperature = 0
	stats.Voltage = 0
	stats.Amperage = 0
	stats.Power = 0
	stats.DesignCapacity = 0
	stats.MaxCapacity = 0
	stats.CurrentCapacity = 0
	stats.DesignCycles = 0
	stats.CPUPower = 0
	stats.GPUPower = 0
	stats.TotalPower = 0

	return stats, nil
}

// Helper functions that use getStats().
func getPowerSource() (types.PowerSource, error) {
	stats, err := getStats()
	if err != nil {
		return types.PowerSourceUnknown, err
	}
	return stats.Source, nil
}

func getBatteryPercentage() (float64, error) {
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	if !stats.IsPresent {
		return 0, types.ErrNoBattery
	}
	return stats.Percentage, nil
}

func getBatteryState() (types.BatteryState, error) {
	stats, err := getStats()
	if err != nil {
		return types.BatteryStateUnknown, err
	}
	return stats.State, nil
}

func getBatteryHealth() (types.BatteryHealth, error) {
	stats, err := getStats()
	if err != nil {
		return types.BatteryHealthUnknown, err
	}
	if !stats.IsPresent {
		return types.BatteryHealthUnknown, types.ErrNoBattery
	}
	return types.BatteryHealthUnknown, nil // Health calculation requires more data for v0.1
}

func getTimeRemaining() (time.Duration, error) {
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	if !stats.IsPresent {
		return 0, types.ErrNoBattery
	}
	return stats.TimeRemaining, nil
}

func getPowerConsumption() (float64, error) {
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	return stats.TotalPower, nil
}
