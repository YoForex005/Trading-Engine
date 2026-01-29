package fix

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// RateLimitTier defines rate limit configurations
type RateLimitTier struct {
	Name              string        `json:"name"`
	OrdersPerSecond   int           `json:"orders_per_second"`
	MessagesPerSecond int           `json:"messages_per_second"`
	MaxSessions       int           `json:"max_sessions"`
	BurstSize         int           `json:"burst_size"` // Allows short bursts
}

var (
	// Predefined rate limit tiers
	TierBasic = RateLimitTier{
		Name:              "basic",
		OrdersPerSecond:   5,
		MessagesPerSecond: 20,
		MaxSessions:       1,
		BurstSize:         10,
	}

	TierStandard = RateLimitTier{
		Name:              "standard",
		OrdersPerSecond:   20,
		MessagesPerSecond: 100,
		MaxSessions:       3,
		BurstSize:         40,
	}

	TierPremium = RateLimitTier{
		Name:              "premium",
		OrdersPerSecond:   100,
		MessagesPerSecond: 500,
		MaxSessions:       10,
		BurstSize:         200,
	}

	TierUnlimited = RateLimitTier{
		Name:              "unlimited",
		OrdersPerSecond:   10000,
		MessagesPerSecond: 50000,
		MaxSessions:       100,
		BurstSize:         10000,
	}
)

// RateLimiter manages rate limiting for FIX sessions
type RateLimiter struct {
	tiers        map[string]*RateLimitTier
	userLimits   map[string]*userRateState
	sessionCount map[string]int // userID -> active session count
	mu           sync.RWMutex
	auditLogger  AuditLogger
}

