'use client';

import { useState, useMemo, useEffect } from 'react';
import { AlertTriangle, CheckCircle, XCircle, Clock, Download, RefreshCw, Settings, TrendingUp, Activity } from 'lucide-react';

type ReconciliationStatus = 'matched' | 'price_mismatch' | 'volume_mismatch' | 'missing_at_lp' | 'missing_internal';

interface TradeDiscrepancy {
  tradeId: string;
  symbol: string;
  internalPrice: number;
  lpPrice: number | null;
  priceDiff: number;
  volume: number;
  internalTime: string;
  lpTime: string | null;
  timeDiff: number; // milliseconds
  status: ReconciliationStatus;
  lpName: string;
  side: 'buy' | 'sell';
}

interface LPStats {
  lpName: string;
  totalTrades: number;
  matched: number;
  discrepancies: number;
  reconciliationRate: number;
}

const STATUS_CONFIG: Record<ReconciliationStatus, { label: string; color: string; bgColor: string; icon: React.ReactNode }> = {
  matched: { label: 'Matched', color: '#22C55E', bgColor: '#22C55E20', icon: <CheckCircle size={14} /> },
  price_mismatch: { label: 'Price Mismatch', color: '#F59E0B', bgColor: '#F59E0B20', icon: <AlertTriangle size={14} /> },
  volume_mismatch: { label: 'Volume Mismatch', color: '#F59E0B', bgColor: '#F59E0B20', icon: <AlertTriangle size={14} /> },
  missing_at_lp: { label: 'Missing at LP', color: '#EF4444', bgColor: '#EF444420', icon: <XCircle size={14} /> },
  missing_internal: { label: 'Missing Internal', color: '#6B7280', bgColor: '#6B728020', icon: <Clock size={14} /> }
};

const SYMBOLS = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'BTCUSD', 'ETHUSD', 'XAUUSD'];
const LPS = ['YOFX', 'ICMarkets', 'LMax', 'Integral'];

// Mock data generator
function generateMockDiscrepancies(): TradeDiscrepancy[] {
  const now = Date.now();
  const discrepancies: TradeDiscrepancy[] = [];

  for (let i = 0; i < 200; i++) {
    const symbol = SYMBOLS[Math.floor(Math.random() * SYMBOLS.length)];
    const lpName = LPS[Math.floor(Math.random() * LPS.length)];
    const basePrice = symbol.includes('JPY') ? 150 : symbol.includes('XAU') ? 2050 : symbol.includes('BTC') ? 95000 : 1.1;
    const internalPrice = basePrice + (Math.random() - 0.5) * basePrice * 0.001;
    const side = Math.random() > 0.5 ? 'buy' : 'sell';

    // 85% matched, 10% discrepancies, 5% missing
    const rand = Math.random();
    let status: ReconciliationStatus;
    let lpPrice: number | null;
    let priceDiff: number;
    let lpTime: string | null;
    let timeDiff: number;

    if (rand < 0.85) {
      // Matched
      status = 'matched';
      lpPrice = internalPrice + (Math.random() - 0.5) * 0.0001; // Minor variation
      priceDiff = Math.abs(internalPrice - lpPrice);
      lpTime = new Date(now - Math.random() * 24 * 60 * 60 * 1000).toISOString();
      timeDiff = Math.floor(Math.random() * 500); // 0-500ms
    } else if (rand < 0.92) {
      // Price mismatch
      status = Math.random() > 0.5 ? 'price_mismatch' : 'volume_mismatch';
      lpPrice = internalPrice + (Math.random() - 0.5) * basePrice * 0.01; // Larger variation
      priceDiff = Math.abs(internalPrice - lpPrice);
      lpTime = new Date(now - Math.random() * 24 * 60 * 60 * 1000).toISOString();
      timeDiff = Math.floor(Math.random() * 2000);
    } else {
      // Missing
      status = Math.random() > 0.5 ? 'missing_at_lp' : 'missing_internal';
      lpPrice = status === 'missing_at_lp' ? null : internalPrice;
      priceDiff = 0;
      lpTime = status === 'missing_at_lp' ? null : new Date(now - Math.random() * 24 * 60 * 60 * 1000).toISOString();
      timeDiff = 0;
    }

    discrepancies.push({
      tradeId: `TRD${String(100000 + i).padStart(6, '0')}`,
      symbol,
      internalPrice,
      lpPrice,
      priceDiff,
      volume: Math.random() * 10,
      internalTime: new Date(now - Math.random() * 24 * 60 * 60 * 1000).toISOString(),
      lpTime,
      timeDiff,
      status,
      lpName,
      side
    });
  }

  return discrepancies.sort((a, b) => new Date(b.internalTime).getTime() - new Date(a.internalTime).getTime());
}

