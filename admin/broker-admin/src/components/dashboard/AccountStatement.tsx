'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
  FileText,
  Download,
  Calendar,
  Users,
  Filter,
  RefreshCw,
  Clock,
  FileSpreadsheet,
  Mail,
  TrendingUp,
  BarChart3,
  CheckSquare,
  Square
} from 'lucide-react';
import { API_CONFIG } from '../../config/api';

// ============================================
// Type Definitions (matching backend)
// ============================================

interface TradeEntry {
  id: number;
  clientId: number;
  symbol: string;
  type: string;
  volume: number;
  entryPrice: number;
  exitPrice: number;
  pnl: number;
  commission: number;
  swap: number;
  openTime: string;
  closeTime: string;
}

interface TransactionEntry {
  id: number;
  clientId: number;
  type: string;
  amount: number;
  currency: string;
  status: string;
  timestamp: string;
  reference: string;
}

interface DailyBalance {
  date: string;
  openingBalance: number;
  deposits: number;
  withdrawals: number;
  tradingPnl: number;
  commissions: number;
  swaps: number;
  closingBalance: number;
}

interface GeneratedStatement {
  clientId: number;
  clientName: string;
  fromDate: string;
  toDate: string;
  trades: TradeEntry[];
  transactions: TransactionEntry[];
  dailyBalances: DailyBalance[];
  totalPnl: number;
  totalDeposits: number;
  totalWithdrawals: number;
  totalCommissions: number;
  finalBalance: number;
  generatedAt: string;
}

interface StatementTemplate {
  id: number;
  name: string;
  type: string;
  description: string;
  defaultDays: number;
}

interface StatementHistory {
  id: number;
  clientId: number;
  clientName: string;
  format: string;
  fromDate: string;
  toDate: string;
  generatedAt: string;
  downloadUrl: string;
  fileSize: number;
}

interface ScheduledStatement {
  id: number;
  clientId: number;
  clientName: string;
  frequency: string;
  format: string;
  emailTo: string;
  isActive: boolean;
  nextRun: string;
  createdAt: string;
  updatedAt: string;
}

interface StatementStats {
  totalGenerated: number;
  popularFormats: { [key: string]: number };
  avgGenerationTime: number;
  topClients: Array<{
    clientId: number;
    clientName: string;
    count: number;
  }>;
  lastUpdated: string;
}

// ============================================
// Helper Functions
// ============================================

const formatCurrency = (amount: number): string => {
  return `$${amount.toFixed(2)}`;
};

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
};

const getFormatColor = (format: string): string => {
  const colors: { [key: string]: string } = {
    csv: '#10B981',
    pdf: '#EF4444',
    xlsx: '#3B82F6'
  };
  return colors[format] || '#6B7280';
};

const getFormatIcon = (format: string) => {
  switch (format) {
    case 'csv':
      return <FileSpreadsheet size={16} />;
    case 'pdf':
      return <FileText size={16} />;
    case 'xlsx':
      return <FileSpreadsheet size={16} />;
    default:
      return <FileText size={16} />;
  }
};

// ============================================
// Main Component
// ============================================