// userRateState tracks rate limiting state for a user
type userRateState struct {
	userID            string
	tier              *RateLimitTier
	orderTokens       int
	messageTokens     int
	lastOrderRefill   time.Time
	lastMessageRefill time.Time
	violations        int
	lastViolation     time.Time
	// PERFORMANCE FIX #2: Integer arithmetic for token precision
	// Accumulate fractional nanoseconds to prevent precision loss over time
	// Previous float math: 0.9s × 5 orders/s = 4.5 → truncates to 4 (lost tokens)
	orderNanos        int64 // Fractional nanoseconds for order tokens
	messageNanos      int64 // Fractional nanoseconds for message tokens
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(auditLogger AuditLogger) *RateLimiter {
	rl := &RateLimiter{
		tiers:        make(map[string]*RateLimitTier),
		userLimits:   make(map[string]*userRateState),
		sessionCount: make(map[string]int),
		auditLogger:  auditLogger,
	}

	// Register default tiers
	rl.tiers["basic"] = &TierBasic
	rl.tiers["standard"] = &TierStandard
	rl.tiers["premium"] = &TierPremium
	rl.tiers["unlimited"] = &TierUnlimited

	return rl
}

// InitializeUser sets up rate limiting for a user
func (rl *RateLimiter) InitializeUser(userID, tierName string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	tier, exists := rl.tiers[tierName]
	if !exists {
		return fmt.Errorf("unknown tier: %s", tierName)
	}

	now := time.Now()
	rl.userLimits[userID] = &userRateState{
		userID:            userID,
		tier:              tier,
		orderTokens:       tier.BurstSize,
		messageTokens:     tier.BurstSize * 5, // Messages get higher burst
		lastOrderRefill:   now,
		lastMessageRefill: now,
		violations:        0,
	}

	rl.auditLogger.LogCredentialOperation("rate_limit_init", userID, fmt.Sprintf("tier=%s", tierName), true)
	return nil
}

// CheckOrderLimit checks if user can send an order
func (rl *RateLimiter) CheckOrderLimit(userID string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	state, exists := rl.userLimits[userID]
	if !exists {
		return false, errors.New("user not initialized")
	}

	// Refill tokens based on time elapsed
	rl.refillTokens(state)

	// Check if user has tokens available
	if state.orderTokens > 0 {
		state.orderTokens--
		return true, nil
	}

	// Rate limit exceeded
	state.violations++
	state.lastViolation = time.Now()

	rl.auditLogger.LogCredentialOperation("rate_limit_exceeded", userID,
		fmt.Sprintf("type=order violations=%d", state.violations), false)

	return false, errors.New("order rate limit exceeded")
}

// CheckMessageLimit checks if user can send a message
func (rl *RateLimiter) CheckMessageLimit(userID string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	state, exists := rl.userLimits[userID]
	if !exists {
		return false, errors.New("user not initialized")
	}

	// Refill tokens based on time elapsed
	rl.refillTokens(state)

	// Check if user has tokens available
	if state.messageTokens > 0 {
		state.messageTokens--
		return true, nil
	}

	// Rate limit exceeded
	state.violations++
	state.lastViolation = time.Now()

	rl.auditLogger.LogCredentialOperation("rate_limit_exceeded", userID,
		fmt.Sprintf("type=message violations=%d", state.violations), false)

	return false, errors.New("message rate limit exceeded")
}

// CheckSessionLimit checks if user can create a new session
func (rl *RateLimiter) CheckSessionLimit(userID string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	state, exists := rl.userLimits[userID]
	if !exists {
		return false, errors.New("user not initialized")
	}

	currentSessions := rl.sessionCount[userID]
	if currentSessions >= state.tier.MaxSessions {
		rl.auditLogger.LogCredentialOperation("session_limit_exceeded", userID,
			fmt.Sprintf("current=%d max=%d", currentSessions, state.tier.MaxSessions), false)
		return false, fmt.Errorf("session limit exceeded: %d/%d", currentSessions, state.tier.MaxSessions)
	}

	return true, nil
}

// IncrementSessionCount increments the active session count for a user
func (rl *RateLimiter) IncrementSessionCount(userID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.sessionCount[userID]++
	rl.auditLogger.LogCredentialOperation("session_opened", userID,
		fmt.Sprintf("count=%d", rl.sessionCount[userID]), true)
}

// DecrementSessionCount decrements the active session count for a user
func (rl *RateLimiter) DecrementSessionCount(userID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if count := rl.sessionCount[userID]; count > 0 {
		rl.sessionCount[userID]--
		rl.auditLogger.LogCredentialOperation("session_closed", userID,
			fmt.Sprintf("count=%d", rl.sessionCount[userID]), true)
	}
}

// GetSessionCount returns the current session count for a user
func (rl *RateLimiter) GetSessionCount(userID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.sessionCount[userID]
}

// refillTokens refills tokens based on elapsed time (token bucket algorithm)
func (rl *RateLimiter) refillTokens(state *userRateState) {
	now := time.Now()

	// PERFORMANCE FIX #2: Integer arithmetic prevents precision loss over time
	// Previous float math: 0.9s × 5 orders/s = 4.5 → truncates to 4 (lost tokens)
	// New approach: Track nanoseconds, no precision loss

	// Calculate order tokens using integer arithmetic
	elapsed := now.Sub(state.lastOrderRefill).Nanoseconds()
	state.orderNanos += elapsed

	// tokens = (nanos × rate) / 1e9
	orderTokensToAdd := (state.orderNanos * int64(state.tier.OrdersPerSecond)) / 1_000_000_000
	if orderTokensToAdd > 0 {
		state.orderTokens += int(orderTokensToAdd)
		if state.orderTokens > state.tier.BurstSize {
			state.orderTokens = state.tier.BurstSize
		}
		// Update remaining nanoseconds after token addition
		state.orderNanos -= orderTokensToAdd * 1_000_000_000 / int64(state.tier.OrdersPerSecond)
		state.lastOrderRefill = now
	}

	// Calculate message tokens using integer arithmetic
	messageElapsed := now.Sub(state.lastMessageRefill).Nanoseconds()
	state.messageNanos += messageElapsed

	// tokens = (nanos × rate) / 1e9
	messageTokensToAdd := (state.messageNanos * int64(state.tier.MessagesPerSecond)) / 1_000_000_000
	if messageTokensToAdd > 0 {
		state.messageTokens += int(messageTokensToAdd)
		maxMessageBurst := state.tier.BurstSize * 5
		if state.messageTokens > maxMessageBurst {
			state.messageTokens = maxMessageBurst
		}
		// Update remaining nanoseconds after token addition
		state.messageNanos -= messageTokensToAdd * 1_000_000_000 / int64(state.tier.MessagesPerSecond)
		state.lastMessageRefill = now
	}
}

// UpdateUserTier changes a user's rate limit tier
func (rl *RateLimiter) UpdateUserTier(userID, newTierName string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	state, exists := rl.userLimits[userID]
	if !exists {
		return errors.New("user not initialized")
	}

	newTier, exists := rl.tiers[newTierName]
	if !exists {
		return fmt.Errorf("unknown tier: %s", newTierName)
	}

	oldTierName := state.tier.Name
	state.tier = newTier
	state.orderTokens = newTier.BurstSize
	state.messageTokens = newTier.BurstSize * 5

	rl.auditLogger.LogCredentialOperation("tier_updated", userID,
		fmt.Sprintf("from=%s to=%s", oldTierName, newTierName), true)

	return nil
}

// GetUserState returns the current rate limit state for a user
func (rl *RateLimiter) GetUserState(userID string) (*UserRateLimitState, error) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	state, exists := rl.userLimits[userID]
	if !exists {
		return nil, errors.New("user not initialized")
	}

	// Refill tokens to get current state
	rl.refillTokens(state)

	return &UserRateLimitState{
		UserID:            userID,
		Tier:              state.tier.Name,
		OrdersPerSecond:   state.tier.OrdersPerSecond,
		MessagesPerSecond: state.tier.MessagesPerSecond,
		MaxSessions:       state.tier.MaxSessions,
		CurrentSessions:   rl.sessionCount[userID],
		AvailableOrderTokens:   state.orderTokens,
		AvailableMessageTokens: state.messageTokens,
		Violations:        state.violations,
		LastViolation:     state.lastViolation,
	}, nil
}

