'use client';

import { useState, useEffect, useMemo } from 'react';
import {
    Users,
    TrendingUp,
    TrendingDown,
    AlertTriangle,
    CheckCircle,
    Clock,
    DollarSign,
    Activity,
    Mail,
    X,
    ChevronDown,
    ChevronUp
} from 'lucide-react';
import { API_CONFIG } from '../../config/api';

type LifecycleStage = 'Lead' | 'NewClient' | 'Active' | 'Dormant' | 'AtRisk' | 'Churned';

interface FunnelMetrics {
    registered: number;
    verified: number;
    deposited: number;
    traded: number;
    active: number;
    verification_rate: number;
    deposit_rate: number;
    trading_rate: number;
    activation_rate: number;
}

interface StageDistribution {
    stage: LifecycleStage;
    count: number;
    percentage: number;
    avg_revenue: number;
    avg_trades: number;
    avg_time_in_stage_days: number;
}

interface RetentionCohort {
    month: string;
    registered: number;
    retained_30d: number;
    retained_60d: number;
    retained_90d: number;
    retained_180d: number;
    retained_365d: number;
    retention_30d_pct: number;
    retention_60d_pct: number;
    retention_90d_pct: number;
    retention_180d_pct: number;
    retention_365d_pct: number;
}

interface AtRiskClient {
    client_id: string;
    name: string;
    email: string;
    last_login_at: string;
    last_trade_at: string;
    churn_risk: number;
    days_since_last_trade: number;
    days_since_last_login: number;
    lifetime_revenue: number;
}

interface LifecycleSegment {
    stage: LifecycleStage;
    client_count: number;
    avg_revenue: number;
    avg_trades: number;
    avg_deposits: number;
}

interface MonthlyTrend {
    month: string;
    new_registrations: number;
    new_activations: number;
    churned: number;
    churn_rate: number;
    active_clients: number;
    revenue: number;
}

interface TimelineEvent {
    date: string;
    event: string;
    description: string;
    value?: number;
}

