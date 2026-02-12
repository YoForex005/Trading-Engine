/**
 * Financial Operations Management
 * Unified deposit/withdrawal processing and internal transfer tracking
 */

'use client';

import React, { useState, useMemo } from 'react';
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  Clock,
  CheckCircle,
  XCircle,
  AlertTriangle,
  ArrowRightLeft,
  Calendar,
  Filter,
  Eye,
  Shield,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// Types
type PaymentMethod = 'Bank Wire' | 'Credit Card' | 'Debit Card' | 'Crypto' | 'E-Wallet';
type DepositStatus = 'pending' | 'completed' | 'failed';
type WithdrawalStatus = 'pending' | 'approved' | 'rejected';
type TransferType = 'internal' | 'rebate' | 'commission' | 'adjustment';
type TransferStatus = 'completed' | 'pending' | 'failed';

interface Deposit {
  id: string;
  clientName: string;
  clientId: string;
  amount: number;
  currency: string;
  method: PaymentMethod;
  status: DepositStatus;
  date: number;
  notes?: string;
  transactionRef?: string;
}

interface Withdrawal {
  id: string;
  clientName: string;
  clientId: string;
  amount: number;
  method: PaymentMethod;
  status: WithdrawalStatus;
  riskScore: number;
  autoApprovable: boolean;
  riskFlags: string[];
  date: number;
}

interface Transfer {
  id: string;
  fromAccount: string;
  fromClientName: string;
  toAccount: string;
  toClientName: string;
  amount: number;
  type: TransferType;
  date: number;
  status: TransferStatus;
  notes?: string;
}

interface DailyFlow {
  date: string;
  deposits: number;
  withdrawals: number;
  runningBalance: number;
}

// Mock Data - 20 deposits
const MOCK_DEPOSITS: Deposit[] = [
  { id: 'DEP-2026-001', clientName: 'John Smith', clientId: '680962851', amount: 5000, currency: 'USD', method: 'Bank Wire', status: 'pending', date: Date.now() - 3600000, transactionRef: 'TXN-BW-001' },
  { id: 'DEP-2026-002', clientName: 'Sarah Johnson', clientId: '680962852', amount: 10000, currency: 'USD', method: 'Credit Card', status: 'completed', date: Date.now() - 7200000 },
  { id: 'DEP-2026-003', clientName: 'Mike Wilson', clientId: '680962853', amount: 2500, currency: 'EUR', method: 'Crypto', status: 'pending', date: Date.now() - 1800000, transactionRef: 'TXN-CR-003' },
  { id: 'DEP-2026-004', clientName: 'Emily Brown', clientId: '680962854', amount: 7500, currency: 'USD', method: 'E-Wallet', status: 'completed', date: Date.now() - 86400000 },
  { id: 'DEP-2026-005', clientName: 'David Lee', clientId: '680962855', amount: 15000, currency: 'GBP', method: 'Bank Wire', status: 'completed', date: Date.now() - 172800000 },
  { id: 'DEP-2026-006', clientName: 'Lisa Chen', clientId: '680962856', amount: 3000, currency: 'USD', method: 'Debit Card', status: 'failed', date: Date.now() - 259200000, notes: 'Insufficient funds' },
  { id: 'DEP-2026-007', clientName: 'Robert Taylor', clientId: '680962857', amount: 20000, currency: 'USD', method: 'Bank Wire', status: 'pending', date: Date.now() - 5400000, transactionRef: 'TXN-BW-007' },
  { id: 'DEP-2026-008', clientName: 'Amanda White', clientId: '680962858', amount: 8500, currency: 'EUR', method: 'Credit Card', status: 'completed', date: Date.now() - 345600000 },
  { id: 'DEP-2026-009', clientName: 'James Martinez', clientId: '680962859', amount: 12000, currency: 'USD', method: 'Crypto', status: 'completed', date: Date.now() - 432000000 },
  { id: 'DEP-2026-010', clientName: 'Jennifer Garcia', clientId: '680962860', amount: 4500, currency: 'USD', method: 'E-Wallet', status: 'pending', date: Date.now() - 9000000, transactionRef: 'TXN-EW-010' },
  { id: 'DEP-2026-011', clientName: 'Thomas Anderson', clientId: '680962861', amount: 6000, currency: 'GBP', method: 'Bank Wire', status: 'completed', date: Date.now() - 518400000 },
  { id: 'DEP-2026-012', clientName: 'Patricia Moore', clientId: '680962862', amount: 9500, currency: 'USD', method: 'Credit Card', status: 'completed', date: Date.now() - 604800000 },
  { id: 'DEP-2026-013', clientName: 'Christopher King', clientId: '680962863', amount: 3500, currency: 'EUR', method: 'Debit Card', status: 'failed', date: Date.now() - 691200000, notes: 'Card declined' },
  { id: 'DEP-2026-014', clientName: 'Nancy Harris', clientId: '680962864', amount: 11000, currency: 'USD', method: 'Bank Wire', status: 'completed', date: Date.now() - 777600000 },
  { id: 'DEP-2026-015', clientName: 'Daniel Clark', clientId: '680962865', amount: 5500, currency: 'USD', method: 'Crypto', status: 'completed', date: Date.now() - 864000000 },
  { id: 'DEP-2026-016', clientName: 'Barbara Lewis', clientId: '680962866', amount: 7000, currency: 'GBP', method: 'E-Wallet', status: 'completed', date: Date.now() - 950400000 },
  { id: 'DEP-2026-017', clientName: 'Matthew Walker', clientId: '680962867', amount: 13500, currency: 'USD', method: 'Bank Wire', status: 'completed', date: Date.now() - 1036800000 },
  { id: 'DEP-2026-018', clientName: 'Susan Hall', clientId: '680962868', amount: 4000, currency: 'EUR', method: 'Credit Card', status: 'completed', date: Date.now() - 1123200000 },
  { id: 'DEP-2026-019', clientName: 'Joseph Allen', clientId: '680962869', amount: 8000, currency: 'USD', method: 'Debit Card', status: 'completed', date: Date.now() - 1209600000 },
  { id: 'DEP-2026-020', clientName: 'Jessica Young', clientId: '680962870', amount: 16000, currency: 'USD', method: 'Bank Wire', status: 'completed', date: Date.now() - 1296000000 },
];

