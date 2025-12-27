/**
 * DataSyncService - Background service for periodic data synchronization
 * Handles: Startup sync, 24-hour periodic refresh, retry logic, error handling
 */

import { DataCache, type CachedCandle } from './DataCache';

const API_BASE = 'http://localhost:8080';
const SYNC_INTERVAL_MS = 24 * 60 * 60 * 1000; // 24 hours
const RETRY_DELAYS = [1000, 3000, 10000]; // Retry delays in ms
const MAX_RETRIES = 3;

interface SyncStatus {
    symbol: string;
    timeframe: string;
    status: 'idle' | 'syncing' | 'success' | 'error';
    lastSync: number | null;
    error: string | null;
}

type SyncCallback = (symbol: string, timeframe: string, candles: CachedCandle[]) => void;

class DataSyncServiceClass {
    private syncStatuses = new Map<string, SyncStatus>();
    private syncTimer: ReturnType<typeof setInterval> | null = null;
    private listeners: SyncCallback[] = [];

    /**
     * Initialize sync service and start background worker
     */
    async start(): Promise<void> {
        console.log('[DataSync] Starting background sync service');

        // Initialize cache
        await DataCache.init();

        // Cleanup old data on startup
        await DataCache.cleanupOldData();

        // Start periodic sync timer (every 24 hours)
        if (!this.syncTimer) {
            this.syncTimer = setInterval(() => {
                this.syncAllSymbols();
            }, SYNC_INTERVAL_MS);
        }
    }

    /**
     * Stop sync service
     */
    stop(): void {
        if (this.syncTimer) {
            clearInterval(this.syncTimer);
            this.syncTimer = null;
        }
        console.log('[DataSync] Stopped');
    }

    /**
     * Subscribe to data updates
     */
    onDataUpdate(callback: SyncCallback): () => void {
        this.listeners.push(callback);
        return () => {
            this.listeners = this.listeners.filter(cb => cb !== callback);
        };
    }

    /**
     * Sync data for a specific symbol/timeframe
     * Called on chart load or manually
     */
    async syncSymbol(symbol: string, timeframe: string, days = 30): Promise<CachedCandle[]> {
        const key = `${symbol}:${timeframe}`;

        // Update status
        this.syncStatuses.set(key, {
            symbol,
            timeframe,
            status: 'syncing',
            lastSync: null,
            error: null
        });

        try {
            // Check cache first
            const cached = await DataCache.getCandles(symbol, timeframe);
            const meta = await DataCache.getSyncMeta(symbol, timeframe);

            // Calculate what we need to fetch
            const now = Date.now();
            const thirtyDaysAgo = now - (days * 24 * 60 * 60 * 1000);

            let fromTime = thirtyDaysAgo;
            let needsFetch = true;

            // If we have cached data, only fetch what's missing
            if (meta && cached.length > 0) {
                const age = now - meta.lastSync;

                // If last sync was within 1 hour and we have recent data, use cache
                if (age < 60 * 60 * 1000 && meta.newestTime > now - 5 * 60 * 1000) {
                    console.log(`[DataSync] Using cached data for ${symbol}/${timeframe}`);
                    needsFetch = false;
                } else {
                    // Fetch only from last known time
                    fromTime = meta.newestTime;
                }
            }

            // Fetch missing data with retry
            let candles = cached;
            if (needsFetch) {
                const newCandles = await this.fetchWithRetry(symbol, timeframe, fromTime, now);

                if (newCandles.length > 0) {
                    // Merge with existing
                    const merged = this.mergeCandles(cached, newCandles);

                    // Store in cache
                    await DataCache.storeCandles(symbol, timeframe, merged);
                    candles = merged;
                }
            }

            // Fill any gaps
            const filled = DataCache.fillGaps(candles, timeframe);

            // Update status
            this.syncStatuses.set(key, {
                symbol,
                timeframe,
                status: 'success',
                lastSync: Date.now(),
                error: null
            });

            // Notify listeners
            this.notifyListeners(symbol, timeframe, filled);

            return filled;

        } catch (error) {
            const errorMsg = error instanceof Error ? error.message : 'Unknown error';
            console.error(`[DataSync] Failed to sync ${symbol}/${timeframe}:`, errorMsg);

            this.syncStatuses.set(key, {
                symbol,
                timeframe,
                status: 'error',
                lastSync: null,
                error: errorMsg
            });

            // Return cached data as fallback
            const fallback = await DataCache.getCandles(symbol, timeframe);
            if (fallback.length > 0) {
                console.log(`[DataSync] Returning cached fallback for ${symbol}/${timeframe}`);
                return DataCache.fillGaps(fallback, timeframe);
            }

            throw error;
        }
    }

