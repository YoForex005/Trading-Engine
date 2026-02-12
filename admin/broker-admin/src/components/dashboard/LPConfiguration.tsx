'use client';

import React, { useState, useMemo } from 'react';
import { API_CONFIG } from '@/config/api';
import {
  Wifi,
  WifiOff,
  AlertCircle,
  Settings,
  TrendingUp,
  Activity,
  CheckCircle,
  Eye,
  EyeOff,
  X as CloseIcon,
  RefreshCw,
  Play,
  ChevronUp,
  ChevronDown,
} from 'lucide-react';

// ============================================================================
// TYPES
// ============================================================================

type LPProtocol = 'FIX' | 'REST' | 'WebSocket';
type LPStatus = 'Connected' | 'Disconnected' | 'Error';
type RoutingMethod = 'default' | 'round-robin' | 'weighted' | 'latency-based';

interface LiquidityProvider {
  id: string;
  name: string;
  protocol: LPProtocol;
  status: LPStatus;
  latency: number; // ms
  uptime: number; // percentage
  symbolsProvided: number;
  dailyVolume: number; // USD
  host: string;
  port: number;
  username: string;
  credentials: string;
  heartbeatInterval: number; // seconds
  maxRetries: number;
  backoffMs: number;
}

interface SymbolMapping {
  lpSymbol: string;
  internalSymbol: string;
}

interface RoutingConfig {
  symbol: string;
  defaultLP: string;
  failoverPriority: string[];
  method: RoutingMethod;
  weight?: Record<string, number>; // LP ID -> weight
}

interface SpreadComparison {
  symbol: string;
  spreads: Record<string, number>; // LP ID -> spread
}

// ============================================================================
// MOCK DATA
// ============================================================================

const MOCK_LPS: LiquidityProvider[] = [
  {
    id: 'lp-yofx',
    name: 'YOFX',
    protocol: 'FIX',
    status: 'Connected',
    latency: 45,
    uptime: 99.8,
    symbolsProvided: 85,
    dailyVolume: 2500000,
    host: 'fix.yofx.com',
    port: 4001,
    username: 'RTX5_DEMO',
    credentials: '••••••••',
    heartbeatInterval: 30,
    maxRetries: 5,
    backoffMs: 5000,
  },
  {
    id: 'lp-binance',
    name: 'Binance',
    protocol: 'WebSocket',
    status: 'Connected',
    latency: 38,
    uptime: 99.9,
    symbolsProvided: 120,
    dailyVolume: 3200000,
    host: 'stream.binance.com',
    port: 9443,
    username: 'api_user',
    credentials: '••••••••',
    heartbeatInterval: 60,
    maxRetries: 3,
    backoffMs: 3000,
  },
  {
    id: 'lp-oanda',
    name: 'OANDA',
    protocol: 'REST',
    status: 'Connected',
    latency: 52,
    uptime: 99.5,
    symbolsProvided: 68,
    dailyVolume: 1800000,
    host: 'api.oanda.com',
    port: 443,
    username: 'oanda_api',
    credentials: '••••••••',
    heartbeatInterval: 0,
    maxRetries: 3,
    backoffMs: 2000,
  },
  {
    id: 'lp-lmax',
    name: 'LMAX',
    protocol: 'FIX',
    status: 'Disconnected',
    latency: 0,
    uptime: 98.2,
    symbolsProvided: 45,
    dailyVolume: 0,
    host: 'fix.lmax.com',
    port: 5001,
    username: 'lmax_user',
    credentials: '••••••••',
    heartbeatInterval: 30,
    maxRetries: 5,
    backoffMs: 10000,
  },
  {
    id: 'lp-currenex',
    name: 'Currenex',
    protocol: 'FIX',
    status: 'Error',
    latency: 0,
    uptime: 95.0,
    symbolsProvided: 52,
    dailyVolume: 0,
    host: 'fix.currenex.com',
    port: 6001,
    username: 'currenex_demo',
    credentials: '••••••••',
    heartbeatInterval: 30,
    maxRetries: 5,
    backoffMs: 5000,
  },
];

