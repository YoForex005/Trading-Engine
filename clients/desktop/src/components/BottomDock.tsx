import { useState, useEffect, useMemo, useRef } from 'react';
import {
    Briefcase,
    PieChart,
    History,
    Newspaper,
    Mail,
    Calendar,
    Building2,
    Bell,
    BookOpen,
    Code2,
    Cpu,
    FileText,
    MoreVertical,
    Check,
    Database
} from 'lucide-react';
import { CalendarTab } from './layout/CalendarTab';
import { CodeBaseTab, ArticlesTab, ExpertsTab } from './layout/ResearchTabs';
import { JournalTab } from './layout/JournalTab';
import { HistoryDownloader } from './HistoryDownloader';

interface BottomDockProps {
    height: number;
    onHeightChange: (height: number) => void;
    account: any;
    positions: any[];
    orders: any[];
    history: any[];
    ledger: any[];
    onClosePosition: (id: number, volume?: number) => void;
    onModifyPosition: (id: number, sl: number, tp: number) => void;
    onCancelOrder: (id: number) => void;
    onCloseBulk: (type: 'ALL' | 'WINNERS' | 'LOSERS', symbol?: string) => void;
}

type TabType = 'Trade' | 'Exposure' | 'History' | 'News' | 'Mailbox' | 'Calendar' | 'Company' | 'Alerts' | 'Articles' | 'Code Base' | 'Experts' | 'Journal' | 'Historical Data';

