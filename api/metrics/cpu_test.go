package metrics_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/sm-moshi/dmetrics-go/api/metrics"
	"github.com/sm-moshi/dmetrics-go/internal/cpu/darwin"
)

// ExampleCPUMetrics demonstrates basic usage of the CPUMetrics interface.
func ExampleCPUMetrics() {
	ctx := context.Background()
	provider := darwin.NewProvider()
	defer provider.Shutdown()

	// Get current CPU stats
	stats, err := provider.GetStats(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)
}

// ExampleCPUMetrics_Watch demonstrates how to monitor CPU metrics.
func ExampleCPUMetrics_Watch() {
	ctx := context.Background()
	provider := darwin.NewProvider()
	defer provider.Shutdown()

	ch, err := provider.Watch(ctx, time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// Only read first update for the example
	stats := <-ch
	fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)
}

// ExampleCPUMetrics_GetPlatform demonstrates platform-specific features.
func ExampleCPUMetrics_GetPlatform() {
	ctx := context.Background()
	provider := darwin.NewProvider()
	defer provider.Shutdown()

	stats, err := provider.GetStats(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if stats.PerformanceCores > 0 {
		fmt.Println("Running on Apple Silicon")
	} else {
		fmt.Println("Running on Intel")
	}
}

func ExampleCPUMetrics_errorHandling() {
	ctx := context.Background()
	provider := darwin.NewProvider()

	// First, shut down the provider
	provider.Shutdown()

	// Now all operations should return ErrShutdown
	_, err := provider.GetStats(ctx)
	fmt.Printf("Error after shutdown: %v\n", err)
	// Output: Error after shutdown: provider has been shut down
}

// TestCPUMetrics_Shutdown verifies that operations return ErrShutdown after calling Shutdown.
func TestCPUMetrics_Shutdown(t *testing.T) {
	ctx := context.Background()
	provider := darwin.NewProvider()

	// First, shut down the provider
	if err := provider.Shutdown(); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	// Test all operations return ErrShutdown
	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "GetStats",
			fn: func() error {
				_, err := provider.GetStats(ctx)
				return err
			},
		},
		{
			name: "GetFrequency",
			fn: func() error {
				_, err := provider.GetFrequency(ctx)
				return err
			},
		},
		{
			name: "Watch",
			fn: func() error {
				_, err := provider.Watch(ctx, time.Second)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err != metrics.ErrShutdown {
				t.Errorf("got error %v, want %v", err, metrics.ErrShutdown)
			}
		})
	}
}

// TestCPUMetrics_Watch verifies Watch behavior.
func TestCPUMetrics_Watch(t *testing.T) {
	ctx := context.Background()
	provider := darwin.NewProvider()
	defer provider.Shutdown()

	// Test invalid interval
	_, err := provider.Watch(ctx, -time.Second)
	if err != metrics.ErrInvalidInterval {
		t.Errorf("Watch with negative interval: got error %v, want %v", err, metrics.ErrInvalidInterval)
	}

	// Test valid interval
	ch, err := provider.Watch(ctx, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Read first update
	select {
	case stats := <-ch:
		if stats.PhysicalCores == 0 {
			t.Error("received invalid stats: PhysicalCores is 0")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for stats")
	}
}

// TestCPUMetrics_GetStats verifies basic stats collection.
func TestCPUMetrics_GetStats(t *testing.T) {
	ctx := context.Background()
	provider := darwin.NewProvider()
	defer provider.Shutdown()

	stats, err := provider.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	// Verify basic stats validity
	if stats.TotalUsage < 0 || stats.TotalUsage > 100 {
		t.Errorf("invalid total usage: got %.2f, want value between 0 and 100", stats.TotalUsage)
	}

	if stats.PhysicalCores <= 0 {
		t.Errorf("invalid physical core count: got %d, want > 0", stats.PhysicalCores)
	}

	if len(stats.CoreUsage) != stats.PhysicalCores {
		t.Errorf("core usage length mismatch: got %d, want %d", len(stats.CoreUsage), stats.PhysicalCores)
	}
}
