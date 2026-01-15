package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SymbolMarginConfig represents symbol-specific margin configuration
// CRITICAL: All DECIMAL columns stored as strings to avoid float precision issues
type SymbolMarginConfig struct {
	Symbol          string
	AssetClass      string // forex_major, forex_minor, stock, crypto, commodity, index
	MaxLeverage     string // DECIMAL(5,2) as string - ESMA limits: 30, 20, 5, 2
	MarginPercent   string // DECIMAL(5,4) as string - 1/leverage * 100
	ContractSize    string // DECIMAL(20,8) as string
	TickSize        string // DECIMAL(10,8) as string
	TickValue       string // DECIMAL(20,8) as string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type SymbolMarginConfigRepository struct {
	pool *pgxpool.Pool
}

func NewSymbolMarginConfigRepository(pool *pgxpool.Pool) *SymbolMarginConfigRepository {
	return &SymbolMarginConfigRepository{pool: pool}
}

// GetBySymbol retrieves margin configuration for a symbol
func (r *SymbolMarginConfigRepository) GetBySymbol(ctx context.Context, symbol string) (*SymbolMarginConfig, error) {
	query := `
		SELECT symbol, asset_class, max_leverage, margin_percentage,
		       contract_size, tick_size, tick_value, created_at, updated_at
		FROM symbol_margin_config
		WHERE symbol = $1
	`

	var config SymbolMarginConfig
	err := r.pool.QueryRow(ctx, query, symbol).Scan(
		&config.Symbol,
		&config.AssetClass,
		&config.MaxLeverage,
		&config.MarginPercent,
		&config.ContractSize,
		&config.TickSize,
		&config.TickValue,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("margin config not found for symbol: %s", symbol)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get margin config: %w", err)
	}

	return &config, nil
}

// GetAll retrieves all symbol margin configurations
func (r *SymbolMarginConfigRepository) GetAll(ctx context.Context) ([]*SymbolMarginConfig, error) {
	query := `
		SELECT symbol, asset_class, max_leverage, margin_percentage,
		       contract_size, tick_size, tick_value, created_at, updated_at
		FROM symbol_margin_config
		ORDER BY symbol
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all margin configs: %w", err)
	}
	defer rows.Close()

	var configs []*SymbolMarginConfig
	for rows.Next() {
		var config SymbolMarginConfig
		err := rows.Scan(
			&config.Symbol,
			&config.AssetClass,
			&config.MaxLeverage,
			&config.MarginPercent,
			&config.ContractSize,
			&config.TickSize,
			&config.TickValue,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan margin config: %w", err)
		}
		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating margin configs: %w", err)
	}

	return configs, nil
}

// Upsert inserts or updates a symbol margin configuration
func (r *SymbolMarginConfigRepository) Upsert(ctx context.Context, config *SymbolMarginConfig) error {
	query := `
		INSERT INTO symbol_margin_config (
			symbol, asset_class, max_leverage, margin_percentage,
			contract_size, tick_size, tick_value, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (symbol) DO UPDATE SET
			asset_class = EXCLUDED.asset_class,
			max_leverage = EXCLUDED.max_leverage,
			margin_percentage = EXCLUDED.margin_percentage,
			contract_size = EXCLUDED.contract_size,
			tick_size = EXCLUDED.tick_size,
			tick_value = EXCLUDED.tick_value,
			updated_at = NOW()
		RETURNING symbol, asset_class, max_leverage, margin_percentage,
		          contract_size, tick_size, tick_value, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		config.Symbol,
		config.AssetClass,
		config.MaxLeverage,
		config.MarginPercent,
		config.ContractSize,
		config.TickSize,
		config.TickValue,
	).Scan(
		&config.Symbol,
		&config.AssetClass,
		&config.MaxLeverage,
		&config.MarginPercent,
		&config.ContractSize,
		&config.TickSize,
		&config.TickValue,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert margin config: %w", err)
	}

	return nil
}

