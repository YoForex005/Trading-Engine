'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
    Users,
    Search,
    X,
    ChevronDown,
    ChevronRight,
    UserCheck,
    UserX,
    Bell,
    Download,
    Activity,
    TrendingUp,
    TrendingDown,
    DollarSign,
    Calendar,
    Clock,
} from 'lucide-react';
import { api } from '@/services/apiClient';
import { API_CONFIG } from '@/config/api';

type KYCStatus = 'Verified' | 'Pending' | 'Rejected';
type AccountStatus = 'Active' | 'Suspended' | 'Blocked';
type SortField = 'clientId' | 'name' | 'balance' | 'equity' | 'registrationDate' | 'lastLogin';
type SortDirection = 'asc' | 'desc';

interface Trade {
    id: string;
    symbol: string;
    type: 'BUY' | 'SELL';
    volume: number;
    pnl: number;
    closedAt: string;
}

interface Client {
    id: string;
    clientId: string;
    name: string;
    email: string;
    balance: number;
    equity: number;
    openPositions: number;
    kycStatus: KYCStatus;
    accountStatus: AccountStatus;
    registrationDate: string;
    lastLogin: string;
    // Expandable detail fields
    tradingStats: {
        totalTrades: number;
        winRate: number;
        profitFactor: number;
        avgTradeSize: number;
    };
    recentTrades: Trade[];
    financialSummary: {
        totalDeposits: number;
        totalWithdrawals: number;
        netDeposits: number;
    };
    notes: string;
}

interface SummaryStats {
    totalClients: number;
    activeTraders: number;
    pendingKYC: number;
    totalEquity: number;
}

const KYC_STATUSES: KYCStatus[] = ['Verified', 'Pending', 'Rejected'];
const ACCOUNT_STATUSES: AccountStatus[] = ['Active', 'Suspended', 'Blocked'];

const KYC_STATUS_COLORS: Record<KYCStatus, string> = {
    Verified: 'bg-green-600 text-white',
    Pending: 'bg-yellow-600 text-black',
    Rejected: 'bg-red-600 text-white',
};

const ACCOUNT_STATUS_COLORS: Record<AccountStatus, string> = {
    Active: 'text-green-400',
    Suspended: 'text-yellow-400',
    Blocked: 'text-red-400',
};

// Mock data generator
const generateMockClients = (): Client[] => {
    const firstNames = ['John', 'Emma', 'Michael', 'Sophia', 'William', 'Olivia', 'James', 'Ava', 'Robert', 'Isabella', 'David', 'Mia', 'Richard', 'Charlotte', 'Joseph'];
    const lastNames = ['Smith', 'Johnson', 'Williams', 'Brown', 'Jones', 'Garcia', 'Miller', 'Davis', 'Rodriguez', 'Martinez', 'Hernandez', 'Lopez', 'Gonzalez', 'Wilson', 'Anderson'];
    const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'BTCUSD', 'ETHUSD', 'XAUUSD'];

    return Array.from({ length: 25 }, (_, i) => {
        const firstName = firstNames[Math.floor(Math.random() * firstNames.length)];
        const lastName = lastNames[Math.floor(Math.random() * lastNames.length)];
        const name = `${firstName} ${lastName}`;
        const clientId = `CL${String(10001 + i).padStart(6, '0')}`;

        const balance = Math.floor(Math.random() * 50000) + 1000;
        const equity = balance + (Math.random() * 2000 - 1000);
        const openPositions = Math.floor(Math.random() * 8);

        const kycStatuses: KYCStatus[] = ['Verified', 'Verified', 'Verified', 'Pending', 'Rejected'];
        const kycStatus = kycStatuses[Math.floor(Math.random() * kycStatuses.length)];

        const accountStatuses: AccountStatus[] = ['Active', 'Active', 'Active', 'Active', 'Suspended', 'Blocked'];
        const accountStatus = accountStatuses[Math.floor(Math.random() * accountStatuses.length)];

        const daysAgo = Math.floor(Math.random() * 365);
        const registrationDate = new Date(Date.now() - daysAgo * 24 * 60 * 60 * 1000).toISOString();
        const lastLoginDaysAgo = Math.floor(Math.random() * 30);
        const lastLogin = new Date(Date.now() - lastLoginDaysAgo * 24 * 60 * 60 * 1000).toISOString();

        const totalTrades = Math.floor(Math.random() * 200) + 10;
        const wins = Math.floor(totalTrades * (0.4 + Math.random() * 0.3));
        const winRate = (wins / totalTrades) * 100;

        const recentTrades: Trade[] = Array.from({ length: 5 }, (_, j) => ({
            id: `T${i}-${j}`,
            symbol: symbols[Math.floor(Math.random() * symbols.length)],
            type: Math.random() > 0.5 ? 'BUY' : 'SELL',
            volume: parseFloat((Math.random() * 2).toFixed(2)),
            pnl: parseFloat((Math.random() * 1000 - 500).toFixed(2)),
            closedAt: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString(),
        }));

        const totalDeposits = Math.floor(Math.random() * 100000) + 10000;
        const totalWithdrawals = Math.floor(Math.random() * 50000);

        return {
            id: `client-${i}`,
            clientId,
            name,
            email: `${firstName.toLowerCase()}.${lastName.toLowerCase()}@example.com`,
            balance,
            equity,
            openPositions,
            kycStatus,
            accountStatus,
            registrationDate,
            lastLogin,
            tradingStats: {
                totalTrades,
                winRate,
                profitFactor: parseFloat((1 + Math.random() * 2).toFixed(2)),
                avgTradeSize: parseFloat((Math.random() * 5).toFixed(2)),
            },
            recentTrades,
            financialSummary: {
                totalDeposits,
                totalWithdrawals,
                netDeposits: totalDeposits - totalWithdrawals,
            },
            notes: Math.random() > 0.7 ? 'High-value client - priority support' : '',
        };
    });
};

