'use client';

import { useState, useEffect, useMemo } from 'react';
import {
    Users,
    UserCheck,
    UserX,
    TrendingUp,
    DollarSign,
    Calendar,
    Plus,
    Search,
    ChevronDown,
    ChevronUp,
    X,
    RotateCcw,
    Trash2,
    Clock,
    CheckCircle,
    AlertTriangle,
    BarChart3
} from 'lucide-react';
import { API_CONFIG } from '../../config/api';

type AccountStatus = 'active' | 'expired' | 'converted';

interface DemoAccount {
    id: string;
    owner_name: string;
    email: string;
    balance: number;
    initial_balance: number;
    leverage: number;
    currency: string;
    created_at: string;
    expires_at: string;
    is_active: boolean;
    total_trades: number;
    total_pnl: number;
    win_rate: number;
    converted_at?: string;
    converted_to_id?: string;
    last_activity_at: string;
}

interface DemoPosition {
    id: string;
    account_id: string;
    symbol: string;
    side: string;
    volume: number;
    open_price: number;
    current_price: number;
    stop_loss?: number;
    take_profit?: number;
    pnl: number;
    pnl_pct: number;
    open_time: string;
    commission: number;
    swap: number;
}

interface DemoStats {
    total_accounts: number;
    active_accounts: number;
    expired_accounts: number;
    converted_accounts: number;
    avg_win_rate: number;
    avg_pnl: number;
    most_traded_symbol: string;
    avg_trades_per_account: number;
    conversion_rate: number;
    total_positions: number;
    total_volume: number;
}

interface ConversionRecord {
    demo_account_id: string;
    owner_name: string;
    email: string;
    demo_balance: number;
    demo_pnl: number;
    demo_trades: number;
    demo_win_rate: number;
    converted_at: string;
    live_account_id: string;
    initial_deposit: number;
    days_until_conversion: number;
}

