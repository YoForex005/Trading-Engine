package performance

import (
	"encoding/json"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// Benchmark critical paths in the trading engine

// BenchmarkOrderExecution tests order execution performance
// Target: <50ms per order
func BenchmarkOrderExecution(b *testing.B) {
	orders := generateOrders(b.N)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		executeOrder(orders[i])
	}
}

// BenchmarkOrderExecutionParallel tests parallel order execution
func BenchmarkOrderExecutionParallel(b *testing.B) {
	orders := generateOrders(b.N)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			executeOrder(orders[i%len(orders)])
			i++
		}
	})
}

// BenchmarkWebSocketBroadcast tests WebSocket message broadcasting
// Target: <10ms latency
func BenchmarkWebSocketBroadcast(b *testing.B) {
	hub := NewWebSocketHub()
	clients := make([]*MockClient, 1000)

	// Setup 1000 mock clients
	for i := 0; i < 1000; i++ {
		clients[i] = NewMockClient()
		hub.Register(clients[i])
	}

	message := []byte(`{"type":"price","symbol":"BTCUSD","price":50000.0}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		hub.Broadcast(message)
	}

	b.StopTimer()
	hub.Shutdown()
}

// BenchmarkOrderMatching tests order matching engine performance
func BenchmarkOrderMatching(b *testing.B) {
	engine := NewMatchingEngine()
	orders := generateOrders(1000)

	// Populate order book
	for i := 0; i < 500; i++ {
		engine.AddOrder(orders[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		engine.MatchOrder(orders[500+i%500])
	}
}

// BenchmarkRiskCalculation tests risk calculation performance
func BenchmarkRiskCalculation(b *testing.B) {
	calculator := NewRiskCalculator()
	positions := generatePositions(100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		calculator.CalculateRisk(positions)
	}
}

// BenchmarkPnLCalculation tests P&L calculation performance
func BenchmarkPnLCalculation(b *testing.B) {
	positions := generatePositions(1000)
	prices := generatePrices(100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		calculatePnL(positions, prices)
	}
}

// BenchmarkDatabaseQuery tests database query performance
func BenchmarkDatabaseQuery(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		queryOrders(db, "user123")
	}
}

// BenchmarkDatabaseInsert tests database insert performance
func BenchmarkDatabaseInsert(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	orders := generateOrders(b.N)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		insertOrder(db, orders[i])
	}
}

// BenchmarkJSONMarshaling tests JSON marshaling performance
func BenchmarkJSONMarshaling(b *testing.B) {
	order := Order{
		ID:       "order123",
		Symbol:   "BTCUSD",
		Side:     "buy",
		Type:     "market",
		Quantity: 1.5,
		Price:    50000.0,
		Status:   "filled",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(order)
	}
}

// BenchmarkJSONUnmarshaling tests JSON unmarshaling performance
func BenchmarkJSONUnmarshaling(b *testing.B) {
	data := []byte(`{"id":"order123","symbol":"BTCUSD","side":"buy","type":"market","quantity":1.5,"price":50000.0,"status":"filled"}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var order Order
		_ = json.Unmarshal(data, &order)
	}
}

// BenchmarkConcurrentOrderProcessing tests high-concurrency order processing
func BenchmarkConcurrentOrderProcessing(b *testing.B) {
	numWorkers := 100
	orders := generateOrders(b.N)
	orderChan := make(chan Order, 1000)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for order := range orderChan {
				executeOrder(order)
			}
		}()
	}

	b.ResetTimer()
	b.ReportAllocs()

	go func() {
		for i := 0; i < b.N; i++ {
			orderChan <- orders[i%len(orders)]
		}
		close(orderChan)
	}()

	wg.Wait()
}

// BenchmarkMemoryAllocation tests memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		orders := make([]Order, 1000)
		for j := 0; j < 1000; j++ {
			orders[j] = Order{
				ID:       generateID(),
				Symbol:   "BTCUSD",
				Quantity: rand.Float64(),
			}
		}
		_ = orders
	}
}

// Helper types and functions

type Order struct {
	ID       string  `json:"id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Type     string  `json:"type"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Status   string  `json:"status"`
}

type Position struct {
	Symbol   string
	Quantity float64
	AvgPrice float64
}

type MockClient struct {
	id       string
	messages chan []byte
}

