import React, { useState, useEffect } from 'react';
import {
    X,
    Monitor,
    BarChart2,
    TrendingUp,
    Cpu,
    Bell,
    Mail,
    Wifi,
    Globe,
    Server,
    Shield,
    Smartphone,
    HardDrive,
    Bot,
    Activity,
    Check,
    AlertTriangle
} from 'lucide-react';

interface OptionsDialogProps {
    isOpen: boolean;
    onClose: () => void;
}

type TabId = 'server' | 'charts' | 'trade' | 'automation' | 'events' | 'notifications' | 'email' | 'ftp' | 'community' | 'signals';

export const OptionsDialog: React.FC<OptionsDialogProps> = ({ isOpen, onClose }) => {
    const [activeTab, setActiveTab] = useState<TabId>('server');

    // Close on Escape
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape' && isOpen) onClose();
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [isOpen, onClose]);

    if (!isOpen) return null;

    const tabs: { id: TabId; label: string }[] = [
        { id: 'server', label: 'Server' },
        { id: 'charts', label: 'Charts' },
        { id: 'trade', label: 'Trade' },
        { id: 'automation', label: 'Automation & Bots' },
        { id: 'events', label: 'Events' },
        { id: 'notifications', label: 'Notifications' },
        { id: 'email', label: 'Email' },
        { id: 'ftp', label: 'FTP' },
        { id: 'community', label: 'Community' },
        { id: 'signals', label: 'Signals' },
    ];

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="w-[800px] h-[600px] bg-[#252528] rounded-lg shadow-2xl border border-zinc-700 flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
                {/* Header */}
                <div className="h-10 bg-[#1e1e1e] border-b border-zinc-700 flex items-center justify-between px-4 select-none">
                    <span className="text-sm font-medium text-zinc-200">Options</span>
                    <button onClick={onClose} className="text-zinc-500 hover:text-white transition-colors">
                        <X size={16} />
                    </button>
                </div>

                {/* Tabs */}
                <div className="bg-[#1e1e1e] border-b border-zinc-700 px-2 flex items-center gap-1 overflow-x-auto scrollbar-hide">
                    {tabs.map((tab) => (
                        <button
                            key={tab.id}
                            onClick={() => setActiveTab(tab.id)}
                            className={`px-3 py-2 text-[12px] font-medium border-b-2 transition-all whitespace-nowrap
                                ${activeTab === tab.id
                                    ? 'border-blue-500 text-blue-400'
                                    : 'border-transparent text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/50'}`}
                        >
                            {tab.label}
                        </button>
                    ))}
                </div>

                {/* Content */}
                <div className="flex-1 overflow-y-auto p-6 bg-[#252528]">
                    {activeTab === 'server' && <ServerTab />}
                    {activeTab === 'charts' && <ChartsTab />}
                    {activeTab === 'trade' && <TradeTab />}
                    {activeTab === 'automation' && <AutomationTab />}
                    {activeTab === 'events' && <PlaceholderTab icon={<Activity size={48} />} title="Events" desc="Configure sound notifications and system events." />}
                    {activeTab === 'notifications' && <PlaceholderTab icon={<Bell size={48} />} title="Notifications" desc="Manage push notifications to RTX5 Mobile." />}
                    {activeTab === 'email' && <PlaceholderTab icon={<Mail size={48} />} title="Email" desc="SMTP server configuration for alerts." />}
                    {activeTab === 'ftp' && <PlaceholderTab icon={<HardDrive size={48} />} title="FTP" desc="Automated report publishing settings." />}
                    {activeTab === 'community' && <PlaceholderTab icon={<Globe size={48} />} title="RTX5 Community" desc="Login to the proprietary RTX5 ecosystem." />}
                    {activeTab === 'signals' && <PlaceholderTab icon={<Wifi size={48} />} title="RTX5 Signals Hub" desc="Subscribe to institutional signal providers." />}
                </div>

                {/* Footer */}
                <div className="h-14 bg-[#1e1e1e] border-t border-zinc-700 flex items-center justify-end px-4 gap-3">
                    <button onClick={onClose} className="px-4 py-1.5 text-xs font-medium text-zinc-300 hover:bg-zinc-700 rounded transition-colors">
                        Cancel
                    </button>
                    <button onClick={onClose} className="px-4 py-1.5 text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white rounded transition-colors shadow-sm">
                        OK
                    </button>
                </div>
            </div>
        </div>
    );
};

