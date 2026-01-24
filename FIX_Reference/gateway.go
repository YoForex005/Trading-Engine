package fix

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// FIX message types - Session Level
	MsgTypeLogon          = "A"
	MsgTypeLogout         = "5"
	MsgTypeHeartbeat      = "0"
	MsgTypeTestRequest    = "1"
	MsgTypeResendRequest  = "2"
	MsgTypeReject         = "3"
	MsgTypeSequenceReset  = "4"
	MsgTypeBusinessReject = "j"

	// FIX message types - Trading
	MsgTypeNewOrderSingle     = "D"
	MsgTypeOrderCancelRequest = "F"
	MsgTypeOrderCancelReject  = "9"
	MsgTypeOrderStatusRequest = "H"
	MsgTypeExecutionReport    = "8"
	MsgTypeOrderMassStatusReq = "AF"

	// FIX message types - Market Data
	MsgTypeMarketDataRequest     = "V"
	MsgTypeMarketDataSnapshot    = "W"
	MsgTypeMarketDataIncremental = "X"
	MsgTypeMarketDataReject      = "Y"

	// FIX message types - Position & Trade History
	MsgTypeRequestForPositions    = "AN"
	MsgTypeRequestForPositionsAck = "AO"
	MsgTypePositionReport         = "AP"
	MsgTypeTradeCaptureReportReq  = "AD"
	MsgTypeTradeCaptureReportAck  = "AQ"
	MsgTypeTradeCaptureReport     = "AE"

	// FIX message types - Security Definition
	MsgTypeSecurityListRequest   = "x"
	MsgTypeSecurityList          = "y"
	MsgTypeSecurityDefinitionReq = "c"
	MsgTypeSecurityDefinition    = "d"

	// Store directory for sequence numbers
	DefaultStoreDir = "./fixstore"
)

// LPSession represents a connection to a Liquidity Provider
type LPSession struct {
	ID             string
	Name           string
	Host           string
	Port           int
	SenderCompID   string
	TargetCompID   string
	Username       string
	Password       string
	TradingAccount string
	BeginString    string
	SSL            bool
	// Proxy settings
	UseProxy      bool
	ProxyHost     string
	ProxyPort     int
	ProxyUsername string
	ProxyPassword string
	Status        string // DISCONNECTED, CONNECTING, CONNECTED, LOGGED_IN
	LastHeartbeat time.Time
	conn          net.Conn

	// Sequence number management (critical for FIX protocol)
	OutSeqNum       int            // Next outgoing sequence number
	InSeqNum        int            // Expected incoming sequence number
	ResetSeqNumFlag bool           // Whether to reset sequence numbers on logon
	msgStore        map[int]string // Store sent messages for potential resend (seqNum -> message)
	msgStoreMu      sync.RWMutex   // Mutex for message store
	storeDir        string         // Directory for persisting sequence numbers
}

// ExecutionReport represents a fill or reject from LP
type ExecutionReport struct {
	OrderID   string
	ExecType  string // NEW, FILLED, REJECTED, CANCELED
	Symbol    string
	Side      string
	Volume    float64
	Price     float64
	LPOrderID string
	Text      string
	Timestamp time.Time
}

// MarketData represents a price quote from LP
type MarketData struct {
	Symbol    string
	Bid       float64
	Ask       float64
	BidSize   float64
	AskSize   float64
	MDReqID   string
	SessionID string
	Timestamp time.Time
}

// MarketDataReject represents a rejected market data subscription
type MarketDataReject struct {
	MDReqID   string
	Reason    string
	Text      string
	SessionID string
}

// Position represents an open position from LP
type Position struct {
	Symbol      string
	Side        string  // BUY or SELL
	Volume      float64 // LongQty or ShortQty
	EntryPrice  float64 // SettlPrice (730)
	CurrentPnL  float64 // Calculated P&L
	PosReqID    string
	SessionID   string
	Account     string
	PositionID  string
	Timestamp   time.Time
	TimestampMs int64 // Millisecond precision
}

// TradeCapture represents a historical trade
type TradeCapture struct {
	TradeID      string // ExecID (17)
	OrderID      string // OrderID (37)
	ClOrdID      string // ClOrdID (11)
	Symbol       string
	Side         string  // BUY or SELL
	Volume       float64 // LastQty (32)
	Price        float64 // LastPx (31)
	Account      string
	TradeDate    string // YYYYMMDD
	TransactTime time.Time
	SessionID    string
	TimestampMs  int64
}

// OrderStatus represents current status of an order
type OrderStatus struct {
	OrderID     string
	ClOrdID     string
	Symbol      string
	Side        string
	OrdStatus   string // 0=New, 1=PartialFill, 2=Filled, 4=Canceled, 8=Rejected
	OrdType     string // 1=Market, 2=Limit
	Price       float64
	OrderQty    float64
	CumQty      float64 // Filled quantity
	LeavesQty   float64 // Remaining quantity
	AvgPx       float64 // Average fill price
	Account     string
	Text        string
	SessionID   string
	Timestamp   time.Time
	TimestampMs int64
}

// FIXGateway manages connections to Liquidity Providers
type FIXGateway struct {
	sessions            map[string]*LPSession
	execReports         chan ExecutionReport
	marketData          chan MarketData
	mdRejects           chan MarketDataReject
	positions           chan Position
	trades              chan TradeCapture
	orderStatuses       chan OrderStatus
	mdSubscriptions     map[string]string      // MDReqID -> Symbol mapping
	symbolSubscriptions map[string]string      // Symbol -> MDReqID mapping (reverse lookup)
	posSubscriptions    map[string]bool        // PosReqID -> active
	quoteCache          map[string]*MarketData // Symbol -> Last known quote (for merging incremental updates)
	quoteCacheMu        sync.RWMutex
	mu                  sync.RWMutex
}

func NewFIXGateway() *FIXGateway {
	// Ensure store directory exists
	storeDir := getEnvOrDefault("FIX_STORE_DIR", DefaultStoreDir)
	os.MkdirAll(storeDir, 0755)

	gw := &FIXGateway{
		sessions: map[string]*LPSession{
			"LMAX_PROD": {
				ID:              "LMAX_PROD",
				Name:            "LMAX Exchange",
				Host:            "fix.lmax.com",
				Port:            443,
				SenderCompID:    "RTX_BROKER",
				TargetCompID:    "LMAX",
				BeginString:     "FIX.4.4",
				Status:          "DISCONNECTED",
				OutSeqNum:       0, // Will be loaded from store or start at 1
				InSeqNum:        0,
				ResetSeqNumFlag: false,
				msgStore:        make(map[int]string),
				storeDir:        storeDir,
			},
			"LMAX_DEMO": {
				ID:              "LMAX_DEMO",
				Name:            "LMAX Demo",
				Host:            "demo-fix.lmax.com",
				Port:            443,
				SenderCompID:    "RTX_BROKER_DEMO",
				TargetCompID:    "LMAX_DEMO",
				BeginString:     "FIX.4.4",
				Status:          "DISCONNECTED",
				OutSeqNum:       0,
				InSeqNum:        0,
				ResetSeqNumFlag: false,
				msgStore:        make(map[int]string),
				storeDir:        storeDir,
			},
			"YOFX1": {
				ID:              "YOFX1",
				Name:            "YOFX Trading Account",
				Host:            getEnvOrDefault("YOFX_HOST", "23.106.238.138"),
				Port:            getEnvIntOrDefault("YOFX_PORT", 12336), // T4B FIX server port
				SenderCompID:    getEnvOrDefault("YOFX1_SENDER_COMP_ID", "YOFX1"),
				TargetCompID:    getEnvOrDefault("YOFX_TARGET_COMP_ID", "YOFX"),
				Username:        getEnvOrDefault("YOFX1_USERNAME", "YOFX1"),
				Password:        getEnvOrDefault("YOFX1_PASSWORD", "Brand#143"),
				TradingAccount:  getEnvOrDefault("YOFX_TRADING_ACCOUNT", "50153"),
				BeginString:     "FIX.4.4",
				SSL:             getEnvOrDefault("YOFX_SSL", "false") == "true",
				UseProxy:        getEnvOrDefault("YOFX_USE_PROXY", "true") == "true",
				ProxyHost:       getEnvOrDefault("YOFX_PROXY_HOST", "81.29.145.69"),
				ProxyPort:       getEnvIntOrDefault("YOFX_PROXY_PORT", 49527),
				ProxyUsername:   getEnvOrDefault("YOFX_PROXY_USERNAME", "fGUqTcsdMsBZlms"),
				ProxyPassword:   getEnvOrDefault("YOFX_PROXY_PASSWORD", "3eo1qF91WA7Fyku"),
				Status:          "DISCONNECTED",
				OutSeqNum:       0,
				InSeqNum:        0,
				ResetSeqNumFlag: getEnvOrDefault("FIX_RESET_SEQ", "false") == "true",
				msgStore:        make(map[int]string),
				storeDir:        storeDir,
			},
			"YOFX2": {
				ID:              "YOFX2",
				Name:            "YOFX Market Data Feed",
				Host:            getEnvOrDefault("YOFX_HOST", "23.106.238.138"),
				Port:            getEnvIntOrDefault("YOFX_PORT", 12336), // T4B FIX server port
				SenderCompID:    getEnvOrDefault("YOFX2_SENDER_COMP_ID", "YOFX2"),
				TargetCompID:    getEnvOrDefault("YOFX_TARGET_COMP_ID", "YOFX"),
				Username:        getEnvOrDefault("YOFX2_USERNAME", "YOFX2"),
				Password:        getEnvOrDefault("YOFX2_PASSWORD", "Brand#143"),
				TradingAccount:  getEnvOrDefault("YOFX_TRADING_ACCOUNT", "50153"),
				BeginString:     "FIX.4.4",
				SSL:             getEnvOrDefault("YOFX_SSL", "false") == "true",
				UseProxy:        getEnvOrDefault("YOFX_USE_PROXY", "true") == "true",
				ProxyHost:       getEnvOrDefault("YOFX_PROXY_HOST", "81.29.145.69"),
				ProxyPort:       getEnvIntOrDefault("YOFX_PROXY_PORT", 49527),
				ProxyUsername:   getEnvOrDefault("YOFX_PROXY_USERNAME", "fGUqTcsdMsBZlms"),
				ProxyPassword:   getEnvOrDefault("YOFX_PROXY_PASSWORD", "3eo1qF91WA7Fyku"),
				Status:          "DISCONNECTED",
				OutSeqNum:       0,
				InSeqNum:        0,
				ResetSeqNumFlag: getEnvOrDefault("FIX_RESET_SEQ", "false") == "true",
				msgStore:        make(map[int]string),
				storeDir:        storeDir,
			},
		},
		execReports:         make(chan ExecutionReport, 1000),
		marketData:          make(chan MarketData, 10000),
		mdRejects:           make(chan MarketDataReject, 100),
		positions:           make(chan Position, 1000),
		trades:              make(chan TradeCapture, 5000),
		orderStatuses:       make(chan OrderStatus, 1000),
		mdSubscriptions:     make(map[string]string),
		symbolSubscriptions: make(map[string]string),
		posSubscriptions:    make(map[string]bool),
		quoteCache:          make(map[string]*MarketData),
	}

	// Load persisted sequence numbers for all sessions
	for _, session := range gw.sessions {
		gw.loadSequenceNumbers(session)
	}

	return gw
}

