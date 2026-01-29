// +build ignore
//
// YoFX Market Hours Symbol Test
// Run this during active market hours:
// - Sunday 22:00 UTC onwards (market open)
// - Monday-Friday 00:00-22:00 UTC (active trading)
//
// This will test the 9 validated symbols and confirm live data flow

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"
)

const SOH = "\x01"

type SymbolResult struct {
	Symbol    string
	Ticks     int
	LastBid   string
	LastOffer string
	FirstTick time.Time
	LastTick  time.Time
	Active    bool
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

	body := fmt.Sprintf("35=V%s49=YOFX2%s56=YOFX%s34=%d%s52=%s%s"+
		"262=%s%s"+
		"263=1%s"+
		"264=0%s"+
		"267=2%s"+
		"269=0%s"+
		"269=1%s"+
		"146=1%s"+
		"55=%s%s",
		SOH, SOH, SOH, seqNum, SOH, timestamp, SOH,
		reqID, SOH,
		SOH,
		SOH,
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

func main() {
	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY",
		"AUDUSD", "USDCHF", "USDCAD",
		"EURGBP", "EURJPY", "GBPJPY",
	}

	nowUTC := time.Now().UTC()
	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              YoFX MARKET HOURS LIVE DATA TEST                            ║")
	fmt.Printf("║              Current Time: %-46s ║\n", nowUTC.Format("2006-01-02 15:04:05 UTC"))
	fmt.Println("║              Testing 9 Validated Symbols                                 ║")
	fmt.Println("║              Duration: 60 seconds per symbol                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝\n")

	// Check if likely during market hours
	hour := nowUTC.Hour()
	day := nowUTC.Weekday()

	if day == time.Saturday || (day == time.Sunday && hour < 22) {
		fmt.Println("⚠️  WARNING: Current time is likely OUTSIDE market hours")
		fmt.Println("   Forex market opens: Sunday 22:00 UTC")
		fmt.Println("   Forex market closes: Friday 22:00 UTC")
		fmt.Println("   You may not receive live data.\n")
	} else {
		fmt.Println("✅ Current time is likely DURING market hours")
		fmt.Println("   Expecting live tick data...\n")
	}

	results := make(map[string]*SymbolResult)
	for _, symbol := range symbols {
		results[symbol] = &SymbolResult{Symbol: symbol}
	}

	conn, err := connectViaProxy()
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("✅ Connected via proxy")

	seqNum := 1

	// Logon
	logon := buildLogon("YOFX2", "YOFX", "YOFX2", "Brand#143", seqNum)
	conn.Write([]byte(logon))
	seqNum++

	// Read logon response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("❌ Logon timeout: %v\n", err)
		return
	}

	response := string(buf[:n])
	if !strings.Contains(response, "35=A") {
		fmt.Printf("❌ Logon failed\n")
		return
	}

	fmt.Println("✅ Logged in")
	fmt.Println("\n📤 Subscribing to all 9 symbols...\n")

	// Subscribe to all symbols
	for _, symbol := range symbols {
		mdReq := buildMarketDataRequest(symbol, seqNum)
		conn.Write([]byte(mdReq))
		seqNum++
		fmt.Printf("   ✓ Subscribed to %s\n", symbol)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n⏳ Monitoring for 60 seconds...\n")

	// Monitor for 60 seconds
	timeout := time.After(60 * time.Second)
	heartbeatTicker := time.NewTicker(10 * time.Second)
	defer heartbeatTicker.Stop()

	lastUpdate := time.Now()

	for {
		select {
		case <-timeout:
			goto RESULTS

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
				case "W", "X":
					// Market data!
					symbol := extractField(msg, "55")
					price := extractField(msg, "270")
					entryType := extractField(msg, "269")

					if result, ok := results[symbol]; ok {
						result.Ticks++
						if result.FirstTick.IsZero() {
							result.FirstTick = time.Now()
						}
						result.LastTick = time.Now()
						result.Active = true

						if entryType == "0" {
							result.LastBid = price
						} else if entryType == "1" {
							result.LastOffer = price
						}

						// Print live update every second
						if time.Since(lastUpdate) > time.Second {
							fmt.Printf("📊 %s: %d ticks | Bid: %s | Offer: %s\n",
								symbol, result.Ticks, result.LastBid, result.LastOffer)
							lastUpdate = time.Now()
						}
					}

				case "Y":
					// Rejection
					symbol := extractField(msg, "55")
					rejectText := extractField(msg, "58")
					if result, ok := results[symbol]; ok {
						fmt.Printf("❌ %s REJECTED: %s\n", symbol, rejectText)
						result.Active = false
					}

				case "0":
					hb := buildHeartbeat(seqNum)
					conn.Write([]byte(hb))
					seqNum++

				case "1":
					hb := buildHeartbeat(seqNum)
					conn.Write([]byte(hb))
					seqNum++
				}
			}
		}
	}

RESULTS:
	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                           RESULTS SUMMARY                                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝\n")

	activeCount := 0
	for _, symbol := range symbols {
		result := results[symbol]
		if result.Active && result.Ticks > 0 {
			activeCount++
			duration := result.LastTick.Sub(result.FirstTick).Seconds()
			ticksPerSec := float64(result.Ticks) / duration
			fmt.Printf("✅ %-8s | Ticks: %-6d | Bid: %-10s | Offer: %-10s | Rate: %.1f/sec\n",
				result.Symbol, result.Ticks, result.LastBid, result.LastOffer, ticksPerSec)
		} else if result.Ticks > 0 {
			fmt.Printf("⚠️  %-8s | Ticks: %-6d | Bid: %-10s | Offer: %-10s | (Stopped)\n",
				result.Symbol, result.Ticks, result.LastBid, result.LastOffer)
		} else {
			fmt.Printf("❌ %-8s | NO DATA\n", result.Symbol)
		}
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  Active Symbols: %d / %d                                                   ║\n", activeCount, len(symbols))
	if activeCount > 0 {
		fmt.Println("║  STATUS: ✅ LIVE DATA CONFIRMED - YoFX is working!                       ║")
	} else {
		fmt.Println("║  STATUS: ❌ NO LIVE DATA - Retry during market hours or check account    ║")
	}
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝")
}
