'use client';

import React, { useState, useMemo } from 'react';
import {
  DollarSign,
  TrendingUp,
  Users,
  PieChart,
  Download,
  ChevronDown,
  ChevronUp,
  Calendar,
  Search,
  Filter,
  X,
} from 'lucide-react';

type CommissionModel = 'per-lot' | 'percentage' | 'tiered';
type SymbolGroup = 'forex' | 'metals' | 'crypto' | 'indices';

interface CommissionRecord {
  id: string;
  date: number;
  clientId: string;
  clientName: string;
  symbol: string;
  symbolGroup: SymbolGroup;
  volume: number; // lots
  commissionModel: CommissionModel;
  rate: number; // USD per lot or %
  commissionAmount: number; // USD
  tradeId: string;
  groupName: string;
}

type SortColumn = keyof CommissionRecord | null;
type SortDirection = 'asc' | 'desc';

export default function CommissionReport() {
  const [dateFrom, setDateFrom] = useState<string>(
    new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  );
  const [dateTo, setDateTo] = useState<string>(new Date().toISOString().split('T')[0]);
  const [clientSearch, setClientSearch] = useState<string>('');
  const [symbolFilter, setSymbolFilter] = useState<string>('all');
  const [modelFilter, setModelFilter] = useState<CommissionModel | 'all'>('all');
  const [groupFilter, setGroupFilter] = useState<string>('all');

  const [sortColumn, setSortColumn] = useState<SortColumn>('date');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
  const [expandedRow, setExpandedRow] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(25);

  // Mock data: 30+ commission records
  const allRecords: CommissionRecord[] = useMemo(() => {
    const records: CommissionRecord[] = [];
    const clients = ['John Doe', 'Jane Smith', 'Alice Brown', 'Bob Wilson', 'Charlie Davis', 'Eva Martinez', 'Frank Taylor', 'Grace Lee', 'Henry Clark', 'Ivy Rodriguez'];
    const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'GOLD', 'SILVER', 'BTCUSD', 'ETHUSD', 'SPX500'];
    const symbolGroups: SymbolGroup[] = ['forex', 'metals', 'crypto', 'indices'];
    const models: CommissionModel[] = ['per-lot', 'percentage', 'tiered'];
    const groups = ['Standard', 'Premium', 'VIP', 'Institutional'];

    for (let i = 0; i < 35; i++) {
      const client = clients[Math.floor(Math.random() * clients.length)];
      const symbol = symbols[Math.floor(Math.random() * symbols.length)];
      let symbolGroup: SymbolGroup;
      if (symbol.includes('USD') || symbol.includes('GBP') || symbol.includes('EUR') || symbol.includes('JPY') || symbol.includes('AUD') || symbol.includes('CAD')) {
        symbolGroup = 'forex';
      } else if (symbol === 'GOLD' || symbol === 'SILVER') {
        symbolGroup = 'metals';
      } else if (symbol.includes('BTC') || symbol.includes('ETH')) {
        symbolGroup = 'crypto';
      } else {
        symbolGroup = 'indices';
      }

      const model = models[Math.floor(Math.random() * models.length)];
      const volume = parseFloat((Math.random() * 10 + 0.1).toFixed(2));
      let rate: number;
      let commissionAmount: number;

      if (model === 'per-lot') {
        rate = parseFloat((Math.random() * 5 + 2).toFixed(2)); // $2-7 per lot
        commissionAmount = parseFloat((volume * rate).toFixed(2));
      } else if (model === 'percentage') {
        rate = parseFloat((Math.random() * 0.05 + 0.01).toFixed(4)); // 0.01% - 0.06%
        const tradeValue = volume * 100000; // assume standard lot
        commissionAmount = parseFloat((tradeValue * rate).toFixed(2));
      } else {
        // tiered
        rate = parseFloat((Math.random() * 3 + 1).toFixed(2));
        commissionAmount = parseFloat((volume * rate * (volume > 5 ? 0.8 : 1)).toFixed(2));
      }

      records.push({
        id: `comm-${1000 + i}`,
        date: Date.now() - Math.floor(Math.random() * 30) * 24 * 60 * 60 * 1000,
        clientId: `client-${100 + i}`,
        clientName: client,
        symbol,
        symbolGroup,
        volume,
        commissionModel: model,
        rate,
        commissionAmount,
        tradeId: `trade-${5000 + i}`,
        groupName: groups[Math.floor(Math.random() * groups.length)],
      });
    }

    return records.sort((a, b) => b.date - a.date);
  }, []);

  // Filtered records
  const filteredRecords = useMemo(() => {
    let filtered = allRecords.filter((record) => {
      const recordDate = new Date(record.date).toISOString().split('T')[0];
      if (recordDate < dateFrom || recordDate > dateTo) return false;
      if (clientSearch && !record.clientName.toLowerCase().includes(clientSearch.toLowerCase())) return false;
      if (symbolFilter !== 'all' && record.symbol !== symbolFilter) return false;
      if (modelFilter !== 'all' && record.commissionModel !== modelFilter) return false;
      if (groupFilter !== 'all' && record.groupName !== groupFilter) return false;
      return true;
    });

    // Sort
    if (sortColumn) {
      filtered.sort((a, b) => {
        const aVal = a[sortColumn];
        const bVal = b[sortColumn];
        if (typeof aVal === 'number' && typeof bVal === 'number') {
          return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
        }
        if (typeof aVal === 'string' && typeof bVal === 'string') {
          return sortDirection === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
        }
        return 0;
      });
    }

    return filtered;
  }, [allRecords, dateFrom, dateTo, clientSearch, symbolFilter, modelFilter, groupFilter, sortColumn, sortDirection]);

  // Pagination
  const totalPages = Math.ceil(filteredRecords.length / pageSize);
  const paginatedRecords = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filteredRecords.slice(start, start + pageSize);
  }, [filteredRecords, currentPage, pageSize]);

  // Summary calculations
  const summary = useMemo(() => {
    const now = Date.now();
    const oneDayAgo = now - 24 * 60 * 60 * 1000;
    const oneWeekAgo = now - 7 * 24 * 60 * 60 * 1000;
    const oneMonthAgo = now - 30 * 24 * 60 * 60 * 1000;

    const todayRecords = filteredRecords.filter((r) => r.date >= oneDayAgo);
    const weekRecords = filteredRecords.filter((r) => r.date >= oneWeekAgo);
    const monthRecords = filteredRecords.filter((r) => r.date >= oneMonthAgo);

    const totalToday = todayRecords.reduce((sum, r) => sum + r.commissionAmount, 0);
    const totalWeek = weekRecords.reduce((sum, r) => sum + r.commissionAmount, 0);
    const totalMonth = monthRecords.reduce((sum, r) => sum + r.commissionAmount, 0);

    const avgPerTrade = filteredRecords.length > 0
      ? filteredRecords.reduce((sum, r) => sum + r.commissionAmount, 0) / filteredRecords.length
      : 0;

    // Top commission generator
    const clientTotals = new Map<string, number>();
    filteredRecords.forEach((r) => {
      clientTotals.set(r.clientName, (clientTotals.get(r.clientName) || 0) + r.commissionAmount);
    });
    let topClient = { name: 'N/A', amount: 0 };
    clientTotals.forEach((amount, name) => {
      if (amount > topClient.amount) {
        topClient = { name, amount };
      }
    });

    // Commission by model
    const modelTotals = {
      'per-lot': 0,
      'percentage': 0,
      'tiered': 0,
    };
    filteredRecords.forEach((r) => {
      modelTotals[r.commissionModel] += r.commissionAmount;
    });

    return {
      totalToday,
      totalWeek,
      totalMonth,
      avgPerTrade,
      topClient,
      modelTotals,
    };
  }, [filteredRecords]);

  // Daily trend data (last 30 days)
  const dailyTrend = useMemo(() => {
    const days: { date: string; total: number }[] = [];
    const now = new Date();
    for (let i = 29; i >= 0; i--) {
      const date = new Date(now.getTime() - i * 24 * 60 * 60 * 1000);
      const dateStr = date.toISOString().split('T')[0];
      const dayRecords = filteredRecords.filter((r) => {
        const rDate = new Date(r.date).toISOString().split('T')[0];
        return rDate === dateStr;
      });
      const total = dayRecords.reduce((sum, r) => sum + r.commissionAmount, 0);
      days.push({ date: dateStr, total });
    }
    return days;
  }, [filteredRecords]);

  // Top 10 clients by commission
  const topClients = useMemo(() => {
    const clientTotals = new Map<string, number>();
    filteredRecords.forEach((r) => {
      clientTotals.set(r.clientName, (clientTotals.get(r.clientName) || 0) + r.commissionAmount);
    });
    const sorted = Array.from(clientTotals.entries())
      .map(([name, amount]) => ({ name, amount }))
      .sort((a, b) => b.amount - a.amount)
      .slice(0, 10);
    return sorted;
  }, [filteredRecords]);

  // Commission by symbol group
  const groupTotals = useMemo(() => {
    const totals: Record<SymbolGroup, number> = {
      forex: 0,
      metals: 0,
      crypto: 0,
      indices: 0,
    };
    filteredRecords.forEach((r) => {
      totals[r.symbolGroup] += r.commissionAmount;
    });
    return totals;
  }, [filteredRecords]);

  const handleSort = (column: SortColumn) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('desc');
    }
  };

  const SortIcon = ({ column }: { column: SortColumn }) => {
    if (sortColumn !== column) return null;
    return sortDirection === 'asc' ? (
      <ChevronUp size={12} className="inline ml-1" />
    ) : (
      <ChevronDown size={12} className="inline ml-1" />
    );
  };

  const exportToCSV = () => {
    const headers = ['Date', 'Client', 'Symbol', 'Volume (lots)', 'Commission Model', 'Rate', 'Commission Amount', 'Trade ID'];
    const rows = filteredRecords.map((r) => [
      new Date(r.date).toLocaleDateString(),
      r.clientName,
      r.symbol,
      r.volume.toFixed(2),
      r.commissionModel,
      r.rate.toFixed(4),
      r.commissionAmount.toFixed(2),
      r.tradeId,
    ]);

    const csv = [headers, ...rows].map((row) => row.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `commissions_${dateFrom}_${dateTo}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value);
  };

  const formatDate = (timestamp: number) => {
    return new Date(timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const getModelColor = (model: CommissionModel) => {
    switch (model) {
      case 'per-lot': return 'text-emerald-400';
      case 'percentage': return 'text-blue-400';
      case 'tiered': return 'text-purple-400';
    }
  };

  const getModelBadge = (model: CommissionModel) => {
    const colors = {
      'per-lot': 'bg-emerald-500/20 text-emerald-400',
      'percentage': 'bg-blue-500/20 text-blue-400',
      'tiered': 'bg-purple-500/20 text-purple-400',
    };
    return (
      <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${colors[model]}`}>
        {model.toUpperCase()}
      </span>
    );
  };

  // Pie chart for commission by model
  const PieChartWidget = () => {
    const total = summary.modelTotals['per-lot'] + summary.modelTotals['percentage'] + summary.modelTotals['tiered'];
    if (total === 0) return <div className="text-zinc-500 text-xs">No data</div>;

    const perLotPct = (summary.modelTotals['per-lot'] / total) * 100;
    const percentagePct = (summary.modelTotals['percentage'] / total) * 100;
    const tieredPct = (summary.modelTotals['tiered'] / total) * 100;

    return (
      <div className="flex items-center gap-3">
        <div className="relative w-16 h-16">
          <svg viewBox="0 0 100 100" className="transform -rotate-90">
            <circle cx="50" cy="50" r="40" fill="none" stroke="#10b981" strokeWidth="20" strokeDasharray={`${perLotPct * 2.51} 251`} />
            <circle cx="50" cy="50" r="40" fill="none" stroke="#3b82f6" strokeWidth="20" strokeDasharray={`${percentagePct * 2.51} 251`} strokeDashoffset={`-${perLotPct * 2.51}`} />
            <circle cx="50" cy="50" r="40" fill="none" stroke="#a855f7" strokeWidth="20" strokeDasharray={`${tieredPct * 2.51} 251`} strokeDashoffset={`-${(perLotPct + percentagePct) * 2.51}`} />
          </svg>
        </div>
        <div className="flex flex-col gap-1 text-[10px]">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-emerald-500 rounded-full"></div>
            <span className="text-zinc-400">Per-Lot {perLotPct.toFixed(0)}%</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
            <span className="text-zinc-400">Percentage {percentagePct.toFixed(0)}%</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-purple-500 rounded-full"></div>
            <span className="text-zinc-400">Tiered {tieredPct.toFixed(0)}%</span>
          </div>
        </div>
      </div>
    );
  };

  // Daily trend bar chart
  const DailyTrendChart = () => {
    const maxValue = Math.max(...dailyTrend.map((d) => d.total), 1);
    return (
      <div className="space-y-2">
        <div className="text-xs font-bold text-zinc-400">Daily Commission Trend (Last 30 Days)</div>
        <div className="flex items-end gap-0.5 h-32">
          {dailyTrend.map((day, index) => {
            const heightPct = (day.total / maxValue) * 100;
            return (
              <div key={index} className="flex-1 flex flex-col items-center group relative">
                <div
                  className="w-full bg-blue-500 hover:bg-blue-400 transition-colors cursor-pointer"
                  style={{ height: `${heightPct}%`, minHeight: day.total > 0 ? '2px' : '0px' }}
                  title={`${day.date}: ${formatCurrency(day.total)}`}
                ></div>
                <div className="absolute bottom-0 left-1/2 -translate-x-1/2 translate-y-full opacity-0 group-hover:opacity-100 bg-zinc-800 text-zinc-300 text-[9px] px-1 py-0.5 rounded mt-1 whitespace-nowrap pointer-events-none z-10">
                  {day.date.split('-')[2]}/{day.date.split('-')[1]}: {formatCurrency(day.total)}
                </div>
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-zinc-500">
          <span>{dailyTrend[0]?.date}</span>
          <span>{dailyTrend[dailyTrend.length - 1]?.date}</span>
        </div>
      </div>
    );
  };

  // Top 10 clients horizontal bar chart
  const TopClientsChart = () => {
    const maxValue = Math.max(...topClients.map((c) => c.amount), 1);
    return (
      <div className="space-y-2">
        <div className="text-xs font-bold text-zinc-400">Top 10 Clients by Commission</div>
        <div className="space-y-1.5">
          {topClients.map((client, index) => {
            const widthPct = (client.amount / maxValue) * 100;
            return (
              <div key={index} className="flex items-center gap-2">
                <div className="w-20 text-[10px] text-zinc-400 truncate">{client.name}</div>
                <div className="flex-1 bg-zinc-800 h-5 rounded overflow-hidden relative">
                  <div
                    className="bg-emerald-500 h-full transition-all"
                    style={{ width: `${widthPct}%` }}
                  ></div>
                  <div className="absolute inset-0 flex items-center justify-end px-2 text-[9px] text-zinc-300 font-bold">
                    {formatCurrency(client.amount)}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  // Commission by symbol group chart
  const SymbolGroupChart = () => {
    const total = Object.values(groupTotals).reduce((sum, val) => sum + val, 0);
    const groups: { name: string; amount: number; color: string }[] = [
      { name: 'Forex', amount: groupTotals.forex, color: 'bg-blue-500' },
      { name: 'Metals', amount: groupTotals.metals, color: 'bg-yellow-500' },
      { name: 'Crypto', amount: groupTotals.crypto, color: 'bg-purple-500' },
      { name: 'Indices', amount: groupTotals.indices, color: 'bg-emerald-500' },
    ];

    return (
      <div className="space-y-2">
        <div className="text-xs font-bold text-zinc-400">Commission by Symbol Group</div>
        <div className="space-y-1.5">
          {groups.map((group, index) => {
            const pct = total > 0 ? (group.amount / total) * 100 : 0;
            return (
              <div key={index} className="flex items-center gap-2">
                <div className="w-16 text-[10px] text-zinc-400">{group.name}</div>
                <div className="flex-1 bg-zinc-800 h-5 rounded overflow-hidden relative">
                  <div
                    className={`${group.color} h-full transition-all`}
                    style={{ width: `${pct}%` }}
                  ></div>
                  <div className="absolute inset-0 flex items-center justify-end px-2 text-[9px] text-zinc-300 font-bold">
                    {formatCurrency(group.amount)} ({pct.toFixed(0)}%)
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-zinc-300 overflow-hidden">
      {/* Header */}
      <div className="px-4 py-3 bg-[#1E2026] border-b border-zinc-700 flex items-center justify-between flex-shrink-0">
        <h2 className="text-sm font-bold text-zinc-300">Commission Report</h2>
        <div className="text-[10px] text-zinc-500">
          {filteredRecords.length} records | Total: {formatCurrency(filteredRecords.reduce((sum, r) => sum + r.commissionAmount, 0))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-4">
        {/* Summary Cards */}
        <div className="grid grid-cols-4 gap-3">
          {/* Total Commissions */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3 space-y-2">
            <div className="flex items-center gap-2">
              <DollarSign size={14} className="text-emerald-400" />
              <span className="text-[10px] text-zinc-500 font-bold uppercase">Total Commissions</span>
            </div>
            <div className="space-y-1">
              <div className="flex items-baseline gap-2">
                <span className="text-xs text-zinc-500">Today:</span>
                <span className="text-sm font-bold text-emerald-400">{formatCurrency(summary.totalToday)}</span>
              </div>
              <div className="flex items-baseline gap-2">
                <span className="text-xs text-zinc-500">Week:</span>
                <span className="text-sm font-bold text-blue-400">{formatCurrency(summary.totalWeek)}</span>
              </div>
              <div className="flex items-baseline gap-2">
                <span className="text-xs text-zinc-500">Month:</span>
                <span className="text-sm font-bold text-purple-400">{formatCurrency(summary.totalMonth)}</span>
              </div>
            </div>
          </div>

          {/* Avg Per Trade */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3 space-y-2">
            <div className="flex items-center gap-2">
              <TrendingUp size={14} className="text-blue-400" />
              <span className="text-[10px] text-zinc-500 font-bold uppercase">Avg Per Trade</span>
            </div>
            <div className="text-2xl font-bold text-blue-400">
              {formatCurrency(summary.avgPerTrade)}
            </div>
          </div>

          {/* Top Client */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3 space-y-2">
            <div className="flex items-center gap-2">
              <Users size={14} className="text-yellow-400" />
              <span className="text-[10px] text-zinc-500 font-bold uppercase">Top Generator</span>
            </div>
            <div className="space-y-1">
              <div className="text-sm font-bold text-zinc-300 truncate">{summary.topClient.name}</div>
              <div className="text-lg font-bold text-yellow-400">{formatCurrency(summary.topClient.amount)}</div>
            </div>
          </div>

          {/* Commission by Model */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3 space-y-2">
            <div className="flex items-center gap-2">
              <PieChart size={14} className="text-purple-400" />
              <span className="text-[10px] text-zinc-500 font-bold uppercase">By Model</span>
            </div>
            <PieChartWidget />
          </div>
        </div>

        {/* Filter Bar */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="grid grid-cols-6 gap-3">
            {/* Date Range */}
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase flex items-center gap-1">
                <Calendar size={10} />
                From
              </label>
              <input
                type="date"
                value={dateFrom}
                onChange={(e) => setDateFrom(e.target.value)}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              />
            </div>

            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase flex items-center gap-1">
                <Calendar size={10} />
                To
              </label>
              <input
                type="date"
                value={dateTo}
                onChange={(e) => setDateTo(e.target.value)}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              />
            </div>

            {/* Client Search */}
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase flex items-center gap-1">
                <Search size={10} />
                Client
              </label>
              <div className="relative">
                <input
                  type="text"
                  placeholder="Search client..."
                  value={clientSearch}
                  onChange={(e) => setClientSearch(e.target.value)}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
                />
                {clientSearch && (
                  <button
                    onClick={() => setClientSearch('')}
                    className="absolute right-1 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300"
                  >
                    <X size={12} />
                  </button>
                )}
              </div>
            </div>

            {/* Symbol Filter */}
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase flex items-center gap-1">
                <Filter size={10} />
                Symbol
              </label>
              <select
                value={symbolFilter}
                onChange={(e) => setSymbolFilter(e.target.value)}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              >
                <option value="all">All Symbols</option>
                <option value="EURUSD">EURUSD</option>
                <option value="GBPUSD">GBPUSD</option>
                <option value="USDJPY">USDJPY</option>
                <option value="GOLD">GOLD</option>
                <option value="BTCUSD">BTCUSD</option>
              </select>
            </div>

            {/* Model Filter */}
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase flex items-center gap-1">
                <Filter size={10} />
                Model
              </label>
              <select
                value={modelFilter}
                onChange={(e) => setModelFilter(e.target.value as CommissionModel | 'all')}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              >
                <option value="all">All Models</option>
                <option value="per-lot">Per-Lot</option>
                <option value="percentage">Percentage</option>
                <option value="tiered">Tiered</option>
              </select>
            </div>

            {/* Group Filter */}
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase flex items-center gap-1">
                <Filter size={10} />
                Group
              </label>
              <select
                value={groupFilter}
                onChange={(e) => setGroupFilter(e.target.value)}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              >
                <option value="all">All Groups</option>
                <option value="Standard">Standard</option>
                <option value="Premium">Premium</option>
                <option value="VIP">VIP</option>
                <option value="Institutional">Institutional</option>
              </select>
            </div>
          </div>

          {/* Export Button */}
          <div className="mt-3 flex justify-end">
            <button
              onClick={exportToCSV}
              className="flex items-center gap-2 px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold rounded transition-colors"
            >
              <Download size={12} />
              Export to CSV
            </button>
          </div>
        </div>

        {/* Commission Table */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="bg-[#252525] border-b border-zinc-700">
                  <th
                    className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('date')}
                  >
                    Date <SortIcon column="date" />
                  </th>
                  <th
                    className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('clientName')}
                  >
                    Client <SortIcon column="clientName" />
                  </th>
                  <th
                    className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('symbol')}
                  >
                    Symbol <SortIcon column="symbol" />
                  </th>
                  <th
                    className="px-3 py-2 text-right text-[10px] font-bold text-zinc-400 uppercase cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('volume')}
                  >
                    Volume (lots) <SortIcon column="volume" />
                  </th>
                  <th
                    className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('commissionModel')}
                  >
                    Model <SortIcon column="commissionModel" />
                  </th>
                  <th
                    className="px-3 py-2 text-right text-[10px] font-bold text-zinc-400 uppercase cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('rate')}
                  >
                    Rate <SortIcon column="rate" />
                  </th>
                  <th
                    className="px-3 py-2 text-right text-[10px] font-bold text-zinc-400 uppercase cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('commissionAmount')}
                  >
                    Commission <SortIcon column="commissionAmount" />
                  </th>
                  <th
                    className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('tradeId')}
                  >
                    Trade ID <SortIcon column="tradeId" />
                  </th>
                </tr>
              </thead>
              <tbody>
                {paginatedRecords.map((record) => (
                  <React.Fragment key={record.id}>
                    <tr
                      className="border-b border-zinc-800 hover:bg-[#25272E] cursor-pointer transition-colors"
                      onClick={() => setExpandedRow(expandedRow === record.id ? null : record.id)}
                    >
                      <td className="px-3 py-2 text-zinc-400">{formatDate(record.date)}</td>
                      <td className="px-3 py-2 text-zinc-300">{record.clientName}</td>
                      <td className="px-3 py-2 text-blue-400 font-mono">{record.symbol}</td>
                      <td className="px-3 py-2 text-right text-zinc-300 font-mono">{record.volume.toFixed(2)}</td>
                      <td className="px-3 py-2">{getModelBadge(record.commissionModel)}</td>
                      <td className="px-3 py-2 text-right text-zinc-400 font-mono">
                        {record.commissionModel === 'percentage' ? `${(record.rate * 100).toFixed(2)}%` : `$${record.rate.toFixed(2)}`}
                      </td>
                      <td className="px-3 py-2 text-right text-emerald-400 font-mono font-bold">{formatCurrency(record.commissionAmount)}</td>
                      <td className="px-3 py-2 text-zinc-500 font-mono text-[10px]">{record.tradeId}</td>
                    </tr>

                    {/* Expandable Row Detail */}
                    {expandedRow === record.id && (
                      <tr>
                        <td colSpan={8} className="px-3 py-3 bg-[#1a1a1c] border-b border-zinc-800">
                          <div className="grid grid-cols-4 gap-4 text-[11px]">
                            <div>
                              <div className="text-zinc-500 font-bold mb-1">Client Details</div>
                              <div className="space-y-0.5">
                                <div><span className="text-zinc-500">ID:</span> <span className="text-zinc-300">{record.clientId}</span></div>
                                <div><span className="text-zinc-500">Group:</span> <span className="text-zinc-300">{record.groupName}</span></div>
                              </div>
                            </div>
                            <div>
                              <div className="text-zinc-500 font-bold mb-1">Symbol Details</div>
                              <div className="space-y-0.5">
                                <div><span className="text-zinc-500">Symbol:</span> <span className="text-blue-400">{record.symbol}</span></div>
                                <div><span className="text-zinc-500">Group:</span> <span className="text-zinc-300 capitalize">{record.symbolGroup}</span></div>
                              </div>
                            </div>
                            <div>
                              <div className="text-zinc-500 font-bold mb-1">Commission Calculation</div>
                              <div className="space-y-0.5">
                                <div><span className="text-zinc-500">Model:</span> <span className={getModelColor(record.commissionModel)}>{record.commissionModel}</span></div>
                                <div><span className="text-zinc-500">Rate:</span> <span className="text-zinc-300">{record.commissionModel === 'percentage' ? `${(record.rate * 100).toFixed(2)}%` : `$${record.rate.toFixed(2)}/lot`}</span></div>
                                <div><span className="text-zinc-500">Volume:</span> <span className="text-zinc-300">{record.volume.toFixed(2)} lots</span></div>
                              </div>
                            </div>
                            <div>
                              <div className="text-zinc-500 font-bold mb-1">Trade Reference</div>
                              <div className="space-y-0.5">
                                <div><span className="text-zinc-500">Trade ID:</span> <span className="text-zinc-300 font-mono">{record.tradeId}</span></div>
                                <div><span className="text-zinc-500">Timestamp:</span> <span className="text-zinc-300">{new Date(record.date).toLocaleString()}</span></div>
                              </div>
                            </div>
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="bg-[#252525] border-t border-zinc-700 px-3 py-2 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="text-[10px] text-zinc-500">Rows per page:</span>
              <select
                value={pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value));
                  setCurrentPage(1);
                }}
                className="bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              >
                <option value={25}>25</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
              </select>
            </div>

            <div className="flex items-center gap-2">
              <span className="text-[10px] text-zinc-500">
                Page {currentPage} of {totalPages}
              </span>
              <button
                onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                disabled={currentPage === 1}
                className="px-2 py-1 bg-[#121316] border border-zinc-700 rounded text-xs text-zinc-300 disabled:opacity-30 disabled:cursor-not-allowed hover:bg-[#25272E]"
              >
                Previous
              </button>
              <button
                onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                disabled={currentPage === totalPages}
                className="px-2 py-1 bg-[#121316] border border-zinc-700 rounded text-xs text-zinc-300 disabled:opacity-30 disabled:cursor-not-allowed hover:bg-[#25272E]"
              >
                Next
              </button>
            </div>
          </div>
        </div>

        {/* Charts Section */}
        <div className="grid grid-cols-3 gap-3">
          {/* Daily Trend */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <DailyTrendChart />
          </div>

          {/* Top Clients */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <TopClientsChart />
          </div>

          {/* Symbol Groups */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <SymbolGroupChart />
          </div>
        </div>
      </div>
    </div>
  );
}
