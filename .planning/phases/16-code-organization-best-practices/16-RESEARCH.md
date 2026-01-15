# Phase 16: Code Organization & Best Practices - Research

**Researched:** 2026-01-16
**Domain:** Go + React/TypeScript codebase organization, clean architecture, professional standards
**Confidence:** HIGH

<research_summary>
## Summary

Researched professional code organization standards for a Go backend + React/TypeScript frontend trading platform. The modern approach (2026) emphasizes pragmatic architecture over rigid rules: start simple, add structure only when needed, and favor clarity over dogma.

For Go backends, the community has moved away from prescriptive "Standard Go Project Layout" toward simpler, domain-driven structures. Clean/Hexagonal architecture is about keeping business logic pure and testable, not about having 20 directories. The golden rule: dependencies point inward toward the domain.

For React/TypeScript frontends, feature-based organization has become the standard for scalable applications. Components should follow single responsibility, custom hooks extract reusable logic, and the architecture enforces unidirectional code flow (shared → features → app).

Key finding: Both ecosystems prioritize **simplicity** and **separation of concerns** over complex abstraction. Tools like golangci-lint and typescript-eslint provide opinionated defaults that catch bugs without configuration overhead.

**Primary recommendation:** Adopt clean architecture principles (domain-centric, dependency inversion) with pragmatic structure. Use slog for structured logging (standard library future-proof), golangci-lint recommended preset, and feature-based React organization with custom hooks for shared logic.
</research_summary>

<standard_stack>
## Standard Stack

### Go Backend Tools

| Library/Tool | Version | Purpose | Why Standard |
|--------------|---------|---------|--------------|
| golangci-lint | 2.8.0+ (2026) | Linting and static analysis | Industry standard Go linter aggregator with 50+ linters |
| slog | stdlib (Go 1.21+) | Structured logging | Official standard library logging, future-proof |
| errors | stdlib | Error handling with wrapping | Built-in error wrapping with errors.Is/errors.As |
| testify/assert | latest | Testing assertions | Most popular Go testing library for readable assertions |
| jscpd | latest | Code duplication detection | Multi-language support (150+ languages including Go) |

### React/TypeScript Frontend Tools

| Library/Tool | Version | Purpose | Why Standard |
|--------------|---------|---------|--------------|
| typescript-eslint | latest | TypeScript linting | Official TypeScript ESLint integration |
| eslint | latest | JavaScript/TypeScript linting | Industry standard linter |
| jscpd | latest | Code duplication detection | Supports TypeScript, uses Rabin-Karp algorithm |
| vitest | latest (already in use) | Testing framework | Modern, fast, Vite-native testing |

### Code Quality Tools (Both)

| Tool | Purpose | When to Use |
|------|---------|-------------|
| jscpd | Duplicate code detection | CI/CD integration, pre-commit hooks |
| PMD CPD | Alternative duplicate detector | If jscpd doesn't fit workflow |
| dupl | Go-specific duplication tool | Go-only projects, simpler than jscpd |

**Installation:**
```bash
# Go tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Frontend tools (already using bun)
bun add -D @typescript-eslint/eslint-plugin @typescript-eslint/parser
bun add -D jscpd

# Code duplication (optional, can use jscpd for both)
go install github.com/mibk/dupl@latest
```
</standard_stack>

<architecture_patterns>
## Architecture Patterns

### Recommended Backend Project Structure (Go)

```
backend/
├── cmd/
│   └── server/           # Application entry point
│       └── main.go
├── internal/             # Private application code
│   ├── domain/          # Business logic (entities, value objects)
│   │   ├── account/
│   │   ├── position/
│   │   ├── order/
│   │   └── trade/
│   ├── ports/           # Interfaces (dependency inversion)
│   │   ├── repositories/
│   │   └── services/
│   ├── adapters/        # External implementations
│   │   ├── http/       # HTTP handlers
│   │   ├── ws/         # WebSocket handlers
│   │   ├── postgres/   # Database repositories
│   │   └── lp/         # Liquidity provider adapters
│   └── shared/         # Shared utilities
│       ├── decimal/
│       ├── logger/
│       └── errors/
├── pkg/                 # Public libraries (if needed)
└── config/             # Configuration
```

