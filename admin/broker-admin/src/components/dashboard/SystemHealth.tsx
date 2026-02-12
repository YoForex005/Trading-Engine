'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { API_CONFIG } from '@/config/api';
import {
  Activity,
  Cpu,
  HardDrive,
  Network,
  TrendingUp,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Database,
  Wifi,
  Server,
  RefreshCw,
  Clock,
} from 'lucide-react';

// ============================================================================
// TYPES
// ============================================================================

interface SystemMetrics {
  cpu: number;
  memory: number;
  disk: number;
  activeConnections: number;
  requestsPerMin: number;
  errorRate: number;
}

interface ServiceStatus {
  name: string;
  status: 'online' | 'offline' | 'degraded';
  latency: number; // ms
  uptime: number; // percentage
  icon: React.ReactNode;
}

type AlertSeverity = 'CRITICAL' | 'WARNING' | 'INFO';

interface SystemAlert {
  id: string;
  timestamp: number;
  severity: AlertSeverity;
  message: string;
  service: string;
}

interface ChartDataPoint {
  time: number;
  cpu: number;
  memory: number;
  requests: number;
}

// ============================================================================
// MOCK DATA GENERATORS
// ============================================================================

const generateMetrics = (): SystemMetrics => ({
  cpu: parseFloat((Math.random() * 30 + 40).toFixed(1)), // 40-70%
  memory: parseFloat((Math.random() * 20 + 60).toFixed(1)), // 60-80%
  disk: parseFloat((Math.random() * 10 + 45).toFixed(1)), // 45-55%
  activeConnections: Math.floor(Math.random() * 200 + 800), // 800-1000
  requestsPerMin: Math.floor(Math.random() * 500 + 1500), // 1500-2000
  errorRate: parseFloat((Math.random() * 0.5).toFixed(2)), // 0-0.5%
});

const generateChartData = (hours: number): ChartDataPoint[] => {
  const now = Date.now();
  const points: ChartDataPoint[] = [];

  for (let i = hours * 12; i >= 0; i--) {
    const time = now - i * 5 * 60 * 1000; // 5-minute intervals
    points.push({
      time,
      cpu: parseFloat((Math.random() * 30 + 40).toFixed(1)),
      memory: parseFloat((Math.random() * 20 + 60).toFixed(1)),
      requests: Math.floor(Math.random() * 500 + 1500),
    });
  }

  return points;
};

const SERVICES: Omit<ServiceStatus, 'status' | 'latency' | 'uptime'>[] = [
  { name: 'Backend API', icon: <Server size={14} /> },
  { name: 'WebSocket Server', icon: <Wifi size={14} /> },
  { name: 'FIX Gateway', icon: <Network size={14} /> },
  { name: 'Database', icon: <Database size={14} /> },
  { name: 'Redis Cache', icon: <Activity size={14} /> },
  { name: 'LP Connection', icon: <TrendingUp size={14} /> },
];

const generateServiceStatus = (): ServiceStatus[] => {
  return SERVICES.map((service) => {
    const rand = Math.random();
    let status: 'online' | 'offline' | 'degraded';
    let latency: number;
    let uptime: number;

    if (rand > 0.95) {
      status = 'offline';
      latency = 0;
      uptime = parseFloat((Math.random() * 5 + 90).toFixed(2));
    } else if (rand > 0.85) {
      status = 'degraded';
      latency = Math.floor(Math.random() * 200 + 100);
      uptime = parseFloat((Math.random() * 5 + 95).toFixed(2));
    } else {
      status = 'online';
      latency = Math.floor(Math.random() * 50 + 10);
      uptime = parseFloat((Math.random() * 1 + 99).toFixed(2));
    }

    return {
      ...service,
      status,
      latency,
      uptime,
    };
  });
};

