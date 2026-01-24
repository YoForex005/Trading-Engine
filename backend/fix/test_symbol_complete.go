// +build ignore

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

	// Market Data Request (V) with ALL required fields
	// Tag 267 (NoMDEntryTypes) was missing - this specifies what type of data we want
	body := fmt.Sprintf("35=V%s49=YOFX2%s56=YOFX%s34=%d%s52=%s%s"+
		"262=%s%s"+           // MDReqID - unique request ID
		"263=1%s"+            // SubscriptionRequestType: 1=Snapshot+Updates
		"264=0%s"+            // MarketDepth: 0=Full Book
		"267=2%s"+            // NoMDEntryTypes: 2 entry types (BID and OFFER)
		"269=0%s"+            // MDEntryType: 0=Bid
		"269=1%s"+            // MDEntryType: 1=Offer/Ask
		"146=1%s"+            // NoRelatedSym: 1 symbol
		"55=%s%s",            // Symbol
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

func testSymbol(symbol string) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Testing Symbol: %-42s  ║\n", symbol)
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n")

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
	buf := make([]byte, 4096)
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

	// Subscribe to symbol
	mdReq := buildMarketDataRequest(symbol, seqNum)
	debugReq := strings.ReplaceAll(mdReq, SOH, "|")
	fmt.Printf("📤 Market Data Request:\n   %s\n", debugReq)
	conn.Write([]byte(mdReq))
	seqNum++

	// Monitor for responses
	timeout := time.After(20 * time.Second)
	heartbeatTicker := time.NewTicker(10 * time.Second)
	defer heartbeatTicker.Stop()

	tickCount := 0
	lastPrice := ""

	for {
		select {
		case <-timeout:
			if tickCount > 0 {
				fmt.Printf("✅ ACTIVE - Received %d tick(s), Last Price: %s\n", tickCount, lastPrice)
			} else {
				fmt.Printf("⚠️  NO DATA - Subscription accepted but no ticks received\n")
			}
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
				debugMsg := strings.ReplaceAll(msg, SOH, "|")

				switch msgType {
				case "W":
					tickCount++
					price := extractField(msg, "270")
					if price != "" {
						lastPrice = price
					}
					if tickCount <= 3 {
						fmt.Printf("📊 Market Data Snapshot (tick #%d): Price=%s\n", tickCount, price)
					} else if tickCount%10 == 0 {
						fmt.Printf("📊 Tick #%d: Price=%s\n", tickCount, price)
					}

				case "X":
					tickCount++
					price := extractField(msg, "270")
					if price != "" {
						lastPrice = price
					}
					if tickCount <= 3 {
						fmt.Printf("📊 Market Data Update (tick #%d): Price=%s\n", tickCount, price)
					} else if tickCount%10 == 0 {
						fmt.Printf("📊 Tick #%d: Price=%s\n", tickCount, price)
					}

				case "Y":
					fmt.Printf("❌ Market Data Request REJECTED\n")
					fmt.Printf("Full message: %s\n", debugMsg)
					rejectReason := extractField(msg, "281")
					rejectText := extractField(msg, "58")
					if rejectReason != "" {
						fmt.Printf("Reject Reason Code (281): %s\n", rejectReason)
					}
					if rejectText != "" {
						fmt.Printf("Reject Text (58): %s\n", rejectText)
					}
					return

				case "3":
					fmt.Printf("❌ Session-level REJECT\n")
					fmt.Printf("Full message: %s\n", debugMsg)
					rejectReason := extractField(msg, "373")
					rejectText := extractField(msg, "58")
					missingTag := extractField(msg, "371")
					if missingTag != "" {
						fmt.Printf("Missing Tag (371): %s\n", missingTag)
					}
					if rejectReason != "" {
						fmt.Printf("Reject Reason (373): %s\n", rejectReason)
					}
					if rejectText != "" {
						fmt.Printf("Reject Text (58): %s\n", rejectText)
					}
					return

				case "0":
					// Heartbeat - respond
					hb := buildHeartbeat(seqNum)
					conn.Write([]byte(hb))
					seqNum++

				case "1":
					// Test request - respond with heartbeat
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
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              YoFX SYMBOL DISCOVERY TEST (CORRECTED)                      ║")
	fmt.Println("║              Testing with Complete Market Data Request                  ║")
	fmt.Println("║              20-Second Monitoring Per Symbol                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝")

	activeSymbols := []string{}

	for _, symbol := range symbols {
		testSymbol(symbol)
		time.Sleep(2 * time.Second)
		// Track which symbols are active (this is a simplified version)
		// In a real implementation, you'd pass results back via channels
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                           TEST COMPLETE                                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝")

	if len(activeSymbols) > 0 {
		fmt.Println("\n📊 ACTIVE SYMBOLS:")
		for _, symbol := range activeSymbols {
			fmt.Printf("   - %s\n", symbol)
		}
	}
}