const SYMBOL_MAPPINGS: Record<string, SymbolMapping[]> = {
  'lp-yofx': [
    { lpSymbol: 'EUR/USD', internalSymbol: 'EURUSD' },
    { lpSymbol: 'GBP/USD', internalSymbol: 'GBPUSD' },
    { lpSymbol: 'USD/JPY', internalSymbol: 'USDJPY' },
  ],
  'lp-binance': [
    { lpSymbol: 'BTCUSDT', internalSymbol: 'BTCUSD' },
    { lpSymbol: 'ETHUSDT', internalSymbol: 'ETHUSD' },
  ],
  'lp-oanda': [
    { lpSymbol: 'EUR_USD', internalSymbol: 'EURUSD' },
    { lpSymbol: 'GBP_USD', internalSymbol: 'GBPUSD' },
  ],
};

const ROUTING_CONFIGS: RoutingConfig[] = [
  {
    symbol: 'EURUSD',
    defaultLP: 'lp-yofx',
    failoverPriority: ['lp-oanda', 'lp-lmax'],
    method: 'default',
  },
  {
    symbol: 'BTCUSD',
    defaultLP: 'lp-binance',
    failoverPriority: [],
    method: 'default',
  },
  {
    symbol: 'GBPUSD',
    defaultLP: 'lp-yofx',
    failoverPriority: ['lp-oanda'],
    method: 'weighted',
    weight: { 'lp-yofx': 70, 'lp-oanda': 30 },
  },
];

const SPREAD_COMPARISON: SpreadComparison[] = [
  {
    symbol: 'EURUSD',
    spreads: { 'lp-yofx': 0.8, 'lp-oanda': 1.2, 'lp-lmax': 0.6 },
  },
  {
    symbol: 'GBPUSD',
    spreads: { 'lp-yofx': 1.5, 'lp-oanda': 1.8, 'lp-lmax': 1.3 },
  },
  {
    symbol: 'USDJPY',
    spreads: { 'lp-yofx': 0.9, 'lp-oanda': 1.1 },
  },
  {
    symbol: 'BTCUSD',
    spreads: { 'lp-binance': 5.0 },
  },
];

// ============================================================================
// MAIN COMPONENT
// ============================================================================

