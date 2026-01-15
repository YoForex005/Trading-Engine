// Package bbook - persistence.go
// DEPRECATED: This file implements JSON file-based persistence.
// As of Phase 2 (Database Migration - Plan 02-03), all persistence is handled by the repository layer.
// This code is kept for reference but should not be used in production.
//
// See: backend/internal/database/repository/
// Migration: backend/internal/migration/migrate_data.go

package bbook

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const dataDir = "./data/bbook"

// PersistenceData holds all data to be persisted
type PersistenceData struct {
	Accounts       map[int64]*Account  `json:"accounts"`
	Positions      map[int64]*Position `json:"positions"`
	Orders         map[int64]*Order    `json:"orders"`
	Trades         []Trade             `json:"trades"`
	NextPositionID int64               `json:"nextPositionId"`
	NextOrderID    int64               `json:"nextOrderId"`
	NextTradeID    int64               `json:"nextTradeId"`
	SavedAt        time.Time           `json:"savedAt"`
}

// Save persists engine state to disk
func (e *Engine) Save() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	data := PersistenceData{
		Accounts:       e.accounts,
		Positions:      e.positions,
		Orders:         e.orders,
		Trades:         e.trades,
		NextPositionID: e.nextPositionID,
		NextOrderID:    e.nextOrderID,
		NextTradeID:    e.nextTradeID,
		SavedAt:        time.Now(),
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(dataDir, "engine_state.json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return err
	}

	log.Printf("[B-Book] State saved: %d accounts, %d positions, %d trades",
		len(e.accounts), len(e.positions), len(e.trades))
	return nil
}

// Load restores engine state from disk
func (e *Engine) Load() error {
	filePath := filepath.Join(dataDir, "engine_state.json")

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("[B-Book] No saved state found, starting fresh")
			return nil
		}
		return err
	}
	defer file.Close()

	var data PersistenceData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		log.Printf("[B-Book] Failed to decode saved state: %v", err)
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.accounts = data.Accounts
	e.positions = data.Positions
	e.orders = data.Orders
	e.trades = data.Trades
	e.nextPositionID = data.NextPositionID
	e.nextOrderID = data.NextOrderID
	e.nextTradeID = data.NextTradeID

	// Initialize nil maps
	if e.accounts == nil {
		e.accounts = make(map[int64]*Account)
	}
	if e.positions == nil {
		e.positions = make(map[int64]*Position)
	}
	if e.orders == nil {
		e.orders = make(map[int64]*Order)
	}
	if e.trades == nil {
		e.trades = make([]Trade, 0)
	}

	log.Printf("[B-Book] State loaded: %d accounts, %d positions, %d trades (saved at %s)",
		len(e.accounts), len(e.positions), len(e.trades), data.SavedAt.Format(time.RFC3339))
	return nil
}

// StartAutoSave starts a background goroutine that saves state periodically
func (e *Engine) StartAutoSave(interval time.Duration, wg *sync.WaitGroup) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := e.Save(); err != nil {
				log.Printf("[B-Book] Auto-save failed: %v", err)
			}
		}
	}()
	log.Printf("[B-Book] Auto-save enabled (every %s)", interval)
}
