'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
  LogIn,
  LogOut,
  TrendingUp,
  TrendingDown,
  DollarSign,
  CreditCard,
  Lock,
  Settings,
  FileCheck,
  Key,
  Users,
  Shield,
  Globe,
  Activity,
  AlertTriangle,
  Search,
  Filter,
  Calendar,
} from 'lucide-react';

type EventType =
  | 'login'
  | 'logout'
  | 'trade_opened'
  | 'trade_closed'
  | 'deposit'
  | 'withdrawal'
  | 'password_change'
  | 'settings_change'
  | 'kyc_submitted'
  | 'api_key_created';

interface ActivityEntry {
  id: string;
  timestamp: Date;
  clientName: string;
  clientId: string;
  eventType: EventType;
  ipAddress: string;
  device: string;
  browser: string;
  details: string;
  country: string;
}

interface SuspiciousActivity {
  id: string;
  type: 'failed_logins' | 'new_ip' | 'concurrent_sessions';
  clientName: string;
  description: string;
  timestamp: Date;
  severity: 'low' | 'medium' | 'high';
}

const EVENT_TYPE_CONFIG: Record<
  EventType,
  { icon: React.ReactNode; label: string; color: string; bgColor: string }
> = {
  login: {
    icon: <LogIn size={12} />,
    label: 'Login',
    color: '#22C55E',
    bgColor: '#22C55E20',
  },
  logout: {
    icon: <LogOut size={12} />,
    label: 'Logout',
    color: '#888',
    bgColor: '#88888820',
  },
  trade_opened: {
    icon: <TrendingUp size={12} />,
    label: 'Trade Opened',
    color: '#3B82F6',
    bgColor: '#3B82F620',
  },
  trade_closed: {
    icon: <TrendingDown size={12} />,
    label: 'Trade Closed',
    color: '#9B59B6',
    bgColor: '#9B59B620',
  },
  deposit: {
    icon: <DollarSign size={12} />,
    label: 'Deposit',
    color: '#10B981',
    bgColor: '#10B98120',
  },
  withdrawal: {
    icon: <CreditCard size={12} />,
    label: 'Withdrawal',
    color: '#F59E0B',
    bgColor: '#F59E0B20',
  },
  password_change: {
    icon: <Lock size={12} />,
    label: 'Password Change',
    color: '#EF4444',
    bgColor: '#EF444420',
  },
  settings_change: {
    icon: <Settings size={12} />,
    label: 'Settings Change',
    color: '#8B5CF6',
    bgColor: '#8B5CF620',
  },
  kyc_submitted: {
    icon: <FileCheck size={12} />,
    label: 'KYC Submitted',
    color: '#F5C542',
    bgColor: '#F5C54220',
  },
  api_key_created: {
    icon: <Key size={12} />,
    label: 'API Key Created',
    color: '#EC4899',
    bgColor: '#EC489920',
  },
};

const COUNTRIES = [
  'United States',
  'United Kingdom',
  'Germany',
  'France',
  'Japan',
  'Australia',
  'Canada',
  'Singapore',
  'UAE',
  'Hong Kong',
  'India',
  'Brazil',
  'Russia',
  'South Korea',
];

const DEVICES = ['Desktop', 'Mobile', 'Tablet'];
const BROWSERS = ['Chrome', 'Firefox', 'Safari', 'Edge', 'Opera'];

