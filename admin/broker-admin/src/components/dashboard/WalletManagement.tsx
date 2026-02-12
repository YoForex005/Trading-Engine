'use client';

import React, { useState, useEffect } from 'react';
import { Search, DollarSign, TrendingUp, AlertTriangle, Plus, Minus, ArrowRightLeft, RefreshCw, CheckCircle, XCircle, Wallet, CreditCard } from 'lucide-react';
import { ADMIN_API_URL } from '../../config/api';

type WalletStatus = 'ACTIVE' | 'SUSPENDED' | 'CLOSED';
type TransactionType = 'DEPOSIT' | 'WITHDRAWAL' | 'TRANSFER' | 'ADJUSTMENT';
type ReconciliationStatus = 'OK' | 'WARNING' | 'CRITICAL';

interface ClientWallet {
  id: string;
  clientName: string;
  clientId: string;
  totalUsdValue: number;
  currencyCount: number;
  status: WalletStatus;
  balances: CurrencyBalance[];
}

interface CurrencyBalance {
  currency: string;
  amount: number;
  usdValue: number;
  exchangeRate: number;
}

interface CurrencyOverview {
  currency: string;
  symbol: string;
  exchangeRate: number;
  totalHeld: number;
  totalUsdValue: number;
  walletCount: number;
  color: string;
}

interface WalletTransaction {
  id: string;
  timestamp: string;
  type: TransactionType;
  clientId: string;
  clientName: string;
  amount: number;
  currency: string;
  usdValue: number;
  reference: string;
  reason?: string;
}

interface WalletStats {
  totalAUM: number;
  walletCount: number;
  avgBalance: number;
  largestWallet: number;
  largestWalletClient: string;
}

interface ReconciliationEntry {
  currency: string;
  expected: number;
  actual: number;
  discrepancy: number;
  status: ReconciliationStatus;
}

const CURRENCIES: CurrencyOverview[] = [
  { currency: 'USD', symbol: '$', exchangeRate: 1.0, totalHeld: 0, totalUsdValue: 0, walletCount: 0, color: '#10B981' },
  { currency: 'EUR', symbol: '€', exchangeRate: 1.08, totalHeld: 0, totalUsdValue: 0, walletCount: 0, color: '#3B82F6' },
  { currency: 'GBP', symbol: '£', exchangeRate: 1.27, totalHeld: 0, totalUsdValue: 0, walletCount: 0, color: '#F59E0B' },
  { currency: 'JPY', symbol: '¥', exchangeRate: 0.0067, totalHeld: 0, totalUsdValue: 0, walletCount: 0, color: '#EF4444' },
  { currency: 'AUD', symbol: 'A$', exchangeRate: 0.65, totalHeld: 0, totalUsdValue: 0, walletCount: 0, color: '#8B5CF6' },
  { currency: 'CAD', symbol: 'C$', exchangeRate: 0.73, totalHeld: 0, totalUsdValue: 0, walletCount: 0, color: '#EC4899' },
  { currency: 'CHF', symbol: 'CHF', exchangeRate: 1.13, totalHeld: 0, totalUsdValue: 0, walletCount: 0, color: '#14B8A6' },
  { currency: 'BTC', symbol: '₿', exchangeRate: 45000, totalHeld: 0, totalUsdValue: 0, walletCount: 0, color: '#F97316' },
];

const ADJUSTMENT_REASONS = [
  'Correction',
  'Bonus',
  'Refund',
  'Fee',
  'Interest',
  'Compensation',
  'Manual Override',
  'System Error Fix',
];

