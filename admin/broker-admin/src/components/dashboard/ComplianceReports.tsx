'use client';

import { useState, useEffect, useMemo } from 'react';
import {
  FileText, Shield, AlertTriangle, CheckCircle2, XCircle,
  Clock, Download, RefreshCw, Filter, Search, Calendar,
  AlertCircle, TrendingUp, BarChart3, X
} from 'lucide-react';
import { API_CONFIG } from '../../config/api';

// ============================================
// TypeScript Interfaces (matching backend)
// ============================================

interface ComplianceReport {
  id: number;
  type: string; // "MiFID_II", "EMIR", "ASIC", "FCA", "CySEC"
  regulator: string;
  period: string;
  format: string; // "PDF", "CSV", "XML"
  status: string; // "pending", "generating", "completed", "failed"
  filePath: string;
  fileSize: number;
  generatedAt: string;
  generatedBy: string;
  dataSummary: ReportSummary;
  submittedTo?: string;
  submittedAt?: string;
}

interface ReportSummary {
  totalTrades: number;
  totalVolume: number;
  totalClients: number;
  totalTransactions: number;
  complianceScore: number;
  issuesFound: number;
}

interface ReportTemplate {
  id: number;
  regulator: string;
  reportType: string;
  name: string;
  description: string;
  requiredData: string[];
  frequency: string;
  deadline: string;
  format: string[];
}

interface ComplianceCheck {
  id: number;
  name: string;
  category: string;
  description: string;
  status: string; // "pass", "fail", "warning"
  score: number; // 0-100
  details: string;
  lastChecked: string;
  priority: string; // "critical", "high", "medium", "low"
}

interface ComplianceBreach {
  id: number;
  date: string;
  type: string;
  severity: string; // "critical", "high", "medium", "low"
  description: string;
  status: string; // "open", "investigating", "resolved", "closed"
  resolution?: string;
  resolvedAt?: string;
  resolvedBy?: string;
  clientId?: number;
  tradeId?: number;
}

interface ComplianceStats {
  overallScore: number;
  checksPassed: number;
  checksFailed: number;
  checksWarning: number;
  totalChecks: number;
  reportsGenerated: number;
  reportsThisQuarter: number;
  openBreaches: number;
  resolvedBreaches: number;
  upcomingDeadlines: RegulatoryDeadline[];
  lastAuditDate: string;
}

interface RegulatoryDeadline {
  regulator: string;
  reportType: string;
  dueDate: string;
  daysUntil: number;
  status: string; // "upcoming", "due_soon", "overdue", "submitted"
}

// ============================================
// Helper Functions
// ============================================

const getStatusBadgeColor = (status: string): string => {
  const colors: Record<string, string> = {
    'completed': '#10B981',
    'pending': '#F59E0B',
    'generating': '#3B82F6',
    'failed': '#EF4444',
    'pass': '#10B981',
    'fail': '#EF4444',
    'warning': '#F59E0B',
    'open': '#EF4444',
    'investigating': '#F59E0B',
    'resolved': '#10B981',
    'closed': '#6B7280'
  };
  return colors[status] || '#6B7280';
};

const getSeverityBadgeColor = (severity: string): string => {
  const colors: Record<string, string> = {
    'critical': '#EF4444',
    'high': '#F59E0B',
    'medium': '#F59E0B',
    'low': '#10B981'
  };
  return colors[severity] || '#6B7280';
};

const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
};

const formatDateTime = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

// ============================================
// Main Component
// ============================================