export default function CustomerLifecycle() {
    const [loading, setLoading] = useState(true);
    const [funnel, setFunnel] = useState<FunnelMetrics | null>(null);
    const [stages, setStages] = useState<StageDistribution[]>([]);
    const [cohorts, setCohorts] = useState<RetentionCohort[]>([]);
    const [atRiskClients, setAtRiskClients] = useState<AtRiskClient[]>([]);
    const [segments, setSegments] = useState<LifecycleSegment[]>([]);
    const [trends, setTrends] = useState<MonthlyTrend[]>([]);
    const [selectedClient, setSelectedClient] = useState<string | null>(null);
    const [clientTimeline, setClientTimeline] = useState<TimelineEvent[]>([]);
    const [sortConfig, setSortConfig] = useState<{ key: keyof AtRiskClient; direction: 'asc' | 'desc' }>({
        key: 'churn_risk',
        direction: 'desc'
    });

    useEffect(() => {
        fetchAllData();
    }, []);

    const fetchAllData = async () => {
        setLoading(true);
        try {
            const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') || '' : '';
            const headers = {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            };

            const [funnelRes, stagesRes, retentionRes, atRiskRes, segmentsRes, trendsRes] = await Promise.all([
                fetch(API_CONFIG.LIFECYCLE_FUNNEL, { headers }),
                fetch(API_CONFIG.LIFECYCLE_OVERVIEW, { headers }),
                fetch(API_CONFIG.LIFECYCLE_RETENTION, { headers }),
                fetch(API_CONFIG.LIFECYCLE_AT_RISK, { headers }),
                fetch(API_CONFIG.LIFECYCLE_SEGMENTS, { headers }),
                fetch(API_CONFIG.LIFECYCLE_TRENDS, { headers })
            ]);

            if (funnelRes.ok) {
                const data = await funnelRes.json();
                setFunnel(data.funnel);
            }

            if (stagesRes.ok) {
                const data = await stagesRes.json();
                setStages(data.distributions || []);
            }

            if (retentionRes.ok) {
                const data = await retentionRes.json();
                setCohorts(data.cohorts || []);
            }

            if (atRiskRes.ok) {
                const data = await atRiskRes.json();
                setAtRiskClients(data.at_risk_clients || []);
            }

            if (segmentsRes.ok) {
                const data = await segmentsRes.json();
                setSegments(data.segments || []);
            }

            if (trendsRes.ok) {
                const data = await trendsRes.json();
                setTrends(data.trends || []);
            }
        } catch (error) {
            console.error('[CustomerLifecycle] Error fetching data:', error);
        } finally {
            setLoading(false);
        }
    };

    const fetchClientTimeline = async (clientId: string) => {
        try {
            const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') || '' : '';
            const res = await fetch(API_CONFIG.LIFECYCLE_CLIENT_TIMELINE(clientId), {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (res.ok) {
                const data = await res.json();
                setClientTimeline(data.events || []);
                setSelectedClient(clientId);
            }
        } catch (error) {
            console.error('[CustomerLifecycle] Error fetching client timeline:', error);
        }
    };

    const handleSort = (key: keyof AtRiskClient) => {
        setSortConfig(prev => ({
            key,
            direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
        }));
    };

    const sortedAtRiskClients = useMemo(() => {
        const sorted = [...atRiskClients].sort((a, b) => {
            const aVal = a[sortConfig.key];
            const bVal = b[sortConfig.key];

            if (typeof aVal === 'string' && typeof bVal === 'string') {
                return sortConfig.direction === 'asc'
                    ? aVal.localeCompare(bVal)
                    : bVal.localeCompare(aVal);
            }

            if (typeof aVal === 'number' && typeof bVal === 'number') {
                return sortConfig.direction === 'asc' ? aVal - bVal : bVal - aVal;
            }

            return 0;
        });
        return sorted;
    }, [atRiskClients, sortConfig]);

    const getRetentionColor = (percentage: number): string => {
        if (percentage >= 70) return 'bg-emerald-500/20 text-emerald-300';
        if (percentage >= 50) return 'bg-yellow-500/20 text-yellow-300';
        if (percentage >= 30) return 'bg-orange-500/20 text-orange-300';
        return 'bg-red-500/20 text-red-300';
    };

    const getStageIcon = (stage: LifecycleStage) => {
        switch (stage) {
            case 'Lead': return <Users size={14} className="text-blue-400" />;
            case 'NewClient': return <CheckCircle size={14} className="text-emerald-400" />;
            case 'Active': return <Activity size={14} className="text-green-400" />;
            case 'Dormant': return <Clock size={14} className="text-yellow-400" />;
            case 'AtRisk': return <AlertTriangle size={14} className="text-orange-400" />;
            case 'Churned': return <X size={14} className="text-red-400" />;
            default: return null;
        }
    };

    const getStageColor = (stage: LifecycleStage): string => {
        switch (stage) {
            case 'Lead': return 'border-blue-500/50 bg-blue-500/10';
            case 'NewClient': return 'border-emerald-500/50 bg-emerald-500/10';
            case 'Active': return 'border-green-500/50 bg-green-500/10';
            case 'Dormant': return 'border-yellow-500/50 bg-yellow-500/10';
            case 'AtRisk': return 'border-orange-500/50 bg-orange-500/10';
            case 'Churned': return 'border-red-500/50 bg-red-500/10';
            default: return 'border-zinc-600 bg-zinc-800';
        }
    };

    const formatMoney = (value: number) => {
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(value);
    };

    const formatDate = (dateStr: string) => {
        try {
            return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
        } catch {
            return dateStr;
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full bg-[#121316]">
                <div className="text-zinc-400 text-sm">Loading customer lifecycle data...</div>
            </div>
        );
    }

    return (
        <div className="h-full overflow-auto scrollbar-thin scrollbar-thumb-zinc-700 bg-[#121316] p-4">
            <div className="max-w-[1600px] mx-auto space-y-4">
                {/* Header */}
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-xl font-bold text-zinc-100">Customer Lifecycle & Retention</h1>
                        <p className="text-xs text-zinc-500 mt-1">Track customer journey and churn risk</p>
                    </div>
                    <button
                        onClick={fetchAllData}
                        className="px-3 py-1.5 bg-zinc-800 border border-zinc-700 rounded text-xs font-medium text-zinc-300 hover:bg-zinc-700 transition-colors"
                    >
                        Refresh Data
                    </button>
                </div>

                {/* Lifecycle Funnel */}
                {funnel && (
                    <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
                        <h2 className="text-sm font-bold text-zinc-200 mb-4 flex items-center gap-2">
                            <TrendingUp size={16} className="text-emerald-500" />
                            Lifecycle Conversion Funnel
                        </h2>
                        <div className="flex items-center justify-center gap-8">
                            {[
                                { label: 'Registered', value: funnel.registered, rate: 100 },
                                { label: 'Verified', value: funnel.verified, rate: funnel.verification_rate },
                                { label: 'Deposited', value: funnel.deposited, rate: funnel.deposit_rate },
                                { label: 'Traded', value: funnel.traded, rate: funnel.trading_rate },
                                { label: 'Active', value: funnel.active, rate: funnel.activation_rate }
                            ].map((stage, idx) => {
                                const width = 80 + (100 - idx * 15);
                                return (
                                    <div key={stage.label} className="flex flex-col items-center gap-2">
                                        <div
                                            className="bg-gradient-to-b from-emerald-500/30 to-emerald-500/10 border-2 border-emerald-500/50 rounded flex items-center justify-center relative"
                                            style={{ width: `${width}px`, height: '60px' }}
                                        >
                                            <div className="text-center">
                                                <div className="text-lg font-bold text-emerald-300">{stage.value}</div>
                                                <div className="text-[9px] text-zinc-500 uppercase">{stage.rate.toFixed(1)}%</div>
                                            </div>
                                        </div>
                                        <div className="text-[10px] font-medium text-zinc-400">{stage.label}</div>
                                        {idx < 4 && (
                                            <div className="absolute right-[-20px] top-1/2 transform -translate-y-1/2 text-zinc-600 text-xl">→</div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                )}

                {/* Stage Distribution Cards */}
                <div className="grid grid-cols-6 gap-3">
                    {stages.map(stage => (
                        <div
                            key={stage.stage}
                            className={`bg-[#1E2026] border rounded-lg p-3 ${getStageColor(stage.stage)}`}
                        >
                            <div className="flex items-center justify-between mb-2">
                                {getStageIcon(stage.stage)}
                                <div className="text-[10px] font-medium text-zinc-500">{stage.percentage.toFixed(1)}%</div>
                            </div>
                            <div className="text-xl font-bold text-zinc-100 mb-1">{stage.count}</div>
                            <div className="text-[10px] font-medium text-zinc-400 mb-2">{stage.stage}</div>
                            <div className="space-y-1 text-[9px] text-zinc-500">
                                <div>Avg Revenue: <span className="text-zinc-300">{formatMoney(stage.avg_revenue)}</span></div>
                                <div>Avg Trades: <span className="text-zinc-300">{stage.avg_trades.toFixed(0)}</span></div>
                            </div>
                        </div>
                    ))}
                </div>

                <div className="grid grid-cols-2 gap-4">
                    {/* Retention Cohort Table */}
                    <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
                        <h2 className="text-sm font-bold text-zinc-200 mb-3">12-Month Retention Cohort Analysis</h2>
                        <div className="overflow-x-auto">
                            <table className="w-full text-[10px]">
                                <thead className="bg-[#2d3436] text-zinc-400">
                                    <tr>
                                        <th className="p-1.5 text-left border-r border-zinc-700">Month</th>
                                        <th className="p-1.5 text-center border-r border-zinc-700">Registered</th>
                                        <th className="p-1.5 text-center border-r border-zinc-700">30d</th>
                                        <th className="p-1.5 text-center border-r border-zinc-700">60d</th>
                                        <th className="p-1.5 text-center border-r border-zinc-700">90d</th>
                                        <th className="p-1.5 text-center border-r border-zinc-700">180d</th>
                                        <th className="p-1.5 text-center">365d</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-zinc-800">
                                    {cohorts.map((cohort, idx) => (
                                        <tr key={cohort.month} className={idx % 2 === 0 ? 'bg-[#1E2026]' : 'bg-[#232323]'}>
                                            <td className="p-1.5 text-zinc-300 font-medium border-r border-zinc-800">{cohort.month}</td>
                                            <td className="p-1.5 text-center text-zinc-400 border-r border-zinc-800">{cohort.registered}</td>
                                            <td className={`p-1.5 text-center font-medium rounded border-r border-zinc-800 ${getRetentionColor(cohort.retention_30d_pct)}`}>
                                                {cohort.retention_30d_pct.toFixed(0)}%
                                            </td>
                                            <td className={`p-1.5 text-center font-medium rounded border-r border-zinc-800 ${getRetentionColor(cohort.retention_60d_pct)}`}>
                                                {cohort.retention_60d_pct.toFixed(0)}%
                                            </td>
                                            <td className={`p-1.5 text-center font-medium rounded border-r border-zinc-800 ${getRetentionColor(cohort.retention_90d_pct)}`}>
                                                {cohort.retention_90d_pct.toFixed(0)}%
                                            </td>
                                            <td className={`p-1.5 text-center font-medium rounded border-r border-zinc-800 ${getRetentionColor(cohort.retention_180d_pct)}`}>
                                                {cohort.retention_180d_pct.toFixed(0)}%
                                            </td>
                                            <td className={`p-1.5 text-center font-medium rounded ${getRetentionColor(cohort.retention_365d_pct)}`}>
                                                {cohort.retention_365d_pct.toFixed(0)}%
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>

                    {/* Monthly Trends Chart */}
                    <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
                        <h2 className="text-sm font-bold text-zinc-200 mb-3">Monthly Trends (12 Months)</h2>
                        <div className="relative h-[200px]">
                            <svg className="w-full h-full">
                                {/* Grid lines */}
                                {[0, 1, 2, 3, 4].map(i => (
                                    <line
                                        key={i}
                                        x1="40"
                                        y1={30 + (i * 35)}
                                        x2="100%"
                                        y2={30 + (i * 35)}
                                        stroke="#383A42"
                                        strokeWidth="1"
                                    />
                                ))}

                                {/* Y-axis labels */}
                                {[0, 25, 50, 75, 100].map((val, i) => (
                                    <text
                                        key={i}
                                        x="5"
                                        y={175 - (i * 35)}
                                        fill="#71717a"
                                        fontSize="9"
                                    >
                                        {val}
                                    </text>
                                ))}

                                {/* Lines */}
                                {trends.length > 1 && (
                                    <>
                                        {/* New Registrations Line */}
                                        <polyline
                                            points={trends.map((t, i) => {
                                                const x = 40 + (i * ((100 - 40) / (trends.length - 1)));
                                                const maxReg = Math.max(...trends.map(tr => tr.new_registrations));
                                                const y = 175 - ((t.new_registrations / maxReg) * 140);
                                                return `${x}%,${y}`;
                                            }).join(' ')}
                                            fill="none"
                                            stroke="#10b981"
                                            strokeWidth="2"
                                        />

                                        {/* New Activations Line */}
                                        <polyline
                                            points={trends.map((t, i) => {
                                                const x = 40 + (i * ((100 - 40) / (trends.length - 1)));
                                                const maxAct = Math.max(...trends.map(tr => tr.new_activations));
                                                const y = 175 - ((t.new_activations / maxAct) * 140);
                                                return `${x}%,${y}`;
                                            }).join(' ')}
                                            fill="none"
                                            stroke="#3b82f6"
                                            strokeWidth="2"
                                        />

                                        {/* Churn Rate Line */}
                                        <polyline
                                            points={trends.map((t, i) => {
                                                const x = 40 + (i * ((100 - 40) / (trends.length - 1)));
                                                const y = 175 - ((t.churn_rate) * 1.4);
                                                return `${x}%,${y}`;
                                            }).join(' ')}
                                            fill="none"
                                            stroke="#ef4444"
                                            strokeWidth="2"
                                        />
                                    </>
                                )}
                            </svg>
                        </div>
                        <div className="flex items-center justify-center gap-4 mt-2 text-[10px]">
                            <div className="flex items-center gap-1.5">
                                <div className="w-3 h-0.5 bg-emerald-500"></div>
                                <span className="text-zinc-400">New Registrations</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                                <div className="w-3 h-0.5 bg-blue-500"></div>
                                <span className="text-zinc-400">New Activations</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                                <div className="w-3 h-0.5 bg-red-500"></div>
                                <span className="text-zinc-400">Churn Rate</span>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Segment Comparison */}
                <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
                    <h2 className="text-sm font-bold text-zinc-200 mb-3">Segment Comparison</h2>
                    <div className="grid grid-cols-6 gap-3">
                        {segments.map(segment => (
                            <div key={segment.stage} className="bg-[#121316] border border-zinc-700 rounded p-3">
                                <div className="text-[10px] font-medium text-zinc-400 mb-2">{segment.stage}</div>
                                <div className="space-y-1.5 text-[10px]">
                                    <div className="flex justify-between">
                                        <span className="text-zinc-500">Clients:</span>
                                        <span className="text-zinc-200 font-medium">{segment.client_count}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-zinc-500">Avg Revenue:</span>
                                        <span className="text-emerald-400 font-medium">{formatMoney(segment.avg_revenue)}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-zinc-500">Avg Trades:</span>
                                        <span className="text-blue-400 font-medium">{segment.avg_trades.toFixed(0)}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-zinc-500">Avg Deposits:</span>
                                        <span className="text-zinc-300 font-medium">{formatMoney(segment.avg_deposits)}</span>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                {/* At-Risk Clients Panel */}
                <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
                    <h2 className="text-sm font-bold text-zinc-200 mb-3 flex items-center gap-2">
                        <AlertTriangle size={16} className="text-orange-500" />
                        High Churn Risk Clients ({atRiskClients.length})
                    </h2>
                    <div className="overflow-x-auto">
                        <table className="w-full text-[11px]">
                            <thead className="bg-[#2d3436] text-zinc-400">
                                <tr>
                                    <th className="p-2 text-left border-r border-zinc-700">
                                        <button onClick={() => handleSort('name')} className="flex items-center gap-1 hover:text-zinc-200">
                                            Name {sortConfig.key === 'name' && (sortConfig.direction === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                                        </button>
                                    </th>
                                    <th className="p-2 text-left border-r border-zinc-700">Email</th>
                                    <th className="p-2 text-right border-r border-zinc-700">
                                        <button onClick={() => handleSort('last_login_at')} className="flex items-center gap-1 ml-auto hover:text-zinc-200">
                                            Last Login {sortConfig.key === 'last_login_at' && (sortConfig.direction === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                                        </button>
                                    </th>
                                    <th className="p-2 text-right border-r border-zinc-700">
                                        <button onClick={() => handleSort('last_trade_at')} className="flex items-center gap-1 ml-auto hover:text-zinc-200">
                                            Last Trade {sortConfig.key === 'last_trade_at' && (sortConfig.direction === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                                        </button>
                                    </th>
                                    <th className="p-2 text-right border-r border-zinc-700">
                                        <button onClick={() => handleSort('churn_risk')} className="flex items-center gap-1 ml-auto hover:text-zinc-200">
                                            Churn Score {sortConfig.key === 'churn_risk' && (sortConfig.direction === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                                        </button>
                                    </th>
                                    <th className="p-2 text-right border-r border-zinc-700">Days Inactive</th>
                                    <th className="p-2 text-right border-r border-zinc-700">Lifetime Revenue</th>
                                    <th className="p-2 text-center">Action</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-zinc-800">
                                {sortedAtRiskClients.slice(0, 10).map((client, idx) => (
                                    <tr
                                        key={client.client_id}
                                        className={`${idx % 2 === 0 ? 'bg-[#1E2026]' : 'bg-[#232323]'} hover:bg-[#2d3436] cursor-pointer transition-colors`}
                                        onClick={() => fetchClientTimeline(client.client_id)}
                                    >
                                        <td className="p-2 text-zinc-200 font-medium border-r border-zinc-800">{client.name}</td>
                                        <td className="p-2 text-zinc-400 border-r border-zinc-800">{client.email}</td>
                                        <td className="p-2 text-right text-zinc-400 border-r border-zinc-800 text-[10px]">{formatDate(client.last_login_at)}</td>
                                        <td className="p-2 text-right text-zinc-400 border-r border-zinc-800 text-[10px]">{formatDate(client.last_trade_at)}</td>
                                        <td className="p-2 text-right border-r border-zinc-800">
                                            <span className={`px-2 py-0.5 rounded font-medium ${
                                                client.churn_risk >= 80 ? 'bg-red-500/20 text-red-300' :
                                                client.churn_risk >= 60 ? 'bg-orange-500/20 text-orange-300' :
                                                'bg-yellow-500/20 text-yellow-300'
                                            }`}>
                                                {client.churn_risk.toFixed(0)}%
                                            </span>
                                        </td>
                                        <td className="p-2 text-right text-zinc-400 border-r border-zinc-800">{client.days_since_last_trade}d</td>
                                        <td className="p-2 text-right text-emerald-400 font-medium border-r border-zinc-800">{formatMoney(client.lifetime_revenue)}</td>
                                        <td className="p-2 text-center">
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    // Handle reach out action
                                                }}
                                                className="px-2 py-1 bg-blue-500/20 text-blue-300 rounded text-[10px] font-medium hover:bg-blue-500/30 transition-colors flex items-center gap-1 mx-auto"
                                            >
                                                <Mail size={12} />
                                                Reach Out
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>

                {/* Individual Client Timeline Modal */}
                {selectedClient && clientTimeline.length > 0 && (
                    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setSelectedClient(null)}>
                        <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4 w-[600px] max-h-[80vh] overflow-auto" onClick={(e) => e.stopPropagation()}>
                            <div className="flex items-center justify-between mb-4">
                                <h2 className="text-sm font-bold text-zinc-200">Client Lifecycle Timeline</h2>
                                <button
                                    onClick={() => setSelectedClient(null)}
                                    className="text-zinc-500 hover:text-zinc-300 transition-colors"
                                >
                                    <X size={16} />
                                </button>
                            </div>
                            <div className="space-y-3">
                                {clientTimeline.map((event, idx) => (
                                    <div key={idx} className="flex gap-3">
                                        <div className="flex-shrink-0 w-20 text-right text-[10px] text-zinc-500">
                                            {formatDate(event.date)}
                                        </div>
                                        <div className="flex-shrink-0 w-3 h-3 rounded-full bg-emerald-500 mt-1"></div>
                                        <div className="flex-1">
                                            <div className="text-[11px] font-medium text-zinc-200">{event.event}</div>
                                            <div className="text-[10px] text-zinc-500">{event.description}</div>
                                            {event.value && (
                                                <div className="text-[10px] text-emerald-400 font-medium mt-0.5">{formatMoney(event.value)}</div>
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
