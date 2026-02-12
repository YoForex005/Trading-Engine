'use client';

import { useState, useMemo } from 'react';
import {
  Mail,
  Grid,
  List,
  Edit,
  Eye,
  Send,
  Plus,
  Save,
  X,
  Code,
  BarChart3,
  Copy,
  Archive,
  FileText,
  Filter
} from 'lucide-react';

type TemplateCategory = 'all' | 'account' | 'trading' | 'payment' | 'security' | 'marketing' | 'system';
type TemplateStatus = 'active' | 'draft' | 'archived';
type ViewMode = 'grid' | 'list';
type EditorView = 'code' | 'preview';

interface EmailTemplate {
  id: string;
  name: string;
  slug: string;
  category: Exclude<TemplateCategory, 'all'>;
  subject: string;
  htmlBody: string;
  plainBody: string;
  status: TemplateStatus;
  lastModified: string;
  stats: {
    sent: number;
    openRate: number;
    clickRate: number;
  };
}

interface MergeTag {
  tag: string;
  label: string;
  description: string;
}

const MERGE_TAGS: MergeTag[] = [
  { tag: '{{client_name}}', label: 'Client Name', description: 'Full name of the client' },
  { tag: '{{client_email}}', label: 'Client Email', description: 'Client email address' },
  { tag: '{{account_number}}', label: 'Account Number', description: 'Trading account number' },
  { tag: '{{amount}}', label: 'Amount', description: 'Transaction amount' },
  { tag: '{{currency}}', label: 'Currency', description: 'Currency code (USD, EUR, etc.)' },
  { tag: '{{symbol}}', label: 'Symbol', description: 'Trading symbol (EURUSD, etc.)' },
  { tag: '{{balance}}', label: 'Balance', description: 'Account balance' },
  { tag: '{{equity}}', label: 'Equity', description: 'Account equity' },
  { tag: '{{date}}', label: 'Date', description: 'Current date' },
  { tag: '{{time}}', label: 'Time', description: 'Current time' },
  { tag: '{{transaction_id}}', label: 'Transaction ID', description: 'Unique transaction ID' },
  { tag: '{{support_email}}', label: 'Support Email', description: 'Platform support email' },
  { tag: '{{platform_name}}', label: 'Platform Name', description: 'RTX5 Platform' },
  { tag: '{{login_url}}', label: 'Login URL', description: 'Link to login page' },
];

const SAMPLE_DATA = {
  client_name: 'John Smith',
  client_email: 'john.smith@example.com',
  account_number: '680962851',
  amount: '$5,000.00',
  currency: 'USD',
  symbol: 'EURUSD',
  balance: '$25,430.50',
  equity: '$25,280.75',
  date: new Date().toLocaleDateString(),
  time: new Date().toLocaleTimeString(),
  transaction_id: 'TXN000123',
  support_email: 'support@rtx5.com',
  platform_name: 'RTX5 Trading Platform',
  login_url: 'https://rtx5.com/login',
};

