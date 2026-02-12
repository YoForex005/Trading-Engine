# Context Menu Integration Guide

## Quick Start

Both context menus are now fully implemented and integrated. This guide shows how to connect them to parent components for full functionality.

## Chart Context Menu - Parent Component Integration

### Option 1: Pass Chart Type/Timeframe as Props (Recommended)

If TradingChart is wrapped by a parent component that controls chart settings:

```typescript
// Parent component (e.g., TradingDashboard.tsx)
import { TradingChart, ChartType, Timeframe } from './TradingChart';

function TradingDashboard() {
  const [chartType, setChartType] = useState<ChartType>('candlestick');
  const [timeframe, setTimeframe] = useState<Timeframe>('M1');
  const [symbol, setSymbol] = useState('EURUSD');

  return (
    <TradingChart
      symbol={symbol}
      chartType={chartType}
      timeframe={timeframe}
      onChangeChartType={setChartType}  // Add this prop
      onChangeTimeframe={setTimeframe}  // Add this prop
      {...otherProps}
    />
  );
}
```

### Option 2: Internal State Management

If you want TradingChart to manage its own settings:

```typescript
// In TradingChart.tsx
export function TradingChart({
    symbol,
    initialChartType = 'candlestick',
    initialTimeframe = 'M1',
    ...
}: ChartProps) {
    const [chartType, setChartType] = useState<ChartType>(initialChartType);
    const [timeframe, setTimeframe] = useState<Timeframe>(initialTimeframe);

    // Update context menu handlers:
    onChangeChartType={(type) => {
        setChartType(type);
    }}
    onChangeTimeframe={(tf) => {
        setTimeframe(tf);
    }}
}
```

## MarketWatch Context Menu - Order Entry Integration

### Method 1: Modal Dialog (Recommended)

```typescript
// Parent component (e.g., TradingDashboard.tsx)
import { MarketWatch } from './MarketWatch';
import { OrderEntry } from './OrderEntry';

function TradingDashboard() {
  const [selectedSymbol, setSelectedSymbol] = useState('EURUSD');
  const [orderEntryOpen, setOrderEntryOpen] = useState(false);
  const [orderEntrySymbol, setOrderEntrySymbol] = useState<string | null>(null);
  const account = useAppStore(state => state.account);

  const handleOpenOrderEntry = (symbol: string) => {
    setOrderEntrySymbol(symbol);
    setOrderEntryOpen(true);
  };

  return (
    <>
      <MarketWatch
        selectedSymbol={selectedSymbol}
        onSelectSymbol={setSelectedSymbol}
        onOpenOrderEntry={handleOpenOrderEntry}  // Add this prop
      />

      {/* Order Entry Modal */}
      {orderEntryOpen && orderEntrySymbol && account && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#131722] rounded-lg shadow-2xl p-6 max-w-2xl w-full">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-bold text-white">
                New Order - {orderEntrySymbol}
              </h2>
              <button
                onClick={() => setOrderEntryOpen(false)}
                className="text-zinc-400 hover:text-white"
              >
                ✕
              </button>
            </div>
            <OrderEntry
              symbol={orderEntrySymbol}
              accountId={account.id}
              balance={account.balance}
              onOrderPlaced={() => {
                setOrderEntryOpen(false);
                // Refresh positions/orders
              }}
            />
          </div>
        </div>
      )}
    </>
  );
}
```

### Method 2: Side Panel

```typescript
// Parent component
function TradingDashboard() {
  const [orderPanelOpen, setOrderPanelOpen] = useState(false);
  const [orderPanelSymbol, setOrderPanelSymbol] = useState<string | null>(null);

  const handleOpenOrderEntry = (symbol: string) => {
    setOrderPanelSymbol(symbol);
    setOrderPanelOpen(true);
  };

  return (
    <div className="flex h-screen">
      <MarketWatch
        onOpenOrderEntry={handleOpenOrderEntry}
        {...props}
      />

      {/* Sliding Side Panel */}
      <div
        className={`absolute right-0 top-0 h-full w-96 bg-[#131722] border-l border-zinc-800 transform transition-transform duration-300 ${
          orderPanelOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        {orderPanelSymbol && (
          <OrderEntry
            symbol={orderPanelSymbol}
            {...props}
          />
        )}
      </div>
    </div>
  );
}
```

## Complete Example: Full Integration

```typescript
// TradingDashboard.tsx
import React, { useState } from 'react';
import { TradingChart, ChartType, Timeframe } from './components/TradingChart';
import { MarketWatch } from './components/MarketWatch';
import { OrderEntry } from './components/OrderEntry';
import { useAppStore } from './store/useAppStore';

