'use client';

import { useState, useEffect } from 'react';
import {
  AlertTriangle,
  Shield,
  TrendingUp,
  TrendingDown,
  Activity,
  Target,
  Zap,
  CheckCircle,
  XCircle,
  AlertCircle,
  ChevronUp,
  ChevronDown
} from 'lucide-react';
import { API_CONFIG } from '../../config/api';

// ===== TYPES =====
interface RiskOverview {
  totalExposure: number;
  marginUtilization: number;
  netPosition: number;
  largestSingleExposure: number;
  riskScore: number; // 1-100
}

interface SymbolGroupExposure {
  group: string;
  longExposure: number;
  shortExposure: number;
  netExposure: number;
  count: number;
}

interface RiskyAccount {
  accountId: string;
  name: string;
  marginLevel: number; // percentage
  exposure: number;
  unrealizedPnL: number;
  riskScore: number;
}

interface RiskAlert {
  id: string;
  type: 'circuit_breaker' | 'margin_warning' | 'concentration';
  severity: 'critical' | 'high' | 'medium' | 'low';
  message: string;
  timestamp: string;
  accountId?: string;
  symbol?: string;
}

interface ExposureHeatmapCell {
  symbol: string;
  accountId: string;
  exposure: number;
  intensity: number; // 0-1 normalized
}

interface RiskTrendPoint {
  date: string;
  riskScore: number;
}

interface CircuitBreaker {
  id: string;
  name: string;
  type: string;
  status: 'armed' | 'triggered' | 'ok';
  threshold: number;
  current: number;
  lastTriggered?: string;
}

interface RiskLimit {
  name: string;
  limit: number;
  current: number;
  unit: string;
  percentage: number;
}

