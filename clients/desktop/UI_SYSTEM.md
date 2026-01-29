# Professional Trading Terminal UI System

## Overview

This UI system provides a comprehensive set of components specifically designed for professional trading terminals. All components follow trading industry standards with proper color coding, animations, and accessibility.

## Design System

### Color Palette

```css
/* Trading Colors */
--color-profit: #00C853      /* Gains, buy orders */
--color-loss: #FF5252        /* Losses, sell orders */

/* Accent Colors */
--color-primary: #2196F3     /* Focus states, primary actions */
--color-warning: #FFA726     /* Warnings, pending states */
--color-info: #26C6DA        /* Information, neutral */

/* Background Hierarchy */
--bg-primary: #09090b        /* Main background */
--bg-secondary: #1E1E1E      /* Panels, cards */
--bg-tertiary: #2D2D2D       /* Hover states */
--bg-quaternary: #3D3D3D     /* Active states */
```

### Typography

```css
/* Font Families */
--font-sans: 'Inter'         /* UI text */
--font-mono: 'Roboto Mono'   /* Numbers, prices, code */

/* Font Sizes */
--text-2xs: 10px
--text-xs: 12px
--text-sm: 14px
--text-base: 16px
--text-lg: 18px
```

## Components

### 1. Sparkline

Micro line charts for showing price trends.

```tsx
import { Sparkline } from './components/ui';

<Sparkline
  data={[1.084, 1.085, 1.083, 1.086]}
  width={60}
  height={24}
  color="profit"  // 'profit' | 'loss' | 'neutral'
  showFill={true}
/>
```

**Features:**
- Automatic trend detection
- Color-coded (profit/loss)
- SVG-based for crisp rendering
- Minimal performance impact

### 2. FlashPrice

Animated price display with flash effect on updates.

```tsx
import { FlashPrice } from './components/ui';

<FlashPrice
  value={1.0850}
  direction="up"  // 'up' | 'down' | 'none'
  flashDuration={200}
  className="text-2xl font-bold"
/>
```

**Features:**
- Flash animation on price change
- Automatic direction detection
- Color-coded text
- Monospace font for alignment

### 3. StatusIndicator

Status badges with pulse animation.

```tsx
import { StatusIndicator } from './components/ui';

<StatusIndicator
  status="connected"  // 'connected' | 'disconnected' | 'pending' | 'idle'
  label="Trading Server"
  showPulse={true}
  size="md"  // 'sm' | 'md' | 'lg'
/>
```

**Features:**
- Animated pulse for active states
- Color-coded by status
- Optional label
- Three size variants

### 4. ProgressBar

Progress indicators for margin, risk, etc.

```tsx
import { ProgressBar } from './components/ui';

<ProgressBar
  value={75}
  max={100}
  variant="primary"  // 'primary' | 'success' | 'danger' | 'warning'
  showLabel={true}
  label="Margin Usage"
/>
```

**Features:**
- Auto color-coding based on thresholds
- Smooth transitions
- Optional label and percentage display
- Three size variants

### 5. HeatMapCell

Color-coded cells for performance visualization.

```tsx
import { HeatMapCell } from './components/ui';

<HeatMapCell
  value={45.5}
  min={-100}
  max={100}
  format={(v) => `${v.toFixed(2)}%`}
/>
```

**Features:**
- Intensity-based coloring
- Custom value formatting
- Automatic profit/loss detection
- Gradient backgrounds

### 6. Badge

Status and category badges.

```tsx
import { Badge } from './components/ui';

<Badge variant="profit" size="md">
  PROFIT
</Badge>
```

**Variants:** profit, loss, neutral, primary, warning, info
**Sizes:** sm, md, lg

### 7. Tooltip

Contextual help tooltips.

```tsx
import { Tooltip } from './components/ui';

<Tooltip content="Click to buy at market" position="top">
  <button>Buy</button>
</Tooltip>
```

**Features:**
- Four positions (top, bottom, left, right)
- Configurable delay
- Auto-positioning to stay on screen

### 8. ContextMenu

Right-click context menus.

```tsx
import { ContextMenu } from './components/ui';

<ContextMenu
  items={[
    { label: 'Buy', onClick: () => {}, icon: <Icon /> },
    { divider: true },
    { label: 'Close', onClick: () => {} },
  ]}
>
  <div>Right-click me</div>
</ContextMenu>
```

**Features:**
- Icon support
- Keyboard shortcuts display
- Dividers
- Disabled states
- Auto-positioning

### 9. ToggleSwitch

Professional toggle switches.

```tsx
import { ToggleSwitch } from './components/ui';

<ToggleSwitch
  checked={enabled}
  onChange={setEnabled}
  label="Auto-Trade"
  size="md"
/>
```

**Sizes:** sm, md, lg

### 10. Slider

Range input sliders.

```tsx
import { Slider } from './components/ui';

<Slider
  value={0.5}
  onChange={setValue}
  min={0}
  max={1}
  step={0.01}
  label="Volume"
  showValue={true}
/>
```

**Features:**
- Smooth dragging
- Custom formatting
- Min/max labels
- Value display

### 11. DataTable

Sortable data tables with virtualization support.

```tsx
import { DataTable } from './components/ui';

<DataTable
  data={positions}
  keyExtractor={(row) => row.id}
  columns={[
    { key: 'symbol', header: 'Symbol', sortable: true },
    { key: 'pnl', header: 'P&L', render: (v) => formatPnL(v) }
  ]}
  onRowClick={(row) => selectPosition(row)}
  stickyHeader={true}
/>
```

