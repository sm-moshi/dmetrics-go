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
		IsPresent:       bool(cStats.is_present),
		Percentage:      float64(cStats.percentage),
		State:           determineBatteryState(cStats),
		Source:          determinePowerSource(cStats),
		Timestamp:       time.Now(),
		TimeRemaining:   determineTimeRemaining(cStats),
		CycleCount:      int(cStats.cycle_count),
		CurrentCapacity: float64(cStats.current_capacity),
		MaxCapacity:     float64(cStats.max_capacity),
		DesignCapacity:  float64(cStats.design_capacity),
		Health:          determineBatteryHealth(cStats),
	}

	return stats, nil
}

// determineBatteryState determines the battery state based on power source info
func determineBatteryState(stats C.power_stats_t) types.BatteryState {
	if !bool(stats.is_present) {
		return types.BatteryStateUnknown
	}
	if bool(stats.is_charged) {
		return types.BatteryStateFull
	}
	if bool(stats.is_charging) {
		return types.BatteryStateCharging
	}
	if bool(stats.is_ac_present) {
		return types.BatteryStateNotCharging
	}
	return types.BatteryStateDischarging
}

// determinePowerSource determines the current power source
func determinePowerSource(stats C.power_stats_t) types.PowerSource {
	if bool(stats.is_ac_present) {
		return types.PowerSourceAC
	}
	if bool(stats.is_present) {
		return types.PowerSourceBattery
	}
	return types.PowerSourceUnknown
}

// determineTimeRemaining converts the time remaining value
func determineTimeRemaining(stats C.power_stats_t) time.Duration {
	// Negative values indicate charging time
	minutes := float64(stats.time_remaining)
	return time.Duration(minutes) * time.Minute
}

// determineBatteryHealth calculates battery health based on capacity values
func determineBatteryHealth(stats C.power_stats_t) types.BatteryHealth {
	if !bool(stats.is_present) {
		return types.BatteryHealthUnknown
	}

	// Calculate health percentage based on current max capacity vs design capacity
	maxCapacity := float64(stats.max_capacity)
	designCapacity := float64(stats.design_capacity)

	if maxCapacity <= 0 || designCapacity <= 0 {
		return types.BatteryHealthUnknown
	}

	healthPercent := (maxCapacity / designCapacity) * batteryHealthPercentMultiplier

	switch {
	case healthPercent >= batteryHealthGoodThreshold:
		return types.BatteryHealthGood
	case healthPercent >= batteryHealthFairThreshold:
		return types.BatteryHealthFair
	default:
		return types.BatteryHealthPoor
	}
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
	return determineBatteryHealth(C.power_stats_t{
		is_present:      C._Bool(stats.IsPresent),
		max_capacity:    C.double(stats.MaxCapacity),
		design_capacity: C.double(stats.DesignCapacity),
	}), nil
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
