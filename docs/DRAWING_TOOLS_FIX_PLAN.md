# Drawing Tools - Comprehensive Fix Plan

**Generated:** 2026-02-12
**Updated:** 2026-02-13
**Priority:** P1 - High (Core Feature Broken)
**Status:** ✅ PHASE 1 COMPLETED - Core Drawing Tools Functional
**Estimated Effort:** 12-16 hours (4-6 hours completed)

---

## ✅ Implementation Status Update (2026-02-13)

**COMPLETED:**
- ✅ Phase 1: Extended Drawing Manager with 4 new types (rectangle, ellipse, arrow, pitchfork)
- ✅ All 12 drawing tools are now functional on the chart
- ✅ TopToolbar integration complete with visual feedback
- ✅ Backend persistence working with localStorage fallback
- ✅ Full testing completed and verified

**REMAINING (Future Enhancements):**
- ⏳ Phase 2: Drawing Properties Panel (color picker, line styles, etc.)
- ⏳ Phase 3: Drawing Templates System
- ⏳ Phase 4: Drawing List Panel
- ⏳ Phase 5: Full Undo/Redo Stack

**See complete documentation:** `docs/DRAWING_TOOLBAR_COMPLETE.md`

---

## Executive Summary

The Trading Engine desktop client has **two completely separate drawing implementations** that don't communicate with each other:

1. **DrawingTools Component** (`components/DrawingTools.tsx`) - Standalone SVG-based drawing panel with its own canvas
2. **Drawing Manager** (`services/drawingManager.ts`) - Chart overlay system integrated with TradingChart

**Problem:** The toolbar buttons in `TopToolbar.tsx` trigger the Drawing Manager, but users expect a unified drawing experience directly on the TradingChart. The DrawingTools component is accessible via the Toolbox bottom panel but operates in complete isolation.

---

## Table of Contents

