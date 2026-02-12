/**
 * Withdrawal Approval Queue
 * Admin interface for approving/rejecting withdrawal requests
 */

'use client';

import React, { useState, useMemo, useEffect } from 'react';
import {
  CheckCircle,
  XCircle,
  Eye,
  ArrowUpDown,
  ChevronLeft,
  ChevronRight,
  X,
  AlertTriangle,
  TrendingUp,
  Clock,
  User,
  CreditCard,
  History,
  Shield,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// Types
type WithdrawalMethod = 'Card' | 'Bank' | 'Crypto';
type WithdrawalStatus = 'pending' | 'approved' | 'rejected';

interface WithdrawalRequest {
  id: string;
  accountId: string;
  userEmail: string;
  amount: number;
  method: WithdrawalMethod;
  riskScore: number;
  flags: string[];
  requestedAt: number;
  status: WithdrawalStatus;
  approvedBy?: string;
  rejectedBy?: string;
  processedAt?: number;
  rejectionReason?: string;
}

interface UserProfile {
  name: string;
  email: string;
  accountAge: number; // in days
  kycStatus: 'verified' | 'pending' | 'rejected';
  totalDeposits: number;
  totalWithdrawals: number;
  depositWithdrawalRatio: number;
}

interface Transaction {
  id: string;
  type: 'deposit' | 'withdrawal';
  amount: number;
  method: string;
  status: string;
  timestamp: number;
}

type SortColumn = 'id' | 'amount' | 'riskScore' | 'requestedAt';
type SortDirection = 'asc' | 'desc';

// Mock data for initial state
const MOCK_PENDING: WithdrawalRequest[] = [
  {
    id: 'WD-2026-001',
    accountId: '680962851',
    userEmail: 'trader@example.com',
    amount: 5000,
    method: 'Bank',
    riskScore: 15,
    flags: [],
    requestedAt: Date.now() - 3600000,
    status: 'pending',
  },
  {
    id: 'WD-2026-002',
    accountId: '680962852',
    userEmail: 'john.doe@example.com',
    amount: 25000,
    method: 'Crypto',
    riskScore: 65,
    flags: ['large_amount', 'ip_change'],
    requestedAt: Date.now() - 7200000,
    status: 'pending',
  },
  {
    id: 'WD-2026-003',
    accountId: '680962853',
    userEmail: 'alice@example.com',
    amount: 1500,
    method: 'Card',
    riskScore: 85,
    flags: ['velocity', 'new_account'],
    requestedAt: Date.now() - 1800000,
    status: 'pending',
  },
];

const MOCK_HISTORY: WithdrawalRequest[] = [
  {
    id: 'WD-2026-000',
    accountId: '680962850',
    userEmail: 'previous@example.com',
    amount: 3000,
    method: 'Bank',
    riskScore: 20,
    flags: [],
    requestedAt: Date.now() - 86400000,
    status: 'approved',
    approvedBy: 'Admin User',
    processedAt: Date.now() - 82800000,
  },
];

export default function WithdrawalQueue() {
  const [pending, setPending] = useState<WithdrawalRequest[]>(MOCK_PENDING);
  const [history, setHistory] = useState<WithdrawalRequest[]>(MOCK_HISTORY);
  const [selectedDetail, setSelectedDetail] = useState<string | null>(null);
  const [approveModal, setApproveModal] = useState<WithdrawalRequest | null>(null);
  const [rejectModal, setRejectModal] = useState<WithdrawalRequest | null>(null);
  const [rejectionReason, setRejectionReason] = useState('');

  // Filters
  const [filterStatus, setFilterStatus] = useState<'all' | WithdrawalStatus>('all');
  const [filterMethod, setFilterMethod] = useState<'all' | WithdrawalMethod>('all');
  const [minAmount, setMinAmount] = useState<string>('');
  const [maxAmount, setMaxAmount] = useState<string>('');

  // Sorting
  const [sortColumn, setSortColumn] = useState<SortColumn>('requestedAt');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  // Pagination
  const [historyPage, setHistoryPage] = useState(0);
  const historyPerPage = 20;

  // Summary stats
  const stats = useMemo(() => {
    const now = Date.now();
    const oneDayAgo = now - 86400000;

    return {
      pendingCount: pending.length,
      approvedToday: history.filter(
        (w) => w.status === 'approved' && w.processedAt && w.processedAt >= oneDayAgo
      ).length,
      rejectedToday: history.filter(
        (w) => w.status === 'rejected' && w.processedAt && w.processedAt >= oneDayAgo
      ).length,
    };
  }, [pending, history]);

  // Filtered pending
  const filteredPending = useMemo(() => {
    let result = [...pending];

    if (filterMethod !== 'all') {
      result = result.filter((w) => w.method === filterMethod);
    }

    if (minAmount) {
      const min = parseFloat(minAmount);
      if (!isNaN(min)) {
        result = result.filter((w) => w.amount >= min);
      }
    }

    if (maxAmount) {
      const max = parseFloat(maxAmount);
      if (!isNaN(max)) {
        result = result.filter((w) => w.amount <= max);
      }
    }

    return result;
  }, [pending, filterMethod, minAmount, maxAmount]);

  // Sorted pending
  const sortedPending = useMemo(() => {
    return [...filteredPending].sort((a, b) => {
      let aVal: any = a[sortColumn];
      let bVal: any = b[sortColumn];

      if (sortColumn === 'id') {
        aVal = a.id;
        bVal = b.id;
      }

      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortDirection === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
      }

      return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
    });
  }, [filteredPending, sortColumn, sortDirection]);

  // Filtered history
  const filteredHistory = useMemo(() => {
    let result = [...history];

    if (filterStatus !== 'all') {
      result = result.filter((w) => w.status === filterStatus);
    }

    if (filterMethod !== 'all') {
      result = result.filter((w) => w.method === filterMethod);
    }

    return result;
  }, [history, filterStatus, filterMethod]);

  // Paginated history
  const paginatedHistory = useMemo(() => {
    const start = historyPage * historyPerPage;
    return filteredHistory.slice(start, start + historyPerPage);
  }, [filteredHistory, historyPage]);

  const totalHistoryPages = Math.ceil(filteredHistory.length / historyPerPage);

  // Handle sort
  const handleSort = (column: SortColumn) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('desc');
    }
  };

  // Handle approve
  const handleApprove = async (request: WithdrawalRequest) => {
    try {
      // In production: POST to API_CONFIG.ADMIN_WITHDRAWALS_APPROVE(request.id)
      // await fetch(`${API_CONFIG.ADMIN_WITHDRAWALS}/${request.id}/approve`, { method: 'POST' });

      // Update local state
      const updatedRequest: WithdrawalRequest = {
        ...request,
        status: 'approved',
        approvedBy: 'Admin User',
        processedAt: Date.now(),
      };

      setPending((prev) => prev.filter((w) => w.id !== request.id));
      setHistory((prev) => [updatedRequest, ...prev]);
      setApproveModal(null);
    } catch (error) {
      console.error('Failed to approve withdrawal:', error);
    }
  };

  // Handle reject
  const handleReject = async (request: WithdrawalRequest) => {
    if (!rejectionReason.trim()) {
      alert('Rejection reason is required');
      return;
    }

    try {
      // In production: POST to API_CONFIG.ADMIN_WITHDRAWALS_REJECT(request.id)
      // await fetch(`${API_CONFIG.ADMIN_WITHDRAWALS}/${request.id}/reject`, {
      //   method: 'POST',
      //   body: JSON.stringify({ reason: rejectionReason }),
      // });

      // Update local state
      const updatedRequest: WithdrawalRequest = {
        ...request,
        status: 'rejected',
        rejectedBy: 'Admin User',
        processedAt: Date.now(),
        rejectionReason,
      };

      setPending((prev) => prev.filter((w) => w.id !== request.id));
      setHistory((prev) => [updatedRequest, ...prev]);
      setRejectModal(null);
      setRejectionReason('');
    } catch (error) {
      console.error('Failed to reject withdrawal:', error);
    }
  };

  // Get risk score color
  const getRiskScoreColor = (score: number): string => {
    if (score < 30) return 'text-emerald-400';
    if (score < 60) return 'text-yellow-400';
    if (score < 80) return 'text-orange-400';
    return 'text-rose-400';
  };

  // Get risk score bg
  const getRiskScoreBg = (score: number): string => {
    if (score < 30) return 'bg-emerald-500/20';
    if (score < 60) return 'bg-yellow-500/20';
    if (score < 80) return 'bg-orange-500/20';
    return 'bg-rose-500/20';
  };

  // Format flag
  const formatFlag = (flag: string): string => {
    return flag
      .split('_')
      .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
      .join(' ');
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-zinc-300 p-4 overflow-auto custom-scrollbar">
      {/* Summary Cards */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        {/* Pending */}
        <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-yellow-500/20 flex items-center justify-center">
              <Clock size={20} className="text-yellow-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-yellow-400">{stats.pendingCount}</div>
              <div className="text-xs text-zinc-500">Pending Approvals</div>
            </div>
          </div>
        </div>

        {/* Approved Today */}
        <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-emerald-500/20 flex items-center justify-center">
              <CheckCircle size={20} className="text-emerald-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-emerald-400">{stats.approvedToday}</div>
              <div className="text-xs text-zinc-500">Approved Today</div>
            </div>
          </div>
        </div>

        {/* Rejected Today */}
        <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-rose-500/20 flex items-center justify-center">
              <XCircle size={20} className="text-rose-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-rose-400">{stats.rejectedToday}</div>
              <div className="text-xs text-zinc-500">Rejected Today</div>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <div className="flex items-center gap-2">
          <label className="text-xs text-zinc-500">Method:</label>
          <select
            value={filterMethod}
            onChange={(e) => setFilterMethod(e.target.value as any)}
            className="px-2 py-1 text-xs bg-[#252525] border border-zinc-700 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
          >
            <option value="all">All</option>
            <option value="Card">Card</option>
            <option value="Bank">Bank</option>
            <option value="Crypto">Crypto</option>
          </select>
        </div>

        <div className="flex items-center gap-2">
          <label className="text-xs text-zinc-500">Min Amount:</label>
          <input
            type="number"
            value={minAmount}
            onChange={(e) => setMinAmount(e.target.value)}
            placeholder="0"
            className="w-24 px-2 py-1 text-xs bg-[#252525] border border-zinc-700 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
          />
        </div>

        <div className="flex items-center gap-2">
          <label className="text-xs text-zinc-500">Max Amount:</label>
          <input
            type="number"
            value={maxAmount}
            onChange={(e) => setMaxAmount(e.target.value)}
            placeholder="∞"
            className="w-24 px-2 py-1 text-xs bg-[#252525] border border-zinc-700 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
          />
        </div>
      </div>

      {/* Pending Queue Table */}
      <div className="mb-6">
        <h2 className="text-sm font-bold text-zinc-400 mb-3 flex items-center gap-2">
          <AlertTriangle size={14} className="text-yellow-400" />
          Pending Queue ({sortedPending.length})
        </h2>
        <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded overflow-hidden">
          <table className="w-full text-xs">
            <thead className="bg-[#252525] border-b border-zinc-700">
              <tr>
                <th
                  className="px-3 py-2 text-left text-zinc-500 font-semibold cursor-pointer hover:text-zinc-300"
                  onClick={() => handleSort('id')}
                >
                  <div className="flex items-center gap-1">
                    Request ID
                    <ArrowUpDown size={12} />
                  </div>
                </th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Account</th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">User Email</th>
                <th
                  className="px-3 py-2 text-right text-zinc-500 font-semibold cursor-pointer hover:text-zinc-300"
                  onClick={() => handleSort('amount')}
                >
                  <div className="flex items-center gap-1 justify-end">
                    Amount
                    <ArrowUpDown size={12} />
                  </div>
                </th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Method</th>
                <th
                  className="px-3 py-2 text-center text-zinc-500 font-semibold cursor-pointer hover:text-zinc-300"
                  onClick={() => handleSort('riskScore')}
                >
                  <div className="flex items-center gap-1 justify-center">
                    Risk Score
                    <ArrowUpDown size={12} />
                  </div>
                </th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Flags</th>
                <th
                  className="px-3 py-2 text-left text-zinc-500 font-semibold cursor-pointer hover:text-zinc-300"
                  onClick={() => handleSort('requestedAt')}
                >
                  <div className="flex items-center gap-1">
                    Requested At
                    <ArrowUpDown size={12} />
                  </div>
                </th>
                <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Actions</th>
              </tr>
            </thead>
            <tbody>
              {sortedPending.length === 0 ? (
                <tr>
                  <td colSpan={9} className="px-3 py-8 text-center text-zinc-500">
                    No pending withdrawals
                  </td>
                </tr>
              ) : (
                sortedPending.map((request) => (
                  <tr
                    key={request.id}
                    className="border-b border-zinc-700/30 hover:bg-zinc-800/30"
                  >
                    <td className="px-3 py-2 text-blue-400 font-mono">{request.id}</td>
                    <td className="px-3 py-2 text-zinc-400">{request.accountId}</td>
                    <td className="px-3 py-2 text-zinc-400">{request.userEmail}</td>
                    <td className="px-3 py-2 text-right font-semibold text-emerald-400">
                      ${request.amount.toLocaleString()}
                    </td>
                    <td className="px-3 py-2">
                      <span className="px-2 py-0.5 rounded bg-zinc-700/50 text-zinc-300">
                        {request.method}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-center">
                      <span
                        className={`px-2 py-0.5 rounded font-semibold ${getRiskScoreBg(
                          request.riskScore
                        )} ${getRiskScoreColor(request.riskScore)}`}
                      >
                        {request.riskScore}
                      </span>
                    </td>
                    <td className="px-3 py-2">
                      <div className="flex flex-wrap gap-1">
                        {request.flags.length === 0 ? (
                          <span className="text-zinc-600 text-xs">None</span>
                        ) : (
                          request.flags.map((flag) => (
                            <span
                              key={flag}
                              className="px-1.5 py-0.5 rounded bg-orange-500/20 text-orange-400 text-[10px]"
                            >
                              {formatFlag(flag)}
                            </span>
                          ))
                        )}
                      </div>
                    </td>
                    <td className="px-3 py-2 text-zinc-500">
                      {new Date(request.requestedAt).toLocaleString()}
                    </td>
                    <td className="px-3 py-2">
                      <div className="flex items-center justify-center gap-2">
                        <button
                          onClick={() => setApproveModal(request)}
                          className="px-2 py-1 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-medium"
                        >
                          Approve
                        </button>
                        <button
                          onClick={() => setRejectModal(request)}
                          className="px-2 py-1 rounded bg-rose-600 hover:bg-rose-700 text-white text-xs font-medium"
                        >
                          Reject
                        </button>
                        <button
                          onClick={() => setSelectedDetail(request.id)}
                          className="p-1 rounded hover:bg-zinc-700 text-zinc-400"
                        >
                          <Eye size={14} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* History Table */}
      <div>
        <h2 className="text-sm font-bold text-zinc-400 mb-3 flex items-center gap-2">
          <History size={14} className="text-blue-400" />
          Processing History ({filteredHistory.length})
        </h2>

        {/* History Filters */}
        <div className="flex items-center gap-3 mb-3">
          <div className="flex items-center gap-2">
            <label className="text-xs text-zinc-500">Status:</label>
            <select
              value={filterStatus}
              onChange={(e) => {
                setFilterStatus(e.target.value as any);
                setHistoryPage(0);
              }}
              className="px-2 py-1 text-xs bg-[#252525] border border-zinc-700 rounded text-zinc-300 focus:outline-none focus:border-blue-500"
            >
              <option value="all">All</option>
              <option value="approved">Approved</option>
              <option value="rejected">Rejected</option>
            </select>
          </div>
        </div>

        <div className="bg-[#1e1e1e] border border-zinc-700/50 rounded overflow-hidden">
          <table className="w-full text-xs">
            <thead className="bg-[#252525] border-b border-zinc-700">
              <tr>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Request ID</th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">User Email</th>
                <th className="px-3 py-2 text-right text-zinc-500 font-semibold">Amount</th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Method</th>
                <th className="px-3 py-2 text-center text-zinc-500 font-semibold">Status</th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Processed By</th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Processed At</th>
                <th className="px-3 py-2 text-left text-zinc-500 font-semibold">Reason</th>
              </tr>
            </thead>
            <tbody>
              {paginatedHistory.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-3 py-8 text-center text-zinc-500">
                    No history records
                  </td>
                </tr>
              ) : (
                paginatedHistory.map((request) => (
                  <tr
                    key={request.id}
                    className="border-b border-zinc-700/30 hover:bg-zinc-800/30"
                  >
                    <td className="px-3 py-2 text-blue-400 font-mono">{request.id}</td>
                    <td className="px-3 py-2 text-zinc-400">{request.userEmail}</td>
                    <td className="px-3 py-2 text-right font-semibold text-zinc-300">
                      ${request.amount.toLocaleString()}
                    </td>
                    <td className="px-3 py-2">
                      <span className="px-2 py-0.5 rounded bg-zinc-700/50 text-zinc-300">
                        {request.method}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-center">
                      {request.status === 'approved' ? (
                        <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 font-semibold">
                          Approved
                        </span>
                      ) : (
                        <span className="px-2 py-0.5 rounded bg-rose-500/20 text-rose-400 font-semibold">
                          Rejected
                        </span>
                      )}
                    </td>
                    <td className="px-3 py-2 text-zinc-400">
                      {request.approvedBy || request.rejectedBy || 'N/A'}
                    </td>
                    <td className="px-3 py-2 text-zinc-500">
                      {request.processedAt
                        ? new Date(request.processedAt).toLocaleString()
                        : 'N/A'}
                    </td>
                    <td className="px-3 py-2 text-zinc-400 max-w-xs truncate">
                      {request.rejectionReason || '-'}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>

          {/* Pagination */}
          {totalHistoryPages > 1 && (
            <div className="flex items-center justify-between px-3 py-2 bg-[#252525] border-t border-zinc-700">
              <div className="text-xs text-zinc-500">
                Page {historyPage + 1} of {totalHistoryPages}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setHistoryPage((p) => Math.max(0, p - 1))}
                  disabled={historyPage === 0}
                  className="p-1 rounded hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed text-zinc-400"
                >
                  <ChevronLeft size={14} />
                </button>
                <button
                  onClick={() => setHistoryPage((p) => Math.min(totalHistoryPages - 1, p + 1))}
                  disabled={historyPage >= totalHistoryPages - 1}
                  className="p-1 rounded hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed text-zinc-400"
                >
                  <ChevronRight size={14} />
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Approve Modal */}
      {approveModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg p-6 w-[400px]">
            <h3 className="text-lg font-bold text-zinc-200 mb-4">Confirm Approval</h3>
            <p className="text-sm text-zinc-400 mb-6">
              Approve withdrawal of{' '}
              <span className="font-bold text-emerald-400">
                ${approveModal.amount.toLocaleString()}
              </span>{' '}
              for user <span className="font-semibold text-blue-400">{approveModal.userEmail}</span>
              ?
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => setApproveModal(null)}
                className="px-4 py-2 rounded bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-sm"
              >
                Cancel
              </button>
              <button
                onClick={() => handleApprove(approveModal)}
                className="px-4 py-2 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-medium"
              >
                Approve
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Reject Modal */}
      {rejectModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg p-6 w-[400px]">
            <h3 className="text-lg font-bold text-zinc-200 mb-4">Reject Withdrawal</h3>
            <p className="text-sm text-zinc-400 mb-4">
              Enter rejection reason for{' '}
              <span className="font-semibold text-blue-400">{rejectModal.userEmail}</span>:
            </p>
            <textarea
              value={rejectionReason}
              onChange={(e) => setRejectionReason(e.target.value)}
              placeholder="Enter rejection reason (required)"
              className="w-full px-3 py-2 bg-[#252525] border border-zinc-700 rounded text-zinc-300 text-sm focus:outline-none focus:border-rose-500 resize-none"
              rows={4}
            />
            <div className="flex gap-3 justify-end mt-4">
              <button
                onClick={() => {
                  setRejectModal(null);
                  setRejectionReason('');
                }}
                className="px-4 py-2 rounded bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-sm"
              >
                Cancel
              </button>
              <button
                onClick={() => handleReject(rejectModal)}
                className="px-4 py-2 rounded bg-rose-600 hover:bg-rose-700 text-white text-sm font-medium"
              >
                Reject
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Detail Panel (Slide-out) */}
      {selectedDetail && (
        <DetailPanel
          requestId={selectedDetail}
          onClose={() => setSelectedDetail(null)}
        />
      )}
    </div>
  );
}

// Detail Panel Component
interface DetailPanelProps {
  requestId: string;
  onClose: () => void;
}

function DetailPanel({ requestId, onClose }: DetailPanelProps) {
  // Mock data - in production, fetch from API based on requestId
  const userProfile: UserProfile = {
    name: 'John Doe',
    email: 'john.doe@example.com',
    accountAge: 120,
    kycStatus: 'verified',
    totalDeposits: 50000,
    totalWithdrawals: 15000,
    depositWithdrawalRatio: 3.33,
  };

  const recentTransactions: Transaction[] = [
    {
      id: 'TX-001',
      type: 'deposit',
      amount: 5000,
      method: 'Bank',
      status: 'completed',
      timestamp: Date.now() - 86400000,
    },
    {
      id: 'TX-002',
      type: 'withdrawal',
      amount: 2000,
      method: 'Bank',
      status: 'completed',
      timestamp: Date.now() - 172800000,
    },
  ];

  return (
    <div className="fixed inset-y-0 right-0 w-[400px] bg-[#1e1e1e] border-l border-zinc-700 z-50 flex flex-col shadow-2xl">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-[#252525] border-b border-zinc-700">
        <h3 className="text-sm font-bold text-zinc-200">Withdrawal Details</h3>
        <button
          onClick={onClose}
          className="p-1 rounded hover:bg-zinc-700 text-zinc-400"
        >
          <X size={16} />
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-6">
        {/* User Profile */}
        <div>
          <h4 className="text-xs font-bold text-zinc-400 mb-3 flex items-center gap-2">
            <User size={12} />
            User Profile
          </h4>
          <div className="bg-[#252525] border border-zinc-700/50 rounded p-3 space-y-2 text-xs">
            <div className="flex justify-between">
              <span className="text-zinc-500">Name:</span>
              <span className="text-zinc-300 font-medium">{userProfile.name}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-500">Email:</span>
              <span className="text-zinc-300 font-medium">{userProfile.email}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-500">Account Age:</span>
              <span className="text-zinc-300">{userProfile.accountAge} days</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-500">KYC Status:</span>
              <span
                className={`px-2 py-0.5 rounded ${
                  userProfile.kycStatus === 'verified'
                    ? 'bg-emerald-500/20 text-emerald-400'
                    : 'bg-yellow-500/20 text-yellow-400'
                }`}
              >
                {userProfile.kycStatus}
              </span>
            </div>
          </div>
        </div>

        {/* Transaction Summary */}
        <div>
          <h4 className="text-xs font-bold text-zinc-400 mb-3 flex items-center gap-2">
            <TrendingUp size={12} />
            Transaction Summary
          </h4>
          <div className="bg-[#252525] border border-zinc-700/50 rounded p-3 space-y-2 text-xs">
            <div className="flex justify-between">
              <span className="text-zinc-500">Total Deposits:</span>
              <span className="text-emerald-400 font-semibold">
                ${userProfile.totalDeposits.toLocaleString()}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-500">Total Withdrawals:</span>
              <span className="text-rose-400 font-semibold">
                ${userProfile.totalWithdrawals.toLocaleString()}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-500">D/W Ratio:</span>
              <span className="text-zinc-300 font-semibold">
                {userProfile.depositWithdrawalRatio.toFixed(2)}x
              </span>
            </div>
          </div>
        </div>

        {/* Recent Transactions */}
        <div>
          <h4 className="text-xs font-bold text-zinc-400 mb-3 flex items-center gap-2">
            <CreditCard size={12} />
            Recent Transactions (Last 5)
          </h4>
          <div className="space-y-2">
            {recentTransactions.map((tx) => (
              <div
                key={tx.id}
                className="bg-[#252525] border border-zinc-700/50 rounded p-2 text-xs"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-blue-400 font-mono">{tx.id}</span>
                  <span
                    className={`px-1.5 py-0.5 rounded text-[10px] ${
                      tx.type === 'deposit'
                        ? 'bg-emerald-500/20 text-emerald-400'
                        : 'bg-rose-500/20 text-rose-400'
                    }`}
                  >
                    {tx.type}
                  </span>
                </div>
                <div className="flex items-center justify-between text-zinc-500">
                  <span>
                    {tx.method} • {new Date(tx.timestamp).toLocaleDateString()}
                  </span>
                  <span className="text-zinc-300 font-semibold">
                    ${tx.amount.toLocaleString()}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Risk Factors */}
        <div>
          <h4 className="text-xs font-bold text-zinc-400 mb-3 flex items-center gap-2">
            <Shield size={12} />
            Risk Factors
          </h4>
          <div className="bg-[#252525] border border-zinc-700/50 rounded p-3 space-y-2 text-xs">
            <div className="flex justify-between">
              <span className="text-zinc-500">IP Change Detected:</span>
              <span className="text-yellow-400">Yes</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-500">Velocity Check:</span>
              <span className="text-emerald-400">Pass</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-500">Same Method:</span>
              <span className="text-emerald-400">Yes</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-500">Duplicate Request:</span>
              <span className="text-emerald-400">No</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
