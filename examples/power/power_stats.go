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

func main() {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		cancel()
	}()

	// Create power metrics provider
	provider := power.NewProvider()

	// Get current power stats
	stats, err := provider.GetStats(ctx)
	if err != nil {
		log.Fatalf("Failed to get power stats: %v", err)
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
	ch, err := provider.Watch(ctx, 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to start monitoring: %v", err)
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
}
