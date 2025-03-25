// Package darwin provides CPU metrics collection for macOS systems.
// It supports both Intel and Apple Silicon processors, providing detailed
// information about CPU frequency, usage, core counts, and load averages.
// The implementation uses native macOS APIs through cgo.
package darwin

/*
#cgo CFLAGS: -x objective-c -I/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include
#cgo LDFLAGS: -F/System/Library/Frameworks -framework CoreFoundation

#include <stdlib.h>
#include <mach/mach_host.h>
#include <mach/processor_info.h>
#include "cpu.h"
*/
import "C"

import (
	"fmt"
	"time"

	"github.com/sm-moshi/dmetrics-go/api/metrics"
	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

const (
	maxCPUPercentage = 100.0
)

// Error codes from C implementation for CPU metrics collection.
const (
	errCPUSuccess           = 0
	errCPUHostProcessorInfo = -1
	errCPUSysctl            = -2
	errCPUMemory            = -3
)

// cpuError converts a C error code to a Go error with appropriate context.
func cpuError(code int) error {
	switch code {
	case errCPUSuccess:
		return nil
	case errCPUHostProcessorInfo:
		return fmt.Errorf("%w: failed to get host processor information", metrics.ErrHardwareAccess)
	case errCPUSysctl:
		return fmt.Errorf("%w: sysctl operation failed", metrics.ErrHardwareAccess)
	case errCPUMemory:
		return fmt.Errorf("%w: memory allocation failed", metrics.ErrHardwareAccess)
	default:
		return fmt.Errorf("%w: unknown error code %d", metrics.ErrHardwareAccess, code)
	}
}

// getStats returns current CPU statistics including usage, frequency, and core information.
// For Apple Silicon Macs, this includes both performance and efficiency core metrics.
// Returns metrics.ErrHardwareAccess if hardware information cannot be accessed.
func getStats() (*types.CPUStats, error) {
	numCPUs := int(C.get_cpu_count())
	if numCPUs <= 0 {
		return nil, fmt.Errorf("%w: failed to get CPU count", metrics.ErrHardwareAccess)
	}

	var cStats C.cpu_stats_t
	if err := int(C.get_cpu_stats(&cStats)); err != errCPUSuccess {
		return nil, cpuError(err)
	}

	// Get platform info
	var platform C.cpu_platform_t
	if err := int(C.get_cpu_platform(&platform)); err != errCPUSuccess {
		return nil, cpuError(err)
	}

	// Get frequencies
	perfFreq := uint64(C.get_perf_core_freq())
	effiFreq := uint64(C.get_effi_core_freq())
	baseFreq := perfFreq
	if baseFreq == 0 {
		baseFreq = effiFreq
	}

	// Get core counts
	perfCores := int(C.get_perf_core_count())
	effiCores := int(C.get_effi_core_count())
	totalCores := perfCores + effiCores
	if totalCores == 0 {
		totalCores = numCPUs
	}

	var loadAvg [3]float64
	if err := int(C.get_load_avg((*C.double)(&loadAvg[0]))); err != errCPUSuccess {
		return nil, cpuError(err)
	}

	return &types.CPUStats{
		User:             float64(cStats.user),
		System:           float64(cStats.system),
		Idle:             float64(cStats.idle),
		Nice:             float64(cStats.nice),
		FrequencyMHz:     baseFreq,
		PerfFrequencyMHz: perfFreq,
		EffiFrequencyMHz: effiFreq,
		PhysicalCores:    totalCores,
		PerformanceCores: perfCores,
		EfficiencyCores:  effiCores,
		TotalUsage:       maxCPUPercentage - float64(cStats.idle),
		LoadAvg:          loadAvg,
		Timestamp:        time.Now(),
	}, nil
}

// getFrequency returns the current CPU frequency in MHz.
// For Apple Silicon Macs, this returns the highest frequency among all cores.
// Returns metrics.ErrHardwareAccess if the frequency cannot be determined.
func getFrequency() (uint64, error) {
	// Try performance cores first
	if freq := uint64(C.get_perf_core_freq()); freq > 0 {
		return freq, nil
	}

	// Try efficiency cores next
	if freq := uint64(C.get_effi_core_freq()); freq > 0 {
		return freq, nil
	}

	// Fall back to traditional method
	freq := uint64(C.get_cpu_freq())
	if freq == 0 {
		return 0, fmt.Errorf("%w: failed to detect CPU frequency", metrics.ErrHardwareAccess)
	}
	return freq, nil
}

// usage returns current CPU usage as a percentage (0-100).
// Returns metrics.ErrHardwareAccess if usage cannot be determined.
func usage() (float64, error) {
	stats, err := getStats()
	if err != nil {
		return 0, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	return stats.TotalUsage, nil
}

// getLoadAvg returns the system load averages for the past 1, 5, and 15 minutes.
// Returns metrics.ErrHardwareAccess if load averages cannot be determined.
func getLoadAvg() ([3]float64, error) {
	var loadAvg [3]float64
	if err := int(C.get_load_avg((*C.double)(&loadAvg[0]))); err != errCPUSuccess {
		return [3]float64{}, cpuError(err)
	}
	return loadAvg, nil
}

// cleanup releases any resources used by the CPU stats collector.
// This function is safe to call multiple times and should be called
// when the CPU metrics are no longer needed.
func cleanup() {
	C.cleanup_cpu_stats()
}
