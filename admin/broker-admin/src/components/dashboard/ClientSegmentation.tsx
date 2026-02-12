'use client';

import React, { useState, useMemo } from 'react';
import { Users, Tag, Plus, X, Edit2, Mail, Settings, TrendingUp, Clock, MapPin, UserCheck, Filter, Search, Download } from 'lucide-react';

type SegmentConditionField = 'balance' | 'volume' | 'lastLogin' | 'accountAge' | 'country' | 'accountType';
type SegmentConditionOperator = '>' | '<' | '=' | '>=' | '<=';

interface SegmentCondition {
  field: SegmentConditionField;
  operator: SegmentConditionOperator;
  value: string | number;
}

interface Segment {
  id: string;
  name: string;
  description: string;
  color: string;
  clientCount: number;
  autoAssignRules: SegmentCondition[];
  createdAt: number;
  updatedAt: number;
}

interface Client {
  id: string;
  name: string;
  email: string;
  balance: number;
  volume: number;
  lastLogin: number;
  accountAge: number;
  country: string;
  accountType: string;
  segments: string[];
}

interface SegmentOverlap {
  segments: string[];
  count: number;
}

// Mock data generator
const generateMockSegments = (): Segment[] => {
  const now = Date.now();
  return [
    {
      id: 'seg-1',
      name: 'VIP Traders',
      description: 'High-value clients with substantial balances and trading volume',
      color: '#F59E0B',
      clientCount: 47,
      autoAssignRules: [
        { field: 'balance', operator: '>', value: 100000 },
        { field: 'volume', operator: '>', value: 1000000 }
      ],
      createdAt: now - 90 * 86400000,
      updatedAt: now - 5 * 86400000
    },
    {
      id: 'seg-2',
      name: 'High Volume',
      description: 'Active traders with high monthly trading volume',
      color: '#3B82F6',
      clientCount: 183,
      autoAssignRules: [
        { field: 'volume', operator: '>', value: 500000 }
      ],
      createdAt: now - 60 * 86400000,
      updatedAt: now - 2 * 86400000
    },
    {
      id: 'seg-3',
      name: 'New Accounts',
      description: 'Recently registered accounts (less than 30 days old)',
      color: '#10B981',
      clientCount: 234,
      autoAssignRules: [
        { field: 'accountAge', operator: '<', value: 30 }
      ],
      createdAt: now - 30 * 86400000,
      updatedAt: now - 1 * 86400000
    },
    {
      id: 'seg-4',
      name: 'Dormant',
      description: 'Inactive accounts with no login in the last 60 days',
      color: '#6B7280',
      clientCount: 298,
      autoAssignRules: [
        { field: 'lastLogin', operator: '>', value: 60 }
      ],
      createdAt: now - 45 * 86400000,
      updatedAt: now - 3 * 86400000
    },
    {
      id: 'seg-5',
      name: 'At Risk',
      description: 'Clients showing signs of potential churn or issues',
      color: '#EF4444',
      clientCount: 76,
      autoAssignRules: [
        { field: 'balance', operator: '<', value: 1000 },
        { field: 'lastLogin', operator: '>', value: 30 }
      ],
      createdAt: now - 20 * 86400000,
      updatedAt: now - 1 * 86400000
    },
    {
      id: 'seg-6',
      name: 'Islamic',
      description: 'Islamic/swap-free accounts',
      color: '#8B5CF6',
      clientCount: 124,
      autoAssignRules: [
        { field: 'accountType', operator: '=', value: 'Islamic' }
      ],
      createdAt: now - 75 * 86400000,
      updatedAt: now - 10 * 86400000
    },
    {
      id: 'seg-7',
      name: 'Crypto Only',
      description: 'Clients trading exclusively cryptocurrency pairs',
      color: '#EC4899',
      clientCount: 156,
      autoAssignRules: [
        { field: 'accountType', operator: '=', value: 'Crypto' }
      ],
      createdAt: now - 40 * 86400000,
      updatedAt: now - 7 * 86400000
    },
    {
      id: 'seg-8',
      name: 'Scalpers',
      description: 'High-frequency traders with many small positions',
      color: '#F97316',
      clientCount: 89,
      autoAssignRules: [
        { field: 'volume', operator: '>', value: 300000 },
        { field: 'accountType', operator: '=', value: 'Standard' }
      ],
      createdAt: now - 50 * 86400000,
      updatedAt: now - 4 * 86400000
    }
  ];
};

