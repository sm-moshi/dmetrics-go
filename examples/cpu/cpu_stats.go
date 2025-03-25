// Package main provides an example of using the CPU metrics functionality
// from the dmetrics-go library. It demonstrates how to collect and display
// CPU statistics including usage, frequency, and core count.
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
)

const cpuUsageBarScale = 5 // Each bar character represents 5% CPU usage

func printStats(ctx context.Context) error {
	provider := cpu.NewProvider()
	stats, err := provider.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get CPU stats: %w", err)
	}

	// Clear screen (ANSI escape sequence)
	fmt.Print("\033[H\033[2J")

	fmt.Printf("CPU Statistics (Updated: %s)\n", stats.Timestamp.Format("15:04:05"))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("  Physical Cores: %d\n", stats.PhysicalCores)
	fmt.Printf("  Frequency: %d MHz\n", stats.FrequencyMHz)
	fmt.Printf("  Total Usage: %.2f%%\n", stats.TotalUsage)
	fmt.Printf("  Load Averages (1, 5, 15 min): %.2f, %.2f, %.2f\n",
		stats.LoadAvg[0], stats.LoadAvg[1], stats.LoadAvg[2])

	fmt.Printf("\nPer-Core Usage:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	for i, usage := range stats.CoreUsage {
		// Create a simple bar graph
		barLength := int(usage / cpuUsageBarScale)
		bar := strings.Repeat("█", barLength)
		fmt.Printf("  Core %2d [%-20s] %.2f%%\n", i, bar, usage)
	}

	return nil
}

func run() error {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Initial check to ensure we can get stats
	if err := printStats(ctx); err != nil {
		return fmt.Errorf("initial stats check failed: %w", err)
	}

	// Print stats every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("Press Ctrl+C to exit...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sigCh:
			fmt.Println("\nShutting down...")
			return nil
		case <-ticker.C:
			if err := printStats(ctx); err != nil {
				if ctx.Err() != nil {
					return nil // Context cancelled, exit silently
				}
				return fmt.Errorf("failed to print stats: %w", err)
			}
		}
	}
}

func main() {
	if err := run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