**Current vs. Recommended:**
- Current: `bbook/`, `lp/`, `auth/`, `ws/` as top-level packages (mixed concerns)
- Recommended: `internal/domain/`, `internal/ports/`, `internal/adapters/` (clean architecture)

### Pattern 1: Clean Architecture Layers (Go)

**What:** Separate domain logic from infrastructure concerns using ports (interfaces) and adapters (implementations)

**When to use:** Any business logic that needs to be testable independent of database/HTTP/external services

**Example:**
```go
// internal/domain/account/account.go (pure domain logic)
package account

type Account struct {
    ID      string
    Balance decimal.Decimal
    Equity  decimal.Decimal
}

func (a *Account) CanOpenPosition(required decimal.Decimal) bool {
    return a.Balance.GreaterThanOrEqual(required)
}

// internal/ports/repositories/account.go (interface)
package repositories

type AccountRepository interface {
    Get(ctx context.Context, id string) (*account.Account, error)
    Update(ctx context.Context, acc *account.Account) error
}

// internal/adapters/postgres/account_repo.go (implementation)
package postgres

type AccountRepository struct {
    db *pgxpool.Pool
}

func (r *AccountRepository) Get(ctx context.Context, id string) (*account.Account, error) {
    // Database implementation
}
```

### Pattern 2: Error Wrapping with Context (Go)

**What:** Use fmt.Errorf with %w to add context while preserving error type

**When to use:** Every error boundary where you want to add contextual information

**Example:**
```go
// Source: https://go.dev/blog/go1.13-errors
func GetAccount(id string) (*Account, error) {
    acc, err := repo.Get(ctx, id)
    if err != nil {
        // Wrap error with context
        return nil, fmt.Errorf("failed to get account %s: %w", id, err)
    }
    return acc, nil
}

// Checking errors downstream
if errors.Is(err, sql.ErrNoRows) {
    // Handle not found
}

// Or extract error type
var pErr *PermissionError
if errors.As(err, &pErr) {
    log.Error("permission denied", "user", pErr.UserID)
}
```

### Pattern 3: Structured Logging with slog (Go)

**What:** Use standard library slog for structured, level-based logging

**When to use:** All logging (replaces fmt.Println, log.Printf)

**Example:**
```go
// Source: https://betterstack.com/community/guides/logging/logging-in-go/
import "log/slog"

// Setup (once at startup)
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Usage
logger.Info("position opened",
    "account_id", acc.ID,
    "symbol", pos.Symbol,
    "volume", pos.Volume,
    "price", pos.OpenPrice,
)

logger.Error("margin calculation failed",
    "account_id", acc.ID,
    "error", err,
)
```

### Recommended Frontend Project Structure (React/TypeScript)

```
clients/desktop/src/
├── features/            # Feature-based organization
│   ├── trading/
│   │   ├── components/  # Feature-specific components
│   │   ├── hooks/       # Feature-specific hooks
│   │   ├── services/    # API calls for this feature
│   │   └── types.ts     # Feature types
│   ├── orders/
│   ├── positions/
│   └── account/
├── shared/              # Shared across features
│   ├── components/      # Common UI components
│   ├── hooks/           # Shared custom hooks
│   ├── utils/           # Helper functions
│   ├── services/        # Shared API clients
│   └── types/           # Shared types
├── indicators/          # Domain-specific module
└── App.tsx
```

**Current vs. Recommended:**
- Current: All components in `components/`, mixed responsibilities
- Recommended: Feature folders with scoped components, shared folder for common code

### Pattern 4: Feature-Based Organization (React)

**What:** Group files by feature/domain rather than by type (all components together)

**When to use:** Applications with multiple distinct features/domains

**Example:**
```
features/
├── trading/
│   ├── TradingChart.tsx       # Main feature component
│   ├── OrderEntry.tsx         # Sub-component
│   ├── useTradingData.ts      # Feature hook
│   └── tradingService.ts      # API calls
└── orders/
    ├── OrdersPanel.tsx
    ├── useOrders.ts
    └── ordersService.ts
```

### Pattern 5: Custom Hooks for Logic Extraction (React)

**What:** Extract reusable stateful logic into custom hooks (prefix with `use`)

**When to use:** Any logic repeated across components, or to separate concerns from UI

