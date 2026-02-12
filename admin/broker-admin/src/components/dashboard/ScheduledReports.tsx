'use client';

import React, { useState, useEffect } from 'react';
import {
  FileText,
  Download,
  Calendar,
  Clock,
  CheckCircle2,
  XCircle,
  Play,
  Plus,
  Edit,
  Trash2,
  Copy,
  Filter,
  Mail,
} from 'lucide-react';

type ReportType =
  | 'TRADES'
  | 'PNL'
  | 'COMMISSIONS'
  | 'DEPOSITS_WITHDRAWALS'
  | 'CLIENT_ACTIVITY'
  | 'RISK_EXPOSURE'
  | 'LP_PERFORMANCE'
  | 'COMPLIANCE'
  | 'AUDIT_LOG'
  | 'SYSTEM_HEALTH';

type ReportFormat = 'PDF' | 'CSV' | 'EXCEL' | 'JSON';
type ScheduleType = 'DAILY' | 'WEEKLY' | 'MONTHLY' | 'QUARTERLY' | 'CRON';
type ExecutionStatus = 'SUCCESS' | 'FAILED' | 'RUNNING' | 'PENDING';

interface ScheduledReport {
  id: string;
  name: string;
  type: ReportType;
  format: ReportFormat;
  schedule: ScheduleType;
  cronExpression?: string;
  nextRun: number;
  lastRun?: number;
  lastStatus?: ExecutionStatus;
  active: boolean;
  recipients: string[];
  filters?: Record<string, any>;
}

interface ExecutionRecord {
  id: string;
  reportId: string;
  reportName: string;
  status: ExecutionStatus;
  startTime: number;
  endTime?: number;
  duration?: number;
  fileSize?: number;
  downloadUrl?: string;
  errorMessage?: string;
}

interface ReportTemplate {
  id: string;
  name: string;
  description: string;
  type: ReportType;
  formats: ReportFormat[];
  previewUrl?: string;
}

// Mock data generators
function generateMockReports(count: number): ScheduledReport[] {
  const types: ReportType[] = ['TRADES', 'PNL', 'COMMISSIONS', 'DEPOSITS_WITHDRAWALS', 'CLIENT_ACTIVITY'];
  const formats: ReportFormat[] = ['PDF', 'CSV', 'EXCEL'];
  const schedules: ScheduleType[] = ['DAILY', 'WEEKLY', 'MONTHLY'];
  const statuses: ExecutionStatus[] = ['SUCCESS', 'FAILED'];

  return Array.from({ length: count }, (_, i) => ({
    id: `report-${i + 1}`,
    name: `${types[i % types.length]} Report ${i + 1}`,
    type: types[i % types.length],
    format: formats[i % formats.length],
    schedule: schedules[i % schedules.length],
    nextRun: Date.now() + (i + 1) * 3600000,
    lastRun: Date.now() - (i + 1) * 86400000,
    lastStatus: statuses[i % statuses.length],
    active: Math.random() > 0.3,
    recipients: [`user${i + 1}@example.com`],
  }));
}

function generateMockExecutions(count: number): ExecutionRecord[] {
  const statuses: ExecutionStatus[] = ['SUCCESS', 'FAILED'];

  return Array.from({ length: count }, (_, i) => {
    const status = statuses[i % statuses.length];
    const startTime = Date.now() - (i + 1) * 3600000;
    const duration = 2000 + Math.random() * 8000;

    return {
      id: `exec-${i + 1}`,
      reportId: `report-${(i % 5) + 1}`,
      reportName: `Report ${(i % 5) + 1}`,
      status,
      startTime,
      endTime: status !== 'RUNNING' ? startTime + duration : undefined,
      duration: status !== 'RUNNING' ? duration : undefined,
      fileSize: status === 'SUCCESS' ? 100000 + Math.random() * 500000 : undefined,
      downloadUrl: status === 'SUCCESS' ? `/downloads/report-${i + 1}.pdf` : undefined,
      errorMessage: status === 'FAILED' ? 'Database connection timeout' : undefined,
    };
  });
}

