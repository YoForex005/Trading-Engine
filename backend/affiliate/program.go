package affiliate

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ProgramManager handles affiliate program management
type ProgramManager struct {
	mu               sync.RWMutex
	programs         map[int64]*AffiliateProgram
	affiliates       map[int64]*Affiliate
	affiliatesByCode map[string]*Affiliate
	activeProgram    *AffiliateProgram
}

// NewProgramManager creates a new program manager
func NewProgramManager() *ProgramManager {
	return &ProgramManager{
		programs:         make(map[int64]*AffiliateProgram),
		affiliates:       make(map[int64]*Affiliate),
		affiliatesByCode: make(map[string]*Affiliate),
	}
}

// CreateProgram creates a new affiliate program
func (pm *ProgramManager) CreateProgram(program *AffiliateProgram) (*AffiliateProgram, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Validate commission model
	if program.CommissionModel != "CPA" && program.CommissionModel != "REVSHARE" && program.CommissionModel != "HYBRID" {
		return nil, errors.New("invalid commission model, must be CPA, REVSHARE, or HYBRID")
	}

	// Validate amounts
	if program.CommissionModel == "CPA" || program.CommissionModel == "HYBRID" {
		if program.CPAAmount <= 0 {
			return nil, errors.New("CPA amount must be greater than 0")
		}
	}

	if program.CommissionModel == "REVSHARE" || program.CommissionModel == "HYBRID" {
		if program.RevSharePercent <= 0 || program.RevSharePercent > 100 {
			return nil, errors.New("RevShare percent must be between 0 and 100")
		}
	}

	// Generate ID
	program.ID = int64(len(pm.programs) + 1)
	program.CreatedAt = time.Now()
	program.UpdatedAt = time.Now()
	program.Status = "ACTIVE"

	pm.programs[program.ID] = program

	// Set as active program if first or only active
	if pm.activeProgram == nil || program.Status == "ACTIVE" {
		pm.activeProgram = program
	}

	log.Printf("[Affiliate] Created program: %s (ID=%d, Model=%s)", program.Name, program.ID, program.CommissionModel)
	return program, nil
}

// GetProgram retrieves a program by ID
func (pm *ProgramManager) GetProgram(id int64) (*AffiliateProgram, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	program, ok := pm.programs[id]
	return program, ok
}

// GetActiveProgram returns the currently active program
func (pm *ProgramManager) GetActiveProgram() *AffiliateProgram {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.activeProgram
}

// RegisterAffiliate registers a new affiliate
func (pm *ProgramManager) RegisterAffiliate(affiliate *Affiliate) (*Affiliate, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Validate email
	if affiliate.Email == "" {
		return nil, errors.New("email is required")
	}

	// Generate unique affiliate code
	if affiliate.AffiliateCode == "" {
		affiliate.AffiliateCode = pm.generateAffiliateCode()
	}

	// Check if code already exists
	if _, exists := pm.affiliatesByCode[affiliate.AffiliateCode]; exists {
		return nil, errors.New("affiliate code already exists")
	}

	// Generate ID
	affiliate.ID = int64(len(pm.affiliates) + 1)
	affiliate.Status = "PENDING" // Requires approval
	affiliate.Tier = 1 // Default tier
	affiliate.CreatedAt = time.Now()
	affiliate.UpdatedAt = time.Now()
	affiliate.LastActivityAt = time.Now()

	// Set default commission model from active program
	if pm.activeProgram != nil && affiliate.CommissionModel == "" {
		affiliate.CommissionModel = pm.activeProgram.CommissionModel
	}

	pm.affiliates[affiliate.ID] = affiliate
	pm.affiliatesByCode[affiliate.AffiliateCode] = affiliate

	log.Printf("[Affiliate] Registered new affiliate: %s (Code=%s, Email=%s)",
		affiliate.ContactName, affiliate.AffiliateCode, affiliate.Email)

	return affiliate, nil
}

// ApproveAffiliate approves a pending affiliate
func (pm *ProgramManager) ApproveAffiliate(affiliateID int64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	affiliate, ok := pm.affiliates[affiliateID]
	if !ok {
		return errors.New("affiliate not found")
	}

	if affiliate.Status == "ACTIVE" {
		return errors.New("affiliate already active")
	}

	affiliate.Status = "ACTIVE"
	affiliate.UpdatedAt = time.Now()

	log.Printf("[Affiliate] Approved affiliate: %s (ID=%d)", affiliate.ContactName, affiliate.ID)
	return nil
}

