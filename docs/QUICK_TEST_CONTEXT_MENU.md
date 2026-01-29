# Quick Test Guide - Context Menu Fixes

**Test Duration**: ~5 minutes
**Status**: Ready to test

---

## 1. Edge Clipping Test (2 min)

**Steps**:
1. Start the desktop client
2. Go to Market Watch panel
3. Right-click symbol at **top-right corner** of screen
   - ✅ Menu should flip left/down to stay in viewport
4. Right-click symbol at **bottom-right corner** of screen
   - ✅ Menu should flip left/up to stay in viewport
5. Right-click symbol at **bottom-left corner** of screen
   - ✅ Menu should flip down/right to stay in viewport

**Pass Criteria**: No menu clipping at any edge.

---

## 2. Keyboard Navigation Test (2 min)

**Steps**:
1. Right-click any symbol to open context menu
2. Press **Down Arrow** 3 times
   - ✅ Should highlight "Quick Buy (0.01)"
3. Press **Up Arrow** 1 time
   - ✅ Should highlight "New Order"
4. Press **N** key
   - ✅ Should jump to "New Order"
5. Press **S** key
   - ✅ Should jump to "Sort" or "Symbols"
6. Press **Right Arrow** on "Sort"
   - ✅ Should open submenu
7. Press **Escape**
   - ✅ Should close menu

**Pass Criteria**: All keyboard shortcuts work.

---

## 3. Submenu Test (1 min)

**Steps**:
1. Right-click any symbol
2. Hover over "Columns"
3. Wait 300ms
   - ✅ Submenu should appear after delay
4. Hover over nested submenu item (if any)
   - ✅ Nested submenu should appear above parent

**Pass Criteria**: 300ms delay, submenus on top.

---

## 4. Column Persistence Test (1 min)

**Steps**:
1. Right-click Market Watch
2. Navigate to **Columns** > **Daily %**
3. Click to toggle off
   - ✅ Column should disappear immediately
4. Refresh page (F5)
   - ✅ Column should stay hidden
5. Right-click > **Columns** > **Daily %**
6. Click to toggle on
   - ✅ Column should reappear

**Pass Criteria**: Column state persists across refresh.

---

## Quick Visual Check

**Menu Appearance**:
- ✅ Dark theme (#1e1e1e background)
- ✅ 8px padding from viewport edges
- ✅ Smooth opacity transition (no flash)
- ✅ Highlighted items have blue background (#3b82f6)
- ✅ Submenus have slight overlap (4px)

---

## If Test Fails

1. Check browser console for errors
2. Verify `localStorage` has `rtx5_marketwatch_cols` key
3. Test in different browser (Chrome vs Firefox)
4. Check `ContextMenu.tsx` changes applied correctly

---

## Report Issues

If any test fails, note:
- Which test failed
- Browser and OS
- Screenshot of issue
- Console errors
