/**
 * Toolbox Bottom Panel (MT5-style)
 * Collapsible panel with tabs: Trade, History, Orders, Alerts, Journal
 */

import { useState, useEffect, useMemo } from 'react';
import {
    ChevronDown,
    ChevronUp,
    X,
    Clock,
    AlertTriangle,
    FileText,
    TrendingUp
} from 'lucide-react';
import { API_ENDPOINTS } from '../config/api';
import { useAppStore } from '../store/useAppStore';
import { AlertManager } from './AlertManager';
import { useAlertStore } from '../store/useAlertStore';
import { MarketSessions } from './MarketSessions';
import { CorrelationMatrix } from './CorrelationMatrix';
import { PositionsPanel } from './PositionsPanel';
import { StrategyTester } from './StrategyTester';
import { TradingSignals } from './TradingSignals';
import { RiskHeatmap } from './RiskHeatmap';
import { SentimentAnalysis } from './SentimentAnalysis';
import { SymbolScreener } from './SymbolScreener';
import { MultiTimeframeAnalysis } from './MultiTimeframeAnalysis';
import { PnLCalendar } from './PnLCalendar';
import { DrawingTools } from './DrawingTools';
import { TradeAnalytics } from './TradeAnalytics';
import AdvancedOrders from './AdvancedOrders';
import { MarketReplay } from './MarketReplay';
import { CopyTradingLeaderboard } from './CopyTradingLeaderboard';
import { NewsImpactAnalyzer } from './NewsImpactAnalyzer';
import { PositionSizeCalculator } from './PositionSizeCalculator';
import { EquityCurveTracker } from './EquityCurveTracker';
import { VolatilityAnalyzer } from './VolatilityAnalyzer';
import { CurrencyConverter } from './CurrencyConverter';
import { TradeJournalV2 } from './TradeJournalV2';
import { SessionHeatMap } from './SessionHeatMap';
import { RiskExposurePanel } from './RiskExposurePanel';
import { OrderFlowVisualizer } from './OrderFlowVisualizer';
import { MarketProfile } from './MarketProfile';
import { TradingSimulator } from './TradingSimulator';

interface ToolboxProps {
    accountId: string;
    wsConnection: WebSocket | null;
}

type TabType = 'Trade' | 'History' | 'Orders' | 'Alerts' | 'Sessions' | 'Correlation' | 'Tester' | 'Signals' | 'RiskMap' | 'Sentiment' | 'Screener' | 'MTA' | 'PnL' | 'Drawing' | 'Analytics' | 'AdvancedOrders' | 'Replay' | 'CopyTrading' | 'NewsImpact' | 'PositionSizer' | 'EquityCurve' | 'Volatility' | 'Currency' | 'Journal' | 'JournalV2' | 'SessionMap' | 'RiskExposure' | 'OrderFlow' | 'MarketProfile' | 'TradingSim';

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

