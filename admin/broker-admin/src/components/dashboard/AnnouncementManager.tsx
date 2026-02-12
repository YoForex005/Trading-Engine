'use client';

import { useState, useMemo } from 'react';
import { Plus, Edit2, Trash2, Eye, Copy, Calendar, Users, Bell, X, Check, AlertTriangle, Info, Megaphone, Wrench } from 'lucide-react';

// Types
type AnnouncementType = 'info' | 'warning' | 'critical' | 'maintenance' | 'promo';
type AnnouncementStatus = 'active' | 'scheduled' | 'expired' | 'draft';
type DisplayPosition = 'top-banner' | 'popup' | 'notification' | 'ticker';
type TargetAudience = 'all' | 'group' | 'account';

interface Announcement {
  id: string;
  title: string;
  message: string;
  type: AnnouncementType;
  status: AnnouncementStatus;
  targetAudience: TargetAudience;
  targetGroups?: string[];
  targetAccounts?: string[];
  startDate: Date;
  endDate: Date;
  displayPosition: DisplayPosition;
  priority: number;
  dismissable: boolean;
  views: number;
  dismissals: number;
  createdAt: Date;
  updatedAt: Date;
}

// Mock data generator
const generateMockAnnouncements = (): Announcement[] => {
  const now = new Date();
  const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000);
  const tomorrow = new Date(now.getTime() + 24 * 60 * 60 * 1000);
  const nextWeek = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000);
  const lastWeek = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  return [
    {
      id: 'ANN-001',
      title: 'Platform Maintenance Scheduled',
      message: 'Our trading platform will undergo scheduled maintenance on Sunday from 2:00 AM to 6:00 AM UTC. Trading will be temporarily unavailable during this period.',
      type: 'maintenance',
      status: 'active',
      targetAudience: 'all',
      startDate: now,
      endDate: nextWeek,
      displayPosition: 'top-banner',
      priority: 9,
      dismissable: false,
      views: 1245,
      dismissals: 0,
      createdAt: yesterday,
      updatedAt: yesterday,
    },
    {
      id: 'ANN-002',
      title: 'New Trading Instruments Available',
      message: 'We are excited to announce the addition of 15 new cryptocurrency pairs including <strong>SOLUSD</strong>, <strong>ADAUSD</strong>, and <strong>DOTUSD</strong>. Start trading now!',
      type: 'promo',
      status: 'active',
      targetAudience: 'all',
      startDate: yesterday,
      endDate: nextWeek,
      displayPosition: 'notification',
      priority: 7,
      dismissable: true,
      views: 3421,
      dismissals: 892,
      createdAt: lastWeek,
      updatedAt: yesterday,
    },
    {
      id: 'ANN-003',
      title: 'Critical: Margin Requirements Updated',
      message: 'Effective immediately, margin requirements for volatile instruments have been increased by 20%. Please review your open positions to avoid margin calls.',
      type: 'critical',
      status: 'active',
      targetAudience: 'all',
      startDate: now,
      endDate: tomorrow,
      displayPosition: 'popup',
      priority: 10,
      dismissable: false,
      views: 2156,
      dismissals: 0,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: 'ANN-004',
      title: 'Holiday Trading Hours',
      message: 'Please note that trading hours will be modified during the upcoming holiday season. Check our <a href="#">holiday schedule</a> for details.',
      type: 'info',
      status: 'scheduled',
      targetAudience: 'all',
      startDate: tomorrow,
      endDate: nextWeek,
      displayPosition: 'top-banner',
      priority: 6,
      dismissable: true,
      views: 0,
      dismissals: 0,
      createdAt: yesterday,
      updatedAt: yesterday,
    },
    {
      id: 'ANN-005',
      title: 'VIP Account Benefits',
      message: 'Congratulations on achieving VIP status! You now have access to <strong>lower spreads</strong>, <strong>priority support</strong>, and <strong>exclusive market research</strong>.',
      type: 'promo',
      status: 'active',
      targetAudience: 'group',
      targetGroups: ['VIP', 'Premium'],
      startDate: lastWeek,
      endDate: nextWeek,
      displayPosition: 'notification',
      priority: 5,
      dismissable: true,
      views: 156,
      dismissals: 45,
      createdAt: lastWeek,
      updatedAt: lastWeek,
    },
    {
      id: 'ANN-006',
      title: 'Swap Rate Changes',
      message: 'Swap rates for major currency pairs will be adjusted starting next Monday due to central bank policy changes.',
      type: 'warning',
      status: 'scheduled',
      targetAudience: 'all',
      startDate: tomorrow,
      endDate: nextWeek,
      displayPosition: 'ticker',
      priority: 7,
      dismissable: true,
      views: 0,
      dismissals: 0,
      createdAt: yesterday,
      updatedAt: yesterday,
    },
    {
      id: 'ANN-007',
      title: 'Server Upgrade Complete',
      message: 'Our server infrastructure upgrade has been completed successfully. You should notice improved order execution speeds and platform stability.',
      type: 'info',
      status: 'expired',
      targetAudience: 'all',
      startDate: lastWeek,
      endDate: yesterday,
      displayPosition: 'notification',
      priority: 4,
      dismissable: true,
      views: 5234,
      dismissals: 4891,
      createdAt: lastWeek,
      updatedAt: lastWeek,
    },
    {
      id: 'ANN-008',
      title: 'Special Commission Rates for High Volume Traders',
      message: 'For the next 30 days, traders with monthly volume exceeding $10M will receive a 25% commission discount. Trade more, save more!',
      type: 'promo',
      status: 'active',
      targetAudience: 'group',
      targetGroups: ['High Volume', 'Institutional'],
      startDate: yesterday,
      endDate: nextWeek,
      displayPosition: 'popup',
      priority: 8,
      dismissable: true,
      views: 432,
      dismissals: 89,
      createdAt: yesterday,
      updatedAt: yesterday,
    },
    {
      id: 'ANN-009',
      title: 'New Risk Management Tools',
      message: 'We have added advanced risk management features including trailing stops, guaranteed stop-loss orders, and position hedging. <a href="#">Learn more</a>',
      type: 'info',
      status: 'draft',
      targetAudience: 'all',
      startDate: tomorrow,
      endDate: nextWeek,
      displayPosition: 'notification',
      priority: 6,
      dismissable: true,
      views: 0,
      dismissals: 0,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: 'ANN-010',
      title: 'Regulatory Compliance Update',
      message: 'Due to new regulatory requirements, all clients must complete identity verification by the end of the month. <strong>Action required</strong>.',
      type: 'warning',
      status: 'active',
      targetAudience: 'all',
      startDate: lastWeek,
      endDate: nextWeek,
      displayPosition: 'top-banner',
      priority: 9,
      dismissable: false,
      views: 2891,
      dismissals: 0,
      createdAt: lastWeek,
      updatedAt: lastWeek,
    },
  ];
};

