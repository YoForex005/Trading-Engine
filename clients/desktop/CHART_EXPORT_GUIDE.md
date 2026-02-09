# Chart Export User Guide

## How to Export Charts as Images

### Quick Start

1. **Open a Chart**
   - Load any trading chart (EURUSD, BTCUSD, etc.)

2. **Access Export Function**
   - Click **File** menu → **Save As Picture**
   - Or use keyboard shortcut (if configured)

3. **Configure Export Options**
   - Choose format (PNG or JPG)
   - Adjust quality/scale as needed
   - Enable/disable overlays

4. **Export**
   - Click **Save Picture** button
   - Image downloads automatically

## Export Options Explained

### Image Format

#### PNG (Recommended)
- **Pros**: Lossless quality, perfect text clarity, supports transparency
- **Cons**: Larger file size
- **Best for**: Charts with text, sharing online, printing

#### JPG
- **Pros**: Smaller file size, adjustable quality
- **Cons**: Lossy compression, may blur text slightly
- **Best for**: Email attachments, quick shares, space-constrained situations

### Quality Settings

**JPG Only** - Slider from 10% to 100%

| Quality | File Size | Use Case |
|---------|-----------|----------|
| 10-30%  | Smallest  | Quick preview, thumbnails |
| 40-60%  | Small     | Email, messaging apps |
| 70-85%  | Medium    | General sharing |
| 90-100% | Large     | Archival, analysis |

**Recommended**: 92% (good balance)

### Resolution Scale

| Scale | Output Size | Use Case |
|-------|-------------|----------|
| 1x    | Standard    | Web sharing, fast export |
| 2x    | 2x larger   | Retina displays, presentations |
| 3x    | 3x larger   | High-DPI displays, large prints |

**Note**: Higher scales increase file size and export time

### Export Area

#### Visible Area (Default)
- Exports only what you see on screen
- Faster export
- Smaller file size

#### Full Chart
- Exports all historical data
- May be very large
- Use for comprehensive archives

### Chart Overlays

#### Symbol & Timeframe
- Adds trading pair and timeframe to top-left
- Example: "EURUSD - 5m"
- Recommended: ✅ On

#### Timestamp
- Adds export date/time to bottom-right
- Example: "2/2/2024, 2:30:15 PM"
- Recommended: ✅ On

#### Watermark
- Adds semi-transparent text diagonally
- Custom text allowed
- Recommended: Optional (for professional sharing)

## File Naming

Exported files follow this pattern:
```
{SYMBOL}_{TIMEFRAME}_{TIMESTAMP}.{FORMAT}
```

### Examples
- `EURUSD_5m_2024-02-02T14-30-15.png`
- `BTCUSD_1h_2024-02-02T14-30-15.jpg`
- `GBPUSD_15m_2024-02-02T14-30-15.png`

**Tip**: Files are automatically named - no need to type anything!

## Common Use Cases

### 1. Share on Social Media
```
Format: PNG
Quality: N/A
Scale: 1x
Overlays: Symbol ✅, Timestamp ✅, Watermark ✅
```

### 2. Email to Colleague
```
Format: JPG
Quality: 80%
Scale: 1x
Overlays: Symbol ✅, Timestamp ✅, Watermark ❌
```

### 3. Trading Journal/Documentation
```
Format: PNG
Quality: N/A
Scale: 2x
Overlays: Symbol ✅, Timestamp ✅, Watermark ❌
```

### 4. Presentation/Report
```
Format: PNG
Quality: N/A
Scale: 2x or 3x
Overlays: Symbol ✅, Timestamp ✅, Watermark ✅
```

### 5. Quick Screenshot for Chat
```
Format: JPG
Quality: 60%
Scale: 1x
Overlays: Symbol ✅, Timestamp ❌, Watermark ❌
```

## Tips & Best Practices

### For Best Quality
- ✅ Use PNG format
- ✅ Use 2x scale for Retina displays
- ✅ Enable symbol & timeframe overlay
- ✅ Export during low system load

### For Smallest File Size
- ✅ Use JPG format
- ✅ Lower quality to 60-70%
- ✅ Use 1x scale
- ✅ Export visible area only

### For Professional Sharing
- ✅ Use PNG format
- ✅ Use 2x scale
- ✅ Enable all overlays
- ✅ Add custom watermark with your name/company

