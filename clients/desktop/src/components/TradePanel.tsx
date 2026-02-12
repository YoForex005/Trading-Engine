/**
 * Trade Panel Component (MT5-style)
 * Comprehensive trade management with tabs: Trade, Exposure, History
 * Features: Real-time P&L, context menus, exposure visualization, history filters
 */

import { useState, useMemo, useEffect } from 'react';
import {
  X,
  TrendingUp,
  TrendingDown,
  Edit2,
  RotateCw,
  ArrowLeftRight,
  Download,
  Calendar,
  ChevronDown,
  MoreVertical
} from 'lucide-react';
import { useAppStore } from '../store/useAppStore';
import type { Position, Trade, Account } from '../store/useAppStore';
import { API_ENDPOINTS } from '../config/api';

interface TradePanelProps {
  onClose?: () => void;
  wsConnection?: WebSocket | null;
}

type TabType = 'Trade' | 'Exposure' | 'History';

interface PositionContextMenu {
  visible: boolean;
  x: number;
  y: number;
  position: Position | null;
}

export const TradePanel: React.FC<TradePanelProps> = ({ onClose, wsConnection }) => {
  const { positions, trades, account, ticks, accountId } = useAppStore();
  const [activeTab, setActiveTab] = useState<TabType>('Trade');
  const [contextMenu, setContextMenu] = useState<PositionContextMenu>({
    visible: false,
    x: 0,
    y: 0,
    position: null,
  });
  const [modifyDialogOpen, setModifyDialogOpen] = useState(false);
  const [modifyingPosition, setModifyingPosition] = useState<Position | null>(null);
  const [newSL, setNewSL] = useState('');
  const [newTP, setNewTP] = useState('');
  const [historyDateFrom, setHistoryDateFrom] = useState('');
  const [historyDateTo, setHistoryDateTo] = useState('');
  const [closingPositionId, setClosingPositionId] = useState<number | null>(null);

  // Calculate live P&L for each position
  const enrichedPositions = useMemo(() => {
    return positions.map((pos) => {
      const tick = ticks[pos.symbol];
      if (!tick) return { ...pos, livePrice: pos.currentPrice, livePnL: pos.unrealizedPnL };

      const livePrice = pos.side === 'BUY' ? tick.bid : tick.ask;
      const priceDiff = pos.side === 'BUY'
        ? livePrice - pos.openPrice
        : pos.openPrice - livePrice;

      const pipValue = 10;
      const livePnL = priceDiff * pos.volume * 100000 / pipValue - (pos.commission + pos.swap);

      return { ...pos, livePrice, livePnL };
    });
  }, [positions, ticks]);

  // Calculate totals
  const totals = useMemo(() => {
    const totalPnL = enrichedPositions.reduce((sum, pos) => sum + pos.livePnL, 0);
    const totalVolume = enrichedPositions.reduce((sum, pos) => sum + pos.volume, 0);
    const totalSwap = enrichedPositions.reduce((sum, pos) => sum + pos.swap, 0);
    const totalCommission = enrichedPositions.reduce((sum, pos) => sum + pos.commission, 0);

    return { totalPnL, totalVolume, totalSwap, totalCommission };
  }, [enrichedPositions]);

  // Calculate exposure by symbol
  const exposure = useMemo(() => {
    const symbolMap = new Map<string, { buy: number; sell: number; net: number }>();

    positions.forEach((pos) => {
      const existing = symbolMap.get(pos.symbol) || { buy: 0, sell: 0, net: 0 };
      if (pos.side === 'BUY') {
        existing.buy += pos.volume;
      } else {
        existing.sell += pos.volume;
      }
      existing.net = existing.buy - existing.sell;
      symbolMap.set(pos.symbol, existing);
    });

    return Array.from(symbolMap.entries()).map(([symbol, data]) => ({
      symbol,
      ...data,
    }));
  }, [positions]);

  // Filter history
  const filteredHistory = useMemo(() => {
    let filtered = trades;

    if (historyDateFrom) {
      filtered = filtered.filter(
        (t) => new Date(t.closeTime || t.openTime) >= new Date(historyDateFrom)
      );
    }
    if (historyDateTo) {
      filtered = filtered.filter(
        (t) => new Date(t.closeTime || t.openTime) <= new Date(historyDateTo)
      );
    }

    return filtered.sort((a, b) => {
      const dateA = new Date(b.closeTime || b.openTime).getTime();
      const dateB = new Date(a.closeTime || a.openTime).getTime();
      return dateA - dateB;
    });
  }, [trades, historyDateFrom, historyDateTo]);

  // Handle right-click on position
  const handlePositionContextMenu = (e: React.MouseEvent, position: Position) => {
    e.preventDefault();
    setContextMenu({
      visible: true,
      x: e.clientX,
      y: e.clientY,
      position,
    });
  };

  // Close position
  const handleClosePosition = async (positionId: number) => {
    if (!accountId) return;
    if (!confirm('Close this position?')) return;

    setClosingPositionId(positionId);
    try {
      const response = await fetch(API_ENDPOINTS.closePosition, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ accountId, positionId }),
      });

      if (!response.ok) {
        const error = await response.text();
        alert(`Failed to close position: ${error}`);
      }
    } catch (error) {
      console.error('Close position error:', error);
      alert('Failed to close position');
    } finally {
      setClosingPositionId(null);
      setContextMenu({ visible: false, x: 0, y: 0, position: null });
    }
  };

  // Modify position SL/TP
  const handleModifyPosition = async () => {
    if (!accountId || !modifyingPosition) return;

    const sl = newSL ? parseFloat(newSL) : modifyingPosition.sl;
    const tp = newTP ? parseFloat(newTP) : modifyingPosition.tp;

    try {
      const response = await fetch(API_ENDPOINTS.modifyOrder, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          accountId,
          positionId: modifyingPosition.id,
          sl,
          tp,
        }),
      });

      if (!response.ok) {
        const error = await response.text();
        alert(`Failed to modify position: ${error}`);
      } else {
        setModifyDialogOpen(false);
        setModifyingPosition(null);
        setNewSL('');
        setNewTP('');
      }
    } catch (error) {
      console.error('Modify position error:', error);
      alert('Failed to modify position');
    }
  };

  // Export history to CSV
  const handleExportCSV = () => {
    const headers = ['Date', 'Symbol', 'Type', 'Volume', 'Open Price', 'Close Price', 'Profit', 'Commission', 'Swap'];
    const rows = filteredHistory.map((t) => [
      t.closeTime || t.openTime,
      t.symbol,
      `${t.side} ${t.type}`,
      t.volume,
      t.openPrice,
      t.closePrice || '',
      t.profit,
      t.commission,
      t.swap,
    ]);

    const csv = [headers, ...rows].map((row) => row.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `trade-history-${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  // Close context menu when clicking outside
  useEffect(() => {
    const handleClickOutside = () => {
      if (contextMenu.visible) {
        setContextMenu({ visible: false, x: 0, y: 0, position: null });
      }
    };
    window.addEventListener('click', handleClickOutside);
    return () => window.removeEventListener('click', handleClickOutside);
  }, [contextMenu.visible]);

  return (
    <div className="flex flex-col h-full bg-[#1e1e1e] text-zinc-300">
      {/* Header with Tabs */}
      <div className="flex items-center justify-between border-b border-zinc-700 bg-[#252526]">
        <div className="flex">
          {(['Trade', 'Exposure', 'History'] as TabType[]).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-2 text-xs font-medium transition-colors border-b-2 ${
                activeTab === tab
                  ? 'text-white border-blue-500 bg-[#1e1e1e]'
                  : 'text-zinc-400 border-transparent hover:text-white hover:bg-zinc-800/50'
              }`}
            >
              {tab}
            </button>
          ))}
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="p-2 mr-2 text-zinc-400 hover:text-white hover:bg-zinc-700/50 rounded transition-colors"
          >
            <X size={16} />
          </button>
        )}
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === 'Trade' && (
          <TradeTab
            positions={enrichedPositions}
            totals={totals}
            account={account}
            onContextMenu={handlePositionContextMenu}
            closingPositionId={closingPositionId}
            onDoubleClick={(pos) => {
              setModifyingPosition(pos);
              setNewSL(pos.sl?.toString() || '');
              setNewTP(pos.tp?.toString() || '');
              setModifyDialogOpen(true);
            }}
          />
        )}

        {activeTab === 'Exposure' && <ExposureTab exposure={exposure} />}

        {activeTab === 'History' && (
          <HistoryTab
            history={filteredHistory}
            dateFrom={historyDateFrom}
            dateTo={historyDateTo}
            onDateFromChange={setHistoryDateFrom}
            onDateToChange={setHistoryDateTo}
            onExport={handleExportCSV}
          />
        )}
      </div>

      {/* Context Menu */}
      {contextMenu.visible && contextMenu.position && (
        <PositionContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          position={contextMenu.position}
          onClose={() => setContextMenu({ visible: false, x: 0, y: 0, position: null })}
          onClosePosition={() => handleClosePosition(contextMenu.position!.id)}
          onModify={() => {
            setModifyingPosition(contextMenu.position);
            setNewSL(contextMenu.position!.sl?.toString() || '');
            setNewTP(contextMenu.position!.tp?.toString() || '');
            setModifyDialogOpen(true);
            setContextMenu({ visible: false, x: 0, y: 0, position: null });
          }}
        />
      )}

      {/* Modify Dialog */}
      {modifyDialogOpen && modifyingPosition && (
        <ModifyDialog
          position={modifyingPosition}
          newSL={newSL}
          newTP={newTP}
          onSLChange={setNewSL}
          onTPChange={setNewTP}
          onCancel={() => {
            setModifyDialogOpen(false);
            setModifyingPosition(null);
            setNewSL('');
            setNewTP('');
          }}
          onConfirm={handleModifyPosition}
        />
      )}
    </div>
  );
};

