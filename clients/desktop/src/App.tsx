import { useState, useEffect, useRef, useCallback } from 'react';
import { Login } from './components/Login';
import { TradingChart, ChartControls } from './components/TradingChart';
import type { ChartType, Timeframe } from './components/TradingChart';
import { ErrorBoundary } from './components/ErrorBoundary';
import { BottomDock } from './components/BottomDock';
import { LayoutDashboard, Zap, Settings, BookOpen, ArrowUpDown } from 'lucide-react';
import { MarketWatch } from './components/MarketWatch';
import { TickChart } from './components/TickChart';
import { DepthOfMarket } from './components/DepthOfMarket';
import { DataSyncService } from './services';

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
  const [tickHistory, setTickHistory] = useState<Record<string, number[]>>({}); // For Sparklines
  const [selectedSymbol, setSelectedSymbol] = useState('EURUSD');
  const [positions, setPositions] = useState<Position[]>([]);
  const [account, setAccount] = useState<Account | null>(null);
  const [volume, setVolume] = useState(0.01);
  const [orderLoading, setOrderLoading] = useState(false);
  const [chartType, setChartType] = useState<ChartType>('candlestick');
  const [timeframe, setTimeframe] = useState<Timeframe>('1m');
  const [isChartMaximized, setIsChartMaximized] = useState(false);
  const [brokerConfig, setBrokerConfig] = useState<BrokerConfig | null>(null);
  const [logs, setLogs] = useState<{ time: string, message: string, type: 'info' | 'error' | 'success' }[]>([]);
  const [toasts, setToasts] = useState<{ id: number, message: string, type: 'info' | 'error' | 'success' }[]>([]);

  // Logging helper with Toast
  const addLog = (message: string, type: 'info' | 'error' | 'success' = 'info') => {
    const timestamp = new Date().toLocaleTimeString();
    setLogs(prev => [...prev, { time: timestamp, message, type }].slice(-100));

    // Add Toast
    const id = Date.now() + Math.random();
    setToasts(prev => [...prev, { id, message, type }]);
    setTimeout(() => setToasts(prev => prev.filter(t => t.id !== id)), 3000);
  };

  const [dockHeight, setDockHeight] = useState(() => {
    const saved = localStorage.getItem('dockHeight');
    return saved ? parseInt(saved, 10) : 250;
  });
  const [accountId, setAccountId] = useState<string>("1");
  const [tradeHistory, setTradeHistory] = useState<any[]>([]);
  const [ledger, setLedger] = useState<any[]>([]);

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

    // Start data sync service on mount
    DataSyncService.start().then(() => {
      console.log('[App] DataSyncService started');
    });
  }, []);

  // Fetch account and positions from B-Book (NOT OANDA)
  useEffect(() => {
    if (!isAuthenticated) return;

    const fetchAccountData = async () => {
      try {
        // Use B-Book API for internal balance (NOT OANDA)
        const [accRes, posRes, histRes, ledgerRes] = await Promise.all([
          fetch(`http://localhost:8080/api/account/summary?accountId=${accountId}`),
          fetch(`http://localhost:8080/api/positions?accountId=${accountId}`),
          fetch(`http://localhost:8080/api/trades?accountId=${accountId}`),
          fetch(`http://localhost:8080/api/ledger?accountId=${accountId}`)
        ]);

        if (accRes.ok) {
          const acc = await accRes.json();
          setAccount(acc);
        }
        if (posRes.ok) {
          const pos = await posRes.json();
          setPositions(pos || []);
        }
        if (histRes.ok) {
          const hist = await histRes.json();
          setTradeHistory(hist);
        }
        if (ledgerRes.ok) {
          const led = await ledgerRes.json();
          setLedger(led);
        }
      } catch (err) {
        console.error('Failed to fetch account data:', err);
      }
    };

    fetchAccountData();
    const interval = setInterval(fetchAccountData, 1000);
    return () => clearInterval(interval);
  }, [isAuthenticated, accountId]);

  const handleLogin = (id: string) => {
    setAccountId(id);
    setIsAuthenticated(true);
  };
  // Connect to WebSocket after login
  useEffect(() => {
    if (!isAuthenticated) return;

    const ws = new WebSocket('ws://localhost:8080/ws');
    wsRef.current = ws;

    ws.onopen = () => {
      addLog('Connected to Real-time Feed', 'success');
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'tick') {
        const midPrice = (data.bid + data.ask) / 2;

        setTicks(prev => ({
          ...prev,
          [data.symbol]: {
            ...data,
            prevBid: prev[data.symbol]?.bid
          }
        }));

        setTickHistory(prev => {
          const hist = prev[data.symbol] || [];
          // Keep last 40 ticks for sparkline
          const newHist = [...hist, midPrice].slice(-40);
          return { ...prev, [data.symbol]: newHist };
        });
      }
    };

    ws.onerror = (error) => {
      addLog('Connection Error: ' + JSON.stringify(error), 'error');
    };

    ws.onclose = () => {
      addLog('Disconnected from feed', 'error');
    };

    return () => {
      ws.close();
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
          accountId: parseInt(accountId), // Default B-Book account
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
      addLog(`Order Executed: ${side} ${volume} ${selectedSymbol}`, 'success');

      // Refresh positions from B-Book
      const posRes = await fetch(`http://localhost:8080/api/positions?accountId=${accountId}`);
      if (posRes.ok) {
        const positions = await posRes.json();
        setPositions(positions || []);
      }
    } catch (err: any) {
      addLog(`Order Failed: ${err.message}`, 'error');
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
      const posRes = await fetch(`http://localhost:8080/api/positions?accountId=${accountId}`);
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
      const posRes = await fetch(`http://localhost:8080/api/positions?accountId=${accountId}`);
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
          accountId: parseInt(accountId),
          type,
          symbol,
        })
      });

      if (!res.ok) throw new Error('Bulk close failed');

      const result = await res.json();
      console.log('Bulk close result:', result);

      // Refresh positions
      const posRes = await fetch(`http://localhost:8080/api/positions?accountId=${accountId}`);
      if (posRes.ok) {
        setPositions(await posRes.json() || []);
      }
    } catch (err) {
      console.error('Bulk close failed:', err);
      alert('Bulk close failed');
    }
  }, []);


  const [oneClickEnabled, setOneClickEnabled] = useState(false);
  const [activeMobileTab, setActiveMobileTab] = useState<'CHART' | 'TRADE' | 'ACCOUNT'>('CHART');

  const handlePlacePendingOrder = useCallback(async (price: number, type: 'LIMIT' | 'STOP') => {
    // Determine side based on current price
    const currentPrice = ticks[selectedSymbol]?.bid || 0;
    if (!currentPrice) return;

    const side = price > currentPrice ? 'SELL' : 'BUY'; // Simple logic: click above = SELL LIMIT, below = BUY LIMIT

    // For now, B-Book might only support MARKET. If so, we can't place pending.
    // But we will try to implement or at least log.
    // Assuming we send a Limit order.
    // TODO: Implement Pending Orders in Backend. For now, alert or log.
    addLog(`One-Click: Placing ${side} LIMIT at ${price.toFixed(5)} (Pending Orders not yet fully supported on B-Book)`, 'info');
    // alert(`One-Click: Would place ${side} LIMIT at ${price.toFixed(5)}`);
  }, [selectedSymbol, ticks]);

  const handleDownloadData = useCallback(async () => {
    try {
      const res = await fetch(`http://localhost:8080/ohlc?symbol=${selectedSymbol}&timeframe=${timeframe}&limit=5000`);
      if (!res.ok) {
        addLog('Failed to fetch data for download', 'error');
        return;
      }
      const data = await res.json();
      if (!Array.isArray(data) || data.length === 0) {
        addLog('No data available to download', 'info');
        return;
      }

      // Create CSV content
      const csvHeader = 'Time,Open,High,Low,Close,Volume\n';
      const csvRows = data.map((c: any) => {
        const date = new Date(c.time * 1000).toISOString();
        return `${date},${c.open},${c.high},${c.low},${c.close},${c.volume || 0}`;
      }).join('\n');
      const csvContent = csvHeader + csvRows;

      // Download file
      const blob = new Blob([csvContent], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${selectedSymbol}_${timeframe}_${Date.now()}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      addLog(`Downloaded ${data.length} candles for ${selectedSymbol}`, 'success');
    } catch (err) {
      console.error('Download failed:', err);
      addLog('Download failed', 'error');
    }
  }, [selectedSymbol, timeframe]);

  if (!isAuthenticated) {
    return <Login onLogin={() => setIsAuthenticated(true)} />;
  }

  const currentTick = ticks[selectedSymbol];
  const sortedSymbols = Object.keys(ticks).sort();

  return (
    <div className="flex h-screen w-full bg-[#09090b] text-zinc-300 overflow-hidden font-sans flex-col">
      {/* Sidebar - Desktop Only */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <div className="hidden md:flex w-14 border-r border-zinc-800 flex-col items-center py-4 gap-4 bg-zinc-900/30">
          <div className="w-8 h-8 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-lg flex items-center justify-center text-black font-bold text-xs shadow-lg shadow-emerald-500/20">
            RTX
          </div>
          <nav className="flex flex-col gap-2 w-full items-center">
            <NavItem icon={<LayoutDashboard size={18} />} active />
            <NavItem icon={<Zap size={18} />} />
            <NavItem icon={<BookOpen size={18} />} />
          </nav>
          <div className="mt-auto">
            <NavItem icon={<Settings size={18} />} />
          </div>
        </div>

        {/* Main Content Area */}
        <div className="flex-1 flex flex-col min-w-0 pb-16 md:pb-0"> {/* Add padding bottom for mobile nav */}
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
          <div className="flex-1 flex overflow-hidden relative">
            {/* Market Watch */}
            <div className={`${!isChartMaximized ? 'flex' : 'hidden'} ${activeMobileTab === 'TRADE' ? 'absolute inset-0 z-20 w-full' : 'hidden md:flex w-64'} flex-col overflow-hidden bg-[#09090b] border-r border-zinc-800`}>
              <MarketWatch
                ticks={ticks}
                tickHistory={tickHistory}
                selectedSymbol={selectedSymbol}
                onSelectSymbol={setSelectedSymbol}
                onNewOrder={(sym) => { setSelectedSymbol(sym); /* Open order modal logic if separate */ }}
              />
            </div>

            {/* Chart Area */}
            <div className={`${activeMobileTab === 'CHART' ? 'flex' : 'hidden md:flex'} flex-1 flex-col bg-[#131722] relative`}>
              {/* Chart Controls */}
              <div className="flex items-center justify-between border-b border-zinc-800 bg-zinc-900/30">
                <ChartControls
                  chartType={chartType}
                  timeframe={timeframe}
                  onChartTypeChange={setChartType}
                  onTimeframeChange={setTimeframe}
                  isMaximized={isChartMaximized}
                  onToggleMaximize={() => setIsChartMaximized(!isChartMaximized)}
                  oneClickEnabled={oneClickEnabled}
                  onToggleOneClick={() => setOneClickEnabled(!oneClickEnabled)}
                  onDownloadData={handleDownloadData}
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
                    oneClickEnabled={oneClickEnabled}
                    onPlacePendingOrder={handlePlacePendingOrder}
                  />
                </ErrorBoundary>
              </div>
            </div>
            {/* NEW: Pro Panel (Tick Chart & DOM) - Only visible when maximized chart is NOT active for now, or toggleable */}
            {!isChartMaximized && (
              <div className="hidden lg:flex w-72 flex-col border-l border-zinc-800 bg-[#131722] overflow-hidden">
                <div className="flex-1 p-2 flex flex-col gap-2 overflow-y-auto">
                  {/* Tick Chart */}
                  <div className="h-48 shrink-0">
                    <TickChart symbol={selectedSymbol} currentTick={currentTick} />
                  </div>
                  {/* DOM */}
                  <div className="flex-1 min-h-[300px]">
                    <DepthOfMarket symbol={selectedSymbol} tick={currentTick} />
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Bottom Dock - Desktop: Always. Mobile: Only if TRADE or ACCOUNT */}
          <div className={`
            ${['TRADE', 'ACCOUNT'].includes(activeMobileTab) ? 'block flex-1' : 'hidden'} 
            md:block md:flex-none
          `}>
            <BottomDock
              height={dockHeight}
              onHeightChange={setDockHeight}
              account={account}
              positions={positions}
              orders={[]}
              // history={tradeHistory}
              // ledger={ledger}
              logs={logs}
              onClosePosition={closePosition}
              onModifyPosition={modifyPosition}
              onCancelOrder={() => { }}
              onCloseBulk={closeBulkPositions}
            />
          </div>
        </div>
      </div>

      {/* Mobile Bottom Navigation */}
      <div className="md:hidden fixed bottom-0 left-0 right-0 h-14 bg-zinc-950 border-t border-zinc-800 flex items-center justify-around z-50">
        <button
          onClick={() => setActiveMobileTab('CHART')}
          className={`flex flex-col items-center gap-1 px-4 py-1.5 rounded transition-colors ${activeMobileTab === 'CHART' ? 'text-emerald-400' : 'text-zinc-500'}`}
        >
          <Zap size={20} />
          <span className="text-[10px] font-medium">Chart</span>
        </button>
        <button
          onClick={() => setActiveMobileTab('TRADE')}
          className={`flex flex-col items-center gap-1 px-4 py-1.5 rounded transition-colors ${activeMobileTab === 'TRADE' ? 'text-emerald-400' : 'text-zinc-500'}`}
        >
          <ArrowUpDown size={20} />
          <span className="text-[10px] font-medium">Trade</span>
        </button>
        <button
          onClick={() => setActiveMobileTab('ACCOUNT')}
          className={`flex flex-col items-center gap-1 px-4 py-1.5 rounded transition-colors ${activeMobileTab === 'ACCOUNT' ? 'text-emerald-400' : 'text-zinc-500'}`}
        >
          <LayoutDashboard size={20} />
          <span className="text-[10px] font-medium">Account</span>
        </button>
      </div>

      {/* Toast Container */}
      <div className="fixed top-16 right-4 z-[60] flex flex-col gap-2 pointer-events-none max-w-sm w-full items-end">
        {toasts.map(toast => (
          <div
            key={toast.id}
            className={`
              pointer-events-auto px-4 py-2.5 rounded-lg shadow-2xl border backdrop-blur-md flex items-center gap-3 min-w-[200px]
              ${toast.type === 'success' ? 'bg-emerald-950/80 border-emerald-500/30 text-emerald-400 shadow-emerald-500/10' : ''}
              ${toast.type === 'error' ? 'bg-red-950/80 border-red-500/30 text-red-500 shadow-red-500/10' : ''}
              ${toast.type === 'info' ? 'bg-blue-950/80 border-blue-500/30 text-blue-400 shadow-blue-500/10' : ''}
            `}
          >
            <div className={`w-2 h-2 rounded-full ${toast.type === 'success' ? 'bg-emerald-500' :
              toast.type === 'error' ? 'bg-red-500' : 'bg-blue-500'
              }`} />
            <span className="text-xs font-medium">{toast.message}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function formatPrice(price: number, symbol: string): string {
  if (symbol.includes('JPY')) {
    return price.toFixed(3);
  }
  return price.toFixed(5);
}



function NavItem({ icon, active }: { icon: React.ReactNode; active?: boolean }) {
  return (
    <button className={`p-2 rounded-lg transition-all ${active ? 'bg-zinc-800 text-emerald-400' : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800/50'
      }`}>
      {icon}
    </button>
  );
}

// Tab component removed as it was unused

export default App;
