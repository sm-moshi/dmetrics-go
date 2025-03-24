package internal

/*
#cgo CFLAGS: -x objective-c -I/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include
#cgo LDFLAGS: -F/System/Library/Frameworks -framework CoreFoundation

#include <stdlib.h>
#include "cpu.h"

// Go-compatible type definitions
typedef struct {
	double user;
	double system;
	double idle;
	double nice;
} go_cpu_stats_t;

typedef struct {
	int is_apple_silicon;
	char brand_string[128];
	unsigned long long frequency;
} go_cpu_platform_t;

typedef struct {
	double user;
	double system;
	double idle;
	double nice;
	int core_id;
} go_cpu_core_stats_t;

// Function declarations
uint64_t get_cpu_freq(void);

// Wrapper functions to avoid macro expansion issues
static inline int go_get_cpu_count() {
	return get_cpu_count();
}

static inline uint64_t go_get_cpu_freq() {
	return get_cpu_freq();
}

static inline int go_get_cpu_stats(go_cpu_stats_t* stats) {
	return get_cpu_stats((cpu_stats_t*)stats);
}

static inline int go_get_load_avg(double loadavg[3]) {
	return get_load_avg(loadavg);
}

static inline int go_get_cpu_platform(go_cpu_platform_t* platform) {
	return get_cpu_platform((cpu_platform_t*)platform);
}

static inline void go_cleanup_cpu_stats() {
	cleanup_cpu_stats();
}

static inline int go_get_cpu_core_stats(go_cpu_core_stats_t* stats, int* num_cores) {
	return get_cpu_core_stats((cpu_core_stats_t*)stats, num_cores);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/sm-moshi/dmetrics-go/cpu/types"
)

const (
	maxCPUPercentage = 100.0
)

// platformInfo caches CPU platform information.
var platformInfo struct {
	sync.Once
	platform C.go_cpu_platform_t
	err      error
}

// GetStats returns current CPU statistics.
func GetStats() (*types.Stats, error) {
	numCPUs := int(C.get_cpu_count())
	cStats := make([]C.cpu_stats_t, numCPUs)

	if err := C.get_cpu_stats(&cStats[0]); err != C.CPU_SUCCESS {
		return nil, cpuError(err)
	}

	// Calculate per-core usage
	coreUsage := make([]float64, numCPUs)
	var totalUsage float64
	for i := 0; i < numCPUs; i++ {
		active := maxCPUPercentage - float64(cStats[i].idle)
		coreUsage[i] = active
		totalUsage += active
	}
	totalUsage /= float64(numCPUs)

	// Get platform info first
	var platform C.cpu_platform_t
	if err := C.get_cpu_platform(&platform); err != C.CPU_SUCCESS {
		return nil, cpuError(err)
	}

	// Get frequencies
	perfFreq := C.get_perf_core_freq()
	effiFreq := C.get_effi_core_freq()
	baseFreq := uint64(perfFreq)
	if baseFreq == 0 {
		baseFreq = uint64(effiFreq)
	}

	// Get core counts
	perfCores := int(C.get_perf_core_count())
	effiCores := int(C.get_effi_core_count())
	totalCores := perfCores + effiCores
	if totalCores == 0 {
		totalCores = int(C.get_cpu_count())
	}

	var loadAvg [3]C.double
	if err := C.get_load_avg(&loadAvg[0]); err != C.CPU_SUCCESS {
		return nil, cpuError(err)
	}

	return &types.Stats{
		User:             float64(cStats[0].user),
		System:           float64(cStats[0].system),
		Idle:             float64(cStats[0].idle),
		Nice:             float64(cStats[0].nice),
		FrequencyMHz:     baseFreq,
		PerfFrequencyMHz: uint64(perfFreq),
		EffiFrequencyMHz: uint64(effiFreq),
		PhysicalCores:    totalCores,
		PerformanceCores: perfCores,
		EfficiencyCores:  effiCores,
		TotalUsage:       totalUsage,
		LoadAvg:          [3]float64{float64(loadAvg[0]), float64(loadAvg[1]), float64(loadAvg[2])},
		Timestamp:        time.Now(),
		CoreUsage:        coreUsage,
	}, nil
}

// GetFrequency returns the current CPU frequency in MHz.
func GetFrequency() (uint64, error) {
	// Try performance cores first
	if freq := C.get_perf_core_freq(); freq > 0 {
		return uint64(freq), nil
	}

	// Try efficiency cores next
	if freq := C.get_effi_core_freq(); freq > 0 {
		return uint64(freq), nil
	}

	// Fall back to traditional method
	freq := C.get_cpu_freq()
	if freq == 0 {
		return 0, cpuError(C.CPU_ERROR_SYSCTL)
	}
	return uint64(freq), nil
}

// IsAppleSilicon returns true if running on Apple Silicon.
func IsAppleSilicon() (bool, error) {
	// Initialise platform info if not already done
	platformInfo.Once.Do(func() {
		var platform C.go_cpu_platform_t
		if ret := C.go_get_cpu_platform(&platform); ret != 0 {
			platformInfo.err = fmt.Errorf("failed to get CPU platform info: %d", ret)
			return
		}
		platformInfo.platform = platform
	})

	if platformInfo.err != nil {
		return false, platformInfo.err
	}

	return platformInfo.platform.is_apple_silicon != 0, nil
}

// Usage returns current CPU usage as a percentage.
func Usage() (float64, error) {
	stats, err := GetStats()
	if err != nil {
		return 0, err
	}

	return stats.TotalUsage, nil
}

// GetLoadAvg returns the system load averages for the past 1, 5, and 15 minutes.
func GetLoadAvg() ([3]float64, error) {
	var loadAvg [3]C.double
	ret := C.get_load_avg(&loadAvg[0])
	if ret != 0 {
		return [3]float64{}, errors.New("failed to get load average")
	}

	return [3]float64{
		float64(loadAvg[0]),
		float64(loadAvg[1]),
		float64(loadAvg[2]),
	}, nil
}

// Cleanup releases any resources used by the CPU stats collector.
func Cleanup() {
	C.go_cleanup_cpu_stats()
}

func init() {
	runtime.SetFinalizer(new(bool), func(_ *bool) {
		C.go_cleanup_cpu_stats()
	})
}

func cpuError(code C.int) error {
	switch code {
	case -3: // CPU_ERROR_MEMORY
		return errors.New("memory allocation error")
	case -2: // CPU_ERROR_SYSCTL
		return errors.New("sysctl error")
	case -1: // CPU_ERROR_HOST_PROCESSOR_INFO
		return errors.New("host processor info error")
	default:
		return fmt.Errorf("unknown error: %d", code)
	}
}
