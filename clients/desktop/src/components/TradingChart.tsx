import { useEffect, useRef, useState } from 'react';
import {
    createChart,
    ColorType,
    CrosshairMode,
    CandlestickSeries,
    LineSeries,
    BarSeries,
    AreaSeries,
} from 'lightweight-charts';
import type { IChartApi, ISeriesApi, Time } from 'lightweight-charts';
import { Maximize2, Minimize2 } from 'lucide-react';
import { DrawingTools, type DrawingType } from './TradingChart/DrawingTools';
import { DrawingOverlay } from './TradingChart/DrawingOverlay';
import type { Drawing } from './TradingChart/types';
import { DataCache, type CachedCandle } from '../services/DataCache';
import { ExternalDataService } from '../services/ExternalDataService';
import { useIndicators } from '../hooks/useIndicators';
import IndicatorManager from './IndicatorManager';
import { IndicatorPane } from './TradingChart/IndicatorPane';
import { IndicatorStorage } from '../services/IndicatorStorage';
import type { IndicatorType, IndicatorParams } from '../indicators/core/IndicatorEngine';

export type ChartType = 'candlestick' | 'heikinAshi' | 'bar' | 'line' | 'area';
export type Timeframe = '1m' | '5m' | '15m' | '1h' | '4h' | '1d';

interface ChartProps {
    symbol: string;
    currentPrice?: { bid: number; ask: number };
    chartType?: ChartType;
    timeframe?: Timeframe;
    positions?: any[]; // Array of open positions
    onClosePosition?: (id: number) => void;
    onModifyPosition?: (id: number, sl: number, tp: number) => void;
    onPlacePendingOrder?: (price: number, type: 'LIMIT' | 'STOP') => void;
    oneClickEnabled?: boolean;
}

interface OHLC {
    time: Time;
    open: number;
    high: number;
    low: number;
    close: number;
}

// ... (previous imports)

