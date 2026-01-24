import { UserPlus, MonitorPlay, Zap, RefreshCw, Layers, Search, Server, Signal } from 'lucide-react';

export default function TopToolbar() {
    return (
        <header className="h-9 bg-charcoal-950 border-b border-charcoal-border flex items-center px-1 shadow-sm flex-shrink-0 z-50 select-none">
            {/* App Branding */}
            <div className="px-3 font-bold text-zinc-200 tracking-tight text-xs select-none">RTX MANAGER</div>
            <div className="h-5 w-[1px] bg-charcoal-border mx-2"></div>

            {/* Toolbar Actions */}
            <div className="flex items-center gap-1">
                <button className="flex items-center gap-1.5 px-2 py-1 bg-charcoal-900 hover:bg-charcoal-800 border border-charcoal-border rounded-none text-xs text-zinc-300 transition-colors group">
                    <UserPlus size={13} className="text-rtx-yellow group-hover:text-rtx-hover" />
                    <span>New Account</span>
                </button>

                <button className="flex items-center gap-1.5 px-2 py-1 bg-charcoal-900 hover:bg-charcoal-800 border border-charcoal-border rounded-none text-xs text-zinc-300 transition-colors">
                    <MonitorPlay size={13} className="text-zinc-400" />
                    <span>Connect</span>
                </button>

                <button className="flex items-center gap-1.5 px-2 py-1 bg-charcoal-900 hover:bg-charcoal-800 border border-charcoal-border rounded-none text-xs text-zinc-300 transition-colors">
                    <Zap size={13} className="text-zinc-400" />
                    <span>Automation</span>
                </button>

                <div className="w-[1px] h-5 bg-charcoal-border mx-1"></div>

                <button className="p-1 hover:bg-charcoal-800 border border-transparent rounded-none text-zinc-400 hover:text-white" title="Refresh">
                    <RefreshCw size={14} />
                </button>
                <button className="p-1 hover:bg-charcoal-800 border border-transparent rounded-none text-zinc-400 hover:text-white" title="Filter">
                    <Layers size={14} />
                </button>
            </div>

            <div className="flex-1"></div>

            {/* Right Side: Search & Server Status */}
            <div className="flex items-center gap-4">
                {/* Global Search */}
                <div className="flex items-center bg-charcoal-900 border border-charcoal-border h-6 w-64 px-2 hover:border-zinc-600 transition-colors">
                    <Search size={11} className="text-zinc-500 mr-2" />
                    <input
                        className="bg-transparent border-none text-xs w-full focus:outline-none text-white placeholder-zinc-600 font-sans"
                        placeholder="Search Login / Name / Email / Group"
                    />
                </div>

                <div className="h-5 w-[1px] bg-charcoal-border"></div>

                {/* Server Info */}
                <div className="flex items-center gap-3 text-xs">
                    <div className="flex flex-col items-end leading-none">
                        <span className="text-zinc-300 font-medium">RTX-Live-01</span>
                        <span className="text-[10px] text-rtx-active text-right">LIVE</span>
                    </div>
                    <div className="flex items-center gap-1.5 text-zinc-400 bg-charcoal-900 px-2 py-1 border border-charcoal-border">
                        <Signal size={12} className="text-rtx-active" />
                        <span className="font-mono">12ms</span>
                    </div>
                </div>
            </div>
        </header>
    );
}
