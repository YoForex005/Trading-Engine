# Chart Integration - Quick Reference

## Command Reference

### Chart Manager Commands
```typescript
import { chartManager } from '@/services';

chartManager.toggleCrosshair()       // Toggle crosshair on/off
chartManager.zoomIn()                // Zoom in (wider candles)
chartManager.zoomOut()               // Zoom out (narrower candles)
chartManager.setZoom(barSpacing)     // Set specific zoom (1-50)
chartManager.resetZoom()             // Reset to default (6px)
chartManager.fitContent()            // Fit all candles
chartManager.scrollToRealtime()      // Scroll to latest
chartManager.isCrosshairEnabled()    // Check crosshair state
chartManager.getBarSpacing()         // Get current zoom level
```

### Drawing Manager Commands
```typescript
import { drawingManager } from '@/services';

// Start drawing
const id = drawingManager.startDrawing('trendline', '#3b82f6');

// Add points
drawingManager.addPoint(timestamp, price);

// Complete drawing
const drawing = drawingManager.finishDrawing();

// Manage drawings
drawingManager.deleteDrawing(id);
drawingManager.clearAllDrawings();
drawingManager.getDrawings();
drawingManager.getActiveDrawing();

// Persistence
drawingManager.saveToStorage('EURUSD');
drawingManager.loadFromStorage('EURUSD');
```

### Indicator Manager Commands
```typescript
import { indicatorManager } from '@/services';

// Add indicator
indicatorManager.addIndicator({
  id: 'ma-1',
  name: 'Moving Average',
  type: 'ma',
  parameters: { period: 20 },
  visible: true
});

// Manage indicators
indicatorManager.removeIndicator('ma-1');
indicatorManager.toggleIndicator('ma-1');
indicatorManager.updateIndicator('ma-1', { period: 50 });
indicatorManager.getIndicators();
indicatorManager.clearAll();

// Persistence
indicatorManager.saveToStorage('EURUSD');
indicatorManager.loadFromStorage('EURUSD');
```

## Command Bus Integration

### Dispatching Commands
```typescript
import { commandBus } from '@/services';

// Chart controls
commandBus.dispatch({ type: 'TOGGLE_CROSSHAIR' });
commandBus.dispatch({ type: 'ZOOM_IN' });
commandBus.dispatch({ type: 'ZOOM_OUT' });
commandBus.dispatch({ type: 'FIT_CONTENT' });

// Drawing tools
commandBus.dispatch({ type: 'SELECT_TRENDLINE' });
commandBus.dispatch({ type: 'SELECT_HLINE' });
commandBus.dispatch({ type: 'SELECT_VLINE' });
commandBus.dispatch({ type: 'SELECT_TEXT' });

// Indicators
commandBus.dispatch({ type: 'OPEN_INDICATORS' });
```

### Subscribing to Commands
```typescript
import { commandBus } from '@/services';

const unsubscribe = commandBus.subscribe('ZOOM_IN', () => {
  console.log('Zoom in command received');
});

// Cleanup
unsubscribe();
```

## Component Usage

### ChartWithIndicators (Recommended)
```tsx
import { ChartWithIndicators } from '@/components';

<ChartWithIndicators
  symbol="EURUSD"
  currentPrice={{ bid: 1.0850, ask: 1.0852 }}
  chartType="candlestick"
  timeframe="1m"
  positions={positions}
  onClosePosition={handleClose}
  onModifyPosition={handleModify}
/>
```

### Indicator Navigator
```tsx
import { IndicatorNavigator } from '@/components';

<IndicatorNavigator
  isOpen={showDialog}
  onClose={() => setShowDialog(false)}
  onAddIndicator={(indicator) => {
    indicatorManager.addIndicator({
      id: indicator.id,
      name: indicator.name,
      type: indicator.id.split('-')[0],
      parameters: indicator.parameters || {},
      visible: true
    });
  }}
/>
```

## Drawing Types

### Trendline (2 points)
```typescript
drawingManager.startDrawing('trendline');
drawingManager.addPoint(timestamp1, price1);
drawingManager.addPoint(timestamp2, price2);
// Auto-completes after 2 points
```

### Horizontal Line (1 point)
```typescript
drawingManager.startDrawing('hline');
drawingManager.addPoint(timestamp, price);
// Auto-completes after 1 point
```

### Vertical Line (1 point)
```typescript
drawingManager.startDrawing('vline');
drawingManager.addPoint(timestamp, price);
// Auto-completes after 1 point
```

