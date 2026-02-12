'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
    Users,
    Activity,
    BarChart3,
    TrendingUp,
    ArrowDownCircle,
    ArrowUpCircle,
    Clock,
    AlertTriangle,
    Bell,
    LayoutDashboard,
    Loader2,
    Wifi,
    WifiOff,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';
import { api } from '@/services/apiClient';
import { getAdminWebSocket } from '@/services/adminWebSocket';
import SparklineChart from './SparklineChart';

interface OverviewData {
    totalAccounts: number;
    activePositions: number;
    todayVolume: number;
    todayPnL: number;
    todayDeposits: number;
    todayWithdrawals: number;
}

interface Trade {
    time: string;
    accountId: string;
    symbol: string;
    type: 'BUY' | 'SELL';
    volume: number;
    price: number;
    pnl: number;
}

interface Alert {
    id: string;
    severity: 'info' | 'warning' | 'error' | 'success';
    message: string;
    timestamp: string;
}

interface SymbolVolume {
    symbol: string;
    volume: number;
}

interface ActivityEvent {
    timestamp: string;
    accountId: string;
    description: string;
    type: 'login' | 'deposit' | 'withdrawal' | 'trade';
}

interface KPIHistory {
    totalAccounts: number[];
    activePositions: number[];
    todayVolume: number[];
    todayPnL: number[];
    todayDeposits: number[];
    todayWithdrawals: number[];
}

interface PulseState {
    totalAccounts: boolean;
    activePositions: boolean;
    todayVolume: boolean;
    todayPnL: boolean;
    todayDeposits: boolean;
    todayWithdrawals: boolean;
}

const SEVERITY_COLORS = {
    info: 'text-blue-400 bg-blue-500/10 border-blue-500/30',
    warning: 'text-yellow-400 bg-yellow-500/10 border-yellow-500/30',
    error: 'text-red-400 bg-red-500/10 border-red-500/30',
    success: 'text-green-400 bg-green-500/10 border-green-500/30',
};

const MAX_HISTORY_LENGTH = 20;

