/**
 * Chart Analysis Service
 * API client for AI-powered forex chart image analysis
 */

export interface DetectionResult {
    status: string;
    detected_pairs: string[];
    confidence: number;
    analysis_id: string;
    timestamp: string;
    message?: string;
}

export interface LivePrice {
    pair: string;
    current_price: string;
    bid: string;
    ask: string;
    spread: string;
    timestamp: string;
}

export interface OHLCCandle {
    time: string;
    open: string;
    high: string;
    low: string;
    close: string;
    volume?: number;
}

export interface OHLCData {
    pair: string;
    timeframe: string;
    candles: OHLCCandle[];
    count: number;
    timestamp: string;
}

export type Timeframe = '1m' | '5m' | '15m' | '1h' | '4h' | '1d';

const API_BASE_URL = 'http://localhost:7999'; // Backend URL

export class ChartAnalysisService {
    /**
     * Upload chart image for AI analysis
     * Returns detected forex pairs
     */
    async uploadChart(imageFile: File): Promise<DetectionResult> {
        const formData = new FormData();
        formData.append('image', imageFile);

        const response = await fetch(`${API_BASE_URL}/api/analysis/upload-chart`, {
            method: 'POST',
            body: formData,
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to analyze chart image');
        }

        return response.json();
    }

    /**
     * Get current live price for a symbol
     */
    async getLivePrice(symbol: string): Promise<LivePrice> {
        const response = await fetch(`${API_BASE_URL}/api/analysis/pairs/${symbol}/live`);

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || `Failed to fetch live price for ${symbol}`);
        }

        return response.json();
    }

    /**
     * Get OHLC candlestick data for a symbol
     */
    async getOHLCData(
        symbol: string,
        timeframe: Timeframe = '5m',
        limit: number = 100
    ): Promise<OHLCData> {
        const url = `${API_BASE_URL}/api/analysis/pairs/${symbol}/ohlc?timeframe=${timeframe}&limit=${limit}`;
        const response = await fetch(url);

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || `Failed to fetch OHLC data for ${symbol}`);
        }

        return response.json();
    }
}

// Singleton instance
export const chartAnalysisService = new ChartAnalysisService();
