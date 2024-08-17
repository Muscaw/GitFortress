package application

import (
	"context"
	"sync"
	"time"
)

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

func ScheduleEvery(wg *sync.WaitGroup, ticker Ticker, ctx context.Context, f func()) {
	defer ticker.Stop()

	// Executing the first time directly
	wg.Add(1)
	f()
	wg.Done()
	for {
		select {
		case <-ticker.C():
			wg.Add(1)
			f()
			wg.Done()
		case <-ctx.Done():
			return
		}
	}
}
