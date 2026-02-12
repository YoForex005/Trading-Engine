package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// SessionType represents the type of trading session
type SessionType string

const (
	SessionRegular     SessionType = "REGULAR"
	SessionPreMarket   SessionType = "PRE_MARKET"
	SessionAfterHours  SessionType = "AFTER_HOURS"
	SessionClosed      SessionType = "CLOSED"
)

// DaySchedule represents trading hours for a specific day
type DaySchedule struct {
	DayOfWeek int       `json:"dayOfWeek"` // 0=Sunday, 6=Saturday
	Sessions  []MarketSession `json:"sessions"`
}

// MarketSession represents a trading session with start and end times
type MarketSession struct {
	Type      SessionType `json:"type"`
	StartTime string      `json:"startTime"` // HH:MM format in UTC
	EndTime   string      `json:"endTime"`   // HH:MM format in UTC
}

// MarketSchedule represents the complete schedule for a symbol
type MarketSchedule struct {
	Symbol       string        `json:"symbol"`
	Timezone     string        `json:"timezone"`     // e.g., "America/New_York", "UTC"
	WeekSchedule []DaySchedule `json:"weekSchedule"` // Index 0-6 for Sunday-Saturday
	Holidays     []time.Time   `json:"holidays"`     // Holiday dates (UTC)
}