export function TradingChart({
    symbol,
    currentPrice,
    chartType = 'candlestick',
    timeframe = '1m',
    positions = [],
    onClosePosition,
    onModifyPosition,
    onPlacePendingOrder,
    oneClickEnabled
}: ChartProps) {
    const chartContainerRef = useRef<HTMLDivElement>(null);
    const chartRef = useRef<IChartApi | null>(null);
    const seriesRef = useRef<ISeriesApi<any> | null>(null);
    const candlesRef = useRef<OHLC[]>([]);
    const [isChartReady, setIsChartReady] = useState(false);

    // Drawing State
    const [activeTool, setActiveTool] = useState<DrawingType>('cursor');
    const [drawings, setDrawings] = useState<Drawing[]>([]);

    // Indicator State
    const [showIndicatorManager, setShowIndicatorManager] = useState(false);
    const indicatorSeriesRef = useRef<Map<string, ISeriesApi<'Line'>>>(new Map());

    // Convert candleRef OHLC to indicator OHLC format (Time -> number)
    const indicatorOHLCData = candlesRef.current.map((c) => ({
        time: typeof c.time === 'string' ? parseInt(c.time) : (c.time as number),
        open: c.open,
        high: c.high,
        low: c.low,
        close: c.close,
    }));

    // Initialize indicators hook
    const {
        indicators,
        overlayIndicators,
        paneGroups,
        addIndicator,
    } = useIndicators({
        ohlcData: indicatorOHLCData,
        autoCalculate: true,
    });

    // Load saved indicators on symbol/timeframe change
    useEffect(() => {
        const savedIndicators = IndicatorStorage.load(symbol, timeframe);
        for (const savedInd of savedIndicators) {
            if (savedInd.type && savedInd.params) {
                addIndicator(savedInd.type as IndicatorType, savedInd.params as IndicatorParams, {
                    color: savedInd.color,
                    lineWidth: savedInd.lineWidth,
                    visible: savedInd.visible,
                    paneIndex: savedInd.paneIndex,
                });
            }
        }
    }, [symbol, timeframe]);

    // Save indicators whenever they change
    useEffect(() => {
        IndicatorStorage.save(symbol, timeframe, indicators);
    }, [indicators, symbol, timeframe]);

    // Fetch drawings on load
    useEffect(() => {
        // Need accountId, assuming 1 for now or passed from props (TODO: Pass accountId to TradingChart)
        const accountId = 1;
        fetch(`http://localhost:8080/api/drawings?accountId=${accountId}&symbol=${symbol}`)
            .then(res => res.json())
            .then(data => setDrawings(data || []))
            .catch(err => console.error('Failed to fetch drawings:', err));
    }, [symbol]);

    const handleUpdateDrawing = async (drawing: Drawing) => {
        // Optimistic update
        setDrawings(prev => {
            const exists = prev.find(d => d.id === drawing.id);
            if (exists) {
                return prev.map(d => d.id === drawing.id ? drawing : d);
            }
            return [...prev, drawing];
        });

        // Save to backend
        try {
            await fetch('http://localhost:8080/api/drawings/save', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(drawing)
            });
        } catch (err) {
            console.error('Failed to save drawing:', err);
        }
    };

    const handleDeleteDrawing = async (id: string) => {
        setDrawings(prev => prev.filter(d => d.id !== id));
        try {
            // Need accountId
            await fetch(`http://localhost:8080/api/drawings/delete?id=${id}&accountId=1`, { method: 'POST' });
        } catch (err) {
            console.error('Failed to delete drawing:', err);
        }
    };

    // State for overlays
    const [overlayPositions, setOverlayPositions] = useState<any[]>([]);

    // Initialize chart ONCE
    useEffect(() => {
        if (!chartContainerRef.current) return;

        try {
            const chart = createChart(chartContainerRef.current, {
                layout: {
                    background: { type: ColorType.Solid, color: 'transparent' },
                    textColor: '#71717a',
                    attributionLogo: false,
                },
                grid: {
                    vertLines: { color: '#27272a' },
                    horzLines: { color: '#27272a' },
                },
                crosshair: {
                    mode: CrosshairMode.Normal,
                    vertLine: { color: '#525252', width: 1, style: 2, labelBackgroundColor: '#18181b' },
                    horzLine: { color: '#525252', width: 1, style: 2, labelBackgroundColor: '#18181b' },
                },
                rightPriceScale: {
                    borderColor: '#27272a',
                    scaleMargins: { top: 0.1, bottom: 0.2 },
                },
                timeScale: {
                    borderColor: '#27272a',
                    timeVisible: true,
                    secondsVisible: false,
                },
                handleScroll: { mouseWheel: true, pressedMouseMove: true, horzTouchDrag: true, vertTouchDrag: false },
                handleScale: { axisPressedMouseMove: true, mouseWheel: true, pinch: true },
            });

            chartRef.current = chart;

            const handleResize = () => {
                if (chartContainerRef.current && chartRef.current) {
                    chartRef.current.applyOptions({
                        width: chartContainerRef.current.clientWidth,
                        height: chartContainerRef.current.clientHeight,
                    });
                }
            };

            window.addEventListener('resize', handleResize);
            handleResize();
            setIsChartReady(true);

            return () => {
                window.removeEventListener('resize', handleResize);
                setIsChartReady(false);
                chart.remove();
                chartRef.current = null;
                seriesRef.current = null;
            };
        } catch (err) {
            console.error('Failed to initialize chart:', err);
        }
    }, []);

    // Create/update series when chart is ready or chartType changes
    useEffect(() => {
        if (!isChartReady || !chartRef.current) return;

        try {
            if (seriesRef.current) {
                try { chartRef.current.removeSeries(seriesRef.current); } catch (e) { }
                seriesRef.current = null;
            }

            let series: any;
            const chart = chartRef.current;

            switch (chartType) {
                case 'candlestick':
                case 'heikinAshi':
                    series = chart.addSeries(CandlestickSeries, {
                        upColor: '#10b981', downColor: '#ef4444',
                        borderUpColor: '#10b981', borderDownColor: '#ef4444',
                        wickUpColor: '#10b981', wickDownColor: '#ef4444',
                    });
                    break;
                case 'bar':
                    series = chart.addSeries(BarSeries, { upColor: '#10b981', downColor: '#ef4444' });
                    break;
                case 'line':
                    series = chart.addSeries(LineSeries, { color: '#10b981', lineWidth: 2 });
                    break;
                case 'area':
                    series = chart.addSeries(AreaSeries, {
                        lineColor: '#10b981', topColor: 'rgba(16, 185, 129, 0.4)',
                        bottomColor: 'rgba(16, 185, 129, 0.0)', lineWidth: 2,
                    });
                    break;
                default:
                    series = chart.addSeries(CandlestickSeries, {
                        upColor: '#10b981', downColor: '#ef4444',
                        borderUpColor: '#10b981', borderDownColor: '#ef4444',
                        wickUpColor: '#10b981', wickDownColor: '#ef4444',
                    });
            }

            seriesRef.current = series;

            if (candlesRef.current.length > 0) {
                const formattedData = formatDataForSeries(candlesRef.current, chartType);
                series.setData(formattedData);
            }
        } catch (err) {
            console.error('Error creating chart series:', err);
        }
    }, [isChartReady, chartType]);

    // Fetch historical OHLC data when symbol or timeframe changes
    useEffect(() => {
        const fetchHistory = async () => {
            if (!seriesRef.current) return;

            try {
                // 1. Get cached data first
                const cachedCandles = await DataCache.getCandles(symbol, timeframe);

                // 2. Fetch from backend API
                const res = await fetch(`http://localhost:8080/ohlc?symbol=${symbol}&timeframe=${timeframe}&limit=1000`);
                let apiCandles: OHLC[] = [];

                if (res.ok) {
                    const data = await res.json();
                    if (Array.isArray(data) && data.length > 0) {
                        apiCandles = data.map((d: any) => ({
                            time: d.time as Time,
                            open: d.open, high: d.high, low: d.low, close: d.close,
                        }));
                    }
                }

                // 3. AUTO-FETCH FROM BINANCE if backend data is insufficient
                let externalCandles: OHLC[] = [];
                const totalLocalCandles = cachedCandles.length + apiCandles.length;

                if (totalLocalCandles < 100 && ExternalDataService.isSupported(symbol)) {
                    console.log(`[TradingChart] Backend data insufficient (${totalLocalCandles}), fetching from Binance...`);
                    const binanceData = await ExternalDataService.fetchOHLC(symbol, timeframe, 500);
                    externalCandles = binanceData.map(c => ({
                        time: c.time as Time,
                        open: c.open, high: c.high, low: c.low, close: c.close,
                    }));
                }

                // 4. Merge all sources: External (oldest) + Cache + Backend (newest)
                const allCandlesMap = new Map<number, OHLC>();

                // External data first (lowest priority)
                for (const c of externalCandles) {
                    allCandlesMap.set(c.time as number, c);
                }
                // Then cached data
                for (const c of cachedCandles) {
                    allCandlesMap.set(c.time, { time: c.time as Time, open: c.open, high: c.high, low: c.low, close: c.close });
                }
                // Backend API data takes highest priority
                for (const c of apiCandles) {
                    allCandlesMap.set(c.time as number, c);
                }

                let mergedCandles = Array.from(allCandlesMap.values()).sort((a, b) => (a.time as number) - (b.time as number));

                // 5. Fill gaps using DataCache logic
                if (mergedCandles.length > 1) {
                    const filledCandles = DataCache.fillGaps(
                        mergedCandles.map(c => ({
                            symbol, timeframe, time: c.time as number,
                            open: c.open, high: c.high, low: c.low, close: c.close
                        })),
                        timeframe
                    );
                    mergedCandles = filledCandles.map((c: CachedCandle) => ({
                        time: c.time as Time, open: c.open, high: c.high, low: c.low, close: c.close
                    }));
                }

                // 6. Store merged data to cache
                if (mergedCandles.length > 0) {
                    await DataCache.storeCandles(
                        symbol,
                        timeframe,
                        mergedCandles.map(c => ({
                            symbol, timeframe, time: c.time as number,
                            open: c.open, high: c.high, low: c.low, close: c.close
                        }))
                    );
                }

                // 7. Update chart
                candlesRef.current = mergedCandles;
                const formattedData = formatDataForSeries(mergedCandles, chartType);
                seriesRef.current.setData(formattedData);

                console.log(`[TradingChart] Loaded ${mergedCandles.length} candles for ${symbol}/${timeframe} (${cachedCandles.length} cached, ${apiCandles.length} backend, ${externalCandles.length} Binance)`);

            } catch (err) {
                console.error('Error fetching historical data:', err);
                candlesRef.current = [];
            }
        };

        fetchHistory();
    }, [symbol, timeframe, chartType]);

    // Update candles with new price
    useEffect(() => {
        if (!currentPrice || !seriesRef.current) return;

        const price = (currentPrice.bid + currentPrice.ask) / 2;
        const now = Math.floor(Date.now() / 1000) as Time;
        const tfSeconds = getTimeframeSeconds(timeframe);
        const candleTime = (Math.floor((now as number) / tfSeconds) * tfSeconds) as Time;

        if (candlesRef.current.length === 0) {
            candlesRef.current.push({ time: candleTime, open: price, high: price, low: price, close: price });
        } else {
            const lastCandle = candlesRef.current[candlesRef.current.length - 1];
            if (lastCandle.time === candleTime) {
                lastCandle.close = price;
                lastCandle.high = Math.max(lastCandle.high, price);
                lastCandle.low = Math.min(lastCandle.low, price);
                seriesRef.current.update(lastCandle);
            } else {
                const newCandle = { time: candleTime, open: price, high: price, low: price, close: price };
                candlesRef.current.push(newCandle);
                seriesRef.current.update(newCandle);
            }
        }
    }, [currentPrice, timeframe]);

    // Update overlay positions on scroll/zoom
    useEffect(() => {
        if (!chartRef.current || !seriesRef.current) return;

        const updateOverlays = () => {
            if (!seriesRef.current) return;
            const newOverlays = positions
                .filter(p => p.symbol === symbol)
                .map(p => {
                    const yEntry = seriesRef.current!.priceToCoordinate(p.openPrice);
                    const ySL = p.sl ? seriesRef.current!.priceToCoordinate(p.sl) : null;
                    const yTP = p.tp ? seriesRef.current!.priceToCoordinate(p.tp) : null;
                    return { ...p, yEntry, ySL, yTP };
                });
            setOverlayPositions(newOverlays);
        };

        const timeScale = chartRef.current.timeScale();
        timeScale.subscribeVisibleTimeRangeChange(updateOverlays);
        // Also update immediately and on resize
        updateOverlays();

        // Hack to update on Y-axis scale changes (monitor visual updates)
        const interval = setInterval(updateOverlays, 100); // 10fps check for smooth enough updates during drag

        return () => {
            timeScale.unsubscribeVisibleTimeRangeChange(updateOverlays);
            clearInterval(interval);
        };
    }, [positions, symbol, isChartReady]);

    // Dragging Logic
    const [draggingState, setDraggingState] = useState<{ id: number; type: 'SL' | 'TP'; startPrice: number } | null>(null);
    const [dragPrice, setDragPrice] = useState<number | null>(null);

    const handleDragStart = (id: number, type: 'SL' | 'TP', currentPrice: number) => {
        setDraggingState({ id, type, startPrice: currentPrice });
        setDragPrice(currentPrice);
        // Disable chart scroll/interaction while dragging overlays
        if (chartRef.current) {
            chartRef.current.applyOptions({ handleScroll: false, handleScale: false });
        }
    };

    useEffect(() => {
        if (!draggingState) return;

        const handleMouseMove = (e: MouseEvent) => {
            if (!chartContainerRef.current || !seriesRef.current) return;
            const rect = chartContainerRef.current.getBoundingClientRect();
            const y = e.clientY - rect.top;
            const price = seriesRef.current.coordinateToPrice(y);
            if (price) setDragPrice(price);
        };

        const handleMouseUp = () => {
            if (draggingState && dragPrice !== null) {
                if (onModifyPosition) {
                    const pos = positions.find(p => p.id === draggingState.id);
                    if (pos) {
                        const sl = draggingState.type === 'SL' ? dragPrice : pos.sl;
                        const tp = draggingState.type === 'TP' ? dragPrice : pos.tp;
                        onModifyPosition(draggingState.id, sl, tp);
                    }
                }
            }

            setDraggingState(null);
            setDragPrice(null);
            // Re-enable chart Scroll
            if (chartRef.current) {
                chartRef.current.applyOptions({
                    handleScroll: { mouseWheel: true, pressedMouseMove: true, horzTouchDrag: true, vertTouchDrag: false },
                    handleScale: { axisPressedMouseMove: true, mouseWheel: true, pinch: true }
                });
            }
        };

        window.addEventListener('mousemove', handleMouseMove);
        window.addEventListener('mouseup', handleMouseUp);
        return () => {
            window.removeEventListener('mousemove', handleMouseMove);
            window.removeEventListener('mouseup', handleMouseUp);
        };
    }, [draggingState, positions, onModifyPosition]);

    // Force re-render of overlay on scroll/scale
    // Force re-render of overlay on scroll/scale
    useEffect(() => {
        if (!chartRef.current) return;
        const chart = chartRef.current;
        const refresh = () => setOverlayPositions(prev => [...prev]); // Trigger re-render
        chart.timeScale().subscribeVisibleTimeRangeChange(refresh);
        return () => chart.timeScale().unsubscribeVisibleTimeRangeChange(refresh);
    }, [isChartReady]);

    // One-Click Trading Handler
    useEffect(() => {
        if (!chartRef.current || !seriesRef.current || !oneClickEnabled || !onPlacePendingOrder) return;

        const handleChartClick = (param: any) => {
            if (!param.point || !param.time || !seriesRef.current) return;
            const price = seriesRef.current.coordinateToPrice(param.point.y);
            if (price) {
                // Determine Limit vs Stop based on current price
                // Simple logic: if click > current (ask), Buy Stop or Sell Limit?
                // Standard: Click Above -> Sell Limit. Click Below -> Buy Limit.
                // Or user context? Let's implement Limit orders for now.
                // Actually, typically One-Click ON CHART places limit orders.
                // Above price = Sell Limit. Below price = Buy Limit.
                onPlacePendingOrder(price, 'LIMIT');
            }
        };

        chartRef.current.subscribeClick(handleChartClick);
        return () => {
            if (chartRef.current) chartRef.current.unsubscribeClick(handleChartClick);
        };
    }, [oneClickEnabled, onPlacePendingOrder, currentPrice]);

    // Manage indicator series rendering
    useEffect(() => {
        if (!isChartReady || !chartRef.current) return;

        const chart = chartRef.current;
        const seriesMap = indicatorSeriesRef.current;

        // Get indicator types currently on chart
        const currentIndicatorIds = new Set(overlayIndicators.map(ind => ind.id));

        // Remove series for indicators no longer on chart
        for (const [id, series] of seriesMap.entries()) {
            if (!currentIndicatorIds.has(id)) {
                try {
                    chart.removeSeries(series);
                    seriesMap.delete(id);
                } catch (err) {
                    console.error('Failed to remove indicator series:', err);
                }
            }
        }

        // Add or update series for overlay indicators
        for (const indicator of overlayIndicators) {
            if (!indicator.visible) continue;

            try {
                let series = seriesMap.get(indicator.id);

                // Create series if it doesn't exist
                if (!series) {
                    series = chart.addSeries(LineSeries, {
                        color: indicator.color,
                        lineWidth: indicator.lineWidth as 1 | 2 | 3 | 4,
                    });
                    seriesMap.set(indicator.id, series);
                }

                // Update series data for single-value indicators
                if (Array.isArray(indicator.data) && indicator.data.length > 0) {
                    const firstValue = indicator.data[0].value;

                    // Handle single-value indicators (e.g., SMA, EMA, RSI)
                    if (typeof firstValue === 'number') {
                        const lineData = indicator.data.map((point) => ({
                            time: point.time as Time,
                            value: point.value as number,
                        }));
                        series.setData(lineData);
                    }
                }
            } catch (err) {
                console.error(`Failed to render indicator ${indicator.type}:`, err);
            }
        }
    }, [overlayIndicators, isChartReady]);


    const handleAddIndicator = (type: IndicatorType, params?: IndicatorParams) => {
        addIndicator(type, params);
        setShowIndicatorManager(false);
    };

    return (
        <div className="relative w-full h-full bg-[#131722] flex flex-col" role="region" aria-label="Trading Chart">
            <div ref={chartContainerRef} className="w-full flex-1" />

            <DrawingTools
                activeTool={activeTool}
                onToolChange={setActiveTool}
                onClearAll={() => {
                    // TODO: Implement clear all
                    setDrawings([]);
                }}
            />

            {/* Drawing Overlay */}
            {isChartReady && (
                <DrawingOverlay
                    chart={chartRef.current}
                    series={seriesRef.current}
                    drawings={drawings}
                    selectedTool={activeTool}
                    symbol={symbol}
                    accountId={1} // TODO: Prop
                    onUpdateDrawing={handleUpdateDrawing}
                    onFinishDrawing={() => setActiveTool('cursor')}
                    onDeleteDrawing={handleDeleteDrawing}
                    containerRef={chartContainerRef as React.RefObject<HTMLDivElement>}
                />
            )}

            {/* HTML Overlays (Positions) */}
            <div className="absolute inset-0 pointer-events-none overflow-hidden z-30">
                {overlayPositions.map((pos) => (
                    <PositionOverlay
                        key={pos.id}
                        pos={pos}
                        draggingState={draggingState}
                        onDragStart={handleDragStart}
                        onClose={() => onClosePosition && onClosePosition(pos.id)}
                    />
                ))}

                {/* Active Drag Line */}
                {draggingState && dragPrice && (
                    <div
                        className="absolute left-0 right-0 border-b border-dashed flex items-center justify-end pr-12 z-40"
                        style={{
                            top: seriesRef.current?.priceToCoordinate(dragPrice) ?? 0,
                            height: 1,
                            borderColor: draggingState.type === 'TP' ? '#10b981' : '#ef4444'
                        }}
                    >
                        <div className={`text-white text-[10px] px-1 rounded-sm -mt-5 flex items-center gap-1 ${draggingState.type === 'TP' ? 'bg-emerald-500' : 'bg-red-500'
                            }`}>
                            <span>{draggingState.type}: {dragPrice.toFixed(5)}</span>
                        </div>
                    </div>
                )}
            </div>

            {/* Legend and Controls */}
            <div className="absolute top-4 left-4 z-10 flex items-start gap-4">
                <div className="pointer-events-none">
                    <div className="text-2xl font-bold text-zinc-100">{symbol}</div>
                    <div className="text-sm text-zinc-500 font-medium">{timeframe}</div>
                </div>
                <button
                    onClick={() => setShowIndicatorManager(true)}
                    className="pointer-events-auto px-3 py-1.5 text-xs font-medium rounded bg-zinc-800/80 hover:bg-zinc-700 text-zinc-300 hover:text-white border border-zinc-700 hover:border-zinc-600 transition-colors"
                    title="Add technical indicators"
                >
                    📊 Indicators
                </button>
            </div>

            {/* Indicator Manager Modal */}
            <IndicatorManager
                isOpen={showIndicatorManager}
                onClose={() => setShowIndicatorManager(false)}
                onAddIndicator={handleAddIndicator}
                currentIndicators={indicators.map(ind => ind.type)}
            />

            {/* Indicator Panes - Render each pane group */}
            {Array.from(paneGroups.entries()).map(([paneIndex, paneIndicators]) => (
                <IndicatorPane
                    key={paneIndex}
                    indicators={paneIndicators}
                    height={150}
                    mainChartRef={chartRef.current}
                />
            ))}
        </div>
    );
}

