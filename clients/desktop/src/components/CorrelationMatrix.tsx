/**
 * Correlation Matrix Panel
 * Shows correlation coefficients between currency pairs
 * Heatmap visualization with color coding: red (negative) → white (neutral) → green (positive)
 */

import { useState, useEffect, useMemo } from 'react';
import { TrendingUp, Settings, X } from 'lucide-react';

// Default major currency pairs
const DEFAULT_SYMBOLS = [
    'EURUSD',
    'GBPUSD',
    'USDJPY',
    'USDCHF',
    'AUDUSD',
    'USDCAD',
    'NZDUSD',
    'EURGBP'
];

interface PriceData {
    symbol: string;
    returns: number[]; // Daily returns
}

// TODO: Backend endpoint needed - /api/market-data/returns or /api/correlation
// This would accept symbols[] and period (days) and return daily price returns.
// Currently the backend has /ticks and /ohlc endpoints which could be used to derive
// returns (fetch OHLC with timeframe=D1, compute daily close-to-close returns),
// but there is no dedicated endpoint for correlation data.
// Mock function to generate 30 days of price returns
function generateMockReturns(symbol: string, days: number = 30): number[] {
    const returns: number[] = [];

    // Seed based on symbol for consistency
    const seed = symbol.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
    let random = seed;

    const seededRandom = () => {
        random = (random * 9301 + 49297) % 233280;
        return random / 233280;
    };

    for (let i = 0; i < days; i++) {
        // Generate returns with some correlation patterns
        let baseReturn = (seededRandom() - 0.5) * 0.02; // -1% to +1%

        // Add correlation patterns
        if (symbol.includes('EUR')) {
            baseReturn += (seededRandom() - 0.5) * 0.01;
        }
        if (symbol.includes('USD')) {
            baseReturn += (seededRandom() - 0.5) * 0.008;
        }
        if (symbol.includes('JPY')) {
            baseReturn -= (seededRandom() - 0.5) * 0.005;
        }

        returns.push(baseReturn);
    }

    return returns;
}

// Calculate Pearson correlation coefficient
function calculateCorrelation(returns1: number[], returns2: number[]): number {
    const n = Math.min(returns1.length, returns2.length);
    if (n === 0) return 0;

    const mean1 = returns1.slice(0, n).reduce((a, b) => a + b, 0) / n;
    const mean2 = returns2.slice(0, n).reduce((a, b) => a + b, 0) / n;

    let numerator = 0;
    let sumSq1 = 0;
    let sumSq2 = 0;

    for (let i = 0; i < n; i++) {
        const diff1 = returns1[i] - mean1;
        const diff2 = returns2[i] - mean2;
        numerator += diff1 * diff2;
        sumSq1 += diff1 * diff1;
        sumSq2 += diff2 * diff2;
    }

    const denominator = Math.sqrt(sumSq1 * sumSq2);
    if (denominator === 0) return 0;

    return numerator / denominator;
}

// Get color for correlation value (-1 to +1)
function getCorrelationColor(correlation: number): string {
    // -1 (red) → 0 (white) → +1 (green)
    if (correlation < 0) {
        // Red gradient
        const intensity = Math.abs(correlation);
        const r = Math.round(239 * intensity + 30 * (1 - intensity)); // 239 (red) to 30 (dark)
        const g = Math.round(68 * intensity + 30 * (1 - intensity));
        const b = Math.round(68 * intensity + 30 * (1 - intensity));
        return `rgb(${r}, ${g}, ${b})`;
    } else if (correlation > 0) {
        // Green gradient
        const intensity = correlation;
        const r = Math.round(16 * intensity + 30 * (1 - intensity));
        const g = Math.round(185 * intensity + 30 * (1 - intensity)); // 185 (green) to 30 (dark)
        const b = Math.round(129 * intensity + 30 * (1 - intensity));
        return `rgb(${r}, ${g}, ${b})`;
    } else {
        // Neutral (white-ish)
        return 'rgb(60, 60, 60)';
    }
}

// Get text color for readability
function getTextColor(correlation: number): string {
    return Math.abs(correlation) > 0.5 ? '#ffffff' : '#d4d4d8';
}

