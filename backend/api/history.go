package api

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/tickstore"
)

// HistoryHandler handles historical tick data API endpoints
type HistoryHandler struct {
	tickStore tickstore.TickStorageService

	// Rate limiting
	rateLimiter *RateLimiter

	// Symbol metadata cache
	symbolCache map[string]SymbolMetadata
	symbolMu    sync.RWMutex
}

// SymbolMetadata contains metadata about a symbol's available data
type SymbolMetadata struct {
	Symbol        string    `json:"symbol"`
	EarliestTick  time.Time `json:"earliest_tick"`
	LatestTick    time.Time `json:"latest_tick"`
	TotalTicks    int64     `json:"total_ticks"`
	AvailableDays int       `json:"available_days"`
	LastUpdated   time.Time `json:"last_updated"`
}

// TicksResponse is the response format for tick data
type TicksResponse struct {
	Symbol     string             `json:"symbol"`
	From       time.Time          `json:"from"`
	To         time.Time          `json:"to"`
	Count      int                `json:"count"`
	TotalCount int64              `json:"total_count,omitempty"`
	Page       int                `json:"page,omitempty"`
	PageSize   int                `json:"page_size,omitempty"`
	HasMore    bool               `json:"has_more,omitempty"`
	Ticks      []tickstore.Tick   `json:"ticks"`
	Format     string             `json:"format"`
}

// BulkTicksRequest is the request format for bulk download
type BulkTicksRequest struct {
	Symbols []string  `json:"symbols"`
	From    time.Time `json:"from"`
	To      time.Time `json:"to"`
	Format  string    `json:"format"` // json, csv, binary
}

// BulkTicksResponse is the response format for bulk download
type BulkTicksResponse struct {
	Symbols []string               `json:"symbols"`
	From    time.Time              `json:"from"`
	To      time.Time              `json:"to"`
	Data    map[string][]tickstore.Tick `json:"data"`
	Count   int                    `json:"count"`
}

// AvailableDataResponse lists available symbols and their date ranges
type AvailableDataResponse struct {
	Symbols []SymbolMetadata `json:"symbols"`
	Total   int              `json:"total"`
}

// BackfillRequest is the request format for backfilling historical data
type BackfillRequest struct {
	Symbol string             `json:"symbol"`
	Ticks  []tickstore.Tick   `json:"ticks"`
	Source string             `json:"source"` // Source of the data (e.g., "external", "provider_name")
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu            sync.Mutex
	tokens        map[string]int
	maxTokens     int
	refillRate    int // tokens per second
	lastRefill    map[string]time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens, refillRate int) *RateLimiter {
	return &RateLimiter{
		tokens:     make(map[string]int),
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: make(map[string]time.Time),
	}
}

// Allow checks if a request is allowed for the given key (IP address)
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Initialize if first request
	if _, exists := rl.tokens[key]; !exists {
		rl.tokens[key] = rl.maxTokens
		rl.lastRefill[key] = now
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(rl.lastRefill[key]).Seconds()
	tokensToAdd := int(elapsed * float64(rl.refillRate))

	if tokensToAdd > 0 {
		rl.tokens[key] = min(rl.maxTokens, rl.tokens[key]+tokensToAdd)
		rl.lastRefill[key] = now
	}

	// Check if we have tokens available
	if rl.tokens[key] > 0 {
		rl.tokens[key]--
		return true
	}

	return false
}

// NewHistoryHandler creates a new history API handler
func NewHistoryHandler(ts tickstore.TickStorageService) *HistoryHandler {
	return &HistoryHandler{
		tickStore:   ts,
		rateLimiter: NewRateLimiter(100, 10), // 100 tokens, refill 10/sec
		symbolCache: make(map[string]SymbolMetadata),
	}
}

// RegisterRoutes registers all history API routes with standard http.ServeMux
func (h *HistoryHandler) RegisterRoutes(mux *http.ServeMux) {
	// Public endpoints
	mux.HandleFunc("/api/history/ticks/", h.handleCORS(h.rateLimitMiddleware(h.HandleGetTicks)))
	mux.HandleFunc("/api/history/ticks", h.handleCORS(h.rateLimitMiddleware(h.HandleGetTicksQuery))) // Query param version
	mux.HandleFunc("/api/history/ticks/bulk", h.handleCORS(h.rateLimitMiddleware(h.HandleBulkDownload)))
	mux.HandleFunc("/api/history/available", h.handleCORS(h.HandleGetAvailable))
	mux.HandleFunc("/api/history/symbols", h.handleCORS(h.HandleGetSymbols))
	mux.HandleFunc("/api/history/info", h.handleCORS(h.HandleGetSymbolInfo)) // Symbol info endpoint

	// Admin endpoints (require authentication in production)
	mux.HandleFunc("/admin/history/backfill", h.handleCORS(h.HandleBackfill))
}