const ALERT_MESSAGES = [
  { severity: 'CRITICAL' as AlertSeverity, message: 'CPU usage exceeded 90%', service: 'Backend API' },
  { severity: 'CRITICAL' as AlertSeverity, message: 'Database connection lost', service: 'Database' },
  { severity: 'WARNING' as AlertSeverity, message: 'High memory usage detected (85%)', service: 'Backend API' },
  { severity: 'WARNING' as AlertSeverity, message: 'WebSocket connection drops detected', service: 'WebSocket Server' },
  { severity: 'WARNING' as AlertSeverity, message: 'Error rate spike: 2.5%', service: 'Backend API' },
  { severity: 'INFO' as AlertSeverity, message: 'System restarted successfully', service: 'Backend API' },
  { severity: 'INFO' as AlertSeverity, message: 'Cache cleared', service: 'Redis Cache' },
  { severity: 'WARNING' as AlertSeverity, message: 'FIX Gateway latency increased to 250ms', service: 'FIX Gateway' },
  { severity: 'CRITICAL' as AlertSeverity, message: 'LP Connection timeout', service: 'LP Connection' },
  { severity: 'INFO' as AlertSeverity, message: 'Backup completed', service: 'Database' },
];

const generateAlerts = (count: number): SystemAlert[] => {
  const now = Date.now();
  return Array.from({ length: count }, (_, i) => {
    const template = ALERT_MESSAGES[Math.floor(Math.random() * ALERT_MESSAGES.length)];
    return {
      id: `alert-${now}-${i}`,
      timestamp: now - Math.floor(Math.random() * 3600000), // last hour
      severity: template.severity,
      message: template.message,
      service: template.service,
    };
  }).sort((a, b) => b.timestamp - a.timestamp);
};

// ============================================================================
// CHART COMPONENT
// ============================================================================

interface AreaChartProps {
  data: ChartDataPoint[];
  width: number;
  height: number;
}

const AreaChart: React.FC<AreaChartProps> = ({ data, width, height }) => {
  const padding = { top: 10, right: 10, bottom: 20, left: 40 };
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;

  const maxCpu = Math.max(...data.map((d) => d.cpu));
  const maxMemory = Math.max(...data.map((d) => d.memory));
  const maxRequests = Math.max(...data.map((d) => d.requests));
  const maxValue = Math.max(100, maxRequests / 20); // Scale requests down

  const getX = (index: number) => (index / (data.length - 1)) * chartWidth;
  const getY = (value: number) => chartHeight - (value / maxValue) * chartHeight;

  const createPath = (dataKey: keyof ChartDataPoint, scale: number = 1) => {
    if (data.length === 0) return '';

    const points = data.map((d, i) => ({
      x: getX(i),
      y: getY((d[dataKey] as number) * scale),
    }));

    const pathData = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ');
    return pathData;
  };

  const createAreaPath = (dataKey: keyof ChartDataPoint, scale: number = 1) => {
    if (data.length === 0) return '';

    const linePath = createPath(dataKey, scale);
    const lastX = getX(data.length - 1);
    const firstX = getX(0);
    return `${linePath} L ${lastX} ${chartHeight} L ${firstX} ${chartHeight} Z`;
  };

  return (
    <svg width={width} height={height}>
      <g transform={`translate(${padding.left}, ${padding.top})`}>
        {/* Grid lines */}
        {[0, 25, 50, 75, 100].map((value) => {
          const y = getY(value);
          return (
            <g key={value}>
              <line
                x1={0}
                y1={y}
                x2={chartWidth}
                y2={y}
                stroke="#3f3f46"
                strokeWidth={1}
                strokeDasharray="2,2"
              />
              <text x={-5} y={y + 3} fontSize={10} fill="#71717a" textAnchor="end">
                {value}
              </text>
            </g>
          );
        })}

        {/* Area fills */}
        <path
          d={createAreaPath('cpu')}
          fill="url(#cpuGradient)"
          opacity={0.3}
        />
        <path
          d={createAreaPath('memory')}
          fill="url(#memoryGradient)"
          opacity={0.3}
        />
        <path
          d={createAreaPath('requests', 1 / 20)}
          fill="url(#requestsGradient)"
          opacity={0.3}
        />

        {/* Lines */}
        <path
          d={createPath('cpu')}
          fill="none"
          stroke="#3b82f6"
          strokeWidth={2}
        />
        <path
          d={createPath('memory')}
          fill="none"
          stroke="#10b981"
          strokeWidth={2}
        />
        <path
          d={createPath('requests', 1 / 20)}
          fill="none"
          stroke="#f59e0b"
          strokeWidth={2}
        />
      </g>

      {/* Gradients */}
      <defs>
        <linearGradient id="cpuGradient" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="#3b82f6" stopOpacity={0.8} />
          <stop offset="100%" stopColor="#3b82f6" stopOpacity={0.1} />
        </linearGradient>
        <linearGradient id="memoryGradient" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="#10b981" stopOpacity={0.8} />
          <stop offset="100%" stopColor="#10b981" stopOpacity={0.1} />
        </linearGradient>
        <linearGradient id="requestsGradient" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="#f59e0b" stopOpacity={0.8} />
          <stop offset="100%" stopColor="#f59e0b" stopOpacity={0.1} />
        </linearGradient>
      </defs>
    </svg>
  );
};