// ===== COMPONENT =====
export default function RiskDashboard() {
  const [overview, setOverview] = useState<RiskOverview | null>(null);
  const [symbolGroupExposure, setSymbolGroupExposure] = useState<SymbolGroupExposure[]>([]);
  const [riskyAccounts, setRiskyAccounts] = useState<RiskyAccount[]>([]);
  const [alerts, setAlerts] = useState<RiskAlert[]>([]);
  const [heatmapData, setHeatmapData] = useState<ExposureHeatmapCell[]>([]);
  const [trendData, setTrendData] = useState<RiskTrendPoint[]>([]);
  const [circuitBreakers, setCircuitBreakers] = useState<CircuitBreaker[]>([]);
  const [limits, setLimits] = useState<RiskLimit[]>([]);
  const [sortField, setSortField] = useState<keyof RiskyAccount>('riskScore');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
  const [loading, setLoading] = useState(true);

  // ===== DATA FETCHING =====
  useEffect(() => {
    fetchAllRiskData();
    const interval = setInterval(fetchAllRiskData, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, []);

  const fetchAllRiskData = async () => {
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('rtx_admin_token') : null;
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const [
        overviewRes,
        exposureRes,
        accountsRes,
        alertsRes,
        heatmapRes,
        trendRes,
        breakersRes,
        limitsRes
      ] = await Promise.all([
        fetch(API_CONFIG.RISK_OVERVIEW, { headers }),
        fetch(API_CONFIG.RISK_EXPOSURE_BY_GROUP, { headers }),
        fetch(API_CONFIG.RISK_RISKY_ACCOUNTS, { headers }),
        fetch(API_CONFIG.RISK_ALERTS, { headers }),
        fetch(API_CONFIG.RISK_HEATMAP, { headers }),
        fetch(API_CONFIG.RISK_TREND, { headers }),
        fetch(API_CONFIG.RISK_CIRCUIT_BREAKERS, { headers }),
        fetch(API_CONFIG.RISK_LIMITS, { headers })
      ]);

      if (overviewRes.ok) setOverview(await overviewRes.json());
      if (exposureRes.ok) setSymbolGroupExposure(await exposureRes.json());
      if (accountsRes.ok) setRiskyAccounts(await accountsRes.json());
      if (alertsRes.ok) setAlerts(await alertsRes.json());
      if (heatmapRes.ok) setHeatmapData(await heatmapRes.json());
      if (trendRes.ok) setTrendData(await trendRes.json());
      if (breakersRes.ok) setCircuitBreakers(await breakersRes.json());
      if (limitsRes.ok) setLimits(await limitsRes.json());

      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch risk data:', error);
      setLoading(false);
    }
  };
  // ===== SORTING =====
  const handleSort = (field: keyof RiskyAccount) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const sortedAccounts = [...riskyAccounts].sort((a, b) => {
    const aVal = a[sortField];
    const bVal = b[sortField];
    const multiplier = sortDirection === 'asc' ? 1 : -1;
    if (typeof aVal === 'string' && typeof bVal === 'string') {
      return aVal.localeCompare(bVal) * multiplier;
    }
    return ((aVal as number) - (bVal as number)) * multiplier;
  });

  // ===== HELPER FUNCTIONS =====
  const getRiskScoreColor = (score: number): string => {
    if (score >= 80) return 'text-red-400';
    if (score >= 60) return 'text-orange-400';
    if (score >= 40) return 'text-yellow-400';
    return 'text-green-400';
  };

  const getSeverityColor = (severity: string): string => {
    switch (severity) {
      case 'critical': return 'bg-red-500/20 text-red-400 border-red-500';
      case 'high': return 'bg-orange-500/20 text-orange-400 border-orange-500';
      case 'medium': return 'bg-yellow-500/20 text-yellow-400 border-yellow-500';
      case 'low': return 'bg-blue-500/20 text-blue-400 border-blue-500';
      default: return 'bg-zinc-700 text-zinc-300';
    }
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'triggered': return 'text-red-400';
      case 'armed': return 'text-yellow-400';
      case 'ok': return 'text-green-400';
      default: return 'text-zinc-400';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'triggered': return <XCircle size={16} />;
      case 'armed': return <AlertCircle size={16} />;
      case 'ok': return <CheckCircle size={16} />;
      default: return <Activity size={16} />;
    }
  };

  const formatCurrency = (val: number): string => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val);
  };

  const formatPercent = (val: number): string => {
    return `${val.toFixed(2)}%`;
  };

  // ===== SVG CHARTS =====
  const renderExposureBarChart = () => {
    if (symbolGroupExposure.length === 0) return null;

    const maxExposure = Math.max(
      ...symbolGroupExposure.map(g => Math.max(Math.abs(g.longExposure), Math.abs(g.shortExposure)))
    );

    const barHeight = 32;
    const barSpacing = 48;
    const chartHeight = symbolGroupExposure.length * barSpacing + 20;
    const chartWidth = 600;
    const labelWidth = 100;
    const scaleWidth = chartWidth - labelWidth - 60;

    return (
      <svg width={chartWidth} height={chartHeight} className="text-xs">
        {symbolGroupExposure.map((group, idx) => {
          const y = idx * barSpacing + 10;
          const longWidth = (Math.abs(group.longExposure) / maxExposure) * (scaleWidth / 2);
          const shortWidth = (Math.abs(group.shortExposure) / maxExposure) * (scaleWidth / 2);
          const centerX = labelWidth + scaleWidth / 2;

          return (
            <g key={group.group}>
              <text x={10} y={y + 16} fill="#888" className="text-xs font-medium">
                {group.group}
              </text>
              {/* Short (left) */}
              <rect
                x={centerX - shortWidth}
                y={y}
                width={shortWidth}
                height={barHeight}
                fill="#ef4444"
                opacity={0.7}
              />
              {/* Long (right) */}
              <rect
                x={centerX}
                y={y}
                width={longWidth}
                height={barHeight}
                fill="#22c55e"
                opacity={0.7}
              />
              {/* Center line */}
              <line
                x1={centerX}
                y1={y}
                x2={centerX}
                y2={y + barHeight}
                stroke="#555"
                strokeWidth={1}
              />
              {/* Values */}
              <text x={centerX - shortWidth - 5} y={y + 20} fill="#ef4444" textAnchor="end" className="text-xs">
                {formatCurrency(group.shortExposure)}
              </text>
              <text x={centerX + longWidth + 5} y={y + 20} fill="#22c55e" textAnchor="start" className="text-xs">
                {formatCurrency(group.longExposure)}
              </text>
            </g>
          );
        })}
      </svg>
    );
  };

  const renderTrendLineChart = () => {
    if (trendData.length === 0) return null;

    const chartWidth = 800;
    const chartHeight = 200;
    const padding = 40;
    const innerWidth = chartWidth - 2 * padding;
    const innerHeight = chartHeight - 2 * padding;

    const minScore = Math.min(...trendData.map(p => p.riskScore));
    const maxScore = Math.max(...trendData.map(p => p.riskScore));
    const range = maxScore - minScore || 1;

    const points = trendData.map((point, idx) => {
      const x = padding + (idx / (trendData.length - 1)) * innerWidth;
      const y = padding + innerHeight - ((point.riskScore - minScore) / range) * innerHeight;
      return { x, y, score: point.riskScore, date: point.date };
    });

    const pathD = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ');

    return (
      <svg width={chartWidth} height={chartHeight} className="text-xs">
        {/* Grid lines */}
        {[0, 25, 50, 75, 100].map(score => {
          const y = padding + innerHeight - ((score - minScore) / range) * innerHeight;
          return (
            <g key={score}>
              <line x1={padding} y1={y} x2={chartWidth - padding} y2={y} stroke="#333" strokeWidth={1} />
              <text x={padding - 10} y={y + 4} fill="#666" textAnchor="end" className="text-xs">
                {score}
              </text>
            </g>
          );
        })}
        {/* Line path */}
        <path d={pathD} fill="none" stroke="#F5C542" strokeWidth={2} />
        {/* Data points */}
        {points.map((p, idx) => (
          <circle key={idx} cx={p.x} cy={p.y} r={4} fill="#F5C542" />
        ))}
        {/* X-axis labels (show every 5th date) */}
        {points.map((p, idx) => {
          if (idx % 5 === 0 || idx === points.length - 1) {
            return (
              <text key={idx} x={p.x} y={chartHeight - 10} fill="#666" textAnchor="middle" className="text-xs">
                {new Date(p.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
              </text>
            );
          }
          return null;
        })}
      </svg>
    );
  };

  // ===== LOADING STATE =====
  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-zinc-400">Loading risk dashboard...</div>
      </div>
    );
  }

  // ===== RENDER =====
  return (
    <div className="h-full overflow-auto bg-[#121316] p-4">
      <div className="max-w-[1800px] mx-auto space-y-4">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-xl font-bold text-zinc-100">Risk Dashboard</h1>
          <div className="text-xs text-zinc-500">
            Last updated: {new Date().toLocaleTimeString()}
          </div>
        </div>

        {/* 1. Risk Overview Cards */}
        <div className="grid grid-cols-5 gap-4">
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-zinc-400">Total Exposure</span>
              <Target size={16} className="text-blue-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {overview ? formatCurrency(overview.totalExposure) : '$0'}
            </div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-zinc-400">Margin Utilization</span>
              <Activity size={16} className="text-purple-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {overview ? formatPercent(overview.marginUtilization) : '0%'}
            </div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-zinc-400">Net Position</span>
              {overview && overview.netPosition >= 0 ? (
                <TrendingUp size={16} className="text-green-400" />
              ) : (
                <TrendingDown size={16} className="text-red-400" />
              )}
            </div>
            <div className={`text-2xl font-bold ${overview && overview.netPosition >= 0 ? 'text-green-400' : 'text-red-400'}`}>
              {overview ? formatCurrency(overview.netPosition) : '$0'}
            </div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-zinc-400">Largest Exposure</span>
              <AlertTriangle size={16} className="text-orange-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {overview ? formatCurrency(overview.largestSingleExposure) : '$0'}
            </div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-zinc-400">Risk Score</span>
              <Shield size={16} className={overview ? getRiskScoreColor(overview.riskScore) : 'text-zinc-400'} />
            </div>
            <div className={`text-2xl font-bold ${overview ? getRiskScoreColor(overview.riskScore) : 'text-zinc-100'}`}>
              {overview ? overview.riskScore : '0'} / 100
            </div>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          {/* 2. Exposure by Symbol Group */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <h2 className="text-sm font-bold text-zinc-100 mb-4">Exposure by Symbol Group</h2>
            <div className="overflow-auto">
              {renderExposureBarChart()}
            </div>
            <div className="flex items-center justify-center gap-6 mt-4 text-xs">
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-green-500 opacity-70 rounded"></div>
                <span className="text-zinc-400">Long</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-red-500 opacity-70 rounded"></div>
                <span className="text-zinc-400">Short</span>
              </div>
            </div>
          </div>

          {/* 4. Risk Alerts Panel */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <h2 className="text-sm font-bold text-zinc-100 mb-4">Risk Alerts</h2>
            <div className="space-y-2 max-h-[400px] overflow-auto">
              {alerts.length === 0 ? (
                <div className="text-center text-zinc-500 py-8">No active alerts</div>
              ) : (
                alerts.map(alert => (
                  <div
                    key={alert.id}
                    className={`border rounded p-3 ${getSeverityColor(alert.severity)}`}
                  >
                    <div className="flex items-start justify-between mb-1">
                      <div className="flex items-center gap-2">
                        <AlertTriangle size={14} />
                        <span className="text-xs font-bold uppercase">{alert.type.replace('_', ' ')}</span>
                      </div>
                      <span className="text-xs opacity-70">
                        {new Date(alert.timestamp).toLocaleTimeString()}
                      </span>
                    </div>
                    <div className="text-xs">{alert.message}</div>
                    {(alert.accountId || alert.symbol) && (
                      <div className="text-xs opacity-70 mt-1">
                        {alert.accountId && `Account: ${alert.accountId}`}
                        {alert.symbol && ` | Symbol: ${alert.symbol}`}
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* 3. Top Risky Accounts Table */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <h2 className="text-sm font-bold text-zinc-100 mb-4">Top Risky Accounts</h2>
          <div className="overflow-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-[#383A42]">
                  <th
                    className="text-left py-2 px-2 text-zinc-400 cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('accountId')}
                  >
                    <div className="flex items-center gap-1">
                      Account ID
                      {sortField === 'accountId' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th
                    className="text-left py-2 px-2 text-zinc-400 cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('name')}
                  >
                    <div className="flex items-center gap-1">
                      Name
                      {sortField === 'name' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th
                    className="text-right py-2 px-2 text-zinc-400 cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('marginLevel')}
                  >
                    <div className="flex items-center justify-end gap-1">
                      Margin Level
                      {sortField === 'marginLevel' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th
                    className="text-right py-2 px-2 text-zinc-400 cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('exposure')}
                  >
                    <div className="flex items-center justify-end gap-1">
                      Exposure
                      {sortField === 'exposure' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th
                    className="text-right py-2 px-2 text-zinc-400 cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('unrealizedPnL')}
                  >
                    <div className="flex items-center justify-end gap-1">
                      Unrealized P&L
                      {sortField === 'unrealizedPnL' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th
                    className="text-right py-2 px-2 text-zinc-400 cursor-pointer hover:text-zinc-200"
                    onClick={() => handleSort('riskScore')}
                  >
                    <div className="flex items-center justify-end gap-1">
                      Risk Score
                      {sortField === 'riskScore' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                </tr>
              </thead>
              <tbody>
                {sortedAccounts.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="text-center py-8 text-zinc-500">
                      No risky accounts detected
                    </td>
                  </tr>
                ) : (
                  sortedAccounts.map(account => (
                    <tr key={account.accountId} className="border-b border-[#383A42]/50 hover:bg-[#25272E]">
                      <td className="py-2 px-2 text-zinc-300 font-mono">{account.accountId}</td>
                      <td className="py-2 px-2 text-zinc-300">{account.name}</td>
                      <td className={`py-2 px-2 text-right font-mono ${account.marginLevel < 100 ? 'text-red-400' : 'text-zinc-300'}`}>
                        {formatPercent(account.marginLevel)}
                      </td>
                      <td className="py-2 px-2 text-right font-mono text-zinc-300">
                        {formatCurrency(account.exposure)}
                      </td>
                      <td className={`py-2 px-2 text-right font-mono ${account.unrealizedPnL >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                        {formatCurrency(account.unrealizedPnL)}
                      </td>
                      <td className={`py-2 px-2 text-right font-bold ${getRiskScoreColor(account.riskScore)}`}>
                        {account.riskScore}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* 5. Exposure Heatmap */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <h2 className="text-sm font-bold text-zinc-100 mb-4">Exposure Heatmap (Symbol × Account)</h2>
          <div className="overflow-auto">
            {heatmapData.length === 0 ? (
              <div className="text-center text-zinc-500 py-8">No heatmap data available</div>
            ) : (
              <div className="grid gap-1" style={{ gridTemplateColumns: `repeat(${Math.ceil(Math.sqrt(heatmapData.length))}, 1fr)` }}>
                {heatmapData.slice(0, 100).map((cell, idx) => {
                  const hue = cell.intensity > 0 ? 120 : 0; // Green for positive, red for negative
                  const saturation = Math.abs(cell.intensity) * 100;
                  const lightness = 30 + (1 - Math.abs(cell.intensity)) * 20;
                  return (
                    <div
                      key={idx}
                      className="relative group w-12 h-12 rounded cursor-pointer border border-[#383A42]"
                      style={{ backgroundColor: `hsl(${hue}, ${saturation}%, ${lightness}%)` }}
                      title={`${cell.symbol} - ${cell.accountId}: ${formatCurrency(cell.exposure)}`}
                    >
                      <div className="absolute hidden group-hover:block bg-zinc-900 text-white text-xs p-2 rounded shadow-lg z-10 -top-16 left-1/2 -translate-x-1/2 whitespace-nowrap">
                        <div className="font-bold">{cell.symbol}</div>
                        <div>Acct: {cell.accountId}</div>
                        <div>{formatCurrency(cell.exposure)}</div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>

        {/* 6. Historical Risk Trend */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <h2 className="text-sm font-bold text-zinc-100 mb-4">Risk Score Trend (30 Days)</h2>
          <div className="overflow-auto">
            {renderTrendLineChart()}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          {/* 7. Circuit Breaker Status */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <h2 className="text-sm font-bold text-zinc-100 mb-4">Circuit Breaker Status</h2>
            <div className="space-y-2">
              {circuitBreakers.length === 0 ? (
                <div className="text-center text-zinc-500 py-8">No circuit breakers configured</div>
              ) : (
                circuitBreakers.map(breaker => (
                  <div key={breaker.id} className="flex items-center justify-between border border-[#383A42] rounded p-3">
                    <div className="flex items-center gap-3">
                      <div className={getStatusColor(breaker.status)}>
                        {getStatusIcon(breaker.status)}
                      </div>
                      <div>
                        <div className="text-sm font-medium text-zinc-200">{breaker.name}</div>
                        <div className="text-xs text-zinc-500">{breaker.type}</div>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-xs text-zinc-400">
                        {breaker.current} / {breaker.threshold}
                      </div>
                      {breaker.lastTriggered && (
                        <div className="text-xs text-zinc-600">
                          Last: {new Date(breaker.lastTriggered).toLocaleDateString()}
                        </div>
                      )}
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* 8. Risk Limits Table */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <h2 className="text-sm font-bold text-zinc-100 mb-4">Risk Limits</h2>
            <div className="space-y-3">
              {limits.length === 0 ? (
                <div className="text-center text-zinc-500 py-8">No limits configured</div>
              ) : (
                limits.map((limit, idx) => (
                  <div key={idx}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-xs text-zinc-300">{limit.name}</span>
                      <span className="text-xs text-zinc-500">
                        {limit.current.toLocaleString()} / {limit.limit.toLocaleString()} {limit.unit}
                      </span>
                    </div>
                    <div className="relative h-2 bg-[#121316] rounded overflow-hidden">
                      <div
                        className={`absolute top-0 left-0 h-full transition-all ${
                          limit.percentage >= 90
                            ? 'bg-red-500'
                            : limit.percentage >= 75
                            ? 'bg-orange-500'
                            : limit.percentage >= 50
                            ? 'bg-yellow-500'
                            : 'bg-green-500'
                        }`}
                        style={{ width: `${Math.min(limit.percentage, 100)}%` }}
                      />
                    </div>
                    <div className="text-xs text-zinc-600 mt-1">{formatPercent(limit.percentage)}</div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
