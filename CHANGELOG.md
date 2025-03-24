# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial CPU metrics implementation for macOS
  - Added CPU usage tracking (total and per-core)
  - Added CPU frequency detection for both Intel and Apple Silicon
  - Added load average monitoring
  - Added platform detection (Apple Silicon vs Intel)
  - Added core count detection (performance and efficiency cores)
- Comprehensive package documentation
  - Added root package `doc.go` with library overview
  - Added `cpu/doc.go` with detailed CPU metrics documentation
  - Added runnable examples and implementation details
- Basic example application
  - Real-time CPU statistics monitoring
  - Per-core usage visualization with Unicode bar graphs
  - Clean shutdown handling and proper resource cleanup

### Changed

- Implemented proper C to Go type bridging using cgo
- Added safe memory management for Mach calls
- Improved error handling with custom error types
- Added constants to replace magic numbers
- Enhanced error wrapping using `%w` verb
- Improved example application reliability
  - Added proper defer handling
  - Enhanced error reporting
  - Fixed resource cleanup on exit

### Fixed

- Fixed CPU frequency detection to use actual values instead of hardcoded ones
- Fixed per-core CPU usage calculation
- Fixed type conversion issues between C and Go
- Fixed memory leaks in CPU stats collection
- Fixed linter warnings
  - Replaced magic numbers with named constants
  - Corrected error handling in deferred calls
  - Improved error wrapping format

### Technical Details

- Implemented `cpu.Stats` structure for comprehensive CPU information
- Added thread-safe platform information caching
- Implemented proper cleanup using finalizers
- Added comprehensive error handling for syscalls

### Pending

- Replace deprecated `rand.Seed` with `rand.New(rand.NewSource())` (Go 1.20+ compatibility)
- Replace weak random number generator with cryptographically secure alternative
- Update test suite to use modern Go random number generation patterns

<!-- markdownlint-configure-file
MD024:
  # Only check sibling headings
  siblings_only: true
-->
