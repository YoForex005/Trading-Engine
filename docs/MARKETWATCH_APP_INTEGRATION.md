# Market Watch → App.tsx Integration Guide

## Overview
This document details the event listeners that App.tsx must implement to handle Market Watch menu actions.

---

## Required Event Listeners in App.tsx

### 1. New Order Dialog (F9)

```typescript
useEffect(() => {
  const handleOpenOrderDialog = (e: CustomEvent) => {
    const { symbol } = e.detail;

    // Open order entry dialog/modal
    setOrderDialogVisible(true);
    setOrderDialogSymbol(symbol);
  };

  window.addEventListener('openOrderDialog', handleOpenOrderDialog as EventListener);
  return () => window.removeEventListener('openOrderDialog', handleOpenOrderDialog as EventListener);
}, []);
```

**Expected Behavior**: Opens the order entry dialog with the selected symbol pre-filled.

---

### 2. Chart Window

```typescript
useEffect(() => {
  const handleOpenChart = (e: CustomEvent) => {
    const { symbol, timeframe = '1H', type = 'candlestick' } = e.detail;

    // Switch to chart view and update chart settings
    setSelectedSymbol(symbol);
    setChartTimeframe(timeframe);
    setChartType(type);

    // Optional: Maximize chart area
    setChartMaximized(true);
  };

  window.addEventListener('openChart', handleOpenChart as EventListener);
  return () => window.removeEventListener('openChart', handleOpenChart as EventListener);
}, []);
```

**Event Detail**:
```typescript
{
  symbol: string;          // e.g., "EURUSD"
  timeframe?: string;      // e.g., "1H", "15m", "TICK"
  type?: ChartType;        // e.g., "candlestick", "line", "area"
}
```

---

### 3. Depth of Market (Alt+B)

```typescript
useEffect(() => {
  const handleOpenDOM = (e: CustomEvent) => {
    const { symbol } = e.detail;

    // Open Depth of Market modal/window
    setDOMVisible(true);
    setDOMSymbol(symbol);
  };

  window.addEventListener('openDepthOfMarket', handleOpenDOM as EventListener);
  return () => window.removeEventListener('openDepthOfMarket', handleOpenDOM as EventListener);
}, []);
```

**Component to Render**:
```typescript
import { DepthOfMarket } from './components/professional/DepthOfMarket';

// In render
{domVisible && (
  <Modal onClose={() => setDOMVisible(false)}>
    <DepthOfMarket symbol={domSymbol} depth={20} />
  </Modal>
)}
```

---

### 4. Popup Prices (F10)

```typescript
useEffect(() => {
  const handlePopupPrices = (e: CustomEvent) => {
    const { symbol } = e.detail;

    // Open floating price window
    setPopupPricesVisible(true);
    setPopupPricesSymbol(symbol);
  };

  window.addEventListener('openPopupPrices', handlePopupPrices as EventListener);
  return () => window.removeEventListener('openPopupPrices', handlePopupPrices as EventListener);
}, []);
```

**Component to Render** (suggested):
```typescript
{popupPricesVisible && (
  <FloatingPriceWindow
    symbol={popupPricesSymbol}
    tick={ticks[popupPricesSymbol]}
    onClose={() => setPopupPricesVisible(false)}
  />
)}
```

---

### 5. Symbols Dialog (Ctrl+U)

```typescript
useEffect(() => {
  const handleOpenSymbolsDialog = () => {
    setSymbolsDialogVisible(true);
  };

  window.addEventListener('openSymbolsDialog', handleOpenSymbolsDialog);
  return () => window.removeEventListener('openSymbolsDialog', handleOpenSymbolsDialog);
}, []);
```

**Component to Render** (suggested):
```typescript
{symbolsDialogVisible && (
  <Modal onClose={() => setSymbolsDialogVisible(false)} size="large">
    <SymbolsManager
      availableSymbols={availableSymbols}
      subscribedSymbols={subscribedSymbols}
      onSubscribe={handleSubscribeSymbol}
      onUnsubscribe={handleUnsubscribeSymbol}
    />
  </Modal>
)}
```

---

### 6. Notifications

```typescript
useEffect(() => {
  const handleNotification = (e: CustomEvent) => {
    const { message, type, duration = 3000 } = e.detail;

    // Show toast notification
    showToast(message, type, duration);
  };

  window.addEventListener('showNotification', handleNotification as EventListener);
  return () => window.removeEventListener('showNotification', handleNotification as EventListener);
}, []);
```

