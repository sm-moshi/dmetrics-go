# 🚧 ROADMAP – dmetrics-go

A Go-native reimplementation of `darwin-metrics`, exposing system metrics on macOS via syscalls and IOKit.

---

## v0.1 – Minimum Viable Port

- System architecture detection
- CPU usage & frequency (sysctl)
- Power source and battery percentage
- Modular package layout

## v0.2 – Thermal & GPU Data

- Temperature sensors (via SMC)
- Fan speed metrics
- GPU memory usage and vendor info
- Begin cgo bridge to Metal

## v0.3 – Full System Monitor

- Process statistics
- Network I/O & interfaces
- System memory + swap usage
- Integration with observability tools

## v1.0 – Stable API

- Full test suite
- Stable `darwinmetrics` Go module
- Package documentation on pkg.go.dev
- Semantic versioning and changelogs

---

## Beyond v1.0

- CLI tool: `dmetrics` (Go binary)
- Prometheus exporter
- TUI system dashboard
- Cross-compilation pipeline
