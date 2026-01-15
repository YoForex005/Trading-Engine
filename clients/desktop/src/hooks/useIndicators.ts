import { useState, useEffect, useMemo, useCallback } from 'react';
import { IndicatorEngine } from '../indicators/core/IndicatorEngine';
import type {
  IndicatorType,
  IndicatorParams,
  IndicatorResult,
  IndicatorDisplayMode,
  OHLC,
} from '../indicators/core/IndicatorEngine';

// Chart indicator configuration
export type ChartIndicator = {
  id: string; // Unique ID for this indicator instance
  type: IndicatorType;
  params: IndicatorParams;
  displayMode: IndicatorDisplayMode;
  visible: boolean;
  color: string;
  lineWidth: number;
  paneIndex?: number; // For indicators in separate panes (0 = first pane, 1 = second, etc.)
};

// Calculated indicator data
export type CalculatedIndicator = ChartIndicator & {
  data: IndicatorResult;
  outputs: string[]; // e.g., ['MACD', 'signal', 'histogram']
};

// Hook options
type UseIndicatorsOptions = {
  ohlcData: OHLC[];
  autoCalculate?: boolean; // Auto-recalculate on data change
};

// Cache structure for performance
type IndicatorCache = {
  [key: string]: {
    data: IndicatorResult;
    timestamp: number;
    dataHash: string; // Hash of input data to detect changes
  };
};

// Simple hash function for OHLC data
function hashOHLCData(data: OHLC[]): string {
  if (data.length === 0) return '';
  // Hash based on length, first, and last candle
  const first = data[0];
  const last = data[data.length - 1];
  return `${data.length}-${first.time}-${first.close}-${last.time}-${last.close}`;
}

// Generate unique ID for indicator
function generateIndicatorId(type: IndicatorType, params: IndicatorParams): string {
  const paramStr = Object.entries(params)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([k, v]) => `${k}=${v}`)
    .join(',');
  return `${type}-${paramStr}-${Date.now()}`;
}

// Default colors for indicators
const DEFAULT_COLORS = [
  '#3b82f6', // blue
  '#ef4444', // red
  '#10b981', // green
  '#f59e0b', // amber
  '#8b5cf6', // purple
  '#06b6d4', // cyan
  '#f97316', // orange
  '#ec4899', // pink
];