// loadSequenceNumbers loads sequence numbers from file store
func (g *FIXGateway) loadSequenceNumbers(session *LPSession) {
	seqFile := filepath.Join(session.storeDir, session.ID+".seqnums")
	file, err := os.Open(seqFile)
	if err != nil {
		// File doesn't exist, start fresh
		log.Printf("[FIX] No stored sequence numbers for %s, starting fresh", session.ID)
		session.OutSeqNum = 0
		session.InSeqNum = 0
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			session.OutSeqNum, _ = strconv.Atoi(parts[0])
			session.InSeqNum, _ = strconv.Atoi(parts[1])
			log.Printf("[FIX] Loaded sequence numbers for %s: Out=%d, In=%d",
				session.ID, session.OutSeqNum, session.InSeqNum)
		}
	}
}

// saveSequenceNumbers persists sequence numbers to file store
func (g *FIXGateway) saveSequenceNumbers(session *LPSession) error {
	seqFile := filepath.Join(session.storeDir, session.ID+".seqnums")
	content := fmt.Sprintf("%d:%d", session.OutSeqNum, session.InSeqNum)
	return os.WriteFile(seqFile, []byte(content), 0644)
}

// resetSequenceNumbers resets sequence numbers to 1
func (g *FIXGateway) resetSequenceNumbers(session *LPSession) {
	session.OutSeqNum = 0
	session.InSeqNum = 0
	session.msgStoreMu.Lock()
	session.msgStore = make(map[int]string) // Clear message store
	session.msgStoreMu.Unlock()
	g.saveSequenceNumbers(session)
	log.Printf("[FIX] Reset sequence numbers for %s", session.ID)
}

// storeMessage saves a sent message for potential resend
func (g *FIXGateway) storeMessage(session *LPSession, seqNum int, msg string) {
	session.msgStoreMu.Lock()
	defer session.msgStoreMu.Unlock()
	session.msgStore[seqNum] = msg

	// Also persist to file for crash recovery (append mode)
	msgFile := filepath.Join(session.storeDir, session.ID+".msgs")
	f, err := os.OpenFile(msgFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		// Store as: seqnum|message (message with SOH replaced by ^A for readability)
		storedMsg := strings.ReplaceAll(msg, "\x01", "^A")
		f.WriteString(fmt.Sprintf("%d|%s\n", seqNum, storedMsg))
	}
}

// getStoredMessage retrieves a stored message by sequence number
func (g *FIXGateway) getStoredMessage(session *LPSession, seqNum int) (string, bool) {
	session.msgStoreMu.RLock()
	defer session.msgStoreMu.RUnlock()
	msg, ok := session.msgStore[seqNum]
	return msg, ok
}

// getNextOutSeqNum returns and increments the outgoing sequence number
func (g *FIXGateway) getNextOutSeqNum(session *LPSession) int {
	session.OutSeqNum++
	g.saveSequenceNumbers(session)
	return session.OutSeqNum
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns the environment variable as int or a default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// Connect initiates a FIX session with real TCP connection
func (g *FIXGateway) Connect(sessionID string) error {
	g.mu.Lock()
	session, ok := g.sessions[sessionID]
	if !ok {
		g.mu.Unlock()
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status == "LOGGED_IN" || session.Status == "CONNECTING" {
		g.mu.Unlock()
		return fmt.Errorf("session already %s", session.Status)
	}

	if session.UseProxy {
		log.Printf("[FIX] Connecting to %s at %s:%d via proxy %s:%d",
			session.Name, session.Host, session.Port, session.ProxyHost, session.ProxyPort)
	} else {
		log.Printf("[FIX] Connecting to %s at %s:%d", session.Name, session.Host, session.Port)
	}
	session.Status = "CONNECTING"
	g.mu.Unlock()

	// Start connection in goroutine
	go g.connectSession(session)

	return nil
}

// connectSession handles the actual TCP connection and FIX logon
func (g *FIXGateway) connectSession(session *LPSession) {
	var conn net.Conn
	var err error

	if session.UseProxy {
		// Connect via HTTP CONNECT proxy
		conn, err = g.dialViaHTTPProxy(session)
	} else {
		// Direct connection
		addr := fmt.Sprintf("%s:%d", session.Host, session.Port)
		conn, err = net.DialTimeout("tcp", addr, 10*time.Second)
	}

	if err != nil {
		log.Printf("[FIX] Failed to connect to %s: %v", session.Name, err)
		g.mu.Lock()
		session.Status = "DISCONNECTED"
		g.mu.Unlock()
		return
	}

	// Wrap with TLS if SSL is enabled
	if session.SSL {
		log.Printf("[FIX] Upgrading connection to TLS for %s", session.Name)
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Often needed for FIX servers with IP addresses
			ServerName:         session.Host,
		}
		tlsConn := tls.Client(conn, tlsConfig)

		// Perform TLS handshake with timeout
		tlsConn.SetDeadline(time.Now().Add(10 * time.Second))
		if err := tlsConn.Handshake(); err != nil {
			log.Printf("[FIX] TLS handshake failed for %s: %v", session.Name, err)
			conn.Close()
			g.mu.Lock()
			session.Status = "DISCONNECTED"
			g.mu.Unlock()
			return
		}
		tlsConn.SetDeadline(time.Time{}) // Clear deadline
		conn = tlsConn
		log.Printf("[FIX] TLS handshake successful for %s", session.Name)
	}

	// Apply TCP optimizations for low-latency trading
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		// Disable Nagle's algorithm for immediate packet transmission
		tcpConn.SetNoDelay(true)
		// Increase socket buffer sizes for high-throughput market data
		tcpConn.SetReadBuffer(131072) // 128KB read buffer
		tcpConn.SetWriteBuffer(65536) // 64KB write buffer
		// Enable TCP keep-alive for connection health monitoring
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		log.Printf("[FIX] TCP optimizations applied (NoDelay=true, ReadBuf=128KB, WriteBuf=64KB)")
	}

	g.mu.Lock()
	session.conn = conn
	session.Status = "CONNECTED"
	g.mu.Unlock()
	log.Printf("[FIX] TCP connected to %s", session.Name)

	// Send FIX 4.4 Logon message
	if err := g.sendLogon(session); err != nil {
		log.Printf("[FIX] Logon failed for %s: %v", session.Name, err)
		conn.Close()
		g.mu.Lock()
		session.Status = "DISCONNECTED"
		session.conn = nil
		g.mu.Unlock()
		return
	}

	g.mu.Lock()
	session.Status = "LOGGED_IN"
	session.LastHeartbeat = time.Now()
	g.mu.Unlock()
	log.Printf("[FIX] Logged in to %s", session.Name)

	// Start heartbeat and message reading goroutines
	go g.heartbeatLoop(session)
	go g.readMessages(session)
}

// dialViaHTTPProxy connects to the target through an HTTP CONNECT proxy
func (g *FIXGateway) dialViaHTTPProxy(session *LPSession) (net.Conn, error) {
	proxyAddr := fmt.Sprintf("%s:%d", session.ProxyHost, session.ProxyPort)
	targetAddr := fmt.Sprintf("%s:%d", session.Host, session.Port)

	log.Printf("[FIX] Connecting to proxy %s", proxyAddr)

	// Connect to proxy server
	conn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %v", err)
	}

	// Try SOCKS5 protocol first
	log.Printf("[FIX] Attempting SOCKS5 handshake to tunnel to %s", targetAddr)
	socks5Conn, socks5Err := g.attemptSocks5(conn, session)
	if socks5Err == nil {
		log.Printf("[FIX] SOCKS5 tunnel established to %s", targetAddr)
		return socks5Conn, nil
	}
	log.Printf("[FIX] SOCKS5 failed (%v), falling back to HTTP CONNECT", socks5Err)

	// Reconnect for HTTP CONNECT (SOCKS5 attempt may have corrupted connection)
	conn.Close()
	conn, err = net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to reconnect to proxy: %v", err)
	}

	// Build HTTP CONNECT request with Basic Auth
	auth := base64Encode(session.ProxyUsername + ":" + session.ProxyPassword)
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Proxy-Authorization: Basic %s\r\n"+
		"User-Agent: RTX-FIX-Gateway/1.0\r\n"+
		"\r\n",
		targetAddr, targetAddr, auth)

	log.Printf("[FIX] Sending HTTP CONNECT request to tunnel to %s", targetAddr)

	// Send CONNECT request
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := conn.Write([]byte(connectReq)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send CONNECT request: %v", err)
	}

	// Read proxy response
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read proxy response: %v", err)
	}

	respStr := string(response[:n])
	log.Printf("[FIX] Proxy response: %s", respStr[:min(len(respStr), 100)])

	// Check for successful connection (HTTP/1.x 200)
	if len(respStr) < 12 || respStr[9:12] != "200" {
		conn.Close()
		return nil, fmt.Errorf("proxy connection failed: %s", respStr)
	}

	// Clear deadlines for ongoing FIX communication
	conn.SetDeadline(time.Time{})

	log.Printf("[FIX] HTTP tunnel established through proxy to %s", targetAddr)
	return conn, nil
}

