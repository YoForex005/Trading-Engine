package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	proxyHost := "81.29.145.69"
	proxyPort := "49527"
	proxyUser := "fGUqTcsdMsBZlms"
	proxyPass := "3eo1qF91WA7Fyku"
	targetHost := "23.106.238.138"
	targetPort := "12336"

	fmt.Printf("Testing SOCKS5 Proxy %s:%s -> %s:%s\n", proxyHost, proxyPort, targetHost, targetPort)

	// 1. Connect to Proxy
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(proxyHost, proxyPort), 5*time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to connect to proxy: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("✅ TCP connection to proxy established")

	// 2. SOCKS5 Handshake
	// Ver 5, 1 Method, Method 2 (Username/Password)
	conn.Write([]byte{0x05, 0x01, 0x02})

	resp := make([]byte, 2)
	if _, err := conn.Read(resp); err != nil {
		fmt.Printf("❌ Failed to read handshake response: %v\n", err)
		return
	}
	if resp[0] != 0x05 || resp[1] != 0x02 {
		fmt.Printf("❌ Proxy does not support Username/Password auth (Got: %v)\n", resp)
		return
	}
	fmt.Println("✅ Proxy accepted Username/Password method")

	// 3. Authenticate
	authMsg := []byte{0x01} // Ver 1
	authMsg = append(authMsg, byte(len(proxyUser)))
	authMsg = append(authMsg, []byte(proxyUser)...)
	authMsg = append(authMsg, byte(len(proxyPass)))
	authMsg = append(authMsg, []byte(proxyPass)...)

	conn.Write(authMsg)

	authResp := make([]byte, 2)
	if _, err := conn.Read(authResp); err != nil {
		fmt.Printf("❌ Failed to read auth response: %v\n", err)
		return
	}
	if authResp[1] != 0x00 {
		fmt.Printf("❌ Authentication failed! (Status: %x)\n", authResp[1])
		return
	}
	fmt.Println("✅ Proxy Authentication successful")

	// 4. Connect Request
	// Ver 5, Cmd 1 (Connect), Rsv 0, AddrTyp 1 (IPv4)
	req := []byte{0x05, 0x01, 0x00, 0x01}
	ip := net.ParseIP(targetHost).To4()
	req = append(req, ip...)
	// Port 12336 = 0x3030
	req = append(req, 0x30, 0x30)

	conn.Write(req)

	connResp := make([]byte, 10)
	if _, err := conn.Read(connResp); err != nil {
		fmt.Printf("❌ Failed to read connect response: %v\n", err)
		return
	}
	if connResp[1] != 0x00 {
		fmt.Printf("❌ SOCKS5 Connect failed (Status: %x)\n", connResp[1])
		return
	}

	fmt.Println("✅ SOCKS5 Tunnel Successfully Established!")
}
