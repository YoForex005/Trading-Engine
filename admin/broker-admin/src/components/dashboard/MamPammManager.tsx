'use client';

import React, { useState, useMemo } from 'react';
import { Users, TrendingUp, DollarSign, Activity, Plus, Settings, Eye, Edit2, AlertCircle, BarChart3, Filter } from 'lucide-react';

type AllocationMethod = 'lot' | 'percentage' | 'equity';
type ManagerStatus = 'active' | 'suspended' | 'closed';

interface Manager {
  id: string;
  name: string;
  accountId: string;
  strategyName: string;
  totalAUM: number;
  investorCount: number;
  performanceMTD: number;
  performanceYTD: number;
  managementFee: number;
  performanceFee: number;
  allocationMethod: AllocationMethod;
  status: ManagerStatus;
  equity: number[];
  maxDrawdown: number;
  maxLotSize: number;
  allowedSymbols: string[];
  highWaterMark: boolean;
}

interface Investor {
  id: string;
  managerId: string;
  name: string;
  accountId: string;
  allocationPercent: number;
  currentEquity: number;
  pnl: number;
  pnlPercent: number;
  joinDate: number;
}

// Generate mock data
const generateMockManagers = (): Manager[] => {
  const strategyNames = [
    'Scalping Pro',
    'Trend Following',
    'Mean Reversion',
    'Breakout Strategy',
    'Range Trading',
    'Momentum Strategy',
    'Swing Trading',
    'Day Trading Elite'
  ];

  const now = Date.now();
  const managers: Manager[] = [];

  for (let i = 0; i < 8; i++) {
    const baseEquity = 50000 + Math.random() * 200000;
    const equity: number[] = [];
    let currentEquity = baseEquity;

    // Generate 12 months of equity curve
    for (let m = 0; m < 12; m++) {
      const change = (Math.random() - 0.45) * baseEquity * 0.15;
      currentEquity += change;
      equity.push(Math.max(baseEquity * 0.5, currentEquity));
    }

    const finalEquity = equity[equity.length - 1];
    const performanceYTD = ((finalEquity - baseEquity) / baseEquity) * 100;
    const performanceMTD = ((equity[11] - equity[10]) / equity[10]) * 100;

    managers.push({
      id: `manager-${i + 1}`,
      name: `Manager ${i + 1}`,
      accountId: `MAM${20000 + i}`,
      strategyName: strategyNames[i],
      totalAUM: 50000 + Math.random() * 500000,
      investorCount: 3 + Math.floor(Math.random() * 8),
      performanceMTD,
      performanceYTD,
      managementFee: 1 + Math.random() * 2,
      performanceFee: 15 + Math.random() * 20,
      allocationMethod: ['lot', 'percentage', 'equity'][Math.floor(Math.random() * 3)] as AllocationMethod,
      status: Math.random() > 0.1 ? 'active' : 'suspended',
      equity,
      maxDrawdown: 10 + Math.random() * 20,
      maxLotSize: 10 + Math.floor(Math.random() * 90),
      allowedSymbols: ['EURUSD', 'GBPUSD', 'USDJPY', 'GOLD', 'BTCUSD'].slice(0, 2 + Math.floor(Math.random() * 4)),
      highWaterMark: Math.random() > 0.5
    });
  }

  return managers;
};

const generateMockInvestors = (managers: Manager[]): Investor[] => {
  const investors: Investor[] = [];
  const now = Date.now();

  managers.forEach(manager => {
    const investorCount = manager.investorCount;
    let remainingPercent = 100;

    for (let i = 0; i < investorCount; i++) {
      const allocationPercent = i === investorCount - 1
        ? remainingPercent
        : Math.random() * (remainingPercent / (investorCount - i));

      remainingPercent -= allocationPercent;

      const equity = (manager.totalAUM * allocationPercent) / 100;
      const pnl = equity * ((manager.performanceYTD / 100) * (0.8 + Math.random() * 0.4));
      const pnlPercent = (pnl / (equity - pnl)) * 100;

      investors.push({
        id: `investor-${investors.length + 1}`,
        managerId: manager.id,
        name: `Investor ${investors.length + 1}`,
        accountId: `INV${30000 + investors.length}`,
        allocationPercent,
        currentEquity: equity,
        pnl,
        pnlPercent,
        joinDate: now - Math.floor(Math.random() * 365 * 86400000)
      });
    }
  });

  return investors;
};

