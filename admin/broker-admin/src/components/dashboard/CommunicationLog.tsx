'use client';

import { useState, useMemo } from 'react';
import { Mail, MessageSquare, Bell, Shield, Headphones, X, Search, Send, Calendar, Filter, ChevronDown } from 'lucide-react';

// Types
type CommunicationType = 'email' | 'sms' | 'push' | 'system' | 'support';
type CommunicationStatus = 'sent' | 'delivered' | 'opened' | 'bounced' | 'failed';

interface CommunicationEntry {
  id: string;
  type: CommunicationType;
  clientId: string;
  clientName: string;
  clientEmail: string;
  clientPhone: string;
  accountType: string;
  subject: string;
  preview: string;
  content: string;
  status: CommunicationStatus;
  sentAt: Date;
  deliveredAt?: Date;
  openedAt?: Date;
}

interface MessageTemplate {
  id: string;
  name: string;
  subject: string;
  body: string;
  channel: 'email' | 'sms' | 'push';
}

export default function CommunicationLog() {
  const [entries, setEntries] = useState<CommunicationEntry[]>(generateMockEntries());
  const [templates] = useState<MessageTemplate[]>(generateMockTemplates());
  const [selectedEntry, setSelectedEntry] = useState<CommunicationEntry | null>(null);
  const [showCompose, setShowCompose] = useState(false);

  // Filters
  const [typeFilter, setTypeFilter] = useState<CommunicationType | 'all'>('all');
  const [statusFilter, setStatusFilter] = useState<CommunicationStatus | 'all'>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortOrder, setSortOrder] = useState<'newest' | 'oldest'>('newest');

  // Compose form state
  const [composeRecipient, setComposeRecipient] = useState('');
  const [composeChannel, setComposeChannel] = useState<'email' | 'sms' | 'push'>('email');
  const [composeTemplate, setComposeTemplate] = useState('');
  const [composeSubject, setComposeSubject] = useState('');
  const [composeBody, setComposeBody] = useState('');

  // Filter and sort entries
  const filteredEntries = useMemo(() => {
    let result = entries;

    if (typeFilter !== 'all') {
      result = result.filter(e => e.type === typeFilter);
    }

    if (statusFilter !== 'all') {
      result = result.filter(e => e.status === statusFilter);
    }

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter(e =>
        e.clientName.toLowerCase().includes(query) ||
        e.subject.toLowerCase().includes(query) ||
        e.preview.toLowerCase().includes(query)
      );
    }

    result.sort((a, b) => {
      if (sortOrder === 'newest') {
        return b.sentAt.getTime() - a.sentAt.getTime();
      } else {
        return a.sentAt.getTime() - b.sentAt.getTime();
      }
    });

    return result;
  }, [entries, typeFilter, statusFilter, searchQuery, sortOrder]);

  // Calculate stats
  const stats = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const todayEntries = entries.filter(e => e.sentAt >= today);
    const delivered = entries.filter(e => e.status === 'delivered' || e.status === 'opened').length;
    const opened = entries.filter(e => e.status === 'opened').length;
    const pending = entries.filter(e => e.status === 'sent').length;

    return {
      totalSentToday: todayEntries.length,
      deliveryRate: entries.length > 0 ? ((delivered / entries.length) * 100).toFixed(1) : '0.0',
      openRate: entries.length > 0 ? ((opened / entries.length) * 100).toFixed(1) : '0.0',
      pending,
    };
  }, [entries]);

  // Calculate donut chart data
  const donutData = useMemo(() => {
    const counts = {
      email: entries.filter(e => e.type === 'email').length,
      sms: entries.filter(e => e.type === 'sms').length,
      push: entries.filter(e => e.type === 'push').length,
      system: entries.filter(e => e.type === 'system').length,
      support: entries.filter(e => e.type === 'support').length,
    };

    const total = Object.values(counts).reduce((sum, count) => sum + count, 0);

    const segments = [
      { label: 'Email', count: counts.email, color: '#3B82F6', percentage: (counts.email / total) * 100 },
      { label: 'SMS', count: counts.sms, color: '#10B981', percentage: (counts.sms / total) * 100 },
      { label: 'Push', count: counts.push, color: '#8B5CF6', percentage: (counts.push / total) * 100 },
      { label: 'System', count: counts.system, color: '#F59E0B', percentage: (counts.system / total) * 100 },
      { label: 'Support', count: counts.support, color: '#EF4444', percentage: (counts.support / total) * 100 },
    ];

    // Calculate SVG paths
    const radius = 80;
    const innerRadius = 50;
    let currentAngle = -90;

    const paths = segments.map(segment => {
      const angleSize = (segment.percentage / 100) * 360;
      const endAngle = currentAngle + angleSize;

      const x1 = 100 + radius * Math.cos((currentAngle * Math.PI) / 180);
      const y1 = 100 + radius * Math.sin((currentAngle * Math.PI) / 180);
      const x2 = 100 + radius * Math.cos((endAngle * Math.PI) / 180);
      const y2 = 100 + radius * Math.sin((endAngle * Math.PI) / 180);
      const x3 = 100 + innerRadius * Math.cos((endAngle * Math.PI) / 180);
      const y3 = 100 + innerRadius * Math.sin((endAngle * Math.PI) / 180);
      const x4 = 100 + innerRadius * Math.cos((currentAngle * Math.PI) / 180);
      const y4 = 100 + innerRadius * Math.sin((currentAngle * Math.PI) / 180);

      const largeArc = angleSize > 180 ? 1 : 0;

      const path = `M ${x1} ${y1} A ${radius} ${radius} 0 ${largeArc} 1 ${x2} ${y2} L ${x3} ${y3} A ${innerRadius} ${innerRadius} 0 ${largeArc} 0 ${x4} ${y4} Z`;

      currentAngle = endAngle;

      return { ...segment, path };
    });

    return paths;
  }, [entries]);

  const typeConfig: Record<CommunicationType, { label: string; icon: React.ReactNode; color: string; bg: string }> = {
    email: { label: 'Email', icon: <Mail size={14} />, color: 'text-blue-400', bg: 'bg-blue-500/10' },
    sms: { label: 'SMS', icon: <MessageSquare size={14} />, color: 'text-green-400', bg: 'bg-green-500/10' },
    push: { label: 'Push', icon: <Bell size={14} />, color: 'text-purple-400', bg: 'bg-purple-500/10' },
    system: { label: 'System', icon: <Shield size={14} />, color: 'text-yellow-400', bg: 'bg-yellow-500/10' },
    support: { label: 'Support', icon: <Headphones size={14} />, color: 'text-red-400', bg: 'bg-red-500/10' },
  };

  const statusConfig: Record<CommunicationStatus, { label: string; color: string; bg: string }> = {
    sent: { label: 'Sent', color: 'text-blue-400', bg: 'bg-blue-500/10' },
    delivered: { label: 'Delivered', color: 'text-green-400', bg: 'bg-green-500/10' },
    opened: { label: 'Opened', color: 'text-purple-400', bg: 'bg-purple-500/10' },
    bounced: { label: 'Bounced', color: 'text-orange-400', bg: 'bg-orange-500/10' },
    failed: { label: 'Failed', color: 'text-red-400', bg: 'bg-red-500/10' },
  };

  const handleTemplateSelect = (templateId: string) => {
    const template = templates.find(t => t.id === templateId);
    if (template) {
      setComposeTemplate(templateId);
      setComposeSubject(template.subject);
      setComposeBody(template.body);
      setComposeChannel(template.channel);
    }
  };

  const handleSendMessage = () => {
    // Mock sending logic
    console.log('Sending message:', { composeRecipient, composeChannel, composeSubject, composeBody });
    setShowCompose(false);
    setComposeRecipient('');
    setComposeTemplate('');
    setComposeSubject('');
    setComposeBody('');
  };

  const formatTimestamp = (date: Date) => {
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (hours < 1) return 'Just now';
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-white">
      {/* Header */}
      <div className="h-12 bg-[#1E2026] border-b border-[#383A42] flex items-center justify-between px-4 flex-shrink-0">
        <h1 className="text-sm font-bold text-[#F5C542]">Communication Log</h1>
        <button
          onClick={() => setShowCompose(true)}
          className="px-3 py-1.5 bg-[#F5C542] text-black text-xs font-bold rounded hover:bg-[#FFD700] transition-colors flex items-center gap-2"
        >
          <Send size={14} />
          Compose Message
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4 p-4 flex-shrink-0">
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Total Sent Today</div>
          <div className="text-2xl font-bold text-white">{stats.totalSentToday.toLocaleString()}</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Delivery Rate</div>
          <div className="text-2xl font-bold text-green-400">{stats.deliveryRate}%</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Open Rate</div>
          <div className="text-2xl font-bold text-purple-400">{stats.openRate}%</div>
        </div>
        <div className="bg-[#1E2026] border border-[#383A42] rounded p-3">
          <div className="text-xs text-[#888] mb-1">Pending</div>
          <div className="text-2xl font-bold text-yellow-400">{stats.pending}</div>
        </div>
      </div>

      {/* Main Content: Chart + Table */}
      <div className="flex-1 overflow-hidden flex gap-4 p-4 pt-0">
        {/* Donut Chart */}
        <div className="w-80 bg-[#1E2026] border border-[#383A42] rounded p-4 flex-shrink-0">
          <div className="text-sm font-bold text-[#CCC] mb-4">Communication Breakdown</div>
          <div className="flex items-center justify-center mb-4">
            <svg width="200" height="200" viewBox="0 0 200 200">
              {donutData.map((segment, index) => (
                <path
                  key={index}
                  d={segment.path}
                  fill={segment.color}
                  opacity="0.9"
                />
              ))}
              <text x="100" y="100" textAnchor="middle" dy=".3em" className="text-2xl font-bold fill-white">
                {entries.length}
              </text>
              <text x="100" y="120" textAnchor="middle" className="text-xs fill-[#888]">
                Total
              </text>
            </svg>
          </div>
          <div className="space-y-2">
            {donutData.map((segment, index) => (
              <div key={index} className="flex items-center justify-between text-xs">
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded" style={{ backgroundColor: segment.color }}></div>
                  <span className="text-[#CCC]">{segment.label}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-[#888]">{segment.count}</span>
                  <span className="text-[#666]">({segment.percentage.toFixed(1)}%)</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Timeline Table */}
        <div className="flex-1 bg-[#1E2026] border border-[#383A42] rounded overflow-hidden flex flex-col">
          {/* Filters */}
          <div className="p-3 border-b border-[#383A42] flex items-center gap-3 flex-shrink-0">
            <div className="flex items-center gap-2 bg-[#25272E] rounded px-3 py-1.5 flex-1">
              <Search size={14} className="text-[#888]" />
              <input
                type="text"
                placeholder="Search client or subject..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="bg-transparent text-xs text-white outline-none flex-1"
              />
            </div>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value as any)}
              className="bg-[#25272E] text-xs text-white rounded px-3 py-1.5 outline-none border border-[#383A42]"
            >
              <option value="all">All Types</option>
              <option value="email">Email</option>
              <option value="sms">SMS</option>
              <option value="push">Push</option>
              <option value="system">System</option>
              <option value="support">Support</option>
            </select>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as any)}
              className="bg-[#25272E] text-xs text-white rounded px-3 py-1.5 outline-none border border-[#383A42]"
            >
              <option value="all">All Status</option>
              <option value="sent">Sent</option>
              <option value="delivered">Delivered</option>
              <option value="opened">Opened</option>
              <option value="bounced">Bounced</option>
              <option value="failed">Failed</option>
            </select>
            <select
              value={sortOrder}
              onChange={(e) => setSortOrder(e.target.value as any)}
              className="bg-[#25272E] text-xs text-white rounded px-3 py-1.5 outline-none border border-[#383A42]"
            >
              <option value="newest">Newest First</option>
              <option value="oldest">Oldest First</option>
            </select>
          </div>

          {/* Table Header */}
          <div className="grid grid-cols-[60px_200px_1fr_150px_120px_80px] gap-4 px-4 py-2 bg-[#25272E] border-b border-[#383A42] text-xs font-bold text-[#888] flex-shrink-0">
            <div>Type</div>
            <div>Client</div>
            <div>Subject / Preview</div>
            <div>Timestamp</div>
            <div>Status</div>
            <div>Actions</div>
          </div>

          {/* Table Body */}
          <div className="flex-1 overflow-y-auto">
            {filteredEntries.map((entry) => {
              const typeInfo = typeConfig[entry.type];
              const statusInfo = statusConfig[entry.status];

              return (
                <div
                  key={entry.id}
                  onClick={() => setSelectedEntry(entry)}
                  className="grid grid-cols-[60px_200px_1fr_150px_120px_80px] gap-4 px-4 py-3 border-b border-[#383A42] hover:bg-[#25272E] cursor-pointer transition-colors text-xs"
                >
                  <div className="flex items-center">
                    <div className={`${typeInfo.bg} ${typeInfo.color} p-2 rounded`}>
                      {typeInfo.icon}
                    </div>
                  </div>
                  <div className="flex flex-col justify-center">
                    <div className="text-white font-medium truncate">{entry.clientName}</div>
                    <div className="text-[#888] text-xs truncate">{entry.clientEmail}</div>
                  </div>
                  <div className="flex flex-col justify-center">
                    <div className="text-white font-medium truncate">{entry.subject}</div>
                    <div className="text-[#888] text-xs truncate">{entry.preview}</div>
                  </div>
                  <div className="flex items-center text-[#888]">
                    {formatTimestamp(entry.sentAt)}
                  </div>
                  <div className="flex items-center">
                    <span className={`${statusInfo.bg} ${statusInfo.color} px-2 py-1 rounded text-xs font-medium`}>
                      {statusInfo.label}
                    </span>
                  </div>
                  <div className="flex items-center">
                    <button className="text-[#F5C542] hover:text-[#FFD700] text-xs font-medium">
                      View
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Detail Modal */}
      {selectedEntry && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setSelectedEntry(null)}>
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg w-[800px] max-h-[600px] flex flex-col" onClick={(e) => e.stopPropagation()}>
            {/* Modal Header */}
            <div className="flex items-center justify-between p-4 border-b border-[#383A42]">
              <div className="flex items-center gap-3">
                <div className={`${typeConfig[selectedEntry.type].bg} ${typeConfig[selectedEntry.type].color} p-2 rounded`}>
                  {typeConfig[selectedEntry.type].icon}
                </div>
                <div>
                  <h3 className="text-sm font-bold text-white">{selectedEntry.subject}</h3>
                  <p className="text-xs text-[#888]">{selectedEntry.clientName} • {selectedEntry.clientEmail}</p>
                </div>
              </div>
              <button onClick={() => setSelectedEntry(null)} className="text-[#888] hover:text-white">
                <X size={20} />
              </button>
            </div>

            {/* Modal Body */}
            <div className="flex-1 overflow-y-auto">
              <div className="p-4 grid grid-cols-[1fr_250px] gap-4">
                {/* Message Content */}
                <div>
                  <div className="text-xs font-bold text-[#888] mb-2">Message Content</div>
                  <div className="bg-[#25272E] border border-[#383A42] rounded p-4 text-sm text-[#CCC] whitespace-pre-wrap">
                    {selectedEntry.content}
                  </div>

                  {/* Delivery Tracking */}
                  <div className="mt-4">
                    <div className="text-xs font-bold text-[#888] mb-2">Delivery Tracking</div>
                    <div className="space-y-2">
                      <div className="flex items-center gap-3">
                        <div className="w-2 h-2 rounded-full bg-green-400"></div>
                        <div className="text-xs text-[#CCC]">Sent</div>
                        <div className="text-xs text-[#888]">{selectedEntry.sentAt.toLocaleString()}</div>
                      </div>
                      {selectedEntry.deliveredAt && (
                        <div className="flex items-center gap-3">
                          <div className="w-2 h-2 rounded-full bg-green-400"></div>
                          <div className="text-xs text-[#CCC]">Delivered</div>
                          <div className="text-xs text-[#888]">{selectedEntry.deliveredAt.toLocaleString()}</div>
                        </div>
                      )}
                      {selectedEntry.openedAt && (
                        <div className="flex items-center gap-3">
                          <div className="w-2 h-2 rounded-full bg-purple-400"></div>
                          <div className="text-xs text-[#CCC]">Opened</div>
                          <div className="text-xs text-[#888]">{selectedEntry.openedAt.toLocaleString()}</div>
                        </div>
                      )}
                    </div>
                  </div>
                </div>

                {/* Client Info Sidebar */}
                <div className="bg-[#25272E] border border-[#383A42] rounded p-4">
                  <div className="text-xs font-bold text-[#888] mb-3">Client Information</div>
                  <div className="space-y-3 text-xs">
                    <div>
                      <div className="text-[#888]">Name</div>
                      <div className="text-white font-medium">{selectedEntry.clientName}</div>
                    </div>
                    <div>
                      <div className="text-[#888]">Email</div>
                      <div className="text-white">{selectedEntry.clientEmail}</div>
                    </div>
                    <div>
                      <div className="text-[#888]">Phone</div>
                      <div className="text-white">{selectedEntry.clientPhone}</div>
                    </div>
                    <div>
                      <div className="text-[#888]">Account Type</div>
                      <div className="text-white">{selectedEntry.accountType}</div>
                    </div>
                  </div>
                  <button className="w-full mt-4 px-4 py-2 bg-[#F5C542] text-black text-xs font-bold rounded hover:bg-[#FFD700] transition-colors">
                    Reply
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Compose Modal */}
      {showCompose && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowCompose(false)}>
          <div className="bg-[#1E2026] border border-[#383A42] rounded-lg w-[700px] max-h-[600px] flex flex-col" onClick={(e) => e.stopPropagation()}>
            {/* Modal Header */}
            <div className="flex items-center justify-between p-4 border-b border-[#383A42]">
              <h3 className="text-sm font-bold text-white">Compose Message</h3>
              <button onClick={() => setShowCompose(false)} className="text-[#888] hover:text-white">
                <X size={20} />
              </button>
            </div>

            {/* Modal Body */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
              {/* Recipient */}
              <div>
                <label className="text-xs font-bold text-[#888] mb-2 block">Recipient</label>
                <input
                  type="text"
                  placeholder="Enter client email or search..."
                  value={composeRecipient}
                  onChange={(e) => setComposeRecipient(e.target.value)}
                  className="w-full bg-[#25272E] border border-[#383A42] rounded px-3 py-2 text-sm text-white outline-none"
                />
              </div>

              {/* Channel */}
              <div>
                <label className="text-xs font-bold text-[#888] mb-2 block">Channel</label>
                <div className="flex gap-2">
                  <button
                    onClick={() => setComposeChannel('email')}
                    className={`flex-1 px-4 py-2 rounded text-xs font-bold ${
                      composeChannel === 'email'
                        ? 'bg-[#F5C542] text-black'
                        : 'bg-[#25272E] text-[#888] hover:bg-[#2A2D35]'
                    }`}
                  >
                    Email
                  </button>
                  <button
                    onClick={() => setComposeChannel('sms')}
                    className={`flex-1 px-4 py-2 rounded text-xs font-bold ${
                      composeChannel === 'sms'
                        ? 'bg-[#F5C542] text-black'
                        : 'bg-[#25272E] text-[#888] hover:bg-[#2A2D35]'
                    }`}
                  >
                    SMS
                  </button>
                  <button
                    onClick={() => setComposeChannel('push')}
                    className={`flex-1 px-4 py-2 rounded text-xs font-bold ${
                      composeChannel === 'push'
                        ? 'bg-[#F5C542] text-black'
                        : 'bg-[#25272E] text-[#888] hover:bg-[#2A2D35]'
                    }`}
                  >
                    Push
                  </button>
                </div>
              </div>

              {/* Template */}
              <div>
                <label className="text-xs font-bold text-[#888] mb-2 block">Template (Optional)</label>
                <select
                  value={composeTemplate}
                  onChange={(e) => handleTemplateSelect(e.target.value)}
                  className="w-full bg-[#25272E] border border-[#383A42] rounded px-3 py-2 text-sm text-white outline-none"
                >
                  <option value="">None - Write Custom Message</option>
                  {templates.filter(t => t.channel === composeChannel).map(template => (
                    <option key={template.id} value={template.id}>{template.name}</option>
                  ))}
                </select>
              </div>

              {/* Subject */}
              {composeChannel === 'email' && (
                <div>
                  <label className="text-xs font-bold text-[#888] mb-2 block">Subject</label>
                  <input
                    type="text"
                    placeholder="Enter subject..."
                    value={composeSubject}
                    onChange={(e) => setComposeSubject(e.target.value)}
                    className="w-full bg-[#25272E] border border-[#383A42] rounded px-3 py-2 text-sm text-white outline-none"
                  />
                </div>
              )}

              {/* Body */}
              <div>
                <label className="text-xs font-bold text-[#888] mb-2 block">Message</label>
                <textarea
                  placeholder="Enter your message..."
                  value={composeBody}
                  onChange={(e) => setComposeBody(e.target.value)}
                  rows={8}
                  className="w-full bg-[#25272E] border border-[#383A42] rounded px-3 py-2 text-sm text-white outline-none resize-none"
                />
              </div>
            </div>

            {/* Modal Footer */}
            <div className="p-4 border-t border-[#383A42] flex justify-end gap-2">
              <button
                onClick={() => setShowCompose(false)}
                className="px-4 py-2 bg-[#25272E] text-white text-xs font-bold rounded hover:bg-[#2A2D35] transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSendMessage}
                className="px-4 py-2 bg-[#F5C542] text-black text-xs font-bold rounded hover:bg-[#FFD700] transition-colors flex items-center gap-2"
              >
                <Send size={14} />
                Send Now
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// Mock data generators
function generateMockEntries(): CommunicationEntry[] {
  const types: CommunicationType[] = ['email', 'sms', 'push', 'system', 'support'];
  const statuses: CommunicationStatus[] = ['sent', 'delivered', 'opened', 'bounced', 'failed'];

  const subjects = [
    'Welcome to RTX5 Trading Platform',
    'KYC Document Verification Required',
    'Deposit Confirmation - $5,000',
    'Margin Call Warning - EURUSD Position',
    'Withdrawal Request Approved',
    'Account Verification Complete',
    'New Feature: Copy Trading Now Available',
    'Monthly Account Statement',
    'Important: Platform Maintenance Scheduled',
    'Your Withdrawal is Being Processed',
    'Trade Execution Confirmation',
    'Password Changed Successfully',
    'New Login from Unknown Device',
    'Promotion: 20% Bonus on Next Deposit',
    'Support Ticket #12345 Resolved',
    'Low Balance Alert',
    'Position Closed - Take Profit Hit',
    'Weekly Market Analysis Report',
    'Account Suspended - Action Required',
    'Referral Bonus Credited',
  ];

  const previews = [
    'Thank you for choosing RTX5 Trading Platform. Your account has been successfully created.',
    'Please upload your identity documents to complete account verification.',
    'Your deposit of $5,000 has been successfully processed and credited to your account.',
    'Warning: Your margin level on EURUSD position is below 50%. Please add funds or close positions.',
    'Your withdrawal request for $2,000 has been approved and will be processed within 1-2 business days.',
    'Congratulations! Your account verification is complete. You can now start trading.',
    'Discover our new Copy Trading feature and follow top traders to replicate their strategies.',
    'Your monthly account statement for January 2026 is now available for download.',
    'Platform maintenance scheduled for Sunday 2 AM - 4 AM EST. Trading will be unavailable.',
    'Your withdrawal of $1,500 is currently being processed. Expected completion: 24-48 hours.',
    'Your market order for 1.0 lot EURUSD has been executed at 1.0850.',
    'Your account password was changed successfully on 2026-02-10 at 14:23:15 UTC.',
    'New login detected from IP 192.168.1.1 in New York, USA.',
    'Limited time offer: Get 20% bonus on your next deposit up to $10,000!',
    'Your support ticket regarding deposit inquiry has been resolved by our team.',
    'Your account balance is low ($250). Please deposit funds to maintain positions.',
    'Your GBPUSD position was closed at take profit level. Profit: $450.',
    'Read our latest market analysis covering major currency pairs and commodities.',
    'Your account has been temporarily suspended due to unusual activity. Contact support.',
    'You earned $100 referral bonus for inviting John Doe to RTX5.',
  ];

  const names = [
    'John Smith', 'Emily Johnson', 'Michael Brown', 'Sarah Davis', 'David Wilson',
    'Lisa Anderson', 'Robert Taylor', 'Jennifer Martinez', 'William Garcia', 'Mary Rodriguez',
    'James Miller', 'Patricia Lee', 'Christopher Walker', 'Linda Hall', 'Daniel Young',
    'Nancy King', 'Matthew Wright', 'Sandra Green', 'Joseph Baker', 'Karen Nelson',
  ];

  const entries: CommunicationEntry[] = [];
  const now = new Date();

  for (let i = 0; i < 50; i++) {
    const type = types[Math.floor(Math.random() * types.length)];
    const status = statuses[Math.floor(Math.random() * statuses.length)];
    const clientName = names[i % names.length];
    const sentAt = new Date(now.getTime() - Math.random() * 7 * 24 * 60 * 60 * 1000); // Last 7 days

    let deliveredAt: Date | undefined;
    let openedAt: Date | undefined;

    if (status !== 'sent' && status !== 'failed' && status !== 'bounced') {
      deliveredAt = new Date(sentAt.getTime() + Math.random() * 60 * 60 * 1000); // Within 1 hour

      if (status === 'opened') {
        openedAt = new Date(deliveredAt.getTime() + Math.random() * 24 * 60 * 60 * 1000); // Within 24 hours
      }
    }

    entries.push({
      id: `COM-${String(i + 1).padStart(5, '0')}`,
      type,
      clientId: `CLI-${String((i % 20) + 1).padStart(5, '0')}`,
      clientName,
      clientEmail: clientName.toLowerCase().replace(' ', '.') + '@example.com',
      clientPhone: `+1 (555) ${String(Math.floor(Math.random() * 900) + 100)}-${String(Math.floor(Math.random() * 9000) + 1000)}`,
      accountType: ['Standard', 'Premium', 'VIP'][Math.floor(Math.random() * 3)],
      subject: subjects[i % subjects.length],
      preview: previews[i % previews.length],
      content: previews[i % previews.length] + '\n\nBest regards,\nRTX5 Support Team',
      status,
      sentAt,
      deliveredAt,
      openedAt,
    });
  }

  return entries.sort((a, b) => b.sentAt.getTime() - a.sentAt.getTime());
}

function generateMockTemplates(): MessageTemplate[] {
  return [
    {
      id: 'TPL-001',
      name: 'Welcome Email',
      subject: 'Welcome to RTX5 Trading Platform',
      body: 'Dear Client,\n\nThank you for choosing RTX5 Trading Platform. Your account has been successfully created.\n\nPlease verify your email and complete KYC to start trading.\n\nBest regards,\nRTX5 Team',
      channel: 'email',
    },
    {
      id: 'TPL-002',
      name: 'KYC Reminder',
      subject: 'Complete Your Account Verification',
      body: 'Dear Client,\n\nYour account verification is pending. Please upload the required documents to complete KYC.\n\nBest regards,\nRTX5 Team',
      channel: 'email',
    },
    {
      id: 'TPL-003',
      name: 'Deposit Confirmation',
      subject: 'Deposit Received',
      body: 'Your deposit has been successfully processed and credited to your trading account.\n\nThank you for trading with RTX5!',
      channel: 'email',
    },
    {
      id: 'TPL-004',
      name: 'Margin Call Warning',
      subject: 'Urgent: Margin Call Alert',
      body: 'WARNING: Your margin level is critically low. Please add funds or close positions to avoid liquidation.\n\nContact support if you need assistance.',
      channel: 'email',
    },
    {
      id: 'TPL-005',
      name: 'Withdrawal Approved',
      subject: 'Withdrawal Request Approved',
      body: 'Your withdrawal request has been approved and will be processed within 1-2 business days.\n\nThank you for your patience.',
      channel: 'email',
    },
    {
      id: 'TPL-006',
      name: 'SMS: Deposit Alert',
      subject: '',
      body: 'RTX5: Your deposit of ${{amount}} has been credited. Start trading now!',
      channel: 'sms',
    },
    {
      id: 'TPL-007',
      name: 'SMS: Margin Call',
      subject: '',
      body: 'RTX5 URGENT: Margin call on {{symbol}}. Add funds or close positions immediately.',
      channel: 'sms',
    },
    {
      id: 'TPL-008',
      name: 'SMS: Withdrawal',
      subject: '',
      body: 'RTX5: Your withdrawal of ${{amount}} has been approved. Processing time: 24-48h.',
      channel: 'sms',
    },
    {
      id: 'TPL-009',
      name: 'Push: New Trade',
      subject: '',
      body: 'Your order for {{symbol}} has been executed at {{price}}',
      channel: 'push',
    },
    {
      id: 'TPL-010',
      name: 'Push: Low Balance',
      subject: '',
      body: 'Account balance is low: ${{balance}}. Deposit funds to continue trading.',
      channel: 'push',
    },
  ];
}
