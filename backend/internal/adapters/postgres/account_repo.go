package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epic1st/rtx/backend/internal/database/repository"
	"github.com/epic1st/rtx/backend/internal/domain/account"
	"github.com/epic1st/rtx/backend/internal/ports"
)

// AccountRepositoryAdapter implements ports.AccountRepository
type AccountRepositoryAdapter struct {
	repo *repository.AccountRepository
}

// Verify interface implementation at compile time
var _ ports.AccountRepository = (*AccountRepositoryAdapter)(nil)

// NewAccountRepositoryAdapter creates a new account repository adapter
func NewAccountRepositoryAdapter(pool *pgxpool.Pool) *AccountRepositoryAdapter {
	return &AccountRepositoryAdapter{
		repo: repository.NewAccountRepository(pool),
	}
}

// toDomain converts repository.Account to domain.Account
func (a *AccountRepositoryAdapter) toDomain(repoAcc *repository.Account) *account.Account {
	return &account.Account{
		ID:            repoAcc.ID,
		AccountNumber: repoAcc.AccountNumber,
		UserID:        repoAcc.UserID,
		Username:      repoAcc.Username,
		Password:      repoAcc.Password,
		Balance:       repoAcc.Balance,
		Equity:        repoAcc.Equity,
		Margin:        repoAcc.Margin,
		FreeMargin:    repoAcc.FreeMargin,
		MarginLevel:   repoAcc.MarginLevel,
		Leverage:      repoAcc.Leverage,
		MarginMode:    repoAcc.MarginMode,
		Currency:      repoAcc.Currency,
		Status:        repoAcc.Status,
		IsDemo:        repoAcc.IsDemo,
		CreatedAt:     repoAcc.CreatedAt.Unix(),
	}
}

// toRepository converts domain.Account to repository.Account
func (a *AccountRepositoryAdapter) toRepository(domainAcc *account.Account) *repository.Account {
	return &repository.Account{
		ID:            domainAcc.ID,
		AccountNumber: domainAcc.AccountNumber,
		UserID:        domainAcc.UserID,
		Username:      domainAcc.Username,
		Password:      domainAcc.Password,
		Balance:       domainAcc.Balance,
		Equity:        domainAcc.Equity,
		Margin:        domainAcc.Margin,
		FreeMargin:    domainAcc.FreeMargin,
		MarginLevel:   domainAcc.MarginLevel,
		Leverage:      domainAcc.Leverage,
		MarginMode:    domainAcc.MarginMode,
		Currency:      domainAcc.Currency,
		Status:        domainAcc.Status,
		IsDemo:        domainAcc.IsDemo,
		CreatedAt:     time.Unix(domainAcc.CreatedAt, 0),
	}
}

// Get retrieves an account by ID
func (a *AccountRepositoryAdapter) Get(ctx context.Context, id int64) (*account.Account, error) {
	repoAcc, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account %d: %w", id, err)
	}
	return a.toDomain(repoAcc), nil
}

// GetByAccountNumber retrieves an account by account number
func (a *AccountRepositoryAdapter) GetByAccountNumber(ctx context.Context, accountNumber string) (*account.Account, error) {
	repoAcc, err := a.repo.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get account %s: %w", accountNumber, err)
	}
	return a.toDomain(repoAcc), nil
}

// GetByUsername retrieves an account by username
func (a *AccountRepositoryAdapter) GetByUsername(ctx context.Context, username string) (*account.Account, error) {
	// Note: This needs to be added to repository.AccountRepository
	// For now, we'll search through all accounts (not efficient, but works for migration)
	accounts, err := a.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, acc := range accounts {
		if acc.Username == username {
			return acc, nil
		}
	}
	return nil, fmt.Errorf("account with username %s not found", username)
}

// Create creates a new account
func (a *AccountRepositoryAdapter) Create(ctx context.Context, acc *account.Account) error {
	repoAcc := a.toRepository(acc)
	if err := a.repo.Create(ctx, repoAcc); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	// Update the domain account with generated ID and timestamps
	acc.ID = repoAcc.ID
	acc.CreatedAt = repoAcc.CreatedAt.Unix()
	return nil
}

// Update updates an existing account
func (a *AccountRepositoryAdapter) Update(ctx context.Context, acc *account.Account) error {
	err := a.repo.UpdateBalance(ctx, acc.ID, acc.Balance, acc.Equity, acc.Margin, acc.FreeMargin, acc.MarginLevel)
	if err != nil {
		return fmt.Errorf("failed to update account %d: %w", acc.ID, err)
	}
	return nil
}

// List retrieves all accounts
func (a *AccountRepositoryAdapter) List(ctx context.Context) ([]*account.Account, error) {
	repoAccounts, err := a.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	accounts := make([]*account.Account, len(repoAccounts))
	for i, repoAcc := range repoAccounts {
		accounts[i] = a.toDomain(repoAcc)
	}
	return accounts, nil
}

// Delete deletes an account (not implemented in repository yet)
func (a *AccountRepositoryAdapter) Delete(ctx context.Context, id int64) error {
	return fmt.Errorf("delete operation not implemented for account %d", id)
}
