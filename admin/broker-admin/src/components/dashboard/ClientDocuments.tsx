'use client';

import React, { useState, useMemo } from 'react';
import {
  FileText,
  CheckCircle,
  XCircle,
  Clock,
  AlertTriangle,
  Upload,
  Download,
  Search,
  Filter,
  User,
  Calendar,
  MessageSquare,
  Bell,
  Eye,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';

// ============================================================================
// TYPES
// ============================================================================

type DocumentStatus = 'PENDING' | 'APPROVED' | 'REJECTED' | 'EXPIRED';

type DocumentType =
  | 'PASSPORT'
  | 'ID_CARD'
  | 'DRIVERS_LICENSE'
  | 'UTILITY_BILL'
  | 'BANK_STATEMENT'
  | 'PROOF_OF_ADDRESS'
  | 'TAX_DOCUMENT'
  | 'OTHER';

interface ClientDocument {
  id: string;
  clientId: string;
  clientName: string;
  type: DocumentType;
  fileName: string;
  fileSize: number; // bytes
  uploadDate: number;
  reviewDate?: number;
  expiryDate?: number;
  status: DocumentStatus;
  reviewNotes?: string;
  rejectionReason?: string;
  fileUrl: string;
}

interface DocumentStats {
  total: number;
  pending: number;
  approved: number;
  rejected: number;
  expiringSoon: number;
  avgReviewTimeHours: number;
}

// ============================================================================
// MOCK DATA GENERATORS
// ============================================================================

const DOCUMENT_TYPES: DocumentType[] = [
  'PASSPORT',
  'ID_CARD',
  'DRIVERS_LICENSE',
  'UTILITY_BILL',
  'BANK_STATEMENT',
  'PROOF_OF_ADDRESS',
  'TAX_DOCUMENT',
  'OTHER',
];

const REJECTION_REASONS = [
  'Document expired',
  'Poor image quality',
  'Information not visible',
  'Document altered or tampered',
  'Name mismatch',
  'Address mismatch',
  'Document type incorrect',
  'Other (see notes)',
];

function generateMockDocuments(count: number): ClientDocument[] {
  const docs: ClientDocument[] = [];
  const now = Date.now();
  const statuses: DocumentStatus[] = ['PENDING', 'APPROVED', 'REJECTED', 'EXPIRED'];

  for (let i = 0; i < count; i++) {
    const uploadDate = now - Math.random() * 30 * 24 * 60 * 60 * 1000; // Last 30 days
    const status = statuses[Math.floor(Math.random() * statuses.length)];
    const type = DOCUMENT_TYPES[Math.floor(Math.random() * DOCUMENT_TYPES.length)];

    docs.push({
      id: `doc-${i + 1}`,
      clientId: `client-${Math.floor(Math.random() * 50) + 1}`,
      clientName: `Client ${String.fromCharCode(65 + (i % 26))}${Math.floor(i / 26) + 1}`,
      type,
      fileName: `${type.toLowerCase()}_${i + 1}.pdf`,
      fileSize: Math.floor(Math.random() * 5000000) + 100000, // 100KB - 5MB
      uploadDate,
      reviewDate: status !== 'PENDING' ? uploadDate + Math.random() * 5 * 24 * 60 * 60 * 1000 : undefined,
      expiryDate:
        type === 'PASSPORT' || type === 'ID_CARD' || type === 'DRIVERS_LICENSE'
          ? now + Math.random() * 365 * 24 * 60 * 60 * 1000
          : undefined,
      status,
      reviewNotes: status === 'APPROVED' ? 'Document verified successfully' : undefined,
      rejectionReason: status === 'REJECTED' ? REJECTION_REASONS[Math.floor(Math.random() * REJECTION_REASONS.length)] : undefined,
      fileUrl: `/uploads/documents/doc-${i + 1}.pdf`,
    });
  }

  return docs.sort((a, b) => a.uploadDate - b.uploadDate); // Oldest first
}

function calculateStats(docs: ClientDocument[]): DocumentStats {
  const now = Date.now();
  const thirtyDaysMs = 30 * 24 * 60 * 60 * 1000;

  const approved = docs.filter((d) => d.status === 'APPROVED');
  const reviewTimes = approved
    .filter((d) => d.reviewDate && d.uploadDate)
    .map((d) => (d.reviewDate! - d.uploadDate) / (1000 * 60 * 60)); // hours

  return {
    total: docs.length,
    pending: docs.filter((d) => d.status === 'PENDING').length,
    approved: approved.length,
    rejected: docs.filter((d) => d.status === 'REJECTED').length,
    expiringSoon: docs.filter((d) => d.expiryDate && d.expiryDate - now < thirtyDaysMs && d.expiryDate > now).length,
    avgReviewTimeHours: reviewTimes.length > 0 ? reviewTimes.reduce((a, b) => a + b, 0) / reviewTimes.length : 0,
  };
}

// ============================================================================
// MAIN COMPONENT
// ============================================================================

export default function ClientDocuments() {
  const [documents, setDocuments] = useState<ClientDocument[]>(() => generateMockDocuments(80));
  const [selectedDocument, setSelectedDocument] = useState<ClientDocument | null>(null);
  const [selectedClient, setSelectedClient] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<DocumentStatus | 'ALL'>('ALL');
  const [typeFilter, setTypeFilter] = useState<DocumentType | 'ALL'>('ALL');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedDocIds, setSelectedDocIds] = useState<Set<string>>(new Set());
  const [reviewNotes, setReviewNotes] = useState('');
  const [rejectionReason, setRejectionReason] = useState('');
  const [view, setView] = useState<'queue' | 'client' | 'expiring'>('queue');

  const stats = useMemo(() => calculateStats(documents), [documents]);

  // Filter documents
  const filteredDocuments = useMemo(() => {
    return documents.filter((doc) => {
      if (statusFilter !== 'ALL' && doc.status !== statusFilter) return false;
      if (typeFilter !== 'ALL' && doc.type !== typeFilter) return false;
      if (searchQuery && !doc.clientName.toLowerCase().includes(searchQuery.toLowerCase())) return false;
      if (selectedClient && doc.clientId !== selectedClient) return false;
      return true;
    });
  }, [documents, statusFilter, typeFilter, searchQuery, selectedClient]);

  const pendingDocuments = useMemo(
    () => documents.filter((d) => d.status === 'PENDING').sort((a, b) => a.uploadDate - b.uploadDate),
    [documents]
  );

  const expiringDocuments = useMemo(() => {
    const now = Date.now();
    const thirtyDaysMs = 30 * 24 * 60 * 60 * 1000;
    return documents
      .filter((d) => d.expiryDate && d.expiryDate - now < thirtyDaysMs && d.expiryDate > now)
      .sort((a, b) => (a.expiryDate || 0) - (b.expiryDate || 0));
  }, [documents]);

  const uniqueClients = useMemo(() => {
    const clientMap = new Map<string, string>();
    documents.forEach((doc) => {
      if (!clientMap.has(doc.clientId)) {
        clientMap.set(doc.clientId, doc.clientName);
      }
    });
    return Array.from(clientMap.entries()).map(([id, name]) => ({ id, name }));
  }, [documents]);

  // Document type breakdown for donut chart
  const typeBreakdown = useMemo(() => {
    const counts = new Map<DocumentType, number>();
    DOCUMENT_TYPES.forEach((type) => counts.set(type, 0));
    documents.forEach((doc) => {
      counts.set(doc.type, (counts.get(doc.type) || 0) + 1);
    });
    return Array.from(counts.entries()).map(([type, count]) => ({ type, count }));
  }, [documents]);

  // Handlers
  const handleApprove = (docId: string) => {
    setDocuments((prev) =>
      prev.map((doc) =>
        doc.id === docId
          ? {
              ...doc,
              status: 'APPROVED' as DocumentStatus,
              reviewDate: Date.now(),
              reviewNotes,
            }
          : doc
      )
    );
    setReviewNotes('');
    setSelectedDocument(null);
  };

  const handleReject = (docId: string) => {
    if (!rejectionReason) {
      alert('Please select a rejection reason');
      return;
    }
    setDocuments((prev) =>
      prev.map((doc) =>
        doc.id === docId
          ? {
              ...doc,
              status: 'REJECTED' as DocumentStatus,
              reviewDate: Date.now(),
              rejectionReason,
              reviewNotes,
            }
          : doc
      )
    );
    setRejectionReason('');
    setReviewNotes('');
    setSelectedDocument(null);
  };

  const handleBulkApprove = () => {
    if (selectedDocIds.size === 0) return;
    setDocuments((prev) =>
      prev.map((doc) =>
        selectedDocIds.has(doc.id)
          ? {
              ...doc,
              status: 'APPROVED' as DocumentStatus,
              reviewDate: Date.now(),
              reviewNotes: 'Bulk approved',
            }
          : doc
      )
    );
    setSelectedDocIds(new Set());
  };

  const handleBulkReject = () => {
    if (selectedDocIds.size === 0) return;
    const reason = prompt('Enter rejection reason for all selected documents:');
    if (!reason) return;
    setDocuments((prev) =>
      prev.map((doc) =>
        selectedDocIds.has(doc.id)
          ? {
              ...doc,
              status: 'REJECTED' as DocumentStatus,
              reviewDate: Date.now(),
              rejectionReason: reason,
            }
          : doc
      )
    );
    setSelectedDocIds(new Set());
  };

  const toggleSelection = (docId: string) => {
    setSelectedDocIds((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(docId)) {
        newSet.delete(docId);
      } else {
        newSet.add(docId);
      }
      return newSet;
    });
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDate = (timestamp: number): string => {
    return new Date(timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const getStatusColor = (status: DocumentStatus): string => {
    switch (status) {
      case 'APPROVED':
        return 'text-[#22c55e]';
      case 'PENDING':
        return 'text-[#F5C542]';
      case 'REJECTED':
        return 'text-[#E74C3C]';
      case 'EXPIRED':
        return 'text-[#666]';
    }
  };

  const getStatusIcon = (status: DocumentStatus) => {
    switch (status) {
      case 'APPROVED':
        return <CheckCircle size={14} className="text-[#22c55e]" />;
      case 'PENDING':
        return <Clock size={14} className="text-[#F5C542]" />;
      case 'REJECTED':
        return <XCircle size={14} className="text-[#E74C3C]" />;
      case 'EXPIRED':
        return <AlertTriangle size={14} className="text-[#666]" />;
    }
  };

  const getTypeIcon = (type: DocumentType) => {
    return <FileText size={14} className="text-[#3B82F6]" />;
  };

  // Render donut chart
  const renderDonutChart = () => {
    const total = typeBreakdown.reduce((sum, item) => sum + item.count, 0);
    if (total === 0) return null;

    const colors = ['#3B82F6', '#10B981', '#F59E0B', '#8B5CF6', '#EC4899', '#F97316', '#06B6D4', '#84CC16'];
    let currentAngle = 0;

    const paths = typeBreakdown.map((item, index) => {
      const percentage = item.count / total;
      const angle = percentage * 360;
      const startAngle = currentAngle;
      const endAngle = currentAngle + angle;
      currentAngle = endAngle;

      const startRad = (startAngle - 90) * (Math.PI / 180);
      const endRad = (endAngle - 90) * (Math.PI / 180);

      const x1 = 50 + 40 * Math.cos(startRad);
      const y1 = 50 + 40 * Math.sin(startRad);
      const x2 = 50 + 40 * Math.cos(endRad);
      const y2 = 50 + 40 * Math.sin(endRad);

      const largeArc = angle > 180 ? 1 : 0;

      return (
        <path
          key={item.type}
          d={`M 50 50 L ${x1} ${y1} A 40 40 0 ${largeArc} 1 ${x2} ${y2} Z`}
          fill={colors[index % colors.length]}
          opacity={0.8}
        />
      );
    });

    return (
      <div className="flex items-center gap-4">
        <svg width="120" height="120" viewBox="0 0 100 100">
          {paths}
          <circle cx="50" cy="50" r="25" fill="#121316" />
        </svg>
        <div className="flex flex-col gap-1 text-xs">
          {typeBreakdown.map((item, index) => (
            <div key={item.type} className="flex items-center gap-2">
              <div className="w-3 h-3" style={{ backgroundColor: colors[index % colors.length] }} />
              <span className="text-[#999]">
                {item.type.replace('_', ' ')}: {item.count}
              </span>
            </div>
          ))}
        </div>
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#CCC] overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-[#383A42] flex-shrink-0">
        <div className="flex items-center gap-2">
          <FileText size={18} className="text-[#3B82F6]" />
          <h1 className="text-sm font-bold text-white">Client Documents / KYC Review</h1>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setView('queue')}
            className={`px-3 py-1 text-xs rounded ${
              view === 'queue' ? 'bg-[#2980B9] text-white' : 'bg-[#1E2026] text-[#999] hover:bg-[#25272E]'
            }`}
          >
            Review Queue
          </button>
          <button
            onClick={() => setView('client')}
            className={`px-3 py-1 text-xs rounded ${
              view === 'client' ? 'bg-[#2980B9] text-white' : 'bg-[#1E2026] text-[#999] hover:bg-[#25272E]'
            }`}
          >
            By Client
          </button>
          <button
            onClick={() => setView('expiring')}
            className={`px-3 py-1 text-xs rounded ${
              view === 'expiring' ? 'bg-[#2980B9] text-white' : 'bg-[#1E2026] text-[#999] hover:bg-[#25272E]'
            }`}
          >
            Expiring Soon
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-6 gap-3 px-4 py-3 border-b border-[#383A42] flex-shrink-0">
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <FileText size={14} className="text-[#3B82F6]" />
            <span className="text-xs text-[#666]">Total</span>
          </div>
          <div className="text-xl font-bold text-white">{stats.total}</div>
        </div>
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <Clock size={14} className="text-[#F5C542]" />
            <span className="text-xs text-[#666]">Pending</span>
          </div>
          <div className="text-xl font-bold text-[#F5C542]">{stats.pending}</div>
        </div>
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <CheckCircle size={14} className="text-[#22c55e]" />
            <span className="text-xs text-[#666]">Approved</span>
          </div>
          <div className="text-xl font-bold text-[#22c55e]">{stats.approved}</div>
        </div>
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <XCircle size={14} className="text-[#E74C3C]" />
            <span className="text-xs text-[#666]">Rejected</span>
          </div>
          <div className="text-xl font-bold text-[#E74C3C]">{stats.rejected}</div>
        </div>
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <AlertTriangle size={14} className="text-[#F59E0B]" />
            <span className="text-xs text-[#666]">Expiring</span>
          </div>
          <div className="text-xl font-bold text-[#F59E0B]">{stats.expiringSoon}</div>
        </div>
        <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
          <div className="flex items-center gap-2 mb-1">
            <Clock size={14} className="text-[#10B981]" />
            <span className="text-xs text-[#666]">Avg Review</span>
          </div>
          <div className="text-xl font-bold text-[#10B981]">{stats.avgReviewTimeHours.toFixed(1)}h</div>
        </div>
      </div>

      {/* Filter Bar */}
      <div className="flex items-center gap-3 px-4 py-2 border-b border-[#383A42] bg-[#1E2026] flex-shrink-0">
        <div className="flex items-center gap-2 bg-[#121316] px-2 py-1 rounded border border-[#383A42] flex-1 max-w-xs">
          <Search size={14} className="text-[#666]" />
          <input
            type="text"
            placeholder="Search by client name..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="bg-transparent text-xs text-white outline-none flex-1"
          />
        </div>

        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as DocumentStatus | 'ALL')}
          className="bg-[#121316] text-xs text-[#CCC] px-2 py-1 rounded border border-[#383A42] outline-none"
        >
          <option value="ALL">All Statuses</option>
          <option value="PENDING">Pending</option>
          <option value="APPROVED">Approved</option>
          <option value="REJECTED">Rejected</option>
          <option value="EXPIRED">Expired</option>
        </select>

        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value as DocumentType | 'ALL')}
          className="bg-[#121316] text-xs text-[#CCC] px-2 py-1 rounded border border-[#383A42] outline-none"
        >
          <option value="ALL">All Types</option>
          {DOCUMENT_TYPES.map((type) => (
            <option key={type} value={type}>
              {type.replace('_', ' ')}
            </option>
          ))}
        </select>

        {view === 'client' && (
          <select
            value={selectedClient || ''}
            onChange={(e) => setSelectedClient(e.target.value || null)}
            className="bg-[#121316] text-xs text-[#CCC] px-2 py-1 rounded border border-[#383A42] outline-none"
          >
            <option value="">Select Client</option>
            {uniqueClients.map((client) => (
              <option key={client.id} value={client.id}>
                {client.name}
              </option>
            ))}
          </select>
        )}

        {selectedDocIds.size > 0 && view === 'queue' && (
          <div className="flex items-center gap-2 ml-auto">
            <span className="text-xs text-[#999]">{selectedDocIds.size} selected</span>
            <button
              onClick={handleBulkApprove}
              className="px-3 py-1 text-xs bg-[#22c55e] text-white rounded hover:bg-[#16a34a]"
            >
              Approve All
            </button>
            <button
              onClick={handleBulkReject}
              className="px-3 py-1 text-xs bg-[#E74C3C] text-white rounded hover:bg-[#C0392B]"
            >
              Reject All
            </button>
          </div>
        )}
      </div>

      {/* Main Content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left Panel - Document List or Type Breakdown */}
        <div className="flex-1 flex flex-col border-r border-[#383A42] overflow-hidden">
          {view === 'queue' && (
            <>
              <div className="px-4 py-2 border-b border-[#383A42] bg-[#1E2026] flex-shrink-0">
                <h2 className="text-xs font-bold text-white">Pending Review Queue ({pendingDocuments.length})</h2>
              </div>
              <div className="flex-1 overflow-y-auto custom-scrollbar">
                <table className="w-full text-xs">
                  <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                    <tr>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">
                        <input
                          type="checkbox"
                          checked={selectedDocIds.size === pendingDocuments.length && pendingDocuments.length > 0}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setSelectedDocIds(new Set(pendingDocuments.map((d) => d.id)));
                            } else {
                              setSelectedDocIds(new Set());
                            }
                          }}
                        />
                      </th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Client</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Type</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">File Name</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Upload Date</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Size</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {pendingDocuments.map((doc) => (
                      <tr
                        key={doc.id}
                        className={`border-b border-[#383A42] hover:bg-[#1E2026] cursor-pointer ${
                          selectedDocument?.id === doc.id ? 'bg-[#25272E]' : ''
                        }`}
                        onClick={() => setSelectedDocument(doc)}
                      >
                        <td className="px-3 py-2" onClick={(e) => e.stopPropagation()}>
                          <input
                            type="checkbox"
                            checked={selectedDocIds.has(doc.id)}
                            onChange={() => toggleSelection(doc.id)}
                          />
                        </td>
                        <td className="px-3 py-2 text-white">{doc.clientName}</td>
                        <td className="px-3 py-2">
                          <div className="flex items-center gap-1">
                            {getTypeIcon(doc.type)}
                            <span className="text-[#999]">{doc.type.replace('_', ' ')}</span>
                          </div>
                        </td>
                        <td className="px-3 py-2 text-[#3B82F6]">{doc.fileName}</td>
                        <td className="px-3 py-2 text-[#999]">{formatDate(doc.uploadDate)}</td>
                        <td className="px-3 py-2 text-[#999]">{formatFileSize(doc.fileSize)}</td>
                        <td className="px-3 py-2">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedDocument(doc);
                            }}
                            className="px-2 py-1 text-xs bg-[#2980B9] text-white rounded hover:bg-[#2574A9]"
                          >
                            Review
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          )}

          {view === 'client' && (
            <>
              <div className="px-4 py-2 border-b border-[#383A42] bg-[#1E2026] flex-shrink-0">
                <h2 className="text-xs font-bold text-white">Client Documents ({filteredDocuments.length})</h2>
              </div>
              <div className="flex-1 overflow-y-auto custom-scrollbar">
                <table className="w-full text-xs">
                  <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                    <tr>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Type</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">File Name</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Status</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Upload Date</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Expiry</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredDocuments.map((doc) => (
                      <tr
                        key={doc.id}
                        className={`border-b border-[#383A42] hover:bg-[#1E2026] cursor-pointer ${
                          selectedDocument?.id === doc.id ? 'bg-[#25272E]' : ''
                        }`}
                        onClick={() => setSelectedDocument(doc)}
                      >
                        <td className="px-3 py-2">
                          <div className="flex items-center gap-1">
                            {getTypeIcon(doc.type)}
                            <span className="text-[#999]">{doc.type.replace('_', ' ')}</span>
                          </div>
                        </td>
                        <td className="px-3 py-2 text-[#3B82F6]">{doc.fileName}</td>
                        <td className="px-3 py-2">
                          <div className="flex items-center gap-1">
                            {getStatusIcon(doc.status)}
                            <span className={getStatusColor(doc.status)}>{doc.status}</span>
                          </div>
                        </td>
                        <td className="px-3 py-2 text-[#999]">{formatDate(doc.uploadDate)}</td>
                        <td className="px-3 py-2 text-[#999]">{doc.expiryDate ? formatDate(doc.expiryDate) : '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          )}

          {view === 'expiring' && (
            <>
              <div className="px-4 py-2 border-b border-[#383A42] bg-[#1E2026] flex-shrink-0">
                <h2 className="text-xs font-bold text-white">Expiring Within 30 Days ({expiringDocuments.length})</h2>
              </div>
              <div className="flex-1 overflow-y-auto custom-scrollbar">
                <table className="w-full text-xs">
                  <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                    <tr>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Client</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Type</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Expiry Date</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Days Left</th>
                      <th className="text-left px-3 py-2 font-bold text-[#999]">Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {expiringDocuments.map((doc) => {
                      const daysLeft = Math.floor(((doc.expiryDate || 0) - Date.now()) / (24 * 60 * 60 * 1000));
                      return (
                        <tr
                          key={doc.id}
                          className={`border-b border-[#383A42] hover:bg-[#1E2026] cursor-pointer ${
                            selectedDocument?.id === doc.id ? 'bg-[#25272E]' : ''
                          }`}
                          onClick={() => setSelectedDocument(doc)}
                        >
                          <td className="px-3 py-2 text-white">{doc.clientName}</td>
                          <td className="px-3 py-2">
                            <div className="flex items-center gap-1">
                              {getTypeIcon(doc.type)}
                              <span className="text-[#999]">{doc.type.replace('_', ' ')}</span>
                            </div>
                          </td>
                          <td className="px-3 py-2 text-[#999]">{formatDate(doc.expiryDate || 0)}</td>
                          <td className="px-3 py-2">
                            <span className={daysLeft < 7 ? 'text-[#E74C3C]' : 'text-[#F5C542]'}>{daysLeft} days</span>
                          </td>
                          <td className="px-3 py-2">
                            <button className="px-2 py-1 text-xs bg-[#F59E0B] text-white rounded hover:bg-[#D97706] flex items-center gap-1">
                              <Bell size={12} />
                              Notify
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </>
          )}
        </div>

        {/* Right Panel - Document Review or Type Breakdown */}
        <div className="w-96 flex flex-col bg-[#1E2026] overflow-hidden">
          {selectedDocument ? (
            <>
              <div className="px-4 py-2 border-b border-[#383A42] flex items-center justify-between flex-shrink-0">
                <h2 className="text-xs font-bold text-white">Document Review</h2>
                <button onClick={() => setSelectedDocument(null)} className="text-[#999] hover:text-white">
                  ✕
                </button>
              </div>

              <div className="flex-1 overflow-y-auto custom-scrollbar p-4">
                {/* Document Info */}
                <div className="mb-4">
                  <div className="text-xs text-[#666] mb-1">Client</div>
                  <div className="text-sm text-white font-bold">{selectedDocument.clientName}</div>
                </div>

                <div className="grid grid-cols-2 gap-3 mb-4">
                  <div>
                    <div className="text-xs text-[#666] mb-1">Type</div>
                    <div className="text-xs text-white">{selectedDocument.type.replace('_', ' ')}</div>
                  </div>
                  <div>
                    <div className="text-xs text-[#666] mb-1">Status</div>
                    <div className="flex items-center gap-1">
                      {getStatusIcon(selectedDocument.status)}
                      <span className={`text-xs ${getStatusColor(selectedDocument.status)}`}>{selectedDocument.status}</span>
                    </div>
                  </div>
                </div>

                <div className="mb-4">
                  <div className="text-xs text-[#666] mb-1">File Name</div>
                  <div className="text-xs text-[#3B82F6]">{selectedDocument.fileName}</div>
                </div>

                <div className="grid grid-cols-2 gap-3 mb-4">
                  <div>
                    <div className="text-xs text-[#666] mb-1">Upload Date</div>
                    <div className="text-xs text-white">{formatDate(selectedDocument.uploadDate)}</div>
                  </div>
                  <div>
                    <div className="text-xs text-[#666] mb-1">File Size</div>
                    <div className="text-xs text-white">{formatFileSize(selectedDocument.fileSize)}</div>
                  </div>
                </div>

                {selectedDocument.expiryDate && (
                  <div className="mb-4">
                    <div className="text-xs text-[#666] mb-1">Expiry Date</div>
                    <div className="text-xs text-white">{formatDate(selectedDocument.expiryDate)}</div>
                  </div>
                )}

                {/* Document Preview */}
                <div className="mb-4 bg-[#121316] border border-[#383A42] rounded p-4 flex items-center justify-center h-48">
                  <div className="text-center text-[#666]">
                    <FileText size={48} className="mx-auto mb-2 opacity-30" />
                    <div className="text-xs">Document Preview</div>
                    <button className="mt-2 px-3 py-1 text-xs bg-[#2980B9] text-white rounded hover:bg-[#2574A9] flex items-center gap-1 mx-auto">
                      <Eye size={12} />
                      View Full Document
                    </button>
                  </div>
                </div>

                {selectedDocument.status === 'PENDING' && (
                  <>
                    {/* Rejection Reason */}
                    <div className="mb-3">
                      <label className="text-xs text-[#666] mb-1 block">Rejection Reason (if rejecting)</label>
                      <select
                        value={rejectionReason}
                        onChange={(e) => setRejectionReason(e.target.value)}
                        className="w-full bg-[#121316] text-xs text-white px-2 py-2 rounded border border-[#383A42] outline-none"
                      >
                        <option value="">Select reason...</option>
                        {REJECTION_REASONS.map((reason) => (
                          <option key={reason} value={reason}>
                            {reason}
                          </option>
                        ))}
                      </select>
                    </div>

                    {/* Review Notes */}
                    <div className="mb-4">
                      <label className="text-xs text-[#666] mb-1 block">Review Notes</label>
                      <textarea
                        value={reviewNotes}
                        onChange={(e) => setReviewNotes(e.target.value)}
                        className="w-full bg-[#121316] text-xs text-white px-2 py-2 rounded border border-[#383A42] outline-none resize-none"
                        rows={3}
                        placeholder="Add notes about this review..."
                      />
                    </div>

                    {/* Action Buttons */}
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleApprove(selectedDocument.id)}
                        className="flex-1 px-3 py-2 text-xs bg-[#22c55e] text-white rounded hover:bg-[#16a34a] font-bold"
                      >
                        Approve
                      </button>
                      <button
                        onClick={() => handleReject(selectedDocument.id)}
                        className="flex-1 px-3 py-2 text-xs bg-[#E74C3C] text-white rounded hover:bg-[#C0392B] font-bold"
                      >
                        Reject
                      </button>
                    </div>
                  </>
                )}

                {selectedDocument.status !== 'PENDING' && (
                  <div className="bg-[#121316] border border-[#383A42] rounded p-3">
                    <div className="text-xs text-[#666] mb-2">Review Information</div>
                    {selectedDocument.reviewDate && (
                      <div className="text-xs text-white mb-2">Reviewed: {formatDate(selectedDocument.reviewDate)}</div>
                    )}
                    {selectedDocument.rejectionReason && (
                      <div className="text-xs text-[#E74C3C] mb-2">Reason: {selectedDocument.rejectionReason}</div>
                    )}
                    {selectedDocument.reviewNotes && (
                      <div className="text-xs text-[#999]">Notes: {selectedDocument.reviewNotes}</div>
                    )}
                  </div>
                )}
              </div>
            </>
          ) : (
            <>
              <div className="px-4 py-2 border-b border-[#383A42] flex-shrink-0">
                <h2 className="text-xs font-bold text-white">Document Type Breakdown</h2>
              </div>
              <div className="flex-1 overflow-y-auto custom-scrollbar p-4">
                <div className="flex justify-center">{renderDonutChart()}</div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
