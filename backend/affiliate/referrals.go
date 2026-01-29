package affiliate

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"sync"
	"time"
)

// ReferralManager handles user referral program
type ReferralManager struct {
	mu             sync.RWMutex
	codes          map[string]*ReferralCode
	codesByUser    map[string][]*ReferralCode
	rewards        []*ReferralReward
	socialShares   map[string]int // Track social shares by code
}

// NewReferralManager creates a new referral manager
func NewReferralManager() *ReferralManager {
	return &ReferralManager{
		codes:        make(map[string]*ReferralCode),
		codesByUser:  make(map[string][]*ReferralCode),
		rewards:      make([]*ReferralReward, 0),
		socialShares: make(map[string]int),
	}
}

// CreateReferralCode creates a new referral code for a user
func (rm *ReferralManager) CreateReferralCode(userID string, referrerBonus, refereeBonus float64, maxUses int, expiryDays int) (*ReferralCode, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Generate unique code
	code := rm.generateReferralCode(userID)

	var expiresAt *time.Time
	if expiryDays > 0 {
		expiry := time.Now().AddDate(0, 0, expiryDays)
		expiresAt = &expiry
	}

	refCode := &ReferralCode{
		ID:            int64(len(rm.codes) + 1),
		UserID:        userID,
		Code:          code,
		ReferrerBonus: referrerBonus,
		RefereeBonus:  refereeBonus,
		TotalUses:     0,
		MaxUses:       maxUses,
		ExpiresAt:     expiresAt,
		IsActive:      true,
		CreatedAt:     time.Now(),
	}

	rm.codes[code] = refCode
	rm.codesByUser[userID] = append(rm.codesByUser[userID], refCode)

	log.Printf("[Referral] Created referral code: %s for user %s (Referrer=$%.2f, Referee=$%.2f)",
		code, userID, referrerBonus, refereeBonus)

	return refCode, nil
}

// ApplyReferralCode applies a referral code when a new user signs up
func (rm *ReferralManager) ApplyReferralCode(code, newUserID string) (*ReferralReward, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Find code
	refCode, ok := rm.codes[strings.ToUpper(code)]
	if !ok {
		return nil, errors.New("invalid referral code")
	}

	// Validate code
	if !refCode.IsActive {
		return nil, errors.New("referral code is inactive")
	}

	if refCode.ExpiresAt != nil && time.Now().After(*refCode.ExpiresAt) {
		return nil, errors.New("referral code has expired")
	}

	if refCode.MaxUses > 0 && refCode.TotalUses >= int64(refCode.MaxUses) {
		return nil, errors.New("referral code has reached maximum uses")
	}

	// Check if user is trying to refer themselves
	if refCode.UserID == newUserID {
		return nil, errors.New("cannot use your own referral code")
	}

	// Check if user already used a referral code
	for _, reward := range rm.rewards {
		if reward.RefereeUserID == newUserID {
			return nil, errors.New("user has already used a referral code")
		}
	}

	// Create reward
	reward := &ReferralReward{
		ID:             int64(len(rm.rewards) + 1),
		ReferralCodeID: refCode.ID,
		ReferrerUserID: refCode.UserID,
		RefereeUserID:  newUserID,
		ReferrerReward: refCode.ReferrerBonus,
		RefereeReward:  refCode.RefereeBonus,
		Status:         "PENDING",
		CreatedAt:      time.Now(),
	}

	rm.rewards = append(rm.rewards, reward)

	// Update code usage
	refCode.TotalUses++

	log.Printf("[Referral] Applied code %s: Referrer=%s gets $%.2f, Referee=%s gets $%.2f",
		code, refCode.UserID, refCode.ReferrerBonus, newUserID, refCode.RefereeBonus)

	return reward, nil
}

// CreditReferralReward credits the reward to both users
func (rm *ReferralManager) CreditReferralReward(rewardID int64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, reward := range rm.rewards {
		if reward.ID == rewardID {
			if reward.Status != "PENDING" {
				return errors.New("reward is not pending")
			}

			reward.Status = "CREDITED"
			now := time.Now()
			reward.CreditedAt = &now

			log.Printf("[Referral] Credited reward: Referrer=%s ($%.2f), Referee=%s ($%.2f)",
				reward.ReferrerUserID, reward.ReferrerReward,
				reward.RefereeUserID, reward.RefereeReward)

			return nil
		}
	}
	return errors.New("reward not found")
}

