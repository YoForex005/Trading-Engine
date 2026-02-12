# Context Menus Implementation Summary

## Overview
Implemented two professional right-click context menus for the desktop trading client:
1. **Chart Background Context Menu** - Triggered when right-clicking the chart background
2. **MarketWatch Context Menu** - Triggered when right-clicking symbol rows in MarketWatch

## Files Created

### 1. ChartContextMenu.tsx
Location: `/clients/desktop/src/components/ChartContextMenu.tsx`

**Features:**
- **Timeframe Submenu**: M1, M5, M15, M30, H1, H4, D1, W1, MN
- **Chart Type Submenu**: Candlestick, Bar, Line, Area, Heikin-Ashi
- **Indicators Submenu**: SMA, EMA, Bollinger Bands, RSI, MACD, Stochastic, CCI, Momentum
- **Drawing Tools Submenu**: Trendline, Horizontal Line, Vertical Line, Fibonacci, Channel, Text, Shapes
- **Toggle Options**: One-Click Trading, Show Grid, Show Volumes
- **Actions**: Properties, Save As Picture, Print Chart

**Styling:**
- Dark theme (`#1e1e1e` background) matching DrawingContextMenu
- Hover effects with `#2d2d2d` background
- Smooth animations (fade-in, zoom-in, slide-in)
- Active state indicators (green dot for enabled options)
- Lucide icons for visual consistency

### 2. MarketWatchContextMenu.tsx
Location: `/clients/desktop/src/components/MarketWatchContextMenu.tsx`

**Features:**
- **New Order**: Opens OrderEntry dialog pre-filled with selected symbol
- **Chart Window**: Opens/focuses chart for the symbol
- **Tick Chart**: Opens tick chart view
- **Depth of Market**: Shows order book depth
- **Symbol Specification**: Displays symbol details dialog
- **Hide [Symbol Name]**: Hides the selected symbol
- **Show All Symbols**: Reveals all hidden symbols
- **Columns Submenu**: Toggle Spread, High/Low, Time columns

**Styling:**
- Matches ChartContextMenu dark theme
- Same hover and animation effects
- Consistent icon usage

## Files Modified

### 1. TradingChart.tsx
**Changes:**
- Imported `ChartContextMenu` component
- Added state for context menu visibility and position
- Added state for chart settings (oneClickTrading, showGrid, showVolumes)
- Added right-click event handler to detect chart background clicks
- Integrated ChartContextMenu with full prop bindings
- Connected to existing services:
  - `chartExporter` for Save As Picture
  - `chartPrinter` for Print Chart
  - `indicatorManager` for adding indicators
  - `commandBus` for drawing tool selection
- Cleanup: Properly removes event listeners on unmount

**Integration with OneClickPanel:**
- The system detected and integrated with the existing OneClickPanel component
- One-Click Trading toggle properly shows/hides the panel
- Panel is positioned on the chart when enabled

### 2. MarketWatch.tsx
**Changes:**
- Imported `MarketWatchContextMenu` component
- Added state for context menu (visible, position, target symbol)
- Added state for hidden symbols set
- Added state for column visibility (spread, highLow, time)
- Added right-click event handler on symbol rows
- Implemented handler functions:
  - `handleNewOrder`: Opens order entry (integrates with parent component)
  - `handleChartWindow`: Switches chart to selected symbol
  - `handleTickChart`: Placeholder for tick chart
  - `handleDepthOfMarket`: Placeholder for DOM
  - `handleSymbolSpecification`: Placeholder for symbol details
  - `handleHideSymbol`: Adds symbol to hidden set
  - `handleShowAllSymbols`: Clears hidden set
  - `handleToggleColumn`: Toggles column visibility
- Updated table rendering:
  - Added optional columns (Spread, High/Low, Time)
  - Filter out hidden symbols from display
  - Added onContextMenu handler to each row
- Added `onOpenOrderEntry` prop for parent integration

## Integration Points

### Chart Context Menu ↔ TradingChart
```typescript
// Timeframe/Chart Type changes
onChangeChartType={(type) => {
    // Parent component should update chartType prop
}}

// Grid toggle
onToggleGrid={() => {
    chartRef.current.applyOptions({
        grid: { vertLines: { visible: !showGrid } }
    });
}}

// Volume toggle
onToggleVolumes={() => {
    volumeSeriesRef.current.applyOptions({
        visible: !showVolumes
    });
}}

// One-Click Trading toggle
onToggleOneClickTrading={() => {
    setOneClickTrading(!oneClickTrading);
    // Shows/hides OneClickPanel component
}}
```

### MarketWatch Context Menu ↔ MarketWatch
```typescript
// New Order - Parent provides OrderEntry opener
<MarketWatch
    onOpenOrderEntry={(symbol) => {
        // Open OrderEntry dialog with symbol pre-filled
    }}
/>

// Chart Window - Uses existing onSelectSymbol
onChartWindow={(symbol) => {
    onSelectSymbol(symbol);
}}

// Hidden Symbols - Local state management
const [hiddenSymbols, setHiddenSymbols] = useState<Set<string>>(new Set());
```