// attemptSocks5 tries SOCKS5 protocol with username/password auth
func (g *FIXGateway) attemptSocks5(conn net.Conn, session *LPSession) (net.Conn, error) {
	// SOCKS5 greeting: version 5, 1 auth method (username/password = 0x02)
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte{0x05, 0x01, 0x02})
	if err != nil {
		return nil, fmt.Errorf("socks5 greeting failed: %v", err)
	}

	// Read server's auth method selection
	resp := make([]byte, 2)
	_, err = conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("socks5 auth selection read failed: %v", err)
	}
	if resp[0] != 0x05 || resp[1] != 0x02 {
		return nil, fmt.Errorf("socks5 auth method not supported: %v", resp)
	}

	// Send username/password auth
	user := session.ProxyUsername
	pass := session.ProxyPassword
	authReq := []byte{0x01, byte(len(user))}
	authReq = append(authReq, []byte(user)...)
	authReq = append(authReq, byte(len(pass)))
	authReq = append(authReq, []byte(pass)...)
	_, err = conn.Write(authReq)
	if err != nil {
		return nil, fmt.Errorf("socks5 auth send failed: %v", err)
	}

	// Read auth response
	authResp := make([]byte, 2)
	_, err = conn.Read(authResp)
	if err != nil {
		return nil, fmt.Errorf("socks5 auth response failed: %v", err)
	}
	if authResp[1] != 0x00 {
		return nil, fmt.Errorf("socks5 auth rejected: status %d", authResp[1])
	}

	// Send connect request
	// CMD=CONNECT(0x01), RSV=0, ATYP=IPv4(0x01)
	ip := net.ParseIP(session.Host).To4()
	if ip == nil {
		return nil, fmt.Errorf("invalid IPv4 address: %s", session.Host)
	}
	port := session.Port
	connectReq := []byte{0x05, 0x01, 0x00, 0x01}
	connectReq = append(connectReq, ip...)
	connectReq = append(connectReq, byte(port>>8), byte(port&0xff))
	_, err = conn.Write(connectReq)
	if err != nil {
		return nil, fmt.Errorf("socks5 connect send failed: %v", err)
	}

	// Read connect response
	connectResp := make([]byte, 10)
	_, err = conn.Read(connectResp)
	if err != nil {
		return nil, fmt.Errorf("socks5 connect response failed: %v", err)
	}
	if connectResp[1] != 0x00 {
		return nil, fmt.Errorf("socks5 connect rejected: status %d", connectResp[1])
	}

	conn.SetDeadline(time.Time{})
	return conn, nil
}

// base64Encode encodes a string to base64
func base64Encode(s string) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	result := ""
	padding := ""

	for len(s)%3 != 0 {
		padding += "="
		s += "\x00"
	}

	for i := 0; i < len(s); i += 3 {
		n := (int(s[i]) << 16) | (int(s[i+1]) << 8) | int(s[i+2])
		result += string(base64Chars[(n>>18)&0x3F])
		result += string(base64Chars[(n>>12)&0x3F])
		result += string(base64Chars[(n>>6)&0x3F])
		result += string(base64Chars[n&0x3F])
	}

	if len(padding) > 0 {
		result = result[:len(result)-len(padding)] + padding
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// sendLogon sends a FIX 4.4 Logon message (MsgType=A)
func (g *FIXGateway) sendLogon(session *LPSession) error {
	// Handle sequence number reset if configured
	if session.ResetSeqNumFlag {
		g.resetSequenceNumbers(session)
		log.Printf("[FIX] Sequence numbers reset for %s (ResetSeqNumFlag=Y)", session.ID)
	}

	// Get next sequence number (increments and persists)
	msgSeqNum := g.getNextOutSeqNum(session)
	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	// Build body first (excluding BeginString, BodyLength, and Checksum)
	// Tag 35=A (Logon), Tag 98=0 (No encryption), Tag 108=30 (HeartBtInt)
	// Tag 141=Y (ResetSeqNumFlag) if resetting, Tag 553=Username, Tag 554=Password
	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+ // SenderCompID
		"56=%s\x01"+ // TargetCompID
		"34=%d\x01"+ // MsgSeqNum
		"52=%s\x01"+ // SendingTime
		"98=0\x01"+ // EncryptMethod (None)
		"108=30\x01", // HeartBtInt (30 seconds)
		MsgTypeLogon,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
	)

	// NOTE: ResetSeqNumFlag (141=Y) is NOT sent in Logon message
	// The YOFX/T4B server does not accept this field and will not respond
	// Sequence number reset is handled internally by resetSequenceNumbers() above
	// if session.ResetSeqNumFlag {
	// 	body += "141=Y\x01" // ResetSeqNumFlag - DISABLED: Server doesn't support
	// }

	// Add credentials
	if session.Username != "" {
		body += fmt.Sprintf("553=%s\x01", session.Username) // Username
	}
	if session.Password != "" {
		body += fmt.Sprintf("554=%s\x01", session.Password) // Password
	}

	// Build complete message
	fullMsg := g.buildMessage(session, body)

	// Store the message for potential resend (excluding Logon per FIX spec - admin messages may be gap-filled)
	g.storeMessage(session, msgSeqNum, fullMsg)

	log.Printf("[FIX] Sending Logon to %s: SenderCompID=%s, TargetCompID=%s, User=%s, SeqNum=%d, ResetSeqNum=%v",
		session.Name, session.SenderCompID, session.TargetCompID, session.Username, msgSeqNum, session.ResetSeqNumFlag)

	// Set write deadline
	session.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := session.conn.Write([]byte(fullMsg))
	if err != nil {
		return fmt.Errorf("failed to send logon: %v", err)
	}

	// Wait for Logon response with timeout
	session.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buffer := make([]byte, 4096)
	n, err := session.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read logon response: %v", err)
	}

	response := string(buffer[:n])
	log.Printf("[FIX] Received from %s: %s", session.Name, g.formatFIXMessage(response))

	// Parse and validate incoming sequence number
	if err := g.validateAndUpdateInSeq(session, response); err != nil {
		log.Printf("[FIX] Warning: %v", err)
		// Don't fail on logon response seq validation - counterparty may have reset
	}

	// Check if we got a Logon response (35=A) or Logout/Reject
	if !g.containsTag(response, "35", MsgTypeLogon) {
		if g.containsTag(response, "35", MsgTypeLogout) {
			return fmt.Errorf("received Logout: %s", g.extractTag(response, "58"))
		}
		if g.containsTag(response, "35", MsgTypeReject) {
			return fmt.Errorf("received Reject: %s", g.extractTag(response, "58"))
		}
		return fmt.Errorf("unexpected response type: %s", g.extractTag(response, "35"))
	}

	// Check if counterparty is also resetting sequence numbers
	if g.containsTag(response, "141", "Y") {
		log.Printf("[FIX] Counterparty also reset sequence numbers")
		session.InSeqNum = 1 // They reset, so expect seq 1
	}

	return nil
}

// buildMessage constructs a complete FIX message with header and checksum
func (g *FIXGateway) buildMessage(session *LPSession, body string) string {
	header := fmt.Sprintf("8=%s\x019=%d\x01", session.BeginString, len(body))
	msgWithoutChecksum := header + body
	checksum := g.calculateChecksum(msgWithoutChecksum)
	return msgWithoutChecksum + fmt.Sprintf("10=%03d\x01", checksum)
}

// validateAndUpdateInSeq validates incoming sequence number and updates expected
func (g *FIXGateway) validateAndUpdateInSeq(session *LPSession, msg string) error {
	inSeqStr := g.extractTag(msg, "34")
	if inSeqStr == "" {
		return fmt.Errorf("missing MsgSeqNum (34) in message")
	}

	inSeq, err := strconv.Atoi(inSeqStr)
	if err != nil {
		return fmt.Errorf("invalid MsgSeqNum: %s", inSeqStr)
	}

	expectedSeq := session.InSeqNum + 1

	if inSeq > expectedSeq {
		// Gap detected - we missed messages
		log.Printf("[FIX] Sequence gap detected for %s: expected %d, got %d",
			session.ID, expectedSeq, inSeq)
		// Send ResendRequest for missing messages
		go g.sendResendRequest(session, expectedSeq, inSeq-1)
	} else if inSeq < expectedSeq {
		// Check for PossDupFlag
		if !g.containsTag(msg, "43", "Y") {
			return fmt.Errorf("sequence number too low: expected %d, got %d (no PossDupFlag)",
				expectedSeq, inSeq)
		}
		// It's a resend, accept it
		log.Printf("[FIX] Received resent message %d (PossDupFlag=Y)", inSeq)
		return nil
	}

	// Update expected incoming sequence number
	session.InSeqNum = inSeq
	g.saveSequenceNumbers(session)
	return nil
}

// calculateChecksum calculates FIX message checksum
func (g *FIXGateway) calculateChecksum(msg string) int {
	sum := 0
	for i := 0; i < len(msg); i++ {
		sum += int(msg[i])
	}
	return sum % 256
}

// validateMessage validates checksum and body length of incoming FIX message
// Returns error if validation fails, nil if valid
func (g *FIXGateway) validateMessage(msg string) error {
	// Validate checksum (tag 10)
	if err := g.validateChecksum(msg); err != nil {
		return fmt.Errorf("checksum validation failed: %v", err)
	}

	// Validate body length (tag 9)
	if err := g.validateBodyLength(msg); err != nil {
		return fmt.Errorf("body length validation failed: %v", err)
	}

	return nil
}

// validateChecksum validates the FIX message checksum
func (g *FIXGateway) validateChecksum(msg string) error {
	// Find checksum tag position
	checksumIdx := strings.LastIndex(msg, "10=")
	if checksumIdx == -1 {
		return fmt.Errorf("checksum tag (10) not found")
	}

	// Extract declared checksum value
	declaredChecksumStr := g.extractTag(msg, "10")
	if len(declaredChecksumStr) != 3 {
		return fmt.Errorf("invalid checksum format: %s", declaredChecksumStr)
	}

	declaredChecksum, err := strconv.Atoi(declaredChecksumStr)
	if err != nil {
		return fmt.Errorf("invalid checksum value: %s", declaredChecksumStr)
	}

	// Calculate checksum of message body (everything before "10=")
	msgWithoutChecksum := msg[:checksumIdx]
	calculatedChecksum := g.calculateChecksum(msgWithoutChecksum)

	if calculatedChecksum != declaredChecksum {
		return fmt.Errorf("checksum mismatch: declared=%03d, calculated=%03d", declaredChecksum, calculatedChecksum)
	}

	return nil
}

