package affiliate

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// CommissionManager handles commission calculation and payouts
type CommissionManager struct {
	mu              sync.RWMutex
	commissions     []*Commission
	payouts         []*Payout
	programManager  *ProgramManager
	trackingManager *TrackingManager
}

// NewCommissionManager creates a new commission manager
func NewCommissionManager(pm *ProgramManager, tm *TrackingManager) *CommissionManager {
	return &CommissionManager{
		commissions:     make([]*Commission, 0),
		payouts:         make([]*Payout, 0),
		programManager:  pm,
		trackingManager: tm,
	}
}

// CalculateCPACommission calculates CPA (Cost Per Acquisition) commission
func (cm *CommissionManager) CalculateCPACommission(affiliateID int64, conversionID int64, accountID int64) (*Commission, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Get affiliate commission rate
	cpa, _, err := cm.programManager.GetCommissionRate(affiliateID)
	if err != nil {
		return nil, err
	}

	if cpa <= 0 {
		return nil, errors.New("CPA commission not configured for this affiliate")
	}

	commission := &Commission{
		ID:             int64(len(cm.commissions) + 1),
		AffiliateID:    affiliateID,
		ConversionID:   &conversionID,
		AccountID:      accountID,
		CommissionType: "CPA",
		Amount:         cpa,
		Currency:       "USD",
		Description:    fmt.Sprintf("CPA commission for new account signup"),
		Status:         "PENDING",
		CreatedAt:      time.Now(),
	}

	cm.commissions = append(cm.commissions, commission)

	log.Printf("[Commission] CPA commission created: Affiliate=%d, Amount=%.2f", affiliateID, cpa)
	return commission, nil
}

// CalculateRevShareCommission calculates revenue share commission
func (cm *CommissionManager) CalculateRevShareCommission(affiliateID int64, accountID int64, tradingVolume, tradingFees float64, period string) (*Commission, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Get affiliate commission rate
	_, revShare, err := cm.programManager.GetCommissionRate(affiliateID)
	if err != nil {
		return nil, err
	}

	if revShare <= 0 {
		return nil, errors.New("RevShare commission not configured for this affiliate")
	}

	// Calculate commission amount (percentage of trading fees)
	amount := tradingFees * (revShare / 100.0)

	commission := &Commission{
		ID:             int64(len(cm.commissions) + 1),
		AffiliateID:    affiliateID,
		AccountID:      accountID,
		CommissionType: "REVSHARE",
		Amount:         amount,
		Currency:       "USD",
		Description:    fmt.Sprintf("RevShare commission for period %s (%.2f%% of $%.2f fees)", period, revShare, tradingFees),
		Period:         period,
		TradingVolume:  tradingVolume,
		TradingFees:    tradingFees,
		Status:         "PENDING",
		CreatedAt:      time.Now(),
	}

	cm.commissions = append(cm.commissions, commission)

	log.Printf("[Commission] RevShare commission created: Affiliate=%d, Period=%s, Amount=%.2f (%.2f%% of %.2f)",
		affiliateID, period, amount, revShare, tradingFees)
	return commission, nil
}

// CalculateHybridCommission calculates both CPA and RevShare
func (cm *CommissionManager) CalculateHybridCommission(affiliateID int64, conversionID int64, accountID int64, tradingFees float64) ([]*Commission, error) {
	commissions := make([]*Commission, 0)

	// Calculate CPA
	cpaComm, err := cm.CalculateCPACommission(affiliateID, conversionID, accountID)
	if err == nil {
		commissions = append(commissions, cpaComm)
	}

	// Calculate RevShare
	period := time.Now().Format("2006-01")
	revShareComm, err := cm.CalculateRevShareCommission(affiliateID, accountID, 0, tradingFees, period)
	if err == nil {
		commissions = append(commissions, revShareComm)
	}

	return commissions, nil
}

// CalculateSubAffiliateCommission calculates commission for sub-affiliates
func (cm *CommissionManager) CalculateSubAffiliateCommission(parentAffiliateID int64, subAffiliateID int64, originalCommission float64) (*Commission, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	program := cm.programManager.GetActiveProgram()
	if program == nil || !program.SubAffiliateEnabled {
		return nil, errors.New("sub-affiliate program not enabled")
	}

	// Calculate sub-affiliate commission (percentage of original commission)
	amount := originalCommission * (program.SubAffiliatePercent / 100.0)

	commission := &Commission{
		ID:             int64(len(cm.commissions) + 1),
		AffiliateID:    parentAffiliateID,
		CommissionType: "SUB_AFFILIATE",
		Amount:         amount,
		Currency:       "USD",
		Description:    fmt.Sprintf("Sub-affiliate commission (%.2f%% of $%.2f)", program.SubAffiliatePercent, originalCommission),
		Status:         "PENDING",
		CreatedAt:      time.Now(),
	}

	cm.commissions = append(cm.commissions, commission)

	log.Printf("[Commission] Sub-affiliate commission created: Parent=%d, Amount=%.2f", parentAffiliateID, amount)
	return commission, nil
}