function PositionOverlay({ pos, draggingState, onDragStart, onClose }: any) {
    const isDraggingThis = draggingState?.id === pos.id;
    const pnl = pos.unrealizedPnL || 0;
    const pnlColor = pnl >= 0 ? 'text-emerald-400' : 'text-red-400';


    return (
        <>
            {/* Entry Line - Blue solid line with price label on right */}
            {pos.yEntry !== null && (
                <div
                    className="absolute left-0 right-0 pointer-events-auto group z-20"
                    style={{ top: pos.yEntry }}
                >
                    {/* Line */}
                    <div className="absolute inset-x-0 top-0 h-[2px] bg-blue-500" />

                    {/* Left Label - Position Info */}
                    <div className="absolute left-3 -top-3 flex items-center gap-2 opacity-90 group-hover:opacity-100 transition-opacity">
                        <div className={`${pos.side === 'BUY' ? 'bg-emerald-500' : 'bg-red-500'} text-white text-[10px] font-bold px-1.5 py-0.5 rounded-sm shadow-lg`}>
                            {pos.side} {pos.volume}
                        </div>
                        <span className={`text-[10px] font-mono ${pnlColor} font-bold`}>
                            {pnl >= 0 ? '+' : ''}{pnl.toFixed(2)}
                        </span>
                    </div>

                    {/* Right Label - Price Tag (MT5 Style) */}
                    <div className="absolute right-0 -top-2.5 flex items-center">
                        <div className="relative">
                            <div className="bg-blue-500 text-white text-[10px] font-mono font-bold px-2 py-0.5 rounded-l-sm shadow-lg">
                                {pos.openPrice.toFixed(5)}
                            </div>
                            {/* Triangle pointer */}
                            <div className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-full w-0 h-0 border-t-[6px] border-t-transparent border-b-[6px] border-b-transparent border-l-[6px] border-l-blue-500" />
                        </div>
                    </div>

                    {/* Close Button on Hover */}
                    <div className="absolute left-24 -top-2.5 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                            onClick={onClose}
                            className="bg-red-500 hover:bg-red-600 text-white text-[10px] px-1.5 py-0.5 rounded-sm shadow-lg font-bold"
                        >
                            ✕ CLOSE
                        </button>
                    </div>
                </div>
            )}

            {/* SL Line - Red dashed with fill zone */}
            {pos.ySL !== null && pos.sl > 0 && (
                <div
                    className={`absolute left-0 right-0 pointer-events-auto cursor-ns-resize group z-20 ${isDraggingThis && draggingState.type === 'SL' ? 'opacity-0' : ''}`}
                    style={{ top: pos.ySL }}
                    onMouseDown={(e) => { e.preventDefault(); e.stopPropagation(); onDragStart(pos.id, 'SL', pos.sl); }}
                >
                    {/* Invisible Hit Area for easier grabbing - Increased height */}
                    <div className="absolute inset-x-0 -top-4 h-8 bg-transparent" />

                    {/* Danger Zone Fill (between entry and SL) */}
                    {pos.yEntry !== null && (
                        <div
                            className="absolute inset-x-0 bg-gradient-to-b from-red-500/10 to-red-500/5 pointer-events-none"
                            style={{
                                top: pos.side === 'BUY' ? 0 : -(pos.yEntry - pos.ySL),
                                height: Math.abs(pos.yEntry - pos.ySL)
                            }}
                        />
                    )}

                    {/* Line */}
                    <div className="absolute inset-x-0 top-0 h-[1px] border-t border-dashed border-red-500 group-hover:border-solid group-hover:h-[2px] shadow-sm" />

                    {/* Right Label */}
                    <div className="absolute right-0 -top-2.5 flex items-center">
                        <div className="relative hover:scale-110 transition-transform origin-right">
                            <div className="bg-red-500 text-white text-[10px] font-mono font-bold px-2 py-0.5 rounded-l-md shadow-lg flex items-center gap-1 border border-red-400/50">
                                <span className="opacity-70">SL</span> {pos.sl.toFixed(5)}
                            </div>
                            <div className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-full w-0 h-0 border-t-[6px] border-t-transparent border-b-[6px] border-b-transparent border-l-[6px] border-l-red-500" />
                        </div>
                    </div>

                    {/* Drag Handle Hint */}
                    <div className="absolute left-3 -top-2 opacity-0 group-hover:opacity-100 transition-opacity text-[10px] text-red-500 font-bold bg-white/10 backdrop-blur-md px-1 rounded">
                        <span>Drag to Modify</span>
                    </div>
                </div>
            )}

            {/* TP Line - Green dashed with fill zone */}
            {pos.yTP !== null && pos.tp > 0 && (
                <div
                    className={`absolute left-0 right-0 pointer-events-auto cursor-ns-resize group z-20 ${isDraggingThis && draggingState.type === 'TP' ? 'opacity-0' : ''}`}
                    style={{ top: pos.yTP }}
                    onMouseDown={(e) => { e.preventDefault(); e.stopPropagation(); onDragStart(pos.id, 'TP', pos.tp); }}
                >
                    {/* Invisible Hit Area for easier grabbing - Increased height */}
                    <div className="absolute inset-x-0 -top-4 h-8 bg-transparent" />

                    {/* Profit Zone Fill (between entry and TP) */}
                    {pos.yEntry !== null && (
                        <div
                            className="absolute inset-x-0 bg-gradient-to-b from-emerald-500/5 to-emerald-500/10 pointer-events-none"
                            style={{
                                top: pos.side === 'BUY' ? -(pos.yEntry - pos.yTP) : 0,
                                height: Math.abs(pos.yEntry - pos.yTP)
                            }}
                        />
                    )}

                    {/* Line */}
                    <div className="absolute inset-x-0 top-0 h-[1px] border-t border-dashed border-emerald-500 group-hover:border-solid group-hover:h-[2px] shadow-sm" />

                    {/* Right Label */}
                    <div className="absolute right-0 -top-2.5 flex items-center">
                        <div className="relative hover:scale-110 transition-transform origin-right">
                            <div className="bg-emerald-500 text-white text-[10px] font-mono font-bold px-2 py-0.5 rounded-l-md shadow-lg flex items-center gap-1 border border-emerald-400/50">
                                <span className="opacity-70">TP</span> {pos.tp.toFixed(5)}
                            </div>
                            <div className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-full w-0 h-0 border-t-[6px] border-t-transparent border-b-[6px] border-b-transparent border-l-[6px] border-l-emerald-500" />
                        </div>
                    </div>

                    {/* Drag Handle Hint */}
                    <div className="absolute left-3 -top-2 opacity-0 group-hover:opacity-100 transition-opacity text-[10px] text-emerald-500 font-bold bg-white/10 backdrop-blur-md px-1 rounded">
                        <span>Drag to Modify</span>
                    </div>
                </div>
            )}
        </>
    );
}

