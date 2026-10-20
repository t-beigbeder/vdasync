package cli

import (
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type SignalHandler struct {
	mx       sync.Mutex
	lgr      *slog.Logger
	cleanups []func(*slog.Logger) error
}

func MakeSignalHandler(lgr *slog.Logger) *SignalHandler {
	sh := &SignalHandler{lgr: lgr}
	chEx := make(chan os.Signal, 1)
	signal.Notify(chEx, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range chEx {
			lgr.Error("main: signal received, preparing to exit", "signal", sig)
			for _, fcu := range sh.cleanups {
				if err := fcu(sh.lgr); err != nil {
					lgr.Error("main: cleanup error", "err", err)
				}
			}
			os.Exit(1)
		}
	}()
	return sh
}

func (sh *SignalHandler) RegisterCleanupFunction(f func(lgr *slog.Logger) error) {
	sh.mx.Lock()
	defer sh.mx.Unlock()
	sh.cleanups = append(sh.cleanups, f)
}
