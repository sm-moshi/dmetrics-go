// Package main provides an example of using both CPU and power metrics
// functionality from the dmetrics-go library. It demonstrates how to monitor
// system resources and power consumption in a unified view.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/sm-moshi/dmetrics-go/internal/cpu"
	"github.com/sm-moshi/dmetrics-go/internal/power"
	"github.com/sm-moshi/dmetrics-go/pkg/metrics/types"
)

const (
	// Display constants
	cpuBarScale    = 5 // Each bar character represents 5% CPU usage
	updateInterval = 2 * time.Second

	// ANSI escape codes
	clearScreen = "\033[H\033[2J"
	bold        = "\033[1m"
	reset       = "\033[0m"
)

type systemStats struct {
	cpu   *types.CPUStats
	power *types.PowerStats
}

// printStats formats and displays the current system statistics
func printStats(stats systemStats) {
	// Clear screen
	fmt.Print(clearScreen)

	// Print timestamp header
	now := time.Now().Format("15:04:05")
	fmt.Printf("%sSystem Statistics (Updated: %s)%s\n", bold, now, reset)
	fmt.Println(strings.Repeat("━", 50))

	// Print CPU information
	if stats.cpu != nil {
		fmt.Printf("\n%sCPU Information%s\n", bold, reset)
		fmt.Printf("  Physical Cores: %d\n", stats.cpu.PhysicalCores)
		fmt.Printf("  Frequency: %d MHz\n", stats.cpu.FrequencyMHz)
		fmt.Printf("  Total Usage: %.2f%%\n", stats.cpu.TotalUsage)
		fmt.Printf("  Load Averages (1, 5, 15 min): %.2f, %.2f, %.2f\n",
			stats.cpu.LoadAvg[0], stats.cpu.LoadAvg[1], stats.cpu.LoadAvg[2])

		fmt.Printf("\n%sCore Usage:%s\n", bold, reset)
		for i, usage := range stats.cpu.CoreUsage {
			barLength := int(usage / cpuBarScale)
			bar := strings.Repeat("█", barLength)
			fmt.Printf("  Core %2d [%-20s] %.2f%%\n", i, bar, usage)
		}
	}

	// Print power information
	if stats.power != nil {
		fmt.Printf("\n%sPower Information%s\n", bold, reset)
		fmt.Printf("  Power Source: %v\n", stats.power.Source)
		fmt.Printf("  CPU Power: %.1f W\n", stats.power.CPUPower)
		fmt.Printf("  GPU Power: %.1f W\n", stats.power.GPUPower)
		fmt.Printf("  Total Power: %.1f W\n", stats.power.TotalPower)

		if stats.power.Source == types.PowerSourceBattery {
			fmt.Printf("\n%sBattery Status%s\n", bold, reset)
			fmt.Printf("  Level: %.1f%%\n", stats.power.Percentage)
			fmt.Printf("  State: %v\n", stats.power.State)
			fmt.Printf("  Health: %v\n", stats.power.Health)
			if stats.power.TimeRemaining > 0 {
				fmt.Printf("  Time Remaining: %v\n",
					stats.power.TimeRemaining.Round(time.Minute))
			}
		}
	}
}

func run() error {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Create providers
	cpuProvider := cpu.NewProvider()
	powerProvider := power.NewProvider()

	// Start power monitoring
	powerCh, err := powerProvider.Watch(ctx, updateInterval)
	if err != nil {
		return fmt.Errorf("failed to start power monitoring: %w", err)
	}

	// Print initial stats
	cpuStats, err := cpuProvider.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get CPU stats: %w", err)
	}

	powerStats, err := powerProvider.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get power stats: %w", err)
	}

	printStats(systemStats{
		cpu:   cpuStats,
		power: powerStats,
	})

	// Update loop
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	fmt.Println("\nPress Ctrl+C to exit...")

	var lastPowerStats *types.PowerStats
	for {
		select {
		case <-ctx.Done():
			return nil
		case powerStats := <-powerCh:
			lastPowerStats = powerStats
		case <-ticker.C:
			cpuStats, err := cpuProvider.GetStats(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil // Context cancelled
				}
				return fmt.Errorf("failed to get CPU stats: %w", err)
			}
			printStats(systemStats{
				cpu:   cpuStats,
				power: lastPowerStats,
			})
		}
	}
}

func main() {
	if err := run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
