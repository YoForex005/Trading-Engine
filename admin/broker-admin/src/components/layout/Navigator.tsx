import React, { useState, useMemo } from 'react';
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
    RefreshCcw,
    LayoutDashboard,
    Coins,
    Shield,
    Banknote,
    Tag,
    Wallet,
    AlertTriangle,
    Clock,
    Target,
    Database,
    Zap,
    MessageSquare,
    Palette,
    Trophy,
    Gift,
    TrendingUp,
    Calendar
} from 'lucide-react';
import ContextMenu, { ContextAction } from '../ui/ContextMenu';
import { useRole, type AdminRole } from '@/hooks/useRole';

interface TreeNode {
    id: string;
    label: string;
    type?: 'server' | 'group' | 'manager' | 'account' | 'report' | 'generic';
    icon?: React.ReactNode;
    children?: TreeNode[];
    expanded?: boolean;
    requiredRole?: AdminRole; // Minimum role required to see this item
    requiredPermission?: string; // Specific permission required
}

const ICON_SIZE = 12;

const INITIAL_DATA: TreeNode[] = [
    {
        id: 'dashboard-overview',
        label: 'Dashboard',
        icon: <LayoutDashboard size={ICON_SIZE} className="text-[#3B82F6]" />,
    },
    {
        id: 'notifications',
        label: 'Notifications',
        icon: <Mail size={ICON_SIZE} className="text-[#F5C542]" />,
    },
    {
        id: 'servers',
        label: 'Servers',
        type: 'server',
        icon: <Monitor size={ICON_SIZE} />,
        expanded: true,
        requiredRole: 'BROKER_ADMIN', // Only admins+ can see servers
        children: [
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
            { id: 'order-flow', label: 'Order Flow', icon: <BarChart3 size={ICON_SIZE} className="text-[#3B82F6]" /> },
        ]
    },
    {
        id: 'reports',
        label: 'Reports & Export',
        type: 'report',
        icon: <FileText size={ICON_SIZE} className="text-[#3B82F6]" />,
        requiredPermission: 'REPORTS_READ',
        children: [
            { id: 'scheduled-reports', label: 'Scheduled Reports', icon: <Calendar size={ICON_SIZE} className="text-[#10B981]" /> },
            { id: 'account-statement', label: 'Account Statement / Trade Export', icon: <FileText size={ICON_SIZE} className="text-[#3B82F6]" /> },
        ]
    },
    {
        id: 'audit-log',
        label: 'Audit Log',
        icon: <Shield size={ICON_SIZE} className="text-[#9B59B6]" />,
        requiredPermission: 'REPORTS_READ',
    },
    {
        id: 'user-management',
        label: 'User Management',
        icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" />,
        requiredRole: 'BROKER_ADMIN',
    },
    {
        id: 'client-management',
        label: 'Client Management',
        icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" />,
        requiredPermission: 'ACCOUNTS_READ',
    },
    {
        id: 'server-reports', label: 'Server Reports', type: 'report', icon: <FileText size={ICON_SIZE} className="text-[#3B82F6]" />, requiredRole: 'BROKER_ADMIN', children: [
            { id: 'broker-pnl', label: 'Broker P&L Summary', icon: <BarChart3 size={ICON_SIZE} className="text-[#22c55e]" /> },
            { id: 'execution-report', label: 'Execution Quality Report', icon: <Activity size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'trade-reconciliation', label: 'Trade Reconciliation', icon: <Activity size={ICON_SIZE} className="text-[#10B981]" /> },
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
        id: 'clients_orders',
        label: 'Clients & Orders',
        type: 'manager',
        icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" />,
        expanded: true,
        children: [
            { id: 'cl-online', label: 'Online Users', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'cl-clients', label: 'Clients', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'accounts', label: 'Trading Accounts (86)', type: 'account', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" />, requiredPermission: 'ACCOUNTS_READ' },
            { id: 'positions', label: 'Positions (20)', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" />, requiredPermission: 'TRADING_READ' },
            { id: 'orders', label: 'Orders', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" />, requiredPermission: 'TRADING_READ' },
            { id: 'admin-users', label: 'Admin Users', icon: <Users size={ICON_SIZE} className="text-[#E74C3C]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'communication-log', label: 'Communication Log', icon: <Mail size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'client-segmentation', label: 'Client Segmentation', icon: <Tag size={ICON_SIZE} className="text-[#F59E0B]" /> },
            { id: 'client-notes', label: 'Client Notes / CRM', icon: <FileText size={ICON_SIZE} className="text-[#10B981]" /> },
            { id: 'customer-lifecycle', label: 'Customer Lifecycle', icon: <TrendingUp size={ICON_SIZE} className="text-[#8B5CF6]" /> },
            { id: 'demo-account-manager', label: 'Demo Account Manager', icon: <UserPlus size={ICON_SIZE} className="text-[#3B82F6]" /> },
        ]
    },
    {
        id: 'finance',
        label: 'Finance',
        icon: <Briefcase size={ICON_SIZE} className="text-[#2ECC71]" />,
        requiredRole: 'BROKER_ADMIN',
        children: [
            { id: 'financial-ops', label: 'Financial Operations', icon: <Briefcase size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'payments', label: 'Payments', icon: <CreditCard size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'withdrawal-queue', label: 'Withdrawal Queue', icon: <Banknote size={ICON_SIZE} className="text-[#F5C542]" /> },
            { id: 'commissions', label: 'Commission Report', icon: <Coins size={ICON_SIZE} className="text-[#2ECC71]" /> },
            { id: 'commission-tiers', label: 'Commission Tiers / Rebates', icon: <TrendingUp size={ICON_SIZE} className="text-[#10B981]" /> },
            { id: 'affiliate-management', label: 'Affiliate / IB Management', icon: <UserPlus size={ICON_SIZE} className="text-[#9B59B6]" /> },
            { id: 'referral-program', label: 'Referral Program', icon: <Users size={ICON_SIZE} className="text-[#EC4899]" /> },
            { id: 'payment-gateway-config', label: 'Payment Gateway Config', icon: <CreditCard size={ICON_SIZE} className="text-[#10B981]" /> },
            { id: 'promotion-manager', label: 'Promotion / Bonus Manager', icon: <Gift size={ICON_SIZE} className="text-[#EC4899]" /> },
            { id: 'wallet-management', label: 'Multi-Currency Wallets', icon: <Wallet size={ICON_SIZE} className="text-[#3B82F6]" /> },
        ]
    },
    { id: 'risk-dashboard', label: 'Risk Dashboard', icon: <ShieldAlert size={ICON_SIZE} className="text-[#E74C3C]" />, requiredPermission: 'RISK_MANAGE' },
    { id: 'risk-rules', label: 'Risk Rules Engine', icon: <Shield size={ICON_SIZE} className="text-[#F59E0B]" />, requiredPermission: 'RISK_MANAGE' },
    { id: 'margin-call-dashboard', label: 'Margin Call Management', icon: <AlertTriangle size={ICON_SIZE} className="text-[#DC2626]" />, requiredPermission: 'RISK_MANAGE' },
    { id: 'client-risk-scoring', label: 'Client Risk Scoring / Credit Assessment', icon: <Shield size={ICON_SIZE} className="text-[#F59E0B]" />, requiredPermission: 'RISK_MANAGE' },
    {
        id: 'trading',
        label: 'Trading',
        icon: <Activity size={ICON_SIZE} className="text-[#E74C3C]" />,
        requiredPermission: 'TRADING_READ',
        children: [
            { id: 'symbol-management', label: 'Symbol Management', icon: <Coins size={ICON_SIZE} className="text-[#F5C542]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'instrument-groups', label: 'Instrument Groups / Symbol Categories', icon: <Settings size={ICON_SIZE} className="text-[#8B5CF6]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'group-management', label: 'Group Management', icon: <Users size={ICON_SIZE} className="text-[#9B59B6]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'trade-copier', label: 'Trade Copier', icon: <RefreshCcw size={ICON_SIZE} className="text-[#3B82F6]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'market-hours-config', label: 'Market Hours Config', icon: <Globe size={ICON_SIZE} className="text-[#10B981]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'trading-sessions', label: 'Trading Sessions / Market Hours', icon: <Clock size={ICON_SIZE} className="text-[#F5C542]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'lp-config', label: 'LP Configuration', icon: <Settings size={ICON_SIZE} className="text-[#3B82F6]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'swap-config', label: 'Swap Rate Config', icon: <RefreshCcw size={ICON_SIZE} className="text-[#F59E0B]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'dealing', label: 'Dealing', icon: <Activity size={ICON_SIZE} className="text-[#E74C3C]" />, requiredPermission: 'TRADING_EXECUTE' },
            { id: 'leverages', label: 'Leverages', icon: <Layers size={ICON_SIZE} className="text-[#2ECC71]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'trading-competitions', label: 'Trading Competitions', icon: <Trophy size={ICON_SIZE} className="text-[#F59E0B]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'spread-monitor', label: 'Spread Monitor', icon: <TrendingUp size={ICON_SIZE} className="text-[#3B82F6]" />, requiredRole: 'BROKER_ADMIN' },
            { id: 'trading-signals', label: 'Trading Signals / Signal Providers', icon: <Target size={ICON_SIZE} className="text-[#10B981]" />, requiredRole: 'BROKER_ADMIN' },
        ]
    },
    { id: 'matching', label: 'Ultency Matching Engine', icon: <Settings size={ICON_SIZE} className="text-[#9B59B6]" />, requiredRole: 'BROKER_ADMIN' },
    { id: 'subscriptions', label: 'Subscriptions', icon: <RefreshCw size={ICON_SIZE} className="text-[#2ECC71]" />, requiredRole: 'BROKER_ADMIN' },
    {
        id: 'system',
        label: 'System',
        icon: <Settings size={ICON_SIZE} className="text-[#9B59B6]" />,
        requiredRole: 'BROKER_ADMIN',
        children: [
            { id: 'backup-manager', label: 'Backup & Recovery', icon: <Database size={ICON_SIZE} className="text-[#10B981]" /> },
            { id: 'system-health', label: 'System Health Monitor', icon: <Activity size={ICON_SIZE} className="text-[#E74C3C]" /> },
            { id: 'performance-monitor', label: 'Performance Monitoring', icon: <Zap size={ICON_SIZE} className="text-[#F39C12]" /> },
            { id: 'server-config', label: 'Server Configuration', icon: <Server size={ICON_SIZE} className="text-[#F5C542]" /> },
            { id: 'platform-config', label: 'Platform Configuration', icon: <Settings size={ICON_SIZE} className="text-[#9B59B6]" /> },
            { id: 'lp-health', label: 'LP Health Monitoring', icon: <Activity size={ICON_SIZE} className="text-[#2ECC71]" /> },
            { id: 'api-keys', label: 'API Key Management', icon: <Shield size={ICON_SIZE} className="text-[#9B59B6]" /> },
            { id: 'webhooks', label: 'Webhook Management', icon: <Globe size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'email-templates', label: 'Email Templates', icon: <Mail size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'announcement-manager', label: 'Announcement Manager', icon: <Mail size={ICON_SIZE} className="text-[#9B59B6]" /> },
            { id: 'client-portal', label: 'Client Portal Config', icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'account-type-config', label: 'Account Type & Leverage', icon: <Settings size={ICON_SIZE} className="text-[#9B59B6]" /> },
            { id: 'support-tickets', label: 'Support Ticket System', icon: <MessageSquare size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'localization-config', label: 'Localization & Languages', icon: <Globe size={ICON_SIZE} className="text-[#10B981]" /> },
            { id: 'white-label', label: 'White Label / Branding', icon: <Palette size={ICON_SIZE} className="text-[#F97316]" /> },
            { id: 'client-activity-log', label: 'Client Activity / Session Log', icon: <Activity size={ICON_SIZE} className="text-[#3B82F6]" /> },
        ]
    },
    {
        id: 'security',
        label: 'Security',
        icon: <ShieldAlert size={ICON_SIZE} className="text-[#E74C3C]" />,
        requiredRole: 'BROKER_ADMIN',
        children: [
            { id: 'fraud-detection', label: 'Fraud Detection / IP Monitoring', icon: <ShieldAlert size={ICON_SIZE} className="text-[#DC2626]" /> },
            { id: 'ip-filter', label: 'IP Access Control', icon: <Shield size={ICON_SIZE} className="text-[#3B82F6]" /> },
            { id: 'audit-trail', label: 'Audit Trail / Activity Log', icon: <Activity size={ICON_SIZE} className="text-[#F59E0B]" /> },
            { id: 'server-logs', label: 'Server Log Viewer', icon: <Monitor size={ICON_SIZE} className="text-[#10B981]" /> },
        ]
    },
    {
        id: 'compliance',
        label: 'Compliance',
        icon: <ShieldAlert size={ICON_SIZE} className="text-[#9B59B6]" />,
        requiredRole: 'BROKER_ADMIN',
        children: [
            { id: 'compliance', label: 'Compliance Dashboard', icon: <Shield size={ICON_SIZE} className="text-[#2ECC71]" /> },
            { id: 'kyc', label: 'KYC/AML Verification', icon: <Users size={ICON_SIZE} className="text-[#F5C542]" /> },
            { id: 'client-documents', label: 'Client Documents / KYC', icon: <FileText size={ICON_SIZE} className="text-[#10B981]" /> },
            { id: 'compliance-reports', label: 'Compliance Reports', icon: <Shield size={ICON_SIZE} className="text-[#F5C542]" /> },
            { id: 'regulatory-reporting', label: 'Regulatory Reporting', icon: <FileText size={ICON_SIZE} className="text-[#8B5CF6]" /> },
        ]
    },
    {
        id: 'groups',
        label: 'Groups (6)',
        type: 'group',
        icon: <Users size={ICON_SIZE} className="text-[#3B82F6]" />,
        requiredRole: 'BROKER_ADMIN',
        children: [
            { id: 'grp-alfx', label: 'ALFX-B (2)', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'grp-ch', label: 'ch (2)', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
            { id: 'grp-demo', label: 'demo (2)', icon: <Folder size={ICON_SIZE} className="text-[#E67E22]" /> },
        ]
    },
    { id: 'plugins', label: 'Plugins', icon: <Plug size={ICON_SIZE} className="text-[#3B82F6]" />, requiredRole: 'SUPER_ADMIN' },
    { id: 'mailbox', label: 'Mailbox (1)', icon: <Mail size={ICON_SIZE} className="text-[#E67E22]" /> },
    { id: 'support', label: 'Support Center', icon: <Headphones size={ICON_SIZE} className="text-[#3B82F6]" /> },
    { id: 'appstore', label: 'App Store', icon: <ShoppingBag size={ICON_SIZE} className="text-[#E67E22]" />, requiredRole: 'BROKER_ADMIN' },
];

interface NavigatorProps {
    onNavigate?: (id: string) => void;
}

export default function Navigator({ onNavigate }: NavigatorProps) {
    const { role, hasPermission, hasRole } = useRole();
    const [data, setData] = useState(INITIAL_DATA);
    const [selectedId, setSelectedId] = useState<string | null>(null);
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; node: TreeNode } | null>(null);

    /**
     * Filter navigation tree based on user role and permissions
     */
    const filteredData = useMemo(() => {
        const filterTree = (nodes: TreeNode[]): TreeNode[] => {
            return nodes
                .filter(node => {
                    // If node requires a role, check role hierarchy
                    if (node.requiredRole && !hasRole(node.requiredRole)) {
                        return false;
                    }

                    // If node requires a permission, check permission
                    if (node.requiredPermission && !hasPermission(node.requiredPermission as any)) {
                        return false;
                    }

                    return true;
                })
                .map(node => {
                    // Recursively filter children
                    if (node.children) {
                        return {
                            ...node,
                            children: filterTree(node.children)
                        };
                    }
                    return node;
                });
        };

        return filterTree(data);
    }, [data, role, hasPermission, hasRole]);

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
            { label: 'Register', onClick: () => {} },
            { label: 'Connect', onClick: () => {} },
            { label: 'Scan Network', onClick: () => {} },
            { label: 'Custom Command', onClick: () => {} },
        ];

        // 1. SERVERS
        if (node.id === 'servers' || node.type === 'server') {
            return [
                { label: 'Login', onClick: () => {} },
                { label: 'Login to Web Terminal', onClick: () => {} },
                { label: 'Change Password', onClick: () => {} },
                { separator: true, label: '' },
                { label: 'Start', onClick: () => {} },
                { label: 'Stop', onClick: () => {} },
                { label: 'Restart', onClick: () => {} },
                { separator: true, label: '' },
                {
                    label: 'Reports',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'Daily', onClick: () => {} },
                        { label: 'Monthly', onClick: () => {} },
                        { label: 'Quarterly', onClick: () => {} },
                        { label: 'Custom', onClick: () => {} },
                    ]
                },
                {
                    label: 'Logs',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'Journal', onClick: () => {} },
                        { label: 'TradeServer', onClick: () => {} },
                        { label: 'Quotes', onClick: () => {} },
                        { label: 'Antispam', onClick: () => {} },
                    ]
                },
                { separator: true, label: '' },
                { label: 'Properties', onClick: () => {} },
            ];
        }

        // 2. GROUPS
        if (node.id.includes('groups') || node.type === 'group') {
            return [
                { label: 'Create Group', onClick: () => {} },
                { label: 'Request Group', onClick: () => {} },
                { separator: true, label: '' },
                { label: 'Symbols', onClick: () => {} },
                { label: 'Commissions', onClick: () => {} },
                { separator: true, label: '' },
                {
                    label: 'Export',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'To HTML', onClick: () => {} },
                        { label: 'To Excel', onClick: () => {} },
                    ]
                },
            ];
        }

        // 3. MANAGERS (Admins)
        if (node.id === 'admins' || node.type === 'manager' || node.id === 'clients_orders') {
            return [
                { label: 'Add Manager', onClick: () => {} },
                { label: 'Rights', onClick: () => {} },
                { separator: true, label: '' },
                { label: 'List', onClick: () => {} },
                {
                    label: 'View',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'Details', onClick: () => {} },
                        { label: 'Summary', onClick: () => {} },
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
                { label: 'Request Real Account', onClick: () => {} },
                { separator: true, label: '' },
                { label: 'Deposit / Withdrawal', onClick: () => {} },
                { label: 'Credit Facility', onClick: () => {} },
                { separator: true, label: '' },
                { label: 'Change Password', onClick: () => {} },
                { label: 'Change Group', onClick: () => {} },
                { separator: true, label: '' },
                {
                    label: 'Reports',
                    hasSubmenu: true,
                    submenu: [
                        { label: 'Statement', onClick: () => {} },
                        { label: 'Deals', onClick: () => {} },
                        { label: 'Positions', onClick: () => {} },
                    ]
                },
                { label: 'Delete Account', danger: true, onClick: () => {} },
            ];
        }

        // GENERIC
        return [
            { label: 'Open', onClick: () => {} },
            { label: 'Refresh', onClick: () => {} },
            { separator: true, label: '' },
            { label: 'Properties', onClick: () => {} },
        ];
    };

    const renderTree = (nodes: TreeNode[], depth = 0) => {
        // Use filteredData instead of nodes
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
                {renderTree(filteredData)}
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
