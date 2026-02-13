# Drawing Tools - Test Verification Script

## 🧪 Manual Testing Checklist

Use this checklist to verify all drawing tools are working correctly.

---

## Prerequisites

1. ✅ Backend server running (Go server on port 8080)
2. ✅ Desktop client running (npm run dev)
3. ✅ Chart loaded with market data
4. ✅ Browser DevTools console open (F12)

---

## Test 1: Rectangle Tool

### Steps
1. Click **Rectangle** button in toolbar
2. Click chart at any point (top-left corner)
3. Click chart at another point (bottom-right corner)

### Expected Results
- ✅ Rectangle appears with solid border
- ✅ Rectangle is selected (red squares on corners)
- ✅ Properties panel opens on right
- ✅ Console shows: `drawing:saved` event
- ✅ Drawing appears in list panel

### Advanced Tests
1. **Change Line Style:** Click "Dashed" in properties panel
   - ✅ Border changes to dashed line
2. **Change Color:** Click red color swatch
   - ✅ Border changes to red
3. **Drag Rectangle:** Click and drag rectangle body
   - ✅ Entire rectangle moves
4. **Drag Corner:** Click and drag red square on corner
   - ✅ Rectangle resizes
5. **Delete:** Click Delete button in properties panel
   - ✅ Rectangle disappears
   - ✅ Console shows: `drawing:deleted` event

---

## Test 2: Ellipse Tool

### Steps
1. Click **Ellipse** button in toolbar
2. Click two opposite corners

### Expected Results
- ✅ Perfect ellipse appears
- ✅ Border has 50% border-radius
- ✅ Selected with red squares
- ✅ Properties panel opens

### Advanced Tests
1. **Dotted Border:** Select "Dotted" line style
   - ✅ Border becomes dotted
2. **Line Width:** Drag slider to 5px
   - ✅ Border becomes thicker
3. **Visibility:** Click eye icon in list panel
   - ✅ Ellipse hides
   - ✅ Click eye icon again
   - ✅ Ellipse reappears

---

## Test 3: Pitchfork Tool

### Steps
1. Click **Pitchfork** button in toolbar
2. Click for center handle point
3. Click for upper point
4. Click for lower point

### Expected Results
- ✅ Three lines appear:
  - Upper line (solid)
  - Median line (dashed, between upper and lower)
  - Lower line (solid)
- ✅ All three lines start from center point
- ✅ Selected with red squares on all 3 points
- ✅ Properties panel opens

### Advanced Tests
1. **Drag Center Point:** Drag the first (center) red square
   - ✅ All three lines pivot from new center
2. **Drag Upper Point:** Drag the second red square
   - ✅ Upper line and median adjust
3. **Drag Lower Point:** Drag the third red square
   - ✅ Lower line and median adjust

---

## Test 4: Enhanced Fibonacci

### Steps
1. Click **Fibonacci** button in toolbar
2. Click start point (e.g., swing low)
3. Click end point (e.g., swing high)

### Expected Results
- ✅ **6 horizontal levels appear:**
  - 0.0% (solid line)
  - 23.6% (dashed line)
  - 38.2% (dashed line)
  - 50.0% (dashed line)
  - 61.8% (dashed line)
  - 100.0% (solid line)
- ✅ **Labels visible on right side** showing percentages
- ✅ **0% and 100% are solid**, others are dashed
- ✅ Selected with red squares on start/end points
- ✅ Properties panel opens

### Advanced Tests
1. **Drag Start Point:** Drag first red square
   - ✅ All levels recalculate
   - ✅ Spacing adjusts proportionally
2. **Drag End Point:** Drag second red square
   - ✅ All levels recalculate
3. **Change Color:** Set color to yellow
   - ✅ All 6 lines and labels change to yellow

---

## Test 5: Horizontal Line

### Steps
1. Click **Horizontal Line** button
2. Click once on chart

### Expected Results
- ✅ Horizontal line spans full chart width
- ✅ Line is at exact price clicked
- ✅ Selected with one red square
- ✅ Properties panel opens

### Advanced Tests
1. **Drag Line:** Click and drag line vertically
   - ✅ Line moves to new price level
2. **Dashed Style:** Select "Dashed"
   - ✅ Line becomes dashed
3. **Width:** Set to 4px
   - ✅ Line becomes thicker

---

## Test 6: Vertical Line

### Steps
1. Click **Vertical Line** button
2. Click once on chart

### Expected Results
- ✅ Vertical line spans full chart height
- ✅ Line is at exact time clicked
- ✅ Selected with one red square

### Advanced Tests
1. **Drag Line:** Click and drag line horizontally
   - ✅ Line moves to new time position

---

## Test 7: Trendline

### Steps
1. Click **Trendline** button
2. Click start point
3. Click end point

### Expected Results
- ✅ Diagonal line connects two points
- ✅ Selected with two red squares
- ✅ Line rotates to match angle