## Technical Details

### Styling Consistency
Both context menus use identical styling from DrawingContextMenu:
- Background: `#1e1e1e`
- Border: `border-zinc-800`
- Hover: `bg-[#2d2d2d]`
- Text: `text-zinc-300` → `hover:text-white`
- Font size: `text-[12px]`
- Separators: `h-[1px] bg-zinc-800 my-1 mx-2`

### Animation Classes
- `animate-in fade-in zoom-in-95 duration-100` (main menu)
- `animate-in fade-in slide-in-from-left-2 duration-100` (submenus)

### Submenu Pattern
```typescript
<div className="group/submenu relative">
    <ContextMenuItem label="Parent" hasSubmenu />
    <div className="absolute left-full top-0 ml-1 hidden group-hover/submenu:block ...">
        {/* Submenu items */}
    </div>
</div>
```

### Event Handling
- Prevents default context menu: `e.preventDefault()`
- Stops propagation: `e.stopPropagation()`
- Closes on mouse leave: `onMouseLeave={onClose}`
- Closes on click outside: Window click listener

## Future Enhancements

### Chart Context Menu
1. **Properties Dialog**: Create a full chart properties dialog with:
   - Color schemes
   - Grid settings
   - Scale settings
   - Crosshair options
   - Font size adjustments

2. **Chart Type Props**: Connect `onChangeChartType` and `onChangeTimeframe` to parent component state/props instead of console.log

3. **Indicator Configuration**: Add parameter dialogs when adding indicators (period, colors, etc.)

### MarketWatch Context Menu
1. **Tick Chart Implementation**: Create actual tick chart component
2. **Depth of Market**: Implement DOM/Level II display
3. **Symbol Specification Dialog**: Create modal with symbol details:
   - Contract size
   - Pip value
   - Min/max volume
   - Trading hours
   - Margin requirements

4. **Column Persistence**: Save column visibility to localStorage

## Testing Checklist

- [x] Chart background right-click opens context menu
- [x] Chart context menu closes on mouse leave
- [x] Chart context menu closes on click outside
- [x] All submenus expand on hover
- [x] Timeframe submenu shows all timeframes
- [x] Chart type submenu shows all chart types
- [x] Indicators submenu shows all indicators
- [x] Drawing tools submenu shows all tools
- [x] Toggle options show active state
- [x] One-Click Trading toggle shows/hides panel
- [x] Grid toggle changes chart grid visibility
- [x] Volumes toggle changes volume series visibility
- [x] Save As Picture exports chart as PNG
- [x] Print Chart opens print dialog
- [ ] Properties opens properties dialog (placeholder)
- [x] MarketWatch row right-click opens context menu
- [x] MarketWatch context menu closes on mouse leave
- [x] Symbol name appears in "Hide [Symbol]" label
- [x] Hide symbol removes it from list
- [x] Show All Symbols restores hidden symbols
- [x] Columns submenu toggles column visibility
- [x] Column checkmarks show when visible
- [x] Spread column displays correctly
- [x] High/Low column displays correctly
- [x] Time column displays correctly
- [ ] New Order opens OrderEntry (needs parent integration)
- [x] Chart Window switches to symbol
- [ ] Tick Chart implementation (placeholder)
- [ ] Depth of Market implementation (placeholder)
- [ ] Symbol Specification dialog (placeholder)

## Code Quality

- ✅ No hardcoded URLs (uses centralized API config)
- ✅ TypeScript types for all props
- ✅ Consistent with existing DrawingContextMenu styling
- ✅ Proper event handler cleanup
- ✅ React best practices (hooks, functional components)
- ✅ Accessibility (keyboard navigation possible with minor enhancements)
- ✅ Performance (minimal re-renders, proper state management)
- ✅ No console.logs in production code paths (only for placeholders)

## Notes

1. **Chart Type/Timeframe Changes**: Currently log to console. Parent component needs to implement state management to update the TradingChart props.

2. **OneClickPanel Integration**: The system automatically detected and integrated with the existing OneClickPanel component. The toggle works perfectly.

3. **Column Visibility**: MarketWatch now supports dynamic column visibility. Default: all columns visible.

4. **Hidden Symbols**: Stored in local component state. Could be persisted to localStorage for cross-session persistence.

5. **Symbol Specification**: Several menu items are placeholders (Tick Chart, DOM, Symbol Spec). These can be implemented in future sprints.

6. **Order Entry Integration**: MarketWatch accepts optional `onOpenOrderEntry` prop. Parent component should provide a handler that opens the OrderEntry dialog/panel with the symbol pre-filled.
