# Print Functionality Documentation

## Overview

The trading platform includes comprehensive print functionality that allows users to print charts with full customization options including page setup, headers, footers, indicators, and drawings.

## Features

### 1. Print Menu Options

Located in `File Menu`:
- **Print (Ctrl+P)**: Quick print with current settings
- **Print Preview**: Preview before printing with zoom controls
- **Print Setup**: Configure print preferences

### 2. Print Setup Dialog

Accessible via `File > Print Setup`

#### Page Setup
- **Orientation**: Portrait or Landscape
- **Paper Size**: A4, Letter, Legal, A3
- **Margins**: Top, Right, Bottom, Left (0-50mm)

#### Print Options
- **Include Header**: Show chart title and date range
- **Include Footer**: Show custom footer text and page numbers
- **Include Grid Lines**: Display chart grid in print
- **Include Indicators Legend**: Show list of active indicators
- **Include Drawings Legend**: Show list of drawings/annotations
- **Scale to Fit Page**: Automatically scale chart to fit page
- **Color Mode**: Color or Black & White

#### Custom Text
- **Header Text**: Custom header (default: Symbol - Timeframe)
- **Footer Text**: Custom footer text

#### Save Preferences
- **Save as Default**: Store preferences for future prints

### 3. Print Preview Dialog

Accessible via `File > Print Preview`

#### Controls
- **Zoom In/Out**: 50% to 200% zoom levels
- **Zoom Reset**: Click percentage to reset to 100%
- **Print Button**: Execute print after preview
- **Cancel**: Close preview without printing

#### Preview Display
- Real-time preview of print output
- Shows all selected elements (header, footer, legends)
- Accurate page dimensions based on settings
- Visual representation of margins

### 4. Chart Capture

The system automatically captures:
- Current chart canvas with all candlesticks
- Applied indicators (when legend enabled)
- Drawing tools and annotations (when legend enabled)
- Current symbol and timeframe
- Date range displayed on chart

## Technical Implementation

### Components

#### 1. PrintSetupDialog
**Location**: `src/components/dialogs/PrintSetupDialog.tsx`

**Props**:
```typescript
interface PrintSetupDialogProps {
  onConfirm: (preferences: PrintPreferences) => void;
  onCancel: () => void;
  accountId?: string;
}
```

**Features**:
- Form validation for margins
- Preference persistence via API
- Checkbox controls for options
- Real-time preference updates

#### 2. PrintPreviewDialog
**Location**: `src/components/dialogs/PrintPreviewDialog.tsx`

**Props**:
```typescript
interface PrintPreviewDialogProps {
  onClose: () => void;
  onPrint: () => void;
  chartData?: ChartPrintData;
  preferences?: PrintPreferences;
}
```

**Features**:
- Zoom controls (50% - 200%)
- IFrame-based preview
- Print button with loading state
- Chart information display

#### 3. FileMenu Integration
**Location**: `src/components/layout/FileMenu.tsx`

**Integration**:
- Print button (Ctrl+P shortcut)
- Print Preview button
- Print Setup button
- Automatic chart capture
- Dialog state management

### Services

#### chartPrinter Service
**Location**: `src/services/chartPrinter.ts`

**Key Methods**:

```typescript
// Generate printable HTML
generatePrintableHTML(data: ChartPrintData, preferences: PrintPreferences): string

// Execute print
printChart(data: ChartPrintData, preferences?: PrintPreferences): Promise<void>

// Capture chart canvas
captureChartCanvas(chartApi: IChartApi): HTMLCanvasElement | null

// Preference management
loadPreferences(accountId: string): Promise<PrintPreferences>
savePreferences(accountId: string, preferences: PrintPreferences): Promise<void>
getPreferences(): PrintPreferences
```

**Print Data Structure**:

```typescript
interface ChartPrintData {
  symbol: string;
  timeframe: string;
  dateRange: {
    from: Date;
    to: Date;
  };
  chartCanvas?: HTMLCanvasElement;
  indicators: Array<{
    name: string;
    parameters: Record<string, any>;
  }>;
  drawings: Array<{
    type: string;
    label?: string;
  }>;
}
```

**Print Preferences**:

```typescript
interface PrintPreferences {
  pageSize: 'A4' | 'Letter' | 'Legal' | 'A3';
  orientation: 'portrait' | 'landscape';
  colorMode: 'color' | 'grayscale';
  includeHeader: boolean;
  includeFooter: boolean;
  includeGrid: boolean;
  includeIndicators: boolean;
  includeDrawings: boolean;
  marginTop: number;
  marginBottom: number;
  marginLeft: number;
  marginRight: number;
  scaleToFit: boolean;
  headerText?: string;
  footerText?: string;
}
```

### Hooks

#### useChartPrint Hook
**Location**: `src/hooks/useChartPrint.ts`

**Usage**:

```typescript
const {
  showSetupDialog,
  showPreviewDialog,
  printPreferences,
  chartData,
  openPrintSetup,
  handleSetupComplete,
  handleDirectPrint,
  closeDialogs,
} = useChartPrint({
  accountId: '1',
  symbol: 'BTCUSD',
  timeframe: '1m',
  chartApi: chartApiRef.current,
  getIndicators: () => indicatorManager.getAllIndicators(),
  getDrawings: () => drawingManager.getAllDrawings(),
});
```

