'use client';

import { useState, useMemo } from 'react';
import {
  CreditCard,
  Bitcoin,
  Wallet,
  Building2,
  Globe,
  CheckCircle,
  XCircle,
  Edit2,
  Settings,
  Plus,
  DollarSign,
  TrendingUp,
  Clock,
  Eye,
  EyeOff,
  Save,
  X
} from 'lucide-react';

// Types
type PaymentMethodType = 'bank_wire' | 'credit_card' | 'crypto' | 'ewallet' | 'local_transfer';
type GatewayStatus = 'enabled' | 'disabled';
type TransactionStatus = 'pending' | 'processing' | 'completed' | 'failed';

interface PaymentMethod {
  id: string;
  type: PaymentMethodType;
  name: string;
  description: string;
  status: GatewayStatus;
  transactionCount: number;
  volume: number;
  fees: number; // percentage
  icon: React.ReactNode;
  color: string;
}

interface GatewayConfig {
  methodId: string;
  apiKey: string;
  secretKey: string;
  webhookUrl: string;
  isSandbox: boolean;
  minDeposit: number;
  maxDeposit: number;
  minWithdrawal: number;
  maxWithdrawal: number;
  processingFee: number;
  supportedCurrencies: string[];
}

interface TransactionLimit {
  method: string;
  dailyDepositLimit: number;
  monthlyDepositLimit: number;
  dailyWithdrawalLimit: number;
  monthlyWithdrawalLimit: number;
}

interface FeeSchedule {
  method: string;
  depositFeePercent: number;
  depositFeeFixed: number;
  withdrawalFeePercent: number;
  withdrawalFeeFixed: number;
}

interface PendingTransaction {
  id: string;
  type: 'deposit' | 'withdrawal';
  method: string;
  client: string;
  amount: number;
  currency: string;
  status: TransactionStatus;
  timestamp: Date;
}

// Mock data generators
const generatePaymentMethods = (): PaymentMethod[] => {
  return [
    {
      id: 'PM-001',
      type: 'bank_wire',
      name: 'Bank Wire Transfer',
      description: 'Traditional bank wire transfers',
      status: 'enabled',
      transactionCount: 342,
      volume: 1245678,
      fees: 0.5,
      icon: <Building2 size={24} />,
      color: 'text-blue-400',
    },
    {
      id: 'PM-002',
      type: 'credit_card',
      name: 'Credit/Debit Card',
      description: 'Visa, Mastercard, Amex',
      status: 'enabled',
      transactionCount: 1456,
      volume: 456890,
      fees: 2.5,
      icon: <CreditCard size={24} />,
      color: 'text-purple-400',
    },
    {
      id: 'PM-003',
      type: 'crypto',
      name: 'Cryptocurrency',
      description: 'BTC, ETH, USDT',
      status: 'enabled',
      transactionCount: 567,
      volume: 789345,
      fees: 1.0,
      icon: <Bitcoin size={24} />,
      color: 'text-orange-400',
    },
    {
      id: 'PM-004',
      type: 'ewallet',
      name: 'E-Wallet',
      description: 'Skrill, Neteller, PayPal',
      status: 'enabled',
      transactionCount: 892,
      volume: 234567,
      fees: 1.5,
      icon: <Wallet size={24} />,
      color: 'text-green-400',
    },
    {
      id: 'PM-005',
      type: 'local_transfer',
      name: 'Local Bank Transfer',
      description: 'Regional banking networks',
      status: 'disabled',
      transactionCount: 145,
      volume: 123456,
      fees: 0.3,
      icon: <Globe size={24} />,
      color: 'text-yellow-400',
    },
  ];
};

const generateGatewayConfig = (methodId: string): GatewayConfig => {
  return {
    methodId,
    apiKey: 'sk_live_' + Math.random().toString(36).substring(2, 15),
    secretKey: 'sk_secret_' + Math.random().toString(36).substring(2, 15),
    webhookUrl: 'https://api.rtx5.com/webhooks/payments',
    isSandbox: false,
    minDeposit: 100,
    maxDeposit: 100000,
    minWithdrawal: 50,
    maxWithdrawal: 50000,
    processingFee: 2.5,
    supportedCurrencies: ['USD', 'EUR', 'GBP', 'JPY', 'AUD'],
  };
};

