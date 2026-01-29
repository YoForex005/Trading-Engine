import React, { useState } from 'react';
import {
    Server,
    BarChart3,
    FileText,
    Users,
    Briefcase,
    CreditCard,
    ShieldAlert,
    Globe,
    Settings,
    ChevronRight,
    ChevronDown,
    Folder,
    Monitor,
    RefreshCw,
    Activity,
    Layers,
    Plug,
    Mail,
    Headphones,
    ShoppingBag,
    Plus,
    UserPlus,
    Trash2,
    RefreshCcw
} from 'lucide-react';
import ContextMenu, { ContextAction } from '../ui/ContextMenu';

interface TreeNode {
    id: string;
    label: string;
    type?: 'server' | 'group' | 'manager' | 'account' | 'report' | 'generic';
    icon?: React.ReactNode;
    children?: TreeNode[];
    expanded?: boolean;
}

const ICON_SIZE = 12;

const INITIAL_DATA: TreeNode[] = [
    {
        id: 'servers', label: 'Servers', type: 'server', icon: <Monitor size={ICON_SIZE} />, expanded: true, children: [
            { id: 'srv-flexy', label: 'FlexyMarkets-Server', type: 'server', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            {
                id: 'srv-alfx', label: 'ALFX-Trade', type: 'server', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" />, expanded: true, children: [
                    { id: 'srv-alfx-user', label: '2015 - Alfx prop', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" /> }
                ]
            },
        ]
    },
    {
        id: 'analytics', label: 'Analytics', type: 'generic', icon: <BarChart3 size={ICON_SIZE} className="text-[#3B82F6]" />, children: [
            { id: 'an-accounts', label: 'Trading Accounts', icon: <BarChart3 size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'an-users', label: 'Online Users', icon: <BarChart3 size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'an-positions', label: 'Open Positions', icon: <BarChart3 size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'an-orders', label: 'Open Orders', icon: <BarChart3 size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'an-ads', label: 'Advertising Campaigns', icon: <BarChart3 size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'an-country', label: 'By Country', icon: <BarChart3 size={ICON_SIZE} className="text-[#3B82F6]" /> },
        ]
    },
    {
        id: 'reports', label: 'Server Reports', type: 'report', icon: <FileText size={ICON_SIZE} className="text-[#3B82F6]" />, children: [
            { id: 'rep-accounts', label: 'Accounts', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'rep-capital', label: 'Capital', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'rep-daily', label: 'Daily', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'rep-emir', label: 'EMIR', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'rep-funds', label: 'Funds', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'rep-gateways', label: 'Gateways', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'rep-nfa', label: 'NFA', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'rep-traders', label: 'Traders', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'rep-trades', label: 'Trades', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'rep-ultency', label: 'Ultency', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
        ]
    },
    {
        id: 'clients_orders', label: 'Clients & Orders', type: 'manager', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" />, expanded: true, children: [
            { id: 'cl-online', label: 'Online Users', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'cl-clients', label: 'Clients', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'accounts', label: 'Trading Accounts (86)', type: 'account', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'positions', label: 'Positions (20)', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'orders', label: 'Orders', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" /> },
        ]
    },
    { id: 'payments', label: 'Payments', icon: <CreditCard size={ICON_SIZE} className="text-[#3B82F6]" /> },
    { id: 'matching', label: 'Ultency Matching Engine', icon: <Settings size={ICON_SIZE} className="text-[#9B59B6]" /> },
    { id: 'subscriptions', label: 'Subscriptions', icon: <RefreshCw size={ICON_SIZE} className="text-[#2ECC71]" /> },
    { id: 'dealing', label: 'Dealing', icon: <Activity size={ICON_SIZE} className="text-[#E74C3C]" /> },
    { id: 'leverages', label: 'Leverages', icon: <Layers size={ICON_SIZE} className="text-[#2ECC71]" /> },
    {
        id: 'groups', label: 'Groups (6)', type: 'group', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" />, children: [
            { id: 'grp-alfx', label: 'ALFX-B (2)', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'grp-ch', label: 'ch (2)', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'grp-demo', label: 'demo (2)', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
        ]
    },
    { id: 'plugins', label: 'Plugins', icon: <Plug size={ICON_SIZE} className="text-[#3B82F6]" /> },
    { id: 'mailbox', label: 'Mailbox (1)', icon: <Mail size={ICON_SIZE} className="text-[#E67E22]" /> },
    { id: 'support', label: 'Support Center', icon: <Headphones size={ICON_SIZE} className="text-[#3B82F6]" /> },
    { id: 'appstore', label: 'App Store', icon: <ShoppingBag size={ICON_SIZE} className="text-[#E67E22]" /> },
];

interface NavigatorProps {
    onNavigate?: (id: string) => void;
}

export default function Navigator({ onNavigate }: NavigatorProps) {
    const [data, setData] = useState(INITIAL_DATA);
    const [selectedId, setSelectedId] = useState<string | null>(null);
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; node: TreeNode } | null>(null);

    const toggleNode = (id: string) => {
        const update = (nodes: TreeNode[]): TreeNode[] => {
            return nodes.map((node) => {
                if (node.id === id) return { ...node, expanded: !node.expanded };
                if (node.children) return { ...node, children: update(node.children) };
                return node;
            });
        };
        setData(update(data));
    };

    const handleNodeClick = (node: TreeNode) => {
        setSelectedId(node.id);
        if (node.children) {
            toggleNode(node.id);
        } else {
            if (onNavigate) onNavigate(node.id);
        }
    };

    const handleContextMenu = (e: React.MouseEvent, node: TreeNode) => {
        e.preventDefault();
        setContextMenu({ x: e.clientX, y: e.clientY, node });
    };

    const getActionsForNode = (node: TreeNode): ContextAction[] => {
        const commonActions = [
            { label: 'Register', onClick: () => console.log('Register') },
            { label: 'Connect', onClick: () => console.log('Connect') },
            { label: 'Scan Network', onClick: () => console.log('Scan') },
            { label: 'Custom Command', onClick: () => console.log('Custom') },
        ];

        // 1. SERVERS
        if (node.id === 'servers' || node.type === 'server') {
            return [
                { label: 'Login', onClick: () => console.log('Login') },
                { label: 'Login to Web Terminal', onClick: () => console.log('WebTerm') },
                { label: 'Change Password', onClick: () => console.log('Pass') },
                { separator: true, label: '' },
                { label: 'Start', onClick: () => console.log('Start') },
                { label: 'Stop', onClick: () => console.log('Stop') },
                { label: 'Restart', onClick: () => console.log('Restart') },
                { separator: true, label: '' },
                {
                    label: 'Reports',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'Daily', onClick: () => console.log('Daily') },
                        { label: 'Monthly', onClick: () => console.log('Monthly') },
                        { label: 'Quarterly', onClick: () => console.log('Quarterly') },
                        { label: 'Custom', onClick: () => console.log('Custom') },
                    ]
                },
                {
                    label: 'Logs',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'Journal', onClick: () => console.log('Journal') },
                        { label: 'TradeServer', onClick: () => console.log('TradeServer') },
                        { label: 'Quotes', onClick: () => console.log('Quotes') },
                        { label: 'Antispam', onClick: () => console.log('Antispam') },
                    ]
                },
                { separator: true, label: '' },
                { label: 'Properties', onClick: () => console.log('Properties') },
            ];
        }

        // 2. GROUPS
        if (node.id.includes('groups') || node.type === 'group') {
            return [
                { label: 'Create Group', onClick: () => console.log('Create Group') },
                { label: 'Request Group', onClick: () => console.log('Request Group') },
                { separator: true, label: '' },
                { label: 'Symbols', onClick: () => console.log('Symbols') },
                { label: 'Commissions', onClick: () => console.log('Commissions') },
                { separator: true, label: '' },
                {
                    label: 'Export',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'To HTML', onClick: () => console.log('HTML') },
                        { label: 'To Excel', onClick: () => console.log('Excel') },
                    ]
                },
            ];
        }

        // 3. MANAGERS (Admins)
        if (node.id === 'admins' || node.type === 'manager' || node.id === 'clients_orders') {
            return [
                { label: 'Add Manager', onClick: () => console.log('Add Manager') },
                { label: 'Rights', onClick: () => console.log('Rights') },
                { separator: true, label: '' },
                { label: 'List', onClick: () => console.log('List') },
                {
                    label: 'View',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'Details', onClick: () => console.log('Details') },
                        { label: 'Summary', onClick: () => console.log('Summary') },
                    ]
                },
            ];
        }

        // 4. ACCOUNTS
        if (node.id === 'accounts' || node.type === 'account') {
            return [
                {
                    label: 'Create New Account',
                    onClick: () => {
                        if (onNavigate) onNavigate('action:create-account');
                    }
                },
                { label: 'Request Real Account', onClick: () => console.log('Request Real') },
                { separator: true, label: '' },
                { label: 'Deposit / Withdrawal', onClick: () => console.log('Balance') },
                { label: 'Credit Facility', onClick: () => console.log('Credit') },
                { separator: true, label: '' },
                { label: 'Change Password', onClick: () => console.log('Change Pass') },
                { label: 'Change Group', onClick: () => console.log('Change Group') },
                { separator: true, label: '' },
                {
                    label: 'Reports',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'Statement', onClick: () => console.log('Statement') },
                        { label: 'Deals', onClick: () => console.log('Deals') },
                        { label: 'Positions', onClick: () => console.log('Positions') },
                    ]
                },
                { label: 'Delete Account', danger: true, onClick: () => console.log('Delete') },
            ];
        }

        // GENERIC
        return [
            { label: 'Open', onClick: () => console.log('Open') },
            { label: 'Refresh', onClick: () => console.log('Refresh') },
            { separator: true, label: '' },
            { label: 'Properties', onClick: () => console.log('Properties') },
        ];
    };

    const renderTree = (nodes: TreeNode[], depth = 0) => {
        return nodes.map((node) => (
            <div key={node.id}>
                <div
                    className={`
                        flex items-center gap-1.5 px-2 py-0.5 cursor-default text-[11px] select-none
                        ${selectedId === node.id ? 'bg-[#2980B9] text-white' : 'text-[#CCC] hover:bg-[#25272E] hover:text-white'}
                    `}
                    style={{ paddingLeft: `${depth * 12 + 8}px` }}
                    onClick={() => handleNodeClick(node)}
                    onContextMenu={(e) => handleContextMenu(e, node)}
                >
                    {node.children ? (
                        node.expanded ? <ChevronDown size={10} className={selectedId === node.id ? 'text-white' : 'text-[#666]'} /> : <ChevronRight size={10} className={selectedId === node.id ? 'text-white' : 'text-[#666]'} />
                    ) : (
                        <div className="w-[10px]" />
                    )}

                    <span className={selectedId === node.id ? 'text-white' : 'text-[#888]'}>{node.icon}</span>
                    <span>{node.label}</span>
                </div>

                {node.children && node.expanded && (
                    <div>
                        {renderTree(node.children, depth + 1)}
                    </div>
                )}
            </div>
        ));
    };

    return (
        <div className="h-full w-60 bg-[#121316] border-r border-[#383A42] flex flex-col font-sans select-none overflow-y-auto custom-scrollbar">
            <div className="px-2 py-1 text-[10px] font-bold text-[#666] uppercase tracking-wider bg-[#1E2026] border-b border-[#383A42]">
                Navigator
            </div>
            <div className="py-1">
                {renderTree(data)}
            </div>

            {contextMenu && (
                <ContextMenu
                    x={contextMenu.x}
                    y={contextMenu.y}
                    onClose={() => setContextMenu(null)}
                    actions={getActionsForNode(contextMenu.node)}
                />
            )}
        </div>
    );
}