// UserRateLimitState represents the current rate limit state for a user
type UserRateLimitState struct {
	UserID                 string    `json:"user_id"`
	Tier                   string    `json:"tier"`
	OrdersPerSecond        int       `json:"orders_per_second"`
	MessagesPerSecond      int       `json:"messages_per_second"`
	MaxSessions            int       `json:"max_sessions"`
	CurrentSessions        int       `json:"current_sessions"`
	AvailableOrderTokens   int       `json:"available_order_tokens"`
	AvailableMessageTokens int       `json:"available_message_tokens"`
	Violations             int       `json:"violations"`
	LastViolation          time.Time `json:"last_violation,omitempty"`
}

// ResetViolations resets violation count for a user
func (rl *RateLimiter) ResetViolations(userID string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	state, exists := rl.userLimits[userID]
	if !exists {
		return errors.New("user not initialized")
	}

	state.violations = 0
	rl.auditLogger.LogCredentialOperation("violations_reset", userID, "reset to 0", true)

	return nil
}

// AddCustomTier adds a custom rate limit tier
func (rl *RateLimiter) AddCustomTier(tier *RateLimitTier) error {
	if tier.Name == "" {
		return errors.New("tier name cannot be empty")
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.tiers[tier.Name] = tier
	rl.auditLogger.LogCredentialOperation("tier_added", "system",
		fmt.Sprintf("tier=%s", tier.Name), true)

	return nil
}

// ListTiers returns all available rate limit tiers
func (rl *RateLimiter) ListTiers() []*RateLimitTier {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	tiers := make([]*RateLimitTier, 0, len(rl.tiers))
	for _, tier := range rl.tiers {
		tiers = append(tiers, tier)
	}

	return tiers
}

// GetAllUserStates returns rate limit states for all users
func (rl *RateLimiter) GetAllUserStates() []*UserRateLimitState {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	states := make([]*UserRateLimitState, 0, len(rl.userLimits))
	for userID, state := range rl.userLimits {
		states = append(states, &UserRateLimitState{
			UserID:                 userID,
			Tier:                   state.tier.Name,
			OrdersPerSecond:        state.tier.OrdersPerSecond,
			MessagesPerSecond:      state.tier.MessagesPerSecond,
			MaxSessions:            state.tier.MaxSessions,
			CurrentSessions:        rl.sessionCount[userID],
			AvailableOrderTokens:   state.orderTokens,
			AvailableMessageTokens: state.messageTokens,
			Violations:             state.violations,
			LastViolation:          state.lastViolation,
		})
	}

	return states
}
