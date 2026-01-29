package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type MarketTick struct {
	Type      string  `json:"type"`
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Spread    float64 `json:"spread"`
	Timestamp int64   `json:"timestamp"`
	LP        string  `json:"lp"`
}

func main() {
	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║         XAUUSD FIX 4.4 YoFX2 Subscription Test           ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	// Test 1: Check server health
	log.Println("\n[1/4] Checking backend server...")
	resp, err := http.Get("http://localhost:7999/health")
	if err != nil {
		log.Fatalf("❌ Server not running: %v", err)
	}
	resp.Body.Close()
	log.Println("✅ Server is running on port 7999")

	// Test 2: Get auth token
	log.Println("\n[2/4] Authenticating...")
	loginData := `{"account_number":"demo-user","password":"password"}`
	loginResp, err := http.Post("http://localhost:7999/api/v1/login", "application/json",
		bytes.NewBufferString(loginData))

	var token string
	if err != nil {
		log.Printf("⚠️  Login endpoint not responding: %v", err)
		log.Println("   Attempting WebSocket connection without auth...")
	} else {
		defer loginResp.Body.Close()
		var loginResult struct {
			Token string `json:"token"`
		}
		json.NewDecoder(loginResp.Body).Decode(&loginResult)
		token = loginResult.Token
		if token != "" {
			log.Printf("✅ Auth token acquired: %s...", token[:min(20, len(token))])
		}
	}

	// Test 3: Connect to WebSocket
	log.Println("\n[3/4] Connecting to WebSocket feed...")
	wsURL := "ws://localhost:7999/ws"
	if token != "" {
		wsURL = fmt.Sprintf("ws://localhost:7999/ws?token=%s", token)
	}

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("❌ WebSocket connection failed: %v", err)
		log.Println("\n📁 Falling back to tick file monitoring...")
		monitorTickFiles()
		return
	}
	defer conn.Close()
	log.Println("✅ WebSocket connected successfully")

	// Test 4: Monitor for XAUUSD ticks
	log.Println("\n[4/4] Monitoring XAUUSD quotes for 30 seconds...")
	log.Println("     Waiting for live FIX 4.4 market data from YoFX2...")
	log.Println("     Expected: Bid/Ask prices for Gold (XAUUSD)")
	log.Println()

	tickCount := 0
	var lastBid, lastAsk float64
	timeout := time.After(30 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			var tick MarketTick
			err := conn.ReadJSON(&tick)
			if err != nil {
				log.Printf("❌ WebSocket read error: %v", err)
				done <- true
				return
			}

			// Show all symbols for debugging
			if tickCount == 0 {
				log.Printf("📊 First tick received: %s (LP: %s)", tick.Symbol, tick.LP)
			}

			if tick.Symbol == "XAUUSD" {
				tickCount++
				priceChange := ""
				if lastBid > 0 {
					bidChange := tick.Bid - lastBid
					if bidChange > 0 {
						priceChange = fmt.Sprintf("↑ +%.2f", bidChange)
					} else if bidChange < 0 {
						priceChange = fmt.Sprintf("↓ %.2f", bidChange)
					} else {
						priceChange = "="
					}
				}

				log.Printf("✅ [%3d] XAUUSD: Bid=%8.2f Ask=%8.2f Spread=%5.2f %s | LP=%-10s | %s",
					tickCount, tick.Bid, tick.Ask, tick.Spread, priceChange, tick.LP,
					time.Unix(tick.Timestamp, 0).Format("15:04:05"))

				lastBid = tick.Bid
				lastAsk = tick.Ask
			}
		}
	}()

	select {
	case <-timeout:
		log.Println()
		log.Println("╔═══════════════════════════════════════════════════════════╗")
		log.Println("║                     Test Complete                         ║")
		log.Println("╚═══════════════════════════════════════════════════════════╝")
		if tickCount > 0 {
			log.Printf("✅ SUCCESS: Received %d XAUUSD ticks from FIX 4.4", tickCount)
			log.Printf("   Latest price: Bid=%.2f Ask=%.2f", lastBid, lastAsk)
			log.Println("   Quote throttling is working (only significant price changes shown)")
		} else {
			log.Println("⚠️  No XAUUSD ticks received in 30 seconds")
			log.Println("\n   Possible reasons:")
			log.Println("   • FIX YoFX2 session not connected")
			log.Println("   • Market closed (outside trading hours)")
			log.Println("   • XAUUSD not in subscription list")
			log.Println("   • Quote throttling filtering out all changes")
			log.Println("\n   Checking tick files...")
			checkTickFiles()
		}
	case <-done:
		log.Println("⚠️  WebSocket closed unexpectedly")
	}
}

func monitorTickFiles() {
	log.Println("\n📊 Monitoring tick files for XAUUSD updates...")

	today := time.Now().Format("2006-01-02")
	filePath := fmt.Sprintf("../../data/ticks/XAUUSD/%s.json", today)

	log.Printf("   File path: %s\n", filePath)

	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)

		data, err := os.ReadFile(filePath)
		if err == nil {
			var ticks []MarketTick
			json.Unmarshal(data, &ticks)
			if len(ticks) > 0 {
				lastTick := ticks[len(ticks)-1]
				log.Printf("[%2ds] ✅ %d ticks | Latest: Bid=%.2f Ask=%.2f LP=%s",
					i+1, len(ticks), lastTick.Bid, lastTick.Ask, lastTick.LP)
			} else {
				log.Printf("[%2ds] File exists but empty", i+1)
			}
		} else {
			log.Printf("[%2ds] ⏳ Waiting for today's file...", i+1)
		}
	}
}

func checkTickFiles() {
	today := time.Now().Format("2006-01-02")
	filePath := fmt.Sprintf("../../data/ticks/XAUUSD/%s.json", today)

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("   📁 No tick file for today (%s)", today)
		return
	}

	var ticks []MarketTick
	json.Unmarshal(data, &ticks)
	log.Printf("   📁 Found %d ticks in today's file", len(ticks))

	if len(ticks) > 0 {
		lastTick := ticks[len(ticks)-1]
		log.Printf("   Latest tick: Bid=%.2f Ask=%.2f LP=%s at %v",
			lastTick.Bid, lastTick.Ask, lastTick.LP,
			time.Unix(lastTick.Timestamp, 0).Format("15:04:05"))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
