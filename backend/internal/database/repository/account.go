package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Account matches bbook.Account structure
type Account struct {
	ID            int64
	AccountNumber string
	UserID        string
	Username      string
	Password      string
	Balance       float64
	Equity        float64
	Margin        float64
	FreeMargin    float64
	MarginLevel   float64
	Leverage      float64
	MarginMode    string
	Currency      string
	Status        string
	IsDemo        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AccountRepository struct {
	pool *pgxpool.Pool
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

// Create inserts a new account
func (r *AccountRepository) Create(ctx context.Context, acc *Account) error {
	query := `
		INSERT INTO accounts (
			account_number, user_id, username, password, balance, equity,
			margin, free_margin, margin_level, leverage, margin_mode,
			currency, status, is_demo, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		acc.AccountNumber, acc.UserID, acc.Username, acc.Password,
		acc.Balance, acc.Equity, acc.Margin, acc.FreeMargin, acc.MarginLevel,
		acc.Leverage, acc.MarginMode, acc.Currency, acc.Status, acc.IsDemo,
	).Scan(&acc.ID, &acc.CreatedAt, &acc.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

// GetByID retrieves account by ID
func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*Account, error) {
	query := `
		SELECT id, account_number, user_id, username, password, balance, equity,
		       margin, free_margin, margin_level, leverage, margin_mode,
		       currency, status, is_demo, created_at, updated_at
		FROM accounts WHERE id = $1
	`

	var acc Account
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.AccountNumber, &acc.UserID, &acc.Username, &acc.Password,
		&acc.Balance, &acc.Equity, &acc.Margin, &acc.FreeMargin, &acc.MarginLevel,
		&acc.Leverage, &acc.MarginMode, &acc.Currency, &acc.Status, &acc.IsDemo,
		&acc.CreatedAt, &acc.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("account not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return &acc, nil
}

// GetByAccountNumber retrieves account by account number
func (r *AccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*Account, error) {
	query := `
		SELECT id, account_number, user_id, username, password, balance, equity,
		       margin, free_margin, margin_level, leverage, margin_mode,
		       currency, status, is_demo, created_at, updated_at
		FROM accounts WHERE account_number = $1
	`

	var acc Account
	err := r.pool.QueryRow(ctx, query, accountNumber).Scan(
		&acc.ID, &acc.AccountNumber, &acc.UserID, &acc.Username, &acc.Password,
		&acc.Balance, &acc.Equity, &acc.Margin, &acc.FreeMargin, &acc.MarginLevel,
		&acc.Leverage, &acc.MarginMode, &acc.Currency, &acc.Status, &acc.IsDemo,
		&acc.CreatedAt, &acc.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("account not found: %s", accountNumber)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return &acc, nil
}

// UpdateBalance updates account balance (use transaction for financial operations)
func (r *AccountRepository) UpdateBalance(ctx context.Context, id int64, balance, equity, margin, freeMargin, marginLevel float64) error {
	// Use REPEATABLE READ isolation for financial consistency (per RESEARCH.md)
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE accounts
		SET balance = $1, equity = $2, margin = $3, free_margin = $4,
		    margin_level = $5, updated_at = NOW()
		WHERE id = $6
	`

	_, err = tx.Exec(ctx, query, balance, equity, margin, freeMargin, marginLevel, id)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// List retrieves all accounts
func (r *AccountRepository) List(ctx context.Context) ([]*Account, error) {
	query := `
		SELECT id, account_number, user_id, username, password, balance, equity,
		       margin, free_margin, margin_level, leverage, margin_mode,
		       currency, status, is_demo, created_at, updated_at
		FROM accounts ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*Account
	for rows.Next() {
		var acc Account
		err := rows.Scan(
			&acc.ID, &acc.AccountNumber, &acc.UserID, &acc.Username, &acc.Password,
			&acc.Balance, &acc.Equity, &acc.Margin, &acc.FreeMargin, &acc.MarginLevel,
			&acc.Leverage, &acc.MarginMode, &acc.Currency, &acc.Status, &acc.IsDemo,
			&acc.CreatedAt, &acc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, &acc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return accounts, nil
}
