'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
  Link as LinkIcon,
  Users,
  TrendingUp,
  DollarSign,
  Check,
  X,
  ChevronDown,
  ChevronUp,
  Search,
  RefreshCw,
  Award,
  Target,
  Activity,
  Percent
} from 'lucide-react';
import { API_CONFIG } from '../../config/api';

// ============================================
// Type Definitions (matching backend)
// ============================================

type ReferralStatus = 'pending' | 'registered' | 'deposited' | 'traded' | 'qualified' | 'expired';
type RewardStatus = 'pending' | 'approved' | 'paid' | 'cancelled';

interface ReferralLink {
  id: string;
  code: string;
  clientID: string;
  clientName: string;
  clicks: number;
  registrations: number;
  conversions: number;
  totalReward: number;
  isActive: boolean;
  createdAt: string;
}

interface Referral {
  id: string;
  referrerID: string;
  referrerName: string;
  referredID: string;
  referredName: string;
  status: ReferralStatus;
  reward: number;
  rewardStatus: RewardStatus;
  registeredAt: string;
}

interface ReferralTier {
  id: string;
  name: string;
  minReferrals: number;
  maxReferrals: number;
  rewardPerReferral: number;
  bonusMultiplier: number;
  isActive: boolean;
}

interface ReferralStats {
  totalReferrers: number;
  totalReferred: number;
  totalConverted: number;
  conversionRate: number;
  totalRewardsPaid: number;
  avgRewardPerReferrer: number;
  topReferrerID: string;
  topReferrerName: string;
  topReferrerCount: number;
}

interface MonthlyTrend {
  month: string;
  referrals: number;
  conversions: number;
  rewardsPaid: number;
}

interface TopReferrer {
  clientID: string;
  clientName: string;
  conversions: number;
  registrations: number;
  totalReward: number;
  linkCode: string;
}

// ============================================
// Helper Functions
// ============================================

const getStatusColor = (status: ReferralStatus): string => {
  const colors: Record<ReferralStatus, string> = {
    pending: '#6B7280',      // gray
    registered: '#3B82F6',   // blue
    deposited: '#F59E0B',    // yellow
    traded: '#F97316',       // orange
    qualified: '#10B981',    // green
    expired: '#EF4444'       // red
  };
  return colors[status] || '#6B7280';
};

const getRewardStatusColor = (status: RewardStatus): string => {
  const colors: Record<RewardStatus, string> = {
    pending: '#6B7280',    // gray
    approved: '#3B82F6',   // blue
    paid: '#10B981',       // green
    cancelled: '#EF4444'   // red
  };
  return colors[status] || '#6B7280';
};

const getTierColor = (name: string): string => {
  const colors: Record<string, string> = {
    'Bronze': '#CD7F32',
    'Silver': '#C0C0C0',
    'Gold': '#FFD700',
    'Platinum': '#E5E4E2'
  };
  return colors[name] || '#6B7280';
};

const getRankBadgeColor = (rank: number): string => {
  if (rank === 1) return '#FFD700'; // gold
  if (rank === 2) return '#C0C0C0'; // silver
  if (rank === 3) return '#CD7F32'; // bronze
  return '#6B7280'; // default gray
};

const formatCurrency = (amount: number): string => {
  return `$${amount.toFixed(2)}`;
};

const formatPercent = (value: number): string => {
  return `${value.toFixed(1)}%`;
};

// ============================================
// Main Component
// ============================================

