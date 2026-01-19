import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Search, Check, ChevronRight, Clock, Plus } from 'lucide-react';

// API Base URL
const API_BASE = 'http://localhost:7999';

// Available symbol from API
interface AvailableSymbol {
    symbol: string;
    name: string;
    category: string;
    digits: number;
    subscribed?: boolean;
}

interface Tick {
    symbol: string;
    bid: number;
    ask: number;
    spread?: number;
    prevBid?: number;
    dailyChange?: number;
    high?: number;
    low?: number;
    volume?: number;
    last?: number;
    open?: number;
    close?: number; // Previous close
    tickHistory?: number[]; // For tick chart
}

interface MarketWatchPanelProps {
    ticks: Record<string, Tick>;
    allSymbols: any[];
    selectedSymbol: string;
    onSymbolSelect: (symbol: string) => void;
    className?: string;
}

type ColumnId = 'symbol' | 'bid' | 'ask' | 'spread' | 'dailyChange' | 'last' | 'high' | 'low' | 'volume' | 'time';
type TabId = 'symbols' | 'details' | 'trading' | 'ticks';

interface ColumnConfig {
    id: ColumnId;
    label: string;
    width: string;
    align: 'left' | 'right' | 'center';
    locked?: boolean; // If true, cannot be hidden
}

const ALL_COLUMNS: ColumnConfig[] = [
    { id: 'symbol', label: 'Symbol', width: 'flex-1', align: 'left', locked: true },
    { id: 'bid', label: 'Bid', width: 'w-16', align: 'right', locked: true },
    { id: 'ask', label: 'Ask', width: 'w-16', align: 'right', locked: true },
    { id: 'spread', label: '!', width: 'w-8', align: 'center', locked: true },
    { id: 'dailyChange', label: 'Daily %', width: 'w-14', align: 'right' },
    { id: 'last', label: 'Last', width: 'w-16', align: 'right' },
    { id: 'high', label: 'High', width: 'w-16', align: 'right' },
    { id: 'low', label: 'Low', width: 'w-16', align: 'right' },
    { id: 'volume', label: 'Vol', width: 'w-14', align: 'right' },
    { id: 'time', label: 'Time', width: 'w-16', align: 'right' },
];

// Default strict columns: Symbol, Bid, Ask, Spread
const DEFAULT_VISIBLE_COLUMNS: ColumnId[] = ['symbol', 'bid', 'ask', 'spread'];

