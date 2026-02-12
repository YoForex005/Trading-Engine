'use client';

import React, { useState, useMemo } from 'react';
import {
    Activity,
    TrendingUp,
    TrendingDown,
    Zap,
    Target,
    CheckCircle,
    AlertTriangle,
    Download,
    Filter,
    Clock,
    BarChart3
} from 'lucide-react';

// Types
type DateRange = 'today' | 'yesterday' | 'thisWeek' | 'thisMonth' | 'custom';
type ExecutionType = 'all' | 'market' | 'limit' | 'stop';

interface Execution {
    id: string;
    timestamp: Date;
    symbol: string;
    lp: string;
    executionTime: number; // milliseconds
    slippage: number; // pips
    volume: number;
    type: 'market' | 'limit' | 'stop';
    requoted: boolean;
    filled: boolean;
}

interface LPQuality {
    lp: string;
    avgLatency: number;
    fillRate: number;
    avgSlippage: number;
    requoteRate: number;
    volumeHandled: number;
    executionCount: number;
}

interface SymbolQuality {
    symbol: string;
    avgExecutionTime: number;
    avgSlippage: number;
    bestLP: string;
    worstLP: string;
    executionCount: number;
}

export default function ExecutionReport() {
    const [dateRange, setDateRange] = useState<DateRange>('today');
    const [selectedSymbol, setSelectedSymbol] = useState<string>('all');
    const [selectedLP, setSelectedLP] = useState<string>('all');
    const [executionType, setExecutionType] = useState<ExecutionType>('all');
    const [customStartDate, setCustomStartDate] = useState('');
    const [customEndDate, setCustomEndDate] = useState('');

    // Generate 200+ mock executions
    const mockExecutions: Execution[] = useMemo(() => {
        const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'NZDUSD', 'USDCHF', 'EURJPY', 'GBPJPY', 'EURGBP', 'XAUUSD', 'BTCUSD', 'ETHUSD', 'SPX500', 'NAS100'];
        const lps = ['YOFX', 'Oanda', 'Binance', 'LP-Alpha', 'LP-Beta', 'LP-Gamma'];
        const types: Array<'market' | 'limit' | 'stop'> = ['market', 'limit', 'stop'];

        const executions: Execution[] = [];
        const now = new Date();

        for (let i = 0; i < 220; i++) {
            const timestamp = new Date(now.getTime() - Math.random() * 24 * 60 * 60 * 1000);
            const symbol = symbols[Math.floor(Math.random() * symbols.length)];
            const lp = lps[Math.floor(Math.random() * lps.length)];
            const type = types[Math.floor(Math.random() * types.length)];

            // Execution time distribution (most < 100ms, some outliers)
            let executionTime: number;
            const rand = Math.random();
            if (rand < 0.4) executionTime = Math.random() * 10; // 0-10ms (40%)
            else if (rand < 0.7) executionTime = 10 + Math.random() * 40; // 10-50ms (30%)
            else if (rand < 0.85) executionTime = 50 + Math.random() * 50; // 50-100ms (15%)
            else if (rand < 0.95) executionTime = 100 + Math.random() * 400; // 100-500ms (10%)
            else executionTime = 500 + Math.random() * 1500; // 500ms+ (5%)

            const slippage = (Math.random() - 0.5) * 5; // -2.5 to +2.5 pips
            const volume = Math.random() * 10 + 0.1;
            const requoted = Math.random() < 0.08; // 8% requote rate
            const filled = Math.random() < 0.96; // 96% fill rate

            executions.push({
                id: `EXE-${10000 + i}`,
                timestamp,
                symbol,
                lp,
                executionTime,
                slippage,
                volume,
                type,
                requoted,
                filled
            });
        }

        return executions.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
    }, []);

    // Filter executions
    const filteredExecutions = useMemo(() => {
        return mockExecutions.filter(exec => {
            if (selectedSymbol !== 'all' && exec.symbol !== selectedSymbol) return false;
            if (selectedLP !== 'all' && exec.lp !== selectedLP) return false;
            if (executionType !== 'all' && exec.type !== executionType) return false;
            return true;
        });
    }, [mockExecutions, selectedSymbol, selectedLP, executionType]);

    // Calculate summary metrics
    const summaryMetrics = useMemo(() => {
        const totalExecutions = filteredExecutions.length;
        const avgExecutionTime = filteredExecutions.reduce((sum, e) => sum + e.executionTime, 0) / totalExecutions || 0;
        const avgSlippage = filteredExecutions.reduce((sum, e) => sum + Math.abs(e.slippage), 0) / totalExecutions || 0;
        const requoteRate = (filteredExecutions.filter(e => e.requoted).length / totalExecutions) * 100 || 0;
        const fillRate = (filteredExecutions.filter(e => e.filled).length / totalExecutions) * 100 || 0;

        return {
            totalExecutions,
            avgExecutionTime,
            avgSlippage,
            requoteRate,
            fillRate
        };
    }, [filteredExecutions]);

    // Execution time distribution (histogram buckets)
    const executionTimeBuckets = useMemo(() => {
        const buckets = [
            { label: '0-10ms', min: 0, max: 10, count: 0, color: '#10B981' },
            { label: '10-50ms', min: 10, max: 50, count: 0, color: '#3B82F6' },
            { label: '50-100ms', min: 50, max: 100, count: 0, color: '#F59E0B' },
            { label: '100-500ms', min: 100, max: 500, count: 0, color: '#EF4444' },
            { label: '500ms+', min: 500, max: Infinity, count: 0, color: '#DC2626' }
        ];

        filteredExecutions.forEach(exec => {
            for (const bucket of buckets) {
                if (exec.executionTime >= bucket.min && exec.executionTime < bucket.max) {
                    bucket.count++;
                    break;
                }
            }
        });

        return buckets;
    }, [filteredExecutions]);

    // LP quality analysis
    const lpQualityData: LPQuality[] = useMemo(() => {
        const lpMap = new Map<string, Execution[]>();

        filteredExecutions.forEach(exec => {
            if (!lpMap.has(exec.lp)) lpMap.set(exec.lp, []);
            lpMap.get(exec.lp)!.push(exec);
        });

        return Array.from(lpMap.entries()).map(([lp, execs]) => {
            const avgLatency = execs.reduce((sum, e) => sum + e.executionTime, 0) / execs.length;
            const fillRate = (execs.filter(e => e.filled).length / execs.length) * 100;
            const avgSlippage = execs.reduce((sum, e) => sum + e.slippage, 0) / execs.length;
            const requoteRate = (execs.filter(e => e.requoted).length / execs.length) * 100;
            const volumeHandled = execs.reduce((sum, e) => sum + e.volume, 0);

            return {
                lp,
                avgLatency,
                fillRate,
                avgSlippage,
                requoteRate,
                volumeHandled,
                executionCount: execs.length
            };
        }).sort((a, b) => a.avgLatency - b.avgLatency);
    }, [filteredExecutions]);

    // Symbol quality analysis
    const symbolQualityData: SymbolQuality[] = useMemo(() => {
        const symbolMap = new Map<string, Execution[]>();

        filteredExecutions.forEach(exec => {
            if (!symbolMap.has(exec.symbol)) symbolMap.set(exec.symbol, []);
            symbolMap.get(exec.symbol)!.push(exec);
        });

        return Array.from(symbolMap.entries()).map(([symbol, execs]) => {
            const avgExecutionTime = execs.reduce((sum, e) => sum + e.executionTime, 0) / execs.length;
            const avgSlippage = execs.reduce((sum, e) => sum + e.slippage, 0) / execs.length;

            // Find best and worst LP for this symbol
            const lpPerformance = new Map<string, number[]>();
            execs.forEach(exec => {
                if (!lpPerformance.has(exec.lp)) lpPerformance.set(exec.lp, []);
                lpPerformance.get(exec.lp)!.push(exec.executionTime);
            });

            let bestLP = '';
            let worstLP = '';
            let bestAvg = Infinity;
            let worstAvg = 0;

            lpPerformance.forEach((times, lp) => {
                const avg = times.reduce((sum, t) => sum + t, 0) / times.length;
                if (avg < bestAvg) {
                    bestAvg = avg;
                    bestLP = lp;
                }
                if (avg > worstAvg) {
                    worstAvg = avg;
                    worstLP = lp;
                }
            });

            return {
                symbol,
                avgExecutionTime,
                avgSlippage,
                bestLP,
                worstLP,
                executionCount: execs.length
            };
        }).sort((a, b) => a.avgExecutionTime - b.avgExecutionTime);
    }, [filteredExecutions]);

    // Time-of-day heatmap data (24 hours)
    const hourlyQuality = useMemo(() => {
        const hours = Array(24).fill(0).map((_, i) => ({
            hour: i,
            avgExecutionTime: 0,
            count: 0
        }));

        filteredExecutions.forEach(exec => {
            const hour = exec.timestamp.getHours();
            hours[hour].avgExecutionTime += exec.executionTime;
            hours[hour].count++;
        });

        hours.forEach(h => {
            if (h.count > 0) h.avgExecutionTime /= h.count;
        });

        return hours;
    }, [filteredExecutions]);

    // Get unique symbols and LPs for filters
    const allSymbols = useMemo(() => {
        const symbols = new Set(mockExecutions.map(e => e.symbol));
        return ['all', ...Array.from(symbols).sort()];
    }, [mockExecutions]);

    const allLPs = useMemo(() => {
        const lps = new Set(mockExecutions.map(e => e.lp));
        return ['all', ...Array.from(lps).sort()];
    }, [mockExecutions]);

    // Format number helpers
    const formatMs = (ms: number) => ms.toFixed(2);
    const formatPips = (pips: number) => pips.toFixed(2);
    const formatPercent = (val: number) => val.toFixed(2);
    const formatVolume = (vol: number) => vol.toFixed(2);

    // Render scatter plot for slippage analysis
    const renderSlippageScatter = () => {
        const width = 600;
        const height = 300;
        const padding = 40;

        const maxVolume = Math.max(...filteredExecutions.map(e => e.volume));
        const maxSlippage = Math.max(...filteredExecutions.map(e => Math.abs(e.slippage)));

        const symbolColors: Record<string, string> = {
            'EURUSD': '#3B82F6',
            'GBPUSD': '#10B981',
            'USDJPY': '#F59E0B',
            'AUDUSD': '#EF4444',
            'USDCAD': '#8B5CF6',
            'XAUUSD': '#F5C542',
            'BTCUSD': '#EC4899',
            'ETHUSD': '#06B6D4'
        };

        return (
            <svg width={width} height={height} className="bg-[#18181b] rounded">
                {/* Axes */}
                <line x1={padding} y1={height - padding} x2={width - padding} y2={height - padding} stroke="#3f3f46" strokeWidth="1" />
                <line x1={padding} y1={padding} x2={padding} y2={height - padding} stroke="#3f3f46" strokeWidth="1" />

                {/* Axis labels */}
                <text x={width / 2} y={height - 10} fill="#71717a" fontSize="10" textAnchor="middle">Volume (lots)</text>
                <text x={15} y={height / 2} fill="#71717a" fontSize="10" textAnchor="middle" transform={`rotate(-90, 15, ${height / 2})`}>Slippage (pips)</text>

                {/* Zero line for slippage */}
                <line
                    x1={padding}
                    y1={height / 2}
                    x2={width - padding}
                    y2={height / 2}
                    stroke="#3f3f46"
                    strokeWidth="1"
                    strokeDasharray="4 2"
                />

                {/* Data points */}
                {filteredExecutions.slice(0, 100).map((exec, idx) => {
                    const x = padding + ((exec.volume / maxVolume) * (width - 2 * padding));
                    const y = (height / 2) - ((exec.slippage / maxSlippage) * ((height / 2) - padding));
                    const color = symbolColors[exec.symbol] || '#71717a';

                    return (
                        <circle
                            key={idx}
                            cx={x}
                            cy={y}
                            r="3"
                            fill={color}
                            opacity="0.6"
                        />
                    );
                })}
            </svg>
        );
    };

    // Render time-of-day heatmap
    const renderHourlyHeatmap = () => {
        const cellWidth = 30;
        const cellHeight = 40;
        const maxTime = Math.max(...hourlyQuality.map(h => h.avgExecutionTime));

        return (
            <div className="flex gap-1">
                {hourlyQuality.map(h => {
                    const intensity = h.count > 0 ? h.avgExecutionTime / maxTime : 0;
                    const color = intensity < 0.3 ? '#10B981' : intensity < 0.6 ? '#F59E0B' : '#EF4444';

                    return (
                        <div key={h.hour} className="flex flex-col items-center">
                            <div
                                className="rounded"
                                style={{
                                    width: cellWidth,
                                    height: cellHeight,
                                    backgroundColor: h.count > 0 ? color : '#27272a',
                                    opacity: h.count > 0 ? 0.3 + (intensity * 0.7) : 0.2
                                }}
                                title={`${h.hour}:00 - Avg: ${formatMs(h.avgExecutionTime)}ms (${h.count} execs)`}
                            />
                            <span className="text-[9px] text-zinc-500 mt-1">{h.hour}</span>
                        </div>
                    );
                })}
            </div>
        );
    };

    return (
        <div className="h-full overflow-auto bg-[#121316] p-4">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-xl font-bold text-zinc-100 flex items-center gap-2">
                        <Activity size={20} className="text-blue-400" />
                        Execution Quality Report
                    </h1>
                    <p className="text-xs text-zinc-500 mt-1">Trade execution analysis and fill quality metrics</p>
                </div>
                <button className="flex items-center gap-2 px-4 py-2 bg-[#27272a] hover:bg-[#3f3f46] text-zinc-300 rounded text-xs transition-colors">
                    <Download size={14} />
                    Export Report
                </button>
            </div>

            {/* Filters */}
            <div className="grid grid-cols-5 gap-3 mb-6 p-4 bg-[#18181b] rounded-lg border border-[#27272a]">
                <div>
                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Date Range</label>
                    <select
                        value={dateRange}
                        onChange={(e) => setDateRange(e.target.value as DateRange)}
                        className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500"
                    >
                        <option value="today">Today</option>
                        <option value="yesterday">Yesterday</option>
                        <option value="thisWeek">This Week</option>
                        <option value="thisMonth">This Month</option>
                        <option value="custom">Custom Range</option>
                    </select>
                </div>

                <div>
                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Symbol</label>
                    <select
                        value={selectedSymbol}
                        onChange={(e) => setSelectedSymbol(e.target.value)}
                        className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500"
                    >
                        {allSymbols.map(s => (
                            <option key={s} value={s}>{s === 'all' ? 'All Symbols' : s}</option>
                        ))}
                    </select>
                </div>

                <div>
                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Liquidity Provider</label>
                    <select
                        value={selectedLP}
                        onChange={(e) => setSelectedLP(e.target.value)}
                        className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500"
                    >
                        {allLPs.map(lp => (
                            <option key={lp} value={lp}>{lp === 'all' ? 'All LPs' : lp}</option>
                        ))}
                    </select>
                </div>

                <div>
                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Execution Type</label>
                    <select
                        value={executionType}
                        onChange={(e) => setExecutionType(e.target.value as ExecutionType)}
                        className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500"
                    >
                        <option value="all">All Types</option>
                        <option value="market">Market Orders</option>
                        <option value="limit">Limit Orders</option>
                        <option value="stop">Stop Orders</option>
                    </select>
                </div>

                <div className="flex items-end">
                    <button className="w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded text-xs transition-colors flex items-center justify-center gap-2">
                        <Filter size={12} />
                        Apply Filters
                    </button>
                </div>
            </div>

            {/* Summary Cards */}
            <div className="grid grid-cols-5 gap-4 mb-6">
                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] text-zinc-500 uppercase font-semibold">Total Executions</span>
                        <Activity size={14} className="text-blue-400" />
                    </div>
                    <div className="text-2xl font-bold text-zinc-100">{summaryMetrics.totalExecutions.toLocaleString()}</div>
                    <div className="text-[10px] text-zinc-500 mt-1">Today</div>
                </div>

                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] text-zinc-500 uppercase font-semibold">Avg Execution Time</span>
                        <Clock size={14} className="text-emerald-400" />
                    </div>
                    <div className="text-2xl font-bold text-zinc-100">{formatMs(summaryMetrics.avgExecutionTime)}<span className="text-sm text-zinc-500 ml-1">ms</span></div>
                    <div className="text-[10px] text-emerald-400 mt-1 flex items-center gap-1">
                        <TrendingDown size={10} />
                        12% faster than yesterday
                    </div>
                </div>

                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] text-zinc-500 uppercase font-semibold">Slippage Rate</span>
                        <Target size={14} className="text-yellow-400" />
                    </div>
                    <div className="text-2xl font-bold text-zinc-100">{formatPips(summaryMetrics.avgSlippage)}<span className="text-sm text-zinc-500 ml-1">pips</span></div>
                    <div className="text-[10px] text-yellow-400 mt-1">Average absolute slippage</div>
                </div>

                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] text-zinc-500 uppercase font-semibold">Requote Rate</span>
                        <AlertTriangle size={14} className="text-orange-400" />
                    </div>
                    <div className="text-2xl font-bold text-zinc-100">{formatPercent(summaryMetrics.requoteRate)}<span className="text-sm text-zinc-500 ml-1">%</span></div>
                    <div className="text-[10px] text-emerald-400 mt-1 flex items-center gap-1">
                        <TrendingDown size={10} />
                        Below industry average
                    </div>
                </div>

                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] text-zinc-500 uppercase font-semibold">Fill Rate</span>
                        <CheckCircle size={14} className="text-emerald-400" />
                    </div>
                    <div className="text-2xl font-bold text-zinc-100">{formatPercent(summaryMetrics.fillRate)}<span className="text-sm text-zinc-500 ml-1">%</span></div>
                    <div className="text-[10px] text-emerald-400 mt-1">Excellent fill quality</div>
                </div>
            </div>

            {/* Charts Row */}
            <div className="grid grid-cols-2 gap-6 mb-6">
                {/* Execution Time Distribution Histogram */}
                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <h3 className="text-sm font-semibold text-zinc-100 mb-4 flex items-center gap-2">
                        <BarChart3 size={14} className="text-blue-400" />
                        Execution Time Distribution
                    </h3>
                    <svg width="100%" height="220" className="bg-[#18181b]">
                        {executionTimeBuckets.map((bucket, idx) => {
                            const maxCount = Math.max(...executionTimeBuckets.map(b => b.count));
                            const barHeight = (bucket.count / maxCount) * 160;
                            const barWidth = 80;
                            const x = 40 + idx * (barWidth + 20);
                            const y = 180 - barHeight;

                            return (
                                <g key={idx}>
                                    <rect
                                        x={x}
                                        y={y}
                                        width={barWidth}
                                        height={barHeight}
                                        fill={bucket.color}
                                        opacity="0.7"
                                        rx="2"
                                    />
                                    <text
                                        x={x + barWidth / 2}
                                        y={y - 5}
                                        fill="#a1a1aa"
                                        fontSize="11"
                                        textAnchor="middle"
                                    >
                                        {bucket.count}
                                    </text>
                                    <text
                                        x={x + barWidth / 2}
                                        y={195}
                                        fill="#71717a"
                                        fontSize="10"
                                        textAnchor="middle"
                                    >
                                        {bucket.label}
                                    </text>
                                </g>
                            );
                        })}
                    </svg>
                </div>

                {/* Slippage Analysis Scatter */}
                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <h3 className="text-sm font-semibold text-zinc-100 mb-4 flex items-center gap-2">
                        <Target size={14} className="text-yellow-400" />
                        Slippage vs Volume Analysis
                    </h3>
                    {renderSlippageScatter()}
                    <div className="flex gap-3 mt-2 text-[9px] text-zinc-500">
                        <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-[#3B82F6]"></div> EURUSD</span>
                        <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-[#10B981]"></div> GBPUSD</span>
                        <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-[#F59E0B]"></div> USDJPY</span>
                        <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-[#EF4444]"></div> AUDUSD</span>
                    </div>
                </div>
            </div>

            {/* Time-of-Day Heatmap */}
            <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a] mb-6">
                <h3 className="text-sm font-semibold text-zinc-100 mb-4 flex items-center gap-2">
                    <Clock size={14} className="text-blue-400" />
                    Execution Quality by Hour (24h)
                </h3>
                {renderHourlyHeatmap()}
                <div className="flex gap-4 mt-3 text-[10px] text-zinc-500">
                    <span className="flex items-center gap-2"><div className="w-4 h-3 rounded bg-[#10B981] opacity-60"></div> Fast (&lt;30ms avg)</span>
                    <span className="flex items-center gap-2"><div className="w-4 h-3 rounded bg-[#F59E0B] opacity-60"></div> Moderate (30-60ms)</span>
                    <span className="flex items-center gap-2"><div className="w-4 h-3 rounded bg-[#EF4444] opacity-60"></div> Slow (&gt;60ms)</span>
                </div>
            </div>

            {/* LP Quality Table */}
            <div className="bg-[#18181b] rounded-lg border border-[#27272a] mb-6 overflow-hidden">
                <div className="p-4 border-b border-[#27272a]">
                    <h3 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                        <Zap size={14} className="text-emerald-400" />
                        Execution Quality by Liquidity Provider
                    </h3>
                </div>
                <div className="overflow-x-auto">
                    <table className="w-full text-xs">
                        <thead className="bg-[#27272a] text-zinc-400 uppercase text-[10px]">
                            <tr>
                                <th className="text-left p-3 font-semibold">LP Name</th>
                                <th className="text-right p-3 font-semibold">Avg Latency</th>
                                <th className="text-right p-3 font-semibold">Fill Rate</th>
                                <th className="text-right p-3 font-semibold">Avg Slippage</th>
                                <th className="text-right p-3 font-semibold">Requote Rate</th>
                                <th className="text-right p-3 font-semibold">Volume</th>
                                <th className="text-right p-3 font-semibold">Executions</th>
                                <th className="text-center p-3 font-semibold">Rating</th>
                            </tr>
                        </thead>
                        <tbody className="text-zinc-300">
                            {lpQualityData.map((lp, idx) => {
                                const isTop3 = idx < 3;
                                const isBottom3 = idx >= lpQualityData.length - 3;

                                return (
                                    <tr key={lp.lp} className={`border-t border-[#27272a] hover:bg-[#27272a]/50 ${isTop3 ? 'bg-emerald-500/5' : isBottom3 ? 'bg-red-500/5' : ''}`}>
                                        <td className="p-3">
                                            <div className="flex items-center gap-2">
                                                {isTop3 && <span className="text-emerald-400">⭐</span>}
                                                {isBottom3 && <span className="text-red-400">⚠️</span>}
                                                <span className="font-medium">{lp.lp}</span>
                                            </div>
                                        </td>
                                        <td className="text-right p-3">
                                            <span className={lp.avgLatency < 30 ? 'text-emerald-400' : lp.avgLatency > 100 ? 'text-red-400' : 'text-zinc-300'}>
                                                {formatMs(lp.avgLatency)} ms
                                            </span>
                                        </td>
                                        <td className="text-right p-3">
                                            <span className={lp.fillRate > 95 ? 'text-emerald-400' : lp.fillRate < 90 ? 'text-red-400' : 'text-zinc-300'}>
                                                {formatPercent(lp.fillRate)}%
                                            </span>
                                        </td>
                                        <td className="text-right p-3">
                                            <span className={Math.abs(lp.avgSlippage) < 0.5 ? 'text-emerald-400' : Math.abs(lp.avgSlippage) > 1.5 ? 'text-red-400' : 'text-zinc-300'}>
                                                {formatPips(lp.avgSlippage)} pips
                                            </span>
                                        </td>
                                        <td className="text-right p-3">
                                            <span className={lp.requoteRate < 5 ? 'text-emerald-400' : lp.requoteRate > 10 ? 'text-red-400' : 'text-zinc-300'}>
                                                {formatPercent(lp.requoteRate)}%
                                            </span>
                                        </td>
                                        <td className="text-right p-3 text-zinc-400">{formatVolume(lp.volumeHandled)} lots</td>
                                        <td className="text-right p-3 text-zinc-400">{lp.executionCount}</td>
                                        <td className="text-center p-3">
                                            {isTop3 && <span className="px-2 py-1 bg-emerald-500/20 text-emerald-400 rounded text-[10px] font-semibold">Best</span>}
                                            {isBottom3 && <span className="px-2 py-1 bg-red-500/20 text-red-400 rounded text-[10px] font-semibold">Worst</span>}
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Symbol Quality Table */}
            <div className="bg-[#18181b] rounded-lg border border-[#27272a] overflow-hidden">
                <div className="p-4 border-b border-[#27272a]">
                    <h3 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                        <Activity size={14} className="text-blue-400" />
                        Execution Quality by Symbol
                    </h3>
                </div>
                <div className="overflow-x-auto">
                    <table className="w-full text-xs">
                        <thead className="bg-[#27272a] text-zinc-400 uppercase text-[10px]">
                            <tr>
                                <th className="text-left p-3 font-semibold">Symbol</th>
                                <th className="text-right p-3 font-semibold">Avg Execution Time</th>
                                <th className="text-right p-3 font-semibold">Avg Slippage</th>
                                <th className="text-left p-3 font-semibold">Best LP</th>
                                <th className="text-left p-3 font-semibold">Worst LP</th>
                                <th className="text-right p-3 font-semibold">Executions</th>
                            </tr>
                        </thead>
                        <tbody className="text-zinc-300">
                            {symbolQualityData.map((sym, idx) => (
                                <tr key={sym.symbol} className="border-t border-[#27272a] hover:bg-[#27272a]/50">
                                    <td className="p-3">
                                        <span className="font-medium text-zinc-100">{sym.symbol}</span>
                                    </td>
                                    <td className="text-right p-3">
                                        <span className={sym.avgExecutionTime < 30 ? 'text-emerald-400' : sym.avgExecutionTime > 100 ? 'text-red-400' : 'text-zinc-300'}>
                                            {formatMs(sym.avgExecutionTime)} ms
                                        </span>
                                    </td>
                                    <td className="text-right p-3">
                                        <span className={Math.abs(sym.avgSlippage) < 0.5 ? 'text-emerald-400' : Math.abs(sym.avgSlippage) > 1.5 ? 'text-red-400' : 'text-zinc-300'}>
                                            {formatPips(sym.avgSlippage)} pips
                                        </span>
                                    </td>
                                    <td className="p-3">
                                        <span className="px-2 py-0.5 bg-emerald-500/20 text-emerald-400 rounded text-[10px]">{sym.bestLP}</span>
                                    </td>
                                    <td className="p-3">
                                        <span className="px-2 py-0.5 bg-red-500/20 text-red-400 rounded text-[10px]">{sym.worstLP}</span>
                                    </td>
                                    <td className="text-right p-3 text-zinc-400">{sym.executionCount}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
