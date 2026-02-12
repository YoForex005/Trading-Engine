package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Rate limit tier configurations
const (
	// STRICT tier - authentication and security-critical operations
	LoginPerMinute           = 5
	LoginPerHour             = 20
	TwoFactorPerMinute       = 3
	TwoFactorPerHour         = 10
	PasswordChangePerMinute  = 3
	PasswordChangePerHour    = 10

	// MODERATE tier - financial operations
	DepositPerMinute      = 10
	DepositPerHour        = 100
	WithdrawalPerMinute   = 5
	WithdrawalPerHour     = 20
	AdminUsersPerMinute   = 10
	AdminUsersPerHour     = 200

	// STANDARD tier - data operations
	ExportPerMinute     = 5
	ExportPerHour       = 30
	AuditQueryPerMinute = 20
	AuditQueryPerHour   = 500

	// Lockout thresholds
	MaxLoginAttempts      = 10
	LoginLockoutDuration  = 15 * time.Minute
	Max2FAAttempts        = 5
	TwoFactorLockoutDuration = 5 * time.Minute

	// Cleanup interval
	CleanupInterval = 5 * time.Minute
)

// RateLimitTier defines the type of rate limiting to apply
type RateLimitTier string

const (
	TierStrictLogin       RateLimitTier = "strict_login"
	TierStrict2FA         RateLimitTier = "strict_2fa"
	TierStrictPassword    RateLimitTier = "strict_password"
	TierModerateDeposit   RateLimitTier = "moderate_deposit"
	TierModerateWithdraw  RateLimitTier = "moderate_withdraw"
	TierModerateAdminUser RateLimitTier = "moderate_admin_user"
	TierStandardExport    RateLimitTier = "standard_export"
	TierStandardAudit     RateLimitTier = "standard_audit"
)

// rateLimitConfig holds the configuration for a specific tier
type rateLimitConfig struct {
	perMinute int
	perHour   int
}

// getTierConfig returns the rate limit configuration for a tier
func getTierConfig(tier RateLimitTier) rateLimitConfig {
	configs := map[RateLimitTier]rateLimitConfig{
		TierStrictLogin:       {LoginPerMinute, LoginPerHour},
		TierStrict2FA:         {TwoFactorPerMinute, TwoFactorPerHour},
		TierStrictPassword:    {PasswordChangePerMinute, PasswordChangePerHour},
		TierModerateDeposit:   {DepositPerMinute, DepositPerHour},
		TierModerateWithdraw:  {WithdrawalPerMinute, WithdrawalPerHour},
		TierModerateAdminUser: {AdminUsersPerMinute, AdminUsersPerHour},
		TierStandardExport:    {ExportPerMinute, ExportPerHour},
		TierStandardAudit:     {AuditQueryPerMinute, AuditQueryPerHour},
	}
	return configs[tier]
}

// requestRecord tracks a single request timestamp
type requestRecord struct {
	timestamp time.Time
	success   bool // true = successful request, false = failed attempt (for lockout tracking)
}

// clientLimits tracks rate limits for a single client (IP or user)
type clientLimits struct {
	requests      []requestRecord
	lockedUntil   time.Time
	failedAttempts int
	mu            sync.Mutex
}

// SensitiveRateLimiter implements sliding window rate limiting for sensitive endpoints
type SensitiveRateLimiter struct {
	ipLimits   map[string]map[RateLimitTier]*clientLimits
	userLimits map[string]map[RateLimitTier]*clientLimits
	mu         sync.RWMutex
}

// NewSensitiveRateLimiter creates a new rate limiter for sensitive endpoints
func NewSensitiveRateLimiter() *SensitiveRateLimiter {
	limiter := &SensitiveRateLimiter{
		ipLimits:   make(map[string]map[RateLimitTier]*clientLimits),
		userLimits: make(map[string]map[RateLimitTier]*clientLimits),
	}

	// Start cleanup goroutine
	go limiter.cleanupExpired()

	return limiter
}