// validateBodyLength validates the FIX message body length
func (g *FIXGateway) validateBodyLength(msg string) error {
	// Extract declared body length
	bodyLengthStr := g.extractTag(msg, "9")
	if bodyLengthStr == "" {
		return fmt.Errorf("body length tag (9) not found")
	}

	declaredLength, err := strconv.Atoi(bodyLengthStr)
	if err != nil {
		return fmt.Errorf("invalid body length value: %s", bodyLengthStr)
	}

	// Find start of body (after "9=XXX\x01")
	bodyStartTag := "9=" + bodyLengthStr + "\x01"
	bodyStartIdx := strings.Index(msg, bodyStartTag)
	if bodyStartIdx == -1 {
		return fmt.Errorf("could not find body start")
	}
	bodyStartIdx += len(bodyStartTag)

	// Find end of body (before "10=")
	checksumIdx := strings.LastIndex(msg, "10=")
	if checksumIdx == -1 {
		return fmt.Errorf("checksum tag not found for body length calculation")
	}

	// Calculate actual body length
	actualLength := checksumIdx - bodyStartIdx

	if actualLength != declaredLength {
		return fmt.Errorf("body length mismatch: declared=%d, actual=%d", declaredLength, actualLength)
	}

	return nil
}

// formatFIXMessage replaces SOH with | for logging
func (g *FIXGateway) formatFIXMessage(msg string) string {
	result := ""
	for _, c := range msg {
		if c == '\x01' {
			result += "|"
		} else {
			result += string(c)
		}
	}
	return result
}

// containsTag checks if a FIX message contains a specific tag=value
func (g *FIXGateway) containsTag(msg string, tag string, value string) bool {
	search := tag + "=" + value + "\x01"
	return len(msg) >= len(search) && (msg == search || len(msg) > len(search) && (msg[:len(search)] == search || msg[len(msg)-len(search):] == search || len(msg) > len(search)+1 && findSubstring(msg, search)))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// extractTag extracts a tag value from a FIX message
func (g *FIXGateway) extractTag(msg string, tag string) string {
	tagPrefix := tag + "="
	start := 0
	for i := 0; i <= len(msg)-len(tagPrefix); i++ {
		if msg[i:i+len(tagPrefix)] == tagPrefix {
			start = i + len(tagPrefix)
			break
		}
	}
	if start == 0 {
		return ""
	}
	end := start
	for end < len(msg) && msg[end] != '\x01' {
		end++
	}
	return msg[start:end]
}

// heartbeatLoop sends periodic heartbeats
func (g *FIXGateway) heartbeatLoop(session *LPSession) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		g.mu.RLock()
		if session.Status != "LOGGED_IN" || session.conn == nil {
			g.mu.RUnlock()
			return
		}
		conn := session.conn
		g.mu.RUnlock()

		// Send Heartbeat (35=0) - no TestReqID for unsolicited heartbeats
		if err := g.sendHeartbeat(session, conn, ""); err != nil {
			log.Printf("[FIX] Heartbeat failed for %s: %v", session.Name, err)
			g.Disconnect(session.ID)
			return
		}

		g.mu.Lock()
		session.LastHeartbeat = time.Now()
		g.mu.Unlock()
	}
}

// sendHeartbeat sends a FIX Heartbeat message
// testReqID should be set when responding to a TestRequest (tag 112)
func (g *FIXGateway) sendHeartbeat(session *LPSession, conn net.Conn, testReqID string) error {
	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01",
		MsgTypeHeartbeat,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
	)

	// Add TestReqID if this is a response to TestRequest
	if testReqID != "" {
		body += fmt.Sprintf("112=%s\x01", testReqID)
	}

	fullMsg := g.buildMessage(session, body)

	// Store heartbeat for potential resend
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err == nil {
		log.Printf("[FIX] Sent Heartbeat to %s: SeqNum=%d", session.Name, msgSeqNum)
	}
	return err
}

// sendTestRequest sends a TestRequest message (35=1)
func (g *FIXGateway) sendTestRequest(session *LPSession, conn net.Conn) error {
	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")
	testReqID := fmt.Sprintf("TEST%d", time.Now().UnixNano())

	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"112=%s\x01", // TestReqID
		MsgTypeTestRequest,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		testReqID,
	)

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err == nil {
		log.Printf("[FIX] Sent TestRequest to %s: SeqNum=%d, TestReqID=%s", session.Name, msgSeqNum, testReqID)
	}
	return err
}

// sendResendRequest sends a ResendRequest message (35=2)
func (g *FIXGateway) sendResendRequest(session *LPSession, beginSeqNo, endSeqNo int) error {
	g.mu.RLock()
	if session.conn == nil {
		g.mu.RUnlock()
		return fmt.Errorf("connection not available")
	}
	conn := session.conn
	g.mu.RUnlock()

	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"7=%d\x01"+ // BeginSeqNo
		"16=%d\x01", // EndSeqNo (0 = infinity)
		MsgTypeResendRequest,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		beginSeqNo,
		endSeqNo,
	)

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err == nil {
		log.Printf("[FIX] Sent ResendRequest to %s: SeqNum=%d, BeginSeqNo=%d, EndSeqNo=%d",
			session.Name, msgSeqNum, beginSeqNo, endSeqNo)
	}
	return err
}

// sendSequenceReset sends a SequenceReset-GapFill message (35=4)
func (g *FIXGateway) sendSequenceReset(session *LPSession, conn net.Conn, newSeqNo int, gapFill bool) error {
	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	gapFillFlag := "N"
	if gapFill {
		gapFillFlag = "Y"
	}

	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"123=%s\x01"+ // GapFillFlag
		"36=%d\x01", // NewSeqNo
		MsgTypeSequenceReset,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		gapFillFlag,
		newSeqNo,
	)

	// Add PossDupFlag for gap fills
	if gapFill {
		body = fmt.Sprintf("35=%s\x01"+
			"49=%s\x01"+
			"56=%s\x01"+
			"34=%d\x01"+
			"43=Y\x01"+ // PossDupFlag
			"52=%s\x01"+
			"122=%s\x01"+ // OrigSendingTime
			"123=%s\x01"+ // GapFillFlag
			"36=%d\x01", // NewSeqNo
			MsgTypeSequenceReset,
			session.SenderCompID,
			session.TargetCompID,
			msgSeqNum,
			sendingTime,
			sendingTime,
			gapFillFlag,
			newSeqNo,
		)
	}

	fullMsg := g.buildMessage(session, body)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err == nil {
		log.Printf("[FIX] Sent SequenceReset to %s: SeqNum=%d, NewSeqNo=%d, GapFill=%v",
			session.Name, msgSeqNum, newSeqNo, gapFill)
	}
	return err
}

// handleResendRequest processes an incoming ResendRequest
func (g *FIXGateway) handleResendRequest(session *LPSession, msg string) {
	beginSeqNo, _ := strconv.Atoi(g.extractTag(msg, "7"))
	endSeqNo, _ := strconv.Atoi(g.extractTag(msg, "16"))

	log.Printf("[FIX] Received ResendRequest from %s: BeginSeqNo=%d, EndSeqNo=%d",
		session.Name, beginSeqNo, endSeqNo)

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return
	}

	// If endSeqNo is 0, it means "to infinity" - use current outSeqNum
	if endSeqNo == 0 {
		endSeqNo = session.OutSeqNum
	}

	// Try to resend stored messages or send GapFill
	for seqNum := beginSeqNo; seqNum <= endSeqNo; seqNum++ {
		storedMsg, found := g.getStoredMessage(session, seqNum)
		if found {
			// Resend with PossDupFlag=Y
			resendMsg := g.addPossDupFlag(storedMsg)
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			conn.Write([]byte(resendMsg))
			log.Printf("[FIX] Resent message %d to %s", seqNum, session.Name)
		} else {
			// Message not found - send GapFill to skip it
			log.Printf("[FIX] Message %d not found, sending GapFill", seqNum)
			g.sendSequenceReset(session, conn, seqNum+1, true)
		}
	}
}

// addPossDupFlag adds PossDupFlag=Y and OrigSendingTime to a message for resend
func (g *FIXGateway) addPossDupFlag(msg string) string {
	// Extract original sending time
	origSendingTime := g.extractTag(msg, "52")
	newSendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	// Insert PossDupFlag after MsgSeqNum (34)
	// And add OrigSendingTime (122), update SendingTime (52)
	// This is a simplified implementation - production code should properly parse and rebuild

	// Replace SendingTime with new time and add PossDupFlag
	msg = strings.Replace(msg, "52="+origSendingTime, "52="+newSendingTime+"\x0143=Y\x01122="+origSendingTime, 1)

	// Recalculate checksum
	checksumPos := strings.LastIndex(msg, "10=")
	if checksumPos > 0 {
		msgWithoutChecksum := msg[:checksumPos]
		checksum := g.calculateChecksum(msgWithoutChecksum)
		msg = msgWithoutChecksum + fmt.Sprintf("10=%03d\x01", checksum)
	}

	return msg
}

// handleSequenceReset processes an incoming SequenceReset message
func (g *FIXGateway) handleSequenceReset(session *LPSession, msg string) {
	newSeqNo, _ := strconv.Atoi(g.extractTag(msg, "36"))
	gapFill := g.containsTag(msg, "123", "Y")

	log.Printf("[FIX] Received SequenceReset from %s: NewSeqNo=%d, GapFill=%v",
		session.Name, newSeqNo, gapFill)

	if gapFill {
		// GapFill - just advance our expected incoming seq
		if newSeqNo > session.InSeqNum {
			session.InSeqNum = newSeqNo - 1 // Will be incremented on next message
			g.saveSequenceNumbers(session)
		}
	} else {
		// Reset - accept the new sequence number
		session.InSeqNum = newSeqNo - 1
		g.saveSequenceNumbers(session)
	}
}

