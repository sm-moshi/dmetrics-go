//go:build darwin
// +build darwin

// Package power provides platform-specific power metrics collection.
package power

import (
	"github.com/sm-moshi/dmetrics-go/api/metrics"
	"github.com/sm-moshi/dmetrics-go/internal/power/darwin"
)

// NewProvider creates a new power metrics provider for the current platform.
// On Darwin systems, this returns a provider that uses IOKit for power metrics.
func NewProvider() metrics.PowerMetrics {
	return darwin.NewProvider()
}
