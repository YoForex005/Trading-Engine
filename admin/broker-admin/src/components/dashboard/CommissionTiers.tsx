'use client';

import React, { useState, useMemo } from 'react';
import {
  TrendingUp,
  Award,
  Users,
  DollarSign,
  ArrowUp,
  ArrowDown,
  Calculator,
  BarChart3,
  Search,
  Filter,
  ChevronDown,
  ChevronUp,
  Star,
  Crown,
  Zap,
  Target,
} from 'lucide-react';

// ============================================================================
// TYPES
// ============================================================================

type TierLevel = 'BRONZE' | 'SILVER' | 'GOLD' | 'PLATINUM' | 'VIP';

interface CommissionTier {
  id: string;
  level: TierLevel;
  name: string;
  minVolume: number; // Monthly trading volume required
  commissionRate: number; // Percentage
  rebateRate: number; // Percentage
  color: string;
  icon: React.ReactNode;
  features: string[];
  clientCount: number;
}

interface ClientTierAssignment {
  id: string;
  clientId: string;
  clientName: string;
  currentTier: TierLevel;
  monthlyVolume: number;
  totalCommissions: number;
  nextTier: TierLevel | null;
  progressToNextTier: number; // 0-100
  joinDate: number;
}

interface CommissionTransaction {
  id: string;
  clientId: string;
  clientName: string;
  tier: TierLevel;
  amount: number;
  type: 'COMMISSION' | 'REBATE';
  volume: number;
  timestamp: number;
}

interface CommissionStats {
  totalCommissions: number;
  totalRebates: number;
  avgCommissionRate: number;
  totalClients: number;
  topEarners: Array<{ clientName: string; amount: number }>;
}

// ============================================================================
// TIER CONFIGURATIONS
// ============================================================================

const TIER_CONFIGS: CommissionTier[] = [
  {
    id: 'bronze',
    level: 'BRONZE',
    name: 'Bronze',
    minVolume: 0,
    commissionRate: 0.05,
    rebateRate: 0.01,
    color: '#CD7F32',
    icon: <Award size={24} className="text-[#CD7F32]" />,
    features: ['Basic commission rate', 'Monthly reports', 'Email support'],
    clientCount: 0,
  },
  {
    id: 'silver',
    level: 'SILVER',
    name: 'Silver',
    minVolume: 100000,
    commissionRate: 0.08,
    rebateRate: 0.015,
    color: '#C0C0C0',
    icon: <Star size={24} className="text-[#C0C0C0]" />,
    features: ['Enhanced commission rate', 'Weekly reports', 'Priority support', '5% volume bonus'],
    clientCount: 0,
  },
  {
    id: 'gold',
    level: 'GOLD',
    name: 'Gold',
    minVolume: 500000,
    commissionRate: 0.12,
    rebateRate: 0.025,
    color: '#FFD700',
    icon: <Target size={24} className="text-[#FFD700]" />,
    features: ['Premium commission rate', 'Daily reports', 'Dedicated support', '10% volume bonus', 'Marketing materials'],
    clientCount: 0,
  },
  {
    id: 'platinum',
    level: 'PLATINUM',
    name: 'Platinum',
    minVolume: 2000000,
    commissionRate: 0.18,
    rebateRate: 0.04,
    color: '#E5E4E2',
    icon: <Zap size={24} className="text-[#E5E4E2]" />,
    features: [
      'Elite commission rate',
      'Real-time reporting',
      'VIP support',
      '15% volume bonus',
      'Custom marketing',
      'Early access to new features',
    ],
    clientCount: 0,
  },
  {
    id: 'vip',
    level: 'VIP',
    name: 'VIP',
    minVolume: 10000000,
    commissionRate: 0.25,
    rebateRate: 0.06,
    color: '#9B59B6',
    icon: <Crown size={24} className="text-[#9B59B6]" />,
    features: [
      'Maximum commission rate',
      'Custom reporting',
      'Personal account manager',
      '20% volume bonus',
      'White-label options',
      'Revenue sharing',
      'Exclusive events',
    ],
    clientCount: 0,
  },
];

// ============================================================================
// MOCK DATA GENERATORS
// ============================================================================

