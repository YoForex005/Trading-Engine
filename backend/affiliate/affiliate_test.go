package affiliate

import (
	"testing"
	"time"
)

func TestProgramManager(t *testing.T) {
	pm := NewProgramManager()

	// Test creating affiliate program
	program := &AffiliateProgram{
		Name:                "Test Program",
		CommissionModel:     "HYBRID",
		CPAAmount:           100.00,
		RevSharePercent:     25.00,
		MinPayout:           50.00,
		PayoutSchedule:      "MONTHLY",
		CookieDuration:      30,
		SubAffiliateEnabled: true,
		SubAffiliatePercent: 10.00,
	}

	createdProgram, err := pm.CreateProgram(program)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}

	if createdProgram.ID == 0 {
		t.Error("Program ID should not be 0")
	}

	if createdProgram.Status != "ACTIVE" {
		t.Errorf("Expected status ACTIVE, got %s", createdProgram.Status)
	}

	// Test registering affiliate
	affiliate := &Affiliate{
		CompanyName: "Test Marketing Inc",
		ContactName: "John Doe",
		Email:       "john@test.com",
		Phone:       "+1-555-1234",
		Country:     "USA",
	}

	registeredAffiliate, err := pm.RegisterAffiliate(affiliate)
	if err != nil {
		t.Fatalf("Failed to register affiliate: %v", err)
	}

	if registeredAffiliate.AffiliateCode == "" {
		t.Error("Affiliate code should be generated")
	}

	if registeredAffiliate.Status != "PENDING" {
		t.Errorf("Expected status PENDING, got %s", registeredAffiliate.Status)
	}

	// Test approving affiliate
	err = pm.ApproveAffiliate(registeredAffiliate.ID)
	if err != nil {
		t.Fatalf("Failed to approve affiliate: %v", err)
	}

	affiliate, ok := pm.GetAffiliate(registeredAffiliate.ID)
	if !ok {
		t.Fatal("Affiliate not found")
	}

	if affiliate.Status != "ACTIVE" {
		t.Errorf("Expected status ACTIVE after approval, got %s", affiliate.Status)
	}

	// Test commission rates
	cpa, revShare, err := pm.GetCommissionRate(registeredAffiliate.ID)
	if err != nil {
		t.Fatalf("Failed to get commission rate: %v", err)
	}

	if cpa != 100.00 {
		t.Errorf("Expected CPA 100.00, got %.2f", cpa)
	}

	if revShare != 25.00 {
		t.Errorf("Expected RevShare 25.00, got %.2f", revShare)
	}

	// Test updating tier
	err = pm.UpdateAffiliateTier(registeredAffiliate.ID, 3)
	if err != nil {
		t.Fatalf("Failed to update tier: %v", err)
	}

	// Commission rates should increase with tier
	cpa, revShare, err = pm.GetCommissionRate(registeredAffiliate.ID)
	if err != nil {
		t.Fatalf("Failed to get commission rate: %v", err)
	}

	expectedCPA := 100.00 * 1.2 // Tier 3 = 1.2x multiplier
	if cpa != expectedCPA {
		t.Errorf("Expected CPA %.2f with tier multiplier, got %.2f", expectedCPA, cpa)
	}
}

func TestTrackingManager(t *testing.T) {
	pm := NewProgramManager()
	tm := NewTrackingManager()

	// Register affiliate
	affiliate := &Affiliate{
		ContactName: "Jane Smith",
		Email:       "jane@test.com",
	}
	registeredAffiliate, _ := pm.RegisterAffiliate(affiliate)
	pm.ApproveAffiliate(registeredAffiliate.ID)

	// Create tracking link
	link := &AffiliateLink{
		AffiliateID: registeredAffiliate.ID,
		Campaign:    "test-campaign",
		Source:      "facebook",
		Medium:      "cpc",
	}

	createdLink, err := tm.CreateLink(link)
	if err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}

	if createdLink.LinkCode == "" {
		t.Error("Link code should be generated")
	}

	if createdLink.FullURL == "" {
		t.Error("Full URL should be generated")
	}

	// Track click
	click, err := tm.TrackClick(
		createdLink.LinkCode,
		"192.168.1.100",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0",
		"https://google.com",
		"https://yourbroker.com/landing",
	)

	if err != nil {
		t.Fatalf("Failed to track click: %v", err)
	}

	if click.ClickID == "" {
		t.Error("Click ID should be generated")
	}

	if click.Device != "DESKTOP" {
		t.Errorf("Expected device DESKTOP, got %s", click.Device)
	}

	if !click.IsUnique {
		t.Error("First click should be unique")
	}

	// Track conversion
	conversion, err := tm.TrackConversion(
		click.ClickID,
		"user-123",
		12345,
		"SIGNUP",
		0,
	)

	if err != nil {
		t.Fatalf("Failed to track conversion: %v", err)
	}

	if conversion.ConversionType != "SIGNUP" {
		t.Errorf("Expected conversion type SIGNUP, got %s", conversion.ConversionType)
	}

	if conversion.Status != "PENDING" {
		t.Errorf("Expected status PENDING, got %s", conversion.Status)
	}

	// Test fraud detection
	// Rapid clicks from same IP should be flagged
	for i := 0; i < 150; i++ {
		tm.TrackClick(
			createdLink.LinkCode,
			"192.168.1.100",
			"Mozilla/5.0",
			"",
			"",
		)
	}

	// Check if latest click is flagged as fraudulent
	clicks := tm.GetAffiliateClicks(registeredAffiliate.ID, time.Now().Add(-1*time.Hour), time.Now())
	fraudCount := 0
	for _, c := range clicks {
		if c.IsFraudulent {
			fraudCount++
		}
	}

	if fraudCount == 0 {
		t.Error("Expected some clicks to be flagged as fraudulent")
	}
}

