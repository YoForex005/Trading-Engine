'use client';

import React, { useState, useMemo } from 'react';
import {
  FileText,
  Download,
  Calendar,
  AlertCircle,
  CheckCircle,
  Clock,
  XCircle,
  TrendingUp,
  FileCheck
} from 'lucide-react';

type ReportType = 'transaction' | 'best_execution' | 'client_money' | 'suspicious_activity' | 'mifid_ii' | 'capital_adequacy';
type ReportStatus = 'generated' | 'pending' | 'submitted' | 'rejected';
type ReportFormat = 'PDF' | 'CSV' | 'XML';

interface Report {
  id: string;
  type: ReportType;
  typeName: string;
  period: string;
  generatedDate: string;
  status: ReportStatus;
  fileSize: string;
  submittedTo?: string;
  submittedDate?: string;
}

interface ReportTypeCard {
  id: ReportType;
  name: string;
  description: string;
  lastGenerated: string;
  color: string;
  icon: React.ReactElement;
}

export default function RegulatoryReporting() {
  const [selectedReportType, setSelectedReportType] = useState<ReportType | null>(null);
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const [dateRange, setDateRange] = useState({ start: '2026-01-01', end: '2026-01-31' });
  const [selectedFormat, setSelectedFormat] = useState<ReportFormat>('PDF');
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<ReportStatus | 'all'>('all');
  const [typeFilter, setTypeFilter] = useState<ReportType | 'all'>('all');

  // Report type cards
  const reportTypes: ReportTypeCard[] = useMemo(() => [
    {
      id: 'transaction',
      name: 'Transaction Report',
      description: 'Detailed log of all trading transactions',
      lastGenerated: '2026-02-08 14:30',
      color: '#3B82F6',
      icon: <FileText size={24} />
    },
    {
      id: 'best_execution',
      name: 'Best Execution Report',
      description: 'Analysis of execution quality and pricing',
      lastGenerated: '2026-02-07 09:15',
      color: '#10B981',
      icon: <TrendingUp size={24} />
    },
    {
      id: 'client_money',
      name: 'Client Money Report',
      description: 'Segregated client funds reconciliation',
      lastGenerated: '2026-02-09 16:45',
      color: '#F59E0B',
      icon: <FileCheck size={24} />
    },
    {
      id: 'suspicious_activity',
      name: 'Suspicious Activity Report',
      description: 'SAR - Flagged transactions for AML',
      lastGenerated: '2026-02-06 11:20',
      color: '#EF4444',
      icon: <AlertCircle size={24} />
    },
    {
      id: 'mifid_ii',
      name: 'MiFID II Report',
      description: 'EU Markets in Financial Instruments Directive',
      lastGenerated: '2026-02-05 13:00',
      color: '#8B5CF6',
      icon: <FileText size={24} />
    },
    {
      id: 'capital_adequacy',
      name: 'Capital Adequacy Report',
      description: 'Regulatory capital requirements compliance',
      lastGenerated: '2026-02-04 10:30',
      color: '#06B6D4',
      icon: <FileCheck size={24} />
    }
  ], []);

  // Mock generated reports
  const allReports: Report[] = useMemo(() => [
    { id: 'REP-2026-001', type: 'transaction', typeName: 'Transaction Report', period: 'Jan 2026', generatedDate: '2026-02-08 14:30', status: 'submitted', fileSize: '12.4 MB', submittedTo: 'FCA', submittedDate: '2026-02-08 15:00' },
    { id: 'REP-2026-002', type: 'best_execution', typeName: 'Best Execution Report', period: 'Jan 2026', generatedDate: '2026-02-07 09:15', status: 'generated', fileSize: '8.7 MB' },
    { id: 'REP-2026-003', type: 'client_money', typeName: 'Client Money Report', period: 'Jan 2026', generatedDate: '2026-02-09 16:45', status: 'submitted', fileSize: '5.2 MB', submittedTo: 'FCA', submittedDate: '2026-02-09 17:00' },
    { id: 'REP-2026-004', type: 'suspicious_activity', typeName: 'Suspicious Activity Report', period: 'Q4 2025', generatedDate: '2026-02-06 11:20', status: 'submitted', fileSize: '2.1 MB', submittedTo: 'NCA', submittedDate: '2026-02-06 12:00' },
    { id: 'REP-2026-005', type: 'mifid_ii', typeName: 'MiFID II Report', period: 'Jan 2026', generatedDate: '2026-02-05 13:00', status: 'pending', fileSize: '15.8 MB' },
    { id: 'REP-2026-006', type: 'capital_adequacy', typeName: 'Capital Adequacy Report', period: 'Q1 2026', generatedDate: '2026-02-04 10:30', status: 'generated', fileSize: '4.3 MB' },
    { id: 'REP-2025-020', type: 'transaction', typeName: 'Transaction Report', period: 'Dec 2025', generatedDate: '2026-01-08 14:30', status: 'submitted', fileSize: '11.9 MB', submittedTo: 'FCA', submittedDate: '2026-01-08 15:30' },
    { id: 'REP-2025-019', type: 'best_execution', typeName: 'Best Execution Report', period: 'Dec 2025', generatedDate: '2026-01-07 09:15', status: 'submitted', fileSize: '9.1 MB', submittedTo: 'FCA', submittedDate: '2026-01-07 10:00' },
    { id: 'REP-2025-018', type: 'client_money', typeName: 'Client Money Report', period: 'Dec 2025', generatedDate: '2026-01-09 16:45', status: 'submitted', fileSize: '5.5 MB', submittedTo: 'FCA', submittedDate: '2026-01-09 17:15' },
    { id: 'REP-2025-017', type: 'mifid_ii', typeName: 'MiFID II Report', period: 'Dec 2025', generatedDate: '2026-01-05 13:00', status: 'submitted', fileSize: '16.2 MB', submittedTo: 'ESMA', submittedDate: '2026-01-05 14:00' },
    { id: 'REP-2025-016', type: 'transaction', typeName: 'Transaction Report', period: 'Nov 2025', generatedDate: '2025-12-08 14:30', status: 'submitted', fileSize: '12.1 MB', submittedTo: 'FCA', submittedDate: '2025-12-08 15:00' },
    { id: 'REP-2025-015', type: 'best_execution', typeName: 'Best Execution Report', period: 'Nov 2025', generatedDate: '2025-12-07 09:15', status: 'submitted', fileSize: '8.9 MB', submittedTo: 'FCA', submittedDate: '2025-12-07 10:00' },
    { id: 'REP-2025-014', type: 'client_money', typeName: 'Client Money Report', period: 'Nov 2025', generatedDate: '2025-12-09 16:45', status: 'submitted', fileSize: '5.3 MB', submittedTo: 'FCA', submittedDate: '2025-12-09 17:00' },
    { id: 'REP-2025-013', type: 'capital_adequacy', typeName: 'Capital Adequacy Report', period: 'Q4 2025', generatedDate: '2025-12-15 10:30', status: 'rejected', fileSize: '4.5 MB', submittedTo: 'PRA', submittedDate: '2025-12-15 11:00' },
    { id: 'REP-2025-012', type: 'transaction', typeName: 'Transaction Report', period: 'Oct 2025', generatedDate: '2025-11-08 14:30', status: 'submitted', fileSize: '11.7 MB', submittedTo: 'FCA', submittedDate: '2025-11-08 15:00' },
    { id: 'REP-2025-011', type: 'best_execution', typeName: 'Best Execution Report', period: 'Oct 2025', generatedDate: '2025-11-07 09:15', status: 'submitted', fileSize: '8.5 MB', submittedTo: 'FCA', submittedDate: '2025-11-07 10:00' },
    { id: 'REP-2025-010', type: 'suspicious_activity', typeName: 'Suspicious Activity Report', period: 'Q3 2025', generatedDate: '2025-11-10 11:20', status: 'submitted', fileSize: '1.8 MB', submittedTo: 'NCA', submittedDate: '2025-11-10 12:00' },
    { id: 'REP-2025-009', type: 'mifid_ii', typeName: 'MiFID II Report', period: 'Oct 2025', generatedDate: '2025-11-05 13:00', status: 'submitted', fileSize: '15.5 MB', submittedTo: 'ESMA', submittedDate: '2025-11-05 14:00' },
    { id: 'REP-2025-008', type: 'transaction', typeName: 'Transaction Report', period: 'Sep 2025', generatedDate: '2025-10-08 14:30', status: 'submitted', fileSize: '12.3 MB', submittedTo: 'FCA', submittedDate: '2025-10-08 15:00' },
    { id: 'REP-2025-007', type: 'client_money', typeName: 'Client Money Report', period: 'Sep 2025', generatedDate: '2025-10-09 16:45', status: 'submitted', fileSize: '5.1 MB', submittedTo: 'FCA', submittedDate: '2025-10-09 17:00' }
  ], []);

  // Calculate stats
  const stats = useMemo(() => {
    const now = new Date('2026-02-11');
    const thisMonth = allReports.filter(r => r.period.includes('Feb 2026') || r.period.includes('Q1 2026'));
    const overdue = allReports.filter(r => r.status === 'pending' && new Date(r.generatedDate) < new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000));
    const submitted = allReports.filter(r => r.status === 'submitted').length;
    const submissionRate = allReports.length > 0 ? Math.round((submitted / allReports.length) * 100) : 0;

    return {
      dueThisMonth: thisMonth.length,
      overdue: overdue.length,
      submissionRate,
      nextDeadline: '2026-02-15 17:00'
    };
  }, [allReports]);

  // Timeline data for SVG chart (last 6 months)
  const timelineData = useMemo(() => {
    const months = ['Sep 2025', 'Oct 2025', 'Nov 2025', 'Dec 2025', 'Jan 2026', 'Feb 2026'];
    return months.map(month => ({
      month,
      submitted: allReports.filter(r => r.period === month && r.status === 'submitted').length,
      pending: allReports.filter(r => r.period === month && r.status === 'pending').length,
      rejected: allReports.filter(r => r.period === month && r.status === 'rejected').length
    }));
  }, [allReports]);

  // Filter reports
  const filteredReports = useMemo(() => {
    return allReports.filter(report => {
      const matchesSearch = report.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           report.typeName.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           report.period.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesStatus = statusFilter === 'all' || report.status === statusFilter;
      const matchesType = typeFilter === 'all' || report.type === typeFilter;
      return matchesSearch && matchesStatus && matchesType;
    });
  }, [allReports, searchTerm, statusFilter, typeFilter]);

  const getStatusBadge = (status: ReportStatus) => {
    const styles = {
      generated: { bg: 'bg-blue-500/20', text: 'text-blue-400', icon: <FileText size={14} /> },
      pending: { bg: 'bg-yellow-500/20', text: 'text-yellow-400', icon: <Clock size={14} /> },
      submitted: { bg: 'bg-green-500/20', text: 'text-green-400', icon: <CheckCircle size={14} /> },
      rejected: { bg: 'bg-red-500/20', text: 'text-red-400', icon: <XCircle size={14} /> }
    };
    const style = styles[status];
    return (
      <span className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium ${style.bg} ${style.text}`}>
        {style.icon}
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    );
  };

  const handleGenerate = () => {
    setShowGenerateModal(false);
    // In real implementation, would trigger report generation
    console.log('Generating report:', selectedReportType, dateRange, selectedFormat);
  };

  const maxSubmissions = Math.max(...timelineData.map(d => d.submitted + d.pending + d.rejected), 1);

  return (
    <div className="h-full flex flex-col bg-[#18181b] overflow-hidden">
      {/* Header */}
      <div className="flex-shrink-0 px-6 py-4 border-b border-zinc-700">
        <h2 className="text-xl font-bold text-zinc-100">Regulatory Reporting</h2>
        <p className="text-sm text-zinc-400 mt-1">Generate and manage compliance reports for regulatory submissions</p>
      </div>

      {/* Stats Cards */}
      <div className="flex-shrink-0 px-6 py-4 grid grid-cols-4 gap-4">
        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-zinc-400 uppercase tracking-wide">Reports Due This Month</p>
              <p className="text-2xl font-bold text-zinc-100 mt-1">{stats.dueThisMonth}</p>
            </div>
            <Calendar className="text-blue-400" size={32} />
          </div>
        </div>

        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-zinc-400 uppercase tracking-wide">Overdue Reports</p>
              <p className="text-2xl font-bold text-red-400 mt-1">{stats.overdue}</p>
            </div>
            <AlertCircle className="text-red-400" size={32} />
          </div>
        </div>

        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-zinc-400 uppercase tracking-wide">Submission Rate</p>
              <p className="text-2xl font-bold text-green-400 mt-1">{stats.submissionRate}%</p>
            </div>
            <CheckCircle className="text-green-400" size={32} />
          </div>
        </div>

        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-zinc-400 uppercase tracking-wide">Next Deadline</p>
              <p className="text-sm font-bold text-zinc-100 mt-1">{stats.nextDeadline}</p>
            </div>
            <Clock className="text-orange-400" size={32} />
          </div>
        </div>
      </div>

      {/* Report Type Cards */}
      <div className="flex-shrink-0 px-6 py-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3 uppercase tracking-wide">Report Types</h3>
        <div className="grid grid-cols-3 gap-4">
          {reportTypes.map(reportType => (
            <div key={reportType.id} className="bg-[#27272a] rounded-lg p-4 border border-zinc-700 hover:border-zinc-600 transition-colors">
              <div className="flex items-start justify-between mb-3">
                <div className="p-2 rounded-lg" style={{ backgroundColor: `${reportType.color}20` }}>
                  <div style={{ color: reportType.color }}>
                    {reportType.icon}
                  </div>
                </div>
                <button
                  onClick={() => {
                    setSelectedReportType(reportType.id);
                    setShowGenerateModal(true);
                  }}
                  className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium rounded transition-colors"
                >
                  Generate
                </button>
              </div>
              <h4 className="text-sm font-semibold text-zinc-100 mb-1">{reportType.name}</h4>
              <p className="text-xs text-zinc-400 mb-2">{reportType.description}</p>
              <div className="flex items-center gap-1 text-xs text-zinc-500">
                <Clock size={12} />
                <span>Last: {reportType.lastGenerated}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Timeline Chart */}
      <div className="flex-shrink-0 px-6 py-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3 uppercase tracking-wide">Submission Timeline (Last 6 Months)</h3>
        <div className="bg-[#27272a] rounded-lg p-4 border border-zinc-700">
          <svg width="100%" height="200" viewBox="0 0 800 200">
            {/* Grid lines */}
            {[0, 1, 2, 3, 4, 5].map(i => (
              <line
                key={i}
                x1="60"
                y1={20 + i * 30}
                x2="780"
                y2={20 + i * 30}
                stroke="#3f3f46"
                strokeWidth="1"
                strokeDasharray="4 4"
              />
            ))}

            {/* Y-axis labels */}
            {[0, 1, 2, 3, 4, 5].map(i => (
              <text key={i} x="50" y={25 + i * 30} textAnchor="end" fill="#71717a" fontSize="11">
                {5 - i}
              </text>
            ))}

            {/* Bars */}
            {timelineData.map((data, idx) => {
              const x = 80 + idx * 120;
              const barWidth = 30;
              const submittedHeight = (data.submitted / maxSubmissions) * 150;
              const pendingHeight = (data.pending / maxSubmissions) * 150;
              const rejectedHeight = (data.rejected / maxSubmissions) * 150;

              return (
                <g key={idx}>
                  {/* Submitted bar */}
                  <rect
                    x={x}
                    y={170 - submittedHeight}
                    width={barWidth}
                    height={submittedHeight}
                    fill="#10b981"
                    rx="2"
                  />
                  {/* Pending bar */}
                  <rect
                    x={x + barWidth + 5}
                    y={170 - pendingHeight}
                    width={barWidth}
                    height={pendingHeight}
                    fill="#f59e0b"
                    rx="2"
                  />
                  {/* Rejected bar */}
                  <rect
                    x={x + (barWidth + 5) * 2}
                    y={170 - rejectedHeight}
                    width={barWidth}
                    height={rejectedHeight}
                    fill="#ef4444"
                    rx="2"
                  />
                  {/* Month label */}
                  <text x={x + barWidth + 20} y="190" textAnchor="middle" fill="#71717a" fontSize="10">
                    {data.month.split(' ')[0]}
                  </text>
                </g>
              );
            })}

            {/* Legend */}
            <g transform="translate(620, 10)">
              <rect x="0" y="0" width="12" height="12" fill="#10b981" rx="2" />
              <text x="18" y="10" fill="#a1a1aa" fontSize="11">Submitted</text>

              <rect x="90" y="0" width="12" height="12" fill="#f59e0b" rx="2" />
              <text x="108" y="10" fill="#a1a1aa" fontSize="11">Pending</text>
            </g>
          </svg>
        </div>
      </div>

      {/* Generated Reports Table */}
      <div className="flex-1 flex flex-col px-6 pb-4 overflow-hidden">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-semibold text-zinc-300 uppercase tracking-wide">Generated Reports</h3>
          <div className="flex gap-2">
            <input
              type="text"
              placeholder="Search reports..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="px-3 py-1.5 bg-[#27272a] border border-zinc-700 rounded text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-zinc-600"
            />
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value as ReportType | 'all')}
              className="px-3 py-1.5 bg-[#27272a] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
            >
              <option value="all">All Types</option>
              <option value="transaction">Transaction</option>
              <option value="best_execution">Best Execution</option>
              <option value="client_money">Client Money</option>
              <option value="suspicious_activity">Suspicious Activity</option>
              <option value="mifid_ii">MiFID II</option>
              <option value="capital_adequacy">Capital Adequacy</option>
            </select>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as ReportStatus | 'all')}
              className="px-3 py-1.5 bg-[#27272a] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
            >
              <option value="all">All Statuses</option>
              <option value="generated">Generated</option>
              <option value="pending">Pending</option>
              <option value="submitted">Submitted</option>
              <option value="rejected">Rejected</option>
            </select>
          </div>
        </div>

        <div className="flex-1 bg-[#27272a] rounded-lg border border-zinc-700 overflow-auto">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-[#27272a] border-b border-zinc-700">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Report ID</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Type</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Period</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Generated</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Status</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">File Size</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Submitted To</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-zinc-400 uppercase tracking-wide">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredReports.map((report, idx) => (
                <tr
                  key={report.id}
                  className={`border-b border-zinc-700/50 hover:bg-zinc-800/50 transition-colors ${
                    idx % 2 === 0 ? 'bg-[#27272a]' : 'bg-[#27272a]/50'
                  }`}
                >
                  <td className="px-4 py-3 font-mono text-zinc-300">{report.id}</td>
                  <td className="px-4 py-3 text-zinc-300">{report.typeName}</td>
                  <td className="px-4 py-3 text-zinc-400">{report.period}</td>
                  <td className="px-4 py-3 text-zinc-400">{report.generatedDate}</td>
                  <td className="px-4 py-3">{getStatusBadge(report.status)}</td>
                  <td className="px-4 py-3 text-zinc-400">{report.fileSize}</td>
                  <td className="px-4 py-3 text-zinc-400">
                    {report.submittedTo ? (
                      <div>
                        <div className="font-medium text-zinc-300">{report.submittedTo}</div>
                        <div className="text-xs text-zinc-500">{report.submittedDate}</div>
                      </div>
                    ) : (
                      <span className="text-zinc-600">—</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button className="inline-flex items-center gap-1 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium rounded transition-colors">
                      <Download size={14} />
                      Download
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {filteredReports.length === 0 && (
            <div className="flex flex-col items-center justify-center py-16 text-zinc-500">
              <FileText size={48} className="mb-3 opacity-20" />
              <p className="text-sm">No reports found matching your filters</p>
            </div>
          )}
        </div>
      </div>

      {/* Generate Report Modal */}
      {showGenerateModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-[#27272a] rounded-lg border border-zinc-700 w-full max-w-md mx-4">
            <div className="px-6 py-4 border-b border-zinc-700">
              <h3 className="text-lg font-semibold text-zinc-100">Generate Report</h3>
              <p className="text-sm text-zinc-400 mt-1">
                {reportTypes.find(r => r.id === selectedReportType)?.name}
              </p>
            </div>

            <div className="px-6 py-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">Date Range</label>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs text-zinc-500 mb-1">Start Date</label>
                    <input
                      type="date"
                      value={dateRange.start}
                      onChange={(e) => setDateRange(prev => ({ ...prev, start: e.target.value }))}
                      className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-zinc-500 mb-1">End Date</label>
                    <input
                      type="date"
                      value={dateRange.end}
                      onChange={(e) => setDateRange(prev => ({ ...prev, end: e.target.value }))}
                      className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                    />
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">Export Format</label>
                <div className="flex gap-2">
                  {(['PDF', 'CSV', 'XML'] as ReportFormat[]).map(format => (
                    <button
                      key={format}
                      onClick={() => setSelectedFormat(format)}
                      className={`flex-1 px-4 py-2 rounded text-sm font-medium transition-colors ${
                        selectedFormat === format
                          ? 'bg-blue-600 text-white'
                          : 'bg-[#18181b] text-zinc-400 hover:bg-zinc-800 border border-zinc-700'
                      }`}
                    >
                      {format}
                    </button>
                  ))}
                </div>
              </div>
            </div>

            <div className="px-6 py-4 border-t border-zinc-700 flex justify-end gap-3">
              <button
                onClick={() => setShowGenerateModal(false)}
                className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-zinc-100 text-sm font-medium rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleGenerate}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded transition-colors"
              >
                Generate Report
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
