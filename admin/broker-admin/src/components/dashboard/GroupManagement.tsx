/**
 * Trading Group Management
 * Manage trading groups with spread markups, commissions, leverage, and symbol access
 */

'use client';

import React, { useState, useMemo } from 'react';
import {
  Plus,
  Edit,
  Trash2,
  Users,
  TrendingUp,
  DollarSign,
  Shield,
  ArrowUpDown,
  X,
  Search,
  Check,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// Types
interface TradingGroup {
  id: string;
  name: string;
  description: string;
  clientCount: number;
  spreadMarkup: number; // percentage
  commissionPerLot: number; // USD
  maxLeverage: number;
  marginCallLevel: number; // percentage
  stopOutLevel: number; // percentage
  allowedSymbols: string[];
  status: 'active' | 'inactive';
  createdAt: number;
  updatedAt: number;
}

interface Client {
  id: string;
  name: string;
  email: string;
  groupId: string | null;
}

type SortColumn = 'name' | 'clientCount' | 'spreadMarkup' | 'commissionPerLot' | 'maxLeverage';
type SortDirection = 'asc' | 'desc';

// Mock data
const MOCK_GROUPS: TradingGroup[] = [
  {
    id: 'grp-1',
    name: 'Standard',
    description: 'Default group for retail clients',
    clientCount: 45,
    spreadMarkup: 0.5,
    commissionPerLot: 7.0,
    maxLeverage: 100,
    marginCallLevel: 100,
    stopOutLevel: 50,
    allowedSymbols: ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD'],
    status: 'active',
    createdAt: Date.now() - 86400000 * 30,
    updatedAt: Date.now() - 86400000 * 5,
  },
  {
    id: 'grp-2',
    name: 'VIP',
    description: 'Premium clients with better rates',
    clientCount: 12,
    spreadMarkup: 0.2,
    commissionPerLot: 5.0,
    maxLeverage: 200,
    marginCallLevel: 80,
    stopOutLevel: 40,
    allowedSymbols: ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'GOLD', 'SILVER'],
    status: 'active',
    createdAt: Date.now() - 86400000 * 60,
    updatedAt: Date.now() - 86400000 * 2,
  },
  {
    id: 'grp-3',
    name: 'Demo',
    description: 'Demo accounts',
    clientCount: 128,
    spreadMarkup: 1.0,
    commissionPerLot: 0.0,
    maxLeverage: 100,
    marginCallLevel: 100,
    stopOutLevel: 50,
    allowedSymbols: ['EURUSD', 'GBPUSD'],
    status: 'active',
    createdAt: Date.now() - 86400000 * 90,
    updatedAt: Date.now() - 86400000 * 1,
  },
];

const MOCK_CLIENTS: Client[] = [
  { id: 'c-1', name: 'John Doe', email: 'john@example.com', groupId: 'grp-1' },
  { id: 'c-2', name: 'Jane Smith', email: 'jane@example.com', groupId: 'grp-2' },
  { id: 'c-3', name: 'Bob Johnson', email: 'bob@example.com', groupId: null },
];

const ALL_SYMBOLS = [
  'EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'NZDUSD', 'USDCHF',
  'EURGBP', 'EURJPY', 'GBPJPY', 'GOLD', 'SILVER', 'BTCUSD', 'ETHUSD',
];

