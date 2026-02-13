/**
 * Currency Converter / Cross-Rate Calculator
 * Multi-currency converter with live rates and cross-rate matrix
 */

import React, { useState, useMemo } from 'react';
import {
  RefreshCw,
  Star,
  TrendingUp,
  DollarSign,
  ArrowLeftRight,
  Calculator,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import { useCurrencyStore } from '../store/useCurrencyStore';
import { useAppStore } from '../store/useAppStore';

const CURRENCIES = [
  'USD', 'EUR', 'GBP', 'JPY', 'CHF', 'AUD', 'CAD', 'NZD',
  'SEK', 'NOK', 'DKK', 'PLN', 'CZK', 'HUF', 'RON', 'BGN',
  'TRY', 'ZAR', 'MXN', 'BRL', 'CNY', 'HKD', 'SGD', 'INR',
  'THB', 'MYR', 'IDR', 'PHP', 'KRW', 'TWD', 'RUB', 'ILS',
];

const MAJOR_CURRENCIES = ['USD', 'EUR', 'GBP', 'JPY', 'CHF', 'AUD', 'CAD', 'NZD'];

interface ExchangeRate {
  pair: string;
  bid: number;
  ask: number;
  change: number; // % change since last update
  history: number[]; // Last 24 hours
}

// TODO: Backend API needed - No /api/currency/rates or /api/exchange-rates endpoint exists yet.
// When a backend endpoint is added, replace generateExchangeRates() with an API call
// and use the live WebSocket tick data (useAppStore ticks) for forex pair rates.
// For now, rates are generated from hardcoded base rates.
function generateExchangeRates(): Map<string, ExchangeRate> {
  const rates = new Map<string, ExchangeRate>();

  // Base rates to USD
  const baseRates: Record<string, number> = {
    'USD': 1.0000, 'EUR': 0.9200, 'GBP': 0.7900, 'JPY': 148.50, 'CHF': 0.8750,
    'AUD': 1.5270, 'CAD': 1.3650, 'NZD': 1.6530, 'SEK': 10.45, 'NOK': 10.85,
    'DKK': 6.87, 'PLN': 3.98, 'CZK': 22.85, 'HUF': 355.20, 'RON': 4.56,
    'BGN': 1.80, 'TRY': 30.75, 'ZAR': 18.45, 'MXN': 17.20, 'BRL': 4.95,
    'CNY': 7.18, 'HKD': 7.82, 'SGD': 1.33, 'INR': 83.10, 'THB': 35.40,
    'MYR': 4.68, 'IDR': 15650, 'PHP': 56.20, 'KRW': 1320, 'TWD': 31.50,
    'RUB': 92.50, 'ILS': 3.65,
  };

  // Generate all pairs
  CURRENCIES.forEach((from) => {
    CURRENCIES.forEach((to) => {
      if (from !== to) {
        const fromRate = baseRates[from] || 1;
        const toRate = baseRates[to] || 1;

        // Calculate cross rate
        const midRate = fromRate / toRate;
        const spread = 0.0005; // 0.05% spread
        const bid = midRate * (1 - spread);
        const ask = midRate * (1 + spread);

        // Generate 24h history (24 hourly points)
        const history: number[] = [];
        let currentRate = midRate;
        for (let i = 0; i < 24; i++) {
          currentRate += (Math.random() - 0.5) * midRate * 0.002; // ±0.2% volatility
          history.push(currentRate);
        }

        const change = ((midRate - history[0]) / history[0]) * 100;

        rates.set(`${from}${to}`, {
          pair: `${from}${to}`,
          bid,
          ask,
          change,
          history,
        });
      }
    });
  });

  return rates;
}

export const CurrencyConverter: React.FC = () => {
  const {
    favorites,
    addFavorite,
    removeFavorite,
    isFavorite,
    addToLastUsed,
    multiConvertCurrencies,
    toggleMultiConvertCurrency,
  } = useCurrencyStore();

  const [amount, setAmount] = useState<number>(1000);
  const [fromCurrency, setFromCurrency] = useState<string>('USD');
  const [toCurrency, setToCurrency] = useState<string>('EUR');
  const [showMultiConvert, setShowMultiConvert] = useState<boolean>(false);

  // Pip value calculator
  const [lotSize, setLotSize] = useState<number>(1);
  const [pipPair, setPipPair] = useState<string>('EURUSD');
  const [accountCurrency, setAccountCurrency] = useState<string>('USD');

  // Use live tick data from WebSocket to enhance exchange rates for available forex pairs
  const liveTicks = useAppStore((state) => state.ticks);

  // Generate exchange rates, overriding with live tick data where available
  const exchangeRates = useMemo(() => {
    const rates = generateExchangeRates();
    // Override rates for forex pairs that have live WebSocket data
    Object.keys(liveTicks).forEach((symbol) => {
      const tick = liveTicks[symbol];
      if (tick && tick.bid > 0 && tick.ask > 0 && symbol.length === 6) {
        const from = symbol.substring(0, 3);
        const to = symbol.substring(3, 6);
        const existing = rates.get(`${from}${to}`);
        if (existing) {
          existing.bid = tick.bid;
          existing.ask = tick.ask;
          rates.set(`${from}${to}`, existing);
        }
      }
    });
    return rates;
  }, [liveTicks]);

  // Get current conversion rate
  const currentRate = useMemo(() => {
    return exchangeRates.get(`${fromCurrency}${toCurrency}`);
  }, [fromCurrency, toCurrency, exchangeRates]);

  // Calculate converted amount
  const convertedAmount = useMemo(() => {
    if (!currentRate) return 0;
    return amount * currentRate.bid;
  }, [amount, currentRate]);

  // Handle reverse currencies
  const handleReverse = () => {
    setFromCurrency(toCurrency);
    setToCurrency(fromCurrency);
  };

  // Toggle favorite
  const handleToggleFavorite = () => {
    const pair = { from: fromCurrency, to: toCurrency };
    if (isFavorite(pair)) {
      removeFavorite(pair);
    } else {
      addFavorite(pair);
      addToLastUsed(pair);
    }
  };

  // Multi-convert results
  const multiConvertResults = useMemo(() => {
    return multiConvertCurrencies.map((currency) => {
      const rate = exchangeRates.get(`${fromCurrency}${currency}`);
      return {
        currency,
        rate: rate?.bid || 0,
        amount: amount * (rate?.bid || 0),
      };
    });
  }, [fromCurrency, multiConvertCurrencies, amount, exchangeRates]);

  // Calculate pip value
  const pipValue = useMemo(() => {
    const pair = pipPair.replace('/', '');
    const baseCurrency = pair.substring(0, 3);
    const quoteCurrency = pair.substring(3, 6);

    // Pip size
    const pipSize = quoteCurrency === 'JPY' ? 0.01 : 0.0001;

    // Standard lot size (100,000 units)
    const standardLot = 100000;
    const units = lotSize * standardLot;

    // Pip value in quote currency
    let pipValueInQuote = units * pipSize;

    // Convert to account currency if needed
    if (quoteCurrency !== accountCurrency) {
      const conversionRate = exchangeRates.get(`${quoteCurrency}${accountCurrency}`);
      if (conversionRate) {
        pipValueInQuote *= conversionRate.bid;
      }
    }

    return pipValueInQuote;
  }, [lotSize, pipPair, accountCurrency, exchangeRates]);

  return (
    <div className="flex flex-col h-full bg-zinc-900 text-zinc-100">
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-2 bg-zinc-800 border-b border-zinc-700">
        <DollarSign className="w-4 h-4 text-emerald-400" />
        <span className="font-semibold text-sm">Currency Converter</span>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-3 gap-4">
          {/* Left Column - Converter */}
          <div className="col-span-2 space-y-4">
            {/* Main Converter */}
            <div className="bg-zinc-800 border border-zinc-700 rounded p-4">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-semibold text-zinc-300">Convert Currency</h3>
                <button
                  onClick={handleToggleFavorite}
                  className={`p-1 rounded transition-colors ${
                    isFavorite({ from: fromCurrency, to: toCurrency })
                      ? 'text-yellow-400'
                      : 'text-zinc-500 hover:text-yellow-400'
                  }`}
                >
                  <Star className="w-4 h-4" fill={isFavorite({ from: fromCurrency, to: toCurrency }) ? 'currentColor' : 'none'} />
                </button>
              </div>

              {/* Amount Input */}
              <div className="mb-4">
                <label className="text-xs text-zinc-400 block mb-1">Amount</label>
                <input
                  type="number"
                  value={amount}
                  onChange={(e) => setAmount(Number(e.target.value))}
                  className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-lg text-zinc-100"
                  min="0"
                  step="10"
                />
              </div>

              {/* From/To Selectors */}
              <div className="grid grid-cols-[1fr,auto,1fr] gap-2 mb-4">
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">From</label>
                  <select
                    value={fromCurrency}
                    onChange={(e) => {
                      setFromCurrency(e.target.value);
                      addToLastUsed({ from: e.target.value, to: toCurrency });
                    }}
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100"
                  >
                    {CURRENCIES.map((curr) => (
                      <option key={curr} value={curr}>{curr}</option>
                    ))}
                  </select>
                </div>

                <button
                  onClick={handleReverse}
                  className="self-end p-2 hover:bg-zinc-700 rounded transition-colors"
                  title="Reverse currencies"
                >
                  <ArrowLeftRight className="w-4 h-4 text-zinc-400" />
                </button>

                <div>
                  <label className="text-xs text-zinc-400 block mb-1">To</label>
                  <select
                    value={toCurrency}
                    onChange={(e) => {
                      setToCurrency(e.target.value);
                      addToLastUsed({ from: fromCurrency, to: e.target.value });
                    }}
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100"
                  >
                    {CURRENCIES.map((curr) => (
                      <option key={curr} value={curr}>{curr}</option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Conversion Result */}
              {currentRate && (
                <div className="bg-zinc-900 border border-zinc-700 rounded p-4">
                  <div className="text-xs text-zinc-400 mb-1">Converted Amount</div>
                  <div className="text-2xl font-bold text-emerald-400 mb-2">
                    {convertedAmount.toFixed(2)} {toCurrency}
                  </div>
                  <div className="flex items-center gap-4 text-xs text-zinc-500">
                    <div>Bid: {currentRate.bid.toFixed(6)}</div>
                    <div>Ask: {currentRate.ask.toFixed(6)}</div>
                    <div className={currentRate.change >= 0 ? 'text-emerald-400' : 'text-red-400'}>
                      {currentRate.change >= 0 ? '+' : ''}{currentRate.change.toFixed(2)}%
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Multi-Convert */}
            <div className="bg-zinc-800 border border-zinc-700 rounded p-4">
              <button
                onClick={() => setShowMultiConvert(!showMultiConvert)}
                className="w-full flex items-center justify-between mb-3"
              >
                <h3 className="text-sm font-semibold text-zinc-300">Multi-Currency Conversion</h3>
                {showMultiConvert ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
              </button>

              {showMultiConvert && (
                <div className="space-y-2">
                  {/* Currency selector */}
                  <div className="flex flex-wrap gap-2 mb-3">
                    {CURRENCIES.filter(c => c !== fromCurrency).map((currency) => (
                      <button
                        key={currency}
                        onClick={() => toggleMultiConvertCurrency(currency)}
                        className={`px-2 py-1 rounded text-xs transition-colors ${
                          multiConvertCurrencies.includes(currency)
                            ? 'bg-emerald-600 text-white'
                            : 'bg-zinc-700 text-zinc-300 hover:bg-zinc-600'
                        }`}
                      >
                        {currency}
                      </button>
                    ))}
                  </div>

                  {/* Results */}
                  <div className="space-y-2">
                    {multiConvertResults.map((result) => (
                      <div key={result.currency} className="flex items-center justify-between bg-zinc-900 p-2 rounded">
                        <span className="text-sm font-medium text-zinc-300">{result.currency}</span>
                        <div className="text-right">
                          <div className="text-sm font-bold text-zinc-100">{result.amount.toFixed(2)}</div>
                          <div className="text-xs text-zinc-500">@ {result.rate.toFixed(6)}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Cross-Rate Matrix */}
            <div className="bg-zinc-800 border border-zinc-700 rounded p-4">
              <h3 className="text-sm font-semibold text-zinc-300 mb-3 flex items-center gap-2">
                <TrendingUp className="w-4 h-4" />
                Cross-Rate Matrix
              </h3>
              <CrossRateMatrix rates={exchangeRates} />
            </div>

            {/* Favorites */}
            {favorites.length > 0 && (
              <div className="bg-zinc-800 border border-zinc-700 rounded p-4">
                <h3 className="text-sm font-semibold text-zinc-300 mb-3 flex items-center gap-2">
                  <Star className="w-4 h-4 text-yellow-400" />
                  Favorite Pairs
                </h3>
                <div className="grid grid-cols-3 gap-2">
                  {favorites.map((fav, i) => {
                    const rate = exchangeRates.get(`${fav.from}${fav.to}`);
                    return (
                      <button
                        key={i}
                        onClick={() => {
                          setFromCurrency(fav.from);
                          setToCurrency(fav.to);
                        }}
                        className="bg-zinc-900 border border-zinc-700 rounded p-2 hover:border-emerald-600 transition-colors text-left"
                      >
                        <div className="text-xs font-medium text-zinc-300">{fav.from}/{fav.to}</div>
                        {rate && (
                          <div className="text-xs text-zinc-500">{rate.bid.toFixed(6)}</div>
                        )}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}
          </div>

          {/* Right Column - Pip Calculator */}
          <div className="space-y-4">
            <div className="bg-zinc-800 border border-zinc-700 rounded p-4">
              <h3 className="text-sm font-semibold text-zinc-300 mb-3 flex items-center gap-2">
                <Calculator className="w-4 h-4" />
                Pip Value Calculator
              </h3>

              <div className="space-y-3">
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">Pair</label>
                  <select
                    value={pipPair}
                    onChange={(e) => setPipPair(e.target.value)}
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-sm text-zinc-100"
                  >
                    {['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'USDCHF', 'NZDUSD', 'EURJPY', 'GBPJPY', 'EURGBP'].map((pair) => (
                      <option key={pair} value={pair}>{pair}</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="text-xs text-zinc-400 block mb-1">Lot Size</label>
                  <input
                    type="number"
                    value={lotSize}
                    onChange={(e) => setLotSize(Number(e.target.value))}
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-sm text-zinc-100"
                    min="0.01"
                    step="0.01"
                  />
                </div>

                <div>
                  <label className="text-xs text-zinc-400 block mb-1">Account Currency</label>
                  <select
                    value={accountCurrency}
                    onChange={(e) => setAccountCurrency(e.target.value)}
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-sm text-zinc-100"
                  >
                    {MAJOR_CURRENCIES.map((curr) => (
                      <option key={curr} value={curr}>{curr}</option>
                    ))}
                  </select>
                </div>

                <div className="bg-zinc-900 border border-zinc-700 rounded p-3 mt-3">
                  <div className="text-xs text-zinc-400 mb-1">Pip Value</div>
                  <div className="text-xl font-bold text-emerald-400">
                    {pipValue.toFixed(2)} {accountCurrency}
                  </div>
                  <div className="text-xs text-zinc-500 mt-1">per pip movement</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

// Cross-Rate Matrix Component
interface CrossRateMatrixProps {
  rates: Map<string, ExchangeRate>;
}

const CrossRateMatrix: React.FC<CrossRateMatrixProps> = ({ rates }) => {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-xs">
        <thead>
          <tr>
            <th className="p-2 text-left text-zinc-400 border border-zinc-700"></th>
            {MAJOR_CURRENCIES.map((currency) => (
              <th key={currency} className="p-2 text-center text-zinc-400 border border-zinc-700 font-medium">
                {currency}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {MAJOR_CURRENCIES.map((fromCurr) => (
            <tr key={fromCurr}>
              <td className="p-2 text-zinc-400 border border-zinc-700 font-medium">{fromCurr}</td>
              {MAJOR_CURRENCIES.map((toCurr) => {
                if (fromCurr === toCurr) {
                  return (
                    <td key={toCurr} className="p-2 text-center border border-zinc-700 bg-zinc-900">
                      <span className="text-zinc-600">—</span>
                    </td>
                  );
                }

                const rate = rates.get(`${fromCurr}${toCurr}`);
                if (!rate) return <td key={toCurr} className="p-2 border border-zinc-700"></td>;

                const changeColor = rate.change >= 0 ? 'text-emerald-400' : 'text-red-400';
                const bgColor = rate.change >= 0 ? 'bg-emerald-900/10' : 'bg-red-900/10';

                return (
                  <td key={toCurr} className={`p-2 text-center border border-zinc-700 ${bgColor}`}>
                    <div className={`font-medium ${changeColor}`}>
                      {rate.bid.toFixed(fromCurr === 'JPY' || toCurr === 'JPY' ? 3 : 5)}
                    </div>
                    <div className="flex items-center justify-center gap-1 mt-1">
                      <Sparkline data={rate.history} width={40} height={12} />
                    </div>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

// Sparkline Chart Component
interface SparklineProps {
  data: number[];
  width: number;
  height: number;
}

const Sparkline: React.FC<SparklineProps> = ({ data, width, height }) => {
  if (data.length < 2) return null;

  const min = Math.min(...data);
  const max = Math.max(...data);
  const range = max - min;

  const points = data.map((value, index) => {
    const x = (index / (data.length - 1)) * width;
    const y = height - ((value - min) / range) * height;
    return `${x},${y}`;
  }).join(' ');

  const isUp = data[data.length - 1] > data[0];
  const color = isUp ? '#10b981' : '#ef4444';

  return (
    <svg width={width} height={height} className="inline-block">
      <polyline
        points={points}
        fill="none"
        stroke={color}
        strokeWidth="1"
      />
    </svg>
  );
};
