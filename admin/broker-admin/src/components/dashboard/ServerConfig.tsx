'use client';

import { useState } from 'react';
import {
    Server,
    Settings,
    Shield,
    Clock,
    Wifi,
    Wrench,
    Save,
    RotateCcw,
    Activity,
    AlertTriangle,
    CheckCircle,
    Download,
    FileText,
} from 'lucide-react';

type ConfigSection = 'trading' | 'risk' | 'session' | 'connection' | 'maintenance';

interface TradingSettings {
    defaultLeverage: string;
    maxLotSize: number;
    minLotSize: number;
    lotStep: number;
    stopLevel: number;
    freezeLevel: number;
    tradeTimeout: number;
    executionMode: 'market' | 'instant';
}

interface RiskSettings {
    marginCallLevel: number;
    stopOutLevel: number;
    maxOpenPositions: number;
    maxPendingOrders: number;
    maxTotalExposure: number;
    negativeBalanceProtection: boolean;
}

interface SessionSettings {
    jwtExpiry: number;
    maxConcurrentSessions: number;
    sessionTimeout: number;
    force2FA: boolean;
    ipRestriction: boolean;
}

interface ConnectionSettings {
    wsHeartbeat: number;
    wsMaxConnections: number;
    apiRateLimit: number;
    corsOrigins: string;
    lpTimeout: number;
}

interface MaintenanceSettings {
    maintenanceMode: boolean;
    maintenanceMessage: string;
    scheduledFrom: string;
    scheduledTo: string;
}

interface ConfigChange {
    id: string;
    timestamp: Date;
    admin: string;
    field: string;
    oldValue: string;
    newValue: string;
}

const defaultTradingSettings: TradingSettings = {
    defaultLeverage: '1:100',
    maxLotSize: 100,
    minLotSize: 0.01,
    lotStep: 0.01,
    stopLevel: 10,
    freezeLevel: 5,
    tradeTimeout: 30,
    executionMode: 'market',
};

const defaultRiskSettings: RiskSettings = {
    marginCallLevel: 100,
    stopOutLevel: 50,
    maxOpenPositions: 100,
    maxPendingOrders: 50,
    maxTotalExposure: 10000000,
    negativeBalanceProtection: true,
};

const defaultSessionSettings: SessionSettings = {
    jwtExpiry: 24,
    maxConcurrentSessions: 3,
    sessionTimeout: 30,
    force2FA: false,
    ipRestriction: false,
};

const defaultConnectionSettings: ConnectionSettings = {
    wsHeartbeat: 30,
    wsMaxConnections: 10000,
    apiRateLimit: 1000,
    corsOrigins: '*',
    lpTimeout: 10,
};

const defaultMaintenanceSettings: MaintenanceSettings = {
    maintenanceMode: false,
    maintenanceMessage: 'The platform is currently under maintenance. We will be back shortly.',
    scheduledFrom: '',
    scheduledTo: '',
};

