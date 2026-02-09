# Chart Export Feature Documentation

## Overview

The Chart Export feature allows users to save trading charts as high-quality images (PNG or JPG format) with various customization options including overlays, scaling, and quality settings.

## Architecture

### Core Components

#### 1. ChartExporter Service (`services/chartExporter.ts`)
The main service class that handles all chart export functionality.

**Key Methods:**
- `exportCanvas()` - Export a canvas element to an image blob
- `exportElement()` - Export a DOM element using html2canvas (fallback)
- `exportActiveChart()` - Auto-detect and export the active chart
- `downloadBlob()` - Trigger browser download
- `generateFilename()` - Generate standardized filenames
- `getPreview()` - Generate thumbnail previews

**Features:**
- Multiple format support (PNG/JPG)
- Resolution scaling (1x, 2x, 3x for Retina displays)
- Quality control for JPG exports
- Overlay rendering (symbol info, timestamp, watermark)
- Canvas compositing with metadata

#### 2. SavePictureDialog Component (`components/dialogs/SavePictureDialog.tsx`)
Modal dialog providing UI for export configuration.

**Features:**
- Live preview of chart
- Format selection (PNG/JPG)
- Quality slider for JPG
- Resolution scale selector
- Export area selection (visible/full)
- Overlay toggles:
  - Symbol & Timeframe label
  - Timestamp
  - Custom watermark
- Real-time error handling
- Progress indication during export

#### 3. useChartExport Hook (`hooks/useChartExport.ts`)
React hook for managing chart export state.

**Provides:**
- Dialog state management
- Chart canvas/container references
- Auto-detection of chart elements
- Symbol and timeframe tracking

### Integration Points

#### FileMenu Integration
The "Save As Picture" menu item is integrated into the File menu:

```tsx
// components/layout/FileMenu.tsx
const handleSaveAsPicture = () => {
  setShowSavePictureDialog(true);
};

<SavePictureDialog
  isOpen={showSavePictureDialog}
  onClose={() => setShowSavePictureDialog(false)}
  symbol={selectedSymbol || 'EURUSD'}
  timeframe={timeframe || '5m'}
/>
```

#### Global State Integration
Uses Zustand store for current symbol and timeframe:

```tsx
const selectedSymbol = useAppStore(state => state.selectedSymbol);
const timeframe = useAppStore(state => state.timeframe);
```

## Export Options

### Format Options

#### PNG (Lossless)
- **Pros:** Perfect quality, transparency support, no compression artifacts
- **Cons:** Larger file size
- **Best for:** Charts with text, technical indicators, sharing online
- **Default quality:** N/A (lossless)

#### JPG (Compressed)
- **Pros:** Smaller file size
- **Cons:** Lossy compression, no transparency
- **Best for:** Quick sharing, email attachments, storage
- **Quality range:** 10-100% (default: 92%)

### Resolution Scaling

| Scale | Use Case | File Size | DPI |
|-------|----------|-----------|-----|
| **1x** | Standard displays, web sharing | Smallest | 96 DPI |
| **2x** | Retina displays, high-quality prints | Medium | 192 DPI |
| **3x** | Ultra-high DPI, professional prints | Largest | 288 DPI |

### Export Areas

- **Visible Area:** Exports only what's currently visible on screen
- **Full Chart:** Exports entire chart including off-screen data (future enhancement)

### Overlays

#### Symbol & Timeframe Label
- Position: Top-left corner
- Format: `{SYMBOL} - {TIMEFRAME}`
- Example: `EURUSD - 5m`
- Semi-transparent background for readability

#### Timestamp
- Position: Bottom-right corner
- Format: Locale-specific date/time
- Example: `2/2/2026, 10:30:45 AM`
- Useful for documenting market conditions

#### Watermark
- Position: Center (diagonal)
- Customizable text
- Default: "Trading Engine"
- Semi-transparent (10% opacity)
- Rotated -30 degrees

## Usage Examples

### Basic Export

```typescript
import { ChartExporter } from './services/chartExporter';

const canvas = document.querySelector('canvas');
const metadata = {
  symbol: 'EURUSD',
  timeframe: '5m',
  timestamp: new Date()
};

const blob = await ChartExporter.exportCanvas(canvas, metadata);
const filename = ChartExporter.generateFilename(metadata, 'png');
ChartExporter.downloadBlob(blob, filename);
```

### Custom Options Export

```typescript
const blob = await ChartExporter.exportCanvas(canvas, metadata, {
  format: 'jpg',
  quality: 0.85,
  scale: 2,
  includeWatermark: true,
  watermarkText: 'My Trading System',
  includeTimestamp: true,
  includeSymbolInfo: true,
  backgroundColor: '#000000'
});
```

### Auto-Detect Active Chart

```typescript
// Automatically finds and exports the largest canvas
const blob = await ChartExporter.exportActiveChart(metadata, options);
```

### Get Preview

```typescript
const previewDataUrl = ChartExporter.getPreview(canvas, 300, 200);
// Use in <img src={previewDataUrl} />
```

## File Naming Convention

Exported files follow this pattern:
```
{SYMBOL}_{TIMEFRAME}_{TIMESTAMP}.{ext}
```

**Examples:**
- `EURUSD_5m_2024-01-01T12-30-45.png`
- `GBPUSD_1h_2024-01-01T15-45-30.jpg`
- `BTCUSD_1d_2024-01-01T09-00-00.png`

**Rules:**
- Symbols are sanitized (special characters removed)
- Timestamps use ISO format with hyphens instead of colons
- Extensions match the selected format

