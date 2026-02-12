'use client';

import { useState, useEffect, useMemo } from 'react';
import {
  TrendingUp,
  TrendingDown,
  Users,
  Activity,
  DollarSign,
  ArrowUpRight,
  ArrowDownRight,
  Briefcase,
  Target,
  Globe,
  UserPlus,
  CreditCard,
  ShoppingCart,
  AlertTriangle,
  CheckCircle,
  ArrowRight
} from 'lucide-react';

// Types
interface KPIMetric {
  label: string;
  value: string;
  change: number; // percentage change vs yesterday
  icon: React.ReactNode;
  color: string;
}

interface RevenueSegment {
  label: string;
  value: number;
  color: string;
}

interface TopClient {
  id: string;
  name: string;
  volume: number;
  trades: number;
  revenue: number;
  riskScore: number; // 0-100
}

interface ActivityEvent {
  id: string;
  type: 'account' | 'deposit' | 'withdrawal' | 'trade' | 'margin_call';
  description: string;
  timestamp: Date;
  icon: React.ReactNode;
  color: string;
}

interface GeoDistribution {
  country: string;
  clientCount: number;
  percentage: number;
}

// Mock data generators
const generateKPIData = (): KPIMetric[] => {
  return [
    {
      label: 'Total Clients',
      value: '1,247',
      change: 3.2,
      icon: <Users size={20} />,
      color: 'text-blue-400',
    },
    {
      label: 'Active Traders (24h)',
      value: '423',
      change: 5.8,
      icon: <Activity size={20} />,
      color: 'text-green-400',
    },
    {
      label: 'Deposits Today',
      value: '$284,532',
      change: -2.1,
      icon: <ArrowDownRight size={20} />,
      color: 'text-emerald-400',
    },
    {
      label: 'Withdrawals Today',
      value: '$156,890',
      change: 4.3,
      icon: <ArrowUpRight size={20} />,
      color: 'text-orange-400',
    },
    {
      label: 'Net Revenue Today',
      value: '$47,234',
      change: 12.5,
      icon: <DollarSign size={20} />,
      color: 'text-yellow-400',
    },
    {
      label: 'Open Positions',
      value: '2,156',
      change: -1.4,
      icon: <Briefcase size={20} />,
      color: 'text-purple-400',
    },
    {
      label: 'Total Volume Today',
      value: '$8.2M',
      change: 8.7,
      icon: <Target size={20} />,
      color: 'text-pink-400',
    },
    {
      label: 'Platform P&L',
      value: '$23,145',
      change: 15.3,
      icon: <TrendingUp size={20} />,
      color: 'text-green-400',
    },
  ];
};

const generateRevenueData = (): RevenueSegment[] => {
  return [
    { label: 'Spread Revenue', value: 18500, color: '#3B82F6' },
    { label: 'Commission Revenue', value: 15200, color: '#10B981' },
    { label: 'Swap Revenue', value: 9800, color: '#F59E0B' },
    { label: 'Other Revenue', value: 3734, color: '#8B5CF6' },
  ];
};

const generateTopClients = (): TopClient[] => {
  const names = [
    'John Smith Trading Ltd',
    'Global Capital Partners',
    'Elite Forex Group',
    'Apex Investment Fund',
    'Quantum Trading Corp',
    'Silver Oak Capital',
    'North Star Investments',
    'Blue Horizon Trading',
    'Crown Wealth Management',
    'Phoenix Capital Group',
  ];

  return names.map((name, idx) => ({
    id: `CL-${String(idx + 1).padStart(4, '0')}`,
    name,
    volume: Math.floor(Math.random() * 500000 + 100000),
    trades: Math.floor(Math.random() * 200 + 50),
    revenue: Math.floor(Math.random() * 5000 + 1000),
    riskScore: Math.floor(Math.random() * 40 + 20), // 20-60 range
  })).sort((a, b) => b.volume - a.volume);
};

const generateRecentActivity = (): ActivityEvent[] => {
  const events: ActivityEvent[] = [];
  const now = new Date();

  const eventTypes = [
    { type: 'account' as const, descriptions: ['New account registration', 'Account verified', 'Account upgraded to VIP'], icon: <UserPlus size={14} />, color: 'text-blue-400' },
    { type: 'deposit' as const, descriptions: ['Deposit processed', 'Wire transfer received', 'Card deposit confirmed'], icon: <CreditCard size={14} />, color: 'text-green-400' },
    { type: 'withdrawal' as const, descriptions: ['Withdrawal approved', 'Withdrawal processed', 'Withdrawal request'], icon: <ArrowUpRight size={14} />, color: 'text-orange-400' },
    { type: 'trade' as const, descriptions: ['Large trade executed', 'High-volume trading session', 'Position closed'], icon: <ShoppingCart size={14} />, color: 'text-purple-400' },
    { type: 'margin_call' as const, descriptions: ['Margin call triggered', 'Stop out executed', 'Margin warning sent'], icon: <AlertTriangle size={14} />, color: 'text-red-400' },
  ];

  for (let i = 0; i < 20; i++) {
    const eventType = eventTypes[Math.floor(Math.random() * eventTypes.length)];
    const description = eventType.descriptions[Math.floor(Math.random() * eventType.descriptions.length)];
    const minutesAgo = i * 15 + Math.floor(Math.random() * 15);

    events.push({
      id: `EVT-${String(i + 1).padStart(4, '0')}`,
      type: eventType.type,
      description,
      timestamp: new Date(now.getTime() - minutesAgo * 60 * 1000),
      icon: eventType.icon,
      color: eventType.color,
    });
  }

  return events;
};

