//go:build windows

package terminal

import (
	"os"
	"syscall"
	"unsafe"
)

type coord struct {
	x int16
	y int16
}

type smallRect struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        uint16
	window            smallRect
	maximumWindowSize coord
}

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

// GetSize queries the Windows console dimensions or falls back to environment variables.
func GetSize(file *os.File, getenv func(string) string) (int, int) {
	if file != nil {
		var csbi consoleScreenBufferInfo
		r1, _, _ := procGetConsoleScreenBufferInfo.Call(
			file.Fd(),
			uintptr(unsafe.Pointer(&csbi)),
		)
		if r1 != 0 {
			w := int(csbi.window.right - csbi.window.left + 1)
			h := int(csbi.window.bottom - csbi.window.top + 1)
			if w > 0 && h > 0 {
				return w, h
			}
		}
	}
	return FallbackDimensions(getenv)
}
