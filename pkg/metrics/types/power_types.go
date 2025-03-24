// Package types provides type definitions for power metrics collection.
package types

import (
	"errors"
	"time"
)

// Common errors for power metrics.
var (
	// ErrInvalidInterval is returned when a non-positive interval is provided for monitoring.
	ErrInvalidInterval = errors.New("interval must be positive")

	// ErrNoBattery is returned when attempting to get battery metrics on a system without a battery.
	ErrNoBattery = errors.New("no battery present")

	// ErrIOKitFailure is returned when IOKit operations fail.
	ErrIOKitFailure = errors.New("IOKit operation failed")
)

// PowerSource represents the current power source type.
type PowerSource string

const (
	// PowerSourceAC indicates the device is running on AC power.
	PowerSourceAC PowerSource = "AC"
	// PowerSourceBattery indicates the device is running on battery power.
	PowerSourceBattery PowerSource = "Battery"
	// PowerSourceUnknown indicates the power source could not be determined.
	PowerSourceUnknown PowerSource = "Unknown"
)

// BatteryState represents the current battery charging state.
type BatteryState string

const (
	// BatteryStateCharging indicates the battery is being charged.
	BatteryStateCharging BatteryState = "Charging"
	// BatteryStateDischarging indicates the battery is being discharged.
	BatteryStateDischarging BatteryState = "Discharging"
	// BatteryStateFull indicates the battery is fully charged.
	BatteryStateFull BatteryState = "Full"
	// BatteryStateNotCharging indicates the battery is not being charged despite AC power.
	BatteryStateNotCharging BatteryState = "NotCharging"
	// BatteryStateUnknown indicates the battery state could not be determined.
	BatteryStateUnknown BatteryState = "Unknown"
)

// BatteryHealth represents the battery's current health status.
type BatteryHealth string

const (
	// BatteryHealthGood indicates the battery is functioning normally.
	BatteryHealthGood BatteryHealth = "Good"
	// BatteryHealthPoor indicates the battery's capacity is significantly reduced.
	BatteryHealthPoor BatteryHealth = "Poor"
	// BatteryHealthReplacement indicates the battery needs replacement.
	BatteryHealthReplacement BatteryHealth = "Replacement"
	// BatteryHealthUnknown indicates the battery health could not be determined.
	BatteryHealthUnknown BatteryHealth = "Unknown"
	// BatteryHealthFair indicates the battery is functioning normally but has a reduced capacity.
	BatteryHealthFair BatteryHealth = "Fair"
)

// PowerStats represents detailed power and battery statistics.
type PowerStats struct {
	// Current power source (AC or Battery)
	Source PowerSource

	// Battery information
	IsPresent       bool          // Whether a battery is present
	State           BatteryState  // Current charging state
	Health          BatteryHealth // Battery health status
	Percentage      float64       // Current charge percentage (0-100)
	TimeRemaining   time.Duration // Estimated time remaining (when discharging)
	TimeToFull      time.Duration // Estimated time until full charge (when charging)
	CycleCount      int           // Number of charge cycles
	Temperature     float64       // Battery temperature in Celsius
	Voltage         float64       // Current voltage in Volts
	Amperage        float64       // Current amperage in Amperes
	Power           float64       // Current power draw in Watts
	DesignCapacity  float64       // Design capacity in Watt-hours
	MaxCapacity     float64       // Current maximum capacity in Watt-hours
	CurrentCapacity float64       // Current capacity in Watt-hours
	DesignCycles    int           // Design cycle count limit

	// System power information
	CPUPower   float64   // CPU power consumption in Watts
	GPUPower   float64   // GPU power consumption in Watts
	TotalPower float64   // Total system power consumption in Watts
	Timestamp  time.Time // Time when these stats were collected
}
