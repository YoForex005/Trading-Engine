/**
 * Advanced Market Watch Panel
 * Professional market overview with sortable columns, color-coded movements, and context menu
 */

import { useState, useMemo, useCallback, useRef, useEffect } from 'react';
import {
  ArrowUpDown,
  TrendingUp,
  TrendingDown,
  Star,
  Search,
  MoreVertical,
  Plus,
  BarChart3,
} from 'lucide-react';
import type { MarketWatchItem, SortConfig, ContextMenuItem, ContextMenuPosition } from '../../types/trading';
import { useAppStore } from '../../store/useAppStore';

type MarketWatchColumn = 'symbol' | 'bid' | 'ask' | 'last' | 'change' | 'changePercent' | 'volume';

export const MarketWatch = () => {
  const { ticks, selectedSymbol, setSelectedSymbol } = useAppStore();

  const [searchTerm, setSearchTerm] = useState('');
  const [sortConfig, setSortConfig] = useState<SortConfig>({ column: 'symbol', direction: 'asc' });
  const [favorites, setFavorites] = useState<Set<string>>(new Set());
  const [showOnlyFavorites, setShowOnlyFavorites] = useState(false);
  const [contextMenu, setContextMenu] = useState<{ symbol: string; position: ContextMenuPosition } | null>(null);

  const contextMenuRef = useRef<HTMLDivElement>(null);

  // Convert ticks to MarketWatchItem format
  const marketItems = useMemo(() => {
    return Object.entries(ticks).map(([symbol, tick]): MarketWatchItem => {
      const mid = (tick.bid + tick.ask) / 2;
      const change = tick.prevBid ? tick.bid - tick.prevBid : 0;
      const changePercent = tick.prevBid ? (change / tick.prevBid) * 100 : 0;

      return {
        symbol,
        bid: tick.bid,
        ask: tick.ask,
        last: mid,
        change,
        changePercent,
        volume: 0, // Would come from backend
        high24h: tick.bid, // Mock data
        low24h: tick.bid, // Mock data
        timestamp: tick.timestamp,
        direction: change > 0 ? 'up' : change < 0 ? 'down' : 'neutral',
      };
    });
  }, [ticks]);

  // Filter and sort market items
  const displayItems = useMemo(() => {
    let filtered = marketItems;

    // Apply search filter
    if (searchTerm) {
      filtered = filtered.filter((item) =>
        item.symbol.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Apply favorites filter
    if (showOnlyFavorites) {
      filtered = filtered.filter((item) => favorites.has(item.symbol));
    }

    // Apply sorting
    filtered.sort((a, b) => {
      const aValue = a[sortConfig.column as keyof MarketWatchItem];
      const bValue = b[sortConfig.column as keyof MarketWatchItem];

      if (typeof aValue === 'string' && typeof bValue === 'string') {
        return sortConfig.direction === 'asc'
          ? aValue.localeCompare(bValue)
          : bValue.localeCompare(aValue);
      }

      if (typeof aValue === 'number' && typeof bValue === 'number') {
        return sortConfig.direction === 'asc' ? aValue - bValue : bValue - aValue;
      }

      return 0;
    });

    return filtered;
  }, [marketItems, searchTerm, sortConfig, showOnlyFavorites, favorites]);

  // Handle column sort
  const handleSort = (column: MarketWatchColumn) => {
    setSortConfig((prev) => ({
      column,
      direction: prev.column === column && prev.direction === 'asc' ? 'desc' : 'asc',
    }));
  };

  // Handle favorite toggle
  const toggleFavorite = useCallback((symbol: string) => {
    setFavorites((prev) => {
      const next = new Set(prev);
      if (next.has(symbol)) {
        next.delete(symbol);
      } else {
        next.add(symbol);
      }
      return next;
    });
  }, []);

  // Handle context menu
  const handleContextMenu = useCallback((e: React.MouseEvent, symbol: string) => {
    e.preventDefault();
    setContextMenu({
      symbol,
      position: { x: e.clientX, y: e.clientY },
    });
  }, []);

  // Close context menu on click outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (contextMenuRef.current && !contextMenuRef.current.contains(e.target as Node)) {
        setContextMenu(null);
      }
    };

    if (contextMenu) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [contextMenu]);

  // Context menu items
  const getContextMenuItems = (symbol: string): ContextMenuItem[] => [
    {
      label: favorites.has(symbol) ? 'Remove from Favorites' : 'Add to Favorites',
      icon: <Star size={14} className={favorites.has(symbol) ? 'fill-yellow-400 text-yellow-400' : ''} />,
      onClick: () => {
        toggleFavorite(symbol);
        setContextMenu(null);
      },
    },
    {
      label: 'View Chart',
      icon: <BarChart3 size={14} />,
      onClick: () => {
        setSelectedSymbol(symbol);
        setContextMenu(null);
      },
    },
    {
      label: 'Quick Buy',
      icon: <TrendingUp size={14} />,
      onClick: () => {
        setSelectedSymbol(symbol);
        setContextMenu(null);
        // Trigger buy action
      },
    },
    {
      label: 'Quick Sell',
      icon: <TrendingDown size={14} />,
      onClick: () => {
        setSelectedSymbol(symbol);
        setContextMenu(null);
        // Trigger sell action
      },
    },
  ];

  return (
    <div className="flex flex-col h-full bg-zinc-900 border-r border-zinc-800">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-zinc-800 bg-zinc-900/50">
        <div className="flex items-center gap-2">
          <h3 className="text-xs font-semibold text-zinc-300 uppercase tracking-wide">Market Watch</h3>
          <span className="text-xs text-emerald-400 font-medium">{displayItems.length}</span>
        </div>
        <button
          onClick={() => setShowOnlyFavorites(!showOnlyFavorites)}
          className={`p-1.5 rounded transition-colors ${
            showOnlyFavorites ? 'bg-yellow-500/20 text-yellow-400' : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800'
          }`}
          title="Show favorites only"
        >
          <Star size={14} className={showOnlyFavorites ? 'fill-yellow-400' : ''} />
        </button>
      </div>

      {/* Search */}
      <div className="px-2 py-2 border-b border-zinc-800">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-zinc-500" />
          <input
            type="text"
            placeholder="Search symbols..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full bg-zinc-800 border border-zinc-700 rounded pl-8 pr-3 py-1.5 text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/20"
          />
        </div>
      </div>

      {/* Column Headers */}
      <div className="grid grid-cols-[auto_1fr_1fr_1fr] gap-1 px-2 py-1.5 border-b border-zinc-800 bg-zinc-900/30 text-[10px] font-medium text-zinc-500 uppercase tracking-wide sticky top-0 z-10">
        <div className="w-8"></div>
        <SortableHeader column="symbol" label="Symbol" sortConfig={sortConfig} onSort={handleSort} />
        <SortableHeader column="bid" label="Bid" sortConfig={sortConfig} onSort={handleSort} align="right" />
        <SortableHeader column="changePercent" label="Change" sortConfig={sortConfig} onSort={handleSort} align="right" />
      </div>

      {/* Symbol List */}
      <div className="flex-1 overflow-y-auto">
        {displayItems.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-zinc-600">
            <Search className="w-8 h-8 mb-2" />
            <p className="text-xs">No symbols found</p>
          </div>
        ) : (
          displayItems.map((item) => (
            <MarketWatchRow
              key={item.symbol}
              item={item}
              isSelected={item.symbol === selectedSymbol}
              isFavorite={favorites.has(item.symbol)}
              onSelect={() => setSelectedSymbol(item.symbol)}
              onToggleFavorite={() => toggleFavorite(item.symbol)}
              onContextMenu={(e) => handleContextMenu(e, item.symbol)}
            />
          ))
        )}
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <ContextMenu
          ref={contextMenuRef}
          position={contextMenu.position}
          items={getContextMenuItems(contextMenu.symbol)}
          onClose={() => setContextMenu(null)}
        />
      )}
    </div>
  );
};

