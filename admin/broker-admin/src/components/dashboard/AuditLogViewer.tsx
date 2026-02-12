'use client';

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
    Shield,
    Search,
    Filter,
    Calendar,
    User,
    Activity,
    CheckCircle,
    XCircle,
    ChevronDown,
    ChevronRight,
    Clock,
    Info,
    Loader2,
    FileText,
    AlertTriangle,
    TrendingUp,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';
import { api } from '@/services/apiClient';

interface AuditEntry {
    id: number;
    adminId: number;
    adminName: string;
    action: string;
    entityType: string;
    entityId: number;
    changes: any;
    reason: string;
    ipAddress: string;
    userAgent: string;
    status: 'SUCCESS' | 'FAILED';
    errorMsg?: string;
    createdAt: string;
}

interface AuditStats {
    totalEntries: number;
    successCount: number;
    failedCount: number;
    actionCounts: { [action: string]: number };
    adminCounts: { [adminName: string]: number };
}

interface AuditFilters {
    admin: string;
    action: string;
    entityType: string;
    status: string;
    fromDate: string;
    toDate: string;
    search: string;
}

type SortField = 'createdAt' | 'adminName' | 'action' | 'entityType' | 'status';
type SortDirection = 'asc' | 'desc';

const STATUS_COLORS = {
    SUCCESS: 'text-green-400 bg-green-500/10 border-green-500/30',
    FAILED: 'text-red-400 bg-red-500/10 border-red-500/30',
};

const ACTION_TYPES = [
    'USER_CREATE', 'USER_UPDATE', 'USER_DELETE', 'USER_DISABLE', 'USER_ENABLE',
    'FUND_DEPOSIT', 'FUND_WITHDRAW', 'FUND_ADJUSTMENT',
    'ORDER_CREATE', 'ORDER_MODIFY', 'ORDER_DELETE', 'ORDER_CLOSE',
    'POSITION_MODIFY', 'POSITION_CLOSE',
    'GROUP_CREATE', 'GROUP_UPDATE', 'GROUP_DELETE',
    'SYMBOL_CREATE', 'SYMBOL_UPDATE', 'SYMBOL_DELETE',
    'LOGIN', 'LOGOUT', '2FA_ENABLE', '2FA_DISABLE',
];

const ENTITY_TYPES = ['USER', 'ORDER', 'POSITION', 'GROUP', 'SYMBOL', 'ACCOUNT', 'ADMIN'];

