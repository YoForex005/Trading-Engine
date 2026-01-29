package bbook

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// SymbolConfig represents a symbol configuration that can be loaded/saved
type SymbolConfig struct {
	Symbols []SymbolSpec `json:"symbols"`
}

// SymbolCategory represents the asset class of a symbol
type SymbolCategory string

const (
	CategoryForexMajor  SymbolCategory = "FOREX_MAJOR"
	CategoryForexMinor  SymbolCategory = "FOREX_MINOR"
	CategoryForexExotic SymbolCategory = "FOREX_EXOTIC"
	CategoryCrypto      SymbolCategory = "CRYPTO"
	CategoryMetals      SymbolCategory = "METALS"
	CategoryCommodities SymbolCategory = "COMMODITIES"
	CategoryIndices     SymbolCategory = "INDICES"
	CategoryBonds       SymbolCategory = "BONDS"
	CategoryUnknown     SymbolCategory = "UNKNOWN"
)

// DetectSymbolCategory auto-detects the category of a symbol based on naming patterns
func DetectSymbolCategory(symbol string) SymbolCategory {
	symbol = strings.ToUpper(symbol)

	// Metals (Gold, Silver, Platinum, Palladium, Copper)
	if strings.HasPrefix(symbol, "XAU") || strings.HasPrefix(symbol, "XAG") ||
		strings.HasPrefix(symbol, "XPT") || strings.HasPrefix(symbol, "XPD") ||
		strings.HasPrefix(symbol, "XCU") {
		return CategoryMetals
	}

	// Crypto (BTC, ETH, BNB, SOL, XRP, etc.)
	cryptoPatterns := []string{"BTC", "ETH", "BNB", "SOL", "XRP", "LTC", "DOGE", "ADA", "DOT", "AVAX"}
	for _, pattern := range cryptoPatterns {
		if strings.Contains(symbol, pattern) {
			return CategoryCrypto
		}
	}

	// Indices (patterns: US30, NAS100, SPX500, JP225, DE30, UK100, etc.)
	indexPatterns := []string{"US30", "US2000", "NAS100", "SPX500", "JP225", "DE30", "UK100", "FR40",
		"EU50", "AU200", "CN50", "HK33", "SG30", "NL25", "CH20", "ESPIX", "CHINAH"}
	for _, pattern := range indexPatterns {
		if strings.Contains(symbol, pattern) {
			return CategoryIndices
		}
	}

	// Bonds (USB, UK10YB, DE10YB patterns)
	if strings.Contains(symbol, "USB") || strings.Contains(symbol, "YB") {
		return CategoryBonds
	}

	// Commodities (OIL, GAS, CORN, WHEAT, SUGAR, SOYBN)
	commodityPatterns := []string{"BCO", "WTICO", "NATGAS", "CORN", "WHEAT", "SUGAR", "SOYBN"}
	for _, pattern := range commodityPatterns {
		if strings.Contains(symbol, pattern) {
			return CategoryCommodities
		}
	}

	// Forex - check if it's 6-7 chars and contains currency codes
	currencies := []string{"USD", "EUR", "GBP", "JPY", "AUD", "NZD", "CAD", "CHF",
		"HKD", "SGD", "NOK", "SEK", "DKK", "PLN", "CZK", "HUF",
		"TRY", "ZAR", "MXN", "THB", "CNH"}

	currencyCount := 0
	for _, cur := range currencies {
		if strings.Contains(symbol, cur) {
			currencyCount++
		}
	}
	if currencyCount >= 2 && len(symbol) >= 6 && len(symbol) <= 7 {
		// Major pairs
		majors := []string{"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "NZDUSD", "USDCAD"}
		for _, major := range majors {
			if symbol == major {
				return CategoryForexMajor
			}
		}
		// Exotic pairs (contain exotic currencies)
		exotics := []string{"TRY", "ZAR", "MXN", "THB", "CNH", "PLN", "CZK", "HUF", "NOK", "SEK", "DKK", "HKD", "SGD"}
		for _, exotic := range exotics {
			if strings.Contains(symbol, exotic) {
				return CategoryForexExotic
			}
		}
		return CategoryForexMinor
	}

	return CategoryUnknown
}

// GenerateSymbolSpec auto-generates symbol specifications based on naming patterns
func GenerateSymbolSpec(symbol string) *SymbolSpec {
	category := DetectSymbolCategory(symbol)
	symbol = strings.ToUpper(symbol)

	spec := &SymbolSpec{
		Symbol:     symbol,
		MinVolume:  0.01,
		MaxVolume:  100,
		VolumeStep: 0.01,
	}

	// Determine pip size based on symbol pattern
	isJPYPair := strings.Contains(symbol, "JPY") && !strings.HasPrefix(symbol, "XAU") && !strings.HasPrefix(symbol, "XAG")
	isHKDPair := strings.Contains(symbol, "HKD")

	switch category {
	case CategoryForexMajor, CategoryForexMinor:
		spec.ContractSize = 100000
		spec.MarginPercent = 1
		if isJPYPair {
			spec.PipSize = 0.01
			spec.PipValue = 9.09
		} else if isHKDPair {
			spec.PipSize = 0.0001
			spec.PipValue = 1.28
		} else {
			spec.PipSize = 0.0001
			spec.PipValue = 10
		}

	case CategoryForexExotic:
		spec.ContractSize = 100000
		spec.MarginPercent = 3
		if isJPYPair {
			spec.PipSize = 0.01
			spec.PipValue = 9.09
		} else if strings.Contains(symbol, "ZAR") || strings.Contains(symbol, "TRY") || strings.Contains(symbol, "MXN") {
			spec.PipSize = 0.0001
			spec.PipValue = 5
		} else {
			spec.PipSize = 0.0001
			spec.PipValue = 10
		}

	case CategoryCrypto:
		spec.ContractSize = 1
		spec.MarginPercent = 10
		if strings.Contains(symbol, "BTC") {
			spec.PipSize = 1
			spec.PipValue = 1
			spec.MaxVolume = 10
		} else if strings.Contains(symbol, "ETH") {
			spec.PipSize = 0.1
			spec.PipValue = 0.1
			spec.MaxVolume = 50
		} else {
			spec.PipSize = 0.01
			spec.PipValue = 0.01
		}

	case CategoryMetals:
		spec.MarginPercent = 2
		if strings.HasPrefix(symbol, "XAU") {
			spec.ContractSize = 100
			spec.PipSize = 0.01
			spec.PipValue = 1
		} else if strings.HasPrefix(symbol, "XAG") {
			spec.ContractSize = 5000
			spec.PipSize = 0.001
			spec.PipValue = 5
		} else if strings.HasPrefix(symbol, "XPT") || strings.HasPrefix(symbol, "XPD") {
			spec.ContractSize = 100
			spec.PipSize = 0.01
			spec.PipValue = 1
		} else {
			spec.ContractSize = 25000
			spec.PipSize = 0.0001
			spec.PipValue = 2.5
		}

	case CategoryCommodities:
		spec.ContractSize = 1000
		spec.MarginPercent = 5
		spec.PipSize = 0.01
		spec.PipValue = 10

	case CategoryIndices:
		spec.ContractSize = 1
		spec.MarginPercent = 5
		spec.PipSize = 0.1
		spec.PipValue = 0.1
		spec.MaxVolume = 500

	case CategoryBonds:
		spec.ContractSize = 1000
		spec.MarginPercent = 2
		spec.PipSize = 0.01
		spec.PipValue = 10

	default:
		// Unknown - use safe defaults
		spec.ContractSize = 100000
		spec.MarginPercent = 5
		spec.PipSize = 0.0001
		spec.PipValue = 10
	}

	return spec
}

// LoadSymbolsFromDirectory scans tick data directory and auto-generates specs
func (e *Engine) LoadSymbolsFromDirectory(tickDataDir string) error {
	entries, err := os.ReadDir(tickDataDir)
	if err != nil {
		return err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		symbol := entry.Name()
		symbol = strings.ToUpper(symbol)

		spec := GenerateSymbolSpec(symbol)
		e.symbols[symbol] = spec
		count++
	}

	log.Printf("[B-Book] Loaded %d symbols from tick data directory", count)
	return nil
}

// LoadSymbolsFromJSON loads symbols from a JSON configuration file
func (e *Engine) LoadSymbolsFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var config SymbolConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	for _, spec := range config.Symbols {
		specCopy := spec
		e.symbols[spec.Symbol] = &specCopy
	}

	log.Printf("[B-Book] Loaded %d symbols from JSON config", len(config.Symbols))
	return nil
}