const MamPammManager: React.FC = () => {
  const [managers] = useState<Manager[]>(generateMockManagers());
  const [investors] = useState<Investor[]>(() => generateMockInvestors(generateMockManagers()));
  const [selectedManager, setSelectedManager] = useState<Manager | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDetailView, setShowDetailView] = useState(false);

  // Create modal form state
  const [formManagerAccount, setFormManagerAccount] = useState('');
  const [formStrategyName, setFormStrategyName] = useState('');
  const [formAllocationMethod, setFormAllocationMethod] = useState<AllocationMethod>('percentage');
  const [formManagementFee, setFormManagementFee] = useState('2.0');
  const [formPerformanceFee, setFormPerformanceFee] = useState('20.0');
  const [formHighWaterMark, setFormHighWaterMark] = useState(true);
  const [formMinInvestment, setFormMinInvestment] = useState('1000');
  const [formMaxDrawdown, setFormMaxDrawdown] = useState('20');
  const [formMaxLotSize, setFormMaxLotSize] = useState('50');

  // Calculate stats
  const stats = useMemo(() => {
    const activeManagers = managers.filter(m => m.status === 'active').length;
    const totalAUM = managers.reduce((sum, m) => sum + m.totalAUM, 0);
    const totalInvestors = managers.reduce((sum, m) => sum + m.investorCount, 0);
    const avgPerformance = managers.reduce((sum, m) => sum + m.performanceYTD, 0) / managers.length;

    return {
      activeManagers,
      totalAUM,
      totalInvestors,
      avgPerformance
    };
  }, [managers]);

  // Get investors for selected manager
  const managerInvestors = useMemo(() => {
    if (!selectedManager) return [];
    return investors.filter(inv => inv.managerId === selectedManager.id);
  }, [investors, selectedManager]);

  const handleCreateManager = () => {
    console.log('Creating MAM group...');
    setShowCreateModal(false);
  };

  const handleViewManager = (manager: Manager) => {
    setSelectedManager(manager);
    setShowDetailView(true);
  };

  const getStatusColor = (status: ManagerStatus): string => {
    const colors: Record<ManagerStatus, string> = {
      active: '#10B981',
      suspended: '#F59E0B',
      closed: '#6B7280'
    };
    return colors[status];
  };

  const getAllocationMethodLabel = (method: AllocationMethod): string => {
    const labels: Record<AllocationMethod, string> = {
      lot: 'Lot Allocation',
      percentage: 'Percentage Allocation',
      equity: 'Equity Allocation'
    };
    return labels[method];
  };

  // SVG Equity Curve Chart
  const EquityCurveChart: React.FC<{ equity: number[] }> = ({ equity }) => {
    const width = 600;
    const height = 200;
    const padding = { top: 20, right: 20, bottom: 30, left: 50 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    const minEquity = Math.min(...equity);
    const maxEquity = Math.max(...equity);
    const range = maxEquity - minEquity;

    const scaleX = (index: number) => padding.left + (index / (equity.length - 1)) * chartWidth;
    const scaleY = (value: number) => padding.top + chartHeight - ((value - minEquity) / range) * chartHeight;

    // Generate path
    const pathData = equity
      .map((value, index) => {
        const x = scaleX(index);
        const y = scaleY(value);
        return index === 0 ? `M ${x} ${y}` : `L ${x} ${y}`;
      })
      .join(' ');

    // Generate area path
    const areaData = `${pathData} L ${scaleX(equity.length - 1)} ${padding.top + chartHeight} L ${scaleX(0)} ${padding.top + chartHeight} Z`;

    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

    return (
      <svg width={width} height={height} className="mx-auto">
        {/* Grid lines */}
        {[0, 1, 2, 3, 4].map(i => {
          const y = padding.top + (chartHeight / 4) * i;
          return (
            <g key={`grid-${i}`}>
              <line
                x1={padding.left}
                y1={y}
                x2={width - padding.right}
                y2={y}
                stroke="#27272a"
                strokeWidth="1"
              />
              <text
                x={padding.left - 10}
                y={y + 5}
                textAnchor="end"
                className="fill-zinc-500 text-xs"
              >
                ${((maxEquity - (range / 4) * i) / 1000).toFixed(0)}k
              </text>
            </g>
          );
        })}

        {/* Area gradient */}
        <defs>
          <linearGradient id="equityGradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="#3B82F6" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#3B82F6" stopOpacity="0.05" />
          </linearGradient>
        </defs>

        {/* Area */}
        <path d={areaData} fill="url(#equityGradient)" />

        {/* Line */}
        <path d={pathData} fill="none" stroke="#3B82F6" strokeWidth="2" />

        {/* Data points */}
        {equity.map((value, index) => (
          <circle
            key={`point-${index}`}
            cx={scaleX(index)}
            cy={scaleY(value)}
            r="3"
            fill="#3B82F6"
          />
        ))}

        {/* X-axis labels */}
        {months.map((month, index) => (
          <text
            key={`month-${index}`}
            x={scaleX(index)}
            y={height - 10}
            textAnchor="middle"
            className="fill-zinc-500 text-xs"
          >
            {month}
          </text>
        ))}
      </svg>
    );
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-zinc-300 p-4 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between mb-4 pb-3 border-b border-zinc-700">
        <div className="flex items-center gap-3">
          <Users size={20} className="text-[#3B82F6]" />
          <h2 className="text-lg font-semibold text-zinc-100">MAM / PAMM Manager</h2>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs rounded transition-colors flex items-center gap-1.5"
        >
          <Plus size={14} />
          Create MAM Group
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-3 mb-4">
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Active Managers</div>
          <div className="text-2xl font-bold text-[#10B981]">{stats.activeManagers}</div>
        </div>
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Total AUM</div>
          <div className="text-2xl font-bold text-[#3B82F6]">${(stats.totalAUM / 1000).toFixed(0)}k</div>
        </div>
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Total Investors</div>
          <div className="text-2xl font-bold text-[#F59E0B]">{stats.totalInvestors}</div>
        </div>
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Avg Performance (YTD)</div>
          <div className={`text-2xl font-bold ${stats.avgPerformance >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
            {stats.avgPerformance >= 0 ? '+' : ''}{stats.avgPerformance.toFixed(2)}%
          </div>
        </div>
      </div>

      {/* Manager List Table */}
      <div className="flex-1 bg-[#1E2026] border border-zinc-700 rounded overflow-hidden flex flex-col">
        <div className="p-3 border-b border-zinc-700">
          <h3 className="text-sm font-semibold text-zinc-100">Money Managers</h3>
        </div>

        <div className="flex-1 overflow-auto">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-[#1E2026] border-b border-zinc-700">
              <tr className="text-left text-xs text-zinc-500">
                <th className="p-3">Manager</th>
                <th className="p-3">Account ID</th>
                <th className="p-3">Strategy</th>
                <th className="p-3 text-right">AUM</th>
                <th className="p-3 text-center">Investors</th>
                <th className="p-3 text-right">MTD</th>
                <th className="p-3 text-right">YTD</th>
                <th className="p-3">Fee Structure</th>
                <th className="p-3">Allocation</th>
                <th className="p-3">Status</th>
                <th className="p-3">Actions</th>
              </tr>
            </thead>
            <tbody>
              {managers.map(manager => (
                <tr
                  key={manager.id}
                  className="border-b border-zinc-800 hover:bg-[#2A2D35] transition-colors"
                >
                  <td className="p-3 text-zinc-100 font-semibold">{manager.name}</td>
                  <td className="p-3 text-zinc-400">{manager.accountId}</td>
                  <td className="p-3 text-zinc-300">{manager.strategyName}</td>
                  <td className="p-3 text-right text-zinc-100">${manager.totalAUM.toFixed(0)}</td>
                  <td className="p-3 text-center">
                    <span className="px-2 py-1 bg-[#3B82F6]/20 text-[#3B82F6] rounded text-xs">
                      {manager.investorCount}
                    </span>
                  </td>
                  <td className={`p-3 text-right font-semibold ${manager.performanceMTD >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                    {manager.performanceMTD >= 0 ? '+' : ''}{manager.performanceMTD.toFixed(2)}%
                  </td>
                  <td className={`p-3 text-right font-semibold ${manager.performanceYTD >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                    {manager.performanceYTD >= 0 ? '+' : ''}{manager.performanceYTD.toFixed(2)}%
                  </td>
                  <td className="p-3 text-zinc-400 text-xs">
                    {manager.managementFee.toFixed(1)}% + {manager.performanceFee.toFixed(0)}%
                  </td>
                  <td className="p-3">
                    <span className="px-2 py-1 bg-zinc-700 text-zinc-300 rounded text-xs">
                      {getAllocationMethodLabel(manager.allocationMethod).split(' ')[0]}
                    </span>
                  </td>
                  <td className="p-3">
                    <span
                      className="px-2 py-1 rounded text-xs font-semibold capitalize"
                      style={{
                        backgroundColor: getStatusColor(manager.status) + '33',
                        color: getStatusColor(manager.status)
                      }}
                    >
                      {manager.status}
                    </span>
                  </td>
                  <td className="p-3">
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => handleViewManager(manager)}
                        className="p-1.5 hover:bg-[#35383F] rounded transition-colors"
                        title="View Details"
                      >
                        <Eye size={14} className="text-[#3B82F6]" />
                      </button>
                      <button
                        className="p-1.5 hover:bg-[#35383F] rounded transition-colors"
                        title="Edit Manager"
                      >
                        <Edit2 size={14} className="text-zinc-400" />
                      </button>
                      <button
                        className="p-1.5 hover:bg-[#35383F] rounded transition-colors"
                        title="Settings"
                      >
                        <Settings size={14} className="text-zinc-400" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create MAM Group Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowCreateModal(false)}>
          <div className="bg-[#1E2026] border border-zinc-700 rounded-lg w-[600px] max-h-[80vh] overflow-auto" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-zinc-700">
              <h3 className="text-lg font-semibold text-zinc-100">Create MAM Group</h3>
              <button onClick={() => setShowCreateModal(false)} className="text-zinc-500 hover:text-zinc-300">
                ×
              </button>
            </div>

            <div className="p-4 space-y-4">
              {/* Manager Account */}
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Manager Account</label>
                <input
                  type="text"
                  value={formManagerAccount}
                  onChange={(e) => setFormManagerAccount(e.target.value)}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                  placeholder="e.g., MAM20001"
                />
              </div>

              {/* Strategy Name */}
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Strategy Name</label>
                <input
                  type="text"
                  value={formStrategyName}
                  onChange={(e) => setFormStrategyName(e.target.value)}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                  placeholder="e.g., Scalping Pro"
                />
              </div>

              {/* Allocation Method */}
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Allocation Method</label>
                <select
                  value={formAllocationMethod}
                  onChange={(e) => setFormAllocationMethod(e.target.value as AllocationMethod)}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100"
                >
                  <option value="lot">Lot Allocation</option>
                  <option value="percentage">Percentage Allocation</option>
                  <option value="equity">Equity Allocation</option>
                </select>
              </div>

              {/* Fee Structure */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-zinc-400 mb-1">Management Fee (%)</label>
                  <input
                    type="number"
                    step="0.1"
                    value={formManagementFee}
                    onChange={(e) => setFormManagementFee(e.target.value)}
                    className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                  />
                </div>
                <div>
                  <label className="block text-xs text-zinc-400 mb-1">Performance Fee (%)</label>
                  <input
                    type="number"
                    step="1"
                    value={formPerformanceFee}
                    onChange={(e) => setFormPerformanceFee(e.target.value)}
                    className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                  />
                </div>
              </div>

              {/* High Water Mark */}
              <div>
                <label className="flex items-center gap-2 text-sm text-zinc-300">
                  <input
                    type="checkbox"
                    checked={formHighWaterMark}
                    onChange={(e) => setFormHighWaterMark(e.target.checked)}
                    className="rounded border-zinc-600 bg-[#121316]"
                  />
                  Enable High Water Mark
                </label>
                <p className="text-xs text-zinc-500 mt-1 ml-6">
                  Performance fees only charged on new equity highs
                </p>
              </div>

              {/* Min Investment */}
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Minimum Investment</label>
                <input
                  type="number"
                  value={formMinInvestment}
                  onChange={(e) => setFormMinInvestment(e.target.value)}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                />
              </div>

              {/* Risk Controls */}
              <div className="border-t border-zinc-700 pt-4">
                <h4 className="text-sm font-semibold text-zinc-100 mb-3">Risk Controls</h4>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Max Drawdown (%)</label>
                    <input
                      type="number"
                      value={formMaxDrawdown}
                      onChange={(e) => setFormMaxDrawdown(e.target.value)}
                      className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-zinc-400 mb-1">Max Lot Size</label>
                    <input
                      type="number"
                      value={formMaxLotSize}
                      onChange={(e) => setFormMaxLotSize(e.target.value)}
                      className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                    />
                  </div>
                </div>
              </div>
            </div>

            <div className="flex justify-end gap-2 p-4 border-t border-zinc-700">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 bg-[#2A2D35] hover:bg-[#35383F] text-zinc-300 text-sm rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateManager}
                className="px-4 py-2 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-sm rounded transition-colors"
              >
                Create MAM Group
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Manager Detail View */}
      {showDetailView && selectedManager && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowDetailView(false)}>
          <div className="bg-[#1E2026] border border-zinc-700 rounded-lg w-[1000px] max-h-[90vh] flex flex-col" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-zinc-700">
              <div>
                <h3 className="text-lg font-semibold text-zinc-100">{selectedManager.name}</h3>
                <div className="text-sm text-zinc-400">{selectedManager.accountId} - {selectedManager.strategyName}</div>
              </div>
              <button onClick={() => setShowDetailView(false)} className="text-zinc-500 hover:text-zinc-300">
                ×
              </button>
            </div>

            <div className="flex-1 overflow-auto p-4 space-y-4">
              {/* Manager Profile Summary */}
              <div className="grid grid-cols-3 gap-4">
                <div className="bg-[#121316] border border-zinc-700 rounded p-3">
                  <div className="text-xs text-zinc-500 mb-1">Total AUM</div>
                  <div className="text-xl font-bold text-zinc-100">${selectedManager.totalAUM.toFixed(0)}</div>
                </div>
                <div className="bg-[#121316] border border-zinc-700 rounded p-3">
                  <div className="text-xs text-zinc-500 mb-1">YTD Performance</div>
                  <div className={`text-xl font-bold ${selectedManager.performanceYTD >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                    {selectedManager.performanceYTD >= 0 ? '+' : ''}{selectedManager.performanceYTD.toFixed(2)}%
                  </div>
                </div>
                <div className="bg-[#121316] border border-zinc-700 rounded p-3">
                  <div className="text-xs text-zinc-500 mb-1">Investors</div>
                  <div className="text-xl font-bold text-zinc-100">{selectedManager.investorCount}</div>
                </div>
              </div>

              {/* Performance Chart */}
              <div className="bg-[#121316] border border-zinc-700 rounded p-4">
                <h4 className="text-sm font-semibold text-zinc-100 mb-3">12-Month Equity Curve</h4>
                <EquityCurveChart equity={selectedManager.equity} />
              </div>

              {/* Configuration Details */}
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-[#121316] border border-zinc-700 rounded p-4">
                  <h4 className="text-sm font-semibold text-zinc-100 mb-3">Fee Structure</h4>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-zinc-500">Management Fee:</span>
                      <span className="text-zinc-100">{selectedManager.managementFee.toFixed(1)}%</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-zinc-500">Performance Fee:</span>
                      <span className="text-zinc-100">{selectedManager.performanceFee.toFixed(0)}%</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-zinc-500">High Water Mark:</span>
                      <span className="text-zinc-100">{selectedManager.highWaterMark ? 'Enabled' : 'Disabled'}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-zinc-500">Allocation Method:</span>
                      <span className="text-zinc-100">{getAllocationMethodLabel(selectedManager.allocationMethod)}</span>
                    </div>
                  </div>
                </div>

                <div className="bg-[#121316] border border-zinc-700 rounded p-4">
                  <h4 className="text-sm font-semibold text-zinc-100 mb-3">Risk Controls</h4>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-zinc-500">Max Drawdown:</span>
                      <span className="text-zinc-100">{selectedManager.maxDrawdown.toFixed(1)}%</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-zinc-500">Max Lot Size:</span>
                      <span className="text-zinc-100">{selectedManager.maxLotSize}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-zinc-500">Allowed Symbols:</span>
                      <span className="text-zinc-100">{selectedManager.allowedSymbols.length} symbols</span>
                    </div>
                    <div className="flex flex-wrap gap-1 mt-2">
                      {selectedManager.allowedSymbols.map(symbol => (
                        <span key={symbol} className="px-2 py-0.5 bg-zinc-700 text-zinc-300 rounded text-xs">
                          {symbol}
                        </span>
                      ))}
                    </div>
                  </div>
                </div>
              </div>

              {/* Investor Allocation Table */}
              <div className="bg-[#121316] border border-zinc-700 rounded p-4">
                <h4 className="text-sm font-semibold text-zinc-100 mb-3">Investor Allocation</h4>
                <table className="w-full text-sm">
                  <thead className="border-b border-zinc-700">
                    <tr className="text-left text-xs text-zinc-500">
                      <th className="pb-2">Investor</th>
                      <th className="pb-2">Account ID</th>
                      <th className="pb-2 text-right">Allocation %</th>
                      <th className="pb-2 text-right">Current Equity</th>
                      <th className="pb-2 text-right">P&L</th>
                      <th className="pb-2 text-right">P&L %</th>
                      <th className="pb-2">Join Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    {managerInvestors.map(investor => (
                      <tr key={investor.id} className="border-b border-zinc-800">
                        <td className="py-2 text-zinc-100">{investor.name}</td>
                        <td className="py-2 text-zinc-400">{investor.accountId}</td>
                        <td className="py-2 text-right text-zinc-100">{investor.allocationPercent.toFixed(1)}%</td>
                        <td className="py-2 text-right text-zinc-100">${investor.currentEquity.toFixed(2)}</td>
                        <td className={`py-2 text-right font-semibold ${investor.pnl >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                          {investor.pnl >= 0 ? '+' : ''}${investor.pnl.toFixed(2)}
                        </td>
                        <td className={`py-2 text-right font-semibold ${investor.pnlPercent >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                          {investor.pnlPercent >= 0 ? '+' : ''}{investor.pnlPercent.toFixed(2)}%
                        </td>
                        <td className="py-2 text-zinc-400">{new Date(investor.joinDate).toLocaleDateString()}</td>
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
};

export default MamPammManager;
