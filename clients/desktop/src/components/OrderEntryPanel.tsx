/**
 * Enhanced Order Entry Panel
 * Supports Market, Limit, Stop, and Stop-Limit orders with SL/TP
 */

import { useState, useMemo } from 'react';
import { TrendingUp, TrendingDown, DollarSign, Shield, Target } from 'lucide-react';
import { useAppStore } from '../store/useAppStore';

const API_BASE = 'http://localhost:7999';

type OrderType = 'MARKET' | 'LIMIT' | 'STOP' | 'STOP_LIMIT';
type OrderSide = 'BUY' | 'SELL';

export const OrderEntryPanel = () => {
  const {
    selectedSymbol,
    ticks,
    orderVolume,
    setOrderVolume,
    isPlacingOrder,
    setLoadingStates,
    accountId,
  } = useAppStore();

  const [orderType, setOrderType] = useState<OrderType>('MARKET');
  const [limitPrice, setLimitPrice] = useState('');
  const [stopPrice, setStopPrice] = useState('');
  const [stopLoss, setStopLoss] = useState('');
  const [takeProfit, setTakeProfit] = useState('');

  const currentTick = ticks[selectedSymbol];

  const spread = useMemo(() => {
    if (!currentTick) return 0;
    return currentTick.ask - currentTick.bid;
  }, [currentTick]);

  const spreadPips = useMemo(() => {
    if (!selectedSymbol || !spread) return 0;
    // Forex pairs typically have 4 or 5 decimal places
    const pipMultiplier = selectedSymbol.includes('JPY') ? 100 : 10000;
    return spread * pipMultiplier;
  }, [selectedSymbol, spread]);

  const calculateEstimatedValue = (side: OrderSide) => {
    if (!currentTick) return 0;
    const price = side === 'BUY' ? currentTick.ask : currentTick.bid;
    return orderVolume * 100000 * price; // 1 lot = 100,000 units
  };

  const placeOrder = async (side: OrderSide) => {
    if (!accountId) {
      alert('Please login first');
      return;
    }

    if (orderVolume <= 0) {
      alert('Invalid volume');
      return;
    }

    setLoadingStates({ isPlacingOrder: true });

    try {
      let endpoint = '';
      let body: any = {
        accountId,
        symbol: selectedSymbol,
        side,
        volume: orderVolume,
      };

      switch (orderType) {
        case 'MARKET':
          endpoint = `${API_BASE}/api/orders/market`;
          break;

        case 'LIMIT':
          if (!limitPrice || parseFloat(limitPrice) <= 0) {
            alert('Invalid limit price');
            setLoadingStates({ isPlacingOrder: false });
            return;
          }
          endpoint = `${API_BASE}/order/limit`;
          body.price = parseFloat(limitPrice);
          break;

        case 'STOP':
          if (!stopPrice || parseFloat(stopPrice) <= 0) {
            alert('Invalid stop price');
            setLoadingStates({ isPlacingOrder: false });
            return;
          }
          endpoint = `${API_BASE}/order/stop`;
          body.triggerPrice = parseFloat(stopPrice);
          break;

        case 'STOP_LIMIT':
          if (!stopPrice || !limitPrice) {
            alert('Invalid stop or limit price');
            setLoadingStates({ isPlacingOrder: false });
            return;
          }
          endpoint = `${API_BASE}/order/stop-limit`;
          body.triggerPrice = parseFloat(stopPrice);
          body.limitPrice = parseFloat(limitPrice);
          break;
      }

      // Add SL/TP if specified
      if (stopLoss && parseFloat(stopLoss) > 0) {
        body.sl = parseFloat(stopLoss);
      }
      if (takeProfit && parseFloat(takeProfit) > 0) {
        body.tp = parseFloat(takeProfit);
      }

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      if (response.ok) {
        const result = await response.json();
        console.log('Order placed:', result);

        // Reset form on success
        if (orderType !== 'MARKET') {
          setLimitPrice('');
          setStopPrice('');
        }
        setStopLoss('');
        setTakeProfit('');

        alert(`${orderType} ${side} order placed successfully!`);
      } else {
        const error = await response.text();
        alert(`Order failed: ${error}`);
      }
    } catch (error) {
      console.error('Order placement error:', error);
      alert('Failed to place order. Check console for details.');
    } finally {
      setLoadingStates({ isPlacingOrder: false });
    }
  };

  if (!currentTick) {
    return (
      <div className="p-4 text-center text-zinc-500">
        Waiting for market data...
      </div>
    );
  }

  return (
    <div className="p-4 space-y-4">
      {/* Symbol Header */}
      <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
        <div>
          <h3 className="text-lg font-bold text-white">{selectedSymbol}</h3>
          <p className="text-xs text-zinc-500">
            Spread: {spreadPips.toFixed(1)} pips ({currentTick.lp || 'Unknown LP'})
          </p>
        </div>
        <div className="text-right">
          <div className="text-xs text-zinc-500">Bid / Ask</div>
          <div className="font-mono text-sm">
            <span className="text-red-400">{currentTick.bid.toFixed(5)}</span>
            {' / '}
            <span className="text-green-400">{currentTick.ask.toFixed(5)}</span>
          </div>
        </div>
      </div>

      {/* Order Type Selection */}
      <div className="grid grid-cols-4 gap-2">
        {(['MARKET', 'LIMIT', 'STOP', 'STOP_LIMIT'] as OrderType[]).map((type) => (
          <button
            key={type}
            onClick={() => setOrderType(type)}
            className={`px-3 py-2 rounded text-xs font-medium transition-colors ${
              orderType === type
                ? 'bg-blue-600 text-white'
                : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
            }`}
          >
            {type.replace('_', ' ')}
          </button>
        ))}
      </div>

      {/* Volume Input */}
      <div>
        <label className="block text-xs text-zinc-400 mb-1">
          <DollarSign className="inline w-3 h-3 mr-1" />
          Volume (Lots)
        </label>
        <input
          type="number"
          step="0.01"
          min="0.01"
          value={orderVolume}
          onChange={(e) => setOrderVolume(parseFloat(e.target.value) || 0.01)}
          className="w-full px-3 py-2 bg-zinc-900 border border-zinc-700 rounded text-white focus:outline-none focus:border-blue-500"
        />
        <div className="text-xs text-zinc-500 mt-1">
          Quick: {[0.01, 0.1, 0.5, 1, 5].map((vol) => (
            <button
              key={vol}
              onClick={() => setOrderVolume(vol)}
              className="ml-2 text-blue-400 hover:text-blue-300"
            >
              {vol}
            </button>
          ))}
        </div>
      </div>

      {/* Price Inputs for non-market orders */}
      {orderType !== 'MARKET' && (
        <div className="space-y-3">
          {(orderType === 'LIMIT' || orderType === 'STOP_LIMIT') && (
            <div>
              <label className="block text-xs text-zinc-400 mb-1">Limit Price</label>
              <input
                type="number"
                step="0.00001"
                value={limitPrice}
                onChange={(e) => setLimitPrice(e.target.value)}
                placeholder={currentTick.ask.toFixed(5)}
                className="w-full px-3 py-2 bg-zinc-900 border border-zinc-700 rounded text-white focus:outline-none focus:border-blue-500"
              />
            </div>
          )}

          {(orderType === 'STOP' || orderType === 'STOP_LIMIT') && (
            <div>
              <label className="block text-xs text-zinc-400 mb-1">Stop Price</label>
              <input
                type="number"
                step="0.00001"
                value={stopPrice}
                onChange={(e) => setStopPrice(e.target.value)}
                placeholder={currentTick.bid.toFixed(5)}
                className="w-full px-3 py-2 bg-zinc-900 border border-zinc-700 rounded text-white focus:outline-none focus:border-blue-500"
              />
            </div>
          )}
        </div>
      )}

      {/* Stop Loss / Take Profit */}
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="block text-xs text-zinc-400 mb-1">
            <Shield className="inline w-3 h-3 mr-1" />
            Stop Loss
          </label>
          <input
            type="number"
            step="0.00001"
            value={stopLoss}
            onChange={(e) => setStopLoss(e.target.value)}
            placeholder="Optional"
            className="w-full px-3 py-2 bg-zinc-900 border border-zinc-700 rounded text-white text-sm focus:outline-none focus:border-red-500"
          />
        </div>
        <div>
          <label className="block text-xs text-zinc-400 mb-1">
            <Target className="inline w-3 h-3 mr-1" />
            Take Profit
          </label>
          <input
            type="number"
            step="0.00001"
            value={takeProfit}
            onChange={(e) => setTakeProfit(e.target.value)}
            placeholder="Optional"
            className="w-full px-3 py-2 bg-zinc-900 border border-zinc-700 rounded text-white text-sm focus:outline-none focus:border-green-500"
          />
        </div>
      </div>

      {/* Order Buttons */}
      <div className="grid grid-cols-2 gap-3 pt-2">
        <button
          onClick={() => placeOrder('BUY')}
          disabled={isPlacingOrder}
          className="flex items-center justify-center gap-2 px-4 py-3 bg-green-600 hover:bg-green-700 disabled:bg-green-900 disabled:cursor-not-allowed text-white font-semibold rounded transition-colors"
        >
          <TrendingUp className="w-4 h-4" />
          BUY {currentTick.ask.toFixed(5)}
        </button>
        <button
          onClick={() => placeOrder('SELL')}
          disabled={isPlacingOrder}
          className="flex items-center justify-center gap-2 px-4 py-3 bg-red-600 hover:bg-red-700 disabled:bg-red-900 disabled:cursor-not-allowed text-white font-semibold rounded transition-colors"
        >
          <TrendingDown className="w-4 h-4" />
          SELL {currentTick.bid.toFixed(5)}
        </button>
      </div>

      {/* Estimated Value */}
      <div className="text-xs text-zinc-500 text-center">
        Est. value: ${calculateEstimatedValue('BUY').toFixed(2)} (BUY) / ${calculateEstimatedValue('SELL').toFixed(2)} (SELL)
      </div>
    </div>
  );
};