export default function AuditLogViewer() {
    const [logs, setLogs] = useState<AuditEntry[]>([]);
    const [stats, setStats] = useState<AuditStats | null>(null);
    const [filters, setFilters] = useState<AuditFilters>({
        admin: '',
        action: '',
        entityType: '',
        status: '',
        fromDate: '',
        toDate: '',
        search: '',
    });
    const [sortField, setSortField] = useState<SortField>('createdAt');
    const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
    const [currentPage, setCurrentPage] = useState(1);
    const [expandedRow, setExpandedRow] = useState<number | null>(null);
    const [showFilters, setShowFilters] = useState(false);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const pageSize = 25;

    /**
     * Fetch audit logs and stats
     */
    const fetchAuditData = useCallback(async () => {
        try {
            setError(null);

            // Build query parameters for logs
            const params = new URLSearchParams();
            if (filters.admin) params.append('admin', filters.admin);
            if (filters.action) params.append('action', filters.action);
            if (filters.entityType) params.append('resource', filters.entityType);
            if (filters.status) params.append('success', filters.status === 'SUCCESS' ? 'true' : 'false');
            if (filters.fromDate) params.append('from', new Date(filters.fromDate).getTime().toString());
            if (filters.toDate) params.append('to', new Date(filters.toDate).getTime().toString());
            params.append('limit', '1000'); // Get large batch for client-side filtering

            // Fetch logs
            const logsUrl = `${API_CONFIG.AUDIT_LOGS}?${params.toString()}`;
            const response = await api.get<{ logs: AuditEntry[]; count: number; total: number }>(logsUrl);

            let fetchedLogs = response.logs || [];

            // Apply client-side search filter if needed
            if (filters.search) {
                const searchLower = filters.search.toLowerCase();
                fetchedLogs = fetchedLogs.filter(log =>
                    log.adminName.toLowerCase().includes(searchLower) ||
                    log.action.toLowerCase().includes(searchLower) ||
                    log.entityType.toLowerCase().includes(searchLower) ||
                    log.reason?.toLowerCase().includes(searchLower) ||
                    log.ipAddress.includes(searchLower)
                );
            }

            setLogs(fetchedLogs);

            // Fetch stats
            const statsData = await api.get<AuditStats>(API_CONFIG.AUDIT_STATS);
            setStats(statsData);
        } catch (err: any) {
            console.error('[AuditLogViewer] Failed to fetch data:', err);
            setError(err.message || 'Failed to load audit data');
        } finally {
            setLoading(false);
        }
    }, [filters]);

    /**
     * Initial load and filter change handler
     */
    useEffect(() => {
        fetchAuditData();
    }, [fetchAuditData]);

    /**
     * Sort logs
     */
    const sortedLogs = useMemo(() => {
        const sorted = [...logs].sort((a, b) => {
            let aVal: any = a[sortField];
            let bVal: any = b[sortField];

            // Handle date sorting
            if (sortField === 'createdAt') {
                aVal = new Date(aVal).getTime();
                bVal = new Date(bVal).getTime();
            }

            // Handle string sorting (case-insensitive)
            if (typeof aVal === 'string') {
                aVal = aVal.toLowerCase();
                bVal = bVal.toLowerCase();
            }

            if (sortDirection === 'asc') {
                return aVal > bVal ? 1 : aVal < bVal ? -1 : 0;
            } else {
                return aVal < bVal ? 1 : aVal > bVal ? -1 : 0;
            }
        });
        return sorted;
    }, [logs, sortField, sortDirection]);

    /**
     * Paginated logs
     */
    const paginatedLogs = useMemo(() => {
        const startIdx = (currentPage - 1) * pageSize;
        return sortedLogs.slice(startIdx, startIdx + pageSize);
    }, [sortedLogs, currentPage]);

    const totalPages = Math.ceil(sortedLogs.length / pageSize);

    /**
     * Handle sort column click
     */
    const handleSort = (field: SortField) => {
        if (sortField === field) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('desc');
        }
    };

    /**
     * Handle filter change
     */
    const handleFilterChange = (key: keyof AuditFilters, value: string) => {
        setFilters(prev => ({ ...prev, [key]: value }));
        setCurrentPage(1);
    };

    /**
     * Clear filters
     */
    const handleClearFilters = () => {
        setFilters({
            admin: '',
            action: '',
            entityType: '',
            status: '',
            fromDate: '',
            toDate: '',
            search: '',
        });
        setCurrentPage(1);
    };

    /**
     * Toggle row expansion
     */
    const toggleRowExpansion = (id: number) => {
        setExpandedRow(expandedRow === id ? null : id);
    };

    /**
     * Format timestamp
     */
    const formatTimestamp = (timestamp: string): string => {
        const date = new Date(timestamp);
        return date.toLocaleString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
        });
    };

    /**
     * Calculate today's activity count
     */
    const todayActivityCount = useMemo(() => {
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        return logs.filter(log => new Date(log.createdAt) >= today).length;
    }, [logs]);

    /**
     * Calculate success rate
     */
    const successRate = useMemo(() => {
        if (!stats || stats.totalEntries === 0) return 0;
        return ((stats.successCount / stats.totalEntries) * 100).toFixed(1);
    }, [stats]);

    if (loading && !stats) {
        return (
            <div className="h-full flex items-center justify-center bg-[#0A0A0B]">
                <div className="text-center">
                    <Loader2 className="w-8 h-8 animate-spin text-blue-500 mx-auto mb-3" />
                    <p className="text-zinc-400">Loading audit logs...</p>
                </div>
            </div>
        );
    }

    if (error && !stats) {
        return (
            <div className="h-full flex items-center justify-center bg-[#0A0A0B]">
                <div className="text-center">
                    <AlertTriangle className="w-12 h-12 text-red-400 mx-auto mb-3" />
                    <p className="text-red-300">{error}</p>
                    <button
                        onClick={fetchAuditData}
                        className="mt-4 px-4 py-2 bg-zinc-800 text-white rounded hover:bg-zinc-700"
                    >
                        Retry
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="h-full bg-[#0A0A0B] flex flex-col overflow-hidden">
            {/* Header */}
            <div className="flex-shrink-0 p-4 border-b border-zinc-800">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <Shield className="w-6 h-6 text-blue-400" />
                        <h1 className="text-2xl font-bold text-white">Audit Log Viewer</h1>
                    </div>
                    <button
                        onClick={() => setShowFilters(!showFilters)}
                        className={`flex items-center gap-2 px-4 py-2 rounded-lg border transition-colors ${
                            showFilters
                                ? 'bg-blue-500/20 border-blue-500 text-blue-400'
                                : 'bg-zinc-800 border-zinc-700 text-zinc-300 hover:bg-zinc-700'
                        }`}
                    >
                        <Filter className="w-4 h-4" />
                        <span className="text-sm font-medium">Filters</span>
                        {showFilters ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                    </button>
                </div>

                {/* Summary Cards */}
                <div className="grid grid-cols-4 gap-4">
                    {/* Total Entries */}
                    <div className="bg-zinc-900 rounded-lg p-3 border border-zinc-800">
                        <div className="flex items-center gap-2 mb-1">
                            <FileText className="w-4 h-4 text-blue-400" />
                            <span className="text-xs text-zinc-400">Total Entries</span>
                        </div>
                        <div className="text-xl font-bold text-white">
                            {stats ? stats.totalEntries.toLocaleString() : '-'}
                        </div>
                    </div>

                    {/* Success Rate */}
                    <div className="bg-zinc-900 rounded-lg p-3 border border-zinc-800">
                        <div className="flex items-center gap-2 mb-1">
                            <TrendingUp className="w-4 h-4 text-green-400" />
                            <span className="text-xs text-zinc-400">Success Rate</span>
                        </div>
                        <div className="text-xl font-bold text-green-400">
                            {successRate}%
                        </div>
                    </div>

                    {/* Failed Actions */}
                    <div className="bg-zinc-900 rounded-lg p-3 border border-zinc-800">
                        <div className="flex items-center gap-2 mb-1">
                            <XCircle className="w-4 h-4 text-red-400" />
                            <span className="text-xs text-zinc-400">Failed Actions</span>
                        </div>
                        <div className="text-xl font-bold text-red-400">
                            {stats ? stats.failedCount.toLocaleString() : '-'}
                        </div>
                    </div>

                    {/* Today's Activity */}
                    <div className="bg-zinc-900 rounded-lg p-3 border border-zinc-800">
                        <div className="flex items-center gap-2 mb-1">
                            <Activity className="w-4 h-4 text-purple-400" />
                            <span className="text-xs text-zinc-400">Today's Activity</span>
                        </div>
                        <div className="text-xl font-bold text-white">
                            {todayActivityCount.toLocaleString()}
                        </div>
                    </div>
                </div>

                {/* Filter Bar */}
                {showFilters && (
                    <div className="mt-4 p-4 bg-zinc-900 rounded-lg border border-zinc-800">
                        <div className="grid grid-cols-4 gap-3 mb-3">
                            {/* Search */}
                            <div>
                                <label className="block text-xs text-zinc-400 mb-1">Search</label>
                                <div className="relative">
                                    <Search className="absolute left-2 top-2.5 w-4 h-4 text-zinc-500" />
                                    <input
                                        type="text"
                                        value={filters.search}
                                        onChange={(e) => handleFilterChange('search', e.target.value)}
                                        placeholder="Search logs..."
                                        className="w-full pl-8 pr-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                                    />
                                </div>
                            </div>

                            {/* Admin */}
                            <div>
                                <label className="block text-xs text-zinc-400 mb-1">Admin</label>
                                <div className="relative">
                                    <User className="absolute left-2 top-2.5 w-4 h-4 text-zinc-500" />
                                    <input
                                        type="text"
                                        value={filters.admin}
                                        onChange={(e) => handleFilterChange('admin', e.target.value)}
                                        placeholder="Filter by admin..."
                                        className="w-full pl-8 pr-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                                    />
                                </div>
                            </div>

                            {/* Action */}
                            <div>
                                <label className="block text-xs text-zinc-400 mb-1">Action</label>
                                <select
                                    value={filters.action}
                                    onChange={(e) => handleFilterChange('action', e.target.value)}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-blue-500"
                                >
                                    <option value="">All Actions</option>
                                    {ACTION_TYPES.map(action => (
                                        <option key={action} value={action}>{action}</option>
                                    ))}
                                </select>
                            </div>

                            {/* Entity Type */}
                            <div>
                                <label className="block text-xs text-zinc-400 mb-1">Entity Type</label>
                                <select
                                    value={filters.entityType}
                                    onChange={(e) => handleFilterChange('entityType', e.target.value)}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-blue-500"
                                >
                                    <option value="">All Types</option>
                                    {ENTITY_TYPES.map(type => (
                                        <option key={type} value={type}>{type}</option>
                                    ))}
                                </select>
                            </div>
                        </div>

                        <div className="grid grid-cols-4 gap-3">
                            {/* Status */}
                            <div>
                                <label className="block text-xs text-zinc-400 mb-1">Status</label>
                                <select
                                    value={filters.status}
                                    onChange={(e) => handleFilterChange('status', e.target.value)}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-blue-500"
                                >
                                    <option value="">All Statuses</option>
                                    <option value="SUCCESS">Success</option>
                                    <option value="FAILED">Failed</option>
                                </select>
                            </div>

                            {/* From Date */}
                            <div>
                                <label className="block text-xs text-zinc-400 mb-1">From Date</label>
                                <div className="relative">
                                    <Calendar className="absolute left-2 top-2.5 w-4 h-4 text-zinc-500" />
                                    <input
                                        type="date"
                                        value={filters.fromDate}
                                        onChange={(e) => handleFilterChange('fromDate', e.target.value)}
                                        className="w-full pl-8 pr-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-blue-500"
                                    />
                                </div>
                            </div>

                            {/* To Date */}
                            <div>
                                <label className="block text-xs text-zinc-400 mb-1">To Date</label>
                                <div className="relative">
                                    <Calendar className="absolute left-2 top-2.5 w-4 h-4 text-zinc-500" />
                                    <input
                                        type="date"
                                        value={filters.toDate}
                                        onChange={(e) => handleFilterChange('toDate', e.target.value)}
                                        className="w-full pl-8 pr-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-blue-500"
                                    />
                                </div>
                            </div>

                            {/* Clear Filters */}
                            <div className="flex items-end">
                                <button
                                    onClick={handleClearFilters}
                                    className="w-full px-4 py-2 bg-zinc-800 border border-zinc-700 text-zinc-300 rounded text-sm hover:bg-zinc-700"
                                >
                                    Clear Filters
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* Audit Log Table */}
            <div className="flex-1 overflow-auto">
                <table className="w-full text-sm">
                    <thead className="bg-zinc-800 sticky top-0 z-10">
                        <tr>
                            <th className="w-10 px-3 py-3 text-left text-xs font-semibold text-zinc-400"></th>
                            <th
                                onClick={() => handleSort('createdAt')}
                                className="px-3 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                            >
                                <div className="flex items-center gap-1">
                                    <Clock className="w-3 h-3" />
                                    Timestamp
                                    {sortField === 'createdAt' && (
                                        <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                    )}
                                </div>
                            </th>
                            <th
                                onClick={() => handleSort('adminName')}
                                className="px-3 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                            >
                                <div className="flex items-center gap-1">
                                    <User className="w-3 h-3" />
                                    Admin
                                    {sortField === 'adminName' && (
                                        <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                    )}
                                </div>
                            </th>
                            <th
                                onClick={() => handleSort('action')}
                                className="px-3 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                            >
                                <div className="flex items-center gap-1">
                                    <Activity className="w-3 h-3" />
                                    Action
                                    {sortField === 'action' && (
                                        <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                    )}
                                </div>
                            </th>
                            <th
                                onClick={() => handleSort('entityType')}
                                className="px-3 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                            >
                                <div className="flex items-center gap-1">
                                    <FileText className="w-3 h-3" />
                                    Entity Type
                                    {sortField === 'entityType' && (
                                        <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                    )}
                                </div>
                            </th>
                            <th className="px-3 py-3 text-left text-xs font-semibold text-zinc-400">Entity ID</th>
                            <th className="px-3 py-3 text-left text-xs font-semibold text-zinc-400">IP Address</th>
                            <th
                                onClick={() => handleSort('status')}
                                className="px-3 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                            >
                                <div className="flex items-center gap-1">
                                    Status
                                    {sortField === 'status' && (
                                        <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                    )}
                                </div>
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        {paginatedLogs.length === 0 ? (
                            <tr>
                                <td colSpan={8} className="px-3 py-12 text-center text-zinc-500">
                                    <Info className="w-8 h-8 mx-auto mb-2 opacity-50" />
                                    No audit logs found
                                </td>
                            </tr>
                        ) : (
                            paginatedLogs.map((log) => (
                                <React.Fragment key={log.id}>
                                    {/* Main Row */}
                                    <tr
                                        onClick={() => toggleRowExpansion(log.id)}
                                        className="border-b border-zinc-800 hover:bg-zinc-800/50 cursor-pointer"
                                    >
                                        <td className="px-3 py-2">
                                            {expandedRow === log.id ? (
                                                <ChevronDown className="w-4 h-4 text-blue-400" />
                                            ) : (
                                                <ChevronRight className="w-4 h-4 text-zinc-600" />
                                            )}
                                        </td>
                                        <td className="px-3 py-2 text-zinc-400">{formatTimestamp(log.createdAt)}</td>
                                        <td className="px-3 py-2 text-white font-medium">{log.adminName}</td>
                                        <td className="px-3 py-2">
                                            <span className="px-2 py-1 bg-blue-500/10 text-blue-400 rounded text-xs font-medium">
                                                {log.action}
                                            </span>
                                        </td>
                                        <td className="px-3 py-2 text-zinc-300">{log.entityType}</td>
                                        <td className="px-3 py-2 text-zinc-400">#{log.entityId}</td>
                                        <td className="px-3 py-2 text-zinc-400 font-mono text-xs">{log.ipAddress}</td>
                                        <td className="px-3 py-2">
                                            <div className={`inline-flex items-center gap-1 px-2 py-1 rounded border text-xs font-semibold ${STATUS_COLORS[log.status]}`}>
                                                {log.status === 'SUCCESS' ? (
                                                    <CheckCircle className="w-3 h-3" />
                                                ) : (
                                                    <XCircle className="w-3 h-3" />
                                                )}
                                                {log.status}
                                            </div>
                                        </td>
                                    </tr>

                                    {/* Expanded Detail Panel */}
                                    {expandedRow === log.id && (
                                        <tr className="bg-zinc-900/50">
                                            <td colSpan={8} className="px-3 py-4">
                                                <div className="max-w-4xl grid grid-cols-2 gap-4">
                                                    {/* Left Column */}
                                                    <div className="space-y-3">
                                                        <div>
                                                            <div className="text-xs text-zinc-500 mb-1">Admin ID</div>
                                                            <div className="text-sm text-white">#{log.adminId}</div>
                                                        </div>
                                                        <div>
                                                            <div className="text-xs text-zinc-500 mb-1">Reason</div>
                                                            <div className="text-sm text-zinc-300">{log.reason || 'N/A'}</div>
                                                        </div>
                                                        <div>
                                                            <div className="text-xs text-zinc-500 mb-1">User Agent</div>
                                                            <div className="text-xs text-zinc-400 font-mono break-all">{log.userAgent}</div>
                                                        </div>
                                                        {log.errorMsg && (
                                                            <div>
                                                                <div className="text-xs text-zinc-500 mb-1">Error Message</div>
                                                                <div className="text-sm text-red-400">{log.errorMsg}</div>
                                                            </div>
                                                        )}
                                                    </div>

                                                    {/* Right Column - Changes */}
                                                    <div>
                                                        <div className="text-xs text-zinc-500 mb-1">Changes</div>
                                                        <pre className="text-xs text-zinc-300 bg-zinc-800 rounded p-3 overflow-auto max-h-40">
                                                            {JSON.stringify(log.changes, null, 2) || 'No changes recorded'}
                                                        </pre>
                                                    </div>
                                                </div>
                                            </td>
                                        </tr>
                                    )}
                                </React.Fragment>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {/* Pagination Footer */}
            {totalPages > 1 && (
                <div className="flex-shrink-0 flex items-center justify-between px-4 py-3 border-t border-zinc-800 bg-zinc-900">
                    <div className="text-sm text-zinc-400">
                        Showing {((currentPage - 1) * pageSize) + 1} - {Math.min(currentPage * pageSize, sortedLogs.length)} of {sortedLogs.length} entries
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                            disabled={currentPage === 1}
                            className="px-3 py-1 bg-zinc-800 text-zinc-300 rounded text-sm hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Previous
                        </button>
                        <div className="text-sm text-zinc-400">
                            Page {currentPage} of {totalPages}
                        </div>
                        <button
                            onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                            disabled={currentPage === totalPages}
                            className="px-3 py-1 bg-zinc-800 text-zinc-300 rounded text-sm hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Next
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}
