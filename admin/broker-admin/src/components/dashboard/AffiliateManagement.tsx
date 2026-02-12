/**
 * Affiliate / Introducing Broker (IB) Management
 * Manage affiliate program, commissions, and payouts
 */

'use client';

import React, { useState, useMemo } from 'react';
import {
  Users,
  DollarSign,
  TrendingUp,
  Clock,
  CheckCircle,
  XCircle,
  Eye,
  X,
  Award,
  UserCheck,
  UserX,
  Edit,
  Download,
  Filter,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// Types
type AffiliateStatus = 'pending' | 'active' | 'suspended' | 'rejected';
type CommissionTier = 'Bronze' | 'Silver' | 'Gold' | 'Platinum';
type PayoutStatus = 'pending' | 'processing' | 'completed' | 'failed';

interface Affiliate {
  id: string;
  name: string;
  email: string;
  status: AffiliateStatus;
  tier: CommissionTier;
  referredClients: number;
  totalVolume: number;
  commissionEarned: number;
  payoutStatus: PayoutStatus;
  joinDate: number;
  lastActivity: number;
}

interface ReferredClient {
  id: string;
  name: string;
  account: string;
  volume: number;
  commission: number;
  joinDate: number;
}

interface MonthlyCommission {
  month: string;
  amount: number;
  volume: number;
}

interface PayoutRecord {
  id: string;
  amount: number;
  method: string;
  status: PayoutStatus;
  date: number;
  transactionRef?: string;
}

// Commission tier configuration
const COMMISSION_TIERS = {
  Bronze: { minVolume: 0, maxVolume: 100000, pipRate: 0.5, color: 'bg-amber-700/20 text-amber-400' },
  Silver: { minVolume: 100000, maxVolume: 500000, pipRate: 0.8, color: 'bg-zinc-400/20 text-zinc-300' },
  Gold: { minVolume: 500000, maxVolume: 1000000, pipRate: 1.0, color: 'bg-yellow-500/20 text-yellow-400' },
  Platinum: { minVolume: 1000000, maxVolume: Infinity, pipRate: 1.2, color: 'bg-purple-500/20 text-purple-400' },
};

// Mock Data - 15+ affiliates
const MOCK_AFFILIATES: Affiliate[] = [
  { id: 'IB-001', name: 'Michael Chen', email: 'michael.chen@example.com', status: 'active', tier: 'Platinum', referredClients: 45, totalVolume: 2500000, commissionEarned: 15800, payoutStatus: 'pending', joinDate: Date.now() - 31536000000, lastActivity: Date.now() - 86400000 },
  { id: 'IB-002', name: 'Sarah Williams', email: 'sarah.w@example.com', status: 'active', tier: 'Gold', referredClients: 32, totalVolume: 850000, commissionEarned: 9200, payoutStatus: 'completed', joinDate: Date.now() - 25920000000, lastActivity: Date.now() - 172800000 },
  { id: 'IB-003', name: 'David Rodriguez', email: 'david.r@example.com', status: 'active', tier: 'Silver', referredClients: 18, totalVolume: 320000, commissionEarned: 4100, payoutStatus: 'processing', joinDate: Date.now() - 20736000000, lastActivity: Date.now() - 259200000 },
  { id: 'IB-004', name: 'Emma Thompson', email: 'emma.t@example.com', status: 'active', tier: 'Platinum', referredClients: 52, totalVolume: 3200000, commissionEarned: 21500, payoutStatus: 'completed', joinDate: Date.now() - 28512000000, lastActivity: Date.now() - 43200000 },
  { id: 'IB-005', name: 'James Anderson', email: 'james.a@example.com', status: 'active', tier: 'Gold', referredClients: 28, totalVolume: 720000, commissionEarned: 7800, payoutStatus: 'pending', joinDate: Date.now() - 18144000000, lastActivity: Date.now() - 345600000 },
  { id: 'IB-006', name: 'Olivia Martinez', email: 'olivia.m@example.com', status: 'active', tier: 'Silver', referredClients: 15, totalVolume: 280000, commissionEarned: 3500, payoutStatus: 'completed', joinDate: Date.now() - 15552000000, lastActivity: Date.now() - 432000000 },
  { id: 'IB-007', name: 'William Taylor', email: 'william.t@example.com', status: 'pending', tier: 'Bronze', referredClients: 5, totalVolume: 45000, commissionEarned: 280, payoutStatus: 'pending', joinDate: Date.now() - 2592000000, lastActivity: Date.now() - 518400000 },
  { id: 'IB-008', name: 'Sophia Garcia', email: 'sophia.g@example.com', status: 'active', tier: 'Platinum', referredClients: 38, totalVolume: 1850000, commissionEarned: 13200, payoutStatus: 'processing', joinDate: Date.now() - 22896000000, lastActivity: Date.now() - 604800000 },
  { id: 'IB-009', name: 'Benjamin Lee', email: 'benjamin.l@example.com', status: 'active', tier: 'Gold', referredClients: 25, totalVolume: 650000, commissionEarned: 6900, payoutStatus: 'completed', joinDate: Date.now() - 19872000000, lastActivity: Date.now() - 691200000 },
  { id: 'IB-010', name: 'Charlotte Brown', email: 'charlotte.b@example.com', status: 'suspended', tier: 'Silver', referredClients: 12, totalVolume: 180000, commissionEarned: 2100, payoutStatus: 'failed', joinDate: Date.now() - 12960000000, lastActivity: Date.now() - 1296000000 },
  { id: 'IB-011', name: 'Alexander Wilson', email: 'alex.w@example.com', status: 'active', tier: 'Bronze', referredClients: 8, totalVolume: 75000, commissionEarned: 450, payoutStatus: 'pending', joinDate: Date.now() - 10368000000, lastActivity: Date.now() - 777600000 },
  { id: 'IB-012', name: 'Isabella Moore', email: 'isabella.m@example.com', status: 'active', tier: 'Platinum', referredClients: 41, totalVolume: 2100000, commissionEarned: 16800, payoutStatus: 'completed', joinDate: Date.now() - 24192000000, lastActivity: Date.now() - 129600000 },
  { id: 'IB-013', name: 'Daniel Jackson', email: 'daniel.j@example.com', status: 'active', tier: 'Gold', referredClients: 22, totalVolume: 580000, commissionEarned: 6200, payoutStatus: 'pending', joinDate: Date.now() - 17280000000, lastActivity: Date.now() - 864000000 },
  { id: 'IB-014', name: 'Mia Harris', email: 'mia.h@example.com', status: 'rejected', tier: 'Bronze', referredClients: 0, totalVolume: 0, commissionEarned: 0, payoutStatus: 'pending', joinDate: Date.now() - 5184000000, lastActivity: Date.now() - 5184000000 },
  { id: 'IB-015', name: 'Ethan Clark', email: 'ethan.c@example.com', status: 'active', tier: 'Silver', referredClients: 19, totalVolume: 340000, commissionEarned: 4500, payoutStatus: 'processing', joinDate: Date.now() - 14688000000, lastActivity: Date.now() - 950400000 },
  { id: 'IB-016', name: 'Ava Lewis', email: 'ava.l@example.com', status: 'pending', tier: 'Bronze', referredClients: 3, totalVolume: 25000, commissionEarned: 150, payoutStatus: 'pending', joinDate: Date.now() - 1728000000, lastActivity: Date.now() - 1036800000 },
  { id: 'IB-017', name: 'Noah Walker', email: 'noah.w@example.com', status: 'active', tier: 'Platinum', referredClients: 48, totalVolume: 2800000, commissionEarned: 19500, payoutStatus: 'completed', joinDate: Date.now() - 27648000000, lastActivity: Date.now() - 216000000 },
  { id: 'IB-018', name: 'Emily Hall', email: 'emily.h@example.com', status: 'active', tier: 'Gold', referredClients: 27, totalVolume: 690000, commissionEarned: 7400, payoutStatus: 'pending', joinDate: Date.now() - 16416000000, lastActivity: Date.now() - 1123200000 },
];

// Mock referred clients (for detail modal)
const generateMockReferredClients = (count: number, affiliateId: string): ReferredClient[] => {
  const clients: ReferredClient[] = [];
  for (let i = 1; i <= count; i++) {
    clients.push({
      id: `${affiliateId}-CLIENT-${i.toString().padStart(3, '0')}`,
      name: `Client ${i}`,
      account: `68096${2800 + i}`,
      volume: Math.floor(Math.random() * 100000) + 10000,
      commission: Math.floor(Math.random() * 500) + 50,
      joinDate: Date.now() - Math.floor(Math.random() * 15552000000),
    });
  }
  return clients;
};

// Mock monthly commissions (for chart)
const generateMonthlyCommissions = (): MonthlyCommission[] => {
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  const currentMonth = new Date().getMonth();
  const data: MonthlyCommission[] = [];

  for (let i = 11; i >= 0; i--) {
    const monthIndex = (currentMonth - i + 12) % 12;
    data.push({
      month: months[monthIndex],
      amount: Math.floor(Math.random() * 2000) + 500,
      volume: Math.floor(Math.random() * 200000) + 50000,
    });
  }

  return data;
};

// Mock payout history
const generatePayoutHistory = (affiliateId: string): PayoutRecord[] => {
  return [
    { id: `PAY-${affiliateId}-001`, amount: 5000, method: 'Bank Wire', status: 'completed', date: Date.now() - 2592000000, transactionRef: 'TXN-BW-5421' },
    { id: `PAY-${affiliateId}-002`, amount: 3500, method: 'PayPal', status: 'completed', date: Date.now() - 5184000000, transactionRef: 'TXN-PP-8932' },
    { id: `PAY-${affiliateId}-003`, amount: 4200, method: 'Bank Wire', status: 'completed', date: Date.now() - 7776000000, transactionRef: 'TXN-BW-1245' },
    { id: `PAY-${affiliateId}-004`, amount: 2800, method: 'Crypto', status: 'completed', date: Date.now() - 10368000000, transactionRef: 'TXN-CR-6789' },
  ];
};

export default function AffiliateManagement() {
  const [affiliates] = useState<Affiliate[]>(MOCK_AFFILIATES);
  const [selectedAffiliate, setSelectedAffiliate] = useState<Affiliate | null>(null);
  const [showDetailModal, setShowDetailModal] = useState(false);

  // Filters
  const [statusFilter, setStatusFilter] = useState<'all' | AffiliateStatus>('all');
  const [tierFilter, setTierFilter] = useState<'all' | CommissionTier>('all');
  const [searchQuery, setSearchQuery] = useState('');

  // Summary stats
  const stats = useMemo(() => {
    const totalAffiliates = affiliates.length;
    const activeIBs = affiliates.filter(a => a.status === 'active').length;
    const totalCommissionsPaid = affiliates
      .filter(a => a.payoutStatus === 'completed')
      .reduce((sum, a) => sum + a.commissionEarned, 0);
    const pendingPayouts = affiliates
      .filter(a => a.payoutStatus === 'pending')
      .reduce((sum, a) => sum + a.commissionEarned, 0);

    return {
      totalAffiliates,
      activeIBs,
      totalCommissionsPaid,
      pendingPayouts,
    };
  }, [affiliates]);

  // Filtered affiliates
  const filteredAffiliates = useMemo(() => {
    let result = [...affiliates];

    if (statusFilter !== 'all') {
      result = result.filter(a => a.status === statusFilter);
    }

    if (tierFilter !== 'all') {
      result = result.filter(a => a.tier === tierFilter);
    }

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter(a =>
        a.name.toLowerCase().includes(query) ||
        a.email.toLowerCase().includes(query) ||
        a.id.toLowerCase().includes(query)
      );
    }

    return result.sort((a, b) => b.commissionEarned - a.commissionEarned);
  }, [affiliates, statusFilter, tierFilter, searchQuery]);

  // Status badge helper
  const getStatusBadge = (status: AffiliateStatus) => {
    switch (status) {
      case 'pending':
        return <span className="px-2 py-0.5 rounded bg-yellow-500/20 text-yellow-400 text-xs font-semibold">Pending</span>;
      case 'active':
        return <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 text-xs font-semibold">Active</span>;
      case 'suspended':
        return <span className="px-2 py-0.5 rounded bg-orange-500/20 text-orange-400 text-xs font-semibold">Suspended</span>;
      case 'rejected':
        return <span className="px-2 py-0.5 rounded bg-rose-500/20 text-rose-400 text-xs font-semibold">Rejected</span>;
    }
  };

  const getPayoutStatusBadge = (status: PayoutStatus) => {
    switch (status) {
      case 'pending':
        return <span className="px-2 py-0.5 rounded bg-yellow-500/20 text-yellow-400 text-xs font-semibold">Pending</span>;
      case 'processing':
        return <span className="px-2 py-0.5 rounded bg-blue-500/20 text-blue-400 text-xs font-semibold">Processing</span>;
      case 'completed':
        return <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 text-xs font-semibold">Completed</span>;
      case 'failed':
        return <span className="px-2 py-0.5 rounded bg-rose-500/20 text-rose-400 text-xs font-semibold">Failed</span>;
    }
  };

  const getTierBadge = (tier: CommissionTier) => {
    const config = COMMISSION_TIERS[tier];
    return <span className={`px-2 py-0.5 rounded text-xs font-semibold ${config.color}`}>{tier}</span>;
  };

  const handleViewDetails = (affiliate: Affiliate) => {
    setSelectedAffiliate(affiliate);
    setShowDetailModal(true);
  };

  const handleApprove = (affiliateId: string) => {
    console.log('Approve affiliate:', affiliateId);
    // TODO: API call to approve affiliate
  };

  const handleSuspend = (affiliateId: string) => {
    console.log('Suspend affiliate:', affiliateId);
    // TODO: API call to suspend affiliate
  };

  const handleAdjustTier = (affiliateId: string) => {
    console.log('Adjust tier for affiliate:', affiliateId);
    // TODO: Show tier adjustment modal
  };

  const handleProcessPayout = (affiliateId: string) => {
    console.log('Process payout for affiliate:', affiliateId);
    // TODO: Show payout processing modal
  };

  return (
    <div className="h-full flex flex-col bg-[#18181b] text-zinc-300 overflow-auto custom-scrollbar">
      {/* Summary Cards */}
      <div className="grid grid-cols-4 gap-4 p-4">
        {/* Total Affiliates */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-blue-500/20 flex items-center justify-center">
              <Users size={20} className="text-blue-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-blue-400">{stats.totalAffiliates}</div>
              <div className="text-xs text-zinc-500">Total Affiliates</div>
            </div>
          </div>
        </div>

        {/* Active IBs */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-emerald-500/20 flex items-center justify-center">
              <UserCheck size={20} className="text-emerald-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-emerald-400">{stats.activeIBs}</div>
              <div className="text-xs text-zinc-500">Active IBs</div>
            </div>
          </div>
        </div>

        {/* Total Commissions Paid */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-purple-500/20 flex items-center justify-center">
              <DollarSign size={20} className="text-purple-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-purple-400">
                ${stats.totalCommissionsPaid.toLocaleString()}
              </div>
              <div className="text-xs text-zinc-500">Commissions Paid</div>
            </div>
          </div>
        </div>

        {/* Pending Payouts */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-yellow-500/20 flex items-center justify-center">
              <Clock size={20} className="text-yellow-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-yellow-400">
                ${stats.pendingPayouts.toLocaleString()}
              </div>
              <div className="text-xs text-zinc-500">Pending Payouts</div>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="mx-4 mb-4 flex items-center gap-3 flex-wrap">
        <div className="flex items-center gap-2">
          <Filter size={14} className="text-zinc-500" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search by name, email, or ID..."
            className="px-3 py-1.5 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500 w-64"
          />
        </div>

        <div className="flex items-center gap-2">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as any)}
            className="px-2 py-1 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
          >
            <option value="all">All Status</option>
            <option value="pending">Pending</option>
            <option value="active">Active</option>
            <option value="suspended">Suspended</option>
            <option value="rejected">Rejected</option>
          </select>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={tierFilter}
            onChange={(e) => setTierFilter(e.target.value as any)}
            className="px-2 py-1 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
          >
            <option value="all">All Tiers</option>
            <option value="Bronze">Bronze</option>
            <option value="Silver">Silver</option>
            <option value="Gold">Gold</option>
            <option value="Platinum">Platinum</option>
          </select>
        </div>
      </div>

      {/* Affiliates Table */}
      <div className="flex-1 mx-4 mb-4 bg-[#27272a] border border-[#3f3f46] rounded overflow-hidden">
        <div className="overflow-auto custom-scrollbar h-full">
          <table className="w-full text-xs">
            <thead className="bg-[#1e1e1e] border-b border-[#3f3f46] sticky top-0">
              <tr>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">ID</th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Name</th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Email</th>
                <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Status</th>
                <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Tier</th>
                <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Referred</th>
                <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Total Volume</th>
                <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Commission</th>
                <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Payout</th>
                <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredAffiliates.map((affiliate) => (
                <tr
                  key={affiliate.id}
                  className="border-b border-[#3f3f46]/30 hover:bg-zinc-800/30"
                >
                  <td className="px-3 py-2 text-blue-400 font-mono">{affiliate.id}</td>
                  <td className="px-3 py-2 text-zinc-300">{affiliate.name}</td>
                  <td className="px-3 py-2 text-zinc-500">{affiliate.email}</td>
                  <td className="px-3 py-2 text-center">{getStatusBadge(affiliate.status)}</td>
                  <td className="px-3 py-2 text-center">{getTierBadge(affiliate.tier)}</td>
                  <td className="px-3 py-2 text-right text-zinc-400">{affiliate.referredClients}</td>
                  <td className="px-3 py-2 text-right font-semibold text-blue-400">
                    ${affiliate.totalVolume.toLocaleString()}
                  </td>
                  <td className="px-3 py-2 text-right font-semibold text-emerald-400">
                    ${affiliate.commissionEarned.toLocaleString()}
                  </td>
                  <td className="px-3 py-2 text-center">{getPayoutStatusBadge(affiliate.payoutStatus)}</td>
                  <td className="px-3 py-2">
                    <div className="flex items-center justify-center gap-1">
                      <button
                        onClick={() => handleViewDetails(affiliate)}
                        className="p-1 rounded hover:bg-zinc-700 text-blue-400"
                        title="View Details"
                      >
                        <Eye size={14} />
                      </button>
                      {affiliate.status === 'pending' && (
                        <button
                          onClick={() => handleApprove(affiliate.id)}
                          className="p-1 rounded hover:bg-zinc-700 text-emerald-400"
                          title="Approve"
                        >
                          <CheckCircle size={14} />
                        </button>
                      )}
                      {affiliate.status === 'active' && (
                        <button
                          onClick={() => handleSuspend(affiliate.id)}
                          className="p-1 rounded hover:bg-zinc-700 text-orange-400"
                          title="Suspend"
                        >
                          <UserX size={14} />
                        </button>
                      )}
                      <button
                        onClick={() => handleAdjustTier(affiliate.id)}
                        className="p-1 rounded hover:bg-zinc-700 text-purple-400"
                        title="Adjust Tier"
                      >
                        <Edit size={14} />
                      </button>
                      {affiliate.payoutStatus === 'pending' && (
                        <button
                          onClick={() => handleProcessPayout(affiliate.id)}
                          className="p-1 rounded hover:bg-zinc-700 text-yellow-400"
                          title="Process Payout"
                        >
                          <DollarSign size={14} />
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

      {/* Detail Modal */}
      {showDetailModal && selectedAffiliate && (
        <AffiliateDetailModal
          affiliate={selectedAffiliate}
          onClose={() => {
            setShowDetailModal(false);
            setSelectedAffiliate(null);
          }}
        />
      )}
    </div>
  );
}

// Affiliate Detail Modal Component
function AffiliateDetailModal({ affiliate, onClose }: { affiliate: Affiliate; onClose: () => void }) {
  const [activeTab, setActiveTab] = useState<'clients' | 'commissions' | 'payouts' | 'tier'>('clients');

  const referredClients = useMemo(() => generateMockReferredClients(affiliate.referredClients, affiliate.id), [affiliate]);
  const monthlyCommissions = useMemo(() => generateMonthlyCommissions(), []);
  const payoutHistory = useMemo(() => generatePayoutHistory(affiliate.id), [affiliate.id]);

  const tierConfig = COMMISSION_TIERS[affiliate.tier];

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-[#27272a] border border-[#3f3f46] rounded-lg max-w-5xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-[#3f3f46]">
          <div>
            <h2 className="text-lg font-bold text-zinc-200">{affiliate.name}</h2>
            <div className="text-xs text-zinc-500">{affiliate.email}</div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-zinc-700 rounded text-zinc-400">
            <X size={20} />
          </button>
        </div>

        {/* Tab Navigation */}
        <div className="flex gap-1 border-b border-[#3f3f46] px-4">
          <button
            onClick={() => setActiveTab('clients')}
            className={`px-4 py-2 text-sm font-semibold transition-colors ${
              activeTab === 'clients'
                ? 'border-b-2 border-blue-400 text-blue-400'
                : 'text-zinc-500 hover:text-zinc-300'
            }`}
          >
            Referred Clients ({affiliate.referredClients})
          </button>
          <button
            onClick={() => setActiveTab('commissions')}
            className={`px-4 py-2 text-sm font-semibold transition-colors ${
              activeTab === 'commissions'
                ? 'border-b-2 border-emerald-400 text-emerald-400'
                : 'text-zinc-500 hover:text-zinc-300'
            }`}
          >
            Commission Breakdown
          </button>
          <button
            onClick={() => setActiveTab('payouts')}
            className={`px-4 py-2 text-sm font-semibold transition-colors ${
              activeTab === 'payouts'
                ? 'border-b-2 border-yellow-400 text-yellow-400'
                : 'text-zinc-500 hover:text-zinc-300'
            }`}
          >
            Payout History
          </button>
          <button
            onClick={() => setActiveTab('tier')}
            className={`px-4 py-2 text-sm font-semibold transition-colors ${
              activeTab === 'tier'
                ? 'border-b-2 border-purple-400 text-purple-400'
                : 'text-zinc-500 hover:text-zinc-300'
            }`}
          >
            Commission Tier
          </button>
        </div>

        {/* Tab Content */}
        <div className="flex-1 overflow-auto custom-scrollbar p-4">
          {activeTab === 'clients' && (
            <div>
              <h3 className="text-sm font-bold text-zinc-400 mb-3">Referred Clients</h3>
              <div className="bg-[#18181b] border border-[#3f3f46] rounded overflow-hidden">
                <table className="w-full text-xs">
                  <thead className="bg-[#1e1e1e] border-b border-[#3f3f46]">
                    <tr>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Client ID</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Name</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Account</th>
                      <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Volume</th>
                      <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Commission</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Join Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    {referredClients.map((client) => (
                      <tr key={client.id} className="border-b border-[#3f3f46]/30 hover:bg-zinc-800/30">
                        <td className="px-3 py-2 text-blue-400 font-mono">{client.id}</td>
                        <td className="px-3 py-2 text-zinc-300">{client.name}</td>
                        <td className="px-3 py-2 text-zinc-500">{client.account}</td>
                        <td className="px-3 py-2 text-right text-blue-400">${client.volume.toLocaleString()}</td>
                        <td className="px-3 py-2 text-right text-emerald-400">${client.commission}</td>
                        <td className="px-3 py-2 text-zinc-500">{new Date(client.joinDate).toLocaleDateString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {activeTab === 'commissions' && (
            <div>
              <h3 className="text-sm font-bold text-zinc-400 mb-3">Monthly Commission Breakdown (Last 12 Months)</h3>
              <svg viewBox="0 0 700 250" className="w-full h-64 bg-[#18181b] border border-[#3f3f46] rounded p-4">
                {/* Grid lines */}
                {[0, 1, 2, 3, 4].map((i) => (
                  <line
                    key={i}
                    x1="50"
                    y1={30 + i * 45}
                    x2="680"
                    y2={30 + i * 45}
                    stroke="#3f3f46"
                    strokeWidth="0.5"
                  />
                ))}

                {/* Bars */}
                {monthlyCommissions.map((month, index) => {
                  const x = 60 + index * 52;
                  const maxVal = Math.max(...monthlyCommissions.map(m => m.amount));
                  const barHeight = (month.amount / maxVal) * 170;

                  return (
                    <g key={index}>
                      {/* Bar */}
                      <rect
                        x={x}
                        y={210 - barHeight}
                        width="40"
                        height={barHeight}
                        fill="#22c55e"
                        opacity="0.7"
                      />
                      {/* Amount label */}
                      <text
                        x={x + 20}
                        y={205 - barHeight}
                        fill="#a1a1aa"
                        fontSize="9"
                        textAnchor="middle"
                      >
                        ${month.amount}
                      </text>
                      {/* Month label */}
                      <text
                        x={x + 20}
                        y="230"
                        fill="#71717a"
                        fontSize="10"
                        textAnchor="middle"
                      >
                        {month.month}
                      </text>
                    </g>
                  );
                })}
              </svg>
            </div>
          )}

          {activeTab === 'payouts' && (
            <div>
              <h3 className="text-sm font-bold text-zinc-400 mb-3">Payout History</h3>
              <div className="bg-[#18181b] border border-[#3f3f46] rounded overflow-hidden">
                <table className="w-full text-xs">
                  <thead className="bg-[#1e1e1e] border-b border-[#3f3f46]">
                    <tr>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Payout ID</th>
                      <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Amount</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Method</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Status</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Date</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Transaction Ref</th>
                    </tr>
                  </thead>
                  <tbody>
                    {payoutHistory.map((payout) => (
                      <tr key={payout.id} className="border-b border-[#3f3f46]/30 hover:bg-zinc-800/30">
                        <td className="px-3 py-2 text-blue-400 font-mono">{payout.id}</td>
                        <td className="px-3 py-2 text-right font-semibold text-emerald-400">
                          ${payout.amount.toLocaleString()}
                        </td>
                        <td className="px-3 py-2 text-zinc-400">{payout.method}</td>
                        <td className="px-3 py-2 text-center">
                          <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 text-xs font-semibold">
                            {payout.status}
                          </span>
                        </td>
                        <td className="px-3 py-2 text-zinc-500">{new Date(payout.date).toLocaleDateString()}</td>
                        <td className="px-3 py-2 text-zinc-500 font-mono text-[10px]">{payout.transactionRef}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {activeTab === 'tier' && (
            <div>
              <h3 className="text-sm font-bold text-zinc-400 mb-3">Commission Tier Information</h3>
              <div className="bg-[#18181b] border border-[#3f3f46] rounded p-4 space-y-4">
                {/* Current Tier */}
                <div>
                  <div className="text-xs text-zinc-500 mb-1">Current Tier</div>
                  <div className="flex items-center gap-3">
                    <Award size={24} className={affiliate.tier === 'Platinum' ? 'text-purple-400' : affiliate.tier === 'Gold' ? 'text-yellow-400' : affiliate.tier === 'Silver' ? 'text-zinc-300' : 'text-amber-700'} />
                    <div>
                      <div className="text-lg font-bold">{affiliate.tier}</div>
                      <div className="text-xs text-zinc-500">{tierConfig.pipRate} pip per trade</div>
                    </div>
                  </div>
                </div>

                {/* Tier Breakdown */}
                <div className="grid grid-cols-2 gap-4 pt-4 border-t border-[#3f3f46]">
                  {Object.entries(COMMISSION_TIERS).map(([tier, config]) => (
                    <div
                      key={tier}
                      className={`p-3 rounded border ${
                        tier === affiliate.tier
                          ? 'border-blue-500 bg-blue-500/10'
                          : 'border-[#3f3f46] bg-[#27272a]'
                      }`}
                    >
                      <div className="flex items-center gap-2 mb-2">
                        <Award
                          size={16}
                          className={
                            tier === 'Platinum'
                              ? 'text-purple-400'
                              : tier === 'Gold'
                              ? 'text-yellow-400'
                              : tier === 'Silver'
                              ? 'text-zinc-300'
                              : 'text-amber-700'
                          }
                        />
                        <span className="text-sm font-bold">{tier}</span>
                      </div>
                      <div className="text-xs text-zinc-500">
                        {config.maxVolume === Infinity
                          ? `$${(config.minVolume / 1000).toFixed(0)}K+ volume`
                          : `$${(config.minVolume / 1000).toFixed(0)}K - $${(config.maxVolume / 1000).toFixed(0)}K`}
                      </div>
                      <div className="text-xs text-emerald-400 font-semibold mt-1">
                        {config.pipRate} pip per trade
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
