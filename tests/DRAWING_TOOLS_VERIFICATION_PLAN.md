# Drawing Tools Verification Plan
**Test Execution Date**: 2026-02-13
**Version**: 1.0
**Status**: Ready for Execution

---

## 📋 Overview

This document provides a step-by-step manual testing plan to verify that all drawing tools in the desktop client are fully functional. Each test includes detailed steps, expected results, and space to record actual results.

---

## 🎯 Pre-Test Setup

### Environment Check
- [ ] Backend server is running on port 8080
- [ ] Desktop client is running on port 5173
- [ ] Browser console is open (F12) for error monitoring
- [ ] Test symbol: XAUUSD (or any symbol with visible chart data)
- [ ] Chart is loaded and displaying candles

### Initial State Verification
1. Open browser to http://localhost:5173
2. Verify chart is visible with price data
3. Open browser DevTools (F12) → Console tab
4. Confirm no critical errors in console
5. Verify toolbar is visible at top of screen

---

## 🧪 Test Suite

---

### **TEST 1: Horizontal Line (H-Line)**

#### **Objective**: Verify horizontal line can be drawn at any price level

#### **Steps**:
1. Click the **Horizontal Line** button in toolbar (Minus icon `-`)
   - Button should highlight in blue when active
2. Observe cursor changes to crosshair
3. Click anywhere on the chart (middle of visible area)
4. Observe result

