package monitoring

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Span represents a distributed tracing span
type Span struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Operation string
	StartTime time.Time
	EndTime   time.Time
	Tags      map[string]interface{}
	Logs      []SpanLog
	mu        sync.RWMutex
}

// SpanLog represents a log entry within a span
type SpanLog struct {
	Timestamp time.Time
	Fields    map[string]interface{}
}

// Tracer manages distributed tracing
type Tracer struct {
	serviceName string
	spans       map[string]*Span
	mu          sync.RWMutex
	logger      *Logger
}

// NewTracer creates a new tracer
func NewTracer(serviceName string) *Tracer {
	return &Tracer{
		serviceName: serviceName,
		spans:       make(map[string]*Span),
		logger:      GetLogger(),
	}
}

// StartSpan starts a new tracing span
func (t *Tracer) StartSpan(operation string) *Span {
	traceID := generateID()
	spanID := generateID()

	span := &Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Operation: operation,
		StartTime: time.Now(),
		Tags:      make(map[string]interface{}),
		Logs:      []SpanLog{},
	}

	t.mu.Lock()
	t.spans[spanID] = span
	t.mu.Unlock()

	return span
}

// StartChildSpan starts a child span
func (t *Tracer) StartChildSpan(parentSpan *Span, operation string) *Span {
	spanID := generateID()

	span := &Span{
		TraceID:   parentSpan.TraceID,
		SpanID:    spanID,
		ParentID:  parentSpan.SpanID,
		Operation: operation,
		StartTime: time.Now(),
		Tags:      make(map[string]interface{}),
		Logs:      []SpanLog{},
	}

	t.mu.Lock()
	t.spans[spanID] = span
	t.mu.Unlock()

	return span
}

// Finish finishes a span and records it
func (s *Span) Finish() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.EndTime = time.Now()

	duration := s.EndTime.Sub(s.StartTime)

	// Log span completion
	fields := map[string]interface{}{
		"trace_id":    s.TraceID,
		"span_id":     s.SpanID,
		"parent_id":   s.ParentID,
		"operation":   s.Operation,
		"duration_ms": duration.Milliseconds(),
	}

	// Add tags to fields
	for k, v := range s.Tags {
		fields[k] = v
	}

	logger := GetLogger()
	logger.WithTracing(INFO, fmt.Sprintf("Span finished: %s (%.2fms)", s.Operation, float64(duration.Milliseconds())),
		s.TraceID, s.SpanID, fields)
}

// SetTag sets a tag on the span
func (s *Span) SetTag(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tags[key] = value
}

// LogFields logs fields to the span
func (s *Span) LogFields(fields map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Logs = append(s.Logs, SpanLog{
		Timestamp: time.Now(),
		Fields:    fields,
	})
}

// Context keys for tracing
type contextKey string

const (
	spanKey contextKey = "span"
)

// ContextWithSpan adds a span to the context
func ContextWithSpan(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanKey, span)
}

// SpanFromContext extracts a span from the context
func SpanFromContext(ctx context.Context) (*Span, bool) {
	span, ok := ctx.Value(spanKey).(*Span)
	return span, ok
}

// generateID generates a random ID for trace/span
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Global tracer instance
var globalTracer = NewTracer("trading-engine")

// GetTracer returns the global tracer
func GetTracer() *Tracer {
	return globalTracer
}

// SetGlobalTracer sets the global tracer instance
func SetGlobalTracer(tracer *Tracer) {
	globalTracer = tracer
}

// TraceOrderExecution traces order execution flow
func TraceOrderExecution(orderID, symbol, orderType string) *Span {
	span := globalTracer.StartSpan("order_execution")
	span.SetTag("order_id", orderID)
	span.SetTag("symbol", symbol)
	span.SetTag("order_type", orderType)
	span.SetTag("service", "trading-engine")
	return span
}

// TraceAPIRequest traces API request handling
func TraceAPIRequest(method, endpoint string) *Span {
	span := globalTracer.StartSpan("api_request")
	span.SetTag("http.method", method)
	span.SetTag("http.endpoint", endpoint)
	span.SetTag("service", "trading-engine")
	return span
}

// TraceLPCommunication traces LP communication
func TraceLPCommunication(lpName, operation string) *Span {
	span := globalTracer.StartSpan("lp_communication")
	span.SetTag("lp_name", lpName)
	span.SetTag("operation", operation)
	span.SetTag("service", "trading-engine")
	return span
}

// TraceDBQuery traces database queries
func TraceDBQuery(operation, table string) *Span {
	span := globalTracer.StartSpan("db_query")
	span.SetTag("db.operation", operation)
	span.SetTag("db.table", table)
	span.SetTag("service", "trading-engine")
	return span
}
