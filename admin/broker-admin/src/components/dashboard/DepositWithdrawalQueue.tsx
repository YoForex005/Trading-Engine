'use client';

import React, { useState, useMemo } from 'react';
import { X, Search, Check, XCircle, Eye, CreditCard, Banknote, Wallet, DollarSign } from 'lucide-react';

type TransactionType = 'deposit' | 'withdrawal';
type TransactionStatus = 'pending' | 'approved' | 'rejected' | 'processing' | 'completed' | 'failed';
type PaymentMethod = 'bank_wire' | 'credit_card' | 'crypto' | 'skrill' | 'neteller' | 'paypal';

interface Transaction {
  id: string;
  clientName: string;
  clientId: string;
  type: TransactionType;
  amount: number;
  currency: string;
  method: PaymentMethod;
  requestedTime: string;
  status: TransactionStatus;
  bankRef?: string;
  walletAddress?: string;
  notes?: string;
  processingTimeline: TimelineEvent[];
}

interface TimelineEvent {
  timestamp: string;
  status: string;
  note: string;
}

export default function DepositWithdrawalQueue() {
  const [selectedType, setSelectedType] = useState<'all' | TransactionType>('all');
  const [selectedStatus, setSelectedStatus] = useState<'all' | TransactionStatus>('all');
  const [selectedMethod, setSelectedMethod] = useState<'all' | PaymentMethod>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [showRejectModal, setShowRejectModal] = useState(false);
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null);
  const [rejectReason, setRejectReason] = useState('');

  // Mock 150 transactions
  const transactions = useMemo<Transaction[]>(() => {
    const types: TransactionType[] = ['deposit', 'withdrawal'];
    const statuses: TransactionStatus[] = ['pending', 'approved', 'rejected', 'processing', 'completed', 'failed'];
    const methods: PaymentMethod[] = ['bank_wire', 'credit_card', 'crypto', 'skrill', 'neteller', 'paypal'];
    const clients = [
      'John Doe', 'Jane Smith', 'Robert Johnson', 'Maria Garcia', 'David Lee',
      'Sarah Wilson', 'Michael Brown', 'Emma Davis', 'James Miller', 'Lisa Anderson',
      'William Taylor', 'Jennifer Thomas', 'Richard Jackson', 'Patricia White', 'Christopher Harris',
      'Nancy Martin', 'Daniel Thompson', 'Karen Martinez', 'Matthew Robinson', 'Betty Clark'
    ];

    const result: Transaction[] = [];
    const now = Date.now();

    for (let i = 0; i < 150; i++) {
      const type = types[Math.floor(Math.random() * types.length)];
      const status = statuses[Math.floor(Math.random() * statuses.length)];
      const method = methods[Math.floor(Math.random() * methods.length)];
      const client = clients[Math.floor(Math.random() * clients.length)];
      const amount = type === 'deposit'
        ? Math.floor(Math.random() * 9000) + 1000
        : Math.floor(Math.random() * 8000) + 500;
      const hoursAgo = Math.floor(Math.random() * 720); // Up to 30 days ago
      const requestedTime = new Date(now - hoursAgo * 60 * 60 * 1000).toISOString();

      const timeline: TimelineEvent[] = [
        { timestamp: requestedTime, status: 'requested', note: 'Transaction requested by client' }
      ];

      if (status !== 'pending') {
        timeline.push({
          timestamp: new Date(now - (hoursAgo - 1) * 60 * 60 * 1000).toISOString(),
          status: status === 'approved' || status === 'completed' ? 'approved' : status,
          note: status === 'approved' ? 'Approved by admin' : status === 'rejected' ? 'Rejected by admin' : 'Status updated'
        });
      }

      if (status === 'completed') {
        timeline.push({
          timestamp: new Date(now - (hoursAgo - 2) * 60 * 60 * 1000).toISOString(),
          status: 'completed',
          note: 'Transaction completed successfully'
        });
      }

      result.push({
        id: `TXN${String(150001 + i).padStart(6, '0')}`,
        clientName: client,
        clientId: `ACC${Math.floor(Math.random() * 900000) + 100000}`,
        type,
        amount,
        currency: 'USD',
        method,
        requestedTime,
        status,
        bankRef: method === 'bank_wire' ? `WIRE${Math.random().toString(36).substr(2, 9).toUpperCase()}` : undefined,
        walletAddress: method === 'crypto' ? `0x${Math.random().toString(16).substr(2, 40)}` : undefined,
        notes: Math.random() > 0.7 ? 'Client requested priority processing' : '',
        processingTimeline: timeline
      });
    }

    return result.sort((a, b) => new Date(b.requestedTime).getTime() - new Date(a.requestedTime).getTime());
  }, []);

  // Filter transactions
  const filteredTransactions = useMemo(() => {
    return transactions.filter(tx => {
      if (selectedType !== 'all' && tx.type !== selectedType) return false;
      if (selectedStatus !== 'all' && tx.status !== selectedStatus) return false;
      if (selectedMethod !== 'all' && tx.method !== selectedMethod) return false;
      if (searchQuery && !tx.clientName.toLowerCase().includes(searchQuery.toLowerCase()) && !tx.id.toLowerCase().includes(searchQuery.toLowerCase())) return false;
      return true;
    });
  }, [transactions, selectedType, selectedStatus, selectedMethod, searchQuery]);

  // Pending transactions (priority view)
  const pendingTransactions = useMemo(() => {
    return transactions.filter(tx => tx.status === 'pending');
  }, [transactions]);

  // Stats
  const stats = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    let todayDeposits = 0;
    let todayWithdrawals = 0;
    let processingCount = 0;

    transactions.forEach(tx => {
      const txDate = new Date(tx.requestedTime);
      if (txDate >= today) {
        if (tx.type === 'deposit' && (tx.status === 'approved' || tx.status === 'completed')) {
          todayDeposits += tx.amount;
        }
        if (tx.type === 'withdrawal' && (tx.status === 'approved' || tx.status === 'completed')) {
          todayWithdrawals += tx.amount;
        }
      }
      if (tx.status === 'processing') {
        processingCount++;
      }
    });

    return {
      pendingCount: pendingTransactions.length,
      todayDeposits,
      todayWithdrawals,
      processingCount
    };
  }, [transactions, pendingTransactions]);

  // Daily volume data for chart (30 days)
  const dailyVolumeData = useMemo(() => {
    const days = 30;
    const data: Array<{ date: string; deposits: number; withdrawals: number }> = [];
    const now = Date.now();

    for (let i = days - 1; i >= 0; i--) {
      const dayStart = new Date(now - i * 24 * 60 * 60 * 1000);
      dayStart.setHours(0, 0, 0, 0);
      const dayEnd = new Date(dayStart);
      dayEnd.setHours(23, 59, 59, 999);

      let deposits = 0;
      let withdrawals = 0;

      transactions.forEach(tx => {
        const txDate = new Date(tx.requestedTime);
        if (txDate >= dayStart && txDate <= dayEnd && (tx.status === 'approved' || tx.status === 'completed')) {
          if (tx.type === 'deposit') {
            deposits += tx.amount;
          } else {
            withdrawals += tx.amount;
          }
        }
      });

      data.push({
        date: dayStart.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        deposits,
        withdrawals
      });
    }

    return data;
  }, [transactions]);

  // Method breakdown
  const methodBreakdown = useMemo(() => {
    const counts: Record<PaymentMethod, number> = {
      bank_wire: 0,
      credit_card: 0,
      crypto: 0,
      skrill: 0,
      neteller: 0,
      paypal: 0
    };

    transactions.forEach(tx => {
      counts[tx.method]++;
    });

    return Object.entries(counts).map(([method, count]) => ({ method: method as PaymentMethod, count }));
  }, [transactions]);

  const getStatusColor = (status: TransactionStatus) => {
    switch (status) {
      case 'pending': return 'text-[#F59E0B]';
      case 'approved': return 'text-[#10B981]';
      case 'rejected': return 'text-[#EF4444]';
      case 'processing': return 'text-[#3B82F6]';
      case 'completed': return 'text-[#10B981]';
      case 'failed': return 'text-[#6B7280]';
    }
  };

  const getStatusBg = (status: TransactionStatus) => {
    switch (status) {
      case 'pending': return 'bg-[#F59E0B]/10 border-[#F59E0B]/20';
      case 'approved': return 'bg-[#10B981]/10 border-[#10B981]/20';
      case 'rejected': return 'bg-[#EF4444]/10 border-[#EF4444]/20';
      case 'processing': return 'bg-[#3B82F6]/10 border-[#3B82F6]/20';
      case 'completed': return 'bg-[#10B981]/10 border-[#10B981]/20';
      case 'failed': return 'bg-[#6B7280]/10 border-[#6B7280]/20';
    }
  };

  const getMethodLabel = (method: PaymentMethod) => {
    return method.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
  };

  const getAge = (requestedTime: string) => {
    const now = Date.now();
    const then = new Date(requestedTime).getTime();
    const diffMs = now - then;
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays}d ${diffHours % 24}h`;
    return `${diffHours}h`;
  };

  const handleApprove = (tx: Transaction) => {
    if (confirm(`Approve ${tx.type} of $${tx.amount.toLocaleString()} for ${tx.clientName}?`)) {
      console.log('Approving transaction:', tx.id);
      // In real implementation, call API
    }
  };

  const handleReject = (tx: Transaction) => {
    setSelectedTransaction(tx);
    setShowRejectModal(true);
  };

  const submitReject = () => {
    if (!rejectReason.trim()) {
      alert('Please provide a rejection reason');
      return;
    }
    console.log('Rejecting transaction:', selectedTransaction?.id, 'Reason:', rejectReason);
    // In real implementation, call API
    setShowRejectModal(false);
    setRejectReason('');
    setSelectedTransaction(null);
  };

  const handleViewDetail = (tx: Transaction) => {
    setSelectedTransaction(tx);
    setShowDetailModal(true);
  };

  return (
    <div className="h-full flex flex-col bg-[#18181b] text-[#E5E5E5] overflow-hidden">
      {/* Header */}
      <div className="flex-shrink-0 px-6 py-4 border-b border-[#3f3f46]">
        <h1 className="text-xl font-semibold">Deposit & Withdrawal Processing</h1>
      </div>

      {/* Stats Cards */}
      <div className="flex-shrink-0 px-6 py-4 grid grid-cols-4 gap-4">
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="text-xs text-[#888] mb-1">Pending Approvals</div>
          <div className="text-2xl font-bold text-[#F59E0B]">{stats.pendingCount}</div>
        </div>
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="text-xs text-[#888] mb-1">Today's Deposits Total</div>
          <div className="text-2xl font-bold text-[#10B981]">${stats.todayDeposits.toLocaleString()}</div>
        </div>
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="text-xs text-[#888] mb-1">Today's Withdrawals Total</div>
          <div className="text-2xl font-bold text-[#EF4444]">${stats.todayWithdrawals.toLocaleString()}</div>
        </div>
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="text-xs text-[#888] mb-1">Processing Count</div>
          <div className="text-2xl font-bold text-[#3B82F6]">{stats.processingCount}</div>
        </div>
      </div>

      {/* Pending Queue (Priority View) */}
      {pendingTransactions.length > 0 && (
        <div className="flex-shrink-0 px-6 py-4">
          <h2 className="text-sm font-semibold mb-3 text-[#F59E0B]">⚠️ Pending Approval Queue ({pendingTransactions.length})</h2>
          <div className="bg-[#27272a] border border-[#3f3f46] rounded overflow-hidden">
            <div className="overflow-x-auto max-h-64">
              <table className="w-full text-xs">
                <thead className="bg-[#1e1e22] border-b border-[#3f3f46] sticky top-0">
                  <tr>
                    <th className="px-3 py-2 text-left font-semibold text-[#888]">Transaction ID</th>
                    <th className="px-3 py-2 text-left font-semibold text-[#888]">Client</th>
                    <th className="px-3 py-2 text-left font-semibold text-[#888]">Type</th>
                    <th className="px-3 py-2 text-right font-semibold text-[#888]">Amount</th>
                    <th className="px-3 py-2 text-left font-semibold text-[#888]">Method</th>
                    <th className="px-3 py-2 text-left font-semibold text-[#888]">Requested</th>
                    <th className="px-3 py-2 text-right font-semibold text-[#888]">Age</th>
                    <th className="px-3 py-2 text-center font-semibold text-[#888]">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {pendingTransactions.map(tx => (
                    <tr key={tx.id} className="border-b border-[#3f3f46] hover:bg-[#2a2a2e] transition-colors">
                      <td className="px-3 py-2 font-mono text-[#3B82F6]">{tx.id}</td>
                      <td className="px-3 py-2">{tx.clientName}</td>
                      <td className="px-3 py-2">
                        <span className={`inline-block px-2 py-0.5 rounded text-xs ${
                          tx.type === 'deposit' ? 'bg-[#10B981]/10 text-[#10B981]' : 'bg-[#EF4444]/10 text-[#EF4444]'
                        }`}>
                          {tx.type.charAt(0).toUpperCase() + tx.type.slice(1)}
                        </span>
                      </td>
                      <td className="px-3 py-2 text-right font-semibold">${tx.amount.toLocaleString()}</td>
                      <td className="px-3 py-2">{getMethodLabel(tx.method)}</td>
                      <td className="px-3 py-2 text-[#888]">{new Date(tx.requestedTime).toLocaleString()}</td>
                      <td className="px-3 py-2 text-right font-semibold text-[#F59E0B]">{getAge(tx.requestedTime)}</td>
                      <td className="px-3 py-2">
                        <div className="flex items-center justify-center gap-1">
                          <button
                            onClick={() => handleApprove(tx)}
                            className="p-1 hover:bg-[#10B981]/20 rounded text-[#10B981] transition-colors"
                            title="Approve"
                          >
                            <Check size={14} />
                          </button>
                          <button
                            onClick={() => handleReject(tx)}
                            className="p-1 hover:bg-[#EF4444]/20 rounded text-[#EF4444] transition-colors"
                            title="Reject"
                          >
                            <XCircle size={14} />
                          </button>
                          <button
                            onClick={() => handleViewDetail(tx)}
                            className="p-1 hover:bg-[#3B82F6]/20 rounded text-[#3B82F6] transition-colors"
                            title="View Details"
                          >
                            <Eye size={14} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="flex-shrink-0 px-6 py-3 border-b border-[#3f3f46] flex items-center gap-4">
        <div className="relative flex-1 max-w-xs">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#666]" />
          <input
            type="text"
            placeholder="Search client or transaction ID..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-[#27272a] border border-[#3f3f46] rounded pl-9 pr-3 py-1.5 text-xs focus:outline-none focus:border-[#3B82F6]"
          />
        </div>
        <select
          value={selectedType}
          onChange={(e) => setSelectedType(e.target.value as any)}
          className="bg-[#27272a] border border-[#3f3f46] rounded px-3 py-1.5 text-xs focus:outline-none focus:border-[#3B82F6]"
        >
          <option value="all">All Types</option>
          <option value="deposit">Deposits</option>
          <option value="withdrawal">Withdrawals</option>
        </select>
        <select
          value={selectedStatus}
          onChange={(e) => setSelectedStatus(e.target.value as any)}
          className="bg-[#27272a] border border-[#3f3f46] rounded px-3 py-1.5 text-xs focus:outline-none focus:border-[#3B82F6]"
        >
          <option value="all">All Statuses</option>
          <option value="pending">Pending</option>
          <option value="approved">Approved</option>
          <option value="rejected">Rejected</option>
          <option value="processing">Processing</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
        </select>
        <select
          value={selectedMethod}
          onChange={(e) => setSelectedMethod(e.target.value as any)}
          className="bg-[#27272a] border border-[#3f3f46] rounded px-3 py-1.5 text-xs focus:outline-none focus:border-[#3B82F6]"
        >
          <option value="all">All Methods</option>
          <option value="bank_wire">Bank Wire</option>
          <option value="credit_card">Credit Card</option>
          <option value="crypto">Crypto</option>
          <option value="skrill">Skrill</option>
          <option value="neteller">Neteller</option>
          <option value="paypal">PayPal</option>
        </select>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 overflow-auto px-6 py-4">
        <div className="grid grid-cols-3 gap-6 mb-6">
          {/* Daily Volume Chart */}
          <div className="col-span-2 bg-[#27272a] border border-[#3f3f46] rounded p-4">
            <h3 className="text-sm font-semibold mb-3">30-Day Volume (Deposits vs Withdrawals)</h3>
            <div className="w-full h-64">
              <svg viewBox="0 0 900 240" className="w-full h-full">
                {/* Grid lines */}
                {[0, 1, 2, 3, 4].map(i => (
                  <line key={i} x1="40" y1={i * 60} x2="880" y2={i * 60} stroke="#3f3f46" strokeWidth="0.5" />
                ))}
                {/* Bars */}
                {dailyVolumeData.map((day, idx) => {
                  const x = 40 + (idx * 28);
                  const maxVal = Math.max(...dailyVolumeData.flatMap(d => [d.deposits, d.withdrawals]));
                  const depositHeight = (day.deposits / maxVal) * 200;
                  const withdrawalHeight = (day.withdrawals / maxVal) * 200;

                  return (
                    <g key={idx}>
                      <rect x={x} y={240 - depositHeight} width="10" height={depositHeight} fill="#10B981" opacity="0.8" />
                      <rect x={x + 12} y={240 - withdrawalHeight} width="10" height={withdrawalHeight} fill="#EF4444" opacity="0.8" />
                      {idx % 5 === 0 && (
                        <text x={x + 11} y="235" fill="#888" fontSize="8" textAnchor="middle">
                          {day.date.split(' ')[1]}
                        </text>
                      )}
                    </g>
                  );
                })}
                <text x="450" y="15" fill="#E5E5E5" fontSize="11" textAnchor="middle" fontWeight="600">
                  Daily Transaction Volume
                </text>
                <rect x="700" y="10" width="10" height="10" fill="#10B981" />
                <text x="715" y="19" fill="#888" fontSize="9">Deposits</text>
                <rect x="780" y="10" width="10" height="10" fill="#EF4444" />
                <text x="795" y="19" fill="#888" fontSize="9">Withdrawals</text>
              </svg>
            </div>
          </div>

          {/* Method Breakdown */}
          <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
            <h3 className="text-sm font-semibold mb-3">Transaction Methods</h3>
            <div className="space-y-2">
              {methodBreakdown.map(({ method, count }) => {
                const total = methodBreakdown.reduce((sum, m) => sum + m.count, 0);
                const percentage = ((count / total) * 100).toFixed(1);
                const colors: Record<PaymentMethod, string> = {
                  bank_wire: '#3B82F6',
                  credit_card: '#10B981',
                  crypto: '#F59E0B',
                  skrill: '#8B5CF6',
                  neteller: '#EC4899',
                  paypal: '#06B6D4'
                };
                return (
                  <div key={method} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className="w-3 h-3 rounded" style={{ backgroundColor: colors[method] }} />
                      <span className="text-xs text-[#E5E5E5]">{getMethodLabel(method)}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-semibold" style={{ color: colors[method] }}>{count}</span>
                      <span className="text-xs text-[#666]">({percentage}%)</span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* All Transactions Table */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded overflow-hidden">
          <div className="px-4 py-3 border-b border-[#3f3f46]">
            <h3 className="text-sm font-semibold">All Transactions ({filteredTransactions.length})</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-[#1e1e22] border-b border-[#3f3f46]">
                <tr>
                  <th className="px-3 py-2 text-left font-semibold text-[#888]">ID</th>
                  <th className="px-3 py-2 text-left font-semibold text-[#888]">Client</th>
                  <th className="px-3 py-2 text-left font-semibold text-[#888]">Type</th>
                  <th className="px-3 py-2 text-right font-semibold text-[#888]">Amount</th>
                  <th className="px-3 py-2 text-left font-semibold text-[#888]">Method</th>
                  <th className="px-3 py-2 text-left font-semibold text-[#888]">Requested</th>
                  <th className="px-3 py-2 text-center font-semibold text-[#888]">Status</th>
                  <th className="px-3 py-2 text-center font-semibold text-[#888]">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredTransactions.map(tx => (
                  <tr key={tx.id} className="border-b border-[#3f3f46] hover:bg-[#2a2a2e] transition-colors">
                    <td className="px-3 py-2 font-mono text-[#3B82F6]">{tx.id}</td>
                    <td className="px-3 py-2">{tx.clientName}</td>
                    <td className="px-3 py-2">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs ${
                        tx.type === 'deposit' ? 'bg-[#10B981]/10 text-[#10B981]' : 'bg-[#EF4444]/10 text-[#EF4444]'
                      }`}>
                        {tx.type.charAt(0).toUpperCase() + tx.type.slice(1)}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-right font-semibold">${tx.amount.toLocaleString()}</td>
                    <td className="px-3 py-2">{getMethodLabel(tx.method)}</td>
                    <td className="px-3 py-2 text-[#888]">{new Date(tx.requestedTime).toLocaleString()}</td>
                    <td className="px-3 py-2 text-center">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs border ${getStatusBg(tx.status)} ${getStatusColor(tx.status)}`}>
                        {tx.status.charAt(0).toUpperCase() + tx.status.slice(1)}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-center">
                      <button
                        onClick={() => handleViewDetail(tx)}
                        className="p-1 hover:bg-[#3B82F6]/20 rounded text-[#3B82F6] transition-colors"
                        title="View Details"
                      >
                        <Eye size={14} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Transaction Detail Modal */}
      {showDetailModal && selectedTransaction && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#27272a] border border-[#3f3f46] rounded-lg w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
            <div className="flex items-center justify-between px-4 py-3 border-b border-[#3f3f46]">
              <h3 className="text-sm font-semibold">Transaction Details - {selectedTransaction.id}</h3>
              <button onClick={() => setShowDetailModal(false)} className="text-[#888] hover:text-[#E5E5E5]">
                <X size={16} />
              </button>
            </div>
            <div className="p-4 overflow-auto flex-1">
              <div className="grid grid-cols-2 gap-4 mb-4">
                <div>
                  <div className="text-xs text-[#888] mb-1">Client</div>
                  <div className="text-sm font-semibold">{selectedTransaction.clientName}</div>
                  <div className="text-xs text-[#666]">{selectedTransaction.clientId}</div>
                </div>
                <div>
                  <div className="text-xs text-[#888] mb-1">Status</div>
                  <span className={`inline-block px-2 py-0.5 rounded text-xs border ${getStatusBg(selectedTransaction.status)} ${getStatusColor(selectedTransaction.status)}`}>
                    {selectedTransaction.status.charAt(0).toUpperCase() + selectedTransaction.status.slice(1)}
                  </span>
                </div>
                <div>
                  <div className="text-xs text-[#888] mb-1">Type</div>
                  <div className="text-sm font-semibold">{selectedTransaction.type.charAt(0).toUpperCase() + selectedTransaction.type.slice(1)}</div>
                </div>
                <div>
                  <div className="text-xs text-[#888] mb-1">Amount</div>
                  <div className="text-sm font-semibold">${selectedTransaction.amount.toLocaleString()} {selectedTransaction.currency}</div>
                </div>
                <div>
                  <div className="text-xs text-[#888] mb-1">Payment Method</div>
                  <div className="text-sm">{getMethodLabel(selectedTransaction.method)}</div>
                </div>
                <div>
                  <div className="text-xs text-[#888] mb-1">Requested Time</div>
                  <div className="text-sm">{new Date(selectedTransaction.requestedTime).toLocaleString()}</div>
                </div>
                {selectedTransaction.bankRef && (
                  <div className="col-span-2">
                    <div className="text-xs text-[#888] mb-1">Bank Reference</div>
                    <div className="text-sm font-mono">{selectedTransaction.bankRef}</div>
                  </div>
                )}
                {selectedTransaction.walletAddress && (
                  <div className="col-span-2">
                    <div className="text-xs text-[#888] mb-1">Wallet Address</div>
                    <div className="text-sm font-mono text-[#3B82F6]">{selectedTransaction.walletAddress}</div>
                  </div>
                )}
                {selectedTransaction.notes && (
                  <div className="col-span-2">
                    <div className="text-xs text-[#888] mb-1">Notes</div>
                    <div className="text-sm">{selectedTransaction.notes}</div>
                  </div>
                )}
              </div>
              <div className="mt-6">
                <div className="text-xs text-[#888] mb-3">Processing Timeline</div>
                <div className="space-y-3">
                  {selectedTransaction.processingTimeline.map((event, idx) => (
                    <div key={idx} className="flex gap-3">
                      <div className="flex flex-col items-center">
                        <div className="w-2 h-2 rounded-full bg-[#3B82F6]" />
                        {idx < selectedTransaction.processingTimeline.length - 1 && (
                          <div className="w-px h-full bg-[#3f3f46] mt-1" />
                        )}
                      </div>
                      <div className="flex-1 pb-4">
                        <div className="text-xs font-semibold text-[#E5E5E5]">{event.status.charAt(0).toUpperCase() + event.status.slice(1)}</div>
                        <div className="text-xs text-[#888]">{new Date(event.timestamp).toLocaleString()}</div>
                        <div className="text-xs text-[#666] mt-1">{event.note}</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Reject Modal */}
      {showRejectModal && selectedTransaction && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#27272a] border border-[#3f3f46] rounded-lg w-full max-w-md">
            <div className="flex items-center justify-between px-4 py-3 border-b border-[#3f3f46]">
              <h3 className="text-sm font-semibold">Reject Transaction - {selectedTransaction.id}</h3>
              <button onClick={() => setShowRejectModal(false)} className="text-[#888] hover:text-[#E5E5E5]">
                <X size={16} />
              </button>
            </div>
            <div className="p-4">
              <div className="mb-4">
                <div className="text-xs text-[#888] mb-2">Client: <span className="text-[#E5E5E5] font-semibold">{selectedTransaction.clientName}</span></div>
                <div className="text-xs text-[#888] mb-2">Amount: <span className="text-[#E5E5E5] font-semibold">${selectedTransaction.amount.toLocaleString()}</span></div>
              </div>
              <div>
                <label className="block text-xs text-[#888] mb-2">Rejection Reason *</label>
                <textarea
                  value={rejectReason}
                  onChange={(e) => setRejectReason(e.target.value)}
                  className="w-full bg-[#1e1e22] border border-[#3f3f46] rounded px-3 py-2 text-xs focus:outline-none focus:border-[#3B82F6] resize-none"
                  rows={4}
                  placeholder="Please provide a clear reason for rejection..."
                />
              </div>
              <div className="flex gap-2 mt-4">
                <button
                  onClick={submitReject}
                  className="flex-1 px-4 py-2 bg-[#EF4444] hover:bg-[#DC2626] text-white text-xs rounded transition-colors"
                >
                  Confirm Rejection
                </button>
                <button
                  onClick={() => setShowRejectModal(false)}
                  className="flex-1 px-4 py-2 bg-[#3f3f46] hover:bg-[#52525b] text-[#E5E5E5] text-xs rounded transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
