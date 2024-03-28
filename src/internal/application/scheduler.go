package application

import (
	"context"
	"time"
)

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

func ScheduleEvery(ticker Ticker, ctx context.Context, f func()) {
	defer ticker.Stop()

	// Executing the first time directly
	f()
	for {
		select {
		case <-ticker.C():
			f()
		case <-ctx.Done():
			return
		}
	}
}