// GetUserReferralCodes returns all referral codes for a user
func (rm *ReferralManager) GetUserReferralCodes(userID string) []*ReferralCode {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.codesByUser[userID]
}

// GetReferralStats returns referral statistics for a user
func (rm *ReferralManager) GetReferralStats(userID string) map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_codes":       0,
		"total_referrals":   0,
		"total_earned":      0.0,
		"pending_rewards":   0,
		"successful_referrals": make([]string, 0),
	}

	codes := rm.codesByUser[userID]
	totalReferrals := int64(0)
	totalEarned := 0.0
	pendingRewards := 0
	successfulReferrals := make([]string, 0)

	for _, code := range codes {
		totalReferrals += code.TotalUses
	}

	for _, reward := range rm.rewards {
		if reward.ReferrerUserID == userID {
			totalEarned += reward.ReferrerReward
			if reward.Status == "PENDING" {
				pendingRewards++
			} else {
				successfulReferrals = append(successfulReferrals, reward.RefereeUserID)
			}
		}
	}

	stats["total_codes"] = len(codes)
	stats["total_referrals"] = totalReferrals
	stats["total_earned"] = totalEarned
	stats["pending_rewards"] = pendingRewards
	stats["successful_referrals"] = successfulReferrals

	return stats
}

// DeactivateCode deactivates a referral code
func (rm *ReferralManager) DeactivateCode(code string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	refCode, ok := rm.codes[strings.ToUpper(code)]
	if !ok {
		return errors.New("code not found")
	}

	refCode.IsActive = false
	log.Printf("[Referral] Deactivated code: %s", code)
	return nil
}

// TrackSocialShare tracks social media shares
func (rm *ReferralManager) TrackSocialShare(code, platform string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	refCode, ok := rm.codes[strings.ToUpper(code)]
	if !ok {
		return errors.New("code not found")
	}

	key := code + ":" + platform
	rm.socialShares[key]++

	log.Printf("[Referral] Social share tracked: Code=%s, Platform=%s, Total=%d",
		code, platform, rm.socialShares[key])

	_ = refCode
	return nil
}

// GetTopReferrers returns top referrers by total referrals
func (rm *ReferralManager) GetTopReferrers(limit int) []map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Count referrals per user
	userReferralCount := make(map[string]int64)
	userEarnings := make(map[string]float64)

	for _, reward := range rm.rewards {
		if reward.Status == "CREDITED" {
			userReferralCount[reward.ReferrerUserID]++
			userEarnings[reward.ReferrerUserID] += reward.ReferrerReward
		}
	}

	// Create sorted list
	type referrerData struct {
		UserID    string
		Referrals int64
		Earnings  float64
	}

	referrers := make([]referrerData, 0)
	for userID, count := range userReferralCount {
		referrers = append(referrers, referrerData{
			UserID:    userID,
			Referrals: count,
			Earnings:  userEarnings[userID],
		})
	}

	// Sort by referral count
	for i := 0; i < len(referrers); i++ {
		for j := i + 1; j < len(referrers); j++ {
			if referrers[j].Referrals > referrers[i].Referrals {
				referrers[i], referrers[j] = referrers[j], referrers[i]
			}
		}
	}

	// Limit results
	if limit > 0 && limit < len(referrers) {
		referrers = referrers[:limit]
	}

	// Convert to map
	result := make([]map[string]interface{}, len(referrers))
	for i, r := range referrers {
		result[i] = map[string]interface{}{
			"user_id":   r.UserID,
			"referrals": r.Referrals,
			"earnings":  r.Earnings,
			"rank":      i + 1,
		}
	}

	return result
}

// GenerateShareableLink generates a shareable referral link
func (rm *ReferralManager) GenerateShareableLink(code, platform string) string {
	baseURL := "https://yourbroker.com/signup"

	switch platform {
	case "facebook":
		return "https://www.facebook.com/sharer/sharer.php?u=" + baseURL + "?ref=" + code
	case "twitter":
		text := "Join the best trading platform! Sign up with my referral code: " + code
		return "https://twitter.com/intent/tweet?text=" + strings.ReplaceAll(text, " ", "%20") + "&url=" + baseURL + "?ref=" + code
	case "whatsapp":
		text := "Join the best trading platform! Sign up with my referral code: " + code + " " + baseURL + "?ref=" + code
		return "https://wa.me/?text=" + strings.ReplaceAll(text, " ", "%20")
	case "email":
		subject := "Join me on this amazing trading platform!"
		body := "I've been trading with this amazing platform and I think you'll love it too! Sign up with my referral code: " + code + "\n\n" + baseURL + "?ref=" + code
		return "mailto:?subject=" + strings.ReplaceAll(subject, " ", "%20") + "&body=" + strings.ReplaceAll(body, " ", "%20")
	default:
		return baseURL + "?ref=" + code
	}
}