interface HistoryTrade {
    id: number;
    symbol: string;
    type: string;
    side: string;
    volume: number;
    openPrice: number;
    closePrice?: number;
    openTime: string;
    closeTime?: string;
    profit: number;
    commission: number;
    swap: number;
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

interface Alert {
    id: string;
    symbol: string;
    condition: string;
    price: number;
    status: 'ACTIVE' | 'TRIGGERED' | 'EXPIRED';
    createdAt: string;
}

interface JournalEntry {
    timestamp: string;
    message: string;
    type: 'INFO' | 'WARNING' | 'ERROR' | 'SUCCESS';
}

export function Toolbox({ accountId, wsConnection }: ToolboxProps) {
    const [isCollapsed, setIsCollapsed] = useState(false);
    const [height, setHeight] = useState(() => {
        const saved = localStorage.getItem('toolbox-height');
        return saved ? parseInt(saved, 10) : 200;
    });
    const [activeTab, setActiveTab] = useState<TabType>('Trade');
    const [isResizing, setIsResizing] = useState(false);

    // Data states
    const [positions, setPositions] = useState<Position[]>([]);
    const [history, setHistory] = useState<HistoryTrade[]>([]);
    const [orders, setOrders] = useState<PendingOrder[]>([]);
    const [journal, setJournal] = useState<JournalEntry[]>([]);

    // Get account from store
    const account = useAppStore((state) => state.account);

    // Get alerts from alert store
    const alerts = useAlertStore((state) => state.alerts);
    const removeAlert = useAlertStore((state) => state.removeAlert);
    const updateAlert = useAlertStore((state) => state.updateAlert);

    // Fetch positions
    useEffect(() => {
        const fetchPositions = async () => {
            try {
                const res = await fetch(`${API_ENDPOINTS.positions}?accountId=${accountId}`);
                if (res.ok) {
                    const data = await res.json();
                    setPositions(data || []);
                }
            } catch (err) {
                console.error('[Toolbox] Failed to fetch positions:', err);
            }
        };

        fetchPositions();
        const interval = setInterval(fetchPositions, 2000); // Poll every 2s
        return () => clearInterval(interval);
    }, [accountId]);

    // Fetch history/trades
    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const res = await fetch(`${API_ENDPOINTS.account}/${accountId}/trades`);
                if (res.ok) {
                    const data = await res.json();
                    setHistory(data || []);
                }
            } catch (err) {
                console.error('[Toolbox] Failed to fetch history:', err);
            }
        };

        fetchHistory();
    }, [accountId]);

    // Fetch pending orders
    useEffect(() => {
        const fetchOrders = async () => {
            try {
                const res = await fetch(`${API_ENDPOINTS.orders}?accountId=${accountId}&status=PENDING`);
                if (res.ok) {
                    const data = await res.json();
                    setOrders(data || []);
                }
            } catch (err) {
                console.error('[Toolbox] Failed to fetch orders:', err);
            }
        };

        fetchOrders();
        const interval = setInterval(fetchOrders, 3000); // Poll every 3s
        return () => clearInterval(interval);
    }, [accountId]);

    // Fetch alerts - alerts are managed by useAlertStore
    // No need to fetch from API as they're persisted in localStorage

    // WebSocket for real-time position updates
    useEffect(() => {
        if (!wsConnection) return;

        const handleMessage = (event: MessageEvent) => {
            try {
                const data = JSON.parse(event.data);
                if (data.type === 'position_update') {
                    // Update positions in real-time
                    setPositions((prev) => {
                        const index = prev.findIndex((p) => p.id === data.position.id);
                        if (index >= 0) {
                            const updated = [...prev];
                            updated[index] = { ...updated[index], ...data.position };
                            return updated;
                        }
                        return prev;
                    });

                    // Add journal entry
                    addJournalEntry(`Position ${data.position.id} updated`, 'INFO');
                } else if (data.type === 'order_filled') {
                    addJournalEntry(`Order filled: ${data.order.symbol} ${data.order.side} ${data.order.volume}`, 'SUCCESS');
                } else if (data.type === 'alert_triggered') {
                    addJournalEntry(`Alert triggered: ${data.alert.symbol} at ${data.alert.price}`, 'WARNING');
                }
            } catch (err) {
                console.error('[Toolbox] WS parse error:', err);
            }
        };

        wsConnection.addEventListener('message', handleMessage);
        return () => wsConnection.removeEventListener('message', handleMessage);
    }, [wsConnection]);

    // Add journal entry
    const addJournalEntry = (message: string, type: JournalEntry['type']) => {
        setJournal((prev) => [
            { timestamp: new Date().toISOString(), message, type },
            ...prev.slice(0, 99) // Keep last 100 entries
        ]);
    };

    // Resize handler
    useEffect(() => {
        const handleMouseMove = (e: MouseEvent) => {
            if (isResizing) {
                const newHeight = window.innerHeight - e.clientY;
                const clampedHeight = Math.max(100, Math.min(newHeight, window.innerHeight * 0.5));
                setHeight(clampedHeight);
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
    }, [isResizing]);

    // Persist height
    useEffect(() => {
        localStorage.setItem('toolbox-height', height.toString());
    }, [height]);

    // Close position
    const handleClosePosition = async (id: number) => {
        try {
            const res = await fetch(API_ENDPOINTS.closePosition, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ positionId: id })
            });

            if (!res.ok) throw new Error('Failed to close position');

            // Refresh positions
            const posRes = await fetch(`${API_ENDPOINTS.positions}?accountId=${accountId}`);
            if (posRes.ok) setPositions(await posRes.json() || []);

            addJournalEntry(`Closed position #${id}`, 'SUCCESS');
        } catch (err) {
            console.error('[Toolbox] Close position failed:', err);
            addJournalEntry(`Failed to close position #${id}`, 'ERROR');
        }
    };

    // Cancel order
    const handleCancelOrder = async (orderId: string) => {
        try {
            const res = await fetch(`${API_ENDPOINTS.orders}/${orderId}`, {
                method: 'DELETE'
            });

            if (!res.ok) throw new Error('Failed to cancel order');

            // Refresh orders
            const ordRes = await fetch(`${API_ENDPOINTS.orders}?accountId=${accountId}&status=PENDING`);
            if (ordRes.ok) setOrders(await ordRes.json() || []);

            addJournalEntry(`Cancelled order #${orderId}`, 'SUCCESS');
        } catch (err) {
            console.error('[Toolbox] Cancel order failed:', err);
            addJournalEntry(`Failed to cancel order #${orderId}`, 'ERROR');
        }
    };

    // Tab counts
    const tradeBadge = positions.length > 0 ? ` (${positions.length})` : '';
    const ordersBadge = orders.length > 0 ? ` (${orders.length})` : '';
    const alertsBadge = alerts.filter(a => a.status === 'active').length > 0
        ? ` (${alerts.filter(a => a.status === 'active').length})`
        : '';

    if (isCollapsed) {
        return (
            <div className="flex-shrink-0 bg-[#1e1e1e] border-t border-zinc-700 h-8 flex items-center px-2 justify-between select-none">
                <div className="text-xs text-zinc-400 font-medium">Toolbox</div>
                <button
                    onClick={() => setIsCollapsed(false)}
                    className="text-zinc-500 hover:text-zinc-300 transition-colors"
                >
                    <ChevronUp size={14} />
                </button>
            </div>
        );
    }

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

            {/* Tab Bar */}
            <div className="flex items-center bg-[#1e1e1e] border-b border-zinc-700 select-none overflow-x-auto scrollbar-thin scrollbar-thumb-zinc-700">
                <div className="flex items-center flex-1">
                    {(['Trade', 'History', 'Orders', 'Alerts', 'Sessions', 'Correlation', 'Tester', 'Signals', 'RiskMap', 'Sentiment', 'Screener', 'MTA', 'PnL', 'Drawing', 'Analytics', 'AdvancedOrders', 'Replay', 'CopyTrading', 'NewsImpact', 'PositionSizer', 'EquityCurve', 'Volatility', 'Currency', 'Journal', 'JournalV2', 'SessionMap', 'RiskExposure', 'OrderFlow', 'MarketProfile', 'TradingSim'] as TabType[]).map((tab) => {
                        const badge = tab === 'Trade' ? tradeBadge : tab === 'Orders' ? ordersBadge : tab === 'Alerts' ? alertsBadge : '';
                        return (
                            <button
                                key={tab}
                                onClick={() => setActiveTab(tab)}
                                className={`px-3 py-1 text-[11px] font-medium transition-colors whitespace-nowrap border-r border-zinc-700/50 relative
                                ${activeTab === tab
                                        ? 'text-zinc-100 bg-[#2d3436] shadow-[inset_0_-2px_0_0_#3b82f6]'
                                        : 'text-zinc-500 hover:text-zinc-300 hover:bg-[#2d3436]/50'}`}
                            >
                                {tab}{badge}
                            </button>
                        );
                    })}
                </div>
                <button
                    onClick={() => setIsCollapsed(true)}
                    className="px-2 text-zinc-500 hover:text-zinc-300 transition-colors"
                    title="Collapse"
                >
                    <ChevronDown size={14} />
                </button>
            </div>

            {/* Content Area */}
            <div className="flex-1 overflow-hidden relative bg-[#1e1e1e] tabular-nums">
                {activeTab === 'Trade' && (
                    <PositionsPanel
                        accountId={accountId}
                        wsConnection={wsConnection}
                    />
                )}
                {activeTab === 'History' && (
                    <HistoryTab history={history} />
                )}
                {activeTab === 'Orders' && (
                    <OrdersTab orders={orders} onCancelOrder={handleCancelOrder} />
                )}
                {activeTab === 'Alerts' && (
                    <AlertManager />
                )}
                {activeTab === 'Sessions' && (
                    <MarketSessions />
                )}
                {activeTab === 'Correlation' && (
                    <CorrelationMatrix />
                )}
                {activeTab === 'Tester' && (
                    <StrategyTester />
                )}
                {activeTab === 'Signals' && (
                    <TradingSignals />
                )}
                {activeTab === 'RiskMap' && (
                    <RiskHeatmap />
                )}
                {activeTab === 'Sentiment' && (
                    <SentimentAnalysis />
                )}
                {activeTab === 'Screener' && (
                    <SymbolScreener />
                )}
                {activeTab === 'MTA' && (
                    <MultiTimeframeAnalysis />
                )}
                {activeTab === 'PnL' && (
                    <PnLCalendar />
                )}
                {activeTab === 'Drawing' && (
                    <DrawingTools />
                )}
                {activeTab === 'Analytics' && (
                    <TradeAnalytics />
                )}
                {activeTab === 'AdvancedOrders' && (
                    <AdvancedOrders />
                )}
                {activeTab === 'Replay' && (
                    <MarketReplay />
                )}
                {activeTab === 'CopyTrading' && (
                    <CopyTradingLeaderboard />
                )}
                {activeTab === 'NewsImpact' && (
                    <NewsImpactAnalyzer />
                )}
                {activeTab === 'PositionSizer' && (
                    <PositionSizeCalculator />
                )}
                {activeTab === 'EquityCurve' && (
                    <EquityCurveTracker />
                )}
                {activeTab === 'Volatility' && (
                    <VolatilityAnalyzer />
                )}
                {activeTab === 'Currency' && (
                    <CurrencyConverter />
                )}
                {activeTab === 'Journal' && (
                    <JournalTab entries={journal} />
                )}
                {activeTab === 'JournalV2' && (
                    <TradeJournalV2 />
                )}
                {activeTab === 'SessionMap' && (
                    <SessionHeatMap />
                )}
                {activeTab === 'RiskExposure' && (
                    <RiskExposurePanel />
                )}
                {activeTab === 'OrderFlow' && (
                    <OrderFlowVisualizer />
                )}
                {activeTab === 'MarketProfile' && (
                    <MarketProfile />
                )}
                {activeTab === 'TradingSim' && (
                    <TradingSimulator />
                )}
            </div>
        </div>
    );
}