**Example:**
```typescript
// Source: https://react.dev/learn/reusing-logic-with-custom-hooks
// hooks/useWebSocket.ts
export function useWebSocket(url: string) {
    const [data, setData] = useState(null);
    const [status, setStatus] = useState<'connecting' | 'connected' | 'disconnected'>('connecting');

    useEffect(() => {
        const ws = new WebSocket(url);

        ws.onopen = () => setStatus('connected');
        ws.onmessage = (event) => setData(JSON.parse(event.data));
        ws.onerror = () => setStatus('disconnected');

        return () => ws.close();
    }, [url]);

    return { data, status };
}

// Usage in component
function TradingChart() {
    const { data, status } = useWebSocket('ws://localhost:8080/market-data');

    if (status !== 'connected') return <div>Connecting...</div>;
    return <Chart data={data} />;
}
```

### Pattern 6: Component Splitting by Responsibility (React)

**What:** Split large components (>500 lines) into smaller, single-purpose components

**When to use:** Component has bloated state, multiple unrelated concerns, or hard-to-read JSX

**Example:**
```typescript
// Source: https://alexkondov.com/refactoring-a-messy-react-component/

// BEFORE: TradingChart.tsx (900+ lines)
function TradingChart() {
    // State for chart, orders, positions, indicators, websocket...
    // 900+ lines of mixed concerns
}

// AFTER: Split into focused components
function TradingChart() {
    return (
        <ChartContainer>
            <ChartCanvas />
            <IndicatorPane />
            <OrderLevels />
            <PositionMarkers />
        </ChartContainer>
    );
}

// Each sub-component is 50-150 lines, single responsibility
function ChartCanvas() { /* chart rendering only */ }
function IndicatorPane() { /* indicator display only */ }
function OrderLevels() { /* order visualization only */ }
```

### Anti-Patterns to Avoid

- **Over-engineering architecture:** Don't create `internal/ports/repositories/interfaces/` if you have 3 files. Start simple.
- **Mixing domain and infrastructure:** Don't put database queries in business logic functions
- **God components:** 900+ line React components that do everything (already identified: TradingChart.tsx)
- **Ignoring error context:** Returning raw errors without wrapping loses debugging information
- **Using fmt.Printf for logging:** Unstructured logs are hard to query in production
- **Not using TypeScript strict mode:** Defeats the purpose of TypeScript
</architecture_patterns>

<dont_hand_roll>
## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Linting configuration | Custom lint rules from scratch | golangci-lint `recommended` preset | 50+ linters pre-configured, actively maintained |
| Structured logging | Custom logger wrapper | slog (stdlib) | Future-proof, standard library, zero dependencies |
| Error wrapping | Custom error types for all cases | errors.Is/errors.As with %w | Standard library, compiler support, community patterns |
| Code duplication detection | Manual code review only | jscpd in CI/CD | Rabin-Karp algorithm, 150+ languages, automated |
| TypeScript linting | Custom ESLint rules | typescript-eslint `recommended` | Official TypeScript team preset, type-aware rules |
| Component state management | Prop drilling everything | Custom hooks for shared logic | React-idiomatic, testable, reusable |
| HTTP routing (already using Gin) | Custom router | Keep using Gin (already good choice) | Battle-tested, middleware ecosystem |
| Decimal precision (already done) | float64 arithmetic | Keep using govalues/decimal | Already integrated in Phase 6 |

**Key insight:** The Go and TypeScript ecosystems have mature, well-maintained solutions for common problems. Fighting against community standards creates maintenance burden. The trading platform already made good choices (Gin, govalues/decimal, pgx) - extend this pattern to code organization and quality tooling.
</dont_hand_roll>

<common_pitfalls>
## Common Pitfalls

### Pitfall 1: Over-Applying Clean Architecture

**What goes wrong:** Creating dozens of directories and interfaces for a simple CRUD operation, leading to navigation hell and cognitive overhead

**Why it happens:** Misunderstanding that clean architecture is about dependency direction, not directory count

**How to avoid:**
- Start with simple structure: `cmd/`, `internal/`, `pkg/` (if needed)
- Add layers (domain/ports/adapters) only when you have genuine abstraction needs
- If you only have one implementation of an interface, question if you need the interface