// ============================================================================
// Sub-Components
// ============================================================================

interface TradeTabProps {
  positions: Array<Position & { livePrice: number; livePnL: number }>;
  totals: { totalPnL: number; totalVolume: number; totalSwap: number; totalCommission: number };
  account: Account | null;
  onContextMenu: (e: React.MouseEvent, position: Position) => void;
  closingPositionId: number | null;
  onDoubleClick: (position: Position) => void;
}

const TradeTab: React.FC<TradeTabProps> = ({
  positions,
  totals,
  account,
  onContextMenu,
  closingPositionId,
  onDoubleClick,
}) => {
  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 overflow-auto">
        <table className="w-full text-xs border-collapse">
          <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10">
            <tr>
              <th className="px-2 py-2 text-left font-medium">Symbol</th>
              <th className="px-2 py-2 text-left font-medium">Type</th>
              <th className="px-2 py-2 text-right font-medium">Volume</th>
              <th className="px-2 py-2 text-right font-medium">Open</th>
              <th className="px-2 py-2 text-right font-medium">Current</th>
              <th className="px-2 py-2 text-right font-medium">SL</th>
              <th className="px-2 py-2 text-right font-medium">TP</th>
              <th className="px-2 py-2 text-right font-medium">Swap</th>
              <th className="px-2 py-2 text-right font-medium">Profit</th>
            </tr>
          </thead>
          <tbody>
            {positions.length === 0 ? (
              <tr>
                <td colSpan={9} className="text-center py-12 text-zinc-500">
                  No open positions
                </td>
              </tr>
            ) : (
              positions.map((pos) => (
                <tr
                  key={pos.id}
                  onContextMenu={(e) => onContextMenu(e, pos)}
                  onDoubleClick={() => onDoubleClick(pos)}
                  className={`hover:bg-zinc-800/50 cursor-pointer transition-colors border-b border-zinc-800 ${
                    closingPositionId === pos.id ? 'opacity-50' : ''
                  }`}
                >
                  <td className="px-2 py-2 font-medium text-zinc-200">{pos.symbol}</td>
                  <td className="px-2 py-2">
                    <span
                      className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                        pos.side === 'BUY'
                          ? 'bg-blue-500/20 text-blue-400'
                          : 'bg-red-500/20 text-red-400'
                      }`}
                    >
                      {pos.side}
                    </span>
                  </td>
                  <td className="px-2 py-2 text-right font-mono">{pos.volume.toFixed(2)}</td>
                  <td className="px-2 py-2 text-right font-mono text-zinc-400">
                    {pos.openPrice.toFixed(5)}
                  </td>
                  <td className="px-2 py-2 text-right font-mono text-zinc-200">
                    {pos.livePrice.toFixed(5)}
                  </td>
                  <td className="px-2 py-2 text-right font-mono text-zinc-500">
                    {pos.sl ? pos.sl.toFixed(5) : '-'}
                  </td>
                  <td className="px-2 py-2 text-right font-mono text-zinc-500">
                    {pos.tp ? pos.tp.toFixed(5) : '-'}
                  </td>
                  <td className="px-2 py-2 text-right font-mono text-zinc-500">
                    ${pos.swap.toFixed(2)}
                  </td>
                  <td
                    className={`px-2 py-2 text-right font-mono font-semibold ${
                      pos.livePnL >= 0 ? 'text-emerald-400' : 'text-red-400'
                    }`}
                  >
                    {pos.livePnL >= 0 ? '+' : ''}${pos.livePnL.toFixed(2)}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Footer - Totals */}
      <div className="border-t border-zinc-700 bg-[#252526] px-4 py-3">
        <div className="grid grid-cols-4 gap-4 text-xs">
          <div>
            <div className="text-zinc-500 mb-1">Total P&L</div>
            <div
              className={`text-lg font-bold font-mono ${
                totals.totalPnL >= 0 ? 'text-emerald-400' : 'text-red-400'
              }`}
            >
              {totals.totalPnL >= 0 ? '+' : ''}${totals.totalPnL.toFixed(2)}
            </div>
          </div>
          <div>
            <div className="text-zinc-500 mb-1">Equity</div>
            <div className="text-lg font-bold font-mono text-zinc-200">
              ${account?.equity.toFixed(2) || '0.00'}
            </div>
          </div>
          <div>
            <div className="text-zinc-500 mb-1">Margin</div>
            <div className="text-sm font-mono text-zinc-300">
              ${account?.margin.toFixed(2) || '0.00'}
            </div>
          </div>
          <div>
            <div className="text-zinc-500 mb-1">Free Margin</div>
            <div className="text-sm font-mono text-zinc-300">
              ${account?.freeMargin.toFixed(2) || '0.00'}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

interface ExposureTabProps {
  exposure: Array<{ symbol: string; buy: number; sell: number; net: number }>;
}

const ExposureTab: React.FC<ExposureTabProps> = ({ exposure }) => {
  const maxVolume = Math.max(...exposure.map((e) => Math.max(e.buy, e.sell)), 1);

  return (
    <div className="flex flex-col h-full overflow-auto p-4">
      {exposure.length === 0 ? (
        <div className="flex items-center justify-center h-full text-zinc-500 text-sm">
          No exposure data available
        </div>
      ) : (
        <div className="space-y-4">
          {exposure.map((exp) => (
            <div key={exp.symbol} className="bg-[#252526] rounded-lg p-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-zinc-200">{exp.symbol}</h3>
                <div className="text-xs text-zinc-400">
                  Net:{' '}
                  <span
                    className={`font-mono font-semibold ${
                      exp.net > 0 ? 'text-blue-400' : exp.net < 0 ? 'text-red-400' : 'text-zinc-400'
                    }`}
                  >
                    {exp.net > 0 ? '+' : ''}
                    {exp.net.toFixed(2)}
                  </span>
                </div>
              </div>

              {/* Buy/Sell Bars */}
              <div className="space-y-2">
                <div>
                  <div className="flex justify-between text-xs text-zinc-500 mb-1">
                    <span>Buy</span>
                    <span className="font-mono">{exp.buy.toFixed(2)}</span>
                  </div>
                  <div className="w-full bg-zinc-800 rounded-full h-2 overflow-hidden">
                    <div
                      className="bg-blue-500 h-full rounded-full transition-all"
                      style={{ width: `${(exp.buy / maxVolume) * 100}%` }}
                    />
                  </div>
                </div>

                <div>
                  <div className="flex justify-between text-xs text-zinc-500 mb-1">
                    <span>Sell</span>
                    <span className="font-mono">{exp.sell.toFixed(2)}</span>
                  </div>
                  <div className="w-full bg-zinc-800 rounded-full h-2 overflow-hidden">
                    <div
                      className="bg-red-500 h-full rounded-full transition-all"
                      style={{ width: `${(exp.sell / maxVolume) * 100}%` }}
                    />
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

interface HistoryTabProps {
  history: Trade[];
  dateFrom: string;
  dateTo: string;
  onDateFromChange: (date: string) => void;
  onDateToChange: (date: string) => void;
  onExport: () => void;
}

const HistoryTab: React.FC<HistoryTabProps> = ({
  history,
  dateFrom,
  dateTo,
  onDateFromChange,
  onDateToChange,
  onExport,
}) => {
  const totalProfit = useMemo(() => {
    return history.reduce((sum, t) => sum + t.profit, 0);
  }, [history]);

  return (
    <div className="flex flex-col h-full">
      {/* Filters */}
      <div className="flex items-center gap-3 px-4 py-3 border-b border-zinc-700 bg-[#252526]">
        <Calendar size={14} className="text-zinc-500" />
        <input
          type="date"
          value={dateFrom}
          onChange={(e) => onDateFromChange(e.target.value)}
          className="bg-[#1e1e1e] border border-zinc-600 rounded px-2 py-1 text-xs text-white focus:outline-none focus:border-blue-500"
        />
        <span className="text-zinc-600">to</span>
        <input
          type="date"
          value={dateTo}
          onChange={(e) => onDateToChange(e.target.value)}
          className="bg-[#1e1e1e] border border-zinc-600 rounded px-2 py-1 text-xs text-white focus:outline-none focus:border-blue-500"
        />
        <div className="flex-1" />
        <button
          onClick={onExport}
          className="flex items-center gap-1 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-xs rounded transition-colors"
        >
          <Download size={14} />
          Export CSV
        </button>
      </div>

      {/* History Table */}
      <div className="flex-1 overflow-auto">
        <table className="w-full text-xs border-collapse">
          <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10">
            <tr>
              <th className="px-2 py-2 text-left font-medium">Close Time</th>
              <th className="px-2 py-2 text-left font-medium">Symbol</th>
              <th className="px-2 py-2 text-left font-medium">Type</th>
              <th className="px-2 py-2 text-right font-medium">Volume</th>
              <th className="px-2 py-2 text-right font-medium">Open</th>
              <th className="px-2 py-2 text-right font-medium">Close</th>
              <th className="px-2 py-2 text-right font-medium">Commission</th>
              <th className="px-2 py-2 text-right font-medium">Swap</th>
              <th className="px-2 py-2 text-right font-medium">Profit</th>
            </tr>
          </thead>
          <tbody>
            {history.length === 0 ? (
              <tr>
                <td colSpan={9} className="text-center py-12 text-zinc-500">
                  No trade history
                </td>
              </tr>
            ) : (
              history.map((trade) => (
                <tr
                  key={trade.id}
                  className="hover:bg-zinc-800/50 transition-colors border-b border-zinc-800"
                >
                  <td className="px-2 py-2 font-mono text-zinc-400">
                    {new Date(trade.closeTime || trade.openTime).toLocaleString()}
                  </td>
                  <td className="px-2 py-2 font-medium text-zinc-200">{trade.symbol}</td>
                  <td className="px-2 py-2">
                    <span
                      className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                        trade.side === 'BUY'
                          ? 'bg-blue-500/20 text-blue-400'
                          : 'bg-red-500/20 text-red-400'
                      }`}
                    >
                      {trade.side}
                    </span>
                  </td>
                  <td className="px-2 py-2 text-right font-mono">{trade.volume.toFixed(2)}</td>
                  <td className="px-2 py-2 text-right font-mono text-zinc-400">
                    {trade.openPrice.toFixed(5)}
                  </td>
                  <td className="px-2 py-2 text-right font-mono text-zinc-400">
                    {trade.closePrice?.toFixed(5) || '-'}
                  </td>
                  <td className="px-2 py-2 text-right font-mono text-zinc-500">
                    ${trade.commission.toFixed(2)}
                  </td>
                  <td className="px-2 py-2 text-right font-mono text-zinc-500">
                    ${trade.swap.toFixed(2)}
                  </td>
                  <td
                    className={`px-2 py-2 text-right font-mono font-semibold ${
                      trade.profit >= 0 ? 'text-emerald-400' : 'text-red-400'
                    }`}
                  >
                    {trade.profit >= 0 ? '+' : ''}${trade.profit.toFixed(2)}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Footer */}
      {history.length > 0 && (
        <div className="border-t border-zinc-700 bg-[#252526] px-4 py-3">
          <div className="flex items-center justify-between">
            <div className="text-xs text-zinc-500">{history.length} closed trades</div>
            <div className="text-sm">
              <span className="text-zinc-500 mr-2">Total Profit:</span>
              <span
                className={`font-mono font-bold ${
                  totalProfit >= 0 ? 'text-emerald-400' : 'text-red-400'
                }`}
              >
                {totalProfit >= 0 ? '+' : ''}${totalProfit.toFixed(2)}
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Dialogs & Menus
// ============================================================================

interface PositionContextMenuProps {
  x: number;
  y: number;
  position: Position;
  onClose: () => void;
  onClosePosition: () => void;
  onModify: () => void;
}

const PositionContextMenu: React.FC<PositionContextMenuProps> = ({
  x,
  y,
  position,
  onClose,
  onClosePosition,
  onModify,
}) => {
  return (
    <div
      className="fixed z-[1000] bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl py-1 w-48 animate-in fade-in zoom-in-95 duration-100"
      style={{ left: x, top: y }}
      onMouseLeave={onClose}
      onClick={(e) => e.stopPropagation()}
    >
      <div
        className="flex items-center gap-3 px-3 py-2 text-xs text-zinc-300 hover:bg-[#2d2d2d] hover:text-white cursor-pointer transition-colors"
        onClick={onClosePosition}
      >
        <X size={14} className="text-zinc-500" />
        <span>Close Position</span>
      </div>
      <div
        className="flex items-center gap-3 px-3 py-2 text-xs text-zinc-300 hover:bg-[#2d2d2d] hover:text-white cursor-pointer transition-colors"
        onClick={onModify}
      >
        <Edit2 size={14} className="text-zinc-500" />
        <span>Modify SL/TP</span>
      </div>
      <div className="h-[1px] bg-zinc-800 my-1 mx-2" />
      <div className="flex items-center gap-3 px-3 py-2 text-xs text-zinc-300 hover:bg-[#2d2d2d] hover:text-white cursor-pointer transition-colors opacity-50 cursor-not-allowed">
        <RotateCw size={14} className="text-zinc-500" />
        <span>Reverse Position</span>
      </div>
      <div className="flex items-center gap-3 px-3 py-2 text-xs text-zinc-300 hover:bg-[#2d2d2d] hover:text-white cursor-pointer transition-colors opacity-50 cursor-not-allowed">
        <ArrowLeftRight size={14} className="text-zinc-500" />
        <span>Partial Close</span>
      </div>
    </div>
  );
};

interface ModifyDialogProps {
  position: Position;
  newSL: string;
  newTP: string;
  onSLChange: (value: string) => void;
  onTPChange: (value: string) => void;
  onCancel: () => void;
  onConfirm: () => void;
}

const ModifyDialog: React.FC<ModifyDialogProps> = ({
  position,
  newSL,
  newTP,
  onSLChange,
  onTPChange,
  onCancel,
  onConfirm,
}) => {
  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg shadow-2xl w-96 animate-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between px-4 py-3 border-b border-zinc-700">
          <h2 className="text-sm font-semibold text-zinc-200">Modify Position</h2>
          <button
            onClick={onCancel}
            className="text-zinc-400 hover:text-white transition-colors"
          >
            <X size={16} />
          </button>
        </div>

        <div className="p-4 space-y-4">
          <div className="text-xs text-zinc-400 space-y-1">
            <div>Symbol: <span className="text-zinc-200 font-medium">{position.symbol}</span></div>
            <div>Type: <span className="text-zinc-200 font-medium">{position.side}</span></div>
            <div>Volume: <span className="text-zinc-200 font-medium">{position.volume}</span></div>
            <div>Open Price: <span className="text-zinc-200 font-mono">{position.openPrice.toFixed(5)}</span></div>
          </div>

          <div>
            <label className="block text-xs text-zinc-400 mb-1">Stop Loss (SL)</label>
            <input
              type="number"
              step="0.00001"
              value={newSL}
              onChange={(e) => onSLChange(e.target.value)}
              placeholder={position.sl?.toFixed(5) || 'None'}
              className="w-full bg-[#252526] border border-zinc-600 rounded px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-blue-500"
            />
          </div>

          <div>
            <label className="block text-xs text-zinc-400 mb-1">Take Profit (TP)</label>
            <input
              type="number"
              step="0.00001"
              value={newTP}
              onChange={(e) => onTPChange(e.target.value)}
              placeholder={position.tp?.toFixed(5) || 'None'}
              className="w-full bg-[#252526] border border-zinc-600 rounded px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-blue-500"
            />
          </div>
        </div>

        <div className="flex items-center justify-end gap-2 px-4 py-3 border-t border-zinc-700 bg-[#252526]">
          <button
            onClick={onCancel}
            className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-white text-sm rounded transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded transition-colors"
          >
            Modify
          </button>
        </div>
      </div>
    </div>
  );
};
