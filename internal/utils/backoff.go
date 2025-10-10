package utils

import (
	"time"
)

type RateLimiterFunc func(prev time.Duration) time.Duration
type Ticker func() bool

func NewBackoffTicker(increase, decrease RateLimiterFunc, maxDuration time.Duration) Ticker {
	var currentInterval time.Duration
	var nextBackoffTime time.Time

	return func() bool {
		now := time.Now()

		if now.Before(nextBackoffTime) {
			currentInterval = increase(currentInterval)
			if currentInterval > maxDuration {
				currentInterval = maxDuration
			}
			return false
		}

		if currentInterval > 0 {
			currentInterval = decrease(currentInterval)
		}

		nextBackoffTime = now.Add(currentInterval)
		return true
	}
}

func Double(start time.Duration) RateLimiterFunc {
	return func(prev time.Duration) time.Duration {
		if prev == 0 {
			return start
		}
		return prev * 2
	}
}

func Halve(minDuration time.Duration) RateLimiterFunc {
	return func(prev time.Duration) time.Duration {
		next := prev / 2
		if next < minDuration {
			return minDuration
		}
		return next
	}
}
