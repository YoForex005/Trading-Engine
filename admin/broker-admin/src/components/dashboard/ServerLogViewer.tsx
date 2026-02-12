'use client';

import React, { useState, useEffect, useRef, useMemo } from 'react';
import {
    Terminal,
    Play,
    Pause,
    Download,
    Search,
    ChevronDown,
    ChevronRight,
    AlertTriangle,
    AlertCircle,
    Info,
    Zap,
    Clock,
    Activity,
    TrendingUp,
    XCircle
} from 'lucide-react';

// Types
type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'FATAL';
type TimeRange = '15min' | '1h' | '6h' | '24h' | 'custom';

interface LogEntry {
    id: string;
    timestamp: Date;
    level: LogLevel;
    module: string;
    message: string;
    stackTrace?: string;
    details?: Record<string, any>;
}

interface ErrorSummary {
    message: string;
    count: number;
    lastOccurrence: Date;
    module: string;
}

interface DashboardStats {
    errorsLastHour: number;
    warningsLastHour: number;
    avgResponseTime: number;
    uptimePercentage: number;
}

const LOG_LEVEL_COLORS: Record<LogLevel, { bg: string; text: string; icon: React.ReactNode }> = {
    DEBUG: { bg: 'bg-[#888]/10', text: 'text-[#888]', icon: <Info size={12} /> },
    INFO: { bg: 'bg-[#3B82F6]/10', text: 'text-[#3B82F6]', icon: <Info size={12} /> },
    WARN: { bg: 'bg-[#F59E0B]/10', text: 'text-[#F59E0B]', icon: <AlertTriangle size={12} /> },
    ERROR: { bg: 'bg-[#EF4444]/10', text: 'text-[#EF4444]', icon: <XCircle size={12} /> },
    FATAL: { bg: 'bg-[#DC2626]/20', text: 'text-[#DC2626]', icon: <AlertCircle size={12} /> },
};

const MODULES = [
    'trading-engine',
    'order-manager',
    'risk-engine',
    'market-data',
    'websocket-server',
    'database',
    'auth-service',
    'api-gateway',
    'payment-processor',
    'notification-service',
];

