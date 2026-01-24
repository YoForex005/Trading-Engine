/**
 * Data Management Dashboard for Admin Panel
 * Comprehensive admin controls for tick data management
 * Features: Storage overview, data operations, monitoring, and configuration
 */

import React, { useState, useEffect } from 'react';
import {
  Database,
  Download,
  Upload,
  RefreshCw,
  HardDrive,
  AlertCircle,
  CheckCircle,
  Trash2,
  Archive,
  TrendingUp,
  Clock,
  FileText,
  Settings
} from 'lucide-react';

const API_BASE = 'http://localhost:8080';

type BackfillTask = {
  id: string;
  symbol: string;
  dateRange: { from: string; to: string };
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress: number;
  totalRecords: number;
  processedRecords: number;
  startTime?: number;
  endTime?: number;
  error?: string;
};

type StorageMetrics = {
  totalSize: number;
  ticksCount: number;
  symbolsCount: number;
  oldestData: string;
  newestData: string;
  avgTicksPerDay: number;
};

interface StorageStats {
  totalTicks: number;
  totalSizeBytes: number;
  totalSizeMB: number;
  symbolCount: number;
  dateRangeStart?: string;
  dateRangeEnd: string;
  symbolStats: Record<string, SymbolStats>;
  missingDataGaps?: DataGap[];
  lastUpdated: string;
}

interface SymbolStats {
  symbol: string;
  tickCount: number;
  sizeBytes: number;
  sizeMB: number;
  availableDates: string[];
  firstDate?: string;
  lastDate?: string;
}

interface DataGap {
  symbol: string;
  startDate: string;
  endDate: string;
  daysMissing: number;
}

interface MonitoringMetrics {
  tickIngestionRate: number;
  storageGrowthMBPerHour: number;
  failedWrites: number;
  lastTickTimestamp: Record<string, string>;
  dataQuality: {
    ticksWithZeroSpread: number;
    ticksWithInvalidPrice: number;
    qualityScore: number;
  };
}

interface ConfigSettings {
  retentionDays: number;
  autoDownload: boolean;
  autoDownloadSymbols?: string[];
  compressionEnabled: boolean;
  backupSchedule: string;
  backupPath: string;
}

