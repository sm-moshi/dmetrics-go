// Package main provides an example of using the power metrics functionality
// from the dmetrics-go library. It demonstrates how to monitor power source,
// battery status, and power consumption in real-time.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/sm-moshi/dmetrics-go/internal/power"
	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

const (
	// updateInterval is the time between power metric updates.
	updateInterval = 5 * time.Second
)

func run() error {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Create power metrics provider
	provider := power.NewProvider()

	// Get current power stats
	stats, err := provider.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get power stats: %w", err)
	}

	// Print current power information
	fmt.Printf("Power Source: %v\n", stats.Source)
	if stats.Source == types.PowerSourceBattery {
		fmt.Printf("Battery Level: %.1f%%\n", stats.Percentage)
		fmt.Printf("Battery State: %v\n", stats.State)
		fmt.Printf("Battery Health: %v\n", stats.Health)
		if stats.TimeRemaining > 0 {
			fmt.Printf("Time Remaining: %v\n", stats.TimeRemaining.Round(time.Minute))
		}
	}

	// Monitor power metrics
	fmt.Println("\nMonitoring power metrics (Ctrl+C to stop)...")
	ch, err := provider.Watch(ctx, updateInterval)
	if err != nil {
		return fmt.Errorf("failed to start monitoring: %w", err)
	}

	// Print power metrics updates
	for stats := range ch {
		fmt.Printf("\n=== Power Update ===\n")
		fmt.Printf("Time: %v\n", stats.Timestamp.Format(time.Kitchen))
		fmt.Printf("Source: %v\n", stats.Source)
		fmt.Printf("CPU Power: %.1fW\n", stats.CPUPower)
		fmt.Printf("GPU Power: %.1fW\n", stats.GPUPower)
		fmt.Printf("Total Power: %.1fW\n", stats.TotalPower)

		if stats.Source == types.PowerSourceBattery {
			fmt.Printf("Battery: %.1f%% (%v)\n", stats.Percentage, stats.State)
			if stats.TimeRemaining > 0 {
				fmt.Printf("Time Remaining: %v\n", stats.TimeRemaining.Round(time.Minute))
			}
		}
		fmt.Println("=================")
	}

	fmt.Println("Monitoring stopped.")
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
