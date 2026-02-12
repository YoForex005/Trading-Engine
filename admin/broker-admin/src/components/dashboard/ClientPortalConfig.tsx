/**
 * Client Portal / Self-Service Configuration
 * Admin page for configuring the client-facing self-service portal
 */

'use client';

import React, { useState } from 'react';
import {
  Globe,
  Users,
  FileText,
  MessageCircle,
  Eye,
  Upload,
  Download,
  Lock,
  Shield,
  CreditCard,
  Clock,
  Mail,
  Phone,
  MessageSquare,
  Image,
  Palette,
  Save,
  RotateCcw,
  CheckCircle,
  XCircle,
  AlertCircle,
  TrendingUp,
  Ticket,
} from 'lucide-react';

interface PortalSettings {
  enabled: boolean;
  logoUrl: string;
  primaryColor: string;
  companyName: string;
  welcomeMessage: string;
  termsUrl: string;
}

interface SelfServiceFeature {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  icon: React.ReactNode;
}

interface DocumentRequirement {
  id: string;
  name: string;
  required: boolean;
  acceptedFormats: string[];
  maxSizeMB: number;
}

interface SupportConfig {
  email: string;
  phone: string;
  liveChatEnabled: boolean;
  workingHours: string;
  autoReplyMessage: string;
}

interface Stats {
  portalVisitorsToday: number;
  openTickets: number;
  pendingKycUploads: number;
  selfServiceActionsToday: number;
}

