package cmd

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func watchLoop(duration, statusInterval time.Duration, printStatus func()) string {
	if statusInterval <= 0 {
		statusInterval = 2 * time.Second
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	ticker := time.NewTicker(statusInterval)
	defer ticker.Stop()

	var timerC <-chan time.Time
	var timer *time.Timer
	if duration > 0 {
		timer = time.NewTimer(duration)
		defer timer.Stop()
		timerC = timer.C
	}

	printStatus()

	for {
		select {
		case <-ticker.C:
			printStatus()
		case sig := <-signals:
			return sig.String()
		case <-timerC:
			return "time limit reached"
		}
	}
}

func joinErrors(errs ...error) error {
	filtered := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			filtered = append(filtered, err)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return errors.Join(filtered...)
}
