# One-Click Trading Panel - Quick Reference

## Quick Start

### Activate Panel
- **Keyboard:** `Ctrl+Shift+T`
- **Menu:** Right-click chart → "One-Click Trading"

### Place Order
1. Set volume (lots)
2. Click **BUY** (blue) or **SELL** (red)
3. Order executes immediately!

## Panel Layout

```
┌─────────────────────────────┐
│ One-Click                 × │  ← Header
├─────────────────────────────┤
│ Symbol:        EURUSD       │  ← Info
│ Spread:        2.0 pips     │
├─────────────────────────────┤
│ [-]   0.01   [+]            │  ← Volume
│        Lots                 │
├─────────────────────────────┤
│  SELL       │      BUY      │  ← Order Buttons
│  1.08500    │    1.08520    │
└─────────────────────────────┘
```

## Controls

### Volume Adjustment
- **Minus (-):** Decrease by 0.01 lots
- **Plus (+):** Increase by 0.01 lots
- **Input:** Type custom value
- **Minimum:** 0.01 lots

### Order Buttons
- **SELL (Red):** Market sell at bid price
- **BUY (Blue):** Market buy at ask price
- **No confirmation** - Instant execution!

### Close Panel
- **X button** in header
- **Keyboard:** `Ctrl+Shift+T` again
- **Menu:** Uncheck in context menu

## Safety Features

### First Use
- Disclaimer appears on first activation
- Must accept to continue
- Won't show again after acceptance

### Auto SL/TP
If configured in settings:
- Stop Loss automatically set
- Take Profit automatically set
- Based on default pips

## Display Information

### Symbol Info
- Current trading pair
- Real-time bid price (red)
- Real-time ask price (blue)
- Current spread in pips

### Status Indicators
- **"Placing order...":** Order being sent
- **Red error:** Order failed (auto-dismisses)
- **Disabled buttons:** Already placing order

## Tips & Tricks

### Fast Trading
1. Keep panel open for active trading
2. Use keyboard shortcut for quick toggle
3. Pre-set your common volume
4. Watch spread before clicking

### Risk Management
1. Set default SL/TP in settings
2. Start with small volumes
3. Check account balance first
4. Monitor open positions

### Keyboard Workflow
```
Ctrl+Shift+T  → Open panel
Type volume   → 0.05
Enter         → Confirm input
Click BUY     → Execute
Ctrl+Shift+T  → Close panel
```

## Common Mistakes to Avoid

❌ **Don't:** Click BUY/SELL multiple times
✅ **Do:** Wait for "Placing order..." to clear

❌ **Don't:** Use during news releases without checking spread
✅ **Do:** Monitor spread indicator before trading

❌ **Don't:** Forget panel is open when not trading
✅ **Do:** Close panel when done trading

❌ **Don't:** Ignore error messages
✅ **Do:** Read errors and check connection

## Troubleshooting

### Panel Won't Open
- Check if logged in
- Verify chart is active
- Try keyboard shortcut

### Orders Not Executing
- Check WebSocket connection
- Verify account has sufficient balance
- Check backend server is running

### Prices Not Updating
- Check WebSocket indicator (green = connected)
- Refresh page if stuck
- Check network connection

## Configuration

### Set Default SL/TP
Open browser console and run:
```javascript
// Set default stop loss in pips
localStorage.setItem('defaultStopLoss', '20');

// Set default take profit in pips
localStorage.setItem('defaultTakeProfit', '30');

// Clear defaults
localStorage.removeItem('defaultStopLoss');
localStorage.removeItem('defaultTakeProfit');
```

### Reset Disclaimer
To see disclaimer again:
```javascript
localStorage.removeItem('oneClickTradingAccepted');
```

## Keyboard Shortcuts Summary

| Shortcut | Action |
|----------|--------|
| `Ctrl+Shift+T` | Toggle One-Click Panel |
| `Tab` | Navigate between controls |
| `Enter` | Confirm volume input |
| `Esc` | (Future) Close panel |

## Volume Presets Reference

| Micro | Mini | Standard | Large |
|-------|------|----------|-------|
| 0.01  | 0.1  | 1.0      | 5.0   |

*Use +/- buttons or type directly*

## Spread Guidelines

| Spread | Trading Condition |
|--------|-------------------|
| 0-2 pips | ✅ Excellent |
| 2-5 pips | ✅ Good |
| 5-10 pips | ⚠️ Fair |
| 10+ pips | ❌ Poor (avoid) |

## Order Flow

```
Click BUY/SELL
     ↓
Validate Volume
     ↓
Apply Default SL/TP (if set)
     ↓
Send to API
     ↓
Show Loading
     ↓
Success → Update Positions
Failure → Show Error
```

## Support

**Issue:** Panel not visible
**Fix:** `Ctrl+Shift+T` or check context menu

**Issue:** Can't set volume
**Fix:** Click input field first

**Issue:** Orders fail silently
**Fix:** Open browser console (F12) for errors

**Issue:** Prices frozen
**Fix:** Check WebSocket connection status

## Best Practices

### Before Trading
1. ✅ Check WebSocket connection
2. ✅ Verify spread is reasonable
3. ✅ Set appropriate volume
4. ✅ Ensure sufficient margin

### During Trading
1. ✅ Monitor positions list
2. ✅ Watch spread changes
3. ✅ Adjust volume as needed
4. ✅ Use with caution during news

### After Trading
1. ✅ Close panel if done
2. ✅ Review executed orders
3. ✅ Check account balance
4. ✅ Manage open positions

## Advanced Usage

### Rapid Scalping
```
1. Set volume to 0.01
2. Keep panel open
3. Watch price action
4. Quick click BUY/SELL
5. Monitor ticks closely
```

### Swing Trading
```
1. Set larger volume (0.5-1.0)
2. Configure default SL/TP
3. Open panel when ready
4. Execute at key levels
5. Close panel after entry
```

## Safety Checklist

Before enabling one-click trading:

- [ ] I understand orders execute immediately
- [ ] I have checked my account balance
- [ ] I have set appropriate default SL/TP
- [ ] I have tested with small volumes first
- [ ] I understand the risks involved
- [ ] I am ready for instant execution

## Remember

> ⚠️ **One-Click Trading = Instant Execution**
>
> No confirmation dialogs!
> No undo button!
> Trade carefully!

## Quick Help

| Question | Answer |
|----------|--------|
| How to open? | `Ctrl+Shift+T` |
| Minimum volume? | 0.01 lots |
| Auto SL/TP? | Set in localStorage |
| How to close? | Click X or `Ctrl+Shift+T` |
| Confirmation? | None - instant! |

---

**Last Updated:** 2026-02-11
**Version:** 1.0
**Component:** OneClickPanel
