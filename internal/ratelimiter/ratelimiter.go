package ratelimiter

import (
	"time"
)

type Limiter interface {
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	RequestPerTimeFrame int
	TimeFrame           time.Duration
	Enabled             bool
}

func (rl *FixedWindowRateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.RLock()
	count, exists := rl.clients[ip]
	rl.RUnlock()
	if !exists || count < rl.limit {
		rl.Lock()
		if !exists {
			go rl.resetCount(ip)
		}
		rl.clients[ip]++
		rl.Unlock()
		return true, 0
	}

	return false, rl.window
}

func (rl *FixedWindowRateLimiter) resetCount(ip string) {
	time.Sleep(rl.window)
	rl.Lock()
	delete(rl.clients, ip)
	rl.Unlock()
}