export default function ClientPortalConfig() {
  const [stats] = useState<Stats>({
    portalVisitorsToday: 342,
    openTickets: 18,
    pendingKycUploads: 23,
    selfServiceActionsToday: 156,
  });

  const [portalSettings, setPortalSettings] = useState<PortalSettings>({
    enabled: true,
    logoUrl: 'https://example.com/logo.png',
    primaryColor: '#3B82F6',
    companyName: 'RTX5 Trading',
    welcomeMessage: 'Welcome to your client portal! Manage your account, upload documents, and get support - all in one place.',
    termsUrl: 'https://example.com/terms',
  });

  const [selfServiceFeatures, setSelfServiceFeatures] = useState<SelfServiceFeature[]>([
    {
      id: 'kyc-upload',
      name: 'KYC Document Upload',
      description: 'Allow clients to upload verification documents directly',
      enabled: true,
      icon: <Upload size={16} />,
    },
    {
      id: 'deposit-requests',
      name: 'Deposit Requests',
      description: 'Enable clients to initiate deposit transactions',
      enabled: true,
      icon: <CreditCard size={16} />,
    },
    {
      id: 'withdrawal-requests',
      name: 'Withdrawal Requests',
      description: 'Allow clients to submit withdrawal requests',
      enabled: true,
      icon: <Download size={16} />,
    },
    {
      id: 'password-change',
      name: 'Password Change',
      description: 'Let clients reset their password independently',
      enabled: true,
      icon: <Lock size={16} />,
    },
    {
      id: 'twofa-setup',
      name: '2FA Setup',
      description: 'Enable clients to configure two-factor authentication',
      enabled: true,
      icon: <Shield size={16} />,
    },
    {
      id: 'support-tickets',
      name: 'Support Ticket Creation',
      description: 'Allow clients to open support tickets',
      enabled: true,
      icon: <MessageCircle size={16} />,
    },
    {
      id: 'account-statement',
      name: 'Account Statement Download',
      description: 'Enable account statement PDF downloads',
      enabled: true,
      icon: <FileText size={16} />,
    },
    {
      id: 'leverage-change',
      name: 'Leverage Change Request',
      description: 'Allow clients to request leverage adjustments',
      enabled: false,
      icon: <TrendingUp size={16} />,
    },
  ]);

  const [documentRequirements, setDocumentRequirements] = useState<DocumentRequirement[]>([
    {
      id: 'id-proof',
      name: 'ID Proof (Passport, Driver License, National ID)',
      required: true,
      acceptedFormats: ['PDF', 'JPG', 'PNG'],
      maxSizeMB: 5,
    },
    {
      id: 'address-proof',
      name: 'Address Proof (Utility Bill, Bank Statement)',
      required: true,
      acceptedFormats: ['PDF', 'JPG', 'PNG'],
      maxSizeMB: 5,
    },
    {
      id: 'selfie',
      name: 'Selfie with ID Document',
      required: true,
      acceptedFormats: ['JPG', 'PNG'],
      maxSizeMB: 3,
    },
    {
      id: 'bank-statement',
      name: 'Bank Statement (Last 3 months)',
      required: false,
      acceptedFormats: ['PDF'],
      maxSizeMB: 10,
    },
  ]);

  const [supportConfig, setSupportConfig] = useState<SupportConfig>({
    email: 'support@rtx5trading.com',
    phone: '+1-800-RTX-5000',
    liveChatEnabled: true,
    workingHours: 'Monday - Friday, 9:00 AM - 6:00 PM EST',
    autoReplyMessage: 'Thank you for contacting RTX5 Trading support. We have received your message and will respond within 24 hours.',
  });

  const [hasChanges, setHasChanges] = useState(false);

  const handlePortalSettingChange = (field: keyof PortalSettings, value: any) => {
    setPortalSettings((prev) => ({ ...prev, [field]: value }));
    setHasChanges(true);
  };

  const handleFeatureToggle = (featureId: string) => {
    setSelfServiceFeatures((prev) =>
      prev.map((f) => (f.id === featureId ? { ...f, enabled: !f.enabled } : f))
    );
    setHasChanges(true);
  };

  const handleDocumentRequirementChange = (docId: string, field: keyof DocumentRequirement, value: any) => {
    setDocumentRequirements((prev) =>
      prev.map((d) => (d.id === docId ? { ...d, [field]: value } : d))
    );
    setHasChanges(true);
  };

  const handleSupportConfigChange = (field: keyof SupportConfig, value: any) => {
    setSupportConfig((prev) => ({ ...prev, [field]: value }));
    setHasChanges(true);
  };

  const handleSave = () => {
    // Save configuration
    setHasChanges(false);
    alert('Client portal configuration saved successfully!');
  };

  const handleReset = () => {
    if (confirm('Are you sure you want to reset all changes?')) {
      // Reset to defaults
      setHasChanges(false);
    }
  };

  return (
    <div className="flex flex-col h-full bg-[#1e1e1e] text-zinc-300 overflow-hidden">
      {/* Header */}
      <div className="px-4 py-3 border-b border-zinc-700 bg-[#252528] flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-zinc-200 flex items-center gap-2">
            <Globe size={20} />
            Client Portal Configuration
          </h2>
          <p className="text-xs text-zinc-500 mt-0.5">
            Configure self-service portal settings and features for your clients
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleReset}
            disabled={!hasChanges}
            className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 disabled:opacity-50 disabled:cursor-not-allowed text-zinc-300 rounded text-sm flex items-center gap-2 transition-colors"
          >
            <RotateCcw size={14} />
            Reset
          </button>
          <button
            onClick={handleSave}
            disabled={!hasChanges}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded text-sm flex items-center gap-2 transition-colors"
          >
            <Save size={14} />
            Save Changes
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="px-4 py-4 border-b border-zinc-700 bg-[#1e1e1e]">
        <div className="grid grid-cols-4 gap-4">
          <div className="bg-[#252528] border border-zinc-700 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-zinc-500 uppercase tracking-wide">Portal Visitors Today</span>
              <Users size={16} className="text-blue-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-200">{stats.portalVisitorsToday}</div>
            <div className="text-xs text-emerald-400 mt-1">+12% from yesterday</div>
          </div>

          <div className="bg-[#252528] border border-zinc-700 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-zinc-500 uppercase tracking-wide">Open Tickets</span>
              <Ticket size={16} className="text-yellow-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-200">{stats.openTickets}</div>
            <div className="text-xs text-zinc-500 mt-1">Awaiting response</div>
          </div>

          <div className="bg-[#252528] border border-zinc-700 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-zinc-500 uppercase tracking-wide">Pending KYC Uploads</span>
              <Upload size={16} className="text-orange-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-200">{stats.pendingKycUploads}</div>
            <div className="text-xs text-zinc-500 mt-1">Requires review</div>
          </div>

          <div className="bg-[#252528] border border-zinc-700 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-zinc-500 uppercase tracking-wide">Self-Service Actions</span>
              <CheckCircle size={16} className="text-emerald-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-200">{stats.selfServiceActionsToday}</div>
            <div className="text-xs text-emerald-400 mt-1">Automated today</div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-6">
        {/* Section 1: Portal Settings */}
        <div className="bg-[#252528] border border-zinc-700 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-200 mb-4 flex items-center gap-2">
            <Globe size={16} />
            Portal Settings
          </h3>

          <div className="space-y-4">
            {/* Enable/Disable Portal */}
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <div className="text-sm font-medium text-zinc-300">Enable Client Portal</div>
                <div className="text-xs text-zinc-500 mt-0.5">
                  Turn the client self-service portal on or off
                </div>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={portalSettings.enabled}
                  onChange={(e) => handlePortalSettingChange('enabled', e.target.checked)}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-zinc-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
              </label>
            </div>

            <div className="grid grid-cols-2 gap-4">
              {/* Logo URL */}
              <div className="space-y-1">
                <label className="block text-xs text-zinc-500 flex items-center gap-1">
                  <Image size={12} />
                  Logo URL
                </label>
                <input
                  type="url"
                  value={portalSettings.logoUrl}
                  onChange={(e) => handlePortalSettingChange('logoUrl', e.target.value)}
                  className="w-full px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500"
                  placeholder="https://example.com/logo.png"
                />
              </div>

              {/* Primary Color */}
              <div className="space-y-1">
                <label className="block text-xs text-zinc-500 flex items-center gap-1">
                  <Palette size={12} />
                  Primary Color
                </label>
                <div className="flex gap-2">
                  <input
                    type="color"
                    value={portalSettings.primaryColor}
                    onChange={(e) => handlePortalSettingChange('primaryColor', e.target.value)}
                    className="w-12 h-10 bg-[#1e1e1e] border border-zinc-700 rounded cursor-pointer"
                  />
                  <input
                    type="text"
                    value={portalSettings.primaryColor}
                    onChange={(e) => handlePortalSettingChange('primaryColor', e.target.value)}
                    className="flex-1 px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500"
                    placeholder="#3B82F6"
                  />
                </div>
              </div>

              {/* Company Name */}
              <div className="space-y-1">
                <label className="block text-xs text-zinc-500">Company Name</label>
                <input
                  type="text"
                  value={portalSettings.companyName}
                  onChange={(e) => handlePortalSettingChange('companyName', e.target.value)}
                  className="w-full px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500"
                  placeholder="RTX5 Trading"
                />
              </div>

              {/* Terms & Conditions URL */}
              <div className="space-y-1">
                <label className="block text-xs text-zinc-500">Terms & Conditions URL</label>
                <input
                  type="url"
                  value={portalSettings.termsUrl}
                  onChange={(e) => handlePortalSettingChange('termsUrl', e.target.value)}
                  className="w-full px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500"
                  placeholder="https://example.com/terms"
                />
              </div>
            </div>

            {/* Welcome Message */}
            <div className="space-y-1">
              <label className="block text-xs text-zinc-500">Welcome Message</label>
              <textarea
                value={portalSettings.welcomeMessage}
                onChange={(e) => handlePortalSettingChange('welcomeMessage', e.target.value)}
                rows={3}
                className="w-full px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500 resize-none"
                placeholder="Welcome message for clients..."
              />
            </div>
          </div>
        </div>

        {/* Section 2: Self-Service Features */}
        <div className="bg-[#252528] border border-zinc-700 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-200 mb-4 flex items-center gap-2">
            <CheckCircle size={16} />
            Self-Service Features
          </h3>

          <div className="grid grid-cols-2 gap-4">
            {selfServiceFeatures.map((feature) => (
              <div
                key={feature.id}
                className="bg-[#1e1e1e] border border-zinc-700 rounded p-3 flex items-start justify-between hover:border-zinc-600 transition-colors"
              >
                <div className="flex items-start gap-3 flex-1">
                  <div className="text-zinc-400 mt-0.5">{feature.icon}</div>
                  <div className="flex-1">
                    <div className="text-sm font-medium text-zinc-300">{feature.name}</div>
                    <div className="text-xs text-zinc-500 mt-0.5">{feature.description}</div>
                  </div>
                </div>
                <label className="relative inline-flex items-center cursor-pointer ml-3">
                  <input
                    type="checkbox"
                    checked={feature.enabled}
                    onChange={() => handleFeatureToggle(feature.id)}
                    className="sr-only peer"
                  />
                  <div className="w-9 h-5 bg-zinc-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>
            ))}
          </div>
        </div>

        {/* Section 3: Document Requirements */}
        <div className="bg-[#252528] border border-zinc-700 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-200 mb-4 flex items-center gap-2">
            <FileText size={16} />
            Document Requirements
          </h3>

          <div className="space-y-3">
            {documentRequirements.map((doc) => (
              <div
                key={doc.id}
                className="bg-[#1e1e1e] border border-zinc-700 rounded p-3"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <div className="text-sm font-medium text-zinc-300">{doc.name}</div>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer ml-3">
                    <span className="text-xs text-zinc-500 mr-2">Required</span>
                    <input
                      type="checkbox"
                      checked={doc.required}
                      onChange={(e) =>
                        handleDocumentRequirementChange(doc.id, 'required', e.target.checked)
                      }
                      className="sr-only peer"
                    />
                    <div className="w-9 h-5 bg-zinc-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-blue-600"></div>
                  </label>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1">
                    <label className="block text-xs text-zinc-500">Accepted Formats</label>
                    <input
                      type="text"
                      value={doc.acceptedFormats.join(', ')}
                      onChange={(e) =>
                        handleDocumentRequirementChange(
                          doc.id,
                          'acceptedFormats',
                          e.target.value.split(',').map((f) => f.trim())
                        )
                      }
                      className="w-full px-2 py-1.5 bg-[#252528] border border-zinc-700 rounded text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
                      placeholder="PDF, JPG, PNG"
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="block text-xs text-zinc-500">Max File Size (MB)</label>
                    <input
                      type="number"
                      value={doc.maxSizeMB}
                      onChange={(e) =>
                        handleDocumentRequirementChange(doc.id, 'maxSizeMB', Number(e.target.value))
                      }
                      min="1"
                      max="50"
                      className="w-full px-2 py-1.5 bg-[#252528] border border-zinc-700 rounded text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Section 4: Support Configuration */}
        <div className="bg-[#252528] border border-zinc-700 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-200 mb-4 flex items-center gap-2">
            <MessageCircle size={16} />
            Support Configuration
          </h3>

          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              {/* Support Email */}
              <div className="space-y-1">
                <label className="block text-xs text-zinc-500 flex items-center gap-1">
                  <Mail size={12} />
                  Support Email
                </label>
                <input
                  type="email"
                  value={supportConfig.email}
                  onChange={(e) => handleSupportConfigChange('email', e.target.value)}
                  className="w-full px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500"
                  placeholder="support@rtx5trading.com"
                />
              </div>

              {/* Support Phone */}
              <div className="space-y-1">
                <label className="block text-xs text-zinc-500 flex items-center gap-1">
                  <Phone size={12} />
                  Support Phone
                </label>
                <input
                  type="tel"
                  value={supportConfig.phone}
                  onChange={(e) => handleSupportConfigChange('phone', e.target.value)}
                  className="w-full px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500"
                  placeholder="+1-800-RTX-5000"
                />
              </div>
            </div>

            {/* Live Chat Toggle */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <MessageSquare size={16} className="text-zinc-400" />
                <div>
                  <div className="text-sm font-medium text-zinc-300">Enable Live Chat</div>
                  <div className="text-xs text-zinc-500">Real-time chat support for clients</div>
                </div>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={supportConfig.liveChatEnabled}
                  onChange={(e) => handleSupportConfigChange('liveChatEnabled', e.target.checked)}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-zinc-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
              </label>
            </div>

            {/* Working Hours */}
            <div className="space-y-1">
              <label className="block text-xs text-zinc-500 flex items-center gap-1">
                <Clock size={12} />
                Working Hours
              </label>
              <input
                type="text"
                value={supportConfig.workingHours}
                onChange={(e) => handleSupportConfigChange('workingHours', e.target.value)}
                className="w-full px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500"
                placeholder="Monday - Friday, 9:00 AM - 6:00 PM EST"
              />
            </div>

            {/* Auto-Reply Message */}
            <div className="space-y-1">
              <label className="block text-xs text-zinc-500">Auto-Reply Message</label>
              <textarea
                value={supportConfig.autoReplyMessage}
                onChange={(e) => handleSupportConfigChange('autoReplyMessage', e.target.value)}
                rows={3}
                className="w-full px-3 py-2 bg-[#1e1e1e] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500 resize-none"
                placeholder="Automated response for support ticket submissions..."
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