export default function LPConfiguration() {
  const [lps, setLps] = useState<LiquidityProvider[]>(MOCK_LPS);
  const [selectedLP, setSelectedLP] = useState<LiquidityProvider | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [showCredentials, setShowCredentials] = useState(false);
  const [testResults, setTestResults] = useState<Record<string, { success: boolean; message: string }>>({});

  // Summary stats
  const stats = useMemo(() => {
    const connectedLPs = lps.filter((lp) => lp.status === 'Connected').length;
    const totalSymbols = new Set(SPREAD_COMPARISON.map((s) => s.symbol)).size;
    const avgLatency = lps.filter((lp) => lp.status === 'Connected').reduce((sum, lp) => sum + lp.latency, 0) / Math.max(connectedLPs, 1);
    const dailyVolume = lps.reduce((sum, lp) => sum + lp.dailyVolume, 0);

    return {
      connectedLPs,
      totalSymbols,
      avgLatency: Math.round(avgLatency),
      dailyVolume,
    };
  }, [lps]);

  const handleEditLP = (lp: LiquidityProvider) => {
    setSelectedLP(lp);
    setShowCredentials(false);
    setShowModal(true);
  };

  const handleTestConnection = (lpId: string) => {
    setTestResults((prev) => ({
      ...prev,
      [lpId]: { success: false, message: 'Testing connection...' },
    }));

    setTimeout(() => {
      const success = Math.random() > 0.3;
      setTestResults((prev) => ({
        ...prev,
        [lpId]: {
          success,
          message: success
            ? 'Connection successful. Latency: 42ms'
            : 'Connection failed. Check host/port settings.',
        },
      }));
    }, 1500);
  };

  const getStatusColor = (status: LPStatus) => {
    switch (status) {
      case 'Connected':
        return 'text-emerald-400';
      case 'Disconnected':
        return 'text-zinc-500';
      case 'Error':
        return 'text-rose-400';
    }
  };

  const getStatusIcon = (status: LPStatus) => {
    switch (status) {
      case 'Connected':
        return <Wifi size={14} className="text-emerald-400" />;
      case 'Disconnected':
        return <WifiOff size={14} className="text-zinc-500" />;
      case 'Error':
        return <AlertCircle size={14} className="text-rose-400" />;
    }
  };

  const getBestSpread = (symbol: string) => {
    const comparison = SPREAD_COMPARISON.find((s) => s.symbol === symbol);
    if (!comparison) return null;

    const spreads = Object.entries(comparison.spreads);
    if (spreads.length === 0) return null;

    return spreads.reduce((min, [lpId, spread]) =>
      spread < min.spread ? { lpId, spread } : min
    , { lpId: spreads[0][0], spread: spreads[0][1] });
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-zinc-300 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-[#1E2026] border-b border-zinc-700">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-bold text-zinc-100">Liquidity Provider Configuration</h1>
        </div>
        <div className="text-xs text-zinc-500">Manage LP connections and routing</div>
      </div>

      {/* Main Content - Scrollable */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Summary Cards */}
        <div className="grid grid-cols-4 gap-3">
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <CheckCircle size={14} className="text-emerald-400" />
              <div className="text-xs text-zinc-500">Connected LPs</div>
            </div>
            <div className="text-2xl font-bold text-emerald-400">{stats.connectedLPs}</div>
            <div className="text-xs text-zinc-500 mt-1">of {lps.length} total</div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp size={14} className="text-blue-400" />
              <div className="text-xs text-zinc-500">Total Symbols</div>
            </div>
            <div className="text-2xl font-bold text-blue-400">{stats.totalSymbols}</div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <Activity size={14} className="text-yellow-400" />
              <div className="text-xs text-zinc-500">Avg Latency</div>
            </div>
            <div className="text-2xl font-bold text-yellow-400">{stats.avgLatency}ms</div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp size={14} className="text-purple-400" />
              <div className="text-xs text-zinc-500">Daily LP Volume</div>
            </div>
            <div className="text-2xl font-bold text-purple-400">
              ${(stats.dailyVolume / 1000000).toFixed(1)}M
            </div>
          </div>
        </div>

        {/* LP Cards Grid */}
        <div className="grid grid-cols-2 gap-3">
          {lps.map((lp) => (
            <div
              key={lp.id}
              className="bg-[#1E2026] border border-zinc-700 rounded p-4"
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  {getStatusIcon(lp.status)}
                  <div>
                    <div className="text-sm font-bold text-zinc-100">{lp.name}</div>
                    <div className="text-xs text-zinc-500">{lp.protocol}</div>
                  </div>
                </div>
                <div className={`text-xs font-bold ${getStatusColor(lp.status)}`}>
                  {lp.status}
                </div>
              </div>

              <div className="grid grid-cols-3 gap-2 text-xs mb-3">
                <div>
                  <div className="text-zinc-500">Latency</div>
                  <div className="font-bold text-zinc-300">
                    {lp.status === 'Connected' ? `${lp.latency}ms` : 'N/A'}
                  </div>
                </div>
                <div>
                  <div className="text-zinc-500">Uptime</div>
                  <div className="font-bold text-zinc-300">{lp.uptime.toFixed(1)}%</div>
                </div>
                <div>
                  <div className="text-zinc-500">Symbols</div>
                  <div className="font-bold text-zinc-300">{lp.symbolsProvided}</div>
                </div>
              </div>

              <div className="mb-3">
                <div className="text-xs text-zinc-500">Daily Volume</div>
                <div className="text-sm font-bold text-zinc-300">
                  ${(lp.dailyVolume / 1000000).toFixed(2)}M
                </div>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleEditLP(lp)}
                  className="flex-1 px-2 py-1.5 text-xs font-bold bg-blue-500/20 text-blue-400 border border-blue-500/50 rounded hover:bg-blue-500/30 flex items-center justify-center gap-1"
                >
                  <Settings size={12} />
                  Configure
                </button>
                <button
                  onClick={() => handleTestConnection(lp.id)}
                  className="flex-1 px-2 py-1.5 text-xs font-bold bg-emerald-500/20 text-emerald-400 border border-emerald-500/50 rounded hover:bg-emerald-500/30 flex items-center justify-center gap-1"
                >
                  <Play size={12} />
                  Test
                </button>
              </div>

              {testResults[lp.id] && (
                <div
                  className={`mt-2 px-2 py-1.5 text-xs rounded ${
                    testResults[lp.id].success
                      ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/50'
                      : 'bg-rose-500/20 text-rose-400 border border-rose-500/50'
                  }`}
                >
                  {testResults[lp.id].message}
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Routing Configuration */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="px-4 py-3 bg-zinc-800/50 border-b border-zinc-700">
            <h2 className="text-sm font-bold text-zinc-100">Routing Configuration</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-zinc-800/30 border-b border-zinc-700">
                <tr>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Symbol</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Default LP</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Failover Priority</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Method</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Weight Distribution</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-700/50">
                {ROUTING_CONFIGS.map((config) => (
                  <tr key={config.symbol} className="hover:bg-zinc-800/30">
                    <td className="px-3 py-2 font-mono text-zinc-300 font-bold">{config.symbol}</td>
                    <td className="px-3 py-2 text-zinc-400">
                      {lps.find((lp) => lp.id === config.defaultLP)?.name || 'N/A'}
                    </td>
                    <td className="px-3 py-2 text-zinc-400">
                      {config.failoverPriority.length > 0
                        ? config.failoverPriority.map((lpId) => lps.find((lp) => lp.id === lpId)?.name).join(' → ')
                        : 'None'}
                    </td>
                    <td className="px-3 py-2">
                      <span className="px-2 py-1 text-xs font-bold bg-blue-500/20 text-blue-400 border border-blue-500/50 rounded">
                        {config.method}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-zinc-400">
                      {config.weight
                        ? Object.entries(config.weight)
                            .map(([lpId, w]) => `${lps.find((lp) => lp.id === lpId)?.name}: ${w}%`)
                            .join(', ')
                        : 'N/A'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Spread Comparison Table */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="px-4 py-3 bg-zinc-800/50 border-b border-zinc-700">
            <h2 className="text-sm font-bold text-zinc-100">LP Spread Comparison</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-zinc-800/30 border-b border-zinc-700">
                <tr>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Symbol</th>
                  {lps.map((lp) => (
                    <th key={lp.id} className="px-3 py-2 text-center font-bold text-zinc-400">
                      {lp.name}
                    </th>
                  ))}
                  <th className="px-3 py-2 text-center font-bold text-zinc-400">Best</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-700/50">
                {SPREAD_COMPARISON.map((comparison) => {
                  const best = getBestSpread(comparison.symbol);

                  return (
                    <tr key={comparison.symbol} className="hover:bg-zinc-800/30">
                      <td className="px-3 py-2 font-mono text-zinc-300 font-bold">
                        {comparison.symbol}
                      </td>
                      {lps.map((lp) => {
                        const spread = comparison.spreads[lp.id];
                        const isBest = best && best.lpId === lp.id;

                        return (
                          <td
                            key={lp.id}
                            className={`px-3 py-2 text-center font-mono ${
                              spread
                                ? isBest
                                  ? 'text-emerald-400 font-bold'
                                  : 'text-zinc-400'
                                : 'text-zinc-600'
                            }`}
                          >
                            {spread ? `${spread.toFixed(1)} pips` : '-'}
                          </td>
                        );
                      })}
                      <td className="px-3 py-2 text-center font-mono text-emerald-400 font-bold">
                        {best ? `${best.spread.toFixed(1)} pips` : '-'}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* LP Detail/Edit Modal */}
      {showModal && selectedLP && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-zinc-700 rounded w-full max-w-2xl max-h-[90vh] overflow-y-auto p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-zinc-100">Configure {selectedLP.name}</h3>
              <button
                onClick={() => setShowModal(false)}
                className="text-zinc-500 hover:text-zinc-300"
              >
                <CloseIcon size={20} />
              </button>
            </div>

            <div className="space-y-4">
              {/* Connection Settings */}
              <div>
                <h4 className="text-sm font-bold text-zinc-300 mb-2">Connection Settings</h4>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Host</label>
                    <input
                      type="text"
                      value={selectedLP.host}
                      className="w-full px-3 py-2 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                      readOnly
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Port</label>
                    <input
                      type="number"
                      value={selectedLP.port}
                      className="w-full px-3 py-2 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                      readOnly
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Username</label>
                    <input
                      type="text"
                      value={selectedLP.username}
                      className="w-full px-3 py-2 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                      readOnly
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Credentials</label>
                    <div className="flex items-center gap-2">
                      <input
                        type={showCredentials ? 'text' : 'password'}
                        value={showCredentials ? 'demo_password_123' : selectedLP.credentials}
                        className="flex-1 px-3 py-2 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                        readOnly
                      />
                      <button
                        onClick={() => setShowCredentials(!showCredentials)}
                        className="p-2 text-zinc-500 hover:text-zinc-300"
                      >
                        {showCredentials ? <EyeOff size={16} /> : <Eye size={16} />}
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              {/* Protocol Config */}
              <div>
                <h4 className="text-sm font-bold text-zinc-300 mb-2">Protocol Configuration</h4>
                <div className="grid grid-cols-3 gap-3">
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Protocol</label>
                    <input
                      type="text"
                      value={selectedLP.protocol}
                      className="w-full px-3 py-2 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                      readOnly
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Heartbeat Interval (s)</label>
                    <input
                      type="number"
                      value={selectedLP.heartbeatInterval}
                      className="w-full px-3 py-2 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                      readOnly
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Max Retries</label>
                    <input
                      type="number"
                      value={selectedLP.maxRetries}
                      className="w-full px-3 py-2 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                      readOnly
                    />
                  </div>
                </div>
              </div>

              {/* Reconnect Policy */}
              <div>
                <h4 className="text-sm font-bold text-zinc-300 mb-2">Reconnect Policy</h4>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Backoff (ms)</label>
                    <input
                      type="number"
                      value={selectedLP.backoffMs}
                      className="w-full px-3 py-2 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                      readOnly
                    />
                  </div>
                </div>
              </div>

              {/* Symbol Mapping Table */}
              <div>
                <h4 className="text-sm font-bold text-zinc-300 mb-2">Symbol Mapping</h4>
                <div className="bg-zinc-800/50 border border-zinc-700 rounded overflow-hidden">
                  <table className="w-full text-xs">
                    <thead className="bg-zinc-800/30 border-b border-zinc-700">
                      <tr>
                        <th className="px-3 py-2 text-left font-bold text-zinc-400">LP Symbol</th>
                        <th className="px-3 py-2 text-center font-bold text-zinc-400">→</th>
                        <th className="px-3 py-2 text-left font-bold text-zinc-400">Internal Symbol</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-700/50">
                      {(SYMBOL_MAPPINGS[selectedLP.id] || []).map((mapping, index) => (
                        <tr key={index} className="hover:bg-zinc-800/30">
                          <td className="px-3 py-2 text-zinc-400">{mapping.lpSymbol}</td>
                          <td className="px-3 py-2 text-center text-zinc-600">→</td>
                          <td className="px-3 py-2 font-mono text-zinc-300 font-bold">
                            {mapping.internalSymbol}
                          </td>
                        </tr>
                      ))}
                      {(!SYMBOL_MAPPINGS[selectedLP.id] || SYMBOL_MAPPINGS[selectedLP.id].length === 0) && (
                        <tr>
                          <td colSpan={3} className="px-3 py-4 text-center text-zinc-500">
                            No symbol mappings configured
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center justify-end gap-2 mt-4">
              <button
                onClick={() => setShowModal(false)}
                className="px-3 py-1.5 text-xs font-bold bg-zinc-700 text-zinc-300 rounded hover:bg-zinc-600"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
