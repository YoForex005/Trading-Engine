# Testing Patterns

**Analysis Date:** 2026-01-15

## Test Framework

**Frontend:**
- Runner: Vitest 4.0.17
- Config: `clients/desktop/vitest.config.ts`
- Assertion Library: Vitest built-in `expect`

**Backend:**
- No test framework detected (0 Go test files found)
- Testing gap: Critical business logic untested

**Run Commands:**
```bash
# Frontend tests
bun test                           # Run all tests
bun test -- --watch                # Watch mode
bun test -- path/to/file.test.ts  # Single file
bun run test:coverage              # Coverage report
bun run test:ui                    # UI mode
```

## Test File Organization

**Frontend Location:**
- Co-located: `src/services/__tests__/IndicatorStorage.test.ts`
- Nested: `src/indicators/core/__tests__/IndicatorEngine.test.ts`
- Pattern: `__tests__/` directories alongside source

**Naming:**
- Unit tests: `{module}.test.ts`
- Integration: Same pattern (no special suffix)

**Backend:**
- Expected pattern: `*_test.go` (none found)
- Recommended: Co-locate with source files

## Test Structure

**Suite Organization:**
```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest';

describe('IndicatorStorage', () => {
  const mockIndicators: ChartIndicator[] = [ ... ];

  beforeEach(() => {
    localStorageMock.clear();
    vi.clearAllMocks();
  });

  describe('save', () => {
    it('should save indicators to localStorage', () => {
      IndicatorStorage.save('EURUSD', '1h', mockIndicators);
      const stored = localStorage.getItem(key);
      expect(stored).not.toBeNull();
    });

    it('should handle invalid data gracefully', () => {
      expect(() => IndicatorStorage.save('', '', [])).not.toThrow();
    });
  });
});
```

**Patterns:**
- Nested `describe()` blocks by functionality
- `it('should...')` pattern describing expected behavior
- Mock data setup at suite level
- Setup/teardown with `beforeEach()`

## Mocking

**Framework:**
- Vitest built-in mocking (`vi`)
- Module mocking via `vi.mock()`

**What to Mock:**
- Browser APIs: localStorage, matchMedia, IntersectionObserver, ResizeObserver
- External services (when implemented)

**Mock Setup:**
`clients/desktop/src/test/setup.ts`:
```typescript
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};

global.localStorage = localStorageMock as any;

global.matchMedia = vi.fn().mockImplementation(query => ({ ... }));
```

## Fixtures and Factories

**Test Data:**
```typescript
// Inline in tests
const mockOHLCData: OHLC[] = [
  { time: 1000, open: 100, high: 105, low: 95, close: 102 },
  { time: 2000, open: 102, high: 108, low: 100, close: 105 },
];

// Factory pattern
const createTestIndicator = (overrides?: Partial<ChartIndicator>) => ({
  id: 'test-id',
  type: 'SMA' as const,
  params: { period: 20 },
  ...overrides,
});
```

**Location:**
- Factory functions: Defined in test file near usage
- Shared fixtures: Could go in `src/test/fixtures/` (not currently used)

## Coverage

**Requirements:**
- Lines: 80%
- Functions: 80%
- Branches: 80%
- Statements: 80%

**Configuration:**
`vitest.config.ts`:
```typescript
coverage: {
  provider: 'v8',
  reporter: ['text', 'json', 'html'],
  lines: 80,
  functions: 80,
  branches: 80,
  statements: 80,
}
```

**View Coverage:**
```bash
bun run test:coverage
open coverage/index.html
```

## Test Types

**Unit Tests:**
- Scope: Test single function in isolation
- Files: `IndicatorStorage.test.ts`, `IndicatorEngine.test.ts`
- Mocking: Mock all external dependencies
- Speed: Fast (<100ms per test)

**Integration Tests:**
- Not explicitly separated from unit tests
- Could test multiple components together

**E2E Tests:**
- Not implemented
- Future: Could use Playwright or Cypress

## Common Patterns

**Async Testing:**
```typescript
it('should handle async operation', async () => {
  const result = await asyncFunction();
  expect(result).toBe('expected');
});
```

**Error Testing:**
```typescript
it('should throw on invalid input', () => {
  expect(() => functionCall()).toThrow('error message');
});
```

**Mock Function Testing:**
```typescript
it('should call function with correct args', () => {
  const mockFn = vi.fn();
  component.onClick(mockFn);
  expect(mockFn).toHaveBeenCalledWith('expected-arg');
});
```

## Test Coverage Gaps

**Backend:**
- **CRITICAL**: 0 Go test files
- Untested: Core engine, LP manager, WebSocket hub, all business logic
- Priority: Add tests for `backend/internal/core/engine.go`, `backend/lpmanager/manager.go`

**Frontend:**
- Good: Indicator engine, indicator storage
- Missing: TradingChart component tests, App component tests, service tests

---

*Testing analysis: 2026-01-15*
*Update when test patterns change*
