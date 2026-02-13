/**
 * Trade Journal / History Panel
 * Detailed trade history view for reviewing past performance
 */

import { useState, useEffect, useMemo } from 'react';
import { Download, TrendingUp, TrendingDown, DollarSign, Target, Activity } from 'lucide-react';
import { API_BASE_URL } from '../config/api';
import { useAppStore } from '../store/useAppStore';

interface Trade {
  id: number;
  openTime: string;
  closeTime: string;
  symbol: string;
  type: 'BUY' | 'SELL';
  volume: number;
  openPrice: number;
  closePrice: number;
  sl: number;
  tp: number;
  commission: number;
  swap: number;
  profit: number;
}

type SortField = keyof Trade;
type SortDirection = 'asc' | 'desc';

export function TradeJournal() {
  const [trades, setTrades] = useState<Trade[]>([]);
  const [loading, setLoading] = useState(true);

  // Filters
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [symbolFilter, setSymbolFilter] = useState('ALL');
  const [typeFilter, setTypeFilter] = useState<'ALL' | 'BUY' | 'SELL'>('ALL');

  // Pagination
  const [currentPage, setCurrentPage] = useState(1);
  const tradesPerPage = 20;

  // Sorting
  const [sortField, setSortField] = useState<SortField>('id');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  // Fetch trades data
  useEffect(() => {
    fetchTrades();
  }, []);

  const fetchTrades = async () => {
    try {
      setLoading(true);
      const accountId = useAppStore.getState().accountId;
      const authToken = useAppStore.getState().authToken;
      const headers: Record<string, string> = { 'Content-Type': 'application/json' };
      if (authToken) headers['Authorization'] = `Bearer ${authToken}`;

      // Backend endpoint: /api/trades (GET)
      const response = await fetch(
        `${API_BASE_URL}/api/trades${accountId ? `?accountId=${accountId}` : ''}`,
        { headers }
      );
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      const raw = Array.isArray(data) ? data : data.trades || [];
      if (raw.length > 0) {
        const mapped: Trade[] = raw.map((t: any, i: number) => ({
          id: t.id || i + 1,
          openTime: t.openTime || t.open_time || t.time || '',
          closeTime: t.closeTime || t.close_time || t.time || '',
          symbol: t.symbol || '',
          type: t.type || t.side || 'BUY',
          volume: t.volume ?? t.lots ?? 0,
          openPrice: t.openPrice ?? t.open_price ?? 0,
          closePrice: t.closePrice ?? t.close_price ?? 0,
          sl: t.sl ?? t.stopLoss ?? 0,
          tp: t.tp ?? t.takeProfit ?? 0,
          commission: t.commission ?? 0,
          swap: t.swap ?? 0,
          profit: t.profit ?? 0,
        }));
        setTrades(mapped.sort((a, b) => new Date(b.closeTime).getTime() - new Date(a.closeTime).getTime()));
      } else {
        setTrades(generateMockTrades());
      }
    } catch (error) {
      console.error('Failed to fetch trades:', error);
      // Use mock data as fallback
      setTrades(generateMockTrades());
    } finally {
      setLoading(false);
    }
  };

  // Filter trades
  const filteredTrades = useMemo(() => {
    return trades.filter(trade => {
      // Date filter
      if (fromDate && new Date(trade.openTime) < new Date(fromDate)) return false;
      if (toDate && new Date(trade.closeTime) > new Date(toDate)) return false;

      // Symbol filter
      if (symbolFilter !== 'ALL' && trade.symbol !== symbolFilter) return false;

      // Type filter
      if (typeFilter !== 'ALL' && trade.type !== typeFilter) return false;

      return true;
    });
  }, [trades, fromDate, toDate, symbolFilter, typeFilter]);

  // Sort trades
  const sortedTrades = useMemo(() => {
    const sorted = [...filteredTrades];
    sorted.sort((a, b) => {
      const aVal = a[sortField];
      const bVal = b[sortField];

      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortDirection === 'asc'
          ? aVal.localeCompare(bVal)
          : bVal.localeCompare(aVal);
      }

      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
      }

      return 0;
    });
    return sorted;
  }, [filteredTrades, sortField, sortDirection]);

  // Paginate trades
  const paginatedTrades = useMemo(() => {
    const startIndex = (currentPage - 1) * tradesPerPage;
    return sortedTrades.slice(startIndex, startIndex + tradesPerPage);
  }, [sortedTrades, currentPage]);

  const totalPages = Math.ceil(sortedTrades.length / tradesPerPage);

  // Calculate summary statistics
  const summary = useMemo(() => {
    const totalTrades = filteredTrades.length;
    const totalPL = filteredTrades.reduce((sum, t) => sum + t.profit, 0);
    const winningTrades = filteredTrades.filter(t => t.profit > 0);
    const losingTrades = filteredTrades.filter(t => t.profit < 0);
    const winRate = totalTrades > 0 ? (winningTrades.length / totalTrades) * 100 : 0;
    const avgWin = winningTrades.length > 0
      ? winningTrades.reduce((sum, t) => sum + t.profit, 0) / winningTrades.length
      : 0;
    const avgLoss = losingTrades.length > 0
      ? losingTrades.reduce((sum, t) => sum + t.profit, 0) / losingTrades.length
      : 0;

    return { totalTrades, totalPL, winRate, avgWin, avgLoss };
  }, [filteredTrades]);

  // Get unique symbols for filter
  const symbols = useMemo(() => {
    const uniqueSymbols = new Set(trades.map(t => t.symbol));
    return ['ALL', ...Array.from(uniqueSymbols)];
  }, [trades]);

  // Handle column header click for sorting
  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  // Export to CSV
  const exportToCSV = () => {
    const headers = ['#', 'Open Time', 'Close Time', 'Symbol', 'Type', 'Volume', 'Open Price', 'Close Price', 'SL', 'TP', 'Commission', 'Swap', 'P&L'];
    const rows = filteredTrades.map(t => [
      t.id,
      t.openTime,
      t.closeTime,
      t.symbol,
      t.type,
      t.volume,
      t.openPrice,
      t.closePrice,
      t.sl,
      t.tp,
      t.commission,
      t.swap,
      t.profit,
    ]);

    const csv = [
      headers.join(','),
      ...rows.map(row => row.join(','))
    ].join('\n');

    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `trade_journal_${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  // Get last 20 trades for chart
  const chartTrades = useMemo(() => {
    return sortedTrades.slice(0, 20);
  }, [sortedTrades]);

  const maxAbsProfit = useMemo(() => {
    return Math.max(...chartTrades.map(t => Math.abs(t.profit)), 1);
  }, [chartTrades]);

  return (
    <div className="flex flex-col h-full bg-[#1e1e1e] text-zinc-300 overflow-hidden">
      {/* Header */}
      <div className="bg-[#252528] border-b border-zinc-700 p-4">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">Trade Journal</h2>
          <button
            onClick={exportToCSV}
            className="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs font-medium transition-colors"
          >
            <Download size={14} />
            Export CSV
          </button>
        </div>

        {/* Filters */}
        <div className="flex items-center gap-3 flex-wrap text-xs">
          <div className="flex items-center gap-2">
            <label className="text-zinc-400">From:</label>
            <input
              type="date"
              value={fromDate}
              onChange={(e) => setFromDate(e.target.value)}
              className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1 outline-none focus:border-blue-500"
            />
          </div>

          <div className="flex items-center gap-2">
            <label className="text-zinc-400">To:</label>
            <input
              type="date"
              value={toDate}
              onChange={(e) => setToDate(e.target.value)}
              className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1 outline-none focus:border-blue-500"
            />
          </div>

          <div className="flex items-center gap-2">
            <label className="text-zinc-400">Symbol:</label>
            <select
              value={symbolFilter}
              onChange={(e) => setSymbolFilter(e.target.value)}
              className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1 outline-none focus:border-blue-500"
            >
              {symbols.map(sym => (
                <option key={sym} value={sym}>{sym}</option>
              ))}
            </select>
          </div>

          <div className="flex items-center gap-2">
            <label className="text-zinc-400">Type:</label>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value as 'ALL' | 'BUY' | 'SELL')}
              className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1 outline-none focus:border-blue-500"
            >
              <option value="ALL">All</option>
              <option value="BUY">Buy</option>
              <option value="SELL">Sell</option>
            </select>
          </div>

          <button
            onClick={() => {
              setFromDate('');
              setToDate('');
              setSymbolFilter('ALL');
              setTypeFilter('ALL');
            }}
            className="px-3 py-1 bg-zinc-700 hover:bg-zinc-600 rounded transition-colors"
          >
            Clear Filters
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-5 gap-3 p-4 bg-[#1e1e1e]">
        <SummaryCard
          icon={<Activity size={18} />}
          label="Total Trades"
          value={summary.totalTrades.toString()}
          color="text-blue-400"
        />
        <SummaryCard
          icon={<DollarSign size={18} />}
          label="Total P&L"
          value={`$${summary.totalPL.toFixed(2)}`}
          color={summary.totalPL >= 0 ? 'text-green-400' : 'text-red-400'}
        />
        <SummaryCard
          icon={<Target size={18} />}
          label="Win Rate"
          value={`${summary.winRate.toFixed(1)}%`}
          color="text-purple-400"
        />
        <SummaryCard
          icon={<TrendingUp size={18} />}
          label="Average Win"
          value={`$${summary.avgWin.toFixed(2)}`}
          color="text-green-400"
        />
        <SummaryCard
          icon={<TrendingDown size={18} />}
          label="Average Loss"
          value={`$${summary.avgLoss.toFixed(2)}`}
          color="text-red-400"
        />
      </div>

      {/* Trades Table */}
      <div className="flex-1 overflow-auto">
        <table className="w-full text-xs">
          <thead className="bg-[#252528] sticky top-0 z-10">
            <tr className="border-b border-zinc-700">
              <SortableHeader label="#" field="id" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="Open Time" field="openTime" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="Close Time" field="closeTime" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="Symbol" field="symbol" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="Type" field="type" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="Volume" field="volume" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="Open Price" field="openPrice" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="Close Price" field="closePrice" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="SL" field="sl" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="TP" field="tp" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="Commission" field="commission" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="Swap" field="swap" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
              <SortableHeader label="P&L" field="profit" sortField={sortField} sortDirection={sortDirection} onSort={handleSort} />
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={13} className="text-center py-8 text-zinc-500">
                  Loading trades...
                </td>
              </tr>
            ) : paginatedTrades.length === 0 ? (
              <tr>
                <td colSpan={13} className="text-center py-8 text-zinc-500">
                  No trades found
                </td>
              </tr>
            ) : (
              paginatedTrades.map((trade) => (
                <tr key={trade.id} className="border-b border-zinc-800 hover:bg-zinc-800/50">
                  <td className="px-3 py-2 text-zinc-400">{trade.id}</td>
                  <td className="px-3 py-2">{formatDateTime(trade.openTime)}</td>
                  <td className="px-3 py-2">{formatDateTime(trade.closeTime)}</td>
                  <td className="px-3 py-2 font-medium">{trade.symbol}</td>
                  <td className={`px-3 py-2 font-medium ${trade.type === 'BUY' ? 'text-green-400' : 'text-red-400'}`}>
                    {trade.type}
                  </td>
                  <td className="px-3 py-2">{trade.volume.toFixed(2)}</td>
                  <td className="px-3 py-2">{trade.openPrice.toFixed(5)}</td>
                  <td className="px-3 py-2">{trade.closePrice.toFixed(5)}</td>
                  <td className="px-3 py-2">{trade.sl > 0 ? trade.sl.toFixed(5) : '-'}</td>
                  <td className="px-3 py-2">{trade.tp > 0 ? trade.tp.toFixed(5) : '-'}</td>
                  <td className="px-3 py-2 text-red-400">{trade.commission.toFixed(2)}</td>
                  <td className="px-3 py-2">{trade.swap.toFixed(2)}</td>
                  <td className={`px-3 py-2 font-medium ${trade.profit >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                    ${trade.profit.toFixed(2)}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="bg-[#252528] border-t border-zinc-700 px-4 py-2 flex items-center justify-between text-xs">
          <div className="text-zinc-400">
            Showing {((currentPage - 1) * tradesPerPage) + 1} - {Math.min(currentPage * tradesPerPage, sortedTrades.length)} of {sortedTrades.length} trades
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
              disabled={currentPage === 1}
              className="px-3 py-1 bg-zinc-700 hover:bg-zinc-600 rounded disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Previous
            </button>
            <span className="text-zinc-300">
              Page {currentPage} of {totalPages}
            </span>
            <button
              onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
              className="px-3 py-1 bg-zinc-700 hover:bg-zinc-600 rounded disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Next
            </button>
          </div>
        </div>
      )}

      {/* P&L Chart */}
      <div className="bg-[#252528] border-t border-zinc-700 p-4">
        <h3 className="text-xs font-semibold text-zinc-400 mb-3">Last 20 Trades P&L</h3>
        <div className="flex items-end gap-1 h-24">
          {chartTrades.length === 0 ? (
            <div className="w-full text-center text-zinc-500 text-xs">No trades to display</div>
          ) : (
            chartTrades.map((trade) => {
              const height = Math.abs(trade.profit) / maxAbsProfit * 100;
              const isProfit = trade.profit >= 0;
              return (
                <div
                  key={trade.id}
                  className="flex-1 relative group"
                  title={`#${trade.id}: $${trade.profit.toFixed(2)}`}
                >
                  <div
                    className={`w-full ${isProfit ? 'bg-green-500' : 'bg-red-500'} rounded-sm transition-all group-hover:opacity-80`}
                    style={{ height: `${height}%`, minHeight: '2px' }}
                  />
                  <div className="absolute -top-6 left-1/2 -translate-x-1/2 bg-zinc-900 px-1 py-0.5 rounded text-[10px] opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
                    ${trade.profit.toFixed(2)}
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}

// Helper Components
interface SummaryCardProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  color: string;
}

function SummaryCard({ icon, label, value, color }: SummaryCardProps) {
  return (
    <div className="bg-[#252528] border border-zinc-700 rounded p-3">
      <div className="flex items-center gap-2 mb-1">
        <div className={color}>{icon}</div>
        <span className="text-[10px] text-zinc-400 uppercase">{label}</span>
      </div>
      <div className={`text-lg font-bold ${color}`}>{value}</div>
    </div>
  );
}

interface SortableHeaderProps {
  label: string;
  field: SortField;
  sortField: SortField;
  sortDirection: SortDirection;
  onSort: (field: SortField) => void;
}

function SortableHeader({ label, field, sortField, sortDirection, onSort }: SortableHeaderProps) {
  const isActive = sortField === field;
  return (
    <th
      onClick={() => onSort(field)}
      className="px-3 py-2 text-left font-medium text-zinc-400 cursor-pointer hover:text-white transition-colors select-none"
    >
      <div className="flex items-center gap-1">
        {label}
        {isActive && (
          <span className="text-blue-400">
            {sortDirection === 'asc' ? '↑' : '↓'}
          </span>
        )}
      </div>
    </th>
  );
}

// Helper Functions
function formatDateTime(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleString('en-US', {
    month: '2-digit',
    day: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

// Mock data generator for development
function generateMockTrades(): Trade[] {
  const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD'];
  const types: ('BUY' | 'SELL')[] = ['BUY', 'SELL'];
  const trades: Trade[] = [];

  for (let i = 1; i <= 50; i++) {
    const symbol = symbols[Math.floor(Math.random() * symbols.length)];
    const type = types[Math.floor(Math.random() * types.length)];
    const volume = Math.random() * 2 + 0.1;
    const openPrice = 1.1 + Math.random() * 0.05;
    const priceChange = (Math.random() - 0.5) * 0.01;
    const closePrice = openPrice + priceChange;
    const profit = (type === 'BUY' ? 1 : -1) * (closePrice - openPrice) * volume * 100000;

    trades.push({
      id: i,
      openTime: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
      closeTime: new Date(Date.now() - Math.random() * 25 * 24 * 60 * 60 * 1000).toISOString(),
      symbol,
      type,
      volume,
      openPrice,
      closePrice,
      sl: Math.random() > 0.5 ? openPrice - (type === 'BUY' ? 0.002 : -0.002) : 0,
      tp: Math.random() > 0.5 ? openPrice + (type === 'BUY' ? 0.003 : -0.003) : 0,
      commission: -Math.random() * 5,
      swap: (Math.random() - 0.5) * 2,
      profit,
    });
  }

  return trades.sort((a, b) => new Date(b.closeTime).getTime() - new Date(a.closeTime).getTime());
}