## Technical Implementation

### Canvas Compositing

The exporter uses a temporary canvas for compositing:

1. Create temporary canvas with scaled dimensions
2. Fill background color
3. Draw original chart (scaled)
4. Draw overlays (symbol info, timestamp, watermark)
5. Convert to blob with specified format and quality

### Resolution Scaling

Scaling is applied during the compositing phase:

```typescript
tempCanvas.width = originalWidth * scale;
tempCanvas.height = originalHeight * scale;
ctx.scale(scale, scale);
ctx.drawImage(originalCanvas, 0, 0);
```

This ensures sharp rendering on high-DPI displays.

### Browser Compatibility

**Supported Features:**
- ✅ Canvas API (all modern browsers)
- ✅ Blob API (all modern browsers)
- ✅ Download via `<a>` element (all modern browsers)
- ⚠️ html2canvas (optional, for complex DOM export)

**Fallback Strategy:**
1. Try canvas export (primary)
2. Try html2canvas if available (fallback)
3. Show error if neither works

## Testing

### Unit Tests

```bash
npm run test -- chartExporter.test.ts
npm run test -- SavePictureDialog.test.tsx
```

### Test Coverage

**ChartExporter Service:**
- ✅ Filename generation
- ✅ Canvas export (PNG/JPG)
- ✅ Scale factor application
- ✅ Overlay rendering
- ✅ Preview generation
- ✅ Download triggering
- ✅ Active chart detection
- ✅ Error handling

**SavePictureDialog Component:**
- ✅ Rendering states (open/closed)
- ✅ Preview display
- ✅ Format selection
- ✅ Quality slider (JPG)
- ✅ Scale selection
- ✅ Overlay toggles
- ✅ Watermark input
- ✅ Export execution
- ✅ Error handling
- ✅ Loading states

### Manual Testing Checklist

- [ ] Open chart with data
- [ ] Click File → Save As Picture
- [ ] Verify preview displays correctly
- [ ] Test PNG export
- [ ] Test JPG export with different quality levels
- [ ] Test 1x, 2x, 3x scaling
- [ ] Test all overlay options
- [ ] Test custom watermark text
- [ ] Verify filename format
- [ ] Test with different symbols and timeframes
- [ ] Test error handling (no chart available)
- [ ] Test cancel button
- [ ] Test close button (X)

## Performance Considerations

### Optimization Strategies

1. **Preview Throttling:** Preview generation is debounced to avoid excessive rendering
2. **Lazy Loading:** html2canvas is loaded only when needed
3. **Memory Management:** Temporary canvases are created and disposed properly
4. **Blob URLs:** Object URLs are revoked after download to prevent memory leaks

### Performance Metrics

| Operation | Expected Time |
|-----------|--------------|
| Preview generation | <100ms |
| 1x PNG export | 200-500ms |
| 2x PNG export | 500-1000ms |
| 3x PNG export | 1-2s |
| JPG export (any scale) | 50% faster than PNG |

### Memory Usage

- Temporary canvas: ~Width × Height × 4 bytes × scale²
- Example: 1920×1080 @ 2x = ~33MB peak usage

## Future Enhancements

### Planned Features

1. **Full Chart Export:** Export entire chart including off-screen data
2. **Batch Export:** Export multiple timeframes at once
3. **Template System:** Save and load export presets
4. **Cloud Upload:** Direct upload to cloud storage
5. **Print Preview:** Better print-specific formatting
6. **PDF Export:** Multi-page PDF reports
7. **Annotation Tools:** Add notes/arrows before export
8. **Comparison Export:** Side-by-side chart comparison images

### Integration Opportunities

- [ ] Desktop app: Native file dialogs (Electron)
- [ ] Mobile app: Share to apps (iOS/Android)
- [ ] Social media: Direct sharing to platforms
- [ ] Trading journal: Auto-export on trade events
- [ ] Email: Send chart via email from app

## Troubleshooting

### Common Issues

**Issue: Preview not showing**
- Check if canvas element exists and has content
- Verify canvas dimensions are valid
- Check browser console for errors

**Issue: Export fails**
- Ensure canvas has valid 2D context
- Check browser compatibility
- Verify no CORS issues with chart images

**Issue: Downloaded file is black/empty**
- Canvas may not be fully rendered
- Try adding a small delay before export
- Check if chart is hidden (display: none)

**Issue: File size too large**
- Use JPG format instead of PNG
- Reduce quality setting (JPG)
- Use 1x scale instead of 2x/3x
- Disable overlays if not needed

### Debug Mode

Enable debug logging:

```typescript
// In chartExporter.ts, add:
const DEBUG = true;

if (DEBUG) {
  console.log('Export options:', options);
  console.log('Canvas dimensions:', canvas.width, canvas.height);
  console.log('Blob size:', blob.size, 'bytes');
}
```

## Security Considerations

- ✅ No external API calls (runs entirely client-side)
- ✅ No data sent to servers
- ✅ Safe filename sanitization
- ✅ User-controlled export options
- ⚠️ Large exports may cause memory issues (limited by scale limits)

## Accessibility

- ✅ Keyboard navigation support
- ✅ Screen reader labels
- ✅ Clear error messages
- ✅ Visual feedback for actions
- ✅ Cancel/close options available

## License

This feature is part of the Trading Engine project and follows the same license terms.

## Support

For issues or questions:
1. Check this documentation
2. Review test files for examples
3. Check browser console for errors
4. Open an issue on GitHub

---

**Last Updated:** 2026-02-02
**Version:** 1.0.0
**Author:** Trading Engine Team
