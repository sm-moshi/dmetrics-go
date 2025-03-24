# 📊 dmetrics-go

[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![Go Reference](https://pkg.go.dev/badge/github.com/sm-smoshi/dmetrics-go.svg)](https://pkg.go.dev/github.com/sm-moshi/dmetrics-go)
[![go.mod](https://img.shields.io/github/go-mod/go-version/sm-moshi/dmetrics-go)](go.mod)
[![LICENSE](https://img.shields.io/github/license/sm-moshi/dmetrics-go)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/sm-moshi/dmetrics-go)](https://goreportcard.com/report/github.com/sm-moshi/dmetrics-go)
[![Codecov](https://codecov.io/gh/sm-moshi/dmetrics-go/branch/main/graph/badge.svg)](https://codecov.io/gh/sm-moshi/dmetrics-go)

A Go-native macOS system metrics library — the port of [`darwin-metrics`](https://github.com/sm-moshi/darwin-metrics) from Rust to Go.

> Exposes CPU, memory, GPU, power, temperature, and process stats using `sysctl`, `IOKit`, `SMC`, and `CoreFoundation`.

⭐ `Star` this repository if you find it valuable and worth maintaining.

👁 `Watch` this repository to get notified about new releases, issues, etc.

## 🚀 Features

- 🧠 Architecture detection (`arm64`, `x86_64`)
- ⚡ CPU usage and frequency via `sysctl`
- 🔋 Power metrics (battery, AC status, charging)
- 🌡️ Temperature sensors (SMC)
- 🌀 Fan speeds
- 🎮 GPU VRAM & vendor info (IOKit)
- 🧵 Process stats and uptime
- 🌐 Network interface stats

## 📦 Installation

```bash
go get github.com/sm-moshi/dmetrics-go
```

## 🛠 Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/sm-moshi/dmetrics-go/internal/cpu"
)

func main() {
    provider := cpu.NewProvider()
    stats, err := provider.GetStats(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Printf("CPU Usage: %.2f%%\n", stats.TotalUsage)
    fmt.Printf("Frequency: %d MHz\n", stats.FrequencyMHz)
}
```

## 🧱 Package Structure

- `api/metrics` - Public interfaces for metrics collection
- `internal/cpu` - Platform-specific CPU metrics implementation
- `internal/gpu` - Platform-specific GPU metrics implementation
- `internal/power` - Platform-specific power metrics implementation
- `internal/temperature` - Platform-specific temperature metrics implementation
- `internal/memory` - Platform-specific memory metrics implementation
- `internal/network` - Platform-specific network metrics implementation
- `internal/process` - Platform-specific process metrics implementation
- `pkg/metrics/types` - Common type definitions for all metrics

## Development

### Setup

1. Install [Go](https://golang.org/doc/install)
2. Install [Visual Studio Code](https://code.visualstudio.com/)
3. Install [Go extension](https://code.visualstudio.com/docs/languages/go)
4. Clone and open this repository
5. `F1` -> `Go: Install/Update Tools` -> (select all) -> OK

### Build

#### Terminal

- `make` - execute the build pipeline
- `make help` - print help for the Make targets

#### Visual Studio Code

`F1` → `Tasks: Run Build Task (Ctrl+Shift+B or ⇧⌘B)` to execute the build pipeline

## 🧪 Testing

```bash
go test ./...
```

> Note: Tests only run on macOS: `// +build darwin`

## 🔍 Code Quality

- Continuous integration via [GitHub Actions](https://github.com/features/actions)
- Code formatting using [gofumpt](https://github.com/mvdan/gofumpt)
- Linting with [golangci-lint](https://github.com/golangci/golangci-lint)
- Dependencies scanning with [Dependabot](https://dependabot.com)
- Security analysis using [CodeQL Action](https://docs.github.com/en/github/finding-security-vulnerabilities-and-errors-in-your-code/about-code-scanning)

## 📜 License

MIT © 2025 [sm-moshi](https://github.com/sm-moshi)

## Contributing

Feel free to create an issue or propose a pull request.

Follow the [Code of Conduct](CODE_OF_CONDUCT.md).
