/**
 * Order Entry / Trade Ticket
 * Professional MT5-style order placement form
 */

import { useState, useEffect, useMemo } from 'react';
import {
  TrendingUp,
  TrendingDown,
  Calculator,
  AlertTriangle,
  Info,
  X,
  Clock,
  Calendar,
} from 'lucide-react';
import { useTradeSettings } from '../store/useSettingsStore';
import { useAppStore } from '../store/useAppStore';
import { API_ENDPOINTS, API_BASE_URL } from '../config/api';

interface OrderEntryProps {
  symbol: string;
  onClose: () => void;
  onOrderPlaced?: () => void;
}

type OrderType =
  | 'Market'
  | 'Buy Limit'
  | 'Sell Limit'
  | 'Buy Stop'
  | 'Sell Stop'
  | 'Buy Stop Limit'
  | 'Sell Stop Limit';

type ExpiryType = 'GTC' | 'Today' | 'Specified';

type PipsMode = 'price' | 'pips';

const LOT_PRESETS = [0.01, 0.1, 0.5, 1.0, 5.0, 10.0];

export default function OrderEntry({ symbol, onClose, onOrderPlaced }: OrderEntryProps) {
  const tradeSettings = useTradeSettings();
  const { ticks, account } = useAppStore();

  // Current prices from ticks
  const currentTick = ticks[symbol];
  const currentBid = currentTick?.bid || 0;
  const currentAsk = currentTick?.ask || 0;

  // Order parameters
  const [orderType, setOrderType] = useState<OrderType>('Market');
  const [volume, setVolume] = useState(tradeSettings.defaultVolume);
  const [price, setPrice] = useState('');
  const [stopLoss, setStopLoss] = useState('');
  const [takeProfit, setTakeProfit] = useState('');
  const [slMode, setSlMode] = useState<PipsMode>('price');
  const [tpMode, setTpMode] = useState<PipsMode>('price');
  const [expiry, setExpiry] = useState<ExpiryType>('GTC');
  const [expiryDate, setExpiryDate] = useState('');
  const [expiryTime, setExpiryTime] = useState('23:59');
  const [comment, setComment] = useState('');

  // Symbol specification from backend (contractSize, currency, leverage)
  const [symbolSpec, setSymbolSpec] = useState<{
    contractSize: number;
    currency: string;
    marginRate: number;
  } | null>(null);
  const [configLeverage, setConfigLeverage] = useState<number | null>(null);

  // Fetch symbol spec from backend: GET /api/symbols/{symbol}/spec
  useEffect(() => {
    let cancelled = false;
    async function fetchSymbolSpec() {
      try {
        const response = await fetch(`${API_BASE_URL}/api/symbols/${symbol}/spec`, {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${localStorage.getItem('jwt_token')}`,
          },
        });
        if (response.ok) {
          const spec = await response.json();
          if (!cancelled && spec) {
            setSymbolSpec({
              contractSize: spec.contractSize || 100000,
              currency: spec.currency || 'USD',
              marginRate: spec.marginRate || 0.01,
            });
          }
        }
      } catch (err) {
        console.warn('[OrderEntry] Failed to fetch symbol spec, using defaults:', err);
      }
    }
    fetchSymbolSpec();
    return () => { cancelled = true; };
  }, [symbol]);

  // Fetch broker config for default leverage: GET /api/config
  useEffect(() => {
    let cancelled = false;
    async function fetchConfig() {
      try {
        const response = await fetch(`${API_BASE_URL}/api/config`, {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${localStorage.getItem('jwt_token')}`,
          },
        });
        if (response.ok) {
          const config = await response.json();
          if (!cancelled && config?.defaultLeverage) {
            setConfigLeverage(config.defaultLeverage);
          }
        }
      } catch (err) {
        console.warn('[OrderEntry] Failed to fetch config, using default leverage:', err);
      }
    }
    fetchConfig();
    return () => { cancelled = true; };
  }, []);

  // UI state
  const [isPlacing, setIsPlacing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Determine if JPY pair (different pip size)
  const isJPY = symbol.includes('JPY');
  const pipSize = isJPY ? 0.01 : 0.0001;
  const pipMultiplier = isJPY ? 100 : 10000;

  // Check if order type is pending
  const isPendingOrder = orderType !== 'Market';

  // Auto-fill price for pending orders
  useEffect(() => {
    if (isPendingOrder && !price) {
      if (orderType.includes('Buy')) {
        setPrice(currentAsk.toFixed(isJPY ? 3 : 5));
      } else {
        setPrice(currentBid.toFixed(isJPY ? 3 : 5));
      }
    }
  }, [orderType, currentBid, currentAsk, isJPY, isPendingOrder, price]);

  // Calculate spread
  const spread = currentAsk - currentBid;
  const spreadPips = spread * pipMultiplier;

  // Convert pips to price
  const pipsToPrice = (pips: number, basePrice: number, direction: 'above' | 'below'): number => {
    const priceChange = pips * pipSize;
    return direction === 'above' ? basePrice + priceChange : basePrice - priceChange;
  };

  // Convert price to pips
  const priceToPips = (priceValue: number, basePrice: number): number => {
    return Math.abs((priceValue - basePrice) / pipSize);
  };

  // Get entry price for calculations
  const entryPrice = useMemo(() => {
    if (orderType === 'Market') {
      return orderType.includes('Buy') ? currentAsk : currentBid;
    }
    return parseFloat(price) || 0;
  }, [orderType, price, currentBid, currentAsk]);

  // Calculate SL price
  const slPrice = useMemo(() => {
    if (!stopLoss) return null;
    const slValue = parseFloat(stopLoss);
    if (isNaN(slValue)) return null;

    if (slMode === 'pips') {
      const isBuy = orderType.includes('Buy');
      return pipsToPrice(slValue, entryPrice, isBuy ? 'below' : 'above');
    }
    return slValue;
  }, [stopLoss, slMode, entryPrice, orderType, pipSize]);

  // Calculate TP price
  const tpPrice = useMemo(() => {
    if (!takeProfit) return null;
    const tpValue = parseFloat(takeProfit);
    if (isNaN(tpValue)) return null;

    if (tpMode === 'pips') {
      const isBuy = orderType.includes('Buy');
      return pipsToPrice(tpValue, entryPrice, isBuy ? 'above' : 'below');
    }
    return tpValue;
  }, [takeProfit, tpMode, entryPrice, orderType, pipSize]);

  // Risk calculator - uses real values from /api/symbols/{symbol}/spec and /api/config when available
  const riskCalc = useMemo(() => {
    const leverage = configLeverage || (account as any)?.leverage || 100;
    const accountCurrency = symbolSpec?.currency || 'USD';
    const contractSize = symbolSpec?.contractSize || 100000;

    // Calculate margin required
    const margin = symbolSpec?.marginRate
      ? volume * contractSize * entryPrice * symbolSpec.marginRate
      : (volume * contractSize * entryPrice) / leverage;

    // Calculate pip value (USD per pip for 1 lot)
    const pipValue = (contractSize * pipSize) * volume;

    // Calculate potential loss/profit
    let potentialLoss = 0;
    let potentialProfit = 0;

    if (slPrice) {
      const slPips = priceToPips(slPrice, entryPrice);
      potentialLoss = slPips * pipValue;
    }

    if (tpPrice) {
      const tpPips = priceToPips(tpPrice, entryPrice);
      potentialProfit = tpPips * pipValue;
    }

    return {
      margin,
      pipValue,
      potentialLoss,
      potentialProfit,
    };
  }, [volume, entryPrice, slPrice, tpPrice, pipSize, account, symbolSpec, configLeverage]);

  // Handle volume preset click
  const handleVolumePreset = (preset: number) => {
    setVolume(preset);
  };

  // Handle order placement
  const handlePlaceOrder = async (side: 'BUY' | 'SELL') => {
    if (isPlacing) return;

    setError(null);
    setIsPlacing(true);

    try {
      // Validate inputs
      if (volume <= 0) {
        throw new Error('Volume must be greater than 0');
      }

      if (isPendingOrder && !price) {
        throw new Error('Price is required for pending orders');
      }

      // Build order data
      const orderData: any = {
        symbol,
        side,
        volume,
        type: orderType,
      };

      // Add price for pending orders
      if (isPendingOrder) {
        orderData.price = parseFloat(price);
      }

      // Add SL/TP if provided
      if (slPrice) {
        orderData.stopLoss = slPrice;
      }
      if (tpPrice) {
        orderData.takeProfit = tpPrice;
      }

      // Add expiry for pending orders
      if (isPendingOrder && expiry !== 'GTC') {
        if (expiry === 'Today') {
          const today = new Date();
          today.setHours(23, 59, 59, 999);
          orderData.expiry = today.toISOString();
        } else if (expiry === 'Specified' && expiryDate) {
          const expiryDateTime = new Date(`${expiryDate}T${expiryTime}`);
          orderData.expiry = expiryDateTime.toISOString();
        }
      }

      // Add comment if provided
      if (comment.trim()) {
        orderData.comment = comment.trim().slice(0, 64);
      }

      // Add deviation for market orders
      if (orderType === 'Market') {
        orderData.deviation = tradeSettings.defaultDeviation;
      }

      // Call API
      const endpoint = orderType === 'Market'
        ? API_ENDPOINTS.order
        : API_ENDPOINTS.orders;

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('jwt_token')}`,
        },
        body: JSON.stringify(orderData),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Order placement failed');
      }

      // Success
      onOrderPlaced?.();
      onClose();
    } catch (err: any) {
      console.error('Order placement failed:', err);
      setError(err.message || 'Failed to place order');
      setTimeout(() => setError(null), 5000);
    } finally {
      setIsPlacing(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg shadow-2xl w-[480px] max-h-[90vh] overflow-y-auto custom-scrollbar">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 bg-[#252525] border-b border-zinc-700">
          <div className="flex items-center gap-3">
            <h3 className="text-sm font-bold text-zinc-200">New Order</h3>
            <span className="px-2 py-0.5 rounded bg-blue-500/20 text-blue-400 text-xs font-mono font-bold">
              {symbol}
            </span>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-zinc-700 text-zinc-400"
          >
            <X size={16} />
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4">
          {/* Error Display */}
          {error && (
            <div className="px-3 py-2 bg-rose-500/10 border border-rose-500/20 rounded text-xs text-rose-400 flex items-center gap-2">
              <AlertTriangle size={14} />
              <span>{error}</span>
            </div>
          )}

          {/* Current Prices */}
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-[#252525] border border-zinc-700/50 rounded p-3">
              <div className="text-xs text-zinc-500 mb-1">Bid</div>
              <div className="text-lg font-mono font-bold text-rose-400">
                {currentBid.toFixed(isJPY ? 3 : 5)}
              </div>
            </div>
            <div className="bg-[#252525] border border-zinc-700/50 rounded p-3">
              <div className="text-xs text-zinc-500 mb-1">Ask</div>
              <div className="text-lg font-mono font-bold text-blue-400">
                {currentAsk.toFixed(isJPY ? 3 : 5)}
              </div>
            </div>
          </div>

          <div className="flex items-center justify-between text-xs text-zinc-500">
            <span>Spread:</span>
            <span className="font-mono text-zinc-400">{spreadPips.toFixed(1)} pips</span>
          </div>

          {/* Order Type */}
          <div>
            <label className="block text-xs text-zinc-400 mb-2">Order Type</label>
            <select
              value={orderType}
              onChange={(e) => setOrderType(e.target.value as OrderType)}
              className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm focus:outline-none focus:border-blue-500"
            >
              <option value="Market">Market Execution</option>
              <option value="Buy Limit">Buy Limit</option>
              <option value="Sell Limit">Sell Limit</option>
              <option value="Buy Stop">Buy Stop</option>
              <option value="Sell Stop">Sell Stop</option>
              <option value="Buy Stop Limit">Buy Stop Limit</option>
              <option value="Sell Stop Limit">Sell Stop Limit</option>
            </select>
          </div>

          {/* Volume */}
          <div>
            <label className="block text-xs text-zinc-400 mb-2">Volume (Lots)</label>
            <input
              type="number"
              step={0.01}
              min={0.01}
              value={volume}
              onChange={(e) => {
                const val = parseFloat(e.target.value);
                if (!isNaN(val) && val >= 0.01) {
                  setVolume(val);
                }
              }}
              className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm font-mono focus:outline-none focus:border-blue-500"
            />
            {/* Lot Presets */}
            <div className="flex flex-wrap gap-2 mt-2">
              {LOT_PRESETS.map((preset) => (
                <button
                  key={preset}
                  onClick={() => handleVolumePreset(preset)}
                  className={`px-2 py-1 rounded text-xs font-medium transition-colors ${
                    volume === preset
                      ? 'bg-blue-500/30 text-blue-400 border border-blue-500/50'
                      : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700 border border-zinc-700'
                  }`}
                >
                  {preset}
                </button>
              ))}
            </div>
          </div>

          {/* Price (for pending orders) */}
          {isPendingOrder && (
            <div>
              <label className="block text-xs text-zinc-400 mb-2">Price</label>
              <input
                type="number"
                step={pipSize}
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm font-mono focus:outline-none focus:border-blue-500"
                placeholder={currentAsk.toFixed(isJPY ? 3 : 5)}
              />
            </div>
          )}

          {/* Stop Loss */}
          <div>
            <label className="block text-xs text-zinc-400 mb-2">Stop Loss (Optional)</label>
            <div className="flex gap-2">
              <input
                type="number"
                step={slMode === 'pips' ? 1 : pipSize}
                value={stopLoss}
                onChange={(e) => setStopLoss(e.target.value)}
                className="flex-1 px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm font-mono focus:outline-none focus:border-blue-500"
                placeholder={slMode === 'pips' ? 'Pips' : 'Price'}
              />
              <button
                onClick={() => setSlMode(slMode === 'price' ? 'pips' : 'price')}
                className="px-3 py-2 bg-zinc-800 hover:bg-zinc-700 border border-zinc-700 rounded text-xs text-zinc-300 font-medium whitespace-nowrap"
              >
                {slMode === 'price' ? 'Price' : 'Pips'}
              </button>
            </div>
            {slPrice && slMode === 'pips' && (
              <div className="text-xs text-zinc-500 mt-1">
                = {slPrice.toFixed(isJPY ? 3 : 5)}
              </div>
            )}
          </div>

          {/* Take Profit */}
          <div>
            <label className="block text-xs text-zinc-400 mb-2">Take Profit (Optional)</label>
            <div className="flex gap-2">
              <input
                type="number"
                step={tpMode === 'pips' ? 1 : pipSize}
                value={takeProfit}
                onChange={(e) => setTakeProfit(e.target.value)}
                className="flex-1 px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm font-mono focus:outline-none focus:border-blue-500"
                placeholder={tpMode === 'pips' ? 'Pips' : 'Price'}
              />
              <button
                onClick={() => setTpMode(tpMode === 'price' ? 'pips' : 'price')}
                className="px-3 py-2 bg-zinc-800 hover:bg-zinc-700 border border-zinc-700 rounded text-xs text-zinc-300 font-medium whitespace-nowrap"
              >
                {tpMode === 'price' ? 'Price' : 'Pips'}
              </button>
            </div>
            {tpPrice && tpMode === 'pips' && (
              <div className="text-xs text-zinc-500 mt-1">
                = {tpPrice.toFixed(isJPY ? 3 : 5)}
              </div>
            )}
          </div>

          {/* Expiry (for pending orders) */}
          {isPendingOrder && (
            <div>
              <label className="block text-xs text-zinc-400 mb-2">Expiry</label>
              <div className="space-y-2">
                <select
                  value={expiry}
                  onChange={(e) => setExpiry(e.target.value as ExpiryType)}
                  className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm focus:outline-none focus:border-blue-500"
                >
                  <option value="GTC">Good Till Cancelled (GTC)</option>
                  <option value="Today">Today</option>
                  <option value="Specified">Specified Date/Time</option>
                </select>

                {expiry === 'Specified' && (
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <input
                        type="date"
                        value={expiryDate}
                        onChange={(e) => setExpiryDate(e.target.value)}
                        className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-xs focus:outline-none focus:border-blue-500"
                      />
                    </div>
                    <div>
                      <input
                        type="time"
                        value={expiryTime}
                        onChange={(e) => setExpiryTime(e.target.value)}
                        className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-xs focus:outline-none focus:border-blue-500"
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Comment */}
          <div>
            <label className="block text-xs text-zinc-400 mb-2">Comment (Optional, max 64 chars)</label>
            <input
              type="text"
              maxLength={64}
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm focus:outline-none focus:border-blue-500"
              placeholder="Order comment..."
            />
            <div className="text-xs text-zinc-600 text-right mt-1">
              {comment.length}/64
            </div>
          </div>

          {/* Risk Calculator */}
          <div className="bg-[#252525] border border-zinc-700/50 rounded p-3">
            <div className="flex items-center gap-2 mb-3">
              <Calculator size={14} className="text-blue-400" />
              <h4 className="text-xs font-bold text-zinc-300">Risk Calculator</h4>
            </div>
            <div className="space-y-2 text-xs">
              <div className="flex justify-between">
                <span className="text-zinc-500">Required Margin:</span>
                <span className="text-zinc-300 font-mono font-semibold">
                  ${riskCalc.margin.toFixed(2)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-zinc-500">Pip Value:</span>
                <span className="text-zinc-300 font-mono font-semibold">
                  ${riskCalc.pipValue.toFixed(2)}
                </span>
              </div>
              {riskCalc.potentialLoss > 0 && (
                <div className="flex justify-between">
                  <span className="text-zinc-500">Potential Loss (SL):</span>
                  <span className="text-rose-400 font-mono font-semibold">
                    -${riskCalc.potentialLoss.toFixed(2)}
                  </span>
                </div>
              )}
              {riskCalc.potentialProfit > 0 && (
                <div className="flex justify-between">
                  <span className="text-zinc-500">Potential Profit (TP):</span>
                  <span className="text-emerald-400 font-mono font-semibold">
                    +${riskCalc.potentialProfit.toFixed(2)}
                  </span>
                </div>
              )}
              {riskCalc.potentialProfit > 0 && riskCalc.potentialLoss > 0 && (
                <div className="flex justify-between">
                  <span className="text-zinc-500">Risk/Reward Ratio:</span>
                  <span className="text-blue-400 font-mono font-semibold">
                    1:{(riskCalc.potentialProfit / riskCalc.potentialLoss).toFixed(2)}
                  </span>
                </div>
              )}
            </div>
          </div>

          {/* One-Click Mode Info */}
          {tradeSettings.oneClickTrading && orderType === 'Market' && (
            <div className="flex items-start gap-2 px-3 py-2 bg-yellow-500/10 border border-yellow-500/20 rounded text-xs text-yellow-400">
              <Info size={14} className="flex-shrink-0 mt-0.5" />
              <span>One-click mode is enabled. Orders will be placed immediately.</span>
            </div>
          )}

          {/* Action Buttons */}
          <div className="grid grid-cols-2 gap-3 pt-2">
            {orderType === 'Market' ? (
              <>
                {/* SELL Button */}
                <button
                  onClick={() => handlePlaceOrder('SELL')}
                  disabled={isPlacing || volume <= 0}
                  className="flex flex-col items-center justify-center px-4 py-3 bg-rose-600 hover:bg-rose-700 disabled:bg-rose-600/30 disabled:cursor-not-allowed text-white rounded-lg transition-colors group"
                >
                  <TrendingDown size={18} className="mb-1 group-hover:scale-110 transition-transform" />
                  <span className="text-sm font-bold">SELL</span>
                  <span className="text-xs font-mono mt-0.5 opacity-90">
                    {currentBid.toFixed(isJPY ? 3 : 5)}
                  </span>
                </button>

                {/* BUY Button */}
                <button
                  onClick={() => handlePlaceOrder('BUY')}
                  disabled={isPlacing || volume <= 0}
                  className="flex flex-col items-center justify-center px-4 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-600/30 disabled:cursor-not-allowed text-white rounded-lg transition-colors group"
                >
                  <TrendingUp size={18} className="mb-1 group-hover:scale-110 transition-transform" />
                  <span className="text-sm font-bold">BUY</span>
                  <span className="text-xs font-mono mt-0.5 opacity-90">
                    {currentAsk.toFixed(isJPY ? 3 : 5)}
                  </span>
                </button>
              </>
            ) : (
              <>
                {/* PLACE SELL ORDER Button */}
                {orderType.includes('Sell') && (
                  <button
                    onClick={() => handlePlaceOrder('SELL')}
                    disabled={isPlacing || volume <= 0 || !price}
                    className="col-span-2 flex items-center justify-center gap-2 px-4 py-3 bg-rose-600 hover:bg-rose-700 disabled:bg-rose-600/30 disabled:cursor-not-allowed text-white rounded-lg transition-colors font-bold"
                  >
                    <TrendingDown size={18} />
                    <span>PLACE SELL ORDER</span>
                  </button>
                )}

                {/* PLACE BUY ORDER Button */}
                {orderType.includes('Buy') && (
                  <button
                    onClick={() => handlePlaceOrder('BUY')}
                    disabled={isPlacing || volume <= 0 || !price}
                    className="col-span-2 flex items-center justify-center gap-2 px-4 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-600/30 disabled:cursor-not-allowed text-white rounded-lg transition-colors font-bold"
                  >
                    <TrendingUp size={18} />
                    <span>PLACE BUY ORDER</span>
                  </button>
                )}
              </>
            )}
          </div>

          {/* Status Indicator */}
          {isPlacing && (
            <div className="text-center">
              <div className="text-xs text-blue-400 font-medium animate-pulse">
                Placing order...
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
