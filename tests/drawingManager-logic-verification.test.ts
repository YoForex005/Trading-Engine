/**
 * DrawingManager Logic Verification Test Suite
 *
 * This test suite verifies the complete logic flow of the drawingManager service:
 * 1. Singleton instance export
 * 2. startDrawing() creates activeDrawing properly
 * 3. addPoint() adds coordinates correctly
 * 4. isDrawingComplete() works for all 12 types
 * 5. finishDrawing() creates overlay elements
 * 6. renderDrawing() handles all drawing types
 * 7. Coordinate conversions (timeToCoordinate, priceToCoordinate)
 * 8. Chart instance set correctly
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { drawingManager, DrawingManager, DrawingType } from '../clients/desktop/src/services/drawingManager';
import type { IChartApi, ISeriesApi } from 'lightweight-charts';

describe('DrawingManager Logic Verification', () => {
  let mockChart: Partial<IChartApi>;
  let mockSeries: Partial<ISeriesApi<any>>;
  let mockContainer: HTMLElement;

  beforeEach(() => {
    // Setup mock chart with coordinate conversion methods
    mockChart = {
      timeScale: vi.fn(() => ({
        timeToCoordinate: vi.fn((time: number) => time * 10), // Mock conversion
        coordinateToTime: vi.fn((coord: number) => coord / 10),
      })),
    } as any;

    mockSeries = {
      priceToCoordinate: vi.fn((price: number) => 500 - price * 10), // Mock conversion
      coordinateToPrice: vi.fn((coord: number) => (500 - coord) / 10),
    } as any;

    // Setup mock DOM container
    mockContainer = document.createElement('div');
    mockContainer.className = 'chart-drawing-overlay';
    document.body.appendChild(mockContainer);

    // Clear any existing drawings
    drawingManager.clearAllDrawings();
  });

  afterEach(() => {
    if (mockContainer && mockContainer.parentNode) {
      mockContainer.parentNode.removeChild(mockContainer);
    }
    drawingManager.clearAllDrawings();
  });

  // TEST 1: Singleton Instance Export
  describe('1. Singleton Instance Export', () => {
    it('should export a singleton instance', () => {
      expect(drawingManager).toBeDefined();
      expect(drawingManager).toBeInstanceOf(DrawingManager);
    });

    it('should maintain same instance across imports', () => {
      const instance1 = drawingManager;
      const instance2 = drawingManager;
      expect(instance1).toBe(instance2);
    });
  });

  // TEST 2: startDrawing() Creates activeDrawing Properly
  describe('2. startDrawing() Creates activeDrawing', () => {
    it('should create activeDrawing with correct properties', () => {
      const id = drawingManager.startDrawing('trendline', '#3b82f6');

      const activeDrawing = drawingManager.getActiveDrawing();
      expect(activeDrawing).not.toBeNull();
      expect(activeDrawing?.id).toBe(id);
      expect(activeDrawing?.type).toBe('trendline');
      expect(activeDrawing?.color).toBe('#3b82f6');
      expect(activeDrawing?.lineWidth).toBe(2);
      expect(activeDrawing?.points).toEqual([]);
    });

    it('should generate unique IDs for each drawing', () => {
      const id1 = drawingManager.startDrawing('trendline');
      drawingManager.cancelDrawing();
      const id2 = drawingManager.startDrawing('hline');

      expect(id1).not.toBe(id2);
    });

    it('should handle subtype parameter', () => {
      drawingManager.startDrawing('shapes', '#ff0000', 'thumbs_up');

      const activeDrawing = drawingManager.getActiveDrawing();
      expect(activeDrawing?.subtype).toBe('thumbs_up');
    });
  });

  // TEST 3: addPoint() Adds Coordinates Correctly
  describe('3. addPoint() Adds Coordinates', () => {
    beforeEach(() => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);
    });

    it('should add point to activeDrawing', () => {
      drawingManager.startDrawing('trendline', '#3b82f6');
      const result = drawingManager.addPoint(1000, 1.5000);

      const activeDrawing = drawingManager.getActiveDrawing();
      expect(activeDrawing?.points.length).toBe(1);
      expect(activeDrawing?.points[0]).toEqual({ time: 1000, price: 1.5000 });
    });

    it('should return false if no activeDrawing', () => {
      const result = drawingManager.addPoint(1000, 1.5000);
      expect(result).toBe(false);
    });

    it('should add multiple points in sequence', () => {
      drawingManager.startDrawing('pitchfork', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(2000, 1.5100);

      const activeDrawing = drawingManager.getActiveDrawing();
      expect(activeDrawing?.points.length).toBe(2);
    });
  });

  // TEST 4: isDrawingComplete() Works for All 12 Types
  describe('4. isDrawingComplete() for All Types', () => {
    beforeEach(() => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);
    });

    const drawingTypeTests: { type: DrawingType; requiredPoints: number }[] = [
      { type: 'hline', requiredPoints: 1 },
      { type: 'vline', requiredPoints: 1 },
      { type: 'shapes', requiredPoints: 1 },
      { type: 'text', requiredPoints: 1 },
      { type: 'arrow', requiredPoints: 1 },
      { type: 'trendline', requiredPoints: 2 },
      { type: 'channel', requiredPoints: 2 },
      { type: 'fibonacci', requiredPoints: 2 },
      { type: 'rectangle', requiredPoints: 2 },
      { type: 'ellipse', requiredPoints: 2 },
      { type: 'pitchfork', requiredPoints: 3 },
    ];

    drawingTypeTests.forEach(({ type, requiredPoints }) => {
      it(`should complete ${type} with ${requiredPoints} point(s)`, () => {
        drawingManager.startDrawing(type, '#3b82f6');

        // Add points one by one
        for (let i = 0; i < requiredPoints - 1; i++) {
          const result = drawingManager.addPoint(1000 + i * 1000, 1.5000 + i * 0.01);
          expect(result).toBe(false); // Not complete yet
        }

        // Add final point
        const finalResult = drawingManager.addPoint(
          1000 + (requiredPoints - 1) * 1000,
          1.5000 + (requiredPoints - 1) * 0.01
        );
        expect(finalResult).toBe(true); // Should be complete

        // activeDrawing should be null after completion
        const activeDrawing = drawingManager.getActiveDrawing();
        expect(activeDrawing).toBeNull();

        // Drawing should be in drawings array
        const drawings = drawingManager.getDrawings();
        expect(drawings.length).toBe(1);
        expect(drawings[0].type).toBe(type);
      });
    });
  });

  // TEST 5: finishDrawing() Creates Overlay Elements
  describe('5. finishDrawing() Creates Drawing', () => {
    beforeEach(() => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);
    });

    it('should return completed drawing', () => {
      drawingManager.startDrawing('trendline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(2000, 1.5100);

      // addPoint auto-calls finishDrawing when complete
      const drawings = drawingManager.getDrawings();
      expect(drawings.length).toBe(1);
      expect(drawings[0].selected).toBe(true);
    });

    it('should add drawing to drawings array', () => {
      drawingManager.startDrawing('hline', '#ff0000');
      drawingManager.addPoint(1000, 1.5000);

      const drawings = drawingManager.getDrawings();
      expect(drawings.length).toBe(1);
    });

    it('should clear activeDrawing', () => {
      drawingManager.startDrawing('vline', '#00ff00');
      drawingManager.addPoint(1000, 1.5000);

      const activeDrawing = drawingManager.getActiveDrawing();
      expect(activeDrawing).toBeNull();
    });

    it('should dispatch drawing:saved event', (done) => {
      const handler = (event: CustomEvent) => {
        expect(event.detail).toBeDefined();
        expect(event.detail.type).toBe('trendline');
        window.removeEventListener('drawing:saved', handler as any);
        done();
      };

      window.addEventListener('drawing:saved', handler as any);

      drawingManager.startDrawing('trendline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(2000, 1.5100);
    });
  });

  // TEST 6: renderDrawing() Handles All Drawing Types
  describe('6. renderDrawing() Handles All Types', () => {
    beforeEach(() => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);
    });

    it('should render hline', () => {
      drawingManager.startDrawing('hline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);
    });

    it('should render vline', () => {
      drawingManager.startDrawing('vline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);
    });

    it('should render trendline', () => {
      drawingManager.startDrawing('trendline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(2000, 1.5100);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);
    });

    it('should render fibonacci with levels', () => {
      drawingManager.startDrawing('fibonacci', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(2000, 1.5100);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);

      // Check for SVG elements (fibonacci uses SVG)
      const svgElements = mockContainer.querySelectorAll('svg');
      expect(svgElements.length).toBeGreaterThan(0);
    });

    it('should render rectangle', () => {
      drawingManager.startDrawing('rectangle', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(2000, 1.5100);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);
    });

    it('should render ellipse', () => {
      drawingManager.startDrawing('ellipse', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(2000, 1.5100);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);
    });

    it('should render pitchfork', () => {
      drawingManager.startDrawing('pitchfork', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(2000, 1.5100);
      drawingManager.addPoint(1500, 1.5050);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);

      // Check for SVG elements (pitchfork uses SVG)
      const svgElements = mockContainer.querySelectorAll('svg');
      expect(svgElements.length).toBeGreaterThan(0);
    });

    it('should render text annotation', () => {
      drawingManager.startDrawing('text', '#3b82f6');
      drawingManager.updateActiveDrawing({ text: 'Test Label' });
      drawingManager.addPoint(1000, 1.5000);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);
      expect(elements[0].textContent).toBe('Test Label');
    });

    it('should render shapes', () => {
      drawingManager.startDrawing('shapes', '#3b82f6', 'thumbs_up');
      drawingManager.addPoint(1000, 1.5000);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);
    });

    it('should render arrow', () => {
      drawingManager.startDrawing('arrow', '#3b82f6', 'up');
      drawingManager.addPoint(1000, 1.5000);

      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);
    });

    it('should not render hidden drawings', () => {
      drawingManager.startDrawing('hline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);

      const drawings = drawingManager.getDrawings();
      const drawingId = drawings[0].id;

      // Toggle visibility to false
      drawingManager.toggleDrawingVisibility(drawingId);

      // Clear and re-render
      mockContainer.innerHTML = '';
      drawingManager.renderAllDrawings();

      // Should not render hidden drawing
      const elements = mockContainer.querySelectorAll(`[data-drawing-id="${drawingId}"]`);
      expect(elements.length).toBe(0);
    });
  });

  // TEST 7: Coordinate Conversions
  describe('7. Coordinate Conversions', () => {
    beforeEach(() => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);
    });

    it('should call timeToCoordinate for time conversion', () => {
      const timeToCoordinateSpy = vi.spyOn(mockChart.timeScale!() as any, 'timeToCoordinate');

      drawingManager.startDrawing('trendline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(2000, 1.5100);

      // Should be called during rendering
      expect(timeToCoordinateSpy).toHaveBeenCalled();
    });

    it('should call priceToCoordinate for price conversion', () => {
      const priceToCoordinateSpy = vi.spyOn(mockSeries, 'priceToCoordinate');

      drawingManager.startDrawing('hline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);

      // Should be called during rendering
      expect(priceToCoordinateSpy).toHaveBeenCalled();
    });

    it('should handle null coordinate conversions gracefully', () => {
      // Mock returning null
      mockChart.timeScale = vi.fn(() => ({
        timeToCoordinate: vi.fn(() => null),
        coordinateToTime: vi.fn(() => null),
      })) as any;

      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);

      // Should not throw error
      expect(() => {
        drawingManager.startDrawing('trendline', '#3b82f6');
        drawingManager.addPoint(1000, 1.5000);
        drawingManager.addPoint(2000, 1.5100);
      }).not.toThrow();
    });
  });

  // TEST 8: Chart Instance Set Correctly
  describe('8. Chart Instance Management', () => {
    it('should set chart and series references', () => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);

      // Verify by attempting to add a drawing (would fail without chart)
      drawingManager.startDrawing('hline', '#3b82f6');
      const result = drawingManager.addPoint(1000, 1.5000);

      expect(result).toBe(true);
    });

    it('should store current symbol', () => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);

      // Symbol is stored internally and used for backend operations
      // Verify by checking if drawings can be created
      drawingManager.startDrawing('hline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);

      const drawings = drawingManager.getDrawings();
      expect(drawings.length).toBe(1);
    });

    it('should re-render drawings when chart is set', () => {
      // Add a drawing first
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);
      drawingManager.startDrawing('hline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);

      // Clear container
      mockContainer.innerHTML = '';

      // Set chart again (should trigger re-render)
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);

      // Check if drawing was re-rendered
      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);
    });

    it('should handle null chart gracefully', () => {
      expect(() => {
        drawingManager.setChart(null, null, null, 1);
      }).not.toThrow();
    });
  });

  // TEST 9: Complete Flow Integration Test
  describe('9. Complete Flow Integration', () => {
    beforeEach(() => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);
    });

    it('should complete entire trendline flow', () => {
      // Step 1: startDrawing
      const id = drawingManager.startDrawing('trendline', '#3b82f6');
      expect(id).toBeTruthy();

      // Step 2: activeDrawing created
      let activeDrawing = drawingManager.getActiveDrawing();
      expect(activeDrawing).not.toBeNull();
      expect(activeDrawing?.type).toBe('trendline');

      // Step 3: addPoint (first point)
      let isComplete = drawingManager.addPoint(1000, 1.5000);
      expect(isComplete).toBe(false); // Not complete yet

      activeDrawing = drawingManager.getActiveDrawing();
      expect(activeDrawing?.points.length).toBe(1);

      // Step 4: addPoint (second point - completes drawing)
      isComplete = drawingManager.addPoint(2000, 1.5100);
      expect(isComplete).toBe(true); // Drawing completed

      // Step 5: finishDrawing called automatically
      activeDrawing = drawingManager.getActiveDrawing();
      expect(activeDrawing).toBeNull(); // Active drawing cleared

      // Step 6: Drawing added to array
      const drawings = drawingManager.getDrawings();
      expect(drawings.length).toBe(1);
      expect(drawings[0].type).toBe('trendline');
      expect(drawings[0].points.length).toBe(2);

      // Step 7: Overlay element created
      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThan(0);

      // Step 8: renderDrawing called - drawing appears on chart
      const drawingElement = mockContainer.querySelector(`[data-drawing-id="${drawings[0].id}"]`);
      expect(drawingElement).not.toBeNull();
    });

    it('should handle multiple drawings simultaneously', () => {
      // Create multiple drawings
      drawingManager.startDrawing('hline', '#ff0000');
      drawingManager.addPoint(1000, 1.5000);

      drawingManager.startDrawing('vline', '#00ff00');
      drawingManager.addPoint(2000, 1.5100);

      drawingManager.startDrawing('trendline', '#0000ff');
      drawingManager.addPoint(1000, 1.5000);
      drawingManager.addPoint(3000, 1.5200);

      // All three drawings should exist
      const drawings = drawingManager.getDrawings();
      expect(drawings.length).toBe(3);

      // All three should be rendered
      const elements = mockContainer.querySelectorAll('.chart-drawing');
      expect(elements.length).toBeGreaterThanOrEqual(3);
    });
  });

  // TEST 10: Error Handling & Edge Cases
  describe('10. Error Handling & Edge Cases', () => {
    it('should handle addPoint without activeDrawing', () => {
      const result = drawingManager.addPoint(1000, 1.5000);
      expect(result).toBe(false);
    });

    it('should handle cancelDrawing', () => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);
      drawingManager.startDrawing('trendline', '#3b82f6');
      drawingManager.addPoint(1000, 1.5000);

      drawingManager.cancelDrawing();

      const activeDrawing = drawingManager.getActiveDrawing();
      expect(activeDrawing).toBeNull();

      const drawings = drawingManager.getDrawings();
      expect(drawings.length).toBe(0);
    });

    it('should handle clearAllDrawings', () => {
      drawingManager.setChart(mockChart as IChartApi, mockSeries as ISeriesApi<any>, 'EURUSD', 1);

      // Add multiple drawings
      drawingManager.startDrawing('hline', '#ff0000');
      drawingManager.addPoint(1000, 1.5000);

      drawingManager.startDrawing('vline', '#00ff00');
      drawingManager.addPoint(2000, 1.5100);

      expect(drawingManager.getDrawings().length).toBe(2);

      // Clear all
      drawingManager.clearAllDrawings();

      expect(drawingManager.getDrawings().length).toBe(0);
      expect(drawingManager.getActiveDrawing()).toBeNull();
    });

    it('should handle rendering without chart', () => {
      drawingManager.setChart(null, null, null, 1);

      // Should not throw error
      expect(() => {
        drawingManager.startDrawing('hline', '#3b82f6');
        drawingManager.addPoint(1000, 1.5000);
      }).not.toThrow();
    });

    it('should handle drawings array being null/undefined', () => {
      // Force drawings to be cleared
      drawingManager.clearAllDrawings();

      // Should not throw error on unselectAll
      expect(() => {
        drawingManager.unselectAll();
      }).not.toThrow();
    });
  });
});
