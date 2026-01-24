// +build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// SessionConfig represents FIX session configuration
type SessionConfig struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
	Purpose           string `json:"purpose"`
	Connection        struct {
		TargetIP   string `json:"targetIP"`
		TargetPort int    `json:"targetPort"`
		SSL        bool   `json:"ssl"`
	} `json:"connection"`
	FixProtocol struct {
		BeginString  string `json:"beginString"`
		SenderCompID string `json:"senderCompID"`
		TargetCompID string `json:"targetCompID"`
	} `json:"fixProtocol"`
	Authentication struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"authentication"`
	TradingAccount string `json:"tradingAccount"`
	Settings       struct {
		HeartBeatInterval int `json:"heartBeatInterval"`
		ReconnectInterval int `json:"reconnectInterval"`
		LogonTimeout      int `json:"logonTimeout"`
	} `json:"settings"`
}

type SessionsConfig struct {
	Sessions []SessionConfig `json:"sessions"`
}

const (
	SOH = "\x01" // FIX field delimiter
)

// connectViaHTTPProxy establishes connection through HTTP CONNECT proxy
func connectViaHTTPProxy(proxyAddr, username, password, targetHost string, targetPort int) (net.Conn, error) {
	// Connect to proxy
	conn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %v", err)
	}

	// Send HTTP CONNECT request
	connectReq := fmt.Sprintf("CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\n",
		targetHost, targetPort, targetHost, targetPort)

	// Add Proxy-Authorization if credentials provided
	if username != "" && password != "" {
		auth := fmt.Sprintf("%s:%s", username, password)
		authEncoded := base64Encode(auth)
		connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", authEncoded)
	}

	connectReq += "\r\n"

	if _, err := conn.Write([]byte(connectReq)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send CONNECT request: %v", err)
	}

	// Read proxy response
	reader := bufio.NewReader(conn)
	responseLine, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read proxy response: %v", err)
	}

	// Parse HTTP status line
	if !strings.HasPrefix(responseLine, "HTTP/1.") {
		conn.Close()
		return nil, fmt.Errorf("invalid HTTP response: %s", responseLine)
	}

	// Check status code
	parts := strings.Fields(responseLine)
	if len(parts) < 2 || parts[1] != "200" {
		conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed: %s", responseLine)
	}

	// Read remaining headers until empty line
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to read proxy headers: %v", err)
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}

	return conn, nil
}

// base64Encode encodes a string to base64
func base64Encode(s string) string {
	var buf [1024]byte
	encoded := buf[:0]
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	data := []byte(s)
	for len(data) > 0 {
		var chunk [3]byte
		chunkLen := copy(chunk[:], data)
		data = data[chunkLen:]

		encoded = append(encoded,
			encodeStd[chunk[0]>>2],
			encodeStd[(chunk[0]&0x03)<<4|chunk[1]>>4])

		if chunkLen > 1 {
			encoded = append(encoded, encodeStd[(chunk[1]&0x0f)<<2|chunk[2]>>6])
		} else {
			encoded = append(encoded, '=')
		}

		if chunkLen > 2 {
			encoded = append(encoded, encodeStd[chunk[2]&0x3f])
		} else {
			encoded = append(encoded, '=')
		}
	}

	return string(encoded)
}

// buildFIXMessage creates a FIX protocol message
func buildFIXMessage(msgType string, fields map[int]string) string {
	// Extract and remove header fields to ensure proper ordering
	beginString := fields[8]
	senderCompID := fields[49]
	targetCompID := fields[56]
	msgSeqNum := fields[34]
	sendingTime := fields[52]

	// Build the body (all fields except standard header and checksum)
	bodyFields := ""

	// Add standard header fields in proper order (after MsgType)
	bodyFields += fmt.Sprintf("49=%s%s", senderCompID, SOH)
	bodyFields += fmt.Sprintf("56=%s%s", targetCompID, SOH)
	bodyFields += fmt.Sprintf("34=%s%s", msgSeqNum, SOH)
	bodyFields += fmt.Sprintf("52=%s%s", sendingTime, SOH)

	// Add all other fields (excluding header fields 8, 9, 35, 49, 56, 34, 52, 10)
	excludedTags := map[int]bool{8: true, 9: true, 35: true, 49: true, 56: true, 34: true, 52: true, 10: true}
	for tag, value := range fields {
		if !excludedTags[tag] {
			bodyFields += fmt.Sprintf("%d=%s%s", tag, value, SOH)
		}
	}

	// Build body (MsgType + standard header + other fields)
	body := fmt.Sprintf("35=%s%s%s", msgType, SOH, bodyFields)

	// Calculate body length
	bodyLength := len(body)

	// Build full message with BeginString and BodyLength
	fullMsg := fmt.Sprintf("8=%s%s9=%d%s%s", beginString, SOH, bodyLength, SOH, body)

	// Calculate and add checksum
	checksum := calculateChecksum(fullMsg)
	fullMsg += fmt.Sprintf("10=%03d%s", checksum, SOH)

	return fullMsg
}

// calculateChecksum computes FIX message checksum
func calculateChecksum(msg string) int {
	sum := 0
	for i := 0; i < len(msg); i++ {
		sum += int(msg[i])
	}
	return sum % 256
}

// parseFIXMessage parses a FIX message into field map
func parseFIXMessage(msg string) map[string]string {
	fields := make(map[string]string)
	parts := strings.Split(msg, SOH)

	for _, part := range parts {
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			fields[kv[0]] = kv[1]
		}
	}

	return fields
}

// testFIXConnection tests connection and authentication to FIX server
func testFIXConnection(config SessionConfig, proxyAddr, proxyUser, proxyPass string) error {
	fmt.Printf("\n=== Testing FIX 4.4 Connection ===\n")
	fmt.Printf("Session: %s (%s)\n", config.Name, config.ID)
	fmt.Printf("Server: %s:%d\n", config.Connection.TargetIP, config.Connection.TargetPort)
	fmt.Printf("Protocol: %s\n", config.FixProtocol.BeginString)
	fmt.Printf("SenderCompID: %s\n", config.FixProtocol.SenderCompID)
	fmt.Printf("TargetCompID: %s\n", config.FixProtocol.TargetCompID)
	fmt.Printf("Username: %s\n", config.Authentication.Username)
	if proxyAddr != "" {
		fmt.Printf("Proxy: %s (authenticated)\n", proxyAddr)
	}
	fmt.Printf("\n")

	// Step 1: Connect to server (via proxy if specified)
	var conn net.Conn
	var err error

	if proxyAddr != "" {
		fmt.Printf("[1/4] Connecting via HTTP proxy %s...\n", proxyAddr)
		conn, err = connectViaHTTPProxy(proxyAddr, proxyUser, proxyPass,
			config.Connection.TargetIP, config.Connection.TargetPort)
	} else {
		fmt.Printf("[1/4] Connecting directly to %s:%d...\n", config.Connection.TargetIP, config.Connection.TargetPort)
		address := fmt.Sprintf("%s:%d", config.Connection.TargetIP, config.Connection.TargetPort)
		conn, err = net.DialTimeout("tcp", address, 10*time.Second)
	}

	if err != nil {
		return fmt.Errorf("❌ Connection failed: %v", err)
	}
	defer conn.Close()

	fmt.Printf("✅ TCP connection established\n\n")

	// Step 2: Send Logon message
	fmt.Printf("[2/4] Sending Logon message with authentication...\n")

	timestamp := time.Now().UTC().Format("20060102-15:04:05.000")

	// Note: Server does NOT accept Account (1) or ResetSeqNumFlag (141) fields
	// These cause the server to not respond - use minimal logon format
	logonFields := map[int]string{
		8:   config.FixProtocol.BeginString,                  // BeginString
		49:  config.FixProtocol.SenderCompID,                 // SenderCompID
		56:  config.FixProtocol.TargetCompID,                 // TargetCompID
		34:  "1",                                             // MsgSeqNum
		52:  timestamp,                                       // SendingTime
		98:  "0",                                             // EncryptMethod (None)
		108: strconv.Itoa(config.Settings.HeartBeatInterval), // HeartBtInt
		553: config.Authentication.Username,                  // Username
		554: config.Authentication.Password,                  // Password
		// 141: "Y",                                          // ResetSeqNumFlag - NOT SUPPORTED
		// 1:   config.TradingAccount,                        // Account - NOT SUPPORTED
	}

	logonMsg := buildFIXMessage("A", logonFields) // A = Logon

	// Debug: Print raw message
	debugMsg := strings.ReplaceAll(logonMsg, SOH, "|")
	fmt.Printf("Logon message (length=%d bytes):\n%s\n", len(logonMsg), debugMsg)

	// Send logon
	_, err = conn.Write([]byte(logonMsg))
	if err != nil {
		return fmt.Errorf("❌ Failed to send Logon: %v", err)
	}

	fmt.Printf("✅ Logon message sent\n\n")

	// Step 3: Wait for response
	fmt.Printf("[3/4] Waiting for server response (30 second timeout)...\n")

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Try to read any available data
	response := make([]byte, 4096)
	n, err := conn.Read(response)

	if err != nil {
		if n > 0 {
			// Got partial data before error
			fmt.Printf("⚠️  Received %d bytes before error: %v\n", n, err)
			responseStr := string(response[:n])
			debugResp := strings.ReplaceAll(responseStr, SOH, "|")
			fmt.Printf("Partial data: %s\n", debugResp)

			// Try to parse what we got
			fields := parseFIXMessage(responseStr)
			if msgType, ok := fields["35"]; ok {
				fmt.Printf("Received message type: %s\n", msgType)
			}
		}

		// Check if this is a network timeout or connection closed
		if strings.Contains(err.Error(), "timeout") {
			return fmt.Errorf("❌ Timeout waiting for response - server not responding (may indicate: wrong credentials, server offline, or IP not whitelisted)")
		} else if strings.Contains(err.Error(), "EOF") {
			return fmt.Errorf("❌ Server closed connection immediately after Logon (may indicate: authentication rejected)")
		}

		return fmt.Errorf("❌ Failed to read response: %v", err)
	}

	responseStr := string(response[:n])

	// Debug: Print raw response
	debugResp := strings.ReplaceAll(responseStr, SOH, "|")
	fmt.Printf("✅ Received response (%d bytes):\n%s\n\n", n, debugResp)

	fields := parseFIXMessage(responseStr)

	// Step 4: Verify authentication
	fmt.Printf("[4/4] Verifying authentication...\n")

	msgType := fields["35"]

	if msgType == "A" {
		// Logon accepted
		fmt.Printf("✅ Authentication SUCCESSFUL!\n")
		fmt.Printf("\nSession Details:\n")
		fmt.Printf("  - Status: LOGGED_IN\n")
		fmt.Printf("  - MsgSeqNum: %s\n", fields["34"])
		fmt.Printf("  - HeartBtInt: %s seconds\n", fields["108"])
		if sendTime, ok := fields["52"]; ok {
			fmt.Printf("  - Server Time: %s\n", sendTime)
		}
		if testReqID, ok := fields["112"]; ok {
			fmt.Printf("  - Test Request ID: %s\n", testReqID)
		}

		// Send Logout to cleanly disconnect
		fmt.Printf("\nSending Logout to disconnect cleanly...\n")
		logoutFields := map[int]string{
			8:  config.FixProtocol.BeginString,
			49: config.FixProtocol.SenderCompID,
			56: config.FixProtocol.TargetCompID,
			34: "2",
			52: time.Now().UTC().Format("20060102-15:04:05.000"),
		}
		logoutMsg := buildFIXMessage("5", logoutFields) // 5 = Logout
		conn.Write([]byte(logoutMsg))

		return nil

	} else if msgType == "3" {
		// Reject
		rejectReason := fields["58"]
		if rejectReason == "" {
			rejectReason = "Unknown reason"
		}
		return fmt.Errorf("❌ Authentication REJECTED: %s (MsgType: %s)", rejectReason, msgType)

	} else if msgType == "5" {
		// Logout (server rejected)
		logoutText := fields["58"]
		if logoutText == "" {
			logoutText = "Server initiated logout"
		}
		return fmt.Errorf("❌ Authentication FAILED: %s", logoutText)

	} else {
		return fmt.Errorf("❌ Unexpected response type: %s", msgType)
	}
}

func main() {
	// Parse proxy from environment or command line
	// Format: username:password@host:port
	proxyStr := os.Getenv("HTTP_PROXY")
	if proxyStr == "" && len(os.Args) > 1 {
		proxyStr = os.Args[1]
	}

	var proxyAddr, proxyUser, proxyPass string
	if proxyStr != "" {
		// Parse proxy string
		parts := strings.Split(proxyStr, "@")
		if len(parts) == 2 {
			authParts := strings.Split(parts[0], ":")
			if len(authParts) == 2 {
				proxyUser = authParts[0]
				proxyPass = authParts[1]
			}
			proxyAddr = parts[1]
		} else {
			proxyAddr = proxyStr
		}
		fmt.Printf("Using HTTP proxy: %s\n", proxyAddr)
	}

	// Load configuration
	configPath := "./config/sessions.json"

	fmt.Printf("Reading configuration from: %s\n", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var sessionsConfig SessionsConfig
	if err := json.Unmarshal(data, &sessionsConfig); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Test each enabled session
	successCount := 0
	failCount := 0

	for _, session := range sessionsConfig.Sessions {
		if !session.Enabled {
			fmt.Printf("\n⏭️  Skipping disabled session: %s\n", session.ID)
			continue
		}

		err := testFIXConnection(session, proxyAddr, proxyUser, proxyPass)

		if err != nil {
			fmt.Printf("\n%v\n", err)
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			failCount++
		} else {
			fmt.Printf("\n✅ Test completed successfully!\n")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			successCount++
		}
	}

	// Print summary
	fmt.Printf("\n\n╔════════════════════════════════════╗\n")
	fmt.Printf("║        TEST SUMMARY                ║\n")
	fmt.Printf("╠════════════════════════════════════╣\n")
	fmt.Printf("║ Total Sessions: %-2d                 ║\n", successCount+failCount)
	fmt.Printf("║ Successful:     %-2d ✅              ║\n", successCount)
	fmt.Printf("║ Failed:         %-2d ❌              ║\n", failCount)
	fmt.Printf("╚════════════════════════════════════╝\n")

	if failCount > 0 {
		os.Exit(1)
	}
}