// CheckIPLimit checks if an IP has exceeded the rate limit for a tier
// Returns (allowed bool, retryAfter int, isLocked bool)
func (srl *SensitiveRateLimiter) CheckIPLimit(ip string, tier RateLimitTier) (bool, int, bool) {
	return srl.checkLimit(ip, tier, true)
}

// CheckUserLimit checks if a user has exceeded the rate limit for a tier
// Returns (allowed bool, retryAfter int, isLocked bool)
func (srl *SensitiveRateLimiter) CheckUserLimit(userID string, tier RateLimitTier) (bool, int, bool) {
	return srl.checkLimit(userID, tier, false)
}

// RecordIPFailure records a failed attempt for an IP (for lockout tracking)
func (srl *SensitiveRateLimiter) RecordIPFailure(ip string, tier RateLimitTier) {
	srl.recordFailure(ip, tier, true)
}

// RecordUserFailure records a failed attempt for a user (for lockout tracking)
func (srl *SensitiveRateLimiter) RecordUserFailure(userID string, tier RateLimitTier) {
	srl.recordFailure(userID, tier, false)
}

// checkLimit is the internal implementation for rate limit checking
func (srl *SensitiveRateLimiter) checkLimit(key string, tier RateLimitTier, isIP bool) (bool, int, bool) {
	srl.mu.Lock()
	defer srl.mu.Unlock()

	config := getTierConfig(tier)
	now := time.Now()

	// Get or create client limits
	var limitsMap map[string]map[RateLimitTier]*clientLimits
	if isIP {
		limitsMap = srl.ipLimits
	} else {
		limitsMap = srl.userLimits
	}

	if limitsMap[key] == nil {
		limitsMap[key] = make(map[RateLimitTier]*clientLimits)
	}
	if limitsMap[key][tier] == nil {
		limitsMap[key][tier] = &clientLimits{
			requests: make([]requestRecord, 0),
		}
	}

	limits := limitsMap[key][tier]
	limits.mu.Lock()
	defer limits.mu.Unlock()

	// Check if locked out
	if now.Before(limits.lockedUntil) {
		retryAfter := int(time.Until(limits.lockedUntil).Seconds())
		return false, retryAfter, true
	}

	// Remove expired requests (older than 1 hour)
	oneHourAgo := now.Add(-1 * time.Hour)
	oneMinuteAgo := now.Add(-1 * time.Minute)

	validRequests := make([]requestRecord, 0)
	for _, req := range limits.requests {
		if req.timestamp.After(oneHourAgo) {
			validRequests = append(validRequests, req)
		}
	}
	limits.requests = validRequests

	// Count requests in last minute and hour
	countMinute := 0
	countHour := 0
	for _, req := range limits.requests {
		if req.timestamp.After(oneMinuteAgo) {
			countMinute++
		}
		countHour++
	}

	// Check limits
	if countMinute >= config.perMinute {
		retryAfter := int(time.Until(limits.requests[len(limits.requests)-config.perMinute].timestamp.Add(1 * time.Minute)).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, retryAfter, false
	}

	if countHour >= config.perHour {
		retryAfter := int(time.Until(limits.requests[0].timestamp.Add(1 * time.Hour)).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, retryAfter, false
	}

	// Record this request
	limits.requests = append(limits.requests, requestRecord{
		timestamp: now,
		success:   true,
	})

	return true, 0, false
}

// recordFailure records a failed attempt and potentially locks out the client
func (srl *SensitiveRateLimiter) recordFailure(key string, tier RateLimitTier, isIP bool) {
	srl.mu.Lock()
	defer srl.mu.Unlock()

	now := time.Now()

	// Get or create client limits
	var limitsMap map[string]map[RateLimitTier]*clientLimits
	if isIP {
		limitsMap = srl.ipLimits
	} else {
		limitsMap = srl.userLimits
	}

	if limitsMap[key] == nil {
		limitsMap[key] = make(map[RateLimitTier]*clientLimits)
	}
	if limitsMap[key][tier] == nil {
		limitsMap[key][tier] = &clientLimits{
			requests: make([]requestRecord, 0),
		}
	}

	limits := limitsMap[key][tier]
	limits.mu.Lock()
	defer limits.mu.Unlock()

	// Increment failed attempts
	limits.failedAttempts++

	// Check if lockout threshold reached
	if tier == TierStrictLogin && limits.failedAttempts >= MaxLoginAttempts {
		limits.lockedUntil = now.Add(LoginLockoutDuration)
		limits.failedAttempts = 0 // Reset counter after lockout
	} else if tier == TierStrict2FA && limits.failedAttempts >= Max2FAAttempts {
		limits.lockedUntil = now.Add(TwoFactorLockoutDuration)
		limits.failedAttempts = 0 // Reset counter after lockout
	}
}

// ResetFailures resets failed attempt counter (call on successful auth)
func (srl *SensitiveRateLimiter) ResetFailures(key string, tier RateLimitTier, isIP bool) {
	srl.mu.Lock()
	defer srl.mu.Unlock()

	var limitsMap map[string]map[RateLimitTier]*clientLimits
	if isIP {
		limitsMap = srl.ipLimits
	} else {
		limitsMap = srl.userLimits
	}

	if limitsMap[key] != nil && limitsMap[key][tier] != nil {
		limits := limitsMap[key][tier]
		limits.mu.Lock()
		limits.failedAttempts = 0
		limits.mu.Unlock()
	}
}

// cleanupExpired removes expired entries periodically
func (srl *SensitiveRateLimiter) cleanupExpired() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		srl.mu.Lock()
		now := time.Now()
		oneHourAgo := now.Add(-1 * time.Hour)

		// Cleanup IP limits
		for ip, tiers := range srl.ipLimits {
			for tier, limits := range tiers {
				limits.mu.Lock()
				// Remove if no recent requests and not locked
				if len(limits.requests) == 0 || (limits.requests[len(limits.requests)-1].timestamp.Before(oneHourAgo) && now.After(limits.lockedUntil)) {
					delete(tiers, tier)
				}
				limits.mu.Unlock()
			}
			if len(tiers) == 0 {
				delete(srl.ipLimits, ip)
			}
		}

		// Cleanup user limits
		for userID, tiers := range srl.userLimits {
			for tier, limits := range tiers {
				limits.mu.Lock()
				// Remove if no recent requests and not locked
				if len(limits.requests) == 0 || (limits.requests[len(limits.requests)-1].timestamp.Before(oneHourAgo) && now.After(limits.lockedUntil)) {
					delete(tiers, tier)
				}
				limits.mu.Unlock()
			}
			if len(tiers) == 0 {
				delete(srl.userLimits, userID)
			}
		}

		srl.mu.Unlock()
	}
}

