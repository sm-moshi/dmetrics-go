//go:build darwin
// +build darwin

// Package cpu provides CPU statistics for macOS systems.
package cpu

import (
	"github.com/sm-moshi/dmetrics-go/cpu/internal"
	"github.com/sm-moshi/dmetrics-go/cpu/types"
)

// Stats is an alias for types.Stats.
type Stats = types.Stats

// Get returns current CPU statistics.
func Get() (*Stats, error) {
	return internal.GetStats()
}

// Usage returns the current total CPU usage percentage (0-100).
func Usage() (float64, error) {
	return internal.Usage()
}

// Frequency returns the current CPU frequency in MHz.
func Frequency() (uint64, error) {
	return internal.GetFrequency()
}

// IsAppleSilicon returns true if running on Apple Silicon.
func IsAppleSilicon() (bool, error) {
	return internal.IsAppleSilicon()
}

// LoadAverage returns the system load averages for the past 1, 5, and 15 minutes.
func LoadAverage() ([3]float64, error) {
	return internal.GetLoadAvg()
}
