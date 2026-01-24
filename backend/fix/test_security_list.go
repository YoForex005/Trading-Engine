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

func buildSecurityListRequest(seqNum int) string {
	timestamp := time.Now().UTC().Format("20060102-15:04:05.000")
	reqID := fmt.Sprintf("SECLIST-%d", time.Now().Unix())

	// Security List Request (MsgType=x)
	// Tag 320: SecurityReqID
	// Tag 559: SecurityListRequestType (0=Symbol, 4=All Securities)
	body := fmt.Sprintf("35=x%s49=YOFX2%s56=YOFX%s34=%d%s52=%s%s"+
		"320=%s%s"+    // SecurityReqID
		"559=4%s",     // SecurityListRequestType: 4=All Securities
		SOH, SOH, SOH, seqNum, SOH, timestamp, SOH,
		reqID, SOH,
		SOH)

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

func extractAllFields(msg, tag string) []string {
	parts := strings.Split(msg, SOH)
	results := []string{}
	for _, part := range parts {
		if strings.HasPrefix(part, tag+"=") {
			results = append(results, strings.TrimPrefix(part, tag+"="))
		}
	}
	return results
}

func main() {
	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              YoFX SECURITY LIST REQUEST TEST                             ║")
	fmt.Println("║              Requesting All Available Symbols from Server                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝\n")

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
		debugResp := strings.ReplaceAll(response, SOH, "|")
		fmt.Printf("Response: %s\n", debugResp)
		return
	}

	fmt.Println("✅ Logged in")

	// Request Security List
	secListReq := buildSecurityListRequest(seqNum)
	debugReq := strings.ReplaceAll(secListReq, SOH, "|")
	fmt.Printf("\n📤 Sending Security List Request:\n   %s\n\n", debugReq)
	conn.Write([]byte(secListReq))
	seqNum++

	fmt.Println("⏳ Waiting for Security List response (30 seconds)...\n")

	// Monitor for responses
	timeout := time.After(30 * time.Second)
	heartbeatTicker := time.NewTicker(10 * time.Second)
	defer heartbeatTicker.Stop()

	symbols := []string{}
	gotSecurityList := false

	for {
		select {
		case <-timeout:
			if gotSecurityList {
				fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════╗")
				fmt.Printf("║  AVAILABLE SYMBOLS: %-52d  ║\n", len(symbols))
				fmt.Println("╚══════════════════════════════════════════════════════════════════════════╝\n")

				if len(symbols) > 0 {
					for i, symbol := range symbols {
						fmt.Printf("%3d. %s\n", i+1, symbol)
					}
				} else {
					fmt.Println("⚠️  No symbols in the response")
				}
			} else {
				fmt.Println("\n❌ No Security List response received")
				fmt.Println("   The server might not support Security List Request (MsgType=x)")
				fmt.Println("   Or it might require different parameters")
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
				case "y":
					// Security List (MsgType=y) - response to our request
					fmt.Println("✅ Received Security List response!")
					gotSecurityList = true

					// Extract all symbols (tag 55)
					foundSymbols := extractAllFields(msg, "55")
					symbols = append(symbols, foundSymbols...)

					// Show partial message
					if len(debugMsg) > 500 {
						fmt.Printf("📥 Message (truncated): %s...\n", debugMsg[:500])
					} else {
						fmt.Printf("📥 Message: %s\n", debugMsg)
					}

					// Extract other useful fields
					numSecurities := extractField(msg, "146")
					if numSecurities != "" {
						fmt.Printf("   Number of securities (tag 146): %s\n", numSecurities)
					}

					secReqResult := extractField(msg, "560")
					if secReqResult != "" {
						fmt.Printf("   Security Request Result (tag 560): %s\n", secReqResult)
					}

				case "3":
					// Reject
					fmt.Printf("❌ Session-level REJECT\n")
					fmt.Printf("Full message: %s\n", debugMsg)
					rejectText := extractField(msg, "58")
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

				default:
					if msgType != "" && msgType != "A" {
						fmt.Printf("📥 Received message type %s\n", msgType)
						if len(debugMsg) > 200 {
							fmt.Printf("   %s...\n", debugMsg[:200])
						} else {
							fmt.Printf("   %s\n", debugMsg)
						}
					}
				}
			}
		}
	}
}
