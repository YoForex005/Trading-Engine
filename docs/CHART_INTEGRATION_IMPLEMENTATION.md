# Chart Integration Implementation - Agent 3

## Overview
Implementation of crosshair toggle, zoom controls (candle density), drawing tools, and indicator system for the TradingChart component.

## Implemented Features

### 1. Chart Manager (`chartManager.ts`)
Manages chart configuration and responds to toolbar commands.

**Features:**
- Crosshair toggle (on/off)
- Zoom in/out (adjusts bar spacing, NOT CSS transform)
- Set zoom level directly
- Reset zoom to default
- Fit content to viewport
- Scroll to realtime

**Key Methods:**
```typescript
chartManager.toggleCrosshair()    // Toggle crosshair visibility
chartManager.zoomIn()              // Increase bar spacing (wider candles)
chartManager.zoomOut()             // Decrease bar spacing (narrower candles)
chartManager.setZoom(barSpacing)   // Set specific bar spacing (1-50px)
chartManager.resetZoom()           // Reset to default (6px)
chartManager.fitContent()          // Fit all candles to viewport
chartManager.scrollToRealtime()    // Scroll to latest candle
```

### 2. Drawing Manager (`drawingManager.ts`)
Manages chart drawings (trendlines, horizontal/vertical lines, text annotations).

**Supported Drawing Types:**
- `trendline` - 2-point line (click twice)
- `hline` - Horizontal line (1 click)
- `vline` - Vertical line (1 click)
- `text` - Text annotation (1 click)

**Key Methods:**
```typescript
drawingManager.startDrawing(type, color)  // Start new drawing
drawingManager.addPoint(time, price)      // Add point to active drawing
drawingManager.finishDrawing()            // Complete and save drawing
drawingManager.deleteDrawing(id)          // Remove a drawing
drawingManager.clearAllDrawings()         // Clear all drawings
drawingManager.saveToStorage(symbol)      // Save to localStorage
drawingManager.loadFromStorage(symbol)    // Load from localStorage
```

**Events:**
- `drawing:saved` - Fired when drawing is completed
- `drawing:deleted` - Fired when drawing is removed
- `drawings:cleared` - Fired when all drawings are cleared

### 3. Indicator Manager (`indicatorManager.ts`)
Calculates and manages technical indicators.

**Implemented Indicators:**
- Simple Moving Average (SMA)
- Exponential Moving Average (EMA)
- Relative Strength Index (RSI)
- MACD (partial - line only)

**Planned Indicators:**
- Bollinger Bands
- Stochastic Oscillator
- CCI (Commodity Channel Index)
- Volume indicators
- Bill Williams indicators

**Key Methods:**
```typescript
indicatorManager.addIndicator(config)       // Add indicator to chart
indicatorManager.removeIndicator(id)        // Remove indicator
indicatorManager.toggleIndicator(id)        // Toggle visibility
indicatorManager.updateIndicator(id, params) // Update parameters
indicatorManager.getIndicators()            // Get all indicators
indicatorManager.clearAll()                 // Remove all indicators
indicatorManager.saveToStorage(symbol)      // Save to localStorage
indicatorManager.loadFromStorage(symbol)    // Load from localStorage
```

**Indicator Configuration:**
```typescript
interface IndicatorConfig {
  id: string;
  name: string;
  type: string;
  parameters: Record<string, any>;
  visible: boolean;
  color?: string;
}
```

### 4. Indicator Navigator (`IndicatorNavigator.tsx`)
Dialog component for browsing and adding indicators.

**Features:**
- Categorized indicator list (Trend, Oscillators, Volume, Bill Williams)
- Search functionality
- Category filtering
- Double-click or button to add
- Shows indicator parameters

**Categories:**
1. **Trend**: MA, EMA, Bollinger Bands, Ichimoku, Parabolic SAR
2. **Oscillators**: RSI, MACD, Stochastic, CCI, Momentum
3. **Volume**: Volume, OBV, VWAP, MFI
4. **Bill Williams**: Alligator, Awesome Oscillator, Fractals, Gator

