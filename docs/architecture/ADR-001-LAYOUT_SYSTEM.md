# ADR-001: Layout System Architecture

**Status:** Approved
**Date:** 2026-01-18
**Deciders:** System Architecture Designer, Development Team
**Context:** Need for professional drag-drop multi-panel layout system

## Context and Problem Statement

The current trading platform uses a fixed 3-column layout that cannot be customized by users. Professional traders require:

- Customizable panel positioning
- Drag-and-drop panel arrangement
- Resizable panels
- Saved workspace layouts
- Multi-monitor support
- Responsive behavior across devices

**Question:** Which layout framework should we use for the professional trading terminal?

## Decision Drivers

1. **Flexibility:** Must support arbitrary panel arrangements
2. **Performance:** Smooth drag/resize with real-time data
3. **Persistence:** Save/restore layouts across sessions
4. **Responsive:** Adapt to different screen sizes
5. **Maintainability:** Active library with good documentation
6. **Bundle Size:** Minimal impact on bundle size
7. **TypeScript Support:** First-class TypeScript support

## Considered Options

### Option 1: React Grid Layout

**Pros:**
- Mature library (10k+ stars, 8+ years active)
- Excellent drag-drop and resize support
- Breakpoint-responsive layouts
- Grid-based positioning (easy to reason about)
- Small bundle size (~50KB)
- Good TypeScript support
- Active maintenance

**Cons:**
- Grid-based constraints (not free-form)
- Requires layout data structure
- CSS-in-JS or external CSS needed

**Example:**
```tsx
import GridLayout from 'react-grid-layout';

<GridLayout
  className="layout"
  layout={layout}
  cols={12}
  rowHeight={30}
  width={1200}
  onLayoutChange={onLayoutChange}
  draggableHandle=".drag-handle"
  resizeHandles={['se', 'sw', 'ne', 'nw']}
>
  <div key="market-watch" data-grid={{x: 0, y: 0, w: 3, h: 10}}>
    <MarketWatchPanel />
  </div>
  <div key="chart" data-grid={{x: 3, y: 0, w: 6, h: 10}}>
    <ChartPanel />
  </div>
  {/* More panels */}
</GridLayout>
```

### Option 2: Golden Layout

**Pros:**
- Feature-rich (tabs, stacking, popping out windows)
- Professional appearance (like IDEs)
- Very flexible panel management
- Multi-window support out-of-box

**Cons:**
- Heavy bundle (~300KB)
- Complex API (steep learning curve)
- Not React-native (requires wrapper)
- Less active maintenance
- jQuery dependency (legacy)

**Example:**
```tsx
const config = {
  content: [{
    type: 'row',
    content: [{
      type: 'component',
      componentName: 'MarketWatch',
      componentState: { symbol: 'EURUSD' }
    }, {
      type: 'column',
      content: [{
        type: 'component',
        componentName: 'Chart'
      }]
    }]
  }]
};

const layout = new GoldenLayout(config, container);
```

### Option 3: react-mosaic

**Pros:**
- Window tiling (like VSCode)
- Split views and tabs
- Minimalist API
- Good TypeScript support

**Cons:**
- Limited to mosaic/tiling patterns
- No free-form drag-drop
- Smaller community
- Less flexible than grid

### Option 4: Custom CSS Grid Solution

**Pros:**
- Full control over implementation
- No library dependencies
- Lightest weight
- Tailored to exact needs

**Cons:**
- Significant development time (2-3 weeks)
- Complex drag-drop implementation
- Maintenance burden
- Potential bugs

## Decision Outcome

**Chosen Option:** React Grid Layout

**Rationale:**

1. **Best Balance:** Good feature set without excessive complexity
2. **Performance:** Grid-based approach is performant with many panels
3. **React-First:** Built for React, not a wrapper
4. **Active Maintenance:** Regular updates, large community
5. **Bundle Size:** Acceptable at ~50KB gzipped
6. **Proven:** Used by many dashboard applications
7. **TypeScript:** Good type definitions

**Trade-offs Accepted:**
- Grid-based constraints vs. free-form positioning
  - **Mitigation:** 12-column grid provides sufficient granularity
- External CSS required
  - **Mitigation:** Minimal CSS, integrates with Tailwind

## Implementation Details

### Layout Data Structure

