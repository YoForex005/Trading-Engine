package services

import (
	"fmt"
	"time"

	"github.com/epic1st/rtx/backend/compliance/models"
	"github.com/epic1st/rtx/backend/compliance/repository"
	"github.com/google/uuid"
)

// BestExecutionService tracks execution quality metrics
type BestExecutionService struct {
	repo *repository.ComplianceRepository
}

func NewBestExecutionService(repo *repository.ComplianceRepository) *BestExecutionService {
	return &BestExecutionService{
		repo: repo,
	}
}

// TrackExecution records execution metrics for best execution reporting
func (s *BestExecutionService) TrackExecution(
	orderID, symbol, venue, lpName string,
	requestedPrice, executedPrice, quantity float64,
	latency int64, // milliseconds
	fillType string, // PASSIVE, AGGRESSIVE, DIRECTED
) error {

	metrics := map[string]interface{}{
		"order_id":        orderID,
		"symbol":          symbol,
		"venue":           venue,
		"lp_name":         lpName,
		"requested_price": requestedPrice,
		"executed_price":  executedPrice,
		"quantity":        quantity,
		"latency_ms":      latency,
		"fill_type":       fillType,
		"timestamp":       time.Now(),
	}

	// Calculate price improvement
	priceImprovement := s.calculatePriceImprovement(requestedPrice, executedPrice, fillType)
	metrics["price_improvement"] = priceImprovement

	return s.repo.SaveExecutionMetrics(metrics)
}

