/**
 * Trading Dashboard Example
 * Complete example showing how to use all trading components together
 */

import { useState } from 'react';
import {
  TradingChart,
  OrderEntry,
  PositionList,
  AccountInfoDashboard,
  OrderBook,
  TradeHistory,
  AdminPanel,
} from '../components';
import { useAppStore } from '../store/useAppStore';
import {
  LayoutDashboard,
  TrendingUp,
  Settings,
  BookOpen,
  History,
} from 'lucide-react';

type ViewMode = 'trading' | 'positions' | 'history' | 'orderbook' | 'admin';

export const TradingDashboard = () => {
  const [viewMode, setViewMode] = useState<ViewMode>('trading');
  const { selectedSymbol, ticks, account } = useAppStore();

  const currentTick = ticks[selectedSymbol];

  return (
    <div className="flex h-screen bg-[#09090b] text-zinc-300">
      {/* Sidebar */}
      <div className="w-16 border-r border-zinc-800 flex flex-col items-center py-4 gap-4 bg-zinc-900/30">
        <div className="w-10 h-10 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-lg flex items-center justify-center text-black font-bold shadow-lg shadow-emerald-500/20">
          RT
        </div>

        <nav className="flex flex-col gap-2 w-full items-center mt-4">
          <NavButton
            icon={<LayoutDashboard size={20} />}
            active={viewMode === 'trading'}
            onClick={() => setViewMode('trading')}
            label="Trading"
          />
          <NavButton
            icon={<TrendingUp size={20} />}
            active={viewMode === 'positions'}
            onClick={() => setViewMode('positions')}
            label="Positions"
          />
          <NavButton
            icon={<BookOpen size={20} />}
            active={viewMode === 'orderbook'}
            onClick={() => setViewMode('orderbook')}
            label="Order Book"
          />
          <NavButton
            icon={<History size={20} />}
            active={viewMode === 'history'}
            onClick={() => setViewMode('history')}
            label="History"
          />
          <NavButton
            icon={<Settings size={20} />}
            active={viewMode === 'admin'}
            onClick={() => setViewMode('admin')}
            label="Admin"
          />
        </nav>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Header */}
        <header className="h-14 border-b border-zinc-800 flex items-center justify-between px-4 bg-zinc-900/20">
          <div className="flex items-center gap-4">
            <h1 className="text-lg font-bold text-white">Trading Platform</h1>
            <div className="h-6 w-px bg-zinc-700"></div>
            <div className="flex items-center gap-2">
              <span className="text-sm text-zinc-500">Symbol:</span>
              <span className="text-sm font-semibold text-white">{selectedSymbol}</span>
            </div>
            {currentTick && (
              <>
                <div className="flex items-center gap-3">
                  <div className="flex flex-col">
                    <span className="text-xs text-zinc-500">Bid</span>
                    <span className="text-sm font-mono text-emerald-400">
                      {currentTick.bid.toFixed(5)}
                    </span>
                  </div>
                  <div className="flex flex-col">
                    <span className="text-xs text-zinc-500">Ask</span>
                    <span className="text-sm font-mono text-red-400">
                      {currentTick.ask.toFixed(5)}
                    </span>
                  </div>
                </div>
              </>
            )}
          </div>
          <div className="flex items-center gap-3">
            {account && (
              <div className="flex items-center gap-4 text-sm">
                <div className="flex flex-col">
                  <span className="text-xs text-zinc-500">Balance</span>
                  <span className="font-mono text-white">${account.balance.toFixed(2)}</span>
                </div>
                <div className="flex flex-col">
                  <span className="text-xs text-zinc-500">Equity</span>
                  <span
                    className={`font-mono ${
                      account.equity >= account.balance ? 'text-emerald-400' : 'text-red-400'
                    }`}
                  >
                    ${account.equity.toFixed(2)}
                  </span>
                </div>
              </div>
            )}
          </div>
        </header>

        {/* Content Area */}
        <div className="flex-1 overflow-hidden">
          {viewMode === 'trading' && <TradingView />}
          {viewMode === 'positions' && <PositionsView />}
          {viewMode === 'orderbook' && <OrderBookView />}
          {viewMode === 'history' && <HistoryView />}
          {viewMode === 'admin' && <AdminView />}
        </div>
      </div>
    </div>
  );
};

// Trading View (Chart + Order Entry)
const TradingView = () => {
  const { selectedSymbol, ticks, account } = useAppStore();
  const currentTick = ticks[selectedSymbol];

  return (
    <div className="flex h-full gap-4 p-4">
      {/* Chart */}
      <div className="flex-1 bg-[#131722] rounded-lg overflow-hidden">
        <TradingChart symbol={selectedSymbol} currentPrice={currentTick} />
      </div>

      {/* Sidebar */}
      <div className="w-80 flex flex-col gap-4">
        {/* Account Info */}
        <div className="bg-zinc-900 rounded-lg border border-zinc-800 p-4">
          <AccountInfoDashboard />
        </div>

        {/* Order Entry */}
        {account && (
          <div className="bg-zinc-900 rounded-lg border border-zinc-800">
            <OrderEntry
              symbol={selectedSymbol}
              currentBid={currentTick?.bid}
              currentAsk={currentTick?.ask}
              accountId={1}
              balance={account.balance}
            />
          </div>
        )}
      </div>
    </div>
  );
};

// Positions View
const PositionsView = () => {
  return (
    <div className="h-full p-4">
      <div className="h-full bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
        <PositionList />
      </div>
    </div>
  );
};

// Order Book View
const OrderBookView = () => {
  const { selectedSymbol, ticks } = useAppStore();
  const currentTick = ticks[selectedSymbol];

  return (
    <div className="flex h-full gap-4 p-4">
      <div className="flex-1 bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
        <OrderBook
          symbol={selectedSymbol}
          currentBid={currentTick?.bid}
          currentAsk={currentTick?.ask}
        />
      </div>
      <div className="w-96 bg-zinc-900 rounded-lg border border-zinc-800 p-4">
        <AccountInfoDashboard />
      </div>
    </div>
  );
};

// History View
const HistoryView = () => {
  const { accountId } = useAppStore();

  return (
    <div className="h-full p-4">
      <div className="h-full bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
        <TradeHistory accountId={accountId || undefined} />
      </div>
    </div>
  );
};

// Admin View
const AdminView = () => {
  return (
    <div className="h-full p-4">
      <div className="h-full bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
        <AdminPanel />
      </div>
    </div>
  );
};

// Navigation Button Component
const NavButton = ({
  icon,
  active,
  onClick,
  label,
}: {
  icon: React.ReactNode;
  active: boolean;
  onClick: () => void;
  label: string;
}) => (
  <button
    onClick={onClick}
    title={label}
    className={`p-3 rounded-lg transition-all ${
      active
        ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
        : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800/50'
    }`}
  >
    {icon}
  </button>
);
