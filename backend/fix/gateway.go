package fix

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// LPSession represents a connection to a Liquidity Provider
type LPSession struct {
	ID           string
	Name         string
	Host         string
	Port         int
	SenderCompID string
	TargetCompID string
	Status       string // DISCONNECTED, CONNECTING, CONNECTED, LOGGED_IN
	LastHeartbeat time.Time
}

// ExecutionReport represents a fill or reject from LP
type ExecutionReport struct {
	OrderID       string
	ExecType      string // NEW, FILLED, REJECTED, CANCELED
	Symbol        string
	Side          string
	Volume        float64
	Price         float64
	LPOrderID     string
	Text          string
	Timestamp     time.Time
}

// FIXGateway manages connections to Liquidity Providers
type FIXGateway struct {
	sessions     map[string]*LPSession
	execReports  chan ExecutionReport
	mu           sync.RWMutex
}

func NewFIXGateway() *FIXGateway {
	gw := &FIXGateway{
		sessions: map[string]*LPSession{
			"LMAX_PROD": {
				ID:           "LMAX_PROD",
				Name:         "LMAX Exchange",
				Host:         "fix.lmax.com",
				Port:         443,
				SenderCompID: "RTX_BROKER",
				TargetCompID: "LMAX",
				Status:       "DISCONNECTED",
			},
			"LMAX_DEMO": {
				ID:           "LMAX_DEMO",
				Name:         "LMAX Demo",
				Host:         "demo-fix.lmax.com",
				Port:         443,
				SenderCompID: "RTX_BROKER_DEMO",
				TargetCompID: "LMAX_DEMO",
				Status:       "DISCONNECTED",
			},
		},
		execReports: make(chan ExecutionReport, 1000),
	}
	return gw
}

// Connect initiates a FIX session (simulated)
func (g *FIXGateway) Connect(sessionID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	session, ok := g.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	log.Printf("[FIX] Connecting to %s at %s:%d", session.Name, session.Host, session.Port)
	session.Status = "CONNECTING"

	// Simulate connection delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		g.mu.Lock()
		session.Status = "CONNECTED"
		g.mu.Unlock()
		log.Printf("[FIX] Connected to %s", session.Name)

		// Simulate logon
		time.Sleep(200 * time.Millisecond)
		g.mu.Lock()
		session.Status = "LOGGED_IN"
		session.LastHeartbeat = time.Now()
		g.mu.Unlock()
		log.Printf("[FIX] Logged in to %s", session.Name)
	}()

	return nil
}

// SendOrder sends a new order to the LP (simulated)
func (g *FIXGateway) SendOrder(sessionID string, symbol string, side string, volume float64, price float64) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	lpOrderID := fmt.Sprintf("LP_%s_%d", sessionID, time.Now().UnixNano())
	log.Printf("[FIX] Sending order to %s: %s %s %.2f @ %.5f", session.Name, side, symbol, volume, price)

	// Simulate fill after delay
	go func() {
		time.Sleep(50 * time.Millisecond) // Simulated LP latency

		// 95% fill rate
		if time.Now().UnixNano()%100 < 95 {
			g.execReports <- ExecutionReport{
				OrderID:   lpOrderID,
				ExecType:  "FILLED",
				Symbol:    symbol,
				Side:      side,
				Volume:    volume,
				Price:     price + 0.00002, // Small slippage
				LPOrderID: lpOrderID,
				Timestamp: time.Now(),
			}
		} else {
			g.execReports <- ExecutionReport{
				OrderID:   lpOrderID,
				ExecType:  "REJECTED",
				Symbol:    symbol,
				Side:      side,
				Volume:    volume,
				Text:      "No liquidity available",
				LPOrderID: lpOrderID,
				Timestamp: time.Now(),
			}
		}
	}()

	return lpOrderID, nil
}

// GetStatus returns all session statuses
func (g *FIXGateway) GetStatus() map[string]string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	status := make(map[string]string)
	for id, session := range g.sessions {
		status[id] = session.Status
	}
	return status
}

// GetExecutionReports returns the channel for execution reports
func (g *FIXGateway) GetExecutionReports() <-chan ExecutionReport {
	return g.execReports
}