export default function ServerLogViewer() {
    const [logs, setLogs] = useState<LogEntry[]>(() => generateMockLogs());
    const [filteredLogs, setFilteredLogs] = useState<LogEntry[]>([]);
    const [expandedLogId, setExpandedLogId] = useState<string | null>(null);
    const [autoRefresh, setAutoRefresh] = useState(true);
    const [timeRange, setTimeRange] = useState<TimeRange>('1h');
    const logContainerRef = useRef<HTMLDivElement>(null);

    // Filters
    const [levelFilters, setLevelFilters] = useState<Set<LogLevel>>(
        new Set(['DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'])
    );
    const [moduleFilter, setModuleFilter] = useState<string>('all');
    const [searchQuery, setSearchQuery] = useState('');

    // Auto-scroll
    useEffect(() => {
        if (autoRefresh && logContainerRef.current) {
            logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
        }
    }, [filteredLogs, autoRefresh]);

    // Auto-refresh simulation (add new logs every 5s)
    useEffect(() => {
        if (!autoRefresh) return;

        const interval = setInterval(() => {
            const newLog = generateSingleLog();
            setLogs(prev => [...prev, newLog].slice(-500)); // Keep last 500
        }, 5000);

        return () => clearInterval(interval);
    }, [autoRefresh]);

    // Calculate dashboard stats
    const stats: DashboardStats = useMemo(() => {
        const oneHourAgo = new Date(Date.now() - 60 * 60 * 1000);
        const recentLogs = logs.filter(log => log.timestamp >= oneHourAgo);

        return {
            errorsLastHour: recentLogs.filter(log => log.level === 'ERROR' || log.level === 'FATAL').length,
            warningsLastHour: recentLogs.filter(log => log.level === 'WARN').length,
            avgResponseTime: 145 + Math.floor(Math.random() * 50), // Simulated
            uptimePercentage: 99.8 + Math.random() * 0.2, // Simulated
        };
    }, [logs]);

    // Calculate error rate chart data (errors per minute over last 60 minutes)
    const errorRateData = useMemo(() => {
        const data: { minute: number; count: number }[] = [];
        const now = Date.now();

        for (let i = 59; i >= 0; i--) {
            const minuteStart = now - i * 60 * 1000;
            const minuteEnd = minuteStart + 60 * 1000;
            const count = logs.filter(
                log =>
                    (log.level === 'ERROR' || log.level === 'FATAL') &&
                    log.timestamp.getTime() >= minuteStart &&
                    log.timestamp.getTime() < minuteEnd
            ).length;
            data.push({ minute: 59 - i, count });
        }

        return data;
    }, [logs]);

    // Calculate top 5 errors
    const topErrors: ErrorSummary[] = useMemo(() => {
        const errorMap = new Map<string, ErrorSummary>();

        logs
            .filter(log => log.level === 'ERROR' || log.level === 'FATAL')
            .forEach(log => {
                const key = log.message.split(':')[0]; // Group by error message prefix
                const existing = errorMap.get(key);
                if (existing) {
                    existing.count++;
                    if (log.timestamp > existing.lastOccurrence) {
                        existing.lastOccurrence = log.timestamp;
                    }
                } else {
                    errorMap.set(key, {
                        message: key,
                        count: 1,
                        lastOccurrence: log.timestamp,
                        module: log.module,
                    });
                }
            });

        return Array.from(errorMap.values())
            .sort((a, b) => b.count - a.count)
            .slice(0, 5);
    }, [logs]);

    // Apply filters
    useEffect(() => {
        let filtered = logs;

        // Time range filter
        const now = Date.now();
        const ranges: Record<TimeRange, number> = {
            '15min': 15 * 60 * 1000,
            '1h': 60 * 60 * 1000,
            '6h': 6 * 60 * 60 * 1000,
            '24h': 24 * 60 * 60 * 1000,
            'custom': 24 * 60 * 60 * 1000, // Default to 24h for custom
        };
        const rangeMs = ranges[timeRange];
        filtered = filtered.filter(log => now - log.timestamp.getTime() <= rangeMs);

        // Level filter
        filtered = filtered.filter(log => levelFilters.has(log.level));

        // Module filter
        if (moduleFilter !== 'all') {
            filtered = filtered.filter(log => log.module === moduleFilter);
        }

        // Search filter
        if (searchQuery) {
            const query = searchQuery.toLowerCase();
            filtered = filtered.filter(
                log =>
                    log.message.toLowerCase().includes(query) ||
                    log.module.toLowerCase().includes(query)
            );
        }

        setFilteredLogs(filtered);
    }, [logs, levelFilters, moduleFilter, searchQuery, timeRange]);

    const toggleLevelFilter = (level: LogLevel) => {
        const newFilters = new Set(levelFilters);
        if (newFilters.has(level)) {
            newFilters.delete(level);
        } else {
            newFilters.add(level);
        }
        setLevelFilters(newFilters);
    };

    const handleExportCSV = () => {
        const csvHeaders = ['Timestamp', 'Level', 'Module', 'Message'];
        const csvRows = filteredLogs.map(log => [
            log.timestamp.toISOString(),
            log.level,
            log.module,
            log.message.replace(/,/g, ';'), // Escape commas
        ]);

        const csvContent = [
            csvHeaders.join(','),
            ...csvRows.map(row => row.map(cell => `"${cell}"`).join(','))
        ].join('\n');

        const blob = new Blob([csvContent], { type: 'text/csv' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `server-logs-${new Date().toISOString()}.csv`;
        a.click();
        URL.revokeObjectURL(url);
    };

    // Generate SVG line chart path
    const chartPath = useMemo(() => {
        if (errorRateData.length === 0) return '';

        const maxCount = Math.max(...errorRateData.map(d => d.count), 1);
        const width = 800;
        const height = 100;
        const xStep = width / (errorRateData.length - 1);

        const points = errorRateData.map((d, i) => {
            const x = i * xStep;
            const y = height - (d.count / maxCount) * height;
            return `${x},${y}`;
        });

        return `M ${points.join(' L ')}`;
    }, [errorRateData]);

    return (
        <div className="h-full flex flex-col bg-[#121316] text-[#E5E7EB]">
            {/* Header */}
            <div className="flex-shrink-0 px-6 py-4 border-b border-[#383A42]">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <Terminal className="text-[#10B981]" size={24} />
                        <h1 className="text-xl font-bold text-[#F5C542]">Server Log Viewer</h1>
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            onClick={() => setAutoRefresh(!autoRefresh)}
                            className={`flex items-center gap-2 px-4 py-2 rounded transition-colors text-sm font-medium ${
                                autoRefresh
                                    ? 'bg-[#10B981] text-white hover:bg-[#059669]'
                                    : 'bg-[#383A42] text-[#888] hover:bg-[#4B5563]'
                            }`}
                        >
                            {autoRefresh ? <Pause size={16} /> : <Play size={16} />}
                            {autoRefresh ? 'Live' : 'Paused'}
                        </button>
                        <button
                            onClick={handleExportCSV}
                            className="flex items-center gap-2 px-4 py-2 bg-[#1E2026] border border-[#383A42] rounded hover:bg-[#25272E] transition-colors text-sm"
                        >
                            <Download size={16} />
                            Export CSV
                        </button>
                    </div>
                </div>

                {/* Error Dashboard Stats */}
                <div className="grid grid-cols-4 gap-4 mb-4">
                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Errors (1h)</span>
                            <XCircle size={16} className="text-[#EF4444]" />
                        </div>
                        <div className="text-2xl font-bold text-[#EF4444]">{stats.errorsLastHour}</div>
                    </div>

                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Warnings (1h)</span>
                            <AlertTriangle size={16} className="text-[#F59E0B]" />
                        </div>
                        <div className="text-2xl font-bold text-[#F59E0B]">{stats.warningsLastHour}</div>
                    </div>

                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Avg Response</span>
                            <Clock size={16} className="text-[#3B82F6]" />
                        </div>
                        <div className="text-2xl font-bold text-[#F5C542]">{stats.avgResponseTime}ms</div>
                    </div>

                    <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-[#888]">Uptime</span>
                            <Activity size={16} className="text-[#10B981]" />
                        </div>
                        <div className="text-2xl font-bold text-[#10B981]">{stats.uptimePercentage.toFixed(2)}%</div>
                    </div>
                </div>

                {/* Error Rate Chart */}
                <div className="bg-[#1E2026] border border-[#383A42] rounded p-4 mb-4">
                    <div className="text-xs font-medium text-[#888] mb-3">Error Rate (errors/min over last 60 min)</div>
                    <svg viewBox="0 0 800 100" className="w-full h-24">
                        <path
                            d={chartPath}
                            fill="none"
                            stroke="#EF4444"
                            strokeWidth="2"
                        />
                        <path
                            d={`${chartPath} L 800,100 L 0,100 Z`}
                            fill="url(#errorGradient)"
                        />
                        <defs>
                            <linearGradient id="errorGradient" x1="0" x2="0" y1="0" y2="1">
                                <stop offset="0%" stopColor="#EF4444" stopOpacity="0.3" />
                                <stop offset="100%" stopColor="#EF4444" stopOpacity="0" />
                            </linearGradient>
                        </defs>
                    </svg>
                </div>

                {/* Top 5 Errors Table */}
                <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
                    <div className="text-xs font-medium text-[#888] mb-3">Top 5 Most Frequent Errors</div>
                    <table className="w-full text-xs">
                        <thead>
                            <tr className="border-b border-[#383A42]">
                                <th className="text-left py-2 text-[#888] font-medium">Error Message</th>
                                <th className="text-left py-2 text-[#888] font-medium">Count</th>
                                <th className="text-left py-2 text-[#888] font-medium">Last Occurrence</th>
                                <th className="text-left py-2 text-[#888] font-medium">Module</th>
                            </tr>
                        </thead>
                        <tbody>
                            {topErrors.map((error, i) => (
                                <tr key={i} className="border-b border-[#383A42]/50">
                                    <td className="py-2 text-[#E5E7EB] font-mono text-xs">{error.message}</td>
                                    <td className="py-2 text-[#EF4444] font-bold">{error.count}</td>
                                    <td className="py-2 text-[#888]">{error.lastOccurrence.toLocaleTimeString()}</td>
                                    <td className="py-2 text-[#3B82F6]">{error.module}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Filter Bar */}
            <div className="flex-shrink-0 px-6 py-3 bg-[#1E2026] border-b border-[#383A42]">
                <div className="flex items-center gap-3 flex-wrap">
                    {/* Time Range Selector */}
                    <select
                        value={timeRange}
                        onChange={(e) => setTimeRange(e.target.value as TimeRange)}
                        className="bg-[#121316] border border-[#383A42] rounded px-3 py-2 text-xs outline-none text-[#E5E7EB]"
                    >
                        <option value="15min">Last 15 min</option>
                        <option value="1h">Last 1 hour</option>
                        <option value="6h">Last 6 hours</option>
                        <option value="24h">Last 24 hours</option>
                        <option value="custom">Custom</option>
                    </select>

                    {/* Level Filters */}
                    <div className="flex items-center gap-2">
                        {(['DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'] as LogLevel[]).map(level => (
                            <label
                                key={level}
                                className={`flex items-center gap-1 px-2 py-1 rounded cursor-pointer transition-colors text-xs ${
                                    levelFilters.has(level)
                                        ? `${LOG_LEVEL_COLORS[level].bg} ${LOG_LEVEL_COLORS[level].text}`
                                        : 'bg-[#383A42]/30 text-[#666]'
                                }`}
                            >
                                <input
                                    type="checkbox"
                                    checked={levelFilters.has(level)}
                                    onChange={() => toggleLevelFilter(level)}
                                    className="w-3 h-3"
                                />
                                <span className="font-medium">{level}</span>
                            </label>
                        ))}
                    </div>

                    {/* Module Filter */}
                    <select
                        value={moduleFilter}
                        onChange={(e) => setModuleFilter(e.target.value)}
                        className="bg-[#121316] border border-[#383A42] rounded px-3 py-2 text-xs outline-none text-[#E5E7EB]"
                    >
                        <option value="all">All Modules</option>
                        {MODULES.map(module => (
                            <option key={module} value={module}>{module}</option>
                        ))}
                    </select>

                    {/* Search */}
                    <div className="flex items-center gap-2 bg-[#121316] border border-[#383A42] rounded px-3 py-2 flex-1 max-w-md">
                        <Search size={14} className="text-[#888]" />
                        <input
                            type="text"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="bg-transparent text-xs outline-none flex-1 text-[#E5E7EB]"
                            placeholder="Search logs..."
                        />
                    </div>
                </div>
            </div>

            {/* Log Stream */}
            <div ref={logContainerRef} className="flex-1 overflow-auto p-4 font-mono text-xs bg-[#0D0E11]">
                {filteredLogs.map(log => {
                    const isExpanded = expandedLogId === log.id;
                    const levelStyle = LOG_LEVEL_COLORS[log.level];

                    return (
                        <div
                            key={log.id}
                            className="mb-1 hover:bg-[#1E2026]/50 transition-colors cursor-pointer"
                            onClick={() => setExpandedLogId(isExpanded ? null : log.id)}
                        >
                            <div className="flex items-start gap-2 py-1 px-2">
                                {isExpanded ? <ChevronDown size={12} className="text-[#888] mt-1" /> : <ChevronRight size={12} className="text-[#888] mt-1" />}
                                <span className="text-[#888] whitespace-nowrap">
                                    {log.timestamp.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })}.
                                    {log.timestamp.getMilliseconds().toString().padStart(3, '0')}
                                </span>
                                <span className={`px-2 py-0.5 rounded font-bold ${levelStyle.bg} ${levelStyle.text} flex items-center gap-1 whitespace-nowrap`}>
                                    {levelStyle.icon}
                                    {log.level}
                                </span>
                                <span className="text-[#3B82F6] whitespace-nowrap">[{log.module}]</span>
                                <span className="text-[#E5E7EB] flex-1">{log.message}</span>
                            </div>

                            {/* Expanded Details */}
                            {isExpanded && (log.stackTrace || log.details) && (
                                <div className="ml-12 mt-2 mb-3 p-3 bg-[#121316] border border-[#383A42] rounded">
                                    {log.stackTrace && (
                                        <div className="mb-2">
                                            <div className="text-[#888] text-xs mb-1">Stack Trace:</div>
                                            <pre className="text-[#E5E7EB] text-xs whitespace-pre-wrap">{log.stackTrace}</pre>
                                        </div>
                                    )}
                                    {log.details && (
                                        <div>
                                            <div className="text-[#888] text-xs mb-1">Details:</div>
                                            <pre className="text-[#E5E7EB] text-xs">{JSON.stringify(log.details, null, 2)}</pre>
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    );
                })}

                {filteredLogs.length === 0 && (
                    <div className="flex flex-col items-center justify-center py-16 text-[#666]">
                        <Terminal size={48} className="mb-4 opacity-20" />
                        <div className="text-lg">No logs match the current filters</div>
                        <div className="text-xs mt-2">Try adjusting your filters or time range</div>
                    </div>
                )}
            </div>

            {/* Footer */}
            <div className="flex-shrink-0 px-6 py-2 bg-[#1E2026] border-t border-[#383A42] flex items-center justify-between text-xs text-[#888]">
                <div>Showing {filteredLogs.length} of {logs.length} log entries</div>
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-[#888]"></div>
                        <span>DEBUG: {logs.filter(l => l.level === 'DEBUG').length}</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-[#3B82F6]"></div>
                        <span>INFO: {logs.filter(l => l.level === 'INFO').length}</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-[#F59E0B]"></div>
                        <span>WARN: {logs.filter(l => l.level === 'WARN').length}</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-[#EF4444]"></div>
                        <span>ERROR: {logs.filter(l => l.level === 'ERROR').length}</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-[#DC2626]"></div>
                        <span>FATAL: {logs.filter(l => l.level === 'FATAL').length}</span>
                    </div>
                </div>
            </div>
        </div>
    );
}

// Generate mock logs
function generateMockLogs(): LogEntry[] {
    const logs: LogEntry[] = [];
    const now = Date.now();

    // Distribution: 70% INFO, 15% WARN, 10% ERROR, 4% DEBUG, 1% FATAL
    const distribution: LogLevel[] = [
        ...Array(70).fill('INFO'),
        ...Array(15).fill('WARN'),
        ...Array(10).fill('ERROR'),
        ...Array(4).fill('DEBUG'),
        ...Array(1).fill('FATAL'),
    ];

    const messages = {
        DEBUG: ['Database query took 12ms', 'Cache hit for key: user:123', 'HTTP request to /api/trades'],
        INFO: ['Trade executed successfully', 'User logged in', 'Market data updated', 'Order placed', 'Withdrawal processed'],
        WARN: ['High memory usage detected', 'Slow query detected (>500ms)', 'API rate limit approaching', 'Connection pool nearing capacity'],
        ERROR: ['Database connection failed', 'Failed to process order', 'API timeout after 30s', 'Invalid authentication token', 'Payment gateway error'],
        FATAL: ['Out of memory', 'Database server unreachable', 'Critical system failure'],
    };

    const stackTraces = [
        'at tradingEngine.processOrder (engine.go:234)\nat orderManager.handleOrder (manager.go:89)\nat httpHandler.ServeHTTP (handler.go:45)',
        'at database.Connect (db.go:123)\nat main.initialize (main.go:56)\nat runtime.main (runtime.go:250)',
        'at apiGateway.handleRequest (gateway.go:178)\nat httpServer.handleConnection (server.go:92)',
    ];

    for (let i = 0; i < 500; i++) {
        const level = distribution[Math.floor(Math.random() * distribution.length)] as LogLevel;
        const module = MODULES[Math.floor(Math.random() * MODULES.length)];
        const message = messages[level][Math.floor(Math.random() * messages[level].length)];
        const hasStackTrace = (level === 'ERROR' || level === 'FATAL') && Math.random() > 0.5;

        logs.push({
            id: `log-${i}`,
            timestamp: new Date(now - (500 - i) * 60 * 1000 - Math.random() * 60 * 1000),
            level,
            module,
            message: `${message}${level === 'ERROR' || level === 'FATAL' ? `: ${['timeout', 'connection refused', 'null pointer', 'invalid input'][Math.floor(Math.random() * 4)]}` : ''}`,
            stackTrace: hasStackTrace ? stackTraces[Math.floor(Math.random() * stackTraces.length)] : undefined,
            details: hasStackTrace ? {
                errorCode: Math.floor(Math.random() * 5000),
                requestId: `req-${Math.random().toString(36).substr(2, 9)}`,
            } : undefined,
        });
    }

    return logs.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime());
}

// Generate single log entry (for auto-refresh)
function generateSingleLog(): LogEntry {
    const levels: LogLevel[] = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'];
    const weights = [0.04, 0.70, 0.15, 0.10, 0.01];
    const rand = Math.random();
    let cumulative = 0;
    let level: LogLevel = 'INFO';

    for (let i = 0; i < weights.length; i++) {
        cumulative += weights[i];
        if (rand <= cumulative) {
            level = levels[i];
            break;
        }
    }

    const module = MODULES[Math.floor(Math.random() * MODULES.length)];
    const messages = {
        DEBUG: ['Cache miss for key: session:' + Math.random().toString(36).substr(2, 5)],
        INFO: ['Request processed in ' + Math.floor(Math.random() * 200) + 'ms'],
        WARN: ['Slow query detected (>500ms)'],
        ERROR: ['API timeout'],
        FATAL: ['System failure'],
    };
    const message = messages[level][Math.floor(Math.random() * messages[level].length)];

    return {
        id: `log-${Date.now()}-${Math.random()}`,
        timestamp: new Date(),
        level,
        module,
        message,
    };
}
