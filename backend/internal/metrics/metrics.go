package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Trading platform metrics for monitoring order processing,
// active positions, and latency.
//
// Usage example:
//   metrics.OrdersProcessed.WithLabelValues("market", "BTCUSD").Inc()
//   metrics.PositionCount.WithLabelValues("ETHUSD").Set(5)
//   timer := prometheus.NewTimer(metrics.OrderLatency.WithLabelValues("market"))
//   defer timer.ObserveDuration()

var (
	// OrdersProcessed tracks total number of orders processed by the trading engine.
	// Labels: type (market, limit, stop), symbol (BTCUSD, ETHUSD, etc.)
	OrdersProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_orders_processed_total",
			Help: "Total orders processed",
		},
		[]string{"type", "symbol"},
	)

	// PositionCount tracks the number of currently active positions.
	// Labels: symbol (BTCUSD, ETHUSD, etc.)
	PositionCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "trading_positions_active",
			Help: "Number of active positions",
		},
		[]string{"symbol"},
	)

	// OrderLatency tracks the duration of order processing in seconds.
	// Labels: type (market, limit, stop)
	// Uses default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
	OrderLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "trading_order_duration_seconds",
			Help:    "Order processing duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)
)