func TestCommissionManager(t *testing.T) {
	pm := NewProgramManager()
	tm := NewTrackingManager()
	cm := NewCommissionManager(pm, tm)

	// Setup program
	pm.CreateProgram(&AffiliateProgram{
		Name:            "Test Program",
		CommissionModel: "HYBRID",
		CPAAmount:       100.00,
		RevSharePercent: 25.00,
	})

	// Register affiliate
	affiliate := &Affiliate{
		ContactName: "Bob Johnson",
		Email:       "bob@test.com",
	}
	registeredAffiliate, _ := pm.RegisterAffiliate(affiliate)
	pm.ApproveAffiliate(registeredAffiliate.ID)

	// Test CPA commission
	cpaComm, err := cm.CalculateCPACommission(registeredAffiliate.ID, 1, 12345)
	if err != nil {
		t.Fatalf("Failed to calculate CPA commission: %v", err)
	}

	if cpaComm.Amount != 100.00 {
		t.Errorf("Expected CPA amount 100.00, got %.2f", cpaComm.Amount)
	}

	if cpaComm.CommissionType != "CPA" {
		t.Errorf("Expected type CPA, got %s", cpaComm.CommissionType)
	}

	if cpaComm.Status != "PENDING" {
		t.Errorf("Expected status PENDING, got %s", cpaComm.Status)
	}

	// Test RevShare commission
	revShareComm, err := cm.CalculateRevShareCommission(
		registeredAffiliate.ID,
		12345,
		100000.00, // $100k trading volume
		1000.00,   // $1000 trading fees
		"2024-01",
	)

	if err != nil {
		t.Fatalf("Failed to calculate RevShare commission: %v", err)
	}

	expectedRevShare := 1000.00 * 0.25 // 25% of $1000
	if revShareComm.Amount != expectedRevShare {
		t.Errorf("Expected RevShare amount %.2f, got %.2f", expectedRevShare, revShareComm.Amount)
	}

	// Test approving commission
	err = cm.ApproveCommission(cpaComm.ID)
	if err != nil {
		t.Fatalf("Failed to approve commission: %v", err)
	}

	// Check affiliate balance updated
	affiliate, _ = pm.GetAffiliate(registeredAffiliate.ID)
	if affiliate.PendingBalance != 100.00 {
		t.Errorf("Expected pending balance 100.00, got %.2f", affiliate.PendingBalance)
	}

	// Test payout
	payout, err := cm.ProcessPayout(registeredAffiliate.ID, 100.00, "BANK")
	if err != nil {
		t.Fatalf("Failed to process payout: %v", err)
	}

	if payout.Status != "PENDING" {
		t.Errorf("Expected payout status PENDING, got %s", payout.Status)
	}

	// Complete payout
	err = cm.CompletePayout(payout.ID, "TXN-123456")
	if err != nil {
		t.Fatalf("Failed to complete payout: %v", err)
	}

	// Check affiliate balances updated
	affiliate, _ = pm.GetAffiliate(registeredAffiliate.ID)
	if affiliate.PendingBalance != 0 {
		t.Errorf("Expected pending balance 0 after payout, got %.2f", affiliate.PendingBalance)
	}

	if affiliate.TotalPaid != 100.00 {
		t.Errorf("Expected total paid 100.00, got %.2f", affiliate.TotalPaid)
	}
}

