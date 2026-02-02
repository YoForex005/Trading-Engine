import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { createPortal } from 'react-dom';
import { Plus, Clock } from 'lucide-react';
import { ContextMenu, type ContextMenuItemConfig } from '../ui/ContextMenu';
import { useContextMenu, useKeyboardShortcuts } from '../../hooks';
import { useAppStore } from '../../store/useAppStore';

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


interface MarketWatchPanelProps {
    allSymbols: any[];
    selectedSymbol: string;
    onSymbolSelect: (symbol: string) => void;
    className?: string;
}

type ColumnId = 'symbol' | 'lp' | 'bid' | 'ask' | 'spread' | 'dailyChange' | 'last' | 'high' | 'low' | 'volume' | 'time';
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
    { id: 'lp', label: 'Source', width: 'w-16', align: 'center', locked: true },
    { id: 'bid', label: 'Bid', width: 'w-16', align: 'right', locked: true },
    { id: 'ask', label: 'Ask', width: 'w-16', align: 'right', locked: true },
    { id: 'spread', label: 'Spread', width: 'w-10', align: 'center', locked: true },
    { id: 'dailyChange', label: 'Daily', width: 'w-12', align: 'right' },
    { id: 'last', label: 'Last', width: 'w-16', align: 'right' },
    { id: 'high', label: 'High', width: 'w-16', align: 'right' },
    { id: 'low', label: 'Low', width: 'w-16', align: 'right' },
    { id: 'volume', label: 'Vol', width: 'w-14', align: 'right' },
    { id: 'time', label: 'Time', width: 'w-16', align: 'right' },
];

// Default strict columns: Symbol, LP, Bid, Ask, Spread, Daily Change
const DEFAULT_VISIBLE_COLUMNS: ColumnId[] = ['symbol', 'lp', 'bid', 'ask', 'spread', 'dailyChange'];

