# Quick Fix: Vite Dev Server Cache Issue

## Problem

You're seeing this error:
```
MarketWatchPanel.tsx:3 Uncaught SyntaxError: The requested module '/src/components/ui/ContextMenu.tsx' does not provide an export named 'ContextMenuItemConfig'
```

**Root Cause:** Vite dev server hasn't picked up the new `ContextMenu.tsx` file. This is a caching issue.

## Solution (Choose One)

### Option 1: Restart Dev Server (Recommended)

```bash
# Stop the current dev server (Ctrl+C in the terminal)
# Then restart:
cd clients/desktop
npm run dev
```

### Option 2: Clear Vite Cache + Restart

```bash
cd clients/desktop

# Clear Vite cache
rm -rf node_modules/.vite

# Clear dist folder
rm -rf dist

# Restart dev server
npm run dev
```

### Option 3: Force Reload in Browser

1. Stop dev server (Ctrl+C)
2. Clear Vite cache: `rm -rf clients/desktop/node_modules/.vite`
3. Restart dev server: `npm run dev`
4. In browser: **Ctrl+Shift+R** (hard reload) or **Ctrl+F5**

## Verification

After restarting, you should see the app load without errors. Check the console:

```javascript
// No errors - the import should work
import { ContextMenu, ContextMenuItemConfig } from '../ui/ContextMenu';
```

## Why This Happens

Vite uses aggressive caching for performance. When new files are created (like our `ContextMenu.tsx`), sometimes the cache doesn't invalidate properly, especially if:

1. The file was created while dev server was running
2. Multiple agents were writing files simultaneously
3. File watchers didn't trigger properly

## Confirm Fix Worked

After restart, right-click on any symbol in Market Watch. You should see:
- ✅ Context menu appears
- ✅ No console errors
- ✅ Submenus work
- ✅ No clipping at viewport edges

---

**If issue persists after trying all options above, check:**

1. Is `ContextMenu.tsx` actually saved? (Check file exists)
2. Are there any TypeScript errors in the file?
3. Is the import path correct in `MarketWatchPanel.tsx`?

Run: `ls -la clients/desktop/src/components/ui/ContextMenu.tsx`

Expected: File exists, ~576 lines, ~18KB
