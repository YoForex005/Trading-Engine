# Keyboard Shortcuts Reference

**Trading Platform - Global Shortcuts**

---

## 📋 Quick Reference Table

| Shortcut | Action | Context | Status |
|----------|--------|---------|--------|
| **F9** | New Order | Global | ✅ Active |
| **F10** | Chart Window | Global | 🚧 Placeholder |
| **Alt+B** | Quick Buy | Global | ✅ Active |
| **Alt+S** | Quick Sell | Global | ✅ Active |
| **Ctrl+U** | Unsubscribe Symbol | Global | 🚧 Placeholder |
| **Esc** | Close Dialog | Order Panel | ✅ Active |
| **↑** | Navigate Up | Context Menu | ✅ Active |
| **↓** | Navigate Down | Context Menu | ✅ Active |
| **→** | Open Submenu | Context Menu | ✅ Active |
| **←** | Close Submenu | Context Menu | ✅ Active |
| **Enter** | Execute Action | Context Menu | ✅ Active |
| **Letter Keys** | Jump to Item | Context Menu | ✅ Active |

---

## 🎯 Global Shortcuts

### Trading Actions

#### F9 - New Order Dialog
- **Description**: Opens the order panel for the currently selected symbol
- **Requirements**: Symbol must be selected in Market Watch
- **Pre-fills**: Symbol, current Bid/Ask prices
- **Implementation**: Full

#### Alt+B - Quick Buy
- **Description**: Instantly executes a market BUY order
- **Volume**: Uses current volume setting from One-Click Trading panel
- **Confirmation**: None (instant execution)
- **Warning**: ⚠️ No confirmation dialog - order executes immediately
- **Implementation**: Full

#### Alt+S - Quick Sell
- **Description**: Instantly executes a market SELL order
- **Volume**: Uses current volume setting from One-Click Trading panel
- **Confirmation**: None (instant execution)
- **Warning**: ⚠️ No confirmation dialog - order executes immediately
- **Implementation**: Full

---

### Window Management

#### F10 - Chart Window
- **Description**: Opens chart window for selected symbol
- **Status**: 🚧 Placeholder (shows alert)
- **Future**: Will open new window/tab with chart
- **Implementation**: Partial (placeholder)

#### Ctrl+U - Unsubscribe Symbol
- **Description**: Unsubscribes from market data for current symbol
- **Status**: 🚧 Placeholder (shows alert)
- **Future**: Will remove symbol from WebSocket subscription
- **Implementation**: Partial (placeholder)

---

### Dialog Control

#### Esc - Close Dialog
- **Description**: Closes the currently open order dialog
- **Context**: Only works when order panel is visible
- **Implementation**: Full

---

## 🖱️ Context Menu Shortcuts

### Navigation

#### Arrow Keys (↑ ↓ → ←)
- **↑**: Move to previous menu item
- **↓**: Move to next menu item
- **→**: Open submenu (if item has submenu)
- **←**: Close submenu and return to parent

#### Enter / Space
- **Description**: Execute the currently focused menu item
- **Behavior**:
  - If item has submenu → Open submenu
  - If item has action → Execute action and close menu

#### Esc
- **Description**: Close menu or submenu
- **Behavior**:
  - If in submenu → Close submenu, return to parent
  - If in main menu → Close entire context menu

---

### First-Letter Navigation

#### Letter Keys (A-Z, 0-9)
- **Description**: Jump to first menu item starting with that letter
- **Case**: Insensitive (both 'c' and 'C' work)
- **Example**:
  - Press 'C' → Jumps to "Chart Window"
  - Press 'N' → Jumps to "New Order"
  - Press 'D' → Jumps to "Depth of Market"

---

## 🎮 Usage Examples

### Example 1: Quick Buy Flow
```
1. Select symbol in Market Watch (EURUSD)
2. Set volume in One-Click Trading panel (0.01)
3. Press Alt+B
4. ✅ BUY order executes immediately
5. Position appears in Positions tab
```

### Example 2: Context Menu Navigation
```
1. Right-click on EURUSD in Market Watch
2. Press ↓ to navigate to "New Order"
3. Press → to open submenu
4. Press ↓ to select "Buy Limit"
5. Press Enter to open order dialog
```

### Example 3: First-Letter Jump
```
1. Right-click on symbol
2. Press 'C' → Jumps to "Chart Window"
3. Press 'D' → Jumps to "Depth of Market"
4. Press 'N' → Jumps to "New Order"
```

---

## ⚙️ Configuration

### Current Settings
- **Hover Delay**: 300ms (MT5 standard)
- **Keyboard Sensitivity**: 7px (MT5 standard)
- **Auto-flip Padding**: 8px from screen edges
- **Z-index Base**: 9999 (menus), +10 per submenu level

### Customization (Future)
- User-defined keyboard shortcuts
- Configurable hover delay
- Custom key bindings

---

## 🐛 Troubleshooting

### Keyboard Shortcuts Not Working

**Problem**: Pressing F9 or Alt+B does nothing

**Solutions**:
1. Check that a symbol is selected in Market Watch
2. Verify browser focus is on the trading platform window
3. Check browser console for errors
4. Ensure no browser extension is intercepting shortcuts

---

### Context Menu Keyboard Navigation Issues

**Problem**: Arrow keys not working in context menu

**Solutions**:
1. Click on menu to ensure it has focus
2. Wait for menu to fully render (50ms delay)
3. Check that no other element has captured keyboard focus

---

### Quick Buy/Sell Executing Wrong Volume

**Problem**: Alt+B/Alt+S uses unexpected volume

**Solution**:
- Check One-Click Trading panel volume setting
- Volume is NOT taken from order dialog
- Adjust volume in One-Click panel before using shortcuts

---

## 📚 Related Documentation

- [UI Interaction Implementation Report](./UI_INTERACTION_IMPLEMENTATION_REPORT.md)
- [Context Menu Implementation](./CONTEXT_MENU_IMPLEMENTATION.md)
- [Keyboard Navigation Flow](./KEYBOARD_NAVIGATION_FLOW.txt)

---

## 🔜 Upcoming Features

### Planned Shortcuts
- `Ctrl+T`: Open Terminal (Bottom Dock)
- `Ctrl+M`: Focus Market Watch
- `Ctrl+Shift+P`: Command Palette
- `F11`: Fullscreen Chart
- `Ctrl+1-9`: Switch between chart tabs

### Planned Context Menu Features
- `Ctrl+Click`: Multi-select symbols
- `Shift+Click`: Range select
- `Del`: Remove from Market Watch
- `Ins`: Add to Market Watch

---

**Last Updated**: 2026-01-20
**Status**: Active Development
