'use client';

import React, { useState, useMemo, useEffect } from 'react';
import {
  Trophy,
  TrendingUp,
  Users,
  DollarSign,
  Calendar,
  Award,
  Clock,
  Plus,
  Eye,
  Settings,
  CheckCircle,
  XCircle,
  Medal,
  Activity,
  Target
} from 'lucide-react';

// Types
type CompetitionStatus = 'active' | 'upcoming' | 'past';
type TabFilter = 'active' | 'upcoming' | 'past';

interface Competition {
  id: string;
  name: string;
  description: string;
  status: CompetitionStatus;
  startDate: number;
  endDate: number;
  entryFee: number;
  prizePool: number;
  maxParticipants: number;
  participants: number;
  eligibleAccountTypes: string[];
  rules: string;
  prizeDistribution: { place: number; percentage: number }[];
  totalVolume: number;
  avgProfit: number;
}

interface Participant {
  id: string;
  competitionId: string;
  rank: number;
  traderName: string;
  accountId: string;
  profitPct: number;
  equityGrowth: number;
  tradesCount: number;
  winRate: number;
  prize: number;
  status: 'active' | 'disqualified';
  equityCurve: { timestamp: number; equity: number }[]; // For chart
}

// Mock Data - 8 competitions
const MOCK_COMPETITIONS: Competition[] = [
  // Active competitions (3)
  {
    id: 'comp-1',
    name: 'January Forex Challenge',
    description: 'Trade major forex pairs for a chance to win',
    status: 'active',
    startDate: Date.now() - 15 * 86400000, // Started 15 days ago
    endDate: Date.now() + 15 * 86400000, // Ends in 15 days
    entryFee: 0,
    prizePool: 10000,
    maxParticipants: 100,
    participants: 45,
    eligibleAccountTypes: ['Standard', 'ECN', 'VIP'],
    rules: 'Trade any forex pairs. No scalping. Minimum 10 trades required.',
    prizeDistribution: [
      { place: 1, percentage: 50 },
      { place: 2, percentage: 30 },
      { place: 3, percentage: 20 }
    ],
    totalVolume: 245680000,
    avgProfit: 12.5
  },
  {
    id: 'comp-2',
    name: 'Crypto Trading Cup',
    description: 'BTC, ETH, and major altcoins competition',
    status: 'active',
    startDate: Date.now() - 7 * 86400000,
    endDate: Date.now() + 23 * 86400000,
    entryFee: 50,
    prizePool: 25000,
    maxParticipants: 150,
    participants: 120,
    eligibleAccountTypes: ['ECN', 'VIP'],
    rules: 'Crypto pairs only. Max leverage 1:10. No weekend gap trading.',
    prizeDistribution: [
      { place: 1, percentage: 40 },
      { place: 2, percentage: 25 },
      { place: 3, percentage: 15 },
      { place: 4, percentage: 10 },
      { place: 5, percentage: 10 }
    ],
    totalVolume: 892340000,
    avgProfit: 18.3
  },
  {
    id: 'comp-3',
    name: 'Scalping Sprint',
    description: 'High-frequency trading competition',
    status: 'active',
    startDate: Date.now() - 3 * 86400000,
    endDate: Date.now() + 4 * 86400000, // Ends in 4 days
    entryFee: 25,
    prizePool: 5000,
    maxParticipants: 50,
    participants: 28,
    eligibleAccountTypes: ['ECN'],
    rules: 'Minimum 50 trades. Max hold time 30 minutes. EUR/USD only.',
    prizeDistribution: [
      { place: 1, percentage: 60 },
      { place: 2, percentage: 25 },
      { place: 3, percentage: 15 }
    ],
    totalVolume: 123450000,
    avgProfit: 8.7
  },

  // Upcoming competitions (2)
  {
    id: 'comp-4',
    name: 'February Grand Prix',
    description: 'Month-long trading championship',
    status: 'upcoming',
    startDate: Date.now() + 5 * 86400000, // Starts in 5 days
    endDate: Date.now() + 35 * 86400000,
    entryFee: 100,
    prizePool: 50000,
    maxParticipants: 200,
    participants: 0,
    eligibleAccountTypes: ['VIP'],
    rules: 'All instruments allowed. Minimum account balance $10,000. Professional traders only.',
    prizeDistribution: [
      { place: 1, percentage: 35 },
      { place: 2, percentage: 20 },
      { place: 3, percentage: 15 },
      { place: 4, percentage: 10 },
      { place: 5, percentage: 8 },
      { place: 6, percentage: 7 },
      { place: 7, percentage: 5 }
    ],
    totalVolume: 0,
    avgProfit: 0
  },
  {
    id: 'comp-5',
    name: 'Gold Trading Masters',
    description: 'Precious metals trading challenge',
    status: 'upcoming',
    startDate: Date.now() + 12 * 86400000,
    endDate: Date.now() + 26 * 86400000,
    entryFee: 0,
    prizePool: 15000,
    maxParticipants: 100,
    participants: 0,
    eligibleAccountTypes: ['Standard', 'ECN', 'VIP'],
    rules: 'XAU/USD and XAG/USD only. No hedging allowed. Max 5 open positions.',
    prizeDistribution: [
      { place: 1, percentage: 50 },
      { place: 2, percentage: 30 },
      { place: 3, percentage: 20 }
    ],
    totalVolume: 0,
    avgProfit: 0
  },

  // Past competitions (3)
  {
    id: 'comp-6',
    name: 'Q4 Championship',
    description: 'Year-end trading championship',
    status: 'past',
    startDate: Date.now() - 90 * 86400000,
    endDate: Date.now() - 30 * 86400000,
    entryFee: 200,
    prizePool: 100000,
    maxParticipants: 250,
    participants: 187,
    eligibleAccountTypes: ['VIP'],
    rules: 'All instruments. Minimum 30 trades. Verified accounts only.',
    prizeDistribution: [
      { place: 1, percentage: 30 },
      { place: 2, percentage: 20 },
      { place: 3, percentage: 15 },
      { place: 4, percentage: 10 },
      { place: 5, percentage: 8 },
      { place: 6, percentage: 7 },
      { place: 7, percentage: 5 },
      { place: 8, percentage: 5 }
    ],
    totalVolume: 1245780000,
    avgProfit: 15.8
  },
  {
    id: 'comp-7',
    name: 'Holiday Trading Contest',
    description: 'December holiday special',
    status: 'past',
    startDate: Date.now() - 60 * 86400000,
    endDate: Date.now() - 15 * 86400000,
    entryFee: 0,
    prizePool: 20000,
    maxParticipants: 150,
    participants: 142,
    eligibleAccountTypes: ['Standard', 'ECN', 'VIP'],
    rules: 'Forex pairs only. Festive spirit required!',
    prizeDistribution: [
      { place: 1, percentage: 40 },
      { place: 2, percentage: 25 },
      { place: 3, percentage: 20 },
      { place: 4, percentage: 15 }
    ],
    totalVolume: 567890000,
    avgProfit: 11.2
  },
  {
    id: 'comp-8',
    name: 'Beginners Challenge',
    description: 'Competition for new traders',
    status: 'past',
    startDate: Date.now() - 45 * 86400000,
    endDate: Date.now() - 10 * 86400000,
    entryFee: 10,
    prizePool: 3000,
    maxParticipants: 80,
    participants: 67,
    eligibleAccountTypes: ['Standard'],
    rules: 'New accounts only (less than 6 months old). Max 3 lots per trade.',
    prizeDistribution: [
      { place: 1, percentage: 50 },
      { place: 2, percentage: 30 },
      { place: 3, percentage: 20 }
    ],
    totalVolume: 89450000,
    avgProfit: 6.4
  }
];

