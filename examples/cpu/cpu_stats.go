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

	// Print header
	fmt.Printf("CPU Statistics (Updated: %s)\n", stats.Timestamp.Format("15:04:05"))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Print core information
	fmt.Printf("Core Configuration:\n")
	fmt.Printf("  Physical Cores: %d\n", stats.PhysicalCores)
	if stats.PerformanceCores > 0 || stats.EfficiencyCores > 0 {
		fmt.Printf("  Performance Cores: %d\n", stats.PerformanceCores)
		fmt.Printf("  Efficiency Cores: %d\n", stats.EfficiencyCores)
	}

	// Print frequency information
	fmt.Printf("\nFrequency Information:\n")
	fmt.Printf("  Base Frequency: %d MHz\n", stats.FrequencyMHz)
	if stats.PerfFrequencyMHz > 0 || stats.EffiFrequencyMHz > 0 {
		if stats.PerfFrequencyMHz > 0 {
			fmt.Printf("  Performance Core Freq: %d MHz\n", stats.PerfFrequencyMHz)
		}
		if stats.EffiFrequencyMHz > 0 {
			fmt.Printf("  Efficiency Core Freq: %d MHz\n", stats.EffiFrequencyMHz)
		}
	}

	// Print usage statistics
	fmt.Printf("\nUsage Statistics:\n")
	fmt.Printf("  Total Usage: %.2f%%\n", stats.TotalUsage)
	fmt.Printf("  User: %.2f%%, System: %.2f%%, Idle: %.2f%%, Nice: %.2f%%\n",
		stats.User, stats.System, stats.Idle, stats.Nice)
	fmt.Printf("  Load Averages (1, 5, 15 min): %.2f, %.2f, %.2f\n",
		stats.LoadAvg[0], stats.LoadAvg[1], stats.LoadAvg[2])

	// Print per-core usage with improved visualization
	fmt.Printf("\nPer-Core Usage:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	for i, usage := range stats.CoreUsage {
		// Create a simple bar graph
		barLength := int(usage / cpuUsageBarScale)
		bar := strings.Repeat("█", barLength)
		coreType := ""
		if i < stats.PerformanceCores {
			coreType = "P" // Performance core
		} else if stats.EfficiencyCores > 0 {
			coreType = "E" // Efficiency core
		}
		if coreType != "" {
			fmt.Printf("  Core %s%d [%-20s] %.2f%%\n", coreType, i, bar, usage)
		} else {
			fmt.Printf("  Core %2d [%-20s] %.2f%%\n", i, bar, usage)
		}
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

	fmt.Println("\nPress Ctrl+C to exit...")

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