// --- TAB COMPONENTS ---

function TradeTab({ positions, account, onClosePosition }: {
    positions: Position[];
    account: any;
    onClosePosition: (id: number) => void;
}) {
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

    return (
        <div className="h-full flex flex-col">
            <div className="flex-1 overflow-auto scrollbar-thin scrollbar-thumb-zinc-700">
                <table className="w-full text-left border-collapse">
                    <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                        <tr>
                            <th className="p-1 pl-2 border-r border-zinc-700 w-16">Ticket</th>
                            <th className="p-1 border-r border-zinc-700 w-20">Symbol</th>
                            <th className="p-1 border-r border-zinc-700 w-12">Type</th>
                            <th className="p-1 border-r border-zinc-700 w-16 text-right">Volume</th>
                            <th className="p-1 border-r border-zinc-700 w-20 text-right">Open Price</th>
                            <th className="p-1 border-r border-zinc-700 w-16 text-right">S/L</th>
                            <th className="p-1 border-r border-zinc-700 w-16 text-right">T/P</th>
                            <th className="p-1 border-r border-zinc-700 w-20 text-right">Current Price</th>
                            <th className="p-1 border-r border-zinc-700 w-16 text-right">Profit</th>
                            <th className="p-1 border-r border-zinc-700 w-16 text-right">Swap</th>
                            <th className="p-1 border-r border-zinc-700 w-20 text-right">Commission</th>
                            <th className="w-6"></th>
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
                                    <td className="p-1 border-r border-zinc-800 text-right">
                                        <span className={pos.sl > 0 ? "text-red-300 bg-red-900/20 px-1 rounded-sm" : "text-zinc-600"}>
                                            {pos.sl > 0 ? pos.sl.toFixed(5) : ''}
                                        </span>
                                    </td>
                                    <td className="p-1 border-r border-zinc-800 text-right">
                                        <span className={pos.tp > 0 ? "text-emerald-300 bg-emerald-900/20 px-1 rounded-sm" : "text-zinc-600"}>
                                            {pos.tp > 0 ? pos.tp.toFixed(5) : ''}
                                        </span>
                                    </td>
                                    <td className="p-1 border-r border-zinc-800 text-right text-zinc-300">{pos.currentPrice.toFixed(5)}</td>
                                    <td className={`p-1 border-r border-zinc-800 text-right font-bold ${pos.unrealizedPnL >= 0 ? 'text-[#4ade80]' : 'text-[#f87171]'}`}>
                                        {formatMoney(pos.unrealizedPnL)}
                                    </td>
                                    <td className="p-1 border-r border-zinc-800 text-right text-zinc-500 text-[10px]">{formatMoney(pos.swap)}</td>
                                    <td className="p-1 border-r border-zinc-800 text-right text-zinc-500 text-[10px]">{formatMoney(pos.commission)}</td>
                                    <td className="p-1 text-center">
                                        <button
                                            onClick={() => onClosePosition(pos.id)}
                                            className="text-zinc-600 hover:text-white hover:bg-zinc-700 w-4 h-4 rounded flex items-center justify-center transition-colors"
                                            title="Close Position"
                                        >
                                            ×
                                        </button>
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {/* Footer Summary */}
            {account && (
                <div className="bg-[#2d3436] border-t border-zinc-600 w-full text-[11px] font-medium text-zinc-300 flex items-center px-4 py-1 gap-6 tabular-nums">
                    <span>Balance: <span className="text-white font-bold">{formatMoney(account.balance)}</span></span>
                    <span>Equity: <span className="text-white font-bold">{formatMoney(account.equity)}</span></span>
                    <span>Margin: <span className="text-white">{formatMoney(account.margin)}</span></span>
                    <span>Free Margin: <span className="text-white">{formatMoney(account.freeMargin)}</span></span>
                    <span>Margin Level: <span className="text-white">{account.marginLevel?.toFixed(2)}%</span></span>
                    <div className="flex-1"></div>
                    <span>Total Profit: <span className={account.unrealizedPL >= 0 ? "text-[#4ade80] font-bold" : "text-[#f87171] font-bold"}>
                        {formatMoney(account.unrealizedPL)}
                    </span></span>
                </div>
            )}
        </div>
    );
}

