//go:build darwin
// +build darwin

// Package darwin provides Darwin-specific power metrics implementation using IOKit.
//
// This package implements power metrics collection for macOS systems by interfacing with
// IOKit Power Sources for battery information and capacity metrics.
//
// Battery Capacity Handling:
// The implementation retrieves three types of capacity measurements:
//   - Current Capacity: The current charge level of the battery
//   - Max Capacity: The maximum capacity the battery can currently hold
//   - Design Capacity: The original design capacity of the battery
//
// Battery Health Calculation:
// Battery health is determined by comparing max capacity to design capacity:
// - Good: ≥80% of design capacity
// - Fair: ≥50% of design capacity
// - Poor: <50% of design capacity
//
// The implementation provides:
// - Battery status (charging state, percentage, health)
// - Power source detection (AC vs Battery)
// - Battery capacity and cycle count information
// - Time estimates (remaining on battery, time to full charge)
//
// Thread Safety:
// All public methods are thread-safe and handle concurrent access appropriately.
//
// Error Handling:
// - Returns types.ErrNoBattery when battery operations are attempted without a battery present
// - Returns descriptive errors for IOKit operation failures
// - Provides clear feedback for operation failures
//
// Example Usage:
//
//	provider := darwin.NewProvider()
//	stats, err := provider.GetStats(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Battery: %.1f%%, Health: %v\n", stats.Percentage, stats.Health)
package darwin
