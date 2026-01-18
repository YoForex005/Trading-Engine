#!/bin/bash

# Script to add LP Analytics routes to main.go
# Usage: ./scripts/add_analytics_routes.sh

set -e

MAIN_GO="cmd/server/main.go"
BACKUP_FILE="cmd/server/main.go.backup.$(date +%Y%m%d_%H%M%S)"

echo "========================================="
echo "LP Analytics Routes Integration"
echo "========================================="
echo ""

# Check if main.go exists
if [ ! -f "$MAIN_GO" ]; then
    echo "❌ Error: $MAIN_GO not found"
    echo "   Please run this script from the backend directory"
    exit 1
fi

# Create backup
echo "📦 Creating backup: $BACKUP_FILE"
cp "$MAIN_GO" "$BACKUP_FILE"

# Check if routes already added
if grep -q "LP ANALYTICS ENDPOINTS" "$MAIN_GO"; then
    echo "⚠️  Warning: LP Analytics endpoints already exist in $MAIN_GO"
    echo "   Routes may have already been added"
    read -p "   Do you want to continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Aborted"
        exit 1
    fi
fi

# Find the insertion point (line with "FIX SESSION MANAGEMENT")
LINE_NUM=$(grep -n "===== FIX SESSION MANAGEMENT =====" "$MAIN_GO" | cut -d: -f1 | head -1)

if [ -z "$LINE_NUM" ]; then
    echo "❌ Error: Could not find insertion point in $MAIN_GO"
    echo "   Looking for: '===== FIX SESSION MANAGEMENT ====='"
    exit 1
fi

echo "📍 Found insertion point at line $LINE_NUM"

# Calculate insertion line (1 line before FIX SESSION MANAGEMENT comment)
INSERT_LINE=$((LINE_NUM - 1))

# Create the code to insert
ANALYTICS_CODE='
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
'

# Insert the code
echo "✍️  Inserting LP Analytics routes..."
{
    head -n $INSERT_LINE "$MAIN_GO"
    echo "$ANALYTICS_CODE"
    tail -n +$((INSERT_LINE + 1)) "$MAIN_GO"
} > "$MAIN_GO.tmp"

# Replace original with modified version
mv "$MAIN_GO.tmp" "$MAIN_GO"

echo "✅ Successfully added LP Analytics routes to $MAIN_GO"
echo ""
echo "========================================="
echo "Verification"
echo "========================================="
echo ""
echo "Added routes:"
echo "  • GET /api/analytics/lp/comparison"
echo "  • GET /api/analytics/lp/performance/{lp_name}"
echo "  • GET /api/analytics/lp/ranking"
echo ""
echo "Backup created at: $BACKUP_FILE"
echo ""
echo "Next steps:"
echo "  1. Review changes: git diff $MAIN_GO"
echo "  2. Test compilation: go build ./cmd/server"
echo "  3. Run tests: go test -v ./internal/api/handlers -run TestHandleLP"
echo "  4. Start server: go run cmd/server/main.go"
echo ""
echo "Expected log output:"
echo "  [Analytics] LP analytics endpoints initialized"
echo ""
echo "If something goes wrong, restore backup:"
echo "  cp $BACKUP_FILE $MAIN_GO"
echo ""
echo "========================================="

# Verify Go syntax
echo "🔍 Verifying Go syntax..."
if go fmt "$MAIN_GO" > /dev/null 2>&1; then
    echo "✅ Go syntax is valid"
else
    echo "❌ Warning: Go syntax may have issues"
    echo "   Run: go fmt $MAIN_GO"
fi

echo ""
echo "🎉 Integration complete!"
echo ""
echo "Test the API:"
echo "  curl http://localhost:7999/api/analytics/lp/comparison"
echo ""
