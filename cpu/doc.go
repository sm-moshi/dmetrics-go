/*
Package cpu provides CPU statistics for macOS systems, supporting both Intel and Apple Silicon processors.

The package offers comprehensive CPU metrics including:
  - Total and per-core CPU usage percentages
  - CPU frequencies (including performance/efficiency cores on Apple Silicon)
  - System load averages
  - Processor architecture detection
  - Core count information

Key types and functions:

Stats struct contains all CPU-related metrics:

	type Stats struct {
	    User             float64    // User CPU time percentage
	    System           float64    // System CPU time percentage
	    Idle             float64    // Idle CPU time percentage
	    Nice             float64    // Nice CPU time percentage
	    FrequencyMHz     uint64     // CPU frequency in MHz
	    PerfFrequencyMHz uint64     // Performance cores frequency (Apple Silicon)
	    EffiFrequencyMHz uint64     // Efficiency cores frequency (Apple Silicon)
	    PhysicalCores    int        // Number of physical CPU cores
	    PerformanceCores int        // Number of performance cores (Apple Silicon)
	    EfficiencyCores  int        // Number of efficiency cores (Apple Silicon)
	    CoreUsage        []float64  // Per-core CPU usage percentages
	    TotalUsage       float64    // Total CPU usage percentage
	    LoadAvg          [3]float64 // Load averages for 1, 5, and 15 minutes
	    Timestamp        time.Time  // Time when the stats were collected
	}

Main functions:

	Get() (*Stats, error)           // Returns all CPU statistics
	Usage() (float64, error)        // Returns total CPU usage percentage
	Frequency() (uint64, error)     // Returns current CPU frequency in MHz
	IsAppleSilicon() (bool, error)  // Checks if running on Apple Silicon
	LoadAverage() ([3]float64, error) // Returns 1, 5, and 15 minute load averages

Example usage:

	// Get comprehensive CPU statistics
	stats, err := cpu.Get()
	if err != nil {
	    log.Fatal(err)
	}

	// Print CPU information
	fmt.Printf("Total CPU Usage: %.2f%%\n", stats.TotalUsage)
	fmt.Printf("CPU Frequency: %d MHz\n", stats.FrequencyMHz)
	fmt.Printf("Load Average (1min): %.2f\n", stats.LoadAvg[0])

	// Print per-core usage
	for i, usage := range stats.CoreUsage {
	    fmt.Printf("Core %d: %.2f%%\n", i, usage)
	}

	// Check for Apple Silicon
	isAS, _ := cpu.IsAppleSilicon()
	if isAS {
	    fmt.Printf("Performance Cores: %d at %d MHz\n",
	        stats.PerformanceCores, stats.PerfFrequencyMHz)
	    fmt.Printf("Efficiency Cores: %d at %d MHz\n",
	        stats.EfficiencyCores, stats.EffiFrequencyMHz)
	}

Implementation details:
  - Uses sysctl calls via cgo for CPU metrics
  - Implements proper memory management and cleanup
  - Thread-safe with mutex protection
  - Supports both Intel and Apple Silicon architectures
  - Calculates accurate CPU usage with proper sampling
  - Handles dynamic frequency scaling
*/
package cpu
