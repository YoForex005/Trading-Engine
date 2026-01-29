# RoutingIndicator - Quick Start

## Installation

Already created at:
```
clients/desktop/src/components/RoutingIndicator.tsx
```

## Import

```tsx
import { RoutingIndicator } from './components/RoutingIndicator';
// Or with types:
import { RoutingIndicator, type RoutingDecision, type RoutingPath } from './components/RoutingIndicator';
```

## Basic Example

```tsx
<RoutingIndicator
  symbol="EUR/USD"
  volume={1.5}
  accountId="account-123"
  side="BUY"
/>
```

## Props at a Glance

```tsx
<RoutingIndicator
  symbol="EUR/USD"        // Required: Trading symbol
  volume={1.0}            // Required: Order volume (lots)
  accountId="acc-123"     // Required: Account ID
  side="BUY"              // Optional: Order side (default: BUY)
  className="mb-4"        // Optional: Additional CSS
/>
```

## Color Scheme Quick Reference

| Routing | Color | Meaning |
|---------|-------|---------|
| A-Book | Blue | Direct LP - Best execution |
| B-Book | Amber | Internal market maker |
| C-Book | Orange | Alternative LP provider |
| Partial | Purple | Mixed routing with hedge |

## Exposure Warnings

```
Green (0-74%)    : Safe - Within limits
Yellow (75-89%)  : Warning - High exposure
Red (≥90%)       : Critical - Exceeds limits
```

## Real-World Usage

### In Order Entry Panel

```tsx
export function OrderPanel() {
  const [volume, setVolume] = useState(1.0);
  const symbol = "EUR/USD";
  const accountId = "acc-123";

  return (
    <>
      <input
        type="number"
        value={volume}
        onChange={(e) => setVolume(parseFloat(e.target.value))}
        placeholder="Volume"
      />

      <RoutingIndicator
        symbol={symbol}
        volume={volume}
        accountId={accountId}
      />

      <button>Place Order</button>
    </>
  );
}
```

### With Side Selection

```tsx
const [side, setSide] = useState<'BUY' | 'SELL'>('BUY');

<RoutingIndicator
  symbol={symbol}
  volume={volume}
  accountId={accountId}
  side={side}
/>
```

## What the Component Does

1. ✓ Fetches routing decision from backend
2. ✓ Shows which path (A/B/C-Book, Partial) will be used
3. ✓ Displays LP name if routed to specific provider
4. ✓ Shows hedge % for partial hedges
5. ✓ Warns about high exposure levels
6. ✓ Explains routing decision with reason
7. ✓ Shows confidence score
8. ✓ Updates in real-time as volume changes

## API Endpoint Required

Backend must implement:
```
GET /api/routing/preview
  ?symbol=EUR/USD
  &volume=1.0
  &accountId=account-123
  &side=BUY
```

Returns:
```json
{
  "path": "ABOOK",
  "lpName": "Prime Liquidity",
  "reason": "High volume qualified for premium LP",
  "confidence": 0.95,
  "exposureImpact": {
    "current": 500000,
    "after": 501500,
    "limit": 1000000,
    "utilizationPercent": 50.15,
    "isWarning": false,
    "isCritical": false
  }
}
```

## Visual States

### Normal State
Shows routing decision with exposure metrics and color-coded indicator.

### Loading State
Displays spinning loader: "Calculating routing..."

### Error State
Shows error message: "Unable to determine routing"

### Hidden State
Component not rendered when:
- `symbol` is empty
- `volume` is 0 or negative

## Common Integration Points

1. **OrderEntry Component** - Show before placing orders
2. **OrderEntryPanel** - Display in advanced order interface
3. **AdminPanel** - Show for account management
4. **Risk Dashboard** - Display alongside risk metrics

## Styling

Component uses Tailwind CSS. To customize spacing:

```tsx
<RoutingIndicator
  symbol="EUR/USD"
  volume={1.0}
  accountId="acc-123"
  className="mb-6 border-2 border-gray-200"
/>
```

## Key Features

- ✓ Real-time routing preview
- ✓ Exposure impact analysis
- ✓ Color-coded routing paths
- ✓ Interactive tooltips
- ✓ Loading and error states
- ✓ Debounced API calls
- ✓ TypeScript support
- ✓ Tailwind CSS styling
- ✓ Responsive design

## Type Safety

All exported types available:

```tsx
import type {
  RoutingPath,           // 'ABOOK' | 'BBOOK' | 'CBOOK' | 'PARTIAL'
  RoutingDecision,       // Full response type
  RoutingIndicatorProps, // Component props
} from './components/RoutingIndicator';
```

## Next Steps

1. Implement backend `/api/routing/preview` endpoint
2. Add component to order entry forms
3. Test with various volumes and symbols
4. Monitor exposure warnings in production
5. Gather user feedback on routing clarity
