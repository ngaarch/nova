package cli

import (
	"fmt"
	"runtime"
)

// Version information set at build time via -ldflags.
var (
	Version   = "1.4.0"
	GitCommit = "none"
	BuildDate = "unknown"
)

// VersionInfo holds structured version and build metadata.
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetVersionInfo returns populated version metadata.
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// Format returns a human-readable version string.
func (v VersionInfo) Format() string {
	return fmt.Sprintf("nova %s (%s %s) %s %s",
		v.Version,
		v.GitCommit,
		v.BuildDate,
		v.GoVersion,
		v.Platform,
	)
}
