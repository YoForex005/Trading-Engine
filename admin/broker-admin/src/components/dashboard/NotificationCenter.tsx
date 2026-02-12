'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Bell,
  TrendingUp,
  TrendingDown,
  Mail,
  MessageSquare,
  Smartphone,
  Globe,
  AlertCircle,
  CheckCircle,
  Info,
  AlertTriangle,
  XCircle,
  Filter,
  Plus,
  Edit2,
  Power,
  Send,
  ChevronDown,
  ChevronUp,
  X,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// ============================================================================
// TYPE DEFINITIONS
// ============================================================================

type NotificationType = 'trade' | 'risk' | 'system' | 'client' | 'payment' | 'kyc' | 'withdrawal' | 'deposit';
type Severity = 'info' | 'success' | 'warning' | 'error' | 'critical';
type Channel = 'in_app' | 'email' | 'sms' | 'push' | 'webhook';
type NotificationStatus = 'read' | 'unread';

interface Notification {
  id: string;
  type: NotificationType;
  severity: Severity;
  title: string;
  message: string;
  fullMessage?: string;
  channel: Channel[];
  status: NotificationStatus;
  timestamp: string;
  clientId?: string;
  recipientCount?: number;
}

interface NotificationStats {
  totalSent: number;
  totalSentTrend: number; // percentage
  readRate: number; // percentage
  readRateTrend: number;
  unreadCount: number;
  unreadCountTrend: number;
  alertsCount: number;
  alertsCountTrend: number;
}

interface DeliveryBreakdown {
  channel: Channel;
  count: number;
  percentage: number;
}

interface SeverityDistribution {
  severity: Severity;
  count: number;
}

interface NotificationRule {
  id: string;
  name: string;
  triggerEvent: string;
  channels: Channel[];
  template: string;
  conditions: string;
  status: 'active' | 'inactive';
  triggerCount: number;
}

// ============================================================================
// MAIN COMPONENT
// ============================================================================

interface NotificationCenterProps {
  onNavigate?: (viewId: string) => void;
}

