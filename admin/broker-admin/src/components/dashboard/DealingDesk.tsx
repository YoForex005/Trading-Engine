'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { API_CONFIG } from '@/config/api';
import {
  TrendingUp,
  TrendingDown,
  CheckCircle,
  XCircle,
  RefreshCw,
  AlertTriangle,
  Clock,
  Ban,
  X as CloseIcon,
  UserX,
} from 'lucide-react';

// ============================================================================
// TYPES
// ============================================================================

type OrderSide = 'BUY' | 'SELL';
type OrderType = 'Market' | 'Limit' | 'Stop';
type OrderStatus = 'pending' | 'accepted' | 'rejected';

interface LiveOrder {
  id: string;
  timestamp: number;
  clientId: string;
  clientName: string;
  symbol: string;
  side: OrderSide;
  volume: number;
  price: number;
  type: OrderType;
  isNew?: boolean;
}

interface PendingOrder {
  id: string;
  timestamp: number;
  clientId: string;
  clientName: string;
  symbol: string;
  side: OrderSide;
  volume: number;
  requestedPrice: number;
  currentPrice: number;
  pendingDuration: number; // seconds
}

interface OpenPosition {
  ticket: string;
  clientId: string;
  clientName: string;
  symbol: string;
  side: OrderSide;
  volume: number;
  openPrice: number;
  currentPrice: number;
  pnl: number;
  swap: number;
  durationMinutes: number;
}

interface DealingStats {
  ordersProcessed: number;
  avgExecutionTime: number; // ms
  requoteRate: number; // percentage
  slippageDistribution: { label: string; value: number }[];
}

// ============================================================================
// MOCK DATA GENERATORS
// ============================================================================

const SYMBOLS = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'XAUUSD', 'BTCUSD', 'ETHUSD'];
const CLIENTS = [
  { id: '100234', name: 'John Smith' },
  { id: '100567', name: 'Sarah Johnson' },
  { id: '101234', name: 'Michael Brown' },
  { id: '102345', name: 'Emily Davis' },
  { id: '103456', name: 'David Wilson' },
  { id: '104567', name: 'Lisa Anderson' },
  { id: '105678', name: 'Robert Taylor' },
  { id: '106789', name: 'Jennifer Martinez' },
];

const generateLiveOrders = (count: number): LiveOrder[] => {
  const now = Date.now();
  return Array.from({ length: count }, (_, i) => {
    const client = CLIENTS[Math.floor(Math.random() * CLIENTS.length)];
    const symbol = SYMBOLS[Math.floor(Math.random() * SYMBOLS.length)];
    const side: OrderSide = Math.random() > 0.5 ? 'BUY' : 'SELL';
    const type: OrderType = ['Market', 'Limit', 'Stop'][Math.floor(Math.random() * 3)] as OrderType;

    return {
      id: `LO-${now}-${i}`,
      timestamp: now - Math.floor(Math.random() * 60000), // last 60 seconds
      clientId: client.id,
      clientName: client.name,
      symbol,
      side,
      volume: parseFloat((Math.random() * 5 + 0.01).toFixed(2)),
      price: parseFloat((Math.random() * 100 + 1).toFixed(5)),
      type,
      isNew: i < 3, // First 3 are new
    };
  });
};

const generatePendingOrders = (count: number): PendingOrder[] => {
  const now = Date.now();
  return Array.from({ length: count }, (_, i) => {
    const client = CLIENTS[Math.floor(Math.random() * CLIENTS.length)];
    const symbol = SYMBOLS[Math.floor(Math.random() * SYMBOLS.length)];
    const side: OrderSide = Math.random() > 0.5 ? 'BUY' : 'SELL';
    const requestedPrice = parseFloat((Math.random() * 100 + 1).toFixed(5));
    const slippage = (Math.random() - 0.5) * 0.0005; // Random slippage
    const currentPrice = parseFloat((requestedPrice + slippage).toFixed(5));

    return {
      id: `PO-${now}-${i}`,
      timestamp: now - Math.floor(Math.random() * 30000), // last 30 seconds
      clientId: client.id,
      clientName: client.name,
      symbol,
      side,
      volume: parseFloat((Math.random() * 5 + 0.01).toFixed(2)),
      requestedPrice,
      currentPrice,
      pendingDuration: Math.floor(Math.random() * 30),
    };
  });
};

