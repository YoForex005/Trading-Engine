'use client';

import { useState, useMemo } from 'react';
import { Settings, ChevronDown, ChevronRight, Save, X, Edit2, RotateCcw, Globe, Shield, Bell, Wrench, Database, Activity, Clock, User } from 'lucide-react';
import { API_ENDPOINTS } from '@/config/api';

// =====================
// TYPES
// =====================

interface SettingsHistory {
  id: string;
  section: string;
  field: string;
  oldValue: string;
  newValue: string;
  changedBy: string;
  changedAt: string;
}

interface GeneralSettings {
  platformName: string;
  companyName: string;
  timezone: string;
  defaultLanguage: string;
  dateFormat: string;
  numberFormat: string;
}

interface TradingSettings {
  defaultLeverage: number;
  maxLeverage: number;
  marginCallLevel: number;
  stopOutLevel: number;
  maxOpenPositions: number;
  maxPendingOrders: number;
  allowedTradeModes: string[];
}

interface SecuritySettings {
  passwordMinLength: number;
  requireUppercase: boolean;
  requireNumbers: boolean;
  requireSpecialChars: boolean;
  maxLoginAttempts: number;
  lockoutDuration: number;
  sessionTimeout: number;
  require2FA: boolean;
  ipRestrictionMode: 'whitelist' | 'blacklist' | 'none';
}

interface NotificationSettings {
  emailEnabled: boolean;
  smsEnabled: boolean;
  pushEnabled: boolean;
  smtpHost: string;
  smtpPort: number;
  smtpUser: string;
  smtpFrom: string;
}

interface MaintenanceSettings {
  maintenanceMode: boolean;
  maintenanceMessage: string;
  scheduledDowntime: string;
  dataRetentionDays: number;
}

interface ApiSettings {
  rateLimit: number;
  apiVersion: string;
  wsMaxConnections: number;
  maxPayloadSize: number;
}

type SectionKey = 'general' | 'trading' | 'security' | 'notifications' | 'maintenance' | 'api';

// =====================
// MAIN COMPONENT
// =====================

