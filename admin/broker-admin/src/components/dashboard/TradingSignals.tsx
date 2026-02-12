'use client';

import React, { useState, useEffect } from 'react';
import { Search, Star, TrendingUp, TrendingDown, CheckCircle, XCircle, AlertCircle, Award, Users, DollarSign, Target, BarChart3, Clock, Filter } from 'lucide-react';
import { ADMIN_API_URL } from '../../config/api';

type SignalProviderStatus = 'ACTIVE' | 'PENDING' | 'SUSPENDED';
type SignalDirection = 'BUY' | 'SELL';
type SignalResult = 'WIN' | 'LOSS' | 'BREAKEVEN' | 'OPEN';
type SubscriptionStatus = 'ACTIVE' | 'CANCELLED' | 'EXPIRED';

interface SignalProvider {
  id: string;
  name: string;
  winRate: number; // percentage
  subscribers: number;
  rating: number; // 1-5 stars
  monthlyFee: number;
  status: SignalProviderStatus;
  totalPips: number;
  joinDate: string;
  bio: string;
  compositeScore: number;
}

interface Signal {
  id: string;
  providerId: string;
  providerName: string;
  symbol: string;
  direction: SignalDirection;
  entryPrice: number;
  stopLoss: number;
  takeProfit: number;
  openTime: string;
  closeTime?: string;
  result?: SignalResult;
  pips?: number;
  status: 'OPEN' | 'CLOSED';
}

interface Subscription {
  id: string;
  clientId: string;
  clientName: string;
  providerId: string;
  providerName: string;
  status: SubscriptionStatus;
  autoTrade: boolean;
  monthlyFee: number;
  startDate: string;
  endDate?: string;
}

interface SignalStats {
  totalProviders: number;
  activeProviders: number;
  pendingProviders: number;
  suspendedProviders: number;
  activeSignals: number;
  totalSignals: number;
  avgWinRate: number;
  monthlyRevenue: number;
  totalSubscribers: number;
}

