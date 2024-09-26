package application

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeTicker struct {
	channel    chan time.Time
	stopCalled bool
}

func (f *fakeTicker) C() <-chan time.Time {
	return f.channel
}

func (f *fakeTicker) Stop() {
	f.stopCalled = true
}

func Test_scheduleEvery(t *testing.T) {
	ticker := fakeTicker{channel: make(chan time.Time), stopCalled: false}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker.channel <- time.Now()
		ticker.channel <- time.Now()
		cancel()
	}()

	runCount := 0
	var wg sync.WaitGroup
	ScheduleEvery(&wg, &ticker, ctx, func() {
		runCount += 1
	})

	// We expect func to be executed 3 times
	// One initial time + 1 time for each tick
	if runCount != 3 {
		t.Fatalf("function executed %v. Expected 3 times", runCount)
	}

	if !ticker.stopCalled {
		t.Fatal("stop not called on ticker")
	}
}
