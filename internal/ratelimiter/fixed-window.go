package ratelimiter

import (
	"sync"
	"time"
)

type FixedWindowLimiter struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindowRateLimiter(limit int, timeFrame time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  timeFrame,
	}
}

func (rl *FixedWindowLimiter) Allow(remoteAddr string) (bool, time.Duration) {
	rl.RLock()
	count, exists := rl.clients[remoteAddr]
	rl.RUnlock()

	if !exists || count < rl.limit {
		rl.Lock()
		if !exists {
			go rl.resetCount(remoteAddr)
		}

		rl.clients[remoteAddr]++
		rl.Unlock()
		return true, 0
	}

	return false, rl.window
}

func (rl *FixedWindowLimiter) resetCount(remoteAddr string) {
	time.Sleep(rl.window)
	rl.Lock()
	delete(rl.clients, remoteAddr)
	rl.Unlock()
}