// Mock Withdrawals - 10 records
const MOCK_WITHDRAWALS: Withdrawal[] = [
  { id: 'WD-2026-001', clientName: 'Trader Alpha', clientId: '680962851', amount: 5000, method: 'Bank Wire', status: 'pending', riskScore: 15, autoApprovable: true, riskFlags: [], date: Date.now() - 3600000 },
  { id: 'WD-2026-002', clientName: 'Trader Beta', clientId: '680962852', amount: 25000, method: 'Crypto', status: 'pending', riskScore: 65, autoApprovable: false, riskFlags: ['large_amount', 'ip_change'], date: Date.now() - 7200000 },
  { id: 'WD-2026-003', clientName: 'Trader Gamma', clientId: '680962853', amount: 1500, method: 'Credit Card', status: 'pending', riskScore: 85, autoApprovable: false, riskFlags: ['velocity', 'new_account'], date: Date.now() - 1800000 },
  { id: 'WD-2026-004', clientName: 'Trader Delta', clientId: '680962854', amount: 3000, method: 'Bank Wire', status: 'approved', riskScore: 20, autoApprovable: true, riskFlags: [], date: Date.now() - 86400000 },
  { id: 'WD-2026-005', clientName: 'Trader Epsilon', clientId: '680962855', amount: 8000, method: 'E-Wallet', status: 'approved', riskScore: 30, autoApprovable: true, riskFlags: [], date: Date.now() - 172800000 },
  { id: 'WD-2026-006', clientName: 'Trader Zeta', clientId: '680962856', amount: 12000, method: 'Bank Wire', status: 'rejected', riskScore: 75, autoApprovable: false, riskFlags: ['duplicate_request'], date: Date.now() - 259200000 },
  { id: 'WD-2026-007', clientName: 'Trader Eta', clientId: '680962857', amount: 2000, method: 'Debit Card', status: 'pending', riskScore: 10, autoApprovable: true, riskFlags: [], date: Date.now() - 5400000 },
  { id: 'WD-2026-008', clientName: 'Trader Theta', clientId: '680962858', amount: 6500, method: 'Bank Wire', status: 'approved', riskScore: 25, autoApprovable: true, riskFlags: [], date: Date.now() - 345600000 },
  { id: 'WD-2026-009', clientName: 'Trader Iota', clientId: '680962859', amount: 18000, method: 'Crypto', status: 'pending', riskScore: 55, autoApprovable: false, riskFlags: ['large_amount'], date: Date.now() - 432000000 },
  { id: 'WD-2026-010', clientName: 'Trader Kappa', clientId: '680962860', amount: 4500, method: 'E-Wallet', status: 'approved', riskScore: 18, autoApprovable: true, riskFlags: [], date: Date.now() - 518400000 },
];

