# Contributing to Trading Engine

This document provides guidelines for contributing to the Trading Engine project.

## Error Handling

### Always Wrap Errors with Context

Use `fmt.Errorf` with `%w` to preserve the error chain and add contextual information:

```go
if err := repo.Get(ctx, id); err != nil {
    return fmt.Errorf("failed to get account %d: %w", id, err)
}
```

**Why:** Contextual error messages make debugging in production much easier. Include relevant IDs, symbols, and other identifying information.

### Use Custom Error Types for Business Logic

Import the shared errors package:

```go
import "github.com/epic1st/rtx/backend/internal/shared/errors"
```

Return typed errors for common scenarios:

```go
// Return NotFoundError when a resource doesn't exist
if !ok {
    return errors.NewNotFound("account", accountID)
}

// Return ValidationError for invalid input
if volume < spec.MinVolume {
    return errors.NewValidation("volume", "volume too small")
}

// Return InsufficientFundsError for balance issues
if balance < required {
    return errors.NewInsufficientFunds(accountID, required, balance)
}
```

### Check Error Types in Handlers

Use `errors.As` to check for specific error types and return appropriate HTTP status codes:

```go
import stderrors "errors"

// In HTTP handler
if err != nil {
    // Check for validation errors (400 Bad Request)
    var valErr *errors.ValidationError
    if stderrors.As(err, &valErr) {
        http.Error(w, valErr.Message, http.StatusBadRequest)
        return
    }

    // Check for not found errors (404 Not Found)
    var notFoundErr *errors.NotFoundError
    if stderrors.As(err, &notFoundErr) {
        http.Error(w, "resource not found", http.StatusNotFound)
        return
    }

    // Check for insufficient funds (422 Unprocessable Entity)
    var fundsErr *errors.InsufficientFundsError
    if stderrors.As(err, &fundsErr) {
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "insufficient funds",
            "required": fundsErr.Required,
            "available": fundsErr.Available,
        })
        return
    }

    // Generic error (500 Internal Server Error)
    logging.Default.Error("operation failed", "error", err)
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
}
```

### Check Errors with errors.Is for Sentinel Values

Use `errors.Is` to check for specific error values:

```go
import stderrors "errors"

if err != nil {
    // Check for specific database errors
    if stderrors.Is(err, pgx.ErrNoRows) {
        return errors.NewNotFound("position", id)
    }

    // Check for context errors
    if stderrors.Is(err, context.Canceled) {
        return fmt.Errorf("operation canceled: %w", err)
    }

    return fmt.Errorf("database query failed: %w", err)
}
```

### Never Ignore Errors

All errors must be handled explicitly:

```go
// BAD - error ignored
repo.Update(ctx, account)

// GOOD - error checked and handled
if err := repo.Update(ctx, account); err != nil {
    return fmt.Errorf("failed to update account: %w", err)
}

// ACCEPTABLE - intentional ignore with comment
_ = file.Close()  // Defer handles error, ignore here
```

The `errcheck` linter is enabled to catch unchecked errors automatically.

### Error Wrapping Best Practices

**DO:**
- Add specific context (IDs, symbols, amounts) to every error
- Use `%w` verb to preserve the error chain
- Return custom error types at API boundaries
- Log errors with structured logging before wrapping

**DON'T:**
- Wrap errors more than 2-3 times in a single call stack
- Use `fmt.Errorf("%v", err)` - this loses the error chain
- Create generic error messages like "operation failed"
- Ignore errors from critical operations (database writes, file I/O)

### Example: Complete Error Handling Flow

```go
// Repository layer - return typed errors
func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*Account, error) {
    var acc Account
    err := r.pool.QueryRow(ctx, query, id).Scan(&acc)

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.NewNotFound("account", fmt.Sprintf("%d", id))
        }
        return nil, fmt.Errorf("failed to get account %d: %w", id, err)
    }

    return &acc, nil
}

// Business logic layer - wrap with business context
func (e *Engine) ExecuteOrder(accountID int64, order Order) (*Position, error) {
    account, err := e.accountRepo.GetByID(ctx, accountID)
    if err != nil {
        return nil, fmt.Errorf("order execution failed for account %d: %w", accountID, err)
    }

    // ... business logic
}

// HTTP handler - check error types and return appropriate status
func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
    position, err := h.engine.ExecuteOrder(accountID, order)
    if err != nil {
        var notFoundErr *errors.NotFoundError
        if errors.As(err, &notFoundErr) {
            http.Error(w, "account not found", http.StatusNotFound)
            return
        }

        var valErr *errors.ValidationError
        if errors.As(err, &valErr) {
            http.Error(w, valErr.Message, http.StatusBadRequest)
            return
        }

        logging.Default.Error("order execution failed", "error", err)
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(position)
}
```

## Code Style

- **NEVER use `enum`** - always use literal unions (string unions)
- Prefer `type` over `interface`
- Use descriptive variable names
- Add comments for complex logic
- Keep functions small and focused

## Testing

- Write tests for all error paths
- Use table-driven tests with `t.Parallel()`
- Test custom error types with `errors.As`
- Verify error messages include context

## Linting

Run linting before submitting:

```bash
cd backend
golangci-lint run
```

The following linters are enabled:
- `govet` - Critical syntax and type checking
- `ineffassign` - Detect ineffectual assignments
- `unused` - Find unused code
- `misspell` - Spell checking
- `errcheck` - Ensure all errors are checked

## Code Duplication (DRY Principle)

### Avoid Copy-Paste Programming

**Don't Repeat Yourself (DRY)** - Code duplication is actively monitored and prevented in CI/CD.

