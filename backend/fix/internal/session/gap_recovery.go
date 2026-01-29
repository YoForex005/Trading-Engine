package session

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// GapStatus represents the status of sequence gap detection
type GapStatus int

const (
	GapStatusNoGap GapStatus = iota
	GapStatusDetected
	GapStatusDuplicate
)

// SequenceGap represents a detected sequence number gap
type SequenceGap struct {
	BeginSeqNo  int
	EndSeqNo    int
	DetectedAt  time.Time
	RequestSent bool
}

// GapRecoveryManager manages sequence gap detection and recovery
type GapRecoveryManager struct {
	sessionID      string
	expectedSeqNum int
	gapTimeout     time.Duration
	maxGapSize     int

	// Gap tracking
	currentGap     *SequenceGap
	queuedMessages []QueuedMessage
	lastSeenSeqNums map[int]time.Time // For duplicate detection

	mutex sync.RWMutex
}

// QueuedMessage represents a message queued during gap recovery
type QueuedMessage struct {
	SeqNum      int
	Message     []byte
	ReceivedAt  time.Time
}

// NewGapRecoveryManager creates a new gap recovery manager
func NewGapRecoveryManager(sessionID string, expectedSeqNum int) *GapRecoveryManager {
	return &GapRecoveryManager{
		sessionID:       sessionID,
		expectedSeqNum:  expectedSeqNum,
		gapTimeout:      500 * time.Millisecond,
		maxGapSize:      1000,
		queuedMessages:  make([]QueuedMessage, 0, 100),
		lastSeenSeqNums: make(map[int]time.Time, 1000),
	}
}

// CheckMessage checks if a received message has a sequence gap
func (m *GapRecoveryManager) CheckMessage(receivedSeqNum int, possResend bool) (GapStatus, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check for duplicate
	if receivedSeqNum < m.expectedSeqNum {
		if !possResend {
			log.Printf("[GAP] [%s] Duplicate message received: SeqNum=%d, Expected=%d (without PossResend flag)",
				m.sessionID, receivedSeqNum, m.expectedSeqNum)
		}
		return GapStatusDuplicate, nil
	}

	// Check for gap
	if receivedSeqNum > m.expectedSeqNum {
		gap := receivedSeqNum - m.expectedSeqNum

		if gap > m.maxGapSize {
			return GapStatusNoGap, fmt.Errorf("gap too large: %d (max=%d)", gap, m.maxGapSize)
		}

		log.Printf("[GAP] [%s] Gap detected: Expected=%d, Received=%d, Gap=%d",
			m.sessionID, m.expectedSeqNum, receivedSeqNum, gap)

		// Create or update gap
		if m.currentGap == nil {
			m.currentGap = &SequenceGap{
				BeginSeqNo: m.expectedSeqNum,
				EndSeqNo:   receivedSeqNum - 1,
				DetectedAt: time.Now(),
				RequestSent: false,
			}
		} else {
			// Extend existing gap
			if receivedSeqNum-1 > m.currentGap.EndSeqNo {
				m.currentGap.EndSeqNo = receivedSeqNum - 1
			}
		}

		return GapStatusDetected, nil
	}

	// Expected sequence number - process normally
	m.expectedSeqNum = receivedSeqNum + 1
	m.lastSeenSeqNums[receivedSeqNum] = time.Now()

	// Clean up old sequence tracking (keep last 1000)
	if len(m.lastSeenSeqNums) > 1000 {
		m.cleanupOldSequences()
	}

	return GapStatusNoGap, nil
}

// QueueMessage queues a message received during gap recovery
func (m *GapRecoveryManager) QueueMessage(seqNum int, msg []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.queuedMessages = append(m.queuedMessages, QueuedMessage{
		SeqNum:     seqNum,
		Message:    msg,
		ReceivedAt: time.Now(),
	})

	log.Printf("[GAP] [%s] Queued message SeqNum=%d (queue size=%d)",
		m.sessionID, seqNum, len(m.queuedMessages))
}

// GetCurrentGap returns the current gap if one exists
func (m *GapRecoveryManager) GetCurrentGap() *SequenceGap {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.currentGap
}

