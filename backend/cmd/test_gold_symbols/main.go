package main

import (
	"log"
	"time"

	"github.com/epic1st/rtx/backend/fix"
)

func main() {
	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║        Testing Gold Symbol Variations on YoFX2           ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	// Common gold symbol formats
	goldSymbols := []string{
		"XAUUSD",     // Standard format
		"XAU/USD",    // With slash
		"XAUUSD.a",   // With suffix
		"GOLD",       // Simple name
		"GOLD/USD",   // Simple with slash
		"GLD",        // Short form
		"XAU",        // Metal code only
		"XAUUSD-C",   // Contract suffix
		"XAUUSD_C",   // Underscore suffix
		"XAUUSD-FUT", // Futures suffix
	}

	// Initialize gateway
	log.Println("\n[1/3] Initializing FIX Gateway...")
	gateway := fix.NewFIXGateway()

	// Connect to YoFX2
	log.Println("\n[2/3] Connecting to YoFX2...")
	err := gateway.Connect("YOFX2")
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	log.Println("⏳ Waiting for login (10 seconds)...")
	time.Sleep(10 * time.Second)

	status := gateway.GetDetailedStatus()
	if status["YOFX2"].Status != "LOGGED_IN" {
		log.Fatal("❌ Not logged in")
	}
	log.Println("✅ YoFX2 logged in")

	// Test each symbol
	log.Println("\n[3/3] Testing gold symbol variations...")
	log.Println("     Sending MarketDataRequest for each variation...")
	log.Println()

	results := make(map[string]string)

	for _, symbol := range goldSymbols {
		log.Printf("📊 Testing: %-15s", symbol)

		mdReqID, err := gateway.SubscribeMarketData("YOFX2", symbol)
		if err != nil {
			log.Printf("   ❌ Subscription error: %v\n", err)
			results[symbol] = "ERROR: " + err.Error()
			continue
		}

		log.Printf("   ⏳ Sent MDReqID: %s", mdReqID)

		// Wait for response
		timeout := time.After(3 * time.Second)
		received := false

		select {
		case md := <-gateway.GetMarketData():
			if md.Symbol == symbol {
				log.Printf("   ✅ SUCCESS! Bid: %.2f Ask: %.2f\n", md.Bid, md.Ask)
				results[symbol] = "SUCCESS"
				received = true
			}
		case reject := <-gateway.GetMarketDataRejects():
			if reject.MDReqID == mdReqID {
				log.Printf("   ❌ REJECTED: %s\n", reject.Text)
				results[symbol] = "REJECTED: " + reject.Text
				received = true
			}
		case <-timeout:
			if !received {
				log.Println("   ⏱️  TIMEOUT (no response)\n")
				results[symbol] = "TIMEOUT"
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Summary
	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║                    Test Results Summary                   ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	successCount := 0
	for _, symbol := range goldSymbols {
		result := results[symbol]
		status := "❌"
		if result == "SUCCESS" {
			status = "✅"
			successCount++
		}
		log.Printf("%-15s %s %s", symbol, status, result)
	}

	log.Println()
	if successCount > 0 {
		log.Printf("🎉 Found %d working symbol(s)!", successCount)
		log.Println("\nTo subscribe to gold:")
		for _, symbol := range goldSymbols {
			if results[symbol] == "SUCCESS" {
				log.Printf("   gateway.SubscribeMarketData(\"YOFX2\", \"%s\")", symbol)
			}
		}
	} else {
		log.Println("⚠️  No gold symbols found on YoFX server")
		log.Println("\nPossible reasons:")
		log.Println("  • Gold not available on this account")
		log.Println("  • Different symbol naming convention")
		log.Println("  • Market closed")
		log.Println("\nTry these EUR/USD pairs instead:")
		log.Println("   EURUSD, EUR/USD, EURUSD.a")
	}
}