**Event Detail**:
```typescript
{
  message: string;           // e.g., "Buy order executed: EURUSD"
  type: 'success' | 'error' | 'info';
  duration?: number;         // ms, default 3000
}
```

---

## Complete App.tsx Integration Example

```typescript
import { useState, useEffect } from 'react';

function App() {
  // State for modals/windows
  const [orderDialogVisible, setOrderDialogVisible] = useState(false);
  const [orderDialogSymbol, setOrderDialogSymbol] = useState('');
  const [domVisible, setDOMVisible] = useState(false);
  const [domSymbol, setDOMSymbol] = useState('');
  const [popupPricesVisible, setPopupPricesVisible] = useState(false);
  const [popupPricesSymbol, setPopupPricesSymbol] = useState('');
  const [symbolsDialogVisible, setSymbolsDialogVisible] = useState(false);
  const [chartMaximized, setChartMaximized] = useState(false);

  // Event listeners for Market Watch actions
  useEffect(() => {
    // 1. New Order Dialog (F9)
    const handleOpenOrderDialog = (e: CustomEvent) => {
      const { symbol } = e.detail;
      setOrderDialogVisible(true);
      setOrderDialogSymbol(symbol);
    };

    // 2. Chart Window
    const handleOpenChart = (e: CustomEvent) => {
      const { symbol, timeframe = '1H', type = 'candlestick' } = e.detail;
      setSelectedSymbol(symbol);
      setChartTimeframe(timeframe);
      setChartType(type);
      setChartMaximized(true);
    };

    // 3. Depth of Market (Alt+B)
    const handleOpenDOM = (e: CustomEvent) => {
      const { symbol } = e.detail;
      setDOMVisible(true);
      setDOMSymbol(symbol);
    };

    // 4. Popup Prices (F10)
    const handlePopupPrices = (e: CustomEvent) => {
      const { symbol } = e.detail;
      setPopupPricesVisible(true);
      setPopupPricesSymbol(symbol);
    };

    // 5. Symbols Dialog (Ctrl+U)
    const handleOpenSymbolsDialog = () => {
      setSymbolsDialogVisible(true);
    };

    // 6. Notifications
    const handleNotification = (e: CustomEvent) => {
      const { message, type, duration = 3000 } = e.detail;
      showToast(message, type, duration);
    };

    // Register all listeners
    window.addEventListener('openOrderDialog', handleOpenOrderDialog as EventListener);
    window.addEventListener('openChart', handleOpenChart as EventListener);
    window.addEventListener('openDepthOfMarket', handleOpenDOM as EventListener);
    window.addEventListener('openPopupPrices', handlePopupPrices as EventListener);
    window.addEventListener('openSymbolsDialog', handleOpenSymbolsDialog);
    window.addEventListener('showNotification', handleNotification as EventListener);

    // Cleanup
    return () => {
      window.removeEventListener('openOrderDialog', handleOpenOrderDialog as EventListener);
      window.removeEventListener('openChart', handleOpenChart as EventListener);
      window.removeEventListener('openDepthOfMarket', handleOpenDOM as EventListener);
      window.removeEventListener('openPopupPrices', handlePopupPrices as EventListener);
      window.removeEventListener('openSymbolsDialog', handleOpenSymbolsDialog);
      window.removeEventListener('showNotification', handleNotification as EventListener);
    };
  }, []);

  return (
    <div className="app">
      {/* Market Watch Panel */}
      <MarketWatchPanel
        ticks={ticks}
        allSymbols={allSymbols}
        selectedSymbol={selectedSymbol}
        onSymbolSelect={setSelectedSymbol}
      />

      {/* Chart Area */}
      <div className={chartMaximized ? 'chart-maximized' : 'chart-normal'}>
        <TradingChart
          symbol={selectedSymbol}
          chartType={chartType}
          timeframe={timeframe}
        />
      </div>

      {/* Modals/Windows */}
      {orderDialogVisible && (
        <OrderEntryDialog
          symbol={orderDialogSymbol}
          onClose={() => setOrderDialogVisible(false)}
        />
      )}

      {domVisible && (
        <Modal onClose={() => setDOMVisible(false)}>
          <DepthOfMarket symbol={domSymbol} depth={20} />
        </Modal>
      )}

      {popupPricesVisible && (
        <FloatingPriceWindow
          symbol={popupPricesSymbol}
          tick={ticks[popupPricesSymbol]}
          onClose={() => setPopupPricesVisible(false)}
        />
      )}

      {symbolsDialogVisible && (
        <Modal onClose={() => setSymbolsDialogVisible(false)} size="large">
          <SymbolsManager
            availableSymbols={availableSymbols}
            subscribedSymbols={subscribedSymbols}
          />
        </Modal>
      )}
    </div>
  );
}
```

