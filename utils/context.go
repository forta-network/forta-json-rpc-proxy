package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// InitMainContext returns an interruptable service for quicker service shutdown and replacement.
func InitMainContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		cancel()
	}()
	return ctx, cancel
}