**Warning signs:**
- Navigating through 5+ directories to change a simple function
- Interfaces with only one implementation
- Team complaining about "where do I put this file?"

### Pitfall 2: Ignoring Error Context

**What goes wrong:** Returning raw errors without wrapping, making production debugging impossible

**Why it happens:** Laziness or unawareness of error wrapping patterns

**How to avoid:**
- Use `fmt.Errorf("context: %w", err)` at every error boundary
- Log errors with structured fields (account_id, symbol, etc.) using slog
- Use errors.Is/errors.As instead of == for error checking

**Warning signs:**
- Production errors say "database error" with no context
- Can't tell which user/account/position caused the error
- Error logs missing critical debugging information

### Pitfall 3: God Components (900+ Line React Files)

**What goes wrong:** Single components handling chart rendering, WebSocket, orders, positions, indicators - unmaintainable and untestable

**Why it happens:** "Just one more feature" syndrome - gradual bloat over time

**How to avoid:**
- Refactor when component exceeds 500 lines (hard limit)
- Extract custom hooks for non-UI logic (useWebSocket, useOrderData)
- Split JSX into sub-components by responsibility
- Use feature-based organization to prevent cross-feature coupling

**Warning signs:**
- Component has 10+ useState calls with unrelated state
- Scrolling through 900+ lines to find one function
- Tests require mocking 15+ dependencies

### Pitfall 4: Not Using slog (Still Using fmt.Printf)

**What goes wrong:** Unstructured logs are impossible to query in production monitoring systems

**Why it happens:** Inertia from old Go patterns, unawareness that slog is now stdlib

**How to avoid:**
- Migrate all logging to slog (structured, level-based)
- Use JSON handler for production (queryable by Prometheus, Loki, etc.)
- Add contextual fields: account_id, symbol, order_id, etc.

**Warning signs:**
- Grep-ing through text logs instead of querying structured data
- Can't filter logs by account or symbol
- No log levels (info/warn/error all mixed)

### Pitfall 5: Skipping golangci-lint Configuration

**What goes wrong:** Inconsistent code style, missing bugs that linters would catch

**Why it happens:** "We'll add it later" - never happens

**How to avoid:**
- Add .golangci.yml with `recommended` preset immediately
- Run in CI/CD (fail builds on violations)
- Use `--new-from-rev=HEAD~` in CI to only lint changed code (faster)

**Warning signs:**
- Code review debates about formatting (should be automated)
- Obvious bugs making it to production (ineffectual assignments, etc.)
- No linting in CI/CD

### Pitfall 6: Duplicate Code Proliferation

**What goes wrong:** Copy-pasted logic diverges, bugs fixed in one place but not others

**Why it happens:** Faster to copy-paste than refactor into shared function

**How to avoid:**
- Run jscpd in CI/CD with threshold (fail if >5% duplication)
- Extract shared functions into `internal/shared/` or `shared/utils/`
- Use custom hooks in React for repeated stateful logic

**Warning signs:**
- Fixing same bug in 3 different files
- Nearly-identical functions with small variations
- jscpd reports >10% duplication
</common_pitfalls>

<code_examples>
## Code Examples

Verified patterns from official sources:

### Go Clean Architecture - Domain Layer

```go
// Source: https://threedots.tech/post/introducing-clean-architecture/
// internal/domain/position/position.go

package position

import "github.com/trading-engine/internal/shared/decimal"

// Position is a pure domain entity (no database, no HTTP)
type Position struct {
    ID         string
    AccountID  string
    Symbol     string
    Volume     decimal.Decimal
    OpenPrice  decimal.Decimal
    CurrentPnL decimal.Decimal
}

// CalculatePnL is pure business logic
func (p *Position) CalculatePnL(currentPrice decimal.Decimal) decimal.Decimal {
    priceDiff := currentPrice.Sub(p.OpenPrice)
    return priceDiff.Mul(p.Volume)
}

// IsProfitable is domain logic
func (p *Position) IsProfitable() bool {
    return p.CurrentPnL.GreaterThan(decimal.Zero)
}
```

### Go Structured Logging with slog