// SeedESMADefaults inserts standard ESMA leverage limits for common asset classes
func (r *SymbolMarginConfigRepository) SeedESMADefaults(ctx context.Context) error {
	// ESMA regulatory leverage limits
	// forex_major: 30:1 (3.33% margin)
	// forex_minor: 20:1 (5% margin)
	// stock: 5:1 (20% margin)
	// crypto: 2:1 (50% margin)
	// commodity: 20:1 (5% margin)
	// index: 20:1 (5% margin)

	defaults := []struct {
		symbol       string
		assetClass   string
		maxLeverage  string
		marginPct    string
		contractSize string
		tickSize     string
		tickValue    string
	}{
		// Major Forex Pairs (30:1)
		{"EURUSD", "forex_major", "30.00", "3.3333", "100000", "0.0001", "10"},
		{"GBPUSD", "forex_major", "30.00", "3.3333", "100000", "0.0001", "10"},
		{"USDJPY", "forex_major", "30.00", "3.3333", "100000", "0.01", "1000"},
		{"USDCHF", "forex_major", "30.00", "3.3333", "100000", "0.0001", "10"},
		{"AUDUSD", "forex_major", "30.00", "3.3333", "100000", "0.0001", "10"},
		{"USDCAD", "forex_major", "30.00", "3.3333", "100000", "0.0001", "10"},
		{"NZDUSD", "forex_major", "30.00", "3.3333", "100000", "0.0001", "10"},

		// Minor/Exotic Forex Pairs (20:1)
		{"EURGBP", "forex_minor", "20.00", "5.0000", "100000", "0.0001", "10"},
		{"EURJPY", "forex_minor", "20.00", "5.0000", "100000", "0.01", "1000"},
		{"GBPJPY", "forex_minor", "20.00", "5.0000", "100000", "0.01", "1000"},
		{"EURCHF", "forex_minor", "20.00", "5.0000", "100000", "0.0001", "10"},

		// Crypto (2:1)
		{"BTCUSD", "crypto", "2.00", "50.0000", "1", "0.01", "1"},
		{"ETHUSD", "crypto", "2.00", "50.0000", "1", "0.01", "1"},
		{"XRPUSD", "crypto", "2.00", "50.0000", "1", "0.0001", "1"},
		{"SOLUSD", "crypto", "2.00", "50.0000", "1", "0.01", "1"},
		{"BNBUSD", "crypto", "2.00", "50.0000", "1", "0.01", "1"},

		// Stock indices (20:1)
		{"US30", "index", "20.00", "5.0000", "1", "0.01", "1"},
		{"US500", "index", "20.00", "5.0000", "1", "0.01", "1"},
		{"NAS100", "index", "20.00", "5.0000", "1", "0.01", "1"},

		// Commodities (20:1)
		{"XAUUSD", "commodity", "20.00", "5.0000", "100", "0.01", "1"},
		{"XAGUSD", "commodity", "20.00", "5.0000", "5000", "0.001", "5"},
		{"WTIUSD", "commodity", "20.00", "5.0000", "1000", "0.01", "10"},
	}

	for _, d := range defaults {
		config := &SymbolMarginConfig{
			Symbol:        d.symbol,
			AssetClass:    d.assetClass,
			MaxLeverage:   d.maxLeverage,
			MarginPercent: d.marginPct,
			ContractSize:  d.contractSize,
			TickSize:      d.tickSize,
			TickValue:     d.tickValue,
		}

		if err := r.Upsert(ctx, config); err != nil {
			return fmt.Errorf("failed to seed default for %s: %w", d.symbol, err)
		}
	}

	return nil
}

// GetByAssetClass retrieves all symbols in an asset class
func (r *SymbolMarginConfigRepository) GetByAssetClass(ctx context.Context, assetClass string) ([]*SymbolMarginConfig, error) {
	query := `
		SELECT symbol, asset_class, max_leverage, margin_percentage,
		       contract_size, tick_size, tick_value, created_at, updated_at
		FROM symbol_margin_config
		WHERE asset_class = $1
		ORDER BY symbol
	`

	rows, err := r.pool.Query(ctx, query, assetClass)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs for asset class: %w", err)
	}
	defer rows.Close()

	var configs []*SymbolMarginConfig
	for rows.Next() {
		var config SymbolMarginConfig
		err := rows.Scan(
			&config.Symbol,
			&config.AssetClass,
			&config.MaxLeverage,
			&config.MarginPercent,
			&config.ContractSize,
			&config.TickSize,
			&config.TickValue,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan margin config: %w", err)
		}
		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating margin configs: %w", err)
	}

	return configs, nil
}