const generateChangeHistory = (): ConfigChange[] => {
    const now = new Date();
    return [
        {
            id: '1',
            timestamp: new Date(now.getTime() - 5 * 60000),
            admin: 'admin@rtx.com',
            field: 'Trading.defaultLeverage',
            oldValue: '1:200',
            newValue: '1:100',
        },
        {
            id: '2',
            timestamp: new Date(now.getTime() - 15 * 60000),
            admin: 'admin@rtx.com',
            field: 'Risk.marginCallLevel',
            oldValue: '120',
            newValue: '100',
        },
        {
            id: '3',
            timestamp: new Date(now.getTime() - 30 * 60000),
            admin: 'admin@rtx.com',
            field: 'Session.force2FA',
            oldValue: 'false',
            newValue: 'true',
        },
        {
            id: '4',
            timestamp: new Date(now.getTime() - 45 * 60000),
            admin: 'admin@rtx.com',
            field: 'Connection.wsHeartbeat',
            oldValue: '60',
            newValue: '30',
        },
        {
            id: '5',
            timestamp: new Date(now.getTime() - 60 * 60000),
            admin: 'admin@rtx.com',
            field: 'Trading.maxLotSize',
            oldValue: '50',
            newValue: '100',
        },
        {
            id: '6',
            timestamp: new Date(now.getTime() - 90 * 60000),
            admin: 'admin@rtx.com',
            field: 'Risk.negativeBalanceProtection',
            oldValue: 'false',
            newValue: 'true',
        },
        {
            id: '7',
            timestamp: new Date(now.getTime() - 120 * 60000),
            admin: 'admin@rtx.com',
            field: 'Session.jwtExpiry',
            oldValue: '48',
            newValue: '24',
        },
        {
            id: '8',
            timestamp: new Date(now.getTime() - 150 * 60000),
            admin: 'admin@rtx.com',
            field: 'Connection.apiRateLimit',
            oldValue: '500',
            newValue: '1000',
        },
        {
            id: '9',
            timestamp: new Date(now.getTime() - 180 * 60000),
            admin: 'admin@rtx.com',
            field: 'Trading.executionMode',
            oldValue: 'instant',
            newValue: 'market',
        },
        {
            id: '10',
            timestamp: new Date(now.getTime() - 210 * 60000),
            admin: 'admin@rtx.com',
            field: 'Risk.stopOutLevel',
            oldValue: '30',
            newValue: '50',
        },
    ];
};

const formatTimeAgo = (date: Date): string => {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    return `${diffHours}h ago`;
};

const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    return `${days}d ${hours}h ${mins}m`;
};

