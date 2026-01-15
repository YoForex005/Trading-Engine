import { useRef, useEffect } from 'react';
import {
  createChart,
  ColorType,
  CrosshairMode,
  CandlestickSeries,
  LineSeries,
  BarSeries,
  AreaSeries,
} from 'lightweight-charts';
import type { IChartApi, ISeriesApi } from 'lightweight-charts';
import type { OHLCData, ChartType } from '../types';

type ChartCanvasProps = {
  data: OHLCData[];
  chartType: ChartType;
  onChartReady?: (chart: IChartApi, series: ISeriesApi<any>) => void;
};

export function ChartCanvas({ data, chartType, onChartReady }: ChartCanvasProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const seriesRef = useRef<ISeriesApi<any> | null>(null);

  // Initialize chart once
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

      return () => {
        window.removeEventListener('resize', handleResize);
        chart.remove();
        chartRef.current = null;
        seriesRef.current = null;
      };
    } catch (err) {
      console.error('Failed to initialize chart:', err);
    }
  }, []);

  // Create/update series when chartType changes
  useEffect(() => {
    if (!chartRef.current) return;

    try {
      // Remove old series
      if (seriesRef.current) {
        try {
          chartRef.current.removeSeries(seriesRef.current);
        } catch {
          // Ignore errors
        }
        seriesRef.current = null;
      }

      let series: any;
      const chart = chartRef.current;

      switch (chartType) {
        case 'candlestick':
        case 'heikinAshi':
          series = chart.addSeries(CandlestickSeries, {
            upColor: '#10b981',
            downColor: '#ef4444',
            borderUpColor: '#10b981',
            borderDownColor: '#ef4444',
            wickUpColor: '#10b981',
            wickDownColor: '#ef4444',
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
            lineColor: '#10b981',
            topColor: 'rgba(16, 185, 129, 0.4)',
            bottomColor: 'rgba(16, 185, 129, 0.0)',
            lineWidth: 2,
          });
          break;
        default:
          series = chart.addSeries(CandlestickSeries, {
            upColor: '#10b981',
            downColor: '#ef4444',
            borderUpColor: '#10b981',
            borderDownColor: '#ef4444',
            wickUpColor: '#10b981',
            wickDownColor: '#ef4444',
          });
      }

      seriesRef.current = series;

      // Notify parent that chart is ready
      if (onChartReady) {
        onChartReady(chart, series);
      }
    } catch (err) {
      console.error('Error creating chart series:', err);
    }
  }, [chartType, onChartReady]);

  // Update data when it changes
  useEffect(() => {
    if (!seriesRef.current || data.length === 0) return;

    try {
      seriesRef.current.setData(data);
    } catch (err) {
      console.error('Error setting chart data:', err);
    }
  }, [data]);

  return (
    <div
      ref={chartContainerRef}
      className="w-full h-full"
      style={{ position: 'relative' }}
    />
  );
}