```typescript
interface PanelLayout {
  i: string;           // Unique panel ID
  x: number;           // Grid column position (0-11)
  y: number;           // Grid row position
  w: number;           // Width in grid units (1-12)
  h: number;           // Height in grid units
  minW?: number;       // Minimum width
  minH?: number;       // Minimum height
  maxW?: number;       // Maximum width
  maxH?: number;       // Maximum height
  static?: boolean;    // Prevent drag/resize
  isDraggable?: boolean;
  isResizable?: boolean;
}

interface Workspace {
  id: string;
  name: string;
  layout: PanelLayout[];
  breakpoints: {
    lg: PanelLayout[];
    md: PanelLayout[];
    sm: PanelLayout[];
  };
}
```

### Responsive Breakpoints

```typescript
const breakpoints = {
  lg: 1200,  // Desktop
  md: 996,   // Tablet landscape
  sm: 768,   // Tablet portrait
  xs: 480,   // Mobile
  xxs: 0     // Small mobile
};

const cols = {
  lg: 12,
  md: 10,
  sm: 6,
  xs: 4,
  xxs: 2
};
```

### Workspace Presets

```typescript
const WORKSPACES: Record<string, Workspace> = {
  trading: {
    id: 'trading',
    name: 'Trading',
    layout: [
      { i: 'market-watch', x: 0, y: 0, w: 3, h: 8, minW: 2, minH: 4 },
      { i: 'chart', x: 3, y: 0, w: 6, h: 8, minW: 4, minH: 6 },
      { i: 'order-entry', x: 9, y: 0, w: 3, h: 4, minW: 2, minH: 3 },
      { i: 'positions', x: 9, y: 4, w: 3, h: 4, minW: 2, minH: 3 },
      { i: 'account', x: 0, y: 8, w: 12, h: 2, static: true },
    ],
    breakpoints: { /* Responsive layouts */ }
  },

  analysis: {
    id: 'analysis',
    name: 'Analysis',
    layout: [
      { i: 'chart-1', x: 0, y: 0, w: 6, h: 6 },
      { i: 'chart-2', x: 6, y: 0, w: 6, h: 6 },
      { i: 'chart-3', x: 0, y: 6, w: 6, h: 6 },
      { i: 'chart-4', x: 6, y: 6, w: 6, h: 6 },
    ],
    breakpoints: { /* Responsive layouts */ }
  },

  scalping: {
    id: 'scalping',
    name: 'Scalping',
    layout: [
      { i: 'orderbook', x: 0, y: 0, w: 3, h: 10 },
      { i: 'time-sales', x: 3, y: 0, w: 2, h: 10 },
      { i: 'chart', x: 5, y: 0, w: 5, h: 10 },
      { i: 'quick-order', x: 10, y: 0, w: 2, h: 5 },
      { i: 'positions', x: 10, y: 5, w: 2, h: 5 },
    ],
    breakpoints: { /* Responsive layouts */ }
  }
};
```

### Panel Component Structure

```typescript
interface PanelProps {
  id: string;
  title: string;
  closable?: boolean;
  collapsible?: boolean;
  settingsEnabled?: boolean;
  onClose?: () => void;
  onSettings?: () => void;
  children: React.ReactNode;
}

const Panel: React.FC<PanelProps> = ({
  id,
  title,
  closable = true,
  collapsible = false,
  settingsEnabled = false,
  onClose,
  onSettings,
  children
}) => {
  const [collapsed, setCollapsed] = useState(false);

  return (
    <div className="panel-container h-full flex flex-col bg-zinc-900 border border-zinc-800 rounded">
      {/* Header with drag handle */}
      <div className="drag-handle flex items-center justify-between px-3 py-2 bg-zinc-800/50 border-b border-zinc-700 cursor-move">
        <h3 className="text-xs font-semibold uppercase tracking-wide text-zinc-400">
          {title}
        </h3>
        <div className="flex items-center gap-1">
          {settingsEnabled && (
            <button onClick={onSettings} className="p-1 hover:bg-zinc-700 rounded">
              <Settings size={14} />
            </button>
          )}
          {collapsible && (
            <button onClick={() => setCollapsed(!collapsed)} className="p-1 hover:bg-zinc-700 rounded">
              {collapsed ? <ChevronDown size={14} /> : <ChevronUp size={14} />}
            </button>
          )}
          {closable && (
            <button onClick={onClose} className="p-1 hover:bg-red-500/20 rounded">
              <X size={14} />
            </button>
          )}
        </div>
      </div>

      {/* Content */}
      {!collapsed && (
        <div className="flex-1 overflow-auto p-2">
          {children}
        </div>
      )}
    </div>
  );
};
```