// SaveSymbolsToJSON saves current symbols to a JSON configuration file
func (e *Engine) SaveSymbolsToJSON(path string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	config := SymbolConfig{
		Symbols: make([]SymbolSpec, 0, len(e.symbols)),
	}

	for _, spec := range e.symbols {
		config.Symbols = append(config.Symbols, *spec)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// RegisterSymbol dynamically registers a new symbol (auto-generates spec if needed)
func (e *Engine) RegisterSymbol(symbol string) *SymbolSpec {
	e.mu.Lock()
	defer e.mu.Unlock()

	symbol = strings.ToUpper(symbol)

	// Check if already exists
	if spec, ok := e.symbols[symbol]; ok {
		return spec
	}

	// Auto-generate specification
	spec := GenerateSymbolSpec(symbol)
	e.symbols[symbol] = spec

	log.Printf("[B-Book] Auto-registered symbol: %s (Category: %s)", symbol, DetectSymbolCategory(symbol))
	return spec
}

// GetOrCreateSymbol gets a symbol spec, creating it dynamically if it doesn't exist
func (e *Engine) GetOrCreateSymbol(symbol string) *SymbolSpec {
	e.mu.RLock()
	spec, ok := e.symbols[symbol]
	e.mu.RUnlock()

	if ok {
		return spec
	}

	return e.RegisterSymbol(symbol)
}

// DiscoverSymbolsFromFIX is called when FIX sends us a quote for a new symbol
func (e *Engine) DiscoverSymbolsFromFIX(symbol string) {
	e.GetOrCreateSymbol(symbol)
}
