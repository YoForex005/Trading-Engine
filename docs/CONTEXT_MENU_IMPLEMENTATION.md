# Context Menu Implementation Summary

## Overview
Implemented a desktop-grade, portal-based context menu system for RTX Market Watch that never clips or gets cut off at viewport edges.

## Architecture

### Core Components

#### 1. `ContextMenu.tsx` - Portal-Based Menu System
**Location**: `clients/desktop/src/components/ui/ContextMenu.tsx`

**Key Features**:
- **React Portal Rendering**: All menus render at `document.body` level, escaping parent container constraints
- **Viewport Collision Detection**: Smart positioning that flips horizontally/vertically to stay on screen
- **Recursive Submenu Support**: Infinite nesting with proper z-index management
- **Menu Stack Management**: Automatic cascade close when parent menus close
- **Keyboard Support**: Escape key closes all menus
- **Click Outside Detection**: Closes menu when clicking outside menu bounds

**Core Functions**:
```typescript
// Main menu position calculation
calculateMenuPosition(
  triggerX: number,
  triggerY: number,
  menuWidth: number,
  menuHeight: number,
  viewportWidth: number,
  viewportHeight: number
): Position

// Submenu collision detection with flip logic
calculateSubmenuPosition(
  triggerRect: DOMRect,
  menuWidth: number,
  menuHeight: number,
  viewportWidth: number,
  viewportHeight: number
): Position
```

**Components**:
- `ContextMenu` - Main portal-based menu container
- `Menu` - Recursive menu renderer with submenu support
- `MenuItem` - Individual menu item with hover/click logic
- `SubmenuPortal` - Portal-wrapped submenu with collision detection
- `MenuSectionHeader` - Utility for section headers
- `MenuDivider` - Utility for dividers

#### 2. `useContextMenu` Hook
**Location**: `clients/desktop/src/hooks/useContextMenu.ts`

**Purpose**: Simplified state management for context menu positioning and data

**API**:
```typescript
const contextMenu = useContextMenu();

// Open menu at position with optional data
contextMenu.open(x: number, y: number, data?: any)

// Close menu
contextMenu.close()

// Convenient handler for React events
contextMenu.handleContextMenu(e: React.MouseEvent, data?: any)

// Access state
contextMenu.state.isOpen // boolean
contextMenu.state.position // { x, y }
contextMenu.state.data // any
```

### Integration with MarketWatchPanel

**Updated**: `clients/desktop/src/components/layout/MarketWatchPanel.tsx`

**Changes**:
1. Replaced inline context menu with portal-based `ContextMenu` component
2. Removed old `ContextMenuItem` and `MenuDivider` components (now imported)
3. Added `useContextMenu` hook for state management
4. Created `menuItems` configuration using `useMemo` for performance
5. Updated all action handlers to use `contextMenu.state.data` and `contextMenu.close()`

**Menu Configuration Example**:
```typescript
const menuItems: ContextMenuItemConfig[] = useMemo(() => [
  { label: 'New Order', shortcut: 'F9', action: handleNewOrder },
  { divider: true },
  {
    label: 'Sort',
    submenu: [
      { label: 'Symbol', checked: sortBy === 'symbol', action: () => setSortBy('symbol') },
      { label: 'Gainers', checked: sortBy === 'gainers', action: () => setSortBy('gainers') }
    ]
  }
], [dependencies]);
```

## Technical Details

### Viewport Collision Detection Algorithm

**Horizontal Flip Logic**:
```typescript
// For submenus, default position is to the right
let x = triggerRect.right;

// If menu would overflow right edge, flip to left
if (x + menuWidth > viewportWidth - 8) {
  x = triggerRect.left - menuWidth;
}
```

**Vertical Adjustment**:
```typescript
let y = triggerRect.top;

// If menu would overflow bottom, adjust upward
if (y + menuHeight > viewportHeight - 8) {
  y = Math.max(8, viewportHeight - menuHeight - 8);
}
```

**Safety Padding**: 8px minimum from all edges

### Z-Index Management

Menus use incremental z-index for proper layering:
- Main menu: `z-index: 9999`
- First submenu: `z-index: 10000`
- Second submenu: `z-index: 10001`
- And so on...

### Position Calculation Timing

Uses `useLayoutEffect` instead of `useEffect` to calculate positions **before** the browser paints, preventing visible jumps or flickers:

```typescript
useLayoutEffect(() => {
  if (menuRef.current) {
    const rect = menuRef.current.getBoundingClientRect();
    const newPos = calculateSubmenuPosition(/* ... */);
    setPosition(newPos);
  }
}, [triggerRect]);
```

## Hard Requirements Met

✅ **NO overflow:hidden on menu containers** - Menus render at document.body level
✅ **NO nesting submenus inside table rows** - All menus are siblings at document.body
✅ **Menus rendered as siblings at document.body** - Using React Portal
✅ **Position calculated AFTER render** - Using useLayoutEffect
✅ **Full viewport awareness** - Collision detection on all edges

## Close Behavior

1. **Click Outside** → Close all menus
2. **Escape Key** → Close all menus
3. **Parent Menu Close** → Cascade to children (automatic via component unmounting)
4. **Menu Item Click** → Close all (unless `autoClose: false`)

## Performance Optimizations

1. **useMemo for menu items** - Prevents unnecessary recalculation
2. **useCallback for handlers** - Prevents child re-renders
3. **Portal rendering** - Menus don't trigger parent re-renders
4. **useLayoutEffect** - Single synchronous layout calculation

## Files Modified

1. `clients/desktop/src/components/ui/ContextMenu.tsx` - **NEW** (complete rewrite)
2. `clients/desktop/src/hooks/useContextMenu.ts` - **NEW**
3. `clients/desktop/src/hooks/index.ts` - Added exports
4. `clients/desktop/src/components/layout/MarketWatchPanel.tsx` - Integrated new system

## Testing Recommendations

1. **Edge Cases**:
   - Right-click near right viewport edge → submenu should flip left
   - Right-click near bottom edge → submenu should adjust upward
   - Deeply nested submenus (3+ levels)
   - Rapid context menu opening/closing

2. **Interaction**:
   - Click outside menu → should close
   - Press Escape → should close
   - Hover over menu items with submenus
   - Click on checked items

3. **Viewport Sizes**:
   - Small viewport (800x600)
   - Large viewport (1920x1080)
   - Narrow viewports where menus must flip horizontally

## Future Enhancements

1. **Keyboard Navigation**: Arrow keys, Enter, Tab
2. **Touch Support**: Long-press to open context menu on mobile
3. **Animation**: Smooth fade-in/out transitions
4. **Accessibility**: ARIA roles, focus management
5. **Virtual Scrolling**: For very long menus (100+ items)
6. **Search Filter**: Searchable menu items for large menus
