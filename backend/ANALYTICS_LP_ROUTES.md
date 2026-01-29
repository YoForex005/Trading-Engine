# LP Analytics API Routes - Integration Instructions

## File: backend/cmd/server/main.go

Add the following code block **after line 637** (after the `})` closing the `/api/admin/lp/` handler and **before** the `// ===== FIX SESSION MANAGEMENT =====` comment):

```go
	// ===== LP ANALYTICS ENDPOINTS =====
	// Initialize analytics handler
	analyticsLPHandler, err := handlers.NewAnalyticsLPHandler()
	if err != nil {
		log.Printf("[Analytics] Failed to initialize analytics handler: %v", err)
		log.Println("[Analytics] LP analytics endpoints will not be available")
	} else {
		log.Println("[Analytics] LP analytics endpoints initialized")

		// LP Comparison endpoint
		http.HandleFunc("/api/analytics/lp/comparison", analyticsLPHandler.HandleLPComparison)

		// LP Performance detail endpoint
		http.HandleFunc("/api/analytics/lp/performance/", analyticsLPHandler.HandleLPPerformance)

		// LP Ranking endpoint
		http.HandleFunc("/api/analytics/lp/ranking", analyticsLPHandler.HandleLPRanking)

		// Cleanup on shutdown
		defer analyticsLPHandler.Close()
	}

```

## Location Context

Insert between:
```go
	})  // ← End of /api/admin/lp/ handler (line 637)

	// YOUR CODE GOES HERE

	// ===== FIX SESSION MANAGEMENT =====  // ← This should come after (line 639)
```

## Verification

After adding the routes, the server should log on startup:
```
[Analytics] LP analytics endpoints initialized
```

If database connection fails, it will log:
```
[Analytics] Failed to initialize analytics handler: <error>
[Analytics] LP analytics endpoints will not be available
```

## API Endpoints Created

1. **GET /api/analytics/lp/comparison**
   - Compare all LPs by performance metrics
   - Query params: `start_time`, `end_time`, `symbol`, `metric`

2. **GET /api/analytics/lp/performance/{lp_name}**
   - Detailed performance for single LP
   - Query params: `start_time`, `end_time`, `symbol`

3. **GET /api/analytics/lp/ranking**
   - LP ranking by specific metric
   - Query params: `start_time`, `end_time`, `metric`, `limit`