// Sortable Header Component
const SortableHeader = ({
  column,
  label,
  sortConfig,
  onSort,
  align = 'left',
}: {
  column: MarketWatchColumn;
  label: string;
  sortConfig: SortConfig;
  onSort: (column: MarketWatchColumn) => void;
  align?: 'left' | 'right';
}) => {
  const isActive = sortConfig.column === column;

  return (
    <button
      onClick={() => onSort(column)}
      className={`flex items-center gap-1 hover:text-zinc-300 transition-colors ${
        align === 'right' ? 'justify-end' : ''
      }`}
    >
      <span>{label}</span>
      <ArrowUpDown
        size={10}
        className={`transition-all ${isActive ? 'text-emerald-400' : 'text-zinc-600'} ${
          isActive && sortConfig.direction === 'desc' ? 'rotate-180' : ''
        }`}
      />
    </button>
  );
};

// Market Watch Row Component
const MarketWatchRow = ({
  item,
  isSelected,
  isFavorite,
  onSelect,
  onToggleFavorite,
  onContextMenu,
}: {
  item: MarketWatchItem;
  isSelected: boolean;
  isFavorite: boolean;
  onSelect: () => void;
  onToggleFavorite: () => void;
  onContextMenu: (e: React.MouseEvent) => void;
}) => {
  const changeColor = item.change > 0 ? 'text-emerald-400' : item.change < 0 ? 'text-red-400' : 'text-zinc-500';
  const bgColor = isSelected ? 'bg-emerald-500/10 border-l-2 border-emerald-500' : 'hover:bg-zinc-800/50 border-l-2 border-transparent';

  const directionIcon = item.direction === 'up'
    ? <TrendingUp className="w-3 h-3 text-emerald-400" />
    : item.direction === 'down'
    ? <TrendingDown className="w-3 h-3 text-red-400" />
    : <div className="w-3 h-3" />;

  return (
    <div
      onClick={onSelect}
      onContextMenu={onContextMenu}
      className={`grid grid-cols-[auto_1fr_1fr_1fr] gap-1 px-2 py-1.5 cursor-pointer transition-all text-xs ${bgColor}`}
    >
      <div className="flex items-center justify-center w-8">
        <button
          onClick={(e) => {
            e.stopPropagation();
            onToggleFavorite();
          }}
          className="p-0.5 hover:bg-zinc-700 rounded transition-colors"
        >
          <Star size={12} className={isFavorite ? 'fill-yellow-400 text-yellow-400' : 'text-zinc-600'} />
        </button>
      </div>

      <div className="flex items-center gap-1 font-medium text-white">
        {directionIcon}
        <span className="truncate">{item.symbol}</span>
      </div>

      <div className="text-right font-mono text-zinc-300">
        {formatPrice(item.bid, item.symbol)}
      </div>

      <div className={`text-right font-mono font-medium ${changeColor}`}>
        {item.changePercent >= 0 ? '+' : ''}
        {item.changePercent.toFixed(2)}%
      </div>
    </div>
  );
};

// Context Menu Component
const ContextMenu = ({
  position,
  items,
  onClose,
  ref,
}: {
  position: ContextMenuPosition;
  items: ContextMenuItem[];
  onClose: () => void;
  ref: React.RefObject<HTMLDivElement>;
}) => {
  return (
    <div
      ref={ref}
      className="fixed bg-zinc-900 border border-zinc-700 rounded-lg shadow-xl py-1 z-50 min-w-[180px]"
      style={{ top: position.y, left: position.x }}
    >
      {items.map((item, index) => (
        <button
          key={index}
          onClick={item.onClick}
          disabled={item.disabled}
          className="w-full flex items-center gap-2 px-3 py-2 text-xs text-zinc-300 hover:bg-zinc-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {item.icon && <span className="flex-shrink-0">{item.icon}</span>}
          <span className="flex-1 text-left">{item.label}</span>
          {item.shortcut && <span className="text-zinc-500 text-[10px]">{item.shortcut}</span>}
        </button>
      ))}
    </div>
  );
};

// Utility function
const formatPrice = (price: number, symbol: string): string => {
  if (symbol.includes('JPY')) {
    return price.toFixed(3);
  }
  return price.toFixed(5);
};
