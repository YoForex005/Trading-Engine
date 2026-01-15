import type { ChartIndicator } from '../hooks/useIndicators';

// Storage key format: indicators:{symbol}:{timeframe}
const STORAGE_KEY_PREFIX = 'indicators';

// Indicator configuration for storage (without calculated data)
type StoredIndicatorConfig = Omit<ChartIndicator, 'id'> & {
  // Store type and params to recreate indicator
  savedAt: number;
};

export class IndicatorStorage {
  // Generate storage key
  private static getKey(symbol: string, timeframe: string): string {
    return `${STORAGE_KEY_PREFIX}:${symbol}:${timeframe}`;
  }

  // Save indicators for a symbol/timeframe
  static save(symbol: string, timeframe: string, indicators: ChartIndicator[]): void {
    try {
      const key = this.getKey(symbol, timeframe);
      const configs: StoredIndicatorConfig[] = indicators.map((ind) => ({
        type: ind.type,
        params: ind.params,
        displayMode: ind.displayMode,
        visible: ind.visible,
        color: ind.color,
        lineWidth: ind.lineWidth,
        paneIndex: ind.paneIndex,
        savedAt: Date.now(),
      }));

      localStorage.setItem(key, JSON.stringify(configs));
    } catch (error) {
      console.error('Error saving indicators:', error);
    }
  }

  // Load indicators for a symbol/timeframe
  static load(symbol: string, timeframe: string): Partial<ChartIndicator>[] {
    try {
      const key = this.getKey(symbol, timeframe);
      const stored = localStorage.getItem(key);

      if (!stored) return [];

      const configs: StoredIndicatorConfig[] = JSON.parse(stored);
      return configs.map((config) => ({
        type: config.type,
        params: config.params,
        displayMode: config.displayMode,
        visible: config.visible,
        color: config.color,
        lineWidth: config.lineWidth,
        paneIndex: config.paneIndex,
      }));
    } catch (error) {
      console.error('Error loading indicators:', error);
      return [];
    }
  }

  // Clear indicators for a symbol/timeframe
  static clear(symbol: string, timeframe: string): void {
    try {
      const key = this.getKey(symbol, timeframe);
      localStorage.removeItem(key);
    } catch (error) {
      console.error('Error clearing indicators:', error);
    }
  }

  // Get all saved symbol/timeframe combinations
  static getAllSaved(): Array<{ symbol: string; timeframe: string }> {
    try {
      const keys: Array<{ symbol: string; timeframe: string }> = [];

      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key && key.startsWith(STORAGE_KEY_PREFIX)) {
          const parts = key.split(':');
          if (parts.length === 3) {
            keys.push({ symbol: parts[1], timeframe: parts[2] });
          }
        }
      }

      return keys;
    } catch (error) {
      console.error('Error getting saved indicators:', error);
      return [];
    }
  }

  // Export all indicators to JSON (for backup/sharing)
  static exportAll(): string {
    try {
      const allData: Record<string, any> = {};

      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key && key.startsWith(STORAGE_KEY_PREFIX)) {
          const value = localStorage.getItem(key);
          if (value) {
            allData[key] = JSON.parse(value);
          }
        }
      }

      return JSON.stringify(allData, null, 2);
    } catch (error) {
      console.error('Error exporting indicators:', error);
      return '{}';
    }
  }

  // Import indicators from JSON
  static importAll(json: string): boolean {
    try {
      const data = JSON.parse(json);

      for (const [key, value] of Object.entries(data)) {
        if (key.startsWith(STORAGE_KEY_PREFIX)) {
          localStorage.setItem(key, JSON.stringify(value));
        }
      }

      return true;
    } catch (error) {
      console.error('Error importing indicators:', error);
      return false;
    }
  }
}
