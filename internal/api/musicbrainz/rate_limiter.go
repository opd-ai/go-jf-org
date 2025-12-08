package musicbrainz

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter for MusicBrainz API
// MusicBrainz allows 1 request per second
type RateLimiter struct {
	tokens   int
	capacity int
	refill   int           // tokens to add per interval
	interval time.Duration // refill interval
	mu       sync.Mutex
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
// For MusicBrainz: capacity=1, refill=1, interval=1s
func NewRateLimiter(capacity, refill int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refill:     refill,
		interval:   interval,
		lastRefill: time.Now(),
	}
}

// NewMusicBrainzRateLimiter creates a rate limiter configured for MusicBrainz API
// MusicBrainz rate limit: 1 request per second
func NewMusicBrainzRateLimiter() *RateLimiter {
	return NewRateLimiter(1, 1, 1*time.Second)
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