```go
// Source: https://betterstack.com/community/guides/logging/logging-in-go/
// internal/adapters/http/handlers.go

package handlers

import (
    "log/slog"
    "net/http"
)

type Handler struct {
    logger *slog.Logger
}

func (h *Handler) ClosePosition(w http.ResponseWriter, r *http.Request) {
    posID := r.PathValue("id")

    h.logger.Info("closing position request",
        "position_id", posID,
        "ip", r.RemoteAddr,
    )

    if err := h.positionService.Close(r.Context(), posID); err != nil {
        h.logger.Error("failed to close position",
            "position_id", posID,
            "error", err,
        )
        http.Error(w, "Internal server error", 500)
        return
    }

    h.logger.Info("position closed successfully", "position_id", posID)
    w.WriteHeader(http.StatusOK)
}
```

### Go Error Wrapping

```go
// Source: https://go.dev/blog/go1.13-errors
// internal/adapters/postgres/position_repo.go

package postgres

import (
    "context"
    "fmt"
    "errors"

    "github.com/jackc/pgx/v5"
)

func (r *PositionRepository) Get(ctx context.Context, id string) (*Position, error) {
    var pos Position

    err := r.db.QueryRow(ctx, "SELECT * FROM positions WHERE id = $1", id).Scan(&pos)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            // Wrap with context, preserve error type
            return nil, fmt.Errorf("position %s not found: %w", id, err)
        }
        return nil, fmt.Errorf("failed to query position %s: %w", id, err)
    }

    return &pos, nil
}

// Downstream usage
pos, err := repo.Get(ctx, posID)
if err != nil {
    // Can still check error type
    if errors.Is(err, pgx.ErrNoRows) {
        return http.StatusNotFound
    }
    return http.StatusInternalServerError
}
```

### React Custom Hook for API

```typescript
// Source: https://react.dev/learn/reusing-logic-with-custom-hooks
// features/positions/hooks/usePositions.ts

import { useState, useEffect } from 'react';
import type { Position } from '../types';

export function usePositions(accountId: string) {
    const [positions, setPositions] = useState<Position[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<Error | null>(null);

    useEffect(() => {
        let cancelled = false;

        async function fetchPositions() {
            try {
                const res = await fetch(`/api/accounts/${accountId}/positions`);
                if (!res.ok) throw new Error('Failed to fetch positions');

                const data = await res.json();
                if (!cancelled) {
                    setPositions(data);
                    setLoading(false);
                }
            } catch (err) {
                if (!cancelled) {
                    setError(err as Error);
                    setLoading(false);
                }
            }
        }

        fetchPositions();

        return () => {
            cancelled = true; // Prevent state updates after unmount
        };
    }, [accountId]);

    return { positions, loading, error };
}

// Usage in component
function PositionsPanel({ accountId }: Props) {
    const { positions, loading, error } = usePositions(accountId);

    if (loading) return <Spinner />;
    if (error) return <ErrorMessage error={error} />;

    return <PositionsList positions={positions} />;
}
```

### React Component Splitting

```typescript
// Source: https://alexkondov.com/refactoring-a-messy-react-component/

// BEFORE: TradingChart.tsx (900+ lines, mixed concerns)

// AFTER: Split into focused modules

// features/trading/TradingChart.tsx (main orchestrator, ~100 lines)
import { ChartCanvas } from './components/ChartCanvas';
import { IndicatorPane } from './components/IndicatorPane';
import { OrderLevels } from './components/OrderLevels';
import { useChartData } from './hooks/useChartData';

export function TradingChart({ symbol }: Props) {
    const { ohlc, indicators } = useChartData(symbol);

    return (
        <div className="trading-chart">
            <ChartCanvas ohlc={ohlc} />
            <IndicatorPane indicators={indicators} />
            <OrderLevels symbol={symbol} />
        </div>
    );
}

// features/trading/components/ChartCanvas.tsx (~150 lines)
export function ChartCanvas({ ohlc }: Props) {
    // Chart rendering only - single responsibility
    // ...
}

// features/trading/hooks/useChartData.ts (~80 lines)
export function useChartData(symbol: string) {
    // Extract data fetching logic from component
    // ...
}
```

### TypeScript ESLint Configuration

```javascript
// Source: https://typescript-eslint.io/getting-started/
// eslint.config.mjs

import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';

export default tseslint.config(
    eslint.configs.recommended,
    ...tseslint.configs.recommended, // Type-safe recommended rules
    ...tseslint.configs.stylistic,   // Stylistic best practices
    {
        rules: {
            // Project-specific overrides if needed
            '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
        },
    },
);
```
</code_examples>

