// +build ignore

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const SOH = "\x01"

type SymbolTest struct {
	Symbol     string
	Subscribed bool
	TickCount  int
	LastPrice  string
	Status     string
}

func calculateChecksum(msg string) int {
	sum := 0
	for i := 0; i < len(msg); i++ {
		sum += int(msg[i])
	}
	return sum % 256
}

func buildLogon(senderCompID, targetCompID, username, password string, seqNum int) string {
	timestamp := time.Now().UTC().Format("20060102-15:04:05.000")
	body := fmt.Sprintf("35=A%s49=%s%s56=%s%s34=%d%s52=%s%s98=0%s108=30%s553=%s%s554=%s%s",
		SOH, senderCompID, SOH, targetCompID, SOH, seqNum, SOH, timestamp, SOH, SOH, SOH, username, SOH, password, SOH)
	bodyLen := len(body)
	fullMsg := fmt.Sprintf("8=FIX.4.4%s9=%d%s%s", SOH, bodyLen, SOH, body)
	checksum := calculateChecksum(fullMsg)
	fullMsg += fmt.Sprintf("10=%03d%s", checksum, SOH)
	return fullMsg
}

func buildMarketDataRequest(symbol string, seqNum int) string {
	timestamp := time.Now().UTC().Format("20060102-15:04:05.000")
	reqID := fmt.Sprintf("MD-%s-%d", symbol, time.Now().Unix())

	// Market Data Request (V) with proper symbol group
	body := fmt.Sprintf("35=V%s49=YOFX2%s56=YOFX%s34=%d%s52=%s%s"+
		"262=%s%s"+           // MDReqID
		"263=1%s"+            // SubscriptionRequestType: 1=Snapshot+Updates
		"264=0%s"+            // MarketDepth: 0=Full Book
		"265=0%s"+            // MDUpdateType: 0=Full Refresh
		"146=1%s"+            // NoRelatedSym group count
		"55=%s%s",            // Symbol
		SOH, SOH, SOH, seqNum, SOH, timestamp, SOH,
		reqID, SOH,
		SOH,
		SOH,
		SOH,
		SOH,
		symbol, SOH)

	bodyLen := len(body)
	fullMsg := fmt.Sprintf("8=FIX.4.4%s9=%d%s%s", SOH, bodyLen, SOH, body)
	checksum := calculateChecksum(fullMsg)
	fullMsg += fmt.Sprintf("10=%03d%s", checksum, SOH)
	return fullMsg
}

func buildHeartbeat(seqNum int) string {
	timestamp := time.Now().UTC().Format("20060102-15:04:05.000")
	body := fmt.Sprintf("35=0%s49=YOFX2%s56=YOFX%s34=%d%s52=%s%s",
		SOH, SOH, SOH, seqNum, SOH, timestamp, SOH)
	bodyLen := len(body)
	fullMsg := fmt.Sprintf("8=FIX.4.4%s9=%d%s%s", SOH, bodyLen, SOH, body)
	checksum := calculateChecksum(fullMsg)
	fullMsg += fmt.Sprintf("10=%03d%s", checksum, SOH)
	return fullMsg
}

func connectViaProxy() (net.Conn, error) {
	proxyAddr := "81.29.145.69:49527"
	proxyUser := "fGUqTcsdMsBZlms"
	proxyPass := "3eo1qF91WA7Fyku"
	targetHost := "23.106.238.138"
	targetPort := 12336

	conn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(proxyUser + ":" + proxyPass))
	connectReq := fmt.Sprintf("CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\nProxy-Authorization: Basic %s\r\n\r\n",
		targetHost, targetPort, targetHost, targetPort, auth)

	conn.Write([]byte(connectReq))

	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')

	if !strings.Contains(response, "200") {
		conn.Close()
		return nil, fmt.Errorf("proxy error: %s", response)
	}

	for {
		line, _ := reader.ReadString('\n')
		if line == "\r\n" || line == "\n" {
			break
		}
	}

	return conn, nil
}

func extractField(msg, tag string) string {
	parts := strings.Split(msg, SOH)
	for _, part := range parts {
		if strings.HasPrefix(part, tag+"=") {
			return strings.TrimPrefix(part, tag+"=")
		}
	}
	return ""
}