// handleCORS wraps a handler with CORS headers
func (h *HistoryHandler) handleCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Range")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Range, Accept-Ranges, Content-Encoding")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// rateLimitMiddleware applies rate limiting to endpoints
func (h *HistoryHandler) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as rate limit key
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		}

		if !h.rateLimiter.Allow(ip) {
			w.Header().Set("Retry-After", "10")
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// HandleGetTicks handles GET /api/history/ticks/{symbol}
// Query params: from, to, format (json/csv/binary), page, page_size
func (h *HistoryHandler) HandleGetTicks(w http.ResponseWriter, r *http.Request) {
	// Extract symbol from URL path: /api/history/ticks/EURUSD
	parts := strings.Split(r.URL.Path, "/")
	symbol := ""
	if len(parts) >= 5 {
		symbol = parts[4]
	}

	if symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	// Validate symbol to prevent path traversal
	if !isValidSymbol(symbol) {
		http.Error(w, "Invalid symbol format", http.StatusBadRequest)
		log.Printf("[HistoryAPI] Invalid symbol attempt: %s", symbol)
		return
	}

	// Parse query parameters
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	// Validate and sanitize page parameter
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 || page > 100000 {
		page = 1
	}

	// Validate and sanitize page_size parameter
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || pageSize < 1 || pageSize > 10000 {
		pageSize = 1000 // Default page size
	}

	// Parse date range
	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			http.Error(w, "Invalid 'from' date format. Use RFC3339", http.StatusBadRequest)
			return
		}
	} else {
		from = time.Now().AddDate(0, 0, -7) // Default: 7 days back
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			http.Error(w, "Invalid 'to' date format. Use RFC3339", http.StatusBadRequest)
			return
		}
	} else {
		to = time.Now()
	}

	// Calculate days back for fetching from DailyStore
	daysBack := int(to.Sub(from).Hours() / 24)
	if daysBack < 1 {
		daysBack = 1
	}

	// Fetch ticks from storage
	// For now, we'll get all ticks and filter/paginate in memory
	// TODO: Optimize with database queries for better performance
	allTicks := h.getTicksInRange(symbol, from, to, daysBack)

	// Apply pagination
	totalCount := len(allTicks)
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= totalCount {
		startIdx = 0
		endIdx = 0
	}
	if endIdx > totalCount {
		endIdx = totalCount
	}

	pageTicks := []tickstore.Tick{}
	if startIdx < endIdx {
		pageTicks = allTicks[startIdx:endIdx]
	}

	hasMore := endIdx < totalCount

	// Handle different formats
	switch format {
	case "csv":
		h.respondCSV(w, symbol, pageTicks)
	case "binary":
		h.respondBinary(w, symbol, pageTicks)
	default: // json
		// Check if compression is requested
		acceptEncoding := r.Header.Get("Accept-Encoding")
		useGzip := strings.Contains(acceptEncoding, "gzip")

		response := TicksResponse{
			Symbol:     symbol,
			From:       from,
			To:         to,
			Count:      len(pageTicks),
			TotalCount: int64(totalCount),
			Page:       page,
			PageSize:   pageSize,
			HasMore:    hasMore,
			Ticks:      pageTicks,
			Format:     format,
		}

		w.Header().Set("Content-Type", "application/json")

		if useGzip {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			json.NewEncoder(gz).Encode(response)
		} else {
			json.NewEncoder(w).Encode(response)
		}
	}

	log.Printf("[HistoryAPI] GET /api/history/ticks/%s: returned %d/%d ticks (page %d)",
		symbol, len(pageTicks), totalCount, page)
}

