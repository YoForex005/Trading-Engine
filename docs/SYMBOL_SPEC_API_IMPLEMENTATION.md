# Symbol Specification API Implementation

## Overview
This document describes the implementation of the Symbol Specification API endpoint that provides detailed trading specifications for financial instruments.

## Implementation Summary

### 1. Backend API Endpoint

**File**: `backend/api/server.go`

#### Location of Changes:
- **Line 3-9**: Added imports (`regexp`, `strings`)
- **Line 611-778**: Added `SymbolSpecification` struct, `getSymbolSpecification()` function, and `HandleGetSymbolSpec()` handler

#### Endpoint Details:
- **URL Pattern**: `/api/symbols/{symbol}/spec`
- **Method**: GET
- **Response**: JSON object with symbol specifications

#### SymbolSpecification Structure:
```go
type SymbolSpecification struct {
    Symbol        string  `json:"symbol"`
    Description   string  `json:"description"`
    ContractSize  float64 `json:"contractSize"`
    PipValue      float64 `json:"pipValue"`
    PipPosition   int     `json:"pipPosition"` // Decimal places
    MinLot        float64 `json:"minLot"`
    MaxLot        float64 `json:"maxLot"`
    LotStep       float64 `json:"lotStep"`
    MarginRate    float64 `json:"marginRate"`   // e.g. 0.01 = 1%
    SwapLong      float64 `json:"swapLong"`
    SwapShort     float64 `json:"swapShort"`
    Commission    float64 `json:"commission"`
    Currency      string  `json:"currency"`
    BaseCurrency  string  `json:"baseCurrency"`
    QuoteCurrency string  `json:"quoteCurrency"`
}
```

#### Supported Symbols (Hardcoded):
1. **EURUSD** - Euro vs US Dollar
2. **GBPUSD** - British Pound vs US Dollar
3. **USDJPY** - US Dollar vs Japanese Yen
4. **XAUUSD** - Gold vs US Dollar
5. **USDCHF** - US Dollar vs Swiss Franc
6. **AUDUSD** - Australian Dollar vs US Dollar

### 2. Route Registration

**File**: `backend/cmd/server/main.go`

**Location**: Lines 752-760

```go
// Symbol Specification API - must be before /api/symbols/available to avoid conflicts
http.HandleFunc("/api/symbols/", func(w http.ResponseWriter, r *http.Request) {
    // Handle /api/symbols/{symbol}/spec
    if strings.HasSuffix(r.URL.Path, "/spec") {
        server.HandleGetSymbolSpec(w, r)
        return
    }
    http.Error(w, "Not found", http.StatusNotFound)
})
```

**Note**: Route is registered BEFORE `/api/symbols/available` to avoid routing conflicts.

### 3. Frontend TypeScript Interface

**File**: `clients/desktop/src/types/trading.ts`

**Location**: Lines 48-65

```typescript
export interface SymbolSpecification {
  symbol: string;
  description: string;
  contractSize: number;
  pipValue: number;
  pipPosition: number; // Decimal places (2=0.01, 5=0.00001)
  minLot: number;
  maxLot: number;
  lotStep: number;
  marginRate: number; // Margin requirement (e.g. 0.01 = 1%)
  swapLong: number;
  swapShort: number;
  commission: number;
  currency: string;
  baseCurrency: string;
  quoteCurrency: string;
}
```

## API Usage Examples

### 1. Get EURUSD Specification
```bash
curl http://localhost:7999/api/symbols/EURUSD/spec
```

**Response**:
```json
{
  "symbol": "EURUSD",
  "description": "Euro vs US Dollar",
  "contractSize": 100000,
  "pipValue": 10.0,
  "pipPosition": 5,
  "minLot": 0.01,
  "maxLot": 100.0,
  "lotStep": 0.01,
  "marginRate": 0.01,
  "swapLong": -0.5,
  "swapShort": 0.2,
  "commission": 0.0,
  "currency": "USD",
  "baseCurrency": "EUR",
  "quoteCurrency": "USD"
}
```

### 2. Get Gold (XAUUSD) Specification
```bash
curl http://localhost:7999/api/symbols/XAUUSD/spec
```

**Response**:
```json
{
  "symbol": "XAUUSD",
  "description": "Gold vs US Dollar",
  "contractSize": 100,
  "pipValue": 1.0,
  "pipPosition": 2,
  "minLot": 0.01,
  "maxLot": 50.0,
  "lotStep": 0.01,
  "marginRate": 0.02,
  "swapLong": -2.5,
  "swapShort": 0.5,
  "commission": 0.0,
  "currency": "USD",
  "baseCurrency": "XAU",
  "quoteCurrency": "USD"
}
```

