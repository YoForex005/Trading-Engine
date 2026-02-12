'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
    Webhook,
    Plus,
    Edit,
    Trash2,
    Play,
    Pause,
    Search,
    X,
    Copy,
    Check,
    AlertTriangle,
    Clock,
    CheckCircle,
    XCircle,
    RefreshCw,
    ChevronDown,
    ChevronRight,
    Code,
} from 'lucide-react';
import { api } from '@/services/apiClient';
import { API_CONFIG } from '@/config/api';

type WebhookEvent =
    | 'trade.opened'
    | 'trade.closed'
    | 'order.placed'
    | 'order.cancelled'
    | 'deposit.received'
    | 'withdrawal.approved'
    | 'margin.warning'
    | 'margin.stopout';

type WebhookStatus = 'active' | 'paused' | 'failed';
type DeliveryStatus = 'success' | 'failed' | 'pending';

interface WebhookHeader {
    key: string;
    value: string;
}

interface DeliveryLogEntry {
    id: string;
    timestamp: string;
    event: WebhookEvent;
    statusCode: number | null;
    responseTime: number | null; // milliseconds
    attempt: number;
    status: DeliveryStatus;
    requestPayload: any;
    responsePayload: any;
}

interface Webhook {
    id: string;
    name: string;
    url: string;
    secret: string; // Only shown once when creating
    events: WebhookEvent[];
    status: WebhookStatus;
    isActive: boolean;
    lastTriggered: string | null;
    successRate: number; // 0-100
    headers: WebhookHeader[];
    deliveryLog: DeliveryLogEntry[];
}

interface WebhookFormData {
    name: string;
    url: string;
    events: WebhookEvent[];
    isActive: boolean;
    headers: WebhookHeader[];
}

const WEBHOOK_EVENTS: { value: WebhookEvent; label: string; description: string }[] = [
    { value: 'trade.opened', label: 'Trade Opened', description: 'When a new trade is executed' },
    { value: 'trade.closed', label: 'Trade Closed', description: 'When a trade is closed' },
    { value: 'order.placed', label: 'Order Placed', description: 'When a pending order is placed' },
    { value: 'order.cancelled', label: 'Order Cancelled', description: 'When an order is cancelled' },
    { value: 'deposit.received', label: 'Deposit Received', description: 'When a client deposits funds' },
    { value: 'withdrawal.approved', label: 'Withdrawal Approved', description: 'When a withdrawal is approved' },
    { value: 'margin.warning', label: 'Margin Warning', description: 'When margin level reaches warning threshold' },
    { value: 'margin.stopout', label: 'Margin Stopout', description: 'When margin stopout occurs' },
];

const STATUS_COLORS: Record<WebhookStatus, string> = {
    active: 'bg-green-600 text-white',
    paused: 'bg-yellow-600 text-black',
    failed: 'bg-red-600 text-white',
};

// Mock data generator
const generateMockWebhooks = (): Webhook[] => {
    const names = [
        'Production Trading Bot',
        'Risk Alert System',
        'Accounting Integration',
        'CRM Sync',
        'Mobile Push Notifications',
    ];

    const urls = [
        'https://api.example.com/webhooks/trading',
        'https://risk.company.com/alerts',
        'https://accounting.internal/sync',
        'https://crm.example.com/hooks/rtx5',
        'https://push.mobile.app/notify',
    ];

    const eventCombos: WebhookEvent[][] = [
        ['trade.opened', 'trade.closed'],
        ['margin.warning', 'margin.stopout'],
        ['deposit.received', 'withdrawal.approved'],
        ['order.placed', 'order.cancelled'],
        ['trade.opened', 'trade.closed', 'order.placed', 'order.cancelled'],
    ];

    return names.map((name, i) => {
        const deliveryLog: DeliveryLogEntry[] = Array.from({ length: 25 }, (_, j) => {
            const hoursAgo = Math.floor(Math.random() * 72);
            const timestamp = new Date(Date.now() - hoursAgo * 60 * 60 * 1000).toISOString();
            const statusCode = Math.random() > 0.2 ? 200 : Math.random() > 0.5 ? 500 : null;
            const status: DeliveryStatus = statusCode === 200 ? 'success' : statusCode ? 'failed' : 'pending';

            return {
                id: `log-${i}-${j}`,
                timestamp,
                event: eventCombos[i][Math.floor(Math.random() * eventCombos[i].length)],
                statusCode,
                responseTime: statusCode ? Math.floor(Math.random() * 500) + 50 : null,
                attempt: Math.floor(Math.random() * 3) + 1,
                status,
                requestPayload: {
                    event: eventCombos[i][0],
                    data: { tradeId: `T${j}`, symbol: 'EURUSD', volume: 1.0 },
                },
                responsePayload: statusCode === 200 ? { success: true } : statusCode ? { error: 'Internal server error' } : null,
            };
        });

        const successCount = deliveryLog.filter(l => l.status === 'success').length;
        const successRate = (successCount / deliveryLog.length) * 100;

        const statuses: WebhookStatus[] = ['active', 'active', 'active', 'paused', 'failed'];
        const status = statuses[i];

        const lastTriggered = deliveryLog[0]?.timestamp || null;

        return {
            id: `webhook-${i}`,
            name,
            url: urls[i],
            secret: `whsec_${Math.random().toString(36).substring(2, 18)}`,
            events: eventCombos[i],
            status,
            isActive: status === 'active',
            lastTriggered,
            successRate,
            headers: i === 2 ? [{ key: 'X-API-Key', value: 'secret-key-123' }] : [],
            deliveryLog,
        };
    });
};

