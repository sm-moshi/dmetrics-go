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
	fmt.Println("=== Power Information ===")
	fmt.Printf("Battery Present: %v\n", stats.IsPresent)
	if stats.IsPresent {
		fmt.Printf("Battery Level: %.1f%%\n", stats.Percentage)
		fmt.Printf("Charging Status: %v\n", stats.State)
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
		fmt.Printf("Battery Present: %v\n", stats.IsPresent)
		if stats.IsPresent {
			fmt.Printf("Battery Level: %.1f%%\n", stats.Percentage)
			fmt.Printf("Charging Status: %v\n", stats.State)
		}
		fmt.Printf("CPU Power: %.1fW\n", stats.CPUPower)
		fmt.Printf("GPU Power: %.1fW\n", stats.GPUPower)
		fmt.Printf("Total Power: %.1fW\n", stats.TotalPower)
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