export default function NotificationCenter({ onNavigate }: NotificationCenterProps = {}) {
  const [loading, setLoading] = useState(true);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [stats, setStats] = useState<NotificationStats | null>(null);
  const [deliveryBreakdown, setDeliveryBreakdown] = useState<DeliveryBreakdown[]>([]);
  const [severityDistribution, setSeverityDistribution] = useState<SeverityDistribution[]>([]);
  const [rules, setRules] = useState<NotificationRule[]>([]);

  // UI state
  const [expandedNotificationId, setExpandedNotificationId] = useState<string | null>(null);
  const [filterType, setFilterType] = useState<NotificationType | 'all'>('all');
  const [filterSeverity, setFilterSeverity] = useState<Severity | 'all'>('all');
  const [filterStatus, setFilterStatus] = useState<NotificationStatus | 'all'>('all');
  const [filterChannel, setFilterChannel] = useState<Channel | 'all'>('all');
  const [showCreateRuleModal, setShowCreateRuleModal] = useState(false);
  const [showSendNotificationPanel, setShowSendNotificationPanel] = useState(false);

  // Form state for create rule modal
  const [newRule, setNewRule] = useState({
    name: '',
    triggerEvent: '',
    channels: [] as Channel[],
    template: '',
    conditions: '',
  });

  // Form state for send notification panel
  const [sendForm, setSendForm] = useState({
    recipientType: 'single' as 'single' | 'broadcast',
    recipientId: '',
    type: 'system' as NotificationType,
    severity: 'info' as Severity,
    title: '',
    message: '',
    channels: ['in_app'] as Channel[],
  });

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

      const [notifRes, statsRes, rulesRes] = await Promise.all([
        fetch(API_CONFIG.NOTIFICATIONS_LIST, { headers }),
        fetch(API_CONFIG.NOTIFICATIONS_STATS, { headers }),
        fetch(API_CONFIG.NOTIFICATIONS_RULES, { headers }),
      ]);

      if (notifRes.ok) setNotifications(await notifRes.json());
      if (statsRes.ok) {
        const statsData = await statsRes.json();
        setStats(statsData.stats);
        setDeliveryBreakdown(statsData.deliveryBreakdown);
        setSeverityDistribution(statsData.severityDistribution);
      }
      if (rulesRes.ok) setRules(await rulesRes.json());

      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch notification data:', error);
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAllData();
  }, [fetchAllData]);

  // ============================================================================
  // HANDLERS
  // ============================================================================

  const handleToggleExpand = (id: string) => {
    setExpandedNotificationId(expandedNotificationId === id ? null : id);
  };

  const handleToggleRule = async (ruleId: string, currentStatus: 'active' | 'inactive') => {
    const newStatus = currentStatus === 'active' ? 'inactive' : 'active';
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('rtx_admin_token') : null;
      await fetch(API_CONFIG.NOTIFICATIONS_RULE_UPDATE(ruleId), {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : '',
        },
        body: JSON.stringify({ status: newStatus }),
      });
      setRules(rules.map(r => r.id === ruleId ? { ...r, status: newStatus } : r));
    } catch (error) {
      console.error('Failed to toggle rule:', error);
    }
  };

  const handleCreateRule = async () => {
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('rtx_admin_token') : null;
      const response = await fetch(API_CONFIG.NOTIFICATIONS_RULES, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : '',
        },
        body: JSON.stringify(newRule),
      });

      if (response.ok) {
        const createdRule = await response.json();
        setRules([...rules, createdRule]);
        setShowCreateRuleModal(false);
        setNewRule({ name: '', triggerEvent: '', channels: [], template: '', conditions: '' });
      }
    } catch (error) {
      console.error('Failed to create rule:', error);
    }
  };

  const handleSendNotification = async () => {
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('rtx_admin_token') : null;
      const endpoint = sendForm.recipientType === 'broadcast'
        ? API_CONFIG.NOTIFICATIONS_BROADCAST
        : API_CONFIG.NOTIFICATIONS_SEND;

      await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : '',
        },
        body: JSON.stringify(sendForm.recipientType === 'broadcast'
          ? { type: sendForm.type, severity: sendForm.severity, title: sendForm.title, message: sendForm.message, channels: sendForm.channels }
          : { ...sendForm, clientId: sendForm.recipientId }
        ),
      });

      setShowSendNotificationPanel(false);
      setSendForm({ recipientType: 'single', recipientId: '', type: 'system', severity: 'info', title: '', message: '', channels: ['in_app'] });
      fetchAllData();
    } catch (error) {
      console.error('Failed to send notification:', error);
    }
  };

  const handleToggleChannel = (channel: Channel, isRule: boolean) => {
    if (isRule) {
      const channels = newRule.channels.includes(channel)
        ? newRule.channels.filter(c => c !== channel)
        : [...newRule.channels, channel];
      setNewRule({ ...newRule, channels });
    } else {
      const channels = sendForm.channels.includes(channel)
        ? sendForm.channels.filter(c => c !== channel)
        : [...sendForm.channels, channel];
      setSendForm({ ...sendForm, channels });
    }
  };

  // ============================================================================
  // FILTERING
  // ============================================================================

  const filteredNotifications = notifications.filter(n => {
    if (filterType !== 'all' && n.type !== filterType) return false;
    if (filterSeverity !== 'all' && n.severity !== filterSeverity) return false;
    if (filterStatus !== 'all' && n.status !== filterStatus) return false;
    if (filterChannel !== 'all' && !n.channel.includes(filterChannel)) return false;
    return true;
  });

  // ============================================================================
  // HELPER FUNCTIONS
  // ============================================================================

  const getSeverityColor = (severity: Severity) => {
    switch (severity) {
      case 'info': return '#3498DB';
      case 'success': return '#2ECC71';
      case 'warning': return '#F39C12';
      case 'error': return '#E74C3C';
      case 'critical': return '#C0392B';
    }
  };

  const getSeverityIcon = (severity: Severity) => {
    switch (severity) {
      case 'info': return <Info size={16} />;
      case 'success': return <CheckCircle size={16} />;
      case 'warning': return <AlertTriangle size={16} />;
      case 'error': return <XCircle size={16} />;
      case 'critical': return <AlertCircle size={16} />;
    }
  };

  const getChannelIcon = (channel: Channel) => {
    switch (channel) {
      case 'in_app': return <Bell size={12} />;
      case 'email': return <Mail size={12} />;
      case 'sms': return <MessageSquare size={12} />;
      case 'push': return <Smartphone size={12} />;
      case 'webhook': return <Globe size={12} />;
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return date.toLocaleDateString();
  };

  // ============================================================================
  // SVG CHART RENDERERS
  // ============================================================================

  const renderDeliveryDonut = () => {
    if (deliveryBreakdown.length === 0) return null;

    const colors = {
      in_app: '#3498DB',
      email: '#9B59B6',
      sms: '#2ECC71',
      push: '#F39C12',
      webhook: '#E74C3C',
    };

    const total = deliveryBreakdown.reduce((sum, d) => sum + d.count, 0);
    let currentAngle = 0;

    return (
      <svg width="200" height="200" viewBox="0 0 200 200">
        {deliveryBreakdown.map((item, idx) => {
          const percentage = (item.count / total) * 100;
          const angle = (percentage / 100) * 360;
          const startAngle = currentAngle;
          const endAngle = startAngle + angle;
          currentAngle = endAngle;

          const startRad = (startAngle - 90) * (Math.PI / 180);
          const endRad = (endAngle - 90) * (Math.PI / 180);

          const x1 = 100 + 70 * Math.cos(startRad);
          const y1 = 100 + 70 * Math.sin(startRad);
          const x2 = 100 + 70 * Math.cos(endRad);
          const y2 = 100 + 70 * Math.sin(endRad);

          const largeArc = angle > 180 ? 1 : 0;

          return (
            <path
              key={idx}
              d={`M 100 100 L ${x1} ${y1} A 70 70 0 ${largeArc} 1 ${x2} ${y2} Z`}
              fill={colors[item.channel]}
              stroke="#121316"
              strokeWidth="2"
            />
          );
        })}
        {/* Center hole */}
        <circle cx="100" cy="100" r="40" fill="#1E2026" />
      </svg>
    );
  };

  const renderSeverityBarChart = () => {
    if (severityDistribution.length === 0) return null;

    const maxCount = Math.max(...severityDistribution.map(s => s.count));
    const barHeight = 30;
    const barSpacing = 45;
    const chartHeight = severityDistribution.length * barSpacing + 20;
    const chartWidth = 400;

    return (
      <svg width={chartWidth} height={chartHeight}>
        {severityDistribution.map((item, idx) => {
          const barWidth = (item.count / maxCount) * 300;
          const y = idx * barSpacing + 10;

          return (
            <g key={idx}>
              <text x="0" y={y + barHeight / 2 + 4} fill="#888" fontSize="12" textAnchor="start">
                {item.severity}
              </text>
              <rect
                x="80"
                y={y}
                width={barWidth}
                height={barHeight}
                fill={getSeverityColor(item.severity)}
                rx="4"
              />
              <text
                x={90 + barWidth}
                y={y + barHeight / 2 + 4}
                fill="#CCC"
                fontSize="12"
                textAnchor="start"
              >
                {item.count}
              </text>
            </g>
          );
        })}
      </svg>
    );
  };

  // ============================================================================
  // RENDER
  // ============================================================================

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-[#888]">Loading notification data...</div>
      </div>
    );
  }

  return (
    <div className="h-full overflow-auto bg-[#121316] p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-bold text-[#F5C542]">Notification Center Management</h2>
        <div className="flex gap-2">
          <button
            onClick={() => setShowCreateRuleModal(true)}
            className="px-3 py-1 bg-[#2980B9] text-white rounded text-xs font-bold flex items-center gap-1"
          >
            <Plus size={14} />
            Create Rule
          </button>
          <button
            onClick={() => setShowSendNotificationPanel(true)}
            className="px-3 py-1 bg-[#2ECC71] text-white rounded text-xs font-bold flex items-center gap-1"
          >
            <Send size={14} />
            Send Notification
          </button>
        </div>
      </div>

      {/* 2. Stats Cards */}
      <div className="grid grid-cols-4 gap-4 mb-4">
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="text-[#888] text-xs mb-1">Total Sent</div>
          <div className="text-2xl font-bold text-white">{stats?.totalSent || 0}</div>
          <div className="flex items-center gap-1 text-xs mt-1">
            {(stats?.totalSentTrend || 0) >= 0 ? (
              <TrendingUp size={12} className="text-[#2ECC71]" />
            ) : (
              <TrendingDown size={12} className="text-[#E74C3C]" />
            )}
            <span className={(stats?.totalSentTrend || 0) >= 0 ? 'text-[#2ECC71]' : 'text-[#E74C3C]'}>
              {Math.abs(stats?.totalSentTrend || 0)}%
            </span>
          </div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="text-[#888] text-xs mb-1">Read Rate</div>
          <div className="text-2xl font-bold text-white">{(stats?.readRate || 0).toFixed(1)}%</div>
          <div className="flex items-center gap-1 text-xs mt-1">
            {(stats?.readRateTrend || 0) >= 0 ? (
              <TrendingUp size={12} className="text-[#2ECC71]" />
            ) : (
              <TrendingDown size={12} className="text-[#E74C3C]" />
            )}
            <span className={(stats?.readRateTrend || 0) >= 0 ? 'text-[#2ECC71]' : 'text-[#E74C3C]'}>
              {Math.abs(stats?.readRateTrend || 0)}%
            </span>
          </div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="text-[#888] text-xs mb-1">Unread Count</div>
          <div className="text-2xl font-bold text-white">{stats?.unreadCount || 0}</div>
          <div className="flex items-center gap-1 text-xs mt-1">
            {(stats?.unreadCountTrend || 0) >= 0 ? (
              <TrendingUp size={12} className="text-[#E74C3C]" />
            ) : (
              <TrendingDown size={12} className="text-[#2ECC71]" />
            )}
            <span className={(stats?.unreadCountTrend || 0) >= 0 ? 'text-[#E74C3C]' : 'text-[#2ECC71]'}>
              {Math.abs(stats?.unreadCountTrend || 0)}%
            </span>
          </div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="text-[#888] text-xs mb-1">Alerts Count</div>
          <div className="text-2xl font-bold text-white">{stats?.alertsCount || 0}</div>
          <div className="flex items-center gap-1 text-xs mt-1">
            {(stats?.alertsCountTrend || 0) >= 0 ? (
              <TrendingUp size={12} className="text-[#E74C3C]" />
            ) : (
              <TrendingDown size={12} className="text-[#2ECC71]" />
            )}
            <span className={(stats?.alertsCountTrend || 0) >= 0 ? 'text-[#E74C3C]' : 'text-[#2ECC71]'}>
              {Math.abs(stats?.alertsCountTrend || 0)}%
            </span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-4">
        {/* 3. Delivery Breakdown Donut Chart */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <h3 className="text-sm font-bold text-[#F5C542] mb-3">Delivery Breakdown by Channel</h3>
          <div className="flex items-center gap-4">
            {renderDeliveryDonut()}
            <div className="space-y-2 text-xs">
              {deliveryBreakdown.map((item, idx) => (
                <div key={idx} className="flex items-center gap-2">
                  <div
                    className="w-3 h-3 rounded"
                    style={{ backgroundColor: { in_app: '#3498DB', email: '#9B59B6', sms: '#2ECC71', push: '#F39C12', webhook: '#E74C3C' }[item.channel] }}
                  />
                  <span className="text-[#CCC]">{item.channel}: {item.count} ({item.percentage.toFixed(1)}%)</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* 4. Severity Distribution Bar Chart */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <h3 className="text-sm font-bold text-[#F5C542] mb-3">Severity Distribution</h3>
          {renderSeverityBarChart()}
        </div>
      </div>

      {/* 5. Notification Rules Table */}
      <div className="bg-[#1E2026] border border-[#383A42] rounded p-4 mb-4">
        <h3 className="text-sm font-bold text-[#F5C542] mb-3">Notification Rules</h3>
        <div className="overflow-auto max-h-64">
          <table className="w-full text-xs">
            <thead className="bg-[#25272E] sticky top-0">
              <tr>
                <th className="text-left px-2 py-2 text-[#888]">Rule Name</th>
                <th className="text-left px-2 py-2 text-[#888]">Trigger Event</th>
                <th className="text-left px-2 py-2 text-[#888]">Channels</th>
                <th className="text-center px-2 py-2 text-[#888]">Status</th>
                <th className="text-right px-2 py-2 text-[#888]">Trigger Count</th>
                <th className="text-center px-2 py-2 text-[#888]">Actions</th>
              </tr>
            </thead>
            <tbody>
              {rules.map((rule, idx) => (
                <tr key={idx} className="border-b border-[#383A42]">
                  <td className="px-2 py-2 text-[#CCC]">{rule.name}</td>
                  <td className="px-2 py-2 text-[#888]">{rule.triggerEvent}</td>
                  <td className="px-2 py-2">
                    <div className="flex gap-1">
                      {rule.channels.map(ch => (
                        <div key={ch} className="text-[#888]">{getChannelIcon(ch)}</div>
                      ))}
                    </div>
                  </td>
                  <td className="px-2 py-2 text-center">
                    <button
                      onClick={() => handleToggleRule(rule.id, rule.status)}
                      className={`px-2 py-1 rounded text-xs font-bold ${
                        rule.status === 'active'
                          ? 'bg-[#2ECC71] text-white'
                          : 'bg-[#888] text-white'
                      }`}
                    >
                      {rule.status}
                    </button>
                  </td>
                  <td className="px-2 py-2 text-right text-[#CCC] font-mono">{rule.triggerCount}</td>
                  <td className="px-2 py-2 text-center">
                    <button className="text-[#3498DB] hover:text-[#2980B9]">
                      <Edit2 size={14} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* 8. Filter Bar */}
      <div className="bg-[#1E2026] border border-[#383A42] rounded p-3 mb-4 flex items-center gap-3">
        <Filter size={14} className="text-[#888]" />
        <select
          value={filterType}
          onChange={(e) => setFilterType(e.target.value as any)}
          className="px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
        >
          <option value="all">All Types</option>
          <option value="trade">Trade</option>
          <option value="risk">Risk</option>
          <option value="system">System</option>
          <option value="client">Client</option>
          <option value="payment">Payment</option>
        </select>
        <select
          value={filterSeverity}
          onChange={(e) => setFilterSeverity(e.target.value as any)}
          className="px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
        >
          <option value="all">All Severities</option>
          <option value="info">Info</option>
          <option value="success">Success</option>
          <option value="warning">Warning</option>
          <option value="error">Error</option>
          <option value="critical">Critical</option>
        </select>
        <select
          value={filterStatus}
          onChange={(e) => setFilterStatus(e.target.value as any)}
          className="px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
        >
          <option value="all">All Status</option>
          <option value="read">Read</option>
          <option value="unread">Unread</option>
        </select>
        <select
          value={filterChannel}
          onChange={(e) => setFilterChannel(e.target.value as any)}
          className="px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
        >
          <option value="all">All Channels</option>
          <option value="in_app">In-App</option>
          <option value="email">Email</option>
          <option value="sms">SMS</option>
          <option value="push">Push</option>
          <option value="webhook">Webhook</option>
        </select>
        <div className="text-xs text-[#888] ml-auto">
          {filteredNotifications.length} notification{filteredNotifications.length !== 1 ? 's' : ''}
        </div>
      </div>

      {/* 1. Notification Feed */}
      <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
        <h3 className="text-sm font-bold text-[#F5C542] mb-3">Notification Feed</h3>
        <div className="space-y-2 max-h-96 overflow-auto">
          {filteredNotifications.map((notif) => (
            <div
              key={notif.id}
              className={`p-3 bg-[#25272E] border rounded cursor-pointer ${
                notif.status === 'unread' ? 'border-[#3498DB]' : 'border-[#383A42]'
              }`}
              onClick={() => handleToggleExpand(notif.id)}
            >
              <div className="flex items-start gap-3">
                <div style={{ color: getSeverityColor(notif.severity) }}>
                  {getSeverityIcon(notif.severity)}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-sm font-bold text-white">{notif.title}</span>
                    <span
                      className="px-2 py-0.5 text-xs rounded"
                      style={{ backgroundColor: getSeverityColor(notif.severity) + '20', color: getSeverityColor(notif.severity) }}
                    >
                      {notif.severity}
                    </span>
                    <div className="flex gap-1">
                      {notif.channel.map(ch => (
                        <div key={ch} className="text-[#888]">{getChannelIcon(ch)}</div>
                      ))}
                    </div>
                    <span className="text-xs text-[#888] ml-auto">{formatTimestamp(notif.timestamp)}</span>
                  </div>
                  <p className="text-xs text-[#CCC] mb-1">
                    {expandedNotificationId === notif.id && notif.fullMessage ? notif.fullMessage : notif.message}
                  </p>
                  {notif.recipientCount && (
                    <div className="text-xs text-[#888]">Sent to {notif.recipientCount} recipient(s)</div>
                  )}
                </div>
                {expandedNotificationId === notif.id ? (
                  <ChevronUp size={16} className="text-[#888]" />
                ) : (
                  <ChevronDown size={16} className="text-[#888]" />
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 6. Create Rule Modal */}
      {showCreateRuleModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-6 w-[500px]">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-bold text-[#F5C542]">Create Notification Rule</h3>
              <button onClick={() => setShowCreateRuleModal(false)} className="text-[#888] hover:text-white">
                <X size={20} />
              </button>
            </div>
            <div className="space-y-3">
              <div>
                <label className="text-xs text-[#888] mb-1 block">Rule Name</label>
                <input
                  type="text"
                  value={newRule.name}
                  onChange={(e) => setNewRule({ ...newRule, name: e.target.value })}
                  className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
                  placeholder="e.g., Large Trade Alert"
                />
              </div>
              <div>
                <label className="text-xs text-[#888] mb-1 block">Trigger Event</label>
                <select
                  value={newRule.triggerEvent}
                  onChange={(e) => setNewRule({ ...newRule, triggerEvent: e.target.value })}
                  className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
                >
                  <option value="">Select event...</option>
                  <option value="trade_executed">Trade Executed</option>
                  <option value="margin_call">Margin Call</option>
                  <option value="withdrawal_request">Withdrawal Request</option>
                  <option value="deposit_received">Deposit Received</option>
                </select>
              </div>
              <div>
                <label className="text-xs text-[#888] mb-1 block">Channels</label>
                <div className="flex gap-2">
                  {(['in_app', 'email', 'sms', 'push', 'webhook'] as Channel[]).map(ch => (
                    <label key={ch} className="flex items-center gap-1 text-xs text-[#CCC]">
                      <input
                        type="checkbox"
                        checked={newRule.channels.includes(ch)}
                        onChange={() => handleToggleChannel(ch, true)}
                      />
                      {ch}
                    </label>
                  ))}
                </div>
              </div>
              <div>
                <label className="text-xs text-[#888] mb-1 block">Template</label>
                <textarea
                  value={newRule.template}
                  onChange={(e) => setNewRule({ ...newRule, template: e.target.value })}
                  className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
                  rows={3}
                  placeholder="Message template..."
                />
              </div>
              <div>
                <label className="text-xs text-[#888] mb-1 block">Conditions (JSON)</label>
                <textarea
                  value={newRule.conditions}
                  onChange={(e) => setNewRule({ ...newRule, conditions: e.target.value })}
                  className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42] font-mono"
                  rows={2}
                  placeholder='{"amount_gte": 10000}'
                />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button
                onClick={() => setShowCreateRuleModal(false)}
                className="px-3 py-1.5 bg-[#383A42] text-white text-xs rounded"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateRule}
                className="px-3 py-1.5 bg-[#2980B9] text-white text-xs rounded font-bold"
              >
                Create Rule
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 7. Send Notification Panel */}
      {showSendNotificationPanel && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-6 w-[500px]">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-bold text-[#F5C542]">Send Notification</h3>
              <button onClick={() => setShowSendNotificationPanel(false)} className="text-[#888] hover:text-white">
                <X size={20} />
              </button>
            </div>
            <div className="space-y-3">
              <div>
                <label className="text-xs text-[#888] mb-1 block">Recipient Type</label>
                <select
                  value={sendForm.recipientType}
                  onChange={(e) => setSendForm({ ...sendForm, recipientType: e.target.value as any })}
                  className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
                >
                  <option value="single">Single Client</option>
                  <option value="broadcast">Broadcast to All</option>
                </select>
              </div>
              {sendForm.recipientType === 'single' && (
                <div>
                  <label className="text-xs text-[#888] mb-1 block">Client ID</label>
                  <input
                    type="text"
                    value={sendForm.recipientId}
                    onChange={(e) => setSendForm({ ...sendForm, recipientId: e.target.value })}
                    className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
                    placeholder="e.g., 12345"
                  />
                </div>
              )}
              <div>
                <label className="text-xs text-[#888] mb-1 block">Type</label>
                <select
                  value={sendForm.type}
                  onChange={(e) => setSendForm({ ...sendForm, type: e.target.value as any })}
                  className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
                >
                  <option value="system">System</option>
                  <option value="trade">Trade</option>
                  <option value="risk">Risk</option>
                  <option value="payment">Payment</option>
                </select>
              </div>
              <div>
                <label className="text-xs text-[#888] mb-1 block">Severity</label>
                <select
                  value={sendForm.severity}
                  onChange={(e) => setSendForm({ ...sendForm, severity: e.target.value as any })}
                  className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
                >
                  <option value="info">Info</option>
                  <option value="success">Success</option>
                  <option value="warning">Warning</option>
                  <option value="error">Error</option>
                  <option value="critical">Critical</option>
                </select>
              </div>
              <div>
                <label className="text-xs text-[#888] mb-1 block">Title</label>
                <input
                  type="text"
                  value={sendForm.title}
                  onChange={(e) => setSendForm({ ...sendForm, title: e.target.value })}
                  className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
                  placeholder="Notification title"
                />
              </div>
              <div>
                <label className="text-xs text-[#888] mb-1 block">Message</label>
                <textarea
                  value={sendForm.message}
                  onChange={(e) => setSendForm({ ...sendForm, message: e.target.value })}
                  className="w-full px-2 py-1 bg-[#25272E] text-[#CCC] text-xs rounded border border-[#383A42]"
                  rows={4}
                  placeholder="Notification message"
                />
              </div>
              <div>
                <label className="text-xs text-[#888] mb-1 block">Channels</label>
                <div className="flex gap-2">
                  {(['in_app', 'email', 'sms', 'push', 'webhook'] as Channel[]).map(ch => (
                    <label key={ch} className="flex items-center gap-1 text-xs text-[#CCC]">
                      <input
                        type="checkbox"
                        checked={sendForm.channels.includes(ch)}
                        onChange={() => handleToggleChannel(ch, false)}
                      />
                      {ch}
                    </label>
                  ))}
                </div>
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button
                onClick={() => setShowSendNotificationPanel(false)}
                className="px-3 py-1.5 bg-[#383A42] text-white text-xs rounded"
              >
                Cancel
              </button>
              <button
                onClick={handleSendNotification}
                className="px-3 py-1.5 bg-[#2ECC71] text-white text-xs rounded font-bold flex items-center gap-1"
              >
                <Send size={14} />
                Send
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
