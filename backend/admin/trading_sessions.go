package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

type TradingSession struct {
	ID              int64          `json:"id"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	OpenTime        string         `json:"openTime"`        // HH:MM format
	CloseTime       string         `json:"closeTime"`       // HH:MM format
	Timezone        string         `json:"timezone"`        // e.g., "UTC", "America/New_York"
	BreakPeriods    []BreakPeriod  `json:"breakPeriods"`
	WeekendClosed   bool           `json:"weekendClosed"`
	ClosedDays      []string       `json:"closedDays"`      // ["Saturday", "Sunday"]
	HolidaySchedule string         `json:"holidaySchedule"` // "US", "UK", "Japan", "None"
	SymbolsCount    int            `json:"symbolsCount"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

type BreakPeriod struct {
	StartTime string `json:"startTime"` // HH:MM format
	EndTime   string `json:"endTime"`   // HH:MM format
	Reason    string `json:"reason"`    // e.g., "Server Maintenance", "Lunch Break"
}

type SymbolSession struct {
	Symbol         string        `json:"symbol"`
	SessionID      int64         `json:"sessionId"`
	SessionName    string        `json:"sessionName"`
	OpenTime       string        `json:"openTime"`
	CloseTime      string        `json:"closeTime"`
	Timezone       string        `json:"timezone"`
	BreakPeriods   []BreakPeriod `json:"breakPeriods"`
	WeekendClosed  bool          `json:"weekendClosed"`
	ClosedDays     []string      `json:"closedDays"`
	CustomOverride bool          `json:"customOverride"` // If symbol has custom hours different from template
}

type SessionTemplate struct {
	ID            int64         `json:"id"`
	Name          string        `json:"name"`
	Type          string        `json:"type"` // "24/5 Forex", "24/7 Crypto", "NYSE", "LSE", "Tokyo", "Custom"
	Description   string        `json:"description"`
	OpenTime      string        `json:"openTime"`
	CloseTime     string        `json:"closeTime"`
	Timezone      string        `json:"timezone"`
	BreakPeriods  []BreakPeriod `json:"breakPeriods"`
	WeekendClosed bool          `json:"weekendClosed"`
	ClosedDays    []string      `json:"closedDays"`
	IsDefault     bool          `json:"isDefault"`
	CreatedAt     time.Time     `json:"createdAt"`
}

type Holiday struct {
	ID              int64     `json:"id"`
	Date            string    `json:"date"` // YYYY-MM-DD format
	Name            string    `json:"name"`
	AffectedSessions []string  `json:"affectedSessions"` // Session names
	ClosureType     string    `json:"closureType"`      // "full" or "partial"
	PartialHours    string    `json:"partialHours,omitempty"` // e.g., "09:00-13:00" for half day
	Region          string    `json:"region"`           // "US", "UK", "Japan", "Global"
	CreatedAt       time.Time `json:"createdAt"`
}

type MarketStatus struct {
	SessionID         int64     `json:"sessionId"`
	SessionName       string    `json:"sessionName"`
	Status            string    `json:"status"` // "open", "closed", "pre-market", "after-hours", "break"
	CurrentTime       time.Time `json:"currentTime"`
	NextStateChange   string    `json:"nextStateChange"`   // Timestamp of next state change
	TimeToNextChange  string    `json:"timeToNextChange"`  // Human readable: "2h 15m"
	IsHoliday         bool      `json:"isHoliday"`
	HolidayName       string    `json:"holidayName,omitempty"`
	CurrentBreak      string    `json:"currentBreak,omitempty"` // If in break period
}

// ============================================
// Service
// ============================================

type TradingSessionService struct {
	sessions        map[int64]*TradingSession
	templates       map[int64]*SessionTemplate
	symbolSessions  map[string]*SymbolSession // symbol -> session mapping
	holidays        map[int64]*Holiday
	nextSessionID   int64
	nextTemplateID  int64
	nextHolidayID   int64
	mu              sync.RWMutex
}

func NewTradingSessionService() *TradingSessionService {
	service := &TradingSessionService{
		sessions:       make(map[int64]*TradingSession),
		templates:      make(map[int64]*SessionTemplate),
		symbolSessions: make(map[string]*SymbolSession),
		holidays:       make(map[int64]*Holiday),
		nextSessionID:  1,
		nextTemplateID: 1,
		nextHolidayID:  1,
	}
	service.initializeMockData()
	return service
}

func (s *TradingSessionService) initializeMockData() {
	now := time.Now()

	// ============================================
	// Initialize 6 Session Templates
	// ============================================
	templates := []SessionTemplate{
		{
			ID:            1,
			Name:          "24/5 Forex",
			Type:          "24/5 Forex",
			Description:   "24-hour trading, 5 days a week (Sunday evening to Friday evening)",
			OpenTime:      "22:00",
			CloseTime:     "22:00",
			Timezone:      "UTC",
			BreakPeriods:  []BreakPeriod{{StartTime: "22:00", EndTime: "22:05", Reason: "Daily Server Maintenance"}},
			WeekendClosed: true,
			ClosedDays:    []string{"Saturday"},
			IsDefault:     true,
			CreatedAt:     now.AddDate(0, -6, 0),
		},
		{
			ID:            2,
			Name:          "24/7 Crypto",
			Type:          "24/7 Crypto",
			Description:   "24-hour trading, 7 days a week",
			OpenTime:      "00:00",
			CloseTime:     "23:59",
			Timezone:      "UTC",
			BreakPeriods:  []BreakPeriod{},
			WeekendClosed: false,
			ClosedDays:    []string{},
			IsDefault:     true,
			CreatedAt:     now.AddDate(0, -6, 0),
		},
		{
			ID:            3,
			Name:          "NYSE Hours",
			Type:          "NYSE",
			Description:   "New York Stock Exchange (9:30 AM - 4:00 PM EST)",
			OpenTime:      "14:30",
			CloseTime:     "21:00",
			Timezone:      "UTC",
			BreakPeriods:  []BreakPeriod{},
			WeekendClosed: true,
			ClosedDays:    []string{"Saturday", "Sunday"},
			IsDefault:     true,
			CreatedAt:     now.AddDate(0, -6, 0),
		},
		{
			ID:            4,
			Name:          "LSE Hours",
			Type:          "LSE",
			Description:   "London Stock Exchange (8:00 AM - 4:30 PM GMT)",
			OpenTime:      "08:00",
			CloseTime:     "16:30",
			Timezone:      "UTC",
			BreakPeriods:  []BreakPeriod{},
			WeekendClosed: true,
			ClosedDays:    []string{"Saturday", "Sunday"},
			IsDefault:     true,
			CreatedAt:     now.AddDate(0, -6, 0),
		},
		{
			ID:            5,
			Name:          "Tokyo Hours",
			Type:          "Tokyo",
			Description:   "Tokyo Stock Exchange (9:00 AM - 3:00 PM JST)",
			OpenTime:      "00:00",
			CloseTime:     "06:00",
			Timezone:      "UTC",
			BreakPeriods:  []BreakPeriod{{StartTime: "02:30", EndTime: "03:30", Reason: "Lunch Break"}},
			WeekendClosed: true,
			ClosedDays:    []string{"Saturday", "Sunday"},
			IsDefault:     true,
			CreatedAt:     now.AddDate(0, -6, 0),
		},
		{
			ID:            6,
			Name:          "Custom Template",
			Type:          "Custom",
			Description:   "User-defined custom trading hours",
			OpenTime:      "08:00",
			CloseTime:     "17:00",
			Timezone:      "UTC",
			BreakPeriods:  []BreakPeriod{},
			WeekendClosed: true,
			ClosedDays:    []string{"Saturday", "Sunday"},
			IsDefault:     false,
			CreatedAt:     now.AddDate(0, -1, 0),
		},
	}

	for i := range templates {
		s.templates[templates[i].ID] = &templates[i]
	}
	s.nextTemplateID = 7

	// ============================================
	// Initialize Trading Sessions from Templates
	// ============================================
	sessions := []TradingSession{
		{
			ID:              1,
			Name:            "Forex Major Pairs",
			Description:     "EUR/USD, GBP/USD, USD/JPY, etc.",
			OpenTime:        "22:00",
			CloseTime:       "22:00",
			Timezone:        "UTC",
			BreakPeriods:    []BreakPeriod{{StartTime: "22:00", EndTime: "22:05", Reason: "Daily Server Maintenance"}},
			WeekendClosed:   true,
			ClosedDays:      []string{"Saturday"},
			HolidaySchedule: "US",
			SymbolsCount:    50,
			CreatedAt:       now.AddDate(0, -6, 0),
			UpdatedAt:       now.AddDate(0, 0, -2),
		},
		{
			ID:              2,
			Name:            "Cryptocurrency",
			Description:     "BTC, ETH, XRP, SOL, BNB",
			OpenTime:        "00:00",
			CloseTime:       "23:59",
			Timezone:        "UTC",
			BreakPeriods:    []BreakPeriod{},
			WeekendClosed:   false,
			ClosedDays:      []string{},
			HolidaySchedule: "None",
			SymbolsCount:    30,
			CreatedAt:       now.AddDate(0, -6, 0),
			UpdatedAt:       now.AddDate(0, 0, -1),
		},
		{
			ID:              3,
			Name:            "US Stocks",
			Description:     "NYSE and NASDAQ listed stocks",
			OpenTime:        "14:30",
			CloseTime:       "21:00",
			Timezone:        "UTC",
			BreakPeriods:    []BreakPeriod{},
			WeekendClosed:   true,
			ClosedDays:      []string{"Saturday", "Sunday"},
			HolidaySchedule: "US",
			SymbolsCount:    40,
			CreatedAt:       now.AddDate(0, -6, 0),
			UpdatedAt:       now.AddDate(0, 0, -3),
		},
		{
			ID:              4,
			Name:            "UK Stocks",
			Description:     "LSE listed stocks",
			OpenTime:        "08:00",
			CloseTime:       "16:30",
			Timezone:        "UTC",
			BreakPeriods:    []BreakPeriod{},
			WeekendClosed:   true,
			ClosedDays:      []string{"Saturday", "Sunday"},
			HolidaySchedule: "UK",
			SymbolsCount:    20,
			CreatedAt:       now.AddDate(0, -6, 0),
			UpdatedAt:       now.AddDate(0, 0, -4),
		},
		{
			ID:              5,
			Name:            "Japan Stocks",
			Description:     "Tokyo Stock Exchange",
			OpenTime:        "00:00",
			CloseTime:       "06:00",
			Timezone:        "UTC",
			BreakPeriods:    []BreakPeriod{{StartTime: "02:30", EndTime: "03:30", Reason: "Lunch Break"}},
			WeekendClosed:   true,
			ClosedDays:      []string{"Saturday", "Sunday"},
			HolidaySchedule: "Japan",
			SymbolsCount:    10,
			CreatedAt:       now.AddDate(0, -6, 0),
			UpdatedAt:       now.AddDate(0, 0, -5),
		},
	}

	for i := range sessions {
		s.sessions[sessions[i].ID] = &sessions[i]
	}
	s.nextSessionID = 6

	// ============================================
	// Map 150 Symbols to Sessions
	// ============================================

	// Forex pairs (50 symbols) - 24/5 Forex session
	forexPairs := []string{
		"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "NZDUSD", "USDCAD",
		"EURGBP", "EURJPY", "EURAUD", "EURCHF", "GBPJPY", "GBPAUD", "GBPCHF",
		"AUDJPY", "AUDNZD", "AUDCAD", "AUDCHF", "NZDJPY", "NZDCAD", "NZDCHF",
		"CADJPY", "CADCHF", "CHFJPY", "EURCZK", "EURHUF", "EURPLN", "EURNOK",
		"EURSEK", "EURDKK", "EURTRY", "GBPNZD", "GBPPLN", "USDMXN", "USDNOK",
		"USDSEK", "USDDKK", "USDPLN", "USDTRY", "USDZAR", "USDHUF", "USDCZK",
		"EURZAR", "GBPZAR", "AUDSGD", "NZDSGD", "SGDJPY", "USDHKD", "USDSGD",
		"USDCNH",
	}

	for _, symbol := range forexPairs {
		s.symbolSessions[symbol] = &SymbolSession{
			Symbol:         symbol,
			SessionID:      1,
			SessionName:    "Forex Major Pairs",
			OpenTime:       "22:00",
			CloseTime:      "22:00",
			Timezone:       "UTC",
			BreakPeriods:   []BreakPeriod{{StartTime: "22:00", EndTime: "22:05", Reason: "Daily Server Maintenance"}},
			WeekendClosed:  true,
			ClosedDays:     []string{"Saturday"},
			CustomOverride: false,
		}
	}

	// Crypto (30 symbols) - 24/7 Crypto session
	cryptoPairs := []string{
		"BTCUSD", "ETHUSD", "XRPUSD", "SOLUSD", "BNBUSD", "ADAUSD", "DOGUSD",
		"DOTUSD", "MATICUSD", "LINKUSD", "AVAXUSD", "UNIUSD", "ATOMUSD", "LTCUSD",
		"ETCUSD", "XMRUSD", "BCHUSD", "ALGOUSD", "FILUSD", "VETUSDT", "ICPUSD",
		"FTMUSD", "SANDUSD", "MANAUSD", "GRTUSD", "ENJUSD", "CHZUSD", "THEUSD",
		"AXSUSD", "SHIBUSDT",
	}

	for _, symbol := range cryptoPairs {
		s.symbolSessions[symbol] = &SymbolSession{
			Symbol:         symbol,
			SessionID:      2,
			SessionName:    "Cryptocurrency",
			OpenTime:       "00:00",
			CloseTime:      "23:59",
			Timezone:       "UTC",
			BreakPeriods:   []BreakPeriod{},
			WeekendClosed:  false,
			ClosedDays:     []string{},
			CustomOverride: false,
		}
	}

	// US Stocks (40 symbols) - NYSE session
	usStocks := []string{
		"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "NVDA", "META", "BRK.B",
		"JNJ", "V", "WMT", "JPM", "MA", "PG", "UNH", "DIS", "HD", "BAC",
		"ADBE", "CRM", "NFLX", "PYPL", "INTC", "CMCSA", "VZ", "T", "PFE",
		"KO", "PEP", "ABT", "MRK", "CSCO", "NKE", "TMO", "ABBV", "ACN",
		"AVGO", "TXN", "DHR", "QCOM",
	}

	for _, symbol := range usStocks {
		s.symbolSessions[symbol] = &SymbolSession{
			Symbol:         symbol,
			SessionID:      3,
			SessionName:    "US Stocks",
			OpenTime:       "14:30",
			CloseTime:      "21:00",
			Timezone:       "UTC",
			BreakPeriods:   []BreakPeriod{},
			WeekendClosed:  true,
			ClosedDays:     []string{"Saturday", "Sunday"},
			CustomOverride: false,
		}
	}

	// UK Stocks (20 symbols) - LSE session
	ukStocks := []string{
		"BP", "HSBA", "RIO", "GSK", "AZN", "SHEL", "DGE", "ULVR", "BATS",
		"REL", "NG", "LSEG", "AAL", "BARC", "LLOY", "PRU", "VOD", "BT.A",
		"IMB", "CCH",
	}

	for _, symbol := range ukStocks {
		s.symbolSessions[symbol] = &SymbolSession{
			Symbol:         symbol,
			SessionID:      4,
			SessionName:    "UK Stocks",
			OpenTime:       "08:00",
			CloseTime:      "16:30",
			Timezone:       "UTC",
			BreakPeriods:   []BreakPeriod{},
			WeekendClosed:  true,
			ClosedDays:     []string{"Saturday", "Sunday"},
			CustomOverride: false,
		}
	}

	// Japan Stocks (10 symbols) - Tokyo session
	japanStocks := []string{
		"7203.T", "9984.T", "6758.T", "8306.T", "6861.T", "9433.T",
		"7267.T", "8035.T", "6098.T", "4063.T",
	}

	for _, symbol := range japanStocks {
		s.symbolSessions[symbol] = &SymbolSession{
			Symbol:         symbol,
			SessionID:      5,
			SessionName:    "Japan Stocks",
			OpenTime:       "00:00",
			CloseTime:      "06:00",
			Timezone:       "UTC",
			BreakPeriods:   []BreakPeriod{{StartTime: "02:30", EndTime: "03:30", Reason: "Lunch Break"}},
			WeekendClosed:  true,
			ClosedDays:     []string{"Saturday", "Sunday"},
			CustomOverride: false,
		}
	}

	// ============================================
	// Initialize 30 Holidays for 2026
	// ============================================
	holidays := []Holiday{
		{ID: 1, Date: "2026-01-01", Name: "New Year's Day", AffectedSessions: []string{"US Stocks", "UK Stocks", "Japan Stocks"}, ClosureType: "full", Region: "Global", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 2, Date: "2026-01-19", Name: "Martin Luther King Jr. Day", AffectedSessions: []string{"US Stocks"}, ClosureType: "full", Region: "US", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 3, Date: "2026-02-16", Name: "Presidents' Day", AffectedSessions: []string{"US Stocks"}, ClosureType: "full", Region: "US", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 4, Date: "2026-04-03", Name: "Good Friday", AffectedSessions: []string{"US Stocks", "UK Stocks"}, ClosureType: "full", Region: "US", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 5, Date: "2026-04-06", Name: "Easter Monday", AffectedSessions: []string{"UK Stocks"}, ClosureType: "full", Region: "UK", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 6, Date: "2026-05-04", Name: "Early May Bank Holiday", AffectedSessions: []string{"UK Stocks"}, ClosureType: "full", Region: "UK", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 7, Date: "2026-05-25", Name: "Memorial Day", AffectedSessions: []string{"US Stocks"}, ClosureType: "full", Region: "US", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 8, Date: "2026-05-25", Name: "Spring Bank Holiday", AffectedSessions: []string{"UK Stocks"}, ClosureType: "full", Region: "UK", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 9, Date: "2026-07-03", Name: "Independence Day (observed)", AffectedSessions: []string{"US Stocks"}, ClosureType: "full", Region: "US", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 10, Date: "2026-08-31", Name: "Summer Bank Holiday", AffectedSessions: []string{"UK Stocks"}, ClosureType: "full", Region: "UK", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 11, Date: "2026-09-07", Name: "Labor Day", AffectedSessions: []string{"US Stocks"}, ClosureType: "full", Region: "US", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 12, Date: "2026-11-26", Name: "Thanksgiving Day", AffectedSessions: []string{"US Stocks"}, ClosureType: "full", Region: "US", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 13, Date: "2026-11-27", Name: "Day after Thanksgiving", AffectedSessions: []string{"US Stocks"}, ClosureType: "partial", PartialHours: "14:30-18:00", Region: "US", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 14, Date: "2026-12-24", Name: "Christmas Eve", AffectedSessions: []string{"US Stocks"}, ClosureType: "partial", PartialHours: "14:30-18:00", Region: "US", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 15, Date: "2026-12-25", Name: "Christmas Day", AffectedSessions: []string{"US Stocks", "UK Stocks", "Japan Stocks"}, ClosureType: "full", Region: "Global", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 16, Date: "2026-12-28", Name: "Boxing Day (observed)", AffectedSessions: []string{"UK Stocks"}, ClosureType: "full", Region: "UK", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 17, Date: "2026-01-12", Name: "Coming of Age Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 18, Date: "2026-02-11", Name: "National Foundation Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 19, Date: "2026-03-20", Name: "Vernal Equinox Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 20, Date: "2026-04-29", Name: "Showa Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 21, Date: "2026-05-03", Name: "Constitution Memorial Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 22, Date: "2026-05-04", Name: "Greenery Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 23, Date: "2026-05-05", Name: "Children's Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 24, Date: "2026-07-20", Name: "Marine Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 25, Date: "2026-08-11", Name: "Mountain Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 26, Date: "2026-09-21", Name: "Respect for the Aged Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 27, Date: "2026-09-22", Name: "Autumnal Equinox Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 28, Date: "2026-10-12", Name: "Sports Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 29, Date: "2026-11-03", Name: "Culture Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
		{ID: 30, Date: "2026-11-23", Name: "Labor Thanksgiving Day", AffectedSessions: []string{"Japan Stocks"}, ClosureType: "full", Region: "Japan", CreatedAt: now.AddDate(0, -6, 0)},
	}

	for i := range holidays {
		s.holidays[holidays[i].ID] = &holidays[i]
	}
	s.nextHolidayID = 31

	log.Printf("[TradingSessionService] Initialized with %d sessions, %d templates, %d symbols, %d holidays",
		len(s.sessions), len(s.templates), len(s.symbolSessions), len(s.holidays))
}

// ============================================
// Service Methods
// ============================================

func (s *TradingSessionService) GetAllSessions() []*TradingSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*TradingSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *TradingSessionService) GetSymbolSession(symbol string) *SymbolSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.symbolSessions[symbol]
}