---

## Event Flow Diagram

```
┌──────────────────────┐
│  Market Watch Panel  │
│  (Right-Click Menu)  │
└──────────┬───────────┘
           │
           │ User clicks "New Order (F9)"
           ▼
┌──────────────────────────────────────────┐
│  marketWatchActions.ts                   │
│  openNewOrderDialog(symbol)              │
└──────────┬───────────────────────────────┘
           │
           │ window.dispatchEvent(...)
           ▼
┌──────────────────────────────────────────┐
│  App.tsx                                 │
│  useEffect event listener                │
│  handleOpenOrderDialog()                 │
└──────────┬───────────────────────────────┘
           │
           │ setState(...)
           ▼
┌──────────────────────────────────────────┐
│  Order Entry Dialog Rendered             │
│  with symbol pre-filled                  │
└──────────────────────────────────────────┘
```

---

## Testing Integration

### 1. Test New Order Dialog
```typescript
// In browser console
window.dispatchEvent(new CustomEvent('openOrderDialog', {
  detail: { symbol: 'EURUSD' }
}));

// Expected: Order dialog opens with EURUSD selected
```

### 2. Test Chart Window
```typescript
window.dispatchEvent(new CustomEvent('openChart', {
  detail: { symbol: 'GBPUSD', timeframe: '15m', type: 'candlestick' }
}));

// Expected: Chart switches to GBPUSD on 15m timeframe
```

### 3. Test Depth of Market
```typescript
window.dispatchEvent(new CustomEvent('openDepthOfMarket', {
  detail: { symbol: 'XAUUSD' }
}));

// Expected: DOM modal opens showing XAUUSD order book
```

### 4. Test Notifications
```typescript
window.dispatchEvent(new CustomEvent('showNotification', {
  detail: {
    message: 'Order executed successfully',
    type: 'success',
    duration: 3000
  }
}));

// Expected: Success toast appears for 3 seconds
```

---

## TypeScript Event Types

Create a types file for custom events:

```typescript
// src/types/events.ts

export interface OpenOrderDialogEvent extends CustomEvent {
  detail: {
    symbol: string;
  };
}

export interface OpenChartEvent extends CustomEvent {
  detail: {
    symbol: string;
    timeframe?: string;
    type?: 'candlestick' | 'line' | 'area' | 'bar';
  };
}

export interface OpenDOMEvent extends CustomEvent {
  detail: {
    symbol: string;
  };
}

export interface OpenPopupPricesEvent extends CustomEvent {
  detail: {
    symbol: string;
  };
}

export interface ShowNotificationEvent extends CustomEvent {
  detail: {
    message: string;
    type: 'success' | 'error' | 'info';
    duration?: number;
  };
}

// Extend WindowEventMap
declare global {
  interface WindowEventMap {
    'openOrderDialog': OpenOrderDialogEvent;
    'openChart': OpenChartEvent;
    'openDepthOfMarket': OpenDOMEvent;
    'openPopupPrices': OpenPopupPricesEvent;
    'openSymbolsDialog': Event;
    'showNotification': ShowNotificationEvent;
  }
}
```

Then use typed event listeners:

```typescript
window.addEventListener('openOrderDialog', (e) => {
  // TypeScript now knows e.detail.symbol exists
  const { symbol } = e.detail;
});
```

---

## Summary

✅ **6 custom events** dispatched from Market Watch
✅ **All events include necessary data** (symbol, timeframe, etc.)
✅ **Type-safe event handling** with TypeScript declarations
✅ **Clean separation of concerns** (Market Watch → Service → App.tsx)
✅ **Easy to test** with manual event dispatch

Implement these event listeners in App.tsx to complete the Market Watch integration.
