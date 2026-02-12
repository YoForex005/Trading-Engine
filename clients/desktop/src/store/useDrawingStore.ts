/**
 * Drawing Store (Zustand)
 * Manages chart drawing tools and drawings state
 */

import { create } from 'zustand';

export type DrawingType =
  | 'trend-line'
  | 'horizontal-line'
  | 'vertical-line'
  | 'channel'
  | 'fib-retracement'
  | 'fib-extension'
  | 'rectangle'
  | 'ellipse'
  | 'pitchfork'
  | 'text-label'
  | 'arrow'
  | 'crosshair-ruler';

export type LineStyle = 'solid' | 'dashed' | 'dotted';

export interface Point {
  x: number;
  y: number;
  price?: number;
  time?: string;
}

export interface Drawing {
  id: string;
  type: DrawingType;
  points: Point[];
  color: string;
  lineWidth: number;
  lineStyle: LineStyle;
  extendLeft: boolean;
  extendRight: boolean;
  showPriceLabels: boolean;
  visible: boolean;
  zIndex: number;
  text?: string; // For text labels
  label?: string; // Custom label
}

export interface DrawingTemplate {
  id: string;
  name: string;
  drawings: Drawing[];
  createdAt: string;
}

interface DrawingState {
  drawings: Drawing[];
  selectedDrawingId: string | null;
  activeTool: DrawingType | null;
  templates: DrawingTemplate[];
  isDrawing: boolean;

  // Actions
  addDrawing: (drawing: Omit<Drawing, 'id' | 'zIndex'>) => void;
  updateDrawing: (id: string, updates: Partial<Drawing>) => void;
  deleteDrawing: (id: string) => void;
  duplicateDrawing: (id: string) => void;
  setSelectedDrawing: (id: string | null) => void;
  setActiveTool: (tool: DrawingType | null) => void;
  setIsDrawing: (isDrawing: boolean) => void;
  toggleDrawingVisibility: (id: string) => void;
  bringToFront: (id: string) => void;
  sendToBack: (id: string) => void;
  clearAllDrawings: () => void;

  // Template actions
  saveTemplate: (name: string) => void;
  loadTemplate: (id: string) => void;
  deleteTemplate: (id: string) => void;
}

// Helper to get next z-index
const getNextZIndex = (drawings: Drawing[]): number => {
  if (drawings.length === 0) return 1;
  return Math.max(...drawings.map(d => d.zIndex)) + 1;
};

// Load from localStorage
const loadFromStorage = () => {
  try {
    const stored = localStorage.getItem('rtx5-drawings');
    return stored ? JSON.parse(stored) : [];
  } catch (error) {
    console.error('Failed to load drawings from localStorage:', error);
    return [];
  }
};

const loadTemplatesFromStorage = () => {
  try {
    const stored = localStorage.getItem('rtx5-drawing-templates');
    return stored ? JSON.parse(stored) : [];
  } catch (error) {
    console.error('Failed to load templates from localStorage:', error);
    return [];
  }
};

export const useDrawingStore = create<DrawingState>((set, get) => ({
  drawings: loadFromStorage(),
  selectedDrawingId: null,
  activeTool: null,
  templates: loadTemplatesFromStorage(),
  isDrawing: false,

  addDrawing: (drawing) => {
    const newDrawing: Drawing = {
      ...drawing,
      id: `drawing-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      zIndex: getNextZIndex(get().drawings),
    };

    set((state) => {
      const newDrawings = [...state.drawings, newDrawing];
      localStorage.setItem('rtx5-drawings', JSON.stringify(newDrawings));
      return { drawings: newDrawings };
    });
  },

  updateDrawing: (id, updates) => {
    set((state) => {
      const newDrawings = state.drawings.map((d) =>
        d.id === id ? { ...d, ...updates } : d
      );
      localStorage.setItem('rtx5-drawings', JSON.stringify(newDrawings));
      return { drawings: newDrawings };
    });
  },

  deleteDrawing: (id) => {
    set((state) => {
      const newDrawings = state.drawings.filter((d) => d.id !== id);
      localStorage.setItem('rtx5-drawings', JSON.stringify(newDrawings));
      return {
        drawings: newDrawings,
        selectedDrawingId: state.selectedDrawingId === id ? null : state.selectedDrawingId,
      };
    });
  },

  duplicateDrawing: (id) => {
    const drawing = get().drawings.find((d) => d.id === id);
    if (!drawing) return;

    const duplicated: Drawing = {
      ...drawing,
      id: `drawing-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      zIndex: getNextZIndex(get().drawings),
      points: drawing.points.map(p => ({ ...p, x: p.x + 20, y: p.y + 20 })), // Offset slightly
    };

    set((state) => {
      const newDrawings = [...state.drawings, duplicated];
      localStorage.setItem('rtx5-drawings', JSON.stringify(newDrawings));
      return { drawings: newDrawings, selectedDrawingId: duplicated.id };
    });
  },

  setSelectedDrawing: (id) => set({ selectedDrawingId: id }),

  setActiveTool: (tool) => set({ activeTool: tool }),

  setIsDrawing: (isDrawing) => set({ isDrawing }),

  toggleDrawingVisibility: (id) => {
    set((state) => {
      const newDrawings = state.drawings.map((d) =>
        d.id === id ? { ...d, visible: !d.visible } : d
      );
      localStorage.setItem('rtx5-drawings', JSON.stringify(newDrawings));
      return { drawings: newDrawings };
    });
  },

  bringToFront: (id) => {
    const maxZ = Math.max(...get().drawings.map(d => d.zIndex));
    get().updateDrawing(id, { zIndex: maxZ + 1 });
  },

  sendToBack: (id) => {
    get().updateDrawing(id, { zIndex: 0 });
    // Re-index all drawings
    set((state) => {
      const newDrawings = state.drawings
        .sort((a, b) => a.zIndex - b.zIndex)
        .map((d, idx) => ({ ...d, zIndex: idx + 1 }));
      localStorage.setItem('rtx5-drawings', JSON.stringify(newDrawings));
      return { drawings: newDrawings };
    });
  },

  clearAllDrawings: () => {
    set({ drawings: [], selectedDrawingId: null });
    localStorage.removeItem('rtx5-drawings');
  },

  saveTemplate: (name) => {
    const { drawings } = get();
    const template: DrawingTemplate = {
      id: `template-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      name,
      drawings: drawings.map(d => ({ ...d })),
      createdAt: new Date().toISOString(),
    };

    set((state) => {
      const newTemplates = [...state.templates, template];
      localStorage.setItem('rtx5-drawing-templates', JSON.stringify(newTemplates));
      return { templates: newTemplates };
    });
  },

  loadTemplate: (id) => {
    const template = get().templates.find(t => t.id === id);
    if (!template) return;

    // Generate new IDs for loaded drawings
    const loadedDrawings: Drawing[] = template.drawings.map(d => ({
      ...d,
      id: `drawing-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      zIndex: getNextZIndex(get().drawings) + (d.zIndex || 0),
    }));

    set((state) => {
      const newDrawings = [...state.drawings, ...loadedDrawings];
      localStorage.setItem('rtx5-drawings', JSON.stringify(newDrawings));
      return { drawings: newDrawings };
    });
  },

  deleteTemplate: (id) => {
    set((state) => {
      const newTemplates = state.templates.filter(t => t.id !== id);
      localStorage.setItem('rtx5-drawing-templates', JSON.stringify(newTemplates));
      return { templates: newTemplates };
    });
  },
}));