function HistoryTab({ history }: { history: HistoryTrade[] }) {
    const formatMoney = (val: number) => {
        return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val);
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

    return (
        <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 bg-[#1e1e1e]">
            <table className="w-full text-left border-collapse">
                <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                    <tr>
                        <th className="p-1 pl-2 border-r border-zinc-700 w-16">Ticket</th>
                        <th className="p-1 border-r border-zinc-700 w-28">Open Time</th>
                        <th className="p-1 border-r border-zinc-700 w-12">Type</th>
                        <th className="p-1 border-r border-zinc-700 w-16 text-right">Volume</th>
                        <th className="p-1 border-r border-zinc-700 w-20">Symbol</th>
                        <th className="p-1 border-r border-zinc-700 w-20 text-right">Open Price</th>
                        <th className="p-1 border-r border-zinc-700 w-28">Close Time</th>
                        <th className="p-1 border-r border-zinc-700 w-20 text-right">Close Price</th>
                        <th className="p-1 border-r border-zinc-700 w-16 text-right">Profit</th>
                        <th className="p-1 text-right pr-4">Date</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800 text-zinc-300 font-mono text-[11px]">
                    {history.length === 0 ? (
                        <tr>
                            <td colSpan={10} className="p-8 text-center text-zinc-600 text-xs italic">
                                No trade history
                            </td>
                        </tr>
                    ) : (
                        [...history].reverse().map((trade, i) => (
                            <tr key={trade.id} className={`${i % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'} hover:bg-[#2d3436] transition-colors`}>
                                <td className="p-1 pl-2 border-r border-zinc-800 text-zinc-500">{trade.id}</td>
                                <td className="p-1 border-r border-zinc-800 text-zinc-400 whitespace-nowrap text-[10px]">{formatDate(trade.openTime)}</td>
                                <td className={`p-1 border-r border-zinc-800 font-bold text-[10px] ${trade.side === 'BUY' ? 'text-blue-400' : 'text-red-400'}`}>
                                    {trade.side}
                                </td>
                                <td className="p-1 border-r border-zinc-800 text-right font-medium text-zinc-200">{trade.volume.toFixed(2)}</td>
                                <td className="p-1 border-r border-zinc-800 font-bold text-zinc-100">{trade.symbol}</td>
                                <td className="p-1 border-r border-zinc-800 text-right text-zinc-400">{trade.openPrice.toFixed(5)}</td>
                                <td className="p-1 border-r border-zinc-800 text-zinc-400 whitespace-nowrap text-[10px]">{trade.closeTime ? formatDate(trade.closeTime) : '-'}</td>
                                <td className="p-1 border-r border-zinc-800 text-right text-zinc-400">{trade.closePrice?.toFixed(5) || '-'}</td>
                                <td className={`p-1 border-r border-zinc-800 text-right font-bold ${trade.profit >= 0 ? 'text-[#4ade80]' : 'text-[#f87171]'}`}>
                                    {formatMoney(trade.profit)}
                                </td>
                                <td className="p-1 text-right pr-4 text-zinc-500 text-[10px]">{formatDate(trade.openTime).split(' ')[0]}</td>
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
    );
}

function OrdersTab({ orders, onCancelOrder }: {
    orders: PendingOrder[];
    onCancelOrder: (orderId: string) => void;
}) {
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

    return (
        <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 bg-[#1e1e1e]">
            <table className="w-full text-left border-collapse">
                <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                    <tr>
                        <th className="p-1 pl-2 border-r border-zinc-700 w-16">Ticket</th>
                        <th className="p-1 border-r border-zinc-700 w-20">Symbol</th>
                        <th className="p-1 border-r border-zinc-700 w-16">Type</th>
                        <th className="p-1 border-r border-zinc-700 w-12">Side</th>
                        <th className="p-1 border-r border-zinc-700 w-16 text-right">Volume</th>
                        <th className="p-1 border-r border-zinc-700 w-20 text-right">Price</th>
                        <th className="p-1 border-r border-zinc-700 w-16 text-right">S/L</th>
                        <th className="p-1 border-r border-zinc-700 w-16 text-right">T/P</th>
                        <th className="p-1 border-r border-zinc-700 w-16">Status</th>
                        <th className="p-1 pr-4">Created</th>
                        <th className="w-6"></th>
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
                                    <span className={order.sl && order.sl > 0 ? "text-red-300 bg-red-900/20 px-1 rounded-sm" : "text-zinc-600"}>
                                        {order.sl && order.sl > 0 ? order.sl.toFixed(5) : ''}
                                    </span>
                                </td>
                                <td className="p-1 border-r border-zinc-800 text-right">
                                    <span className={order.tp && order.tp > 0 ? "text-emerald-300 bg-emerald-900/20 px-1 rounded-sm" : "text-zinc-600"}>
                                        {order.tp && order.tp > 0 ? order.tp.toFixed(5) : ''}
                                    </span>
                                </td>
                                <td className="p-1 border-r border-zinc-800 text-zinc-400 text-[10px]">
                                    <span className="bg-yellow-900/20 text-yellow-300 px-1 rounded-sm">{order.status}</span>
                                </td>
                                <td className="p-1 border-r border-zinc-800 text-zinc-500 text-[10px] pr-4">{formatDate(order.createdAt)}</td>
                                <td className="p-1 text-center">
                                    <button
                                        onClick={() => onCancelOrder(order.id)}
                                        className="text-zinc-600 hover:text-white hover:bg-zinc-700 w-4 h-4 rounded flex items-center justify-center transition-colors"
                                        title="Cancel Order"
                                    >
                                        ×
                                    </button>
                                </td>
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
    );
}

function AlertsTab({ alerts }: { alerts: Alert[] }) {
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

    return (
        <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 bg-[#1e1e1e]">
            <table className="w-full text-left border-collapse">
                <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                    <tr>
                        <th className="p-1 pl-2 border-r border-zinc-700 w-20">Symbol</th>
                        <th className="p-1 border-r border-zinc-700">Condition</th>
                        <th className="p-1 border-r border-zinc-700 w-20 text-right">Price</th>
                        <th className="p-1 border-r border-zinc-700 w-16">Status</th>
                        <th className="p-1 pr-4">Created</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800 text-zinc-300 font-mono text-[11px]">
                    {alerts.length === 0 ? (
                        <tr>
                            <td colSpan={5} className="p-8 text-center text-zinc-600 text-xs italic">
                                No active alerts
                            </td>
                        </tr>
                    ) : (
                        alerts.map((alert, i) => (
                            <tr key={alert.id} className={`${i % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'} hover:bg-[#2d3436] transition-colors`}>
                                <td className="p-1 pl-2 border-r border-zinc-800 font-bold text-zinc-100">{alert.symbol}</td>
                                <td className="p-1 border-r border-zinc-800 text-zinc-400">{alert.condition}</td>
                                <td className="p-1 border-r border-zinc-800 text-right text-zinc-300">{alert.price.toFixed(5)}</td>
                                <td className="p-1 border-r border-zinc-800">
                                    <span className={`text-[10px] px-1 rounded-sm ${alert.status === 'ACTIVE' ? 'bg-green-900/20 text-green-300' :
                                            alert.status === 'TRIGGERED' ? 'bg-yellow-900/20 text-yellow-300' :
                                                'bg-zinc-800 text-zinc-500'
                                        }`}>
                                        {alert.status}
                                    </span>
                                </td>
                                <td className="p-1 text-zinc-500 text-[10px] pr-4">{formatDate(alert.createdAt)}</td>
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
    );
}

function JournalTab({ entries }: { entries: JournalEntry[] }) {
    const formatTime = (timestamp: string) => {
        try {
            return new Date(timestamp).toLocaleTimeString('en-US', {
                hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false
            });
        } catch {
            return timestamp;
        }
    };

    const getTypeColor = (type: JournalEntry['type']) => {
        switch (type) {
            case 'SUCCESS': return 'text-emerald-400';
            case 'WARNING': return 'text-yellow-400';
            case 'ERROR': return 'text-red-400';
            default: return 'text-zinc-400';
        }
    };

    const getTypeIcon = (type: JournalEntry['type']) => {
        switch (type) {
            case 'SUCCESS': return <TrendingUp size={12} />;
            case 'WARNING': return <AlertTriangle size={12} />;
            case 'ERROR': return <X size={12} />;
            default: return <FileText size={12} />;
        }
    };

    return (
        <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 bg-[#1e1e1e] font-mono">
            <div className="p-2 space-y-1">
                {entries.length === 0 ? (
                    <div className="p-8 text-center text-zinc-600 text-xs italic">
                        No journal entries
                    </div>
                ) : (
                    entries.map((entry, i) => (
                        <div
                            key={i}
                            className="flex items-start gap-2 px-2 py-1 hover:bg-zinc-800/50 rounded text-[11px]"
                        >
                            <Clock size={10} className="text-zinc-600 mt-0.5 flex-shrink-0" />
                            <span className="text-zinc-500 tabular-nums min-w-[65px]">{formatTime(entry.timestamp)}</span>
                            <span className={`flex-shrink-0 ${getTypeColor(entry.type)}`}>
                                {getTypeIcon(entry.type)}
                            </span>
                            <span className="text-zinc-300 flex-1">{entry.message}</span>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}