const REPORT_TEMPLATES: ReportTemplate[] = [
  {
    id: 'tpl-1',
    name: 'Daily Trading Summary',
    description: 'Comprehensive daily trading activity report with volume, P&L, and top performers',
    type: 'TRADES',
    formats: ['PDF', 'EXCEL', 'CSV'],
  },
  {
    id: 'tpl-2',
    name: 'Weekly P&L Analysis',
    description: 'Weekly profit and loss breakdown by symbol, client, and trading group',
    type: 'PNL',
    formats: ['PDF', 'EXCEL'],
  },
  {
    id: 'tpl-3',
    name: 'Monthly Compliance Report',
    description: 'Regulatory compliance report with KYC status, transaction monitoring, and audit logs',
    type: 'COMPLIANCE',
    formats: ['PDF'],
  },
  {
    id: 'tpl-4',
    name: 'Commission Statement',
    description: 'Detailed commission breakdown for affiliates and introducing brokers',
    type: 'COMMISSIONS',
    formats: ['PDF', 'CSV'],
  },
  {
    id: 'tpl-5',
    name: 'Risk Exposure Dashboard',
    description: 'Real-time risk exposure by symbol, client, and group with margin utilization',
    type: 'RISK_EXPOSURE',
    formats: ['PDF', 'EXCEL'],
  },
];

