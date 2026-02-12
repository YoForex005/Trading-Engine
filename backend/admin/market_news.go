package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// NewsArticle represents a market news article
type NewsArticle struct {
	ID              int64          `json:"id"`
	Source          string         `json:"source"`
	Category        string         `json:"category"`
	Title           string         `json:"title"`
	Summary         string         `json:"summary"`
	Content         string         `json:"content"`
	Author          string         `json:"author"`
	PublishedAt     time.Time      `json:"published_at"`
	URL             string         `json:"url"`
	Sentiment       SentimentScore `json:"sentiment"`
	AffectedSymbols []string       `json:"affected_symbols"`
	Tags            []string       `json:"tags"`
	ImageURL        string         `json:"image_url,omitempty"`
	ViewCount       int            `json:"view_count"`
	Relevance       float64        `json:"relevance"`
}

// SentimentScore represents sentiment analysis scores
type SentimentScore struct {
	Bullish float64 `json:"bullish"`
	Bearish float64 `json:"bearish"`
	Neutral float64 `json:"neutral"`
	Overall string  `json:"overall"`
}

// EconomicEvent represents an economic calendar event
type EconomicEvent struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Country         string    `json:"country"`
	CountryCode     string    `json:"country_code"`
	Category        string    `json:"category"`
	Impact          string    `json:"impact"`
	Currency        string    `json:"currency"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	Forecast        *float64  `json:"forecast,omitempty"`
	Previous        *float64  `json:"previous,omitempty"`
	Actual          *float64  `json:"actual,omitempty"`
	Unit            string    `json:"unit"`
	AffectedSymbols []string  `json:"affected_symbols"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	TradingPause    bool      `json:"trading_pause"`
}

// NewsSource represents a news source configuration
type NewsSource struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Enabled     bool      `json:"enabled"`
	Priority    int       `json:"priority"`
	URL         string    `json:"url"`
	APIKey      string    `json:"api_key,omitempty"`
	UpdateFreq  int       `json:"update_frequency_mins"`
	LastUpdate  time.Time `json:"last_update"`
	ArticleCount int      `json:"article_count"`
	Reliability float64   `json:"reliability"`
}

