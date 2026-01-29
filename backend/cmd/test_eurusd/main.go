package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/epic1st/rtx/backend/fix"
)

func main() {
	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║      EURUSD Subscription via FIX 4.4 YoFX2 Session       ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	// Step 1: Create FIX Gateway
	log.Println("\n[1/5] Initializing FIX Gateway...")
	gateway := fix.NewFIXGateway()
	log.Println("✅ FIX Gateway initialized with YoFX2 session")

	// Step 2: Connect to YoFX2
	log.Println("\n[2/5] Connecting to YoFX2 session...")
	err := gateway.Connect("YOFX2")
	if err != nil {
		log.Fatalf("❌ Failed to connect YoFX2: %v", err)
	}

	log.Println("⏳ Waiting for FIX logon to complete (10 seconds)...")
	time.Sleep(10 * time.Second)

	// Step 3: Check connection status
	status := gateway.GetDetailedStatus()
	yofx2Status, ok := status["YOFX2"]
	if !ok {
		log.Fatal("❌ YoFX2 session not found in status")
	}

	log.Printf("\n[3/5] YoFX2 Session Status:")
	log.Printf("   Status: %s", yofx2Status.Status)
	log.Printf("   OutSeqNum: %d", yofx2Status.OutSeqNum)
	log.Printf("   InSeqNum: %d", yofx2Status.InSeqNum)
	log.Printf("   LastHeartbeat: %v", yofx2Status.LastHeartbeat.Format("15:04:05"))

	if yofx2Status.Status != "LOGGED_IN" {
		log.Fatal("❌ YoFX2 not logged in. Status: " + yofx2Status.Status)
	}
	log.Println("✅ YoFX2 logged in successfully!")

	// Step 4: Subscribe to EURUSD
	log.Println("\n[4/5] Subscribing to EURUSD market data...")
	mdReqID, err := gateway.SubscribeMarketData("YOFX2", "EURUSD")
	if err != nil {
		log.Fatalf("❌ Failed to subscribe EURUSD: %v", err)
	}
	log.Printf("✅ Subscription request sent! MDReqID: %s", mdReqID)

	// Step 5: Monitor market data feed
	log.Println("\n[5/5] Monitoring EURUSD market data (60 seconds)...")
	log.Println("     Waiting for MarketDataSnapshot (35=W) messages...")
	log.Println()

	tickCount := 0
	rejectCount := 0
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Start goroutine to monitor market data
	go func() {
		for {
			select {
			case md := <-gateway.GetMarketData():
				if md.Symbol == "EURUSD" {
					tickCount++
					log.Printf("✅ [%3d] EURUSD Market Data:", tickCount)
					log.Printf("       Bid: %.2f | Ask: %.2f | Spread: %.2f", md.Bid, md.Ask, md.Ask-md.Bid)
					log.Printf("       Session: %s | Time: %s", md.SessionID, md.Timestamp.Format("15:04:05"))
					log.Println()
				}
			case reject := <-gateway.GetMarketDataRejects():
				if reject.MDReqID == mdReqID {
					rejectCount++
					log.Printf("⚠️  [REJECT #%d] MDReqID: %s", rejectCount, reject.MDReqID)
					log.Printf("       Reason: %s", reject.Reason)
					log.Printf("       Text: %s", reject.Text)
					log.Println()
				}
			case <-ticker.C:
				// Heartbeat - show we're still waiting
				if tickCount == 0 {
					log.Printf("⏳ Waiting for EURUSD data... (Ticks: %d, Rejects: %d)", tickCount, rejectCount)
				}
			}
		}
	}()

	// Wait for timeout
	<-timeout

	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║                   Monitoring Complete                     ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	if tickCount > 0 {
		log.Printf("✅ SUCCESS: Received %d EURUSD ticks from FIX 4.4 YoFX2", tickCount)
		log.Println()
		log.Println("📊 Next Steps:")
		log.Println("   • EURUSD data is now flowing through the backend")
		log.Println("   • Check backend/data/ticks/EURUSD/ for persisted ticks")
		log.Println("   • Connect to ws://localhost:7999/ws to stream live quotes")
	} else {
		log.Printf("⚠️  No EURUSD ticks received (Rejects: %d)", rejectCount)
		log.Println()
		log.Println("Possible reasons:")
		log.Println("  • Symbol not available on YoFX server")
		log.Println("  • Market closed (check trading hours)")
		log.Println("  • Subscription rejected by server")
		log.Println("  • Network/proxy issues")
		log.Println()
		log.Println("Checking tick files...")
		checkTickFiles()
	}

	// Keep gateway alive
	log.Println("\n⏸️  Press Ctrl+C to exit (gateway will remain connected)")
	select {}
}

func checkTickFiles() {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:7999/api/v1/ticks/EURUSD?limit=5")
	if err != nil {
		log.Printf("   📁 Cannot check backend API: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("   📁 Backend API returned: %d", resp.StatusCode)
		return
	}

	var ticks []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&ticks)

	if len(ticks) > 0 {
		log.Printf("   📁 Found %d recent EURUSD ticks in backend storage", len(ticks))
		lastTick := ticks[len(ticks)-1]
		log.Printf("       Latest: Bid=%.2f Ask=%.2f LP=%s",
			lastTick["bid"], lastTick["ask"], lastTick["lp"])
	} else {
		log.Println("   📁 No EURUSD ticks found in backend storage")
	}
}