#### **Expected Results**:
✅ Button highlights in blue when clicked
✅ Cursor changes to crosshair
✅ Horizontal line appears spanning full chart width at clicked price level
✅ Line is blue color (#3b82f6)
✅ Line is 2px thick
✅ Line extends from left edge to right edge of chart

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Button highlight | ⬜ | |
| Cursor change | ⬜ | |
| Line appears | ⬜ | |
| Line color correct | ⬜ | |
| Line spans full width | ⬜ | |
| Line at clicked price | ⬜ | |

#### **Failure Troubleshooting**:
- If line doesn't appear: Check console for errors, verify drawingManager is initialized
- If line wrong position: Check series.coordinateToPrice() conversion
- If cursor doesn't change: Check activeDrawingType state in TradingChart.tsx

---

### **TEST 2: Trendline**

#### **Objective**: Verify trendline connects two points clicked by user

#### **Steps**:
1. Click **Trendline** button (diagonal line icon)
   - Button should highlight in blue
2. Click first point on chart (e.g., low of a candle)
3. Click second point on chart (e.g., low of another candle)
4. Observe result

#### **Expected Results**:
✅ Button highlights in blue
✅ Cursor changes to crosshair
✅ After first click: No line yet (waiting for 2nd point)
✅ After second click: Line appears connecting both points
✅ Line is blue (#3b82f6)
✅ Line rotates to match angle between points
✅ Red square handles appear at both endpoints (selected state)

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Button highlight | ⬜ | |
| Cursor change | ⬜ | |
| First click (no line) | ⬜ | |
| Second click (line appears) | ⬜ | |
| Line connects points | ⬜ | |
| Line color/thickness | ⬜ | |
| Red handles visible | ⬜ | |

#### **Failure Troubleshooting**:
- If line doesn't appear after 2nd click: Check isDrawingComplete() logic
- If line wrong angle: Check transform rotation calculation
- If no handles: Check selected state rendering in renderDrawing()

---

### **TEST 3: Rectangle**

#### **Objective**: Verify rectangle can be drawn between two diagonal corners

#### **Steps**:
1. Click **Rectangle** button (Square icon)
2. Click first corner (top-left area of chart)
3. Click opposite corner (bottom-right area of chart)
4. Observe result

#### **Expected Results**:
✅ Button highlights in blue
✅ After first click: No rectangle yet
✅ After second click: Rectangle appears
✅ Rectangle spans from first corner to second corner
✅ Rectangle has blue border (#3b82f6)
✅ Rectangle interior is transparent
✅ Red square handles at both corners

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Button highlight | ⬜ | |
| First click (no rect) | ⬜ | |
| Second click (rect appears) | ⬜ | |
| Correct size/position | ⬜ | |
| Border color correct | ⬜ | |
| Transparent fill | ⬜ | |
| Handles visible | ⬜ | |

---

### **TEST 4: Fibonacci Retracement**

#### **Objective**: Verify Fibonacci levels appear with correct ratios and labels

#### **Steps**:
1. Click **Fibonacci** button (horizontal lines icon)
2. Click high point on chart (peak of price movement)
3. Click low point on chart (trough of price movement)
4. Observe result

#### **Expected Results**:
✅ Button highlights in blue
✅ After first click: No levels yet
✅ After second click: 6 horizontal lines appear
✅ Lines at ratios: 0%, 23.6%, 38.2%, 50%, 61.8%, 100%
✅ Each line has label on right side showing percentage
✅ 0% and 100% lines are solid, others are dashed
✅ Lines span horizontally between the two click points
✅ Red handles at high and low points

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Button highlight | ⬜ | |
| 6 levels appear | ⬜ | |
| Correct ratios (0-100%) | ⬜ | |
| Labels visible | ⬜ | |
| 0%/100% solid | ⬜ | |
| Others dashed | ⬜ | |
| Handles at endpoints | ⬜ | |

#### **Fibonacci Level Verification**:
Count and verify each level:
- [ ] 0.0% (solid line)
- [ ] 23.6% (dashed)
- [ ] 38.2% (dashed)
- [ ] 50.0% (dashed)
- [ ] 61.8% (dashed)
- [ ] 100.0% (solid line)

---

### **TEST 5: Ellipse**

#### **Objective**: Verify ellipse can be drawn

#### **Steps**:
1. Click **Ellipse** button (Circle icon)
2. Click first corner point
3. Click opposite corner point
4. Observe result

#### **Expected Results**:
✅ Button highlights
✅ Ellipse appears after 2nd click
✅ Ellipse fits within bounding box of two points
✅ Border is blue, circular/oval shape
✅ Transparent fill
✅ Red handles at corners

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Button highlight | ⬜ | |
| Ellipse appears | ⬜ | |
| Correct shape | ⬜ | |
| Border/fill correct | ⬜ | |
| Handles visible | ⬜ | |

---

### **TEST 6: Text Annotation**

#### **Objective**: Verify text can be added to chart

#### **Steps**:
1. Click **Text** button (Type icon `T`)
2. Click location on chart
3. When prompt appears, type: "Test Annotation"
4. Click OK
5. Observe result

#### **Expected Results**:
✅ Button highlights
✅ Prompt dialog appears after click
✅ Text "Test Annotation" appears at clicked location
✅ Text is blue color
✅ Text has dark background box
✅ Text is readable and positioned correctly

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Button highlight | ⬜ | |
| Prompt appears | ⬜ | |
| Text displays | ⬜ | |
| Correct position | ⬜ | |
| Readable styling | ⬜ | |

---

### **TEST 7: Vertical Line**

#### **Objective**: Verify vertical line appears at clicked time

#### **Steps**:
1. Click **Vertical Line** button (thin vertical bar)
2. Click on chart at specific time point
3. Observe result

#### **Expected Results**:
✅ Button highlights
✅ Vertical line appears from top to bottom of chart
✅ Line positioned at clicked time coordinate
✅ Line is blue and 2px wide

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Button highlight | ⬜ | |
| Line appears | ⬜ | |
| Full height | ⬜ | |
| Correct position | ⬜ | |

---

### **TEST 8: Arrow Tool**

#### **Objective**: Verify arrow annotation appears

#### **Steps**:
1. Click **Arrow** button (ArrowUpRight icon)
2. Click location on chart
3. Observe result

#### **Expected Results**:
✅ Button highlights
✅ Arrow symbol appears at clicked location
✅ Arrow is visible and blue colored
✅ Single click creates arrow (1-point tool)

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Button highlight | ⬜ | |
| Arrow appears | ⬜ | |
| Correct styling | ⬜ | |

---

### **TEST 9: Selection & Editing**

#### **Objective**: Verify drawings can be selected and display edit handles

#### **Steps**:
1. First, draw a trendline (see Test 2)
2. Click **Cursor** button (MousePointer2 icon) to exit drawing mode
3. Click on the previously drawn trendline
4. Observe selection state

#### **Expected Results**:
✅ Cursor button highlights when clicked
✅ Drawing mode exits (cursor back to normal)
✅ When trendline is clicked, it becomes selected
✅ **Red square handles** appear at both endpoints
✅ Handles are 7px × 7px squares
✅ Handles are red (#ff0000) with white border

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Cursor button works | ⬜ | |
| Exit drawing mode | ⬜ | |
| Selection works | ⬜ | |
| Red handles appear | ⬜ | |
| Handle size/color | ⬜ | |

#### **Visual Check**:
Handles should look like this:
```
┌─┐
│■│  ← 7px red square with 1px white border
└─┘
```

---

### **TEST 10: Drag to Move Drawing**

#### **Objective**: Verify entire drawing can be repositioned

#### **Steps**:
1. Select a drawing (see Test 9)
2. Click and hold on the drawing line itself (not handles)
3. Drag to a new position
4. Release mouse
5. Observe result

#### **Expected Results**:
✅ Drawing follows mouse during drag
✅ Drawing updates position smoothly
✅ Drawing snaps to new position on release
✅ Drawing maintains its shape/size
✅ Drawing is automatically saved to backend/storage

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Drag initiates | ⬜ | |
| Smooth movement | ⬜ | |
| Position updates | ⬜ | |
| Shape preserved | ⬜ | |
| Auto-saved | ⬜ | |

---

### **TEST 11: Drag Handle to Resize**

#### **Objective**: Verify drawing can be resized by dragging endpoint handles

#### **Steps**:
1. Select a trendline (red handles visible)
2. Click and hold on one of the red square handles
3. Drag handle to a new position
4. Release mouse
5. Observe result

#### **Expected Results**:
✅ Handle follows mouse during drag
✅ Drawing stretches/resizes in real-time
✅ Other endpoint remains fixed
✅ Drawing updates smoothly
✅ New size persists after release

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Handle drags | ⬜ | |
| Real-time resize | ⬜ | |
| Other point fixed | ⬜ | |
| Smooth update | ⬜ | |

---

### **TEST 12: Right-Click Context Menu**

#### **Objective**: Verify right-click menu appears with delete option

#### **Steps**:
1. Select a drawing
2. Right-click directly on the drawing
3. Observe context menu
4. Click "Delete" option
5. Observe result

#### **Expected Results**:
✅ Context menu appears at cursor position
✅ Menu shows drawing options (Delete, Properties, etc.)
✅ Clicking Delete removes the drawing
✅ Drawing disappears from chart
✅ Deletion is saved to backend/storage

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Context menu appears | ⬜ | |
| Delete option visible | ⬜ | |
| Click delete works | ⬜ | |
| Drawing removed | ⬜ | |

---

### **TEST 13: Delete via Keyboard**

#### **Objective**: Verify Delete key removes selected drawing

#### **Steps**:
1. Select a drawing (click on it)
2. Press **Delete** key on keyboard
3. Observe result

#### **Expected Results**:
✅ Drawing is immediately removed
✅ No confirmation dialog
✅ Drawing disappears from chart
✅ Other drawings remain unaffected

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Delete key works | ⬜ | |
| Drawing removed | ⬜ | |
| Others unaffected | ⬜ | |

---

### **TEST 14: Undo Delete**

#### **Objective**: Verify Ctrl+Z can restore deleted drawings

#### **Steps**:
1. Draw a horizontal line
2. Delete it (via Delete key or context menu)
3. Press **Ctrl+Z**
4. Observe result

#### **Expected Results**:
✅ Deleted drawing reappears at original position
✅ Drawing retains original properties (color, size, etc.)
✅ Drawing is restored to drawings array

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Ctrl+Z works | ⬜ | |
| Drawing restored | ⬜ | |
| Original position | ⬜ | |

---

### **TEST 15: Channel Tool**

#### **Objective**: Verify equidistant channel can be drawn

#### **Steps**:
1. Click **Channel** button (parallel lines icon)
2. Click first point
3. Click second point
4. Observe result

#### **Expected Results**:
✅ Button highlights
✅ Channel appears as rectangular border
✅ Two parallel lines visible
✅ Blue color (#3b82f6)

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Button highlight | ⬜ | |
| Channel appears | ⬜ | |
| Parallel lines | ⬜ | |
| Correct styling | ⬜ | |

---

### **TEST 16: Pitchfork Tool**

#### **Objective**: Verify Andrew's Pitchfork can be drawn (3 points)

#### **Steps**:
1. Click **Pitchfork** button
2. Click first point (center/pivot)
3. Click second point (upper)
4. Click third point (lower)
5. Observe result

#### **Expected Results**:
✅ After 3 clicks, 3 lines appear from center point
✅ Lines form pitchfork shape
✅ Median line is dashed
✅ Upper and lower lines are solid

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| 3 clicks required | ⬜ | |
| 3 lines appear | ⬜ | |
| Pitchfork shape | ⬜ | |
| Median dashed | ⬜ | |

---

### **TEST 17: Shapes Dropdown**

#### **Objective**: Verify shapes dropdown shows symbol options

#### **Steps**:
1. Click **Shapes** button (dropdown with chevron)
2. Observe dropdown menu
3. Click "Thumbs Up"
4. Click on chart
5. Observe result

#### **Expected Results**:
✅ Dropdown opens showing shape options
✅ Options include: Thumbs Up, Thumbs Down, Arrows, Stop, Check, etc.
✅ Clicking "Thumbs Up" activates that shape
✅ Clicking chart places 👍 symbol

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Dropdown opens | ⬜ | |
| Options visible | ⬜ | |
| Selection works | ⬜ | |
| Symbol placed | ⬜ | |

---

### **TEST 18: Persistence (Page Reload)**

#### **Objective**: Verify drawings persist after page refresh

#### **Steps**:
1. Draw 3 different types of drawings (e.g., hline, trendline, rectangle)
2. Wait 2 seconds (for auto-save)
3. Refresh page (F5)
4. Wait for chart to reload
5. Observe result

#### **Expected Results**:
✅ All 3 drawings reappear after reload
✅ Drawings are at original positions
✅ Drawings have original colors/styles
✅ No console errors during load

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Drawings persist | ⬜ | |
| Correct positions | ⬜ | |
| Correct styling | ⬜ | |
| No errors | ⬜ | |

---

### **TEST 19: Multiple Drawings**

#### **Objective**: Verify multiple drawings can coexist

#### **Steps**:
1. Draw a horizontal line
2. Draw a trendline
3. Draw a rectangle
4. Draw a Fibonacci retracement
5. Draw a text annotation
6. Observe all 5 on chart

#### **Expected Results**:
✅ All 5 drawings visible simultaneously
✅ No overlapping/z-index issues
✅ Each can be selected independently
✅ No performance degradation

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| 5 drawings visible | ⬜ | |
| No overlap issues | ⬜ | |
| Individual selection | ⬜ | |
| Smooth performance | ⬜ | |

---

### **TEST 20: Drawing List Panel**

#### **Objective**: Verify drawing list panel shows/hides drawings

#### **Steps**:
1. Draw 2-3 different drawings
2. Click **Drawing List** button (Layers icon in toolbar)
3. Observe panel appears
4. Toggle visibility of one drawing
5. Observe result

#### **Expected Results**:
✅ Drawing List panel appears on right side
✅ All drawings listed by type
✅ Eye icon allows show/hide toggle
✅ Hidden drawing disappears from chart
✅ Hidden drawing still in list (grayed out)

#### **Actual Results**:
| Checkpoint | Pass/Fail | Notes |
|------------|-----------|-------|
| Panel opens | ⬜ | |
| Drawings listed | ⬜ | |
| Toggle works | ⬜ | |
| Hide/show functional | ⬜ | |

---

## 📊 Test Summary

### Results Table
Fill this in after completing all tests:

| Test # | Test Name | Status | Issues Found |
|--------|-----------|--------|--------------|
| 1 | Horizontal Line | ⬜ Pass / ⬜ Fail | |
| 2 | Trendline | ⬜ Pass / ⬜ Fail | |
| 3 | Rectangle | ⬜ Pass / ⬜ Fail | |
| 4 | Fibonacci | ⬜ Pass / ⬜ Fail | |
| 5 | Ellipse | ⬜ Pass / ⬜ Fail | |
| 6 | Text Annotation | ⬜ Pass / ⬜ Fail | |
| 7 | Vertical Line | ⬜ Pass / ⬜ Fail | |
| 8 | Arrow Tool | ⬜ Pass / ⬜ Fail | |
| 9 | Selection & Editing | ⬜ Pass / ⬜ Fail | |
| 10 | Drag to Move | ⬜ Pass / ⬜ Fail | |
| 11 | Drag Handle to Resize | ⬜ Pass / ⬜ Fail | |
| 12 | Right-Click Menu | ⬜ Pass / ⬜ Fail | |
| 13 | Delete via Keyboard | ⬜ Pass / ⬜ Fail | |
| 14 | Undo Delete | ⬜ Pass / ⬜ Fail | |
| 15 | Channel Tool | ⬜ Pass / ⬜ Fail | |
| 16 | Pitchfork Tool | ⬜ Pass / ⬜ Fail | |
| 17 | Shapes Dropdown | ⬜ Pass / ⬜ Fail | |
| 18 | Persistence | ⬜ Pass / ⬜ Fail | |
| 19 | Multiple Drawings | ⬜ Pass / ⬜ Fail | |
| 20 | Drawing List Panel | ⬜ Pass / ⬜ Fail | |

### Overall Statistics
- **Total Tests**: 20
- **Tests Passed**: ___
- **Tests Failed**: ___
- **Pass Rate**: ___%

---

## 🐛 Known Issues & Fixes

### Common Failure Patterns

#### Issue 1: Drawings Don't Appear
**Symptoms**: Clicking toolbar button but no drawing appears
**Likely Cause**: Chart/series not initialized, or drawing overlay container missing
**Fix**:
```typescript
// Check if overlay container exists in TradingChart.tsx
<div className="chart-drawing-overlay absolute inset-0 pointer-events-none" />
```

#### Issue 2: Handles Don't Show on Selection
**Symptoms**: Drawing is selected but no red squares appear
**Likely Cause**: Selected state not rendering nodes
**Fix**: Check `renderDrawing()` method around line 462-494 in drawingManager.ts

#### Issue 3: Fibonacci Levels Missing Labels
**Symptoms**: Lines appear but no percentage labels
**Likely Cause**: SVG text elements not rendering
**Fix**: Verify text element creation in lines 620-627 of drawingManager.ts

#### Issue 4: Context Menu Doesn't Appear
**Symptoms**: Right-click doesn't show menu
**Likely Cause**: Event listener not attached or DrawingContextMenu component issue
**Fix**: Check event listener in line 500-505 of drawingManager.ts

---

## 🔍 Debug Checklist

If tests fail, check these in browser console:

1. **Check drawingManager initialization**:
```javascript
console.log(window.drawingManager || 'Not initialized');
```

2. **Check active tool state**:
```javascript
console.log(drawingManager.getActiveDrawing());
```

3. **Check drawings array**:
```javascript
console.log(drawingManager.getDrawings());
```

4. **Check chart reference**:
```javascript
console.log(drawingManager.chart); // Should not be null
```

5. **Monitor drawing events**:
```javascript
window.addEventListener('drawing:saved', (e) => console.log('Drawing saved:', e.detail));
window.addEventListener('drawing:selected', (e) => console.log('Drawing selected:', e.detail));
window.addEventListener('drawing:deleted', (e) => console.log('Drawing deleted:', e.detail));
```

---

## 📸 Screenshot Checklist

Take screenshots of these states:
- [ ] Clean chart (before drawings)
- [ ] Horizontal line drawn
- [ ] Trendline with red handles visible
- [ ] Fibonacci retracement with all 6 levels
- [ ] Rectangle drawn
- [ ] Multiple drawings on same chart
- [ ] Drawing List panel open
- [ ] Context menu showing delete option
- [ ] Any failures encountered

---

## ✅ Sign-Off

**Tester Name**: ____________________
**Date Completed**: ____________________
**Overall Result**: ⬜ PASS / ⬜ FAIL
**Ready for Production**: ⬜ YES / ⬜ NO

**Notes**:
_____________________________________________________________________
_____________________________________________________________________
_____________________________________________________________________

---

**Next Steps After Testing**:
1. Document all failures in Issues section
2. Create bug tickets for each failed test
3. Fix critical issues (P0: drawing doesn't appear)
4. Re-run failed tests after fixes
5. Store results in memory for future reference