### 5. TradingChart Integration
Modified `TradingChart.tsx` to integrate all features.

**Command Bus Subscriptions:**
```typescript
'TOGGLE_CROSSHAIR' → chartManager.toggleCrosshair()
'ZOOM_IN'          → chartManager.zoomIn()
'ZOOM_OUT'         → chartManager.zoomOut()
'FIT_CONTENT'      → chartManager.fitContent()
'SELECT_TRENDLINE' → drawingManager.startDrawing('trendline')
'SELECT_HLINE'     → drawingManager.startDrawing('hline')
'SELECT_VLINE'     → drawingManager.startDrawing('vline')
'SELECT_TEXT'      → drawingManager.startDrawing('text')
'OPEN_INDICATORS'  → Open indicator navigator dialog
```

**Dynamic Command Bus Loading:**
The chart uses dynamic imports to handle the command bus, allowing graceful degradation if Agent 1 hasn't completed yet:

```typescript
const { commandBus } = await import('../services/commandBus');
```

### 6. ChartWithIndicators Wrapper
High-level component that combines TradingChart and IndicatorNavigator.

**Usage:**
```tsx
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

## Technical Implementation Details

### Zoom Implementation (NOT CSS Scale)
**CRITICAL**: Zoom is implemented via `barSpacing` in lightweight-charts, NOT CSS transform.

```typescript
// CORRECT - Adjusts candle density
chartRef.current.applyOptions({
  timeScale: {
    barSpacing: newValue  // 1-50px per candle
  }
});