### 3. Invalid Symbol (404 Error)
```bash
curl http://localhost:7999/api/symbols/INVALID/spec
```

**Response**: HTTP 404 - "Symbol not found"

### 4. Invalid Characters (400 Error)
```bash
curl http://localhost:7999/api/symbols/EUR-USD/spec
```

**Response**: HTTP 400 - "Invalid symbol"

## Validation

### Symbol Validation Rules:
- **Pattern**: `^[A-Z0-9]+$` (uppercase letters and numbers only)
- **Invalid characters**: Hyphens, underscores, lowercase letters, special characters
- **Case sensitivity**: Must be uppercase

## Error Responses

| HTTP Status | Error | Reason |
|-------------|-------|--------|
| 400 | Invalid path | URL path is malformed |
| 400 | Invalid symbol | Symbol contains invalid characters |
| 404 | Symbol not found | Symbol not in hardcoded specs map |

## CORS Configuration

All endpoints include CORS headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Headers: Content-Type, Authorization`
- `Content-Type: application/json`

## Testing

A test script has been created at: `test_symbol_spec_api.sh`

Run tests:
```bash
bash test_symbol_spec_api.sh
```

**Requirements**:
- Backend server running on `localhost:7999`
- `curl` and `jq` installed

## Future Enhancements

### TODO: Move to Database
Currently, symbol specifications are hardcoded in the `getSymbolSpecification()` function. Future implementation should:

1. **Create Database Table**:
```sql
CREATE TABLE symbol_specifications (
    symbol VARCHAR(20) PRIMARY KEY,
    description TEXT,
    contract_size DECIMAL(20, 2),
    pip_value DECIMAL(10, 2),
    pip_position INT,
    min_lot DECIMAL(10, 2),
    max_lot DECIMAL(10, 2),
    lot_step DECIMAL(10, 2),
    margin_rate DECIMAL(5, 4),
    swap_long DECIMAL(10, 2),
    swap_short DECIMAL(10, 2),
    commission DECIMAL(10, 2),
    currency VARCHAR(3),
    base_currency VARCHAR(10),
    quote_currency VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

2. **Add Admin Endpoints**:
   - `POST /admin/symbols/specs` - Add new symbol spec
   - `PUT /admin/symbols/{symbol}/spec` - Update spec
   - `DELETE /admin/symbols/{symbol}/spec` - Remove spec

3. **Import from FIX Protocol**:
   - Parse SecurityDefinition messages (35=d)
   - Auto-populate specs from LP responses
   - Sync with YOFX symbol metadata

## Integration Points

### Frontend Integration
```typescript
import { SymbolSpecification } from '@/types/trading';

async function fetchSymbolSpec(symbol: string): Promise<SymbolSpecification> {
  const response = await fetch(`http://localhost:7999/api/symbols/${symbol}/spec`);
  if (!response.ok) {
    throw new Error(`Symbol spec not found: ${symbol}`);
  }
  return response.json();
}

// Usage in components
const spec = await fetchSymbolSpec('EURUSD');
console.log(`Min lot: ${spec.minLot}, Max lot: ${spec.maxLot}`);
console.log(`Margin requirement: ${spec.marginRate * 100}%`);
```

### Use Cases
1. **Order Validation**: Check min/max lot sizes before placing orders
2. **Risk Calculation**: Use margin rate and pip value for position sizing
3. **Swap Display**: Show overnight interest charges
4. **Contract Info**: Display trading specs in UI
5. **Lot Size Calculator**: Calculate appropriate lot sizes based on risk

## Files Modified

1. **Backend**:
   - `backend/api/server.go` (added struct, function, handler)
   - `backend/cmd/server/main.go` (registered route)

2. **Frontend**:
   - `clients/desktop/src/types/trading.ts` (added TypeScript interface)

3. **Documentation**:
   - `docs/SYMBOL_SPEC_API_IMPLEMENTATION.md` (this file)
   - `test_symbol_spec_api.sh` (test script)

## Line Number Summary

| File | Lines Modified | Description |
|------|----------------|-------------|
| `backend/api/server.go` | 3-9 | Added imports |
| `backend/api/server.go` | 611-778 | Added struct, function, handler |
| `backend/cmd/server/main.go` | 752-760 | Registered route |
| `clients/desktop/src/types/trading.ts` | 48-65 | Added TypeScript interface |

## Verification Checklist

- [x] Backend struct defined
- [x] Backend handler implemented
- [x] Route registered in main.go
- [x] TypeScript interface created
- [x] CORS headers configured
- [x] Input validation implemented
- [x] Error handling complete
- [x] Documentation written
- [x] Test script created

## Status

**IMPLEMENTATION COMPLETE** ✓

All code changes have been implemented. The API is ready for testing once the backend server is compiled and started.
