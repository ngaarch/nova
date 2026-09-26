//go:build linux

package terminal

import (
	"syscall"
	"unsafe"
)

// MakeRaw puts the terminal connected to fd into raw mode and returns a restore function.
func MakeRaw(fd int) (func(), error) {
	var oldTermios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCGETS, uintptr(unsafe.Pointer(&oldTermios)))
	if errno != 0 {
		return nil, errno
	}

	raw := oldTermios
	raw.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Cflag |= syscall.CS8
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCSETS, uintptr(unsafe.Pointer(&raw)))
	if errno != 0 {
		return nil, errno
	}

	restore := func() {
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCSETS, uintptr(unsafe.Pointer(&oldTermios)))
	}

	return restore, nil
}
