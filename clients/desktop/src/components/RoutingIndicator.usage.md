# RoutingIndicator Component Usage Guide

## Overview

The `RoutingIndicator` component displays order routing decisions to users before order submission. It shows which routing path (A-Book/B-Book/C-Book/Partial) will be used and provides exposure impact warnings.

## Component Location

```
/Users/epic1st/Documents/trading engine/clients/desktop/src/components/RoutingIndicator.tsx
```

## Basic Usage

```tsx
import { RoutingIndicator } from './components/RoutingIndicator';

export function OrderPanel() {
  const [symbol, setSymbol] = useState('EUR/USD');
  const [volume, setVolume] = useState(1.0);
  const [accountId, setAccountId] = useState('account-123');

  return (
    <div className="p-4">
      <input
        value={symbol}
        onChange={(e) => setSymbol(e.target.value)}
        placeholder="Symbol"
      />
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
        side="BUY"
      />
    </div>
  );
}
```

## Props

| Prop | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `symbol` | `string` | Yes | - | Trading symbol (e.g., 'EUR/USD', 'GBPUSD') |
| `volume` | `number` | Yes | - | Order volume in lots (must be > 0) |
| `accountId` | `string \| number` | Yes | - | Account identifier for routing decision |
| `side` | `'BUY' \| 'SELL'` | No | `'BUY'` | Order direction |
| `className` | `string` | No | `''` | Additional CSS classes |

## Features

### 1. Routing Path Display

Shows the order routing decision with visual color coding:
- **A-Book (Blue)**: Direct LP routing - best execution for transparent clients
- **B-Book (Amber)**: Internal book - market maker as counterparty
- **C-Book (Orange)**: Market maker routing - alternative LP provider
- **Partial (Purple)**: Mixed routing with hedge ratio

### 2. Confidence Score

Displays a confidence percentage (0-100%) showing the certainty of the routing decision.

### 3. Liquidity Provider Details

For A-Book routing, displays the target LP name.

### 4. Hedge Percentage

For Partial routing, shows the hedge ratio percentage.

### 5. Exposure Impact Analysis

Displays detailed exposure metrics:
- Current exposure
- Exposure after order
- Maximum allowed exposure
- Utilization percentage with progress bar
- Color-coded status (Green/Yellow/Red)

### 6. Risk Warnings

- **Warning Level (≥75% utilization)**: Yellow indicator with warning message
- **Critical Level (≥90% utilization)**: Red indicator with critical message

### 7. Routing Reason

Shows explanation text for why this particular routing was selected.

### 8. Interactive Tooltips

Hover over icons to see additional context about routing and exposure.

## Visual States

### Loading State

```
Calculating routing...
```
Shown while fetching the routing decision from the backend.

### Error State

```
Unable to determine routing
```
Displayed if the API call fails or routing decision cannot be made.

### Hidden State

Component doesn't render if:
- `symbol` is empty
- `volume` is 0 or negative

## API Endpoint

The component fetches from:

```
GET /api/routing/preview?symbol={symbol}&volume={volume}&accountId={accountId}&side={side}
```

### Response Format

```typescript
interface RoutingDecision {
  path: 'ABOOK' | 'BBOOK' | 'CBOOK' | 'PARTIAL';
  lpName?: string;
  hedgePercentage?: number;
  reason: string;
  confidence: number;
  exposureImpact: {
    current: number;
    after: number;
    limit: number;
    utilizationPercent: number;
    isWarning: boolean;
    isCritical: boolean;
  };
}
```

## Styling

The component uses **Tailwind CSS** for all styling with:
- Responsive design
- Color-coded routing paths
- Smooth transitions and animations
- Accessible contrast ratios

## Integration Points

### 1. OrderEntry Component

Integrate before order submission:

```tsx
<RoutingIndicator
  symbol={symbol}
  volume={volume}
  accountId={accountId}
  side={side}
  className="mb-4"
/>

<button onClick={handlePlaceOrder}>
  Place Order
</button>
```

### 2. OrderEntryPanel

Add to advanced order panel for A-Book/B-Book visibility.

### 3. AdminPanel

Display routing for account management and execution mode visibility.

## Behavior

### Automatic Updates

The component automatically refetches routing when:
- Symbol changes
- Volume changes
- Account ID changes
- Order side changes

### Debouncing

Requests are debounced by 500ms to avoid excessive API calls while user is typing/adjusting values.

### Error Handling

- Network errors are caught and displayed
- Invalid parameters (empty symbol, zero volume) hide the component
- Failed API responses show user-friendly error message

## Color Scheme

| Routing | Background | Border | Badge | Icon |
|---------|-----------|--------|-------|------|
| A-Book | Blue-50 | Blue-300 | Blue-100 | Blue-600 |
| B-Book | Amber-50 | Amber-300 | Amber-100 | Amber-600 |
| C-Book | Orange-50 | Orange-300 | Orange-100 | Orange-600 |
| Partial | Purple-50 | Purple-300 | Purple-100 | Purple-600 |

## Example Integration

```tsx
import { useState } from 'react';
import { RoutingIndicator } from './components/RoutingIndicator';
import { OrderEntry } from './components/OrderEntry';

export function AdvancedOrderPanel({ accountId, symbol, currentBid, currentAsk }) {
  const [volume, setVolume] = useState(1.0);
  const [side, setSide] = useState<'BUY' | 'SELL'>('BUY');

  return (
    <div className="p-6 bg-white rounded-lg shadow">
      <h2 className="text-lg font-bold mb-4">Advanced Order Entry</h2>

      {/* Input fields */}
      <div className="mb-4">
        <label className="block text-sm font-medium mb-2">Volume</label>
        <input
          type="number"
          value={volume}
          onChange={(e) => setVolume(parseFloat(e.target.value))}
          className="w-full px-3 py-2 border rounded"
          step="0.01"
          min="0"
        />
      </div>

      {/* Routing Indicator */}
      <div className="mb-6">
        <RoutingIndicator
          symbol={symbol}
          volume={volume}
          accountId={accountId}
          side={side}
        />
      </div>

      {/* Order Entry */}
      <OrderEntry
        symbol={symbol}
        currentBid={currentBid}
        currentAsk={currentAsk}
        accountId={parseInt(accountId)}
        balance={10000}
      />
    </div>
  );
}
```

## Backend Implementation Notes

To implement the `/api/routing/preview` endpoint, consider:

1. **Routing Logic**
   - Check current account exposure for the symbol
   - Determine best LP based on volume
   - Calculate hedge ratio if applicable
   - Generate confidence score

2. **Exposure Calculation**
   - Sum all open positions for the symbol
   - Add new order volume
   - Check against symbol-level limits
   - Calculate utilization percentage

3. **Response Data**
   - Include reason string explaining the routing decision
   - Set confidence based on data quality and certainty
   - Flag warnings at 75% and critical at 90% utilization

## Performance Considerations

- Component uses `useEffect` with proper dependency array
- Debounces API requests by 500ms
- Cleans up timers on unmount
- Memoized color configuration
- Lightweight re-renders on state changes

## Accessibility

- Semantic HTML structure
- ARIA-compliant tooltips
- Color + icon indicators (not just color)
- Clear text labels and descriptions
- Keyboard-navigable tooltips

## Future Enhancements

Potential additions:
- Caching of routing decisions
- Historical routing statistics
- User preference for preferred routing paths
- A/B testing different routing strategies
- Real-time routing availability status
