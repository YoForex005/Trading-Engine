import React, { useState, useEffect } from 'react';
import ContextMenu, { ContextAction } from '../ui/ContextMenu';
import { ChevronRight, ChevronDown, Filter, Search, Wifi, WifiOff } from 'lucide-react';
import { useAdminData } from '../../hooks/useAdminData';

// Mock Data as fallback (kept for initial UI development)
const MOCK_ACCOUNTS_FULL: any[] = Array.from({ length: 50 }).map((_, i) => ({
    id: 1000 + i,
    login: (5001092 + i).toString(),
    name: i % 3 === 0 ? 'John Doe' : i % 3 === 1 ? 'Jane Smith' : 'Risk Desk',
    group: i % 2 === 0 ? 'real\\standard' : 'demo\\pro',
    leverage: '1:500',
    balance: (10000 + Math.random() * 50000).toFixed(2),
    credit: '0.00',
    equity: (10000 + Math.random() * 50000 + (Math.random() > 0.5 ? 500 : -500)).toFixed(2),
    margin: (100 + Math.random() * 1000).toFixed(2),
    freeMargin: (9000).toFixed(2),
    marginLevel: (5000).toFixed(0),
    profit: (Math.random() * 1000 - 200).toFixed(2),
    floatingPL: (Math.random() * 500 - 100).toFixed(2),
    swap: '-2.50',
    commission: '-5.00',
    status: i === 3 ? 'SUSPENDED' : i === 4 ? 'MARGIN_CALL' : 'ACTIVE',
    flags: 'Fully Hedged',
    country: i % 5 === 0 ? 'United Kingdom' : i % 5 === 1 ? 'United States' : i % 5 === 2 ? 'Germany' : i % 5 === 3 ? 'France' : 'Japan',
    email: 'user@example.com',
    phone: '+44 7911 123456',
    regTime: '2025.10.15 14:30',
    lastAccess: '2026.01.19 16:45',
    lastIP: '192.168.1.1',
    mqID: '883922',
    currency: 'USD',
    comment: 'VIP Client',
    leadSource: 'Google Ads',
    leadCampaign: 'Q4_Promo',
    agentAccount: '',
    bankAccount: '',
}));

