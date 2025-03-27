//go:build darwin
// +build darwin

// Package darwin provides Darwin-specific CPU metrics implementation.
package darwin

/*
#cgo CFLAGS: -x objective-c -I/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include
#cgo LDFLAGS: -F/System/Library/Frameworks -framework CoreFoundation

#include "cpu.h"
#include <stdlib.h>

// Forward declarations of functions used in Go
int get_cpu_count(void);
uint64_t get_cpu_freq(void);
uint64_t get_perf_core_freq(void);
uint64_t get_effi_core_freq(void);
int get_perf_core_count(void);
int get_effi_core_count(void);
int get_cpu_stats(cpu_stats_t *stats);
int get_cpu_platform(cpu_platform_t *platform);
int get_load_avg(double loadavg[3]);
void init_cpu_stats(void);
void cleanup_cpu_stats(void);
int get_cpu_core_stats(cpu_core_stats_t *stats, int *num_cores);
*/
import "C"

import (
	"context"
	"time"
	"unsafe"

	"github.com/sm-moshi/dmetrics-go/api/metrics"
	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

const (
	brandStringLength = 128 // Length of the CPU brand string buffer
)

// Provider implements the CPU metrics collection for Darwin systems.
// It provides thread-safe access to CPU statistics, frequency information,
// and core configuration details. The provider supports both Intel and
// Apple Silicon architectures, with additional methods specific to
// Apple Silicon systems.
//
// The provider maintains minimal state and is safe for concurrent use.
// All methods are thread-safe and can be called from multiple goroutines.
// Resource cleanup is handled automatically through the Shutdown method.
type Provider struct {
	// Provider is stateless and uses system calls directly
}

// NewProvider creates a new Darwin CPU metrics provider.
func NewProvider() *Provider {
	initCleanup()
	return &Provider{}
}

// GetStats returns current CPU statistics.
func (p *Provider) GetStats(context.Context) (*types.CPUStats, error) {
	stats, err := getStats()
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetUsage returns the current total CPU usage percentage (0-100).
// The interval parameter determines the sampling period for calculating usage.
func (p *Provider) GetUsage(interval time.Duration) (float64, error) {
	// Create a timer for the interval
	timer := time.NewTimer(interval)
	defer timer.Stop()

	// Get initial usage
	initial, err := usage()
	if err != nil {
		return 0, err
	}

	// Wait for interval completion
	<-timer.C
	final, err := usage()
	if err != nil {
		return 0, err
	}
	// Return the difference in usage over the interval
	return final - initial, nil
}

// GetFrequency returns the current CPU frequency in MHz.
func (p *Provider) GetFrequency() (uint64, error) {
	return getFrequency()
}

// GetCoreCount returns the number of CPU cores.
func (p *Provider) GetCoreCount() (int, error) {
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	return stats.PhysicalCores, nil
}

// GetEfficiencyCoreCount returns the number of efficiency cores on Apple Silicon.
// Returns 0 on Intel processors.
func (p *Provider) GetEfficiencyCoreCount() (int, error) {
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	return stats.EfficiencyCores, nil
}

// GetPerformanceCoreCount returns the number of performance cores on Apple Silicon.
// Returns 0 on Intel processors.
func (p *Provider) GetPerformanceCoreCount() (int, error) {
	stats, err := getStats()
	if err != nil {
		return 0, err
	}
	return stats.PerformanceCores, nil
}

// GetPerformanceFrequency returns the current frequency of performance cores in MHz.
// This method is only applicable to Apple Silicon Macs.
func (p *Provider) GetPerformanceFrequency() (uint64, error) {
	freq := uint64(C.get_perf_core_freq())
	if freq == 0 {
		return 0, metrics.ErrUnsupportedPlatform
	}
	return freq, nil
}

// GetEfficiencyFrequency returns the current frequency of efficiency cores in MHz.
// This method is only applicable to Apple Silicon Macs.
func (p *Provider) GetEfficiencyFrequency() (uint64, error) {
	freq := uint64(C.get_effi_core_freq())
	if freq == 0 {
		return 0, metrics.ErrUnsupportedPlatform
	}
	return freq, nil
}

// GetPlatform returns information about the CPU platform.
func (p *Provider) GetPlatform() (*types.CPUPlatform, error) {
	var cPlatform C.cpu_platform_t
	if err := int(C.get_cpu_platform(&cPlatform)); err != errCPUSuccess {
		return nil, cpuError(err)
	}

	return &types.CPUPlatform{
		IsAppleSilicon:   cPlatform.is_apple_silicon != 0,
		BrandString:      C.GoStringN((*C.char)(unsafe.Pointer(&cPlatform.brand_string[0])), brandStringLength),
		FrequencyMHz:     uint64(cPlatform.frequency),
		PerfFrequencyMHz: uint64(cPlatform.perf_freq),
		EffiFrequencyMHz: uint64(cPlatform.effi_freq),
		PerformanceCores: int(cPlatform.perf_cores),
		EfficiencyCores:  int(cPlatform.effi_cores),
	}, nil
}

// validateWatchParams validates the interval parameter for Watch.
func (p *Provider) validateWatchParams(interval time.Duration) error {
	if interval <= 0 {
		return types.ErrInvalidInterval
	}
	return nil
}

// createStatsChannel creates a buffered channel for CPU stats.
func (p *Provider) createStatsChannel() chan *types.CPUStats {
	return make(chan *types.CPUStats, 1)
}

// sendStatsWithDropping sends stats to the channel, dropping old values if needed.
func (p *Provider) sendStatsWithDropping(ch chan *types.CPUStats, stats *types.CPUStats) {
	select {
	case ch <- stats:
		return
	default:
		// Channel is full, drop old value and send new
		select {
		case <-ch:
		default:
		}
		// Try to send again
		select {
		case ch <- stats:
		default:
		}
	}
}

// collectAndSendStats collects CPU stats and sends them to the channel.
func (p *Provider) collectAndSendStats(ctx context.Context, ch chan *types.CPUStats) {
	stats, err := p.GetStats(ctx)
	if err != nil {
		return // Error is already logged in GetStats
	}
	p.sendStatsWithDropping(ch, stats)
}

// runWatchLoop runs the main monitoring loop.
func (p *Provider) runWatchLoop(ctx context.Context, interval time.Duration, ch chan *types.CPUStats) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.collectAndSendStats(ctx, ch)
		}
	}
}

// Watch monitors CPU statistics and sends updates through the returned channel.
// The interval parameter determines how often updates are sent.
// The context can be used to stop monitoring. When the context is cancelled,
// the channel will be closed after any pending updates are sent.
//
// The returned channel is buffered with a capacity of 1 to prevent blocking
// when the consumer is slower than the update interval. If a consumer cannot
// keep up with updates, the oldest update will be dropped in favor of the newest.
//
// If an error occurs during monitoring, the error will be logged and the
// channel will be closed. The caller should always ensure proper cleanup by
// cancelling the context when monitoring is no longer needed.
func (p *Provider) Watch(ctx context.Context, interval time.Duration) (<-chan *types.CPUStats, error) {
	if err := p.validateWatchParams(interval); err != nil {
		return nil, err
	}

	ch := p.createStatsChannel()
	go func() {
		defer close(ch)
		p.runWatchLoop(ctx, interval, ch)
	}()

	return ch, nil
}

// Shutdown cleans up resources used by the provider.
func (p *Provider) Shutdown() error {
	cleanup()
	return nil
}