// Mock Transfers - 18 records
const MOCK_TRANSFERS: Transfer[] = [
  { id: 'TRF-2026-001', fromAccount: '680962851', fromClientName: 'John Smith', toAccount: '680962852', toClientName: 'Sarah Johnson', amount: 1000, type: 'internal', date: Date.now() - 3600000, status: 'completed' },
  { id: 'TRF-2026-002', fromAccount: 'BROKER-COMMISSION', fromClientName: 'Broker Account', toAccount: '680962853', toClientName: 'Mike Wilson', amount: 50, type: 'rebate', date: Date.now() - 7200000, status: 'completed', notes: 'Q1 2026 rebate' },
  { id: 'TRF-2026-003', fromAccount: '680962854', fromClientName: 'Emily Brown', toAccount: 'BROKER-COMMISSION', toClientName: 'Broker Account', amount: 35, type: 'commission', date: Date.now() - 10800000, status: 'completed' },
  { id: 'TRF-2026-004', fromAccount: '680962855', fromClientName: 'David Lee', toAccount: '680962856', toClientName: 'Lisa Chen', amount: 2500, type: 'internal', date: Date.now() - 86400000, status: 'completed' },
  { id: 'TRF-2026-005', fromAccount: 'BROKER-ADMIN', fromClientName: 'Admin Account', toAccount: '680962857', toClientName: 'Robert Taylor', amount: 100, type: 'adjustment', date: Date.now() - 172800000, status: 'completed', notes: 'Balance correction' },
  { id: 'TRF-2026-006', fromAccount: '680962858', fromClientName: 'Amanda White', toAccount: '680962859', toClientName: 'James Martinez', amount: 500, type: 'internal', date: Date.now() - 259200000, status: 'completed' },
  { id: 'TRF-2026-007', fromAccount: 'BROKER-COMMISSION', fromClientName: 'Broker Account', toAccount: '680962860', toClientName: 'Jennifer Garcia', amount: 75, type: 'rebate', date: Date.now() - 345600000, status: 'completed' },
  { id: 'TRF-2026-008', fromAccount: '680962861', fromClientName: 'Thomas Anderson', toAccount: 'BROKER-COMMISSION', toClientName: 'Broker Account', amount: 42, type: 'commission', date: Date.now() - 432000000, status: 'completed' },
  { id: 'TRF-2026-009', fromAccount: '680962862', fromClientName: 'Patricia Moore', toAccount: '680962863', toClientName: 'Christopher King', amount: 1500, type: 'internal', date: Date.now() - 518400000, status: 'completed' },
  { id: 'TRF-2026-010', fromAccount: '680962864', fromClientName: 'Nancy Harris', toAccount: '680962865', toClientName: 'Daniel Clark', amount: 800, type: 'internal', date: Date.now() - 604800000, status: 'completed' },
  { id: 'TRF-2026-011', fromAccount: 'BROKER-COMMISSION', fromClientName: 'Broker Account', toAccount: '680962866', toClientName: 'Barbara Lewis', amount: 60, type: 'rebate', date: Date.now() - 691200000, status: 'completed' },
  { id: 'TRF-2026-012', fromAccount: '680962867', fromClientName: 'Matthew Walker', toAccount: 'BROKER-COMMISSION', toClientName: 'Broker Account', amount: 55, type: 'commission', date: Date.now() - 777600000, status: 'completed' },
  { id: 'TRF-2026-013', fromAccount: '680962868', fromClientName: 'Susan Hall', toAccount: '680962869', toClientName: 'Joseph Allen', amount: 3000, type: 'internal', date: Date.now() - 864000000, status: 'completed' },
  { id: 'TRF-2026-014', fromAccount: 'BROKER-ADMIN', fromClientName: 'Admin Account', toAccount: '680962870', toClientName: 'Jessica Young', amount: 250, type: 'adjustment', date: Date.now() - 950400000, status: 'completed', notes: 'Compensation' },
  { id: 'TRF-2026-015', fromAccount: '680962851', fromClientName: 'John Smith', toAccount: '680962854', toClientName: 'Emily Brown', amount: 1200, type: 'internal', date: Date.now() - 1036800000, status: 'completed' },
  { id: 'TRF-2026-016', fromAccount: '680962855', fromClientName: 'David Lee', toAccount: 'BROKER-COMMISSION', toClientName: 'Broker Account', amount: 48, type: 'commission', date: Date.now() - 1123200000, status: 'completed' },
  { id: 'TRF-2026-017', fromAccount: 'BROKER-COMMISSION', fromClientName: 'Broker Account', toAccount: '680962858', toClientName: 'Amanda White', amount: 90, type: 'rebate', date: Date.now() - 1209600000, status: 'completed' },
  { id: 'TRF-2026-018', fromAccount: '680962862', fromClientName: 'Patricia Moore', toAccount: '680962867', toClientName: 'Matthew Walker', amount: 2000, type: 'internal', date: Date.now() - 1296000000, status: 'completed' },
];

