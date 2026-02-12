'use client';

import React, { useState, useMemo } from 'react';
import {
    MessageSquare,
    Clock,
    CheckCircle,
    AlertTriangle,
    Search,
    Filter,
    X,
    Send,
    Paperclip,
    User,
    Calendar,
    Download,
    Edit3,
    Lock
} from 'lucide-react';

// Types
type TicketPriority = 'critical' | 'high' | 'medium' | 'low';
type TicketCategory = 'account' | 'trading' | 'deposit' | 'withdrawal' | 'technical' | 'other';
type TicketStatus = 'open' | 'in_progress' | 'waiting' | 'resolved' | 'closed';

interface TicketMessage {
    id: string;
    sender: 'client' | 'admin';
    senderName: string;
    content: string;
    timestamp: Date;
    isInternal?: boolean; // Internal notes not visible to client
}

interface Ticket {
    id: string;
    subject: string;
    clientName: string;
    clientId: string;
    priority: TicketPriority;
    category: TicketCategory;
    status: TicketStatus;
    assignedTo: string | null;
    createdAt: Date;
    lastUpdated: Date;
    messages: TicketMessage[];
    attachments?: string[];
    slaBreached: boolean;
    slaApproaching: boolean;
}

const PRIORITY_COLORS: Record<TicketPriority, { bg: string; text: string; label: string }> = {
    critical: { bg: 'bg-red-500/20', text: 'text-red-400', label: 'Critical' },
    high: { bg: 'bg-orange-500/20', text: 'text-orange-400', label: 'High' },
    medium: { bg: 'bg-yellow-500/20', text: 'text-yellow-400', label: 'Medium' },
    low: { bg: 'bg-blue-500/20', text: 'text-blue-400', label: 'Low' }
};

const STATUS_COLORS: Record<TicketStatus, { bg: string; text: string; label: string }> = {
    open: { bg: 'bg-blue-500/20', text: 'text-blue-400', label: 'Open' },
    in_progress: { bg: 'bg-purple-500/20', text: 'text-purple-400', label: 'In Progress' },
    waiting: { bg: 'bg-yellow-500/20', text: 'text-yellow-400', label: 'Waiting' },
    resolved: { bg: 'bg-emerald-500/20', text: 'text-emerald-400', label: 'Resolved' },
    closed: { bg: 'bg-zinc-500/20', text: 'text-zinc-400', label: 'Closed' }
};

const CATEGORY_LABELS: Record<TicketCategory, string> = {
    account: 'Account',
    trading: 'Trading',
    deposit: 'Deposit',
    withdrawal: 'Withdrawal',
    technical: 'Technical',
    other: 'Other'
};

const CANNED_RESPONSES = [
    { id: '1', label: 'Deposit Received', content: 'Thank you for your deposit. Your account has been credited and you can start trading immediately.' },
    { id: '2', label: 'KYC Required', content: 'To process your withdrawal, we need to verify your identity. Please upload a government-issued ID and proof of address in the Client Portal.' },
    { id: '3', label: 'Issue Escalated', content: 'Your issue has been escalated to our specialist team. They will review your case and respond within 24 hours.' },
    { id: '4', label: 'Withdrawal Processing', content: 'Your withdrawal request is being processed. Funds will be transferred to your account within 2-3 business days.' },
    { id: '5', label: 'Technical Support', content: 'Our technical team is investigating this issue. We will update you as soon as we have more information.' },
    { id: '6', label: 'Account Verified', content: 'Your account has been successfully verified. You now have full access to all features.' }
];

