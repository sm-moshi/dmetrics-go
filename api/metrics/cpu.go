// Package metrics provides the public interfaces for system metrics collection.
package metrics

import (
	"context"
	"time"

	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

// CPUMetrics defines the interface for CPU metrics collection.
type CPUMetrics interface {
	// GetFrequency returns the current CPU frequency in MHz.
	// It returns an error if the frequency cannot be determined.
	GetFrequency(ctx context.Context) (uint64, error)

	// GetUsage returns the current CPU usage as a percentage (0-100).
	// The interval parameter determines the sampling period.
	// It returns an error if the usage cannot be determined.
	GetUsage(ctx context.Context, interval time.Duration) (float64, error)

	// GetCoreCount returns the number of CPU cores available.
	// It returns an error if the core count cannot be determined.
	GetCoreCount(ctx context.Context) (int, error)

	// GetStats returns detailed CPU statistics.
	// It returns an error if the statistics cannot be determined.
	GetStats(ctx context.Context) (*types.CPUStats, error)

	// Watch starts monitoring CPU metrics and sends updates through the returned channel.
	// The interval parameter determines how often updates are sent.
	// The returned channel will be closed when the context is cancelled.
	Watch(ctx context.Context, interval time.Duration) (<-chan types.CPUStats, error)
}

// CPUStats represents a snapshot of CPU statistics.
type CPUStats struct {
	// Timestamp when the stats were collected
	Timestamp time.Time

	// Frequency in MHz
	Frequency float64

	// Usage as percentage (0-100)
	Usage float64

	// Error if any occurred during collection
	Error error
}