// SuspendAffiliate suspends an affiliate
func (pm *ProgramManager) SuspendAffiliate(affiliateID int64, reason string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	affiliate, ok := pm.affiliates[affiliateID]
	if !ok {
		return errors.New("affiliate not found")
	}

	affiliate.Status = "SUSPENDED"
	affiliate.UpdatedAt = time.Now()

	log.Printf("[Affiliate] Suspended affiliate: %s (ID=%d, Reason=%s)",
		affiliate.ContactName, affiliate.ID, reason)
	return nil
}

// BanAffiliate permanently bans an affiliate
func (pm *ProgramManager) BanAffiliate(affiliateID int64, reason string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	affiliate, ok := pm.affiliates[affiliateID]
	if !ok {
		return errors.New("affiliate not found")
	}

	affiliate.Status = "BANNED"
	affiliate.UpdatedAt = time.Now()

	log.Printf("[Affiliate] Banned affiliate: %s (ID=%d, Reason=%s)",
		affiliate.ContactName, affiliate.ID, reason)
	return nil
}

// GetAffiliate retrieves an affiliate by ID
func (pm *ProgramManager) GetAffiliate(id int64) (*Affiliate, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	affiliate, ok := pm.affiliates[id]
	return affiliate, ok
}

// GetAffiliateByCode retrieves an affiliate by their code
func (pm *ProgramManager) GetAffiliateByCode(code string) (*Affiliate, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	affiliate, ok := pm.affiliatesByCode[strings.ToUpper(code)]
	return affiliate, ok
}

// ListAffiliates returns all affiliates
func (pm *ProgramManager) ListAffiliates(status string) []*Affiliate {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []*Affiliate
	for _, aff := range pm.affiliates {
		if status == "" || aff.Status == status {
			result = append(result, aff)
		}
	}
	return result
}

// UpdateAffiliateTier updates an affiliate's commission tier
func (pm *ProgramManager) UpdateAffiliateTier(affiliateID int64, tier int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	affiliate, ok := pm.affiliates[affiliateID]
	if !ok {
		return errors.New("affiliate not found")
	}

	if tier < 1 || tier > 5 {
		return errors.New("tier must be between 1 and 5")
	}

	affiliate.Tier = tier
	affiliate.UpdatedAt = time.Now()

	log.Printf("[Affiliate] Updated tier for %s to %d", affiliate.ContactName, tier)
	return nil
}

// SetCustomCommission sets custom commission rates for an affiliate
func (pm *ProgramManager) SetCustomCommission(affiliateID int64, cpa *float64, revShare *float64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	affiliate, ok := pm.affiliates[affiliateID]
	if !ok {
		return errors.New("affiliate not found")
	}

	affiliate.CustomCPA = cpa
	affiliate.CustomRevShare = revShare
	affiliate.UpdatedAt = time.Now()

	log.Printf("[Affiliate] Set custom commission for %s (CPA=%.2f, RevShare=%.2f%%)",
		affiliate.ContactName,
		cpa != nil && *cpa != 0,
		revShare != nil && *revShare != 0)
	return nil
}

// UpdatePayoutMethod updates affiliate's payout method
func (pm *ProgramManager) UpdatePayoutMethod(affiliateID int64, method, bankDetails, cryptoAddress string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	affiliate, ok := pm.affiliates[affiliateID]
	if !ok {
		return errors.New("affiliate not found")
	}

	// Validate method
	validMethods := []string{"BANK", "PAYPAL", "CRYPTO", "WIRE"}
	valid := false
	for _, vm := range validMethods {
		if method == vm {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("invalid payout method")
	}

	affiliate.PayoutMethod = method
	affiliate.BankDetails = bankDetails
	affiliate.CryptoAddress = cryptoAddress
	affiliate.UpdatedAt = time.Now()

	log.Printf("[Affiliate] Updated payout method for %s to %s", affiliate.ContactName, method)
	return nil
}

// AddSubAffiliate adds a sub-affiliate relationship
func (pm *ProgramManager) AddSubAffiliate(parentID int64, subAffiliate *Affiliate) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	parent, ok := pm.affiliates[parentID]
	if !ok {
		return errors.New("parent affiliate not found")
	}

	if parent.Status != "ACTIVE" {
		return errors.New("parent affiliate must be active")
	}

	// Check if program allows sub-affiliates
	if pm.activeProgram != nil && !pm.activeProgram.SubAffiliateEnabled {
		return errors.New("sub-affiliate program is not enabled")
	}

	subAffiliate.ParentAffiliateID = &parentID
	return nil
}

// GetSubAffiliates returns all sub-affiliates for a parent
func (pm *ProgramManager) GetSubAffiliates(parentID int64) []*Affiliate {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var subs []*Affiliate
	for _, aff := range pm.affiliates {
		if aff.ParentAffiliateID != nil && *aff.ParentAffiliateID == parentID {
			subs = append(subs, aff)
		}
	}
	return subs
}

