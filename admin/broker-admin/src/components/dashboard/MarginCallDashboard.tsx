'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { AlertTriangle, TrendingDown, Activity, Percent, Bell, X, Settings as SettingsIcon } from 'lucide-react';

interface AtRiskAccount {
  accountId: string;
  clientName: string;
  equity: number;
  marginUsed: number;
  freeMargin: number;
  marginLevel: number;
  status: 'safe' | 'warning' | 'margin-call' | 'stop-out';
  lastUpdated: Date;
}

interface MarginCallHistory {
  id: string;
  accountId: string;
  clientName: string;
  timestamp: Date;
  marginLevel: number;
  action: 'notified' | 'positions-closed' | 'auto-liquidated' | 'resolved';
  outcome: 'recovered' | 'liquidated' | 'pending';
  closedPositions?: number;
}

interface AutoCloseSettings {
  marginCallLevel: number;
  stopOutLevel: number;
  notifyOnWarning: boolean;
  notifyOnMarginCall: boolean;
  autoCloseEnabled: boolean;
}

// Mock data generator
const generateMockAccounts = (): AtRiskAccount[] => {
  const statuses: Array<'safe' | 'warning' | 'margin-call' | 'stop-out'> = [
    'safe', 'safe', 'safe', 'safe', 'safe', 'safe', 'safe', 'safe',
    'warning', 'warning', 'warning', 'warning', 'warning',
    'margin-call', 'margin-call', 'margin-call', 'margin-call',
    'stop-out', 'stop-out'
  ];

  return Array.from({ length: 30 }, (_, i) => {
    const status = statuses[i] || 'safe';
    let marginLevel: number;

    switch (status) {
      case 'safe':
        marginLevel = 200 + Math.random() * 300; // 200-500%
        break;
      case 'warning':
        marginLevel = 100 + Math.random() * 100; // 100-200%
        break;
      case 'margin-call':
        marginLevel = 50 + Math.random() * 50; // 50-100%
        break;
      case 'stop-out':
        marginLevel = Math.random() * 50; // 0-50%
        break;
    }

    const marginUsed = 5000 + Math.random() * 45000;
    const equity = (marginUsed * marginLevel) / 100;
    const freeMargin = equity - marginUsed;

    return {
      accountId: `AC${String(680962851 + i).padStart(9, '0')}`,
      clientName: `Client ${i + 1}`,
      equity,
      marginUsed,
      freeMargin,
      marginLevel,
      status,
      lastUpdated: new Date(Date.now() - Math.random() * 300000), // Last 5 minutes
    };
  });
};

const generateMockHistory = (): MarginCallHistory[] => {
  const actions: Array<'notified' | 'positions-closed' | 'auto-liquidated' | 'resolved'> =
    ['notified', 'positions-closed', 'auto-liquidated', 'resolved'];
  const outcomes: Array<'recovered' | 'liquidated' | 'pending'> =
    ['recovered', 'liquidated', 'pending'];

  return Array.from({ length: 15 }, (_, i) => ({
    id: `MC${String(1000 + i).padStart(4, '0')}`,
    accountId: `AC${String(680962851 + i).padStart(9, '0')}`,
    clientName: `Client ${i + 1}`,
    timestamp: new Date(Date.now() - Math.random() * 86400000), // Last 24 hours
    marginLevel: 30 + Math.random() * 70,
    action: actions[Math.floor(Math.random() * actions.length)],
    outcome: outcomes[Math.floor(Math.random() * outcomes.length)],
    closedPositions: Math.random() > 0.5 ? Math.floor(Math.random() * 5) + 1 : undefined,
  }));
};