const ReferralProgram: React.FC = () => {
  // State for data
  const [links, setLinks] = useState<ReferralLink[]>([]);
  const [selectedLink, setSelectedLink] = useState<ReferralLink | null>(null);
  const [linkReferrals, setLinkReferrals] = useState<Referral[]>([]);
  const [allReferrals, setAllReferrals] = useState<Referral[]>([]);
  const [tiers, setTiers] = useState<ReferralTier[]>([]);
  const [stats, setStats] = useState<ReferralStats | null>(null);
  const [monthlyTrends, setMonthlyTrends] = useState<MonthlyTrend[]>([]);
  const [topReferrers, setTopReferrers] = useState<TopReferrer[]>([]);

  // UI State
  const [linkSearch, setLinkSearch] = useState('');
  const [linkSortBy, setLinkSortBy] = useState<'clicks' | 'registrations' | 'conversions' | 'reward'>('conversions');
  const [referralSearch, setReferralSearch] = useState('');
  const [referralStatusFilter, setReferralStatusFilter] = useState<ReferralStatus | 'all'>('all');
  const [isLoading, setIsLoading] = useState(false);
  const [showApproveModal, setShowApproveModal] = useState(false);
  const [selectedReferralForApproval, setSelectedReferralForApproval] = useState<Referral | null>(null);

  // Fetch all data on mount
  useEffect(() => {
    fetchAllData();
  }, []);

  const fetchAllData = async () => {
    setIsLoading(true);
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
      const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` })
      };

      const [linksRes, referralsRes, tiersRes, statsRes, topRes] = await Promise.all([
        fetch(`${API_CONFIG.REFERRAL_LINKS}?sort=${linkSortBy}`, { headers }),
        fetch(API_CONFIG.REFERRAL_LIST, { headers }),
        fetch(API_CONFIG.REFERRAL_TIERS, { headers }),
        fetch(API_CONFIG.REFERRAL_STATS, { headers }),
        fetch(API_CONFIG.REFERRAL_TOP_REFERRERS, { headers })
      ]);

      const linksData = await linksRes.json();
      const referralsData = await referralsRes.json();
      const tiersData = await tiersRes.json();
      const statsData = await statsRes.json();
      const topData = await topRes.json();

      setLinks(linksData.links || []);
      setAllReferrals(referralsData.referrals || []);
      setTiers(tiersData.tiers || []);
      setStats(statsData.stats || null);
      setMonthlyTrends(statsData.monthlyTrends || []);
      setTopReferrers(topData.topReferrers || []);
    } catch (error) {
      console.error('Error fetching referral data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const fetchLinkDetails = async (linkId: string) => {
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
      const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` })
      };

      const res = await fetch(`${API_CONFIG.REFERRAL_LINK_DETAIL(linkId)}`, { headers });
      const data = await res.json();

      setLinkReferrals(data.referrals || []);
    } catch (error) {
      console.error('Error fetching link details:', error);
    }
  };

  const handleLinkClick = (link: ReferralLink) => {
    if (selectedLink?.id === link.id) {
      setSelectedLink(null);
      setLinkReferrals([]);
    } else {
      setSelectedLink(link);
      fetchLinkDetails(link.id);
    }
  };

  const handleToggleLinkActive = async (linkId: string, currentActive: boolean) => {
    // In production, this would call an API endpoint to toggle active status
    // For now, we'll update locally
    setLinks(prev => prev.map(link =>
      link.id === linkId ? { ...link, isActive: !currentActive } : link
    ));
  };

  const handleApproveReward = async () => {
    if (!selectedReferralForApproval) return;

    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
      const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` })
      };

      const res = await fetch(API_CONFIG.REFERRAL_APPROVE(selectedReferralForApproval.id), {
        method: 'PUT',
        headers
      });

      if (res.ok) {
        const updatedReferral = await res.json();
        // Update in both lists
        setAllReferrals(prev => prev.map(ref =>
          ref.id === updatedReferral.id ? updatedReferral : ref
        ));
        setLinkReferrals(prev => prev.map(ref =>
          ref.id === updatedReferral.id ? updatedReferral : ref
        ));
        setShowApproveModal(false);
        setSelectedReferralForApproval(null);
        // Refresh stats
        fetchAllData();
      }
    } catch (error) {
      console.error('Error approving referral:', error);
    }
  };

  const handleUpdateTier = async (tierId: string, updates: Partial<ReferralTier>) => {
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
      const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` })
      };

      const res = await fetch(API_CONFIG.REFERRAL_TIER_UPDATE(tierId), {
        method: 'PUT',
        headers,
        body: JSON.stringify(updates)
      });

      if (res.ok) {
        const updatedTier = await res.json();
        setTiers(prev => prev.map(tier =>
          tier.id === updatedTier.id ? updatedTier : tier
        ));
      }
    } catch (error) {
      console.error('Error updating tier:', error);
    }
  };

  // Filtered and sorted links
  const filteredLinks = useMemo(() => {
    return links.filter(link =>
      link.code.toLowerCase().includes(linkSearch.toLowerCase()) ||
      link.clientName.toLowerCase().includes(linkSearch.toLowerCase())
    );
  }, [links, linkSearch]);

  // Filtered referrals
  const filteredReferrals = useMemo(() => {
    return allReferrals.filter(ref => {
      const matchesSearch = ref.referrerName.toLowerCase().includes(referralSearch.toLowerCase()) ||
        ref.referredName.toLowerCase().includes(referralSearch.toLowerCase());
      const matchesStatus = referralStatusFilter === 'all' || ref.status === referralStatusFilter;
      return matchesSearch && matchesStatus;
    });
  }, [allReferrals, referralSearch, referralStatusFilter]);

  // ============================================
  // Render Stats Overview Cards
  // ============================================

  const renderStatsCards = () => {
    if (!stats) return null;

    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 mb-6">
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-[#9CA3AF]">Total Referrers</span>
            <Users size={18} className="text-[#3B82F6]" />
          </div>
          <div className="text-2xl font-semibold text-white">{stats.totalReferrers}</div>
        </div>

        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-[#9CA3AF]">Total Referred</span>
            <Users size={18} className="text-[#10B981]" />
          </div>
          <div className="text-2xl font-semibold text-white">{stats.totalReferred}</div>
        </div>

        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-[#9CA3AF]">Conversion Rate</span>
            <Percent size={18} className="text-[#F59E0B]" />
          </div>
          <div className="text-2xl font-semibold text-white">{formatPercent(stats.conversionRate)}</div>
        </div>

        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-[#9CA3AF]">Total Rewards Paid</span>
            <DollarSign size={18} className="text-[#10B981]" />
          </div>
          <div className="text-2xl font-semibold text-white">{formatCurrency(stats.totalRewardsPaid)}</div>
        </div>

        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-[#9CA3AF]">Avg Reward / Referrer</span>
            <TrendingUp size={18} className="text-[#F5C542]" />
          </div>
          <div className="text-2xl font-semibold text-white">{formatCurrency(stats.avgRewardPerReferrer)}</div>
        </div>
      </div>
    );
  };

  // ============================================
  // Render Referral Link List Table
  // ============================================

  const renderLinkList = () => {
    return (
      <div className="bg-[#1E2026] border border-[#383A42] rounded mb-6">
        <div className="p-4 border-b border-[#383A42]">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <LinkIcon size={20} className="text-[#F5C542]" />
              <h2 className="text-lg font-semibold text-white">Referral Links ({filteredLinks.length})</h2>
            </div>
            <button
              onClick={fetchAllData}
              disabled={isLoading}
              className="px-3 py-1.5 bg-[#383A42] text-white rounded hover:bg-[#4A4C54] transition-colors disabled:opacity-50"
            >
              <RefreshCw size={16} className={isLoading ? 'animate-spin' : ''} />
            </button>
          </div>

          <div className="flex gap-4">
            <div className="flex-1 relative">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#9CA3AF]" />
              <input
                type="text"
                placeholder="Search by code or client name..."
                value={linkSearch}
                onChange={(e) => setLinkSearch(e.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-[#121316] border border-[#383A42] rounded text-white placeholder-[#6B7280] focus:outline-none focus:border-[#F5C542]"
              />
            </div>
            <select
              value={linkSortBy}
              onChange={(e) => setLinkSortBy(e.target.value as typeof linkSortBy)}
              className="px-4 py-2 bg-[#121316] border border-[#383A42] rounded text-white focus:outline-none focus:border-[#F5C542]"
            >
              <option value="conversions">Sort by Conversions</option>
              <option value="clicks">Sort by Clicks</option>
              <option value="registrations">Sort by Registrations</option>
              <option value="reward">Sort by Total Reward</option>
            </select>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-[#121316]">
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Code</th>
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Client</th>
                <th className="text-right p-3 text-sm font-medium text-[#9CA3AF]">Clicks</th>
                <th className="text-right p-3 text-sm font-medium text-[#9CA3AF]">Registrations</th>
                <th className="text-right p-3 text-sm font-medium text-[#9CA3AF]">Conversions</th>
                <th className="text-right p-3 text-sm font-medium text-[#9CA3AF]">Total Reward</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Status</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredLinks.map((link) => (
                <React.Fragment key={link.id}>
                  <tr
                    onClick={() => handleLinkClick(link)}
                    className="border-t border-[#383A42] hover:bg-[#252830] cursor-pointer transition-colors"
                  >
                    <td className="p-3 text-white font-mono text-sm">{link.code}</td>
                    <td className="p-3 text-white">{link.clientName}</td>
                    <td className="p-3 text-right text-white">{link.clicks}</td>
                    <td className="p-3 text-right text-white">{link.registrations}</td>
                    <td className="p-3 text-right text-white font-semibold">{link.conversions}</td>
                    <td className="p-3 text-right text-[#10B981] font-semibold">{formatCurrency(link.totalReward)}</td>
                    <td className="p-3 text-center">
                      <span
                        className="px-2 py-1 rounded text-xs font-medium"
                        style={{
                          backgroundColor: link.isActive ? 'rgba(16, 185, 129, 0.1)' : 'rgba(107, 114, 128, 0.1)',
                          color: link.isActive ? '#10B981' : '#6B7280'
                        }}
                      >
                        {link.isActive ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="p-3 text-center">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleToggleLinkActive(link.id, link.isActive);
                        }}
                        className="px-3 py-1 bg-[#383A42] text-white rounded text-xs hover:bg-[#4A4C54] transition-colors"
                      >
                        {link.isActive ? 'Deactivate' : 'Activate'}
                      </button>
                    </td>
                  </tr>
                  {selectedLink?.id === link.id && (
                    <tr>
                      <td colSpan={8} className="p-0">
                        <div className="bg-[#121316] p-4 border-t border-[#383A42]">
                          <h3 className="text-sm font-semibold text-white mb-3">Referrals for {link.code} ({linkReferrals.length})</h3>
                          {linkReferrals.length === 0 ? (
                            <p className="text-sm text-[#9CA3AF]">No referrals yet</p>
                          ) : (
                            <div className="space-y-2">
                              {linkReferrals.map((ref) => (
                                <div key={ref.id} className="flex items-center justify-between bg-[#1E2026] p-3 rounded">
                                  <div className="flex-1">
                                    <div className="text-sm text-white">{ref.referredName}</div>
                                    <div className="text-xs text-[#9CA3AF]">Registered: {new Date(ref.registeredAt).toLocaleDateString()}</div>
                                  </div>
                                  <div className="flex items-center gap-3">
                                    <span
                                      className="px-2 py-1 rounded text-xs font-medium"
                                      style={{
                                        backgroundColor: `${getStatusColor(ref.status)}20`,
                                        color: getStatusColor(ref.status)
                                      }}
                                    >
                                      {ref.status}
                                    </span>
                                    <span className="text-sm text-white">{formatCurrency(ref.reward)}</span>
                                    <span
                                      className="px-2 py-1 rounded text-xs font-medium"
                                      style={{
                                        backgroundColor: `${getRewardStatusColor(ref.rewardStatus)}20`,
                                        color: getRewardStatusColor(ref.rewardStatus)
                                      }}
                                    >
                                      {ref.rewardStatus}
                                    </span>
                                  </div>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  };

  // ============================================
  // Render Referral List Table
  // ============================================

  const renderReferralList = () => {
    return (
      <div className="bg-[#1E2026] border border-[#383A42] rounded mb-6">
        <div className="p-4 border-b border-[#383A42]">
          <div className="flex items-center gap-2 mb-4">
            <Users size={20} className="text-[#F5C542]" />
            <h2 className="text-lg font-semibold text-white">All Referrals ({filteredReferrals.length})</h2>
          </div>

          <div className="flex gap-4">
            <div className="flex-1 relative">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#9CA3AF]" />
              <input
                type="text"
                placeholder="Search by referrer or referred name..."
                value={referralSearch}
                onChange={(e) => setReferralSearch(e.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-[#121316] border border-[#383A42] rounded text-white placeholder-[#6B7280] focus:outline-none focus:border-[#F5C542]"
              />
            </div>
            <select
              value={referralStatusFilter}
              onChange={(e) => setReferralStatusFilter(e.target.value as typeof referralStatusFilter)}
              className="px-4 py-2 bg-[#121316] border border-[#383A42] rounded text-white focus:outline-none focus:border-[#F5C542]"
            >
              <option value="all">All Statuses</option>
              <option value="pending">Pending</option>
              <option value="registered">Registered</option>
              <option value="deposited">Deposited</option>
              <option value="traded">Traded</option>
              <option value="qualified">Qualified</option>
              <option value="expired">Expired</option>
            </select>
          </div>
        </div>

        <div className="overflow-x-auto max-h-[400px] overflow-y-auto">
          <table className="w-full">
            <thead className="sticky top-0 bg-[#121316]">
              <tr>
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Referrer</th>
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Referred</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Status</th>
                <th className="text-right p-3 text-sm font-medium text-[#9CA3AF]">Reward</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Reward Status</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Registered</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredReferrals.map((ref) => (
                <tr key={ref.id} className="border-t border-[#383A42] hover:bg-[#252830] transition-colors">
                  <td className="p-3 text-white">{ref.referrerName}</td>
                  <td className="p-3 text-white">{ref.referredName}</td>
                  <td className="p-3 text-center">
                    <span
                      className="px-2 py-1 rounded text-xs font-medium"
                      style={{
                        backgroundColor: `${getStatusColor(ref.status)}20`,
                        color: getStatusColor(ref.status)
                      }}
                    >
                      {ref.status}
                    </span>
                  </td>
                  <td className="p-3 text-right text-white font-semibold">{formatCurrency(ref.reward)}</td>
                  <td className="p-3 text-center">
                    <span
                      className="px-2 py-1 rounded text-xs font-medium"
                      style={{
                        backgroundColor: `${getRewardStatusColor(ref.rewardStatus)}20`,
                        color: getRewardStatusColor(ref.rewardStatus)
                      }}
                    >
                      {ref.rewardStatus}
                    </span>
                  </td>
                  <td className="p-3 text-center text-sm text-[#9CA3AF]">
                    {new Date(ref.registeredAt).toLocaleDateString()}
                  </td>
                  <td className="p-3 text-center">
                    {ref.status === 'qualified' && ref.rewardStatus === 'pending' && (
                      <button
                        onClick={() => {
                          setSelectedReferralForApproval(ref);
                          setShowApproveModal(true);
                        }}
                        className="px-3 py-1 bg-[#10B981] text-white rounded text-xs hover:bg-[#059669] transition-colors"
                      >
                        Approve Reward
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  };

  // ============================================
  // Render Tier Configuration Cards
  // ============================================

  const renderTierCards = () => {
    return (
      <div className="mb-6">
        <div className="flex items-center gap-2 mb-4">
          <Award size={20} className="text-[#F5C542]" />
          <h2 className="text-lg font-semibold text-white">Referral Tiers</h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {tiers.map((tier) => (
            <div key={tier.id} className="bg-[#1E2026] border border-[#383A42] rounded p-4">
              <div className="flex items-center justify-between mb-3">
                <h3
                  className="text-lg font-semibold"
                  style={{ color: getTierColor(tier.name) }}
                >
                  {tier.name}
                </h3>
                <span
                  className="px-2 py-1 rounded text-xs font-medium"
                  style={{
                    backgroundColor: tier.isActive ? 'rgba(16, 185, 129, 0.1)' : 'rgba(107, 114, 128, 0.1)',
                    color: tier.isActive ? '#10B981' : '#6B7280'
                  }}
                >
                  {tier.isActive ? 'Active' : 'Inactive'}
                </span>
              </div>

              <div className="space-y-2 mb-4">
                <div className="flex justify-between text-sm">
                  <span className="text-[#9CA3AF]">Range:</span>
                  <span className="text-white">
                    {tier.minReferrals}-{tier.maxReferrals === 999999 ? '∞' : tier.maxReferrals}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-[#9CA3AF]">Reward:</span>
                  <span className="text-[#10B981] font-semibold">{formatCurrency(tier.rewardPerReferral)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-[#9CA3AF]">Multiplier:</span>
                  <span className="text-[#F5C542] font-semibold">{tier.bonusMultiplier}x</span>
                </div>
              </div>

              <button
                onClick={() => handleUpdateTier(tier.id, { isActive: !tier.isActive })}
                className="w-full px-3 py-1.5 bg-[#383A42] text-white rounded text-sm hover:bg-[#4A4C54] transition-colors"
              >
                {tier.isActive ? 'Deactivate' : 'Activate'}
              </button>
            </div>
          ))}
        </div>
      </div>
    );
  };

  // ============================================
  // Render Top Referrers Leaderboard
  // ============================================

  const renderTopReferrers = () => {
    return (
      <div className="bg-[#1E2026] border border-[#383A42] rounded mb-6">
        <div className="p-4 border-b border-[#383A42]">
          <div className="flex items-center gap-2">
            <Target size={20} className="text-[#F5C542]" />
            <h2 className="text-lg font-semibold text-white">Top Referrers</h2>
          </div>
        </div>

        <div className="overflow-x-auto max-h-[400px] overflow-y-auto">
          <table className="w-full">
            <thead className="sticky top-0 bg-[#121316]">
              <tr>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF] w-16">Rank</th>
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Referrer</th>
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Link Code</th>
                <th className="text-right p-3 text-sm font-medium text-[#9CA3AF]">Conversions</th>
                <th className="text-right p-3 text-sm font-medium text-[#9CA3AF]">Registrations</th>
                <th className="text-right p-3 text-sm font-medium text-[#9CA3AF]">Total Reward</th>
              </tr>
            </thead>
            <tbody>
              {topReferrers.map((referrer, index) => {
                const rank = index + 1;
                return (
                  <tr key={referrer.clientID} className="border-t border-[#383A42] hover:bg-[#252830] transition-colors">
                    <td className="p-3 text-center">
                      <div
                        className="inline-flex items-center justify-center w-8 h-8 rounded-full font-bold text-sm"
                        style={{
                          backgroundColor: `${getRankBadgeColor(rank)}20`,
                          color: getRankBadgeColor(rank)
                        }}
                      >
                        #{rank}
                      </div>
                    </td>
                    <td className="p-3 text-white font-semibold">{referrer.clientName}</td>
                    <td className="p-3 text-[#9CA3AF] font-mono text-sm">{referrer.linkCode}</td>
                    <td className="p-3 text-right text-white font-semibold">{referrer.conversions}</td>
                    <td className="p-3 text-right text-[#9CA3AF]">{referrer.registrations}</td>
                    <td className="p-3 text-right text-[#10B981] font-semibold">{formatCurrency(referrer.totalReward)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    );
  };

  // ============================================
  // Render Monthly Trend Chart (SVG Line Chart)
  // ============================================

  const renderMonthlyTrendChart = () => {
    if (monthlyTrends.length === 0) return null;

    const maxReferrals = Math.max(...monthlyTrends.map(t => t.referrals));
    const maxConversions = Math.max(...monthlyTrends.map(t => t.conversions));
    const maxRewards = Math.max(...monthlyTrends.map(t => t.rewardsPaid));
    const maxValue = Math.max(maxReferrals, maxConversions, maxRewards / 100); // Scale rewards down

    const chartWidth = 800;
    const chartHeight = 300;
    const padding = { top: 20, right: 20, bottom: 40, left: 50 };
    const innerWidth = chartWidth - padding.left - padding.right;
    const innerHeight = chartHeight - padding.top - padding.bottom;

    const xStep = innerWidth / (monthlyTrends.length - 1);

    const getY = (value: number) => {
      return padding.top + innerHeight - (value / maxValue) * innerHeight;
    };

    const referralsPath = monthlyTrends
      .map((trend, i) => {
        const x = padding.left + i * xStep;
        const y = getY(trend.referrals);
        return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
      })
      .join(' ');

    const conversionsPath = monthlyTrends
      .map((trend, i) => {
        const x = padding.left + i * xStep;
        const y = getY(trend.conversions);
        return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
      })
      .join(' ');

    const rewardsPath = monthlyTrends
      .map((trend, i) => {
        const x = padding.left + i * xStep;
        const y = getY(trend.rewardsPaid / 100); // Scale down
        return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
      })
      .join(' ');

    return (
      <div className="bg-[#1E2026] border border-[#383A42] rounded mb-6">
        <div className="p-4 border-b border-[#383A42]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Activity size={20} className="text-[#F5C542]" />
              <h2 className="text-lg font-semibold text-white">Monthly Trends (12 Months)</h2>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <div className="w-4 h-0.5 bg-[#3B82F6]"></div>
                <span className="text-xs text-[#9CA3AF]">Referrals</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-0.5 bg-[#10B981]"></div>
                <span className="text-xs text-[#9CA3AF]">Conversions</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-0.5 bg-[#F59E0B]"></div>
                <span className="text-xs text-[#9CA3AF]">Rewards ($100s)</span>
              </div>
            </div>
          </div>
        </div>

        <div className="p-4 overflow-x-auto">
          <svg width={chartWidth} height={chartHeight} className="mx-auto">
            {/* Grid lines */}
            {[0, 0.25, 0.5, 0.75, 1].map((ratio) => {
              const y = padding.top + innerHeight * (1 - ratio);
              return (
                <g key={ratio}>
                  <line
                    x1={padding.left}
                    y1={y}
                    x2={chartWidth - padding.right}
                    y2={y}
                    stroke="#383A42"
                    strokeWidth="1"
                  />
                  <text
                    x={padding.left - 10}
                    y={y + 4}
                    textAnchor="end"
                    fill="#9CA3AF"
                    fontSize="10"
                  >
                    {Math.round(maxValue * ratio)}
                  </text>
                </g>
              );
            })}

            {/* X-axis labels */}
            {monthlyTrends.map((trend, i) => {
              const x = padding.left + i * xStep;
              return (
                <text
                  key={trend.month}
                  x={x}
                  y={chartHeight - 10}
                  textAnchor="middle"
                  fill="#9CA3AF"
                  fontSize="10"
                >
                  {trend.month.split('-')[1]}
                </text>
              );
            })}

            {/* Lines */}
            <path d={referralsPath} fill="none" stroke="#3B82F6" strokeWidth="2" />
            <path d={conversionsPath} fill="none" stroke="#10B981" strokeWidth="2" />
            <path d={rewardsPath} fill="none" stroke="#F59E0B" strokeWidth="2" />

            {/* Data points */}
            {monthlyTrends.map((trend, i) => {
              const x = padding.left + i * xStep;
              return (
                <g key={trend.month}>
                  <circle cx={x} cy={getY(trend.referrals)} r="3" fill="#3B82F6" />
                  <circle cx={x} cy={getY(trend.conversions)} r="3" fill="#10B981" />
                  <circle cx={x} cy={getY(trend.rewardsPaid / 100)} r="3" fill="#F59E0B" />
                </g>
              );
            })}
          </svg>
        </div>
      </div>
    );
  };

  // ============================================
  // Render Approve Reward Modal
  // ============================================

  const renderApproveModal = () => {
    if (!showApproveModal || !selectedReferralForApproval) return null;

    return (
      <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-6 max-w-md w-full mx-4">
          <h2 className="text-xl font-semibold text-white mb-4">Approve Referral Reward</h2>

          <div className="space-y-3 mb-6">
            <div className="flex justify-between">
              <span className="text-[#9CA3AF]">Referrer:</span>
              <span className="text-white font-semibold">{selectedReferralForApproval.referrerName}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-[#9CA3AF]">Referred:</span>
              <span className="text-white font-semibold">{selectedReferralForApproval.referredName}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-[#9CA3AF]">Reward Amount:</span>
              <span className="text-[#10B981] font-semibold text-lg">{formatCurrency(selectedReferralForApproval.reward)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-[#9CA3AF]">Status:</span>
              <span
                className="px-2 py-1 rounded text-xs font-medium"
                style={{
                  backgroundColor: `${getStatusColor(selectedReferralForApproval.status)}20`,
                  color: getStatusColor(selectedReferralForApproval.status)
                }}
              >
                {selectedReferralForApproval.status}
              </span>
            </div>
          </div>

          <div className="bg-[#121316] border border-[#383A42] rounded p-3 mb-6">
            <p className="text-sm text-[#9CA3AF]">
              This will approve the reward and mark it as paid. The referrer will receive {formatCurrency(selectedReferralForApproval.reward)} in their account.
            </p>
          </div>

          <div className="flex gap-3">
            <button
              onClick={() => {
                setShowApproveModal(false);
                setSelectedReferralForApproval(null);
              }}
              className="flex-1 px-4 py-2 bg-[#383A42] text-white rounded hover:bg-[#4A4C54] transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleApproveReward}
              className="flex-1 px-4 py-2 bg-[#10B981] text-white rounded hover:bg-[#059669] transition-colors flex items-center justify-center gap-2"
            >
              <Check size={18} />
              Approve & Pay
            </button>
          </div>
        </div>
      </div>
    );
  };

  // ============================================
  // Main Render
  // ============================================

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white mb-2">Referral Program Management</h1>
        <p className="text-[#9CA3AF]">Manage referral links, track conversions, and configure reward tiers</p>
      </div>

      {renderStatsCards()}
      {renderTierCards()}
      {renderLinkList()}
      {renderReferralList()}
      {renderTopReferrers()}
      {renderMonthlyTrendChart()}
      {renderApproveModal()}
    </div>
  );
};

export default ReferralProgram;
