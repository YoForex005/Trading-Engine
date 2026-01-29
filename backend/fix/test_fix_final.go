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
	
	// Minimal logon - just required fields (no Account, no ResetSeqNumFlag)
	body := fmt.Sprintf("35=A%s49=%s%s56=%s%s34=%d%s52=%s%s98=0%s108=30%s553=%s%s554=%s%s",
		SOH, senderCompID, SOH, targetCompID, SOH, seqNum, SOH, timestamp, SOH, SOH, SOH, username, SOH, password, SOH)
	
	bodyLen := len(body)
	fullMsg := fmt.Sprintf("8=FIX.4.4%s9=%d%s%s", SOH, bodyLen, SOH, body)
	checksum := calculateChecksum(fullMsg)
	fullMsg += fmt.Sprintf("10=%03d%s", checksum, SOH)
	
	return fullMsg
}

func buildLogout(senderCompID, targetCompID string, seqNum int) string {
	timestamp := time.Now().UTC().Format("20060102-15:04:05.000")
	body := fmt.Sprintf("35=5%s49=%s%s56=%s%s34=%d%s52=%s%s", SOH, senderCompID, SOH, targetCompID, SOH, seqNum, SOH, timestamp, SOH)
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

func testSession(sender, target, username, password, sessionName string) bool {
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Testing: %-50s  ║\n", sessionName)
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n")
	
	conn, err := connectViaProxy()
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return false
	}
	defer conn.Close()
	
	fmt.Printf("✅ Proxy tunnel established\n")
	
	// Send Logon
	logon := buildLogon(sender, target, username, password, 1)
	debugMsg := strings.ReplaceAll(logon, SOH, "|")
	fmt.Printf("📤 Logon: %s\n", debugMsg)
	
	conn.Write([]byte(logon))
	
	// Read response
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	
	if err != nil {
		fmt.Printf("❌ No response: %v\n", err)
		return false
	}
	
	response := string(buf[:n])
	debugResp := strings.ReplaceAll(response, SOH, "|")
	fmt.Printf("📥 Response: %s\n", debugResp)
	
	if strings.Contains(response, "35=A") {
		fmt.Printf("\n✅ AUTHENTICATION SUCCESSFUL!\n")
		fmt.Printf("   Session: %s\n", sessionName)
		fmt.Printf("   Status: LOGGED_IN\n")
		
		// Send clean logout
		logout := buildLogout(sender, target, 2)
		conn.Write([]byte(logout))
		fmt.Printf("📤 Sent Logout for clean disconnect\n")
		
		return true
	} else if strings.Contains(response, "35=5") {
		fmt.Printf("❌ LOGOUT received - authentication rejected\n")
	} else if strings.Contains(response, "35=3") {
		fmt.Printf("❌ REJECT received\n")
	}
	
	return false
}

func main() {
	fmt.Println("\n╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║         FIX 4.4 CONNECTION VERIFICATION TEST                 ║")
	fmt.Println("║         Server: 23.106.238.138:12336                         ║")
	fmt.Println("║         Proxy: 81.29.145.69:49527 (Whitelisted)              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	
	success := 0
	fail := 0
	
	// Test YOFX1
	if testSession("YOFX1", "YOFX", "YOFX1", "Brand#143", "YOFX1 - Trading Account") {
		success++
	} else {
		fail++
	}
	
	time.Sleep(2 * time.Second)
	
	// Test YOFX2
	if testSession("YOFX2", "YOFX", "YOFX2", "Brand#143", "YOFX2 - Market Data Feed") {
		success++
	} else {
		fail++
	}
	
	fmt.Println("\n╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    FINAL RESULTS                             ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Sessions Tested:    2                                       ║\n")
	fmt.Printf("║  Successful:         %d  ✅                                    ║\n", success)
	fmt.Printf("║  Failed:             %d  ❌                                    ║\n", fail)
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	if success == 2 {
		fmt.Println("║  STATUS: ALL CONNECTIONS VERIFIED                           ║")
	} else if success > 0 {
		fmt.Println("║  STATUS: PARTIAL SUCCESS                                    ║")
	} else {
		fmt.Println("║  STATUS: VERIFICATION FAILED                                ║")
	}
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}