// Generate mock participants for a competition
const generateMockParticipants = (competitionId: string, count: number): Participant[] => {
  const traders = [
    'Alex Chen', 'Maria Rodriguez', 'John Smith', 'Sarah Kim', 'David Lee',
    'Emma Watson', 'Michael Chang', 'Lisa Anderson', 'Tom Wilson', 'Anna Garcia',
    'Chris Taylor', 'Jessica Brown', 'Kevin Wang', 'Nina Patel', 'Ryan O\'Brien',
    'Sophie Martinez', 'Daniel Park', 'Olivia Thompson', 'James Miller', 'Emily Davis'
  ];

  const participants: Participant[] = [];

  for (let i = 0; i < count; i++) {
    const profitPct = (Math.random() * 80 - 20); // -20% to +60%
    const equityGrowth = profitPct * 100; // Assuming $10k starting equity
    const tradesCount = Math.floor(Math.random() * 200) + 10;
    const winRate = 40 + Math.random() * 40; // 40-80% win rate

    // Generate equity curve (20 data points)
    const startEquity = 10000;
    const equityCurve: { timestamp: number; equity: number }[] = [];
    let currentEquity = startEquity;
    const now = Date.now();
    const duration = 30 * 86400000; // 30 days

    for (let j = 0; j < 20; j++) {
      const timestamp = now - duration + (j * duration / 19);
      const change = (Math.random() - 0.45) * (profitPct / 10); // Gradual progress toward final profit
      currentEquity += currentEquity * (change / 100);
      equityCurve.push({ timestamp, equity: currentEquity });
    }

    participants.push({
      id: `part-${competitionId}-${i + 1}`,
      competitionId,
      rank: i + 1,
      traderName: traders[i % traders.length],
      accountId: `ACC-${10000 + i}`,
      profitPct: Math.round(profitPct * 100) / 100,
      equityGrowth: Math.round(equityGrowth * 100) / 100,
      tradesCount,
      winRate: Math.round(winRate * 100) / 100,
      prize: 0, // Will be calculated based on prize distribution
      status: 'active',
      equityCurve
    });
  }

  // Sort by profit percentage descending
  participants.sort((a, b) => b.profitPct - a.profitPct);

  // Update ranks and assign prizes
  participants.forEach((p, index) => {
    p.rank = index + 1;
  });

  return participants;
};

