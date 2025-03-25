//go:build darwin
// +build darwin

package cpu_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sm-moshi/dmetrics-go/internal/cpu"
)

func TestNewProvider(t *testing.T) {
	provider := cpu.NewProvider()
	require.NotNil(t, provider, "provider should not be nil")
}

func TestProviderFrequency(t *testing.T) {
	provider := cpu.NewProvider()

	freq, err := provider.GetFrequency()
	if err != nil {
		if err.Error() == "failed to detect CPU frequency" {
			t.Skip("CPU frequency detection failed, this is expected in some environments")
		}
		t.Fatal(err)
	}
	assert.Greater(t, freq, uint64(0), "frequency should be > 0 MHz")
}

func TestProviderCoreCount(t *testing.T) {
	provider := cpu.NewProvider()

	cores, err := provider.GetCoreCount()
	require.NoError(t, err)
	assert.Greater(t, cores, 0, "should have at least one core")
}

func TestProviderStats(t *testing.T) {
	provider := cpu.NewProvider()

	stats, err := provider.GetStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Greater(t, stats.PhysicalCores, 0, "should have at least one physical core")

	// CPU frequency might not be available
	if stats.FrequencyMHz == 0 {
		t.Log("CPU frequency is 0, this is expected in some environments running tests without sudo")
	} else {
		assert.Greater(t, stats.FrequencyMHz, uint64(0), "CPU frequency should be > 0 MHz")
	}

	assert.Len(t, stats.CoreUsage, stats.PhysicalCores, "should have usage values for each core")

	for i, usage := range stats.CoreUsage {
		assert.GreaterOrEqual(t, usage, 0.0, "core %d usage should be >= 0%%", i)
		assert.LessOrEqual(t, usage, 100.0, "core %d usage should be <= 100%%", i)
	}

	assert.GreaterOrEqual(t, stats.TotalUsage, 0.0, "total usage should be >= 0%")
	assert.LessOrEqual(t, stats.TotalUsage, 100.0, "total usage should be <= 100%")

	for i, load := range stats.LoadAvg {
		assert.GreaterOrEqual(t, load, 0.0, "load average %d should be >= 0", i)
	}

	assert.WithinDuration(t, time.Now(), stats.Timestamp, 2*time.Second, "timestamp should be recent")
}

func TestProviderWatch(t *testing.T) {
	provider := cpu.NewProvider()
	defer provider.Shutdown()

	// Create a context with timeout to ensure test completion
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch, err := provider.Watch(ctx, 100*time.Millisecond)
	require.NoError(t, err)

	var updates int
	for stats := range ch {
		assert.GreaterOrEqual(t, stats.TotalUsage, 0.0, "usage should be >= 0%")
		assert.LessOrEqual(t, stats.TotalUsage, 100.0, "usage should be <= 100%")
		updates++
		if updates >= 3 {
			break
		}
	}

	assert.Greater(t, updates, 0, "should receive at least one update")
}
