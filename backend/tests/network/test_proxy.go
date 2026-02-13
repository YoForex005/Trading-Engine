package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	target := "81.29.145.69:49527"
	fmt.Printf("Testing TCP connection to Proxy %s...\n", target)

	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("✅ Connection successful!\n")
}