// ============================================================================
// MAIN COMPONENT
// ============================================================================

export default function SystemHealth() {
  const [metrics, setMetrics] = useState<SystemMetrics>(generateMetrics());
  const [services, setServices] = useState<ServiceStatus[]>(generateServiceStatus());
  const [alerts, setAlerts] = useState<SystemAlert[]>(generateAlerts(20));
  const [chartData, setChartData] = useState<ChartDataPoint[]>(generateChartData(24));
  const [countdown, setCountdown] = useState(5);

  // Auto-refresh every 5 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      setMetrics(generateMetrics());
      setServices(generateServiceStatus());
      setChartData((prev) => {
        const newPoint: ChartDataPoint = {
          time: Date.now(),
          cpu: parseFloat((Math.random() * 30 + 40).toFixed(1)),
          memory: parseFloat((Math.random() * 20 + 60).toFixed(1)),
          requests: Math.floor(Math.random() * 500 + 1500),
        };
        return [...prev.slice(1), newPoint];
      });
      setCountdown(5);
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  // Countdown timer
  useEffect(() => {
    const interval = setInterval(() => {
      setCountdown((prev) => (prev > 0 ? prev - 1 : 5));
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  const getSeverityColor = (severity: AlertSeverity) => {
    switch (severity) {
      case 'CRITICAL':
        return 'bg-rose-500/20 text-rose-400 border-rose-500/50';
      case 'WARNING':
        return 'bg-yellow-500/20 text-yellow-400 border-yellow-500/50';
      case 'INFO':
        return 'bg-blue-500/20 text-blue-400 border-blue-500/50';
    }
  };

  const getStatusColor = (status: 'online' | 'offline' | 'degraded') => {
    switch (status) {
      case 'online':
        return 'bg-emerald-500';
      case 'degraded':
        return 'bg-yellow-500';
      case 'offline':
        return 'bg-rose-500';
    }
  };

  const getMetricColor = (value: number, type: 'cpu' | 'memory' | 'disk' | 'error') => {
    if (type === 'error') {
      if (value > 1) return 'text-rose-400';
      if (value > 0.5) return 'text-yellow-400';
      return 'text-emerald-400';
    }

    if (value > 80) return 'text-rose-400';
    if (value > 60) return 'text-yellow-400';
    return 'text-emerald-400';
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-zinc-300 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-[#1E2026] border-b border-zinc-700">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-bold text-zinc-100">System Health Monitor</h1>
          <div className="flex items-center gap-2 text-xs text-zinc-500">
            <RefreshCw size={12} className="animate-spin" />
            <span>Auto-refresh in {countdown}s</span>
          </div>
        </div>
        <div className="text-xs text-zinc-500">
          Last updated: {new Date().toLocaleTimeString()}
        </div>
      </div>

      {/* Main Content - Scrollable */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Metric Cards */}
        <div className="grid grid-cols-6 gap-3">
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <Cpu size={14} className="text-blue-400" />
              <div className="text-xs text-zinc-500">CPU Usage</div>
            </div>
            <div className={`text-2xl font-bold ${getMetricColor(metrics.cpu, 'cpu')}`}>
              {metrics.cpu}%
            </div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <Activity size={14} className="text-emerald-400" />
              <div className="text-xs text-zinc-500">Memory Usage</div>
            </div>
            <div className={`text-2xl font-bold ${getMetricColor(metrics.memory, 'memory')}`}>
              {metrics.memory}%
            </div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <HardDrive size={14} className="text-purple-400" />
              <div className="text-xs text-zinc-500">Disk Usage</div>
            </div>
            <div className={`text-2xl font-bold ${getMetricColor(metrics.disk, 'disk')}`}>
              {metrics.disk}%
            </div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <Network size={14} className="text-cyan-400" />
              <div className="text-xs text-zinc-500">Active Connections</div>
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {metrics.activeConnections}
            </div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp size={14} className="text-yellow-400" />
              <div className="text-xs text-zinc-500">Requests/min</div>
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {metrics.requestsPerMin.toLocaleString()}
            </div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <AlertTriangle size={14} className="text-rose-400" />
              <div className="text-xs text-zinc-500">Error Rate</div>
            </div>
            <div className={`text-2xl font-bold ${getMetricColor(metrics.errorRate, 'error')}`}>
              {metrics.errorRate}%
            </div>
          </div>
        </div>

        {/* Performance Chart */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="px-4 py-3 bg-zinc-800/50 border-b border-zinc-700">
            <h2 className="text-sm font-bold text-zinc-100">Performance Trends (24h)</h2>
            <div className="flex items-center gap-4 mt-2 text-xs">
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 bg-blue-500 rounded"></div>
                <span className="text-zinc-400">CPU %</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 bg-emerald-500 rounded"></div>
                <span className="text-zinc-400">Memory %</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 bg-yellow-500 rounded"></div>
                <span className="text-zinc-400">Requests (÷20)</span>
              </div>
            </div>
          </div>
          <div className="p-4">
            <AreaChart data={chartData} width={1200} height={300} />
          </div>
        </div>

        {/* Service Status Grid */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="px-4 py-3 bg-zinc-800/50 border-b border-zinc-700">
            <h2 className="text-sm font-bold text-zinc-100">Service Status</h2>
          </div>
          <div className="p-4">
            <div className="grid grid-cols-3 gap-3">
              {services.map((service) => (
                <div
                  key={service.name}
                  className="bg-zinc-800/50 border border-zinc-700 rounded p-3"
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <div className={`w-2 h-2 rounded-full ${getStatusColor(service.status)}`}></div>
                      <div className="text-xs font-bold text-zinc-300">{service.name}</div>
                    </div>
                    <div className="text-zinc-500">{service.icon}</div>
                  </div>
                  <div className="grid grid-cols-2 gap-2 text-xs">
                    <div>
                      <div className="text-zinc-500">Latency</div>
                      <div className={`font-bold ${service.latency > 100 ? 'text-yellow-400' : 'text-emerald-400'}`}>
                        {service.status === 'offline' ? 'N/A' : `${service.latency}ms`}
                      </div>
                    </div>
                    <div>
                      <div className="text-zinc-500">Uptime</div>
                      <div className={`font-bold ${service.uptime < 95 ? 'text-rose-400' : 'text-emerald-400'}`}>
                        {service.uptime}%
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Alert Log */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="px-4 py-3 bg-zinc-800/50 border-b border-zinc-700">
            <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
              <AlertTriangle size={14} className="text-yellow-400" />
              System Alerts (Last 20)
            </h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-zinc-800/30 border-b border-zinc-700">
                <tr>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Time</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Severity</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Service</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Message</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-700/50">
                {alerts.map((alert) => (
                  <tr key={alert.id} className="hover:bg-zinc-800/30">
                    <td className="px-3 py-2 text-zinc-500">
                      {new Date(alert.timestamp).toLocaleTimeString()}
                    </td>
                    <td className="px-3 py-2">
                      <span className={`px-2 py-1 text-xs font-bold rounded border ${getSeverityColor(alert.severity)}`}>
                        {alert.severity}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-zinc-400">{alert.service}</td>
                    <td className="px-3 py-2 text-zinc-300">{alert.message}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