const generateTransactionLimits = (): TransactionLimit[] => {
  const methods = ['Bank Wire', 'Credit Card', 'Cryptocurrency', 'E-Wallet', 'Local Transfer'];
  return methods.map((method) => ({
    method,
    dailyDepositLimit: Math.floor(Math.random() * 50000 + 50000),
    monthlyDepositLimit: Math.floor(Math.random() * 500000 + 1000000),
    dailyWithdrawalLimit: Math.floor(Math.random() * 30000 + 30000),
    monthlyWithdrawalLimit: Math.floor(Math.random() * 300000 + 500000),
  }));
};

const generateFeeSchedule = (): FeeSchedule[] => {
  return [
    { method: 'Bank Wire', depositFeePercent: 0.5, depositFeeFixed: 25, withdrawalFeePercent: 0.5, withdrawalFeeFixed: 25 },
    { method: 'Credit Card', depositFeePercent: 2.5, depositFeeFixed: 0, withdrawalFeePercent: 3.0, withdrawalFeeFixed: 0 },
    { method: 'Cryptocurrency', depositFeePercent: 1.0, depositFeeFixed: 0, withdrawalFeePercent: 1.0, withdrawalFeeFixed: 0 },
    { method: 'E-Wallet', depositFeePercent: 1.5, depositFeeFixed: 0, withdrawalFeePercent: 2.0, withdrawalFeeFixed: 0 },
    { method: 'Local Transfer', depositFeePercent: 0.3, depositFeeFixed: 10, withdrawalFeePercent: 0.3, withdrawalFeeFixed: 10 },
  ];
};

const generatePendingTransactions = (): PendingTransaction[] => {
  const methods = ['Bank Wire', 'Credit Card', 'Crypto', 'E-Wallet', 'Local Transfer'];
  const clients = ['John Smith', 'Emma Wilson', 'Michael Chen', 'Sarah Johnson', 'David Brown'];
  const statuses: TransactionStatus[] = ['pending', 'processing'];
  const transactions: PendingTransaction[] = [];
  const now = new Date();

  for (let i = 0; i < 20; i++) {
    transactions.push({
      id: `TXN-${String(i + 1).padStart(5, '0')}`,
      type: Math.random() > 0.5 ? 'deposit' : 'withdrawal',
      method: methods[Math.floor(Math.random() * methods.length)],
      client: clients[Math.floor(Math.random() * clients.length)],
      amount: Math.floor(Math.random() * 10000 + 100),
      currency: 'USD',
      status: statuses[Math.floor(Math.random() * statuses.length)],
      timestamp: new Date(now.getTime() - Math.random() * 3600000),
    });
  }

  return transactions.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
};

