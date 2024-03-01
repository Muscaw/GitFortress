package application

import "time"

func ScheduleEvery(d time.Duration, f func()) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	// Executing the first time directly
	f()
	for range ticker.C {
		f()
	}
}