// RateLimitResponse is the structured JSON error response
type RateLimitResponse struct {
	Error      string `json:"error"`
	RetryAfter int    `json:"retryAfter"`
	Message    string `json:"message"`
	IsLocked   bool   `json:"isLocked,omitempty"`
}

// SensitiveRateLimitMiddleware returns middleware for rate limiting based on tier
// keyExtractor should return the client identifier (IP or user ID)
func (srl *SensitiveRateLimiter) SensitiveRateLimitMiddleware(tier RateLimitTier, useIP bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var key string
			if useIP {
				key = getClientIP(r)
			} else {
				// Extract user ID from context (set by auth middleware)
				if userID := r.Context().Value("userID"); userID != nil {
					key = userID.(string)
				} else {
					// Fall back to IP if no user ID
					key = getClientIP(r)
				}
			}

			var allowed bool
			var retryAfter int
			var isLocked bool

			if useIP {
				allowed, retryAfter, isLocked = srl.CheckIPLimit(key, tier)
			} else {
				allowed, retryAfter, isLocked = srl.CheckUserLimit(key, tier)
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				w.WriteHeader(http.StatusTooManyRequests)

				message := "Rate limit exceeded. Please try again later."
				if isLocked {
					message = "Account temporarily locked due to too many failed attempts."
				}

				json.NewEncoder(w).Encode(RateLimitResponse{
					Error:      "rate_limited",
					RetryAfter: retryAfter,
					Message:    message,
					IsLocked:   isLocked,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
