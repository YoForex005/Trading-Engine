/**
 * LP Health Monitoring Dashboard
 * Monitor liquidity provider connections, latency, and quote rates
 */

'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
  Activity,
  AlertTriangle,
  CheckCircle,
  ChevronDown,
  ChevronUp,
  Clock,
  XCircle,
  Zap,
  TrendingUp,
  AlertCircle,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// Types
type LPStatus = 'connected' | 'disconnected' | 'degraded';

interface LPConnection {
  id: string;
  name: string;
  status: LPStatus;
  latency: number; // ms
  quotesPerSecond: number;
  lastQuoteTime: number; // timestamp
  uptime: number; // percentage
  latencyHistory: number[]; // last 20 latency readings for sparkline
  recentGaps: QuoteGap[];
  errorLog: ErrorLog[];
}

interface QuoteGap {
  startTime: number;
  endTime: number;
  duration: number; // ms
}

interface ErrorLog {
  timestamp: number;
  type: 'connection' | 'timeout' | 'data' | 'auth';
  message: string;
}

// Mock data generator
const generateMockLP = (id: string, name: string, baseLatency: number): LPConnection => {
  const now = Date.now();
  const statuses: LPStatus[] = ['connected', 'connected', 'connected', 'degraded'];
  const status = statuses[Math.floor(Math.random() * statuses.length)];

  return {
    id,
    name,
    status,
    latency: baseLatency + Math.random() * 20,
    quotesPerSecond: Math.random() * 100 + 50,
    lastQuoteTime: now - Math.random() * 5000,
    uptime: 95 + Math.random() * 5,
    latencyHistory: Array.from({ length: 20 }, () => baseLatency + Math.random() * 30),
    recentGaps: [
      {
        startTime: now - 3600000,
        endTime: now - 3598000,
        duration: 2000,
      },
    ],
    errorLog: [
      {
        timestamp: now - 7200000,
        type: 'timeout',
        message: 'Quote request timeout after 5000ms',
      },
    ],
  };
};

const MOCK_LPS: LPConnection[] = [
  generateMockLP('lp-1', 'YOFX', 45),
  generateMockLP('lp-2', 'Oanda', 65),
  generateMockLP('lp-3', 'Binance', 35),
];

