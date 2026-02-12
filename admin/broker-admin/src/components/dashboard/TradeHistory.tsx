'use client';

import { useState, useMemo } from 'react';
import {
    FileText,
    Download,
    ChevronUp,
    ChevronDown,
    ChevronLeft,
    ChevronRight,
    Calendar,
    Filter,
    TrendingUp,
    TrendingDown,
} from 'lucide-react';

type TradeType = 'BUY' | 'SELL';
type ViewMode = 'trades' | 'statement';

interface Trade {
    ticket: string;
    openTime: Date;
    closeTime: Date;
    type: TradeType;
    symbol: string;
    volume: number;
    openPrice: number;
    closePrice: number;
    sl: number;
    tp: number;
    commission: number;
    swap: number;
    profit: number;
    accountId: string;
}

const generateMockTrades = (): Trade[] => {
    const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCHF', 'USDCAD', 'NZDUSD', 'XAUUSD', 'BTCUSD'];
    const accounts = ['100001', '100002', '100003', '100004', '100005'];
    const trades: Trade[] = [];
    const now = new Date();

    for (let i = 0; i < 60; i++) {
        const symbol = symbols[Math.floor(Math.random() * symbols.length)];
        const account = accounts[Math.floor(Math.random() * accounts.length)];
        const type: TradeType = Math.random() > 0.5 ? 'BUY' : 'SELL';
        const volume = parseFloat((Math.random() * 5 + 0.01).toFixed(2));
        const openPrice = parseFloat((Math.random() * 100 + 1.05).toFixed(5));
        const priceChange = parseFloat(((Math.random() - 0.5) * 0.01).toFixed(5));
        const closePrice = parseFloat((openPrice + priceChange).toFixed(5));
        const pip = symbol.includes('JPY') ? 0.01 : 0.0001;
        const pips = type === 'BUY'
            ? (closePrice - openPrice) / pip
            : (openPrice - closePrice) / pip;
        const profit = parseFloat((pips * 10 * volume - Math.random() * 2).toFixed(2));

        trades.push({
            ticket: `${1000000 + i}`,
            openTime: new Date(now.getTime() - Math.random() * 30 * 24 * 60 * 60 * 1000),
            closeTime: new Date(now.getTime() - Math.random() * 25 * 24 * 60 * 60 * 1000),
            type,
            symbol,
            volume,
            openPrice,
            closePrice,
            sl: type === 'BUY' ? parseFloat((openPrice - 0.01).toFixed(5)) : parseFloat((openPrice + 0.01).toFixed(5)),
            tp: type === 'BUY' ? parseFloat((openPrice + 0.02).toFixed(5)) : parseFloat((openPrice - 0.02).toFixed(5)),
            commission: parseFloat((Math.random() * -5).toFixed(2)),
            swap: parseFloat(((Math.random() - 0.5) * 3).toFixed(2)),
            profit,
            accountId: account,
        });
    }

    return trades.sort((a, b) => b.closeTime.getTime() - a.closeTime.getTime());
};

type SortField = 'ticket' | 'closeTime' | 'type' | 'symbol' | 'volume' | 'profit';
type SortDirection = 'asc' | 'desc';

