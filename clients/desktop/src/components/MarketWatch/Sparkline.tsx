import { useMemo } from 'react';

interface SparklineProps {
    data: number[];
    width?: number;
    height?: number;
    color?: string;
    strokeWidth?: number;
}

export function Sparkline({ data, width = 60, height = 20, color = '#10b981', strokeWidth = 1.5 }: SparklineProps) {
    const points = useMemo(() => {
        if (!data || data.length < 2) return '';

        const min = Math.min(...data);
        const max = Math.max(...data);
        const range = max - min || 1; // Avoid division by zero

        const stepX = width / (data.length - 1);

        return data
            .map((val, index) => {
                const x = index * stepX;
                const normalizedY = (val - min) / range;
                const y = height - normalizedY * height; // Invert Y because SVG 0 is top
                return `${x},${y}`;
            })
            .join(' ');
    }, [data, width, height]);

    if (!data || data.length < 2) return <div style={{ width, height }} />;

    return (
        <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`} className="overflow-visible">
            <polyline
                points={points}
                fill="none"
                stroke={color}
                strokeWidth={strokeWidth}
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </svg>
    );
}