const ServerTab = () => (
    <div className="space-y-6 text-xs text-zinc-300">
        <div className="grid grid-cols-[100px_1fr] gap-4 items-center">
            <label className="text-right text-zinc-400">Server:</label>
            <select className="w-full bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1.5 outline-none focus:border-blue-500">
                <option>FlexyMarkets-Server</option>
                <option>FlexyMarkets-Demo</option>
            </select>

            <label className="text-right text-zinc-400">Login:</label>
            <input type="text" value="900500" readOnly className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1.5 outline-none focus:border-blue-500" />

            <label className="text-right text-zinc-400">Password:</label>
            <div className="flex gap-2">
                <input type="password" value="********" readOnly className="flex-1 bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1.5 outline-none focus:border-blue-500" />
                <button className="px-3 py-1 bg-zinc-700 hover:bg-zinc-600 rounded text-zinc-200">Change</button>
            </div>
        </div>

        <div className="ml-[116px] space-y-3">
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" className="rounded bg-zinc-700 border-zinc-600 text-blue-500 focus:ring-0 focus:ring-offset-0" />
                <span>Enable proxy server</span>
                <button className="ml-auto px-3 py-0.5 bg-zinc-800 border border-zinc-700 rounded hover:bg-zinc-700">Proxy...</button>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" defaultChecked className="rounded bg-zinc-700 border-zinc-600 text-blue-500 focus:ring-0 focus:ring-offset-0" />
                <span>Keep personal settings and data at startup</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" defaultChecked className="rounded bg-zinc-700 border-zinc-600 text-blue-500 focus:ring-0 focus:ring-offset-0" />
                <span>Enable news</span>
            </label>
        </div>
    </div>
);

const ChartsTab = () => (
    <div className="space-y-4 text-xs text-zinc-300">
        <div className="space-y-2">
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" defaultChecked className="rounded bg-zinc-700 border-zinc-600 text-blue-500" />
                <span>Show trade history</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" defaultChecked className="rounded bg-zinc-700 border-zinc-600 text-blue-500" />
                <span>Show trade levels</span>
            </label>
            <div className="ml-6">
                <select className="w-full bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1">
                    <option>Enable dragging of trade levels</option>
                    <option>Disable dragging</option>
                </select>
            </div>
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" defaultChecked className="rounded bg-zinc-700 border-zinc-600 text-blue-500" />
                <span>Preload chart data for open positions and orders</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" className="rounded bg-zinc-700 border-zinc-600 text-blue-500" />
                <span>Show object properties after creation</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" defaultChecked className="rounded bg-zinc-700 border-zinc-600 text-blue-500" />
                <span>Select object after creation</span>
            </label>
        </div>

        <div className="grid grid-cols-[120px_1fr] gap-4 pt-4 border-t border-zinc-700/50">
            <label className="text-right text-zinc-400">Magnet sensitivity:</label>
            <select className="w-20 bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1">
                <option>10</option>
                <option>5</option>
                <option>20</option>
            </select>

            <label className="text-right text-zinc-400">Max bars in chart:</label>
            <select className="w-32 bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1">
                <option>100000</option>
                <option>Unlimited</option>
            </select>
        </div>
    </div>
);

