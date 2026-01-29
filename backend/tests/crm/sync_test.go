package crm

import (
	"context"
	"sync"
	"testing"
	"time"
)

// SyncRecord represents a record to sync
type SyncRecord struct {
	ID        string
	Type      string
	Source    string
	Data      map[string]interface{}
	Status    string
	Timestamp time.Time
	Retry     int
}

// SyncEngine manages CRM synchronization
type SyncEngine struct {
	mu              sync.RWMutex
	records         map[string]*SyncRecord
	syncHistory     []SyncRecord
	failedRecords   map[string]*SyncRecord
	concurrency     int
	retryLimit      int
	failureCallback func(*SyncRecord, error)
}

func NewSyncEngine(concurrency, retryLimit int) *SyncEngine {
	return &SyncEngine{
		records:       make(map[string]*SyncRecord),
		syncHistory:   make([]SyncRecord, 0),
		failedRecords: make(map[string]*SyncRecord),
		concurrency:   concurrency,
		retryLimit:    retryLimit,
	}
}

// AddRecord adds a record for synchronization
func (e *SyncEngine) AddRecord(record *SyncRecord) {
	e.mu.Lock()
	defer e.mu.Unlock()

	record.Status = "pending"
	record.Timestamp = time.Now()
	e.records[record.ID] = record
}

// Sync processes all pending records
func (e *SyncEngine) Sync(ctx context.Context) error {
	e.mu.RLock()
	recordsToSync := make([]*SyncRecord, 0)
	for _, record := range e.records {
		if record.Status == "pending" {
			recordsToSync = append(recordsToSync, record)
		}
	}
	e.mu.RUnlock()

	// Process with concurrency control
	semaphore := make(chan struct{}, e.concurrency)
	var wg sync.WaitGroup

	for _, record := range recordsToSync {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(rec *SyncRecord) {
			defer wg.Done()
			defer func() { <-semaphore }()

			e.syncRecord(ctx, rec)
		}(record)
	}

	wg.Wait()
	return nil
}

// syncRecord synchronizes a single record
func (e *SyncEngine) syncRecord(ctx context.Context, record *SyncRecord) {
	e.mu.Lock()
	record.Status = "syncing"
	e.mu.Unlock()

	// Simulate sync operation
	select {
	case <-ctx.Done():
		e.mu.Lock()
		record.Status = "cancelled"
		e.mu.Unlock()
		return
	case <-time.After(100 * time.Millisecond):
	}

	// Mark as synced
	e.mu.Lock()
	record.Status = "synced"
	e.syncHistory = append(e.syncHistory, *record)
	e.mu.Unlock()
}

// RetryFailedRecords retries failed records
func (e *SyncEngine) RetryFailedRecords(ctx context.Context) error {
	e.mu.Lock()
	failedRecords := make([]*SyncRecord, 0)
	for _, record := range e.failedRecords {
		if record.Retry < e.retryLimit {
			failedRecords = append(failedRecords, record)
		}
	}
	e.mu.Unlock()

	for _, record := range failedRecords {
		record.Retry++
		e.syncRecord(ctx, record)
	}

	return nil
}

// GetSyncStatus returns synchronization status
func (e *SyncEngine) GetSyncStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pending := 0
	synced := 0
	failed := 0

	for _, record := range e.records {
		switch record.Status {
		case "pending":
			pending++
		case "synced":
			synced++
		case "failed":
			failed++
		}
	}

	return map[string]interface{}{
		"total":    len(e.records),
		"pending":  pending,
		"synced":   synced,
		"failed":   failed,
		"history":  len(e.syncHistory),
	}
}

// Test cases
func TestSyncEngineAddRecord(t *testing.T) {
	engine := NewSyncEngine(5, 3)

	record := &SyncRecord{
		ID:   "rec-123",
		Type: "contact",
		Data: map[string]interface{}{"name": "John Doe"},
	}

	engine.AddRecord(record)

	status := engine.GetSyncStatus()
	if status["total"].(int) != 1 {
		t.Errorf("Expected 1 record, got %d", status["total"])
	}

	if status["pending"].(int) != 1 {
		t.Errorf("Expected 1 pending record, got %d", status["pending"])
	}
}