function generateMockTemplates(): EmailTemplate[] {
  const templates: Omit<EmailTemplate, 'id' | 'lastModified' | 'stats'>[] = [
    // Account templates
    {
      name: 'Welcome Email',
      slug: 'welcome',
      category: 'account',
      subject: 'Welcome to {{platform_name}}',
      htmlBody: '<h1>Welcome {{client_name}}!</h1><p>Your account {{account_number}} has been successfully created.</p><p><a href="{{login_url}}">Login Now</a></p>',
      plainBody: 'Welcome {{client_name}}! Your account {{account_number}} has been successfully created. Login at: {{login_url}}',
      status: 'active',
    },
    {
      name: 'Account Verification',
      slug: 'account_verification',
      category: 'account',
      subject: 'Verify Your Email Address',
      htmlBody: '<h2>Email Verification Required</h2><p>Hi {{client_name}},</p><p>Please verify your email to activate your account.</p>',
      plainBody: 'Hi {{client_name}}, Please verify your email to activate your account.',
      status: 'active',
    },
    {
      name: 'Password Reset',
      slug: 'password_reset',
      category: 'security',
      subject: 'Reset Your Password',
      htmlBody: '<h2>Password Reset Request</h2><p>Hi {{client_name}},</p><p>Click the link below to reset your password.</p>',
      plainBody: 'Hi {{client_name}}, Click the link to reset your password.',
      status: 'active',
    },
    {
      name: 'Account Approved',
      slug: 'account_approved',
      category: 'account',
      subject: 'Your Account Has Been Approved',
      htmlBody: '<h1>Congratulations!</h1><p>{{client_name}}, your account has been approved. Start trading now!</p>',
      plainBody: '{{client_name}}, your account has been approved. Start trading now!',
      status: 'active',
    },
    // Trading templates
    {
      name: 'Trade Executed',
      slug: 'trade_executed',
      category: 'trading',
      subject: 'Trade Executed: {{symbol}}',
      htmlBody: '<h2>Trade Confirmation</h2><p>Your trade on {{symbol}} has been executed. Amount: {{amount}}</p>',
      plainBody: 'Your trade on {{symbol}} has been executed. Amount: {{amount}}',
      status: 'active',
    },
    {
      name: 'Margin Call Warning',
      slug: 'margin_call',
      category: 'trading',
      subject: 'Urgent: Margin Call Warning',
      htmlBody: '<h2 style="color: red;">Margin Call Warning</h2><p>{{client_name}}, your account equity is below margin requirements.</p>',
      plainBody: '{{client_name}}, your account equity is below margin requirements.',
      status: 'active',
    },
    {
      name: 'Stop Loss Triggered',
      slug: 'stop_loss',
      category: 'trading',
      subject: 'Stop Loss Triggered on {{symbol}}',
      htmlBody: '<h2>Stop Loss Executed</h2><p>Your stop loss order on {{symbol}} has been triggered.</p>',
      plainBody: 'Your stop loss order on {{symbol}} has been triggered.',
      status: 'active',
    },
    {
      name: 'Daily Trading Summary',
      slug: 'daily_summary',
      category: 'trading',
      subject: 'Your Daily Trading Summary',
      htmlBody: '<h2>Daily Summary for {{date}}</h2><p>Balance: {{balance}} | Equity: {{equity}}</p>',
      plainBody: 'Daily Summary for {{date}}. Balance: {{balance}} | Equity: {{equity}}',
      status: 'active',
    },
    // Payment templates
    {
      name: 'Deposit Received',
      slug: 'deposit_received',
      category: 'payment',
      subject: 'Deposit Received: {{amount}}',
      htmlBody: '<h2>Deposit Confirmation</h2><p>We have received your deposit of {{amount}}. Transaction ID: {{transaction_id}}</p>',
      plainBody: 'We have received your deposit of {{amount}}. Transaction ID: {{transaction_id}}',
      status: 'active',
    },
    {
      name: 'Withdrawal Approved',
      slug: 'withdrawal_approved',
      category: 'payment',
      subject: 'Withdrawal Approved: {{amount}}',
      htmlBody: '<h2>Withdrawal Approved</h2><p>Your withdrawal of {{amount}} has been approved and will be processed shortly.</p>',
      plainBody: 'Your withdrawal of {{amount}} has been approved and will be processed shortly.',
      status: 'active',
    },
    {
      name: 'Withdrawal Rejected',
      slug: 'withdrawal_rejected',
      category: 'payment',
      subject: 'Withdrawal Request Declined',
      htmlBody: '<h2>Withdrawal Declined</h2><p>Your withdrawal request for {{amount}} has been declined. Contact support for details.</p>',
      plainBody: 'Your withdrawal request for {{amount}} has been declined. Contact support: {{support_email}}',
      status: 'active',
    },
    {
      name: 'Payment Pending',
      slug: 'payment_pending',
      category: 'payment',
      subject: 'Payment Processing: {{amount}}',
      htmlBody: '<h2>Payment in Progress</h2><p>Your payment of {{amount}} is being processed. Transaction: {{transaction_id}}</p>',
      plainBody: 'Your payment of {{amount}} is being processed. Transaction: {{transaction_id}}',
      status: 'active',
    },
    // Security templates
    {
      name: 'Login Alert',
      slug: 'login_alert',
      category: 'security',
      subject: 'New Login Detected',
      htmlBody: '<h2>Security Alert</h2><p>A new login was detected on your account at {{time}} on {{date}}.</p>',
      plainBody: 'A new login was detected on your account at {{time}} on {{date}}.',
      status: 'active',
    },
    {
      name: 'Two-Factor Authentication',
      slug: '2fa_setup',
      category: 'security',
      subject: 'Enable Two-Factor Authentication',
      htmlBody: '<h2>Secure Your Account</h2><p>{{client_name}}, enable 2FA to add extra security to your account.</p>',
      plainBody: '{{client_name}}, enable 2FA to add extra security to your account.',
      status: 'draft',
    },
    {
      name: 'Suspicious Activity',
      slug: 'suspicious_activity',
      category: 'security',
      subject: 'Unusual Account Activity Detected',
      htmlBody: '<h2 style="color: red;">Security Alert</h2><p>Unusual activity detected on account {{account_number}}.</p>',
      plainBody: 'Unusual activity detected on account {{account_number}}. Please review immediately.',
      status: 'active',
    },
    // Marketing templates
    {
      name: 'Monthly Newsletter',
      slug: 'newsletter',
      category: 'marketing',
      subject: 'RTX5 Monthly Newsletter - {{date}}',
      htmlBody: '<h1>Market Insights</h1><p>Hi {{client_name}}, check out this month\'s trading opportunities!</p>',
      plainBody: 'Hi {{client_name}}, check out this month\'s trading opportunities!',
      status: 'active',
    },
    {
      name: 'Promotion Announcement',
      slug: 'promotion',
      category: 'marketing',
      subject: 'Special Offer: Reduced Spreads',
      htmlBody: '<h1>Limited Time Offer!</h1><p>Trade with reduced spreads this week. Don\'t miss out!</p>',
      plainBody: 'Limited Time Offer! Trade with reduced spreads this week.',
      status: 'draft',
    },
    {
      name: 'Webinar Invitation',
      slug: 'webinar',
      category: 'marketing',
      subject: 'Invitation: Trading Strategies Webinar',
      htmlBody: '<h2>Join Our Webinar</h2><p>Learn advanced trading strategies. Register now!</p>',
      plainBody: 'Learn advanced trading strategies. Register for our webinar now!',
      status: 'archived',
    },
    // System templates
    {
      name: 'Maintenance Notice',
      slug: 'maintenance',
      category: 'system',
      subject: 'Scheduled Maintenance on {{date}}',
      htmlBody: '<h2>System Maintenance</h2><p>Our platform will undergo maintenance on {{date}} at {{time}}.</p>',
      plainBody: 'Our platform will undergo maintenance on {{date}} at {{time}}.',
      status: 'active',
    },
    {
      name: 'System Update',
      slug: 'system_update',
      category: 'system',
      subject: 'Platform Update: New Features',
      htmlBody: '<h2>What\'s New</h2><p>We\'ve added new features to enhance your trading experience!</p>',
      plainBody: 'We\'ve added new features to enhance your trading experience!',
      status: 'active',
    },
  ];

  return templates.map((t, index) => ({
    ...t,
    id: `TPL${String(index + 1).padStart(4, '0')}`,
    lastModified: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
    stats: {
      sent: Math.floor(Math.random() * 10000 + 500),
      openRate: Math.random() * 40 + 50,
      clickRate: Math.random() * 15 + 10,
    },
  }));
}

