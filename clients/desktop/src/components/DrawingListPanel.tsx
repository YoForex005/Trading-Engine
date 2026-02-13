import React, { useState, useEffect } from 'react';
import {
  Eye,
  EyeOff,
  Trash2,
  X,
  Minus,
  Type,
  TrendingUp,
  Square,
  Circle,
  ArrowUpRight,
  Grid3x3
} from 'lucide-react';
import { drawingManager } from '../services/drawingManager';
import type { Drawing, DrawingType } from '../services/drawingManager';

interface DrawingListPanelProps {
  visible: boolean;
  onClose: () => void;
}

export const DrawingListPanel: React.FC<DrawingListPanelProps> = ({
  visible,
  onClose
}) => {
  const [drawings, setDrawings] = useState<Drawing[]>([]);
  const [hiddenDrawings, setHiddenDrawings] = useState<Set<string>>(new Set());

  // Load drawings on mount and listen for changes
  useEffect(() => {
    const updateDrawings = () => {
      const currentDrawings = drawingManager.getDrawings();
      setDrawings([...currentDrawings].reverse()); // Reverse chronological order
    };

    // Initial load
    updateDrawings();

    // Listen for drawing events
    const handleDrawingEvent = () => {
      updateDrawings();
    };

    window.addEventListener('drawing:saved', handleDrawingEvent);
    window.addEventListener('drawing:deleted', handleDrawingEvent);
    window.addEventListener('drawings:cleared', handleDrawingEvent);

    return () => {
      window.removeEventListener('drawing:saved', handleDrawingEvent);
      window.removeEventListener('drawing:deleted', handleDrawingEvent);
      window.removeEventListener('drawings:cleared', handleDrawingEvent);
    };
  }, []);

  // Toggle visibility of a drawing
  const toggleVisibility = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const newHidden = new Set(hiddenDrawings);
    if (newHidden.has(id)) {
      newHidden.delete(id);
    } else {
      newHidden.add(id);
    }
    setHiddenDrawings(newHidden);

    // Hide/show the drawing element
    const element = document.querySelector(`[data-drawing-id="${id}"]`) as HTMLElement;
    if (element) {
      element.style.display = newHidden.has(id) ? 'none' : '';
    }

    // Also hide/show nodes
    const nodes = document.querySelectorAll(`.drawing-node-${id}`);
    nodes.forEach(node => {
      const nodeElement = node as HTMLElement;
      if (newHidden.has(id)) {
        nodeElement.style.display = 'none';
      } else {
        // Restore original display based on selection state
        const drawing = drawings.find(d => d.id === id);
        nodeElement.style.display = drawing?.selected ? 'block' : 'none';
      }
    });
  };

  // Delete a drawing
  const deleteDrawing = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (confirm('Delete this drawing?')) {
      drawingManager.deleteDrawing(id);
    }
  };

  // Select a drawing on the chart
  const selectDrawing = (id: string) => {
    drawingManager.selectDrawing(id);
  };

  // Get icon for drawing type
  const getDrawingIcon = (drawing: Drawing) => {
    const iconClass = "w-4 h-4";
    const color = drawing.color || '#3b82f6';

    switch (drawing.type) {
      case 'hline':
        return <Minus className={iconClass} style={{ color }} />;
      case 'vline':
        return <div className={iconClass} style={{ backgroundColor: color, width: '2px', height: '16px' }} />;
      case 'trendline':
        return <TrendingUp className={iconClass} style={{ color }} />;
      case 'text':
        return <Type className={iconClass} style={{ color }} />;
      case 'rectangle':
        return <Square className={iconClass} style={{ color }} />;
      case 'ellipse':
        return <Circle className={iconClass} style={{ color }} />;
      case 'arrow':
        return <ArrowUpRight className={iconClass} style={{ color }} />;
      case 'channel':
        return (
          <svg width="16" height="16" viewBox="0 0 18 18" fill="none" stroke={color} strokeWidth="2">
            <path d="M3 12L12 3M6 15L15 6" />
          </svg>
        );
      case 'fibonacci':
        return (
          <svg width="16" height="16" viewBox="0 0 18 18" fill="none" stroke={color} strokeWidth="1.5">
            <path d="M2 16L16 2M2 13h14M2 10h14M2 7h14M2 4h14" />
          </svg>
        );
      case 'pitchfork':
        return (
          <svg width="16" height="16" viewBox="0 0 18 18" fill="none" stroke={color} strokeWidth="1.5">
            <path d="M9 2v14M9 2l-4 6M9 2l4 6M5 16l4-8l4 8" />
          </svg>
        );
      case 'shapes':
        return <span style={{ color, fontSize: '14px' }}>{getShapeSymbol(drawing.subtype)}</span>;
      default:
        return <Grid3x3 className={iconClass} style={{ color }} />;
    }
  };

  // Get shape symbol
  const getShapeSymbol = (subtype?: string): string => {
    switch (subtype) {
      case 'thumbs_up': return '👍';
      case 'thumbs_down': return '👎';
      case 'arrow_up': return '↑';
      case 'arrow_down': return '↓';
      case 'stop': return '🛑';
      case 'check': return '✅';
      case 'buy': return '▲';
      case 'sell': return '▼';
      case 'arrow': return '➔';
      default: return '●';
    }
  };

  // Get drawing label
  const getDrawingLabel = (drawing: Drawing): string => {
    if (drawing.type === 'text' && drawing.text) {
      return `Text: ${drawing.text.substring(0, 20)}${drawing.text.length > 20 ? '...' : ''}`;
    }
    if (drawing.type === 'shapes' && drawing.subtype) {
      return drawing.subtype.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
    }
    return drawing.type.charAt(0).toUpperCase() + drawing.type.slice(1).replace(/([A-Z])/g, ' $1');
  };

  if (!visible) return null;

  return (
    <div className="absolute left-4 top-20 z-50 w-64 bg-[#1e1e1e] border border-zinc-800 rounded-lg shadow-2xl overflow-hidden animate-in slide-in-from-left duration-200">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 bg-[#252525] border-b border-zinc-800">
        <div className="flex items-center gap-2">
          <Grid3x3 size={14} className="text-blue-400" />
          <span className="text-xs font-bold text-zinc-200">Drawing List</span>
          <span className="text-[10px] text-zinc-500 bg-zinc-800 px-1.5 py-0.5 rounded">
            {drawings.length}
          </span>
        </div>
        <button
          onClick={onClose}
          className="text-zinc-500 hover:text-zinc-300 transition-colors"
        >
          <X size={14} />
        </button>
      </div>

      {/* Drawing List */}
      <div className="max-h-96 overflow-y-auto custom-scrollbar">
        {drawings.length === 0 ? (
          <div className="px-3 py-8 text-center">
            <Grid3x3 size={32} className="mx-auto text-zinc-700 mb-2" />
            <p className="text-xs text-zinc-500">No drawings yet</p>
            <p className="text-[10px] text-zinc-600 mt-1">
              Use drawing tools to add shapes
            </p>
          </div>
        ) : (
          <div className="py-1">
            {drawings.map((drawing) => {
              const isHidden = hiddenDrawings.has(drawing.id);
              const isSelected = drawing.selected;

              return (
                <div
                  key={drawing.id}
                  onClick={() => selectDrawing(drawing.id)}
                  className={`
                    flex items-center gap-2 px-3 py-2 cursor-pointer transition-all
                    ${isSelected
                      ? 'bg-blue-500/10 border-l-2 border-blue-500'
                      : 'hover:bg-[#252525] border-l-2 border-transparent'
                    }
                    ${isHidden ? 'opacity-50' : ''}
                  `}
                >
                  {/* Icon with color preview */}
                  <div className="flex-shrink-0">
                    {getDrawingIcon(drawing)}
                  </div>

                  {/* Label */}
                  <div className="flex-1 min-w-0">
                    <p className="text-xs text-zinc-300 truncate">
                      {getDrawingLabel(drawing)}
                    </p>
                    {drawing.points.length > 0 && (
                      <p className="text-[10px] text-zinc-600">
                        {drawing.points.length} point{drawing.points.length > 1 ? 's' : ''}
                      </p>
                    )}
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-1">
                    {/* Visibility toggle */}
                    <button
                      onClick={(e) => toggleVisibility(drawing.id, e)}
                      className="p-1 hover:bg-zinc-700 rounded transition-colors"
                      title={isHidden ? 'Show' : 'Hide'}
                    >
                      {isHidden ? (
                        <EyeOff size={12} className="text-zinc-500" />
                      ) : (
                        <Eye size={12} className="text-zinc-400" />
                      )}
                    </button>

                    {/* Delete */}
                    <button
                      onClick={(e) => deleteDrawing(drawing.id, e)}
                      className="p-1 hover:bg-red-500/10 rounded transition-colors group"
                      title="Delete"
                    >
                      <Trash2 size={12} className="text-zinc-500 group-hover:text-red-400" />
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Footer */}
      {drawings.length > 0 && (
        <div className="px-3 py-2 bg-[#252525] border-t border-zinc-800">
          <button
            onClick={() => {
              if (confirm('Clear all drawings?')) {
                drawingManager.clearAllDrawings();
              }
            }}
            className="w-full px-2 py-1 text-[10px] text-zinc-400 hover:text-red-400 hover:bg-red-500/10 rounded transition-colors"
          >
            Clear All
          </button>
        </div>
      )}
    </div>
  );
};
