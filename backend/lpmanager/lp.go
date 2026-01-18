package lpmanager

import (
	"time"
)

// Quote represents a price quote from any LP
type Quote struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Timestamp int64   `json:"timestamp"`
	LP        string  `json:"lp"`
}

// SymbolInfo represents information about a tradeable symbol
type SymbolInfo struct {
	Symbol        string  `json:"symbol"`
	DisplayName   string  `json:"displayName"`
	BaseCurrency  string  `json:"baseCurrency"`
	QuoteCurrency string  `json:"quoteCurrency"`
	MinLotSize    float64 `json:"minLotSize"`
	MaxLotSize    float64 `json:"maxLotSize"`
	LotStep       float64 `json:"lotStep"`
	PipValue      float64 `json:"pipValue"`
	Type          string  `json:"type"` // "forex", "crypto", "commodity", "index"
}

// LPStatus represents the current status of an LP
type LPStatus struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Connected    bool      `json:"connected"`
	Enabled      bool      `json:"enabled"`
	SymbolCount  int       `json:"symbolCount"`
	LastTick     time.Time `json:"lastTick"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
}

// LPAdapter is the interface that all LP implementations must satisfy
type LPAdapter interface {
	// ID returns the unique identifier for this LP
	ID() string

	// Name returns the human-readable name of this LP
	Name() string

	// Type returns the connection type (REST, WebSocket, FIX)
	Type() string

	// Connect establishes connection to the LP
	Connect() error

	// Disconnect closes the connection to the LP
	Disconnect() error

	// IsConnected returns true if currently connected
	IsConnected() bool

	// GetSymbols fetches all available symbols from the LP
	GetSymbols() ([]SymbolInfo, error)

	// Subscribe starts streaming prices for the given symbols
	Subscribe(symbols []string) error

	// Unsubscribe stops streaming prices for the given symbols
	Unsubscribe(symbols []string) error

	// GetQuotesChan returns the channel for receiving quotes
	GetQuotesChan() <-chan Quote

	// GetStatus returns the current status of this LP
	GetStatus() LPStatus
}

// LPConfig represents the configuration for an LP
type LPConfig struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"` // "OANDA", "BINANCE", "FLEXYMARKETS", etc.
	Enabled  bool              `json:"enabled"`
	Priority int               `json:"priority"` // Lower = higher priority
	Settings map[string]string `json:"settings"` // API keys, endpoints, etc.
	Symbols  []string          `json:"symbols"`  // Empty = all available
}

// LPManagerConfig represents the full LP manager configuration
type LPManagerConfig struct {
	LPs          []LPConfig `json:"lps"`
	PrimaryLP    string     `json:"primaryLp"` // ID of primary LP for execution
	LastModified int64      `json:"lastModified"`
}

// NewDefaultConfig creates a default LP configuration
func NewDefaultConfig() *LPManagerConfig {
	return &LPManagerConfig{
		LPs: []LPConfig{
			{
				ID:       "oanda",
				Name:     "OANDA",
				Type:     "OANDA",
				Enabled:  true,
				Priority: 1,
				Settings: map[string]string{},
				Symbols:  []string{}, // All available
			},
			{
				ID:       "binance",
				Name:     "Binance",
				Type:     "BINANCE",
				Enabled:  true,
				Priority: 2,
				Settings: map[string]string{},
				Symbols:  []string{}, // All available
			},
		},
		PrimaryLP:    "oanda",
		LastModified: time.Now().Unix(),
	}
}
