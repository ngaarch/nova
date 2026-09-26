//go:build linux || darwin || freebsd || openbsd || netbsd

package terminal

import (
	"syscall"
	"unsafe"
)

// IsTerminal checks if the given file descriptor refers to an interactive terminal.
func IsTerminal(fd uintptr) bool {
	var ws winsize
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	)
	return err == 0
}