### Layout Persistence

```typescript
class LayoutManager {
  private static STORAGE_KEY = 'terminal-layouts';

  static saveLayout(workspaceId: string, layout: PanelLayout[]): void {
    const layouts = this.getLayouts();
    layouts[workspaceId] = layout;
    localStorage.setItem(this.STORAGE_KEY, JSON.stringify(layouts));
  }

  static getLayout(workspaceId: string): PanelLayout[] | null {
    const layouts = this.getLayouts();
    return layouts[workspaceId] || null;
  }

  static getLayouts(): Record<string, PanelLayout[]> {
    const data = localStorage.getItem(this.STORAGE_KEY);
    return data ? JSON.parse(data) : {};
  }

  static deleteLayout(workspaceId: string): void {
    const layouts = this.getLayouts();
    delete layouts[workspaceId];
    localStorage.setItem(this.STORAGE_KEY, JSON.stringify(layouts));
  }

  static exportLayouts(): string {
    return localStorage.getItem(this.STORAGE_KEY) || '{}';
  }

  static importLayouts(data: string): void {
    try {
      const layouts = JSON.parse(data);
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(layouts));
    } catch (e) {
      throw new Error('Invalid layout data');
    }
  }
}
```

## Consequences

### Positive

1. **User Productivity:** Traders can customize layouts to their workflow
2. **Professional UX:** Drag-drop interaction feels premium
3. **Multi-Workflow Support:** Different layouts for different trading styles
4. **Responsive Design:** Automatically adapts to screen sizes
5. **State Persistence:** Layouts survive page reloads
6. **Developer Experience:** React Grid Layout has good DX

### Negative

1. **Learning Curve:** Users need to learn layout customization
   - **Mitigation:** Provide tutorial on first launch
2. **Layout Complexity:** Grid data structure to maintain
   - **Mitigation:** LayoutManager abstraction
3. **Bundle Size:** +50KB to bundle
   - **Mitigation:** Code splitting, lazy loading
4. **Grid Constraints:** Limited to grid-based positioning
   - **Mitigation:** 12-column grid provides flexibility

### Neutral

1. **CSS Dependency:** Requires external CSS
2. **Breakpoint Management:** Need to define responsive layouts
3. **Panel Registry:** Need to maintain panel registry

## Validation

### Performance Benchmarks

**Target:** 60 FPS drag/resize with 12 panels
**Measurement:** Chrome DevTools Performance profiler

**Expected Results:**
- Initial render: < 100ms
- Drag operation: < 16ms per frame (60 FPS)
- Layout change: < 50ms
- Memory usage: < 50MB for layout system

### User Testing

**Scenarios:**
1. Can user drag panel to new position in < 5 seconds?
2. Can user resize panel to desired size in < 5 seconds?
3. Can user save and restore custom layout?
4. Does layout adapt correctly on tablet (breakpoint test)?

**Success Criteria:**
- 90% of users complete layout tasks successfully
- < 5 support tickets related to layout in first month

## Related Decisions

- ADR-002: State Management (Zustand store slicing)
- ADR-003: Panel Registry Pattern
- ADR-004: WebSocket Data Flow

## References

- [React Grid Layout GitHub](https://github.com/react-grid-layout/react-grid-layout)
- [React Grid Layout Demo](https://react-grid-layout.github.io/react-grid-layout/examples/0-showcase.html)
- [Golden Layout](https://golden-layout.com/)
- [react-mosaic](https://github.com/nomcopter/react-mosaic-component)

## Notes

**Migration Strategy:**
1. Week 1: Implement layout system with Trading preset
2. Week 2: Add Analysis and Scalping presets
3. Week 3: Add custom layout saving
4. Week 4: User testing and refinement

**Future Enhancements:**
- Layout templates marketplace
- Multi-monitor window popping (requires Electron)
- AI-suggested layouts based on trading patterns
- Layout sharing between users
