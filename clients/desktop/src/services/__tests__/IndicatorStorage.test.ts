import { describe, it, expect, beforeEach, vi } from 'vitest';

// Mock localStorage before importing IndicatorStorage
class LocalStorageMock implements Storage {
  private store = new Map<string, string>();

  getItem(key: string): string | null {
    return this.store.get(key) ?? null;
  }

  setItem(key: string, value: string): void {
    this.store.set(key, value);
  }

  removeItem(key: string): void {
    this.store.delete(key);
  }

  clear(): void {
    this.store.clear();
  }

  key(index: number): string | null {
    const keys = Array.from(this.store.keys());
    return keys[index] ?? null;
  }

  get length(): number {
    return this.store.size;
  }
}

const localStorageMock = new LocalStorageMock();
globalThis.localStorage = localStorageMock as any;

import { IndicatorStorage } from '../IndicatorStorage';
import type { ChartIndicator } from '../../hooks/useIndicators';

describe('IndicatorStorage', () => {
  // Mock indicators for testing
  const mockIndicators: ChartIndicator[] = [
    {
      id: 'sma-20-1234',
      type: 'SMA',
      params: { period: 20 },
      displayMode: 'overlay',
      visible: true,
      color: '#3b82f6',
      lineWidth: 2,
    },
    {
      id: 'rsi-14-5678',
      type: 'RSI',
      params: { period: 14 },
      displayMode: 'pane',
      visible: true,
      color: '#ef4444',
      lineWidth: 2,
      paneIndex: 0,
    },
  ];

  beforeEach(() => {
    localStorageMock.clear();
    vi.clearAllMocks();
  });

  describe('save', () => {
    it('should save indicators to localStorage', () => {
      IndicatorStorage.save('EURUSD', '1h', mockIndicators);

      const key = 'indicators:EURUSD:1h';
      const stored = localStorage.getItem(key);

      expect(stored).not.toBeNull();

      const parsed = JSON.parse(stored!);
      expect(parsed).toHaveLength(2);
      expect(parsed[0].type).toBe('SMA');
      expect(parsed[1].type).toBe('RSI');
    });

    it('should include savedAt timestamp', () => {
      const beforeSave = Date.now();
      IndicatorStorage.save('GBPUSD', '15m', mockIndicators);
      const afterSave = Date.now();

      const stored = localStorage.getItem('indicators:GBPUSD:15m');
      const parsed = JSON.parse(stored!);

      expect(parsed[0].savedAt).toBeGreaterThanOrEqual(beforeSave);
      expect(parsed[0].savedAt).toBeLessThanOrEqual(afterSave);
    });

    it('should not include id field in stored data', () => {
      IndicatorStorage.save('USDJPY', '5m', mockIndicators);

      const stored = localStorage.getItem('indicators:USDJPY:5m');
      const parsed = JSON.parse(stored!);

      expect(parsed[0]).not.toHaveProperty('id');
      expect(parsed[1]).not.toHaveProperty('id');
    });

    it('should handle empty indicators array', () => {
      IndicatorStorage.save('BTCUSD', '1d', []);

      const stored = localStorage.getItem('indicators:BTCUSD:1d');
      const parsed = JSON.parse(stored!);

      expect(parsed).toEqual([]);
    });

    it('should handle localStorage errors gracefully', () => {
      // Mock localStorage.setItem to throw error
      const originalSetItem = localStorage.setItem;
      localStorage.setItem = vi.fn(() => {
        throw new Error('localStorage is full');
      });

      // Should not throw
      expect(() => {
        IndicatorStorage.save('ETHUSD', '1h', mockIndicators);
      }).not.toThrow();

      // Restore original
      localStorage.setItem = originalSetItem;
    });
  });

  describe('load', () => {
    it('should load saved indicators', () => {
      IndicatorStorage.save('EURUSD', '1h', mockIndicators);
      const loaded = IndicatorStorage.load('EURUSD', '1h');

      expect(loaded).toHaveLength(2);
      expect(loaded[0].type).toBe('SMA');
      expect(loaded[0].params).toEqual({ period: 20 });
      expect(loaded[1].type).toBe('RSI');
      expect(loaded[1].params).toEqual({ period: 14 });
    });

    it('should return empty array for non-existent key', () => {
      const loaded = IndicatorStorage.load('XAUUSD', '4h');

      expect(loaded).toEqual([]);
    });

    it('should handle corrupted data gracefully', () => {
      // Save corrupted JSON
      localStorage.setItem('indicators:CORRUPT:1h', 'invalid json{]');

      const loaded = IndicatorStorage.load('CORRUPT', '1h');
      expect(loaded).toEqual([]);
    });

    it('should not include savedAt in loaded data', () => {
      IndicatorStorage.save('GBPJPY', '30m', mockIndicators);
      const loaded = IndicatorStorage.load('GBPJPY', '30m');

      expect(loaded[0]).not.toHaveProperty('savedAt');
    });

    it('should preserve all indicator properties', () => {
      IndicatorStorage.save('AUDCAD', '1h', mockIndicators);
      const loaded = IndicatorStorage.load('AUDCAD', '1h');

      expect(loaded[0]).toMatchObject({
        type: 'SMA',
        params: { period: 20 },
        displayMode: 'overlay',
        visible: true,
        color: '#3b82f6',
        lineWidth: 2,
      });

      expect(loaded[1]).toMatchObject({
        type: 'RSI',
        params: { period: 14 },
        displayMode: 'pane',
        visible: true,
        color: '#ef4444',
        lineWidth: 2,
        paneIndex: 0,
      });
    });
  });

  describe('clear', () => {
    it('should remove indicators from localStorage', () => {
      IndicatorStorage.save('NZDUSD', '1h', mockIndicators);

      expect(localStorage.getItem('indicators:NZDUSD:1h')).not.toBeNull();

      IndicatorStorage.clear('NZDUSD', '1h');

      expect(localStorage.getItem('indicators:NZDUSD:1h')).toBeNull();
    });

    it('should handle clearing non-existent key', () => {
      expect(() => {
        IndicatorStorage.clear('NONEXISTENT', '1h');
      }).not.toThrow();
    });

    it('should handle localStorage errors gracefully', () => {
      const originalRemoveItem = localStorage.removeItem;
      localStorage.removeItem = vi.fn(() => {
        throw new Error('localStorage error');
      });

      expect(() => {
        IndicatorStorage.clear('EURUSD', '1h');
      }).not.toThrow();

      localStorage.removeItem = originalRemoveItem;
    });
  });

  describe('getAllSaved', () => {
    it('should return all saved symbol/timeframe combinations', () => {
      IndicatorStorage.save('EURUSD', '1h', mockIndicators);
      IndicatorStorage.save('GBPUSD', '15m', mockIndicators);
      IndicatorStorage.save('USDJPY', '5m', mockIndicators);

      const all = IndicatorStorage.getAllSaved();

      expect(all).toHaveLength(3);
      expect(all).toContainEqual({ symbol: 'EURUSD', timeframe: '1h' });
      expect(all).toContainEqual({ symbol: 'GBPUSD', timeframe: '15m' });
      expect(all).toContainEqual({ symbol: 'USDJPY', timeframe: '5m' });
    });

    it('should return empty array when no indicators saved', () => {
      const all = IndicatorStorage.getAllSaved();
      expect(all).toEqual([]);
    });

    it('should ignore non-indicator localStorage keys', () => {
      localStorage.setItem('other:key', 'value');
      localStorage.setItem('indicators:EURUSD:1h', '[]');

      const all = IndicatorStorage.getAllSaved();
      expect(all).toHaveLength(1);
      expect(all[0]).toEqual({ symbol: 'EURUSD', timeframe: '1h' });
    });

    it('should handle malformed keys gracefully', () => {
      localStorage.setItem('indicators:malformed', 'value');
      localStorage.setItem('indicators:EURUSD:1h', '[]');

      const all = IndicatorStorage.getAllSaved();
      expect(all).toHaveLength(1);
      expect(all[0]).toEqual({ symbol: 'EURUSD', timeframe: '1h' });
    });
  });

  describe('exportAll', () => {
    it('should export all indicators as JSON', () => {
      IndicatorStorage.save('EURUSD', '1h', mockIndicators);
      IndicatorStorage.save('GBPUSD', '15m', [mockIndicators[0]]);

      const exported = IndicatorStorage.exportAll();
      const parsed = JSON.parse(exported);

      expect(parsed).toHaveProperty('indicators:EURUSD:1h');
      expect(parsed).toHaveProperty('indicators:GBPUSD:15m');
      expect(parsed['indicators:EURUSD:1h']).toHaveLength(2);
      expect(parsed['indicators:GBPUSD:15m']).toHaveLength(1);
    });

    it('should return empty object when nothing saved', () => {
      const exported = IndicatorStorage.exportAll();
      expect(JSON.parse(exported)).toEqual({});
    });

    it('should handle export errors gracefully', () => {
      const originalGetItem = localStorage.getItem;
      localStorage.getItem = vi.fn(() => {
        throw new Error('localStorage error');
      });

      const exported = IndicatorStorage.exportAll();
      expect(exported).toBe('{}');

      localStorage.getItem = originalGetItem;
    });

    it('should format JSON with indentation', () => {
      IndicatorStorage.save('EURUSD', '1h', mockIndicators);
      const exported = IndicatorStorage.exportAll();

      // Check that it's pretty-printed (contains newlines and spaces)
      expect(exported).toContain('\n');
      expect(exported).toContain('  ');
    });
  });

  describe('importAll', () => {
    it('should import indicators from JSON', () => {
      const exportData = {
        'indicators:EURUSD:1h': [
          { type: 'SMA', params: { period: 20 }, displayMode: 'overlay', visible: true, color: '#3b82f6', lineWidth: 2, savedAt: Date.now() },
        ],
        'indicators:GBPUSD:15m': [
          { type: 'RSI', params: { period: 14 }, displayMode: 'pane', visible: true, color: '#ef4444', lineWidth: 2, savedAt: Date.now() },
        ],
      };

      const result = IndicatorStorage.importAll(JSON.stringify(exportData));

      expect(result).toBe(true);
      expect(localStorage.getItem('indicators:EURUSD:1h')).not.toBeNull();
      expect(localStorage.getItem('indicators:GBPUSD:15m')).not.toBeNull();
    });

    it('should overwrite existing data on import', () => {
      IndicatorStorage.save('EURUSD', '1h', mockIndicators);

      const exportData = {
        'indicators:EURUSD:1h': [
          { type: 'EMA', params: { period: 50 }, displayMode: 'overlay', visible: true, color: '#10b981', lineWidth: 3, savedAt: Date.now() },
        ],
      };

      IndicatorStorage.importAll(JSON.stringify(exportData));

      const loaded = IndicatorStorage.load('EURUSD', '1h');
      expect(loaded).toHaveLength(1);
      expect(loaded[0].type).toBe('EMA');
    });

    it('should return false for invalid JSON', () => {
      const result = IndicatorStorage.importAll('invalid json{]');
      expect(result).toBe(false);
    });

    it('should ignore non-indicator keys on import', () => {
      const exportData = {
        'indicators:EURUSD:1h': [
          { type: 'SMA', params: { period: 20 }, displayMode: 'overlay', visible: true, color: '#3b82f6', lineWidth: 2, savedAt: Date.now() },
        ],
        'other:key': 'should be ignored',
      };

      IndicatorStorage.importAll(JSON.stringify(exportData));

      expect(localStorage.getItem('indicators:EURUSD:1h')).not.toBeNull();
      expect(localStorage.getItem('other:key')).toBeNull();
    });

    it('should handle import errors gracefully', () => {
      const originalSetItem = localStorage.setItem;
      localStorage.setItem = vi.fn(() => {
        throw new Error('localStorage full');
      });

      const exportData = { 'indicators:EURUSD:1h': [] };
      const result = IndicatorStorage.importAll(JSON.stringify(exportData));

      expect(result).toBe(false);

      localStorage.setItem = originalSetItem;
    });
  });

  describe('Round-trip save/load', () => {
    it('should preserve all data through save and load cycle', () => {
      const original = mockIndicators;
      IndicatorStorage.save('TEST', '1h', original);
      const loaded = IndicatorStorage.load('TEST', '1h');

      expect(loaded).toHaveLength(original.length);

      loaded.forEach((loadedInd, index) => {
        const originalInd = original[index];
        expect(loadedInd.type).toBe(originalInd.type);
        expect(loadedInd.params).toEqual(originalInd.params);
        expect(loadedInd.displayMode).toBe(originalInd.displayMode);
        expect(loadedInd.visible).toBe(originalInd.visible);
        expect(loadedInd.color).toBe(originalInd.color);
        expect(loadedInd.lineWidth).toBe(originalInd.lineWidth);
        expect(loadedInd.paneIndex).toBe(originalInd.paneIndex);
      });
    });
  });

  describe('Round-trip export/import', () => {
    it('should preserve all data through export and import cycle', () => {
      IndicatorStorage.save('EURUSD', '1h', mockIndicators);
      IndicatorStorage.save('GBPUSD', '15m', [mockIndicators[0]]);

      const exported = IndicatorStorage.exportAll();

      localStorage.clear();

      const importResult = IndicatorStorage.importAll(exported);
      expect(importResult).toBe(true);

      const loadedEUR = IndicatorStorage.load('EURUSD', '1h');
      const loadedGBP = IndicatorStorage.load('GBPUSD', '15m');

      expect(loadedEUR).toHaveLength(2);
      expect(loadedGBP).toHaveLength(1);

      expect(loadedEUR[0].type).toBe('SMA');
      expect(loadedEUR[1].type).toBe('RSI');
      expect(loadedGBP[0].type).toBe('SMA');
    });
  });
});
