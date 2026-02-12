'use client';

import React, { useState, useMemo } from 'react';
import {
  Palette,
  Image,
  Mail,
  Monitor,
  Globe,
  ChevronDown,
  ChevronRight,
  Save,
  RotateCcw,
  Check,
  X,
  Info,
  Eye
} from 'lucide-react';

type ButtonStyle = 'rounded' | 'square' | 'pill';
type ThemeMode = 'dark' | 'light';

interface BrandConfig {
  // Brand Identity
  companyName: string;
  logoUrl: string;
  faviconUrl: string;
  tagline: string;
  primaryColor: string;
  secondaryColor: string;
  accentColor: string;

  // Login Page
  loginBgUrl: string;
  welcomeMessage: string;
  showRegistration: boolean;
  termsUrl: string;
  privacyUrl: string;

  // Email Branding
  emailLogoUrl: string;
  emailFooter: string;
  replyToEmail: string;
  fromName: string;
  emailSignature: string;

  // Platform Theme
  defaultMode: ThemeMode;
  sidebarColor: string;
  headerColor: string;
  buttonStyle: ButtonStyle;
  fontFamily: string;

  // Domain Config
  customDomain: string;
  sslActive: boolean;
  dnsVerified: boolean;
}

interface BrandPreset {
  name: string;
  primaryColor: string;
  secondaryColor: string;
  accentColor: string;
  sidebarColor: string;
  headerColor: string;
}

