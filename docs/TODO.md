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

- [x] **CPU Metrics**
  - Use `sysctl` via `cgo` to get CPU usage, core count, frequency.
  - Recreate `cpu::usage()` and `cpu::frequency()`.

- [ ] **Memory Metrics**
  - Implement physical and virtual memory statistics via `sysctl`.
  - Track active, free, used, and swap memory pages.

- [ ] **GPU Metrics (IOKit)**
  - Read GPU vendor, model, and VRAM usage via `IOService` registry.
  - Mirror `gpu::get_vram_size()` and `get_gpu_vendor()`.

- [ ] **Temperature Sensors (SMC)**
  - Implement `SMCOpen`, `SMCReadKey`, `SMCClose` bridges.
  - Use fan sensor keys, die temp, CPU/GPU thermals.

- [ ] **Fan Speed Metrics**
  - Use `F0Ac`, `F1Ac` (etc.) SMC keys to read RPM per fan.
  - Structure this as `temperature/fans.go`.

- [ ] **Power Source / Battery Info**
  - Read power source type (AC/battery) via `IOPSCopyPowerSourcesInfo`.
  - Extract battery charge, status, cycle count, and health.

- [ ] **Process Monitoring**
  - Use `libproc` via `cgo` to list processes and gather PID stats.
  - Add CPU time, memory, command name, parent/child tracking.

- [ ] **Network Statistics**
  - Use BSD syscalls (`getifaddrs`, `sysctl`) to collect per-interface traffic.
  - Expose bandwidth usage, interface state, error counters.

---

## 🔬 Testing & Validation

- [x] **Unit Tests**
  - Each package must have a `*_test.go` suite with table-driven tests.
  - Include success and failure path testing.

- [x] **Integration Tests**
  - Test inter-package behavior (e.g. combined CPU + Memory polling).
  - Verify FFI and system interactions don't panic or leak.

- [ ] **macOS Version Compatibility**
  - Validate on macOS 12, 13, 14+.
  - Prefer public, stable symbols to avoid future breakage.

---

## 📚 Documentation

- [~] **Public API Docs**
  - Add `//` comments to all exported functions, types, and constants.

- [~] **Module-Level Docs**
  - Add `doc.go` to root and each package with overview usage.

- [ ] **Examples**
  - Create examples showing common usage patterns.

---

## 🔁 FFI & CGO Integration

- [x] **Bridge Headers**
  - Create `.h` and `.m`/`.c` files for Objective-C/SMC/IOKit glue.

- [x] **Safe Wrappers**
  - Do not expose `unsafe.Pointer` to user-facing code.
  - Abstract over CFTypes, io_service_t, etc.

- [x] **Memory Safety**
  - Ensure all allocated memory is `free()`d properly.
  - Use `defer` to manage lifecycle of CF objects and C strings.

---

## 🛠 Misc

- [x] Set up `golangci-lint` and `goimports` formatting.
- [x] Setup CI workflow: `go test ./...` + static checks on PR.
- [x] Prepare semantic versioning (v0.1, v0.2, v1.0).
