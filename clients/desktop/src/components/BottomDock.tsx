import { useState, useEffect } from 'react';

interface BottomDockProps {
    height: number;
    onHeightChange: (height: number) => void;
    account: any;
    positions: any[];
    orders: any[]; // Kept for future use
    history: any[]; // New
    ledger: any[]; // New
    onClosePosition: (id: number, volume?: number) => void;
    onModifyPosition: (id: number, sl: number, tp: number) => void;
    onCancelOrder: (id: number) => void; // Kept for future use
    onCloseBulk: (type: 'ALL' | 'WINNERS' | 'LOSERS', symbol?: string) => void;
}

type TabType = 'trade' | 'history' | 'journal';

export function BottomDock({
    height,
    onHeightChange,
    account,
    positions,
    history,
    ledger,
    onClosePosition,
    onModifyPosition,
}: BottomDockProps) {
    const [activeTab, setActiveTab] = useState<TabType>('trade');
    const [isResizing, setIsResizing] = useState(false);

    // Formatting helpers
    const formatMoney = (val: number | undefined | null) => {
        const v = val ?? 0;
        if (isNaN(v)) return '$0.00';
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(v);
    };

    const formatDate = (dateStr: string) => {
        try {
            return new Date(dateStr).toLocaleString('en-US', {
                month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit'
            });
        } catch {
            return dateStr;
        }
    };

    // Resize handler
    useEffect(() => {
        const handleMouseMove = (e: MouseEvent) => {
            if (isResizing) {
                const newHeight = window.innerHeight - e.clientY;
                onHeightChange(Math.max(200, Math.min(newHeight, 600)));
            }
        };
        const handleMouseUp = () => setIsResizing(false);

        if (isResizing) {
            window.addEventListener('mousemove', handleMouseMove);
            window.addEventListener('mouseup', handleMouseUp);
        }
        return () => {
            window.removeEventListener('mousemove', handleMouseMove);
            window.removeEventListener('mouseup', handleMouseUp);
        };
    }, [isResizing, onHeightChange]);

    return (
        <div
            className="flex-shrink-0 bg-[#1e222d] border-t border-[#2a2e39] flex flex-col relative"
            style={{ height: `${height}px`, zIndex: 30 }}
        >
            {/* Resize Handle */}
            <div
                className="h-1 cursor-ns-resize hover:bg-blue-500/50 transition-colors w-full absolute top-0 left-0 z-50"
                onMouseDown={() => setIsResizing(true)}
            />

            {/* Header Row: Tabs + Account Summary */}
            <div className="flex items-center justify-between border-b border-[#2a2e39] bg-[#131722] px-2 h-9 select-none">
                {/* Tabs */}
                <div className="flex items-center gap-1">
                    {['trade', 'history', 'journal'].map((tab) => (
                        <button
                            key={tab}
                            onClick={() => setActiveTab(tab as TabType)}
                            className={`px-3 py-1.5 text-xs font-semibold uppercase tracking-wider rounded-t-sm transition-all
                            ${activeTab === tab
                                    ? 'text-blue-400 bg-[#1e222d] border-t-2 border-t-blue-500'
                                    : 'text-zinc-500 hover:text-zinc-300 hover:bg-[#1e222d]/50 border-t-2 border-transparent'}`}
                        >
                            {tab}
                        </button>
                    ))}
                </div>

                {/* Account Summary (Top Bar) */}
                {account && (
                    <div className="flex items-center gap-4 text-xs font-mono mr-2">
                        <div className="flex items-center gap-2">
                            <span className="text-zinc-500">Bal:</span>
                            <span className="text-zinc-200">{formatMoney(account.balance)}</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className="text-zinc-500">Eq:</span>
                            <span className="text-zinc-200">{formatMoney(account.equity)}</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className="text-zinc-500">Marg:</span>
                            <span className={`${account.marginLevel < 100 ? 'text-red-500' : 'text-emerald-400'}`}>
                                {account.marginLevel?.toFixed(0)}%
                            </span>
                        </div>
                        <div className="flex items-center gap-2 pl-2 border-l border-zinc-700">
                            <span className="text-zinc-500">P/L:</span>
                            <span className={`font-bold ${account.unrealizedPL >= 0 ? 'text-emerald-400' : 'text-red-500'}`}>
                                {formatMoney(account.unrealizedPL)}
                            </span>
                        </div>
                    </div>
                )}
            </div>

            {/* Content Area */}
            <div className="flex-1 overflow-hidden relative bg-[#1e222d] tabular-nums">
                {activeTab === 'trade' && (
                    <TradeTab
                        positions={positions}
                        onClosePosition={onClosePosition}
                        onModifyPosition={onModifyPosition}
                        formatMoney={formatMoney}
                        formatDate={formatDate}
                    />
                )}
                {activeTab === 'history' && (
                    <HistoryTab history={history} formatMoney={formatMoney} formatDate={formatDate} />
                )}
                {activeTab === 'journal' && (
                    <JournalTab ledger={ledger} formatMoney={formatMoney} formatDate={formatDate} />
                )}
            </div>
        </div>
    );
}