### For Analysis/Reference
- ✅ Use PNG format
- ✅ Use 1x or 2x scale
- ✅ Enable timestamp
- ✅ Keep symbol overlay

## Keyboard Shortcuts

| Action | Shortcut | Notes |
|--------|----------|-------|
| Open Export Dialog | (Not set) | Use File menu |
| Close Dialog | Esc | Cancels export |
| Save/Export | Enter | When dialog is focused |

**Tip**: You can request custom keyboard shortcuts from the development team!

## Troubleshooting

### "No chart available to export"
**Problem**: No active chart loaded
**Solution**: Open a chart first, then try exporting

### Preview shows black/blank image
**Problem**: Chart not fully rendered yet
**Solution**: Wait a few seconds after loading chart, then try again

### Export button doesn't work
**Problem**: Chart data not ready
**Solution**: Refresh page and try again

### File doesn't download
**Problem**: Browser blocked download
**Solution**: Check browser settings, allow downloads from this site

### Image quality is poor
**Problem**: Low JPG quality or 1x scale
**Solution**: Use PNG or increase JPG quality to 90%+

### Export takes a long time
**Problem**: High resolution scale (3x) with large chart
**Solution**: Use 1x or 2x scale for faster exports

## Advanced Features

### Custom Watermarks
Add your personal or company branding:
1. Enable "Include Watermark" checkbox
2. Enter custom text (e.g., "© Your Name 2024")
3. Text appears diagonally across chart

### Multiple Exports
To export multiple charts:
1. Switch to first chart
2. Export
3. Switch to next chart
4. Export
5. Repeat as needed

### Batch Export (Future Feature)
Coming soon: Export all open charts at once!

## Technical Details

### Supported Formats
- PNG (Portable Network Graphics)
- JPG/JPEG (Joint Photographic Experts Group)

### Maximum Resolution
- Limited by browser memory
- 3x scale recommended maximum
- Larger scales may fail on low-memory devices

### Browser Compatibility
- ✅ Chrome/Edge (Chromium)
- ✅ Firefox
- ✅ Safari
- ✅ Brave
- ✅ Opera

### File Size Estimates

Based on typical 1920x1080 chart:

| Format | Quality | Scale | Approx. Size |
|--------|---------|-------|--------------|
| PNG    | N/A     | 1x    | 100-300 KB   |
| PNG    | N/A     | 2x    | 400-800 KB   |
| PNG    | N/A     | 3x    | 800-1500 KB  |
| JPG    | 60%     | 1x    | 50-100 KB    |
| JPG    | 80%     | 1x    | 80-150 KB    |
| JPG    | 92%     | 1x    | 100-200 KB   |

**Note**: Actual sizes vary based on chart complexity

## FAQ

**Q: Can I export multiple charts at once?**
A: Not yet - currently one at a time. Batch export coming soon!

**Q: Where are my exported files saved?**
A: Your browser's default download folder (usually ~/Downloads)

**Q: Can I change the filename?**
A: Files are auto-named. You can rename after download.

**Q: Does export include indicators?**
A: Yes! Everything visible on the chart is exported.

**Q: Can I export to PDF?**
A: Not currently. Use Print functionality for PDF export.

**Q: Is there a limit to file size?**
A: Browser-dependent, but generally no practical limit.

**Q: Can I export historical data not on screen?**
A: Use "Full Chart" export area option (may be large).

**Q: Does export work offline?**
A: Yes! No internet connection needed for export.

**Q: Can I automate exports?**
A: Not yet - manual only. API coming in future version.

## Need Help?

If you encounter issues:

1. Check this guide for solutions
2. Check browser console for errors (F12)
3. Try refreshing the page
4. Contact support with:
   - Browser name and version
   - Chart being exported
   - Error message (if any)
   - Screenshot of issue

## Version History

### v1.0.0 (Current)
- ✅ PNG/JPG export
- ✅ Quality control
- ✅ Resolution scaling (1x, 2x, 3x)
- ✅ Overlay options
- ✅ Preview generation
- ✅ Auto-naming

### Planned Features
- 🔄 Batch export (multiple charts)
- 🔄 Export to clipboard
- 🔄 Custom dimensions
- 🔄 Export templates/presets
- 🔄 Keyboard shortcuts
- 🔄 Cloud upload integration

---

**Happy Trading! 📈**
