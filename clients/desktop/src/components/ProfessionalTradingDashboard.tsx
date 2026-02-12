
import React, { useState, useEffect, useRef } from 'react';
import { createChart, CandlestickSeries, HistogramSeries, LineSeries } from 'lightweight-charts';
import type { IChartApi, ISeriesApi, CandlestickData, Time } from 'lightweight-charts';
import { OneClickTrading } from './OneClickTrading';
import { MarketWatch } from './MarketWatch';
import './ProfessionalTradingDashboard.css';
import { WS_ENDPOINTS, buildApiUrl } from '../config/api';
import { useWebSocket } from '../hooks/useWebSocket';

// Icons as React components for better performance
const Icons = {
    Crosshair: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
            <path d="M10 1v4M10 15v4M1 10h4M15 10h4M10 8a2 2 0 100 4 2 2 0 000-4z" stroke="currentColor" fill="none" strokeWidth="1.5" />
        </svg>
    ),
    TrendLine: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M3 17L17 3" />
        </svg>
    ),
    HorizontalLine: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M3 10h14" />
        </svg>
    ),
    VerticalLine: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M10 3v14" />
        </svg>
    ),
    Fibonacci: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M3 17L17 3M3 14h14M3 11h14M3 8h14M3 5h14" />
        </svg>
    ),
    Brush: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M12 2l4 4-10 10-4 2 2-4z" />
        </svg>
    ),
    Text: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
            <path d="M6 4h8v2h-3v10h-2V6H6V4z" />
        </svg>
    ),
    Measure: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M3 3l14 14M17 3v14H3" />
        </svg>
    ),
    ZoomIn: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <circle cx="8" cy="8" r="5" />
            <path d="M12 12l5 5M8 5v6M5 8h6" />
        </svg>
    ),
    ZoomOut: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <circle cx="8" cy="8" r="5" />
            <path d="M12 12l5 5M5 8h6" />
        </svg>
    ),
    Settings: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <circle cx="10" cy="10" r="3" />
            <path d="M10 1v2m0 14v2M19 10h-2M3 10H1m15.5-6.5l-1.4 1.4M5.9 14.1l-1.4 1.4m11-11l-1.4 1.4M5.9 5.9L4.5 4.5" />
        </svg>
    ),
    Indicator: () => (
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M2 10h3l2-6 4 12 2-6h5" />
        </svg>
    ),
    ChevronDown: () => (
        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M4 6l4 4 4-4" />
        </svg>
    ),
};

interface OHLC {
    time: number;
    open: number;
    high: number;
    low: number;
    close: number;
    volume?: number;
}

type ChartTool = 'crosshair' | 'trendline' | 'horizontal' | 'vertical' | 'fibonacci' | 'brush' | 'text' | 'measure' | 'zoom_in' | 'zoom_out';
type Timeframe = '1m' | '5m' | '15m' | '30m' | '1h' | '4h' | '1D' | '1W';
type Indicator = 'MA' | 'EMA' | 'RSI' | 'MACD' | 'BB' | 'Volume';

