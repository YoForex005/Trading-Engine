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

func main() {
	proxyAddr := "81.29.145.69:49527"
	proxyUser := "fGUqTcsdMsBZlms"
	proxyPass := "3eo1qF91WA7Fyku"
	targetHost := "23.106.238.138"
	targetPort := 12336

	fmt.Println("=== Proxy Tunnel Diagnostic Test ===")
	fmt.Printf("Proxy: %s\n", proxyAddr)
	fmt.Printf("Target: %s:%d\n\n", targetHost, targetPort)

	// Step 1: Connect to proxy
	fmt.Println("[1] Connecting to proxy...")
	conn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to connect to proxy: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("✅ Connected to proxy")

	// Step 2: Send CONNECT request
	fmt.Println("\n[2] Sending HTTP CONNECT request...")
	auth := base64.StdEncoding.EncodeToString([]byte(proxyUser + ":" + proxyPass))
	connectReq := fmt.Sprintf("CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\nProxy-Authorization: Basic %s\r\n\r\n",
		targetHost, targetPort, targetHost, targetPort, auth)
	
	fmt.Printf("Request:\n%s\n", strings.ReplaceAll(connectReq, "\r\n", "\r\n\n"))
	
	_, err = conn.Write([]byte(connectReq))
	if err != nil {
		fmt.Printf("❌ Failed to send CONNECT: %v\n", err)
		return
	}
	fmt.Println("✅ CONNECT request sent")

	// Step 3: Read proxy response
	fmt.Println("\n[3] Reading proxy response...")
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	reader := bufio.NewReader(conn)
	
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ Failed to read proxy response: %v\n", err)
		return
	}
	fmt.Printf("Response: %s", response)
	
	// Check status
	if strings.Contains(response, "200") {
		fmt.Println("✅ Proxy tunnel established!")
		
		// Read remaining headers
		for {
			line, _ := reader.ReadString('\n')
			if line == "\r\n" || line == "\n" {
				break
			}
			fmt.Printf("Header: %s", line)
		}
		
		// Step 4: Test if we can send/receive data through tunnel
		fmt.Println("\n[4] Tunnel is open. Testing data flow...")
		fmt.Println("Waiting 5 seconds to see if server sends any data...")
		
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				fmt.Println("⚠️  No data from server (timeout) - server waiting for us to send first")
			} else {
				fmt.Printf("⚠️  Read error: %v\n", err)
			}
		} else {
			fmt.Printf("📥 Received %d bytes: %x\n", n, buf[:n])
		}
		
		fmt.Println("\n✅ Tunnel is working - FIX server port is reachable through proxy")
		fmt.Println("The FIX Logon message sent but server not responding suggests:")
		fmt.Println("  1. Credentials may be incorrect")
		fmt.Println("  2. Account may not be enabled for FIX API")
		fmt.Println("  3. Server may require different message format")
		
	} else {
		fmt.Printf("❌ Proxy returned error: %s", response)
		// Read rest of error response
		for {
			line, err := reader.ReadString('\n')
			if err != nil || line == "\r\n" {
				break
			}
			fmt.Printf("  %s", line)
		}
	}
}
