//go:build !windows

package interactive

import (
	"os"
	"os/signal"
	"syscall"
)

func setupResizeSignal() (chan os.Signal, func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	stop := func() {
		signal.Stop(sigChan)
	}
	return sigChan, stop
}