const ProfessionalTradingDashboard: React.FC = () => {
    const chartContainerRef = useRef<HTMLDivElement>(null);
    const chartRef = useRef<IChartApi | null>(null);
    const candleSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null);
    const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null);
    const indicatorSeriesRef = useRef<Map<string, ISeriesApi<'Line'>>>(new Map());
    const wsRef = useRef<WebSocket | null>(null);

    // Chart Tab State
    const [openCharts, setOpenCharts] = useState<{ symbol: string; timeframe: Timeframe }[]>([
        { symbol: 'BTCUSD', timeframe: '1m' }
    ]);
    const [activeChartIndex, setActiveChartIndex] = useState(0);

    // Derived state for current view
    const selectedSymbol = openCharts[activeChartIndex]?.symbol || 'BTCUSD';
    const selectedTimeframe = openCharts[activeChartIndex]?.timeframe || '1m';

    // Helper to update current chart
    const setSelectedTimeframe = (tf: Timeframe) => {
        setOpenCharts(prev => prev.map((c, i) => i === activeChartIndex ? { ...c, timeframe: tf } : c));
    };

    const setSelectedSymbol = (symbol: string) => {
        const existingIndex = openCharts.findIndex(c => c.symbol === symbol);
        if (existingIndex >= 0) {
            setActiveChartIndex(existingIndex);
        } else {
            setOpenCharts(prev => [...prev, { symbol, timeframe: '1m' }]);
            setActiveChartIndex(openCharts.length);
        }
    };

    const handleCloseChart = (index: number) => {
        if (openCharts.length <= 1) return; // Don't close last chart

        const newCharts = openCharts.filter((_, i) => i !== index);
        setOpenCharts(newCharts);

        if (index === activeChartIndex) {
            setActiveChartIndex(Math.max(0, index - 1));
        } else if (index < activeChartIndex) {
            setActiveChartIndex(activeChartIndex - 1);
        }
    };

    const [selectedTool, setSelectedTool] = useState<ChartTool>('crosshair');
    // Removed old useState definitions for selectedTimeframe and selectedSymbol as they are now derived
    const [activeIndicators, setActiveIndicators] = useState<Indicator[]>(['Volume']);
    const [marketWatchPosition, setMarketWatchPosition] = useState<'left' | 'right'>('left');

    const handlePanelDragStart = (e: React.DragEvent) => {
        e.dataTransfer.setData('text/plain', 'market-watch-panel');
        e.dataTransfer.effectAllowed = 'move';
    };

    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'move';
    };

    const handleDrop = (e: React.DragEvent, position: 'left' | 'right') => {
        e.preventDefault();
        setMarketWatchPosition(position);
    };

    const [currentPrice, setCurrentPrice] = useState<number>(0);
    const [bidPrice, setBidPrice] = useState<number>(0);
    const [askPrice, setAskPrice] = useState<number>(0);
    const [tradingVolume, setTradingVolume] = useState<number>(0.01);
    const [showOneClickTrading, setShowOneClickTrading] = useState(false);
    const [priceChange, setPriceChange] = useState<number>(0);
    const [priceChangePercent, setPriceChangePercent] = useState<number>(0);
    const [high24h, setHigh24h] = useState<number>(0);
    const [low24h, setLow24h] = useState<number>(0);
    const [volume24h, setVolume24h] = useState<number>(0);

    const [showIndicatorMenu, setShowIndicatorMenu] = useState(false);
    const [showSymbolMenu, setShowSymbolMenu] = useState(false);
    const [showSettingsMenu, setShowSettingsMenu] = useState(false);

    const timeframes: Timeframe[] = ['1m', '5m', '15m', '30m', '1h', '4h', '1D', '1W'];
    const indicators: Indicator[] = ['MA', 'EMA', 'RSI', 'MACD', 'BB', 'Volume'];
    // Full watchlist matching backend configuration
    const symbols = [
        'GBPUSD', 'USDJPY', 'GBPJPY', 'CADJPY', 'EURJPY', 'EURCAD', 'AUDCAD', 'CADCHF',
        'GBPAUD', 'GBPCAD', 'AUDUSD', 'EURGBP', 'AUDCHF', 'AUDJPY', 'CHFJPY', 'EURUSD',
        'USDCHF', 'USDCAD', 'NZDUSD', 'EURAUD', 'EURCHF', 'AUDNZD', 'GBPCHF', 'NZDCAD',
        'NZDCHF', 'XAUUSD', 'XAGUSD', 'GBPNZD', 'NZDJPY', 'ETHUSD', 'BTCUSD', 'CN50USD',
        'BNBUSD'
    ];

    // Initialize chart
    useEffect(() => {
        if (!chartContainerRef.current) return;

        const chart = createChart(chartContainerRef.current, {
            width: chartContainerRef.current.clientWidth,
            height: chartContainerRef.current.clientHeight,
            layout: {
                background: { color: '#0F1419' },
                textColor: '#9BA3B5',
            },
            grid: {
                vertLines: { color: '#1A202C', style: 1 },
                horzLines: { color: '#1A202C', style: 1 },
            },
            crosshair: {
                mode: 1,
                vertLine: {
                    color: '#3B82F6',
                    width: 1,
                    style: 3,
                    labelBackgroundColor: '#2563EB',
                },
                horzLine: {
                    color: '#3B82F6',
                    width: 1,
                    style: 3,
                    labelBackgroundColor: '#2563EB',
                },
            },
            rightPriceScale: {
                borderColor: '#2D3748',
                borderVisible: true,
                scaleMargins: {
                    top: 0.1,
                    bottom: 0.2,
                },
            },
            timeScale: {
                borderColor: '#2D3748',
                borderVisible: true,
                timeVisible: true,
                secondsVisible: false,
            },
        });

        // Add candlestick series
        const candleSeries = chart.addSeries(CandlestickSeries, {
            upColor: '#10B981',
            downColor: '#EF4444',
            borderUpColor: '#10B981',
            borderDownColor: '#EF4444',
            wickUpColor: '#10B981',
            wickDownColor: '#EF4444',
            lastValueVisible: false, // Fix: Eradicate ghost labels
            priceLineVisible: false,
        });

        // Add volume series
        const volumeSeries = chart.addSeries(HistogramSeries, {
            color: '#3B82F6',
            priceFormat: {
                type: 'volume',
            },
            priceScaleId: '',
            lastValueVisible: false, // Fix: Eradicate ghost labels
            priceLineVisible: false,
        });

        chartRef.current = chart;
        candleSeriesRef.current = candleSeries;
        volumeSeriesRef.current = volumeSeries;

        // Handle resize
        const handleResize = () => {
            if (chartContainerRef.current && chartRef.current) {
                chartRef.current.applyOptions({
                    width: chartContainerRef.current.clientWidth,
                    height: chartContainerRef.current.clientHeight,
                });
            }
        };

        window.addEventListener('resize', handleResize);

        // Load initial data
        fetchHistoricalData();

        return () => {
            window.removeEventListener('resize', handleResize);
            if (chartRef.current) {
                chartRef.current.remove();
            }
        };
    }, []);

    // Handle Keyboard Shortcuts
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            // Check for Alt+T (using both key and code for robustness)
            if (e.altKey && (e.key === 't' || e.key === 'T' || e.code === 'KeyT')) {
                e.preventDefault();
                setShowOneClickTrading(prev => !prev);
            }

            // Ctrl+F4 - Close active chart
            if (e.ctrlKey && e.key === 'F4') {
                e.preventDefault();
                if (openCharts.length > 1) {
                    handleCloseChart(activeChartIndex);
                }
            }
        };

        // Listen for close-active-chart event from menu
        const handleCloseEvent = () => {
            if (openCharts.length > 1) {
                handleCloseChart(activeChartIndex);
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        window.addEventListener('close-active-chart', handleCloseEvent);

        return () => {
            window.removeEventListener('keydown', handleKeyDown);
            window.removeEventListener('close-active-chart', handleCloseEvent);
        };
    }, [openCharts.length, activeChartIndex]);



    // Update indicators
    useEffect(() => {
        if (!chartRef.current || !candleSeriesRef.current) return;

        const currentIndicators = indicatorSeriesRef.current;

        // Remove inactive indicators
        currentIndicators.forEach((series: ISeriesApi<'Line'>, name: string) => {
            if (!activeIndicators.includes(name as Indicator)) {
                chartRef.current?.removeSeries(series);
                currentIndicators.delete(name);
            }
        });

        // Add active indicators
        activeIndicators.forEach((indicator: Indicator) => {
            if (indicator === 'Volume') return; // Handled separately
            if (!currentIndicators.has(indicator)) {
                const color = indicator === 'MA' ? '#3B82F6' : '#F59E0B';
                const series = chartRef.current!.addSeries(LineSeries, {
                    color,
                    lineWidth: 2,
                    title: indicator,
                    priceScaleId: 'right',
                    lastValueVisible: false, // Fix: Eradicate ghost labels
                    priceLineVisible: false,
                });
                currentIndicators.set(indicator, series);
            }
        });
    }, [activeIndicators]);

    // Fetch historical data
    const fetchHistoricalData = async () => {
        try {
            const response = await fetch(
                buildApiUrl(`/api/history/ohlc?symbol=${selectedSymbol}&timeframe=${selectedTimeframe}&limit=500`)
            );
            const rawData = await response.json();
            const candlesArray = Array.isArray(rawData) ? rawData : (rawData.candles || []);

            if (candlesArray.length > 0 && candleSeriesRef.current && volumeSeriesRef.current) {
                const formattedCandles: CandlestickData<Time>[] = candlesArray.map((c: any) => ({
                    time: (c.time || c.timestamp) as Time,
                    open: c.open,
                    high: c.high,
                    low: c.low,
                    close: c.close,
                }));

                const formattedVolumes = candlesArray.map((c: any) => ({
                    time: (c.time || c.timestamp) as Time,
                    value: c.volume || 0,
                    color: c.close >= c.open ? '#10B98133' : '#EF444433',
                }));

                candleSeriesRef.current.setData(formattedCandles);
                volumeSeriesRef.current.setData(formattedVolumes);

                // Update price info
                const latest = candlesArray[candlesArray.length - 1];
                const first = candlesArray[0];
                const lastClose = latest.close;
                const firstOpen = first.open;

                setCurrentPrice(lastClose);
                setPriceChange(lastClose - firstOpen);
                setPriceChangePercent(((lastClose - firstOpen) / firstOpen) * 100);

                const highs = candlesArray.map((c: any) => c.high);
                const lows = candlesArray.map((c: any) => c.low);
                setHigh24h(Math.max(...highs));
                setLow24h(Math.min(...lows));

                const totalVolume = candlesArray.reduce((sum: number, c: any) => sum + (c.volume || 0), 0);
                setVolume24h(totalVolume);

                // Set the current forming candle
                formingCandleRef.current = {
                    time: (latest.time || latest.timestamp),
                    open: latest.open,
                    high: latest.high,
                    low: latest.low,
                    close: latest.close,
                    volume: latest.volume || 0
                };
            }
        } catch (error) {
            console.error('Error fetching historical data:', error);
        }
    };

    const formingCandleRef = useRef<OHLC | null>(null);

    // Use centralized WebSocket with auto-reconnection
    const { subscribe } = useWebSocket({
        url: WS_ENDPOINTS.general,
        autoConnect: true,
    });

    // WebSocket connection for live updates
    useEffect(() => {
        const unsubscribe = subscribe('*', (message: any) => {
            try {
                const data = message;

                // Handle 'tick' for current price display (Bid/Ask/Spread)
                if (data.type === 'tick' && data.symbol === selectedSymbol) {
                    const price = data.bid;
                    const ask = data.ask || price;

                    setCurrentPrice(price);
                    setBidPrice(price);
                    setAskPrice(ask);

                    // Note: Chart updates are now handled via 'candle_update' from server
                    // to ensure consistency and performance.
                }

                // Handle 'candle_update' for Chart
                if (data.type === 'candle_update' && data.symbol === selectedSymbol && data.timeframe === 'M1') {
                    // Only verify mapping for M1 current timeframe,
                    // generic handling would need timeframe check against selectedTimeframe
                    if (selectedTimeframe !== '1m') return;

                    const candleVal = {
                        time: data.time as Time,
                        open: data.open,
                        high: data.high,
                        low: data.low,
                        close: data.close,
                    };

                    const volumeVal = {
                        time: data.time as Time,
                        value: data.volume,
                        color: data.close >= data.open ? '#10B98133' : '#EF444433'
                    };

                    if (candleSeriesRef.current) {
                        candleSeriesRef.current.update(candleVal);
                    }
                    if (volumeSeriesRef.current) {
                        volumeSeriesRef.current.update(volumeVal);
                    }

                    // Update forming candle ref for other logic if needed
                    formingCandleRef.current = {
                        time: data.time,
                        open: data.open,
                        high: data.high,
                        low: data.low,
                        close: data.close,
                        volume: data.volume
                    };
                }
            } catch (error) {
                console.error('WebSocket message error:', error);
            }
        });

        return () => {
            unsubscribe();
        };
    }, [selectedSymbol, selectedTimeframe, subscribe]);

    // Refetch chart history when selected symbol changes
    useEffect(() => {
        fetchHistoricalData();
    }, [selectedSymbol, selectedTimeframe]);

    // Subscribe to ALL Market Watch symbols to ensure live updates for the panel
    useEffect(() => {
        const subscribeAll = async () => {
            console.log('Subscribing to full Market Watch list...');
            // Use the symbols array defined above
            symbols.forEach(async (sym) => {
                try {
                    await fetch(`${buildApiUrl('')}/api/symbols/subscribe`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ symbol: sym })
                    });
                } catch (err) {
                    console.error(`Failed to subscribe ${sym}:`, err);
                }
            });
        };
        subscribeAll();
    }, []);

    const toggleIndicator = (indicator: Indicator) => {
        setActiveIndicators(prev =>
            prev.includes(indicator)
                ? prev.filter(i => i !== indicator)
                : [...prev, indicator]
        );
    };

    const formatPrice = (price: number) => {
        return price.toLocaleString('en-US', {
            minimumFractionDigits: 2,
            maximumFractionDigits: 5,
        });
    };

    const formatVolume = (volume: number) => {
        if (volume >= 1_000_000_000) return `${(volume / 1_000_000_000).toFixed(2)}B`;
        if (volume >= 1_000_000) return `${(volume / 1_000_000).toFixed(2)}M`;
        if (volume >= 1_000) return `${(volume / 1_000).toFixed(2)}K`;
        return volume.toFixed(2);
    };

    const placeOrder = async (side: string) => {
        try {
            const response = await fetch(`${buildApiUrl('')}/api/orders/market`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    symbol: selectedSymbol,
                    side: side,
                    volume: tradingVolume,
                    type: 'MARKET'
                }),
            });

            const data = await response.json();

            if (response.ok) {
                console.log('Order placed successfully:', data);
                // Optional: Show success notification or updated account info
            } else {
                console.error('Order placement failed:', data);
                // Fallback to legacy endpoint if simple B-Book fails or for debugging
                if (response.status === 404) {
                    // Try legacy OANDA/A-Book endpoint
                    const legacyResponse = await fetch(`${buildApiUrl('')}/order`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            symbol: selectedSymbol,
                            side: side,
                            volume: tradingVolume,
                            type: 'MARKET'
                        })
                    });
                    if (legacyResponse.ok) console.log('Legacy Order placed');
                }
            }
        } catch (error) {
            console.error('Error placing order:', error);
        }
    };

    return (
        <div className="trading-dashboard bg-[#131722] text-[#d1d4dc] h-screen flex font-sans overflow-hidden">
            {marketWatchPosition === 'left' ? (
                <div
                    className="h-full border-r border-[#2D3748]"
                    draggable
                    onDragStart={handlePanelDragStart}
                    style={{ cursor: 'grab' }}
                >
                    <MarketWatch
                        selectedSymbol={selectedSymbol}
                        onSelectSymbol={setSelectedSymbol}
                        watchlist={symbols}
                    />
                </div>
            ) : (
                <div
                    className="w-6 hover:bg-[#2D3748]/30 transition-colors flex items-center justify-center cursor-pointer border-r border-[#2D3748]"
                    onDragOver={handleDragOver}
                    onDrop={(e) => handleDrop(e, 'left')}
                    title="Drop to dock Left"
                >
                    <div className="w-0.5 h-8 bg-[#4A5568] rounded-full" />
                </div>
            )}

            <div className="flex-1 flex flex-col min-w-0">
                {/* Top Navigation Bar */}
                <div className="top-nav">
                    <div className="nav-left">
                        {/* Symbol Selector */}
                        <div className="symbol-selector">
                            <button
                                className="symbol-button"
                                onClick={() => setShowSymbolMenu(!showSymbolMenu)}
                            >
                                <span className="symbol-name">{selectedSymbol}</span>
                                <Icons.ChevronDown />
                            </button>
                            {showSymbolMenu && (
                                <div className="dropdown-menu symbol-menu">
                                    {symbols.map(symbol => (
                                        <div
                                            key={symbol}
                                            className={`menu-item ${symbol === selectedSymbol ? 'active' : ''}`}
                                            onClick={() => {
                                                setSelectedSymbol(symbol);
                                                setShowSymbolMenu(false);
                                            }}
                                        >
                                            {symbol}
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        {/* Price Display */}
                        <div className="price-display">
                            <div className="current-price">
                                <span className="price-value">${formatPrice(currentPrice)}</span>
                                <span className={`price-change ${priceChange >= 0 ? 'positive' : 'negative'}`}>
                                    {priceChange >= 0 ? '+' : ''}{formatPrice(priceChange)} ({priceChange >= 0 ? '+' : ''}{priceChangePercent.toFixed(2)}%)
                                </span>
                            </div>
                            <div className="price-stats">
                                <span className="stat">24h High: <strong>${formatPrice(high24h)}</strong></span>
                                <span className="stat">24h Low: <strong>${formatPrice(low24h)}</strong></span>
                                <span className="stat">24h Vol: <strong>{formatVolume(volume24h)}</strong></span>
                            </div>
                        </div>
                    </div>

                    <div className="nav-center">
                        {/* Timeframe Selector */}
                        <div className="timeframe-selector">
                            {timeframes.map(tf => (
                                <button
                                    key={tf}
                                    className={`timeframe-button ${tf === selectedTimeframe ? 'active' : ''}`}
                                    onClick={() => setSelectedTimeframe(tf)}
                                >
                                    {tf}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="nav-right">
                        {/* Indicators */}
                        <div className="indicators-container">
                            <button
                                className="indicator-button"
                                onClick={() => setShowIndicatorMenu(!showIndicatorMenu)}
                            >
                                <Icons.Indicator />
                                <span>Indicators</span>
                                <Icons.ChevronDown />
                            </button>
                            {showIndicatorMenu && (
                                <div className="dropdown-menu indicator-menu">
                                    {indicators.map(indicator => (
                                        <div
                                            key={indicator}
                                            className={`menu-item ${activeIndicators.includes(indicator) ? 'active' : ''}`}
                                            onClick={() => toggleIndicator(indicator)}
                                        >
                                            <input
                                                type="checkbox"
                                                checked={activeIndicators.includes(indicator)}
                                                readOnly
                                            />
                                            <span>{indicator}</span>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        {/* Settings */}
                        <button
                            className="settings-button"
                            onClick={() => setShowSettingsMenu(!showSettingsMenu)}
                        >
                            <Icons.Settings />
                        </button>
                        {showSettingsMenu && (
                            <div className="dropdown-menu settings-menu">
                                <div className="menu-item">Chart Settings</div>
                                <div className="menu-item">Color Theme</div>
                                <div className="menu-item">Export Data</div>
                                <div className="menu-item">Screenshot</div>
                                <div className="menu-divider"></div>
                                <div
                                    className="menu-item text-red-500 hover:bg-red-500/10"
                                    onClick={async () => {
                                        if (window.confirm('Reset all local data cache? This will refresh the page.')) {
                                            try {
                                                const { ticksDB } = await import('../db/ticksDB');
                                                await ticksDB.clearAll();
                                                window.location.reload();
                                            } catch (e) {
                                                console.error('Failed to reset cache:', e);
                                                alert('Failed to reset cache');
                                            }
                                        }
                                    }}
                                >
                                    Reset Cache
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                {/* Main Content Area */}
                <div className="main-content">
                    {/* Left Toolbar */}
                    <div className="left-toolbar">
                        <div className="toolbar-section">
                            <button
                                className={`tool-button ${selectedTool === 'crosshair' ? 'active' : ''}`}
                                onClick={() => setSelectedTool('crosshair')}
                                title="Crosshair"
                            >
                                <Icons.Crosshair />
                            </button>
                            <button
                                className={`tool-button ${selectedTool === 'trendline' ? 'active' : ''}`}
                                onClick={() => setSelectedTool('trendline')}
                                title="Trend Line"
                            >
                                <Icons.TrendLine />
                            </button>
                            <button
                                className={`tool-button ${selectedTool === 'horizontal' ? 'active' : ''}`}
                                onClick={() => setSelectedTool('horizontal')}
                                title="Horizontal Line"
                            >
                                <Icons.HorizontalLine />
                            </button>
                            <button
                                className={`tool-button ${selectedTool === 'vertical' ? 'active' : ''}`}
                                onClick={() => setSelectedTool('vertical')}
                                title="Vertical Line"
                            >
                                <Icons.VerticalLine />
                            </button>
                        </div>

                        <div className="toolbar-divider"></div>

                        <div className="toolbar-section">
                            <button
                                className={`tool-button ${selectedTool === 'fibonacci' ? 'active' : ''}`}
                                onClick={() => setSelectedTool('fibonacci')}
                                title="Fibonacci Retracement"
                            >
                                <Icons.Fibonacci />
                            </button>
                            <button
                                className={`tool-button ${selectedTool === 'brush' ? 'active' : ''}`}
                                onClick={() => setSelectedTool('brush')}
                                title="Drawing Brush"
                            >
                                <Icons.Brush />
                            </button>
                            <button
                                className={`tool-button ${selectedTool === 'text' ? 'active' : ''}`}
                                onClick={() => setSelectedTool('text')}
                                title="Text Tool"
                            >
                                <Icons.Text />
                            </button>
                            <button
                                className={`tool-button ${selectedTool === 'measure' ? 'active' : ''}`}
                                onClick={() => setSelectedTool('measure')}
                                title="Measure Tool"
                            >
                                <Icons.Measure />
                            </button>
                        </div>

                        <div className="toolbar-divider"></div>

                        <div className="toolbar-section">
                            <button
                                className={`tool-button ${selectedTool === 'zoom_in' ? 'active' : ''}`}
                                onClick={() => {
                                    setSelectedTool('zoom_in');
                                    if (chartRef.current) {
                                        chartRef.current.timeScale().scrollToPosition(2, false);
                                    }
                                }}
                                title="Zoom In"
                            >
                                <Icons.ZoomIn />
                            </button>
                            <button
                                className={`tool-button ${selectedTool === 'zoom_out' ? 'active' : ''}`}
                                onClick={() => {
                                    setSelectedTool('zoom_out');
                                    if (chartRef.current) {
                                        chartRef.current.timeScale().scrollToPosition(-2, false);
                                    }
                                }}
                                title="Zoom Out"
                            >
                                <Icons.ZoomOut />
                            </button>
                        </div>
                    </div>

                    {/* Chart Area Wrapper */}
                    <div className="relative flex-1 overflow-hidden">
                        <div className="absolute inset-0" ref={chartContainerRef} style={{ zIndex: 0 }} />
                        <OneClickTrading
                            symbol={selectedSymbol}
                            bid={bidPrice || currentPrice}
                            ask={askPrice || currentPrice}
                            volume={tradingVolume}
                            onVolumeChange={setTradingVolume}
                            onBuy={() => placeOrder('BUY')}
                            onSell={() => placeOrder('SELL')}
                            isHidden={!showOneClickTrading}
                        />
                    </div>
                </div>

                {/* Active Indicators Display */}
                {activeIndicators.length > 0 && (
                    <div className="active-indicators">
                        {activeIndicators.map(indicator => (
                            <span key={indicator} className="indicator-tag">
                                {indicator}
                                <button
                                    className="remove-indicator"
                                    onClick={() => toggleIndicator(indicator)}
                                >
                                    ×
                                </button>
                            </span>
                        ))}
                    </div>
                )}

                {/* Chart Tabs Bar */}
                <div className="h-8 bg-[#131722] border-t border-[#2D3748] flex items-center px-1 select-none">
                    {openCharts.map((chart, index) => (
                        <div
                            key={`${chart.symbol}-${index}`}
                            className={`
                                group flex items-center gap-2 px-3 h-full cursor-pointer border-r border-[#2D3748] text-xs font-medium transition-colors
                                ${index === activeChartIndex
                                    ? 'bg-[#1E293B] text-[#3B82F6] border-b-2 border-b-[#3B82F6]'
                                    : 'text-[#718096] hover:bg-[#1A202C] hover:text-[#94A3B8]'}
                            `}
                            onClick={() => setActiveChartIndex(index)}
                        >
                            <span>{chart.symbol}, {chart.timeframe}</span>
                            <span
                                className="opacity-0 group-hover:opacity-100 hover:text-red-400 p-0.5 rounded"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    handleCloseChart(index);
                                }}
                            >
                                ×
                            </span>
                        </div>
                    ))}
                    <div
                        className="px-3 h-full flex items-center justify-center text-[#718096] hover:text-white cursor-pointer hover:bg-[#1A202C]"
                        onClick={() => setSelectedSymbol('EURUSD')}
                        title="Open New Chart"
                    >
                        +
                    </div>
                </div>
            </div>
            {marketWatchPosition === 'right' ? (
                <div
                    className="h-full border-l border-[#2D3748]"
                    draggable
                    onDragStart={handlePanelDragStart}
                    style={{ cursor: 'grab' }}
                >
                    <MarketWatch
                        selectedSymbol={selectedSymbol}
                        onSelectSymbol={setSelectedSymbol}
                        watchlist={symbols}
                    />
                </div>
            ) : (
                <div
                    className="w-6 hover:bg-[#2D3748]/30 transition-colors flex items-center justify-center cursor-pointer border-l border-[#2D3748]"
                    onDragOver={handleDragOver}
                    onDrop={(e) => handleDrop(e, 'right')}
                    title="Drop to dock Right"
                >
                    <div className="w-0.5 h-8 bg-[#4A5568] rounded-full" />
                </div>
            )}
        </div>
    );
};

export default ProfessionalTradingDashboard;