const generateOpenPositions = (count: number): OpenPosition[] => {
  const now = Date.now();
  return Array.from({ length: count }, (_, i) => {
    const client = CLIENTS[Math.floor(Math.random() * CLIENTS.length)];
    const symbol = SYMBOLS[Math.floor(Math.random() * SYMBOLS.length)];
    const side: OrderSide = Math.random() > 0.5 ? 'BUY' : 'SELL';
    const openPrice = parseFloat((Math.random() * 100 + 1).toFixed(5));
    const priceChange = (Math.random() - 0.5) * 2; // Random price movement
    const currentPrice = parseFloat((openPrice + priceChange).toFixed(5));
    const volume = parseFloat((Math.random() * 5 + 0.01).toFixed(2));
    const pnl = parseFloat(((currentPrice - openPrice) * volume * (side === 'BUY' ? 1 : -1) * 100).toFixed(2));

    return {
      ticket: `T${100000 + i}`,
      clientId: client.id,
      clientName: client.name,
      symbol,
      side,
      volume,
      openPrice,
      currentPrice,
      pnl,
      swap: parseFloat((Math.random() * 10 - 5).toFixed(2)),
      durationMinutes: Math.floor(Math.random() * 1440), // up to 24 hours
    };
  });
};

// ============================================================================
// MAIN COMPONENT
// ============================================================================