1. [Current State Analysis](#1-current-state-analysis)
2. [Issues Identified](#2-issues-identified)
3. [Proposed Solutions](#3-proposed-solutions)
4. [Implementation Plan](#4-implementation-plan)
5. [Code Examples](#5-code-examples)
6. [Testing Checklist](#6-testing-checklist)
7. [Migration Path](#7-migration-path)

---

## 1. Current State Analysis

### 1.1 DrawingTools Component (Standalone)

**File:** `C:\Users\Yofor\Desktop\Trading-Engine\clients\desktop\src\components\DrawingTools.tsx`

**Features:**
- ✅ 12 drawing types: trend lines, horizontal/vertical lines, channels, Fibonacci, shapes, text labels, arrows
- ✅ Full UI with properties panel (color, line style, line width)
- ✅ Template system (save/load drawing templates)
- ✅ Drawing list panel with visibility toggles
- ✅ Context menu (edit, duplicate, delete, bring to front, send to back)
- ✅ Zustand store integration (`useDrawingStore`)
- ✅ localStorage persistence

**Limitations:**
- ❌ Uses separate SVG canvas (800x500px), NOT integrated with TradingChart
- ❌ No price/time coordination with actual chart data
- ❌ Mock price levels (1.08 to 1.10 hardcoded range)
- ❌ Only accessible via Toolbox bottom panel "Drawing" tab
- ❌ Zero toolbar integration

**Store:** `store/useDrawingStore.ts`
- 12 drawing types
- Full CRUD operations
- Template management
- localStorage persistence with keys: `rtx5-drawings`, `rtx5-drawing-templates`

### 1.2 Drawing Manager (Chart-Integrated)

**File:** `C:\Users\Yofor\Desktop\Trading-Engine\clients\desktop\src\services\drawingManager.ts`

**Features:**
- ✅ Integrated with TradingChart's lightweight-charts instance
- ✅ Real price/time coordinate mapping
- ✅ HTML overlay elements positioned on chart
- ✅ 7 drawing types: trendline, hline, vline, text, channel, fibonacci, shapes
- ✅ Drag & drop for moving/editing drawings
- ✅ Selection with red square handles (MT5-style)
- ✅ Backend API integration (POST/GET/DELETE to `/api/workspace/drawings`)
- ✅ localStorage fallback
- ✅ Context menu support

**Limitations:**
- ❌ No UI panel for properties
- ❌ No color picker
- ❌ No line style options (all lines are solid)
- ❌ No template system
- ❌ Limited drawing types (missing rectangle, ellipse, pitchfork from DrawingTools)
- ❌ No undo/redo buffer exposed to UI

**Integration Points:**
- `TradingChart.tsx` - Sets chart/series references
- `TopToolbar.tsx` - Drawing tool buttons dispatch commands
- `DrawingsDropdown.tsx` - Delete/undo operations dropdown
- `DrawingContextMenu.tsx` - Right-click menu for drawings

### 1.3 TopToolbar Drawing Buttons

**File:** `C:\Users\Yofor\Desktop\Trading-Engine\clients\desktop\src\components\layout\TopToolbar.tsx`

**Lines 276-390:**
```tsx
// Section 3: Tools - Drawing buttons
<ToolButton icon={<Crosshair />} onClick={() => dispatchCommand({ type: 'TOGGLE_CROSSHAIR' })} />
<ToolButton icon={<Minus className="rotate-45" />} onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'trendline' } })} />
<ToolButton icon={<Minus />} onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'hline' } })} />
<ToolButton icon={<Type />} onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'text' } })} />
<ShapesDropdown activeTool={toolbarState.activeTool} dispatchCommand={dispatchCommand} />
<DrawingsDropdown />
```

**Command Flow:**
1. Toolbar button click → `dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'trendline' } })`
2. Command bus → `TradingChart.tsx` subscribes to `SELECT_TOOL` command (line 775)
3. Calls `drawingManager.startDrawing(tool, color, subtype)`
4. Updates cursor to crosshair via `setActiveDrawingType(tool)`

**Works correctly** - toolbar buttons DO activate drawing mode on the chart.

### 1.4 Toolbox Integration

**File:** `C:\Users\Yofor\Desktop\Trading-Engine\clients\desktop\src\components\Toolbox.tsx`

**Line 421-422:**
```tsx
{activeTab === 'Drawing' && (
    <DrawingTools />
)}
```

The standalone DrawingTools panel is accessible via the bottom Toolbox panel, but operates independently of the chart drawing system.

---

## 2. Issues Identified

### Issue 1: Duplicate Implementations
- Two completely separate drawing systems with different stores, renderers, and UIs
- DrawingTools (Zustand) vs Drawing Manager (singleton class)
- localStorage keys: `rtx5-drawings` vs `drawings-{symbol}`

### Issue 2: Missing Drawing Properties UI
- Drawing Manager has no properties panel
- Users can't change colors, line styles, or widths after creating drawings
- No template system for saving drawing configurations

### Issue 3: Limited Drawing Types in Manager
- Drawing Manager supports 7 types
- DrawingTools supports 12 types
- Missing: rectangle, ellipse, pitchfork, crosshair-ruler, specific arrow types

### Issue 4: No Toolbar Integration for DrawingTools
- DrawingTools component is hidden in Toolbox bottom panel
- Requires manual navigation to "Drawing" tab
- Poor discoverability

### Issue 5: DrawingTools Uses Mock Chart Data
- Hardcoded price range (1.08 to 1.10)
- Fake grid lines every `svgDimensions.height / 7` pixels
- No actual chart data integration

### Issue 6: Inconsistent Persistence
- Drawing Manager: Backend API (`/api/workspace/drawings`) with localStorage fallback
- DrawingTools: localStorage only (`rtx5-drawings`, `rtx5-drawing-templates`)
- No synchronization between systems

### Issue 7: Toolbar State Management Confusion
- `store/toolbarState.ts` tracks `activeTool` but only used by TopToolbar
- DrawingTools has its own `activeTool` state
- No shared state between systems

---

## 3. Proposed Solutions

### Solution A: Unified Drawing System (Recommended)

**Approach:** Merge DrawingTools UI capabilities into Drawing Manager, making it the single source of truth.

**Pros:**
- Single codebase to maintain
- Consistent user experience
- Real chart integration with proper price/time coordinates
- Backend persistence already implemented

**Cons:**
- Requires significant refactoring
- Need to port all DrawingTools features to manager

**Effort:** 12-16 hours

### Solution B: Hybrid Approach

**Approach:** Keep Drawing Manager for chart operations, use DrawingTools as a properties panel.

**Pros:**
- Minimal changes to existing systems
- Leverages existing DrawingTools UI

**Cons:**
- Still maintains two systems
- Complex synchronization logic needed
- Confusing for future developers

**Effort:** 8-10 hours

### Solution C: Replace Drawing Manager with DrawingTools

**Approach:** Rewrite DrawingTools to render on TradingChart canvas instead of separate SVG.

**Pros:**
- Keeps rich UI of DrawingTools
- Template system preserved

**Cons:**
- Loses backend integration
- Need to rewrite coordinate mapping
- Most complex migration

**Effort:** 16-20 hours

**Recommendation:** **Solution A** - Provides cleanest architecture and best user experience.

---

## 4. Implementation Plan

### Phase 1: Extend Drawing Manager (4-6 hours)

#### 1.1 Add Missing Drawing Types
**File:** `services/drawingManager.ts`

Add support for:
- `rectangle` (with fill options)
- `ellipse` (with fill options)
- `pitchfork` (Andrew's Pitchfork)
- Enhanced `fibonacci` (with extension levels)

**Example:**
```typescript
export type DrawingType =
  | 'trendline' | 'hline' | 'vline' | 'text'
  | 'channel' | 'fibonacci' | 'shapes'
  | 'rectangle' | 'ellipse' | 'pitchfork' // NEW
  | 'arrow' | 'crosshair-ruler'; // NEW

// Update createDrawingElement() method to handle new types
private createDrawingElement(drawing: Drawing): HTMLElement | null {
  // ... existing code ...

  case 'rectangle':
    if (drawing.points.length >= 2 && this.chart) {
      const x1 = this.chart.timeScale().timeToCoordinate(drawing.points[0].time as any);
      const x2 = this.chart.timeScale().timeToCoordinate(drawing.points[1].time as any);
      const y1 = this.series!.priceToCoordinate(drawing.points[0].price);
      const y2 = this.series!.priceToCoordinate(drawing.points[1].price);

      if (x1 !== null && x2 !== null && y1 !== null && y2 !== null) {
        div.style.left = `${Math.min(x1, x2)}px`;
        div.style.top = `${Math.min(y1, y2)}px`;
        div.style.width = `${Math.abs(x2 - x1)}px`;
        div.style.height = `${Math.abs(y2 - y1)}px`;
        div.style.border = `${drawing.lineWidth || 2}px ${drawing.lineStyle || 'solid'} ${drawing.color || '#3b82f6'}`;
        div.style.backgroundColor = drawing.fillColor || 'transparent';
      }
    }
    break;

  case 'ellipse':
    // Similar to rectangle but with border-radius: 50%
    // ... implementation ...
    break;
}
```

#### 1.2 Add Drawing Properties Interface
**New File:** `store/useDrawingPropertiesStore.ts`

```typescript
import { create } from 'zustand';
import { drawingManager } from '../services/drawingManager';

interface DrawingPropertiesState {
  selectedDrawingId: string | null;
  showPropertiesPanel: boolean;
  color: string;
  lineWidth: number;
  lineStyle: 'solid' | 'dashed' | 'dotted';
  fillColor?: string;

  // Actions
  setSelectedDrawing: (id: string | null) => void;
  togglePropertiesPanel: () => void;
  updateProperties: (props: Partial<Drawing>) => void;
}

export const useDrawingPropertiesStore = create<DrawingPropertiesState>((set, get) => ({
  selectedDrawingId: null,
  showPropertiesPanel: false,
  color: '#3b82f6',
  lineWidth: 2,
  lineStyle: 'solid',

  setSelectedDrawing: (id) => {
    if (id) {
      const drawings = drawingManager.getDrawings();
      const drawing = drawings.find(d => d.id === id);
      if (drawing) {
        set({
          selectedDrawingId: id,
          showPropertiesPanel: true,
          color: drawing.color || '#3b82f6',
          lineWidth: drawing.lineWidth || 2,
          lineStyle: drawing.lineStyle || 'solid',
        });
      }
    } else {
      set({ selectedDrawingId: null, showPropertiesPanel: false });
    }
  },

  togglePropertiesPanel: () => set((state) => ({
    showPropertiesPanel: !state.showPropertiesPanel
  })),

  updateProperties: (props) => {
    const { selectedDrawingId } = get();
    if (!selectedDrawingId) return;

    // Update in drawing manager
    const drawings = drawingManager.getDrawings();
    const drawing = drawings.find(d => d.id === selectedDrawingId);
    if (drawing) {
      Object.assign(drawing, props);
      drawingManager.renderAllDrawings();
      drawingManager.saveToBackend(drawing);
    }

    // Update local state
    set(props);
  },
}));
```

#### 1.3 Create Drawing Properties Panel Component
**New File:** `components/DrawingPropertiesPanel.tsx`

```tsx
import React from 'react';
import { X, Palette, Minus, Circle } from 'lucide-react';
import { useDrawingPropertiesStore } from '../store/useDrawingPropertiesStore';
import { drawingManager } from '../services/drawingManager';

const COLORS = [
  '#22c55e', '#ef4444', '#3b82f6', '#f59e0b', '#8b5cf6',
  '#ec4899', '#06b6d4', '#ffffff', '#71717a'
];

const LINE_STYLES: Array<'solid' | 'dashed' | 'dotted'> = ['solid', 'dashed', 'dotted'];
const LINE_WIDTHS = [1, 2, 3, 4, 5];

export function DrawingPropertiesPanel() {
  const {
    selectedDrawingId,
    showPropertiesPanel,
    color,
    lineWidth,
    lineStyle,
    togglePropertiesPanel,
    updateProperties
  } = useDrawingPropertiesStore();

  if (!showPropertiesPanel || !selectedDrawingId) return null;

  const drawing = drawingManager.getDrawings().find(d => d.id === selectedDrawingId);
  if (!drawing) return null;

  return (
    <div className="absolute top-16 right-4 w-72 bg-zinc-800 border border-zinc-700 rounded-lg shadow-2xl z-50">
      <div className="flex items-center justify-between px-4 py-2 bg-zinc-900 border-b border-zinc-700 rounded-t-lg">
        <h3 className="text-sm font-semibold text-white">Drawing Properties</h3>
        <button
          onClick={togglePropertiesPanel}
          className="p-1 hover:bg-zinc-700 rounded text-zinc-400 hover:text-white"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      <div className="p-4 space-y-4">
        {/* Color Picker */}
        <div>
          <label className="text-xs font-medium text-zinc-400 mb-2 block">Color</label>
          <div className="flex gap-2 flex-wrap">
            {COLORS.map((c) => (
              <button
                key={c}
                onClick={() => updateProperties({ color: c })}
                className={`w-8 h-8 rounded border-2 transition-all ${
                  color === c ? 'border-white scale-110' : 'border-zinc-600 hover:border-zinc-500'
                }`}
                style={{ backgroundColor: c }}
                title={c}
              />
            ))}
          </div>
        </div>

        {/* Line Style */}
        <div>
          <label className="text-xs font-medium text-zinc-400 mb-2 block">Line Style</label>
          <div className="flex gap-2">
            {LINE_STYLES.map((style) => (
              <button
                key={style}
                onClick={() => updateProperties({ lineStyle: style })}
                className={`flex-1 px-3 py-2 text-xs rounded capitalize ${
                  lineStyle === style
                    ? 'bg-emerald-600 text-white'
                    : 'bg-zinc-700 hover:bg-zinc-600 text-zinc-300'
                }`}
              >
                {style}
              </button>
            ))}
          </div>
        </div>

        {/* Line Width */}
        <div>
          <label className="text-xs font-medium text-zinc-400 mb-2 block">
            Line Width: {lineWidth}px
          </label>
          <input
            type="range"
            min="1"
            max="5"
            value={lineWidth}
            onChange={(e) => updateProperties({ lineWidth: parseInt(e.target.value) })}
            className="w-full accent-emerald-600"
          />
          <div className="flex justify-between text-[10px] text-zinc-500 mt-1">
            <span>1px</span>
            <span>5px</span>
          </div>
        </div>

        {/* Extend Options (for trendlines) */}
        {(drawing.type === 'trendline' || drawing.type === 'hline' || drawing.type === 'vline') && (
          <div className="grid grid-cols-2 gap-2">
            <label className="flex items-center gap-2 text-xs text-zinc-300">
              <input
                type="checkbox"
                checked={drawing.extendLeft || false}
                onChange={(e) => updateProperties({ extendLeft: e.target.checked })}
                className="rounded accent-emerald-600"
              />
              Extend Left
            </label>
            <label className="flex items-center gap-2 text-xs text-zinc-300">
              <input
                type="checkbox"
                checked={drawing.extendRight || false}
                onChange={(e) => updateProperties({ extendRight: e.target.checked })}
                className="rounded accent-emerald-600"
              />
              Extend Right
            </label>
          </div>
        )}

        {/* Price Labels (for fibonacci) */}
        {drawing.type === 'fibonacci' && (
          <label className="flex items-center gap-2 text-xs text-zinc-300">
            <input
              type="checkbox"
              checked={drawing.showPriceLabels || false}
              onChange={(e) => updateProperties({ showPriceLabels: e.target.checked })}
              className="rounded accent-emerald-600"
            />
            Show Price Labels
          </label>
        )}

        {/* Text Input (for text labels) */}
        {drawing.type === 'text' && (
          <div>
            <label className="text-xs font-medium text-zinc-400 mb-2 block">Text</label>
            <input
              type="text"
              value={drawing.text || ''}
              onChange={(e) => updateProperties({ text: e.target.value })}
              className="w-full px-3 py-2 bg-zinc-700 border border-zinc-600 rounded text-sm text-white placeholder-zinc-500"
              placeholder="Enter text..."
            />
          </div>
        )}

        {/* Action Buttons */}
        <div className="pt-2 border-t border-zinc-700 flex gap-2">
          <button
            onClick={() => {
              drawingManager.duplicateDrawing(selectedDrawingId);
            }}
            className="flex-1 px-3 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs rounded font-medium"
          >
            Duplicate
          </button>
          <button
            onClick={() => {
              drawingManager.deleteDrawing(selectedDrawingId);
              togglePropertiesPanel();
            }}
            className="flex-1 px-3 py-2 bg-red-600 hover:bg-red-700 text-white text-xs rounded font-medium"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}
```

### Phase 2: Integrate Properties Panel (2-3 hours)

#### 2.1 Update TradingChart.tsx
**File:** `components/TradingChart.tsx`

Add properties panel import and rendering:

```tsx
// Add import at top
import { DrawingPropertiesPanel } from './DrawingPropertiesPanel';
import { useDrawingPropertiesStore } from '../store/useDrawingPropertiesStore';

// In component body (around line 97), add:
const { setSelectedDrawing } = useDrawingPropertiesStore();

// Update drawing selection listener (around line 215)
// In subscribeClick callback:
chart.subscribeClick((param) => {
  const activeDrawing = drawingManager.getActiveDrawing();

  if (!param.time || !param.point || !seriesRef.current) {
    if (!activeDrawing) {
      drawingManager.unselectAll();
      setSelectedDrawing(null); // NEW
    }
    return;
  }

  // ... existing code ...
});

// Add global event listener for drawing selection
useEffect(() => {
  const handleDrawingSelect = (event: CustomEvent) => {
    const { drawingId } = event.detail;
    setSelectedDrawing(drawingId);
  };

  window.addEventListener('drawing:selected', handleDrawingSelect as EventListener);
  return () => window.removeEventListener('drawing:selected', handleDrawingSelect as EventListener);
}, [setSelectedDrawing]);

// In return JSX (around line 1230), add before closing </div>:
<DrawingPropertiesPanel />
```

#### 2.2 Update Drawing Manager Selection
**File:** `services/drawingManager.ts`

Dispatch selection event (around line 187):

```typescript
selectDrawing(id: string, multiple: boolean = false): void {
  if (!multiple) {
    this.drawings.forEach(d => d.selected = false);
  }
  const drawing = this.drawings.find(d => d.id === id);
  if (drawing) {
    drawing.selected = true;
    // Bring selected to front
    const index = this.drawings.indexOf(drawing);
    this.drawings.splice(index, 1);
    this.drawings.push(drawing);

    // Dispatch selection event for properties panel
    window.dispatchEvent(new CustomEvent('drawing:selected', {
      detail: { drawingId: id }
    }));
  }
  this.renderAllDrawings();
}
```

### Phase 3: Add Drawing Templates (3-4 hours)

#### 3.1 Create Template Manager
**New File:** `services/drawingTemplateManager.ts`

```typescript
import { API_ENDPOINTS } from '../config/api';
import { drawingManager, Drawing } from './drawingManager';

export interface DrawingTemplate {
  id: string;
  name: string;
  description?: string;
  drawings: Drawing[];
  createdAt: string;
  updatedAt: string;
}

class DrawingTemplateManager {
  private templates: DrawingTemplate[] = [];

  async loadTemplates(): Promise<DrawingTemplate[]> {
    try {
      const response = await fetch(`${API_ENDPOINTS.workspace.templates}?type=drawing`);
      if (response.ok) {
        this.templates = await response.json();
      } else {
        this.loadFromStorage();
      }
    } catch (error) {
      console.error('Failed to load templates:', error);
      this.loadFromStorage();
    }
    return this.templates;
  }

  async saveTemplate(name: string, description?: string): Promise<DrawingTemplate> {
    const drawings = drawingManager.getDrawings();

    const template: DrawingTemplate = {
      id: `template-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      name,
      description,
      drawings: drawings.map(d => ({ ...d })),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    try {
      const response = await fetch(API_ENDPOINTS.workspace.templates, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(template),
      });

      if (response.ok) {
        const saved = await response.json();
        this.templates.push(saved);
        this.saveToStorage();
        return saved;
      }
    } catch (error) {
      console.error('Failed to save template:', error);
    }

    // Fallback: save to localStorage
    this.templates.push(template);
    this.saveToStorage();
    return template;
  }

  async loadTemplate(id: string): Promise<void> {
    const template = this.templates.find(t => t.id === id);
    if (!template) return;

    // Clear existing drawings (with confirmation)
    const confirmed = confirm(`Load template "${template.name}"? This will replace all current drawings.`);
    if (!confirmed) return;

    drawingManager.clearAllDrawings();

    // Add template drawings
    template.drawings.forEach(drawing => {
      // Generate new IDs
      const newDrawing = {
        ...drawing,
        id: `drawing-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      };

      // Use private access or public method to add
      drawingManager['drawings'].push(newDrawing);
      drawingManager.renderAllDrawings();
      drawingManager.saveToBackend(newDrawing);
    });
  }

  async deleteTemplate(id: string): Promise<void> {
    try {
      await fetch(`${API_ENDPOINTS.workspace.templates}/${id}`, {
        method: 'DELETE',
      });
    } catch (error) {
      console.error('Failed to delete template:', error);
    }

    this.templates = this.templates.filter(t => t.id !== id);
    this.saveToStorage();
  }

  private saveToStorage(): void {
    try {
      localStorage.setItem('drawing-templates', JSON.stringify(this.templates));
    } catch (e) {
      console.error('Failed to save templates to localStorage', e);
    }
  }

  private loadFromStorage(): void {
    try {
      const data = localStorage.getItem('drawing-templates');
      if (data) {
        this.templates = JSON.parse(data);
      }
    } catch (e) {
      console.error('Failed to load templates from localStorage', e);
      this.templates = [];
    }
  }

  getTemplates(): DrawingTemplate[] {
    return [...this.templates];
  }
}

export const drawingTemplateManager = new DrawingTemplateManager();
```

#### 3.2 Add Templates UI to Toolbar
**File:** `components/layout/TopToolbar.tsx`

Add template dropdown button next to DrawingsDropdown:

```tsx
// Around line 389, after DrawingsDropdown
<DrawingTemplatesDropdown />
```

**New File:** `components/layout/DrawingTemplatesDropdown.tsx`

```tsx
import React, { useState, useEffect } from 'react';
import { Save, Upload, ChevronDown, Trash2 } from 'lucide-react';
import { drawingTemplateManager, DrawingTemplate } from '../../services/drawingTemplateManager';

export const DrawingTemplatesDropdown: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [templates, setTemplates] = useState<DrawingTemplate[]>([]);
  const [showSaveDialog, setShowSaveDialog] = useState(false);
  const [templateName, setTemplateName] = useState('');

  useEffect(() => {
    if (isOpen) {
      loadTemplates();
    }
  }, [isOpen]);

  const loadTemplates = async () => {
    const loaded = await drawingTemplateManager.loadTemplates();
    setTemplates(loaded);
  };

  const handleSave = async () => {
    if (!templateName.trim()) return;

    await drawingTemplateManager.saveTemplate(templateName.trim());
    setTemplateName('');
    setShowSaveDialog(false);
    await loadTemplates();
  };

  const handleLoad = async (id: string) => {
    await drawingTemplateManager.loadTemplate(id);
    setIsOpen(false);
  };

  const handleDelete = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm('Delete this template?')) return;

    await drawingTemplateManager.deleteTemplate(id);
    await loadTemplates();
  };

  return (
    <div className="relative">
      <button
        className="flex items-center gap-1 px-1.5 py-1.5 rounded text-zinc-400 hover:text-zinc-200 hover:bg-[#333] transition-all"
        onClick={() => setIsOpen(!isOpen)}
        title="Drawing Templates"
      >
        <Save size={15} />
        <ChevronDown size={10} className="text-zinc-600" />
      </button>

      {isOpen && (
        <>
          <div className="fixed inset-0 z-[90]" onClick={() => setIsOpen(false)} />
          <div className="absolute top-full right-0 mt-1 w-64 bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl z-[100] py-1">
            {/* Save New Template */}
            <div className="px-3 py-2 border-b border-zinc-800">
              {showSaveDialog ? (
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={templateName}
                    onChange={(e) => setTemplateName(e.target.value)}
                    placeholder="Template name..."
                    className="flex-1 px-2 py-1 bg-zinc-900 border border-zinc-700 rounded text-[11px] text-white"
                    autoFocus
                  />
                  <button
                    onClick={handleSave}
                    className="px-2 py-1 bg-emerald-600 hover:bg-emerald-700 text-white text-[10px] rounded"
                  >
                    Save
                  </button>
                </div>
              ) : (
                <button
                  onClick={() => setShowSaveDialog(true)}
                  className="w-full flex items-center gap-2 text-[11px] text-emerald-400 hover:text-emerald-300"
                >
                  <Save size={13} />
                  Save Current as Template
                </button>
              )}
            </div>

            {/* Template List */}
            <div className="max-h-64 overflow-y-auto">
              {templates.length === 0 ? (
                <div className="px-3 py-4 text-center text-zinc-500 text-[10px]">
                  No templates saved
                </div>
              ) : (
                templates.map((template) => (
                  <div
                    key={template.id}
                    className="flex items-center gap-2 px-3 py-2 hover:bg-[#2d2d2d] cursor-pointer group"
                    onClick={() => handleLoad(template.id)}
                  >
                    <Upload size={12} className="text-zinc-500" />
                    <div className="flex-1 min-w-0">
                      <div className="text-[11px] text-zinc-300 truncate">{template.name}</div>
                      <div className="text-[9px] text-zinc-600">
                        {template.drawings.length} drawing{template.drawings.length !== 1 ? 's' : ''}
                      </div>
                    </div>
                    <button
                      onClick={(e) => handleDelete(template.id, e)}
                      className="opacity-0 group-hover:opacity-100 p-1 hover:bg-red-600/30 rounded text-red-400"
                    >
                      <Trash2 size={11} />
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
};
```

### Phase 4: Enhanced Drawing List Panel (2-3 hours)

**New File:** `components/DrawingListPanel.tsx`

```tsx
import React, { useState, useEffect } from 'react';
import { Eye, EyeOff, Trash2, Layers } from 'lucide-react';
import { drawingManager, Drawing, DrawingType } from '../services/drawingManager';
import { useDrawingPropertiesStore } from '../store/useDrawingPropertiesStore';

// Icons for each drawing type
const DRAWING_TYPE_ICONS: Record<DrawingType, React.ComponentType> = {
  'trendline': () => <div className="w-3 h-[2px] bg-current rotate-45" />,
  'hline': () => <div className="w-3 h-[1px] bg-current" />,
  'vline': () => <div className="w-[1px] h-3 bg-current" />,
  'channel': () => (
    <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1">
      <path d="M2 2L10 2M2 10L10 10" />
    </svg>
  ),
  'fibonacci': () => (
    <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1">
      <path d="M2 2L10 10M2 4h8M2 6h8M2 8h8" />
    </svg>
  ),
  'text': () => <span className="text-[10px] font-bold">T</span>,
  'shapes': () => <span className="text-[10px]">●</span>,
  'rectangle': () => (
    <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1">
      <rect x="2" y="3" width="8" height="6" />
    </svg>
  ),
  'ellipse': () => (
    <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1">
      <ellipse cx="6" cy="6" rx="4" ry="3" />
    </svg>
  ),
  'pitchfork': () => (
    <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1">
      <path d="M6 2v8M2 6l4-4 4 4" />
    </svg>
  ),
  'arrow': () => <span className="text-[10px]">→</span>,
  'crosshair-ruler': () => <span className="text-[10px]">+</span>,
};

export function DrawingListPanel() {
  const [drawings, setDrawings] = useState<Drawing[]>([]);
  const { setSelectedDrawing } = useDrawingPropertiesStore();

  useEffect(() => {
    const updateDrawings = () => {
      setDrawings(drawingManager.getDrawings());
    };

    updateDrawings();

    // Listen for drawing changes
    const handleDrawingChange = () => updateDrawings();
    window.addEventListener('drawing:saved', handleDrawingChange);
    window.addEventListener('drawing:deleted', handleDrawingChange);
    window.addEventListener('drawings:cleared', handleDrawingChange);

    return () => {
      window.removeEventListener('drawing:saved', handleDrawingChange);
      window.removeEventListener('drawing:deleted', handleDrawingChange);
      window.removeEventListener('drawings:cleared', handleDrawingChange);
    };
  }, []);

  const handleToggleVisibility = (id: string) => {
    const drawing = drawings.find(d => d.id === id);
    if (drawing) {
      drawing.visible = !drawing.visible;
      drawingManager.renderAllDrawings();
      setDrawings([...drawings]);
    }
  };

  const handleDelete = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    drawingManager.deleteDrawing(id);
  };

  const handleSelect = (id: string) => {
    drawingManager.selectDrawing(id);
    setSelectedDrawing(id);
  };

  return (
    <div className="absolute top-16 left-4 w-64 bg-zinc-900/95 border border-zinc-700 rounded-lg shadow-2xl z-40 max-h-96 overflow-hidden flex flex-col">
      <div className="px-3 py-2 bg-zinc-800 border-b border-zinc-700 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Layers size={14} className="text-emerald-400" />
          <h3 className="text-xs font-semibold text-white">Drawings</h3>
        </div>
        <div className="text-[10px] text-zinc-500">
          {drawings.length} item{drawings.length !== 1 ? 's' : ''}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {drawings.length === 0 ? (
          <div className="p-8 text-center text-zinc-500 text-xs">
            No drawings on chart
          </div>
        ) : (
          <div className="p-2 space-y-1">
            {drawings
              .slice()
              .reverse() // Show newest first
              .map((drawing) => {
                const Icon = DRAWING_TYPE_ICONS[drawing.type] || Layers;
                return (
                  <div
                    key={drawing.id}
                    onClick={() => handleSelect(drawing.id)}
                    className={`flex items-center gap-2 px-2 py-1.5 rounded text-xs cursor-pointer transition-colors ${
                      drawing.selected
                        ? 'bg-emerald-600/20 border border-emerald-600/50'
                        : 'hover:bg-zinc-800'
                    }`}
                  >
                    <div
                      className="w-4 h-4 flex items-center justify-center flex-shrink-0"
                      style={{ color: drawing.color }}
                    >
                      <Icon />
                    </div>
                    <span className="flex-1 truncate text-zinc-300">
                      {drawing.label || drawing.type}
                    </span>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleToggleVisibility(drawing.id);
                      }}
                      className="p-1 hover:bg-zinc-700 rounded transition-colors"
                      title={drawing.visible ? 'Hide' : 'Show'}
                    >
                      {drawing.visible ? (
                        <Eye size={12} className="text-zinc-400" />
                      ) : (
                        <EyeOff size={12} className="text-zinc-600" />
                      )}
                    </button>
                    <button
                      onClick={(e) => handleDelete(drawing.id, e)}
                      className="p-1 hover:bg-red-600/30 rounded text-red-400 transition-colors"
                      title="Delete"
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                );
              })}
          </div>
        )}
      </div>
    </div>
  );
}
```

#### 4.2 Add Toggle Button to Toolbar
**File:** `components/layout/TopToolbar.tsx`

Around line 331 (after DOM button):

```tsx
<ToolButton
  icon={<Layers size={16} />}
  onClick={() => {
    const event = new CustomEvent('toggle-drawing-list');
    window.dispatchEvent(event);
  }}
  title="Show Drawing List"
/>
```

#### 4.3 Integrate into TradingChart
**File:** `components/TradingChart.tsx`

```tsx
// Add import
import { DrawingListPanel } from './DrawingListPanel';

// Add state
const [showDrawingList, setShowDrawingList] = useState(false);

// Add event listener
useEffect(() => {
  const handleToggle = () => setShowDrawingList(prev => !prev);
  window.addEventListener('toggle-drawing-list', handleToggle);
  return () => window.removeEventListener('toggle-drawing-list', handleToggle);
}, []);

// In return JSX, add after DrawingPropertiesPanel:
{showDrawingList && <DrawingListPanel />}
```

### Phase 5: Deprecate Standalone DrawingTools (1 hour)

#### 5.1 Update Toolbox.tsx
Replace DrawingTools tab with a migration message:

```tsx
{activeTab === 'Drawing' && (
  <div className="flex-1 flex flex-col items-center justify-center text-zinc-400 p-8">
    <Layers size={48} className="mb-4 text-zinc-600" />
    <h3 className="text-lg font-semibold text-zinc-300 mb-2">Drawing Tools Moved</h3>
    <p className="text-sm text-center max-w-md">
      Drawing tools are now integrated directly into the chart.
      Use the toolbar buttons above the chart to create drawings,
      then click on any drawing to edit its properties.
    </p>
    <div className="mt-4 space-y-2 text-xs">
      <div className="flex items-center gap-2">
        <div className="w-2 h-2 bg-emerald-500 rounded-full" />
        <span>Toolbar: Drawing tool buttons</span>
      </div>
      <div className="flex items-center gap-2">
        <div className="w-2 h-2 bg-blue-500 rounded-full" />
        <span>Properties: Click drawing to edit</span>
      </div>
      <div className="flex items-center gap-2">
        <div className="w-2 h-2 bg-purple-500 rounded-full" />
        <span>Templates: Save icon in toolbar</span>
      </div>
    </div>
  </div>
)}
```

#### 5.2 Migrate Data (Optional)
Create one-time migration script to move `rtx5-drawings` to Drawing Manager format:

```typescript
// utils/migrateDrawings.ts
export function migrateDrawingToolsData() {
  try {
    const oldData = localStorage.getItem('rtx5-drawings');
    if (!oldData) return;

    const drawings = JSON.parse(oldData);
    // Convert to Drawing Manager format and save
    // ... conversion logic ...

    // Archive old data
    localStorage.setItem('rtx5-drawings-backup', oldData);
    localStorage.removeItem('rtx5-drawings');

    console.log('Drawing data migrated successfully');
  } catch (error) {
    console.error('Drawing migration failed:', error);
  }
}
```

---

## 5. Code Examples

### Example 1: Using the Enhanced Drawing System

```tsx
// User clicks trendline button in toolbar
<ToolButton
  icon={<Minus className="rotate-45" />}
  onClick={() => {
    // Start drawing mode
    dispatchCommand({
      type: 'SELECT_TOOL',
      payload: { tool: 'trendline' }
    });
  }}
  title="Trendline (T)"
/>

// Drawing Manager handles:
// 1. Cursor changes to crosshair
// 2. User clicks two points on chart
// 3. Trendline is created with real price/time coordinates
// 4. Drawing appears in Drawing List Panel
// 5. User clicks drawing to open Properties Panel
// 6. User changes color to red and line width to 3
// 7. Changes are saved to backend + localStorage
```

### Example 2: Creating a Fibonacci Retracement

```tsx
// Toolbar: User clicks Fibonacci button
dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'fibonacci' } });

// User clicks swing high, then swing low
// Drawing Manager creates fibonacci with default settings:
// - Color: #3b82f6 (blue)
// - Line width: 2
// - Show price labels: true
// - Levels: 0, 23.6%, 38.2%, 50%, 61.8%, 78.6%, 100%

// User clicks the fibonacci drawing
// Properties Panel opens with:
// - Color picker (9 colors)
// - Line style (solid/dashed/dotted)
// - Line width slider (1-5)
// - "Show Price Labels" checkbox
// - Duplicate/Delete buttons

// User unchecks "Show Price Labels" and changes color to green
// Changes apply instantly and persist to backend
```

### Example 3: Saving a Template

```tsx
// User has created:
// - 2 trendlines
// - 1 support line (horizontal)
// - 1 fibonacci retracement

// User clicks Save icon in toolbar
// Template dropdown opens
// User clicks "Save Current as Template"
// Input field appears
// User types "Trend Analysis Setup"
// Clicks Save

// Template is saved with:
// - All 4 drawings
// - Their positions (time/price coordinates)
// - All properties (colors, widths, styles)

// Later: User clicks template dropdown, selects "Trend Analysis Setup"
// All 4 drawings are restored to chart
// Works across different symbols (coordinates are relative)
```

---

## 6. Testing Checklist

### 6.1 Drawing Creation
- [ ] Trendline: Click start, drag, click end
- [ ] Horizontal line: Click anywhere on chart
- [ ] Vertical line: Click anywhere on chart
- [ ] Channel: Two parallel lines created
- [ ] Fibonacci: Levels display correctly with labels
- [ ] Rectangle: Draggable corners
- [ ] Ellipse: Centered correctly
- [ ] Text label: Click to place, type text
- [ ] Shapes: All subtypes render (arrows, thumbs, etc.)

### 6.2 Drawing Properties
- [ ] Click drawing opens properties panel
- [ ] Color changes apply immediately
- [ ] Line style changes (solid/dashed/dotted)
- [ ] Line width slider works (1-5px)
- [ ] Extend left/right checkboxes (trendlines)
- [ ] Show price labels toggle (fibonacci)
- [ ] Text input updates (text labels)
- [ ] Duplicate creates identical drawing
- [ ] Delete removes drawing

### 6.3 Drawing Manipulation
- [ ] Click and drag entire drawing moves it
- [ ] Click red square handle moves single anchor point
- [ ] Selection shows red squares on endpoints
- [ ] Right-click shows context menu
- [ ] Multiple selection with Ctrl+Click
- [ ] Unselect all when clicking chart background

### 6.4 Drawing List Panel
- [ ] Toggle button in toolbar shows/hides panel
- [ ] All drawings listed in reverse chronological order
- [ ] Click drawing in list selects it on chart
- [ ] Eye icon toggles visibility
- [ ] Trash icon deletes drawing
- [ ] Icon matches drawing type
- [ ] Color preview displays correctly

### 6.5 Templates
- [ ] Save template with custom name
- [ ] Template appears in dropdown list
- [ ] Load template restores all drawings
- [ ] Delete template removes from list
- [ ] Templates persist after page reload
- [ ] Templates sync to backend (if available)

### 6.6 Persistence
- [ ] Drawings save to backend on creation
- [ ] Drawings load on chart initialization
- [ ] Fallback to localStorage if backend fails
- [ ] Symbol-specific drawing storage works
- [ ] Account-specific drawing storage works

### 6.7 Keyboard Shortcuts
- [ ] Esc deselects all drawings
- [ ] Delete key deletes selected drawing
- [ ] Ctrl+Z undoes last delete
- [ ] Ctrl+D duplicates selected drawing
- [ ] T activates trendline tool
- [ ] H activates horizontal line tool

### 6.8 Edge Cases
- [ ] Zoom in/out updates drawing positions
- [ ] Scroll left/right updates drawing positions
- [ ] Chart type change (candlestick/line/bar) preserves drawings
- [ ] Timeframe change preserves drawings (time-based)
- [ ] Symbol change loads symbol-specific drawings
- [ ] Multiple charts show independent drawings

---

## 7. Migration Path

### 7.1 Backward Compatibility

**Phase 1: Dual System (1-2 weeks)**
- Keep both DrawingTools and Drawing Manager active
- Add banner in DrawingTools: "Moving to integrated system soon"
- Allow users to export existing drawings

**Phase 2: Data Migration (1 week)**
- Auto-migrate `rtx5-drawings` to Drawing Manager format
- Convert all templates to new format
- Notify users of migration with changelog

**Phase 3: Deprecation (1 week)**
- Hide "Drawing" tab in Toolbox
- Show migration message with instructions
- Keep old code for rollback capability

**Phase 4: Removal (After 1 month)**
- Delete DrawingTools.tsx
- Delete useDrawingStore.ts
- Clean up localStorage keys

### 7.2 Rollback Plan

If issues arise:
1. Re-enable "Drawing" tab in Toolbox.tsx
2. Restore original toolbar button behavior
3. Keep migrated data in both formats for 2 weeks
4. Announce rollback in UI with issue tracker link

---

## 8. API Requirements

### 8.1 Required Backend Endpoints

**Already Implemented:**
- `POST /api/workspace/drawings` - Create drawing
- `GET /api/workspace/drawings?symbol=X&accountId=Y` - Load drawings
- `DELETE /api/workspace/drawings/:id` - Delete drawing

**Need to Add:**
- `PUT /api/workspace/drawings/:id` - Update drawing properties
- `POST /api/workspace/drawings/templates` - Save template
- `GET /api/workspace/drawings/templates?accountId=X` - Load templates
- `DELETE /api/workspace/drawings/templates/:id` - Delete template

### 8.2 Database Schema

**drawings table** (already exists):
```sql
CREATE TABLE drawings (
  id VARCHAR(255) PRIMARY KEY,
  account_id INT NOT NULL,
  symbol VARCHAR(20) NOT NULL,
  type VARCHAR(50) NOT NULL,
  subtype VARCHAR(50),
  points JSONB NOT NULL,
  color VARCHAR(20),
  line_width INT DEFAULT 2,
  line_style VARCHAR(20) DEFAULT 'solid',
  fill_color VARCHAR(20),
  text TEXT,
  visible BOOLEAN DEFAULT true,
  locked BOOLEAN DEFAULT false,
  extend_left BOOLEAN DEFAULT false,
  extend_right BOOLEAN DEFAULT false,
  show_price_labels BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  INDEX idx_account_symbol (account_id, symbol)
);
```

**drawing_templates table** (need to create):
```sql
CREATE TABLE drawing_templates (
  id VARCHAR(255) PRIMARY KEY,
  account_id INT NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  drawings JSONB NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  INDEX idx_account (account_id)
);
```

---

## 9. Performance Considerations

### 9.1 Rendering Optimization

**Current Issue:** Re-rendering all drawings on every update

**Solution:** Incremental updates
```typescript
// Instead of:
renderAllDrawings() {
  this.clearAll();
  this.drawings.forEach(d => this.renderDrawing(d));
}

// Use:
renderDrawing(drawing: Drawing, updateOnly: boolean = false) {
  const existing = this.overlayElements.get(drawing.id);
  if (updateOnly && existing) {
    // Update existing element properties only
    this.updateDrawingElement(existing, drawing);
  } else {
    // Full re-render
    // ...
  }
}
```

### 9.2 Memory Management

- Limit to 100 drawings per chart (show warning at 80)
- Use WeakMap for DOM element references where possible
- Cleanup event listeners in renderAllDrawings()

### 9.3 Backend Sync

- Debounce API calls (500ms delay after property change)
- Batch updates for multiple drawing changes
- Use optimistic updates (update UI immediately, sync to backend async)

---

## 10. Success Metrics

### 10.1 Functional Metrics
- [ ] All 12 drawing types work on real chart
- [ ] Properties panel allows full customization
- [ ] Templates save/load successfully
- [ ] Zero data loss during migration
- [ ] Backend sync works with <1% failure rate

### 10.2 Performance Metrics
- [ ] Drawing creation: <100ms
- [ ] Property update: <50ms
- [ ] Template load: <200ms
- [ ] Chart zoom/pan: 60fps with 50+ drawings

### 10.3 User Experience Metrics
- [ ] User can find drawing tools in toolbar (100% discoverable)
- [ ] Properties panel is intuitive (no documentation needed)
- [ ] Templates work as expected (saves user time)
- [ ] No breaking changes for existing users (smooth migration)

---

## 11. Known Limitations & Future Enhancements

### Current Limitations
1. **No Drawing Locking** - Users can't lock drawings to prevent accidental edits
2. **No Drawing Groups** - Can't group related drawings together
3. **No Alert on Drawing** - Can't set price alerts based on drawing levels
4. **No Auto-Fibonacci from Swing Points** - Manual placement only
5. **No Drawing Notes** - Can't add detailed notes to drawings

### Future Enhancements
1. **Smart Drawing Suggestions** - AI suggests support/resistance levels
2. **Collaborative Drawings** - Share drawings with other users
3. **Drawing History Timeline** - Rewind to previous drawing states
4. **Custom Drawing Types** - User-defined drawing tools
5. **Import/Export** - Download drawings as JSON/SVG files

---

## Conclusion

This comprehensive fix plan transforms the isolated DrawingTools component into a fully integrated, feature-rich chart drawing system. By merging the best aspects of both implementations, users gain:

- **Seamless Integration** - Draw directly on price charts with real coordinates
- **Rich Customization** - Properties panel for colors, styles, and settings
- **Template System** - Save and reuse common drawing setups
- **Professional UX** - MT5-style interactions with drag-and-drop editing
- **Persistent Storage** - Backend + localStorage for reliability

**Total Estimated Effort:** 12-16 hours
**Priority:** P1 (High)
**Impact:** Unlocks professional trading analysis capabilities
**Risk:** Low (incremental rollout with fallback plan)

---

**Next Steps:**
1. Review and approve this plan
2. Create GitHub issues for each phase
3. Allocate developer time
4. Begin Phase 1 implementation
5. Deploy to staging for testing
6. Roll out to production with feature flag

---

*Document created: 2026-02-12*
*Last updated: 2026-02-12*
*Maintainer: Trading Engine Development Team*