export default function MarginCallDashboard() {
  const [accounts, setAccounts] = useState<AtRiskAccount[]>(generateMockAccounts());
  const [history, setHistory] = useState<MarginCallHistory[]>(generateMockHistory());
  const [settings, setSettings] = useState<AutoCloseSettings>({
    marginCallLevel: 100,
    stopOutLevel: 50,
    notifyOnWarning: true,
    notifyOnMarginCall: true,
    autoCloseEnabled: true,
  });
  const [showSettings, setShowSettings] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<string | null>(null);

  // Auto-refresh every 5 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      setAccounts(generateMockAccounts());
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  // Calculate statistics
  const stats = useMemo(() => {
    const atRisk = accounts.filter(a => a.status !== 'safe').length;
    const marginCalls = accounts.filter(a => a.status === 'margin-call' || a.status === 'stop-out').length;
    const liquidationsToday = history.filter(h =>
      h.outcome === 'liquidated' &&
      h.timestamp.getTime() > Date.now() - 86400000
    ).length;
    const avgMarginLevel = accounts.reduce((sum, a) => sum + a.marginLevel, 0) / accounts.length;

    return {
      atRisk,
      marginCalls,
      liquidationsToday,
      avgMarginLevel,
    };
  }, [accounts, history]);

  // Margin level distribution for histogram
  const distribution = useMemo(() => {
    const buckets = [
      { label: '0-50%', min: 0, max: 50, count: 0, color: '#EF4444' },
      { label: '50-100%', min: 50, max: 100, count: 0, color: '#F97316' },
      { label: '100-200%', min: 100, max: 200, count: 0, color: '#F59E0B' },
      { label: '200-300%', min: 200, max: 300, count: 0, color: '#84CC16' },
      { label: '300%+', min: 300, max: Infinity, count: 0, color: '#22C55E' },
    ];

    accounts.forEach(acc => {
      const bucket = buckets.find(b => acc.marginLevel >= b.min && acc.marginLevel < b.max);
      if (bucket) bucket.count++;
    });

    return buckets;
  }, [accounts]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'safe': return '#22C55E';
      case 'warning': return '#F59E0B';
      case 'margin-call': return '#F97316';
      case 'stop-out': return '#EF4444';
      default: return '#888';
    }
  };

  const getRowBgColor = (marginLevel: number) => {
    if (marginLevel > 200) return 'rgba(34, 197, 94, 0.1)'; // Green
    if (marginLevel > 100) return 'rgba(245, 158, 11, 0.1)'; // Yellow
    if (marginLevel > 50) return 'rgba(249, 115, 22, 0.1)'; // Orange
    return 'rgba(239, 68, 68, 0.1)'; // Red
  };

  const handleNotify = (accountId: string) => {
    console.log(`Sending notification for account ${accountId}`);
    // In production, call API
  };

  const handleClosePositions = (accountId: string) => {
    console.log(`Closing positions for account ${accountId}`);
    // In production, call API
  };

  const handleSaveSettings = () => {
    console.log('Saving settings:', settings);
    setShowSettings(false);
    // In production, call API
  };

  // Overall margin health gauge calculation
  const overallHealth = useMemo(() => {
    const avgMargin = stats.avgMarginLevel;
    const healthScore = Math.min(100, (avgMargin / 300) * 100); // 300% = 100 health
    return healthScore;
  }, [stats.avgMarginLevel]);

  return (
    <div className="h-full flex flex-col bg-[#121316] text-[#E4E4E7] overflow-hidden">
      {/* Header */}
      <div className="h-12 bg-[#1E2026] border-b border-[#383A42] flex items-center justify-between px-4 flex-shrink-0">
        <div className="flex items-center gap-3">
          <AlertTriangle size={18} className="text-[#F97316]" />
          <h1 className="text-sm font-bold text-[#F5C542]">Margin Call Management</h1>
        </div>
        <div className="flex items-center gap-2">
          <div className="text-xs text-[#888]">
            Auto-refresh: <span className="text-[#22C55E]">5s</span>
          </div>
          <button
            onClick={() => setShowSettings(!showSettings)}
            className="px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs rounded transition-colors flex items-center gap-1.5"
          >
            <SettingsIcon size={12} />
            Settings
          </button>
        </div>
      </div>

      {/* Settings Modal */}
      {showSettings && (
        <div className="absolute inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded w-[500px] shadow-xl">
            <div className="h-10 bg-[#25272E] border-b border-[#383A42] flex items-center justify-between px-4">
              <div className="text-sm font-bold text-[#F5C542]">Auto-Close Settings</div>
              <button onClick={() => setShowSettings(false)} className="text-[#888] hover:text-white">
                <X size={14} />
              </button>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <label className="text-xs text-[#888] block mb-1">Margin Call Level (%)</label>
                <input
                  type="number"
                  value={settings.marginCallLevel}
                  onChange={(e) => setSettings({ ...settings, marginCallLevel: Number(e.target.value) })}
                  className="w-full bg-[#121316] border border-[#383A42] text-white text-sm px-3 py-2 rounded focus:border-[#3B82F6] outline-none"
                />
              </div>
              <div>
                <label className="text-xs text-[#888] block mb-1">Stop Out Level (%)</label>
                <input
                  type="number"
                  value={settings.stopOutLevel}
                  onChange={(e) => setSettings({ ...settings, stopOutLevel: Number(e.target.value) })}
                  className="w-full bg-[#121316] border border-[#383A42] text-white text-sm px-3 py-2 rounded focus:border-[#3B82F6] outline-none"
                />
              </div>
              <div className="space-y-2">
                <label className="flex items-center gap-2 text-sm cursor-pointer">
                  <input
                    type="checkbox"
                    checked={settings.notifyOnWarning}
                    onChange={(e) => setSettings({ ...settings, notifyOnWarning: e.target.checked })}
                    className="w-4 h-4"
                  />
                  <span className="text-[#CCC]">Notify on Warning (100-200%)</span>
                </label>
                <label className="flex items-center gap-2 text-sm cursor-pointer">
                  <input
                    type="checkbox"
                    checked={settings.notifyOnMarginCall}
                    onChange={(e) => setSettings({ ...settings, notifyOnMarginCall: e.target.checked })}
                    className="w-4 h-4"
                  />
                  <span className="text-[#CCC]">Notify on Margin Call (50-100%)</span>
                </label>
                <label className="flex items-center gap-2 text-sm cursor-pointer">
                  <input
                    type="checkbox"
                    checked={settings.autoCloseEnabled}
                    onChange={(e) => setSettings({ ...settings, autoCloseEnabled: e.target.checked })}
                    className="w-4 h-4"
                  />
                  <span className="text-[#CCC]">Enable Auto-Close at Stop Out Level</span>
                </label>
              </div>
              <div className="flex justify-end gap-2 pt-4 border-t border-[#383A42]">
                <button
                  onClick={() => setShowSettings(false)}
                  className="px-4 py-2 bg-[#383A42] hover:bg-[#4A4C54] text-white text-xs rounded transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSaveSettings}
                  className="px-4 py-2 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs rounded transition-colors"
                >
                  Save Settings
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Main Content */}
      <div className="flex-1 overflow-auto p-4 space-y-4">
        {/* Stats Cards */}
        <div className="grid grid-cols-4 gap-4">
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="text-xs text-[#888]">Accounts at Risk</div>
              <AlertTriangle size={14} className="text-[#F97316]" />
            </div>
            <div className="text-2xl font-bold text-[#F97316]">{stats.atRisk}</div>
            <div className="text-xs text-[#666] mt-1">
              {((stats.atRisk / accounts.length) * 100).toFixed(1)}% of total
            </div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="text-xs text-[#888]">Active Margin Calls</div>
              <TrendingDown size={14} className="text-[#EF4444]" />
            </div>
            <div className="text-2xl font-bold text-[#EF4444]">{stats.marginCalls}</div>
            <div className="text-xs text-[#666] mt-1">Require immediate action</div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="text-xs text-[#888]">Liquidations Today</div>
              <X size={14} className="text-[#DC2626]" />
            </div>
            <div className="text-2xl font-bold text-[#DC2626]">{stats.liquidationsToday}</div>
            <div className="text-xs text-[#666] mt-1">Last 24 hours</div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="text-xs text-[#888]">Avg Margin Level</div>
              <Percent size={14} className="text-[#22C55E]" />
            </div>
            <div className="text-2xl font-bold text-[#22C55E]">{stats.avgMarginLevel.toFixed(1)}%</div>
            <div className="text-xs text-[#666] mt-1">Platform average</div>
          </div>
        </div>

        {/* Charts Row */}
        <div className="grid grid-cols-2 gap-4">
          {/* Overall Margin Health Gauge */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="text-sm font-bold text-[#F5C542] mb-4">Overall Platform Margin Health</div>
            <div className="flex items-center justify-center">
              <svg width="200" height="120" viewBox="0 0 200 120">
                {/* Background arc */}
                <path
                  d="M 20,100 A 80,80 0 0,1 180,100"
                  fill="none"
                  stroke="#383A42"
                  strokeWidth="20"
                  strokeLinecap="round"
                />
                {/* Health arc */}
                <path
                  d="M 20,100 A 80,80 0 0,1 180,100"
                  fill="none"
                  stroke={overallHealth > 66 ? '#22C55E' : overallHealth > 33 ? '#F59E0B' : '#EF4444'}
                  strokeWidth="20"
                  strokeLinecap="round"
                  strokeDasharray={`${(overallHealth / 100) * 251.2} 251.2`}
                />
                {/* Center value */}
                <text x="100" y="80" textAnchor="middle" fill="#E4E4E7" fontSize="28" fontWeight="bold">
                  {overallHealth.toFixed(0)}
                </text>
                <text x="100" y="100" textAnchor="middle" fill="#888" fontSize="12">
                  Health Score
                </text>
              </svg>
            </div>
            <div className="text-xs text-center text-[#888] mt-2">
              Based on average margin level: {stats.avgMarginLevel.toFixed(1)}%
            </div>
          </div>

          {/* Margin Level Distribution Histogram */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="text-sm font-bold text-[#F5C542] mb-4">Margin Level Distribution</div>
            <div className="flex items-end justify-between gap-2 h-32">
              {distribution.map((bucket, idx) => {
                const maxCount = Math.max(...distribution.map(b => b.count));
                const height = (bucket.count / maxCount) * 100;

                return (
                  <div key={idx} className="flex-1 flex flex-col items-center gap-1">
                    <div className="text-xs text-[#888] mb-1">{bucket.count}</div>
                    <div
                      className="w-full rounded-t"
                      style={{ height: `${height}%`, backgroundColor: bucket.color }}
                    />
                    <div className="text-[10px] text-[#666] mt-1 text-center">{bucket.label}</div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* At-Risk Accounts Table */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded">
          <div className="h-9 bg-[#25272E] border-b border-[#383A42] flex items-center px-4">
            <div className="text-sm font-bold text-[#F5C542]">At-Risk Accounts ({accounts.filter(a => a.status !== 'safe').length})</div>
          </div>
          <div className="overflow-auto max-h-[300px]">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                <tr className="text-[#888]">
                  <th className="text-left py-2 px-3 font-normal">Account ID</th>
                  <th className="text-left py-2 px-3 font-normal">Client</th>
                  <th className="text-right py-2 px-3 font-normal">Equity</th>
                  <th className="text-right py-2 px-3 font-normal">Margin Used</th>
                  <th className="text-right py-2 px-3 font-normal">Free Margin</th>
                  <th className="text-right py-2 px-3 font-normal">Margin Level</th>
                  <th className="text-center py-2 px-3 font-normal">Status</th>
                  <th className="text-center py-2 px-3 font-normal">Actions</th>
                </tr>
              </thead>
              <tbody>
                {accounts
                  .filter(a => a.status !== 'safe')
                  .sort((a, b) => a.marginLevel - b.marginLevel)
                  .map((account) => (
                    <tr
                      key={account.accountId}
                      className="border-b border-[#383A42] hover:bg-[#25272E] cursor-pointer"
                      style={{ backgroundColor: getRowBgColor(account.marginLevel) }}
                      onClick={() => setSelectedAccount(account.accountId)}
                    >
                      <td className="py-2 px-3 text-[#3B82F6] font-mono">{account.accountId}</td>
                      <td className="py-2 px-3 text-[#CCC]">{account.clientName}</td>
                      <td className="py-2 px-3 text-right text-[#CCC] font-mono">
                        ${account.equity.toFixed(2)}
                      </td>
                      <td className="py-2 px-3 text-right text-[#CCC] font-mono">
                        ${account.marginUsed.toFixed(2)}
                      </td>
                      <td className="py-2 px-3 text-right font-mono" style={{ color: account.freeMargin >= 0 ? '#22C55E' : '#EF4444' }}>
                        ${account.freeMargin.toFixed(2)}
                      </td>
                      <td className="py-2 px-3 text-right font-bold font-mono" style={{ color: getStatusColor(account.status) }}>
                        {account.marginLevel.toFixed(1)}%
                      </td>
                      <td className="py-2 px-3 text-center">
                        <span
                          className="px-2 py-0.5 rounded text-[10px] font-bold uppercase"
                          style={{
                            backgroundColor: getStatusColor(account.status) + '20',
                            color: getStatusColor(account.status),
                          }}
                        >
                          {account.status.replace('-', ' ')}
                        </span>
                      </td>
                      <td className="py-2 px-3">
                        <div className="flex items-center justify-center gap-1">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleNotify(account.accountId);
                            }}
                            className="px-2 py-1 bg-[#3B82F6] hover:bg-[#2563EB] text-white rounded text-[10px] transition-colors flex items-center gap-1"
                          >
                            <Bell size={10} />
                            Notify
                          </button>
                          {(account.status === 'margin-call' || account.status === 'stop-out') && (
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleClosePositions(account.accountId);
                              }}
                              className="px-2 py-1 bg-[#EF4444] hover:bg-[#DC2626] text-white rounded text-[10px] transition-colors flex items-center gap-1"
                            >
                              <X size={10} />
                              Close
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Margin Call History */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded">
          <div className="h-9 bg-[#25272E] border-b border-[#383A42] flex items-center px-4">
            <div className="text-sm font-bold text-[#F5C542]">Margin Call History (Last 24 Hours)</div>
          </div>
          <div className="overflow-auto max-h-[250px]">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                <tr className="text-[#888]">
                  <th className="text-left py-2 px-3 font-normal">ID</th>
                  <th className="text-left py-2 px-3 font-normal">Account</th>
                  <th className="text-left py-2 px-3 font-normal">Client</th>
                  <th className="text-left py-2 px-3 font-normal">Timestamp</th>
                  <th className="text-right py-2 px-3 font-normal">Margin Level</th>
                  <th className="text-center py-2 px-3 font-normal">Action</th>
                  <th className="text-center py-2 px-3 font-normal">Outcome</th>
                  <th className="text-right py-2 px-3 font-normal">Closed Positions</th>
                </tr>
              </thead>
              <tbody>
                {history
                  .sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime())
                  .map((record) => (
                    <tr key={record.id} className="border-b border-[#383A42] hover:bg-[#25272E]">
                      <td className="py-2 px-3 text-[#888] font-mono">{record.id}</td>
                      <td className="py-2 px-3 text-[#3B82F6] font-mono">{record.accountId}</td>
                      <td className="py-2 px-3 text-[#CCC]">{record.clientName}</td>
                      <td className="py-2 px-3 text-[#888]">
                        {record.timestamp.toLocaleString('en-US', {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </td>
                      <td className="py-2 px-3 text-right text-[#F97316] font-mono">
                        {record.marginLevel.toFixed(1)}%
                      </td>
                      <td className="py-2 px-3 text-center">
                        <span className="px-2 py-0.5 rounded text-[10px] bg-[#3B82F6]/20 text-[#3B82F6]">
                          {record.action.replace('-', ' ')}
                        </span>
                      </td>
                      <td className="py-2 px-3 text-center">
                        <span
                          className="px-2 py-0.5 rounded text-[10px] font-bold"
                          style={{
                            backgroundColor:
                              record.outcome === 'recovered'
                                ? '#22C55E20'
                                : record.outcome === 'liquidated'
                                ? '#EF444420'
                                : '#F59E0B20',
                            color:
                              record.outcome === 'recovered'
                                ? '#22C55E'
                                : record.outcome === 'liquidated'
                                ? '#EF4444'
                                : '#F59E0B',
                          }}
                        >
                          {record.outcome}
                        </span>
                      </td>
                      <td className="py-2 px-3 text-right text-[#888]">
                        {record.closedPositions ?? '-'}
                      </td>
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
