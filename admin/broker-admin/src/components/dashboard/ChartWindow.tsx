'use client';

import React, { useState, useEffect, useRef } from 'react';
import { X, ChevronDown } from 'lucide-react';
import { API_CONFIG, TIMEFRAMES, Timeframe } from '@/config/api';
import { useWebSocket, MarketTick } from '@/hooks/useWebSocket';

interface ChartWindowProps {
    symbol: string;
    timeframe?: Timeframe;
    isActive?: boolean;
    onClose?: () => void;
    onFocus?: () => void;
}

export default function ChartWindow({
    symbol,
    timeframe: initialTimeframe = 'M5',
    isActive = false,
    onClose,
    onFocus
}: ChartWindowProps) {
    const [timeframe, setTimeframe] = useState<Timeframe>(initialTimeframe);
    const [showTimeframeMenu, setShowTimeframeMenu] = useState(false);
    const [price, setPrice] = useState(0);
    const [ticks, setTicks] = useState<number[]>([]);
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const animFrameRef = useRef<number | null>(null);

    // Handle incoming WebSocket ticks
    const handleTick = (tick: MarketTick) => {
        if (tick.symbol === symbol) {
            const midPrice = (tick.bid + tick.ask) / 2;
            setPrice(midPrice);
            setTicks(prev => {
                const updated = [...prev, midPrice];
                // Keep last 100 ticks
                if (updated.length > 100) updated.shift();
                return updated;
            });
        }
    };

    // Connect to WebSocket
    const { isConnected } = useWebSocket({
        url: API_CONFIG.MARKET_WS_URL,
        onMessage: handleTick,
    });

    // Load initial OHLC data
    useEffect(() => {
        async function fetchOHLC() {
            try {
                // Import api dynamically to avoid circular dependencies
                const { api } = await import('@/services/apiClient');
                const data = await api.get(API_CONFIG.OHLC_DATA(symbol, timeframe));
                if (data && data.length > 0) {
                    // Initialize ticks from OHLC data (use close prices)
                    const closePrices = data.slice(-100).map((candle: any) => candle.close);
                    setTicks(closePrices);
                    setPrice(closePrices[closePrices.length - 1]);
                }
            } catch (e) {
                console.error('Failed to fetch OHLC data:', e);
            }
        }
        fetchOHLC();
    }, [symbol, timeframe]);

    // Canvas drawing
    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        const draw = () => {
            if (!ctx || !canvas) return;

            // Clear canvas
            ctx.fillStyle = '#0A0A0B';
            ctx.fillRect(0, 0, canvas.width, canvas.height);

            // Draw grid
            ctx.strokeStyle = '#1A1A1C';
            ctx.lineWidth = 1;
            ctx.beginPath();
            for (let x = 0; x < canvas.width; x += 40) {
                ctx.moveTo(x, 0);
                ctx.lineTo(x, canvas.height);
            }
            for (let y = 0; y < canvas.height; y += 40) {
                ctx.moveTo(0, y);
                ctx.lineTo(canvas.width, y);
            }
            ctx.stroke();

            // Draw ticks if we have data
            if (ticks.length > 1) {
                const min = Math.min(...ticks);
                const max = Math.max(...ticks);
                const range = max - min || 1;

                ctx.strokeStyle = '#3B82F6';
                ctx.lineWidth = 2;
                ctx.beginPath();

                ticks.forEach((t, i) => {
                    const x = (i / Math.max(ticks.length - 1, 1)) * canvas.width;
                    const y = canvas.height - ((t - min) / range) * (canvas.height - 20);
                    if (i === 0) ctx.moveTo(x, y);
                    else ctx.lineTo(x, y);
                });
                ctx.stroke();

                // Draw current price line
                if (ticks.length > 0) {
                    const lastPrice = ticks[ticks.length - 1];
                    const lastY = canvas.height - ((lastPrice - min) / range) * (canvas.height - 20);

                    ctx.strokeStyle = '#F5C542';
                    ctx.setLineDash([4, 4]);
                    ctx.beginPath();
                    ctx.moveTo(0, lastY);
                    ctx.lineTo(canvas.width, lastY);
                    ctx.stroke();
                    ctx.setLineDash([]);

                    // Price label
                    ctx.fillStyle = '#F5C542';
                    ctx.font = '11px monospace';
                    ctx.fillText(lastPrice.toFixed(5), canvas.width - 70, lastY - 5);
                }
            }

            animFrameRef.current = requestAnimationFrame(draw);
        };

        draw();

        return () => {
            if (animFrameRef.current) {
                cancelAnimationFrame(animFrameRef.current);
            }
        };
    }, [ticks]);

    return (
        <div
            className={`relative bg-[#0A0A0B] border ${isActive ? 'border-[#F5C542]' : 'border-[#383A42]'} flex flex-col h-full`}
            onClick={onFocus}
        >
            {/* Title Bar */}
            <div className="h-6 bg-[#1E2026] border-b border-[#383A42] flex items-center justify-between px-2 select-none">
                <div className="flex items-center gap-2">
                    <span className="text-white text-xs font-bold">{symbol}</span>
                    <div className="relative">
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                setShowTimeframeMenu(!showTimeframeMenu);
                            }}
                            className="flex items-center gap-1 text-[#888] hover:text-white text-[10px]"
                        >
                            {timeframe}
                            <ChevronDown size={10} />
                        </button>
                        {showTimeframeMenu && (
                            <div className="absolute top-full left-0 mt-1 bg-[#1E1E1E] border border-[#444] z-50 min-w-[60px]">
                                {TIMEFRAMES.map(tf => (
                                    <div
                                        key={tf.value}
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            setTimeframe(tf.value);
                                            setShowTimeframeMenu(false);
                                        }}
                                        className={`px-3 py-1 text-xs cursor-pointer hover:bg-[#3399FF] hover:text-white ${
                                            timeframe === tf.value ? 'bg-[#3399FF] text-white' : 'text-[#CCC]'
                                        }`}
                                    >
                                        {tf.label}
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                    <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-[#2ECC71]' : 'bg-[#E74C3C]'}`}
                         title={isConnected ? 'Connected' : 'Disconnected'} />
                </div>
                {onClose && (
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            onClose();
                        }}
                        className="text-[#888] hover:text-white"
                    >
                        <X size={12} />
                    </button>
                )}
            </div>

            {/* Chart Canvas */}
            <div className="flex-1 relative">
                <canvas
                    ref={canvasRef}
                    width={800}
                    height={600}
                    className="w-full h-full"
                />
                {ticks.length === 0 && (
                    <div className="absolute inset-0 flex items-center justify-center text-[#666] text-xs">
                        Loading chart data...
                    </div>
                )}
            </div>

            {/* Time axis (simplified) */}
            <div className="h-4 bg-[#0A0A0B] border-t border-[#383A42] flex justify-between px-2 text-[9px] text-[#555]">
                <span>{new Date().toLocaleDateString()}</span>
                <span>{new Date().toLocaleTimeString()}</span>
            </div>
        </div>
    );
}