func NewMockClient() *MockClient {
	return &MockClient{
		id:       generateID(),
		messages: make(chan []byte, 100),
	}
}

type WebSocketHub struct {
	clients    map[*MockClient]bool
	broadcast  chan []byte
	register   chan *MockClient
	unregister chan *MockClient
	mu         sync.RWMutex
}

func NewWebSocketHub() *WebSocketHub {
	hub := &WebSocketHub{
		clients:    make(map[*MockClient]bool),
		broadcast:  make(chan []byte, 1000),
		register:   make(chan *MockClient),
		unregister: make(chan *MockClient),
	}
	go hub.run()
	return hub
}

func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			delete(h.clients, client)
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.messages <- message:
				default:
					// Client buffer full, skip
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *WebSocketHub) Register(client *MockClient) {
	h.register <- client
}

func (h *WebSocketHub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (h *WebSocketHub) Shutdown() {
	close(h.broadcast)
	close(h.register)
	close(h.unregister)
}

type MatchingEngine struct {
	orderBook map[string][]Order
	mu        sync.RWMutex
}

func NewMatchingEngine() *MatchingEngine {
	return &MatchingEngine{
		orderBook: make(map[string][]Order),
	}
}

func (e *MatchingEngine) AddOrder(order Order) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.orderBook[order.Symbol] = append(e.orderBook[order.Symbol], order)
}

func (e *MatchingEngine) MatchOrder(order Order) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	orders := e.orderBook[order.Symbol]
	for _, existing := range orders {
		if existing.Side != order.Side && existing.Price == order.Price {
			return true
		}
	}
	return false
}

type RiskCalculator struct{}

func NewRiskCalculator() *RiskCalculator {
	return &RiskCalculator{}
}

func (r *RiskCalculator) CalculateRisk(positions []Position) float64 {
	var totalExposure float64
	for _, pos := range positions {
		totalExposure += pos.Quantity * pos.AvgPrice
	}
	return totalExposure * 0.1 // 10% margin requirement
}

func generateOrders(n int) []Order {
	orders := make([]Order, n)
	symbols := []string{"BTCUSD", "ETHUSD", "XRPUSD", "BNBUSD", "SOLUSD"}
	sides := []string{"buy", "sell"}
	types := []string{"market", "limit"}

	for i := 0; i < n; i++ {
		orders[i] = Order{
			ID:       generateID(),
			Symbol:   symbols[rand.Intn(len(symbols))],
			Side:     sides[rand.Intn(len(sides))],
			Type:     types[rand.Intn(len(types))],
			Quantity: rand.Float64()*10 + 0.1,
			Price:    rand.Float64()*50000 + 1000,
			Status:   "pending",
		}
	}
	return orders
}

func generatePositions(n int) []Position {
	positions := make([]Position, n)
	symbols := []string{"BTCUSD", "ETHUSD", "XRPUSD"}

	for i := 0; i < n; i++ {
		positions[i] = Position{
			Symbol:   symbols[rand.Intn(len(symbols))],
			Quantity: rand.Float64() * 10,
			AvgPrice: rand.Float64()*50000 + 1000,
		}
	}
	return positions
}

func generatePrices(n int) map[string]float64 {
	prices := make(map[string]float64)
	symbols := []string{"BTCUSD", "ETHUSD", "XRPUSD"}

	for _, symbol := range symbols {
		prices[symbol] = rand.Float64()*50000 + 1000
	}
	return prices
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func executeOrder(order Order) {
	// Simulate order execution
	time.Sleep(time.Microsecond * 100) // 0.1ms
}

func calculatePnL(positions []Position, prices map[string]float64) float64 {
	var totalPnL float64
	for _, pos := range positions {
		if currentPrice, ok := prices[pos.Symbol]; ok {
			totalPnL += (currentPrice - pos.AvgPrice) * pos.Quantity
		}
	}
	return totalPnL
}

// Mock database functions
type MockDB struct{}

func setupTestDB(b *testing.B) *MockDB {
	return &MockDB{}
}

func (db *MockDB) Close() error {
	return nil
}

func queryOrders(db *MockDB, userID string) []Order {
	// Simulate database query
	time.Sleep(time.Microsecond * 500) // 0.5ms
	return make([]Order, 10)
}

func insertOrder(db *MockDB, order Order) error {
	// Simulate database insert
	time.Sleep(time.Microsecond * 200) // 0.2ms
	return nil
}