export default function ComplianceReports() {
  const [reports, setReports] = useState<ComplianceReport[]>([]);
  const [templates, setTemplates] = useState<ReportTemplate[]>([]);
  const [checks, setChecks] = useState<ComplianceCheck[]>([]);
  const [breaches, setBreaches] = useState<ComplianceBreach[]>([]);
  const [stats, setStats] = useState<ComplianceStats | null>(null);

  const [selectedReport, setSelectedReport] = useState<ComplianceReport | null>(null);
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const [showDetailView, setShowDetailView] = useState(false);

  const [filterRegulator, setFilterRegulator] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [filterType, setFilterType] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  const [breachFilterStatus, setBreachFilterStatus] = useState('');
  const [breachFilterSeverity, setBreachFilterSeverity] = useState('');

  // Generate Report Modal State
  const [generateForm, setGenerateForm] = useState({
    type: 'MiFID_II',
    regulator: 'MiFID II',
    period: 'Q1_2026',
    format: 'PDF'
  });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    const token = typeof window !== 'undefined' ?
      localStorage.getItem('rtx_token') || localStorage.getItem('admin_token') || localStorage.getItem('jwt_token') : null;

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    try {
      const [reportsRes, templatesRes, checksRes, breachesRes, statsRes] = await Promise.all([
        fetch(`${API_CONFIG.COMPLIANCE_REPORTS_LIST}`, { headers }),
        fetch(`${API_CONFIG.COMPLIANCE_REPORT_TEMPLATES}`, { headers }),
        fetch(`${API_CONFIG.COMPLIANCE_CHECKS}`, { headers }),
        fetch(`${API_CONFIG.COMPLIANCE_BREACHES}`, { headers }),
        fetch(`${API_CONFIG.COMPLIANCE_STATS}`, { headers })
      ]);

      const reportsData = await reportsRes.json();
      const templatesData = await templatesRes.json();
      const checksData = await checksRes.json();
      const breachesData = await breachesRes.json();
      const statsData = await statsRes.json();

      setReports(reportsData.reports || []);
      setTemplates(templatesData.templates || []);
      setChecks(checksData.checks || []);
      setBreaches(breachesData.breaches || []);
      setStats(statsData);
    } catch (error) {
      console.error('Failed to fetch compliance data:', error);
    }
  };

  const handleGenerateReport = async () => {
    const token = typeof window !== 'undefined' ?
      localStorage.getItem('rtx_token') || localStorage.getItem('admin_token') || localStorage.getItem('jwt_token') : null;

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    try {
      const response = await fetch(`${API_CONFIG.COMPLIANCE_REPORT_GENERATE}`, {
        method: 'POST',
        headers,
        body: JSON.stringify(generateForm)
      });

      if (response.ok) {
        setShowGenerateModal(false);
        fetchData(); // Refresh data
      }
    } catch (error) {
      console.error('Failed to generate report:', error);
    }
  };

  const handleRunChecks = async () => {
    const token = typeof window !== 'undefined' ?
      localStorage.getItem('rtx_token') || localStorage.getItem('admin_token') || localStorage.getItem('jwt_token') : null;

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    try {
      const response = await fetch(`${API_CONFIG.COMPLIANCE_CHECKS_RUN}`, {
        method: 'POST',
        headers
      });

      if (response.ok) {
        fetchData(); // Refresh checks
      }
    } catch (error) {
      console.error('Failed to run checks:', error);
    }
  };

  const viewReportDetail = async (reportId: number) => {
    const token = typeof window !== 'undefined' ?
      localStorage.getItem('rtx_token') || localStorage.getItem('admin_token') || localStorage.getItem('jwt_token') : null;

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    try {
      const response = await fetch(`${API_CONFIG.COMPLIANCE_REPORT_DETAIL(reportId)}`, { headers });
      const report = await response.json();
      setSelectedReport(report);
      setShowDetailView(true);
    } catch (error) {
      console.error('Failed to fetch report detail:', error);
    }
  };

  // Filtered Reports
  const filteredReports = useMemo(() => {
    return reports.filter(report => {
      if (filterRegulator && report.regulator !== filterRegulator) return false;
      if (filterStatus && report.status !== filterStatus) return false;
      if (filterType && report.type !== filterType) return false;
      if (searchQuery && !report.type.toLowerCase().includes(searchQuery.toLowerCase()) &&
          !report.regulator.toLowerCase().includes(searchQuery.toLowerCase())) return false;
      return true;
    });
  }, [reports, filterRegulator, filterStatus, filterType, searchQuery]);

  // Filtered Breaches
  const filteredBreaches = useMemo(() => {
    return breaches.filter(breach => {
      if (breachFilterStatus && breach.status !== breachFilterStatus) return false;
      if (breachFilterSeverity && breach.severity !== breachFilterSeverity) return false;
      return true;
    });
  }, [breaches, breachFilterStatus, breachFilterSeverity]);

  // Calculate report status distribution for pie chart
  const reportStatusDistribution = useMemo(() => {
    const distribution: Record<string, number> = {
      'completed': 0,
      'generating': 0,
      'pending': 0,
      'failed': 0
    };

    reports.forEach(report => {
      if (distribution[report.status] !== undefined) {
        distribution[report.status]++;
      }
    });

    return [
      { status: 'Completed', count: distribution.completed, color: '#10B981' },
      { status: 'Generating', count: distribution.generating, color: '#3B82F6' },
      { status: 'Pending', count: distribution.pending, color: '#F59E0B' },
      { status: 'Failed', count: distribution.failed, color: '#EF4444' }
    ];
  }, [reports]);

  // SVG Pie Chart
  const renderPieChart = () => {
    const total = reportStatusDistribution.reduce((sum, item) => sum + item.count, 0);
    if (total === 0) return null;

    let currentAngle = 0;
    const radius = 80;
    const centerX = 100;
    const centerY = 100;

    const slices = reportStatusDistribution.map(item => {
      const percentage = item.count / total;
      const angle = percentage * 360;
      const startAngle = currentAngle;
      const endAngle = currentAngle + angle;
      currentAngle = endAngle;

      const startRad = (startAngle - 90) * (Math.PI / 180);
      const endRad = (endAngle - 90) * (Math.PI / 180);

      const x1 = centerX + radius * Math.cos(startRad);
      const y1 = centerY + radius * Math.sin(startRad);
      const x2 = centerX + radius * Math.cos(endRad);
      const y2 = centerY + radius * Math.sin(endRad);

      const largeArcFlag = angle > 180 ? 1 : 0;

      const pathData = [
        `M ${centerX} ${centerY}`,
        `L ${x1} ${y1}`,
        `A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2}`,
        'Z'
      ].join(' ');

      return { ...item, pathData, percentage: (percentage * 100).toFixed(1) };
    });

    return (
      <div className="flex items-center gap-6">
        <svg width="200" height="200" viewBox="0 0 200 200">
          {slices.map((slice, idx) => (
            <path
              key={idx}
              d={slice.pathData}
              fill={slice.color}
              stroke="#1E2026"
              strokeWidth="2"
            />
          ))}
        </svg>
        <div className="flex flex-col gap-2">
          {slices.map((slice, idx) => (
            <div key={idx} className="flex items-center gap-2">
              <div
                className="w-4 h-4 rounded"
                style={{ backgroundColor: slice.color }}
              />
              <span className="text-sm text-[#A1A1AA]">
                {slice.status}: {slice.count} ({slice.percentage}%)
              </span>
            </div>
          ))}
        </div>
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-white overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-[#383A42]">
        <div className="flex items-center gap-3">
          <Shield className="text-[#F5C542]" size={24} />
          <div>
            <h1 className="text-lg font-semibold">Compliance Reports & Regulatory Reporting</h1>
            <p className="text-sm text-[#A1A1AA]">MiFID II, EMIR, ASIC, FCA, CySEC</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleRunChecks}
            className="flex items-center gap-2 px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] rounded text-sm transition-colors"
          >
            <RefreshCw size={16} />
            Run All Checks
          </button>
          <button
            onClick={() => setShowGenerateModal(true)}
            className="flex items-center gap-2 px-3 py-1.5 bg-[#F5C542] hover:bg-[#E5B532] text-black rounded text-sm font-medium transition-colors"
          >
            <FileText size={16} />
            Generate Report
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-4 gap-4 p-4 border-b border-[#383A42]">
          <div className="bg-[#1E2026] p-4 rounded border border-[#383A42]">
            <div className="text-sm text-[#A1A1AA]">Total Reports</div>
            <div className="text-2xl font-bold mt-1">{stats.reportsGenerated}</div>
            <div className="text-xs text-[#A1A1AA] mt-1">This Quarter: {stats.reportsThisQuarter}</div>
          </div>
          <div className="bg-[#1E2026] p-4 rounded border border-[#383A42]">
            <div className="text-sm text-[#A1A1AA]">Pending Submissions</div>
            <div className="text-2xl font-bold mt-1 text-[#F59E0B]">
              {reports.filter(r => r.status === 'pending' || r.status === 'generating').length}
            </div>
            <div className="text-xs text-[#A1A1AA] mt-1">Awaiting submission</div>
          </div>
          <div className="bg-[#1E2026] p-4 rounded border border-[#383A42]">
            <div className="text-sm text-[#A1A1AA]">Compliance Score</div>
            <div className="text-2xl font-bold mt-1 text-[#10B981]">{stats.overallScore.toFixed(1)}%</div>
            <div className="text-xs text-[#A1A1AA] mt-1">
              {stats.checksPassed} passed, {stats.checksFailed} failed
            </div>
          </div>
          <div className="bg-[#1E2026] p-4 rounded border border-[#383A42]">
            <div className="text-sm text-[#A1A1AA]">Active Breaches</div>
            <div className="text-2xl font-bold mt-1 text-[#EF4444]">{stats.openBreaches}</div>
            <div className="text-xs text-[#A1A1AA] mt-1">
              {stats.resolvedBreaches} resolved total
            </div>
          </div>
        </div>
      )}

      {/* Main Content - Two Columns */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left Column - Reports List & Checks */}
        <div className="flex-1 flex flex-col border-r border-[#383A42] overflow-hidden">
          {/* Report List Table */}
          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="flex items-center justify-between p-4 border-b border-[#383A42]">
              <h2 className="font-semibold">Reports ({filteredReports.length})</h2>
              <div className="flex items-center gap-2">
                <div className="relative">
                  <Search className="absolute left-2 top-1/2 -translate-y-1/2 text-[#666]" size={16} />
                  <input
                    type="text"
                    placeholder="Search reports..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-8 pr-3 py-1.5 bg-[#1E2026] border border-[#383A42] rounded text-sm w-48 focus:outline-none focus:border-[#F5C542]"
                  />
                </div>
                <select
                  value={filterRegulator}
                  onChange={(e) => setFilterRegulator(e.target.value)}
                  className="px-3 py-1.5 bg-[#1E2026] border border-[#383A42] rounded text-sm focus:outline-none focus:border-[#F5C542]"
                >
                  <option value="">All Regulators</option>
                  <option value="MiFID II">MiFID II</option>
                  <option value="EMIR">EMIR</option>
                  <option value="ASIC">ASIC</option>
                  <option value="FCA">FCA</option>
                  <option value="CySEC">CySEC</option>
                </select>
                <select
                  value={filterStatus}
                  onChange={(e) => setFilterStatus(e.target.value)}
                  className="px-3 py-1.5 bg-[#1E2026] border border-[#383A42] rounded text-sm focus:outline-none focus:border-[#F5C542]"
                >
                  <option value="">All Status</option>
                  <option value="completed">Completed</option>
                  <option value="generating">Generating</option>
                  <option value="pending">Pending</option>
                  <option value="failed">Failed</option>
                </select>
              </div>
            </div>

            <div className="flex-1 overflow-auto">
              <table className="w-full">
                <thead className="bg-[#1E2026] sticky top-0 z-10">
                  <tr className="text-left text-sm text-[#A1A1AA]">
                    <th className="p-3 font-medium">ID</th>
                    <th className="p-3 font-medium">Type</th>
                    <th className="p-3 font-medium">Regulator</th>
                    <th className="p-3 font-medium">Period</th>
                    <th className="p-3 font-medium">Status</th>
                    <th className="p-3 font-medium">Generated</th>
                    <th className="p-3 font-medium">Format</th>
                    <th className="p-3 font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredReports.map((report) => (
                    <tr
                      key={report.id}
                      className="border-b border-[#383A42] hover:bg-[#1E2026] cursor-pointer transition-colors"
                      onClick={() => viewReportDetail(report.id)}
                    >
                      <td className="p-3 text-sm">#{report.id}</td>
                      <td className="p-3 text-sm">{report.type.replace(/_/g, ' ')}</td>
                      <td className="p-3 text-sm">{report.regulator}</td>
                      <td className="p-3 text-sm">{report.period.replace(/_/g, ' ')}</td>
                      <td className="p-3">
                        <span
                          className="px-2 py-1 rounded text-xs font-medium"
                          style={{
                            backgroundColor: `${getStatusBadgeColor(report.status)}20`,
                            color: getStatusBadgeColor(report.status)
                          }}
                        >
                          {report.status}
                        </span>
                      </td>
                      <td className="p-3 text-sm text-[#A1A1AA]">{formatDate(report.generatedAt)}</td>
                      <td className="p-3 text-sm">{report.format}</td>
                      <td className="p-3">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            // Download report
                          }}
                          className="text-[#F5C542] hover:text-[#E5B532]"
                        >
                          <Download size={16} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Compliance Checks Panel */}
          <div className="flex-none border-t border-[#383A42] h-64 flex flex-col">
            <div className="flex items-center justify-between p-3 border-b border-[#383A42]">
              <h3 className="font-semibold text-sm">Compliance Checks ({checks.length})</h3>
            </div>
            <div className="flex-1 overflow-auto">
              <div className="grid grid-cols-2 gap-2 p-3">
                {checks.map((check) => (
                  <div
                    key={check.id}
                    className="bg-[#1E2026] p-3 rounded border border-[#383A42]"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1">
                        <div className="text-sm font-medium mb-1">{check.name}</div>
                        <div className="text-xs text-[#A1A1AA]">{check.category}</div>
                      </div>
                      {check.status === 'pass' && <CheckCircle2 className="text-[#10B981]" size={18} />}
                      {check.status === 'fail' && <XCircle className="text-[#EF4444]" size={18} />}
                      {check.status === 'warning' && <AlertTriangle className="text-[#F59E0B]" size={18} />}
                    </div>
                    <div className="flex items-center justify-between text-xs">
                      <span className="text-[#A1A1AA]">Score: {check.score.toFixed(1)}%</span>
                      <span className="text-[#A1A1AA]">{formatDateTime(check.lastChecked)}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Right Column - Breach Log & Chart */}
        <div className="w-96 flex flex-col overflow-hidden">
          {/* Report Status Pie Chart */}
          <div className="flex-none p-4 border-b border-[#383A42]">
            <h3 className="font-semibold mb-4">Report Status Distribution</h3>
            {renderPieChart()}
          </div>

          {/* Breach Log Table */}
          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="flex items-center justify-between p-3 border-b border-[#383A42]">
              <h3 className="font-semibold text-sm">Breach Log ({filteredBreaches.length})</h3>
              <div className="flex gap-2">
                <select
                  value={breachFilterSeverity}
                  onChange={(e) => setBreachFilterSeverity(e.target.value)}
                  className="px-2 py-1 bg-[#1E2026] border border-[#383A42] rounded text-xs focus:outline-none focus:border-[#F5C542]"
                >
                  <option value="">All Severity</option>
                  <option value="critical">Critical</option>
                  <option value="high">High</option>
                  <option value="medium">Medium</option>
                  <option value="low">Low</option>
                </select>
                <select
                  value={breachFilterStatus}
                  onChange={(e) => setBreachFilterStatus(e.target.value)}
                  className="px-2 py-1 bg-[#1E2026] border border-[#383A42] rounded text-xs focus:outline-none focus:border-[#F5C542]"
                >
                  <option value="">All Status</option>
                  <option value="open">Open</option>
                  <option value="investigating">Investigating</option>
                  <option value="resolved">Resolved</option>
                  <option value="closed">Closed</option>
                </select>
              </div>
            </div>
            <div className="flex-1 overflow-auto">
              <div className="p-3 space-y-2">
                {filteredBreaches.map((breach) => (
                  <div
                    key={breach.id}
                    className="bg-[#1E2026] p-3 rounded border border-[#383A42]"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <span
                            className="px-2 py-0.5 rounded text-xs font-medium"
                            style={{
                              backgroundColor: `${getSeverityBadgeColor(breach.severity)}20`,
                              color: getSeverityBadgeColor(breach.severity)
                            }}
                          >
                            {breach.severity}
                          </span>
                          <span className="text-xs text-[#A1A1AA]">{breach.type}</span>
                        </div>
                        <div className="text-sm">{breach.description}</div>
                      </div>
                    </div>
                    <div className="flex items-center justify-between text-xs text-[#A1A1AA]">
                      <span>{formatDate(breach.date)}</span>
                      <span
                        className="px-2 py-0.5 rounded"
                        style={{
                          backgroundColor: `${getStatusBadgeColor(breach.status)}20`,
                          color: getStatusBadgeColor(breach.status)
                        }}
                      >
                        {breach.status}
                      </span>
                    </div>
                    {breach.resolution && (
                      <div className="mt-2 pt-2 border-t border-[#383A42] text-xs text-[#10B981]">
                        ✓ {breach.resolution}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Generate Report Modal */}
      {showGenerateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg w-[500px]">
            <div className="flex items-center justify-between p-4 border-b border-[#383A42]">
              <h2 className="font-semibold">Generate Compliance Report</h2>
              <button
                onClick={() => setShowGenerateModal(false)}
                className="text-[#A1A1AA] hover:text-white"
              >
                <X size={20} />
              </button>
            </div>
            <div className="p-4 space-y-4">
              <div>
                <label className="block text-sm font-medium mb-2">Regulator</label>
                <select
                  value={generateForm.regulator}
                  onChange={(e) => setGenerateForm({ ...generateForm, regulator: e.target.value, type: e.target.value.replace(' ', '_') })}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded focus:outline-none focus:border-[#F5C542]"
                >
                  <option value="MiFID II">MiFID II</option>
                  <option value="EMIR">EMIR</option>
                  <option value="ASIC">ASIC</option>
                  <option value="FCA">FCA</option>
                  <option value="CySEC">CySEC</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">Report Period</label>
                <select
                  value={generateForm.period}
                  onChange={(e) => setGenerateForm({ ...generateForm, period: e.target.value })}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded focus:outline-none focus:border-[#F5C542]"
                >
                  <option value="Q1_2026">Q1 2026</option>
                  <option value="Q2_2026">Q2 2026</option>
                  <option value="Q3_2026">Q3 2026</option>
                  <option value="Q4_2025">Q4 2025</option>
                  <option value="Monthly_Jan_2026">Monthly - Jan 2026</option>
                  <option value="Annual_2025">Annual 2025</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">Format</label>
                <select
                  value={generateForm.format}
                  onChange={(e) => setGenerateForm({ ...generateForm, format: e.target.value })}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded focus:outline-none focus:border-[#F5C542]"
                >
                  <option value="PDF">PDF</option>
                  <option value="CSV">CSV</option>
                  <option value="XML">XML</option>
                </select>
              </div>
            </div>
            <div className="flex items-center justify-end gap-2 p-4 border-t border-[#383A42]">
              <button
                onClick={() => setShowGenerateModal(false)}
                className="px-4 py-2 bg-[#383A42] hover:bg-[#4A4C54] rounded text-sm transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleGenerateReport}
                className="px-4 py-2 bg-[#F5C542] hover:bg-[#E5B532] text-black rounded text-sm font-medium transition-colors"
              >
                Generate Report
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Report Detail View Modal */}
      {showDetailView && selectedReport && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg w-[700px] max-h-[80vh] overflow-hidden flex flex-col">
            <div className="flex items-center justify-between p-4 border-b border-[#383A42]">
              <h2 className="font-semibold">Report Detail - #{selectedReport.id}</h2>
              <button
                onClick={() => setShowDetailView(false)}
                className="text-[#A1A1AA] hover:text-white"
              >
                <X size={20} />
              </button>
            </div>
            <div className="flex-1 overflow-auto p-4 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-sm text-[#A1A1AA] mb-1">Report Type</div>
                  <div className="font-medium">{selectedReport.type.replace(/_/g, ' ')}</div>
                </div>
                <div>
                  <div className="text-sm text-[#A1A1AA] mb-1">Regulator</div>
                  <div className="font-medium">{selectedReport.regulator}</div>
                </div>
                <div>
                  <div className="text-sm text-[#A1A1AA] mb-1">Period</div>
                  <div className="font-medium">{selectedReport.period.replace(/_/g, ' ')}</div>
                </div>
                <div>
                  <div className="text-sm text-[#A1A1AA] mb-1">Format</div>
                  <div className="font-medium">{selectedReport.format}</div>
                </div>
                <div>
                  <div className="text-sm text-[#A1A1AA] mb-1">Status</div>
                  <span
                    className="px-2 py-1 rounded text-xs font-medium inline-block"
                    style={{
                      backgroundColor: `${getStatusBadgeColor(selectedReport.status)}20`,
                      color: getStatusBadgeColor(selectedReport.status)
                    }}
                  >
                    {selectedReport.status}
                  </span>
                </div>
                <div>
                  <div className="text-sm text-[#A1A1AA] mb-1">File Size</div>
                  <div className="font-medium">{formatFileSize(selectedReport.fileSize)}</div>
                </div>
                <div>
                  <div className="text-sm text-[#A1A1AA] mb-1">Generated At</div>
                  <div className="font-medium">{formatDateTime(selectedReport.generatedAt)}</div>
                </div>
                <div>
                  <div className="text-sm text-[#A1A1AA] mb-1">Generated By</div>
                  <div className="font-medium">{selectedReport.generatedBy}</div>
                </div>
              </div>

              <div className="border-t border-[#383A42] pt-4">
                <h3 className="font-semibold mb-3">Data Summary</h3>
                <div className="grid grid-cols-3 gap-4">
                  <div className="bg-[#121316] p-3 rounded">
                    <div className="text-sm text-[#A1A1AA]">Total Trades</div>
                    <div className="text-xl font-bold">{selectedReport.dataSummary.totalTrades.toLocaleString()}</div>
                  </div>
                  <div className="bg-[#121316] p-3 rounded">
                    <div className="text-sm text-[#A1A1AA]">Total Volume</div>
                    <div className="text-xl font-bold">${(selectedReport.dataSummary.totalVolume / 1000000).toFixed(1)}M</div>
                  </div>
                  <div className="bg-[#121316] p-3 rounded">
                    <div className="text-sm text-[#A1A1AA]">Total Clients</div>
                    <div className="text-xl font-bold">{selectedReport.dataSummary.totalClients}</div>
                  </div>
                  <div className="bg-[#121316] p-3 rounded">
                    <div className="text-sm text-[#A1A1AA]">Transactions</div>
                    <div className="text-xl font-bold">{selectedReport.dataSummary.totalTransactions}</div>
                  </div>
                  <div className="bg-[#121316] p-3 rounded">
                    <div className="text-sm text-[#A1A1AA]">Compliance Score</div>
                    <div className="text-xl font-bold text-[#10B981]">{selectedReport.dataSummary.complianceScore.toFixed(1)}%</div>
                  </div>
                  <div className="bg-[#121316] p-3 rounded">
                    <div className="text-sm text-[#A1A1AA]">Issues Found</div>
                    <div className="text-xl font-bold text-[#F59E0B]">{selectedReport.dataSummary.issuesFound}</div>
                  </div>
                </div>
              </div>

              {selectedReport.submittedTo && (
                <div className="border-t border-[#383A42] pt-4">
                  <h3 className="font-semibold mb-3">Submission History</h3>
                  <div className="bg-[#121316] p-3 rounded">
                    <div className="flex items-center gap-2 mb-2">
                      <CheckCircle2 className="text-[#10B981]" size={18} />
                      <span className="font-medium">Submitted to {selectedReport.submittedTo}</span>
                    </div>
                    <div className="text-sm text-[#A1A1AA]">
                      Submitted at: {selectedReport.submittedAt ? formatDateTime(selectedReport.submittedAt) : 'N/A'}
                    </div>
                  </div>
                </div>
              )}
            </div>
            <div className="flex items-center justify-end gap-2 p-4 border-t border-[#383A42]">
              <button
                onClick={() => setShowDetailView(false)}
                className="px-4 py-2 bg-[#383A42] hover:bg-[#4A4C54] rounded text-sm transition-colors"
              >
                Close
              </button>
              <button
                className="flex items-center gap-2 px-4 py-2 bg-[#F5C542] hover:bg-[#E5B532] text-black rounded text-sm font-medium transition-colors"
              >
                <Download size={16} />
                Download Report
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