export function TradingDashboard() {
  // Chart settings
  const [chartType, setChartType] = useState<ChartType>('candlestick');
  const [timeframe, setTimeframe] = useState<Timeframe>('M1');
  const [selectedSymbol, setSelectedSymbol] = useState('EURUSD');

  // Order entry modal
  const [orderEntryOpen, setOrderEntryOpen] = useState(false);
  const [orderEntrySymbol, setOrderEntrySymbol] = useState<string | null>(null);

  // Global state
  const account = useAppStore(state => state.account);
  const ticks = useAppStore(state => state.ticks);

  const currentTick = ticks[selectedSymbol];

  const handleOpenOrderEntry = (symbol: string) => {
    setOrderEntrySymbol(symbol);
    setOrderEntryOpen(true);
  };

  return (
    <div className="flex h-screen bg-[#131722]">
      {/* Left Sidebar - Market Watch */}
      <div className="w-64 border-r border-zinc-800">
        <MarketWatch
          selectedSymbol={selectedSymbol}
          onSelectSymbol={setSelectedSymbol}
          onOpenOrderEntry={handleOpenOrderEntry}
        />
      </div>

      {/* Main Chart Area */}
      <div className="flex-1">
        <TradingChart
          symbol={selectedSymbol}
          chartType={chartType}
          timeframe={timeframe}
          currentPrice={
            currentTick
              ? { bid: currentTick.bid, ask: currentTick.ask }
              : undefined
          }
          // These props need to be added to TradingChart interface
          onChangeChartType={setChartType}
          onChangeTimeframe={setTimeframe}
        />
      </div>

      {/* Order Entry Modal */}
      {orderEntryOpen && orderEntrySymbol && account && (
        <div
          className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
          onClick={() => setOrderEntryOpen(false)}
        >
          <div
            className="bg-[#131722] rounded-lg shadow-2xl p-6 max-w-2xl w-full mx-4"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-bold text-white">
                New Order - {orderEntrySymbol}
              </h2>
              <button
                onClick={() => setOrderEntryOpen(false)}
                className="text-zinc-400 hover:text-white text-2xl"
              >
                ×
              </button>
            </div>
            <OrderEntry
              symbol={orderEntrySymbol}
              currentBid={ticks[orderEntrySymbol]?.bid}
              currentAsk={ticks[orderEntrySymbol]?.ask}
              accountId={parseInt(account.accountId || '0')}
              balance={account.balance}
              onOrderPlaced={() => {
                setOrderEntryOpen(false);
                // Optionally refresh positions/orders
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
```

## TradingChart Props Update

To fully support the context menu, update the TradingChart props interface:

```typescript
// In TradingChart.tsx
interface ChartProps {
    symbol: string;
    currentPrice?: { bid: number; ask: number };
    chartType?: ChartType;
    timeframe?: Timeframe;
    positions?: any[];
    onClosePosition?: (id: number) => void;
    onModifyPosition?: (id: number, sl: number, tp: number) => void;

    // Add these for context menu integration:
    onChangeChartType?: (type: ChartType) => void;
    onChangeTimeframe?: (tf: Timeframe) => void;
}

export function TradingChart({
    // ... existing props
    onChangeChartType,
    onChangeTimeframe,
}: ChartProps) {
    // Update context menu handlers:
    <ChartContextMenu
        onChangeChartType={(type) => {
            if (onChangeChartType) {
                onChangeChartType(type);
            } else {
                // Fallback: log or show notification
                console.log('Change chart type to:', type);
            }
        }}
        onChangeTimeframe={(tf) => {
            if (onChangeTimeframe) {
                onChangeTimeframe(tf);
            } else {
                console.log('Change timeframe to:', tf);
            }
        }}
        // ... other props
    />
}
```

## Testing the Integration

### Test Chart Context Menu:
1. Right-click chart background (not on a drawing)
2. Hover over "Timeframe" → Select M5 → Chart should update
3. Hover over "Chart Type" → Select Line Chart → Chart should update
4. Click "One-Click Trading" → Panel should appear on chart
5. Click "Show Grid" → Grid should toggle
6. Click "Show Volumes" → Volume bars should toggle
7. Hover over "Indicators" → Select RSI → Indicator should appear
8. Click "Save As Picture" → PNG should download
9. Click "Print Chart" → Print dialog should open

### Test MarketWatch Context Menu:
1. Right-click on a symbol row (e.g., EURUSD)
2. Click "New Order" → Order Entry modal should open with EURUSD selected
3. Click "Chart Window" → Chart should switch to that symbol
4. Click "Hide EURUSD" → Symbol should disappear from list
5. Right-click another symbol → Click "Show All Symbols" → EURUSD should reappear
6. Hover over "Columns" → Toggle "Spread" → Spread column should disappear
7. Hover over "Columns" → Toggle "Spread" again → Spread column should reappear

## Common Issues & Solutions

### Issue: Context menu doesn't open
**Solution:** Check that event handlers are properly attached and not blocked by other elements.

```typescript
// Ensure chart container has context menu handler
chartContainerRef.current.addEventListener('contextmenu', handleChartContextMenu);
```

### Issue: Context menu doesn't close
**Solution:** Make sure click-outside and mouse-leave handlers are working.

```typescript
// Add global click handler
window.addEventListener('click', () => setContextMenu(null));

// Add mouse leave to context menu
<div onMouseLeave={onClose}>...</div>
```

### Issue: Timeframe/Chart type changes don't update chart
**Solution:** Ensure parent component is passing updated props to TradingChart.

```typescript
// Parent must update state and pass as prop
const [chartType, setChartType] = useState<ChartType>('candlestick');

<TradingChart chartType={chartType} />
```

### Issue: OrderEntry modal doesn't have current prices
**Solution:** Pass tick data to OrderEntry.

```typescript
const currentTick = useAppStore(state => state.ticks[orderEntrySymbol]);

<OrderEntry
  currentBid={currentTick?.bid}
  currentAsk={currentTick?.ask}
/>
```

## Next Steps

1. **Implement Properties Dialog**: Create a comprehensive chart settings dialog
2. **Add Tick Chart Component**: Implement tick chart view
3. **Add Depth of Market**: Implement Level II order book display
4. **Symbol Specification Dialog**: Create modal with full symbol details
5. **Persist Column Visibility**: Save MarketWatch column settings to localStorage
6. **Keyboard Shortcuts**: Add keyboard shortcuts for context menu actions
7. **Right-Click on Drawings**: Ensure drawing context menu takes precedence over chart context menu

## API Reference

### ChartContextMenu Props
```typescript
interface ChartContextMenuProps {
    visible: boolean;
    x: number;
    y: number;
    currentChartType: ChartType;
    currentTimeframe: Timeframe;
    onChangeChartType: (type: ChartType) => void;
    onChangeTimeframe: (tf: Timeframe) => void;
    onToggleOneClickTrading: () => void;
    onToggleGrid: () => void;
    onToggleVolumes: () => void;
    onOpenProperties: () => void;
    onSaveAsPicture: () => void;
    onPrintChart: () => void;
    onClose: () => void;
    oneClickTrading?: boolean;
    showGrid?: boolean;
    showVolumes?: boolean;
}
```

### MarketWatchContextMenu Props
```typescript
interface MarketWatchContextMenuProps {
    visible: boolean;
    x: number;
    y: number;
    symbol: string;
    onNewOrder: (symbol: string) => void;
    onChartWindow: (symbol: string) => void;
    onTickChart: (symbol: string) => void;
    onDepthOfMarket: (symbol: string) => void;
    onSymbolSpecification: (symbol: string) => void;
    onHideSymbol: (symbol: string) => void;
    onShowAllSymbols: () => void;
    onToggleColumn: (column: string) => void;
    onClose: () => void;
    columnVisibility?: {
        spread: boolean;
        highLow: boolean;
        time: boolean;
    };
}
```
