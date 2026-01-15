# Coding Conventions

**Analysis Date:** 2026-01-15

## Naming Patterns

**Files (TypeScript/React):**
- PascalCase for components: `TradingChart.tsx`, `Login.tsx`, `IndicatorManager.tsx`
- camelCase for hooks: `useIndicators.ts`, `useChartData.ts`
- PascalCase for services: `DataCache.ts`, `IndicatorStorage.ts`, `ExternalDataService.ts`
- Test files: `*.test.ts` co-located with source

**Files (Go):**
- lowercase: `manager.go`, `engine.go`, `hub.go`, `client.go`
- Test files: `*_test.go` (none found - testing gap)

**Functions (TypeScript):**
- camelCase for all functions
- `use` prefix for hooks: `useIndicators()`, `useChartData()`
- Event handlers: `handleClick`, `handleSubmit`

**Functions (Go):**
- PascalCase for exported: `NewEngine()`, `GetAccount()`, `ExecuteMarketOrder()`
- camelCase for unexported: `initDefaultSymbols()`, `saveConfigLocked()`

**Variables (TypeScript):**
- camelCase: `selectedSymbol`, `isAuthenticated`, `orderLoading`
- UPPER_SNAKE_CASE for constants: `API_BASE`, `DB_NAME`, `MAX_AGE_DAYS`

**Variables (Go):**
- camelCase for unexported fields: `priceCallback`, `bbookEngine`, `activeAggregators`
- PascalCase for exported fields

**Types (TypeScript):**
- PascalCase for types: `ChartIndicator`, `AccountUpdate`, `Position`
- Union types preferred over enums: `type ChartType = 'candlestick' | 'heikinAshi' | 'bar' | 'line'`
- No enum keyword used (per CLAUDE.md)

**Types (Go):**
- PascalCase: `type Engine struct`, `type LPAdapter interface`
- JSON tags use camelCase: `` `json:"accountId"` ``

## Code Style

**Formatting (TypeScript):**
- 2-space indentation
- Single quotes for strings
- Double quotes for JSX attributes
- Semicolons required
- Trailing commas in multiline objects/arrays
- No explicit Prettier config (using ESLint flat config)

**Formatting (Go):**
- Tab indentation (gofmt standard)
- Standard Go formatting via `gofmt`

**Linting (TypeScript):**
- ESLint 9.39.1 with flat config (`eslint.config.js`)
- TypeScript ESLint recommended rules
- React Hooks linting
- Run: `bun run lint`

**TypeScript Config:**
- `strict: true` - strictest type checking
- `noUnusedLocals: true` - error on unused variables
- `noUnusedParameters: true` - error on unused parameters
- Target: ES2022

## Import Organization

**TypeScript Order:**
1. External packages (react, lightweight-charts)
2. Internal modules (relative imports)
3. Type imports last

**Go Order:**
1. Standard library
2. External packages (gorilla/websocket, golang-jwt/jwt)
3. Internal packages (github.com/epic1st/rtx/backend/*)

## Error Handling

**TypeScript Patterns:**
- Try/catch for async operations
- ErrorBoundary for React component errors (`clients/desktop/src/components/ErrorBoundary.tsx`)
- Console.error for logging

**Go Patterns:**
- Return `error` as last return value
- Check `if err != nil` immediately
- Gap: Many ignored errors with blank identifier (`body, _ := io.ReadAll(...)`)
- Log before returning errors: `log.Printf("[Service] Error: %v", err)`

## Logging

**TypeScript:**
- `console.log()` for debugging
- `console.error()` for errors
- Logged to browser console

**Go:**
- `log.Printf()` with service prefix: `[LP Manager]`, `[Binance]`, `[PnL Engine]`
- Format: `[Service] Message` pattern
- Logged to stdout

## Comments

**TypeScript:**
- JSDoc-style for complex functions
- Inline comments explaining non-obvious logic
- Section headers: `// Cache structure for performance`

**Go:**
- Godoc comments for exported items: `// Engine handles B-Book execution`
- Inline comments for logic blocks
- TODO comments: `// TODO: Implement account-specific WebSocket`

## Function Design

**Size:**
- Keep functions focused and small
- Current reality: Some large functions (800+ lines in engine.go)
- Goal: Extract helpers for complex logic

**Parameters (TypeScript):**
- Destructure object parameters: `function useIndicators({ ohlcData, autoCalculate = true })`
- Use options object for 4+ parameters

**Parameters (Go):**
- Named parameters for clarity
- Pointer receivers for methods that modify state
- Context as first parameter (not consistently used)

**Return Values (TypeScript):**
- Explicit returns
- Return early for guard clauses
- Hooks return arrays: `[state, setState]`

**Return Values (Go):**
- Return `error` as last value
- Return early for error cases

## Module Design

**Exports (TypeScript):**
- Named exports preferred
- Default exports for React components
- Barrel files (index.ts) for public API

**Exports (Go):**
- Uppercase = exported
- Lowercase = package-private
- Interfaces in separate files

**Anti-Patterns to Avoid:**
- **TypeScript**: `any` type (found extensively - should be fixed)
- **Go**: Ignoring errors with blank identifier (common - should be fixed)
- **Both**: Hardcoded configuration (use environment variables)

---

*Convention analysis: 2026-01-15*
*Update when patterns change*
