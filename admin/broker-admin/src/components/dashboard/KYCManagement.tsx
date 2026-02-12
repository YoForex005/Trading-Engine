'use client';

import React, { useState, useMemo } from 'react';
import {
    Shield,
    CheckCircle,
    XCircle,
    Clock,
    AlertTriangle,
    Search,
    Calendar,
    Filter,
    X,
    FileText,
    User,
    Download,
    RefreshCw
} from 'lucide-react';

// Types
type KYCStatus = 'pending' | 'under_review' | 'approved' | 'rejected' | 'expired';
type DocumentType = 'passport' | 'national_id' | 'drivers_license';
type RiskLevel = 'low' | 'medium' | 'high';

interface KYCApplication {
    id: string;
    clientName: string;
    clientId: string;
    documentType: DocumentType;
    documentNumber: string;
    documentExpiry: string;
    status: KYCStatus;
    riskLevel: RiskLevel;
    amlCheck: boolean | null;
    pepCheck: boolean | null;
    sanctionsCheck: boolean | null;
    submitted: string;
    reviewedBy?: string;
    reviewedAt?: string;
    rejectionReason?: string;
    timeline: TimelineEvent[];
}

interface TimelineEvent {
    date: string;
    event: string;
    by?: string;
}

export default function KYCManagement() {
    const [selectedStatus, setSelectedStatus] = useState<string>('all');
    const [selectedDocType, setSelectedDocType] = useState<string>('all');
    const [selectedRiskLevel, setSelectedRiskLevel] = useState<string>('all');
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedApp, setSelectedApp] = useState<KYCApplication | null>(null);
    const [selectedApps, setSelectedApps] = useState<Set<string>>(new Set());
    const [rejectReason, setRejectReason] = useState('');

    // Mock data - 15 applications
    const mockApplications: KYCApplication[] = useMemo(() => [
        {
            id: 'KYC-001',
            clientName: 'John Smith',
            clientId: 'CLI-1001',
            documentType: 'passport',
            documentNumber: 'P12345678',
            documentExpiry: '2027-06-15',
            status: 'pending',
            riskLevel: 'low',
            amlCheck: null,
            pepCheck: null,
            sanctionsCheck: null,
            submitted: '2026-02-10T14:30:00Z',
            timeline: [
                { date: '2026-02-10T14:30:00Z', event: 'Application submitted' }
            ]
        },
        {
            id: 'KYC-002',
            clientName: 'Emma Wilson',
            clientId: 'CLI-1002',
            documentType: 'national_id',
            documentNumber: 'N87654321',
            documentExpiry: '2028-03-20',
            status: 'under_review',
            riskLevel: 'medium',
            amlCheck: true,
            pepCheck: false,
            sanctionsCheck: true,
            submitted: '2026-02-09T10:15:00Z',
            timeline: [
                { date: '2026-02-09T10:15:00Z', event: 'Application submitted' },
                { date: '2026-02-09T15:20:00Z', event: 'AML check passed' },
                { date: '2026-02-10T09:00:00Z', event: 'Under review', by: 'Admin User' }
            ]
        },
        {
            id: 'KYC-003',
            clientName: 'Michael Chen',
            clientId: 'CLI-1003',
            documentType: 'drivers_license',
            documentNumber: 'DL556677',
            documentExpiry: '2026-12-01',
            status: 'approved',
            riskLevel: 'low',
            amlCheck: true,
            pepCheck: true,
            sanctionsCheck: true,
            submitted: '2026-02-11T08:00:00Z',
            reviewedBy: 'Admin User',
            reviewedAt: '2026-02-11T10:30:00Z',
            timeline: [
                { date: '2026-02-11T08:00:00Z', event: 'Application submitted' },
                { date: '2026-02-11T08:15:00Z', event: 'AML check passed' },
                { date: '2026-02-11T10:30:00Z', event: 'Approved', by: 'Admin User' }
            ]
        },
        {
            id: 'KYC-004',
            clientName: 'Sarah Johnson',
            clientId: 'CLI-1004',
            documentType: 'passport',
            documentNumber: 'P98765432',
            documentExpiry: '2029-01-10',
            status: 'rejected',
            riskLevel: 'high',
            amlCheck: false,
            pepCheck: true,
            sanctionsCheck: false,
            submitted: '2026-02-10T16:45:00Z',
            reviewedBy: 'Admin User',
            reviewedAt: '2026-02-11T11:00:00Z',
            rejectionReason: 'Failed AML and sanctions checks',
            timeline: [
                { date: '2026-02-10T16:45:00Z', event: 'Application submitted' },
                { date: '2026-02-10T17:00:00Z', event: 'AML check failed' },
                { date: '2026-02-11T11:00:00Z', event: 'Rejected', by: 'Admin User' }
            ]
        },
        {
            id: 'KYC-005',
            clientName: 'David Lee',
            clientId: 'CLI-1005',
            documentType: 'national_id',
            documentNumber: 'N11223344',
            documentExpiry: '2026-03-15',
            status: 'expired',
            riskLevel: 'low',
            amlCheck: true,
            pepCheck: true,
            sanctionsCheck: true,
            submitted: '2024-01-15T12:00:00Z',
            reviewedBy: 'System',
            reviewedAt: '2026-03-15T00:00:00Z',
            timeline: [
                { date: '2024-01-15T12:00:00Z', event: 'Application submitted' },
                { date: '2024-01-15T14:00:00Z', event: 'Approved', by: 'Admin User' },
                { date: '2026-03-15T00:00:00Z', event: 'Expired' }
            ]
        },
        {
            id: 'KYC-006',
            clientName: 'Lisa Anderson',
            clientId: 'CLI-1006',
            documentType: 'passport',
            documentNumber: 'P55667788',
            documentExpiry: '2027-08-25',
            status: 'pending',
            riskLevel: 'low',
            amlCheck: null,
            pepCheck: null,
            sanctionsCheck: null,
            submitted: '2026-02-11T09:00:00Z',
            timeline: [
                { date: '2026-02-11T09:00:00Z', event: 'Application submitted' }
            ]
        },
        {
            id: 'KYC-007',
            clientName: 'Robert Brown',
            clientId: 'CLI-1007',
            documentType: 'drivers_license',
            documentNumber: 'DL998877',
            documentExpiry: '2028-05-30',
            status: 'under_review',
            riskLevel: 'medium',
            amlCheck: true,
            pepCheck: true,
            sanctionsCheck: true,
            submitted: '2026-02-10T11:30:00Z',
            timeline: [
                { date: '2026-02-10T11:30:00Z', event: 'Application submitted' },
                { date: '2026-02-10T12:00:00Z', event: 'All checks passed' },
                { date: '2026-02-11T08:00:00Z', event: 'Under review', by: 'Admin User' }
            ]
        },
        {
            id: 'KYC-008',
            clientName: 'Jennifer Davis',
            clientId: 'CLI-1008',
            documentType: 'national_id',
            documentNumber: 'N44556677',
            documentExpiry: '2027-11-12',
            status: 'approved',
            riskLevel: 'low',
            amlCheck: true,
            pepCheck: true,
            sanctionsCheck: true,
            submitted: '2026-02-11T07:15:00Z',
            reviewedBy: 'Admin User',
            reviewedAt: '2026-02-11T09:45:00Z',
            timeline: [
                { date: '2026-02-11T07:15:00Z', event: 'Application submitted' },
                { date: '2026-02-11T07:30:00Z', event: 'All checks passed' },
                { date: '2026-02-11T09:45:00Z', event: 'Approved', by: 'Admin User' }
            ]
        },
        {
            id: 'KYC-009',
            clientName: 'William Garcia',
            clientId: 'CLI-1009',
            documentType: 'passport',
            documentNumber: 'P33445566',
            documentExpiry: '2026-04-20',
            status: 'pending',
            riskLevel: 'high',
            amlCheck: null,
            pepCheck: null,
            sanctionsCheck: null,
            submitted: '2026-02-11T10:00:00Z',
            timeline: [
                { date: '2026-02-11T10:00:00Z', event: 'Application submitted' }
            ]
        },
        {
            id: 'KYC-010',
            clientName: 'Mary Martinez',
            clientId: 'CLI-1010',
            documentType: 'drivers_license',
            documentNumber: 'DL223344',
            documentExpiry: '2029-07-08',
            status: 'rejected',
            riskLevel: 'medium',
            amlCheck: true,
            pepCheck: false,
            sanctionsCheck: true,
            submitted: '2026-02-10T13:20:00Z',
            reviewedBy: 'Admin User',
            reviewedAt: '2026-02-11T08:30:00Z',
            rejectionReason: 'PEP check flagged - requires additional verification',
            timeline: [
                { date: '2026-02-10T13:20:00Z', event: 'Application submitted' },
                { date: '2026-02-10T14:00:00Z', event: 'PEP check failed' },
                { date: '2026-02-11T08:30:00Z', event: 'Rejected', by: 'Admin User' }
            ]
        },
        {
            id: 'KYC-011',
            clientName: 'James Rodriguez',
            clientId: 'CLI-1011',
            documentType: 'national_id',
            documentNumber: 'N77889900',
            documentExpiry: '2028-09-15',
            status: 'pending',
            riskLevel: 'low',
            amlCheck: null,
            pepCheck: null,
            sanctionsCheck: null,
            submitted: '2026-02-11T11:30:00Z',
            timeline: [
                { date: '2026-02-11T11:30:00Z', event: 'Application submitted' }
            ]
        },
        {
            id: 'KYC-012',
            clientName: 'Patricia Taylor',
            clientId: 'CLI-1012',
            documentType: 'passport',
            documentNumber: 'P66778899',
            documentExpiry: '2027-02-28',
            status: 'under_review',
            riskLevel: 'low',
            amlCheck: true,
            pepCheck: true,
            sanctionsCheck: true,
            submitted: '2026-02-10T15:00:00Z',
            timeline: [
                { date: '2026-02-10T15:00:00Z', event: 'Application submitted' },
                { date: '2026-02-10T15:30:00Z', event: 'All checks passed' },
                { date: '2026-02-11T09:00:00Z', event: 'Under review', by: 'Admin User' }
            ]
        },
        {
            id: 'KYC-013',
            clientName: 'Christopher Thomas',
            clientId: 'CLI-1013',
            documentType: 'drivers_license',
            documentNumber: 'DL445566',
            documentExpiry: '2030-12-31',
            status: 'approved',
            riskLevel: 'low',
            amlCheck: true,
            pepCheck: true,
            sanctionsCheck: true,
            submitted: '2026-02-11T06:00:00Z',
            reviewedBy: 'Admin User',
            reviewedAt: '2026-02-11T08:00:00Z',
            timeline: [
                { date: '2026-02-11T06:00:00Z', event: 'Application submitted' },
                { date: '2026-02-11T06:15:00Z', event: 'All checks passed' },
                { date: '2026-02-11T08:00:00Z', event: 'Approved', by: 'Admin User' }
            ]
        },
        {
            id: 'KYC-014',
            clientName: 'Linda Moore',
            clientId: 'CLI-1014',
            documentType: 'national_id',
            documentNumber: 'N99887766',
            documentExpiry: '2026-03-10',
            status: 'pending',
            riskLevel: 'medium',
            amlCheck: null,
            pepCheck: null,
            sanctionsCheck: null,
            submitted: '2026-02-11T12:00:00Z',
            timeline: [
                { date: '2026-02-11T12:00:00Z', event: 'Application submitted' }
            ]
        },
        {
            id: 'KYC-015',
            clientName: 'Daniel Jackson',
            clientId: 'CLI-1015',
            documentType: 'passport',
            documentNumber: 'P11223344',
            documentExpiry: '2028-06-18',
            status: 'rejected',
            riskLevel: 'high',
            amlCheck: false,
            pepCheck: false,
            sanctionsCheck: false,
            submitted: '2026-02-09T14:00:00Z',
            reviewedBy: 'Admin User',
            reviewedAt: '2026-02-10T10:00:00Z',
            rejectionReason: 'All verification checks failed - suspicious activity',
            timeline: [
                { date: '2026-02-09T14:00:00Z', event: 'Application submitted' },
                { date: '2026-02-09T14:30:00Z', event: 'All checks failed' },
                { date: '2026-02-10T10:00:00Z', event: 'Rejected', by: 'Admin User' }
            ]
        }
    ], []);

    // Filter applications
    const filteredApplications = useMemo(() => {
        return mockApplications.filter(app => {
            if (selectedStatus !== 'all' && app.status !== selectedStatus) return false;
            if (selectedDocType !== 'all' && app.documentType !== selectedDocType) return false;
            if (selectedRiskLevel !== 'all' && app.riskLevel !== selectedRiskLevel) return false;
            if (searchQuery && !app.clientName.toLowerCase().includes(searchQuery.toLowerCase()) && !app.clientId.toLowerCase().includes(searchQuery.toLowerCase())) return false;
            return true;
        });
    }, [mockApplications, selectedStatus, selectedDocType, selectedRiskLevel, searchQuery]);

    // Calculate stats
    const stats = useMemo(() => {
        const today = new Date().toISOString().split('T')[0];
        const expiringThreshold = new Date();
        expiringThreshold.setDate(expiringThreshold.getDate() + 30);

        return {
            pending: mockApplications.filter(a => a.status === 'pending').length,
            approvedToday: mockApplications.filter(a => a.status === 'approved' && a.reviewedAt?.startsWith(today)).length,
            rejectedToday: mockApplications.filter(a => a.status === 'rejected' && a.reviewedAt?.startsWith(today)).length,
            underReview: mockApplications.filter(a => a.status === 'under_review').length,
            expiringSoon: mockApplications.filter(a => {
                const expiry = new Date(a.documentExpiry);
                return expiry < expiringThreshold && expiry > new Date();
            }).length
        };
    }, [mockApplications]);

    // Status badge component
    const StatusBadge = ({ status }: { status: KYCStatus }) => {
        const colors = {
            pending: 'bg-[#f59e0b] text-white',
            under_review: 'bg-[#3b82f6] text-white',
            approved: 'bg-[#22c55e] text-white',
            rejected: 'bg-[#ef4444] text-white',
            expired: 'bg-[#71717a] text-white'
        };
        return (
            <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${colors[status]}`}>
                {status.replace('_', ' ').toUpperCase()}
            </span>
        );
    };

    // Risk badge component
    const RiskBadge = ({ risk }: { risk: RiskLevel }) => {
        const colors = {
            low: 'bg-[#22c55e] text-white',
            medium: 'bg-[#f59e0b] text-white',
            high: 'bg-[#ef4444] text-white'
        };
        return (
            <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${colors[risk]}`}>
                {risk.toUpperCase()}
            </span>
        );
    };

    // Check indicator component
    const CheckIndicator = ({ passed }: { passed: boolean | null }) => {
        if (passed === null) return <Clock size={14} className="text-[#71717a]" />;
        return passed ? <CheckCircle size={14} className="text-[#22c55e]" /> : <XCircle size={14} className="text-[#ef4444]" />;
    };

    // Handle bulk selection
    const toggleSelection = (id: string) => {
        const newSelected = new Set(selectedApps);
        if (newSelected.has(id)) {
            newSelected.delete(id);
        } else {
            newSelected.add(id);
        }
        setSelectedApps(newSelected);
    };

    const toggleAll = () => {
        if (selectedApps.size === filteredApplications.length) {
            setSelectedApps(new Set());
        } else {
            setSelectedApps(new Set(filteredApplications.map(a => a.id)));
        }
    };

    // Mock actions
    const handleApprove = (id: string) => {
        console.log('Approve:', id);
        setSelectedApp(null);
        setRejectReason('');
    };

    const handleReject = (id: string, reason: string) => {
        console.log('Reject:', id, reason);
        setSelectedApp(null);
        setRejectReason('');
    };

    const handleRunAML = (id: string) => {
        console.log('Run AML check:', id);
    };

    const handleBulkAML = () => {
        console.log('Bulk AML check:', Array.from(selectedApps));
    };

    const handleBulkApprove = () => {
        console.log('Bulk approve:', Array.from(selectedApps));
        setSelectedApps(new Set());
    };

    const handleBulkReject = () => {
        console.log('Bulk reject:', Array.from(selectedApps));
        setSelectedApps(new Set());
    };

    return (
        <div className="h-full flex flex-col bg-[#18181b] text-white p-4 overflow-hidden">
            {/* Header */}
            <div className="flex items-center justify-between mb-4">
                <h1 className="text-lg font-bold text-[#e4e4e7] flex items-center gap-2">
                    <Shield size={18} className="text-[#3b82f6]" />
                    KYC/AML Verification Management
                </h1>
            </div>

            {/* Stats Cards */}
            <div className="grid grid-cols-5 gap-2 mb-4">
                <div className="bg-[#27272a] p-3 rounded border border-[#f59e0b]">
                    <div className="text-[#f59e0b] text-[10px] mb-1">Pending Reviews</div>
                    <div className="text-2xl font-bold text-[#f59e0b]">{stats.pending}</div>
                </div>
                <div className="bg-[#27272a] p-3 rounded border border-[#22c55e]">
                    <div className="text-[#22c55e] text-[10px] mb-1">Approved Today</div>
                    <div className="text-2xl font-bold text-[#22c55e]">{stats.approvedToday}</div>
                </div>
                <div className="bg-[#27272a] p-3 rounded border border-[#ef4444]">
                    <div className="text-[#ef4444] text-[10px] mb-1">Rejected Today</div>
                    <div className="text-2xl font-bold text-[#ef4444]">{stats.rejectedToday}</div>
                </div>
                <div className="bg-[#27272a] p-3 rounded border border-[#3b82f6]">
                    <div className="text-[#3b82f6] text-[10px] mb-1">Under Review</div>
                    <div className="text-2xl font-bold text-[#3b82f6]">{stats.underReview}</div>
                </div>
                <div className="bg-[#27272a] p-3 rounded border border-[#f59e0b]">
                    <div className="text-[#f59e0b] text-[10px] mb-1">Expiring Soon</div>
                    <div className="text-2xl font-bold text-[#f59e0b]">{stats.expiringSoon}</div>
                </div>
            </div>

            {/* Filter Bar */}
            <div className="flex items-center gap-2 mb-4 text-xs">
                <Filter size={14} className="text-[#a1a1aa]" />
                <select
                    value={selectedStatus}
                    onChange={(e) => setSelectedStatus(e.target.value)}
                    className="px-2 py-1 bg-[#27272a] border border-[#3f3f46] rounded text-[#a1a1aa]"
                >
                    <option value="all">All Status</option>
                    <option value="pending">Pending</option>
                    <option value="under_review">Under Review</option>
                    <option value="approved">Approved</option>
                    <option value="rejected">Rejected</option>
                    <option value="expired">Expired</option>
                </select>
                <select
                    value={selectedDocType}
                    onChange={(e) => setSelectedDocType(e.target.value)}
                    className="px-2 py-1 bg-[#27272a] border border-[#3f3f46] rounded text-[#a1a1aa]"
                >
                    <option value="all">All Documents</option>
                    <option value="passport">Passport</option>
                    <option value="national_id">National ID</option>
                    <option value="drivers_license">Driver's License</option>
                </select>
                <select
                    value={selectedRiskLevel}
                    onChange={(e) => setSelectedRiskLevel(e.target.value)}
                    className="px-2 py-1 bg-[#27272a] border border-[#3f3f46] rounded text-[#a1a1aa]"
                >
                    <option value="all">All Risk Levels</option>
                    <option value="low">Low Risk</option>
                    <option value="medium">Medium Risk</option>
                    <option value="high">High Risk</option>
                </select>
                <div className="flex-1 flex items-center gap-2 px-2 py-1 bg-[#27272a] border border-[#3f3f46] rounded">
                    <Search size={14} className="text-[#a1a1aa]" />
                    <input
                        type="text"
                        placeholder="Search client name or ID..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="flex-1 bg-transparent text-[#a1a1aa] outline-none text-xs"
                    />
                </div>
            </div>

            {/* Bulk Actions */}
            {selectedApps.size > 0 && (
                <div className="flex items-center gap-2 mb-3 p-2 bg-[#27272a] rounded border border-[#3b82f6]">
                    <span className="text-xs text-[#a1a1aa]">{selectedApps.size} selected</span>
                    <button
                        onClick={handleBulkAML}
                        className="px-2 py-1 bg-[#3b82f6] text-white text-xs rounded hover:bg-[#2563eb]"
                    >
                        Run Bulk AML Check
                    </button>
                    <button
                        onClick={handleBulkApprove}
                        className="px-2 py-1 bg-[#22c55e] text-white text-xs rounded hover:bg-[#16a34a]"
                    >
                        Approve Selected
                    </button>
                    <button
                        onClick={handleBulkReject}
                        className="px-2 py-1 bg-[#ef4444] text-white text-xs rounded hover:bg-[#dc2626]"
                    >
                        Reject Selected
                    </button>
                </div>
            )}

            {/* Applications Table */}
            <div className="flex-1 overflow-auto bg-[#27272a] rounded border border-[#3f3f46]">
                <table className="w-full text-xs">
                    <thead className="sticky top-0 bg-[#27272a] border-b border-[#3f3f46]">
                        <tr className="text-[#a1a1aa]">
                            <th className="text-left p-2">
                                <input
                                    type="checkbox"
                                    checked={selectedApps.size === filteredApplications.length && filteredApplications.length > 0}
                                    onChange={toggleAll}
                                    className="cursor-pointer"
                                />
                            </th>
                            <th className="text-left p-2">Client Name</th>
                            <th className="text-left p-2">Client ID</th>
                            <th className="text-left p-2">Document Type</th>
                            <th className="text-left p-2">Status</th>
                            <th className="text-left p-2">Risk Level</th>
                            <th className="text-center p-2">AML</th>
                            <th className="text-center p-2">PEP</th>
                            <th className="text-center p-2">Sanctions</th>
                            <th className="text-left p-2">Submitted</th>
                            <th className="text-left p-2">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {filteredApplications.map((app) => (
                            <tr
                                key={app.id}
                                className="border-b border-[#18181b] hover:bg-[#18181b] cursor-pointer"
                                onClick={() => setSelectedApp(app)}
                            >
                                <td className="p-2" onClick={(e) => e.stopPropagation()}>
                                    <input
                                        type="checkbox"
                                        checked={selectedApps.has(app.id)}
                                        onChange={() => toggleSelection(app.id)}
                                        className="cursor-pointer"
                                    />
                                </td>
                                <td className="p-2 text-[#e4e4e7]">{app.clientName}</td>
                                <td className="p-2 text-[#3b82f6]">{app.clientId}</td>
                                <td className="p-2 text-[#a1a1aa]">{app.documentType.replace('_', ' ')}</td>
                                <td className="p-2"><StatusBadge status={app.status} /></td>
                                <td className="p-2"><RiskBadge risk={app.riskLevel} /></td>
                                <td className="p-2 text-center"><CheckIndicator passed={app.amlCheck} /></td>
                                <td className="p-2 text-center"><CheckIndicator passed={app.pepCheck} /></td>
                                <td className="p-2 text-center"><CheckIndicator passed={app.sanctionsCheck} /></td>
                                <td className="p-2 text-[#a1a1aa]">{new Date(app.submitted).toLocaleDateString()}</td>
                                <td className="p-2" onClick={(e) => e.stopPropagation()}>
                                    <button
                                        onClick={() => setSelectedApp(app)}
                                        className="text-[#3b82f6] hover:text-[#2563eb]"
                                    >
                                        View
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {/* Application Detail Modal */}
            {selectedApp && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setSelectedApp(null)}>
                    <div className="bg-[#27272a] rounded-lg w-[800px] max-h-[90vh] overflow-auto" onClick={(e) => e.stopPropagation()}>
                        {/* Modal Header */}
                        <div className="flex items-center justify-between p-4 border-b border-[#3f3f46]">
                            <h2 className="text-lg font-bold text-[#e4e4e7]">KYC Application - {selectedApp.id}</h2>
                            <button onClick={() => setSelectedApp(null)} className="text-[#a1a1aa] hover:text-white">
                                <X size={20} />
                            </button>
                        </div>

                        {/* Modal Content */}
                        <div className="p-4 space-y-4">
                            {/* Client Info */}
                            <div className="bg-[#18181b] p-3 rounded">
                                <h3 className="text-sm font-semibold text-[#e4e4e7] mb-2 flex items-center gap-2">
                                    <User size={14} />
                                    Client Information
                                </h3>
                                <div className="grid grid-cols-2 gap-2 text-xs">
                                    <div>
                                        <span className="text-[#a1a1aa]">Name:</span>
                                        <span className="text-[#e4e4e7] ml-2">{selectedApp.clientName}</span>
                                    </div>
                                    <div>
                                        <span className="text-[#a1a1aa]">Client ID:</span>
                                        <span className="text-[#3b82f6] ml-2">{selectedApp.clientId}</span>
                                    </div>
                                    <div>
                                        <span className="text-[#a1a1aa]">Status:</span>
                                        <span className="ml-2"><StatusBadge status={selectedApp.status} /></span>
                                    </div>
                                    <div>
                                        <span className="text-[#a1a1aa]">Risk Level:</span>
                                        <span className="ml-2"><RiskBadge risk={selectedApp.riskLevel} /></span>
                                    </div>
                                </div>
                            </div>

                            {/* Document Details */}
                            <div className="bg-[#18181b] p-3 rounded">
                                <h3 className="text-sm font-semibold text-[#e4e4e7] mb-2 flex items-center gap-2">
                                    <FileText size={14} />
                                    Document Details
                                </h3>
                                <div className="grid grid-cols-3 gap-2 text-xs">
                                    <div>
                                        <span className="text-[#a1a1aa]">Type:</span>
                                        <span className="text-[#e4e4e7] ml-2">{selectedApp.documentType.replace('_', ' ')}</span>
                                    </div>
                                    <div>
                                        <span className="text-[#a1a1aa]">Number:</span>
                                        <span className="text-[#e4e4e7] ml-2">{selectedApp.documentNumber}</span>
                                    </div>
                                    <div>
                                        <span className="text-[#a1a1aa]">Expiry:</span>
                                        <span className="text-[#e4e4e7] ml-2">{new Date(selectedApp.documentExpiry).toLocaleDateString()}</span>
                                    </div>
                                </div>
                            </div>

                            {/* Verification Checks */}
                            <div className="bg-[#18181b] p-3 rounded">
                                <h3 className="text-sm font-semibold text-[#e4e4e7] mb-2">Verification Checks</h3>
                                <div className="grid grid-cols-3 gap-2">
                                    <div className="flex items-center gap-2 text-xs">
                                        <CheckIndicator passed={selectedApp.amlCheck} />
                                        <span className="text-[#a1a1aa]">AML Check</span>
                                    </div>
                                    <div className="flex items-center gap-2 text-xs">
                                        <CheckIndicator passed={selectedApp.pepCheck} />
                                        <span className="text-[#a1a1aa]">PEP Check</span>
                                    </div>
                                    <div className="flex items-center gap-2 text-xs">
                                        <CheckIndicator passed={selectedApp.sanctionsCheck} />
                                        <span className="text-[#a1a1aa]">Sanctions Check</span>
                                    </div>
                                </div>
                                {selectedApp.amlCheck === null && (
                                    <button
                                        onClick={() => handleRunAML(selectedApp.id)}
                                        className="mt-2 px-3 py-1 bg-[#3b82f6] text-white text-xs rounded hover:bg-[#2563eb] flex items-center gap-1"
                                    >
                                        <RefreshCw size={12} />
                                        Run AML Check
                                    </button>
                                )}
                            </div>

                            {/* Document Previews */}
                            <div className="bg-[#18181b] p-3 rounded">
                                <h3 className="text-sm font-semibold text-[#e4e4e7] mb-2">Document Uploads</h3>
                                <div className="grid grid-cols-4 gap-2">
                                    {['ID Front', 'ID Back', 'Proof of Address', 'Selfie'].map((doc) => (
                                        <div key={doc} className="bg-[#27272a] p-3 rounded border border-[#3f3f46] text-center">
                                            <FileText size={24} className="mx-auto text-[#a1a1aa] mb-1" />
                                            <div className="text-[10px] text-[#a1a1aa]">{doc}</div>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            {/* Timeline */}
                            <div className="bg-[#18181b] p-3 rounded">
                                <h3 className="text-sm font-semibold text-[#e4e4e7] mb-2">Timeline</h3>
                                <div className="space-y-2">
                                    {selectedApp.timeline.map((event, idx) => (
                                        <div key={idx} className="flex items-start gap-2 text-xs">
                                            <div className="w-2 h-2 rounded-full bg-[#3b82f6] mt-1"></div>
                                            <div className="flex-1">
                                                <div className="text-[#e4e4e7]">{event.event}</div>
                                                <div className="text-[#71717a] text-[10px]">
                                                    {new Date(event.date).toLocaleString()}
                                                    {event.by && ` by ${event.by}`}
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            {/* Rejection Reason */}
                            {selectedApp.rejectionReason && (
                                <div className="bg-[#18181b] p-3 rounded border border-[#ef4444]">
                                    <h3 className="text-sm font-semibold text-[#ef4444] mb-2">Rejection Reason</h3>
                                    <p className="text-xs text-[#a1a1aa]">{selectedApp.rejectionReason}</p>
                                </div>
                            )}

                            {/* Actions */}
                            {selectedApp.status === 'pending' || selectedApp.status === 'under_review' ? (
                                <div className="flex items-start gap-2">
                                    <div className="flex-1">
                                        <label className="text-xs text-[#a1a1aa] mb-1 block">Rejection Reason (optional)</label>
                                        <textarea
                                            value={rejectReason}
                                            onChange={(e) => setRejectReason(e.target.value)}
                                            placeholder="Enter reason for rejection..."
                                            className="w-full px-2 py-1 bg-[#18181b] border border-[#3f3f46] rounded text-xs text-[#e4e4e7] resize-none"
                                            rows={2}
                                        />
                                    </div>
                                    <div className="flex gap-2">
                                        <button
                                            onClick={() => handleApprove(selectedApp.id)}
                                            className="px-4 py-2 bg-[#22c55e] text-white text-xs rounded hover:bg-[#16a34a]"
                                        >
                                            Approve
                                        </button>
                                        <button
                                            onClick={() => handleReject(selectedApp.id, rejectReason)}
                                            className="px-4 py-2 bg-[#ef4444] text-white text-xs rounded hover:bg-[#dc2626]"
                                        >
                                            Reject
                                        </button>
                                    </div>
                                </div>
                            ) : null}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
