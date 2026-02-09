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
import { useAppStore } from '../store/useAppStore';
import { DrawingContextMenu } from './DrawingContextMenu';

export type ChartType = 'candlestick' | 'heikinAshi' | 'bar' | 'line' | 'area';
export type Timeframe = 'M1' | 'M5' | 'M15' | 'M30' | 'H1' | 'H4' | 'D1' | 'W1' | 'MN';

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
    currentPrice: _, // Unused but kept for prop compatibility
    chartType = 'candlestick',
    timeframe = 'M1',
    positions = [],
    onClosePosition,
    onModifyPosition
}: ChartProps) {
    const chartContainerRef = useRef<HTMLDivElement>(null);
    const chartRef = useRef<IChartApi | null>(null);
    const seriesRef = useRef<ISeriesApi<any> | null>(null);
    const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null);
    const bidLineRef = useRef<any>(null);
    const canvasRef = useRef<HTMLCanvasElement | null>(null);
    const isDisposedRef = useRef(false); // Track if chart is disposed to prevent operations on disposed chart

    // Separate state: historical loaded once, forming updates from ticks
    const historicalCandlesRef = useRef<OHLC[]>([]); // Loaded from API, never modified
    const formingCandleRef = useRef<OHLC | null>(null); // Updates from real-time ticks

    const [isChartReady, setIsChartReady] = useState(false);

    // State for overlays
    const [overlayPositions, setOverlayPositions] = useState<any[]>([]);

    // Get real-time tick data from AppStore (Fast Path)
    const currentTick = useAppStore(state => state.ticks[symbol]);

    // Legend Data State
    const [legendData, setLegendData] = useState<Partial<OHLC>>({});

    // Active tool state for cursor management
    const [activeDrawingType, setActiveDrawingType] = useState<string | null>(null);

    // Initialize chart ONCE
    useEffect(() => {
        if (!chartContainerRef.current) return;

        try {
            const chart = createChart(chartContainerRef.current, {
                layout: {
                    background: { type: ColorType.Solid, color: '#000000' }, // Black background
                    textColor: '#d1d4dc',
                    attributionLogo: false,
                },
                grid: {
                    vertLines: { color: '#2b2b2b', style: 2 }, // Dotted Grid
                    horzLines: { color: '#2b2b2b', style: 2 }, // Dotted Grid
                },
                crosshair: {
                    mode: CrosshairMode.Normal,
                    vertLine: {
                        color: '#758696',
                        width: 1,
                        style: 2,
                        labelBackgroundColor: '#131722',
                        labelVisible: true,
                    },
                    horzLine: {
                        color: '#758696',
                        width: 1,
                        style: 2,
                        labelBackgroundColor: '#131722',
                        labelVisible: true,
                    },
                },
                rightPriceScale: {
                    borderColor: '#2b2b2b',
                    scaleMargins: { top: 0.1, bottom: 0.2 },
                    ticksVisible: true,
                    borderVisible: true,
                },
                timeScale: {
                    borderColor: '#2b2b2b',
                    timeVisible: true,
                    secondsVisible: false,
                    ticksVisible: true,
                    borderVisible: true,
                },
                handleScroll: { mouseWheel: true, pressedMouseMove: true, horzTouchDrag: true, vertTouchDrag: false },
                handleScale: { axisPressedMouseMove: true, mouseWheel: false, pinch: true },
            });

            chartRef.current = chart;

            // Set chart reference in managers
            chartManager.setChart(chart);
            indicatorManager.setChart(chart);

            // Subscribe to crosshair move for Legend
            chart.subscribeCrosshairMove((param) => {
                if (param.time && seriesRef.current) {
                    const data = param.seriesData.get(seriesRef.current) as OHLC;
                    if (data) {
                        setLegendData({
                            open: data.open,
                            high: data.high,
                            low: data.low,
                            close: data.close,
                            volume: (data as any).value || (data as any).volume
                        });
                    }
                } else {
                    const last = historicalCandlesRef.current.length > 0
                        ? historicalCandlesRef.current[historicalCandlesRef.current.length - 1]
                        : formingCandleRef.current;

                    if (last) {
                        setLegendData({
                            open: last.open,
                            high: last.high,
                            low: last.low,
                            close: last.close,
                            volume: last.volume
                        });
                    }
                }
            });

            // Subscribe to clicks for drawing tools
            chart.subscribeClick((param) => {
                const activeDrawing = drawingManager.getActiveDrawing();

                if (!param.time || !param.point || !seriesRef.current) {
                    if (!activeDrawing) {
                        drawingManager.unselectAll();
                    }
                    return;
                }

                if (activeDrawing) {
                    const price = seriesRef.current.coordinateToPrice(param.point.y);
                    if (price !== null) {
                        if (activeDrawing.type === 'text') {
                            const text = prompt('Enter annotation text:');
                            if (text) {
                                drawingManager.updateActiveDrawing({ text });
                                drawingManager.addPoint(param.time as number, price);
                            } else {
                                drawingManager.cancelDrawing();
                            }
                        } else {
                            drawingManager.addPoint(param.time as number, price);
                        }
                    }
                } else {
                    // If no active drawing, maybe we clicked to unselect?
                    // The drawing manager handles clicks on drawings via DOM events,
                    // but we can unselect if we clicked the background.
                    // We'll use a timeout to let the drawing click handle it first if any.
                    setTimeout(() => {
                        // If no drawing was selected (handled by DOM), unselect all
                    }, 50);
                    // For now, simpler:
                    drawingManager.unselectAll();
                }
            });

            const handleResize = () => {
                if (chartContainerRef.current && chartRef.current) {
                    chartRef.current.applyOptions({
                        width: chartContainerRef.current.clientWidth,
                        height: chartContainerRef.current.clientHeight,
                    });
                }
            };

            const resizeObserver = new ResizeObserver(() => handleResize());
            resizeObserver.observe(chartContainerRef.current);

            // Initial resize
            handleResize();
            setIsChartReady(true);

            // Get canvas reference for export functionality
            // Store the canvas ref globally for SavePictureDialog
            const findCanvas = () => {
                const canvas = chartContainerRef.current?.querySelector('canvas');
                if (canvas) {
                    canvasRef.current = canvas;
                    // Store globally for access from FileMenu
                    (window as any).__activeChartCanvas = canvas;
                    (window as any).__activeChartContainer = chartContainerRef.current;
                }
            };

            // Try immediately and then with a delay (canvas may render async)
            findCanvas();
            const canvasInterval = setInterval(findCanvas, 100);
            setTimeout(() => clearInterval(canvasInterval), 2000);

            // Reset disposed flag on mount
            isDisposedRef.current = false;

            return () => {
                // CRITICAL: Set disposed flag FIRST to prevent any pending operations
                isDisposedRef.current = true;

                resizeObserver.disconnect();
                setIsChartReady(false);

                // Clear manager references
                chartManager.setChart(null);
                indicatorManager.setChart(null);
                drawingManager.setChart(null, null);

                // Clear global refs
                if ((window as any).__activeChartCanvas === canvasRef.current) {
                    (window as any).__activeChartCanvas = null;
                }
                if ((window as any).__activeChartContainer === chartContainerRef.current) {
                    (window as any).__activeChartContainer = null;
                }

                clearInterval(canvasInterval);

                // Remove chart last (after all cleanup)
                try {
                    chart.remove();
                } catch (e) {
                    // Ignore - chart may already be disposed
                }

                chartRef.current = null;
                seriesRef.current = null;
                canvasRef.current = null;
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
                // SCORCHED EARTH: Force-remove price lines BEFORE series removal to prevent leaks
                activePriceLines.current.forEach(line => {
                    try { seriesRef.current?.removePriceLine(line); } catch (e) { }
                });
                activePriceLines.current.clear();
                bidLineRef.current = null;

                try { chartRef.current.removeSeries(seriesRef.current); } catch (e) { }
                seriesRef.current = null;
            }

            let series: any;
            const chart = chartRef.current;

            const upColor = '#26a69a'; // Teal/Green (Classic)
            const downColor = '#ef5350'; // Red (Classic)

            switch (chartType) {
                case 'candlestick':
                case 'heikinAshi':
                    series = chart.addSeries(CandlestickSeries, {
                        upColor: '#000000', // Black body (Hollow look)
                        downColor: '#FFFFFF', // White body (Solid)
                        borderUpColor: '#00FF00', // Green border
                        borderDownColor: '#FFFFFF', // White border
                        wickUpColor: '#00FF00', // Green wick
                        wickDownColor: '#FFFFFF', // White wick
                        priceFormat: { type: 'price', precision: 5, minMove: 0.00001 },
                        lastValueVisible: false, priceLineVisible: false
                    });
                    break;
                case 'bar':
                    series = chart.addSeries(BarSeries, {
                        upColor: upColor, downColor: downColor,
                        lastValueVisible: false, priceLineVisible: false
                    });
                    break;
                case 'line':
                    series = chart.addSeries(LineSeries, {
                        color: upColor, lineWidth: 2,
                        lastValueVisible: false, priceLineVisible: false
                    });
                    break;
                case 'area':
                    series = chart.addSeries(AreaSeries, {
                        lineColor: upColor, topColor: 'rgba(38, 166, 154, 0.4)',
                        bottomColor: 'rgba(38, 166, 154, 0.0)', lineWidth: 2,
                        lastValueVisible: false, priceLineVisible: false
                    });
                    break;
                default:
                    series = chart.addSeries(CandlestickSeries, {
                        upColor: upColor, downColor: downColor,
                        borderUpColor: upColor, borderDownColor: downColor,
                        wickUpColor: upColor, wickDownColor: downColor,
                        lastValueVisible: false, priceLineVisible: false
                    });
            }

            seriesRef.current = series;

            // Add volume histogram
            if (volumeSeriesRef.current) {
                try { chartRef.current.removeSeries(volumeSeriesRef.current); } catch (e) { }
            }
            const volumeSeries = chart.addSeries(HistogramSeries, {
                color: '#00FF00', // Match bull candle color
                priceFormat: {
                    type: 'volume',
                },
                priceScaleId: '', // Overlay mode
                lastValueVisible: false, // Fix: Hide volume labels on Y-axis
                priceLineVisible: false, // Fix: Hide volume horizontal lines
            });
            volumeSeries.priceScale().applyOptions({
                scaleMargins: {
                    top: 0.8,
                    bottom: 0,
                },
            });
            volumeSeriesRef.current = volumeSeries;

            // Update drawing manager with new series
            drawingManager.setChart(chartRef.current, series, symbol);

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
                        color: bar.close >= bar.open ? 'rgba(34, 197, 94, 0.5)' : 'rgba(239, 68, 68, 0.5)',
                    }));
                    volumeSeriesRef.current.setData(volumeData);
                }

                // Update indicator manager with combined OHLC data
                indicatorManager.setOHLCData(allCandles);

                // Init Legend with last candle
                const last = allCandles[allCandles.length - 1];
                setLegendData({
                    open: last.open,
                    high: last.high,
                    low: last.low,
                    close: last.close,
                    volume: last.volume
                });
            }
        } catch (err) {
            console.error('Error creating chart series:', err);
        }
    }, [isChartReady, chartType]);

    // Fetch historical OHLC data when symbol or timeframe changes
    useEffect(() => {
        const fetchHistory = async () => {
            if (!seriesRef.current) return;
            console.log('[TradingChart] Fetching history for:', symbol, timeframe);

            try {
                // Fetch pre-aggregated OHLC candles (lightweight)
                // Use limit=1500 to fill a wide screen
                const res = await fetch(`http://localhost:7999/api/history/ohlc?symbol=${symbol}&timeframe=${timeframe}&limit=1500`);
                console.log('[TradingChart] Fetch response status:', res.status);

                if (!res.ok) {
                    console.warn(`No historical data for ${symbol}: ${res.status} ${res.statusText}`);
                    // ... existing error handling ...
                    historicalCandlesRef.current = [];
                    formingCandleRef.current = null;
                    seriesRef.current.setData([]);
                    setLegendData({});
                    return;
                }

                const responseData = await res.json();
                // API wraps candles: {symbol, timeframe, candles: [...], count}
                const rawCandles: OHLC[] = Array.isArray(responseData)
                    ? responseData
                    : responseData.candles || [];
                console.log('[TradingChart] Received candles:', rawCandles.length);

                // Validation and Sanitization
                const validCandles = rawCandles
                    .filter(c => c && c.time && !isNaN(c.open) && !isNaN(c.close))
                    .sort((a, b) => (a.time as number) - (b.time as number))
                    .filter((c, index, array) => index === 0 || c.time > array[index - 1].time); // Deduplicate

                if (validCandles.length > 0) {
                    historicalCandlesRef.current = validCandles;
                    formingCandleRef.current = null;

                    // Merge history with forming candle (Inline Logic)
                    let allCandles = [...validCandles];
                    if (formingCandleRef.current) {
                        const formingCandle = formingCandleRef.current as OHLC;
                        const last = allCandles[allCandles.length - 1];
                        // If forming candle is newer or same time, append/update
                        // Ensure we don't have gaps or duplicates
                        if (last && last.time === formingCandle.time) {
                            allCandles[allCandles.length - 1] = formingCandle;
                        } else if (last && formingCandle.time > last.time) {
                            allCandles.push(formingCandle);
                        }
                    }

                    const formattedData = formatDataForSeries(allCandles, chartType);
                    console.log('[TradingChart] Setting data series (Inlined):', formattedData.length);
                    if (seriesRef.current) {
                        seriesRef.current.setData(formattedData);
                    }

                    if (volumeSeriesRef.current && allCandles.length > 0) {
                        const volumeData = allCandles.map(bar => ({
                            time: bar.time,
                            value: (bar as any).volume || 0,
                            color: bar.close >= bar.open ? 'rgba(34, 197, 94, 0.5)' : 'rgba(239, 68, 68, 0.5)',
                        }));
                        volumeSeriesRef.current.setData(volumeData);
                    }
                    indicatorManager.setOHLCData(allCandles);

                    const last = allCandles[allCandles.length - 1];
                    setLegendData({
                        open: last.open,
                        high: last.high,
                        low: last.low,
                        close: last.close,
                        volume: (last as any).volume
                    });

                    setLegendData({
                        open: last.open,
                        high: last.high,
                        low: last.low,
                        close: last.close,
                        volume: (last as any).volume
                    });
                } else {
                    // Fallback to Mock Data (Client-side)
                    const mockCandles = generateMockData(symbol, timeframe, 1000);
                    historicalCandlesRef.current = mockCandles;
                    formingCandleRef.current = null;

                    const formattedData = formatDataForSeries(mockCandles, chartType);
                    seriesRef.current.setData(formattedData);

                    if (volumeSeriesRef.current) {
                        const volumeData = mockCandles.map(bar => ({
                            time: bar.time,
                            value: bar.volume || 0,
                            color: bar.close >= bar.open ? 'rgba(34, 197, 94, 0.5)' : 'rgba(239, 68, 68, 0.5)',
                        }));
                        volumeSeriesRef.current.setData(volumeData);
                    }
                    indicatorManager.setOHLCData(mockCandles);
                }
            } catch (err) {
                console.error(`Error fetching historical data for ${symbol}:`, err);
                historicalCandlesRef.current = [];
            }
        };

        fetchHistory();
    }, [symbol, timeframe, chartType, isChartReady]);

    // Helper to get timeframe in seconds
    const getTimeframeSeconds = (tf: Timeframe) => {
        switch (tf) {
            case 'M1': return 60;
            case 'M5': return 300;
            case 'M15': return 900;
            case 'M30': return 1800;
            case 'H1': return 3600;
            case 'H4': return 14400;
            case 'D1': return 86400;
            case 'W1': return 604800;
            case 'MN': return 2592000;
            default: return 60;
        }
    };

    // Real-time CANDLE + PRICE LINE updates
    useEffect(() => {
        // Guard: Prevent operations on disposed chart
        if (isDisposedRef.current) return;
        if (!currentTick || !seriesRef.current || !chartRef.current) return;

        // 1. Update Price Line (Removed redundant creation, managed by lifecycle effect)

        // 2. Real-time Candle Aggregation (Grow/Shrink)
        // Calculate timestamp for the current candle based on timeframe
        const tfSeconds = getTimeframeSeconds(timeframe);
        const tfMs = tfSeconds * 1000;
        const tickTime = currentTick.timestamp;
        const candleTime = (Math.floor(tickTime / tfMs) * tfMs) / 1000 as Time; // Lightweight charts uses seconds for Time

        const price = currentTick.bid;
        const volume = currentTick.volume || 1;

        let newCandle: OHLC;

        // Check if we are in a new candle period or updating the current one
        if (formingCandleRef.current && (formingCandleRef.current.time as number) === (candleTime as number)) {
            // Update existing candle
            const current = formingCandleRef.current;
            newCandle = {
                ...current,
                high: Math.max(current.high, price),
                low: Math.min(current.low, price),
                close: price,
                volume: (current.volume || 0) + volume,
            };
        } else {
            // Start new candle
            // If we have a previous candle, we might want to ensure it's closed properly in the series?
            // Usually the series.update handles the transition if the time is newer.
            // But we need an Open price. If it's a new candle, Open = current price (approx) or prev Close.
            // For simplicity, Open = current price if we don't have history.

            const prevClose = formingCandleRef.current?.close || price;
            newCandle = {
                time: candleTime,
                open: prevClose, // Or price if gap? usually prev close.
                high: price,
                low: price,
                close: price,
                volume: volume,
            };

            // Correction: If we really just started, Open should probably be the first tick price of this bar.
            // Converting a tick to a new bar: Open = price.
            if (!formingCandleRef.current || (formingCandleRef.current.time as number) < (candleTime as number)) {
                newCandle.open = price;
                newCandle.high = price;
                newCandle.low = price;
            }
        }

        formingCandleRef.current = newCandle;

        // Update Series
        try {
            if (chartType === 'line' || chartType === 'area') {
                seriesRef.current.update({
                    time: newCandle.time,
                    value: newCandle.close
                });
            } else {
                seriesRef.current.update(newCandle);
            }

            // Update Volume
            if (volumeSeriesRef.current) {
                volumeSeriesRef.current.update({
                    time: newCandle.time,
                    value: newCandle.volume || 0,
                    color: newCandle.close >= newCandle.open ? 'rgba(34, 197, 94, 0.5)' : 'rgba(239, 68, 68, 0.5)',
                });
            }
        } catch (err) {
            // Use console.debug to reduce noise
            // console.debug('Chart update error:', err);
        }

    }, [currentTick, timeframe, chartType]);

    // Update overlay positions
    useEffect(() => {
        // Guard: Prevent operations on disposed chart
        if (isDisposedRef.current) return;
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
        updateOverlays();
        const interval = setInterval(updateOverlays, 100);
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

    // Manage Real-time Price Line (Removed redundant creation, managed by lifecycle effect)

    // Reset price line ref when series changes
    useEffect(() => {
        return () => {
            bidLineRef.current = null;
        };
    }, [chartType, symbol]); // When chart type or symbol changes, series is recreated

    // Command bus subscriptions
    useEffect(() => {
        let unsubscribers: Array<() => void> = [];
        const setupCommandBus = async () => {
            try {
                const { commandBus } = await import('../services/commandBus');

                const unsubscribersList = [
                    commandBus.subscribe('TOGGLE_CROSSHAIR', () => chartManager.toggleCrosshair()),
                    commandBus.subscribe('ZOOM_IN', () => chartManager.zoomIn()),
                    commandBus.subscribe('ZOOM_OUT', () => chartManager.zoomOut()),
                    commandBus.subscribe('FIT_CONTENT', () => chartManager.fitContent()),
                    commandBus.subscribe('SELECT_TOOL', (payload: any) => {
                        if (['trendline', 'hline', 'vline', 'text', 'channel', 'fibonacci', 'shapes'].includes(payload.tool)) {
                            drawingManager.startDrawing(payload.tool, '#3b82f6', payload.subtype);
                            setActiveDrawingType(payload.tool);
                        } else if (payload.tool === 'cursor') {
                            drawingManager.cancelDrawing();
                            setActiveDrawingType(null);
                        }
                    }),
                    commandBus.subscribe('SET_AUTO_SCROLL', (payload: any) => {
                        if (chartRef.current) {
                            chartRef.current.applyOptions({
                                timeScale: { shiftVisibleRangeOnNewBar: payload.enabled }
                            });
                        }
                    }),
                    commandBus.subscribe('SET_CHART_SHIFT', (payload: any) => {
                        if (chartRef.current) {
                            chartRef.current.applyOptions({
                                timeScale: { rightOffset: payload.enabled ? 12 : 0 }
                            });
                        }
                    }),
                    commandBus.subscribe('DELETE_SELECTED_DRAWING', () => {
                        drawingManager.deleteSelected();
                    }),
                    commandBus.subscribe('UNDO', () => {
                        drawingManager.undoDelete();
                    })
                ];
                unsubscribers = unsubscribersList;
            } catch (error) {
                console.log('Command bus not yet available');
            }
        };
        setupCommandBus();
        return () => unsubscribers.forEach(unsub => unsub());
    }, []);

    // Indicator persistence
    useEffect(() => {
        if (!symbol) return;
        // Clean up previous indicators series from the chart
        indicatorManager.clearAll();
        indicatorManager.loadFromStorage(symbol);
        return () => {
            indicatorManager.saveToStorage(symbol);
            indicatorManager.clearAll();
        };
    }, [symbol]);

    // Update drawing positions
    useEffect(() => {
        if (!chartRef.current || !seriesRef.current) return;
        const timeScale = chartRef.current.timeScale();
        const handleVisibleRangeChange = () => drawingManager.updateDrawingPositions();
        timeScale.subscribeVisibleTimeRangeChange(handleVisibleRangeChange);
        return () => timeScale.unsubscribeVisibleTimeRangeChange(handleVisibleRangeChange);
    }, [isChartReady]);

    // Update bid/ask price lines
    // Clean up price lines on unmount or symbol change to prevent "ghost" lines
    // Manage Bid Price Line Lifecycle: Create/Destroy on component mount/symbol change
    // Manage Bid Price Line Lifecycle: Create/Destroy on component mount/symbol change
    // We use a Set to track all created lines to ensure absolute cleanup
    const activePriceLines = useRef<Set<any>>(new Set());

    useEffect(() => {
        if (!seriesRef.current) return;
        const currentSeries = seriesRef.current;

        // 1. SCORCHED EARTH: Remove ALL known price lines for this series
        activePriceLines.current.forEach(line => {
            try {
                currentSeries.removePriceLine(line);
            } catch (e) { /* ignore cleanup errors */ }
        });
        activePriceLines.current.clear();
        bidLineRef.current = null;

        // 2. Create NEW Price Line
        try {
            const newLine = currentSeries.createPriceLine({
                price: 0,
                color: '#ef4444',
                lineWidth: 1,
                lineStyle: 1, // Solid
                axisLabelVisible: true,
                title: 'Bid',
            });

            bidLineRef.current = newLine;
            activePriceLines.current.add(newLine);
        } catch (e) {
            console.warn("Failed to create price line", e);
        }

        return () => {
            // Cleanup on unmount/dep change
            activePriceLines.current.forEach(line => {
                try {
                    currentSeries.removePriceLine(line);
                } catch (e) { }
            });
            activePriceLines.current.clear();
            bidLineRef.current = null;
        };
    }, [seriesRef.current, symbol]);

    // Fast Update Effect: Only updates position, never creates/destroys
    useEffect(() => {
        if (bidLineRef.current && currentTick) {
            // Safety check: ensure line still exists in our set
            if (activePriceLines.current.has(bidLineRef.current)) {
                bidLineRef.current.applyOptions({
                    price: currentTick.bid,
                    title: '', // Remove text title from label, only show price
                });
            }
        }
    }, [currentTick]);

    return (
        <div className="relative w-full h-full bg-white">
            <div
                ref={chartContainerRef}
                className="w-full h-full"
                style={{
                    cursor: activeDrawingType ? 'crosshair' : 'default'
                }}
            />

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

            {/* Context Menu for Drawings */}
            <DrawingContextMenu onClose={() => { }} />

            {/* Legend - Updated with OHLC */}
            <div className="absolute top-4 left-4 z-10 pointer-events-none font-mono text-zinc-800">
                <div className="flex items-center gap-4">
                    <span className="text-lg font-bold text-zinc-900">{symbol}, {timeframe}</span>
                    <div className="hidden sm:flex items-center gap-3 text-xs font-bold">
                        {/* Safe optional chaining for legend data */}
                        <span className={(legendData.close ?? 0) >= (legendData.open ?? 0) ? 'text-emerald-600' : 'text-rose-600'}>
                            O: <span className="text-zinc-700">{legendData.open?.toFixed(5)}</span>
                        </span>
                        <span className={(legendData.close ?? 0) >= (legendData.open ?? 0) ? 'text-emerald-600' : 'text-rose-600'}>
                            H: <span className="text-zinc-700">{legendData.high?.toFixed(5)}</span>
                        </span>
                        <span className={(legendData.close ?? 0) >= (legendData.open ?? 0) ? 'text-emerald-600' : 'text-rose-600'}>
                            L: <span className="text-zinc-700">{legendData.low?.toFixed(5)}</span>
                        </span>
                        <span className={(legendData.close ?? 0) >= (legendData.open ?? 0) ? 'text-emerald-600' : 'text-rose-600'}>
                            C: <span className="text-zinc-700">{legendData.close?.toFixed(5)}</span>
                        </span>
                        <span className="text-zinc-500">
                            Vol: <span className="text-zinc-700">{legendData.volume?.toLocaleString() || 0}</span>
                        </span>
                    </div>
                </div>
            </div>

            {/* DEBUG OVERLAY - REMOVED */}
        </div>
    );
}

function PositionOverlay({ pos, draggingState, onDragStart, onClose }: any) {
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
        case 'M1': return 60;
        case 'M5': return 300;
        case 'M15': return 900;
        case 'M30': return 1800;
        case 'H1': return 3600;
        case 'H4': return 14400;
        case 'D1': return 86400;
        case 'W1': return 604800;
        case 'MN': return 2592000;
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

function generateMockData(symbol: string, timeframe: Timeframe, limit: number): OHLC[] {
    const candles: OHLC[] = [];
    const tfSeconds = getTimeframeSeconds(timeframe);
    let time = (Math.floor(Date.now() / 1000) - limit * tfSeconds) as Time;

    // Start price based on symbol
    let price = 1.0850;
    if (symbol.includes('JPY')) price = 145.00;
    if (symbol.includes('BTC')) price = 42000.00;
    if (symbol.includes('XAU')) price = 2030.00;

    for (let i = 0; i < limit; i++) {
        const volatility = price * 0.0005; // 0.05% per candle
        const change = (Math.random() - 0.5) * volatility;
        const close = price + change;
        const high = Math.max(price, close) + Math.random() * volatility * 0.5;
        const low = Math.min(price, close) - Math.random() * volatility * 0.5;

        candles.push({
            time: time as Time,
            open: price,
            high: high,
            low: low,
            close: close,
            volume: Math.floor(Math.random() * 100)
        });

        price = close;
        time = (time as number + tfSeconds) as Time;
    }
    return candles;
}