// readMessages reads incoming FIX messages
func (g *FIXGateway) readMessages(session *LPSession) {
	buffer := make([]byte, 8192)
	var partialMsg string // Buffer for incomplete messages

	for {
		g.mu.RLock()
		if session.Status != "LOGGED_IN" || session.conn == nil {
			g.mu.RUnlock()
			return
		}
		conn := session.conn
		g.mu.RUnlock()

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout is ok, just retry
			}
			log.Printf("[FIX] Read error for %s: %v", session.Name, err)
			g.Disconnect(session.ID)
			return
		}

		// Combine with any partial message from previous read
		data := partialMsg + string(buffer[:n])
		partialMsg = ""

		// Process all complete messages (may be multiple in one read)
		messages := g.splitFIXMessages(data)
		for i, msg := range messages {
			// Last message might be incomplete
			if i == len(messages)-1 && !strings.HasSuffix(msg, "\x01") {
				partialMsg = msg
				continue
			}

			g.processMessage(session, conn, msg)
		}
	}
}

// splitFIXMessages splits a buffer into individual FIX messages
func (g *FIXGateway) splitFIXMessages(data string) []string {
	var messages []string
	start := 0

	for {
		// Find start of message (8=FIX)
		msgStart := strings.Index(data[start:], "8=FIX")
		if msgStart == -1 {
			// No more complete messages
			if start < len(data) {
				messages = append(messages, data[start:])
			}
			break
		}
		msgStart += start

		// Find end of message (10=XXX\x01)
		checksumIdx := strings.Index(data[msgStart:], "10=")
		if checksumIdx == -1 {
			// Incomplete message
			messages = append(messages, data[msgStart:])
			break
		}
		checksumIdx += msgStart

		// Find the SOH after checksum
		endIdx := strings.Index(data[checksumIdx:], "\x01")
		if endIdx == -1 {
			messages = append(messages, data[msgStart:])
			break
		}
		endIdx += checksumIdx + 1

		messages = append(messages, data[msgStart:endIdx])
		start = endIdx
	}

	return messages
}

// processMessage handles a single FIX message
func (g *FIXGateway) processMessage(session *LPSession, conn net.Conn, msg string) {
	log.Printf("[FIX] Received from %s: %s", session.Name, g.formatFIXMessage(msg))

	// Validate message checksum and body length
	if err := g.validateMessage(msg); err != nil {
		log.Printf("[FIX] Message validation error for %s: %v", session.Name, err)
		// Log warning but continue - some LPs may have minor protocol deviations
	}

	// Validate and update sequence number (skip for SequenceReset which has special handling)
	msgType := g.extractTag(msg, "35")
	if msgType != MsgTypeSequenceReset {
		if err := g.validateAndUpdateInSeq(session, msg); err != nil {
			log.Printf("[FIX] Sequence error for %s: %v", session.Name, err)
			// For serious sequence errors, we may need to disconnect
			// But continue processing for now
		}
	}

	// Handle different message types
	switch msgType {
	case MsgTypeLogout: // Logout (35=5)
		text := g.extractTag(msg, "58")
		log.Printf("[FIX] Received Logout from %s: %s", session.Name, text)
		g.Disconnect(session.ID)
		return

	case MsgTypeHeartbeat: // Heartbeat (35=0)
		// Just update last heartbeat time
		g.mu.Lock()
		session.LastHeartbeat = time.Now()
		g.mu.Unlock()
		log.Printf("[FIX] Received Heartbeat from %s", session.Name)

	case MsgTypeTestRequest: // TestRequest (35=1)
		testReqID := g.extractTag(msg, "112")
		log.Printf("[FIX] Received TestRequest from %s: TestReqID=%s", session.Name, testReqID)
		// Respond with Heartbeat containing the TestReqID
		g.sendHeartbeat(session, conn, testReqID)

	case MsgTypeResendRequest: // ResendRequest (35=2)
		g.handleResendRequest(session, msg)

	case MsgTypeReject: // Reject (35=3)
		refSeqNum := g.extractTag(msg, "45")
		text := g.extractTag(msg, "58")
		log.Printf("[FIX] Received Reject from %s: RefSeqNum=%s, Text=%s", session.Name, refSeqNum, text)

	case MsgTypeSequenceReset: // SequenceReset (35=4)
		g.handleSequenceReset(session, msg)

	case MsgTypeExecutionReport: // ExecutionReport (35=8)
		g.handleExecutionReport(session, msg)

	case MsgTypeMarketDataSnapshot: // MarketDataSnapshot (35=W)
		g.handleMarketDataSnapshot(session, msg)

	case MsgTypeMarketDataIncremental: // MarketDataIncremental (35=X)
		g.handleMarketDataIncremental(session, msg)

	case MsgTypeMarketDataReject: // MarketDataReject (35=Y)
		g.handleMarketDataReject(session, msg)

	case MsgTypeOrderCancelReject: // OrderCancelReject (35=9)
		g.handleOrderCancelReject(session, msg)

	case MsgTypeRequestForPositionsAck: // RequestForPositionsAck (35=AO)
		g.handleRequestForPositionsAck(session, msg)

	case MsgTypePositionReport: // PositionReport (35=AP)
		g.handlePositionReport(session, msg)

	case MsgTypeTradeCaptureReportAck: // TradeCaptureReportAck (35=AQ)
		g.handleTradeCaptureReportAck(session, msg)

	case MsgTypeTradeCaptureReport: // TradeCaptureReport (35=AE)
		g.handleTradeCaptureReport(session, msg)

	case MsgTypeBusinessReject: // BusinessMessageReject (35=j)
		refMsgType := g.extractTag(msg, "372")
		reason := g.extractTag(msg, "380")
		text := g.extractTag(msg, "58")
		log.Printf("[FIX] BusinessReject from %s: RefMsgType=%s, Reason=%s, Text=%s", session.Name, refMsgType, reason, text)

	default:
		log.Printf("[FIX] Received message type %s from %s", msgType, session.Name)
	}
}

// handleExecutionReport processes incoming execution reports
func (g *FIXGateway) handleExecutionReport(session *LPSession, msg string) {
	report := ExecutionReport{
		OrderID:   g.extractTag(msg, "37"),
		Symbol:    g.extractTag(msg, "55"),
		Side:      g.extractTag(msg, "54"),
		LPOrderID: g.extractTag(msg, "17"),
		Text:      g.extractTag(msg, "58"),
		Timestamp: time.Now(),
	}

	execType := g.extractTag(msg, "150")
	switch execType {
	case "0":
		report.ExecType = "NEW"
	case "F", "2":
		report.ExecType = "FILLED"
	case "8":
		report.ExecType = "REJECTED"
	case "4":
		report.ExecType = "CANCELED"
	default:
		report.ExecType = execType
	}

	// Parse volume and price
	if qty := g.extractTag(msg, "32"); qty != "" {
		fmt.Sscanf(qty, "%f", &report.Volume)
	}
	if px := g.extractTag(msg, "31"); px != "" {
		fmt.Sscanf(px, "%f", &report.Price)
	}

	log.Printf("[FIX] Execution Report from %s: %s %s %s @ %.5f", session.Name, report.ExecType, report.Side, report.Symbol, report.Price)
	g.execReports <- report
}

// Disconnect closes a FIX session
func (g *FIXGateway) Disconnect(sessionID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	session, ok := g.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.conn != nil {
		// Send Logout message (35=5) with proper sequence number before closing
		session.OutSeqNum++
		msgSeqNum := session.OutSeqNum
		g.saveSequenceNumbers(session)

		sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")
		body := fmt.Sprintf("35=%s\x01"+
			"49=%s\x01"+
			"56=%s\x01"+
			"34=%d\x01"+
			"52=%s\x01",
			MsgTypeLogout,
			session.SenderCompID,
			session.TargetCompID,
			msgSeqNum,
			sendingTime,
		)

		fullMsg := g.buildMessage(session, body)

		session.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		session.conn.Write([]byte(fullMsg))
		log.Printf("[FIX] Sent Logout to %s: SeqNum=%d", session.Name, msgSeqNum)

		session.conn.Close()
		session.conn = nil
	}

	session.Status = "DISCONNECTED"
	log.Printf("[FIX] Disconnected from %s", session.Name)
	return nil
}

// SendOrder sends a NewOrderSingle (35=D) to the LP
func (g *FIXGateway) SendOrder(sessionID string, symbol string, side string, volume float64, price float64) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return "", fmt.Errorf("connection not available")
	}

	// Generate unique ClOrdID (Client Order ID)
	clOrdID := fmt.Sprintf("%s_%d", sessionID, time.Now().UnixNano())

	// Get next sequence number
	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")
	transactTime := sendingTime

	// Convert side to FIX format: "BUY"/"SELL" -> "1"/"2"
	fixSide := "1" // Buy
	if side == "SELL" || side == "2" {
		fixSide = "2" // Sell
	}

	// Build NewOrderSingle message (35=D)
	// Tag 11=ClOrdID, Tag 55=Symbol, Tag 54=Side, Tag 38=OrderQty
	// Tag 40=OrdType (1=Market, 2=Limit), Tag 44=Price (for Limit)
	// Tag 60=TransactTime, Tag 21=HandlInst (1=Auto, no intervention)
	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+ // SenderCompID
		"56=%s\x01"+ // TargetCompID
		"34=%d\x01"+ // MsgSeqNum
		"52=%s\x01"+ // SendingTime
		"11=%s\x01"+ // ClOrdID
		"55=%s\x01"+ // Symbol
		"54=%s\x01"+ // Side
		"38=%.2f\x01"+ // OrderQty (volume)
		"40=2\x01"+ // OrdType (2=Limit)
		"44=%.5f\x01"+ // Price
		"60=%s\x01"+ // TransactTime
		"21=1\x01", // HandlInst (1=Auto)
		MsgTypeNewOrderSingle,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		clOrdID,
		symbol,
		fixSide,
		volume,
		price,
		transactTime,
	)

	// Add account if specified
	if session.TradingAccount != "" {
		body += fmt.Sprintf("1=%s\x01", session.TradingAccount) // Account
	}

	fullMsg := g.buildMessage(session, body)

	// Store the order message for potential resend
	g.storeMessage(session, msgSeqNum, fullMsg)

	// Send the order
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send order: %v", err)
	}

	log.Printf("[FIX] Sent NewOrderSingle to %s: ClOrdID=%s, Symbol=%s, Side=%s, Qty=%.2f, Price=%.5f, SeqNum=%d",
		session.Name, clOrdID, symbol, side, volume, price, msgSeqNum)

	return clOrdID, nil
}

