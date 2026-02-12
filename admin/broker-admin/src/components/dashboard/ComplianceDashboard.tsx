'use client';

import React, { useState, useMemo } from 'react';
import {
  Shield,
  CheckCircle,
  AlertTriangle,
  FileText,
  TrendingUp,
  Bell,
  Filter,
  Download,
  Eye,
  Calendar,
} from 'lucide-react';

type KYCStatus = 'verified' | 'pending' | 'expired' | 'rejected';
type RiskLevel = 'low' | 'medium' | 'high';
type ReportType = 'SAR' | 'CTR' | 'FATCA';
type ReportStatus = 'filed' | 'pending' | 'overdue';

interface KYCClient {
  id: string;
  name: string;
  status: KYCStatus;
  verifiedDate: number | null;
  expiryDate: number | null;
}

interface Transaction {
  id: string;
  clientName: string;
  amount: number;
  type: 'deposit' | 'withdrawal' | 'transfer';
  date: number;
  flagged: boolean;
  reviewed: boolean;
}

interface RegulatoryReport {
  id: string;
  type: ReportType;
  period: string;
  status: ReportStatus;
  filedDate: number | null;
  dueDate: number;
}

interface RiskDistribution {
  low: number;
  medium: number;
  high: number;
}

export default function ComplianceDashboard() {
  const [selectedRiskFilter, setSelectedRiskFilter] = useState<RiskLevel | 'all'>('all');
  const [transactionDateFrom, setTransactionDateFrom] = useState<string>(
    new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  );
  const [transactionDateTo, setTransactionDateTo] = useState<string>(new Date().toISOString().split('T')[0]);
  const [amountThreshold, setAmountThreshold] = useState<number>(10000);
  const [showFlaggedOnly, setShowFlaggedOnly] = useState<boolean>(false);

  // Mock data: Compliance Score
  const complianceScore = 87; // 0-100
  const lastAuditDate = Date.now() - 45 * 24 * 60 * 60 * 1000;
  const nextReviewDate = Date.now() + 30 * 24 * 60 * 60 * 1000;

  // Mock data: KYC Clients (50 clients)
  const [kycClients, setKycClients] = useState<KYCClient[]>(() => {
    const clients: KYCClient[] = [];
    const names = ['John Doe', 'Jane Smith', 'Alice Johnson', 'Bob Wilson', 'Charlie Brown', 'Eva Martinez', 'Frank Lee', 'Grace Chen', 'Henry Davis', 'Ivy Rodriguez'];
    const statuses: KYCStatus[] = ['verified', 'pending', 'expired', 'rejected'];

    for (let i = 0; i < 50; i++) {
      const status = statuses[Math.floor(Math.random() * statuses.length)];
      const verifiedDate = status === 'verified' || status === 'expired'
        ? Date.now() - Math.floor(Math.random() * 365) * 24 * 60 * 60 * 1000
        : null;
      const expiryDate = status === 'verified'
        ? Date.now() + Math.floor(Math.random() * 60 - 10) * 24 * 60 * 60 * 1000 // some expiring soon
        : status === 'expired'
        ? Date.now() - Math.floor(Math.random() * 30) * 24 * 60 * 60 * 1000
        : null;

      clients.push({
        id: `kyc-${i + 1}`,
        name: `${names[i % names.length]} ${i + 1}`,
        status,
        verifiedDate,
        expiryDate,
      });
    }
    return clients;
  });

  // Mock data: Transactions (40 transactions)
  const [transactions, setTransactions] = useState<Transaction[]>(() => {
    const txns: Transaction[] = [];
    const clients = ['John Doe', 'Jane Smith', 'Alice Johnson', 'Bob Wilson', 'Charlie Brown'];
    const types: ('deposit' | 'withdrawal' | 'transfer')[] = ['deposit', 'withdrawal', 'transfer'];

    for (let i = 0; i < 40; i++) {
      const amount = Math.floor(Math.random() * 50000) + 5000;
      const flagged = amount > 25000 || Math.random() > 0.8;

      txns.push({
        id: `txn-${i + 1}`,
        clientName: clients[Math.floor(Math.random() * clients.length)],
        amount,
        type: types[Math.floor(Math.random() * types.length)],
        date: Date.now() - Math.floor(Math.random() * 60) * 24 * 60 * 60 * 1000,
        flagged,
        reviewed: flagged ? Math.random() > 0.5 : true,
      });
    }
    return txns;
  });

  // Mock data: Regulatory Reports (12 reports)
  const [regulatoryReports] = useState<RegulatoryReport[]>(() => {
    const reports: RegulatoryReport[] = [];
    const types: ReportType[] = ['SAR', 'CTR', 'FATCA'];
    const statuses: ReportStatus[] = ['filed', 'pending', 'overdue'];

    for (let i = 0; i < 12; i++) {
      const status = statuses[Math.floor(Math.random() * statuses.length)];
      const dueDate = Date.now() - Math.floor(Math.random() * 90 - 30) * 24 * 60 * 60 * 1000;
      const filedDate = status === 'filed'
        ? dueDate - Math.floor(Math.random() * 7) * 24 * 60 * 60 * 1000
        : null;

      reports.push({
        id: `report-${i + 1}`,
        type: types[Math.floor(Math.random() * types.length)],
        period: `Q${Math.floor(i / 3) + 1} 2026`,
        status,
        filedDate,
        dueDate,
      });
    }
    return reports;
  });

  // Mock data: Risk Distribution
  const riskDistribution: RiskDistribution = {
    low: 25,
    medium: 18,
    high: 7,
  };

  // Calculate KYC stats
  const kycStats = useMemo(() => {
    const total = kycClients.length;
    const verified = kycClients.filter((c) => c.status === 'verified').length;
    const pending = kycClients.filter((c) => c.status === 'pending').length;
    const expired = kycClients.filter((c) => c.status === 'expired').length;
    const rejected = kycClients.filter((c) => c.status === 'rejected').length;
    const completionRate = total > 0 ? (verified / total) * 100 : 0;

    return { total, verified, pending, expired, rejected, completionRate };
  }, [kycClients]);

  // Clients expiring within 30 days
  const expiringClients = useMemo(() => {
    const now = Date.now();
    const thirtyDaysFromNow = now + 30 * 24 * 60 * 60 * 1000;
    return kycClients
      .filter((c) => c.status === 'verified' && c.expiryDate && c.expiryDate > now && c.expiryDate <= thirtyDaysFromNow)
      .sort((a, b) => (a.expiryDate || 0) - (b.expiryDate || 0))
      .slice(0, 10);
  }, [kycClients]);

  // AML Alerts
  const amlAlerts = useMemo(() => {
    return transactions.filter((t) => t.flagged && !t.reviewed).length;
  }, [transactions]);

  // Suspicious Transactions this month
  const suspiciousTransactions = useMemo(() => {
    const oneMonthAgo = Date.now() - 30 * 24 * 60 * 60 * 1000;
    return transactions.filter((t) => t.flagged && t.date >= oneMonthAgo).length;
  }, [transactions]);

  // Regulatory Reports this quarter
  const reportsFiled = useMemo(() => {
    return regulatoryReports.filter((r) => r.status === 'filed').length;
  }, [regulatoryReports]);

  // Filtered transactions
  const filteredTransactions = useMemo(() => {
    return transactions.filter((t) => {
      const dateFrom = new Date(transactionDateFrom).getTime();
      const dateTo = new Date(transactionDateTo).getTime() + 24 * 60 * 60 * 1000; // end of day
      if (t.date < dateFrom || t.date > dateTo) return false;
      if (t.amount < amountThreshold) return false;
      if (showFlaggedOnly && !t.flagged) return false;
      return true;
    });
  }, [transactions, transactionDateFrom, transactionDateTo, amountThreshold, showFlaggedOnly]);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value);
  };

  const formatDate = (timestamp: number) => {
    return new Date(timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const getComplianceColor = (score: number) => {
    if (score > 80) return { text: 'text-emerald-400', bg: 'bg-emerald-500' };
    if (score > 60) return { text: 'text-yellow-400', bg: 'bg-yellow-500' };
    return { text: 'text-rose-400', bg: 'bg-rose-500' };
  };

  const getKYCBadge = (status: KYCStatus) => {
    const badges = {
      verified: 'bg-emerald-500/20 text-emerald-400',
      pending: 'bg-yellow-500/20 text-yellow-400',
      expired: 'bg-rose-500/20 text-rose-400',
      rejected: 'bg-zinc-700 text-zinc-500',
    };
    return <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${badges[status]}`}>{status.toUpperCase()}</span>;
  };

  const getReportStatusBadge = (status: ReportStatus) => {
    const badges = {
      filed: 'bg-emerald-500/20 text-emerald-400',
      pending: 'bg-yellow-500/20 text-yellow-400',
      overdue: 'bg-rose-500/20 text-rose-400',
    };
    return <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${badges[status]}`}>{status.toUpperCase()}</span>;
  };

  const handleSendReminder = (clientId: string) => {
    alert(`Reminder sent to client ${clientId}`);
  };

  const handleReviewTransaction = (txnId: string) => {
    setTransactions(transactions.map((t) => (t.id === txnId ? { ...t, reviewed: true } : t)));
  };

  const handleGenerateReport = () => {
    alert('Generating compliance report...');
  };

  // Circular progress for compliance score
  const CircularProgress = ({ score }: { score: number }) => {
    const radius = 45;
    const circumference = 2 * Math.PI * radius;
    const offset = circumference - (score / 100) * circumference;
    const colors = getComplianceColor(score);

    return (
      <div className="relative w-32 h-32">
        <svg className="transform -rotate-90 w-32 h-32">
          <circle
            cx="64"
            cy="64"
            r={radius}
            stroke="#383A42"
            strokeWidth="8"
            fill="none"
          />
          <circle
            cx="64"
            cy="64"
            r={radius}
            stroke={colors.bg.replace('bg-', '#')}
            strokeWidth="8"
            fill="none"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            strokeLinecap="round"
            className="transition-all duration-500"
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <div className={`text-3xl font-bold ${colors.text}`}>{score}</div>
          <div className="text-[10px] text-zinc-500">SCORE</div>
        </div>
      </div>
    );
  };

  // Donut chart for KYC distribution
  const KYCDonutChart = () => {
    const total = kycStats.total;
    const data = [
      { label: 'Verified', value: kycStats.verified, color: '#10b981' },
      { label: 'Pending', value: kycStats.pending, color: '#eab308' },
      { label: 'Expired', value: kycStats.expired, color: '#f43f5e' },
      { label: 'Rejected', value: kycStats.rejected, color: '#71717a' },
    ];

    const radius = 40;
    const centerX = 50;
    const centerY = 50;
    let currentAngle = 0;

    return (
      <div className="flex items-center gap-4">
        <svg viewBox="0 0 100 100" className="w-24 h-24">
          {data.map((segment, index) => {
            const percentage = (segment.value / total) * 100;
            const angle = (percentage / 100) * 360;
            const startAngle = currentAngle;
            const endAngle = currentAngle + angle;

            currentAngle = endAngle;

            const startRad = (startAngle - 90) * (Math.PI / 180);
            const endRad = (endAngle - 90) * (Math.PI / 180);

            const x1 = centerX + radius * Math.cos(startRad);
            const y1 = centerY + radius * Math.sin(startRad);
            const x2 = centerX + radius * Math.cos(endRad);
            const y2 = centerY + radius * Math.sin(endRad);

            const largeArc = angle > 180 ? 1 : 0;

            return (
              <path
                key={index}
                d={`M ${centerX} ${centerY} L ${x1} ${y1} A ${radius} ${radius} 0 ${largeArc} 1 ${x2} ${y2} Z`}
                fill={segment.color}
              />
            );
          })}
          <circle cx={centerX} cy={centerY} r="20" fill="#1E2026" />
        </svg>

        <div className="flex flex-col gap-1 text-[10px]">
          {data.map((segment, index) => (
            <div key={index} className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full" style={{ backgroundColor: segment.color }}></div>
              <span className="text-zinc-400">{segment.label}: {segment.value}</span>
            </div>
          ))}
        </div>
      </div>
    );
  };

  // Risk heatmap bar chart
  const RiskHeatmap = () => {
    const maxValue = Math.max(riskDistribution.low, riskDistribution.medium, riskDistribution.high);
    const risks = [
      { level: 'low' as RiskLevel, label: 'Low Risk', count: riskDistribution.low, color: 'bg-emerald-500' },
      { level: 'medium' as RiskLevel, label: 'Medium Risk', count: riskDistribution.medium, color: 'bg-yellow-500' },
      { level: 'high' as RiskLevel, label: 'High Risk', count: riskDistribution.high, color: 'bg-rose-500' },
    ];

    return (
      <div className="space-y-2">
        {risks.map((risk) => {
          const widthPct = (risk.count / maxValue) * 100;
          return (
            <div key={risk.level} className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <span className="text-zinc-400">{risk.label}</span>
                <span className="text-zinc-300 font-bold">{risk.count}</span>
              </div>
              <div
                className="h-6 bg-zinc-800 rounded overflow-hidden cursor-pointer hover:opacity-80 transition-opacity"
                onClick={() => setSelectedRiskFilter(risk.level)}
              >
                <div
                  className={`h-full ${risk.color} flex items-center justify-end px-2`}
                  style={{ width: `${widthPct}%` }}
                >
                  <span className="text-xs text-white font-bold">{risk.count} clients</span>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-zinc-300 overflow-hidden">
      {/* Header */}
      <div className="px-4 py-3 bg-[#1E2026] border-b border-zinc-700 flex items-center justify-between flex-shrink-0">
        <h2 className="text-sm font-bold text-zinc-300 flex items-center gap-2">
          <Shield size={16} className="text-blue-400" />
          Compliance Dashboard
        </h2>
        <button
          onClick={handleGenerateReport}
          className="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded transition-colors"
        >
          <Download size={12} />
          Generate Report
        </button>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-4">
        {/* Compliance Score Card */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-4 flex items-center gap-6">
          <CircularProgress score={complianceScore} />
          <div className="flex-1 space-y-2">
            <div className="text-lg font-bold text-zinc-300">Overall Compliance Score</div>
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div className="flex items-center gap-2">
                <Calendar size={12} className="text-zinc-500" />
                <span className="text-zinc-500">Last Audit:</span>
                <span className="text-zinc-300">{formatDate(lastAuditDate)}</span>
              </div>
              <div className="flex items-center gap-2">
                <Calendar size={12} className="text-zinc-500" />
                <span className="text-zinc-500">Next Review:</span>
                <span className="text-zinc-300">{formatDate(nextReviewDate)}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Stats Cards Row */}
        <div className="grid grid-cols-4 gap-3">
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <CheckCircle size={14} className="text-emerald-400" />
              <span className="text-[10px] text-zinc-500 font-bold uppercase">KYC Completion</span>
            </div>
            <div className="text-2xl font-bold text-emerald-400">{kycStats.completionRate.toFixed(1)}%</div>
            <div className="text-[10px] text-zinc-500 mt-1">{kycStats.verified} of {kycStats.total} verified</div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <AlertTriangle size={14} className="text-yellow-400" />
              <span className="text-[10px] text-zinc-500 font-bold uppercase">AML Alerts</span>
            </div>
            <div className="text-2xl font-bold text-yellow-400">{amlAlerts}</div>
            <div className="text-[10px] text-zinc-500 mt-1">Open alerts</div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp size={14} className="text-rose-400" />
              <span className="text-[10px] text-zinc-500 font-bold uppercase">Suspicious Txns</span>
            </div>
            <div className="text-2xl font-bold text-rose-400">{suspiciousTransactions}</div>
            <div className="text-[10px] text-zinc-500 mt-1">This month</div>
          </div>

          <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <FileText size={14} className="text-blue-400" />
              <span className="text-[10px] text-zinc-500 font-bold uppercase">Reports Filed</span>
            </div>
            <div className="text-2xl font-bold text-blue-400">{reportsFiled}</div>
            <div className="text-[10px] text-zinc-500 mt-1">This quarter</div>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          {/* KYC Overview Panel */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-4 space-y-4">
            <div className="text-sm font-bold text-zinc-400 uppercase">KYC Overview</div>

            <KYCDonutChart />

            <div className="space-y-2">
              <div className="text-xs font-bold text-zinc-400 uppercase">Expiring Soon (30 days)</div>
              <div className="space-y-1 max-h-48 overflow-y-auto">
                {expiringClients.length === 0 && (
                  <div className="text-xs text-zinc-500 text-center py-4">No clients expiring soon</div>
                )}
                {expiringClients.map((client) => (
                  <div key={client.id} className="flex items-center justify-between p-2 bg-[#121316] rounded text-xs">
                    <div className="flex-1">
                      <div className="text-zinc-300 font-bold">{client.name}</div>
                      <div className="text-zinc-500 text-[10px]">Expires: {client.expiryDate ? formatDate(client.expiryDate) : 'N/A'}</div>
                    </div>
                    <button
                      onClick={() => handleSendReminder(client.id)}
                      className="flex items-center gap-1 px-2 py-1 bg-blue-600 hover:bg-blue-500 text-white text-[10px] font-bold rounded transition-colors"
                    >
                      <Bell size={10} />
                      Remind
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Risk Heatmap */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-4 space-y-4">
            <div className="text-sm font-bold text-zinc-400 uppercase">Client Risk Distribution</div>
            <RiskHeatmap />
            {selectedRiskFilter !== 'all' && (
              <div className="text-xs text-zinc-400 bg-[#121316] p-2 rounded">
                Filtered by: <span className="text-zinc-300 font-bold capitalize">{selectedRiskFilter} Risk</span>
                <button
                  onClick={() => setSelectedRiskFilter('all')}
                  className="ml-2 text-blue-400 hover:text-blue-300"
                >
                  Clear
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Transaction Monitoring */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="p-3 border-b border-zinc-700 flex items-center justify-between">
            <div className="text-sm font-bold text-zinc-400 uppercase">Transaction Monitoring</div>
            <div className="text-[10px] text-zinc-500">Auto-flag threshold: ${amountThreshold.toLocaleString()}</div>
          </div>

          <div className="p-3 border-b border-zinc-700 grid grid-cols-4 gap-3">
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase">From</label>
              <input
                type="date"
                value={transactionDateFrom}
                onChange={(e) => setTransactionDateFrom(e.target.value)}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              />
            </div>
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase">To</label>
              <input
                type="date"
                value={transactionDateTo}
                onChange={(e) => setTransactionDateTo(e.target.value)}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              />
            </div>
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase">Min Amount</label>
              <input
                type="number"
                value={amountThreshold}
                onChange={(e) => setAmountThreshold(Number(e.target.value))}
                className="w-full bg-[#121316] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              />
            </div>
            <div className="space-y-1">
              <label className="text-[10px] text-zinc-500 font-bold uppercase">Filter</label>
              <label className="flex items-center gap-2 cursor-pointer bg-[#121316] border border-zinc-700 rounded px-2 py-1">
                <input
                  type="checkbox"
                  checked={showFlaggedOnly}
                  onChange={(e) => setShowFlaggedOnly(e.target.checked)}
                  className="text-blue-500"
                />
                <span className="text-xs text-zinc-300">Flagged only</span>
              </label>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="bg-[#252525] border-b border-zinc-700">
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Client</th>
                  <th className="px-3 py-2 text-right text-[10px] font-bold text-zinc-400 uppercase">Amount</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Type</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Date</th>
                  <th className="px-3 py-2 text-center text-[10px] font-bold text-zinc-400 uppercase">Flagged</th>
                  <th className="px-3 py-2 text-center text-[10px] font-bold text-zinc-400 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredTransactions.slice(0, 15).map((txn) => (
                  <tr key={txn.id} className="border-b border-zinc-800 hover:bg-[#25272E] transition-colors">
                    <td className="px-3 py-2 text-zinc-300">{txn.clientName}</td>
                    <td className="px-3 py-2 text-right text-emerald-400 font-mono font-bold">{formatCurrency(txn.amount)}</td>
                    <td className="px-3 py-2 text-zinc-400 capitalize">{txn.type}</td>
                    <td className="px-3 py-2 text-zinc-400">{formatDate(txn.date)}</td>
                    <td className="px-3 py-2 text-center">
                      {txn.flagged ? (
                        <AlertTriangle size={14} className="inline text-yellow-400" />
                      ) : (
                        <CheckCircle size={14} className="inline text-zinc-600" />
                      )}
                    </td>
                    <td className="px-3 py-2 text-center">
                      {txn.flagged && !txn.reviewed ? (
                        <button
                          onClick={() => handleReviewTransaction(txn.id)}
                          className="flex items-center gap-1 px-2 py-1 bg-blue-600 hover:bg-blue-500 text-white text-[10px] font-bold rounded transition-colors mx-auto"
                        >
                          <Eye size={10} />
                          Review
                        </button>
                      ) : (
                        <span className="text-zinc-600 text-[10px]">—</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Regulatory Reports */}
        <div className="bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="p-3 border-b border-zinc-700">
            <div className="text-sm font-bold text-zinc-400 uppercase">Regulatory Reports</div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="bg-[#252525] border-b border-zinc-700">
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Report Type</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Period</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Status</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Filed Date</th>
                  <th className="px-3 py-2 text-left text-[10px] font-bold text-zinc-400 uppercase">Due Date</th>
                </tr>
              </thead>
              <tbody>
                {regulatoryReports.map((report) => (
                  <tr key={report.id} className="border-b border-zinc-800 hover:bg-[#25272E] transition-colors">
                    <td className="px-3 py-2 text-blue-400 font-bold">{report.type}</td>
                    <td className="px-3 py-2 text-zinc-400">{report.period}</td>
                    <td className="px-3 py-2">{getReportStatusBadge(report.status)}</td>
                    <td className="px-3 py-2 text-zinc-400">
                      {report.filedDate ? formatDate(report.filedDate) : <span className="text-zinc-600">—</span>}
                    </td>
                    <td className="px-3 py-2 text-zinc-400">{formatDate(report.dueDate)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
