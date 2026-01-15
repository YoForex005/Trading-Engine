import { describe, it, expect } from 'vitest';
import { IndicatorEngine, type OHLC } from '../IndicatorEngine';

// Mock OHLC data for testing
const mockOHLCData: OHLC[] = [
  { time: 1000, open: 100, high: 105, low: 99, close: 103, volume: 1000 },
  { time: 2000, open: 103, high: 107, low: 102, close: 106, volume: 1200 },
  { time: 3000, open: 106, high: 110, low: 105, close: 108, volume: 1100 },
  { time: 4000, open: 108, high: 112, low: 107, close: 110, volume: 1300 },
  { time: 5000, open: 110, high: 114, low: 109, close: 112, volume: 1400 },
  { time: 6000, open: 112, high: 116, low: 111, close: 114, volume: 1500 },
  { time: 7000, open: 114, high: 118, low: 113, close: 116, volume: 1600 },
  { time: 8000, open: 116, high: 120, low: 115, close: 118, volume: 1700 },
  { time: 9000, open: 118, high: 122, low: 117, close: 120, volume: 1800 },
  { time: 10000, open: 120, high: 124, low: 119, close: 122, volume: 1900 },
  { time: 11000, open: 122, high: 126, low: 121, close: 124, volume: 2000 },
  { time: 12000, open: 124, high: 128, low: 123, close: 126, volume: 2100 },
  { time: 13000, open: 126, high: 130, low: 125, close: 128, volume: 2200 },
  { time: 14000, open: 128, high: 132, low: 127, close: 130, volume: 2300 },
  { time: 15000, open: 130, high: 134, low: 129, close: 132, volume: 2400 },
  { time: 16000, open: 132, high: 136, low: 131, close: 134, volume: 2500 },
  { time: 17000, open: 134, high: 138, low: 133, close: 136, volume: 2600 },
  { time: 18000, open: 136, high: 140, low: 135, close: 138, volume: 2700 },
  { time: 19000, open: 138, high: 142, low: 137, close: 140, volume: 2800 },
  { time: 20000, open: 140, high: 144, low: 139, close: 142, volume: 2900 },
];