// HandleBulkDownload handles POST /api/history/ticks/bulk
func (h *HistoryHandler) HandleBulkDownload(w http.ResponseWriter, r *http.Request) {
	var req BulkTicksRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Symbols) == 0 {
		http.Error(w, "At least one symbol is required", http.StatusBadRequest)
		return
	}

	if len(req.Symbols) > 50 {
		http.Error(w, "Maximum 50 symbols allowed per request", http.StatusBadRequest)
		return
	}

	// Validate all symbols to prevent path traversal
	for _, symbol := range req.Symbols {
		if !isValidSymbol(symbol) {
			http.Error(w, fmt.Sprintf("Invalid symbol format: %s", symbol), http.StatusBadRequest)
			log.Printf("[HistoryAPI] Invalid symbol attempt in bulk download: %s", symbol)
			return
		}
	}

	// Default format
	if req.Format == "" {
		req.Format = "json"
	}

	// Calculate days back
	daysBack := int(req.To.Sub(req.From).Hours() / 24)
	if daysBack < 1 {
		daysBack = 7
	}

	// Fetch ticks for all symbols
	data := make(map[string][]tickstore.Tick)
	totalCount := 0

	for _, symbol := range req.Symbols {
		ticks := h.getTicksInRange(symbol, req.From, req.To, daysBack)
		data[symbol] = ticks
		totalCount += len(ticks)
	}

	response := BulkTicksResponse{
		Symbols: req.Symbols,
		From:    req.From,
		To:      req.To,
		Data:    data,
		Count:   totalCount,
	}

	// Always use gzip for bulk downloads
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"bulk_ticks_%s_%s.json.gz\"",
		req.From.Format("2006-01-02"), req.To.Format("2006-01-02")))

	gz := gzip.NewWriter(w)
	defer gz.Close()

	json.NewEncoder(gz).Encode(response)

	log.Printf("[HistoryAPI] POST /api/history/ticks/bulk: returned %d ticks for %d symbols",
		totalCount, len(req.Symbols))
}

// HandleGetAvailable handles GET /api/history/available
func (h *HistoryHandler) HandleGetAvailable(w http.ResponseWriter, r *http.Request) {
	symbols := h.tickStore.GetSymbols()

	metadataList := make([]SymbolMetadata, 0, len(symbols))

	for _, symbol := range symbols {
		metadata := h.getSymbolMetadata(symbol)
		metadataList = append(metadataList, metadata)
	}

	response := AvailableDataResponse{
		Symbols: metadataList,
		Total:   len(metadataList),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("[HistoryAPI] GET /api/history/available: returned %d symbols", len(symbols))
}

// HandleGetSymbols handles GET /api/history/symbols
func (h *HistoryHandler) HandleGetSymbols(w http.ResponseWriter, r *http.Request) {
	// Get all symbols with metadata
	symbols := h.tickStore.GetSymbols()

	type SymbolInfo struct {
		Symbol       string    `json:"symbol"`
		DisplayName  string    `json:"display_name"`
		Category     string    `json:"category"`
		Available    bool      `json:"available"`
		TickCount    int       `json:"tick_count"`
		LastUpdated  time.Time `json:"last_updated,omitempty"`
	}

	symbolList := make([]SymbolInfo, 0, len(symbols))

	for _, symbol := range symbols {
		metadata := h.getSymbolMetadata(symbol)

		info := SymbolInfo{
			Symbol:      symbol,
			DisplayName: symbol,
			Category:    h.categorizeSymbol(symbol),
			Available:   metadata.TotalTicks > 0,
			TickCount:   int(metadata.TotalTicks),
			LastUpdated: metadata.LastUpdated,
		}

		symbolList = append(symbolList, info)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbols": symbolList,
		"total":   len(symbolList),
	})

	log.Printf("[HistoryAPI] GET /api/history/symbols: returned %d symbols", len(symbolList))
}