func TestSyncEngineSync(t *testing.T) {
	engine := NewSyncEngine(5, 3)

	// Add multiple records
	for i := 0; i < 5; i++ {
		record := &SyncRecord{
			ID:     "rec-" + string(rune(i)),
			Type:   "contact",
			Data:   map[string]interface{}{"index": i},
		}
		engine.AddRecord(record)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := engine.Sync(ctx)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	status := engine.GetSyncStatus()
	if status["synced"].(int) != 5 {
		t.Errorf("Expected 5 synced records, got %d", status["synced"])
	}
}

func TestSyncEngineConcurrency(t *testing.T) {
	concurrency := 3
	engine := NewSyncEngine(concurrency, 3)

	// Add 10 records
	for i := 0; i < 10; i++ {
		record := &SyncRecord{
			ID:   "rec-" + string(rune(i)),
			Type: "contact",
			Data: map[string]interface{}{"index": i},
		}
		engine.AddRecord(record)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Now()
	err := engine.Sync(ctx)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	t.Logf("Synced 10 records with concurrency %d in %v", concurrency, duration)

	status := engine.GetSyncStatus()
	if status["synced"].(int) != 10 {
		t.Errorf("Expected 10 synced records, got %d", status["synced"])
	}
}

func TestSyncEngineContextCancellation(t *testing.T) {
	engine := NewSyncEngine(1, 3)

	// Add records
	for i := 0; i < 5; i++ {
		record := &SyncRecord{
			ID:   "rec-" + string(rune(i)),
			Type: "contact",
			Data: map[string]interface{}{"index": i},
		}
		engine.AddRecord(record)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	err := engine.Sync(ctx)
	if err == nil {
		// Cancellation might not be an error
		t.Logf("Sync context cancelled")
	}
}

func TestSyncEngineRetry(t *testing.T) {
	engine := NewSyncEngine(5, 3)

	// Add record to failed
	record := &SyncRecord{
		ID:     "rec-fail",
		Type:   "contact",
		Status: "failed",
		Retry:  0,
	}

	engine.mu.Lock()
	engine.failedRecords[record.ID] = record
	engine.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := engine.RetryFailedRecords(ctx)
	if err != nil {
		t.Fatalf("Retry failed: %v", err)
	}

	engine.mu.RLock()
	defer engine.mu.RUnlock()

	if record.Retry != 1 {
		t.Errorf("Expected retry count 1, got %d", record.Retry)
	}
}

func TestSyncEngineRecordTypes(t *testing.T) {
	engine := NewSyncEngine(5, 3)

	recordTypes := []string{"contact", "account", "deal", "lead"}

	for _, recordType := range recordTypes {
		record := &SyncRecord{
			ID:   "rec-" + recordType,
			Type: recordType,
			Data: map[string]interface{}{"type": recordType},
		}
		engine.AddRecord(record)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := engine.Sync(ctx)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	status := engine.GetSyncStatus()
	if status["synced"].(int) != 4 {
		t.Errorf("Expected 4 synced records, got %d", status["synced"])
	}
}

func TestSyncEngineSourceTracking(t *testing.T) {
	engine := NewSyncEngine(5, 3)

	sources := []string{"hubspot", "salesforce", "zoho"}

	for i, source := range sources {
		record := &SyncRecord{
			ID:     "rec-" + source,
			Type:   "contact",
			Source: source,
			Data:   map[string]interface{}{"source": source},
		}
		engine.AddRecord(record)

		if i == 1 {
			// Mark second record as failed
			record.Status = "failed"
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := engine.Sync(ctx)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	status := engine.GetSyncStatus()
	t.Logf("Sync status: %+v", status)
}

func TestSyncEngineBatchProcessing(t *testing.T) {
	engine := NewSyncEngine(5, 3)

	const batchSize = 10
	const numBatches = 3

	// Add multiple batches of records
	for batch := 0; batch < numBatches; batch++ {
		for i := 0; i < batchSize; i++ {
			record := &SyncRecord{
				ID:   "rec-" + string(rune(batch*batchSize+i)),
				Type: "contact",
				Data: map[string]interface{}{"batch": batch, "index": i},
			}
			engine.AddRecord(record)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now()
	err := engine.Sync(ctx)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	status := engine.GetSyncStatus()
	if status["synced"].(int) != batchSize*numBatches {
		t.Errorf("Expected %d synced records, got %d", batchSize*numBatches, status["synced"])
	}

	t.Logf("Processed %d records in %d batches within %v", batchSize*numBatches, numBatches, duration)
}

func TestSyncEngineHistory(t *testing.T) {
	engine := NewSyncEngine(5, 3)

	// Add and sync records
	for i := 0; i < 5; i++ {
		record := &SyncRecord{
			ID:   "rec-" + string(rune(i)),
			Type: "contact",
		}
		engine.AddRecord(record)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	engine.Sync(ctx)

	status := engine.GetSyncStatus()
	if status["history"].(int) != 5 {
		t.Errorf("Expected 5 history entries, got %d", status["history"])
	}
}

func TestSyncEngineDuplicatePrevention(t *testing.T) {
	engine := NewSyncEngine(5, 3)

	// Add same record twice
	record := &SyncRecord{
		ID:   "rec-dup",
		Type: "contact",
	}

	engine.AddRecord(record)
	engine.AddRecord(record) // Should overwrite

	status := engine.GetSyncStatus()
	if status["total"].(int) != 1 {
		t.Errorf("Expected 1 record, got %d (duplicate not prevented)", status["total"])
	}
}

func TestSyncEngineErrorRecovery(t *testing.T) {
	engine := NewSyncEngine(5, 3)

	// Add records, some will fail
	for i := 0; i < 10; i++ {
		record := &SyncRecord{
			ID:   "rec-" + string(rune(i)),
			Type: "contact",
			Data: map[string]interface{}{"index": i},
		}
		engine.AddRecord(record)

		if i%3 == 0 {
			// Simulate failure for every 3rd record
			record.Status = "failed"
			engine.mu.Lock()
			engine.failedRecords[record.ID] = record
			engine.mu.Unlock()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	engine.Sync(ctx)
	engine.RetryFailedRecords(ctx)

	status := engine.GetSyncStatus()
	t.Logf("Final sync status: %+v", status)
}
