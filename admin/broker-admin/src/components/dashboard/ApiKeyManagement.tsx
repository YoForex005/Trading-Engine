'use client';

import React, { useState, useMemo } from 'react';
import {
    Key,
    Plus,
    Trash2,
    Search,
    X,
    Copy,
    Check,
    AlertTriangle,
    Shield,
    Eye,
    EyeOff,
    Clock,
    Activity,
    Zap,
    ExternalLink,
    Send,
    CheckCircle,
    XCircle,
    Globe
} from 'lucide-react';

// Types
type PermissionType =
    | 'read_accounts'
    | 'write_accounts'
    | 'read_trades'
    | 'write_trades'
    | 'read_symbols'
    | 'admin_settings'
    | 'webhooks';

type WebhookEvent =
    | 'trade.opened'
    | 'trade.closed'
    | 'deposit'
    | 'withdrawal'
    | 'account.created'
    | 'margin_call';

type ApiKeyStatus = 'active' | 'revoked' | 'expired';
type WebhookStatus = 'active' | 'disabled' | 'error';

interface ApiKey {
    id: string;
    name: string;
    description: string;
    maskedKey: string; // e.g., "sk_live_***abc123"
    createdAt: Date;
    lastUsed: Date | null;
    permissions: PermissionType[];
    status: ApiKeyStatus;
    expiresAt: Date | null;
    usage: {
        requests24h: number;
        requests7d: number;
        requests30d: number;
    };
    rateLimit: {
        limit: number;
        window: string;
    };
    ipWhitelist: string[];
}

interface Webhook {
    id: string;
    name: string;
    url: string;
    events: WebhookEvent[];
    status: WebhookStatus;
    lastTriggered: Date | null;
    successRate: number; // percentage
    secret: string; // masked
    deliveries: WebhookDelivery[];
}

interface WebhookDelivery {
    id: string;
    timestamp: Date;
    event: WebhookEvent;
    statusCode: number;
    responseTime: number; // ms
    success: boolean;
    payload: any;
    response: string;
}

interface CreateApiKeyForm {
    name: string;
    description: string;
    permissions: PermissionType[];
    expiry: 'never' | '30d' | '90d' | '1y';
}

interface CreateWebhookForm {
    name: string;
    url: string;
    events: WebhookEvent[];
}

const PERMISSION_LABELS: Record<PermissionType, string> = {
    read_accounts: 'Read Accounts',
    write_accounts: 'Write Accounts',
    read_trades: 'Read Trades',
    write_trades: 'Write Trades',
    read_symbols: 'Read Symbols',
    admin_settings: 'Admin Settings',
    webhooks: 'Webhooks',
};

const EVENT_LABELS: Record<WebhookEvent, string> = {
    'trade.opened': 'Trade Opened',
    'trade.closed': 'Trade Closed',
    'deposit': 'Deposit',
    'withdrawal': 'Withdrawal',
    'account.created': 'Account Created',
    'margin_call': 'Margin Call',
};