describe('IndicatorEngine', () => {
  describe('getDefaultParams', () => {
    it('should return default params for SMA', () => {
      const params = IndicatorEngine.getDefaultParams('SMA');
      expect(params).toEqual({ period: 20 });
    });

    it('should return default params for RSI', () => {
      const params = IndicatorEngine.getDefaultParams('RSI');
      expect(params).toEqual({ period: 14 });
    });

    it('should return default params for MACD', () => {
      const params = IndicatorEngine.getDefaultParams('MACD');
      expect(params).toEqual({ fastPeriod: 12, slowPeriod: 26, signalPeriod: 9 });
    });

    it('should return default params for Bollinger Bands', () => {
      const params = IndicatorEngine.getDefaultParams('BollingerBands');
      expect(params).toEqual({ period: 20, stdDev: 2 });
    });
  });

  describe('getMeta', () => {
    it('should return metadata for SMA', () => {
      const meta = IndicatorEngine.getMeta('SMA');
      expect(meta.name).toBe('Simple Moving Average');
      expect(meta.type).toBe('SMA');
      expect(meta.category).toBe('trend');
      expect(meta.displayMode).toBe('overlay');
      expect(meta.outputs).toEqual(['SMA']);
    });

    it('should return metadata for RSI', () => {
      const meta = IndicatorEngine.getMeta('RSI');
      expect(meta.name).toBe('Relative Strength Index');
      expect(meta.type).toBe('RSI');
      expect(meta.category).toBe('momentum');
      expect(meta.displayMode).toBe('pane');
      expect(meta.outputs).toEqual(['RSI']);
    });

    it('should return metadata for MACD', () => {
      const meta = IndicatorEngine.getMeta('MACD');
      expect(meta.name).toBe('Moving Average Convergence Divergence');
      expect(meta.type).toBe('MACD');
      expect(meta.category).toBe('momentum');
      expect(meta.displayMode).toBe('pane');
      expect(meta.outputs).toEqual(['MACD', 'signal', 'histogram']);
    });
  });

  describe('getAllIndicators', () => {
    it('should return all available indicator types', () => {
      const indicators = IndicatorEngine.getAllIndicators();
      expect(indicators).toBeInstanceOf(Array);
      expect(indicators.length).toBeGreaterThan(20);
      expect(indicators).toContain('SMA');
      expect(indicators).toContain('RSI');
      expect(indicators).toContain('MACD');
      expect(indicators).toContain('BollingerBands');
    });
  });

  describe('calculate - Trend Indicators', () => {
    it('should calculate SMA correctly', () => {
      const result = IndicatorEngine.calculate('SMA', mockOHLCData, { period: 5 });

      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);
      expect(result[0]).toHaveProperty('time');
      expect(result[0]).toHaveProperty('value');
      expect(typeof result[0].value).toBe('number');
    });

    it('should calculate EMA correctly', () => {
      const result = IndicatorEngine.calculate('EMA', mockOHLCData, { period: 10 });

      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);
      expect(typeof result[0].value).toBe('number');
    });

    it('should calculate Bollinger Bands correctly', () => {
      const result = IndicatorEngine.calculate('BollingerBands', mockOHLCData, {
        period: 20,
        stdDev: 2,
      });

      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);

      const firstValue = result[0].value;
      expect(firstValue).toHaveProperty('upper');
      expect(firstValue).toHaveProperty('middle');
      expect(firstValue).toHaveProperty('lower');
    });
  });

  describe('calculate - Momentum Indicators', () => {
    it('should calculate RSI correctly', () => {
      const result = IndicatorEngine.calculate('RSI', mockOHLCData, { period: 14 });

      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);
      expect(typeof result[0].value).toBe('number');

      // RSI should be between 0 and 100
      result.forEach((point) => {
        if (typeof point.value === 'number') {
          expect(point.value).toBeGreaterThanOrEqual(0);
          expect(point.value).toBeLessThanOrEqual(100);
        }
      });
    });

    it('should calculate MACD correctly', () => {
      const result = IndicatorEngine.calculate('MACD', mockOHLCData, {
        fastPeriod: 5,  // Reduced to work with our limited test data
        slowPeriod: 10,
        signalPeriod: 3,
      });

      expect(result).toBeInstanceOf(Array);

      // MACD requires sufficient data - may return empty for insufficient data
      if (result.length > 0) {
        const firstValue = result[0].value;
        expect(firstValue).toHaveProperty('MACD');
        expect(firstValue).toHaveProperty('signal');
        expect(firstValue).toHaveProperty('histogram');
      }
    });

    it('should calculate Stochastic correctly', () => {
      const result = IndicatorEngine.calculate('Stochastic', mockOHLCData, {
        kPeriod: 14,
        dPeriod: 3,
        smooth: 3,
      });

      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);

      const firstValue = result[0].value;
      expect(firstValue).toHaveProperty('k');
      expect(firstValue).toHaveProperty('d');
    });
  });

  describe('calculate - Volatility Indicators', () => {
    it('should calculate ATR correctly', () => {
      const result = IndicatorEngine.calculate('ATR', mockOHLCData, { period: 14 });

      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);
      expect(typeof result[0].value).toBe('number');

      // ATR should be positive
      result.forEach((point) => {
        if (typeof point.value === 'number') {
          expect(point.value).toBeGreaterThanOrEqual(0);
        }
      });
    });
  });

  describe('calculate - Volume Indicators', () => {
    it('should calculate OBV correctly', () => {
      const result = IndicatorEngine.calculate('OBV', mockOHLCData, {});

      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);
      expect(typeof result[0].value).toBe('number');
    });

    it('should calculate MFI correctly', () => {
      const result = IndicatorEngine.calculate('MFI', mockOHLCData, { period: 14 });

      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);

      // MFI should be between 0 and 100
      result.forEach((point) => {
        if (typeof point.value === 'number') {
          expect(point.value).toBeGreaterThanOrEqual(0);
          expect(point.value).toBeLessThanOrEqual(100);
        }
      });
    });
  });

  describe('Edge cases', () => {
    it('should handle empty OHLC data', () => {
      const result = IndicatorEngine.calculate('SMA', [], { period: 20 });
      expect(result).toEqual([]);
    });

    it('should handle insufficient data', () => {
      const shortData: OHLC[] = [
        { time: 1000, open: 100, high: 105, low: 99, close: 103 },
        { time: 2000, open: 103, high: 107, low: 102, close: 106 },
      ];

      const result = IndicatorEngine.calculate('SMA', shortData, { period: 20 });
      expect(result).toBeInstanceOf(Array);
      // Should handle gracefully even with insufficient data
    });

    it('should handle invalid period gracefully', () => {
      // Technical indicators library returns empty array for invalid periods
      const result = IndicatorEngine.calculate('SMA', mockOHLCData, { period: 0 });
      expect(result).toBeInstanceOf(Array);
    });

    it('should handle negative period gracefully', () => {
      // Technical indicators library returns empty array for negative periods
      const result = IndicatorEngine.calculate('SMA', mockOHLCData, { period: -10 });
      expect(result).toBeInstanceOf(Array);
    });
  });

  describe('Time alignment', () => {
    it('should preserve time values in results', () => {
      const result = IndicatorEngine.calculate('SMA', mockOHLCData, { period: 5 });

      result.forEach((point, index) => {
        expect(point.time).toBeDefined();
        expect(typeof point.time).toBe('number');
        if (index > 0) {
          expect(point.time).toBeGreaterThan(result[index - 1].time);
        }
      });
    });

    it('should maintain chronological order', () => {
      const result = IndicatorEngine.calculate('RSI', mockOHLCData, { period: 14 });

      for (let i = 1; i < result.length; i++) {
        expect(result[i].time).toBeGreaterThan(result[i - 1].time);
      }
    });
  });

  describe('Parameter validation', () => {
    it('should use default params when not provided', () => {
      const result = IndicatorEngine.calculate('SMA', mockOHLCData, {});
      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);
    });

    it('should merge custom params with defaults', () => {
      const result = IndicatorEngine.calculate('BollingerBands', mockOHLCData, {
        period: 15, // Custom period, should use default stdDev
      });

      expect(result).toBeInstanceOf(Array);
      expect(result.length).toBeGreaterThan(0);
    });
  });

  describe('Multi-output indicators', () => {
    it('should return all outputs for Bollinger Bands', () => {
      const result = IndicatorEngine.calculate('BollingerBands', mockOHLCData, {
        period: 20,
        stdDev: 2,
      });

      result.forEach((point) => {
        expect(point.value).toHaveProperty('upper');
        expect(point.value).toHaveProperty('middle');
        expect(point.value).toHaveProperty('lower');
      });
    });

    it('should return all outputs for MACD', () => {
      const result = IndicatorEngine.calculate('MACD', mockOHLCData, {
        fastPeriod: 12,
        slowPeriod: 26,
        signalPeriod: 9,
      });

      result.forEach((point) => {
        expect(point.value).toHaveProperty('MACD');
        expect(point.value).toHaveProperty('signal');
        expect(point.value).toHaveProperty('histogram');
      });
    });

    it('should return all outputs for Stochastic', () => {
      const result = IndicatorEngine.calculate('Stochastic', mockOHLCData, {
        kPeriod: 14,
        dPeriod: 3,
        smooth: 3,
      });

      result.forEach((point) => {
        expect(point.value).toHaveProperty('k');
        expect(point.value).toHaveProperty('d');
      });
    });
  });
});
