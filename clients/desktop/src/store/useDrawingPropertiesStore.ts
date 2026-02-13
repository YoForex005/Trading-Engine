/**
 * Drawing Properties Store (Zustand)
 * Manages drawing properties panel state and selected drawing properties
 */

import { create } from 'zustand';
import type { Drawing } from '../services/drawingManager';

export interface DrawingProperties {
  color: string;
  lineWidth: number;
  lineStyle: 'solid' | 'dashed' | 'dotted';
  extendLeft?: boolean;
  extendRight?: boolean;
  showPriceLabels?: boolean;
  text?: string;
  visible?: boolean;
}

interface DrawingPropertiesState {
  // UI State
  isPanelOpen: boolean;
  selectedDrawingId: string | null;

  // Current Properties
  properties: DrawingProperties;

  // Actions
  togglePanel: () => void;
  openPanel: () => void;
  closePanel: () => void;
  setSelectedDrawing: (drawing: Drawing | null) => void;
  updateProperty: <K extends keyof DrawingProperties>(
    key: K,
    value: DrawingProperties[K]
  ) => void;
  updateProperties: (properties: Partial<DrawingProperties>) => void;
  resetProperties: () => void;
}

const DEFAULT_PROPERTIES: DrawingProperties = {
  color: '#3b82f6',
  lineWidth: 2,
  lineStyle: 'solid',
  extendLeft: false,
  extendRight: false,
  showPriceLabels: true,
  text: '',
  visible: true,
};

export const useDrawingPropertiesStore = create<DrawingPropertiesState>((set, get) => ({
  isPanelOpen: false,
  selectedDrawingId: null,
  properties: { ...DEFAULT_PROPERTIES },

  togglePanel: () => set((state) => ({ isPanelOpen: !state.isPanelOpen })),

  openPanel: () => set({ isPanelOpen: true }),

  closePanel: () => set({ isPanelOpen: false }),

  setSelectedDrawing: (drawing) => {
    if (!drawing) {
      set({
        selectedDrawingId: null,
        isPanelOpen: false,
        properties: { ...DEFAULT_PROPERTIES },
      });
      return;
    }

    set({
      selectedDrawingId: drawing.id,
      isPanelOpen: true,
      properties: {
        color: drawing.color || DEFAULT_PROPERTIES.color,
        lineWidth: drawing.lineWidth || DEFAULT_PROPERTIES.lineWidth,
        lineStyle: drawing.lineStyle || DEFAULT_PROPERTIES.lineStyle,
        extendLeft: false, // TODO: Add to Drawing interface
        extendRight: false, // TODO: Add to Drawing interface
        showPriceLabels: true, // TODO: Add to Drawing interface
        text: drawing.text || '',
        visible: drawing.visible !== false, // Default to true if undefined
      },
    });
  },

  updateProperty: (key, value) => {
    set((state) => ({
      properties: {
        ...state.properties,
        [key]: value,
      },
    }));
  },

  updateProperties: (newProperties) => {
    set((state) => ({
      properties: {
        ...state.properties,
        ...newProperties,
      },
    }));
  },

  resetProperties: () => {
    set({ properties: { ...DEFAULT_PROPERTIES } });
  },
}));
