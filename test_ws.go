package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}

func main() {
	serverAddr := flag.String("server", "localhost:7999", "server address")
	username := flag.String("user", "1", "username")
	password := flag.String("pass", "password", "password")
	flag.Parse()

	log.Printf("1. Authenticating as %s...", *username)

	// 1. Login to get Token
	loginURL := fmt.Sprintf("http://%s/login", *serverAddr)
	jsonData, _ := json.Marshal(map[string]string{
		"username": *username,
		"password": *password,
	})

	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("Login failed (Status %d): %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		log.Fatalf("Failed to decode login response: %v", err)
	}

	log.Printf("   Success! Token: %s...", loginResp.Token[:20])

	// 2. Connect to WebSocket
	u := url.URL{Scheme: "ws", Host: *serverAddr, Path: "/ws", RawQuery: "token=" + loginResp.Token}
	log.Printf("2. Connecting to %s...", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("WebSocket connection failed: %v", err)
	}
	defer c.Close()

	log.Println("   Connected! Waiting for messages...")

	done := make(chan struct{})

	go func() {
		defer close(done)
		msgCount := 0
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			msgCount++
			if msgCount <= 5 {
				log.Printf("   recv: %s", message)
			} else if msgCount%50 == 0 {
				log.Printf("   recv: (total %d messages received)", msgCount)
			}
		}
	}()

	// Keep alive for 30 seconds then exit
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		log.Println("interrupt")
	case <-time.After(30 * time.Second):
		log.Println("Test completed successfully (30s timeout)")
	}

	c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