export const MarketWatchPanel: React.FC<MarketWatchPanelProps> = ({
    ticks,
    allSymbols,
    selectedSymbol,
    onSymbolSelect,
    className
}) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [visibleColumns, setVisibleColumns] = useState<ColumnId[]>(() => {
        const saved = localStorage.getItem('rtx5_marketwatch_cols');
        return saved ? JSON.parse(saved) : DEFAULT_VISIBLE_COLUMNS;
    });
    const [activeTab, setActiveTab] = useState<TabId>('symbols');

    // Sort State
    const [sortBy, setSortBy] = useState<'symbol' | 'gainers' | 'losers' | 'volume' | null>(null);

    // Context Menu State
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; symbol?: string } | null>(null);
    const [activeSubMenu, setActiveSubMenu] = useState<string | null>(null);
    const menuRef = useRef<HTMLDivElement>(null);

    // Symbol Search & Subscribe State
    const [availableSymbols, setAvailableSymbols] = useState<AvailableSymbol[]>([]);
    const [showSearchDropdown, setShowSearchDropdown] = useState(false);
    const [subscribedSymbols, setSubscribedSymbols] = useState<string[]>([]);
    const [isSubscribing, setIsSubscribing] = useState<string | null>(null);
    const searchRef = useRef<HTMLDivElement>(null);

    // Fetch available symbols from API on mount
    useEffect(() => {
        const fetchAvailableSymbols = async () => {
            try {
                const response = await fetch(`${API_BASE}/api/symbols/available`);
                if (response.ok) {
                    const data = await response.json();
                    setAvailableSymbols(data);
                }
            } catch (error) {
                console.error('Failed to fetch available symbols:', error);
            }
        };
        fetchAvailableSymbols();
    }, []);

    // Fetch subscribed symbols
    useEffect(() => {
        const fetchSubscribed = async () => {
            try {
                const response = await fetch(`${API_BASE}/api/symbols/subscribed`);
                if (response.ok) {
                    const data = await response.json();
                    setSubscribedSymbols(data || []);
                }
            } catch (error) {
                console.error('Failed to fetch subscribed symbols:', error);
            }
        };
        fetchSubscribed();
        // Refresh every 5 seconds
        const interval = setInterval(fetchSubscribed, 5000);
        return () => clearInterval(interval);
    }, []);

    // Subscribe to a symbol
    const subscribeToSymbol = useCallback(async (symbol: string) => {
        setIsSubscribing(symbol);
        try {
            const response = await fetch(`${API_BASE}/api/symbols/subscribe`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ symbol })
            });
            const result = await response.json();
            if (result.success) {
                setSubscribedSymbols(prev => [...new Set([...prev, symbol])]);
                setSearchTerm('');
                setShowSearchDropdown(false);
                // Select the newly subscribed symbol
                onSymbolSelect(symbol);
            } else {
                console.error('Subscribe failed:', result.error);
                alert(`Failed to subscribe to ${symbol}: ${result.error || 'Unknown error'}`);
            }
        } catch (error) {
            console.error('Subscribe error:', error);
            alert(`Failed to subscribe to ${symbol}`);
        } finally {
            setIsSubscribing(null);
        }
    }, [onSymbolSelect]);

    // Close search dropdown on click outside
    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (searchRef.current && !searchRef.current.contains(e.target as Node)) {
                setShowSearchDropdown(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    // Filter available symbols for search dropdown
    const filteredAvailableSymbols = searchTerm.length > 0
        ? availableSymbols.filter(s =>
            s.symbol.toLowerCase().includes(searchTerm.toLowerCase()) ||
            s.name.toLowerCase().includes(searchTerm.toLowerCase())
        )
        : availableSymbols;

    useEffect(() => {
        localStorage.setItem('rtx5_marketwatch_cols', JSON.stringify(visibleColumns));
    }, [visibleColumns]);

    // Close menu on click outside
    useEffect(() => {
        const handleClick = (e: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
                setContextMenu(null);
                setActiveSubMenu(null);
            }
        };
        document.addEventListener('mousedown', handleClick);
        return () => document.removeEventListener('mousedown', handleClick);
    }, []);

    const toggleColumn = (colId: ColumnId) => {
        setVisibleColumns(prev => {
            if (prev.includes(colId)) {
                return prev.filter(c => c !== colId);
            } else {
                // Insert in order of ALL_COLUMNS definition to maintain table structure
                const newCols = [...prev, colId];
                return ALL_COLUMNS.filter(c => newCols.includes(c.id)).map(c => c.id);
            }
        });
    };

    const handleContextMenu = (e: React.MouseEvent, symbol?: string) => {
        e.preventDefault();
        setContextMenu({ x: e.clientX, y: e.clientY, symbol: symbol || selectedSymbol });
        setActiveSubMenu(null);
    };

    // Context Menu Action Handlers
    const handleNewOrder = useCallback(() => {
        const symbol = contextMenu?.symbol || selectedSymbol;
        if (symbol) {
            // Dispatch event to open order dialog (App.tsx should listen)
            window.dispatchEvent(new CustomEvent('openOrderDialog', { detail: { symbol } }));
        }
        setContextMenu(null);
    }, [contextMenu?.symbol, selectedSymbol]);

    const handleQuickBuy = useCallback(async () => {
        const symbol = contextMenu?.symbol || selectedSymbol;
        const tick = ticks[symbol];
        if (!symbol || !tick) {
            alert('Please select a symbol first');
            setContextMenu(null);
            return;
        }
        try {
            const response = await fetch(`${API_BASE}/api/orders/market`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    symbol,
                    side: 'BUY',
                    quantity: 0.01, // Minimum lot
                    accountId: 'RTX-000001'
                })
            });
            const result = await response.json();
            if (result.error) {
                alert(`Buy failed: ${result.error}`);
            }
        } catch (error) {
            console.error('Quick buy error:', error);
        }
        setContextMenu(null);
    }, [contextMenu?.symbol, selectedSymbol, ticks]);

    const handleQuickSell = useCallback(async () => {
        const symbol = contextMenu?.symbol || selectedSymbol;
        const tick = ticks[symbol];
        if (!symbol || !tick) {
            alert('Please select a symbol first');
            setContextMenu(null);
            return;
        }
        try {
            const response = await fetch(`${API_BASE}/api/orders/market`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    symbol,
                    side: 'SELL',
                    quantity: 0.01, // Minimum lot
                    accountId: 'RTX-000001'
                })
            });
            const result = await response.json();
            if (result.error) {
                alert(`Sell failed: ${result.error}`);
            }
        } catch (error) {
            console.error('Quick sell error:', error);
        }
        setContextMenu(null);
    }, [contextMenu?.symbol, selectedSymbol, ticks]);

    const handleChartWindow = useCallback(() => {
        const symbol = contextMenu?.symbol || selectedSymbol;
        if (symbol) {
            onSymbolSelect(symbol);
            // Dispatch event to maximize chart
            window.dispatchEvent(new CustomEvent('openChart', { detail: { symbol } }));
        }
        setContextMenu(null);
    }, [contextMenu?.symbol, selectedSymbol, onSymbolSelect]);

    const handleHideSymbol = useCallback(() => {
        const symbol = contextMenu?.symbol;
        if (symbol) {
            // Store hidden symbols in localStorage
            const hidden = JSON.parse(localStorage.getItem('rtx5_hidden_symbols') || '[]');
            if (!hidden.includes(symbol)) {
                hidden.push(symbol);
                localStorage.setItem('rtx5_hidden_symbols', JSON.stringify(hidden));
            }
        }
        setContextMenu(null);
    }, [contextMenu?.symbol]);

    const handleShowAll = useCallback(() => {
        localStorage.setItem('rtx5_hidden_symbols', '[]');
        setContextMenu(null);
    }, []);

    // Symbol Processing
    const uniqueSymbols = Array.from(new Set([
        ...allSymbols.map(s => s.symbol || s),
        ...Object.keys(ticks)
    ]));

    let processedSymbols = uniqueSymbols.filter(s =>
        (s || '').toLowerCase().includes(searchTerm.toLowerCase())
    );

    // Sorting Logic
    if (sortBy === 'symbol') {
        processedSymbols.sort((a, b) => a.localeCompare(b));
    } else if (sortBy === 'gainers') {
        processedSymbols.sort((a, b) => (ticks[b]?.dailyChange || 0) - (ticks[a]?.dailyChange || 0));
    } else if (sortBy === 'losers') {
        processedSymbols.sort((a, b) => (ticks[a]?.dailyChange || 0) - (ticks[b]?.dailyChange || 0));
    } else if (sortBy === 'volume') {
        processedSymbols.sort((a, b) => (ticks[b]?.volume || 0) - (ticks[a]?.volume || 0));
    } else {
        // Default sort (usually alphabetical or by adding order)
        processedSymbols.sort();
    }


    return (
        <div className={`flex flex-col bg-[#1e1e1e] border-b border-zinc-700 select-none ${className}`} onContextMenu={handleContextMenu}>
            {/* Context Menu */}
            {contextMenu && (
                <div
                    ref={menuRef}
                    className="fixed z-[9999] w-64 bg-[#1e1e1e] border border-zinc-600 shadow-2xl rounded-sm py-1 text-xs text-zinc-200"
                    style={{ top: contextMenu.y, left: contextMenu.x }}
                >
                    {contextMenu?.symbol && (
                        <div className="px-3 py-1.5 text-[10px] font-bold text-emerald-400 border-b border-zinc-700 mb-1">
                            {contextMenu.symbol}
                        </div>
                    )}
                    <div className="px-3 py-1.5 text-[10px] font-bold text-zinc-500 uppercase tracking-wider pl-2">Trading Actions</div>
                    <ContextMenuItem label="New Order" shortcut="F9" action={handleNewOrder} />
                    <ContextMenuItem label="Quick Buy (0.01)" action={handleQuickBuy} />
                    <ContextMenuItem label="Quick Sell (0.01)" action={handleQuickSell} />
                    <ContextMenuItem label="Chart Window" action={handleChartWindow} />
                    <ContextMenuItem label="Tick Chart" action={() => { setActiveTab('ticks'); setContextMenu(null); }} />
                    <ContextMenuItem label="Depth of Market" shortcut="Alt+B" action={() => { alert('Depth of Market - Coming soon'); setContextMenu(null); }} />
                    <ContextMenuItem label="Popup Prices" shortcut="F10" action={() => { alert('Popup Prices - Coming soon'); setContextMenu(null); }} />
                    <MenuDivider />

                    <div className="px-3 py-1.5 text-[10px] font-bold text-zinc-500 uppercase tracking-wider pl-2">Visibility</div>
                    <ContextMenuItem label="Hide" shortcut="Delete" action={handleHideSymbol} />
                    <ContextMenuItem label="Show All" action={handleShowAll} />
                    <MenuDivider />

                    <div className="px-3 py-1.5 text-[10px] font-bold text-zinc-500 uppercase tracking-wider pl-2">Configuration</div>
                    <ContextMenuItem label="Symbols" shortcut="Ctrl+U" action={() => setContextMenu(null)} />

                    {/* Sets Submenu */}
                    <div
                        className="relative"
                        onMouseEnter={() => setActiveSubMenu('sets')}
                        onMouseLeave={() => setActiveSubMenu(null)}
                    >
                        <ContextMenuItem label="Sets" hasSubmenu active={activeSubMenu === 'sets'} />
                        {activeSubMenu === 'sets' && (
                            <div className="absolute left-full top-0 -ml-1 w-48 bg-[#1e1e1e] border border-zinc-600 shadow-xl rounded-sm py-1">
                                <ContextMenuItem label="forex.all" action={() => setContextMenu(null)} />
                                <ContextMenuItem label="forex.major" action={() => setContextMenu(null)} />
                                <ContextMenuItem label="forex.crosses" action={() => setContextMenu(null)} />
                                <MenuDivider />
                                <ContextMenuItem label="Save as..." icon={<Clock size={12} />} action={() => setContextMenu(null)} />
                                <ContextMenuItem label="Remove" hasSubmenu />
                            </div>
                        )}
                    </div>

                    {/* Sort Submenu */}
                    <div
                        className="relative"
                        onMouseEnter={() => setActiveSubMenu('sort')}
                        onMouseLeave={() => setActiveSubMenu(null)}
                    >
                        <ContextMenuItem label="Sort" hasSubmenu active={activeSubMenu === 'sort'} />
                        {activeSubMenu === 'sort' && (
                            <div className="absolute left-full top-0 -ml-1 w-48 bg-[#1e1e1e] border border-zinc-600 shadow-xl rounded-sm py-1">
                                <ContextMenuItem label="Symbol" checked={sortBy === 'symbol'} action={() => setSortBy('symbol')} />
                                <ContextMenuItem label="Gainers" checked={sortBy === 'gainers'} action={() => setSortBy('gainers')} />
                                <ContextMenuItem label="Losers" checked={sortBy === 'losers'} action={() => setSortBy('losers')} />
                                <ContextMenuItem label="Volume" checked={sortBy === 'volume'} action={() => setSortBy('volume')} />
                                <MenuDivider />
                                <ContextMenuItem label="Reset" action={() => setSortBy(null)} />
                            </div>
                        )}
                    </div>

                    <ContextMenuItem label="Export" action={() => setContextMenu(null)} />

                    <MenuDivider />
                    <div className="px-3 py-1.5 text-[10px] font-bold text-zinc-500 uppercase tracking-wider pl-2">System Options</div>
                    <ContextMenuItem label="Use System Colors" checked={true} />
                    <ContextMenuItem label="Show Milliseconds" />
                    <ContextMenuItem label="Auto Remove Expired" checked={true} />
                    <ContextMenuItem label="Auto Arrange" checked={true} />
                    <ContextMenuItem label="Grid" checked={true} />
                    <MenuDivider />

                    {/* Columns Submenu */}
                    <div
                        className="relative"
                        onMouseEnter={() => setActiveSubMenu('columns')}
                        onMouseLeave={() => setActiveSubMenu(null)}
                    >
                        <ContextMenuItem label="Columns" hasSubmenu active={activeSubMenu === 'columns'} />
                        {activeSubMenu === 'columns' && (
                            <div className="absolute left-full top-0 -ml-1 w-48 bg-[#1e1e1e] border border-zinc-600 shadow-xl rounded-sm py-1 max-h-64 overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-700">
                                {ALL_COLUMNS.map(col => (
                                    (!col.locked) && (
                                        <ContextMenuItem
                                            key={col.id}
                                            label={col.label === '!' ? 'Spread' : col.label}
                                            checked={visibleColumns.includes(col.id)}
                                            action={() => toggleColumn(col.id)}
                                            autoClose={false}
                                        />
                                    )
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            )}

            {/* Header Bar */}
            <div className="px-2 py-1 bg-[#2d3436] border-b border-zinc-700 text-xs font-bold text-zinc-400 uppercase tracking-wider flex justify-between items-center">
                <span>Market Watch: {new Date().toLocaleTimeString()}</span>
            </div>

            {/* Search (Only in Symbols view) */}
            {activeTab === 'symbols' && (
                <div className="p-1 border-b border-zinc-700 bg-[#1e1e1e]" ref={searchRef}>
                    <div className="relative">
                        <input
                            type="text"
                            placeholder="Click to add symbol..."
                            value={searchTerm}
                            onChange={(e) => {
                                setSearchTerm(e.target.value);
                                setShowSearchDropdown(true);
                            }}
                            onFocus={() => setShowSearchDropdown(true)}
                            className="w-full bg-[#2d3436] border border-zinc-600 rounded-sm px-2 py-0.5 text-xs text-zinc-300 focus:outline-none focus:border-emerald-500 placeholder:text-zinc-500"
                        />
                        <Search size={10} className="absolute right-2 top-1.5 text-zinc-500" />

                        {/* Symbol Search Dropdown */}
                        {showSearchDropdown && (
                            <div className="absolute z-50 top-full left-0 right-0 mt-1 bg-[#1e1e1e] border border-zinc-600 rounded shadow-xl max-h-64 overflow-y-auto">
                                {/* Dynamic categories from API response */}
                                {(() => {
                                    // Get unique categories from available symbols (dynamic, not hardcoded)
                                    const uniqueCategories = [...new Set(filteredAvailableSymbols.map(s => s.category))];

                                    return uniqueCategories.map(category => {
                                        const categorySymbols = filteredAvailableSymbols.filter(s => s.category === category);
                                        if (categorySymbols.length === 0) return null;

                                        // Format category label (e.g., "forex.major" -> "Forex Major")
                                        const categoryLabel = category
                                            .split('.')
                                            .map(s => s.charAt(0).toUpperCase() + s.slice(1))
                                            .join(' ');

                                        return (
                                            <div key={category}>
                                                <div className="px-2 py-1 text-[9px] font-bold text-zinc-500 uppercase tracking-wider bg-[#2d3436] sticky top-0">
                                                    {categoryLabel}
                                                </div>
                                                {categorySymbols.map(sym => {
                                                    const isSubscribed = subscribedSymbols.includes(sym.symbol) || Object.keys(ticks).includes(sym.symbol);
                                                    const isLoading = isSubscribing === sym.symbol;

                                                    return (
                                                        <div
                                                            key={sym.symbol}
                                                            className={`flex items-center justify-between px-2 py-1.5 hover:bg-[#3b82f6] hover:text-white cursor-pointer text-xs group ${isSubscribed ? 'bg-emerald-900/20' : ''}`}
                                                            onClick={() => {
                                                                if (!isSubscribed && !isLoading) {
                                                                    subscribeToSymbol(sym.symbol);
                                                                } else if (isSubscribed) {
                                                                    onSymbolSelect(sym.symbol);
                                                                    setShowSearchDropdown(false);
                                                                    setSearchTerm('');
                                                                }
                                                            }}
                                                        >
                                                            <div className="flex flex-col">
                                                                <span className="font-medium text-zinc-200">{sym.symbol}</span>
                                                                <span className="text-[9px] text-zinc-500">{sym.name}</span>
                                                            </div>
                                                            <div className="flex items-center gap-1">
                                                                {isLoading ? (
                                                                    <span className="text-[9px] text-yellow-400 animate-pulse">Adding...</span>
                                                                ) : isSubscribed ? (
                                                                    <span className="text-[9px] text-emerald-400 flex items-center gap-0.5">
                                                                        <Check size={10} /> Active
                                                                    </span>
                                                                ) : (
                                                                    <span className="text-[9px] text-blue-400 flex items-center gap-0.5 opacity-0 group-hover:opacity-100">
                                                                        <Plus size={10} /> Add
                                                                    </span>
                                                                )}
                                                            </div>
                                                        </div>
                                                    );
                                                })}
                                            </div>
                                        );
                                    });
                                })()}
                                {filteredAvailableSymbols.length === 0 && (
                                    <div className="px-3 py-4 text-center text-zinc-500 text-xs">
                                        No symbols found matching "{searchTerm}"
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </div>
            )}

            {/* Content Area Based on Tab */}
            <div className="flex-1 overflow-hidden flex flex-col relative">

                {/* 1. SYMBOLS TAB */}
                {activeTab === 'symbols' && (
                    <div className="flex-1 flex flex-col">
                        {/* Column Headers */}
                        <div className="flex px-2 py-1 bg-[#2d3436] text-[10px] font-bold text-zinc-500 border-b border-zinc-700">
                            {ALL_COLUMNS.filter(c => visibleColumns.includes(c.id)).map(col => (
                                <span key={col.id} className={`${col.width} ${col.align === 'right' ? 'text-right' : col.align === 'center' ? 'text-center' : 'text-left'} px-1`}>
                                    {col.label}
                                </span>
                            ))}
                        </div>
                        <div className="flex-1 overflow-y-auto overflow-x-hidden scrollbar-thin scrollbar-thumb-zinc-700">
                            {processedSymbols.map((symbol, idx) => (
                                <MarketWatchRow
                                    key={symbol}
                                    symbol={symbol}
                                    tick={ticks[symbol]}
                                    selected={symbol === selectedSymbol}
                                    onClick={() => onSymbolSelect(symbol)}
                                    index={idx}
                                    columns={ALL_COLUMNS.filter(c => visibleColumns.includes(c.id))}
                                    onContextMenu={handleContextMenu}
                                />
                            ))}
                        </div>
                    </div>
                )}

                {/* 2. DETAILS TAB */}
                {activeTab === 'details' && (
                    <DetailsView symbol={selectedSymbol} tick={ticks[selectedSymbol]} />
                )}

                {/* 3. TRADING TAB */}
                {activeTab === 'trading' && (
                    <div className="flex-1 overflow-y-auto p-1 scrollbar-thin scrollbar-thumb-zinc-700 grid grid-cols-1 gap-1">
                        {/* Display mini trading panels for visible symbols (limited to 10 for performance if list is long) */}
                        {processedSymbols.slice(0, 20).map(symbol => (
                            <TradingPanelRow key={symbol} symbol={symbol} tick={ticks[symbol]} />
                        ))}
                    </div>
                )}

                {/* 4. TICKS TAB */}
                {activeTab === 'ticks' && (
                    <TicksView symbol={selectedSymbol} tick={ticks[selectedSymbol]} />
                )}

            </div>

            {/* Bottom Tabs */}
            <div className="flex bg-[#2d3436] border-t border-zinc-700 p-0.5 gap-1">
                <TabButton label="Symbols" active={activeTab === 'symbols'} onClick={() => setActiveTab('symbols')} />
                <TabButton label="Details" active={activeTab === 'details'} onClick={() => setActiveTab('details')} />
                <TabButton label="Trading" active={activeTab === 'trading'} onClick={() => setActiveTab('trading')} />
                <TabButton label="Ticks" active={activeTab === 'ticks'} onClick={() => setActiveTab('ticks')} />
            </div>
        </div>
    );
};

// --- Sub-Components ---

const TabButton = ({ label, active, onClick }: { label: string, active: boolean, onClick: () => void }) => (
    <div
        onClick={onClick}
        className={`px-3 py-0.5 text-[11px] font-bold cursor-pointer rounded-sm transition-colors border-t-2 ${active
                ? 'bg-[#1e1e1e] text-zinc-200 border-emerald-500'
                : 'text-zinc-500 border-transparent hover:text-zinc-300'
            }`}
    >
        {label}
    </div>
);

const DetailsView = ({ symbol, tick }: { symbol: string, tick?: Tick }) => {
    if (!tick) return <div className="flex-1 flex items-center justify-center text-zinc-500 text-xs">select a symbol</div>;

    const Row = ({ label, value, color }: { label: string, value: string, color?: string }) => (
        <div className="flex justify-between items-center py-1 border-b border-zinc-800/50 text-xs">
            <span className="text-zinc-500 font-medium">{label}</span>
            <span className={`font-mono ${color || 'text-zinc-300'}`}>{value}</span>
        </div>
    );

    return (
        <div className="flex-1 p-3 overflow-y-auto">
            <div className="text-sm font-bold text-white mb-1">{symbol}</div>
            <div className="text-[10px] text-zinc-400 mb-4">{/* Description placeholder */}Generic Stock/Forex Pair</div>

            <div className="space-y-0.5">
                <Row label="Bid" value={tick.bid.toFixed(5)} color="text-[#f87171]" />
                <Row label="Bid High" value={(tick.high || tick.bid).toFixed(5)} color="text-[#4ade80]" />
                <Row label="Bid Low" value={(tick.low || tick.bid).toFixed(5)} color="text-[#f87171]" />
                <div className="h-2"></div>
                <Row label="Ask" value={tick.ask.toFixed(5)} color="text-[#4ade80]" />
                <Row label="Ask High" value={(tick.high || tick.ask).toFixed(5)} color="text-[#4ade80]" />
                <Row label="Ask Low" value={(tick.low || tick.ask).toFixed(5)} color="text-[#f87171]" />
                <div className="h-2"></div>
                <Row label="Open Price" value={(tick.open || tick.bid).toFixed(5)} />
                <Row label="Close Price" value={(tick.close || tick.bid).toFixed(5)} />
                <div className="h-2"></div>
                <Row label="Daily Change" value={`${(tick.dailyChange || 0) >= 0 ? '+' : ''}${(tick.dailyChange || 0).toFixed(2)}%`} color={(tick.dailyChange || 0) >= 0 ? "text-[#4ade80]" : "text-[#f87171]"} />
            </div>
        </div>
    );
};

const TradingPanelRow = ({ symbol, tick }: { symbol: string, tick?: Tick }) => {
    if (!tick) return null;
    return (
        <div className="bg-[#2d3436] rounded border border-zinc-700 p-1 flex items-center justify-between">
            <div className="flex flex-col w-1/4">
                <span className="text-zinc-100 font-bold text-xs">{symbol}</span>
                <span className="text-[9px] text-zinc-500">{new Date().toLocaleTimeString()}</span>
            </div>

            <div className="flex gap-1 flex-1 justify-end">
                {/* Sell Btn */}
                <div className="flex flex-col bg-red-900/20 border border-red-800/50 rounded px-2 py-1 w-20 cursor-pointer hover:bg-red-900/40 transition-colors group">
                    <span className="text-[9px] text-red-400 font-bold group-hover:text-red-300">SELL</span>
                    <span className="text-sm font-mono text-zinc-200">{tick.bid.toFixed(5)}</span>
                </div>
                {/* Lots */}
                <div className="flex flex-col justify-center items-center w-12">
                    <input type="text" defaultValue="0.10" className="w-10 bg-[#1e1e1e] border border-zinc-600 rounded text-center text-xs text-zinc-300 py-0.5" />
                </div>
                {/* Buy Btn */}
                <div className="flex flex-col bg-emerald-900/20 border border-emerald-800/50 rounded px-2 py-1 w-20 cursor-pointer hover:bg-emerald-900/40 transition-colors group">
                    <span className="text-[9px] text-emerald-400 font-bold group-hover:text-emerald-300">BUY</span>
                    <span className="text-sm font-mono text-zinc-200">{tick.ask.toFixed(5)}</span>
                </div>
            </div>
        </div>
    );
};

const TicksView = ({ symbol, tick }: { symbol: string, tick?: Tick }) => {
    // Mock tick history generator for visualization
    // In a real app, this would come from a history buffer prop
    const [history, setHistory] = useState<{ bid: number, ask: number, time: number }[]>([]);

    useEffect(() => {
        if (tick) {
            setHistory(prev => {
                const newH = [...prev, { bid: tick.bid, ask: tick.ask, time: Date.now() }];
                if (newH.length > 50) return newH.slice(newH.length - 50);
                return newH;
            });
        }
    }, [tick]);

    if (!tick) return <div className="flex-1 flex items-center justify-center text-zinc-500 text-xs">select a symbol</div>;

    // Simple SVG scaler
    const minP = Math.min(...history.map(h => h.bid)) * 0.9999;
    const maxP = Math.max(...history.map(h => h.ask)) * 1.0001;
    const range = maxP - minP || 0.0001;
    const width = 300; // viewbox width
    const height = 200; // viewbox height

    const getY = (p: number) => height - ((p - minP) / range) * height;
    const getX = (i: number) => (i / (50 - 1)) * width;

    const bidPath = history.map((h, i) => `${i === 0 ? 'M' : 'L'} ${getX(i)} ${getY(h.bid)}`).join(' ');
    const askPath = history.map((h, i) => `${i === 0 ? 'M' : 'L'} ${getX(i)} ${getY(h.ask)}`).join(' ');

    return (
        <div className="flex-1 flex flex-col p-2 bg-[#1e1e1e]">
            <div className="flex justify-between items-center mb-2">
                <span className="text-sm font-bold text-white">{symbol}</span>
                <span className="text-[10px] text-zinc-400">Real-time Ticks</span>
            </div>
            <div className="flex-1 border border-zinc-700/50 bg-[#121212] relative overflow-hidden rounded-sm">
                <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-full" preserveAspectRatio="none">
                    {/* Grid lines */}
                    <line x1="0" y1={height / 2} x2={width} y2={height / 2} stroke="#333" strokeDasharray="4" strokeWidth="1" />
                    <line x1="0" y1={height / 4} x2={width} y2={height / 4} stroke="#222" strokeDasharray="4" strokeWidth="1" />
                    <line x1="0" y1={height * 0.75} x2={width} y2={height * 0.75} stroke="#222" strokeDasharray="4" strokeWidth="1" />

                    {/* Paths */}
                    <path d={askPath} fill="none" stroke="#f87171" strokeWidth="1.5" />
                    <path d={bidPath} fill="none" stroke="#3b82f6" strokeWidth="1.5" />
                </svg>

                <div className="absolute top-1 right-1 text-[9px] text-red-400 font-mono">{tick.ask.toFixed(5)}</div>
                <div className="absolute bottom-1 right-1 text-[9px] text-blue-400 font-mono">{tick.bid.toFixed(5)}</div>
            </div>
        </div>
    );
};

const ContextMenuItem = ({
    label,
    shortcut,
    checked,
    hasSubmenu,
    active,
    action,
    icon,
    autoClose = true
}: {
    label: string,
    shortcut?: string,
    checked?: boolean,
    hasSubmenu?: boolean,
    active?: boolean,
    action?: () => void,
    icon?: React.ReactNode,
    autoClose?: boolean
}) => (
    <div
        className={`flex items-center px-3 py-1.5 hover:bg-[#3b82f6] hover:text-white cursor-pointer group relative ${active ? 'bg-[#3b82f6] text-white' : 'text-zinc-300'}`}
        onClick={() => {
            if (action) action();
            // Note: autoClose logic is handled in the action prop usually by setting parent state to null
        }}
    >
        <div className="w-5 flex items-center justify-center mr-1">
            {checked && <Check size={12} />}
            {icon && !checked && <span className="text-zinc-400 group-hover:text-white">{icon}</span>}
        </div>
        <span className="flex-1">{label}</span>
        {shortcut && <span className="text-[10px] text-zinc-500 group-hover:text-zinc-200 ml-4 font-mono">{shortcut}</span>}
        {hasSubmenu && <ChevronRight size={12} className="ml-2 text-zinc-500 group-hover:text-white" />}
    </div>
);

const MenuDivider = () => <div className="h-[1px] bg-zinc-700 my-1 mx-2"></div>;

function MarketWatchRow({ symbol, tick, selected, onClick, index, columns, onContextMenu }: {
    symbol: string;
    tick: Tick;
    selected: boolean;
    onClick: () => void;
    index: number;
    columns: ColumnConfig[];
    onContextMenu: (e: React.MouseEvent, symbol: string) => void;
}) {
    // Determine colors
    const bidDir = tick?.prevBid !== undefined
        ? tick.bid > tick.prevBid ? 'up' : tick.bid < tick.prevBid ? 'down' : 'none'
        : 'none';

    // Institutional colors: Soft Green (#4ade80) for UP, Soft Red (#f87171) for DOWN
    const bidColor = bidDir === 'up' ? 'text-[#4ade80]' : bidDir === 'down' ? 'text-[#f87171]' : 'text-zinc-300';
    const askColor = bidDir === 'up' ? 'text-[#4ade80]' : bidDir === 'down' ? 'text-[#f87171]' : 'text-zinc-300';

    return (
        <div
            onClick={onClick}
            onContextMenu={(e) => onContextMenu(e, symbol)}
            className={`flex items-center px-2 py-0.5 cursor-pointer text-xs font-medium border-b border-zinc-800/30 
                ${selected ? 'bg-[#2d3436] text-white' : index % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'}
                hover:bg-[#2d3436]/80 hover:text-white transition-colors
            `}
        >
            {columns.map(col => {
                let content: React.ReactNode = '-';
                let cellClass = '';

                if (tick) {
                    switch (col.id) {
                        case 'symbol':
                            content = (
                                <div className="flex items-center gap-1">
                                    <div className={`w-1.5 h-1.5 rounded-full ${bidDir === 'up' ? 'bg-[#4ade80]' : bidDir === 'down' ? 'bg-[#f87171]' : 'bg-zinc-600'}`}></div>
                                    <span className={selected ? 'text-white font-bold' : 'text-zinc-200'}>{symbol}</span>
                                </div>
                            );
                            break;
                        case 'bid':
                            content = formatPrice(tick.bid, symbol);
                            cellClass = bidColor;
                            break;
                        case 'ask':
                            content = formatPrice(tick.ask, symbol);
                            cellClass = askColor;
                            break;
                        case 'spread':
                            content = tick.spread?.toFixed(0) || '!';
                            cellClass = 'text-zinc-400 text-[10px]';
                            break;
                        case 'dailyChange':
                            const chg = tick.dailyChange || 0;
                            content = `${chg > 0 ? '+' : ''}${chg.toFixed(2)}%`;
                            cellClass = chg > 0 ? 'text-[#4ade80]' : chg < 0 ? 'text-[#f87171]' : 'text-zinc-400';
                            break;
                        case 'high': content = formatPrice(tick.high || 0, symbol); break;
                        case 'low': content = formatPrice(tick.low || 0, symbol); break;
                        case 'volume': content = tick.volume?.toLocaleString() || '-'; break;
                        case 'time': content = new Date().toLocaleTimeString('en-US', { hour12: false }); break;
                        default: content = '-';
                    }
                }

                return (
                    <div key={col.id} className={`${col.width} px-1 truncate ${col.align === 'right' ? 'text-right' : col.align === 'center' ? 'text-center' : 'text-left'} ${cellClass} font-mono`}>
                        {content}
                    </div>
                );
            })}
        </div>
    );
}

function formatPrice(price: number, symbol: string): string {
    if (!price) return '---';
    if (symbol.includes('JPY')) {
        return price.toFixed(3);
    }
    return price.toFixed(5);
}
