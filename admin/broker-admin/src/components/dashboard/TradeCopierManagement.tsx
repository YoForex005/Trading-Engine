'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
    Copy,
    Plus,
    Edit,
    Trash2,
    Play,
    Pause,
    Search,
    X,
    ChevronDown,
    ChevronRight,
    TrendingUp,
    TrendingDown,
    Users,
    Activity,
    BarChart3,
    CheckCircle,
    XCircle,
} from 'lucide-react';
import { api } from '@/services/apiClient';
import { API_CONFIG } from '@/config/api';

type CopyMode = 'proportional' | 'fixed' | 'inverse';
type CopyStatus = 'active' | 'paused' | 'stopped';
type CopyAction = 'open' | 'close' | 'modify';
type LogStatus = 'success' | 'failed';

interface Account {
    id: string;
    accountId: string;
    name: string;
    balance: number;
}

interface CopyRelation {
    id: string;
    masterId: string;
    masterName: string;
    masterAccountId: string;
    followerId: string;
    followerName: string;
    followerAccountId: string;
    copyMode: CopyMode;
    lotMultiplier?: number; // For proportional
    fixedLotSize?: number; // For fixed
    maxLots: number;
    skipSymbols: string[];
    status: CopyStatus;
    createdAt: string;
    tradesCopied: number;
}

interface CopyLogEntry {
    id: string;
    relationId: string;
    timestamp: string;
    masterTradeId: string;
    followerTradeId: string | null;
    action: CopyAction;
    symbol: string;
    volume: number;
    status: LogStatus;
    error: string | null;
}

interface MasterPerformance {
    accountId: string;
    winRate: number;
    totalTrades: number;
    pnl: number;
    followersCount: number;
    equityCurve: number[];
}

interface FormData {
    masterId: string;
    followerId: string;
    copyMode: CopyMode;
    lotMultiplier: number;
    fixedLotSize: number;
    maxLots: number;
    skipSymbols: string[];
    isActive: boolean;
}

const COPY_MODE_COLORS: Record<CopyMode, string> = {
    proportional: 'bg-blue-600 text-white',
    fixed: 'bg-green-600 text-white',
    inverse: 'bg-purple-600 text-white',
};

const STATUS_COLORS: Record<CopyStatus, string> = {
    active: 'bg-green-600 text-white',
    paused: 'bg-yellow-600 text-black',
    stopped: 'bg-red-600 text-white',
};

const AVAILABLE_SYMBOLS = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'BTCUSD', 'XAUUSD'];

// Mock data generators
const generateMockAccounts = (): Account[] => {
    const names = ['Alice Johnson', 'Bob Smith', 'Charlie Brown', 'Diana Prince', 'Ethan Hunt', 'Fiona Green', 'George Miller', 'Hannah Lee', 'Ivan Drago', 'Julia Roberts'];
    return names.map((name, i) => ({
        id: `acc-${i}`,
        accountId: `${10001 + i}`,
        name,
        balance: 5000 + Math.random() * 95000,
    }));
};

const generateMockRelations = (accounts: Account[]): CopyRelation[] => {
    return Array.from({ length: 8 }, (_, i) => {
        const master = accounts[Math.floor(Math.random() * 3)]; // First 3 are masters
        const follower = accounts[3 + Math.floor(Math.random() * 7)]; // Rest are followers

        const modes: CopyMode[] = ['proportional', 'fixed', 'inverse'];
        const copyMode = modes[i % 3];

        const statuses: CopyStatus[] = ['active', 'active', 'active', 'paused', 'stopped'];
        const status = statuses[i % statuses.length];

        const daysAgo = Math.floor(Math.random() * 90);
        const createdAt = new Date(Date.now() - daysAgo * 24 * 60 * 60 * 1000).toISOString();

        return {
            id: `rel-${i}`,
            masterId: master.id,
            masterName: master.name,
            masterAccountId: master.accountId,
            followerId: follower.id,
            followerName: follower.name,
            followerAccountId: follower.accountId,
            copyMode,
            lotMultiplier: copyMode === 'proportional' ? 0.5 + Math.random() * 1.5 : undefined,
            fixedLotSize: copyMode === 'fixed' ? 0.1 + Math.random() * 0.9 : undefined,
            maxLots: 5 + Math.floor(Math.random() * 10),
            skipSymbols: Math.random() > 0.5 ? ['BTCUSD'] : [],
            status,
            createdAt,
            tradesCopied: Math.floor(Math.random() * 200),
        };
    });
};

