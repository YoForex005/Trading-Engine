/**
 * useChartData - Hook for fetching chart data with caching and gap handling
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import { DataSyncService } from './DataSyncService';
import { DataCache, type CachedCandle } from './DataCache';

interface UseChartDataOptions {
    symbol: string;
    timeframe: string;
    enabled?: boolean;
}

interface UseChartDataResult {
    candles: CachedCandle[];
    isLoading: boolean;
    error: string | null;
    isCached: boolean;
    lastSync: number | null;
    refresh: () => Promise<void>;
}

export function useChartData({ symbol, timeframe, enabled = true }: UseChartDataOptions): UseChartDataResult {
    const [candles, setCandles] = useState<CachedCandle[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isCached, setIsCached] = useState(false);
    const [lastSync, setLastSync] = useState<number | null>(null);
    const mountedRef = useRef(true);

    const loadData = useCallback(async () => {
        if (!enabled || !symbol) return;

        setIsLoading(true);
        setError(null);

        try {
            // First, try to load from cache for instant display
            const cached = await DataCache.getCandles(symbol, timeframe);

            if (cached.length > 0 && mountedRef.current) {
                const filled = DataCache.fillGaps(cached, timeframe);
                setCandles(filled);
                setIsCached(true);
                setIsLoading(false);
            }

            // Then sync in background
            const synced = await DataSyncService.syncSymbol(symbol, timeframe);

            if (mountedRef.current) {
                setCandles(synced);
                setLastSync(Date.now());
                setIsCached(false);
                setIsLoading(false);
            }

        } catch (err) {
            if (mountedRef.current) {
                const errorMsg = err instanceof Error ? err.message : 'Failed to load data';
                setError(errorMsg);
                setIsLoading(false);

                // Try to use cached data as fallback
                const fallback = await DataCache.getCandles(symbol, timeframe);
                if (fallback.length > 0) {
                    setCandles(DataCache.fillGaps(fallback, timeframe));
                    setIsCached(true);
                }
            }
        }
    }, [symbol, timeframe, enabled]);

    // Initial load
    useEffect(() => {
        mountedRef.current = true;
        loadData();

        return () => {
            mountedRef.current = false;
        };
    }, [loadData]);

    // Subscribe to data updates from sync service
    useEffect(() => {
        const unsubscribe = DataSyncService.onDataUpdate((s, tf, data) => {
            if (s === symbol && tf === timeframe && mountedRef.current) {
                setCandles(data);
                setLastSync(Date.now());
                setIsCached(false);
            }
        });

        return unsubscribe;
    }, [symbol, timeframe]);

    return {
        candles,
        isLoading,
        error,
        isCached,
        lastSync,
        refresh: loadData
    };
}