const AccountStatement: React.FC = () => {
  // State for data
  const [templates, setTemplates] = useState<StatementTemplate[]>([]);
  const [history, setHistory] = useState<StatementHistory[]>([]);
  const [scheduled, setScheduled] = useState<ScheduledStatement[]>([]);
  const [stats, setStats] = useState<StatementStats | null>(null);
  const [generatedStatement, setGeneratedStatement] = useState<GeneratedStatement | null>(null);

  // UI State
  const [selectedClientId, setSelectedClientId] = useState<number>(1);
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [includeTrades, setIncludeTrades] = useState(true);
  const [includeDeposits, setIncludeDeposits] = useState(true);
  const [includeFees, setIncludeFees] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  // Fetch all data on mount
  useEffect(() => {
    // Set default date range (last 30 days)
    const today = new Date();
    const thirtyDaysAgo = new Date(today);
    thirtyDaysAgo.setDate(today.getDate() - 30);

    setFromDate(thirtyDaysAgo.toISOString().split('T')[0]);
    setToDate(today.toISOString().split('T')[0]);

    fetchAllData();
  }, []);

  const fetchAllData = async () => {
    setIsLoading(true);
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
      const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` })
      };

      const [templatesRes, historyRes, scheduledRes, statsRes] = await Promise.all([
        fetch(API_CONFIG.STATEMENT_TEMPLATES, { headers }),
        fetch(API_CONFIG.STATEMENT_HISTORY, { headers }),
        fetch(API_CONFIG.STATEMENT_SCHEDULED, { headers }),
        fetch(API_CONFIG.STATEMENT_STATS, { headers })
      ]);

      const templatesData = await templatesRes.json();
      const historyData = await historyRes.json();
      const scheduledData = await scheduledRes.json();
      const statsData = await statsRes.json();

      setTemplates(templatesData.templates || []);
      setHistory(historyData.history || []);
      setScheduled(scheduledData.schedules || []);
      setStats(statsData || null);
    } catch (error) {
      console.error('Error fetching statement data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleGenerateStatement = async () => {
    if (!selectedClientId || !fromDate || !toDate) {
      alert('Please select a client and date range');
      return;
    }

    setIsGenerating(true);
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
      const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` })
      };

      const url = `${API_CONFIG.STATEMENT_GENERATE(selectedClientId)}?from=${fromDate}&to=${toDate}&includeTrades=${includeTrades}&includeDeposits=${includeDeposits}&includeFees=${includeFees}`;

      const res = await fetch(url, { headers });
      const data = await res.json();

      setGeneratedStatement(data);
    } catch (error) {
      console.error('Error generating statement:', error);
    } finally {
      setIsGenerating(false);
    }
  };

  const handleExport = async (format: 'csv' | 'pdf' | 'xlsx') => {
    if (!generatedStatement) {
      alert('Please generate a statement first');
      return;
    }

    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
      const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` })
      };

      const res = await fetch(API_CONFIG.STATEMENT_EXPORT, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          clientId: generatedStatement.clientId,
          fromDate: generatedStatement.fromDate,
          toDate: generatedStatement.toDate,
          format
        })
      });

      const data = await res.json();

      // Refresh history
      fetchAllData();

      // In production, this would trigger a download
      alert(`Statement exported as ${format.toUpperCase()}. Download URL: ${data.downloadUrl}`);
    } catch (error) {
      console.error('Error exporting statement:', error);
    }
  };

  const handleQuickGenerate = (template: StatementTemplate) => {
    const today = new Date();
    const fromDate = new Date(today);
    fromDate.setDate(today.getDate() - template.defaultDays);

    setFromDate(fromDate.toISOString().split('T')[0]);
    setToDate(today.toISOString().split('T')[0]);
    setIncludeTrades(true);
    setIncludeDeposits(true);
    setIncludeFees(true);

    // Auto-generate
    setTimeout(() => handleGenerateStatement(), 100);
  };

  const handleToggleSchedule = async (scheduleId: number, currentActive: boolean) => {
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
      const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` })
      };

      await fetch(API_CONFIG.STATEMENT_SCHEDULE_UPDATE(scheduleId), {
        method: 'PUT',
        headers,
        body: JSON.stringify({ isActive: !currentActive })
      });

      setScheduled(prev => prev.map(sch =>
        sch.id === scheduleId ? { ...sch, isActive: !currentActive } : sch
      ));
    } catch (error) {
      console.error('Error toggling schedule:', error);
    }
  };

  // Generate mock client list (1-100)
  const clientOptions = useMemo(() => {
    const clients = [];
    for (let i = 1; i <= 100; i++) {
      clients.push({ id: i, name: `Client ${i}` });
    }
    return clients;
  }, []);

  // ============================================
  // Render Stats Cards
  // ============================================

  const renderStatsCards = () => {
    if (!stats) return null;

    const mostPopularFormat = Object.entries(stats.popularFormats).sort((a, b) => b[1] - a[1])[0];

    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-[#9CA3AF]">Total Generated</span>
            <FileText size={18} className="text-[#3B82F6]" />
          </div>
          <div className="text-2xl font-semibold text-white">{stats.totalGenerated}</div>
        </div>

        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-[#9CA3AF]">Popular Format</span>
            <FileSpreadsheet size={18} className="text-[#10B981]" />
          </div>
          <div className="text-2xl font-semibold text-white uppercase">
            {mostPopularFormat ? mostPopularFormat[0] : 'N/A'}
          </div>
        </div>

        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-[#9CA3AF]">Avg Generation Time</span>
            <Clock size={18} className="text-[#F59E0B]" />
          </div>
          <div className="text-2xl font-semibold text-white">{stats.avgGenerationTime}ms</div>
        </div>

        <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-[#9CA3AF]">Top Client</span>
            <Users size={18} className="text-[#F5C542]" />
          </div>
          <div className="text-lg font-semibold text-white truncate">
            {stats.topClients[0]?.clientName || 'N/A'}
          </div>
          <div className="text-xs text-[#9CA3AF]">
            {stats.topClients[0]?.count || 0} statements
          </div>
        </div>
      </div>
    );
  };

  // ============================================
  // Render Statement Generator Panel
  // ============================================

  const renderGeneratorPanel = () => {
    return (
      <div className="bg-[#1E2026] border border-[#383A42] rounded mb-6">
        <div className="p-4 border-b border-[#383A42]">
          <div className="flex items-center gap-2">
            <Filter size={20} className="text-[#F5C542]" />
            <h2 className="text-lg font-semibold text-white">Generate Statement</h2>
          </div>
        </div>

        <div className="p-4 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm text-[#9CA3AF] mb-2">Client</label>
              <select
                value={selectedClientId}
                onChange={(e) => setSelectedClientId(Number(e.target.value))}
                className="w-full px-4 py-2 bg-[#121316] border border-[#383A42] rounded text-white focus:outline-none focus:border-[#F5C542]"
              >
                {clientOptions.map(client => (
                  <option key={client.id} value={client.id}>{client.name}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm text-[#9CA3AF] mb-2">From Date</label>
              <input
                type="date"
                value={fromDate}
                onChange={(e) => setFromDate(e.target.value)}
                className="w-full px-4 py-2 bg-[#121316] border border-[#383A42] rounded text-white focus:outline-none focus:border-[#F5C542]"
              />
            </div>

            <div>
              <label className="block text-sm text-[#9CA3AF] mb-2">To Date</label>
              <input
                type="date"
                value={toDate}
                onChange={(e) => setToDate(e.target.value)}
                className="w-full px-4 py-2 bg-[#121316] border border-[#383A42] rounded text-white focus:outline-none focus:border-[#F5C542]"
              />
            </div>
          </div>

          <div className="flex gap-6">
            <label className="flex items-center gap-2 cursor-pointer">
              <div onClick={() => setIncludeTrades(!includeTrades)}>
                {includeTrades ? (
                  <CheckSquare size={20} className="text-[#F5C542]" />
                ) : (
                  <Square size={20} className="text-[#6B7280]" />
                )}
              </div>
              <span className="text-white">Include Trades</span>
            </label>

            <label className="flex items-center gap-2 cursor-pointer">
              <div onClick={() => setIncludeDeposits(!includeDeposits)}>
                {includeDeposits ? (
                  <CheckSquare size={20} className="text-[#F5C542]" />
                ) : (
                  <Square size={20} className="text-[#6B7280]" />
                )}
              </div>
              <span className="text-white">Include Deposits/Withdrawals</span>
            </label>

            <label className="flex items-center gap-2 cursor-pointer">
              <div onClick={() => setIncludeFees(!includeFees)}>
                {includeFees ? (
                  <CheckSquare size={20} className="text-[#F5C542]" />
                ) : (
                  <Square size={20} className="text-[#6B7280]" />
                )}
              </div>
              <span className="text-white">Include Fees/Commissions</span>
            </label>
          </div>

          <button
            onClick={handleGenerateStatement}
            disabled={isGenerating}
            className="w-full px-4 py-2 bg-[#F5C542] text-[#121316] rounded font-semibold hover:bg-[#F5D563] transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {isGenerating ? (
              <>
                <RefreshCw size={18} className="animate-spin" />
                Generating...
              </>
            ) : (
              <>
                <FileText size={18} />
                Generate Statement
              </>
            )}
          </button>
        </div>
      </div>
    );
  };

  // ============================================
  // Render Generated Statement View
  // ============================================

  const renderGeneratedStatement = () => {
    if (!generatedStatement) return null;

    return (
      <div className="bg-[#1E2026] border border-[#383A42] rounded mb-6">
        <div className="p-4 border-b border-[#383A42]">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-white">Statement: {generatedStatement.clientName}</h2>
              <p className="text-sm text-[#9CA3AF]">
                {generatedStatement.fromDate} to {generatedStatement.toDate}
              </p>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => handleExport('csv')}
                className="px-3 py-1.5 bg-[#10B981] text-white rounded text-sm hover:bg-[#059669] transition-colors flex items-center gap-1"
              >
                {getFormatIcon('csv')}
                CSV
              </button>
              <button
                onClick={() => handleExport('pdf')}
                className="px-3 py-1.5 bg-[#EF4444] text-white rounded text-sm hover:bg-[#DC2626] transition-colors flex items-center gap-1"
              >
                {getFormatIcon('pdf')}
                PDF
              </button>
              <button
                onClick={() => handleExport('xlsx')}
                className="px-3 py-1.5 bg-[#3B82F6] text-white rounded text-sm hover:bg-[#2563EB] transition-colors flex items-center gap-1"
              >
                {getFormatIcon('xlsx')}
                Excel
              </button>
            </div>
          </div>
        </div>

        <div className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
            <div className="bg-[#121316] p-3 rounded">
              <div className="text-xs text-[#9CA3AF] mb-1">Total P&L</div>
              <div className={`text-lg font-semibold ${generatedStatement.totalPnl >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                {formatCurrency(generatedStatement.totalPnl)}
              </div>
            </div>
            <div className="bg-[#121316] p-3 rounded">
              <div className="text-xs text-[#9CA3AF] mb-1">Total Deposits</div>
              <div className="text-lg font-semibold text-[#10B981]">{formatCurrency(generatedStatement.totalDeposits)}</div>
            </div>
            <div className="bg-[#121316] p-3 rounded">
              <div className="text-xs text-[#9CA3AF] mb-1">Total Withdrawals</div>
              <div className="text-lg font-semibold text-[#EF4444]">{formatCurrency(generatedStatement.totalWithdrawals)}</div>
            </div>
            <div className="bg-[#121316] p-3 rounded">
              <div className="text-xs text-[#9CA3AF] mb-1">Final Balance</div>
              <div className="text-lg font-semibold text-white">{formatCurrency(generatedStatement.finalBalance)}</div>
            </div>
          </div>

          <h3 className="text-md font-semibold text-white mb-3">Daily Breakdown</h3>
          <div className="overflow-x-auto max-h-[400px] overflow-y-auto">
            <table className="w-full">
              <thead className="sticky top-0 bg-[#121316]">
                <tr>
                  <th className="text-left p-2 text-sm font-medium text-[#9CA3AF]">Date</th>
                  <th className="text-right p-2 text-sm font-medium text-[#9CA3AF]">Opening</th>
                  <th className="text-right p-2 text-sm font-medium text-[#9CA3AF]">Deposits</th>
                  <th className="text-right p-2 text-sm font-medium text-[#9CA3AF]">Withdrawals</th>
                  <th className="text-right p-2 text-sm font-medium text-[#9CA3AF]">Trading P&L</th>
                  <th className="text-right p-2 text-sm font-medium text-[#9CA3AF]">Commissions</th>
                  <th className="text-right p-2 text-sm font-medium text-[#9CA3AF]">Swaps</th>
                  <th className="text-right p-2 text-sm font-medium text-[#9CA3AF]">Closing</th>
                </tr>
              </thead>
              <tbody>
                {generatedStatement.dailyBalances.map((day, idx) => (
                  <tr key={idx} className="border-t border-[#383A42] hover:bg-[#252830] transition-colors">
                    <td className="p-2 text-white text-sm">{day.date}</td>
                    <td className="p-2 text-right text-white text-sm">{formatCurrency(day.openingBalance)}</td>
                    <td className="p-2 text-right text-[#10B981] text-sm">{formatCurrency(day.deposits)}</td>
                    <td className="p-2 text-right text-[#EF4444] text-sm">{formatCurrency(day.withdrawals)}</td>
                    <td className={`p-2 text-right text-sm ${day.tradingPnl >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                      {formatCurrency(day.tradingPnl)}
                    </td>
                    <td className="p-2 text-right text-[#9CA3AF] text-sm">{formatCurrency(day.commissions)}</td>
                    <td className={`p-2 text-right text-sm ${day.swaps >= 0 ? 'text-[#10B981]' : 'text-[#EF4444]'}`}>
                      {formatCurrency(day.swaps)}
                    </td>
                    <td className="p-2 text-right text-white font-semibold text-sm">{formatCurrency(day.closingBalance)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    );
  };

  // ============================================
  // Render Template Gallery
  // ============================================

  const renderTemplateGallery = () => {
    return (
      <div className="mb-6">
        <div className="flex items-center gap-2 mb-4">
          <BarChart3 size={20} className="text-[#F5C542]" />
          <h2 className="text-lg font-semibold text-white">Quick Templates</h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {templates.map((template) => (
            <div key={template.id} className="bg-[#1E2026] border border-[#383A42] rounded p-4 hover:border-[#F5C542] transition-colors">
              <h3 className="text-white font-semibold mb-2">{template.name}</h3>
              <p className="text-sm text-[#9CA3AF] mb-3">{template.description}</p>
              <div className="flex items-center justify-between">
                <span className="text-xs text-[#6B7280]">
                  {template.defaultDays > 0 ? `${template.defaultDays} days` : 'Custom'}
                </span>
                <button
                  onClick={() => handleQuickGenerate(template)}
                  className="px-3 py-1 bg-[#F5C542] text-[#121316] rounded text-xs font-semibold hover:bg-[#F5D563] transition-colors"
                >
                  Generate
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  // ============================================
  // Render Statement History Table
  // ============================================

  const renderHistoryTable = () => {
    return (
      <div className="bg-[#1E2026] border border-[#383A42] rounded mb-6">
        <div className="p-4 border-b border-[#383A42]">
          <div className="flex items-center gap-2">
            <Calendar size={20} className="text-[#F5C542]" />
            <h2 className="text-lg font-semibold text-white">Statement History ({history.length})</h2>
          </div>
        </div>

        <div className="overflow-x-auto max-h-[400px] overflow-y-auto">
          <table className="w-full">
            <thead className="sticky top-0 bg-[#121316]">
              <tr>
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Client</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Format</th>
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Date Range</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Generated</th>
                <th className="text-right p-3 text-sm font-medium text-[#9CA3AF]">File Size</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Download</th>
              </tr>
            </thead>
            <tbody>
              {history.map((item) => (
                <tr key={item.id} className="border-t border-[#383A42] hover:bg-[#252830] transition-colors">
                  <td className="p-3 text-white">{item.clientName}</td>
                  <td className="p-3 text-center">
                    <span
                      className="px-2 py-1 rounded text-xs font-medium uppercase flex items-center justify-center gap-1"
                      style={{
                        backgroundColor: `${getFormatColor(item.format)}20`,
                        color: getFormatColor(item.format)
                      }}
                    >
                      {getFormatIcon(item.format)}
                      {item.format}
                    </span>
                  </td>
                  <td className="p-3 text-white text-sm">{item.fromDate} → {item.toDate}</td>
                  <td className="p-3 text-center text-[#9CA3AF] text-sm">
                    {new Date(item.generatedAt).toLocaleDateString()}
                  </td>
                  <td className="p-3 text-right text-[#9CA3AF] text-sm">{formatFileSize(item.fileSize)}</td>
                  <td className="p-3 text-center">
                    <button className="px-3 py-1 bg-[#383A42] text-white rounded text-xs hover:bg-[#4A4C54] transition-colors flex items-center gap-1 mx-auto">
                      <Download size={14} />
                      Download
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  };

  // ============================================
  // Render Scheduled Statements Management
  // ============================================

  const renderScheduledStatements = () => {
    return (
      <div className="bg-[#1E2026] border border-[#383A42] rounded mb-6">
        <div className="p-4 border-b border-[#383A42]">
          <div className="flex items-center gap-2">
            <Clock size={20} className="text-[#F5C542]" />
            <h2 className="text-lg font-semibold text-white">Scheduled Statements ({scheduled.length})</h2>
          </div>
        </div>

        <div className="overflow-x-auto max-h-[400px] overflow-y-auto">
          <table className="w-full">
            <thead className="sticky top-0 bg-[#121316]">
              <tr>
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Client</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Frequency</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Format</th>
                <th className="text-left p-3 text-sm font-medium text-[#9CA3AF]">Email</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Next Run</th>
                <th className="text-center p-3 text-sm font-medium text-[#9CA3AF]">Status</th>
              </tr>
            </thead>
            <tbody>
              {scheduled.map((sch) => (
                <tr key={sch.id} className="border-t border-[#383A42] hover:bg-[#252830] transition-colors">
                  <td className="p-3 text-white">{sch.clientName}</td>
                  <td className="p-3 text-center">
                    <span className="px-2 py-1 bg-[#383A42] text-white rounded text-xs capitalize">
                      {sch.frequency}
                    </span>
                  </td>
                  <td className="p-3 text-center">
                    <span
                      className="px-2 py-1 rounded text-xs font-medium uppercase"
                      style={{
                        backgroundColor: `${getFormatColor(sch.format)}20`,
                        color: getFormatColor(sch.format)
                      }}
                    >
                      {sch.format}
                    </span>
                  </td>
                  <td className="p-3 text-[#9CA3AF] text-sm flex items-center gap-1">
                    <Mail size={14} />
                    {sch.emailTo}
                  </td>
                  <td className="p-3 text-center text-white text-sm">
                    {new Date(sch.nextRun).toLocaleDateString()}
                  </td>
                  <td className="p-3 text-center">
                    <button
                      onClick={() => handleToggleSchedule(sch.id, sch.isActive)}
                      className={`px-3 py-1 rounded text-xs font-semibold transition-colors ${
                        sch.isActive
                          ? 'bg-[#10B981] text-white hover:bg-[#059669]'
                          : 'bg-[#6B7280] text-white hover:bg-[#4B5563]'
                      }`}
                    >
                      {sch.isActive ? 'Active' : 'Inactive'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  };

  // ============================================
  // Main Render
  // ============================================

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white mb-2">Account Statement & Trade Export</h1>
        <p className="text-[#9CA3AF]">Generate, export, and schedule client account statements</p>
      </div>

      {renderStatsCards()}
      {renderTemplateGallery()}
      {renderGeneratorPanel()}
      {renderGeneratedStatement()}
      {renderHistoryTable()}
      {renderScheduledStatements()}
    </div>
  );
};

export default AccountStatement;
