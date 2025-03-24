package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sm-moshi/dmetrics-go/cpu"
)

const cpuUsageBarScale = 5 // Each bar character represents 5% CPU usage

func printStats() error {
	// Get all CPU statistics
	stats, err := cpu.Get()
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

func main() {
	// Initial check to ensure we can get stats
	if err := printStats(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}

	// Print stats every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("Press Ctrl+C to exit...")

	for {
		if err := printStats(); err != nil {
			log.Printf("Error: %v", err)
			return
		}
		<-ticker.C
	}
}
