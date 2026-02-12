'use client';

import { useState, useRef, useEffect } from 'react';
import { UserPlus, MonitorPlay, Zap, RefreshCw, Layers, Search, Server, Signal, LayoutGrid, Grid2x2, Grid3x3, SquareSplitVertical, LogOut } from 'lucide-react';
import { ChartLayout } from '../dashboard/ChartGrid';
import { useAuth } from '@/hooks/useAuth';
import NotificationCenter from '../dashboard/NotificationCenter';

interface TopToolbarProps {
    chartLayout?: ChartLayout;
    onLayoutChange?: (layout: ChartLayout) => void;
    onNavigate?: (viewId: string) => void;
}

export default function TopToolbar({ chartLayout = '1x1', onLayoutChange, onNavigate }: TopToolbarProps) {
    const { logout } = useAuth();
    const [showNotifications, setShowNotifications] = useState(false);
    const notificationRef = useRef<HTMLDivElement>(null);

    // Mock unread count - in real app, this would come from API/WebSocket
    const [unreadCount] = useState(12);

    const handleLogout = () => {
        if (confirm('Are you sure you want to logout?')) {
            logout();
        }
    };

    // Close dropdown when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (notificationRef.current && !notificationRef.current.contains(event.target as Node)) {
                setShowNotifications(false);
            }
        };

        if (showNotifications) {
            document.addEventListener('mousedown', handleClickOutside);
        }

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [showNotifications]);
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

                <div className="w-[1px] h-5 bg-charcoal-border mx-1"></div>

                {/* Notification Bell */}
                <div className="relative" ref={notificationRef}>
                    <button
                        onClick={() => setShowNotifications(!showNotifications)}
                        className="relative p-1 hover:bg-charcoal-800 border border-transparent rounded-none text-zinc-400 hover:text-white transition-colors"
                        title="Notifications"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"></path>
                            <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0"></path>
                        </svg>
                        {unreadCount > 0 && (
                            <span className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-[9px] font-bold rounded-full flex items-center justify-center">
                                {unreadCount > 9 ? '9+' : unreadCount}
                            </span>
                        )}
                    </button>

                    {/* Notification Dropdown */}
                    {showNotifications && (
                        <div className="absolute top-full right-0 mt-2 z-50">
                            <NotificationCenter
                                onNavigate={(viewId: string) => {
                                    if (onNavigate) onNavigate(viewId);
                                    setShowNotifications(false);
                                }}
                            />
                        </div>
                    )}
                </div>

                {/* Chart Layout Buttons */}
                {onLayoutChange && (
                    <>
                        <div className="w-[1px] h-5 bg-charcoal-border mx-1"></div>
                        <div className="flex items-center gap-1">
                            <button
                                className={`p-1 hover:bg-charcoal-800 border rounded-none transition-colors ${
                                    chartLayout === '1x1'
                                        ? 'border-rtx-yellow text-rtx-yellow'
                                        : 'border-transparent text-zinc-400 hover:text-white'
                                }`}
                                title="Single Chart (1x1)"
                                onClick={() => onLayoutChange('1x1')}
                            >
                                <LayoutGrid size={14} />
                            </button>
                            <button
                                className={`p-1 hover:bg-charcoal-800 border rounded-none transition-colors ${
                                    chartLayout === '2x1'
                                        ? 'border-rtx-yellow text-rtx-yellow'
                                        : 'border-transparent text-zinc-400 hover:text-white'
                                }`}
                                title="Two Charts Side by Side (2x1)"
                                onClick={() => onLayoutChange('2x1')}
                            >
                                <SquareSplitVertical size={14} />
                            </button>
                            <button
                                className={`p-1 hover:bg-charcoal-800 border rounded-none transition-colors ${
                                    chartLayout === '2x2'
                                        ? 'border-rtx-yellow text-rtx-yellow'
                                        : 'border-transparent text-zinc-400 hover:text-white'
                                }`}
                                title="Four Charts (2x2)"
                                onClick={() => onLayoutChange('2x2')}
                            >
                                <Grid2x2 size={14} />
                            </button>
                            <button
                                className={`p-1 hover:bg-charcoal-800 border rounded-none transition-colors ${
                                    chartLayout === '3x2'
                                        ? 'border-rtx-yellow text-rtx-yellow'
                                        : 'border-transparent text-zinc-400 hover:text-white'
                                }`}
                                title="Six Charts (3x2)"
                                onClick={() => onLayoutChange('3x2')}
                            >
                                <Grid3x3 size={14} />
                            </button>
                        </div>
                    </>
                )}
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

                <div className="h-5 w-[1px] bg-charcoal-border"></div>

                {/* Logout Button */}
                <button
                    onClick={handleLogout}
                    className="flex items-center gap-1.5 px-2 py-1 bg-charcoal-900 hover:bg-red-900/20 border border-charcoal-border hover:border-red-500 rounded-none text-xs text-zinc-300 hover:text-red-400 transition-colors"
                    title="Logout"
                >
                    <LogOut size={13} />
                    <span>Logout</span>
                </button>
            </div>
        </header>
    );
}
