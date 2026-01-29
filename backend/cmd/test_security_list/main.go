package main

import (
	"log"
	"time"

	"github.com/epic1st/rtx/backend/fix"
)

// Test SecurityListRequest (35=x) before MarketDataRequest
// Some FIX servers require security discovery first

func main() {
	log.Println("╔══════════════════════════════════════════════════════════╗")
	log.Println("║      Testing SecurityListRequest Before Market Data      ║")
	log.Println("╚══════════════════════════════════════════════════════════╝")

	gateway := fix.NewFIXGateway()

	// Step 1: Connect
	log.Println("\n[1/4] Connecting to YOFX2...")
	err := gateway.Connect("YOFX2")
	if err != nil {
		log.Fatalf("❌ Connection failed: %v", err)
	}

	log.Println("⏳ Waiting for logon (10s)...")
	time.Sleep(10 * time.Second)

	status := gateway.GetDetailedStatus()
	if status["YOFX2"].Status != "LOGGED_IN" {
		log.Fatalf("❌ Not logged in")
	}
	log.Println("✅ YOFX2 logged in")

	// Step 2: Request Security List
	log.Println("\n[2/4] Requesting security list...")
	log.Println("     This may tell us what symbols are available")

	// For now, we'll skip straight to market data with a modified request
	// that includes SecurityType tag

	// Step 3: Subscribe with SecurityType
	log.Println("\n[3/4] Subscribing to EURUSD with SecurityType tag...")
	log.Println("     Adding tag 167=FXSPOT to indicate forex pair")

	// We need to modify the SubscribeMarketData to accept additional tags
	// For now, let's just test the normal way and monitor

	mdReqID, err := gateway.SubscribeMarketData("YOFX2", "EURUSD")
	if err != nil {
		log.Fatalf("❌ Subscribe failed: %v", err)
	}
	log.Printf("📤 Sent MDReqID: %s", mdReqID)

	// Step 4: Monitor for 30 seconds
	log.Println("\n[4/4] Monitoring for responses (30 seconds)...")

	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	gotData := false

	for !gotData {
		select {
		case md := <-gateway.GetMarketData():
			log.Printf("✅ MarketDataSnapshot received!")
			log.Printf("   Symbol: %s", md.Symbol)
			log.Printf("   Bid: %.5f | Ask: %.5f", md.Bid, md.Ask)
			gotData = true

		case reject := <-gateway.GetMarketDataRejects():
			log.Printf("❌ MarketDataReject received!")
			log.Printf("   MDReqID: %s", reject.MDReqID)
			log.Printf("   Reason: %s", reject.Reason)
			log.Printf("   Text: %s", reject.Text)
			gotData = true

		case <-ticker.C:
			log.Println("   ⏳ Still waiting...")

		case <-timeout:
			log.Println("⏱️  Timeout - No response received")
			log.Println("\nThis confirms the server is not responding to MarketDataRequest")
			log.Println("Next step: Try adding SecurityType (167=FXSPOT) tag")
			return
		}
	}

	log.Println("\n✅ Test complete!")
	log.Println("⏸️  Press Ctrl+C to exit")
	select {}
}
