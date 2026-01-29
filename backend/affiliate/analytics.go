package affiliate

import (
	"fmt"
	"sync"
	"time"
)

// AnalyticsEngine handles affiliate performance analytics
type AnalyticsEngine struct {
	mu              sync.RWMutex
	programManager  *ProgramManager
	trackingManager *TrackingManager
	commissionMgr   *CommissionManager
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine(pm *ProgramManager, tm *TrackingManager, cm *CommissionManager) *AnalyticsEngine {
	return &AnalyticsEngine{
		programManager:  pm,
		trackingManager: tm,
		commissionMgr:   cm,
	}
}

// GetDashboardStats returns real-time dashboard statistics
func (ae *AnalyticsEngine) GetDashboardStats(affiliateID int64) (*AffiliateStats, error) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	stats := &AffiliateStats{
		AffiliateID: affiliateID,
	}

	// Get affiliate
	affiliate, ok := ae.programManager.GetAffiliate(affiliateID)
	if !ok {
		return nil, fmt.Errorf("affiliate not found")
	}

	// Lifetime stats from affiliate record
	stats.TotalClicks = affiliate.LifetimeClicks
	stats.TotalSignups = affiliate.LifetimeSignups
	stats.TotalDeposits = affiliate.LifetimeDeposits
	stats.TotalEarnings = affiliate.TotalEarnings
	stats.PendingBalance = affiliate.PendingBalance
	stats.ConversionRate = affiliate.ConversionRate

	// Calculate available balance (approved but not paid)
	approvedCommissions := ae.commissionMgr.GetAffiliateCommissions(
		affiliateID,
		"APPROVED",
		time.Time{},
		time.Now(),
	)
	availableBalance := 0.0
	for _, comm := range approvedCommissions {
		availableBalance += comm.Amount
	}
	stats.AvailableBalance = availableBalance

	// Today's stats
	todayClicks := ae.trackingManager.GetAffiliateClicks(affiliateID, todayStart, now)
	stats.TodayClicks = int64(len(todayClicks))

	todayConversions := ae.trackingManager.GetAffiliateConversions(affiliateID, todayStart, now)
	stats.TodaySignups = int64(len(todayConversions))

	todayCommissions := ae.commissionMgr.GetAffiliateCommissions(affiliateID, "", todayStart, now)
	for _, comm := range todayCommissions {
		stats.TodayEarnings += comm.Amount
		if comm.ConversionID != nil {
			stats.TodayDeposits++
		}
	}

	// Week stats
	weekClicks := ae.trackingManager.GetAffiliateClicks(affiliateID, weekStart, now)
	stats.WeekClicks = int64(len(weekClicks))

	weekConversions := ae.trackingManager.GetAffiliateConversions(affiliateID, weekStart, now)
	stats.WeekSignups = int64(len(weekConversions))

	weekCommissions := ae.commissionMgr.GetAffiliateCommissions(affiliateID, "", weekStart, now)
	for _, comm := range weekCommissions {
		stats.WeekEarnings += comm.Amount
		if comm.ConversionID != nil {
			stats.WeekDeposits++
		}
	}

	// Month stats
	monthClicks := ae.trackingManager.GetAffiliateClicks(affiliateID, monthStart, now)
	stats.MonthClicks = int64(len(monthClicks))

	monthConversions := ae.trackingManager.GetAffiliateConversions(affiliateID, monthStart, now)
	stats.MonthSignups = int64(len(monthConversions))

	monthCommissions := ae.commissionMgr.GetAffiliateCommissions(affiliateID, "", monthStart, now)
	for _, comm := range monthCommissions {
		stats.MonthEarnings += comm.Amount
		if comm.ConversionID != nil {
			stats.MonthDeposits++
		}
	}

	return stats, nil
}

// ConversionFunnelData represents funnel analysis
type ConversionFunnelData struct {
	TotalClicks       int64   `json:"totalClicks"`
	UniqueClicks      int64   `json:"uniqueClicks"`
	Signups           int64   `json:"signups"`
	Deposits          int64   `json:"deposits"`
	FirstTrades       int64   `json:"firstTrades"`
	ClickToSignup     float64 `json:"clickToSignup"`     // %
	SignupToDeposit   float64 `json:"signupToDeposit"`   // %
	DepositToTrade    float64 `json:"depositToTrade"`    // %
	OverallConversion float64 `json:"overallConversion"` // %
}

