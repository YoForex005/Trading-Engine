#!/bin/bash
# split_main_go.sh
# Splits the monolithic main.go into logical modules

set -e

echo "=== Splitting main.go into Logical Modules ==="
echo ""

MAIN_FILE="./cmd/server/main.go"
BACKUP_DIR="./scripts/backup_$(date +%Y%m%d_%H%M%S)"

# Check if main.go exists
if [ ! -f "$MAIN_FILE" ]; then
    echo "Error: $MAIN_FILE not found"
    exit 1
fi

# Count lines
LINE_COUNT=$(wc -l < "$MAIN_FILE")
echo "Current main.go size: $LINE_COUNT lines"

if [ "$LINE_COUNT" -lt 500 ]; then
    echo "✓ main.go is already under 500 lines. No split needed."
    exit 0
fi

# Create backup
mkdir -p "$BACKUP_DIR"
cp "$MAIN_FILE" "$BACKUP_DIR/main.go.bak"
echo "✓ Backup created: $BACKUP_DIR/main.go.bak"
echo ""

# Extract routes to routes.go
echo "Extracting routes to routes.go..."

cat > "./cmd/server/routes.go" << 'EOF'
package main

import (
    "log"
    "net/http"
    "strings"

    "github.com/epic1st/rtx/backend/admin"
    "github.com/epic1st/rtx/backend/api"
    "github.com/epic1st/rtx/backend/internal/api/handlers"
    "github.com/epic1st/rtx/backend/ws"
)

// RegisterRoutes registers all HTTP handlers
func RegisterRoutes(
    hub *ws.Hub,
    authService *auth.Service,
    adminHandler *admin.Handler,
    // Add other dependencies as needed
) {
    log.Println("[ROUTES] Registering HTTP routes...")

    // Admin routes
    registerAdminRoutes(adminHandler)

    // Market data routes
    registerMarketDataRoutes()

    // WebSocket routes
    registerWebSocketRoutes(hub)

    // API routes
    registerAPIRoutes(authService)

    log.Println("[ROUTES] All routes registered")
}

func registerAdminRoutes(handler *admin.Handler) {
    // TODO: Extract all /admin/* routes here
    log.Println("[ROUTES] Admin routes registered")
}

func registerMarketDataRoutes() {
    // TODO: Extract all /market-data/* routes here
    log.Println("[ROUTES] Market data routes registered")
}

func registerWebSocketRoutes(hub *ws.Hub) {
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        ws.ServeWs(hub, w, r)
    })
    log.Println("[ROUTES] WebSocket routes registered")
}

func registerAPIRoutes(authService *auth.Service) {
    // TODO: Extract all /api/* routes here
    log.Println("[ROUTES] API routes registered")
}
EOF

echo "✓ Created routes.go with skeleton"

# Extract configuration to config.go
echo "Extracting configuration to config.go..."

cat > "./cmd/server/config.go" << 'EOF'
package main

import (
    "log"
    "os"

    "github.com/epic1st/rtx/backend/config"
    "github.com/joho/godotenv"
)

// LoadConfiguration loads and validates all configuration
func LoadConfiguration() (*config.Config, error) {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Printf("[WARN] No .env file found: %v", err)
    } else {
        log.Printf("[INIT] Loaded .env file")
    }

    // GC tuning
    if os.Getenv("GOGC") == "" {
        os.Setenv("GOGC", "50")
        log.Println("[GC] Set GOGC=50 for more frequent garbage collection")
    }
    if os.Getenv("GOMEMLIMIT") == "" {
        os.Setenv("GOMEMLIMIT", "2GiB")
        log.Println("[GC] Set GOMEMLIMIT=2GiB to prevent OOM crashes")
    }

    // Load main config
    cfg, err := config.Load()
    if err != nil {
        return nil, err
    }

    // Production validation
    if cfg.Environment == "production" {
        log.Println("[PRODUCTION] Running production environment validation...")
        validator := config.NewProductionValidator(cfg)
        if err := validator.ValidateProduction(); err != nil {
            log.Fatalf("❌ PRODUCTION VALIDATION FAILED: %v", err)
        }
        log.Println("[PRODUCTION] ✓ Production validation passed")
    }

    return cfg, nil
}
EOF

echo "✓ Created config.go"

# Extract dependencies to dependencies.go
echo "Extracting dependencies to dependencies.go..."

cat > "./cmd/server/dependencies.go" << 'EOF'
package main

import (
    "log"

    "github.com/epic1st/rtx/backend/admin"
    "github.com/epic1st/rtx/backend/auth"
    "github.com/epic1st/rtx/backend/config"
    "github.com/epic1st/rtx/backend/tickstore"
    "github.com/epic1st/rtx/backend/ws"
)

// Dependencies holds all initialized services
type Dependencies struct {
    Config       *config.Config
    AuthService  *auth.Service
    AdminHandler *admin.Handler
    TickStore    *tickstore.Service
    Hub          *ws.Hub
    // Add other dependencies as needed
}

// InitializeDependencies creates and wires all dependencies
func InitializeDependencies(cfg *config.Config) (*Dependencies, error) {
    log.Println("[INIT] Initializing dependencies...")

    deps := &Dependencies{
        Config: cfg,
    }

    // Initialize services in dependency order
    // TODO: Extract all initialization logic here

    log.Println("[INIT] All dependencies initialized")
    return deps, nil
}

// Cleanup performs graceful shutdown of all dependencies
func (d *Dependencies) Cleanup() {
    log.Println("[SHUTDOWN] Cleaning up dependencies...")
    // TODO: Add cleanup logic
}
EOF

echo "✓ Created dependencies.go"

# Create new minimal main.go
echo "Creating new minimal main.go..."

cat > "./cmd/server/main.go.new" << 'EOF'
package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // Load configuration
    cfg, err := LoadConfiguration()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Initialize dependencies
    deps, err := InitializeDependencies(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize dependencies: %v", err)
    }
    defer deps.Cleanup()

    // Register routes
    RegisterRoutes(deps.Hub, deps.AuthService, deps.AdminHandler)

    // Start server
    port := cfg.ServerPort
    if port == "" {
        port = "7999"
    }

    log.Printf("[SERVER] Starting on port %s...", port)

    // Graceful shutdown
    go func() {
        if err := http.ListenAndServe(":"+port, nil); err != nil {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan

    log.Println("[SERVER] Shutting down gracefully...")
}
EOF

echo "✓ Created main.go.new (minimal entry point)"
echo ""

echo "=== Manual Steps Required ==="
echo ""
echo "1. Review the generated files:"
echo "   - cmd/server/routes.go"
echo "   - cmd/server/config.go"
echo "   - cmd/server/dependencies.go"
echo "   - cmd/server/main.go.new"
echo ""
echo "2. Extract route registrations from old main.go to routes.go"
echo "   - All http.HandleFunc() calls"
echo "   - Group by domain (admin, api, ws, market-data)"
echo ""
echo "3. Extract initialization logic to dependencies.go"
echo "   - All New*() constructor calls"
echo "   - Service wiring"
echo ""
echo "4. Test the new structure:"
echo "   go build ./cmd/server/"
echo ""
echo "5. If build succeeds, replace old main.go:"
echo "   mv cmd/server/main.go.new cmd/server/main.go"
echo ""
echo "Backup location: $BACKUP_DIR"
