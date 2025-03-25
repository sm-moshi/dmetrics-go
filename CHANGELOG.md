# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- CPU metrics implementation for macOS
  - Per-core and total CPU utilisation tracking
  - Platform detection (Intel/Apple Silicon)
  - Core count detection
  - Load average monitoring
  - Context-aware cancellation support
  - Thread-safe concurrent access
  - Performance characterisation documentation

### Changed

- Improved C to Go type bridging
- Enhanced memory management with proper cleanup
- Optimised error handling with proper propagation
- Fixed CPU utilisation calculation
- Restructured test organisation

### Fixed

- Per-core CPU utilisation calculation
- Type conversion between C and Go
- Memory leaks in C implementation
- Variable shadowing in tests
- Thread safety in provider implementation

### Known Issues

- CPU frequency detection returns 0 (deferred to post-v0.1)
- Some integration tests need complexity optimisation

### Technical Details

- `cpu.Stats` structure provides:
  - Total CPU utilisation (percentage)
  - Per-core utilisation (percentage array)
  - Core count
  - Load averages (1, 5, 15 minutes)
- Thread-safe with minimal lock contention
- Performance characteristics:
  - Initial collection: ~500ms
  - Subsequent calls: ~1-2ms
  - Memory utilisation: ~4KB per core

### Pending

- CPU frequency detection improvements
- Integration test optimisation

<!-- markdownlint-configure-file
MD024:
  # Only check sibling headings
  siblings_only: true
-->