### Advanced Tests
1. **Extend Right:** Check "Extend Right" in properties panel
   - ⚠️ (Enhancement needed - not yet implemented)
2. **Drag Points:** Drag either red square
   - ✅ Line pivots/extends accordingly

---

## Test 8: Text Annotation

### Steps
1. Click **Text** button
2. Click on chart
3. Type "Test Annotation" in prompt
4. Click OK

### Expected Results
- ✅ Text appears at clicked location
- ✅ Text has dark background with border
- ✅ Selected with one red square
- ✅ Properties panel opens with text input

### Advanced Tests
1. **Edit Text:** Type new text in properties panel
   - ✅ Text updates on chart
2. **Drag Text:** Click and drag text box
   - ✅ Text moves to new location
3. **Change Color:** Set to green
   - ✅ Text color changes to green

---

## Test 9: Shapes

### Steps
1. Click **Shapes** dropdown button
2. Select a shape (e.g., "Thumbs Up 👍")
3. Click on chart

### Expected Results
- ✅ Selected emoji/symbol appears
- ✅ Positioned at clicked location
- ✅ Selected with one red square

### Advanced Tests
1. Test all 9 shape types:
   - ✅ 👍 Thumbs Up
   - ✅ 👎 Thumbs Down
   - ✅ ↑ Arrow Up
   - ✅ ↓ Arrow Down
   - ✅ 🛑 Stop
   - ✅ ✅ Check
   - ✅ ▲ Buy
   - ✅ ▼ Sell
   - ✅ ➔ Arrow Right

---

## Test 10: Arrow

### Steps
1. Click **Arrow** dropdown button
2. Select arrow direction (up/down/right)
3. Click on chart

### Expected Results
- ✅ Arrow symbol appears (⬆ / ⬇ / ➜)
- ✅ Positioned at clicked location
- ✅ Selected with one red square

---

## Test 11: Channel

### Steps
1. Click **Channel** button
2. Click two points

### Expected Results
- ✅ Rectangular channel appears
- ✅ Parallel lines visible
- ✅ Selected with two red squares

---

## Test 12: Multi-Drawing Management

### Steps
1. Create 3 drawings (any types)
2. Click **Drawing List** icon

### Expected Results
- ✅ List panel shows all 3 drawings
- ✅ Each has icon with color preview
- ✅ Click item to select on chart
- ✅ Eye icon toggles visibility
- ✅ Trash icon deletes drawing

### Advanced Tests
1. **Hide Drawing:** Click eye icon on one
   - ✅ Drawing hides on chart
   - ✅ List item becomes semi-transparent
2. **Show Drawing:** Click eye icon again
   - ✅ Drawing reappears
3. **Delete from List:** Click trash icon
   - ✅ Confirm dialog appears
   - ✅ Drawing removed from chart and list
4. **Clear All:** Click "Clear All" at bottom
   - ✅ Confirm dialog appears
   - ✅ All drawings removed

---

## Test 13: Selection & Multi-Select

### Steps
1. Create 3 drawings
2. Click on drawing #1

### Expected Results
- ✅ Drawing #1 selected (red squares visible)
- ✅ Other drawings deselected

### Multi-Select Test
1. Hold Ctrl, click drawing #2
   - ✅ Both #1 and #2 selected
2. Hold Ctrl, click drawing #3
   - ✅ All three selected
3. Press Delete key
   - ✅ All three deleted

---

## Test 14: Duplicate Function

### Steps
1. Create one rectangle with custom color (red)
2. Select the rectangle
3. Click "Duplicate" in properties panel

### Expected Results
- ✅ New rectangle appears offset from original
- ✅ New rectangle has same color (red)
- ✅ New rectangle has same line style
- ✅ Original remains selected
- ✅ Console shows: `drawing:duplicated` event

---

## Test 15: Persistence (Backend Save)

### Steps
1. Create 2 drawings (any types)
2. **Reload the page** (F5)
3. Wait for chart to load

### Expected Results
- ✅ Both drawings reappear
- ✅ Same positions
- ✅ Same colors and styles
- ✅ Visible in drawing list

### LocalStorage Fallback Test
1. Stop backend server
2. Create a drawing
3. Reload page
   - ✅ Drawing still appears (from localStorage)

---

## Test 16: Drag & Drop Performance

### Steps
1. Create 5 different drawings
2. Select one drawing
3. Drag it rapidly around the chart

### Expected Results
- ✅ Drawing follows cursor smoothly
- ✅ Red squares move with drawing
- ✅ No lag or stuttering
- ✅ Position updates in real-time
- ✅ On release, saves to backend

---

## Test 17: Context Menu

### Steps
1. Create one drawing
2. **Right-click** on the drawing

### Expected Results
- ✅ Context menu appears at cursor
- ✅ Menu options visible:
  - Properties
  - Duplicate
  - Delete
  - Lock/Unlock

