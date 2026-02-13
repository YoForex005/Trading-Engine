/**
 * DrawingTools Component
 * Advanced chart drawing tools panel
 */

import React, { useState, useRef, useCallback, useMemo } from 'react';
import {
  TrendingUp,
  Minus,
  ArrowRight,
  Square,
  Circle,
  Triangle,
  Type,
  ArrowUpRight,
  Crosshair,
  Route,
  TrendingDown,
  GitBranch,
  Eye,
  EyeOff,
  Trash2,
  Copy,
  Layers,
  Download,
  Upload,
  Save,
  Palette,
  Settings,
  X,
} from 'lucide-react';
import {
  useDrawingStore,
  type DrawingType,
  type DrawingPoint,
  type StoreDrawing,
} from '../store/useDrawingStore';

// Drawing tool configuration - mapped to canonical DrawingType values
const DRAWING_TOOLS = [
  { type: 'trendline' as DrawingType, label: 'Trend Line', icon: TrendingUp, shortcut: 'T' },
  { type: 'hline' as DrawingType, label: 'Horizontal Line', icon: Minus, shortcut: 'H' },
  { type: 'vline' as DrawingType, label: 'Vertical Line', icon: ArrowRight, shortcut: 'V' },
  { type: 'channel' as DrawingType, label: 'Channel', icon: GitBranch, shortcut: 'C' },
  { type: 'fibonacci' as DrawingType, label: 'Fibonacci Retracement', icon: TrendingDown, shortcut: 'F' },
  { type: 'rectangle' as DrawingType, label: 'Rectangle', icon: Square, shortcut: 'R' },
  { type: 'ellipse' as DrawingType, label: 'Ellipse', icon: Circle, shortcut: 'O' },
  { type: 'pitchfork' as DrawingType, label: 'Pitchfork', icon: Triangle, shortcut: 'P' },
  { type: 'text' as DrawingType, label: 'Text Label', icon: Type, shortcut: 'L' },
  { type: 'arrow' as DrawingType, label: 'Arrow', icon: ArrowUpRight, shortcut: 'A' },
  { type: 'shapes' as DrawingType, label: 'Shapes', icon: Crosshair, shortcut: 'X' },
];

const COLORS = [
  '#22c55e', // green
  '#ef4444', // red
  '#3b82f6', // blue
  '#f59e0b', // amber
  '#8b5cf6', // purple
  '#ec4899', // pink
  '#06b6d4', // cyan
  '#ffffff', // white
  '#71717a', // gray
];

type LineStyle = 'solid' | 'dashed' | 'dotted';
const LINE_STYLES: LineStyle[] = ['solid', 'dashed', 'dotted'];
const LINE_WIDTHS = [1, 2, 3, 4, 5];