export default function LPHealthDashboard() {
  const [lps, setLps] = useState<LPConnection[]>(MOCK_LPS);
  const [expandedRow, setExpandedRow] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState(Date.now());

  // Auto-refresh every 5 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      setLps((prevLps) =>
        prevLps.map((lp) => {
          const newLatency = lp.latency + (Math.random() - 0.5) * 10;
          const newHistory = [...lp.latencyHistory.slice(1), newLatency];

          // Check for disconnection and alert
          const prevStatus = lp.status;
          const newStatus =
            Math.random() > 0.95
              ? 'degraded'
              : Math.random() > 0.98
              ? 'disconnected'
              : 'connected';

          if (prevStatus === 'connected' && newStatus === 'disconnected') {
            // Alert user
            console.warn(`LP ${lp.name} disconnected!`);
            if (typeof window !== 'undefined' && 'Notification' in window) {
              new Notification('LP Disconnected', {
                body: `${lp.name} has disconnected`,
                icon: '/alert.png',
              });
            }
          }

          return {
            ...lp,
            status: newStatus,
            latency: Math.max(10, newLatency),
            quotesPerSecond: Math.random() * 100 + 50,
            lastQuoteTime: Date.now() - Math.random() * 5000,
            latencyHistory: newHistory,
          };
        })
      );
      setLastUpdate(Date.now());
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  // Summary stats
  const stats = useMemo(() => {
    const connectedCount = lps.filter((lp) => lp.status === 'connected').length;
    const avgLatency =
      lps.reduce((sum, lp) => sum + lp.latency, 0) / lps.length;
    const totalQuotes = lps.reduce((sum, lp) => sum + lp.quotesPerSecond, 0);

    return {
      connectedCount,
      avgLatency,
      totalQuotes,
    };
  }, [lps]);

  // Get latency color
  const getLatencyColor = (latency: number): string => {
    if (latency < 50) return 'text-emerald-400';
    if (latency < 100) return 'text-yellow-400';
    return 'text-rose-400';
  };

  // Get status badge
  const getStatusBadge = (status: LPStatus) => {
    switch (status) {
      case 'connected':
        return (
          <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 font-semibold text-xs flex items-center gap-1">
            <CheckCircle size={12} />
            Connected
          </span>
        );
      case 'degraded':
        return (
          <span className="px-2 py-0.5 rounded bg-yellow-500/20 text-yellow-400 font-semibold text-xs flex items-center gap-1">
            <AlertTriangle size={12} />
            Degraded
          </span>
        );
      case 'disconnected':
        return (
          <span className="px-2 py-0.5 rounded bg-rose-500/20 text-rose-400 font-semibold text-xs flex items-center gap-1">
            <XCircle size={12} />
            Disconnected
          </span>
        );
    }
  };

  // Format time ago
  const timeAgo = (timestamp: number): string => {
    const seconds = Math.floor((Date.now() - timestamp) / 1000);
    if (seconds < 5) return 'just now';
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    return `${hours}h ago`;
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-zinc-300 p-4 overflow-auto custom-scrollbar">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-zinc-200">LP Health Monitoring</h1>
          <p className="text-sm text-zinc-500 mt-1">
            Real-time monitoring of liquidity provider connections
          </p>
        </div>
        <div className="text-xs text-zinc-500">
          Last updated: {timeAgo(lastUpdate)}
        </div>
      </div>

      {/* Status Cards */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        {/* Connected LPs */}
        <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-emerald-500/20 flex items-center justify-center">
              <Activity size={20} className="text-emerald-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-emerald-400">
                {stats.connectedCount}/{lps.length}
              </div>
              <div className="text-xs text-zinc-500">Connected LPs</div>
            </div>
          </div>
        </div>

        {/* Avg Latency */}
        <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-blue-500/20 flex items-center justify-center">
              <Zap size={20} className="text-blue-400" />
            </div>
            <div>
              <div className={`text-2xl font-bold ${getLatencyColor(stats.avgLatency)}`}>
                {stats.avgLatency.toFixed(0)}ms
              </div>
              <div className="text-xs text-zinc-500">Avg Latency</div>
            </div>
          </div>
        </div>

        {/* Total Quote Rate */}
        <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-purple-500/20 flex items-center justify-center">
              <TrendingUp size={20} className="text-purple-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-purple-400">
                {stats.totalQuotes.toFixed(0)}
              </div>
              <div className="text-xs text-zinc-500">Quotes/sec</div>
            </div>
          </div>
        </div>
      </div>

      {/* LP Table */}
      <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded overflow-hidden flex-1">
        <table className="w-full text-xs">
          <thead className="bg-[#252525] border-b border-zinc-700">
            <tr>
              <th className="px-4 py-3 text-left text-zinc-500 font-semibold w-8"></th>
              <th className="px-4 py-3 text-left text-zinc-500 font-semibold">LP Name</th>
              <th className="px-4 py-3 text-center text-zinc-500 font-semibold">Status</th>
              <th className="px-4 py-3 text-center text-zinc-500 font-semibold">Latency</th>
              <th className="px-4 py-3 text-center text-zinc-500 font-semibold">Quotes/sec</th>
              <th className="px-4 py-3 text-left text-zinc-500 font-semibold">Last Quote</th>
              <th className="px-4 py-3 text-center text-zinc-500 font-semibold">Uptime</th>
              <th className="px-4 py-3 text-left text-zinc-500 font-semibold">Latency Trend</th>
            </tr>
          </thead>
          <tbody>
            {lps.map((lp) => (
              <React.Fragment key={lp.id}>
                {/* Main Row */}
                <tr
                  className="border-b border-zinc-700/30 hover:bg-zinc-800/30 cursor-pointer"
                  onClick={() => setExpandedRow(expandedRow === lp.id ? null : lp.id)}
                >
                  <td className="px-4 py-3">
                    {expandedRow === lp.id ? (
                      <ChevronUp size={14} className="text-zinc-500" />
                    ) : (
                      <ChevronDown size={14} className="text-zinc-500" />
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <div className="font-semibold text-zinc-200">{lp.name}</div>
                  </td>
                  <td className="px-4 py-3 text-center">{getStatusBadge(lp.status)}</td>
                  <td className="px-4 py-3 text-center">
                    <span className={`font-mono font-semibold ${getLatencyColor(lp.latency)}`}>
                      {lp.latency.toFixed(0)}ms
                    </span>
                  </td>
                  <td className="px-4 py-3 text-center">
                    <span className="font-mono text-zinc-300">{lp.quotesPerSecond.toFixed(1)}</span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1 text-zinc-400">
                      <Clock size={12} />
                      <span>{timeAgo(lp.lastQuoteTime)}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-center">
                    <span className="font-mono text-emerald-400">{lp.uptime.toFixed(2)}%</span>
                  </td>
                  <td className="px-4 py-3">
                    <LatencySparkline data={lp.latencyHistory} />
                  </td>
                </tr>

                {/* Expanded Details */}
                {expandedRow === lp.id && (
                  <tr className="bg-[#252525] border-b border-zinc-700">
                    <td colSpan={8} className="px-4 py-4">
                      <div className="grid grid-cols-3 gap-4">
                        {/* Latency Histogram */}
                        <div>
                          <h4 className="text-xs font-bold text-zinc-400 mb-3">Latency Distribution</h4>
                          <LatencyHistogram data={lp.latencyHistory} />
                        </div>

                        {/* Recent Quote Gaps */}
                        <div>
                          <h4 className="text-xs font-bold text-zinc-400 mb-3">Recent Quote Gaps</h4>
                          <div className="space-y-2">
                            {lp.recentGaps.length === 0 ? (
                              <div className="text-xs text-zinc-600">No gaps detected</div>
                            ) : (
                              lp.recentGaps.map((gap, idx) => (
                                <div
                                  key={idx}
                                  className="px-2 py-1.5 bg-zinc-800 rounded text-xs"
                                >
                                  <div className="flex items-center gap-1 text-yellow-400 mb-1">
                                    <AlertCircle size={10} />
                                    <span className="font-semibold">{gap.duration}ms gap</span>
                                  </div>
                                  <div className="text-zinc-500">
                                    {new Date(gap.startTime).toLocaleTimeString()}
                                  </div>
                                </div>
                              ))
                            )}
                          </div>
                        </div>

                        {/* Error Log */}
                        <div>
                          <h4 className="text-xs font-bold text-zinc-400 mb-3">Error Log</h4>
                          <div className="space-y-2 max-h-32 overflow-y-auto custom-scrollbar">
                            {lp.errorLog.length === 0 ? (
                              <div className="text-xs text-zinc-600">No errors</div>
                            ) : (
                              lp.errorLog.map((error, idx) => (
                                <div
                                  key={idx}
                                  className="px-2 py-1.5 bg-zinc-800 rounded text-xs"
                                >
                                  <div className="flex items-center gap-1 text-rose-400 mb-1">
                                    <XCircle size={10} />
                                    <span className="font-semibold capitalize">{error.type}</span>
                                  </div>
                                  <div className="text-zinc-400 text-[10px] leading-relaxed">
                                    {error.message}
                                  </div>
                                  <div className="text-zinc-600 text-[10px] mt-1">
                                    {new Date(error.timestamp).toLocaleString()}
                                  </div>
                                </div>
                              ))
                            )}
                          </div>
                        </div>
                      </div>
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// Latency Sparkline Component
interface LatencySparklineProps {
  data: number[];
}

function LatencySparkline({ data }: LatencySparklineProps) {
  const maxValue = Math.max(...data);
  const minValue = Math.min(...data);
  const range = maxValue - minValue || 1;

  // Create SVG path
  const points = data.map((value, index) => {
    const x = (index / (data.length - 1)) * 100;
    const y = 100 - ((value - minValue) / range) * 100;
    return `${x},${y}`;
  });

  const pathD = `M ${points.join(' L ')}`;

  // Determine color based on latest value
  const latestValue = data[data.length - 1];
  const color = latestValue < 50 ? '#10b981' : latestValue < 100 ? '#eab308' : '#f43f5e';

  return (
    <div className="w-24 h-8">
      <svg width="100%" height="100%" viewBox="0 0 100 100" preserveAspectRatio="none">
        <path
          d={pathD}
          fill="none"
          stroke={color}
          strokeWidth="3"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
    </div>
  );
}

// Latency Histogram Component
interface LatencyHistogramProps {
  data: number[];
}

function LatencyHistogram({ data }: LatencyHistogramProps) {
  // Create histogram bins
  const bins = [
    { label: '0-50ms', count: 0, color: 'bg-emerald-500' },
    { label: '50-100ms', count: 0, color: 'bg-yellow-500' },
    { label: '100+ms', count: 0, color: 'bg-rose-500' },
  ];

  data.forEach((value) => {
    if (value < 50) bins[0].count++;
    else if (value < 100) bins[1].count++;
    else bins[2].count++;
  });

  const maxCount = Math.max(...bins.map((b) => b.count));

  return (
    <div className="space-y-2">
      {bins.map((bin, idx) => (
        <div key={idx} className="flex items-center gap-2">
          <div className="text-xs text-zinc-500 w-20">{bin.label}</div>
          <div className="flex-1 bg-zinc-800 rounded h-4 overflow-hidden">
            <div
              className={`h-full ${bin.color} transition-all`}
              style={{ width: `${(bin.count / maxCount) * 100}%` }}
            />
          </div>
          <div className="text-xs text-zinc-400 font-mono w-8 text-right">{bin.count}</div>
        </div>
      ))}
    </div>
  );
}
