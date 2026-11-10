package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// runCtx is cancelled by the first Ctrl+C (or SIGTERM). Long-running loops
// and prompts check it and return normally, so deferred clean-up such as
// leaveTemporaryGroups still runs on the main goroutine (the client is not
// safe for concurrent use, so the signal handler itself does no API calls).
var runCtx = context.Background()

// errInterrupted is returned when the user aborted the run.
var errInterrupted = errors.New("abgebrochen (Strg+C)")

// installInterruptHandler sets up runCtx and returns a function that removes
// the handler again. After the first signal the default behaviour is
// restored, so a second Ctrl+C terminates the program immediately.
func installInterruptHandler() func() {
	ctx, cancel := context.WithCancel(context.Background())
	runCtx = ctx

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})

	go func() {
		select {
		case <-signals:
			signal.Stop(signals)
			fmt.Fprintln(os.Stderr, "\nAbbruch angefordert – laufender Schritt wird beendet, temporäre Gruppen werden verlassen (erneut Strg+C beendet sofort).")
			cancel()
		case <-done:
		}
	}()

	return func() {
		close(done)
		signal.Stop(signals)
		cancel()
		runCtx = context.Background()
	}
}

// interrupted returns errInterrupted once the user has aborted the run.
func interrupted() error {
	if runCtx.Err() != nil {
		return errInterrupted
	}
	return nil
}
