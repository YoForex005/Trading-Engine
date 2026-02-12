'use client';

import React, { useState, useMemo } from 'react';
import { Search, User, Calendar, Phone, Mail, Flag, UserCog, Plus, Clock, AlertCircle, CheckCircle, Filter } from 'lucide-react';

type NoteCategory = 'General' | 'Sales Call' | 'Support Issue' | 'Risk Alert' | 'KYC Follow-up' | 'VIP Service' | 'Compliance';
type NotePriority = 'low' | 'normal' | 'high' | 'urgent';

interface Client {
  id: string;
  name: string;
  email: string;
  accountId: string;
  accountType: string;
  balance: number;
  joinDate: number;
  lastActivity: number;
  assignedManager: string;
}

interface Note {
  id: string;
  clientId: string;
  author: string;
  timestamp: number;
  category: NoteCategory;
  priority: NotePriority;
  text: string;
  hasFollowUp: boolean;
  followUpDate?: number;
  followUpCompleted?: boolean;
}

// Mock data generators
const generateMockClients = (): Client[] => {
  const clients: Client[] = [];
  const accountTypes = ['Standard', 'ECN', 'Pro', 'Islamic', 'VIP'];
  const managers = ['John Smith', 'Sarah Johnson', 'Michael Chen', 'Emily Rodriguez', 'David Kim'];
  const now = Date.now();

  for (let i = 1; i <= 50; i++) {
    clients.push({
      id: `client-${i}`,
      name: `Client ${i}`,
      email: `client${i}@example.com`,
      accountId: `ACC${10000 + i}`,
      accountType: accountTypes[Math.floor(Math.random() * accountTypes.length)],
      balance: Math.random() * 100000,
      joinDate: now - Math.floor(Math.random() * 365 * 86400000),
      lastActivity: now - Math.floor(Math.random() * 90 * 86400000),
      assignedManager: managers[Math.floor(Math.random() * managers.length)]
    });
  }

  return clients;
};

const generateMockNotes = (clients: Client[]): Note[] => {
  const notes: Note[] = [];
  const categories: NoteCategory[] = ['General', 'Sales Call', 'Support Issue', 'Risk Alert', 'KYC Follow-up', 'VIP Service', 'Compliance'];
  const priorities: NotePriority[] = ['low', 'normal', 'high', 'urgent'];
  const authors = ['John Smith', 'Sarah Johnson', 'Michael Chen', 'Emily Rodriguez', 'David Kim', 'Admin User'];

  const noteTexts = [
    'Client expressed interest in upgrading to VIP account',
    'Follow-up call scheduled for next week to discuss trading strategy',
    'Resolved deposit issue - funds credited successfully',
    'Client approaching margin call threshold - monitoring closely',
    'KYC documents received and under review',
    'Premium support requested for time-sensitive trades',
    'Compliance check completed - all documents verified',
    'Client requested information about leverage options',
    'Discussed portfolio diversification strategies',
    'Technical issue with platform - escalated to IT',
    'Client satisfied with recent execution quality',
    'Risk assessment updated based on recent trading activity',
    'VIP concierge service initiated',
    'Regulatory reporting requirements discussed',
    'Client requested copy of account statement',
    'Positive feedback on customer service experience',
    'Margin requirements explained in detail',
    'Islamic account conversion completed',
    'High-frequency trading patterns observed',
    'Client onboarding completed successfully'
  ];

  const now = Date.now();

  // Generate 200+ notes across all clients
  for (let i = 0; i < 220; i++) {
    const client = clients[Math.floor(Math.random() * clients.length)];
    const hasFollowUp = Math.random() > 0.7;
    const daysAgo = Math.floor(Math.random() * 60);
    const timestamp = now - daysAgo * 86400000;

    let followUpDate: number | undefined;
    let followUpCompleted: boolean | undefined;

    if (hasFollowUp) {
      const followUpDays = Math.floor(Math.random() * 30) - 10; // -10 to +20 days from now
      followUpDate = now + followUpDays * 86400000;
      followUpCompleted = followUpDate < now ? Math.random() > 0.3 : false;
    }

    notes.push({
      id: `note-${i + 1}`,
      clientId: client.id,
      author: authors[Math.floor(Math.random() * authors.length)],
      timestamp,
      category: categories[Math.floor(Math.random() * categories.length)],
      priority: priorities[Math.floor(Math.random() * priorities.length)],
      text: noteTexts[Math.floor(Math.random() * noteTexts.length)],
      hasFollowUp,
      followUpDate,
      followUpCompleted
    });
  }

  return notes.sort((a, b) => b.timestamp - a.timestamp);
};

