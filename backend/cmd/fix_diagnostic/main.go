package main

import (
	"fmt"
	"log"
	"time"

	"github.com/epic1st/rtx/backend/fix"
)

// Comprehensive YOFX Market Data Diagnostic Tool
// Tests multiple configurations to find what works

func main() {
	log.Println("╔══════════════════════════════════════════════════════════╗")
	log.Println("║   YOFX Market Data Diagnostic - Testing All Variations   ║")
	log.Println("╚══════════════════════════════════════════════════════════╝")

	// Test configurations
	testCases := []struct {
		name         string
		symbol       string
		marketDepth  int
		subType      int
		updateType   int
		description  string
	}{
		{"Standard EURUSD", "EURUSD", 0, 1, 0, "Full book, Snapshot+Updates"},
		{"Depth=1 EURUSD", "EURUSD", 1, 1, 0, "Top of book, Snapshot+Updates"},
		{"Snapshot Only", "EURUSD", 0, 0, 0, "Full book, Snapshot only"},
		{"EUR/USD Slash", "EUR/USD", 0, 1, 0, "With slash separator"},
		{"GBPUSD Test", "GBPUSD", 0, 1, 0, "Different major pair"},
		{"USDJPY Test", "USDJPY", 0, 1, 0, "JPY cross"},
	}

	gateway := fix.NewFIXGateway()

	// Connect to YOFX2
	log.Println("\n[1/4] Connecting to YOFX2...")
	err := gateway.Connect("YOFX2")
	if err != nil {
		log.Fatalf("❌ Connection failed: %v", err)
	}

	log.Println("⏳ Waiting for logon (10s)...")
	time.Sleep(10 * time.Second)

	status := gateway.GetDetailedStatus()
	if status["YOFX2"].Status != "LOGGED_IN" {
		log.Fatalf("❌ Not logged in. Status: %s", status["YOFX2"].Status)
	}
	log.Println("✅ YOFX2 logged in successfully")

	// Test each configuration
	log.Println("\n[2/4] Testing market data subscriptions...")
	log.Println("     Each test will wait 15 seconds for response")
	log.Println()

	results := make(map[string]string)

	for i, tc := range testCases {
		log.Printf("═══ Test %d/%d: %s ═══", i+1, len(testCases), tc.name)
		log.Printf("    Symbol: %s", tc.symbol)
		log.Printf("    MarketDepth: %d (264=%d)", tc.marketDepth, tc.marketDepth)
		log.Printf("    SubscriptionType: %d (263=%d)", tc.subType, tc.subType)
		log.Printf("    UpdateType: %d (265=%d)", tc.updateType, tc.updateType)

		// Subscribe (we'll need to modify SubscribeMarketData to accept these params)
		mdReqID, err := gateway.SubscribeMarketData("YOFX2", tc.symbol)
		if err != nil {
			log.Printf("    ❌ Subscription error: %v\n", err)
			results[tc.name] = "ERROR: " + err.Error()
			continue
		}

		log.Printf("    📤 Sent MDReqID: %s", mdReqID)
		log.Printf("    ⏳ Waiting 15 seconds for response...")

		// Monitor for response
		timeout := time.After(15 * time.Second)
		ticker := time.NewTicker(1 * time.Second)
		gotResponse := false

		responseLoop:
		for elapsed := 0; elapsed < 15; elapsed++ {
			select {
			case md := <-gateway.GetMarketData():
				if md.Symbol == tc.symbol {
					log.Printf("    ✅ SUCCESS! MarketDataSnapshot received!")
					log.Printf("       Bid: %.5f | Ask: %.5f | Spread: %.5f", md.Bid, md.Ask, md.Ask-md.Bid)
					log.Printf("       Session: %s | Time: %s", md.SessionID, md.Timestamp.Format("15:04:05"))
					results[tc.name] = fmt.Sprintf("SUCCESS - Bid:%.5f Ask:%.5f", md.Bid, md.Ask)
					gotResponse = true
					break responseLoop
				}

			case reject := <-gateway.GetMarketDataRejects():
				if reject.MDReqID == mdReqID {
					log.Printf("    ❌ REJECTED!")
					log.Printf("       Reason: %s", reject.Reason)
					log.Printf("       Text: %s", reject.Text)
					results[tc.name] = fmt.Sprintf("REJECTED: %s", reject.Text)
					gotResponse = true
					break responseLoop
				}

			case <-timeout:
				break responseLoop

			case <-ticker.C:
				// Continue waiting
			}
		}

		if !gotResponse {
			log.Printf("    ⏱️  TIMEOUT - No response after 15 seconds")
			results[tc.name] = "TIMEOUT - No response"
		}

		log.Println()
		ticker.Stop()

		// Small delay between tests
		time.Sleep(2 * time.Second)
	}

	// Summary
	log.Println("╔══════════════════════════════════════════════════════════╗")
	log.Println("║                    Test Results Summary                  ║")
	log.Println("╚══════════════════════════════════════════════════════════╝")
	log.Println()

	successCount := 0
	for _, tc := range testCases {
		result := results[tc.name]
		status := "❌"
		if len(result) >= 7 && result[:7] == "SUCCESS" {
			status = "✅"
			successCount++
		} else if len(result) >= 8 && result[:8] == "REJECTED" {
			status = "⚠️ "
		} else if len(result) >= 7 && result[:7] == "TIMEOUT" {
			status = "⏱️ "
		}

		log.Printf("%-25s %s %s", tc.name, status, result)
	}

	log.Println()
	if successCount > 0 {
		log.Printf("🎉 Found %d working configuration(s)!", successCount)
		log.Println("\nWorking configurations:")
		for _, tc := range testCases {
			result := results[tc.name]
			if len(result) >= 7 && result[:7] == "SUCCESS" {
				log.Printf("  • %s: Symbol=%s, MarketDepth=%d, SubType=%d",
					tc.name, tc.symbol, tc.marketDepth, tc.subType)
			}
		}
	} else {
		log.Println("⚠️  No working configurations found")
		log.Println("\nPossible issues:")
		log.Println("  1. Account doesn't have market data permissions")
		log.Println("  2. Market is closed (check trading hours)")
		log.Println("  3. YOFX requires different symbol format")
		log.Println("  4. Server-side configuration needed")
		log.Println("\nNext steps:")
		log.Println("  • Contact YOFX support for market data entitlements")
		log.Println("  • Request list of valid symbols and formats")
		log.Println("  • Verify account has market data subscription")
	}

	log.Println("\n⏸️  Keeping connection alive. Press Ctrl+C to exit")
	select {}
}