export default function AccountsView() {
    // PERFORMANCE FIX: Real-time WebSocket data instead of mocks
    const { accounts: liveAccounts, isConnected, connectionState, refresh } = useAdminData({
        autoConnect: true,
        fallbackToPolling: true,
    });

    // Use live data if available, otherwise fall back to mocks
    const accounts = liveAccounts.length > 0 ? liveAccounts : MOCK_ACCOUNTS_FULL;
    const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
    const [lastSelectedId, setLastSelectedId] = useState<number | null>(null);
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; account: any } | null>(null);
    const [showTreeNav, setShowTreeNav] = useState(true);
    const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set(['real\\standard', 'demo\\pro']));
    const [showFilters, setShowFilters] = useState(false);
    const [filters, setFilters] = useState({
        group: [] as string[],
        status: [] as string[],
        country: [] as string[],
        search: ''
    });

    // Initial Column State - Load from localStorage or defaults
    const [visibleColumns, setVisibleColumns] = useState<Record<string, boolean>>(() => {
        const saved = localStorage.getItem('mt5-accounts-columns');
        if (saved) return JSON.parse(saved);
        return {
            login: true, name: true, group: true, leverage: true, balance: true,
            credit: true, equity: true, margin: true, freeMargin: true, marginLevel: true,
            profit: true, floatingPL: true, swap: true, status: true, country: true,
            email: true, comment: true,
            // Hidden by default
            commission: false, currency: false, flags: false,
            phone: false, regTime: false, lastAccess: false, lastIP: false,
            mqID: false, agentAccount: false, bankAccount: false,
            leadSource: false, leadCampaign: false
        };
    });

    const COLUMNS_DEF = [
        { key: 'login', label: 'Login', w: 'w-16', align: 'left' },
        { key: 'name', label: 'Name', w: 'w-32', align: 'left' },
        { key: 'group', label: 'Group', w: 'w-24', align: 'left' },
        { key: 'leverage', label: 'Lev', w: 'w-12', align: 'left' },
        { key: 'balance', label: 'Balance', w: 'w-24', align: 'right', mono: true },
        { key: 'credit', label: 'Credit', w: 'w-16', align: 'right', mono: true },
        { key: 'equity', label: 'Equity', w: 'w-24', align: 'right', mono: true, colorLogic: true },
        { key: 'margin', label: 'Margin', w: 'w-20', align: 'right', mono: true },
        { key: 'freeMargin', label: 'Free Margin', w: 'w-24', align: 'right', mono: true },
        { key: 'marginLevel', label: 'Margin %', w: 'w-16', align: 'right', mono: true },
        { key: 'profit', label: 'Profit', w: 'w-20', align: 'right', mono: true, colorLogic: true },
        { key: 'floatingPL', label: 'Floating P/L', w: 'w-20', align: 'right', mono: true, colorLogic: true },
        { key: 'swap', label: 'Swap', w: 'w-16', align: 'right', mono: true },
        { key: 'commission', label: 'Commission', w: 'w-20', align: 'right', mono: true },
        { key: 'currency', label: 'Curr', w: 'w-12', align: 'left' },
        { key: 'status', label: 'Status', w: 'w-20', align: 'left' },
        { key: 'flags', label: 'Flags', w: 'w-24', align: 'left' },
        { key: 'country', label: 'Country', w: 'w-24', align: 'left' },
        { key: 'email', label: 'Email', w: 'w-40', align: 'left' },
        { key: 'comment', label: 'Comment', w: 'w-32', align: 'left' },
        { key: 'regTime', label: 'Registration', w: 'w-32', align: 'left' },
        { key: 'lastAccess', label: 'Last Access', w: 'w-32', align: 'left' },
        { key: 'lastIP', label: 'Last IP', w: 'w-24', align: 'left' },
        { key: 'agentAccount', label: 'Agent', w: 'w-20', align: 'left' },
        { key: 'bankAccount', label: 'Bank', w: 'w-20', align: 'left' },
        { key: 'leadSource', label: 'Lead Source', w: 'w-24', align: 'left' },
        { key: 'leadCampaign', label: 'Campaign', w: 'w-24', align: 'left' },
        { key: 'phone', label: 'Phone', w: 'w-24', align: 'left' },
        { key: 'mqID', label: 'MQ ID', w: 'w-20', align: 'left' },
    ];

    const activeColumns = COLUMNS_DEF.filter(c => visibleColumns[c.key]);

    const handleRowClick = (e: React.MouseEvent, id: number) => {
        if (e.ctrlKey) {
            const newSet = new Set(selectedIds);
            if (newSet.has(id)) newSet.delete(id);
            else newSet.add(id);
            setSelectedIds(newSet);
        } else {
            setSelectedIds(new Set([id]));
        }
        setLastSelectedId(id);
    };

    const handleRightClick = (e: React.MouseEvent, acc: any) => {
        e.preventDefault();
        // If clicking outside selection, reset selection to clicked item
        if (!selectedIds.has(acc.id)) {
            setSelectedIds(new Set([acc.id]));
        }
        setContextMenu({ x: e.clientX, y: e.clientY, account: acc });
    };

    // Persist column config
    useEffect(() => {
        localStorage.setItem('mt5-accounts-columns', JSON.stringify(visibleColumns));
    }, [visibleColumns]);

    const toggleColumn = (key: string) => {
        setVisibleColumns(prev => ({ ...prev, [key]: !prev[key] }));
    };

    const toggleGroup = (group: string) => {
        setExpandedGroups(prev => {
            const next = new Set(prev);
            if (next.has(group)) next.delete(group);
            else next.add(group);
            return next;
        });
    };

    // Filter logic
    const filteredAccounts = accounts.filter(acc => {
        if (filters.group.length > 0 && !filters.group.includes(acc.group)) return false;
        if (filters.status.length > 0 && !filters.status.includes(acc.status)) return false;
        if (filters.country.length > 0 && !filters.country.includes(acc.country)) return false;
        if (filters.search && !acc.login.includes(filters.search) &&
            !acc.name.toLowerCase().includes(filters.search.toLowerCase())) return false;
        return true;
    });

    // Group accounts for tree navigator
    const groupedAccounts = filteredAccounts.reduce((acc, account) => {
        const group = account.group;
        if (!acc[group]) acc[group] = [];
        acc[group].push(account);
        return acc;
    }, {} as Record<string, typeof accounts>);

    const uniqueGroups = [...new Set(accounts.map(a => a.group))];
    const uniqueStatuses = [...new Set(accounts.map(a => a.status))];
    const uniqueCountries = [...new Set(accounts.map(a => a.country))];

    const toggleFilter = (type: 'group' | 'status' | 'country', value: string) => {
        setFilters(prev => {
            const arr = prev[type];
            const next = arr.includes(value)
                ? arr.filter(v => v !== value)
                : [...arr, value];
            return { ...prev, [type]: next };
        });
    };

    const getMenuActions = (acc: any): ContextAction[] => [
        { label: 'New Account', onClick: () => console.log('New'), shortcut: 'Ctrl+Shift+N' },
        { label: 'Account Details', onClick: () => console.log('Details'), shortcut: 'Enter' },
        {
            label: 'Bulk Operations',
            hasSubmenu: true,
            submenu: [
                { label: 'Charges', onClick: () => console.log('Charges') },
                { label: 'Check Balance', onClick: () => console.log('Check Balance') },
                { label: 'Fix Balance', onClick: () => console.log('Fix Balance') },
                { label: 'Fix Personal Data', onClick: () => console.log('Fix Data') },
                { label: 'Bulk Closing', onClick: () => console.log('Closing') },
                { label: 'Bulk Payments', onClick: () => console.log('Payments') },
                { label: 'Split Positions', onClick: () => console.log('Split') },
            ]
        },
        { label: 'Internal Mail / Email', onClick: () => console.log('Mail') },
        { label: 'Push Notification / SMS', onClick: () => console.log('Push'), separator: true },

        {
            label: 'Select By',
            hasSubmenu: true,
            submenu: [
                { label: 'Group', onClick: () => console.log('Sel Group') },
                { label: 'Country', onClick: () => console.log('Sel Country') },
                { label: 'Custom', onClick: () => console.log('Sel Custom') },
            ]
        },
        { label: 'Filter', onClick: () => console.log('Filter'), hasSubmenu: true, submenu: [{ label: 'Advanced Filter', onClick: () => { } }] },
        {
            label: 'Copy As',
            hasSubmenu: true,
            submenu: [
                { label: 'Copy to Clipboard', onClick: () => console.log('Copy') },
                { label: 'CSV', onClick: () => console.log('CSV') },
                { label: 'HTML', onClick: () => console.log('HTML') },
            ]
        },
        { label: 'Export', onClick: () => console.log('Export') },
        { label: 'Import', onClick: () => console.log('Import') },
        { label: 'Find', onClick: () => console.log('Find'), shortcut: 'Ctrl+F', separator: true },

        { label: 'Auto Scroll', onClick: () => console.log('Auto Scroll') },
        { label: 'Auto Arrange', onClick: () => console.log('Auto Arrange'), shortcut: 'A', checked: true },
        { label: 'Grid', onClick: () => console.log('Grid'), shortcut: 'G', checked: true, separator: true },

        {
            label: 'Columns',
            hasSubmenu: true,
            // Map ALL columns to checkbox items
            submenu: COLUMNS_DEF.map(col => ({
                label: col.label,
                checked: visibleColumns[col.key],
                onClick: () => toggleColumn(col.key)
            }))
        },
    ];

    return (
        <div className="flex flex-col h-full bg-[#121316] select-none cursor-default font-sans text-[11px]">
            {/* Toolbar with Filters */}
            <div className="h-7 bg-[#1E2026] border-b border-[#383A42] flex items-center px-2 gap-2">
                <button
                    onClick={() => setShowTreeNav(!showTreeNav)}
                    className="h-5 px-2 bg-[#333] border border-[#444] text-[#CCC] hover:bg-[#444]"
                >
                    {showTreeNav ? '◀ Hide Tree' : '▶ Show Tree'}
                </button>
                <button
                    onClick={() => setShowFilters(!showFilters)}
                    className="h-5 px-2 bg-[#333] border border-[#444] text-[#CCC] hover:bg-[#444] flex items-center gap-1"
                >
                    <Filter size={12} />
                    Filters
                </button>
                {/* Connection Status Indicator */}
                <div className={`h-5 px-2 flex items-center gap-1 text-[10px] border ${
                    isConnected
                        ? 'bg-emerald-900/20 border-emerald-700 text-emerald-400'
                        : connectionState === 'connecting'
                        ? 'bg-yellow-900/20 border-yellow-700 text-yellow-400'
                        : 'bg-zinc-800 border-zinc-600 text-zinc-400'
                }`}>
                    {isConnected ? (
                        <>
                            <Wifi size={10} />
                            <span>Live</span>
                        </>
                    ) : connectionState === 'connecting' ? (
                        <>
                            <div className="w-2 h-2 rounded-full bg-yellow-400 animate-pulse" />
                            <span>Connecting...</span>
                        </>
                    ) : (
                        <>
                            <WifiOff size={10} />
                            <span>Mock Data</span>
                        </>
                    )}
                </div>
                <div className="flex-1" />
                <div className="relative">
                    <Search size={12} className="absolute left-1 top-1 text-[#666]" />
                    <input
                        type="text"
                        placeholder="Search login, name..."
                        value={filters.search}
                        onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
                        className="h-5 pl-6 pr-2 w-48 bg-[#252526] border border-[#444] text-[#CCC] text-[10px]"
                    />
                </div>
                {selectedIds.size > 0 && (
                    <div className="flex items-center gap-1 border-l border-[#444] pl-2">
                        <span className="text-[#F5C542] font-bold">{selectedIds.size} selected</span>
                        <button className="h-5 px-2 bg-[#333] border border-[#444] text-[#CCC] hover:bg-[#444]">
                            Change Group
                        </button>
                        <button className="h-5 px-2 bg-[#333] border border-[#444] text-[#CCC] hover:bg-[#444]">
                            Disable
                        </button>
                        <button className="h-5 px-2 bg-[#E74C3C] border border-[#922B21] text-white hover:brightness-110">
                            Bulk Action
                        </button>
                    </div>
                )}
            </div>

            {/* Filter Panel */}
            {showFilters && (
                <div className="bg-[#1E2026] border-b border-[#383A42] p-2 flex gap-4">
                    <div className="flex-1">
                        <div className="text-[#888] mb-1">Group:</div>
                        <div className="flex flex-wrap gap-1">
                            {uniqueGroups.map(g => (
                                <label key={g} className="flex items-center gap-1 px-2 py-0.5 bg-[#252526] border border-[#444] cursor-pointer hover:bg-[#333]">
                                    <input
                                        type="checkbox"
                                        checked={filters.group.includes(g)}
                                        onChange={() => toggleFilter('group', g)}
                                        className="w-3 h-3"
                                    />
                                    <span className="text-[10px]">{g}</span>
                                </label>
                            ))}
                        </div>
                    </div>
                    <div className="flex-1">
                        <div className="text-[#888] mb-1">Status:</div>
                        <div className="flex flex-wrap gap-1">
                            {uniqueStatuses.map(s => (
                                <label key={s} className="flex items-center gap-1 px-2 py-0.5 bg-[#252526] border border-[#444] cursor-pointer hover:bg-[#333]">
                                    <input
                                        type="checkbox"
                                        checked={filters.status.includes(s)}
                                        onChange={() => toggleFilter('status', s)}
                                        className="w-3 h-3"
                                    />
                                    <span className={`text-[10px] ${
                                        s === 'ACTIVE' ? 'text-rtx-active' :
                                        s === 'MARGIN_CALL' ? 'text-rtx-warning' :
                                        'text-rtx-suspended'
                                    }`}>{s}</span>
                                </label>
                            ))}
                        </div>
                    </div>
                    <div className="flex-1">
                        <div className="text-[#888] mb-1">Country:</div>
                        <div className="flex flex-wrap gap-1">
                            {uniqueCountries.map(c => (
                                <label key={c} className="flex items-center gap-1 px-2 py-0.5 bg-[#252526] border border-[#444] cursor-pointer hover:bg-[#333]">
                                    <input
                                        type="checkbox"
                                        checked={filters.country.includes(c)}
                                        onChange={() => toggleFilter('country', c)}
                                        className="w-3 h-3"
                                    />
                                    <span className="text-[10px]">{c}</span>
                                </label>
                            ))}
                        </div>
                    </div>
                </div>
            )}

            {/* Main Layout: Tree Navigator + Table */}
            <div className="flex-1 flex overflow-hidden">
                {/* Tree Navigator */}
                {showTreeNav && (
                    <div className="w-48 bg-[#1E2026] border-r border-[#383A42] overflow-auto flex-shrink-0">
                        <div className="p-1">
                            <div className="font-bold text-[#F5C542] mb-1 px-1">Servers</div>
                            <div className="pl-2">
                                <div className="font-bold text-[#CCC] mb-1 px-1">Groups</div>
                                {uniqueGroups.map(group => {
                                    const count = groupedAccounts[group]?.length || 0;
                                    const isExpanded = expandedGroups.has(group);
                                    return (
                                        <div key={group}>
                                            <div
                                                onClick={() => toggleGroup(group)}
                                                className="flex items-center gap-1 px-1 py-0.5 cursor-pointer hover:bg-[#333] text-[#CCC]"
                                            >
                                                {isExpanded ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
                                                <span className="flex-1">{group}</span>
                                                <span className="text-[#666] bg-[#252526] px-1 rounded-sm text-[9px]">{count}</span>
                                            </div>
                                            {isExpanded && groupedAccounts[group]?.map(acc => (
                                                <div
                                                    key={acc.id}
                                                    onClick={(e) => handleRowClick(e, acc.id)}
                                                    className={`pl-6 py-0.5 cursor-pointer text-[10px] truncate ${
                                                        selectedIds.has(acc.id) ? 'bg-[#2B3D55] text-white' : 'text-[#AAA] hover:bg-[#333]'
                                                    }`}
                                                >
                                                    {acc.login} - {acc.name}
                                                </div>
                                            ))}
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    </div>
                )}

                {/* Data Table */}
                <div className="flex-1 overflow-auto custom-scrollbar relative">
                    <table className="min-w-max w-full border-collapse table-fixed">
                        <thead className="sticky top-0 z-10 bg-[#1E2026] text-[#A0A0A0] shadow-[0_1px_0_0_#000]">
                            <tr className="h-5">
                                {activeColumns.map(col => (
                                    <th key={col.key} className={`${col.w} px-2 py-0 border-r border-[#383A42] text-${col.align} font-normal truncate`}>
                                        {col.label}
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody className="bg-[#121316]">
                            {filteredAccounts.map((acc, i) => {
                            const isSelected = selectedIds.has(acc.id);

                            // Color Logic
                            const equityVal = parseFloat(acc.equity);
                            const balanceVal = parseFloat(acc.balance);
                            const profitVal = parseFloat(acc.profit);
                            const floatingPLVal = parseFloat(acc.floatingPL);

                            let equityColor = 'text-[#F0F0F0]';
                            if (equityVal < balanceVal) equityColor = 'text-rtx-loss';
                            if (equityVal > balanceVal) equityColor = 'text-rtx-profit';

                            let profitColor = 'text-[#F0F0F0]';
                            if (profitVal < 0) profitColor = 'text-rtx-loss';
                            if (profitVal > 0) profitColor = 'text-rtx-profit';

                            let floatingPLColor = 'text-[#F0F0F0]';
                            if (floatingPLVal < 0) floatingPLColor = 'text-rtx-loss';
                            if (floatingPLVal > 0) floatingPLColor = 'text-rtx-profit';

                            let statusColor = 'text-[#F0F0F0]';
                            if (acc.status === 'ACTIVE') statusColor = 'text-rtx-active';
                            if (acc.status === 'MARGIN_CALL') statusColor = 'text-rtx-warning';
                            if (acc.status === 'SUSPENDED') statusColor = 'text-rtx-suspended';

                            return (
                                <tr
                                    key={acc.id}
                                    onClick={(e) => handleRowClick(e, acc.id)}
                                    onContextMenu={(e) => handleRightClick(e, acc)}
                                    className={`
                                        h-5 leading-5 border-b border-[#2A2C33]
                                        ${isSelected ? 'bg-[#2B3D55] text-white' : (i % 2 === 0 ? 'bg-transparent' : 'bg-[#16171A]')}
                                        group
                                    `}
                                >
                                    {activeColumns.map(col => {
                                        const content = acc[col.key];
                                        let cellClass = `px-2 border-r border-[#383A42] truncate align-middle text-${col.align} ${col.mono ? 'font-mono' : ''}`;

                                        if (!isSelected) {
                                            if (col.key === 'equity') cellClass += ` ${equityColor}`;
                                            if (col.key === 'profit') cellClass += ` ${profitColor}`;
                                            if (col.key === 'floatingPL') cellClass += ` ${floatingPLColor}`;
                                            if (col.key === 'status') cellClass += ` ${statusColor}`;
                                        } else {
                                            cellClass += ' text-white';
                                        }

                                        return (
                                            <td key={col.key} className={cellClass}>
                                                {col.key === 'balance' || col.key === 'equity' || col.key === 'margin' ||
                                                 col.key === 'freeMargin' || col.key === 'credit' || col.key === 'profit' ||
                                                 col.key === 'floatingPL' || col.key === 'swap' || col.key === 'commission'
                                                    ? Number(content).toLocaleString('en-US', { minimumFractionDigits: 2 })
                                                    : content}
                                            </td>
                                        );
                                    })}
                                </tr>
                            );
                        })}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Bottom Status Bar */}
            <div className="h-5 flex items-center bg-[#1E2026] border-t border-[#383A42] px-2 text-[#888] gap-4 mt-auto text-[10px]">
                <span className="w-32">{filteredAccounts.length} / {accounts.length} Accounts</span>
                <span className="w-32">{selectedIds.size} Selected</span>
                {filters.group.length > 0 && (
                    <span className="text-[#F5C542]">Group Filter: {filters.group.join(', ')}</span>
                )}
                {filters.status.length > 0 && (
                    <span className="text-[#F5C542]">Status Filter: {filters.status.join(', ')}</span>
                )}
                <div className="flex-1" />
                <span className="font-mono text-[#AAA]">Server: 12ms</span>
            </div>

            {contextMenu && (
                <ContextMenu
                    x={contextMenu.x}
                    y={contextMenu.y}
                    onClose={() => setContextMenu(null)}
                    actions={getMenuActions(contextMenu.account)}
                />
            )}
        </div>
    );
}
