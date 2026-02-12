'use client';

import { useState, useMemo } from 'react';
import { TrendingUp, TrendingDown, DollarSign, Download, BarChart3, PieChart, Activity } from 'lucide-react';

type TimeRange = 'today' | 'week' | 'mtd' | 'qtd' | 'ytd' | 'custom';

interface SymbolGroupPnL {
  group: string;
  volume: number;
  revenue: number;
  cost: number;
  netPnL: number;
  pnlPercent: number;
}

interface ProfitableSymbol {
  symbol: string;
  volume: number;
  revenue: number;
  netPnL: number;
  trades: number;
}

interface ProfitableClient {
  clientId: string;
  clientName: string;
  volume: number;
  pnlContribution: number;
  trades: number;
  winRate: number;
}

interface RevenueSource {
  spreads: number;
  commissions: number;
  swaps: number;
  markup: number;
  fees: number;
}

interface MonthlyRevenue {
  month: string;
  sources: RevenueSource;
  total: number;
}

interface DailyPnL {
  date: string;
  pnl: number;
}

const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(amount);
};

const formatNumber = (num: number): string => {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(num);
};

// Mock data generators
function generateMonthlyRevenue(): MonthlyRevenue[] {
  const months = ['Aug', 'Sep', 'Oct', 'Nov', 'Dec', 'Jan'];
  return months.map(month => {
    const spreads = Math.random() * 500000 + 300000;
    const commissions = Math.random() * 200000 + 100000;
    const swaps = Math.random() * 100000 + 50000;
    const markup = Math.random() * 150000 + 75000;
    const fees = Math.random() * 50000 + 25000;
    return {
      month,
      sources: { spreads, commissions, swaps, markup, fees },
      total: spreads + commissions + swaps + markup + fees
    };
  });
}

function generateSymbolGroupPnL(): SymbolGroupPnL[] {
  const groups = [
    { group: 'Forex Majors', baseVol: 5000000, baseRev: 250000 },
    { group: 'Forex Minors', baseVol: 2000000, baseRev: 120000 },
    { group: 'Forex Exotics', baseVol: 500000, baseRev: 80000 },
    { group: 'Metals', baseVol: 1500000, baseRev: 180000 },
    { group: 'Crypto', baseVol: 3000000, baseRev: 300000 },
    { group: 'Indices', baseVol: 2500000, baseRev: 200000 }
  ];

  return groups.map(g => {
    const volume = g.baseVol + Math.random() * g.baseVol * 0.5;
    const revenue = g.baseRev + Math.random() * g.baseRev * 0.3;
    const cost = revenue * (Math.random() * 0.3 + 0.1); // 10-40% cost
    const netPnL = revenue - cost;
    return {
      group: g.group,
      volume,
      revenue,
      cost,
      netPnL,
      pnlPercent: (netPnL / revenue) * 100
    };
  });
}

function generateProfitableSymbols(): ProfitableSymbol[] {
  const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'BTCUSD', 'XAUUSD', 'US30', 'ETHUSD', 'AUDUSD', 'USDCAD', 'NZDUSD'];
  return symbols.map(symbol => {
    const volume = Math.random() * 2000000 + 500000;
    const revenue = Math.random() * 150000 + 50000;
    const netPnL = revenue * (Math.random() * 0.5 + 0.5);
    const trades = Math.floor(Math.random() * 5000 + 1000);
    return { symbol, volume, revenue, netPnL, trades };
  }).sort((a, b) => b.netPnL - a.netPnL);
}

function generateProfitableClients(): ProfitableClient[] {
  const clients = Array.from({ length: 10 }, (_, i) => ({
    clientId: `CLI${String(10001 + i).padStart(6, '0')}`,
    clientName: `Client ${String.fromCharCode(65 + i)}`,
    volume: Math.random() * 1000000 + 200000,
    pnlContribution: Math.random() * 100000 + 20000,
    trades: Math.floor(Math.random() * 2000 + 500),
    winRate: Math.random() * 30 + 35
  }));
  return clients.sort((a, b) => b.pnlContribution - a.pnlContribution);
}

