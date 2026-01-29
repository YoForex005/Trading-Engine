/**
 * Advanced Order Entry Panel
 * Supports Market, Limit, Stop, and Stop-Limit orders with SL/TP
 */

import { useState, useEffect } from 'react';
import { api } from '../services/api';
import type { MarginPreview, LotCalculation } from '../services/api';
import { DollarSign, TrendingUp, AlertCircle, Calculator } from 'lucide-react';
import { RoutingIndicator } from './RoutingIndicator';

export type OrderType = 'MARKET' | 'LIMIT' | 'STOP' | 'STOP_LIMIT';
export type OrderSide = 'BUY' | 'SELL';

interface OrderEntryProps {
  symbol: string;
  currentBid?: number;
  currentAsk?: number;
  accountId: number;
  balance: number;
  onOrderPlaced?: () => void;
}

export function OrderEntry({
  symbol,
  currentBid = 0,
  currentAsk = 0,
  accountId,
  onOrderPlaced,
}: OrderEntryProps) {
  const [orderType, setOrderType] = useState<OrderType>('MARKET');
  const [side, setSide] = useState<OrderSide>('BUY');
  const [volume, setVolume] = useState(0.01);
  const [price, setPrice] = useState(0);
  const [triggerPrice, setTriggerPrice] = useState(0);
  const [sl, setSl] = useState(0);
  const [tp, setTp] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Risk calculator
  const [riskPercent, setRiskPercent] = useState(1);
  const [slPips, setSlPips] = useState(20);
  const [showRiskCalc, setShowRiskCalc] = useState(false);
  const [lotCalc, setLotCalc] = useState<LotCalculation | null>(null);

  // Margin preview
  const [marginPreview, setMarginPreview] = useState<MarginPreview | null>(null);

  // Auto-fill price for limit/stop orders
  useEffect(() => {
    if (orderType === 'LIMIT' || orderType === 'STOP' || orderType === 'STOP_LIMIT') {
      const basePrice = side === 'BUY' ? currentAsk : currentBid;
      if (!price || price === 0) {
        setPrice(basePrice);
      }
      if ((orderType === 'STOP' || orderType === 'STOP_LIMIT') && (!triggerPrice || triggerPrice === 0)) {
        setTriggerPrice(basePrice);
      }
    }
  }, [orderType, side, currentBid, currentAsk]);

  // Fetch margin preview when volume changes
  useEffect(() => {
    if (volume > 0 && symbol) {
      const timer = setTimeout(() => {
        api.risk
          .previewMargin(symbol, volume, side)
          .then(setMarginPreview)
          .catch(() => setMarginPreview(null));
      }, 500);

      return () => clearTimeout(timer);
    }
  }, [volume, symbol, side]);

  // Calculate recommended lot size from risk
  const calculateLot = async () => {
    if (!slPips || slPips <= 0) {
      setError('Enter Stop Loss in pips first');
      return;
    }

    try {
      const result = await api.risk.calculateLot(symbol, riskPercent, slPips);
      setLotCalc(result);
      setVolume(result.recommendedLot);
      setShowRiskCalc(true);
    } catch (err: any) {
      setError(err.message || 'Failed to calculate lot size');
    }
  };

  const handlePlaceOrder = async () => {
    setError(null);
    setLoading(true);

    try {
      if (orderType === 'MARKET') {
        await api.orders.placeMarketOrder({
          accountId,
          symbol,
          side,
          volume,
          sl: sl || undefined,
          tp: tp || undefined,
        });
      } else if (orderType === 'LIMIT') {
        await api.orders.placeLimitOrder({
          symbol,
          side,
          volume,
          price,
          sl: sl || undefined,
          tp: tp || undefined,
        });
      } else if (orderType === 'STOP') {
        await api.orders.placeStopOrder({
          symbol,
          side,
          volume,
          triggerPrice,
          sl: sl || undefined,
          tp: tp || undefined,
        });
      } else if (orderType === 'STOP_LIMIT') {
        await api.orders.placeStopLimitOrder({
          symbol,
          side,
          volume,
          triggerPrice,
          limitPrice: price,
          sl: sl || undefined,
          tp: tp || undefined,
        });
      }

      // Reset form and notify parent
      setVolume(0.01);
      setSl(0);
      setTp(0);
      setPrice(0);
      setTriggerPrice(0);
      onOrderPlaced?.();
    } catch (err: any) {
      setError(err.message || 'Order failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col gap-3 p-3 bg-zinc-900/50 rounded-lg border border-zinc-800">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-zinc-300">Order Entry</h3>
        <span className="text-xs text-zinc-500">{symbol}</span>
      </div>

      {/* Error Display */}
      {error && (
        <div className="flex items-center gap-2 px-3 py-2 bg-red-500/10 border border-red-500/20 rounded text-xs text-red-400">
          <AlertCircle size={14} />
          <span>{error}</span>
        </div>
      )}

      {/* Order Type Selector */}
      <div className="flex gap-1 p-1 bg-zinc-800/50 rounded">
        {(['MARKET', 'LIMIT', 'STOP', 'STOP_LIMIT'] as OrderType[]).map((type) => (
          <button
            key={type}
            onClick={() => setOrderType(type)}
            className={`flex-1 px-2 py-1 text-xs font-medium rounded transition-colors ${
              orderType === type
                ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                : 'text-zinc-500 hover:text-zinc-300'
            }`}
          >
            {type.replace('_', '-')}
          </button>
        ))}
      </div>

      {/* Side Selector */}
      <div className="flex gap-2">
        <button
          onClick={() => setSide('BUY')}
          className={`flex-1 px-4 py-2 rounded font-medium text-sm transition-colors ${
            side === 'BUY'
              ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
              : 'bg-emerald-500/5 text-emerald-600 border border-emerald-500/10 hover:bg-emerald-500/10'
          }`}
        >
          BUY {currentAsk > 0 && <span className="font-mono ml-1">{currentAsk.toFixed(5)}</span>}
        </button>
        <button
          onClick={() => setSide('SELL')}
          className={`flex-1 px-4 py-2 rounded font-medium text-sm transition-colors ${
            side === 'SELL'
              ? 'bg-red-500/20 text-red-400 border border-red-500/30'
              : 'bg-red-500/5 text-red-600 border border-red-500/10 hover:bg-red-500/10'
          }`}
        >
          SELL {currentBid > 0 && <span className="font-mono ml-1">{currentBid.toFixed(5)}</span>}
        </button>
      </div>

      {/* Volume Input */}
      <div className="flex flex-col gap-1">
        <label className="text-xs text-zinc-500 font-medium">Volume (Lots)</label>
        <input
          type="number"
          step={0.01}
          min={0.01}
          value={volume}
          onChange={(e) => setVolume(parseFloat(e.target.value) || 0)}
          className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono text-zinc-200 focus:outline-none focus:border-emerald-500/50"
        />
      </div>

      {/* Price Inputs (for Limit/Stop orders) */}
      {(orderType === 'LIMIT' || orderType === 'STOP_LIMIT') && (
        <div className="flex flex-col gap-1">
          <label className="text-xs text-zinc-500 font-medium">
            {orderType === 'STOP_LIMIT' ? 'Limit Price' : 'Price'}
          </label>
          <input
            type="number"
            step={0.00001}
            value={price}
            onChange={(e) => setPrice(parseFloat(e.target.value) || 0)}
            className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono text-zinc-200 focus:outline-none focus:border-emerald-500/50"
          />
        </div>
      )}

      {(orderType === 'STOP' || orderType === 'STOP_LIMIT') && (
        <div className="flex flex-col gap-1">
          <label className="text-xs text-zinc-500 font-medium">Trigger Price (Stop)</label>
          <input
            type="number"
            step={0.00001}
            value={triggerPrice}
            onChange={(e) => setTriggerPrice(parseFloat(e.target.value) || 0)}
            className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono text-zinc-200 focus:outline-none focus:border-emerald-500/50"
          />
        </div>
      )}

      {/* SL/TP Inputs */}
      <div className="grid grid-cols-2 gap-2">
        <div className="flex flex-col gap-1">
          <label className="text-xs text-zinc-500 font-medium">Stop Loss</label>
          <input
            type="number"
            step={0.00001}
            value={sl}
            onChange={(e) => setSl(parseFloat(e.target.value) || 0)}
            placeholder="Optional"
            className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono text-zinc-200 focus:outline-none focus:border-red-500/50"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-xs text-zinc-500 font-medium">Take Profit</label>
          <input
            type="number"
            step={0.00001}
            value={tp}
            onChange={(e) => setTp(parseFloat(e.target.value) || 0)}
            placeholder="Optional"
            className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono text-zinc-200 focus:outline-none focus:border-emerald-500/50"
          />
        </div>
      </div>

      {/* Risk Calculator Toggle */}
      <button
        onClick={() => setShowRiskCalc(!showRiskCalc)}
        className="flex items-center justify-center gap-2 px-3 py-2 bg-zinc-800/50 hover:bg-zinc-800 border border-zinc-700 rounded text-xs text-zinc-400 transition-colors"
      >
        <Calculator size={14} />
        {showRiskCalc ? 'Hide' : 'Show'} Risk Calculator
      </button>

      {/* Risk Calculator */}
      {showRiskCalc && (
        <div className="flex flex-col gap-2 p-3 bg-zinc-800/30 rounded border border-zinc-700">
          <div className="flex items-center gap-2 text-xs text-zinc-400">
            <TrendingUp size={12} />
            <span className="font-medium">Position Sizing</span>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-zinc-500">Risk %</label>
              <input
                type="number"
                step={0.1}
                min={0.1}
                max={10}
                value={riskPercent}
                onChange={(e) => setRiskPercent(parseFloat(e.target.value) || 1)}
                className="w-full bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-sm font-mono text-zinc-200 focus:outline-none focus:border-emerald-500/50"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-zinc-500">SL (pips)</label>
              <input
                type="number"
                step={1}
                min={1}
                value={slPips}
                onChange={(e) => setSlPips(parseFloat(e.target.value) || 20)}
                className="w-full bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-sm font-mono text-zinc-200 focus:outline-none focus:border-emerald-500/50"
              />
            </div>
          </div>
          <button
            onClick={calculateLot}
            className="w-full px-3 py-2 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/20 rounded text-xs font-medium transition-colors"
          >
            Calculate Lot Size
          </button>
          {lotCalc && (
            <div className="flex flex-col gap-1 text-xs">
              <div className="flex justify-between text-zinc-400">
                <span>Recommended:</span>
                <span className="font-mono text-emerald-400">{lotCalc.recommendedLot.toFixed(2)} lots</span>
              </div>
              <div className="flex justify-between text-zinc-500">
                <span>Risk Amount:</span>
                <span className="font-mono">${lotCalc.riskAmount.toFixed(2)}</span>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Margin Preview */}
      {marginPreview && (
        <div className="flex flex-col gap-1 p-2 bg-zinc-800/30 rounded border border-zinc-700 text-xs">
          <div className="flex items-center gap-2 text-zinc-400 mb-1">
            <DollarSign size={12} />
            <span className="font-medium">Margin Required</span>
          </div>
          <div className="flex justify-between">
            <span className="text-zinc-500">Required:</span>
            <span className="font-mono text-zinc-300">${marginPreview.requiredMargin.toFixed(2)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-zinc-500">Free After:</span>
            <span className={`font-mono ${marginPreview.canTrade ? 'text-emerald-400' : 'text-red-400'}`}>
              ${marginPreview.freeMarginAfter.toFixed(2)}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-zinc-500">Margin Level:</span>
            <span className={`font-mono ${marginPreview.marginLevelAfter >= 100 ? 'text-emerald-400' : 'text-red-400'}`}>
              {marginPreview.marginLevelAfter.toFixed(0)}%
            </span>
          </div>
        </div>
      )}

      {/* Routing Decision Preview */}
      {volume > 0 && (
        <RoutingIndicator
          symbol={symbol}
          volume={volume}
          accountId={accountId.toString()}
          side={side}
        />
      )}

      {/* Place Order Button */}
      <button
        onClick={handlePlaceOrder}
        disabled={loading || !volume || volume <= 0}
        className={`w-full px-4 py-3 rounded font-semibold text-sm transition-all ${
          side === 'BUY'
            ? 'bg-emerald-500 hover:bg-emerald-600 text-black disabled:bg-emerald-500/30'
            : 'bg-red-500 hover:bg-red-600 text-white disabled:bg-red-500/30'
        } disabled:cursor-not-allowed`}
      >
        {loading ? 'Placing Order...' : `${orderType} ${side} ${volume} Lots`}
      </button>
    </div>
  );
}
