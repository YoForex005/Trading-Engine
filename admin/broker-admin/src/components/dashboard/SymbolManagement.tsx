'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
    Search,
    Plus,
    Edit2,
    Power,
    X,
    Loader2,
    ChevronLeft,
    ChevronRight,
    ArrowUpDown,
    Coins,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';
import { api } from '@/services/apiClient';

type SymbolType = 'Forex' | 'Crypto' | 'Index' | 'Commodity';

interface Symbol {
    symbol: string;
    description: string;
    type: SymbolType;
    digits: number;
    spreadMarkup: number;
    minSpread: number;
    maxSpread: number;
    commission: number;
    swapLong: number;
    swapShort: number;
    tradingSessions: string;
    status: 'active' | 'disabled';
}

type SortColumn = keyof Symbol | null;
type SortDirection = 'asc' | 'desc';

const ITEMS_PER_PAGE = 25;

export default function SymbolManagement() {
    const [symbols, setSymbols] = useState<Symbol[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const [sortColumn, setSortColumn] = useState<SortColumn>('symbol');
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

    // Modal states
    const [showEditModal, setShowEditModal] = useState(false);
    const [showAddModal, setShowAddModal] = useState(false);
    const [editingSymbol, setEditingSymbol] = useState<Symbol | null>(null);
    const [modalLoading, setModalLoading] = useState(false);

    // Form state
    const [formData, setFormData] = useState<Partial<Symbol>>({
        symbol: '',
        description: '',
        type: 'Forex',
        digits: 5,
        spreadMarkup: 0,
        minSpread: 0,
        maxSpread: 0,
        commission: 0,
        swapLong: 0,
        swapShort: 0,
        tradingSessions: 'Mon-Fri 00:00-23:59',
        status: 'active',
    });

    /**
     * Fetch symbols from API
     */
    const fetchSymbols = async () => {
        try {
            setLoading(true);
            setError(null);
            const data = await api.get<Symbol[]>(API_CONFIG.SYMBOLS_LIST);
            setSymbols(data);
        } catch (err: any) {
            setError(err.message || 'Failed to load symbols');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSymbols();
    }, []);

    /**
     * Handle column sort
     */
    const handleSort = (column: SortColumn) => {
        if (sortColumn === column) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortColumn(column);
            setSortDirection('asc');
        }
    };

    /**
     * Filter and sort symbols
     */
    const filteredAndSortedSymbols = useMemo(() => {
        let filtered = symbols.filter((symbol) =>
            symbol.symbol.toLowerCase().includes(searchQuery.toLowerCase()) ||
            symbol.description.toLowerCase().includes(searchQuery.toLowerCase())
        );

        if (sortColumn) {
            filtered.sort((a, b) => {
                const aVal = a[sortColumn];
                const bVal = b[sortColumn];

                if (typeof aVal === 'number' && typeof bVal === 'number') {
                    return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
                }

                const aStr = String(aVal || '');
                const bStr = String(bVal || '');
                return sortDirection === 'asc'
                    ? aStr.localeCompare(bStr)
                    : bStr.localeCompare(aStr);
            });
        }

        return filtered;
    }, [symbols, searchQuery, sortColumn, sortDirection]);

    /**
     * Paginate symbols
     */
    const paginatedSymbols = useMemo(() => {
        const start = (currentPage - 1) * ITEMS_PER_PAGE;
        const end = start + ITEMS_PER_PAGE;
        return filteredAndSortedSymbols.slice(start, end);
    }, [filteredAndSortedSymbols, currentPage]);

    const totalPages = Math.ceil(filteredAndSortedSymbols.length / ITEMS_PER_PAGE);

    /**
     * Open Edit Modal
     */
    const handleEdit = async (symbol: Symbol) => {
        setEditingSymbol(symbol);
        setFormData(symbol);
        setShowEditModal(true);
    };

    /**
     * Open Add Modal
     */
    const handleAdd = () => {
        setEditingSymbol(null);
        setFormData({
            symbol: '',
            description: '',
            type: 'Forex',
            digits: 5,
            spreadMarkup: 0,
            minSpread: 0,
            maxSpread: 0,
            commission: 0,
            swapLong: 0,
            swapShort: 0,
            tradingSessions: 'Mon-Fri 00:00-23:59',
            status: 'active',
        });
        setShowAddModal(true);
    };

    /**
     * Toggle symbol status (enable/disable)
     */
    const handleToggleStatus = async (symbol: Symbol) => {
        try {
            const newStatus = symbol.status === 'active' ? 'disabled' : 'active';
            await api.put(API_CONFIG.SYMBOL_UPDATE_STATUS(symbol.symbol), { status: newStatus });

            // Update local state
            setSymbols(symbols.map(s =>
                s.symbol === symbol.symbol ? { ...s, status: newStatus } : s
            ));
        } catch (err: any) {
            alert(`Failed to toggle status: ${err.message}`);
        }
    };

    /**
     * Save symbol (Add or Edit)
     */
    const handleSave = async () => {
        try {
            setModalLoading(true);

            if (showAddModal) {
                // Create new symbol
                await api.post(API_CONFIG.SYMBOL_CREATE, formData);
            } else if (showEditModal && editingSymbol) {
                // Update existing symbol
                await api.put(API_CONFIG.SYMBOL_UPDATE_SPREAD(editingSymbol.symbol), formData);
            }

            // Refresh symbols list
            await fetchSymbols();

            // Close modal
            setShowEditModal(false);
            setShowAddModal(false);
            setEditingSymbol(null);
        } catch (err: any) {
            alert(`Failed to save symbol: ${err.message}`);
        } finally {
            setModalLoading(false);
        }
    };

    /**
     * Close modal
     */
    const handleCloseModal = () => {
        setShowEditModal(false);
        setShowAddModal(false);
        setEditingSymbol(null);
    };

    if (loading) {
        return (
            <div className="h-full flex items-center justify-center bg-[#0A0A0B]">
                <div className="text-center">
                    <Loader2 className="w-8 h-8 animate-spin text-blue-500 mx-auto mb-3" />
                    <p className="text-zinc-400">Loading symbols...</p>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="h-full flex items-center justify-center bg-[#0A0A0B]">
                <div className="text-center">
                    <p className="text-red-300 mb-4">{error}</p>
                    <button
                        onClick={fetchSymbols}
                        className="px-4 py-2 bg-zinc-800 text-white rounded hover:bg-zinc-700"
                    >
                        Retry
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="h-full bg-[#0A0A0B] p-4 overflow-hidden flex flex-col">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                    <Coins className="w-6 h-6 text-yellow-400" />
                    <h1 className="text-2xl font-bold text-white">Symbol Management</h1>
                </div>
                <div className="flex items-center gap-3">
                    {/* Search */}
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-zinc-400" />
                        <input
                            type="text"
                            value={searchQuery}
                            onChange={(e) => {
                                setSearchQuery(e.target.value);
                                setCurrentPage(1);
                            }}
                            placeholder="Search symbols..."
                            className="pl-10 pr-4 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 w-64"
                        />
                    </div>
                    {/* Add Symbol Button */}
                    <button
                        onClick={handleAdd}
                        className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors flex items-center gap-2 font-semibold"
                    >
                        <Plus className="w-4 h-4" />
                        Add Symbol
                    </button>
                </div>
            </div>

            {/* Symbols Table */}
            <div className="flex-1 bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden flex flex-col">
                <div className="overflow-auto flex-1">
                    <table className="w-full text-sm">
                        <thead className="bg-zinc-800 sticky top-0">
                            <tr>
                                {[
                                    { key: 'symbol', label: 'Symbol' },
                                    { key: 'description', label: 'Description' },
                                    { key: 'type', label: 'Type' },
                                    { key: 'digits', label: 'Digits' },
                                    { key: 'spreadMarkup', label: 'Spread Markup' },
                                    { key: 'minSpread', label: 'Min Spread' },
                                    { key: 'maxSpread', label: 'Max Spread' },
                                    { key: 'status', label: 'Status' },
                                ].map((col) => (
                                    <th
                                        key={col.key}
                                        onClick={() => handleSort(col.key as SortColumn)}
                                        className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider cursor-pointer hover:bg-zinc-700 transition-colors"
                                    >
                                        <div className="flex items-center gap-2">
                                            {col.label}
                                            <ArrowUpDown className="w-3 h-3" />
                                        </div>
                                    </th>
                                ))}
                                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {paginatedSymbols.length === 0 ? (
                                <tr>
                                    <td colSpan={9} className="px-4 py-12 text-center text-zinc-500">
                                        {searchQuery ? 'No symbols match your search' : 'No symbols configured'}
                                    </td>
                                </tr>
                            ) : (
                                paginatedSymbols.map((symbol) => (
                                    <tr
                                        key={symbol.symbol}
                                        className="border-b border-zinc-800 hover:bg-zinc-800/50 transition-colors"
                                    >
                                        <td className="px-4 py-3 text-white font-bold">{symbol.symbol}</td>
                                        <td className="px-4 py-3 text-zinc-300">{symbol.description}</td>
                                        <td className="px-4 py-3 text-zinc-300">{symbol.type}</td>
                                        <td className="px-4 py-3 text-zinc-300">{symbol.digits}</td>
                                        <td className="px-4 py-3 text-zinc-300">{symbol.spreadMarkup}</td>
                                        <td className="px-4 py-3 text-zinc-300">{symbol.minSpread}</td>
                                        <td className="px-4 py-3 text-zinc-300">{symbol.maxSpread}</td>
                                        <td className="px-4 py-3">
                                            <span
                                                className={`px-2 py-1 rounded text-xs font-semibold ${
                                                    symbol.status === 'active'
                                                        ? 'bg-green-500/20 text-green-400'
                                                        : 'bg-red-500/20 text-red-400'
                                                }`}
                                            >
                                                {symbol.status.charAt(0).toUpperCase() + symbol.status.slice(1)}
                                            </span>
                                        </td>
                                        <td className="px-4 py-3">
                                            <div className="flex items-center gap-2">
                                                <button
                                                    onClick={() => handleEdit(symbol)}
                                                    className="p-1.5 bg-blue-500/20 text-blue-400 rounded hover:bg-blue-500/30 transition-colors"
                                                    title="Edit"
                                                >
                                                    <Edit2 className="w-4 h-4" />
                                                </button>
                                                <button
                                                    onClick={() => handleToggleStatus(symbol)}
                                                    className={`p-1.5 rounded transition-colors ${
                                                        symbol.status === 'active'
                                                            ? 'bg-red-500/20 text-red-400 hover:bg-red-500/30'
                                                            : 'bg-green-500/20 text-green-400 hover:bg-green-500/30'
                                                    }`}
                                                    title={symbol.status === 'active' ? 'Disable' : 'Enable'}
                                                >
                                                    <Power className="w-4 h-4" />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                {/* Pagination */}
                {totalPages > 1 && (
                    <div className="p-4 border-t border-zinc-800 flex items-center justify-between">
                        <div className="text-sm text-zinc-400">
                            Showing {(currentPage - 1) * ITEMS_PER_PAGE + 1} to{' '}
                            {Math.min(currentPage * ITEMS_PER_PAGE, filteredAndSortedSymbols.length)} of{' '}
                            {filteredAndSortedSymbols.length} symbols
                        </div>
                        <div className="flex items-center gap-2">
                            <button
                                onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                                disabled={currentPage === 1}
                                className="p-2 bg-zinc-800 text-white rounded hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                <ChevronLeft className="w-4 h-4" />
                            </button>
                            <span className="text-sm text-white">
                                Page {currentPage} of {totalPages}
                            </span>
                            <button
                                onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                                disabled={currentPage === totalPages}
                                className="p-2 bg-zinc-800 text-white rounded hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                <ChevronRight className="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* Edit/Add Symbol Modal */}
            {(showEditModal || showAddModal) && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
                    <div className="bg-zinc-900 rounded-lg border border-zinc-800 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
                        {/* Modal Header */}
                        <div className="p-4 border-b border-zinc-800 flex items-center justify-between sticky top-0 bg-zinc-900">
                            <h2 className="text-lg font-semibold text-white">
                                {showAddModal ? 'Add New Symbol' : `Edit Symbol: ${editingSymbol?.symbol}`}
                            </h2>
                            <button
                                onClick={handleCloseModal}
                                className="p-1 hover:bg-zinc-800 rounded transition-colors"
                            >
                                <X className="w-5 h-5 text-zinc-400" />
                            </button>
                        </div>

                        {/* Modal Body */}
                        <div className="p-6 space-y-4">
                            {/* Symbol Name */}
                            <div>
                                <label className="text-sm text-zinc-400 mb-1 block">Symbol Name</label>
                                <input
                                    type="text"
                                    value={formData.symbol}
                                    onChange={(e) => setFormData({ ...formData, symbol: e.target.value.toUpperCase() })}
                                    disabled={showEditModal}
                                    placeholder="e.g., EURUSD"
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                                />
                            </div>

                            {/* Description */}
                            <div>
                                <label className="text-sm text-zinc-400 mb-1 block">Description</label>
                                <input
                                    type="text"
                                    value={formData.description}
                                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                    placeholder="e.g., Euro vs US Dollar"
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                            </div>

                            {/* Type */}
                            <div>
                                <label className="text-sm text-zinc-400 mb-1 block">Type</label>
                                <select
                                    value={formData.type}
                                    onChange={(e) => setFormData({ ...formData, type: e.target.value as SymbolType })}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    <option value="Forex">Forex</option>
                                    <option value="Crypto">Crypto</option>
                                    <option value="Index">Index</option>
                                    <option value="Commodity">Commodity</option>
                                </select>
                            </div>

                            {/* Digits */}
                            <div>
                                <label className="text-sm text-zinc-400 mb-1 block">Digits (1-8)</label>
                                <input
                                    type="number"
                                    min="1"
                                    max="8"
                                    value={formData.digits}
                                    onChange={(e) => setFormData({ ...formData, digits: parseInt(e.target.value) || 5 })}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                            </div>

                            {/* Spread Settings - 3 columns */}
                            <div className="grid grid-cols-3 gap-4">
                                <div>
                                    <label className="text-sm text-zinc-400 mb-1 block">Spread Markup</label>
                                    <input
                                        type="number"
                                        step="0.1"
                                        value={formData.spreadMarkup}
                                        onChange={(e) => setFormData({ ...formData, spreadMarkup: parseFloat(e.target.value) || 0 })}
                                        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>
                                <div>
                                    <label className="text-sm text-zinc-400 mb-1 block">Min Spread</label>
                                    <input
                                        type="number"
                                        step="0.1"
                                        value={formData.minSpread}
                                        onChange={(e) => setFormData({ ...formData, minSpread: parseFloat(e.target.value) || 0 })}
                                        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>
                                <div>
                                    <label className="text-sm text-zinc-400 mb-1 block">Max Spread</label>
                                    <input
                                        type="number"
                                        step="0.1"
                                        value={formData.maxSpread}
                                        onChange={(e) => setFormData({ ...formData, maxSpread: parseFloat(e.target.value) || 0 })}
                                        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>
                            </div>

                            {/* Commission */}
                            <div>
                                <label className="text-sm text-zinc-400 mb-1 block">Commission per Lot</label>
                                <input
                                    type="number"
                                    step="0.01"
                                    value={formData.commission}
                                    onChange={(e) => setFormData({ ...formData, commission: parseFloat(e.target.value) || 0 })}
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                            </div>

                            {/* Swap Settings - 2 columns */}
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="text-sm text-zinc-400 mb-1 block">Swap Long</label>
                                    <input
                                        type="number"
                                        step="0.01"
                                        value={formData.swapLong}
                                        onChange={(e) => setFormData({ ...formData, swapLong: parseFloat(e.target.value) || 0 })}
                                        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>
                                <div>
                                    <label className="text-sm text-zinc-400 mb-1 block">Swap Short</label>
                                    <input
                                        type="number"
                                        step="0.01"
                                        value={formData.swapShort}
                                        onChange={(e) => setFormData({ ...formData, swapShort: parseFloat(e.target.value) || 0 })}
                                        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>
                            </div>

                            {/* Trading Sessions */}
                            <div>
                                <label className="text-sm text-zinc-400 mb-1 block">Trading Sessions</label>
                                <input
                                    type="text"
                                    value={formData.tradingSessions}
                                    onChange={(e) => setFormData({ ...formData, tradingSessions: e.target.value })}
                                    placeholder="e.g., Mon-Fri 00:00-23:59"
                                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                            </div>
                        </div>

                        {/* Modal Footer */}
                        <div className="p-4 border-t border-zinc-800 flex items-center justify-end gap-3">
                            <button
                                onClick={handleCloseModal}
                                className="px-4 py-2 bg-zinc-800 text-white rounded hover:bg-zinc-700 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleSave}
                                disabled={modalLoading || !formData.symbol}
                                className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                            >
                                {modalLoading ? (
                                    <>
                                        <Loader2 className="w-4 h-4 animate-spin" />
                                        Saving...
                                    </>
                                ) : (
                                    'Save'
                                )}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