// UpdateStats updates affiliate statistics
func (pm *ProgramManager) UpdateStats(affiliateID int64, clicks, signups, deposits int64, earnings float64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	affiliate, ok := pm.affiliates[affiliateID]
	if !ok {
		return errors.New("affiliate not found")
	}

	affiliate.LifetimeClicks += clicks
	affiliate.LifetimeSignups += signups
	affiliate.LifetimeDeposits += deposits
	affiliate.TotalEarnings += earnings
	affiliate.PendingBalance += earnings

	// Calculate conversion rate
	if affiliate.LifetimeClicks > 0 {
		affiliate.ConversionRate = float64(affiliate.LifetimeSignups) / float64(affiliate.LifetimeClicks) * 100
	}

	affiliate.LastActivityAt = time.Now()
	affiliate.UpdatedAt = time.Now()

	return nil
}

// generateAffiliateCode generates a unique affiliate code
func (pm *ProgramManager) generateAffiliateCode() string {
	for {
		// Generate 8-character code
		b := make([]byte, 6)
		rand.Read(b)
		code := base64.URLEncoding.EncodeToString(b)[:8]
		code = strings.ToUpper(strings.ReplaceAll(code, "-", "X"))
		code = strings.ReplaceAll(code, "_", "Y")

		// Check if unique
		if _, exists := pm.affiliatesByCode[code]; !exists {
			return code
		}
	}
}

// GetCommissionRate returns the commission rate for an affiliate
func (pm *ProgramManager) GetCommissionRate(affiliateID int64) (cpa float64, revShare float64, err error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	affiliate, ok := pm.affiliates[affiliateID]
	if !ok {
		return 0, 0, errors.New("affiliate not found")
	}

	// Use custom rates if set
	if affiliate.CustomCPA != nil {
		cpa = *affiliate.CustomCPA
	} else if pm.activeProgram != nil {
		cpa = pm.activeProgram.CPAAmount
	}

	if affiliate.CustomRevShare != nil {
		revShare = *affiliate.CustomRevShare
	} else if pm.activeProgram != nil {
		revShare = pm.activeProgram.RevSharePercent
	}

	// Apply tier multiplier (tier 2 = 1.1x, tier 3 = 1.2x, etc.)
	if affiliate.Tier > 1 {
		multiplier := 1.0 + float64(affiliate.Tier-1)*0.1
		cpa *= multiplier
		revShare *= multiplier
	}

	return cpa, revShare, nil
}

// GetLeaderboard returns top affiliates by earnings
func (pm *ProgramManager) GetLeaderboard(limit int, period string) []*Affiliate {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Clone and sort by total earnings
	affiliates := make([]*Affiliate, 0, len(pm.affiliates))
	for _, aff := range pm.affiliates {
		if aff.Status == "ACTIVE" {
			affiliates = append(affiliates, aff)
		}
	}

	// Simple bubble sort by total earnings (descending)
	for i := 0; i < len(affiliates); i++ {
		for j := i + 1; j < len(affiliates); j++ {
			if affiliates[j].TotalEarnings > affiliates[i].TotalEarnings {
				affiliates[i], affiliates[j] = affiliates[j], affiliates[i]
			}
		}
	}

	if limit > 0 && limit < len(affiliates) {
		return affiliates[:limit]
	}
	return affiliates
}

// ExportReport generates an affiliate performance report
func (pm *ProgramManager) ExportReport(affiliateID int64, startDate, endDate time.Time) (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	affiliate, ok := pm.affiliates[affiliateID]
	if !ok {
		return "", errors.New("affiliate not found")
	}

	report := fmt.Sprintf("Affiliate Performance Report\n")
	report += fmt.Sprintf("================================\n")
	report += fmt.Sprintf("Affiliate: %s\n", affiliate.ContactName)
	report += fmt.Sprintf("Code: %s\n", affiliate.AffiliateCode)
	report += fmt.Sprintf("Period: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	report += fmt.Sprintf("\nLifetime Statistics:\n")
	report += fmt.Sprintf("  Total Clicks: %d\n", affiliate.LifetimeClicks)
	report += fmt.Sprintf("  Total Signups: %d\n", affiliate.LifetimeSignups)
	report += fmt.Sprintf("  Total Deposits: %d\n", affiliate.LifetimeDeposits)
	report += fmt.Sprintf("  Conversion Rate: %.2f%%\n", affiliate.ConversionRate)
	report += fmt.Sprintf("\nEarnings:\n")
	report += fmt.Sprintf("  Total Earned: $%.2f\n", affiliate.TotalEarnings)
	report += fmt.Sprintf("  Total Paid: $%.2f\n", affiliate.TotalPaid)
	report += fmt.Sprintf("  Pending Balance: $%.2f\n", affiliate.PendingBalance)

	return report, nil
}
