'use client';

import React, { useState, useMemo } from 'react';
import {
    FileText,
    TrendingUp,
    Users,
    BookOpen,
    Download,
    Loader2,
    Calendar,
    Filter,
    ArrowUpDown,
    BarChart3,
    DollarSign,
    Activity,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';
import { api } from '@/services/apiClient';

type ReportType = 'trades' | 'positions' | 'accounts' | 'ledger';
type ExportFormat = 'csv' | 'json';
type SortDirection = 'asc' | 'desc';

interface ReportCard {
    id: ReportType;
    title: string;
    description: string;
    icon: React.ReactNode;
}

const REPORT_TYPES: ReportCard[] = [
    {
        id: 'trades',
        title: 'Trade History',
        description: 'Export completed trades with P&L',
        icon: <TrendingUp className="w-6 h-6" />,
    },
    {
        id: 'positions',
        title: 'Open Positions',
        description: 'Current open positions snapshot',
        icon: <Activity className="w-6 h-6" />,
    },
    {
        id: 'accounts',
        title: 'Account List',
        description: 'All trading accounts with balances',
        icon: <Users className="w-6 h-6" />,
    },
    {
        id: 'ledger',
        title: 'Ledger / Transactions',
        description: 'Deposits, withdrawals, and balance changes',
        icon: <BookOpen className="w-6 h-6" />,
    },
];

export default function ReportsView() {
    const [selectedReport, setSelectedReport] = useState<ReportType>('trades');
    const [format, setFormat] = useState<ExportFormat>('json');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [data, setData] = useState<any[]>([]);
    const [totalRecords, setTotalRecords] = useState(0);

    // Filters
    const [dateFrom, setDateFrom] = useState(() => {
        const date = new Date();
        date.setDate(date.getDate() - 30); // Default: last 30 days
        return date.toISOString().split('T')[0];
    });
    const [dateTo, setDateTo] = useState(() => new Date().toISOString().split('T')[0]);
    const [symbolFilter, setSymbolFilter] = useState('');
    const [accountFilter, setAccountFilter] = useState('');

    // Sorting
    const [sortColumn, setSortColumn] = useState<string | null>(null);
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

    /**
     * Generate report based on selected type and filters
     */
    const handleGenerateReport = async () => {
        setLoading(true);
        setError(null);
        setData([]);

        try {
            let endpoint = '';
            const params = new URLSearchParams();
            params.append('format', format);

            switch (selectedReport) {
                case 'trades':
                    endpoint = API_CONFIG.EXPORT_TRADES;
                    params.append('from', dateFrom);
                    params.append('to', dateTo);
                    if (symbolFilter) params.append('symbol', symbolFilter);
                    if (accountFilter) params.append('accountId', accountFilter);
                    break;
                case 'positions':
                    endpoint = API_CONFIG.EXPORT_POSITIONS;
                    break;
                case 'accounts':
                    endpoint = API_CONFIG.EXPORT_ACCOUNTS;
                    break;
                case 'ledger':
                    endpoint = API_CONFIG.EXPORT_LEDGER;
                    params.append('from', dateFrom);
                    params.append('to', dateTo);
                    if (accountFilter) params.append('accountId', accountFilter);
                    break;
            }

            const url = `${endpoint}?${params.toString()}`;

            if (format === 'csv') {
                // Direct download for CSV
                const response = await fetch(url, {
                    headers: {
                        Authorization: `Bearer ${localStorage.getItem('admin_token')}`,
                    },
                });

                if (!response.ok) {
                    throw new Error(`Export failed: ${response.statusText}`);
                }

                const blob = await response.blob();
                const downloadUrl = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = downloadUrl;
                a.download = `${selectedReport}_${new Date().toISOString().split('T')[0]}.csv`;
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                window.URL.revokeObjectURL(downloadUrl);

                // Show success message
                setTotalRecords(0);
                setData([]);
            } else {
                // Fetch JSON for preview
                const response = await api.get<any[]>(url);
                setData(response.slice(0, 50)); // First 50 rows
                setTotalRecords(response.length);
            }
        } catch (err: any) {
            setError(err.message || 'Failed to generate report');
        } finally {
            setLoading(false);
        }
    };

    /**
     * Download current data as CSV
     */
    const handleDownloadCSV = () => {
        if (data.length === 0) return;

        // Convert JSON to CSV
        const headers = Object.keys(data[0]);
        const csvRows = [
            headers.join(','),
            ...data.map(row =>
                headers.map(header => {
                    const value = row[header];
                    // Escape values with commas or quotes
                    if (typeof value === 'string' && (value.includes(',') || value.includes('"'))) {
                        return `"${value.replace(/"/g, '""')}"`;
                    }
                    return value;
                }).join(',')
            ),
        ];

        const csvContent = csvRows.join('\n');
        const blob = new Blob([csvContent], { type: 'text/csv' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${selectedReport}_${new Date().toISOString().split('T')[0]}.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
    };

    /**
     * Handle column sorting
     */
    const handleSort = (column: string) => {
        if (sortColumn === column) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortColumn(column);
            setSortDirection('asc');
        }
    };

    /**
     * Sorted data based on current sort state
     */
    const sortedData = useMemo(() => {
        if (!sortColumn || data.length === 0) return data;

        return [...data].sort((a, b) => {
            const aVal = a[sortColumn];
            const bVal = b[sortColumn];

            if (typeof aVal === 'number' && typeof bVal === 'number') {
                return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
            }

            const aStr = String(aVal || '');
            const bStr = String(bVal || '');
            return sortDirection === 'asc'
                ? aStr.localeCompare(bStr)
                : bStr.localeCompare(aStr);
        });
    }, [data, sortColumn, sortDirection]);

    /**
     * Calculate quick stats based on current data
     */
    const stats = useMemo(() => {
        if (data.length === 0) return null;

        switch (selectedReport) {
            case 'trades':
                const totalPnL = data.reduce((sum, trade) => sum + (trade.pnl || 0), 0);
                const winners = data.filter(t => (t.pnl || 0) > 0).length;
                const winRate = data.length > 0 ? (winners / data.length) * 100 : 0;
                return {
                    'Total Trades': data.length.toLocaleString(),
                    'Total P&L': `$${totalPnL.toFixed(2)}`,
                    'Win Rate': `${winRate.toFixed(1)}%`,
                };

            case 'positions':
                const totalExposure = data.reduce((sum, pos) => sum + Math.abs(pos.volume || 0), 0);
                const unrealizedPnL = data.reduce((sum, pos) => sum + (pos.unrealizedPnL || 0), 0);
                return {
                    'Total Positions': data.length.toLocaleString(),
                    'Net Exposure': totalExposure.toFixed(2),
                    'Unrealized P&L': `$${unrealizedPnL.toFixed(2)}`,
                };

            case 'accounts':
                const totalBalance = data.reduce((sum, acc) => sum + (acc.balance || 0), 0);
                const activeAccounts = data.filter(a => a.status === 'active').length;
                return {
                    'Total Accounts': data.length.toLocaleString(),
                    'Total Balance': `$${totalBalance.toLocaleString()}`,
                    'Active': `${activeAccounts} / ${data.length}`,
                };

            case 'ledger':
                const deposits = data.filter(l => l.type === 'deposit').reduce((sum, l) => sum + (l.amount || 0), 0);
                const withdrawals = data.filter(l => l.type === 'withdrawal').reduce((sum, l) => sum + Math.abs(l.amount || 0), 0);
                const netFlow = deposits - withdrawals;
                return {
                    'Total Deposits': `$${deposits.toLocaleString()}`,
                    'Total Withdrawals': `$${withdrawals.toLocaleString()}`,
                    'Net Flow': `$${netFlow.toLocaleString()}`,
                };

            default:
                return null;
        }
    }, [data, selectedReport]);

    /**
     * Check if filters should be shown
     */
    const showDateFilter = selectedReport === 'trades' || selectedReport === 'ledger';
    const showSymbolFilter = selectedReport === 'trades';
    const showAccountFilter = selectedReport === 'trades' || selectedReport === 'ledger';

    return (
        <div className="h-full flex flex-col bg-[#0A0A0B] p-4 overflow-hidden">
            {/* Report Type Selector */}
            <div className="mb-4">
                <h2 className="text-lg font-semibold text-white mb-3">Select Report Type</h2>
                <div className="grid grid-cols-4 gap-3">
                    {REPORT_TYPES.map((report) => (
                        <button
                            key={report.id}
                            onClick={() => setSelectedReport(report.id)}
                            className={`p-4 rounded-lg border-2 transition-all ${
                                selectedReport === report.id
                                    ? 'bg-blue-500/20 border-blue-500 ring-2 ring-blue-500/50'
                                    : 'bg-zinc-900 border-zinc-800 hover:border-zinc-700 hover:bg-zinc-800'
                            }`}
                        >
                            <div className="flex items-center gap-3 mb-2">
                                <div className={selectedReport === report.id ? 'text-blue-400' : 'text-zinc-400'}>
                                    {report.icon}
                                </div>
                                <h3 className="font-semibold text-white text-sm">{report.title}</h3>
                            </div>
                            <p className="text-xs text-zinc-400 text-left">{report.description}</p>
                        </button>
                    ))}
                </div>
            </div>

            {/* Filters Section */}
            <div className="bg-zinc-900 rounded-lg p-4 mb-4">
                <div className="flex items-center gap-2 mb-3">
                    <Filter className="w-4 h-4 text-zinc-400" />
                    <h3 className="font-semibold text-white text-sm">Filters</h3>
                </div>

                <div className="grid grid-cols-5 gap-3">
                    {/* Date Range */}
                    {showDateFilter && (
                        <>
                            <div>
                                <label className="text-xs text-zinc-400 mb-1 block">From Date</label>
                                <input
                                    type="date"
                                    value={dateFrom}
                                    onChange={(e) => setDateFrom(e.target.value)}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                            </div>
                            <div>
                                <label className="text-xs text-zinc-400 mb-1 block">To Date</label>
                                <input
                                    type="date"
                                    value={dateTo}
                                    onChange={(e) => setDateTo(e.target.value)}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                            </div>
                        </>
                    )}

                    {/* Symbol Filter */}
                    {showSymbolFilter && (
                        <div>
                            <label className="text-xs text-zinc-400 mb-1 block">Symbol</label>
                            <input
                                type="text"
                                value={symbolFilter}
                                onChange={(e) => setSymbolFilter(e.target.value.toUpperCase())}
                                placeholder="e.g., EURUSD"
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>
                    )}

                    {/* Account Filter */}
                    {showAccountFilter && (
                        <div>
                            <label className="text-xs text-zinc-400 mb-1 block">Account ID</label>
                            <input
                                type="text"
                                value={accountFilter}
                                onChange={(e) => setAccountFilter(e.target.value)}
                                placeholder="e.g., 123456"
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>
                    )}

                    {/* Format Selector */}
                    <div>
                        <label className="text-xs text-zinc-400 mb-1 block">Export Format</label>
                        <div className="flex gap-2">
                            <label className="flex items-center gap-2 px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm cursor-pointer hover:bg-zinc-750">
                                <input
                                    type="radio"
                                    value="json"
                                    checked={format === 'json'}
                                    onChange={() => setFormat('json')}
                                    className="accent-blue-500"
                                />
                                <span className="text-white">JSON</span>
                            </label>
                            <label className="flex items-center gap-2 px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm cursor-pointer hover:bg-zinc-750">
                                <input
                                    type="radio"
                                    value="csv"
                                    checked={format === 'csv'}
                                    onChange={() => setFormat('csv')}
                                    className="accent-blue-500"
                                />
                                <span className="text-white">CSV</span>
                            </label>
                        </div>
                    </div>
                </div>
            </div>

            {/* Action Bar */}
            <div className="flex items-center gap-3 mb-4">
                <button
                    onClick={handleGenerateReport}
                    disabled={loading}
                    className="px-6 py-2.5 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 font-semibold"
                >
                    {loading ? (
                        <>
                            <Loader2 className="w-4 h-4 animate-spin" />
                            Generating...
                        </>
                    ) : (
                        <>
                            <BarChart3 className="w-4 h-4" />
                            Generate Report
                        </>
                    )}
                </button>

                {data.length > 0 && format === 'json' && (
                    <button
                        onClick={handleDownloadCSV}
                        className="px-6 py-2.5 bg-emerald-500/20 border border-emerald-500/30 text-emerald-400 rounded-lg hover:bg-emerald-500/30 transition-colors flex items-center gap-2 font-semibold"
                    >
                        <Download className="w-4 h-4" />
                        Download CSV
                    </button>
                )}

                {/* Record Count */}
                {totalRecords > 0 && (
                    <div className="ml-auto text-sm text-zinc-400">
                        Showing <span className="text-white font-semibold">{data.length}</span> of{' '}
                        <span className="text-white font-semibold">{totalRecords.toLocaleString()}</span> records
                    </div>
                )}
            </div>

            {/* Error Message */}
            {error && (
                <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-300 text-sm">
                    {error}
                </div>
            )}

            {/* Main Content Area */}
            <div className="flex-1 flex gap-4 overflow-hidden">
                {/* Preview Table */}
                <div className="flex-1 bg-zinc-900 rounded-lg overflow-hidden flex flex-col">
                    <div className="p-3 border-b border-zinc-800 flex items-center justify-between">
                        <h3 className="font-semibold text-white">Preview</h3>
                        {format === 'csv' && (
                            <span className="text-xs text-zinc-500">
                                CSV format will download directly when generated
                            </span>
                        )}
                    </div>

                    <div className="flex-1 overflow-auto">
                        {data.length === 0 ? (
                            <div className="h-full flex items-center justify-center text-zinc-500">
                                <div className="text-center">
                                    <FileText className="w-12 h-12 mx-auto mb-3 opacity-30" />
                                    <p className="text-sm">No data to display</p>
                                    <p className="text-xs mt-1">Generate a report to preview data</p>
                                </div>
                            </div>
                        ) : (
                            <table className="w-full text-sm">
                                <thead className="bg-zinc-800 sticky top-0">
                                    <tr>
                                        {Object.keys(data[0]).map((key) => (
                                            <th
                                                key={key}
                                                onClick={() => handleSort(key)}
                                                className="px-4 py-2 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider cursor-pointer hover:bg-zinc-700 transition-colors"
                                            >
                                                <div className="flex items-center gap-2">
                                                    {key}
                                                    <ArrowUpDown className="w-3 h-3" />
                                                </div>
                                            </th>
                                        ))}
                                    </tr>
                                </thead>
                                <tbody>
                                    {sortedData.map((row, idx) => (
                                        <tr
                                            key={idx}
                                            className="border-b border-zinc-800 hover:bg-zinc-800/50 transition-colors"
                                        >
                                            {Object.values(row).map((value: any, cellIdx) => (
                                                <td key={cellIdx} className="px-4 py-2 text-zinc-300">
                                                    {typeof value === 'number'
                                                        ? value.toLocaleString()
                                                        : String(value || '-')}
                                                </td>
                                            ))}
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        )}
                    </div>
                </div>

                {/* Quick Stats Sidebar */}
                {stats && (
                    <div className="w-64 bg-zinc-900 rounded-lg p-4 space-y-3">
                        <h3 className="font-semibold text-white text-sm mb-3 flex items-center gap-2">
                            <DollarSign className="w-4 h-4 text-emerald-400" />
                            Quick Stats
                        </h3>
                        {Object.entries(stats).map(([label, value]) => (
                            <div key={label} className="p-3 bg-zinc-800 rounded-lg">
                                <div className="text-xs text-zinc-400 mb-1">{label}</div>
                                <div className="text-lg font-bold text-white">{value}</div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}
