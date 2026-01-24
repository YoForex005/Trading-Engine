import { useEffect, useRef, useState } from 'react';
import {
    createChart,
    ColorType,
    CrosshairMode,
    CandlestickSeries,
    LineSeries,
    BarSeries,
    AreaSeries,
    HistogramSeries,
} from 'lightweight-charts';
import type { IChartApi, ISeriesApi, Time } from 'lightweight-charts';
import { X } from 'lucide-react';
import { chartManager } from '../services/chartManager';
import { drawingManager } from '../services/drawingManager';
import { indicatorManager } from '../services/indicatorManager';
import { useCurrentTick } from '../store/useMarketDataStore';

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
    volume?: number;
}

// Helper function to combine historical and forming candles
function getAllCandles(historicalCandles: OHLC[], formingCandle: OHLC | null): OHLC[] {
    const allCandles = [...historicalCandles];
    if (formingCandle) {
        allCandles.push(formingCandle);
    }
    return allCandles;
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
    const volumeSeriesRef = useRef<ISeriesApi<'Histogram'>>();
    const bidLineRef = useRef<any>(null);
    const askLineRef = useRef<any>(null);

    // Separate state: historical loaded once, forming updates from ticks
    const historicalCandlesRef = useRef<OHLC[]>([]); // Loaded from API, never modified
    const formingCandleRef = useRef<OHLC | null>(null); // Updates from real-time ticks

    const [isChartReady, setIsChartReady] = useState(false);

    // State for overlays
    const [overlayPositions, setOverlayPositions] = useState<any[]>([]);

    // Get real-time tick data from Zustand store
    const currentTick = useCurrentTick(symbol);

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
                    vertLines: { color: 'rgba(70, 70, 70, 0.3)', style: 2 }, // 2 = dotted
                    horzLines: { color: 'rgba(70, 70, 70, 0.3)', style: 2 }, // 2 = dotted
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

            // Set chart reference in managers
            chartManager.setChart(chart);
            indicatorManager.setChart(chart);

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

                // Clear manager references
                chartManager.setChart(null);
                indicatorManager.setChart(null);
                drawingManager.setChart(null, null);

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
                        upColor: '#14b8a6', downColor: '#ef4444',
                        borderUpColor: '#14b8a6', borderDownColor: '#ef4444',
                        wickUpColor: '#14b8a6', wickDownColor: '#ef4444',
                    });
                    break;
                case 'bar':
                    series = chart.addSeries(BarSeries, { upColor: '#14b8a6', downColor: '#ef4444' });
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
                        upColor: '#14b8a6', downColor: '#ef4444',
                        borderUpColor: '#14b8a6', borderDownColor: '#ef4444',
                        wickUpColor: '#14b8a6', wickDownColor: '#ef4444',
                    });
            }

            seriesRef.current = series;

            // Add volume histogram
            if (volumeSeriesRef.current) {
                try { chartRef.current.removeSeries(volumeSeriesRef.current); } catch (e) { }
            }
            const volumeSeries = chart.addSeries(HistogramSeries, {
                color: '#06b6d4', // Cyan
                priceFormat: {
                    type: 'volume',
                },
                priceScaleId: '', // Use default scale
                scaleMargins: {
                    top: 0.8, // Position at bottom 20%
                    bottom: 0,
                },
            });
            volumeSeriesRef.current = volumeSeries;

            // Update drawing manager with new series
            drawingManager.setChart(chartRef.current, series);

            // Get combined candles (historical + forming)
            const allCandles = getAllCandles(historicalCandlesRef.current, formingCandleRef.current);
            if (allCandles.length > 0) {
                const formattedData = formatDataForSeries(allCandles, chartType);
                series.setData(formattedData);

                // Set volume data
                if (volumeSeriesRef.current) {
                    const volumeData = allCandles.map(bar => ({
                        time: bar.time,
                        value: bar.volume || 0,
                        color: bar.close >= bar.open ? 'rgba(6, 182, 212, 0.5)' : 'rgba(239, 68, 68, 0.5)',
                    }));
                    volumeSeriesRef.current.setData(volumeData);
                }

                // Update indicator manager with combined OHLC data
                indicatorManager.setOHLCData(allCandles);
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
                // FIXED: Correct port (7999) and endpoint (/api/history/ticks)
                // Backend server runs on port 7999 (see backend/cmd/server/main.go:1915)
                // Historical tick data API at /api/history/ticks (see backend/api/history.go)
                const dateStr = new Date().toISOString().split('T')[0]; // YYYY-MM-DD
                const res = await fetch(`http://localhost:7999/api/history/ticks?symbol=${symbol}&date=${dateStr}&limit=5000`);

                if (!res.ok) {
                    console.warn(`No historical data for ${symbol}: ${res.status} ${res.statusText}`);
                    historicalCandlesRef.current = [];
                    formingCandleRef.current = null;
                    seriesRef.current.setData([]);
                    return;
                }

                const data = await res.json();

                // Convert tick data to OHLC candles
                if (data.ticks && Array.isArray(data.ticks) && data.ticks.length > 0) {
                    const candles = buildOHLCFromTicks(data.ticks, timeframe);

                    if (candles.length > 0) {
                        // Only update historical candles, forming candle stays separate
                        historicalCandlesRef.current = candles;

                        // Reset forming candle when new historical data is loaded
                        formingCandleRef.current = null;

                        // Get combined candles for display
                        const allCandles = getAllCandles(candles, formingCandleRef.current);
                        const formattedData = formatDataForSeries(allCandles, chartType);
                        seriesRef.current.setData(formattedData);

                        // Set volume data
                        if (volumeSeriesRef.current && allCandles.length > 0) {
                            const volumeData = allCandles.map(bar => ({
                                time: bar.time,
                                value: bar.volume || 0,
                                color: bar.close >= bar.open ? 'rgba(6, 182, 212, 0.5)' : 'rgba(239, 68, 68, 0.5)',
                            }));
                            volumeSeriesRef.current.setData(volumeData);
                        }

                        // Update indicator manager with combined data
                        indicatorManager.setOHLCData(allCandles);
                    } else {
                        console.warn(`No candles built from ${data.ticks.length} ticks for ${symbol}`);
                        historicalCandlesRef.current = [];
                        formingCandleRef.current = null;
                        seriesRef.current.setData([]);
                    }
                } else {
                    console.warn(`No tick data returned for ${symbol}`);
                    historicalCandlesRef.current = [];
                    formingCandleRef.current = null;
                    seriesRef.current.setData([]);
                    if (volumeSeriesRef.current) {
                        volumeSeriesRef.current.setData([]);
                    }
                }
            } catch (err) {
                console.error(`Error fetching historical data for ${symbol}:`, err);
                historicalCandlesRef.current = [];
            }
        };

        fetchHistory();
    }, [symbol, timeframe, chartType]);

    // Update candles with real-time tick data (MT5-correct time-bucket aggregation)
    useEffect(() => {
        if (!currentTick || !seriesRef.current) return;

        const price = (currentTick.bid + currentTick.ask) / 2;
        const tickTime = Math.floor(Date.now() / 1000);
        const tfSeconds = getTimeframeSeconds(timeframe);

        // MT5-CORRECT: Calculate time bucket for this tick
        const candleTime = (Math.floor(tickTime / tfSeconds) * tfSeconds) as Time;

        // Initialize forming candle if needed
        if (!formingCandleRef.current) {
            formingCandleRef.current = {
                time: candleTime,
                open: price,
                high: price,
                low: price,
                close: price,
                volume: 1
            };
            seriesRef.current.update(formingCandleRef.current);
            return;
        }

        // Check if we need to start a new candle (NEW TIME BUCKET)
        if (formingCandleRef.current.time !== candleTime) {
            // Close the previous candle (move to historical)
            historicalCandlesRef.current.push(formingCandleRef.current);

            // Start a new forming candle
            formingCandleRef.current = {
                time: candleTime,
                open: price,
                high: price,
                low: price,
                close: price,
                volume: 1
            };

            // Update chart with the new candle
            seriesRef.current.update(formingCandleRef.current);

            // Update volume series with closed candle
            if (volumeSeriesRef.current && historicalCandlesRef.current.length > 0) {
                const closedCandle = historicalCandlesRef.current[historicalCandlesRef.current.length - 1];
                volumeSeriesRef.current.update({
                    time: closedCandle.time,
                    value: closedCandle.volume || 0,
                    color: closedCandle.close >= closedCandle.open
                        ? 'rgba(6, 182, 212, 0.5)'
                        : 'rgba(239, 68, 68, 0.5)',
                });
            }
        } else {
            // Update the forming candle (SAME TIME BUCKET)
            formingCandleRef.current.high = Math.max(formingCandleRef.current.high, price);
            formingCandleRef.current.low = Math.min(formingCandleRef.current.low, price);
            formingCandleRef.current.close = price;
            formingCandleRef.current.volume = (formingCandleRef.current.volume || 0) + 1;

            // Update chart with updated forming candle
            seriesRef.current.update(formingCandleRef.current);
        }
    }, [currentTick, timeframe]);

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

    // Command bus subscriptions (will work once Agent 1 creates commandBus)
    useEffect(() => {
        // Note: This assumes commandBus will be created by Agent 1
        // When commandBus is available, these subscriptions will activate

        // Try to import commandBus dynamically
        let unsubscribers: Array<() => void> = [];

        const setupCommandBus = async () => {
            try {
                // Dynamic import to avoid errors if not yet created
                const { commandBus } = await import('../services/commandBus');

                // Subscribe to crosshair toggle
                const unsubCrosshair = commandBus.subscribe('TOGGLE_CROSSHAIR', () => {
                    chartManager.toggleCrosshair();
                });

                // Subscribe to zoom in
                const unsubZoomIn = commandBus.subscribe('ZOOM_IN', () => {
                    chartManager.zoomIn();
                });

                // Subscribe to zoom out
                const unsubZoomOut = commandBus.subscribe('ZOOM_OUT', () => {
                    chartManager.zoomOut();
                });

                // Subscribe to fit content
                const unsubFit = commandBus.subscribe('FIT_CONTENT', () => {
                    chartManager.fitContent();
                });

                // Subscribe to drawing tool selection
                const unsubTrendline = commandBus.subscribe('SELECT_TRENDLINE', () => {
                    drawingManager.startDrawing('trendline');
                });

                const unsubHLine = commandBus.subscribe('SELECT_HLINE', () => {
                    drawingManager.startDrawing('hline');
                });

                const unsubVLine = commandBus.subscribe('SELECT_VLINE', () => {
                    drawingManager.startDrawing('vline');
                });

                const unsubText = commandBus.subscribe('SELECT_TEXT', () => {
                    drawingManager.startDrawing('text');
                });

                unsubscribers = [
                    unsubCrosshair,
                    unsubZoomIn,
                    unsubZoomOut,
                    unsubFit,
                    unsubTrendline,
                    unsubHLine,
                    unsubVLine,
                    unsubText
                ];
            } catch (error) {
                // commandBus not yet available, will work when Agent 1 completes
                console.log('Command bus not yet available');
            }
        };

        setupCommandBus();

        return () => {
            unsubscribers.forEach(unsub => unsub());
        };
    }, []);

    // Load drawings and indicators for symbol
    useEffect(() => {
        if (!symbol) return;

        // Load saved drawings
        drawingManager.loadFromStorage(symbol);

        // Load saved indicators
        indicatorManager.loadFromStorage(symbol);

        return () => {
            // Save on unmount
            drawingManager.saveToStorage(symbol);
            indicatorManager.saveToStorage(symbol);
        };
    }, [symbol]);

    // Update drawing positions on chart scroll/zoom
    useEffect(() => {
        if (!chartRef.current || !seriesRef.current) return;

        const timeScale = chartRef.current.timeScale();

        const handleVisibleRangeChange = () => {
            drawingManager.updateDrawingPositions();
        };

        timeScale.subscribeVisibleTimeRangeChange(handleVisibleRangeChange);

        return () => {
            timeScale.unsubscribeVisibleTimeRangeChange(handleVisibleRangeChange);
        };
    }, [isChartReady]);

    // Update bid/ask price lines with real-time ticks
    useEffect(() => {
        if (!seriesRef.current || !currentTick) return;

        try {
            // Remove previous price lines
            if (bidLineRef.current) {
                seriesRef.current.removePriceLine(bidLineRef.current);
                bidLineRef.current = null;
            }
            if (askLineRef.current) {
                seriesRef.current.removePriceLine(askLineRef.current);
                askLineRef.current = null;
            }

            // Create new bid line
            bidLineRef.current = seriesRef.current.createPriceLine({
                price: currentTick.bid,
                color: '#ef4444', // Red
                lineWidth: 1,
                lineStyle: 2, // Dashed
                axisLabelVisible: true,
                title: `Bid ${currentTick.bid.toFixed(5)}`,
            });

            // Create new ask line
            askLineRef.current = seriesRef.current.createPriceLine({
                price: currentTick.ask,
                color: '#14b8a6', // Teal
                lineWidth: 1,
                lineStyle: 2, // Dashed
                axisLabelVisible: true,
                title: `Ask ${currentTick.ask.toFixed(5)}`,
            });
        } catch (error) {
            console.error('Error updating bid/ask price lines:', error);
        }
    }, [currentTick]);

    return (
        <div className="relative w-full h-full bg-[#131722]">
            <div ref={chartContainerRef} className="w-full h-full" />

            {/* Drawing overlay container */}
            <div className="chart-drawing-overlay absolute inset-0 pointer-events-none" />

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

/**
 * Converts tick data from backend API to OHLC candles
 * Backend returns: { timestamp: number (unix ms), bid: number, ask: number, spread: number }
 */
function buildOHLCFromTicks(ticks: any[], timeframe: Timeframe): OHLC[] {
    if (!ticks || ticks.length === 0) return [];

    const tfSeconds = getTimeframeSeconds(timeframe);
    const candleMap = new Map<number, OHLC>();

    // Group ticks into candles by time bucket
    for (const tick of ticks) {
        const price = (tick.bid + tick.ask) / 2; // Mid price
        const timestamp = Math.floor(tick.timestamp / 1000); // Convert ms to seconds
        const candleTime = (Math.floor(timestamp / tfSeconds) * tfSeconds) as Time;

        if (!candleMap.has(candleTime as number)) {
            // Create new candle
            candleMap.set(candleTime as number, {
                time: candleTime,
                open: price,
                high: price,
                low: price,
                close: price,
                volume: 1,
            });
        } else {
            // Update existing candle
            const candle = candleMap.get(candleTime as number)!;
            candle.high = Math.max(candle.high, price);
            candle.low = Math.min(candle.low, price);
            candle.close = price;
            candle.volume = (candle.volume || 0) + 1; // Tick count as volume
        }
    }

    // Sort candles by time
    return Array.from(candleMap.values()).sort((a, b) => (a.time as number) - (b.time as number));
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