export default function ClientManagement() {
    const [clients, setClients] = useState<Client[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [kycFilter, setKycFilter] = useState<'All' | KYCStatus>('All');
    const [accountFilter, setAccountFilter] = useState<'All' | AccountStatus>('All');
    const [expandedRow, setExpandedRow] = useState<string | null>(null);
    const [sortField, setSortField] = useState<SortField>('registrationDate');
    const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
    const [currentPage, setCurrentPage] = useState(1);
    const [showNotificationModal, setShowNotificationModal] = useState(false);
    const [notificationClient, setNotificationClient] = useState<Client | null>(null);
    const [notificationSubject, setNotificationSubject] = useState('');
    const [notificationMessage, setNotificationMessage] = useState('');

    const pageSize = 25;

    // Fetch clients on mount
    useEffect(() => {
        fetchClients();
    }, []);

    const fetchClients = async () => {
        try {
            setLoading(true);
            // In production, replace with actual API call:
            // const data = await api.get<Client[]>(API_CONFIG.CLIENT_LIST);
            // For now, use mock data
            await new Promise(resolve => setTimeout(resolve, 500));
            const mockData = generateMockClients();
            setClients(mockData);
        } catch (error) {
            console.error('[ClientManagement] Failed to fetch clients:', error);
        } finally {
            setLoading(false);
        }
    };

    // Calculate summary stats
    const summaryStats = useMemo((): SummaryStats => {
        const totalClients = clients.length;
        const sevenDaysAgo = Date.now() - 7 * 24 * 60 * 60 * 1000;
        const activeTraders = clients.filter(c => new Date(c.lastLogin).getTime() > sevenDaysAgo).length;
        const pendingKYC = clients.filter(c => c.kycStatus === 'Pending').length;
        const totalEquity = clients.reduce((sum, c) => sum + c.equity, 0);

        return {
            totalClients,
            activeTraders,
            pendingKYC,
            totalEquity,
        };
    }, [clients]);

    // Filter and sort clients
    const filteredAndSortedClients = useMemo(() => {
        let filtered = clients;

        // Search filter
        if (searchQuery.trim()) {
            const query = searchQuery.toLowerCase();
            filtered = filtered.filter(
                c =>
                    c.name.toLowerCase().includes(query) ||
                    c.email.toLowerCase().includes(query) ||
                    c.clientId.toLowerCase().includes(query)
            );
        }

        // KYC filter
        if (kycFilter !== 'All') {
            filtered = filtered.filter(c => c.kycStatus === kycFilter);
        }

        // Account status filter
        if (accountFilter !== 'All') {
            filtered = filtered.filter(c => c.accountStatus === accountFilter);
        }

        // Sort
        filtered.sort((a, b) => {
            let aVal: any = a[sortField];
            let bVal: any = b[sortField];

            if (sortField === 'registrationDate' || sortField === 'lastLogin') {
                aVal = new Date(aVal).getTime();
                bVal = new Date(bVal).getTime();
            }

            if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
            if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
            return 0;
        });

        return filtered;
    }, [clients, searchQuery, kycFilter, accountFilter, sortField, sortDirection]);

    // Paginate
    const totalPages = Math.ceil(filteredAndSortedClients.length / pageSize);
    const paginatedClients = useMemo(() => {
        const start = (currentPage - 1) * pageSize;
        return filteredAndSortedClients.slice(start, start + pageSize);
    }, [filteredAndSortedClients, currentPage]);

    const handleSort = (field: SortField) => {
        if (sortField === field) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('desc');
        }
    };

    const handleToggleExpand = (clientId: string) => {
        setExpandedRow(expandedRow === clientId ? null : clientId);
    };

    const handleSuspend = async (client: Client) => {
        if (!confirm(`Are you sure you want to ${client.accountStatus === 'Active' ? 'suspend' : 'activate'} ${client.name}?`)) {
            return;
        }

        try {
            // In production:
            // await api.put(API_CONFIG.CLIENT_SUSPEND(client.id), {});

            // Mock update
            setClients(prev =>
                prev.map(c =>
                    c.id === client.id
                        ? { ...c, accountStatus: c.accountStatus === 'Active' ? 'Suspended' : 'Active' }
                        : c
                )
            );
        } catch (error) {
            console.error('[ClientManagement] Failed to toggle client status:', error);
        }
    };

    const handleSendNotification = (client: Client) => {
        setNotificationClient(client);
        setNotificationSubject('');
        setNotificationMessage('');
        setShowNotificationModal(true);
    };

    const submitNotification = async () => {
        if (!notificationClient || !notificationSubject.trim() || !notificationMessage.trim()) {
            alert('Please fill in all fields');
            return;
        }

        try {
            // In production:
            // await api.post(API_CONFIG.CLIENT_SEND_NOTIFICATION(notificationClient.id), {
            //     subject: notificationSubject,
            //     message: notificationMessage,
            // });

            alert(`Notification sent to ${notificationClient.name}`);
            setShowNotificationModal(false);
        } catch (error) {
            console.error('[ClientManagement] Failed to send notification:', error);
        }
    };

    const handleExportCSV = () => {
        const headers = ['Client ID', 'Name', 'Email', 'Balance', 'Equity', 'Open Positions', 'KYC Status', 'Account Status', 'Registration Date', 'Last Login'];
        const rows = filteredAndSortedClients.map(c => [
            c.clientId,
            c.name,
            c.email,
            c.balance.toFixed(2),
            c.equity.toFixed(2),
            c.openPositions,
            c.kycStatus,
            c.accountStatus,
            new Date(c.registrationDate).toLocaleDateString(),
            new Date(c.lastLogin).toLocaleDateString(),
        ]);

        const csv = [headers, ...rows].map(row => row.join(',')).join('\n');
        const blob = new Blob([csv], { type: 'text/csv' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `clients_${new Date().toISOString().split('T')[0]}.csv`;
        a.click();
        URL.revokeObjectURL(url);
    };

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value);
    };

    const formatRelativeTime = (dateString: string) => {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

        if (diffDays === 0) return 'Today';
        if (diffDays === 1) return 'Yesterday';
        if (diffDays < 7) return `${diffDays} days ago`;
        if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
        if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
        return `${Math.floor(diffDays / 365)} years ago`;
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full bg-[#121316]">
                <div className="text-zinc-400">Loading clients...</div>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full bg-[#121316] text-white overflow-hidden">
            {/* Header */}
            <div className="flex-shrink-0 bg-zinc-900 border-b border-zinc-800 p-4">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <Users className="text-blue-400" size={24} />
                        <div>
                            <h1 className="text-xl font-bold text-white">Client Management</h1>
                            <p className="text-xs text-zinc-400">{summaryStats.totalClients} total clients</p>
                        </div>
                    </div>
                    <button
                        onClick={handleExportCSV}
                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm font-medium transition-colors"
                    >
                        <Download size={16} />
                        Export CSV
                    </button>
                </div>

                {/* Search and Filters */}
                <div className="flex items-center gap-3">
                    <div className="flex-1 relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={16} />
                        <input
                            type="text"
                            placeholder="Search by name, email, or client ID..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full pl-10 pr-10 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                        />
                        {searchQuery && (
                            <button
                                onClick={() => setSearchQuery('')}
                                className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-white"
                            >
                                <X size={16} />
                            </button>
                        )}
                    </div>

                    <select
                        value={kycFilter}
                        onChange={(e) => setKycFilter(e.target.value as any)}
                        className="px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-blue-500"
                    >
                        <option value="All">All KYC Status</option>
                        {KYC_STATUSES.map(status => (
                            <option key={status} value={status}>{status}</option>
                        ))}
                    </select>

                    <select
                        value={accountFilter}
                        onChange={(e) => setAccountFilter(e.target.value as any)}
                        className="px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-blue-500"
                    >
                        <option value="All">All Account Status</option>
                        {ACCOUNT_STATUSES.map(status => (
                            <option key={status} value={status}>{status}</option>
                        ))}
                    </select>
                </div>
            </div>

            {/* Summary Cards */}
            <div className="flex-shrink-0 grid grid-cols-4 gap-4 p-4 bg-zinc-900/50">
                <div className="bg-zinc-900 rounded-lg p-4 border border-zinc-800">
                    <div className="flex items-center justify-between mb-2">
                        <div className="text-xs text-zinc-400 uppercase tracking-wide">Total Clients</div>
                        <Users size={16} className="text-blue-400" />
                    </div>
                    <div className="text-2xl font-bold text-white">{summaryStats.totalClients}</div>
                </div>

                <div className="bg-zinc-900 rounded-lg p-4 border border-zinc-800">
                    <div className="flex items-center justify-between mb-2">
                        <div className="text-xs text-zinc-400 uppercase tracking-wide">Active Traders</div>
                        <Activity size={16} className="text-green-400" />
                    </div>
                    <div className="text-2xl font-bold text-white">{summaryStats.activeTraders}</div>
                    <div className="text-xs text-zinc-500 mt-1">Last 7 days</div>
                </div>

                <div className="bg-zinc-900 rounded-lg p-4 border border-zinc-800">
                    <div className="flex items-center justify-between mb-2">
                        <div className="text-xs text-zinc-400 uppercase tracking-wide">Pending KYC</div>
                        <UserCheck size={16} className="text-yellow-400" />
                    </div>
                    <div className="text-2xl font-bold text-white">{summaryStats.pendingKYC}</div>
                </div>

                <div className="bg-zinc-900 rounded-lg p-4 border border-zinc-800">
                    <div className="flex items-center justify-between mb-2">
                        <div className="text-xs text-zinc-400 uppercase tracking-wide">Total Client Equity</div>
                        <DollarSign size={16} className="text-emerald-400" />
                    </div>
                    <div className="text-2xl font-bold text-white">{formatCurrency(summaryStats.totalEquity)}</div>
                </div>
            </div>

            {/* Table */}
            <div className="flex-1 overflow-auto custom-scrollbar">
                <table className="w-full text-xs">
                    <thead className="sticky top-0 bg-zinc-900 border-b border-zinc-800 z-10">
                        <tr>
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400 w-8"></th>
                            <th
                                className="text-left px-3 py-2 font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                onClick={() => handleSort('clientId')}
                            >
                                Client ID {sortField === 'clientId' && (sortDirection === 'asc' ? '▲' : '▼')}
                            </th>
                            <th
                                className="text-left px-3 py-2 font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                onClick={() => handleSort('name')}
                            >
                                Name {sortField === 'name' && (sortDirection === 'asc' ? '▲' : '▼')}
                            </th>
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400">Email</th>
                            <th
                                className="text-right px-3 py-2 font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                onClick={() => handleSort('balance')}
                            >
                                Balance {sortField === 'balance' && (sortDirection === 'asc' ? '▲' : '▼')}
                            </th>
                            <th
                                className="text-right px-3 py-2 font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                onClick={() => handleSort('equity')}
                            >
                                Equity {sortField === 'equity' && (sortDirection === 'asc' ? '▲' : '▼')}
                            </th>
                            <th className="text-center px-3 py-2 font-semibold text-zinc-400">Open Pos</th>
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400">KYC Status</th>
                            <th
                                className="text-left px-3 py-2 font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                onClick={() => handleSort('registrationDate')}
                            >
                                Registration {sortField === 'registrationDate' && (sortDirection === 'asc' ? '▲' : '▼')}
                            </th>
                            <th
                                className="text-left px-3 py-2 font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                onClick={() => handleSort('lastLogin')}
                            >
                                Last Login {sortField === 'lastLogin' && (sortDirection === 'asc' ? '▲' : '▼')}
                            </th>
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {paginatedClients.map((client) => (
                            <React.Fragment key={client.id}>
                                <tr className="border-b border-zinc-800 hover:bg-zinc-900/50 cursor-pointer">
                                    <td
                                        className="px-3 py-2"
                                        onClick={() => handleToggleExpand(client.id)}
                                    >
                                        {expandedRow === client.id ? (
                                            <ChevronDown size={14} className="text-zinc-500" />
                                        ) : (
                                            <ChevronRight size={14} className="text-zinc-500" />
                                        )}
                                    </td>
                                    <td className="px-3 py-2 font-mono text-blue-400">{client.clientId}</td>
                                    <td className="px-3 py-2 font-medium text-white">{client.name}</td>
                                    <td className="px-3 py-2 text-zinc-400">{client.email}</td>
                                    <td className="px-3 py-2 text-right font-mono text-white">{formatCurrency(client.balance)}</td>
                                    <td className="px-3 py-2 text-right font-mono text-white">{formatCurrency(client.equity)}</td>
                                    <td className="px-3 py-2 text-center text-white">{client.openPositions}</td>
                                    <td className="px-3 py-2">
                                        <span className={`px-2 py-1 rounded text-xs font-medium ${KYC_STATUS_COLORS[client.kycStatus]}`}>
                                            {client.kycStatus}
                                        </span>
                                    </td>
                                    <td className="px-3 py-2 text-zinc-400">
                                        {new Date(client.registrationDate).toLocaleDateString()}
                                    </td>
                                    <td className="px-3 py-2 text-zinc-400">
                                        {formatRelativeTime(client.lastLogin)}
                                    </td>
                                    <td className="px-3 py-2">
                                        <div className="flex items-center gap-2">
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleSuspend(client);
                                                }}
                                                className={`p-1 rounded transition-colors ${
                                                    client.accountStatus === 'Active'
                                                        ? 'hover:bg-yellow-600/20 text-yellow-400'
                                                        : 'hover:bg-green-600/20 text-green-400'
                                                }`}
                                                title={client.accountStatus === 'Active' ? 'Suspend' : 'Activate'}
                                            >
                                                {client.accountStatus === 'Active' ? <UserX size={14} /> : <UserCheck size={14} />}
                                            </button>
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleSendNotification(client);
                                                }}
                                                className="p-1 hover:bg-blue-600/20 text-blue-400 rounded transition-colors"
                                                title="Send Notification"
                                            >
                                                <Bell size={14} />
                                            </button>
                                        </div>
                                    </td>
                                </tr>

                                {/* Expandable Detail */}
                                {expandedRow === client.id && (
                                    <tr className="bg-zinc-900/70 border-b border-zinc-800">
                                        <td colSpan={11} className="px-6 py-4">
                                            <div className="grid grid-cols-2 gap-6">
                                                {/* Left Column */}
                                                <div className="space-y-4">
                                                    {/* Trading Stats */}
                                                    <div>
                                                        <h3 className="text-sm font-bold text-white mb-3 flex items-center gap-2">
                                                            <Activity size={16} className="text-blue-400" />
                                                            Trading Statistics
                                                        </h3>
                                                        <div className="grid grid-cols-2 gap-3">
                                                            <div className="bg-zinc-800 rounded p-3">
                                                                <div className="text-xs text-zinc-400 mb-1">Total Trades</div>
                                                                <div className="text-lg font-bold text-white">{client.tradingStats.totalTrades}</div>
                                                            </div>
                                                            <div className="bg-zinc-800 rounded p-3">
                                                                <div className="text-xs text-zinc-400 mb-1">Win Rate</div>
                                                                <div className="text-lg font-bold text-green-400">{client.tradingStats.winRate.toFixed(1)}%</div>
                                                            </div>
                                                            <div className="bg-zinc-800 rounded p-3">
                                                                <div className="text-xs text-zinc-400 mb-1">Profit Factor</div>
                                                                <div className="text-lg font-bold text-white">{client.tradingStats.profitFactor}</div>
                                                            </div>
                                                            <div className="bg-zinc-800 rounded p-3">
                                                                <div className="text-xs text-zinc-400 mb-1">Avg Trade Size</div>
                                                                <div className="text-lg font-bold text-white">{client.tradingStats.avgTradeSize} lots</div>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    {/* Financial Summary */}
                                                    <div>
                                                        <h3 className="text-sm font-bold text-white mb-3 flex items-center gap-2">
                                                            <DollarSign size={16} className="text-emerald-400" />
                                                            Financial Summary
                                                        </h3>
                                                        <div className="bg-zinc-800 rounded p-3 space-y-2">
                                                            <div className="flex justify-between">
                                                                <span className="text-xs text-zinc-400">Total Deposits</span>
                                                                <span className="text-xs font-mono text-green-400">
                                                                    {formatCurrency(client.financialSummary.totalDeposits)}
                                                                </span>
                                                            </div>
                                                            <div className="flex justify-between">
                                                                <span className="text-xs text-zinc-400">Total Withdrawals</span>
                                                                <span className="text-xs font-mono text-red-400">
                                                                    {formatCurrency(client.financialSummary.totalWithdrawals)}
                                                                </span>
                                                            </div>
                                                            <div className="flex justify-between pt-2 border-t border-zinc-700">
                                                                <span className="text-xs font-bold text-white">Net Deposits</span>
                                                                <span className={`text-xs font-mono font-bold ${
                                                                    client.financialSummary.netDeposits >= 0 ? 'text-green-400' : 'text-red-400'
                                                                }`}>
                                                                    {formatCurrency(client.financialSummary.netDeposits)}
                                                                </span>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    {/* Account Notes */}
                                                    <div>
                                                        <h3 className="text-sm font-bold text-white mb-3">Account Notes</h3>
                                                        <textarea
                                                            value={client.notes}
                                                            onChange={(e) => {
                                                                const newNotes = e.target.value;
                                                                setClients(prev =>
                                                                    prev.map(c =>
                                                                        c.id === client.id ? { ...c, notes: newNotes } : c
                                                                    )
                                                                );
                                                            }}
                                                            placeholder="Add notes about this client..."
                                                            className="w-full h-20 px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-xs text-white placeholder-zinc-500 resize-none focus:outline-none focus:border-blue-500"
                                                        />
                                                    </div>
                                                </div>

                                                {/* Right Column - Recent Trades */}
                                                <div>
                                                    <h3 className="text-sm font-bold text-white mb-3 flex items-center gap-2">
                                                        <TrendingUp size={16} className="text-blue-400" />
                                                        Recent Trades (Last 5)
                                                    </h3>
                                                    <div className="bg-zinc-800 rounded overflow-hidden">
                                                        <table className="w-full text-xs">
                                                            <thead className="bg-zinc-900">
                                                                <tr>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Symbol</th>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Type</th>
                                                                    <th className="text-right px-3 py-2 text-zinc-400 font-semibold">Volume</th>
                                                                    <th className="text-right px-3 py-2 text-zinc-400 font-semibold">P&L</th>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Closed</th>
                                                                </tr>
                                                            </thead>
                                                            <tbody>
                                                                {client.recentTrades.map((trade) => (
                                                                    <tr key={trade.id} className="border-t border-zinc-700">
                                                                        <td className="px-3 py-2 font-mono text-blue-400">{trade.symbol}</td>
                                                                        <td className="px-3 py-2">
                                                                            <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                                                                                trade.type === 'BUY' ? 'bg-green-600 text-white' : 'bg-red-600 text-white'
                                                                            }`}>
                                                                                {trade.type}
                                                                            </span>
                                                                        </td>
                                                                        <td className="px-3 py-2 text-right font-mono text-white">{trade.volume}</td>
                                                                        <td className={`px-3 py-2 text-right font-mono font-medium ${
                                                                            trade.pnl >= 0 ? 'text-green-400' : 'text-red-400'
                                                                        }`}>
                                                                            {trade.pnl >= 0 ? '+' : ''}{formatCurrency(trade.pnl)}
                                                                        </td>
                                                                        <td className="px-3 py-2 text-zinc-400">
                                                                            {formatRelativeTime(trade.closedAt)}
                                                                        </td>
                                                                    </tr>
                                                                ))}
                                                            </tbody>
                                                        </table>
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

                {paginatedClients.length === 0 && (
                    <div className="text-center py-12 text-zinc-500">
                        No clients found matching your criteria
                    </div>
                )}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
                <div className="flex-shrink-0 flex items-center justify-between px-4 py-3 bg-zinc-900 border-t border-zinc-800">
                    <div className="text-xs text-zinc-400">
                        Showing {(currentPage - 1) * pageSize + 1} to {Math.min(currentPage * pageSize, filteredAndSortedClients.length)} of {filteredAndSortedClients.length} clients
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            onClick={() => setCurrentPage(1)}
                            disabled={currentPage === 1}
                            className="px-3 py-1 bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed rounded text-xs transition-colors"
                        >
                            First
                        </button>
                        <button
                            onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                            disabled={currentPage === 1}
                            className="px-3 py-1 bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed rounded text-xs transition-colors"
                        >
                            Previous
                        </button>
                        <span className="px-3 py-1 text-xs text-zinc-400">
                            Page {currentPage} of {totalPages}
                        </span>
                        <button
                            onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                            disabled={currentPage === totalPages}
                            className="px-3 py-1 bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed rounded text-xs transition-colors"
                        >
                            Next
                        </button>
                        <button
                            onClick={() => setCurrentPage(totalPages)}
                            disabled={currentPage === totalPages}
                            className="px-3 py-1 bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed rounded text-xs transition-colors"
                        >
                            Last
                        </button>
                    </div>
                </div>
            )}

            {/* Notification Modal */}
            {showNotificationModal && notificationClient && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
                    <div className="bg-zinc-900 rounded-lg shadow-xl border border-zinc-700 w-full max-w-md">
                        <div className="flex items-center justify-between p-4 border-b border-zinc-800">
                            <h2 className="text-lg font-bold text-white">Send Notification</h2>
                            <button
                                onClick={() => setShowNotificationModal(false)}
                                className="text-zinc-400 hover:text-white transition-colors"
                            >
                                <X size={20} />
                            </button>
                        </div>

                        <div className="p-4 space-y-4">
                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-1">To</label>
                                <div className="text-sm text-white">{notificationClient.name} ({notificationClient.email})</div>
                            </div>

                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-1">Subject</label>
                                <input
                                    type="text"
                                    value={notificationSubject}
                                    onChange={(e) => setNotificationSubject(e.target.value)}
                                    placeholder="Enter subject"
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                                />
                            </div>

                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-1">Message</label>
                                <textarea
                                    value={notificationMessage}
                                    onChange={(e) => setNotificationMessage(e.target.value)}
                                    placeholder="Enter message"
                                    className="w-full h-32 px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white placeholder-zinc-500 resize-none focus:outline-none focus:border-blue-500"
                                />
                            </div>
                        </div>

                        <div className="flex items-center justify-end gap-3 p-4 border-t border-zinc-800">
                            <button
                                onClick={() => setShowNotificationModal(false)}
                                className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded text-sm font-medium transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={submitNotification}
                                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm font-medium transition-colors"
                            >
                                Send Notification
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