// HandleBackfill handles POST /admin/history/backfill
func (h *HistoryHandler) HandleBackfill(w http.ResponseWriter, r *http.Request) {
	var req BackfillRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	// Validate symbol to prevent path traversal
	if !isValidSymbol(req.Symbol) {
		http.Error(w, "Invalid symbol format", http.StatusBadRequest)
		log.Printf("[HistoryAPI] Invalid symbol attempt in backfill: %s", req.Symbol)
		return
	}

	if len(req.Ticks) == 0 {
		http.Error(w, "No ticks provided", http.StatusBadRequest)
		return
	}

	// Get DailyStore from TickStore (if available)
	if ts, ok := h.tickStore.(*tickstore.TickStore); ok {
		dailyStore := ts.GetDailyStore()

		if err := dailyStore.MergeHistoricalData(req.Symbol, req.Ticks); err != nil {
			http.Error(w, fmt.Sprintf("Backfill failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Invalidate cache for this symbol
		h.symbolMu.Lock()
		delete(h.symbolCache, req.Symbol)
		h.symbolMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"symbol":  req.Symbol,
			"count":   len(req.Ticks),
			"source":  req.Source,
			"message": fmt.Sprintf("Successfully backfilled %d ticks for %s", len(req.Ticks), req.Symbol),
		})

		log.Printf("[HistoryAPI] POST /admin/history/backfill: backfilled %d ticks for %s from %s",
			len(req.Ticks), req.Symbol, req.Source)
	} else {
		http.Error(w, "Backfill not supported with current storage implementation", http.StatusNotImplemented)
	}
}

// Helper: getTicksInRange fetches ticks in a date range
func (h *HistoryHandler) getTicksInRange(symbol string, from, to time.Time, daysBack int) []tickstore.Tick {
	// Try to get from DailyStore if available
	if ts, ok := h.tickStore.(*tickstore.TickStore); ok {
		dailyStore := ts.GetDailyStore()
		allTicks := dailyStore.GetHistory(symbol, 0, daysBack)

		// Filter by date range
		filtered := make([]tickstore.Tick, 0)
		for _, tick := range allTicks {
			if (tick.Timestamp.Equal(from) || tick.Timestamp.After(from)) &&
			   (tick.Timestamp.Equal(to) || tick.Timestamp.Before(to)) {
				filtered = append(filtered, tick)
			}
		}
		return filtered
	}

	// Fallback to regular GetHistory
	return h.tickStore.GetHistory(symbol, 10000) // Max 10k ticks
}

// Helper: getSymbolMetadata gets or computes metadata for a symbol
func (h *HistoryHandler) getSymbolMetadata(symbol string) SymbolMetadata {
	h.symbolMu.RLock()
	if cached, exists := h.symbolCache[symbol]; exists {
		// Cache valid for 5 minutes
		if time.Since(cached.LastUpdated) < 5*time.Minute {
			h.symbolMu.RUnlock()
			return cached
		}
	}
	h.symbolMu.RUnlock()

	// Compute metadata
	metadata := SymbolMetadata{
		Symbol:      symbol,
		LastUpdated: time.Now(),
	}

	// Get tick data to compute metadata
	if ts, ok := h.tickStore.(*tickstore.TickStore); ok {
		dailyStore := ts.GetDailyStore()
		dates := dailyStore.GetAvailableDates(symbol)
		metadata.AvailableDays = len(dates)

		if len(dates) > 0 {
			// Get earliest and latest ticks
			earliestTicks := dailyStore.GetHistory(symbol, 1, 365) // Look back 1 year
			if len(earliestTicks) > 0 {
				metadata.EarliestTick = earliestTicks[0].Timestamp
				metadata.LatestTick = earliestTicks[len(earliestTicks)-1].Timestamp
			}

			// Estimate total ticks (rough)
			metadata.TotalTicks = int64(len(earliestTicks))
		}
	} else {
		ticks := h.tickStore.GetHistory(symbol, 0)
		metadata.TotalTicks = int64(len(ticks))
		if len(ticks) > 0 {
			metadata.EarliestTick = ticks[0].Timestamp
			metadata.LatestTick = ticks[len(ticks)-1].Timestamp
		}
	}

	// Cache it
	h.symbolMu.Lock()
	h.symbolCache[symbol] = metadata
	h.symbolMu.Unlock()

	return metadata
}

// Helper: categorizeSymbol categorizes a symbol by type
func (h *HistoryHandler) categorizeSymbol(symbol string) string {
	symbol = strings.ToUpper(symbol)

	if strings.Contains(symbol, "USD") || strings.Contains(symbol, "EUR") ||
	   strings.Contains(symbol, "GBP") || strings.Contains(symbol, "JPY") {
		return "forex"
	}

	if strings.Contains(symbol, "XAU") || strings.Contains(symbol, "XAG") {
		return "metals"
	}

	if strings.Contains(symbol, "WTI") || strings.Contains(symbol, "BRENT") {
		return "energy"
	}

	return "other"
}

// Helper: respondCSV responds with CSV format
func (h *HistoryHandler) respondCSV(w http.ResponseWriter, symbol string, ticks []tickstore.Tick) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_ticks.csv\"", symbol))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	writer.Write([]string{"Timestamp", "Symbol", "Bid", "Ask", "Spread", "LP"})

	// Data
	for _, tick := range ticks {
		writer.Write([]string{
			tick.Timestamp.Format(time.RFC3339),
			tick.Symbol,
			fmt.Sprintf("%.5f", tick.Bid),
			fmt.Sprintf("%.5f", tick.Ask),
			fmt.Sprintf("%.5f", tick.Spread),
			tick.LP,
		})
	}
}