// ApproveCommission approves a pending commission
func (cm *CommissionManager) ApproveCommission(commissionID int64) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, comm := range cm.commissions {
		if comm.ID == commissionID {
			if comm.Status != "PENDING" {
				return errors.New("commission is not pending")
			}

			comm.Status = "APPROVED"

			// Update affiliate pending balance
			if err := cm.programManager.UpdateStats(comm.AffiliateID, 0, 0, 0, comm.Amount); err != nil {
				return err
			}

			log.Printf("[Commission] Approved commission: ID=%d, Affiliate=%d, Amount=%.2f",
				commissionID, comm.AffiliateID, comm.Amount)
			return nil
		}
	}
	return errors.New("commission not found")
}

// ReverseCommission reverses a paid commission
func (cm *CommissionManager) ReverseCommission(commissionID int64, reason string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, comm := range cm.commissions {
		if comm.ID == commissionID {
			if comm.Status != "PAID" {
				return errors.New("commission is not paid")
			}

			comm.Status = "REVERSED"

			// Update affiliate balances
			if err := cm.programManager.UpdateStats(comm.AffiliateID, 0, 0, 0, -comm.Amount); err != nil {
				return err
			}

			log.Printf("[Commission] Reversed commission: ID=%d, Affiliate=%d, Amount=%.2f, Reason=%s",
				commissionID, comm.AffiliateID, comm.Amount, reason)
			return nil
		}
	}
	return errors.New("commission not found")
}

// GetAffiliateCommissions returns all commissions for an affiliate
func (cm *CommissionManager) GetAffiliateCommissions(affiliateID int64, status string, startDate, endDate time.Time) []*Commission {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []*Commission
	for _, comm := range cm.commissions {
		if comm.AffiliateID == affiliateID &&
			(status == "" || comm.Status == status) &&
			comm.CreatedAt.After(startDate) &&
			comm.CreatedAt.Before(endDate) {
			result = append(result, comm)
		}
	}
	return result
}

// ProcessPayout processes a payout to an affiliate
func (cm *CommissionManager) ProcessPayout(affiliateID int64, amount float64, method string) (*Payout, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	affiliate, ok := cm.programManager.GetAffiliate(affiliateID)
	if !ok {
		return nil, errors.New("affiliate not found")
	}

	// Check minimum payout
	program := cm.programManager.GetActiveProgram()
	if program != nil && amount < program.MinPayout {
		return nil, fmt.Errorf("amount %.2f is below minimum payout %.2f", amount, program.MinPayout)
	}

	// Check affiliate balance
	if affiliate.PendingBalance < amount {
		return nil, errors.New("insufficient balance")
	}

	// Create payout
	payout := &Payout{
		ID:          int64(len(cm.payouts) + 1),
		AffiliateID: affiliateID,
		Amount:      amount,
		Currency:    "USD",
		Method:      method,
		Status:      "PENDING",
		CreatedAt:   time.Now(),
	}

	// Set payout details based on method
	switch method {
	case "BANK":
		payout.BankDetails = affiliate.BankDetails
	case "CRYPTO":
		payout.CryptoAddress = affiliate.CryptoAddress
	case "PAYPAL":
		// Use email for PayPal
		payout.BankDetails = affiliate.Email
	}

	// Update affiliate balances
	affiliate.PendingBalance -= amount
	affiliate.UpdatedAt = time.Now()

	// Update commission statuses
	remainingAmount := amount
	for _, comm := range cm.commissions {
		if comm.AffiliateID == affiliateID && comm.Status == "APPROVED" && remainingAmount > 0 {
			comm.Status = "PAID"
			comm.PayoutID = &payout.ID
			now := time.Now()
			comm.PaidAt = &now
			remainingAmount -= comm.Amount
		}
	}

	cm.payouts = append(cm.payouts, payout)

	log.Printf("[Payout] Created payout: ID=%d, Affiliate=%d, Amount=%.2f, Method=%s",
		payout.ID, affiliateID, amount, method)

	return payout, nil
}

// CompletePayout marks a payout as completed
func (cm *CommissionManager) CompletePayout(payoutID int64, transactionRef string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, payout := range cm.payouts {
		if payout.ID == payoutID {
			if payout.Status == "COMPLETED" {
				return errors.New("payout already completed")
			}

			payout.Status = "COMPLETED"
			payout.TransactionRef = transactionRef
			now := time.Now()
			payout.CompletedAt = &now

			// Update affiliate total paid
			affiliate, ok := cm.programManager.GetAffiliate(payout.AffiliateID)
			if ok {
				affiliate.TotalPaid += payout.Amount
				affiliate.UpdatedAt = time.Now()
			}

			log.Printf("[Payout] Completed payout: ID=%d, Affiliate=%d, Amount=%.2f, Ref=%s",
				payoutID, payout.AffiliateID, payout.Amount, transactionRef)
			return nil
		}
	}
	return errors.New("payout not found")
}

