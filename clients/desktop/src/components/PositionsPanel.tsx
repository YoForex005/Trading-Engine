/**
 * Enhanced Positions/Orders Panel
 * Professional positions and orders management with inline editing
 * Features: SL/TP inline editing, partial close, close all, real-time P&L
 */

import { useState, useEffect, useCallback } from 'react';
import { X, TrendingUp, TrendingDown, AlertTriangle, Edit2, Trash2 } from 'lucide-react';
import { API_ENDPOINTS } from '../config/api';

interface Position {
    id: number;
    symbol: string;
    side: 'BUY' | 'SELL';
    volume: number;
    openPrice: number;
    currentPrice: number;
    openTime: string;
    sl: number;
    tp: number;
    swap: number;
    commission: number;
    unrealizedPnL: number;
}

interface PendingOrder {
    id: string;
    symbol: string;
    type: 'LIMIT' | 'STOP' | 'STOP_LIMIT';
    side: 'BUY' | 'SELL';
    volume: number;
    price: number;
    triggerPrice?: number;
    sl?: number;
    tp?: number;
    status: 'PENDING' | 'TRIGGERED' | 'FILLED' | 'CANCELLED';
    createdAt: string;
}

type TabType = 'Positions' | 'Orders';

interface PositionsPanelProps {
    accountId: string;
    wsConnection: WebSocket | null;
}

interface EditingState {
    positionId: number;
    field: 'sl' | 'tp';
    value: string;
}

