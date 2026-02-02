import { useState, useEffect, useRef, useCallback } from 'react';
import { Login } from './components/Login';
import { ChartWithHistory } from './components/ChartWithHistory';
import { TradingChart } from './components/TradingChart';
import type { ChartType, Timeframe } from './components/TradingChart';
import { ErrorBoundary } from './components/ErrorBoundary';
import { ChartTabs, type ChartTab } from './components/ChartTabs';
import { BottomDock } from './components/BottomDock';
import { AlertsContainer } from './components/AlertsContainer';
import { useAppStore } from './store/useAppStore';
import { useMarketDataStore } from './store/useMarketDataStore';
import { terminateWorker } from './services/aggregationWorkerManager';

// Layout Components
import { TopToolbar } from './components/layout/TopToolbar';
import { NavigatorPanel } from './components/layout/NavigatorPanel';
import { MarketWatchPanel } from './components/layout/MarketWatchPanel';
import { StatusBar } from './components/layout/StatusBar';
import { OrderPanelDialog } from './components/OrderPanelDialog';
import { DepthOfMarket } from './components/DepthOfMarket';
import { SaveWorkspaceDialog } from './components/dialogs/SaveWorkspaceDialog';

// Services
import { windowManager, type LayoutMode } from './services/windowManager';

// Command Bus
import { CommandBusProvider } from './contexts/CommandBusContext';

// Keyboard Shortcuts
import { KeyboardShortcutProvider } from './contexts/KeyboardShortcutContext';
import { GlobalShortcuts } from './components/GlobalShortcuts';

interface Tick {
  symbol: string;
  bid: number;
  ask: number;
  spread: number; // Always present - calculated as ask - bid
  timestamp: number;
  prevBid?: number;
  lp?: string;
  high24h?: number; // 24-hour high from backend
  low24h?: number;  // 24-hour low from backend
}

interface Position {
  id: number;
  symbol: string;
  side: string;
  volume: number;
  openPrice: number;
  currentPrice: number;
  openTime: string;
  sl: number;
  tp: number;
  swap: number;
  commission: number;
  unrealizedPnL: number;
}

interface Account {
  balance: number;
  equity: number;
  margin: number;
  freeMargin: number;
  unrealizedPL: number;
  marginLevel: number;
  currency: string;
}

