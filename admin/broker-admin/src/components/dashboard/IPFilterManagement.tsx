'use client';

import React, { useState, useMemo } from 'react';
import {
  Shield,
  Plus,
  Upload,
  Search,
  Filter,
  X,
  Trash2,
  CheckCircle,
  XCircle,
  Calendar,
  User,
  AlertCircle,
} from 'lucide-react';

type FilterType = 'whitelist' | 'blacklist';
type FilterStatus = 'active' | 'expired';

interface IPFilterRule {
  id: string;
  ip: string; // IP or CIDR notation
  type: FilterType;
  reason: string;
  createdBy: string;
  createdAt: number;
  expiresAt: number | null; // null = never expires
  status: FilterStatus;
}

export default function IPFilterManagement() {
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<'all' | FilterType>('all');
  const [statusFilter, setStatusFilter] = useState<'all' | FilterStatus>('all');
  const [showAddModal, setShowAddModal] = useState<boolean>(false);
  const [showImportModal, setShowImportModal] = useState<boolean>(false);
  const [checkIP, setCheckIP] = useState<string>('');
  const [checkResult, setCheckResult] = useState<{ allowed: boolean; reason: string } | null>(null);

  // Add Rule Modal State
  const [newIP, setNewIP] = useState<string>('');
  const [newType, setNewType] = useState<FilterType>('whitelist');
  const [newReason, setNewReason] = useState<string>('');
  const [newExpiryType, setNewExpiryType] = useState<'never' | 'custom'>('never');
  const [newExpiryDate, setNewExpiryDate] = useState<string>('');

  // Import Modal State
  const [importJSON, setImportJSON] = useState<string>('');
  const [importPreview, setImportPreview] = useState<any[]>([]);

  // Mock data: 10 IP filter rules
  const [rules, setRules] = useState<IPFilterRule[]>([
    {
      id: 'rule-1',
      ip: '192.168.1.100',
      type: 'whitelist',
      reason: 'Office network - headquarters',
      createdBy: 'admin@broker.com',
      createdAt: Date.now() - 15 * 24 * 60 * 60 * 1000,
      expiresAt: null,
      status: 'active',
    },
    {
      id: 'rule-2',
      ip: '10.0.0.0/24',
      type: 'whitelist',
      reason: 'VPN subnet for remote staff',
      createdBy: 'it-admin@broker.com',
      createdAt: Date.now() - 10 * 24 * 60 * 60 * 1000,
      expiresAt: null,
      status: 'active',
    },
    {
      id: 'rule-3',
      ip: '45.76.123.45',
      type: 'blacklist',
      reason: 'Suspected fraud attempt',
      createdBy: 'security@broker.com',
      createdAt: Date.now() - 5 * 24 * 60 * 60 * 1000,
      expiresAt: Date.now() + 25 * 24 * 60 * 60 * 1000,
      status: 'active',
    },
    {
      id: 'rule-4',
      ip: '203.0.113.0/24',
      type: 'blacklist',
      reason: 'Bot network detected',
      createdBy: 'admin@broker.com',
      createdAt: Date.now() - 3 * 24 * 60 * 60 * 1000,
      expiresAt: null,
      status: 'active',
    },
    {
      id: 'rule-5',
      ip: '172.16.0.1',
      type: 'whitelist',
      reason: 'API integration partner',
      createdBy: 'api-admin@broker.com',
      createdAt: Date.now() - 20 * 24 * 60 * 60 * 1000,
      expiresAt: Date.now() + 60 * 24 * 60 * 60 * 1000,
      status: 'active',
    },
    {
      id: 'rule-6',
      ip: '88.99.123.200',
      type: 'blacklist',
      reason: 'DDoS attack source',
      createdBy: 'security@broker.com',
      createdAt: Date.now() - 30 * 24 * 60 * 60 * 1000,
      expiresAt: Date.now() - 5 * 24 * 60 * 60 * 1000, // expired 5 days ago
      status: 'expired',
    },
    {
      id: 'rule-7',
      ip: '198.51.100.42',
      type: 'whitelist',
      reason: 'Temporary contractor access',
      createdBy: 'hr@broker.com',
      createdAt: Date.now() - 40 * 24 * 60 * 60 * 1000,
      expiresAt: Date.now() - 10 * 24 * 60 * 60 * 1000, // expired 10 days ago
      status: 'expired',
    },
    {
      id: 'rule-8',
      ip: '104.244.42.0/24',
      type: 'blacklist',
      reason: 'Spam originating subnet',
      createdBy: 'admin@broker.com',
      createdAt: Date.now() - 2 * 24 * 60 * 60 * 1000,
      expiresAt: Date.now() + 90 * 24 * 60 * 60 * 1000,
      status: 'active',
    },
    {
      id: 'rule-9',
      ip: '192.0.2.100',
      type: 'whitelist',
      reason: 'Testing environment',
      createdBy: 'dev@broker.com',
      createdAt: Date.now() - 7 * 24 * 60 * 60 * 1000,
      expiresAt: Date.now() + 30 * 24 * 60 * 60 * 1000,
      status: 'active',
    },
    {
      id: 'rule-10',
      ip: '185.220.100.0/22',
      type: 'blacklist',
      reason: 'Known Tor exit nodes',
      createdBy: 'security@broker.com',
      createdAt: Date.now() - 60 * 24 * 60 * 60 * 1000,
      expiresAt: null,
      status: 'active',
    },
  ]);

  // Filtered rules
  const filteredRules = useMemo(() => {
    return rules.filter((rule) => {
      if (searchQuery && !rule.ip.toLowerCase().includes(searchQuery.toLowerCase())) return false;
      if (typeFilter !== 'all' && rule.type !== typeFilter) return false;
      if (statusFilter !== 'all' && rule.status !== statusFilter) return false;
      return true;
    });
  }, [rules, searchQuery, typeFilter, statusFilter]);

  // Stats
  const stats = useMemo(() => {
    const totalRules = rules.length;
    const whitelistCount = rules.filter((r) => r.type === 'whitelist').length;
    const blacklistCount = rules.filter((r) => r.type === 'blacklist').length;
    const blockedToday = Math.floor(Math.random() * 150) + 50; // mock count
    return { totalRules, whitelistCount, blacklistCount, blockedToday };
  }, [rules]);

  const formatDate = (timestamp: number) => {
    return new Date(timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const formatDateTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const validateIP = (ip: string): boolean => {
    // Simple validation for IP or CIDR
    const ipPattern = /^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/;
    return ipPattern.test(ip);
  };

  const handleAddRule = () => {
    if (!newIP || !validateIP(newIP)) {
      alert('Invalid IP or CIDR format');
      return;
    }

    const expiresAt = newExpiryType === 'never'
      ? null
      : new Date(newExpiryDate).getTime();

    const newRule: IPFilterRule = {
      id: `rule-${Date.now()}`,
      ip: newIP,
      type: newType,
      reason: newReason || 'No reason provided',
      createdBy: 'current-user@broker.com',
      createdAt: Date.now(),
      expiresAt,
      status: 'active',
    };

    setRules([...rules, newRule]);
    setShowAddModal(false);
    setNewIP('');
    setNewType('whitelist');
    setNewReason('');
    setNewExpiryType('never');
    setNewExpiryDate('');
  };

  const handleDeleteRule = (id: string) => {
    if (confirm('Are you sure you want to delete this rule?')) {
      setRules(rules.filter((r) => r.id !== id));
    }
  };

  const handleImportPreview = () => {
    try {
      const parsed = JSON.parse(importJSON);
      if (Array.isArray(parsed)) {
        setImportPreview(parsed);
      } else {
        alert('Invalid JSON format. Expected an array.');
      }
    } catch (e) {
      alert('Invalid JSON syntax');
    }
  };

  const handleImport = () => {
    const newRules = importPreview.map((item, index) => ({
      id: `rule-import-${Date.now()}-${index}`,
      ip: item.ip || '',
      type: (item.type as FilterType) || 'blacklist',
      reason: item.reason || 'Imported rule',
      createdBy: 'current-user@broker.com',
      createdAt: Date.now(),
      expiresAt: item.expiresAt ? new Date(item.expiresAt).getTime() : null,
      status: 'active' as FilterStatus,
    }));

    setRules([...rules, ...newRules]);
    setShowImportModal(false);
    setImportJSON('');
    setImportPreview([]);
  };

  const handleCheckIP = () => {
    if (!checkIP || !validateIP(checkIP.split('/')[0])) {
      setCheckResult({ allowed: false, reason: 'Invalid IP format' });
      return;
    }

    // Simple check: if IP is in blacklist, blocked; if in whitelist, allowed
    const blacklisted = rules.find((r) => r.type === 'blacklist' && r.status === 'active' && r.ip === checkIP);
    const whitelisted = rules.find((r) => r.type === 'whitelist' && r.status === 'active' && r.ip === checkIP);

    if (blacklisted) {
      setCheckResult({ allowed: false, reason: `Blocked: ${blacklisted.reason}` });
    } else if (whitelisted) {
      setCheckResult({ allowed: true, reason: `Allowed: ${whitelisted.reason}` });
    } else {
      setCheckResult({ allowed: true, reason: 'No matching rule - default allow' });
    }
  };

  const getTypeBadge = (type: FilterType) => {
    if (type === 'whitelist') {
      return <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/20 text-emerald-400">WHITELIST</span>;
    } else {
      return <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-rose-500/20 text-rose-400">BLACKLIST</span>;
    }
  };

  const getStatusBadge = (status: FilterStatus) => {
    if (status === 'active') {
      return <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-blue-500/20 text-blue-400">ACTIVE</span>;
    } else {
      return <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-zinc-700 text-zinc-500">EXPIRED</span>;
    }
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-zinc-300 overflow-hidden">
      {/* Header */}
      <div className="px-4 py-3 bg-[#1E2026] border-b border-zinc-700 flex items-center justify-between flex-shrink-0">
        <h2 className="text-sm font-bold text-zinc-300 flex items-center gap-2">
          <Shield size={16} className="text-blue-400" />
          IP Access Control
        </h2>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowImportModal(true)}
            className="flex items-center gap-2 px-3 py-1.5 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-xs font-bold rounded transition-colors"
          >
            <Upload size={12} />
            Import
          </button>
          <button
            onClick={() => setShowAddModal(true)}
            className="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded transition-colors"
          >
            <Plus size={12} />
            Add Rule
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-4">
        {/* Stats Cards */}
        <div className="grid grid-cols-4 gap-3">
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="text-[10px] text-zinc-500 font-bold uppercase mb-1">Total Rules</div>
            <div className="text-2xl font-bold text-zinc-300">{stats.totalRules}</div>
          </div>
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="text-[10px] text-zinc-500 font-bold uppercase mb-1">Whitelist</div>
            <div className="text-2xl font-bold text-emerald-400">{stats.whitelistCount}</div>
          </div>
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="text-[10px] text-zinc-500 font-bold uppercase mb-1">Blacklist</div>
            <div className="text-2xl font-bold text-rose-400">{stats.blacklistCount}</div>
          </div>
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="text-[10px] text-zinc-500 font-bold uppercase mb-1">Blocked Today</div>
            <div className="text-2xl font-bold text-yellow-400">{stats.blockedToday}</div>
          </div>
        </div>

        {/* Filter/Search Bar */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="grid grid-cols-3 gap-3">
            {/* Search */}
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase flex items-center gap-1">
                <Search size={10} />
                Search IP
              </label>
              <div className="relative">
                <input
                  type="text"
                  placeholder="e.g., 192.168.1.100"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
                />
                {searchQuery && (
                  <button
                    onClick={() => setSearchQuery('')}
                    className="absolute right-1 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300"
                  >
                    <X size={12} />
                  </button>
                )}
              </div>
            </div>

            {/* Type Filter */}
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase flex items-center gap-1">
                <Filter size={10} />
                Type
              </label>
              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value as 'all' | FilterType)}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              >
                <option value="all">All Types</option>
                <option value="whitelist">Whitelist</option>
                <option value="blacklist">Blacklist</option>
              </select>
            </div>

            {/* Status Filter */}
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase flex items-center gap-1">
                <Filter size={10} />
                Status
              </label>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value as 'all' | FilterStatus)}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              >
                <option value="all">All Status</option>
                <option value="active">Active</option>
                <option value="expired">Expired</option>
              </select>
            </div>
          </div>
        </div>

        {/* Rules Table */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="bg-[#252525] border-b border-zinc-700">
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">IP/CIDR</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Type</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Reason</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Created By</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Created At</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Expires At</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Status</th>
                  <th className="px-3 py-2 text-center text-[10px] font-bold text-zinc-400 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredRules.map((rule) => (
                  <tr
                    key={rule.id}
                    className={`border-b border-zinc-800 hover:bg-[#25272E] transition-colors ${
                      rule.status === 'expired' ? 'opacity-40' : ''
                    }`}
                  >
                    <td className="px-3 py-2 text-blue-400 font-mono">{rule.ip}</td>
                    <td className="px-3 py-2">{getTypeBadge(rule.type)}</td>
                    <td className="px-3 py-2 text-zinc-300 max-w-xs truncate">{rule.reason}</td>
                    <td className="px-3 py-2 text-zinc-400 text-[10px]">{rule.createdBy}</td>
                    <td className="px-3 py-2 text-zinc-400">{formatDate(rule.createdAt)}</td>
                    <td className="px-3 py-2 text-zinc-400">
                      {rule.expiresAt ? formatDate(rule.expiresAt) : <span className="text-zinc-600">Never</span>}
                    </td>
                    <td className="px-3 py-2">{getStatusBadge(rule.status)}</td>
                    <td className="px-3 py-2 text-center">
                      <button
                        onClick={() => handleDeleteRule(rule.id)}
                        className="text-rose-400 hover:text-rose-300 transition-colors"
                        title="Delete rule"
                      >
                        <Trash2 size={14} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {filteredRules.length === 0 && (
            <div className="p-8 text-center text-zinc-500">
              <AlertCircle size={32} className="mx-auto mb-2 opacity-30" />
              <div className="text-sm">No rules found matching your filters</div>
            </div>
          )}
        </div>

        {/* IP Check Tool */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-4">
          <div className="text-xs font-bold text-zinc-400 uppercase mb-3">Quick IP Check</div>
          <div className="flex items-start gap-3">
            <div className="flex-1">
              <input
                type="text"
                placeholder="Enter IP address to check (e.g., 192.168.1.100)"
                value={checkIP}
                onChange={(e) => setCheckIP(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleCheckIP()}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              />
            </div>
            <button
              onClick={handleCheckIP}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded transition-colors"
            >
              Check
            </button>
          </div>

          {checkResult && (
            <div className={`mt-3 p-3 rounded border ${
              checkResult.allowed
                ? 'bg-emerald-500/10 border-emerald-500/30'
                : 'bg-rose-500/10 border-rose-500/30'
            }`}>
              <div className="flex items-center gap-2">
                {checkResult.allowed ? (
                  <CheckCircle size={16} className="text-emerald-400" />
                ) : (
                  <XCircle size={16} className="text-rose-400" />
                )}
                <span className={`text-sm font-bold ${checkResult.allowed ? 'text-emerald-400' : 'text-rose-400'}`}>
                  {checkResult.allowed ? 'ALLOWED' : 'BLOCKED'}
                </span>
              </div>
              <div className="text-xs text-zinc-400 mt-1">{checkResult.reason}</div>
            </div>
          )}
        </div>
      </div>

      {/* Add Rule Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-zinc-700 rounded w-[500px] max-h-[80vh] overflow-y-auto">
            <div className="p-4 border-b border-zinc-700 flex items-center justify-between">
              <h3 className="text-sm font-bold text-zinc-300">Add IP Filter Rule</h3>
              <button onClick={() => setShowAddModal(false)} className="text-zinc-500 hover:text-zinc-300">
                <X size={16} />
              </button>
            </div>

            <div className="p-4 space-y-4">
              {/* IP/CIDR Input */}
              <div className="space-y-1">
                <label className="text-xs font-bold text-zinc-400">IP Address / CIDR</label>
                <input
                  type="text"
                  placeholder="e.g., 192.168.1.100 or 10.0.0.0/24"
                  value={newIP}
                  onChange={(e) => setNewIP(e.target.value)}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
                />
                <div className="text-[10px] text-zinc-500">Single IP (192.168.1.100) or CIDR range (192.168.1.0/24)</div>
              </div>

              {/* Type Radio */}
              <div className="space-y-1">
                <label className="text-xs font-bold text-zinc-400">Type</label>
                <div className="flex gap-4">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      value="whitelist"
                      checked={newType === 'whitelist'}
                      onChange={(e) => setNewType(e.target.value as FilterType)}
                      className="text-blue-500"
                    />
                    <span className="text-xs text-zinc-300">Whitelist (Allow)</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      value="blacklist"
                      checked={newType === 'blacklist'}
                      onChange={(e) => setNewType(e.target.value as FilterType)}
                      className="text-blue-500"
                    />
                    <span className="text-xs text-zinc-300">Blacklist (Block)</span>
                  </label>
                </div>
              </div>

              {/* Reason */}
              <div className="space-y-1">
                <label className="text-xs font-bold text-zinc-400">Reason</label>
                <textarea
                  placeholder="Enter reason for this rule..."
                  value={newReason}
                  onChange={(e) => setNewReason(e.target.value)}
                  rows={3}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-300 focus:outline-none focus:border-blue-500 resize-none"
                />
              </div>

              {/* Expiry */}
              <div className="space-y-1">
                <label className="text-xs font-bold text-zinc-400">Expiration</label>
                <div className="flex gap-4 mb-2">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      value="never"
                      checked={newExpiryType === 'never'}
                      onChange={(e) => setNewExpiryType(e.target.value as 'never' | 'custom')}
                      className="text-blue-500"
                    />
                    <span className="text-xs text-zinc-300">Never</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      value="custom"
                      checked={newExpiryType === 'custom'}
                      onChange={(e) => setNewExpiryType(e.target.value as 'never' | 'custom')}
                      className="text-blue-500"
                    />
                    <span className="text-xs text-zinc-300">Custom Date</span>
                  </label>
                </div>

                {newExpiryType === 'custom' && (
                  <input
                    type="date"
                    value={newExpiryDate}
                    onChange={(e) => setNewExpiryDate(e.target.value)}
                    className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
                  />
                )}
              </div>
            </div>

            <div className="p-4 border-t border-zinc-700 flex justify-end gap-2">
              <button
                onClick={() => setShowAddModal(false)}
                className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-xs font-bold rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleAddRule}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded transition-colors"
              >
                Add Rule
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Bulk Import Modal */}
      {showImportModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-zinc-700 rounded w-[600px] max-h-[80vh] overflow-y-auto">
            <div className="p-4 border-b border-zinc-700 flex items-center justify-between">
              <h3 className="text-sm font-bold text-zinc-300">Bulk Import IP Rules</h3>
              <button onClick={() => setShowImportModal(false)} className="text-zinc-500 hover:text-zinc-300">
                <X size={16} />
              </button>
            </div>

            <div className="p-4 space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-bold text-zinc-400">JSON Input</label>
                <textarea
                  placeholder='[{"ip": "1.2.3.4", "type": "blacklist", "reason": "spam"}]'
                  value={importJSON}
                  onChange={(e) => setImportJSON(e.target.value)}
                  rows={6}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-300 font-mono focus:outline-none focus:border-blue-500 resize-none"
                />
                <div className="text-[10px] text-zinc-500">
                  Format: {`[{"ip": "1.2.3.4", "type": "blacklist", "reason": "spam", "expiresAt": "2026-12-31"}]`}
                </div>
              </div>

              <button
                onClick={handleImportPreview}
                className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-xs font-bold rounded transition-colors"
              >
                Preview
              </button>

              {/* Preview Table */}
              {importPreview.length > 0 && (
                <div className="border border-zinc-700 rounded overflow-hidden">
                  <div className="bg-[#252525] px-3 py-2 text-[10px] font-bold text-zinc-400 uppercase">
                    Preview ({importPreview.length} rules)
                  </div>
                  <div className="max-h-60 overflow-y-auto">
                    <table className="w-full text-xs">
                      <thead className="bg-[#1a1a1c] sticky top-0">
                        <tr>
                          <th className="px-2 py-1 text-left text-[10px] text-zinc-500">IP</th>
                          <th className="px-2 py-1 text-left text-[10px] text-zinc-500">Type</th>
                          <th className="px-2 py-1 text-left text-[10px] text-zinc-500">Reason</th>
                        </tr>
                      </thead>
                      <tbody>
                        {importPreview.map((item, index) => (
                          <tr key={index} className="border-t border-zinc-800">
                            <td className="px-2 py-1 text-blue-400 font-mono">{item.ip}</td>
                            <td className="px-2 py-1">
                              {item.type === 'whitelist' ? (
                                <span className="text-emerald-400 text-[10px]">WHITELIST</span>
                              ) : (
                                <span className="text-rose-400 text-[10px]">BLACKLIST</span>
                              )}
                            </td>
                            <td className="px-2 py-1 text-zinc-400">{item.reason}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </div>

            <div className="p-4 border-t border-zinc-700 flex justify-end gap-2">
              <button
                onClick={() => setShowImportModal(false)}
                className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-xs font-bold rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleImport}
                disabled={importPreview.length === 0}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
              >
                Import {importPreview.length} Rules
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
