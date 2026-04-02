package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_AllowWithinLimits(t *testing.T) {
	rl := NewRateLimiter(5, 100)

	for i := 0; i < 5; i++ {
		err := rl.Allow()
		require.NoError(t, err, "request %d should be allowed", i+1)
	}
}

func TestRateLimiter_RejectPerMinute(t *testing.T) {
	rl := NewRateLimiter(3, 100)

	// Use up the per-minute limit
	for i := 0; i < 3; i++ {
		require.NoError(t, rl.Allow())
	}

	// Next request should be rejected
	err := rl.Allow()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "per minute")
}

func TestRateLimiter_RejectPerHour(t *testing.T) {
	rl := NewRateLimiter(100, 3)

	// Use up the per-hour limit
	for i := 0; i < 3; i++ {
		require.NoError(t, rl.Allow())
	}

	// Next request should be rejected
	err := rl.Allow()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "per hour")
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(100, 1000)

	done := make(chan error, 50)
	for i := 0; i < 50; i++ {
		go func() {
			done <- rl.Allow()
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}
	// No panic or race condition
}

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 100)
	assert.NotNil(t, rl)
	assert.Equal(t, 10, rl.perMinute)
	assert.Equal(t, 100, rl.perHour)
	assert.Empty(t, rl.timestamps)
}
