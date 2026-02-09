# Chart Export - Quick Reference Guide

## Quick Start

### User Guide

1. **Open the Export Dialog**
   - Click `File` menu → `Save As Picture`
   - Or use keyboard shortcut (if configured)

2. **Choose Format**
   - **PNG:** Best quality, larger files
   - **JPG:** Smaller files, adjustable quality

3. **Select Options**
   - **Scale:** 1x (standard) / 2x (retina) / 3x (ultra)
   - **Overlays:** Toggle symbol info, timestamp, watermark
   - **Quality:** JPG only - adjust slider (10-100%)

4. **Export**
   - Click "Save Picture"
   - File downloads automatically
   - Named: `SYMBOL_TIMEFRAME_TIMESTAMP.ext`

### Developer Quick Start

```typescript
// 1. Import the service
import { ChartExporter } from '@/services/chartExporter';

// 2. Get canvas reference
const canvas = document.querySelector('canvas');

// 3. Prepare metadata
const metadata = {
  symbol: 'EURUSD',
  timeframe: '5m',
  timestamp: new Date()
};

// 4. Export
const blob = await ChartExporter.exportCanvas(canvas, metadata);

// 5. Download
const filename = ChartExporter.generateFilename(metadata, 'png');
ChartExporter.downloadBlob(blob, filename);
```

## API Reference

### ChartExporter Methods

```typescript
// Export canvas to blob
static async exportCanvas(
  canvas: HTMLCanvasElement,
  metadata: ChartMetadata,
  options?: Partial<ChartExportOptions>
): Promise<Blob>

// Export DOM element to blob
static async exportElement(
  element: HTMLElement,
  metadata: ChartMetadata,
  options?: Partial<ChartExportOptions>
): Promise<Blob>

// Auto-detect and export active chart
static async exportActiveChart(
  metadata: ChartMetadata,
  options?: Partial<ChartExportOptions>
): Promise<Blob>

// Trigger browser download
static downloadBlob(blob: Blob, filename: string): void

// Generate standard filename
static generateFilename(
  metadata: ChartMetadata,
  format: 'png' | 'jpg'
): string

// Get preview thumbnail
static getPreview(
  canvas: HTMLCanvasElement,
  maxWidth?: number,
  maxHeight?: number
): string
```

### Types

```typescript
interface ChartExportOptions {
  format: 'png' | 'jpg';
  quality: number;              // 0.1 to 1.0 (JPG only)
  scale: number;                // 1, 2, or 3
  exportArea: 'visible' | 'full';
  includeWatermark: boolean;
  watermarkText?: string;
  includeTimestamp: boolean;
  includeSymbolInfo: boolean;
  backgroundColor?: string;
}

interface ChartMetadata {
  symbol: string;
  timeframe: string;
  timestamp: Date;
}
```

### Default Options

```typescript
{
  format: 'png',
  quality: 0.92,
  scale: 1,
  exportArea: 'visible',
  includeWatermark: false,
  includeTimestamp: true,
  includeSymbolInfo: true,
  backgroundColor: '#000000'
}
```

## Common Use Cases

### High-Quality Print Export

```typescript
await ChartExporter.exportCanvas(canvas, metadata, {
  format: 'png',
  scale: 3,
  includeSymbolInfo: true,
  includeTimestamp: true
});
```

### Quick Email Export

```typescript
await ChartExporter.exportCanvas(canvas, metadata, {
  format: 'jpg',
  quality: 0.75,
  scale: 1,
  includeTimestamp: false
});
```

### Branded Export

```typescript
await ChartExporter.exportCanvas(canvas, metadata, {
  format: 'png',
  scale: 2,
  includeWatermark: true,
  watermarkText: 'My Trading System',
  includeSymbolInfo: true,
  includeTimestamp: true
});
```

### Web Sharing Export

```typescript
await ChartExporter.exportCanvas(canvas, metadata, {
  format: 'jpg',
  quality: 0.85,
  scale: 1,
  includeWatermark: false,
  includeTimestamp: false
});
```

## Component Usage

### SavePictureDialog Props

```typescript
interface SavePictureDialogProps {
  isOpen: boolean;
  onClose: () => void;
  chartCanvas?: HTMLCanvasElement | null;
  chartContainer?: HTMLElement | null;
  symbol: string;
  timeframe: string;
}
```