export default function ScheduledReports() {
  const [reports, setReports] = useState<ScheduledReport[]>([]);
  const [executions, setExecutions] = useState<ExecutionRecord[]>([]);
  const [selectedReport, setSelectedReport] = useState<ScheduledReport | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [runningReportId, setRunningReportId] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'list' | 'calendar'>('list');

  useEffect(() => {
    setReports(generateMockReports(12));
    setExecutions(generateMockExecutions(30));
  }, []);

  const handleToggleActive = (reportId: string) => {
    setReports((prev) =>
      prev.map((r) => (r.id === reportId ? { ...r, active: !r.active } : r))
    );
  };

  const handleRunNow = async (reportId: string) => {
    setRunningReportId(reportId);
    // Simulate execution
    setTimeout(() => {
      const newExecution: ExecutionRecord = {
        id: `exec-${Date.now()}`,
        reportId,
        reportName: reports.find((r) => r.id === reportId)?.name || 'Report',
        status: 'SUCCESS',
        startTime: Date.now() - 3000,
        endTime: Date.now(),
        duration: 3000,
        fileSize: 250000,
        downloadUrl: `/downloads/${reportId}.pdf`,
      };
      setExecutions((prev) => [newExecution, ...prev]);
      setRunningReportId(null);
    }, 3000);
  };

  const handleCreateFromTemplate = (template: ReportTemplate) => {
    const newReport: ScheduledReport = {
      id: `report-${Date.now()}`,
      name: template.name,
      type: template.type,
      format: template.formats[0],
      schedule: 'DAILY',
      nextRun: Date.now() + 86400000,
      active: true,
      recipients: [],
    };
    setReports((prev) => [newReport, ...prev]);
  };

  // Calculate stats
  const totalReports = reports.length;
  const activeReports = reports.filter((r) => r.active).length;
  const executionsToday = executions.filter(
    (e) => new Date(e.startTime).toDateString() === new Date().toDateString()
  ).length;
  const successfulExecutions = executions.filter((e) => e.status === 'SUCCESS').length;
  const successRate = executions.length > 0 ? (successfulExecutions / executions.length) * 100 : 0;
  const failedCount = executions.filter((e) => e.status === 'FAILED').length;

  const getTypeColor = (type: ReportType) => {
    const colors: Record<ReportType, string> = {
      TRADES: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
      PNL: 'bg-green-500/10 text-green-400 border-green-500/20',
      COMMISSIONS: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
      DEPOSITS_WITHDRAWALS: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
      CLIENT_ACTIVITY: 'bg-pink-500/10 text-pink-400 border-pink-500/20',
      RISK_EXPOSURE: 'bg-red-500/10 text-red-400 border-red-500/20',
      LP_PERFORMANCE: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20',
      COMPLIANCE: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
      AUDIT_LOG: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20',
      SYSTEM_HEALTH: 'bg-teal-500/10 text-teal-400 border-teal-500/20',
    };
    return colors[type];
  };

  const getFormatIcon = (format: ReportFormat) => {
    return <FileText size={14} />;
  };

  const getStatusBadge = (status: ExecutionStatus) => {
    const styles: Record<ExecutionStatus, string> = {
      SUCCESS: 'bg-green-500/10 text-green-400 border-green-500/20',
      FAILED: 'bg-red-500/10 text-red-400 border-red-500/20',
      RUNNING: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
      PENDING: 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20',
    };
    const icons: Record<ExecutionStatus, React.ReactNode> = {
      SUCCESS: <CheckCircle2 size={12} />,
      FAILED: <XCircle size={12} />,
      RUNNING: <Clock size={12} className="animate-spin" />,
      PENDING: <Clock size={12} />,
    };

    return (
      <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs border ${styles[status]}`}>
        {icons[status]}
        {status}
      </span>
    );
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDuration = (ms: number) => {
    const seconds = Math.floor(ms / 1000);
    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    return `${minutes}m ${seconds % 60}s`;
  };

  // Render mini calendar
  const renderCalendar = () => {
    const today = new Date();
    const daysInMonth = new Date(today.getFullYear(), today.getMonth() + 1, 0).getDate();
    const firstDay = new Date(today.getFullYear(), today.getMonth(), 1).getDay();

    const scheduledDates = new Set(
      reports
        .filter((r) => r.active)
        .map((r) => new Date(r.nextRun).getDate())
    );

    const days = [];
    for (let i = 0; i < firstDay; i++) {
      days.push(<div key={`empty-${i}`} className="h-8" />);
    }
    for (let day = 1; day <= daysInMonth; day++) {
      const isScheduled = scheduledDates.has(day);
      const isToday = day === today.getDate();
      days.push(
        <div
          key={day}
          className={`h-8 flex items-center justify-center text-xs rounded ${
            isToday ? 'bg-blue-500/20 text-blue-400 font-bold' : ''
          } ${isScheduled && !isToday ? 'bg-green-500/10 text-green-400' : 'text-zinc-400'}`}
        >
          {day}
        </div>
      );
    }

    return (
      <div className="grid grid-cols-7 gap-1">
        {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map((day) => (
          <div key={day} className="h-6 flex items-center justify-center text-xs text-zinc-500 font-bold">
            {day}
          </div>
        ))}
        {days}
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full bg-zinc-900 text-zinc-100 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-zinc-800 border-b border-zinc-700 flex-shrink-0">
        <div>
          <h2 className="text-lg font-bold text-zinc-100">Scheduled Reports</h2>
          <p className="text-xs text-zinc-400">Manage automated report generation and delivery</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setViewMode(viewMode === 'list' ? 'calendar' : 'list')}
            className="px-3 py-1.5 bg-zinc-700 hover:bg-zinc-600 rounded text-xs flex items-center gap-1"
          >
            <Calendar size={14} />
            {viewMode === 'list' ? 'Calendar' : 'List'} View
          </button>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 rounded text-xs flex items-center gap-1 font-bold"
          >
            <Plus size={14} />
            Create Report
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4 space-y-4">
        {/* Stats Cards */}
        <div className="grid grid-cols-6 gap-4">
          <div className="bg-zinc-800 rounded p-4 border border-zinc-700">
            <div className="text-xs text-zinc-400 mb-1">Total Reports</div>
            <div className="text-2xl font-bold text-zinc-100">{totalReports}</div>
          </div>
          <div className="bg-zinc-800 rounded p-4 border border-zinc-700">
            <div className="text-xs text-zinc-400 mb-1">Active</div>
            <div className="text-2xl font-bold text-green-400">{activeReports}</div>
          </div>
          <div className="bg-zinc-800 rounded p-4 border border-zinc-700">
            <div className="text-xs text-zinc-400 mb-1">Executions Today</div>
            <div className="text-2xl font-bold text-blue-400">{executionsToday}</div>
          </div>
          <div className="bg-zinc-800 rounded p-4 border border-zinc-700">
            <div className="text-xs text-zinc-400 mb-1">Success Rate</div>
            <div className="text-2xl font-bold text-green-400">{successRate.toFixed(1)}%</div>
          </div>
          <div className="bg-zinc-800 rounded p-4 border border-zinc-700">
            <div className="text-xs text-zinc-400 mb-1">Failed</div>
            <div className="text-2xl font-bold text-red-400">{failedCount}</div>
          </div>
          <div className="bg-zinc-800 rounded p-4 border border-zinc-700">
            <div className="text-xs text-zinc-400 mb-1">Next Run</div>
            <div className="text-sm font-mono text-zinc-100">
              {reports.length > 0
                ? new Date(Math.min(...reports.filter((r) => r.active).map((r) => r.nextRun))).toLocaleTimeString([], {
                    hour: '2-digit',
                    minute: '2-digit',
                  })
                : 'N/A'}
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="grid grid-cols-3 gap-4">
          {/* Left: Reports List or Calendar */}
          <div className="col-span-2 space-y-4">
            {viewMode === 'list' ? (
              <div className="bg-zinc-800 rounded border border-zinc-700">
                <div className="px-4 py-3 border-b border-zinc-700">
                  <h3 className="text-sm font-bold text-zinc-100">Reports</h3>
                </div>
                <div className="overflow-auto max-h-[600px]">
                  <table className="w-full text-xs">
                    <thead className="sticky top-0 bg-zinc-800">
                      <tr className="border-b border-zinc-700">
                        <th className="text-left py-2 px-3 text-zinc-400 font-normal">Name</th>
                        <th className="text-left py-2 px-3 text-zinc-400 font-normal">Type</th>
                        <th className="text-center py-2 px-3 text-zinc-400 font-normal">Format</th>
                        <th className="text-left py-2 px-3 text-zinc-400 font-normal">Schedule</th>
                        <th className="text-left py-2 px-3 text-zinc-400 font-normal">Next Run</th>
                        <th className="text-center py-2 px-3 text-zinc-400 font-normal">Last Status</th>
                        <th className="text-center py-2 px-3 text-zinc-400 font-normal">Active</th>
                        <th className="text-center py-2 px-3 text-zinc-400 font-normal">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {reports.map((report) => (
                        <tr
                          key={report.id}
                          onClick={() => setSelectedReport(report)}
                          className={`border-b border-zinc-700 hover:bg-zinc-750 cursor-pointer ${
                            selectedReport?.id === report.id ? 'bg-zinc-750' : ''
                          }`}
                        >
                          <td className="py-2 px-3 font-medium">{report.name}</td>
                          <td className="py-2 px-3">
                            <span className={`inline-block px-2 py-0.5 rounded text-xs border ${getTypeColor(report.type)}`}>
                              {report.type.replace('_', ' ')}
                            </span>
                          </td>
                          <td className="text-center py-2 px-3">{getFormatIcon(report.format)}</td>
                          <td className="py-2 px-3 text-zinc-400">{report.schedule}</td>
                          <td className="py-2 px-3 text-zinc-400">
                            {new Date(report.nextRun).toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                          </td>
                          <td className="text-center py-2 px-3">
                            {report.lastStatus ? getStatusBadge(report.lastStatus) : '-'}
                          </td>
                          <td className="text-center py-2 px-3">
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleToggleActive(report.id);
                              }}
                              className={`w-10 h-5 rounded-full relative transition-colors ${
                                report.active ? 'bg-green-500' : 'bg-zinc-700'
                              }`}
                            >
                              <div
                                className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${
                                  report.active ? 'translate-x-5' : 'translate-x-0.5'
                                }`}
                              />
                            </button>
                          </td>
                          <td className="text-center py-2 px-3">
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleRunNow(report.id);
                              }}
                              disabled={runningReportId === report.id}
                              className="px-2 py-1 bg-blue-600 hover:bg-blue-700 disabled:bg-zinc-700 rounded text-xs flex items-center gap-1 mx-auto"
                            >
                              {runningReportId === report.id ? (
                                <Clock size={12} className="animate-spin" />
                              ) : (
                                <Play size={12} />
                              )}
                              Run
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            ) : (
              <div className="bg-zinc-800 rounded p-4 border border-zinc-700">
                <h3 className="text-sm font-bold text-zinc-100 mb-4">
                  {new Date().toLocaleString('default', { month: 'long', year: 'numeric' })}
                </h3>
                {renderCalendar()}
                <div className="mt-4 flex items-center gap-4 text-xs">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 bg-blue-500/20 rounded" />
                    <span className="text-zinc-400">Today</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 bg-green-500/10 rounded" />
                    <span className="text-zinc-400">Scheduled</span>
                  </div>
                </div>
              </div>
            )}

            {/* Report Templates */}
            <div className="bg-zinc-800 rounded border border-zinc-700">
              <div className="px-4 py-3 border-b border-zinc-700">
                <h3 className="text-sm font-bold text-zinc-100">Report Templates</h3>
              </div>
              <div className="p-4 grid grid-cols-5 gap-3">
                {REPORT_TEMPLATES.map((template) => (
                  <div
                    key={template.id}
                    className="bg-zinc-750 rounded p-3 border border-zinc-700 hover:border-blue-500/50 cursor-pointer"
                    onClick={() => handleCreateFromTemplate(template)}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <FileText size={16} className="text-blue-400" />
                      <Copy size={12} className="text-zinc-500" />
                    </div>
                    <div className="text-xs font-bold text-zinc-100 mb-1">{template.name}</div>
                    <div className="text-xs text-zinc-500 mb-2 line-clamp-2">{template.description}</div>
                    <div className="flex items-center gap-1 flex-wrap">
                      {template.formats.map((fmt) => (
                        <span key={fmt} className="px-1.5 py-0.5 bg-zinc-700 rounded text-xs text-zinc-400">
                          {fmt}
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Right: Report Detail / Execution History */}
          <div className="space-y-4">
            {selectedReport ? (
              <>
                {/* Report Detail */}
                <div className="bg-zinc-800 rounded border border-zinc-700">
                  <div className="px-4 py-3 border-b border-zinc-700 flex items-center justify-between">
                    <h3 className="text-sm font-bold text-zinc-100">Report Details</h3>
                    <div className="flex items-center gap-2">
                      <button className="p-1 hover:bg-zinc-700 rounded">
                        <Edit size={14} className="text-zinc-400" />
                      </button>
                      <button className="p-1 hover:bg-zinc-700 rounded">
                        <Trash2 size={14} className="text-red-400" />
                      </button>
                    </div>
                  </div>
                  <div className="p-4 space-y-3">
                    <div>
                      <div className="text-xs text-zinc-500 mb-1">Name</div>
                      <div className="text-sm font-medium text-zinc-100">{selectedReport.name}</div>
                    </div>
                    <div>
                      <div className="text-xs text-zinc-500 mb-1">Type</div>
                      <span className={`inline-block px-2 py-0.5 rounded text-xs border ${getTypeColor(selectedReport.type)}`}>
                        {selectedReport.type.replace('_', ' ')}
                      </span>
                    </div>
                    <div>
                      <div className="text-xs text-zinc-500 mb-1">Format</div>
                      <div className="text-sm text-zinc-100">{selectedReport.format}</div>
                    </div>
                    <div>
                      <div className="text-xs text-zinc-500 mb-1">Schedule</div>
                      <div className="text-sm text-zinc-100">{selectedReport.schedule}</div>
                    </div>
                    <div>
                      <div className="text-xs text-zinc-500 mb-1">Recipients</div>
                      <div className="flex items-center gap-1 flex-wrap">
                        {selectedReport.recipients.length > 0 ? (
                          selectedReport.recipients.map((email, i) => (
                            <span key={i} className="inline-flex items-center gap-1 px-2 py-0.5 bg-zinc-700 rounded text-xs text-zinc-300">
                              <Mail size={10} />
                              {email}
                            </span>
                          ))
                        ) : (
                          <span className="text-xs text-zinc-500">No recipients</span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Execution History */}
                <div className="bg-zinc-800 rounded border border-zinc-700">
                  <div className="px-4 py-3 border-b border-zinc-700">
                    <h3 className="text-sm font-bold text-zinc-100">Execution History</h3>
                  </div>
                  <div className="p-4 max-h-96 overflow-y-auto space-y-2">
                    {executions
                      .filter((e) => e.reportId === selectedReport.id)
                      .map((execution) => (
                        <div key={execution.id} className="bg-zinc-750 rounded p-3 border border-zinc-700">
                          <div className="flex items-start justify-between mb-2">
                            {getStatusBadge(execution.status)}
                            <span className="text-xs text-zinc-500">
                              {new Date(execution.startTime).toLocaleString([], {
                                month: 'short',
                                day: 'numeric',
                                hour: '2-digit',
                                minute: '2-digit',
                              })}
                            </span>
                          </div>
                          {execution.duration && (
                            <div className="text-xs text-zinc-400 mb-1">Duration: {formatDuration(execution.duration)}</div>
                          )}
                          {execution.fileSize && (
                            <div className="text-xs text-zinc-400 mb-1">Size: {formatFileSize(execution.fileSize)}</div>
                          )}
                          {execution.errorMessage && (
                            <div className="text-xs text-red-400 mb-2">{execution.errorMessage}</div>
                          )}
                          {execution.downloadUrl && (
                            <button className="flex items-center gap-1 text-xs text-blue-400 hover:text-blue-300">
                              <Download size={12} />
                              Download
                            </button>
                          )}
                        </div>
                      ))}
                    {executions.filter((e) => e.reportId === selectedReport.id).length === 0 && (
                      <div className="text-center py-8 text-zinc-500 text-xs">No execution history</div>
                    )}
                  </div>
                </div>
              </>
            ) : (
              <div className="bg-zinc-800 rounded border border-zinc-700 p-8 text-center">
                <FileText size={48} className="text-zinc-700 mx-auto mb-4" />
                <div className="text-sm text-zinc-500">Select a report to view details</div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Create Report Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-zinc-800 rounded-lg border border-zinc-700 w-[600px] max-h-[80vh] overflow-y-auto">
            <div className="px-6 py-4 border-b border-zinc-700 flex items-center justify-between">
              <h3 className="text-lg font-bold text-zinc-100">Create Scheduled Report</h3>
              <button onClick={() => setShowCreateModal(false)} className="text-zinc-500 hover:text-zinc-300">
                ×
              </button>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Report Name</label>
                <input
                  type="text"
                  placeholder="Enter report name"
                  className="w-full bg-zinc-700 text-zinc-100 text-sm px-3 py-2 rounded border border-zinc-600"
                />
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Report Type</label>
                <select className="w-full bg-zinc-700 text-zinc-100 text-sm px-3 py-2 rounded border border-zinc-600">
                  <option value="TRADES">Trades</option>
                  <option value="PNL">P&L</option>
                  <option value="COMMISSIONS">Commissions</option>
                  <option value="DEPOSITS_WITHDRAWALS">Deposits & Withdrawals</option>
                  <option value="CLIENT_ACTIVITY">Client Activity</option>
                  <option value="RISK_EXPOSURE">Risk Exposure</option>
                  <option value="LP_PERFORMANCE">LP Performance</option>
                  <option value="COMPLIANCE">Compliance</option>
                  <option value="AUDIT_LOG">Audit Log</option>
                  <option value="SYSTEM_HEALTH">System Health</option>
                </select>
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Format</label>
                <div className="flex gap-2">
                  {(['PDF', 'CSV', 'EXCEL', 'JSON'] as ReportFormat[]).map((fmt) => (
                    <button
                      key={fmt}
                      className="flex-1 px-3 py-2 bg-zinc-700 hover:bg-zinc-600 rounded text-sm border border-zinc-600"
                    >
                      {fmt}
                    </button>
                  ))}
                </div>
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Schedule</label>
                <select className="w-full bg-zinc-700 text-zinc-100 text-sm px-3 py-2 rounded border border-zinc-600">
                  <option value="DAILY">Daily</option>
                  <option value="WEEKLY">Weekly</option>
                  <option value="MONTHLY">Monthly</option>
                  <option value="QUARTERLY">Quarterly</option>
                  <option value="CRON">Custom (Cron)</option>
                </select>
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Recipients (comma-separated emails)</label>
                <input
                  type="text"
                  placeholder="user1@example.com, user2@example.com"
                  className="w-full bg-zinc-700 text-zinc-100 text-sm px-3 py-2 rounded border border-zinc-600"
                />
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Filters (optional)</label>
                <textarea
                  placeholder="JSON filter configuration"
                  rows={3}
                  className="w-full bg-zinc-700 text-zinc-100 text-sm px-3 py-2 rounded border border-zinc-600 font-mono"
                />
              </div>
            </div>
            <div className="px-6 py-4 border-t border-zinc-700 flex justify-end gap-3">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 rounded text-sm"
              >
                Cancel
              </button>
              <button
                onClick={() => {
                  setShowCreateModal(false);
                }}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm font-bold"
              >
                Create Report
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
