/**
 * Drawing Properties Panel
 * Floating panel for customizing drawing appearance (color, line style, width, etc.)
 */

import React, { useEffect } from 'react';
import { X, Copy, Trash2, Eye, EyeOff } from 'lucide-react';
import { useDrawingPropertiesStore } from '../store/useDrawingPropertiesStore';
import { drawingManager } from '../services/drawingManager';

const COLOR_PALETTE = [
  '#ef4444', // Red
  '#f97316', // Orange
  '#eab308', // Yellow
  '#22c55e', // Green
  '#3b82f6', // Blue
  '#8b5cf6', // Purple
  '#ec4899', // Pink
  '#ffffff', // White
  '#6b7280', // Gray
];

const LINE_WIDTHS = [1, 2, 3, 4, 5];

export const DrawingPropertiesPanel: React.FC = () => {
  const {
    isPanelOpen,
    selectedDrawingId,
    properties,
    updateProperty,
    closePanel,
  } = useDrawingPropertiesStore();

  // Listen for drawing selection events from drawingManager
  useEffect(() => {
    const handleDrawingSelected = (event: CustomEvent) => {
      const drawingId = event.detail?.drawingId;
      if (drawingId) {
        const drawing = drawingManager.getDrawings().find(d => d.id === drawingId);
        if (drawing) {
          useDrawingPropertiesStore.getState().setSelectedDrawing(drawing);
        }
      }
    };

    window.addEventListener('drawing:selected', handleDrawingSelected as EventListener);
    return () => {
      window.removeEventListener('drawing:selected', handleDrawingSelected as EventListener);
    };
  }, []);

  // Apply property changes to the selected drawing
  useEffect(() => {
    if (!selectedDrawingId) return;

    const drawing = drawingManager.getDrawings().find(d => d.id === selectedDrawingId);
    if (!drawing) return;

    // Update drawing properties
    drawing.color = properties.color;
    drawing.lineWidth = properties.lineWidth;
    drawing.lineStyle = properties.lineStyle;
    drawing.text = properties.text;
    drawing.visible = properties.visible !== false;

    // Trigger re-render and backend save
    drawingManager.renderAllDrawings();
    drawingManager.saveToBackend(drawing);
  }, [
    selectedDrawingId,
    properties.color,
    properties.lineWidth,
    properties.lineStyle,
    properties.text,
    properties.visible,
  ]);

  const handleDuplicate = () => {
    if (!selectedDrawingId) return;
    const drawing = drawingManager.getDrawings().find(d => d.id === selectedDrawingId);
    if (drawing) {
      // Create duplicate with offset
      const duplicateId = drawingManager.startDrawing(
        drawing.type,
        drawing.color || '#3b82f6',
        drawing.subtype
      );

      // Copy properties
      drawingManager.updateActiveDrawing({
        color: drawing.color,
        lineWidth: drawing.lineWidth,
        lineStyle: drawing.lineStyle,
        text: drawing.text,
      });

      // Copy points with offset
      drawing.points.forEach(point => {
        drawingManager.addPoint(point.time + 3600, point.price); // 1 hour offset
      });

      drawingManager.renderAllDrawings();
    }
  };

  const handleDelete = () => {
    if (!selectedDrawingId) return;
    drawingManager.deleteDrawing(selectedDrawingId);
    closePanel();
  };

  if (!isPanelOpen || !selectedDrawingId) return null;

  const currentDrawing = drawingManager.getDrawings().find(d => d.id === selectedDrawingId);
  if (!currentDrawing) return null;

  const isTrendline = currentDrawing.type === 'trendline';
  const isFibonacci = currentDrawing.type === 'fibonacci';
  const isText = currentDrawing.type === 'text';

  return (
    <div
      className="fixed right-4 top-20 z-[999] bg-[#1a1a1a] border border-zinc-800 rounded-lg shadow-2xl w-72 animate-in slide-in-from-right duration-200"
      onClick={(e) => e.stopPropagation()}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-zinc-800">
        <h3 className="text-sm font-semibold text-zinc-200">Drawing Properties</h3>
        <button
          onClick={closePanel}
          className="text-zinc-500 hover:text-zinc-300 transition-colors"
        >
          <X size={16} />
        </button>
      </div>

      {/* Content */}
      <div className="p-4 space-y-4 max-h-[70vh] overflow-y-auto">
        {/* Color Picker */}
        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-2">
            Color
          </label>
          <div className="grid grid-cols-9 gap-2">
            {COLOR_PALETTE.map((color) => (
              <button
                key={color}
                onClick={() => updateProperty('color', color)}
                className={`w-7 h-7 rounded border-2 transition-all hover:scale-110 ${
                  properties.color === color
                    ? 'border-white ring-2 ring-blue-500'
                    : 'border-zinc-700'
                }`}
                style={{ backgroundColor: color }}
                title={color}
              />
            ))}
          </div>
          {/* Custom color input */}
          <div className="mt-2 flex items-center gap-2">
            <input
              type="color"
              value={properties.color}
              onChange={(e) => updateProperty('color', e.target.value)}
              className="w-10 h-8 rounded cursor-pointer bg-transparent"
            />
            <input
              type="text"
              value={properties.color}
              onChange={(e) => updateProperty('color', e.target.value)}
              className="flex-1 bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              placeholder="#3b82f6"
            />
          </div>
        </div>

        {/* Line Style */}
        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-2">
            Line Style
          </label>
          <div className="grid grid-cols-3 gap-2">
            {(['solid', 'dashed', 'dotted'] as const).map((style) => (
              <button
                key={style}
                onClick={() => updateProperty('lineStyle', style)}
                className={`px-3 py-2 rounded text-xs font-medium transition-colors ${
                  properties.lineStyle === style
                    ? 'bg-blue-600 text-white'
                    : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
                }`}
              >
                {style.charAt(0).toUpperCase() + style.slice(1)}
              </button>
            ))}
          </div>
        </div>

        {/* Line Width */}
        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-2">
            Line Width: {properties.lineWidth}px
          </label>
          <div className="flex items-center gap-2">
            <input
              type="range"
              min="1"
              max="5"
              step="1"
              value={properties.lineWidth}
              onChange={(e) => updateProperty('lineWidth', parseInt(e.target.value))}
              className="flex-1 h-2 bg-zinc-800 rounded-lg appearance-none cursor-pointer accent-blue-500"
            />
            <div className="flex gap-1">
              {LINE_WIDTHS.map((width) => (
                <button
                  key={width}
                  onClick={() => updateProperty('lineWidth', width)}
                  className={`w-8 h-8 flex items-center justify-center rounded text-xs font-medium transition-colors ${
                    properties.lineWidth === width
                      ? 'bg-blue-600 text-white'
                      : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
                  }`}
                >
                  {width}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Trendline Extensions */}
        {isTrendline && (
          <div className="space-y-2">
            <label className="block text-xs font-medium text-zinc-400">
              Extensions
            </label>
            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 text-xs text-zinc-300 cursor-pointer">
                <input
                  type="checkbox"
                  checked={properties.extendLeft || false}
                  onChange={(e) => updateProperty('extendLeft', e.target.checked)}
                  className="w-4 h-4 rounded border-zinc-700 bg-zinc-800 text-blue-600 focus:ring-blue-500 focus:ring-offset-0"
                />
                Extend Left
              </label>
              <label className="flex items-center gap-2 text-xs text-zinc-300 cursor-pointer">
                <input
                  type="checkbox"
                  checked={properties.extendRight || false}
                  onChange={(e) => updateProperty('extendRight', e.target.checked)}
                  className="w-4 h-4 rounded border-zinc-700 bg-zinc-800 text-blue-600 focus:ring-blue-500 focus:ring-offset-0"
                />
                Extend Right
              </label>
            </div>
          </div>
        )}

        {/* Fibonacci Price Labels */}
        {isFibonacci && (
          <div>
            <label className="flex items-center gap-2 text-xs text-zinc-300 cursor-pointer">
              <input
                type="checkbox"
                checked={properties.showPriceLabels || false}
                onChange={(e) => updateProperty('showPriceLabels', e.target.checked)}
                className="w-4 h-4 rounded border-zinc-700 bg-zinc-800 text-blue-600 focus:ring-blue-500 focus:ring-offset-0"
              />
              Show Price Labels
            </label>
          </div>
        )}

        {/* Text Input for Labels */}
        {isText && (
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-2">
              Text Label
            </label>
            <textarea
              value={properties.text || ''}
              onChange={(e) => updateProperty('text', e.target.value)}
              placeholder="Enter text..."
              rows={3}
              className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-300 focus:outline-none focus:border-blue-500 resize-none"
            />
          </div>
        )}

        {/* Visibility Toggle */}
        <div>
          <label className="flex items-center gap-2 text-xs text-zinc-300 cursor-pointer">
            <input
              type="checkbox"
              checked={properties.visible !== false}
              onChange={(e) => updateProperty('visible', e.target.checked)}
              className="w-4 h-4 rounded border-zinc-700 bg-zinc-800 text-blue-600 focus:ring-blue-500 focus:ring-offset-0"
            />
            {properties.visible !== false ? <Eye size={14} /> : <EyeOff size={14} />}
            <span>Visible on Chart</span>
          </label>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-2 pt-2 border-t border-zinc-800">
          <button
            onClick={handleDuplicate}
            className="flex-1 flex items-center justify-center gap-2 px-3 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 rounded text-xs font-medium transition-colors"
          >
            <Copy size={14} />
            Duplicate
          </button>
          <button
            onClick={handleDelete}
            className="flex-1 flex items-center justify-center gap-2 px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded text-xs font-medium transition-colors"
          >
            <Trash2 size={14} />
            Delete
          </button>
        </div>
      </div>
    </div>
  );
};
