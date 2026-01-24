package fix

import (
	"bytes"
	"sync"
	"time"
)

// MessagePool provides object pooling for FIX messages to reduce GC pressure
// This is critical for high-frequency market data handling
type MessagePool struct {
	bufferPool  sync.Pool
	messagePool sync.Pool
}

// PooledBuffer is a reusable buffer for FIX message building
type PooledBuffer struct {
	Buffer *bytes.Buffer
	pool   *MessagePool
}

// PooledMessage is a reusable parsed FIX message
type PooledMessage struct {
	MsgType      string
	MsgSeqNum    int
	SenderCompID string
	TargetCompID string
	SendingTime  time.Time
	Body         map[string]string
	Raw          []byte
	pool         *MessagePool
}

// NewMessagePool creates a new message pool
func NewMessagePool() *MessagePool {
	mp := &MessagePool{}

	mp.bufferPool = sync.Pool{
		New: func() interface{} {
			return &PooledBuffer{
				Buffer: bytes.NewBuffer(make([]byte, 0, 4096)),
				pool:   mp,
			}
		},
	}

	mp.messagePool = sync.Pool{
		New: func() interface{} {
			return &PooledMessage{
				Body: make(map[string]string, 32),
				pool: mp,
			}
		},
	}

	return mp
}

// GetBuffer gets a buffer from the pool
func (mp *MessagePool) GetBuffer() *PooledBuffer {
	pb := mp.bufferPool.Get().(*PooledBuffer)
	pb.Buffer.Reset()
	return pb
}

// PutBuffer returns a buffer to the pool
func (mp *MessagePool) PutBuffer(pb *PooledBuffer) {
	if pb != nil {
		pb.Buffer.Reset()
		mp.bufferPool.Put(pb)
	}
}

// GetMessage gets a message from the pool
func (mp *MessagePool) GetMessage() *PooledMessage {
	pm := mp.messagePool.Get().(*PooledMessage)
	pm.Reset()
	return pm
}

// PutMessage returns a message to the pool
func (mp *MessagePool) PutMessage(pm *PooledMessage) {
	if pm != nil {
		mp.messagePool.Put(pm)
	}
}

// Reset clears the message for reuse
func (pm *PooledMessage) Reset() {
	pm.MsgType = ""
	pm.MsgSeqNum = 0
	pm.SenderCompID = ""
	pm.TargetCompID = ""
	pm.SendingTime = time.Time{}
	pm.Raw = pm.Raw[:0]
	for k := range pm.Body {
		delete(pm.Body, k)
	}
}

// Release returns the message to its pool
func (pm *PooledMessage) Release() {
	if pm.pool != nil {
		pm.pool.PutMessage(pm)
	}
}

// Release returns the buffer to its pool
func (pb *PooledBuffer) Release() {
	if pb.pool != nil {
		pb.pool.PutBuffer(pb)
	}
}

// MessageStoreCleanup provides bounded message storage with cleanup
type MessageStoreCleanup struct {
	mu       sync.RWMutex
	messages map[int]storedMessage
	maxAge   time.Duration
	maxSize  int
}

type storedMessage struct {
	msg     string
	storedAt time.Time
}

// NewMessageStoreCleanup creates a bounded message store
func NewMessageStoreCleanup(maxSize int, maxAge time.Duration) *MessageStoreCleanup {
	msc := &MessageStoreCleanup{
		messages: make(map[int]storedMessage),
		maxAge:   maxAge,
		maxSize:  maxSize,
	}

	// Start cleanup goroutine
	go msc.periodicCleanup()

	return msc
}

// Store adds a message to the store
func (msc *MessageStoreCleanup) Store(seqNum int, msg string) {
	msc.mu.Lock()
	defer msc.mu.Unlock()

	msc.messages[seqNum] = storedMessage{
		msg:      msg,
		storedAt: time.Now(),
	}

	// If over size limit, remove oldest
	if len(msc.messages) > msc.maxSize {
		msc.removeOldest()
	}
}

// Get retrieves a message by sequence number
func (msc *MessageStoreCleanup) Get(seqNum int) (string, bool) {
	msc.mu.RLock()
	defer msc.mu.RUnlock()

	sm, ok := msc.messages[seqNum]
	if !ok {
		return "", false
	}
	return sm.msg, true
}

func (msc *MessageStoreCleanup) removeOldest() {
	oldestSeq := -1
	oldestTime := time.Now()

	for seq, sm := range msc.messages {
		if sm.storedAt.Before(oldestTime) {
			oldestTime = sm.storedAt
			oldestSeq = seq
		}
	}

	if oldestSeq >= 0 {
		delete(msc.messages, oldestSeq)
	}
}

func (msc *MessageStoreCleanup) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		msc.cleanup()
	}
}

func (msc *MessageStoreCleanup) cleanup() {
	msc.mu.Lock()
	defer msc.mu.Unlock()

	now := time.Now()
	for seq, sm := range msc.messages {
		if now.Sub(sm.storedAt) > msc.maxAge {
			delete(msc.messages, seq)
		}
	}
}

// Clear removes all messages
func (msc *MessageStoreCleanup) Clear() {
	msc.mu.Lock()
	defer msc.mu.Unlock()
	msc.messages = make(map[int]storedMessage)
}

// Count returns the number of stored messages
func (msc *MessageStoreCleanup) Count() int {
	msc.mu.RLock()
	defer msc.mu.RUnlock()
	return len(msc.messages)
}
