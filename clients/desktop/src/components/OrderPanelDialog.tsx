import { useState } from 'react';
import { X } from 'lucide-react';

interface OrderPanelProps {
  isOpen: boolean;
  onClose: () => void;
  symbol: string;
  currentPrice: { bid: number; ask: number };
  onSubmitOrder: (order: { type: 'buy' | 'sell'; volume: number; sl?: number; tp?: number }) => void;
}

export function OrderPanelDialog({
  isOpen,
  onClose,
  symbol,
  currentPrice,
  onSubmitOrder
}: OrderPanelProps) {
  const [volume, setVolume] = useState(() => {
    // Remember last lot size from localStorage
    return parseFloat(localStorage.getItem('lastLotSize') || '0.01');
  });
  const [sl, setSl] = useState<number | undefined>();
  const [tp, setTp] = useState<number | undefined>();

  const handleBuy = () => {
    localStorage.setItem('lastLotSize', volume.toString());
    onSubmitOrder({ type: 'buy', volume, sl, tp });
    onClose();
  };

  const handleSell = () => {
    localStorage.setItem('lastLotSize', volume.toString());
    onSubmitOrder({ type: 'sell', volume, sl, tp });
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg p-6 w-96 shadow-lg">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-bold text-zinc-100">New Order</h2>
          <button
            onClick={onClose}
            className="text-zinc-500 hover:text-zinc-300 transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Symbol Info */}
        <div className="mb-4 p-3 bg-[#252525] rounded border border-zinc-700">
          <div className="text-sm text-zinc-400 mb-1">Symbol</div>
          <div className="text-xl font-bold text-emerald-400">{symbol}</div>
          <div className="grid grid-cols-2 gap-2 mt-2 text-xs">
            <div>
              <span className="text-zinc-500">Bid:</span>
              <div className="font-mono text-red-400">{currentPrice.bid.toFixed(5)}</div>
            </div>
            <div>
              <span className="text-zinc-500">Ask:</span>
              <div className="font-mono text-blue-400">{currentPrice.ask.toFixed(5)}</div>
            </div>
          </div>
        </div>

        {/* Volume Input */}
        <div className="mb-4">
          <label className="text-sm text-zinc-300 block mb-2">Volume (Lots)</label>
          <input
            type="number"
            value={volume}
            onChange={(e) => setVolume(parseFloat(e.target.value) || 0.01)}
            step="0.01"
            min="0.01"
            className="w-full bg-[#252525] border border-zinc-700 rounded px-3 py-2 text-zinc-100 font-mono text-sm focus:outline-none focus:border-blue-500 transition-colors"
          />
          <div className="text-xs text-zinc-500 mt-1">
            Quick fills:
            <button
              onClick={() => setVolume(0.01)}
              className="ml-2 px-2 py-1 bg-zinc-700/50 hover:bg-zinc-600 rounded text-zinc-300 text-xs"
            >
              0.01
            </button>
            <button
              onClick={() => setVolume(0.1)}
              className="ml-1 px-2 py-1 bg-zinc-700/50 hover:bg-zinc-600 rounded text-zinc-300 text-xs"
            >
              0.1
            </button>
            <button
              onClick={() => setVolume(1)}
              className="ml-1 px-2 py-1 bg-zinc-700/50 hover:bg-zinc-600 rounded text-zinc-300 text-xs"
            >
              1.0
            </button>
          </div>
        </div>

        {/* Stop Loss Input */}
        <div className="mb-4">
          <label className="text-sm text-zinc-300 block mb-2">Stop Loss (SL)</label>
          <input
            type="number"
            value={sl ?? ''}
            onChange={(e) => setSl(e.target.value ? parseFloat(e.target.value) : undefined)}
            placeholder="Optional"
            step="0.00001"
            className="w-full bg-[#252525] border border-zinc-700 rounded px-3 py-2 text-zinc-100 font-mono text-sm focus:outline-none focus:border-blue-500 transition-colors"
          />
        </div>

        {/* Take Profit Input */}
        <div className="mb-6">
          <label className="text-sm text-zinc-300 block mb-2">Take Profit (TP)</label>
          <input
            type="number"
            value={tp ?? ''}
            onChange={(e) => setTp(e.target.value ? parseFloat(e.target.value) : undefined)}
            placeholder="Optional"
            step="0.00001"
            className="w-full bg-[#252525] border border-zinc-700 rounded px-3 py-2 text-zinc-100 font-mono text-sm focus:outline-none focus:border-blue-500 transition-colors"
          />
        </div>

        {/* Action Buttons */}
        <div className="flex gap-3">
          <button
            onClick={handleSell}
            className="flex-1 bg-red-600 hover:bg-red-700 text-white font-bold py-2.5 px-4 rounded transition-colors flex items-center justify-center gap-2"
          >
            <span>SELL</span>
            <span className="text-xs font-mono">{currentPrice.bid.toFixed(5)}</span>
          </button>
          <button
            onClick={handleBuy}
            className="flex-1 bg-blue-600 hover:bg-blue-700 text-white font-bold py-2.5 px-4 rounded transition-colors flex items-center justify-center gap-2"
          >
            <span>BUY</span>
            <span className="text-xs font-mono">{currentPrice.ask.toFixed(5)}</span>
          </button>
        </div>

        {/* Cancel Button */}
        <button
          onClick={onClose}
          className="w-full mt-3 bg-zinc-700/30 hover:bg-zinc-700/50 text-zinc-300 font-medium py-2 px-4 rounded transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  );
}