Before copying code, ask:
1. Can this be extracted into a shared function/hook?
2. Is this logic truly identical, or just similar?
3. Will this logic change together across all copies?

**If yes to all three, extract it.**

### Rule of Three

Extract shared logic when duplicated **3+ times**, not at first duplication:
- Code block is **>5 lines**
- Logic has **identical purpose** across instances
- Changes will affect all instances together

### Extracting Shared Logic

#### Backend (Go)

Put shared utilities in `backend/internal/shared/`:

| Type | Location | Use Case |
|------|----------|----------|
| HTTP utilities | `internal/shared/httputil/` | CORS, response helpers, request parsing |
| Database helpers | `internal/shared/database/` | Query error handling, connection utilities |
| Validation | `internal/shared/validation/` | Decimal/string validation helpers |
| Logging | `internal/shared/logging/` | Structured logging utilities |

**Example: HTTP Response Utilities**

```go
// ❌ DON'T - Duplicated across handlers
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(data)

// ✅ DO - Use shared utility
import "github.com/trading-engine/backend/internal/shared/httputil"
httputil.RespondWithJSON(w, http.StatusOK, data)
```

**Example: CORS Handling**

```go
// ❌ DON'T - Duplicated in every handler
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

// ✅ DO - Use shared helper
httputil.SetCORSHeaders(w)
if httputil.HandleOPTIONS(w, r) {
    return
}
```

**Example: Database Error Handling**

```go
// ❌ DON'T - Duplicated error handling
if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, errors.NewNotFound("account", id)
    }
    return nil, fmt.Errorf("failed to get account %s: %w", id, err)
}

// ✅ DO - Use shared helper
import "github.com/trading-engine/backend/internal/shared/database"
if err := database.HandleQueryError(err, "account", id); err != nil {
    return nil, err
}
```

**Example: Validation**

```go
// ❌ DON'T - Duplicated validation logic
if order.Volume.LessThanOrEqual(decimal.Zero) {
    return fmt.Errorf("volume must be positive")
}

// ✅ DO - Use shared validator
import "github.com/trading-engine/backend/internal/shared/validation"
if err := validation.ValidatePositive(order.Volume, "volume"); err != nil {
    return err
}
```

#### Frontend (TypeScript)

Put shared utilities in `clients/desktop/src/shared/`:

| Type | Location | Use Case |
|------|----------|----------|
| UI Components | `shared/components/` | LoadingSpinner, ErrorMessage |
| API Client | `shared/services/` | HTTP fetch utilities |
| Validation | `shared/utils/validation.ts` | Form validation |
| Formatting | `shared/utils/formatting.ts` | Currency, dates, numbers |

**Example: API Client**

```typescript
// ❌ DON'T - Duplicated fetch logic
const res = await fetch(`/api/accounts/${id}`);
if (!res.ok) throw new Error('Failed to fetch account');
const data = await res.json();

// ✅ DO - Use shared API client
import { api } from '@/shared/services/api';
const data = await api.get<Account>(`/api/accounts/${id}`);
```

**Example: Validation**

```typescript
// ❌ DON'T - Duplicated validation
if (!symbol || symbol.trim() === '') {
  setError('Symbol is required');
  return;
}
if (volume <= 0) {
  setError('Volume must be positive');
  return;
}

// ✅ DO - Use shared validators
import { validators } from '@/shared/utils/validation';

const symbolError = validators.required(symbol, 'Symbol');
if (symbolError) {
  setError(symbolError);
  return;
}

const volumeError = validators.positive(volume, 'Volume');
if (volumeError) {
  setError(volumeError);
  return;
}
```

**Example: Loading/Error UI**

```tsx
// ❌ DON'T - Duplicated loading/error components
if (loading) return <div className="spinner">Loading...</div>;
if (error) return <div className="error">{error.message}</div>;

// ✅ DO - Use shared components
import { LoadingSpinner } from '@/shared/components/LoadingSpinner';
import { ErrorMessage } from '@/shared/components/ErrorMessage';

if (loading) return <LoadingSpinner />;
if (error) return <ErrorMessage error={error} />;
```

### CI/CD Enforcement

Code duplication is checked automatically via jscpd:

- **Threshold:** 5% (builds fail if exceeded)
- **Report:** Available in GitHub Actions artifacts
- **Runs on:** Every pull request and push to main

**View duplication report locally:**

```bash
# Run duplication check
npx jscpd .

# View HTML report
open .jscpd-report/html/index.html

# Check specific directory
npx jscpd backend/
npx jscpd clients/desktop/src/
```

### When Duplication is Acceptable

- **Test fixtures** - Similar setup, different assertions
- **Configuration files** - Similar structure, different values
- **Small snippets** (<5 lines) where abstraction adds complexity
- **Domain-specific logic** - Similar code with different business purposes

**Example of acceptable duplication:**

```go
// These serve different business purposes - don't force into shared function
func ValidateBuyOrder(order *Order) error {
    if order.Price.GreaterThan(market.Ask) {
        return errors.New("BUY_LIMIT price must be below market")
    }
    return nil
}

func ValidateSellOrder(order *Order) error {
    if order.Price.LessThan(market.Bid) {
        return errors.New("SELL_LIMIT price must be above market")
    }
    return nil
}
```

### Anti-Patterns to Avoid

- **Over-abstraction:** Don't create a utility for 3 lines of code used twice
- **Premature extraction:** Wait until code is duplicated 3+ times before extracting
- **Ignoring context:** Similar code with different purposes shouldn't be merged
- **Breaking semantics:** Don't force-fit unrelated code into shared utility just to reduce duplication percentage

## Documentation

- Update this file when adding new error types
- Document error handling patterns in code comments
- Include error examples in API documentation
