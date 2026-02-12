package core

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// SetDB attaches a PostgreSQL connection to the engine for persistence
func (e *Engine) SetDB(db *sql.DB) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.db = db
	log.Println("[Persistence] Database connection attached to engine")
}

// SetNotificationManager attaches a notification manager for sending notifications
func (e *Engine) SetNotificationManager(manager NotificationSender) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.notificationManager = manager
	log.Println("[Engine] Notification manager attached to engine")
}

// persistAccount saves a newly created account to the database
func (e *Engine) persistAccount(account *Account) error {
	if e.db == nil {
		return nil // Skip persistence if DB not configured
	}

	query := `
		INSERT INTO rtx_accounts (
			id, user_id, account_number, currency, balance, leverage,
			margin_mode, status, is_demo, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			balance = EXCLUDED.balance,
			leverage = EXCLUDED.leverage,
			margin_mode = EXCLUDED.margin_mode,
			status = EXCLUDED.status,
			updated_at = NOW()
	`

	createdAt := time.Unix(account.CreatedAt, 0)
	if account.CreatedAt == 0 {
		createdAt = time.Now()
	}

	_, err := e.db.Exec(query,
		account.ID,
		account.UserID,
		account.AccountNumber,
		account.Currency,
		account.Balance,
		account.Leverage,
		account.MarginMode,
		account.Status,
		account.IsDemo,
		createdAt,
	)

	if err != nil {
		log.Printf("[Persistence] ERROR persisting account %s: %v", account.AccountNumber, err)
		return err
	}

	log.Printf("[Persistence] Account %s persisted to database (ID: %d)", account.AccountNumber, account.ID)
	return nil
}