func (s *TradingSessionService) UpdateSymbolSession(symbol string, openTime, closeTime, timezone string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.symbolSessions[symbol]
	if !exists {
		return nil // Symbol not found
	}

	session.OpenTime = openTime
	session.CloseTime = closeTime
	session.Timezone = timezone
	session.CustomOverride = true

	return nil
}

func (s *TradingSessionService) GetTemplates() []*SessionTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]*SessionTemplate, 0, len(s.templates))
	for _, template := range s.templates {
		templates = append(templates, template)
	}
	return templates
}

func (s *TradingSessionService) CreateTemplate(name, templateType, description, openTime, closeTime, timezone string, weekendClosed bool, closedDays []string) *SessionTemplate {
	s.mu.Lock()
	defer s.mu.Unlock()

	template := &SessionTemplate{
		ID:            s.nextTemplateID,
		Name:          name,
		Type:          templateType,
		Description:   description,
		OpenTime:      openTime,
		CloseTime:     closeTime,
		Timezone:      timezone,
		BreakPeriods:  []BreakPeriod{},
		WeekendClosed: weekendClosed,
		ClosedDays:    closedDays,
		IsDefault:     false,
		CreatedAt:     time.Now(),
	}

	s.templates[s.nextTemplateID] = template
	s.nextTemplateID++

	return template
}