### Text Annotation (1 point + text)
```typescript
const id = drawingManager.startDrawing('text');
drawingManager.addPoint(timestamp, price);
const drawing = drawingManager.finishDrawing();
drawing.text = 'My annotation';
```

## Indicator Types

### Available Now
- `ma` / `sma` - Simple Moving Average
- `ema` - Exponential Moving Average
- `rsi` - Relative Strength Index
- `macd` - MACD (line only)

### Parameters

**Moving Average:**
```typescript
{ period: 20, type: 'SMA' }
```

**EMA:**
```typescript
{ period: 12 }
```

**RSI:**
```typescript
{ period: 14 }
```

**MACD:**
```typescript
{ fast: 12, slow: 26, signal: 9 }
```

## Events

### Drawing Events
```typescript
window.addEventListener('drawing:saved', (e) => {
  console.log('Drawing saved:', e.detail);
});

window.addEventListener('drawing:deleted', (e) => {
  console.log('Drawing deleted:', e.detail);
});

window.addEventListener('drawings:cleared', () => {
  console.log('All drawings cleared');
});
```

## Keyboard Shortcuts (Future)
These will be implemented by Agent 2:

- `C` - Toggle crosshair
- `+` / `=` - Zoom in
- `-` - Zoom out
- `T` - Trendline tool
- `H` - Horizontal line tool
- `V` - Vertical line tool
- `Shift+T` - Text tool
- `I` - Open indicators
- `Escape` - Cancel active drawing

## Storage Keys

### Drawings
- Key: `chart-drawings-{symbol}`
- Format: JSON array of Drawing objects

### Indicators
- Key: `chart-indicators-{symbol}`
- Format: JSON array of IndicatorConfig objects

## Common Patterns

### Add Multiple Indicators
```typescript
const indicators = [
  { id: 'ma-20', name: 'MA(20)', type: 'ma', parameters: { period: 20 } },
  { id: 'ma-50', name: 'MA(50)', type: 'ma', parameters: { period: 50 } },
  { id: 'rsi-14', name: 'RSI(14)', type: 'rsi', parameters: { period: 14 } }
];

indicators.forEach(ind => {
  indicatorManager.addIndicator({ ...ind, visible: true });
});
```

### Toggle All Indicators
```typescript
const indicators = indicatorManager.getIndicators();
indicators.forEach(ind => {
  indicatorManager.toggleIndicator(ind.id);
});
```

### Clear Everything
```typescript
drawingManager.clearAllDrawings();
indicatorManager.clearAll();
```

### Save All State
```typescript
const symbol = 'EURUSD';
drawingManager.saveToStorage(symbol);
indicatorManager.saveToStorage(symbol);
```

### Load All State
```typescript
const symbol = 'EURUSD';
drawingManager.loadFromStorage(symbol);
indicatorManager.loadFromStorage(symbol);
```

## Troubleshooting

### Crosshair Not Toggling
Check if chartManager has chart reference:
```typescript
chartManager.setChart(chartRef.current);
```

### Zoom Not Working
Verify barSpacing is being updated (NOT CSS transform):
```typescript
console.log(chartManager.getBarSpacing());
```

### Drawings Not Appearing
Check if series reference is set:
```typescript
drawingManager.setChart(chartRef.current, seriesRef.current);
```

### Indicators Not Calculating
Ensure OHLC data is set:
```typescript
indicatorManager.setOHLCData(candlesData);
```

### Command Bus Not Working
Verify dynamic import:
```typescript
try {
  const { commandBus } = await import('../services/commandBus');
} catch (error) {
  console.error('Command bus not available');
}
```

## Performance Tips

1. **Drawings**: Limit to 50-100 per chart
2. **Indicators**: Limit to 5-10 active at once
3. **Updates**: Debounce drawing position updates
4. **Storage**: Compress large datasets before saving
5. **Calculations**: Use Web Workers for heavy computation (future)

## Files Reference

**Services:**
- `src/services/chartManager.ts`
- `src/services/drawingManager.ts`
- `src/services/indicatorManager.ts`

**Components:**
- `src/components/TradingChart.tsx` (modified)
- `src/components/ChartWithIndicators.tsx`
- `src/components/IndicatorNavigator.tsx`

**Examples:**
- `src/examples/ChartIntegrationDemo.tsx`

**Documentation:**
- `docs/CHART_INTEGRATION_IMPLEMENTATION.md`
- `docs/CHART_INTEGRATION_QUICK_REFERENCE.md`