export function PositionsPanel({ accountId, wsConnection }: PositionsPanelProps) {
    const [activeTab, setActiveTab] = useState<TabType>('Positions');
    const [positions, setPositions] = useState<Position[]>([]);
    const [orders, setOrders] = useState<PendingOrder[]>([]);
    const [editing, setEditing] = useState<EditingState | null>(null);
    const [showCloseAllConfirm, setShowCloseAllConfirm] = useState(false);
    const [showPartialClose, setShowPartialClose] = useState<{ positionId: number; maxVolume: number } | null>(null);
    const [partialCloseVolume, setPartialCloseVolume] = useState<string>('');
    const [loading, setLoading] = useState(false);

    // Fetch positions
    const fetchPositions = useCallback(async () => {
        try {
            const res = await fetch(`${API_ENDPOINTS.positions}?accountId=${accountId}`);
            if (res.ok) {
                const data = await res.json();
                setPositions(data || []);
            }
        } catch (err) {
            console.error('[PositionsPanel] Failed to fetch positions:', err);
        }
    }, [accountId]);

    // Fetch pending orders
    const fetchOrders = useCallback(async () => {
        try {
            const res = await fetch(`${API_ENDPOINTS.orders}?accountId=${accountId}&status=PENDING`);
            if (res.ok) {
                const data = await res.json();
                setOrders(data || []);
            }
        } catch (err) {
            console.error('[PositionsPanel] Failed to fetch orders:', err);
        }
    }, [accountId]);

    // Fetch on mount and set intervals
    useEffect(() => {
        fetchPositions();
        fetchOrders();

        // Real-time updates every 1 second for P&L
        const posInterval = setInterval(fetchPositions, 1000);
        const ordInterval = setInterval(fetchOrders, 3000);

        return () => {
            clearInterval(posInterval);
            clearInterval(ordInterval);
        };
    }, [fetchPositions, fetchOrders]);

    // WebSocket updates
    useEffect(() => {
        if (!wsConnection) return;

        const handleMessage = (event: MessageEvent) => {
            try {
                const data = JSON.parse(event.data);
                if (data.type === 'position_update') {
                    setPositions((prev) => {
                        const index = prev.findIndex((p) => p.id === data.position.id);
                        if (index >= 0) {
                            const updated = [...prev];
                            updated[index] = { ...updated[index], ...data.position };
                            return updated;
                        }
                        return prev;
                    });
                }
            } catch (err) {
                console.error('[PositionsPanel] WS parse error:', err);
            }
        };

        wsConnection.addEventListener('message', handleMessage);
        return () => wsConnection.removeEventListener('message', handleMessage);
    }, [wsConnection]);

    // Close position
    const handleClosePosition = async (id: number) => {
        if (!confirm('Close this position?')) return;

        try {
            setLoading(true);
            const res = await fetch(API_ENDPOINTS.closePosition, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ positionId: id })
            });

            if (!res.ok) throw new Error('Failed to close position');
            await fetchPositions();
        } catch (err) {
            console.error('[PositionsPanel] Close position failed:', err);
            alert('Failed to close position');
        } finally {
            setLoading(false);
        }
    };

    // Close all positions
    const handleCloseAll = async () => {
        try {
            setLoading(true);
            const closePromises = positions.map(pos =>
                fetch(API_ENDPOINTS.closePosition, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ positionId: pos.id })
                })
            );

            await Promise.all(closePromises);
            await fetchPositions();
            setShowCloseAllConfirm(false);
        } catch (err) {
            console.error('[PositionsPanel] Close all failed:', err);
            alert('Failed to close all positions');
        } finally {
            setLoading(false);
        }
    };

    // Partial close
    const handlePartialClose = async () => {
        if (!showPartialClose) return;

        const volume = parseFloat(partialCloseVolume);
        if (isNaN(volume) || volume <= 0 || volume > showPartialClose.maxVolume) {
            alert(`Invalid volume. Must be between 0 and ${showPartialClose.maxVolume}`);
            return;
        }

        try {
            setLoading(true);
            const res = await fetch(`${API_ENDPOINTS.positions}/${showPartialClose.positionId}/partial-close`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ volume })
            });

            if (!res.ok) throw new Error('Failed to partially close position');
            await fetchPositions();
            setShowPartialClose(null);
            setPartialCloseVolume('');
        } catch (err) {
            console.error('[PositionsPanel] Partial close failed:', err);
            alert('Failed to partially close position');
        } finally {
            setLoading(false);
        }
    };

    // Modify SL/TP
    const handleModifySLTP = async (positionId: number, field: 'sl' | 'tp', value: string) => {
        const numValue = parseFloat(value);
        if (isNaN(numValue) || numValue < 0) {
            alert('Invalid price value');
            return;
        }

        try {
            setLoading(true);
            const res = await fetch(`${API_ENDPOINTS.positions}/${positionId}/modify`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ [field]: numValue })
            });

            if (!res.ok) throw new Error('Failed to modify position');
            await fetchPositions();
            setEditing(null);
        } catch (err) {
            console.error('[PositionsPanel] Modify SL/TP failed:', err);
            alert('Failed to modify position');
        } finally {
            setLoading(false);
        }
    };

    // Cancel order
    const handleCancelOrder = async (orderId: string) => {
        if (!confirm('Cancel this order?')) return;

        try {
            setLoading(true);
            const res = await fetch(`${API_ENDPOINTS.orders}/${orderId}`, {
                method: 'DELETE'
            });

            if (!res.ok) throw new Error('Failed to cancel order');
            await fetchOrders();
        } catch (err) {
            console.error('[PositionsPanel] Cancel order failed:', err);
            alert('Failed to cancel order');
        } finally {
            setLoading(false);
        }
    };

    // Format functions
    const formatMoney = (val: number | undefined | null) => {
        const v = val ?? 0;
        return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v);
    };

    const formatDate = (dateStr: string) => {
        try {
            return new Date(dateStr).toLocaleString('en-US', {
                year: 'numeric', month: '2-digit', day: '2-digit',
                hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false
            }).replace(',', '');
        } catch {
            return dateStr;
        }
    };

    // Calculate totals
    const totalPnL = positions.reduce((sum, pos) => sum + (pos.unrealizedPnL || 0), 0);
    const totalMargin = positions.reduce((sum, pos) => sum + (pos.volume * 1000), 0); // Simplified
    const totalSwap = positions.reduce((sum, pos) => sum + (pos.swap || 0), 0);

    return (
        <div className="h-full flex flex-col bg-[#1e1e1e]">
            {/* Tab Bar */}
            <div className="flex items-center bg-[#1e1e1e] border-b border-zinc-700 flex-shrink-0">
                <div className="flex items-center flex-1">
                    {(['Positions', 'Orders'] as TabType[]).map((tab) => (
                        <button
                            key={tab}
                            onClick={() => setActiveTab(tab)}
                            className={`px-4 py-2 text-xs font-medium transition-colors whitespace-nowrap border-r border-zinc-700/50 relative
                                ${activeTab === tab
                                    ? 'text-zinc-100 bg-[#2d3436] shadow-[inset_0_-2px_0_0_#3b82f6]'
                                    : 'text-zinc-500 hover:text-zinc-300 hover:bg-[#2d3436]/50'}`}
                        >
                            {tab} {tab === 'Positions' && positions.length > 0 ? `(${positions.length})` : ''}
                            {tab === 'Orders' && orders.length > 0 ? `(${orders.length})` : ''}
                        </button>
                    ))}
                </div>
                {activeTab === 'Positions' && positions.length > 0 && (
                    <button
                        onClick={() => setShowCloseAllConfirm(true)}
                        className="px-3 py-1 mr-2 text-xs font-medium bg-red-500/20 text-red-400 hover:bg-red-500/30 rounded transition-colors"
                        disabled={loading}
                    >
                        Close All
                    </button>
                )}
            </div>

            {/* Content */}
            <div className="flex-1 overflow-hidden relative">
                {activeTab === 'Positions' && (
                    <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700">
                        <table className="w-full text-left border-collapse">
                            <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                                <tr>
                                    <th className="p-1 pl-2 border-r border-zinc-700">Ticket</th>
                                    <th className="p-1 border-r border-zinc-700">Symbol</th>
                                    <th className="p-1 border-r border-zinc-700">Type</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">Volume</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">Open</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">Current</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">S/L</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">T/P</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">P&L</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">Swap</th>
                                    <th className="p-1 border-r border-zinc-700">Time</th>
                                    <th className="p-1 pr-2">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-zinc-800 text-zinc-300 font-mono text-[11px]">
                                {positions.length === 0 ? (
                                    <tr>
                                        <td colSpan={12} className="p-8 text-center text-zinc-600 text-xs italic">
                                            No open positions
                                        </td>
                                    </tr>
                                ) : (
                                    positions.map((pos, i) => (
                                        <tr key={pos.id} className={`${i % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'} hover:bg-[#2d3436] transition-colors group`}>
                                            <td className="p-1 pl-2 border-r border-zinc-800 text-zinc-500">{pos.id}</td>
                                            <td className="p-1 border-r border-zinc-800 font-bold text-zinc-100">{pos.symbol}</td>
                                            <td className={`p-1 border-r border-zinc-800 font-bold text-[10px] ${pos.side === 'BUY' ? 'text-blue-400' : 'text-red-400'}`}>
                                                {pos.side}
                                            </td>
                                            <td className="p-1 border-r border-zinc-800 text-right font-medium text-zinc-200">{pos.volume.toFixed(2)}</td>
                                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-400">{pos.openPrice.toFixed(5)}</td>
                                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-300">{pos.currentPrice.toFixed(5)}</td>
                                            {/* SL - Inline Editing */}
                                            <td className="p-1 border-r border-zinc-800 text-right">
                                                {editing?.positionId === pos.id && editing.field === 'sl' ? (
                                                    <input
                                                        type="number"
                                                        step="0.00001"
                                                        value={editing.value}
                                                        onChange={(e) => setEditing({ ...editing, value: e.target.value })}
                                                        onBlur={() => handleModifySLTP(pos.id, 'sl', editing.value)}
                                                        onKeyDown={(e) => {
                                                            if (e.key === 'Enter') handleModifySLTP(pos.id, 'sl', editing.value);
                                                            if (e.key === 'Escape') setEditing(null);
                                                        }}
                                                        autoFocus
                                                        className="w-full bg-zinc-800 text-zinc-100 px-1 py-0.5 rounded text-[11px] font-mono"
                                                    />
                                                ) : (
                                                    <div
                                                        onClick={() => setEditing({ positionId: pos.id, field: 'sl', value: pos.sl.toString() })}
                                                        className="cursor-pointer hover:bg-zinc-700 rounded px-1"
                                                    >
                                                        <span className={pos.sl > 0 ? "text-red-300" : "text-zinc-600"}>
                                                            {pos.sl > 0 ? pos.sl.toFixed(5) : '—'}
                                                        </span>
                                                    </div>
                                                )}
                                            </td>
                                            {/* TP - Inline Editing */}
                                            <td className="p-1 border-r border-zinc-800 text-right">
                                                {editing?.positionId === pos.id && editing.field === 'tp' ? (
                                                    <input
                                                        type="number"
                                                        step="0.00001"
                                                        value={editing.value}
                                                        onChange={(e) => setEditing({ ...editing, value: e.target.value })}
                                                        onBlur={() => handleModifySLTP(pos.id, 'tp', editing.value)}
                                                        onKeyDown={(e) => {
                                                            if (e.key === 'Enter') handleModifySLTP(pos.id, 'tp', editing.value);
                                                            if (e.key === 'Escape') setEditing(null);
                                                        }}
                                                        autoFocus
                                                        className="w-full bg-zinc-800 text-zinc-100 px-1 py-0.5 rounded text-[11px] font-mono"
                                                    />
                                                ) : (
                                                    <div
                                                        onClick={() => setEditing({ positionId: pos.id, field: 'tp', value: pos.tp.toString() })}
                                                        className="cursor-pointer hover:bg-zinc-700 rounded px-1"
                                                    >
                                                        <span className={pos.tp > 0 ? "text-emerald-300" : "text-zinc-600"}>
                                                            {pos.tp > 0 ? pos.tp.toFixed(5) : '—'}
                                                        </span>
                                                    </div>
                                                )}
                                            </td>
                                            <td className={`p-1 border-r border-zinc-800 text-right font-bold ${pos.unrealizedPnL >= 0 ? 'text-[#4ade80]' : 'text-[#f87171]'}`}>
                                                {formatMoney(pos.unrealizedPnL)}
                                            </td>
                                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-500 text-[10px]">{formatMoney(pos.swap)}</td>
                                            <td className="p-1 border-r border-zinc-800 text-zinc-500 text-[10px]">{formatDate(pos.openTime).split(' ')[1]}</td>
                                            <td className="p-1 pr-2">
                                                <div className="flex items-center gap-1">
                                                    <button
                                                        onClick={() => setShowPartialClose({ positionId: pos.id, maxVolume: pos.volume })}
                                                        className="text-zinc-600 hover:text-yellow-400 hover:bg-zinc-700 w-5 h-5 rounded flex items-center justify-center transition-colors"
                                                        title="Partial Close"
                                                        disabled={loading}
                                                    >
                                                        <TrendingDown size={12} />
                                                    </button>
                                                    <button
                                                        onClick={() => handleClosePosition(pos.id)}
                                                        className="text-zinc-600 hover:text-white hover:bg-zinc-700 w-5 h-5 rounded flex items-center justify-center transition-colors"
                                                        title="Close Position"
                                                        disabled={loading}
                                                    >
                                                        <X size={12} />
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                )}

                {activeTab === 'Orders' && (
                    <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700">
                        <table className="w-full text-left border-collapse">
                            <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                                <tr>
                                    <th className="p-1 pl-2 border-r border-zinc-700">Ticket</th>
                                    <th className="p-1 border-r border-zinc-700">Symbol</th>
                                    <th className="p-1 border-r border-zinc-700">Type</th>
                                    <th className="p-1 border-r border-zinc-700">Side</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">Volume</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">Price</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">S/L</th>
                                    <th className="p-1 border-r border-zinc-700 text-right">T/P</th>
                                    <th className="p-1 border-r border-zinc-700">Status</th>
                                    <th className="p-1 border-r border-zinc-700">Created</th>
                                    <th className="p-1 pr-2">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-zinc-800 text-zinc-300 font-mono text-[11px]">
                                {orders.length === 0 ? (
                                    <tr>
                                        <td colSpan={11} className="p-8 text-center text-zinc-600 text-xs italic">
                                            No pending orders
                                        </td>
                                    </tr>
                                ) : (
                                    orders.map((order, i) => (
                                        <tr key={order.id} className={`${i % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'} hover:bg-[#2d3436] transition-colors group`}>
                                            <td className="p-1 pl-2 border-r border-zinc-800 text-zinc-500">{order.id}</td>
                                            <td className="p-1 border-r border-zinc-800 font-bold text-zinc-100">{order.symbol}</td>
                                            <td className="p-1 border-r border-zinc-800 text-zinc-400 text-[10px]">{order.type}</td>
                                            <td className={`p-1 border-r border-zinc-800 font-bold text-[10px] ${order.side === 'BUY' ? 'text-blue-400' : 'text-red-400'}`}>
                                                {order.side}
                                            </td>
                                            <td className="p-1 border-r border-zinc-800 text-right font-medium text-zinc-200">{order.volume.toFixed(2)}</td>
                                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-400">{order.price.toFixed(5)}</td>
                                            <td className="p-1 border-r border-zinc-800 text-right">
                                                <span className={order.sl && order.sl > 0 ? "text-red-300" : "text-zinc-600"}>
                                                    {order.sl && order.sl > 0 ? order.sl.toFixed(5) : '—'}
                                                </span>
                                            </td>
                                            <td className="p-1 border-r border-zinc-800 text-right">
                                                <span className={order.tp && order.tp > 0 ? "text-emerald-300" : "text-zinc-600"}>
                                                    {order.tp && order.tp > 0 ? order.tp.toFixed(5) : '—'}
                                                </span>
                                            </td>
                                            <td className="p-1 border-r border-zinc-800">
                                                <span className="bg-yellow-900/20 text-yellow-300 px-1 rounded-sm text-[10px]">{order.status}</span>
                                            </td>
                                            <td className="p-1 border-r border-zinc-800 text-zinc-500 text-[10px]">{formatDate(order.createdAt)}</td>
                                            <td className="p-1 pr-2">
                                                <button
                                                    onClick={() => handleCancelOrder(order.id)}
                                                    className="text-zinc-600 hover:text-white hover:bg-zinc-700 w-5 h-5 rounded flex items-center justify-center transition-colors"
                                                    title="Cancel Order"
                                                    disabled={loading}
                                                >
                                                    <X size={12} />
                                                </button>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>

            {/* Bottom Summary */}
            {activeTab === 'Positions' && positions.length > 0 && (
                <div className="bg-[#2d3436] border-t border-zinc-600 px-4 py-1.5 flex items-center gap-6 text-[11px] font-medium text-zinc-300 tabular-nums flex-shrink-0">
                    <span>
                        Total P&L: <span className={totalPnL >= 0 ? "text-[#4ade80] font-bold" : "text-[#f87171] font-bold"}>
                            {formatMoney(totalPnL)}
                        </span>
                    </span>
                    <span>Total Margin: <span className="text-white">{formatMoney(totalMargin)}</span></span>
                    <span>Total Swap: <span className="text-white">{formatMoney(totalSwap)}</span></span>
                    <div className="flex-1"></div>
                    <span className="text-zinc-500">{positions.length} position{positions.length !== 1 ? 's' : ''}</span>
                </div>
            )}

            {/* Close All Confirmation Modal */}
            {showCloseAllConfirm && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
                    <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg shadow-2xl p-6 max-w-md">
                        <div className="flex items-center gap-3 mb-4">
                            <AlertTriangle className="w-6 h-6 text-yellow-400" />
                            <h3 className="text-lg font-semibold text-zinc-100">Close All Positions?</h3>
                        </div>
                        <p className="text-sm text-zinc-400 mb-6">
                            This will close all {positions.length} open positions. This action cannot be undone.
                        </p>
                        <div className="flex gap-3">
                            <button
                                onClick={() => setShowCloseAllConfirm(false)}
                                className="flex-1 px-4 py-2 text-sm font-medium bg-zinc-800 hover:bg-zinc-700 text-zinc-300 rounded transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleCloseAll}
                                disabled={loading}
                                className="flex-1 px-4 py-2 text-sm font-medium bg-red-500 hover:bg-red-600 text-white rounded transition-colors disabled:opacity-50"
                            >
                                {loading ? 'Closing...' : 'Close All'}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Partial Close Modal */}
            {showPartialClose && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
                    <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg shadow-2xl p-6 max-w-md">
                        <div className="flex items-center justify-between mb-4">
                            <h3 className="text-lg font-semibold text-zinc-100">Partial Close</h3>
                            <button onClick={() => setShowPartialClose(null)} className="text-zinc-500 hover:text-zinc-300">
                                <X size={20} />
                            </button>
                        </div>
                        <p className="text-sm text-zinc-400 mb-4">
                            Enter volume to close (max: {showPartialClose.maxVolume.toFixed(2)} lots)
                        </p>
                        <input
                            type="number"
                            step="0.01"
                            min="0.01"
                            max={showPartialClose.maxVolume}
                            value={partialCloseVolume}
                            onChange={(e) => setPartialCloseVolume(e.target.value)}
                            placeholder="0.00"
                            className="w-full bg-zinc-800 text-zinc-100 px-3 py-2 rounded text-sm font-mono mb-4 focus:outline-none focus:ring-2 focus:ring-blue-500"
                            autoFocus
                        />
                        <div className="flex gap-3">
                            <button
                                onClick={() => setShowPartialClose(null)}
                                className="flex-1 px-4 py-2 text-sm font-medium bg-zinc-800 hover:bg-zinc-700 text-zinc-300 rounded transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handlePartialClose}
                                disabled={loading}
                                className="flex-1 px-4 py-2 text-sm font-medium bg-blue-500 hover:bg-blue-600 text-white rounded transition-colors disabled:opacity-50"
                            >
                                {loading ? 'Closing...' : 'Partial Close'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