export const MarketWatchPanel: React.FC<MarketWatchPanelProps> = ({
    allSymbols,
    selectedSymbol,
    onSymbolSelect,
    className
}) => {
    // PERFORMANCE FIX: Removed global ticks subscription to prevent massive re-renders
    // Each row now subscribes to its own symbol updates via useTick()

    const [searchTerm, setSearchTerm] = useState('');
    const [visibleColumns, setVisibleColumns] = useState<ColumnId[]>(() => {
        const saved = localStorage.getItem('rtx5_marketwatch_cols');
        return saved ? JSON.parse(saved) : DEFAULT_VISIBLE_COLUMNS;
    });
    const [activeTab, setActiveTab] = useState<TabId>('symbols');

    // Sort State
    const [sortBy, setSortBy] = useState<'symbol' | 'gainers' | 'losers' | 'volume' | null>(null);

    // Context Menu State (using new hook)
    const contextMenu = useContextMenu();

    // Symbol Search & Subscribe State
    const [availableSymbols, setAvailableSymbols] = useState<AvailableSymbol[]>([]);
    const [showSearchDropdown, setShowSearchDropdown] = useState(false);
    const [subscribedSymbols, setSubscribedSymbols] = useState<string[]>([]);
    const [isSubscribing, setIsSubscribing] = useState<string | null>(null);
    const searchRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLInputElement>(null);

    // Keyboard navigation state (Agent 1 - MT5 parity)
    const [selectedDropdownIndex, setSelectedDropdownIndex] = useState<number>(0);
    const [dropdownPos, setDropdownPos] = useState<{ top: number; left: number; width: number } | null>(null);

    // Update dropdown position
    useEffect(() => {
        if (showSearchDropdown && inputRef.current) {
            const updatePos = () => {
                if (inputRef.current) {
                    const rect = inputRef.current.getBoundingClientRect();
                    setDropdownPos({
                        top: rect.bottom + 5, // Render below the input
                        left: rect.left,
                        width: rect.width
                    });
                }
            };
            updatePos();
            window.addEventListener('resize', updatePos);
            window.addEventListener('scroll', updatePos, true); // Capture scroll
            return () => {
                window.removeEventListener('resize', updatePos);
                window.removeEventListener('scroll', updatePos, true);
            };
        }
    }, [showSearchDropdown]);

    // Hidden symbols state (for Show All / Hide to work reactively)
    const [hiddenSymbols, setHiddenSymbols] = useState<string[]>(() => {
        const saved = localStorage.getItem('rtx5_hidden_symbols');
        return saved ? JSON.parse(saved) : [];
    });

    // System Options State (persisted to localStorage)
    const [systemOptions, setSystemOptions] = useState(() => {
        const saved = localStorage.getItem('rtx5_marketwatch_options');
        return saved ? JSON.parse(saved) : {
            useSystemColors: true,
            showMilliseconds: false,
            autoRemoveExpired: true,
            autoArrange: true,
            showGrid: true,
        };
    });

    // LP Filter State (persisted to localStorage)
    const [showOnlyRealData, setShowOnlyRealData] = useState(() => {
        const saved = localStorage.getItem('rtx5_marketwatch_lp_filter');
        return saved ? JSON.parse(saved) : false;
    });

    // Persist system options
    useEffect(() => {
        localStorage.setItem('rtx5_marketwatch_options', JSON.stringify(systemOptions));
    }, [systemOptions]);

    // Persist LP filter
    useEffect(() => {
        localStorage.setItem('rtx5_marketwatch_lp_filter', JSON.stringify(showOnlyRealData));
    }, [showOnlyRealData]);

    const toggleSystemOption = (key: string) => {
        setSystemOptions((prev: Record<string, boolean>) => ({ ...prev, [key]: !prev[key] }));
    };

    // Fetch available symbols from API on mount
    useEffect(() => {
        const fetchAvailableSymbols = async () => {
            try {
                const response = await fetch(`${API_BASE}/api/symbols/available`);
                if (response.ok) {
                    const data = await response.json();
                    setAvailableSymbols(data);

                    // Check if we need to auto-populate deeply (Agent 4)
                    const saved = localStorage.getItem('rtx5_subscribed_symbols');
                    if (!saved || JSON.parse(saved).length === 0) {
                        // If no saved symbols (or empty list), subscribe to ALL available symbols to ensure full list
                        const allSymbols = data.map((s: any) => s.symbol);
                        setSubscribedSymbols(allSymbols);
                        console.log('[MarketWatch] Auto-subscribed to all available symbols:', allSymbols.length);
                    }
                }
            } catch (error) {
                console.error('Failed to fetch available symbols:', error);
            }
        };
        fetchAvailableSymbols();
    }, []);

    // Load persisted subscribed symbols from localStorage on mount
    useEffect(() => {
        const saved = localStorage.getItem('rtx5_subscribed_symbols');
        if (saved) {
            try {
                const symbols = JSON.parse(saved);
                if (symbols && symbols.length > 0) {
                    setSubscribedSymbols(symbols);
                    console.log('[MarketWatch] Loaded subscribed symbols from localStorage:', symbols);
                }
            } catch (e) {
                console.error('[MarketWatch] Failed to load subscribed symbols:', e);
            }
        }
    }, []);

    // Fetch subscribed symbols from backend and merge with localStorage
    useEffect(() => {
        const fetchSubscribed = async () => {
            try {
                const response = await fetch(`${API_BASE}/api/symbols/subscribed`);
                if (response.ok) {
                    const data = await response.json();
                    setSubscribedSymbols(prev => {
                        // Merge backend subscriptions with local storage
                        const merged = [...new Set([...prev, ...(data || [])])];
                        return merged;
                    });
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

    // Persist subscribed symbols to localStorage whenever they change
    useEffect(() => {
        if (subscribedSymbols.length > 0) {
            localStorage.setItem('rtx5_subscribed_symbols', JSON.stringify(subscribedSymbols));
            console.log('[MarketWatch] Persisted subscribed symbols:', subscribedSymbols);
        }
    }, [subscribedSymbols]);

    // Subscribe to a symbol with optimistic updates (Agent 3 fix - instant UI feedback)
    const subscribeToSymbol = useCallback(async (symbol: string) => {
        setIsSubscribing(symbol);

        // Optimistic update - add symbol immediately for instant UI feedback
        setSubscribedSymbols(prev => [...new Set([...prev, symbol])]);
        setSearchTerm('');
        setShowSearchDropdown(false);
        onSymbolSelect(symbol);

        // Auto-focus input after subscription (Agent 1 - MT5 UX)
        setTimeout(() => inputRef.current?.focus(), 150);

        try {
            const response = await fetch(`${API_BASE}/api/symbols/subscribe`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ symbol })
            });
            const result = await response.json();

            if (!result.success) {
                // Rollback on failure
                setSubscribedSymbols(prev => prev.filter(s => s !== symbol));
                console.error('Subscribe failed:', result.error);
                alert(`Failed to subscribe to ${symbol}: ${result.error || 'Unknown error'}`);
            } else {
                console.log(`[MarketWatch] Successfully subscribed to ${symbol}`);
            }
        } catch (error) {
            // Rollback on error
            setSubscribedSymbols(prev => prev.filter(s => s !== symbol));
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

    // Filter available symbols for search dropdown (moved before handleSearchKeyDown)
    const filteredAvailableSymbols = useMemo(() => {
        return searchTerm.length > 0
            ? availableSymbols.filter(s =>
                s.symbol.toLowerCase().includes(searchTerm.toLowerCase()) ||
                s.name.toLowerCase().includes(searchTerm.toLowerCase())
            )
            : availableSymbols;
    }, [searchTerm, availableSymbols]);

    // Keyboard navigation handler (Agent 1 - MT5 parity)
    const handleSearchKeyDown = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
        if (!showSearchDropdown) return;

        const maxIndex = filteredAvailableSymbols.length - 1;

        switch (e.key) {
            case 'ArrowDown':
                e.preventDefault();
                setSelectedDropdownIndex(prev => Math.min(prev + 1, maxIndex));
                break;
            case 'ArrowUp':
                e.preventDefault();
                setSelectedDropdownIndex(prev => Math.max(prev - 1, 0));
                break;
            case 'Enter':
                e.preventDefault();
                if (filteredAvailableSymbols[selectedDropdownIndex]) {
                    const selectedSymbol = filteredAvailableSymbols[selectedDropdownIndex];
                    const currentTicks = useAppStore.getState().ticks;
                    const isSubscribed = subscribedSymbols.includes(selectedSymbol.symbol) || Object.keys(currentTicks).includes(selectedSymbol.symbol);
                    if (!isSubscribed) {
                        subscribeToSymbol(selectedSymbol.symbol);
                    } else {
                        onSymbolSelect(selectedSymbol.symbol);
                        setShowSearchDropdown(false);
                        setSearchTerm('');
                    }
                } else if (searchTerm.trim().length > 0) {
                    // Agent 4: Allow forcing subscription to a symbol even if not in the Available list
                    // This handles cases where backend supports dynamic symbol creation or the list is stale
                    subscribeToSymbol(searchTerm.trim().toUpperCase());
                }
                break;
            case 'Escape':
                e.preventDefault();
                setShowSearchDropdown(false);
                setSearchTerm('');
                inputRef.current?.blur();
                break;
        }
    }, [showSearchDropdown, filteredAvailableSymbols, selectedDropdownIndex, subscribedSymbols, subscribeToSymbol, onSymbolSelect]);

    // Reset selected index when filtered list changes
    useEffect(() => {
        setSelectedDropdownIndex(0);
    }, [searchTerm]);

    useEffect(() => {
        localStorage.setItem('rtx5_marketwatch_cols', JSON.stringify(visibleColumns));
    }, [visibleColumns]);

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

    const handleContextMenuOpen = useCallback((e: React.MouseEvent, symbol?: string) => {
        e.preventDefault();
        e.stopPropagation();
        contextMenu.open(e.clientX, e.clientY, symbol || selectedSymbol);
    }, [contextMenu, selectedSymbol]);

    // Context Menu Action Handlers
    const handleNewOrder = useCallback(() => {
        const symbol = contextMenu.state.data || selectedSymbol;
        if (symbol) {
            // Dispatch event to open order dialog (App.tsx should listen)
            window.dispatchEvent(new CustomEvent('openOrderDialog', { detail: { symbol } }));
        }
        contextMenu.close();
    }, [contextMenu, selectedSymbol]);

    const handleQuickBuy = useCallback(async () => {
        const symbol = contextMenu.state.data || selectedSymbol;
        const currentTicks = useAppStore.getState().ticks;
        const tick = currentTicks[symbol];
        if (!symbol || !tick) {
            alert('Please select a symbol first');
            contextMenu.close();
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
        contextMenu.close();
    }, [contextMenu, selectedSymbol]);

    const handleQuickSell = useCallback(async () => {
        const symbol = contextMenu.state.data || selectedSymbol;
        const currentTicks = useAppStore.getState().ticks;
        const tick = currentTicks[symbol];
        if (!symbol || !tick) {
            alert('Please select a symbol first');
            contextMenu.close();
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
        contextMenu.close();
    }, [contextMenu, selectedSymbol]);

    const handleChartWindow = useCallback(() => {
        const symbol = contextMenu.state.data || selectedSymbol;
        if (symbol) {
            onSymbolSelect(symbol);
            // Dispatch event to maximize chart
            window.dispatchEvent(new CustomEvent('openChart', { detail: { symbol } }));
        }
        contextMenu.close();
    }, [contextMenu, selectedSymbol, onSymbolSelect]);

    const handleHideSymbol = useCallback(() => {
        const symbol = contextMenu.state.data;
        if (symbol) {
            setHiddenSymbols(prev => {
                if (!prev.includes(symbol)) {
                    const updated = [...prev, symbol];
                    localStorage.setItem('rtx5_hidden_symbols', JSON.stringify(updated));
                    return updated;
                }
                return prev;
            });
        }
        contextMenu.close();
    }, [contextMenu]);

    const handleShowAll = useCallback(() => {
        setHiddenSymbols([]);
        localStorage.setItem('rtx5_hidden_symbols', '[]');
        contextMenu.close();
    }, [contextMenu]);

    // Export symbols as CSV
    const handleExport = useCallback(() => {
        const currentTicks = useAppStore.getState().ticks;
        const headers = ['Symbol', 'Bid', 'Ask', 'Spread (pips)', 'Daily Change %'];
        const rows = Object.keys(currentTicks).map(sym => {
            const t = currentTicks[sym];
            const spreadInPips = Math.round((t.spread || (t.ask - t.bid)) * 10000);
            return [sym, t.bid, t.ask, spreadInPips, t.dailyChange || 0].join(',');
        });
        const csv = [headers.join(','), ...rows].join('\n');
        const blob = new Blob([csv], { type: 'text/csv' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `marketwatch_${new Date().toISOString().slice(0, 10)}.csv`;
        a.click();
        URL.revokeObjectURL(url);
        contextMenu.close();
    }, [contextMenu]);

    // Open Depth of Market modal
    const handleDepthOfMarket = useCallback(() => {
        const symbol = contextMenu.state.data || selectedSymbol;
        window.dispatchEvent(new CustomEvent('openDepthOfMarket', { detail: { symbol } }));
        contextMenu.close();
    }, [contextMenu, selectedSymbol]);

    // Open Popup Prices window
    const handlePopupPrices = useCallback(() => {
        const symbol = contextMenu.state.data || selectedSymbol;
        window.dispatchEvent(new CustomEvent('openPopupPrices', { detail: { symbol } }));
        contextMenu.close();
    }, [contextMenu, selectedSymbol]);

    // Open Symbols management dialog
    const handleOpenSymbols = useCallback(() => {
        window.dispatchEvent(new CustomEvent('openSymbolsDialog'));
        contextMenu.close();
    }, [contextMenu]);

    // Symbol Processing
    // Global keyboard shortcuts integration (MT5-style)
    useKeyboardShortcuts({
        NEW_ORDER: handleNewOrder,
        DEPTH_OF_MARKET: handleDepthOfMarket,
        SYMBOLS_DIALOG: handleOpenSymbols,
        POPUP_PRICES: handlePopupPrices,
        CLOSE_MODAL: () => contextMenu.close(),
    });

    // Convert to useMemo for reactivity (Agent 3 fix - eliminates state sync race condition)
    const processedSymbols = useMemo(() => {
        const currentTicks = useAppStore.getState().ticks;
        // Combine all potential symbol sources
        const uniqueSymbols = Array.from(new Set([
            // Agent 4 Fix: removed ...allSymbols from here to preventing showing unsubscribed 'ghost' rows
            ...Object.keys(currentTicks),
            ...subscribedSymbols
        ]));

        let result = uniqueSymbols.filter(s => {
            if (!s || s.trim().length === 0) return false;
            if (s.startsWith('-')) return false; // Filter out spacer rows
            if (hiddenSymbols.includes(s)) return false;

            // LP Filter: Show only real data (YOFX) if enabled
            if (showOnlyRealData) {
                const tick = currentTicks[s];
                if (tick && tick.lp !== 'YOFX') return false;
            }

            // CRITICAL: Filter out FIX variant symbols with suffixes and numbers
            // Reject: USDCAD1, USDCHF.H, USDCHF.h, USDCHF.c, AUDCHF!, BTCUSD.H, etc.
            // Allow: EURUSD, XAUUSD, BTCUSD, US30, JP225, GOLD.APL2025, SILVER.DEC2024

            // Pattern: Reject if symbol ends with:
            // - Digit followed by nothing (e.g., USDCAD1)
            // - Dot followed by single letter (e.g., .H, .h, .c) UNLESS it's a known futures contract
            // - Exclamation mark (e.g., AUDCHF!)

            const isFuturesContract = /\.(APL|DEC|MAR|JUN|SEP)\d{4}$/.test(s); // GOLD.APL2025, SILVER.DEC2024
            if (isFuturesContract) return true; // Allow futures contracts

            const hasInvalidSuffix = /[!]$/.test(s); // Ends with !
            const hasSingleLetterSuffix = /\.[a-zA-Z]$/.test(s); // Ends with .H, .h, .c
            const hasSlash = s.includes('/'); // Reject AUD/CHF, BTC/USD

            // Reject symbols ending with digits like '1', '2' etc ONLY if they follow a 6-character forex-like string
            // Allow: AU200, US30, JP225, GER40, etc.
            // Reject: USDCAD1, AUDCHF.H1, etc.
            const hasVariantNumberSuffix = /[A-Z]{6}\d+$/.test(s) || /[A-Z]{3,}\.[a-zA-Z]\d+$/.test(s);

            if (hasInvalidSuffix || hasSingleLetterSuffix || hasSlash || hasVariantNumberSuffix) {
                return false; // Reject variants
            }

            return true;
        });

        // Sorting Logic
        if (sortBy === 'symbol') {
            result.sort((a, b) => a.localeCompare(b));
        } else if (sortBy === 'gainers') {
            result.sort((a, b) => (currentTicks[b]?.dailyChange || 0) - (currentTicks[a]?.dailyChange || 0));
        } else if (sortBy === 'losers') {
            result.sort((a, b) => (currentTicks[a]?.dailyChange || 0) - (currentTicks[b]?.dailyChange || 0));
        } else if (sortBy === 'volume') {
            result.sort((a, b) => (currentTicks[b]?.volume || 0) - (currentTicks[a]?.volume || 0));
        } else {
            // Default sort
            result.sort();
        }

        return result;
    }, [allSymbols, subscribedSymbols, hiddenSymbols, sortBy, showOnlyRealData]);

    // Build context menu items configuration
    const menuItems: ContextMenuItemConfig[] = useMemo(() => [
        { label: 'Trading Actions', divider: true },
        { label: 'New Order', shortcut: 'F9', action: handleNewOrder },
        { label: 'Quick Buy (0.01)', action: handleQuickBuy },
        { label: 'Quick Sell (0.01)', action: handleQuickSell },
        { label: 'Chart Window', action: handleChartWindow },
        { label: 'Tick Chart', action: () => { setActiveTab('ticks'); contextMenu.close(); } },
        { label: 'Depth of Market', shortcut: 'Alt+B', action: handleDepthOfMarket },
        { label: 'Popup Prices', shortcut: 'F10', action: handlePopupPrices },
        { label: 'Separator 1', divider: true },
        { label: 'Visibility', divider: true },
        { label: 'Hide', shortcut: 'Delete', action: handleHideSymbol },
        { label: 'Show All', action: handleShowAll },
        { label: 'Separator 2', divider: true },
        { label: 'Configuration', divider: true },
        { label: 'Symbols', shortcut: 'Ctrl+U', action: handleOpenSymbols },
        {
            label: 'Sets',
            submenu: [
                { label: 'forex.all', action: () => contextMenu.close() },
                { label: 'forex.major', action: () => contextMenu.close() },
                { label: 'forex.crosses', action: () => contextMenu.close() },
                { label: 'Separator Sets', divider: true },
                { label: 'Save as...', icon: <Clock size={12} />, action: () => contextMenu.close() },
                { label: 'Remove', submenu: [] }
            ]
        },
        {
            label: 'Sort',
            submenu: [
                { label: 'Symbol', checked: sortBy === 'symbol', action: () => setSortBy('symbol') },
                { label: 'Gainers', checked: sortBy === 'gainers', action: () => setSortBy('gainers') },
                { label: 'Losers', checked: sortBy === 'losers', action: () => setSortBy('losers') },
                { label: 'Volume', checked: sortBy === 'volume', action: () => setSortBy('volume') },
                { label: 'Separator Sort', divider: true },
                { label: 'Reset', action: () => setSortBy(null) }
            ]
        },
        { label: 'Export', action: handleExport },
        { label: 'Separator 3', divider: true },
        { label: 'System Options', divider: true },
        { label: 'Use System Colors', checked: systemOptions.useSystemColors, action: () => toggleSystemOption('useSystemColors') },
        { label: 'Show Milliseconds', checked: systemOptions.showMilliseconds, action: () => toggleSystemOption('showMilliseconds') },
        { label: 'Auto Remove Expired', checked: systemOptions.autoRemoveExpired, action: () => toggleSystemOption('autoRemoveExpired') },
        { label: 'Auto Arrange', checked: systemOptions.autoArrange, action: () => toggleSystemOption('autoArrange') },
        { label: 'Grid', checked: systemOptions.showGrid, action: () => toggleSystemOption('showGrid') },
        { label: 'Separator 4', divider: true },
        {
            label: 'Columns',
            submenu: ALL_COLUMNS
                .filter(col => !col.locked)
                .map(col => ({
                    label: col.label === '!' ? 'Spread' : col.label,
                    checked: visibleColumns.includes(col.id),
                    action: () => toggleColumn(col.id),
                    autoClose: false
                }))
        }
    ], [
        sortBy, systemOptions, visibleColumns,
        handleNewOrder, handleQuickBuy, handleQuickSell, handleChartWindow,
        handleDepthOfMarket, handlePopupPrices, handleHideSymbol, handleShowAll,
        handleOpenSymbols, handleExport, contextMenu, setActiveTab, toggleColumn
    ]);

    return (
        <div className={`flex flex-col bg-[#1e1e1e] border-b border-zinc-700 select-none ${className}`} onContextMenu={handleContextMenuOpen}>
            {/* Portal-Based Context Menu */}
            {contextMenu.state.isOpen && (
                <ContextMenu
                    items={menuItems}
                    onClose={contextMenu.close}
                    position={contextMenu.state.position}
                    triggerSymbol={contextMenu.state.data}
                />
            )}

            {/* Header Bar */}
            <div className="px-2 py-1 bg-[#2d3436] border-b border-zinc-700 text-xs font-bold text-zinc-400 uppercase tracking-wider flex justify-between items-center">
                <span>Market Watch: {new Date().toLocaleTimeString()}</span>
                <button
                    onClick={() => setShowOnlyRealData(!showOnlyRealData)}
                    className={`px-2 py-0.5 text-[10px] rounded transition-colors ${
                        showOnlyRealData
                            ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/50'
                            : 'text-zinc-500 hover:text-zinc-300 border border-zinc-700 hover:border-zinc-600'
                    }`}
                    title={showOnlyRealData ? 'Showing only real data (YOFX)' : 'Showing all data (Real + Simulated)'}
                >
                    {showOnlyRealData ? '✓ Real Only' : 'All Data'}
                </button>
            </div>

            {/* Content Area Based on Tab */}
            <div className="flex-1 overflow-hidden flex flex-col relative min-h-0">
                {/* 1. SYMBOLS TAB */}
                {activeTab === 'symbols' && (
                    <div className="flex-1 flex flex-col h-full">
                        {/* Column Headers */}
                        <div className="flex px-2 py-1 bg-[#2d3436] text-[10px] font-bold text-zinc-500 border-b border-zinc-700">
                            {ALL_COLUMNS.filter(c => visibleColumns.includes(c.id)).map(col => (
                                <span key={col.id} className={`${col.width} ${col.align === 'right' ? 'text-right' : col.align === 'center' ? 'text-center' : 'text-left'} px-1`}>
                                    {col.label}
                                </span>
                            ))}
                        </div>

                        {/* Scrollable Container */}
                        <div className="flex-1 overflow-y-auto overflow-x-hidden scrollbar-thin scrollbar-thumb-zinc-500 relative">
                            {/* Click to add row - Sticky at Top (Inside Scroll Container) */}
                            <div className="sticky top-0 z-20 px-2 py-1 border-b border-zinc-700/50 bg-[#252526] hover:bg-[#2d2d2d] transition-colors min-h-[32px] flex items-center shadow-md box-border" ref={searchRef}>
                                {showSearchDropdown ? (
                                    <div className="relative w-full">
                                        <input
                                            ref={inputRef}
                                            autoFocus
                                            type="text"
                                            className="w-full bg-[#1e1e1e] border border-blue-500 text-xs text-white px-1 outline-none h-5"
                                            value={searchTerm}
                                            onChange={(e) => setSearchTerm(e.target.value)}
                                            onKeyDown={handleSearchKeyDown}
                                            disabled={!!isSubscribing}
                                            onBlur={() => {
                                                // Optional: delay close to allow click
                                            }}
                                        />
                                        {/* Dropdown via Portal */}
                                        {dropdownPos && createPortal(
                                            <div
                                                className="fixed z-[9999] bg-[#1e1e1e] border border-zinc-600 rounded shadow-xl overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-500 scrollbar-track-zinc-800"
                                                style={{
                                                    top: dropdownPos.top,
                                                    left: dropdownPos.left,
                                                    width: dropdownPos.width,
                                                    maxHeight: '256px'
                                                }}
                                            >
                                                {filteredAvailableSymbols.map((sym, idx) => (
                                                    <div
                                                        key={sym.symbol}
                                                        className={`px-2 py-1 hover:bg-[#3b82f6] hover:text-white cursor-pointer text-xs flex justify-between ${idx === selectedDropdownIndex ? 'bg-blue-900/40' : ''}`}
                                                        onClick={() => {
                                                            subscribeToSymbol(sym.symbol);
                                                            setSearchTerm('');
                                                        }}
                                                    >
                                                        <span>{sym.symbol}</span>
                                                        <span className="text-[10px] text-zinc-500">{sym.name}</span>
                                                    </div>
                                                ))}
                                                {filteredAvailableSymbols.length === 0 && (
                                                    <div className="px-2 py-1 text-zinc-500 text-xs italic">No symbols found</div>
                                                )}
                                            </div>,
                                            document.body
                                        )}
                                    </div>
                                ) : (
                                    <div
                                        className="flex items-center justify-between w-full cursor-text opacity-70 hover:opacity-100 transition-opacity"
                                        onClick={() => {
                                            setShowSearchDropdown(true);
                                            setSearchTerm('');
                                        }}
                                    >
                                        <div className="flex items-center gap-2">
                                            <Plus size={14} className="text-green-500" />
                                            <span className="text-xs text-zinc-400 font-mono italic">click to add...</span>
                                        </div>
                                        <span className="text-xs text-zinc-500 font-mono">
                                            {subscribedSymbols.length} / {availableSymbols.length || 80}
                                        </span>
                                    </div>
                                )}
                            </div>

                            {/* Symbol List */}
                            {processedSymbols.map((symbol, idx) => (
                                <MarketWatchRow
                                    key={symbol}
                                    symbol={symbol}
                                    selected={symbol === selectedSymbol}
                                    onClick={() => onSymbolSelect(symbol)}
                                    index={idx}
                                    columns={ALL_COLUMNS.filter(c => visibleColumns.includes(c.id))}
                                    onContextMenu={handleContextMenuOpen}
                                    isSubscribed={subscribedSymbols.includes(symbol)}
                                />
                            ))}
                        </div>
                    </div>
                )}

                {/* 2. DETAILS TAB */}
                {activeTab === 'details' && (
                    <DetailsView symbol={selectedSymbol} />
                )}

                {/* 3. TRADING TAB */}
                {activeTab === 'trading' && (
                    <div className="flex-1 overflow-y-auto p-1 scrollbar-thin scrollbar-thumb-zinc-700 grid grid-cols-1 gap-1">
                        {processedSymbols.slice(0, 20).map(symbol => (
                            <TradingPanelRow key={symbol} symbol={symbol} />
                        ))}
                    </div>
                )}

                {/* 4. TICKS TAB */}
                {activeTab === 'ticks' && (
                    <TicksView symbol={selectedSymbol} />
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

const DetailsView = ({ symbol }: { symbol: string }) => {
    const tick = useTick(symbol);
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

const TradingPanelRow = ({ symbol }: { symbol: string }) => {
    const tick = useTick(symbol);
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

const TicksView = ({ symbol }: { symbol: string }) => {
    const tick = useTick(symbol);
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

import { useTick } from '../../store/useAppStore';

// MarketWatchRow component with per-row subscription optimization (AGENT 4 FIX)
const MarketWatchRow = React.memo(function MarketWatchRow({ symbol, selected, onClick, index, columns, onContextMenu, isSubscribed }: {
    symbol: string;
    selected: boolean;
    onClick: () => void;
    index: number;
    columns: ColumnConfig[];
    onContextMenu: (e: React.MouseEvent, symbol: string) => void;
    isSubscribed?: boolean;
}) {
    if (!symbol) return null;

    // Hook-based subscription for this specific symbol only
    const tick = useTick(symbol);

    // Flash animation state (Agent 2 - MT5 parity)
    const [flashBid, setFlashBid] = useState<'up' | 'down' | 'none'>('none');
    const [flashAsk, setFlashAsk] = useState<'up' | 'down' | 'none'>('none');

    // Detect bid price changes and trigger flash
    useEffect(() => {
        if (tick?.prevBid !== undefined) {
            if (tick.bid > tick.prevBid) {
                setFlashBid('up');
                const timer = setTimeout(() => setFlashBid('none'), 200);
                return () => clearTimeout(timer);
            } else if (tick.bid < tick.prevBid) {
                setFlashBid('down');
                const timer = setTimeout(() => setFlashBid('none'), 200);
                return () => clearTimeout(timer);
            }
        }
    }, [tick?.bid, tick?.prevBid]);

    // Detect ask price changes and trigger flash
    useEffect(() => {
        if (tick?.ask && tick?.prevBid !== undefined) {
            const prevAsk = tick.prevBid + (tick.spread || 0);
            if (tick.ask > prevAsk) {
                setFlashAsk('up');
                const timer = setTimeout(() => setFlashAsk('none'), 200);
                return () => clearTimeout(timer);
            } else if (tick.ask < prevAsk) {
                setFlashAsk('down');
                const timer = setTimeout(() => setFlashAsk('none'), 200);
                return () => clearTimeout(timer);
            }
        }
    }, [tick?.ask, tick?.prevBid, tick?.spread]);

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
                ${selected ? 'bg-[#0078D7] text-white' : index % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#232323]'}
                hover:bg-[#2d3436]/80 hover:text-white transition-colors
            `}
        >
            {columns.map(col => {
                let content: React.ReactNode = '';
                let cellClass = '';

                if (tick) {
                    switch (col.id) {
                        case 'symbol':
                            content = (
                                <div className="flex items-center gap-1">
                                    {bidDir === 'up' ? (
                                        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4" className="text-[#4ade80]">
                                            <path d="M12 19V5M5 12l7-7 7 7" />
                                        </svg>
                                    ) : bidDir === 'down' ? (
                                        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4" className="text-[#f87171]">
                                            <path d="M12 5v14M5 12l7 7 7-7" />
                                        </svg>
                                    ) : (
                                        <div className="w-1.5 h-1.5 rounded-full bg-zinc-600"></div>
                                    )}
                                    <span className={selected ? 'text-white font-bold' : 'text-zinc-200'}>{symbol}</span>
                                </div>
                            );
                            break;
                        case 'lp':
                            const lpValue = tick.lp || 'SIM';
                            const isReal = lpValue === 'YOFX';
                            content = (
                                <div className={`flex items-center justify-center gap-1 ${isReal ? 'text-[#4ade80]' : 'text-[#f87171]'}`} title={isReal ? '✓ Real market data from YOFX broker' : '⚠️ Simulated data - not real market prices'}>
                                    {isReal ? (
                                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" className="flex-shrink-0">
                                            <path d="M20 6L9 17l-5-5" />
                                        </svg>
                                    ) : (
                                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" className="flex-shrink-0">
                                            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                                            <line x1="12" y1="9" x2="12" y2="13" />
                                            <line x1="12" y1="17" x2="12.01" y2="17" />
                                        </svg>
                                    )}
                                    <span className="text-[9px] font-bold uppercase">{lpValue}</span>
                                </div>
                            );
                            cellClass = 'cursor-help';
                            break;
                        case 'bid':
                            content = formatPrice(tick.bid, symbol);
                            cellClass = `${bidColor} transition-colors duration-200 ${flashBid === 'up' ? 'bg-emerald-500/30' :
                                flashBid === 'down' ? 'bg-red-500/30' : ''
                                }`;
                            break;
                        case 'ask':
                            content = formatPrice(tick.ask, symbol);
                            cellClass = `${askColor} transition-colors duration-200 ${flashAsk === 'up' ? 'bg-emerald-500/30' :
                                flashAsk === 'down' ? 'bg-red-500/30' : ''
                                }`;
                            break;
                        case 'spread':
                            // Always recalculate spread dynamically (Agent 4 fix - MT5 parity)
                            const rawSpread = tick.ask - tick.bid;
                            const spreadFormat = getSpreadFormat(symbol);
                            const spreadInPips = rawSpread * spreadFormat.multiplier;
                            content = spreadInPips > 0 ? spreadInPips.toFixed(spreadFormat.decimals) : '';
                            cellClass = 'text-zinc-400 text-[10px]';
                            break;
                        case 'dailyChange':
                            const chg = tick.dailyChange || 0;
                            content = `${chg > 0 ? '+' : ''}${chg.toFixed(2)}%`;
                            cellClass = chg > 0 ? 'text-[#4ade80]' : chg < 0 ? 'text-[#f87171]' : 'text-zinc-400';
                            break;
                        case 'high': content = formatPrice(tick.high || 0, symbol); break;
                        case 'low': content = formatPrice(tick.low || 0, symbol); break;
                        case 'volume': content = tick.volume?.toLocaleString() || ''; break;
                        case 'time': content = new Date().toLocaleTimeString('en-US', { hour12: false }); break;
                        default: content = '';
                    }
                } else if (isSubscribed) {
                    // PERFORMANCE FIX: Show subscription status for symbols waiting for data (MT5 parity)
                    switch (col.id) {
                        case 'symbol':
                            content = (
                                <div className="flex items-center gap-1">
                                    <Clock size={10} className="text-yellow-500 animate-pulse" />
                                    <span className={selected ? 'text-white font-bold' : 'text-zinc-200'}>{symbol}</span>
                                </div>
                            );
                            break;
                        case 'lp':
                            content = <span className="text-yellow-600 text-[9px] italic">...</span>;
                            cellClass = 'text-center';
                            break;
                        case 'bid':
                        case 'ask':
                        case 'spread':
                            content = <span className="text-yellow-600 text-[9px] italic">Waiting...</span>;
                            cellClass = 'text-center';
                            break;
                        default:
                            content = '';
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
}, (prevProps, nextProps) => {
    // Only re-render if essential UI state changed
    // Tick data is handled internally by useTick() now
    return prevProps.selected === nextProps.selected &&
        prevProps.isSubscribed === nextProps.isSubscribed;
});

function formatPrice(price: number, symbol: string): string {
    if (!price) return '---';
    if (symbol.includes('JPY')) {
        return price.toFixed(3);
    }
    return price.toFixed(5);
}

// Symbol-aware spread formatting (Agent 4 recommendation - MT5 parity fix)
function getSpreadFormat(symbol: string): { pipSize: number; multiplier: number; decimals: number } {
    // Gold and precious metals (2 decimals, pip = 0.01)
    if (symbol.includes('XAU') || symbol.includes('XAG') || symbol.includes('GOLD') || symbol.includes('SILVER')) {
        return { pipSize: 0.01, multiplier: 100, decimals: 0 };
    }

    // JPY pairs (3 decimals, pip = 0.01)
    if (symbol.includes('JPY')) {
        return { pipSize: 0.01, multiplier: 100, decimals: 1 };
    }

    // Crypto (variable decimals, pip = 0.01)
    if (symbol.includes('BTC') || symbol.includes('ETH') || symbol.includes('USDT')) {
        return { pipSize: 0.01, multiplier: 100, decimals: 1 };
    }

    // Oil and commodities
    if (symbol.includes('WTI') || symbol.includes('BRENT') || symbol.includes('OIL')) {
        return { pipSize: 0.01, multiplier: 100, decimals: 1 };
    }

    // Standard forex (5 decimals, pip = 0.0001)
    return { pipSize: 0.0001, multiplier: 10000, decimals: 1 };
}
