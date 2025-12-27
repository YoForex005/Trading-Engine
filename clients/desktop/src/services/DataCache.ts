/**
 * DataCache - IndexedDB-based cache for OHLC chart data
 * Stores up to 30 days of historical data per symbol/timeframe
 */

const DB_NAME = 'TradingEngineCache';
const DB_VERSION = 1;
const STORE_NAME = 'ohlc_data';
const META_STORE = 'sync_meta';
const MAX_AGE_DAYS = 30;

interface CachedCandle {
    symbol: string;
    timeframe: string;
    time: number;
    open: number;
    high: number;
    low: number;
    close: number;
}

interface SyncMeta {
    symbol: string;
    timeframe: string;
    lastSync: number;
    oldestTime: number;
    newestTime: number;
}

class DataCacheService {
    private db: IDBDatabase | null = null;
    private initPromise: Promise<void> | null = null;

    async init(): Promise<void> {
        if (this.db) return;
        if (this.initPromise) return this.initPromise;

        this.initPromise = new Promise((resolve, reject) => {
            const request = indexedDB.open(DB_NAME, DB_VERSION);

            request.onerror = () => reject(request.error);
            request.onsuccess = () => {
                this.db = request.result;
                console.log('[DataCache] Initialized');
                resolve();
            };

            request.onupgradeneeded = (event) => {
                const db = (event.target as IDBOpenDBRequest).result;

                // OHLC data store with compound index
                if (!db.objectStoreNames.contains(STORE_NAME)) {
                    const store = db.createObjectStore(STORE_NAME, { keyPath: ['symbol', 'timeframe', 'time'] });
                    store.createIndex('by_symbol_tf', ['symbol', 'timeframe'], { unique: false });
                    store.createIndex('by_time', 'time', { unique: false });
                }

                // Sync metadata store
                if (!db.objectStoreNames.contains(META_STORE)) {
                    db.createObjectStore(META_STORE, { keyPath: ['symbol', 'timeframe'] });
                }
            };
        });

        return this.initPromise;
    }

    /**
     * Store candles in IndexedDB
     */
    async storeCandles(symbol: string, timeframe: string, candles: CachedCandle[]): Promise<void> {
        await this.init();
        if (!this.db || !candles.length) return;

        return new Promise((resolve, reject) => {
            const tx = this.db!.transaction([STORE_NAME, META_STORE], 'readwrite');
            const store = tx.objectStore(STORE_NAME);
            const metaStore = tx.objectStore(META_STORE);

            // Add all candles
            for (const candle of candles) {
                store.put({ ...candle, symbol, timeframe });
            }

            // Update sync meta
            const times = candles.map(c => c.time);
            const meta: SyncMeta = {
                symbol,
                timeframe,
                lastSync: Date.now(),
                oldestTime: Math.min(...times),
                newestTime: Math.max(...times)
            };
            metaStore.put(meta);

            tx.oncomplete = () => {
                console.log(`[DataCache] Stored ${candles.length} candles for ${symbol}/${timeframe}`);
                resolve();
            };
            tx.onerror = () => reject(tx.error);
        });
    }

    /**
     * Get candles from cache
     */
    async getCandles(symbol: string, timeframe: string, from?: number, to?: number): Promise<CachedCandle[]> {
        await this.init();
        if (!this.db) return [];

        return new Promise((resolve, reject) => {
            const tx = this.db!.transaction(STORE_NAME, 'readonly');
            const store = tx.objectStore(STORE_NAME);
            const index = store.index('by_symbol_tf');
            const range = IDBKeyRange.only([symbol, timeframe]);
            const request = index.openCursor(range);

            const results: CachedCandle[] = [];

            request.onsuccess = () => {
                const cursor = request.result;
                if (cursor) {
                    const candle = cursor.value as CachedCandle;
                    // Filter by time range if specified
                    if ((!from || candle.time >= from) && (!to || candle.time <= to)) {
                        results.push(candle);
                    }
                    cursor.continue();
                } else {
                    // Sort by time
                    results.sort((a, b) => a.time - b.time);
                    resolve(results);
                }
            };
            request.onerror = () => reject(request.error);
        });
    }

    /**
     * Get sync metadata
     */
    async getSyncMeta(symbol: string, timeframe: string): Promise<SyncMeta | null> {
        await this.init();
        if (!this.db) return null;

        return new Promise((resolve, reject) => {
            const tx = this.db!.transaction(META_STORE, 'readonly');
            const store = tx.objectStore(META_STORE);
            const request = store.get([symbol, timeframe]);

            request.onsuccess = () => resolve(request.result || null);
            request.onerror = () => reject(request.error);
        });
    }

    /**
     * Detect gaps in data and return missing ranges
     */
    detectGaps(candles: CachedCandle[], timeframe: string): { from: number; to: number }[] {
        if (candles.length < 2) return [];

        const intervalMs = this.getIntervalMs(timeframe);
        const gaps: { from: number; to: number }[] = [];

        for (let i = 1; i < candles.length; i++) {
            const expected = candles[i - 1].time + intervalMs;
            const actual = candles[i].time;

            // If gap is larger than 2x expected interval, it's a gap
            if (actual - expected > intervalMs * 1.5) {
                gaps.push({ from: expected, to: actual });
            }
        }

        return gaps;
    }

    /**
     * Fill gaps with last known value
     */
    fillGaps(candles: CachedCandle[], timeframe: string): CachedCandle[] {
        if (candles.length < 2) return candles;

        const intervalMs = this.getIntervalMs(timeframe);
        const filled: CachedCandle[] = [];

        for (let i = 0; i < candles.length; i++) {
            filled.push(candles[i]);

            if (i < candles.length - 1) {
                const current = candles[i];
                const next = candles[i + 1];
                let time = current.time + intervalMs;

                // Fill gaps with flat candles using last close
                while (time < next.time - intervalMs / 2) {
                    filled.push({
                        symbol: current.symbol,
                        timeframe: current.timeframe,
                        time,
                        open: current.close,
                        high: current.close,
                        low: current.close,
                        close: current.close
                    });
                    time += intervalMs;
                }
            }
        }

        return filled;
    }

    /**
     * Clean up old data (older than 30 days)
     */
    async cleanupOldData(): Promise<number> {
        await this.init();
        if (!this.db) return 0;

        const cutoff = Date.now() - (MAX_AGE_DAYS * 24 * 60 * 60 * 1000);

        return new Promise((resolve, reject) => {
            const tx = this.db!.transaction(STORE_NAME, 'readwrite');
            const store = tx.objectStore(STORE_NAME);
            const index = store.index('by_time');
            const range = IDBKeyRange.upperBound(cutoff);
            const request = index.openCursor(range);

            let deleted = 0;

            request.onsuccess = () => {
                const cursor = request.result;
                if (cursor) {
                    cursor.delete();
                    deleted++;
                    cursor.continue();
                }
            };

            tx.oncomplete = () => {
                console.log(`[DataCache] Cleaned up ${deleted} old records`);
                resolve(deleted);
            };
            tx.onerror = () => reject(tx.error);
        });
    }

    private getIntervalMs(timeframe: string): number {
        const intervals: Record<string, number> = {
            '1m': 60 * 1000,
            '5m': 5 * 60 * 1000,
            '15m': 15 * 60 * 1000,
            '1h': 60 * 60 * 1000,
            '4h': 4 * 60 * 60 * 1000,
            '1d': 24 * 60 * 60 * 1000
        };
        return intervals[timeframe] || 60 * 1000;
    }
}

// Singleton instance
export const DataCache = new DataCacheService();
export type { CachedCandle, SyncMeta };