export default function DealingDesk() {
  const [liveOrders, setLiveOrders] = useState<LiveOrder[]>([]);
  const [pendingOrders, setPendingOrders] = useState<PendingOrder[]>([]);
  const [openPositions, setOpenPositions] = useState<OpenPosition[]>([]);
  const [tradingHalted, setTradingHalted] = useState(false);
  const [disabledSymbol, setDisabledSymbol] = useState<string | null>(null);
  const [selectedClient, setSelectedClient] = useState<string>('');

  // Filters
  const [positionClientFilter, setPositionClientFilter] = useState('');
  const [positionSymbolFilter, setPositionSymbolFilter] = useState('');

  // Stats
  const [stats, setStats] = useState<DealingStats>({
    ordersProcessed: 1247,
    avgExecutionTime: 45, // ms
    requoteRate: 2.3, // %
    slippageDistribution: [
      { label: '-2 to -1 pips', value: 8 },
      { label: '-1 to 0 pips', value: 22 },
      { label: '0 to 1 pips', value: 45 },
      { label: '1 to 2 pips', value: 18 },
      { label: '2+ pips', value: 7 },
    ],
  });

  // Initialize mock data
  useEffect(() => {
    setLiveOrders(generateLiveOrders(20));
    setPendingOrders(generatePendingOrders(5));
    setOpenPositions(generateOpenPositions(15));
  }, []);

  // Auto-refresh live orders every 2 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      setLiveOrders((prev) => {
        const newOrders = generateLiveOrders(3);
        const updated = [...newOrders.map(o => ({ ...o, isNew: true })), ...prev.slice(0, 17)];
        return updated;
      });
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  // Clear "new" flag after 3 seconds
  useEffect(() => {
    const timer = setTimeout(() => {
      setLiveOrders((prev) => prev.map(o => ({ ...o, isNew: false })));
    }, 3000);

    return () => clearTimeout(timer);
  }, [liveOrders]);

  // Update pending durations
  useEffect(() => {
    const interval = setInterval(() => {
      setPendingOrders((prev) =>
        prev.map((o) => ({ ...o, pendingDuration: o.pendingDuration + 1 }))
      );
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  // Handle pending order actions
  const handleAcceptOrder = (orderId: string) => {
    setPendingOrders((prev) => prev.filter((o) => o.id !== orderId));
    setStats((prev) => ({ ...prev, ordersProcessed: prev.ordersProcessed + 1 }));
  };

  const handleRejectOrder = (orderId: string) => {
    setPendingOrders((prev) => prev.filter((o) => o.id !== orderId));
    setStats((prev) => ({
      ...prev,
      ordersProcessed: prev.ordersProcessed + 1,
      requoteRate: prev.requoteRate + 0.1,
    }));
  };

  const handleRequoteOrder = (orderId: string) => {
    setPendingOrders((prev) => prev.filter((o) => o.id !== orderId));
    setStats((prev) => ({
      ...prev,
      ordersProcessed: prev.ordersProcessed + 1,
      requoteRate: prev.requoteRate + 0.5,
    }));
  };

  const handleClosePosition = (ticket: string) => {
    if (confirm(`Close position ${ticket}?`)) {
      setOpenPositions((prev) => prev.filter((p) => p.ticket !== ticket));
    }
  };

  const handleCloseAllForClient = () => {
    if (!selectedClient) {
      alert('Please select a client');
      return;
    }
    if (confirm(`Close all positions for client ${selectedClient}?`)) {
      setOpenPositions((prev) => prev.filter((p) => p.clientId !== selectedClient));
      setSelectedClient('');
    }
  };

  const handleDisableSymbol = () => {
    if (!disabledSymbol) {
      alert('Please select a symbol');
      return;
    }
    if (confirm(`Temporarily disable trading on ${disabledSymbol}?`)) {
      alert(`Trading disabled for ${disabledSymbol}`);
      // In real implementation, would call API
    }
  };

  // Filtered positions
  const filteredPositions = useMemo(() => {
    return openPositions.filter((p) => {
      const matchesClient = !positionClientFilter || p.clientId.includes(positionClientFilter) || p.clientName.toLowerCase().includes(positionClientFilter.toLowerCase());
      const matchesSymbol = !positionSymbolFilter || p.symbol.toLowerCase().includes(positionSymbolFilter.toLowerCase());
      return matchesClient && matchesSymbol;
    });
  }, [openPositions, positionClientFilter, positionSymbolFilter]);

  // Exposure summary
  const exposureSummary = useMemo(() => {
    const totalPnL = filteredPositions.reduce((sum, p) => sum + p.pnl, 0);
    const totalSwap = filteredPositions.reduce((sum, p) => sum + p.swap, 0);
    return { totalPnL, totalSwap, count: filteredPositions.length };
  }, [filteredPositions]);

  const formatDuration = (minutes: number) => {
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return hours > 0 ? `${hours}h ${mins}m` : `${mins}m`;
  };

  const getSlippage = (requested: number, current: number) => {
    const diff = current - requested;
    return diff;
  };

  const getSlippageColor = (requested: number, current: number, side: OrderSide) => {
    const diff = current - requested;
    // For BUY: lower is better, for SELL: higher is better
    if (side === 'BUY') {
      return diff < 0 ? 'text-emerald-400' : 'text-rose-400';
    } else {
      return diff > 0 ? 'text-emerald-400' : 'text-rose-400';
    }
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-zinc-300 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-[#1E2026] border-b border-zinc-700">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-bold text-zinc-100">Dealing Desk</h1>
          {tradingHalted && (
            <span className="px-2 py-1 text-xs font-bold bg-rose-500/20 text-rose-400 border border-rose-500/50 rounded">
              TRADING HALTED
            </span>
          )}
        </div>
        <div className="text-xs text-zinc-500">Real-time Order Management</div>
      </div>

      {/* Main Content - Scrollable */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Stats Cards Row */}
        <div className="grid grid-cols-4 gap-3">
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="text-xs text-zinc-500 mb-1">Orders Processed</div>
            <div className="text-2xl font-bold text-zinc-100">{stats.ordersProcessed}</div>
            <div className="text-xs text-emerald-400 mt-1">Today</div>
          </div>
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="text-xs text-zinc-500 mb-1">Avg Execution Time</div>
            <div className="text-2xl font-bold text-zinc-100">{stats.avgExecutionTime}ms</div>
            <div className="text-xs text-emerald-400 mt-1">-5ms from yesterday</div>
          </div>
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="text-xs text-zinc-500 mb-1">Requote Rate</div>
            <div className="text-2xl font-bold text-zinc-100">{stats.requoteRate.toFixed(1)}%</div>
            <div className="text-xs text-yellow-400 mt-1">+0.2% from yesterday</div>
          </div>
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="text-xs text-zinc-500 mb-1">Pending Orders</div>
            <div className="text-2xl font-bold text-yellow-400">{pendingOrders.length}</div>
            <div className="text-xs text-zinc-500 mt-1">Requires action</div>
          </div>
        </div>

        {/* Quick Actions Panel */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-4">
          <h2 className="text-sm font-bold text-zinc-100 mb-3 flex items-center gap-2">
            <AlertTriangle size={14} className="text-rose-400" />
            Quick Actions
          </h2>
          <div className="grid grid-cols-3 gap-3">
            {/* Halt Trading */}
            <div className="flex items-center gap-2">
              <button
                onClick={() => setTradingHalted(!tradingHalted)}
                className={`px-3 py-2 text-xs font-bold rounded transition-colors ${
                  tradingHalted
                    ? 'bg-rose-500/20 text-rose-400 border border-rose-500/50 hover:bg-rose-500/30'
                    : 'bg-zinc-700/50 text-zinc-300 border border-zinc-600 hover:bg-zinc-600/50'
                }`}
              >
                <Ban size={12} className="inline mr-1" />
                {tradingHalted ? 'Resume Trading' : 'Halt Trading'}
              </button>
            </div>

            {/* Disable Symbol */}
            <div className="flex items-center gap-2">
              <select
                value={disabledSymbol || ''}
                onChange={(e) => setDisabledSymbol(e.target.value)}
                className="flex-1 px-2 py-1.5 text-xs bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
              >
                <option value="">Select Symbol</option>
                {SYMBOLS.map((s) => (
                  <option key={s} value={s}>{s}</option>
                ))}
              </select>
              <button
                onClick={handleDisableSymbol}
                className="px-3 py-1.5 text-xs font-bold bg-yellow-500/20 text-yellow-400 border border-yellow-500/50 rounded hover:bg-yellow-500/30"
              >
                Disable
              </button>
            </div>

            {/* Close All for Client */}
            <div className="flex items-center gap-2">
              <select
                value={selectedClient}
                onChange={(e) => setSelectedClient(e.target.value)}
                className="flex-1 px-2 py-1.5 text-xs bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
              >
                <option value="">Select Client</option>
                {CLIENTS.map((c) => (
                  <option key={c.id} value={c.id}>{c.name} ({c.id})</option>
                ))}
              </select>
              <button
                onClick={handleCloseAllForClient}
                className="px-3 py-1.5 text-xs font-bold bg-rose-500/20 text-rose-400 border border-rose-500/50 rounded hover:bg-rose-500/30"
              >
                <UserX size={12} className="inline mr-1" />
                Close All
              </button>
            </div>
          </div>
        </div>

        {/* Two Column Layout: Live Orders + Pending Orders */}
        <div className="grid grid-cols-2 gap-4">
          {/* Live Order Flow */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden flex flex-col" style={{ height: '300px' }}>
            <div className="flex items-center justify-between px-3 py-2 bg-zinc-800/50 border-b border-zinc-700">
              <h2 className="text-xs font-bold text-zinc-100 flex items-center gap-2">
                <RefreshCw size={12} className="text-blue-400" />
                Live Order Flow
              </h2>
              <span className="text-xs text-zinc-500">Auto-refresh: 2s</span>
            </div>
            <div className="flex-1 overflow-y-auto">
              <div className="divide-y divide-zinc-700/50">
                {liveOrders.map((order) => (
                  <div
                    key={order.id}
                    className={`px-3 py-2 text-xs transition-colors ${
                      order.isNew ? 'bg-blue-500/10 animate-pulse' : 'hover:bg-zinc-800/30'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="text-zinc-500">
                          {new Date(order.timestamp).toLocaleTimeString()}
                        </span>
                        <span className="text-zinc-400">{order.clientName}</span>
                        <span className="font-mono text-zinc-300">{order.symbol}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span
                          className={`px-1.5 py-0.5 font-bold rounded ${
                            order.side === 'BUY'
                              ? 'bg-emerald-500/20 text-emerald-400'
                              : 'bg-rose-500/20 text-rose-400'
                          }`}
                        >
                          {order.side}
                        </span>
                        <span className="text-zinc-400">{order.volume}</span>
                        <span className="text-zinc-500">@{order.price.toFixed(5)}</span>
                        <span className="text-xs text-zinc-600">{order.type}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Pending Orders Queue */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden flex flex-col" style={{ height: '300px' }}>
            <div className="flex items-center justify-between px-3 py-2 bg-zinc-800/50 border-b border-zinc-700">
              <h2 className="text-xs font-bold text-zinc-100 flex items-center gap-2">
                <Clock size={12} className="text-yellow-400" />
                Pending Orders ({pendingOrders.length})
              </h2>
              <span className="text-xs text-zinc-500">Instant Execution Mode</span>
            </div>
            <div className="flex-1 overflow-y-auto">
              {pendingOrders.length === 0 ? (
                <div className="flex items-center justify-center h-full text-xs text-zinc-500">
                  No pending orders
                </div>
              ) : (
                <div className="divide-y divide-zinc-700/50">
                  {pendingOrders.map((order) => {
                    const slippage = getSlippage(order.requestedPrice, order.currentPrice);
                    const slippageColor = getSlippageColor(order.requestedPrice, order.currentPrice, order.side);

                    return (
                      <div key={order.id} className="px-3 py-2 text-xs hover:bg-zinc-800/30">
                        <div className="flex items-center justify-between mb-1">
                          <div className="flex items-center gap-2">
                            <span className="text-zinc-500 flex items-center gap-1">
                              <Clock size={10} />
                              {order.pendingDuration}s
                            </span>
                            <span className="text-zinc-400">{order.clientName}</span>
                            <span className="font-mono text-zinc-300">{order.symbol}</span>
                            <span
                              className={`px-1.5 py-0.5 font-bold rounded ${
                                order.side === 'BUY'
                                  ? 'bg-emerald-500/20 text-emerald-400'
                                  : 'bg-rose-500/20 text-rose-400'
                              }`}
                            >
                              {order.side}
                            </span>
                            <span className="text-zinc-400">{order.volume}</span>
                          </div>
                        </div>
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3 text-xs">
                            <span className="text-zinc-500">
                              Req: {order.requestedPrice.toFixed(5)}
                            </span>
                            <span className="text-zinc-400">
                              Curr: {order.currentPrice.toFixed(5)}
                            </span>
                            <span className={slippageColor}>
                              Slip: {slippage > 0 ? '+' : ''}{slippage.toFixed(5)}
                            </span>
                          </div>
                          <div className="flex items-center gap-1">
                            <button
                              onClick={() => handleAcceptOrder(order.id)}
                              className="px-2 py-1 bg-emerald-500/20 text-emerald-400 border border-emerald-500/50 rounded hover:bg-emerald-500/30 font-bold"
                            >
                              <CheckCircle size={12} className="inline mr-1" />
                              Accept
                            </button>
                            <button
                              onClick={() => handleRequoteOrder(order.id)}
                              className="px-2 py-1 bg-yellow-500/20 text-yellow-400 border border-yellow-500/50 rounded hover:bg-yellow-500/30 font-bold"
                            >
                              <RefreshCw size={12} className="inline mr-1" />
                              Requote
                            </button>
                            <button
                              onClick={() => handleRejectOrder(order.id)}
                              className="px-2 py-1 bg-rose-500/20 text-rose-400 border border-rose-500/50 rounded hover:bg-rose-500/30 font-bold"
                            >
                              <XCircle size={12} className="inline mr-1" />
                              Reject
                            </button>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Open Positions Monitor */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 bg-zinc-800/50 border-b border-zinc-700">
            <h2 className="text-xs font-bold text-zinc-100">Open Positions Monitor ({filteredPositions.length})</h2>
            <div className="flex items-center gap-2">
              <input
                type="text"
                placeholder="Filter by client..."
                value={positionClientFilter}
                onChange={(e) => setPositionClientFilter(e.target.value)}
                className="px-2 py-1 text-xs bg-zinc-800 border border-zinc-600 rounded text-zinc-300 placeholder-zinc-600 focus:outline-none focus:border-blue-500"
              />
              <input
                type="text"
                placeholder="Filter by symbol..."
                value={positionSymbolFilter}
                onChange={(e) => setPositionSymbolFilter(e.target.value)}
                className="px-2 py-1 text-xs bg-zinc-800 border border-zinc-600 rounded text-zinc-300 placeholder-zinc-600 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          {/* Table */}
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-zinc-800/30 border-b border-zinc-700">
                <tr>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Ticket</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Client</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Symbol</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Side</th>
                  <th className="px-3 py-2 text-right font-bold text-zinc-400">Volume</th>
                  <th className="px-3 py-2 text-right font-bold text-zinc-400">Open Price</th>
                  <th className="px-3 py-2 text-right font-bold text-zinc-400">Current Price</th>
                  <th className="px-3 py-2 text-right font-bold text-zinc-400">P&L</th>
                  <th className="px-3 py-2 text-right font-bold text-zinc-400">Swap</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Duration</th>
                  <th className="px-3 py-2 text-center font-bold text-zinc-400">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-700/50">
                {filteredPositions.map((position) => (
                  <tr key={position.ticket} className="hover:bg-zinc-800/30">
                    <td className="px-3 py-2 font-mono text-zinc-300">{position.ticket}</td>
                    <td className="px-3 py-2 text-zinc-400">
                      {position.clientName}
                      <div className="text-xs text-zinc-600">{position.clientId}</div>
                    </td>
                    <td className="px-3 py-2 font-mono text-zinc-300">{position.symbol}</td>
                    <td className="px-3 py-2">
                      <span
                        className={`px-1.5 py-0.5 font-bold rounded text-xs ${
                          position.side === 'BUY'
                            ? 'bg-emerald-500/20 text-emerald-400'
                            : 'bg-rose-500/20 text-rose-400'
                        }`}
                      >
                        {position.side}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-right text-zinc-400">{position.volume.toFixed(2)}</td>
                    <td className="px-3 py-2 text-right font-mono text-zinc-400">{position.openPrice.toFixed(5)}</td>
                    <td className="px-3 py-2 text-right font-mono text-zinc-300">{position.currentPrice.toFixed(5)}</td>
                    <td
                      className={`px-3 py-2 text-right font-bold ${
                        position.pnl >= 0 ? 'text-emerald-400' : 'text-rose-400'
                      }`}
                    >
                      {position.pnl >= 0 ? '+' : ''}{position.pnl.toFixed(2)}
                    </td>
                    <td
                      className={`px-3 py-2 text-right ${
                        position.swap >= 0 ? 'text-emerald-400' : 'text-rose-400'
                      }`}
                    >
                      {position.swap >= 0 ? '+' : ''}{position.swap.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 text-zinc-500">{formatDuration(position.durationMinutes)}</td>
                    <td className="px-3 py-2 text-center">
                      <button
                        onClick={() => handleClosePosition(position.ticket)}
                        className="px-2 py-1 text-xs font-bold bg-rose-500/20 text-rose-400 border border-rose-500/50 rounded hover:bg-rose-500/30"
                      >
                        <CloseIcon size={10} className="inline mr-1" />
                        Close
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Exposure Summary Footer */}
          <div className="px-3 py-2 bg-zinc-800/30 border-t border-zinc-700 flex items-center justify-between text-xs">
            <div className="font-bold text-zinc-300">Total Exposure ({exposureSummary.count} positions)</div>
            <div className="flex items-center gap-4">
              <div>
                <span className="text-zinc-500">Total P&L: </span>
                <span
                  className={`font-bold ${
                    exposureSummary.totalPnL >= 0 ? 'text-emerald-400' : 'text-rose-400'
                  }`}
                >
                  {exposureSummary.totalPnL >= 0 ? '+' : ''}{exposureSummary.totalPnL.toFixed(2)}
                </span>
              </div>
              <div>
                <span className="text-zinc-500">Total Swap: </span>
                <span
                  className={`font-bold ${
                    exposureSummary.totalSwap >= 0 ? 'text-emerald-400' : 'text-rose-400'
                  }`}
                >
                  {exposureSummary.totalSwap >= 0 ? '+' : ''}{exposureSummary.totalSwap.toFixed(2)}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Slippage Distribution Chart */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-4">
          <h2 className="text-sm font-bold text-zinc-100 mb-3">Slippage Distribution</h2>
          <div className="space-y-2">
            {stats.slippageDistribution.map((item, index) => (
              <div key={index} className="flex items-center gap-3">
                <div className="w-32 text-xs text-zinc-400">{item.label}</div>
                <div className="flex-1 h-6 bg-zinc-800 rounded overflow-hidden">
                  <div
                    className="h-full bg-blue-500/50 flex items-center justify-end px-2"
                    style={{ width: `${item.value}%` }}
                  >
                    <span className="text-xs font-bold text-zinc-100">{item.value}%</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