// Mock data generator
const generateMockActivity = (): ActivityEntry[] => {
  const eventTypes: EventType[] = [
    'login',
    'logout',
    'trade_opened',
    'trade_closed',
    'deposit',
    'withdrawal',
    'password_change',
    'settings_change',
    'kyc_submitted',
    'api_key_created',
  ];

  return Array.from({ length: 100 }, (_, i) => {
    const eventType = eventTypes[Math.floor(Math.random() * eventTypes.length)];
    const country = COUNTRIES[Math.floor(Math.random() * COUNTRIES.length)];

    let details = '';
    switch (eventType) {
      case 'login':
        details = 'Successful login';
        break;
      case 'logout':
        details = 'User logged out';
        break;
      case 'trade_opened':
        details = `Opened EURUSD ${Math.random() > 0.5 ? 'BUY' : 'SELL'} 0.${Math.floor(
          Math.random() * 99
        )} lots`;
        break;
      case 'trade_closed':
        details = `Closed position #${1000 + i}, P&L: ${
          Math.random() > 0.5 ? '+' : '-'
        }$${(Math.random() * 500).toFixed(2)}`;
        break;
      case 'deposit':
        details = `Deposited $${(500 + Math.random() * 4500).toFixed(2)} via Credit Card`;
        break;
      case 'withdrawal':
        details = `Withdrew $${(100 + Math.random() * 1900).toFixed(2)} to bank account`;
        break;
      case 'password_change':
        details = 'Changed account password';
        break;
      case 'settings_change':
        details = 'Updated notification preferences';
        break;
      case 'kyc_submitted':
        details = 'Submitted verification documents';
        break;
      case 'api_key_created':
        details = 'Created new API key';
        break;
    }

    return {
      id: `ACT${String(10000 + i).padStart(5, '0')}`,
      timestamp: new Date(Date.now() - Math.random() * 86400000), // Last 24 hours
      clientName: `Client ${Math.floor(Math.random() * 500) + 1}`,
      clientId: `AC${String(680962851 + Math.floor(Math.random() * 1000)).padStart(9, '0')}`,
      eventType,
      ipAddress: `${Math.floor(Math.random() * 256)}.${Math.floor(
        Math.random() * 256
      )}.${Math.floor(Math.random() * 256)}.${Math.floor(Math.random() * 256)}`,
      device: DEVICES[Math.floor(Math.random() * DEVICES.length)],
      browser: BROWSERS[Math.floor(Math.random() * BROWSERS.length)],
      details,
      country,
    };
  });
};

const generateSuspiciousActivity = (): SuspiciousActivity[] => {
  return [
    {
      id: 'SA001',
      type: 'failed_logins',
      clientName: 'Client 42',
      description: '5 failed login attempts in 10 minutes',
      timestamp: new Date(Date.now() - 300000),
      severity: 'high',
    },
    {
      id: 'SA002',
      type: 'new_ip',
      clientName: 'Client 127',
      description: 'Login from new IP address (Russia)',
      timestamp: new Date(Date.now() - 600000),
      severity: 'medium',
    },
    {
      id: 'SA003',
      type: 'concurrent_sessions',
      clientName: 'Client 89',
      description: '3 concurrent sessions from different countries',
      timestamp: new Date(Date.now() - 900000),
      severity: 'high',
    },
    {
      id: 'SA004',
      type: 'failed_logins',
      clientName: 'Client 234',
      description: '3 failed login attempts',
      timestamp: new Date(Date.now() - 1200000),
      severity: 'low',
    },
  ];
};