export default function WebhookManagement() {
    const [webhooks, setWebhooks] = useState<Webhook[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showEditModal, setShowEditModal] = useState(false);
    const [showSecretModal, setShowSecretModal] = useState(false);
    const [showTestModal, setShowTestModal] = useState(false);
    const [showPayloadModal, setShowPayloadModal] = useState(false);
    const [editingWebhook, setEditingWebhook] = useState<Webhook | null>(null);
    const [testingWebhook, setTestingWebhook] = useState<Webhook | null>(null);
    const [newSecret, setNewSecret] = useState('');
    const [copied, setCopied] = useState(false);
    const [expandedRow, setExpandedRow] = useState<string | null>(null);
    const [selectedLog, setSelectedLog] = useState<DeliveryLogEntry | null>(null);
    const [testEvent, setTestEvent] = useState<WebhookEvent>('trade.opened');
    const [testResult, setTestResult] = useState<any>(null);

    const [formData, setFormData] = useState<WebhookFormData>({
        name: '',
        url: '',
        events: [],
        isActive: true,
        headers: [],
    });

    useEffect(() => {
        fetchWebhooks();
    }, []);

    const fetchWebhooks = async () => {
        try {
            setLoading(true);
            // In production: const data = await api.get<Webhook[]>(API_CONFIG.WEBHOOKS_LIST);
            await new Promise(resolve => setTimeout(resolve, 500));
            const mockData = generateMockWebhooks();
            setWebhooks(mockData);
        } catch (error) {
            console.error('[WebhookManagement] Failed to fetch webhooks:', error);
        } finally {
            setLoading(false);
        }
    };

    const filteredWebhooks = useMemo(() => {
        if (!searchQuery.trim()) return webhooks;
        const query = searchQuery.toLowerCase();
        return webhooks.filter(w =>
            w.name.toLowerCase().includes(query) ||
            w.url.toLowerCase().includes(query)
        );
    }, [webhooks, searchQuery]);

    const handleCreate = () => {
        setFormData({
            name: '',
            url: '',
            events: [],
            isActive: true,
            headers: [],
        });
        setShowCreateModal(true);
    };

    const handleEdit = (webhook: Webhook) => {
        setEditingWebhook(webhook);
        setFormData({
            name: webhook.name,
            url: webhook.url,
            events: webhook.events,
            isActive: webhook.isActive,
            headers: webhook.headers,
        });
        setShowEditModal(true);
    };

    const handleToggleEvent = (event: WebhookEvent) => {
        setFormData(prev => ({
            ...prev,
            events: prev.events.includes(event)
                ? prev.events.filter(e => e !== event)
                : [...prev.events, event],
        }));
    };

    const handleAddHeader = () => {
        setFormData(prev => ({
            ...prev,
            headers: [...prev.headers, { key: '', value: '' }],
        }));
    };

    const handleRemoveHeader = (index: number) => {
        setFormData(prev => ({
            ...prev,
            headers: prev.headers.filter((_, i) => i !== index),
        }));
    };

    const handleUpdateHeader = (index: number, field: 'key' | 'value', value: string) => {
        setFormData(prev => ({
            ...prev,
            headers: prev.headers.map((h, i) => i === index ? { ...h, [field]: value } : h),
        }));
    };

    const submitCreate = async () => {
        if (!formData.name.trim() || !formData.url.trim()) {
            alert('Please fill in name and URL');
            return;
        }
        if (formData.events.length === 0) {
            alert('Please select at least one event');
            return;
        }

        try {
            // In production: const response = await api.post(API_CONFIG.WEBHOOKS_CREATE, formData);
            const generatedSecret = `whsec_${Math.random().toString(36).substring(2, 18)}`;
            setNewSecret(generatedSecret);
            setShowCreateModal(false);
            setShowSecretModal(true);
            await fetchWebhooks();
        } catch (error) {
            console.error('[WebhookManagement] Failed to create webhook:', error);
        }
    };

    const submitEdit = async () => {
        if (!editingWebhook) return;
        try {
            // In production: await api.put(API_CONFIG.WEBHOOKS_UPDATE(editingWebhook.id), formData);
            setWebhooks(prev =>
                prev.map(w =>
                    w.id === editingWebhook.id
                        ? { ...w, ...formData }
                        : w
                )
            );
            setShowEditModal(false);
        } catch (error) {
            console.error('[WebhookManagement] Failed to update webhook:', error);
        }
    };

    const handleDelete = async (webhook: Webhook) => {
        if (!confirm(`Delete webhook "${webhook.name}"?`)) return;
        try {
            // In production: await api.delete(API_CONFIG.WEBHOOKS_DELETE(webhook.id));
            setWebhooks(prev => prev.filter(w => w.id !== webhook.id));
        } catch (error) {
            console.error('[WebhookManagement] Failed to delete webhook:', error);
        }
    };

    const handleToggleActive = async (webhook: Webhook) => {
        try {
            // In production: await api.put(API_CONFIG.WEBHOOKS_TOGGLE(webhook.id));
            setWebhooks(prev =>
                prev.map(w =>
                    w.id === webhook.id
                        ? { ...w, isActive: !w.isActive, status: !w.isActive ? 'active' : 'paused' }
                        : w
                )
            );
        } catch (error) {
            console.error('[WebhookManagement] Failed to toggle webhook:', error);
        }
    };

    const handleTest = (webhook: Webhook) => {
        setTestingWebhook(webhook);
        setTestEvent('trade.opened');
        setTestResult(null);
        setShowTestModal(true);
    };

    const submitTest = async () => {
        if (!testingWebhook) return;
        try {
            // In production: const response = await api.post(API_CONFIG.WEBHOOKS_TEST(testingWebhook.id), { event: testEvent });
            await new Promise(resolve => setTimeout(resolve, 1000));
            setTestResult({
                status: 200,
                headers: { 'content-type': 'application/json' },
                body: { success: true, message: 'Test webhook received' },
            });
        } catch (error) {
            console.error('[WebhookManagement] Failed to test webhook:', error);
        }
    };

    const handleRetryDelivery = async (webhook: Webhook, logEntry: DeliveryLogEntry) => {
        try {
            // In production: await api.post(API_CONFIG.WEBHOOKS_RETRY(webhook.id, logEntry.id));
            alert(`Retrying delivery for event ${logEntry.event}`);
        } catch (error) {
            console.error('[WebhookManagement] Failed to retry delivery:', error);
        }
    };

    const handleCopySecret = () => {
        navigator.clipboard.writeText(newSecret);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const formatRelativeTime = (dateString: string) => {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffMins = Math.floor(diffMs / (1000 * 60));
        const diffHours = Math.floor(diffMs / (1000 * 60 * 60));

        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins} min ago`;
        if (diffHours < 24) return `${diffHours} hours ago`;
        return date.toLocaleDateString();
    };

    const truncateUrl = (url: string, maxLength = 40) => {
        return url.length > maxLength ? url.substring(0, maxLength) + '...' : url;
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full bg-[#121316]">
                <div className="text-zinc-400">Loading webhooks...</div>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full bg-[#121316] text-white overflow-hidden">
            {/* Header */}
            <div className="flex-shrink-0 bg-zinc-900 border-b border-zinc-800 p-4">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <Webhook className="text-blue-400" size={24} />
                        <div>
                            <h1 className="text-xl font-bold text-white">Webhook Management</h1>
                            <p className="text-xs text-zinc-400">{webhooks.length} webhooks configured</p>
                        </div>
                    </div>
                    <button
                        onClick={handleCreate}
                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm font-medium transition-colors"
                    >
                        <Plus size={16} />
                        Add Webhook
                    </button>
                </div>

                {/* Search */}
                <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={16} />
                    <input
                        type="text"
                        placeholder="Search by name or URL..."
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
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400">Name</th>
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400">URL</th>
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400">Events</th>
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400">Status</th>
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400">Last Triggered</th>
                            <th className="text-right px-3 py-2 font-semibold text-zinc-400">Success Rate</th>
                            <th className="text-left px-3 py-2 font-semibold text-zinc-400">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {filteredWebhooks.map((webhook) => (
                            <React.Fragment key={webhook.id}>
                                <tr className="border-b border-zinc-800 hover:bg-zinc-900/50">
                                    <td
                                        className="px-3 py-2 cursor-pointer"
                                        onClick={() => setExpandedRow(expandedRow === webhook.id ? null : webhook.id)}
                                    >
                                        {expandedRow === webhook.id ? (
                                            <ChevronDown size={14} className="text-zinc-500" />
                                        ) : (
                                            <ChevronRight size={14} className="text-zinc-500" />
                                        )}
                                    </td>
                                    <td className="px-3 py-2 font-medium text-white">{webhook.name}</td>
                                    <td className="px-3 py-2 font-mono text-blue-400 text-xs" title={webhook.url}>
                                        {truncateUrl(webhook.url)}
                                    </td>
                                    <td className="px-3 py-2">
                                        <div className="text-zinc-400">{webhook.events.length} events</div>
                                    </td>
                                    <td className="px-3 py-2">
                                        <span className={`px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[webhook.status]}`}>
                                            {webhook.status.charAt(0).toUpperCase() + webhook.status.slice(1)}
                                        </span>
                                    </td>
                                    <td className="px-3 py-2 text-zinc-400">
                                        {webhook.lastTriggered ? formatRelativeTime(webhook.lastTriggered) : 'Never'}
                                    </td>
                                    <td className="px-3 py-2 text-right">
                                        <span className={webhook.successRate >= 90 ? 'text-green-400' : webhook.successRate >= 70 ? 'text-yellow-400' : 'text-red-400'}>
                                            {webhook.successRate.toFixed(0)}%
                                        </span>
                                    </td>
                                    <td className="px-3 py-2">
                                        <div className="flex items-center gap-1">
                                            <button
                                                onClick={() => handleEdit(webhook)}
                                                className="p-1 hover:bg-blue-600/20 text-blue-400 rounded transition-colors"
                                                title="Edit"
                                            >
                                                <Edit size={14} />
                                            </button>
                                            <button
                                                onClick={() => handleToggleActive(webhook)}
                                                className={`p-1 rounded transition-colors ${
                                                    webhook.isActive
                                                        ? 'hover:bg-yellow-600/20 text-yellow-400'
                                                        : 'hover:bg-green-600/20 text-green-400'
                                                }`}
                                                title={webhook.isActive ? 'Pause' : 'Activate'}
                                            >
                                                {webhook.isActive ? <Pause size={14} /> : <Play size={14} />}
                                            </button>
                                            <button
                                                onClick={() => handleTest(webhook)}
                                                className="p-1 hover:bg-purple-600/20 text-purple-400 rounded transition-colors"
                                                title="Test"
                                            >
                                                <Code size={14} />
                                            </button>
                                            <button
                                                onClick={() => handleDelete(webhook)}
                                                className="p-1 hover:bg-red-600/20 text-red-400 rounded transition-colors"
                                                title="Delete"
                                            >
                                                <Trash2 size={14} />
                                            </button>
                                        </div>
                                    </td>
                                </tr>

                                {/* Delivery Log (Expanded) */}
                                {expandedRow === webhook.id && (
                                    <tr className="bg-zinc-900/70 border-b border-zinc-800">
                                        <td colSpan={8} className="px-6 py-4">
                                            <div>
                                                <h3 className="text-sm font-bold text-white mb-3">Delivery Log (Last 25)</h3>
                                                <div className="bg-zinc-800 rounded overflow-hidden max-h-96 overflow-y-auto custom-scrollbar">
                                                    <table className="w-full text-xs">
                                                        <thead className="bg-zinc-900 sticky top-0">
                                                            <tr>
                                                                <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Timestamp</th>
                                                                <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Event</th>
                                                                <th className="text-center px-3 py-2 text-zinc-400 font-semibold">Status Code</th>
                                                                <th className="text-right px-3 py-2 text-zinc-400 font-semibold">Response Time</th>
                                                                <th className="text-center px-3 py-2 text-zinc-400 font-semibold">Attempt</th>
                                                                <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Status</th>
                                                                <th className="text-left px-3 py-2 text-zinc-400 font-semibold">Actions</th>
                                                            </tr>
                                                        </thead>
                                                        <tbody>
                                                            {webhook.deliveryLog.map(log => (
                                                                <tr key={log.id} className="border-t border-zinc-700 hover:bg-zinc-700/50">
                                                                    <td className="px-3 py-2 text-zinc-400">
                                                                        {formatRelativeTime(log.timestamp)}
                                                                    </td>
                                                                    <td className="px-3 py-2 font-mono text-blue-400">{log.event}</td>
                                                                    <td className="px-3 py-2 text-center font-mono text-white">
                                                                        {log.statusCode || '-'}
                                                                    </td>
                                                                    <td className="px-3 py-2 text-right text-zinc-400">
                                                                        {log.responseTime ? `${log.responseTime}ms` : '-'}
                                                                    </td>
                                                                    <td className="px-3 py-2 text-center text-zinc-400">{log.attempt}</td>
                                                                    <td className="px-3 py-2">
                                                                        {log.status === 'success' && (
                                                                            <span className="flex items-center gap-1 text-green-400">
                                                                                <CheckCircle size={12} /> Success
                                                                            </span>
                                                                        )}
                                                                        {log.status === 'failed' && (
                                                                            <span className="flex items-center gap-1 text-red-400">
                                                                                <XCircle size={12} /> Failed
                                                                            </span>
                                                                        )}
                                                                        {log.status === 'pending' && (
                                                                            <span className="flex items-center gap-1 text-yellow-400">
                                                                                <Clock size={12} /> Pending
                                                                            </span>
                                                                        )}
                                                                    </td>
                                                                    <td className="px-3 py-2">
                                                                        <div className="flex items-center gap-1">
                                                                            <button
                                                                                onClick={() => {
                                                                                    setSelectedLog(log);
                                                                                    setShowPayloadModal(true);
                                                                                }}
                                                                                className="p-1 hover:bg-blue-600/20 text-blue-400 rounded transition-colors"
                                                                                title="View Payload"
                                                                            >
                                                                                <Code size={12} />
                                                                            </button>
                                                                            {log.status === 'failed' && (
                                                                                <button
                                                                                    onClick={() => handleRetryDelivery(webhook, log)}
                                                                                    className="p-1 hover:bg-green-600/20 text-green-400 rounded transition-colors"
                                                                                    title="Retry"
                                                                                >
                                                                                    <RefreshCw size={12} />
                                                                                </button>
                                                                            )}
                                                                        </div>
                                                                    </td>
                                                                </tr>
                                                            ))}
                                                        </tbody>
                                                    </table>
                                                </div>
                                            </div>
                                        </td>
                                    </tr>
                                )}
                            </React.Fragment>
                        ))}
                    </tbody>
                </table>

                {filteredWebhooks.length === 0 && (
                    <div className="text-center py-12 text-zinc-500">
                        No webhooks found
                    </div>
                )}
            </div>

            {/* Create/Edit Modal */}
            {(showCreateModal || showEditModal) && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 overflow-y-auto">
                    <div className="bg-zinc-900 rounded-lg shadow-xl border border-zinc-700 w-full max-w-2xl my-8">
                        <div className="flex items-center justify-between p-4 border-b border-zinc-800">
                            <h2 className="text-lg font-bold text-white">
                                {showCreateModal ? 'Create Webhook' : 'Edit Webhook'}
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
                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-1">
                                    Name <span className="text-red-400">*</span>
                                </label>
                                <input
                                    type="text"
                                    value={formData.name}
                                    onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                                    placeholder="e.g., Production Trading Bot"
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                                />
                            </div>

                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-1">
                                    URL <span className="text-red-400">*</span>
                                </label>
                                <input
                                    type="url"
                                    value={formData.url}
                                    onChange={(e) => setFormData(prev => ({ ...prev, url: e.target.value }))}
                                    placeholder="https://api.example.com/webhooks"
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                                />
                                <div className="text-xs text-zinc-500 mt-1">Must start with https://</div>
                            </div>

                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-2">
                                    Events <span className="text-red-400">*</span>
                                </label>
                                <div className="grid grid-cols-2 gap-2">
                                    {WEBHOOK_EVENTS.map(event => (
                                        <label
                                            key={event.value}
                                            className="flex items-start gap-2 px-3 py-2 bg-zinc-800 rounded cursor-pointer hover:bg-zinc-700 transition-colors"
                                        >
                                            <input
                                                type="checkbox"
                                                checked={formData.events.includes(event.value)}
                                                onChange={() => handleToggleEvent(event.value)}
                                                className="mt-0.5 w-4 h-4 accent-blue-600"
                                            />
                                            <div className="flex-1">
                                                <div className="text-xs font-medium text-white">{event.label}</div>
                                                <div className="text-xs text-zinc-400">{event.description}</div>
                                            </div>
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
                                    <span className="text-sm text-white">Active (start delivering events immediately)</span>
                                </label>
                            </div>

                            <div>
                                <div className="flex items-center justify-between mb-2">
                                    <label className="text-xs font-semibold text-zinc-400">
                                        Custom Headers (Optional)
                                    </label>
                                    <button
                                        onClick={handleAddHeader}
                                        className="text-xs text-blue-400 hover:text-blue-300"
                                    >
                                        + Add Header
                                    </button>
                                </div>
                                {formData.headers.map((header, index) => (
                                    <div key={index} className="flex gap-2 mb-2">
                                        <input
                                            type="text"
                                            value={header.key}
                                            onChange={(e) => handleUpdateHeader(index, 'key', e.target.value)}
                                            placeholder="Header name"
                                            className="flex-1 px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                                        />
                                        <input
                                            type="text"
                                            value={header.value}
                                            onChange={(e) => handleUpdateHeader(index, 'value', e.target.value)}
                                            placeholder="Header value"
                                            className="flex-1 px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                                        />
                                        <button
                                            onClick={() => handleRemoveHeader(index)}
                                            className="p-2 hover:bg-red-600/20 text-red-400 rounded transition-colors"
                                        >
                                            <X size={16} />
                                        </button>
                                    </div>
                                ))}
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

            {/* Secret Reveal Modal */}
            {showSecretModal && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
                    <div className="bg-zinc-900 rounded-lg shadow-xl border border-zinc-700 w-full max-w-lg">
                        <div className="flex items-center justify-between p-4 border-b border-zinc-800">
                            <h2 className="text-lg font-bold text-white">Webhook Created</h2>
                        </div>

                        <div className="p-4 space-y-4">
                            <div className="bg-yellow-900/20 border border-yellow-600 rounded p-3 flex gap-3">
                                <AlertTriangle size={20} className="text-yellow-400 flex-shrink-0 mt-0.5" />
                                <div className="text-xs text-yellow-200">
                                    <div className="font-bold mb-1">Save this secret now!</div>
                                    <div>For security, this is the only time you will see the webhook signing secret. Use it to verify webhook signatures.</div>
                                </div>
                            </div>

                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-1">
                                    Signing Secret
                                </label>
                                <div className="flex items-center gap-2">
                                    <div className="flex-1 px-3 py-2 bg-zinc-800 border border-zinc-700 rounded font-mono text-xs text-purple-400 break-all select-all">
                                        {newSecret}
                                    </div>
                                    <button
                                        onClick={handleCopySecret}
                                        className="p-2 bg-blue-600 hover:bg-blue-700 rounded transition-colors flex-shrink-0"
                                        title="Copy"
                                    >
                                        {copied ? <Check size={16} /> : <Copy size={16} />}
                                    </button>
                                </div>
                            </div>
                        </div>

                        <div className="flex items-center justify-end gap-3 p-4 border-t border-zinc-800">
                            <button
                                onClick={() => {
                                    setShowSecretModal(false);
                                    setNewSecret('');
                                }}
                                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm font-medium transition-colors"
                            >
                                I've Saved the Secret
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Test Modal */}
            {showTestModal && testingWebhook && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
                    <div className="bg-zinc-900 rounded-lg shadow-xl border border-zinc-700 w-full max-w-2xl">
                        <div className="flex items-center justify-between p-4 border-b border-zinc-800">
                            <h2 className="text-lg font-bold text-white">Test Webhook</h2>
                            <button
                                onClick={() => setShowTestModal(false)}
                                className="text-zinc-400 hover:text-white transition-colors"
                            >
                                <X size={20} />
                            </button>
                        </div>

                        <div className="p-4 space-y-4">
                            <div>
                                <div className="text-sm text-zinc-400 mb-1">Webhook:</div>
                                <div className="text-base font-semibold text-white">{testingWebhook.name}</div>
                            </div>

                            <div>
                                <label className="block text-xs font-semibold text-zinc-400 mb-1">
                                    Select Event Type
                                </label>
                                <select
                                    value={testEvent}
                                    onChange={(e) => setTestEvent(e.target.value as WebhookEvent)}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-blue-500"
                                >
                                    {WEBHOOK_EVENTS.map(event => (
                                        <option key={event.value} value={event.value}>
                                            {event.label}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <button
                                onClick={submitTest}
                                className="w-full px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded text-sm font-medium transition-colors"
                            >
                                Send Test Event
                            </button>

                            {testResult && (
                                <div>
                                    <div className="text-sm font-semibold text-white mb-2">Response:</div>
                                    <div className="bg-zinc-800 rounded p-3 font-mono text-xs">
                                        <div className="text-green-400 mb-2">Status: {testResult.status}</div>
                                        <div className="text-zinc-400 mb-2">
                                            Headers: {JSON.stringify(testResult.headers, null, 2)}
                                        </div>
                                        <div className="text-white">
                                            Body: {JSON.stringify(testResult.body, null, 2)}
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>

                        <div className="flex items-center justify-end gap-3 p-4 border-t border-zinc-800">
                            <button
                                onClick={() => setShowTestModal(false)}
                                className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded text-sm font-medium transition-colors"
                            >
                                Close
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Payload Modal */}
            {showPayloadModal && selectedLog && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
                    <div className="bg-zinc-900 rounded-lg shadow-xl border border-zinc-700 w-full max-w-3xl">
                        <div className="flex items-center justify-between p-4 border-b border-zinc-800">
                            <h2 className="text-lg font-bold text-white">Delivery Details</h2>
                            <button
                                onClick={() => setShowPayloadModal(false)}
                                className="text-zinc-400 hover:text-white transition-colors"
                            >
                                <X size={20} />
                            </button>
                        </div>

                        <div className="p-4 space-y-4 max-h-[calc(100vh-200px)] overflow-y-auto custom-scrollbar">
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <div className="text-xs text-zinc-400 mb-1">Event</div>
                                    <div className="text-sm font-mono text-blue-400">{selectedLog.event}</div>
                                </div>
                                <div>
                                    <div className="text-xs text-zinc-400 mb-1">Status Code</div>
                                    <div className="text-sm font-mono text-white">{selectedLog.statusCode || 'N/A'}</div>
                                </div>
                                <div>
                                    <div className="text-xs text-zinc-400 mb-1">Response Time</div>
                                    <div className="text-sm text-white">{selectedLog.responseTime ? `${selectedLog.responseTime}ms` : 'N/A'}</div>
                                </div>
                                <div>
                                    <div className="text-xs text-zinc-400 mb-1">Attempt</div>
                                    <div className="text-sm text-white">#{selectedLog.attempt}</div>
                                </div>
                            </div>

                            <div>
                                <div className="text-sm font-semibold text-white mb-2">Request Payload:</div>
                                <pre className="bg-zinc-800 rounded p-3 font-mono text-xs text-white overflow-x-auto">
                                    {JSON.stringify(selectedLog.requestPayload, null, 2)}
                                </pre>
                            </div>

                            <div>
                                <div className="text-sm font-semibold text-white mb-2">Response Payload:</div>
                                <pre className="bg-zinc-800 rounded p-3 font-mono text-xs text-white overflow-x-auto">
                                    {selectedLog.responsePayload
                                        ? JSON.stringify(selectedLog.responsePayload, null, 2)
                                        : 'No response received'}
                                </pre>
                            </div>
                        </div>

                        <div className="flex items-center justify-end gap-3 p-4 border-t border-zinc-800">
                            <button
                                onClick={() => setShowPayloadModal(false)}
                                className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded text-sm font-medium transition-colors"
                            >
                                Close
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