function getTimeframeSeconds(tf: Timeframe): number {
    switch (tf) {
        case '1m': return 60;
        case '5m': return 300;
        case '15m': return 900;
        case '1h': return 3600;
        case '4h': return 14400;
        case '1d': return 86400;
        default: return 60;
    }
}

function formatDataForSeries(candles: OHLC[], chartType: ChartType): any[] {
    if (chartType === 'heikinAshi') {
        return candles.map((c, i) => toHeikinAshi(c, i > 0 ? candles[i - 1] : undefined));
    }
    if (chartType === 'line' || chartType === 'area') {
        return candles.map(c => ({ time: c.time, value: c.close }));
    }
    return candles;
}



function toHeikinAshi(candle: OHLC, prevCandle?: OHLC): OHLC {
    const haClose = (candle.open + candle.high + candle.low + candle.close) / 4;
    const haOpen = prevCandle ? (prevCandle.open + prevCandle.close) / 2 : (candle.open + candle.close) / 2;
    const haHigh = Math.max(candle.high, haOpen, haClose);
    const haLow = Math.min(candle.low, haOpen, haClose);
    return { time: candle.time, open: haOpen, high: haHigh, low: haLow, close: haClose };
}

// Chart Controls Component
// Chart Controls Component
export function ChartControls({
    chartType, timeframe, onChartTypeChange, onTimeframeChange, isMaximized, onToggleMaximize, oneClickEnabled, onToggleOneClick, onDownloadData, onOpenIndicators
}: {
    chartType: ChartType;
    timeframe: Timeframe;
    onChartTypeChange: (type: ChartType) => void;
    onTimeframeChange: (tf: Timeframe) => void;
    isMaximized: boolean;
    onToggleMaximize: () => void;
    oneClickEnabled?: boolean;
    onToggleOneClick?: () => void;
    onDownloadData?: () => void;
    onOpenIndicators?: () => void;
}) {
    const chartTypes: { value: ChartType; label: string }[] = [
        { value: 'candlestick', label: 'Candles' },
        { value: 'heikinAshi', label: 'HA' },
        { value: 'line', label: 'Line' },
    ];

    const timeframes: { value: Timeframe; label: string }[] = [
        { value: '1m', label: 'M1' }, { value: '5m', label: 'M5' },
        { value: '15m', label: 'M15' }, { value: '1h', label: 'H1' },
    ];

    return (
        <div className="flex items-center gap-2 px-3 py-1.5 bg-zinc-900/80 border-b border-zinc-800">
            {/* Timeframes */}
            <div className="flex items-center bg-zinc-800/50 rounded-md p-0.5">
                {timeframes.map((tf) => (
                    <button
                        key={tf.value}
                        onClick={() => onTimeframeChange(tf.value)}
                        className={`px-2.5 py-1 text-[11px] font-medium rounded transition-all duration-150 ${timeframe === tf.value
                            ? 'bg-emerald-500/20 text-emerald-400 shadow-sm'
                            : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-700/50'
                            }`}
                    >
                        {tf.label}
                    </button>
                ))}
            </div>

            <div className="w-px h-4 bg-zinc-700/50 mx-1" />

            {/* Chart Types */}
            <div className="flex items-center bg-zinc-800/50 rounded-md p-0.5">
                {chartTypes.map((ct) => (
                    <button
                        key={ct.value}
                        onClick={() => onChartTypeChange(ct.value)}
                        className={`px-2 py-1 text-[11px] font-medium rounded transition-all duration-150 ${chartType === ct.value
                            ? 'bg-emerald-500/20 text-emerald-400 shadow-sm'
                            : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-700/50'
                            }`}
                    >
                        {ct.label}
                    </button>
                ))}
            </div>

            <div className="flex-1" />

            {/* Download Data Button */}
            {onDownloadData && (
                <button
                    onClick={onDownloadData}
                    className="px-2 py-1 text-[10px] font-medium rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 transition-colors mr-2"
                    title="Download chart data as CSV"
                >
                    ⬇ CSV
                </button>
            )}

            {/* Indicators Button */}
            {onOpenIndicators && (
                <button
                    onClick={onOpenIndicators}
                    className="px-2 py-1 text-[10px] font-medium rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 transition-colors mr-2"
                    title="Add technical indicators"
                >
                    📊 Indicators
                </button>
            )}

            {/* One-Click Toggle */}
            {onToggleOneClick && (
                <button
                    onClick={onToggleOneClick}
                    className={`px-2 py-1 text-[10px] font-bold uppercase rounded border transition-colors mr-2 ${oneClickEnabled
                        ? 'bg-blue-500/20 text-blue-400 border-blue-500/30'
                        : 'bg-transparent text-zinc-500 border-zinc-700 hover:text-zinc-300'
                        }`}
                    title="One-Click Trading: Click chart to place pending order"
                >
                    One-Click {oneClickEnabled ? 'ON' : 'OFF'}
                </button>
            )}

            <button
                onClick={onToggleMaximize}
                className="p-1.5 text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 rounded transition-colors"
                title={isMaximized ? "Restore" : "Maximize"}
            >
                {isMaximized ? <Minimize2 size={16} /> : <Maximize2 size={16} />}
            </button>
        </div>
    );
}

// Helper icons (need to import Minimize2, Maximize2 in this file or pass icons? Better to import)

