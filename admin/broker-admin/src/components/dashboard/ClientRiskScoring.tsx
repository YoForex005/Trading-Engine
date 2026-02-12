'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { AlertTriangle, TrendingUp, TrendingDown, RefreshCw, Search, Settings, ChevronRight, AlertCircle } from 'lucide-react';
import { API_CONFIG } from '../../config/api';

// ============================================
// Types
// ============================================

interface RiskFactor {
  id: number;
  name: string;
  weight: number; // percentage
  description: string;
}

interface ClientRiskScore {
  clientId: number;
  clientName: string;
  overallScore: number; // 0-100
  category: 'Low' | 'Medium' | 'High' | 'Critical';
  factorScores: Record<string, number>;
  scoreHistory?: ScoreHistoryPoint[];
  lastCalculated: string;
  lastChanged: string;
  previousScore: number;
}

interface ScoreHistoryPoint {
  date: string;
  score: number;
}

interface RiskDistribution {
  bucket: string;
  count: number;
  percentage: number;
}

interface RiskAlert {
  clientId: number;
  clientName: string;
  oldScore: number;
  newScore: number;
  change: number;
  oldCategory: string;
  newCategory: string;
  detectedAt: string;
}

interface RiskStats {
  averageScore: number;
  medianScore: number;
  lowRiskCount: number;
  mediumRiskCount: number;
  highRiskCount: number;
  criticalCount: number;
  totalClients: number;
  scoreTrend: 'improving' | 'stable' | 'deteriorating';
  lastUpdated: string;
}

// ============================================
// Main Component
// ============================================

