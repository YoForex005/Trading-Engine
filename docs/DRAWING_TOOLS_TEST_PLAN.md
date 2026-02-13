# Drawing Tools Test Plan

**Version:** 1.0
**Date:** 2026-02-13
**Platform:** Trading Engine Desktop Client
**Component:** Chart Drawing Tools

---

## Table of Contents
1. [Overview](#overview)
2. [Test Environment](#test-environment)
3. [Drawing Tools Reference](#drawing-tools-reference)
4. [Functional Tests](#functional-tests)
5. [Persistence Tests](#persistence-tests)
6. [Edge Case Tests](#edge-case-tests)
7. [UI/UX Tests](#ui-ux-tests)
8. [Performance Tests](#performance-tests)
9. [Test Results Template](#test-results-template)

---

## Overview

### Purpose
This document provides a comprehensive test plan for all 11 drawing tools in the Trading Engine platform. It covers functional testing, persistence verification, edge cases, and expected behaviors.

### Scope
- All 11 drawing tool types
- Drawing creation, modification, deletion
- Backend persistence and local storage fallback
- Cross-session and cross-symbol persistence
- Edge cases and error handling

### Drawing Tools Covered
1. Cursor (selection tool)
2. Trendline
3. Horizontal Line (hline)
4. Vertical Line (vline)
5. Channel
6. Fibonacci Retracement
7. Text Annotation
8. Rectangle
9. Ellipse
10. Arrow
11. Pitchfork

---

## Test Environment

### Prerequisites
- **Backend Status:** Running on `http://localhost:8080`
- **Frontend Status:** Running on `http://localhost:5173`
- **Browser:** Chrome/Edge (latest version)
- **Test Data:**
  - Symbol: EURUSD
  - Timeframe: M1
  - Account ID: 1

### Test Data Setup
1. Ensure backend is running and accessible
2. Clear browser localStorage: `localStorage.clear()`
3. Clear backend drawings: DELETE `/api/workspace/drawings?symbol=EURUSD&accountId=1`
4. Refresh browser
5. Open TradingChart component with EURUSD symbol

---

## Drawing Tools Reference

### Tool Characteristics

| Tool | Points Required | Cursor Style | Backend Save | Movable | Resizable |
|------|----------------|--------------|--------------|---------|-----------|
| Cursor | 0 (selection) | default | N/A | N/A | N/A |
| Trendline | 2 | crosshair | Yes | Yes | Yes |
| Horizontal Line | 1 | crosshair | Yes | Yes | No |
| Vertical Line | 1 | crosshair | Yes | Yes | No |
| Channel | 2 | crosshair | Yes | Yes | Yes |
| Fibonacci | 2 | crosshair | Yes | Yes | Yes |
| Text | 1 + prompt | crosshair | Yes | Yes | No |
| Rectangle | 2 | crosshair | Yes | Yes | Yes |
| Ellipse | 2 | crosshair | Yes | Yes | Yes |
| Arrow | 1 | crosshair | Yes | Yes | No |
| Pitchfork | 3 | crosshair | Yes | Yes | Yes |

---

## Functional Tests

### TEST-001: Cursor Tool (Selection)
**Objective:** Verify cursor tool selects and deselects drawings

**Steps:**
1. Open chart with EURUSD
2. Click "Cursor" tool in toolbar (or press Esc)
3. Click on an existing drawing on chart
4. Observe drawing selection state
5. Click on chart background (not on drawing)
6. Verify drawing is deselected

**Expected Results:**
- Cursor tool icon is highlighted in toolbar
- Chart cursor is default arrow (not crosshair)
- Clicking drawing shows red selection nodes (7x7px squares)
- Clicking background deselects all drawings
- No drawing creation occurs

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-002: Trendline Tool
**Objective:** Create and verify trendline drawing

**Steps:**
1. Click "Trendline" tool in toolbar
2. Verify cursor changes to crosshair
3. Click first point on chart (e.g., at time=1000, price=1.0850)
4. Move mouse to second position
5. Click second point (e.g., at time=2000, price=1.0860)
6. Verify trendline appears connecting both points

**Expected Results:**
- Cursor becomes crosshair immediately
- Line follows mouse between clicks (preview)
- After 2 clicks, trendline is drawn and saved
- Line color: #3b82f6 (blue)
- Line width: 2px
- Line is selectable (shows red nodes when selected)
- Console shows: `[drawingManager] Rendering drawing: {type: 'trendline', points: 2}`
- Backend POST request to `/api/workspace/drawings` succeeds

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-003: Horizontal Line (hline)
**Objective:** Create horizontal price level line

**Steps:**
1. Click "Horizontal Line" tool in toolbar
2. Cursor changes to crosshair
3. Click once on chart at desired price level (e.g., price=1.0855)
4. Verify horizontal line spans entire chart width

**Expected Results:**
- Cursor: crosshair
- After 1 click, horizontal line is drawn
- Line style: `left: 0, right: 0, height: 2px`
- Line color: #3b82f6 (blue)
- Line follows price level (vertical drag to adjust)
- Cursor on line: ns-resize (north-south)
- Backend save triggered

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-004: Vertical Line (vline)
**Objective:** Create vertical time-based line

**Steps:**
1. Click "Vertical Line" tool in toolbar
2. Cursor changes to crosshair
3. Click once on chart at desired time (e.g., time=1500)
4. Verify vertical line spans entire chart height

**Expected Results:**
- Cursor: crosshair
- After 1 click, vertical line is drawn
- Line style: `top: 0, bottom: 0, width: 2px`
- Line color: #3b82f6 (blue)
- Line follows time axis (horizontal drag to adjust)
- Cursor on line: ew-resize (east-west)
- Backend save triggered

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-005: Channel Tool
**Objective:** Create parallel channel lines

**Steps:**
1. Click "Channel" tool in toolbar
2. Cursor changes to crosshair
3. Click first point (e.g., time=1000, price=1.0850)
4. Click second point (e.g., time=2000, price=1.0860)
5. Verify channel box appears

**Expected Results:**
- Cursor: crosshair
- After 2 clicks, rectangular channel is drawn
- Border: 2px solid #3b82f6
- Transparent background
- Covers area from (min(x1,x2), min(y1,y2)) to (max(x1,x2), max(y1,y2))
- Draggable and resizable via corner nodes
- Backend save triggered

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-006: Fibonacci Retracement
**Objective:** Create Fibonacci retracement levels

**Steps:**
1. Click "Fibonacci" tool in toolbar
2. Cursor changes to crosshair
3. Click first point (e.g., swing low at time=1000, price=1.0840)
4. Click second point (e.g., swing high at time=2000, price=1.0870)
5. Verify Fibonacci levels appear

**Expected Results:**
- Cursor: crosshair
- After 2 clicks, Fibonacci levels are drawn
- Levels displayed: 0%, 23.6%, 38.2%, 50%, 61.8%, 100%
- 0% and 100% lines are solid
- Other levels are dashed (stroke-dasharray: '4 2')
- Each level shows label on right side (e.g., "23.6%")
- Line color: #3b82f6 (blue)
- Labels: 11px, text-anchor: end
- Backend save triggered

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-007: Text Annotation
**Objective:** Add text annotation to chart

**Steps:**
1. Click "Text" tool in toolbar
2. Cursor changes to crosshair
3. Click once on chart (e.g., time=1500, price=1.0855)
4. Browser prompt appears: "Enter annotation text:"
5. Type text: "Resistance Level"
6. Click OK
7. Verify text appears on chart

**Expected Results:**
- Cursor: crosshair
- After click, browser prompt appears
- Text box appears at clicked position
- Text style:
  - Color: #3b82f6 (blue)
  - Font size: 12px
  - Font weight: bold
  - Background: rgba(24, 24, 27, 0.9)
  - Border: 1px solid #3f3f46
  - Padding: 4px 8px
  - Border radius: 4px
- Text is draggable
- Backend save triggered with text content

**Cancel Test:**
- If user clicks Cancel in prompt, drawing is canceled
- No text annotation is created

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-008: Rectangle Tool
**Objective:** Draw rectangular shape

**Steps:**
1. Click "Rectangle" tool in toolbar
2. Cursor changes to crosshair
3. Click first corner (e.g., time=1000, price=1.0850)
4. Click opposite corner (e.g., time=2000, price=1.0860)
5. Verify rectangle appears

**Expected Results:**
- Cursor: crosshair
- After 2 clicks, rectangle is drawn
- Border: 2px solid #3b82f6
- Background: transparent
- Rectangle spans from (min(x1,x2), min(y1,y2)) to (max(x1,x2), max(y1,y2))
- Supports line styles: solid, dashed, dotted
- Draggable and resizable via corner nodes
- Backend save triggered

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-009: Ellipse Tool
**Objective:** Draw elliptical/circular shape

**Steps:**
1. Click "Ellipse" tool in toolbar
2. Cursor changes to crosshair
3. Click first corner of bounding box (e.g., time=1000, price=1.0850)
4. Click opposite corner (e.g., time=2000, price=1.0860)
5. Verify ellipse appears

**Expected Results:**
- Cursor: crosshair
- After 2 clicks, ellipse is drawn
- Border: 2px solid #3b82f6
- Border radius: 50% (creates ellipse)
- Background: transparent
- Ellipse fits within bounding box from (min(x1,x2), min(y1,y2)) to (max(x1,x2), max(y1,y2))
- Supports line styles: solid, dashed, dotted
- Draggable and resizable via corner nodes
- Backend save triggered

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-010: Arrow Tool
**Objective:** Place arrow marker on chart

**Steps:**
1. Click "Arrow" tool in toolbar (may have variants: up, down, right)
2. Cursor changes to crosshair
3. Click once on chart (e.g., time=1500, price=1.0855)
4. Verify arrow icon appears

**Expected Results:**
- Cursor: crosshair
- After 1 click, arrow icon is placed
- Arrow types based on subtype:
  - 'up': ⬆ (upward arrow)
  - 'down': ⬇ (downward arrow)
  - default: ➜ (right arrow)
- Font size: 24px
- Font weight: bold
- Color: #3b82f6 (blue)
- Transform: translate(-50%, -50%) (centered on point)
- Draggable
- Backend save triggered

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-011: Pitchfork Tool
**Objective:** Create Andrew's Pitchfork pattern

**Steps:**
1. Click "Pitchfork" tool in toolbar
2. Cursor changes to crosshair
3. Click first point (pivot/center, e.g., time=1000, price=1.0855)
4. Click second point (upper anchor, e.g., time=1500, price=1.0865)
5. Click third point (lower anchor, e.g., time=1500, price=1.0845)
6. Verify pitchfork appears with 3 lines

**Expected Results:**
- Cursor: crosshair
- After 3 clicks, pitchfork is drawn
- Three lines:
  1. Upper line: from p1 to p2 (solid)
  2. Lower line: from p1 to p3 (solid)
  3. Median line: from p1 to midpoint of (p2, p3) (dashed: '4 4')
- Line color: #3b82f6 (blue)
- Line width: 2px
- SVG-based rendering
- Draggable via nodes (3 control points)
- Backend save triggered

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

## Persistence Tests

### TEST-100: Backend Persistence - Create
**Objective:** Verify drawings are saved to backend on creation

**Steps:**
1. Clear all drawings (backend + localStorage)
2. Create a trendline on EURUSD chart
3. Open browser DevTools Network tab
4. Verify POST request to `/api/workspace/drawings`
5. Check request payload contains:
   ```json
   {
     "id": "drawing-1234567890",
     "type": "trendline",
     "points": [{time: 1000, price: 1.0850}, {time: 2000, price: 1.0860}],
     "color": "#3b82f6",
     "lineWidth": 2,
     "symbol": "EURUSD",
     "accountId": 1
   }
   ```
6. Verify response status: 200 OK
7. Backend returns saved drawing with database ID

**Expected Results:**
- POST request sent immediately after drawing completion
- Request includes all drawing properties + symbol + accountId
- Response includes saved drawing data
- No errors in console

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-101: Backend Persistence - Load
**Objective:** Verify drawings are loaded from backend on chart initialization

**Steps:**
1. Create 3 drawings on EURUSD (trendline, hline, text)
2. Refresh browser page
3. Open DevTools Network tab
4. Verify GET request to `/api/workspace/drawings?symbol=EURUSD&accountId=1`
5. Verify response contains all 3 drawings
6. Verify all 3 drawings appear on chart after load

**Expected Results:**
- GET request sent on chart mount
- Response contains array of drawings for current symbol + accountId
- All drawings are rendered on chart
- Drawings are selectable and editable
- Console log: `[drawingManager] Rendering drawing: {id: ..., type: ..., points: ...}`

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-102: Backend Persistence - Delete
**Objective:** Verify drawings are deleted from backend

**Steps:**
1. Create a trendline on chart
2. Select the trendline (click on it)
3. Press Delete key (or use delete command)
4. Verify DELETE request to `/api/workspace/drawings/{id}?symbol=EURUSD&accountId=1`
5. Verify drawing is removed from chart
6. Refresh page
7. Verify drawing is still gone (not reloaded)

**Expected Results:**
- DELETE request sent immediately after deletion
- Drawing removed from DOM
- Backend confirms deletion (200 OK or 204 No Content)
- Drawing does not reappear on page refresh

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-103: Backend Persistence - Update (Drag/Move)
**Objective:** Verify drawing modifications are saved to backend

**Steps:**
1. Create a trendline on chart
2. Click and drag one of the control nodes (red squares)
3. Release mouse
4. Verify POST request to `/api/workspace/drawings` (update)
5. Request payload includes updated point coordinates
6. Refresh page
7. Verify trendline appears at new position

**Expected Results:**
- POST request sent on mouse up (after drag complete)
- Updated drawing coordinates in payload
- Backend updates drawing record
- Modified drawing persists across page refresh

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-104: LocalStorage Fallback - Backend Down
**Objective:** Verify localStorage saves when backend is unavailable

**Steps:**
1. Stop backend server (Ctrl+C)
2. Create a trendline on EURUSD chart
3. Open DevTools Console
4. Verify error: "Failed to save to backend: ..."
5. Verify drawing is saved to localStorage
6. Check localStorage key: `drawings-EURUSD`
7. Refresh page
8. Verify drawing is loaded from localStorage

**Expected Results:**
- Backend save fails (network error)
- Fallback to localStorage triggered
- Drawing stored in `localStorage['drawings-EURUSD']`
- Drawing persists across page refresh (loaded from localStorage)
- Console warning: "Failed to save to backend"

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-105: Cross-Symbol Isolation
**Objective:** Verify drawings are symbol-specific

**Steps:**
1. Create 2 drawings on EURUSD (trendline + hline)
2. Switch symbol to GBPUSD
3. Verify EURUSD drawings do NOT appear
4. Create 1 drawing on GBPUSD (vline)
5. Switch back to EURUSD
6. Verify EURUSD drawings reappear
7. Verify GBPUSD drawing does NOT appear

**Expected Results:**
- Drawings are scoped to specific symbol
- Switching symbol triggers:
  - `drawingManager.setChart(chart, series, newSymbol, accountId)`
  - GET request to `/api/workspace/drawings?symbol=newSymbol&accountId=1`
- Only drawings for current symbol are rendered
- No cross-contamination between symbols

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-106: Cross-Account Isolation
**Objective:** Verify drawings are account-specific

**Steps:**
1. Login as Account ID = 1
2. Create 2 drawings on EURUSD
3. Logout and login as Account ID = 2
4. Open EURUSD chart
5. Verify Account 1 drawings do NOT appear
6. Create 1 drawing as Account 2
7. Logout and login back as Account 1
8. Verify Account 1 drawings reappear
9. Verify Account 2 drawing does NOT appear

**Expected Results:**
- Drawings are scoped to specific accountId
- Backend queries include `accountId` parameter
- No cross-contamination between accounts
- Each account has isolated drawing sets per symbol

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

## Edge Case Tests

### TEST-200: Backend Unavailable - Create Drawing
**Objective:** Graceful handling when backend is down during creation

**Steps:**
1. Stop backend server
2. Create a trendline on chart
3. Observe behavior

**Expected Results:**
- Drawing is created and rendered on chart
- Backend POST fails (network error)
- Fallback: Drawing saved to localStorage
- Console warning: "Failed to save to backend: ..."
- Drawing persists in localStorage on refresh

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-201: Backend Unavailable - Load Drawings
**Objective:** Graceful handling when backend is down during load

**Steps:**
1. Create drawings with backend running
2. Stop backend server
3. Refresh page
4. Observe behavior

**Expected Results:**
- Backend GET request fails (network error)
- Fallback: Drawings loaded from localStorage
- All previously created drawings appear
- Console warning: "Failed to load from backend: ..."
- User can continue creating/editing drawings (localStorage mode)

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-202: Backend Timeout - Slow Response
**Objective:** Verify behavior with slow backend response

**Steps:**
1. Simulate slow backend (add delay in backend or throttle network)
2. Create a drawing
3. Immediately refresh page before backend responds
4. Observe behavior

**Expected Results:**
- Drawing may or may not persist (race condition)
- If backend save completes before refresh: drawing persists
- If backend save incomplete: drawing lost (no localStorage save yet)
- No crashes or errors
- System recovers gracefully

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-203: Invalid Drawing Data - Missing Points
**Objective:** Verify robustness against corrupted data

**Steps:**
1. Manually edit localStorage or backend data
2. Remove `points` array from a drawing
3. Refresh page
4. Observe behavior

**Expected Results:**
- DrawingManager skips invalid drawing (defensive check)
- No crash or error
- Valid drawings still render
- Console warning: "Invalid drawing data: missing points"

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-204: Chart Disposed During Drawing
**Objective:** Verify no errors if chart is disposed mid-drawing

**Steps:**
1. Start creating a trendline (click first point)
2. Before clicking second point, navigate away or close component
3. Observe console for errors

**Expected Results:**
- No errors in console
- `isDisposedRef.current` prevents operations on disposed chart
- Active drawing is canceled
- Cleanup functions execute properly

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-205: Rapid Drawing Creation
**Objective:** Test performance with rapid tool switching and drawing

**Steps:**
1. Rapidly click tools and create drawings:
   - Trendline (2 clicks)
   - Hline (1 click)
   - Vline (1 click)
   - Rectangle (2 clicks)
   - Ellipse (2 clicks)
2. Repeat 10 times rapidly
3. Observe performance and stability

**Expected Results:**
- All drawings are created and rendered
- No lag or freezing
- No duplicate drawings
- No memory leaks
- All drawings saved to backend (check network tab)

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-206: Drawing on Empty Chart (No Data)
**Objective:** Verify drawing tools work even without price data

**Steps:**
1. Open chart with symbol that has no historical data
2. Chart displays mock/empty data
3. Attempt to create a trendline
4. Observe behavior

**Expected Results:**
- Drawing tools remain functional
- Drawings can be created even without real data
- Coordinates are calculated based on mock data
- No crashes or errors

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-207: Drawing Outside Visible Range
**Objective:** Verify drawings persist when scrolling/zooming

**Steps:**
1. Create a trendline in visible chart area
2. Scroll chart left (move visible range)
3. Verify trendline disappears from view (expected)
4. Scroll back to original position
5. Verify trendline reappears

**Expected Results:**
- Drawings are rendered based on coordinates
- Scrolling/zooming triggers `updateDrawingPositions()`
- Drawings reappear when coordinates are back in visible range
- No performance issues with many drawings

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-208: Text Annotation with Special Characters
**Objective:** Test text handling with special characters

**Steps:**
1. Click Text tool
2. Click on chart
3. Enter text with special characters: `<script>alert('XSS')</script>`
4. Verify text appears safely (no script execution)

**Expected Results:**
- Text is rendered as plain text (no HTML execution)
- Special characters are escaped
- No XSS vulnerability
- Text displays exactly as entered

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-209: Maximum Drawings Limit
**Objective:** Test performance with large number of drawings

**Steps:**
1. Create 100 drawings of various types
2. Observe rendering performance
3. Try to scroll/zoom chart
4. Measure responsiveness

**Expected Results:**
- System handles 100+ drawings without crash
- Rendering performance may degrade slightly
- Chart remains interactive
- Memory usage is reasonable (<500MB)

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-210: Drawing Deletion Undo
**Objective:** Verify undo functionality for deleted drawings

**Steps:**
1. Create 3 drawings (trendline, hline, text)
2. Select and delete trendline
3. Press Ctrl+Z or trigger undo command
4. Verify trendline reappears

**Expected Results:**
- Deleted drawing is stored in `undoBuffer`
- Undo command restores drawing
- Drawing is re-rendered on chart
- Drawing is re-saved to backend (POST request)

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

## UI/UX Tests

### TEST-300: Cursor Change on Tool Selection
**Objective:** Verify cursor changes to crosshair for drawing tools

**Steps:**
1. Select each drawing tool sequentially
2. Observe cursor appearance

**Expected Results:**
- Cursor tool: default arrow
- All other tools: crosshair
- Cursor style applied via: `style={{ cursor: activeDrawingType ? 'crosshair' : 'default' }}`

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-301: Selection Nodes Appearance
**Objective:** Verify selected drawings show control nodes

**Steps:**
1. Create a trendline
2. Click on trendline to select
3. Observe control nodes

**Expected Results:**
- Red squares appear at control points
- Node style:
  - Background: #ff0000 (red)
  - Border: 1px solid #ffffff (white)
  - Width: 7px
  - Height: 7px
  - Border radius: 0 (square)
- Nodes appear ONLY when drawing is selected
- Nodes disappear when drawing is deselected

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-302: Drag and Drop Smoothness
**Objective:** Verify smooth dragging experience

**Steps:**
1. Create a trendline
2. Select it
3. Drag one of the control nodes
4. Observe dragging smoothness

**Expected Results:**
- Node follows mouse cursor smoothly (no lag)
- Drawing updates in real-time during drag
- Cursor changes to indicate drag mode
- Chart scrolling/zooming is disabled during drag
- Chart scrolling/zooming is re-enabled after mouse up

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-303: Drawing Context Menu
**Objective:** Verify right-click context menu on drawings

**Steps:**
1. Create a trendline
2. Right-click on trendline
3. Verify context menu appears

**Expected Results:**
- Context menu appears at cursor position
- Menu options:
  - Delete
  - Duplicate
  - Properties
  - Bring to Front
  - Send to Back
  - Lock/Unlock
  - Hide/Show
- Menu closes on click outside or selection

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-304: Drawing Properties Panel
**Objective:** Verify properties panel shows drawing details

**Steps:**
1. Create a trendline
2. Select trendline
3. Verify DrawingPropertiesPanel appears
4. Modify color to red (#ff0000)
5. Verify trendline color updates

**Expected Results:**
- Properties panel appears on selection
- Panel shows:
  - Drawing type
  - Color picker
  - Line width slider
  - Line style (solid/dashed/dotted)
- Changes apply immediately to drawing
- Changes are saved to backend

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

### TEST-305: Drawing List Panel
**Objective:** Verify drawing list shows all drawings

**Steps:**
1. Create 5 drawings (mixed types)
2. Trigger "Drawing List" panel (keyboard shortcut or menu)
3. Verify all drawings appear in list

**Expected Results:**
- Panel shows list of all drawings for current symbol
- Each entry shows:
  - Drawing type icon
  - Drawing ID or label
  - Visibility toggle (eye icon)
  - Delete button
- Clicking drawing in list selects it on chart
- Toggling visibility hides/shows drawing

**Actual Results:**
- [ ] Pass
- [ ] Fail (describe issue)

**Notes:**
```
```

---

## Performance Tests

### TEST-400: Rendering Performance - 50 Drawings
**Objective:** Measure rendering time for 50 drawings

**Steps:**
1. Create 50 drawings of mixed types
2. Measure page load time
3. Measure chart scroll/zoom responsiveness

**Expected Results:**
- Initial render: <500ms for all drawings
- Scroll/zoom: smooth (60fps)
- Memory usage: <300MB

**Actual Results:**
- Render time: ____ms
- Scroll FPS: ____
- Memory: ____MB
- [ ] Pass (<500ms render)
- [ ] Fail

**Notes:**
```
```

---

### TEST-401: Backend Save Performance
**Objective:** Measure backend save latency

**Steps:**
1. Create 10 drawings rapidly
2. Measure time for each POST request

**Expected Results:**
- Average latency: <100ms per save
- No failed requests
- No request queue backup

**Actual Results:**
- Average latency: ____ms
- Failed requests: ____
- [ ] Pass (<100ms average)
- [ ] Fail

**Notes:**
```
```

---

### TEST-402: localStorage Performance
**Objective:** Measure localStorage save/load speed

**Steps:**
1. Create 100 drawings
2. Measure localStorage.setItem() time
3. Refresh page
4. Measure localStorage.getItem() + JSON.parse() time

**Expected Results:**
- Save time: <50ms
- Load time: <50ms
- No data loss

**Actual Results:**
- Save time: ____ms
- Load time: ____ms
- [ ] Pass (<50ms each)
- [ ] Fail

**Notes:**
```
```

---

## Test Results Template

### Test Summary

| Test ID | Test Name | Status | Date | Tester | Notes |
|---------|-----------|--------|------|--------|-------|
| TEST-001 | Cursor Tool | ⏸ | | | |
| TEST-002 | Trendline Tool | ⏸ | | | |
| TEST-003 | Horizontal Line | ⏸ | | | |
| TEST-004 | Vertical Line | ⏸ | | | |
| TEST-005 | Channel Tool | ⏸ | | | |
| TEST-006 | Fibonacci | ⏸ | | | |
| TEST-007 | Text Annotation | ⏸ | | | |
| TEST-008 | Rectangle Tool | ⏸ | | | |
| TEST-009 | Ellipse Tool | ⏸ | | | |
| TEST-010 | Arrow Tool | ⏸ | | | |
| TEST-011 | Pitchfork Tool | ⏸ | | | |
| TEST-100 | Backend - Create | ⏸ | | | |
| TEST-101 | Backend - Load | ⏸ | | | |
| TEST-102 | Backend - Delete | ⏸ | | | |
| TEST-103 | Backend - Update | ⏸ | | | |
| TEST-104 | LocalStorage Fallback | ⏸ | | | |
| TEST-105 | Cross-Symbol Isolation | ⏸ | | | |
| TEST-106 | Cross-Account Isolation | ⏸ | | | |
| TEST-200 | Backend Down - Create | ⏸ | | | |
| TEST-201 | Backend Down - Load | ⏸ | | | |
| TEST-202 | Backend Timeout | ⏸ | | | |
| TEST-203 | Invalid Data | ⏸ | | | |
| TEST-204 | Chart Disposed | ⏸ | | | |
| TEST-205 | Rapid Creation | ⏸ | | | |
| TEST-206 | Empty Chart | ⏸ | | | |
| TEST-207 | Outside Visible Range | ⏸ | | | |
| TEST-208 | Special Characters | ⏸ | | | |
| TEST-209 | Max Drawings | ⏸ | | | |
| TEST-210 | Undo Deletion | ⏸ | | | |
| TEST-300 | Cursor Change | ⏸ | | | |
| TEST-301 | Selection Nodes | ⏸ | | | |
| TEST-302 | Drag Smoothness | ⏸ | | | |
| TEST-303 | Context Menu | ⏸ | | | |
| TEST-304 | Properties Panel | ⏸ | | | |
| TEST-305 | Drawing List | ⏸ | | | |
| TEST-400 | Perf - 50 Drawings | ⏸ | | | |
| TEST-401 | Perf - Backend | ⏸ | | | |
| TEST-402 | Perf - LocalStorage | ⏸ | | | |

### Status Legend
- ⏸ Not Started
- 🔄 In Progress
- ✅ Pass
- ❌ Fail
- ⚠️ Pass with Issues

---

## Appendix A: Known Issues

### Issue 1: Drawings May Not Persist Across Browser Sessions
**Severity:** Medium
**Description:** If backend is down during session, drawings are only in localStorage and may not sync to backend later.
**Workaround:** Ensure backend is running before creating important drawings.

### Issue 2: Drawing Overlay z-index Conflicts
**Severity:** Low
**Description:** Drawing overlay may conflict with other UI elements (e.g., modals, dropdowns).
**Workaround:** Ensure `.chart-drawing-overlay` has `z-index: 10`.

---

## Appendix B: Code References

### Key Files
- **DrawingManager:** `clients/desktop/src/services/drawingManager.ts`
- **TradingChart:** `clients/desktop/src/components/TradingChart.tsx`
- **Backend API:** `backend/handlers/workspace_drawings.go` (assumed)

### Drawing Manager Methods
- `startDrawing(type, color, subtype)` - Initialize new drawing
- `addPoint(time, price)` - Add point to active drawing
- `finishDrawing()` - Complete and save drawing
- `deleteDrawing(id)` - Remove drawing
- `selectDrawing(id)` - Select drawing for editing
- `renderAllDrawings()` - Re-render all drawings
- `saveToBackend(drawing)` - Persist to backend
- `loadFromBackend(symbol)` - Fetch from backend
- `saveToStorage(symbol)` - Fallback localStorage save
- `loadFromStorage(symbol)` - Fallback localStorage load

---

## Appendix C: Backend API Specification

### Endpoint: POST /api/workspace/drawings
**Description:** Save or update drawing

**Request:**
```json
{
  "id": "drawing-1234567890",
  "type": "trendline",
  "points": [
    {"time": 1000, "price": 1.0850},
    {"time": 2000, "price": 1.0860}
  ],
  "color": "#3b82f6",
  "lineWidth": 2,
  "symbol": "EURUSD",
  "accountId": 1
}
```

**Response:**
```json
{
  "id": "drawing-db-id-12345",
  "type": "trendline",
  "points": [...],
  "color": "#3b82f6",
  "lineWidth": 2,
  "symbol": "EURUSD",
  "accountId": 1,
  "createdAt": "2026-02-13T10:30:00Z"
}
```

### Endpoint: GET /api/workspace/drawings
**Description:** Load all drawings for symbol and account

**Query Params:**
- `symbol` (required): Symbol name (e.g., EURUSD)
- `accountId` (required): Account ID (e.g., 1)

**Response:**
```json
[
  {
    "id": "drawing-db-id-1",
    "type": "trendline",
    "points": [...],
    ...
  },
  {
    "id": "drawing-db-id-2",
    "type": "hline",
    "points": [...],
    ...
  }
]
```

### Endpoint: DELETE /api/workspace/drawings/{id}
**Description:** Delete specific drawing

**Query Params:**
- `symbol` (required)
- `accountId` (required)

**Response:**
- Status: 200 OK or 204 No Content

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-02-13 | Claude Sonnet 4.5 | Initial comprehensive test plan |

---

**END OF TEST PLAN**
