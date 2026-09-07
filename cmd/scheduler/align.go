package main

import "time"

func durationUntilNextInterval(now time.Time, interval time.Duration) time.Duration {
	if interval <= 0 {
		interval = time.Minute
	}
	now = now.UTC()
	next := now.Truncate(interval).Add(interval)
	return next.Sub(now)
}