// SendMarketOrder sends a market order (OrdType=1) to the LP
func (g *FIXGateway) SendMarketOrder(sessionID string, symbol string, side string, volume float64) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return "", fmt.Errorf("connection not available")
	}

	clOrdID := fmt.Sprintf("%s_%d", sessionID, time.Now().UnixNano())

	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	fixSide := "1"
	if side == "SELL" || side == "2" {
		fixSide = "2"
	}

	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"11=%s\x01"+
		"55=%s\x01"+
		"54=%s\x01"+
		"38=%.2f\x01"+
		"40=1\x01"+ // OrdType (1=Market)
		"60=%s\x01"+
		"21=1\x01",
		MsgTypeNewOrderSingle,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		clOrdID,
		symbol,
		fixSide,
		volume,
		sendingTime,
	)

	if session.TradingAccount != "" {
		body += fmt.Sprintf("1=%s\x01", session.TradingAccount)
	}

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send market order: %v", err)
	}

	log.Printf("[FIX] Sent Market Order to %s: ClOrdID=%s, Symbol=%s, Side=%s, Qty=%.2f, SeqNum=%d",
		session.Name, clOrdID, symbol, side, volume, msgSeqNum)

	return clOrdID, nil
}

// CancelOrder sends an OrderCancelRequest (35=F) to the LP
func (g *FIXGateway) CancelOrder(sessionID string, origClOrdID string, symbol string, side string) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return "", fmt.Errorf("connection not available")
	}

	clOrdID := fmt.Sprintf("CXLREQ_%d", time.Now().UnixNano())

	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	fixSide := "1"
	if side == "SELL" || side == "2" {
		fixSide = "2"
	}

	// OrderCancelRequest (35=F)
	body := fmt.Sprintf("35=F\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"11=%s\x01"+ // ClOrdID (new)
		"41=%s\x01"+ // OrigClOrdID
		"55=%s\x01"+ // Symbol
		"54=%s\x01"+ // Side
		"60=%s\x01", // TransactTime
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		clOrdID,
		origClOrdID,
		symbol,
		fixSide,
		sendingTime,
	)

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send cancel request: %v", err)
	}

	log.Printf("[FIX] Sent OrderCancelRequest to %s: ClOrdID=%s, OrigClOrdID=%s, SeqNum=%d",
		session.Name, clOrdID, origClOrdID, msgSeqNum)

	return clOrdID, nil
}

// GetStatus returns all session statuses
func (g *FIXGateway) GetStatus() map[string]string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	status := make(map[string]string)
	for id, session := range g.sessions {
		status[id] = session.Status
	}
	return status
}

// SessionInfo contains detailed information about a FIX session
type SessionInfo struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	OutSeqNum      int       `json:"outSeqNum"`
	InSeqNum       int       `json:"inSeqNum"`
	LastHeartbeat  time.Time `json:"lastHeartbeat"`
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	SenderCompID   string    `json:"senderCompID"`
	TargetCompID   string    `json:"targetCompID"`
	TradingAccount string    `json:"tradingAccount"`
}

// GetDetailedStatus returns detailed information about all sessions
func (g *FIXGateway) GetDetailedStatus() map[string]SessionInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	info := make(map[string]SessionInfo)
	for id, session := range g.sessions {
		info[id] = SessionInfo{
			ID:             session.ID,
			Name:           session.Name,
			Status:         session.Status,
			OutSeqNum:      session.OutSeqNum,
			InSeqNum:       session.InSeqNum,
			LastHeartbeat:  session.LastHeartbeat,
			Host:           session.Host,
			Port:           session.Port,
			SenderCompID:   session.SenderCompID,
			TargetCompID:   session.TargetCompID,
			TradingAccount: session.TradingAccount,
		}
	}
	return info
}

// ResetSessionSequences manually resets sequence numbers for a session
// Use this when counterparty has reset their sequences (e.g., new trading day)
func (g *FIXGateway) ResetSessionSequences(sessionID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	session, ok := g.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.OutSeqNum = 0
	session.InSeqNum = 0
	session.msgStoreMu.Lock()
	session.msgStore = make(map[int]string)
	session.msgStoreMu.Unlock()

	// Clear stored messages file
	msgFile := filepath.Join(session.storeDir, session.ID+".msgs")
	os.Remove(msgFile)

	g.saveSequenceNumbers(session)
	log.Printf("[FIX] Manually reset sequence numbers for %s", sessionID)

	return nil
}

// SetResetSeqNumFlag sets whether to reset sequence numbers on next logon
func (g *FIXGateway) SetResetSeqNumFlag(sessionID string, reset bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	session, ok := g.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.ResetSeqNumFlag = reset
	log.Printf("[FIX] Set ResetSeqNumFlag=%v for %s", reset, sessionID)

	return nil
}

// GetExecutionReports returns the channel for execution reports
func (g *FIXGateway) GetExecutionReports() <-chan ExecutionReport {
	return g.execReports
}

// GetMarketData returns the channel for market data quotes
func (g *FIXGateway) GetMarketData() <-chan MarketData {
	return g.marketData
}

// GetMarketDataRejects returns the channel for market data subscription rejects
func (g *FIXGateway) GetMarketDataRejects() <-chan MarketDataReject {
	return g.mdRejects
}

// RequestSecurityDefinition requests security definition for a symbol (35=c)
// Some FIX servers require this before MarketDataRequest
func (g *FIXGateway) RequestSecurityDefinition(sessionID string, symbol string) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return "", fmt.Errorf("connection not available")
	}

	// Generate unique SecurityReqID
	secReqID := fmt.Sprintf("SECDEF_%s_%d", symbol, time.Now().UnixNano())

	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	// Build Security Definition Request (35=c) - FIX 4.4
	// Tag 320 = SecurityReqID (required)
	// Tag 321 = SecurityRequestType: 0=Request Security identity and specifications
	// Tag 55  = Symbol
	// Tag 167 = SecurityType: FXSPOT for forex
	// Tag 460 = Product: 4=CURRENCY
	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"320=%s\x01"+ // SecurityReqID
		"321=0\x01"+ // SecurityRequestType: 0=Request security identity
		"55=%s\x01"+ // Symbol
		"167=FXSPOT\x01"+ // SecurityType
		"460=4\x01", // Product: CURRENCY
		MsgTypeSecurityDefinitionReq,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		secReqID,
		symbol,
	)

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send security definition request: %v", err)
	}

	log.Printf("[FIX] Sent SecurityDefinitionRequest to %s: SecurityReqID=%s, Symbol=%s, SeqNum=%d",
		session.Name, secReqID, symbol, msgSeqNum)

	// Wait briefly for response before allowing MarketDataRequest
	time.Sleep(500 * time.Millisecond)

	return secReqID, nil
}

// SubscribeMarketData subscribes to real-time quotes for a symbol (35=V)
// IMPORTANT: Call RequestSecurityDefinition first for better FIX 4.4 compliance
func (g *FIXGateway) SubscribeMarketData(sessionID string, symbol string) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return "", fmt.Errorf("connection not available")
	}

	// Generate unique MDReqID
	mdReqID := fmt.Sprintf("MD_%s_%d", symbol, time.Now().UnixNano())

	// Store subscription (both directions for easy lookup)
	g.mu.Lock()
	g.mdSubscriptions[mdReqID] = symbol
	g.symbolSubscriptions[symbol] = mdReqID
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	// Build Market Data Request (35=V) - FIX 4.4 FULL format with required tags
	// YOFX Key findings:
	// - NO Account tag (1) - causes rejection
	// - NO EUR/USD format - causes "Unknown symbol" rejection
	// - NO MDUpdateType (265) - causes rejection ("Unsupported MDUpdateType '1'")
	// - ADDED SecurityType (167) - Required for FIX 4.4 compliance
	// - ADDED Product (460) - Required: 4=CURRENCY for FX pairs
	// - ADDED SecurityExchange (207) - Required: Exchange identifier
	// - ADDED Currency (15) - Quote currency (USD for EURUSD)
	// - MarketDepth (264) = 0 (Full book)

	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"262=%s\x01"+ // MDReqID
		"263=1\x01"+ // SubscriptionRequestType: 1=Snapshot+Updates (streaming)
		"264=0\x01"+ // MarketDepth: 0=Full book
		"267=2\x01"+ // NoMDEntryTypes: 2 (Bid and Offer)
		"269=0\x01"+ // MDEntryType: 0=Bid
		"269=1\x01"+ // MDEntryType: 1=Offer
		"146=1\x01"+ // NoRelatedSym: 1
		"55=%s\x01"+ // Symbol (EURUSD format)
		"460=4\x01"+ // Product: 4=CURRENCY (FX spot)
		"167=FXSPOT\x01"+ // SecurityType: FXSPOT for forex pairs
		"207=YOFX\x01"+ // SecurityExchange: YOFX exchange identifier
		"15=USD\x01", // Currency: Quote currency (second currency in pair)
		MsgTypeMarketDataRequest,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		mdReqID,
		symbol,
	)

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send market data request: %v", err)
	}

	log.Printf("[FIX] Sent MarketDataRequest to %s: MDReqID=%s, Symbol=%s, SeqNum=%d",
		session.Name, mdReqID, symbol, msgSeqNum)

	return mdReqID, nil
}

// IsSymbolSubscribed checks if a symbol is already subscribed for market data
func (g *FIXGateway) IsSymbolSubscribed(symbol string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, exists := g.symbolSubscriptions[symbol]
	return exists
}

