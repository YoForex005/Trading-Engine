import { useState, useEffect } from 'react';

interface PendingOrder {
    id: string;
    symbol: string;
    side: string;
    type: string;
    subtype: string;
    volume: number;
    entryPrice?: number;
    triggerPrice?: number;
    sl?: number;
    tp?: number;
    status: string;
    createdAt: string;
    ocoLinkId?: number;
    expiryTime?: string;
}

interface PendingOrdersPanelProps {
    symbol?: string; // Filter by symbol if provided
}

export function PendingOrdersPanel({ symbol }: PendingOrdersPanelProps) {
    const [orders, setOrders] = useState<PendingOrder[]>([]);
    const [loading, setLoading] = useState(false);
    const [editingOrderId, setEditingOrderId] = useState<string | null>(null);
    const [editTrigger, setEditTrigger] = useState('');
    const [editSL, setEditSL] = useState('');
    const [editTP, setEditTP] = useState('');
    const [editVolume, setEditVolume] = useState('');

    // Fetch pending orders
    useEffect(() => {
        const fetchOrders = async () => {
            try {
                const res = await fetch('http://localhost:8080/api/orders?status=PENDING');
                if (res.ok) {
                    const data = await res.json();
                    setOrders(data || []);
                }
            } catch (err) {
                console.error('Failed to fetch pending orders:', err);
            }
        };

        fetchOrders();
        const interval = setInterval(fetchOrders, 2000);
        return () => clearInterval(interval);
    }, []);

    const handleEdit = (order: PendingOrder) => {
        setEditingOrderId(order.id);
        setEditTrigger(order.triggerPrice?.toString() || '');
        setEditSL(order.sl?.toString() || '');
        setEditTP(order.tp?.toString() || '');
        setEditVolume(order.volume?.toString() || '');
    };

    const handleSaveModify = async (orderId: string) => {
        setLoading(true);
        try {
            const modifyData: any = {};
            if (editTrigger) modifyData.triggerPrice = parseFloat(editTrigger);
            if (editSL) modifyData.sl = parseFloat(editSL);
            if (editTP) modifyData.tp = parseFloat(editTP);
            if (editVolume) modifyData.volume = parseFloat(editVolume);

            const response = await fetch(`http://localhost:8080/api/orders/modify?id=${orderId}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(modifyData),
            });

            if (response.ok) {
                const data = await response.json();
                // Update local state
                setOrders(orders.map(o => o.id === orderId ? { ...o, ...data.order } : o));
                setEditingOrderId(null);
            } else {
                const errorText = await response.text();
                console.error('Failed to modify order:', errorText);
                alert('Failed to modify order: ' + errorText);
            }
        } catch (err) {
            console.error('Failed to modify order:', err);
            alert('Failed to modify order');
        } finally {
            setLoading(false);
        }
    };

    const handleCancelEdit = () => {
        setEditingOrderId(null);
    };

    const handleCancel = async (orderId: string) => {
        if (!confirm('Are you sure you want to cancel this order?')) {
            return;
        }

        setLoading(true);
        try {
            await fetch(`http://localhost:8080/api/orders/cancel?id=${orderId}`, {
                method: 'DELETE',
            });
            setOrders(orders.filter(o => o.id !== orderId));
        } catch (err) {
            console.error('Failed to cancel order:', err);
        } finally {
            setLoading(false);
        }
    };

    const filteredOrders = symbol
        ? orders.filter(o => o.symbol === symbol)
        : orders;

    if (filteredOrders.length === 0) {
        return (
            <div className="text-xs text-zinc-500 text-center py-4">
                No pending orders
            </div>
        );
    }

    return (
        <div className="space-y-2">
            {filteredOrders.map((order) => {
                const isEditing = editingOrderId === order.id;

                return (
                    <div
                        key={order.id}
                        className="bg-zinc-800/50 border border-zinc-700 rounded-lg p-3"
                    >
                        <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center gap-2">
                                <span className={`text-xs font-semibold ${order.side === 'BUY' ? 'text-emerald-400' : 'text-red-400'
                                    }`}>
                                    {order.type}
                                </span>
                                <span className="text-xs text-zinc-400">{order.symbol}</span>
                                {order.ocoLinkId && (
                                    <span className="text-xs bg-purple-500/20 text-purple-400 px-2 py-0.5 rounded">
                                        OCO→#{order.ocoLinkId}
                                    </span>
                                )}
                            </div>
                            <div className="flex gap-2">
                                {!isEditing && (
                                    <button
                                        onClick={() => handleEdit(order)}
                                        disabled={loading}
                                        className="text-xs text-blue-400 hover:text-blue-300 disabled:opacity-50"
                                    >
                                        Modify
                                    </button>
                                )}
                                <button
                                    onClick={() => handleCancel(order.id)}
                                    disabled={loading}
                                    className="text-xs text-red-400 hover:text-red-300 disabled:opacity-50"
                                >
                                    Cancel
                                </button>
                            </div>
                        </div>

                        {isEditing ? (
                            <div className="space-y-2">
                                <div className="grid grid-cols-2 gap-2">
                                    <div>
                                        <label className="text-xs text-zinc-500">Trigger Price</label>
                                        <input
                                            type="number"
                                            value={editTrigger}
                                            onChange={(e) => setEditTrigger(e.target.value)}
                                            step="0.00001"
                                            className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200"
                                        />
                                    </div>
                                    <div>
                                        <label className="text-xs text-zinc-500">Volume</label>
                                        <input
                                            type="number"
                                            value={editVolume}
                                            onChange={(e) => setEditVolume(e.target.value)}
                                            step="0.01"
                                            className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200"
                                        />
                                    </div>
                                    <div>
                                        <label className="text-xs text-zinc-500">SL</label>
                                        <input
                                            type="number"
                                            value={editSL}
                                            onChange={(e) => setEditSL(e.target.value)}
                                            step="0.00001"
                                            className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200"
                                        />
                                    </div>
                                    <div>
                                        <label className="text-xs text-zinc-500">TP</label>
                                        <input
                                            type="number"
                                            value={editTP}
                                            onChange={(e) => setEditTP(e.target.value)}
                                            step="0.00001"
                                            className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200"
                                        />
                                    </div>
                                </div>
                                <div className="flex gap-2 justify-end">
                                    <button
                                        onClick={handleCancelEdit}
                                        className="text-xs px-3 py-1 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 rounded"
                                    >
                                        Cancel
                                    </button>
                                    <button
                                        onClick={() => handleSaveModify(order.id)}
                                        disabled={loading}
                                        className="text-xs px-3 py-1 bg-emerald-600 hover:bg-emerald-500 text-white rounded disabled:opacity-50"
                                    >
                                        Save
                                    </button>
                                </div>
                            </div>
                        ) : (
                            <>
                                <div className="grid grid-cols-3 gap-2 text-xs">
                                    <div>
                                        <span className="text-zinc-500">Volume:</span>
                                        <span className="text-zinc-300 ml-1">{order.volume}</span>
                                    </div>
                                    <div>
                                        <span className="text-zinc-500">Price:</span>
                                        <span className="text-zinc-300 ml-1">
                                            {order.entryPrice?.toFixed(5) || order.triggerPrice?.toFixed(5)}
                                        </span>
                                    </div>
                                    <div>
                                        <span className="text-zinc-500">Status:</span>
                                        <span className={`ml-1 ${order.status === 'PENDING' ? 'text-yellow-400' :
                                            order.status === 'TRIGGERED' ? 'text-emerald-400' : 'text-zinc-400'
                                            }`}>
                                            {order.status}
                                        </span>
                                    </div>
                                </div>

                                {(order.sl || order.tp) && (
                                    <div className="flex gap-4 mt-2 text-xs">
                                        {order.sl && (
                                            <div>
                                                <span className="text-zinc-500">SL:</span>
                                                <span className="text-red-400 ml-1">{order.sl}</span>
                                            </div>
                                        )}
                                        {order.tp && (
                                            <div>
                                                <span className="text-zinc-500">TP:</span>
                                                <span className="text-emerald-400 ml-1">{order.tp}</span>
                                            </div>
                                        )}
                                    </div>
                                )}

                                {order.expiryTime && (
                                    <div className="mt-2 text-xs text-orange-400">
                                        Expires: {new Date(order.expiryTime).toLocaleString()}
                                    </div>
                                )}
                            </>
                        )}
                    </div>
                );
            })}
        </div>
    );
}