func (s *TradingSessionService) GetHolidays() []*Holiday {
	s.mu.RLock()
	defer s.mu.RUnlock()

	holidays := make([]*Holiday, 0, len(s.holidays))
	for _, holiday := range s.holidays {
		holidays = append(holidays, holiday)
	}
	return holidays
}

func (s *TradingSessionService) CreateHoliday(date, name string, affectedSessions []string, closureType, partialHours, region string) *Holiday {
	s.mu.Lock()
	defer s.mu.Unlock()

	holiday := &Holiday{
		ID:               s.nextHolidayID,
		Date:             date,
		Name:             name,
		AffectedSessions: affectedSessions,
		ClosureType:      closureType,
		PartialHours:     partialHours,
		Region:           region,
		CreatedAt:        time.Now(),
	}

	s.holidays[s.nextHolidayID] = holiday
	s.nextHolidayID++

	return holiday
}

func (s *TradingSessionService) GetMarketStatus() []MarketStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make([]MarketStatus, 0, len(s.sessions))
	now := time.Now().UTC()
	todayStr := now.Format("2006-01-02")

	for _, session := range s.sessions {
		status := MarketStatus{
			SessionID:   session.ID,
			SessionName: session.Name,
			CurrentTime: now,
		}

		// Check if today is a holiday
		isHoliday := false
		var holidayName string
		for _, holiday := range s.holidays {
			if holiday.Date == todayStr {
				for _, affectedSession := range holiday.AffectedSessions {
					if affectedSession == session.Name {
						isHoliday = true
						holidayName = holiday.Name
						break
					}
				}
			}
		}

		status.IsHoliday = isHoliday
		status.HolidayName = holidayName

		if isHoliday {
			status.Status = "closed"
			status.NextStateChange = now.AddDate(0, 0, 1).Format("2006-01-02 00:00:00")
			status.TimeToNextChange = "until tomorrow"
		} else {
			// Simple status calculation (simplified for demo)
			currentHour := now.Hour()
			openHour, _ := strconv.Atoi(strings.Split(session.OpenTime, ":")[0])
			closeHour, _ := strconv.Atoi(strings.Split(session.CloseTime, ":")[0])

			if currentHour >= openHour && currentHour < closeHour {
				status.Status = "open"
				closeTime := now.Add(time.Duration(closeHour-currentHour) * time.Hour)
				status.NextStateChange = closeTime.Format("2006-01-02 15:04:05")
				status.TimeToNextChange = formatDuration(closeTime.Sub(now))
			} else if currentHour < openHour {
				status.Status = "pre-market"
				openTime := now.Add(time.Duration(openHour-currentHour) * time.Hour)
				status.NextStateChange = openTime.Format("2006-01-02 15:04:05")
				status.TimeToNextChange = formatDuration(openTime.Sub(now))
			} else {
				status.Status = "after-hours"
				nextOpen := now.AddDate(0, 0, 1).Add(time.Duration(openHour) * time.Hour)
				status.NextStateChange = nextOpen.Format("2006-01-02 15:04:05")
				status.TimeToNextChange = formatDuration(nextOpen.Sub(now))
			}
		}

		statuses = append(statuses, status)
	}

	return statuses
}

