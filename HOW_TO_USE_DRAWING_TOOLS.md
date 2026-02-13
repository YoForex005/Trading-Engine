# How to Use Drawing Tools - Quick Guide

## 🎯 Quick Start (3 Steps)

1. **Click a drawing tool button** in the top toolbar (Rectangle, Ellipse, Pitchfork, etc.)
2. **Click on the chart** to place anchor points
3. **Done!** Drawing auto-saves when you place the required number of points

---

## 📐 All Drawing Tools

### Single-Click Tools (Click Once)
- **Horizontal Line** - Click once to draw horizontal price level
- **Vertical Line** - Click once to draw vertical time marker
- **Text** - Click once, enter text in prompt
- **Shapes** - Click once to place shape (👍👎↑↓🛑✅▲▼➔)
- **Arrow** - Click once to place arrow (up/down/right)

### Two-Click Tools (Click Twice)
- **Trendline** - Click start point, then end point
- **Rectangle** - Click top-left corner, then bottom-right corner
- **Ellipse** - Click first corner, then opposite corner
- **Fibonacci** - Click start point, then end point (shows 6 levels)
- **Channel** - Click start point, then end point

### Three-Click Tools (Click 3 Times)
- **Pitchfork** - Click center handle, then upper point, then lower point

---

## ✏️ Editing Drawings

### Select a Drawing
- **Click** on any drawing to select it
- **Selected drawings** show red squares on anchor points

### Move Entire Drawing
- **Click and drag** the drawing body to move it

### Adjust Anchor Points
- **Click and drag** the red squares to adjust individual points

### Multi-Select
- **Ctrl+Click** to select multiple drawings

### Deselect
- **Click** on empty chart area
- **Press Esc** key

---

## 🎨 Customize Appearance

### Properties Panel (Auto-Opens on Selection)
1. **Color**
   - Click any preset color
   - OR use color picker for custom color
   - OR type hex code (#3b82f6)

2. **Line Style**
   - Solid (━━━)
   - Dashed (╌╌╌)
   - Dotted (┄┄┄)

3. **Line Width**
   - Drag slider: 1-5px
   - OR click number buttons

4. **Special Options**
   - **Trendline:** Extend left/right checkboxes
   - **Fibonacci:** Show price labels checkbox
   - **Text:** Multi-line text input

5. **Actions**
   - **Duplicate:** Clone drawing with offset
   - **Delete:** Remove drawing

---

## 📋 Manage Drawings

### Drawing List Panel
**Toggle:** Click list icon in toolbar

**Features:**
- See all drawings for current symbol
- Click item to select on chart
- Eye icon to show/hide
- Trash icon to delete
- "Clear All" button at bottom

---

## ⌨️ Keyboard Shortcuts

| Key | Action |
|-----|--------|
| **Delete** | Remove selected drawing(s) |
| **Esc** | Cancel active drawing / Deselect all |
| **Ctrl+Click** | Multi-select drawings |

---

## 🖱️ Right-Click Context Menu

**Right-click on any drawing** to access:
- Duplicate
- Delete
- Properties
- Lock/Unlock

---

## 💾 Auto-Save Features

✅ **Automatic Backend Save**
- Drawings save to backend when:
  - Created
  - Moved
  - Modified
  - Deleted

✅ **Persistent Storage**
- Drawings persist across page reloads
- Per-symbol storage (drawings only show on correct chart)
- LocalStorage fallback if backend unavailable

---

## 🔍 Tool-Specific Tips

### Fibonacci Retracement
- **6 Levels Shown:** 0%, 23.6%, 38.2%, 50%, 61.8%, 100%
- **Styling:** 0% and 100% are solid, others are dashed
- **Labels:** Percentage labels on right side

### Pitchfork (Andrews)
- **3 Lines:** Upper, Median (dashed), Lower
- **Usage:** Click center first, then two extremes

### Rectangle & Ellipse
- **All line styles work:** Solid, dashed, dotted borders
- **Transparent fill:** Only border is visible
- **Perfect ellipse:** Uses CSS border-radius: 50%

### Shapes
- **9 Subtypes Available:**
  - 👍 Thumbs Up
  - 👎 Thumbs Down
  - ↑ Arrow Up
  - ↓ Arrow Down
  - 🛑 Stop
  - ✅ Check
  - ▲ Buy Signal
  - ▼ Sell Signal
  - ➔ Right Arrow

---

## 🎯 Common Workflows

### Mark Support/Resistance
1. Click **Horizontal Line** tool
2. Click on support/resistance level
3. Optionally: Change color in properties panel

### Fibonacci Analysis
1. Click **Fibonacci** tool
2. Click swing low
3. Click swing high
4. View 6 retracement levels

### Trend Analysis
1. Click **Trendline** tool
2. Click first touch point
3. Click second touch point
4. Optionally: Enable "Extend Right" in properties

### Annotations
1. Click **Text** tool
2. Click location on chart
3. Type your note (e.g., "Potential breakout zone")
4. Click OK

### Chart Patterns
1. Use **Rectangle** or **Channel** for ranges
2. Use **Pitchfork** for price action analysis
3. Use **Shapes** for entry/exit markers

---

## 🚨 Troubleshooting

### Drawing Not Appearing?
- Check if drawing is hidden (eye icon in list panel)
- Ensure you clicked enough times for the tool type
- Zoom out to see if drawing is off-screen

### Can't Select Drawing?
- Make sure you're in cursor mode (not active drawing mode)
- Click directly on the drawing line/shape
- Try clicking the list panel instead

### Properties Panel Not Opening?
- Drawing must be selected first
- Click on the drawing to select it
- Panel appears on right side

### Drawings Lost After Reload?
- Check backend API connection
- Look in browser localStorage as fallback
- Ensure correct symbol is selected

---

## 📊 Best Practices

✅ **Use Colors Wisely**
- Red for resistance, bearish
- Green for support, bullish
- Blue for neutral analysis

✅ **Keep Chart Clean**
- Delete old/irrelevant drawings
- Use "Clear All" periodically
- Hide instead of delete if unsure

✅ **Label Everything**
- Use text annotations
- Add context to your analysis
- Future you will thank you

✅ **Lock Important Drawings**
- Right-click → Lock
- Prevents accidental modification
- Unlock when needed

---

## 🎓 Pro Tips

💡 **Duplicate for Similar Levels**
- Create one drawing
- Customize color, style, width
- Duplicate for similar analysis
- Faster than creating from scratch

💡 **Use Visibility Toggle**
- Hide drawings you don't need right now
- Reduces chart clutter
- Easy to restore later

💡 **Multi-Select for Batch Delete**
- Ctrl+Click multiple drawings
- Press Delete once
- Removes all selected

💡 **Organize by Color**
- All support lines = green
- All resistance lines = red
- All annotations = yellow
- Easy visual scanning

---

## 📱 Touch Support (Future)

Currently optimized for desktop with mouse. Mobile/tablet support planned for future updates.

---

## 🆘 Need Help?

1. **Check Console:** Open browser DevTools (F12) for error messages
2. **Review Main Report:** See `DRAWING_TOOLS_FINAL_STATUS_REPORT.md` for technical details
3. **Test Simple First:** Try horizontal line before complex tools
4. **Check Backend:** Ensure trading backend is running

---

**Last Updated:** 2026-02-13
**Status:** ✅ All Features Working
