'use client';

import React, { useState, useMemo } from 'react';
import {
  Gift,
  Plus,
  Calendar,
  DollarSign,
  Users,
  TrendingUp,
  X,
  CheckCircle,
  Clock,
  XCircle,
  Pause,
  Eye,
  Edit,
  Trash2
} from 'lucide-react';

type PromotionType = 'deposit_match' | 'no_deposit' | 'cashback' | 'rebate' | 'loyalty';
type PromotionStatus = 'active' | 'scheduled' | 'expired' | 'paused';
type ValueType = 'percentage' | 'fixed';

interface Promotion {
  id: string;
  name: string;
  type: PromotionType;
  valueType: ValueType;
  value: number;
  minDeposit?: number;
  maxBonus?: number;
  eligibleClients: number;
  claimedCount: number;
  totalCost: number;
  startDate: string;
  endDate: string;
  status: PromotionStatus;
  conversionRate: number;
  eligibleAccountTypes: string[];
  terms: string;
}

interface Client {
  accountId: string;
  name: string;
  email: string;
  claimedDate: string;
  bonusAmount: number;
  deposited: number;
}

export default function PromotionManager() {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [selectedPromotion, setSelectedPromotion] = useState<Promotion | null>(null);
  const [statusFilter, setStatusFilter] = useState<PromotionStatus | 'all'>('all');

  const [newPromotion, setNewPromotion] = useState({
    name: '',
    type: 'deposit_match' as PromotionType,
    valueType: 'percentage' as ValueType,
    value: 0,
    minDeposit: 0,
    maxBonus: 0,
    eligibleAccountTypes: [] as string[],
    startDate: '',
    endDate: '',
    terms: ''
  });

  // Mock promotions
  const allPromotions: Promotion[] = useMemo(() => [
    {
      id: 'PROMO-001',
      name: 'Welcome Bonus',
      type: 'deposit_match',
      valueType: 'percentage',
      value: 30,
      minDeposit: 100,
      maxBonus: 500,
      eligibleClients: 1240,
      claimedCount: 187,
      totalCost: 12450,
      startDate: '2026-01-01',
      endDate: '2026-03-31',
      status: 'active',
      conversionRate: 42,
      eligibleAccountTypes: ['Standard', 'Pro'],
      terms: 'First deposit only. Minimum deposit $100. Maximum bonus $500.'
    },
    {
      id: 'PROMO-002',
      name: 'Deposit Match 50%',
      type: 'deposit_match',
      valueType: 'percentage',
      value: 50,
      minDeposit: 500,
      maxBonus: 2000,
      eligibleClients: 856,
      claimedCount: 143,
      totalCost: 18900,
      startDate: '2026-02-01',
      endDate: '2026-02-28',
      status: 'active',
      conversionRate: 38,
      eligibleAccountTypes: ['Pro', 'VIP'],
      terms: 'Valid for deposits above $500. Maximum bonus $2000.'
    },
    {
      id: 'PROMO-003',
      name: 'No-Deposit Bonus',
      type: 'no_deposit',
      valueType: 'fixed',
      value: 25,
      eligibleClients: 3420,
      claimedCount: 892,
      totalCost: 22300,
      startDate: '2026-01-15',
      endDate: '2026-12-31',
      status: 'active',
      conversionRate: 18,
      eligibleAccountTypes: ['Standard'],
      terms: 'Free $25 bonus for new account registration. No deposit required.'
    },
    {
      id: 'PROMO-004',
      name: 'Loyalty Bonus',
      type: 'loyalty',
      valueType: 'percentage',
      value: 10,
      minDeposit: 1000,
      maxBonus: 1000,
      eligibleClients: 445,
      claimedCount: 67,
      totalCost: 8920,
      startDate: '2026-01-01',
      endDate: '2026-12-31',
      status: 'active',
      conversionRate: 51,
      eligibleAccountTypes: ['VIP'],
      terms: 'Monthly loyalty bonus for VIP accounts with $1000+ monthly volume.'
    },
    {
      id: 'PROMO-005',
      name: 'Referral Bonus',
      type: 'rebate',
      valueType: 'fixed',
      value: 100,
      eligibleClients: 2100,
      claimedCount: 234,
      totalCost: 23400,
      startDate: '2025-12-01',
      endDate: '2026-06-30',
      status: 'active',
      conversionRate: 28,
      eligibleAccountTypes: ['Standard', 'Pro', 'VIP'],
      terms: 'Earn $100 for each referred friend who deposits $500+.'
    },
    {
      id: 'PROMO-006',
      name: 'VIP Cashback',
      type: 'cashback',
      valueType: 'percentage',
      value: 5,
      minDeposit: 5000,
      eligibleClients: 189,
      claimedCount: 102,
      totalCost: 14567,
      startDate: '2026-02-01',
      endDate: '2026-12-31',
      status: 'active',
      conversionRate: 62,
      eligibleAccountTypes: ['VIP'],
      terms: '5% cashback on all trading losses for VIP accounts.'
    },
    {
      id: 'PROMO-007',
      name: 'Summer Special',
      type: 'deposit_match',
      valueType: 'percentage',
      value: 100,
      minDeposit: 200,
      maxBonus: 1000,
      eligibleClients: 0,
      claimedCount: 0,
      totalCost: 0,
      startDate: '2026-06-01',
      endDate: '2026-08-31',
      status: 'scheduled',
      conversionRate: 0,
      eligibleAccountTypes: ['Standard', 'Pro'],
      terms: 'Summer promotion: 100% deposit match up to $1000.'
    },
    {
      id: 'PROMO-008',
      name: 'Black Friday Deal',
      type: 'deposit_match',
      valueType: 'percentage',
      value: 75,
      minDeposit: 300,
      maxBonus: 1500,
      eligibleClients: 1567,
      claimedCount: 423,
      totalCost: 28940,
      startDate: '2025-11-25',
      endDate: '2025-11-30',
      status: 'expired',
      conversionRate: 45,
      eligibleAccountTypes: ['Standard', 'Pro'],
      terms: 'Limited time Black Friday offer.'
    },
    {
      id: 'PROMO-009',
      name: 'Holiday Bonus',
      type: 'no_deposit',
      valueType: 'fixed',
      value: 50,
      eligibleClients: 2890,
      claimedCount: 1234,
      totalCost: 61700,
      startDate: '2025-12-20',
      endDate: '2026-01-05',
      status: 'expired',
      conversionRate: 22,
      eligibleAccountTypes: ['Standard', 'Pro'],
      terms: 'Holiday season gift - $50 no deposit bonus.'
    },
    {
      id: 'PROMO-010',
      name: 'VIP Exclusive',
      type: 'cashback',
      valueType: 'percentage',
      value: 10,
      eligibleClients: 67,
      claimedCount: 34,
      totalCost: 8920,
      startDate: '2026-01-01',
      endDate: '2026-03-31',
      status: 'paused',
      conversionRate: 78,
      eligibleAccountTypes: ['VIP'],
      terms: 'Temporarily paused. 10% cashback for top VIP clients.'
    }
  ], []);

  // Mock claimed clients for detail view
  const mockClaimedClients: Client[] = useMemo(() => {
    if (!selectedPromotion) return [];
    return Array.from({ length: Math.min(selectedPromotion.claimedCount, 20) }, (_, i) => ({
      accountId: `ACC-${10000 + i}`,
      name: `Client ${i + 1}`,
      email: `client${i + 1}@example.com`,
      claimedDate: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
      bonusAmount: selectedPromotion.valueType === 'percentage'
        ? Math.round(Math.random() * (selectedPromotion.maxBonus || 500))
        : selectedPromotion.value,
      deposited: Math.round(Math.random() * 5000 + 100)
    }));
  }, [selectedPromotion]);

  // Calculate stats
  const stats = useMemo(() => {
    const activePromos = allPromotions.filter(p => p.status === 'active').length;
    const totalBonusPaid = allPromotions.reduce((sum, p) => sum + p.totalCost, 0);
    const avgConversion = allPromotions.filter(p => p.status === 'active').reduce((sum, p, _, arr) =>
      sum + p.conversionRate / arr.length, 0);
    const roiEstimate = 2.3;

    return {
      activePromotions: activePromos,
      totalBonusPaid: totalBonusPaid.toFixed(0),
      avgConversion: avgConversion.toFixed(0),
      roiEstimate: roiEstimate.toFixed(1)
    };
  }, [allPromotions]);

  // Monthly payout data for chart (last 6 months)
  const monthlyPayouts = useMemo(() => {
    return [
      { month: 'Sep', amount: 38420 },
      { month: 'Oct', amount: 42190 },
      { month: 'Nov', amount: 51230 },
      { month: 'Dec', amount: 67890 },
      { month: 'Jan', amount: 45670 },
      { month: 'Feb', amount: 39120 }
    ];
  }, []);

  const maxPayout = Math.max(...monthlyPayouts.map(m => m.amount));

  // Filter promotions
  const filteredPromotions = useMemo(() => {
    return allPromotions.filter(promo =>
      statusFilter === 'all' || promo.status === statusFilter
    );
  }, [allPromotions, statusFilter]);

  const getStatusBadge = (status: PromotionStatus) => {
    const styles = {
      active: { bg: 'bg-green-500/20', text: 'text-green-400', icon: <CheckCircle size={14} /> },
      scheduled: { bg: 'bg-blue-500/20', text: 'text-blue-400', icon: <Clock size={14} /> },
      expired: { bg: 'bg-zinc-500/20', text: 'text-zinc-400', icon: <XCircle size={14} /> },
      paused: { bg: 'bg-yellow-500/20', text: 'text-yellow-400', icon: <Pause size={14} /> }
    };
    const style = styles[status];
    return (
      <span className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium ${style.bg} ${style.text}`}>
        {style.icon}
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    );
  };

  const getPromotionTypeLabel = (type: PromotionType) => {
    const labels = {
      deposit_match: 'Deposit Match',
      no_deposit: 'No Deposit',
      cashback: 'Cashback',
      rebate: 'Rebate',
      loyalty: 'Loyalty'
    };
    return labels[type];
  };

  const formatValue = (promo: Promotion) => {
    if (promo.valueType === 'percentage') {
      return `${promo.value}%`;
    }
    return `$${promo.value}`;
  };

  const handleCreatePromotion = () => {
    console.log('Creating promotion:', newPromotion);
    setShowCreateModal(false);
    // Reset form
    setNewPromotion({
      name: '',
      type: 'deposit_match',
      valueType: 'percentage',
      value: 0,
      minDeposit: 0,
      maxBonus: 0,
      eligibleAccountTypes: [],
      startDate: '',
      endDate: '',
      terms: ''
    });
  };

  return (
    <div className="h-full flex flex-col bg-[#18181b] overflow-hidden">
      {/* Header */}
      <div className="flex-shrink-0 px-6 py-4 border-b border-zinc-700 flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-zinc-100">Promotion & Bonus Management</h2>
          <p className="text-sm text-zinc-400 mt-1">Manage trading bonuses and promotional campaigns</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm font-medium transition-colors"
        >
          <Plus size={16} />
          Create Promotion
        </button>
      </div>

      {/* Stats Cards */}
      <div className="flex-shrink-0 px-6 py-4 grid grid-cols-4 gap-4">
        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-zinc-400 uppercase tracking-wide">Active Promotions</p>
              <p className="text-2xl font-bold text-zinc-100 mt-1">{stats.activePromotions}</p>
            </div>
            <Gift className="text-green-400" size={32} />
          </div>
        </div>

        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-zinc-400 uppercase tracking-wide">Total Bonus Paid</p>
              <p className="text-2xl font-bold text-blue-400 mt-1">${stats.totalBonusPaid}</p>
            </div>
            <DollarSign className="text-blue-400" size={32} />
          </div>
        </div>

        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-zinc-400 uppercase tracking-wide">Avg Conversion</p>
              <p className="text-2xl font-bold text-orange-400 mt-1">{stats.avgConversion}%</p>
            </div>
            <Users className="text-orange-400" size={32} />
          </div>
        </div>

        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-zinc-400 uppercase tracking-wide">ROI Estimate</p>
              <p className="text-2xl font-bold text-green-400 mt-1">{stats.roiEstimate}x</p>
            </div>
            <TrendingUp className="text-green-400" size={32} />
          </div>
        </div>
      </div>

      {/* Monthly Payouts Chart */}
      <div className="flex-shrink-0 px-6 py-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3 uppercase tracking-wide">
          Bonus Payouts by Month (Last 6 Months)
        </h3>
        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <svg width="100%" height="180" viewBox="0 0 800 180">
            {/* Grid lines */}
            {[0, 1, 2, 3, 4].map(i => (
              <line
                key={i}
                x1="60"
                y1={20 + i * 30}
                x2="780"
                y2={20 + i * 30}
                stroke="#3f3f46"
                strokeWidth="1"
                strokeDasharray="4 4"
              />
            ))}

            {/* Y-axis labels */}
            {[0, 1, 2, 3, 4].map(i => (
              <text key={i} x="50" y={25 + i * 30} textAnchor="end" fill="#71717a" fontSize="11">
                ${((4 - i) * (maxPayout / 4) / 1000).toFixed(0)}k
              </text>
            ))}

            {/* Bars */}
            {monthlyPayouts.map((data, idx) => {
              const x = 80 + idx * 120;
              const barHeight = (data.amount / maxPayout) * 140;
              const y = 160 - barHeight;

              return (
                <g key={idx}>
                  <rect
                    x={x}
                    y={y}
                    width="80"
                    height={barHeight}
                    fill="#3b82f6"
                    rx="4"
                  />
                  <text
                    x={x + 40}
                    y={y - 5}
                    textAnchor="middle"
                    fill="#a1a1aa"
                    fontSize="10"
                  >
                    ${(data.amount / 1000).toFixed(1)}k
                  </text>
                  <text
                    x={x + 40}
                    y="175"
                    textAnchor="middle"
                    fill="#71717a"
                    fontSize="11"
                  >
                    {data.month}
                  </text>
                </g>
              );
            })}
          </svg>
        </div>
      </div>

      {/* Promotions Grid */}
      <div className="flex-1 flex flex-col px-6 pb-4 overflow-hidden">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-semibold text-zinc-300 uppercase tracking-wide">Promotions</h3>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as PromotionStatus | 'all')}
            className="px-3 py-1.5 bg-[#27272a] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
          >
            <option value="all">All Statuses</option>
            <option value="active">Active</option>
            <option value="scheduled">Scheduled</option>
            <option value="expired">Expired</option>
            <option value="paused">Paused</option>
          </select>
        </div>

        <div className="flex-1 overflow-auto grid grid-cols-3 gap-4 content-start">
          {filteredPromotions.map((promo) => (
            <div
              key={promo.id}
              className="bg-[#27272a] rounded-lg border border-zinc-700 p-4 hover:border-zinc-600 transition-colors"
            >
              <div className="flex items-start justify-between mb-3">
                <div>
                  <h4 className="text-sm font-semibold text-zinc-100">{promo.name}</h4>
                  <p className="text-xs text-zinc-500 mt-0.5">{getPromotionTypeLabel(promo.type)}</p>
                </div>
                {getStatusBadge(promo.status)}
              </div>

              <div className="space-y-2 mb-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-zinc-400">Value:</span>
                  <span className="text-sm font-semibold text-blue-400">{formatValue(promo)}</span>
                </div>
                {promo.minDeposit && (
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-zinc-400">Min Deposit:</span>
                    <span className="text-sm font-medium text-zinc-300">${promo.minDeposit}</span>
                  </div>
                )}
                {promo.maxBonus && (
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-zinc-400">Max Bonus:</span>
                    <span className="text-sm font-medium text-zinc-300">${promo.maxBonus}</span>
                  </div>
                )}
                <div className="flex items-center justify-between">
                  <span className="text-xs text-zinc-400">Eligible Clients:</span>
                  <span className="text-sm font-medium text-zinc-300">{promo.eligibleClients}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-zinc-400">Claimed:</span>
                  <span className="text-sm font-medium text-green-400">{promo.claimedCount}</span>
                </div>
              </div>

              <div className="flex items-center gap-1 text-xs text-zinc-500 mb-4">
                <Calendar size={12} />
                <span>{promo.startDate} to {promo.endDate}</span>
              </div>

              <div className="flex gap-2">
                <button
                  onClick={() => {
                    setSelectedPromotion(promo);
                    setShowDetailModal(true);
                  }}
                  className="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 bg-zinc-700 hover:bg-zinc-600 text-zinc-100 text-xs font-medium rounded transition-colors"
                >
                  <Eye size={14} />
                  Details
                </button>
                <button className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium rounded transition-colors">
                  <Edit size={14} />
                </button>
                <button className="px-3 py-1.5 bg-red-600 hover:bg-red-700 text-white text-xs font-medium rounded transition-colors">
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Create Promotion Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#27272a] rounded-lg border border-zinc-700 w-full max-w-2xl max-h-[90vh] overflow-auto">
            <div className="px-6 py-4 border-b border-zinc-700 flex items-center justify-between sticky top-0 bg-[#27272a]">
              <h3 className="text-lg font-semibold text-zinc-100">Create New Promotion</h3>
              <button
                onClick={() => setShowCreateModal(false)}
                className="p-1 hover:bg-zinc-700 rounded transition-colors"
              >
                <X size={20} className="text-zinc-400" />
              </button>
            </div>

            <div className="px-6 py-4 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-2">Promotion Name</label>
                  <input
                    type="text"
                    value={newPromotion.name}
                    onChange={(e) => setNewPromotion({ ...newPromotion, name: e.target.value })}
                    className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                    placeholder="e.g., Welcome Bonus"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-2">Type</label>
                  <select
                    value={newPromotion.type}
                    onChange={(e) => setNewPromotion({ ...newPromotion, type: e.target.value as PromotionType })}
                    className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                  >
                    <option value="deposit_match">Deposit Match</option>
                    <option value="no_deposit">No Deposit</option>
                    <option value="cashback">Cashback</option>
                    <option value="rebate">Rebate</option>
                    <option value="loyalty">Loyalty</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-2">Value Type</label>
                  <select
                    value={newPromotion.valueType}
                    onChange={(e) => setNewPromotion({ ...newPromotion, valueType: e.target.value as ValueType })}
                    className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                  >
                    <option value="percentage">Percentage (%)</option>
                    <option value="fixed">Fixed ($)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-2">
                    Value {newPromotion.valueType === 'percentage' ? '(%)' : '($)'}
                  </label>
                  <input
                    type="number"
                    value={newPromotion.value}
                    onChange={(e) => setNewPromotion({ ...newPromotion, value: parseFloat(e.target.value) })}
                    className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                    placeholder="0"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-2">Min Deposit ($)</label>
                  <input
                    type="number"
                    value={newPromotion.minDeposit}
                    onChange={(e) => setNewPromotion({ ...newPromotion, minDeposit: parseFloat(e.target.value) })}
                    className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                    placeholder="0"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-2">Max Bonus ($)</label>
                  <input
                    type="number"
                    value={newPromotion.maxBonus}
                    onChange={(e) => setNewPromotion({ ...newPromotion, maxBonus: parseFloat(e.target.value) })}
                    className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                    placeholder="0"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-2">Start Date</label>
                  <input
                    type="date"
                    value={newPromotion.startDate}
                    onChange={(e) => setNewPromotion({ ...newPromotion, startDate: e.target.value })}
                    className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-2">End Date</label>
                  <input
                    type="date"
                    value={newPromotion.endDate}
                    onChange={(e) => setNewPromotion({ ...newPromotion, endDate: e.target.value })}
                    className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">Terms & Conditions</label>
                <textarea
                  value={newPromotion.terms}
                  onChange={(e) => setNewPromotion({ ...newPromotion, terms: e.target.value })}
                  rows={4}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                  placeholder="Enter promotion terms and conditions..."
                />
              </div>
            </div>

            <div className="px-6 py-4 border-t border-zinc-700 flex justify-end gap-3">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-zinc-100 text-sm font-medium rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleCreatePromotion}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded transition-colors"
              >
                Create Promotion
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Detail Modal */}
      {showDetailModal && selectedPromotion && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#27272a] rounded-lg border border-zinc-700 w-full max-w-4xl max-h-[90vh] overflow-auto">
            <div className="px-6 py-4 border-b border-zinc-700 flex items-center justify-between sticky top-0 bg-[#27272a]">
              <div>
                <h3 className="text-lg font-semibold text-zinc-100">{selectedPromotion.name}</h3>
                <p className="text-sm text-zinc-400">{getPromotionTypeLabel(selectedPromotion.type)}</p>
              </div>
              <button
                onClick={() => setShowDetailModal(false)}
                className="p-1 hover:bg-zinc-700 rounded transition-colors"
              >
                <X size={20} className="text-zinc-400" />
              </button>
            </div>

            <div className="px-6 py-4">
              {/* Stats Row */}
              <div className="grid grid-cols-3 gap-4 mb-6">
                <div className="bg-[#18181b] border border-zinc-700 rounded p-4">
                  <p className="text-xs text-zinc-400 uppercase tracking-wide mb-1">Total Cost to Broker</p>
                  <p className="text-2xl font-bold text-red-400">${selectedPromotion.totalCost.toLocaleString()}</p>
                </div>
                <div className="bg-[#18181b] border border-zinc-700 rounded p-4">
                  <p className="text-xs text-zinc-400 uppercase tracking-wide mb-1">Conversion Rate</p>
                  <p className="text-2xl font-bold text-green-400">{selectedPromotion.conversionRate}%</p>
                </div>
                <div className="bg-[#18181b] border border-zinc-700 rounded p-4">
                  <p className="text-xs text-zinc-400 uppercase tracking-wide mb-1">Claims / Eligible</p>
                  <p className="text-2xl font-bold text-blue-400">
                    {selectedPromotion.claimedCount} / {selectedPromotion.eligibleClients}
                  </p>
                </div>
              </div>

              {/* Claimed Clients Table */}
              <h4 className="text-sm font-semibold text-zinc-300 mb-3 uppercase tracking-wide">
                Claimed Clients ({mockClaimedClients.length})
              </h4>
              <div className="bg-[#18181b] border border-zinc-700 rounded overflow-auto max-h-96">
                <table className="w-full text-sm">
                  <thead className="sticky top-0 bg-[#18181b] border-b border-zinc-700">
                    <tr>
                      <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Account ID</th>
                      <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Client Name</th>
                      <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Email</th>
                      <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Claimed Date</th>
                      <th className="text-right px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Bonus Amount</th>
                      <th className="text-right px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Deposited</th>
                    </tr>
                  </thead>
                  <tbody>
                    {mockClaimedClients.map((client, idx) => (
                      <tr
                        key={client.accountId}
                        className={`border-b border-zinc-700/50 ${
                          idx % 2 === 0 ? 'bg-[#18181b]' : 'bg-[#18181b]/50'
                        }`}
                      >
                        <td className="px-4 py-3 font-mono text-zinc-300">{client.accountId}</td>
                        <td className="px-4 py-3 text-zinc-300">{client.name}</td>
                        <td className="px-4 py-3 text-zinc-400">{client.email}</td>
                        <td className="px-4 py-3 text-zinc-400">{client.claimedDate}</td>
                        <td className="px-4 py-3 text-right font-semibold text-green-400">${client.bonusAmount}</td>
                        <td className="px-4 py-3 text-right text-zinc-300">${client.deposited}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
