/**
 * ExternalDataService - Fetch historical OHLC from public APIs (Binance)
 * Automatically fetches data without authentication
 */

interface ExternalCandle {
    time: number;
    open: number;
    high: number;
    low: number;
    close: number;
    volume: number;
}

// Symbol mapping: Platform -> Binance
const BINANCE_SYMBOL_MAP: Record<string, string> = {
    'BTCUSD': 'BTCUSDT',
    'ETHUSD': 'ETHUSDT',
    'BNBUSD': 'BNBUSDT',
    'XRPUSD': 'XRPUSDT',
    'ADAUSD': 'ADAUSDT',
    'SOLUSD': 'SOLUSDT',
    'DOGEUSD': 'DOGEUSDT',
    'DOTUSD': 'DOTUSDT',
    'LTCUSD': 'LTCUSDT',
    'LINKUSD': 'LINKUSDT',
};

// Timeframe mapping: Platform -> Binance interval
const BINANCE_INTERVAL_MAP: Record<string, string> = {
    '1m': '1m',
    '5m': '5m',
    '15m': '15m',
    '1h': '1h',
    '4h': '4h',
    '1d': '1d',
};

class ExternalDataServiceClass {
    private baseUrl = 'https://api.binance.com/api/v3';

    /**
     * Check if symbol is supported by Binance
     */
    isSupported(symbol: string): boolean {
        return symbol in BINANCE_SYMBOL_MAP;
    }

    /**
     * Fetch historical OHLC from Binance
     */
    async fetchOHLC(symbol: string, timeframe: string, limit: number = 500): Promise<ExternalCandle[]> {
        const binanceSymbol = BINANCE_SYMBOL_MAP[symbol];
        const binanceInterval = BINANCE_INTERVAL_MAP[timeframe] || '1m';

        if (!binanceSymbol) {
            console.log(`[ExternalData] Symbol ${symbol} not supported by Binance`);
            return [];
        }

        try {
            const url = `${this.baseUrl}/klines?symbol=${binanceSymbol}&interval=${binanceInterval}&limit=${limit}`;
            console.log(`[ExternalData] Fetching from Binance: ${url}`);

            const res = await fetch(url);
            if (!res.ok) {
                console.error(`[ExternalData] Binance API error: ${res.status}`);
                return [];
            }

            const data = await res.json();

            // Binance kline format: [openTime, open, high, low, close, volume, closeTime, ...]
            const candles: ExternalCandle[] = data.map((k: any[]) => ({
                time: Math.floor(k[0] / 1000), // Convert ms to seconds
                open: parseFloat(k[1]),
                high: parseFloat(k[2]),
                low: parseFloat(k[3]),
                close: parseFloat(k[4]),
                volume: parseFloat(k[5]),
            }));

            console.log(`[ExternalData] Fetched ${candles.length} candles from Binance for ${symbol}`);
            return candles;

        } catch (err) {
            console.error('[ExternalData] Fetch error:', err);
            return [];
        }
    }

    /**
     * Fetch historical data with time range (for filling gaps)
     */
    async fetchOHLCRange(symbol: string, timeframe: string, startTime: number, endTime: number): Promise<ExternalCandle[]> {
        const binanceSymbol = BINANCE_SYMBOL_MAP[symbol];
        const binanceInterval = BINANCE_INTERVAL_MAP[timeframe] || '1m';

        if (!binanceSymbol) {
            return [];
        }

        try {
            const url = `${this.baseUrl}/klines?symbol=${binanceSymbol}&interval=${binanceInterval}&startTime=${startTime * 1000}&endTime=${endTime * 1000}&limit=1000`;
            const res = await fetch(url);

            if (!res.ok) return [];

            const data = await res.json();

            return data.map((k: any[]) => ({
                time: Math.floor(k[0] / 1000),
                open: parseFloat(k[1]),
                high: parseFloat(k[2]),
                low: parseFloat(k[3]),
                close: parseFloat(k[4]),
                volume: parseFloat(k[5]),
            }));

        } catch (err) {
            console.error('[ExternalData] Range fetch error:', err);
            return [];
        }
    }
}

export const ExternalDataService = new ExternalDataServiceClass();
export type { ExternalCandle };