// ============================================
// HTTP Handlers
// ============================================

type TradingSessionHandler struct {
	service     *TradingSessionService
	authService *auth.Service
}

func NewTradingSessionHandler(service *TradingSessionService, authService *auth.Service) *TradingSessionHandler {
	return &TradingSessionHandler{
		service:     service,
		authService: authService,
	}
}

// 1. GET /admin/sessions/market-hours - List all trading sessions
func (h *TradingSessionHandler) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessions := h.service.GetAllSessions()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// 2. GET /admin/sessions/market-hours/:symbol - Get session for specific symbol
func (h *TradingSessionHandler) HandleGetSymbolSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	symbol := strings.TrimPrefix(r.URL.Path, "/admin/sessions/market-hours/")
	if symbol == "" {
		http.Error(w, "Symbol required", http.StatusBadRequest)
		return
	}

	session := h.service.GetSymbolSession(symbol)
	if session == nil {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(session)
}

// 3. PUT /admin/sessions/market-hours/:symbol - Update symbol trading hours
func (h *TradingSessionHandler) HandleUpdateSymbolSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	symbol := strings.TrimPrefix(r.URL.Path, "/admin/sessions/market-hours/")
	if symbol == "" {
		http.Error(w, "Symbol required", http.StatusBadRequest)
		return
	}

	var req struct {
		OpenTime  string `json:"openTime"`
		CloseTime string `json:"closeTime"`
		Timezone  string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateSymbolSession(symbol, req.OpenTime, req.CloseTime, req.Timezone); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Symbol session updated successfully",
		"symbol":  symbol,
	})
}