export default function SupportTickets() {
    const [selectedTicket, setSelectedTicket] = useState<Ticket | null>(null);
    const [filterStatus, setFilterStatus] = useState<TicketStatus | 'all'>('all');
    const [filterPriority, setFilterPriority] = useState<TicketPriority | 'all'>('all');
    const [filterCategory, setFilterCategory] = useState<TicketCategory | 'all'>('all');
    const [filterAssigned, setFilterAssigned] = useState<string>('all');
    const [searchQuery, setSearchQuery] = useState('');
    const [replyText, setReplyText] = useState('');
    const [isInternalNote, setIsInternalNote] = useState(false);

    // Generate 25+ mock tickets
    const mockTickets: Ticket[] = useMemo(() => {
        const subjects = [
            'Cannot withdraw funds',
            'Deposit not reflected in account',
            'Platform login issues',
            'Margin call notification error',
            'Commission discrepancy',
            'KYC document upload failed',
            'API key not working',
            'Chart data missing for EURUSD',
            'Withdrawal taking too long',
            'Account leverage change request',
            'Spread seems too wide',
            'Stop loss not triggered',
            'Multiple positions opened by mistake',
            'Password reset not working',
            'Email notifications not received',
            'Mobile app crashing',
            'Trade history export issue',
            'Swap charges too high',
            'Account verification status',
            'Bonus credit not applied',
            'Demo account balance reset request',
            'Copy trading not working',
            'Webhook configuration help',
            'Symbol not available for trading',
            'Slippage on market order',
            'Two-factor authentication issue',
            'Account statement request',
            'Referral commission missing'
        ];

        const clients = [
            { name: 'John Smith', id: 'CLI-1001' },
            { name: 'Emma Wilson', id: 'CLI-1002' },
            { name: 'Michael Chen', id: 'CLI-1003' },
            { name: 'Sarah Johnson', id: 'CLI-1004' },
            { name: 'David Lee', id: 'CLI-1005' },
            { name: 'Lisa Anderson', id: 'CLI-1006' },
            { name: 'Robert Brown', id: 'CLI-1007' },
            { name: 'Jennifer Davis', id: 'CLI-1008' },
            { name: 'William Garcia', id: 'CLI-1009' },
            { name: 'Mary Martinez', id: 'CLI-1010' }
        ];

        const admins = ['Admin-Alice', 'Admin-Bob', 'Admin-Charlie', null]; // null = unassigned
        const priorities: TicketPriority[] = ['critical', 'high', 'medium', 'low'];
        const categories: TicketCategory[] = ['account', 'trading', 'deposit', 'withdrawal', 'technical', 'other'];
        const statuses: TicketStatus[] = ['open', 'in_progress', 'waiting', 'resolved', 'closed'];

        return subjects.map((subject, idx) => {
            const client = clients[idx % clients.length];
            const priority = priorities[idx % priorities.length];
            const category = categories[idx % categories.length];
            const status = statuses[idx % statuses.length];
            const assignedTo = admins[idx % admins.length];
            const createdAt = new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000); // Last 7 days
            const lastUpdated = new Date(createdAt.getTime() + Math.random() * 24 * 60 * 60 * 1000);

            // SLA logic: Critical = 2h, High = 4h, Medium = 8h, Low = 24h
            const slaMinutes = priority === 'critical' ? 120 : priority === 'high' ? 240 : priority === 'medium' ? 480 : 1440;
            const minutesSinceCreated = (Date.now() - createdAt.getTime()) / (1000 * 60);
            const slaBreached = status === 'open' && minutesSinceCreated > slaMinutes;
            const slaApproaching = status === 'open' && minutesSinceCreated > (slaMinutes * 0.8) && !slaBreached;

            // Generate conversation thread
            const messages: TicketMessage[] = [
                {
                    id: `msg-${idx}-1`,
                    sender: 'client',
                    senderName: client.name,
                    content: `Hello, I'm having an issue: ${subject}. Can you please help?`,
                    timestamp: createdAt
                }
            ];

            if (status !== 'open') {
                messages.push({
                    id: `msg-${idx}-2`,
                    sender: 'admin',
                    senderName: assignedTo || 'Support Team',
                    content: 'Thank you for contacting us. We are looking into this issue and will get back to you shortly.',
                    timestamp: new Date(createdAt.getTime() + 30 * 60 * 1000)
                });
            }

            if (status === 'in_progress' || status === 'waiting' || status === 'resolved' || status === 'closed') {
                messages.push({
                    id: `msg-${idx}-3`,
                    sender: 'admin',
                    senderName: assignedTo || 'Support Team',
                    content: 'Investigating the issue...',
                    timestamp: new Date(createdAt.getTime() + 90 * 60 * 1000),
                    isInternal: true
                });
            }

            if (status === 'resolved' || status === 'closed') {
                messages.push({
                    id: `msg-${idx}-4`,
                    sender: 'admin',
                    senderName: assignedTo || 'Support Team',
                    content: 'The issue has been resolved. Please check your account and let us know if you need any further assistance.',
                    timestamp: new Date(createdAt.getTime() + 4 * 60 * 60 * 1000)
                });
            }

            return {
                id: `TKT-${10000 + idx}`,
                subject,
                clientName: client.name,
                clientId: client.id,
                priority,
                category,
                status,
                assignedTo,
                createdAt,
                lastUpdated,
                messages,
                attachments: idx % 5 === 0 ? ['screenshot.png', 'error_log.txt'] : undefined,
                slaBreached,
                slaApproaching
            };
        });
    }, []);

    // Filter tickets
    const filteredTickets = useMemo(() => {
        return mockTickets.filter(ticket => {
            if (filterStatus !== 'all' && ticket.status !== filterStatus) return false;
            if (filterPriority !== 'all' && ticket.priority !== filterPriority) return false;
            if (filterCategory !== 'all' && ticket.category !== filterCategory) return false;
            if (filterAssigned !== 'all' && ticket.assignedTo !== filterAssigned) return false;
            if (searchQuery && !ticket.subject.toLowerCase().includes(searchQuery.toLowerCase()) &&
                !ticket.clientName.toLowerCase().includes(searchQuery.toLowerCase())) return false;
            return true;
        });
    }, [mockTickets, filterStatus, filterPriority, filterCategory, filterAssigned, searchQuery]);

    // Calculate summary metrics
    const summaryMetrics = useMemo(() => {
        const openTickets = mockTickets.filter(t => t.status === 'open' || t.status === 'in_progress' || t.status === 'waiting').length;
        const resolvedToday = mockTickets.filter(t => {
            const today = new Date();
            today.setHours(0, 0, 0, 0);
            return (t.status === 'resolved' || t.status === 'closed') && t.lastUpdated >= today;
        }).length;

        // Calculate average response time (simplified: time from creation to first admin reply)
        const responseTimes = mockTickets
            .filter(t => t.messages.length > 1 && t.messages[1].sender === 'admin')
            .map(t => {
                const created = t.createdAt.getTime();
                const firstReply = t.messages[1].timestamp.getTime();
                return (firstReply - created) / (1000 * 60); // minutes
            });
        const avgResponseTime = responseTimes.length > 0
            ? responseTimes.reduce((sum, t) => sum + t, 0) / responseTimes.length
            : 0;

        // SLA compliance
        const totalSLATickets = mockTickets.filter(t => t.status !== 'open').length;
        const breachedTickets = mockTickets.filter(t => t.slaBreached).length;
        const slaCompliance = totalSLATickets > 0 ? ((totalSLATickets - breachedTickets) / totalSLATickets) * 100 : 100;

        return {
            openTickets,
            avgResponseTime,
            resolvedToday,
            slaCompliance
        };
    }, [mockTickets]);

    // Get unique admins for filter
    const allAdmins = useMemo(() => {
        const admins = new Set(mockTickets.map(t => t.assignedTo).filter(Boolean) as string[]);
        return ['all', ...Array.from(admins)];
    }, [mockTickets]);

    const handleSendReply = () => {
        if (!selectedTicket || !replyText.trim()) return;

        // In real implementation, this would send the reply via API
        console.log('Sending reply:', {
            ticketId: selectedTicket.id,
            content: replyText,
            isInternal: isInternalNote
        });

        // Reset form
        setReplyText('');
        setIsInternalNote(false);
    };

    const handleCannedResponse = (content: string) => {
        setReplyText(content);
    };

    const formatTime = (date: Date) => {
        const now = new Date();
        const diff = (now.getTime() - date.getTime()) / 1000 / 60; // minutes

        if (diff < 60) return `${Math.floor(diff)}m ago`;
        if (diff < 1440) return `${Math.floor(diff / 60)}h ago`;
        return `${Math.floor(diff / 1440)}d ago`;
    };

    const formatDateTime = (date: Date) => {
        return date.toLocaleString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    return (
        <div className="h-full overflow-auto bg-[#121316] p-4">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-xl font-bold text-zinc-100 flex items-center gap-2">
                        <MessageSquare size={20} className="text-blue-400" />
                        Support Tickets
                    </h1>
                    <p className="text-xs text-zinc-500 mt-1">Customer support ticket management system</p>
                </div>
                <button className="flex items-center gap-2 px-4 py-2 bg-[#27272a] hover:bg-[#3f3f46] text-zinc-300 rounded text-xs transition-colors">
                    <Download size={14} />
                    Export Report
                </button>
            </div>

            {/* Summary Cards */}
            <div className="grid grid-cols-4 gap-4 mb-6">
                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] text-zinc-500 uppercase font-semibold">Open Tickets</span>
                        <MessageSquare size={14} className="text-blue-400" />
                    </div>
                    <div className="text-2xl font-bold text-zinc-100">{summaryMetrics.openTickets}</div>
                    <div className="text-[10px] text-zinc-500 mt-1">Require attention</div>
                </div>

                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] text-zinc-500 uppercase font-semibold">Avg Response Time</span>
                        <Clock size={14} className="text-yellow-400" />
                    </div>
                    <div className="text-2xl font-bold text-zinc-100">{summaryMetrics.avgResponseTime.toFixed(0)}<span className="text-sm text-zinc-500 ml-1">min</span></div>
                    <div className="text-[10px] text-emerald-400 mt-1">Within SLA target</div>
                </div>

                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] text-zinc-500 uppercase font-semibold">Resolved Today</span>
                        <CheckCircle size={14} className="text-emerald-400" />
                    </div>
                    <div className="text-2xl font-bold text-zinc-100">{summaryMetrics.resolvedToday}</div>
                    <div className="text-[10px] text-zinc-500 mt-1">Tickets closed</div>
                </div>

                <div className="bg-[#18181b] p-4 rounded-lg border border-[#27272a]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-[10px] text-zinc-500 uppercase font-semibold">SLA Compliance</span>
                        <AlertTriangle size={14} className={summaryMetrics.slaCompliance >= 95 ? "text-emerald-400" : "text-orange-400"} />
                    </div>
                    <div className="text-2xl font-bold text-zinc-100">{summaryMetrics.slaCompliance.toFixed(1)}<span className="text-sm text-zinc-500 ml-1">%</span></div>
                    <div className={`text-[10px] mt-1 ${summaryMetrics.slaCompliance >= 95 ? 'text-emerald-400' : 'text-orange-400'}`}>
                        {summaryMetrics.slaCompliance >= 95 ? 'Excellent' : 'Needs improvement'}
                    </div>
                </div>
            </div>

            {/* Filters */}
            <div className="grid grid-cols-6 gap-3 mb-6 p-4 bg-[#18181b] rounded-lg border border-[#27272a]">
                <div>
                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Status</label>
                    <select
                        value={filterStatus}
                        onChange={(e) => setFilterStatus(e.target.value as any)}
                        className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500"
                    >
                        <option value="all">All Statuses</option>
                        <option value="open">Open</option>
                        <option value="in_progress">In Progress</option>
                        <option value="waiting">Waiting</option>
                        <option value="resolved">Resolved</option>
                        <option value="closed">Closed</option>
                    </select>
                </div>

                <div>
                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Priority</label>
                    <select
                        value={filterPriority}
                        onChange={(e) => setFilterPriority(e.target.value as any)}
                        className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500"
                    >
                        <option value="all">All Priorities</option>
                        <option value="critical">Critical</option>
                        <option value="high">High</option>
                        <option value="medium">Medium</option>
                        <option value="low">Low</option>
                    </select>
                </div>

                <div>
                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Category</label>
                    <select
                        value={filterCategory}
                        onChange={(e) => setFilterCategory(e.target.value as any)}
                        className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500"
                    >
                        <option value="all">All Categories</option>
                        <option value="account">Account</option>
                        <option value="trading">Trading</option>
                        <option value="deposit">Deposit</option>
                        <option value="withdrawal">Withdrawal</option>
                        <option value="technical">Technical</option>
                        <option value="other">Other</option>
                    </select>
                </div>

                <div>
                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Assigned To</label>
                    <select
                        value={filterAssigned}
                        onChange={(e) => setFilterAssigned(e.target.value)}
                        className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500"
                    >
                        {allAdmins.map(admin => (
                            <option key={admin} value={admin}>{admin === 'all' ? 'All Admins' : admin === 'null' ? 'Unassigned' : admin}</option>
                        ))}
                    </select>
                </div>

                <div className="col-span-2">
                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Search</label>
                    <div className="relative">
                        <Search size={14} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-zinc-500" />
                        <input
                            type="text"
                            placeholder="Search by subject or client..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full pl-9 pr-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500"
                        />
                    </div>
                </div>
            </div>

            {/* Ticket List Table */}
            <div className="bg-[#18181b] rounded-lg border border-[#27272a] overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="w-full text-xs">
                        <thead className="bg-[#27272a] text-zinc-400 uppercase text-[10px]">
                            <tr>
                                <th className="text-left p-3 font-semibold">ID</th>
                                <th className="text-left p-3 font-semibold">Subject</th>
                                <th className="text-left p-3 font-semibold">Client</th>
                                <th className="text-center p-3 font-semibold">Priority</th>
                                <th className="text-center p-3 font-semibold">Category</th>
                                <th className="text-center p-3 font-semibold">Status</th>
                                <th className="text-left p-3 font-semibold">Assigned To</th>
                                <th className="text-right p-3 font-semibold">Created</th>
                                <th className="text-right p-3 font-semibold">Last Updated</th>
                            </tr>
                        </thead>
                        <tbody className="text-zinc-300">
                            {filteredTickets.map((ticket) => {
                                const rowClass = ticket.slaBreached
                                    ? 'bg-red-500/10 border-l-4 border-red-500'
                                    : ticket.slaApproaching
                                    ? 'bg-yellow-500/10 border-l-4 border-yellow-500'
                                    : '';

                                return (
                                    <tr
                                        key={ticket.id}
                                        className={`border-t border-[#27272a] hover:bg-[#27272a]/50 cursor-pointer ${rowClass}`}
                                        onClick={() => setSelectedTicket(ticket)}
                                    >
                                        <td className="p-3">
                                            <span className="font-medium text-blue-400">{ticket.id}</span>
                                        </td>
                                        <td className="p-3">
                                            <div className="max-w-xs truncate font-medium text-zinc-100">{ticket.subject}</div>
                                        </td>
                                        <td className="p-3">
                                            <div className="flex items-center gap-2">
                                                <User size={12} className="text-zinc-500" />
                                                <span>{ticket.clientName}</span>
                                            </div>
                                        </td>
                                        <td className="p-3 text-center">
                                            <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${PRIORITY_COLORS[ticket.priority].bg} ${PRIORITY_COLORS[ticket.priority].text}`}>
                                                {PRIORITY_COLORS[ticket.priority].label}
                                            </span>
                                        </td>
                                        <td className="p-3 text-center">
                                            <span className="text-zinc-400">{CATEGORY_LABELS[ticket.category]}</span>
                                        </td>
                                        <td className="p-3 text-center">
                                            <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${STATUS_COLORS[ticket.status].bg} ${STATUS_COLORS[ticket.status].text}`}>
                                                {STATUS_COLORS[ticket.status].label}
                                            </span>
                                        </td>
                                        <td className="p-3">
                                            {ticket.assignedTo ? (
                                                <span className="text-zinc-300">{ticket.assignedTo}</span>
                                            ) : (
                                                <span className="text-zinc-500 italic">Unassigned</span>
                                            )}
                                        </td>
                                        <td className="p-3 text-right text-zinc-400">{formatTime(ticket.createdAt)}</td>
                                        <td className="p-3 text-right text-zinc-400">{formatTime(ticket.lastUpdated)}</td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Ticket Detail Modal */}
            {selectedTicket && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-[#18181b] rounded-lg border border-[#27272a] w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
                        {/* Modal Header */}
                        <div className="p-4 border-b border-[#27272a] flex items-center justify-between">
                            <div>
                                <div className="flex items-center gap-3">
                                    <h2 className="text-lg font-bold text-zinc-100">{selectedTicket.id}</h2>
                                    <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${PRIORITY_COLORS[selectedTicket.priority].bg} ${PRIORITY_COLORS[selectedTicket.priority].text}`}>
                                        {PRIORITY_COLORS[selectedTicket.priority].label}
                                    </span>
                                    <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${STATUS_COLORS[selectedTicket.status].bg} ${STATUS_COLORS[selectedTicket.status].text}`}>
                                        {STATUS_COLORS[selectedTicket.status].label}
                                    </span>
                                </div>
                                <p className="text-sm text-zinc-300 mt-1">{selectedTicket.subject}</p>
                            </div>
                            <button
                                onClick={() => setSelectedTicket(null)}
                                className="p-2 hover:bg-[#27272a] rounded transition-colors"
                            >
                                <X size={18} className="text-zinc-400" />
                            </button>
                        </div>

                        {/* Modal Body */}
                        <div className="flex-1 overflow-auto p-4">
                            {/* Client Info */}
                            <div className="grid grid-cols-3 gap-4 mb-4 p-3 bg-[#27272a] rounded">
                                <div>
                                    <div className="text-[10px] text-zinc-500 uppercase font-semibold mb-1">Client</div>
                                    <div className="text-sm text-zinc-100">{selectedTicket.clientName} ({selectedTicket.clientId})</div>
                                </div>
                                <div>
                                    <div className="text-[10px] text-zinc-500 uppercase font-semibold mb-1">Category</div>
                                    <div className="text-sm text-zinc-100">{CATEGORY_LABELS[selectedTicket.category]}</div>
                                </div>
                                <div>
                                    <div className="text-[10px] text-zinc-500 uppercase font-semibold mb-1">Created</div>
                                    <div className="text-sm text-zinc-100">{formatDateTime(selectedTicket.createdAt)}</div>
                                </div>
                            </div>

                            {/* Controls */}
                            <div className="grid grid-cols-3 gap-3 mb-4">
                                <div>
                                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Status</label>
                                    <select className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs">
                                        <option value="open">Open</option>
                                        <option value="in_progress">In Progress</option>
                                        <option value="waiting">Waiting</option>
                                        <option value="resolved">Resolved</option>
                                        <option value="closed">Closed</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Priority</label>
                                    <select className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs">
                                        <option value="critical">Critical</option>
                                        <option value="high">High</option>
                                        <option value="medium">Medium</option>
                                        <option value="low">Low</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Assign To</label>
                                    <select className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs">
                                        <option value="">Unassigned</option>
                                        <option value="Admin-Alice">Admin-Alice</option>
                                        <option value="Admin-Bob">Admin-Bob</option>
                                        <option value="Admin-Charlie">Admin-Charlie</option>
                                    </select>
                                </div>
                            </div>

                            {/* Conversation Thread */}
                            <div className="mb-4">
                                <h3 className="text-sm font-semibold text-zinc-100 mb-3">Conversation</h3>
                                <div className="space-y-3">
                                    {selectedTicket.messages.map((msg) => (
                                        <div
                                            key={msg.id}
                                            className={`p-3 rounded ${
                                                msg.isInternal
                                                    ? 'bg-purple-500/10 border border-purple-500/30'
                                                    : msg.sender === 'client'
                                                    ? 'bg-[#27272a]'
                                                    : 'bg-blue-500/10 border border-blue-500/30'
                                            }`}
                                        >
                                            <div className="flex items-center justify-between mb-2">
                                                <div className="flex items-center gap-2">
                                                    <span className="text-xs font-semibold text-zinc-100">{msg.senderName}</span>
                                                    <span className={`text-[10px] px-1.5 py-0.5 rounded ${
                                                        msg.isInternal
                                                            ? 'bg-purple-500/20 text-purple-400'
                                                            : msg.sender === 'client'
                                                            ? 'bg-zinc-700 text-zinc-400'
                                                            : 'bg-blue-500/20 text-blue-400'
                                                    }`}>
                                                        {msg.isInternal ? 'Internal Note' : msg.sender === 'client' ? 'Client' : 'Admin'}
                                                    </span>
                                                    {msg.isInternal && <Lock size={10} className="text-purple-400" />}
                                                </div>
                                                <span className="text-[10px] text-zinc-500">{formatDateTime(msg.timestamp)}</span>
                                            </div>
                                            <p className="text-xs text-zinc-300">{msg.content}</p>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            {/* Attachments */}
                            {selectedTicket.attachments && selectedTicket.attachments.length > 0 && (
                                <div className="mb-4">
                                    <h3 className="text-sm font-semibold text-zinc-100 mb-2">Attachments</h3>
                                    <div className="flex gap-2">
                                        {selectedTicket.attachments.map((file, idx) => (
                                            <div key={idx} className="flex items-center gap-2 px-3 py-2 bg-[#27272a] rounded text-xs text-zinc-300">
                                                <Paperclip size={12} className="text-zinc-500" />
                                                <span>{file}</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}

                            {/* Reply Form */}
                            <div>
                                <h3 className="text-sm font-semibold text-zinc-100 mb-3">Send Reply</h3>

                                {/* Canned Responses */}
                                <div className="mb-3">
                                    <label className="text-[10px] text-zinc-500 uppercase font-semibold mb-1 block">Quick Responses</label>
                                    <select
                                        onChange={(e) => {
                                            const response = CANNED_RESPONSES.find(r => r.id === e.target.value);
                                            if (response) handleCannedResponse(response.content);
                                        }}
                                        className="w-full px-3 py-1.5 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs"
                                    >
                                        <option value="">Select a template...</option>
                                        {CANNED_RESPONSES.map(resp => (
                                            <option key={resp.id} value={resp.id}>{resp.label}</option>
                                        ))}
                                    </select>
                                </div>

                                <textarea
                                    value={replyText}
                                    onChange={(e) => setReplyText(e.target.value)}
                                    placeholder="Type your reply..."
                                    className="w-full h-24 px-3 py-2 bg-[#27272a] border border-[#3f3f46] text-zinc-300 rounded text-xs focus:outline-none focus:border-blue-500 resize-none"
                                />

                                <div className="flex items-center justify-between mt-3">
                                    <div className="flex items-center gap-2">
                                        <input
                                            type="checkbox"
                                            id="internal-note"
                                            checked={isInternalNote}
                                            onChange={(e) => setIsInternalNote(e.target.checked)}
                                            className="rounded border-[#3f3f46]"
                                        />
                                        <label htmlFor="internal-note" className="text-xs text-zinc-400 flex items-center gap-1">
                                            <Lock size={12} />
                                            Internal Note (not visible to client)
                                        </label>
                                    </div>
                                    <button
                                        onClick={handleSendReply}
                                        disabled={!replyText.trim()}
                                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-zinc-700 disabled:text-zinc-500 text-white rounded text-xs transition-colors"
                                    >
                                        <Send size={14} />
                                        Send Reply
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
