import { createChart, LineSeries } from 'lightweight-charts';
import type { IChartApi, ISeriesApi, Time } from 'lightweight-charts';
import { useEffect, useRef } from 'react';

interface Tick {
    symbol: string;
    bid: number;
    ask: number;
    spread?: number;
    timestamp: number;
}

interface TickChartProps {
    symbol: string;
    currentTick: Tick | undefined;
}

export function TickChart({ symbol, currentTick }: TickChartProps) {
    const chartContainerRef = useRef<HTMLDivElement>(null);
    const chartRef = useRef<IChartApi | null>(null);
    const bidSeriesRef = useRef<ISeriesApi<"Line"> | null>(null);
    const askSeriesRef = useRef<ISeriesApi<"Line"> | null>(null);

    // Initialize Chart
    useEffect(() => {
        if (!chartContainerRef.current) return;

        const chart = createChart(chartContainerRef.current, {
            layout: {
                background: { color: '#09090b' },
                textColor: '#71717a',
            },
            grid: {
                vertLines: { color: '#27272a' },
                horzLines: { color: '#27272a' },
            },
            width: chartContainerRef.current.clientWidth,
            height: 300,
            timeScale: {
                timeVisible: true,
                secondsVisible: true,
            },
        });

        const bidSeries = chart.addSeries(LineSeries, {
            color: '#10b981', // Emerald
            lineWidth: 2,
            priceLineVisible: false,
        });

        const askSeries = chart.addSeries(LineSeries, {
            color: '#ef4444', // Red
            lineWidth: 2,
            priceLineVisible: false,
        });

        chartRef.current = chart;
        bidSeriesRef.current = bidSeries;
        askSeriesRef.current = askSeries;

        // Fetch initial history
        fetch(`http://localhost:8080/api/ticks/history?symbol=${symbol}&limit=100`)
            .then(res => res.json())
            .then((data: any[]) => {
                if (!data || !chartRef.current) return;

                // Map data to lightweight-charts format
                // Note: Lightweight charts expects ascending time. 
                // TickStore returns most recent last (ascending).
                const bidData = data.map(t => ({
                    time: (new Date(t.timestamp).getTime() / 1000) as Time,
                    value: t.bid,
                }));
                const askData = data.map(t => ({
                    time: (new Date(t.timestamp).getTime() / 1000) as Time,
                    value: t.ask,
                }));

                bidSeries.setData(bidData);
                askSeries.setData(askData);
                chart.timeScale().fitContent();
            })
            .catch(err => console.error("Failed to load tick history", err));

        const handleResize = () => {
            if (chartContainerRef.current) {
                chart.applyOptions({ width: chartContainerRef.current.clientWidth });
            }
        };

        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
            chart.remove();
        };
    }, [symbol]);

    // Update Real-time
    useEffect(() => {
        if (!currentTick || !bidSeriesRef.current || !askSeriesRef.current) return;
        if (currentTick.symbol !== symbol) return;

        const time = (Date.now() / 1000) as Time;

        bidSeriesRef.current.update({
            time: time,
            value: currentTick.bid,
        });

        askSeriesRef.current.update({
            time: time,
            value: currentTick.ask,
        });

    }, [currentTick, symbol]);

    return (
        <div className="flex flex-col h-full bg-[#09090b] border border-zinc-800 rounded-lg overflow-hidden">
            <div className="flex items-center justify-between px-3 py-2 border-b border-zinc-800 bg-zinc-900/50">
                <span className="text-xs font-bold text-zinc-300">Tick Chart: {symbol}</span>
                <div className="flex gap-4 text-[10px] font-mono">
                    <span className="text-emerald-400">Bid: {currentTick?.bid}</span>
                    <span className="text-red-400">Ask: {currentTick?.ask}</span>
                </div>
            </div>
            <div ref={chartContainerRef} className="flex-1 w-full relative" />
        </div>
    );
}