const TABS: { id: TabType; icon?: any }[] = [
    { id: 'Trade' },
    { id: 'Exposure' },
    { id: 'History' },
    { id: 'Historical Data' },
    { id: 'News' },
    { id: 'Mailbox' },
    { id: 'Calendar' },
    { id: 'Company' },
    { id: 'Alerts' },
    { id: 'Articles' },
    { id: 'Code Base' },
    { id: 'Experts' },
    { id: 'Journal' },
];

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
    const [activeTab, setActiveTab] = useState<TabType>('Trade');
    const [isResizing, setIsResizing] = useState(false);

    // Helpers
    const formatMoney = (val: number | undefined | null) => {
        const v = val ?? 0;
        if (isNaN(v)) return '0.00';
        return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v);
    };

    const formatDate = (dateStr: string) => {
        try {
            return new Date(dateStr).toLocaleString('en-US', {
                year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false
            }).replace(',', '');
        } catch {
            return dateStr;
        }
    };

    // Resize Handler
    useEffect(() => {
        const handleMouseMove = (e: MouseEvent) => {
            if (isResizing) {
                const newHeight = window.innerHeight - e.clientY;
                onHeightChange(Math.max(150, Math.min(newHeight, 800)));
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
            className="flex-shrink-0 bg-[#1e1e1e] border-t border-zinc-700 flex flex-col relative font-sans text-[12px] shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.2)] select-none"
            style={{ height: `${height}px`, zIndex: 30 }}
        >
            {/* Resize Handle */}
            <div
                className="h-1 cursor-ns-resize hover:bg-emerald-500/50 transition-colors w-full absolute top-0 left-0 z-50"
                onMouseDown={() => setIsResizing(true)}
            />

            {/* Content Area */}
            <div className="flex-1 overflow-hidden relative bg-[#1e1e1e] tabular-nums flex flex-col">
                {activeTab === 'Trade' && (
                    <TradeTab
                        positions={positions}
                        account={account}
                        onClosePosition={onClosePosition}
                        onModifyPosition={onModifyPosition}
                        formatMoney={formatMoney}
                        formatDate={formatDate}
                    />
                )}
                {activeTab === 'Exposure' && (
                    <ExposureTab positions={positions} formatMoney={formatMoney} />
                )}
                {activeTab === 'History' && (
                    <HistoryTab history={history} formatMoney={formatMoney} formatDate={formatDate} />
                )}

                {/* Implemented Tabs from Phase 13 & 14 */}
                {activeTab === 'Calendar' && (
                    <CalendarTab />
                )}
                {activeTab === 'Code Base' && (
                    <CodeBaseTab />
                )}
                {activeTab === 'Articles' && (
                    <ArticlesTab />
                )}
                {activeTab === 'Experts' && (
                    <ExpertsTab />
                )}
                {activeTab === 'Journal' && (
                    <JournalTab />
                )}
                {activeTab === 'Historical Data' && (
                    <HistoryDownloader />
                )}

                {/* Placeholder Tabs */}
                {['News', 'Mailbox', 'Company', 'Alerts'].includes(activeTab) && (
                    <PlaceholderTab title={activeTab} />
                )}
            </div>

            {/* Bottom Tabs */}
            <div className="flex items-center bg-[#1e1e1e] border-t border-zinc-700 select-none overflow-x-auto scrollbar-thin scrollbar-thumb-zinc-700">
                {TABS.map((tab) => (
                    <button
                        key={tab.id}
                        onClick={() => setActiveTab(tab.id)}
                        className={`px-3 py-1 text-[11px] font-medium transition-colors whitespace-nowrap border-r border-zinc-700/50 relative
                        ${activeTab === tab.id
                                ? 'text-zinc-100 bg-[#2d3436] shadow-[inset_0_-2px_0_0_#3b82f6]'
                                : 'text-zinc-500 hover:text-zinc-300 hover:bg-[#2d3436]/50'}`}
                    >
                        {tab.id}
                    </button>
                ))}
            </div>
        </div>
    );
}

// --- TAB COMPONENTS ---

function TradeTab({ positions, account, onClosePosition, onModifyPosition, formatMoney, formatDate }: any) {
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

    return (
        <div className="h-full flex flex-col relative bg-[#1e1e1e]">
            <div className="flex-1 overflow-auto scrollbar-thin scrollbar-thumb-zinc-700">
                <table className="w-full text-left border-collapse">
                    <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                        <tr>
                            <th className="p-1 pl-2 border-r border-zinc-700 w-20">Ticket</th>
                            <th className="p-1 border-r border-zinc-700 w-28">Time</th>
                            <th className="p-1 border-r border-zinc-700 w-12">Type</th>
                            <th className="p-1 border-r border-zinc-700 w-16 text-right">Volume</th>
                            <th className="p-1 border-r border-zinc-700 w-20">Symbol</th>
                            <th className="p-1 border-r border-zinc-700 w-20 text-right">Price</th>
                            <th className="p-1 border-r border-zinc-700 w-20 text-right">S / L</th>
                            <th className="p-1 border-r border-zinc-700 w-20 text-right">T / P</th>
                            <th className="p-1 border-r border-zinc-700 w-20 text-right">Price</th>
                            <th className="p-1 border-r border-zinc-700 w-16 text-right">Swap</th>
                            <th className="p-1 pr-4 text-right">Profit</th>
                            <th className="w-6"></th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800 text-zinc-300 font-mono text-[11px]">
                        {positions.map((pos: any, i: number) => (
                            <tr key={pos.id} className={`${i % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'} hover:bg-[#2d3436] transition-colors group`}>
                                <td className="p-1 pl-2 border-r border-zinc-800 text-zinc-500">{pos.id}</td>
                                <td className="p-1 border-r border-zinc-800 text-zinc-400 whitespace-nowrap text-[10px]">{formatDate(pos.openTime)}</td>
                                <td className={`p-1 border-r border-zinc-800 font-bold text-[10px] ${pos.side === 'BUY' ? 'text-blue-400' : 'text-red-400'}`}>{pos.side}</td>
                                <td className="p-1 border-r border-zinc-800 text-right font-medium text-zinc-200">{pos.volume.toFixed(2)}</td>
                                <td className="p-1 border-r border-zinc-800 font-bold text-zinc-100">{pos.symbol}</td>
                                <td className="p-1 border-r border-zinc-800 text-right text-zinc-400">{pos.openPrice.toFixed(5)}</td>

                                {/* SL Edit */}
                                <td className="p-1 border-r border-zinc-800 text-right" onDoubleClick={() => startEditing(pos)}>
                                    {editingId === pos.id ? (
                                        <input
                                            className="w-full bg-zinc-800 text-white text-right px-1 text-[10px]"
                                            value={editSL}
                                            onChange={e => setEditSL(e.target.value)}
                                            onBlur={() => saveEdit(pos)}
                                            onKeyDown={e => e.key === 'Enter' && saveEdit(pos)}
                                            autoFocus
                                        />
                                    ) : (
                                        <span className={pos.sl > 0 ? "text-red-300 bg-red-900/20 px-1 rounded-sm" : "text-zinc-600"}>
                                            {pos.sl > 0 ? pos.sl.toFixed(5) : ''}
                                        </span>
                                    )}
                                </td>

                                {/* TP Edit */}
                                <td className="p-1 border-r border-zinc-800 text-right" onDoubleClick={() => startEditing(pos)}>
                                    {editingId === pos.id ? (
                                        <input
                                            className="w-full bg-zinc-800 text-white text-right px-1 text-[10px]"
                                            value={editTP}
                                            onChange={e => setEditTP(e.target.value)}
                                            onBlur={() => saveEdit(pos)}
                                            onKeyDown={e => e.key === 'Enter' && saveEdit(pos)}
                                        />
                                    ) : (
                                        <span className={pos.tp > 0 ? "text-emerald-300 bg-emerald-900/20 px-1 rounded-sm" : "text-zinc-600"}>
                                            {pos.tp > 0 ? pos.tp.toFixed(5) : ''}
                                        </span>
                                    )}
                                </td>

                                <td className="p-1 border-r border-zinc-800 text-right text-zinc-300">{pos.currentPrice.toFixed(5)}</td>
                                <td className="p-1 border-r border-zinc-800 text-right text-zinc-500 text-[10px]">{formatMoney(pos.swap)}</td>
                                <td className={`p-1 text-right font-bold pr-4 ${pos.unrealizedPnL >= 0 ? 'text-[#4ade80]' : 'text-[#f87171]'}`}>
                                    {formatMoney(pos.unrealizedPnL)}
                                </td>
                                <td className="p-1 text-center">
                                    <button
                                        onClick={() => onClosePosition(pos.id)}
                                        className="text-zinc-600 hover:text-white hover:bg-zinc-700 w-4 h-4 rounded flex items-center justify-center transition-colors"
                                    >
                                        ×
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {/* Footer Summary Row */}
            {account && (
                <div className="bg-[#2d3436] border-t border-zinc-600 w-full overflow-hidden text-[11px] font-medium text-zinc-300 flex items-center px-4 py-1 gap-6 tabular-nums">
                    <span>Balance: <span className="text-white font-bold">{formatMoney(account.balance)}</span></span>
                    <span>Equity: <span className="text-white font-bold">{formatMoney(account.equity)}</span></span>
                    <span>Margin: <span className="text-white">{formatMoney(account.margin)}</span></span>
                    <span>Free Margin: <span className="text-white">{formatMoney(account.freeMargin)}</span></span>
                    <span>Margin Level: <span className="text-white">{account.marginLevel?.toFixed(2)}%</span></span>
                    <div className="flex-1"></div>
                    <span>Total Profit: <span className={account.unrealizedPL >= 0 ? "text-[#4ade80] font-bold" : "text-[#f87171] font-bold"}>{formatMoney(account.unrealizedPL)}</span></span>
                </div>
            )}
        </div>
    );
}

function ExposureTab({ positions, formatMoney }: any) {
    // Calculate exposure
    const exposure = useMemo(() => {
        const map: Record<string, { volume: number, rate: number, usdValue: number, type: 'Long' | 'Short' | 'Net' }> = {};
        positions.forEach((p: any) => {
            const asset = p.symbol.substring(0, 3); // Simple 3-char assumption or use currency base
            if (!map[asset]) map[asset] = { volume: 0, rate: 1.0, usdValue: 0, type: 'Net' };
            map[asset].volume += p.volume * (p.side === 'BUY' ? 1 : -1);
            // Mock rate/USD calc
            map[asset].usdValue += (p.volume * 100000) * (p.side === 'BUY' ? 1 : -1);
        });

        // Add fake USD/Balance exposure
        map['USD'] = { volume: 12308, rate: 1.0, usdValue: 12308, type: 'Long' };

        return Object.entries(map).map(([asset, data]) => ({ asset, ...data }));
    }, [positions]);

    return (
        <div className="h-full flex bg-[#1e1e1e]">
            {/* Table */}
            <div className="flex-1 overflow-auto border-r border-zinc-700">
                <table className="w-full text-left border-collapse">
                    <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                        <tr>
                            <th className="p-1 pl-2 border-r border-zinc-700 w-24">Assets</th>
                            <th className="p-1 border-r border-zinc-700 w-24 text-right">Volume</th>
                            <th className="p-1 border-r border-zinc-700 w-24 text-right">Rate</th>
                            <th className="p-1 border-r border-zinc-700 w-24 text-right">USD</th>
                            <th className="p-1 text-center">Graph</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800 text-zinc-300 font-mono text-[11px]">
                        {exposure.map((exp: any, i: number) => (
                            <tr key={exp.asset} className={`${i % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'}`}>
                                <td className="p-1 pl-2 border-r border-zinc-800 font-bold text-[#ebc034]">{exp.asset}</td>
                                <td className="p-1 border-r border-zinc-800 text-right">{exp.volume.toFixed(2)}</td>
                                <td className="p-1 border-r border-zinc-800 text-right">{exp.rate.toFixed(5)}</td>
                                <td className="p-1 border-r border-zinc-800 text-right">{formatMoney(exp.usdValue)}</td>
                                <td className="p-1 px-4 align-middle">
                                    <div className="w-full h-3 bg-zinc-800 rounded-sm overflow-hidden flex">
                                        {exp.usdValue > 0 && <div className="h-full bg-blue-500" style={{ width: '60%' }}></div>}
                                        {exp.usdValue < 0 && <div className="h-full bg-red-500" style={{ width: '40%' }}></div>}
                                    </div>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {/* Pie Chart Area */}
            <div className="w-64 bg-[#1e1e1e] flex flex-col items-center justify-center p-4">
                <div className="text-[10px] text-zinc-400 mb-2 font-bold uppercase tracking-wider">Long Positions</div>
                <div className="w-32 h-32 rounded-full border-4 border-blue-500 flex items-center justify-center relative shadow-[0_0_15px_rgba(59,130,246,0.5)]">
                    {/* Placeholder for SVG Pie */}
                    <span className="text-[10px] text-zinc-500">USD</span>
                </div>
            </div>
        </div>
    );
}

function HistoryTab({ history, formatMoney, formatDate }: any) {
    if (!history || history.length === 0) return <div className="p-8 text-center text-zinc-600 text-xs italic">No history available</div>;

    return (
        <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 bg-[#1e1e1e]">
            <table className="w-full text-left border-collapse">
                <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                    <tr>
                        <th className="p-1 pl-2 border-r border-zinc-700 w-20">Ticket</th>
                        <th className="p-1 border-r border-zinc-700 w-28">Time</th>
                        <th className="p-1 border-r border-zinc-700 w-12">Type</th>
                        <th className="p-1 border-r border-zinc-700 w-16 text-right">Volume</th>
                        <th className="p-1 border-r border-zinc-700 w-20">Symbol</th>
                        <th className="p-1 border-r border-zinc-700 w-20 text-right">Price</th>
                        <th className="p-1 border-r border-zinc-700 w-20 text-right">S/L</th>
                        <th className="p-1 border-r border-zinc-700 w-20 text-right">T/P</th>
                        <th className="p-1 border-r border-zinc-700 w-28 text-right">Time</th>
                        <th className="p-1 border-r border-zinc-700 w-20 text-right">Price</th>
                        <th className="p-1 border-r border-zinc-700 w-16 text-right">Swap</th>
                        <th className="p-1 pr-4 text-right">Profit</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800 text-zinc-300 font-mono text-[11px]">
                    {[...history].reverse().map((trade: any, i: number) => (
                        <tr key={trade.id} className={`${i % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'} hover:bg-[#2d3436] transition-colors`}>
                            <td className="p-1 pl-2 border-r border-zinc-800 text-zinc-500">{trade.id}</td>
                            <td className="p-1 border-r border-zinc-800 text-zinc-400 whitespace-nowrap text-[10px]">{formatDate(trade.openTime)}</td>
                            <td className={`p-1 border-r border-zinc-800 font-bold text-[10px] ${trade.side === 'BUY' ? 'text-blue-400' : 'text-red-400'}`}>{trade.side}</td>
                            <td className="p-1 border-r border-zinc-800 text-right font-medium text-zinc-200">{trade.volume.toFixed(2)}</td>
                            <td className="p-1 border-r border-zinc-800 font-bold text-zinc-100">{trade.symbol}</td>
                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-400">{trade.openPrice.toFixed(5)}</td>
                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-500">{trade.sl?.toFixed(5)}</td>
                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-500">{trade.tp?.toFixed(5)}</td>
                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-400 text-[10px]">{formatDate(trade.closeTime)}</td>
                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-400">{trade.closePrice.toFixed(5)}</td>
                            <td className="p-1 border-r border-zinc-800 text-right text-zinc-500 text-[10px]">{formatMoney(trade.swap)}</td>
                            <td className={`p-1 text-right font-bold pr-4 ${trade.profit >= 0 ? 'text-[#4ade80]' : 'text-[#f87171]'}`}>
                                {formatMoney(trade.profit)}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}



function PlaceholderTab({ title }: { title: string }) {
    return (
        <div className="flex-1 flex flex-col items-center justify-center text-zinc-600 bg-[#1e1e1e]">
            <Briefcase size={32} className="mb-2 opacity-20" />
            <div className="text-xs font-medium">RTX5 {title}</div>
            <div className="text-[10px]">No active items to display</div>
        </div>
    );
}
