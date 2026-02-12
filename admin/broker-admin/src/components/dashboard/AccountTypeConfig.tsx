import React, { useState } from 'react';
import { Plus, Edit, X, Check, TrendingUp, Users, Shield } from 'lucide-react';

type SpreadType = 'fixed' | 'variable' | 'raw';
type AccountStatus = 'active' | 'inactive';

interface AccountType {
  id: string;
  name: string;
  description: string;
  minDeposit: number;
  maxLeverage: number;
  spreadType: SpreadType;
  commissionPerLot: number;
  marginCallLevel: number;
  stopOutLevel: number;
  maxOpenPositions: number;
  hedgingAllowed: boolean;
  swapFree: boolean;
  oneClickTradingDefault: boolean;
  status: AccountStatus;
  accountCount: number; // Number of accounts using this type
}

interface LeverageOverride {
  id: string;
  symbol: string;
  assetClass: string;
  maxLeverage: number;
}

const INITIAL_ACCOUNT_TYPES: AccountType[] = [
  {
    id: 'demo',
    name: 'Demo',
    description: 'Risk-free practice account with virtual funds',
    minDeposit: 0,
    maxLeverage: 500,
    spreadType: 'variable',
    commissionPerLot: 0,
    marginCallLevel: 100,
    stopOutLevel: 50,
    maxOpenPositions: 100,
    hedgingAllowed: true,
    swapFree: false,
    oneClickTradingDefault: true,
    status: 'active',
    accountCount: 1247,
  },
  {
    id: 'micro',
    name: 'Micro',
    description: 'Small lot sizes for beginners with lower capital requirements',
    minDeposit: 100,
    maxLeverage: 100,
    spreadType: 'variable',
    commissionPerLot: 0,
    marginCallLevel: 100,
    stopOutLevel: 50,
    maxOpenPositions: 50,
    hedgingAllowed: true,
    swapFree: false,
    oneClickTradingDefault: false,
    status: 'active',
    accountCount: 523,
  },
  {
    id: 'standard',
    name: 'Standard',
    description: 'Standard trading account with competitive spreads',
    minDeposit: 500,
    maxLeverage: 200,
    spreadType: 'variable',
    commissionPerLot: 0,
    marginCallLevel: 100,
    stopOutLevel: 50,
    maxOpenPositions: 100,
    hedgingAllowed: true,
    swapFree: false,
    oneClickTradingDefault: false,
    status: 'active',
    accountCount: 892,
  },
  {
    id: 'ecn',
    name: 'ECN',
    description: 'Direct market access with raw spreads and commission',
    minDeposit: 1000,
    maxLeverage: 100,
    spreadType: 'raw',
    commissionPerLot: 6,
    marginCallLevel: 100,
    stopOutLevel: 50,
    maxOpenPositions: 200,
    hedgingAllowed: true,
    swapFree: false,
    oneClickTradingDefault: false,
    status: 'active',
    accountCount: 341,
  },
  {
    id: 'vip',
    name: 'VIP',
    description: 'Premium account with tightest spreads and dedicated support',
    minDeposit: 10000,
    maxLeverage: 50,
    spreadType: 'raw',
    commissionPerLot: 4,
    marginCallLevel: 120,
    stopOutLevel: 60,
    maxOpenPositions: 500,
    hedgingAllowed: true,
    swapFree: false,
    oneClickTradingDefault: true,
    status: 'active',
    accountCount: 87,
  },
  {
    id: 'islamic',
    name: 'Islamic (Swap-Free)',
    description: 'Sharia-compliant account with no overnight interest',
    minDeposit: 500,
    maxLeverage: 100,
    spreadType: 'variable',
    commissionPerLot: 0,
    marginCallLevel: 100,
    stopOutLevel: 50,
    maxOpenPositions: 100,
    hedgingAllowed: true,
    swapFree: true,
    oneClickTradingDefault: false,
    status: 'active',
    accountCount: 156,
  },
];