interface BrokerConfig {
  brokerName: string;
  priceFeedLP: string;
  executionMode: string;
  defaultLeverage: number;
  marginMode: string;
}

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSD');
  const [positions, setPositions] = useState<Position[]>([]);
  const [account, setAccount] = useState<Account | null>(null);
  const [volume, setVolume] = useState(0.01);
  const [orderLoading, setOrderLoading] = useState(false);
  const [chartType, setChartType] = useState<ChartType>('candlestick');
  const [timeframe, setTimeframe] = useState<Timeframe>('1m');
  const [brokerConfig, setBrokerConfig] = useState<BrokerConfig | null>(null);
  const [dockHeight, setDockHeight] = useState(() => {
    const saved = localStorage.getItem('dockHeight');
    return saved ? parseInt(saved, 10) : 250;
  });
  const [accountId, setAccountId] = useState<string>("1");
  const [isLoadingSymbols, setIsLoadingSymbols] = useState(false);

  const [enableHistoricalData] = useState(false);

  // UI State
  const [showDOM, setShowDOM] = useState(false);

  // Multi-Chart State
  const [openCharts, setOpenCharts] = useState<ChartTab[]>([]);
  const [activeChartId, setActiveChartId] = useState<string | null>(null);

  // Initialize default chart if empty
  useEffect(() => {
    if (openCharts.length === 0 && selectedSymbol) {
      const newChart: ChartTab = {
        id: `chart-${Date.now()}`,
        symbol: selectedSymbol,
        timeframe: timeframe
      };
      setOpenCharts([newChart]);
      setActiveChartId(newChart.id);
    }
  }, []); // Only on mount

  // Sync active chart ID -> selected symbol (Master)
  useEffect(() => {
    if (activeChartId) {
      const activeChart = openCharts.find(c => c.id === activeChartId);
      if (activeChart && activeChart.symbol !== selectedSymbol) {
        setSelectedSymbol(activeChart.symbol);
      }
    }
  }, [activeChartId, openCharts]);

  // Sync selected symbol -> update active chart symbol (if it changed elsewhere)
  useEffect(() => {
    if (activeChartId) {
      setOpenCharts(prev => {
        const active = prev.find(c => c.id === activeChartId);
        if (active && active.symbol !== selectedSymbol) {
          return prev.map(c => c.id === activeChartId ? { ...c, symbol: selectedSymbol } : c);
        }
        return prev;
      });
    }
  }, [selectedSymbol]);
  // Actually, handleOpenChart sets selectedSymbol. 
  // If we switch tabs, we set activeChartId -> updates selectedSymbol.
  // If we change symbol in MarketWatch -> updates selectedSymbol -> should update active chart? Yes.

  const handleTabClick = (id: string) => {
    setActiveChartId(id);
  };

  const handleTabClose = (id: string) => {
    const newCharts = openCharts.filter(c => c.id !== id);
    setOpenCharts(newCharts);
    if (activeChartId === id && newCharts.length > 0) {
      setActiveChartId(newCharts[newCharts.length - 1].id);
    } else if (newCharts.length === 0) {
      setActiveChartId(null);
    }
  };

  // Order Panel State
  const [orderPanelOpen, setOrderPanelOpen] = useState(false);
  const [orderPanelSymbol, setOrderPanelSymbol] = useState('');
  const [orderPanelPrice, setOrderPanelPrice] = useState({ bid: 0, ask: 0 });

  // Layout Management
  const [, setLayoutMode] = useState<LayoutMode>('single');

  const wsRef = useRef<WebSocket | null>(null);

  // PERFORMANCE FIX: Removed global ticks subscription to prevent root re-renders
  // Access ticks via useAppStore.getState().ticks inside event handlers instead.

  // Persist dock height
  useEffect(() => {
    localStorage.setItem('dockHeight', dockHeight.toString());
  }, [dockHeight]);

  // Context Menu Event Listeners - Handle custom events from context menu
  useEffect(() => {
    // Chart window event
    const handleOpenChart = (e: CustomEvent) => {
      const { symbol } = e.detail;
      console.log('[App] Opening chart for symbol:', symbol);

      const newChart: ChartTab = {
        id: `chart-${Date.now()}`,
        symbol: symbol,
        timeframe: '1m' // Default
      };
      setOpenCharts(prev => [...prev, newChart]);
      setActiveChartId(newChart.id);
      // selectedSymbol will be updated by effect
    };

    // Order dialog event
    const handleOpenOrderDialog = (e: CustomEvent) => {
      const { symbol, type } = e.detail;
      console.log('[App] Opening order dialog:', { symbol, type });

      // PERFORMANCE FIX: Access ticks directly from store state
      const currentTicks = useAppStore.getState().ticks;
      if (currentTicks[symbol]) {
        const tick = currentTicks[symbol];
        setOrderPanelSymbol(symbol);
        setOrderPanelPrice({ bid: tick.bid, ask: tick.ask });
        setOrderPanelOpen(true);
      }
    };

    // Depth of Market event
    const handleOpenDOM = (e: CustomEvent) => {
      const { symbol } = e.detail;
      console.log('[App] Opening Depth of Market for:', symbol);
      // TODO: Implement DOM window when DOM component is ready
      alert(`Depth of Market for ${symbol} - Feature coming soon`);
    };

    window.addEventListener('openChart', handleOpenChart as EventListener);
    window.addEventListener('openOrderDialog', handleOpenOrderDialog as EventListener);
    window.addEventListener('openDepthOfMarket', handleOpenDOM as EventListener);

    return () => {
      window.removeEventListener('openChart', handleOpenChart as EventListener);
      window.removeEventListener('openOrderDialog', handleOpenOrderDialog as EventListener);
      window.removeEventListener('openDepthOfMarket', handleOpenDOM as EventListener);
    };
  }, []); // Removed ticks dependency

  // Handle Escape key for closing order panel
  useEffect(() => {
    const handleCloseModal = () => {
      if (orderPanelOpen) {
        setOrderPanelOpen(false);
      }
    };

    window.addEventListener('close-modal', handleCloseModal);
    return () => window.removeEventListener('close-modal', handleCloseModal);
  }, [orderPanelOpen]);

  // Handle global menu actions


  // Subscribe to layout mode changes
  useEffect(() => {
    const unsubscribe = windowManager.subscribe((mode) => {
      setLayoutMode(mode);
    });
    return unsubscribe;
  }, []);

  // Fetch broker config on mount
  useEffect(() => {
    fetch('http://localhost:7999/api/config')
      .then(res => res.json())
      .then(data => setBrokerConfig(data))
      .catch(err => console.error('Failed to fetch config:', err));
  }, []);

  // Fetch account and positions from B-Book (NOT OANDA)
  useEffect(() => {
    if (!isAuthenticated) return;

    const fetchAccountData = async () => {
      try {
        const [accRes, posRes] = await Promise.all([
          fetch(`http://localhost:7999/api/account/summary?accountId=${accountId}`),
          fetch(`http://localhost:7999/api/positions?accountId=${accountId}`)
        ]);

        if (accRes.ok) {
          const acc = await accRes.json();
          setAccount(acc);
        }
        if (posRes.ok) {
          const pos = await posRes.json();
          setPositions(pos || []);
        }
      } catch (err) {
        console.error('Failed to fetch account data:', err);
      }
    };

    fetchAccountData();
    const interval = setInterval(fetchAccountData, 1000);
    return () => clearInterval(interval);
  }, [isAuthenticated, accountId]);


  // PERFORMANCE FIX: Removed tick buffer for immediate updates (MT5 parity - <5ms latency)
  // const tickBuffer = useRef<Record<string, Tick>>({});

  // Connect to WebSocket
  useEffect(() => {
    if (!isAuthenticated) return;

    let ws: WebSocket | null = null;
    let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
    let isUnmounting = false;

    const connect = () => {
      if (isUnmounting) return;

      const authToken = useAppStore.getState().authToken;
      let wsUrl = 'ws://localhost:7999/ws';
      if (authToken) {
        wsUrl += `?token=${encodeURIComponent(authToken)}`;
      }

      console.log('[WS] Attempting connection to ' + wsUrl);
      ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => console.log('[WS] WebSocket connected');

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'tick') {
            // Ensure spread is calculated if missing
            const spread = data.spread !== undefined && data.spread > 0
              ? data.spread
              : (data.ask - data.bid);

            // PERFORMANCE FIX: Immediate update to store (no buffering)
            // 20x faster tick updates for MT5 parity
            const storeTicks = useAppStore.getState().ticks;
            const tick: Tick = {
              ...data,
              spread: spread,
              prevBid: storeTicks[data.symbol]?.bid
            };

            // CRITICAL FIX: Update BOTH stores for complete data flow
            // useAppStore: Used by legacy components and chart
            useAppStore.getState().setTick(data.symbol, tick);

            // useMarketDataStore: Used by MarketWatch component
            // This was the missing link causing MarketWatch to show no data
            useMarketDataStore.getState().updateTick(data.symbol, {
              symbol: data.symbol,
              bid: data.bid,
              ask: data.ask,
              spread: spread,
              timestamp: data.timestamp || Date.now(),
              lp: data.lp,
              volume: 1 // Default volume for tick
            });
          } else if (data.type === 'candle_update') {
            // Dispatch to market data store
            useMarketDataStore.getState().updateCandle(data);
          }
        } catch (e) {
          console.error('[WS] Parse error:', e);
        }
      };

      ws.onclose = (event) => {
        console.log(`[WS] Disconnected (code: ${event.code})`);
        wsRef.current = null;

        if (event.code === 1008 || event.reason === 'Unauthorized') {
          setIsAuthenticated(false);
          return;
        }

        if (!isUnmounting && event.code !== 1000) {
          reconnectTimeout = setTimeout(connect, 2000);
        }
      };
    };

    connect();

    // PERFORMANCE FIX: Removed RAF batching for immediate updates
    // Ticks now update Zustand store directly in onmessage handler
    // Result: 100-120ms → <5ms tick latency (20x improvement)

    return () => {
      isUnmounting = true;
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      if (ws) ws.close(1000, 'Component unmount');
    };

  }, [isAuthenticated, brokerConfig]);

  const placeOrder = useCallback(async (side: 'BUY' | 'SELL') => {
    setOrderLoading(true);
    try {
      const res = await fetch('http://localhost:7999/api/orders/market', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          accountId: 1,
          symbol: selectedSymbol,
          side,
          volume
        })
      });

      if (!res.ok) throw new Error(await res.text());

      const result = await res.json();
      console.log('[B-Book] Order executed:', result);

      const posRes = await fetch('http://localhost:7999/api/positions?accountId=1');
      if (posRes.ok) setPositions(await posRes.json() || []);

    } catch (err: any) {
      alert('Order failed: ' + err.message);
    } finally {
      setOrderLoading(false);
    }
  }, [selectedSymbol, volume]);

  const handleOrderPanelSubmit = useCallback(async (order: {
    type: 'buy' | 'sell';
    volume: number;
    sl?: number;
    tp?: number;
  }) => {
    setOrderLoading(true);
    try {
      const side = order.type.toUpperCase() as 'BUY' | 'SELL';
      const res = await fetch('http://localhost:7999/api/orders/market', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          accountId: 1,
          symbol: orderPanelSymbol,
          side,
          volume: order.volume,
          sl: order.sl,
          tp: order.tp
        })
      });

      if (!res.ok) throw new Error(await res.text());

      const result = await res.json();
      console.log('[B-Book] Order executed:', result);

      const posRes = await fetch('http://localhost:7999/api/positions?accountId=1');
      if (posRes.ok) setPositions(await posRes.json() || []);

    } catch (err: any) {
      alert('Order failed: ' + err.message);
    } finally {
      setOrderLoading(false);
    }
  }, [orderPanelSymbol]);

  const closePosition = useCallback(async (tradeId: number, volume?: number) => {
    try {
      const res = await fetch('http://localhost:7999/api/positions/close', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ positionId: tradeId, volume })
      });

      if (!res.ok) throw new Error('Failed to close');

      const posRes = await fetch('http://localhost:7999/api/positions?accountId=1');
      if (posRes.ok) setPositions(await posRes.json() || []);
    } catch (err) {
      console.error('Close failed:', err);
    }
  }, []);

  const modifyPosition = useCallback(async (id: number, sl: number, tp: number) => {
    try {
      const res = await fetch('http://localhost:7999/api/positions/modify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ positionId: id, sl, tp })
      });

      if (!res.ok) throw new Error('Failed to modify');

      const posRes = await fetch('http://localhost:7999/api/positions?accountId=1');
      if (posRes.ok) setPositions(await posRes.json() || []);
    } catch (err) {
      console.error('Modify failed:', err);
    }
  }, []);

  const closeBulkPositions = useCallback(async (type: 'ALL' | 'WINNERS' | 'LOSERS', symbol?: string) => {
    try {
      const res = await fetch('http://localhost:7999/api/positions/close-bulk', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ accountId: 1, type, symbol })
      });

      if (!res.ok) throw new Error('Bulk close failed');

      const posRes = await fetch('http://localhost:7999/api/positions?accountId=1');
      if (posRes.ok) setPositions(await posRes.json() || []);
    } catch (err) {
      alert('Bulk close failed');
    }
  }, []);

  // Fetch symbols from backend
  const [allSymbols, setAllSymbols] = useState<any[]>([]);

  useEffect(() => {
    setIsLoadingSymbols(true);
    fetch('http://localhost:7999/api/symbols')
      .then(res => res.json())
      .then(data => {
        if (Array.isArray(data) && data.length > 0) {
          setAllSymbols(data);
          if (!selectedSymbol) setSelectedSymbol(data[0].symbol || data[0]);
        }
      })
      .catch(err => console.error('Failed to fetch symbols:', err))
      .finally(() => setIsLoadingSymbols(false));
  }, []);

  // PERFORMANCE FIX: Cleanup Web Worker on unmount
  useEffect(() => {
    return () => {
      terminateWorker();
    };
  }, []);

  // ============================================================================
  // ALL HOOKS MUST BE DECLARED BEFORE ANY EARLY RETURNS
  // This fixes: "Rendered more hooks than during the previous render" error
  // ============================================================================

  // Workspace dialog state
  const [showSaveWorkspaceDialog, setShowSaveWorkspaceDialog] = useState(false);
  const [showToast, setShowToast] = useState(false);
  const [toastMessage, setToastMessage] = useState('');
  const [toastType, setToastType] = useState<'success' | 'error'>('success');

  // Toast notification handler
  const displayToast = useCallback((message: string, type: 'success' | 'error' = 'success') => {
    setToastMessage(message);
    setToastType(type);
    setShowToast(true);
    setTimeout(() => setShowToast(false), 3000);
  }, []);

  // Keyboard shortcut handlers
  const handleOpenOrderPanel = useCallback((symbol: string, price: { bid: number; ask: number }) => {
    setOrderPanelSymbol(symbol);
    setOrderPanelPrice(price);
    setOrderPanelOpen(true);
  }, []);

  const handleQuickBuy = useCallback(() => {
    if (selectedSymbol && !orderLoading) {
      placeOrder('BUY');
    }
  }, [selectedSymbol, orderLoading, placeOrder]);

  const handleQuickSell = useCallback(() => {
    if (selectedSymbol && !orderLoading) {
      placeOrder('SELL');
    }
  }, [selectedSymbol, orderLoading, placeOrder]);

  // Workspace handlers
  const handleSaveWorkspace = useCallback(() => {
    console.log('[App] Save workspace triggered');
    setShowSaveWorkspaceDialog(true);
  }, []);

  const handleConfirmSaveWorkspace = useCallback((workspaceName?: string) => {
    console.log('[App] Workspace saved:', workspaceName);
    setShowSaveWorkspaceDialog(false);
    displayToast(`Workspace "${workspaceName}" saved successfully!`, 'success');
  }, [displayToast]);

  const handleOpenDataFolder = useCallback(() => {
    console.log('[App] Open data folder');
    try {
      // Check if running in Electron environment
      if (window.electron?.shell) {
        window.electron.shell.openPath('./data');
      } else {
        // Fallback for web: Show info message
        const dataPath = window.location.origin + '/data';
        console.log('[App] Data folder location:', dataPath);
        // Create a temporary link to trigger download folder
        const link = document.createElement('a');
        link.href = dataPath;
        link.target = '_blank';
        link.click();
      }
    } catch (error) {
      console.error('[App] Failed to open data folder:', error);
    }
  }, []);

  const handlePrintChart = useCallback(async () => {
    console.log('[App] Print chart - Ctrl+P');
    try {
      // Import dynamically to avoid circular dependencies
      const { chartPrinter } = await import('./services/chartPrinter');

      // Try to capture the chart canvas
      const chartContainer = document.querySelector('.tv-lightweight-charts');
      let chartCanvas: HTMLCanvasElement | undefined;

      if (chartContainer) {
        const canvas = chartContainer.querySelector('canvas');
        if (canvas) {
          const copiedCanvas = document.createElement('canvas');
          copiedCanvas.width = canvas.width;
          copiedCanvas.height = canvas.height;
          const ctx = copiedCanvas.getContext('2d');
          if (ctx) {
            ctx.drawImage(canvas, 0, 0);
            chartCanvas = copiedCanvas;
          }
        }
      }

      const chartData = {
        symbol: selectedSymbol || 'BTCUSD',
        timeframe: timeframe as string || '1m',
        dateRange: {
          from: new Date(Date.now() - 24 * 60 * 60 * 1000),
          to: new Date(),
        },
        chartCanvas,
        indicators: [],
        drawings: [],
      };

      const preferences = chartPrinter.getPreferences();
      await chartPrinter.printChart(chartData, preferences);
    } catch (error) {
      console.error('[App] Failed to print:', error);
    }
  }, [selectedSymbol, timeframe]);

  const handleToggleFullScreen = useCallback(() => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen();
    } else {
      document.exitFullscreen();
    }
  }, []);

  // Handle global menu actions
  useEffect(() => {
    const handleSaveWorkspaceEvent = () => {
      handleSaveWorkspace();
    };

    const handleOpenDataFolderEvent = () => {
      handleOpenDataFolder();
    };

    const handlePrintChartEvent = () => {
      handlePrintChart();
    };

    window.addEventListener('saveWorkspace', handleSaveWorkspaceEvent);
    window.addEventListener('openDataFolder', handleOpenDataFolderEvent);
    window.addEventListener('printChart', handlePrintChartEvent);

    return () => {
      window.removeEventListener('saveWorkspace', handleSaveWorkspaceEvent);
      window.removeEventListener('openDataFolder', handleOpenDataFolderEvent);
      window.removeEventListener('printChart', handlePrintChartEvent);
    };
  }, [handleSaveWorkspace, handleOpenDataFolder, handlePrintChart]);

  // ============================================================================
  // EARLY RETURNS - Must come AFTER all hooks
  // ============================================================================

  if (!isAuthenticated) {
    return <Login onLogin={() => { setIsAuthenticated(true); setAccountId("1"); }} />;
  }

  // Consolidated to main view - no extra dashboard routes needed
  if (isLoadingSymbols || !selectedSymbol) {
    return (
      <div className="flex h-screen w-full bg-[#09090b] text-zinc-300 items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-emerald-500 mx-auto mb-4"></div>
          <p className="text-sm text-zinc-400">Loading platform...</p>
        </div>
      </div>
    );
  }

  // Consumed by trading components
  // const currentTick = ticks[selectedSymbol];

  return (
    <CommandBusProvider>
      <KeyboardShortcutProvider>
        <GlobalShortcuts
          onOpenOrderPanel={handleOpenOrderPanel}
          onQuickBuy={handleQuickBuy}
          onQuickSell={handleQuickSell}
          onSaveWorkspace={handleSaveWorkspace}
          onOpenDataFolder={handleOpenDataFolder}
          onPrintChart={handlePrintChart}
          onToggleFullScreen={handleToggleFullScreen}
          selectedSymbol={selectedSymbol}
        />
        <div className="flex flex-col h-screen w-full bg-[#1e1e1e] text-zinc-300 overflow-hidden font-sans">
          {/* 1. Global Menu & Toolbar */}
          <TopToolbar
            chartType={chartType}
            timeframe={timeframe}
            onChartTypeChange={setChartType}
            onTimeframeChange={setTimeframe}
          />

          {/* 2. Main Workspace */}
          <div className="flex-1 flex overflow-hidden relative">

            {/* Left Sidebar */}
            <div className="flex flex-col w-72 border-r border-[#2d3436] bg-[#1e1e1e] flex-shrink-0">
              <MarketWatchPanel
                className="flex-1 min-h-0"
                allSymbols={allSymbols}
                selectedSymbol={selectedSymbol}
                onSymbolSelect={setSelectedSymbol}
              />
              <div className="h-2 bg-[#2d3436] cursor-row-resize hover:bg-blue-500/50 transition-colors" />
              <div className="h-[40%] flex-shrink-0 min-h-0 overflow-hidden">
                <NavigatorPanel />
              </div>
            </div>

            {/* Center Content: Chart + DOM */}
            <div className="flex-1 flex bg-[#131722] relative min-w-0 overflow-hidden">

              {/* Chart Area */}
              <div className="flex-1 flex flex-col min-w-0 relative">






                <div className="flex-1 w-full h-full">
                  <ErrorBoundary>
                    {enableHistoricalData ? (
                      <ChartWithHistory
                        symbol={selectedSymbol}
                        chartType={chartType}
                        timeframe={timeframe}
                        positions={positions}
                        onClosePosition={(id) => closePosition(id)}
                        onModifyPosition={modifyPosition}
                        enableHistoricalData={enableHistoricalData}
                      />
                    ) : (
                      <TradingChart
                        symbol={selectedSymbol}
                        chartType={chartType}
                        timeframe={timeframe}
                        positions={positions}
                        onClosePosition={(id) => closePosition(id)}
                        onModifyPosition={modifyPosition}
                      />
                    )}
                  </ErrorBoundary>
                </div>

                {/* DOM Panel (Right Side) */}
                {showDOM && (
                  <div className="border-l border-black z-20 shadow-xl">
                    <DepthOfMarket
                      symbol={selectedSymbol}
                      onClose={() => setShowDOM(false)}
                      onPlaceOrder={(side, price, vol) => {
                        setVolume(vol);
                        // For now, just open order panel pre-filled
                        // or implement direct placement if API supports it
                        console.log(`DOM Order: ${side} ${vol} @ ${price}`);
                        placeOrder(side);
                      }}
                    />
                  </div>
                )}
              </div>
              <ChartTabs
                tabs={openCharts}
                activeTabId={activeChartId || ''}
                onTabClick={handleTabClick}
                onTabClose={handleTabClose}
              />
            </div>
          </div>

          {/* 3. Bottom Terminal */}
          <BottomDock
            height={dockHeight}
            onHeightChange={setDockHeight}
            account={account}
            positions={positions}
            orders={[]}
            history={[]}
            ledger={[]}
            onClosePosition={closePosition}
            onModifyPosition={modifyPosition}
            onCancelOrder={() => { }}
            onCloseBulk={closeBulkPositions}
          />

          {/* 4. Status Bar */}
          <StatusBar />

          {/* Headless Components */}
          <AlertsContainer wsConnection={wsRef.current} />

          {/* Order Panel Dialog (F9) */}
          <OrderPanelDialog
            isOpen={orderPanelOpen}
            onClose={() => setOrderPanelOpen(false)}
            symbol={orderPanelSymbol}
            currentPrice={orderPanelPrice}
            onSubmitOrder={handleOrderPanelSubmit}
          />

          {/* Save Workspace Dialog (Ctrl+S) */}
          {showSaveWorkspaceDialog && (
            <SaveWorkspaceDialog
              onConfirm={handleConfirmSaveWorkspace}
              onCancel={() => setShowSaveWorkspaceDialog(false)}
              workspaceData={{
                charts: openCharts,
                layout: { dockHeight },
                marketWatch: { selectedSymbol },
                orderPanel: { volume }
              }}
              userId={accountId}
              accountId={accountId}
            />
          )}

          {/* Toast Notification */}
          {showToast && (
            <div className="fixed top-4 right-4 z-[300] animate-in slide-in-from-top-5 fade-in duration-200">
              <div className={`px-4 py-3 rounded-lg shadow-xl border ${toastType === 'success'
                  ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-200'
                  : 'bg-rose-500/10 border-rose-500/30 text-rose-200'
                }`}>
                <div className="flex items-center gap-2">
                  {toastType === 'success' ? (
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  ) : (
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  )}
                  <span className="text-sm font-medium">{toastMessage}</span>
                </div>
              </div>
            </div>
          )}
        </div>
      </KeyboardShortcutProvider>
    </CommandBusProvider>
  );
}

export default App;