const generateMockLogs = (relations: CopyRelation[]): CopyLogEntry[] => {
    return Array.from({ length: 30 }, (_, i) => {
        const relation = relations[Math.floor(Math.random() * relations.length)];
        const hoursAgo = Math.floor(Math.random() * 48);
        const timestamp = new Date(Date.now() - hoursAgo * 60 * 60 * 1000).toISOString();

        const actions: CopyAction[] = ['open', 'close', 'modify'];
        const action = actions[Math.floor(Math.random() * actions.length)];

        const status: LogStatus = Math.random() > 0.15 ? 'success' : 'failed';

        return {
            id: `log-${i}`,
            relationId: relation.id,
            timestamp,
            masterTradeId: `MT${10000 + i}`,
            followerTradeId: status === 'success' ? `FT${20000 + i}` : null,
            action,
            symbol: AVAILABLE_SYMBOLS[Math.floor(Math.random() * AVAILABLE_SYMBOLS.length)],
            volume: 0.1 + Math.random() * 2,
            status,
            error: status === 'failed' ? 'Insufficient margin' : null,
        };
    });
};

const generateMockPerformance = (masterId: string, relations: CopyRelation[]): MasterPerformance => {
    const followersCount = relations.filter(r => r.masterId === masterId).length;
    return {
        accountId: masterId,
        winRate: 55 + Math.random() * 20,
        totalTrades: 100 + Math.floor(Math.random() * 400),
        pnl: (Math.random() - 0.3) * 50000,
        followersCount,
        equityCurve: Array.from({ length: 20 }, (_, i) => 10000 + Math.random() * 5000 + i * 200),
    };
};

