package types

import (
	"net"
	"sync"
	"time"
)

// SessionState represents the state of a FIX session
type SessionState string

const (
	SessionStateDisconnected  SessionState = "DISCONNECTED"
	SessionStateConnecting    SessionState = "CONNECTING"
	SessionStateConnected     SessionState = "CONNECTED"
	SessionStateLoggingIn     SessionState = "LOGGING_IN"
	SessionStateLoggedIn      SessionState = "LOGGED_IN"
	SessionStateLoggingOut    SessionState = "LOGGING_OUT"
	SessionStateReconnecting  SessionState = "RECONNECTING"
)

// Session represents a FIX session
type Session struct {
	// Identity
	ID           string
	Name         string
	SenderCompID string
	TargetCompID string
	BeginString  string

	// Connection details
	Host string
	Port int
	SSL  bool

	// Credentials
	Username       string
	Password       string
	TradingAccount string

	// State
	State         SessionState
	LastHeartbeat time.Time
	Connection    net.Conn

	// Sequence numbers
	OutSeqNum       int
	InSeqNum        int
	ResetSeqNumFlag bool

	// Message store
	MessageStore    map[int][]byte
	MessageStoreMu  sync.RWMutex

	// Configuration
	HeartbeatInterval time.Duration
	StoreDir          string

	// Proxy settings
	UseProxy      bool
	ProxyHost     string
	ProxyPort     int
	ProxyUsername string
	ProxyPassword string

	// Runtime
	mutex sync.RWMutex
}

// SessionInfo contains detailed session information for API responses
type SessionInfo struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	State          SessionState   `json:"state"`
	OutSeqNum      int            `json:"outSeqNum"`
	InSeqNum       int            `json:"inSeqNum"`
	LastHeartbeat  time.Time      `json:"lastHeartbeat"`
	Host           string         `json:"host"`
	Port           int            `json:"port"`
	SenderCompID   string         `json:"senderCompID"`
	TargetCompID   string         `json:"targetCompID"`
	TradingAccount string         `json:"tradingAccount"`
	Uptime         time.Duration  `json:"uptime"`
}

// TradingSchedule represents trading hours for a session
type TradingSchedule struct {
	TimeZone    *time.Location
	OpenTime    time.Time
	CloseTime   time.Time
	DaysOfWeek  []time.Weekday
	Holidays    []time.Time
}

// ReconnectionPolicy defines reconnection behavior
type ReconnectionPolicy struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	ResetOnSuccess  bool
}

// SessionConfig represents session configuration
type SessionConfig struct {
	ID                 string
	Name               string
	Enabled            bool
	Host               string
	Port               int
	SenderCompID       string
	TargetCompID       string
	Username           string
	Password           string
	TradingAccount     string
	BeginString        string
	SSL                bool
	HeartbeatInterval  time.Duration
	ResetSeqNumFlag    bool
	UseProxy           bool
	ProxyHost          string
	ProxyPort          int
	ProxyUsername      string
	ProxyPassword      string
	TradingSchedule    *TradingSchedule
	ReconnectionPolicy *ReconnectionPolicy
	RateLimits         *RateLimits
}

// RateLimits defines rate limiting configuration
type RateLimits struct {
	MaxOrdersPerSecond      int
	MaxSubscriptionsTotal   int
	MaxCancelsPerSecond     int
}

// SessionHealth represents session health status
type SessionHealth struct {
	State           SessionState  `json:"state"`
	Connected       bool          `json:"connected"`
	LoggedIn        bool          `json:"loggedIn"`
	LastHeartbeat   time.Time     `json:"lastHeartbeat"`
	SequencesInSync bool          `json:"sequencesInSync"`
	LastError       error         `json:"lastError,omitempty"`
	Uptime          time.Duration `json:"uptime"`
}

// MessageStore interface for message persistence
type MessageStore interface {
	SaveMessage(sessionID string, seqNum int, msg []byte) error
	LoadMessage(sessionID string, seqNum int) ([]byte, error)
	LoadMessages(sessionID string, beginSeq, endSeq int) ([][]byte, error)
	ClearMessages(sessionID string, beforeSeq int) error
}

// SequenceStore interface for sequence number persistence
type SequenceStore interface {
	SaveSequences(sessionID string, outSeq, inSeq int) error
	LoadSequences(sessionID string) (outSeq, inSeq int, err error)
	ResetSequences(sessionID string) error
}
