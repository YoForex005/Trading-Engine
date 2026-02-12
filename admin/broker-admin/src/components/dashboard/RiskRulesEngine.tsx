'use client';

import React, { useState, useMemo } from 'react';
import { API_CONFIG } from '@/config/api';
import {
  Shield,
  Plus,
  Edit,
  Trash2,
  Play,
  Pause,
  ChevronUp,
  ChevronDown,
  GripVertical,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Mail,
  BellRing,
  Ban,
  X as CloseIcon,
  TrendingDown,
  Users,
} from 'lucide-react';

// ============================================================================
// TYPES
// ============================================================================

type RiskMetric = 'equity' | 'margin_level' | 'daily_loss' | 'position_count' | 'position_size' | 'total_exposure' | 'drawdown';
type RiskOperator = '>' | '<' | '>=' | '<=' | '==';
type RiskAction = 'alert' | 'email' | 'restrict' | 'close_positions' | 'reduce_leverage' | 'notify_manager';
type RuleStatus = 'active' | 'disabled';

interface RiskRule {
  id: string;
  name: string;
  priority: number;
  metric: RiskMetric;
  operator: RiskOperator;
  value: number;
  action: RiskAction;
  status: RuleStatus;
  lastTriggered: number | null;
  hitCount: number;
}

interface RuleTriggerLog {
  id: string;
  timestamp: number;
  ruleName: string;
  accountId: string;
  accountName: string;
  actionTaken: string;
  metricValue: number;
}

// ============================================================================
// CONSTANTS
// ============================================================================

const METRICS: { value: RiskMetric; label: string; unit: string }[] = [
  { value: 'equity', label: 'Account Equity', unit: 'USD' },
  { value: 'margin_level', label: 'Margin Level', unit: '%' },
  { value: 'daily_loss', label: 'Daily Loss', unit: 'USD' },
  { value: 'position_count', label: 'Position Count', unit: '' },
  { value: 'position_size', label: 'Single Position Size', unit: 'lots' },
  { value: 'total_exposure', label: 'Total Exposure', unit: 'USD' },
  { value: 'drawdown', label: 'Drawdown', unit: '%' },
];

const OPERATORS: { value: RiskOperator; label: string }[] = [
  { value: '>', label: 'Greater than (>)' },
  { value: '<', label: 'Less than (<)' },
  { value: '>=', label: 'Greater or equal (>=)' },
  { value: '<=', label: 'Less or equal (<=)' },
  { value: '==', label: 'Equal to (==)' },
];

const ACTIONS: { value: RiskAction; label: string; icon: React.ReactNode; color: string }[] = [
  { value: 'alert', label: 'Send Alert', icon: <BellRing size={14} />, color: 'blue' },
  { value: 'email', label: 'Send Email', icon: <Mail size={14} />, color: 'cyan' },
  { value: 'restrict', label: 'Restrict Trading', icon: <Ban size={14} />, color: 'yellow' },
  { value: 'close_positions', label: 'Force Close Positions', icon: <XCircle size={14} />, color: 'rose' },
  { value: 'reduce_leverage', label: 'Reduce Leverage', icon: <TrendingDown size={14} />, color: 'orange' },
  { value: 'notify_manager', label: 'Notify Risk Manager', icon: <Users size={14} />, color: 'purple' },
];

// ============================================================================
// MOCK DATA
// ============================================================================