const generateMockClients = (segments: Segment[]): Client[] => {
  const clients: Client[] = [];
  const now = Date.now();
  const countries = ['USA', 'UK', 'Germany', 'France', 'Australia', 'Singapore', 'UAE', 'Japan'];
  const accountTypes = ['Standard', 'ECN', 'Islamic', 'Crypto', 'Pro'];

  for (let i = 0; i < 300; i++) {
    const balance = Math.random() * 200000;
    const volume = Math.random() * 2000000;
    const lastLoginDays = Math.floor(Math.random() * 120);
    const accountAgeDays = Math.floor(Math.random() * 365);
    const country = countries[Math.floor(Math.random() * countries.length)];
    const accountType = accountTypes[Math.floor(Math.random() * accountTypes.length)];

    // Determine which segments this client belongs to
    const clientSegments: string[] = [];
    segments.forEach(seg => {
      let matchesAll = true;
      seg.autoAssignRules.forEach(rule => {
        let value: number | string = 0;
        switch (rule.field) {
          case 'balance': value = balance; break;
          case 'volume': value = volume; break;
          case 'lastLogin': value = lastLoginDays; break;
          case 'accountAge': value = accountAgeDays; break;
          case 'country': value = country; break;
          case 'accountType': value = accountType; break;
        }

        switch (rule.operator) {
          case '>': if (!(value > rule.value)) matchesAll = false; break;
          case '<': if (!(value < rule.value)) matchesAll = false; break;
          case '>=': if (!(value >= rule.value)) matchesAll = false; break;
          case '<=': if (!(value <= rule.value)) matchesAll = false; break;
          case '=': if (value !== rule.value) matchesAll = false; break;
        }
      });

      if (matchesAll) {
        clientSegments.push(seg.id);
      }
    });

    clients.push({
      id: `client-${i + 1}`,
      name: `Trader ${i + 1}`,
      email: `trader${i + 1}@example.com`,
      balance,
      volume,
      lastLogin: now - lastLoginDays * 86400000,
      accountAge: accountAgeDays,
      country,
      accountType,
      segments: clientSegments
    });
  }

  return clients;
};

