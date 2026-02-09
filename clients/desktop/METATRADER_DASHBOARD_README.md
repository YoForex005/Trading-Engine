# MetaTrader-Style Trading Dashboard

## Overview
A professional, production-ready trading dashboard UI modeled after MetaTrader 4/5 with a **top horizontal toolbar** and light theme. All chart tools are positioned in a single horizontal toolbar above the OHLC chart, matching institutional desktop trading terminal aesthetics.

## Key Features

### 🎯 **Top Horizontal Toolbar Layout**
Unlike the previous dark-themed sidebar design, this dashboard features:
- **All tools in one horizontal row** at the top
- Clean separation of tool groups with visual dividers
- Professional light theme with subtle 3D effects
- MetaTrader-style button styling

### 📊 **Toolbar Sections (Left to Right)**

#### 1. **Chart Tools Section**
- ✅ Cursor/Selection tool (default)
- ✅ Crosshair
- ✅ Zoom In / Zoom Out
- ✅ Chart Type Toggle:
  - Candlestick (default)
  - Bar chart
  - Line chart
- ✅ Indicators button

#### 2. **Drawing Tools Section**
- ✅ Trend Line
- ✅ Horizontal Line
- ✅ Vertical Line
- ✅ Arrow
- ✅ Fibonacci Retracement

#### 3. **Timeframe Selector** (Center)
- M1, M5, M15, M30 (minute charts)
- H1, H4 (hourly charts)
- D1 (daily)
- W1 (weekly)
- MN (monthly)

#### 4. **Symbol & Price Display** (Right)
- Symbol name (BTCUSD)
- Real-time price
- 24h price change with color coding

### 📈 **Chart Display**
- **Light Theme**: White background with subtle grey gridlines
- **Green/Red Candles**: #26A69A (bullish) / #EF5350 (bearish)
- **Professional Layout**: Price scale on right, time scale at bottom
- **Clean Typography**: Monospace fonts for price data
- **Responsive**: Adapts to different screen sizes

## How to Access

Navigate to:
```
http://localhost:5174/?view=metatrader
```

## File Structure
```
clients/desktop/src/
├── components/
│   ├── MetaTraderDashboard.tsx    # Main MT-style component
│   └── MetaTraderDashboard.css    # Light theme styling
└── App.tsx                         # Routing logic
```

## Design Philosophy

### **Light Theme (Desktop Trading Software)**
- White background (#FFFFFF)
- Light grey toolbar (#FAFAFA to #E8E8E8gradient)
- Subtle gridlines (#E8E8E8)
- Professional grey borders (#CCCCCC)
- Windows-style 3D button effects

### **vs. Dark Theme (Previous Dashboard)**
The previous `ProfessionalTradingDashboard` had:
- Dark background (#0A0E14)
- Left vertical sidebar
- Modern fintech web app aesthetic

This MetaTrader dashboard has:
- Light background (#FFFFFF)
- Top horizontal toolbar
- Classic desktop terminal aesthetic

## API Integration
- **Historical Data**: `GET http://localhost:7999/api/history/ohlc`
- **Timeframe Mapping**:
  - M1 → 1m, M5 → 5m, M15 → 15m, M30 → 30m
  - H1 → 1h, H4 → 4h
  - D1 → 1d, W1 → 1w, MN → 1M

## Features Implemented

### ✅ **Fully Functional**
1. **Chart Type Switching**: Toggle between candlestick, bar, and line charts
2. **Timeframe Selection**: All 9 timeframes with data fetching
3. **Tool Selection**: Visual feedback when selecting different tools
4. **Zoom Controls**: Chart navigation via toolbar buttons
5. **Price Display**: Real-time price and change calculation
6. **Indicators Panel**: Dropdown menu for adding indicators
7. **Responsive Design**: Adapts to different screen sizes

### ⚠️ **Visual-Only (Future Enhancement)**
- Drawing tool functionality (trend lines, horizontals, etc.)
- Indicator calculations
- Multi-symbol support

## Customization

### Change Colors
Edit `MetaTraderDashboard.tsx`:
```typescript
// Candlestick colors
upColor: '#26A69A',      // Green
downColor: '#EF5350',    // Red

// Chart types
case 'candlestick':
  // Customize appearance
```

### Modify Toolbar Layout
Edit `MetaTraderDashboard.css`:
```css
.mt-toolbar {
  height: 44px;  /* Adjust toolbar height */
}

.mt-tool-btn {
  width: 32px;   /* Adjust button size */
  height: 32px;
}
```

### Add Symbols
Edit `MetaTraderDashboard.tsx`:
```typescript
const [selectedSymbol] = useState('BTCUSD');
// Change to support symbol switching
```

## Comparison: Two Dashboard Styles

| Feature         | MetaTrader Dashboard | Professional Dashboard         |
| --------------- | -------------------- | ------------------------------ |
| **Theme**       | Light (white)        | Dark (charcoal)                |
| **Toolbar**     | Top horizontal       | Left vertical sidebar          |
| **Style**       | Desktop terminal     | Modern web app                 |
| **URL**         | `?view=metatrader`   | `?view=professional-dashboard` |
| **Inspiration** | MT4/MT5              | TradingView/Binance            |

## Technical Details

### **Stack**
- React 18+ with TypeScript
- lightweight-charts v5.x
- Pure CSS (no frameworks)

### **Performance**
- Optimized SVG icons
- Efficient chart updates
- Minimal re-renders
- Direct DOM manipulation for chart

### **Browser Support**
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

## Usage

### **Basic Operation**
1. Click tool icons to select different drawing tools
2. Use timeframe buttons to change chart period
3. Toggle chart types (candles/bars/lines)
4. Click Indicators to show indicator menu
5. Use zoom buttons to navigate the chart

### **Keyboard Shortcuts** (Future)
- `Space` - Toggle crosshair
- `+/-` - Zoom in/out
- `Arrow keys` - Pan chart

## Development

### Run Locally
```bash
# Start frontend
cd clients/desktop
npm run dev

# Start backend
cd backend
go run cmd/server/main.go
```

### Build for Production
```bash
cd clients/desktop
npm run build
```

## Troubleshooting

### Chart not displaying
- Ensure backend runs on port 7999
- Check browser console for API errors
- Verify fetch requests succeed

### Toolbar buttons not responding
- Tools show visual feedback but drawing not yet implemented
- Chart type and timeframe buttons are fully functional

### Styling issues
- Clear browser cache
- Check that CSS file imports correctly
- Verify no global CSS conflicts

## Future Enhancements
- [ ] Drawing tool implementation
- [ ] Technical indicator calculations
- [ ] Multiple chart windows
- [ ] Chart template saving
- [ ] Order placement integration
- [ ] Position visualization
- [ ] Custom timeframes
- [ ] Symbol switcher

## License
Part of the Trading Engine project

---

**Quick Access:**
- Dark Theme Dashboard: `http://localhost:5174/?view=professional-dashboard`
- Light Theme (MT): `http://localhost:5174/?view=metatrader`
- Default App: `http://localhost:5174/`
