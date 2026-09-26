//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !windows

package terminal

import "os"

// GetSize returns fallback dimensions on unsupported operating systems.
func GetSize(file *os.File, getenv func(string) string) (int, int) {
	return FallbackDimensions(getenv)
}
