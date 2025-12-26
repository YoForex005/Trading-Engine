package oanda

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// Practice environment
	StreamingURL = "https://stream-fxpractice.oanda.com"
	RestURL      = "https://api-fxpractice.oanda.com"
)

// Config holds OANDA API configuration
type Config struct {
	APIKey    string
	AccountID string
}

// Price represents a real-time price update from OANDA
type Price struct {
	Type        string    `json:"type"`
	Time        time.Time `json:"time"`
	Instrument  string    `json:"instrument"`
	Tradeable   bool      `json:"tradeable"`
	Bids        []PriceLevel `json:"bids"`
	Asks        []PriceLevel `json:"asks"`
}

type PriceLevel struct {
	Price     string `json:"price"`
	Liquidity int    `json:"liquidity"`
}

// Heartbeat from OANDA stream
type Heartbeat struct {
	Type string `json:"type"`
	Time string `json:"time"`
}

// Client handles OANDA API interactions
type Client struct {
	config     Config
	httpClient *http.Client
	pricesChan chan Price
	stopChan   chan struct{}
	mu         sync.RWMutex
	connected  bool
}

// NewClient creates a new OANDA client
func NewClient(apiKey string) *Client {
	return &Client{
		config: Config{
			APIKey: apiKey,
		},
		httpClient: &http.Client{
			Timeout: 0, // No timeout for streaming
		},
		pricesChan: make(chan Price, 100),
		stopChan:   make(chan struct{}),
	}
}

// GetAccounts fetches all accounts for this API key
func (c *Client) GetAccounts() ([]string, error) {
	req, err := http.NewRequest("GET", RestURL+"/v3/accounts", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OANDA API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Accounts []struct {
			ID string `json:"id"`
		} `json:"accounts"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var ids []string
	for _, acc := range result.Accounts {
		ids = append(ids, acc.ID)
	}

	if len(ids) > 0 {
		c.config.AccountID = ids[0]
	}

	return ids, nil
}

// StreamPrices connects to OANDA streaming API for real-time prices
func (c *Client) StreamPrices(instruments []string) error {
	instrumentList := strings.Join(instruments, ",")
	url := fmt.Sprintf("%s/v3/accounts/%s/pricing/stream?instruments=%s",
		StreamingURL, c.config.AccountID, instrumentList)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Accept-Datetime-Format", "RFC3339")

	log.Printf("[OANDA] Connecting to price stream for: %s", instrumentList)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("OANDA stream error: %s - %s", resp.Status, string(body))
	}

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	log.Println("[OANDA] Connected to price stream")

	go c.readStream(resp.Body)

	return nil
}

func (c *Client) readStream(body io.ReadCloser) {
	defer body.Close()
	defer func() {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
	}()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-c.stopChan:
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Check if it's a heartbeat or price
		var typeCheck struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(line), &typeCheck); err != nil {
			continue
		}

		if typeCheck.Type == "PRICE" {
			var price Price
			if err := json.Unmarshal([]byte(line), &price); err != nil {
				log.Printf("[OANDA] Error parsing price: %v", err)
				continue
			}
			
			select {
			case c.pricesChan <- price:
			default:
				// Channel full, skip this price
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[OANDA] Stream error: %v", err)
	}
}

// GetPricesChan returns the channel for receiving prices
func (c *Client) GetPricesChan() <-chan Price {
	return c.pricesChan
}

// IsConnected checks if the stream is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Stop closes the streaming connection
func (c *Client) Stop() {
	close(c.stopChan)
}

// GetAccountSummary fetches account balance and details
func (c *Client) GetAccountSummary() (*AccountSummary, error) {
	url := fmt.Sprintf("%s/v3/accounts/%s/summary", RestURL, c.config.AccountID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OANDA API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Account AccountSummary `json:"account"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Account, nil
}

// Instrument represents an OANDA instrument
type Instrument struct {
	Name                string `json:"name"`
	Type                string `json:"type"`
	DisplayName         string `json:"displayName"`
	PipLocation         int    `json:"pipLocation"`
	DisplayPrecision    int    `json:"displayPrecision"`
	TradeUnitsPrecision int    `json:"tradeUnitsPrecision"`
	MinimumTradeSize    string `json:"minimumTradeSize"`
	MaximumTrailingStopDistance string `json:"maximumTrailingStopDistance"`
	MinimumTrailingStopDistance string `json:"minimumTrailingStopDistance"`
	MaximumPositionSize string `json:"maximumPositionSize"`
	MaximumOrderUnits   string `json:"maximumOrderUnits"`
	MarginRate          string `json:"marginRate"`
}

// GetInstruments fetches all tradeable instruments
func (c *Client) GetInstruments() ([]Instrument, error) {
	url := fmt.Sprintf("%s/v3/accounts/%s/instruments", RestURL, c.config.AccountID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OANDA API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Instruments []Instrument `json:"instruments"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Instruments, nil
}

// AccountSummary represents OANDA account info
type AccountSummary struct {
	ID                   string `json:"id"`
	Currency             string `json:"currency"`
	Balance              string `json:"balance"`
	UnrealizedPL         string `json:"unrealizedPL"`
	NAV                  string `json:"NAV"`
	MarginUsed           string `json:"marginUsed"`
	MarginAvailable      string `json:"marginAvailable"`
	OpenTradeCount       int    `json:"openTradeCount"`
	OpenPositionCount    int    `json:"openPositionCount"`
	PendingOrderCount    int    `json:"pendingOrderCount"`
}

// PlaceMarketOrder places a market order on OANDA
func (c *Client) PlaceMarketOrder(instrument string, units int) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/v3/accounts/%s/orders", RestURL, c.config.AccountID)

	orderData := map[string]interface{}{
		"order": map[string]interface{}{
			"type":       "MARKET",
			"instrument": instrument,
			"units":      fmt.Sprintf("%d", units), // Positive for buy, negative for sell
			"timeInForce": "FOK",
		},
	}

	body, _ := json.Marshal(orderData)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("OANDA order error: %s - %s", resp.Status, string(respBody))
	}

	var result OrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// OrderResponse represents the response from placing an order
type OrderResponse struct {
	OrderCreateTransaction struct {
		ID         string `json:"id"`
		Instrument string `json:"instrument"`
		Units      string `json:"units"`
		Type       string `json:"type"`
	} `json:"orderCreateTransaction"`
	OrderFillTransaction struct {
		ID        string `json:"id"`
		Price     string `json:"price"`
		PL        string `json:"pl"`
		TradeOpened struct {
			TradeID string `json:"tradeID"`
			Units   string `json:"units"`
		} `json:"tradeOpened"`
	} `json:"orderFillTransaction"`
}

// GetOpenTrades fetches all open trades
func (c *Client) GetOpenTrades() ([]Trade, error) {
	url := fmt.Sprintf("%s/v3/accounts/%s/openTrades", RestURL, c.config.AccountID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Trades []Trade `json:"trades"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Trades, nil
}

// Trade represents an open trade
type Trade struct {
	ID           string `json:"id"`
	Instrument   string `json:"instrument"`
	Price        string `json:"price"`
	OpenTime     string `json:"openTime"`
	CurrentUnits string `json:"currentUnits"`
	UnrealizedPL string `json:"unrealizedPL"`
}

// CloseTrade closes a specific trade
func (c *Client) CloseTrade(tradeID string) error {
	url := fmt.Sprintf("%s/v3/accounts/%s/trades/%s/close", RestURL, c.config.AccountID, tradeID)

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OANDA close trade error: %s - %s", resp.Status, string(body))
	}

	return nil
}
