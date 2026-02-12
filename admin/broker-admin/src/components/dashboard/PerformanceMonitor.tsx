'use client';

import { useState, useEffect, useCallback } from 'react';
import { Activity, Server, Zap, Clock, TrendingUp, TrendingDown, AlertTriangle, CheckCircle, XCircle, RefreshCw } from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// ============================================================================
// TYPE DEFINITIONS
// ============================================================================

interface SystemMetrics {
  cpu: number; // percentage 0-100
  memory: number; // percentage 0-100
  goroutines: number;
  gcPause: number; // milliseconds
  rps: number; // requests per second
  avgLatency: number; // milliseconds
  timestamp: string;
}

interface HistoricalDataPoint {
  timestamp: string;
  cpu: number;
  memory: number;
  latency: number;
}

interface EndpointPerformance {
  endpoint: string;
  method: string;
  p50: number; // milliseconds
  p99: number; // milliseconds
  errorRate: number; // percentage
  throughput: number; // requests/minute
}

interface ServiceHealth {
  name: string;
  status: 'healthy' | 'degraded' | 'unhealthy';
  uptime: number; // seconds
  responseTime: number; // milliseconds
  lastCheck: string;
}

interface SlowQuery {
  endpoint: string;
  method: string;
  avgLatency: number;
  callCount: number;
}

interface ErrorRatePoint {
  timestamp: string;
  errorRate: number;
}

interface ConnectionMetrics {
  timestamp: string;
  httpConnections: number;
  wsConnections: number;
}

type SortField = 'endpoint' | 'p99' | 'errorRate' | 'throughput';
type SortDirection = 'asc' | 'desc';
type RefreshInterval = 5 | 10 | 30 | 60; // seconds

// ============================================================================
// MAIN COMPONENT
// ============================================================================