// WRONG - Would scale entire chart
chartRef.current.style.transform = 'scale(...)';
```

**Bar Spacing Values:**
- Minimum: 1px (most dense)
- Default: 6px (balanced)
- Maximum: 50px (widest candles)

### Crosshair Toggle
Toggles between `CrosshairMode.Normal` and `CrosshairMode.Hidden`:

```typescript
chartRef.current.applyOptions({
  crosshair: {
    mode: enabled ? CrosshairMode.Normal : CrosshairMode.Hidden
  }
});
```

### Drawing Persistence
Drawings are saved to localStorage per symbol:

- Key format: `chart-drawings-{symbol}`
- Automatically saved on unmount
- Automatically loaded on mount
- Survives chart type changes

### Indicator Calculations
All calculations use the OHLC data array:

```typescript
// Example: SMA calculation
for (let i = period - 1; i < data.length; i++) {
  let sum = 0;
  for (let j = 0; j < period; j++) {
    sum += data[i - j].close;
  }
  result.push({
    time: data[i].time,
    value: sum / period
  });
}
```

## Testing Guide

### Manual Testing Checklist

#### Crosshair Toggle
1. Click crosshair button
2. Verify crosshair disappears
3. Click again to re-enable
4. Verify crosshair reappears

#### Zoom Controls
1. Click + button
2. Verify candles get wider (NOT CSS scale)
3. Click - button
4. Verify candles get narrower
5. Verify minimum (1px) and maximum (50px) limits

#### Drawing Tools - Trendline
1. Click trendline button
2. Click chart at point 1
3. Click chart at point 2
4. Verify line drawn between points
5. Switch chart type (candlestick → bar)
6. Verify line persists

#### Drawing Tools - H-Line
1. Click H-line button
2. Click chart once
3. Verify horizontal line drawn at price level
4. Zoom in/out
5. Verify line remains at correct price

#### Drawing Tools - V-Line
1. Click V-line button
2. Click chart once
3. Verify vertical line drawn at time

#### Drawing Tools - Text
1. Click text button
2. Click chart
3. Verify text annotation appears

#### Indicator Navigator
1. Click "Indicators" button
2. Verify dialog opens
3. Search for "RSI"
4. Double-click RSI
5. Verify RSI appears on chart (separate scale)
6. Verify RSI in active indicators list

#### Indicator - Moving Average
1. Open indicator navigator
2. Add "Moving Average"
3. Verify MA line overlays price chart
4. Switch timeframe (1m → 5m)
5. Verify MA recalculates

#### Persistence
1. Add drawing and indicator
2. Switch to different symbol
3. Switch back to original symbol
4. Verify drawing and indicator are restored

## Integration with Agent 1 (Command Bus)

The chart integration is designed to work with the command bus created by Agent 1.

**Dynamic Loading Pattern:**
```typescript
try {
  const { commandBus } = await import('../services/commandBus');
  // Subscribe to commands
} catch (error) {
  console.log('Command bus not yet available');
}
```

**Expected Commands from Toolbar (Agent 2):**
- `TOGGLE_CROSSHAIR`
- `ZOOM_IN`
- `ZOOM_OUT`
- `FIT_CONTENT`
- `SELECT_TRENDLINE`
- `SELECT_HLINE`
- `SELECT_VLINE`
- `SELECT_TEXT`
- `OPEN_INDICATORS`

## Files Created

### Services
- `clients/desktop/src/services/chartManager.ts` (139 lines)
- `clients/desktop/src/services/drawingManager.ts` (339 lines)
- `clients/desktop/src/services/indicatorManager.ts` (366 lines)

### Components
- `clients/desktop/src/components/IndicatorNavigator.tsx` (213 lines)
- `clients/desktop/src/components/ChartWithIndicators.tsx` (62 lines)

### Examples
- `clients/desktop/src/examples/ChartIntegrationDemo.tsx` (283 lines)

### Modified Files
- `clients/desktop/src/components/TradingChart.tsx` (added 110 lines of integration code)
- `clients/desktop/src/services/index.ts` (added exports)
- `clients/desktop/src/components/index.ts` (added exports)

## Dependencies

**Required:**
- `lightweight-charts` - Already installed
- Command bus from Agent 1 - Dynamically imported
- Toolbar state from Agent 2 - Via command bus

**No New Dependencies Required** ✓

## Performance Considerations

1. **Drawing Updates**: Throttled to visual updates only
2. **Indicator Calculations**: Only recalculate when OHLC data changes
3. **Event Listeners**: Properly cleaned up on unmount
4. **LocalStorage**: Async operations for persistence
5. **Command Bus**: Dynamic imports prevent blocking

## Future Enhancements

1. **Drawing Tools**
   - Rectangle/ellipse shapes
   - Fibonacci retracements
   - Gann fans
   - Edit existing drawings
   - Drawing templates

2. **Indicators**
   - Complete MACD (signal + histogram)
   - Bollinger Bands
   - Stochastic
   - Custom indicator builder
   - Indicator alerts

3. **Performance**
   - WebWorker for indicator calculations
   - Virtualized drawing rendering
   - Indicator result caching

4. **UI/UX**
   - Drawing color picker
   - Indicator parameter editor
   - Indicator templates
   - Drawing layers

## Known Limitations

1. **V-Line Implementation**: Requires timeScale coordinate conversion (more complex)
2. **MACD**: Only line series implemented (signal and histogram pending)
3. **Drawing Rendering**: Currently uses HTML overlays (could be optimized with chart primitives)
4. **Indicator Persistence**: Uses localStorage (could be database-backed)

## Acceptance Criteria - Status

✅ Crosshair toggles on/off (not always on)
✅ Zoom changes candle density (bar spacing), NOT CSS transform
✅ Trendline can be drawn (2 clicks)
✅ H-line can be drawn (1 click)
✅ V-line can be drawn (1 click) - Basic implementation
✅ Drawings persist across chart type changes
✅ Indicator navigator dialog opens
✅ At least 1 indicator (MA, EMA, RSI) can be added to chart
✅ All components integrate via command bus
✅ Dynamic loading handles missing dependencies

## Demo Usage

See `ChartIntegrationDemo.tsx` for a complete working example with toolbar.

```tsx
import { ChartIntegrationDemo } from './examples/ChartIntegrationDemo';

function App() {
  return <ChartIntegrationDemo />;
}
```

## Conclusion

All deliverables have been implemented and are ready for integration with Agent 1 (command bus) and Agent 2 (toolbar). The implementation uses dynamic imports to gracefully handle dependencies and provides a solid foundation for advanced chart analysis features.