<sota_updates>
## State of the Art (2024-2026)

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Custom logger wrappers | slog (stdlib) | Go 1.21 (2023) | Standard library logging, no deps, future-proof |
| golang-standards/project-layout as gospel | Pragmatic, project-specific structure | 2024-2025 | "It depends" mindset, simpler projects |
| zap/zerolog only | slog as frontend + zap/zerolog backend | 2024-2025 | Best of both: slog API + performance backend |
| Prop drilling | Custom hooks + context | React 18+ (2022+) | Cleaner state sharing, better composition |
| Class components | Functional components + hooks | React 16.8+ (2019) | Standard since 2020, still relevant |
| ESLint old config | Flat config (eslint.config.mjs) | ESLint 9 (2024) | Simpler configuration, better TypeScript support |
| Type-based folder structure | Feature-based structure | 2023-2024 trend | Scales better for large apps |
| Manual error checking with == | errors.Is/errors.As | Go 1.13+ (2019) | Better error handling, still under-adopted |

**New tools/patterns to consider:**

- **slog with OpenTelemetry integration (2026):** Correlate logs with distributed traces for observability
- **React Server Components:** Not applicable for desktop app, but worth knowing for future web versions
- **golangci-lint v2.0 (2026):** New YAML format, some linters became formatters
- **typescript-eslint v8+:** Type-aware linting with better performance
- **jscpd for CI/CD:** Automated duplicate detection catching code smells before merge

**Deprecated/outdated:**

- **golang-standards/project-layout as "the standard":** Never was official, community moved to pragmatism
- **fmt.Printf for logging:** Use slog (structured, queryable)
- **Custom error types for everything:** Use error wrapping with %w instead
- **900+ line components:** React 18+ patterns make splitting trivial
- **Ignoring linters:** golangci-lint and typescript-eslint catch real bugs, not just style
</sota_updates>

<open_questions>
## Open Questions

Things that couldn't be fully resolved:

1. **Current file size distribution**
   - What we know: TradingChart.tsx mentioned as 900+ lines in roadmap
   - What's unclear: How many other files exceed recommended thresholds (500 lines)?
   - Recommendation: Run `find backend -name "*.go" -exec wc -l {} \; | sort -rn | head -20` and same for frontend to identify largest files during planning

2. **Existing code duplication percentage**
   - What we know: DRY principle listed as success criterion
   - What's unclear: Current duplication level (baseline for improvement)
   - Recommendation: Run jscpd during planning phase to get baseline metrics: `npx jscpd --threshold 0 --reporters console,json .`

3. **Current error handling patterns**
   - What we know: Some code uses error wrapping (seen in Phase 6 work)
   - What's unclear: What percentage of codebase uses proper error wrapping vs. raw errors?
   - Recommendation: Grep for `fmt.Errorf` without %w to find opportunities for improvement

4. **Logging consistency**
   - What we know: Roadmap mentions "structured logging best practices" as goal
   - What's unclear: Current logging library (if any) and usage patterns
   - Recommendation: Search for logging calls (`log.`, `fmt.Print`, `slog.`) during planning to audit current state

5. **Test coverage gaps from refactoring**
   - What we know: Phase 3 (Testing Infrastructure) partially complete (6/7 plans)
   - What's unclear: How refactoring might break existing tests?
   - Recommendation: Run tests before and after each refactoring step, use table-driven tests for new code
</open_questions>

<sources>
## Sources

### Primary (HIGH confidence)

