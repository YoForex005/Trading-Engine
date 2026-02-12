'use client';

import React, { useState, useCallback } from 'react';
import ChartWindow from './ChartWindow';
import { Timeframe } from '@/config/api';

export type ChartLayout = '1x1' | '2x1' | '2x2' | '3x2';

interface ChartInstance {
    id: string;
    symbol: string;
    timeframe: Timeframe;
}

interface ChartGridProps {
    layout: ChartLayout;
    initialSymbol?: string;
    onSymbolRequest?: () => void;
}

export default function ChartGrid({ layout, initialSymbol = 'EURUSD', onSymbolRequest }: ChartGridProps) {
    const [charts, setCharts] = useState<ChartInstance[]>([
        { id: '1', symbol: initialSymbol, timeframe: 'M5' }
    ]);
    const [activeChartId, setActiveChartId] = useState<string>('1');

    // Get grid dimensions based on layout
    const getGridClass = useCallback(() => {
        switch (layout) {
            case '1x1':
                return 'grid-cols-1 grid-rows-1';
            case '2x1':
                return 'grid-cols-2 grid-rows-1';
            case '2x2':
                return 'grid-cols-2 grid-rows-2';
            case '3x2':
                return 'grid-cols-3 grid-rows-2';
            default:
                return 'grid-cols-1 grid-rows-1';
        }
    }, [layout]);

    // Get max charts based on layout
    const getMaxCharts = useCallback(() => {
        switch (layout) {
            case '1x1': return 1;
            case '2x1': return 2;
            case '2x2': return 4;
            case '3x2': return 6;
            default: return 1;
        }
    }, [layout]);

    // Adjust charts when layout changes
    React.useEffect(() => {
        const maxCharts = getMaxCharts();
        if (charts.length > maxCharts) {
            // Remove excess charts
            setCharts(prev => prev.slice(0, maxCharts));
        } else if (charts.length < maxCharts && layout !== '1x1') {
            // Add placeholder charts
            setCharts(prev => {
                const newCharts = [...prev];
                while (newCharts.length < maxCharts) {
                    newCharts.push({
                        id: `${Date.now()}-${newCharts.length}`,
                        symbol: initialSymbol,
                        timeframe: 'M5'
                    });
                }
                return newCharts;
            });
        }
    }, [layout, getMaxCharts, charts.length, initialSymbol]);

    // Open symbol in active/focused chart
    const openSymbolInActiveChart = useCallback((symbol: string) => {
        setCharts(prev => prev.map(chart =>
            chart.id === activeChartId
                ? { ...chart, symbol }
                : chart
        ));
    }, [activeChartId]);

    // Open symbol in new chart window
    const openSymbolInNewChart = useCallback((symbol: string) => {
        const maxCharts = getMaxCharts();
        if (charts.length < maxCharts) {
            const newChart: ChartInstance = {
                id: `${Date.now()}`,
                symbol,
                timeframe: 'M5'
            };
            setCharts(prev => [...prev, newChart]);
            setActiveChartId(newChart.id);
        } else {
            // Replace the last chart if at max capacity
            setCharts(prev => {
                const updated = [...prev];
                updated[updated.length - 1] = {
                    id: `${Date.now()}`,
                    symbol,
                    timeframe: 'M5'
                };
                return updated;
            });
        }
    }, [charts.length, getMaxCharts]);

    // Expose methods to parent via custom event
    React.useEffect(() => {
        const handleOpenSymbol = (e: CustomEvent) => {
            if (e.detail.newWindow) {
                openSymbolInNewChart(e.detail.symbol);
            } else {
                openSymbolInActiveChart(e.detail.symbol);
            }
        };

        window.addEventListener('chart:openSymbol' as any, handleOpenSymbol as any);
        return () => {
            window.removeEventListener('chart:openSymbol' as any, handleOpenSymbol as any);
        };
    }, [openSymbolInActiveChart, openSymbolInNewChart]);

    const handleCloseChart = useCallback((chartId: string) => {
        if (charts.length > 1) {
            setCharts(prev => prev.filter(c => c.id !== chartId));
            if (activeChartId === chartId && charts.length > 0) {
                setActiveChartId(charts[0].id);
            }
        }
    }, [charts, activeChartId]);

    return (
        <div className={`grid ${getGridClass()} gap-1 h-full w-full bg-[#121316] p-1`}>
            {charts.map((chart) => (
                <ChartWindow
                    key={chart.id}
                    symbol={chart.symbol}
                    timeframe={chart.timeframe}
                    isActive={chart.id === activeChartId}
                    onClose={charts.length > 1 ? () => handleCloseChart(chart.id) : undefined}
                    onFocus={() => setActiveChartId(chart.id)}
                />
            ))}
        </div>
    );
}

// Export helper function to trigger chart opening from outside
export function openSymbolInChart(symbol: string, newWindow: boolean = false) {
    const event = new CustomEvent('chart:openSymbol', {
        detail: { symbol, newWindow }
    });
    window.dispatchEvent(event);
}