### Styles

#### Print CSS
**Location**: `src/styles/print.css`

**Features**:
- `@media print` rules for browser printing
- Hide non-essential UI elements
- Full-page chart layout
- Color adjustment for accurate printing
- Page break controls
- Print-specific styling for legends and headers

**Key Rules**:
```css
@media print {
  /* Hide UI elements */
  .no-print, nav, .toolbar, .sidebar, button {
    display: none !important;
  }

  /* Full page chart */
  .chart-container {
    width: 100% !important;
    height: 100vh !important;
  }

  /* Accurate colors */
  * {
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }
}
```

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+P` | Quick Print (uses current preferences) |

**Implementation**: Registered in `GlobalShortcuts` component and `App.tsx`

## User Workflow

### Quick Print (Ctrl+P)
1. Press `Ctrl+P` or `File > Print`
2. Uses saved preferences or defaults
3. Chart is automatically captured
4. Browser print dialog opens
5. User completes print

### Print with Setup
1. Open `File > Print Setup`
2. Configure all print options
3. Optionally save as default
4. Click "OK"
5. Print Preview opens automatically
6. Review and adjust zoom
7. Click "Print" to execute

### Print with Preview Only
1. Open `File > Print Preview`
2. Review chart with zoom controls
3. Click "Print" to execute
4. Or "Cancel" to abort

## API Integration

### Save Preferences Endpoint
```
POST /api/accounts/:accountId/print-preferences
Content-Type: application/json

{
  "pageSize": "A4",
  "orientation": "landscape",
  "colorMode": "color",
  ...
}
```

### Load Preferences Endpoint
```
GET /api/accounts/:accountId/print-preferences

Response:
{
  "pageSize": "A4",
  "orientation": "landscape",
  ...
}
```

## Browser Compatibility

### Supported Browsers
- ✅ Chrome/Edge (Chromium) - Full support
- ✅ Firefox - Full support
- ✅ Safari - Full support (requires -webkit-print-color-adjust)
- ⚠️ Internet Explorer - Not supported

### Required Features
- Canvas API (for chart capture)
- CSS @media print rules
- IFrame printing
- print-color-adjust CSS property

## Testing

### Unit Tests
**Location**: `src/tests/print-functionality.test.tsx`

**Coverage**:
- Hook initialization and state management
- chartPrinter service methods
- HTML generation with various options
- Preference handling
- Dialog interactions

### Manual Testing

#### Test Print Setup
1. Open Print Setup
2. Change each setting
3. Verify preview updates
4. Save preferences
5. Reload and verify persistence

#### Test Print Preview
1. Open Preview
2. Test zoom controls (50%, 75%, 100%, 150%, 200%)
3. Verify chart visibility
4. Check legends and headers
5. Execute print

#### Test Direct Print
1. Press Ctrl+P
2. Verify chart capture
3. Check browser print dialog
4. Cancel and verify no errors

#### Test Chart Capture
1. Add indicators to chart
2. Draw annotations
3. Print and verify all elements visible
4. Test with different chart types
5. Test with different timeframes

## Troubleshooting

### Chart Not Printing
**Problem**: Blank page or no chart visible

**Solutions**:
1. Ensure chart is loaded before printing
2. Check browser console for canvas errors
3. Verify chart container has `.tv-lightweight-charts` class
4. Try Print Preview first to debug

### Colors Not Printing
**Problem**: Chart appears black and white

**Solutions**:
1. Check Color Mode setting (should be "Color")
2. Verify browser supports print-color-adjust
3. Check printer supports color printing
4. Update browser to latest version

### Print Quality Poor
**Problem**: Blurry or pixelated output

**Solutions**:
1. Enable "Scale to Fit Page" option
2. Use landscape orientation for charts
3. Increase browser zoom before printing
4. Select higher quality in printer settings

### Preferences Not Saving
**Problem**: Settings reset after reload

**Solutions**:
1. Verify accountId is provided
2. Check API endpoint is accessible
3. Check browser console for save errors
4. Verify backend endpoint is implemented

## Future Enhancements

### Planned Features
- [ ] Multi-page printing for long charts
- [ ] PDF export (requires html2pdf library)
- [ ] Print templates (predefined layouts)
- [ ] Batch printing (multiple charts)
- [ ] Print to file (save as PNG/PDF)
- [ ] Advanced chart annotations in print
- [ ] Print queue management
- [ ] Print history tracking

### API Enhancements
- [ ] Store print history
- [ ] Print analytics
- [ ] Template management
- [ ] Organization-wide preferences

## Best Practices

### For Developers
1. Always capture chart canvas before opening dialogs
2. Handle canvas capture errors gracefully
3. Test across multiple browsers
4. Validate user input in Print Setup
5. Provide visual feedback during print
6. Clean up resources after print

### For Users
1. Use Print Preview before printing
2. Save preferences for repeated use
3. Test with one page before batch printing
4. Use landscape for most charts
5. Include legends for complex charts
6. Verify printer settings match preferences

## Support

For issues or questions:
- Check browser console for errors
- Verify chart is loaded before printing
- Test with default preferences
- Try different browsers
- Contact support with error logs
