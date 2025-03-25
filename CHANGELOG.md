# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial CPU metrics implementation for macOS
  - Added CPU usage tracking (total and per-core)
  - Added platform detection (Apple Silicon vs Intel)
  - Added core count detection (performance and efficiency cores)
  - Added load average monitoring
  - Added proper context handling for cancellation
  - Implemented thread-safe concurrent access
  - CPU frequency detection framework in place (currently non-functional)
- Comprehensive package documentation
  - Added root package `doc.go` with library overview
  - Added `cpu/doc.go` with detailed CPU metrics documentation
  - Added runnable examples and implementation details
  - Added context usage and thread safety documentation
- Basic example application
  - Real-time CPU statistics monitoring
  - Per-core usage visualization with Unicode bar graphs
  - Clean shutdown handling and proper resource cleanup
  - Added context-aware cancellation examples
- Power metrics module for Darwin systems
  - Battery status monitoring (percentage, state, health)
  - Power source detection (AC vs Battery)
  - System power consumption metrics (CPU, GPU, total)
  - Real-time monitoring with configurable intervals
  - Added proper context handling for cancellation
  - Implemented thread-safe concurrent access
- Test suite implementation in progress:
  - Unit tests with table-driven approach
  - Initial integration tests (some complexity issues remain)
  - Benchmarks for performance critical paths
  - Context cancellation testing
  - Race condition detection

### Changed

- Implemented proper C to Go type bridging using cgo
- Added safe memory management for Mach calls
- Improved error handling with custom error types
- Added constants to replace magic numbers
- Enhanced error wrapping using `%w` verb
- Improved CPU frequency detection error handling
  - Added proper error propagation from low-level calls
  - Implemented graceful fallback chain for frequency detection
  - Updated tests to handle frequency detection failures
  - Fixed variable shadowing in CPU usage calculation
- Improved example application reliability
  - Added proper defer handling
  - Enhanced error reporting
  - Fixed resource cleanup on exit
- Refactored initialization patterns
  - Removed usage of `init()` functions
  - Moved initialization to constructors
  - Added proper error handling in initialization
- Enhanced test structure
  - Split large test functions into focused units
  - Added concurrent access test cases
  - Improved test readability and maintainability

### Fixed

- Fixed per-core CPU usage calculation
- Fixed type conversion issues between C and Go
- Fixed memory leaks in CPU stats collection
- Fixed variable shadowing in GetUsage method
- Fixed linter warnings
  - Replaced magic numbers with named constants
  - Corrected error handling in deferred calls
  - Improved error wrapping format
  - Removed unused variables and imports
  - Fixed function length issues
  - Removed weak random number generator usage
  - Eliminated unused mutex declarations

### Known Issues

- CPU frequency detection may return 0 on some systems (expected behaviour)
- Integration tests show complexity issues exceeding thresholds
- TestPowerMetricsIntegration needs refactoring (complexity: 23/20)
- Some integration tests remain incomplete

### Technical Details

- Implemented `cpu.Stats` structure for comprehensive CPU information
- Added thread-safe platform information caching
- Implemented proper cleanup using finalizers
- Added comprehensive error handling for syscalls
- Enhanced concurrency safety with proper mutex usage
- Improved context handling across all providers

### Pending

- Replace deprecated `rand.Seed` with `rand.New(rand.NewSource())` (Go 1.20+ compatibility)
- Replace weak random number generator with cryptographically secure alternative
- Update test suite to use modern Go random number generation patterns
- Fix CPU frequency detection implementation
- Refactor complex integration tests
- Complete remaining integration test coverage

<!-- markdownlint-configure-file
MD024:
  # Only check sibling headings
  siblings_only: true
-->
