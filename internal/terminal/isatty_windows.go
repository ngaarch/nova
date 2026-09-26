//go:build windows

package terminal

import (
	"syscall"
)

// IsTerminal checks if the given file descriptor refers to an interactive Windows console.
func IsTerminal(fd uintptr) bool {
	var st uint32
	err := syscall.GetConsoleMode(syscall.Handle(fd), &st)
	return err == nil
}
