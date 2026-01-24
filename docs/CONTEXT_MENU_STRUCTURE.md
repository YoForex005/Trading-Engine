# Context Menu Structure - Visual Reference

## Menu Hierarchy

```
RTX Market Watch Context Menu
├── [Symbol: EURUSD] ← Dynamic header
├─┬ Trading Actions
│ ├── New Order (F9)
│ ├── Quick Buy (0.01)
│ ├── Quick Sell (0.01)
│ ├── Chart Window
│ ├── Tick Chart
│ ├── Depth of Market (Alt+B)
│ └── Popup Prices (F10)
├─┬ Visibility
│ ├── Hide (Delete)
│ └── Show All
├─┬ Configuration
│ ├── Symbols (Ctrl+U)
│ ├─┬ Sets →
│ │ ├── forex.all
│ │ ├── forex.major
│ │ ├── forex.crosses
│ │ ├── ─────────
│ │ ├── 🕐 Save as...
│ │ └─┬ Remove →
│ │   └── (empty)
│ ├─┬ Sort →
│ │ ├── ☑ Symbol
│ │ ├── ☐ Gainers
│ │ ├── ☐ Losers
│ │ ├── ☐ Volume
│ │ ├── ─────────
│ │ └── Reset
│ └── Export
├─┬ System Options
│ ├── ☑ Use System Colors
│ ├── ☐ Show Milliseconds
│ ├── ☑ Auto Remove Expired
│ ├── ☑ Auto Arrange
│ └── ☑ Grid
└─┬ Columns →
  ├── ☐ Daily %
  ├── ☐ Last
  ├── ☐ High
  ├── ☐ Low
  ├── ☐ Vol
  └── ☐ Time
```

## Rendering Flow

```
User Right-Clicks Row
         ↓
handleContextMenuOpen(e, symbol)
         ↓
contextMenu.open(x, y, symbol)
         ↓
State Updated: { isOpen: true, position: { x, y }, data: symbol }
         ↓
ContextMenu Component Renders via Portal
         ↓
         ├→ createPortal(menu, document.body)
         │           ↓
         │    Menu renders at body level
         │    (escapes parent constraints)
         │           ↓
         │    useLayoutEffect calculates position
         │           ↓
         │    calculateMenuPosition() checks viewport
         │           ↓
         │    Position adjusted to avoid clipping
         │
         └→ User hovers over "Sort" item
                    ↓
             onMouseEnter triggers
                    ↓
             Submenu state updated
                    ↓
             SubmenuPortal renders
                    ↓
             useLayoutEffect calculates submenu position
                    ↓
             calculateSubmenuPosition() checks viewport
                    ↓
             Submenu flips left if would overflow right
                    ↓
             Submenu adjusts up if would overflow bottom
```

## Collision Detection Examples

### Scenario 1: Right-Click Near Right Edge
```
Viewport: [0────────────────────1920px]
                               │
                         Click │ X (1850, 300)
                               │
                  Main Menu    │
                  ┌─────────┐  │
                  │ Sort  → │──┼─ Would overflow!
                  └─────────┘  │
                               │
                          Submenu would be at x=2010
                          (off screen by 90px)
                               │
                     FLIP LEFT │
                               │
              ┌─────────┐      │
              │ Symbol  │      │
          ┌───│ Gainers │      │
          │   │ Losers  │      │
    Submenu   │ Volume  │      │
    appears   └─────────┘      │
    to LEFT                    │
```

### Scenario 2: Right-Click Near Bottom Edge
```
Viewport Height: 1080px





                    Click X (500, 1000)
                    ┌──────────────┐
                    │ Sort       → │
                    └──────────────┘
                           │
                    Submenu would start at y=1000
                           │
                    ┌─────────────┐
                    │ Symbol      │
                    │ Gainers     │
                    │ Losers      │
                    │ Volume      │ ← Would overflow bottom!
                    └─────────────┘

    ADJUSTED UPWARD
                    ┌─────────────┐
                    │ Symbol      │
                    │ Gainers     │ ← Adjusted to y=920
                    │ Losers      │   (160px height)
                    │ Volume      │
                    └─────────────┘ ← Bottom at y=1080 (perfect!)
```