export const DataManagement: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'overview' | 'management' | 'monitoring' | 'config'>('overview');
  const [backfillTasks, setBackfillTasks] = useState<BackfillTask[]>([]);
  const [storageMetrics, setStorageMetrics] = useState<StorageMetrics | null>(null);
  const [stats, setStats] = useState<StorageStats | null>(null);
  const [monitoring, setMonitoring] = useState<MonitoringMetrics | null>(null);
  const [config, setConfig] = useState<ConfigSettings>({
    retentionDays: 180,
    autoDownload: false,
    compressionEnabled: true,
    backupSchedule: '0 2 * * *',
    backupPath: 'backups'
  });
  const [selectedSymbol, setSelectedSymbol] = useState('');
  const [dateRange, setDateRange] = useState({
    from: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    to: new Date().toISOString().split('T')[0]
  });
  const [isLoading, setIsLoading] = useState(false);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  // Import state
  const [importSymbol, setImportSymbol] = useState('');
  const [importFile, setImportFile] = useState<File | null>(null);

  // Cleanup state
  const [cleanupSymbol, setCleanupSymbol] = useState('');
  const [cleanupDays, setCleanupDays] = useState(180);
  const [confirmCleanup, setConfirmCleanup] = useState(false);

  // Compress state
  const [compressSymbol, setCompressSymbol] = useState('');
  const [compressDays, setCompressDays] = useState(90);

  // Backup state
  const [backupSymbols, setBackupSymbols] = useState<string[]>([]);

  useEffect(() => {
    loadStorageMetrics();
    loadBackfillTasks();
  }, []);

  const loadStorageMetrics = async () => {
    try {
      const response = await fetch('/api/admin/data/metrics');
      const data = await response.json();
      setStorageMetrics(data);
    } catch (error) {
      console.error('Failed to load storage metrics:', error);
    }
  };

  const loadBackfillTasks = async () => {
    try {
      const response = await fetch('/api/admin/data/backfill/tasks');
      const data = await response.json();
      setBackfillTasks(data);
    } catch (error) {
      console.error('Failed to load backfill tasks:', error);
    }
  };

  const handleBackfill = async () => {
    if (!selectedSymbol) return;

    setIsLoading(true);
    try {
      const response = await fetch('/api/admin/data/backfill', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          symbol: selectedSymbol,
          dateRange
        })
      });

      if (response.ok) {
        loadBackfillTasks();
      }
    } catch (error) {
      console.error('Failed to start backfill:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleExport = async () => {
    if (!selectedSymbol) return;

    try {
      const params = new URLSearchParams({
        symbol: selectedSymbol,
        from: dateRange.from,
        to: dateRange.to
      });

      const response = await fetch(`/api/admin/data/export?${params}`);
      const blob = await response.blob();

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${selectedSymbol}_${dateRange.from}_${dateRange.to}.csv`;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export data:', error);
    }
  };

  const handleImport = async (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('symbol', selectedSymbol);

    try {
      const response = await fetch('/api/admin/data/import', {
        method: 'POST',
        body: formData
      });

      if (response.ok) {
        loadStorageMetrics();
        alert('Data imported successfully');
      }
    } catch (error) {
      console.error('Failed to import data:', error);
    }
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Database className="w-6 h-6 text-blue-500" />
          <h1 className="text-2xl font-bold">Data Management</h1>
        </div>

        <button
          onClick={loadStorageMetrics}
          className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded flex items-center gap-2 transition-colors"
        >
          <RefreshCw className="w-4 h-4" />
          Refresh
        </button>
      </div>

      {/* Storage Metrics */}
      {storageMetrics && (
        <div className="grid grid-cols-4 gap-4">
          <div className="p-4 bg-gray-800 rounded-lg border border-gray-700">
            <div className="flex items-center gap-2 mb-2">
              <HardDrive className="w-5 h-5 text-blue-400" />
              <span className="text-sm text-gray-400">Total Size</span>
            </div>
            <div className="text-2xl font-bold">{formatBytes(storageMetrics.totalSize)}</div>
          </div>

          <div className="p-4 bg-gray-800 rounded-lg border border-gray-700">
            <div className="text-sm text-gray-400 mb-2">Total Ticks</div>
            <div className="text-2xl font-bold">{storageMetrics.ticksCount.toLocaleString()}</div>
          </div>

          <div className="p-4 bg-gray-800 rounded-lg border border-gray-700">
            <div className="text-sm text-gray-400 mb-2">Symbols</div>
            <div className="text-2xl font-bold">{storageMetrics.symbolsCount}</div>
          </div>

          <div className="p-4 bg-gray-800 rounded-lg border border-gray-700">
            <div className="text-sm text-gray-400 mb-2">Avg Ticks/Day</div>
            <div className="text-2xl font-bold">{storageMetrics.avgTicksPerDay.toLocaleString()}</div>
          </div>
        </div>
      )}

      {/* Backfill Operations */}
      <div className="bg-gray-800 rounded-lg border border-gray-700 p-6">
        <h2 className="text-lg font-semibold mb-4">Backfill Operations</h2>

        <div className="grid grid-cols-3 gap-4 mb-4">
          <div>
            <label className="block text-sm mb-2">Symbol</label>
            <input
              type="text"
              value={selectedSymbol}
              onChange={(e) => setSelectedSymbol(e.target.value)}
              placeholder="e.g., EURUSD"
              className="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded focus:outline-none focus:border-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm mb-2">From Date</label>
            <input
              type="date"
              value={dateRange.from}
              onChange={(e) => setDateRange({ ...dateRange, from: e.target.value })}
              className="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded focus:outline-none focus:border-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm mb-2">To Date</label>
            <input
              type="date"
              value={dateRange.to}
              onChange={(e) => setDateRange({ ...dateRange, to: e.target.value })}
              className="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded focus:outline-none focus:border-blue-500"
            />
          </div>
        </div>

        <div className="flex gap-3">
          <button
            onClick={handleBackfill}
            disabled={!selectedSymbol || isLoading}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed rounded font-medium flex items-center gap-2 transition-colors"
          >
            <Download className="w-4 h-4" />
            Start Backfill
          </button>

          <button
            onClick={handleExport}
            disabled={!selectedSymbol}
            className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-700 disabled:cursor-not-allowed rounded font-medium flex items-center gap-2 transition-colors"
          >
            <Upload className="w-4 h-4" />
            Export Data
          </button>

          <label className="px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-medium flex items-center gap-2 cursor-pointer transition-colors">
            <Upload className="w-4 h-4" />
            Import Data
            <input
              type="file"
              accept=".csv,.json"
              onChange={(e) => e.target.files?.[0] && handleImport(e.target.files[0])}
              className="hidden"
            />
          </label>
        </div>
      </div>

      {/* Backfill Tasks */}
      <div className="bg-gray-800 rounded-lg border border-gray-700 p-6">
        <h2 className="text-lg font-semibold mb-4">Active Backfill Tasks</h2>

        {backfillTasks.length === 0 ? (
          <div className="text-center text-gray-400 py-8">
            No active backfill tasks
          </div>
        ) : (
          <div className="space-y-3">
            {backfillTasks.map((task) => (
              <div
                key={task.id}
                className="p-4 bg-gray-900 rounded border border-gray-700"
              >
                <div className="flex items-center justify-between mb-2">
                  <div>
                    <div className="font-medium">{task.symbol}</div>
                    <div className="text-sm text-gray-400">
                      {task.dateRange.from} to {task.dateRange.to}
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    {task.status === 'completed' ? (
                      <CheckCircle className="w-5 h-5 text-green-500" />
                    ) : task.status === 'failed' ? (
                      <AlertCircle className="w-5 h-5 text-red-500" />
                    ) : (
                      <RefreshCw className="w-5 h-5 text-blue-500 animate-spin" />
                    )}
                    <span className="text-sm font-medium uppercase">{task.status}</span>
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-400">Progress</span>
                    <span>
                      {Math.round(task.progress)}% ({task.processedRecords.toLocaleString()} / {task.totalRecords.toLocaleString()})
                    </span>
                  </div>

                  <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
                    <div
                      className={`h-full transition-all ${
                        task.status === 'failed'
                          ? 'bg-red-500'
                          : task.status === 'completed'
                          ? 'bg-green-500'
                          : 'bg-blue-500'
                      }`}
                      style={{ width: `${task.progress}%` }}
                    />
                  </div>

                  {task.error && (
                    <div className="text-sm text-red-400">
                      Error: {task.error}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
