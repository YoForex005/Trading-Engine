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

func buildLogon(senderCompID, targetCompID, username, password, account string, seqNum int) string {
	timestamp := time.Now().UTC().Format("20060102-15:04:05.000")
	
	// Build body first (everything between BodyLength and Checksum)
	body := fmt.Sprintf("35=A%s49=%s%s56=%s%s34=%d%s52=%s%s98=0%s108=30%s141=Y%s553=%s%s554=%s%s1=%s%s",
		SOH, senderCompID, SOH, targetCompID, SOH, seqNum, SOH, timestamp, SOH, SOH, SOH, SOH, username, SOH, password, SOH, account, SOH)
	
	bodyLen := len(body)
	
	// Build full message
	fullMsg := fmt.Sprintf("8=FIX.4.4%s9=%d%s%s", SOH, bodyLen, SOH, body)
	
	// Calculate checksum
	checksum := calculateChecksum(fullMsg)
	fullMsg += fmt.Sprintf("10=%03d%s", checksum, SOH)
	
	return fullMsg
}

func buildLogonMinimal(senderCompID, targetCompID, username, password string, seqNum int) string {
	timestamp := time.Now().UTC().Format("20060102-15:04:05.000")
	
	// Minimal logon - just required fields
	body := fmt.Sprintf("35=A%s49=%s%s56=%s%s34=%d%s52=%s%s98=0%s108=30%s553=%s%s554=%s%s",
		SOH, senderCompID, SOH, targetCompID, SOH, seqNum, SOH, timestamp, SOH, SOH, SOH, username, SOH, password, SOH)
	
	bodyLen := len(body)
	fullMsg := fmt.Sprintf("8=FIX.4.4%s9=%d%s%s", SOH, bodyLen, SOH, body)
	checksum := calculateChecksum(fullMsg)
	fullMsg += fmt.Sprintf("10=%03d%s", checksum, SOH)
	
	return fullMsg
}

func testLogon(conn net.Conn, label, msg string) bool {
	debugMsg := strings.ReplaceAll(msg, SOH, "|")
	fmt.Printf("\n--- Testing: %s ---\n", label)
	fmt.Printf("Sending: %s\n", debugMsg)
	
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("❌ Send error: %v\n", err)
		return false
	}
	
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			fmt.Printf("⏱️  Timeout - no response\n")
		} else if strings.Contains(err.Error(), "EOF") {
			fmt.Printf("🔌 Connection closed by server\n")
		} else {
			fmt.Printf("❌ Read error: %v\n", err)
		}
		return false
	}
	
	response := string(buf[:n])
	debugResp := strings.ReplaceAll(response, SOH, "|")
	fmt.Printf("📥 Response (%d bytes): %s\n", n, debugResp)
	
	// Parse message type
	if strings.Contains(response, "35=A") {
		fmt.Printf("✅ LOGON ACCEPTED!\n")
		return true
	} else if strings.Contains(response, "35=5") {
		fmt.Printf("❌ LOGOUT received\n")
	} else if strings.Contains(response, "35=3") {
		fmt.Printf("❌ REJECT received\n")
	}
	return false
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
	
	// Read remaining headers
	for {
		line, _ := reader.ReadString('\n')
		if line == "\r\n" || line == "\n" {
			break
		}
	}
	
	return conn, nil
}

func main() {
	fmt.Println("=== FIX 4.4 Logon Variation Tests ===\n")
	
	configs := []struct {
		label      string
		sender     string
		target     string
		username   string
		password   string
		account    string
		useAccount bool
	}{
		{"YOFX1 Full", "YOFX1", "YOFX", "YOFX1", "Brand#143", "50153", true},
		{"YOFX1 Minimal", "YOFX1", "YOFX", "YOFX1", "Brand#143", "", false},
		{"YOFX2 Full", "YOFX2", "YOFX", "YOFX2", "Brand#143", "50153", true},
	}
	
	for i, cfg := range configs {
		fmt.Printf("\n========== Test %d: %s ==========\n", i+1, cfg.label)
		
		conn, err := connectViaProxy()
		if err != nil {
			fmt.Printf("❌ Connection failed: %v\n", err)
			continue
		}
		
		var msg string
		if cfg.useAccount {
			msg = buildLogon(cfg.sender, cfg.target, cfg.username, cfg.password, cfg.account, 1)
		} else {
			msg = buildLogonMinimal(cfg.sender, cfg.target, cfg.username, cfg.password, 1)
		}
		
		testLogon(conn, cfg.label, msg)
		conn.Close()
		
		time.Sleep(2 * time.Second)
	}
	
	fmt.Println("\n\n========== SUMMARY ==========")
	fmt.Println("If all tests timeout, possible causes:")
	fmt.Println("1. Credentials incorrect (username/password)")
	fmt.Println("2. Account not enabled for FIX API access")
	fmt.Println("3. SenderCompID/TargetCompID mismatch")
	fmt.Println("4. Server expects SSL/TLS connection")
	fmt.Println("5. Additional required fields missing")
	fmt.Println("\nRecommendation: Contact broker to verify FIX credentials and settings")
}
