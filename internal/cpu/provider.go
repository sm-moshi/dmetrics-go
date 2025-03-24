//go:build darwin
// +build darwin

// Package cpu provides CPU metrics collection functionality.
package cpu

import (
	"github.com/sm-moshi/dmetrics-go/api/metrics"
	"github.com/sm-moshi/dmetrics-go/internal/cpu/darwin"
)

// NewProvider creates a new CPU metrics provider for the current platform.
func NewProvider() metrics.CPUMetrics {
	return darwin.NewProvider()
}