const ClientNotes: React.FC = () => {
  const [clients] = useState<Client[]>(generateMockClients());
  const [notes, setNotes] = useState<Note[]>(() => generateMockNotes(generateMockClients()));
  const [selectedClient, setSelectedClient] = useState<Client | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [showAddNote, setShowAddNote] = useState(false);
  const [filterCategory, setFilterCategory] = useState<NoteCategory | 'All'>('All');

  // Add note form state
  const [noteCategory, setNoteCategory] = useState<NoteCategory>('General');
  const [notePriority, setNotePriority] = useState<NotePriority>('normal');
  const [noteText, setNoteText] = useState('');
  const [hasFollowUp, setHasFollowUp] = useState(false);
  const [followUpDate, setFollowUpDate] = useState('');

  // Search/filter clients
  const filteredClients = useMemo(() => {
    if (!searchTerm) return clients.slice(0, 10);

    const term = searchTerm.toLowerCase();
    return clients.filter(c =>
      c.name.toLowerCase().includes(term) ||
      c.email.toLowerCase().includes(term) ||
      c.accountId.toLowerCase().includes(term)
    ).slice(0, 10);
  }, [clients, searchTerm]);

  // Get notes for selected client
  const clientNotes = useMemo(() => {
    if (!selectedClient) return [];

    let filtered = notes.filter(n => n.clientId === selectedClient.id);

    if (filterCategory !== 'All') {
      filtered = filtered.filter(n => n.category === filterCategory);
    }

    return filtered;
  }, [notes, selectedClient, filterCategory]);

  // Calculate stats
  const stats = useMemo(() => {
    const now = Date.now();
    const todayStart = new Date().setHours(0, 0, 0, 0);

    const notesToday = notes.filter(n => n.timestamp >= todayStart).length;

    const pendingFollowUps = notes.filter(n =>
      n.hasFollowUp && !n.followUpCompleted && n.followUpDate && n.followUpDate >= now
    ).length;

    const overdueFollowUps = notes.filter(n =>
      n.hasFollowUp && !n.followUpCompleted && n.followUpDate && n.followUpDate < now
    ).length;

    const thirtyDaysAgo = now - 30 * 86400000;
    const clientsWithRecentNotes = new Set(notes.filter(n => n.timestamp >= thirtyDaysAgo).map(n => n.clientId));
    const clientsWithoutNotes = clients.length - clientsWithRecentNotes.size;

    return {
      notesToday,
      pendingFollowUps,
      overdueFollowUps,
      clientsWithoutNotes
    };
  }, [notes, clients]);

  // Get all follow-ups
  const allFollowUps = useMemo(() => {
    const now = Date.now();
    return notes
      .filter(n => n.hasFollowUp && !n.followUpCompleted && n.followUpDate)
      .map(n => {
        const client = clients.find(c => c.id === n.clientId);
        const daysUntil = Math.ceil((n.followUpDate! - now) / 86400000);
        let urgency: 'overdue' | 'today' | 'upcoming' = 'upcoming';

        if (daysUntil < 0) urgency = 'overdue';
        else if (daysUntil === 0) urgency = 'today';

        return { ...n, client, daysUntil, urgency };
      })
      .sort((a, b) => a.followUpDate! - b.followUpDate!);
  }, [notes, clients]);

  const handleAddNote = () => {
    if (!selectedClient || !noteText.trim()) return;

    const newNote: Note = {
      id: `note-${Date.now()}`,
      clientId: selectedClient.id,
      author: 'Current User',
      timestamp: Date.now(),
      category: noteCategory,
      priority: notePriority,
      text: noteText,
      hasFollowUp,
      followUpDate: hasFollowUp && followUpDate ? new Date(followUpDate).getTime() : undefined,
      followUpCompleted: false
    };

    setNotes([newNote, ...notes]);

    // Reset form
    setNoteText('');
    setNoteCategory('General');
    setNotePriority('normal');
    setHasFollowUp(false);
    setFollowUpDate('');
    setShowAddNote(false);
  };

  const handleMarkFollowUpComplete = (noteId: string) => {
    setNotes(notes.map(n =>
      n.id === noteId ? { ...n, followUpCompleted: true } : n
    ));
  };

  const getCategoryColor = (category: NoteCategory): string => {
    const colors: Record<NoteCategory, string> = {
      'General': '#6B7280',
      'Sales Call': '#3B82F6',
      'Support Issue': '#F59E0B',
      'Risk Alert': '#EF4444',
      'KYC Follow-up': '#8B5CF6',
      'VIP Service': '#F59E0B',
      'Compliance': '#10B981'
    };
    return colors[category];
  };

  const getPriorityColor = (priority: NotePriority): string => {
    const colors: Record<NotePriority, string> = {
      low: '#6B7280',
      normal: '#3B82F6',
      high: '#F59E0B',
      urgent: '#EF4444'
    };
    return colors[priority];
  };

  const formatDate = (timestamp: number): string => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - timestamp;
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffDays === 0) {
      const hours = date.getHours().toString().padStart(2, '0');
      const minutes = date.getMinutes().toString().padStart(2, '0');
      return `Today at ${hours}:${minutes}`;
    } else if (diffDays === 1) {
      return 'Yesterday';
    } else if (diffDays < 7) {
      return `${diffDays} days ago`;
    } else {
      return date.toLocaleDateString();
    }
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-zinc-300 p-4 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between mb-4 pb-3 border-b border-zinc-700">
        <div className="flex items-center gap-3">
          <User size={20} className="text-[#3B82F6]" />
          <h2 className="text-lg font-semibold text-zinc-100">Client Notes & CRM</h2>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-3 mb-4">
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Notes Today</div>
          <div className="text-2xl font-bold text-[#3B82F6]">{stats.notesToday}</div>
        </div>
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Pending Follow-ups</div>
          <div className="text-2xl font-bold text-[#F59E0B]">{stats.pendingFollowUps}</div>
        </div>
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Overdue Items</div>
          <div className="text-2xl font-bold text-[#EF4444]">{stats.overdueFollowUps}</div>
        </div>
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Clients Without Notes (30d)</div>
          <div className="text-2xl font-bold text-zinc-500">{stats.clientsWithoutNotes}</div>
        </div>
      </div>

      <div className="flex-1 grid grid-cols-12 gap-4 overflow-hidden">
        {/* Left Panel - Client Search & Selection */}
        <div className="col-span-3 flex flex-col bg-[#1E2026] border border-zinc-700 rounded overflow-hidden">
          <div className="p-3 border-b border-zinc-700">
            <div className="relative">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search clients..."
                className="w-full bg-[#121316] border border-zinc-700 rounded pl-10 pr-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
              />
            </div>
          </div>

          <div className="flex-1 overflow-auto">
            {filteredClients.map(client => (
              <div
                key={client.id}
                onClick={() => setSelectedClient(client)}
                className={`p-3 border-b border-zinc-800 cursor-pointer transition-colors ${
                  selectedClient?.id === client.id
                    ? 'bg-[#3B82F6]/20 border-l-4 border-l-[#3B82F6]'
                    : 'hover:bg-[#2A2D35]'
                }`}
              >
                <div className="text-sm font-semibold text-zinc-100">{client.name}</div>
                <div className="text-xs text-zinc-500">{client.accountId}</div>
                <div className="text-xs text-zinc-400 mt-1">{client.email}</div>
              </div>
            ))}
            {filteredClients.length === 0 && (
              <div className="p-4 text-center text-zinc-500 text-sm">
                No clients found
              </div>
            )}
          </div>
        </div>

        {/* Middle Panel - Client Profile & Notes */}
        <div className="col-span-6 flex flex-col overflow-hidden">
          {selectedClient ? (
            <>
              {/* Client Profile Card */}
              <div className="bg-[#1E2026] border border-zinc-700 rounded p-4 mb-4">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h3 className="text-lg font-semibold text-zinc-100">{selectedClient.name}</h3>
                    <div className="text-sm text-zinc-400">{selectedClient.email}</div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-zinc-500">Account ID</div>
                    <div className="text-sm font-semibold text-zinc-100">{selectedClient.accountId}</div>
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-4 mb-3">
                  <div>
                    <div className="text-xs text-zinc-500">Account Type</div>
                    <div className="text-sm text-zinc-100">{selectedClient.accountType}</div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500">Balance</div>
                    <div className="text-sm text-zinc-100">${selectedClient.balance.toFixed(2)}</div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500">Join Date</div>
                    <div className="text-sm text-zinc-100">{new Date(selectedClient.joinDate).toLocaleDateString()}</div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div>
                    <div className="text-xs text-zinc-500">Last Activity</div>
                    <div className="text-sm text-zinc-100">{formatDate(selectedClient.lastActivity)}</div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500">Assigned Manager</div>
                    <div className="text-sm text-zinc-100">{selectedClient.assignedManager}</div>
                  </div>
                </div>

                {/* Quick Actions */}
                <div className="flex gap-2">
                  <button className="flex-1 px-3 py-2 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs rounded transition-colors flex items-center justify-center gap-1.5">
                    <Phone size={12} />
                    Schedule Call
                  </button>
                  <button className="flex-1 px-3 py-2 bg-[#10B981] hover:bg-[#059669] text-white text-xs rounded transition-colors flex items-center justify-center gap-1.5">
                    <Mail size={12} />
                    Send Email
                  </button>
                  <button className="flex-1 px-3 py-2 bg-[#F59E0B] hover:bg-[#D97706] text-white text-xs rounded transition-colors flex items-center justify-center gap-1.5">
                    <Flag size={12} />
                    Flag for Review
                  </button>
                  <button className="flex-1 px-3 py-2 bg-[#8B5CF6] hover:bg-[#7C3AED] text-white text-xs rounded transition-colors flex items-center justify-center gap-1.5">
                    <UserCog size={12} />
                    Assign Manager
                  </button>
                </div>
              </div>

              {/* Add Note Button */}
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <h3 className="text-sm font-semibold text-zinc-100">Notes Timeline</h3>
                  <div className="text-xs text-zinc-500">({clientNotes.length} notes)</div>
                </div>
                <div className="flex items-center gap-2">
                  <select
                    value={filterCategory}
                    onChange={(e) => setFilterCategory(e.target.value as NoteCategory | 'All')}
                    className="bg-[#1E2026] border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300"
                  >
                    <option value="All">All Categories</option>
                    <option value="General">General</option>
                    <option value="Sales Call">Sales Call</option>
                    <option value="Support Issue">Support Issue</option>
                    <option value="Risk Alert">Risk Alert</option>
                    <option value="KYC Follow-up">KYC Follow-up</option>
                    <option value="VIP Service">VIP Service</option>
                    <option value="Compliance">Compliance</option>
                  </select>
                  <button
                    onClick={() => setShowAddNote(!showAddNote)}
                    className="px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs rounded transition-colors flex items-center gap-1.5"
                  >
                    <Plus size={14} />
                    Add Note
                  </button>
                </div>
              </div>

              {/* Add Note Form */}
              {showAddNote && (
                <div className="bg-[#1E2026] border border-zinc-700 rounded p-4 mb-3">
                  <div className="grid grid-cols-2 gap-3 mb-3">
                    <div>
                      <label className="block text-xs text-zinc-400 mb-1">Category</label>
                      <select
                        value={noteCategory}
                        onChange={(e) => setNoteCategory(e.target.value as NoteCategory)}
                        className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100"
                      >
                        <option value="General">General</option>
                        <option value="Sales Call">Sales Call</option>
                        <option value="Support Issue">Support Issue</option>
                        <option value="Risk Alert">Risk Alert</option>
                        <option value="KYC Follow-up">KYC Follow-up</option>
                        <option value="VIP Service">VIP Service</option>
                        <option value="Compliance">Compliance</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-xs text-zinc-400 mb-1">Priority</label>
                      <select
                        value={notePriority}
                        onChange={(e) => setNotePriority(e.target.value as NotePriority)}
                        className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100"
                      >
                        <option value="low">Low</option>
                        <option value="normal">Normal</option>
                        <option value="high">High</option>
                        <option value="urgent">Urgent</option>
                      </select>
                    </div>
                  </div>

                  <div className="mb-3">
                    <label className="block text-xs text-zinc-400 mb-1">Note</label>
                    <textarea
                      value={noteText}
                      onChange={(e) => setNoteText(e.target.value)}
                      className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 resize-none focus:outline-none focus:border-[#3B82F6]"
                      rows={3}
                      placeholder="Enter note details..."
                    />
                  </div>

                  <div className="flex items-center gap-4 mb-3">
                    <label className="flex items-center gap-2 text-sm text-zinc-300">
                      <input
                        type="checkbox"
                        checked={hasFollowUp}
                        onChange={(e) => setHasFollowUp(e.target.checked)}
                        className="rounded border-zinc-600 bg-[#121316]"
                      />
                      Add follow-up reminder
                    </label>
                    {hasFollowUp && (
                      <input
                        type="date"
                        value={followUpDate}
                        onChange={(e) => setFollowUpDate(e.target.value)}
                        className="bg-[#121316] border border-zinc-700 rounded px-3 py-1.5 text-sm text-zinc-100"
                      />
                    )}
                  </div>

                  <div className="flex justify-end gap-2">
                    <button
                      onClick={() => setShowAddNote(false)}
                      className="px-4 py-2 bg-[#2A2D35] hover:bg-[#35383F] text-zinc-300 text-sm rounded transition-colors"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={handleAddNote}
                      className="px-4 py-2 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-sm rounded transition-colors"
                    >
                      Add Note
                    </button>
                  </div>
                </div>
              )}

              {/* Notes Timeline */}
              <div className="flex-1 overflow-auto space-y-3">
                {clientNotes.map(note => (
                  <div key={note.id} className="bg-[#1E2026] border border-zinc-700 rounded p-3">
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <span
                          className="px-2 py-0.5 rounded text-xs font-semibold"
                          style={{ backgroundColor: getCategoryColor(note.category) + '33', color: getCategoryColor(note.category) }}
                        >
                          {note.category}
                        </span>
                        <span
                          className="px-2 py-0.5 rounded text-xs font-semibold"
                          style={{ backgroundColor: getPriorityColor(note.priority) + '33', color: getPriorityColor(note.priority) }}
                        >
                          {note.priority}
                        </span>
                      </div>
                      <div className="text-xs text-zinc-500">{formatDate(note.timestamp)}</div>
                    </div>

                    <p className="text-sm text-zinc-300 mb-2">{note.text}</p>

                    <div className="flex items-center justify-between">
                      <div className="text-xs text-zinc-500">
                        by <span className="text-zinc-400">{note.author}</span>
                      </div>
                      {note.hasFollowUp && note.followUpDate && (
                        <div className="flex items-center gap-2">
                          {note.followUpCompleted ? (
                            <span className="text-xs text-[#10B981] flex items-center gap-1">
                              <CheckCircle size={12} />
                              Follow-up completed
                            </span>
                          ) : (
                            <>
                              <span className="text-xs text-zinc-400 flex items-center gap-1">
                                <Calendar size={12} />
                                Follow-up: {new Date(note.followUpDate).toLocaleDateString()}
                              </span>
                              <button
                                onClick={() => handleMarkFollowUpComplete(note.id)}
                                className="text-xs text-[#3B82F6] hover:text-[#2563EB]"
                              >
                                Mark Complete
                              </button>
                            </>
                          )}
                        </div>
                      )}
                    </div>
                  </div>
                ))}

                {clientNotes.length === 0 && (
                  <div className="text-center py-12 text-zinc-500">
                    No notes for this client
                  </div>
                )}
              </div>
            </>
          ) : (
            <div className="flex items-center justify-center h-full text-zinc-500">
              Select a client to view notes
            </div>
          )}
        </div>

        {/* Right Panel - Follow-up Reminders */}
        <div className="col-span-3 bg-[#1E2026] border border-zinc-700 rounded p-4 overflow-auto">
          <h3 className="text-sm font-semibold text-zinc-100 mb-3 flex items-center gap-2">
            <Clock size={16} className="text-[#F59E0B]" />
            Follow-up Reminders
          </h3>

          <div className="space-y-2">
            {allFollowUps.slice(0, 20).map(followUp => {
              const urgencyColors = {
                overdue: '#EF4444',
                today: '#F59E0B',
                upcoming: '#10B981'
              };

              return (
                <div
                  key={followUp.id}
                  className="p-3 rounded border"
                  style={{
                    backgroundColor: urgencyColors[followUp.urgency] + '11',
                    borderColor: urgencyColors[followUp.urgency] + '33'
                  }}
                >
                  <div className="flex items-start justify-between mb-1">
                    <div className="text-sm font-semibold text-zinc-100">
                      {followUp.client?.name}
                    </div>
                    <span
                      className="px-2 py-0.5 rounded text-xs font-semibold"
                      style={{ backgroundColor: urgencyColors[followUp.urgency] + '33', color: urgencyColors[followUp.urgency] }}
                    >
                      {followUp.urgency === 'overdue' ? `${Math.abs(followUp.daysUntil)}d overdue` :
                       followUp.urgency === 'today' ? 'Today' :
                       `In ${followUp.daysUntil}d`}
                    </span>
                  </div>

                  <div className="text-xs text-zinc-400 mb-2">{followUp.client?.accountId}</div>

                  <div className="text-xs text-zinc-300 mb-2 line-clamp-2">{followUp.text}</div>

                  <div className="flex items-center justify-between">
                    <span
                      className="px-2 py-0.5 rounded text-xs"
                      style={{ backgroundColor: getCategoryColor(followUp.category) + '33', color: getCategoryColor(followUp.category) }}
                    >
                      {followUp.category}
                    </span>
                    <button
                      onClick={() => {
                        setSelectedClient(followUp.client!);
                        handleMarkFollowUpComplete(followUp.id);
                      }}
                      className="text-xs text-[#3B82F6] hover:text-[#2563EB]"
                    >
                      Complete
                    </button>
                  </div>
                </div>
              );
            })}

            {allFollowUps.length === 0 && (
              <div className="text-center py-8 text-zinc-500 text-sm">
                No pending follow-ups
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ClientNotes;