func TestReferralManager(t *testing.T) {
	rm := NewReferralManager()

	// Create referral code
	code, err := rm.CreateReferralCode("user-123", 25.00, 25.00, 10, 30)
	if err != nil {
		t.Fatalf("Failed to create referral code: %v", err)
	}

	if code.Code == "" {
		t.Error("Referral code should be generated")
	}

	if code.ReferrerBonus != 25.00 {
		t.Errorf("Expected referrer bonus 25.00, got %.2f", code.ReferrerBonus)
	}

	// Apply referral code
	reward, err := rm.ApplyReferralCode(code.Code, "user-456")
	if err != nil {
		t.Fatalf("Failed to apply referral code: %v", err)
	}

	if reward.ReferrerUserID != "user-123" {
		t.Errorf("Expected referrer user-123, got %s", reward.ReferrerUserID)
	}

	if reward.RefereeUserID != "user-456" {
		t.Errorf("Expected referee user-456, got %s", reward.RefereeUserID)
	}

	// Test duplicate usage
	_, err = rm.ApplyReferralCode(code.Code, "user-456")
	if err == nil {
		t.Error("Expected error when user tries to use referral code twice")
	}

	// Test self-referral
	_, err = rm.ApplyReferralCode(code.Code, "user-123")
	if err == nil {
		t.Error("Expected error when user tries to refer themselves")
	}

	// Test max uses
	for i := 0; i < 10; i++ {
		rm.ApplyReferralCode(code.Code, "user-"+string(rune(500+i)))
	}

	_, err = rm.ApplyReferralCode(code.Code, "user-999")
	if err == nil {
		t.Error("Expected error when referral code reaches max uses")
	}

	// Test referral stats
	stats := rm.GetReferralStats("user-123")
	if stats["total_referrals"].(int64) != 10 {
		t.Errorf("Expected 10 referrals, got %d", stats["total_referrals"])
	}

	// Test gamification badges
	badges := rm.GetUserBadges("user-123")
	if len(badges) < 2 {
		t.Errorf("Expected at least 2 badges for 10 referrals, got %d", len(badges))
	}
}

func TestAnalyticsEngine(t *testing.T) {
	pm := NewProgramManager()
	tm := NewTrackingManager()
	cm := NewCommissionManager(pm, tm)
	ae := NewAnalyticsEngine(pm, tm, cm)

	// Setup
	pm.CreateProgram(&AffiliateProgram{
		Name:            "Test Program",
		CommissionModel: "CPA",
		CPAAmount:       100.00,
	})

	affiliate := &Affiliate{
		ContactName: "Analytics Test",
		Email:       "analytics@test.com",
	}
	registeredAffiliate, _ := pm.RegisterAffiliate(affiliate)
	pm.ApproveAffiliate(registeredAffiliate.ID)

	link := &AffiliateLink{
		AffiliateID: registeredAffiliate.ID,
		Campaign:    "analytics-test",
	}
	createdLink, _ := tm.CreateLink(link)

	// Generate test data
	for i := 0; i < 100; i++ {
		click, _ := tm.TrackClick(
			createdLink.LinkCode,
			"192.168.1."+string(rune(i)),
			"Mozilla/5.0",
			"",
			"",
		)

		// 10% conversion rate
		if i%10 == 0 {
			tm.TrackConversion(click.ClickID, "user-"+string(rune(i)), int64(i), "SIGNUP", 0)
			cm.CalculateCPACommission(registeredAffiliate.ID, int64(i), int64(i))
		}
	}

	// Test dashboard stats
	stats, err := ae.GetDashboardStats(registeredAffiliate.ID)
	if err != nil {
		t.Fatalf("Failed to get dashboard stats: %v", err)
	}

	if stats.TotalClicks == 0 {
		t.Error("Expected total clicks > 0")
	}

	// Test conversion funnel
	funnel := ae.GetConversionFunnel(
		registeredAffiliate.ID,
		time.Now().Add(-1*time.Hour),
		time.Now(),
	)

	if funnel.TotalClicks == 0 {
		t.Error("Expected total clicks in funnel")
	}

	if funnel.ClickToSignup == 0 {
		t.Error("Expected click-to-signup conversion rate")
	}

	// Test time series data
	timeSeries := ae.GetTimeSeriesData(
		registeredAffiliate.ID,
		time.Now().Add(-24*time.Hour),
		time.Now(),
		"hour",
	)

	if len(timeSeries) == 0 {
		t.Error("Expected time series data")
	}
}

func BenchmarkClickTracking(b *testing.B) {
	pm := NewProgramManager()
	tm := NewTrackingManager()

	affiliate := &Affiliate{
		ContactName: "Bench Test",
		Email:       "bench@test.com",
	}
	registeredAffiliate, _ := pm.RegisterAffiliate(affiliate)

	link := &AffiliateLink{
		AffiliateID: registeredAffiliate.ID,
		Campaign:    "benchmark",
	}
	createdLink, _ := tm.CreateLink(link)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.TrackClick(
			createdLink.LinkCode,
			"192.168.1.1",
			"Mozilla/5.0",
			"",
			"",
		)
	}
}

func BenchmarkCommissionCalculation(b *testing.B) {
	pm := NewProgramManager()
	tm := NewTrackingManager()
	cm := NewCommissionManager(pm, tm)

	pm.CreateProgram(&AffiliateProgram{
		Name:            "Bench Program",
		CommissionModel: "REVSHARE",
		RevSharePercent: 25.00,
	})

	affiliate := &Affiliate{
		ContactName: "Bench Test",
		Email:       "bench@test.com",
	}
	registeredAffiliate, _ := pm.RegisterAffiliate(affiliate)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.CalculateRevShareCommission(
			registeredAffiliate.ID,
			12345,
			100000.00,
			1000.00,
			"2024-01",
		)
	}
}