export default function TradingCompetitions() {
  const [activeFilter, setActiveFilter] = useState<TabFilter>('active');
  const [selectedCompetition, setSelectedCompetition] = useState<Competition | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDetailView, setShowDetailView] = useState(false);

  // Mock participants for selected competition
  const [participants, setParticipants] = useState<Participant[]>([]);

  useEffect(() => {
    if (selectedCompetition) {
      const mockParts = generateMockParticipants(selectedCompetition.id, selectedCompetition.participants);

      // Assign prizes based on prize distribution
      const comp = selectedCompetition;
      comp.prizeDistribution.forEach((dist) => {
        if (mockParts[dist.place - 1]) {
          mockParts[dist.place - 1].prize = (comp.prizePool * dist.percentage) / 100;
        }
      });

      setParticipants(mockParts);
    }
  }, [selectedCompetition]);

  // Filter competitions
  const filteredCompetitions = useMemo(() => {
    return MOCK_COMPETITIONS.filter(c => c.status === activeFilter);
  }, [activeFilter]);

  // Calculate time remaining
  const getTimeRemaining = (endDate: number) => {
    const now = Date.now();
    const diff = endDate - now;

    if (diff <= 0) return 'Ended';

    const days = Math.floor(diff / 86400000);
    const hours = Math.floor((diff % 86400000) / 3600000);
    const minutes = Math.floor((diff % 3600000) / 60000);

    if (days > 0) return `${days}d ${hours}h remaining`;
    if (hours > 0) return `${hours}h ${minutes}m remaining`;
    return `${minutes}m remaining`;
  };

  // Calculate countdown for upcoming
  const getCountdown = (startDate: number) => {
    const now = Date.now();
    const diff = startDate - now;

    if (diff <= 0) return 'Starting soon';

    const days = Math.floor(diff / 86400000);
    const hours = Math.floor((diff % 86400000) / 3600000);

    if (days > 0) return `Starts in ${days}d ${hours}h`;
    return `Starts in ${hours}h`;
  };

  // Format date
  const formatDate = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  // Get status badge color
  const getStatusColor = (status: CompetitionStatus) => {
    switch (status) {
      case 'active': return 'bg-emerald-900/30 text-emerald-400 border-emerald-700';
      case 'upcoming': return 'bg-blue-900/30 text-blue-400 border-blue-700';
      case 'past': return 'bg-zinc-700/30 text-zinc-400 border-zinc-600';
    }
  };

  // Equity curve SVG chart
  const EquityCurveChart: React.FC<{ participants: Participant[] }> = ({ participants }) => {
    const width = 800;
    const height = 300;
    const padding = { top: 20, right: 20, bottom: 30, left: 60 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    // Get top 5
    const top5 = participants.slice(0, 5);

    if (top5.length === 0) return null;

    // Find min/max equity across all curves
    let minEquity = Infinity;
    let maxEquity = -Infinity;

    top5.forEach(p => {
      p.equityCurve.forEach(point => {
        minEquity = Math.min(minEquity, point.equity);
        maxEquity = Math.max(maxEquity, point.equity);
      });
    });

    // Add 10% padding to y-axis
    const yPadding = (maxEquity - minEquity) * 0.1;
    minEquity -= yPadding;
    maxEquity += yPadding;

    // Scale functions
    const scaleX = (index: number, total: number) => padding.left + (index / (total - 1)) * chartWidth;
    const scaleY = (equity: number) => padding.top + chartHeight - ((equity - minEquity) / (maxEquity - minEquity)) * chartHeight;

    // Line colors
    const colors = ['#F59E0B', '#3B82F6', '#10B981', '#8B5CF6', '#EC4899'];

    return (
      <svg width={width} height={height} className="overflow-visible">
        {/* Grid lines */}
        {[0, 1, 2, 3, 4].map(i => {
          const y = padding.top + (chartHeight / 4) * i;
          const equity = maxEquity - ((maxEquity - minEquity) / 4) * i;
          return (
            <g key={i}>
              <line
                x1={padding.left}
                y1={y}
                x2={width - padding.right}
                y2={y}
                stroke="#383A42"
                strokeWidth="1"
              />
              <text x={padding.left - 10} y={y + 4} textAnchor="end" fontSize="10" fill="#999">
                ${(equity / 1000).toFixed(1)}k
              </text>
            </g>
          );
        })}

        {/* Equity curves */}
        {top5.map((participant, pIndex) => {
          const pathData = participant.equityCurve.map((point, index) => {
            const x = scaleX(index, participant.equityCurve.length);
            const y = scaleY(point.equity);
            return `${index === 0 ? 'M' : 'L'} ${x} ${y}`;
          }).join(' ');

          return (
            <path
              key={participant.id}
              d={pathData}
              fill="none"
              stroke={colors[pIndex]}
              strokeWidth="2"
              opacity="0.8"
            />
          );
        })}

        {/* Legend */}
        {top5.map((participant, index) => (
          <g key={`legend-${participant.id}`} transform={`translate(${width - padding.right - 150}, ${padding.top + index * 20})`}>
            <circle cx="5" cy="0" r="4" fill={colors[index]} />
            <text x="15" y="4" fontSize="11" fill="#CCC">
              #{participant.rank} {participant.traderName} ({participant.profitPct > 0 ? '+' : ''}{participant.profitPct}%)
            </text>
          </g>
        ))}
      </svg>
    );
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-zinc-300 p-6 overflow-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-zinc-100 flex items-center">
            <Trophy className="mr-3 text-amber-400" size={28} />
            Trading Competitions
          </h1>
          <p className="text-sm text-zinc-500 mt-1">Create and manage trading contests for clients</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 rounded bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium transition-colors flex items-center"
        >
          <Plus size={16} className="mr-2" />
          Create Competition
        </button>
      </div>

      {/* Tab Filters */}
      <div className="flex gap-2 mb-6 border-b border-zinc-700 pb-0">
        <button
          onClick={() => setActiveFilter('active')}
          className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 ${
            activeFilter === 'active'
              ? 'border-emerald-400 text-emerald-400'
              : 'border-transparent text-zinc-500 hover:text-zinc-300'
          }`}
        >
          Active ({MOCK_COMPETITIONS.filter(c => c.status === 'active').length})
        </button>
        <button
          onClick={() => setActiveFilter('upcoming')}
          className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 ${
            activeFilter === 'upcoming'
              ? 'border-blue-400 text-blue-400'
              : 'border-transparent text-zinc-500 hover:text-zinc-300'
          }`}
        >
          Upcoming ({MOCK_COMPETITIONS.filter(c => c.status === 'upcoming').length})
        </button>
        <button
          onClick={() => setActiveFilter('past')}
          className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 ${
            activeFilter === 'past'
              ? 'border-zinc-400 text-zinc-400'
              : 'border-transparent text-zinc-500 hover:text-zinc-300'
          }`}
        >
          Past ({MOCK_COMPETITIONS.filter(c => c.status === 'past').length})
        </button>
      </div>

      {/* Competition Cards Grid */}
      <div className="grid grid-cols-3 gap-6">
        {filteredCompetitions.map(comp => (
          <div
            key={comp.id}
            className="bg-[#1E2026] border border-zinc-700 rounded p-5 hover:border-zinc-600 transition-colors"
          >
            {/* Header */}
            <div className="flex items-start justify-between mb-3">
              <div className="flex-1">
                <h3 className="text-base font-bold text-zinc-100 mb-1">{comp.name}</h3>
                <p className="text-xs text-zinc-500 line-clamp-2">{comp.description}</p>
              </div>
              <span className={`text-[10px] px-2 py-0.5 rounded border ${getStatusColor(comp.status)} uppercase font-bold ml-2`}>
                {comp.status}
              </span>
            </div>

            {/* Countdown / Timer */}
            {comp.status === 'active' && (
              <div className="flex items-center text-xs text-emerald-400 mb-3 bg-emerald-900/20 px-2 py-1 rounded">
                <Clock size={12} className="mr-1" />
                {getTimeRemaining(comp.endDate)}
              </div>
            )}
            {comp.status === 'upcoming' && (
              <div className="flex items-center text-xs text-blue-400 mb-3 bg-blue-900/20 px-2 py-1 rounded">
                <Calendar size={12} className="mr-1" />
                {getCountdown(comp.startDate)}
              </div>
            )}

            {/* Stats */}
            <div className="grid grid-cols-2 gap-3 mb-4">
              <div className="bg-zinc-900/50 rounded p-2">
                <div className="text-[10px] text-zinc-500 uppercase mb-1">Prize Pool</div>
                <div className="text-sm font-bold text-emerald-400">${comp.prizePool.toLocaleString()}</div>
              </div>
              <div className="bg-zinc-900/50 rounded p-2">
                <div className="text-[10px] text-zinc-500 uppercase mb-1">Participants</div>
                <div className="text-sm font-bold text-zinc-100">
                  {comp.participants} / {comp.maxParticipants}
                </div>
              </div>
            </div>

            {/* Dates */}
            <div className="text-xs text-zinc-500 mb-4">
              <div className="flex items-center mb-1">
                <Calendar size={10} className="mr-1" />
                Start: {formatDate(comp.startDate)}
              </div>
              <div className="flex items-center">
                <Calendar size={10} className="mr-1" />
                End: {formatDate(comp.endDate)}
              </div>
            </div>

            {/* Actions */}
            <div className="flex gap-2">
              <button
                onClick={() => {
                  setSelectedCompetition(comp);
                  setShowDetailView(true);
                }}
                className="flex-1 px-3 py-1.5 rounded bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 text-xs font-medium transition-colors flex items-center justify-center"
              >
                <Eye size={12} className="mr-1" />
                View Details
              </button>
              <button
                onClick={() => setSelectedCompetition(comp)}
                className="flex-1 px-3 py-1.5 rounded bg-zinc-700/50 hover:bg-zinc-700 text-zinc-300 text-xs font-medium transition-colors flex items-center justify-center"
              >
                <Settings size={12} className="mr-1" />
                Manage
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Create Competition Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#1E2026] border border-zinc-700 rounded w-full max-w-3xl max-h-[90vh] overflow-auto">
            <div className="px-6 py-4 border-b border-zinc-700 flex items-center justify-between sticky top-0 bg-[#1E2026] z-10">
              <h2 className="text-lg font-bold text-zinc-100">Create New Competition</h2>
              <button
                onClick={() => setShowCreateModal(false)}
                className="text-zinc-500 hover:text-zinc-300"
              >
                ✕
              </button>
            </div>

            <div className="p-6 space-y-6">
              <div className="grid grid-cols-2 gap-4">
                <div className="col-span-2">
                  <label className="block text-sm text-zinc-400 mb-2">Competition Name</label>
                  <input
                    type="text"
                    placeholder="e.g., March Trading Challenge"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  />
                </div>

                <div className="col-span-2">
                  <label className="block text-sm text-zinc-400 mb-2">Description</label>
                  <textarea
                    placeholder="Brief description of the competition..."
                    rows={2}
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100 resize-none"
                  />
                </div>

                <div>
                  <label className="block text-sm text-zinc-400 mb-2">Start Date</label>
                  <input
                    type="date"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  />
                </div>

                <div>
                  <label className="block text-sm text-zinc-400 mb-2">End Date</label>
                  <input
                    type="date"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  />
                </div>

                <div>
                  <label className="block text-sm text-zinc-400 mb-2">Entry Fee ($)</label>
                  <input
                    type="number"
                    placeholder="0 for free"
                    min="0"
                    step="10"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  />
                </div>

                <div>
                  <label className="block text-sm text-zinc-400 mb-2">Prize Pool ($)</label>
                  <input
                    type="number"
                    placeholder="10000"
                    min="0"
                    step="1000"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  />
                </div>

                <div>
                  <label className="block text-sm text-zinc-400 mb-2">Max Participants</label>
                  <input
                    type="number"
                    placeholder="100"
                    min="10"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  />
                </div>

                <div className="col-span-2">
                  <label className="block text-sm text-zinc-400 mb-2">Eligible Account Types</label>
                  <div className="flex gap-4">
                    <label className="flex items-center">
                      <input type="checkbox" className="mr-2" defaultChecked />
                      <span className="text-sm text-zinc-300">Standard</span>
                    </label>
                    <label className="flex items-center">
                      <input type="checkbox" className="mr-2" defaultChecked />
                      <span className="text-sm text-zinc-300">ECN</span>
                    </label>
                    <label className="flex items-center">
                      <input type="checkbox" className="mr-2" />
                      <span className="text-sm text-zinc-300">VIP</span>
                    </label>
                  </div>
                </div>

                <div className="col-span-2">
                  <label className="block text-sm text-zinc-400 mb-2">Competition Rules</label>
                  <textarea
                    placeholder="Enter competition rules and requirements..."
                    rows={4}
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100 resize-none"
                  />
                </div>

                <div className="col-span-2">
                  <label className="block text-sm text-zinc-400 mb-2">Prize Distribution</label>
                  <div className="grid grid-cols-3 gap-3">
                    <div>
                      <label className="text-xs text-zinc-500">1st Place %</label>
                      <input type="number" defaultValue="50" className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-zinc-100 text-sm" />
                    </div>
                    <div>
                      <label className="text-xs text-zinc-500">2nd Place %</label>
                      <input type="number" defaultValue="30" className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-zinc-100 text-sm" />
                    </div>
                    <div>
                      <label className="text-xs text-zinc-500">3rd Place %</label>
                      <input type="number" defaultValue="20" className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-zinc-100 text-sm" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div className="px-6 py-4 border-t border-zinc-700 flex justify-end gap-3 sticky bottom-0 bg-[#1E2026]">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-sm transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 rounded bg-blue-600 hover:bg-blue-700 text-white text-sm transition-colors"
              >
                Create Competition
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Competition Detail View */}
      {showDetailView && selectedCompetition && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#1E2026] border border-zinc-700 rounded w-full max-w-7xl max-h-[95vh] overflow-auto">
            <div className="px-6 py-4 border-b border-zinc-700 flex items-center justify-between sticky top-0 bg-[#1E2026] z-10">
              <div>
                <h2 className="text-xl font-bold text-zinc-100">{selectedCompetition.name}</h2>
                <p className="text-sm text-zinc-500 mt-1">{selectedCompetition.description}</p>
              </div>
              <button
                onClick={() => setShowDetailView(false)}
                className="text-zinc-500 hover:text-zinc-300"
              >
                ✕
              </button>
            </div>

            <div className="p-6 space-y-6">
              {/* Stats Cards */}
              <div className="grid grid-cols-4 gap-4">
                <div className="bg-zinc-900/50 border border-zinc-700 rounded p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs text-zinc-500 uppercase">Total Participants</span>
                    <Users size={16} className="text-blue-400" />
                  </div>
                  <div className="text-2xl font-bold text-zinc-100">{selectedCompetition.participants}</div>
                  <div className="text-xs text-zinc-500 mt-1">of {selectedCompetition.maxParticipants} max</div>
                </div>

                <div className="bg-zinc-900/50 border border-zinc-700 rounded p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs text-zinc-500 uppercase">Total Volume Traded</span>
                    <Activity size={16} className="text-purple-400" />
                  </div>
                  <div className="text-2xl font-bold text-zinc-100">${(selectedCompetition.totalVolume / 1000000).toFixed(1)}M</div>
                  <div className="text-xs text-zinc-500 mt-1">cumulative</div>
                </div>

                <div className="bg-zinc-900/50 border border-zinc-700 rounded p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs text-zinc-500 uppercase">Avg Profit %</span>
                    <TrendingUp size={16} className="text-emerald-400" />
                  </div>
                  <div className="text-2xl font-bold text-emerald-400">+{selectedCompetition.avgProfit}%</div>
                  <div className="text-xs text-zinc-500 mt-1">all participants</div>
                </div>

                <div className="bg-zinc-900/50 border border-zinc-700 rounded p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs text-zinc-500 uppercase">Prize Pool</span>
                    <DollarSign size={16} className="text-amber-400" />
                  </div>
                  <div className="text-2xl font-bold text-amber-400">${selectedCompetition.prizePool.toLocaleString()}</div>
                  <div className="text-xs text-zinc-500 mt-1">total prizes</div>
                </div>
              </div>

              {/* Equity Curve Chart */}
              {participants.length > 0 && (
                <div className="bg-zinc-900/50 border border-zinc-700 rounded p-4">
                  <h3 className="text-sm font-bold text-zinc-100 mb-4">Top 5 Equity Curves</h3>
                  <EquityCurveChart participants={participants} />
                </div>
              )}

              {/* Leaderboard */}
              <div className="bg-zinc-900/50 border border-zinc-700 rounded">
                <div className="px-4 py-3 border-b border-zinc-700">
                  <h3 className="text-sm font-bold text-zinc-100 uppercase">Live Standings</h3>
                </div>

                <div className="overflow-x-auto">
                  <table className="w-full text-xs">
                    <thead className="bg-zinc-800/50 border-b border-zinc-700">
                      <tr className="text-zinc-400">
                        <th className="text-left p-3 w-16">Rank</th>
                        <th className="text-left p-3">Trader</th>
                        <th className="text-right p-3">Profit %</th>
                        <th className="text-right p-3">Equity Growth</th>
                        <th className="text-right p-3">Trades</th>
                        <th className="text-right p-3">Win Rate</th>
                        <th className="text-right p-3">Prize</th>
                        <th className="text-center p-3">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {participants.slice(0, 20).map((p) => (
                        <tr key={p.id} className="border-b border-zinc-800 hover:bg-zinc-900/30">
                          <td className="p-3">
                            <div className="flex items-center">
                              {p.rank === 1 && <Medal size={14} className="mr-1 text-amber-400" />}
                              {p.rank === 2 && <Medal size={14} className="mr-1 text-zinc-400" />}
                              {p.rank === 3 && <Medal size={14} className="mr-1 text-orange-600" />}
                              <span className={`font-bold ${p.rank <= 3 ? 'text-zinc-100' : 'text-zinc-400'}`}>
                                #{p.rank}
                              </span>
                            </div>
                          </td>
                          <td className="p-3">
                            <div className="font-medium text-zinc-100">{p.traderName}</div>
                            <div className="text-[10px] text-zinc-500">{p.accountId}</div>
                          </td>
                          <td className={`p-3 text-right font-mono font-bold ${p.profitPct >= 0 ? 'text-emerald-400' : 'text-rose-400'}`}>
                            {p.profitPct >= 0 ? '+' : ''}{p.profitPct}%
                          </td>
                          <td className={`p-3 text-right font-mono ${p.equityGrowth >= 0 ? 'text-emerald-400' : 'text-rose-400'}`}>
                            ${p.equityGrowth >= 0 ? '+' : ''}{p.equityGrowth.toFixed(2)}
                          </td>
                          <td className="p-3 text-right text-zinc-300">{p.tradesCount}</td>
                          <td className="p-3 text-right text-zinc-300">{p.winRate.toFixed(1)}%</td>
                          <td className="p-3 text-right font-bold text-amber-400">
                            {p.prize > 0 ? `$${p.prize.toFixed(0)}` : '-'}
                          </td>
                          <td className="p-3 text-center">
                            {p.status === 'active' ? (
                              <button
                                className="text-rose-400 hover:text-rose-300 text-xs"
                                title="Disqualify"
                              >
                                <XCircle size={14} />
                              </button>
                            ) : (
                              <span className="text-zinc-600 text-xs">Disqualified</span>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Prize Distribution Table */}
              <div className="bg-zinc-900/50 border border-zinc-700 rounded p-4">
                <h3 className="text-sm font-bold text-zinc-100 mb-3 uppercase">Prize Distribution</h3>
                <div className="grid grid-cols-4 gap-3">
                  {selectedCompetition.prizeDistribution.map((dist) => (
                    <div key={dist.place} className="bg-zinc-800/50 rounded p-3 text-center">
                      <div className="text-xs text-zinc-500 mb-1">{dist.place === 1 ? '1st' : dist.place === 2 ? '2nd' : dist.place === 3 ? '3rd' : `${dist.place}th`} Place</div>
                      <div className="text-lg font-bold text-amber-400">${((selectedCompetition.prizePool * dist.percentage) / 100).toFixed(0)}</div>
                      <div className="text-[10px] text-zinc-500">{dist.percentage}%</div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
