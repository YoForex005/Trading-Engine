# Quick Integration Guide - Drawing Tools

**For**: Developer implementing UI integration
**Time**: 5 minutes to read, 2-3 days to implement
**Difficulty**: Medium

---

## What You're Doing

Adding 3 UI panels to TradingChart:
1. Properties panel (edit drawing color, width, style)
2. Drawing list (see all drawings, toggle visibility)
3. Template manager (save/load drawing sets)

**Source**: Extract from `DrawingTools.tsx`
**Target**: Add to `TradingChart.tsx`
**Connect**: Via `drawingManager` events

---

## Quick Start

### Step 1: Create DrawingPropertiesPanel.tsx

**Copy from**: `DrawingTools.tsx` lines 636-758 (properties panel)

**Key Changes**:
```typescript
// BEFORE (in DrawingTools.tsx)
import { useDrawingStore } from '../store/useDrawingStore';
const { selectedDrawing, updateDrawing } = useDrawingStore();

// AFTER (in DrawingPropertiesPanel.tsx)
import { drawingManager } from '../services/drawingManager';

interface Props {
  drawing: Drawing | null;
  onClose: () => void;
}

export function DrawingPropertiesPanel({ drawing, onClose }: Props) {
  if (!drawing) return null;

  const handleUpdate = (updates: Partial<Drawing>) => {
    // Update via drawingManager instead of store
    Object.assign(drawing, updates);
    drawingManager.saveToBackend(drawing);
    drawingManager.renderAllDrawings();
  };

  // Rest is same UI code
}
```

### Step 2: Create DrawingListPanel.tsx

**Copy from**: `DrawingTools.tsx` lines 575-633 (drawing list sidebar)

**Key Changes**:
```typescript
import { drawingManager, type Drawing } from '../services/drawingManager';

export function DrawingListPanel() {
  const [drawings, setDrawings] = useState<Drawing[]>([]);

  useEffect(() => {
    // Refresh on events
    const refresh = () => setDrawings(drawingManager.getDrawings());

    window.addEventListener('drawing:saved', refresh);
    window.addEventListener('drawing:deleted', refresh);
    window.addEventListener('drawing:visibility-changed', refresh);
    window.addEventListener('drawings:cleared', refresh);

    refresh();

    return () => {
      window.removeEventListener('drawing:saved', refresh);
      window.removeEventListener('drawing:deleted', refresh);
      window.removeEventListener('drawing:visibility-changed', refresh);
      window.removeEventListener('drawings:cleared', refresh);
    };
  }, []);

  return (
    <div className="w-64 bg-zinc-800 border-l border-zinc-700">
      {/* Same UI as DrawingTools.tsx */}
      {drawings.map(drawing => (
        <div key={drawing.id}>
          <span>{drawing.type}</span>
          <button onClick={() => drawingManager.toggleDrawingVisibility(drawing.id)}>
            {drawing.visible ? <Eye /> : <EyeOff />}
          </button>
          <button onClick={() => drawingManager.deleteDrawing(drawing.id)}>
            <Trash2 />
          </button>
        </div>
      ))}
    </div>
  );
}
```

### Step 3: Integrate into TradingChart.tsx

**Add to imports**:
```typescript
import { DrawingPropertiesPanel } from './DrawingPropertiesPanel';
import { DrawingListPanel } from './DrawingListPanel';
import type { Drawing } from '../services/drawingManager';
```

**Add state**:
```typescript
const [selectedDrawing, setSelectedDrawing] = useState<Drawing | null>(null);
const [showPropertiesPanel, setShowPropertiesPanel] = useState(false);
const [showDrawingList, setShowDrawingList] = useState(false);
```

**Add effect to track selection**:
```typescript
useEffect(() => {
  const handleDrawingSelection = () => {
    const drawings = drawingManager.getDrawings();
    const selected = drawings.find(d => d.selected);
    setSelectedDrawing(selected || null);
    if (selected) setShowPropertiesPanel(true);
  };

  // Listen to drawing events
  window.addEventListener('drawing:saved', handleDrawingSelection);
  window.addEventListener('drawing:deleted', handleDrawingSelection);

  return () => {
    window.removeEventListener('drawing:saved', handleDrawingSelection);
    window.removeEventListener('drawing:deleted', handleDrawingSelection);
  };
}, []);
```

**Add to JSX** (before closing div):
```typescript
return (
  <div className="relative w-full h-full">
    {/* Existing chart code */}

    {/* NEW: Properties Panel */}
    {showPropertiesPanel && selectedDrawing && (
      <DrawingPropertiesPanel
        drawing={selectedDrawing}
        onClose={() => setShowPropertiesPanel(false)}
      />
    )}

    {/* NEW: Drawing List (toggle with keyboard shortcut) */}
    {showDrawingList && <DrawingListPanel />}
  </div>
);
```

**Add keyboard shortcut** (in existing keyboard effect):
```typescript
useEffect(() => {
  const handleKeyDown = (e: KeyboardEvent) => {
    // Existing shortcuts...

    // NEW: Toggle drawing list (Ctrl+Shift+D)
    if (e.ctrlKey && e.shiftKey && e.key === 'D') {
      e.preventDefault();
      setShowDrawingList(prev => !prev);
    }

    // NEW: Close properties panel (Esc)
    if (e.key === 'Escape' && showPropertiesPanel) {
      e.preventDefault();
      setShowPropertiesPanel(false);
    }
  };

  window.addEventListener('keydown', handleKeyDown);
  return () => window.removeEventListener('keydown', handleKeyDown);
}, [showPropertiesPanel]);
```

