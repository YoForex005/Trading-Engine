package health

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LivenessHandler checks if the process is alive.
// Returns 200 OK if the process is running.
// This is a simple check suitable for Kubernetes liveness probes.
//
// Usage:
//
//	http.HandleFunc("/health/live", health.LivenessHandler)
func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ReadinessHandler checks if the application can serve traffic.
// Verifies database connectivity before returning success.
// Returns 200 Ready if all dependencies are healthy.
// Returns 503 Service Unavailable if database is down.
//
// Why separate from liveness?
// - Liveness stays simple: Kubernetes won't restart if DB is temporarily down
// - Readiness checks dependencies: Kubernetes stops routing traffic if DB unavailable
//
// Usage:
//
//	http.HandleFunc("/health/ready", health.ReadinessHandler(pool))
func ReadinessHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		if err := pool.Ping(context.Background()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Database unavailable"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	}
}