### Basic Integration

```tsx
import { SavePictureDialog } from '@/components/dialogs/SavePictureDialog';

function MyComponent() {
  const [showDialog, setShowDialog] = useState(false);

  return (
    <>
      <button onClick={() => setShowDialog(true)}>
        Export Chart
      </button>

      <SavePictureDialog
        isOpen={showDialog}
        onClose={() => setShowDialog(false)}
        symbol="EURUSD"
        timeframe="5m"
      />
    </>
  );
}
```

### With Custom Canvas Reference

```tsx
function MyChart() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [showDialog, setShowDialog] = useState(false);

  return (
    <>
      <canvas ref={canvasRef} />

      <SavePictureDialog
        isOpen={showDialog}
        onClose={() => setShowDialog(false)}
        chartCanvas={canvasRef.current}
        symbol="EURUSD"
        timeframe="5m"
      />
    </>
  );
}
```

## File Naming

### Format

```
{SYMBOL}_{TIMEFRAME}_{TIMESTAMP}.{ext}
```

### Examples

```
EURUSD_5m_2024-01-01T12-30-45.png
GBPUSD_1h_2024-01-01T15-45-30.jpg
BTCUSD_1d_2024-01-01T09-00-00.png
```

### Sanitization Rules

- Special characters removed from symbol
- Slashes and spaces replaced
- Colons in timestamp replaced with hyphens
- ISO 8601 format for timestamp

## Keyboard Shortcuts (Suggested)

| Shortcut | Action |
|----------|--------|
| `Ctrl+E` | Open export dialog |
| `Ctrl+Shift+S` | Quick PNG export |
| `Enter` | Confirm export |
| `Esc` | Cancel dialog |

## Performance Tips

1. **Use JPG for large charts:** 50% faster than PNG
2. **Use 1x scale for quick exports:** 4x faster than 2x
3. **Disable overlays if not needed:** Slightly faster
4. **Avoid 3x scale unless needed:** Very large files

## File Size Reference

| Configuration | Approximate Size |
|--------------|------------------|
| 1920×1080, PNG, 1x | 200-500 KB |
| 1920×1080, PNG, 2x | 800 KB - 2 MB |
| 1920×1080, PNG, 3x | 2-5 MB |
| 1920×1080, JPG 92%, 1x | 100-300 KB |
| 1920×1080, JPG 75%, 1x | 50-150 KB |

## Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| "No chart available to export" | Canvas not found | Ensure chart is rendered |
| "Failed to get 2D context" | Canvas error | Refresh page |
| "Failed to create blob" | Browser limitation | Try smaller scale |
| "Export failed" | General error | Check console for details |

## Browser Support

| Browser | PNG | JPG | Scale | Overlays |
|---------|-----|-----|-------|----------|
| Chrome 90+ | ✅ | ✅ | ✅ | ✅ |
| Firefox 88+ | ✅ | ✅ | ✅ | ✅ |
| Safari 14+ | ✅ | ✅ | ✅ | ✅ |
| Edge 90+ | ✅ | ✅ | ✅ | ✅ |

## Troubleshooting

**Preview not showing?**
- Canvas might be empty
- Chart might not be rendered yet

**Export takes too long?**
- Reduce scale (use 1x instead of 2x/3x)
- Use JPG instead of PNG

**File size too large?**
- Use JPG format
- Reduce quality slider
- Use lower scale

**Timestamp wrong timezone?**
- Uses browser's local timezone
- Convert before export if needed

## Testing Commands

```bash
# Run all tests
npm run test

# Run export tests only
npm run test -- chartExporter.test.ts
npm run test -- SavePictureDialog.test.tsx

# Run with coverage
npm run test:coverage

# Watch mode
npm run test -- --watch
```

## Related Documentation

- [Full Documentation](./CHART_EXPORT_FEATURE.md)
- [Integration Guide](./INTEGRATION_TEST_PLAN.md)
- [Market Watch Documentation](./MARKETWATCH_QUICK_REFERENCE.md)

---

**Quick Links:**
- [Source Code](../clients/desktop/src/services/chartExporter.ts)
- [Dialog Component](../clients/desktop/src/components/dialogs/SavePictureDialog.tsx)
- [Tests](../clients/desktop/src/services/chartExporter.test.ts)

**Last Updated:** 2026-02-02