// CreateReferralContest creates a referral contest/campaign
func (rm *ReferralManager) CreateReferralContest(name string, startDate, endDate time.Time, prizes map[int]float64) error {
	// In production, store contest in database
	log.Printf("[Referral] Created contest: %s (Start=%s, End=%s, Prizes=%d)",
		name, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), len(prizes))
	return nil
}

// GetContestLeaderboard returns contest leaderboard
func (rm *ReferralManager) GetContestLeaderboard(contestID int64, startDate, endDate time.Time) []map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Count referrals per user within contest period
	userReferralCount := make(map[string]int)

	for _, reward := range rm.rewards {
		if reward.CreatedAt.After(startDate) && reward.CreatedAt.Before(endDate) && reward.Status == "CREDITED" {
			userReferralCount[reward.ReferrerUserID]++
		}
	}

	// Create sorted list
	type contestant struct {
		UserID    string
		Referrals int
	}

	contestants := make([]contestant, 0)
	for userID, count := range userReferralCount {
		contestants = append(contestants, contestant{
			UserID:    userID,
			Referrals: count,
		})
	}

	// Sort by referral count
	for i := 0; i < len(contestants); i++ {
		for j := i + 1; j < len(contestants); j++ {
			if contestants[j].Referrals > contestants[i].Referrals {
				contestants[i], contestants[j] = contestants[j], contestants[i]
			}
		}
	}

	// Convert to map
	result := make([]map[string]interface{}, len(contestants))
	for i, c := range contestants {
		result[i] = map[string]interface{}{
			"rank":      i + 1,
			"user_id":   c.UserID,
			"referrals": c.Referrals,
		}
	}

	return result
}

// generateReferralCode generates a unique referral code
func (rm *ReferralManager) generateReferralCode(userID string) string {
	for {
		// Generate 6-character code
		b := make([]byte, 4)
		rand.Read(b)
		code := base64.URLEncoding.EncodeToString(b)[:6]
		code = strings.ToUpper(strings.ReplaceAll(code, "-", ""))
		code = strings.ReplaceAll(code, "_", "")

		// Prepend first 2 chars of userID hash for personalization
		if len(userID) >= 2 {
			code = strings.ToUpper(userID[:2]) + code
		}

		// Check if unique
		if _, exists := rm.codes[code]; !exists {
			return code
		}
	}
}

// GamificationBadge represents achievement badges
type GamificationBadge struct {
	ID          int64
	Name        string
	Description string
	Icon        string
	Threshold   int // Number of referrals needed
	Reward      float64 // Bonus reward for earning badge
}

var badges = []GamificationBadge{
	{ID: 1, Name: "First Referral", Description: "Made your first referral", Icon: "🌟", Threshold: 1, Reward: 10.0},
	{ID: 2, Name: "Rising Star", Description: "Referred 5 users", Icon: "⭐", Threshold: 5, Reward: 50.0},
	{ID: 3, Name: "Influencer", Description: "Referred 10 users", Icon: "💫", Threshold: 10, Reward: 100.0},
	{ID: 4, Name: "Master Recruiter", Description: "Referred 25 users", Icon: "👑", Threshold: 25, Reward: 250.0},
	{ID: 5, Name: "Legend", Description: "Referred 50 users", Icon: "🏆", Threshold: 50, Reward: 500.0},
}

// GetUserBadges returns earned badges for a user
func (rm *ReferralManager) GetUserBadges(userID string) []GamificationBadge {
	stats := rm.GetReferralStats(userID)
	totalReferrals := stats["total_referrals"].(int64)

	earnedBadges := make([]GamificationBadge, 0)
	for _, badge := range badges {
		if int64(badge.Threshold) <= totalReferrals {
			earnedBadges = append(earnedBadges, badge)
		}
	}

	return earnedBadges
}

// GetNextBadge returns the next badge a user can earn
func (rm *ReferralManager) GetNextBadge(userID string) *GamificationBadge {
	stats := rm.GetReferralStats(userID)
	totalReferrals := stats["total_referrals"].(int64)

	for _, badge := range badges {
		if int64(badge.Threshold) > totalReferrals {
			return &badge
		}
	}

	return nil // User has all badges
}