export default function ClientRiskScoring() {
  const [clients, setClients] = useState<ClientRiskScore[]>([]);
  const [selectedClient, setSelectedClient] = useState<ClientRiskScore | null>(null);
  const [factors, setFactors] = useState<RiskFactor[]>([]);
  const [distribution, setDistribution] = useState<RiskDistribution[]>([]);
  const [alerts, setAlerts] = useState<RiskAlert[]>([]);
  const [stats, setStats] = useState<RiskStats | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isRecalculating, setIsRecalculating] = useState<number | null>(null);

  const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
  const headers = {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
  };

  // Fetch data on mount
  useEffect(() => {
    Promise.all([
      fetchClients(),
      fetchFactors(),
      fetchDistribution(),
      fetchAlerts(),
      fetchStats(),
    ]).finally(() => setIsLoading(false));
  }, []);

  const fetchClients = async () => {
    try {
      const res = await fetch(`${API_CONFIG.RISK_SCORING_CLIENTS}?sortBy=score&sortOrder=desc`, { headers });
      const data = await res.json();
      setClients(data.clients || []);
    } catch (error) {
      console.error('Error fetching clients:', error);
    }
  };

  const fetchFactors = async () => {
    try {
      const res = await fetch(API_CONFIG.RISK_SCORING_FACTORS, { headers });
      const data = await res.json();
      setFactors(data.factors || []);
    } catch (error) {
      console.error('Error fetching factors:', error);
    }
  };

  const fetchDistribution = async () => {
    try {
      const res = await fetch(API_CONFIG.RISK_SCORING_DISTRIBUTION, { headers });
      const data = await res.json();
      setDistribution(data.distribution || []);
    } catch (error) {
      console.error('Error fetching distribution:', error);
    }
  };

  const fetchAlerts = async () => {
    try {
      const res = await fetch(API_CONFIG.RISK_SCORING_ALERTS, { headers });
      const data = await res.json();
      setAlerts(data.alerts || []);
    } catch (error) {
      console.error('Error fetching alerts:', error);
    }
  };

  const fetchStats = async () => {
    try {
      const res = await fetch(API_CONFIG.RISK_SCORING_STATS, { headers });
      const data = await res.json();
      setStats(data);
    } catch (error) {
      console.error('Error fetching stats:', error);
    }
  };

  const handleRecalculate = async (clientId: number) => {
    setIsRecalculating(clientId);
    try {
      const res = await fetch(API_CONFIG.RISK_SCORING_RECALCULATE(clientId), {
        method: 'POST',
        headers,
      });
      const updatedClient = await res.json();

      // Update client in list
      setClients(prev => prev.map(c => c.clientId === clientId ? updatedClient : c));

      // Update selected client if it's the one being recalculated
      if (selectedClient?.clientId === clientId) {
        setSelectedClient(updatedClient);
      }

      // Refresh stats and alerts
      await Promise.all([fetchStats(), fetchAlerts()]);
    } catch (error) {
      console.error('Error recalculating score:', error);
    } finally {
      setIsRecalculating(null);
    }
  };

  const handleSelectClient = async (clientId: number) => {
    try {
      const res = await fetch(API_CONFIG.RISK_SCORING_CLIENT_DETAIL(clientId), { headers });
      const data = await res.json();
      setSelectedClient(data);
    } catch (error) {
      console.error('Error fetching client detail:', error);
    }
  };

  const handleUpdateFactor = async (factorId: number, newWeight: number) => {
    try {
      await fetch(API_CONFIG.RISK_SCORING_FACTOR_UPDATE(factorId), {
        method: 'PUT',
        headers,
        body: JSON.stringify({ weight: newWeight }),
      });
      await fetchFactors();
    } catch (error) {
      console.error('Error updating factor:', error);
      alert('Failed to update factor. Ensure total weight equals 100%.');
    }
  };

  // Filtered clients
  const filteredClients = useMemo(() => {
    let filtered = clients;

    if (categoryFilter) {
      filtered = filtered.filter(c => c.category === categoryFilter);
    }

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(c => c.clientName.toLowerCase().includes(query));
    }

    return filtered;
  }, [clients, categoryFilter, searchQuery]);

  // Helper functions
  const getCategoryColor = (category: string): string => {
    const colors: Record<string, string> = {
      Low: '#10B981',
      Medium: '#F59E0B',
      High: '#F97316',
      Critical: '#EF4444',
    };
    return colors[category] || '#6B7280';
  };

  const formatDate = (dateStr: string): string => {
    return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const formatScore = (score: number): string => {
    return score.toFixed(1);
  };

  // SVG Line Chart for Score History
  const renderScoreHistoryChart = (history: ScoreHistoryPoint[]) => {
    if (!history || history.length === 0) return null;

    const width = 400;
    const height = 150;
    const padding = { top: 10, right: 10, bottom: 20, left: 30 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    const maxScore = 100;
    const minScore = 0;

    const points = history.map((point, i) => {
      const x = padding.left + (i / (history.length - 1)) * chartWidth;
      const y = padding.top + chartHeight - ((point.score - minScore) / (maxScore - minScore)) * chartHeight;
      return { x, y, score: point.score };
    });

    const pathD = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ');

    return (
      <svg width={width} height={height} className="bg-[#121316] rounded">
        {/* Grid lines */}
        {[0, 25, 50, 75, 100].map(score => {
          const y = padding.top + chartHeight - ((score - minScore) / (maxScore - minScore)) * chartHeight;
          return (
            <g key={score}>
              <line x1={padding.left} y1={y} x2={width - padding.right} y2={y} stroke="#383A42" strokeWidth={1} />
              <text x={padding.left - 5} y={y + 3} textAnchor="end" fill="#888" fontSize="10">{score}</text>
            </g>
          );
        })}

        {/* Line path */}
        <path d={pathD} stroke="#F5C542" strokeWidth={2} fill="none" />

        {/* Points */}
        {points.map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r={2} fill="#F5C542" />
        ))}

        {/* X-axis labels */}
        <text x={padding.left} y={height - 5} fill="#888" fontSize="10">30d ago</text>
        <text x={width - padding.right} y={height - 5} textAnchor="end" fill="#888" fontSize="10">Today</text>
      </svg>
    );
  };

  // SVG Horizontal Bar Chart for Factor Scores
  const renderFactorBarChart = (factorScores: Record<string, number>) => {
    const factorNames = Object.keys(factorScores);
    const barHeight = 20;
    const width = 400;
    const height = factorNames.length * (barHeight + 10) + 20;

    return (
      <svg width={width} height={height} className="bg-[#121316] rounded">
        {factorNames.map((name, i) => {
          const score = factorScores[name];
          const barWidth = (score / 100) * (width - 150);
          const y = i * (barHeight + 10) + 10;

          return (
            <g key={name}>
              <text x={5} y={y + barHeight / 2 + 4} fill="#CCC" fontSize="11">{name}</text>
              <rect x={150} y={y} width={barWidth} height={barHeight} fill="#3B82F6" opacity={0.8} />
              <text x={155 + barWidth} y={y + barHeight / 2 + 4} fill="#FFF" fontSize="11" fontWeight="bold">{score.toFixed(1)}</text>
            </g>
          );
        })}
      </svg>
    );
  };

  // SVG Histogram for Distribution
  const renderDistributionHistogram = () => {
    if (distribution.length === 0) return null;

    const width = 500;
    const height = 200;
    const padding = { top: 10, right: 10, bottom: 30, left: 40 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    const maxCount = Math.max(...distribution.map(d => d.count));
    const barWidth = chartWidth / distribution.length;

    return (
      <svg width={width} height={height} className="bg-[#121316] rounded">
        {distribution.map((bucket, i) => {
          const barH = (bucket.count / maxCount) * chartHeight;
          const x = padding.left + i * barWidth;
          const y = padding.top + chartHeight - barH;

          return (
            <g key={bucket.bucket}>
              <rect x={x + 2} y={y} width={barWidth - 4} height={barH} fill="#10B981" opacity={0.8} />
              <text x={x + barWidth / 2} y={padding.top + chartHeight + 15} textAnchor="middle" fill="#888" fontSize="10">{bucket.bucket.split('-')[0]}</text>
              <text x={x + barWidth / 2} y={y - 5} textAnchor="middle" fill="#FFF" fontSize="10" fontWeight="bold">{bucket.count}</text>
            </g>
          );
        })}

        {/* Y-axis label */}
        <text x={5} y={20} fill="#888" fontSize="11">Clients</text>
        <text x={width / 2} y={height - 5} textAnchor="middle" fill="#888" fontSize="11">Risk Score Range</text>
      </svg>
    );
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full bg-[#121316]">
        <div className="text-[#888]">Loading client risk scoring...</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#CCC] font-['Segoe_UI',Tahoma,sans-serif] overflow-hidden">
      {/* Header */}
      <div className="h-10 bg-[#1E2026] border-b border-[#383A42] flex items-center px-3 gap-3 flex-shrink-0">
        <AlertTriangle size={16} className="text-[#F59E0B]" />
        <span className="text-sm font-semibold text-white">Client Risk Scoring / Credit Assessment</span>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="flex gap-3 p-3 bg-[#1A1C21] border-b border-[#383A42] flex-shrink-0">
          <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-2.5">
            <div className="text-xs text-[#888] mb-1">Average Score</div>
            <div className="text-xl font-bold text-white">{formatScore(stats.averageScore)}</div>
          </div>
          <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-2.5">
            <div className="text-xs text-[#888] mb-1">Median Score</div>
            <div className="text-xl font-bold text-white">{formatScore(stats.medianScore)}</div>
          </div>
          <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-2.5">
            <div className="text-xs text-[#888] mb-1">High Risk</div>
            <div className="text-xl font-bold text-[#F97316]">{stats.highRiskCount}</div>
          </div>
          <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-2.5">
            <div className="text-xs text-[#888] mb-1">Critical</div>
            <div className="text-xl font-bold text-[#EF4444]">{stats.criticalCount}</div>
          </div>
          <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-2.5">
            <div className="text-xs text-[#888] mb-1">Trend</div>
            <div className={`text-sm font-bold ${stats.scoreTrend === 'improving' ? 'text-[#10B981]' : stats.scoreTrend === 'deteriorating' ? 'text-[#EF4444]' : 'text-[#888]'}`}>
              {stats.scoreTrend === 'improving' && <TrendingDown size={16} className="inline mr-1" />}
              {stats.scoreTrend === 'deteriorating' && <TrendingUp size={16} className="inline mr-1" />}
              {stats.scoreTrend.charAt(0).toUpperCase() + stats.scoreTrend.slice(1)}
            </div>
          </div>
        </div>
      )}

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left: Client List */}
        <div className="w-96 bg-[#1A1C21] border-r border-[#383A42] flex flex-col overflow-hidden">
          {/* Filters */}
          <div className="p-3 border-b border-[#383A42] space-y-2">
            <div className="relative">
              <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-[#666]" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search by client name..."
                className="w-full h-8 bg-[#1E2026] border border-[#383A42] rounded pl-8 pr-3 text-xs text-white placeholder-[#666] focus:outline-none focus:border-[#F5C542]"
              />
            </div>
            <select
              value={categoryFilter}
              onChange={(e) => setCategoryFilter(e.target.value)}
              className="w-full h-8 bg-[#1E2026] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
            >
              <option value="">All Categories</option>
              <option value="Low">Low Risk</option>
              <option value="Medium">Medium Risk</option>
              <option value="High">High Risk</option>
              <option value="Critical">Critical Risk</option>
            </select>
          </div>

          {/* Client List */}
          <div className="flex-1 overflow-y-auto">
            {filteredClients.map(client => (
              <div
                key={client.clientId}
                onClick={() => handleSelectClient(client.clientId)}
                className={`p-3 border-b border-[#383A42] cursor-pointer transition-colors ${
                  selectedClient?.clientId === client.clientId ? 'bg-[#25272E]' : 'hover:bg-[#1E2026]'
                }`}
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-semibold text-white truncate">{client.clientName}</div>
                    <div className="text-xs text-[#888]">ID: {client.clientId}</div>
                  </div>
                  {selectedClient?.clientId === client.clientId && (
                    <ChevronRight size={16} className="text-[#F5C542] flex-shrink-0 ml-2" />
                  )}
                </div>
                <div className="flex items-center justify-between">
                  <span className="px-2 py-0.5 rounded text-[10px] font-semibold" style={{ backgroundColor: getCategoryColor(client.category) + '33', color: getCategoryColor(client.category) }}>
                    {client.category}
                  </span>
                  <div className="text-lg font-bold text-white">{formatScore(client.overallScore)}</div>
                </div>
                <div className="flex items-center justify-between mt-1.5">
                  <span className="text-xs text-[#666]">Updated: {formatDate(client.lastCalculated)}</span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleRecalculate(client.clientId);
                    }}
                    disabled={isRecalculating === client.clientId}
                    className="px-2 py-0.5 bg-[#3B82F6] text-white text-[10px] rounded hover:bg-[#2563EB] transition-colors disabled:opacity-50"
                  >
                    {isRecalculating === client.clientId ? <RefreshCw size={10} className="animate-spin" /> : 'Recalc'}
                  </button>
                </div>
              </div>
            ))}
            {filteredClients.length === 0 && (
              <div className="flex items-center justify-center h-32 text-[#666]">
                No clients match the filter
              </div>
            )}
          </div>
        </div>

        {/* Right: Details & Charts */}
        <div className="flex-1 flex flex-col overflow-hidden bg-[#121316]">
          {selectedClient ? (
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
              {/* Client Detail Header */}
              <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
                <div className="flex items-start justify-between mb-2">
                  <div>
                    <h3 className="text-lg font-bold text-white">{selectedClient.clientName}</h3>
                    <div className="text-xs text-[#888]">Client ID: {selectedClient.clientId}</div>
                  </div>
                  <div className="text-right">
                    <div className="text-3xl font-bold text-white">{formatScore(selectedClient.overallScore)}</div>
                    <span className="px-2 py-0.5 rounded text-xs font-semibold" style={{ backgroundColor: getCategoryColor(selectedClient.category) + '33', color: getCategoryColor(selectedClient.category) }}>
                      {selectedClient.category} Risk
                    </span>
                  </div>
                </div>
                <div className="text-xs text-[#666] mt-2">
                  Previous: {formatScore(selectedClient.previousScore)} | Last Updated: {formatDate(selectedClient.lastCalculated)}
                </div>
              </div>

              {/* Score History Chart */}
              {selectedClient.scoreHistory && selectedClient.scoreHistory.length > 0 && (
                <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
                  <h4 className="text-sm font-semibold text-white mb-3">30-Day Score History</h4>
                  {renderScoreHistoryChart(selectedClient.scoreHistory)}
                </div>
              )}

              {/* Factor Scores Bar Chart */}
              <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
                <h4 className="text-sm font-semibold text-white mb-3">Risk Factor Breakdown</h4>
                {renderFactorBarChart(selectedClient.factorScores)}
              </div>
            </div>
          ) : (
            <div className="flex-1 flex flex-col overflow-y-auto p-4 space-y-4">
              {/* Risk Factor Configuration */}
              <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
                <h4 className="text-sm font-semibold text-white mb-3 flex items-center gap-2">
                  <Settings size={14} />
                  Risk Factor Configuration
                </h4>
                <div className="space-y-2">
                  {factors.map(factor => (
                    <div key={factor.id} className="flex items-center gap-3">
                      <div className="w-40 text-xs text-[#CCC]">{factor.name}</div>
                      <input
                        type="range"
                        min="0"
                        max="100"
                        step="0.5"
                        value={factor.weight}
                        onChange={(e) => handleUpdateFactor(factor.id, parseFloat(e.target.value))}
                        className="flex-1"
                      />
                      <div className="w-12 text-xs text-white font-semibold text-right">{factor.weight.toFixed(1)}%</div>
                    </div>
                  ))}
                  <div className="mt-3 pt-3 border-t border-[#383A42] flex justify-between text-xs">
                    <span className="text-[#888]">Total Weight:</span>
                    <span className={`font-bold ${factors.reduce((sum, f) => sum + f.weight, 0) === 100 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                      {factors.reduce((sum, f) => sum + f.weight, 0).toFixed(1)}%
                    </span>
                  </div>
                </div>
              </div>

              {/* Score Distribution Histogram */}
              <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
                <h4 className="text-sm font-semibold text-white mb-3">Score Distribution</h4>
                {renderDistributionHistogram()}
              </div>

              {/* Risk Alerts */}
              <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
                <h4 className="text-sm font-semibold text-white mb-3 flex items-center gap-2">
                  <AlertCircle size={14} />
                  Recent Risk Alerts ({'>'}10 point changes in 7 days)
                </h4>
                <div className="space-y-2 max-h-60 overflow-y-auto">
                  {alerts.slice(0, 20).map((alert, idx) => (
                    <div key={idx} className="p-2 bg-[#121316] rounded border border-[#383A42] text-xs">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="font-semibold text-white">{alert.clientName}</div>
                          <div className="text-[#888]">ID: {alert.clientId}</div>
                        </div>
                        <div className="text-right">
                          <div className="flex items-center gap-1">
                            {alert.change > 0 ? (
                              <TrendingUp size={14} className="text-[#EF4444]" />
                            ) : (
                              <TrendingDown size={14} className="text-[#10B981]" />
                            )}
                            <span className={alert.change > 0 ? 'text-[#EF4444]' : 'text-[#10B981]'} style={{ fontWeight: 'bold' }}>
                              {alert.change > 0 ? '+' : ''}{formatScore(alert.change)}
                            </span>
                          </div>
                          <div className="text-[#888]">{formatScore(alert.oldScore)} → {formatScore(alert.newScore)}</div>
                        </div>
                      </div>
                      <div className="flex items-center gap-2 mt-1.5">
                        <span className="px-1.5 py-0.5 rounded text-[9px]" style={{ backgroundColor: getCategoryColor(alert.oldCategory) + '33', color: getCategoryColor(alert.oldCategory) }}>
                          {alert.oldCategory}
                        </span>
                        <span className="text-[#666]">→</span>
                        <span className="px-1.5 py-0.5 rounded text-[9px]" style={{ backgroundColor: getCategoryColor(alert.newCategory) + '33', color: getCategoryColor(alert.newCategory) }}>
                          {alert.newCategory}
                        </span>
                        <span className="ml-auto text-[#666]">{formatDate(alert.detectedAt)}</span>
                      </div>
                    </div>
                  ))}
                  {alerts.length === 0 && (
                    <div className="text-center text-[#666] py-4">No recent risk alerts</div>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