function replaceMergeTags(text: string, data: Record<string, string>): string {
  let result = text;
  Object.entries(data).forEach(([key, value]) => {
    result = result.replace(new RegExp(`{{${key}}}`, 'g'), value);
  });
  return result;
}

export default function EmailTemplateManager() {
  const [templates] = useState<EmailTemplate[]>(generateMockTemplates());
  const [selectedCategory, setSelectedCategory] = useState<TemplateCategory>('all');
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [editingTemplate, setEditingTemplate] = useState<EmailTemplate | null>(null);
  const [editorView, setEditorView] = useState<EditorView>('code');
  const [searchQuery, setSearchQuery] = useState('');
  const [showNewTemplateForm, setShowNewTemplateForm] = useState(false);

  const categories: { id: TemplateCategory; label: string; count: number }[] = useMemo(() => {
    return [
      { id: 'all', label: 'All Templates', count: templates.length },
      { id: 'account', label: 'Account', count: templates.filter(t => t.category === 'account').length },
      { id: 'trading', label: 'Trading', count: templates.filter(t => t.category === 'trading').length },
      { id: 'payment', label: 'Payment', count: templates.filter(t => t.category === 'payment').length },
      { id: 'security', label: 'Security', count: templates.filter(t => t.category === 'security').length },
      { id: 'marketing', label: 'Marketing', count: templates.filter(t => t.category === 'marketing').length },
      { id: 'system', label: 'System', count: templates.filter(t => t.category === 'system').length },
    ];
  }, [templates]);

  const filteredTemplates = useMemo(() => {
    return templates.filter(t => {
      const matchesCategory = selectedCategory === 'all' || t.category === selectedCategory;
      const matchesSearch = searchQuery === '' ||
        t.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        t.subject.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesCategory && matchesSearch;
    });
  }, [templates, selectedCategory, searchQuery]);

  const handleEditTemplate = (template: EmailTemplate) => {
    setEditingTemplate({ ...template });
    setEditorView('code');
  };

  const handleSaveTemplate = () => {
    // In real app, would save to backend
    console.log('Saving template:', editingTemplate);
    setEditingTemplate(null);
  };

  const handleDiscardChanges = () => {
    setEditingTemplate(null);
  };

  const handleSendTest = () => {
    console.log('Sending test email for template:', editingTemplate?.id);
    alert('Test email sent to admin@rtx5.com');
  };

  const handleInsertTag = (tag: string) => {
    if (!editingTemplate) return;
    // Insert at cursor position (simplified - in real app would track cursor)
    setEditingTemplate({
      ...editingTemplate,
      htmlBody: editingTemplate.htmlBody + ' ' + tag,
    });
  };

  const getCategoryBadgeColor = (category: Exclude<TemplateCategory, 'all'>) => {
    const colors = {
      account: 'bg-blue-500/20 text-blue-400',
      trading: 'bg-green-500/20 text-green-400',
      payment: 'bg-yellow-500/20 text-yellow-400',
      security: 'bg-red-500/20 text-red-400',
      marketing: 'bg-purple-500/20 text-purple-400',
      system: 'bg-gray-500/20 text-gray-400',
    };
    return colors[category];
  };

  const getStatusColor = (status: TemplateStatus) => {
    const colors = {
      active: 'text-green-400',
      draft: 'text-yellow-400',
      archived: 'text-gray-500',
    };
    return colors[status];
  };

  if (editingTemplate) {
    const previewHtml = replaceMergeTags(editingTemplate.htmlBody, SAMPLE_DATA);

    return (
      <div className="h-full bg-[#121316] text-[#E0E0E0] flex flex-col">
        {/* Editor Header */}
        <div className="h-14 bg-[#1E2026] border-b border-[#383A42] flex items-center justify-between px-6">
          <div className="flex items-center gap-4">
            <button
              onClick={handleDiscardChanges}
              className="p-2 hover:bg-[#25272E] rounded transition-colors"
            >
              <X size={20} />
            </button>
            <div>
              <h2 className="text-lg font-semibold">{editingTemplate.name}</h2>
              <div className="flex items-center gap-2 text-xs text-[#888]">
                <span className={`px-2 py-0.5 rounded text-xs font-medium ${getCategoryBadgeColor(editingTemplate.category)}`}>
                  {editingTemplate.category.toUpperCase()}
                </span>
                <span className={getStatusColor(editingTemplate.status)}>
                  {editingTemplate.status.toUpperCase()}
                </span>
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={handleSendTest}
              className="px-4 py-2 bg-[#25272E] hover:bg-[#2C2E36] text-[#888] rounded flex items-center gap-2 transition-colors"
            >
              <Send size={16} />
              Send Test
            </button>
            <button
              onClick={handleSaveTemplate}
              className="px-4 py-2 bg-[#F5C542] hover:bg-[#F5C542]/90 text-[#000] rounded flex items-center gap-2 transition-colors font-semibold"
            >
              <Save size={16} />
              Save Changes
            </button>
          </div>
        </div>

        {/* Editor Body */}
        <div className="flex-1 flex overflow-hidden">
          {/* Editor Panel */}
          <div className="flex-1 flex flex-col border-r border-[#383A42]">
            {/* Subject Line */}
            <div className="p-4 border-b border-[#383A42]">
              <label className="block text-xs text-[#888] mb-1">Subject Line</label>
              <input
                type="text"
                value={editingTemplate.subject}
                onChange={(e) => setEditingTemplate({ ...editingTemplate, subject: e.target.value })}
                className="w-full bg-[#1E2026] border border-[#383A42] rounded px-3 py-2 text-sm focus:outline-none focus:border-[#F5C542]"
              />
            </div>

            {/* Code/Preview Toggle */}
            <div className="flex items-center gap-2 px-4 py-2 border-b border-[#383A42] bg-[#1A1C22]">
              <button
                onClick={() => setEditorView('code')}
                className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                  editorView === 'code'
                    ? 'bg-[#F5C542] text-[#000]'
                    : 'bg-[#25272E] text-[#888] hover:bg-[#2C2E36]'
                }`}
              >
                <Code size={14} className="inline mr-1" />
                Code
              </button>
              <button
                onClick={() => setEditorView('preview')}
                className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                  editorView === 'preview'
                    ? 'bg-[#F5C542] text-[#000]'
                    : 'bg-[#25272E] text-[#888] hover:bg-[#2C2E36]'
                }`}
              >
                <Eye size={14} className="inline mr-1" />
                Preview
              </button>
            </div>

            {/* HTML Body Editor */}
            <div className="flex-1 overflow-auto">
              {editorView === 'code' ? (
                <textarea
                  value={editingTemplate.htmlBody}
                  onChange={(e) => setEditingTemplate({ ...editingTemplate, htmlBody: e.target.value })}
                  className="w-full h-full bg-[#121316] border-0 p-4 text-sm font-mono focus:outline-none resize-none"
                  placeholder="Enter HTML content..."
                />
              ) : (
                <div className="p-4 bg-white text-black" dangerouslySetInnerHTML={{ __html: previewHtml }} />
              )}
            </div>

            {/* Plain Text Body */}
            <div className="border-t border-[#383A42] p-4">
              <label className="block text-xs text-[#888] mb-1">Plain Text Body</label>
              <textarea
                value={editingTemplate.plainBody}
                onChange={(e) => setEditingTemplate({ ...editingTemplate, plainBody: e.target.value })}
                className="w-full bg-[#1E2026] border border-[#383A42] rounded px-3 py-2 text-sm focus:outline-none focus:border-[#F5C542] resize-none"
                rows={3}
                placeholder="Enter plain text version..."
              />
            </div>
          </div>

          {/* Variables Sidebar */}
          <div className="w-80 bg-[#1A1C22] flex flex-col">
            <div className="p-4 border-b border-[#383A42]">
              <h3 className="text-sm font-semibold mb-1">Available Variables</h3>
              <p className="text-xs text-[#888]">Click to insert into template</p>
            </div>
            <div className="flex-1 overflow-auto p-4 space-y-2">
              {MERGE_TAGS.map((tag) => (
                <button
                  key={tag.tag}
                  onClick={() => handleInsertTag(tag.tag)}
                  className="w-full text-left p-3 bg-[#25272E] hover:bg-[#2C2E36] rounded border border-[#383A42] transition-colors group"
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-xs font-mono text-[#F5C542]">{tag.tag}</span>
                    <Copy size={14} className="text-[#888] opacity-0 group-hover:opacity-100 transition-opacity" />
                  </div>
                  <div className="text-xs font-medium text-[#E0E0E0]">{tag.label}</div>
                  <div className="text-xs text-[#888] mt-1">{tag.description}</div>
                </button>
              ))}
            </div>

            {/* Template Stats */}
            <div className="border-t border-[#383A42] p-4 space-y-3">
              <h3 className="text-xs font-semibold text-[#888]">Template Statistics</h3>
              <div className="space-y-2">
                <div className="flex justify-between text-xs">
                  <span className="text-[#888]">Emails Sent:</span>
                  <span className="font-semibold">{editingTemplate.stats.sent.toLocaleString()}</span>
                </div>
                <div className="flex justify-between text-xs">
                  <span className="text-[#888]">Open Rate:</span>
                  <span className="font-semibold text-green-400">{editingTemplate.stats.openRate.toFixed(1)}%</span>
                </div>
                <div className="flex justify-between text-xs">
                  <span className="text-[#888]">Click Rate:</span>
                  <span className="font-semibold text-blue-400">{editingTemplate.stats.clickRate.toFixed(1)}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full bg-[#121316] text-[#E0E0E0] flex flex-col">
      {/* Header */}
      <div className="h-14 bg-[#1E2026] border-b border-[#383A42] flex items-center justify-between px-6">
        <div className="flex items-center gap-4">
          <Mail size={20} className="text-[#F5C542]" />
          <h1 className="text-lg font-semibold">Email Templates</h1>
          <span className="text-xs text-[#888]">{filteredTemplates.length} templates</span>
        </div>
        <div className="flex items-center gap-2">
          <input
            type="text"
            placeholder="Search templates..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-64 bg-[#25272E] border border-[#383A42] rounded px-3 py-1.5 text-sm focus:outline-none focus:border-[#F5C542]"
          />
          <div className="flex items-center gap-1 bg-[#25272E] rounded p-1">
            <button
              onClick={() => setViewMode('grid')}
              className={`p-1.5 rounded transition-colors ${
                viewMode === 'grid' ? 'bg-[#F5C542] text-[#000]' : 'text-[#888] hover:text-[#E0E0E0]'
              }`}
            >
              <Grid size={16} />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`p-1.5 rounded transition-colors ${
                viewMode === 'list' ? 'bg-[#F5C542] text-[#000]' : 'text-[#888] hover:text-[#E0E0E0]'
              }`}
            >
              <List size={16} />
            </button>
          </div>
          <button
            onClick={() => setShowNewTemplateForm(true)}
            className="px-4 py-1.5 bg-[#F5C542] hover:bg-[#F5C542]/90 text-[#000] rounded flex items-center gap-2 transition-colors font-semibold text-sm"
          >
            <Plus size={16} />
            New Template
          </button>
        </div>
      </div>

      {/* Category Tabs */}
      <div className="bg-[#1A1C22] border-b border-[#383A42] px-6">
        <div className="flex items-center gap-1 -mb-px">
          {categories.map((cat) => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(cat.id)}
              className={`px-4 py-3 text-sm font-medium transition-colors border-b-2 ${
                selectedCategory === cat.id
                  ? 'text-[#F5C542] border-[#F5C542]'
                  : 'text-[#888] border-transparent hover:text-[#E0E0E0]'
              }`}
            >
              {cat.label}
              <span className="ml-2 text-xs opacity-60">({cat.count})</span>
            </button>
          ))}
        </div>
      </div>

      {/* Templates Grid/List */}
      <div className="flex-1 overflow-auto p-6">
        {viewMode === 'grid' ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {filteredTemplates.map((template) => (
              <div
                key={template.id}
                className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4 hover:border-[#F5C542] transition-colors cursor-pointer group"
                onClick={() => handleEditTemplate(template)}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className={`px-2 py-1 rounded text-xs font-medium ${getCategoryBadgeColor(template.category)}`}>
                    {template.category.toUpperCase()}
                  </div>
                  <Edit size={16} className="text-[#888] opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
                <h3 className="font-semibold mb-2 text-sm">{template.name}</h3>
                <p className="text-xs text-[#888] mb-3 line-clamp-2">{template.subject}</p>
                <div className="flex items-center justify-between text-xs">
                  <span className={getStatusColor(template.status)}>{template.status}</span>
                  <span className="text-[#888]">
                    {new Date(template.lastModified).toLocaleDateString()}
                  </span>
                </div>
                <div className="mt-3 pt-3 border-t border-[#383A42] grid grid-cols-3 gap-2 text-xs">
                  <div>
                    <div className="text-[#888]">Sent</div>
                    <div className="font-semibold">{template.stats.sent.toLocaleString()}</div>
                  </div>
                  <div>
                    <div className="text-[#888]">Open</div>
                    <div className="font-semibold text-green-400">{template.stats.openRate.toFixed(0)}%</div>
                  </div>
                  <div>
                    <div className="text-[#888]">Click</div>
                    <div className="font-semibold text-blue-400">{template.stats.clickRate.toFixed(0)}%</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="space-y-2">
            {filteredTemplates.map((template) => (
              <div
                key={template.id}
                className="bg-[#1E2026] border border-[#383A42] rounded-lg p-4 hover:border-[#F5C542] transition-colors cursor-pointer group flex items-center gap-4"
                onClick={() => handleEditTemplate(template)}
              >
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <h3 className="font-semibold text-sm">{template.name}</h3>
                    <div className={`px-2 py-0.5 rounded text-xs font-medium ${getCategoryBadgeColor(template.category)}`}>
                      {template.category}
                    </div>
                    <span className={`text-xs ${getStatusColor(template.status)}`}>{template.status}</span>
                  </div>
                  <p className="text-xs text-[#888]">{template.subject}</p>
                </div>
                <div className="flex items-center gap-6 text-xs">
                  <div className="text-center">
                    <div className="text-[#888]">Sent</div>
                    <div className="font-semibold">{template.stats.sent.toLocaleString()}</div>
                  </div>
                  <div className="text-center">
                    <div className="text-[#888]">Open Rate</div>
                    <div className="font-semibold text-green-400">{template.stats.openRate.toFixed(1)}%</div>
                  </div>
                  <div className="text-center">
                    <div className="text-[#888]">Click Rate</div>
                    <div className="font-semibold text-blue-400">{template.stats.clickRate.toFixed(1)}%</div>
                  </div>
                  <div className="text-center">
                    <div className="text-[#888]">Modified</div>
                    <div className="font-semibold">{new Date(template.lastModified).toLocaleDateString()}</div>
                  </div>
                </div>
                <Edit size={18} className="text-[#888] opacity-0 group-hover:opacity-100 transition-opacity" />
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
