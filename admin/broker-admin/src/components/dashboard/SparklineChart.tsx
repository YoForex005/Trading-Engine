'use client';

import React, { useMemo } from 'react';

interface SparklineChartProps {
    data: number[];
    width?: number;
    height?: number;
    className?: string;
}

export default function SparklineChart({
    data,
    width = 80,
    height = 30,
    className = '',
}: SparklineChartProps) {
    // Calculate trend direction (compare first and last values)
    const trend = useMemo(() => {
        if (data.length < 2) return 'neutral';
        const firstValue = data[0];
        const lastValue = data[data.length - 1];
        return lastValue > firstValue ? 'up' : lastValue < firstValue ? 'down' : 'neutral';
    }, [data]);

    // Calculate path data for SVG polyline
    const pathData = useMemo(() => {
        if (data.length === 0) return '';

        const min = Math.min(...data);
        const max = Math.max(...data);
        const range = max - min || 1; // Avoid division by zero

        const points = data.map((value, index) => {
            const x = (index / (data.length - 1 || 1)) * width;
            const y = height - ((value - min) / range) * height;
            return `${x},${y}`;
        });

        return points.join(' ');
    }, [data, width, height]);

    // Determine color based on trend
    const color = trend === 'up' ? '#10B981' : trend === 'down' ? '#EF4444' : '#6B7280';

    if (data.length === 0) {
        return (
            <svg
                width={width}
                height={height}
                className={className}
                style={{ display: 'block' }}
            >
                <line
                    x1="0"
                    y1={height / 2}
                    x2={width}
                    y2={height / 2}
                    stroke="#374151"
                    strokeWidth="1"
                    strokeDasharray="2,2"
                />
            </svg>
        );
    }

    return (
        <svg
            width={width}
            height={height}
            className={className}
            style={{ display: 'block' }}
        >
            <polyline
                points={pathData}
                fill="none"
                stroke={color}
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </svg>
    );
}
