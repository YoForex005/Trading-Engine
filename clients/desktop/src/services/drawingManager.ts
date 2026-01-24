/**
 * Drawing Manager
 * Manages chart drawings (trendlines, horizontal/vertical lines, text annotations)
 */

import type { IChartApi, ISeriesApi, Time } from 'lightweight-charts';

export type DrawingType = 'trendline' | 'hline' | 'vline' | 'text';

export interface DrawingPoint {
  time: number;
  price: number;
}

export interface Drawing {
  id: string;
  type: DrawingType;
  points: DrawingPoint[];
  text?: string;
  color?: string;
  lineWidth?: number;
}

export class DrawingManager {
  private drawings: Drawing[] = [];
  private activeDrawing: Drawing | null = null;
  private drawingIdCounter: number = 0;
  private chart: IChartApi | null = null;
  private series: ISeriesApi<any> | null = null;
  private overlayElements: Map<string, HTMLElement> = new Map();

  constructor() {}

  /**
   * Set the chart and series references
   */
  setChart(chart: IChartApi | null, series: ISeriesApi<any> | null): void {
    this.chart = chart;
    this.series = series;

    // Re-render all drawings when chart is set
    if (this.chart) {
      this.renderAllDrawings();
    }
  }

  /**
   * Start a new drawing
   */
  startDrawing(type: DrawingType, color: string = '#3b82f6'): string {
    const id = `drawing-${this.drawingIdCounter++}`;

    this.activeDrawing = {
      id,
      type,
      points: [],
      color,
      lineWidth: 2
    };

    return id;
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
        return drawing.points.length >= 1;
      case 'trendline':
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

    const completedDrawing = { ...this.activeDrawing };
    this.drawings.push(completedDrawing);
    this.activeDrawing = null;

    // Render the new drawing
    this.renderDrawing(completedDrawing);

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
   * Delete a drawing by ID
   */
  deleteDrawing(id: string): boolean {
    const index = this.drawings.findIndex(d => d.id === id);

    if (index === -1) return false;

    this.drawings.splice(index, 1);

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

    return true;
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

    // For now, we'll use a simple approach with HTML overlays
    // In production, you might want to use chart primitives or SVG
    const container = document.querySelector('.chart-drawing-overlay');
    if (!container) return;

    const element = this.createDrawingElement(drawing);
    if (element) {
      container.appendChild(element);
      this.overlayElements.set(drawing.id, element);
    }
  }

  /**
   * Create HTML element for drawing
   */
  private createDrawingElement(drawing: Drawing): HTMLElement | null {
    if (!this.series) return null;

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
            div.style.cursor = 'ns-resize';
          }
        }
        break;

      case 'vline':
        // Vertical lines require time-to-coordinate conversion
        // This is more complex and requires timeScale API
        break;

      case 'trendline':
        if (drawing.points.length >= 2) {
          const y1 = this.series.priceToCoordinate(drawing.points[0].price);
          const y2 = this.series.priceToCoordinate(drawing.points[1].price);

          if (y1 !== null && y2 !== null) {
            // Calculate line position and angle
            const length = Math.sqrt(
              Math.pow(100, 2) + Math.pow(y2 - y1, 2)
            );
            const angle = Math.atan2(y2 - y1, 100) * (180 / Math.PI);

            div.style.width = `${length}px`;
            div.style.height = `${drawing.lineWidth || 2}px`;
            div.style.backgroundColor = drawing.color || '#3b82f6';
            div.style.transformOrigin = 'left center';
            div.style.transform = `rotate(${angle}deg)`;
            div.style.left = '0';
            div.style.top = `${y1}px`;
          }
        }
        break;

      case 'text':
        if (drawing.points.length > 0 && drawing.text) {
          const y = this.series.priceToCoordinate(drawing.points[0].price);
          if (y !== null) {
            div.style.left = '10px';
            div.style.top = `${y}px`;
            div.style.color = drawing.color || '#3b82f6';
            div.style.fontSize = '12px';
            div.style.fontWeight = 'bold';
            div.style.padding = '4px 8px';
            div.style.backgroundColor = 'rgba(0, 0, 0, 0.7)';
            div.style.borderRadius = '4px';
            div.textContent = drawing.text;
          }
        }
        break;
    }

    return div;
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
    this.renderAllDrawings();
  }

  /**
   * Save drawings to localStorage
   */
  saveToStorage(symbol: string): void {
    try {
      const key = `chart-drawings-${symbol}`;
      localStorage.setItem(key, JSON.stringify(this.drawings));
    } catch (error) {
      console.error('Failed to save drawings:', error);
    }
  }

  /**
   * Load drawings from localStorage
   */
  loadFromStorage(symbol: string): void {
    try {
      const key = `chart-drawings-${symbol}`;
      const data = localStorage.getItem(key);

      if (data) {
        this.drawings = JSON.parse(data);
        this.renderAllDrawings();
      }
    } catch (error) {
      console.error('Failed to load drawings:', error);
    }
  }
}

// Export singleton instance
export const drawingManager = new DrawingManager();