export default function TradeHistory() {
    const [viewMode, setViewMode] = useState<ViewMode>('trades');
    const [trades] = useState<Trade[]>(generateMockTrades());
    const [dateFrom, setDateFrom] = useState('');
    const [dateTo, setDateTo] = useState('');
    const [selectedSymbol, setSelectedSymbol] = useState('all');
    const [selectedAccount, setSelectedAccount] = useState('all');
    const [sortField, setSortField] = useState<SortField>('closeTime');
    const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 50;

    const symbols = useMemo(() => {
        const uniqueSymbols = [...new Set(trades.map((t) => t.symbol))];
        return ['all', ...uniqueSymbols.sort()];
    }, [trades]);

    const accounts = useMemo(() => {
        const uniqueAccounts = [...new Set(trades.map((t) => t.accountId))];
        return ['all', ...uniqueAccounts.sort()];
    }, [trades]);

    const filteredTrades = useMemo(() => {
        let filtered = trades;

        if (dateFrom) {
            const fromDate = new Date(dateFrom);
            filtered = filtered.filter((t) => t.closeTime >= fromDate);
        }

        if (dateTo) {
            const toDate = new Date(dateTo);
            toDate.setHours(23, 59, 59, 999);
            filtered = filtered.filter((t) => t.closeTime <= toDate);
        }

        if (selectedSymbol !== 'all') {
            filtered = filtered.filter((t) => t.symbol === selectedSymbol);
        }

        if (selectedAccount !== 'all') {
            filtered = filtered.filter((t) => t.accountId === selectedAccount);
        }

        return filtered;
    }, [trades, dateFrom, dateTo, selectedSymbol, selectedAccount]);

    const sortedTrades = useMemo(() => {
        const sorted = [...filteredTrades];
        sorted.sort((a, b) => {
            let aVal: any = a[sortField];
            let bVal: any = b[sortField];

            if (sortField === 'closeTime') {
                aVal = a.closeTime.getTime();
                bVal = b.closeTime.getTime();
            }

            if (sortDirection === 'asc') {
                return aVal > bVal ? 1 : -1;
            } else {
                return aVal < bVal ? 1 : -1;
            }
        });
        return sorted;
    }, [filteredTrades, sortField, sortDirection]);

    const paginatedTrades = useMemo(() => {
        const start = (currentPage - 1) * itemsPerPage;
        const end = start + itemsPerPage;
        return sortedTrades.slice(start, end);
    }, [sortedTrades, currentPage]);

    const totalPages = Math.ceil(sortedTrades.length / itemsPerPage);

    const summary = useMemo(() => {
        return {
            totalTrades: filteredTrades.length,
            totalVolume: filteredTrades.reduce((sum, t) => sum + t.volume, 0),
            totalCommission: filteredTrades.reduce((sum, t) => sum + t.commission, 0),
            totalSwap: filteredTrades.reduce((sum, t) => sum + t.swap, 0),
            totalProfit: filteredTrades.reduce((sum, t) => sum + t.profit, 0),
            winningTrades: filteredTrades.filter((t) => t.profit > 0).length,
            losingTrades: filteredTrades.filter((t) => t.profit < 0).length,
        };
    }, [filteredTrades]);

    const handleSort = (field: SortField) => {
        if (sortField === field) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('desc');
        }
    };

    const exportCSV = () => {
        const headers = ['Ticket', 'Open Time', 'Close Time', 'Type', 'Symbol', 'Volume', 'Open Price', 'Close Price', 'SL', 'TP', 'Commission', 'Swap', 'Profit'];
        const rows = filteredTrades.map((t) => [
            t.ticket,
            t.openTime.toLocaleString(),
            t.closeTime.toLocaleString(),
            t.type,
            t.symbol,
            t.volume,
            t.openPrice,
            t.closePrice,
            t.sl,
            t.tp,
            t.commission,
            t.swap,
            t.profit,
        ]);

        const csv = [headers, ...rows].map((row) => row.join(',')).join('\n');
        const blob = new Blob([csv], { type: 'text/csv' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `trade-history-${new Date().toISOString().split('T')[0]}.csv`;
        a.click();
        URL.revokeObjectURL(url);
    };

    const exportPDF = () => {
        alert('PDF export functionality would be implemented here using a library like jsPDF');
    };

    const SortIcon = ({ field }: { field: SortField }) => {
        if (sortField !== field) return null;
        return sortDirection === 'asc' ? <ChevronUp size={14} /> : <ChevronDown size={14} />;
    };

    // Account Statement Chart Data
    const statementData = useMemo(() => {
        if (viewMode !== 'statement') return [];

        const sorted = [...filteredTrades].sort((a, b) => a.closeTime.getTime() - b.closeTime.getTime());
        let balance = 10000; // Starting balance
        const points: { time: Date; balance: number; equity: number }[] = [
            { time: sorted[0]?.closeTime || new Date(), balance: balance, equity: balance }
        ];

        sorted.forEach((trade) => {
            balance += trade.profit + trade.commission + trade.swap;
            points.push({
                time: trade.closeTime,
                balance: balance,
                equity: balance,
            });
        });

        return points;
    }, [filteredTrades, viewMode]);

    const renderChart = () => {
        if (statementData.length === 0) return null;

        const width = 800;
        const height = 300;
        const padding = { top: 20, right: 20, bottom: 40, left: 60 };
        const chartWidth = width - padding.left - padding.right;
        const chartHeight = height - padding.top - padding.bottom;

        const minBalance = Math.min(...statementData.map((d) => d.balance));
        const maxBalance = Math.max(...statementData.map((d) => d.balance));
        const balanceRange = maxBalance - minBalance;

        const points = statementData.map((d, i) => {
            const x = padding.left + (i / (statementData.length - 1)) * chartWidth;
            const y = padding.top + chartHeight - ((d.balance - minBalance) / balanceRange) * chartHeight;
            return `${x},${y}`;
        }).join(' ');

        return (
            <div className="bg-[#1E2026] border border-[#383A42] rounded p-6 mb-6">
                <h3 className="text-lg font-semibold text-white mb-4">Account Balance Over Time</h3>
                <svg width={width} height={height} className="w-full">
                    {/* Y-axis labels */}
                    {[0, 0.25, 0.5, 0.75, 1].map((ratio) => {
                        const value = minBalance + balanceRange * ratio;
                        const y = padding.top + chartHeight - ratio * chartHeight;
                        return (
                            <g key={ratio}>
                                <line
                                    x1={padding.left}
                                    y1={y}
                                    x2={width - padding.right}
                                    y2={y}
                                    stroke="#383A42"
                                    strokeWidth="1"
                                />
                                <text
                                    x={padding.left - 10}
                                    y={y}
                                    fill="#888"
                                    fontSize="10"
                                    textAnchor="end"
                                    dominantBaseline="middle"
                                >
                                    ${value.toFixed(0)}
                                </text>
                            </g>
                        );
                    })}

                    {/* Chart line */}
                    <polyline
                        points={points}
                        fill="none"
                        stroke="#2ECC71"
                        strokeWidth="2"
                    />

                    {/* X-axis */}
                    <line
                        x1={padding.left}
                        y1={height - padding.bottom}
                        x2={width - padding.right}
                        y2={height - padding.bottom}
                        stroke="#383A42"
                        strokeWidth="2"
                    />

                    {/* Y-axis */}
                    <line
                        x1={padding.left}
                        y1={padding.top}
                        x2={padding.left}
                        y2={height - padding.bottom}
                        stroke="#383A42"
                        strokeWidth="2"
                    />

                    {/* Labels */}
                    <text
                        x={width / 2}
                        y={height - 5}
                        fill="#888"
                        fontSize="12"
                        textAnchor="middle"
                    >
                        Time
                    </text>
                    <text
                        x={15}
                        y={height / 2}
                        fill="#888"
                        fontSize="12"
                        textAnchor="middle"
                        transform={`rotate(-90, 15, ${height / 2})`}
                    >
                        Balance (USD)
                    </text>
                </svg>
                <div className="mt-4 grid grid-cols-3 gap-4">
                    <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                        <div className="text-xs text-[#888]">Starting Balance</div>
                        <div className="text-lg font-semibold text-white">
                            ${statementData[0]?.balance.toFixed(2)}
                        </div>
                    </div>
                    <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                        <div className="text-xs text-[#888]">Current Balance</div>
                        <div className="text-lg font-semibold text-white">
                            ${statementData[statementData.length - 1]?.balance.toFixed(2)}
                        </div>
                    </div>
                    <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                        <div className="text-xs text-[#888]">Total Change</div>
                        <div className={`text-lg font-semibold ${summary.totalProfit >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                            {summary.totalProfit >= 0 ? '+' : ''}${summary.totalProfit.toFixed(2)}
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    return (
        <div className="h-full flex flex-col bg-[#121316] overflow-hidden">
            {/* Header */}
            <div className="px-6 py-4 border-b border-[#383A42] flex items-center justify-between flex-shrink-0">
                <div className="flex items-center gap-2">
                    <FileText size={20} className="text-[#F5C542]" />
                    <h1 className="text-xl font-bold text-white">Trade History & Statement</h1>
                </div>
                <div className="flex items-center gap-2">
                    <button
                        onClick={exportCSV}
                        className="px-3 py-1.5 bg-[#2ECC71] hover:bg-[#27AE60] text-white text-sm rounded flex items-center gap-2"
                    >
                        <Download size={14} />
                        Export CSV
                    </button>
                    <button
                        onClick={exportPDF}
                        className="px-3 py-1.5 bg-[#E74C3C] hover:bg-[#C0392B] text-white text-sm rounded flex items-center gap-2"
                    >
                        <Download size={14} />
                        Export PDF
                    </button>
                </div>
            </div>

            {/* View Mode Toggle */}
            <div className="px-6 py-3 border-b border-[#383A42] flex items-center gap-2 flex-shrink-0">
                <button
                    onClick={() => setViewMode('trades')}
                    className={`px-4 py-2 text-sm font-semibold rounded ${
                        viewMode === 'trades'
                            ? 'bg-[#2980B9] text-white'
                            : 'bg-[#25272E] text-[#888] hover:text-white'
                    }`}
                >
                    Trade History
                </button>
                <button
                    onClick={() => setViewMode('statement')}
                    className={`px-4 py-2 text-sm font-semibold rounded ${
                        viewMode === 'statement'
                            ? 'bg-[#2980B9] text-white'
                            : 'bg-[#25272E] text-[#888] hover:text-white'
                    }`}
                >
                    Account Statement
                </button>
            </div>

            {/* Filters */}
            <div className="px-6 py-4 border-b border-[#383A42] bg-[#1E2026] flex items-center gap-4 flex-shrink-0">
                <div className="flex items-center gap-2">
                    <Calendar size={14} className="text-[#888]" />
                    <input
                        type="date"
                        value={dateFrom}
                        onChange={(e) => {
                            setDateFrom(e.target.value);
                            setCurrentPage(1);
                        }}
                        className="px-3 py-1.5 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                    />
                    <span className="text-[#888] text-sm">to</span>
                    <input
                        type="date"
                        value={dateTo}
                        onChange={(e) => {
                            setDateTo(e.target.value);
                            setCurrentPage(1);
                        }}
                        className="px-3 py-1.5 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                    />
                </div>
                <div className="flex items-center gap-2">
                    <Filter size={14} className="text-[#888]" />
                    <select
                        value={selectedSymbol}
                        onChange={(e) => {
                            setSelectedSymbol(e.target.value);
                            setCurrentPage(1);
                        }}
                        className="px-3 py-1.5 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                    >
                        {symbols.map((symbol) => (
                            <option key={symbol} value={symbol}>
                                {symbol === 'all' ? 'All Symbols' : symbol}
                            </option>
                        ))}
                    </select>
                    <select
                        value={selectedAccount}
                        onChange={(e) => {
                            setSelectedAccount(e.target.value);
                            setCurrentPage(1);
                        }}
                        className="px-3 py-1.5 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                    >
                        {accounts.map((account) => (
                            <option key={account} value={account}>
                                {account === 'all' ? 'All Accounts' : `Account ${account}`}
                            </option>
                        ))}
                    </select>
                </div>
                <div className="flex-1"></div>
                <div className="text-sm text-[#888]">
                    {filteredTrades.length} trades found
                </div>
            </div>

            {/* Main Content */}
            <div className="flex-1 overflow-y-auto custom-scrollbar">
                <div className="p-6">
                    {/* Account Statement Chart */}
                    {viewMode === 'statement' && renderChart()}

                    {/* Trade Table */}
                    <div className="bg-[#1E2026] border border-[#383A42] rounded overflow-hidden">
                        <div className="overflow-x-auto">
                            <table className="w-full text-sm">
                                <thead className="bg-[#25272E] border-b border-[#383A42]">
                                    <tr>
                                        <th
                                            onClick={() => handleSort('ticket')}
                                            className="px-4 py-3 text-left text-xs font-semibold text-[#CCC] cursor-pointer hover:text-white"
                                        >
                                            <div className="flex items-center gap-1">
                                                Ticket
                                                <SortIcon field="ticket" />
                                            </div>
                                        </th>
                                        <th
                                            onClick={() => handleSort('closeTime')}
                                            className="px-4 py-3 text-left text-xs font-semibold text-[#CCC] cursor-pointer hover:text-white"
                                        >
                                            <div className="flex items-center gap-1">
                                                Close Time
                                                <SortIcon field="closeTime" />
                                            </div>
                                        </th>
                                        <th
                                            onClick={() => handleSort('type')}
                                            className="px-4 py-3 text-left text-xs font-semibold text-[#CCC] cursor-pointer hover:text-white"
                                        >
                                            <div className="flex items-center gap-1">
                                                Type
                                                <SortIcon field="type" />
                                            </div>
                                        </th>
                                        <th
                                            onClick={() => handleSort('symbol')}
                                            className="px-4 py-3 text-left text-xs font-semibold text-[#CCC] cursor-pointer hover:text-white"
                                        >
                                            <div className="flex items-center gap-1">
                                                Symbol
                                                <SortIcon field="symbol" />
                                            </div>
                                        </th>
                                        <th
                                            onClick={() => handleSort('volume')}
                                            className="px-4 py-3 text-right text-xs font-semibold text-[#CCC] cursor-pointer hover:text-white"
                                        >
                                            <div className="flex items-center justify-end gap-1">
                                                Volume
                                                <SortIcon field="volume" />
                                            </div>
                                        </th>
                                        <th className="px-4 py-3 text-right text-xs font-semibold text-[#CCC]">
                                            Open Price
                                        </th>
                                        <th className="px-4 py-3 text-right text-xs font-semibold text-[#CCC]">
                                            Close Price
                                        </th>
                                        <th className="px-4 py-3 text-right text-xs font-semibold text-[#CCC]">
                                            SL
                                        </th>
                                        <th className="px-4 py-3 text-right text-xs font-semibold text-[#CCC]">
                                            TP
                                        </th>
                                        <th className="px-4 py-3 text-right text-xs font-semibold text-[#CCC]">
                                            Commission
                                        </th>
                                        <th className="px-4 py-3 text-right text-xs font-semibold text-[#CCC]">
                                            Swap
                                        </th>
                                        <th
                                            onClick={() => handleSort('profit')}
                                            className="px-4 py-3 text-right text-xs font-semibold text-[#CCC] cursor-pointer hover:text-white"
                                        >
                                            <div className="flex items-center justify-end gap-1">
                                                Profit
                                                <SortIcon field="profit" />
                                            </div>
                                        </th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {paginatedTrades.map((trade, index) => (
                                        <tr
                                            key={trade.ticket}
                                            className={`border-b border-[#383A42] hover:bg-[#25272E] ${
                                                index % 2 === 0 ? 'bg-[#1E2026]' : 'bg-[#1a1c22]'
                                            }`}
                                        >
                                            <td className="px-4 py-3 text-white font-mono">{trade.ticket}</td>
                                            <td className="px-4 py-3 text-[#CCC]">
                                                {trade.closeTime.toLocaleString()}
                                            </td>
                                            <td className="px-4 py-3">
                                                <span
                                                    className={`px-2 py-0.5 text-xs rounded ${
                                                        trade.type === 'BUY'
                                                            ? 'bg-blue-500/10 text-blue-400'
                                                            : 'bg-red-500/10 text-red-400'
                                                    }`}
                                                >
                                                    {trade.type}
                                                </span>
                                            </td>
                                            <td className="px-4 py-3 text-white font-semibold">{trade.symbol}</td>
                                            <td className="px-4 py-3 text-right text-white font-mono">
                                                {trade.volume.toFixed(2)}
                                            </td>
                                            <td className="px-4 py-3 text-right text-[#CCC] font-mono">
                                                {trade.openPrice.toFixed(5)}
                                            </td>
                                            <td className="px-4 py-3 text-right text-[#CCC] font-mono">
                                                {trade.closePrice.toFixed(5)}
                                            </td>
                                            <td className="px-4 py-3 text-right text-[#888] font-mono text-xs">
                                                {trade.sl.toFixed(5)}
                                            </td>
                                            <td className="px-4 py-3 text-right text-[#888] font-mono text-xs">
                                                {trade.tp.toFixed(5)}
                                            </td>
                                            <td className="px-4 py-3 text-right text-red-400 font-mono">
                                                ${trade.commission.toFixed(2)}
                                            </td>
                                            <td className={`px-4 py-3 text-right font-mono ${
                                                trade.swap >= 0 ? 'text-green-400' : 'text-red-400'
                                            }`}>
                                                ${trade.swap.toFixed(2)}
                                            </td>
                                            <td className="px-4 py-3 text-right font-semibold font-mono">
                                                <div className="flex items-center justify-end gap-1">
                                                    {trade.profit >= 0 ? (
                                                        <TrendingUp size={14} className="text-green-400" />
                                                    ) : (
                                                        <TrendingDown size={14} className="text-red-400" />
                                                    )}
                                                    <span className={trade.profit >= 0 ? 'text-green-400' : 'text-red-400'}>
                                                        ${trade.profit.toFixed(2)}
                                                    </span>
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                                {/* Summary Row */}
                                <tfoot className="bg-[#25272E] border-t-2 border-[#F5C542]">
                                    <tr>
                                        <td colSpan={4} className="px-4 py-3 text-white font-bold">
                                            Summary ({summary.totalTrades} trades)
                                        </td>
                                        <td className="px-4 py-3 text-right text-white font-bold font-mono">
                                            {summary.totalVolume.toFixed(2)}
                                        </td>
                                        <td colSpan={4} className="px-4 py-3"></td>
                                        <td className="px-4 py-3 text-right text-red-400 font-bold font-mono">
                                            ${summary.totalCommission.toFixed(2)}
                                        </td>
                                        <td className={`px-4 py-3 text-right font-bold font-mono ${
                                            summary.totalSwap >= 0 ? 'text-green-400' : 'text-red-400'
                                        }`}>
                                            ${summary.totalSwap.toFixed(2)}
                                        </td>
                                        <td className={`px-4 py-3 text-right font-bold font-mono ${
                                            summary.totalProfit >= 0 ? 'text-green-400' : 'text-red-400'
                                        }`}>
                                            ${summary.totalProfit.toFixed(2)}
                                        </td>
                                    </tr>
                                </tfoot>
                            </table>
                        </div>
                    </div>

                    {/* Pagination */}
                    {totalPages > 1 && (
                        <div className="mt-6 flex items-center justify-between">
                            <div className="text-sm text-[#888]">
                                Page {currentPage} of {totalPages} • Showing {((currentPage - 1) * itemsPerPage) + 1}-{Math.min(currentPage * itemsPerPage, sortedTrades.length)} of {sortedTrades.length} trades
                            </div>
                            <div className="flex items-center gap-2">
                                <button
                                    onClick={() => setCurrentPage(1)}
                                    disabled={currentPage === 1}
                                    className="px-3 py-1.5 bg-[#25272E] border border-[#383A42] text-white text-sm rounded hover:bg-[#383A42] disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    First
                                </button>
                                <button
                                    onClick={() => setCurrentPage(currentPage - 1)}
                                    disabled={currentPage === 1}
                                    className="px-3 py-1.5 bg-[#25272E] border border-[#383A42] text-white text-sm rounded hover:bg-[#383A42] disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                                >
                                    <ChevronLeft size={14} />
                                    Previous
                                </button>
                                <button
                                    onClick={() => setCurrentPage(currentPage + 1)}
                                    disabled={currentPage === totalPages}
                                    className="px-3 py-1.5 bg-[#25272E] border border-[#383A42] text-white text-sm rounded hover:bg-[#383A42] disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                                >
                                    Next
                                    <ChevronRight size={14} />
                                </button>
                                <button
                                    onClick={() => setCurrentPage(totalPages)}
                                    disabled={currentPage === totalPages}
                                    className="px-3 py-1.5 bg-[#25272E] border border-[#383A42] text-white text-sm rounded hover:bg-[#383A42] disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    Last
                                </button>
                            </div>
                        </div>
                    )}

                    {/* Statistics */}
                    <div className="mt-6 grid grid-cols-4 gap-4">
                        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                            <div className="text-xs text-[#888]">Win Rate</div>
                            <div className="text-2xl font-bold text-white mt-1">
                                {summary.totalTrades > 0
                                    ? ((summary.winningTrades / summary.totalTrades) * 100).toFixed(1)
                                    : '0.0'}%
                            </div>
                            <div className="text-xs text-[#666] mt-1">
                                {summary.winningTrades}W / {summary.losingTrades}L
                            </div>
                        </div>
                        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                            <div className="text-xs text-[#888]">Total Volume</div>
                            <div className="text-2xl font-bold text-white mt-1">
                                {summary.totalVolume.toFixed(2)}
                            </div>
                            <div className="text-xs text-[#666] mt-1">lots</div>
                        </div>
                        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                            <div className="text-xs text-[#888]">Total Fees</div>
                            <div className="text-2xl font-bold text-red-400 mt-1">
                                ${(summary.totalCommission + summary.totalSwap).toFixed(2)}
                            </div>
                            <div className="text-xs text-[#666] mt-1">commission + swap</div>
                        </div>
                        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                            <div className="text-xs text-[#888]">Net Profit</div>
                            <div className={`text-2xl font-bold mt-1 ${
                                summary.totalProfit >= 0 ? 'text-green-400' : 'text-red-400'
                            }`}>
                                {summary.totalProfit >= 0 ? '+' : ''}${summary.totalProfit.toFixed(2)}
                            </div>
                            <div className="text-xs text-[#666] mt-1">
                                {summary.totalProfit >= 0 ? 'profit' : 'loss'}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