// Type configurations
const typeConfig: Record<AnnouncementType, { label: string; color: string; icon: React.ReactNode; bg: string }> = {
  info: {
    label: 'Info',
    color: 'text-blue-400',
    icon: <Info size={14} />,
    bg: 'bg-blue-500/10',
  },
  warning: {
    label: 'Warning',
    color: 'text-yellow-400',
    icon: <AlertTriangle size={14} />,
    bg: 'bg-yellow-500/10',
  },
  critical: {
    label: 'Critical',
    color: 'text-red-400',
    icon: <AlertTriangle size={14} />,
    bg: 'bg-red-500/10',
  },
  maintenance: {
    label: 'Maintenance',
    color: 'text-orange-400',
    icon: <Wrench size={14} />,
    bg: 'bg-orange-500/10',
  },
  promo: {
    label: 'Promo',
    color: 'text-purple-400',
    icon: <Megaphone size={14} />,
    bg: 'bg-purple-500/10',
  },
};

const statusConfig: Record<AnnouncementStatus, { label: string; color: string; bg: string }> = {
  active: { label: 'Active', color: 'text-green-400', bg: 'bg-green-500/20' },
  scheduled: { label: 'Scheduled', color: 'text-blue-400', bg: 'bg-blue-500/20' },
  expired: { label: 'Expired', color: 'text-gray-400', bg: 'bg-gray-500/20' },
  draft: { label: 'Draft', color: 'text-yellow-400', bg: 'bg-yellow-500/20' },
};

