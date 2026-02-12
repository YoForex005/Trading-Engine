'use client';

import { useState, useEffect, useMemo } from 'react';
import {
  Database, HardDrive, Clock, Calendar, CheckCircle, XCircle,
  RefreshCw, Trash2, Play, Plus, X, AlertTriangle, Download,
  ChevronDown, ChevronUp, Settings
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// Types matching backend
type BackupType = 'full' | 'incremental' | 'differential' | 'config_only' | 'database_only';
type BackupStatus = 'pending' | 'in_progress' | 'completed' | 'failed' | 'expired';
type BackupFrequency = 'hourly' | 'daily' | 'weekly' | 'monthly';
type RestoreStatus = 'pending' | 'in_progress' | 'completed' | 'failed';

interface Backup {
  id: string;
  name: string;
  type: BackupType;
  status: BackupStatus;
  size: number; // bytes
  createdAt: string;
  completedAt?: string;
  createdBy: string;
  storagePath: string;
  retention: number; // days
  components: string[];
  error?: string;
  duration: number; // seconds
}

interface BackupSchedule {
  id: string;
  name: string;
  frequency: BackupFrequency;
  nextRun: string;
  lastRun?: string;
  type: BackupType;
  retentionDays: number;
  isActive: boolean;
  components: string[];
  createdAt: string;
  updatedAt: string;
}

interface RestorePoint {
  id: string;
  backupId: string;
  timestamp: string;
  description: string;
  status: RestoreStatus;
  restoredBy: string;
  restoredAt: string;
  duration: number; // seconds
  error?: string;
}

interface BackupComponent {
  name: string;
  size: number; // bytes
  itemCount: number;
  lastBackup?: string;
  enabled: boolean;
}

interface BackupStats {
  totalBackups: number;
  storageUsed: number; // bytes
  lastSuccessful: string;
  nextScheduled: string;
  failedLast24h: number;
  completedLast24h: number;
  averageSize: number;
  averageDuration: number; // seconds
}

type SortField = 'name' | 'type' | 'status' | 'size' | 'createdAt' | 'duration';
type SortDirection = 'asc' | 'desc';

export default function BackupManager() {
  const [backups, setBackups] = useState<Backup[]>([]);
  const [schedules, setSchedules] = useState<BackupSchedule[]>([]);
  const [components, setComponents] = useState<BackupComponent[]>([]);
  const [restorePoints, setRestorePoints] = useState<RestorePoint[]>([]);
  const [stats, setStats] = useState<BackupStats | null>(null);
  const [loading, setLoading] = useState(true);

  // Modals
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showRestoreModal, setShowRestoreModal] = useState<Backup | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);

  // Create backup form
  const [backupType, setBackupType] = useState<BackupType>('full');
  const [selectedComponents, setSelectedComponents] = useState<string[]>([]);
  const [backupName, setBackupName] = useState('');

  // Restore progress
  const [restoring, setRestoring] = useState(false);

  // Sorting
  const [sortField, setSortField] = useState<SortField>('createdAt');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  useEffect(() => {
    fetchAllData();
  }, []);

  const fetchAllData = async () => {
    setLoading(true);
    const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
    if (!token) {
      setLoading(false);
      return;
    }

    const headers = {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    };

    try {
      const [backupsRes, schedulesRes, componentsRes, statsRes] = await Promise.all([
        fetch(API_CONFIG.BACKUP_LIST, { headers }),
        fetch(API_CONFIG.BACKUP_SCHEDULES, { headers }),
        fetch(API_CONFIG.BACKUP_COMPONENTS, { headers }),
        fetch(API_CONFIG.BACKUP_STATS, { headers })
      ]);

      if (backupsRes.ok) {
        const data = await backupsRes.json();
        setBackups(data.backups || []);
      }

      if (schedulesRes.ok) {
        const data = await schedulesRes.json();
        setSchedules(data.schedules || []);
      }

      if (componentsRes.ok) {
        const data = await componentsRes.json();
        setComponents(data.components || []);
      }

      if (statsRes.ok) {
        setStats(await statsRes.json());
      }
    } catch (error) {
      console.error('Failed to fetch backup data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateBackup = async () => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
    if (!token) return;

    const componentsToBackup = selectedComponents.length > 0 ? selectedComponents :
      backupType === 'full' ? components.map(c => c.name) :
      backupType === 'config_only' ? ['config'] :
      backupType === 'database_only' ? ['database'] :
      ['database', 'config'];

    try {
      const res = await fetch(API_CONFIG.BACKUP_CREATE, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          type: backupType,
          components: componentsToBackup,
          name: backupName || `${backupType}-backup-${new Date().toISOString().split('T')[0]}`
        })
      });

      if (res.ok) {
        fetchAllData();
        setShowCreateModal(false);
        setBackupName('');
        setSelectedComponents([]);
      }
    } catch (error) {
      console.error('Failed to create backup:', error);
    }
  };

  const handleDeleteBackup = async (backupId: string) => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
    if (!token) return;

    try {
      const res = await fetch(API_CONFIG.BACKUP_DELETE(backupId), {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (res.ok) {
        fetchAllData();
        setShowDeleteConfirm(null);
      }
    } catch (error) {
      console.error('Failed to delete backup:', error);
    }
  };

  const handleRestoreBackup = async (backupId: string) => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
    if (!token) return;

    setRestoring(true);
    try {
      const res = await fetch(API_CONFIG.BACKUP_RESTORE(backupId), {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (res.ok) {
        const restorePoint = await res.json();
        setRestorePoints([restorePoint, ...restorePoints]);
        setShowRestoreModal(null);
        fetchAllData();
      }
    } catch (error) {
      console.error('Failed to restore backup:', error);
    } finally {
      setRestoring(false);
    }
  };

  const handleToggleSchedule = async (scheduleId: string, isActive: boolean) => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
    if (!token) return;

    try {
      const res = await fetch(API_CONFIG.BACKUP_SCHEDULE_UPDATE(scheduleId), {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ isActive })
      });

      if (res.ok) {
        setSchedules(schedules.map(s => s.id === scheduleId ? { ...s, isActive } : s));
      }
    } catch (error) {
      console.error('Failed to toggle schedule:', error);
    }
  };

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const sortedBackups = useMemo(() => {
    const sorted = [...backups].sort((a, b) => {
      let aVal: any = a[sortField];
      let bVal: any = b[sortField];

      if (sortField === 'createdAt') {
        aVal = new Date(aVal).getTime();
        bVal = new Date(bVal).getTime();
      }

      if (sortDirection === 'asc') {
        return aVal > bVal ? 1 : -1;
      } else {
        return aVal < bVal ? 1 : -1;
      }
    });
    return sorted;
  }, [backups, sortField, sortDirection]);

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  const formatDuration = (seconds: number) => {
    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${minutes}m ${secs}s`;
  };

  const getTypeBadge = (type: BackupType) => {
    const colors = {
      full: 'bg-[#3B82F6]/20 text-[#3B82F6] border-[#3B82F6]/30',
      incremental: 'bg-[#10B981]/20 text-[#10B981] border-[#10B981]/30',
      differential: 'bg-[#F59E0B]/20 text-[#F59E0B] border-[#F59E0B]/30',
      config_only: 'bg-[#8B5CF6]/20 text-[#8B5CF6] border-[#8B5CF6]/30',
      database_only: 'bg-[#EC4899]/20 text-[#EC4899] border-[#EC4899]/30'
    };
    return colors[type] || colors.full;
  };

  const getStatusBadge = (status: BackupStatus) => {
    const colors = {
      completed: 'bg-[#10B981]/20 text-[#10B981]',
      failed: 'bg-[#EF4444]/20 text-[#EF4444]',
      in_progress: 'bg-[#3B82F6]/20 text-[#3B82F6]',
      pending: 'bg-[#F59E0B]/20 text-[#F59E0B]',
      expired: 'bg-[#666]/20 text-[#666]'
    };
    return colors[status] || colors.pending;
  };

  const totalComponentSize = components.reduce((sum, c) => sum + c.size, 0);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full bg-[#121316]">
        <div className="text-[#666] text-sm">Loading backup data...</div>
      </div>
    );
  }

  return (
    <div className="h-full bg-[#121316] overflow-auto">
      <div className="p-4 space-y-4">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Database className="text-[#F5C542]" size={20} />
            <h1 className="text-[#F5C542] font-bold text-lg">Backup & Recovery Management</h1>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="bg-[#2980B9] hover:bg-[#3498DB] text-white text-xs px-3 py-1.5 rounded flex items-center gap-1"
          >
            <Plus size={14} />
            Create Backup
          </button>
        </div>

        {/* Stats Overview Cards */}
        {stats && (
          <div className="grid grid-cols-5 gap-3">
            <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-[#888] text-xs">Total Backups</span>
                <Database size={14} className="text-[#3B82F6]" />
              </div>
              <div className="text-white text-2xl font-bold">{stats.totalBackups}</div>
              <div className="text-[#666] text-xs mt-1">{stats.completedLast24h} today</div>
            </div>

            <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-[#888] text-xs">Storage Used</span>
                <HardDrive size={14} className="text-[#F59E0B]" />
              </div>
              <div className="text-white text-2xl font-bold">{formatBytes(stats.storageUsed)}</div>
              <div className="text-[#666] text-xs mt-1">Avg: {formatBytes(stats.averageSize)}</div>
            </div>

            <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-[#888] text-xs">Last Successful</span>
                <CheckCircle size={14} className="text-[#10B981]" />
              </div>
              <div className="text-white text-sm font-bold">{new Date(stats.lastSuccessful).toLocaleDateString()}</div>
              <div className="text-[#666] text-xs mt-1">{new Date(stats.lastSuccessful).toLocaleTimeString()}</div>
            </div>

            <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-[#888] text-xs">Next Scheduled</span>
                <Calendar size={14} className="text-[#8B5CF6]" />
              </div>
              <div className="text-white text-sm font-bold">{new Date(stats.nextScheduled).toLocaleDateString()}</div>
              <div className="text-[#666] text-xs mt-1">{new Date(stats.nextScheduled).toLocaleTimeString()}</div>
            </div>

            <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-[#888] text-xs">Failed (24h)</span>
                <XCircle size={14} className="text-[#EF4444]" />
              </div>
              <div className="text-white text-2xl font-bold">{stats.failedLast24h}</div>
              <div className="text-[#666] text-xs mt-1">Avg duration: {formatDuration(stats.averageDuration)}</div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-3 gap-4">
          {/* Component Overview Cards */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded col-span-2">
            <div className="border-b border-[#383A42] p-3">
              <h2 className="text-white font-bold text-sm flex items-center gap-2">
                <Settings size={14} className="text-[#10B981]" />
                System Components ({components.length})
              </h2>
            </div>
            <div className="p-3">
              <div className="grid grid-cols-4 gap-2">
                {components.map(comp => (
                  <div key={comp.name} className="bg-[#121316] border border-[#383A42] rounded p-2">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-white text-xs font-semibold capitalize">{comp.name}</span>
                      {comp.enabled && <div className="w-2 h-2 rounded-full bg-[#10B981]"></div>}
                    </div>
                    <div className="text-[#F5C542] text-sm font-bold mb-0.5">{formatBytes(comp.size)}</div>
                    <div className="text-[#666] text-xs">{comp.itemCount.toLocaleString()} items</div>
                    {comp.lastBackup && (
                      <div className="text-[#888] text-xs mt-1">
                        Last: {new Date(comp.lastBackup).toLocaleDateString()}
                      </div>
                    )}
                  </div>
                ))}
              </div>

              {/* Storage Usage Bar Chart */}
              <div className="mt-4">
                <div className="text-[#888] text-xs mb-2">Storage Distribution</div>
                <div className="flex gap-1 h-8">
                  {components.map((comp, idx) => {
                    const percentage = (comp.size / totalComponentSize) * 100;
                    const colors = ['#3B82F6', '#10B981', '#F59E0B', '#8B5CF6', '#EC4899', '#EF4444', '#14B8A6', '#F97316'];
                    return (
                      <div
                        key={comp.name}
                        className="relative group"
                        style={{
                          width: `${percentage}%`,
                          backgroundColor: colors[idx % colors.length],
                          minWidth: percentage > 2 ? 'auto' : '4px'
                        }}
                        title={`${comp.name}: ${formatBytes(comp.size)} (${percentage.toFixed(1)}%)`}
                      >
                        {percentage > 10 && (
                          <div className="absolute inset-0 flex items-center justify-center text-white text-xs font-semibold">
                            {percentage.toFixed(0)}%
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          </div>

          {/* Schedule Management Panel */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded">
            <div className="border-b border-[#383A42] p-3">
              <h2 className="text-white font-bold text-sm flex items-center gap-2">
                <Clock size={14} className="text-[#8B5CF6]" />
                Backup Schedules ({schedules.filter(s => s.isActive).length}/{schedules.length} active)
              </h2>
            </div>
            <div className="p-3 max-h-80 overflow-y-auto space-y-2">
              {schedules.map(schedule => (
                <div key={schedule.id} className="bg-[#121316] border border-[#383A42] rounded p-2">
                  <div className="flex items-start justify-between mb-1">
                    <div className="flex-1">
                      <div className="text-white text-xs font-semibold mb-0.5">{schedule.name}</div>
                      <div className="text-[#888] text-xs capitalize">{schedule.frequency} • {schedule.type.replace('_', ' ')}</div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={schedule.isActive}
                        onChange={e => handleToggleSchedule(schedule.id, e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-7 h-4 bg-[#383A42] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-3 after:w-3 after:transition-all peer-checked:bg-[#10B981]"></div>
                    </label>
                  </div>
                  <div className="flex items-center gap-2 text-xs mt-2">
                    <span className="text-[#666]">Next:</span>
                    <span className="text-[#3B82F6]">{new Date(schedule.nextRun).toLocaleString()}</span>
                  </div>
                  {schedule.lastRun && (
                    <div className="flex items-center gap-2 text-xs mt-1">
                      <span className="text-[#666]">Last:</span>
                      <span className="text-[#888]">{new Date(schedule.lastRun).toLocaleString()}</span>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Backup List Table */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded">
          <div className="border-b border-[#383A42] p-3">
            <h2 className="text-white font-bold text-sm flex items-center gap-2">
              <Database size={14} className="text-[#3B82F6]" />
              Backup History ({sortedBackups.length})
            </h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="bg-[#121316] border-b border-[#383A42]">
                  <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('name')}>
                    <div className="flex items-center gap-1">
                      Name {sortField === 'name' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('type')}>
                    <div className="flex items-center gap-1">
                      Type {sortField === 'type' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('status')}>
                    <div className="flex items-center gap-1">
                      Status {sortField === 'status' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('size')}>
                    <div className="flex items-center gap-1">
                      Size {sortField === 'size' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('createdAt')}>
                    <div className="flex items-center gap-1">
                      Created {sortField === 'createdAt' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th className="text-left text-[#888] font-semibold p-2 cursor-pointer hover:bg-[#1E2026]" onClick={() => handleSort('duration')}>
                    <div className="flex items-center gap-1">
                      Duration {sortField === 'duration' && (sortDirection === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
                    </div>
                  </th>
                  <th className="text-left text-[#888] font-semibold p-2">Actions</th>
                </tr>
              </thead>
              <tbody>
                {sortedBackups.map(backup => (
                  <tr key={backup.id} className="border-b border-[#383A42] hover:bg-[#1E2026]">
                    <td className="p-2 text-white">{backup.name}</td>
                    <td className="p-2">
                      <span className={`px-2 py-0.5 rounded text-xs border ${getTypeBadge(backup.type)}`}>
                        {backup.type.replace('_', ' ')}
                      </span>
                    </td>
                    <td className="p-2">
                      <span className={`px-2 py-0.5 rounded text-xs ${getStatusBadge(backup.status)}`}>
                        {backup.status}
                      </span>
                    </td>
                    <td className="p-2 text-[#CCC]">{formatBytes(backup.size)}</td>
                    <td className="p-2 text-[#888]">{new Date(backup.createdAt).toLocaleString()}</td>
                    <td className="p-2 text-[#888]">{formatDuration(backup.duration)}</td>
                    <td className="p-2">
                      <div className="flex items-center gap-2">
                        {backup.status === 'completed' && (
                          <button
                            onClick={() => setShowRestoreModal(backup)}
                            className="text-[#10B981] hover:text-[#22C55E]"
                            title="Restore"
                          >
                            <RefreshCw size={14} />
                          </button>
                        )}
                        <button
                          onClick={() => setShowDeleteConfirm(backup.id)}
                          className="text-[#EF4444] hover:text-[#F87171]"
                          title="Delete"
                        >
                          <Trash2 size={14} />
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

      {/* Create Backup Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4 w-[500px]">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-white font-bold text-sm">Create Backup</h3>
              <button onClick={() => setShowCreateModal(false)} className="text-[#666] hover:text-white">
                <X size={16} />
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="text-[#888] text-xs mb-2 block">Backup Name</label>
                <input
                  type="text"
                  value={backupName}
                  onChange={e => setBackupName(e.target.value)}
                  placeholder="my-backup-2026-02-11"
                  className="w-full bg-[#121316] border border-[#383A42] text-white text-xs px-3 py-2 rounded"
                />
              </div>

              <div>
                <label className="text-[#888] text-xs mb-2 block">Backup Type</label>
                <select
                  value={backupType}
                  onChange={e => setBackupType(e.target.value as BackupType)}
                  className="w-full bg-[#121316] border border-[#383A42] text-white text-xs px-3 py-2 rounded"
                >
                  <option value="full">Full Backup (All Components)</option>
                  <option value="incremental">Incremental Backup (Changes Only)</option>
                  <option value="differential">Differential Backup (Since Last Full)</option>
                  <option value="config_only">Config Only</option>
                  <option value="database_only">Database Only</option>
                </select>
              </div>

              <div>
                <label className="text-[#888] text-xs mb-2 block">Components to Include</label>
                <div className="grid grid-cols-2 gap-2">
                  {components.map(comp => (
                    <label key={comp.name} className="flex items-center gap-2 text-xs text-[#CCC] cursor-pointer">
                      <input
                        type="checkbox"
                        checked={selectedComponents.includes(comp.name)}
                        onChange={e => {
                          if (e.target.checked) {
                            setSelectedComponents([...selectedComponents, comp.name]);
                          } else {
                            setSelectedComponents(selectedComponents.filter(c => c !== comp.name));
                          }
                        }}
                        className="rounded"
                      />
                      <span className="capitalize">{comp.name}</span>
                      <span className="text-[#666]">({formatBytes(comp.size)})</span>
                    </label>
                  ))}
                </div>
                <div className="text-[#666] text-xs mt-2">
                  {selectedComponents.length === 0 ? 'All components will be included based on backup type' : `Selected: ${selectedComponents.length} component(s)`}
                </div>
              </div>
            </div>

            <div className="flex gap-2 mt-4">
              <button
                onClick={() => setShowCreateModal(false)}
                className="flex-1 bg-[#383A42] hover:bg-[#4A4C54] text-white text-xs py-2 rounded"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateBackup}
                className="flex-1 bg-[#2980B9] hover:bg-[#3498DB] text-white text-xs py-2 rounded"
              >
                Create Backup
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Restore Confirmation Modal */}
      {showRestoreModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4 w-[450px]">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-white font-bold text-sm">Restore Backup</h3>
              <button onClick={() => setShowRestoreModal(null)} className="text-[#666] hover:text-white">
                <X size={16} />
              </button>
            </div>

            <div className="mb-4">
              <div className="flex items-center gap-2 p-3 bg-[#F59E0B]/10 border border-[#F59E0B]/30 rounded mb-3">
                <AlertTriangle size={16} className="text-[#F59E0B]" />
                <div className="text-[#F59E0B] text-xs">
                  <div className="font-semibold mb-1">Warning: This action will restore system data</div>
                  <div>This will overwrite current data with the backup. Make sure you have a recent backup before proceeding.</div>
                </div>
              </div>

              <div className="bg-[#121316] border border-[#383A42] rounded p-3 space-y-2 text-xs">
                <div className="flex justify-between">
                  <span className="text-[#888]">Backup:</span>
                  <span className="text-white">{showRestoreModal.name}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#888]">Type:</span>
                  <span className="text-white capitalize">{showRestoreModal.type.replace('_', ' ')}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#888]">Size:</span>
                  <span className="text-white">{formatBytes(showRestoreModal.size)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#888]">Created:</span>
                  <span className="text-white">{new Date(showRestoreModal.createdAt).toLocaleString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#888]">Components:</span>
                  <span className="text-white">{showRestoreModal.components.join(', ')}</span>
                </div>
              </div>
            </div>

            {restoring && (
              <div className="mb-4 p-3 bg-[#3B82F6]/10 border border-[#3B82F6]/30 rounded">
                <div className="flex items-center gap-2 text-[#3B82F6] text-xs">
                  <RefreshCw size={14} className="animate-spin" />
                  <span>Restoring backup... This may take several minutes.</span>
                </div>
              </div>
            )}

            <div className="flex gap-2">
              <button
                onClick={() => setShowRestoreModal(null)}
                disabled={restoring}
                className="flex-1 bg-[#383A42] hover:bg-[#4A4C54] text-white text-xs py-2 rounded disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={() => handleRestoreBackup(showRestoreModal.id)}
                disabled={restoring}
                className="flex-1 bg-[#F59E0B] hover:bg-[#FBBF24] text-white text-xs py-2 rounded disabled:opacity-50"
              >
                {restoring ? 'Restoring...' : 'Confirm Restore'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4 w-96">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-white font-bold text-sm">Delete Backup</h3>
              <button onClick={() => setShowDeleteConfirm(null)} className="text-[#666] hover:text-white">
                <X size={16} />
              </button>
            </div>
            <div className="text-[#CCC] text-xs mb-4">
              Are you sure you want to delete this backup? This action cannot be undone.
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setShowDeleteConfirm(null)}
                className="flex-1 bg-[#383A42] hover:bg-[#4A4C54] text-white text-xs py-2 rounded"
              >
                Cancel
              </button>
              <button
                onClick={() => handleDeleteBackup(showDeleteConfirm)}
                className="flex-1 bg-[#DC2626] hover:bg-[#EF4444] text-white text-xs py-2 rounded"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
