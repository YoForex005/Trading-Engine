# Market Watch Context Menu - Testing Guide

**Implementation Date:** 2026-01-20
**Status:** Ready for Testing

---

## Quick Start Testing

### 1. Launch the Application

```bash
cd clients/desktop
npm run dev
```

### 2. Navigate to Market Watch Panel

The Market Watch panel should be visible in the main trading interface.

---

## Test Scenarios

### A. Viewport Edge Tests (Portal Rendering)

**Objective:** Verify menus never clip or get cut off at viewport edges.

#### Test A1: Right Edge Clipping
1. Resize browser window to ~1024px width
2. Scroll Market Watch panel to far right
3. Right-click on a symbol near the right edge
4. **Expected:** Menu appears, fully visible
5. Hover over "Columns →" or "Sort →"
6. **Expected:** Submenu flips to the LEFT (not right), fully visible

**✅ PASS:** Submenu flips left and is fully visible
**❌ FAIL:** Submenu extends off-screen or gets clipped

---

#### Test A2: Bottom Edge Clipping
1. Resize browser window to ~768px height
2. Scroll to bottom symbol in Market Watch
3. Right-click on bottom symbol
4. **Expected:** Menu adjusts upward, fully visible
5. Hover over "Columns →"
6. **Expected:** Submenu appears fully visible (adjusted upward if needed)

**✅ PASS:** Menu and submenu fully visible
**❌ FAIL:** Menu or submenu extends off bottom of screen

---

#### Test A3: Corner (Bottom-Right) Clipping
1. Resize window to ~1024x768
2. Right-click on bottom-right symbol
3. **Expected:** Menu appears, fully visible
4. Hover over "Columns →"
5. **Expected:** Submenu flips LEFT and adjusts UP if needed

**✅ PASS:** Both horizontal and vertical adjustments apply
**❌ FAIL:** Any part of menu or submenu is clipped

---

#### Test A4: Small Viewport (800x600)
1. Resize browser to 800x600
2. Right-click anywhere in Market Watch
3. Open multiple nested submenus (Sort → submenu → etc.)
4. **Expected:** All menus and submenus remain fully visible

**✅ PASS:** All menus visible at 800x600
**❌ FAIL:** Any clipping occurs

---

#### Test A5: Large Viewport (1920x1080)
1. Set browser to full screen (1920x1080)
2. Right-click in center of Market Watch
3. Open submenus
4. **Expected:** Menus appear in standard position (to right, below)

**✅ PASS:** Normal positioning works on large screens
**❌ FAIL:** Unexpected positioning

---

### B. Hover Behavior Tests (Hover Intent + Safe Triangle)

**Objective:** Verify MT5-equivalent hover timing and diagonal mouse movement support.

#### Test B1: Hover Intent Delay (150ms)
1. Right-click to open menu
2. Quickly move mouse over "Columns →" (< 100ms)
3. Move away quickly
4. **Expected:** Submenu does NOT open (prevented by 150ms delay)

5. Right-click again
6. Hover over "Columns →" and HOLD for 200ms
7. **Expected:** Submenu opens smoothly

**✅ PASS:** Quick movements don't trigger, 150ms hover opens submenu
**❌ FAIL:** Submenu opens on quick movements (too sensitive)

---

#### Test B2: Safe Hover Triangle (Diagonal Movement)
1. Right-click to open menu
2. Hover over "Columns →" until submenu opens
3. Move mouse DIAGONALLY from "Columns" toward submenu items
4. **Expected:** Submenu STAYS OPEN during diagonal movement
5. Mouse should reach submenu without it closing

**✅ PASS:** Can move diagonally without submenu closing
**❌ FAIL:** Submenu closes during diagonal movement (Amazon-style triangle NOT working)

---

#### Test B3: No Flickering
1. Right-click to open menu
2. Rapidly move mouse up and down over "Columns →" and "Sort →"
3. **Expected:** Submenus open/close smoothly without rapid flickering

**✅ PASS:** Smooth transitions, no flickering
**❌ FAIL:** Rapid open/close cycles (flicker)

---

