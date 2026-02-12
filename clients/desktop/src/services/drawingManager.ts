/**
 * Drawing Manager
 * Manages chart drawings (trendlines, horizontal/vertical lines, text annotations)
 */

import type { IChartApi, ISeriesApi } from 'lightweight-charts';
import { API_ENDPOINTS } from '../config/api';

export type DrawingType = 'trendline' | 'hline' | 'vline' | 'text' | 'channel' | 'fibonacci' | 'shapes';

export interface DrawingPoint {
  time: number;
  price: number;
}

export interface Drawing {
  id: string;
  type: DrawingType;
  subtype?: string;
  points: DrawingPoint[];
  text?: string;
  color?: string;
  lineWidth?: number;
  lineStyle?: 'solid' | 'dashed' | 'dotted';
  selected?: boolean;
  locked?: boolean;
}

export interface DrawingHistory {
  action: 'delete' | 'add' | 'modify';
  drawing: Drawing;
}

export class DrawingManager {
  private drawings: Drawing[] = [];
  private activeDrawing: Drawing | null = null;
  private chart: IChartApi | null = null;
  private series: ISeriesApi<any> | null = null;
  private overlayElements: Map<string, HTMLElement> = new Map();
  private currentSymbol: string | null = null;
  private currentAccountId: number = 1;
  private undoBuffer: DrawingHistory[] = [];

  // Dragging state
  private draggingDrawingId: string | null = null;
  private draggingNodeIndex: number | null = null;
  private dragStartPos: { x: number; y: number } | null = null;
  private dragInitialPoints: DrawingPoint[] = [];

  constructor() {
    this.setupGlobalListeners();
  }

  private setupGlobalListeners(): void {
    window.addEventListener('mousemove', this.handleMouseMove.bind(this));
    window.addEventListener('mouseup', this.handleMouseUp.bind(this));
  }

  /**
   * Set the chart and series references
   */
  setChart(chart: IChartApi | null, series: ISeriesApi<any> | null, symbol: string | null = null, accountId: number = 1): void {
    this.chart = chart;
    this.series = series;
    this.currentAccountId = accountId;

    if (symbol && symbol !== this.currentSymbol) {
      this.currentSymbol = symbol;
      this.loadFromBackend(symbol);
    }

    // Re-render all drawings when chart is set
    if (this.chart) {
      this.renderAllDrawings();
    }
  }

  /**
   * Start a new drawing
   */
  startDrawing(type: DrawingType, color: string = '#3b82f6', subtype?: string): string {
    const id = `drawing-${Date.now()}`;

    this.activeDrawing = {
      id,
      type,
      subtype,
      points: [],
      color,
      lineWidth: 2
    };

    return id;
  }

  /**
   * Update the active drawing properties
   */
  updateActiveDrawing(props: Partial<Drawing>): void {
    if (this.activeDrawing) {
      this.activeDrawing = { ...this.activeDrawing, ...props };
    }
  }

  /**
   * Add a point to the active drawing
   */
  addPoint(time: number, price: number): boolean {
    if (!this.activeDrawing) return false;

    this.activeDrawing.points.push({ time, price });

    // Check if drawing is complete
    const isComplete = this.isDrawingComplete(this.activeDrawing);

    if (isComplete) {
      this.finishDrawing();
    }

    return isComplete;
  }

  /**
   * Check if a drawing is complete based on its type
   */
  private isDrawingComplete(drawing: Drawing): boolean {
    switch (drawing.type) {
      case 'hline':
      case 'vline':
      case 'shapes':
        return drawing.points.length >= 1;
      case 'trendline':
      case 'channel':
      case 'fibonacci':
        return drawing.points.length >= 2;
      case 'text':
        return drawing.points.length >= 1;
      default:
        return false;
    }
  }

  /**
   * Finish the active drawing
   */
  finishDrawing(): Drawing | null {
    if (!this.activeDrawing) return null;

    const completedDrawing = { ...this.activeDrawing, selected: true };
    this.drawings.push(completedDrawing);
    this.activeDrawing = null;

    // Render the new drawing
    this.renderDrawing(completedDrawing);

    // Save to backend
    this.saveToBackend(completedDrawing);

    // Dispatch event for saving
    window.dispatchEvent(new CustomEvent('drawing:saved', {
      detail: completedDrawing
    }));

    return completedDrawing;
  }