---

## That's It! 🎉

You now have:
- ✅ Properties panel that opens when clicking a drawing
- ✅ Drawing list with visibility toggle
- ✅ All connected to existing drawingManager
- ✅ Keyboard shortcuts (Ctrl+Shift+D for list, Esc to close)

---

## Testing

```bash
# Start dev server
npm run dev

# Test flow:
1. Draw a trendline on chart
2. Click the trendline
3. Properties panel should open
4. Change color → trendline updates
5. Press Ctrl+Shift+D → drawing list opens
6. Click eye icon → trendline hides
7. Click again → trendline shows
```

---

## Common Issues

### Issue: Properties panel doesn't open
**Fix**: Check that drawing has `selected: true` when clicked
```typescript
// In TradingChart, when chart is clicked:
chart.subscribeClick((param) => {
  // ... existing click handling

  // Check if we clicked a drawing
  const drawings = drawingManager.getDrawings();
  const clicked = drawings.find(d => /* check if clicked */);
  if (clicked) {
    drawingManager.selectDrawing(clicked.id);
  }
});
```

### Issue: Drawing list not updating
**Fix**: Ensure all event listeners are registered
```typescript
// Check these events are dispatched in drawingManager:
'drawing:saved'
'drawing:deleted'
'drawing:visibility-changed'
'drawings:cleared'
```

### Issue: Visibility toggle not working
**Fix**: Use the new method in drawingManager
```typescript
// drawingManager already has this (added today):
drawingManager.toggleDrawingVisibility(id);

// NOT this:
drawing.visible = !drawing.visible; // Won't trigger re-render
```

---

## Files to Create

```
clients/desktop/src/components/
  DrawingPropertiesPanel.tsx    (~250 lines)
  DrawingListPanel.tsx           (~150 lines)
```

## Files to Modify

```
clients/desktop/src/components/
  TradingChart.tsx               (+50 lines)
```

---

## UI Components Needed

All already imported in project:

```typescript
import {
  X,           // Close button
  Eye,         // Show
  EyeOff,      // Hide
  Trash2,      // Delete
  Settings,    // Properties icon
  Palette,     // Color picker icon
} from 'lucide-react';
```

---

## Styling

Use existing Tailwind classes from DrawingTools:

```typescript
// Panel background
className="bg-zinc-800 border border-zinc-700 rounded-lg shadow-2xl"

// Button
className="px-2 py-1 text-xs bg-zinc-700 hover:bg-zinc-600 rounded"

// Selected item
className="bg-emerald-600/20 border border-emerald-600/50"
```

---

## Advanced Features (Later)

After basic integration works, add:

### Templates (Phase 2)

```typescript
// Add methods to drawingManager (already planned):
drawingManager.saveAsTemplate('My Setup');
drawingManager.loadTemplate(templateId);
drawingManager.getTemplates();
```

### Color Picker (Phase 3)

Replace this:
```typescript
const newColor = prompt('Enter color:');
```

With this:
```typescript
<DrawingColorPicker
  value={drawing.color}
  onChange={(color) => handleUpdate({ color })}
/>
```

### Keyboard Shortcuts (Phase 3)

```typescript
// T = Trendline
// H = Horizontal Line
// V = Vertical Line
// R = Rectangle
// etc.
```

---

## Need Help?

**Documentation**:
- Full analysis: `docs/DRAWING_INTEGRATION_TEST_REPORT.md`
- Action plan: `docs/INTEGRATION_ACTION_PLAN.md`
- Executive summary: `DRAWING_INTEGRATION_EXECUTIVE_SUMMARY.md`

**Code References**:
- Existing drawing system: `services/drawingManager.ts`
- UI to extract: `components/DrawingTools.tsx`
- Target file: `components/TradingChart.tsx`

**Events Reference**:
```typescript
// Listen to these:
'drawing:saved'              // Drawing created/updated
'drawing:deleted'            // Drawing removed
'drawing:visibility-changed' // Show/hide toggled
'drawing:duplicated'         // Drawing copied
'drawings:cleared'           // All removed

// Dispatch from drawingManager:
window.dispatchEvent(new CustomEvent('drawing:saved', {
  detail: drawing
}));
```

---

## Checklist

Before starting:
- [ ] Read this guide
- [ ] Review `DrawingTools.tsx` structure
- [ ] Review `drawingManager.ts` API

Phase 1 (Day 1):
- [ ] Create `DrawingPropertiesPanel.tsx`
- [ ] Create `DrawingListPanel.tsx`
- [ ] Test in isolation (Storybook or separate route)

Phase 2 (Day 2):
- [ ] Integrate into `TradingChart.tsx`
- [ ] Add state management
- [ ] Add event listeners
- [ ] Test click → properties open

Phase 3 (Day 3):
- [ ] Add keyboard shortcuts
- [ ] Polish UI
- [ ] Test all features
- [ ] Update documentation

---

**Good Luck!** 🚀

*Estimated time: 2-3 days*
*Difficulty: Medium*
*Risk: Low*
