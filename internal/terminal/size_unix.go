//go:build linux || darwin || freebsd || openbsd || netbsd

package terminal

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

type winsize struct {
	ws_row    uint16
	ws_col    uint16
	ws_xpixel uint16
	ws_ypixel uint16
}

// GetSize queries the terminal dimensions of the file descriptor or falls back to environment variables.
// Explicit COLUMNS or LINES environment variable takes precedence over ioctl.
func GetSize(file *os.File, getenv func(string) string) (int, int) {
	if cols := getenv("COLUMNS"); cols != "" {
		if w, err := strconv.Atoi(cols); err == nil && w > 0 {
			h := 24
			if lines := getenv("LINES"); lines != "" {
				if hVal, err := strconv.Atoi(lines); err == nil && hVal > 0 {
					h = hVal
				}
			}
			return w, h
		}
	}

	if file != nil {
		var ws winsize
		_, _, err := syscall.Syscall(
			syscall.SYS_IOCTL,
			file.Fd(),
			uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(&ws)),
		)
		if err == 0 && ws.ws_col > 0 && ws.ws_row > 0 {
			return int(ws.ws_col), int(ws.ws_row)
		}
	}
	return FallbackDimensions(getenv)
}
