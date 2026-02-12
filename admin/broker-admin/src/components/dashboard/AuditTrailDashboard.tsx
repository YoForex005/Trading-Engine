'use client';

import React, { useState, useMemo } from 'react';
import {
    Calendar,
    User,
    Filter,
    Search,
    Download,
    ChevronDown,
    ChevronRight,
    Shield,
    Eye,
    AlertTriangle,
    Activity,
    Database,
    Settings,
    Users,
    TrendingUp,
    FileText,
    Lock,
    Unlock,
    Plus,
    Edit,
    Trash2,
    LogIn,
    Upload
} from 'lucide-react';

// Types
type ActionType = 'CREATE' | 'UPDATE' | 'DELETE' | 'LOGIN' | 'EXPORT';
type ModuleType = 'Accounts' | 'Symbols' | 'Orders' | 'Settings' | 'Users' | 'Groups' | 'LP' | 'Risk' | 'Reports' | 'System';

interface AuditEntry {
    id: string;
    timestamp: Date;
    adminName: string;
    adminId: string;
    actionType: ActionType;
    module: ModuleType;
    target: string;
    description: string;
    ipAddress: string;
    status: 'success' | 'failed' | 'blocked';
    details: {
        before?: any;
        after?: any;
        reason?: string;
        metadata?: Record<string, any>;
    };
}

interface AuditStats {
    totalToday: number;
    uniqueAdmins: number;
    mostActiveModule: string;
    failedActions: number;
}

// Color mapping for action types
const ACTION_COLORS: Record<ActionType, { bg: string; text: string; icon: React.ReactNode }> = {
    CREATE: { bg: 'bg-[#10B981]/10', text: 'text-[#10B981]', icon: <Plus size={14} /> },
    UPDATE: { bg: 'bg-[#3B82F6]/10', text: 'text-[#3B82F6]', icon: <Edit size={14} /> },
    DELETE: { bg: 'bg-[#EF4444]/10', text: 'text-[#EF4444]', icon: <Trash2 size={14} /> },
    LOGIN: { bg: 'bg-[#F59E0B]/10', text: 'text-[#F59E0B]', icon: <LogIn size={14} /> },
    EXPORT: { bg: 'bg-[#A855F7]/10', text: 'text-[#A855F7]', icon: <Upload size={14} /> },
};

// Module icons
const MODULE_ICONS: Record<ModuleType, React.ReactNode> = {
    Accounts: <Users size={14} />,
    Symbols: <TrendingUp size={14} />,
    Orders: <FileText size={14} />,
    Settings: <Settings size={14} />,
    Users: <User size={14} />,
    Groups: <Users size={14} />,
    LP: <Database size={14} />,
    Risk: <Shield size={14} />,
    Reports: <FileText size={14} />,
    System: <Activity size={14} />,
};