export default function PlatformConfig() {
  // Expanded sections state
  const [expandedSections, setExpandedSections] = useState<Set<SectionKey>>(new Set(['general']));

  // Editing states
  const [editingSection, setEditingSection] = useState<SectionKey | null>(null);

  // Settings states
  const [generalSettings, setGeneralSettings] = useState<GeneralSettings>({
    platformName: 'RTX5 Trading Platform',
    companyName: 'RTX5 Financial Services',
    timezone: 'UTC+00:00',
    defaultLanguage: 'English (US)',
    dateFormat: 'YYYY-MM-DD',
    numberFormat: '1,000.00',
  });

  const [tradingSettings, setTradingSettings] = useState<TradingSettings>({
    defaultLeverage: 100,
    maxLeverage: 500,
    marginCallLevel: 50,
    stopOutLevel: 20,
    maxOpenPositions: 200,
    maxPendingOrders: 100,
    allowedTradeModes: ['market', 'limit', 'stop', 'stop-limit'],
  });

  const [securitySettings, setSecuritySettings] = useState<SecuritySettings>({
    passwordMinLength: 8,
    requireUppercase: true,
    requireNumbers: true,
    requireSpecialChars: true,
    maxLoginAttempts: 5,
    lockoutDuration: 30,
    sessionTimeout: 30,
    require2FA: false,
    ipRestrictionMode: 'none',
  });

  const [notificationSettings, setNotificationSettings] = useState<NotificationSettings>({
    emailEnabled: true,
    smsEnabled: false,
    pushEnabled: true,
    smtpHost: 'smtp.gmail.com',
    smtpPort: 587,
    smtpUser: 'noreply@rtx5.com',
    smtpFrom: 'RTX5 Platform <noreply@rtx5.com>',
  });

  const [maintenanceSettings, setMaintenanceSettings] = useState<MaintenanceSettings>({
    maintenanceMode: false,
    maintenanceMessage: 'System is under maintenance. Please try again later.',
    scheduledDowntime: 'None scheduled',
    dataRetentionDays: 730,
  });

  const [apiSettings, setApiSettings] = useState<ApiSettings>({
    rateLimit: 1000,
    apiVersion: 'v1.2.0',
    wsMaxConnections: 10000,
    maxPayloadSize: 5,
  });

  // Settings history (mock data)
  const [settingsHistory] = useState<SettingsHistory[]>([
    { id: '1', section: 'Trading', field: 'Max Leverage', oldValue: '400', newValue: '500', changedBy: 'admin@rtx5.com', changedAt: '2026-02-11 08:30:15' },
    { id: '2', section: 'Security', field: 'Password Min Length', oldValue: '6', newValue: '8', changedBy: 'admin@rtx5.com', changedAt: '2026-02-10 14:22:03' },
    { id: '3', section: 'Notifications', field: 'SMS Enabled', oldValue: 'true', newValue: 'false', changedBy: 'support@rtx5.com', changedAt: '2026-02-09 11:15:45' },
    { id: '4', section: 'API', field: 'Rate Limit', oldValue: '500', newValue: '1000', changedBy: 'admin@rtx5.com', changedAt: '2026-02-08 16:50:30' },
    { id: '5', section: 'General', field: 'Default Language', oldValue: 'English (UK)', newValue: 'English (US)', changedBy: 'admin@rtx5.com', changedAt: '2026-02-07 09:05:12' },
  ]);

  // Toggle section expansion
  const toggleSection = (section: SectionKey) => {
    setExpandedSections(prev => {
      const newSet = new Set(prev);
      if (newSet.has(section)) {
        newSet.delete(section);
      } else {
        newSet.add(section);
      }
      return newSet;
    });
  };

  // Edit handlers
  const handleEdit = (section: SectionKey) => {
    setEditingSection(section);
  };

  const handleSave = (section: SectionKey) => {
    // In production: POST to API_ENDPOINTS.PLATFORM_CONFIG_UPDATE
    console.log('Saving settings for section:', section);
    setEditingSection(null);
  };

  const handleCancel = () => {
    setEditingSection(null);
    // Reset changes (in production: reload from API)
  };

  const handleReset = (section: SectionKey) => {
    // In production: POST to API_ENDPOINTS.PLATFORM_CONFIG_RESET
    console.log('Resetting section to defaults:', section);
  };

  // =====================
  // RENDER SECTIONS
  // =====================

  const renderGeneralSection = () => {
    const isEditing = editingSection === 'general';
    return (
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-xs text-[#888] mb-1">Platform Name</label>
          {isEditing ? (
            <input
              type="text"
              value={generalSettings.platformName}
              onChange={(e) => setGeneralSettings({ ...generalSettings, platformName: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{generalSettings.platformName}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Company Name</label>
          {isEditing ? (
            <input
              type="text"
              value={generalSettings.companyName}
              onChange={(e) => setGeneralSettings({ ...generalSettings, companyName: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{generalSettings.companyName}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Timezone</label>
          {isEditing ? (
            <select
              value={generalSettings.timezone}
              onChange={(e) => setGeneralSettings({ ...generalSettings, timezone: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            >
              <option>UTC+00:00</option>
              <option>UTC-05:00 (EST)</option>
              <option>UTC+01:00 (CET)</option>
              <option>UTC+08:00 (SGT)</option>
            </select>
          ) : (
            <div className="text-sm text-[#E0E0E0]">{generalSettings.timezone}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Default Language</label>
          {isEditing ? (
            <select
              value={generalSettings.defaultLanguage}
              onChange={(e) => setGeneralSettings({ ...generalSettings, defaultLanguage: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            >
              <option>English (US)</option>
              <option>English (UK)</option>
              <option>Spanish</option>
              <option>French</option>
              <option>German</option>
              <option>Chinese</option>
            </select>
          ) : (
            <div className="text-sm text-[#E0E0E0]">{generalSettings.defaultLanguage}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Date Format</label>
          {isEditing ? (
            <select
              value={generalSettings.dateFormat}
              onChange={(e) => setGeneralSettings({ ...generalSettings, dateFormat: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            >
              <option>YYYY-MM-DD</option>
              <option>DD/MM/YYYY</option>
              <option>MM/DD/YYYY</option>
            </select>
          ) : (
            <div className="text-sm text-[#E0E0E0]">{generalSettings.dateFormat}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Number Format</label>
          {isEditing ? (
            <select
              value={generalSettings.numberFormat}
              onChange={(e) => setGeneralSettings({ ...generalSettings, numberFormat: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            >
              <option>1,000.00</option>
              <option>1.000,00</option>
              <option>1 000.00</option>
            </select>
          ) : (
            <div className="text-sm text-[#E0E0E0]">{generalSettings.numberFormat}</div>
          )}
        </div>
      </div>
    );
  };

  const renderTradingSection = () => {
    const isEditing = editingSection === 'trading';
    return (
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-xs text-[#888] mb-1">Default Leverage</label>
          {isEditing ? (
            <input
              type="number"
              value={tradingSettings.defaultLeverage}
              onChange={(e) => setTradingSettings({ ...tradingSettings, defaultLeverage: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">1:{tradingSettings.defaultLeverage}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Max Leverage</label>
          {isEditing ? (
            <input
              type="number"
              value={tradingSettings.maxLeverage}
              onChange={(e) => setTradingSettings({ ...tradingSettings, maxLeverage: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">1:{tradingSettings.maxLeverage}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Margin Call Level</label>
          {isEditing ? (
            <input
              type="number"
              value={tradingSettings.marginCallLevel}
              onChange={(e) => setTradingSettings({ ...tradingSettings, marginCallLevel: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{tradingSettings.marginCallLevel}%</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Stop Out Level</label>
          {isEditing ? (
            <input
              type="number"
              value={tradingSettings.stopOutLevel}
              onChange={(e) => setTradingSettings({ ...tradingSettings, stopOutLevel: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{tradingSettings.stopOutLevel}%</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Max Open Positions</label>
          {isEditing ? (
            <input
              type="number"
              value={tradingSettings.maxOpenPositions}
              onChange={(e) => setTradingSettings({ ...tradingSettings, maxOpenPositions: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{tradingSettings.maxOpenPositions}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Max Pending Orders</label>
          {isEditing ? (
            <input
              type="number"
              value={tradingSettings.maxPendingOrders}
              onChange={(e) => setTradingSettings({ ...tradingSettings, maxPendingOrders: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{tradingSettings.maxPendingOrders}</div>
          )}
        </div>
        <div className="col-span-2">
          <label className="block text-xs text-[#888] mb-1">Allowed Trade Modes</label>
          {isEditing ? (
            <div className="flex gap-4 mt-2">
              {['market', 'limit', 'stop', 'stop-limit'].map((mode) => (
                <label key={mode} className="flex items-center gap-2 text-xs text-[#E0E0E0]">
                  <input
                    type="checkbox"
                    checked={tradingSettings.allowedTradeModes.includes(mode)}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setTradingSettings({ ...tradingSettings, allowedTradeModes: [...tradingSettings.allowedTradeModes, mode] });
                      } else {
                        setTradingSettings({ ...tradingSettings, allowedTradeModes: tradingSettings.allowedTradeModes.filter(m => m !== mode) });
                      }
                    }}
                    className="rounded"
                  />
                  {mode}
                </label>
              ))}
            </div>
          ) : (
            <div className="text-sm text-[#E0E0E0]">{tradingSettings.allowedTradeModes.join(', ')}</div>
          )}
        </div>
      </div>
    );
  };

  const renderSecuritySection = () => {
    const isEditing = editingSection === 'security';
    return (
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-xs text-[#888] mb-1">Password Min Length</label>
          {isEditing ? (
            <input
              type="number"
              value={securitySettings.passwordMinLength}
              onChange={(e) => setSecuritySettings({ ...securitySettings, passwordMinLength: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{securitySettings.passwordMinLength} characters</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Max Login Attempts</label>
          {isEditing ? (
            <input
              type="number"
              value={securitySettings.maxLoginAttempts}
              onChange={(e) => setSecuritySettings({ ...securitySettings, maxLoginAttempts: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{securitySettings.maxLoginAttempts} attempts</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Lockout Duration (minutes)</label>
          {isEditing ? (
            <input
              type="number"
              value={securitySettings.lockoutDuration}
              onChange={(e) => setSecuritySettings({ ...securitySettings, lockoutDuration: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{securitySettings.lockoutDuration} min</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Session Timeout (minutes)</label>
          {isEditing ? (
            <input
              type="number"
              value={securitySettings.sessionTimeout}
              onChange={(e) => setSecuritySettings({ ...securitySettings, sessionTimeout: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{securitySettings.sessionTimeout} min</div>
          )}
        </div>
        <div className="col-span-2">
          <label className="block text-xs text-[#888] mb-2">Password Requirements</label>
          {isEditing ? (
            <div className="flex gap-4">
              <label className="flex items-center gap-2 text-xs text-[#E0E0E0]">
                <input
                  type="checkbox"
                  checked={securitySettings.requireUppercase}
                  onChange={(e) => setSecuritySettings({ ...securitySettings, requireUppercase: e.target.checked })}
                  className="rounded"
                />
                Require Uppercase
              </label>
              <label className="flex items-center gap-2 text-xs text-[#E0E0E0]">
                <input
                  type="checkbox"
                  checked={securitySettings.requireNumbers}
                  onChange={(e) => setSecuritySettings({ ...securitySettings, requireNumbers: e.target.checked })}
                  className="rounded"
                />
                Require Numbers
              </label>
              <label className="flex items-center gap-2 text-xs text-[#E0E0E0]">
                <input
                  type="checkbox"
                  checked={securitySettings.requireSpecialChars}
                  onChange={(e) => setSecuritySettings({ ...securitySettings, requireSpecialChars: e.target.checked })}
                  className="rounded"
                />
                Require Special Characters
              </label>
            </div>
          ) : (
            <div className="text-sm text-[#E0E0E0]">
              {[
                securitySettings.requireUppercase && 'Uppercase',
                securitySettings.requireNumbers && 'Numbers',
                securitySettings.requireSpecialChars && 'Special Chars'
              ].filter(Boolean).join(', ')}
            </div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">2FA Enforcement</label>
          {isEditing ? (
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={securitySettings.require2FA}
                onChange={(e) => setSecuritySettings({ ...securitySettings, require2FA: e.target.checked })}
                className="rounded"
              />
              <span className="text-xs text-[#E0E0E0]">Require 2FA for all users</span>
            </label>
          ) : (
            <div className="text-sm text-[#E0E0E0]">{securitySettings.require2FA ? 'Enabled' : 'Disabled'}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">IP Restriction Mode</label>
          {isEditing ? (
            <select
              value={securitySettings.ipRestrictionMode}
              onChange={(e) => setSecuritySettings({ ...securitySettings, ipRestrictionMode: e.target.value as 'whitelist' | 'blacklist' | 'none' })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            >
              <option value="none">None</option>
              <option value="whitelist">Whitelist</option>
              <option value="blacklist">Blacklist</option>
            </select>
          ) : (
            <div className="text-sm text-[#E0E0E0] capitalize">{securitySettings.ipRestrictionMode}</div>
          )}
        </div>
      </div>
    );
  };

  const renderNotificationsSection = () => {
    const isEditing = editingSection === 'notifications';
    return (
      <div className="grid grid-cols-2 gap-4">
        <div className="col-span-2">
          <label className="block text-xs text-[#888] mb-2">Notification Channels</label>
          {isEditing ? (
            <div className="flex gap-4">
              <label className="flex items-center gap-2 text-xs text-[#E0E0E0]">
                <input
                  type="checkbox"
                  checked={notificationSettings.emailEnabled}
                  onChange={(e) => setNotificationSettings({ ...notificationSettings, emailEnabled: e.target.checked })}
                  className="rounded"
                />
                Email Notifications
              </label>
              <label className="flex items-center gap-2 text-xs text-[#E0E0E0]">
                <input
                  type="checkbox"
                  checked={notificationSettings.smsEnabled}
                  onChange={(e) => setNotificationSettings({ ...notificationSettings, smsEnabled: e.target.checked })}
                  className="rounded"
                />
                SMS Notifications
              </label>
              <label className="flex items-center gap-2 text-xs text-[#E0E0E0]">
                <input
                  type="checkbox"
                  checked={notificationSettings.pushEnabled}
                  onChange={(e) => setNotificationSettings({ ...notificationSettings, pushEnabled: e.target.checked })}
                  className="rounded"
                />
                Push Notifications
              </label>
            </div>
          ) : (
            <div className="text-sm text-[#E0E0E0]">
              {[
                notificationSettings.emailEnabled && 'Email',
                notificationSettings.smsEnabled && 'SMS',
                notificationSettings.pushEnabled && 'Push'
              ].filter(Boolean).join(', ')}
            </div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">SMTP Host</label>
          {isEditing ? (
            <input
              type="text"
              value={notificationSettings.smtpHost}
              onChange={(e) => setNotificationSettings({ ...notificationSettings, smtpHost: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{notificationSettings.smtpHost}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">SMTP Port</label>
          {isEditing ? (
            <input
              type="number"
              value={notificationSettings.smtpPort}
              onChange={(e) => setNotificationSettings({ ...notificationSettings, smtpPort: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{notificationSettings.smtpPort}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">SMTP User</label>
          {isEditing ? (
            <input
              type="text"
              value={notificationSettings.smtpUser}
              onChange={(e) => setNotificationSettings({ ...notificationSettings, smtpUser: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{notificationSettings.smtpUser}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">From Address</label>
          {isEditing ? (
            <input
              type="text"
              value={notificationSettings.smtpFrom}
              onChange={(e) => setNotificationSettings({ ...notificationSettings, smtpFrom: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{notificationSettings.smtpFrom}</div>
          )}
        </div>
      </div>
    );
  };

  const renderMaintenanceSection = () => {
    const isEditing = editingSection === 'maintenance';
    return (
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-xs text-[#888] mb-1">Maintenance Mode</label>
          {isEditing ? (
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={maintenanceSettings.maintenanceMode}
                onChange={(e) => setMaintenanceSettings({ ...maintenanceSettings, maintenanceMode: e.target.checked })}
                className="rounded"
              />
              <span className="text-xs text-[#E0E0E0]">Enable maintenance mode</span>
            </label>
          ) : (
            <div className={`text-sm ${maintenanceSettings.maintenanceMode ? 'text-[#F59E0B]' : 'text-[#10B981]'}`}>
              {maintenanceSettings.maintenanceMode ? 'Enabled' : 'Disabled'}
            </div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Data Retention (days)</label>
          {isEditing ? (
            <input
              type="number"
              value={maintenanceSettings.dataRetentionDays}
              onChange={(e) => setMaintenanceSettings({ ...maintenanceSettings, dataRetentionDays: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{maintenanceSettings.dataRetentionDays} days</div>
          )}
        </div>
        <div className="col-span-2">
          <label className="block text-xs text-[#888] mb-1">Maintenance Message</label>
          {isEditing ? (
            <textarea
              value={maintenanceSettings.maintenanceMessage}
              onChange={(e) => setMaintenanceSettings({ ...maintenanceSettings, maintenanceMessage: e.target.value })}
              rows={3}
              className="w-full px-3 py-2 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded resize-none"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{maintenanceSettings.maintenanceMessage}</div>
          )}
        </div>
        <div className="col-span-2">
          <label className="block text-xs text-[#888] mb-1">Scheduled Downtime</label>
          {isEditing ? (
            <input
              type="text"
              value={maintenanceSettings.scheduledDowntime}
              onChange={(e) => setMaintenanceSettings({ ...maintenanceSettings, scheduledDowntime: e.target.value })}
              placeholder="e.g., 2026-02-15 02:00 - 04:00 UTC"
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{maintenanceSettings.scheduledDowntime}</div>
          )}
        </div>
      </div>
    );
  };

  const renderApiSection = () => {
    const isEditing = editingSection === 'api';
    return (
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-xs text-[#888] mb-1">Rate Limit (requests/minute)</label>
          {isEditing ? (
            <input
              type="number"
              value={apiSettings.rateLimit}
              onChange={(e) => setApiSettings({ ...apiSettings, rateLimit: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{apiSettings.rateLimit} req/min</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">API Version</label>
          {isEditing ? (
            <input
              type="text"
              value={apiSettings.apiVersion}
              onChange={(e) => setApiSettings({ ...apiSettings, apiVersion: e.target.value })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{apiSettings.apiVersion}</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">WebSocket Max Connections</label>
          {isEditing ? (
            <input
              type="number"
              value={apiSettings.wsMaxConnections}
              onChange={(e) => setApiSettings({ ...apiSettings, wsMaxConnections: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{apiSettings.wsMaxConnections.toLocaleString()} connections</div>
          )}
        </div>
        <div>
          <label className="block text-xs text-[#888] mb-1">Max Payload Size (MB)</label>
          {isEditing ? (
            <input
              type="number"
              value={apiSettings.maxPayloadSize}
              onChange={(e) => setApiSettings({ ...apiSettings, maxPayloadSize: parseInt(e.target.value) })}
              className="w-full px-3 py-1.5 bg-[#1E2026] border border-[#383A42] text-[#E0E0E0] text-xs rounded"
            />
          ) : (
            <div className="text-sm text-[#E0E0E0]">{apiSettings.maxPayloadSize} MB</div>
          )}
        </div>
      </div>
    );
  };

  // Section config
  const sections: { key: SectionKey; label: string; icon: React.ReactNode; render: () => React.ReactNode }[] = [
    { key: 'general', label: 'General', icon: <Globe size={16} />, render: renderGeneralSection },
    { key: 'trading', label: 'Trading', icon: <Activity size={16} />, render: renderTradingSection },
    { key: 'security', label: 'Security', icon: <Shield size={16} />, render: renderSecuritySection },
    { key: 'notifications', label: 'Notifications', icon: <Bell size={16} />, render: renderNotificationsSection },
    { key: 'maintenance', label: 'Maintenance', icon: <Wrench size={16} />, render: renderMaintenanceSection },
    { key: 'api', label: 'API', icon: <Database size={16} />, render: renderApiSection },
  ];

  return (
    <div className="flex flex-col h-full bg-[#121316]">
      {/* Header */}
      <div className="h-12 bg-[#1E2026] border-b border-[#383A42] px-4 flex items-center justify-between flex-shrink-0">
        <div className="flex items-center gap-3">
          <Settings size={18} className="text-[#F5C542]" />
          <h1 className="text-sm font-bold text-[#E0E0E0]">Platform Configuration</h1>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-6">
        <div className="max-w-5xl mx-auto space-y-4">
          {/* Configuration Sections */}
          {sections.map((section) => {
            const isExpanded = expandedSections.has(section.key);
            const isEditing = editingSection === section.key;

            return (
              <div key={section.key} className="bg-[#1E2026] border border-[#383A42] rounded">
                {/* Section Header */}
                <div
                  className="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-[#25272E] transition-colors"
                  onClick={() => !isEditing && toggleSection(section.key)}
                >
                  <div className="flex items-center gap-3">
                    {isExpanded ? <ChevronDown size={16} className="text-[#888]" /> : <ChevronRight size={16} className="text-[#888]" />}
                    <div className="text-[#F5C542]">{section.icon}</div>
                    <span className="text-sm font-semibold text-[#E0E0E0]">{section.label}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {isEditing ? (
                      <>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleSave(section.key);
                          }}
                          className="p-1.5 bg-[#10B981]/20 hover:bg-[#10B981]/30 rounded transition-colors"
                        >
                          <Save size={14} className="text-[#10B981]" />
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleCancel();
                          }}
                          className="p-1.5 bg-[#EF4444]/20 hover:bg-[#EF4444]/30 rounded transition-colors"
                        >
                          <X size={14} className="text-[#EF4444]" />
                        </button>
                      </>
                    ) : (
                      <>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleEdit(section.key);
                          }}
                          className="p-1.5 bg-[#3B82F6]/20 hover:bg-[#3B82F6]/30 rounded transition-colors"
                        >
                          <Edit2 size={14} className="text-[#3B82F6]" />
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleReset(section.key);
                          }}
                          className="p-1.5 bg-[#888]/20 hover:bg-[#888]/30 rounded transition-colors"
                        >
                          <RotateCcw size={14} className="text-[#888]" />
                        </button>
                      </>
                    )}
                  </div>
                </div>

                {/* Section Content */}
                {isExpanded && (
                  <div className="px-4 pb-4 pt-2 border-t border-[#383A42]">
                    {section.render()}
                  </div>
                )}
              </div>
            );
          })}

          {/* Settings History */}
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4 mt-6">
            <div className="flex items-center gap-2 mb-3">
              <Clock size={16} className="text-[#F5C542]" />
              <h2 className="text-sm font-semibold text-[#E0E0E0]">Recent Changes</h2>
            </div>
            <div className="space-y-2">
              {settingsHistory.map((entry) => (
                <div key={entry.id} className="flex items-center justify-between p-2 bg-[#121316] rounded text-xs">
                  <div className="flex items-center gap-3">
                    <User size={12} className="text-[#888]" />
                    <span className="text-[#E0E0E0]">{entry.changedBy}</span>
                    <span className="text-[#888]">changed</span>
                    <span className="text-[#F5C542]">{entry.section} &gt; {entry.field}</span>
                    <span className="text-[#888]">from</span>
                    <span className="text-[#EF4444]">{entry.oldValue}</span>
                    <span className="text-[#888]">to</span>
                    <span className="text-[#10B981]">{entry.newValue}</span>
                  </div>
                  <span className="text-[#888]">{entry.changedAt}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