**Features:**
- Sortable columns
- Custom cell rendering
- Sticky headers
- Row click handlers
- Max height scrolling

### 12. ResizablePanel

Draggable panel resize.

```tsx
import { ResizablePanel } from './components/ui';

<ResizablePanel
  defaultSize={300}
  minSize={200}
  maxSize={500}
  direction="horizontal"
  onResize={(size) => saveLayout(size)}
>
  <div>Panel content</div>
</ResizablePanel>
```

**Features:**
- Horizontal/vertical resize
- Min/max constraints
- Resize callback
- Visual handle

### 13. CollapsiblePanel

Expandable panels with header.

```tsx
import { CollapsiblePanel } from './components/ui';

<CollapsiblePanel
  title="Advanced Settings"
  icon={<Icon />}
  defaultOpen={false}
  actions={<button>Reset</button>}
>
  <div>Panel content</div>
</CollapsiblePanel>
```

**Features:**
- Animated expand/collapse
- Custom header icon
- Action buttons
- Default state

### 14. LoadingSkeleton

Loading state placeholders.

```tsx
import { LoadingSkeleton, TableSkeleton, CardSkeleton } from './components/ui';

<LoadingSkeleton type="text" count={3} />
<TableSkeleton rows={5} columns={4} />
<CardSkeleton count={2} />
```

**Types:** text, card, row, circle, custom
**Features:** Shimmer animation, customizable dimensions

### 15. KeyboardShortcuts

Global keyboard shortcuts.

```tsx
import { KeyboardShortcuts, KeyboardShortcutsModal } from './components/ui';

const shortcuts = [
  { key: 'F1', description: 'Buy', action: () => buy(), category: 'Trading' },
  { key: 'Ctrl+W', description: 'Close', action: () => close(), category: 'Trading' }
];

<KeyboardShortcuts shortcuts={shortcuts} enabled={true} />
<KeyboardShortcutsModal shortcuts={shortcuts} onClose={() => {}} />
```

**Features:**
- Global keyboard listeners
- Help modal (triggered by '?')
- Categorized shortcuts
- Visual keyboard indicators

## Animations

### Flash Animations

```tsx
// Price updates
className="animate-flash-profit"
className="animate-flash-loss"

// Glow effect
className="animate-pulse-glow"
```

### Loading States

```tsx
// Skeleton shimmer
className="skeleton"

// Pulse
className="animate-pulse"
```

## Layout Utilities

### Grid System

```tsx
// Auto-fit columns
className="grid grid-auto-fit gap-4"

// Auto-fill columns
className="grid grid-auto-fill gap-4"
```

### Flex Utilities

```tsx
// Trading terminal layout
className="flex h-screen"

// Panel with resizer
className="flex-1 min-w-0"
```

## Accessibility

All components include:

- **ARIA labels**: Proper labeling for screen readers
- **Keyboard navigation**: Full keyboard support
- **Focus states**: Visible focus indicators
- **Color contrast**: WCAG AA compliant
- **Semantic HTML**: Proper element usage

## Performance Optimization

### Virtualization

For large lists, use the DataTable with max height:

```tsx
<DataTable
  data={largeDataset}
  maxHeight="400px"  // Enables virtual scrolling
  {...props}
/>
```

### Debouncing

Search inputs and resize handlers are automatically debounced.

### Memoization

All computation-heavy components use `useMemo` and `useCallback`.

## Best Practices

### 1. Color Usage

- **Profit/Loss**: Always use semantic colors (#00C853 for profit, #FF5252 for loss)
- **Neutrality**: Use zinc colors for non-directional data
- **Hierarchy**: Use background layers to create visual depth

### 2. Typography

- **Numbers**: Always use monospace font for alignment
- **Labels**: Use sans-serif for readability
- **Size hierarchy**: Follow the type scale (2xs → 2xl)

### 3. Spacing

- **Consistent gaps**: Use the spacing scale (1 → 8)
- **Padding**: Match padding to content importance
- **Margins**: Use margin for component separation

### 4. Responsive Design

```tsx
// Mobile-first approach
className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3"

// Responsive text
className="text-sm md:text-base lg:text-lg"
```

### 5. Dark Theme

All components are optimized for dark theme:
- High contrast ratios
- Reduced eye strain
- Professional appearance

## Examples

### Complete Trading Panel

```tsx
import {
  ResizablePanel,
  DataTable,
  FlashPrice,
  Sparkline,
  Badge,
  ProgressBar
} from './components/ui';

function TradingPanel() {
  return (
    <ResizablePanel defaultSize={400}>
      <div className="panel p-4">
        <div className="flex items-center justify-between mb-4">
          <FlashPrice value={1.0850} direction="up" />
          <Badge variant="profit">LONG</Badge>
        </div>

        <Sparkline data={priceHistory} color="profit" />

        <ProgressBar
          value={marginUsed}
          label="Margin"
          showLabel
        />

        <DataTable
          data={positions}
          columns={columns}
          stickyHeader
          maxHeight="300px"
        />
      </div>
    </ResizablePanel>
  );
}
```

## Integration with Existing Code

All components can be integrated into the existing trading terminal:

1. Import from `./components/ui`
2. Replace existing elements with UI components
3. Apply consistent theme variables
4. Add keyboard shortcuts for efficiency

## Future Enhancements

- [ ] Drag-and-drop layouts
- [ ] Chart overlays
- [ ] Multi-theme support
- [ ] Advanced filtering
- [ ] Export functionality
- [ ] Mobile optimization
- [ ] Virtual scrolling for tables
- [ ] Real-time data binding hooks

## Support

For issues or feature requests, refer to the component source files in `/src/components/ui/`.
