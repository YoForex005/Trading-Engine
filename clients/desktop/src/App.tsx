import { useState, useEffect, useRef, useCallback } from 'react';
import { Login } from './components/Login';
import { TradingChart, ChartControls } from './components/TradingChart';
import type { ChartType, Timeframe } from './components/TradingChart';
import { ErrorBoundary } from './components/ErrorBoundary';
import { BottomDock } from './components/BottomDock';
import { AlertsContainer } from './components/AlertsContainer';
import { AnalyticsDashboard } from './components/AnalyticsDashboard';
import { LayoutDashboard, Zap, Settings, BookOpen, ArrowUpDown, BarChart3 } from 'lucide-react';
import { useAppStore } from './store/useAppStore';

type AppView = 'trading' | 'analytics';

interface Tick {
  symbol: string;
  bid: number;
  ask: number;
  spread?: number;
  timestamp: number;
  prevBid?: number;
  lp?: string;
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
  const [ticks, setTicks] = useState<Record<string, Tick>>({});
  const [selectedSymbol, setSelectedSymbol] = useState('');
  const [positions, setPositions] = useState<Position[]>([]);
  const [account, setAccount] = useState<Account | null>(null);
  const [volume, setVolume] = useState(0.01);
  const [orderLoading, setOrderLoading] = useState(false);
  const [chartType, setChartType] = useState<ChartType>('candlestick');
  const [timeframe, setTimeframe] = useState<Timeframe>('1m');
  const [isChartMaximized, setIsChartMaximized] = useState(false);
  const [brokerConfig, setBrokerConfig] = useState<BrokerConfig | null>(null);
  const [dockHeight, setDockHeight] = useState(() => {
    const saved = localStorage.getItem('dockHeight');
    return saved ? parseInt(saved, 10) : 250;
  });
  const [accountId, setAccountId] = useState<string>("1");
  const [isLoadingSymbols, setIsLoadingSymbols] = useState(false);
  const [currentView, setCurrentView] = useState<AppView>('trading');

  const wsRef = useRef<WebSocket | null>(null);

  // Persist dock height
  useEffect(() => {
    localStorage.setItem('dockHeight', dockHeight.toString());
  }, [dockHeight]);

  // Fetch broker config on mount
  useEffect(() => {
    fetch('http://localhost:8080/api/config')
      .then(res => res.json())
      .then(data => setBrokerConfig(data))
      .catch(err => console.error('Failed to fetch config:', err));
  }, []);

  // Fetch account and positions from B-Book (NOT OANDA)
  useEffect(() => {
    if (!isAuthenticated) return;

    const fetchAccountData = async () => {
      try {
        // Use B-Book API for internal balance (NOT OANDA)
        const [accRes, posRes] = await Promise.all([
          fetch(`http://localhost:8080/api/account/summary?accountId=${accountId}`),
          fetch(`http://localhost:8080/api/positions?accountId=${accountId}`)
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


  // Tick buffer for throttled updates (prevents UI lag)
  const tickBuffer = useRef<Record<string, Tick>>({});

  // Connect to WebSocket after login with auto-reconnect and throttled updates
  useEffect(() => {
    if (!isAuthenticated) {
      console.log('[WS] Not authenticated, skipping WebSocket connection');
      return;
    }

    let ws: WebSocket | null = null;
    let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
    let flushInterval: ReturnType<typeof setInterval> | null = null;
    let isUnmounting = false;

    const connect = () => {
      if (isUnmounting) return;

      // Get auth token from store
      const authToken = useAppStore.getState().authToken;

      // Build WebSocket URL with auth token as query parameter
      let wsUrl = 'ws://localhost:8080/ws';
      if (authToken) {
        wsUrl += `?token=${encodeURIComponent(authToken)}`;
      }

      console.log('[WS] Attempting connection to ' + wsUrl);
      ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('[WS] WebSocket connected, waiting for tick data...');
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'tick') {
            // Buffer tick (don't trigger render immediately)
            tickBuffer.current[data.symbol] = {
              ...data,
              prevBid: ticks[data.symbol]?.bid
            };
          } else {
            console.log('[WS] Non-tick message:', data.type || 'unknown');
          }
        } catch (e) {
          console.error('[WS] Parse error:', e);
        }
      };

      ws.onerror = (error) => {
        console.error('[WS] Error:', error);
      };

      ws.onclose = (event) => {
        console.log(`[WS] Disconnected (code: ${event.code}, reason: ${event.reason})`);
        wsRef.current = null;

        // Check if auth failed (code 1008 is policy violation, used for 401)
        if (event.code === 1008 || event.reason === 'Unauthorized') {
          console.error('[WS] Authentication failed - logging out');
          setIsAuthenticated(false);
          return;
        }

        // Auto-reconnect after 2 seconds if not intentionally closed
        if (!isUnmounting && event.code !== 1000) {
          console.log('[WS] Reconnecting in 2 seconds...');
          reconnectTimeout = setTimeout(connect, 2000);
        }
      };
    };

    connect();

    // Flush tick buffer to state at 10 FPS (100ms) for smooth UI updates
    flushInterval = setInterval(() => {
      const buffer = tickBuffer.current;
      if (Object.keys(buffer).length > 0) {
        setTicks(prev => ({ ...prev, ...buffer }));
        tickBuffer.current = {};
      }
    }, 100);

    return () => {
      isUnmounting = true;
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      if (flushInterval) clearInterval(flushInterval);
      if (ws) {
        ws.close(1000, 'Component unmount');
      }
    };

  }, [isAuthenticated, brokerConfig]);

  const placeOrder = useCallback(async (side: 'BUY' | 'SELL') => {
    setOrderLoading(true);
    try {
      // Use B-Book API (RTX internal) instead of OANDA
      const res = await fetch('http://localhost:8080/api/orders/market', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          accountId: 1, // Default B-Book account
          symbol: selectedSymbol,
          side,
          volume
        })
      });

      if (!res.ok) {
        const error = await res.text();
        throw new Error(error);
      }

      const result = await res.json();
      console.log('[B-Book] Order executed:', result);

      // Refresh positions from B-Book
      const posRes = await fetch('http://localhost:8080/api/positions?accountId=1');
      if (posRes.ok) {
        const positions = await posRes.json();
        setPositions(positions || []);
      }
    } catch (err: any) {
      console.error('Order failed:', err);
      alert('Order failed: ' + err.message);
    } finally {
      setOrderLoading(false);
    }
  }, [selectedSymbol, volume]);

