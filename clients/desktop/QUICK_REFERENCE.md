# Save As Picture - Quick Reference

## For Developers

### Quick Test in Console
```javascript
// 1. Open app and load a chart
// 2. Open browser console (F12)
// 3. Run:
await testChartExport()
```

### Manual Export Code
```typescript
import { ChartExporter } from './services/chartExporter';

const canvas = (window as any).__activeChartCanvas;
const metadata = {
  symbol: 'EURUSD',
  timeframe: '5m',
  timestamp: new Date()
};

const blob = await ChartExporter.exportCanvas(canvas, metadata, {
  format: 'png',
  scale: 2,
  includeSymbolInfo: true,
  includeTimestamp: true
});

ChartExporter.downloadBlob(blob, 'my-chart.png');
```

### Check Chart Reference
```javascript
// Should return HTMLCanvasElement if chart is loaded
console.log(window.__activeChartCanvas);
console.log(window.__activeChartContainer);
```

## For Users

### Export Steps
1. Load a chart
2. File → Save As Picture
3. Configure options
4. Click "Save Picture"
5. File downloads automatically

### Recommended Settings
**General Use**:
- Format: PNG
- Scale: 1x or 2x
- Symbol overlay: On
- Timestamp: On

**Sharing**:
- Format: JPG
- Quality: 80%
- Watermark: On (your name)

## Files Reference

### Core Files
```
src/components/TradingChart.tsx          # Canvas source
src/components/layout/FileMenu.tsx       # Menu integration
src/components/dialogs/SavePictureDialog.tsx  # Export UI
src/services/chartExporter.ts            # Export logic
```

### Documentation
```
SAVE_AS_PICTURE_IMPLEMENTATION.md        # Technical details
CHART_EXPORT_GUIDE.md                    # User guide
SAVE_AS_PICTURE_SUMMARY.md               # Project summary
QUICK_REFERENCE.md                       # This file
```

### Test Files
```
src/utils/testChartExport.ts             # Test utilities
```

## Common Patterns

### Get Active Canvas
```typescript
const canvas = (window as any).__activeChartCanvas as HTMLCanvasElement;
if (!canvas) {
  console.error('No chart available');
  return;
}
```

### Export with Custom Options
```typescript
const options: Partial<ChartExportOptions> = {
  format: 'jpg',
  quality: 0.85,
  scale: 2,
  includeWatermark: true,
  watermarkText: 'Trading Engine Pro'
};

const blob = await ChartExporter.exportCanvas(canvas, metadata, options);
```

### Generate Preview
```typescript
const previewDataURL = ChartExporter.getPreview(canvas, 300, 200);
// Use in <img src={previewDataURL} />
```

## Troubleshooting

### Canvas is null
```javascript
// Wait for chart to initialize
setTimeout(() => {
  const canvas = window.__activeChartCanvas;
  // Try export again
}, 1000);
```

### Export fails silently
```javascript
// Add try-catch
try {
  const blob = await ChartExporter.exportCanvas(...);
  console.log('Export successful');
} catch (error) {
  console.error('Export failed:', error);
}
```

### File doesn't download
```javascript
// Check blob creation
const blob = await ChartExporter.exportCanvas(...);
console.log('Blob size:', blob.size, 'Type:', blob.type);

// Manually trigger download
const url = URL.createObjectURL(blob);
const a = document.createElement('a');
a.href = url;
a.download = 'chart.png';
a.click();
URL.revokeObjectURL(url);
```

## Performance Tips

### Optimize Export Speed
- Use 1x scale for quick exports
- Use JPG format (faster than PNG)
- Disable overlays if not needed
- Export visible area only

### Optimize File Size
- Use JPG with 70-80% quality
- Use 1x scale
- Disable unnecessary overlays

### Optimize Quality
- Use PNG format (lossless)
- Use 2x or 3x scale
- Enable all overlays

## API Quick Reference

### ChartExporter Methods
```typescript
// Export from canvas
exportCanvas(canvas, metadata, options): Promise<Blob>

// Export from DOM element
exportElement(element, metadata, options): Promise<Blob>

// Auto-detect and export
exportActiveChart(metadata, options): Promise<Blob>

// Download blob
downloadBlob(blob, filename): void

// Generate filename
generateFilename(metadata, format): string

// Get preview
getPreview(canvas, maxWidth, maxHeight): string
```

### Types
```typescript
interface ChartMetadata {
  symbol: string;
  timeframe: string;
  timestamp: Date;
}

interface ChartExportOptions {
  format: 'png' | 'jpg';
  quality: number;        // 0.1 to 1.0
  scale: number;          // 1, 2, 3
  exportArea: 'visible' | 'full';
  includeWatermark: boolean;
  watermarkText?: string;
  includeTimestamp: boolean;
  includeSymbolInfo: boolean;
  backgroundColor?: string;
}
```

## Version Info

- **Version**: 1.0.0
- **Status**: Production Ready ✅
- **Last Updated**: 2026-02-02
- **Dependencies**: lightweight-charts

## Quick Links

- [User Guide](./CHART_EXPORT_GUIDE.md)
- [Technical Docs](./SAVE_AS_PICTURE_IMPLEMENTATION.md)
- [Project Summary](../SAVE_AS_PICTURE_SUMMARY.md)

---

**Need Help?** Check the full documentation or contact support.