export default function FinancialOperations() {
  const [activeTab, setActiveTab] = useState<'deposits' | 'withdrawals' | 'transfers'>('deposits');
  const [deposits] = useState<Deposit[]>(MOCK_DEPOSITS);
  const [withdrawals] = useState<Withdrawal[]>(MOCK_WITHDRAWALS);
  const [transfers] = useState<Transfer[]>(MOCK_TRANSFERS);

  // Deposit filters
  const [depositStatus, setDepositStatus] = useState<'all' | DepositStatus>('all');
  const [depositMethod, setDepositMethod] = useState<'all' | PaymentMethod>('all');
  const [depositMinAmount, setDepositMinAmount] = useState('');
  const [depositMaxAmount, setDepositMaxAmount] = useState('');

  // Withdrawal filters
  const [withdrawalStatus, setWithdrawalStatus] = useState<'all' | WithdrawalStatus>('all');
  const [withdrawalMethod, setWithdrawalMethod] = useState<'all' | PaymentMethod>('all');

  // Transfer filters
  const [transferType, setTransferType] = useState<'all' | TransferType>('all');

  // Summary stats
  const stats = useMemo(() => {
    const now = Date.now();
    const oneDayAgo = now - 86400000;

    const depositsToday = deposits.filter(d => d.status === 'completed' && d.date >= oneDayAgo);
    const withdrawalsToday = withdrawals.filter(w => w.status === 'approved' && w.date >= oneDayAgo);

    const totalDepositsToday = depositsToday.reduce((sum, d) => sum + d.amount, 0);
    const totalWithdrawalsToday = withdrawalsToday.reduce((sum, w) => sum + w.amount, 0);
    const netCashFlow = totalDepositsToday - totalWithdrawalsToday;
    const pendingOps = deposits.filter(d => d.status === 'pending').length +
                       withdrawals.filter(w => w.status === 'pending').length;

    return {
      totalDepositsToday,
      totalWithdrawalsToday,
      netCashFlow,
      pendingOps,
    };
  }, [deposits, withdrawals]);

  // Generate 14-day financial flow data for chart
  const dailyFlowData = useMemo((): DailyFlow[] => {
    const data: DailyFlow[] = [];
    const now = Date.now();
    let runningBalance = 0;

    for (let i = 13; i >= 0; i--) {
      const dayStart = now - (i * 86400000);
      const dayEnd = dayStart + 86400000;

      const dayDeposits = deposits
        .filter(d => d.status === 'completed' && d.date >= dayStart && d.date < dayEnd)
        .reduce((sum, d) => sum + d.amount, 0);

      const dayWithdrawals = withdrawals
        .filter(w => w.status === 'approved' && w.date >= dayStart && w.date < dayEnd)
        .reduce((sum, w) => sum + w.amount, 0);

      runningBalance += (dayDeposits - dayWithdrawals);

      data.push({
        date: new Date(dayStart).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        deposits: dayDeposits,
        withdrawals: dayWithdrawals,
        runningBalance,
      });
    }

    return data;
  }, [deposits, withdrawals]);

  // Filtered deposits
  const filteredDeposits = useMemo(() => {
    let result = [...deposits];

    if (depositStatus !== 'all') {
      result = result.filter(d => d.status === depositStatus);
    }

    if (depositMethod !== 'all') {
      result = result.filter(d => d.method === depositMethod);
    }

    if (depositMinAmount) {
      const min = parseFloat(depositMinAmount);
      if (!isNaN(min)) result = result.filter(d => d.amount >= min);
    }

    if (depositMaxAmount) {
      const max = parseFloat(depositMaxAmount);
      if (!isNaN(max)) result = result.filter(d => d.amount <= max);
    }

    return result.sort((a, b) => b.date - a.date);
  }, [deposits, depositStatus, depositMethod, depositMinAmount, depositMaxAmount]);

  // Filtered withdrawals
  const filteredWithdrawals = useMemo(() => {
    let result = [...withdrawals];

    if (withdrawalStatus !== 'all') {
      result = result.filter(w => w.status === withdrawalStatus);
    }

    if (withdrawalMethod !== 'all') {
      result = result.filter(w => w.method === withdrawalMethod);
    }

    return result.sort((a, b) => b.date - a.date);
  }, [withdrawals, withdrawalStatus, withdrawalMethod]);

  // Filtered transfers
  const filteredTransfers = useMemo(() => {
    let result = [...transfers];

    if (transferType !== 'all') {
      result = result.filter(t => t.type === transferType);
    }

    return result.sort((a, b) => b.date - a.date);
  }, [transfers, transferType]);

  // Status badge helper
  const getDepositStatusBadge = (status: DepositStatus) => {
    switch (status) {
      case 'pending':
        return <span className="px-2 py-0.5 rounded bg-yellow-500/20 text-yellow-400 text-xs font-semibold">Pending</span>;
      case 'completed':
        return <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 text-xs font-semibold">Completed</span>;
      case 'failed':
        return <span className="px-2 py-0.5 rounded bg-rose-500/20 text-rose-400 text-xs font-semibold">Failed</span>;
    }
  };

  const getWithdrawalStatusBadge = (status: WithdrawalStatus) => {
    switch (status) {
      case 'pending':
        return <span className="px-2 py-0.5 rounded bg-yellow-500/20 text-yellow-400 text-xs font-semibold">Pending</span>;
      case 'approved':
        return <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 text-xs font-semibold">Approved</span>;
      case 'rejected':
        return <span className="px-2 py-0.5 rounded bg-rose-500/20 text-rose-400 text-xs font-semibold">Rejected</span>;
    }
  };

  const getTransferTypeBadge = (type: TransferType) => {
    const colors = {
      internal: 'bg-blue-500/20 text-blue-400',
      rebate: 'bg-emerald-500/20 text-emerald-400',
      commission: 'bg-purple-500/20 text-purple-400',
      adjustment: 'bg-orange-500/20 text-orange-400',
    };
    return <span className={`px-2 py-0.5 rounded text-xs font-semibold ${colors[type]}`}>{type}</span>;
  };

  const getRiskScoreColor = (score: number): string => {
    if (score < 30) return 'text-emerald-400';
    if (score < 60) return 'text-yellow-400';
    if (score < 80) return 'text-orange-400';
    return 'text-rose-400';
  };

  const getRiskScoreBg = (score: number): string => {
    if (score < 30) return 'bg-emerald-500/20';
    if (score < 60) return 'bg-yellow-500/20';
    if (score < 80) return 'bg-orange-500/20';
    return 'bg-rose-500/20';
  };

  return (
    <div className="h-full flex flex-col bg-[#18181b] text-zinc-300 overflow-auto custom-scrollbar">
      {/* Summary Cards */}
      <div className="grid grid-cols-4 gap-4 p-4">
        {/* Total Deposits Today */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-emerald-500/20 flex items-center justify-center">
              <TrendingUp size={20} className="text-emerald-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-emerald-400">
                ${stats.totalDepositsToday.toLocaleString()}
              </div>
              <div className="text-xs text-zinc-500">Deposits Today</div>
            </div>
          </div>
        </div>

        {/* Total Withdrawals Today */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-rose-500/20 flex items-center justify-center">
              <TrendingDown size={20} className="text-rose-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-rose-400">
                ${stats.totalWithdrawalsToday.toLocaleString()}
              </div>
              <div className="text-xs text-zinc-500">Withdrawals Today</div>
            </div>
          </div>
        </div>

        {/* Net Cash Flow */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className={`w-10 h-10 rounded ${stats.netCashFlow >= 0 ? 'bg-emerald-500/20' : 'bg-rose-500/20'} flex items-center justify-center`}>
              <DollarSign size={20} className={stats.netCashFlow >= 0 ? 'text-emerald-400' : 'text-rose-400'} />
            </div>
            <div>
              <div className={`text-2xl font-bold ${stats.netCashFlow >= 0 ? 'text-emerald-400' : 'text-rose-400'}`}>
                {stats.netCashFlow >= 0 ? '+' : ''}${stats.netCashFlow.toLocaleString()}
              </div>
              <div className="text-xs text-zinc-500">Net Cash Flow</div>
            </div>
          </div>
        </div>

        {/* Pending Operations */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-yellow-500/20 flex items-center justify-center">
              <Clock size={20} className="text-yellow-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-yellow-400">{stats.pendingOps}</div>
              <div className="text-xs text-zinc-500">Pending Operations</div>
            </div>
          </div>
        </div>
      </div>

      {/* Financial Summary Chart */}
      <div className="mx-4 mb-4 bg-[#27272a] border border-[#3f3f46] rounded p-4">
        <h3 className="text-sm font-bold text-zinc-400 mb-4">14-Day Financial Flow</h3>
        <svg viewBox="0 0 700 200" className="w-full h-40">
          {/* Grid lines */}
          {[0, 1, 2, 3, 4].map((i) => (
            <line
              key={i}
              x1="40"
              y1={20 + i * 40}
              x2="680"
              y2={20 + i * 40}
              stroke="#3f3f46"
              strokeWidth="0.5"
            />
          ))}

          {/* Bars */}
          {dailyFlowData.map((day, index) => {
            const x = 50 + index * 45;
            const maxVal = Math.max(...dailyFlowData.map(d => Math.max(d.deposits, d.withdrawals)));
            const depositHeight = (day.deposits / maxVal) * 140;
            const withdrawalHeight = (day.withdrawals / maxVal) * 140;

            return (
              <g key={index}>
                {/* Deposit bar (green) */}
                <rect
                  x={x}
                  y={180 - depositHeight}
                  width="18"
                  height={depositHeight}
                  fill="#22c55e"
                  opacity="0.7"
                />
                {/* Withdrawal bar (red) */}
                <rect
                  x={x + 20}
                  y={180 - withdrawalHeight}
                  width="18"
                  height={withdrawalHeight}
                  fill="#ef4444"
                  opacity="0.7"
                />
                {/* Date label */}
                <text
                  x={x + 19}
                  y="195"
                  fill="#71717a"
                  fontSize="8"
                  textAnchor="middle"
                >
                  {day.date}
                </text>
              </g>
            );
          })}

          {/* Running balance line overlay */}
          <polyline
            points={dailyFlowData
              .map((day, index) => {
                const x = 50 + index * 45 + 19;
                const maxBalance = Math.max(...dailyFlowData.map(d => Math.abs(d.runningBalance)));
                const y = 100 - (day.runningBalance / maxBalance) * 60;
                return `${x},${y}`;
              })
              .join(' ')}
            fill="none"
            stroke="#3b82f6"
            strokeWidth="2"
          />

          {/* Legend */}
          <g transform="translate(50, 10)">
            <rect x="0" y="0" width="12" height="8" fill="#22c55e" opacity="0.7" />
            <text x="16" y="7" fill="#71717a" fontSize="10">Deposits</text>

            <rect x="80" y="0" width="12" height="8" fill="#ef4444" opacity="0.7" />
            <text x="96" y="7" fill="#71717a" fontSize="10">Withdrawals</text>

            <line x1="160" y1="4" x2="172" y2="4" stroke="#3b82f6" strokeWidth="2" />
            <text x="176" y="7" fill="#71717a" fontSize="10">Running Balance</text>
          </g>
        </svg>
      </div>

      {/* Tab Navigation */}
      <div className="mx-4 mb-4 flex gap-1 border-b border-[#3f3f46]">
        <button
          onClick={() => setActiveTab('deposits')}
          className={`px-4 py-2 text-sm font-semibold transition-colors ${
            activeTab === 'deposits'
              ? 'border-b-2 border-emerald-400 text-emerald-400'
              : 'text-zinc-500 hover:text-zinc-300'
          }`}
        >
          Deposits
        </button>
        <button
          onClick={() => setActiveTab('withdrawals')}
          className={`px-4 py-2 text-sm font-semibold transition-colors ${
            activeTab === 'withdrawals'
              ? 'border-b-2 border-rose-400 text-rose-400'
              : 'text-zinc-500 hover:text-zinc-300'
          }`}
        >
          Withdrawals
        </button>
        <button
          onClick={() => setActiveTab('transfers')}
          className={`px-4 py-2 text-sm font-semibold transition-colors ${
            activeTab === 'transfers'
              ? 'border-b-2 border-blue-400 text-blue-400'
              : 'text-zinc-500 hover:text-zinc-300'
          }`}
        >
          Transfer History
        </button>
      </div>

      {/* Tab Content */}
      <div className="flex-1 px-4 pb-4">
        {activeTab === 'deposits' && (
          <div className="h-full flex flex-col">
            {/* Filters */}
            <div className="flex items-center gap-3 mb-3 flex-wrap">
              <div className="flex items-center gap-2">
                <Filter size={14} className="text-zinc-500" />
                <select
                  value={depositStatus}
                  onChange={(e) => setDepositStatus(e.target.value as any)}
                  className="px-2 py-1 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-emerald-500"
                >
                  <option value="all">All Status</option>
                  <option value="pending">Pending</option>
                  <option value="completed">Completed</option>
                  <option value="failed">Failed</option>
                </select>
              </div>

              <div className="flex items-center gap-2">
                <select
                  value={depositMethod}
                  onChange={(e) => setDepositMethod(e.target.value as any)}
                  className="px-2 py-1 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-emerald-500"
                >
                  <option value="all">All Methods</option>
                  <option value="Bank Wire">Bank Wire</option>
                  <option value="Credit Card">Credit Card</option>
                  <option value="Debit Card">Debit Card</option>
                  <option value="Crypto">Crypto</option>
                  <option value="E-Wallet">E-Wallet</option>
                </select>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="number"
                  value={depositMinAmount}
                  onChange={(e) => setDepositMinAmount(e.target.value)}
                  placeholder="Min Amount"
                  className="w-28 px-2 py-1 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="number"
                  value={depositMaxAmount}
                  onChange={(e) => setDepositMaxAmount(e.target.value)}
                  placeholder="Max Amount"
                  className="w-28 px-2 py-1 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-emerald-500"
                />
              </div>
            </div>

            {/* Deposits Table */}
            <div className="flex-1 bg-[#27272a] border border-[#3f3f46] rounded overflow-hidden">
              <div className="overflow-auto custom-scrollbar h-full">
                <table className="w-full text-xs">
                  <thead className="bg-[#1e1e1e] border-b border-[#3f3f46] sticky top-0">
                    <tr>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">ID</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Client</th>
                      <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Amount</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Currency</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Method</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Status</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Date</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredDeposits.map((deposit) => (
                      <tr
                        key={deposit.id}
                        className="border-b border-[#3f3f46]/30 hover:bg-zinc-800/30"
                      >
                        <td className="px-3 py-2 text-blue-400 font-mono">{deposit.id}</td>
                        <td className="px-3 py-2">
                          <div className="text-zinc-300">{deposit.clientName}</div>
                          <div className="text-[10px] text-zinc-600">{deposit.clientId}</div>
                        </td>
                        <td className="px-3 py-2 text-right font-semibold text-emerald-400">
                          ${deposit.amount.toLocaleString()}
                        </td>
                        <td className="px-3 py-2 text-center text-zinc-400">{deposit.currency}</td>
                        <td className="px-3 py-2">
                          <span className="px-2 py-0.5 rounded bg-zinc-700/50 text-zinc-300">
                            {deposit.method}
                          </span>
                        </td>
                        <td className="px-3 py-2 text-center">{getDepositStatusBadge(deposit.status)}</td>
                        <td className="px-3 py-2 text-zinc-500">
                          {new Date(deposit.date).toLocaleString()}
                        </td>
                        <td className="px-3 py-2">
                          <div className="flex items-center justify-center gap-2">
                            {deposit.status === 'pending' && (
                              <>
                                <button className="px-2 py-1 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-xs">
                                  Process
                                </button>
                                <button className="px-2 py-1 rounded bg-rose-600 hover:bg-rose-700 text-white text-xs">
                                  Reject
                                </button>
                              </>
                            )}
                            <button className="p-1 rounded hover:bg-zinc-700 text-zinc-400">
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

        {activeTab === 'withdrawals' && (
          <div className="h-full flex flex-col">
            {/* Filters */}
            <div className="flex items-center gap-3 mb-3 flex-wrap">
              <div className="flex items-center gap-2">
                <Filter size={14} className="text-zinc-500" />
                <select
                  value={withdrawalStatus}
                  onChange={(e) => setWithdrawalStatus(e.target.value as any)}
                  className="px-2 py-1 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-rose-500"
                >
                  <option value="all">All Status</option>
                  <option value="pending">Pending</option>
                  <option value="approved">Approved</option>
                  <option value="rejected">Rejected</option>
                </select>
              </div>

              <div className="flex items-center gap-2">
                <select
                  value={withdrawalMethod}
                  onChange={(e) => setWithdrawalMethod(e.target.value as any)}
                  className="px-2 py-1 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-rose-500"
                >
                  <option value="all">All Methods</option>
                  <option value="Bank Wire">Bank Wire</option>
                  <option value="Credit Card">Credit Card</option>
                  <option value="Debit Card">Debit Card</option>
                  <option value="Crypto">Crypto</option>
                  <option value="E-Wallet">E-Wallet</option>
                </select>
              </div>
            </div>

            {/* Withdrawals Table */}
            <div className="flex-1 bg-[#27272a] border border-[#3f3f46] rounded overflow-hidden">
              <div className="overflow-auto custom-scrollbar h-full">
                <table className="w-full text-xs">
                  <thead className="bg-[#1e1e1e] border-b border-[#3f3f46] sticky top-0">
                    <tr>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">ID</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Client</th>
                      <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Amount</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Method</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Risk Score</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Auto-Approve</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Status</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Date</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredWithdrawals.map((withdrawal) => (
                      <tr
                        key={withdrawal.id}
                        className={`border-b border-[#3f3f46]/30 hover:bg-zinc-800/30 ${
                          withdrawal.status === 'pending' && !withdrawal.autoApprovable ? 'bg-rose-500/5' : ''
                        }`}
                      >
                        <td className="px-3 py-2 text-blue-400 font-mono">{withdrawal.id}</td>
                        <td className="px-3 py-2">
                          <div className="text-zinc-300">{withdrawal.clientName}</div>
                          <div className="text-[10px] text-zinc-600">{withdrawal.clientId}</div>
                        </td>
                        <td className="px-3 py-2 text-right font-semibold text-rose-400">
                          ${withdrawal.amount.toLocaleString()}
                        </td>
                        <td className="px-3 py-2">
                          <span className="px-2 py-0.5 rounded bg-zinc-700/50 text-zinc-300">
                            {withdrawal.method}
                          </span>
                        </td>
                        <td className="px-3 py-2 text-center">
                          <span
                            className={`px-2 py-0.5 rounded font-semibold ${getRiskScoreBg(
                              withdrawal.riskScore
                            )} ${getRiskScoreColor(withdrawal.riskScore)}`}
                          >
                            {withdrawal.riskScore}
                          </span>
                        </td>
                        <td className="px-3 py-2 text-center">
                          {withdrawal.autoApprovable ? (
                            <CheckCircle size={14} className="text-emerald-400 inline" />
                          ) : (
                            <XCircle size={14} className="text-rose-400 inline" />
                          )}
                        </td>
                        <td className="px-3 py-2 text-center">{getWithdrawalStatusBadge(withdrawal.status)}</td>
                        <td className="px-3 py-2 text-zinc-500">
                          {new Date(withdrawal.date).toLocaleString()}
                        </td>
                        <td className="px-3 py-2">
                          <div className="flex items-center justify-center gap-2">
                            {withdrawal.status === 'pending' && (
                              <>
                                <button className="px-2 py-1 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-xs">
                                  Approve
                                </button>
                                <button className="px-2 py-1 rounded bg-rose-600 hover:bg-rose-700 text-white text-xs">
                                  Reject
                                </button>
                              </>
                            )}
                            <button className="p-1 rounded hover:bg-zinc-700 text-zinc-400">
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

        {activeTab === 'transfers' && (
          <div className="h-full flex flex-col">
            {/* Filters */}
            <div className="flex items-center gap-3 mb-3 flex-wrap">
              <div className="flex items-center gap-2">
                <Filter size={14} className="text-zinc-500" />
                <select
                  value={transferType}
                  onChange={(e) => setTransferType(e.target.value as any)}
                  className="px-2 py-1 text-xs bg-[#27272a] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                >
                  <option value="all">All Types</option>
                  <option value="internal">Internal</option>
                  <option value="rebate">Rebate</option>
                  <option value="commission">Commission</option>
                  <option value="adjustment">Adjustment</option>
                </select>
              </div>
            </div>

            {/* Transfers Table */}
            <div className="flex-1 bg-[#27272a] border border-[#3f3f46] rounded overflow-hidden">
              <div className="overflow-auto custom-scrollbar h-full">
                <table className="w-full text-xs">
                  <thead className="bg-[#1e1e1e] border-b border-[#3f3f46] sticky top-0">
                    <tr>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">ID</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">From Account</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">
                        <ArrowRightLeft size={12} className="inline" />
                      </th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">To Account</th>
                      <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Amount</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Type</th>
                      <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Date</th>
                      <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredTransfers.map((transfer) => (
                      <tr
                        key={transfer.id}
                        className="border-b border-[#3f3f46]/30 hover:bg-zinc-800/30"
                      >
                        <td className="px-3 py-2 text-blue-400 font-mono">{transfer.id}</td>
                        <td className="px-3 py-2">
                          <div className="text-zinc-300">{transfer.fromClientName}</div>
                          <div className="text-[10px] text-zinc-600">{transfer.fromAccount}</div>
                        </td>
                        <td className="px-3 py-2 text-center text-zinc-500">→</td>
                        <td className="px-3 py-2">
                          <div className="text-zinc-300">{transfer.toClientName}</div>
                          <div className="text-[10px] text-zinc-600">{transfer.toAccount}</div>
                        </td>
                        <td className="px-3 py-2 text-right font-semibold text-blue-400">
                          ${transfer.amount.toLocaleString()}
                        </td>
                        <td className="px-3 py-2 text-center">{getTransferTypeBadge(transfer.type)}</td>
                        <td className="px-3 py-2 text-zinc-500">
                          {new Date(transfer.date).toLocaleString()}
                        </td>
                        <td className="px-3 py-2 text-center">
                          <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 text-xs font-semibold">
                            {transfer.status}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