  const closePosition = useCallback(async (tradeId: number, volume?: number) => {
    try {
      // Use B-Book API to close position
      const res = await fetch('http://localhost:8080/api/positions/close', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          positionId: tradeId,
          volume: volume // optional partial close
        })
      });

      if (!res.ok) throw new Error('Failed to close');

      // Refresh positions from B-Book
      const posRes = await fetch('http://localhost:8080/api/positions?accountId=1');
      if (posRes.ok) {
        setPositions(await posRes.json() || []);
      }
    } catch (err) {
      console.error('Close failed:', err);
    }
  }, []);

  const modifyPosition = useCallback(async (id: number, sl: number, tp: number) => {
    try {
      const res = await fetch('http://localhost:8080/api/positions/modify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          positionId: id,
          sl,
          tp
        })
      });

      if (!res.ok) throw new Error('Failed to modify');

      // Refresh positions
      const posRes = await fetch('http://localhost:8080/api/positions?accountId=1');
      if (posRes.ok) {
        setPositions(await posRes.json() || []);
      }
    } catch (err) {
      console.error('Modify failed:', err);
    }
  }, []);

  const closeBulkPositions = useCallback(async (type: 'ALL' | 'WINNERS' | 'LOSERS', symbol?: string) => {
    try {
      const res = await fetch('http://localhost:8080/api/positions/close-bulk', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          accountId: 1,
          type,
          symbol
        })
      });

      if (!res.ok) throw new Error('Bulk close failed');

      const result = await res.json();
      console.log('Bulk close result:', result);

      // Refresh positions
      const posRes = await fetch('http://localhost:8080/api/positions?accountId=1');
      if (posRes.ok) {
        setPositions(await posRes.json() || []);
      }
    } catch (err) {
      console.error('Bulk close failed:', err);
      alert('Bulk close failed');
    }
  }, []);

  // Fetch symbols from backend for Search
  const [allSymbols, setAllSymbols] = useState<any[]>([]);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    setIsLoadingSymbols(true);
    fetch('http://localhost:8080/api/symbols')
      .then(res => res.json())
      .then(data => {
        if (Array.isArray(data) && data.length > 0) {
          setAllSymbols(data);
          // Set the first symbol as default if no symbol is currently selected
          if (!selectedSymbol) {
            const firstSymbol = data[0].symbol || data[0];
            setSelectedSymbol(firstSymbol);
          }
        }
      })
      .catch(err => console.error('Failed to fetch symbols:', err))
      .finally(() => setIsLoadingSymbols(false));
  }, []);

  if (!isAuthenticated) {
    return <Login onLogin={() => { setIsAuthenticated(true); setAccountId("1"); }} />;
  }

  // Show loading state while symbols are being fetched
  if (isLoadingSymbols || !selectedSymbol) {
    return (
      <div className="flex h-screen w-full bg-[#09090b] text-zinc-300 items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-emerald-500 mx-auto mb-4"></div>
          <p className="text-sm text-zinc-400">Loading symbols...</p>
        </div>
      </div>
    );
  }

  const currentTick = ticks[selectedSymbol];

  // Filter symbols based on search
  // If search is empty, show ticking symbols (legacy behavior) OR show all valid symbols?
  // Architect said: "Source of truth: GetSymbols". So we should show allSymbols.
  // But wait, if they aren't ticking, we might show "---" for price. That is fine.

  // Combine allSymbols with ticks keys to ensure we catch everything
  const uniqueSymbols = Array.from(new Set([
    ...allSymbols.map(s => s.symbol),
    ...Object.keys(ticks)
  ])).sort();

  const filteredSymbols = uniqueSymbols.filter(s =>
    s.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="flex h-screen w-full bg-[#09090b] text-zinc-300 overflow-hidden font-sans flex-col">
      {/* Alerts Container - Real-time WebSocket alerts */}
      <AlertsContainer wsConnection={wsRef.current} />

      {/* Sidebar - NOW FLEX ROW for layout if we want, but let's keep simple first */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <div className="w-14 border-r border-zinc-800 flex flex-col items-center py-4 gap-4 bg-zinc-900/30">
          <div className="w-8 h-8 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-lg flex items-center justify-center text-black font-bold text-xs shadow-lg shadow-emerald-500/20">
            RTX
          </div>
          <nav className="flex flex-col gap-2 w-full items-center">
            <NavItem
              icon={<LayoutDashboard size={18} />}
              active={currentView === 'trading'}
              onClick={() => setCurrentView('trading')}
              title="Trading"
            />
            <NavItem
              icon={<BarChart3 size={18} />}
              active={currentView === 'analytics'}
              onClick={() => setCurrentView('analytics')}
              title="Analytics"
            />
            <NavItem icon={<Zap size={18} />} title="Automation" />
            <NavItem icon={<BookOpen size={18} />} title="History" />
          </nav>
          <div className="mt-auto">
            <NavItem icon={<Settings size={18} />} title="Settings" />
          </div>
        </div>

        {/* Main Content Area */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Analytics View */}
          {currentView === 'analytics' && (
            <AnalyticsDashboard
              onBack={() => setCurrentView('trading')}
              className="flex-1"
            />
          )}

          {/* Trading View */}
          {currentView === 'trading' && (
            <>
          {/* Header */}
          <header className="h-9 border-b border-zinc-800 flex items-center justify-between px-3 bg-zinc-900/20">
            <div className="flex items-center gap-3">
              <span className="text-sm font-semibold">{selectedSymbol}</span>
              {currentTick && (
                <>
                  <span className="text-xs font-mono text-emerald-400">{formatPrice(currentTick.bid, selectedSymbol)}</span>
                  <span className="text-xs font-mono text-red-400">{formatPrice(currentTick.ask, selectedSymbol)}</span>
                  <span className="text-[10px] font-mono text-zinc-600">
                    {currentTick.spread?.toFixed(1)} pips
                  </span>
                </>
              )}
            </div>
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-1.5 px-2 py-0.5 bg-emerald-500/10 rounded border border-emerald-500/20">
                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
                <span className="text-[10px] text-emerald-400 font-medium">{brokerConfig?.priceFeedLP || 'LP'}</span>
              </div>
            </div>
          </header>

          {/* Middle Section: Market Watch + Chart */}
          <div className="flex-1 flex overflow-hidden">
            {/* Market Watch */}
            {!isChartMaximized && (
              <div className="w-56 border-r border-zinc-800 overflow-y-auto flex flex-col">
                <div className="p-2 border-b border-zinc-800 text-[10px] font-medium text-zinc-500 uppercase flex justify-between items-center sticky top-0 bg-zinc-900/90 backdrop-blur z-10">
                  <span>Market Watch</span>
                  <span className="text-emerald-400 normal-case">{filteredSymbols.length} pairs</span>
                </div>

                {/* Search Input */}
                <div className="p-2 border-b border-zinc-800 sticky top-8 bg-zinc-900 z-10">
                  <input
                    type="text"
                    placeholder="Search symbols..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 focus:outline-none focus:border-emerald-500/50"
                  />
                </div>

                <div className="flex-1 overflow-y-auto">
                  {filteredSymbols.map((symbol) => (
                    <MarketWatchRow
                      key={symbol}
                      symbol={symbol}
                      tick={ticks[symbol]}
                      selected={symbol === selectedSymbol}
                      onClick={() => setSelectedSymbol(symbol)}
                    />
                  ))}
                  {filteredSymbols.length === 0 && (
                    <div className="p-4 text-center text-xs text-zinc-600 italic">
                      No symbols found
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Chart Area */}
            <div className="flex-1 flex flex-col bg-[#131722] relative">
              {/* Chart Controls */}
              <div className="flex items-center justify-between border-b border-zinc-800 bg-zinc-900/30">
                <ChartControls
                  chartType={chartType}
                  timeframe={timeframe}
                  onChartTypeChange={setChartType}
                  onTimeframeChange={setTimeframe}
                  isMaximized={isChartMaximized}
                  onToggleMaximize={() => setIsChartMaximized(!isChartMaximized)}
                />

                {/* Trade Controls */}
                <div className="flex items-center gap-2 px-2 py-1">
                  <div className="flex items-center gap-1 bg-zinc-900 border border-zinc-800 rounded px-2 py-0.5">
                    <span className="text-[10px] font-medium text-zinc-500">VOL</span>
                    <input
                      type="number"
                      step={0.01}
                      min={0.01}
                      value={volume}
                      onChange={(e) => setVolume(parseFloat(e.target.value))}
                      className="w-14 bg-transparent text-sm font-mono text-center focus:outline-none"
                    />
                  </div>
                  <button
                    onClick={() => placeOrder('SELL')}
                    disabled={orderLoading || !currentTick}
                    className="flex flex-col items-center px-4 py-1 bg-red-500/10 hover:bg-red-500/20 text-red-500 rounded border border-red-500/20 transition-colors disabled:opacity-50"
                  >
                    <span className="text-[10px] font-bold">SELL</span>
                    <span className="text-xs font-mono">{currentTick?.bid ? formatPrice(currentTick.bid, selectedSymbol) : '---'}</span>
                  </button>
                  <button
                    onClick={() => placeOrder('BUY')}
                    disabled={orderLoading || !currentTick}
                    className="flex flex-col items-center px-4 py-1 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 rounded border border-emerald-500/20 transition-colors disabled:opacity-50"
                  >
                    <span className="text-[10px] font-bold">BUY</span>
                    <span className="text-xs font-mono">{currentTick?.ask ? formatPrice(currentTick.ask, selectedSymbol) : '---'}</span>
                  </button>
                </div>
              </div>

              {/* Chart */}
              <div className="flex-1 relative">
                <ErrorBoundary>
                  <TradingChart
                    symbol={selectedSymbol}
                    currentPrice={currentTick}
                    chartType={chartType}
                    timeframe={timeframe}
                    positions={positions}
                    onClosePosition={(id) => closePosition(id)}
                    onModifyPosition={modifyPosition}
                  />
                </ErrorBoundary>
              </div>
            </div>
          </div>
            </>
          )}
        </div>
      </div>

      {/* Bottom Dock - Only show in trading view */}
      {currentView === 'trading' && (
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
      )}
    </div>
  );
}

function formatPrice(price: number, symbol: string): string {
  if (symbol.includes('JPY')) {
    return price.toFixed(3);
  }
  return price.toFixed(5);
}

function MarketWatchRow({ symbol, tick, selected, onClick }: {
  symbol: string;
  tick: Tick;
  selected: boolean;
  onClick: () => void;
}) {
  const direction = tick?.prevBid !== undefined
    ? tick.bid > tick.prevBid ? 'up' : tick.bid < tick.prevBid ? 'down' : 'none'
    : 'none';

  return (
    <div
      onClick={onClick}
      className={`flex items-center justify-between px-2 py-1 cursor-pointer transition-all text-xs ${selected ? 'bg-emerald-500/10 border-l-2 border-emerald-500' : 'hover:bg-zinc-800/50 border-l-2 border-transparent'
        }`}
    >
      <div className="flex items-center gap-1.5">
        <ArrowUpDown className={`w-2.5 h-2.5 ${direction === 'up' ? 'text-emerald-400' : direction === 'down' ? 'text-red-400' : 'text-zinc-600'
          }`} />
        <span className="font-medium">{symbol}</span>
      </div>
      <div className="text-right font-mono">
        <div className="text-emerald-400">{tick?.bid ? formatPrice(tick.bid, symbol) : '---'}</div>
        <div className="text-[10px] text-red-400">{tick?.ask ? formatPrice(tick.ask, symbol) : '---'}</div>
      </div>
    </div>
  );
}

function NavItem({ icon, active, onClick, title }: {
  icon: React.ReactNode;
  active?: boolean;
  onClick?: () => void;
  title?: string;
}) {
  return (
    <button
      onClick={onClick}
      title={title}
      className={`p-2 rounded-lg transition-all ${active ? 'bg-zinc-800 text-emerald-400' : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800/50'
      }`}
    >
      {icon}
    </button>
  );
}

// Tab component removed as it was unused

export default App;