export default function AnnouncementManager() {
  const [announcements, setAnnouncements] = useState<Announcement[]>(generateMockAnnouncements());
  const [selectedAnnouncement, setSelectedAnnouncement] = useState<Announcement | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isPreviewOpen, setIsPreviewOpen] = useState(false);
  const [previewAnnouncement, setPreviewAnnouncement] = useState<Announcement | null>(null);
  const [filterStatus, setFilterStatus] = useState<AnnouncementStatus | 'all'>('all');
  const [filterType, setFilterType] = useState<AnnouncementType | 'all'>('all');
  const [searchQuery, setSearchQuery] = useState('');

  // Form state
  const [formData, setFormData] = useState<Partial<Announcement>>({
    title: '',
    message: '',
    type: 'info',
    targetAudience: 'all',
    displayPosition: 'notification',
    priority: 5,
    dismissable: true,
  });

  // Calculate summary metrics
  const metrics = useMemo(() => {
    const active = announcements.filter(a => a.status === 'active').length;
    const scheduled = announcements.filter(a => a.status === 'scheduled').length;
    const totalViews = announcements.reduce((sum, a) => sum + a.views, 0);
    const totalDismissals = announcements.reduce((sum, a) => sum + a.dismissals, 0);
    const dismissRate = totalViews > 0 ? ((totalDismissals / totalViews) * 100).toFixed(1) : '0.0';

    return { active, scheduled, totalViews, dismissRate };
  }, [announcements]);

  // Filtered and sorted announcements
  const filteredAnnouncements = useMemo(() => {
    return announcements
      .filter(a => {
        if (filterStatus !== 'all' && a.status !== filterStatus) return false;
        if (filterType !== 'all' && a.type !== filterType) return false;
        if (searchQuery && !a.title.toLowerCase().includes(searchQuery.toLowerCase()) &&
            !a.message.toLowerCase().includes(searchQuery.toLowerCase())) return false;
        return true;
      })
      .sort((a, b) => {
        // Active announcements first, then by priority
        if (a.status === 'active' && b.status !== 'active') return -1;
        if (a.status !== 'active' && b.status === 'active') return 1;
        return b.priority - a.priority;
      });
  }, [announcements, filterStatus, filterType, searchQuery]);

  const handleCreate = () => {
    setSelectedAnnouncement(null);
    setFormData({
      title: '',
      message: '',
      type: 'info',
      targetAudience: 'all',
      displayPosition: 'notification',
      priority: 5,
      dismissable: true,
    });
    setIsModalOpen(true);
  };

  const handleEdit = (announcement: Announcement) => {
    setSelectedAnnouncement(announcement);
    setFormData(announcement);
    setIsModalOpen(true);
  };

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this announcement?')) {
      setAnnouncements(prev => prev.filter(a => a.id !== id));
    }
  };

  const handlePreview = (announcement: Announcement) => {
    setPreviewAnnouncement(announcement);
    setIsPreviewOpen(true);
  };

  const handleSave = () => {
    if (!formData.title || !formData.message) {
      alert('Please fill in title and message');
      return;
    }

    const now = new Date();
    if (selectedAnnouncement) {
      // Update existing
      setAnnouncements(prev =>
        prev.map(a =>
          a.id === selectedAnnouncement.id
            ? { ...a, ...formData, updatedAt: now } as Announcement
            : a
        )
      );
    } else {
      // Create new
      const newAnnouncement: Announcement = {
        id: `ANN-${String(announcements.length + 1).padStart(3, '0')}`,
        title: formData.title!,
        message: formData.message!,
        type: formData.type || 'info',
        status: formData.status || 'draft',
        targetAudience: formData.targetAudience || 'all',
        targetGroups: formData.targetGroups,
        targetAccounts: formData.targetAccounts,
        startDate: formData.startDate || now,
        endDate: formData.endDate || new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000),
        displayPosition: formData.displayPosition || 'notification',
        priority: formData.priority || 5,
        dismissable: formData.dismissable !== undefined ? formData.dismissable : true,
        views: 0,
        dismissals: 0,
        createdAt: now,
        updatedAt: now,
      };
      setAnnouncements(prev => [newAnnouncement, ...prev]);
    }

    setIsModalOpen(false);
  };

  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  };

  const renderPreview = (announcement: Announcement, position: DisplayPosition) => {
    const config = typeConfig[announcement.type];

    switch (position) {
      case 'top-banner':
        return (
          <div className={`w-full p-3 flex items-center gap-3 ${config.bg} border-l-4`} style={{ borderLeftColor: config.color.replace('text-', '#') }}>
            <div className={config.color}>{config.icon}</div>
            <div className="flex-1">
              <div className="text-sm font-semibold text-white">{announcement.title}</div>
              <div className="text-xs text-gray-300 mt-0.5" dangerouslySetInnerHTML={{ __html: announcement.message }} />
            </div>
            {announcement.dismissable && (
              <button className="text-gray-400 hover:text-white">
                <X size={16} />
              </button>
            )}
          </div>
        );

      case 'popup':
        return (
          <div className="w-96 bg-[#1E2026] border border-[#383A42] rounded-lg shadow-2xl p-4">
            <div className="flex items-start gap-3 mb-3">
              <div className={`${config.color} mt-0.5`}>{config.icon}</div>
              <div className="flex-1">
                <div className="text-base font-semibold text-white mb-1">{announcement.title}</div>
                <div className="text-sm text-gray-300" dangerouslySetInnerHTML={{ __html: announcement.message }} />
              </div>
              {announcement.dismissable && (
                <button className="text-gray-400 hover:text-white">
                  <X size={18} />
                </button>
              )}
            </div>
            <div className="flex gap-2 justify-end">
              <button className="px-3 py-1.5 text-xs bg-blue-600 hover:bg-blue-700 text-white rounded">
                Learn More
              </button>
            </div>
          </div>
        );

      case 'notification':
        return (
          <div className="w-80 bg-[#1E2026] border border-[#383A42] rounded shadow-lg p-3 flex items-start gap-2">
            <div className={config.color}>{config.icon}</div>
            <div className="flex-1 min-w-0">
              <div className="text-sm font-semibold text-white mb-0.5">{announcement.title}</div>
              <div className="text-xs text-gray-400 line-clamp-2" dangerouslySetInnerHTML={{ __html: announcement.message }} />
            </div>
            {announcement.dismissable && (
              <button className="text-gray-500 hover:text-white flex-shrink-0">
                <X size={14} />
              </button>
            )}
          </div>
        );

      case 'ticker':
        return (
          <div className={`w-full py-1.5 px-3 flex items-center gap-2 ${config.bg} text-xs`}>
            <div className={config.color}>{config.icon}</div>
            <div className="flex-1 text-gray-200 truncate">
              <span className="font-semibold">{announcement.title}:</span> {announcement.message.replace(/<[^>]*>/g, '')}
            </div>
          </div>
        );
    }
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#CCC]">
      {/* Header */}
      <div className="flex-shrink-0 border-b border-[#383A42] bg-[#1E2026] px-4 py-3">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-lg font-bold text-white">Announcement Management</h1>
            <p className="text-xs text-gray-400 mt-0.5">Manage platform-wide announcements and banners</p>
          </div>
          <button
            onClick={handleCreate}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded transition-colors"
          >
            <Plus size={16} />
            Create Announcement
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="flex-shrink-0 grid grid-cols-4 gap-4 p-4 border-b border-[#383A42] bg-[#1A1C20]">
        <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-gray-400">Active Announcements</span>
            <Bell size={16} className="text-green-400" />
          </div>
          <div className="text-2xl font-bold text-white">{metrics.active}</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-gray-400">Scheduled</span>
            <Calendar size={16} className="text-blue-400" />
          </div>
          <div className="text-2xl font-bold text-white">{metrics.scheduled}</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-gray-400">Total Views</span>
            <Eye size={16} className="text-purple-400" />
          </div>
          <div className="text-2xl font-bold text-white">{metrics.totalViews.toLocaleString()}</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-gray-400">Dismiss Rate</span>
            <X size={16} className="text-orange-400" />
          </div>
          <div className="text-2xl font-bold text-white">{metrics.dismissRate}%</div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex-shrink-0 flex items-center gap-3 px-4 py-3 border-b border-[#383A42] bg-[#1A1C20]">
        <input
          type="text"
          placeholder="Search announcements..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="flex-1 px-3 py-1.5 bg-[#121316] border border-[#383A42] rounded text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
        />
        <select
          value={filterStatus}
          onChange={(e) => setFilterStatus(e.target.value as any)}
          className="px-3 py-1.5 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
        >
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="scheduled">Scheduled</option>
          <option value="expired">Expired</option>
          <option value="draft">Draft</option>
        </select>
        <select
          value={filterType}
          onChange={(e) => setFilterType(e.target.value as any)}
          className="px-3 py-1.5 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
        >
          <option value="all">All Types</option>
          <option value="info">Info</option>
          <option value="warning">Warning</option>
          <option value="critical">Critical</option>
          <option value="maintenance">Maintenance</option>
          <option value="promo">Promo</option>
        </select>
      </div>

      {/* Announcements Table */}
      <div className="flex-1 overflow-auto">
        <table className="w-full text-xs">
          <thead className="bg-[#1E2026] sticky top-0 z-10">
            <tr className="border-b border-[#383A42]">
              <th className="text-left px-4 py-3 font-semibold text-gray-400">ID</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-400">Title</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-400">Type</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-400">Target Audience</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-400">Start Date</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-400">End Date</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-400">Status</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-400">Views</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-400">Priority</th>
              <th className="text-right px-4 py-3 font-semibold text-gray-400">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredAnnouncements.map((announcement, idx) => {
              const typeConf = typeConfig[announcement.type];
              const statusConf = statusConfig[announcement.status];
              const isActive = announcement.status === 'active';

              return (
                <tr
                  key={announcement.id}
                  className={`border-b border-[#2A2C34] hover:bg-[#1E2026] transition-colors ${isActive ? 'bg-green-500/5' : ''}`}
                >
                  <td className="px-4 py-3">
                    <span className={`font-mono ${isActive ? 'text-green-400' : 'text-gray-400'}`}>
                      {announcement.id}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className={`font-medium ${isActive ? 'text-white' : 'text-gray-300'}`}>
                      {announcement.title}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div className={`inline-flex items-center gap-1.5 px-2 py-1 rounded ${typeConf.bg} ${typeConf.color}`}>
                      {typeConf.icon}
                      <span>{typeConf.label}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-gray-400">
                    {announcement.targetAudience === 'all' ? 'All Users' : announcement.targetAudience === 'group' ? `Groups (${announcement.targetGroups?.length || 0})` : `Accounts (${announcement.targetAccounts?.length || 0})`}
                  </td>
                  <td className="px-4 py-3 text-gray-400">{formatDate(announcement.startDate)}</td>
                  <td className="px-4 py-3 text-gray-400">{formatDate(announcement.endDate)}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-block px-2 py-1 rounded text-xs font-medium ${statusConf.bg} ${statusConf.color}`}>
                      {statusConf.label}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-400">
                    {announcement.views.toLocaleString()}
                    {announcement.dismissable && announcement.views > 0 && (
                      <span className="text-gray-500 ml-1">
                        ({((announcement.dismissals / announcement.views) * 100).toFixed(0)}%)
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`font-semibold ${announcement.priority >= 8 ? 'text-red-400' : announcement.priority >= 5 ? 'text-yellow-400' : 'text-gray-400'}`}>
                      {announcement.priority}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => handlePreview(announcement)}
                        className="p-1.5 text-blue-400 hover:text-blue-300 hover:bg-blue-500/10 rounded transition-colors"
                        title="Preview"
                      >
                        <Eye size={14} />
                      </button>
                      <button
                        onClick={() => handleEdit(announcement)}
                        className="p-1.5 text-yellow-400 hover:text-yellow-300 hover:bg-yellow-500/10 rounded transition-colors"
                        title="Edit"
                      >
                        <Edit2 size={14} />
                      </button>
                      <button
                        onClick={() => handleDelete(announcement.id)}
                        className="p-1.5 text-red-400 hover:text-red-300 hover:bg-red-500/10 rounded transition-colors"
                        title="Delete"
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>

        {filteredAnnouncements.length === 0 && (
          <div className="flex flex-col items-center justify-center py-16 text-gray-500">
            <Bell size={48} className="mb-4 opacity-20" />
            <p className="text-sm">No announcements found</p>
          </div>
        )}
      </div>

      {/* Create/Edit Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg w-full max-w-3xl max-h-[90vh] overflow-hidden flex flex-col">
            {/* Modal Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-[#383A42]">
              <h2 className="text-lg font-bold text-white">
                {selectedAnnouncement ? 'Edit Announcement' : 'Create Announcement'}
              </h2>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-gray-400 hover:text-white transition-colors"
              >
                <X size={20} />
              </button>
            </div>

            {/* Modal Body */}
            <div className="flex-1 overflow-auto p-6 space-y-4">
              {/* Title */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1.5">Title *</label>
                <input
                  type="text"
                  value={formData.title || ''}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
                  placeholder="Enter announcement title"
                />
              </div>

              {/* Message */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1.5">Message *</label>
                <textarea
                  value={formData.message || ''}
                  onChange={(e) => setFormData({ ...formData, message: e.target.value })}
                  rows={4}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500 resize-none"
                  placeholder="Enter announcement message. Use <strong>bold</strong>, <em>italic</em>, or <a href='#'>links</a>"
                />
                <p className="text-xs text-gray-500 mt-1">Supports HTML: &lt;strong&gt;, &lt;em&gt;, &lt;a href=""&gt;</p>
              </div>

              {/* Type and Priority */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1.5">Type</label>
                  <select
                    value={formData.type || 'info'}
                    onChange={(e) => setFormData({ ...formData, type: e.target.value as AnnouncementType })}
                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
                  >
                    <option value="info">Info</option>
                    <option value="warning">Warning</option>
                    <option value="critical">Critical</option>
                    <option value="maintenance">Maintenance</option>
                    <option value="promo">Promo</option>
                  </select>
                  {formData.type && (
                    <div className={`mt-2 inline-flex items-center gap-1.5 px-2 py-1 rounded text-xs ${typeConfig[formData.type].bg} ${typeConfig[formData.type].color}`}>
                      {typeConfig[formData.type].icon}
                      <span>Preview: {typeConfig[formData.type].label}</span>
                    </div>
                  )}
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1.5">Priority (1-10)</label>
                  <input
                    type="number"
                    min="1"
                    max="10"
                    value={formData.priority || 5}
                    onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) || 5 })}
                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
                  />
                </div>
              </div>

              {/* Target Audience */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1.5">Target Audience</label>
                <select
                  value={formData.targetAudience || 'all'}
                  onChange={(e) => setFormData({ ...formData, targetAudience: e.target.value as TargetAudience })}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
                >
                  <option value="all">All Users</option>
                  <option value="group">Specific Trading Groups</option>
                  <option value="account">Specific Accounts</option>
                </select>
                {formData.targetAudience === 'group' && (
                  <input
                    type="text"
                    placeholder="Enter group names (comma-separated)"
                    className="w-full mt-2 px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
                    onChange={(e) => setFormData({ ...formData, targetGroups: e.target.value.split(',').map(s => s.trim()) })}
                  />
                )}
                {formData.targetAudience === 'account' && (
                  <input
                    type="text"
                    placeholder="Enter account IDs (comma-separated)"
                    className="w-full mt-2 px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
                    onChange={(e) => setFormData({ ...formData, targetAccounts: e.target.value.split(',').map(s => s.trim()) })}
                  />
                )}
              </div>

              {/* Display Position */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1.5">Display Position</label>
                <select
                  value={formData.displayPosition || 'notification'}
                  onChange={(e) => setFormData({ ...formData, displayPosition: e.target.value as DisplayPosition })}
                  className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
                >
                  <option value="top-banner">Top Banner</option>
                  <option value="popup">Popup Modal</option>
                  <option value="notification">Notification Toast</option>
                  <option value="ticker">Scrolling Ticker</option>
                </select>
              </div>

              {/* Schedule */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1.5">Start Date</label>
                  <input
                    type="datetime-local"
                    value={formData.startDate ? new Date(formData.startDate.getTime() - formData.startDate.getTimezoneOffset() * 60000).toISOString().slice(0, 16) : ''}
                    onChange={(e) => setFormData({ ...formData, startDate: new Date(e.target.value) })}
                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1.5">End Date</label>
                  <input
                    type="datetime-local"
                    value={formData.endDate ? new Date(formData.endDate.getTime() - formData.endDate.getTimezoneOffset() * 60000).toISOString().slice(0, 16) : ''}
                    onChange={(e) => setFormData({ ...formData, endDate: new Date(e.target.value) })}
                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-blue-500"
                  />
                </div>
              </div>

              {/* Dismissable Toggle */}
              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  id="dismissable"
                  checked={formData.dismissable !== undefined ? formData.dismissable : true}
                  onChange={(e) => setFormData({ ...formData, dismissable: e.target.checked })}
                  className="w-4 h-4 text-blue-600 bg-[#121316] border-[#383A42] rounded focus:ring-blue-500"
                />
                <label htmlFor="dismissable" className="text-sm text-gray-300">
                  Allow users to dismiss this announcement
                </label>
              </div>
            </div>

            {/* Modal Footer */}
            <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-[#383A42]">
              <button
                onClick={() => setIsModalOpen(false)}
                className="px-4 py-2 bg-[#2A2C34] hover:bg-[#33353D] text-white text-sm font-medium rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSave}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded transition-colors"
              >
                {selectedAnnouncement ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Preview Modal */}
      {isPreviewOpen && previewAnnouncement && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
            {/* Preview Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-[#383A42]">
              <h2 className="text-lg font-bold text-white">Announcement Preview</h2>
              <button
                onClick={() => setIsPreviewOpen(false)}
                className="text-gray-400 hover:text-white transition-colors"
              >
                <X size={20} />
              </button>
            </div>

            {/* Preview Body */}
            <div className="flex-1 overflow-auto p-6 space-y-6">
              <div>
                <h3 className="text-sm font-semibold text-gray-300 mb-3">Top Banner</h3>
                {renderPreview(previewAnnouncement, 'top-banner')}
              </div>
              <div>
                <h3 className="text-sm font-semibold text-gray-300 mb-3">Popup Modal</h3>
                <div className="flex justify-center">
                  {renderPreview(previewAnnouncement, 'popup')}
                </div>
              </div>
              <div>
                <h3 className="text-sm font-semibold text-gray-300 mb-3">Notification Toast</h3>
                <div className="flex justify-end">
                  {renderPreview(previewAnnouncement, 'notification')}
                </div>
              </div>
              <div>
                <h3 className="text-sm font-semibold text-gray-300 mb-3">Scrolling Ticker</h3>
                {renderPreview(previewAnnouncement, 'ticker')}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