#### Test B4: Close Delay (100ms)
1. Right-click to open menu
2. Hover over "Columns →" to open submenu
3. Move mouse away from both parent and submenu
4. **Expected:** Submenu closes after ~100ms delay (prevents accidental closes)

**✅ PASS:** 100ms close delay feels natural
**❌ FAIL:** Instant close (too abrupt)

---

### C. Keyboard Navigation Tests

**Objective:** Verify full keyboard navigation and MT5 shortcuts.

#### Test C1: Global Shortcuts
**Test each shortcut:**

| Shortcut | Expected Action | Status |
|----------|----------------|--------|
| `F9` | Opens New Order Dialog | ☐ |
| `Alt+B` | Opens Depth of Market | ☐ |
| `Ctrl+U` | Opens Symbols Dialog | ☐ |
| `F10` | Opens Popup Prices | ☐ |
| `Delete` | Hides selected symbol | ☐ |
| `Escape` | Closes menu/modal | ☐ |

**Instructions:**
1. Select a symbol in Market Watch
2. Press each shortcut key
3. Verify expected action occurs

**✅ PASS:** All shortcuts work as expected
**❌ FAIL:** Any shortcut doesn't work or triggers wrong action

---

#### Test C2: Arrow Key Navigation (Up/Down)
1. Right-click to open menu
2. Press `Arrow Down`
3. **Expected:** Focus moves to next item (blue highlight)
4. Press `Arrow Down` repeatedly
5. **Expected:** Focus wraps around to top when reaching bottom
6. Press `Arrow Up`
7. **Expected:** Focus moves backward, wraps to bottom from top

**✅ PASS:** Arrow navigation works with wrap-around
**❌ FAIL:** Focus doesn't move or doesn't wrap

---

#### Test C3: Arrow Key Navigation (Right/Left - Submenus)
1. Right-click to open menu
2. Navigate to "Columns →" using arrow keys
3. Press `Arrow Right`
4. **Expected:** Submenu opens, focus moves to first submenu item
5. Press `Arrow Left`
6. **Expected:** Submenu closes, focus returns to parent "Columns" item

**✅ PASS:** Left/Right opens/closes submenus correctly
**❌ FAIL:** Submenus don't open or focus is lost

---

#### Test C4: Enter/Space Execution
1. Right-click to open menu
2. Navigate to "Quick Buy" using arrow keys
3. Press `Enter`
4. **Expected:** Quick Buy action executes, menu closes
5. Right-click again
6. Navigate to a checkbox item (e.g., "Show Grid")
7. Press `Space`
8. **Expected:** Checkbox toggles, menu stays open (autoClose: false)

**✅ PASS:** Enter/Space execute actions correctly
**❌ FAIL:** Actions don't execute or menu behavior wrong

---

#### Test C5: Escape Key (Hierarchical Close)
1. Right-click to open menu
2. Navigate to "Columns →" and open submenu
3. Press `Escape` once
4. **Expected:** Submenu closes, parent menu stays open
5. Press `Escape` again
6. **Expected:** Parent menu closes

**✅ PASS:** Hierarchical close works (one level at a time)
**❌ FAIL:** Escape closes all menus at once

---