export default function TradeCopierManagement() {
    const [accounts, setAccounts] = useState<Account[]>([]);
    const [relations, setRelations] = useState<CopyRelation[]>([]);
    const [logs, setLogs] = useState<CopyLogEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showEditModal, setShowEditModal] = useState(false);
    const [editingRelation, setEditingRelation] = useState<CopyRelation | null>(null);
    const [expandedRow, setExpandedRow] = useState<string | null>(null);
    const [selectedMaster, setSelectedMaster] = useState<string | null>(null);
    const [logFilter, setLogFilter] = useState<CopyAction | 'all'>('all');

    const [formData, setFormData] = useState<FormData>({
        masterId: '',
        followerId: '',
        copyMode: 'proportional',
        lotMultiplier: 1.0,
        fixedLotSize: 0.5,
        maxLots: 10,
        skipSymbols: [],
        isActive: true,
    });

    useEffect(() => {
        fetchData();
    }, []);

    const fetchData = async () => {
        try {
            setLoading(true);
            // In production: await api calls
            await new Promise(resolve => setTimeout(resolve, 500));
            const mockAccounts = generateMockAccounts();
            const mockRelations = generateMockRelations(mockAccounts);
            const mockLogs = generateMockLogs(mockRelations);

            setAccounts(mockAccounts);
            setRelations(mockRelations);
            setLogs(mockLogs);

            if (mockRelations.length > 0) {
                setSelectedMaster(mockRelations[0].masterId);
            }
        } catch (error) {
            console.error('[TradeCopier] Failed to fetch data:', error);
        } finally {
            setLoading(false);
        }
    };

    const stats = useMemo(() => {
        const totalRelations = relations.length;
        const activeRelations = relations.filter(r => r.status === 'active').length;
        const today = new Date().toDateString();
        const tradesCopiedToday = logs.filter(l => new Date(l.timestamp).toDateString() === today).length;
        const totalVolumeCopied = logs.reduce((sum, l) => sum + l.volume, 0);

        return { totalRelations, activeRelations, tradesCopiedToday, totalVolumeCopied };
    }, [relations, logs]);

    const masterPerformance = useMemo(() => {
        if (!selectedMaster) return null;
        return generateMockPerformance(selectedMaster, relations);
    }, [selectedMaster, relations]);

    const filteredRelations = useMemo(() => {
        if (!searchQuery.trim()) return relations;
        const query = searchQuery.toLowerCase();
        return relations.filter(r =>
            r.masterName.toLowerCase().includes(query) ||
            r.followerName.toLowerCase().includes(query) ||
            r.masterAccountId.includes(query) ||
            r.followerAccountId.includes(query)
        );
    }, [relations, searchQuery]);

    const filteredLogs = useMemo(() => {
        let filtered = logs.filter(l => l.relationId === expandedRow);
        if (logFilter !== 'all') {
            filtered = filtered.filter(l => l.action === logFilter);
        }
        return filtered;
    }, [logs, expandedRow, logFilter]);

    const handleCreate = () => {
        setFormData({
            masterId: '',
            followerId: '',
            copyMode: 'proportional',
            lotMultiplier: 1.0,
            fixedLotSize: 0.5,
            maxLots: 10,
            skipSymbols: [],
            isActive: true,
        });
        setShowCreateModal(true);
    };

    const handleEdit = (relation: CopyRelation) => {
        setEditingRelation(relation);
        setFormData({
            masterId: relation.masterId,
            followerId: relation.followerId,
            copyMode: relation.copyMode,
            lotMultiplier: relation.lotMultiplier || 1.0,
            fixedLotSize: relation.fixedLotSize || 0.5,
            maxLots: relation.maxLots,
            skipSymbols: relation.skipSymbols,
            isActive: relation.status === 'active',
        });
        setShowEditModal(true);
    };

    const handleToggleSymbol = (symbol: string) => {
        setFormData(prev => ({
            ...prev,
            skipSymbols: prev.skipSymbols.includes(symbol)
                ? prev.skipSymbols.filter(s => s !== symbol)
                : [...prev.skipSymbols, symbol],
        }));
    };

    const submitCreate = async () => {
        if (!formData.masterId || !formData.followerId) {
            alert('Please select both master and follower accounts');
            return;
        }

        try {
            // In production: await api.post(API_CONFIG.TRADE_COPIER_CREATE, formData);
            setShowCreateModal(false);
            await fetchData();
        } catch (error) {
            console.error('[TradeCopier] Failed to create relation:', error);
        }
    };

    const submitEdit = async () => {
        if (!editingRelation) return;
        try {
            // In production: await api.put(API_CONFIG.TRADE_COPIER_UPDATE(editingRelation.id), formData);
            setShowEditModal(false);
            await fetchData();
        } catch (error) {
            console.error('[TradeCopier] Failed to update relation:', error);
        }
    };

    const handleDelete = async (relation: CopyRelation) => {
        if (!confirm(`Delete copy relation from ${relation.masterName} to ${relation.followerName}?`)) return;
        try {
            // In production: await api.delete(API_CONFIG.TRADE_COPIER_DELETE(relation.id));
            await fetchData();
        } catch (error) {
            console.error('[TradeCopier] Failed to delete relation:', error);
        }
    };

    const handleTogglePause = async (relation: CopyRelation) => {
        try {
            // In production: await api.put(API_CONFIG.TRADE_COPIER_TOGGLE(relation.id));
            setRelations(prev =>
                prev.map(r =>
                    r.id === relation.id
                        ? { ...r, status: r.status === 'active' ? 'paused' : 'active' }
                        : r
                )
            );
        } catch (error) {
            console.error('[TradeCopier] Failed to toggle relation:', error);
        }
    };

    const formatRelativeTime = (dateString: string) => {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffMins = Math.floor(diffMs / (1000 * 60));
        const diffHours = Math.floor(diffMs / (1000 * 60 * 60));

        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins}m ago`;
        if (diffHours < 24) return `${diffHours}h ago`;
        return date.toLocaleDateString();
    };

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0 }).format(value);
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full bg-[#121316]">
                <div className="text-zinc-400">Loading trade copier...</div>
            </div>
        );
    }

    return (
        <div className="flex h-full bg-[#121316] text-white overflow-hidden">
            {/* Main Content */}
            <div className="flex-1 flex flex-col overflow-hidden">
                {/* Header */}
                <div className="flex-shrink-0 bg-zinc-900 border-b border-zinc-800 p-4">
                    <div className="flex items-center justify-between mb-4">
                        <div className="flex items-center gap-3">
                            <Copy className="text-blue-400" size={24} />
                            <div>
                                <h1 className="text-xl font-bold text-white">Trade Copier Management</h1>
                                <p className="text-xs text-zinc-400">{stats.totalRelations} copy relations</p>
                            </div>
                        </div>
                        <button
                            onClick={handleCreate}
                            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm font-medium transition-colors"
                        >
                            <Plus size={16} />
                            Add Copy Relation
                        </button>
                    </div>

                    {/* Stats Cards */}
                    <div className="grid grid-cols-4 gap-4 mb-4">
                        <div className="bg-zinc-800 rounded-lg p-3 border border-zinc-700">
                            <div className="text-xs text-zinc-400 mb-1">Total Relations</div>
                            <div className="text-2xl font-bold text-white">{stats.totalRelations}</div>
                        </div>
                        <div className="bg-zinc-800 rounded-lg p-3 border border-zinc-700">
                            <div className="text-xs text-zinc-400 mb-1">Active Relations</div>
                            <div className="text-2xl font-bold text-green-400">{stats.activeRelations}</div>
                        </div>
                        <div className="bg-zinc-800 rounded-lg p-3 border border-zinc-700">
                            <div className="text-xs text-zinc-400 mb-1">Copied Today</div>
                            <div className="text-2xl font-bold text-blue-400">{stats.tradesCopiedToday}</div>
                        </div>
                        <div className="bg-zinc-800 rounded-lg p-3 border border-zinc-700">
                            <div className="text-xs text-zinc-400 mb-1">Total Volume</div>
                            <div className="text-2xl font-bold text-purple-400">{stats.totalVolumeCopied.toFixed(1)}</div>
                        </div>
                    </div>

                    {/* Search */}
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={16} />
                        <input
                            type="text"
                            placeholder="Search by master or follower name/account..."
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
                </div>

                {/* Table */}
                <div className="flex-1 overflow-auto custom-scrollbar">
                    <table className="w-full text-xs">
                        <thead className="sticky top-0 bg-zinc-900 border-b border-zinc-800 z-10">
                            <tr>
                                <th className="text-left px-3 py-2 font-semibold text-zinc-400 w-8"></th>
                                <th className="text-left px-3 py-2 font-semibold text-zinc-400">ID</th>
                                <th className="text-left px-3 py-2 font-semibold text-zinc-400">Master Account</th>
                                <th className="text-left px-3 py-2 font-semibold text-zinc-400">Follower Account</th>
                                <th className="text-left px-3 py-2 font-semibold text-zinc-400">Copy Mode</th>
                                <th className="text-left px-3 py-2 font-semibold text-zinc-400">Lot Config</th>
                                <th className="text-left px-3 py-2 font-semibold text-zinc-400">Status</th>
                                <th className="text-left px-3 py-2 font-semibold text-zinc-400">Created</th>
                                <th className="text-left px-3 py-2 font-semibold text-zinc-400">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredRelations.map((relation) => (
                                <React.Fragment key={relation.id}>
                                    <tr className="border-b border-zinc-800 hover:bg-zinc-900/50">
                                        <td
                                            className="px-3 py-2 cursor-pointer"
                                            onClick={() => setExpandedRow(expandedRow === relation.id ? null : relation.id)}
                                        >
                                            {expandedRow === relation.id ? (
                                                <ChevronDown size={14} className="text-zinc-500" />
                                            ) : (
                                                <ChevronRight size={14} className="text-zinc-500" />
                                            )}
                                        </td>
                                        <td className="px-3 py-2 font-mono text-blue-400">{relation.id}</td>
                                        <td className="px-3 py-2">
                                            <div className="font-medium text-white">{relation.masterName}</div>
                                            <div className="text-zinc-400 text-xs">#{relation.masterAccountId}</div>
                                        </td>
                                        <td className="px-3 py-2">
                                            <div className="font-medium text-white">{relation.followerName}</div>
                                            <div className="text-zinc-400 text-xs">#{relation.followerAccountId}</div>
                                        </td>
                                        <td className="px-3 py-2">
                                            <span className={`px-2 py-0.5 rounded text-xs font-medium ${COPY_MODE_COLORS[relation.copyMode]}`}>
                                                {relation.copyMode.charAt(0).toUpperCase() + relation.copyMode.slice(1)}
                                            </span>
                                        </td>
                                        <td className="px-3 py-2 text-zinc-400">
                                            {relation.copyMode === 'proportional' && `${relation.lotMultiplier}x`}
                                            {relation.copyMode === 'fixed' && `${relation.fixedLotSize} lots`}
                                            {relation.copyMode === 'inverse' && `${relation.lotMultiplier}x inv`}
                                            <div className="text-xs">Max: {relation.maxLots}</div>
                                        </td>
                                        <td className="px-3 py-2">
                                            <span className={`px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[relation.status]}`}>
                                                {relation.status.charAt(0).toUpperCase() + relation.status.slice(1)}
                                            </span>
                                        </td>
                                        <td className="px-3 py-2 text-zinc-400">
                                            {new Date(relation.createdAt).toLocaleDateString()}
                                        </td>
                                        <td className="px-3 py-2">
                                            <div className="flex items-center gap-1">
                                                <button
                                                    onClick={() => handleEdit(relation)}
                                                    className="p-1 hover:bg-blue-600/20 text-blue-400 rounded transition-colors"
                                                    title="Edit"
                                                >
                                                    <Edit size={14} />
                                                </button>
                                                {relation.status !== 'stopped' && (
                                                    <button
                                                        onClick={() => handleTogglePause(relation)}
                                                        className={`p-1 rounded transition-colors ${
                                                            relation.status === 'active'
                                                                ? 'hover:bg-yellow-600/20 text-yellow-400'
                                                                : 'hover:bg-green-600/20 text-green-400'
                                                        }`}
                                                        title={relation.status === 'active' ? 'Pause' : 'Resume'}
                                                    >
                                                        {relation.status === 'active' ? <Pause size={14} /> : <Play size={14} />}
                                                    </button>
                                                )}
                                                <button
                                                    onClick={() => handleDelete(relation)}
                                                    className="p-1 hover:bg-red-600/20 text-red-400 rounded transition-colors"
                                                    title="Delete"
                                                >
                                                    <Trash2 size={14} />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>

                                    {/* Copy Log (Expanded) */}
                                    {expandedRow === relation.id && (
                                        <tr className="bg-zinc-900/70 border-b border-zinc-800">
                                            <td colSpan={9} className="px-6 py-4">
                                                <div>
                                                    <div className="flex items-center justify-between mb-3">
                                                        <h3 className="text-sm font-bold text-white">Copy Log</h3>
                                                        <select
                                                            value={logFilter}
                                                            onChange={(e) => setLogFilter(e.target.value as any)}
                                                            className="px-3 py-1 bg-zinc-800 border border-zinc-700 rounded text-xs text-white"
                                                        >
                                                            <option value="all">All Actions</option>
                                                            <option value="open">Open Only</option>
                                                            <option value="close">Close Only</option>
                                                            <option value="modify">Modify Only</option>
                                                        </select>
                                                    </div>
                                                    <div className="bg-zinc-800 rounded overflow-hidden max-h-80 overflow-y-auto custom-scrollbar">
                                                        <table className="w-full text-xs">
                                                            <thead className="bg-zinc-900 sticky top-0">
                                                                <tr>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Time</th>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Master Trade</th>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Follower Trade</th>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Action</th>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Symbol</th>
                                                                    <th className="text-right px-3 py-2 text-zinc-400 font-semibold">Volume</th>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Status</th>
                                                                    <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Error</th>
                                                                </tr>
                                                            </thead>
                                                            <tbody>
                                                                {filteredLogs.map(log => (
                                                                    <tr key={log.id} className="border-t border-zinc-700 hover:bg-zinc-700/50">
                                                                        <td className="px-3 py-2 text-zinc-400">{formatRelativeTime(log.timestamp)}</td>
                                                                        <td className="px-3 py-2 font-mono text-blue-400">{log.masterTradeId}</td>
                                                                        <td className="px-3 py-2 font-mono text-purple-400">{log.followerTradeId || '-'}</td>
                                                                        <td className="px-3 py-2">
                                                                            <span className={`px-2 py-0.5 rounded text-xs ${
                                                                                log.action === 'open' ? 'bg-green-600 text-white' :
                                                                                log.action === 'close' ? 'bg-red-600 text-white' :
                                                                                'bg-yellow-600 text-black'
                                                                            }`}>
                                                                                {log.action}
                                                                            </span>
                                                                        </td>
                                                                        <td className="px-3 py-2 font-mono text-white">{log.symbol}</td>
                                                                        <td className="px-3 py-2 text-right text-white">{log.volume.toFixed(2)}</td>
                                                                        <td className="px-3 py-2">
                                                                            {log.status === 'success' ? (
                                                                                <span className="flex items-center gap-1 text-green-400">
                                                                                    <CheckCircle size={12} /> Success
                                                                                </span>
                                                                            ) : (
                                                                                <span className="flex items-center gap-1 text-red-400">
                                                                                    <XCircle size={12} /> Failed
                                                                                </span>
                                                                            )}
                                                                        </td>
                                                                        <td className="px-3 py-2 text-red-400 text-xs">{log.error || '-'}</td>
                                                                    </tr>
                                                                ))}
                                                            </tbody>
                                                        </table>
                                                        {filteredLogs.length === 0 && (
                                                            <div className="text-center py-8 text-zinc-500">No log entries</div>
                                                        )}
                                                    </div>
                                                </div>
                                            </td>
                                        </tr>
                                    )}
                                </React.Fragment>
                            ))}
                        </tbody>
                    </table>

                    {filteredRelations.length === 0 && (
                        <div className="text-center py-12 text-zinc-500">No copy relations found</div>
                    )}
                </div>
            </div>

            {/* Sidebar - Master Performance */}
            {masterPerformance && (
                <div className="w-80 bg-zinc-900 border-l border-zinc-800 flex flex-col overflow-hidden">
                    <div className="p-4 border-b border-zinc-800">
                        <h2 className="text-sm font-bold text-white mb-3">Master Performance</h2>
                        <select
                            value={selectedMaster || ''}
                            onChange={(e) => setSelectedMaster(e.target.value)}
                            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white"
                        >
                            {Array.from(new Set(relations.map(r => r.masterId))).map(masterId => {
                                const rel = relations.find(r => r.masterId === masterId);
                                return (
                                    <option key={masterId} value={masterId}>
                                        {rel?.masterName} (#{rel?.masterAccountId})
                                    </option>
                                );
                            })}
                        </select>
                    </div>

                    <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-4">
                        <div className="grid grid-cols-2 gap-3">
                            <div className="bg-zinc-800 rounded p-3">
                                <div className="text-xs text-zinc-400 mb-1">Win Rate</div>
                                <div className="text-xl font-bold text-green-400">{masterPerformance.winRate.toFixed(1)}%</div>
                            </div>
                            <div className="bg-zinc-800 rounded p-3">
                                <div className="text-xs text-zinc-400 mb-1">Total Trades</div>
                                <div className="text-xl font-bold text-white">{masterPerformance.totalTrades}</div>
                            </div>
                            <div className="bg-zinc-800 rounded p-3">
                                <div className="text-xs text-zinc-400 mb-1">P&L</div>
                                <div className={`text-xl font-bold ${masterPerformance.pnl >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                                    {formatCurrency(masterPerformance.pnl)}
                                </div>
                            </div>
                            <div className="bg-zinc-800 rounded p-3">
                                <div className="text-xs text-zinc-400 mb-1">Followers</div>
                                <div className="text-xl font-bold text-blue-400">{masterPerformance.followersCount}</div>
                            </div>
                        </div>

                        <div className="bg-zinc-800 rounded p-3">
                            <div className="text-xs text-zinc-400 mb-2">Equity Curve</div>
                            <svg width="100%" height="100" className="block">
                                <polyline
                                    points={masterPerformance.equityCurve.map((val, i) => {
                                        const x = (i / (masterPerformance.equityCurve.length - 1)) * 100;
                                        const min = Math.min(...masterPerformance.equityCurve);
                                        const max = Math.max(...masterPerformance.equityCurve);
                                        const range = max - min || 1;
                                        const y = 100 - ((val - min) / range) * 100;
                                        return `${x}%,${y}%`;
                                    }).join(' ')}
                                    fill="none"
                                    stroke="#10B981"
                                    strokeWidth="2"
                                    vectorEffect="non-scaling-stroke"
                                />
                            </svg>
                        </div>
                    </div>
                </div>
            )}

            {/* Create/Edit Modal */}
            {(showCreateModal || showEditModal) && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 overflow-y-auto">
                    <div className="bg-zinc-900 rounded-lg shadow-xl border border-zinc-700 w-full max-w-2xl my-8">
                        <div className="flex items-center justify-between p-4 border-b border-zinc-800">
                            <h2 className="text-lg font-bold text-white">
                                {showCreateModal ? 'Create Copy Relation' : 'Edit Copy Relation'}
                            </h2>
                            <button
                                onClick={() => {
                                    setShowCreateModal(false);
                                    setShowEditModal(false);
                                }}
                                className="text-zinc-400 hover:text-white transition-colors"
                            >
                                <X size={20} />
                            </button>
                        </div>

                        <div className="p-4 space-y-4 max-h-[calc(100vh-200px)] overflow-y-auto custom-scrollbar">
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-xs font-semibold text-zinc-400 mb-1">
                                        Master Account <span className="text-red-400">*</span>
                                    </label>
                                    <select
                                        value={formData.masterId}
                                        onChange={(e) => setFormData(prev => ({ ...prev, masterId: e.target.value }))}
                                        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white"
                                    >
                                        <option value="">Select Master</option>
                                        {accounts.map(acc => (
                                            <option key={acc.id} value={acc.id}>
                                                {acc.name} (#{acc.accountId})
                                            </option>
                                        ))}
                                    </select>
                                </div>

                                <div>
                                    <label className="block text-xs font-semibold text-zinc-400 mb-1">
                                        Follower Account <span className="text-red-400">*</span>
                                    </label>
                                    <select
                                        value={formData.followerId}
                                        onChange={(e) => setFormData(prev => ({ ...prev, followerId: e.target.value }))}
                                        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white"
                                    >
                                        <option value="">Select Follower</option>
                                        {accounts.filter(a => a.id !== formData.masterId).map(acc => (
                                            <option key={acc.id} value={acc.id}>
                                                {acc.name} (#{acc.accountId})
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            </div>

                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-2">
                                    Copy Mode <span className="text-red-400">*</span>
                                </label>
                                <div className="grid grid-cols-3 gap-2">
                                    {(['proportional', 'fixed', 'inverse'] as CopyMode[]).map(mode => (
                                        <label
                                            key={mode}
                                            className={`flex items-center justify-center gap-2 px-3 py-2 rounded cursor-pointer transition-colors ${
                                                formData.copyMode === mode
                                                    ? COPY_MODE_COLORS[mode]
                                                    : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
                                            }`}
                                        >
                                            <input
                                                type="radio"
                                                name="copyMode"
                                                checked={formData.copyMode === mode}
                                                onChange={() => setFormData(prev => ({ ...prev, copyMode: mode }))}
                                                className="hidden"
                                            />
                                            <span className="text-xs font-medium capitalize">{mode}</span>
                                        </label>
                                    ))}
                                </div>
                            </div>

                            {formData.copyMode === 'proportional' && (
                                <div>
                                    <label className="block text-xs font-semibold text-zinc-400 mb-1">
                                        Lot Multiplier
                                    </label>
                                    <input
                                        type="number"
                                        step="0.1"
                                        min="0.1"
                                        value={formData.lotMultiplier}
                                        onChange={(e) => setFormData(prev => ({ ...prev, lotMultiplier: parseFloat(e.target.value) }))}
                                        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white"
                                    />
                                    <div className="text-xs text-zinc-500 mt-1">
                                        Follower lot = Master lot × {formData.lotMultiplier}x
                                    </div>
                                </div>
                            )}

                            {formData.copyMode === 'fixed' && (
                                <div>
                                    <label className="block text-xs font-semibold text-zinc-400 mb-1">
                                        Fixed Lot Size
                                    </label>
                                    <input
                                        type="number"
                                        step="0.01"
                                        min="0.01"
                                        value={formData.fixedLotSize}
                                        onChange={(e) => setFormData(prev => ({ ...prev, fixedLotSize: parseFloat(e.target.value) }))}
                                        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white"
                                    />
                                    <div className="text-xs text-zinc-500 mt-1">
                                        All trades copied with fixed {formData.fixedLotSize} lots
                                    </div>
                                </div>
                            )}

                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-1">
                                    Max Lots (Safety Limit)
                                </label>
                                <input
                                    type="number"
                                    step="1"
                                    min="1"
                                    value={formData.maxLots}
                                    onChange={(e) => setFormData(prev => ({ ...prev, maxLots: parseInt(e.target.value) }))}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white"
                                />
                            </div>

                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-2">
                                    Skip Symbols (Optional)
                                </label>
                                <div className="grid grid-cols-3 gap-2">
                                    {AVAILABLE_SYMBOLS.map(symbol => (
                                        <label
                                            key={symbol}
                                            className="flex items-center gap-2 px-3 py-2 bg-zinc-800 rounded cursor-pointer hover:bg-zinc-700"
                                        >
                                            <input
                                                type="checkbox"
                                                checked={formData.skipSymbols.includes(symbol)}
                                                onChange={() => handleToggleSymbol(symbol)}
                                                className="w-4 h-4 accent-blue-600"
                                            />
                                            <span className="text-xs text-white">{symbol}</span>
                                        </label>
                                    ))}
                                </div>
                            </div>

                            <div>
                                <label className="flex items-center gap-2 cursor-pointer">
                                    <input
                                        type="checkbox"
                                        checked={formData.isActive}
                                        onChange={(e) => setFormData(prev => ({ ...prev, isActive: e.target.checked }))}
                                        className="w-4 h-4 accent-blue-600"
                                    />
                                    <span className="text-sm text-white">Active (start copying immediately)</span>
                                </label>
                            </div>
                        </div>

                        <div className="flex items-center justify-end gap-3 p-4 border-t border-zinc-800">
                            <button
                                onClick={() => {
                                    setShowCreateModal(false);
                                    setShowEditModal(false);
                                }}
                                className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded text-sm font-medium transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={showCreateModal ? submitCreate : submitEdit}
                                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm font-medium transition-colors"
                            >
                                {showCreateModal ? 'Create' : 'Save'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
