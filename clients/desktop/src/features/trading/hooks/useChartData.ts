import { useState, useEffect } from 'react';
import type { OHLCData, Timeframe } from '../types';
import { DataCache } from '@/services/DataCache';
import { ExternalDataService } from '@/services/ExternalDataService';

export function useChartData(symbol: string, timeframe: Timeframe) {
  const [ohlc, setOhlc] = useState<OHLCData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function fetchOHLC() {
      try {
        setLoading(true);
        setError(null);

        // 1. Get cached data first
        const cachedCandles = await DataCache.getCandles(symbol, timeframe);

        // 2. Fetch from backend API
        const res = await fetch(`http://localhost:8080/ohlc?symbol=${symbol}&timeframe=${timeframe}&limit=1000`);
        let apiCandles: OHLCData[] = [];

        if (res.ok) {
          const data = await res.json();
          if (Array.isArray(data) && data.length > 0) {
            apiCandles = data.map((d: any) => ({
              time: d.time,
              open: d.open,
              high: d.high,
              low: d.low,
              close: d.close,
            }));
          }
        }

        // 3. Auto-fetch from external source if backend data is insufficient
        let externalCandles: OHLCData[] = [];
        const totalLocalCandles = cachedCandles.length + apiCandles.length;

        if (totalLocalCandles < 100 && ExternalDataService.isSupported(symbol)) {
          console.log(`[useChartData] Backend data insufficient (${totalLocalCandles}), fetching externally...`);
          const externalData = await ExternalDataService.fetchOHLC(symbol, timeframe, 500);
          externalCandles = externalData.map((d: any) => ({
            time: d.time,
            open: d.open,
            high: d.high,
            low: d.low,
            close: d.close,
          }));

          // Cache external data
          if (externalCandles.length > 0) {
            await DataCache.setCandles(symbol, timeframe, externalCandles);
          }
        }

        // 4. Merge all data sources
        const allCandles = [...cachedCandles, ...apiCandles, ...externalCandles];

        // Remove duplicates by timestamp
        const uniqueCandles = Array.from(
          new Map(allCandles.map(c => [c.time, c])).values()
        ).sort((a, b) => {
          const timeA = typeof a.time === 'string' ? parseInt(a.time) : (a.time as number);
          const timeB = typeof b.time === 'string' ? parseInt(b.time) : (b.time as number);
          return timeA - timeB;
        });

        if (!cancelled) {
          setOhlc(uniqueCandles);
          setLoading(false);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err as Error);
          setLoading(false);
        }
      }
    }

    fetchOHLC();

    return () => {
      cancelled = true;
    };
  }, [symbol, timeframe]);

  return { ohlc, loading, error };
}
