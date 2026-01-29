import React, { useState, useEffect } from 'react';
import { useMarketDataStore } from '../store/useMarketDataStore';
import { marketApi } from '../services/api';

interface MarketWatchProps {
    onSelectSymbol: (symbol: string) => void;
    selectedSymbol: string;
    watchlist?: string[];
}

export const MarketWatch: React.FC<MarketWatchProps> = ({ onSelectSymbol, selectedSymbol, watchlist }) => {
    const symbolData = useMarketDataStore(state => state.symbolData);
    const [symbols, setSymbols] = useState<string[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // Fetch symbols from API on mount
    useEffect(() => {
        const fetchSymbols = async () => {
            try {
                setIsLoading(true);
                setError(null);
                const symbolsData = await marketApi.getSymbols();

                // Extract symbol names from the API response
                const symbolNames = symbolsData
                    .filter(s => s.enabled !== false) // Filter out disabled symbols
                    .map(s => s.symbol);

                setSymbols(symbolNames);
            } catch (err) {
                console.error('[MarketWatch] Failed to fetch symbols:', err);
                setError('Failed to load symbols');
                setSymbols([]); // Fallback to empty list
            } finally {
                setIsLoading(false);
            }
        };

        // Use provided watchlist or fetch from API
        if (watchlist && watchlist.length > 0) {
            setSymbols(watchlist);
            setIsLoading(false);
        } else {
            fetchSymbols();
        }
    }, [watchlist]);

    // Use the fetched symbols or the provided watchlist
    const symbolsToShow = symbols;

    return (
        <div className="market-watch bg-[#131722] border-r border-[#363c4e] flex flex-col w-64 h-full">
            <div className="p-2 border-b border-[#363c4e] text-[#d1d4dc] font-semibold text-xs uppercase flex justify-between items-center">
                <span>Market Watch</span>
                <span className="text-[10px] text-gray-500">{new Date().toLocaleTimeString()}</span>
            </div>


            <div className="flex-1 overflow-y-auto custom-scrollbar">
                {isLoading ? (
                    <div className="flex items-center justify-center h-32 text-gray-400 text-xs">
                        <div className="flex flex-col items-center gap-2">
                            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-[#2962ff]"></div>
                            <span>Loading symbols...</span>
                        </div>
                    </div>
                ) : error ? (
                    <div className="flex items-center justify-center h-32 text-[#ef5350] text-xs px-4 text-center">
                        <div className="flex flex-col items-center gap-2">
                            <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            <span>{error}</span>
                        </div>
                    </div>
                ) : symbolsToShow.length === 0 ? (
                    <div className="flex items-center justify-center h-32 text-gray-500 text-xs px-4 text-center">
                        <div className="flex flex-col items-center gap-2">
                            <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 0 00-2 2v7m16 0v5a2 0 01-2 2H6a2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                            </svg>
                            <span>No symbols available</span>
                        </div>
                    </div>
                ) : (
                    <table className="w-full text-xs collapse">
                        <thead className="sticky top-0 bg-[#131722] z-10">
                            <tr className="text-gray-500 font-normal">
                                <th className="text-left p-2 pl-3">Symbol</th>
                                <th className="text-right p-2">Bid</th>
                                <th className="text-right p-2 pr-3">Ask</th>
                            </tr>
                        </thead>
                        <tbody>
                            {symbolsToShow.map(symbol => {
                                const data = symbolData[symbol];
                                const quote = data?.currentTick;

                                // Determine color based on price change logic (placeholder or real if available)
                                // Ideally compare with previous tick
                                const prev = data?.previousTick;
                                const bidColor = quote && prev && quote.bid > prev.bid ? 'text-[#26a69a]' : (quote && prev && quote.bid < prev.bid ? 'text-[#ef5350]' : 'text-[#d1d4dc]');

                                return (
                                    <tr
                                        key={symbol}
                                        onClick={() => onSelectSymbol(symbol)}
                                        className={`cursor-pointer hover:bg-[#2a2e39] transition-colors ${selectedSymbol === symbol ? 'bg-[#2a2e39]' : ''}`}
                                    >
                                        <td className="p-2 pl-3 py-1.5 flex items-center gap-1.5">
                                            <span className={`w-1.5 h-1.5 rounded-full ${bidColor === 'text-[#26a69a]' ? 'bg-[#26a69a]' : bidColor === 'text-[#ef5350]' ? 'bg-[#ef5350]' : 'bg-gray-500'}`}></span>
                                            <span className="text-[#d1d4dc] font-medium">{symbol}</span>
                                        </td>
                                        <td className={`text-right p-2 py-1.5 ${bidColor}`}>
                                            {quote ? quote.bid.toFixed(5) : '---'}
                                        </td>
                                        <td className={`text-right p-2 py-1.5 text-[#d1d4dc]`}>
                                            {quote ? quote.ask.toFixed(5) : '---'}
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                )}
            </div>
        </div>
    );
};