function generateDailyPnL(): DailyPnL[] {
  const data: DailyPnL[] = [];
  const now = new Date();
  for (let i = 29; i >= 0; i--) {
    const date = new Date(now);
    date.setDate(date.getDate() - i);
    const pnl = Math.random() * 100000 - 20000;
    data.push({
      date: date.toISOString().split('T')[0],
      pnl
    });
  }
  return data;
}

export default function BrokerPnlReport() {
  const [timeRange, setTimeRange] = useState<TimeRange>('mtd');
  const [customDateRange, setCustomDateRange] = useState({ start: '', end: '' });

  // Mock data
  const monthlyRevenue = useMemo(() => generateMonthlyRevenue(), []);
  const symbolGroupPnL = useMemo(() => generateSymbolGroupPnL(), []);
  const profitableSymbols = useMemo(() => generateProfitableSymbols(), []);
  const profitableClients = useMemo(() => generateProfitableClients(), []);
  const dailyPnL = useMemo(() => generateDailyPnL(), []);

  // Calculate summary stats
  const summaryStats = useMemo(() => {
    const totalRevenueMTD = 1234567.89;
    const totalRevenueYTD = 8765432.10;
    const netPnLMTD = 987654.32;
    const commissionRevenue = 345678.90;
    const spreadRevenue = 567890.12;
    const swapRevenue = 123456.78;
    const bBookPnL = 654321.00;
    const aBookPnL = 333333.32;

    return {
      totalRevenueMTD,
      totalRevenueYTD,
      netPnLMTD,
      commissionRevenue,
      spreadRevenue,
      swapRevenue,
      bBookPnL,
      aBookPnL
    };
  }, []);

  const handleExportCSV = () => {
    const headers = ['Symbol Group', 'Volume', 'Revenue', 'Cost', 'Net P&L', 'P&L %'];
    const rows = symbolGroupPnL.map(g => [
      g.group,
      formatNumber(g.volume),
      formatCurrency(g.revenue),
      formatCurrency(g.cost),
      formatCurrency(g.netPnL),
      g.pnlPercent.toFixed(2) + '%'
    ]);

    const csv = [headers, ...rows].map(row => row.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `broker-pnl-report-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
  };

  const handleExportPDF = () => {
    alert('PDF export would be implemented with a library like jsPDF or server-side rendering');
  };

  // SVG bar chart dimensions
  const chartWidth = 600;
  const chartHeight = 200;
  const barWidth = chartWidth / monthlyRevenue.length - 10;
  const maxRevenue = Math.max(...monthlyRevenue.map(m => m.total));

  // SVG line chart for daily P&L
  const lineChartWidth = 800;
  const lineChartHeight = 150;
  const maxPnL = Math.max(...dailyPnL.map(d => Math.abs(d.pnl)));
  const linePoints = dailyPnL.map((d, i) => {
    const x = (i / (dailyPnL.length - 1)) * lineChartWidth;
    const y = lineChartHeight / 2 - (d.pnl / maxPnL) * (lineChartHeight / 2 - 10);
    return `${x},${y}`;
  }).join(' ');

  return (
    <div className="h-full flex flex-col bg-[#121316] text-[#E8E9ED] overflow-hidden">
      {/* Header */}
      <div className="h-12 bg-[#1E2026] border-b border-[#383A42] px-4 flex items-center justify-between flex-shrink-0">
        <div className="flex items-center gap-3">
          <BarChart3 size={18} className="text-[#22C55E]" />
          <h2 className="text-sm font-bold text-[#E8E9ED]">Broker P&L Report</h2>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value as TimeRange)}
            className="px-3 py-1 text-xs bg-[#25272E] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
          >
            <option value="today">Today</option>
            <option value="week">This Week</option>
            <option value="mtd">MTD</option>
            <option value="qtd">QTD</option>
            <option value="ytd">YTD</option>
            <option value="custom">Custom</option>
          </select>
          {timeRange === 'custom' && (
            <>
              <input
                type="date"
                value={customDateRange.start}
                onChange={(e) => setCustomDateRange(prev => ({ ...prev, start: e.target.value }))}
                className="px-2 py-1 text-xs bg-[#25272E] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
              />
              <input
                type="date"
                value={customDateRange.end}
                onChange={(e) => setCustomDateRange(prev => ({ ...prev, end: e.target.value }))}
                className="px-2 py-1 text-xs bg-[#25272E] border border-[#383A42] rounded text-[#E8E9ED] outline-none"
              />
            </>
          )}
          <button
            onClick={handleExportCSV}
            className="px-3 py-1 text-xs font-medium bg-[#25272E] text-[#E8E9ED] rounded hover:bg-[#2D2F38] flex items-center gap-1.5"
          >
            <Download size={12} />
            CSV
          </button>
          <button
            onClick={handleExportPDF}
            className="px-3 py-1 text-xs font-medium bg-[#3B82F6] text-white rounded hover:bg-[#2563EB] flex items-center gap-1.5"
          >
            <Download size={12} />
            PDF
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-6 gap-3 p-4 flex-shrink-0">
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Total Revenue (MTD)</div>
          <div className="text-xl font-bold text-[#22C55E]">{formatCurrency(summaryStats.totalRevenueMTD)}</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Total Revenue (YTD)</div>
          <div className="text-xl font-bold text-[#3B82F6]">{formatCurrency(summaryStats.totalRevenueYTD)}</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Net P&L (MTD)</div>
          <div className="text-xl font-bold text-[#F5C542]">{formatCurrency(summaryStats.netPnLMTD)}</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Commission Revenue</div>
          <div className="text-xl font-bold text-[#E8E9ED]">{formatCurrency(summaryStats.commissionRevenue)}</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Spread Revenue</div>
          <div className="text-xl font-bold text-[#E8E9ED]">{formatCurrency(summaryStats.spreadRevenue)}</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Swap Revenue</div>
          <div className="text-xl font-bold text-[#E8E9ED]">{formatCurrency(summaryStats.swapRevenue)}</div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto px-4 pb-4">
        <div className="grid grid-cols-2 gap-4">
          {/* Revenue Breakdown Chart */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <h3 className="text-sm font-bold text-[#E8E9ED] mb-3">Revenue Breakdown (6 Months)</h3>
            <svg width={chartWidth} height={chartHeight + 30} className="w-full">
              {monthlyRevenue.map((month, i) => {
                const x = i * (chartWidth / monthlyRevenue.length) + 5;
                let yOffset = chartHeight;
                const colors = {
                  spreads: '#22C55E',
                  commissions: '#3B82F6',
                  swaps: '#F59E0B',
                  markup: '#A855F7',
                  fees: '#EC4899'
                };

                return (
                  <g key={i}>
                    {Object.entries(month.sources).map(([source, value]) => {
                      const barHeight = (value / maxRevenue) * chartHeight;
                      const rect = (
                        <rect
                          key={source}
                          x={x}
                          y={yOffset - barHeight}
                          width={barWidth}
                          height={barHeight}
                          fill={colors[source as keyof typeof colors]}
                          opacity={0.8}
                        />
                      );
                      yOffset -= barHeight;
                      return rect;
                    })}
                    <text x={x + barWidth / 2} y={chartHeight + 15} fontSize="10" fill="#888" textAnchor="middle">
                      {month.month}
                    </text>
                  </g>
                );
              })}
            </svg>
            <div className="flex flex-wrap gap-3 mt-3 text-xs">
              <div className="flex items-center gap-1"><div className="w-3 h-3 bg-[#22C55E] rounded" /><span>Spreads</span></div>
              <div className="flex items-center gap-1"><div className="w-3 h-3 bg-[#3B82F6] rounded" /><span>Commissions</span></div>
              <div className="flex items-center gap-1"><div className="w-3 h-3 bg-[#F59E0B] rounded" /><span>Swaps</span></div>
              <div className="flex items-center gap-1"><div className="w-3 h-3 bg-[#A855F7] rounded" /><span>Markup</span></div>
              <div className="flex items-center gap-1"><div className="w-3 h-3 bg-[#EC4899] rounded" /><span>Fees</span></div>
            </div>
          </div>

          {/* B-Book vs A-Book Comparison */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <h3 className="text-sm font-bold text-[#E8E9ED] mb-3">B-Book vs A-Book P&L</h3>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm text-[#888]">B-Book (Client Losses)</span>
                  <span className="text-lg font-bold text-[#22C55E]">{formatCurrency(summaryStats.bBookPnL)}</span>
                </div>
                <div className="h-3 bg-[#383A42] rounded-full overflow-hidden">
                  <div
                    className="h-full bg-[#22C55E] rounded-full"
                    style={{ width: `${(summaryStats.bBookPnL / (summaryStats.bBookPnL + summaryStats.aBookPnL)) * 100}%` }}
                  />
                </div>
                <div className="text-xs text-[#888] mt-1">
                  {((summaryStats.bBookPnL / (summaryStats.bBookPnL + summaryStats.aBookPnL)) * 100).toFixed(1)}% of total P&L
                </div>
              </div>
              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm text-[#888]">A-Book (Spread Markup)</span>
                  <span className="text-lg font-bold text-[#3B82F6]">{formatCurrency(summaryStats.aBookPnL)}</span>
                </div>
                <div className="h-3 bg-[#383A42] rounded-full overflow-hidden">
                  <div
                    className="h-full bg-[#3B82F6] rounded-full"
                    style={{ width: `${(summaryStats.aBookPnL / (summaryStats.bBookPnL + summaryStats.aBookPnL)) * 100}%` }}
                  />
                </div>
                <div className="text-xs text-[#888] mt-1">
                  {((summaryStats.aBookPnL / (summaryStats.bBookPnL + summaryStats.aBookPnL)) * 100).toFixed(1)}% of total P&L
                </div>
              </div>
              <div className="pt-3 border-t border-[#383A42]">
                <div className="flex justify-between items-center">
                  <span className="text-sm font-bold text-[#E8E9ED]">Total Net P&L</span>
                  <span className="text-xl font-bold text-[#F5C542]">{formatCurrency(summaryStats.bBookPnL + summaryStats.aBookPnL)}</span>
                </div>
              </div>
            </div>
          </div>

          {/* P&L by Symbol Group */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded overflow-hidden col-span-2">
            <div className="px-4 py-3 border-b border-[#383A42]">
              <h3 className="text-sm font-bold text-[#E8E9ED]">P&L by Symbol Group</h3>
            </div>
            <table className="w-full text-xs">
              <thead className="bg-[#25272E] text-[#888]">
                <tr>
                  <th className="px-4 py-2 text-left font-medium">Symbol Group</th>
                  <th className="px-4 py-2 text-right font-medium">Volume</th>
                  <th className="px-4 py-2 text-right font-medium">Revenue</th>
                  <th className="px-4 py-2 text-right font-medium">Cost (A-Book)</th>
                  <th className="px-4 py-2 text-right font-medium">Net P&L</th>
                  <th className="px-4 py-2 text-right font-medium">P&L %</th>
                </tr>
              </thead>
              <tbody>
                {symbolGroupPnL.map((group, idx) => (
                  <tr key={idx} className="border-b border-[#383A42] hover:bg-[#25272E]">
                    <td className="px-4 py-2 text-[#E8E9ED] font-bold">{group.group}</td>
                    <td className="px-4 py-2 text-right text-[#E8E9ED] font-mono">{formatNumber(group.volume)}</td>
                    <td className="px-4 py-2 text-right text-[#22C55E] font-mono">{formatCurrency(group.revenue)}</td>
                    <td className="px-4 py-2 text-right text-[#EF4444] font-mono">{formatCurrency(group.cost)}</td>
                    <td className="px-4 py-2 text-right font-bold font-mono" style={{ color: group.netPnL >= 0 ? '#22C55E' : '#EF4444' }}>
                      {formatCurrency(group.netPnL)}
                    </td>
                    <td className="px-4 py-2 text-right text-[#888] font-mono">{group.pnlPercent.toFixed(1)}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Top 10 Most Profitable Symbols */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded overflow-hidden">
            <div className="px-4 py-3 border-b border-[#383A42]">
              <h3 className="text-sm font-bold text-[#E8E9ED]">Top 10 Most Profitable Symbols</h3>
            </div>
            <table className="w-full text-xs">
              <thead className="bg-[#25272E] text-[#888]">
                <tr>
                  <th className="px-4 py-2 text-left font-medium">Symbol</th>
                  <th className="px-4 py-2 text-right font-medium">Volume</th>
                  <th className="px-4 py-2 text-right font-medium">Net P&L</th>
                  <th className="px-4 py-2 text-right font-medium">Trades</th>
                </tr>
              </thead>
              <tbody>
                {profitableSymbols.map((symbol, idx) => (
                  <tr key={idx} className="border-b border-[#383A42] hover:bg-[#25272E]">
                    <td className="px-4 py-2 text-[#E8E9ED] font-bold">{symbol.symbol}</td>
                    <td className="px-4 py-2 text-right text-[#888] font-mono">{formatNumber(symbol.volume)}</td>
                    <td className="px-4 py-2 text-right text-[#22C55E] font-mono font-bold">{formatCurrency(symbol.netPnL)}</td>
                    <td className="px-4 py-2 text-right text-[#888]">{symbol.trades.toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Top 10 Most Profitable Clients */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded overflow-hidden">
            <div className="px-4 py-3 border-b border-[#383A42]">
              <h3 className="text-sm font-bold text-[#E8E9ED]">Top 10 Most Profitable Clients</h3>
            </div>
            <table className="w-full text-xs">
              <thead className="bg-[#25272E] text-[#888]">
                <tr>
                  <th className="px-4 py-2 text-left font-medium">Client</th>
                  <th className="px-4 py-2 text-right font-medium">Volume</th>
                  <th className="px-4 py-2 text-right font-medium">P&L Contribution</th>
                  <th className="px-4 py-2 text-right font-medium">Win Rate</th>
                </tr>
              </thead>
              <tbody>
                {profitableClients.map((client, idx) => (
                  <tr key={idx} className="border-b border-[#383A42] hover:bg-[#25272E]">
                    <td className="px-4 py-2">
                      <div className="text-[#E8E9ED] font-bold">{client.clientName}</div>
                      <div className="text-[#888] text-xs font-mono">{client.clientId}</div>
                    </td>
                    <td className="px-4 py-2 text-right text-[#888] font-mono">{formatNumber(client.volume)}</td>
                    <td className="px-4 py-2 text-right text-[#22C55E] font-mono font-bold">{formatCurrency(client.pnlContribution)}</td>
                    <td className="px-4 py-2 text-right text-[#888]">{client.winRate.toFixed(1)}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Daily P&L Trend */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4 col-span-2">
            <h3 className="text-sm font-bold text-[#E8E9ED] mb-3">Daily P&L Trend (30 Days)</h3>
            <svg width={lineChartWidth} height={lineChartHeight + 30} className="w-full">
              {/* Zero line */}
              <line x1="0" y1={lineChartHeight / 2} x2={lineChartWidth} y2={lineChartHeight / 2} stroke="#383A42" strokeWidth="1" strokeDasharray="5,5" />

              {/* Area fill */}
              <path
                d={`M 0,${lineChartHeight / 2} L ${linePoints} L ${lineChartWidth},${lineChartHeight / 2} Z`}
                fill="#22C55E"
                opacity="0.1"
              />

              {/* Line */}
              <polyline
                points={linePoints}
                fill="none"
                stroke="#22C55E"
                strokeWidth="2"
              />

              {/* Points */}
              {dailyPnL.map((d, i) => {
                const x = (i / (dailyPnL.length - 1)) * lineChartWidth;
                const y = lineChartHeight / 2 - (d.pnl / maxPnL) * (lineChartHeight / 2 - 10);
                return (
                  <circle
                    key={i}
                    cx={x}
                    cy={y}
                    r="3"
                    fill={d.pnl >= 0 ? '#22C55E' : '#EF4444'}
                  />
                );
              })}

              {/* Date labels (show every 5th) */}
              {dailyPnL.filter((_, i) => i % 5 === 0).map((d, i) => {
                const x = ((i * 5) / (dailyPnL.length - 1)) * lineChartWidth;
                return (
                  <text key={i} x={x} y={lineChartHeight + 15} fontSize="10" fill="#888" textAnchor="middle">
                    {new Date(d.date).getDate()}
                  </text>
                );
              })}
            </svg>
          </div>
        </div>
      </div>
    </div>
  );
}