// MarketHoliday represents a market holiday
type MarketHoliday struct {
	Date        time.Time `json:"date"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// CurrentSessionInfo provides current market status
type CurrentSessionInfo struct {
	Symbol       string      `json:"symbol"`
	IsOpen       bool        `json:"isOpen"`
	CurrentType  SessionType `json:"currentType"`
	NextOpen     time.Time   `json:"nextOpen"`
	NextClose    time.Time   `json:"nextClose"`
	Timezone     string      `json:"timezone"`
	CheckedAt    time.Time   `json:"checkedAt"`
}

// MarketHoursService manages trading session schedules
type MarketHoursService struct {
	schedules map[string]*MarketSchedule // Symbol -> Schedule
	holidays  []MarketHoliday              // Global holidays
	mu        sync.RWMutex
}

// NewMarketHoursService creates a new market hours service with default schedules
func NewMarketHoursService() *MarketHoursService {
	service := &MarketHoursService{
		schedules: make(map[string]*MarketSchedule),
		holidays:  []MarketHoliday{},
	}

	// Initialize default schedules
	service.initializeDefaultSchedules()

	return service
}

// initializeDefaultSchedules sets up default market hours for different asset classes
func (s *MarketHoursService) initializeDefaultSchedules() {
	// Forex: 24/5 (Sunday 22:00 UTC to Friday 22:00 UTC)
	forexSymbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "NZDUSD", "EURGBP", "EURJPY"}
	for _, symbol := range forexSymbols {
		s.schedules[symbol] = &MarketSchedule{
			Symbol:   symbol,
			Timezone: "UTC",
			WeekSchedule: []DaySchedule{
				// Sunday
				{DayOfWeek: 0, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "22:00", EndTime: "23:59"}}},
				// Monday-Thursday
				{DayOfWeek: 1, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 2, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 3, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 4, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				// Friday
				{DayOfWeek: 5, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "22:00"}}},
				// Saturday - Closed
				{DayOfWeek: 6, Sessions: []MarketSession{{Type: SessionClosed, StartTime: "00:00", EndTime: "23:59"}}},
			},
			Holidays: []time.Time{},
		}
	}

	// Crypto: 24/7
	cryptoSymbols := []string{"BTCUSD", "ETHUSD", "BNBUSD", "XRPUSD", "SOLUSD"}
	for _, symbol := range cryptoSymbols {
		s.schedules[symbol] = &MarketSchedule{
			Symbol:   symbol,
			Timezone: "UTC",
			WeekSchedule: []DaySchedule{
				{DayOfWeek: 0, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 1, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 2, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 3, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 4, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 5, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 6, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
			},
			Holidays: []time.Time{},
		}
	}

	// US Indices: NYSE hours (14:30 UTC - 21:00 UTC, Monday-Friday)
	// Pre-market: 09:00-14:30 UTC, After-hours: 21:00-01:00 UTC
	usIndices := []string{"SPX500USD", "US30USD", "NAS100USD", "US2000USD"}
	for _, symbol := range usIndices {
		s.schedules[symbol] = &MarketSchedule{
			Symbol:   symbol,
			Timezone: "America/New_York",
			WeekSchedule: []DaySchedule{
				// Sunday - Closed
				{DayOfWeek: 0, Sessions: []MarketSession{{Type: SessionClosed, StartTime: "00:00", EndTime: "23:59"}}},
				// Monday-Friday
				{DayOfWeek: 1, Sessions: []MarketSession{
					{Type: SessionPreMarket, StartTime: "09:00", EndTime: "14:30"},
					{Type: SessionRegular, StartTime: "14:30", EndTime: "21:00"},
					{Type: SessionAfterHours, StartTime: "21:00", EndTime: "01:00"},
				}},
				{DayOfWeek: 2, Sessions: []MarketSession{
					{Type: SessionPreMarket, StartTime: "09:00", EndTime: "14:30"},
					{Type: SessionRegular, StartTime: "14:30", EndTime: "21:00"},
					{Type: SessionAfterHours, StartTime: "21:00", EndTime: "01:00"},
				}},
				{DayOfWeek: 3, Sessions: []MarketSession{
					{Type: SessionPreMarket, StartTime: "09:00", EndTime: "14:30"},
					{Type: SessionRegular, StartTime: "14:30", EndTime: "21:00"},
					{Type: SessionAfterHours, StartTime: "21:00", EndTime: "01:00"},
				}},
				{DayOfWeek: 4, Sessions: []MarketSession{
					{Type: SessionPreMarket, StartTime: "09:00", EndTime: "14:30"},
					{Type: SessionRegular, StartTime: "14:30", EndTime: "21:00"},
					{Type: SessionAfterHours, StartTime: "21:00", EndTime: "01:00"},
				}},
				{DayOfWeek: 5, Sessions: []MarketSession{
					{Type: SessionPreMarket, StartTime: "09:00", EndTime: "14:30"},
					{Type: SessionRegular, StartTime: "14:30", EndTime: "21:00"},
					{Type: SessionAfterHours, StartTime: "21:00", EndTime: "01:00"},
				}},
				// Saturday - Closed
				{DayOfWeek: 6, Sessions: []MarketSession{{Type: SessionClosed, StartTime: "00:00", EndTime: "23:59"}}},
			},
			Holidays: []time.Time{},
		}
	}

	// Commodities: Gold/Silver - Sunday 23:00 UTC to Friday 22:00 UTC
	commodities := []string{"XAUUSD", "XAGUSD", "WTICOUSD", "BCOUSD"}
	for _, symbol := range commodities {
		s.schedules[symbol] = &MarketSchedule{
			Symbol:   symbol,
			Timezone: "UTC",
			WeekSchedule: []DaySchedule{
				{DayOfWeek: 0, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "23:00", EndTime: "23:59"}}},
				{DayOfWeek: 1, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 2, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 3, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 4, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "23:59"}}},
				{DayOfWeek: 5, Sessions: []MarketSession{{Type: SessionRegular, StartTime: "00:00", EndTime: "22:00"}}},
				{DayOfWeek: 6, Sessions: []MarketSession{{Type: SessionClosed, StartTime: "00:00", EndTime: "23:59"}}},
			},
			Holidays: []time.Time{},
		}
	}

	log.Printf("[MarketHours] Initialized default schedules for %d symbols (Forex: %d, Crypto: %d, Indices: %d, Commodities: %d)",
		len(s.schedules),
		len(forexSymbols),
		len(cryptoSymbols),
		len(usIndices),
		len(commodities),
	)
}

// IsMarketOpen checks if the market is currently open for a given symbol
func (s *MarketHoursService) IsMarketOpen(symbol string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedule, exists := s.schedules[symbol]
	if !exists {
		// If no schedule exists, assume market is open (default behavior)
		return true
	}

	now := time.Now().UTC()

	// Check if today is a holiday
	for _, holiday := range schedule.Holidays {
		if isSameDay(now, holiday) {
			return false
		}
	}

	// Get current day of week (0=Sunday, 6=Saturday)
	dayOfWeek := int(now.Weekday())

	// Find schedule for current day
	var daySchedule *DaySchedule
	for _, ds := range schedule.WeekSchedule {
		if ds.DayOfWeek == dayOfWeek {
			daySchedule = &ds
			break
		}
	}

	if daySchedule == nil {
		return false
	}

	// Check if current time falls within any session
	currentTime := now.Format("15:04")
	for _, session := range daySchedule.Sessions {
		if session.Type == SessionClosed {
			continue
		}

		// Handle sessions that cross midnight
		if session.EndTime < session.StartTime {
			// Session crosses midnight
			if currentTime >= session.StartTime || currentTime <= session.EndTime {
				return true
			}
		} else {
			// Normal session
			if currentTime >= session.StartTime && currentTime <= session.EndTime {
				return true
			}
		}
	}

	return false
}

// GetCurrentSessionInfo returns detailed current session information for a symbol
func (s *MarketHoursService) GetCurrentSessionInfo(symbol string) (*CurrentSessionInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedule, exists := s.schedules[symbol]
	if !exists {
		return nil, fmt.Errorf("no schedule found for symbol: %s", symbol)
	}

	now := time.Now().UTC()
	info := &CurrentSessionInfo{
		Symbol:    symbol,
		Timezone:  schedule.Timezone,
		CheckedAt: now,
	}

	// Check if market is open
	info.IsOpen = s.isMarketOpenAt(schedule, now)

	// Determine current session type
	info.CurrentType = s.getCurrentSessionType(schedule, now)

	// Calculate next open/close times
	info.NextOpen, info.NextClose = s.getNextOpenClose(schedule, now)

	return info, nil
}

// isMarketOpenAt checks if market is open at a specific time
func (s *MarketHoursService) isMarketOpenAt(schedule *MarketSchedule, t time.Time) bool {
	// Check if it's a holiday
	for _, holiday := range schedule.Holidays {
		if isSameDay(t, holiday) {
			return false
		}
	}

	dayOfWeek := int(t.Weekday())
	currentTime := t.Format("15:04")

	for _, ds := range schedule.WeekSchedule {
		if ds.DayOfWeek == dayOfWeek {
			for _, session := range ds.Sessions {
				if session.Type == SessionClosed {
					continue
				}

				if session.EndTime < session.StartTime {
					if currentTime >= session.StartTime || currentTime <= session.EndTime {
						return true
					}
				} else {
					if currentTime >= session.StartTime && currentTime <= session.EndTime {
						return true
					}
				}
			}
		}
	}

	return false
}

// getCurrentSessionType determines the current session type
func (s *MarketHoursService) getCurrentSessionType(schedule *MarketSchedule, t time.Time) SessionType {
	dayOfWeek := int(t.Weekday())
	currentTime := t.Format("15:04")

	for _, ds := range schedule.WeekSchedule {
		if ds.DayOfWeek == dayOfWeek {
			for _, session := range ds.Sessions {
				if session.EndTime < session.StartTime {
					if currentTime >= session.StartTime || currentTime <= session.EndTime {
						return session.Type
					}
				} else {
					if currentTime >= session.StartTime && currentTime <= session.EndTime {
						return session.Type
					}
				}
			}
		}
	}

	return SessionClosed
}

// getNextOpenClose calculates next market open and close times
func (s *MarketHoursService) getNextOpenClose(schedule *MarketSchedule, now time.Time) (time.Time, time.Time) {
	var nextOpen, nextClose time.Time

	// Start searching from current time
	searchTime := now
	maxDays := 14 // Search up to 2 weeks ahead

	for day := 0; day < maxDays; day++ {
		checkTime := searchTime.AddDate(0, 0, day)
		dayOfWeek := int(checkTime.Weekday())

		for _, ds := range schedule.WeekSchedule {
			if ds.DayOfWeek == dayOfWeek {
				for _, session := range ds.Sessions {
					if session.Type == SessionClosed {
						continue
					}

					// Parse session times for this day
					sessionStart := parseTimeOnDate(checkTime, session.StartTime)
					sessionEnd := parseTimeOnDate(checkTime, session.EndTime)

					// If session crosses midnight, adjust end time
					if session.EndTime < session.StartTime {
						sessionEnd = sessionEnd.AddDate(0, 0, 1)
					}

					// Find next open time
					if nextOpen.IsZero() && sessionStart.After(now) {
						nextOpen = sessionStart
					}

					// Find next close time
					if nextClose.IsZero() && sessionEnd.After(now) {
						nextClose = sessionEnd
					}

					if !nextOpen.IsZero() && !nextClose.IsZero() {
						return nextOpen, nextClose
					}
				}
			}
		}
	}

	return nextOpen, nextClose
}

// GetAllSchedules returns all symbol schedules
func (s *MarketHoursService) GetAllSchedules() map[string]*MarketSchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy to avoid race conditions
	schedules := make(map[string]*MarketSchedule, len(s.schedules))
	for k, v := range s.schedules {
		schedules[k] = v
	}

	return schedules
}

// UpdateSchedule updates or creates a schedule for a symbol
func (s *MarketHoursService) UpdateSchedule(schedule *MarketSchedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if schedule.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	s.schedules[schedule.Symbol] = schedule

	log.Printf("[MarketHours] Schedule updated for symbol: %s", schedule.Symbol)

	return nil
}

// AddHoliday adds a market holiday
func (s *MarketHoursService) AddHoliday(holiday MarketHoliday) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.holidays = append(s.holidays, holiday)

	log.Printf("[MarketHours] Holiday added: %s on %s", holiday.Name, holiday.Date.Format("2006-01-02"))

	return nil
}

// Helper functions

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func parseTimeOnDate(date time.Time, timeStr string) time.Time {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return date
	}

	var hour, minute int
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)

	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, time.UTC)
}

// ============================================
// HTTP HANDLERS
// ============================================

// MarketHoursHandler handles market hours HTTP requests
type MarketHoursHandler struct {
	service     *MarketHoursService
	authService *auth.Service
}

// NewMarketHoursHandler creates a new market hours handler
func NewMarketHoursHandler(service *MarketHoursService, authService *auth.Service) *MarketHoursHandler {
	return &MarketHoursHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetSymbolHours returns current session info for a specific symbol
// GET /api/market-hours/:symbol
func (h *MarketHoursHandler) HandleGetSymbolHours(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid symbol", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(pathParts[3])

	// Get session info
	info, err := h.service.GetCurrentSessionInfo(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// HandleGetAllSchedules returns all symbol schedules
// GET /api/market-hours
func (h *MarketHoursHandler) HandleGetAllSchedules(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all schedules
	schedules := h.service.GetAllSchedules()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schedules": schedules,
		"count":     len(schedules),
	})
}

// HandleUpdateSchedule updates market hours for a symbol (admin only)
// PUT /admin/market-hours/:symbol
func (h *MarketHoursHandler) HandleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid symbol", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(pathParts[3])

	// Parse request body
	var schedule MarketSchedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set symbol from URL
	schedule.Symbol = symbol

	// Update schedule
	if err := h.service.UpdateSchedule(&schedule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Schedule updated successfully",
		"symbol":   symbol,
		"schedule": schedule,
	})
}

// HandleAddHoliday adds a market holiday (admin only)
// POST /admin/market-hours/holidays
func (h *MarketHoursHandler) HandleAddHoliday(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Date        string `json:"date"` // YYYY-MM-DD format
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.Date == "" {
		http.Error(w, "date is required (YYYY-MM-DD format)", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	// Parse date
	holidayDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		http.Error(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Create holiday
	holiday := MarketHoliday{
		Date:        holidayDate,
		Name:        req.Name,
		Description: req.Description,
	}

	// Add holiday
	if err := h.service.AddHoliday(holiday); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Holiday added successfully",
		"holiday": holiday,
	})
}

// validateAdminAuth validates JWT token and checks admin role
func (h *MarketHoursHandler) validateAdminAuth(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]
	claims, err := auth.ValidateTokenWithDefault(tokenString)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
