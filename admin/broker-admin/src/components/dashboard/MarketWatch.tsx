'use client';

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { ChevronDown, ChevronUp, Plus, X, Search, Settings, RefreshCw } from 'lucide-react';
import { useWebSocket, MarketTick } from '@/hooks/useWebSocket';

const API_BASE = 'http://localhost:7999';
const WS_URL = 'ws://localhost:7999/ws';

interface SymbolData {
    symbol: string;
    bid: number;
    ask: number;
    spread: number;
    dailyChange: number;
    dailyChangePercent: number;
    lastUpdate: number;
    lp: string;
}

interface SymbolSpec {
    symbol: string;
    contractSize: number;
    pipSize: number;
    disabled: boolean;
}

// Default symbols shown initially
const DEFAULT_WATCHLIST = ['AUDCHF', 'AUDJPY', 'XAUUSD', 'BTCUSD', 'US500', 'US30', 'UT100'];

// Storage key for user's watchlist
const STORAGE_KEY = 'rtx_market_watchlist';

export default function MarketWatch() {
    const [watchlist, setWatchlist] = useState<string[]>([]);
    const [symbolData, setSymbolData] = useState<Record<string, SymbolData>>({});
    const [allSymbols, setAllSymbols] = useState<SymbolSpec[]>([]);
    const [showSymbolPicker, setShowSymbolPicker] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [sortColumn, setSortColumn] = useState<string | null>(null);
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');

    // Load watchlist from localStorage on mount
    useEffect(() => {
        const saved = localStorage.getItem(STORAGE_KEY);
        if (saved) {
            try {
                setWatchlist(JSON.parse(saved));
            } catch {
                setWatchlist(DEFAULT_WATCHLIST);
            }
        } else {
            setWatchlist(DEFAULT_WATCHLIST);
        }
    }, []);

    // Save watchlist to localStorage when it changes
    useEffect(() => {
        if (watchlist.length > 0) {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(watchlist));
        }
    }, [watchlist]);

    // Fetch all available symbols from API
    useEffect(() => {
        async function fetchSymbols() {
            try {
                const res = await fetch(`${API_BASE}/api/symbols`);
                if (res.ok) {
                    const data = await res.json();
                    setAllSymbols(data.symbols || []);
                }
            } catch (e) {
                console.error('Failed to fetch symbols:', e);
            }
        }
        fetchSymbols();
    }, []);

    // Handle incoming WebSocket ticks
    const handleTick = useCallback((tick: MarketTick) => {
        setSymbolData(prev => {
            const existing = prev[tick.symbol];
            const dailyChange = existing ? tick.bid - (existing.bid || tick.bid) : 0;
            const dailyChangePercent = existing && existing.bid > 0
                ? (dailyChange / existing.bid) * 100
                : 0;

            // Ensure spread is calculated if not provided
            const spread = tick.spread !== undefined && tick.spread > 0
                ? tick.spread
                : (tick.ask - tick.bid);

            return {
                ...prev,
                [tick.symbol]: {
                    symbol: tick.symbol,
                    bid: tick.bid,
                    ask: tick.ask,
                    spread: spread,
                    dailyChange: Math.abs(dailyChange) < 0.0001 ? (existing?.dailyChange || 0) : dailyChange,
                    dailyChangePercent: Math.abs(dailyChangePercent) < 0.0001 ? (existing?.dailyChangePercent || 0) : dailyChangePercent,
                    lastUpdate: tick.timestamp,
                    lp: tick.lp,
                }
            };
        });
    }, []);

    // Connect to WebSocket (no auth for now, can add JWT later)
    const { isConnected, error } = useWebSocket({
        url: WS_URL,
        onMessage: handleTick,
    });

    // Add symbol to watchlist
    const addSymbol = (symbol: string) => {
        if (!watchlist.includes(symbol)) {
            setWatchlist([...watchlist, symbol]);
        }
        setShowSymbolPicker(false);
        setSearchQuery('');
    };

    // Remove symbol from watchlist
    const removeSymbol = (symbol: string) => {
        setWatchlist(watchlist.filter(s => s !== symbol));
    };

    // Filtered symbols for picker
    const filteredSymbols = useMemo(() => {
        return allSymbols
            .filter(s => !s.disabled)
            .filter(s => !watchlist.includes(s.symbol))
            .filter(s => s.symbol.toLowerCase().includes(searchQuery.toLowerCase()));
    }, [allSymbols, watchlist, searchQuery]);

    // Sorted watchlist data
    const sortedWatchlist = useMemo(() => {
        const data = watchlist.map(symbol => symbolData[symbol] || {
            symbol,
            bid: 0,
            ask: 0,
            spread: 0,
            dailyChange: 0,
            dailyChangePercent: 0,
            lastUpdate: 0,
            lp: '-',
        });

        if (!sortColumn) return data;

        return [...data].sort((a, b) => {
            const aVal = a[sortColumn as keyof SymbolData];
            const bVal = b[sortColumn as keyof SymbolData];
            const multiplier = sortDirection === 'asc' ? 1 : -1;

            if (typeof aVal === 'number' && typeof bVal === 'number') {
                return (aVal - bVal) * multiplier;
            }
            return String(aVal).localeCompare(String(bVal)) * multiplier;
        });
    }, [watchlist, symbolData, sortColumn, sortDirection]);

    // Handle column sort
    const handleSort = (column: string) => {
        if (sortColumn === column) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortColumn(column);
            setSortDirection('asc');
        }
    };

    // Format price with appropriate decimals
    const formatPrice = (price: number, symbol: string) => {
        if (price === 0) return '-';
        // Crypto/Indices use 2 decimals, forex uses 5
        const isForex = /^[A-Z]{6}$/.test(symbol) && !symbol.includes('XAU') && !symbol.includes('BTC');
        return price.toFixed(isForex ? 5 : 2);
    };

    // Format spread in pips
    const formatSpread = (spread: number) => {
        if (!spread || spread === 0) return '-';
        // Display as integer pips (spread * 10000 for forex pairs)
        const pips = Math.round(spread * 10000);
        return pips > 0 ? pips.toString() : '-';
    };

    return (
        <div className="h-full w-64 bg-[#121316] border-r border-[#383A42] flex flex-col font-sans select-none text-[11px]">
            {/* Header */}
            <div className="px-2 py-1 flex items-center justify-between border-b border-[#383A42] bg-[#1E2026]">
                <div className="flex items-center gap-1">
                    <span className="text-[10px] font-bold text-[#888] uppercase tracking-wider">Market Watch:</span>
                    <span className="text-[10px] text-[#666]">
                        {new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                    </span>
                </div>
                <div className="flex items-center gap-1">
                    {isConnected ? (
                        <div className="w-2 h-2 bg-[#2ECC71] rounded-full" title="Connected" />
                    ) : (
                        <div className="w-2 h-2 bg-[#E74C3C] rounded-full" title={error || 'Disconnected'} />
                    )}
                    <RefreshCw size={10} className="text-[#666] hover:text-white cursor-pointer" />
                    <Settings size={10} className="text-[#666] hover:text-white cursor-pointer" />
                </div>
            </div>

            {/* Column Headers */}
            <div className="grid grid-cols-[1fr_60px_60px_35px_55px] gap-0 bg-[#1E2026] border-b border-[#383A42] text-[#888] text-[9px]">
                <div className="px-2 py-1 cursor-pointer hover:text-white" onClick={() => handleSort('symbol')}>
                    Symbol {sortColumn === 'symbol' && (sortDirection === 'asc' ? '▲' : '▼')}
                </div>
                <div className="px-1 py-1 text-right cursor-pointer hover:text-white" onClick={() => handleSort('bid')}>
                    Bid
                </div>
                <div className="px-1 py-1 text-right cursor-pointer hover:text-white" onClick={() => handleSort('ask')}>
                    Ask
                </div>
                <div className="px-1 py-1 text-right cursor-pointer hover:text-white" onClick={() => handleSort('spread')}>
                    Spr...
                </div>
                <div className="px-1 py-1 text-right cursor-pointer hover:text-white" onClick={() => handleSort('dailyChangePercent')}>
                    Daily C...
                </div>
            </div>

            {/* Symbol Rows */}
            <div className="flex-1 overflow-y-auto custom-scrollbar">
                {sortedWatchlist.map((data, index) => {
                    const changeColor = data.dailyChangePercent >= 0 ? 'text-[#2ECC71]' : 'text-[#E74C3C]';
                    const bidColor = data.bid > 0 ? 'text-[#E74C3C]' : 'text-[#888]';
                    const askColor = data.ask > 0 ? 'text-[#3B82F6]' : 'text-[#888]';

                    return (
                        <div
                            key={data.symbol}
                            className="grid grid-cols-[1fr_60px_60px_35px_55px] gap-0 border-b border-[#252526] hover:bg-[#25272E] group"
                            onContextMenu={(e) => {
                                e.preventDefault();
                                removeSymbol(data.symbol);
                            }}
                        >
                            <div className="px-2 py-1 flex items-center gap-1 text-[#CCC]">
                                <div className={`w-1.5 h-1.5 ${data.dailyChangePercent >= 0 ? 'bg-[#2ECC71]' : 'bg-[#E74C3C]'}`} />
                                {data.symbol}
                            </div>
                            <div className={`px-1 py-1 text-right font-mono ${bidColor}`}>
                                {formatPrice(data.bid, data.symbol)}
                            </div>
                            <div className={`px-1 py-1 text-right font-mono ${askColor}`}>
                                {formatPrice(data.ask, data.symbol)}
                            </div>
                            <div className="px-1 py-1 text-right font-mono text-[#CCC]">
                                {formatSpread(data.spread)}
                            </div>
                            <div className={`px-1 py-1 text-right font-mono ${changeColor}`}>
                                {data.dailyChangePercent.toFixed(2)}%
                            </div>
                        </div>
                    );
                })}

                {/* Click to Add Row */}
                <div
                    className="px-2 py-1 text-[#666] hover:text-[#3B82F6] hover:bg-[#25272E] cursor-pointer flex items-center gap-1"
                    onClick={() => setShowSymbolPicker(true)}
                >
                    <Plus size={10} />
                    click to add...
                </div>
            </div>

            {/* Footer Tabs */}
            <div className="border-t border-[#383A42] bg-[#1E2026] flex text-[9px] text-[#888]">
                <div className="px-2 py-1 border-b-2 border-[#3B82F6] text-white bg-[#252526]">Symbols</div>
                <div className="px-2 py-1 hover:text-white cursor-pointer">Details</div>
                <div className="px-2 py-1 hover:text-white cursor-pointer">Trading</div>
                <div className="px-2 py-1 hover:text-white cursor-pointer">Ticks</div>
            </div>

            {/* Symbol Count */}
            <div className="px-2 py-0.5 text-[9px] text-[#666] border-t border-[#383A42] bg-[#1E2026]">
                {watchlist.length} / {allSymbols.filter(s => !s.disabled).length}
            </div>

            {/* Symbol Picker Modal */}
            {showSymbolPicker && (
                <div className="absolute inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-[#1E1E1E] border border-[#444] w-[300px] max-h-[400px] flex flex-col">
                        <div className="flex items-center justify-between p-2 border-b border-[#444] bg-[#2D2D30]">
                            <span className="text-white font-bold text-xs">Add Symbol</span>
                            <X size={14} className="text-[#888] hover:text-white cursor-pointer" onClick={() => setShowSymbolPicker(false)} />
                        </div>
                        <div className="p-2 border-b border-[#444]">
                            <div className="flex items-center gap-1 bg-[#252526] border border-[#444] px-2">
                                <Search size={12} className="text-[#888]" />
                                <input
                                    type="text"
                                    placeholder="Search symbols..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    className="flex-1 bg-transparent border-none outline-none h-6 text-white text-xs placeholder-[#666]"
                                    autoFocus
                                />
                            </div>
                        </div>
                        <div className="flex-1 overflow-y-auto max-h-[300px]">
                            {filteredSymbols.map(spec => (
                                <div
                                    key={spec.symbol}
                                    className="px-3 py-1.5 hover:bg-[#3399FF] hover:text-white cursor-pointer text-[#CCC] text-xs"
                                    onClick={() => addSymbol(spec.symbol)}
                                >
                                    {spec.symbol}
                                </div>
                            ))}
                            {filteredSymbols.length === 0 && (
                                <div className="px-3 py-4 text-[#666] text-xs text-center">
                                    No symbols found
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