export default function PerformanceMonitor() {
  const [loading, setLoading] = useState(true);
  const [systemMetrics, setSystemMetrics] = useState<SystemMetrics | null>(null);
  const [historicalData, setHistoricalData] = useState<HistoricalDataPoint[]>([]);
  const [endpoints, setEndpoints] = useState<EndpointPerformance[]>([]);
  const [services, setServices] = useState<ServiceHealth[]>([]);
  const [slowQueries, setSlowQueries] = useState<SlowQuery[]>([]);
  const [errorRateData, setErrorRateData] = useState<ErrorRatePoint[]>([]);
  const [connectionData, setConnectionData] = useState<ConnectionMetrics[]>([]);

  // UI state
  const [sortField, setSortField] = useState<SortField>('p99');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
  const [visibleMetrics, setVisibleMetrics] = useState({ cpu: true, memory: true, latency: true });
  const [refreshInterval, setRefreshInterval] = useState<RefreshInterval>(10);
  const [autoRefresh, setAutoRefresh] = useState(true);

  // ============================================================================
  // DATA FETCHING
  // ============================================================================

  const fetchAllData = useCallback(async () => {
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('rtx_admin_token') : null;
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const [metricsRes, historyRes, endpointsRes, servicesRes, slowRes, errorsRes, connectionsRes] = await Promise.all([
        fetch(API_CONFIG.PERFORMANCE_METRICS, { headers }),
        fetch(API_CONFIG.PERFORMANCE_HISTORY, { headers }),
        fetch(API_CONFIG.PERFORMANCE_ENDPOINTS, { headers }),
        fetch(API_CONFIG.PERFORMANCE_SERVICES, { headers }),
        fetch(API_CONFIG.PERFORMANCE_SLOW_QUERIES, { headers }),
        fetch(API_CONFIG.PERFORMANCE_ERROR_RATE, { headers }),
        fetch(API_CONFIG.PERFORMANCE_CONNECTIONS, { headers }),
      ]);

      if (metricsRes.ok) setSystemMetrics(await metricsRes.json());
      if (historyRes.ok) setHistoricalData(await historyRes.json());
      if (endpointsRes.ok) setEndpoints(await endpointsRes.json());
      if (servicesRes.ok) setServices(await servicesRes.json());
      if (slowRes.ok) setSlowQueries(await slowRes.json());
      if (errorsRes.ok) setErrorRateData(await errorsRes.json());
      if (connectionsRes.ok) setConnectionData(await connectionsRes.json());

      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch performance data:', error);
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAllData();
  }, [fetchAllData]);

  // Auto-refresh effect
  useEffect(() => {
    if (!autoRefresh) return;

    const intervalId = setInterval(() => {
      fetchAllData();
    }, refreshInterval * 1000);

    return () => clearInterval(intervalId);
  }, [autoRefresh, refreshInterval, fetchAllData]);

  // ============================================================================
  // SORTING & FILTERING
  // ============================================================================

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const sortedEndpoints = [...endpoints].sort((a, b) => {
    let aVal: number, bVal: number;
    switch (sortField) {
      case 'p99':
        aVal = a.p99;
        bVal = b.p99;
        break;
      case 'errorRate':
        aVal = a.errorRate;
        bVal = b.errorRate;
        break;
      case 'throughput':
        aVal = a.throughput;
        bVal = b.throughput;
        break;
      default:
        return 0;
    }
    return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
  });

  // ============================================================================
  // HELPER FUNCTIONS
  // ============================================================================

  const getMetricColor = (value: number, threshold1: number, threshold2: number) => {
    if (value >= threshold2) return '#E74C3C'; // Red
    if (value >= threshold1) return '#F39C12'; // Orange
    return '#2ECC71'; // Green
  };

  const getServiceStatusColor = (status: ServiceHealth['status']) => {
    switch (status) {
      case 'healthy': return '#2ECC71';
      case 'degraded': return '#F39C12';
      case 'unhealthy': return '#E74C3C';
    }
  };

  const getServiceStatusIcon = (status: ServiceHealth['status']) => {
    switch (status) {
      case 'healthy': return <CheckCircle size={16} />;
      case 'degraded': return <AlertTriangle size={16} />;
      case 'unhealthy': return <XCircle size={16} />;
    }
  };

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${mins}m`;
    return `${mins}m`;
  };

  const formatNumber = (num: number, decimals = 0) => {
    return num.toFixed(decimals);
  };

  // ============================================================================
  // SVG CHART RENDERERS
  // ============================================================================

  // Circular gauge for CPU/Memory
  const renderCircularGauge = (value: number, label: string, color: string) => {
    const radius = 40;
    const circumference = 2 * Math.PI * radius;
    const offset = circumference - (value / 100) * circumference;

    return (
      <div className="flex flex-col items-center">
        <svg width="120" height="120" className="transform -rotate-90">
          {/* Background circle */}
          <circle
            cx="60"
            cy="60"
            r={radius}
            fill="none"
            stroke="#383A42"
            strokeWidth="8"
          />
          {/* Progress circle */}
          <circle
            cx="60"
            cy="60"
            r={radius}
            fill="none"
            stroke={color}
            strokeWidth="8"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            strokeLinecap="round"
          />
          {/* Center text */}
          <text
            x="60"
            y="60"
            textAnchor="middle"
            dy="0.3em"
            fill="#FFF"
            fontSize="20"
            fontWeight="bold"
            transform="rotate(90 60 60)"
          >
            {value.toFixed(0)}%
          </text>
        </svg>
        <div className="text-xs text-[#888] mt-2">{label}</div>
      </div>
    );
  };

  // Multi-line chart for historical metrics
  const renderHistoricalChart = () => {
    if (historicalData.length === 0) return null;

    const chartWidth = 800;
    const chartHeight = 200;
    const padding = { top: 20, right: 20, bottom: 30, left: 50 };
    const plotWidth = chartWidth - padding.left - padding.right;
    const plotHeight = chartHeight - padding.top - padding.bottom;

    // Scales
    const maxCpu = Math.max(...historicalData.map(d => d.cpu), 100);
    const maxMemory = Math.max(...historicalData.map(d => d.memory), 100);
    const maxLatency = Math.max(...historicalData.map(d => d.latency), 1);

    const scaleX = (index: number) => padding.left + (index / (historicalData.length - 1)) * plotWidth;
    const scaleCpu = (val: number) => padding.top + plotHeight - (val / maxCpu) * plotHeight;
    const scaleMemory = (val: number) => padding.top + plotHeight - (val / maxMemory) * plotHeight;
    const scaleLatency = (val: number) => padding.top + plotHeight - (val / maxLatency) * plotHeight;

    // Generate path strings
    const cpuPath = historicalData.map((d, i) => `${i === 0 ? 'M' : 'L'} ${scaleX(i)} ${scaleCpu(d.cpu)}`).join(' ');
    const memoryPath = historicalData.map((d, i) => `${i === 0 ? 'M' : 'L'} ${scaleX(i)} ${scaleMemory(d.memory)}`).join(' ');
    const latencyPath = historicalData.map((d, i) => `${i === 0 ? 'M' : 'L'} ${scaleX(i)} ${scaleLatency(d.latency)}`).join(' ');

    return (
      <svg width={chartWidth} height={chartHeight}>
        {/* Grid lines */}
        {[0, 25, 50, 75, 100].map(val => {
          const y = padding.top + plotHeight - (val / 100) * plotHeight;
          return (
            <g key={val}>
              <line x1={padding.left} y1={y} x2={chartWidth - padding.right} y2={y} stroke="#383A42" strokeWidth="1" />
              <text x={padding.left - 5} y={y} textAnchor="end" fill="#888" fontSize="10" dy="0.3em">{val}</text>
            </g>
          );
        })}

        {/* CPU line */}
        {visibleMetrics.cpu && <path d={cpuPath} fill="none" stroke="#3498DB" strokeWidth="2" />}

        {/* Memory line */}
        {visibleMetrics.memory && <path d={memoryPath} fill="none" stroke="#9B59B6" strokeWidth="2" />}

        {/* Latency line */}
        {visibleMetrics.latency && <path d={latencyPath} fill="none" stroke="#F39C12" strokeWidth="2" />}
      </svg>
    );
  };

  // Area chart for error rate
  const renderErrorRateChart = () => {
    if (errorRateData.length === 0) return null;

    const chartWidth = 600;
    const chartHeight = 150;
    const padding = { top: 20, right: 20, bottom: 30, left: 50 };
    const plotWidth = chartWidth - padding.left - padding.right;
    const plotHeight = chartHeight - padding.top - padding.bottom;

    const maxRate = Math.max(...errorRateData.map(d => d.errorRate), 5);
    const threshold = 2; // 2% error rate threshold

    const scaleX = (index: number) => padding.left + (index / (errorRateData.length - 1)) * plotWidth;
    const scaleY = (val: number) => padding.top + plotHeight - (val / maxRate) * plotHeight;

    const linePath = errorRateData.map((d, i) => `${i === 0 ? 'M' : 'L'} ${scaleX(i)} ${scaleY(d.errorRate)}`).join(' ');
    const areaPath = `${linePath} L ${scaleX(errorRateData.length - 1)} ${padding.top + plotHeight} L ${padding.left} ${padding.top + plotHeight} Z`;

    const thresholdY = scaleY(threshold);

    return (
      <svg width={chartWidth} height={chartHeight}>
        {/* Threshold line */}
        <line x1={padding.left} y1={thresholdY} x2={chartWidth - padding.right} y2={thresholdY} stroke="#E74C3C" strokeWidth="1" strokeDasharray="4 2" />

        {/* Area fill */}
        <path d={areaPath} fill="#E74C3C" fillOpacity="0.2" />

        {/* Line */}
        <path d={linePath} fill="none" stroke="#E74C3C" strokeWidth="2" />
      </svg>
    );
  };

  // Dual line chart for connections
  const renderConnectionChart = () => {
    if (connectionData.length === 0) return null;

    const chartWidth = 600;
    const chartHeight = 150;
    const padding = { top: 20, right: 20, bottom: 30, left: 50 };
    const plotWidth = chartWidth - padding.left - padding.right;
    const plotHeight = chartHeight - padding.top - padding.bottom;

    const maxHttp = Math.max(...connectionData.map(d => d.httpConnections), 1);
    const maxWs = Math.max(...connectionData.map(d => d.wsConnections), 1);
    const maxConn = Math.max(maxHttp, maxWs);

    const scaleX = (index: number) => padding.left + (index / (connectionData.length - 1)) * plotWidth;
    const scaleY = (val: number) => padding.top + plotHeight - (val / maxConn) * plotHeight;

    const httpPath = connectionData.map((d, i) => `${i === 0 ? 'M' : 'L'} ${scaleX(i)} ${scaleY(d.httpConnections)}`).join(' ');
    const wsPath = connectionData.map((d, i) => `${i === 0 ? 'M' : 'L'} ${scaleX(i)} ${scaleY(d.wsConnections)}`).join(' ');

    return (
      <svg width={chartWidth} height={chartHeight}>
        {/* HTTP connections line */}
        <path d={httpPath} fill="none" stroke="#3498DB" strokeWidth="2" />

        {/* WebSocket connections line */}
        <path d={wsPath} fill="none" stroke="#2ECC71" strokeWidth="2" />
      </svg>
    );
  };

  // ============================================================================
  // RENDER
  // ============================================================================

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-[#888]">Loading performance data...</div>
      </div>
    );
  }

  return (
    <div className="h-full overflow-auto bg-[#121316] p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-bold text-[#F5C542]">Performance Monitoring</h2>

        {/* Auto-refresh controls */}
        <div className="flex items-center gap-3">
          <button
            onClick={() => setAutoRefresh(!autoRefresh)}
            className={`px-3 py-1 rounded text-xs font-bold flex items-center gap-1 ${
              autoRefresh ? 'bg-[#2ECC71] text-white' : 'bg-[#383A42] text-[#888]'
            }`}
          >
            <RefreshCw size={12} className={autoRefresh ? 'animate-spin' : ''} />
            Auto-refresh {autoRefresh ? 'ON' : 'OFF'}
          </button>
          <select
            value={refreshInterval}
            onChange={(e) => setRefreshInterval(Number(e.target.value) as RefreshInterval)}
            className="px-2 py-1 bg-[#1E2026] text-[#CCC] text-xs rounded border border-[#383A42]"
            disabled={!autoRefresh}
          >
            <option value={5}>5s</option>
            <option value={10}>10s</option>
            <option value={30}>30s</option>
            <option value={60}>1m</option>
          </select>
          <button
            onClick={fetchAllData}
            className="px-3 py-1 bg-[#3498DB] text-white rounded text-xs font-bold"
          >
            Refresh Now
          </button>
        </div>
      </div>

      {/* 1. System Metrics Cards */}
      <div className="grid grid-cols-6 gap-4 mb-4">
        <div className="col-span-1 bg-[#1E2026] border border-[#383A42] rounded p-4 flex items-center justify-center">
          {systemMetrics && renderCircularGauge(systemMetrics.cpu, 'CPU', getMetricColor(systemMetrics.cpu, 70, 85))}
        </div>
        <div className="col-span-1 bg-[#1E2026] border border-[#383A42] rounded p-4 flex items-center justify-center">
          {systemMetrics && renderCircularGauge(systemMetrics.memory, 'Memory', getMetricColor(systemMetrics.memory, 70, 85))}
        </div>
        <div className="col-span-1 bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="text-[#888] text-xs mb-1">Goroutines</div>
          <div className="text-2xl font-bold text-white">{systemMetrics?.goroutines || 0}</div>
          <Activity size={16} className="text-[#3498DB] mt-2" />
        </div>
        <div className="col-span-1 bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="text-[#888] text-xs mb-1">GC Pause</div>
          <div className="text-2xl font-bold text-white">{formatNumber(systemMetrics?.gcPause || 0, 1)} ms</div>
          <Zap size={16} className="text-[#F39C12] mt-2" />
        </div>
        <div className="col-span-1 bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="text-[#888] text-xs mb-1">RPS</div>
          <div className="text-2xl font-bold text-white">{formatNumber(systemMetrics?.rps || 0)}</div>
          <TrendingUp size={16} className="text-[#2ECC71] mt-2" />
        </div>
        <div className="col-span-1 bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="text-[#888] text-xs mb-1">Avg Latency</div>
          <div className="text-2xl font-bold text-white">{formatNumber(systemMetrics?.avgLatency || 0, 1)} ms</div>
          <Clock size={16} className="text-[#9B59B6] mt-2" />
        </div>
      </div>

      {/* 2. Historical Metrics Chart */}
      <div className="bg-[#1E2026] border border-[#383A42] rounded p-4 mb-4">
        <div className="flex justify-between items-center mb-3">
          <h3 className="text-sm font-bold text-[#F5C542]">Historical Metrics (60 min)</h3>
          <div className="flex gap-3 text-xs">
            <label className="flex items-center gap-1 cursor-pointer">
              <input
                type="checkbox"
                checked={visibleMetrics.cpu}
                onChange={() => setVisibleMetrics({ ...visibleMetrics, cpu: !visibleMetrics.cpu })}
              />
              <span style={{ color: '#3498DB' }}>CPU</span>
            </label>
            <label className="flex items-center gap-1 cursor-pointer">
              <input
                type="checkbox"
                checked={visibleMetrics.memory}
                onChange={() => setVisibleMetrics({ ...visibleMetrics, memory: !visibleMetrics.memory })}
              />
              <span style={{ color: '#9B59B6' }}>Memory</span>
            </label>
            <label className="flex items-center gap-1 cursor-pointer">
              <input
                type="checkbox"
                checked={visibleMetrics.latency}
                onChange={() => setVisibleMetrics({ ...visibleMetrics, latency: !visibleMetrics.latency })}
              />
              <span style={{ color: '#F39C12' }}>Latency</span>
            </label>
          </div>
        </div>
        {renderHistoricalChart()}
      </div>

      <div className="grid grid-cols-2 gap-4 mb-4">
        {/* 3. Endpoint Performance Table */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <h3 className="text-sm font-bold text-[#F5C542] mb-3">Endpoint Performance</h3>
          <div className="overflow-auto max-h-80">
            <table className="w-full text-xs">
              <thead className="bg-[#25272E] sticky top-0">
                <tr>
                  <th className="text-left px-2 py-2 text-[#888]">Endpoint</th>
                  <th className="text-right px-2 py-2 text-[#888] cursor-pointer" onClick={() => handleSort('p99')}>
                    P99 {sortField === 'p99' && (sortDirection === 'asc' ? '▲' : '▼')}
                  </th>
                  <th className="text-right px-2 py-2 text-[#888] cursor-pointer" onClick={() => handleSort('errorRate')}>
                    Error % {sortField === 'errorRate' && (sortDirection === 'asc' ? '▲' : '▼')}
                  </th>
                  <th className="text-right px-2 py-2 text-[#888] cursor-pointer" onClick={() => handleSort('throughput')}>
                    Throughput {sortField === 'throughput' && (sortDirection === 'asc' ? '▲' : '▼')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {sortedEndpoints.map((ep, idx) => (
                  <tr key={idx} className="border-b border-[#383A42]">
                    <td className="px-2 py-2 text-[#CCC]">
                      <span className="text-[#F39C12] mr-1">{ep.method}</span>
                      {ep.endpoint}
                    </td>
                    <td
                      className="px-2 py-2 text-right font-mono"
                      style={{ color: ep.p99 > 500 ? '#E74C3C' : ep.p99 > 200 ? '#F39C12' : '#2ECC71' }}
                    >
                      {formatNumber(ep.p99, 1)}ms
                    </td>
                    <td
                      className="px-2 py-2 text-right font-mono"
                      style={{ color: ep.errorRate > 5 ? '#E74C3C' : ep.errorRate > 2 ? '#F39C12' : '#2ECC71' }}
                    >
                      {formatNumber(ep.errorRate, 2)}%
                    </td>
                    <td className="px-2 py-2 text-right text-[#CCC] font-mono">{formatNumber(ep.throughput, 1)}/m</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* 4. Service Health Grid */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <h3 className="text-sm font-bold text-[#F5C542] mb-3">Service Health</h3>
          <div className="grid grid-cols-2 gap-2">
            {services.map((svc, idx) => (
              <div key={idx} className="bg-[#25272E] border border-[#383A42] rounded p-3">
                <div className="flex items-center justify-between mb-2">
                  <div className="text-xs font-bold text-white">{svc.name}</div>
                  <div className="flex items-center gap-1" style={{ color: getServiceStatusColor(svc.status) }}>
                    {getServiceStatusIcon(svc.status)}
                  </div>
                </div>
                <div className="text-xs text-[#888]">Uptime: {formatUptime(svc.uptime)}</div>
                <div className="text-xs text-[#888]">Response: {formatNumber(svc.responseTime, 1)}ms</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-4">
        {/* 5. Slow Queries Panel */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <h3 className="text-sm font-bold text-[#F5C542] mb-3">Slowest Endpoints</h3>
          <div className="space-y-2">
            {slowQueries.slice(0, 10).map((query, idx) => {
              const maxLatency = Math.max(...slowQueries.map(q => q.avgLatency));
              const barWidth = (query.avgLatency / maxLatency) * 100;
              return (
                <div key={idx}>
                  <div className="flex justify-between text-xs mb-1">
                    <div className="text-[#CCC]">
                      <span className="text-[#F39C12] mr-1">{query.method}</span>
                      {query.endpoint}
                    </div>
                    <div className="text-[#888] font-mono">{formatNumber(query.avgLatency, 1)}ms</div>
                  </div>
                  <div className="h-1.5 bg-[#383A42] rounded overflow-hidden">
                    <div
                      className="h-full rounded"
                      style={{
                        width: `${barWidth}%`,
                        backgroundColor: query.avgLatency > 500 ? '#E74C3C' : query.avgLatency > 200 ? '#F39C12' : '#3498DB',
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* 6. Error Rate Trend */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <h3 className="text-sm font-bold text-[#F5C542] mb-3">Error Rate Trend (Last Hour)</h3>
          {renderErrorRateChart()}
          <div className="text-xs text-[#888] mt-2">Threshold: 2% (dashed line)</div>
        </div>
      </div>

      {/* 7. Connection Monitor */}
      <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
        <h3 className="text-sm font-bold text-[#F5C542] mb-3">Connection Monitor</h3>
        {renderConnectionChart()}
        <div className="flex gap-4 text-xs mt-2">
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded" style={{ backgroundColor: '#3498DB' }} />
            <span className="text-[#888]">HTTP Connections</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded" style={{ backgroundColor: '#2ECC71' }} />
            <span className="text-[#888]">WebSocket Connections</span>
          </div>
        </div>
      </div>
    </div>
  );
}