export default function WhiteLabelConfig() {
  const [expandedSections, setExpandedSections] = useState<Set<string>>(
    new Set(['brand-identity'])
  );

  const [config, setConfig] = useState<BrandConfig>({
    companyName: 'RTX5 Trading',
    logoUrl: 'https://via.placeholder.com/150x40/3b82f6/ffffff?text=RTX5',
    faviconUrl: 'https://via.placeholder.com/32x32/3b82f6/ffffff?text=R',
    tagline: 'Advanced Trading Platform',
    primaryColor: '#3b82f6',
    secondaryColor: '#6366f1',
    accentColor: '#22c55e',
    loginBgUrl: 'https://via.placeholder.com/1920x1080/1e293b/ffffff?text=Login+Background',
    welcomeMessage: 'Welcome to our professional trading platform',
    showRegistration: true,
    termsUrl: 'https://example.com/terms',
    privacyUrl: 'https://example.com/privacy',
    emailLogoUrl: 'https://via.placeholder.com/200x60/3b82f6/ffffff?text=Email+Logo',
    emailFooter: '© 2026 RTX5 Trading. All rights reserved.',
    replyToEmail: 'support@rtx5trading.com',
    fromName: 'RTX5 Trading Support',
    emailSignature: '<p>Best regards,<br/>The RTX5 Team</p>',
    defaultMode: 'dark',
    sidebarColor: '#18181b',
    headerColor: '#27272a',
    buttonStyle: 'rounded',
    fontFamily: 'Inter',
    customDomain: 'trading.yourdomain.com',
    sslActive: true,
    dnsVerified: false
  });

  const [savedConfig, setSavedConfig] = useState<BrandConfig>(config);

  const brandPresets: BrandPreset[] = useMemo(() => [
    {
      name: 'Ocean Blue',
      primaryColor: '#3b82f6',
      secondaryColor: '#0ea5e9',
      accentColor: '#06b6d4',
      sidebarColor: '#0c4a6e',
      headerColor: '#075985'
    },
    {
      name: 'Forest Green',
      primaryColor: '#22c55e',
      secondaryColor: '#16a34a',
      accentColor: '#84cc16',
      sidebarColor: '#14532d',
      headerColor: '#166534'
    },
    {
      name: 'Royal Purple',
      primaryColor: '#8b5cf6',
      secondaryColor: '#7c3aed',
      accentColor: '#a855f7',
      sidebarColor: '#4c1d95',
      headerColor: '#5b21b6'
    },
    {
      name: 'Sunset Orange',
      primaryColor: '#f97316',
      secondaryColor: '#ea580c',
      accentColor: '#fb923c',
      sidebarColor: '#7c2d12',
      headerColor: '#9a3412'
    },
    {
      name: 'Midnight Dark',
      primaryColor: '#1e293b',
      secondaryColor: '#334155',
      accentColor: '#64748b',
      sidebarColor: '#0f172a',
      headerColor: '#1e293b'
    },
    {
      name: 'Clean White',
      primaryColor: '#ffffff',
      secondaryColor: '#f1f5f9',
      accentColor: '#3b82f6',
      sidebarColor: '#f8fafc',
      headerColor: '#ffffff'
    }
  ], []);

  const toggleSection = (sectionId: string) => {
    setExpandedSections(prev => {
      const newSet = new Set(prev);
      if (newSet.has(sectionId)) {
        newSet.delete(sectionId);
      } else {
        newSet.add(sectionId);
      }
      return newSet;
    });
  };

  const handleSave = () => {
    setSavedConfig(config);
    console.log('Saving white label config:', config);
  };

  const handleReset = () => {
    setConfig(savedConfig);
  };

  const applyPreset = (preset: BrandPreset) => {
    setConfig(prev => ({
      ...prev,
      primaryColor: preset.primaryColor,
      secondaryColor: preset.secondaryColor,
      accentColor: preset.accentColor,
      sidebarColor: preset.sidebarColor,
      headerColor: preset.headerColor
    }));
  };

  const hasChanges = JSON.stringify(config) !== JSON.stringify(savedConfig);

  const renderSection = (
    id: string,
    title: string,
    icon: React.ReactElement,
    content: React.ReactElement
  ) => {
    const isExpanded = expandedSections.has(id);
    return (
      <div className="bg-[#27272a] rounded-lg border border-zinc-700 overflow-hidden">
        <button
          onClick={() => toggleSection(id)}
          className="w-full px-6 py-4 flex items-center justify-between hover:bg-zinc-800/50 transition-colors"
        >
          <div className="flex items-center gap-3">
            <div className="text-blue-400">{icon}</div>
            <h3 className="text-sm font-semibold text-zinc-100">{title}</h3>
          </div>
          {isExpanded ? (
            <ChevronDown size={20} className="text-zinc-400" />
          ) : (
            <ChevronRight size={20} className="text-zinc-400" />
          )}
        </button>
        {isExpanded && <div className="px-6 py-4 border-t border-zinc-700">{content}</div>}
      </div>
    );
  };

  const ColorPicker = ({
    label,
    value,
    onChange
  }: {
    label: string;
    value: string;
    onChange: (color: string) => void;
  }) => (
    <div>
      <label className="block text-sm font-medium text-zinc-300 mb-2">{label}</label>
      <div className="flex gap-3">
        <input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="flex-1 px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 font-mono focus:outline-none focus:border-zinc-600"
          placeholder="#3b82f6"
        />
        <div
          className="w-12 h-10 rounded border border-zinc-700 cursor-pointer"
          style={{ backgroundColor: value }}
          title={value}
        />
      </div>
    </div>
  );

  return (
    <div className="h-full flex bg-[#18181b] overflow-hidden">
      {/* Main Configuration Panel */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex-shrink-0 px-6 py-4 border-b border-zinc-700 flex items-center justify-between">
          <div>
            <h2 className="text-xl font-bold text-zinc-100">White Label Configuration</h2>
            <p className="text-sm text-zinc-400 mt-1">Customize platform branding and appearance</p>
          </div>
          <div className="flex gap-3">
            <button
              onClick={handleReset}
              disabled={!hasChanges}
              className={`flex items-center gap-2 px-4 py-2 rounded text-sm font-medium transition-colors ${
                hasChanges
                  ? 'bg-zinc-700 hover:bg-zinc-600 text-zinc-100'
                  : 'bg-zinc-800 text-zinc-600 cursor-not-allowed'
              }`}
            >
              <RotateCcw size={16} />
              Reset
            </button>
            <button
              onClick={handleSave}
              disabled={!hasChanges}
              className={`flex items-center gap-2 px-4 py-2 rounded text-sm font-medium transition-colors ${
                hasChanges
                  ? 'bg-blue-600 hover:bg-blue-700 text-white'
                  : 'bg-blue-900/30 text-blue-700 cursor-not-allowed'
              }`}
            >
              <Save size={16} />
              Save Changes
            </button>
          </div>
        </div>

        {/* Brand Presets */}
        <div className="flex-shrink-0 px-6 py-4 border-b border-zinc-700">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3 uppercase tracking-wide">
            Brand Presets
          </h3>
          <div className="flex gap-2 overflow-x-auto pb-2">
            {brandPresets.map((preset) => (
              <button
                key={preset.name}
                onClick={() => applyPreset(preset)}
                className="flex-shrink-0 px-4 py-2 bg-[#27272a] border border-zinc-700 rounded hover:border-zinc-600 transition-colors group"
              >
                <div className="flex items-center gap-2">
                  <div className="flex gap-1">
                    <div
                      className="w-4 h-4 rounded-full border border-zinc-600"
                      style={{ backgroundColor: preset.primaryColor }}
                    />
                    <div
                      className="w-4 h-4 rounded-full border border-zinc-600"
                      style={{ backgroundColor: preset.secondaryColor }}
                    />
                    <div
                      className="w-4 h-4 rounded-full border border-zinc-600"
                      style={{ backgroundColor: preset.accentColor }}
                    />
                  </div>
                  <span className="text-xs font-medium text-zinc-300 group-hover:text-zinc-100">
                    {preset.name}
                  </span>
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Configuration Sections */}
        <div className="flex-1 overflow-auto px-6 py-4 space-y-4">
          {/* Section 1: Brand Identity */}
          {renderSection(
            'brand-identity',
            'Brand Identity',
            <Palette size={20} />,
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Company Name
                </label>
                <input
                  type="text"
                  value={config.companyName}
                  onChange={(e) => setConfig({ ...config, companyName: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">Tagline</label>
                <input
                  type="text"
                  value={config.tagline}
                  onChange={(e) => setConfig({ ...config, tagline: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">Logo URL</label>
                <input
                  type="text"
                  value={config.logoUrl}
                  onChange={(e) => setConfig({ ...config, logoUrl: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
                <img
                  src={config.logoUrl}
                  alt="Logo preview"
                  className="mt-2 h-10 border border-zinc-700 rounded bg-zinc-800"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Favicon URL
                </label>
                <input
                  type="text"
                  value={config.faviconUrl}
                  onChange={(e) => setConfig({ ...config, faviconUrl: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
                <img
                  src={config.faviconUrl}
                  alt="Favicon preview"
                  className="mt-2 w-8 h-8 border border-zinc-700 rounded bg-zinc-800"
                />
              </div>

              <ColorPicker
                label="Primary Color"
                value={config.primaryColor}
                onChange={(color) => setConfig({ ...config, primaryColor: color })}
              />

              <ColorPicker
                label="Secondary Color"
                value={config.secondaryColor}
                onChange={(color) => setConfig({ ...config, secondaryColor: color })}
              />

              <ColorPicker
                label="Accent Color"
                value={config.accentColor}
                onChange={(color) => setConfig({ ...config, accentColor: color })}
              />
            </div>
          )}

          {/* Section 2: Login Page */}
          {renderSection(
            'login-page',
            'Login Page',
            <Image size={20} />,
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-2">
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Background Image URL
                </label>
                <input
                  type="text"
                  value={config.loginBgUrl}
                  onChange={(e) => setConfig({ ...config, loginBgUrl: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div className="col-span-2">
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Welcome Message
                </label>
                <textarea
                  value={config.welcomeMessage}
                  onChange={(e) => setConfig({ ...config, welcomeMessage: e.target.value })}
                  rows={3}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={config.showRegistration}
                  onChange={(e) => setConfig({ ...config, showRegistration: e.target.checked })}
                  className="w-4 h-4 rounded border-zinc-700 bg-zinc-800"
                />
                <label className="text-sm font-medium text-zinc-300">
                  Show Registration Option
                </label>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Terms & Conditions URL
                </label>
                <input
                  type="text"
                  value={config.termsUrl}
                  onChange={(e) => setConfig({ ...config, termsUrl: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Privacy Policy URL
                </label>
                <input
                  type="text"
                  value={config.privacyUrl}
                  onChange={(e) => setConfig({ ...config, privacyUrl: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>
            </div>
          )}

          {/* Section 3: Email Branding */}
          {renderSection(
            'email-branding',
            'Email Branding',
            <Mail size={20} />,
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-2">
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Header Logo URL
                </label>
                <input
                  type="text"
                  value={config.emailLogoUrl}
                  onChange={(e) => setConfig({ ...config, emailLogoUrl: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Reply-To Email
                </label>
                <input
                  type="email"
                  value={config.replyToEmail}
                  onChange={(e) => setConfig({ ...config, replyToEmail: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  From Name
                </label>
                <input
                  type="text"
                  value={config.fromName}
                  onChange={(e) => setConfig({ ...config, fromName: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div className="col-span-2">
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Footer Text
                </label>
                <textarea
                  value={config.emailFooter}
                  onChange={(e) => setConfig({ ...config, emailFooter: e.target.value })}
                  rows={2}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div className="col-span-2">
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Signature HTML
                </label>
                <textarea
                  value={config.emailSignature}
                  onChange={(e) => setConfig({ ...config, emailSignature: e.target.value })}
                  rows={3}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 font-mono focus:outline-none focus:border-zinc-600"
                />
              </div>
            </div>
          )}

          {/* Section 4: Platform Theme */}
          {renderSection(
            'platform-theme',
            'Platform Theme',
            <Monitor size={20} />,
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Default Mode
                </label>
                <div className="flex gap-3">
                  {(['dark', 'light'] as ThemeMode[]).map((mode) => (
                    <button
                      key={mode}
                      onClick={() => setConfig({ ...config, defaultMode: mode })}
                      className={`flex-1 px-4 py-2 rounded text-sm font-medium transition-colors ${
                        config.defaultMode === mode
                          ? 'bg-blue-600 text-white'
                          : 'bg-[#18181b] text-zinc-400 hover:bg-zinc-800 border border-zinc-700'
                      }`}
                    >
                      {mode.charAt(0).toUpperCase() + mode.slice(1)}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Font Family
                </label>
                <select
                  value={config.fontFamily}
                  onChange={(e) => setConfig({ ...config, fontFamily: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                >
                  <option value="Inter">Inter</option>
                  <option value="Roboto">Roboto</option>
                  <option value="Open Sans">Open Sans</option>
                  <option value="Lato">Lato</option>
                  <option value="Poppins">Poppins</option>
                </select>
              </div>

              <ColorPicker
                label="Sidebar Color"
                value={config.sidebarColor}
                onChange={(color) => setConfig({ ...config, sidebarColor: color })}
              />

              <ColorPicker
                label="Header Color"
                value={config.headerColor}
                onChange={(color) => setConfig({ ...config, headerColor: color })}
              />

              <div className="col-span-2">
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Button Style
                </label>
                <div className="flex gap-3">
                  {(['rounded', 'square', 'pill'] as ButtonStyle[]).map((style) => (
                    <button
                      key={style}
                      onClick={() => setConfig({ ...config, buttonStyle: style })}
                      className={`flex-1 px-4 py-2 text-sm font-medium transition-colors ${
                        config.buttonStyle === style
                          ? 'bg-blue-600 text-white'
                          : 'bg-[#18181b] text-zinc-400 hover:bg-zinc-800 border border-zinc-700'
                      } ${
                        style === 'rounded'
                          ? 'rounded'
                          : style === 'square'
                          ? 'rounded-none'
                          : 'rounded-full'
                      }`}
                    >
                      {style.charAt(0).toUpperCase() + style.slice(1)}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Section 5: Domain Config */}
          {renderSection(
            'domain-config',
            'Domain Configuration',
            <Globe size={20} />,
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Custom Domain
                </label>
                <input
                  type="text"
                  value={config.customDomain}
                  onChange={(e) => setConfig({ ...config, customDomain: e.target.value })}
                  className="w-full px-3 py-2 bg-[#18181b] border border-zinc-700 rounded text-sm text-zinc-100 focus:outline-none focus:border-zinc-600"
                  placeholder="trading.yourdomain.com"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="bg-[#18181b] border border-zinc-700 rounded p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium text-zinc-300">SSL Status</span>
                    <div className="flex items-center gap-2">
                      {config.sslActive ? (
                        <>
                          <Check size={16} className="text-green-400" />
                          <span className="text-xs font-medium text-green-400">Active</span>
                        </>
                      ) : (
                        <>
                          <X size={16} className="text-red-400" />
                          <span className="text-xs font-medium text-red-400">Inactive</span>
                        </>
                      )}
                    </div>
                  </div>
                  <p className="text-xs text-zinc-500">
                    {config.sslActive
                      ? 'SSL certificate is valid and active'
                      : 'SSL certificate needs to be configured'}
                  </p>
                </div>

                <div className="bg-[#18181b] border border-zinc-700 rounded p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium text-zinc-300">DNS Verification</span>
                    <div className="flex items-center gap-2">
                      {config.dnsVerified ? (
                        <>
                          <Check size={16} className="text-green-400" />
                          <span className="text-xs font-medium text-green-400">Verified</span>
                        </>
                      ) : (
                        <>
                          <X size={16} className="text-yellow-400" />
                          <span className="text-xs font-medium text-yellow-400">Pending</span>
                        </>
                      )}
                    </div>
                  </div>
                  <p className="text-xs text-zinc-500">
                    {config.dnsVerified
                      ? 'Domain DNS records are properly configured'
                      : 'Waiting for DNS propagation'}
                  </p>
                </div>
              </div>

              <div className="bg-blue-500/10 border border-blue-500/30 rounded p-4">
                <div className="flex items-start gap-3">
                  <Info size={20} className="text-blue-400 flex-shrink-0 mt-0.5" />
                  <div>
                    <h4 className="text-sm font-semibold text-blue-300 mb-2">
                      DNS Setup Instructions
                    </h4>
                    <div className="text-xs text-zinc-400 space-y-1">
                      <p>1. Add a CNAME record pointing to: platform.rtx5.com</p>
                      <p>2. Add TXT record for domain verification</p>
                      <p>3. Wait 24-48 hours for DNS propagation</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Live Preview Panel */}
      <div className="w-80 border-l border-zinc-700 bg-[#27272a] p-6 overflow-auto">
        <div className="flex items-center gap-2 mb-4">
          <Eye size={18} className="text-zinc-400" />
          <h3 className="text-sm font-semibold text-zinc-300 uppercase tracking-wide">
            Live Preview
          </h3>
        </div>

        {/* Mock UI Preview */}
        <div className="bg-[#18181b] border border-zinc-700 rounded-lg overflow-hidden">
          {/* Mock Header */}
          <div
            className="h-12 flex items-center justify-between px-4 border-b border-zinc-700"
            style={{ backgroundColor: config.headerColor }}
          >
            <img src={config.logoUrl} alt="Logo" className="h-6" />
            <div className="flex gap-2">
              <div className="w-6 h-6 rounded-full bg-zinc-700" />
              <div className="w-6 h-6 rounded-full bg-zinc-700" />
            </div>
          </div>

          {/* Mock Content with Sidebar */}
          <div className="flex h-64">
            {/* Sidebar */}
            <div
              className="w-16 border-r border-zinc-700 p-2 space-y-2"
              style={{ backgroundColor: config.sidebarColor }}
            >
              {[1, 2, 3, 4, 5].map((i) => (
                <div
                  key={i}
                  className="w-full h-10 rounded flex items-center justify-center"
                  style={{
                    backgroundColor: i === 1 ? config.primaryColor : 'transparent'
                  }}
                >
                  <div className="w-5 h-5 bg-zinc-600 rounded" />
                </div>
              ))}
            </div>

            {/* Main Content */}
            <div className="flex-1 p-4">
              <div className="space-y-3">
                <div className="h-3 bg-zinc-700 rounded w-3/4" />
                <div className="h-3 bg-zinc-700 rounded w-1/2" />
                <div className="flex gap-2 mt-4">
                  <button
                    className={`px-3 py-1.5 text-xs font-medium text-white transition-colors ${
                      config.buttonStyle === 'rounded'
                        ? 'rounded'
                        : config.buttonStyle === 'square'
                        ? 'rounded-none'
                        : 'rounded-full'
                    }`}
                    style={{ backgroundColor: config.primaryColor }}
                  >
                    Primary
                  </button>
                  <button
                    className={`px-3 py-1.5 text-xs font-medium text-white transition-colors ${
                      config.buttonStyle === 'rounded'
                        ? 'rounded'
                        : config.buttonStyle === 'square'
                        ? 'rounded-none'
                        : 'rounded-full'
                    }`}
                    style={{ backgroundColor: config.secondaryColor }}
                  >
                    Secondary
                  </button>
                  <button
                    className={`px-3 py-1.5 text-xs font-medium text-white transition-colors ${
                      config.buttonStyle === 'rounded'
                        ? 'rounded'
                        : config.buttonStyle === 'square'
                        ? 'rounded-none'
                        : 'rounded-full'
                    }`}
                    style={{ backgroundColor: config.accentColor }}
                  >
                    Accent
                  </button>
                </div>
                <div className="h-20 bg-zinc-700 rounded mt-4" />
              </div>
            </div>
          </div>

          {/* Login Page Preview */}
          <div className="border-t border-zinc-700 p-4 bg-zinc-800/50">
            <p className="text-xs font-semibold text-zinc-400 mb-2">Login Page Preview:</p>
            <div className="bg-zinc-900 rounded p-3">
              <p className="text-xs text-zinc-300 mb-2">{config.welcomeMessage}</p>
              <div className="space-y-2">
                <div className="h-6 bg-zinc-700 rounded" />
                <div className="h-6 bg-zinc-700 rounded" />
                <button
                  className={`w-full h-6 text-xs font-medium text-white ${
                    config.buttonStyle === 'rounded'
                      ? 'rounded'
                      : config.buttonStyle === 'square'
                      ? 'rounded-none'
                      : 'rounded-full'
                  }`}
                  style={{ backgroundColor: config.primaryColor }}
                >
                  Login
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Color Palette Summary */}
        <div className="mt-6">
          <h4 className="text-xs font-semibold text-zinc-400 mb-3 uppercase tracking-wide">
            Color Palette
          </h4>
          <div className="space-y-2">
            {[
              { label: 'Primary', color: config.primaryColor },
              { label: 'Secondary', color: config.secondaryColor },
              { label: 'Accent', color: config.accentColor },
              { label: 'Sidebar', color: config.sidebarColor },
              { label: 'Header', color: config.headerColor }
            ].map((item) => (
              <div key={item.label} className="flex items-center gap-3">
                <div
                  className="w-8 h-8 rounded border border-zinc-700"
                  style={{ backgroundColor: item.color }}
                />
                <div className="flex-1">
                  <p className="text-xs font-medium text-zinc-300">{item.label}</p>
                  <p className="text-xs font-mono text-zinc-500">{item.color}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