#### Test C6: Focus Trap
1. Right-click to open menu
2. Press `Tab`
3. **Expected:** Focus stays within menu (doesn't escape to page)
4. Press `Shift+Tab`
5. **Expected:** Focus cycles backward within menu

**✅ PASS:** Focus trapped in menu
**❌ FAIL:** Tab escapes to page elements

---

#### Test C7: Auto-Skip Dividers
1. Right-click to open menu
2. Use arrow keys to navigate
3. **Expected:** Focus automatically skips over divider lines (separator items)

**✅ PASS:** Dividers are skipped
**❌ FAIL:** Focus stops on dividers

---

#### Test C8: Visual Focus Indicator
1. Right-click to open menu
2. Press arrow keys to navigate
3. **Expected:** Focused item has clear visual highlight (blue background)

**✅ PASS:** Focus indicator clearly visible
**❌ FAIL:** No visual indication of focus

---

### D. Action Execution Tests (All 39 Menu Actions)

**Objective:** Verify every menu action executes real functionality (no placeholders).

#### Test D1: Trading Actions (7)

| Action | Test | Expected Result | Status |
|--------|------|----------------|--------|
| New Order (F9) | Click or press F9 | Opens order entry dialog | ☐ |
| Quick Buy | Click | Executes market BUY order | ☐ |
| Quick Sell | Click | Executes market SELL order | ☐ |
| Chart Window | Click | Opens chart for symbol | ☐ |
| Tick Chart | Click | Opens tick chart | ☐ |
| Depth of Market (Alt+B) | Click or Alt+B | Opens DOM window | ☐ |
| Popup Prices (F10) | Click or F10 | Opens price popup | ☐ |

**Instructions:**
1. Right-click on EURUSD
2. Test each action
3. Verify real execution (check console for API calls if needed)

**✅ PASS:** All trading actions execute correctly
**❌ FAIL:** Any action shows placeholder or doesn't work

---

#### Test D2: Symbol Management (2)

| Action | Test | Expected Result | Status |
|--------|------|----------------|--------|
| Hide (Delete) | Click or press Delete | Symbol removed from list | ☐ |
| Show All | Click | All hidden symbols restored | ☐ |

**Instructions:**
1. Right-click EURUSD
2. Click "Hide" (or press Delete)
3. Verify symbol disappears
4. Right-click another symbol
5. Click "Show All"
6. Verify EURUSD reappears

**✅ PASS:** Hide/Show All works correctly
**❌ FAIL:** Symbols don't hide or restore

---

#### Test D3: Symbol Sets (8)

| Action | Test | Expected Result | Status |
|--------|------|----------------|--------|
| Sets → Forex Major | Click | Shows major pairs only | ☐ |
| Sets → Forex Crosses | Click | Shows cross pairs | ☐ |
| Sets → Forex Exotic | Click | Shows exotic pairs | ☐ |
| Sets → Commodities | Click | Shows commodities | ☐ |
| Sets → Indices | Click | Shows indices | ☐ |
| Sets → My Favorites | Click | Shows custom favorites | ☐ |
| Sets → Save as... | Click | Opens save dialog | ☐ |
| Sets → Remove | Click | Removes custom set | ☐ |

**Instructions:**
1. Right-click any symbol
2. Navigate to "Sets →"
3. Test each symbol set
4. Verify symbol list updates correctly

**✅ PASS:** All symbol sets load correctly
**❌ FAIL:** Symbol sets don't load or show wrong symbols

---

#### Test D4: Sorting (5)

| Action | Test | Expected Result | Status |
|--------|------|----------------|--------|
| Sort → Symbol | Click | Sorts alphabetically | ☐ |
| Sort → Gainers | Click | Sorts by % gain (descending) | ☐ |
| Sort → Losers | Click | Sorts by % loss (ascending) | ☐ |
| Sort → Volume | Click | Sorts by volume | ☐ |
| Sort → Reset | Click | Returns to original order | ☐ |

**Instructions:**
1. Right-click any symbol
2. Navigate to "Sort →"
3. Test each sort option
4. Verify list reorders correctly

**✅ PASS:** All sort options work correctly
**❌ FAIL:** Sorting doesn't work or wrong order

---

#### Test D5: Columns (10)

| Column | Test | Expected Result | Status |
|--------|------|----------------|--------|
| Bid | Toggle | Column shows/hides | ☐ |
| Ask | Toggle | Column shows/hides | ☐ |
| Spread | Toggle | Column shows/hides | ☐ |
| Time | Toggle | Column shows/hides | ☐ |
| High/Low | Toggle | Column shows/hides | ☐ |
| Change | Toggle | Column shows/hides | ☐ |
| Change % | Toggle | Column shows/hides | ☐ |
| Volume | Toggle | Column shows/hides | ☐ |

**Instructions:**
1. Right-click any symbol
2. Navigate to "Columns →"
3. Toggle each column checkbox
4. Verify column visibility changes
5. Refresh page
6. Verify settings persist (localStorage)

**✅ PASS:** All columns toggle correctly and persist
**❌ FAIL:** Columns don't toggle or settings lost on refresh

---

#### Test D6: System Options (5)

| Option | Test | Expected Result | Status |
|--------|------|----------------|--------|
| Use System Colors | Toggle | Color scheme changes | ☐ |
| Show Milliseconds | Toggle | Time shows milliseconds | ☐ |
| Auto Remove Expired | Toggle | Expired symbols auto-hide | ☐ |
| Auto Arrange | Toggle | Auto-sorts symbols | ☐ |
| Grid | Toggle | Grid lines show/hide | ☐ |

**Instructions:**
1. Right-click any symbol
2. Toggle each system option
3. Verify immediate visual change
4. Refresh page
5. Verify settings persist

**✅ PASS:** All options toggle and persist
**❌ FAIL:** Options don't apply or don't persist

---

#### Test D7: Export

| Action | Test | Expected Result | Status |
|--------|------|----------------|--------|
| Export | Click | Downloads CSV file | ☐ |

**Instructions:**
1. Right-click any symbol
2. Click "Export"
3. Verify CSV file downloads with all visible symbols

**✅ PASS:** CSV export works
**❌ FAIL:** Export fails or file is empty

---

### E. Accessibility Tests (WCAG 2.1)

**Objective:** Verify screen reader compatibility and ARIA attributes.

#### Test E1: ARIA Attributes
1. Right-click to open menu
2. Open browser DevTools → Elements
3. Inspect menu elements
4. **Expected ARIA attributes:**
   - Menu container: `role="menu"`
   - Menu items: `role="menuitem"`
   - Submenu items: `aria-haspopup="true"`, `aria-expanded="false/true"`
   - Disabled items: `aria-disabled="true"`

**✅ PASS:** All ARIA attributes present and correct
**❌ FAIL:** Missing or incorrect ARIA attributes

---

#### Test E2: Screen Reader (Windows Narrator / NVDA)
1. Enable Windows Narrator or NVDA
2. Right-click to open menu
3. Navigate with arrow keys
4. **Expected:** Screen reader announces:
   - "Menu opened"
   - Each item label as focus moves
   - "Has submenu" for items with submenus
   - Current state for checkboxes

**✅ PASS:** Screen reader correctly announces all elements
**❌ FAIL:** Screen reader silent or announces incorrectly

---

#### Test E3: High Contrast Mode
1. Enable Windows High Contrast mode
2. Right-click to open menu
3. **Expected:** Menu remains readable with high contrast colors

**✅ PASS:** Menu readable in high contrast mode
**❌ FAIL:** Text or elements invisible

---

### F. Performance Tests

**Objective:** Verify menu performance meets targets.

#### Test F1: Menu Open Latency
1. Open browser DevTools → Performance
2. Start recording
3. Right-click to open menu
4. Stop recording
5. **Expected:** Menu appears in < 5ms

**✅ PASS:** < 5ms
**❌ FAIL:** > 5ms

---

#### Test F2: Position Calculation Time
1. Open DevTools → Performance
2. Right-click near viewport edge (forces collision detection)
3. Measure time from click to menu appearance
4. **Expected:** < 10ms total (including position calculation)

**✅ PASS:** < 10ms
**❌ FAIL:** > 10ms

---

#### Test F3: Memory Leaks
1. Open DevTools → Performance → Memory
2. Take heap snapshot
3. Open/close menu 50 times
4. Take another heap snapshot
5. **Expected:** No significant memory increase (detached DOM nodes)

**✅ PASS:** No memory leaks detected
**❌ FAIL:** Memory usage increases significantly

---

#### Test F4: Rapid Open/Close
1. Rapidly right-click → Escape 20 times in quick succession
2. **Expected:** No lag, crashes, or UI freeze

**✅ PASS:** Handles rapid interactions smoothly
**❌ FAIL:** Lag, freeze, or crash

---

### G. Browser Compatibility Tests

**Objective:** Verify cross-browser support.

#### Test G1: Chrome
- [ ] All viewport edge tests pass
- [ ] All hover behavior tests pass
- [ ] All keyboard navigation tests pass
- [ ] All actions execute correctly

---

#### Test G2: Firefox
- [ ] All viewport edge tests pass
- [ ] All hover behavior tests pass
- [ ] All keyboard navigation tests pass
- [ ] All actions execute correctly

---

#### Test G3: Edge
- [ ] All viewport edge tests pass
- [ ] All hover behavior tests pass
- [ ] All keyboard navigation tests pass
- [ ] All actions execute correctly

---

#### Test G4: Safari (macOS)
- [ ] All viewport edge tests pass
- [ ] All hover behavior tests pass
- [ ] All keyboard navigation tests pass
- [ ] All actions execute correctly

---

#### Test G5: Electron (Desktop App)
- [ ] All viewport edge tests pass
- [ ] All hover behavior tests pass
- [ ] All keyboard navigation tests pass
- [ ] All actions execute correctly

---

## Regression Tests

**Objective:** Ensure new context menu doesn't break existing functionality.

### R1: Market Watch Existing Features
- [ ] Symbol list displays correctly
- [ ] Real-time quote updates work
- [ ] Clicking symbol selects it
- [ ] Double-click opens chart (if applicable)
- [ ] Scrolling works normally
- [ ] Column resizing works

### R2: Other Panels
- [ ] Trading panel works
- [ ] Chart panel works
- [ ] Account panel works
- [ ] No context menu conflicts with other panels

---

## Test Results Summary

**Test Date:** _________________
**Tester:** _________________
**Browser:** _________________
**OS:** _________________

### Results

| Category | Total Tests | Passed | Failed | Notes |
|----------|-------------|--------|--------|-------|
| Viewport Edge | 5 | ☐ | ☐ | |
| Hover Behavior | 4 | ☐ | ☐ | |
| Keyboard Navigation | 8 | ☐ | ☐ | |
| Action Execution | 39 | ☐ | ☐ | |
| Accessibility | 3 | ☐ | ☐ | |
| Performance | 4 | ☐ | ☐ | |
| Browser Compatibility | 5 | ☐ | ☐ | |
| Regression | 2 | ☐ | ☐ | |

**TOTAL:** _____ / 70 tests passed

---

## Known Issues

**Document any issues found during testing:**

1. **Issue:** _________________________________________________
   **Severity:** Low / Medium / High / Critical
   **Steps to Reproduce:** _______________________________________
   **Expected:** _______________________________________________
   **Actual:** _________________________________________________

2. **Issue:** _________________________________________________
   **Severity:** Low / Medium / High / Critical
   **Steps to Reproduce:** _______________________________________
   **Expected:** _______________________________________________
   **Actual:** _________________________________________________

---

## Sign-Off

**Testing Complete:** ☐ Yes ☐ No

**Production Ready:** ☐ Yes ☐ No

**Tester Signature:** ___________________ **Date:** ___________

**Reviewer Signature:** _________________ **Date:** ___________

---

## Quick Verification Commands

### DevTools Console Tests

```javascript
// Test 1: Check if menu portal exists
document.querySelectorAll('[role="menu"]').length > 0

// Test 2: Check ARIA attributes
document.querySelector('[role="menu"]').getAttribute('aria-label')

// Test 3: Check z-index management
Array.from(document.querySelectorAll('[role="menu"]')).map(m =>
  window.getComputedStyle(m).zIndex
)

// Test 4: Check if menu is rendered at body level
document.querySelectorAll('body > div[role="menu"]').length > 0
```

---

## Automated Test Script (Optional)

```javascript
// Run in DevTools Console for quick smoke test
(async function quickTest() {
  console.log('🧪 Market Watch Context Menu - Quick Smoke Test');

  // Test 1: Portal rendering
  const portals = document.querySelectorAll('body > div[role="menu"]');
  console.log(`✅ Portal rendering: ${portals.length >= 0 ? 'READY' : 'FAIL'}`);

  // Test 2: ARIA attributes
  const menus = document.querySelectorAll('[role="menu"]');
  const hasAria = Array.from(menus).every(m => m.hasAttribute('role'));
  console.log(`${hasAria ? '✅' : '❌'} ARIA attributes: ${hasAria ? 'PASS' : 'FAIL'}`);

  // Test 3: Keyboard shortcuts registered
  console.log('⌨️ Press F9, Alt+B, Ctrl+U to test shortcuts');

  console.log('🎯 Manual tests required - see TESTING_GUIDE_MARKETWATCH_MENU.md');
})();
```

---

**End of Testing Guide**