export default function GroupManagement() {
  const [groups, setGroups] = useState<TradingGroup[]>(MOCK_GROUPS);
  const [clients, setClients] = useState<Client[]>(MOCK_CLIENTS);
  const [editModal, setEditModal] = useState<{ mode: 'create' | 'edit'; group: TradingGroup | null } | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);
  const [assignModal, setAssignModal] = useState<string | null>(null);

  // Sorting
  const [sortColumn, setSortColumn] = useState<SortColumn>('name');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

  // Filter
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive'>('all');

  // Filtered and sorted groups
  const filteredGroups = useMemo(() => {
    let result = [...groups];

    // Filter by search
    if (searchQuery) {
      result = result.filter(
        (g) =>
          g.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
          g.description.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Filter by status
    if (statusFilter !== 'all') {
      result = result.filter((g) => g.status === statusFilter);
    }

    // Sort
    result.sort((a, b) => {
      let aVal: any = a[sortColumn];
      let bVal: any = b[sortColumn];

      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortDirection === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
      }

      return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
    });

    return result;
  }, [groups, searchQuery, statusFilter, sortColumn, sortDirection]);

  // Handle sort
  const handleSort = (column: SortColumn) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('asc');
    }
  };

  // Handle create/edit
  const handleSave = async (groupData: Partial<TradingGroup>) => {
    if (editModal?.mode === 'create') {
      const newGroup: TradingGroup = {
        id: `grp-${Date.now()}`,
        name: groupData.name || 'New Group',
        description: groupData.description || '',
        clientCount: 0,
        spreadMarkup: groupData.spreadMarkup || 0,
        commissionPerLot: groupData.commissionPerLot || 0,
        maxLeverage: groupData.maxLeverage || 100,
        marginCallLevel: groupData.marginCallLevel || 100,
        stopOutLevel: groupData.stopOutLevel || 50,
        allowedSymbols: groupData.allowedSymbols || [],
        status: 'active',
        createdAt: Date.now(),
        updatedAt: Date.now(),
      };
      setGroups([...groups, newGroup]);
    } else if (editModal?.mode === 'edit' && editModal.group) {
      setGroups(
        groups.map((g) =>
          g.id === editModal.group!.id
            ? { ...g, ...groupData, updatedAt: Date.now() }
            : g
        )
      );
    }
    setEditModal(null);
  };

  // Handle delete
  const handleDelete = async (id: string) => {
    // In production: POST to API_CONFIG.ADMIN_GROUPS_DELETE(id)
    setGroups(groups.filter((g) => g.id !== id));
    setDeleteConfirm(null);
  };

  // Handle assign client to group
  const handleAssignClient = (clientId: string, groupId: string) => {
    setClients(clients.map((c) => (c.id === clientId ? { ...c, groupId } : c)));
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-zinc-300 p-4 overflow-auto custom-scrollbar">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-zinc-200">Trading Groups</h1>
          <p className="text-sm text-zinc-500 mt-1">
            Manage client groups with custom spreads, commissions, and leverage
          </p>
        </div>
        <button
          onClick={() =>
            setEditModal({
              mode: 'create',
              group: null,
            })
          }
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={16} />
          Create Group
        </button>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <div className="flex-1 min-w-[200px]">
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search groups..."
              className="w-full pl-9 pr-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-zinc-300 text-sm focus:outline-none focus:border-blue-500"
            />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <label className="text-xs text-zinc-500">Status:</label>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as any)}
            className="px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-zinc-300 text-sm focus:outline-none focus:border-blue-500"
          >
            <option value="all">All</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded overflow-hidden flex-1">
        <table className="w-full text-xs">
          <thead className="bg-[#252525] border-b border-zinc-700">
            <tr>
              <th
                className="px-4 py-3 text-left text-zinc-500 font-semibold cursor-pointer hover:text-zinc-300"
                onClick={() => handleSort('name')}
              >
                <div className="flex items-center gap-1">
                  Group Name
                  <ArrowUpDown size={12} />
                </div>
              </th>
              <th
                className="px-4 py-3 text-center text-zinc-500 font-semibold cursor-pointer hover:text-zinc-300"
                onClick={() => handleSort('clientCount')}
              >
                <div className="flex items-center gap-1 justify-center">
                  Clients
                  <ArrowUpDown size={12} />
                </div>
              </th>
              <th
                className="px-4 py-3 text-center text-zinc-500 font-semibold cursor-pointer hover:text-zinc-300"
                onClick={() => handleSort('spreadMarkup')}
              >
                <div className="flex items-center gap-1 justify-center">
                  Spread Markup
                  <ArrowUpDown size={12} />
                </div>
              </th>
              <th
                className="px-4 py-3 text-center text-zinc-500 font-semibold cursor-pointer hover:text-zinc-300"
                onClick={() => handleSort('commissionPerLot')}
              >
                <div className="flex items-center gap-1 justify-center">
                  Commission
                  <ArrowUpDown size={12} />
                </div>
              </th>
              <th
                className="px-4 py-3 text-center text-zinc-500 font-semibold cursor-pointer hover:text-zinc-300"
                onClick={() => handleSort('maxLeverage')}
              >
                <div className="flex items-center gap-1 justify-center">
                  Max Leverage
                  <ArrowUpDown size={12} />
                </div>
              </th>
              <th className="px-4 py-3 text-center text-zinc-500 font-semibold">Status</th>
              <th className="px-4 py-3 text-center text-zinc-500 font-semibold">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredGroups.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-zinc-500">
                  No groups found
                </td>
              </tr>
            ) : (
              filteredGroups.map((group) => (
                <tr
                  key={group.id}
                  className="border-b border-zinc-700/30 hover:bg-zinc-800/30"
                >
                  <td className="px-4 py-3">
                    <div>
                      <div className="font-semibold text-zinc-200">{group.name}</div>
                      <div className="text-zinc-500 text-[10px] mt-0.5">
                        {group.description}
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-center">
                    <div className="flex items-center justify-center gap-1">
                      <Users size={12} className="text-blue-400" />
                      <span className="font-semibold text-zinc-300">{group.clientCount}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-center">
                    <div className="flex items-center justify-center gap-1">
                      <TrendingUp size={12} className="text-emerald-400" />
                      <span className="font-mono text-zinc-300">{group.spreadMarkup.toFixed(1)}%</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-center">
                    <div className="flex items-center justify-center gap-1">
                      <DollarSign size={12} className="text-yellow-400" />
                      <span className="font-mono text-zinc-300">${group.commissionPerLot.toFixed(2)}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-center">
                    <div className="flex items-center justify-center gap-1">
                      <Shield size={12} className="text-purple-400" />
                      <span className="font-mono text-zinc-300">1:{group.maxLeverage}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-center">
                    {group.status === 'active' ? (
                      <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 font-semibold">
                        Active
                      </span>
                    ) : (
                      <span className="px-2 py-0.5 rounded bg-zinc-700/50 text-zinc-500 font-semibold">
                        Inactive
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-center gap-2">
                      <button
                        onClick={() => setAssignModal(group.id)}
                        className="p-1.5 rounded hover:bg-blue-500/20 text-blue-400 transition-colors"
                        title="Assign Clients"
                      >
                        <Users size={14} />
                      </button>
                      <button
                        onClick={() => setEditModal({ mode: 'edit', group })}
                        className="p-1.5 rounded hover:bg-zinc-700 text-zinc-400 transition-colors"
                        title="Edit Group"
                      >
                        <Edit size={14} />
                      </button>
                      <button
                        onClick={() => setDeleteConfirm(group.id)}
                        className="p-1.5 rounded hover:bg-rose-500/20 text-rose-400 transition-colors"
                        title="Delete Group"
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Create/Edit Modal */}
      {editModal && (
        <GroupEditModal
          mode={editModal.mode}
          group={editModal.group}
          onSave={handleSave}
          onClose={() => setEditModal(null)}
        />
      )}

      {/* Delete Confirmation Modal */}
      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg p-6 w-[400px]">
            <h3 className="text-lg font-bold text-zinc-200 mb-4">Confirm Deletion</h3>
            <p className="text-sm text-zinc-400 mb-6">
              Are you sure you want to delete this group? This action cannot be undone.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => setDeleteConfirm(null)}
                className="px-4 py-2 rounded bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-sm"
              >
                Cancel
              </button>
              <button
                onClick={() => handleDelete(deleteConfirm)}
                className="px-4 py-2 rounded bg-rose-600 hover:bg-rose-700 text-white text-sm font-medium"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Assign Clients Modal */}
      {assignModal && (
        <AssignClientsModal
          groupId={assignModal}
          clients={clients}
          onAssign={handleAssignClient}
          onClose={() => setAssignModal(null)}
        />
      )}
    </div>
  );
}