export default function ServerConfig() {
    const [activeSection, setActiveSection] = useState<ConfigSection>('trading');
    const [tradingSettings, setTradingSettings] = useState<TradingSettings>(defaultTradingSettings);
    const [riskSettings, setRiskSettings] = useState<RiskSettings>(defaultRiskSettings);
    const [sessionSettings, setSessionSettings] = useState<SessionSettings>(defaultSessionSettings);
    const [connectionSettings, setConnectionSettings] = useState<ConnectionSettings>(defaultConnectionSettings);
    const [maintenanceSettings, setMaintenanceSettings] = useState<MaintenanceSettings>(defaultMaintenanceSettings);
    const [changeHistory] = useState<ConfigChange[]>(generateChangeHistory());
    const [hasChanges, setHasChanges] = useState(false);

    // Mock server info
    const serverInfo = {
        name: 'RTX-Live-01',
        version: 'v2.4.0',
        uptime: 2592000, // 30 days in seconds
        environment: 'production' as 'development' | 'staging' | 'production',
        goVersion: 'go1.21.5',
        os: 'Linux 5.15.0-91-generic',
        cpuUsage: 45,
        memoryUsage: 62,
        lastRestart: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
    };

    const handleSave = () => {
        // In real app, this would POST to API
        console.log('Saving configuration:', {
            tradingSettings,
            riskSettings,
            sessionSettings,
            connectionSettings,
            maintenanceSettings,
        });
        setHasChanges(false);
        alert('Configuration saved successfully!');
    };

    const handleReset = () => {
        if (confirm('Are you sure you want to reset all settings to defaults?')) {
            setTradingSettings(defaultTradingSettings);
            setRiskSettings(defaultRiskSettings);
            setSessionSettings(defaultSessionSettings);
            setConnectionSettings(defaultConnectionSettings);
            setMaintenanceSettings(defaultMaintenanceSettings);
            setHasChanges(false);
        }
    };

    const getEnvBadgeColor = (env: string): string => {
        switch (env) {
            case 'production':
                return 'bg-red-500/10 text-red-400 border-red-500/20';
            case 'staging':
                return 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20';
            case 'development':
                return 'bg-green-500/10 text-green-400 border-green-500/20';
            default:
                return 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20';
        }
    };

    return (
        <div className="h-full flex flex-col bg-[#121316] overflow-hidden">
            {/* Header */}
            <div className="px-6 py-4 border-b border-[#383A42] flex items-center justify-between flex-shrink-0">
                <div className="flex items-center gap-2">
                    <Server size={20} className="text-[#F5C542]" />
                    <h1 className="text-xl font-bold text-white">Server Configuration</h1>
                </div>
                <div className="flex items-center gap-2">
                    <button
                        onClick={handleReset}
                        className="px-3 py-1.5 bg-[#E74C3C] hover:bg-[#C0392B] text-white text-sm rounded flex items-center gap-2"
                    >
                        <RotateCcw size={14} />
                        Reset to Defaults
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={!hasChanges}
                        className={`px-3 py-1.5 text-white text-sm rounded flex items-center gap-2 ${
                            hasChanges
                                ? 'bg-[#2ECC71] hover:bg-[#27AE60]'
                                : 'bg-[#25272E] text-[#666] cursor-not-allowed'
                        }`}
                    >
                        <Save size={14} />
                        Save Changes
                    </button>
                </div>
            </div>

            {/* Maintenance Mode Warning */}
            {maintenanceSettings.maintenanceMode && (
                <div className="px-6 py-3 bg-yellow-900/20 border-b border-yellow-600 flex items-center gap-3">
                    <AlertTriangle size={20} className="text-yellow-400" />
                    <div>
                        <div className="font-semibold text-yellow-400">Maintenance Mode Enabled</div>
                        <div className="text-sm text-yellow-200">{maintenanceSettings.maintenanceMessage}</div>
                    </div>
                </div>
            )}

            <div className="flex-1 flex overflow-hidden">
                {/* Left: Main Content */}
                <div className="flex-1 overflow-y-auto custom-scrollbar">
                    {/* Server Info Card */}
                    <div className="p-6">
                        <div className="bg-[#1E2026] border border-[#383A42] rounded p-6">
                            <div className="flex items-start justify-between mb-4">
                                <div>
                                    <div className="flex items-center gap-3 mb-2">
                                        <h2 className="text-lg font-bold text-white">{serverInfo.name}</h2>
                                        <span className="px-2 py-1 bg-[#2980B9] text-white text-xs rounded">
                                            {serverInfo.version}
                                        </span>
                                        <span className={`px-2 py-1 text-xs rounded border ${getEnvBadgeColor(serverInfo.environment)}`}>
                                            {serverInfo.environment.toUpperCase()}
                                        </span>
                                    </div>
                                    <div className="text-sm text-[#888] space-y-1">
                                        <div>Uptime: {formatUptime(serverInfo.uptime)}</div>
                                        <div>{serverInfo.goVersion} • {serverInfo.os}</div>
                                        <div>Last restart: {serverInfo.lastRestart.toLocaleString()}</div>
                                    </div>
                                </div>
                                <CheckCircle size={24} className="text-green-400" />
                            </div>

                            {/* CPU and Memory Usage */}
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <div className="flex items-center justify-between text-sm mb-2">
                                        <span className="text-[#CCC]">CPU Usage</span>
                                        <span className="text-white font-semibold">{serverInfo.cpuUsage}%</span>
                                    </div>
                                    <div className="h-2 bg-[#25272E] rounded overflow-hidden">
                                        <div
                                            className="h-full bg-blue-500"
                                            style={{ width: `${serverInfo.cpuUsage}%` }}
                                        ></div>
                                    </div>
                                </div>
                                <div>
                                    <div className="flex items-center justify-between text-sm mb-2">
                                        <span className="text-[#CCC]">Memory Usage</span>
                                        <span className="text-white font-semibold">{serverInfo.memoryUsage}%</span>
                                    </div>
                                    <div className="h-2 bg-[#25272E] rounded overflow-hidden">
                                        <div
                                            className="h-full bg-green-500"
                                            style={{ width: `${serverInfo.memoryUsage}%` }}
                                        ></div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Section Tabs */}
                    <div className="px-6 pb-4">
                        <div className="flex gap-2 border-b border-[#383A42]">
                            <button
                                onClick={() => setActiveSection('trading')}
                                className={`px-4 py-2 text-sm font-semibold border-b-2 transition-colors ${
                                    activeSection === 'trading'
                                        ? 'border-[#F5C542] text-white'
                                        : 'border-transparent text-[#888] hover:text-white'
                                }`}
                            >
                                <div className="flex items-center gap-2">
                                    <Activity size={14} />
                                    Trading
                                </div>
                            </button>
                            <button
                                onClick={() => setActiveSection('risk')}
                                className={`px-4 py-2 text-sm font-semibold border-b-2 transition-colors ${
                                    activeSection === 'risk'
                                        ? 'border-[#F5C542] text-white'
                                        : 'border-transparent text-[#888] hover:text-white'
                                }`}
                            >
                                <div className="flex items-center gap-2">
                                    <Shield size={14} />
                                    Risk
                                </div>
                            </button>
                            <button
                                onClick={() => setActiveSection('session')}
                                className={`px-4 py-2 text-sm font-semibold border-b-2 transition-colors ${
                                    activeSection === 'session'
                                        ? 'border-[#F5C542] text-white'
                                        : 'border-transparent text-[#888] hover:text-white'
                                }`}
                            >
                                <div className="flex items-center gap-2">
                                    <Clock size={14} />
                                    Session
                                </div>
                            </button>
                            <button
                                onClick={() => setActiveSection('connection')}
                                className={`px-4 py-2 text-sm font-semibold border-b-2 transition-colors ${
                                    activeSection === 'connection'
                                        ? 'border-[#F5C542] text-white'
                                        : 'border-transparent text-[#888] hover:text-white'
                                }`}
                            >
                                <div className="flex items-center gap-2">
                                    <Wifi size={14} />
                                    Connection
                                </div>
                            </button>
                            <button
                                onClick={() => setActiveSection('maintenance')}
                                className={`px-4 py-2 text-sm font-semibold border-b-2 transition-colors ${
                                    activeSection === 'maintenance'
                                        ? 'border-[#F5C542] text-white'
                                        : 'border-transparent text-[#888] hover:text-white'
                                }`}
                            >
                                <div className="flex items-center gap-2">
                                    <Wrench size={14} />
                                    Maintenance
                                </div>
                            </button>
                        </div>
                    </div>

                    {/* Configuration Forms */}
                    <div className="px-6 pb-6">
                        {/* Trading Settings */}
                        {activeSection === 'trading' && (
                            <div className="bg-[#1E2026] border border-[#383A42] rounded p-6">
                                <h3 className="text-lg font-semibold text-white mb-4">Trading Settings</h3>
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Default Leverage</label>
                                        <select
                                            value={tradingSettings.defaultLeverage}
                                            onChange={(e) => {
                                                setTradingSettings({ ...tradingSettings, defaultLeverage: e.target.value });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        >
                                            <option value="1:10">1:10</option>
                                            <option value="1:20">1:20</option>
                                            <option value="1:50">1:50</option>
                                            <option value="1:100">1:100</option>
                                            <option value="1:200">1:200</option>
                                            <option value="1:500">1:500</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Execution Mode</label>
                                        <select
                                            value={tradingSettings.executionMode}
                                            onChange={(e) => {
                                                setTradingSettings({ ...tradingSettings, executionMode: e.target.value as 'market' | 'instant' });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        >
                                            <option value="market">Market Execution</option>
                                            <option value="instant">Instant Execution</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Maximum Lot Size</label>
                                        <input
                                            type="number"
                                            value={tradingSettings.maxLotSize}
                                            onChange={(e) => {
                                                setTradingSettings({ ...tradingSettings, maxLotSize: parseFloat(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Minimum Lot Size</label>
                                        <input
                                            type="number"
                                            step="0.01"
                                            value={tradingSettings.minLotSize}
                                            onChange={(e) => {
                                                setTradingSettings({ ...tradingSettings, minLotSize: parseFloat(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Lot Step</label>
                                        <input
                                            type="number"
                                            step="0.01"
                                            value={tradingSettings.lotStep}
                                            onChange={(e) => {
                                                setTradingSettings({ ...tradingSettings, lotStep: parseFloat(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Stop Level (points)</label>
                                        <input
                                            type="number"
                                            value={tradingSettings.stopLevel}
                                            onChange={(e) => {
                                                setTradingSettings({ ...tradingSettings, stopLevel: parseInt(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                        <div className="text-xs text-[#666] mt-1">Minimum SL/TP distance from current price</div>
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Freeze Level (points)</label>
                                        <input
                                            type="number"
                                            value={tradingSettings.freezeLevel}
                                            onChange={(e) => {
                                                setTradingSettings({ ...tradingSettings, freezeLevel: parseInt(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                        <div className="text-xs text-[#666] mt-1">Order modification freeze distance</div>
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Trade Timeout (seconds)</label>
                                        <input
                                            type="number"
                                            value={tradingSettings.tradeTimeout}
                                            onChange={(e) => {
                                                setTradingSettings({ ...tradingSettings, tradeTimeout: parseInt(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                </div>
                            </div>
                        )}

                        {/* Risk Settings */}
                        {activeSection === 'risk' && (
                            <div className="bg-[#1E2026] border border-[#383A42] rounded p-6">
                                <h3 className="text-lg font-semibold text-white mb-4">Risk Management Settings</h3>
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Margin Call Level (%)</label>
                                        <input
                                            type="number"
                                            value={riskSettings.marginCallLevel}
                                            onChange={(e) => {
                                                setRiskSettings({ ...riskSettings, marginCallLevel: parseFloat(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Stop Out Level (%)</label>
                                        <input
                                            type="number"
                                            value={riskSettings.stopOutLevel}
                                            onChange={(e) => {
                                                setRiskSettings({ ...riskSettings, stopOutLevel: parseFloat(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Max Open Positions per Account</label>
                                        <input
                                            type="number"
                                            value={riskSettings.maxOpenPositions}
                                            onChange={(e) => {
                                                setRiskSettings({ ...riskSettings, maxOpenPositions: parseInt(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Max Pending Orders per Account</label>
                                        <input
                                            type="number"
                                            value={riskSettings.maxPendingOrders}
                                            onChange={(e) => {
                                                setRiskSettings({ ...riskSettings, maxPendingOrders: parseInt(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Maximum Total Exposure (USD)</label>
                                        <input
                                            type="number"
                                            value={riskSettings.maxTotalExposure}
                                            onChange={(e) => {
                                                setRiskSettings({ ...riskSettings, maxTotalExposure: parseFloat(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="flex items-center gap-2 text-sm text-[#CCC] cursor-pointer">
                                            <input
                                                type="checkbox"
                                                checked={riskSettings.negativeBalanceProtection}
                                                onChange={(e) => {
                                                    setRiskSettings({ ...riskSettings, negativeBalanceProtection: e.target.checked });
                                                    setHasChanges(true);
                                                }}
                                                className="w-4 h-4"
                                            />
                                            Enable Negative Balance Protection
                                        </label>
                                        <div className="text-xs text-[#666] mt-1 ml-6">Prevent accounts from going negative</div>
                                    </div>
                                </div>
                            </div>
                        )}

                        {/* Session Settings */}
                        {activeSection === 'session' && (
                            <div className="bg-[#1E2026] border border-[#383A42] rounded p-6">
                                <h3 className="text-lg font-semibold text-white mb-4">Session & Security Settings</h3>
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">JWT Token Expiry (hours)</label>
                                        <input
                                            type="number"
                                            value={sessionSettings.jwtExpiry}
                                            onChange={(e) => {
                                                setSessionSettings({ ...sessionSettings, jwtExpiry: parseInt(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Session Timeout (minutes)</label>
                                        <input
                                            type="number"
                                            value={sessionSettings.sessionTimeout}
                                            onChange={(e) => {
                                                setSessionSettings({ ...sessionSettings, sessionTimeout: parseInt(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Max Concurrent Sessions per Admin</label>
                                        <input
                                            type="number"
                                            value={sessionSettings.maxConcurrentSessions}
                                            onChange={(e) => {
                                                setSessionSettings({ ...sessionSettings, maxConcurrentSessions: parseInt(e.target.value) });
                                                setHasChanges(true);
                                            }}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        />
                                    </div>
                                    <div className="space-y-3">
                                        <label className="flex items-center gap-2 text-sm text-[#CCC] cursor-pointer">
                                            <input
                                                type="checkbox"
                                                checked={sessionSettings.force2FA}
                                                onChange={(e) => {
                                                    setSessionSettings({ ...sessionSettings, force2FA: e.target.checked });
                                                    setHasChanges(true);
                                                }}
                                                className="w-4 h-4"
                                            />
                                            Force 2FA for All Admins
                                        </label>
                                        <label className="flex items-center gap-2 text-sm text-[#CCC] cursor-pointer">
                                            <input
                                                type="checkbox"
                                                checked={sessionSettings.ipRestriction}
                                                onChange={(e) => {
                                                    setSessionSettings({ ...sessionSettings, ipRestriction: e.target.checked });
                                                    setHasChanges(true);
                                                }}
                                                className="w-4 h-4"
                                            />
                                            Enable IP Restriction
                                        </label>
                                    </div>
                                </div>
                            </div>
                        )}

                        {/* Connection Settings */}
                        {activeSection === 'connection' && (
                            <div className="bg-[#1E2026] border border-[#383A42] rounded p-6">
                                <h3 className="text-lg font-semibold text-white mb-4">Connection & Network Settings</h3>
                                <div className="space-y-4">
                                    <div className="grid grid-cols-2 gap-4">
                                        <div>
                                            <label className="block text-sm text-[#CCC] mb-2">WebSocket Heartbeat Interval (seconds)</label>
                                            <input
                                                type="number"
                                                value={connectionSettings.wsHeartbeat}
                                                onChange={(e) => {
                                                    setConnectionSettings({ ...connectionSettings, wsHeartbeat: parseInt(e.target.value) });
                                                    setHasChanges(true);
                                                }}
                                                className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm text-[#CCC] mb-2">WebSocket Max Connections</label>
                                            <input
                                                type="number"
                                                value={connectionSettings.wsMaxConnections}
                                                onChange={(e) => {
                                                    setConnectionSettings({ ...connectionSettings, wsMaxConnections: parseInt(e.target.value) });
                                                    setHasChanges(true);
                                                }}
                                                className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm text-[#CCC] mb-2">API Rate Limit (requests/minute)</label>
                                            <input
                                                type="number"
                                                value={connectionSettings.apiRateLimit}
                                                onChange={(e) => {
                                                    setConnectionSettings({ ...connectionSettings, apiRateLimit: parseInt(e.target.value) });
                                                    setHasChanges(true);
                                                }}
                                                className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm text-[#CCC] mb-2">LP Connection Timeout (seconds)</label>
                                            <input
                                                type="number"
                                                value={connectionSettings.lpTimeout}
                                                onChange={(e) => {
                                                    setConnectionSettings({ ...connectionSettings, lpTimeout: parseInt(e.target.value) });
                                                    setHasChanges(true);
                                                }}
                                                className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                            />
                                        </div>
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">CORS Allowed Origins</label>
                                        <textarea
                                            value={connectionSettings.corsOrigins}
                                            onChange={(e) => {
                                                setConnectionSettings({ ...connectionSettings, corsOrigins: e.target.value });
                                                setHasChanges(true);
                                            }}
                                            rows={3}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9] font-mono"
                                            placeholder="*, https://example.com, https://admin.example.com"
                                        ></textarea>
                                        <div className="text-xs text-[#666] mt-1">Comma-separated list of allowed origins. Use * for all.</div>
                                    </div>
                                </div>
                            </div>
                        )}

                        {/* Maintenance Settings */}
                        {activeSection === 'maintenance' && (
                            <div className="bg-[#1E2026] border border-[#383A42] rounded p-6">
                                <h3 className="text-lg font-semibold text-white mb-4">Maintenance & Operations</h3>
                                <div className="space-y-4">
                                    <div>
                                        <label className="flex items-center gap-2 text-sm text-[#CCC] cursor-pointer">
                                            <input
                                                type="checkbox"
                                                checked={maintenanceSettings.maintenanceMode}
                                                onChange={(e) => {
                                                    setMaintenanceSettings({ ...maintenanceSettings, maintenanceMode: e.target.checked });
                                                    setHasChanges(true);
                                                }}
                                                className="w-4 h-4"
                                            />
                                            Enable Maintenance Mode
                                        </label>
                                        <div className="text-xs text-[#666] mt-1 ml-6">
                                            Blocks all client access and shows maintenance message
                                        </div>
                                    </div>
                                    <div>
                                        <label className="block text-sm text-[#CCC] mb-2">Maintenance Message</label>
                                        <textarea
                                            value={maintenanceSettings.maintenanceMessage}
                                            onChange={(e) => {
                                                setMaintenanceSettings({ ...maintenanceSettings, maintenanceMessage: e.target.value });
                                                setHasChanges(true);
                                            }}
                                            rows={3}
                                            className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                        ></textarea>
                                    </div>
                                    <div className="grid grid-cols-2 gap-4">
                                        <div>
                                            <label className="block text-sm text-[#CCC] mb-2">Scheduled Maintenance From</label>
                                            <input
                                                type="datetime-local"
                                                value={maintenanceSettings.scheduledFrom}
                                                onChange={(e) => {
                                                    setMaintenanceSettings({ ...maintenanceSettings, scheduledFrom: e.target.value });
                                                    setHasChanges(true);
                                                }}
                                                className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm text-[#CCC] mb-2">Scheduled Maintenance To</label>
                                            <input
                                                type="datetime-local"
                                                value={maintenanceSettings.scheduledTo}
                                                onChange={(e) => {
                                                    setMaintenanceSettings({ ...maintenanceSettings, scheduledTo: e.target.value });
                                                    setHasChanges(true);
                                                }}
                                                className="w-full px-3 py-2 bg-[#25272E] border border-[#383A42] text-white text-sm rounded focus:outline-none focus:border-[#2980B9]"
                                            />
                                        </div>
                                    </div>
                                    <div className="flex gap-2 pt-2">
                                        <button
                                            onClick={() => alert('Backup initiated (placeholder)')}
                                            className="px-4 py-2 bg-[#2980B9] hover:bg-[#3498DB] text-white text-sm rounded flex items-center gap-2"
                                        >
                                            <Download size={14} />
                                            Backup Now
                                        </button>
                                        <button
                                            onClick={() => alert('Opening server logs (placeholder)')}
                                            className="px-4 py-2 bg-[#1E2026] hover:bg-[#25272E] border border-[#383A42] text-white text-sm rounded flex items-center gap-2"
                                        >
                                            <FileText size={14} />
                                            View Server Logs
                                        </button>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                {/* Right: Change History */}
                <div className="w-96 border-l border-[#383A42] bg-[#1E2026] flex flex-col">
                    <div className="px-4 py-3 border-b border-[#383A42]">
                        <h3 className="text-sm font-semibold text-white">Change History</h3>
                        <div className="text-xs text-[#666] mt-1">Last 10 configuration changes</div>
                    </div>
                    <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-3">
                        {changeHistory.map((change) => (
                            <div key={change.id} className="bg-[#121316] border border-[#383A42] rounded p-3">
                                <div className="flex items-start justify-between mb-2">
                                    <div className="text-xs font-semibold text-white">{change.field}</div>
                                    <div className="text-xs text-[#666]" title={change.timestamp.toLocaleString()}>
                                        {formatTimeAgo(change.timestamp)}
                                    </div>
                                </div>
                                <div className="text-xs text-[#888] mb-2">
                                    <span className="text-red-400 line-through">{change.oldValue}</span>
                                    {' → '}
                                    <span className="text-green-400">{change.newValue}</span>
                                </div>
                                <div className="text-xs text-[#666]">by {change.admin}</div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
