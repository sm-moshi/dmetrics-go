//go:build darwin
// +build darwin

package darwin

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/ps/IOPSKeys.h>
#include <IOKit/ps/IOPowerSources.h>
#include <stdbool.h>
#include <stdint.h>
#include "power.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

const (
	// batteryHealthPercentMultiplier is used to convert capacity ratio to percentage.
	batteryHealthPercentMultiplier = 100

	// batteryHealthGoodThreshold is the minimum percentage for good health status.
	batteryHealthGoodThreshold = 80

	// batteryHealthFairThreshold is the minimum percentage for fair health status.
	batteryHealthFairThreshold = 50
)

var (
	smcMu          sync.RWMutex
	smcInitialized bool
)

// getStats returns the current power and battery statistics.
func getStats() (*types.PowerStats, error) {
	smcMu.RLock()
	defer smcMu.RUnlock()

	if !smcInitialized {
		return nil, errors.New("SMC not initialized")
	}

	var cStats C.power_stats_t
	if ok := C.get_power_source_info(&cStats); !ok {
		return nil, errors.New("failed to get power source info")
	}

	stats := &types.PowerStats{
		IsPresent:      bool(cStats.is_present),
		Percentage:     float64(cStats.percentage),
		Temperature:    float64(cStats.temperature),
		Voltage:        float64(cStats.voltage),
		Amperage:       float64(cStats.amperage),
		Power:          float64(cStats.power),
		DesignCapacity: float64(cStats.design_capacity),
		MaxCapacity:    float64(cStats.max_capacity),
		CycleCount:     int(cStats.cycle_count),
		TimeRemaining:  time.Duration(cStats.time_remaining) * time.Minute,
		TimeToFull:     time.Duration(cStats.time_to_full) * time.Minute,
		Timestamp:      time.Now(),
	}

	// Set battery state
	switch {
	case bool(cStats.is_charging):
		stats.State = types.BatteryStateCharging
	case bool(cStats.is_charged):
		stats.State = types.BatteryStateFull
	case bool(cStats.is_present):
		stats.State = types.BatteryStateDischarging
	default:
		stats.State = types.BatteryStateUnknown
	}

	// Set power source
	if !cStats.is_present {
		stats.Source = types.PowerSourceAC
	} else {
		stats.Source = types.PowerSourceBattery
	}

	// Get system power info
	var sysPower C.system_power_t
	if ok := C.get_system_power_info(&sysPower); ok {
		stats.CPUPower = float64(sysPower.cpu_power)
		stats.GPUPower = float64(sysPower.gpu_power)
		stats.TotalPower = float64(sysPower.total_power)
	}

	return stats, nil
}

// Initialise SMC on package init.
func init() {
	smcMu.Lock()
	defer smcMu.Unlock()

	if ok := C.init_smc(); !ok {
		// Log error but don't fail - some metrics may still work
		fmt.Printf("Warning: Failed to initialise SMC connection\n")
		return
	}
	smcInitialized = true
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

	// Calculate health based on current vs design capacity
	if !stats.IsPresent {
		return types.BatteryHealthUnknown, types.ErrNoBattery
	}

	healthPercent := (stats.MaxCapacity / stats.DesignCapacity) * batteryHealthPercentMultiplier
	switch {
	case healthPercent >= batteryHealthGoodThreshold:
		return types.BatteryHealthGood, nil
	case healthPercent >= batteryHealthFairThreshold:
		return types.BatteryHealthFair, nil
	default:
		return types.BatteryHealthPoor, nil
	}
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
