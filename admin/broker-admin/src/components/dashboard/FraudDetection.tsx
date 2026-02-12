'use client';

import { useState, useEffect, useMemo } from 'react';
import {
  Shield, AlertTriangle, Globe, ChevronDown, ChevronUp, X,
  Search, Check, Ban, Eye, Activity, MapPin, TrendingUp, Users, Clock
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// Types matching backend
type LoginStatus = 'success' | 'failed' | 'blocked';
type AlertType = 'multiple_countries' | 'velocity_login' | 'impossible_travel' |
  'new_device' | 'unusual_hours' | 'large_withdrawal' | 'multiple_failed_logins' |
  'vpn_detected' | 'tor_detected' | 'blacklisted_country';
type Severity = 'low' | 'medium' | 'high' | 'critical';
type AlertStatus = 'active' | 'resolved' | 'ignored';

interface LoginAttempt {
  id: string;
  client_id: string;
  client_name: string;
  ip: string;
  country: string;
  city: string;
  device: string;
  browser: string;
  status: LoginStatus;
  timestamp: string;
  risk_score: number;
  flags?: string[];
}

interface FraudAlert {
  id: string;
  client_id: string;
  client_name: string;
  alert_type: AlertType;
  severity: Severity;
  description: string;
  detected_at: string;
  resolved_at?: string;
  resolved_by?: string;
  status: AlertStatus;
  metadata?: Record<string, any>;
}

interface FraudRule {
  id: string;
  name: string;
  type: AlertType;
  condition: string;
  action: string;
  is_active: boolean;
  trigger_count: number;
  created_at: string;
  updated_at: string;
}

interface GeoReport {
  country: string;
  login_count: number;
  unique_clients: number;
  risk_level: string;
  flagged_count: number;
  success_rate: number;
}

interface FraudStats {
  total_alerts: number;
  active_alerts: number;
  resolved_alerts: number;
  resolution_rate: number;
  alerts_by_type: Record<string, number>;
  top_risky_countries: string[];
  flagged_clients: number;
  avg_risk_score: number;
  blocked_logins: number;
}

type SortField = 'timestamp' | 'ip' | 'country' | 'status' | 'risk_score';
type SortDirection = 'asc' | 'desc';

export default function FraudDetection() {
  const [logins, setLogins] = useState<LoginAttempt[]>([]);
  const [alerts, setAlerts] = useState<FraudAlert[]>([]);
  const [rules, setRules] = useState<FraudRule[]>([]);
  const [geoData, setGeoData] = useState<GeoReport[]>([]);
  const [stats, setStats] = useState<FraudStats | null>(null);
  const [loading, setLoading] = useState(true);

  // Filters
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [riskFilter, setRiskFilter] = useState<string>('all');
  const [alertStatusFilter, setAlertStatusFilter] = useState<string>('all');

  // Sorting
  const [sortField, setSortField] = useState<SortField>('timestamp');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  // Modals
  const [resolveAlertModal, setResolveAlertModal] = useState<FraudAlert | null>(null);
  const [resolveAction, setResolveAction] = useState<'resolved' | 'ignored'>('resolved');
  const [ipToBlacklist, setIpToBlacklist] = useState('');
  const [showBlacklistModal, setShowBlacklistModal] = useState(false);

  // Expanded rows
  const [expandedLogin, setExpandedLogin] = useState<string | null>(null);

  useEffect(() => {
    fetchAllData();
  }, []);

  const fetchAllData = async () => {
    setLoading(true);
    const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
    if (!token) {
      setLoading(false);
      return;
    }

    const headers = {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    };

    try {
      const [loginsRes, alertsRes, rulesRes, geoRes, statsRes] = await Promise.all([
        fetch(API_CONFIG.FRAUD_LOGINS, { headers }),
        fetch(API_CONFIG.FRAUD_ALERTS, { headers }),
        fetch(API_CONFIG.FRAUD_RULES, { headers }),
        fetch(API_CONFIG.FRAUD_GEO, { headers }),
        fetch(API_CONFIG.FRAUD_STATS, { headers })
      ]);

      if (loginsRes.ok) {
        const data = await loginsRes.json();
        setLogins(data.data || []);
      }

      if (alertsRes.ok) {
        const data = await alertsRes.json();
        setAlerts(data.alerts || []);
      }

      if (rulesRes.ok) {
        const data = await rulesRes.json();
        setRules(data.rules || []);
      }

      if (geoRes.ok) {
        const data = await geoRes.json();
        setGeoData(data.countries || []);
      }

      if (statsRes.ok) {
        const data = await statsRes.json();
        setStats(data.stats);
      }
    } catch (error) {
      console.error('Failed to fetch fraud data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleResolveAlert = async (alertId: string, action: 'resolved' | 'ignored') => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
    if (!token) return;

    try {
      const res = await fetch(API_CONFIG.FRAUD_ALERT_RESOLVE(alertId), {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status: action })
      });

      if (res.ok) {
        fetchAllData();
        setResolveAlertModal(null);
      }
    } catch (error) {
      console.error('Failed to resolve alert:', error);
    }
  };

  const handleToggleRule = async (ruleId: string, isActive: boolean) => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
    if (!token) return;

    try {
      const res = await fetch(API_CONFIG.FRAUD_RULE_UPDATE(ruleId), {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ is_active: isActive })
      });

      if (res.ok) {
        setRules(rules.map(r => r.id === ruleId ? { ...r, is_active: isActive } : r));
      }
    } catch (error) {
      console.error('Failed to update rule:', error);
    }
  };

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const filteredAndSortedLogins = useMemo(() => {
    let filtered = logins.filter(login => {
      if (statusFilter !== 'all' && login.status !== statusFilter) return false;
      if (riskFilter === 'high' && login.risk_score < 70) return false;
      if (riskFilter === 'medium' && (login.risk_score < 40 || login.risk_score >= 70)) return false;
      if (riskFilter === 'low' && login.risk_score >= 40) return false;
      if (searchQuery && !(
        login.ip.toLowerCase().includes(searchQuery.toLowerCase()) ||
        login.client_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        login.country.toLowerCase().includes(searchQuery.toLowerCase())
      )) return false;
      return true;
    });

    filtered.sort((a, b) => {
      let aVal: any = a[sortField];
      let bVal: any = b[sortField];

      if (sortField === 'timestamp') {
        aVal = new Date(aVal).getTime();
        bVal = new Date(bVal).getTime();
      }

      if (sortDirection === 'asc') {
        return aVal > bVal ? 1 : -1;
      } else {
        return aVal < bVal ? 1 : -1;
      }
    });

    return filtered;
  }, [logins, statusFilter, riskFilter, searchQuery, sortField, sortDirection]);

  const filteredAlerts = useMemo(() => {
    return alerts.filter(alert => {
      if (alertStatusFilter === 'all') return true;
      return alert.status === alertStatusFilter;
    });
  }, [alerts, alertStatusFilter]);

  const getSeverityBadge = (severity: Severity) => {
    const colors = {
      low: 'bg-[#10B981]/20 text-[#10B981] border-[#10B981]/30',
      medium: 'bg-[#F59E0B]/20 text-[#F59E0B] border-[#F59E0B]/30',
      high: 'bg-[#EF4444]/20 text-[#EF4444] border-[#EF4444]/30',
      critical: 'bg-[#DC2626]/20 text-[#DC2626] border-[#DC2626]/30'
    };
    return colors[severity] || colors.low;
  };

  const getStatusBadge = (status: LoginStatus) => {
    const colors = {
      success: 'bg-[#10B981]/20 text-[#10B981]',
      failed: 'bg-[#EF4444]/20 text-[#EF4444]',
      blocked: 'bg-[#DC2626]/20 text-[#DC2626]'
    };
    return colors[status] || colors.success;
  };

  const getRiskBadge = (score: number) => {
    if (score >= 70) return 'bg-[#DC2626]/20 text-[#DC2626]';
    if (score >= 40) return 'bg-[#F59E0B]/20 text-[#F59E0B]';
    return 'bg-[#10B981]/20 text-[#10B981]';
  };

  const formatAlertType = (type: AlertType) => {
    return type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full bg-[#121316]">
        <div className="text-[#666] text-sm">Loading fraud detection data...</div>
      </div>
    );
  }

  return (
    <div className="h-full bg-[#121316] overflow-auto">
      <div className="p-4 space-y-4">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Shield className="text-[#F5C542]" size={20} />
            <h1 className="text-[#F5C542] font-bold text-lg">Fraud Detection & IP Monitoring</h1>
          </div>
        </div>

        {/* Stats Overview Cards */}
        {stats && (
          <div className="grid grid-cols-4 gap-3">
            <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-[#888] text-xs">Total Alerts</span>
                <AlertTriangle size={14} className="text-[#F59E0B]" />
              </div>
              <div className="text-white text-2xl font-bold">{stats.total_alerts}</div>
              <div className="text-[#666] text-xs mt-1">{stats.active_alerts} active</div>
            </div>

            <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-[#888] text-xs">Resolution Rate</span>
                <Check size={14} className="text-[#10B981]" />
              </div>
              <div className="text-white text-2xl font-bold">{stats.resolution_rate.toFixed(1)}%</div>
              <div className="text-[#666] text-xs mt-1">{stats.resolved_alerts} resolved</div>
            </div>

            <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-[#888] text-xs">Flagged Clients</span>
                <Users size={14} className="text-[#EF4444]" />
              </div>
              <div className="text-white text-2xl font-bold">{stats.flagged_clients}</div>
              <div className="text-[#666] text-xs mt-1">Avg risk: {stats.avg_risk_score.toFixed(1)}</div>
            </div>

            <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-[#888] text-xs">Blocked Logins</span>
                <Ban size={14} className="text-[#DC2626]" />
              </div>
              <div className="text-white text-2xl font-bold">{stats.blocked_logins}</div>
              <div className="text-[#666] text-xs mt-1">Security blocks</div>
            </div>
          </div>
        )}

        {/* Fraud Alert Cards */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded">
          <div className="border-b border-[#383A42] p-3 flex items-center justify-between">
            <h2 className="text-white font-bold text-sm flex items-center gap-2">
              <AlertTriangle size={14} className="text-[#F59E0B]" />
              Fraud Alerts ({filteredAlerts.length})
            </h2>
            <select
              value={alertStatusFilter}
              onChange={e => setAlertStatusFilter(e.target.value)}
              className="bg-[#121316] border border-[#383A42] text-white text-xs px-2 py-1 rounded"
            >
              <option value="all">All Alerts</option>
              <option value="active">Active</option>
              <option value="resolved">Resolved</option>
              <option value="ignored">Ignored</option>
            </select>
          </div>
          <div className="p-3 max-h-64 overflow-y-auto space-y-2">
            {filteredAlerts.slice(0, 10).map(alert => (
              <div key={alert.id} className="bg-[#121316] border border-[#383A42] rounded p-3">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className={`text-xs px-2 py-0.5 rounded border ${getSeverityBadge(alert.severity)}`}>
                        {alert.severity.toUpperCase()}
                      </span>
                      <span className="text-[#888] text-xs">{formatAlertType(alert.alert_type)}</span>
                    </div>
                    <div className="text-white text-xs mb-1">{alert.description}</div>
                    <div className="flex items-center gap-3 text-[#666] text-xs">
                      <span>{alert.client_name} ({alert.client_id})</span>
                      <span>•</span>
                      <span>{new Date(alert.detected_at).toLocaleString()}</span>
                      {alert.resolved_at && (
                        <>
                          <span>•</span>
                          <span className="text-[#10B981]">Resolved by {alert.resolved_by}</span>
                        </>
                      )}
                    </div>
                  </div>
                  {alert.status === 'active' && (
                    <button
                      onClick={() => setResolveAlertModal(alert)}
                      className="bg-[#2980B9] hover:bg-[#3498DB] text-white text-xs px-3 py-1 rounded"
                    >
                      Resolve
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          {/* Login Attempts Table */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded col-span-2">
            <div className="border-b border-[#383A42] p-3">
              <h2 className="text-white font-bold text-sm mb-3 flex items-center gap-2">
                <Activity size={14} className="text-[#3B82F6]" />
                Login Attempts ({filteredAndSortedLogins.length})
              </h2>
              <div className="flex items-center gap-2">
                <div className="relative flex-1">
                  <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-[#666]" />
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={e => setSearchQuery(e.target.value)}
                    placeholder="Search by IP, client, country..."
                    className="w-full bg-[#121316] border border-[#383A42] text-white text-xs pl-8 pr-3 py-1.5 rounded"
                  />
                </div>
                <select
                  value={statusFilter}
                  onChange={e => setStatusFilter(e.target.value)}
                  className="bg-[#121316] border border-[#383A42] text-white text-xs px-2 py-1.5 rounded"
                >
                  <option value="all">All Status</option>
                  <option value="success">Success</option>
                  <option value="failed">Failed</option>
                  <option value="blocked">Blocked</option>
                </select>
                <select
                  value={riskFilter}
                  onChange={e => setRiskFilter(e.target.value)}
                  className="bg-[#121316] border border-[#383A42] text-white text-xs px-2 py-1.5 rounded"
                >
                  <option value="all">All Risk</option>
                  <option value="high">High Risk (70+)</option>
                  <option value="medium">Medium Risk (40-70)</option>
                  <option value="low">Low Risk (&lt;40)</option>
                </select>
                <button
                  onClick={() => setShowBlacklistModal(true)}
                  className="bg-[#DC2626] hover:bg-[#EF4444] text-white text-xs px-3 py-1.5 rounded flex items-center gap-1"
                >
                  <Ban size={12} />
                  Blacklist IP
                </button>
              </div>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="bg-[#121316] border-b border-[#383A42]">
                    <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('timestamp')}>
                      <div className="flex items-center gap-1">
                        Date/Time {sortField === 'timestamp' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                      </div>
                    </th>
                    <th className="text-left text-[#888] font-semibold p-2">Client</th>
                    <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('ip')}>
                      <div className="flex items-center gap-1">
                        IP Address {sortField === 'ip' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                      </div>
                    </th>
                    <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('country')}>
                      <div className="flex items-center gap-1">
                        Location {sortField === 'country' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                      </div>
                    </th>
                    <th className="text-left text-[#888] font-semibold p-2">Device</th>
                    <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('status')}>
                      <div className="flex items-center gap-1">
                        Status {sortField === 'status' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                      </div>
                    </th>
                    <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('risk_score')}>
                      <div className="flex items-center gap-1">
                        Risk Score {sortField === 'risk_score' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                      </div>
                    </th>
                    <th className="text-left text-[#888] font-semibold p-2">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredAndSortedLogins.slice(0, 100).map(login => (
                    <>
                      <tr key={login.id} className="border-b border-[#383A42] hover:bg-[#1E2026]">
                        <td className="p-2 text-[#CCC]">{new Date(login.timestamp).toLocaleString()}</td>
                        <td className="p-2 text-white">{login.client_name}</td>
                        <td className="p-2 text-[#3B82F6] font-mono">{login.ip}</td>
                        <td className="p-2 text-[#CCC]">{login.city}, {login.country}</td>
                        <td className="p-2 text-[#888]">{login.device} / {login.browser}</td>
                        <td className="p-2">
                          <span className={`px-2 py-0.5 rounded text-xs ${getStatusBadge(login.status)}`}>
                            {login.status}
                          </span>
                        </td>
                        <td className="p-2">
                          <span className={`px-2 py-0.5 rounded text-xs ${getRiskBadge(login.risk_score)}`}>
                            {login.risk_score.toFixed(1)}
                          </span>
                        </td>
                        <td className="p-2">
                          <button
                            onClick={() => setExpandedLogin(expandedLogin === login.id ? null : login.id)}
                            className="text-[#3B82F6] hover:text-[#60A5FA]"
                          >
                            <Eye size={14} />
                          </button>
                        </td>
                      </tr>
                      {expandedLogin === login.id && (
                        <tr className="bg-[#121316] border-b border-[#383A42]">
                          <td colSpan={8} className="p-3">
                            <div className="grid grid-cols-3 gap-3 text-xs">
                              <div>
                                <span className="text-[#888]">Client ID:</span>
                                <span className="text-white ml-2">{login.client_id}</span>
                              </div>
                              <div>
                                <span className="text-[#888]">Login ID:</span>
                                <span className="text-white ml-2">{login.id}</span>
                              </div>
                              <div>
                                <span className="text-[#888]">Flags:</span>
                                <span className="text-[#F59E0B] ml-2">{login.flags?.join(', ') || 'None'}</span>
                              </div>
                            </div>
                          </td>
                        </tr>
                      )}
                    </>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Geo Distribution Map */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded">
            <div className="border-b border-[#383A42] p-3">
              <h2 className="text-white font-bold text-sm flex items-center gap-2">
                <Globe size={14} className="text-[#10B981]" />
                Geographic Distribution
              </h2>
            </div>
            <div className="p-3 max-h-96 overflow-y-auto">
              <div className="space-y-2">
                {geoData.slice(0, 15).map(geo => (
                  <div key={geo.country} className="flex items-center justify-between">
                    <div className="flex items-center gap-2 flex-1">
                      <MapPin size={12} className="text-[#666]" />
                      <span className="text-white text-xs">{geo.country}</span>
                      <span className={`text-xs px-1.5 py-0.5 rounded ${
                        geo.risk_level === 'high' ? 'bg-[#DC2626]/20 text-[#DC2626]' :
                        geo.risk_level === 'medium' ? 'bg-[#F59E0B]/20 text-[#F59E0B]' :
                        'bg-[#10B981]/20 text-[#10B981]'
                      }`}>
                        {geo.risk_level}
                      </span>
                    </div>
                    <div className="flex items-center gap-3 text-xs">
                      <span className="text-[#888]">{geo.login_count} logins</span>
                      <span className="text-[#666]">•</span>
                      <span className="text-[#888]">{geo.unique_clients} clients</span>
                      <span className="text-[#666]">•</span>
                      <span className={geo.flagged_count > 0 ? 'text-[#EF4444]' : 'text-[#666]'}>
                        {geo.flagged_count} flagged
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Detection Rules Management */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded">
            <div className="border-b border-[#383A42] p-3">
              <h2 className="text-white font-bold text-sm flex items-center gap-2">
                <Shield size={14} className="text-[#8B5CF6]" />
                Detection Rules ({rules.filter(r => r.is_active).length}/{rules.length} active)
              </h2>
            </div>
            <div className="p-3 max-h-96 overflow-y-auto space-y-2">
              {rules.map(rule => (
                <div key={rule.id} className="bg-[#121316] border border-[#383A42] rounded p-2">
                  <div className="flex items-start justify-between mb-1">
                    <div className="flex-1">
                      <div className="text-white text-xs font-semibold mb-0.5">{rule.name}</div>
                      <div className="text-[#888] text-xs mb-1">{rule.condition}</div>
                      <div className="flex items-center gap-2 text-xs">
                        <span className="text-[#666]">Action: {rule.action}</span>
                        <span className="text-[#666]">•</span>
                        <span className="text-[#666]">Triggered: {rule.trigger_count}x</span>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={rule.is_active}
                        onChange={e => handleToggleRule(rule.id, e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-9 h-5 bg-[#383A42] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#10B981]"></div>
                    </label>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Resolve Alert Modal */}
      {resolveAlertModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4 w-96">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-white font-bold text-sm">Resolve Alert</h3>
              <button onClick={() => setResolveAlertModal(null)} className="text-[#666] hover:text-white">
                <X size={16} />
              </button>
            </div>
            <div className="mb-4">
              <div className="text-[#CCC] text-xs mb-2">{resolveAlertModal.description}</div>
              <div className="text-[#888] text-xs">Client: {resolveAlertModal.client_name}</div>
            </div>
            <div className="mb-4">
              <label className="text-[#888] text-xs mb-2 block">Action</label>
              <select
                value={resolveAction}
                onChange={e => setResolveAction(e.target.value as 'resolved' | 'ignored')}
                className="w-full bg-[#121316] border border-[#383A42] text-white text-xs px-2 py-1.5 rounded"
              >
                <option value="resolved">Mark as Resolved</option>
                <option value="ignored">Mark as Ignored</option>
              </select>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setResolveAlertModal(null)}
                className="flex-1 bg-[#383A42] hover:bg-[#4A4C54] text-white text-xs py-2 rounded"
              >
                Cancel
              </button>
              <button
                onClick={() => handleResolveAlert(resolveAlertModal.id, resolveAction)}
                className="flex-1 bg-[#2980B9] hover:bg-[#3498DB] text-white text-xs py-2 rounded"
              >
                Confirm
              </button>
            </div>
          </div>
        </div>
      )}

      {/* IP Blacklist Modal */}
      {showBlacklistModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4 w-96">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-white font-bold text-sm">Add IP to Blacklist</h3>
              <button onClick={() => setShowBlacklistModal(false)} className="text-[#666] hover:text-white">
                <X size={16} />
              </button>
            </div>
            <div className="mb-4">
              <label className="text-[#888] text-xs mb-2 block">IP Address</label>
              <input
                type="text"
                value={ipToBlacklist}
                onChange={e => setIpToBlacklist(e.target.value)}
                placeholder="192.168.1.1"
                className="w-full bg-[#121316] border border-[#383A42] text-white text-xs px-3 py-2 rounded"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setShowBlacklistModal(false)}
                className="flex-1 bg-[#383A42] hover:bg-[#4A4C54] text-white text-xs py-2 rounded"
              >
                Cancel
              </button>
              <button
                onClick={() => {
                  // TODO: Implement blacklist API call
                  console.log('Blacklist IP:', ipToBlacklist);
                  setShowBlacklistModal(false);
                  setIpToBlacklist('');
                }}
                className="flex-1 bg-[#DC2626] hover:bg-[#EF4444] text-white text-xs py-2 rounded"
              >
                Add to Blacklist
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