const generateMockRules = (): RiskRule[] => [
  {
    id: 'rule-1',
    name: 'Stop Out Protection',
    priority: 1,
    metric: 'margin_level',
    operator: '<',
    value: 30,
    action: 'close_positions',
    status: 'active',
    lastTriggered: Date.now() - 3600000,
    hitCount: 12,
  },
  {
    id: 'rule-2',
    name: 'Margin Call Alert',
    priority: 2,
    metric: 'margin_level',
    operator: '<',
    value: 50,
    action: 'alert',
    status: 'active',
    lastTriggered: Date.now() - 7200000,
    hitCount: 34,
  },
  {
    id: 'rule-3',
    name: 'Daily Loss Limit',
    priority: 3,
    metric: 'daily_loss',
    operator: '>',
    value: 5000,
    action: 'restrict',
    status: 'active',
    lastTriggered: Date.now() - 86400000,
    hitCount: 8,
  },
  {
    id: 'rule-4',
    name: 'Max Position Count',
    priority: 4,
    metric: 'position_count',
    operator: '>',
    value: 20,
    action: 'alert',
    status: 'active',
    lastTriggered: null,
    hitCount: 0,
  },
  {
    id: 'rule-5',
    name: 'Large Position Warning',
    priority: 5,
    metric: 'position_size',
    operator: '>',
    value: 10,
    action: 'email',
    status: 'active',
    lastTriggered: Date.now() - 14400000,
    hitCount: 5,
  },
  {
    id: 'rule-6',
    name: 'High Exposure Alert',
    priority: 6,
    metric: 'total_exposure',
    operator: '>',
    value: 100000,
    action: 'notify_manager',
    status: 'active',
    lastTriggered: Date.now() - 21600000,
    hitCount: 3,
  },
  {
    id: 'rule-7',
    name: 'Drawdown Protection',
    priority: 7,
    metric: 'drawdown',
    operator: '>',
    value: 20,
    action: 'reduce_leverage',
    status: 'active',
    lastTriggered: Date.now() - 172800000,
    hitCount: 2,
  },
  {
    id: 'rule-8',
    name: 'Low Equity Warning',
    priority: 8,
    metric: 'equity',
    operator: '<',
    value: 1000,
    action: 'alert',
    status: 'disabled',
    lastTriggered: Date.now() - 259200000,
    hitCount: 15,
  },
  {
    id: 'rule-9',
    name: 'Critical Margin Level',
    priority: 9,
    metric: 'margin_level',
    operator: '<=',
    value: 20,
    action: 'close_positions',
    status: 'active',
    lastTriggered: null,
    hitCount: 0,
  },
  {
    id: 'rule-10',
    name: 'Excessive Exposure',
    priority: 10,
    metric: 'total_exposure',
    operator: '>=',
    value: 150000,
    action: 'restrict',
    status: 'active',
    lastTriggered: Date.now() - 43200000,
    hitCount: 1,
  },
];

const CLIENTS = [
  { id: '100234', name: 'John Smith' },
  { id: '100567', name: 'Sarah Johnson' },
  { id: '101234', name: 'Michael Brown' },
  { id: '102345', name: 'Emily Davis' },
  { id: '103456', name: 'David Wilson' },
  { id: '104567', name: 'Lisa Anderson' },
];

const generateMockLogs = (rules: RiskRule[]): RuleTriggerLog[] => {
  const logs: RuleTriggerLog[] = [];
  const now = Date.now();

  for (let i = 0; i < 50; i++) {
    const rule = rules[Math.floor(Math.random() * rules.length)];
    const client = CLIENTS[Math.floor(Math.random() * CLIENTS.length)];
    const action = ACTIONS.find((a) => a.value === rule.action);

    logs.push({
      id: `log-${now}-${i}`,
      timestamp: now - Math.floor(Math.random() * 86400000),
      ruleName: rule.name,
      accountId: client.id,
      accountName: client.name,
      actionTaken: action?.label || 'Unknown Action',
      metricValue: parseFloat((Math.random() * 100).toFixed(2)),
    });
  }

  return logs.sort((a, b) => b.timestamp - a.timestamp);
};

// ============================================================================
// MAIN COMPONENT
// ============================================================================