// Group Edit Modal Component
interface GroupEditModalProps {
  mode: 'create' | 'edit';
  group: TradingGroup | null;
  onSave: (data: Partial<TradingGroup>) => void;
  onClose: () => void;
}

function GroupEditModal({ mode, group, onSave, onClose }: GroupEditModalProps) {
  const [formData, setFormData] = useState<Partial<TradingGroup>>(
    group || {
      name: '',
      description: '',
      spreadMarkup: 0.5,
      commissionPerLot: 7.0,
      maxLeverage: 100,
      marginCallLevel: 100,
      stopOutLevel: 50,
      allowedSymbols: [],
    }
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  const toggleSymbol = (symbol: string) => {
    const current = formData.allowedSymbols || [];
    if (current.includes(symbol)) {
      setFormData({ ...formData, allowedSymbols: current.filter((s) => s !== symbol) });
    } else {
      setFormData({ ...formData, allowedSymbols: [...current, symbol] });
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg w-full max-w-2xl max-h-[90vh] overflow-y-auto custom-scrollbar">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 bg-[#252525] border-b border-zinc-700 sticky top-0">
          <h3 className="text-lg font-bold text-zinc-200">
            {mode === 'create' ? 'Create New Group' : 'Edit Group'}
          </h3>
          <button onClick={onClose} className="p-1 rounded hover:bg-zinc-700 text-zinc-400">
            <X size={16} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Name */}
          <div>
            <label className="block text-xs text-zinc-400 mb-2">Group Name *</label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm focus:outline-none focus:border-blue-500"
              placeholder="e.g., VIP Clients"
            />
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs text-zinc-400 mb-2">Description</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm focus:outline-none focus:border-blue-500 resize-none"
              rows={2}
              placeholder="Optional description..."
            />
          </div>

          {/* Spread & Commission */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs text-zinc-400 mb-2">Spread Markup (%)</label>
              <input
                type="number"
                step={0.1}
                min={0}
                value={formData.spreadMarkup}
                onChange={(e) =>
                  setFormData({ ...formData, spreadMarkup: parseFloat(e.target.value) || 0 })
                }
                className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm font-mono focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-xs text-zinc-400 mb-2">Commission per Lot ($)</label>
              <input
                type="number"
                step={0.5}
                min={0}
                value={formData.commissionPerLot}
                onChange={(e) =>
                  setFormData({ ...formData, commissionPerLot: parseFloat(e.target.value) || 0 })
                }
                className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm font-mono focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          {/* Leverage */}
          <div>
            <label className="block text-xs text-zinc-400 mb-2">Max Leverage</label>
            <select
              value={formData.maxLeverage}
              onChange={(e) =>
                setFormData({ ...formData, maxLeverage: parseInt(e.target.value) })
              }
              className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm focus:outline-none focus:border-blue-500"
            >
              <option value={30}>1:30</option>
              <option value={50}>1:50</option>
              <option value={100}>1:100</option>
              <option value={200}>1:200</option>
              <option value={400}>1:400</option>
              <option value={500}>1:500</option>
            </select>
          </div>

          {/* Margin Levels */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs text-zinc-400 mb-2">Margin Call Level (%)</label>
              <input
                type="number"
                step={5}
                min={0}
                max={200}
                value={formData.marginCallLevel}
                onChange={(e) =>
                  setFormData({ ...formData, marginCallLevel: parseFloat(e.target.value) || 100 })
                }
                className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm font-mono focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-xs text-zinc-400 mb-2">Stop Out Level (%)</label>
              <input
                type="number"
                step={5}
                min={0}
                max={100}
                value={formData.stopOutLevel}
                onChange={(e) =>
                  setFormData({ ...formData, stopOutLevel: parseFloat(e.target.value) || 50 })
                }
                className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm font-mono focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          {/* Allowed Symbols */}
          <div>
            <label className="block text-xs text-zinc-400 mb-2">
              Allowed Symbols ({formData.allowedSymbols?.length || 0} selected)
            </label>
            <div className="grid grid-cols-4 gap-2 p-3 bg-[#252525] border border-zinc-700 rounded max-h-48 overflow-y-auto custom-scrollbar">
              {ALL_SYMBOLS.map((symbol) => (
                <button
                  key={symbol}
                  type="button"
                  onClick={() => toggleSymbol(symbol)}
                  className={`px-2 py-1.5 rounded text-xs font-medium transition-colors ${
                    formData.allowedSymbols?.includes(symbol)
                      ? 'bg-blue-500/30 text-blue-400 border border-blue-500/50'
                      : 'bg-zinc-800 text-zinc-500 hover:bg-zinc-700 border border-zinc-700'
                  }`}
                >
                  {formData.allowedSymbols?.includes(symbol) && (
                    <Check size={10} className="inline mr-1" />
                  )}
                  {symbol}
                </button>
              ))}
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-3 justify-end pt-4 border-t border-zinc-700">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-sm"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 rounded bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium"
            >
              {mode === 'create' ? 'Create Group' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// Assign Clients Modal Component
interface AssignClientsModalProps {
  groupId: string;
  clients: Client[];
  onAssign: (clientId: string, groupId: string) => void;
  onClose: () => void;
}

function AssignClientsModal({ groupId, clients, onAssign, onClose }: AssignClientsModalProps) {
  const unassignedClients = clients.filter((c) => !c.groupId);

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg w-[500px] max-h-[600px] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 bg-[#252525] border-b border-zinc-700">
          <h3 className="text-lg font-bold text-zinc-200">Assign Clients to Group</h3>
          <button onClick={onClose} className="p-1 rounded hover:bg-zinc-700 text-zinc-400">
            <X size={16} />
          </button>
        </div>

        {/* Client List */}
        <div className="flex-1 overflow-y-auto custom-scrollbar p-4">
          {unassignedClients.length === 0 ? (
            <div className="text-center text-zinc-500 py-8">
              No unassigned clients available
            </div>
          ) : (
            <div className="space-y-2">
              {unassignedClients.map((client) => (
                <div
                  key={client.id}
                  className="flex items-center justify-between p-3 bg-[#252525] border border-zinc-700 rounded hover:border-zinc-600"
                >
                  <div>
                    <div className="text-sm font-medium text-zinc-200">{client.name}</div>
                    <div className="text-xs text-zinc-500">{client.email}</div>
                  </div>
                  <button
                    onClick={() => {
                      onAssign(client.id, groupId);
                      onClose();
                    }}
                    className="px-3 py-1.5 rounded bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium"
                  >
                    Assign
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
