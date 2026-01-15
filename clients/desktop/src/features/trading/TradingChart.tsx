import { useState, useRef } from 'react';
import type { IChartApi, ISeriesApi } from 'lightweight-charts';
import { Maximize2, Minimize2 } from 'lucide-react';
import { ChartCanvas } from './components/ChartCanvas';
import { DrawingTools } from './components/DrawingTools';
import { DrawingOverlay } from './components/DrawingOverlay';
import { IndicatorPane } from './components/IndicatorPane';
import { useChartData, useDrawings, useIndicators } from './hooks';
import { IndicatorStorage } from '@/services/IndicatorStorage';
import type { ChartType, Timeframe, Drawing } from './types';
import type { DrawingType } from './components/DrawingTools';

type TradingChartProps = {
  symbol: string;
  currentPrice?: { bid: number; ask: number };
  chartType?: ChartType;
  timeframe?: Timeframe;
  positions?: any[];
  onClosePosition?: (id: number) => void;
  onModifyPosition?: (id: number, sl: number, tp: number) => void;
  onPlacePendingOrder?: (price: number, type: 'LIMIT' | 'STOP') => void;
  oneClickEnabled?: boolean;
  accountId?: number;
};

export function TradingChart({
  symbol,
  currentPrice,
  chartType = 'candlestick',
  timeframe = '1m',
  positions = [],
  onClosePosition,
  onModifyPosition,
  onPlacePendingOrder,
  oneClickEnabled,
  accountId = 1,
}: TradingChartProps) {
  // Chart references
  const chartRef = useRef<IChartApi | null>(null);
  const seriesRef = useRef<ISeriesApi<any> | null>(null);

  // Drawing state
  const [activeTool, setActiveTool] = useState<DrawingType>('cursor');
  const [isFullscreen, setIsFullscreen] = useState(false);

  // Custom hooks for data management
  const { ohlc, loading: chartLoading, error: chartError } = useChartData(symbol, timeframe);
  const {
    drawings,
    loading: drawingsLoading,
    updateDrawing,
    deleteDrawing,
  } = useDrawings(symbol, accountId);

  // Convert OHLC to indicator format
  const indicatorOHLCData = ohlc.map((c) => ({
    time: typeof c.time === 'string' ? parseInt(c.time) : (c.time as number),
    open: c.open,
    high: c.high,
    low: c.low,
    close: c.close,
  }));

  // Indicators hook
  const {
    indicators,
    overlayIndicators,
    paneGroups,
    addIndicator,
    removeIndicator,
    updateIndicator,
    toggleIndicator,
  } = useIndicators({
    ohlcData: indicatorOHLCData,
    autoCalculate: true,
  });

  // Handle chart ready callback
  const handleChartReady = (chart: IChartApi, series: ISeriesApi<any>) => {
    chartRef.current = chart;
    seriesRef.current = series;
  };

  // Handle drawing updates
  const handleUpdateDrawing = async (drawing: Drawing) => {
    try {
      await updateDrawing(drawing);
    } catch (err) {
      console.error('Failed to update drawing:', err);
    }
  };

  const handleDeleteDrawing = async (id: string) => {
    try {
      await deleteDrawing(id);
    } catch (err) {
      console.error('Failed to delete drawing:', err);
    }
  };

  // Toggle fullscreen
  const toggleFullscreen = () => {
    setIsFullscreen((prev) => !prev);
  };

  // Loading state
  if (chartLoading) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-zinc-900 text-zinc-400">
        Loading chart data...
      </div>
    );
  }

  // Error state
  if (chartError) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-zinc-900 text-red-400">
        Error loading chart: {chartError.message}
      </div>
    );
  }

  return (
    <div
      className={`relative bg-zinc-900 ${
        isFullscreen ? 'fixed inset-0 z-50' : 'w-full h-full'
      }`}
    >
      {/* Toolbar */}
      <div className="absolute top-2 left-2 z-10 flex gap-2">
        <DrawingTools activeTool={activeTool} onToolChange={setActiveTool} />
        <button
          onClick={toggleFullscreen}
          className="p-2 bg-zinc-800 hover:bg-zinc-700 rounded text-zinc-300"
          title={isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'}
        >
          {isFullscreen ? <Minimize2 size={16} /> : <Maximize2 size={16} />}
        </button>
      </div>

      {/* Main chart canvas */}
      <ChartCanvas data={ohlc} chartType={chartType} onChartReady={handleChartReady} />

      {/* Drawing overlay */}
      {chartRef.current && (
        <DrawingOverlay
          chart={chartRef.current}
          activeTool={activeTool}
          drawings={drawings}
          onUpdateDrawing={handleUpdateDrawing}
          onDeleteDrawing={handleDeleteDrawing}
        />
      )}

      {/* Indicator panes */}
      <div className="absolute bottom-0 left-0 right-0">
        {Array.from(paneGroups.entries()).map(([paneIndex, paneIndicators]) => (
          <IndicatorPane
            key={paneIndex}
            indicators={paneIndicators}
            onToggle={toggleIndicator}
            onRemove={removeIndicator}
            onUpdate={updateIndicator}
          />
        ))}
      </div>

      {/* Symbol and price info */}
      <div className="absolute top-2 right-2 z-10 bg-zinc-800/90 px-3 py-2 rounded text-xs">
        <div className="font-semibold text-zinc-100">{symbol}</div>
        {currentPrice && (
          <div className="flex gap-3 mt-1">
            <span className="text-green-400">Bid: {currentPrice.bid.toFixed(5)}</span>
            <span className="text-red-400">Ask: {currentPrice.ask.toFixed(5)}</span>
          </div>
        )}
      </div>
    </div>
  );
}