func testSymbol(symbol string, results chan<- SymbolTest) {
	result := SymbolTest{
		Symbol: symbol,
		Status: "TESTING",
	}

	conn, err := connectViaProxy()
	if err != nil {
		result.Status = "CONN_FAILED"
		results <- result
		return
	}
	defer conn.Close()

	seqNum := 1

	// Logon
	logon := buildLogon("YOFX2", "YOFX", "YOFX2", "Brand#143", seqNum)
	conn.Write([]byte(logon))
	seqNum++

	// Read logon response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || !strings.Contains(string(buf[:n]), "35=A") {
		result.Status = "LOGON_FAILED"
		results <- result
		return
	}

	// Subscribe to symbol
	mdReq := buildMarketDataRequest(symbol, seqNum)
	conn.Write([]byte(mdReq))
	seqNum++

	// Monitor for 20 seconds
	timeout := time.After(20 * time.Second)
	heartbeatTicker := time.NewTicker(10 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-timeout:
			if result.TickCount > 0 {
				result.Status = "ACTIVE"
			} else if result.Subscribed {
				result.Status = "NO_DATA"
			} else {
				result.Status = "REJECTED"
			}
			results <- result
			return

		case <-heartbeatTicker.C:
			hb := buildHeartbeat(seqNum)
			conn.Write([]byte(hb))
			seqNum++

		default:
			conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, err := conn.Read(buf)
			if err != nil {
				continue
			}

			response := string(buf[:n])
			messages := strings.Split(response, "8=FIX")

			for _, msg := range messages {
				if len(msg) < 10 {
					continue
				}

				msgType := extractField(msg, "35")

				switch msgType {
				case "W": // Market Data Snapshot
					result.Subscribed = true
					result.TickCount++
					price := extractField(msg, "270")
					if price != "" {
						result.LastPrice = price
					}

				case "X": // Market Data Incremental Refresh
					result.Subscribed = true
					result.TickCount++
					price := extractField(msg, "270")
					if price != "" {
						result.LastPrice = price
					}

				case "Y": // Market Data Request Reject
					rejectReason := extractField(msg, "281")
					result.Status = "REJECT: " + rejectReason
					results <- result
					return

				case "0": // Heartbeat
					// Respond with heartbeat
					hb := buildHeartbeat(seqNum)
					conn.Write([]byte(hb))
					seqNum++
				}
			}
		}
	}
}

func main() {
	symbols := []string{
		"EURUSD",
		"GBPUSD",
		"USDJPY",
		"AUDUSD",
		"USDCHF",
		"USDCAD",
		"NZDUSD",
		"EURGBP",
		"EURJPY",
		"GBPJPY",
		"XAUUSD",
		"XAGUSD",
		"BTCUSD",
		"ETHUSD",
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              YoFX SYMBOL DISCOVERY TEST                                  ║")
	fmt.Println("║              Testing 14 Common Symbols                                   ║")
	fmt.Println("║              20-Second Monitoring Per Symbol                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝\n")

	results := make(chan SymbolTest, len(symbols))
	var wg sync.WaitGroup

	// Test symbols sequentially to avoid overwhelming the server
	for _, symbol := range symbols {
		fmt.Printf("Testing %s...\n", symbol)
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			testSymbol(sym, results)
		}(symbol)
		wg.Wait() // Wait for this one to complete before starting next
		time.Sleep(2 * time.Second) // Cool-down between tests
	}

	close(results)

	// Collect and display results
	var activeSymbols []SymbolTest
	var noDataSymbols []SymbolTest
	var rejectedSymbols []SymbolTest
	var failedSymbols []SymbolTest

	for result := range results {
		switch {
		case result.Status == "ACTIVE":
			activeSymbols = append(activeSymbols, result)
		case result.Status == "NO_DATA":
			noDataSymbols = append(noDataSymbols, result)
		case strings.HasPrefix(result.Status, "REJECT"):
			rejectedSymbols = append(rejectedSymbols, result)
		default:
			failedSymbols = append(failedSymbols, result)
		}
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                           RESULTS SUMMARY                                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝")

	fmt.Println("\n✅ ACTIVE SYMBOLS (Subscription Accepted + Data Flowing):")
	if len(activeSymbols) == 0 {
		fmt.Println("   None")
	}
	for _, s := range activeSymbols {
		fmt.Printf("   %-10s | Ticks: %-5d | Last Price: %s\n", s.Symbol, s.TickCount, s.LastPrice)
	}

	fmt.Println("\n⚠️  ACCEPTED BUT NO DATA (Subscription OK, No Ticks):")
	if len(noDataSymbols) == 0 {
		fmt.Println("   None")
	}
	for _, s := range noDataSymbols {
		fmt.Printf("   %-10s | Status: %s\n", s.Symbol, s.Status)
	}

	fmt.Println("\n❌ REJECTED SYMBOLS:")
	if len(rejectedSymbols) == 0 {
		fmt.Println("   None")
	}
	for _, s := range rejectedSymbols {
		fmt.Printf("   %-10s | Reason: %s\n", s.Symbol, s.Status)
	}

	fmt.Println("\n🔴 CONNECTION/LOGON FAILED:")
	if len(failedSymbols) == 0 {
		fmt.Println("   None")
	}
	for _, s := range failedSymbols {
		fmt.Printf("   %-10s | Status: %s\n", s.Symbol, s.Status)
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  Total Tested:     %-2d                                                    ║\n", len(symbols))
	fmt.Printf("║  Active (Data):    %-2d  ✅                                                ║\n", len(activeSymbols))
	fmt.Printf("║  No Data:          %-2d  ⚠️                                                 ║\n", len(noDataSymbols))
	fmt.Printf("║  Rejected:         %-2d  ❌                                                ║\n", len(rejectedSymbols))
	fmt.Printf("║  Failed:           %-2d  🔴                                                ║\n", len(failedSymbols))
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝")

	if len(activeSymbols) > 0 {
		fmt.Println("\n📊 RECOMMENDATION:")
		fmt.Println("   Use these active symbols for your trading engine:")
		for _, s := range activeSymbols {
			fmt.Printf("   - %s\n", s.Symbol)
		}
	} else {
		fmt.Println("\n⚠️  WARNING: No symbols are delivering live data!")
		fmt.Println("   Check server configuration or contact YoFX support.")
	}
}
