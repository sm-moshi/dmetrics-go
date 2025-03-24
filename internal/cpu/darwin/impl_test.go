//go:build darwin
// +build darwin

package darwin

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create a package-level random source.
//
//nolint:gosec // G404: acceptable use of weak RNG for test timing variations
var (
	rngMu sync.Mutex
	rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func TestGetStats(t *testing.T) {
	stats, err := getStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Print all CPU statistics
	t.Logf("\nCPU Statistics:")
	t.Logf("  Physical Cores: %d", stats.PhysicalCores)
	t.Logf("  Frequency: %d MHz", stats.FrequencyMHz)
	t.Logf("  Total Usage: %.2f%%", stats.TotalUsage)
	t.Logf("  Load Averages: %.2f, %.2f, %.2f",
		stats.LoadAvg[0], stats.LoadAvg[1], stats.LoadAvg[2])

	t.Logf("\nPer-Core Usage:")
	for i, usage := range stats.CoreUsage {
		t.Logf("  Core %d: %.2f%%", i, usage)
	}

	// Test physical cores
	assert.Greater(t, stats.PhysicalCores, 0, "should have at least one physical core")

	// Test CPU frequency
	assert.Greater(t, stats.FrequencyMHz, uint64(0), "CPU frequency should be greater than 0 MHz")

	// Test core usage values
	assert.Len(t, stats.CoreUsage, stats.PhysicalCores, "should have usage values for each core")
	for i, usage := range stats.CoreUsage {
		assert.GreaterOrEqual(t, usage, 0.0, "core %d usage should be >= 0%%", i)
		assert.LessOrEqual(t, usage, 100.0, "core %d usage should be <= 100%%", i)
	}

	// Test total usage
	assert.GreaterOrEqual(t, stats.TotalUsage, 0.0, "total usage should be >= 0%")
	assert.LessOrEqual(t, stats.TotalUsage, 100.0, "total usage should be <= 100%")

	// Test load averages
	for i, load := range stats.LoadAvg {
		assert.GreaterOrEqual(t, load, 0.0, "load average %d should be >= 0", i)
	}

	// Test timestamp
	assert.WithinDuration(t, time.Now(), stats.Timestamp, 2*time.Second, "timestamp should be recent")
}

func TestUsage(t *testing.T) {
	usage, err := usage()
	require.NoError(t, err)
	t.Logf("\nCurrent CPU Usage: %.2f%%", usage)
	assert.GreaterOrEqual(t, usage, 0.0, "usage should be > 0%")
	assert.LessOrEqual(t, usage, 100.0, "usage should be <= 100%")
}

func TestFrequency(t *testing.T) {
	freq, err := getFrequency()
	require.NoError(t, err)
	t.Logf("\nCurrent CPU Frequency: %d MHz", freq)
	assert.Greater(t, freq, uint64(0), "frequency should be > 0 MHz")
}

func TestLoadAverage(t *testing.T) {
	loads, err := getLoadAvg()
	require.NoError(t, err)
	t.Logf("\nLoad Averages (1, 5, 15 min): %.2f, %.2f, %.2f",
		loads[0], loads[1], loads[2])

	// Test each load average period (1, 5, 15 minutes)
	for i, load := range loads {
		assert.GreaterOrEqual(t, load, 0.0, "load average %d should be >= 0", i)
	}

	// 1-minute load should be available
	assert.NotEqual(t, 0.0, loads[0], "1-minute load average should be available")
}

func TestGetStatsConcurrent(t *testing.T) {
	const numGoroutines = 5
	const numIterations = 3

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				stats, err := getStats()
				if err != nil {
					t.Logf("Goroutine %d, iteration %d failed: %v", id, j, err)
					errChan <- err
					return
				}
				if stats == nil {
					errChan <- fmt.Errorf("goroutine %d received nil stats", id)
					return
				}

				// Use the package-level RNG with mutex protection
				rngMu.Lock()
				sleepTime := time.Duration(10+rng.Intn(40)) * time.Millisecond
				rngMu.Unlock()
				time.Sleep(sleepTime)
			}
			errChan <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect results
	for err := range errChan {
		assert.NoError(t, err)
	}
}
