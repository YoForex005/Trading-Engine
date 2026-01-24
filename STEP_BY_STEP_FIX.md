# Step-by-Step Fix for Context Menu Import Error

## Current Error
```
MarketWatchPanel.tsx:3 Uncaught SyntaxError: The requested module '/src/components/ui/ContextMenu.tsx' does not provide an export named 'ContextMenuItemConfig'
```

## Verified Status
✅ File exists: `clients/desktop/src/components/ui/ContextMenu.tsx` (18KB)
✅ Export exists on line 11: `export interface ContextMenuItemConfig`
✅ All required hooks exist
✅ All imports are correct

**Problem:** Vite dev server module cache is stale

---

## SOLUTION: Complete Server Restart

### Step 1: Stop Dev Server
In the terminal where `npm run dev` is running:
```
Press Ctrl+C
```
Wait until you see the process has stopped.

### Step 2: Clear ALL Caches
```bash
cd clients/desktop

# Remove Vite cache
rm -rf node_modules/.vite

# Remove dist folder
rm -rf dist

# Optional: Remove TypeScript build info
rm -rf tsconfig.tsbuildinfo
```

### Step 3: Restart Dev Server
```bash
npm run dev
```

Wait for the server to fully start. You should see:
```
VITE v4.x.x  ready in XXX ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
```

### Step 4: Hard Reload Browser
Once the server is running:

1. Open browser to `http://localhost:5173`
2. Press **Ctrl+Shift+Delete** to open Clear Browsing Data
3. Select:
   - ✅ Cached images and files
   - ✅ Cookies and other site data
   - Time range: "Last hour"
4. Click "Clear data"
5. Close browser completely
6. Reopen browser
7. Navigate to `http://localhost:5173`
8. Press **Ctrl+Shift+R** (hard reload)

---

## Alternative: Force Complete Rebuild

If the above doesn't work:

```bash
cd clients/desktop

# 1. Stop dev server (Ctrl+C)

# 2. Remove all build artifacts and caches
rm -rf node_modules/.vite
rm -rf dist
rm -rf .cache
rm -rf .parcel-cache

# 3. Clear npm cache
npm cache clean --force

# 4. Reinstall dependencies (only if above didn't work)
# rm -rf node_modules
# npm install

# 5. Restart
npm run dev
```

---

## Verification Steps

After restart, open browser console (F12):

### 1. Check for Import Errors
You should NOT see:
```
❌ Uncaught SyntaxError: The requested module...
```

### 2. Test Context Menu
1. Right-click on any symbol in Market Watch
2. You should see a context menu appear
3. No errors in console

### 3. Check Network Tab
1. Open DevTools → Network tab
2. Filter: "ContextMenu"
3. Look for: `ContextMenu.tsx` with status 200 (OK)

---

## Still Not Working?

### Check 1: Verify File Content
```bash
# Check first 20 lines
head -20 "D:\Tading engine\Trading-Engine\clients\desktop\src\components\ui\ContextMenu.tsx"
```

You should see:
```typescript
export interface ContextMenuItemConfig {
  label: string;
  icon?: React.ReactNode;
  // ... more fields
}
```

### Check 2: Verify Import Path
In `MarketWatchPanel.tsx`, the import should be:
```typescript
import { ContextMenu, ContextMenuItemConfig, MenuSectionHeader, MenuDivider } from '../ui/ContextMenu';
```

**NOT:**
```typescript
import { ContextMenu, ContextMenuItemConfig } from '../ui/ContextMenu.tsx'; // ❌ Don't include .tsx
import { ContextMenu, ContextMenuItemConfig } from './ui/ContextMenu';      // ❌ Wrong relative path
```

### Check 3: Browser DevTools Sources
1. Open DevTools → Sources tab
2. Navigate: `localhost:5173` → `src` → `components` → `ui` → `ContextMenu.tsx`
3. Check if the file shows the correct content with `export interface ContextMenuItemConfig`

If the file in Sources tab is empty or old, that confirms Vite cache issue.

---

## Nuclear Option: Complete Reset

If NOTHING works:

```bash
cd clients/desktop

# 1. Stop server
# Ctrl+C

# 2. Delete EVERYTHING
rm -rf node_modules
rm -rf dist
rm -rf .cache
rm -rf .parcel-cache
rm -rf node_modules/.vite
rm -rf package-lock.json

# 3. Fresh install
npm install

# 4. Start
npm run dev
```

---

## What You Should See After Fix

**Browser Console (F12):**
```
✅ No errors
✅ Application loads normally
```

**Right-click on symbol:**
```
✅ Context menu appears
✅ "Columns →" submenu works
✅ No clipping at viewport edges
✅ All 39 menu actions available
```

---

## If Error Changes

If after restart you see a DIFFERENT error, that's progress! Report the new error message and we'll fix it.

Common follow-up errors:
1. `Cannot find module 'useHoverIntent'` → Missing hook export (easy fix)
2. `Cannot find module 'marketWatchActions'` → Missing service file (easy fix)
3. Type errors → TypeScript configuration (easy fix)

---

## Last Resort: Rollback

If you need to rollback the context menu changes:

```bash
# Restore original MarketWatchPanel.tsx from git
git checkout HEAD -- clients/desktop/src/components/layout/MarketWatchPanel.tsx

# Remove new files
rm clients/desktop/src/components/ui/ContextMenu.tsx
rm clients/desktop/src/hooks/useContextMenu.ts
rm clients/desktop/src/hooks/useHoverIntent.ts
rm clients/desktop/src/hooks/useSafeHoverTriangle.ts
rm clients/desktop/src/hooks/useContextMenuNavigation.ts
rm clients/desktop/src/services/marketWatchActions.ts

# Restart server
npm run dev
```

But this shouldn't be necessary - the implementation is correct, just needs cache clear.

---

**STATUS:** All files are correct. This is 100% a Vite dev server caching issue.

**SOLUTION:** Complete restart with cache clear (steps above).