export function DrawingTools() {
  const {
    drawings,
    selectedDrawingId,
    activeTool,
    templates,
    isDrawing,
    addDrawing,
    updateDrawing,
    deleteDrawing,
    duplicateDrawing,
    setSelectedDrawing,
    setActiveTool,
    setIsDrawing,
    toggleDrawingVisibility,
    bringToFront,
    sendToBack,
    clearAllDrawings,
    saveTemplate,
    loadTemplate,
    deleteTemplate,
  } = useDrawingStore();

  const [showProperties, setShowProperties] = useState(false);
  const [showTemplates, setShowTemplates] = useState(false);
  const [templateName, setTemplateName] = useState('');
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; drawingId: string } | null>(null);
  const [currentPoints, setCurrentPoints] = useState<DrawingPoint[]>([]);

  const canvasRef = useRef<SVGSVGElement>(null);
  const [svgDimensions] = useState({ width: 800, height: 500 });

  // Selected drawing
  const selectedDrawing = useMemo(
    () => drawings.find(d => d.id === selectedDrawingId),
    [drawings, selectedDrawingId]
  );

  // Handle tool selection
  const handleToolSelect = useCallback((tool: DrawingType) => {
    setActiveTool(activeTool === tool ? null : tool);
    setSelectedDrawing(null);
  }, [activeTool, setActiveTool, setSelectedDrawing]);

  // Handle canvas mouse events
  const handleCanvasMouseDown = useCallback((e: React.MouseEvent<SVGSVGElement>) => {
    if (!activeTool || !canvasRef.current) return;

    const rect = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    // Calculate mock price based on y position (inverse, top = higher price)
    const price = 1.1000 - (y / svgDimensions.height) * 0.02; // Range: 1.08 to 1.10

    const point: DrawingPoint = { time: Math.floor(Date.now() / 1000), price };

    setIsDrawing(true);
    setCurrentPoints([point]);
  }, [activeTool, setIsDrawing, svgDimensions.height]);

  const handleCanvasMouseMove = useCallback((e: React.MouseEvent<SVGSVGElement>) => {
    if (!isDrawing || !canvasRef.current || currentPoints.length === 0) return;

    const rect = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    const price = 1.1000 - (y / svgDimensions.height) * 0.02;

    const point: DrawingPoint = { time: Math.floor(Date.now() / 1000), price };

    // Update preview
    setCurrentPoints([currentPoints[0], point]);
  }, [isDrawing, currentPoints, svgDimensions.height]);

  const handleCanvasMouseUp = useCallback(() => {
    if (!isDrawing || !activeTool || currentPoints.length < 2) {
      setIsDrawing(false);
      setCurrentPoints([]);
      return;
    }

    // Create the drawing
    const newDrawing: Omit<StoreDrawing, 'id' | 'zIndex'> = {
      type: activeTool,
      points: currentPoints,
      color: '#22c55e',
      lineWidth: 2,
      lineStyle: 'solid',
      extendLeft: false,
      extendRight: false,
      showPriceLabels: true,
      visible: true,
      text: activeTool === 'text' ? 'New Label' : undefined,
      label: DRAWING_TOOLS.find(t => t.type === activeTool)?.label,
    };

    addDrawing(newDrawing);
    setIsDrawing(false);
    setCurrentPoints([]);
    setActiveTool(null); // Deactivate tool after drawing
  }, [isDrawing, activeTool, currentPoints, addDrawing, setIsDrawing, setActiveTool]);

  // Context menu
  const handleDrawingRightClick = useCallback((e: React.MouseEvent, drawingId: string) => {
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY, drawingId });
  }, []);

  const handleContextMenuAction = useCallback((action: 'edit' | 'duplicate' | 'delete' | 'front' | 'back') => {
    if (!contextMenu) return;

    switch (action) {
      case 'edit':
        setSelectedDrawing(contextMenu.drawingId);
        setShowProperties(true);
        break;
      case 'duplicate':
        duplicateDrawing(contextMenu.drawingId);
        break;
      case 'delete':
        deleteDrawing(contextMenu.drawingId);
        break;
      case 'front':
        bringToFront(contextMenu.drawingId);
        break;
      case 'back':
        sendToBack(contextMenu.drawingId);
        break;
    }

    setContextMenu(null);
  }, [contextMenu, setSelectedDrawing, duplicateDrawing, deleteDrawing, bringToFront, sendToBack]);

  // Render drawing on canvas (convert DrawingPoint to screen coordinates)
  const renderDrawing = useCallback((drawing: StoreDrawing) => {
    if (!drawing.visible || drawing.points.length < 2) return null;

    // Convert DrawingPoint { time, price } to screen coordinates
    const toScreenCoords = (point: DrawingPoint) => {
      const x = (point.time % 1000) * (svgDimensions.width / 1000);
      const y = svgDimensions.height - ((point.price - 1.08) / 0.02) * svgDimensions.height;
      return { x, y };
    };

    const screenPoints = drawing.points.map(toScreenCoords);
    const [p1, p2] = screenPoints;
    const strokeDasharray =
      drawing.lineStyle === 'dashed' ? '8,4' :
      drawing.lineStyle === 'dotted' ? '2,2' :
      undefined;

    switch (drawing.type) {
      case 'trendline':
      case 'hline':
      case 'vline':
      case 'arrow':
        return (
          <line
            key={drawing.id}
            x1={p1.x}
            y1={p1.y}
            x2={p2.x}
            y2={p2.y}
            stroke={drawing.color}
            strokeWidth={drawing.lineWidth}
            strokeDasharray={strokeDasharray}
            onContextMenu={(e) => handleDrawingRightClick(e, drawing.id)}
            onClick={() => setSelectedDrawing(drawing.id)}
            style={{ cursor: 'pointer' }}
          />
        );

      case 'rectangle':
        return (
          <rect
            key={drawing.id}
            x={Math.min(p1.x, p2.x)}
            y={Math.min(p1.y, p2.y)}
            width={Math.abs(p2.x - p1.x)}
            height={Math.abs(p2.y - p1.y)}
            stroke={drawing.color}
            strokeWidth={drawing.lineWidth}
            fill="none"
            strokeDasharray={strokeDasharray}
            onContextMenu={(e) => handleDrawingRightClick(e, drawing.id)}
            onClick={() => setSelectedDrawing(drawing.id)}
            style={{ cursor: 'pointer' }}
          />
        );

      case 'ellipse':
        const cx = (p1.x + p2.x) / 2;
        const cy = (p1.y + p2.y) / 2;
        const rx = Math.abs(p2.x - p1.x) / 2;
        const ry = Math.abs(p2.y - p1.y) / 2;
        return (
          <ellipse
            key={drawing.id}
            cx={cx}
            cy={cy}
            rx={rx}
            ry={ry}
            stroke={drawing.color}
            strokeWidth={drawing.lineWidth}
            fill="none"
            strokeDasharray={strokeDasharray}
            onContextMenu={(e) => handleDrawingRightClick(e, drawing.id)}
            onClick={() => setSelectedDrawing(drawing.id)}
            style={{ cursor: 'pointer' }}
          />
        );

      case 'channel':
        // Draw two parallel lines
        const dy = p2.y - p1.y;
        const offset = 30;
        return (
          <g key={drawing.id}>
            <line
              x1={p1.x}
              y1={p1.y}
              x2={p2.x}
              y2={p2.y}
              stroke={drawing.color}
              strokeWidth={drawing.lineWidth}
              strokeDasharray={strokeDasharray}
            />
            <line
              x1={p1.x}
              y1={p1.y + offset}
              x2={p2.x}
              y2={p2.y + offset}
              stroke={drawing.color}
              strokeWidth={drawing.lineWidth}
              strokeDasharray={strokeDasharray}
              onContextMenu={(e) => handleDrawingRightClick(e, drawing.id)}
              onClick={() => setSelectedDrawing(drawing.id)}
              style={{ cursor: 'pointer' }}
            />
          </g>
        );

      case 'fibonacci':
        // Draw Fibonacci retracement levels
        const fibLevels = [0, 0.236, 0.382, 0.5, 0.618, 0.786, 1];
        const height = p2.y - p1.y;
        return (
          <g key={drawing.id}>
            {fibLevels.map((level, idx) => {
              const y = p1.y + height * level;
              return (
                <g key={idx}>
                  <line
                    x1={Math.min(p1.x, p2.x)}
                    y1={y}
                    x2={Math.max(p1.x, p2.x)}
                    y2={y}
                    stroke={drawing.color}
                    strokeWidth={drawing.lineWidth}
                    strokeDasharray={strokeDasharray}
                    opacity={0.7}
                  />
                  {drawing.showPriceLabels && (
                    <text
                      x={Math.max(p1.x, p2.x) + 5}
                      y={y + 4}
                      fill={drawing.color}
                      fontSize="11"
                      fontFamily="monospace"
                    >
                      {(level * 100).toFixed(1)}%
                    </text>
                  )}
                </g>
              );
            })}
            <rect
              x={Math.min(p1.x, p2.x)}
              y={Math.min(p1.y, p2.y)}
              width={Math.abs(p2.x - p1.x)}
              height={Math.abs(p2.y - p1.y)}
              fill="none"
              stroke="transparent"
              strokeWidth={10}
              onContextMenu={(e) => handleDrawingRightClick(e, drawing.id)}
              onClick={() => setSelectedDrawing(drawing.id)}
              style={{ cursor: 'pointer' }}
            />
          </g>
        );

      case 'text':
        return (
          <text
            key={drawing.id}
            x={p1.x}
            y={p1.y}
            fill={drawing.color}
            fontSize="14"
            fontWeight="bold"
            fontFamily="sans-serif"
            onContextMenu={(e) => handleDrawingRightClick(e, drawing.id)}
            onClick={() => setSelectedDrawing(drawing.id)}
            style={{ cursor: 'pointer' }}
          >
            {drawing.text || 'Label'}
          </text>
        );

      default:
        return null;
    }
  }, [handleDrawingRightClick, setSelectedDrawing, svgDimensions]);

  // Render current drawing preview
  const renderCurrentDrawing = useCallback(() => {
    if (!isDrawing || currentPoints.length < 2 || !activeTool) return null;

    // Convert DrawingPoint to screen coordinates
    const toScreenCoords = (point: DrawingPoint) => {
      const x = (point.time % 1000) * (svgDimensions.width / 1000);
      const y = svgDimensions.height - ((point.price - 1.08) / 0.02) * svgDimensions.height;
      return { x, y };
    };

    const screenPoints = currentPoints.map(toScreenCoords);
    const [p1, p2] = screenPoints;

    switch (activeTool) {
      case 'trendline':
      case 'hline':
      case 'vline':
      case 'arrow':
        return (
          <line
            x1={p1.x}
            y1={p1.y}
            x2={p2.x}
            y2={p2.y}
            stroke="#22c55e"
            strokeWidth={2}
            strokeDasharray="4,4"
            opacity={0.6}
          />
        );

      case 'rectangle':
        return (
          <rect
            x={Math.min(p1.x, p2.x)}
            y={Math.min(p1.y, p2.y)}
            width={Math.abs(p2.x - p1.x)}
            height={Math.abs(p2.y - p1.y)}
            stroke="#22c55e"
            strokeWidth={2}
            fill="none"
            strokeDasharray="4,4"
            opacity={0.6}
          />
        );

      case 'ellipse':
        const cx = (p1.x + p2.x) / 2;
        const cy = (p1.y + p2.y) / 2;
        const rx = Math.abs(p2.x - p1.x) / 2;
        const ry = Math.abs(p2.y - p1.y) / 2;
        return (
          <ellipse
            cx={cx}
            cy={cy}
            rx={rx}
            ry={ry}
            stroke="#22c55e"
            strokeWidth={2}
            fill="none"
            strokeDasharray="4,4"
            opacity={0.6}
          />
        );

      default:
        return null;
    }
  }, [isDrawing, currentPoints, activeTool, svgDimensions]);

  return (
    <div className="flex flex-col h-full bg-zinc-900 text-white">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 bg-zinc-800 border-b border-zinc-700">
        <div className="flex items-center gap-2">
          <Palette className="w-4 h-4 text-emerald-400" />
          <h2 className="text-sm font-semibold">Drawing Tools</h2>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowTemplates(!showTemplates)}
            className="px-2 py-1 text-xs bg-zinc-700 hover:bg-zinc-600 rounded flex items-center gap-1"
            title="Templates"
          >
            <Download className="w-3 h-3" />
            Templates
          </button>
          <button
            onClick={() => setShowProperties(!showProperties)}
            className="px-2 py-1 text-xs bg-zinc-700 hover:bg-zinc-600 rounded flex items-center gap-1"
            title="Properties"
          >
            <Settings className="w-3 h-3" />
            Properties
          </button>
          <button
            onClick={clearAllDrawings}
            className="px-2 py-1 text-xs bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded"
            title="Clear All"
          >
            <Trash2 className="w-3 h-3" />
          </button>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        {/* Left Sidebar - Toolbar */}
        <div className="w-20 bg-zinc-800 border-r border-zinc-700 overflow-y-auto">
          <div className="p-2 space-y-1">
            {DRAWING_TOOLS.map((tool) => {
              const Icon = tool.icon;
              const isActive = activeTool === tool.type;
              return (
                <button
                  key={tool.type}
                  onClick={() => handleToolSelect(tool.type)}
                  className={`w-full p-2 rounded flex flex-col items-center justify-center gap-1 transition-colors ${
                    isActive
                      ? 'bg-emerald-600 text-white'
                      : 'bg-zinc-700 hover:bg-zinc-600 text-zinc-300'
                  }`}
                  title={`${tool.label} (${tool.shortcut})`}
                >
                  <Icon className="w-5 h-5" />
                  <span className="text-[9px] font-medium">{tool.shortcut}</span>
                </button>
              );
            })}
          </div>
        </div>

        {/* Main Canvas Area */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Canvas */}
          <div className="flex-1 bg-zinc-900 overflow-auto p-4">
            <div className="inline-block">
              <svg
                ref={canvasRef}
                width={svgDimensions.width}
                height={svgDimensions.height}
                className="bg-zinc-950 border border-zinc-700 rounded cursor-crosshair"
                onMouseDown={handleCanvasMouseDown}
                onMouseMove={handleCanvasMouseMove}
                onMouseUp={handleCanvasMouseUp}
                onClick={() => !isDrawing && setContextMenu(null)}
              >
                {/* Grid */}
                <defs>
                  <pattern id="grid" width="20" height="20" patternUnits="userSpaceOnUse">
                    <path d="M 20 0 L 0 0 0 20" fill="none" stroke="#27272a" strokeWidth="0.5" />
                  </pattern>
                </defs>
                <rect width="100%" height="100%" fill="url(#grid)" />

                {/* Mock price levels */}
                {[...Array(6)].map((_, i) => {
                  const y = (i + 1) * (svgDimensions.height / 7);
                  const price = (1.1000 - (y / svgDimensions.height) * 0.02).toFixed(5);
                  return (
                    <g key={i}>
                      <line
                        x1={0}
                        y1={y}
                        x2={svgDimensions.width}
                        y2={y}
                        stroke="#3f3f46"
                        strokeWidth="0.5"
                        strokeDasharray="2,2"
                      />
                      <text x="5" y={y - 3} fill="#71717a" fontSize="10" fontFamily="monospace">
                        {price}
                      </text>
                    </g>
                  );
                })}

                {/* Render all drawings sorted by z-index */}
                {drawings
                  .slice()
                  .sort((a, b) => (a.zIndex || 0) - (b.zIndex || 0))
                  .map(renderDrawing)}

                {/* Render current drawing preview */}
                {renderCurrentDrawing()}
              </svg>
            </div>
          </div>

          {/* Info Bar */}
          <div className="px-4 py-2 bg-zinc-800 border-t border-zinc-700 flex items-center justify-between text-xs">
            <div className="flex items-center gap-4">
              <span className="text-zinc-400">
                Active Tool: <span className="text-white">{activeTool ? DRAWING_TOOLS.find(t => t.type === activeTool)?.label : 'None'}</span>
              </span>
              <span className="text-zinc-400">
                Drawings: <span className="text-white">{drawings.length}</span>
              </span>
            </div>
            <div className="text-zinc-400">
              Click and drag to draw • Right-click for options
            </div>
          </div>
        </div>

        {/* Right Sidebar - Drawing List */}
        <div className="w-64 bg-zinc-800 border-l border-zinc-700 overflow-y-auto">
          <div className="p-3">
            <h3 className="text-xs font-semibold mb-2 text-zinc-400">DRAWINGS</h3>
            <div className="space-y-1">
              {drawings.length === 0 ? (
                <div className="text-xs text-zinc-500 text-center py-4">
                  No drawings yet
                </div>
              ) : (
                drawings
                  .slice()
                  .sort((a, b) => (b.zIndex || 0) - (a.zIndex || 0))
                  .map((drawing) => {
                    const tool = DRAWING_TOOLS.find(t => t.type === drawing.type);
                    const Icon = tool?.icon || Layers;
                    return (
                      <div
                        key={drawing.id}
                        className={`flex items-center gap-2 p-2 rounded text-xs ${
                          selectedDrawingId === drawing.id
                            ? 'bg-emerald-600/20 border border-emerald-600/50'
                            : 'bg-zinc-700 hover:bg-zinc-600'
                        }`}
                        onClick={() => setSelectedDrawing(drawing.id)}
                      >
                        <Icon className="w-3 h-3 flex-shrink-0" style={{ color: drawing.color }} />
                        <span className="flex-1 truncate">{drawing.label || tool?.label}</span>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            toggleDrawingVisibility(drawing.id);
                          }}
                          className="p-1 hover:bg-zinc-500 rounded"
                          title={drawing.visible ? 'Hide' : 'Show'}
                        >
                          {drawing.visible ? (
                            <Eye className="w-3 h-3" />
                          ) : (
                            <EyeOff className="w-3 h-3 text-zinc-500" />
                          )}
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            deleteDrawing(drawing.id);
                          }}
                          className="p-1 hover:bg-red-600/30 rounded text-red-400"
                          title="Delete"
                        >
                          <Trash2 className="w-3 h-3" />
                        </button>
                      </div>
                    );
                  })
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Properties Panel */}
      {showProperties && selectedDrawing && (
        <div className="absolute top-16 right-4 w-72 bg-zinc-800 border border-zinc-700 rounded-lg shadow-2xl z-50">
          <div className="flex items-center justify-between px-4 py-2 bg-zinc-900 border-b border-zinc-700 rounded-t-lg">
            <h3 className="text-sm font-semibold">Drawing Properties</h3>
            <button
              onClick={() => setShowProperties(false)}
              className="p-1 hover:bg-zinc-700 rounded"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
          <div className="p-4 space-y-4">
            {/* Color */}
            <div>
              <label className="text-xs font-medium text-zinc-400 mb-2 block">Color</label>
              <div className="flex gap-2 flex-wrap">
                {COLORS.map((color) => (
                  <button
                    key={color}
                    onClick={() => updateDrawing(selectedDrawing.id, { color })}
                    className={`w-8 h-8 rounded border-2 ${
                      selectedDrawing.color === color ? 'border-white' : 'border-zinc-600'
                    }`}
                    style={{ backgroundColor: color }}
                    title={color}
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
                    onClick={() => updateDrawing(selectedDrawing.id, { lineStyle: style })}
                    className={`flex-1 px-3 py-2 text-xs rounded ${
                      selectedDrawing.lineStyle === style
                        ? 'bg-emerald-600 text-white'
                        : 'bg-zinc-700 hover:bg-zinc-600'
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
                Line Width: {selectedDrawing.lineWidth}px
              </label>
              <input
                type="range"
                min="1"
                max="5"
                value={selectedDrawing.lineWidth}
                onChange={(e) =>
                  updateDrawing(selectedDrawing.id, { lineWidth: parseInt(e.target.value) })
                }
                className="w-full"
              />
            </div>

            {/* Extend Options */}
            <div className="grid grid-cols-2 gap-2">
              <label className="flex items-center gap-2 text-xs">
                <input
                  type="checkbox"
                  checked={selectedDrawing.extendLeft}
                  onChange={(e) =>
                    updateDrawing(selectedDrawing.id, { extendLeft: e.target.checked })
                  }
                  className="rounded"
                />
                Extend Left
              </label>
              <label className="flex items-center gap-2 text-xs">
                <input
                  type="checkbox"
                  checked={selectedDrawing.extendRight}
                  onChange={(e) =>
                    updateDrawing(selectedDrawing.id, { extendRight: e.target.checked })
                  }
                  className="rounded"
                />
                Extend Right
              </label>
            </div>

            {/* Price Labels */}
            <label className="flex items-center gap-2 text-xs">
              <input
                type="checkbox"
                checked={selectedDrawing.showPriceLabels}
                onChange={(e) =>
                  updateDrawing(selectedDrawing.id, { showPriceLabels: e.target.checked })
                }
                className="rounded"
              />
              Show Price Labels
            </label>

            {/* Text input for text labels */}
            {selectedDrawing.type === 'text' && (
              <div>
                <label className="text-xs font-medium text-zinc-400 mb-2 block">Text</label>
                <input
                  type="text"
                  value={selectedDrawing.text || ''}
                  onChange={(e) => updateDrawing(selectedDrawing.id, { text: e.target.value })}
                  className="w-full px-3 py-2 bg-zinc-700 border border-zinc-600 rounded text-sm"
                  placeholder="Enter text..."
                />
              </div>
            )}
          </div>
        </div>
      )}

      {/* Templates Panel */}
      {showTemplates && (
        <div className="absolute top-16 right-4 w-80 bg-zinc-800 border border-zinc-700 rounded-lg shadow-2xl z-50">
          <div className="flex items-center justify-between px-4 py-2 bg-zinc-900 border-b border-zinc-700 rounded-t-lg">
            <h3 className="text-sm font-semibold">Drawing Templates</h3>
            <button
              onClick={() => setShowTemplates(false)}
              className="p-1 hover:bg-zinc-700 rounded"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
          <div className="p-4 space-y-4">
            {/* Save Current as Template */}
            <div>
              <label className="text-xs font-medium text-zinc-400 mb-2 block">
                Save Current Drawings as Template
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={templateName}
                  onChange={(e) => setTemplateName(e.target.value)}
                  placeholder="Template name..."
                  className="flex-1 px-3 py-2 bg-zinc-700 border border-zinc-600 rounded text-sm"
                />
                <button
                  onClick={() => {
                    if (templateName.trim()) {
                      saveTemplate(templateName.trim());
                      setTemplateName('');
                    }
                  }}
                  disabled={!templateName.trim() || drawings.length === 0}
                  className="px-3 py-2 bg-emerald-600 hover:bg-emerald-700 disabled:bg-zinc-700 disabled:text-zinc-500 rounded text-sm flex items-center gap-1"
                >
                  <Save className="w-3 h-3" />
                  Save
                </button>
              </div>
            </div>

            {/* Template List */}
            <div>
              <label className="text-xs font-medium text-zinc-400 mb-2 block">
                Saved Templates ({templates.length})
              </label>
              <div className="space-y-2 max-h-64 overflow-y-auto">
                {templates.length === 0 ? (
                  <div className="text-xs text-zinc-500 text-center py-4">
                    No templates saved
                  </div>
                ) : (
                  templates.map((template) => (
                    <div
                      key={template.id}
                      className="flex items-center gap-2 p-2 bg-zinc-700 rounded text-xs"
                    >
                      <div className="flex-1">
                        <div className="font-medium">{template.name}</div>
                        <div className="text-zinc-400">
                          {template.drawings.length} drawing{template.drawings.length !== 1 ? 's' : ''}
                        </div>
                      </div>
                      <button
                        onClick={() => loadTemplate(template.id)}
                        className="px-2 py-1 bg-emerald-600 hover:bg-emerald-700 rounded flex items-center gap-1"
                        title="Load"
                      >
                        <Upload className="w-3 h-3" />
                        Load
                      </button>
                      <button
                        onClick={() => deleteTemplate(template.id)}
                        className="p-1 hover:bg-red-600/30 rounded text-red-400"
                        title="Delete"
                      >
                        <Trash2 className="w-3 h-3" />
                      </button>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Context Menu */}
      {contextMenu && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setContextMenu(null)}
          />
          <div
            className="fixed bg-zinc-800 border border-zinc-700 rounded-lg shadow-2xl py-1 z-50 min-w-[160px]"
            style={{ left: contextMenu.x, top: contextMenu.y }}
          >
            <button
              onClick={() => handleContextMenuAction('edit')}
              className="w-full px-4 py-2 text-left text-sm hover:bg-zinc-700 flex items-center gap-2"
            >
              <Settings className="w-3 h-3" />
              Edit Properties
            </button>
            <button
              onClick={() => handleContextMenuAction('duplicate')}
              className="w-full px-4 py-2 text-left text-sm hover:bg-zinc-700 flex items-center gap-2"
            >
              <Copy className="w-3 h-3" />
              Duplicate
            </button>
            <div className="h-px bg-zinc-700 my-1" />
            <button
              onClick={() => handleContextMenuAction('front')}
              className="w-full px-4 py-2 text-left text-sm hover:bg-zinc-700 flex items-center gap-2"
            >
              <Layers className="w-3 h-3" />
              Bring to Front
            </button>
            <button
              onClick={() => handleContextMenuAction('back')}
              className="w-full px-4 py-2 text-left text-sm hover:bg-zinc-700 flex items-center gap-2"
            >
              <Layers className="w-3 h-3" />
              Send to Back
            </button>
            <div className="h-px bg-zinc-700 my-1" />
            <button
              onClick={() => handleContextMenuAction('delete')}
              className="w-full px-4 py-2 text-left text-sm hover:bg-zinc-700 text-red-400 flex items-center gap-2"
            >
              <Trash2 className="w-3 h-3" />
              Delete
            </button>
          </div>
        </>
      )}
    </div>
  );
}