export default function ApiKeyManagement() {
    const [apiKeys, setApiKeys] = useState<ApiKey[]>(() => generateMockApiKeys());
    const [webhooks, setWebhooks] = useState<Webhook[]>(() => generateMockWebhooks());
    const [activeTab, setActiveTab] = useState<'api-keys' | 'webhooks'>('api-keys');

    // Modals
    const [showCreateKeyModal, setShowCreateKeyModal] = useState(false);
    const [showKeyDetailModal, setShowKeyDetailModal] = useState(false);
    const [showCreateWebhookModal, setShowCreateWebhookModal] = useState(false);
    const [showWebhookDetailModal, setShowWebhookDetailModal] = useState(false);
    const [selectedKey, setSelectedKey] = useState<ApiKey | null>(null);
    const [selectedWebhook, setSelectedWebhook] = useState<Webhook | null>(null);

    // Forms
    const [createKeyForm, setCreateKeyForm] = useState<CreateApiKeyForm>({
        name: '',
        description: '',
        permissions: [],
        expiry: 'never',
    });
    const [createWebhookForm, setCreateWebhookForm] = useState<CreateWebhookForm>({
        name: '',
        url: '',
        events: [],
    });

    // Search
    const [searchQuery, setSearchQuery] = useState('');
    const [copiedKey, setCopiedKey] = useState<string | null>(null);
    const [revealedKey, setRevealedKey] = useState<string | null>(null);

    // Calculate stats
    const stats = useMemo(() => {
        const activeKeys = apiKeys.filter(k => k.status === 'active').length;
        const totalApiCallsToday = apiKeys.reduce((sum, k) => sum + k.usage.requests24h, 0);
        const webhookDeliveriesToday = webhooks.reduce((sum, w) => {
            const today = new Date();
            today.setHours(0, 0, 0, 0);
            return sum + w.deliveries.filter(d => d.timestamp >= today).length;
        }, 0);
        const failedDeliveries = webhooks.reduce((sum, w) => {
            return sum + w.deliveries.filter(d => !d.success).length;
        }, 0);

        return {
            activeKeys,
            totalApiCallsToday,
            webhookDeliveriesToday,
            failedDeliveries,
        };
    }, [apiKeys, webhooks]);

    // Filter keys/webhooks
    const filteredKeys = useMemo(() => {
        return apiKeys.filter(key =>
            key.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            key.description.toLowerCase().includes(searchQuery.toLowerCase())
        );
    }, [apiKeys, searchQuery]);

    const filteredWebhooks = useMemo(() => {
        return webhooks.filter(webhook =>
            webhook.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            webhook.url.toLowerCase().includes(searchQuery.toLowerCase())
        );
    }, [webhooks, searchQuery]);

    const handleCopyKey = (key: string) => {
        navigator.clipboard.writeText(key);
        setCopiedKey(key);
        setTimeout(() => setCopiedKey(null), 2000);
    };

    const handleCreateApiKey = () => {
        // In real app, this would call API
        const newKey: ApiKey = {
            id: `key-${Date.now()}`,
            name: createKeyForm.name,
            description: createKeyForm.description,
            maskedKey: `sk_live_***${Math.random().toString(36).substr(2, 7)}`,
            createdAt: new Date(),
            lastUsed: null,
            permissions: createKeyForm.permissions,
            status: 'active',
            expiresAt: createKeyForm.expiry === 'never' ? null : new Date(Date.now() + parseInt(createKeyForm.expiry) * 24 * 60 * 60 * 1000),
            usage: { requests24h: 0, requests7d: 0, requests30d: 0 },
            rateLimit: { limit: 1000, window: '1 hour' },
            ipWhitelist: [],
        };
        setApiKeys([...apiKeys, newKey]);
        setShowCreateKeyModal(false);
        setCreateKeyForm({ name: '', description: '', permissions: [], expiry: 'never' });
    };

    const handleCreateWebhook = () => {
        const newWebhook: Webhook = {
            id: `webhook-${Date.now()}`,
            name: createWebhookForm.name,
            url: createWebhookForm.url,
            events: createWebhookForm.events,
            status: 'active',
            lastTriggered: null,
            successRate: 100,
            secret: `whsec_***${Math.random().toString(36).substr(2, 10)}`,
            deliveries: [],
        };
        setWebhooks([...webhooks, newWebhook]);
        setShowCreateWebhookModal(false);
        setCreateWebhookForm({ name: '', url: '', events: [] });
    };

    const handleTestWebhook = (webhookId: string) => {
        const webhook = webhooks.find(w => w.id === webhookId);
        if (!webhook) return;

        const testDelivery: WebhookDelivery = {
            id: `delivery-${Date.now()}`,
            timestamp: new Date(),
            event: 'trade.opened',
            statusCode: 200,
            responseTime: Math.floor(Math.random() * 500) + 100,
            success: true,
            payload: { type: 'test', message: 'Test webhook delivery' },
            response: 'OK',
        };

        const updatedWebhook = {
            ...webhook,
            lastTriggered: new Date(),
            deliveries: [testDelivery, ...webhook.deliveries],
        };

        setWebhooks(webhooks.map(w => w.id === webhookId ? updatedWebhook : w));
    };

    const handleRevokeKey = (keyId: string) => {
        setApiKeys(apiKeys.map(k => k.id === keyId ? { ...k, status: 'revoked' as ApiKeyStatus } : k));
    };

    const handleDeleteWebhook = (webhookId: string) => {
        setWebhooks(webhooks.filter(w => w.id !== webhookId));
    };

    return (
        <div className="h-full flex flex-col bg-[#121316] text-[#E5E7EB]">
            {/* Header */}
            <div className="flex-shrink-0 px-6 py-4 border-b border-[#383A42]">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <Key className="text-[#3B82F6]" size={24} />
                        <h1 className="text-xl font-bold text-[#F5C542]">API Keys & Webhooks</h1>
                    </div>
                    <button
                        onClick={() => activeTab === 'api-keys' ? setShowCreateKeyModal(true) : setShowCreateWebhookModal(true)}
                        className="flex items-center gap-2 px-4 py-2 bg-[#3B82F6] text-white rounded hover:bg-[#2563EB] transition-colors text-sm font-medium"
                    >
                        <Plus size={16} />
                        {activeTab === 'api-keys' ? 'Create API Key' : 'Create Webhook'}
                    </button>
                </div>

                {/* Stats Cards */}
                <div className="grid grid-cols-4 gap-4">
                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Active Keys</span>
                            <Key size={16} className="text-[#3B82F6]" />
                        </div>
                        <div className="text-2xl font-bold text-[#F5C542]">{stats.activeKeys}</div>
                    </div>

                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">API Calls Today</span>
                            <Activity size={16} className="text-[#10B981]" />
                        </div>
                        <div className="text-2xl font-bold text-[#F5C542]">{stats.totalApiCallsToday.toLocaleString()}</div>
                    </div>

                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Webhook Deliveries Today</span>
                            <Zap size={16} className="text-[#A855F7]" />
                        </div>
                        <div className="text-2xl font-bold text-[#F5C542]">{stats.webhookDeliveriesToday}</div>
                    </div>

                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Failed Deliveries</span>
                            <AlertTriangle size={16} className="text-[#EF4444]" />
                        </div>
                        <div className="text-2xl font-bold text-[#EF4444]">{stats.failedDeliveries}</div>
                    </div>
                </div>
            </div>

            {/* Tabs */}
            <div className="flex-shrink-0 px-6 py-3 bg-[#1E2026] border-b border-[#383A42] flex items-center gap-4">
                <button
                    onClick={() => setActiveTab('api-keys')}
                    className={`px-4 py-2 rounded text-sm font-medium transition-colors ${
                        activeTab === 'api-keys'
                            ? 'bg-[#3B82F6] text-white'
                            : 'bg-[#121316] text-[#888] hover:text-[#E5E7EB]'
                    }`}
                >
                    <div className="flex items-center gap-2">
                        <Key size={14} />
                        <span>API Keys ({apiKeys.length})</span>
                    </div>
                </button>
                <button
                    onClick={() => setActiveTab('webhooks')}
                    className={`px-4 py-2 rounded text-sm font-medium transition-colors ${
                        activeTab === 'webhooks'
                            ? 'bg-[#3B82F6] text-white'
                            : 'bg-[#121316] text-[#888] hover:text-[#E5E7EB]'
                    }`}
                >
                    <div className="flex items-center gap-2">
                        <Zap size={14} />
                        <span>Webhooks ({webhooks.length})</span>
                    </div>
                </button>

                <div className="ml-auto flex items-center gap-2 bg-[#121316] border border-[#383A42] rounded px-3 py-2 flex-1 max-w-md">
                    <Search size={14} className="text-[#888]" />
                    <input
                        type="text"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="bg-transparent text-xs outline-none flex-1 text-[#E5E7EB]"
                        placeholder={`Search ${activeTab === 'api-keys' ? 'API keys' : 'webhooks'}...`}
                    />
                </div>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-auto">
                {activeTab === 'api-keys' ? (
                    /* API Keys Table */
                    <table className="w-full text-xs">
                        <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                            <tr>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Name</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Key</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Permissions</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Created</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Last Used</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Status</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredKeys.map((key) => (
                                <tr
                                    key={key.id}
                                    className="border-b border-[#383A42] hover:bg-[#1E2026] cursor-pointer transition-colors"
                                    onClick={() => {
                                        setSelectedKey(key);
                                        setShowKeyDetailModal(true);
                                    }}
                                >
                                    <td className="py-2 px-4">
                                        <div>
                                            <div className="text-[#E5E7EB] font-medium">{key.name}</div>
                                            <div className="text-[#888] text-xs">{key.description}</div>
                                        </div>
                                    </td>
                                    <td className="py-2 px-4">
                                        <div className="flex items-center gap-2">
                                            <code className="text-[#888] font-mono">
                                                {revealedKey === key.id ? key.maskedKey.replace('***', 'live_actual_key') : key.maskedKey}
                                            </code>
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleCopyKey(key.maskedKey);
                                                }}
                                                className="p-1 hover:bg-[#383A42] rounded transition-colors"
                                            >
                                                {copiedKey === key.maskedKey ? (
                                                    <Check size={12} className="text-[#10B981]" />
                                                ) : (
                                                    <Copy size={12} className="text-[#888]" />
                                                )}
                                            </button>
                                        </div>
                                    </td>
                                    <td className="py-2 px-4">
                                        <div className="flex flex-wrap gap-1">
                                            {key.permissions.slice(0, 2).map((perm) => (
                                                <span
                                                    key={perm}
                                                    className="px-2 py-0.5 bg-[#3B82F6]/10 text-[#3B82F6] rounded text-xs"
                                                >
                                                    {PERMISSION_LABELS[perm]}
                                                </span>
                                            ))}
                                            {key.permissions.length > 2 && (
                                                <span className="px-2 py-0.5 bg-[#888]/10 text-[#888] rounded text-xs">
                                                    +{key.permissions.length - 2}
                                                </span>
                                            )}
                                        </div>
                                    </td>
                                    <td className="py-2 px-4 text-[#888]">
                                        {key.createdAt.toLocaleDateString()}
                                    </td>
                                    <td className="py-2 px-4 text-[#888]">
                                        {key.lastUsed ? key.lastUsed.toLocaleDateString() : 'Never'}
                                    </td>
                                    <td className="py-2 px-4">
                                        <span
                                            className={`px-2 py-1 rounded text-xs font-medium ${
                                                key.status === 'active'
                                                    ? 'bg-[#10B981]/10 text-[#10B981]'
                                                    : key.status === 'expired'
                                                    ? 'bg-[#F59E0B]/10 text-[#F59E0B]'
                                                    : 'bg-[#EF4444]/10 text-[#EF4444]'
                                            }`}
                                        >
                                            {key.status.toUpperCase()}
                                        </span>
                                    </td>
                                    <td className="py-2 px-4">
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleRevokeKey(key.id);
                                            }}
                                            className="p-1 hover:bg-[#EF4444]/10 rounded transition-colors"
                                            disabled={key.status === 'revoked'}
                                        >
                                            <Trash2 size={14} className="text-[#EF4444]" />
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                ) : (
                    /* Webhooks Table */
                    <table className="w-full text-xs">
                        <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                            <tr>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Name</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">URL</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Events</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Status</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Last Triggered</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Success Rate</th>
                                <th className="text-left py-2 px-4 font-semibold text-[#888]">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredWebhooks.map((webhook) => (
                                <tr
                                    key={webhook.id}
                                    className="border-b border-[#383A42] hover:bg-[#1E2026] cursor-pointer transition-colors"
                                    onClick={() => {
                                        setSelectedWebhook(webhook);
                                        setShowWebhookDetailModal(true);
                                    }}
                                >
                                    <td className="py-2 px-4">
                                        <div className="text-[#E5E7EB] font-medium">{webhook.name}</div>
                                    </td>
                                    <td className="py-2 px-4">
                                        <div className="flex items-center gap-2">
                                            <Globe size={12} className="text-[#888]" />
                                            <code className="text-[#888] font-mono text-xs truncate max-w-xs">
                                                {webhook.url}
                                            </code>
                                        </div>
                                    </td>
                                    <td className="py-2 px-4">
                                        <div className="flex flex-wrap gap-1">
                                            {webhook.events.slice(0, 2).map((event) => (
                                                <span
                                                    key={event}
                                                    className="px-2 py-0.5 bg-[#A855F7]/10 text-[#A855F7] rounded text-xs"
                                                >
                                                    {EVENT_LABELS[event]}
                                                </span>
                                            ))}
                                            {webhook.events.length > 2 && (
                                                <span className="px-2 py-0.5 bg-[#888]/10 text-[#888] rounded text-xs">
                                                    +{webhook.events.length - 2}
                                                </span>
                                            )}
                                        </div>
                                    </td>
                                    <td className="py-2 px-4">
                                        <span
                                            className={`px-2 py-1 rounded text-xs font-medium ${
                                                webhook.status === 'active'
                                                    ? 'bg-[#10B981]/10 text-[#10B981]'
                                                    : webhook.status === 'disabled'
                                                    ? 'bg-[#888]/10 text-[#888]'
                                                    : 'bg-[#EF4444]/10 text-[#EF4444]'
                                            }`}
                                        >
                                            {webhook.status.toUpperCase()}
                                        </span>
                                    </td>
                                    <td className="py-2 px-4 text-[#888]">
                                        {webhook.lastTriggered ? webhook.lastTriggered.toLocaleString() : 'Never'}
                                    </td>
                                    <td className="py-2 px-4">
                                        <div className="flex items-center gap-2">
                                            <div className="w-16 bg-[#383A42] rounded-full h-1.5">
                                                <div
                                                    className={`h-full rounded-full ${
                                                        webhook.successRate >= 90
                                                            ? 'bg-[#10B981]'
                                                            : webhook.successRate >= 70
                                                            ? 'bg-[#F59E0B]'
                                                            : 'bg-[#EF4444]'
                                                    }`}
                                                    style={{ width: `${webhook.successRate}%` }}
                                                />
                                            </div>
                                            <span className="text-[#888] text-xs">{webhook.successRate}%</span>
                                        </div>
                                    </td>
                                    <td className="py-2 px-4">
                                        <div className="flex items-center gap-2">
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleTestWebhook(webhook.id);
                                                }}
                                                className="p-1 hover:bg-[#3B82F6]/10 rounded transition-colors"
                                            >
                                                <Send size={14} className="text-[#3B82F6]" />
                                            </button>
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleDeleteWebhook(webhook.id);
                                                }}
                                                className="p-1 hover:bg-[#EF4444]/10 rounded transition-colors"
                                            >
                                                <Trash2 size={14} className="text-[#EF4444]" />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}

                {(activeTab === 'api-keys' ? filteredKeys : filteredWebhooks).length === 0 && (
                    <div className="flex flex-col items-center justify-center py-16 text-[#666]">
                        <Key size={48} className="mb-4 opacity-20" />
                        <div className="text-lg">No {activeTab === 'api-keys' ? 'API keys' : 'webhooks'} found</div>
                        <div className="text-xs mt-2">Create one to get started</div>
                    </div>
                )}
            </div>

            {/* Create API Key Modal */}
            {showCreateKeyModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-auto">
                        <div className="flex items-center justify-between mb-4">
                            <h2 className="text-lg font-bold text-[#F5C542]">Create API Key</h2>
                            <button
                                onClick={() => setShowCreateKeyModal(false)}
                                className="p-1 hover:bg-[#383A42] rounded transition-colors"
                            >
                                <X size={20} className="text-[#888]" />
                            </button>
                        </div>

                        <div className="space-y-4">
                            <div>
                                <label className="block text-xs font-medium text-[#888] mb-1">Name</label>
                                <input
                                    type="text"
                                    value={createKeyForm.name}
                                    onChange={(e) => setCreateKeyForm({ ...createKeyForm, name: e.target.value })}
                                    className="w-full bg-[#121316] border border-[#383A42] rounded px-3 py-2 text-sm text-[#E5E7EB] outline-none focus:border-[#3B82F6]"
                                    placeholder="Production API Key"
                                />
                            </div>

                            <div>
                                <label className="block text-xs font-medium text-[#888] mb-1">Description</label>
                                <textarea
                                    value={createKeyForm.description}
                                    onChange={(e) => setCreateKeyForm({ ...createKeyForm, description: e.target.value })}
                                    className="w-full bg-[#121316] border border-[#383A42] rounded px-3 py-2 text-sm text-[#E5E7EB] outline-none focus:border-[#3B82F6]"
                                    rows={2}
                                    placeholder="API key for production trading system"
                                />
                            </div>

                            <div>
                                <label className="block text-xs font-medium text-[#888] mb-2">Permissions</label>
                                <div className="grid grid-cols-2 gap-2">
                                    {Object.entries(PERMISSION_LABELS).map(([perm, label]) => (
                                        <label
                                            key={perm}
                                            className="flex items-center gap-2 p-2 bg-[#121316] border border-[#383A42] rounded cursor-pointer hover:border-[#3B82F6] transition-colors"
                                        >
                                            <input
                                                type="checkbox"
                                                checked={createKeyForm.permissions.includes(perm as PermissionType)}
                                                onChange={(e) => {
                                                    const perms = e.target.checked
                                                        ? [...createKeyForm.permissions, perm as PermissionType]
                                                        : createKeyForm.permissions.filter((p) => p !== perm);
                                                    setCreateKeyForm({ ...createKeyForm, permissions: perms });
                                                }}
                                                className="w-4 h-4"
                                            />
                                            <span className="text-xs text-[#E5E7EB]">{label}</span>
                                        </label>
                                    ))}
                                </div>
                            </div>

                            <div>
                                <label className="block text-xs font-medium text-[#888] mb-1">Expiry</label>
                                <select
                                    value={createKeyForm.expiry}
                                    onChange={(e) => setCreateKeyForm({ ...createKeyForm, expiry: e.target.value as any })}
                                    className="w-full bg-[#121316] border border-[#383A42] rounded px-3 py-2 text-sm text-[#E5E7EB] outline-none focus:border-[#3B82F6]"
                                >
                                    <option value="never">Never</option>
                                    <option value="30d">30 days</option>
                                    <option value="90d">90 days</option>
                                    <option value="1y">1 year</option>
                                </select>
                            </div>

                            <div className="flex gap-2 mt-6">
                                <button
                                    onClick={handleCreateApiKey}
                                    disabled={!createKeyForm.name || createKeyForm.permissions.length === 0}
                                    className="flex-1 px-4 py-2 bg-[#3B82F6] text-white rounded hover:bg-[#2563EB] transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    Create Key
                                </button>
                                <button
                                    onClick={() => setShowCreateKeyModal(false)}
                                    className="px-4 py-2 bg-[#383A42] text-[#E5E7EB] rounded hover:bg-[#4B5563] transition-colors text-sm font-medium"
                                >
                                    Cancel
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* Create Webhook Modal */}
            {showCreateWebhookModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-auto">
                        <div className="flex items-center justify-between mb-4">
                            <h2 className="text-lg font-bold text-[#F5C542]">Create Webhook</h2>
                            <button
                                onClick={() => setShowCreateWebhookModal(false)}
                                className="p-1 hover:bg-[#383A42] rounded transition-colors"
                            >
                                <X size={20} className="text-[#888]" />
                            </button>
                        </div>

                        <div className="space-y-4">
                            <div>
                                <label className="block text-xs font-medium text-[#888] mb-1">Name</label>
                                <input
                                    type="text"
                                    value={createWebhookForm.name}
                                    onChange={(e) => setCreateWebhookForm({ ...createWebhookForm, name: e.target.value })}
                                    className="w-full bg-[#121316] border border-[#383A42] rounded px-3 py-2 text-sm text-[#E5E7EB] outline-none focus:border-[#3B82F6]"
                                    placeholder="Trading Events Webhook"
                                />
                            </div>

                            <div>
                                <label className="block text-xs font-medium text-[#888] mb-1">URL</label>
                                <input
                                    type="url"
                                    value={createWebhookForm.url}
                                    onChange={(e) => setCreateWebhookForm({ ...createWebhookForm, url: e.target.value })}
                                    className="w-full bg-[#121316] border border-[#383A42] rounded px-3 py-2 text-sm text-[#E5E7EB] outline-none focus:border-[#3B82F6] font-mono"
                                    placeholder="https://your-app.com/webhooks/trading"
                                />
                            </div>

                            <div>
                                <label className="block text-xs font-medium text-[#888] mb-2">Events</label>
                                <div className="grid grid-cols-2 gap-2">
                                    {Object.entries(EVENT_LABELS).map(([event, label]) => (
                                        <label
                                            key={event}
                                            className="flex items-center gap-2 p-2 bg-[#121316] border border-[#383A42] rounded cursor-pointer hover:border-[#3B82F6] transition-colors"
                                        >
                                            <input
                                                type="checkbox"
                                                checked={createWebhookForm.events.includes(event as WebhookEvent)}
                                                onChange={(e) => {
                                                    const events = e.target.checked
                                                        ? [...createWebhookForm.events, event as WebhookEvent]
                                                        : createWebhookForm.events.filter((ev) => ev !== event);
                                                    setCreateWebhookForm({ ...createWebhookForm, events });
                                                }}
                                                className="w-4 h-4"
                                            />
                                            <span className="text-xs text-[#E5E7EB]">{label}</span>
                                        </label>
                                    ))}
                                </div>
                            </div>

                            <div className="bg-[#121316] border border-[#F59E0B]/30 rounded p-3">
                                <div className="flex items-start gap-2">
                                    <Shield size={16} className="text-[#F59E0B] mt-0.5" />
                                    <div>
                                        <div className="text-xs font-medium text-[#F59E0B] mb-1">Secret Key</div>
                                        <div className="text-xs text-[#888]">
                                            A secret key will be auto-generated for signing webhook requests
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div className="flex gap-2 mt-6">
                                <button
                                    onClick={handleCreateWebhook}
                                    disabled={!createWebhookForm.name || !createWebhookForm.url || createWebhookForm.events.length === 0}
                                    className="flex-1 px-4 py-2 bg-[#3B82F6] text-white rounded hover:bg-[#2563EB] transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    Create Webhook
                                </button>
                                <button
                                    onClick={() => setShowCreateWebhookModal(false)}
                                    className="px-4 py-2 bg-[#383A42] text-[#E5E7EB] rounded hover:bg-[#4B5563] transition-colors text-sm font-medium"
                                >
                                    Cancel
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* API Key Detail Modal */}
            {showKeyDetailModal && selectedKey && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-6 w-full max-w-3xl max-h-[80vh] overflow-auto">
                        <div className="flex items-center justify-between mb-4">
                            <h2 className="text-lg font-bold text-[#F5C542]">{selectedKey.name}</h2>
                            <button
                                onClick={() => setShowKeyDetailModal(false)}
                                className="p-1 hover:bg-[#383A42] rounded transition-colors"
                            >
                                <X size={20} className="text-[#888]" />
                            </button>
                        </div>

                        <div className="space-y-4">
                            <div className="grid grid-cols-3 gap-4">
                                <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                                    <div className="text-xs text-[#888] mb-1">Requests (24h)</div>
                                    <div className="text-xl font-bold text-[#F5C542]">{selectedKey.usage.requests24h.toLocaleString()}</div>
                                </div>
                                <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                                    <div className="text-xs text-[#888] mb-1">Requests (7d)</div>
                                    <div className="text-xl font-bold text-[#F5C542]">{selectedKey.usage.requests7d.toLocaleString()}</div>
                                </div>
                                <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                                    <div className="text-xs text-[#888] mb-1">Requests (30d)</div>
                                    <div className="text-xl font-bold text-[#F5C542]">{selectedKey.usage.requests30d.toLocaleString()}</div>
                                </div>
                            </div>

                            <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                                <div className="text-xs font-medium text-[#888] mb-2">Rate Limit</div>
                                <div className="text-sm text-[#E5E7EB]">
                                    {selectedKey.rateLimit.limit.toLocaleString()} requests per {selectedKey.rateLimit.window}
                                </div>
                            </div>

                            <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                                <div className="text-xs font-medium text-[#888] mb-2">IP Whitelist</div>
                                {selectedKey.ipWhitelist.length > 0 ? (
                                    <div className="space-y-1">
                                        {selectedKey.ipWhitelist.map((ip) => (
                                            <div key={ip} className="text-sm text-[#E5E7EB] font-mono">{ip}</div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="text-sm text-[#888]">No IP restrictions</div>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* Webhook Detail Modal */}
            {showWebhookDetailModal && selectedWebhook && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-6 w-full max-w-4xl max-h-[80vh] overflow-auto">
                        <div className="flex items-center justify-between mb-4">
                            <h2 className="text-lg font-bold text-[#F5C542]">{selectedWebhook.name}</h2>
                            <button
                                onClick={() => setShowWebhookDetailModal(false)}
                                className="p-1 hover:bg-[#383A42] rounded transition-colors"
                            >
                                <X size={20} className="text-[#888]" />
                            </button>
                        </div>

                        <div className="space-y-4">
                            <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                                <div className="text-xs font-medium text-[#888] mb-1">URL</div>
                                <code className="text-sm text-[#E5E7EB] font-mono">{selectedWebhook.url}</code>
                            </div>

                            <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                                <div className="text-xs font-medium text-[#888] mb-1">Secret Key</div>
                                <code className="text-sm text-[#888] font-mono">{selectedWebhook.secret}</code>
                            </div>

                            <div>
                                <div className="text-xs font-medium text-[#888] mb-2">Recent Deliveries</div>
                                <div className="space-y-2">
                                    {selectedWebhook.deliveries.slice(0, 10).map((delivery) => (
                                        <div
                                            key={delivery.id}
                                            className="bg-[#121316] border border-[#383A42] rounded p-3 flex items-center justify-between"
                                        >
                                            <div className="flex items-center gap-3">
                                                {delivery.success ? (
                                                    <CheckCircle size={16} className="text-[#10B981]" />
                                                ) : (
                                                    <XCircle size={16} className="text-[#EF4444]" />
                                                )}
                                                <div>
                                                    <div className="text-sm text-[#E5E7EB]">{EVENT_LABELS[delivery.event]}</div>
                                                    <div className="text-xs text-[#888]">
                                                        {delivery.timestamp.toLocaleString()}
                                                    </div>
                                                </div>
                                            </div>
                                            <div className="text-right">
                                                <div className="text-sm text-[#E5E7EB]">
                                                    <span className={delivery.success ? 'text-[#10B981]' : 'text-[#EF4444]'}>
                                                        {delivery.statusCode}
                                                    </span>
                                                </div>
                                                <div className="text-xs text-[#888]">{delivery.responseTime}ms</div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

// Mock data generators
function generateMockApiKeys(): ApiKey[] {
    const keys: ApiKey[] = [];
    const now = new Date();

    const keyData = [
        { name: 'Production API Key', desc: 'Main production trading system', perms: ['read_accounts', 'write_accounts', 'read_trades', 'write_trades'] as PermissionType[] },
        { name: 'Analytics Service', desc: 'Read-only analytics platform', perms: ['read_accounts', 'read_trades'] as PermissionType[] },
        { name: 'Mobile App', desc: 'Mobile trading application', perms: ['read_accounts', 'read_trades', 'write_trades'] as PermissionType[] },
        { name: 'Reporting Tool', desc: 'Internal reporting dashboard', perms: ['read_accounts', 'read_trades', 'read_symbols'] as PermissionType[] },
        { name: 'Admin Panel', desc: 'Administrative interface', perms: ['admin_settings', 'read_accounts', 'write_accounts'] as PermissionType[] },
        { name: 'Testing Environment', desc: 'Development testing key', perms: ['read_accounts', 'read_trades'] as PermissionType[] },
        { name: 'Risk Management', desc: 'Risk monitoring system', perms: ['read_accounts', 'read_trades'] as PermissionType[] },
        { name: 'Webhook Manager', desc: 'Webhook configuration service', perms: ['webhooks'] as PermissionType[] },
        { name: 'Legacy System', desc: 'Deprecated legacy integration', perms: ['read_accounts'] as PermissionType[] },
        { name: 'Partner Integration', desc: 'Third-party partner API', perms: ['read_accounts', 'read_trades', 'read_symbols'] as PermissionType[] },
    ];

    keyData.forEach((data, i) => {
        keys.push({
            id: `key-${i}`,
            name: data.name,
            description: data.desc,
            maskedKey: `sk_live_***${Math.random().toString(36).substr(2, 7)}`,
            createdAt: new Date(now.getTime() - Math.random() * 180 * 24 * 60 * 60 * 1000),
            lastUsed: i < 7 ? new Date(now.getTime() - Math.random() * 7 * 24 * 60 * 60 * 1000) : null,
            permissions: data.perms,
            status: i === 8 ? 'revoked' : i === 9 ? 'expired' : 'active',
            expiresAt: i === 9 ? new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000) : null,
            usage: {
                requests24h: Math.floor(Math.random() * 10000),
                requests7d: Math.floor(Math.random() * 50000),
                requests30d: Math.floor(Math.random() * 200000),
            },
            rateLimit: { limit: 1000, window: '1 hour' },
            ipWhitelist: i < 3 ? [`192.168.1.${i}`, `10.0.0.${i}`] : [],
        });
    });

    return keys;
}

function generateMockWebhooks(): Webhook[] {
    const webhooks: Webhook[] = [];
    const now = new Date();

    const webhookData = [
        { name: 'Trading Events', url: 'https://api.example.com/webhooks/trading', events: ['trade.opened', 'trade.closed'] as WebhookEvent[] },
        { name: 'Deposit Notifications', url: 'https://notify.example.com/deposits', events: ['deposit'] as WebhookEvent[] },
        { name: 'Withdrawal Alerts', url: 'https://alerts.example.com/withdrawals', events: ['withdrawal'] as WebhookEvent[] },
        { name: 'Account Events', url: 'https://crm.example.com/webhooks/accounts', events: ['account.created'] as WebhookEvent[] },
        { name: 'Risk Alerts', url: 'https://risk.example.com/margin-calls', events: ['margin_call'] as WebhookEvent[] },
        { name: 'All Events Monitor', url: 'https://monitor.example.com/all', events: ['trade.opened', 'trade.closed', 'deposit', 'withdrawal', 'account.created', 'margin_call'] as WebhookEvent[] },
        { name: 'Slack Notifications', url: 'https://hooks.slack.com/services/xxx', events: ['margin_call', 'withdrawal'] as WebhookEvent[] },
        { name: 'Legacy Integration', url: 'https://old.example.com/webhook', events: ['trade.opened'] as WebhookEvent[] },
    ];

    webhookData.forEach((data, i) => {
        const deliveries: WebhookDelivery[] = [];
        for (let j = 0; j < 20; j++) {
            const isSuccess = Math.random() > 0.1;
            deliveries.push({
                id: `delivery-${i}-${j}`,
                timestamp: new Date(now.getTime() - j * 60 * 60 * 1000),
                event: data.events[Math.floor(Math.random() * data.events.length)],
                statusCode: isSuccess ? 200 : [400, 500, 503][Math.floor(Math.random() * 3)],
                responseTime: Math.floor(Math.random() * 1000) + 100,
                success: isSuccess,
                payload: { type: 'test', id: `test-${j}` },
                response: isSuccess ? 'OK' : 'Error',
            });
        }

        webhooks.push({
            id: `webhook-${i}`,
            name: data.name,
            url: data.url,
            events: data.events,
            status: i === 7 ? 'disabled' : i === 6 ? 'error' : 'active',
            lastTriggered: i < 6 ? new Date(now.getTime() - Math.random() * 24 * 60 * 60 * 1000) : null,
            successRate: Math.floor(90 + Math.random() * 10),
            secret: `whsec_***${Math.random().toString(36).substr(2, 10)}`,
            deliveries,
        });
    });

    return webhooks;
}