// GetConversionFunnel analyzes the conversion funnel
func (ae *AnalyticsEngine) GetConversionFunnel(affiliateID int64, startDate, endDate time.Time) *ConversionFunnelData {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	funnel := &ConversionFunnelData{}

	// Get clicks
	clicks := ae.trackingManager.GetAffiliateClicks(affiliateID, startDate, endDate)
	funnel.TotalClicks = int64(len(clicks))

	uniqueClicks := int64(0)
	for _, click := range clicks {
		if click.IsUnique {
			uniqueClicks++
		}
	}
	funnel.UniqueClicks = uniqueClicks

	// Get conversions
	conversions := ae.trackingManager.GetAffiliateConversions(affiliateID, startDate, endDate)
	for _, conv := range conversions {
		switch conv.ConversionType {
		case "SIGNUP":
			funnel.Signups++
		case "DEPOSIT":
			funnel.Deposits++
		case "FIRST_TRADE":
			funnel.FirstTrades++
		}
	}

	// Calculate conversion rates
	if funnel.TotalClicks > 0 {
		funnel.ClickToSignup = float64(funnel.Signups) / float64(funnel.TotalClicks) * 100
	}
	if funnel.Signups > 0 {
		funnel.SignupToDeposit = float64(funnel.Deposits) / float64(funnel.Signups) * 100
	}
	if funnel.Deposits > 0 {
		funnel.DepositToTrade = float64(funnel.FirstTrades) / float64(funnel.Deposits) * 100
	}
	if funnel.TotalClicks > 0 {
		funnel.OverallConversion = float64(funnel.FirstTrades) / float64(funnel.TotalClicks) * 100
	}

	return funnel
}

// TrafficSourceData represents traffic by source
type TrafficSourceData struct {
	Source      string  `json:"source"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
	Revenue     float64 `json:"revenue"`
	ROI         float64 `json:"roi"`
}

// GetTrafficSources analyzes traffic by source
func (ae *AnalyticsEngine) GetTrafficSources(affiliateID int64, startDate, endDate time.Time) []TrafficSourceData {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	sourceMap := make(map[string]*TrafficSourceData)

	// Aggregate clicks by source
	clicks := ae.trackingManager.GetAffiliateClicks(affiliateID, startDate, endDate)
	for _, click := range clicks {
		source := click.Referrer
		if source == "" {
			source = "Direct"
		}

		if _, ok := sourceMap[source]; !ok {
			sourceMap[source] = &TrafficSourceData{Source: source}
		}
		sourceMap[source].Clicks++
	}

	// Add conversion data
	conversions := ae.trackingManager.GetAffiliateConversions(affiliateID, startDate, endDate)
	for _, conv := range conversions {
		// Find click to get source
		for _, click := range clicks {
			if click.ClickID == conv.ClickID {
				source := click.Referrer
				if source == "" {
					source = "Direct"
				}
				if _, ok := sourceMap[source]; ok {
					sourceMap[source].Conversions++
				}
				break
			}
		}
	}

	// Add commission data
	commissions := ae.commissionMgr.GetAffiliateCommissions(affiliateID, "", startDate, endDate)
	for _, comm := range commissions {
		// Simplified: distribute evenly across sources
		// In production, track source per commission
		for _, data := range sourceMap {
			if data.Conversions > 0 {
				data.Revenue += comm.Amount / float64(len(sourceMap))
			}
		}
	}

	// Convert to slice
	result := make([]TrafficSourceData, 0, len(sourceMap))
	for _, data := range sourceMap {
		// Calculate ROI (simplified - assumes $0 cost)
		if data.Clicks > 0 {
			data.ROI = (data.Revenue / float64(data.Clicks)) * 100
		}
		result = append(result, *data)
	}

	return result
}

// GeoData represents geographic performance
type GeoData struct {
	Country     string  `json:"country"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
	Revenue     float64 `json:"revenue"`
}

