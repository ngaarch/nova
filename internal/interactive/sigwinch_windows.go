//go:build windows

package interactive

import (
	"os"
)

func setupResizeSignal() (chan os.Signal, func()) {
	sigChan := make(chan os.Signal, 1)
	stop := func() {}
	return sigChan, stop
}