// GenerateRTS27Report generates RTS 27 execution quality report
// Required annually in EU/UK
func (s *BestExecutionService) GenerateRTS27Report(year int, quarter string, instrumentClass string) (*models.BestExecutionReport, error) {
	startDate, endDate := s.getQuarterDates(year, quarter)

	metrics, err := s.repo.GetExecutionMetrics(instrumentClass, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate aggregate metrics
	totalExecutions := len(metrics)
	if totalExecutions == 0 {
		return nil, fmt.Errorf("no executions found for period")
	}

	var totalPriceImprovement, totalLatency, totalSlippage float64
	rejectedOrders := 0

	for _, m := range metrics {
		totalPriceImprovement += m["price_improvement"].(float64)
		totalLatency += m["latency_ms"].(float64)
		if m["slippage"] != nil {
			totalSlippage += m["slippage"].(float64)
		}
		if m["rejected"].(bool) {
			rejectedOrders++
		}
	}

	report := &models.BestExecutionReport{
		ID:                   uuid.New().String(),
		ReportType:           "RTS_27",
		Period:               fmt.Sprintf("Q%s_%d", quarter, year),
		InstrumentClass:      instrumentClass,
		PriceImprovement:     totalPriceImprovement / float64(totalExecutions),
		FillRate:             float64(totalExecutions-rejectedOrders) / float64(totalExecutions),
		AverageExecutionTime: totalLatency / float64(totalExecutions),
		SlippageRate:         totalSlippage / float64(totalExecutions),
		RejectionRate:        float64(rejectedOrders) / float64(totalExecutions),
		CreatedAt:            time.Now(),
	}

	if err := s.repo.SaveBestExecutionReport(report); err != nil {
		return nil, fmt.Errorf("failed to save RTS 27 report: %w", err)
	}

	return report, nil
}

// GenerateRTS28Report generates RTS 28 top execution venues report
// Required annually in EU/UK
func (s *BestExecutionService) GenerateRTS28Report(year int, instrumentClass string) (*models.BestExecutionReport, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	venueStatsRaw, err := s.repo.GetVenueStatistics(instrumentClass, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Convert []interface{} to []VenueStats
	venueStats := make([]VenueStats, 0, len(venueStatsRaw))
	for _, v := range venueStatsRaw {
		if vs, ok := v.(VenueStats); ok {
			venueStats = append(venueStats, vs)
		}
	}

	// Get top 5 venues by execution volume
	topVenues := s.getTopVenues(venueStats, 5)

	for _, venue := range topVenues {
		report := &models.BestExecutionReport{
			ID:              uuid.New().String(),
			ReportType:      "RTS_28",
			Period:          fmt.Sprintf("%d", year),
			InstrumentClass: instrumentClass,
			Venue:           venue.Name,
			LPName:          venue.LPName,
			ExecutionVolume: venue.Volume,
			NumberOfOrders:  venue.OrderCount,
			PassiveOrders:   venue.PassiveCount,
			AggressiveOrders: venue.AggressiveCount,
			DirectedOrders:  venue.DirectedCount,
			CreatedAt:       time.Now(),
		}

		if err := s.repo.SaveBestExecutionReport(report); err != nil {
			return nil, fmt.Errorf("failed to save RTS 28 report: %w", err)
		}
	}

	return nil, nil
}

// PublishReport publishes best execution report to public website
// Required by MiFID II
func (s *BestExecutionService) PublishReport(reportID string) error {
	report, err := s.repo.GetBestExecutionReport(reportID)
	if err != nil {
		return err
	}

	// In production, this would upload to public website
	// For now, we mark as published
	now := time.Now()
	report.PublishedAt = &now

	return s.repo.UpdateBestExecutionReportPublished(reportID, now)
}

// CalculateExecutionQuality calculates real-time execution quality score
func (s *BestExecutionService) CalculateExecutionQuality(lpName string, symbol string, hours int) (float64, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	metrics, err := s.repo.GetExecutionMetricsByLP(lpName, symbol, startTime, time.Now())
	if err != nil {
		return 0, err
	}

	if len(metrics) == 0 {
		return 0, fmt.Errorf("no metrics found")
	}

	// Quality score components (0-100 scale)
	var priceScore, speedScore, fillScore, reliabilityScore float64

	totalExecutions := float64(len(metrics))
	var totalImprovement, totalLatency float64
	successfulFills := 0

	for _, m := range metrics {
		// Price improvement score
		totalImprovement += m["price_improvement"].(float64)

		// Speed score
		latency := m["latency_ms"].(float64)
		totalLatency += latency

		// Fill score
		if !m["rejected"].(bool) {
			successfulFills++
		}
	}

	// Calculate component scores
	priceScore = (totalImprovement / totalExecutions) * 100
	if priceScore > 100 {
		priceScore = 100
	}

	avgLatency := totalLatency / totalExecutions
	speedScore = (1000 - avgLatency) / 10 // Latency in ms, lower is better
	if speedScore > 100 {
		speedScore = 100
	}
	if speedScore < 0 {
		speedScore = 0
	}

	fillScore = (float64(successfulFills) / totalExecutions) * 100
	reliabilityScore = fillScore // Simplified

	// Weighted average
	qualityScore := (priceScore*0.4 + speedScore*0.3 + fillScore*0.2 + reliabilityScore*0.1)

	return qualityScore, nil
}

// Helper methods

func (s *BestExecutionService) calculatePriceImprovement(requested, executed float64, fillType string) float64 {
	if fillType == "PASSIVE" {
		// Passive orders should get better prices
		return (requested - executed) / requested * 100
	}
	return 0
}

func (s *BestExecutionService) getQuarterDates(year int, quarter string) (time.Time, time.Time) {
	var startMonth, endMonth time.Month

	switch quarter {
	case "1":
		startMonth = time.January
		endMonth = time.March
	case "2":
		startMonth = time.April
		endMonth = time.June
	case "3":
		startMonth = time.July
		endMonth = time.September
	case "4":
		startMonth = time.October
		endMonth = time.December
	}

	startDate := time.Date(year, startMonth, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, endMonth+1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	return startDate, endDate
}

type VenueStats struct {
	Name            string
	LPName          string
	Volume          float64
	OrderCount      int
	PassiveCount    int
	AggressiveCount int
	DirectedCount   int
}

func (s *BestExecutionService) getTopVenues(stats []VenueStats, top int) []VenueStats {
	// Sort by volume and return top N
	// Simplified implementation
	if len(stats) <= top {
		return stats
	}
	return stats[:top]
}