const ClientSegmentation: React.FC = () => {
  const [segments, setSegments] = useState<Segment[]>(generateMockSegments());
  const [clients] = useState<Client[]>(() => generateMockClients(generateMockSegments()));
  const [selectedSegment, setSelectedSegment] = useState<Segment | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showSegmentClients, setShowSegmentClients] = useState(false);
  const [editingSegment, setEditingSegment] = useState<Segment | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedClients, setSelectedClients] = useState<string[]>([]);
  const [viewMode, setViewMode] = useState<'cards' | 'tags'>('cards');

  // Form state for create/edit modal
  const [formName, setFormName] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formColor, setFormColor] = useState('#3B82F6');
  const [formRules, setFormRules] = useState<SegmentCondition[]>([]);

  // Calculate stats
  const totalSegments = segments.length;
  const taggedClients = clients.filter(c => c.segments.length > 0).length;
  const untaggedClients = clients.length - taggedClients;
  const mostPopularSegment = segments.reduce((max, seg) =>
    seg.clientCount > max.clientCount ? seg : max, segments[0]
  );

  // Calculate segment overlaps for Venn diagram
  const segmentOverlaps = useMemo(() => {
    const overlaps: SegmentOverlap[] = [];
    const topSegments = segments.slice(0, 3);

    // Single segments
    topSegments.forEach(seg => {
      const count = clients.filter(c => c.segments.includes(seg.id) && c.segments.length === 1).length;
      overlaps.push({ segments: [seg.id], count });
    });

    // Pairwise overlaps
    for (let i = 0; i < topSegments.length; i++) {
      for (let j = i + 1; j < topSegments.length; j++) {
        const count = clients.filter(c =>
          c.segments.includes(topSegments[i].id) &&
          c.segments.includes(topSegments[j].id) &&
          c.segments.length === 2
        ).length;
        overlaps.push({ segments: [topSegments[i].id, topSegments[j].id], count });
      }
    }

    // Triple overlap
    if (topSegments.length === 3) {
      const count = clients.filter(c =>
        c.segments.includes(topSegments[0].id) &&
        c.segments.includes(topSegments[1].id) &&
        c.segments.includes(topSegments[2].id)
      ).length;
      overlaps.push({ segments: topSegments.map(s => s.id), count });
    }

    return overlaps;
  }, [segments, clients]);

  const handleCreateSegment = () => {
    setEditingSegment(null);
    setFormName('');
    setFormDescription('');
    setFormColor('#3B82F6');
    setFormRules([]);
    setShowCreateModal(true);
  };

  const handleEditSegment = (segment: Segment) => {
    setEditingSegment(segment);
    setFormName(segment.name);
    setFormDescription(segment.description);
    setFormColor(segment.color);
    setFormRules(segment.autoAssignRules);
    setShowCreateModal(true);
  };

  const handleSaveSegment = () => {
    if (!formName.trim()) return;

    const now = Date.now();
    if (editingSegment) {
      // Update existing segment
      setSegments(segments.map(seg =>
        seg.id === editingSegment.id
          ? { ...seg, name: formName, description: formDescription, color: formColor, autoAssignRules: formRules, updatedAt: now }
          : seg
      ));
    } else {
      // Create new segment
      const newSegment: Segment = {
        id: `seg-${Date.now()}`,
        name: formName,
        description: formDescription,
        color: formColor,
        clientCount: 0,
        autoAssignRules: formRules,
        createdAt: now,
        updatedAt: now
      };
      setSegments([...segments, newSegment]);
    }

    setShowCreateModal(false);
  };

  const handleDeleteSegment = (segmentId: string) => {
    if (confirm('Are you sure you want to delete this segment?')) {
      setSegments(segments.filter(s => s.id !== segmentId));
    }
  };

  const handleViewClients = (segment: Segment) => {
    setSelectedSegment(segment);
    setShowSegmentClients(true);
    setSelectedClients([]);
  };

  const handleBulkAction = (action: string) => {
    if (selectedClients.length === 0) {
      alert('Please select at least one client');
      return;
    }

    switch (action) {
      case 'email':
        alert(`Sending email to ${selectedClients.length} clients...`);
        break;
      case 'group':
        alert(`Changing group for ${selectedClients.length} clients...`);
        break;
      case 'commission':
        alert(`Applying commission override to ${selectedClients.length} clients...`);
        break;
    }
  };

  const addRule = () => {
    setFormRules([...formRules, { field: 'balance', operator: '>', value: 0 }]);
  };

  const updateRule = (index: number, updates: Partial<SegmentCondition>) => {
    const newRules = [...formRules];
    newRules[index] = { ...newRules[index], ...updates };
    setFormRules(newRules);
  };

  const removeRule = (index: number) => {
    setFormRules(formRules.filter((_, i) => i !== index));
  };

  const filteredClients = useMemo(() => {
    if (!selectedSegment) return [];

    return clients
      .filter(c => c.segments.includes(selectedSegment.id))
      .filter(c =>
        searchTerm === '' ||
        c.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        c.email.toLowerCase().includes(searchTerm.toLowerCase())
      );
  }, [clients, selectedSegment, searchTerm]);

  const toggleClientSelection = (clientId: string) => {
    if (selectedClients.includes(clientId)) {
      setSelectedClients(selectedClients.filter(id => id !== clientId));
    } else {
      setSelectedClients([...selectedClients, clientId]);
    }
  };

  const selectAllClients = () => {
    if (selectedClients.length === filteredClients.length) {
      setSelectedClients([]);
    } else {
      setSelectedClients(filteredClients.map(c => c.id));
    }
  };

  // Venn Diagram Component
  const VennDiagram: React.FC<{ segments: Segment[]; overlaps: SegmentOverlap[] }> = ({ segments: topSegments, overlaps }) => {
    if (topSegments.length < 3) return null;

    const width = 400;
    const height = 300;
    const radius = 80;

    // Circle centers for 3-way Venn
    const centers = [
      { x: width / 2 - 50, y: height / 2 - 30 },
      { x: width / 2 + 50, y: height / 2 - 30 },
      { x: width / 2, y: height / 2 + 40 }
    ];

    return (
      <svg width={width} height={height} className="mx-auto">
        {/* Circles */}
        {topSegments.slice(0, 3).map((seg, i) => (
          <circle
            key={seg.id}
            cx={centers[i].x}
            cy={centers[i].y}
            r={radius}
            fill={seg.color}
            fillOpacity={0.3}
            stroke={seg.color}
            strokeWidth={2}
          />
        ))}

        {/* Labels and counts */}
        {topSegments.slice(0, 3).map((seg, i) => {
          const labelOffset = i === 0 ? { x: -60, y: 0 } : i === 1 ? { x: 60, y: 0 } : { x: 0, y: 60 };
          const overlap = overlaps.find(o => o.segments.length === 1 && o.segments[0] === seg.id);

          return (
            <g key={`label-${seg.id}`}>
              <text
                x={centers[i].x + labelOffset.x}
                y={centers[i].y + labelOffset.y}
                textAnchor="middle"
                className="fill-zinc-300 text-xs font-medium"
              >
                {seg.name}
              </text>
              <text
                x={centers[i].x}
                y={centers[i].y}
                textAnchor="middle"
                className="fill-zinc-100 text-sm font-bold"
              >
                {overlap?.count || 0}
              </text>
            </g>
          );
        })}

        {/* Pairwise overlap counts */}
        {overlaps.filter(o => o.segments.length === 2).map((overlap, i) => {
          const positions = [
            { x: width / 2, y: height / 2 - 30 }, // Top middle (0-1)
            { x: width / 2 - 35, y: height / 2 + 20 }, // Bottom left (0-2)
            { x: width / 2 + 35, y: height / 2 + 20 }  // Bottom right (1-2)
          ];

          return (
            <text
              key={`overlap-${i}`}
              x={positions[i].x}
              y={positions[i].y}
              textAnchor="middle"
              className="fill-zinc-100 text-xs font-semibold"
            >
              {overlap.count}
            </text>
          );
        })}

        {/* Triple overlap */}
        {overlaps.find(o => o.segments.length === 3) && (
          <text
            x={width / 2}
            y={height / 2 + 5}
            textAnchor="middle"
            className="fill-zinc-100 text-xs font-semibold"
          >
            {overlaps.find(o => o.segments.length === 3)?.count || 0}
          </text>
        )}
      </svg>
    );
  };

  return (
    <div className="h-full flex flex-col bg-[#121316] text-zinc-300 p-4 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between mb-4 pb-3 border-b border-zinc-700">
        <div className="flex items-center gap-3">
          <Tag size={20} className="text-[#3B82F6]" />
          <h2 className="text-lg font-semibold text-zinc-100">Client Segmentation & Tags</h2>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setViewMode('cards')}
            className={`px-3 py-1.5 text-xs rounded transition-colors ${
              viewMode === 'cards'
                ? 'bg-[#3B82F6] text-white'
                : 'bg-[#1E2026] text-zinc-400 hover:bg-[#2A2D35]'
            }`}
          >
            Cards
          </button>
          <button
            onClick={() => setViewMode('tags')}
            className={`px-3 py-1.5 text-xs rounded transition-colors ${
              viewMode === 'tags'
                ? 'bg-[#3B82F6] text-white'
                : 'bg-[#1E2026] text-zinc-400 hover:bg-[#2A2D35]'
            }`}
          >
            Tag Cloud
          </button>
          <button
            onClick={handleCreateSegment}
            className="ml-2 px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs rounded transition-colors flex items-center gap-1.5"
          >
            <Plus size={14} />
            Create Segment
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-3 mb-4">
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Total Segments</div>
          <div className="text-2xl font-bold text-zinc-100">{totalSegments}</div>
        </div>
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Clients Tagged</div>
          <div className="text-2xl font-bold text-[#10B981]">{taggedClients}</div>
        </div>
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Untagged Clients</div>
          <div className="text-2xl font-bold text-[#EF4444]">{untaggedClients}</div>
        </div>
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-3">
          <div className="text-xs text-zinc-500 mb-1">Most Popular</div>
          <div className="text-sm font-semibold text-zinc-100 truncate">{mostPopularSegment?.name}</div>
          <div className="text-xs text-zinc-500">{mostPopularSegment?.clientCount} clients</div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto">
        {viewMode === 'cards' ? (
          <div className="space-y-4">
            {/* Segment Cards Grid */}
            <div className="grid grid-cols-4 gap-3">
              {segments.map(segment => (
                <div
                  key={segment.id}
                  className="bg-[#1E2026] border border-zinc-700 rounded p-3 hover:border-zinc-600 transition-all cursor-pointer group"
                  onClick={() => handleViewClients(segment)}
                >
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: segment.color }}
                      />
                      <h3 className="text-sm font-semibold text-zinc-100">{segment.name}</h3>
                    </div>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={(e) => { e.stopPropagation(); handleEditSegment(segment); }}
                        className="p-1 hover:bg-[#2A2D35] rounded"
                      >
                        <Edit2 size={12} className="text-zinc-400" />
                      </button>
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDeleteSegment(segment.id); }}
                        className="p-1 hover:bg-[#2A2D35] rounded"
                      >
                        <X size={12} className="text-[#EF4444]" />
                      </button>
                    </div>
                  </div>
                  <p className="text-xs text-zinc-500 mb-3 line-clamp-2">{segment.description}</p>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-1.5 text-zinc-400">
                      <Users size={12} />
                      <span className="text-xs">{segment.clientCount} clients</span>
                    </div>
                    {segment.autoAssignRules.length > 0 && (
                      <div className="text-xs text-[#3B82F6] flex items-center gap-1">
                        <Settings size={10} />
                        Auto
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>

            {/* Venn Diagram */}
            <div className="bg-[#1E2026] border border-zinc-700 rounded p-4 mt-4">
              <h3 className="text-sm font-semibold text-zinc-100 mb-3">Segment Overlap (Top 3)</h3>
              <VennDiagram segments={segments.slice(0, 3)} overlaps={segmentOverlaps} />
            </div>
          </div>
        ) : (
          /* Tag Cloud View */
          <div className="bg-[#1E2026] border border-zinc-700 rounded p-6">
            <div className="flex flex-wrap gap-3 items-center justify-center">
              {segments.map(segment => {
                const fontSize = Math.max(12, Math.min(32, 12 + segment.clientCount / 10));
                return (
                  <button
                    key={segment.id}
                    onClick={() => handleViewClients(segment)}
                    className="px-3 py-1.5 rounded-full transition-all hover:scale-110"
                    style={{
                      backgroundColor: segment.color + '33',
                      borderColor: segment.color,
                      borderWidth: '1px',
                      fontSize: `${fontSize}px`,
                      color: segment.color
                    }}
                  >
                    {segment.name} ({segment.clientCount})
                  </button>
                );
              })}
            </div>
          </div>
        )}
      </div>

      {/* Create/Edit Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowCreateModal(false)}>
          <div className="bg-[#1E2026] border border-zinc-700 rounded-lg w-[600px] max-h-[80vh] overflow-auto" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-zinc-700">
              <h3 className="text-lg font-semibold text-zinc-100">
                {editingSegment ? 'Edit Segment' : 'Create Segment'}
              </h3>
              <button onClick={() => setShowCreateModal(false)} className="text-zinc-500 hover:text-zinc-300">
                <X size={20} />
              </button>
            </div>

            <div className="p-4 space-y-4">
              {/* Name */}
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Segment Name</label>
                <input
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                  placeholder="e.g., VIP Traders"
                />
              </div>

              {/* Description */}
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Description</label>
                <textarea
                  value={formDescription}
                  onChange={(e) => setFormDescription(e.target.value)}
                  className="w-full bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6] resize-none"
                  rows={3}
                  placeholder="Brief description of this segment"
                />
              </div>

              {/* Color */}
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Color</label>
                <div className="flex items-center gap-3">
                  <input
                    type="color"
                    value={formColor}
                    onChange={(e) => setFormColor(e.target.value)}
                    className="w-12 h-10 bg-[#121316] border border-zinc-700 rounded cursor-pointer"
                  />
                  <input
                    type="text"
                    value={formColor}
                    onChange={(e) => setFormColor(e.target.value)}
                    className="flex-1 bg-[#121316] border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                  />
                </div>
              </div>

              {/* Auto-Assign Rules */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="text-xs text-zinc-400">Auto-Assign Rules</label>
                  <button
                    onClick={addRule}
                    className="text-xs text-[#3B82F6] hover:text-[#2563EB] flex items-center gap-1"
                  >
                    <Plus size={12} />
                    Add Rule
                  </button>
                </div>
                <div className="space-y-2">
                  {formRules.map((rule, index) => (
                    <div key={index} className="flex items-center gap-2">
                      <select
                        value={rule.field}
                        onChange={(e) => updateRule(index, { field: e.target.value as SegmentConditionField })}
                        className="flex-1 bg-[#121316] border border-zinc-700 rounded px-2 py-1.5 text-xs text-zinc-100"
                      >
                        <option value="balance">Balance</option>
                        <option value="volume">Volume</option>
                        <option value="lastLogin">Last Login (days ago)</option>
                        <option value="accountAge">Account Age (days)</option>
                        <option value="country">Country</option>
                        <option value="accountType">Account Type</option>
                      </select>
                      <select
                        value={rule.operator}
                        onChange={(e) => updateRule(index, { operator: e.target.value as SegmentConditionOperator })}
                        className="w-16 bg-[#121316] border border-zinc-700 rounded px-2 py-1.5 text-xs text-zinc-100"
                      >
                        <option value=">">{'>'}</option>
                        <option value="<">{'<'}</option>
                        <option value=">=">{'>='}</option>
                        <option value="<=">{'<='}</option>
                        <option value="=">=</option>
                      </select>
                      <input
                        type="text"
                        value={rule.value}
                        onChange={(e) => updateRule(index, { value: isNaN(Number(e.target.value)) ? e.target.value : Number(e.target.value) })}
                        className="flex-1 bg-[#121316] border border-zinc-700 rounded px-2 py-1.5 text-xs text-zinc-100"
                        placeholder="Value"
                      />
                      <button
                        onClick={() => removeRule(index)}
                        className="p-1.5 hover:bg-[#2A2D35] rounded"
                      >
                        <X size={14} className="text-[#EF4444]" />
                      </button>
                    </div>
                  ))}
                  {formRules.length === 0 && (
                    <div className="text-xs text-zinc-500 text-center py-2">
                      No rules defined. Clients must be manually assigned.
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="flex justify-end gap-2 p-4 border-t border-zinc-700">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 bg-[#2A2D35] hover:bg-[#35383F] text-zinc-300 text-sm rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveSegment}
                className="px-4 py-2 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-sm rounded transition-colors"
              >
                {editingSegment ? 'Save Changes' : 'Create Segment'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Segment Clients Modal */}
      {showSegmentClients && selectedSegment && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowSegmentClients(false)}>
          <div className="bg-[#1E2026] border border-zinc-700 rounded-lg w-[900px] max-h-[80vh] flex flex-col" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-zinc-700">
              <div className="flex items-center gap-3">
                <div
                  className="w-4 h-4 rounded-full"
                  style={{ backgroundColor: selectedSegment.color }}
                />
                <h3 className="text-lg font-semibold text-zinc-100">{selectedSegment.name}</h3>
                <span className="text-sm text-zinc-500">({filteredClients.length} clients)</span>
              </div>
              <button onClick={() => setShowSegmentClients(false)} className="text-zinc-500 hover:text-zinc-300">
                <X size={20} />
              </button>
            </div>

            {/* Search and Actions */}
            <div className="p-4 border-b border-zinc-700 space-y-3">
              <div className="flex items-center gap-3">
                <div className="flex-1 relative">
                  <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    placeholder="Search clients..."
                    className="w-full bg-[#121316] border border-zinc-700 rounded pl-10 pr-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-[#3B82F6]"
                  />
                </div>
                <button className="px-3 py-2 bg-[#2A2D35] hover:bg-[#35383F] text-zinc-300 text-sm rounded transition-colors flex items-center gap-2">
                  <Download size={14} />
                  Export
                </button>
              </div>

              {selectedClients.length > 0 && (
                <div className="flex items-center gap-2">
                  <span className="text-sm text-zinc-400">{selectedClients.length} selected</span>
                  <button
                    onClick={() => handleBulkAction('email')}
                    className="px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs rounded transition-colors flex items-center gap-1"
                  >
                    <Mail size={12} />
                    Send Email
                  </button>
                  <button
                    onClick={() => handleBulkAction('group')}
                    className="px-3 py-1.5 bg-[#10B981] hover:bg-[#059669] text-white text-xs rounded transition-colors flex items-center gap-1"
                  >
                    <Users size={12} />
                    Change Group
                  </button>
                  <button
                    onClick={() => handleBulkAction('commission')}
                    className="px-3 py-1.5 bg-[#F59E0B] hover:bg-[#D97706] text-white text-xs rounded transition-colors flex items-center gap-1"
                  >
                    <TrendingUp size={12} />
                    Commission Override
                  </button>
                </div>
              )}
            </div>

            {/* Client List Table */}
            <div className="flex-1 overflow-auto p-4">
              <table className="w-full text-sm">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-zinc-700">
                  <tr className="text-left text-xs text-zinc-500">
                    <th className="pb-2 pl-2">
                      <input
                        type="checkbox"
                        checked={selectedClients.length === filteredClients.length && filteredClients.length > 0}
                        onChange={selectAllClients}
                        className="rounded border-zinc-600 bg-[#121316]"
                      />
                    </th>
                    <th className="pb-2">Client</th>
                    <th className="pb-2">Email</th>
                    <th className="pb-2 text-right">Balance</th>
                    <th className="pb-2 text-right">Volume</th>
                    <th className="pb-2">Last Login</th>
                    <th className="pb-2">Account Type</th>
                    <th className="pb-2">Country</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredClients.map(client => {
                    const daysSinceLogin = Math.floor((Date.now() - client.lastLogin) / 86400000);
                    return (
                      <tr
                        key={client.id}
                        className="border-b border-zinc-800 hover:bg-[#2A2D35] transition-colors"
                      >
                        <td className="py-2 pl-2">
                          <input
                            type="checkbox"
                            checked={selectedClients.includes(client.id)}
                            onChange={() => toggleClientSelection(client.id)}
                            className="rounded border-zinc-600 bg-[#121316]"
                          />
                        </td>
                        <td className="py-2 text-zinc-100">{client.name}</td>
                        <td className="py-2 text-zinc-400">{client.email}</td>
                        <td className="py-2 text-right text-zinc-100">${client.balance.toFixed(2)}</td>
                        <td className="py-2 text-right text-zinc-400">${client.volume.toFixed(0)}</td>
                        <td className="py-2 text-zinc-400 flex items-center gap-1">
                          <Clock size={12} />
                          {daysSinceLogin}d ago
                        </td>
                        <td className="py-2 text-zinc-400">{client.accountType}</td>
                        <td className="py-2 text-zinc-400 flex items-center gap-1">
                          <MapPin size={12} />
                          {client.country}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>

              {filteredClients.length === 0 && (
                <div className="text-center py-12 text-zinc-500">
                  No clients found in this segment
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ClientSegmentation;