## Component Tree (Runtime)

```
document.body
└─┬ <div> (Portal root from ContextMenu)
  ├── <div ref={menuRef}>
  │     └─┬ <Menu position={x,y} zIndex={9999}>
  │       ├── <div className="fixed..."> (Main menu)
  │       │     ├── <MenuItem> "New Order"
  │       │     ├── <MenuItem> "Quick Buy"
  │       │     ├── <MenuItem submenu> "Sets"
  │       │     ├── <MenuItem submenu> "Sort"
  │       │     └── ...
  │       │
  │       └─┬ {activeSubmenu && <SubmenuPortal>}
  │         └─┬ <div ref={submenuRef}> (Submenu wrapper)
  │           └─┬ <Menu position={x2,y2} zIndex={10000}>
  │             ├── <div className="fixed..."> (Submenu)
  │             │     ├── <MenuItem> "Symbol"
  │             │     ├── <MenuItem> "Gainers"
  │             │     └── ...
  │             │
  │             └─┬ {activeSubmenu2 && <SubmenuPortal>}
  │               └── ... (3rd level submenu)
  └── (No overflow:hidden anywhere in the tree!)
```

## State Flow Diagram

```
┌────────────────────────────────────────────────┐
│  useContextMenu Hook State                     │
├────────────────────────────────────────────────┤
│  state: {                                      │
│    isOpen: boolean                             │
│    position: { x: number, y: number }         │
│    data: any (symbol string)                  │
│  }                                             │
└──────────────────┬─────────────────────────────┘
                   │
                   ▼
    ┌──────────────────────────────┐
    │  ContextMenu Component       │
    ├──────────────────────────────┤
    │  props: {                    │
    │    items: ContextMenuItemConfig[] │
    │    onClose: () => void       │
    │    position: { x, y }        │
    │    triggerSymbol?: string    │
    │  }                           │
    └──────────┬───────────────────┘
               │
               ▼
    ┌──────────────────────────────┐
    │  Menu Component              │
    ├──────────────────────────────┤
    │  Local State:                │
    │    adjustedPosition          │
    │    activeSubmenu             │
    └──────────┬───────────────────┘
               │
               ├──────────────────────┐
               │                      │
               ▼                      ▼
    ┌─────────────────┐   ┌──────────────────┐
    │  MenuItem       │   │  SubmenuPortal   │
    │  (no submenu)   │   │  (with submenu)  │
    └─────────────────┘   └──────────────────┘
```

## Event Handlers

### Click Flow
```
User clicks menu item
       ↓
MenuItem.onClick
       ↓
e.stopPropagation() ← Prevent bubbling to document
       ↓
item.action() ← Execute user-defined action
       ↓
onClose() ← Close entire menu tree
       ↓
Portal unmounts
```

### Hover Flow (Submenu)
```
User hovers "Sort" item
       ↓
MenuItem.onMouseEnter
       ↓
onSubmenuOpen(items, rect)
       ↓
Parent Menu sets activeSubmenu state
       ↓
SubmenuPortal renders
       ↓
useLayoutEffect calculates position
       ↓
Submenu appears (flipped if needed)
```

### Close Flow
```
Click Outside
       ↓
document.mousedown listener
       ↓
Check if click is outside menuRef
       ↓
onClose() called
       ↓
contextMenu.close()
       ↓
State: { isOpen: false }
       ↓
Portal unmounts
```

## Performance Characteristics

| Operation | Time Complexity | Notes |
|-----------|----------------|-------|
| Menu Open | O(1) | Simple state update |
| Position Calculation | O(1) | Single getBoundingClientRect + math |
| Submenu Open | O(1) | State update triggers new portal |
| Menu Item Render | O(n) | n = number of items (memoized) |
| Close All | O(1) | Single state update cascades |

## Browser Compatibility

✅ Chrome 90+
✅ Firefox 88+
✅ Safari 14+
✅ Edge 90+
❌ IE11 (not supported - uses React Portals and modern CSS)