    /**
     * Sync all common trading symbols for all timeframes
     * Call this for a full data sync
     */
    async syncAllSymbols(): Promise<void> {
        const symbols = [
            'BTCUSD', 'ETHUSD', 'XRPUSD', 'SOLUSD', 'DOGEUSD',
            'EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD',
            'XAUUSD', 'XAGUSD'
        ];
        const timeframes = ['1m', '5m', '15m', '1h', '4h', '1d'];

        console.log(`[DataSync] Starting full sync for ${symbols.length} symbols × ${timeframes.length} timeframes`);

        for (const symbol of symbols) {
            for (const timeframe of timeframes) {
                try {
                    await this.syncSymbol(symbol, timeframe, 30);
                    console.log(`[DataSync] ✓ ${symbol}/${timeframe}`);
                } catch (err) {
                    console.warn(`[DataSync] ✗ ${symbol}/${timeframe}:`, err);
                }
                // Small delay to avoid overwhelming backend
                await this.sleep(100);
            }
        }

        console.log('[DataSync] Full sync complete!');
    }

    /**
     * Fetch data from API with retry logic
     */
    private async fetchWithRetry(
        symbol: string,
        timeframe: string,
        from: number,
        to: number
    ): Promise<CachedCandle[]> {
        let lastError: Error | null = null;

        for (let attempt = 0; attempt < MAX_RETRIES; attempt++) {
            try {
                return await this.fetchOHLC(symbol, timeframe, from, to);
            } catch (error) {
                lastError = error instanceof Error ? error : new Error('Unknown error');
                console.warn(`[DataSync] Fetch attempt ${attempt + 1} failed:`, lastError.message);

                if (attempt < MAX_RETRIES - 1) {
                    const delay = RETRY_DELAYS[attempt] || RETRY_DELAYS[RETRY_DELAYS.length - 1];
                    await this.sleep(delay);
                }
            }
        }

        throw lastError || new Error('Max retries exceeded');
    }

    /**
     * Fetch OHLC data from backend
     */
    private async fetchOHLC(
        symbol: string,
        timeframe: string,
        from: number,
        to: number
    ): Promise<CachedCandle[]> {
        const url = new URL(`${API_BASE}/ohlc`);
        url.searchParams.set('symbol', symbol);
        url.searchParams.set('timeframe', timeframe);
        url.searchParams.set('limit', '5000');

        const response = await fetch(url.toString());

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const data = await response.json();

        if (!Array.isArray(data)) {
            return [];
        }

        // Transform to CachedCandle format
        return data.map((candle: any) => ({
            symbol,
            timeframe,
            time: candle.time * 1000, // Convert to ms if in seconds
            open: candle.open,
            high: candle.high,
            low: candle.low,
            close: candle.close
        })).filter((c: CachedCandle) => c.time >= from && c.time <= to);
    }

    /**
     * Merge existing candles with new ones (deduplicate by time)
     */
    private mergeCandles(existing: CachedCandle[], newCandles: CachedCandle[]): CachedCandle[] {
        const map = new Map<number, CachedCandle>();

        for (const candle of existing) {
            map.set(candle.time, candle);
        }

        for (const candle of newCandles) {
            map.set(candle.time, candle); // New data overwrites old
        }

        return Array.from(map.values()).sort((a, b) => a.time - b.time);
    }

    /**
     * Get sync status for a symbol/timeframe
     */
    getSyncStatus(symbol: string, timeframe: string): SyncStatus | null {
        return this.syncStatuses.get(`${symbol}:${timeframe}`) || null;
    }

    /**
     * Notify all listeners of data update
     */
    private notifyListeners(symbol: string, timeframe: string, candles: CachedCandle[]): void {
        for (const callback of this.listeners) {
            try {
                callback(symbol, timeframe, candles);
            } catch (e) {
                console.error('[DataSync] Listener error:', e);
            }
        }
    }

    private sleep(ms: number): Promise<void> {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

// Singleton instance
export const DataSyncService = new DataSyncServiceClass();

// Expose to window for manual triggering from console
if (typeof window !== 'undefined') {
    (window as any).DataSync = DataSyncService;
}