export default function WalletManagement() {
  const [wallets, setWallets] = useState<ClientWallet[]>([]);
  const [selectedWallet, setSelectedWallet] = useState<ClientWallet | null>(null);
  const [currencies, setCurrencies] = useState<CurrencyOverview[]>(CURRENCIES);
  const [transactions, setTransactions] = useState<WalletTransaction[]>([]);
  const [stats, setStats] = useState<WalletStats | null>(null);
  const [reconciliation, setReconciliation] = useState<ReconciliationEntry[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [showAdjustModal, setShowAdjustModal] = useState(false);
  const [showTransferModal, setShowTransferModal] = useState(false);

  // Adjustment modal state
  const [adjustCurrency, setAdjustCurrency] = useState('USD');
  const [adjustAmount, setAdjustAmount] = useState('');
  const [adjustReason, setAdjustReason] = useState('Correction');

  // Transfer modal state
  const [transferFrom, setTransferFrom] = useState('');
  const [transferTo, setTransferTo] = useState('');
  const [transferAmount, setTransferAmount] = useState('');
  const [transferCurrency, setTransferCurrency] = useState('USD');

  useEffect(() => {
    loadWallets();
    loadCurrencies();
    loadTransactions();
    loadStats();
    loadReconciliation();
  }, []);

  const loadWallets = () => {
    // Mock data - replace with fetch from ADMIN_API_URL
    const mockWallets: ClientWallet[] = Array.from({ length: 50 }, (_, i) => {
      const balances: CurrencyBalance[] = [];
      const activeCurrencies = Math.floor(Math.random() * 5) + 2; // 2-6 currencies per client
      const availableCurrencies = [...CURRENCIES];

      for (let j = 0; j < activeCurrencies; j++) {
        const currIndex = Math.floor(Math.random() * availableCurrencies.length);
        const curr = availableCurrencies[currIndex];
        availableCurrencies.splice(currIndex, 1);

        const amount = curr.currency === 'BTC'
          ? Math.random() * 2
          : curr.currency === 'JPY'
          ? Math.random() * 10000000
          : Math.random() * 100000 + 5000;

        const usdValue = amount * curr.exchangeRate;
        balances.push({
          currency: curr.currency,
          amount,
          usdValue,
          exchangeRate: curr.exchangeRate,
        });
      }

      const totalUsdValue = balances.reduce((sum, b) => sum + b.usdValue, 0);

      return {
        id: `wallet-${i + 1}`,
        clientName: `Client ${i + 1}`,
        clientId: `CL${String(i + 1).padStart(5, '0')}`,
        totalUsdValue,
        currencyCount: balances.length,
        status: ['ACTIVE', 'ACTIVE', 'ACTIVE', 'ACTIVE', 'SUSPENDED', 'CLOSED'][Math.floor(Math.random() * 6)] as WalletStatus,
        balances,
      };
    });
    setWallets(mockWallets);
  };

  const loadCurrencies = () => {
    // Calculate totals from wallets
    const currencyCopy = CURRENCIES.map(c => ({ ...c }));

    wallets.forEach(wallet => {
      wallet.balances.forEach(balance => {
        const curr = currencyCopy.find(c => c.currency === balance.currency);
        if (curr) {
          curr.totalHeld += balance.amount;
          curr.totalUsdValue += balance.usdValue;
          curr.walletCount++;
        }
      });
    });

    setCurrencies(currencyCopy);
  };

  const loadTransactions = () => {
    // Mock transaction data
    const types: TransactionType[] = ['DEPOSIT', 'WITHDRAWAL', 'TRANSFER', 'ADJUSTMENT'];
    const mockTransactions: WalletTransaction[] = Array.from({ length: 200 }, (_, i) => {
      const type = types[Math.floor(Math.random() * types.length)];
      const currency = CURRENCIES[Math.floor(Math.random() * CURRENCIES.length)];
      const amount = currency.currency === 'BTC'
        ? Math.random() * 0.5
        : currency.currency === 'JPY'
        ? Math.random() * 1000000
        : Math.random() * 10000 + 100;
      const usdValue = amount * currency.exchangeRate;

      return {
        id: `txn-${i + 1}`,
        timestamp: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
        type,
        clientId: `CL${String(Math.floor(Math.random() * 50) + 1).padStart(5, '0')}`,
        clientName: `Client ${Math.floor(Math.random() * 50) + 1}`,
        amount,
        currency: currency.currency,
        usdValue,
        reference: `REF${String(i + 1).padStart(8, '0')}`,
        reason: type === 'ADJUSTMENT' ? ADJUSTMENT_REASONS[Math.floor(Math.random() * ADJUSTMENT_REASONS.length)] : undefined,
      };
    });
    setTransactions(mockTransactions.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()));
  };

  const loadStats = () => {
    const totalAUM = wallets.reduce((sum, w) => sum + w.totalUsdValue, 0);
    const activeWallets = wallets.filter(w => w.status === 'ACTIVE');
    const avgBalance = activeWallets.length > 0 ? totalAUM / activeWallets.length : 0;
    const largest = wallets.reduce((max, w) => w.totalUsdValue > max.value ? { value: w.totalUsdValue, client: w.clientName } : max, { value: 0, client: '' });

    setStats({
      totalAUM,
      walletCount: activeWallets.length,
      avgBalance,
      largestWallet: largest.value,
      largestWalletClient: largest.client,
    });
  };

  const loadReconciliation = () => {
    const mockRecon: ReconciliationEntry[] = CURRENCIES.map(curr => {
      const expected = curr.totalHeld;
      const variance = (Math.random() - 0.5) * expected * 0.01; // ±0.5% variance
      const actual = expected + variance;
      const discrepancy = actual - expected;
      const discrepancyPercent = expected > 0 ? Math.abs(discrepancy / expected) * 100 : 0;

      let status: ReconciliationStatus = 'OK';
      if (discrepancyPercent > 1) status = 'CRITICAL';
      else if (discrepancyPercent > 0.1) status = 'WARNING';

      return {
        currency: curr.currency,
        expected,
        actual,
        discrepancy,
        status,
      };
    });
    setReconciliation(mockRecon);
  };

  useEffect(() => {
    if (wallets.length > 0) {
      loadCurrencies();
      loadStats();
      loadReconciliation();
    }
  }, [wallets]);

  const handleAdjustBalance = () => {
    if (!selectedWallet || !adjustAmount) return;

    const amount = parseFloat(adjustAmount);
    if (isNaN(amount)) return;

    // Mock adjustment - in production, would call API
    console.log('Adjusting balance:', {
      wallet: selectedWallet.id,
      currency: adjustCurrency,
      amount,
      reason: adjustReason,
    });

    // Update local state
    const updatedWallets = wallets.map(w => {
      if (w.id === selectedWallet.id) {
        const balanceIndex = w.balances.findIndex(b => b.currency === adjustCurrency);
        const newBalances = [...w.balances];

        if (balanceIndex >= 0) {
          newBalances[balanceIndex] = {
            ...newBalances[balanceIndex],
            amount: newBalances[balanceIndex].amount + amount,
            usdValue: (newBalances[balanceIndex].amount + amount) * newBalances[balanceIndex].exchangeRate,
          };
        } else {
          const curr = CURRENCIES.find(c => c.currency === adjustCurrency);
          if (curr) {
            newBalances.push({
              currency: adjustCurrency,
              amount: Math.max(0, amount),
              usdValue: Math.max(0, amount) * curr.exchangeRate,
              exchangeRate: curr.exchangeRate,
            });
          }
        }

        return {
          ...w,
          balances: newBalances,
          totalUsdValue: newBalances.reduce((sum, b) => sum + b.usdValue, 0),
          currencyCount: newBalances.length,
        };
      }
      return w;
    });

    setWallets(updatedWallets);
    setSelectedWallet(updatedWallets.find(w => w.id === selectedWallet.id) || null);
    setShowAdjustModal(false);
    setAdjustAmount('');
  };

  const handleInternalTransfer = () => {
    if (!transferFrom || !transferTo || !transferAmount) return;

    const amount = parseFloat(transferAmount);
    if (isNaN(amount) || amount <= 0) return;

    console.log('Processing transfer:', {
      from: transferFrom,
      to: transferTo,
      amount,
      currency: transferCurrency,
    });

    // Mock transfer - update local state
    const curr = CURRENCIES.find(c => c.currency === transferCurrency);
    if (!curr) return;

    const updatedWallets = wallets.map(w => {
      if (w.id === transferFrom) {
        // Deduct from sender
        const balanceIndex = w.balances.findIndex(b => b.currency === transferCurrency);
        if (balanceIndex >= 0) {
          const newBalances = [...w.balances];
          newBalances[balanceIndex] = {
            ...newBalances[balanceIndex],
            amount: Math.max(0, newBalances[balanceIndex].amount - amount),
            usdValue: Math.max(0, newBalances[balanceIndex].amount - amount) * curr.exchangeRate,
          };
          return {
            ...w,
            balances: newBalances,
            totalUsdValue: newBalances.reduce((sum, b) => sum + b.usdValue, 0),
          };
        }
      } else if (w.id === transferTo) {
        // Add to receiver
        const balanceIndex = w.balances.findIndex(b => b.currency === transferCurrency);
        const newBalances = [...w.balances];

        if (balanceIndex >= 0) {
          newBalances[balanceIndex] = {
            ...newBalances[balanceIndex],
            amount: newBalances[balanceIndex].amount + amount,
            usdValue: (newBalances[balanceIndex].amount + amount) * curr.exchangeRate,
          };
        } else {
          newBalances.push({
            currency: transferCurrency,
            amount,
            usdValue: amount * curr.exchangeRate,
            exchangeRate: curr.exchangeRate,
          });
        }

        return {
          ...w,
          balances: newBalances,
          totalUsdValue: newBalances.reduce((sum, b) => sum + b.usdValue, 0),
          currencyCount: newBalances.length,
        };
      }
      return w;
    });

    setWallets(updatedWallets);
    setShowTransferModal(false);
    setTransferFrom('');
    setTransferTo('');
    setTransferAmount('');
  };

  const filteredWallets = wallets.filter(w =>
    w.clientName.toLowerCase().includes(searchTerm.toLowerCase()) ||
    w.clientId.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getStatusBadge = (status: WalletStatus) => {
    const colors = {
      ACTIVE: 'bg-[#10B981] text-white',
      SUSPENDED: 'bg-[#F59E0B] text-white',
      CLOSED: 'bg-[#6B7280] text-white',
    };
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[status]}`}>
        {status}
      </span>
    );
  };

  const getTransactionTypeBadge = (type: TransactionType) => {
    const colors = {
      DEPOSIT: 'bg-[#10B981] text-white',
      WITHDRAWAL: 'bg-[#EF4444] text-white',
      TRANSFER: 'bg-[#3B82F6] text-white',
      ADJUSTMENT: 'bg-[#F59E0B] text-white',
    };
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[type]}`}>
        {type}
      </span>
    );
  };

  const getReconciliationBadge = (status: ReconciliationStatus) => {
    const colors = {
      OK: 'bg-[#10B981] text-white',
      WARNING: 'bg-[#F59E0B] text-white',
      CRITICAL: 'bg-[#EF4444] text-white',
    };
    const icons = {
      OK: <CheckCircle size={14} />,
      WARNING: <AlertTriangle size={14} />,
      CRITICAL: <XCircle size={14} />,
    };
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium flex items-center gap-1 ${colors[status]}`}>
        {icons[status]}
        {status}
      </span>
    );
  };

  const renderPieChart = () => {
    const total = currencies.reduce((sum, c) => sum + c.totalUsdValue, 0);
    if (total === 0) return null;

    let currentAngle = 0;
    const radius = 80;
    const centerX = 100;
    const centerY = 100;

    return (
      <svg width="200" height="200" className="mx-auto">
        {currencies.map((curr, index) => {
          const percentage = (curr.totalUsdValue / total) * 100;
          const sliceAngle = (percentage / 100) * 2 * Math.PI;

          const x1 = centerX + radius * Math.cos(currentAngle);
          const y1 = centerY + radius * Math.sin(currentAngle);

          currentAngle += sliceAngle;

          const x2 = centerX + radius * Math.cos(currentAngle);
          const y2 = centerY + radius * Math.sin(currentAngle);

          const largeArcFlag = sliceAngle > Math.PI ? 1 : 0;

          const pathData = [
            `M ${centerX} ${centerY}`,
            `L ${x1} ${y1}`,
            `A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2}`,
            'Z'
          ].join(' ');

          return (
            <g key={curr.currency}>
              <path
                d={pathData}
                fill={curr.color}
                stroke="#121316"
                strokeWidth="2"
              />
            </g>
          );
        })}
        <circle cx={centerX} cy={centerY} r={radius * 0.5} fill="#121316" />
      </svg>
    );
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-[#E0E0E0]">
      {/* Header */}
      <div className="h-12 bg-[#1E2026] border-b border-[#383A42] flex items-center justify-between px-4 flex-shrink-0">
        <div className="flex items-center gap-3">
          <Wallet size={20} className="text-[#10B981]" />
          <span className="font-bold text-sm">Multi-Currency Wallet Management</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowTransferModal(true)}
            className="px-3 py-1 text-xs rounded bg-[#3B82F6] text-white hover:bg-[#2563EB] flex items-center gap-1"
          >
            <ArrowRightLeft size={14} />
            Internal Transfer
          </button>
          <button
            onClick={() => loadReconciliation()}
            className="px-3 py-1 text-xs rounded bg-[#10B981] text-white hover:bg-[#059669] flex items-center gap-1"
          >
            <RefreshCw size={14} />
            Reconcile
          </button>
        </div>
      </div>

      {/* AUM Stats Cards */}
      {stats && (
        <div className="grid grid-cols-4 gap-3 p-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
          <div className="bg-[#121316] border border-[#383A42] rounded p-3">
            <div className="flex items-center gap-2 mb-1">
              <DollarSign size={16} className="text-[#10B981]" />
              <span className="text-xs text-[#888]">Total AUM</span>
            </div>
            <div className="text-xl font-bold text-[#E0E0E0]">${stats.totalAUM.toLocaleString(undefined, { maximumFractionDigits: 0 })}</div>
            <div className="text-xs text-[#888] mt-1">across all wallets</div>
          </div>
          <div className="bg-[#121316] border border-[#383A42] rounded p-3">
            <div className="flex items-center gap-2 mb-1">
              <Wallet size={16} className="text-[#3B82F6]" />
              <span className="text-xs text-[#888]">Active Wallets</span>
            </div>
            <div className="text-xl font-bold text-[#E0E0E0]">{stats.walletCount}</div>
            <div className="text-xs text-[#888] mt-1">of {wallets.length} total</div>
          </div>
          <div className="bg-[#121316] border border-[#383A42] rounded p-3">
            <div className="flex items-center gap-2 mb-1">
              <TrendingUp size={16} className="text-[#F59E0B]" />
              <span className="text-xs text-[#888]">Avg Balance</span>
            </div>
            <div className="text-xl font-bold text-[#E0E0E0]">${stats.avgBalance.toLocaleString(undefined, { maximumFractionDigits: 0 })}</div>
            <div className="text-xs text-[#888] mt-1">per active wallet</div>
          </div>
          <div className="bg-[#121316] border border-[#383A42] rounded p-3">
            <div className="flex items-center gap-2 mb-1">
              <CreditCard size={16} className="text-[#8B5CF6]" />
              <span className="text-xs text-[#888]">Largest Wallet</span>
            </div>
            <div className="text-xl font-bold text-[#E0E0E0]">${stats.largestWallet.toLocaleString(undefined, { maximumFractionDigits: 0 })}</div>
            <div className="text-xs text-[#888] mt-1">{stats.largestWalletClient}</div>
          </div>
        </div>
      )}

      {/* Currency Overview Cards */}
      <div className="grid grid-cols-8 gap-2 p-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
        {currencies.map(curr => (
          <div key={curr.currency} className="bg-[#121316] border border-[#383A42] rounded p-2">
            <div className="flex items-center justify-between mb-1">
              <span className="font-bold text-xs" style={{ color: curr.color }}>{curr.currency}</span>
              <span className="text-[10px] text-[#888]">{curr.symbol}</span>
            </div>
            <div className="text-xs text-[#E0E0E0] mb-1">
              {curr.currency === 'BTC' ? curr.totalHeld.toFixed(4) : curr.totalHeld.toLocaleString(undefined, { maximumFractionDigits: 0 })}
            </div>
            <div className="text-[10px] text-[#888]">
              ${curr.totalUsdValue.toLocaleString(undefined, { maximumFractionDigits: 0 })}
            </div>
            <div className="text-[10px] text-[#888] mt-1">{curr.walletCount} wallets</div>
          </div>
        ))}
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-hidden flex">
        {/* Left Panel - Wallet List */}
        <div className="w-1/3 flex flex-col border-r border-[#383A42]">
          {/* Search */}
          <div className="p-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
            <div className="relative">
              <Search size={16} className="absolute left-2 top-1/2 -translate-y-1/2 text-[#888]" />
              <input
                type="text"
                placeholder="Search clients..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-8 pr-3 py-1.5 bg-[#121316] border border-[#383A42] rounded text-xs text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
              />
            </div>
          </div>

          {/* Wallet Table */}
          <div className="flex-1 overflow-auto">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                <tr>
                  <th className="text-left p-2 font-medium text-[#888]">Client</th>
                  <th className="text-right p-2 font-medium text-[#888]">USD Value</th>
                  <th className="text-center p-2 font-medium text-[#888]">Currencies</th>
                  <th className="text-center p-2 font-medium text-[#888]">Status</th>
                </tr>
              </thead>
              <tbody>
                {filteredWallets.map(wallet => (
                  <tr
                    key={wallet.id}
                    onClick={() => setSelectedWallet(wallet)}
                    className={`border-b border-[#383A42] hover:bg-[#1E2026] cursor-pointer ${selectedWallet?.id === wallet.id ? 'bg-[#1E2026]' : ''}`}
                  >
                    <td className="p-2">
                      <div className="font-medium text-[#E0E0E0]">{wallet.clientName}</div>
                      <div className="text-[10px] text-[#888]">{wallet.clientId}</div>
                    </td>
                    <td className="p-2 text-right text-[#E0E0E0]">
                      ${wallet.totalUsdValue.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                    </td>
                    <td className="p-2 text-center">
                      <span className="px-2 py-0.5 rounded bg-[#383A42] text-[#E0E0E0]">
                        {wallet.currencyCount}
                      </span>
                    </td>
                    <td className="p-2 text-center">{getStatusBadge(wallet.status)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Center Panel - Wallet Detail & Reconciliation */}
        <div className="w-1/3 flex flex-col border-r border-[#383A42]">
          {selectedWallet ? (
            <>
              {/* Wallet Detail */}
              <div className="p-4 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h3 className="font-bold text-sm mb-1">{selectedWallet.clientName}</h3>
                    <div className="text-xs text-[#888]">{selectedWallet.clientId}</div>
                  </div>
                  <button
                    onClick={() => setShowAdjustModal(true)}
                    className="px-2 py-1 text-xs rounded bg-[#10B981] text-white hover:bg-[#059669] flex items-center gap-1"
                  >
                    <Plus size={12} />
                    Adjust
                  </button>
                </div>
                <div className="text-lg font-bold text-[#10B981] mb-3">
                  ${selectedWallet.totalUsdValue.toLocaleString(undefined, { maximumFractionDigits: 2 })} USD
                </div>
              </div>

              {/* Currency Balances */}
              <div className="flex-1 overflow-auto p-4">
                <h4 className="font-medium text-xs mb-2 text-[#888]">Currency Balances</h4>
                <div className="space-y-2">
                  {selectedWallet.balances.map(balance => {
                    const curr = CURRENCIES.find(c => c.currency === balance.currency);
                    return (
                      <div key={balance.currency} className="bg-[#121316] border border-[#383A42] rounded p-3">
                        <div className="flex items-center justify-between mb-2">
                          <span className="font-bold text-sm" style={{ color: curr?.color || '#E0E0E0' }}>
                            {balance.currency}
                          </span>
                          <span className="text-xs text-[#888]">
                            Rate: {curr?.symbol}{balance.exchangeRate.toLocaleString()}
                          </span>
                        </div>
                        <div className="text-lg font-bold text-[#E0E0E0]">
                          {balance.currency === 'BTC'
                            ? balance.amount.toFixed(6)
                            : balance.amount.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                        </div>
                        <div className="text-xs text-[#888] mt-1">
                          ≈ ${balance.usdValue.toLocaleString(undefined, { maximumFractionDigits: 2 })} USD
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-[#888]">
              <div className="text-center">
                <Wallet size={48} className="mx-auto mb-2 opacity-20" />
                <div className="text-sm">Select a wallet to view details</div>
              </div>
            </div>
          )}
        </div>

        {/* Right Panel - Transactions & Reconciliation */}
        <div className="flex-1 flex flex-col">
          {/* Currency Distribution */}
          <div className="p-4 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
            <h4 className="font-medium text-xs mb-3 text-[#888]">Currency Distribution</h4>
            {renderPieChart()}
            <div className="grid grid-cols-4 gap-2 mt-3">
              {currencies.slice(0, 8).map(curr => (
                <div key={curr.currency} className="flex items-center gap-1">
                  <div className="w-3 h-3 rounded" style={{ backgroundColor: curr.color }} />
                  <span className="text-[10px] text-[#888]">{curr.currency}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Transaction Log */}
          <div className="flex-1 flex flex-col">
            <div className="p-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
              <h4 className="font-medium text-xs text-[#888]">Recent Transactions (Last 50)</h4>
            </div>
            <div className="flex-1 overflow-auto">
              <table className="w-full text-xs">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                  <tr>
                    <th className="text-left p-2 font-medium text-[#888]">Date</th>
                    <th className="text-center p-2 font-medium text-[#888]">Type</th>
                    <th className="text-left p-2 font-medium text-[#888]">Client</th>
                    <th className="text-right p-2 font-medium text-[#888]">Amount</th>
                    <th className="text-right p-2 font-medium text-[#888]">USD Value</th>
                  </tr>
                </thead>
                <tbody>
                  {transactions.slice(0, 50).map(txn => (
                    <tr key={txn.id} className="border-b border-[#383A42] hover:bg-[#1E2026]">
                      <td className="p-2 text-[#888]">
                        {new Date(txn.timestamp).toLocaleDateString()}
                      </td>
                      <td className="p-2 text-center">{getTransactionTypeBadge(txn.type)}</td>
                      <td className="p-2 text-[#E0E0E0]">{txn.clientName}</td>
                      <td className="p-2 text-right">
                        <span className={txn.type === 'WITHDRAWAL' ? 'text-[#EF4444]' : 'text-[#10B981]'}>
                          {txn.type === 'WITHDRAWAL' ? '-' : '+'}
                          {txn.currency === 'BTC' ? txn.amount.toFixed(6) : txn.amount.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                          {' '}{txn.currency}
                        </span>
                      </td>
                      <td className="p-2 text-right text-[#888]">
                        ${txn.usdValue.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Reconciliation Panel */}
          <div className="h-1/3 border-t border-[#383A42] flex flex-col">
            <div className="p-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
              <h4 className="font-medium text-xs text-[#888]">Reconciliation Status</h4>
            </div>
            <div className="flex-1 overflow-auto">
              <table className="w-full text-xs">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                  <tr>
                    <th className="text-left p-2 font-medium text-[#888]">Currency</th>
                    <th className="text-right p-2 font-medium text-[#888]">Expected</th>
                    <th className="text-right p-2 font-medium text-[#888]">Actual</th>
                    <th className="text-right p-2 font-medium text-[#888]">Discrepancy</th>
                    <th className="text-center p-2 font-medium text-[#888]">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {reconciliation.map(entry => (
                    <tr key={entry.currency} className="border-b border-[#383A42] hover:bg-[#1E2026]">
                      <td className="p-2 font-medium text-[#E0E0E0]">{entry.currency}</td>
                      <td className="p-2 text-right text-[#888]">
                        {entry.currency === 'BTC' ? entry.expected.toFixed(6) : entry.expected.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                      </td>
                      <td className="p-2 text-right text-[#E0E0E0]">
                        {entry.currency === 'BTC' ? entry.actual.toFixed(6) : entry.actual.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                      </td>
                      <td className="p-2 text-right">
                        <span className={entry.discrepancy >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}>
                          {entry.discrepancy >= 0 ? '+' : ''}
                          {entry.currency === 'BTC' ? entry.discrepancy.toFixed(6) : entry.discrepancy.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                        </span>
                      </td>
                      <td className="p-2 text-center">{getReconciliationBadge(entry.status)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      {/* Adjust Balance Modal */}
      {showAdjustModal && selectedWallet && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-6 w-96">
            <h3 className="text-lg font-bold mb-4 text-[#E0E0E0]">Adjust Balance</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-xs text-[#888] mb-1">Client</label>
                <div className="text-sm text-[#E0E0E0]">{selectedWallet.clientName}</div>
              </div>
              <div>
                <label className="block text-xs text-[#888] mb-1">Currency</label>
                <select
                  value={adjustCurrency}
                  onChange={(e) => setAdjustCurrency(e.target.value)}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
                >
                  {CURRENCIES.map(curr => (
                    <option key={curr.currency} value={curr.currency}>{curr.currency}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs text-[#888] mb-1">Amount (+/-)</label>
                <input
                  type="number"
                  step="0.01"
                  value={adjustAmount}
                  onChange={(e) => setAdjustAmount(e.target.value)}
                  placeholder="Enter amount (use - for deduction)"
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
                />
              </div>
              <div>
                <label className="block text-xs text-[#888] mb-1">Reason</label>
                <select
                  value={adjustReason}
                  onChange={(e) => setAdjustReason(e.target.value)}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
                >
                  {ADJUSTMENT_REASONS.map(reason => (
                    <option key={reason} value={reason}>{reason}</option>
                  ))}
                </select>
              </div>
            </div>
            <div className="flex gap-2 mt-6">
              <button
                onClick={handleAdjustBalance}
                className="flex-1 px-4 py-2 bg-[#10B981] text-white rounded hover:bg-[#059669]"
              >
                Confirm
              </button>
              <button
                onClick={() => {
                  setShowAdjustModal(false);
                  setAdjustAmount('');
                }}
                className="flex-1 px-4 py-2 bg-[#6B7280] text-white rounded hover:bg-[#4B5563]"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Internal Transfer Modal */}
      {showTransferModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-6 w-96">
            <h3 className="text-lg font-bold mb-4 text-[#E0E0E0]">Internal Transfer</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-xs text-[#888] mb-1">From Account</label>
                <select
                  value={transferFrom}
                  onChange={(e) => setTransferFrom(e.target.value)}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
                >
                  <option value="">Select account...</option>
                  {wallets.filter(w => w.status === 'ACTIVE').map(w => (
                    <option key={w.id} value={w.id}>{w.clientName} ({w.clientId})</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs text-[#888] mb-1">To Account</label>
                <select
                  value={transferTo}
                  onChange={(e) => setTransferTo(e.target.value)}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
                >
                  <option value="">Select account...</option>
                  {wallets.filter(w => w.status === 'ACTIVE' && w.id !== transferFrom).map(w => (
                    <option key={w.id} value={w.id}>{w.clientName} ({w.clientId})</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs text-[#888] mb-1">Currency</label>
                <select
                  value={transferCurrency}
                  onChange={(e) => setTransferCurrency(e.target.value)}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
                >
                  {CURRENCIES.map(curr => (
                    <option key={curr.currency} value={curr.currency}>{curr.currency}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs text-[#888] mb-1">Amount</label>
                <input
                  type="number"
                  step="0.01"
                  value={transferAmount}
                  onChange={(e) => setTransferAmount(e.target.value)}
                  placeholder="Enter amount"
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-[#E0E0E0] focus:outline-none focus:border-[#10B981]"
                />
              </div>
            </div>
            <div className="flex gap-2 mt-6">
              <button
                onClick={handleInternalTransfer}
                className="flex-1 px-4 py-2 bg-[#3B82F6] text-white rounded hover:bg-[#2563EB]"
              >
                Transfer
              </button>
              <button
                onClick={() => {
                  setShowTransferModal(false);
                  setTransferFrom('');
                  setTransferTo('');
                  setTransferAmount('');
                }}
                className="flex-1 px-4 py-2 bg-[#6B7280] text-white rounded hover:bg-[#4B5563]"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
