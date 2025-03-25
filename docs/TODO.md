# ✅ TODO – dmetrics-go

This TODO list tracks the porting process from [darwin-metrics (Rust)](https://github.com/sm-moshi/darwin-metrics) to an idiomatic, modular, and testable Go library (`dmetrics-go`). This port will follow Go's best practices for package structure, error handling, concurrency, and system-level programming using `cgo`.

---

## 📦 Project Bootstrapping

- [x] **Create Go module structure**
  - Mirror the Rust crate layout into idiomatic Go packages.
  - Create placeholder files for each subsystem.

- [x] **Define Go error handling model**
  - Replace Rust's `thiserror`-based enums with structured Go error types.
  - Add support for `errors.Is`, `errors.As`, and wrapping with `fmt.Errorf`.

- [x] **Initialize `go.mod` and define module path**
  - Set up the Go module name (`dmetrics-go` or GitHub URL).
  - Add required `x/sys`, `cgo`, and testing packages.

- [x] **Apply `// +build darwin` tags**
  - Ensure platform-specific logic is excluded on non-macOS systems.
  - Set up build constraints for any file using macOS-only frameworks (IOKit, CoreFoundation, Metal).

---

## 🧠 Core API Porting

- [ ] **CPU Metrics**
  - [x] Use `sysctl` via `cgo` to get CPU usage, core count
  - [x] Recreate `cpu::usage()`
  - [ ] Fix `cpu::frequency()` implementation (currently returns 0)
  - [x] Implement thread-safe concurrent access
  - [x] Add proper context handling for cancellation

- [ ] **Memory Metrics**
  - Implement physical and virtual memory statistics via `sysctl`
  - Track active, free, used, and swap memory pages

- [ ] **GPU Metrics (IOKit)**
  - Read GPU vendor, model, and VRAM usage via `IOService` registry
  - Mirror `gpu::get_vram_size()` and `get_gpu_vendor()`

- [ ] **Temperature Sensors (SMC)**
  - Implement `SMCOpen`, `SMCReadKey`, `SMCClose` bridges
  - Use fan sensor keys, die temp, CPU/GPU thermals

- [ ] **Fan Speed Metrics**
  - Use `F0Ac`, `F1Ac` (etc.) SMC keys to read RPM per fan
  - Structure this as `temperature/fans.go`

- [x] **Power Source / Battery Info**
  - [x] Read power source type (AC/battery) via `IOPSCopyPowerSourcesInfo`
  - [x] Extract battery charge and status
  - [x] Implement thread-safe concurrent access
  - [x] Add proper context handling for cancellation
  - [ ] Future v0.2: Add cycle count and health monitoring via SMC
  - [ ] Future v0.2: Add detailed power consumption metrics

- [ ] **Process Monitoring**
  - Use `libproc` via `cgo` to list processes and gather PID stats
  - Add CPU time, memory, command name, parent/child tracking

- [ ] **Network Statistics**
  - Use BSD syscalls (`getifaddrs`, `sysctl`) to collect per-interface traffic
  - Expose bandwidth usage, interface state, error counters

---

## 🔬 Testing & Validation

- [ ] **Unit Tests**
  - [x] Each package has comprehensive `*_test.go` suites with table-driven tests
  - [x] Success and failure path testing implemented
  - [x] Tests are properly structured and focused
  - [x] Concurrent access testing added
  - [ ] Refactor TestPowerMetricsIntegration (complexity: 23/20)

- [ ] **Integration Tests**
  - [ ] Complete test inter-package behavior (e.g. combined CPU + Power polling)
  - [ ] Verify FFI and system interactions don't panic or leak
  - [x] Concurrent access testing implemented
  - [x] Context cancellation properly tested
  - [ ] Address test complexity issues

- [ ] **macOS Version Compatibility**
  - Validate on macOS 12, 13, 14, 15+
  - Prefer public, stable symbols to avoid future breakage

---

## 📚 Documentation

- [x] **Public API Docs**
  - Added comprehensive comments to all exported functions, types, and constants
  - Documented context usage and thread safety guarantees
  - Added examples of proper error handling

- [x] **Module-Level Docs**
  - Added `doc.go` to root and each package with overview usage
  - Documented concurrency and context patterns

- [x] **Examples**
  - Created examples showing common usage patterns
  - Added examples demonstrating proper context usage
  - Included examples of concurrent access patterns

---

## 🔁 FFI & CGO Integration

- [x] **Bridge Headers**
  - Created `.h` and `.m`/`.c` files for Objective-C/SMC/IOKit glue
  - Properly documented C function signatures

- [x] **Safe Wrappers**
  - No exposure of `unsafe.Pointer` in user-facing code
  - Abstract over CFTypes, io_service_t, etc.
  - Added context support for cancellation
  - Implemented proper mutex protection

- [x] **Memory Safety**
  - Ensure all allocated memory is `free()`d properly
  - Use `defer` to manage lifecycle of CF objects and C strings
  - Added cleanup finalizers where needed

---

## 🛠 Misc

- [x] Set up `golangci-lint` and `goimports` formatting
  - All linter issues resolved
  - Code follows Go best practices
  - No init functions used

- [x] Setup CI workflow: `go test ./...` + static checks on PR
  - Added race detector to test suite
  - Implemented comprehensive test coverage

- [ ] Prepare semantic versioning (v0.1, v0.2, v1.0)
  - [ ] Fix CPU frequency detection for v0.1
  - [ ] Complete integration tests for v0.1
  - [ ] Address test complexity issues for v0.1
  - [ ] Document breaking changes