const INITIAL_LEVERAGE_OVERRIDES: LeverageOverride[] = [
  { id: '1', symbol: 'BTCUSD', assetClass: 'Crypto', maxLeverage: 5 },
  { id: '2', symbol: 'ETHUSD', assetClass: 'Crypto', maxLeverage: 5 },
  { id: '3', symbol: 'XRPUSD', assetClass: 'Crypto', maxLeverage: 5 },
  { id: '4', symbol: 'US30', assetClass: 'Indices', maxLeverage: 50 },
  { id: '5', symbol: 'SPX500', assetClass: 'Indices', maxLeverage: 50 },
  { id: '6', symbol: 'NAS100', assetClass: 'Indices', maxLeverage: 50 },
  { id: '7', symbol: 'XAUUSD', assetClass: 'Metals', maxLeverage: 100 },
  { id: '8', symbol: 'XAGUSD', assetClass: 'Metals', maxLeverage: 100 },
];

const LEVERAGE_OPTIONS = [10, 20, 25, 50, 100, 200, 300, 400, 500, 1000, 2000];

export default function AccountTypeConfig() {
  const [accountTypes, setAccountTypes] = useState<AccountType[]>(INITIAL_ACCOUNT_TYPES);
  const [leverageOverrides, setLeverageOverrides] = useState<LeverageOverride[]>(INITIAL_LEVERAGE_OVERRIDES);
  const [editingType, setEditingType] = useState<AccountType | null>(null);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [showComparison, setShowComparison] = useState(false);

  // Calculate summary stats
  const totalTypes = accountTypes.length;
  const activeTypes = accountTypes.filter((t) => t.status === 'active').length;
  const totalAccounts = accountTypes.reduce((sum, t) => sum + t.accountCount, 0);

  const handleSaveType = (type: AccountType) => {
    if (editingType) {
      // Update existing
      setAccountTypes((prev) => prev.map((t) => (t.id === type.id ? type : t)));
    } else {
      // Create new
      setAccountTypes((prev) => [...prev, { ...type, id: `custom-${Date.now()}`, accountCount: 0 }]);
    }
    setEditingType(null);
    setIsCreateModalOpen(false);
  };

  const handleDeleteType = (id: string) => {
    if (confirm('Are you sure you want to delete this account type?')) {
      setAccountTypes((prev) => prev.filter((t) => t.id !== id));
    }
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#E4E4E7] overflow-y-auto">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
        <h1 className="text-base font-bold text-[#F5C542]">Account Type & Leverage Configuration</h1>
        <button
          onClick={() => {
            setEditingType(null);
            setIsCreateModalOpen(true);
          }}
          className="flex items-center gap-2 px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs font-bold rounded transition-colors"
        >
          <Plus size={14} />
          Create Account Type
        </button>
      </div>

      <div className="flex-1 p-4 space-y-4">
        {/* Summary Cards */}
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center gap-2 mb-2">
              <Shield size={16} className="text-[#3B82F6]" />
              <div className="text-xs text-[#888]">Total Account Types</div>
            </div>
            <div className="text-2xl font-bold text-[#E4E4E7]">{totalTypes}</div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center gap-2 mb-2">
              <Check size={16} className="text-[#10b981]" />
              <div className="text-xs text-[#888]">Active Types</div>
            </div>
            <div className="text-2xl font-bold text-[#10b981]">{activeTypes}</div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center gap-2 mb-3">
              <Users size={16} className="text-[#F5C542]" />
              <div className="text-xs text-[#888]">Accounts per Type</div>
            </div>
            {/* SVG Bar Chart */}
            <svg className="w-full h-16">
              {accountTypes.map((type, i) => {
                const maxCount = Math.max(...accountTypes.map((t) => t.accountCount));
                const barHeight = (type.accountCount / maxCount) * 50;
                const barWidth = 100 / accountTypes.length - 2;
                const x = (i * 100) / accountTypes.length + 1;
                return (
                  <g key={type.id}>
                    <rect
                      x={`${x}%`}
                      y={50 - barHeight}
                      width={`${barWidth}%`}
                      height={barHeight}
                      fill="#3B82F6"
                      opacity="0.8"
                    />
                    <text x={`${x + barWidth / 2}%`} y="62" fontSize="8" fill="#888" textAnchor="middle">
                      {type.name.slice(0, 3)}
                    </text>
                  </g>
                );
              })}
            </svg>
          </div>
        </div>

        {/* Account Type Cards Grid */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-bold text-[#F5C542]">Account Types</h2>
            <button
              onClick={() => setShowComparison(!showComparison)}
              className="text-xs text-[#3B82F6] hover:text-[#2563EB] transition-colors"
            >
              {showComparison ? 'Hide Comparison' : 'Show Comparison Table'}
            </button>
          </div>

          <div className="grid grid-cols-3 gap-3">
            {accountTypes.map((type) => (
              <div
                key={type.id}
                className="bg-[#1E2026] border border-[#383A42] rounded p-3 hover:border-[#3B82F6] transition-colors"
              >
                <div className="flex items-start justify-between mb-2">
                  <div>
                    <h3 className="text-sm font-bold text-[#E4E4E7]">{type.name}</h3>
                    <span
                      className={`text-xs px-2 py-0.5 rounded ${
                        type.status === 'active'
                          ? 'bg-[#10b981] bg-opacity-20 text-[#10b981]'
                          : 'bg-[#888] bg-opacity-20 text-[#888]'
                      }`}
                    >
                      {type.status}
                    </span>
                  </div>
                  <button
                    onClick={() => {
                      setEditingType(type);
                      setIsCreateModalOpen(true);
                    }}
                    className="text-[#888] hover:text-[#3B82F6] transition-colors"
                  >
                    <Edit size={14} />
                  </button>
                </div>

                <p className="text-xs text-[#888] mb-3">{type.description}</p>

                <div className="space-y-1.5 text-xs">
                  <div className="flex justify-between">
                    <span className="text-[#888]">Min Deposit:</span>
                    <span className="text-[#E4E4E7] font-bold">${type.minDeposit.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[#888]">Max Leverage:</span>
                    <span className="text-[#E4E4E7] font-bold">1:{type.maxLeverage}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[#888]">Spread Type:</span>
                    <span className="text-[#E4E4E7]">{type.spreadType}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[#888]">Commission:</span>
                    <span className="text-[#E4E4E7]">${type.commissionPerLot}/lot</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[#888]">Accounts:</span>
                    <span className="text-[#3B82F6] font-bold">{type.accountCount}</span>
                  </div>
                  {type.swapFree && (
                    <div className="pt-1 border-t border-[#383A42]">
                      <span className="text-[#10b981] text-xs">✓ Swap-Free</span>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Comparison Table */}
        {showComparison && (
          <div className="bg-[#1E2026] border border-[#383A42] rounded overflow-hidden">
            <div className="px-4 py-2 bg-[#25272E] border-b border-[#383A42]">
              <h3 className="text-sm font-bold text-[#F5C542]">Account Type Comparison</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead className="bg-[#25272E]">
                  <tr>
                    <th className="text-left px-3 py-2 text-[#888] font-bold border-r border-[#383A42]">Parameter</th>
                    {accountTypes.map((type) => (
                      <th key={type.id} className="text-center px-3 py-2 text-[#E4E4E7] font-bold border-r border-[#383A42] last:border-r-0">
                        {type.name}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#383A42]">
                  <tr className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">Min Deposit</td>
                    {accountTypes.map((type) => (
                      <td key={type.id} className="text-center px-3 py-2 text-[#E4E4E7] border-r border-[#383A42] last:border-r-0">
                        ${type.minDeposit.toLocaleString()}
                      </td>
                    ))}
                  </tr>
                  <tr className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">Max Leverage</td>
                    {accountTypes.map((type) => (
                      <td key={type.id} className="text-center px-3 py-2 text-[#E4E4E7] border-r border-[#383A42] last:border-r-0">
                        1:{type.maxLeverage}
                      </td>
                    ))}
                  </tr>
                  <tr className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">Spread Type</td>
                    {accountTypes.map((type) => (
                      <td key={type.id} className="text-center px-3 py-2 text-[#E4E4E7] border-r border-[#383A42] last:border-r-0">
                        {type.spreadType}
                      </td>
                    ))}
                  </tr>
                  <tr className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">Commission/Lot</td>
                    {accountTypes.map((type) => (
                      <td key={type.id} className="text-center px-3 py-2 text-[#E4E4E7] border-r border-[#383A42] last:border-r-0">
                        ${type.commissionPerLot}
                      </td>
                    ))}
                  </tr>
                  <tr className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">Margin Call %</td>
                    {accountTypes.map((type) => (
                      <td key={type.id} className="text-center px-3 py-2 text-[#E4E4E7] border-r border-[#383A42] last:border-r-0">
                        {type.marginCallLevel}%
                      </td>
                    ))}
                  </tr>
                  <tr className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">Stop Out %</td>
                    {accountTypes.map((type) => (
                      <td key={type.id} className="text-center px-3 py-2 text-[#E4E4E7] border-r border-[#383A42] last:border-r-0">
                        {type.stopOutLevel}%
                      </td>
                    ))}
                  </tr>
                  <tr className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">Max Positions</td>
                    {accountTypes.map((type) => (
                      <td key={type.id} className="text-center px-3 py-2 text-[#E4E4E7] border-r border-[#383A42] last:border-r-0">
                        {type.maxOpenPositions}
                      </td>
                    ))}
                  </tr>
                  <tr className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">Hedging</td>
                    {accountTypes.map((type) => (
                      <td key={type.id} className="text-center px-3 py-2 border-r border-[#383A42] last:border-r-0">
                        <span className={type.hedgingAllowed ? 'text-[#10b981]' : 'text-[#ef4444]'}>
                          {type.hedgingAllowed ? '✓' : '✗'}
                        </span>
                      </td>
                    ))}
                  </tr>
                  <tr className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">Swap-Free</td>
                    {accountTypes.map((type) => (
                      <td key={type.id} className="text-center px-3 py-2 border-r border-[#383A42] last:border-r-0">
                        <span className={type.swapFree ? 'text-[#10b981]' : 'text-[#888]'}>
                          {type.swapFree ? '✓' : '✗'}
                        </span>
                      </td>
                    ))}
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Leverage Schedule Table */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded overflow-hidden">
          <div className="px-4 py-2 bg-[#25272E] border-b border-[#383A42]">
            <h3 className="text-sm font-bold text-[#F5C542]">Leverage Schedule (Per-Symbol Overrides)</h3>
            <p className="text-xs text-[#888] mt-1">
              These leverage limits override account type defaults for specific symbols
            </p>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-[#25272E]">
                <tr>
                  <th className="text-left px-3 py-2 text-[#888] font-bold">Symbol</th>
                  <th className="text-left px-3 py-2 text-[#888] font-bold">Asset Class</th>
                  <th className="text-right px-3 py-2 text-[#888] font-bold">Max Leverage</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#383A42]">
                {leverageOverrides.map((override) => (
                  <tr key={override.id} className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#E4E4E7] font-bold">{override.symbol}</td>
                    <td className="px-3 py-2 text-[#888]">{override.assetClass}</td>
                    <td className="px-3 py-2 text-right">
                      <span className="text-[#F5C542] font-bold">1:{override.maxLeverage}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Create/Edit Modal */}
      {isCreateModalOpen && (
        <AccountTypeModal
          accountType={editingType}
          onSave={handleSaveType}
          onClose={() => {
            setIsCreateModalOpen(false);
            setEditingType(null);
          }}
        />
      )}
    </div>
  );
}

// ============================================================================
// Account Type Create/Edit Modal
// ============================================================================
interface AccountTypeModalProps {
  accountType: AccountType | null;
  onSave: (type: AccountType) => void;
  onClose: () => void;
}

function AccountTypeModal({ accountType, onSave, onClose }: AccountTypeModalProps) {
  const [formData, setFormData] = useState<AccountType>(
    accountType || {
      id: '',
      name: '',
      description: '',
      minDeposit: 500,
      maxLeverage: 100,
      spreadType: 'variable',
      commissionPerLot: 0,
      marginCallLevel: 100,
      stopOutLevel: 50,
      maxOpenPositions: 100,
      hedgingAllowed: true,
      swapFree: false,
      oneClickTradingDefault: false,
      status: 'active',
      accountCount: 0,
    }
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-[#1E2026] border border-[#383A42] rounded-lg w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-[#383A42] sticky top-0 bg-[#1E2026]">
          <h2 className="text-sm font-bold text-[#F5C542]">
            {accountType ? 'Edit Account Type' : 'Create Account Type'}
          </h2>
          <button onClick={onClose} className="text-[#888] hover:text-[#E4E4E7] transition-colors">
            <X size={18} />
          </button>
        </div>

        {/* Modal Body */}
        <form onSubmit={handleSubmit} className="p-4 space-y-4">
          {/* Name & Description */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs text-[#888] mb-1 block">Account Type Name*</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                required
              />
            </div>
            <div>
              <label className="text-xs text-[#888] mb-1 block">Status</label>
              <select
                value={formData.status}
                onChange={(e) => setFormData({ ...formData, status: e.target.value as AccountStatus })}
                className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>
          </div>

          <div>
            <label className="text-xs text-[#888] mb-1 block">Description*</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7] h-16 resize-none"
              required
            />
          </div>

          {/* Financial Parameters */}
          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="text-xs text-[#888] mb-1 block">Minimum Deposit*</label>
              <input
                type="number"
                value={formData.minDeposit}
                onChange={(e) => setFormData({ ...formData, minDeposit: parseFloat(e.target.value) })}
                className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                required
                min="0"
              />
            </div>
            <div>
              <label className="text-xs text-[#888] mb-1 block">Max Leverage*</label>
              <select
                value={formData.maxLeverage}
                onChange={(e) => setFormData({ ...formData, maxLeverage: parseInt(e.target.value) })}
                className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
              >
                {LEVERAGE_OPTIONS.map((lev) => (
                  <option key={lev} value={lev}>
                    1:{lev}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-xs text-[#888] mb-1 block">Commission per Lot</label>
              <input
                type="number"
                value={formData.commissionPerLot}
                onChange={(e) => setFormData({ ...formData, commissionPerLot: parseFloat(e.target.value) })}
                className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                min="0"
                step="0.5"
              />
            </div>
          </div>

          {/* Spread Type */}
          <div>
            <label className="text-xs text-[#888] mb-1 block">Spread Type*</label>
            <div className="flex gap-3">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="spreadType"
                  checked={formData.spreadType === 'fixed'}
                  onChange={() => setFormData({ ...formData, spreadType: 'fixed' })}
                />
                <span className="text-xs text-[#E4E4E7]">Fixed</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="spreadType"
                  checked={formData.spreadType === 'variable'}
                  onChange={() => setFormData({ ...formData, spreadType: 'variable' })}
                />
                <span className="text-xs text-[#E4E4E7]">Variable</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="spreadType"
                  checked={formData.spreadType === 'raw'}
                  onChange={() => setFormData({ ...formData, spreadType: 'raw' })}
                />
                <span className="text-xs text-[#E4E4E7]">Raw</span>
              </label>
            </div>
          </div>

          {/* Risk Parameters */}
          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="text-xs text-[#888] mb-1 block">Margin Call Level (%)*</label>
              <input
                type="number"
                value={formData.marginCallLevel}
                onChange={(e) => setFormData({ ...formData, marginCallLevel: parseFloat(e.target.value) })}
                className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                required
                min="0"
                max="200"
              />
            </div>
            <div>
              <label className="text-xs text-[#888] mb-1 block">Stop Out Level (%)*</label>
              <input
                type="number"
                value={formData.stopOutLevel}
                onChange={(e) => setFormData({ ...formData, stopOutLevel: parseFloat(e.target.value) })}
                className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                required
                min="0"
                max="100"
              />
            </div>
            <div>
              <label className="text-xs text-[#888] mb-1 block">Max Open Positions*</label>
              <input
                type="number"
                value={formData.maxOpenPositions}
                onChange={(e) => setFormData({ ...formData, maxOpenPositions: parseInt(e.target.value) })}
                className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                required
                min="1"
              />
            </div>
          </div>

          {/* Feature Toggles */}
          <div className="space-y-2 bg-[#25272E] border border-[#383A42] rounded p-3">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.hedgingAllowed}
                onChange={(e) => setFormData({ ...formData, hedgingAllowed: e.target.checked })}
              />
              <span className="text-xs text-[#E4E4E7]">Allow Hedging (multiple positions on same symbol)</span>
            </label>

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.swapFree}
                onChange={(e) => setFormData({ ...formData, swapFree: e.target.checked })}
              />
              <span className="text-xs text-[#E4E4E7]">Swap-Free (Islamic Account)</span>
            </label>

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.oneClickTradingDefault}
                onChange={(e) => setFormData({ ...formData, oneClickTradingDefault: e.target.checked })}
              />
              <span className="text-xs text-[#E4E4E7]">Enable One-Click Trading by Default</span>
            </label>
          </div>

          {/* Submit Buttons */}
          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 bg-[#25272E] hover:bg-[#2D2F35] text-[#888] text-xs font-bold rounded transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs font-bold rounded transition-colors"
            >
              {accountType ? 'Save Changes' : 'Create Account Type'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