export default function ClientActivityLog() {
  const [activities, setActivities] = useState<ActivityEntry[]>(generateMockActivity());
  const [suspiciousActivities] = useState<SuspiciousActivity[]>(generateSuspiciousActivity());
  const [selectedEventTypes, setSelectedEventTypes] = useState<EventType[]>([]);
  const [searchClient, setSearchClient] = useState('');
  const [searchIP, setSearchIP] = useState('');
  const [showFilters, setShowFilters] = useState(false);

  // Auto-refresh every 5 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      setActivities(generateMockActivity());
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  // Calculate statistics
  const stats = useMemo(() => {
    const now = Date.now();
    const oneDayAgo = now - 86400000;

    const activeSessions = 234; // Mock value
    const loginsToday = activities.filter(
      (a) => a.eventType === 'login' && a.timestamp.getTime() > oneDayAgo
    ).length;
    const failedLogins = 23; // Mock value
    const uniqueIPs = new Set(activities.map((a) => a.ipAddress)).size;

    return { activeSessions, loginsToday, failedLogins, uniqueIPs };
  }, [activities]);

  // Activity by hour (24h)
  const activityByHour = useMemo(() => {
    const hours = Array.from({ length: 24 }, (_, i) => ({
      hour: i,
      count: 0,
    }));

    activities.forEach((activity) => {
      const hour = activity.timestamp.getHours();
      hours[hour].count++;
    });

    return hours;
  }, [activities]);

  // Geographic distribution
  const geoDistribution = useMemo(() => {
    const countryMap = new Map<string, number>();

    activities
      .filter((a) => a.eventType === 'login')
      .forEach((a) => {
        countryMap.set(a.country, (countryMap.get(a.country) || 0) + 1);
      });

    return Array.from(countryMap.entries())
      .map(([country, count]) => ({ country, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10);
  }, [activities]);

  // Filtered activities
  const filteredActivities = useMemo(() => {
    let filtered = activities;

    if (selectedEventTypes.length > 0) {
      filtered = filtered.filter((a) => selectedEventTypes.includes(a.eventType));
    }

    if (searchClient) {
      filtered = filtered.filter(
        (a) =>
          a.clientName.toLowerCase().includes(searchClient.toLowerCase()) ||
          a.clientId.toLowerCase().includes(searchClient.toLowerCase())
      );
    }

    if (searchIP) {
      filtered = filtered.filter((a) => a.ipAddress.includes(searchIP));
    }

    return filtered.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
  }, [activities, selectedEventTypes, searchClient, searchIP]);

  const toggleEventType = (type: EventType) => {
    setSelectedEventTypes((prev) =>
      prev.includes(type) ? prev.filter((t) => t !== type) : [...prev, type]
    );
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'high':
        return '#EF4444';
      case 'medium':
        return '#F59E0B';
      case 'low':
        return '#F5C542';
      default:
        return '#888';
    }
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-[#E4E4E7] overflow-hidden">
      {/* Header */}
      <div className="h-12 bg-[#1E2026] border-b border-[#383A42] flex items-center justify-between px-4 flex-shrink-0">
        <div className="flex items-center gap-3">
          <Activity size={18} className="text-[#3B82F6]" />
          <h1 className="text-sm font-bold text-[#F5C542]">Client Activity & Session Log</h1>
        </div>
        <div className="flex items-center gap-3">
          <div className="text-xs text-[#888]">
            Auto-refresh: <span className="text-[#22C55E]">5s</span>
          </div>
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`px-3 py-1.5 ${
              showFilters ? 'bg-[#3B82F6]' : 'bg-[#383A42]'
            } hover:bg-[#3B82F6] text-white text-xs rounded transition-colors flex items-center gap-1.5`}
          >
            <Filter size={12} />
            Filters
          </button>
        </div>
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <div className="bg-[#1E2026] border-b border-[#383A42] p-4 space-y-3 flex-shrink-0">
          <div className="grid grid-cols-3 gap-4">
            {/* Client Search */}
            <div>
              <label className="text-xs text-[#888] block mb-1">Search Client</label>
              <div className="relative">
                <Search size={12} className="absolute left-2 top-2.5 text-[#888]" />
                <input
                  type="text"
                  value={searchClient}
                  onChange={(e) => setSearchClient(e.target.value)}
                  placeholder="Client name or ID"
                  className="w-full bg-[#121316] border border-[#383A42] text-white text-xs pl-7 pr-3 py-2 rounded focus:border-[#3B82F6] outline-none"
                />
              </div>
            </div>

            {/* IP Address Search */}
            <div>
              <label className="text-xs text-[#888] block mb-1">IP Address</label>
              <input
                type="text"
                value={searchIP}
                onChange={(e) => setSearchIP(e.target.value)}
                placeholder="e.g., 192.168"
                className="w-full bg-[#121316] border border-[#383A42] text-white text-xs px-3 py-2 rounded focus:border-[#3B82F6] outline-none"
              />
            </div>

            {/* Date Range Placeholder */}
            <div>
              <label className="text-xs text-[#888] block mb-1">Date Range</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  placeholder="Last 24h"
                  className="flex-1 bg-[#121316] border border-[#383A42] text-white text-xs px-3 py-2 rounded focus:border-[#3B82F6] outline-none"
                  disabled
                />
              </div>
            </div>
          </div>

          {/* Event Type Multi-Select */}
          <div>
            <label className="text-xs text-[#888] block mb-2">Event Types</label>
            <div className="flex flex-wrap gap-2">
              {Object.entries(EVENT_TYPE_CONFIG).map(([type, config]) => (
                <button
                  key={type}
                  onClick={() => toggleEventType(type as EventType)}
                  className={`px-2 py-1 text-xs rounded flex items-center gap-1 transition-colors ${
                    selectedEventTypes.includes(type as EventType)
                      ? 'text-white'
                      : 'text-[#888] bg-[#25272E] hover:bg-[#383A42]'
                  }`}
                  style={{
                    backgroundColor: selectedEventTypes.includes(type as EventType)
                      ? config.color
                      : undefined,
                  }}
                >
                  {config.icon}
                  {config.label}
                </button>
              ))}
            </div>
          </div>

          {/* Clear Filters */}
          {(selectedEventTypes.length > 0 || searchClient || searchIP) && (
            <button
              onClick={() => {
                setSelectedEventTypes([]);
                setSearchClient('');
                setSearchIP('');
              }}
              className="text-xs text-[#3B82F6] hover:text-[#60A5FA]"
            >
              Clear all filters
            </button>
          )}
        </div>
      )}

      {/* Main Content */}
      <div className="flex-1 overflow-auto p-4 space-y-4">
        {/* Stats Cards */}
        <div className="grid grid-cols-4 gap-4">
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="text-xs text-[#888]">Active Sessions Now</div>
              <Users size={14} className="text-[#22C55E]" />
            </div>
            <div className="text-2xl font-bold text-[#22C55E]">{stats.activeSessions}</div>
            <div className="text-xs text-[#666] mt-1">Currently online</div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="text-xs text-[#888]">Logins Today</div>
              <LogIn size={14} className="text-[#3B82F6]" />
            </div>
            <div className="text-2xl font-bold text-[#3B82F6]">{stats.loginsToday.toLocaleString()}</div>
            <div className="text-xs text-[#666] mt-1">Last 24 hours</div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="text-xs text-[#888]">Failed Logins</div>
              <Shield size={14} className="text-[#EF4444]" />
            </div>
            <div className="text-2xl font-bold text-[#EF4444]">{stats.failedLogins}</div>
            <div className="text-xs text-[#666] mt-1">Security alerts</div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="text-xs text-[#888]">Unique IPs</div>
              <Globe size={14} className="text-[#9B59B6]" />
            </div>
            <div className="text-2xl font-bold text-[#9B59B6]">{stats.uniqueIPs}</div>
            <div className="text-xs text-[#666] mt-1">Distinct addresses</div>
          </div>
        </div>

        {/* Charts Row */}
        <div className="grid grid-cols-2 gap-4">
          {/* Activity Volume by Hour */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="text-sm font-bold text-[#F5C542] mb-4">Activity Volume (Last 24h)</div>
            <div className="h-32">
              <svg width="100%" height="100%" viewBox="0 0 400 120" preserveAspectRatio="none">
                {/* Grid lines */}
                {[0, 25, 50, 75, 100].map((y) => (
                  <line
                    key={y}
                    x1="0"
                    y1={y * 1.2}
                    x2="400"
                    y2={y * 1.2}
                    stroke="#383A42"
                    strokeWidth="1"
                  />
                ))}

                {/* Area chart */}
                <defs>
                  <linearGradient id="activityGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                    <stop offset="0%" stopColor="#3B82F6" stopOpacity="0.3" />
                    <stop offset="100%" stopColor="#3B82F6" stopOpacity="0.05" />
                  </linearGradient>
                </defs>

                <path
                  d={`M 0,120 ${activityByHour
                    .map((h, i) => {
                      const x = (i / 23) * 400;
                      const maxCount = Math.max(...activityByHour.map((h) => h.count));
                      const y = 120 - (h.count / maxCount) * 100;
                      return `L ${x},${y}`;
                    })
                    .join(' ')} L 400,120 Z`}
                  fill="url(#activityGradient)"
                />

                <polyline
                  points={activityByHour
                    .map((h, i) => {
                      const x = (i / 23) * 400;
                      const maxCount = Math.max(...activityByHour.map((h) => h.count));
                      const y = 120 - (h.count / maxCount) * 100;
                      return `${x},${y}`;
                    })
                    .join(' ')}
                  fill="none"
                  stroke="#3B82F6"
                  strokeWidth="2"
                />
              </svg>
            </div>
          </div>

          {/* Geographic Distribution */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="text-sm font-bold text-[#F5C542] mb-4">Top Countries (Logins)</div>
            <div className="space-y-2 text-xs">
              {geoDistribution.map((geo, idx) => {
                const maxCount = geoDistribution[0]?.count || 1;
                const percentage = (geo.count / maxCount) * 100;

                return (
                  <div key={geo.country}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-[#CCC]">{geo.country}</span>
                      <span className="text-[#888] font-mono">{geo.count}</span>
                    </div>
                    <div className="w-full h-1.5 bg-[#383A42] rounded overflow-hidden">
                      <div
                        className="h-full bg-[#3B82F6] rounded"
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Suspicious Activity Alerts */}
        {suspiciousActivities.length > 0 && (
          <div className="bg-[#1E2026] border border-[#383A42] rounded">
            <div className="h-9 bg-[#25272E] border-b border-[#383A42] flex items-center px-4">
              <AlertTriangle size={14} className="text-[#F59E0B] mr-2" />
              <div className="text-sm font-bold text-[#F5C542]">
                Suspicious Activity Alerts ({suspiciousActivities.length})
              </div>
            </div>
            <div className="p-3 space-y-2">
              {suspiciousActivities.map((alert) => (
                <div
                  key={alert.id}
                  className="bg-[#121316] border border-[#383A42] rounded p-3 flex items-center justify-between"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className="w-1 h-12 rounded"
                      style={{ backgroundColor: getSeverityColor(alert.severity) }}
                    />
                    <div>
                      <div className="text-xs font-bold text-[#CCC]">{alert.clientName}</div>
                      <div className="text-xs text-[#888] mt-1">{alert.description}</div>
                      <div className="text-[10px] text-[#666] mt-1">
                        {alert.timestamp.toLocaleString('en-US', {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span
                      className="px-2 py-0.5 rounded text-[10px] font-bold uppercase"
                      style={{
                        backgroundColor: getSeverityColor(alert.severity) + '20',
                        color: getSeverityColor(alert.severity),
                      }}
                    >
                      {alert.severity}
                    </span>
                    <button className="px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs rounded transition-colors">
                      Investigate
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Activity Feed */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded">
          <div className="h-9 bg-[#25272E] border-b border-[#383A42] flex items-center justify-between px-4">
            <div className="text-sm font-bold text-[#F5C542]">
              Live Activity Feed ({filteredActivities.length})
            </div>
            <div className="text-xs text-[#888]">Real-time updates</div>
          </div>
          <div className="overflow-auto max-h-[400px]">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                <tr className="text-[#888]">
                  <th className="text-left py-2 px-3 font-normal">Timestamp</th>
                  <th className="text-left py-2 px-3 font-normal">Client</th>
                  <th className="text-left py-2 px-3 font-normal">Event</th>
                  <th className="text-left py-2 px-3 font-normal">Details</th>
                  <th className="text-left py-2 px-3 font-normal">IP Address</th>
                  <th className="text-left py-2 px-3 font-normal">Device</th>
                  <th className="text-left py-2 px-3 font-normal">Location</th>
                </tr>
              </thead>
              <tbody>
                {filteredActivities.slice(0, 50).map((activity) => {
                  const config = EVENT_TYPE_CONFIG[activity.eventType];
                  return (
                    <tr
                      key={activity.id}
                      className="border-b border-[#383A42] hover:bg-[#25272E]"
                    >
                      <td className="py-2 px-3 text-[#888] font-mono">
                        {activity.timestamp.toLocaleString('en-US', {
                          hour: '2-digit',
                          minute: '2-digit',
                          second: '2-digit',
                        })}
                      </td>
                      <td className="py-2 px-3">
                        <div className="text-[#3B82F6] font-bold">{activity.clientName}</div>
                        <div className="text-[#666] text-[10px] font-mono">{activity.clientId}</div>
                      </td>
                      <td className="py-2 px-3">
                        <span
                          className="px-2 py-1 rounded text-[10px] font-bold flex items-center gap-1 w-fit"
                          style={{ backgroundColor: config.bgColor, color: config.color }}
                        >
                          {config.icon}
                          {config.label}
                        </span>
                      </td>
                      <td className="py-2 px-3 text-[#CCC]">{activity.details}</td>
                      <td className="py-2 px-3 text-[#888] font-mono">{activity.ipAddress}</td>
                      <td className="py-2 px-3 text-[#888]">
                        {activity.device} / {activity.browser}
                      </td>
                      <td className="py-2 px-3 text-[#888]">{activity.country}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