export default function TradeReconciliation() {
  const [discrepancies, setDiscrepancies] = useState<TradeDiscrepancy[]>([]);
  const [selectedSymbol, setSelectedSymbol] = useState<string>('all');
  const [selectedLP, setSelectedLP] = useState<string>('all');
  const [selectedStatus, setSelectedStatus] = useState<string>('all');
  const [dateRange, setDateRange] = useState({ start: '', end: '' });
  const [showSettings, setShowSettings] = useState(false);
  const [priceThreshold, setPriceThreshold] = useState(0.001);
  const [volumeThreshold, setVolumeThreshold] = useState(0.01);
  const [isReconciling, setIsReconciling] = useState(false);

  useEffect(() => {
    setDiscrepancies(generateMockDiscrepancies());
  }, []);

  // Calculate stats
  const stats = useMemo(() => {
    const total = discrepancies.length;
    const matched = discrepancies.filter(d => d.status === 'matched').length;
    const unmatched = total - matched;
    const discrepancyCount = discrepancies.filter(d =>
      d.status === 'price_mismatch' || d.status === 'volume_mismatch'
    ).length;
    const reconciliationRate = total > 0 ? (matched / total) * 100 : 0;

    return { total, matched, unmatched, discrepancyCount, reconciliationRate };
  }, [discrepancies]);

  // LP comparison stats
  const lpStats = useMemo(() => {
    const stats: Record<string, LPStats> = {};

    discrepancies.forEach(d => {
      if (!stats[d.lpName]) {
        stats[d.lpName] = {
          lpName: d.lpName,
          totalTrades: 0,
          matched: 0,
          discrepancies: 0,
          reconciliationRate: 0
        };
      }

      stats[d.lpName].totalTrades++;
      if (d.status === 'matched') {
        stats[d.lpName].matched++;
      } else if (d.status === 'price_mismatch' || d.status === 'volume_mismatch') {
        stats[d.lpName].discrepancies++;
      }
    });

    // Calculate reconciliation rates
    Object.values(stats).forEach(stat => {
      stat.reconciliationRate = stat.totalTrades > 0 ? (stat.matched / stat.totalTrades) * 100 : 0;
    });

    return Object.values(stats).sort((a, b) => b.reconciliationRate - a.reconciliationRate);
  }, [discrepancies]);

  // Filtered discrepancies
  const filtered = useMemo(() => {
    return discrepancies.filter(d => {
      if (selectedSymbol !== 'all' && d.symbol !== selectedSymbol) return false;
      if (selectedLP !== 'all' && d.lpName !== selectedLP) return false;
      if (selectedStatus !== 'all' && d.status !== selectedStatus) return false;
      if (dateRange.start && new Date(d.internalTime) < new Date(dateRange.start)) return false;
      if (dateRange.end && new Date(d.internalTime) > new Date(dateRange.end)) return false;
      return true;
    });
  }, [discrepancies, selectedSymbol, selectedLP, selectedStatus, dateRange]);

  const handleAutoReconcile = () => {
    setIsReconciling(true);
    setTimeout(() => {
      // Simulate auto-reconciliation
      const updated = discrepancies.map(d => {
        if (d.status === 'price_mismatch' && d.priceDiff <= priceThreshold) {
          return { ...d, status: 'matched' as ReconciliationStatus };
        }
        return d;
      });
      setDiscrepancies(updated);
      setIsReconciling(false);
    }, 2000);
  };

  const handleExportCSV = () => {
    const headers = ['Trade ID', 'Symbol', 'Internal Price', 'LP Price', 'Price Diff', 'Volume', 'Internal Time', 'LP Time', 'Time Diff (ms)', 'Status', 'LP Name'];
    const rows = filtered.map(d => [
      d.tradeId,
      d.symbol,
      d.internalPrice.toFixed(5),
      d.lpPrice?.toFixed(5) || 'N/A',
      d.priceDiff.toFixed(5),
      d.volume.toFixed(2),
      new Date(d.internalTime).toLocaleString(),
      d.lpTime ? new Date(d.lpTime).toLocaleString() : 'N/A',
      d.timeDiff.toString(),
      STATUS_CONFIG[d.status].label,
      d.lpName
    ]);

    const csv = [headers, ...rows].map(row => row.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `trade-reconciliation-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-[#E8E9ED] overflow-hidden">
      {/* Header */}
      <div className="h-12 bg-[#1E2026] border-b border-[#383A42] px-4 flex items-center justify-between flex-shrink-0">
        <div className="flex items-center gap-3">
          <Activity size={18} className="text-[#3B82F6]" />
          <h2 className="text-sm font-bold text-[#E8E9ED]">Trade Reconciliation</h2>
          <div className="flex items-center gap-1.5 text-xs">
            <div className="w-1.5 h-1.5 rounded-full bg-[#22C55E] animate-pulse" />
            <span className="text-[#888]">Live</span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleAutoReconcile}
            disabled={isReconciling}
            className="px-3 py-1 text-xs font-medium bg-[#3B82F6] text-white rounded hover:bg-[#2563EB] disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1.5"
          >
            <RefreshCw size={12} className={isReconciling ? 'animate-spin' : ''} />
            {isReconciling ? 'Reconciling...' : 'Auto-Reconcile'}
          </button>
          <button
            onClick={handleExportCSV}
            className="px-3 py-1 text-xs font-medium bg-[#25272E] text-[#E8E9ED] rounded hover:bg-[#2D2F38] flex items-center gap-1.5"
          >
            <Download size={12} />
            Export CSV
          </button>
          <button
            onClick={() => setShowSettings(!showSettings)}
            className="px-3 py-1 text-xs font-medium bg-[#25272E] text-[#E8E9ED] rounded hover:bg-[#2D2F38] flex items-center gap-1.5"
          >
            <Settings size={12} />
            Thresholds
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="flex gap-3 p-4 flex-shrink-0">
        <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Total Trades Today</div>
          <div className="text-2xl font-bold text-[#E8E9ED]">{stats.total.toLocaleString()}</div>
        </div>
        <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Matched</div>
          <div className="text-2xl font-bold text-[#22C55E]">{stats.matched.toLocaleString()}</div>
        </div>
        <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Unmatched</div>
          <div className="text-2xl font-bold text-[#F59E0B]">{stats.unmatched.toLocaleString()}</div>
        </div>
        <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Discrepancies</div>
          <div className="text-2xl font-bold text-[#EF4444]">{stats.discrepancyCount.toLocaleString()}</div>
        </div>
        <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Reconciliation Rate</div>
          <div className="text-2xl font-bold text-[#3B82F6]">{stats.reconciliationRate.toFixed(1)}%</div>
        </div>
      </div>

      {/* Filters */}
      <div className="px-4 pb-3 flex gap-3 flex-shrink-0">
        <select
          value={selectedSymbol}
          onChange={(e) => setSelectedSymbol(e.target.value)}
          className="px-3 py-1.5 text-xs bg-[#1E2026] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
        >
          <option value="all">All Symbols</option>
          {SYMBOLS.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
        <select
          value={selectedLP}
          onChange={(e) => setSelectedLP(e.target.value)}
          className="px-3 py-1.5 text-xs bg-[#1E2026] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
        >
          <option value="all">All LPs</option>
          {LPS.map(lp => <option key={lp} value={lp}>{lp}</option>)}
        </select>
        <select
          value={selectedStatus}
          onChange={(e) => setSelectedStatus(e.target.value)}
          className="px-3 py-1.5 text-xs bg-[#1E2026] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
        >
          <option value="all">All Status</option>
          <option value="matched">Matched</option>
          <option value="price_mismatch">Price Mismatch</option>
          <option value="volume_mismatch">Volume Mismatch</option>
          <option value="missing_at_lp">Missing at LP</option>
          <option value="missing_internal">Missing Internal</option>
        </select>
        <input
          type="date"
          value={dateRange.start}
          onChange={(e) => setDateRange(prev => ({ ...prev, start: e.target.value }))}
          className="px-3 py-1.5 text-xs bg-[#1E2026] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
        />
        <input
          type="date"
          value={dateRange.end}
          onChange={(e) => setDateRange(prev => ({ ...prev, end: e.target.value }))}
          className="px-3 py-1.5 text-xs bg-[#1E2026] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
        />
      </div>

      {/* Main Content */}
      <div className="flex-1 flex gap-3 px-4 pb-4 overflow-hidden">
        {/* Discrepancy Table */}
        <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded overflow-hidden flex flex-col">
          <div className="px-3 py-2 border-b border-[#383A42] text-xs font-bold text-[#E8E9ED]">
            Trade Discrepancies ({filtered.length})
          </div>
          <div className="flex-1 overflow-auto">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#25272E] text-[#888] border-b border-[#383A42]">
                <tr>
                  <th className="px-3 py-2 text-left font-medium">Trade ID</th>
                  <th className="px-3 py-2 text-left font-medium">Symbol</th>
                  <th className="px-3 py-2 text-right font-medium">Internal Price</th>
                  <th className="px-3 py-2 text-right font-medium">LP Price</th>
                  <th className="px-3 py-2 text-right font-medium">Price Diff</th>
                  <th className="px-3 py-2 text-right font-medium">Volume</th>
                  <th className="px-3 py-2 text-left font-medium">Internal Time</th>
                  <th className="px-3 py-2 text-left font-medium">LP Time</th>
                  <th className="px-3 py-2 text-right font-medium">Time Diff</th>
                  <th className="px-3 py-2 text-left font-medium">Status</th>
                  <th className="px-3 py-2 text-left font-medium">LP</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((d, idx) => {
                  const config = STATUS_CONFIG[d.status];
                  const rowBg = d.status === 'matched' ? '#22C55E10' :
                                d.status === 'price_mismatch' || d.status === 'volume_mismatch' ? '#F59E0B10' :
                                d.status === 'missing_at_lp' ? '#EF444410' : '#6B728010';

                  return (
                    <tr key={idx} className="border-b border-[#383A42] hover:bg-[#25272E]" style={{ backgroundColor: rowBg }}>
                      <td className="px-3 py-2 text-[#E8E9ED] font-mono">{d.tradeId}</td>
                      <td className="px-3 py-2 text-[#E8E9ED] font-bold">{d.symbol}</td>
                      <td className="px-3 py-2 text-right text-[#E8E9ED] font-mono">{d.internalPrice.toFixed(5)}</td>
                      <td className="px-3 py-2 text-right text-[#E8E9ED] font-mono">{d.lpPrice?.toFixed(5) || 'N/A'}</td>
                      <td className="px-3 py-2 text-right text-[#F59E0B] font-mono">{d.priceDiff.toFixed(5)}</td>
                      <td className="px-3 py-2 text-right text-[#E8E9ED] font-mono">{d.volume.toFixed(2)}</td>
                      <td className="px-3 py-2 text-[#888] font-mono text-xs">{new Date(d.internalTime).toLocaleTimeString()}</td>
                      <td className="px-3 py-2 text-[#888] font-mono text-xs">{d.lpTime ? new Date(d.lpTime).toLocaleTimeString() : 'N/A'}</td>
                      <td className="px-3 py-2 text-right text-[#888] font-mono text-xs">{d.timeDiff}ms</td>
                      <td className="px-3 py-2">
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium" style={{ color: config.color, backgroundColor: config.bgColor }}>
                          {config.icon}
                          {config.label}
                        </span>
                      </td>
                      <td className="px-3 py-2 text-[#3B82F6] font-medium">{d.lpName}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>

        {/* LP Comparison Sidebar */}
        <div className="w-80 bg-[#1E2026] border border-[#383A42] rounded overflow-hidden flex flex-col">
          <div className="px-3 py-2 border-b border-[#383A42] text-xs font-bold text-[#E8E9ED] flex items-center gap-2">
            <TrendingUp size={14} className="text-[#3B82F6]" />
            LP Reconciliation Rates
          </div>
          <div className="flex-1 overflow-auto p-3 space-y-2">
            {lpStats.map((lp, idx) => (
              <div key={idx} className="bg-[#25272E] border border-[#383A42] rounded p-3">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-bold text-[#E8E9ED]">{lp.lpName}</span>
                  <span className="text-sm font-bold" style={{ color: lp.reconciliationRate >= 95 ? '#22C55E' : lp.reconciliationRate >= 85 ? '#F59E0B' : '#EF4444' }}>
                    {lp.reconciliationRate.toFixed(1)}%
                  </span>
                </div>
                <div className="space-y-1 text-xs text-[#888]">
                  <div className="flex justify-between">
                    <span>Total Trades:</span>
                    <span className="text-[#E8E9ED]">{lp.totalTrades}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Matched:</span>
                    <span className="text-[#22C55E]">{lp.matched}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Discrepancies:</span>
                    <span className="text-[#EF4444]">{lp.discrepancies}</span>
                  </div>
                </div>
                {/* Progress bar */}
                <div className="mt-2 h-1.5 bg-[#383A42] rounded-full overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all"
                    style={{
                      width: `${lp.reconciliationRate}%`,
                      backgroundColor: lp.reconciliationRate >= 95 ? '#22C55E' : lp.reconciliationRate >= 85 ? '#F59E0B' : '#EF4444'
                    }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Threshold Settings Modal */}
      {showSettings && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded w-96 p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-bold text-[#E8E9ED]">Reconciliation Thresholds</h3>
              <button onClick={() => setShowSettings(false)} className="text-[#888] hover:text-[#E8E9ED]">
                <XCircle size={16} />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="text-xs text-[#888] block mb-1">Price Tolerance</label>
                <input
                  type="number"
                  step="0.0001"
                  value={priceThreshold}
                  onChange={(e) => setPriceThreshold(parseFloat(e.target.value))}
                  className="w-full px-3 py-2 text-sm bg-[#25272E] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
                />
                <p className="text-xs text-[#888] mt-1">Trades within this price difference will be considered matched</p>
              </div>
              <div>
                <label className="text-xs text-[#888] block mb-1">Volume Tolerance (%)</label>
                <input
                  type="number"
                  step="0.01"
                  value={volumeThreshold}
                  onChange={(e) => setVolumeThreshold(parseFloat(e.target.value))}
                  className="w-full px-3 py-2 text-sm bg-[#25272E] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
                />
                <p className="text-xs text-[#888] mt-1">Trades within this volume difference will be considered matched</p>
              </div>
              <button
                onClick={() => setShowSettings(false)}
                className="w-full px-4 py-2 text-sm font-medium bg-[#3B82F6] text-white rounded hover:bg-[#2563EB]"
              >
                Save Settings
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