export default function DemoAccountManager() {
    const [loading, setLoading] = useState(true);
    const [stats, setStats] = useState<DemoStats | null>(null);
    const [accounts, setAccounts] = useState<DemoAccount[]>([]);
    const [conversions, setConversions] = useState<ConversionRecord[]>([]);
    const [selectedAccount, setSelectedAccount] = useState<DemoAccount | null>(null);
    const [accountPositions, setAccountPositions] = useState<DemoPosition[]>([]);
    const [expandedAccountId, setExpandedAccountId] = useState<string | null>(null);
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showActionModal, setShowActionModal] = useState<{
        type: 'extend' | 'reset' | 'leverage' | 'deactivate';
        account: DemoAccount;
    } | null>(null);

    // Filters
    const [statusFilter, setStatusFilter] = useState<'all' | AccountStatus>('all');
    const [searchQuery, setSearchQuery] = useState('');

    // Sorting
    const [sortConfig, setSortConfig] = useState<{
        key: keyof DemoAccount;
        direction: 'asc' | 'desc';
    }>({ key: 'created_at', direction: 'desc' });

    // Create form state
    const [createForm, setCreateForm] = useState({
        owner_name: '',
        email: '',
        initial_balance: 10000,
        leverage: 100,
        expiry_days: 30,
        currency: 'USD'
    });

    useEffect(() => {
        fetchAllData();
    }, []);

    const fetchAllData = async () => {
        setLoading(true);
        try {
            const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') || '' : '';
            const headers = {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            };

            const [statsRes, accountsRes, conversionsRes] = await Promise.all([
                fetch(API_CONFIG.DEMO_ACCOUNTS_STATS, { headers }),
                fetch(API_CONFIG.DEMO_ACCOUNTS_LIST, { headers }),
                fetch(API_CONFIG.DEMO_ACCOUNTS_CONVERSIONS, { headers })
            ]);

            if (statsRes.ok) {
                const data = await statsRes.json();
                setStats(data.stats || data);
            }

            if (accountsRes.ok) {
                const data = await accountsRes.json();
                setAccounts(data.data || data.accounts || []);
            }

            if (conversionsRes.ok) {
                const data = await conversionsRes.json();
                setConversions(data.conversions || []);
            }
        } catch (error) {
            console.error('[DemoAccountManager] Error fetching data:', error);
        } finally {
            setLoading(false);
        }
    };

    const fetchAccountPositions = async (accountId: string) => {
        try {
            const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') || '' : '';
            const res = await fetch(API_CONFIG.DEMO_ACCOUNT_POSITIONS(accountId), {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (res.ok) {
                const data = await res.json();
                setAccountPositions(data.positions || []);
            }
        } catch (error) {
            console.error('[DemoAccountManager] Error fetching positions:', error);
        }
    };

    const handleCreateAccount = async () => {
        try {
            const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') || '' : '';
            const res = await fetch(API_CONFIG.DEMO_ACCOUNTS_CREATE, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(createForm)
            });

            if (res.ok) {
                setShowCreateModal(false);
                setCreateForm({
                    owner_name: '',
                    email: '',
                    initial_balance: 10000,
                    leverage: 100,
                    expiry_days: 30,
                    currency: 'USD'
                });
                fetchAllData();
            }
        } catch (error) {
            console.error('[DemoAccountManager] Error creating account:', error);
        }
    };

    const handleAccountAction = async () => {
        if (!showActionModal) return;

        try {
            const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') || '' : '';
            let endpoint = '';
            let body: any = {};

            switch (showActionModal.type) {
                case 'extend':
                    endpoint = API_CONFIG.DEMO_ACCOUNT_UPDATE(showActionModal.account.id);
                    body = { extend_days: 30 };
                    break;
                case 'reset':
                    endpoint = API_CONFIG.DEMO_ACCOUNT_UPDATE(showActionModal.account.id);
                    body = { reset_balance: true };
                    break;
                case 'leverage':
                    endpoint = API_CONFIG.DEMO_ACCOUNT_UPDATE(showActionModal.account.id);
                    body = { leverage: 200 };
                    break;
                case 'deactivate':
                    endpoint = API_CONFIG.DEMO_ACCOUNT_DELETE(showActionModal.account.id);
                    break;
            }

            const method = showActionModal.type === 'deactivate' ? 'DELETE' : 'PUT';
            const res = await fetch(endpoint, {
                method,
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: method !== 'DELETE' ? JSON.stringify(body) : undefined
            });

            if (res.ok) {
                setShowActionModal(null);
                fetchAllData();
            }
        } catch (error) {
            console.error('[DemoAccountManager] Error performing action:', error);
        }
    };

    const filteredAndSortedAccounts = useMemo(() => {
        let filtered = [...accounts];

        // Status filter
        if (statusFilter === 'active') {
            filtered = filtered.filter(acc => acc.is_active && !acc.converted_at);
        } else if (statusFilter === 'expired') {
            filtered = filtered.filter(acc => !acc.is_active);
        } else if (statusFilter === 'converted') {
            filtered = filtered.filter(acc => acc.converted_at);
        }

        // Search filter
        if (searchQuery) {
            const query = searchQuery.toLowerCase();
            filtered = filtered.filter(acc =>
                acc.owner_name.toLowerCase().includes(query) ||
                acc.email.toLowerCase().includes(query)
            );
        }

        // Sort
        filtered.sort((a, b) => {
            const aVal = a[sortConfig.key];
            const bVal = b[sortConfig.key];

            if (typeof aVal === 'string' && typeof bVal === 'string') {
                return sortConfig.direction === 'asc'
                    ? aVal.localeCompare(bVal)
                    : bVal.localeCompare(aVal);
            }

            if (typeof aVal === 'number' && typeof bVal === 'number') {
                return sortConfig.direction === 'asc' ? aVal - bVal : bVal - aVal;
            }

            return 0;
        });

        return filtered;
    }, [accounts, statusFilter, searchQuery, sortConfig]);

    const handleSort = (key: keyof DemoAccount) => {
        setSortConfig(prev => ({
            key,
            direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
        }));
    };

    const getStatusBadge = (account: DemoAccount) => {
        if (account.converted_at) {
            return <span className="px-2 py-0.5 text-[10px] font-bold bg-[#10B981] text-white rounded">CONVERTED</span>;
        }
        if (!account.is_active) {
            return <span className="px-2 py-0.5 text-[10px] font-bold bg-[#EF4444] text-white rounded">EXPIRED</span>;
        }
        return <span className="px-2 py-0.5 text-[10px] font-bold bg-[#3B82F6] text-white rounded">ACTIVE</span>;
    };

    // Monthly conversions chart data
    const monthlyConversions = useMemo(() => {
        const months: { [key: string]: number } = {};
        conversions.forEach(conv => {
            const month = new Date(conv.converted_at).toLocaleString('en', { month: 'short', year: 'numeric' });
            months[month] = (months[month] || 0) + 1;
        });
        return Object.entries(months).slice(-12);
    }, [conversions]);

    const maxConversions = Math.max(...monthlyConversions.map(([_, count]) => count), 1);

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full bg-[#121316] text-[#888]">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#F5C542] mx-auto mb-4"></div>
                    <div>Loading demo accounts...</div>
                </div>
            </div>
        );
    }

    return (
        <div className="h-full flex flex-col bg-[#121316] text-[#CCC] overflow-hidden">
            {/* Header */}
            <div className="flex-shrink-0 px-4 py-3 bg-[#1E2026] border-b border-[#383A42] flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <Users size={18} className="text-[#3B82F6]" />
                    <h2 className="text-sm font-bold text-white">Demo Account Manager</h2>
                </div>
                <button
                    onClick={() => setShowCreateModal(true)}
                    className="px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs font-bold rounded flex items-center gap-1.5"
                >
                    <Plus size={14} />
                    Create Demo Account
                </button>
            </div>

            {/* Stats Cards */}
            {stats && (
                <div className="flex-shrink-0 grid grid-cols-5 gap-3 p-4 bg-[#1E2026] border-b border-[#383A42]">
                    <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                        <div className="flex items-center gap-2 mb-1">
                            <Users size={14} className="text-[#3B82F6]" />
                            <div className="text-[10px] text-[#888]">Total Accounts</div>
                        </div>
                        <div className="text-xl font-bold text-white">{stats.total_accounts}</div>
                    </div>
                    <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                        <div className="flex items-center gap-2 mb-1">
                            <UserCheck size={14} className="text-[#10B981]" />
                            <div className="text-[10px] text-[#888]">Active</div>
                        </div>
                        <div className="text-xl font-bold text-[#10B981]">{stats.active_accounts}</div>
                    </div>
                    <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                        <div className="flex items-center gap-2 mb-1">
                            <UserX size={14} className="text-[#EF4444]" />
                            <div className="text-[10px] text-[#888]">Expired</div>
                        </div>
                        <div className="text-xl font-bold text-[#EF4444]">{stats.expired_accounts}</div>
                    </div>
                    <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                        <div className="flex items-center gap-2 mb-1">
                            <TrendingUp size={14} className="text-[#F5C542]" />
                            <div className="text-[10px] text-[#888]">Conversion Rate</div>
                        </div>
                        <div className="text-xl font-bold text-[#F5C542]">{stats.conversion_rate.toFixed(1)}%</div>
                    </div>
                    <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                        <div className="flex items-center gap-2 mb-1">
                            <BarChart3 size={14} className="text-[#9B59B6]" />
                            <div className="text-[10px] text-[#888]">Avg Win Rate</div>
                        </div>
                        <div className="text-xl font-bold text-[#9B59B6]">{stats.avg_win_rate.toFixed(1)}%</div>
                    </div>
                </div>
            )}

            {/* Filter Bar */}
            <div className="flex-shrink-0 px-4 py-2 bg-[#1E2026] border-b border-[#383A42] flex items-center gap-3">
                <div className="flex items-center gap-2">
                    <div className="text-xs text-[#888]">Status:</div>
                    {(['all', 'active', 'expired', 'converted'] as const).map(status => (
                        <button
                            key={status}
                            onClick={() => setStatusFilter(status)}
                            className={`px-3 py-1 text-xs font-bold rounded ${
                                statusFilter === status
                                    ? 'bg-[#3B82F6] text-white'
                                    : 'bg-[#121316] text-[#888] hover:bg-[#25272E]'
                            }`}
                        >
                            {status.charAt(0).toUpperCase() + status.slice(1)}
                        </button>
                    ))}
                </div>
                <div className="flex-1"></div>
                <div className="relative">
                    <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-[#888]" />
                    <input
                        type="text"
                        placeholder="Search by name or email..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-8 pr-3 py-1 bg-[#121316] border border-[#383A42] rounded text-xs text-white placeholder-[#666] focus:outline-none focus:border-[#3B82F6] w-64"
                    />
                </div>
            </div>

            {/* Main Content */}
            <div className="flex-1 flex gap-4 p-4 overflow-hidden">
                {/* Account List Table */}
                <div className="flex-1 flex flex-col bg-[#1E2026] border border-[#383A42] rounded overflow-hidden">
                    <div className="flex-shrink-0 px-3 py-2 bg-[#121316] border-b border-[#383A42] flex items-center gap-2">
                        <Users size={14} className="text-[#3B82F6]" />
                        <div className="text-xs font-bold text-white">Demo Accounts ({filteredAndSortedAccounts.length})</div>
                    </div>
                    <div className="flex-1 overflow-auto custom-scrollbar">
                        <table className="w-full text-xs">
                            <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                                <tr>
                                    <th className="text-left px-3 py-2 text-[#888] font-bold cursor-pointer hover:text-white" onClick={() => handleSort('owner_name')}>
                                        Owner Name {sortConfig.key === 'owner_name' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th className="text-left px-3 py-2 text-[#888] font-bold cursor-pointer hover:text-white" onClick={() => handleSort('email')}>
                                        Email {sortConfig.key === 'email' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th className="text-right px-3 py-2 text-[#888] font-bold cursor-pointer hover:text-white" onClick={() => handleSort('balance')}>
                                        Balance {sortConfig.key === 'balance' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th className="text-center px-3 py-2 text-[#888] font-bold cursor-pointer hover:text-white" onClick={() => handleSort('leverage')}>
                                        Leverage {sortConfig.key === 'leverage' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th className="text-center px-3 py-2 text-[#888] font-bold">Status</th>
                                    <th className="text-right px-3 py-2 text-[#888] font-bold cursor-pointer hover:text-white" onClick={() => handleSort('total_trades')}>
                                        Trades {sortConfig.key === 'total_trades' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th className="text-right px-3 py-2 text-[#888] font-bold cursor-pointer hover:text-white" onClick={() => handleSort('win_rate')}>
                                        Win Rate {sortConfig.key === 'win_rate' && (sortConfig.direction === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th className="text-center px-3 py-2 text-[#888] font-bold">Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {filteredAndSortedAccounts.map(account => (
                                    <>
                                        <tr
                                            key={account.id}
                                            className="border-b border-[#383A42] hover:bg-[#25272E] cursor-pointer"
                                            onClick={() => {
                                                setSelectedAccount(account);
                                                if (expandedAccountId !== account.id) {
                                                    setExpandedAccountId(account.id);
                                                    fetchAccountPositions(account.id);
                                                } else {
                                                    setExpandedAccountId(null);
                                                }
                                            }}
                                        >
                                            <td className="px-3 py-2 text-white font-bold">{account.owner_name}</td>
                                            <td className="px-3 py-2 text-[#888]">{account.email}</td>
                                            <td className="px-3 py-2 text-right">
                                                <span className={account.balance >= account.initial_balance ? 'text-[#10B981]' : 'text-[#EF4444]'}>
                                                    ${account.balance.toFixed(2)}
                                                </span>
                                            </td>
                                            <td className="px-3 py-2 text-center text-[#888]">1:{account.leverage}</td>
                                            <td className="px-3 py-2 text-center">{getStatusBadge(account)}</td>
                                            <td className="px-3 py-2 text-right text-[#888]">{account.total_trades}</td>
                                            <td className="px-3 py-2 text-right text-[#888]">{account.win_rate.toFixed(1)}%</td>
                                            <td className="px-3 py-2 text-center">
                                                <div className="flex items-center justify-center gap-1">
                                                    <button
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            setShowActionModal({ type: 'extend', account });
                                                        }}
                                                        className="p-1 hover:bg-[#383A42] rounded text-[#3B82F6]"
                                                        title="Extend Expiry"
                                                    >
                                                        <Clock size={14} />
                                                    </button>
                                                    <button
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            setShowActionModal({ type: 'reset', account });
                                                        }}
                                                        className="p-1 hover:bg-[#383A42] rounded text-[#F5C542]"
                                                        title="Reset Balance"
                                                    >
                                                        <RotateCcw size={14} />
                                                    </button>
                                                    <button
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            setShowActionModal({ type: 'deactivate', account });
                                                        }}
                                                        className="p-1 hover:bg-[#383A42] rounded text-[#EF4444]"
                                                        title="Deactivate"
                                                    >
                                                        <Trash2 size={14} />
                                                    </button>
                                                    {expandedAccountId === account.id ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                                                </div>
                                            </td>
                                        </tr>
                                        {/* Position Viewer - Expandable Row */}
                                        {expandedAccountId === account.id && (
                                            <tr className="bg-[#121316]">
                                                <td colSpan={8} className="px-3 py-3">
                                                    <div className="text-xs font-bold text-white mb-2">Open Positions ({accountPositions.length})</div>
                                                    {accountPositions.length === 0 ? (
                                                        <div className="text-xs text-[#666] italic">No open positions</div>
                                                    ) : (
                                                        <table className="w-full text-xs">
                                                            <thead>
                                                                <tr className="border-b border-[#383A42]">
                                                                    <th className="text-left py-1 text-[#888]">Symbol</th>
                                                                    <th className="text-center py-1 text-[#888]">Side</th>
                                                                    <th className="text-right py-1 text-[#888]">Volume</th>
                                                                    <th className="text-right py-1 text-[#888]">Open Price</th>
                                                                    <th className="text-right py-1 text-[#888]">Current Price</th>
                                                                    <th className="text-right py-1 text-[#888]">P&L</th>
                                                                    <th className="text-right py-1 text-[#888]">P&L %</th>
                                                                </tr>
                                                            </thead>
                                                            <tbody>
                                                                {accountPositions.map(pos => (
                                                                    <tr key={pos.id} className="border-b border-[#25272E]">
                                                                        <td className="py-1 text-white">{pos.symbol}</td>
                                                                        <td className="text-center py-1">
                                                                            <span className={pos.side === 'buy' ? 'text-[#10B981]' : 'text-[#EF4444]'}>
                                                                                {pos.side.toUpperCase()}
                                                                            </span>
                                                                        </td>
                                                                        <td className="text-right py-1 text-[#888]">{pos.volume.toFixed(2)}</td>
                                                                        <td className="text-right py-1 text-[#888]">{pos.open_price.toFixed(5)}</td>
                                                                        <td className="text-right py-1 text-[#888]">{pos.current_price.toFixed(5)}</td>
                                                                        <td className={`text-right py-1 ${pos.pnl >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                                                                            ${pos.pnl.toFixed(2)}
                                                                        </td>
                                                                        <td className={`text-right py-1 ${pos.pnl_pct >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                                                                            {pos.pnl_pct >= 0 ? '+' : ''}{pos.pnl_pct.toFixed(2)}%
                                                                        </td>
                                                                    </tr>
                                                                ))}
                                                            </tbody>
                                                        </table>
                                                    )}
                                                </td>
                                            </tr>
                                        )}
                                    </>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>

                {/* Conversion Tracking Panel */}
                <div className="w-96 flex flex-col bg-[#1E2026] border border-[#383A42] rounded overflow-hidden">
                    <div className="flex-shrink-0 px-3 py-2 bg-[#121316] border-b border-[#383A42] flex items-center gap-2">
                        <TrendingUp size={14} className="text-[#10B981]" />
                        <div className="text-xs font-bold text-white">Conversion Tracking</div>
                    </div>

                    {/* Monthly Conversions Bar Chart */}
                    <div className="flex-shrink-0 p-3 border-b border-[#383A42]">
                        <div className="text-xs text-[#888] mb-2">Monthly Conversions (Last 12 Months)</div>
                        <div className="h-32 flex items-end gap-1">
                            {monthlyConversions.map(([month, count]) => (
                                <div key={month} className="flex-1 flex flex-col items-center gap-1">
                                    <div
                                        className="w-full bg-[#10B981] rounded-t"
                                        style={{ height: `${(count / maxConversions) * 100}%` }}
                                        title={`${month}: ${count} conversions`}
                                    ></div>
                                    <div className="text-[9px] text-[#666] rotate-45 origin-top-left whitespace-nowrap">
                                        {month.split(' ')[0]}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Conversion List */}
                    <div className="flex-1 overflow-auto custom-scrollbar">
                        {conversions.slice(0, 20).map(conv => (
                            <div key={conv.demo_account_id} className="px-3 py-2 border-b border-[#383A42] hover:bg-[#25272E]">
                                <div className="flex items-center justify-between mb-1">
                                    <div className="text-xs font-bold text-white">{conv.owner_name}</div>
                                    <CheckCircle size={12} className="text-[#10B981]" />
                                </div>
                                <div className="text-[10px] text-[#666] mb-1">{conv.email}</div>
                                <div className="grid grid-cols-2 gap-2 text-[10px]">
                                    <div>
                                        <span className="text-[#888]">Demo P&L:</span>{' '}
                                        <span className={conv.demo_pnl >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}>
                                            ${conv.demo_pnl.toFixed(2)}
                                        </span>
                                    </div>
                                    <div>
                                        <span className="text-[#888]">Trades:</span>{' '}
                                        <span className="text-white">{conv.demo_trades}</span>
                                    </div>
                                    <div>
                                        <span className="text-[#888]">Win Rate:</span>{' '}
                                        <span className="text-white">{conv.demo_win_rate.toFixed(1)}%</span>
                                    </div>
                                    <div>
                                        <span className="text-[#888]">Days:</span>{' '}
                                        <span className="text-white">{conv.days_until_conversion}d</span>
                                    </div>
                                    <div className="col-span-2">
                                        <span className="text-[#888]">Initial Deposit:</span>{' '}
                                        <span className="text-[#10B981] font-bold">${conv.initial_deposit.toFixed(2)}</span>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>

            {/* Create Modal */}
            {showCreateModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-[#1E2026] border border-[#383A42] rounded-lg shadow-xl w-96">
                        <div className="px-4 py-3 border-b border-[#383A42] flex items-center justify-between">
                            <div className="text-sm font-bold text-white">Create Demo Account</div>
                            <button onClick={() => setShowCreateModal(false)} className="text-[#888] hover:text-white">
                                <X size={16} />
                            </button>
                        </div>
                        <div className="p-4 space-y-3">
                            <div>
                                <label className="block text-xs text-[#888] mb-1">Owner Name</label>
                                <input
                                    type="text"
                                    value={createForm.owner_name}
                                    onChange={(e) => setCreateForm({ ...createForm, owner_name: e.target.value })}
                                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-xs text-white"
                                />
                            </div>
                            <div>
                                <label className="block text-xs text-[#888] mb-1">Email</label>
                                <input
                                    type="email"
                                    value={createForm.email}
                                    onChange={(e) => setCreateForm({ ...createForm, email: e.target.value })}
                                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-xs text-white"
                                />
                            </div>
                            <div>
                                <label className="block text-xs text-[#888] mb-1">Starting Balance</label>
                                <select
                                    value={createForm.initial_balance}
                                    onChange={(e) => setCreateForm({ ...createForm, initial_balance: Number(e.target.value) })}
                                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-xs text-white"
                                >
                                    <option value={10000}>$10,000</option>
                                    <option value={25000}>$25,000</option>
                                    <option value={50000}>$50,000</option>
                                    <option value={100000}>$100,000</option>
                                </select>
                            </div>
                            <div>
                                <label className="block text-xs text-[#888] mb-1">Leverage</label>
                                <select
                                    value={createForm.leverage}
                                    onChange={(e) => setCreateForm({ ...createForm, leverage: Number(e.target.value) })}
                                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-xs text-white"
                                >
                                    <option value={50}>50:1</option>
                                    <option value={100}>100:1</option>
                                    <option value={200}>200:1</option>
                                    <option value={500}>500:1</option>
                                </select>
                            </div>
                            <div>
                                <label className="block text-xs text-[#888] mb-1">Expiry Days</label>
                                <input
                                    type="number"
                                    value={createForm.expiry_days}
                                    onChange={(e) => setCreateForm({ ...createForm, expiry_days: Number(e.target.value) })}
                                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-xs text-white"
                                />
                            </div>
                            <div>
                                <label className="block text-xs text-[#888] mb-1">Currency</label>
                                <select
                                    value={createForm.currency}
                                    onChange={(e) => setCreateForm({ ...createForm, currency: e.target.value })}
                                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-xs text-white"
                                >
                                    <option value="USD">USD</option>
                                    <option value="EUR">EUR</option>
                                    <option value="GBP">GBP</option>
                                </select>
                            </div>
                        </div>
                        <div className="px-4 py-3 border-t border-[#383A42] flex justify-end gap-2">
                            <button
                                onClick={() => setShowCreateModal(false)}
                                className="px-3 py-1.5 bg-[#383A42] hover:bg-[#4A4C54] text-white text-xs font-bold rounded"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleCreateAccount}
                                className="px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs font-bold rounded"
                            >
                                Create Account
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Action Confirmation Modal */}
            {showActionModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-[#1E2026] border border-[#383A42] rounded-lg shadow-xl w-96">
                        <div className="px-4 py-3 border-b border-[#383A42]">
                            <div className="text-sm font-bold text-white">Confirm Action</div>
                        </div>
                        <div className="p-4">
                            <div className="text-xs text-[#CCC]">
                                {showActionModal.type === 'extend' && `Extend expiry for ${showActionModal.account.owner_name} by 30 days?`}
                                {showActionModal.type === 'reset' && `Reset balance to initial amount for ${showActionModal.account.owner_name}?`}
                                {showActionModal.type === 'leverage' && `Change leverage for ${showActionModal.account.owner_name}?`}
                                {showActionModal.type === 'deactivate' && `Deactivate account for ${showActionModal.account.owner_name}? This cannot be undone.`}
                            </div>
                        </div>
                        <div className="px-4 py-3 border-t border-[#383A42] flex justify-end gap-2">
                            <button
                                onClick={() => setShowActionModal(null)}
                                className="px-3 py-1.5 bg-[#383A42] hover:bg-[#4A4C54] text-white text-xs font-bold rounded"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleAccountAction}
                                className={`px-3 py-1.5 text-white text-xs font-bold rounded ${
                                    showActionModal.type === 'deactivate'
                                        ? 'bg-[#EF4444] hover:bg-[#DC2626]'
                                        : 'bg-[#3B82F6] hover:bg-[#2563EB]'
                                }`}
                            >
                                Confirm
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