function generateMockClients(count: number): ClientTierAssignment[] {
  const clients: ClientTierAssignment[] = [];
  const now = Date.now();

  for (let i = 0; i < count; i++) {
    const tierIndex = Math.floor(Math.random() * TIER_CONFIGS.length);
    const currentTier = TIER_CONFIGS[tierIndex];
    const nextTier = tierIndex < TIER_CONFIGS.length - 1 ? TIER_CONFIGS[tierIndex + 1] : null;

    const monthlyVolume = currentTier.minVolume + Math.random() * (nextTier ? nextTier.minVolume - currentTier.minVolume : 5000000);
    const totalCommissions = monthlyVolume * (currentTier.commissionRate / 100);

    const progressToNextTier = nextTier
      ? ((monthlyVolume - currentTier.minVolume) / (nextTier.minVolume - currentTier.minVolume)) * 100
      : 100;

    clients.push({
      id: `client-${i + 1}`,
      clientId: `CL${10000 + i}`,
      clientName: `Client ${String.fromCharCode(65 + (i % 26))}${Math.floor(i / 26) + 1}`,
      currentTier: currentTier.level,
      monthlyVolume,
      totalCommissions,
      nextTier: nextTier?.level || null,
      progressToNextTier: Math.min(100, progressToNextTier),
      joinDate: now - Math.random() * 365 * 24 * 60 * 60 * 1000,
    });
  }

  return clients;
}

function generateMockTransactions(count: number, clients: ClientTierAssignment[]): CommissionTransaction[] {
  const transactions: CommissionTransaction[] = [];
  const now = Date.now();

  for (let i = 0; i < count; i++) {
    const client = clients[Math.floor(Math.random() * clients.length)];
    const type = Math.random() > 0.3 ? 'COMMISSION' : 'REBATE';
    const volume = Math.random() * 50000 + 5000;
    const tier = TIER_CONFIGS.find((t) => t.level === client.currentTier)!;
    const rate = type === 'COMMISSION' ? tier.commissionRate : tier.rebateRate;
    const amount = volume * (rate / 100);

    transactions.push({
      id: `txn-${i + 1}`,
      clientId: client.clientId,
      clientName: client.clientName,
      tier: client.currentTier,
      amount,
      type,
      volume,
      timestamp: now - Math.random() * 30 * 24 * 60 * 60 * 1000,
    });
  }

  return transactions.sort((a, b) => b.timestamp - a.timestamp);
}

function calculateStats(clients: ClientTierAssignment[], transactions: CommissionTransaction[]): CommissionStats {
  const commissions = transactions.filter((t) => t.type === 'COMMISSION');
  const rebates = transactions.filter((t) => t.type === 'REBATE');

  const totalCommissions = commissions.reduce((sum, t) => sum + t.amount, 0);
  const totalRebates = rebates.reduce((sum, t) => sum + t.amount, 0);
  const avgCommissionRate =
    clients.reduce((sum, c) => {
      const tier = TIER_CONFIGS.find((t) => t.level === c.currentTier);
      return sum + (tier?.commissionRate || 0);
    }, 0) / clients.length;

  const clientEarnings = new Map<string, number>();
  transactions.forEach((t) => {
    clientEarnings.set(t.clientName, (clientEarnings.get(t.clientName) || 0) + t.amount);
  });

  const topEarners = Array.from(clientEarnings.entries())
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5)
    .map(([clientName, amount]) => ({ clientName, amount }));

  return {
    totalCommissions,
    totalRebates,
    avgCommissionRate,
    totalClients: clients.length,
    topEarners,
  };
}

// ============================================================================
// MAIN COMPONENT
// ============================================================================