function HistoryTab({ history, formatMoney, formatDate }: any) {
    if (!history || history.length === 0) {
        return <div className="p-12 text-center text-zinc-600 italic text-xs">No closed trades history</div>;
    }

    return (
        <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 scrollbar-track-transparent">
            <table className="w-full text-left border-collapse text-xs">
                <thead className="sticky top-0 bg-[#1e222d] text-zinc-400 z-10 font-semibold shadow-sm">
                    <tr>
                        <th className="p-2 pl-4 border-b border-[#2a2e39] w-16">ID</th>
                        <th className="p-2 border-b border-[#2a2e39] w-36">Open Time</th>
                        <th className="p-2 border-b border-[#2a2e39] w-36">Close Time</th>
                        <th className="p-2 border-b border-[#2a2e39] w-16">Type</th>
                        <th className="p-2 border-b border-[#2a2e39] w-16">Vol</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Symbol</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Open Price</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Close Price</th>
                        <th className="p-2 border-b border-[#2a2e39] w-20">Swap</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Profit</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-[#2a2e39] text-zinc-300">
                    {[...history].reverse().map((trade: any) => (
                        <tr key={trade.id} className="hover:bg-[#2a2e39]/80 bg-[#1e222d] transition-colors">
                            <td className="p-2 pl-4 text-zinc-500">{trade.id}</td>
                            <td className="p-2 text-zinc-500 whitespace-nowrap">{formatDate(trade.openTime)}</td>
                            <td className="p-2 text-zinc-500 whitespace-nowrap">{formatDate(trade.closeTime)}</td>
                            <td className={`p-2 font-bold ${trade.side === 'BUY' ? 'text-emerald-400' : 'text-red-400'}`}>{trade.side}</td>
                            <td className="p-2 font-mono text-zinc-200">{trade.volume}</td>
                            <td className="p-2 font-semibold text-white">{trade.symbol}</td>
                            <td className="p-2 text-zinc-400">{trade.openPrice.toFixed(5)}</td>
                            <td className="p-2 text-zinc-400">{trade.closePrice.toFixed(5)}</td>
                            <td className="p-2 text-zinc-500">{formatMoney(trade.swap)}</td>
                            <td className={`p-2 font-bold ${trade.profit >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                                {formatMoney(trade.profit)}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}

function JournalTab({ ledger, formatMoney, formatDate }: any) {
    if (!ledger || ledger.length === 0) {
        return <div className="p-12 text-center text-zinc-600 italic text-xs">No ledger entries</div>;
    }

    return (
        <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 scrollbar-track-transparent">
            <table className="w-full text-left border-collapse text-xs">
                <thead className="sticky top-0 bg-[#1e222d] text-zinc-400 z-10 font-semibold shadow-sm">
                    <tr>
                        <th className="p-2 pl-4 border-b border-[#2a2e39] w-16">ID</th>
                        <th className="p-2 border-b border-[#2a2e39] w-36">Time</th>
                        <th className="p-2 border-b border-[#2a2e39] w-32">Type</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Amount</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Balance</th>
                        <th className="p-2 border-b border-[#2a2e39]">Description</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-[#2a2e39] text-zinc-300">
                    {[...ledger].reverse().map((entry: any) => (
                        <tr key={entry.id} className="hover:bg-[#2a2e39]/80 bg-[#1e222d] transition-colors">
                            <td className="p-2 pl-4 text-zinc-500">{entry.id}</td>
                            <td className="p-2 text-zinc-500 whitespace-nowrap">{formatDate(entry.time)}</td>
                            <td className="p-2 font-semibold text-blue-300">{entry.type}</td>
                            <td className={`p-2 font-bold ${entry.amount >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                                {formatMoney(entry.amount)}
                            </td>
                            <td className="p-2 font-mono text-zinc-200">{formatMoney(entry.balanceAfter)}</td>
                            <td className="p-2 text-zinc-400">{entry.description}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}

function TradeTab({ positions, onClosePosition, onModifyPosition, formatMoney }: any) {
    // Editing state for SL/TP inline
    const [editingId, setEditingId] = useState<number | null>(null);
    const [editSL, setEditSL] = useState("");
    const [editTP, setEditTP] = useState("");

    const startEditing = (pos: any) => {
        setEditingId(pos.id);
        setEditSL(pos.sl || "");
        setEditTP(pos.tp || "");
    };

    const saveEdit = (pos: any) => {
        onModifyPosition(pos.id, parseFloat(editSL) || 0, parseFloat(editTP) || 0);
        setEditingId(null);
    };

    const formatDate = (dateStr: string) => {
        try {
            return new Date(dateStr).toLocaleString('en-US', {
                month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit'
            });
        } catch {
            return dateStr;
        }
    };

    return (
        <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 scrollbar-track-transparent">
            <table className="w-full text-left border-collapse text-xs">
                <thead className="sticky top-0 bg-[#1e222d] text-zinc-400 z-10 font-semibold shadow-sm">
                    <tr>
                        <th className="p-2 pl-4 border-b border-[#2a2e39] w-16">ID</th>
                        <th className="p-2 border-b border-[#2a2e39] w-36">Time</th>
                        <th className="p-2 border-b border-[#2a2e39] w-16">Type</th>
                        <th className="p-2 border-b border-[#2a2e39] w-16">Vol</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Symbol</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Price</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24 hover:text-blue-400 transition-colors cursor-help" title="Stop Loss">S/L</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24 hover:text-blue-400 transition-colors cursor-help" title="Take Profit">T/P</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Current</th>
                        <th className="p-2 border-b border-[#2a2e39] w-20">Swap</th>
                        <th className="p-2 border-b border-[#2a2e39] w-24">Profit</th>
                        <th className="p-2 border-b border-[#2a2e39] w-16 text-center"></th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-[#2a2e39] text-zinc-300">
                    {positions.map((pos: any) => (
                        <tr key={pos.id} className="hover:bg-[#2a2e39]/80 bg-[#1e222d] transition-colors group">
                            <td className="p-2 pl-4 text-zinc-500">{pos.id}</td>
                            <td className="p-2 text-zinc-500 whitespace-nowrap">
                                {formatDate(pos.openTime)}
                            </td>
                            <td className={`p-2 font-bold ${pos.side === 'BUY' ? 'text-emerald-400' : 'text-red-400'}`}>
                                {pos.side}
                            </td>
                            <td className="p-2 font-mono text-zinc-200">{pos.volume}</td>
                            <td className="p-2 font-semibold text-white">{pos.symbol}</td>
                            <td className="p-2 text-zinc-400">{pos.openPrice.toFixed(5)}</td>

                            {/* SL Editable Cell */}
                            <td className="p-2 relative">
                                {editingId === pos.id ? (
                                    <input
                                        type="number"
                                        step="0.00001"
                                        autoFocus
                                        className="w-20 bg-[#131722] border border-blue-500 px-1 py-0.5 text-xs text-white focus:outline-none rounded"
                                        value={editSL}
                                        onChange={(e) => setEditSL(e.target.value)}
                                        onKeyDown={(e) => e.key === 'Enter' && saveEdit(pos)}
                                        onBlur={() => saveEdit(pos)}
                                    />
                                ) : (
                                    <span
                                        onClick={() => startEditing(pos)}
                                        className={`cursor-pointer border-b border-dashed border-zinc-700 hover:border-blue-500 hover:text-blue-400 transition-colors
                                            ${pos.sl > 0 ? 'text-zinc-300' : 'text-zinc-600'}`}
                                    >
                                        {pos.sl > 0 ? pos.sl.toFixed(5) : 'Add'}
                                    </span>
                                )}
                            </td>

                            {/* TP Editable Cell */}
                            <td className="p-2 relative">
                                {editingId === pos.id ? (
                                    <input
                                        type="number"
                                        step="0.00001"
                                        className="w-20 bg-[#131722] border border-blue-500 px-1 py-0.5 text-xs text-white focus:outline-none rounded mt-1"
                                        value={editTP}
                                        onChange={(e) => setEditTP(e.target.value)}
                                        onKeyDown={(e) => e.key === 'Enter' && saveEdit(pos)}
                                        onClick={(e) => e.stopPropagation()}
                                    />
                                ) : (
                                    <span
                                        onClick={() => startEditing(pos)}
                                        className={`cursor-pointer border-b border-dashed border-zinc-700 hover:border-blue-500 hover:text-blue-400 transition-colors
                                            ${pos.tp > 0 ? 'text-zinc-300' : 'text-zinc-600'}`}
                                    >
                                        {pos.tp > 0 ? pos.tp.toFixed(5) : 'Add'}
                                    </span>
                                )}
                            </td>

                            <td className="p-2 text-zinc-300">{pos.currentPrice.toFixed(5)}</td>
                            <td className="p-2 text-zinc-500">{formatMoney(pos.swap)}</td>
                            <td className={`p-2 font-bold ${pos.unrealizedPnL >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                                {formatMoney(pos.unrealizedPnL)}
                            </td>
                            <td className="p-2 text-center flex items-center justify-end gap-1">
                                <button
                                    onClick={() => onClosePosition(pos.id, pos.volume / 2)}
                                    className="text-zinc-500 hover:text-blue-400 hover:bg-blue-500/10 rounded px-1.5 py-0.5 text-[10px] transition-all border border-transparent hover:border-blue-500/30"
                                    title="Close Half (50%)"
                                >
                                    ½
                                </button>
                                <button
                                    onClick={() => onClosePosition(pos.id)}
                                    className="text-zinc-500 hover:text-white hover:bg-red-500/80 rounded p-1 transition-all"
                                    title="Close Position"
                                >
                                    ✕
                                </button>
                            </td>
                        </tr>
                    ))}
                    {positions.length === 0 && (
                        <tr>
                            <td colSpan={12} className="p-12 text-center text-zinc-600 italic">
                                No open positions active
                            </td>
                        </tr>
                    )}
                </tbody>
            </table>
        </div>
    );
}