// persistBalance updates the account balance in the database
func (e *Engine) persistBalance(accountID int64, newBalance float64) error {
	if e.db == nil {
		return nil // Skip persistence if DB not configured
	}

	query := `
		UPDATE rtx_accounts
		SET balance = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := e.db.Exec(query, newBalance, accountID)
	if err != nil {
		log.Printf("[Persistence] ERROR updating balance for account %d: %v", accountID, err)
		return err
	}

	return nil
}

// persistLedgerEntry saves a ledger transaction to the database
func (e *Engine) persistLedgerEntry(entry *LedgerEntry) error {
	if e.db == nil {
		return nil // Skip persistence if DB not configured
	}

	query := `
		INSERT INTO rtx_ledger (
			id, account_id, type, amount, balance_after, currency,
			description, ref_type, ref_id, admin_id, payment_method,
			payment_ref, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	// Convert admin_id string to UUID format (nullable)
	var adminID interface{}
	if entry.AdminID != "" {
		adminID = entry.AdminID
	}

	// Convert ref_id to proper type (nullable)
	var refID interface{}
	if entry.RefID > 0 {
		refID = entry.RefID
	}

	_, err := e.db.Exec(query,
		entry.ID,
		entry.AccountID,
		entry.Type,
		entry.Amount,
		entry.BalanceAfter,
		entry.Currency,
		entry.Description,
		entry.RefType,
		refID,
		adminID,
		entry.PaymentMethod,
		entry.PaymentRef,
		entry.Status,
		entry.CreatedAt,
	)

	if err != nil {
		log.Printf("[Persistence] ERROR persisting ledger entry %d: %v", entry.ID, err)
		return err
	}

	return nil
}

// LoadFromDB restores accounts and balances from the database on startup
func (e *Engine) LoadFromDB() error {
	if e.db == nil {
		log.Println("[Persistence] No database configured - skipping data load")
		return nil
	}

	log.Println("[Persistence] Loading accounts from database...")

	// Start a transaction for consistent read
	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Load accounts
	accountQuery := `
		SELECT id, user_id, account_number, currency, balance, leverage,
		       margin_mode, status, is_demo, EXTRACT(EPOCH FROM created_at)::BIGINT
		FROM rtx_accounts
		WHERE status != 'CLOSED'
		ORDER BY id
	`

	rows, err := tx.Query(accountQuery)
	if err != nil {
		return fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	accountCount := 0
	maxAccountID := int64(0)

	e.mu.Lock()
	defer e.mu.Unlock()

	for rows.Next() {
		account := &Account{}
		var userIDStr string

		err := rows.Scan(
			&account.ID,
			&userIDStr,
			&account.AccountNumber,
			&account.Currency,
			&account.Balance,
			&account.Leverage,
			&account.MarginMode,
			&account.Status,
			&account.IsDemo,
			&account.CreatedAt,
		)

		if err != nil {
			log.Printf("[Persistence] ERROR scanning account row: %v", err)
			continue
		}

		account.UserID = userIDStr
		account.Equity = account.Balance // Initial equity = balance
		account.FreeMargin = account.Balance
		account.MarginLevel = 0
		account.Margin = 0
		account.Positions = []*Position{}
		account.Orders = []*Order{}

		e.accounts[account.ID] = account
		e.ledger.balances[account.ID] = account.Balance

		accountCount++
		if account.ID > maxAccountID {
			maxAccountID = account.ID
		}

		log.Printf("[Persistence] Loaded account %s (ID: %d, Balance: %.2f)",
			account.AccountNumber, account.ID, account.Balance)
	}

	// Update next account ID counter
	if maxAccountID > 0 {
		// Set next ID to be one more than the max ID we found
		// The CreateAccount function uses len(accounts)+1, but we need to ensure
		// it doesn't collide with existing IDs
		if int64(len(e.accounts)) <= maxAccountID {
			// We'll need to track this differently - for now just log a warning
			log.Printf("[Persistence] WARNING: Max account ID (%d) >= account count (%d)",
				maxAccountID, len(e.accounts))
		}
	}

	// Load ledger entries to rebuild transaction history
	ledgerQuery := `
		SELECT id, account_id, type, amount, balance_after, currency,
		       description, ref_type, COALESCE(ref_id, 0), COALESCE(admin_id::TEXT, ''),
		       COALESCE(payment_method, ''), COALESCE(payment_ref, ''),
		       status, created_at
		FROM rtx_ledger
		WHERE account_id = ANY(SELECT id FROM rtx_accounts WHERE status != 'CLOSED')
		ORDER BY account_id, created_at
	`

	ledgerRows, err := tx.Query(ledgerQuery)
	if err != nil {
		return fmt.Errorf("failed to query ledger: %w", err)
	}
	defer ledgerRows.Close()

	ledgerCount := 0
	maxLedgerID := int64(0)

	for ledgerRows.Next() {
		entry := LedgerEntry{}

		err := ledgerRows.Scan(
			&entry.ID,
			&entry.AccountID,
			&entry.Type,
			&entry.Amount,
			&entry.BalanceAfter,
			&entry.Currency,
			&entry.Description,
			&entry.RefType,
			&entry.RefID,
			&entry.AdminID,
			&entry.PaymentMethod,
			&entry.PaymentRef,
			&entry.Status,
			&entry.CreatedAt,
		)

		if err != nil {
			log.Printf("[Persistence] ERROR scanning ledger row: %v", err)
			continue
		}

		// Add to in-memory ledger
		e.ledger.entries[entry.AccountID] = append(e.ledger.entries[entry.AccountID], entry)
		ledgerCount++

		if entry.ID > maxLedgerID {
			maxLedgerID = entry.ID
		}
	}

	// Update ledger next ID counter
	if maxLedgerID >= e.ledger.nextID {
		e.ledger.nextID = maxLedgerID + 1
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load open positions (using e.db since tx is committed)
	positionQuery := `
		SELECT id, account_id, symbol, side, volume, open_price, open_time,
		       sl, tp, swap, commission, unrealized_pnl, status
		FROM rtx_positions
		WHERE status = 'OPEN'
		ORDER BY id
	`

	posRows, err := e.db.Query(positionQuery)
	if err != nil {
		return fmt.Errorf("failed to query positions: %w", err)
	}
	defer posRows.Close()

	positionCount := 0
	maxPositionID := int64(0)

	for posRows.Next() {
		pos := &Position{}

		err := posRows.Scan(
			&pos.ID,
			&pos.AccountID,
			&pos.Symbol,
			&pos.Side,
			&pos.Volume,
			&pos.OpenPrice,
			&pos.OpenTime,
			&pos.SL,
			&pos.TP,
			&pos.Swap,
			&pos.Commission,
			&pos.UnrealizedPnL,
			&pos.Status,
		)

		if err != nil {
			log.Printf("[Persistence] ERROR scanning position row: %v", err)
			continue
		}

		e.positions[pos.ID] = pos
		positionCount++

		// Link position to account
		if account, ok := e.accounts[pos.AccountID]; ok {
			account.Positions = append(account.Positions, pos)
		}

		if pos.ID > maxPositionID {
			maxPositionID = pos.ID
		}

		log.Printf("[Persistence] Loaded position #%d: %s %s %.2f lots @ %.5f",
			pos.ID, pos.Side, pos.Symbol, pos.Volume, pos.OpenPrice)
	}

	// Update next position ID counter
	if maxPositionID >= e.nextPositionID {
		e.nextPositionID = maxPositionID + 1
	}

	// Load pending orders (using e.db since tx is committed)
	orderQuery := `
		SELECT id, account_id, symbol, type, side, volume, price, trigger_price,
		       sl, tp, status, reject_reason, filled_price, filled_at,
		       position_id, created_at
		FROM rtx_orders
		WHERE status = 'PENDING'
		ORDER BY id
	`

	orderRows, err := e.db.Query(orderQuery)
	if err != nil {
		return fmt.Errorf("failed to query orders: %w", err)
	}
	defer orderRows.Close()

	orderCount := 0
	maxOrderID := int64(0)

	for orderRows.Next() {
		order := &Order{}
		var filledAt *time.Time

		err := orderRows.Scan(
			&order.ID,
			&order.AccountID,
			&order.Symbol,
			&order.Type,
			&order.Side,
			&order.Volume,
			&order.Price,
			&order.TriggerPrice,
			&order.SL,
			&order.TP,
			&order.Status,
			&order.RejectReason,
			&order.FilledPrice,
			&filledAt,
			&order.PositionID,
			&order.CreatedAt,
		)

		if err != nil {
			log.Printf("[Persistence] ERROR scanning order row: %v", err)
			continue
		}

		order.FilledAt = filledAt

		e.orders[order.ID] = order
		orderCount++

		// Link order to account
		if account, ok := e.accounts[order.AccountID]; ok {
			account.Orders = append(account.Orders, order)
		}

		if order.ID > maxOrderID {
			maxOrderID = order.ID
		}

		log.Printf("[Persistence] Loaded order #%d: %s %s %s %.2f lots",
			order.ID, order.Type, order.Side, order.Symbol, order.Volume)
	}

	// Update next order ID counter
	if maxOrderID >= e.nextOrderID {
		e.nextOrderID = maxOrderID + 1
	}

	// Load max trade ID to restore nextTradeID counter
	var maxTradeID int64
	tradeIDQuery := `SELECT COALESCE(MAX(id), 0) FROM rtx_trades`
	if err := e.db.QueryRow(tradeIDQuery).Scan(&maxTradeID); err != nil {
		log.Printf("[Persistence] WARNING: Failed to load max trade ID: %v", err)
	} else if maxTradeID >= e.nextTradeID {
		e.nextTradeID = maxTradeID + 1
		log.Printf("[Persistence] Restored nextTradeID to %d", e.nextTradeID)
	}

	log.Printf("[Persistence] ✓ Loaded %d accounts, %d positions, %d orders, and %d ledger entries from database",
		accountCount, positionCount, orderCount, ledgerCount)

	return nil
}

// persistPosition saves a newly created position to the database
func (e *Engine) persistPosition(pos *Position) error {
	if e.db == nil {
		return nil // Skip persistence if DB not configured
	}

	query := `
		INSERT INTO rtx_positions (
			id, account_id, symbol, side, volume, open_price, open_time,
			sl, tp, swap, commission, unrealized_pnl, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			volume = EXCLUDED.volume,
			sl = EXCLUDED.sl,
			tp = EXCLUDED.tp,
			swap = EXCLUDED.swap,
			commission = EXCLUDED.commission,
			unrealized_pnl = EXCLUDED.unrealized_pnl,
			status = EXCLUDED.status
	`

	_, err := e.db.Exec(query,
		pos.ID,
		pos.AccountID,
		pos.Symbol,
		pos.Side,
		pos.Volume,
		pos.OpenPrice,
		pos.OpenTime,
		pos.SL,
		pos.TP,
		pos.Swap,
		pos.Commission,
		pos.UnrealizedPnL,
		pos.Status,
	)

	if err != nil {
		log.Printf("[Persistence] ERROR persisting position %d: %v", pos.ID, err)
		return err
	}

	log.Printf("[Persistence] Position #%d persisted to database", pos.ID)
	return nil
}

// updatePosition updates a position in the database (for SL/TP modifications)
func (e *Engine) updatePosition(pos *Position) error {
	if e.db == nil {
		return nil
	}

	query := `
		UPDATE rtx_positions
		SET sl = $1, tp = $2, volume = $3, swap = $4, commission = $5,
		    unrealized_pnl = $6, status = $7
		WHERE id = $8
	`

	_, err := e.db.Exec(query, pos.SL, pos.TP, pos.Volume, pos.Swap,
		pos.Commission, pos.UnrealizedPnL, pos.Status, pos.ID)

	if err != nil {
		log.Printf("[Persistence] ERROR updating position %d: %v", pos.ID, err)
		return err
	}

	return nil
}

// closePosition updates a position's close details in the database
func (e *Engine) closePosition(posID int64, closePrice float64, closeTime time.Time, closeReason string) error {
	if e.db == nil {
		return nil
	}

	query := `
		UPDATE rtx_positions
		SET status = 'CLOSED', close_price = $1, close_time = $2, close_reason = $3
		WHERE id = $4
	`

	_, err := e.db.Exec(query, closePrice, closeTime, closeReason, posID)

	if err != nil {
		log.Printf("[Persistence] ERROR closing position %d: %v", posID, err)
		return err
	}

	log.Printf("[Persistence] Position #%d closed in database", posID)
	return nil
}

// persistOrder saves an order to the database
func (e *Engine) persistOrder(order *Order) error {
	if e.db == nil {
		return nil
	}

	query := `
		INSERT INTO rtx_orders (
			id, account_id, symbol, type, side, volume, price, trigger_price,
			sl, tp, status, reject_reason, filled_price, filled_at,
			position_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			reject_reason = EXCLUDED.reject_reason,
			filled_price = EXCLUDED.filled_price,
			filled_at = EXCLUDED.filled_at,
			position_id = EXCLUDED.position_id
	`

	_, err := e.db.Exec(query,
		order.ID,
		order.AccountID,
		order.Symbol,
		order.Type,
		order.Side,
		order.Volume,
		order.Price,
		order.TriggerPrice,
		order.SL,
		order.TP,
		order.Status,
		order.RejectReason,
		order.FilledPrice,
		order.FilledAt,
		order.PositionID,
		order.CreatedAt,
	)

	if err != nil {
		log.Printf("[Persistence] ERROR persisting order %d: %v", order.ID, err)
		return err
	}

	log.Printf("[Persistence] Order #%d persisted to database", order.ID)
	return nil
}

// persistTrade saves a trade execution record to the database
func (e *Engine) persistTrade(trade *Trade) error {
	if e.db == nil {
		return nil
	}

	query := `
		INSERT INTO rtx_trades (
			id, order_id, position_id, account_id, symbol, side, volume,
			price, realized_pnl, commission, swap, executed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	// Handle potential zero values for swap (not in all Trade structs)
	swap := 0.0
	if trade.Commission != 0 { // Assuming swap might be stored elsewhere
		swap = 0.0
	}

	_, err := e.db.Exec(query,
		trade.ID,
		trade.OrderID,
		trade.PositionID,
		trade.AccountID,
		trade.Symbol,
		trade.Side,
		trade.Volume,
		trade.Price,
		trade.RealizedPnL,
		trade.Commission,
		swap,
		trade.ExecutedAt,
	)

	if err != nil {
		log.Printf("[Persistence] ERROR persisting trade %d: %v", trade.ID, err)
		return err
	}

	log.Printf("[Persistence] Trade #%d persisted to database", trade.ID)
	return nil
}

// Transaction-aware executors (use sql.Tx interface)
type dbExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// persistPositionTx saves a position using a transaction
func (e *Engine) persistPositionTx(tx dbExecutor, pos *Position) error {
	query := `
		INSERT INTO rtx_positions (
			id, account_id, symbol, side, volume, open_price, open_time,
			sl, tp, swap, commission, unrealized_pnl, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			volume = EXCLUDED.volume,
			sl = EXCLUDED.sl,
			tp = EXCLUDED.tp,
			swap = EXCLUDED.swap,
			commission = EXCLUDED.commission,
			unrealized_pnl = EXCLUDED.unrealized_pnl,
			status = EXCLUDED.status
	`

	_, err := tx.Exec(query,
		pos.ID,
		pos.AccountID,
		pos.Symbol,
		pos.Side,
		pos.Volume,
		pos.OpenPrice,
		pos.OpenTime,
		pos.SL,
		pos.TP,
		pos.Swap,
		pos.Commission,
		pos.UnrealizedPnL,
		pos.Status,
	)

	if err != nil {
		log.Printf("[Persistence] ERROR persisting position %d: %v", pos.ID, err)
		return err
	}

	return nil
}

// persistOrderTx saves an order using a transaction
func (e *Engine) persistOrderTx(tx dbExecutor, order *Order) error {
	query := `
		INSERT INTO rtx_orders (
			id, account_id, symbol, type, side, volume, price, trigger_price,
			sl, tp, status, reject_reason, filled_price, filled_at,
			position_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			reject_reason = EXCLUDED.reject_reason,
			filled_price = EXCLUDED.filled_price,
			filled_at = EXCLUDED.filled_at,
			position_id = EXCLUDED.position_id
	`

	_, err := tx.Exec(query,
		order.ID,
		order.AccountID,
		order.Symbol,
		order.Type,
		order.Side,
		order.Volume,
		order.Price,
		order.TriggerPrice,
		order.SL,
		order.TP,
		order.Status,
		order.RejectReason,
		order.FilledPrice,
		order.FilledAt,
		order.PositionID,
		order.CreatedAt,
	)

	if err != nil {
		log.Printf("[Persistence] ERROR persisting order %d: %v", order.ID, err)
		return err
	}

	return nil
}

// persistTradeTx saves a trade using a transaction
func (e *Engine) persistTradeTx(tx dbExecutor, trade *Trade) error {
	query := `
		INSERT INTO rtx_trades (
			id, order_id, position_id, account_id, symbol, side, volume,
			price, realized_pnl, commission, swap, executed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	swap := 0.0

	_, err := tx.Exec(query,
		trade.ID,
		trade.OrderID,
		trade.PositionID,
		trade.AccountID,
		trade.Symbol,
		trade.Side,
		trade.Volume,
		trade.Price,
		trade.RealizedPnL,
		trade.Commission,
		swap,
		trade.ExecutedAt,
	)

	if err != nil {
		log.Printf("[Persistence] ERROR persisting trade %d: %v", trade.ID, err)
		return err
	}

	return nil
}

// closePositionTx updates position close details using a transaction
func (e *Engine) closePositionTx(tx dbExecutor, posID int64, closePrice float64, closeTime time.Time, closeReason string) error {
	query := `
		UPDATE rtx_positions
		SET status = 'CLOSED', close_price = $1, close_time = $2, close_reason = $3
		WHERE id = $4
	`

	_, err := tx.Exec(query, closePrice, closeTime, closeReason, posID)

	if err != nil {
		log.Printf("[Persistence] ERROR closing position %d: %v", posID, err)
		return err
	}

	return nil
}

// updatePositionTx updates a position using a transaction
func (e *Engine) updatePositionTx(tx dbExecutor, pos *Position) error {
	query := `
		UPDATE rtx_positions
		SET sl = $1, tp = $2, volume = $3, swap = $4, commission = $5,
		    unrealized_pnl = $6, status = $7
		WHERE id = $8
	`

	_, err := tx.Exec(query, pos.SL, pos.TP, pos.Volume, pos.Swap,
		pos.Commission, pos.UnrealizedPnL, pos.Status, pos.ID)

	if err != nil {
		log.Printf("[Persistence] ERROR updating position %d: %v", pos.ID, err)
		return err
	}

	return nil
}
