package tmdb

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter for TMDB API
// TMDB allows 40 requests per 10 seconds
type RateLimiter struct {
	tokens   int
	capacity int
	refill   int           // tokens to add per interval
	interval time.Duration // refill interval
	mu       sync.Mutex
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
// For TMDB: capacity=40, refill=40, interval=10s
func NewRateLimiter(capacity, refill int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refill:     refill,
		interval:   interval,
		lastRefill: time.Now(),
	}
}

// NewTMDBRateLimiter creates a rate limiter configured for TMDB API
// TMDB rate limit: 40 requests per 10 seconds
func NewTMDBRateLimiter() *RateLimiter {
	return NewRateLimiter(40, 40, 10*time.Second)
}

// Allow checks if a request can proceed and consumes a token
// Returns true if request is allowed, false if rate limited
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

// Wait blocks until a token is available, then consumes it
// Calculates optimal wait time instead of busy-waiting
func (rl *RateLimiter) Wait() {
	for {
		rl.mu.Lock()
		rl.refillTokens()
		
		if rl.tokens > 0 {
			rl.tokens--
			rl.mu.Unlock()
			return
		}
		
		// Calculate time until next refill while holding the lock
		timeSinceRefill := time.Since(rl.lastRefill)
		timeUntilRefill := rl.interval - timeSinceRefill
		rl.mu.Unlock()
		
		// Wait for next refill or minimum time
		if timeUntilRefill > 0 {
			time.Sleep(timeUntilRefill)
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// refillTokens adds tokens based on elapsed time since last refill
// Must be called with mutex locked
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	if elapsed >= rl.interval {
		intervals := int(elapsed / rl.interval)
		tokensToAdd := intervals * rl.refill

		rl.tokens = min(rl.capacity, rl.tokens+tokensToAdd)
		rl.lastRefill = rl.lastRefill.Add(time.Duration(intervals) * rl.interval)
	}
}

// Available returns the number of tokens currently available
func (rl *RateLimiter) Available() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()
	return rl.tokens
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