const generateGeoDistribution = (): GeoDistribution[] => {
  const countries = [
    { country: 'United States', clientCount: 342 },
    { country: 'United Kingdom', clientCount: 287 },
    { country: 'Germany', clientCount: 156 },
    { country: 'Australia', clientCount: 124 },
    { country: 'Canada', clientCount: 98 },
    { country: 'Singapore', clientCount: 87 },
    { country: 'United Arab Emirates', clientCount: 65 },
    { country: 'Japan', clientCount: 52 },
    { country: 'France', clientCount: 36 },
  ];

  const total = countries.reduce((sum, c) => sum + c.clientCount, 0);

  return countries.map(c => ({
    ...c,
    percentage: (c.clientCount / total) * 100,
  }));
};

export default function BrokerOverview() {
  const [kpiData, setKpiData] = useState<KPIMetric[]>(generateKPIData());
  const [revenueData] = useState<RevenueSegment[]>(generateRevenueData());
  const [topClients] = useState<TopClient[]>(generateTopClients());
  const [recentActivity] = useState<ActivityEvent[]>(generateRecentActivity());
  const [geoDistribution] = useState<GeoDistribution[]>(generateGeoDistribution());
  const [lastRefresh, setLastRefresh] = useState(new Date());

  // Auto-refresh every 10 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      setKpiData(generateKPIData());
      setLastRefresh(new Date());
    }, 10000);

    return () => clearInterval(interval);
  }, []);

  // Calculate donut chart paths
  const donutChart = useMemo(() => {
    const total = revenueData.reduce((sum, segment) => sum + segment.value, 0);
    const radius = 80;
    const innerRadius = 50;
    const centerX = 100;
    const centerY = 100;

    let currentAngle = -90; // Start at top

    return revenueData.map((segment) => {
      const percentage = (segment.value / total) * 100;
      const angleSize = (percentage / 100) * 360;
      const endAngle = currentAngle + angleSize;

      const startAngleRad = (currentAngle * Math.PI) / 180;
      const endAngleRad = (endAngle * Math.PI) / 180;

      const x1 = centerX + radius * Math.cos(startAngleRad);
      const y1 = centerY + radius * Math.sin(startAngleRad);
      const x2 = centerX + radius * Math.cos(endAngleRad);
      const y2 = centerY + radius * Math.sin(endAngleRad);
      const x3 = centerX + innerRadius * Math.cos(endAngleRad);
      const y3 = centerY + innerRadius * Math.sin(endAngleRad);
      const x4 = centerX + innerRadius * Math.cos(startAngleRad);
      const y4 = centerY + innerRadius * Math.sin(startAngleRad);

      const largeArc = angleSize > 180 ? 1 : 0;

      const path = `
        M ${x1} ${y1}
        A ${radius} ${radius} 0 ${largeArc} 1 ${x2} ${y2}
        L ${x3} ${y3}
        A ${innerRadius} ${innerRadius} 0 ${largeArc} 0 ${x4} ${y4}
        Z
      `;

      const result = {
        path,
        color: segment.color,
        percentage,
        label: segment.label,
        value: segment.value,
      };

      currentAngle = endAngle;
      return result;
    });
  }, [revenueData]);

  const totalRevenue = revenueData.reduce((sum, s) => sum + s.value, 0);

  const formatTime = (date: Date) => {
    const now = new Date();
    const diff = Math.floor((now.getTime() - date.getTime()) / 1000 / 60);

    if (diff < 1) return 'Just now';
    if (diff < 60) return `${diff}m ago`;
    if (diff < 1440) return `${Math.floor(diff / 60)}h ago`;
    return date.toLocaleDateString();
  };

  const getRiskColor = (score: number) => {
    if (score < 30) return 'text-green-400';
    if (score < 60) return 'text-yellow-400';
    return 'text-red-400';
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#CCC] overflow-auto">
      {/* Header */}
      <div className="flex-shrink-0 border-b border-[#383A42] bg-[#1E2026] px-4 py-3">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-lg font-bold text-white">Broker Dashboard Overview</h1>
            <p className="text-xs text-gray-400 mt-0.5">
              Real-time platform metrics • Last updated: {lastRefresh.toLocaleTimeString()}
            </p>
          </div>
          <div className="flex items-center gap-2 text-xs text-gray-400">
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
              Live • Auto-refresh: 10s
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 p-4 space-y-4">
        {/* KPI Cards - 2 rows of 4 */}
        <div className="grid grid-cols-4 gap-4">
          {kpiData.map((metric, idx) => (
            <div
              key={idx}
              className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4 hover:border-[#4A4C54] transition-colors"
            >
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-gray-400">{metric.label}</span>
                <div className={metric.color}>{metric.icon}</div>
              </div>
              <div className="text-2xl font-bold text-white mb-1">{metric.value}</div>
              <div className="flex items-center gap-1 text-xs">
                {metric.change >= 0 ? (
                  <>
                    <TrendingUp size={12} className="text-green-400" />
                    <span className="text-green-400">+{metric.change.toFixed(1)}%</span>
                  </>
                ) : (
                  <>
                    <TrendingDown size={12} className="text-red-400" />
                    <span className="text-red-400">{metric.change.toFixed(1)}%</span>
                  </>
                )}
                <span className="text-gray-500 ml-1">vs yesterday</span>
              </div>
            </div>
          ))}
        </div>

        {/* Second Row - Revenue Chart + Top Clients */}
        <div className="grid grid-cols-3 gap-4">
          {/* Revenue Breakdown Donut Chart */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <h3 className="text-sm font-semibold text-white mb-4">Revenue Breakdown</h3>
            <div className="flex items-center justify-center gap-6">
              {/* Donut Chart */}
              <div className="relative">
                <svg width="200" height="200" viewBox="0 0 200 200">
                  {donutChart.map((segment, idx) => (
                    <path
                      key={idx}
                      d={segment.path}
                      fill={segment.color}
                      className="hover:opacity-80 transition-opacity cursor-pointer"
                    />
                  ))}
                </svg>
                <div className="absolute inset-0 flex flex-col items-center justify-center">
                  <div className="text-xs text-gray-400">Total</div>
                  <div className="text-lg font-bold text-white">
                    ${(totalRevenue / 1000).toFixed(1)}K
                  </div>
                </div>
              </div>

              {/* Legend */}
              <div className="space-y-2">
                {donutChart.map((segment, idx) => (
                  <div key={idx} className="flex items-center gap-2 text-xs">
                    <div
                      className="w-3 h-3 rounded-sm"
                      style={{ backgroundColor: segment.color }}
                    ></div>
                    <div className="flex-1">
                      <div className="text-gray-300">{segment.label}</div>
                      <div className="text-gray-500">
                        ${(segment.value / 1000).toFixed(1)}K ({segment.percentage.toFixed(1)}%)
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Top 10 Clients by Volume */}
          <div className="col-span-2 bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <h3 className="text-sm font-semibold text-white mb-3">Top 10 Clients by Volume</h3>
            <div className="overflow-auto max-h-64">
              <table className="w-full text-xs">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                  <tr>
                    <th className="text-left py-2 text-gray-400 font-medium">Client</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Volume</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Trades</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Revenue</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Risk</th>
                  </tr>
                </thead>
                <tbody>
                  {topClients.map((client, idx) => (
                    <tr key={client.id} className="border-b border-[#2A2C34] hover:bg-[#25272E]">
                      <td className="py-2">
                        <div className="flex items-center gap-2">
                          <span className="text-gray-500">#{idx + 1}</span>
                          <span className="text-gray-200">{client.name}</span>
                        </div>
                      </td>
                      <td className="text-right text-white font-medium">
                        ${(client.volume / 1000).toFixed(1)}K
                      </td>
                      <td className="text-right text-gray-400">{client.trades}</td>
                      <td className="text-right text-green-400">${client.revenue.toLocaleString()}</td>
                      <td className="text-right">
                        <span className={`font-medium ${getRiskColor(client.riskScore)}`}>
                          {client.riskScore}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* Third Row - Recent Activity + Geographic Distribution */}
        <div className="grid grid-cols-3 gap-4">
          {/* Recent Activity Feed */}
          <div className="col-span-2 bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <h3 className="text-sm font-semibold text-white mb-3">Recent Activity</h3>
            <div className="space-y-2 max-h-80 overflow-auto">
              {recentActivity.map((event) => (
                <div
                  key={event.id}
                  className="flex items-center gap-3 px-3 py-2 bg-[#1A1C20] rounded hover:bg-[#25272E] transition-colors"
                >
                  <div className={event.color}>{event.icon}</div>
                  <div className="flex-1 min-w-0">
                    <div className="text-xs text-gray-200">{event.description}</div>
                    <div className="text-xs text-gray-500 mt-0.5">{formatTime(event.timestamp)}</div>
                  </div>
                  <ArrowRight size={14} className="text-gray-600 flex-shrink-0" />
                </div>
              ))}
            </div>
          </div>

          {/* Geographic Distribution */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <h3 className="text-sm font-semibold text-white mb-3 flex items-center gap-2">
              <Globe size={14} className="text-blue-400" />
              Geographic Distribution
            </h3>
            <div className="space-y-3">
              {geoDistribution.map((geo, idx) => (
                <div key={idx}>
                  <div className="flex items-center justify-between mb-1 text-xs">
                    <span className="text-gray-300">{geo.country}</span>
                    <span className="text-gray-400">{geo.clientCount}</span>
                  </div>
                  <div className="w-full h-1.5 bg-[#2A2C34] rounded-full overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-blue-500 to-blue-400 rounded-full transition-all duration-500"
                      style={{ width: `${geo.percentage}%` }}
                    ></div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