export default function DashboardOverview() {
    const [overview, setOverview] = useState<OverviewData | null>(null);
    const [kpiHistory, setKpiHistory] = useState<KPIHistory>({
        totalAccounts: [],
        activePositions: [],
        todayVolume: [],
        todayPnL: [],
        todayDeposits: [],
        todayWithdrawals: [],
    });
    const [pulseState, setPulseState] = useState<PulseState>({
        totalAccounts: false,
        activePositions: false,
        todayVolume: false,
        todayPnL: false,
        todayDeposits: false,
        todayWithdrawals: false,
    });
    const [recentTrades, setRecentTrades] = useState<Trade[]>([]);
    const [alerts, setAlerts] = useState<Alert[]>([]);
    const [topSymbols, setTopSymbols] = useState<SymbolVolume[]>([]);
    const [activityFeed, setActivityFeed] = useState<ActivityEvent[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [wsConnected, setWsConnected] = useState(false);

    const wsRef = useRef<ReturnType<typeof getAdminWebSocket> | null>(null);
    const pulseTimeoutsRef = useRef<{ [key in keyof PulseState]?: NodeJS.Timeout }>({});

    /**
     * Trigger pulse animation for a specific KPI
     */
    const triggerPulse = useCallback((kpi: keyof PulseState) => {
        // Clear existing timeout for this KPI
        if (pulseTimeoutsRef.current[kpi]) {
            clearTimeout(pulseTimeoutsRef.current[kpi]);
        }

        // Set pulse to true
        setPulseState(prev => ({ ...prev, [kpi]: true }));

        // Auto-clear after 500ms
        pulseTimeoutsRef.current[kpi] = setTimeout(() => {
            setPulseState(prev => ({ ...prev, [kpi]: false }));
        }, 500);
    }, []);

    /**
     * Update KPI history with new value
     */
    const updateHistory = useCallback((kpi: keyof KPIHistory, value: number) => {
        setKpiHistory(prev => {
            const newHistory = [...prev[kpi], value];
            // Keep only last 20 values
            if (newHistory.length > MAX_HISTORY_LENGTH) {
                newHistory.shift();
            }
            return { ...prev, [kpi]: newHistory };
        });
    }, []);

    /**
     * Update overview data and trigger pulse animations
     */
    const updateOverviewData = useCallback((newData: OverviewData) => {
        setOverview(prevData => {
            // Trigger pulse for changed values
            if (prevData) {
                if (prevData.totalAccounts !== newData.totalAccounts) {
                    triggerPulse('totalAccounts');
                    updateHistory('totalAccounts', newData.totalAccounts);
                }
                if (prevData.activePositions !== newData.activePositions) {
                    triggerPulse('activePositions');
                    updateHistory('activePositions', newData.activePositions);
                }
                if (prevData.todayVolume !== newData.todayVolume) {
                    triggerPulse('todayVolume');
                    updateHistory('todayVolume', newData.todayVolume);
                }
                if (prevData.todayPnL !== newData.todayPnL) {
                    triggerPulse('todayPnL');
                    updateHistory('todayPnL', newData.todayPnL);
                }
                if (prevData.todayDeposits !== newData.todayDeposits) {
                    triggerPulse('todayDeposits');
                    updateHistory('todayDeposits', newData.todayDeposits);
                }
                if (prevData.todayWithdrawals !== newData.todayWithdrawals) {
                    triggerPulse('todayWithdrawals');
                    updateHistory('todayWithdrawals', newData.todayWithdrawals);
                }
            } else {
                // First load - initialize history without pulse
                updateHistory('totalAccounts', newData.totalAccounts);
                updateHistory('activePositions', newData.activePositions);
                updateHistory('todayVolume', newData.todayVolume);
                updateHistory('todayPnL', newData.todayPnL);
                updateHistory('todayDeposits', newData.todayDeposits);
                updateHistory('todayWithdrawals', newData.todayWithdrawals);
            }

            return newData;
        });
    }, [triggerPulse, updateHistory]);

    /**
     * Fetch all dashboard data (HTTP fallback)
     */
    const fetchDashboardData = useCallback(async () => {
        try {
            setError(null);

            // Fetch KPI overview data
            const overviewData = await api.get<OverviewData>(API_CONFIG.ANALYTICS_OVERVIEW);
            updateOverviewData(overviewData);

            // Fetch recent trades (last 10)
            const tradesUrl = `${API_CONFIG.EXPORT_TRADES}?format=json&limit=10`;
            const trades = await api.get<Trade[]>(tradesUrl);
            setRecentTrades(trades);

            // Fetch alerts/notifications
            const alertsUrl = `${API_CONFIG.NOTIFICATIONS}?limit=5`;
            const notifs = await api.get<Alert[]>(alertsUrl);
            setAlerts(notifs);

            // Calculate top 5 traded symbols from recent trades
            const symbolMap = new Map<string, number>();
            trades.forEach((trade) => {
                const current = symbolMap.get(trade.symbol) || 0;
                symbolMap.set(trade.symbol, current + trade.volume);
            });
            const sortedSymbols = Array.from(symbolMap.entries())
                .map(([symbol, volume]) => ({ symbol, volume }))
                .sort((a, b) => b.volume - a.volume)
                .slice(0, 5);
            setTopSymbols(sortedSymbols);

            // Generate mock activity feed (replace with real API when available)
            const mockActivity: ActivityEvent[] = [
                { timestamp: new Date().toISOString(), accountId: '123456', description: 'User logged in', type: 'login' },
                { timestamp: new Date(Date.now() - 300000).toISOString(), accountId: '789012', description: 'Deposited $5,000', type: 'deposit' },
                { timestamp: new Date(Date.now() - 600000).toISOString(), accountId: '345678', description: 'Executed trade on EURUSD', type: 'trade' },
                { timestamp: new Date(Date.now() - 900000).toISOString(), accountId: '901234', description: 'Withdrew $2,500', type: 'withdrawal' },
                { timestamp: new Date(Date.now() - 1200000).toISOString(), accountId: '567890', description: 'User logged in', type: 'login' },
            ];
            setActivityFeed(mockActivity);
        } catch (err: any) {
            console.error('[Dashboard] Failed to fetch data:', err);
            setError(err.message || 'Failed to load dashboard data');
        } finally {
            setLoading(false);
        }
    }, [updateOverviewData]);

    /**
     * Setup WebSocket connection and subscriptions
     */
    useEffect(() => {
        // Initial HTTP fetch
        fetchDashboardData();

        // Initialize WebSocket
        try {
            const ws = getAdminWebSocket(API_CONFIG.ADMIN_WS_URL);
            wsRef.current = ws;

            // Listen to connection state changes
            const unsubscribeState = ws.onStateChange((state) => {
                setWsConnected(state === 'connected');
            });

            // Subscribe to admin_stats channel for real-time KPI updates
            const unsubscribeStats = ws.on('admin_stats' as any, (event) => {
                const statsData = event.data as OverviewData;
                updateOverviewData(statsData);
            });

            // Connect WebSocket
            ws.connect();

            // Cleanup
            return () => {
                unsubscribeState();
                unsubscribeStats();
                // Clear all pulse timeouts
                Object.values(pulseTimeoutsRef.current).forEach(timeout => {
                    if (timeout) clearTimeout(timeout);
                });
            };
        } catch (err) {
            console.error('[Dashboard] WebSocket initialization failed:', err);
            // Fall back to polling if WebSocket fails
            const interval = setInterval(fetchDashboardData, 10000);
            return () => clearInterval(interval);
        }
    }, [fetchDashboardData, updateOverviewData]);

    /**
     * Format number with commas
     */
    const formatNumber = (num: number): string => {
        return num.toLocaleString('en-US', { maximumFractionDigits: 2 });
    };

    /**
     * Format currency
     */
    const formatCurrency = (num: number): string => {
        return `$${formatNumber(Math.abs(num))}`;
    };

    /**
     * Format time (HH:MM:SS)
     */
    const formatTime = (timestamp: string): string => {
        const date = new Date(timestamp);
        return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    };

    if (loading && !overview) {
        return (
            <div className="h-full flex items-center justify-center bg-[#0A0A0B]">
                <div className="text-center">
                    <Loader2 className="w-8 h-8 animate-spin text-blue-500 mx-auto mb-3" />
                    <p className="text-zinc-400">Loading dashboard...</p>
                </div>
            </div>
        );
    }

    if (error && !overview) {
        return (
            <div className="h-full flex items-center justify-center bg-[#0A0A0B]">
                <div className="text-center">
                    <AlertTriangle className="w-12 h-12 text-red-400 mx-auto mb-3" />
                    <p className="text-red-300">{error}</p>
                    <button
                        onClick={fetchDashboardData}
                        className="mt-4 px-4 py-2 bg-zinc-800 text-white rounded hover:bg-zinc-700"
                    >
                        Retry
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="h-full bg-[#0A0A0B] p-4 overflow-y-auto">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                    <LayoutDashboard className="w-6 h-6 text-blue-400" />
                    <h1 className="text-2xl font-bold text-white">Dashboard Overview</h1>
                    {/* Live Indicator */}
                    {wsConnected ? (
                        <div className="flex items-center gap-1.5 px-2 py-1 bg-green-500/10 border border-green-500/30 rounded text-xs font-medium text-green-400">
                            <Wifi className="w-3 h-3" />
                            <span>Live</span>
                        </div>
                    ) : (
                        <div className="flex items-center gap-1.5 px-2 py-1 bg-zinc-800 border border-zinc-700 rounded text-xs font-medium text-zinc-500">
                            <WifiOff className="w-3 h-3" />
                            <span>Polling</span>
                        </div>
                    )}
                </div>
                <div className="text-sm text-zinc-400">
                    {wsConnected ? 'Real-time updates via WebSocket' : 'Auto-refreshing every 10 seconds'}
                </div>
            </div>

            {/* KPI Cards - Top Row */}
            <div className="grid grid-cols-6 gap-4 mb-6">
                {/* Total Accounts */}
                <div className={`bg-zinc-900 rounded-lg p-4 border border-zinc-800 transition-all duration-300 ${pulseState.totalAccounts ? 'ring-2 ring-green-500 bg-green-500/5' : ''}`}>
                    <div className="flex items-center gap-2 mb-2">
                        <Users className="w-4 h-4 text-blue-400" />
                        <span className="text-xs text-zinc-400">Total Accounts</span>
                    </div>
                    <div className="flex items-end justify-between">
                        <div className="text-2xl font-bold text-white">
                            {overview ? formatNumber(overview.totalAccounts) : '-'}
                        </div>
                        {kpiHistory.totalAccounts.length > 1 && (
                            <SparklineChart data={kpiHistory.totalAccounts} width={60} height={24} />
                        )}
                    </div>
                </div>

                {/* Active Positions */}
                <div className={`bg-zinc-900 rounded-lg p-4 border border-zinc-800 transition-all duration-300 ${pulseState.activePositions ? 'ring-2 ring-green-500 bg-green-500/5' : ''}`}>
                    <div className="flex items-center gap-2 mb-2">
                        <Activity className="w-4 h-4 text-purple-400" />
                        <span className="text-xs text-zinc-400">Active Positions</span>
                    </div>
                    <div className="flex items-end justify-between">
                        <div className="text-2xl font-bold text-white">
                            {overview ? formatNumber(overview.activePositions) : '-'}
                        </div>
                        {kpiHistory.activePositions.length > 1 && (
                            <SparklineChart data={kpiHistory.activePositions} width={60} height={24} />
                        )}
                    </div>
                </div>

                {/* Today's Volume */}
                <div className={`bg-zinc-900 rounded-lg p-4 border border-zinc-800 transition-all duration-300 ${pulseState.todayVolume ? 'ring-2 ring-green-500 bg-green-500/5' : ''}`}>
                    <div className="flex items-center gap-2 mb-2">
                        <BarChart3 className="w-4 h-4 text-yellow-400" />
                        <span className="text-xs text-zinc-400">Today's Volume</span>
                    </div>
                    <div className="flex items-end justify-between">
                        <div className="text-2xl font-bold text-white">
                            {overview ? formatNumber(overview.todayVolume) : '-'}
                        </div>
                        {kpiHistory.todayVolume.length > 1 && (
                            <SparklineChart data={kpiHistory.todayVolume} width={60} height={24} />
                        )}
                    </div>
                </div>

                {/* Today's P&L */}
                <div className={`bg-zinc-900 rounded-lg p-4 border border-zinc-800 transition-all duration-300 ${pulseState.todayPnL ? 'ring-2 ring-green-500 bg-green-500/5' : ''}`}>
                    <div className="flex items-center gap-2 mb-2">
                        <TrendingUp className={`w-4 h-4 ${overview && overview.todayPnL >= 0 ? 'text-green-400' : 'text-red-400'}`} />
                        <span className="text-xs text-zinc-400">Today's P&L</span>
                    </div>
                    <div className="flex items-end justify-between">
                        <div className={`text-2xl font-bold ${overview && overview.todayPnL >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                            {overview ? (overview.todayPnL >= 0 ? '+' : '-') + formatCurrency(overview.todayPnL) : '-'}
                        </div>
                        {kpiHistory.todayPnL.length > 1 && (
                            <SparklineChart data={kpiHistory.todayPnL} width={60} height={24} />
                        )}
                    </div>
                </div>

                {/* Total Deposits Today */}
                <div className={`bg-zinc-900 rounded-lg p-4 border border-zinc-800 transition-all duration-300 ${pulseState.todayDeposits ? 'ring-2 ring-green-500 bg-green-500/5' : ''}`}>
                    <div className="flex items-center gap-2 mb-2">
                        <ArrowDownCircle className="w-4 h-4 text-green-400" />
                        <span className="text-xs text-zinc-400">Deposits Today</span>
                    </div>
                    <div className="flex items-end justify-between">
                        <div className="text-2xl font-bold text-white">
                            {overview ? formatCurrency(overview.todayDeposits) : '-'}
                        </div>
                        {kpiHistory.todayDeposits.length > 1 && (
                            <SparklineChart data={kpiHistory.todayDeposits} width={60} height={24} />
                        )}
                    </div>
                </div>

                {/* Total Withdrawals Today */}
                <div className={`bg-zinc-900 rounded-lg p-4 border border-zinc-800 transition-all duration-300 ${pulseState.todayWithdrawals ? 'ring-2 ring-green-500 bg-green-500/5' : ''}`}>
                    <div className="flex items-center gap-2 mb-2">
                        <ArrowUpCircle className="w-4 h-4 text-red-400" />
                        <span className="text-xs text-zinc-400">Withdrawals Today</span>
                    </div>
                    <div className="flex items-end justify-between">
                        <div className="text-2xl font-bold text-white">
                            {overview ? formatCurrency(overview.todayWithdrawals) : '-'}
                        </div>
                        {kpiHistory.todayWithdrawals.length > 1 && (
                            <SparklineChart data={kpiHistory.todayWithdrawals} width={60} height={24} />
                        )}
                    </div>
                </div>
            </div>

            {/* Middle Row - Recent Trades + Active Alerts */}
            <div className="grid grid-cols-2 gap-4 mb-6">
                {/* Recent Trades */}
                <div className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
                    <div className="p-4 border-b border-zinc-800 flex items-center justify-between">
                        <h3 className="font-semibold text-white flex items-center gap-2">
                            <Clock className="w-4 h-4 text-blue-400" />
                            Recent Trades
                        </h3>
                        <span className="text-xs text-zinc-500">Last 10 trades</span>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead className="bg-zinc-800">
                                <tr>
                                    <th className="px-3 py-2 text-left text-xs font-semibold text-zinc-400">Time</th>
                                    <th className="px-3 py-2 text-left text-xs font-semibold text-zinc-400">Account</th>
                                    <th className="px-3 py-2 text-left text-xs font-semibold text-zinc-400">Symbol</th>
                                    <th className="px-3 py-2 text-left text-xs font-semibold text-zinc-400">Type</th>
                                    <th className="px-3 py-2 text-right text-xs font-semibold text-zinc-400">Volume</th>
                                    <th className="px-3 py-2 text-right text-xs font-semibold text-zinc-400">Price</th>
                                    <th className="px-3 py-2 text-right text-xs font-semibold text-zinc-400">P&L</th>
                                </tr>
                            </thead>
                            <tbody>
                                {recentTrades.length === 0 ? (
                                    <tr>
                                        <td colSpan={7} className="px-3 py-8 text-center text-zinc-500">
                                            No recent trades
                                        </td>
                                    </tr>
                                ) : (
                                    recentTrades.map((trade, idx) => (
                                        <tr key={idx} className="border-b border-zinc-800 hover:bg-zinc-800/50">
                                            <td className="px-3 py-2 text-zinc-400">{formatTime(trade.time)}</td>
                                            <td className="px-3 py-2 text-zinc-300">{trade.accountId}</td>
                                            <td className="px-3 py-2 text-white font-medium">{trade.symbol}</td>
                                            <td className={`px-3 py-2 font-semibold ${trade.type === 'BUY' ? 'text-green-400' : 'text-red-400'}`}>
                                                {trade.type}
                                            </td>
                                            <td className="px-3 py-2 text-right text-zinc-300">{formatNumber(trade.volume)}</td>
                                            <td className="px-3 py-2 text-right text-zinc-300">{trade.price.toFixed(5)}</td>
                                            <td className={`px-3 py-2 text-right font-semibold ${trade.pnl >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                                                {trade.pnl >= 0 ? '+' : ''}{formatCurrency(trade.pnl)}
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>

                {/* Active Alerts */}
                <div className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
                    <div className="p-4 border-b border-zinc-800 flex items-center justify-between">
                        <h3 className="font-semibold text-white flex items-center gap-2">
                            <Bell className="w-4 h-4 text-yellow-400" />
                            Active Alerts
                        </h3>
                        <a href="#" className="text-xs text-blue-400 hover:underline">View All</a>
                    </div>
                    <div className="p-4 space-y-3">
                        {alerts.length === 0 ? (
                            <div className="text-center py-8 text-zinc-500">
                                No active alerts
                            </div>
                        ) : (
                            alerts.map((alert) => (
                                <div
                                    key={alert.id}
                                    className={`p-3 rounded border ${SEVERITY_COLORS[alert.severity]}`}
                                >
                                    <div className="flex items-start gap-2">
                                        <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                                        <div className="flex-1">
                                            <p className="text-sm font-medium">{alert.message}</p>
                                            <p className="text-xs opacity-70 mt-1">{formatTime(alert.timestamp)}</p>
                                        </div>
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            </div>

            {/* Bottom Row - Top Symbols + Activity Feed */}
            <div className="grid grid-cols-2 gap-4">
                {/* Top 5 Traded Symbols */}
                <div className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
                    <div className="p-4 border-b border-zinc-800">
                        <h3 className="font-semibold text-white flex items-center gap-2">
                            <BarChart3 className="w-4 h-4 text-emerald-400" />
                            Top 5 Traded Symbols
                        </h3>
                    </div>
                    <div className="p-4 space-y-3">
                        {topSymbols.length === 0 ? (
                            <div className="text-center py-8 text-zinc-500">
                                No trading data available
                            </div>
                        ) : (
                            topSymbols.map((item, idx) => {
                                const maxVolume = Math.max(...topSymbols.map(s => s.volume));
                                const percentage = (item.volume / maxVolume) * 100;
                                return (
                                    <div key={idx} className="space-y-1">
                                        <div className="flex items-center justify-between text-sm">
                                            <span className="text-white font-medium">{item.symbol}</span>
                                            <span className="text-zinc-400">{formatNumber(item.volume)}</span>
                                        </div>
                                        <div className="h-2 bg-zinc-800 rounded-full overflow-hidden">
                                            <div
                                                className="h-full bg-gradient-to-r from-emerald-500 to-blue-500 rounded-full transition-all"
                                                style={{ width: `${percentage}%` }}
                                            />
                                        </div>
                                    </div>
                                );
                            })
                        )}
                    </div>
                </div>

                {/* Account Activity Feed */}
                <div className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
                    <div className="p-4 border-b border-zinc-800">
                        <h3 className="font-semibold text-white flex items-center gap-2">
                            <Activity className="w-4 h-4 text-purple-400" />
                            Account Activity Feed
                        </h3>
                    </div>
                    <div className="p-4 space-y-3">
                        {activityFeed.length === 0 ? (
                            <div className="text-center py-8 text-zinc-500">
                                No recent activity
                            </div>
                        ) : (
                            activityFeed.map((event, idx) => (
                                <div key={idx} className="flex items-start gap-3 pb-3 border-b border-zinc-800 last:border-0">
                                    <div className="w-2 h-2 bg-blue-400 rounded-full mt-2 flex-shrink-0" />
                                    <div className="flex-1">
                                        <p className="text-sm text-zinc-300">
                                            <span className="text-white font-medium">#{event.accountId}</span> {event.description}
                                        </p>
                                        <p className="text-xs text-zinc-500 mt-1">{formatTime(event.timestamp)}</p>
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