export default function TradingSignals() {
  const [providers, setProviders] = useState<SignalProvider[]>([]);
  const [selectedProvider, setSelectedProvider] = useState<SignalProvider | null>(null);
  const [activeSignals, setActiveSignals] = useState<Signal[]>([]);
  const [signalHistory, setSignalHistory] = useState<Signal[]>([]);
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);
  const [stats, setStats] = useState<SignalStats | null>(null);
  const [leaderboard, setLeaderboard] = useState<SignalProvider[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState<SignalProviderStatus | 'ALL'>('ALL');
  const [view, setView] = useState<'providers' | 'signals' | 'subscriptions' | 'leaderboard'>('providers');

  useEffect(() => {
    loadProviders();
    loadActiveSignals();
    loadSignalHistory();
    loadSubscriptions();
    loadStats();
    loadLeaderboard();
  }, []);

  const loadProviders = () => {
    // Mock data - replace with fetch from ADMIN_API_URL
    const mockProviders: SignalProvider[] = Array.from({ length: 15 }, (_, i) => ({
      id: `provider-${i + 1}`,
      name: `${['Forex', 'Pro', 'Elite', 'Master', 'Alpha', 'Prime', 'Signal', 'Trade', 'Profit', 'Expert', 'Quantum', 'Smart', 'Swift', 'Tiger', 'Apex'][i]} Trader`,
      winRate: 55 + Math.random() * 30,
      subscribers: Math.floor(Math.random() * 500) + 50,
      rating: 3 + Math.random() * 2,
      monthlyFee: [29, 49, 79, 99, 149, 199][Math.floor(Math.random() * 6)],
      status: ['ACTIVE', 'ACTIVE', 'ACTIVE', 'ACTIVE', 'PENDING', 'SUSPENDED'][Math.floor(Math.random() * 6)] as SignalProviderStatus,
      totalPips: Math.floor(Math.random() * 5000) + 500,
      joinDate: new Date(Date.now() - Math.random() * 365 * 24 * 60 * 60 * 1000).toISOString(),
      bio: 'Professional trader with 10+ years experience in forex markets.',
      compositeScore: 70 + Math.random() * 30,
    }));
    setProviders(mockProviders);
  };

  const loadActiveSignals = () => {
    // Mock active signals
    const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'NZDUSD', 'XAUUSD', 'BTCUSD'];
    const mockSignals: Signal[] = Array.from({ length: 12 }, (_, i) => {
      const symbol = symbols[Math.floor(Math.random() * symbols.length)];
      const direction: SignalDirection = Math.random() > 0.5 ? 'BUY' : 'SELL';
      const entry = 1.0 + Math.random() * 0.2;
      const sl = direction === 'BUY' ? entry - 0.005 : entry + 0.005;
      const tp = direction === 'BUY' ? entry + 0.01 : entry - 0.01;
      return {
        id: `signal-${i + 1}`,
        providerId: `provider-${Math.floor(Math.random() * 15) + 1}`,
        providerName: `Provider ${Math.floor(Math.random() * 15) + 1}`,
        symbol,
        direction,
        entryPrice: entry,
        stopLoss: sl,
        takeProfit: tp,
        openTime: new Date(Date.now() - Math.random() * 24 * 60 * 60 * 1000).toISOString(),
        status: 'OPEN',
        result: 'OPEN',
      };
    });
    setActiveSignals(mockSignals);
  };

  const loadSignalHistory = () => {
    // Mock closed signals
    const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'NZDUSD', 'XAUUSD', 'BTCUSD'];
    const results: SignalResult[] = ['WIN', 'LOSS', 'BREAKEVEN'];
    const mockHistory: Signal[] = Array.from({ length: 50 }, (_, i) => {
      const symbol = symbols[Math.floor(Math.random() * symbols.length)];
      const direction: SignalDirection = Math.random() > 0.5 ? 'BUY' : 'SELL';
      const result = results[Math.floor(Math.random() * results.length)];
      const pips = result === 'WIN' ? Math.floor(Math.random() * 50) + 10 : result === 'LOSS' ? -(Math.floor(Math.random() * 30) + 5) : 0;
      return {
        id: `history-${i + 1}`,
        providerId: `provider-${Math.floor(Math.random() * 15) + 1}`,
        providerName: `Provider ${Math.floor(Math.random() * 15) + 1}`,
        symbol,
        direction,
        entryPrice: 1.0 + Math.random() * 0.2,
        stopLoss: 1.0,
        takeProfit: 1.0,
        openTime: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
        closeTime: new Date(Date.now() - Math.random() * 20 * 24 * 60 * 60 * 1000).toISOString(),
        status: 'CLOSED',
        result,
        pips,
      };
    });
    setSignalHistory(mockHistory);
  };

  const loadSubscriptions = () => {
    // Mock subscriptions
    const mockSubs: Subscription[] = Array.from({ length: 30 }, (_, i) => ({
      id: `sub-${i + 1}`,
      clientId: `client-${i + 1}`,
      clientName: `Client ${i + 1}`,
      providerId: `provider-${Math.floor(Math.random() * 15) + 1}`,
      providerName: `Provider ${Math.floor(Math.random() * 15) + 1}`,
      status: ['ACTIVE', 'ACTIVE', 'ACTIVE', 'CANCELLED', 'EXPIRED'][Math.floor(Math.random() * 5)] as SubscriptionStatus,
      autoTrade: Math.random() > 0.3,
      monthlyFee: [29, 49, 79, 99, 149, 199][Math.floor(Math.random() * 6)],
      startDate: new Date(Date.now() - Math.random() * 180 * 24 * 60 * 60 * 1000).toISOString(),
      endDate: Math.random() > 0.7 ? new Date(Date.now() + Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString() : undefined,
    }));
    setSubscriptions(mockSubs);
  };

  const loadStats = () => {
    const activeProviders = providers.filter(p => p.status === 'ACTIVE').length;
    const pendingProviders = providers.filter(p => p.status === 'PENDING').length;
    const suspendedProviders = providers.filter(p => p.status === 'SUSPENDED').length;
    const totalSubscribers = subscriptions.filter(s => s.status === 'ACTIVE').length;
    const avgWinRate = providers.reduce((sum, p) => sum + p.winRate, 0) / (providers.length || 1);
    const monthlyRevenue = subscriptions.filter(s => s.status === 'ACTIVE').reduce((sum, s) => sum + s.monthlyFee, 0);

    setStats({
      totalProviders: providers.length,
      activeProviders,
      pendingProviders,
      suspendedProviders,
      activeSignals: activeSignals.length,
      totalSignals: signalHistory.length + activeSignals.length,
      avgWinRate,
      monthlyRevenue,
      totalSubscribers,
    });
  };

  const loadLeaderboard = () => {
    const sorted = [...providers]
      .filter(p => p.status === 'ACTIVE')
      .sort((a, b) => b.compositeScore - a.compositeScore)
      .slice(0, 10);
    setLeaderboard(sorted);
  };

  useEffect(() => {
    loadStats();
    loadLeaderboard();
  }, [providers, activeSignals, signalHistory, subscriptions]);

  const handleApproveProvider = (providerId: string) => {
    setProviders(prev => prev.map(p => p.id === providerId ? { ...p, status: 'ACTIVE' } : p));
  };

  const handleSuspendProvider = (providerId: string) => {
    const reason = prompt('Enter suspension reason:');
    if (reason) {
      setProviders(prev => prev.map(p => p.id === providerId ? { ...p, status: 'SUSPENDED' } : p));
    }
  };

  const toggleAutoTrade = (subId: string) => {
    setSubscriptions(prev => prev.map(s => s.id === subId ? { ...s, autoTrade: !s.autoTrade } : s));
  };

  const filteredProviders = providers.filter(p => {
    const matchesSearch = p.name.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesFilter = filterStatus === 'ALL' || p.status === filterStatus;
    return matchesSearch && matchesFilter;
  });

  const getStatusBadge = (status: SignalProviderStatus) => {
    const colors = {
      ACTIVE: 'bg-[#10B981] text-white',
      PENDING: 'bg-[#F59E0B] text-white',
      SUSPENDED: 'bg-[#EF4444] text-white',
    };
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[status]}`}>
        {status}
      </span>
    );
  };

  const getResultBadge = (result: SignalResult) => {
    const colors = {
      WIN: 'bg-[#10B981] text-white',
      LOSS: 'bg-[#EF4444] text-white',
      BREAKEVEN: 'bg-[#F59E0B] text-white',
      OPEN: 'bg-[#3B82F6] text-white',
    };
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[result]}`}>
        {result}
      </span>
    );
  };

  const getSubscriptionBadge = (status: SubscriptionStatus) => {
    const colors = {
      ACTIVE: 'bg-[#10B981] text-white',
      CANCELLED: 'bg-[#6B7280] text-white',
      EXPIRED: 'bg-[#EF4444] text-white',
    };
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[status]}`}>
        {status}
      </span>
    );
  };

  const getRankBadge = (rank: number) => {
    const colors = {
      1: 'bg-gradient-to-r from-[#FFD700] to-[#FFA500] text-white',
      2: 'bg-gradient-to-r from-[#C0C0C0] to-[#A8A8A8] text-white',
      3: 'bg-gradient-to-r from-[#CD7F32] to-[#A0522D] text-white',
    };
    const color = rank <= 3 ? colors[rank as 1 | 2 | 3] : 'bg-[#505258] text-white';
    return (
      <div className={`w-8 h-8 rounded-full flex items-center justify-center font-bold ${color}`}>
        {rank}
      </div>
    );
  };

  const renderStars = (rating: number) => {
    return (
      <div className="flex gap-0.5">
        {Array.from({ length: 5 }, (_, i) => (
          <Star
            key={i}
            size={14}
            className={i < Math.floor(rating) ? 'fill-[#F59E0B] text-[#F59E0B]' : 'text-[#505258]'}
          />
        ))}
      </div>
    );
  };

  const renderPerformanceChart = (provider: SignalProvider) => {
    // Simple SVG line chart showing provider performance trend
    const points = Array.from({ length: 30 }, (_, i) => ({
      x: i,
      y: 50 + Math.sin(i / 5) * 20 + Math.random() * 10,
    }));
    const width = 300;
    const height = 100;
    const pathData = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${(p.x / 30) * width} ${height - (p.y / 100) * height}`).join(' ');

    return (
      <svg width={width} height={height} className="bg-[#1E2026] rounded">
        <path d={pathData} stroke="#10B981" strokeWidth="2" fill="none" />
      </svg>
    );
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-[#E0E0E0]">
      {/* Header */}
      <div className="h-12 bg-[#1E2026] border-b border-[#383A42] flex items-center justify-between px-4 flex-shrink-0">
        <div className="flex items-center gap-3">
          <Target size={20} className="text-[#10B981]" />
          <span className="font-bold text-sm">Trading Signals / Signal Provider Management</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setView('providers')}
            className={`px-3 py-1 text-xs rounded ${view === 'providers' ? 'bg-[#10B981] text-white' : 'bg-[#383A42] text-[#888] hover:bg-[#505258]'}`}
          >
            Providers
          </button>
          <button
            onClick={() => setView('signals')}
            className={`px-3 py-1 text-xs rounded ${view === 'signals' ? 'bg-[#10B981] text-white' : 'bg-[#383A42] text-[#888] hover:bg-[#505258]'}`}
          >
            Active Signals
          </button>
          <button
            onClick={() => setView('subscriptions')}
            className={`px-3 py-1 text-xs rounded ${view === 'subscriptions' ? 'bg-[#10B981] text-white' : 'bg-[#383A42] text-[#888] hover:bg-[#505258]'}`}
          >
            Subscriptions
          </button>
          <button
            onClick={() => setView('leaderboard')}
            className={`px-3 py-1 text-xs rounded ${view === 'leaderboard' ? 'bg-[#10B981] text-white' : 'bg-[#383A42] text-[#888] hover:bg-[#505258]'}`}
          >
            Leaderboard
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-5 gap-3 p-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
          <div className="bg-[#121316] border border-[#383A42] rounded p-3">
            <div className="flex items-center gap-2 mb-1">
              <Users size={16} className="text-[#3B82F6]" />
              <span className="text-xs text-[#888]">Total Providers</span>
            </div>
            <div className="text-xl font-bold text-[#E0E0E0]">{stats.totalProviders}</div>
            <div className="text-xs text-[#10B981] mt-1">
              {stats.activeProviders} active • {stats.pendingProviders} pending
            </div>
          </div>
          <div className="bg-[#121316] border border-[#383A42] rounded p-3">
            <div className="flex items-center gap-2 mb-1">
              <Target size={16} className="text-[#10B981]" />
              <span className="text-xs text-[#888]">Active Signals</span>
            </div>
            <div className="text-xl font-bold text-[#E0E0E0]">{stats.activeSignals}</div>
            <div className="text-xs text-[#888] mt-1">of {stats.totalSignals} total</div>
          </div>
          <div className="bg-[#121316] border border-[#383A42] rounded p-3">
            <div className="flex items-center gap-2 mb-1">
              <BarChart3 size={16} className="text-[#10B981]" />
              <span className="text-xs text-[#888]">Avg Win Rate</span>
            </div>
            <div className="text-xl font-bold text-[#10B981]">{stats.avgWinRate.toFixed(1)}%</div>
            <div className="text-xs text-[#888] mt-1">across all providers</div>
          </div>
          <div className="bg-[#121316] border border-[#383A42] rounded p-3">
            <div className="flex items-center gap-2 mb-1">
              <DollarSign size={16} className="text-[#F59E0B]" />
              <span className="text-xs text-[#888]">Monthly Revenue</span>
            </div>
            <div className="text-xl font-bold text-[#E0E0E0]">${stats.monthlyRevenue.toLocaleString()}</div>
            <div className="text-xs text-[#888] mt-1">from subscriptions</div>
          </div>
          <div className="bg-[#121316] border border-[#383A42] rounded p-3">
            <div className="flex items-center gap-2 mb-1">
              <Users size={16} className="text-[#3B82F6]" />
              <span className="text-xs text-[#888]">Total Subscribers</span>
            </div>
            <div className="text-xl font-bold text-[#E0E0E0]">{stats.totalSubscribers}</div>
            <div className="text-xs text-[#888] mt-1">active subscriptions</div>
          </div>
        </div>
      )}

      {/* Main Content */}
      <div className="flex-1 overflow-hidden flex">
        {/* Left Panel - Provider List or Leaderboard */}
        {(view === 'providers' || view === 'leaderboard') && (
          <div className="w-2/3 flex flex-col border-r border-[#383A42]">
            {/* Search and Filter */}
            {view === 'providers' && (
              <div className="p-3 bg-[#1E2026] border-b border-[#383A42] flex gap-2 flex-shrink-0">
                <div className="flex-1 relative">
                  <Search size={16} className="absolute left-2 top-1/2 -translate-y-1/2 text-[#888]" />
                  <input
                    type="text"
                    placeholder="Search providers..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full pl-8 pr-3 py-1.5 bg-[#121316] border border-[#383A42] rounded text-xs text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
                  />
                </div>
                <select
                  value={filterStatus}
                  onChange={(e) => setFilterStatus(e.target.value as any)}
                  className="px-3 py-1.5 bg-[#121316] border border-[#383A42] rounded text-xs text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
                >
                  <option value="ALL">All Status</option>
                  <option value="ACTIVE">Active</option>
                  <option value="PENDING">Pending</option>
                  <option value="SUSPENDED">Suspended</option>
                </select>
              </div>
            )}

            {/* Provider Table or Leaderboard */}
            <div className="flex-1 overflow-auto">
              {view === 'providers' && (
                <table className="w-full text-xs">
                  <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                    <tr>
                      <th className="text-left p-2 font-medium text-[#888]">Provider</th>
                      <th className="text-right p-2 font-medium text-[#888]">Win Rate</th>
                      <th className="text-right p-2 font-medium text-[#888]">Subscribers</th>
                      <th className="text-left p-2 font-medium text-[#888]">Rating</th>
                      <th className="text-right p-2 font-medium text-[#888]">Monthly Fee</th>
                      <th className="text-right p-2 font-medium text-[#888]">Total Pips</th>
                      <th className="text-center p-2 font-medium text-[#888]">Status</th>
                      <th className="text-center p-2 font-medium text-[#888]">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredProviders.map((provider) => (
                      <tr
                        key={provider.id}
                        onClick={() => setSelectedProvider(provider)}
                        className={`border-b border-[#383A42] hover:bg-[#1E2026] cursor-pointer ${selectedProvider?.id === provider.id ? 'bg-[#1E2026]' : ''}`}
                      >
                        <td className="p-2">
                          <div className="font-medium text-[#E0E0E0]">{provider.name}</div>
                          <div className="text-[#888] text-[10px]">ID: {provider.id}</div>
                        </td>
                        <td className="p-2 text-right">
                          <span className={provider.winRate >= 65 ? 'text-[#10B981]' : provider.winRate >= 50 ? 'text-[#F59E0B]' : 'text-[#EF4444]'}>
                            {provider.winRate.toFixed(1)}%
                          </span>
                        </td>
                        <td className="p-2 text-right text-[#E0E0E0]">{provider.subscribers}</td>
                        <td className="p-2">{renderStars(provider.rating)}</td>
                        <td className="p-2 text-right text-[#E0E0E0]">${provider.monthlyFee}</td>
                        <td className="p-2 text-right">
                          <span className={provider.totalPips >= 2000 ? 'text-[#10B981]' : 'text-[#E0E0E0]'}>
                            {provider.totalPips.toLocaleString()}
                          </span>
                        </td>
                        <td className="p-2 text-center">{getStatusBadge(provider.status)}</td>
                        <td className="p-2 text-center">
                          <div className="flex gap-1 justify-center">
                            {provider.status === 'PENDING' && (
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleApproveProvider(provider.id);
                                }}
                                className="p-1 rounded hover:bg-[#10B981] hover:text-white"
                                title="Approve"
                              >
                                <CheckCircle size={14} />
                              </button>
                            )}
                            {provider.status === 'ACTIVE' && (
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleSuspendProvider(provider.id);
                                }}
                                className="p-1 rounded hover:bg-[#EF4444] hover:text-white"
                                title="Suspend"
                              >
                                <XCircle size={14} />
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}

              {view === 'leaderboard' && (
                <div className="p-4">
                  <div className="mb-4 flex items-center gap-2">
                    <Award size={20} className="text-[#F59E0B]" />
                    <h3 className="font-bold text-sm">Top 10 Signal Providers</h3>
                  </div>
                  <div className="space-y-2">
                    {leaderboard.map((provider, index) => (
                      <div
                        key={provider.id}
                        className="flex items-center gap-3 p-3 bg-[#1E2026] border border-[#383A42] rounded hover:border-[#10B981] cursor-pointer"
                        onClick={() => setSelectedProvider(provider)}
                      >
                        {getRankBadge(index + 1)}
                        <div className="flex-1">
                          <div className="font-medium text-[#E0E0E0]">{provider.name}</div>
                          <div className="flex items-center gap-3 mt-1">
                            {renderStars(provider.rating)}
                            <span className="text-xs text-[#888]">Win Rate: {provider.winRate.toFixed(1)}%</span>
                            <span className="text-xs text-[#888]">{provider.subscribers} subscribers</span>
                          </div>
                        </div>
                        <div className="text-right">
                          <div className="text-sm font-bold text-[#10B981]">{provider.compositeScore.toFixed(1)}</div>
                          <div className="text-xs text-[#888]">Score</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Active Signals View */}
        {view === 'signals' && (
          <div className="flex-1 flex flex-col">
            <div className="p-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
              <h3 className="font-bold text-sm">Active Trading Signals ({activeSignals.length})</h3>
            </div>
            <div className="flex-1 overflow-auto">
              <table className="w-full text-xs">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                  <tr>
                    <th className="text-left p-2 font-medium text-[#888]">Symbol</th>
                    <th className="text-center p-2 font-medium text-[#888]">Direction</th>
                    <th className="text-right p-2 font-medium text-[#888]">Entry</th>
                    <th className="text-right p-2 font-medium text-[#888]">Stop Loss</th>
                    <th className="text-right p-2 font-medium text-[#888]">Take Profit</th>
                    <th className="text-left p-2 font-medium text-[#888]">Provider</th>
                    <th className="text-left p-2 font-medium text-[#888]">Time Opened</th>
                  </tr>
                </thead>
                <tbody>
                  {activeSignals.map((signal) => (
                    <tr key={signal.id} className="border-b border-[#383A42] hover:bg-[#1E2026]">
                      <td className="p-2 font-medium text-[#E0E0E0]">{signal.symbol}</td>
                      <td className="p-2 text-center">
                        <span className={`px-2 py-0.5 rounded text-xs font-bold ${signal.direction === 'BUY' ? 'bg-[#10B981] text-white' : 'bg-[#EF4444] text-white'}`}>
                          {signal.direction}
                        </span>
                      </td>
                      <td className="p-2 text-right text-[#E0E0E0]">{signal.entryPrice.toFixed(5)}</td>
                      <td className="p-2 text-right text-[#EF4444]">{signal.stopLoss.toFixed(5)}</td>
                      <td className="p-2 text-right text-[#10B981]">{signal.takeProfit.toFixed(5)}</td>
                      <td className="p-2 text-[#888]">{signal.providerName}</td>
                      <td className="p-2 text-[#888]">{new Date(signal.openTime).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Signal History Section */}
            <div className="h-1/2 border-t border-[#383A42] flex flex-col">
              <div className="p-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
                <h3 className="font-bold text-sm">Signal History (Last 50)</h3>
              </div>
              <div className="flex-1 overflow-auto">
                <table className="w-full text-xs">
                  <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                    <tr>
                      <th className="text-left p-2 font-medium text-[#888]">Symbol</th>
                      <th className="text-center p-2 font-medium text-[#888]">Direction</th>
                      <th className="text-center p-2 font-medium text-[#888]">Result</th>
                      <th className="text-right p-2 font-medium text-[#888]">Pips</th>
                      <th className="text-left p-2 font-medium text-[#888]">Provider</th>
                      <th className="text-left p-2 font-medium text-[#888]">Closed</th>
                    </tr>
                  </thead>
                  <tbody>
                    {signalHistory.map((signal) => (
                      <tr key={signal.id} className="border-b border-[#383A42] hover:bg-[#1E2026]">
                        <td className="p-2 font-medium text-[#E0E0E0]">{signal.symbol}</td>
                        <td className="p-2 text-center">
                          <span className={`text-xs ${signal.direction === 'BUY' ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                            {signal.direction}
                          </span>
                        </td>
                        <td className="p-2 text-center">{getResultBadge(signal.result!)}</td>
                        <td className="p-2 text-right">
                          <span className={signal.pips! >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}>
                            {signal.pips! >= 0 ? '+' : ''}{signal.pips}
                          </span>
                        </td>
                        <td className="p-2 text-[#888]">{signal.providerName}</td>
                        <td className="p-2 text-[#888]">{signal.closeTime ? new Date(signal.closeTime).toLocaleString() : '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {/* Subscriptions View */}
        {view === 'subscriptions' && (
          <div className="flex-1 flex flex-col">
            <div className="p-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
              <h3 className="font-bold text-sm">Client-Provider Subscriptions ({subscriptions.length})</h3>
            </div>
            <div className="flex-1 overflow-auto">
              <table className="w-full text-xs">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                  <tr>
                    <th className="text-left p-2 font-medium text-[#888]">Client</th>
                    <th className="text-left p-2 font-medium text-[#888]">Provider</th>
                    <th className="text-center p-2 font-medium text-[#888]">Status</th>
                    <th className="text-center p-2 font-medium text-[#888]">Auto-Trade</th>
                    <th className="text-right p-2 font-medium text-[#888]">Monthly Fee</th>
                    <th className="text-left p-2 font-medium text-[#888]">Start Date</th>
                    <th className="text-left p-2 font-medium text-[#888]">End Date</th>
                  </tr>
                </thead>
                <tbody>
                  {subscriptions.map((sub) => (
                    <tr key={sub.id} className="border-b border-[#383A42] hover:bg-[#1E2026]">
                      <td className="p-2 text-[#E0E0E0]">{sub.clientName}</td>
                      <td className="p-2 text-[#E0E0E0]">{sub.providerName}</td>
                      <td className="p-2 text-center">{getSubscriptionBadge(sub.status)}</td>
                      <td className="p-2 text-center">
                        <button
                          onClick={() => toggleAutoTrade(sub.id)}
                          className={`px-2 py-0.5 rounded text-xs font-medium ${sub.autoTrade ? 'bg-[#10B981] text-white' : 'bg-[#6B7280] text-white'}`}
                          disabled={sub.status !== 'ACTIVE'}
                        >
                          {sub.autoTrade ? 'ON' : 'OFF'}
                        </button>
                      </td>
                      <td className="p-2 text-right text-[#E0E0E0]">${sub.monthlyFee}</td>
                      <td className="p-2 text-[#888]">{new Date(sub.startDate).toLocaleDateString()}</td>
                      <td className="p-2 text-[#888]">{sub.endDate ? new Date(sub.endDate).toLocaleDateString() : 'Active'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Right Panel - Provider Detail */}
        {(view === 'providers' || view === 'leaderboard') && selectedProvider && (
          <div className="w-1/3 flex flex-col bg-[#1E2026]">
            <div className="p-4 border-b border-[#383A42] flex-shrink-0">
              <div className="flex items-start justify-between mb-3">
                <div>
                  <h3 className="font-bold text-sm mb-1">{selectedProvider.name}</h3>
                  <div className="flex items-center gap-2">
                    {renderStars(selectedProvider.rating)}
                    <span className="text-xs text-[#888]">({selectedProvider.rating.toFixed(1)})</span>
                  </div>
                </div>
                {getStatusBadge(selectedProvider.status)}
              </div>
              <div className="grid grid-cols-2 gap-2 text-xs">
                <div>
                  <div className="text-[#888]">Win Rate</div>
                  <div className="font-bold text-[#10B981]">{selectedProvider.winRate.toFixed(1)}%</div>
                </div>
                <div>
                  <div className="text-[#888]">Subscribers</div>
                  <div className="font-bold text-[#E0E0E0]">{selectedProvider.subscribers}</div>
                </div>
                <div>
                  <div className="text-[#888]">Monthly Fee</div>
                  <div className="font-bold text-[#E0E0E0]">${selectedProvider.monthlyFee}</div>
                </div>
                <div>
                  <div className="text-[#888]">Total Pips</div>
                  <div className="font-bold text-[#10B981]">{selectedProvider.totalPips.toLocaleString()}</div>
                </div>
              </div>
            </div>

            <div className="p-4 border-b border-[#383A42] flex-shrink-0">
              <h4 className="font-medium text-xs mb-2 text-[#888]">Performance Chart (30 days)</h4>
              {renderPerformanceChart(selectedProvider)}
            </div>

            <div className="flex-1 overflow-auto p-4">
              <h4 className="font-medium text-xs mb-2 text-[#888]">Recent Signals</h4>
              <div className="space-y-2">
                {activeSignals
                  .filter(s => s.providerId === selectedProvider.id)
                  .slice(0, 5)
                  .map(signal => (
                    <div key={signal.id} className="bg-[#121316] border border-[#383A42] rounded p-2">
                      <div className="flex items-center justify-between mb-1">
                        <span className="font-medium text-[#E0E0E0]">{signal.symbol}</span>
                        <span className={`px-2 py-0.5 rounded text-xs font-bold ${signal.direction === 'BUY' ? 'bg-[#10B981] text-white' : 'bg-[#EF4444] text-white'}`}>
                          {signal.direction}
                        </span>
                      </div>
                      <div className="grid grid-cols-3 gap-1 text-[10px] text-[#888]">
                        <div>Entry: {signal.entryPrice.toFixed(5)}</div>
                        <div>SL: {signal.stopLoss.toFixed(5)}</div>
                        <div>TP: {signal.takeProfit.toFixed(5)}</div>
                      </div>
                      <div className="text-[10px] text-[#888] mt-1">
                        {new Date(signal.openTime).toLocaleString()}
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