export function useIndicators({
  ohlcData,
  autoCalculate = true,
}: UseIndicatorsOptions) {
  const [indicators, setIndicators] = useState<ChartIndicator[]>([]);
  const [calculatedData, setCalculatedData] = useState<Map<string, CalculatedIndicator>>(
    new Map()
  );
  const [cache, setCache] = useState<IndicatorCache>({});
  const [isCalculating, setIsCalculating] = useState(false);

  // Data hash for cache invalidation
  const dataHash = useMemo(() => hashOHLCData(ohlcData), [ohlcData]);

  // Add indicator
  const addIndicator = useCallback(
    (type: IndicatorType, params?: IndicatorParams, options?: Partial<ChartIndicator>) => {
      const meta = IndicatorEngine.getMeta(type);
      const finalParams = { ...meta.defaultParams, ...params };
      const id = generateIndicatorId(type, finalParams);

      const newIndicator: ChartIndicator = {
        id,
        type,
        params: finalParams,
        displayMode: meta.displayMode,
        visible: true,
        color: DEFAULT_COLORS[indicators.length % DEFAULT_COLORS.length],
        lineWidth: 2,
        ...options,
      };

      setIndicators((prev) => [...prev, newIndicator]);
      return id;
    },
    [indicators.length]
  );

  // Remove indicator
  const removeIndicator = useCallback((id: string) => {
    setIndicators((prev) => prev.filter((ind) => ind.id !== id));
    setCalculatedData((prev) => {
      const newMap = new Map(prev);
      newMap.delete(id);
      return newMap;
    });
  }, []);

  // Update indicator
  const updateIndicator = useCallback(
    (id: string, updates: Partial<ChartIndicator>) => {
      setIndicators((prev) =>
        prev.map((ind) => (ind.id === id ? { ...ind, ...updates } : ind))
      );

      // If params changed, invalidate cache for this indicator
      if (updates.params) {
        setCache((prev) => {
          const newCache = { ...prev };
          delete newCache[id];
          return newCache;
        });
      }
    },
    []
  );

  // Toggle indicator visibility
  const toggleIndicator = useCallback((id: string) => {
    setIndicators((prev) =>
      prev.map((ind) => (ind.id === id ? { ...ind, visible: !ind.visible } : ind))
    );
  }, []);

  // Clear all indicators
  const clearIndicators = useCallback(() => {
    setIndicators([]);
    setCalculatedData(new Map());
    setCache({});
  }, []);

  // Calculate single indicator
  const calculateIndicator = useCallback(
    (indicator: ChartIndicator): CalculatedIndicator | null => {
      const cacheKey = indicator.id;

      // Check cache
      const cached = cache[cacheKey];
      if (cached && cached.dataHash === dataHash) {
        const meta = IndicatorEngine.getMeta(indicator.type);
        return {
          ...indicator,
          data: cached.data,
          outputs: meta.outputs,
        };
      }

      // Calculate
      try {
        const data = IndicatorEngine.calculate(indicator.type, ohlcData, indicator.params);
        const meta = IndicatorEngine.getMeta(indicator.type);

        // Update cache
        setCache((prev) => ({
          ...prev,
          [cacheKey]: {
            data,
            timestamp: Date.now(),
            dataHash,
          },
        }));

        return {
          ...indicator,
          data,
          outputs: meta.outputs,
        };
      } catch (error) {
        console.error(`Error calculating indicator ${indicator.type}:`, error);
        return null;
      }
    },
    [cache, dataHash, ohlcData]
  );

  // Calculate all indicators
  const calculateAll = useCallback(() => {
    if (ohlcData.length === 0) return;

    setIsCalculating(true);

    const newCalculatedData = new Map<string, CalculatedIndicator>();

    for (const indicator of indicators) {
      const calculated = calculateIndicator(indicator);
      if (calculated) {
        newCalculatedData.set(indicator.id, calculated);
      }
    }

    setCalculatedData(newCalculatedData);
    setIsCalculating(false);
  }, [indicators, calculateIndicator, ohlcData.length]);

  // Auto-calculate when data or indicators change
  useEffect(() => {
    if (autoCalculate) {
      calculateAll();
    }
  }, [autoCalculate, calculateAll]);

  // Get indicators by display mode
  const overlayIndicators = useMemo(
    () => Array.from(calculatedData.values()).filter((ind) => ind.displayMode === 'overlay'),
    [calculatedData]
  );

  const paneIndicators = useMemo(
    () => Array.from(calculatedData.values()).filter((ind) => ind.displayMode === 'pane'),
    [calculatedData]
  );

  // Get indicators grouped by pane
  const paneGroups = useMemo(() => {
    const groups = new Map<number, CalculatedIndicator[]>();

    for (const indicator of paneIndicators) {
      const paneIndex = indicator.paneIndex ?? 0;
      const group = groups.get(paneIndex) || [];
      group.push(indicator);
      groups.set(paneIndex, group);
    }

    return groups;
  }, [paneIndicators]);

  return {
    // State
    indicators,
    calculatedData,
    overlayIndicators,
    paneIndicators,
    paneGroups,
    isCalculating,

    // Actions
    addIndicator,
    removeIndicator,
    updateIndicator,
    toggleIndicator,
    clearIndicators,
    calculateAll,
    calculateIndicator,

    // Utilities
    getIndicator: (id: string) => indicators.find((ind) => ind.id === id),
    getCalculatedIndicator: (id: string) => calculatedData.get(id),
  };
}
