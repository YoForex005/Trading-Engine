# Drawing Tools Quick Verification Checklist ✅

**Quick 5-Minute Smoke Test**

---

## Prerequisites
- [ ] Desktop client running at http://localhost:5173
- [ ] Chart loaded with price data
- [ ] Browser console open (F12)

---

## Core Drawing Tests (Critical Path)

### 1. Horizontal Line ⏤
- [ ] Click H-Line button (Minus icon)
- [ ] Click on chart
- [ ] **✅ PASS**: Blue horizontal line appears spanning full width
- [ ] **❌ FAIL**: No line appears / Wrong position

---

### 2. Trendline ⟋
- [ ] Click Trendline button (diagonal line)
- [ ] Click point 1
- [ ] Click point 2
- [ ] **✅ PASS**: Line connects both points
- [ ] **❌ FAIL**: No line / Wrong angle

---

### 3. Rectangle ▭
- [ ] Click Rectangle button
- [ ] Click corner 1
- [ ] Click corner 2
- [ ] **✅ PASS**: Rectangle appears between corners
- [ ] **❌ FAIL**: No rectangle / Wrong size

---

### 4. Fibonacci Levels
- [ ] Click Fibonacci button
- [ ] Click high point
- [ ] Click low point
- [ ] **✅ PASS**: 6 levels appear (0%, 23.6%, 38.2%, 50%, 61.8%, 100%)
- [ ] **❌ FAIL**: Missing levels / No labels

---

### 5. Selection & Handles
- [ ] Click Cursor button
- [ ] Click on a previously drawn line
- [ ] **✅ PASS**: Red square handles appear at endpoints
- [ ] **❌ FAIL**: No handles / Wrong color

---

### 6. Drag to Move
- [ ] Select a drawing (handles visible)
- [ ] Drag the drawing to new location
- [ ] **✅ PASS**: Drawing moves smoothly
- [ ] **❌ FAIL**: Drawing doesn't move

---

### 7. Delete
- [ ] Select a drawing
- [ ] Press Delete key
- [ ] **✅ PASS**: Drawing disappears
- [ ] **❌ FAIL**: Drawing remains

---

### 8. Persistence
- [ ] Draw 2-3 drawings
- [ ] Refresh page (F5)
- [ ] **✅ PASS**: All drawings reappear
- [ ] **❌ FAIL**: Drawings lost

---

## Extended Tests (Optional)

### Text Annotation
- [ ] Click Text button
- [ ] Click chart → Enter text → OK
- [ ] Text appears with background box

### Vertical Line
- [ ] Click V-Line button
- [ ] Click chart
- [ ] Vertical line from top to bottom

### Ellipse
- [ ] Click Ellipse button
- [ ] Click 2 corners
- [ ] Circular/oval shape appears

### Arrow
- [ ] Click Arrow button
- [ ] Click chart
- [ ] Arrow symbol appears

### Shapes Dropdown
- [ ] Click Shapes dropdown
- [ ] Select "Thumbs Up"
- [ ] Click chart → 👍 appears

### Context Menu
- [ ] Right-click on drawing
- [ ] Menu appears with Delete option
- [ ] Delete works

### Undo (Ctrl+Z)
- [ ] Delete a drawing
- [ ] Press Ctrl+Z
- [ ] Drawing restored

### Channel
- [ ] Click Channel button
- [ ] Click 2 points
- [ ] Parallel lines appear

### Pitchfork
- [ ] Click Pitchfork button
- [ ] Click 3 points
- [ ] 3 lines from center point

### Drawing List Panel
- [ ] Click Drawing List button (Layers)
- [ ] Panel shows all drawings
- [ ] Toggle visibility works

---

## Common Issues & Quick Fixes

| Issue | Likely Cause | Quick Fix |
|-------|--------------|-----------|
| **Button doesn't highlight** | Command bus not connected | Check console for errors |
| **No drawing appears** | Chart/series not initialized | Verify chart loaded first |
| **Wrong position** | Coordinate conversion error | Check series.priceToCoordinate() |
| **No handles on select** | Selected state not rendering | Check renderDrawing() nodes logic |
| **Fibonacci missing labels** | SVG text not created | Check text element in SVG |
| **Drag doesn't work** | Mouse event not attached | Check handleMouseDown() |
| **Persistence fails** | LocalStorage/backend error | Check saveToBackend() logs |
| **Context menu missing** | Event listener not attached | Check contextmenu listener |

---

## Browser Console Checks

Run these in console to debug:

```javascript
// 1. Check drawing manager
console.log(drawingManager);

// 2. Check active drawing
console.log(drawingManager.getActiveDrawing());

// 3. Check all drawings
console.log(drawingManager.getDrawings());

// 4. Check chart reference
console.log(drawingManager.chart);

// 5. Monitor events
window.addEventListener('drawing:saved', e => console.log('Saved:', e.detail));
window.addEventListener('drawing:selected', e => console.log('Selected:', e.detail));
window.addEventListener('drawing:deleted', e => console.log('Deleted:', e.detail));
```

---

## Expected File Structure

The drawing tools implementation spans these files:

```
clients/desktop/src/
├── components/
│   ├── TradingChart.tsx           (Main chart component)
│   ├── DrawingContextMenu.tsx     (Right-click menu)
│   ├── DrawingListPanel.tsx       (Drawing list sidebar)
│   └── DrawingPropertiesPanel.tsx (Properties editor)
├── services/
│   └── drawingManager.ts          (Core drawing logic)
└── hooks/
    └── useToolbarState.ts         (Toolbar state management)
```

---

## Quick Visual Reference

### What Red Handles Should Look Like:
```
    ●━━━━━━━━━━━━━━━━━━━●
    ↑                     ↑
 7×7px red square    7×7px red square
 with white border   with white border
```

### Fibonacci Levels Layout:
```
High ─────────────────── 100.0% (solid)
     ─ ─ ─ ─ ─ ─ ─ ─ ─  61.8%  (dashed)
     ─ ─ ─ ─ ─ ─ ─ ─ ─  50.0%  (dashed)
     ─ ─ ─ ─ ─ ─ ─ ─ ─  38.2%  (dashed)
     ─ ─ ─ ─ ─ ─ ─ ─ ─  23.6%  (dashed)
Low  ─────────────────── 0.0%   (solid)
```

---

## Automated Test Runner

Open this in browser for automated verification:
```
file:///C:/Users/Yofor/Desktop/Trading-Engine/tests/automated-drawing-test.html
```

---

## Pass/Fail Criteria

### ✅ ALL TESTS PASS = Production Ready
- All 8 core tests pass
- No console errors during drawing
- Drawings persist after reload
- Selection/editing works smoothly

### ⚠️ PARTIAL PASS = Needs Fixes
- 6-7 core tests pass
- Minor visual issues (colors, sizes)
- Some tools work, others need fixes

### ❌ FAIL = Not Ready
- Less than 5 core tests pass
- Critical failures (drawings don't appear)
- Console shows errors
- Crashes or freezes

---

## Sign-Off

After completing this checklist:

**Date**: _______________
**Tester**: _______________
**Result**: ⬜ PASS / ⬜ PARTIAL / ⬜ FAIL
**Production Ready**: ⬜ YES / ⬜ NO

**Notes**:
_______________________________________________________
_______________________________________________________
_______________________________________________________
