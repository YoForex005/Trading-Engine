package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	target := "23.106.238.138:12336"
	fmt.Printf("Testing TCP connection to %s...\n", target)

	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("✅ Connection successful!\n")
	fmt.Printf("Local Addr: %s\n", conn.LocalAddr())
	fmt.Printf("Remote Addr: %s\n", conn.RemoteAddr())
}
