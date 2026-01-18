package notifications

import (
	"sync"
	"time"
)

// RateLimiter implements rate limiting for notifications
type RateLimiter struct {
	configs map[NotificationChannel]RateLimitConfig
	buckets map[string]*rateBucket
	mu      sync.RWMutex
}

// rateBucket tracks usage for a user+channel combination
type rateBucket struct {
	perMinute []time.Time
	perHour   []time.Time
	perDay    []time.Time
	mu        sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(configs map[NotificationChannel]RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		configs: configs,
		buckets: make(map[string]*rateBucket),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a notification is allowed under rate limits
func (rl *RateLimiter) Allow(userID string, channel NotificationChannel) bool {
	config, exists := rl.configs[channel]
	if !exists {
		// No rate limit configured, allow
		return true
	}

	bucketKey := userID + ":" + string(channel)

	rl.mu.Lock()
	bucket, exists := rl.buckets[bucketKey]
	if !exists {
		bucket = &rateBucket{
			perMinute: make([]time.Time, 0),
			perHour:   make([]time.Time, 0),
			perDay:    make([]time.Time, 0),
		}
		rl.buckets[bucketKey] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()

	// Clean old entries
	bucket.perMinute = rl.filterOld(bucket.perMinute, now.Add(-time.Minute))
	bucket.perHour = rl.filterOld(bucket.perHour, now.Add(-time.Hour))
	bucket.perDay = rl.filterOld(bucket.perDay, now.Add(-24*time.Hour))

	// Check limits
	if config.MaxPerMinute > 0 && len(bucket.perMinute) >= config.MaxPerMinute {
		return false
	}
	if config.MaxPerHour > 0 && len(bucket.perHour) >= config.MaxPerHour {
		return false
	}
	if config.MaxPerDay > 0 && len(bucket.perDay) >= config.MaxPerDay {
		return false
	}

	// Add current timestamp
	bucket.perMinute = append(bucket.perMinute, now)
	bucket.perHour = append(bucket.perHour, now)
	bucket.perDay = append(bucket.perDay, now)

	return true
}

// filterOld removes timestamps older than the cutoff
func (rl *RateLimiter) filterOld(timestamps []time.Time, cutoff time.Time) []time.Time {
	result := make([]time.Time, 0, len(timestamps))
	for _, t := range timestamps {
		if t.After(cutoff) {
			result = append(result, t)
		}
	}
	return result
}

// GetUsage returns current usage for a user+channel
func (rl *RateLimiter) GetUsage(userID string, channel NotificationChannel) (perMinute, perHour, perDay int) {
	bucketKey := userID + ":" + string(channel)

	rl.mu.RLock()
	bucket, exists := rl.buckets[bucketKey]
	rl.mu.RUnlock()

	if !exists {
		return 0, 0, 0
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	bucket.perMinute = rl.filterOld(bucket.perMinute, now.Add(-time.Minute))
	bucket.perHour = rl.filterOld(bucket.perHour, now.Add(-time.Hour))
	bucket.perDay = rl.filterOld(bucket.perDay, now.Add(-24*time.Hour))

	return len(bucket.perMinute), len(bucket.perHour), len(bucket.perDay)
}

// Reset resets rate limits for a user+channel
func (rl *RateLimiter) Reset(userID string, channel NotificationChannel) {
	bucketKey := userID + ":" + string(channel)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.buckets, bucketKey)
}

// cleanup periodically removes old buckets
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()

		now := time.Now()
		cutoff := now.Add(-25 * time.Hour) // Keep 1 hour extra

		for key, bucket := range rl.buckets {
			bucket.mu.Lock()

			// If all lists are empty or very old, remove bucket
			if len(bucket.perDay) == 0 ||
			   (len(bucket.perDay) > 0 && bucket.perDay[len(bucket.perDay)-1].Before(cutoff)) {
				delete(rl.buckets, key)
			}

			bucket.mu.Unlock()
		}

		rl.mu.Unlock()
	}
}

// DefaultRateLimitConfigs returns sensible default rate limits
func DefaultRateLimitConfigs() map[NotificationChannel]RateLimitConfig {
	return map[NotificationChannel]RateLimitConfig{
		ChannelEmail: {
			Channel:      ChannelEmail,
			MaxPerMinute: 5,
			MaxPerHour:   20,
			MaxPerDay:    100,
		},
		ChannelSMS: {
			Channel:      ChannelSMS,
			MaxPerMinute: 2,  // SMS is expensive
			MaxPerHour:   10,
			MaxPerDay:    30,
		},
		ChannelPush: {
			Channel:      ChannelPush,
			MaxPerMinute: 10,
			MaxPerHour:   50,
			MaxPerDay:    200,
		},
		ChannelWebhook: {
			Channel:      ChannelWebhook,
			MaxPerMinute: 10,
			MaxPerHour:   100,
			MaxPerDay:    1000,
		},
		ChannelInApp: {
			Channel:      ChannelInApp,
			MaxPerMinute: 20,
			MaxPerHour:   100,
			MaxPerDay:    500,
		},
	}
}
