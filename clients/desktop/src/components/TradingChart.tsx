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
import { X } from 'lucide-react';

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
}

interface OHLC {
    time: Time;
    open: number;
    high: number;
    low: number;
    close: number;
}

export function TradingChart({
    symbol,
    currentPrice,
    chartType = 'candlestick',
    timeframe = '1m',
    positions = [],
    onClosePosition,
    onModifyPosition
}: ChartProps) {
    const chartContainerRef = useRef<HTMLDivElement>(null);
    const chartRef = useRef<IChartApi | null>(null);
    const seriesRef = useRef<ISeriesApi<any> | null>(null);
    const candlesRef = useRef<OHLC[]>([]);
    const [isChartReady, setIsChartReady] = useState(false);

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
                const res = await fetch(`http://localhost:8080/ohlc?symbol=${symbol}&timeframe=${timeframe}&limit=500`);

                if (!res.ok) {
                    candlesRef.current = [];
                    seriesRef.current.setData([]);
                    return;
                }

                const data = await res.json();

                if (Array.isArray(data) && data.length > 0) {
                    const candles: OHLC[] = data.map((d: any) => ({
                        time: d.time as Time,
                        open: d.open, high: d.high, low: d.low, close: d.close,
                    }));

                    candlesRef.current = candles;
                    const formattedData = formatDataForSeries(candles, chartType);
                    seriesRef.current.setData(formattedData);
                } else {
                    candlesRef.current = [];
                    seriesRef.current.setData([]);
                }
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

    return (
        <div className="relative w-full h-full bg-[#131722]">
            <div ref={chartContainerRef} className="w-full h-full" />

            {/* HTML Overlays */}
            <div className="absolute inset-0 pointer-events-none overflow-hidden">
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
                        className="absolute left-0 right-0 border-b border-dashed flex items-center justify-end pr-12 z-30"
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

            {/* Legend */}
            <div className="absolute top-4 left-4 z-10 pointer-events-none">
                <div className="text-2xl font-bold text-zinc-100">{symbol}</div>
                <div className="text-sm text-zinc-500 font-medium">{timeframe}</div>
            </div>
        </div>
    );
}

function PositionOverlay({ pos, draggingState, onDragStart, onClose }: any) {
    // If this position is being dragged, override the Y position of the active line
    const isDraggingThis = draggingState?.id === pos.id;

    return (
        <>
            {/* Entry Line */}
            {pos.yEntry !== null && (
                <div
                    className="absolute left-0 right-0 border-b border-dotted border-blue-500 flex items-center justify-end pr-2 pointer-events-auto group hover:border-solid hover:border-2 z-20"
                    style={{ top: pos.yEntry, height: 1 }}
                >
                    <div className="bg-blue-500 text-white text-[10px] px-1 rounded-sm flex items-center gap-1 -mt-5 opacity-0 group-hover:opacity-100 transition-opacity">
                        <span>#{pos.id} {pos.side} {pos.volume}</span>
                        <X size={12} className="cursor-pointer hover:text-red-300" onClick={onClose} />
                    </div>
                </div>
            )}

            {/* SL Line */}
            {(pos.ySL !== null || !pos.sl) && (
                <div
                    className={`absolute left-0 right-0 border-b border-dashed border-red-500 flex items-center justify-end pr-12 pointer-events-auto cursor-ns-resize group z-20 ${isDraggingThis && draggingState.type === 'SL' ? 'opacity-0' : ''}`}
                    style={{ top: pos.ySL ?? pos.yEntry, height: 1, opacity: pos.ySL ? 1 : 0.5 }}
                    onMouseDown={(e) => {
                        e.stopPropagation();
                        onDragStart(pos.id, 'SL', pos.sl || pos.openPrice);
                    }}
                >
                    <div className="bg-red-500 text-white text-[10px] px-1 rounded-sm -mt-5 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        <span>SL: {pos.sl}</span>
                    </div>
                </div>
            )}

            {/* TP Line */}
            {(pos.yTP !== null || !pos.tp) && (
                <div
                    className={`absolute left-0 right-0 border-b border-dashed border-emerald-500 flex items-center justify-end pr-12 pointer-events-auto cursor-ns-resize group z-20 ${isDraggingThis && draggingState.type === 'TP' ? 'opacity-0' : ''}`}
                    style={{ top: pos.yTP ?? pos.yEntry, height: 1, opacity: pos.yTP ? 1 : 0.5 }}
                    onMouseDown={(e) => {
                        e.stopPropagation();
                        onDragStart(pos.id, 'TP', pos.tp || pos.openPrice);
                    }}
                >
                    <div className="bg-emerald-500 text-white text-[10px] px-1 rounded-sm -mt-5 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        <span>TP: {pos.tp}</span>
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
    chartType, timeframe, onChartTypeChange, onTimeframeChange, isMaximized, onToggleMaximize
}: {
    chartType: ChartType;
    timeframe: Timeframe;
    onChartTypeChange: (type: ChartType) => void;
    onTimeframeChange: (tf: Timeframe) => void;
    isMaximized: boolean;
    onToggleMaximize: () => void;
}) {
    const chartTypes: { value: ChartType; label: string }[] = [
        { value: 'candlestick', label: 'Candles' },
        { value: 'heikinAshi', label: 'Heiken Ashi' },
        { value: 'bar', label: 'OHLC' },
        { value: 'line', label: 'Line' },
        { value: 'area', label: 'Area' },
    ];

    const timeframes: { value: Timeframe; label: string }[] = [
        { value: '1m', label: 'M1' }, { value: '5m', label: 'M5' },
        { value: '15m', label: 'M15' }, { value: '1h', label: 'H1' },
        { value: '4h', label: 'H4' }, { value: '1d', label: 'D1' },
    ];

    return (
        <div className="flex items-center gap-2 px-3 py-1.5 bg-zinc-900/80 border-b border-zinc-800">
            <div className="flex items-center">
                <span className="text-[10px] font-medium text-zinc-500 uppercase tracking-wider mr-2">TF</span>
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
            </div>

            <div className="w-px h-5 bg-zinc-700/50 mx-1" />

            <div className="flex items-center">
                <span className="text-[10px] font-medium text-zinc-500 uppercase tracking-wider mr-2">Type</span>
                <div className="flex items-center bg-zinc-800/50 rounded-md p-0.5">
                    {chartTypes.map((ct) => (
                        <button
                            key={ct.value}
                            onClick={() => onChartTypeChange(ct.value)}
                            className={`px-2.5 py-1 text-[11px] font-medium rounded transition-all duration-150 ${chartType === ct.value
                                ? 'bg-emerald-500/20 text-emerald-400 shadow-sm'
                                : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-700/50'
                                }`}
                        >
                            {ct.label}
                        </button>
                    ))}
                </div>
            </div>

            <div className="flex-1" />

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
import { Maximize2, Minimize2 } from 'lucide-react';