// TradingPauseRule represents an auto-trading pause rule
type TradingPauseRule struct {
	ID              int64    `json:"id"`
	EventImpact     string   `json:"event_impact"`
	EventCategories []string `json:"event_categories"`
	PauseBefore     int      `json:"pause_before_mins"`
	PauseAfter      int      `json:"pause_after_mins"`
	AffectedSymbols []string `json:"affected_symbols"`
	Enabled         bool     `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// MarketNewsService manages market news and economic calendar
type MarketNewsService struct {
	mu               sync.RWMutex
	articles         map[int64]*NewsArticle
	events           map[int64]*EconomicEvent
	sources          map[int64]*NewsSource
	pauseRules       map[int64]*TradingPauseRule
	nextArticleID    int64
	nextEventID      int64
	nextSourceID     int64
	nextPauseRuleID  int64
}

// NewMarketNewsService creates a new market news service
func NewMarketNewsService() *MarketNewsService {
	s := &MarketNewsService{
		articles:        make(map[int64]*NewsArticle),
		events:          make(map[int64]*EconomicEvent),
		sources:         make(map[int64]*NewsSource),
		pauseRules:      make(map[int64]*TradingPauseRule),
		nextArticleID:   1,
		nextEventID:     1,
		nextSourceID:    1,
		nextPauseRuleID: 1,
	}
	s.initMockData()
	return s
}

func (s *MarketNewsService) initMockData() {
	// Initialize 5 news sources
	sources := []struct {
		name        string
		url         string
		updateFreq  int
		reliability float64
	}{
		{"Reuters", "https://reuters.com/markets", 5, 0.95},
		{"Bloomberg", "https://bloomberg.com/news", 3, 0.98},
		{"ForexFactory", "https://forexfactory.com/news", 10, 0.85},
		{"DailyFX", "https://dailyfx.com/market-news", 15, 0.82},
		{"Investing.com", "https://investing.com/news", 5, 0.88},
	}

	for _, src := range sources {
		source := &NewsSource{
			ID:          s.nextSourceID,
			Name:        src.name,
			Enabled:     true,
			Priority:    int(s.nextSourceID),
			URL:         src.url,
			UpdateFreq:  src.updateFreq,
			LastUpdate:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Minute),
			ArticleCount: 0,
			Reliability: src.reliability,
		}
		s.sources[s.nextSourceID] = source
		s.nextSourceID++
	}

	// Generate 50 news articles
	categories := []string{"forex", "commodities", "indices", "crypto", "central_bank", "geopolitical"}
	titles := map[string][]string{
		"forex": {
			"Dollar Strengthens on Fed Rate Hike Expectations",
			"EUR/USD Falls to 3-Month Low on ECB Dovish Stance",
			"Yen Weakens as BoJ Maintains Ultra-Loose Policy",
			"Sterling Rallies on Strong UK Employment Data",
			"Swiss Franc Gains Safe-Haven Status Amid Turmoil",
		},
		"commodities": {
			"Gold Surges Past $2,000 on Inflation Fears",
			"Oil Prices Drop 5% on China Demand Concerns",
			"Silver Breaks Out Above Key Resistance Level",
			"Copper Rallies on Supply Shortage Worries",
			"Natural Gas Futures Spike on Cold Weather Forecast",
		},
		"indices": {
			"S&P 500 Hits All-Time High on Tech Earnings",
			"Nasdaq Drops 2% on Interest Rate Fears",
			"FTSE 100 Rallies on Mining Stocks Strength",
			"DAX Reaches New Record High on Economic Data",
			"Nikkei 225 Falls on Yen Strength",
		},
		"crypto": {
			"Bitcoin Breaks $50,000 Resistance Level",
			"Ethereum Surges 15% on Network Upgrade News",
			"Crypto Markets Rally on ETF Approval Hopes",
			"Regulatory Concerns Weigh on Altcoin Prices",
			"Institutional Demand Drives Bitcoin Higher",
		},
		"central_bank": {
			"Fed Signals Three More Rate Hikes This Year",
			"ECB Keeps Rates Unchanged, Dovish Outlook Remains",
			"Bank of England Surprises with 50bp Rate Hike",
			"BoJ Maintains Ultra-Loose Monetary Policy Stance",
			"RBA Signals End to Tightening Cycle",
		},
		"geopolitical": {
			"Trade Tensions Escalate Between US and China",
			"Middle East Tensions Drive Oil Prices Higher",
			"Brexit Negotiations Enter Critical Phase",
			"OPEC+ Announces Surprise Production Cut",
			"G7 Summit Focuses on Global Economic Challenges",
		},
	}

	symbolsByCategory := map[string][]string{
		"forex":        {"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCHF"},
		"commodities":  {"XAUUSD", "XAGUSD", "WTICOUSD", "NATGASUSD", "XCUUSD"},
		"indices":      {"SPX500USD", "NAS100USD", "UK100GBP", "DE30EUR", "JP225USD"},
		"crypto":       {"BTCUSD", "ETHUSD", "XRPUSD", "BNBUSD", "SOLUSD"},
		"central_bank": {"EURUSD", "GBPUSD", "USDJPY", "AUDUSD"},
		"geopolitical": {"XAUUSD", "WTICOUSD", "EURUSD", "GBPUSD"},
	}

	sourcesArr := []string{"Reuters", "Bloomberg", "ForexFactory", "DailyFX", "Investing.com"}

	for i := 0; i < 50; i++ {
		category := categories[i%len(categories)]
		source := sourcesArr[i%len(sourcesArr)]
		titlesList := titles[category]
		title := titlesList[rand.Intn(len(titlesList))]

		// Generate sentiment scores
		sentimentType := rand.Intn(3)
		var sentiment SentimentScore
		switch sentimentType {
		case 0: // Bullish
			sentiment = SentimentScore{
				Bullish: 65.0 + rand.Float64()*30.0,
				Bearish: 5.0 + rand.Float64()*10.0,
				Neutral: 10.0 + rand.Float64()*20.0,
				Overall: "bullish",
			}
		case 1: // Bearish
			sentiment = SentimentScore{
				Bullish: 5.0 + rand.Float64()*10.0,
				Bearish: 65.0 + rand.Float64()*30.0,
				Neutral: 10.0 + rand.Float64()*20.0,
				Overall: "bearish",
			}
		case 2: // Neutral
			sentiment = SentimentScore{
				Bullish: 25.0 + rand.Float64()*20.0,
				Bearish: 25.0 + rand.Float64()*20.0,
				Neutral: 40.0 + rand.Float64()*20.0,
				Overall: "neutral",
			}
		}

		article := &NewsArticle{
			ID:       s.nextArticleID,
			Source:   source,
			Category: category,
			Title:    title,
			Summary:  fmt.Sprintf("Summary of %s article from %s", category, source),
			Content:  fmt.Sprintf("Full content of %s news article discussing market movements and analysis", category),
			Author:   fmt.Sprintf("Author %d", rand.Intn(20)+1),
			PublishedAt: time.Now().Add(-time.Duration(rand.Intn(72)) * time.Hour),
			URL:      fmt.Sprintf("https://%s.com/article/%d", strings.ToLower(source), s.nextArticleID),
			Sentiment: sentiment,
			AffectedSymbols: symbolsByCategory[category],
			Tags:     []string{category, "analysis", "markets"},
			ViewCount: rand.Intn(10000) + 100,
			Relevance: 0.5 + rand.Float64()*0.5,
		}

		s.articles[s.nextArticleID] = article
		s.nextArticleID++

		// Update source article count
		for _, src := range s.sources {
			if src.Name == source {
				src.ArticleCount++
				break
			}
		}
	}

	// Generate 30 economic calendar events
	events := []struct {
		name        string
		country     string
		countryCode string
		category    string
		impact      string
		currency    string
		unit        string
		symbols     []string
	}{
		{"Non-Farm Payrolls (NFP)", "United States", "US", "employment", "high", "USD", "K", []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD"}},
		{"Consumer Price Index (CPI)", "United States", "US", "inflation", "high", "USD", "%", []string{"EURUSD", "GBPUSD", "USDJPY", "XAUUSD"}},
		{"FOMC Interest Rate Decision", "United States", "US", "monetary_policy", "high", "USD", "%", []string{"EURUSD", "GBPUSD", "USDJPY", "SPX500USD"}},
		{"Gross Domestic Product (GDP)", "United States", "US", "growth", "high", "USD", "%", []string{"EURUSD", "USDJPY", "SPX500USD"}},
		{"Retail Sales", "United States", "US", "consumer", "medium", "USD", "%", []string{"EURUSD", "USDJPY"}},
		{"ECB Interest Rate Decision", "Eurozone", "EU", "monetary_policy", "high", "EUR", "%", []string{"EURUSD", "EURGBP", "EURJPY"}},
		{"Eurozone CPI", "Eurozone", "EU", "inflation", "high", "EUR", "%", []string{"EURUSD", "EURGBP"}},
		{"Eurozone GDP", "Eurozone", "EU", "growth", "high", "EUR", "%", []string{"EURUSD", "DE30EUR"}},
		{"UK Interest Rate Decision", "United Kingdom", "GB", "monetary_policy", "high", "GBP", "%", []string{"GBPUSD", "EURGBP", "GBPJPY"}},
		{"UK CPI", "United Kingdom", "GB", "inflation", "high", "GBP", "%", []string{"GBPUSD", "EURGBP"}},
		{"UK Employment Change", "United Kingdom", "GB", "employment", "medium", "GBP", "K", []string{"GBPUSD"}},
		{"BoJ Interest Rate Decision", "Japan", "JP", "monetary_policy", "high", "JPY", "%", []string{"USDJPY", "EURJPY", "GBPJPY"}},
		{"Japan CPI", "Japan", "JP", "inflation", "medium", "JPY", "%", []string{"USDJPY"}},
		{"Manufacturing PMI (US)", "United States", "US", "business", "medium", "USD", "index", []string{"EURUSD", "SPX500USD"}},
		{"Services PMI (US)", "United States", "US", "business", "medium", "USD", "index", []string{"EURUSD", "SPX500USD"}},
		{"ISM Manufacturing PMI", "United States", "US", "business", "medium", "USD", "index", []string{"EURUSD", "USDJPY"}},
		{"ADP Employment Change", "United States", "US", "employment", "medium", "USD", "K", []string{"EURUSD", "USDJPY"}},
		{"Unemployment Rate", "United States", "US", "employment", "high", "USD", "%", []string{"EURUSD", "USDJPY"}},
		{"Trade Balance", "United States", "US", "trade", "low", "USD", "B", []string{"EURUSD"}},
		{"Producer Price Index (PPI)", "United States", "US", "inflation", "medium", "USD", "%", []string{"EURUSD", "XAUUSD"}},
		{"Building Permits", "United States", "US", "housing", "low", "USD", "M", []string{"EURUSD"}},
		{"Housing Starts", "United States", "US", "housing", "low", "USD", "M", []string{"EURUSD"}},
		{"Consumer Confidence", "United States", "US", "consumer", "medium", "USD", "index", []string{"EURUSD", "SPX500USD"}},
		{"Industrial Production", "United States", "US", "production", "low", "USD", "%", []string{"EURUSD"}},
		{"RBA Interest Rate Decision", "Australia", "AU", "monetary_policy", "high", "AUD", "%", []string{"AUDUSD", "AUDNZD"}},
		{"China GDP", "China", "CN", "growth", "high", "CNY", "%", []string{"AUDUSD", "USDCNH", "AU200AUD"}},
		{"China CPI", "China", "CN", "inflation", "medium", "CNY", "%", []string{"AUDUSD", "USDCNH"}},
		{"OPEC Meeting", "OPEC", "OPEC", "commodities", "high", "OIL", "", []string{"WTICOUSD", "BCOUSD"}},
		{"EIA Crude Oil Inventories", "United States", "US", "commodities", "medium", "USD", "M barrels", []string{"WTICOUSD"}},
		{"RBNZ Interest Rate Decision", "New Zealand", "NZ", "monetary_policy", "high", "NZD", "%", []string{"NZDUSD", "AUDNZD"}},
	}

	for _, evt := range events {
		hoursAhead := rand.Intn(168) + 1 // 1-168 hours (1-7 days)
		scheduledAt := time.Now().Add(time.Duration(hoursAhead) * time.Hour)

		// Generate forecast and previous values
		var forecast, previous, actual *float64
		baseValue := rand.Float64() * 5.0 // 0-5%

		prev := baseValue + (rand.Float64()-0.5)*2.0
		previous = &prev

		fore := prev + (rand.Float64()-0.5)*1.0
		forecast = &fore

		// Only past events have actual values
		if hoursAhead > 100 || rand.Intn(10) < 2 {
			act := fore + (rand.Float64()-0.5)*0.5
			actual = &act
		}

		status := "scheduled"
		if actual != nil {
			status = "completed"
		}

		tradingPause := evt.impact == "high"

		event := &EconomicEvent{
			ID:              s.nextEventID,
			Name:            evt.name,
			Country:         evt.country,
			CountryCode:     evt.countryCode,
			Category:        evt.category,
			Impact:          evt.impact,
			Currency:        evt.currency,
			ScheduledAt:     scheduledAt,
			Forecast:        forecast,
			Previous:        previous,
			Actual:          actual,
			Unit:            evt.unit,
			AffectedSymbols: evt.symbols,
			Description:     fmt.Sprintf("Economic indicator measuring %s for %s", evt.category, evt.country),
			Status:          status,
			TradingPause:    tradingPause,
		}

		s.events[s.nextEventID] = event
		s.nextEventID++
	}

	// Generate trading pause rules
	pauseRules := []struct {
		impact      string
		categories  []string
		pauseBefore int
		pauseAfter  int
		symbols     []string
	}{
		{"high", []string{"monetary_policy", "employment", "inflation"}, 15, 30, []string{"EURUSD", "GBPUSD", "USDJPY"}},
		{"high", []string{"growth"}, 10, 20, []string{"SPX500USD", "NAS100USD"}},
		{"high", []string{"commodities"}, 5, 15, []string{"WTICOUSD", "XAUUSD"}},
	}

	for _, rule := range pauseRules {
		pauseRule := &TradingPauseRule{
			ID:              s.nextPauseRuleID,
			EventImpact:     rule.impact,
			EventCategories: rule.categories,
			PauseBefore:     rule.pauseBefore,
			PauseAfter:      rule.pauseAfter,
			AffectedSymbols: rule.symbols,
			Enabled:         true,
			CreatedAt:       time.Now().AddDate(0, 0, -rand.Intn(90)),
			UpdatedAt:       time.Now(),
		}

		s.pauseRules[s.nextPauseRuleID] = pauseRule
		s.nextPauseRuleID++
	}

	log.Println("[MarketNews] Mock data initialized: 50 articles from 5 sources, 30 economic events, 3 trading pause rules")
}

// GetArticles returns news articles with optional filtering
func (s *MarketNewsService) GetArticles(source, category string, since *time.Time) []*NewsArticle {
	s.mu.RLock()
	defer s.mu.RUnlock()

	articles := make([]*NewsArticle, 0)
	for _, article := range s.articles {
		if source != "" && article.Source != source {
			continue
		}
		if category != "" && article.Category != category {
			continue
		}
		if since != nil && article.PublishedAt.Before(*since) {
			continue
		}
		articles = append(articles, article)
	}
	return articles
}

// GetArticle returns a specific article by ID
func (s *MarketNewsService) GetArticle(id int64) (*NewsArticle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	article, exists := s.articles[id]
	if !exists {
		return nil, fmt.Errorf("article not found")
	}
	return article, nil
}

// GetEconomicEvents returns upcoming economic events
func (s *MarketNewsService) GetEconomicEvents(impact, country string) []*EconomicEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]*EconomicEvent, 0)
	for _, event := range s.events {
		if impact != "" && event.Impact != impact {
			continue
		}
		if country != "" && event.CountryCode != country {
			continue
		}
		events = append(events, event)
	}
	return events
}

// GetEconomicEvent returns a specific event by ID
func (s *MarketNewsService) GetEconomicEvent(id int64) (*EconomicEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[id]
	if !exists {
		return nil, fmt.Errorf("event not found")
	}
	return event, nil
}

// UpdateEconomicEvent updates event actual values
func (s *MarketNewsService) UpdateEconomicEvent(id int64, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, exists := s.events[id]
	if !exists {
		return fmt.Errorf("event not found")
	}

	if actual, ok := updates["actual"].(float64); ok {
		event.Actual = &actual
		event.Status = "completed"
	}
	if status, ok := updates["status"].(string); ok {
		event.Status = status
	}

	return nil
}

// GetNewsSources returns all news source configurations
func (s *MarketNewsService) GetNewsSources() []*NewsSource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sources := make([]*NewsSource, 0, len(s.sources))
	for _, source := range s.sources {
		sources = append(sources, source)
	}
	return sources
}

// UpdateNewsSource updates source settings
func (s *MarketNewsService) UpdateNewsSource(id int64, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	source, exists := s.sources[id]
	if !exists {
		return fmt.Errorf("source not found")
	}

	if enabled, ok := updates["enabled"].(bool); ok {
		source.Enabled = enabled
	}
	if priority, ok := updates["priority"].(float64); ok {
		source.Priority = int(priority)
	}
	if updateFreq, ok := updates["update_frequency_mins"].(float64); ok {
		source.UpdateFreq = int(updateFreq)
	}

	return nil
}

// GetStats returns news and calendar statistics
func (s *MarketNewsService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Article stats by category
	articlesByCategory := make(map[string]int)
	articlesBySentiment := make(map[string]int)
	totalViews := 0

	for _, article := range s.articles {
		articlesByCategory[article.Category]++
		articlesBySentiment[article.Sentiment.Overall]++
		totalViews += article.ViewCount
	}

	// Event stats by impact
	eventsByImpact := make(map[string]int)
	upcomingEvents := 0
	completedEvents := 0

	for _, event := range s.events {
		eventsByImpact[event.Impact]++
		if event.Status == "scheduled" {
			upcomingEvents++
		} else {
			completedEvents++
		}
	}

	return map[string]interface{}{
		"total_articles":         len(s.articles),
		"total_events":           len(s.events),
		"total_sources":          len(s.sources),
		"articles_by_category":   articlesByCategory,
		"articles_by_sentiment":  articlesBySentiment,
		"events_by_impact":       eventsByImpact,
		"upcoming_events":        upcomingEvents,
		"completed_events":       completedEvents,
		"total_article_views":    totalViews,
		"active_trading_pauses":  len(s.pauseRules),
	}
}

// MarketNewsHandler handles HTTP requests for market news
type MarketNewsHandler struct {
	service     *MarketNewsService
	authService *AuthService
}

// NewMarketNewsHandler creates a new market news handler
func NewMarketNewsHandler(service *MarketNewsService, authService *AuthService) *MarketNewsHandler {
	return &MarketNewsHandler{
		service:     service,
		authService: authService,
	}
}

// HandleListArticles handles GET /admin/market-news/articles
func (h *MarketNewsHandler) HandleListArticles(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	source := r.URL.Query().Get("source")
	category := r.URL.Query().Get("category")
	sinceStr := r.URL.Query().Get("since")

	var since *time.Time
	if sinceStr != "" {
		t, err := time.Parse(time.RFC3339, sinceStr)
		if err == nil {
			since = &t
		}
	}

	articles := h.service.GetArticles(source, category, since)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"articles": articles,
		"total":    len(articles),
	})
}

// HandleGetArticle handles GET /admin/market-news/articles/:id
func (h *MarketNewsHandler) HandleGetArticle(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/market-news/articles/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	article, err := h.service.GetArticle(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(article)
}

// HandleListCalendar handles GET /admin/market-news/calendar
func (h *MarketNewsHandler) HandleListCalendar(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	impact := r.URL.Query().Get("impact")
	country := r.URL.Query().Get("country")

	events := h.service.GetEconomicEvents(impact, country)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

// HandleGetCalendarEvent handles GET /admin/market-news/calendar/:id
func (h *MarketNewsHandler) HandleGetCalendarEvent(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/market-news/calendar/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := h.service.GetEconomicEvent(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(event)
}

// HandleUpdateCalendarEvent handles PUT /admin/market-news/calendar/:id
func (h *MarketNewsHandler) HandleUpdateCalendarEvent(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/market-news/calendar/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateEconomicEvent(id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Event updated successfully",
	})
}

// HandleListSources handles GET /admin/market-news/sources
func (h *MarketNewsHandler) HandleListSources(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sources := h.service.GetNewsSources()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sources": sources,
		"total":   len(sources),
	})
}

// HandleUpdateSource handles PUT /admin/market-news/sources/:id
func (h *MarketNewsHandler) HandleUpdateSource(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/market-news/sources/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateNewsSource(id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Source updated successfully",
	})
}

// HandleGetStats handles GET /admin/market-news/stats
func (h *MarketNewsHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}