const TradeTab = () => (
    <div className="space-y-5 text-xs text-zinc-300">
        <div className="grid grid-cols-[100px_1fr_1fr] gap-4 items-center">
            <label className="text-right text-zinc-400">Symbol:</label>
            <select className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1.5 focus:border-blue-500">
                <option>Automatic</option>
                <option>Last Used</option>
            </select>
            <div className="text-zinc-500 italic">EURUSD</div>

            <label className="text-right text-zinc-400">Volume:</label>
            <select className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1.5 focus:border-blue-500">
                <option>Last Used</option>
                <option>Default</option>
            </select>
            <input type="number" defaultValue="0.01" className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1.5 w-24" />

            <label className="text-right text-zinc-400">Deviation:</label>
            <select className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1.5 focus:border-blue-500">
                <option>Last Used</option>
                <option>Default</option>
            </select>
            <input type="number" defaultValue="0" className="bg-[#1e1e1e] border border-zinc-700 rounded px-2 py-1.5 w-24" />
        </div>

        <div className="pt-4 border-t border-zinc-700/50">
            <label className="flex items-center gap-2 cursor-pointer group">
                <div className="relative">
                    <input type="checkbox" className="peer sr-only" />
                    <div className="w-9 h-5 bg-zinc-700 rounded-full peer peer-checked:bg-blue-600 transition-colors"></div>
                    <div className="absolute left-1 top-1 w-3 h-3 bg-white rounded-full peer-checked:translate-x-4 transition-transform"></div>
                </div>
                <span className="font-medium group-hover:text-white transition-colors">One Click Trading</span>
            </label>
            <p className="mt-2 text-zinc-500 ml-11 text-[11px]">
                Allows performing trade operations with a single mouse click without additional confirmation.
            </p>
        </div>
    </div>
);

const AutomationTab = () => (
    <div className="space-y-5 text-xs text-zinc-300">
        <div className="space-y-3">
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" className="rounded bg-zinc-700 border-zinc-600 text-blue-500" />
                <span className="font-medium text-zinc-200">Enable automated strategies</span>
            </label>
            <div className="ml-6 space-y-2 text-zinc-400">
                <label className="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" defaultChecked disabled className="rounded bg-zinc-700 border-zinc-600 text-zinc-500 cursor-not-allowed" />
                    <span>Disable strategies when the account has been changed</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" defaultChecked disabled className="rounded bg-zinc-700 border-zinc-600 text-zinc-500 cursor-not-allowed" />
                    <span>Disable strategies when the profile has been changed</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" className="rounded bg-zinc-700 border-zinc-600 text-zinc-500" />
                    <span>Disable strategies when the charts symbol or period has been changed</span>
                </label>
            </div>
        </div>

        <div className="space-y-3 pt-4 border-t border-zinc-700/50">
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" defaultChecked className="rounded bg-zinc-700 border-zinc-600 text-blue-500" />
                <span className="flex items-center gap-2">
                    External library access (DLL)
                    <AlertTriangle size={12} className="text-amber-500" />
                    <span className="text-zinc-500">(potentially dangerous, enable only for trusted applications)</span>
                </span>
            </label>
        </div>

        <div className="space-y-2 pt-4 border-t border-zinc-700/50">
            <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" defaultChecked className="rounded bg-zinc-700 border-zinc-600 text-blue-500" />
                <span>Allow WebRequest for listed URL:</span>
            </label>
            <div className="border border-zinc-700 rounded bg-[#1e1e1e] h-32 overflow-y-auto">
                <div className="flex items-center gap-2 px-2 py-1.5 border-b border-zinc-800 hover:bg-zinc-800/50">
                    <Globe size={12} className="text-blue-400" />
                    <span>https://backend.yoforexai.com</span>
                </div>
                <div className="flex items-center gap-2 px-2 py-1.5 text-zinc-500 cursor-text hover:bg-zinc-800/50">
                    <span className="text-emerald-500 font-bold">+</span>
                    <span className="italic">add new URL like 'https://api.rtx-terminal.com'</span>
                </div>
            </div>
        </div>
    </div>
);

const PlaceholderTab = ({ icon, title, desc }: { icon: React.ReactNode, title: string, desc: string }) => (
    <div className="h-full flex flex-col items-center justify-center text-center space-y-4 opacity-50">
        <div className="text-zinc-600">{icon}</div>
        <div>
            <h3 className="text-lg font-medium text-zinc-300">{title}</h3>
            <p className="text-zinc-500 max-w-xs mx-auto">{desc}</p>
        </div>
    </div>
);