// GetSubscribedSymbols returns a list of all subscribed symbols
func (g *FIXGateway) GetSubscribedSymbols() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	symbols := make([]string, 0, len(g.symbolSubscriptions))
	for symbol := range g.symbolSubscriptions {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// UnsubscribeMarketData unsubscribes from market data for a symbol (35=V with 263=2)
func (g *FIXGateway) UnsubscribeMarketData(sessionID string, mdReqID string) error {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	symbol := g.mdSubscriptions[mdReqID]
	g.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("connection not available")
	}

	g.mu.Lock()
	delete(g.mdSubscriptions, mdReqID)
	if symbol != "" {
		delete(g.symbolSubscriptions, symbol)
	}
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	// Build Unsubscribe request (35=V with 263=2)
	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"262=%s\x01"+ // MDReqID (same as subscribe)
		"263=2\x01"+ // SubscriptionRequestType: 2=Unsubscribe
		"264=1\x01"+
		"267=2\x01"+
		"269=0\x01"+
		"269=1\x01"+
		"146=1\x01"+
		"55=%s\x01",
		MsgTypeMarketDataRequest,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		mdReqID,
		symbol,
	)

	fullMsg := g.buildMessage(session, body)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return fmt.Errorf("failed to send unsubscribe request: %v", err)
	}

	log.Printf("[FIX] Sent MarketData Unsubscribe to %s: MDReqID=%s, Symbol=%s",
		session.Name, mdReqID, symbol)

	return nil
}

// UnsubscribeMarketDataBySymbol unsubscribes from market data by symbol name (convenience wrapper)
func (g *FIXGateway) UnsubscribeMarketDataBySymbol(sessionID string, symbol string) error {
	g.mu.RLock()
	mdReqID, exists := g.symbolSubscriptions[symbol]
	g.mu.RUnlock()

	if !exists {
		return fmt.Errorf("symbol not subscribed: %s", symbol)
	}

	return g.UnsubscribeMarketData(sessionID, mdReqID)
}

// RequestSecurityList sends a SecurityListRequest (35=x) to discover available symbols
func (g *FIXGateway) RequestSecurityList(sessionID string) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return "", fmt.Errorf("connection not available")
	}

	// Generate unique request ID
	securityReqID := fmt.Sprintf("SECLIST_%d", time.Now().UnixNano())

	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	// Build Security List Request (35=x)
	// 320 = SecurityReqID (required)
	// 559 = SecurityListRequestType: 0=Symbol, 1=SecurityType/Exchange, 2=Product, 4=All
	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"320=%s\x01"+ // SecurityReqID
		"559=4\x01", // SecurityListRequestType: 4=All Securities
		MsgTypeSecurityListRequest,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		securityReqID,
	)

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send security list request: %v", err)
	}

	log.Printf("[FIX] Sent SecurityListRequest to %s: SecurityReqID=%s, SeqNum=%d",
		session.Name, securityReqID, msgSeqNum)

	return securityReqID, nil
}

// handleMarketDataSnapshot processes incoming market data (35=W)
func (g *FIXGateway) handleMarketDataSnapshot(session *LPSession, msg string) {
	symbol := g.extractTag(msg, "55")
	mdReqID := g.extractTag(msg, "262")

	md := MarketData{
		Symbol:    symbol,
		MDReqID:   mdReqID,
		SessionID: session.ID,
		Timestamp: time.Now(),
	}

	// Parse NoMDEntries (268) and extract bid/ask
	// Format: 268=N, then N entries with 269 (type), 270 (price), 271 (size)
	numEntries := g.extractTag(msg, "268")
	if numEntries == "" {
		numEntries = "2" // Default to 2 (bid+ask)
	}

	// Find all 269/270/271 pairs
	parts := strings.Split(msg, "\x01")
	var currentType string
	for _, part := range parts {
		if strings.HasPrefix(part, "269=") {
			currentType = part[4:]
		} else if strings.HasPrefix(part, "270=") {
			price := 0.0
			fmt.Sscanf(part[4:], "%f", &price)
			if currentType == "0" { // Bid
				md.Bid = price
			} else if currentType == "1" { // Offer/Ask
				md.Ask = price
			}
		} else if strings.HasPrefix(part, "271=") {
			size := 0.0
			fmt.Sscanf(part[4:], "%f", &size)
			if currentType == "0" {
				md.BidSize = size
			} else if currentType == "1" {
				md.AskSize = size
			}
		}
	}

	// Update quote cache with snapshot data
	g.quoteCacheMu.Lock()
	g.quoteCache[symbol] = &MarketData{
		Symbol:    symbol,
		Bid:       md.Bid,
		Ask:       md.Ask,
		BidSize:   md.BidSize,
		AskSize:   md.AskSize,
		SessionID: session.ID,
		Timestamp: md.Timestamp,
	}
	g.quoteCacheMu.Unlock()

	log.Printf("[FIX] MarketData from %s: %s Bid=%.5f Ask=%.5f",
		session.Name, symbol, md.Bid, md.Ask)

	// Send to channel (non-blocking)
	select {
	case g.marketData <- md:
	default:
		log.Printf("[FIX] MarketData channel full, dropping quote for %s", symbol)
	}
}

// handleMarketDataReject processes market data request reject (35=Y)
func (g *FIXGateway) handleMarketDataReject(session *LPSession, msg string) {
	mdReqID := g.extractTag(msg, "262")
	reason := g.extractTag(msg, "281")
	text := g.extractTag(msg, "58")

	log.Printf("[FIX] MarketDataReject from %s: MDReqID=%s, Reason=%s, Text=%s",
		session.Name, mdReqID, reason, text)

	reject := MarketDataReject{
		MDReqID:   mdReqID,
		Reason:    reason,
		Text:      text,
		SessionID: session.ID,
	}

	// Remove from subscriptions
	g.mu.Lock()
	delete(g.mdSubscriptions, mdReqID)
	g.mu.Unlock()

	select {
	case g.mdRejects <- reject:
	default:
	}
}

// ============================================================================
// POSITION MANAGEMENT (35=AN, 35=AP)
// ============================================================================

// GetPositions returns the channel for position updates
func (g *FIXGateway) GetPositions() <-chan Position {
	return g.positions
}

// GetTrades returns the channel for trade history
func (g *FIXGateway) GetTrades() <-chan TradeCapture {
	return g.trades
}

// GetOrderStatuses returns the channel for order status updates
func (g *FIXGateway) GetOrderStatuses() <-chan OrderStatus {
	return g.orderStatuses
}

// RequestPositions requests current open positions (35=AN)
func (g *FIXGateway) RequestPositions(sessionID string, symbol string) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return "", fmt.Errorf("connection not available")
	}

	posReqID := fmt.Sprintf("POS_%d", time.Now().UnixNano())
	now := time.Now().UTC()

	g.mu.Lock()
	g.posSubscriptions[posReqID] = true
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := now.Format("20060102-15:04:05.000")
	clearingDate := now.Format("20060102")

	// Build RequestForPositions (35=AN)
	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"710=%s\x01"+ // PosReqID
		"724=0\x01"+ // PosReqType: 0=Positions (open)
		"263=0\x01"+ // SubscriptionRequestType: 0=Snapshot
		"1=%s\x01"+ // Account
		"581=1\x01"+ // AccountType: 1=Customer
		"715=%s\x01"+ // ClearingBusinessDate
		"60=%s\x01", // TransactTime
		MsgTypeRequestForPositions,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		posReqID,
		session.TradingAccount,
		clearingDate,
		sendingTime,
	)

	// Add symbol filter if specified
	if symbol != "" {
		body += fmt.Sprintf("55=%s\x01", symbol)
	}

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send position request: %v", err)
	}

	log.Printf("[FIX] Sent RequestForPositions to %s: PosReqID=%s, Account=%s, Symbol=%s",
		session.Name, posReqID, session.TradingAccount, symbol)

	return posReqID, nil
}

// handlePositionReport processes incoming position reports (35=AP)
func (g *FIXGateway) handlePositionReport(session *LPSession, msg string) {
	now := time.Now()

	posReqID := g.extractTag(msg, "710")
	result := g.extractTag(msg, "728")
	symbol := g.extractTag(msg, "55")
	account := g.extractTag(msg, "1")

	// Check if no positions found (728=2)
	if result == "2" {
		log.Printf("[FIX] PositionReport from %s: No positions found (PosReqID=%s)", session.Name, posReqID)
		return
	}

	pos := Position{
		Symbol:      symbol,
		PosReqID:    posReqID,
		SessionID:   session.ID,
		Account:     account,
		Timestamp:   now,
		TimestampMs: now.UnixMilli(),
	}

	// Parse position quantities
	longQty := g.extractTag(msg, "704")
	shortQty := g.extractTag(msg, "705")
	settlPrice := g.extractTag(msg, "730")

	if longQty != "" && longQty != "0" {
		fmt.Sscanf(longQty, "%f", &pos.Volume)
		pos.Side = "BUY"
	} else if shortQty != "" && shortQty != "0" {
		fmt.Sscanf(shortQty, "%f", &pos.Volume)
		pos.Side = "SELL"
	}

	if settlPrice != "" {
		fmt.Sscanf(settlPrice, "%f", &pos.EntryPrice)
	}

	log.Printf("[FIX] PositionReport from %s: %s %s %.2f @ %.5f (PosReqID=%s)",
		session.Name, pos.Side, pos.Symbol, pos.Volume, pos.EntryPrice, posReqID)

	select {
	case g.positions <- pos:
	default:
		log.Printf("[FIX] Position channel full, dropping position for %s", symbol)
	}
}

// ============================================================================
// ORDER STATUS (35=H, 35=AF)
// ============================================================================

// RequestOrderStatus requests status of a specific order (35=H)
func (g *FIXGateway) RequestOrderStatus(sessionID string, clOrdID string, symbol string, side string) error {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("connection not available")
	}

	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	fixSide := "1"
	if side == "SELL" || side == "2" {
		fixSide = "2"
	}

	// Build OrderStatusRequest (35=H)
	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"11=%s\x01"+ // ClOrdID
		"55=%s\x01"+ // Symbol
		"54=%s\x01"+ // Side
		"1=%s\x01", // Account
		MsgTypeOrderStatusRequest,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		clOrdID,
		symbol,
		fixSide,
		session.TradingAccount,
	)

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return fmt.Errorf("failed to send order status request: %v", err)
	}

	log.Printf("[FIX] Sent OrderStatusRequest to %s: ClOrdID=%s, Symbol=%s",
		session.Name, clOrdID, symbol)

	return nil
}