// 4. GET /admin/sessions/templates - List session templates
func (h *TradingSessionHandler) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templates := h.service.GetTemplates()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// 5. POST /admin/sessions/templates - Create custom session template
func (h *TradingSessionHandler) HandleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name          string   `json:"name"`
		Type          string   `json:"type"`
		Description   string   `json:"description"`
		OpenTime      string   `json:"openTime"`
		CloseTime     string   `json:"closeTime"`
		Timezone      string   `json:"timezone"`
		WeekendClosed bool     `json:"weekendClosed"`
		ClosedDays    []string `json:"closedDays"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	template := h.service.CreateTemplate(req.Name, req.Type, req.Description, req.OpenTime, req.CloseTime, req.Timezone, req.WeekendClosed, req.ClosedDays)
	json.NewEncoder(w).Encode(template)
}

// 6. GET /admin/sessions/holidays - List holiday calendar
func (h *TradingSessionHandler) HandleListHolidays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	holidays := h.service.GetHolidays()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"holidays": holidays,
		"count":    len(holidays),
	})
}

// 7. POST /admin/sessions/holidays - Add holiday market closure
func (h *TradingSessionHandler) HandleCreateHoliday(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Date             string   `json:"date"`
		Name             string   `json:"name"`
		AffectedSessions []string `json:"affectedSessions"`
		ClosureType      string   `json:"closureType"`
		PartialHours     string   `json:"partialHours"`
		Region           string   `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	holiday := h.service.CreateHoliday(req.Date, req.Name, req.AffectedSessions, req.ClosureType, req.PartialHours, req.Region)
	json.NewEncoder(w).Encode(holiday)
}

// 8. GET /admin/sessions/status - Get current market status per session
func (h *TradingSessionHandler) HandleGetMarketStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	statuses := h.service.GetMarketStatus()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"statuses": statuses,
		"count":    len(statuses),
	})
}
