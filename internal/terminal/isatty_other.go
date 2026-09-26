//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !windows

package terminal

// IsTerminal fallback for unsupported operating systems.
func IsTerminal(fd uintptr) bool {
	return false
}
