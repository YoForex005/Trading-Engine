import React from 'react';
import { Activity, CheckCircle, RefreshCw, XCircle, Power } from 'lucide-react';

interface LPStatus {
    status: string;
    type: string;
    streaming?: boolean;
    symbols?: string[];
}

interface LPStatusViewProps {
    status: Record<string, LPStatus>;
    onToggleLP?: (lpId: string, active: boolean) => void;
}

export default function LPStatusView({ status, onToggleLP }: LPStatusViewProps) {
    const handleToggle = async (lpId: string, currentStatus: string) => {
        const newActive = currentStatus !== 'CONNECTED';

        try {
            const res = await fetch(`http://localhost:8080/admin/lp/${lpId}/toggle`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ active: newActive })
            });

            if (res.ok) {
                onToggleLP?.(lpId, newActive);
                window.location.reload(); // Refresh to see updated status
            }
        } catch (e) {
            console.error('Failed to toggle LP', e);
        }
    };

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-bold text-gray-100 flex items-center gap-2">
                    <Activity className="w-5 h-5 text-blue-400" />
                    Liquidity Providers
                </h2>
            </div>

            {Object.entries(status || {}).length === 0 ? (
                <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-8 text-center">
                    <Activity className="w-12 h-12 mx-auto text-zinc-600 mb-4" />
                    <h3 className="font-semibold mb-2">No LP Connections</h3>
                    <p className="text-sm text-zinc-500">Running in pure B-Book mode (internal execution only)</p>
                </div>
            ) : (
                Object.entries(status || {}).map(([id, s]: [string, LPStatus]) => (
                    <div key={id} className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                                {s.status === 'CONNECTED' ? (
                                    <CheckCircle className="text-emerald-500" />
                                ) : s.status === 'CONNECTING' ? (
                                    <RefreshCw className="text-yellow-500 animate-spin" />
                                ) : (
                                    <XCircle className="text-red-500" />
                                )}
                                <div>
                                    <h3 className="font-semibold">{id}</h3>
                                    <p className="text-xs text-zinc-500">{s.type}</p>
                                    {s.symbols && s.symbols.length > 0 && (
                                        <p className="text-xs text-blue-400 mt-1">
                                            Symbols: {s.symbols.join(', ')}
                                        </p>
                                    )}
                                </div>
                            </div>
                            <div className="flex items-center gap-3">
                                <span className={`px-3 py-1 rounded-full text-xs ${s.status === 'CONNECTED'
                                        ? 'bg-emerald-500/20 text-emerald-400'
                                        : 'bg-zinc-700 text-zinc-400'
                                    }`}>
                                    {s.status}
                                </span>
                                <button
                                    onClick={() => handleToggle(id, s.status)}
                                    className={`p-2 rounded-lg transition-colors ${s.status === 'CONNECTED'
                                            ? 'bg-red-500/10 text-red-400 hover:bg-red-500/20'
                                            : 'bg-green-500/10 text-green-400 hover:bg-green-500/20'
                                        }`}
                                    title={s.status === 'CONNECTED' ? 'Disconnect' : 'Connect'}
                                >
                                    <Power className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    </div>
                ))
            )}

            <div className="text-xs text-zinc-500 mt-4">
                * Disconnecting an LP will stop price feeds from that provider. Connected clients may experience gaps.
            </div>
        </div>
    );
}