export default function CommissionTiers() {
  const [clients, setClients] = useState<ClientTierAssignment[]>(() => generateMockClients(60));
  const [transactions, setTransactions] = useState<CommissionTransaction[]>(() =>
    generateMockTransactions(200, generateMockClients(60))
  );
  const [selectedTier, setSelectedTier] = useState<TierLevel | 'ALL'>('ALL');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortField, setSortField] = useState<keyof ClientTierAssignment>('totalCommissions');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
  const [calculatorVolume, setCalculatorVolume] = useState<number>(100000);
  const [calculatorTier, setCalculatorTier] = useState<TierLevel>('SILVER');

  const stats = useMemo(() => calculateStats(clients, transactions), [clients, transactions]);

  // Update tier client counts
  const tiersWithCounts = useMemo(() => {
    const counts = new Map<TierLevel, number>();
    clients.forEach((c) => {
      counts.set(c.currentTier, (counts.get(c.currentTier) || 0) + 1);
    });
    return TIER_CONFIGS.map((tier) => ({
      ...tier,
      clientCount: counts.get(tier.level) || 0,
    }));
  }, [clients]);

  // Filter and sort clients
  const filteredClients = useMemo(() => {
    let result = clients.filter((client) => {
      if (selectedTier !== 'ALL' && client.currentTier !== selectedTier) return false;
      if (searchQuery && !client.clientName.toLowerCase().includes(searchQuery.toLowerCase())) return false;
      return true;
    });

    result.sort((a, b) => {
      const aVal = a[sortField];
      const bVal = b[sortField];
      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
      }
      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortDirection === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
      }
      return 0;
    });

    return result;
  }, [clients, selectedTier, searchQuery, sortField, sortDirection]);

  const handleSort = (field: keyof ClientTierAssignment) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const handleTierChange = (clientId: string, newTier: TierLevel) => {
    setClients((prev) =>
      prev.map((c) => {
        if (c.id !== clientId) return c;
        const tier = TIER_CONFIGS.find((t) => t.level === newTier)!;
        return {
          ...c,
          currentTier: newTier,
          totalCommissions: c.monthlyVolume * (tier.commissionRate / 100),
        };
      })
    );
  };

  const formatCurrency = (amount: number): string => {
    return `$${amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  };

  const formatDate = (timestamp: number): string => {
    return new Date(timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const getTierColor = (tier: TierLevel): string => {
    return TIER_CONFIGS.find((t) => t.level === tier)?.color || '#666';
  };

  // Commission calculator
  const calculatedCommission = useMemo(() => {
    const tier = TIER_CONFIGS.find((t) => t.level === calculatorTier);
    if (!tier) return 0;
    return calculatorVolume * (tier.commissionRate / 100);
  }, [calculatorVolume, calculatorTier]);

  const calculatedRebate = useMemo(() => {
    const tier = TIER_CONFIGS.find((t) => t.level === calculatorTier);
    if (!tier) return 0;
    return calculatorVolume * (tier.rebateRate / 100);
  }, [calculatorVolume, calculatorTier]);

  // Tier comparison chart
  const renderTierComparisonChart = () => {
    const maxRate = Math.max(...TIER_CONFIGS.map((t) => t.commissionRate));
    const barWidth = 60;
    const gap = 20;
    const chartHeight = 200;

    return (
      <svg width={TIER_CONFIGS.length * (barWidth + gap)} height={chartHeight + 40} className="mx-auto">
        {TIER_CONFIGS.map((tier, index) => {
          const barHeight = (tier.commissionRate / maxRate) * chartHeight;
          const x = index * (barWidth + gap);
          const y = chartHeight - barHeight;

          return (
            <g key={tier.id}>
              <rect x={x} y={y} width={barWidth} height={barHeight} fill={tier.color} opacity={0.8} />
              <text x={x + barWidth / 2} y={chartHeight + 20} textAnchor="middle" fill="#999" fontSize="11">
                {tier.name}
              </text>
              <text x={x + barWidth / 2} y={y - 5} textAnchor="middle" fill="#CCC" fontSize="12" fontWeight="bold">
                {tier.commissionRate}%
              </text>
            </g>
          );
        })}
      </svg>
    );
  };

  // Rebate distribution breakdown
  const rebateDistribution = useMemo(() => {
    const distribution = new Map<TierLevel, number>();
    TIER_CONFIGS.forEach((tier) => distribution.set(tier.level, 0));

    transactions
      .filter((t) => t.type === 'REBATE')
      .forEach((t) => {
        distribution.set(t.tier, (distribution.get(t.tier) || 0) + t.amount);
      });

    return TIER_CONFIGS.map((tier) => ({
      tier: tier.name,
      amount: distribution.get(tier.level) || 0,
      color: tier.color,
    }));
  }, [transactions]);

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#CCC] overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-[#383A42] flex-shrink-0">
        <div className="flex items-center gap-2">
          <TrendingUp size={18} className="text-[#10B981]" />
          <h1 className="text-sm font-bold text-white">Commission Tiers / Rebate Management</h1>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-5 gap-3 px-4 py-3 border-b border-[#383A42] flex-shrink-0">
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <DollarSign size={14} className="text-[#22c55e]" />
            <span className="text-xs text-[#666]">Total Commissions</span>
          </div>
          <div className="text-xl font-bold text-[#22c55e]">{formatCurrency(stats.totalCommissions)}</div>
        </div>
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <DollarSign size={14} className="text-[#3B82F6]" />
            <span className="text-xs text-[#666]">Total Rebates</span>
          </div>
          <div className="text-xl font-bold text-[#3B82F6]">{formatCurrency(stats.totalRebates)}</div>
        </div>
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <BarChart3 size={14} className="text-[#F5C542]" />
            <span className="text-xs text-[#666]">Avg Commission Rate</span>
          </div>
          <div className="text-xl font-bold text-[#F5C542]">{stats.avgCommissionRate.toFixed(2)}%</div>
        </div>
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <Users size={14} className="text-[#9B59B6]" />
            <span className="text-xs text-[#666]">Total Clients</span>
          </div>
          <div className="text-xl font-bold text-white">{stats.totalClients}</div>
        </div>
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="text-xs text-[#666] mb-2">Top Earner</div>
          {stats.topEarners[0] && (
            <>
              <div className="text-xs text-white font-bold truncate">{stats.topEarners[0].clientName}</div>
              <div className="text-xs text-[#22c55e]">{formatCurrency(stats.topEarners[0].amount)}</div>
            </>
          )}
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        {/* Left Panel - Tier Cards & Calculator */}
        <div className="w-80 flex flex-col border-r border-[#383A42] overflow-y-auto custom-scrollbar">
          {/* Tier Configuration Cards */}
          <div className="px-4 py-3 border-b border-[#383A42] bg-[#1E2026]">
            <h2 className="text-xs font-bold text-white">Commission Tiers</h2>
          </div>
          <div className="p-3 space-y-3">
            {tiersWithCounts.map((tier) => (
              <div
                key={tier.id}
                className="bg-[#1E2026] rounded border border-[#383A42] p-3 hover:border-[#505258] cursor-pointer transition-colors"
                onClick={() => setSelectedTier(tier.level)}
                style={{ borderLeftWidth: '4px', borderLeftColor: tier.color }}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    {tier.icon}
                    <span className="text-sm font-bold text-white">{tier.name}</span>
                  </div>
                  <div className="text-xs text-[#666]">{tier.clientCount} clients</div>
                </div>
                <div className="grid grid-cols-2 gap-2 mb-2">
                  <div>
                    <div className="text-xs text-[#666]">Commission</div>
                    <div className="text-sm font-bold text-[#22c55e]">{tier.commissionRate}%</div>
                  </div>
                  <div>
                    <div className="text-xs text-[#666]">Rebate</div>
                    <div className="text-sm font-bold text-[#3B82F6]">{tier.rebateRate}%</div>
                  </div>
                </div>
                <div className="text-xs text-[#666] mb-1">Min Volume: {formatCurrency(tier.minVolume)}</div>
                <div className="text-xs text-[#999] space-y-1">
                  {tier.features.slice(0, 3).map((feature, idx) => (
                    <div key={idx}>• {feature}</div>
                  ))}
                  {tier.features.length > 3 && <div className="text-[#666]">+{tier.features.length - 3} more</div>}
                </div>
              </div>
            ))}
          </div>

          {/* Commission Calculator */}
          <div className="px-4 py-3 border-b border-[#383A42] bg-[#1E2026]">
            <div className="flex items-center gap-2">
              <Calculator size={14} className="text-[#10B981]" />
              <h2 className="text-xs font-bold text-white">Commission Calculator</h2>
            </div>
          </div>
          <div className="p-4 bg-[#1E2026] border-b border-[#383A42]">
            <div className="mb-3">
              <label className="text-xs text-[#666] mb-1 block">Trading Volume</label>
              <input
                type="number"
                value={calculatorVolume}
                onChange={(e) => setCalculatorVolume(Number(e.target.value))}
                className="w-full bg-[#121316] text-xs text-white px-2 py-2 rounded border border-[#383A42] outline-none"
                min="0"
                step="1000"
              />
            </div>
            <div className="mb-3">
              <label className="text-xs text-[#666] mb-1 block">Tier Level</label>
              <select
                value={calculatorTier}
                onChange={(e) => setCalculatorTier(e.target.value as TierLevel)}
                className="w-full bg-[#121316] text-xs text-white px-2 py-2 rounded border border-[#383A42] outline-none"
              >
                {TIER_CONFIGS.map((tier) => (
                  <option key={tier.id} value={tier.level}>
                    {tier.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="bg-[#121316] rounded p-3 border border-[#383A42]">
              <div className="mb-2">
                <div className="text-xs text-[#666]">Commission Earned</div>
                <div className="text-lg font-bold text-[#22c55e]">{formatCurrency(calculatedCommission)}</div>
              </div>
              <div>
                <div className="text-xs text-[#666]">Rebate Amount</div>
                <div className="text-lg font-bold text-[#3B82F6]">{formatCurrency(calculatedRebate)}</div>
              </div>
            </div>
          </div>

          {/* Rebate Distribution */}
          <div className="px-4 py-3 border-b border-[#383A42] bg-[#1E2026]">
            <h2 className="text-xs font-bold text-white">Rebate Distribution</h2>
          </div>
          <div className="p-4 space-y-2">
            {rebateDistribution.map((item) => (
              <div key={item.tier} className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded" style={{ backgroundColor: item.color }} />
                  <span className="text-xs text-[#CCC]">{item.tier}</span>
                </div>
                <span className="text-xs font-bold text-white">{formatCurrency(item.amount)}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Main Content Area */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Filter Bar */}
          <div className="flex items-center gap-3 px-4 py-2 border-b border-[#383A42] bg-[#1E2026] flex-shrink-0">
            <div className="flex items-center gap-2 bg-[#121316] px-2 py-1 rounded border border-[#383A42] flex-1 max-w-xs">
              <Search size={14} className="text-[#666]" />
              <input
                type="text"
                placeholder="Search clients..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="bg-transparent text-xs text-white outline-none flex-1"
              />
            </div>

            <select
              value={selectedTier}
              onChange={(e) => setSelectedTier(e.target.value as TierLevel | 'ALL')}
              className="bg-[#121316] text-xs text-[#CCC] px-2 py-1 rounded border border-[#383A42] outline-none"
            >
              <option value="ALL">All Tiers</option>
              {TIER_CONFIGS.map((tier) => (
                <option key={tier.id} value={tier.level}>
                  {tier.name}
                </option>
              ))}
            </select>

            <button
              onClick={() => {
                setSelectedTier('ALL');
                setSearchQuery('');
              }}
              className="px-3 py-1 text-xs bg-[#383A42] text-white rounded hover:bg-[#505258]"
            >
              Clear Filters
            </button>
          </div>

          {/* Client Tier Assignment Table */}
          <div className="flex-1 overflow-y-auto custom-scrollbar">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                <tr>
                  <th
                    className="text-left px-3 py-2 font-bold text-[#999] cursor-pointer hover:text-white"
                    onClick={() => handleSort('clientName')}
                  >
                    <div className="flex items-center gap-1">
                      Client
                      {sortField === 'clientName' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th
                    className="text-left px-3 py-2 font-bold text-[#999] cursor-pointer hover:text-white"
                    onClick={() => handleSort('currentTier')}
                  >
                    <div className="flex items-center gap-1">
                      Current Tier
                      {sortField === 'currentTier' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th
                    className="text-left px-3 py-2 font-bold text-[#999] cursor-pointer hover:text-white"
                    onClick={() => handleSort('monthlyVolume')}
                  >
                    <div className="flex items-center gap-1">
                      Monthly Volume
                      {sortField === 'monthlyVolume' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th
                    className="text-left px-3 py-2 font-bold text-[#999] cursor-pointer hover:text-white"
                    onClick={() => handleSort('totalCommissions')}
                  >
                    <div className="flex items-center gap-1">
                      Total Commissions
                      {sortField === 'totalCommissions' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th className="text-left px-3 py-2 font-bold text-[#999]">Next Tier Progress</th>
                  <th className="text-left px-3 py-2 font-bold text-[#999]">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredClients.map((client) => {
                  const tierConfig = TIER_CONFIGS.find((t) => t.level === client.currentTier);
                  return (
                    <tr key={client.id} className="border-b border-[#383A42] hover:bg-[#1E2026]">
                      <td className="px-3 py-2">
                        <div className="text-white font-bold">{client.clientName}</div>
                        <div className="text-[#666] text-xs">{client.clientId}</div>
                      </td>
                      <td className="px-3 py-2">
                        <div
                          className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-bold"
                          style={{ backgroundColor: `${getTierColor(client.currentTier)}20`, color: getTierColor(client.currentTier) }}
                        >
                          {tierConfig?.icon}
                          {client.currentTier}
                        </div>
                      </td>
                      <td className="px-3 py-2 text-white">{formatCurrency(client.monthlyVolume)}</td>
                      <td className="px-3 py-2 text-[#22c55e] font-bold">{formatCurrency(client.totalCommissions)}</td>
                      <td className="px-3 py-2">
                        {client.nextTier ? (
                          <div>
                            <div className="flex items-center gap-2 mb-1">
                              <div className="flex-1 bg-[#383A42] rounded h-2 overflow-hidden">
                                <div
                                  className="bg-[#10B981] h-full transition-all"
                                  style={{ width: `${client.progressToNextTier}%` }}
                                />
                              </div>
                              <span className="text-xs text-[#666]">{Math.round(client.progressToNextTier)}%</span>
                            </div>
                            <div className="text-xs text-[#999]">Next: {client.nextTier}</div>
                          </div>
                        ) : (
                          <span className="text-xs text-[#666]">Max Tier</span>
                        )}
                      </td>
                      <td className="px-3 py-2">
                        <div className="flex items-center gap-1">
                          {client.nextTier && (
                            <button
                              onClick={() => handleTierChange(client.id, client.nextTier!)}
                              className="px-2 py-1 text-xs bg-[#22c55e] text-white rounded hover:bg-[#16a34a] flex items-center gap-1"
                              title="Upgrade Tier"
                            >
                              <ArrowUp size={12} />
                            </button>
                          )}
                          {client.currentTier !== 'BRONZE' && (
                            <button
                              onClick={() => {
                                const currentIndex = TIER_CONFIGS.findIndex((t) => t.level === client.currentTier);
                                if (currentIndex > 0) {
                                  handleTierChange(client.id, TIER_CONFIGS[currentIndex - 1].level);
                                }
                              }}
                              className="px-2 py-1 text-xs bg-[#E74C3C] text-white rounded hover:bg-[#C0392B] flex items-center gap-1"
                              title="Downgrade Tier"
                            >
                              <ArrowDown size={12} />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Tier Comparison Chart */}
          <div className="border-t border-[#383A42] bg-[#1E2026] p-4 flex-shrink-0">
            <h3 className="text-xs font-bold text-white mb-3">Tier Commission Rate Comparison</h3>
            {renderTierComparisonChart()}
          </div>
        </div>

        {/* Right Panel - Transaction History */}
        <div className="w-96 flex flex-col border-l border-[#383A42] overflow-hidden">
          <div className="px-4 py-2 border-b border-[#383A42] bg-[#1E2026] flex-shrink-0">
            <h2 className="text-xs font-bold text-white">Recent Transactions</h2>
          </div>
          <div className="flex-1 overflow-y-auto custom-scrollbar">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                <tr>
                  <th className="text-left px-3 py-2 font-bold text-[#999]">Date</th>
                  <th className="text-left px-3 py-2 font-bold text-[#999]">Client</th>
                  <th className="text-left px-3 py-2 font-bold text-[#999]">Type</th>
                  <th className="text-right px-3 py-2 font-bold text-[#999]">Amount</th>
                </tr>
              </thead>
              <tbody>
                {transactions.slice(0, 100).map((txn) => (
                  <tr key={txn.id} className="border-b border-[#383A42] hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#999]">{formatDate(txn.timestamp)}</td>
                    <td className="px-3 py-2">
                      <div className="text-white">{txn.clientName}</div>
                      <div
                        className="text-xs"
                        style={{ color: getTierColor(txn.tier) }}
                      >
                        {txn.tier}
                      </div>
                    </td>
                    <td className="px-3 py-2">
                      <span
                        className={`px-2 py-1 rounded text-xs ${
                          txn.type === 'COMMISSION'
                            ? 'bg-[#22c55e]/20 text-[#22c55e]'
                            : 'bg-[#3B82F6]/20 text-[#3B82F6]'
                        }`}
                      >
                        {txn.type}
                      </span>
                    </td>
                    <td className={`px-3 py-2 text-right font-bold ${txn.type === 'COMMISSION' ? 'text-[#22c55e]' : 'text-[#3B82F6]'}`}>
                      {formatCurrency(txn.amount)}
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