export default function PaymentGatewayConfig() {
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethod[]>(generatePaymentMethods());
  const [transactionLimits] = useState<TransactionLimit[]>(generateTransactionLimits());
  const [feeSchedule, setFeeSchedule] = useState<FeeSchedule[]>(generateFeeSchedule());
  const [pendingTransactions] = useState<PendingTransaction[]>(generatePendingTransactions());

  const [isConfigModalOpen, setIsConfigModalOpen] = useState(false);
  const [selectedMethod, setSelectedMethod] = useState<PaymentMethod | null>(null);
  const [gatewayConfig, setGatewayConfig] = useState<GatewayConfig | null>(null);
  const [showApiKey, setShowApiKey] = useState(false);
  const [showSecretKey, setShowSecretKey] = useState(false);

  // Summary metrics
  const metrics = useMemo(() => {
    const activeGateways = paymentMethods.filter((m) => m.status === 'enabled').length;
    const depositsToday = pendingTransactions
      .filter((t) => t.type === 'deposit')
      .reduce((sum, t) => sum + t.amount, 0);
    const withdrawalsToday = pendingTransactions
      .filter((t) => t.type === 'withdrawal')
      .reduce((sum, t) => sum + t.amount, 0);
    const feesCollected = paymentMethods.reduce((sum, m) => sum + (m.volume * m.fees) / 100, 0);

    return { activeGateways, depositsToday, withdrawalsToday, feesCollected };
  }, [paymentMethods, pendingTransactions]);

  const handleConfigureGateway = (method: PaymentMethod) => {
    setSelectedMethod(method);
    setGatewayConfig(generateGatewayConfig(method.id));
    setIsConfigModalOpen(true);
    setShowApiKey(false);
    setShowSecretKey(false);
  };

  const handleToggleStatus = (methodId: string) => {
    setPaymentMethods((prev) =>
      prev.map((m) =>
        m.id === methodId
          ? { ...m, status: m.status === 'enabled' ? 'disabled' : 'enabled' }
          : m
      )
    );
  };

  const handleSaveConfig = () => {
    // In real app, would save to backend
    setIsConfigModalOpen(false);
  };

  const maskKey = (key: string, show: boolean) => {
    if (show) return key;
    return key.substring(0, 10) + '••••••••••••••••' + key.substring(key.length - 4);
  };

  const formatTime = (date: Date) => {
    const now = new Date();
    const diff = Math.floor((now.getTime() - date.getTime()) / 1000 / 60);
    if (diff < 1) return 'Just now';
    if (diff < 60) return `${diff}m ago`;
    return `${Math.floor(diff / 60)}h ago`;
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#CCC]">
      {/* Header */}
      <div className="flex-shrink-0 border-b border-[#383A42] bg-[#1E2026] px-4 py-3">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-lg font-bold text-white">Payment Gateway Configuration</h1>
            <p className="text-xs text-gray-400 mt-0.5">Manage payment processor integrations and settings</p>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4 space-y-4">
        {/* Summary Cards */}
        <div className="grid grid-cols-4 gap-4">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-gray-400">Active Gateways</span>
              <CheckCircle size={16} className="text-green-400" />
            </div>
            <div className="text-2xl font-bold text-white">{metrics.activeGateways}</div>
            <div className="text-xs text-gray-500 mt-1">of {paymentMethods.length} total</div>
          </div>
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-gray-400">Deposits Today</span>
              <TrendingUp size={16} className="text-blue-400" />
            </div>
            <div className="text-2xl font-bold text-white">
              ${metrics.depositsToday.toLocaleString()}
            </div>
            <div className="text-xs text-gray-500 mt-1">from pending queue</div>
          </div>
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-gray-400">Withdrawals Today</span>
              <DollarSign size={16} className="text-orange-400" />
            </div>
            <div className="text-2xl font-bold text-white">
              ${metrics.withdrawalsToday.toLocaleString()}
            </div>
            <div className="text-xs text-gray-500 mt-1">from pending queue</div>
          </div>
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-gray-400">Processing Fees</span>
              <DollarSign size={16} className="text-green-400" />
            </div>
            <div className="text-2xl font-bold text-white">
              ${metrics.feesCollected.toLocaleString()}
            </div>
            <div className="text-xs text-gray-500 mt-1">total collected</div>
          </div>
        </div>

        {/* Payment Method Cards */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
          <h3 className="text-sm font-semibold text-white mb-4">Payment Methods</h3>
          <div className="grid grid-cols-5 gap-4">
            {paymentMethods.map((method) => (
              <div
                key={method.id}
                className={`bg-[#1A1C20] border ${
                  method.status === 'enabled' ? 'border-green-500/30' : 'border-[#383A42]'
                } rounded-lg p-4 hover:border-blue-500/50 transition-colors`}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className={method.color}>{method.icon}</div>
                  <div
                    onClick={() => handleToggleStatus(method.id)}
                    className="cursor-pointer"
                  >
                    {method.status === 'enabled' ? (
                      <CheckCircle size={18} className="text-green-400" />
                    ) : (
                      <XCircle size={18} className="text-gray-500" />
                    )}
                  </div>
                </div>
                <div className="text-sm font-semibold text-white mb-1">{method.name}</div>
                <div className="text-xs text-gray-400 mb-3">{method.description}</div>
                <div className="space-y-1.5 text-xs mb-3">
                  <div className="flex justify-between">
                    <span className="text-gray-500">Transactions</span>
                    <span className="text-white font-medium">{method.transactionCount}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Volume</span>
                    <span className="text-white font-medium">
                      ${(method.volume / 1000).toFixed(0)}K
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Fee Rate</span>
                    <span className="text-green-400 font-medium">{method.fees}%</span>
                  </div>
                </div>
                <button
                  onClick={() => handleConfigureGateway(method)}
                  className="w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium rounded transition-colors flex items-center justify-center gap-1.5"
                >
                  <Settings size={12} />
                  Configure
                </button>
              </div>
            ))}
          </div>
        </div>

        {/* Transaction Limits and Fee Schedule */}
        <div className="grid grid-cols-2 gap-4">
          {/* Transaction Limits */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <h3 className="text-sm font-semibold text-white mb-3">Transaction Limits</h3>
            <div className="overflow-auto max-h-80">
              <table className="w-full text-xs">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                  <tr>
                    <th className="text-left py-2 text-gray-400 font-medium">Method</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Daily Dep</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Monthly Dep</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Daily Wd</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Monthly Wd</th>
                  </tr>
                </thead>
                <tbody>
                  {transactionLimits.map((limit, idx) => (
                    <tr key={idx} className="border-b border-[#2A2C34] hover:bg-[#25272E]">
                      <td className="py-2 text-gray-200">{limit.method}</td>
                      <td className="text-right text-white">
                        ${(limit.dailyDepositLimit / 1000).toFixed(0)}K
                      </td>
                      <td className="text-right text-white">
                        ${(limit.monthlyDepositLimit / 1000).toFixed(0)}K
                      </td>
                      <td className="text-right text-white">
                        ${(limit.dailyWithdrawalLimit / 1000).toFixed(0)}K
                      </td>
                      <td className="text-right text-white">
                        ${(limit.monthlyWithdrawalLimit / 1000).toFixed(0)}K
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Fee Schedule */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
            <h3 className="text-sm font-semibold text-white mb-3">Fee Schedule</h3>
            <div className="overflow-auto max-h-80">
              <table className="w-full text-xs">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                  <tr>
                    <th className="text-left py-2 text-gray-400 font-medium">Method</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Dep %</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Dep Fixed</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Wd %</th>
                    <th className="text-right py-2 text-gray-400 font-medium">Wd Fixed</th>
                  </tr>
                </thead>
                <tbody>
                  {feeSchedule.map((fee, idx) => (
                    <tr key={idx} className="border-b border-[#2A2C34] hover:bg-[#25272E]">
                      <td className="py-2 text-gray-200">{fee.method}</td>
                      <td className="text-right text-green-400">{fee.depositFeePercent}%</td>
                      <td className="text-right text-gray-400">
                        {fee.depositFeeFixed > 0 ? `$${fee.depositFeeFixed}` : '-'}
                      </td>
                      <td className="text-right text-orange-400">{fee.withdrawalFeePercent}%</td>
                      <td className="text-right text-gray-400">
                        {fee.withdrawalFeeFixed > 0 ? `$${fee.withdrawalFeeFixed}` : '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* Pending Transactions Queue */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
          <h3 className="text-sm font-semibold text-white mb-3 flex items-center gap-2">
            <Clock size={14} className="text-yellow-400" />
            Pending Transactions Queue
          </h3>
          <div className="overflow-auto max-h-96">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                <tr>
                  <th className="text-left px-3 py-2 text-gray-400 font-medium">Transaction ID</th>
                  <th className="text-left px-3 py-2 text-gray-400 font-medium">Type</th>
                  <th className="text-left px-3 py-2 text-gray-400 font-medium">Method</th>
                  <th className="text-left px-3 py-2 text-gray-400 font-medium">Client</th>
                  <th className="text-right px-3 py-2 text-gray-400 font-medium">Amount</th>
                  <th className="text-left px-3 py-2 text-gray-400 font-medium">Status</th>
                  <th className="text-left px-3 py-2 text-gray-400 font-medium">Time</th>
                </tr>
              </thead>
              <tbody>
                {pendingTransactions.map((txn) => (
                  <tr key={txn.id} className="border-b border-[#2A2C34] hover:bg-[#25272E]">
                    <td className="px-3 py-2 font-mono text-gray-400">{txn.id}</td>
                    <td className="px-3 py-2">
                      <span
                        className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                          txn.type === 'deposit'
                            ? 'bg-green-500/20 text-green-400'
                            : 'bg-orange-500/20 text-orange-400'
                        }`}
                      >
                        {txn.type}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-gray-300">{txn.method}</td>
                    <td className="px-3 py-2 text-gray-200">{txn.client}</td>
                    <td className="px-3 py-2 text-right text-white font-medium">
                      ${txn.amount.toLocaleString()} {txn.currency}
                    </td>
                    <td className="px-3 py-2">
                      <span
                        className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                          txn.status === 'pending'
                            ? 'bg-yellow-500/20 text-yellow-400'
                            : 'bg-blue-500/20 text-blue-400'
                        }`}
                      >
                        {txn.status}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-gray-500">{formatTime(txn.timestamp)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Gateway Configuration Modal */}
      {isConfigModalOpen && selectedMethod && gatewayConfig && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
            {/* Modal Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-[#383A42]">
              <div>
                <h2 className="text-lg font-bold text-white flex items-center gap-2">
                  <div className={selectedMethod.color}>{selectedMethod.icon}</div>
                  Configure {selectedMethod.name}
                </h2>
                <p className="text-xs text-gray-400 mt-0.5">{selectedMethod.description}</p>
              </div>
              <button
                onClick={() => setIsConfigModalOpen(false)}
                className="text-gray-400 hover:text-white transition-colors"
              >
                <X size={20} />
              </button>
            </div>

            {/* Modal Body */}
            <div className="flex-1 overflow-auto p-6 space-y-4">
              {/* API Credentials */}
              <div className="space-y-3">
                <h3 className="text-sm font-semibold text-white">API Credentials</h3>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">API Key</label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={maskKey(gatewayConfig.apiKey, showApiKey)}
                      readOnly
                      className="flex-1 px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white font-mono"
                    />
                    <button
                      onClick={() => setShowApiKey(!showApiKey)}
                      className="px-3 py-2 bg-[#2A2C34] hover:bg-[#33353D] text-gray-300 rounded transition-colors"
                    >
                      {showApiKey ? <EyeOff size={16} /> : <Eye size={16} />}
                    </button>
                  </div>
                </div>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Secret Key</label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={maskKey(gatewayConfig.secretKey, showSecretKey)}
                      readOnly
                      className="flex-1 px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white font-mono"
                    />
                    <button
                      onClick={() => setShowSecretKey(!showSecretKey)}
                      className="px-3 py-2 bg-[#2A2C34] hover:bg-[#33353D] text-gray-300 rounded transition-colors"
                    >
                      {showSecretKey ? <EyeOff size={16} /> : <Eye size={16} />}
                    </button>
                  </div>
                </div>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Webhook URL</label>
                  <input
                    type="text"
                    value={gatewayConfig.webhookUrl}
                    onChange={(e) =>
                      setGatewayConfig({ ...gatewayConfig, webhookUrl: e.target.value })
                    }
                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white"
                  />
                </div>
              </div>

              {/* Environment */}
              <div>
                <h3 className="text-sm font-semibold text-white mb-3">Environment</h3>
                <div className="flex items-center gap-4">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      checked={!gatewayConfig.isSandbox}
                      onChange={() => setGatewayConfig({ ...gatewayConfig, isSandbox: false })}
                      className="text-blue-600"
                    />
                    <span className="text-sm text-gray-300">Production</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      checked={gatewayConfig.isSandbox}
                      onChange={() => setGatewayConfig({ ...gatewayConfig, isSandbox: true })}
                      className="text-blue-600"
                    />
                    <span className="text-sm text-gray-300">Sandbox</span>
                  </label>
                </div>
              </div>

              {/* Transaction Limits */}
              <div className="space-y-3">
                <h3 className="text-sm font-semibold text-white">Transaction Limits</h3>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs text-gray-400 mb-1">Min Deposit ($)</label>
                    <input
                      type="number"
                      value={gatewayConfig.minDeposit}
                      onChange={(e) =>
                        setGatewayConfig({
                          ...gatewayConfig,
                          minDeposit: parseInt(e.target.value) || 0,
                        })
                      }
                      className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-400 mb-1">Max Deposit ($)</label>
                    <input
                      type="number"
                      value={gatewayConfig.maxDeposit}
                      onChange={(e) =>
                        setGatewayConfig({
                          ...gatewayConfig,
                          maxDeposit: parseInt(e.target.value) || 0,
                        })
                      }
                      className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-400 mb-1">Min Withdrawal ($)</label>
                    <input
                      type="number"
                      value={gatewayConfig.minWithdrawal}
                      onChange={(e) =>
                        setGatewayConfig({
                          ...gatewayConfig,
                          minWithdrawal: parseInt(e.target.value) || 0,
                        })
                      }
                      className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-400 mb-1">Max Withdrawal ($)</label>
                    <input
                      type="number"
                      value={gatewayConfig.maxWithdrawal}
                      onChange={(e) =>
                        setGatewayConfig({
                          ...gatewayConfig,
                          maxWithdrawal: parseInt(e.target.value) || 0,
                        })
                      }
                      className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white"
                    />
                  </div>
                </div>
              </div>

              {/* Processing Fee */}
              <div>
                <h3 className="text-sm font-semibold text-white mb-3">Processing Fee</h3>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Fee Percentage (%)</label>
                  <input
                    type="number"
                    step="0.1"
                    value={gatewayConfig.processingFee}
                    onChange={(e) =>
                      setGatewayConfig({
                        ...gatewayConfig,
                        processingFee: parseFloat(e.target.value) || 0,
                      })
                    }
                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white"
                  />
                </div>
              </div>

              {/* Supported Currencies */}
              <div>
                <h3 className="text-sm font-semibold text-white mb-3">Supported Currencies</h3>
                <div className="grid grid-cols-5 gap-2">
                  {['USD', 'EUR', 'GBP', 'JPY', 'AUD', 'CAD', 'CHF', 'CNY', 'HKD', 'SGD'].map(
                    (currency) => (
                      <label
                        key={currency}
                        className="flex items-center gap-2 cursor-pointer text-sm"
                      >
                        <input
                          type="checkbox"
                          checked={gatewayConfig.supportedCurrencies.includes(currency)}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setGatewayConfig({
                                ...gatewayConfig,
                                supportedCurrencies: [
                                  ...gatewayConfig.supportedCurrencies,
                                  currency,
                                ],
                              });
                            } else {
                              setGatewayConfig({
                                ...gatewayConfig,
                                supportedCurrencies: gatewayConfig.supportedCurrencies.filter(
                                  (c) => c !== currency
                                ),
                              });
                            }
                          }}
                          className="text-blue-600"
                        />
                        <span className="text-gray-300">{currency}</span>
                      </label>
                    )
                  )}
                </div>
              </div>
            </div>

            {/* Modal Footer */}
            <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-[#383A42]">
              <button
                onClick={() => setIsConfigModalOpen(false)}
                className="px-4 py-2 bg-[#2A2C34] hover:bg-[#33353D] text-white text-sm font-medium rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveConfig}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded transition-colors flex items-center gap-2"
              >
                <Save size={14} />
                Save Configuration
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