// RequestAllOrderStatus requests status of all orders (35=AF)
func (g *FIXGateway) RequestAllOrderStatus(sessionID string) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return "", fmt.Errorf("connection not available")
	}

	massStatusReqID := fmt.Sprintf("MASS_%d", time.Now().UnixNano())

	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")

	// Build OrderMassStatusRequest (35=AF)
	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"584=%s\x01"+ // MassStatusReqID
		"585=7\x01"+ // MassStatusReqType: 7=All orders
		"1=%s\x01", // Account
		MsgTypeOrderMassStatusReq,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		massStatusReqID,
		session.TradingAccount,
	)

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send mass status request: %v", err)
	}

	log.Printf("[FIX] Sent OrderMassStatusRequest to %s: MassStatusReqID=%s",
		session.Name, massStatusReqID)

	return massStatusReqID, nil
}

// ============================================================================
// TRADE HISTORY (35=AD, 35=AE)
// ============================================================================

// RequestTradeHistory requests historical trades (35=AD)
func (g *FIXGateway) RequestTradeHistory(sessionID string, startTime, endTime time.Time) (string, error) {
	g.mu.RLock()
	session, ok := g.sessions[sessionID]
	g.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Status != "LOGGED_IN" {
		return "", fmt.Errorf("session not logged in: %s", session.Status)
	}

	g.mu.RLock()
	conn := session.conn
	g.mu.RUnlock()

	if conn == nil {
		return "", fmt.Errorf("connection not available")
	}

	tradeReqID := fmt.Sprintf("TRADE_%d", time.Now().UnixNano())

	g.mu.Lock()
	msgSeqNum := g.getNextOutSeqNum(session)
	g.mu.Unlock()

	sendingTime := time.Now().UTC().Format("20060102-15:04:05.000")
	startTimeStr := startTime.UTC().Format("20060102-15:04:05.000")
	endTimeStr := endTime.UTC().Format("20060102-15:04:05.000")

	// Build TradeCaptureReportRequest (35=AD)
	body := fmt.Sprintf("35=%s\x01"+
		"49=%s\x01"+
		"56=%s\x01"+
		"34=%d\x01"+
		"52=%s\x01"+
		"568=%s\x01"+ // TradeRequestID
		"569=1\x01"+ // TradeRequestType: 1=Matched trades
		"263=0\x01"+ // SubscriptionRequestType: 0=Snapshot
		"1=%s\x01"+ // Account
		"580=2\x01"+ // NoDates: 2 (date range)
		"60=%s\x01"+ // TransactTime (start)
		"60=%s\x01", // TransactTime (end)
		MsgTypeTradeCaptureReportReq,
		session.SenderCompID,
		session.TargetCompID,
		msgSeqNum,
		sendingTime,
		tradeReqID,
		session.TradingAccount,
		startTimeStr,
		endTimeStr,
	)

	fullMsg := g.buildMessage(session, body)
	g.storeMessage(session, msgSeqNum, fullMsg)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(fullMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send trade history request: %v", err)
	}

	log.Printf("[FIX] Sent TradeCaptureReportRequest to %s: TradeReqID=%s, From=%s, To=%s",
		session.Name, tradeReqID, startTimeStr, endTimeStr)

	return tradeReqID, nil
}

// handleTradeCaptureReportAck handles trade capture request acknowledgment (35=AQ)
func (g *FIXGateway) handleTradeCaptureReportAck(session *LPSession, msg string) {
	tradeReqID := g.extractTag(msg, "568")
	result := g.extractTag(msg, "749")
	status := g.extractTag(msg, "750")
	totalReports := g.extractTag(msg, "748")
	text := g.extractTag(msg, "58")

	log.Printf("[FIX] TradeCaptureReportAck from %s: TradeReqID=%s, Result=%s, Status=%s, TotalReports=%s, Text=%s",
		session.Name, tradeReqID, result, status, totalReports, text)

	if result == "99" || status == "2" {
		log.Printf("[FIX] Trade request rejected: %s", text)
	}
}

// handleTradeCaptureReport handles individual trade reports (35=AE)
func (g *FIXGateway) handleTradeCaptureReport(session *LPSession, msg string) {
	now := time.Now()

	trade := TradeCapture{
		TradeID:      g.extractTag(msg, "17"), // ExecID
		OrderID:      g.extractTag(msg, "37"), // OrderID
		ClOrdID:      g.extractTag(msg, "11"), // ClOrdID
		Symbol:       g.extractTag(msg, "55"),
		Account:      g.extractTag(msg, "1"),
		TradeDate:    g.extractTag(msg, "75"),
		SessionID:    session.ID,
		TransactTime: now,
		TimestampMs:  now.UnixMilli(),
	}

	// Parse side
	side := g.extractTag(msg, "54")
	if side == "1" {
		trade.Side = "BUY"
	} else if side == "2" {
		trade.Side = "SELL"
	}

	// Parse quantity and price
	if qty := g.extractTag(msg, "32"); qty != "" {
		fmt.Sscanf(qty, "%f", &trade.Volume)
	}
	if px := g.extractTag(msg, "31"); px != "" {
		fmt.Sscanf(px, "%f", &trade.Price)
	}

	// Parse transaction time
	if transactTime := g.extractTag(msg, "60"); transactTime != "" {
		if t, err := time.Parse("20060102-15:04:05.000", transactTime); err == nil {
			trade.TransactTime = t
		} else if t, err := time.Parse("20060102-15:04:05", transactTime); err == nil {
			trade.TransactTime = t
		}
	}

	log.Printf("[FIX] TradeCaptureReport from %s: %s %s %.2f @ %.5f (TradeID=%s, Date=%s)",
		session.Name, trade.Side, trade.Symbol, trade.Volume, trade.Price, trade.TradeID, trade.TradeDate)

	select {
	case g.trades <- trade:
	default:
		log.Printf("[FIX] Trade channel full, dropping trade %s", trade.TradeID)
	}
}

// handleOrderCancelReject handles order cancel rejections (35=9)
func (g *FIXGateway) handleOrderCancelReject(session *LPSession, msg string) {
	clOrdID := g.extractTag(msg, "11")
	origClOrdID := g.extractTag(msg, "41")
	ordStatus := g.extractTag(msg, "39")
	cxlRejReason := g.extractTag(msg, "102")
	text := g.extractTag(msg, "58")

	log.Printf("[FIX] OrderCancelReject from %s: ClOrdID=%s, OrigClOrdID=%s, OrdStatus=%s, Reason=%s, Text=%s",
		session.Name, clOrdID, origClOrdID, ordStatus, cxlRejReason, text)
}

// handleRequestForPositionsAck handles position request acknowledgment (35=AO)
func (g *FIXGateway) handleRequestForPositionsAck(session *LPSession, msg string) {
	posReqID := g.extractTag(msg, "710")
	result := g.extractTag(msg, "728")
	status := g.extractTag(msg, "729")
	totalReports := g.extractTag(msg, "727")
	text := g.extractTag(msg, "58")

	log.Printf("[FIX] RequestForPositionsAck from %s: PosReqID=%s, Result=%s, Status=%s, TotalReports=%s, Text=%s",
		session.Name, posReqID, result, status, totalReports, text)

	if result == "2" {
		log.Printf("[FIX] No positions found for PosReqID=%s", posReqID)
	} else if result == "1" || status == "2" {
		log.Printf("[FIX] Position request rejected: %s", text)
	}
}

// handleMarketDataIncremental handles incremental market data updates (35=X)
// Merges with cached quotes to preserve bid when only ask updates (and vice versa)
func (g *FIXGateway) handleMarketDataIncremental(session *LPSession, msg string) {
	now := time.Now()

	// Parse incremental updates
	// Format: 268=N entries with 279 (action), 269 (type), 270 (price), 271 (size), 55 (symbol)
	parts := strings.Split(msg, "\x01")

	var currentSymbol, currentType string
	var currentAction string // 0=New, 1=Change, 2=Delete
	var currentPrice, currentSize float64

	for _, part := range parts {
		if strings.HasPrefix(part, "55=") {
			currentSymbol = part[3:]
		} else if strings.HasPrefix(part, "279=") {
			currentAction = part[4:] // 0=New, 1=Change, 2=Delete
		} else if strings.HasPrefix(part, "269=") {
			currentType = part[4:] // 0=Bid, 1=Offer
		} else if strings.HasPrefix(part, "270=") {
			fmt.Sscanf(part[4:], "%f", &currentPrice)
		} else if strings.HasPrefix(part, "271=") {
			fmt.Sscanf(part[4:], "%f", &currentSize)

			// When we have size, we have a complete entry
			// Skip delete actions (279=2)
			if currentSymbol != "" && currentPrice > 0 && currentAction != "2" {
				// Get cached quote to merge with incremental update
				g.quoteCacheMu.RLock()
				cached := g.quoteCache[currentSymbol]
				g.quoteCacheMu.RUnlock()

				// Start with cached values or zeros
				var bid, ask, bidSize, askSize float64
				if cached != nil {
					bid = cached.Bid
					ask = cached.Ask
					bidSize = cached.BidSize
					askSize = cached.AskSize
				}

				// Apply incremental update
				if currentType == "0" { // Bid update
					bid = currentPrice
					bidSize = currentSize
				} else if currentType == "1" { // Ask update
					ask = currentPrice
					askSize = currentSize
				}

				// Create merged market data
				md := MarketData{
					Symbol:    currentSymbol,
					SessionID: session.ID,
					Timestamp: now,
					Bid:       bid,
					Ask:       ask,
					BidSize:   bidSize,
					AskSize:   askSize,
				}

				// Update cache with merged data
				g.quoteCacheMu.Lock()
				g.quoteCache[currentSymbol] = &MarketData{
					Symbol:    currentSymbol,
					Bid:       bid,
					Ask:       ask,
					BidSize:   bidSize,
					AskSize:   askSize,
					SessionID: session.ID,
					Timestamp: now,
				}
				g.quoteCacheMu.Unlock()

				select {
				case g.marketData <- md:
				default:
				}
			}
		}
	}
}