  /**
   * Cancel the active drawing
   */
  cancelDrawing(): void {
    this.activeDrawing = null;
  }

  /**
   * Selection Management
   */
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
    }
    this.renderAllDrawings();
  }

  unselectAll(): void {
    // Defensive: Ensure drawings is an array
    if (!Array.isArray(this.drawings)) {
      this.drawings = [];
      return;
    }
    this.drawings.forEach(d => d.selected = false);
    this.renderAllDrawings();
  }

  unselectDrawing(id: string): void {
    const drawing = this.drawings.find(d => d.id === id);
    if (drawing) {
      drawing.selected = false;
    }
    this.renderAllDrawings();
  }

  /**
   * Delete Operations
   */
  deleteDrawing(id: string): boolean {
    const index = this.drawings.findIndex(d => d.id === id);

    if (index === -1) return false;

    const removed = this.drawings.splice(index, 1)[0];
    this.undoBuffer.push({ action: 'delete', drawing: removed });

    // Remove backend record
    if (this.currentSymbol) {
      this.deleteFromBackend(id, this.currentSymbol);
    }

    // Remove overlay element
    const element = this.overlayElements.get(id);
    if (element && element.parentNode) {
      element.parentNode.removeChild(element);
    }
    this.overlayElements.delete(id);

    // Dispatch event
    window.dispatchEvent(new CustomEvent('drawing:deleted', {
      detail: { id }
    }));

    this.renderAllDrawings(); // Re-render to clear nodes
    return true;
  }

  /**
   * Drag & Drop Logic
   */
  private handleMouseDown(e: MouseEvent, id: string, nodeIndex: number | null = null): void {
    e.stopPropagation();
    const drawing = this.drawings.find(d => d.id === id);
    if (!drawing || drawing.locked) return;

    this.selectDrawing(id, e.ctrlKey || e.metaKey);

    this.draggingDrawingId = id;
    this.draggingNodeIndex = nodeIndex;
    this.dragStartPos = { x: e.clientX, y: e.clientY };
    this.dragInitialPoints = JSON.parse(JSON.stringify(drawing.points));
  }

  private handleMouseMove(e: MouseEvent): void {
    if (!this.draggingDrawingId || !this.dragStartPos || !this.chart || !this.series) return;

    const drawing = this.drawings.find(d => d.id === this.draggingDrawingId);
    if (!drawing) return;

    const dx = e.clientX - this.dragStartPos.x;
    const dy = e.clientY - this.dragStartPos.y;

    if (this.draggingNodeIndex !== null) {
      // Move single node
      const point = drawing.points[this.draggingNodeIndex];
      const initialPoint = this.dragInitialPoints[this.draggingNodeIndex];

      const x = this.chart.timeScale().timeToCoordinate(initialPoint.time as any);
      const y = this.series.priceToCoordinate(initialPoint.price);

      if (x !== null && y !== null) {
        const newX = x + dx;
        const newY = y + dy;

        const newTime = this.chart.timeScale().coordinateToTime(newX);
        const newPrice = this.series.coordinateToPrice(newY);

        if (newTime !== null && newPrice !== null) {
          point.time = newTime as number;
          point.price = newPrice;
        }
      }
    } else {
      // Move entire drawing
      drawing.points = this.dragInitialPoints.map(p => {
        const x = this.chart!.timeScale().timeToCoordinate(p.time as any);
        const y = this.series!.priceToCoordinate(p.price);

        if (x !== null && y !== null) {
          const newX = x + dx;
          const newY = y + dy;
          const newTime = this.chart!.timeScale().coordinateToTime(newX);
          const newPrice = this.series!.coordinateToPrice(newY);

          if (newTime !== null && newPrice !== null) {
            return { time: newTime as number, price: newPrice };
          }
        }
        return p;
      });
    }

    this.renderAllDrawings();
  }

  private handleMouseUp(): void {
    if (this.draggingDrawingId) {
      const drawing = this.drawings.find(d => d.id === this.draggingDrawingId);
      if (drawing) {
        this.saveToBackend(drawing);
      }
    }
    this.draggingDrawingId = null;
    this.draggingNodeIndex = null;
    this.dragStartPos = null;
    this.dragInitialPoints = [];
  }

  deleteSelected(): void {
    const toDelete = this.drawings.filter(d => d.selected);
    toDelete.forEach(d => this.deleteDrawing(d.id));
  }

  deleteByType(type: DrawingType): void {
    const toDelete = this.drawings.filter(d => d.type === type);
    toDelete.forEach(d => this.deleteDrawing(d.id));
  }

  undoDelete(): void {
    const last = this.undoBuffer.pop();
    if (last && last.action === 'delete') {
      this.drawings.push(last.drawing);
      this.saveToBackend(last.drawing);
      this.renderAllDrawings();
    }
  }

  /**
   * Clear all drawings
   */
  clearAllDrawings(): void {
    this.drawings = [];
    this.activeDrawing = null;

    // Remove all overlay elements
    this.overlayElements.forEach(element => {
      if (element.parentNode) {
        element.parentNode.removeChild(element);
      }
    });
    this.overlayElements.clear();

    // Dispatch event
    window.dispatchEvent(new CustomEvent('drawings:cleared'));
  }

  /**
   * Get all drawings
   */
  getDrawings(): Drawing[] {
    return [...this.drawings];
  }

  /**
   * Get active drawing
   */
  getActiveDrawing(): Drawing | null {
    return this.activeDrawing ? { ...this.activeDrawing } : null;
  }

  /**
   * Render a single drawing
   */
  private renderDrawing(drawing: Drawing): void {
    if (!this.chart || !this.series) return;

    const container = document.querySelector('.chart-drawing-overlay');
    if (!container) return;

    const element = this.createDrawingElement(drawing);
    if (element) {
      container.appendChild(element);
      this.overlayElements.set(drawing.id, element);

      // Add nodes for path-based drawings (Red Squares if selected)
      // Ensure points array exists before iterating
      if (Array.isArray(drawing.points)) {
        drawing.points.forEach((point, index) => {
        const node = document.createElement('div');
        node.className = `drawing-node drawing-node-${drawing.id} ${drawing.selected ? 'drawing-node-selected' : ''}`;

        // Selection state from MT5 image: red squares on endpoints
        if (drawing.selected) {
          node.style.display = 'block';
        } else {
          node.style.display = 'none';
        }

        const x = this.chart!.timeScale().timeToCoordinate(point.time as any);
        const y = this.series!.priceToCoordinate(point.price);

        if (x !== null && y !== null) {
          node.style.left = `${x}px`;
          node.style.top = `${y}px`;

          if (drawing.selected) {
            node.style.backgroundColor = '#ff0000'; // MT5 Red Square
            node.style.border = '1px solid #ffffff';
            node.style.width = '7px';
            node.style.height = '7px';
            node.style.borderRadius = '0'; // Square
          }

          node.addEventListener('mousedown', (e) => this.handleMouseDown(e, drawing.id, index));
          container.appendChild(node);
        }
      });
      } // End of Array.isArray check

      // Add mouse down to the element itself for dragging entire drawing
      element.addEventListener('mousedown', (e) => this.handleMouseDown(e, drawing.id));

      // Context menu listener
      element.addEventListener('contextmenu', (e) => {
        e.preventDefault();
        window.dispatchEvent(new CustomEvent('drawing:contextmenu', {
          detail: { x: e.clientX, y: e.clientY, drawingId: drawing.id }
        }));
      });
    }
  }

  /**
   * Create HTML element for drawing
   */
  private createDrawingElement(drawing: Drawing): HTMLElement | null {
    // Defensive: Ensure both chart and series exist
    if (!this.chart || !this.series) return null;

    const div = document.createElement('div');
    div.className = 'chart-drawing';
    div.setAttribute('data-drawing-id', drawing.id);
    div.style.position = 'absolute';
    div.style.pointerEvents = 'auto';

    switch (drawing.type) {
      case 'hline':
        if (drawing.points.length > 0) {
          const y = this.series.priceToCoordinate(drawing.points[0].price);
          if (y !== null) {
            div.style.left = '0';
            div.style.right = '0';
            div.style.top = `${y}px`;
            div.style.height = `${drawing.lineWidth || 2}px`;
            div.style.backgroundColor = drawing.color || '#3b82f6';
            if (drawing.lineStyle === 'dashed') {
              div.style.backgroundColor = 'transparent';
              div.style.backgroundImage = `linear-gradient(to right, ${drawing.color || '#3b82f6'} 50%, transparent 50%)`;
              div.style.backgroundSize = '10px 100%';
            } else if (drawing.lineStyle === 'dotted') {
              div.style.backgroundColor = 'transparent';
              div.style.backgroundImage = `radial-gradient(${drawing.color || '#3b82f6'} 1px, transparent 1px)`;
              div.style.backgroundSize = '4px 100%';
            }
            div.style.cursor = drawing.locked ? 'default' : 'ns-resize';
          }
        }
        break;

      case 'vline':
        if (drawing.points.length > 0 && this.chart) {
          const x = this.chart.timeScale().timeToCoordinate(drawing.points[0].time as any);
          if (x !== null) {
            div.style.left = `${x}px`;
            div.style.top = '0';
            div.style.bottom = '0';
            div.style.width = `${drawing.lineWidth || 2}px`;
            div.style.backgroundColor = drawing.color || '#3b82f6';
            div.style.cursor = 'ew-resize';
          }
        }
        break;

      case 'trendline':
      case 'channel':
      case 'fibonacci':
        if (drawing.points.length >= 2 && this.chart) {
          const x1 = this.chart.timeScale().timeToCoordinate(drawing.points[0].time as any);
          const x2 = this.chart.timeScale().timeToCoordinate(drawing.points[1].time as any);
          const y1 = this.series.priceToCoordinate(drawing.points[0].price);
          const y2 = this.series.priceToCoordinate(drawing.points[1].price);

          if (x1 !== null && x2 !== null && y1 !== null && y2 !== null) {
            if (drawing.type === 'trendline') {
              const length = Math.sqrt(Math.pow(x2 - x1, 2) + Math.pow(y2 - y1, 2));
              const angle = Math.atan2(y2 - y1, x2 - x1) * (180 / Math.PI);

              div.style.width = `${length}px`;
              div.style.height = `${drawing.lineWidth || 2}px`;
              div.style.backgroundColor = drawing.color || '#3b82f6';
              div.style.transformOrigin = 'left center';
              div.style.transform = `rotate(${angle}deg)`;
              div.style.left = `${x1}px`;
              div.style.top = `${y1}px`;
            } else if (drawing.type === 'fibonacci') {
              div.style.left = `${Math.min(x1, x2)}px`;
              div.style.top = `${Math.min(y1, y2)}px`;
              div.style.width = `${Math.abs(x2 - x1)}px`;
              div.style.height = `${Math.abs(y2 - y1)}px`;
              div.style.border = `1px dashed ${drawing.color || '#3b82f6'}`;
              div.style.backgroundColor = 'rgba(59, 130, 246, 0.1)';
            } else {
              div.style.left = `${Math.min(x1, x2)}px`;
              div.style.top = `${Math.min(y1, y2)}px`;
              div.style.width = `${Math.abs(x2 - x1)}px`;
              div.style.height = `${Math.abs(y2 - y1)}px`;
              div.style.border = `${drawing.lineWidth || 2}px solid ${drawing.color || '#3b82f6'}`;
            }
          }
        }
        break;

      case 'text':
        if (drawing.points.length > 0 && drawing.text && this.chart) {
          const x = this.chart.timeScale().timeToCoordinate(drawing.points[0].time as any);
          const y = this.series.priceToCoordinate(drawing.points[0].price);
          if (x !== null && y !== null) {
            div.style.left = `${x}px`;
            div.style.top = `${y}px`;
            div.style.color = drawing.color || '#3b82f6';
            div.style.fontSize = '12px';
            div.style.fontWeight = 'bold';
            div.style.padding = '4px 8px';
            div.style.backgroundColor = 'rgba(24, 24, 27, 0.9)';
            div.style.borderRadius = '4px';
            div.style.border = '1px solid #3f3f46';
            div.style.whiteSpace = 'nowrap';
            div.textContent = drawing.text;
          }
        }
        break;

      case 'shapes':
        if (drawing.points.length > 0 && this.chart) {
          const x = this.chart.timeScale().timeToCoordinate(drawing.points[0].time as any);
          const y = this.series.priceToCoordinate(drawing.points[0].price);
          if (x !== null && y !== null) {
            div.style.left = `${x}px`;
            div.style.top = `${y}px`;
            div.style.color = drawing.color || '#3b82f6';
            div.style.transform = 'translate(-50%, -50%)';
            div.innerHTML = `<div style="font-size: 20px;">${this.getSymbolForSubtype(drawing.subtype)}</div>`;
          }
        }
        break;
    }

    return div;
  }

  private getSymbolForSubtype(subtype?: string): string {
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
  }

  /**
   * Render all drawings
   */
  renderAllDrawings(): void {
    // Clear existing overlays
    this.overlayElements.forEach(element => {
      if (element.parentNode) {
        element.parentNode.removeChild(element);
      }
    });

    // Also clear all drawing nodes
    const container = document.querySelector('.chart-drawing-overlay');
    if (container) {
      container.querySelectorAll('.drawing-node').forEach(node => {
        if (node.parentNode) node.parentNode.removeChild(node);
      });
    }

    this.overlayElements.clear();

    // Render each drawing
    this.drawings.forEach(drawing => {
      this.renderDrawing(drawing);
    });
  }

  /**
   * Update drawings on chart scroll/zoom
   */
  updateDrawingPositions(): void {
    // Defensive: Ensure chart and series exist before rendering
    if (!this.chart || !this.series) return;
    this.renderAllDrawings();
  }

  /**
   * Sync with backend
   */
  async saveToBackend(drawing: Drawing): Promise<void> {
    if (!this.currentSymbol) return;

    try {
      const response = await fetch(API_ENDPOINTS.workspace.drawings, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...drawing,
          symbol: this.currentSymbol,
          accountId: this.currentAccountId
        }),
      });

      if (response.ok) {
        const savedDrawing = await response.json();
        const index = this.drawings.findIndex(d => d.id === drawing.id);
        if (index !== -1) {
          this.drawings[index].id = savedDrawing.id;
          this.renderAllDrawings();
        }
      }
    } catch (error) {
      console.error('Failed to save to backend:', error);
      this.saveToStorage(this.currentSymbol);
    }
  }

  async loadFromBackend(symbol: string): Promise<void> {
    try {
      const response = await fetch(`${API_ENDPOINTS.workspace.drawings}?symbol=${symbol}&accountId=${this.currentAccountId}`);
      if (response.ok) {
        const drawings = await response.json();
        // Ensure drawings is always an array, never null or undefined
        this.drawings = Array.isArray(drawings) ? drawings : [];
        this.renderAllDrawings();
      } else {
        this.loadFromStorage(symbol);
      }
    } catch (error) {
      console.error('Failed to load from backend:', error);
      this.loadFromStorage(symbol);
    }
  }

  async deleteFromBackend(id: string, symbol: string): Promise<void> {
    try {
      await fetch(`${API_ENDPOINTS.workspace.drawings}/${id}?symbol=${symbol}&accountId=${this.currentAccountId}`, {
        method: 'DELETE',
      });
    } catch (error) {
      console.error('Failed to delete from backend:', error);
    }
  }

  /**
   * Fallback Storage
   */
  saveToStorage(symbol: string): void {
    try {
      localStorage.setItem(`drawings-${symbol}`, JSON.stringify(this.drawings));
    } catch (e) {
      console.error('Failed to save to local storage', e);
    }
  }

  loadFromStorage(symbol: string): void {
    try {
      const data = localStorage.getItem(`drawings-${symbol}`);
      if (data) {
        const parsed = JSON.parse(data);
        // Ensure drawings is always an array, never null or undefined
        this.drawings = Array.isArray(parsed) ? parsed : [];
        this.renderAllDrawings();
      }
    } catch (e) {
      console.error('Failed to load from local storage', e);
      // Reset to empty array on error
      this.drawings = [];
    }
  }
}

// Export singleton instance
export const drawingManager = new DrawingManager();
