// Package types provides type definitions for CPU metrics collection.
package types

import "time"

// CPUStats represents detailed CPU statistics.
type CPUStats struct {
	User             float64    // User CPU time percentage
	System           float64    // System CPU time percentage
	Idle             float64    // Idle CPU time percentage
	Nice             float64    // Nice CPU time percentage
	FrequencyMHz     uint64     // CPU frequency in MHz
	PerfFrequencyMHz uint64     // Performance cores frequency (Apple Silicon)
	EffiFrequencyMHz uint64     // Efficiency cores frequency (Apple Silicon)
	PhysicalCores    int        // Number of physical CPU cores
	PerformanceCores int        // Number of performance cores (Apple Silicon)
	EfficiencyCores  int        // Number of efficiency cores (Apple Silicon)
	CoreUsage        []float64  // Per-core CPU usage percentages
	TotalUsage       float64    // Total CPU usage percentage
	LoadAvg          [3]float64 // Load averages for 1, 5, and 15 minutes
	Timestamp        time.Time  // Time when the stats were collected
}

// CPUPlatform represents CPU platform information.
type CPUPlatform struct {
	IsAppleSilicon bool   // Whether this is an Apple Silicon Mac
	BrandString    string // CPU brand string
	FrequencyMHz   uint64 // Base/Max CPU frequency in MHz

	// Apple Silicon specific fields
	PerfFrequencyMHz uint64 // Performance cores frequency
	EffiFrequencyMHz uint64 // Efficiency cores frequency
	PerformanceCores int    // Number of performance cores
	EfficiencyCores  int    // Number of efficiency cores
}