// ShouldSendResendRequest determines if a ResendRequest should be sent
func (m *GapRecoveryManager) ShouldSendResendRequest() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.currentGap == nil {
		return false
	}

	// Wait for gap timeout before requesting resend (in case of out-of-order delivery)
	if time.Since(m.currentGap.DetectedAt) < m.gapTimeout {
		return false
	}

	// Don't send duplicate requests
	if m.currentGap.RequestSent {
		return false
	}

	return true
}

// MarkResendRequestSent marks that a ResendRequest has been sent for the current gap
func (m *GapRecoveryManager) MarkResendRequestSent() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.currentGap != nil {
		m.currentGap.RequestSent = true
		log.Printf("[GAP] [%s] ResendRequest sent for gap %d-%d",
			m.sessionID, m.currentGap.BeginSeqNo, m.currentGap.EndSeqNo)
	}
}

// FillGap marks a gap as filled when a message is received
func (m *GapRecoveryManager) FillGap(seqNum int) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.currentGap == nil {
		return false
	}

	// Check if this message fills the gap
	if seqNum >= m.currentGap.BeginSeqNo && seqNum <= m.currentGap.EndSeqNo {
		log.Printf("[GAP] [%s] Received gap fill message: SeqNum=%d",
			m.sessionID, seqNum)

		// If this fills the entire gap, clear it
		if seqNum == m.currentGap.EndSeqNo {
			log.Printf("[GAP] [%s] Gap fully filled: %d-%d",
				m.sessionID, m.currentGap.BeginSeqNo, m.currentGap.EndSeqNo)
			m.currentGap = nil
			return true
		}

		// Update gap range
		if seqNum == m.currentGap.BeginSeqNo {
			m.currentGap.BeginSeqNo++
		}

		return false
	}

	return false
}

// GetQueuedMessages returns and clears queued messages
func (m *GapRecoveryManager) GetQueuedMessages() []QueuedMessage {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	messages := m.queuedMessages
	m.queuedMessages = make([]QueuedMessage, 0, 100)

	log.Printf("[GAP] [%s] Releasing %d queued messages", m.sessionID, len(messages))
	return messages
}

// IsGapFilled checks if the current gap has been completely filled
func (m *GapRecoveryManager) IsGapFilled() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.currentGap == nil
}

// UpdateExpectedSeqNum updates the expected sequence number
func (m *GapRecoveryManager) UpdateExpectedSeqNum(seqNum int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.expectedSeqNum = seqNum
}

// GetExpectedSeqNum returns the current expected sequence number
func (m *GapRecoveryManager) GetExpectedSeqNum() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.expectedSeqNum
}

// cleanupOldSequences removes old sequence number tracking
func (m *GapRecoveryManager) cleanupOldSequences() {
	cutoff := time.Now().Add(-10 * time.Minute)
	for seqNum, timestamp := range m.lastSeenSeqNums {
		if timestamp.Before(cutoff) {
			delete(m.lastSeenSeqNums, seqNum)
		}
	}
}

// Reset resets the gap recovery manager (e.g., after reconnection)
func (m *GapRecoveryManager) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.currentGap = nil
	m.queuedMessages = make([]QueuedMessage, 0, 100)
	m.lastSeenSeqNums = make(map[int]time.Time, 1000)

	log.Printf("[GAP] [%s] Gap recovery manager reset", m.sessionID)
}

// GetStats returns gap recovery statistics
func (m *GapRecoveryManager) GetStats() GapRecoveryStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := GapRecoveryStats{
		ExpectedSeqNum: m.expectedSeqNum,
		QueuedMessages: len(m.queuedMessages),
		TrackedSeqNums: len(m.lastSeenSeqNums),
	}

	if m.currentGap != nil {
		stats.HasActiveGap = true
		stats.GapBegin = m.currentGap.BeginSeqNo
		stats.GapEnd = m.currentGap.EndSeqNo
		stats.GapSize = m.currentGap.EndSeqNo - m.currentGap.BeginSeqNo + 1
		stats.GapAge = time.Since(m.currentGap.DetectedAt)
		stats.ResendRequested = m.currentGap.RequestSent
	}

	return stats
}

// GapRecoveryStats represents gap recovery statistics
type GapRecoveryStats struct {
	ExpectedSeqNum   int
	HasActiveGap     bool
	GapBegin         int
	GapEnd           int
	GapSize          int
	GapAge           time.Duration
	ResendRequested  bool
	QueuedMessages   int
	TrackedSeqNums   int
}
