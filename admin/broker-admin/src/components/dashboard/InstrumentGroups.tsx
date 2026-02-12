'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { Plus, Edit2, Trash2, Search, Users, TrendingUp, Settings, ChevronDown, ChevronRight } from 'lucide-react';
import { API_CONFIG } from '../../config/api';

// ============================================
// Types
// ============================================

type SpreadType = 'fixed' | 'variable';
type CommissionModel = 'none' | 'per_lot' | 'percentage';
type MarginMode = 'retail' | 'professional' | 'hedge';

interface InstrumentGroup {
  id: string;
  name: string;
  description: string;
  defaultLeverage: number;
  spreadType: SpreadType;
  commissionModel: CommissionModel;
  commissionRate: number;
  marginMode: MarginMode;
  tradingHoursTemplate: string;
  symbolCount: number;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

interface SymbolInGroup {
  symbol: string;
  groupId: string;
  groupName: string;
  leverage: number;
  spreadType: SpreadType;
  commissionModel: CommissionModel;
  commissionRate: number;
  marginMode: MarginMode;
  tradingHours: string;
  isActive: boolean;
  addedAt: string;
}

interface InstrumentStats {
  totalGroups: number;
  totalSymbols: number;
  mostActiveGroupId: string;
  mostActiveGroupName: string;
  mostActiveSymbols: number;
}

// ============================================
// Main Component
// ============================================

export default function InstrumentGroups() {
  const [groups, setGroups] = useState<InstrumentGroup[]>([]);
  const [selectedGroup, setSelectedGroup] = useState<InstrumentGroup | null>(null);
  const [groupSymbols, setGroupSymbols] = useState<SymbolInGroup[]>([]);
  const [stats, setStats] = useState<InstrumentStats | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isAddSymbolModalOpen, setIsAddSymbolModalOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [sortField, setSortField] = useState<keyof SymbolInGroup>('symbol');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');

  // Create/Edit form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    defaultLeverage: 100,
    spreadType: 'variable' as SpreadType,
    commissionModel: 'none' as CommissionModel,
    commissionRate: 0,
    marginMode: 'retail' as MarginMode,
    tradingHoursTemplate: '24/5',
  });

  const [newSymbol, setNewSymbol] = useState('');

  const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
  const headers = {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
  };

  // Fetch data on mount
  useEffect(() => {
    fetchGroups();
    fetchStats();
  }, []);

  // Fetch group symbols when selected group changes
  useEffect(() => {
    if (selectedGroup) {
      fetchGroupSymbols(selectedGroup.id);
    }
  }, [selectedGroup]);

  const fetchGroups = async () => {
    try {
      const res = await fetch(API_CONFIG.INSTRUMENT_GROUPS, { headers });
      const data = await res.json();
      setGroups(data.groups || []);
    } catch (error) {
      console.error('Error fetching groups:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const fetchGroupSymbols = async (groupId: string) => {
    try {
      const res = await fetch(API_CONFIG.INSTRUMENT_GROUP_SYMBOLS(groupId), { headers });
      const data = await res.json();
      setGroupSymbols(data.symbols || []);
    } catch (error) {
      console.error('Error fetching group symbols:', error);
    }
  };

  const fetchStats = async () => {
    try {
      const res = await fetch(API_CONFIG.INSTRUMENT_STATS, { headers });
      const data = await res.json();
      setStats(data);
    } catch (error) {
      console.error('Error fetching stats:', error);
    }
  };

  const handleCreateGroup = async () => {
    try {
      const res = await fetch(API_CONFIG.INSTRUMENT_GROUPS, {
        method: 'POST',
        headers,
        body: JSON.stringify(formData),
      });
      if (res.ok) {
        await fetchGroups();
        await fetchStats();
        setIsCreateModalOpen(false);
        resetForm();
      }
    } catch (error) {
      console.error('Error creating group:', error);
    }
  };

  const handleUpdateGroup = async () => {
    if (!selectedGroup) return;
    try {
      const res = await fetch(API_CONFIG.INSTRUMENT_GROUP_UPDATE(selectedGroup.id), {
        method: 'PUT',
        headers,
        body: JSON.stringify(formData),
      });
      if (res.ok) {
        await fetchGroups();
        setIsEditModalOpen(false);
        // Update selected group
        const updatedGroup = groups.find(g => g.id === selectedGroup.id);
        if (updatedGroup) setSelectedGroup(updatedGroup);
        resetForm();
      }
    } catch (error) {
      console.error('Error updating group:', error);
    }
  };

  const handleDeleteGroup = async (groupId: string) => {
    if (!confirm('Are you sure you want to delete this group? This action cannot be undone.')) return;
    try {
      const res = await fetch(API_CONFIG.INSTRUMENT_GROUP_DELETE(groupId), {
        method: 'DELETE',
        headers,
      });
      if (res.ok) {
        await fetchGroups();
        await fetchStats();
        if (selectedGroup?.id === groupId) {
          setSelectedGroup(null);
          setGroupSymbols([]);
        }
      } else {
        const error = await res.json();
        alert(error.message || 'Failed to delete group');
      }
    } catch (error) {
      console.error('Error deleting group:', error);
    }
  };

  const handleAddSymbol = async () => {
    if (!selectedGroup || !newSymbol.trim()) return;
    try {
      const res = await fetch(API_CONFIG.INSTRUMENT_GROUP_ADD_SYMBOL(selectedGroup.id), {
        method: 'POST',
        headers,
        body: JSON.stringify({ symbol: newSymbol.toUpperCase() }),
      });
      if (res.ok) {
        await fetchGroupSymbols(selectedGroup.id);
        await fetchGroups();
        setIsAddSymbolModalOpen(false);
        setNewSymbol('');
      } else {
        const error = await res.json();
        alert(error.message || 'Failed to add symbol');
      }
    } catch (error) {
      console.error('Error adding symbol:', error);
    }
  };

  const handleRemoveSymbol = async (symbol: string) => {
    if (!selectedGroup) return;
    if (!confirm(`Remove ${symbol} from ${selectedGroup.name}?`)) return;
    try {
      const res = await fetch(API_CONFIG.INSTRUMENT_GROUP_REMOVE_SYMBOL(selectedGroup.id, symbol), {
        method: 'DELETE',
        headers,
      });
      if (res.ok) {
        await fetchGroupSymbols(selectedGroup.id);
        await fetchGroups();
      }
    } catch (error) {
      console.error('Error removing symbol:', error);
    }
  };

  const handleToggleSymbolActive = async (symbol: SymbolInGroup) => {
    // In a real implementation, this would call an API endpoint to toggle active status
    console.log('Toggle active for symbol:', symbol.symbol);
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      defaultLeverage: 100,
      spreadType: 'variable',
      commissionModel: 'none',
      commissionRate: 0,
      marginMode: 'retail',
      tradingHoursTemplate: '24/5',
    });
  };

  const openEditModal = (group: InstrumentGroup) => {
    setFormData({
      name: group.name,
      description: group.description,
      defaultLeverage: group.defaultLeverage,
      spreadType: group.spreadType,
      commissionModel: group.commissionModel,
      commissionRate: group.commissionRate,
      marginMode: group.marginMode,
      tradingHoursTemplate: group.tradingHoursTemplate,
    });
    setIsEditModalOpen(true);
  };

  const handleSort = (field: keyof SymbolInGroup) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  // Filtered and sorted symbols
  const filteredSymbols = useMemo(() => {
    let filtered = groupSymbols;

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(s => s.symbol.toLowerCase().includes(query));
    }

    // Sort
    filtered = [...filtered].sort((a, b) => {
      const aVal = a[sortField];
      const bVal = b[sortField];
      if (aVal === bVal) return 0;
      const result = aVal > bVal ? 1 : -1;
      return sortDirection === 'asc' ? result : -result;
    });

    return filtered;
  }, [groupSymbols, searchQuery, sortField, sortDirection]);

  // Helper functions
  const getSpreadTypeBadgeColor = (type: SpreadType): string => {
    return type === 'fixed' ? '#3B82F6' : '#10B981';
  };

  const getCommissionBadgeColor = (model: CommissionModel): string => {
    const colors: Record<CommissionModel, string> = {
      none: '#6B7280',
      per_lot: '#F59E0B',
      percentage: '#8B5CF6',
    };
    return colors[model] || '#6B7280';
  };

  const getMarginModeBadgeColor = (mode: MarginMode): string => {
    const colors: Record<MarginMode, string> = {
      retail: '#3B82F6',
      professional: '#10B981',
      hedge: '#F59E0B',
    };
    return colors[mode] || '#6B7280';
  };

  const formatCommissionModel = (model: CommissionModel): string => {
    const labels: Record<CommissionModel, string> = {
      none: 'None',
      per_lot: 'Per Lot',
      percentage: 'Percentage',
    };
    return labels[model] || model;
  };

  const formatMarginMode = (mode: MarginMode): string => {
    const labels: Record<MarginMode, string> = {
      retail: 'Retail',
      professional: 'Professional',
      hedge: 'Hedge',
    };
    return labels[mode] || mode;
  };

  const formatSpreadType = (type: SpreadType): string => {
    return type === 'fixed' ? 'Fixed' : 'Variable';
  };

  const getGroupIcon = (name: string): string => {
    if (name.includes('Forex')) return '💱';
    if (name.includes('Metals')) return '🥇';
    if (name.includes('Energies')) return '⚡';
    if (name.includes('Indices')) return '📊';
    if (name.includes('Crypto')) return '₿';
    if (name.includes('Commodities')) return '🌾';
    return '📦';
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full bg-[#121316]">
        <div className="text-[#888]">Loading instrument groups...</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#CCC] font-['Segoe_UI',Tahoma,sans-serif] overflow-hidden">
      {/* Header */}
      <div className="h-10 bg-[#1E2026] border-b border-[#383A42] flex items-center px-3 gap-3 flex-shrink-0">
        <Settings size={16} className="text-[#F5C542]" />
        <span className="text-sm font-semibold text-white">Instrument Groups / Symbol Category Management</span>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          className="ml-auto px-3 py-1 bg-[#F5C542] text-[#121316] text-xs font-bold rounded hover:bg-[#E5B532] transition-colors flex items-center gap-1.5"
        >
          <Plus size={14} />
          Create Group
        </button>
      </div>

      {/* Stats Overview Cards */}
      {stats && (
        <div className="flex gap-3 p-3 bg-[#1A1C21] border-b border-[#383A42] flex-shrink-0">
          <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-2.5">
            <div className="text-xs text-[#888] mb-1">Total Groups</div>
            <div className="text-xl font-bold text-white">{stats.totalGroups}</div>
          </div>
          <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-2.5">
            <div className="text-xs text-[#888] mb-1">Total Symbols</div>
            <div className="text-xl font-bold text-white">{stats.totalSymbols}</div>
          </div>
          <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded p-2.5">
            <div className="text-xs text-[#888] mb-1">Most Active Group</div>
            <div className="text-sm font-bold text-[#F5C542]">{stats.mostActiveGroupName}</div>
            <div className="text-xs text-[#888] mt-0.5">{stats.mostActiveSymbols} symbols</div>
          </div>
        </div>
      )}

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Panel: Group Cards */}
        <div className="w-80 bg-[#1A1C21] border-r border-[#383A42] flex flex-col overflow-hidden">
          <div className="p-3 border-b border-[#383A42]">
            <div className="text-xs font-semibold text-[#888] mb-2">INSTRUMENT GROUPS ({groups.length})</div>
          </div>
          <div className="flex-1 overflow-y-auto">
            {groups.map(group => (
              <div
                key={group.id}
                onClick={() => setSelectedGroup(group)}
                className={`p-3 border-b border-[#383A42] cursor-pointer transition-colors ${
                  selectedGroup?.id === group.id ? 'bg-[#25272E]' : 'hover:bg-[#1E2026]'
                }`}
              >
                <div className="flex items-start gap-2 mb-2">
                  <span className="text-xl">{getGroupIcon(group.name)}</span>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-semibold text-white truncate">{group.name}</div>
                    <div className="text-xs text-[#888] truncate">{group.description}</div>
                  </div>
                  {selectedGroup?.id === group.id && (
                    <ChevronRight size={16} className="text-[#F5C542] flex-shrink-0" />
                  )}
                </div>
                <div className="flex gap-2 flex-wrap text-[10px] mb-2">
                  <span className="px-1.5 py-0.5 rounded" style={{ backgroundColor: getSpreadTypeBadgeColor(group.spreadType) + '33', color: getSpreadTypeBadgeColor(group.spreadType) }}>
                    {formatSpreadType(group.spreadType)}
                  </span>
                  <span className="px-1.5 py-0.5 rounded" style={{ backgroundColor: getCommissionBadgeColor(group.commissionModel) + '33', color: getCommissionBadgeColor(group.commissionModel) }}>
                    {formatCommissionModel(group.commissionModel)}
                  </span>
                  <span className="px-1.5 py-0.5 rounded" style={{ backgroundColor: getMarginModeBadgeColor(group.marginMode) + '33', color: getMarginModeBadgeColor(group.marginMode) }}>
                    {formatMarginMode(group.marginMode)}
                  </span>
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-[#888]">
                    <Users size={12} className="inline mr-1" />
                    {group.symbolCount} symbols
                  </span>
                  <span className="text-[#888]">1:{group.defaultLeverage}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Right Panel: Group Detail & Symbols */}
        {selectedGroup ? (
          <div className="flex-1 flex flex-col overflow-hidden bg-[#121316]">
            {/* Group Detail Header */}
            <div className="bg-[#1E2026] border-b border-[#383A42] p-3 flex-shrink-0">
              <div className="flex items-start justify-between mb-3">
                <div>
                  <div className="text-lg font-bold text-white mb-1">{selectedGroup.name}</div>
                  <div className="text-xs text-[#888]">{selectedGroup.description}</div>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => openEditModal(selectedGroup)}
                    className="px-3 py-1.5 bg-[#3B82F6] text-white text-xs font-bold rounded hover:bg-[#2563EB] transition-colors flex items-center gap-1"
                  >
                    <Edit2 size={12} />
                    Edit Group
                  </button>
                  <button
                    onClick={() => handleDeleteGroup(selectedGroup.id)}
                    className="px-3 py-1.5 bg-[#EF4444] text-white text-xs font-bold rounded hover:bg-[#DC2626] transition-colors flex items-center gap-1"
                  >
                    <Trash2 size={12} />
                    Delete
                  </button>
                </div>
              </div>
              <div className="grid grid-cols-4 gap-3 text-xs">
                <div>
                  <div className="text-[#888] mb-1">Default Leverage</div>
                  <div className="text-white font-semibold">1:{selectedGroup.defaultLeverage}</div>
                </div>
                <div>
                  <div className="text-[#888] mb-1">Spread Type</div>
                  <div className="text-white font-semibold">{formatSpreadType(selectedGroup.spreadType)}</div>
                </div>
                <div>
                  <div className="text-[#888] mb-1">Commission</div>
                  <div className="text-white font-semibold">
                    {selectedGroup.commissionModel === 'none' ? 'None' :
                     selectedGroup.commissionModel === 'per_lot' ? `$${selectedGroup.commissionRate}/lot` :
                     `${selectedGroup.commissionRate}%`}
                  </div>
                </div>
                <div>
                  <div className="text-[#888] mb-1">Trading Hours</div>
                  <div className="text-white font-semibold">{selectedGroup.tradingHoursTemplate}</div>
                </div>
              </div>
            </div>

            {/* Search & Add Symbol */}
            <div className="bg-[#1A1C21] border-b border-[#383A42] p-3 flex gap-3 flex-shrink-0">
              <div className="flex-1 relative">
                <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-[#666]" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search symbols in this group..."
                  className="w-full h-8 bg-[#1E2026] border border-[#383A42] rounded pl-8 pr-3 text-xs text-white placeholder-[#666] focus:outline-none focus:border-[#F5C542]"
                />
              </div>
              <button
                onClick={() => setIsAddSymbolModalOpen(true)}
                className="px-3 py-1.5 bg-[#10B981] text-white text-xs font-bold rounded hover:bg-[#059669] transition-colors flex items-center gap-1"
              >
                <Plus size={12} />
                Add Symbol
              </button>
            </div>

            {/* Symbols Table */}
            <div className="flex-1 overflow-auto">
              <table className="w-full text-xs border-collapse">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42] z-10">
                  <tr>
                    <th onClick={() => handleSort('symbol')} className="text-left px-3 py-2 font-semibold text-[#888] cursor-pointer hover:text-white transition-colors">
                      Symbol {sortField === 'symbol' && (sortDirection === 'asc' ? '↑' : '↓')}
                    </th>
                    <th onClick={() => handleSort('leverage')} className="text-right px-3 py-2 font-semibold text-[#888] cursor-pointer hover:text-white transition-colors">
                      Leverage {sortField === 'leverage' && (sortDirection === 'asc' ? '↑' : '↓')}
                    </th>
                    <th onClick={() => handleSort('spreadType')} className="text-left px-3 py-2 font-semibold text-[#888] cursor-pointer hover:text-white transition-colors">
                      Spread {sortField === 'spreadType' && (sortDirection === 'asc' ? '↑' : '↓')}
                    </th>
                    <th onClick={() => handleSort('commissionModel')} className="text-left px-3 py-2 font-semibold text-[#888] cursor-pointer hover:text-white transition-colors">
                      Commission {sortField === 'commissionModel' && (sortDirection === 'asc' ? '↑' : '↓')}
                    </th>
                    <th onClick={() => handleSort('commissionRate')} className="text-right px-3 py-2 font-semibold text-[#888] cursor-pointer hover:text-white transition-colors">
                      Rate {sortField === 'commissionRate' && (sortDirection === 'asc' ? '↑' : '↓')}
                    </th>
                    <th onClick={() => handleSort('marginMode')} className="text-left px-3 py-2 font-semibold text-[#888] cursor-pointer hover:text-white transition-colors">
                      Margin {sortField === 'marginMode' && (sortDirection === 'asc' ? '↑' : '↓')}
                    </th>
                    <th onClick={() => handleSort('tradingHours')} className="text-left px-3 py-2 font-semibold text-[#888] cursor-pointer hover:text-white transition-colors">
                      Hours {sortField === 'tradingHours' && (sortDirection === 'asc' ? '↑' : '↓')}
                    </th>
                    <th className="text-center px-3 py-2 font-semibold text-[#888]">Active</th>
                    <th className="text-center px-3 py-2 font-semibold text-[#888]">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredSymbols.map((symbol, idx) => (
                    <tr
                      key={symbol.symbol}
                      className={`border-b border-[#383A42] ${idx % 2 === 0 ? 'bg-[#121316]' : 'bg-[#1A1C21]'} hover:bg-[#25272E] transition-colors`}
                    >
                      <td className="px-3 py-2 text-white font-semibold">{symbol.symbol}</td>
                      <td className="px-3 py-2 text-right text-white">1:{symbol.leverage}</td>
                      <td className="px-3 py-2">
                        <span className="px-2 py-0.5 rounded text-[10px] font-semibold" style={{ backgroundColor: getSpreadTypeBadgeColor(symbol.spreadType) + '33', color: getSpreadTypeBadgeColor(symbol.spreadType) }}>
                          {formatSpreadType(symbol.spreadType)}
                        </span>
                      </td>
                      <td className="px-3 py-2">
                        <span className="px-2 py-0.5 rounded text-[10px] font-semibold" style={{ backgroundColor: getCommissionBadgeColor(symbol.commissionModel) + '33', color: getCommissionBadgeColor(symbol.commissionModel) }}>
                          {formatCommissionModel(symbol.commissionModel)}
                        </span>
                      </td>
                      <td className="px-3 py-2 text-right text-white">
                        {symbol.commissionModel === 'none' ? '-' :
                         symbol.commissionModel === 'per_lot' ? `$${symbol.commissionRate}` :
                         `${symbol.commissionRate}%`}
                      </td>
                      <td className="px-3 py-2">
                        <span className="px-2 py-0.5 rounded text-[10px] font-semibold" style={{ backgroundColor: getMarginModeBadgeColor(symbol.marginMode) + '33', color: getMarginModeBadgeColor(symbol.marginMode) }}>
                          {formatMarginMode(symbol.marginMode)}
                        </span>
                      </td>
                      <td className="px-3 py-2 text-[#888]">{symbol.tradingHours}</td>
                      <td className="px-3 py-2 text-center">
                        <button
                          onClick={() => handleToggleSymbolActive(symbol)}
                          className={`w-10 h-5 rounded-full transition-colors ${symbol.isActive ? 'bg-[#10B981]' : 'bg-[#6B7280]'}`}
                        >
                          <div className={`w-4 h-4 bg-white rounded-full shadow transition-transform ${symbol.isActive ? 'translate-x-5' : 'translate-x-0.5'}`} />
                        </button>
                      </td>
                      <td className="px-3 py-2 text-center">
                        <button
                          onClick={() => handleRemoveSymbol(symbol.symbol)}
                          className="px-2 py-1 bg-[#EF4444] text-white text-[10px] rounded hover:bg-[#DC2626] transition-colors"
                        >
                          Remove
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {filteredSymbols.length === 0 && (
                <div className="flex items-center justify-center h-32 text-[#666]">
                  {searchQuery ? 'No symbols match your search' : 'No symbols in this group'}
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="flex-1 flex items-center justify-center bg-[#121316] text-[#666]">
            <div className="text-center">
              <Settings size={48} className="mx-auto mb-3 opacity-20" />
              <div className="text-sm">Select a group to view symbols</div>
            </div>
          </div>
        )}
      </div>

      {/* Create Group Modal */}
      {isCreateModalOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded shadow-xl w-[500px]">
            <div className="h-10 bg-[#25272E] border-b border-[#383A42] flex items-center px-3">
              <span className="text-sm font-semibold text-white">Create Instrument Group</span>
            </div>
            <div className="p-4 space-y-3 max-h-[600px] overflow-y-auto">
              <div>
                <label className="block text-xs text-[#888] mb-1">Group Name *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  placeholder="e.g., Forex Majors"
                />
              </div>
              <div>
                <label className="block text-xs text-[#888] mb-1">Description *</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="w-full h-16 bg-[#121316] border border-[#383A42] rounded px-2 py-1.5 text-xs text-white resize-none focus:outline-none focus:border-[#F5C542]"
                  placeholder="Describe this instrument group..."
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-[#888] mb-1">Default Leverage *</label>
                  <input
                    type="number"
                    value={formData.defaultLeverage}
                    onChange={(e) => setFormData({ ...formData, defaultLeverage: parseInt(e.target.value) || 100 })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  />
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Spread Type *</label>
                  <select
                    value={formData.spreadType}
                    onChange={(e) => setFormData({ ...formData, spreadType: e.target.value as SpreadType })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  >
                    <option value="fixed">Fixed</option>
                    <option value="variable">Variable</option>
                  </select>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-[#888] mb-1">Commission Model *</label>
                  <select
                    value={formData.commissionModel}
                    onChange={(e) => setFormData({ ...formData, commissionModel: e.target.value as CommissionModel })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  >
                    <option value="none">None</option>
                    <option value="per_lot">Per Lot</option>
                    <option value="percentage">Percentage</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Commission Rate</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.commissionRate}
                    onChange={(e) => setFormData({ ...formData, commissionRate: parseFloat(e.target.value) || 0 })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                    disabled={formData.commissionModel === 'none'}
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-[#888] mb-1">Margin Mode *</label>
                  <select
                    value={formData.marginMode}
                    onChange={(e) => setFormData({ ...formData, marginMode: e.target.value as MarginMode })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  >
                    <option value="retail">Retail</option>
                    <option value="professional">Professional</option>
                    <option value="hedge">Hedge</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Trading Hours Template *</label>
                  <input
                    type="text"
                    value={formData.tradingHoursTemplate}
                    onChange={(e) => setFormData({ ...formData, tradingHoursTemplate: e.target.value })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                    placeholder="e.g., 24/5, 24/7, Market Hours"
                  />
                </div>
              </div>
            </div>
            <div className="h-12 bg-[#25272E] border-t border-[#383A42] flex items-center justify-end px-3 gap-2">
              <button
                onClick={() => {
                  setIsCreateModalOpen(false);
                  resetForm();
                }}
                className="px-4 py-1.5 bg-[#383A42] text-white text-xs font-bold rounded hover:bg-[#4A4C54] transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateGroup}
                className="px-4 py-1.5 bg-[#F5C542] text-[#121316] text-xs font-bold rounded hover:bg-[#E5B532] transition-colors"
              >
                Create Group
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Group Modal */}
      {isEditModalOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded shadow-xl w-[500px]">
            <div className="h-10 bg-[#25272E] border-b border-[#383A42] flex items-center px-3">
              <span className="text-sm font-semibold text-white">Edit Group: {selectedGroup?.name}</span>
            </div>
            <div className="p-4 space-y-3 max-h-[600px] overflow-y-auto">
              <div>
                <label className="block text-xs text-[#888] mb-1">Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="w-full h-16 bg-[#121316] border border-[#383A42] rounded px-2 py-1.5 text-xs text-white resize-none focus:outline-none focus:border-[#F5C542]"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-[#888] mb-1">Default Leverage</label>
                  <input
                    type="number"
                    value={formData.defaultLeverage}
                    onChange={(e) => setFormData({ ...formData, defaultLeverage: parseInt(e.target.value) || 100 })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  />
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Spread Type</label>
                  <select
                    value={formData.spreadType}
                    onChange={(e) => setFormData({ ...formData, spreadType: e.target.value as SpreadType })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  >
                    <option value="fixed">Fixed</option>
                    <option value="variable">Variable</option>
                  </select>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-[#888] mb-1">Commission Model</label>
                  <select
                    value={formData.commissionModel}
                    onChange={(e) => setFormData({ ...formData, commissionModel: e.target.value as CommissionModel })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  >
                    <option value="none">None</option>
                    <option value="per_lot">Per Lot</option>
                    <option value="percentage">Percentage</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Commission Rate</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.commissionRate}
                    onChange={(e) => setFormData({ ...formData, commissionRate: parseFloat(e.target.value) || 0 })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                    disabled={formData.commissionModel === 'none'}
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-[#888] mb-1">Margin Mode</label>
                  <select
                    value={formData.marginMode}
                    onChange={(e) => setFormData({ ...formData, marginMode: e.target.value as MarginMode })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  >
                    <option value="retail">Retail</option>
                    <option value="professional">Professional</option>
                    <option value="hedge">Hedge</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Trading Hours Template</label>
                  <input
                    type="text"
                    value={formData.tradingHoursTemplate}
                    onChange={(e) => setFormData({ ...formData, tradingHoursTemplate: e.target.value })}
                    className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                  />
                </div>
              </div>
            </div>
            <div className="h-12 bg-[#25272E] border-t border-[#383A42] flex items-center justify-end px-3 gap-2">
              <button
                onClick={() => {
                  setIsEditModalOpen(false);
                  resetForm();
                }}
                className="px-4 py-1.5 bg-[#383A42] text-white text-xs font-bold rounded hover:bg-[#4A4C54] transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleUpdateGroup}
                className="px-4 py-1.5 bg-[#3B82F6] text-white text-xs font-bold rounded hover:bg-[#2563EB] transition-colors"
              >
                Update Group
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Symbol Modal */}
      {isAddSymbolModalOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded shadow-xl w-[400px]">
            <div className="h-10 bg-[#25272E] border-b border-[#383A42] flex items-center px-3">
              <span className="text-sm font-semibold text-white">Add Symbol to {selectedGroup?.name}</span>
            </div>
            <div className="p-4">
              <label className="block text-xs text-[#888] mb-1">Symbol Name *</label>
              <input
                type="text"
                value={newSymbol}
                onChange={(e) => setNewSymbol(e.target.value.toUpperCase())}
                className="w-full h-8 bg-[#121316] border border-[#383A42] rounded px-2 text-xs text-white focus:outline-none focus:border-[#F5C542]"
                placeholder="e.g., EURUSD"
              />
              <div className="text-xs text-[#666] mt-2">
                Symbol will inherit group's default trading conditions (leverage, spread, commission, margin mode, trading hours).
              </div>
            </div>
            <div className="h-12 bg-[#25272E] border-t border-[#383A42] flex items-center justify-end px-3 gap-2">
              <button
                onClick={() => {
                  setIsAddSymbolModalOpen(false);
                  setNewSymbol('');
                }}
                className="px-4 py-1.5 bg-[#383A42] text-white text-xs font-bold rounded hover:bg-[#4A4C54] transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleAddSymbol}
                className="px-4 py-1.5 bg-[#10B981] text-white text-xs font-bold rounded hover:bg-[#059669] transition-colors"
              >
                Add Symbol
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