// FailPayout marks a payout as failed and restores balance
func (cm *CommissionManager) FailPayout(payoutID int64, reason string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, payout := range cm.payouts {
		if payout.ID == payoutID {
			if payout.Status == "COMPLETED" {
				return errors.New("cannot fail completed payout")
			}

			payout.Status = "FAILED"
			payout.Notes = reason

			// Restore affiliate balance
			affiliate, ok := cm.programManager.GetAffiliate(payout.AffiliateID)
			if ok {
				affiliate.PendingBalance += payout.Amount
				affiliate.UpdatedAt = time.Now()
			}

			// Restore commission statuses
			for _, comm := range cm.commissions {
				if comm.PayoutID != nil && *comm.PayoutID == payoutID {
					comm.Status = "APPROVED"
					comm.PayoutID = nil
					comm.PaidAt = nil
				}
			}

			log.Printf("[Payout] Failed payout: ID=%d, Affiliate=%d, Amount=%.2f, Reason=%s",
				payoutID, payout.AffiliateID, payout.Amount, reason)
			return nil
		}
	}
	return errors.New("payout not found")
}

// GetAffiliatePayout returns all payouts for an affiliate
func (cm *CommissionManager) GetAffiliatePayouts(affiliateID int64) []*Payout {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []*Payout
	for _, payout := range cm.payouts {
		if payout.AffiliateID == affiliateID {
			result = append(result, payout)
		}
	}
	return result
}

// CalculateMonthlyRevShare calculates monthly revenue share for all affiliates
func (cm *CommissionManager) CalculateMonthlyRevShare(period string, accountTradingData map[int64]struct {
	Volume float64
	Fees   float64
}) error {
	log.Printf("[Commission] Calculating monthly RevShare for period: %s", period)

	// Get all active affiliates
	affiliates := cm.programManager.ListAffiliates("ACTIVE")

	for _, affiliate := range affiliates {
		// Get all referred accounts
		// (In production, track this in database)
		totalFees := 0.0
		totalVolume := 0.0

		// Calculate total fees for affiliate's referred accounts
		for accountID, data := range accountTradingData {
			// Check if account was referred by this affiliate
			// (Simplified - in production check conversions table)
			_ = accountID
			totalFees += data.Fees
			totalVolume += data.Volume
		}

		if totalFees > 0 {
			// Calculate commission
			_, err := cm.CalculateRevShareCommission(
				affiliate.ID,
				0, // Use first account for simplicity
				totalVolume,
				totalFees,
				period,
			)
			if err != nil {
				log.Printf("[Commission] Error calculating RevShare for affiliate %d: %v", affiliate.ID, err)
			}
		}
	}

	log.Printf("[Commission] Monthly RevShare calculation completed for period: %s", period)
	return nil
}

// AutoApprovePendingCommissions auto-approves commissions based on criteria
func (cm *CommissionManager) AutoApprovePendingCommissions(minDaysOld int) int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -minDaysOld)
	approved := 0

	for _, comm := range cm.commissions {
		if comm.Status == "PENDING" && comm.CreatedAt.Before(cutoff) {
			comm.Status = "APPROVED"

			// Update affiliate balance
			if err := cm.programManager.UpdateStats(comm.AffiliateID, 0, 0, 0, comm.Amount); err == nil {
				approved++
			}
		}
	}

	log.Printf("[Commission] Auto-approved %d commissions older than %d days", approved, minDaysOld)
	return approved
}

// GetPayoutReport generates a payout report
func (cm *CommissionManager) GetPayoutReport(startDate, endDate time.Time) map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	report := map[string]interface{}{
		"period":    fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"total_payouts": 0,
		"total_amount":  0.0,
		"by_status":     make(map[string]int),
		"by_method":     make(map[string]float64),
	}

	totalPayouts := 0
	totalAmount := 0.0
	byStatus := make(map[string]int)
	byMethod := make(map[string]float64)

	for _, payout := range cm.payouts {
		if payout.CreatedAt.After(startDate) && payout.CreatedAt.Before(endDate) {
			totalPayouts++
			totalAmount += payout.Amount
			byStatus[payout.Status]++
			byMethod[payout.Method] += payout.Amount
		}
	}

	report["total_payouts"] = totalPayouts
	report["total_amount"] = totalAmount
	report["by_status"] = byStatus
	report["by_method"] = byMethod

	return report
}