### Test Menu Actions
1. Click "Duplicate"
   - ✅ New drawing created
2. Right-click again, click "Delete"
   - ✅ Drawing removed

---

## Test 18: Edge Cases

### Empty State
1. Delete all drawings
2. Open drawing list panel
   - ✅ Shows "No drawings yet" message
   - ✅ Helpful icon displayed

### Rapid Clicking
1. Click Rectangle tool
2. Click chart rapidly 10 times
   - ✅ Only creates one rectangle (auto-completes after 2 points)

### Cancel Drawing
1. Click Fibonacci tool
2. Click once (only 1 point)
3. Press **Esc** key
   - ✅ Active drawing cancelled
   - ✅ No drawing created

### Zoom/Scroll
1. Create one drawing
2. Zoom in/out (mouse wheel)
   - ✅ Drawing scales with chart
3. Scroll horizontally
   - ✅ Drawing moves with chart
4. Scroll vertically (price scale)
   - ✅ Drawing moves with price

---

## Test 19: Browser Console (No Errors)

### Steps
1. Open DevTools Console (F12)
2. Create 3 drawings
3. Edit 2 drawings
4. Delete 1 drawing

### Expected Results
- ✅ No red error messages
- ✅ Only info logs:
  - `drawing:saved`
  - `drawing:deleted`
  - `drawing:selected`
- ✅ No TypeScript errors
- ✅ No network errors (check Network tab)

---

## Test 20: Network Persistence

### Steps
1. Open DevTools Network tab (F12)
2. Create one drawing
3. Check network requests

### Expected Results
- ✅ POST request to `/api/workspace/drawings`
- ✅ Status: 200 OK
- ✅ Response contains saved drawing with ID
- ✅ Request payload includes:
  - Drawing type
  - Points (time/price)
  - Color, lineWidth, lineStyle
  - Symbol
  - AccountId

### Modify Test
1. Drag the drawing to new position
2. Check network requests
   - ✅ Another POST request sent
   - ✅ Updated coordinates in payload

---

## 🎯 Success Criteria

### All Tests Must Pass
- [ ] Test 1: Rectangle ✅
- [ ] Test 2: Ellipse ✅
- [ ] Test 3: Pitchfork ✅
- [ ] Test 4: Fibonacci ✅
- [ ] Test 5: Horizontal Line ✅
- [ ] Test 6: Vertical Line ✅
- [ ] Test 7: Trendline ✅
- [ ] Test 8: Text ✅
- [ ] Test 9: Shapes ✅
- [ ] Test 10: Arrow ✅
- [ ] Test 11: Channel ✅
- [ ] Test 12: Multi-Drawing ✅
- [ ] Test 13: Selection ✅
- [ ] Test 14: Duplicate ✅
- [ ] Test 15: Persistence ✅
- [ ] Test 16: Drag Performance ✅
- [ ] Test 17: Context Menu ✅
- [ ] Test 18: Edge Cases ✅
- [ ] Test 19: No Console Errors ✅
- [ ] Test 20: Network Requests ✅

---

## 📊 Test Results Template

```
Date: _____________
Tester: _____________
Environment: _____________

Results:
✅ Passed: ___ / 20
⚠️  Warnings: ___
❌ Failed: ___

Notes:
_______________________________________
_______________________________________
_______________________________________

Signature: _____________
```

---

## 🚨 If Tests Fail

### Debugging Steps
1. **Check browser console** for error messages
2. **Verify backend is running** (curl http://localhost:8080/health)
3. **Clear browser cache** and reload
4. **Check network tab** for failed API requests
5. **Verify chart is loaded** with market data
6. **Check localStorage** (Application tab in DevTools)

### Common Issues

**Drawing Not Appearing**
- Ensure enough points clicked for tool type
- Check if hidden (eye icon in list)
- Zoom out to see if off-screen

**Properties Panel Not Opening**
- Drawing must be selected first
- Click directly on drawing

**Persistence Not Working**
- Backend server must be running
- Check CORS settings
- Verify API endpoint configuration

**Drag Not Working**
- Drawing must be selected
- Click on drawing body, not background
- Check if drawing is locked

---

## 📝 Report Issues

If any test fails consistently:

1. **Document:**
   - Which test failed
   - Steps to reproduce
   - Browser console errors
   - Network tab errors

2. **Check Files:**
   - `drawingManager.ts` - Core logic
   - `TradingChart.tsx` - Integration
   - `DrawingPropertiesPanel.tsx` - UI panel
   - `DrawingListPanel.tsx` - List UI

3. **Verify:**
   - TypeScript compilation successful
   - No runtime JavaScript errors
   - Backend API accessible

---

**Test Script Version:** 1.0
**Last Updated:** 2026-02-13
**Expected Result:** ✅ ALL TESTS PASS