**Go Official:**
- [Go Modules Layout (go.dev)](https://go.dev/doc/modules/layout) - Official Go project structure guidance
- [Go 1.13 Errors Blog (go.dev)](https://go.dev/blog/go1.13-errors) - Error wrapping with errors.Is/errors.As
- [Error Handling and Go (go.dev)](https://go.dev/blog/error-handling-and-go) - Official error patterns

**TypeScript ESLint Official:**
- [typescript-eslint Shared Configs](https://typescript-eslint.io/users/configs/) - Recommended presets
- [typescript-eslint Getting Started](https://typescript-eslint.io/getting-started/) - Setup guide

**React Official:**
- [Reusing Logic with Custom Hooks (react.dev)](https://react.dev/learn/reusing-logic-with-custom-hooks) - Custom hook patterns

**golangci-lint Official:**
- [golangci-lint Configuration](https://golangci-lint.run/docs/configuration/) - Official config docs
- [golangci-lint Changelog](https://golangci-lint.run/docs/product/changelog/) - v2.8.0 (Jan 2026)

### Secondary (MEDIUM confidence)

**Go Architecture & Best Practices:**
- [Clean Architecture in Go (Three Dots Labs)](https://threedots.tech/post/introducing-clean-architecture/) - Practical clean architecture
- [Hexagonal Architecture in Go (Skoredin)](https://skoredin.pro/blog/golang/hexagonal-architecture-go) - 2026 perspective
- [No Nonsense Go Package Layout (2024)](https://laurentsv.com/blog/2024/10/19/no-nonsense-go-package-layout.html) - Pragmatic advice
- [11 Tips for Go Projects (Alex Edwards)](https://www.alexedwards.net/blog/11-tips-for-structuring-your-go-projects) - Industry veteran advice

**Go Logging:**
- [Logging in Go with Slog (Better Stack)](https://betterstack.com/community/guides/logging/logging-in-go/) - Comprehensive slog guide
- [High-Performance Logging with slog and zerolog (Leapcell)](https://leapcell.io/blog/high-performance-structured-logging-in-go-with-slog-and-zerolog) - Performance comparison
- [Golang Logging Libraries (Uptrace)](https://uptrace.dev/blog/golang-logging) - 2025 comparison

**React & TypeScript:**
- [Bulletproof React Structure](https://github.com/alan2207/bulletproof-react/blob/master/docs/project-structure.md) - Feature-based organization
- [React Folder Structure (Profy Dev)](https://profy.dev/article/react-folder-structure) - Screaming architecture
- [React Project Structure (Developer Way)](https://www.developerway.com/posts/react-project-structure) - Decomposition patterns
- [Refactoring Messy React Components (Alex Kondov)](https://alexkondov.com/refactoring-a-messy-react-component/) - Practical refactoring
- [Splitting Components in React (Medium)](https://thiraphat-ps-dev.medium.com/splitting-components-in-react-a-path-to-cleaner-and-more-maintainable-code-f0828eca627c) - Component patterns

**Code Quality:**
- [jscpd (npm)](https://www.npmjs.com/package/jscpd) - Code duplication detection
- [dupl (GitHub)](https://github.com/mibk/dupl) - Go duplicate detector
- [Golden Config for golangci-lint (GitHub Gist)](https://gist.github.com/maratori/47a4d00457a92aa426dbd48a18776322) - Community config

### Tertiary (LOW confidence - needs validation)

None - all findings verified with official documentation or authoritative community sources
</sources>

<metadata>
## Metadata

**Research scope:**
- Core technology: Go (backend), React/TypeScript (frontend)
- Ecosystem: golangci-lint, slog, typescript-eslint, jscpd
- Patterns: Clean architecture, feature-based organization, custom hooks, error wrapping
- Pitfalls: God components, missing error context, unstructured logging, no linting

**Confidence breakdown:**
- Standard stack: HIGH - Official tools (slog, typescript-eslint) and widely-adopted community tools (golangci-lint, jscpd)
- Architecture: HIGH - Verified with official Go/React docs, industry authorities (Three Dots Labs, Alex Edwards, Bulletproof React)
- Pitfalls: HIGH - Common patterns from production experience, documented in multiple sources
- Code examples: HIGH - From official docs (go.dev, react.dev, typescript-eslint.io) or well-known authorities

**Research date:** 2026-01-16
**Valid until:** 2026-02-16 (30 days - stable ecosystems, though golangci-lint updates frequently)

**Notes:**
- Go and React/TypeScript are mature, stable ecosystems - patterns change slowly
- golangci-lint releases frequently (v2.8.0 on 2026-01-07), check for updates during planning
- Trading platform already made good architectural choices (repository pattern, dependency injection, decimal precision) - this research extends those patterns to code organization
</metadata>

---

*Phase: 16-code-organization-best-practices*
*Research completed: 2026-01-16*
*Ready for planning: yes*
