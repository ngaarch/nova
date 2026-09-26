//go:build !linux && !darwin

package terminal

import "errors"

// MakeRaw is a stub on platforms where direct termios ioctl is unsupported.
func MakeRaw(fd int) (func(), error) {
	return func() {}, errors.New("terminal raw mode unsupported on this platform")
}
