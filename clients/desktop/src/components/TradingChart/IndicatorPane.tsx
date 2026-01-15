import { useEffect, useRef } from 'react';
import {
  createChart,
  ColorType,
  CrosshairMode,
  LineSeries,
} from 'lightweight-charts';
import type { IChartApi, ISeriesApi, Time } from 'lightweight-charts';
import type { CalculatedIndicator } from '../../hooks/useIndicators';

interface IndicatorPaneProps {
  indicators: CalculatedIndicator[];
  height: number;
  mainChartRef?: IChartApi | null;
}

export function IndicatorPane({
  indicators,
  height,
  mainChartRef,
}: IndicatorPaneProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const seriesMapRef = useRef<Map<string, ISeriesApi<'Line'>>>(new Map());

  // Initialize lightweight-charts for indicator pane
  useEffect(() => {
    if (!containerRef.current) return;

    try {
      const chart = createChart(containerRef.current, {
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
          vertLine: {
            color: '#525252',
            width: 1,
            style: 2,
            labelBackgroundColor: '#18181b',
          },
          horzLine: {
            color: '#525252',
            width: 1,
            style: 2,
            labelBackgroundColor: '#18181b',
          },
        },
        rightPriceScale: {
          borderColor: '#27272a',
          scaleMargins: { top: 0.1, bottom: 0.1 },
        },
        timeScale: {
          borderColor: '#27272a',
          timeVisible: true,
          secondsVisible: false,
        },
        handleScroll: {
          mouseWheel: true,
          pressedMouseMove: true,
          horzTouchDrag: true,
          vertTouchDrag: false,
        },
        handleScale: {
          axisPressedMouseMove: true,
          mouseWheel: true,
          pinch: true,
        },
      });

      chartRef.current = chart;

      // Synchronize time scale with main chart
      if (mainChartRef) {
        const mainTimeScale = mainChartRef.timeScale();
        const paneTimeScale = chart.timeScale();

        // Subscribe to main chart changes and sync to pane
        const handleMainTimeScaleChange = () => {
          const visibleRange = mainTimeScale.getVisibleRange();
          if (visibleRange) {
            paneTimeScale.setVisibleRange(visibleRange);
          }
        };

        // Subscribe to pane changes and sync to main
        const handlePaneTimeScaleChange = () => {
          const visibleRange = paneTimeScale.getVisibleRange();
          if (visibleRange) {
            mainTimeScale.setVisibleRange(visibleRange);
          }
        };

        mainTimeScale.subscribeVisibleTimeRangeChange(handleMainTimeScaleChange);
        paneTimeScale.subscribeVisibleTimeRangeChange(handlePaneTimeScaleChange);

        return () => {
          mainTimeScale.unsubscribeVisibleTimeRangeChange(handleMainTimeScaleChange);
          paneTimeScale.unsubscribeVisibleTimeRangeChange(handlePaneTimeScaleChange);
        };
      }
    } catch (err) {
      console.error('Failed to initialize indicator pane:', err);
    }
  }, [mainChartRef]);

  // Handle window resize
  useEffect(() => {
    if (!chartRef.current || !containerRef.current) return;

    const handleResize = () => {
      if (containerRef.current && chartRef.current) {
        chartRef.current.applyOptions({
          width: containerRef.current.clientWidth,
          height,
        });
      }
    };

    handleResize();
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [height]);

  // Render indicator series
  useEffect(() => {
    if (!chartRef.current) return;

    const chart = chartRef.current;
    const seriesMap = seriesMapRef.current;

    // Get current indicator IDs
    const currentIds = new Set(indicators.map(ind => ind.id));

    // Remove series for indicators no longer visible
    for (const [id, series] of seriesMap.entries()) {
      if (!currentIds.has(id)) {
        try {
          chart.removeSeries(series);
          seriesMap.delete(id);
        } catch (err) {
          console.error('Failed to remove series:', err);
        }
      }
    }

    // Add or update series for visible indicators
    for (const indicator of indicators) {
      if (!indicator.visible) continue;

      try {
        let series = seriesMap.get(indicator.id);

        // Create new series if needed
        if (!series) {
          series = chart.addSeries(LineSeries, {
            color: indicator.color,
            lineWidth: indicator.lineWidth as 1 | 2 | 3 | 4,
          });
          seriesMap.set(indicator.id, series);
        }

        // Update data
        if (Array.isArray(indicator.data) && indicator.data.length > 0) {
          const firstValue = indicator.data[0].value;

          // Handle single-value indicators (e.g., RSI, Stochastic %K)
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

    // Fit chart to content
    if (indicators.length > 0) {
      try {
        chart.timeScale().fitContent();
      } catch (err) {
        console.error('Failed to fit content:', err);
      }
    }
  }, [indicators]);

  if (indicators.length === 0) {
    return null;
  }

  return (
    <div
      ref={containerRef}
      className="relative bg-zinc-900 border-t border-zinc-800 w-full"
      style={{ height }}
    >
      {/* Indicator Labels */}
      <div className="absolute top-2 left-4 z-10 pointer-events-none">
        <div className="text-xs text-zinc-500 font-medium">
          {indicators.map((ind) => ind.type).join(', ')}
        </div>
      </div>
    </div>
  );
}
