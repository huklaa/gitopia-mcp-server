package handler

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements a sliding-window rate limiter for chain transactions.
type RateLimiter struct {
	mu         sync.Mutex
	perMinute  int
	perHour    int
	timestamps []time.Time
}

// NewRateLimiter creates a new RateLimiter with the given per-minute and per-hour limits.
func NewRateLimiter(perMinute, perHour int) *RateLimiter {
	return &RateLimiter{
		perMinute:  perMinute,
		perHour:    perHour,
		timestamps: make([]time.Time, 0),
	}
}

// Allow checks if a new request is allowed. If allowed, it records the request
// and returns nil. Otherwise it returns an error describing the limit hit.
func (r *RateLimiter) Allow() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	oneMinuteAgo := now.Add(-time.Minute)
	oneHourAgo := now.Add(-time.Hour)

	// Prune entries older than one hour
	pruned := r.timestamps[:0]
	for _, t := range r.timestamps {
		if t.After(oneHourAgo) {
			pruned = append(pruned, t)
		}
	}
	r.timestamps = pruned

	// Count requests in last minute
	minuteCount := 0
	for _, t := range r.timestamps {
		if t.After(oneMinuteAgo) {
			minuteCount++
		}
	}

	if minuteCount >= r.perMinute {
		return fmt.Errorf("rate limit exceeded: %d transactions per minute (limit: %d)", minuteCount, r.perMinute)
	}

	if len(r.timestamps) >= r.perHour {
		return fmt.Errorf("rate limit exceeded: %d transactions per hour (limit: %d)", len(r.timestamps), r.perHour)
	}

	r.timestamps = append(r.timestamps, now)
	return nil
}