// Helper: respondBinary responds with binary format (custom protocol)
func (h *HistoryHandler) respondBinary(w http.ResponseWriter, symbol string, ticks []tickstore.Tick) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_ticks.bin\"", symbol))

	// Simple binary format: each tick is 40 bytes
	// 8 bytes timestamp (Unix nano), 8 bytes bid, 8 bytes ask, 8 bytes spread, 8 bytes reserved
	// For simplicity, we'll use JSON encoding for now
	// TODO: Implement efficient binary protocol
	json.NewEncoder(w).Encode(ticks)
}

// Helper: min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper: isValidSymbol validates symbol format to prevent path traversal
func isValidSymbol(symbol string) bool {
	// Only allow alphanumeric characters (A-Z, 0-9)
	// Prevents path traversal attacks like "../../../etc/passwd"
	for _, c := range symbol {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return len(symbol) > 0 && len(symbol) <= 20
}

// HandleGetSymbolInfo handles GET /api/history/info?symbol=XXX
// Returns metadata about a specific symbol's historical data
func (h *HistoryHandler) HandleGetSymbolInfo(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	// Validate symbol to prevent path traversal
	if !isValidSymbol(symbol) {
		http.Error(w, "Invalid symbol format", http.StatusBadRequest)
		log.Printf("[HistoryAPI] Invalid symbol attempt: %s", symbol)
		return
	}

	metadata := h.getSymbolMetadata(symbol)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbol":        metadata.Symbol,
		"earliestDate":  metadata.EarliestTick.Format("2006-01-02"),
		"latestDate":    metadata.LatestTick.Format("2006-01-02"),
		"tickCount":     metadata.TotalTicks,
		"availableDays": metadata.AvailableDays,
		"lastUpdated":   metadata.LastUpdated.Format(time.RFC3339),
	})

	log.Printf("[HistoryAPI] GET /api/history/info?symbol=%s: returned metadata", symbol)
}

// HandleGetTicksQuery handles GET /api/history/ticks?symbol=XXX&date=YYYY-MM-DD&offset=0&limit=5000
// Alternative endpoint that accepts query parameters instead of path parameters
func (h *HistoryHandler) HandleGetTicksQuery(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	// Validate symbol to prevent path traversal
	if !isValidSymbol(symbol) {
		http.Error(w, "Invalid symbol format", http.StatusBadRequest)
		log.Printf("[HistoryAPI] Invalid symbol attempt: %s", symbol)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		// Default to today
		dateStr = time.Now().Format("2006-01-02")
	}

	// Validate and sanitize offset parameter
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 || offset > 1000000 {
		offset = 0
	}

	// Validate and sanitize limit parameter
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 5000
	}
	if limit > 50000 {
		limit = 50000 // Cap at 50k
	}

	// Parse date and get ticks for that date
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Get ticks from tickstore
	var allTicks []tickstore.Tick
	if _, ok := h.tickStore.(*tickstore.TickStore); ok {
		// Get ticks for the specific date (1 day back from target date end)
		endDate := targetDate.AddDate(0, 0, 1)
		allTicks = h.getTicksInRange(symbol, targetDate, endDate, 1)
	} else {
		allTicks = h.tickStore.GetHistory(symbol, limit+offset)
	}

	// Apply offset and limit
	total := len(allTicks)
	if offset >= total {
		allTicks = []tickstore.Tick{}
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		allTicks = allTicks[offset:end]
	}

	// Convert to response format
	ticks := make([]map[string]interface{}, len(allTicks))
	for i, t := range allTicks {
		ticks[i] = map[string]interface{}{
			"timestamp": t.Timestamp.UnixMilli(),
			"bid":       t.Bid,
			"ask":       t.Ask,
			"spread":    t.Spread,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbol": symbol,
		"date":   dateStr,
		"ticks":  ticks,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})

	log.Printf("[HistoryAPI] GET /api/history/ticks?symbol=%s&date=%s: returned %d/%d ticks",
		symbol, dateStr, len(ticks), total)
}