export function CorrelationMatrix() {
    const [selectedSymbols, setSelectedSymbols] = useState<string[]>(DEFAULT_SYMBOLS);
    const [hoveredCell, setHoveredCell] = useState<{ row: string; col: string } | null>(null);
    const [showSettings, setShowSettings] = useState(false);
    const [availableSymbols] = useState<string[]>([
        ...DEFAULT_SYMBOLS,
        'EURJPY',
        'GBPJPY',
        'AUDJPY',
        'NZDJPY',
        'EURCHF',
        'GBPCHF',
        'AUDCAD',
        'AUDNZD'
    ]);

    // Generate price data for selected symbols
    const priceData = useMemo<PriceData[]>(() => {
        return selectedSymbols.map(symbol => ({
            symbol,
            returns: generateMockReturns(symbol, 30)
        }));
    }, [selectedSymbols]);

    // Calculate correlation matrix
    const correlationMatrix = useMemo(() => {
        const matrix: { [key: string]: { [key: string]: number } } = {};

        selectedSymbols.forEach(symbol1 => {
            matrix[symbol1] = {};
            selectedSymbols.forEach(symbol2 => {
                if (symbol1 === symbol2) {
                    matrix[symbol1][symbol2] = 1.0; // Perfect correlation with itself
                } else {
                    const data1 = priceData.find(d => d.symbol === symbol1);
                    const data2 = priceData.find(d => d.symbol === symbol2);
                    if (data1 && data2) {
                        matrix[symbol1][symbol2] = calculateCorrelation(data1.returns, data2.returns);
                    } else {
                        matrix[symbol1][symbol2] = 0;
                    }
                }
            });
        });

        return matrix;
    }, [selectedSymbols, priceData]);

    const handleCellClick = (symbol1: string, symbol2: string) => {
        if (symbol1 === symbol2) return;

        // Dispatch custom event to highlight both symbols on chart
        const event = new CustomEvent('highlightSymbolPair', {
            detail: { symbol1, symbol2 }
        });
        window.dispatchEvent(event);

        console.log(`[CorrelationMatrix] Highlighting pair: ${symbol1} <-> ${symbol2}`);
    };

    const toggleSymbol = (symbol: string) => {
        if (selectedSymbols.includes(symbol)) {
            if (selectedSymbols.length > 2) {
                setSelectedSymbols(selectedSymbols.filter(s => s !== symbol));
            }
        } else {
            setSelectedSymbols([...selectedSymbols, symbol]);
        }
    };

    return (
        <div className="h-full flex flex-col bg-[#1e1e1e] text-zinc-300 overflow-hidden">
            {/* Header */}
            <div className="bg-[#2d3436] border-b border-zinc-700 px-4 py-2 flex items-center justify-between flex-shrink-0">
                <div className="flex items-center gap-2">
                    <TrendingUp className="w-4 h-4 text-blue-400" />
                    <span className="text-sm font-semibold text-zinc-100">Correlation Matrix</span>
                    <span className="text-[10px] text-zinc-500 font-mono">(30-day)</span>
                </div>
                <button
                    onClick={() => setShowSettings(!showSettings)}
                    className={`p-1.5 rounded transition-colors ${showSettings ? 'bg-blue-500/20 text-blue-400' : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-700'
                        }`}
                    title="Symbol Settings"
                >
                    <Settings className="w-4 h-4" />
                </button>
            </div>

            {/* Settings Panel */}
            {showSettings && (
                <div className="bg-[#2d3436] border-b border-zinc-700 p-3 flex-shrink-0">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-xs font-semibold text-zinc-400">Select Symbols</span>
                        <button
                            onClick={() => setShowSettings(false)}
                            className="text-zinc-500 hover:text-zinc-300"
                        >
                            <X className="w-3 h-3" />
                        </button>
                    </div>
                    <div className="flex flex-wrap gap-1.5">
                        {availableSymbols.map(symbol => {
                            const isSelected = selectedSymbols.includes(symbol);
                            return (
                                <button
                                    key={symbol}
                                    onClick={() => toggleSymbol(symbol)}
                                    className={`px-2 py-1 rounded text-[10px] font-mono font-medium transition-colors ${isSelected
                                        ? 'bg-blue-500 text-white'
                                        : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
                                        }`}
                                    disabled={isSelected && selectedSymbols.length <= 2}
                                >
                                    {symbol}
                                </button>
                            );
                        })}
                    </div>
                    <p className="text-[9px] text-zinc-600 mt-2">
                        {selectedSymbols.length} symbols selected (minimum 2)
                    </p>
                </div>
            )}

            {/* Matrix Display */}
            <div className="flex-1 overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 p-4">
                <div className="inline-block min-w-full">
                    <table className="border-collapse">
                        <thead>
                            <tr>
                                <th className="sticky left-0 top-0 z-20 bg-[#2d3436] border-r border-b border-zinc-700 p-2"></th>
                                {selectedSymbols.map(symbol => (
                                    <th
                                        key={symbol}
                                        className="sticky top-0 z-10 bg-[#2d3436] border-b border-zinc-700 p-2 text-[10px] font-mono font-bold text-zinc-300 min-w-[60px]"
                                    >
                                        <div className="transform -rotate-45 origin-center whitespace-nowrap">
                                            {symbol}
                                        </div>
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody>
                            {selectedSymbols.map((rowSymbol, rowIdx) => (
                                <tr key={rowSymbol}>
                                    <td className="sticky left-0 z-10 bg-[#2d3436] border-r border-zinc-700 p-2 text-[10px] font-mono font-bold text-zinc-300 whitespace-nowrap">
                                        {rowSymbol}
                                    </td>
                                    {selectedSymbols.map((colSymbol, colIdx) => {
                                        const correlation = correlationMatrix[rowSymbol]?.[colSymbol] ?? 0;
                                        const bgColor = getCorrelationColor(correlation);
                                        const textColor = getTextColor(correlation);
                                        const isHovered =
                                            hoveredCell?.row === rowSymbol && hoveredCell?.col === colSymbol;
                                        const isDiagonal = rowSymbol === colSymbol;

                                        return (
                                            <td
                                                key={colSymbol}
                                                className={`border border-zinc-800 p-2 text-center text-[11px] font-mono font-bold transition-all duration-150 ${isDiagonal ? 'opacity-30' : 'cursor-pointer hover:ring-2 hover:ring-blue-400 hover:z-30'
                                                    } ${isHovered ? 'ring-2 ring-white z-30' : ''}`}
                                                style={{
                                                    backgroundColor: bgColor,
                                                    color: textColor
                                                }}
                                                onMouseEnter={() =>
                                                    setHoveredCell({ row: rowSymbol, col: colSymbol })
                                                }
                                                onMouseLeave={() => setHoveredCell(null)}
                                                onClick={() => handleCellClick(rowSymbol, colSymbol)}
                                                title={`${rowSymbol} vs ${colSymbol}: ${correlation.toFixed(3)}`}
                                            >
                                                {correlation.toFixed(2)}
                                            </td>
                                        );
                                    })}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                {/* Legend */}
                <div className="mt-6 bg-zinc-900 border border-zinc-800 rounded p-3 max-w-md">
                    <h4 className="text-xs font-semibold text-zinc-400 mb-2">Correlation Scale</h4>
                    <div className="flex items-center gap-2 mb-2">
                        <div
                            className="w-8 h-4 rounded"
                            style={{ backgroundColor: getCorrelationColor(-1) }}
                        />
                        <span className="text-[10px] text-zinc-500 flex-1">-1.0 (Perfect Negative)</span>
                    </div>
                    <div className="flex items-center gap-2 mb-2">
                        <div
                            className="w-8 h-4 rounded"
                            style={{ backgroundColor: getCorrelationColor(0) }}
                        />
                        <span className="text-[10px] text-zinc-500 flex-1">0.0 (No Correlation)</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div
                            className="w-8 h-4 rounded"
                            style={{ backgroundColor: getCorrelationColor(1) }}
                        />
                        <span className="text-[10px] text-zinc-500 flex-1">+1.0 (Perfect Positive)</span>
                    </div>
                    <p className="text-[9px] text-zinc-600 mt-3">
                        Click a cell to highlight the symbol pair on the chart. Hover for exact values.
                    </p>
                </div>
            </div>
        </div>
    );
}
