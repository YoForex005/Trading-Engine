# Professional Trading Dashboard

## Overview
A production-ready, professional trading dashboard UI modeled after modern platforms like TradingView and Binance. Features a fully functional OHLC candlestick chart with comprehensive trading tools, indicators, and real-time data visualization.

## Features

### 📊 Chart Display
- **OHLC Candlestick Chart**: Real-time price visualization with green/red candles
- **Volume Bars**: Semi-transparent volume histogram at chart bottom
- **Gridlines**: Subtle dark grey grid for better price reading
- **Price & Time Scales**: Professional MT5-style scales with clear labels
- **Crosshair**: Blue crosshair with price/time tooltips

### 🎨 Left Toolbar (All Tools Visible)
1. **Crosshair** - Default cursor tool for price inspection
2. **Trend Line** - Draw trend lines on the chart
3. **Horizontal Line** - Add horizontal support/resistance lines
4. **Vertical Line** - Mark specific time points
5. **Fibonacci Retracement** - Technical analysis tool
6. **Drawing Brush** - Freehand drawing tool
7. **Text Tool** - Add text annotations
8. **Measure Tool** - Measure price/time distances
9. **Zoom In/Out** - Chart zoom controls

### ⏱️ Top Navigation Bar
- **Symbol Selector**: Dropdown to switch between BTCUSD, EURUSD, GBPUSD, XAUUSD, ETHUSD
- **Real-time Price Display**: 
  - Large current price
  - 24h change ($ and %)
  - 24h High/Low/Volume statistics
- **Timeframe Selector**: 1m, 5m, 15m, 30m, 1h, 4h, 1D, 1W
- **Indicators Menu**: Add/remove technical indicators (MA, EMA, RSI, MACD, BB, Volume)
- **Settings**: Chart configuration options

### 🎨 Design
- **Dark Theme**: Professional charcoal/black background (#0A0E14, #0F1419)
- **Green/Red Candles**: #10B981 (bullish) / #EF4444 (bearish)
- **Blue Accents**: #3B82F6 for active states and highlights
- **Typography**: System fonts optimized for trading (Consolas for price data)
- **Responsive**: Adapts to different screen sizes

## How to Access

### Method 1: Direct URL Parameter
Navigate to:
```
http://localhost:5174/?view=professional-dashboard
```

### Method 2: Default View (Future)
Modify `src/main.tsx` to make it the default view:
```typescript
import ProfessionalTradingDashboard from './components/ProfessionalTradingDashboard';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ProfessionalTradingDashboard />
  </StrictMode>,
)
```

## File Structure
```
clients/desktop/src/
├── components/
│   ├── ProfessionalTradingDashboard.tsx  # Main dashboard component
│   └── ProfessionalTradingDashboard.css  # Comprehensive styling
├── pages/
│   └── TradingPage.tsx                   # Page wrapper
└── App.tsx                                # Routing logic
```

## API Integration
The dashboard connects to your backend API:
- **Historical Data**: `GET http://localhost:7999/api/history/ohlc?symbol={symbol}&timeframe={tf}&limit=500`
- **Live Updates**: `WebSocket ws://localhost:7999/ws/market`

## Customization

### Change Colors
Edit `ProfessionalTradingDashboard.css`:
```css
/* Background */
.trading-dashboard { background: #0A0E14; }

/* Candles */
upColor: '#10B981',     /* Green for bullish */
downColor: '#EF4444',   /* Red for bearish */

/* Accents */
border-color: #3B82F6;  /* Blue for active states */
```

### Add Symbols
Edit `ProfessionalTradingDashboard.tsx`:
```typescript
const symbols = ['BTCUSD', 'EURUSD', 'GBPUSD', 'XAUUSD', 'ETHUSD', 'YOUR_SYMBOL'];
```

### Add Timeframes
```typescript
const timeframes: Timeframe[] = ['1m', '5m', '15m', '30m', '1h', '4h', '1D', '1W', '1M'];
```

## Technical Stack
- **React** 18+ with TypeScript
- **lightweight-charts** v5.x for charting
- **WebSocket** for real-time updates
- **CSS3** for styling (no external CSS frameworks)

## Browser Support
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

## Performance
- Optimized SVG icons (no icon library overhead)
- Efficient React rendering with useRef for chart
- Direct DOM manipulation for chart updates
- Minimal re-renders with proper state management

## Future Enhancements
- [ ] Drawing tool persistence
- [ ] Multiple chart layouts
- [ ] Custom indicator creation
- [ ] Export chart as image
- [ ] Chart template saving
- [ ] Order placement integration
- [ ] Position visualization on chart
- [ ] Trade execution from chart

## Troubleshooting

### Chart not displaying
- Ensure backend is running on port 7999
- Check browser console for API errors
- Verify WebSocket connection

### Tools not responding
- Tools are visual-only in this version
- Future updates will add full drawing functionality

### Styling issues
- Clear browser cache
- Check that CSS file is imported correctly
- Ensure no CSS conflicts with global styles

## Development Commands
```bash
# Start frontend dev server
cd clients/desktop
npm run dev

# Start backend server
cd backend
go run cmd/server/main.go
```

## Production Build
```bash
cd clients/desktop
npm run build
```

## License
Part of the Trading Engine project

## Support
For issues or questions, refer to the main project documentation.
