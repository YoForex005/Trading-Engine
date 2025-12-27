import React, { useState, useMemo, useEffect } from 'react';
import { Search, Filter, Star, MoreHorizontal } from 'lucide-react';
import { MarketWatchRow } from './MarketWatchRow';
import { getSymbolCategory } from './utils';
import { Check, X } from 'lucide-react'; // For Context Menu icons if needed

interface Tick {
    symbol: string;
    bid: number;
    ask: number;
    spread?: number;
    timestamp: number;
    prevBid?: number;
    lp?: string;
}

interface MarketWatchProps {
    ticks: Record<string, Tick>;
    tickHistory: Record<string, number[]>;
    selectedSymbol: string;
    onSelectSymbol: (symbol: string) => void;
    onNewOrder: (symbol: string) => void;
}

export function MarketWatch({ ticks, tickHistory, selectedSymbol, onSelectSymbol, onNewOrder }: MarketWatchProps) {
    const [search, setSearch] = useState('');
    const [activeTab, setActiveTab] = useState<'All' | 'Forex' | 'Crypto' | 'Metals' | 'Indices' | 'Fav'>('All');
    const [favorites, setFavorites] = useState<string[]>(() => {
        const saved = localStorage.getItem('favorites');
        return saved ? JSON.parse(saved) : [];
    });

    // Context Menu State
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; symbol: string } | null>(null);

    useEffect(() => {
        localStorage.setItem('favorites', JSON.stringify(favorites));
    }, [favorites]);

    const toggleFavorite = (symbol: string) => {
        setFavorites(prev =>
            prev.includes(symbol) ? prev.filter(s => s !== symbol) : [...prev, symbol]
        );
    };

    // Close context menu on click elsewhere
    useEffect(() => {
        const handleClick = () => setContextMenu(null);
        window.addEventListener('click', handleClick);
        return () => window.removeEventListener('click', handleClick);
    }, []);

    const handleContextMenu = (e: React.MouseEvent, symbol: string) => {
        e.preventDefault();
        setContextMenu({ x: e.clientX, y: e.clientY, symbol });
    };

    // Filter and Sort Symbols
    const filteredSymbols = useMemo(() => {
        let symbols = Object.keys(ticks);

        // Sort alphabetically first
        symbols.sort();

        // Filter by Tab
        if (activeTab === 'Fav') {
            symbols = symbols.filter(s => favorites.includes(s));
        } else if (activeTab !== 'All') {
            symbols = symbols.filter(s => getSymbolCategory(s) === activeTab);
        }

        // Filter by Search
        if (search) {
            symbols = symbols.filter(s => s.toLowerCase().includes(search.toLowerCase()));
        }

        return symbols;
    }, [ticks, activeTab, search, favorites]);

    return (
        <div className="flex flex-col h-full bg-[#131722] border-r border-zinc-800">
            {/* Header & Search */}
            <div className="p-2 border-b border-zinc-800 space-y-2">
                <div className="relative">
                    <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-zinc-500" />
                    <input
                        type="text"
                        placeholder="Search..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        className="w-full bg-zinc-900 border border-zinc-800 rounded px-8 py-1 text-xs focus:outline-none focus:border-emerald-500 transition-colors"
                    />
                </div>

                {/* Tabs */}
                <div className="flex gap-1 overflow-x-auto no-scrollbar mask-fade-right pb-1">
                    {['All', 'Forex', 'Crypto', 'Metals', 'Indices', 'Fav'].map(tab => (
                        <button
                            key={tab}
                            onClick={() => setActiveTab(tab as any)}
                            className={`px-2 py-1 rounded text-[10px] font-medium whitespace-nowrap transition-colors ${activeTab === tab ? 'bg-emerald-500/20 text-emerald-400' : 'text-zinc-500 hover:text-zinc-300'
                                }`}
                        >
                            {tab === 'Fav' ? <Star size={10} fill={activeTab === 'Fav' ? 'currentColor' : 'none'} /> : tab}
                        </button>
                    ))}
                </div>
            </div>

            {/* List */}
            <div className="flex-1 overflow-y-auto">
                {filteredSymbols.length === 0 ? (
                    <div className="p-4 text-center text-xs text-zinc-600">No symbols found</div>
                ) : (
                    filteredSymbols.map(symbol => (
                        <MarketWatchRow
                            key={symbol}
                            symbol={symbol}
                            tick={ticks[symbol]}
                            history={tickHistory[symbol] || []}
                            selected={selectedSymbol === symbol}
                            onClick={() => onSelectSymbol(symbol)}
                            onContextMenu={(e) => handleContextMenu(e, symbol)}
                        />
                    ))
                )}
            </div>

            {/* Context Menu */}
            {contextMenu && (
                <div
                    className="fixed z-50 bg-zinc-800 border border-zinc-700 rounded-lg shadow-xl py-1 w-40 text-xs text-zinc-300"
                    style={{ top: contextMenu.y, left: contextMenu.x }}
                >
                    <button
                        className="w-full text-left px-3 py-2 hover:bg-zinc-700 flex items-center gap-2"
                        onClick={() => onNewOrder(contextMenu.symbol)}
                    >
                        New Order
                    </button>
                    <button
                        className="w-full text-left px-3 py-2 hover:bg-zinc-700 flex items-center gap-2"
                        onClick={() => onSelectSymbol(contextMenu.symbol)}
                    >
                        Chart Window
                    </button>
                    <div className="h-px bg-zinc-700 my-1" />
                    <button
                        className="w-full text-left px-3 py-2 hover:bg-zinc-700 flex items-center gap-2"
                        onClick={() => toggleFavorite(contextMenu.symbol)}
                    >
                        {favorites.includes(contextMenu.symbol) ? 'Remove from Favorites' : 'Add to Favorites'}
                    </button>
                    <button className="w-full text-left px-3 py-2 hover:bg-zinc-700 text-zinc-500 cursor-not-allowed">
                        Specification
                    </button>
                    <button className="w-full text-left px-3 py-2 hover:bg-zinc-700 text-zinc-500 cursor-not-allowed">
                        Depth of Market
                    </button>
                </div>
            )}
        </div>
    );
}