export default function RiskRulesEngine() {
  const [rules, setRules] = useState<RiskRule[]>(generateMockRules());
  const [logs, setLogs] = useState<RuleTriggerLog[]>(generateMockLogs(generateMockRules()));
  const [showModal, setShowModal] = useState(false);
  const [editingRule, setEditingRule] = useState<RiskRule | null>(null);

  // Modal form state
  const [formName, setFormName] = useState('');
  const [formMetric, setFormMetric] = useState<RiskMetric>('margin_level');
  const [formOperator, setFormOperator] = useState<RiskOperator>('>');
  const [formValue, setFormValue] = useState('');
  const [formAction, setFormAction] = useState<RiskAction>('alert');

  // Summary stats
  const stats = useMemo(() => {
    const activeRules = rules.filter((r) => r.status === 'active').length;
    const triggeredToday = logs.filter((log) => log.timestamp > Date.now() - 86400000).length;
    const accountsAffected = new Set(logs.filter((log) => log.timestamp > Date.now() - 86400000).map((log) => log.accountId)).size;
    const autoActions = logs.filter((log) => log.timestamp > Date.now() - 86400000 && !log.actionTaken.includes('Alert') && !log.actionTaken.includes('Email')).length;

    return {
      activeRules,
      triggeredToday,
      accountsAffected,
      autoActions,
    };
  }, [rules, logs]);

  const handleCreateRule = () => {
    setEditingRule(null);
    setFormName('');
    setFormMetric('margin_level');
    setFormOperator('>');
    setFormValue('');
    setFormAction('alert');
    setShowModal(true);
  };

  const handleEditRule = (rule: RiskRule) => {
    setEditingRule(rule);
    setFormName(rule.name);
    setFormMetric(rule.metric);
    setFormOperator(rule.operator);
    setFormValue(rule.value.toString());
    setFormAction(rule.action);
    setShowModal(true);
  };

  const handleSaveRule = () => {
    if (!formName || !formValue) {
      alert('Please fill all fields');
      return;
    }

    if (editingRule) {
      // Update existing rule
      setRules((prev) =>
        prev.map((r) =>
          r.id === editingRule.id
            ? {
                ...r,
                name: formName,
                metric: formMetric,
                operator: formOperator,
                value: parseFloat(formValue),
                action: formAction,
              }
            : r
        )
      );
    } else {
      // Create new rule
      const newRule: RiskRule = {
        id: `rule-${Date.now()}`,
        name: formName,
        priority: rules.length + 1,
        metric: formMetric,
        operator: formOperator,
        value: parseFloat(formValue),
        action: formAction,
        status: 'active',
        lastTriggered: null,
        hitCount: 0,
      };
      setRules((prev) => [...prev, newRule]);
    }

    setShowModal(false);
  };

  const handleDeleteRule = (ruleId: string) => {
    if (confirm('Delete this rule?')) {
      setRules((prev) => prev.filter((r) => r.id !== ruleId));
    }
  };

  const handleToggleStatus = (ruleId: string) => {
    setRules((prev) =>
      prev.map((r) =>
        r.id === ruleId ? { ...r, status: r.status === 'active' ? 'disabled' : 'active' } : r
      )
    );
  };

  const handleMovePriority = (ruleId: string, direction: 'up' | 'down') => {
    setRules((prev) => {
      const index = prev.findIndex((r) => r.id === ruleId);
      if (index === -1) return prev;

      const newRules = [...prev];
      const targetIndex = direction === 'up' ? index - 1 : index + 1;

      if (targetIndex < 0 || targetIndex >= newRules.length) return prev;

      [newRules[index], newRules[targetIndex]] = [newRules[targetIndex], newRules[index]];

      return newRules.map((r, i) => ({ ...r, priority: i + 1 }));
    });
  };

  const getActionColor = (action: RiskAction) => {
    const actionObj = ACTIONS.find((a) => a.value === action);
    return actionObj?.color || 'gray';
  };

  const getActionIcon = (action: RiskAction) => {
    const actionObj = ACTIONS.find((a) => a.value === action);
    return actionObj?.icon || null;
  };

  const formatCondition = (rule: RiskRule) => {
    const metric = METRICS.find((m) => m.value === rule.metric);
    return `${metric?.label} ${rule.operator} ${rule.value}${metric?.unit ? ' ' + metric.unit : ''}`;
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-zinc-300 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-[#1E2026] border-b border-zinc-700">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-bold text-zinc-100">Risk Management Rules Engine</h1>
        </div>
        <button
          onClick={handleCreateRule}
          className="px-3 py-1.5 text-xs font-bold bg-blue-500/20 text-blue-400 border border-blue-500/50 rounded hover:bg-blue-500/30 flex items-center gap-1"
        >
          <Plus size={14} />
          Create Rule
        </button>
      </div>

      {/* Main Content - Scrollable */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Summary Cards */}
        <div className="grid grid-cols-4 gap-3">
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <CheckCircle size={14} className="text-emerald-400" />
              <div className="text-xs text-zinc-500">Active Rules</div>
            </div>
            <div className="text-2xl font-bold text-emerald-400">{stats.activeRules}</div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <AlertTriangle size={14} className="text-yellow-400" />
              <div className="text-xs text-zinc-500">Rules Triggered Today</div>
            </div>
            <div className="text-2xl font-bold text-yellow-400">{stats.triggeredToday}</div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <Users size={14} className="text-blue-400" />
              <div className="text-xs text-zinc-500">Accounts Affected</div>
            </div>
            <div className="text-2xl font-bold text-blue-400">{stats.accountsAffected}</div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <Shield size={14} className="text-rose-400" />
              <div className="text-xs text-zinc-500">Auto-Actions Taken</div>
            </div>
            <div className="text-2xl font-bold text-rose-400">{stats.autoActions}</div>
          </div>
        </div>

        {/* Rules List */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="px-4 py-3 bg-zinc-800/50 border-b border-zinc-700">
            <h2 className="text-sm font-bold text-zinc-100">Risk Rules ({rules.length})</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-zinc-800/30 border-b border-zinc-700">
                <tr>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400 w-8">#</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Name</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Condition</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Action</th>
                  <th className="px-3 py-2 text-center font-bold text-zinc-400">Status</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Last Triggered</th>
                  <th className="px-3 py-2 text-right font-bold text-zinc-400">Hit Count</th>
                  <th className="px-3 py-2 text-center font-bold text-zinc-400">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-700/50">
                {rules.map((rule, index) => (
                  <tr key={rule.id} className="hover:bg-zinc-800/30">
                    <td className="px-3 py-2 text-zinc-500">
                      <div className="flex items-center gap-1">
                        <GripVertical size={12} className="text-zinc-600 cursor-move" />
                        {rule.priority}
                      </div>
                    </td>
                    <td className="px-3 py-2 text-zinc-300 font-bold">{rule.name}</td>
                    <td className="px-3 py-2 text-zinc-400 font-mono text-xs">
                      {formatCondition(rule)}
                    </td>
                    <td className="px-3 py-2">
                      <span
                        className={`px-2 py-1 text-xs font-bold rounded border flex items-center gap-1 w-fit bg-${getActionColor(rule.action)}-500/20 text-${getActionColor(rule.action)}-400 border-${getActionColor(rule.action)}-500/50`}
                      >
                        {getActionIcon(rule.action)}
                        {ACTIONS.find((a) => a.value === rule.action)?.label}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-center">
                      {rule.status === 'active' ? (
                        <span className="px-2 py-1 text-xs font-bold bg-emerald-500/20 text-emerald-400 border border-emerald-500/50 rounded">
                          Active
                        </span>
                      ) : (
                        <span className="px-2 py-1 text-xs font-bold bg-zinc-700/50 text-zinc-500 border border-zinc-600 rounded">
                          Disabled
                        </span>
                      )}
                    </td>
                    <td className="px-3 py-2 text-zinc-500">
                      {rule.lastTriggered
                        ? new Date(rule.lastTriggered).toLocaleString()
                        : 'Never'}
                    </td>
                    <td className="px-3 py-2 text-right font-bold text-zinc-300">{rule.hitCount}</td>
                    <td className="px-3 py-2">
                      <div className="flex items-center justify-center gap-1">
                        <button
                          onClick={() => handleMovePriority(rule.id, 'up')}
                          disabled={index === 0}
                          className="p-1 text-zinc-500 hover:text-zinc-300 disabled:opacity-30 disabled:cursor-not-allowed"
                          title="Move up"
                        >
                          <ChevronUp size={14} />
                        </button>
                        <button
                          onClick={() => handleMovePriority(rule.id, 'down')}
                          disabled={index === rules.length - 1}
                          className="p-1 text-zinc-500 hover:text-zinc-300 disabled:opacity-30 disabled:cursor-not-allowed"
                          title="Move down"
                        >
                          <ChevronDown size={14} />
                        </button>
                        <button
                          onClick={() => handleToggleStatus(rule.id)}
                          className="p-1 text-zinc-500 hover:text-yellow-400"
                          title={rule.status === 'active' ? 'Disable' : 'Enable'}
                        >
                          {rule.status === 'active' ? <Pause size={14} /> : <Play size={14} />}
                        </button>
                        <button
                          onClick={() => handleEditRule(rule)}
                          className="p-1 text-zinc-500 hover:text-blue-400"
                          title="Edit"
                        >
                          <Edit size={14} />
                        </button>
                        <button
                          onClick={() => handleDeleteRule(rule.id)}
                          className="p-1 text-zinc-500 hover:text-rose-400"
                          title="Delete"
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Rule History Log */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="px-4 py-3 bg-zinc-800/50 border-b border-zinc-700">
            <h2 className="text-sm font-bold text-zinc-100">Rule Trigger History (Last 50)</h2>
          </div>
          <div className="overflow-x-auto" style={{ maxHeight: '400px' }}>
            <table className="w-full text-xs">
              <thead className="bg-zinc-800/30 border-b border-zinc-700 sticky top-0">
                <tr>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Timestamp</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Rule Name</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Account</th>
                  <th className="px-3 py-2 text-left font-bold text-zinc-400">Action Taken</th>
                  <th className="px-3 py-2 text-right font-bold text-zinc-400">Metric Value</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-700/50">
                {logs.map((log) => (
                  <tr key={log.id} className="hover:bg-zinc-800/30">
                    <td className="px-3 py-2 text-zinc-500">
                      {new Date(log.timestamp).toLocaleString()}
                    </td>
                    <td className="px-3 py-2 text-zinc-300 font-bold">{log.ruleName}</td>
                    <td className="px-3 py-2 text-zinc-400">
                      {log.accountName}
                      <div className="text-xs text-zinc-600">{log.accountId}</div>
                    </td>
                    <td className="px-3 py-2 text-zinc-400">{log.actionTaken}</td>
                    <td className="px-3 py-2 text-right font-mono text-zinc-300">
                      {log.metricValue.toFixed(2)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Create/Edit Rule Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-zinc-700 rounded w-full max-w-md p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-zinc-100">
                {editingRule ? 'Edit Rule' : 'Create New Rule'}
              </h3>
              <button
                onClick={() => setShowModal(false)}
                className="text-zinc-500 hover:text-zinc-300"
              >
                <CloseIcon size={20} />
              </button>
            </div>

            <div className="space-y-3">
              {/* Rule Name */}
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Rule Name</label>
                <input
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  className="w-full px-3 py-2 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                  placeholder="e.g., Margin Call Alert"
                />
              </div>

              {/* Condition Builder */}
              <div className="bg-zinc-800/50 border border-zinc-700 rounded p-3">
                <div className="text-xs text-zinc-400 mb-2 font-bold">IF</div>

                <div className="space-y-2">
                  {/* Metric */}
                  <select
                    value={formMetric}
                    onChange={(e) => setFormMetric(e.target.value as RiskMetric)}
                    className="w-full px-2 py-1.5 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                  >
                    {METRICS.map((m) => (
                      <option key={m.value} value={m.value}>
                        {m.label}
                      </option>
                    ))}
                  </select>

                  {/* Operator */}
                  <select
                    value={formOperator}
                    onChange={(e) => setFormOperator(e.target.value as RiskOperator)}
                    className="w-full px-2 py-1.5 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                  >
                    {OPERATORS.map((op) => (
                      <option key={op.value} value={op.value}>
                        {op.label}
                      </option>
                    ))}
                  </select>

                  {/* Value */}
                  <input
                    type="number"
                    value={formValue}
                    onChange={(e) => setFormValue(e.target.value)}
                    className="w-full px-3 py-1.5 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                    placeholder="Value"
                  />
                </div>

                <div className="text-xs text-zinc-400 mt-2 mb-2 font-bold">THEN</div>

                {/* Action */}
                <select
                  value={formAction}
                  onChange={(e) => setFormAction(e.target.value as RiskAction)}
                  className="w-full px-2 py-1.5 text-sm bg-zinc-800 border border-zinc-600 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                >
                  {ACTIONS.map((action) => (
                    <option key={action.value} value={action.value}>
                      {action.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center justify-end gap-2 mt-4">
              <button
                onClick={() => setShowModal(false)}
                className="px-3 py-1.5 text-xs font-bold bg-zinc-700 text-zinc-300 rounded hover:bg-zinc-600"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveRule}
                className="px-3 py-1.5 text-xs font-bold bg-blue-500/20 text-blue-400 border border-blue-500/50 rounded hover:bg-blue-500/30"
              >
                {editingRule ? 'Update Rule' : 'Create Rule'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
