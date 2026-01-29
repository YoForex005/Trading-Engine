'use client';

import React, { useState, useEffect } from 'react';
import {
    Plus,
    Power,
    PowerOff,
    RefreshCw,
    Trash2,
    Settings,
    Loader2,
    CheckCircle,
    XCircle,
    Wifi,
    WifiOff,
    ChevronDown,
    ChevronUp
} from 'lucide-react';

interface LPConfig {
    id: string;
    name: string;
    type: string;
    enabled: boolean;
    priority: number;
    settings: Record<string, string>;
    symbols: string[];
}

interface LPStatus {
    id: string;
    name: string;
    type: string;
    connected: boolean;
    enabled: boolean;
    symbolCount: number;
    lastTick?: string;
    errorMessage?: string;
}

const LP_TYPES = [
    { value: 'OANDA', label: 'OANDA (REST API)' },
    { value: 'BINANCE', label: 'Binance (WebSocket)' },
    { value: 'FLEXYMARKETS', label: 'FlexyMarkets (WebSocket)' },
    { value: 'LMAX', label: 'LMAX (FIX 4.4)' },
    { value: 'CURRENEX', label: 'Currenex (FIX 4.4)' },
    { value: 'CUSTOM', label: 'Custom LP' },
];

export default function LPManagementView() {
    const [lps, setLps] = useState<LPConfig[]>([]);
    const [status, setStatus] = useState<Record<string, LPStatus>>({});
    const [loading, setLoading] = useState(true);
    const [expandedLP, setExpandedLP] = useState<string | null>(null);
    const [showAddModal, setShowAddModal] = useState(false);
    const [newLP, setNewLP] = useState<Partial<LPConfig>>({
        type: 'BINANCE',
        enabled: true,
        priority: 10,
        settings: {},
        symbols: [],
    });

    useEffect(() => {
        fetchLPs();
        fetchStatus();
        const interval = setInterval(fetchStatus, 5000);
        return () => clearInterval(interval);
    }, []);

    const fetchLPs = async () => {
        try {
            const res = await fetch('http://localhost:8080/admin/lps');
            if (res.ok) {
                const data = await res.json();
                setLps(data || []);
            }
        } catch (e) {
            console.error('Failed to fetch LPs:', e);
        } finally {
            setLoading(false);
        }
    };

    const fetchStatus = async () => {
        try {
            const res = await fetch('http://localhost:8080/admin/lp-status');
            if (res.ok) {
                const data = await res.json();
                setStatus(data || {});
            }
        } catch (e) {
            console.error('Failed to fetch LP status:', e);
        }
    };

    const handleToggle = async (id: string) => {
        try {
            const res = await fetch(`http://localhost:8080/admin/lps/${id}/toggle`, {
                method: 'POST',
            });
            if (res.ok) {
                fetchLPs();
                fetchStatus();
            }
        } catch (e) {
            console.error('Failed to toggle LP:', e);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm(`Are you sure you want to remove LP "${id}"?`)) return;

        try {
            const res = await fetch(`http://localhost:8080/admin/lps/${id}`, {
                method: 'DELETE',
            });
            if (res.ok) {
                fetchLPs();
            }
        } catch (e) {
            console.error('Failed to delete LP:', e);
        }
    };

    const handleAddLP = async () => {
        if (!newLP.name || !newLP.type) {
            alert('Name and Type are required');
            return;
        }

        const config: LPConfig = {
            id: newLP.name?.toLowerCase().replace(/\s+/g, '_') || '',
            name: newLP.name || '',
            type: newLP.type || '',
            enabled: newLP.enabled ?? true,
            priority: newLP.priority ?? 10,
            settings: newLP.settings || {},
            symbols: newLP.symbols || [],
        };

        try {
            const res = await fetch('http://localhost:8080/admin/lps', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config),
            });

            if (res.ok) {
                setShowAddModal(false);
                setNewLP({ type: 'BINANCE', enabled: true, priority: 10, settings: {}, symbols: [] });
                fetchLPs();
            } else {
                const error = await res.json();
                alert(error.error || 'Failed to add LP');
            }
        } catch (e) {
            console.error('Failed to add LP:', e);
        }
    };

    const fetchSymbols = async (id: string) => {
        try {
            const res = await fetch(`http://localhost:8080/admin/lps/${id}/symbols`);
            if (res.ok) {
                const data = await res.json();
                alert(`${id} has ${data.count} available symbols`);
            }
        } catch (e) {
            console.error('Failed to fetch symbols:', e);
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <Loader2 className="w-8 h-8 animate-spin text-emerald-500" />
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex justify-between items-center">
                <div>
                    <h2 className="text-xl font-semibold">Liquidity Providers</h2>
                    <p className="text-sm text-zinc-400">Manage and configure your price feed sources</p>
                </div>
                <button
                    onClick={() => setShowAddModal(true)}
                    className="flex items-center gap-2 px-4 py-2 bg-emerald-500/20 border border-emerald-500/30 text-emerald-400 rounded-lg hover:bg-emerald-500/30 transition-colors"
                >
                    <Plus className="w-4 h-4" />
                    Add LP
                </button>
            </div>

            {/* LP List */}
            <div className="space-y-4">
                {lps.length === 0 ? (
                    <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-8 text-center">
                        <p className="text-zinc-400">No LPs configured. Add one to get started.</p>
                    </div>
                ) : (
                    lps.map((lp) => {
                        const lpStatus = status[lp.id] || {};
                        const isExpanded = expandedLP === lp.id;

                        return (
                            <div
                                key={lp.id}
                                className={`bg-zinc-900/50 border rounded-xl overflow-hidden transition-colors ${lp.enabled ? 'border-zinc-700' : 'border-zinc-800 opacity-60'
                                    }`}
                            >
                                {/* LP Header */}
                                <div className="p-4 flex items-center justify-between">
                                    <div className="flex items-center gap-4">
                                        {/* Status Indicator */}
                                        <div className={`w-3 h-3 rounded-full ${lpStatus.connected ? 'bg-green-500 animate-pulse' : 'bg-red-500'
                                            }`} />

                                        {/* LP Info */}
                                        <div>
                                            <div className="flex items-center gap-2">
                                                <span className="font-medium">{lp.name}</span>
                                                <span className="text-xs px-2 py-0.5 bg-zinc-800 rounded text-zinc-400">
                                                    {lp.type}
                                                </span>
                                            </div>
                                            <div className="text-xs text-zinc-500 flex items-center gap-2">
                                                {lpStatus.connected ? (
                                                    <span className="flex items-center gap-1 text-green-400">
                                                        <Wifi className="w-3 h-3" /> Connected
                                                    </span>
                                                ) : (
                                                    <span className="flex items-center gap-1 text-red-400">
                                                        <WifiOff className="w-3 h-3" /> Disconnected
                                                    </span>
                                                )}
                                                <span>•</span>
                                                <span>{lpStatus.symbolCount || 0} symbols</span>
                                                <span>•</span>
                                                <span>Priority: {lp.priority}</span>
                                            </div>
                                        </div>
                                    </div>

                                    {/* Actions */}
                                    <div className="flex items-center gap-2">
                                        <button
                                            onClick={() => handleToggle(lp.id)}
                                            className={`p-2 rounded-lg transition-colors ${lp.enabled
                                                    ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30'
                                                    : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
                                                }`}
                                            title={lp.enabled ? 'Disable' : 'Enable'}
                                        >
                                            {lp.enabled ? <Power className="w-4 h-4" /> : <PowerOff className="w-4 h-4" />}
                                        </button>

                                        <button
                                            onClick={() => fetchSymbols(lp.id)}
                                            className="p-2 rounded-lg bg-zinc-800 text-zinc-400 hover:bg-zinc-700 transition-colors"
                                            title="View Symbols"
                                        >
                                            <RefreshCw className="w-4 h-4" />
                                        </button>

                                        <button
                                            onClick={() => setExpandedLP(isExpanded ? null : lp.id)}
                                            className="p-2 rounded-lg bg-zinc-800 text-zinc-400 hover:bg-zinc-700 transition-colors"
                                        >
                                            {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                                        </button>

                                        <button
                                            onClick={() => handleDelete(lp.id)}
                                            className="p-2 rounded-lg bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors"
                                            title="Remove LP"
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </button>
                                    </div>
                                </div>

                                {/* Expanded Details */}
                                {isExpanded && (
                                    <div className="border-t border-zinc-800 p-4 bg-zinc-900/30">
                                        <div className="grid grid-cols-2 gap-4 text-sm">
                                            <div>
                                                <span className="text-zinc-500">ID:</span>
                                                <span className="ml-2 font-mono">{lp.id}</span>
                                            </div>
                                            <div>
                                                <span className="text-zinc-500">Type:</span>
                                                <span className="ml-2">{lp.type}</span>
                                            </div>
                                            <div>
                                                <span className="text-zinc-500">Symbols:</span>
                                                <span className="ml-2">
                                                    {lp.symbols.length > 0 ? lp.symbols.join(', ') : 'All available'}
                                                </span>
                                            </div>
                                            {lpStatus.errorMessage && (
                                                <div className="col-span-2 text-red-400">
                                                    <span className="text-zinc-500">Error:</span>
                                                    <span className="ml-2">{lpStatus.errorMessage}</span>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                )}
                            </div>
                        );
                    })
                )}
            </div>

            {/* Add LP Modal */}
            {showAddModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-full max-w-md">
                        <h3 className="text-lg font-semibold mb-4">Add Liquidity Provider</h3>

                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm text-zinc-400 mb-1">LP Type</label>
                                <select
                                    value={newLP.type}
                                    onChange={(e) => setNewLP({ ...newLP, type: e.target.value })}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                                >
                                    {LP_TYPES.map((t) => (
                                        <option key={t.value} value={t.value}>{t.label}</option>
                                    ))}
                                </select>
                            </div>

                            <div>
                                <label className="block text-sm text-zinc-400 mb-1">Name</label>
                                <input
                                    type="text"
                                    value={newLP.name || ''}
                                    onChange={(e) => setNewLP({ ...newLP, name: e.target.value })}
                                    placeholder="e.g., Binance Crypto"
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                                />
                            </div>

                            <div>
                                <label className="block text-sm text-zinc-400 mb-1">Priority (lower = higher)</label>
                                <input
                                    type="number"
                                    value={newLP.priority || 10}
                                    onChange={(e) => setNewLP({ ...newLP, priority: parseInt(e.target.value) })}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                                />
                            </div>

                            <label className="flex items-center gap-2">
                                <input
                                    type="checkbox"
                                    checked={newLP.enabled ?? true}
                                    onChange={(e) => setNewLP({ ...newLP, enabled: e.target.checked })}
                                    className="accent-emerald-500"
                                />
                                <span className="text-sm">Enable immediately</span>
                            </label>
                        </div>

                        <div className="flex justify-end gap-3 mt-6">
                            <button
                                onClick={() => setShowAddModal(false)}
                                className="px-4 py-2 text-zinc-400 hover:text-white transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleAddLP}
                                className="px-4 py-2 bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 transition-colors"
                            >
                                Add LP
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