export default function AuditTrailDashboard() {
    const [entries, setEntries] = useState<AuditEntry[]>(() => generateMockEntries());
    const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
    const [filters, setFilters] = useState({
        dateFrom: '',
        dateTo: '',
        adminUser: '',
        actionType: '' as ActionType | '',
        module: '' as ModuleType | '',
        search: '',
    });

    // Calculate stats
    const stats: AuditStats = useMemo(() => {
        const today = new Date();
        today.setHours(0, 0, 0, 0);

        const todayEntries = entries.filter(e => e.timestamp >= today);
        const uniqueAdmins = new Set(entries.map(e => e.adminId)).size;

        const moduleCounts: Record<string, number> = {};
        entries.forEach(e => {
            moduleCounts[e.module] = (moduleCounts[e.module] || 0) + 1;
        });
        const mostActiveModule = Object.entries(moduleCounts).sort((a, b) => b[1] - a[1])[0]?.[0] || 'N/A';

        const failedActions = entries.filter(e => e.status === 'failed' || e.status === 'blocked').length;

        return {
            totalToday: todayEntries.length,
            uniqueAdmins,
            mostActiveModule,
            failedActions,
        };
    }, [entries]);

    // Filter entries
    const filteredEntries = useMemo(() => {
        return entries.filter(entry => {
            // Date range filter
            if (filters.dateFrom) {
                const fromDate = new Date(filters.dateFrom);
                if (entry.timestamp < fromDate) return false;
            }
            if (filters.dateTo) {
                const toDate = new Date(filters.dateTo);
                toDate.setHours(23, 59, 59, 999);
                if (entry.timestamp > toDate) return false;
            }

            // Admin user filter
            if (filters.adminUser && !entry.adminName.toLowerCase().includes(filters.adminUser.toLowerCase())) {
                return false;
            }

            // Action type filter
            if (filters.actionType && entry.actionType !== filters.actionType) {
                return false;
            }

            // Module filter
            if (filters.module && entry.module !== filters.module) {
                return false;
            }

            // Search filter
            if (filters.search) {
                const searchLower = filters.search.toLowerCase();
                return (
                    entry.description.toLowerCase().includes(searchLower) ||
                    entry.target.toLowerCase().includes(searchLower) ||
                    entry.adminName.toLowerCase().includes(searchLower)
                );
            }

            return true;
        });
    }, [entries, filters]);

    const toggleRow = (id: string) => {
        const newExpanded = new Set(expandedRows);
        if (newExpanded.has(id)) {
            newExpanded.delete(id);
        } else {
            newExpanded.add(id);
        }
        setExpandedRows(newExpanded);
    };

    const handleExportCSV = () => {
        const csvHeaders = ['Timestamp', 'Admin', 'Action Type', 'Module', 'Target', 'Description', 'IP Address', 'Status'];
        const csvRows = filteredEntries.map(entry => [
            entry.timestamp.toISOString(),
            entry.adminName,
            entry.actionType,
            entry.module,
            entry.target,
            entry.description,
            entry.ipAddress,
            entry.status,
        ]);

        const csvContent = [
            csvHeaders.join(','),
            ...csvRows.map(row => row.map(cell => `"${cell}"`).join(','))
        ].join('\n');

        const blob = new Blob([csvContent], { type: 'text/csv' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `audit-trail-${new Date().toISOString().split('T')[0]}.csv`;
        a.click();
        URL.revokeObjectURL(url);
    };

    return (
        <div className="h-full flex flex-col bg-[#121316] text-[#E5E7EB]">
            {/* Header */}
            <div className="flex-shrink-0 px-6 py-4 border-b border-[#383A42]">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <Shield className="text-[#F59E0B]" size={24} />
                        <h1 className="text-xl font-bold text-[#F5C542]">Audit Trail Dashboard</h1>
                    </div>
                    <button
                        onClick={handleExportCSV}
                        className="flex items-center gap-2 px-4 py-2 bg-[#1E2026] border border-[#383A42] rounded hover:bg-[#25272E] transition-colors text-sm"
                    >
                        <Download size={16} />
                        Export CSV
                    </button>
                </div>

                {/* Stats Cards */}
                <div className="grid grid-cols-4 gap-4">
                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Actions Today</span>
                            <Activity size={16} className="text-[#3B82F6]" />
                        </div>
                        <div className="text-2xl font-bold text-[#F5C542]">{stats.totalToday.toLocaleString()}</div>
                    </div>

                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Active Admins</span>
                            <Users size={16} className="text-[#10B981]" />
                        </div>
                        <div className="text-2xl font-bold text-[#F5C542]">{stats.uniqueAdmins}</div>
                    </div>

                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Most Active Module</span>
                            <TrendingUp size={16} className="text-[#A855F7]" />
                        </div>
                        <div className="text-lg font-bold text-[#F5C542]">{stats.mostActiveModule}</div>
                    </div>

                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Failed/Blocked</span>
                            <AlertTriangle size={16} className="text-[#EF4444]" />
                        </div>
                        <div className="text-2xl font-bold text-[#EF4444]">{stats.failedActions}</div>
                    </div>
                </div>
            </div>

            {/* Filter Bar */}
            <div className="flex-shrink-0 px-6 py-3 bg-[#1E2026] border-b border-[#383A42]">
                <div className="grid grid-cols-6 gap-3">
                    <div className="flex items-center gap-2 bg-[#121316] border border-[#383A42] rounded px-3 py-2">
                        <Calendar size={14} className="text-[#888]" />
                        <input
                            type="date"
                            value={filters.dateFrom}
                            onChange={(e) => setFilters({ ...filters, dateFrom: e.target.value })}
                            className="bg-transparent text-xs outline-none flex-1 text-[#E5E7EB]"
                            placeholder="From"
                        />
                    </div>

                    <div className="flex items-center gap-2 bg-[#121316] border border-[#383A42] rounded px-3 py-2">
                        <Calendar size={14} className="text-[#888]" />
                        <input
                            type="date"
                            value={filters.dateTo}
                            onChange={(e) => setFilters({ ...filters, dateTo: e.target.value })}
                            className="bg-transparent text-xs outline-none flex-1 text-[#E5E7EB]"
                            placeholder="To"
                        />
                    </div>

                    <div className="flex items-center gap-2 bg-[#121316] border border-[#383A42] rounded px-3 py-2">
                        <User size={14} className="text-[#888]" />
                        <input
                            type="text"
                            value={filters.adminUser}
                            onChange={(e) => setFilters({ ...filters, adminUser: e.target.value })}
                            className="bg-transparent text-xs outline-none flex-1 text-[#E5E7EB]"
                            placeholder="Admin user..."
                        />
                    </div>

                    <select
                        value={filters.actionType}
                        onChange={(e) => setFilters({ ...filters, actionType: e.target.value as ActionType | '' })}
                        className="bg-[#121316] border border-[#383A42] rounded px-3 py-2 text-xs outline-none text-[#E5E7EB]"
                    >
                        <option value="">All Action Types</option>
                        <option value="CREATE">CREATE</option>
                        <option value="UPDATE">UPDATE</option>
                        <option value="DELETE">DELETE</option>
                        <option value="LOGIN">LOGIN</option>
                        <option value="EXPORT">EXPORT</option>
                    </select>

                    <select
                        value={filters.module}
                        onChange={(e) => setFilters({ ...filters, module: e.target.value as ModuleType | '' })}
                        className="bg-[#121316] border border-[#383A42] rounded px-3 py-2 text-xs outline-none text-[#E5E7EB]"
                    >
                        <option value="">All Modules</option>
                        <option value="Accounts">Accounts</option>
                        <option value="Symbols">Symbols</option>
                        <option value="Orders">Orders</option>
                        <option value="Settings">Settings</option>
                        <option value="Users">Users</option>
                        <option value="Groups">Groups</option>
                        <option value="LP">LP</option>
                        <option value="Risk">Risk</option>
                        <option value="Reports">Reports</option>
                        <option value="System">System</option>
                    </select>

                    <div className="flex items-center gap-2 bg-[#121316] border border-[#383A42] rounded px-3 py-2">
                        <Search size={14} className="text-[#888]" />
                        <input
                            type="text"
                            value={filters.search}
                            onChange={(e) => setFilters({ ...filters, search: e.target.value })}
                            className="bg-transparent text-xs outline-none flex-1 text-[#E5E7EB]"
                            placeholder="Search actions..."
                        />
                    </div>
                </div>
            </div>

            {/* Timeline Table */}
            <div className="flex-1 overflow-auto">
                <table className="w-full text-xs">
                    <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                        <tr>
                            <th className="text-left py-2 px-4 font-semibold text-[#888] w-8"></th>
                            <th className="text-left py-2 px-4 font-semibold text-[#888]">Timestamp</th>
                            <th className="text-left py-2 px-4 font-semibold text-[#888]">Admin</th>
                            <th className="text-left py-2 px-4 font-semibold text-[#888]">Action</th>
                            <th className="text-left py-2 px-4 font-semibold text-[#888]">Module</th>
                            <th className="text-left py-2 px-4 font-semibold text-[#888]">Target</th>
                            <th className="text-left py-2 px-4 font-semibold text-[#888]">Description</th>
                            <th className="text-left py-2 px-4 font-semibold text-[#888]">IP Address</th>
                            <th className="text-left py-2 px-4 font-semibold text-[#888]">Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        {filteredEntries.map((entry) => {
                            const isExpanded = expandedRows.has(entry.id);
                            const actionColor = ACTION_COLORS[entry.actionType];

                            return (
                                <React.Fragment key={entry.id}>
                                    <tr
                                        className="border-b border-[#383A42] hover:bg-[#1E2026] cursor-pointer transition-colors"
                                        onClick={() => toggleRow(entry.id)}
                                    >
                                        <td className="py-2 px-4">
                                            {isExpanded ? (
                                                <ChevronDown size={14} className="text-[#888]" />
                                            ) : (
                                                <ChevronRight size={14} className="text-[#888]" />
                                            )}
                                        </td>
                                        <td className="py-2 px-4 text-[#888]">
                                            {entry.timestamp.toLocaleString('en-US', {
                                                month: 'short',
                                                day: '2-digit',
                                                hour: '2-digit',
                                                minute: '2-digit',
                                                second: '2-digit',
                                            })}
                                        </td>
                                        <td className="py-2 px-4">
                                            <div className="flex items-center gap-2">
                                                <User size={14} className="text-[#3B82F6]" />
                                                <span className="text-[#E5E7EB]">{entry.adminName}</span>
                                            </div>
                                        </td>
                                        <td className="py-2 px-4">
                                            <div className={`flex items-center gap-1 px-2 py-1 rounded ${actionColor.bg} ${actionColor.text} w-fit`}>
                                                {actionColor.icon}
                                                <span className="font-medium">{entry.actionType}</span>
                                            </div>
                                        </td>
                                        <td className="py-2 px-4">
                                            <div className="flex items-center gap-2 text-[#888]">
                                                {MODULE_ICONS[entry.module]}
                                                <span>{entry.module}</span>
                                            </div>
                                        </td>
                                        <td className="py-2 px-4 text-[#E5E7EB] font-mono">{entry.target}</td>
                                        <td className="py-2 px-4 text-[#888] max-w-xs truncate">{entry.description}</td>
                                        <td className="py-2 px-4 text-[#888] font-mono">{entry.ipAddress}</td>
                                        <td className="py-2 px-4">
                                            <span
                                                className={`px-2 py-1 rounded text-xs font-medium ${
                                                    entry.status === 'success'
                                                        ? 'bg-[#10B981]/10 text-[#10B981]'
                                                        : entry.status === 'failed'
                                                        ? 'bg-[#EF4444]/10 text-[#EF4444]'
                                                        : 'bg-[#F59E0B]/10 text-[#F59E0B]'
                                                }`}
                                            >
                                                {entry.status.toUpperCase()}
                                            </span>
                                        </td>
                                    </tr>

                                    {/* Expanded Details Row */}
                                    {isExpanded && (
                                        <tr className="bg-[#0D0E11] border-b border-[#383A42]">
                                            <td colSpan={9} className="py-4 px-8">
                                                <div className="grid grid-cols-2 gap-6">
                                                    {/* Left Column - Before/After */}
                                                    {(entry.details.before || entry.details.after) && (
                                                        <div>
                                                            <h4 className="text-sm font-semibold text-[#F5C542] mb-2">Changes</h4>
                                                            {entry.details.before && (
                                                                <div className="mb-3">
                                                                    <div className="text-xs text-[#888] mb-1">Before:</div>
                                                                    <pre className="bg-[#121316] border border-[#383A42] rounded p-2 text-xs text-[#E5E7EB] overflow-auto max-h-32">
                                                                        {JSON.stringify(entry.details.before, null, 2)}
                                                                    </pre>
                                                                </div>
                                                            )}
                                                            {entry.details.after && (
                                                                <div>
                                                                    <div className="text-xs text-[#888] mb-1">After:</div>
                                                                    <pre className="bg-[#121316] border border-[#383A42] rounded p-2 text-xs text-[#E5E7EB] overflow-auto max-h-32">
                                                                        {JSON.stringify(entry.details.after, null, 2)}
                                                                    </pre>
                                                                </div>
                                                            )}
                                                        </div>
                                                    )}

                                                    {/* Right Column - Metadata & Reason */}
                                                    <div>
                                                        {entry.details.reason && (
                                                            <div className="mb-3">
                                                                <h4 className="text-sm font-semibold text-[#F5C542] mb-2">Reason</h4>
                                                                <div className="bg-[#121316] border border-[#383A42] rounded p-2 text-xs text-[#E5E7EB]">
                                                                    {entry.details.reason}
                                                                </div>
                                                            </div>
                                                        )}
                                                        {entry.details.metadata && (
                                                            <div>
                                                                <h4 className="text-sm font-semibold text-[#F5C542] mb-2">Metadata</h4>
                                                                <div className="space-y-1">
                                                                    {Object.entries(entry.details.metadata).map(([key, value]) => (
                                                                        <div key={key} className="flex items-center gap-2 text-xs">
                                                                            <span className="text-[#888]">{key}:</span>
                                                                            <span className="text-[#E5E7EB] font-mono">{String(value)}</span>
                                                                        </div>
                                                                    ))}
                                                                </div>
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            </td>
                                        </tr>
                                    )}
                                </React.Fragment>
                            );
                        })}
                    </tbody>
                </table>

                {filteredEntries.length === 0 && (
                    <div className="flex flex-col items-center justify-center py-16 text-[#666]">
                        <Shield size={48} className="mb-4 opacity-20" />
                        <div className="text-lg">No audit entries found</div>
                        <div className="text-xs mt-2">Try adjusting your filters</div>
                    </div>
                )}
            </div>

            {/* Footer Stats */}
            <div className="flex-shrink-0 px-6 py-2 bg-[#1E2026] border-t border-[#383A42] flex items-center justify-between text-xs text-[#888]">
                <div>Showing {filteredEntries.length} of {entries.length} entries</div>
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-[#10B981]"></div>
                        <span>Success: {entries.filter(e => e.status === 'success').length}</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-[#EF4444]"></div>
                        <span>Failed: {entries.filter(e => e.status === 'failed').length}</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-[#F59E0B]"></div>
                        <span>Blocked: {entries.filter(e => e.status === 'blocked').length}</span>
                    </div>
                </div>
            </div>
        </div>
    );
}

// Generate mock audit entries
function generateMockEntries(): AuditEntry[] {
    const admins = [
        { id: 'adm-001', name: 'John Anderson' },
        { id: 'adm-002', name: 'Sarah Chen' },
        { id: 'adm-003', name: 'Mike Rodriguez' },
        { id: 'adm-004', name: 'Emily Watson' },
        { id: 'adm-005', name: 'David Kim' },
    ];

    const actionTypes: ActionType[] = ['CREATE', 'UPDATE', 'DELETE', 'LOGIN', 'EXPORT'];
    const modules: ModuleType[] = ['Accounts', 'Symbols', 'Orders', 'Settings', 'Users', 'Groups', 'LP', 'Risk', 'Reports', 'System'];
    const statuses: Array<'success' | 'failed' | 'blocked'> = ['success', 'success', 'success', 'success', 'success', 'failed', 'blocked'];

    const entries: AuditEntry[] = [];
    const now = new Date();

    for (let i = 0; i < 100; i++) {
        const admin = admins[Math.floor(Math.random() * admins.length)];
        const actionType = actionTypes[Math.floor(Math.random() * actionTypes.length)];
        const module = modules[Math.floor(Math.random() * modules.length)];
        const status = statuses[Math.floor(Math.random() * statuses.length)];

        const timestamp = new Date(now.getTime() - Math.random() * 7 * 24 * 60 * 60 * 1000);

        let description = '';
        let target = '';
        let details: AuditEntry['details'] = {};

        // Generate contextual description based on action type and module
        switch (actionType) {
            case 'CREATE':
                target = `${module.toLowerCase()}-${Math.floor(Math.random() * 10000)}`;
                description = `Created new ${module.toLowerCase()} record`;
                details = {
                    after: {
                        name: `New ${module} Item`,
                        status: 'active',
                        createdBy: admin.name,
                    },
                    metadata: {
                        source: 'admin-panel',
                        environment: 'production',
                    },
                };
                break;
            case 'UPDATE':
                target = `${module.toLowerCase()}-${Math.floor(Math.random() * 10000)}`;
                description = `Updated ${module.toLowerCase()} configuration`;
                details = {
                    before: { status: 'pending', leverage: '1:100' },
                    after: { status: 'active', leverage: '1:200' },
                    metadata: {
                        fields_changed: ['status', 'leverage'],
                        validation: 'passed',
                    },
                };
                break;
            case 'DELETE':
                target = `${module.toLowerCase()}-${Math.floor(Math.random() * 10000)}`;
                description = `Deleted ${module.toLowerCase()} record`;
                details = {
                    before: { id: target, status: 'inactive' },
                    reason: status === 'blocked' ? 'Insufficient permissions' : 'Cleanup routine',
                    metadata: { permanent: true },
                };
                break;
            case 'LOGIN':
                target = 'admin-panel';
                description = status === 'success' ? 'Successful login' : 'Failed login attempt';
                details = {
                    reason: status === 'failed' ? 'Invalid credentials' : undefined,
                    metadata: {
                        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
                        twoFactorUsed: Math.random() > 0.5,
                    },
                };
                break;
            case 'EXPORT':
                target = `${module.toLowerCase()}-report`;
                description = `Exported ${module.toLowerCase()} data`;
                details = {
                    metadata: {
                        format: 'CSV',
                        recordCount: Math.floor(Math.random() * 10000),
                        fileSize: `${Math.floor(Math.random() * 5000)}KB`,
                    },
                };
                break;
        }

        if (status === 'blocked') {
            details.reason = 'Action blocked by security policy';
        } else if (status === 'failed') {
            details.reason = 'Operation failed: validation error';
        }

        entries.push({
            id: `audit-${i}`,
            timestamp,
            adminName: admin.name,
            adminId: admin.id,
            actionType,
            module,
            target,
            description,
            ipAddress: `192.168.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`,
            status,
            details,
        });
    }

    // Sort by timestamp descending (newest first)
    return entries.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
}