// GetGeographicPerformance analyzes performance by country
func (ae *AnalyticsEngine) GetGeographicPerformance(affiliateID int64, startDate, endDate time.Time) []GeoData {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	geoMap := make(map[string]*GeoData)

	// Aggregate clicks by country
	clicks := ae.trackingManager.GetAffiliateClicks(affiliateID, startDate, endDate)
	for _, click := range clicks {
		country := click.Country
		if country == "" {
			country = "Unknown"
		}

		if _, ok := geoMap[country]; !ok {
			geoMap[country] = &GeoData{Country: country}
		}
		geoMap[country].Clicks++
	}

	// Add conversion data
	conversions := ae.trackingManager.GetAffiliateConversions(affiliateID, startDate, endDate)
	for _, conv := range conversions {
		// Find click to get country
		for _, click := range clicks {
			if click.ClickID == conv.ClickID {
				country := click.Country
				if country == "" {
					country = "Unknown"
				}
				if _, ok := geoMap[country]; ok {
					geoMap[country].Conversions++
				}
				break
			}
		}
	}

	// Convert to slice and sort by clicks
	result := make([]GeoData, 0, len(geoMap))
	for _, data := range geoMap {
		result = append(result, *data)
	}

	// Sort by clicks (descending)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Clicks > result[i].Clicks {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// TimeSeriesData represents performance over time
type TimeSeriesData struct {
	Date        string  `json:"date"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
	Revenue     float64 `json:"revenue"`
}

// GetTimeSeriesData returns performance data over time
func (ae *AnalyticsEngine) GetTimeSeriesData(affiliateID int64, startDate, endDate time.Time, granularity string) []TimeSeriesData {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	dataMap := make(map[string]*TimeSeriesData)

	// Generate date keys based on granularity
	dateFormat := "2006-01-02"
	if granularity == "month" {
		dateFormat = "2006-01"
	} else if granularity == "hour" {
		dateFormat = "2006-01-02 15:00"
	}

	// Initialize all dates in range
	current := startDate
	for current.Before(endDate) {
		dateKey := current.Format(dateFormat)
		dataMap[dateKey] = &TimeSeriesData{Date: dateKey}

		switch granularity {
		case "hour":
			current = current.Add(1 * time.Hour)
		case "month":
			current = current.AddDate(0, 1, 0)
		default: // day
			current = current.AddDate(0, 0, 1)
		}
	}

	// Aggregate clicks
	clicks := ae.trackingManager.GetAffiliateClicks(affiliateID, startDate, endDate)
	for _, click := range clicks {
		dateKey := click.CreatedAt.Format(dateFormat)
		if data, ok := dataMap[dateKey]; ok {
			data.Clicks++
		}
	}

	// Aggregate conversions
	conversions := ae.trackingManager.GetAffiliateConversions(affiliateID, startDate, endDate)
	for _, conv := range conversions {
		dateKey := conv.CreatedAt.Format(dateFormat)
		if data, ok := dataMap[dateKey]; ok {
			data.Conversions++
		}
	}

	// Aggregate revenue
	commissions := ae.commissionMgr.GetAffiliateCommissions(affiliateID, "", startDate, endDate)
	for _, comm := range commissions {
		dateKey := comm.CreatedAt.Format(dateFormat)
		if data, ok := dataMap[dateKey]; ok {
			data.Revenue += comm.Amount
		}
	}

	// Convert to sorted slice
	result := make([]TimeSeriesData, 0, len(dataMap))
	for _, data := range dataMap {
		result = append(result, *data)
	}

	// Sort by date
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Date < result[i].Date {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// LinkPerformance represents individual link performance
type LinkPerformance struct {
	LinkCode    string  `json:"linkCode"`
	Campaign    string  `json:"campaign"`
	Clicks      int64   `json:"clicks"`
	UniqueClicks int64  `json:"uniqueClicks"`
	Conversions int64   `json:"conversions"`
	Revenue     float64 `json:"revenue"`
	CTR         float64 `json:"ctr"` // Click-through rate
	CVR         float64 `json:"cvr"` // Conversion rate
}

// GetLinkPerformance analyzes performance by link
func (ae *AnalyticsEngine) GetLinkPerformance(affiliateID int64, startDate, endDate time.Time) []LinkPerformance {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	// This would query link stats from tracking manager
	// Simplified implementation
	result := []LinkPerformance{}

	return result
}

// GeneratePerformanceReport generates a comprehensive performance report
func (ae *AnalyticsEngine) GeneratePerformanceReport(affiliateID int64, startDate, endDate time.Time) map[string]interface{} {
	stats, _ := ae.GetDashboardStats(affiliateID)
	funnel := ae.GetConversionFunnel(affiliateID, startDate, endDate)
	traffic := ae.GetTrafficSources(affiliateID, startDate, endDate)
	geo := ae.GetGeographicPerformance(affiliateID, startDate, endDate)
	timeSeries := ae.GetTimeSeriesData(affiliateID, startDate, endDate, "day")

	report := map[string]interface{}{
		"period": map[string]string{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
		},
		"summary":          stats,
		"conversion_funnel": funnel,
		"traffic_sources":  traffic,
		"geographic":       geo,
		"time_series":      timeSeries,
		"generated_at":     time.Now().Format(time.RFC3339),
	}

	return report
}

// ComparePerformance compares two time periods
func (ae *AnalyticsEngine) ComparePerformance(affiliateID int64, period1Start, period1End, period2Start, period2End time.Time) map[string]interface{} {
	funnel1 := ae.GetConversionFunnel(affiliateID, period1Start, period1End)
	funnel2 := ae.GetConversionFunnel(affiliateID, period2Start, period2End)

	comparison := map[string]interface{}{
		"period1": map[string]interface{}{
			"start": period1Start.Format("2006-01-02"),
			"end":   period1End.Format("2006-01-02"),
			"data":  funnel1,
		},
		"period2": map[string]interface{}{
			"start": period2Start.Format("2006-01-02"),
			"end":   period2End.Format("2006-01-02"),
			"data":  funnel2,
		},
		"changes": map[string]interface{}{
			"clicks":      funnel2.TotalClicks - funnel1.TotalClicks,
			"conversions": funnel2.Signups - funnel1.Signups,
			"cvr_change":  funnel2.OverallConversion - funnel1.OverallConversion,
		},
	}

	return comparison
}
