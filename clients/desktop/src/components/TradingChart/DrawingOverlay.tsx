import { useEffect, useState, useCallback } from 'react';
import type { IChartApi, ISeriesApi } from 'lightweight-charts';
import type { Drawing, Point } from './types';
import { Trash2 } from 'lucide-react';

interface DrawingOverlayProps {
    chart: IChartApi | null;
    series: ISeriesApi<any> | null;
    drawings: Drawing[];
    selectedTool: string;
    symbol: string;
    accountId: number;
    onUpdateDrawing: (drawing: Drawing) => void;
    onDeleteDrawing: (id: string) => void;
    onFinishDrawing: () => void; // Called when a drawing is completed so tool can reset if needed
    containerRef: React.RefObject<HTMLDivElement>;
}

export function DrawingOverlay({
    chart,
    series,
    drawings,
    selectedTool,
    symbol,
    accountId,
    onUpdateDrawing,
    onDeleteDrawing,
    onFinishDrawing,
    containerRef
}: DrawingOverlayProps) {
    const [activeDrawing, setActiveDrawing] = useState<Drawing | null>(null);
    const [hoveredDrawingId, setHoveredDrawingId] = useState<string | null>(null);
    const [mousePos, setMousePos] = useState<{ x: number, y: number } | null>(null);

    // Coordinate conversion helpers
    const timeToX = useCallback((time: number) => {
        if (!chart) return null;
        return chart.timeScale().timeToCoordinate(time as any);
    }, [chart]);

    const priceToY = useCallback((price: number) => {
        if (!series) return null;
        return series.priceToCoordinate(price);
    }, [series]);

    const xToTime = useCallback((x: number) => {
        if (!chart) return null;
        return chart.timeScale().coordinateToTime(x) as number;
    }, [chart]);

    const yToPrice = useCallback((y: number) => {
        if (!series) return null;
        return series.coordinateToPrice(y);
    }, [series]);

    // Handle Mouse Interactions
    useEffect(() => {
        if (!containerRef.current || !chart || !series) return;
        const container = containerRef.current;

        const handleMouseDown = (e: MouseEvent) => {
            if (selectedTool === 'cursor') return;

            const rect = container.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const y = e.clientY - rect.top;

            const time = xToTime(x);
            const price = yToPrice(y);

            if (!time || !price) return;

            if (!activeDrawing) {
                // Start a new drawing
                const newId = `draw_${Date.now()}`;
                const newDrawing: Drawing = {
                    id: newId,
                    accountId,
                    symbol,
                    type: selectedTool as any,
                    points: [{ time, price }],
                    options: { color: '#3b82f6' }
                };

                // For single-point tools (like horizontal line potentially), we might finish immediately or drag
                // But for now let's assume click-click or drag mechanics.
                // Let's implement Click-to-Start, Click-to-Finish for Line/Rect.
                setActiveDrawing(newDrawing);
            } else {
                // Finish drawing
                const finishDrawing = {
                    ...activeDrawing,
                    points: [...activeDrawing.points, { time, price }]
                };
                onUpdateDrawing(finishDrawing);
                setActiveDrawing(null);
                onFinishDrawing();
            }
        };

        const handleMouseMove = (e: MouseEvent) => {
            const rect = container.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const y = e.clientY - rect.top;
            setMousePos({ x, y });
        };

        container.addEventListener('mousedown', handleMouseDown);
        container.addEventListener('mousemove', handleMouseMove);

        return () => {
            container.removeEventListener('mousedown', handleMouseDown);
            container.removeEventListener('mousemove', handleMouseMove);
        };
    }, [chart, series, selectedTool, activeDrawing, xToTime, yToPrice, symbol, accountId, onUpdateDrawing, onFinishDrawing, containerRef]);


    // -- RENDERERS --

    const renderLine = (p1: Point, p2: Point, color: string, isPreview = false) => {
        const x1 = timeToX(p1.time);
        const y1 = priceToY(p1.price);
        const x2 = timeToX(p2.time);
        const y2 = priceToY(p2.price);

        if (x1 === null || y1 === null || x2 === null || y2 === null) return null;

        return (
            <line
                x1={x1} y1={y1} x2={x2} y2={y2}
                stroke={color}
                strokeWidth={2}
                strokeDasharray={isPreview ? "5,5" : undefined}
            />
        );
    };

    const renderRect = (p1: Point, p2: Point, color: string, isPreview = false) => {
        const x1 = timeToX(p1.time);
        const y1 = priceToY(p1.price);
        const x2 = timeToX(p2.time);
        const y2 = priceToY(p2.price);

        if (x1 === null || y1 === null || x2 === null || y2 === null) return null;

        const width = x2 - x1;
        const height = y2 - y1;

        return (
            <rect
                x={Math.min(x1, x2)}
                y={Math.min(y1, y2)}
                width={Math.abs(width)}
                height={Math.abs(height)}
                fill={color}
                fillOpacity={0.2}
                stroke={color}
                strokeWidth={1}
                strokeDasharray={isPreview ? "5,5" : undefined}
            />
        );
    };

    const renderHorizontalLine = (p1: Point, color: string) => {
        const y1 = priceToY(p1.price);
        if (y1 === null) return null;

        return (
            <line
                x1={0} y1={y1} x2="100%" y2={y1}
                stroke={color}
                strokeWidth={1}
            />
        );
    };

    // Render Logic
    return (
        <svg className="absolute inset-0 pointer-events-none w-full h-full z-10">
            {/* Render Saved Drawings */}
            {drawings.map(d => {
                const isHovered = d.id === hoveredDrawingId;
                const color = isHovered ? '#60a5fa' : (d.options?.color || '#3b82f6');

                if (d.type === 'line' && d.points.length >= 2) {
                    return <g key={d.id}>{renderLine(d.points[0], d.points[1], color)}</g>;
                }
                if (d.type === 'rectangle' && d.points.length >= 2) {
                    return <g key={d.id}>{renderRect(d.points[0], d.points[1], color)}</g>;
                }
                if (d.type === 'horizontal_line' && d.points.length >= 1) {
                    return <g key={d.id}>{renderHorizontalLine(d.points[0], color)}</g>;
                }
                return null;
            })}

            {/* Render Active Drawing (Preview) */}
            {activeDrawing && mousePos && (
                <g>
                    {activeDrawing.type === 'line' && activeDrawing.points[0] && (
                        (() => {
                            const p1 = activeDrawing.points[0];
                            // We need to map mousePos back to time/price to keep logic consistent
                            // Or just use mousePos x/y directly for the preview endpoint
                            // Let's use mouse coordinates for the second point visually
                            const x1 = timeToX(p1.time);
                            const y1 = priceToY(p1.price);
                            if (x1 && y1) {
                                return <line x1={x1} y1={y1} x2={mousePos.x} y2={mousePos.y} stroke="#3b82f6" strokeWidth={2} strokeDasharray="5,5" />;
                            }
                        })()
                    )}
                    {activeDrawing.type === 'rectangle' && activeDrawing.points[0] && (
                        (() => {
                            const p1 = activeDrawing.points[0];
                            const x1 = timeToX(p1.time);
                            const y1 = priceToY(p1.price);
                            if (x1 && y1) {
                                const width = mousePos.x - x1;
                                const height = mousePos.y - y1;
                                return (
                                    <rect
                                        x={Math.min(x1, mousePos.x)}
                                        y={Math.min(y1, mousePos.y)}
                                        width={Math.abs(width)}
                                        height={Math.abs(height)}
                                        fill="#3b82f6" fillOpacity={0.2} stroke="#3b82f6" strokeWidth={1} strokeDasharray="5,5"
                                    />
                                );
                            }
                        })()
                    )}
                </g>
            )}
        </svg>
    );
}
