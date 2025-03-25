// Package types provides type definitions for CPU metrics collection.
package types

import "time"

// CPUStats represents detailed CPU statistics.
// All percentage values are normalized to range [0.0, 100.0].
// Time-based fields are calculated over the interval between measurements.
type CPUStats struct {
	// User is the percentage of CPU time spent in user space
	User float64

	// System is the percentage of CPU time spent in kernel space
	System float64

	// Idle is the percentage of CPU time spent idle
	Idle float64

	// Nice is the percentage of CPU time spent on low priority processes
	Nice float64

	// FrequencyMHz is the current CPU frequency in MHz
	// Note: May return 0 without elevated permissions
	FrequencyMHz uint64

	// PerfFrequencyMHz is the performance cores frequency (Apple Silicon)
	// Returns 0 on Intel processors
	PerfFrequencyMHz uint64

	// EffiFrequencyMHz is the efficiency cores frequency (Apple Silicon)
	// Returns 0 on Intel processors
	EffiFrequencyMHz uint64

	// PhysicalCores is the number of physical CPU cores
	PhysicalCores int

	// PerformanceCores is the number of performance cores (Apple Silicon)
	// Returns 0 on Intel processors
	PerformanceCores int

	// EfficiencyCores is the number of efficiency cores (Apple Silicon)
	// Returns 0 on Intel processors
	EfficiencyCores int

	// CoreUsage contains per-core CPU usage percentages
	// Index corresponds to logical core number
	// Values are normalized to [0.0, 100.0]
	CoreUsage []float64

	// TotalUsage is the total CPU usage percentage across all cores
	// Normalized to [0.0, 100.0]
	TotalUsage float64

	// LoadAvg contains load averages for 1, 5, and 15 minutes
	// Each value represents the average system load over the period
	// where 1.0 means full utilisation of one core
	LoadAvg [3]float64

	// Timestamp records when these stats were collected
	// Used for calculating deltas between measurements
	Timestamp time.Time
}

// CPUPlatform represents CPU platform information.
// This structure provides static information about the CPU
// that doesn't change during runtime.
type CPUPlatform struct {
	// IsAppleSilicon indicates whether this is an Apple Silicon Mac
	// Used to determine availability of architecture-specific features
	IsAppleSilicon bool

	// BrandString contains the CPU model name and details
	// Example: "Apple M1 Pro" or "Intel(R) Core(TM) i9-9900K CPU @ 3.60GHz"
	BrandString string

	// FrequencyMHz is the base/max CPU frequency in MHz
	// Note: May return 0 without elevated permissions
	FrequencyMHz uint64

	// Apple Silicon specific fields
	// These fields are only populated when IsAppleSilicon is true

	// PerfFrequencyMHz is the performance cores frequency
	PerfFrequencyMHz uint64

	// EffiFrequencyMHz is the efficiency cores frequency
	EffiFrequencyMHz uint64

	// PerformanceCores is the number of performance cores
	PerformanceCores int

	// EfficiencyCores is the number of efficiency cores
	EfficiencyCores int
}
