// Package internal provides version information for dmetrics-go packages.
package internal

// Version information
const (
	// Version is the current version of dmetrics-go.
	Version = "v0.1.0"

	// MinimumDarwinVersion is the minimum required version of macOS.
	MinimumDarwinVersion = "10.15" // Catalina

	// BuildDate will be injected at build time.
	BuildDate = "unknown"

	// GitCommit will be injected at build time.
	GitCommit = "unknown"
)

// GetVersionInfo returns a map containing version information.
func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"min_darwin": MinimumDarwinVersion,
		"build_date": BuildDate,
		"git_commit": GitCommit,
	}
}
