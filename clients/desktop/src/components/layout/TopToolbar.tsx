import React, { useEffect } from 'react';
import {
    CandlestickChart,
    LineChart,
    MousePointer2,
    Crosshair,
    Minus,
    Type,
    Bell,
    Settings,
    BarChart3,
    Layers,
    ZoomIn,
    ZoomOut,
    Play,
    ChevronDown,
    LayoutTemplate,
    LayoutList
} from 'lucide-react';
import type { ChartType, Timeframe } from '../TradingChart';
import { useToolbarState } from '../../hooks/useToolbarState';
import { registerKeyboardShortcutsWithInputCheck } from '../../services/keyboardShortcuts';
import { useCommandBus } from '../../hooks/useCommandBus';
import type { Command } from '../../types/commands';

interface TopToolbarProps {
    chartType: ChartType;
    timeframe: Timeframe;
    onChartTypeChange: (type: ChartType) => void;
    onTimeframeChange: (tf: Timeframe) => void;
    onToggleDOM?: () => void;
}

import { MenuBar } from './MenuBar';
import { DrawingsDropdown } from './DrawingsDropdown';

export const TopToolbar: React.FC<TopToolbarProps> = ({
    chartType,
    timeframe,
    onChartTypeChange,
    onTimeframeChange,
    onToggleDOM
}) => {
    const { state: toolbarState, dispatch: dispatchToolbar } = useToolbarState();
    const { dispatch: dispatchCommand } = useCommandBus();

    // Register keyboard shortcuts on mount
    useEffect(() => {
        const unregister = registerKeyboardShortcutsWithInputCheck(dispatchCommand);
        return unregister;
    }, [dispatchCommand]);

    // Sync props with toolbar state
    useEffect(() => {
        if (chartType !== toolbarState.chartType) {
            dispatchToolbar({ type: 'SET_CHART_TYPE', chartType: chartType as any });
        }
    }, [chartType, toolbarState.chartType, dispatchToolbar]);

    useEffect(() => {
        if (timeframe !== toolbarState.timeframe) {
            dispatchToolbar({ type: 'SET_TIMEFRAME', timeframe });
        }
    }, [timeframe, toolbarState.timeframe, dispatchToolbar]);

    return (
        <div className="flex flex-col z-50 relative">
            {/* 1. Header Bar (Topmost) */}
            <HeaderBar
                symbol="XAUUSD"
                timeframe={timeframe}
                account="900500"
                server="HexyMarkets-Server"
            />

            {/* 2. Menu Bar (File, View, etc.) */}
            <MenuBar />

            {/* 3. Main Toolbar (Icons Row) */}
            <MainToolbar
                chartType={chartType}
                timeframe={timeframe}
                onChartTypeChange={onChartTypeChange}
                onTimeframeChange={onTimeframeChange}
                onToggleDOM={onToggleDOM}
                toolbarState={toolbarState}
                dispatchCommand={dispatchCommand}
                dispatchToolbar={dispatchToolbar}
            />
        </div>
    );
};

// --- Sub-Components ---

const HeaderBar = ({ symbol, timeframe, account, server }: { symbol: string, timeframe: string, account: string, server: string }) => (
    <div className="h-8 bg-[#1e1e1e] flex items-center justify-between px-3 text-[11px] font-medium border-b border-[#2a2e39] select-none text-zinc-400">
        {/* Left: Platform Logo & Account */}
        <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-zinc-300">
                <div className="w-4 h-4 rounded bg-gradient-to-br from-blue-500 to-teal-400 flex items-center justify-center text-[9px] text-white font-bold tracking-tight">
                    H
                </div>
                <span className="font-bold tracking-wide text-zinc-200">HEXY</span>
            </div>
            <div className="flex items-center gap-1.5 pl-3 border-l border-zinc-700/50">
                <span className="text-zinc-500">ACC:</span>
                <span className="text-zinc-300 font-mono">{account}</span>
                <span className="text-zinc-600 px-1">•</span>
                <span className="text-zinc-500">{server}</span>
            </div>
        </div>

        {/* Center: Active Symbol Info */}
        <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 flex items-center gap-2">
            <div className="bg-[#252525] px-3 py-0.5 rounded-full border border-zinc-800 flex items-center gap-2 shadow-sm">
                <span className="text-emerald-400 font-bold tracking-wider">{symbol}</span>
                <span className="w-[1px] h-3 bg-zinc-700"></span>
                <span className="text-blue-400 font-mono font-bold">{timeframe.toUpperCase()}</span>
            </div>
        </div>

        {/* Right: System Status */}
        <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5 px-2 py-0.5 bg-[#252525] rounded border border-zinc-800">
                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
                <span className="text-emerald-500 font-bold text-[10px]">CONNECTED</span>
                <span className="text-zinc-600 text-[10px] ml-1">12ms</span>
            </div>
            <div className="flex items-center gap-3 text-zinc-500 border-l border-zinc-700/50 pl-3">
                <Bell size={13} className="hover:text-zinc-300 cursor-pointer transition-colors" />
                <Settings size={13} className="hover:text-zinc-300 cursor-pointer transition-colors" />
            </div>
        </div>
    </div>
);

interface MainToolbarProps extends TopToolbarProps {
    toolbarState: any;
    dispatchCommand: (cmd: Command) => void;
    dispatchToolbar: (action: any) => void;
}

const MainToolbar = ({
    chartType,
    timeframe,
    onChartTypeChange,
    onTimeframeChange,
    onToggleDOM,
    toolbarState,
    dispatchCommand,
    dispatchToolbar
}: MainToolbarProps) => {
    return (
        <div className="h-10 bg-gradient-to-b from-[#252525] to-[#1e1e1e] border-b border-black flex items-center px-2 gap-3 shadow-md">
            {/* Section 1: Trading Actions */}
            <ToolbarGroup>
                <ToolButton
                    icon={<Play size={15} className="fill-emerald-500 text-emerald-500" />}
                    label="Algo Trading"
                    activeColor="emerald"
                    onClick={() => dispatchCommand({ type: 'TOGGLE_ALGO_TRADING', payload: {} })}
                />
                <ToolButton
                    icon={<Layers size={15} />}
                    label="New Order (F9)"
                    title="Open Order Panel (F9)"
                    onClick={() => {
                        // Simulate F9 key press to trigger order panel
                        const event = new KeyboardEvent('keydown', {
                            key: 'F9',
                            bubbles: true,
                            cancelable: true
                        });
                        window.dispatchEvent(event);
                    }}
                />
            </ToolbarGroup>

            <Divider />

            {/* Section 2: Chart Controls */}
            <ToolbarGroup>
                <ToolButton
                    active={chartType === 'bar'}
                    onClick={() => {
                        onChartTypeChange('bar');
                        dispatchCommand({ type: 'SET_CHART_TYPE', payload: { chartType: 'bar' } });
                    }}
                    icon={<BarChart3 size={16} />}
                    title="Bar Chart"
                />
                <ToolButton
                    active={chartType === 'candlestick'}
                    onClick={() => {
                        onChartTypeChange('candlestick');
                        dispatchCommand({ type: 'SET_CHART_TYPE', payload: { chartType: 'candlestick' } });
                    }}
                    icon={<CandlestickChart size={16} />}
                    title="Candlesticks"
                />
                <ToolButton
                    active={chartType === 'area'}
                    onClick={() => {
                        onChartTypeChange('area');
                        dispatchCommand({ type: 'SET_CHART_TYPE', payload: { chartType: 'area' } });
                    }}
                    icon={<LineChart size={16} />}
                    title="Line Chart"
                />
            </ToolbarGroup>

            <Divider />

            {/* Section 3: Tools */}
            <ToolbarGroup>
                <ToolButton
                    icon={<Crosshair size={16} />}
                    active={toolbarState.crosshairEnabled}
                    onClick={() => dispatchCommand({ type: 'TOGGLE_CROSSHAIR', payload: {} })}
                    title="Crosshair (C)"
                />
                <ToolButton
                    icon={<ZoomIn size={16} />}
                    onClick={() => dispatchCommand({ type: 'ZOOM_IN', payload: {} })}
                    title="Zoom In (+)"
                />
                <ToolButton
                    icon={<ZoomOut size={16} />}
                    onClick={() => dispatchCommand({ type: 'ZOOM_OUT', payload: {} })}
                    title="Zoom Out (-)"
                />
                <ToolButton
                    icon={<LayoutTemplate size={16} />}
                    onClick={() => dispatchCommand({ type: 'TILE_WINDOWS', payload: { mode: 'grid' } })}
                    title="Tile Windows (Grid Layout)"
                />
                <Divider />
                <ToolButton
                    icon={
                        <svg width="16" height="16" viewBox="0 0 18 18" fill="none" stroke="currentColor" strokeWidth="1.5">
                            <rect x="3" y="5" width="12" height="8" />
                            <path d="M15 9l-3-2v4z" fill="currentColor" stroke="none" />
                        </svg>
                    }
                    active={toolbarState.autoScrollEnabled}
                    onClick={() => dispatchToolbar({ type: 'TOGGLE_AUTO_SCROLL' })}
                    title="Auto Scroll to End"
                />
                <ToolButton
                    icon={
                        <svg width="16" height="16" viewBox="0 0 18 18" fill="none" stroke="currentColor" strokeWidth="1.5">
                            <rect x="3" y="5" width="8" height="8" />
                            <path d="M11 9l-3-2v4z" fill="currentColor" stroke="none" />
                        </svg>
                    }
                    active={toolbarState.chartShiftEnabled}
                    onClick={() => dispatchToolbar({ type: 'TOGGLE_CHART_SHIFT' })}
                    title="Chart Shift (Right Margin)"
                />
                <Divider />
                <ToolButton
                    icon={<LayoutList size={16} />}
                    onClick={onToggleDOM}
                    title="Depth of Market (DOM)"
                    label="DOM"
                />
            </ToolbarGroup>

            <Divider />

            {/* Section 4: Drawing Tools */}
            <ToolbarGroup>
                <ToolButton
                    icon={<MousePointer2 size={15} />}
                    active={toolbarState.activeTool === null}
                    onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'cursor' } })}
                    title="Cursor (Esc)"
                />
                <ToolButton
                    icon={<div className="w-[1.5px] h-4 bg-current mx-auto" />}
                    active={toolbarState.activeTool === 'vline'}
                    onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'vline' } })}
                    title="Vertical Line"
                />
                <ToolButton
                    icon={<Minus size={15} />}
                    active={toolbarState.activeTool === 'hline'}
                    onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'hline' } })}
                    title="Horizontal Line (H)"
                />
                <ToolButton
                    icon={<Minus size={15} className="rotate-45" />}
                    active={toolbarState.activeTool === 'trendline'}
                    onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'trendline' } })}
                    title="Trendline (T)"
                />
                <ToolButton
                    icon={
                        <svg width="15" height="15" viewBox="0 0 18 18" fill="none" stroke="currentColor" strokeWidth="2">
                            <path d="M3 12L12 3M6 15L15 6" />
                        </svg>
                    }
                    active={toolbarState.activeTool === 'channel'}
                    onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'channel' } })}
                    title="Equidistant Channel"
                />
                <ToolButton
                    icon={
                        <svg width="15" height="15" viewBox="0 0 18 18" fill="none" stroke="currentColor" strokeWidth="1.5">
                            <path d="M2 16L16 2M2 13h14M2 10h14M2 7h14M2 4h14" />
                        </svg>
                    }
                    active={toolbarState.activeTool === 'fibonacci'}
                    onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'fibonacci' } })}
                    title="Fibonacci Retracement"
                />
                <ToolButton
                    icon={<Type size={15} />}
                    active={toolbarState.activeTool === 'text'}
                    onClick={() => dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'text' } })}
                    title="Text (X)"
                />
                <ShapesDropdown
                    activeTool={toolbarState.activeTool}
                    dispatchCommand={dispatchCommand}
                />
                <Divider />
                <DrawingsDropdown />
            </ToolbarGroup>

            <div className="flex-1"></div>

            {/* Section 5: Timeframes */}
            <div className="flex items-center gap-0.5 bg-[#181818] p-0.5 rounded-md border border-zinc-800">
                {(['M1', 'M5', 'M15', 'M30', 'H1', 'H4', 'D1', 'W1', 'MN'] as string[]).map((tf) => (
                    <button
                        key={tf}
                        onClick={() => {
                            onTimeframeChange(tf.toLowerCase() as Timeframe);
                            dispatchCommand({ type: 'SET_TIMEFRAME', payload: { timeframe: tf.toLowerCase() } });
                        }}
                        className={`px-2.5 py-1 text-[10px] font-bold rounded-[3px] transition-all ${timeframe.toUpperCase() === tf
                            ? 'bg-blue-600 text-white shadow-sm'
                            : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800'
                            }`}
                    >
                        {tf}
                    </button>
                ))}
            </div>
        </div>
    );
};

// --- Utility Components ---

const ToolbarGroup = ({ children }: { children: React.ReactNode }) => (
    <div className="flex items-center gap-1">
        {children}
    </div>
);

const Divider = () => (
    <div className="w-[1px] h-5 bg-zinc-700/50 mx-1"></div>
);

interface ToolButtonProps {
    icon: React.ReactNode;
    label?: string;
    title?: string;
    active?: boolean;
    activeColor?: 'blue' | 'emerald';
    onClick?: () => void;
}

const ToolButton = ({ icon, label, title, active, activeColor = 'blue', onClick }: ToolButtonProps) => (
    <button
        onClick={onClick}
        title={title || label}
        className={`flex items-center gap-1.5 px-1.5 py-1.5 rounded transition-all
            ${active
                ? activeColor === 'blue'
                    ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20 shadow-inner'
                    : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shadow-inner'
                : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#333] border border-transparent'
            }
        `}
    >
        {icon}
        {label && <span className="text-[11px] font-medium hidden lg:inline-block">{label}</span>}
    </button>
);

const ShapesIcons = {
    ThumbsUp: () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3" /></svg>,
    ThumbsDown: () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M10 15v4a3 3 0 0 0 3 3l4-9V2H5.72a2 2 0 0 0-2 1.7l-1.38 9a2 2 0 0 0 2 2.3zm7-13h3a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2h-3" /></svg>,
    ArrowUp: () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 19V5M5 12l7-7 7 7" /></svg>,
    ArrowDown: () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 5v14M5 12l7 7 7-7" /></svg>,
    Stop: () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="2"><circle cx="12" cy="12" r="10" /><line x1="4.93" y1="4.93" x2="19.07" y2="19.07" /></svg>,
    Check: () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#22c55e" strokeWidth="2"><polyline points="20 6 9 17 4 12" /></svg>,
}

const ShapesDropdown = ({ activeTool, dispatchCommand }: { activeTool: string | null, dispatchCommand: any }) => {
    const [isOpen, setIsOpen] = React.useState(false);

    const handleSelect = (subtype: string) => {
        dispatchCommand({ type: 'SELECT_TOOL', payload: { tool: 'shapes', subtype } });
        setIsOpen(false);
    };

    return (
        <div className="relative">
            <ToolButton
                icon={
                    <div className="flex items-center gap-0.5">
                        <svg width="14" height="14" viewBox="0 0 18 18" fill="currentColor">
                            <path d="M3 3h6v6h-6zM10 3l5 3-5 3zM12.5 10a2.5 2.5 0 100 5 2.5 2.5 0 000-5z" />
                        </svg>
                        <ChevronDown size={10} className="text-zinc-500" />
                    </div>
                }
                active={activeTool === 'shapes'}
                onClick={() => setIsOpen(!isOpen)}
                title="Draw Shapes & Objects"
            />

            {isOpen && (
                <div className="absolute top-full left-0 mt-1 w-44 bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl z-[100] py-1 overflow-hidden animate-in fade-in slide-in-from-top-2 duration-150">
                    {[
                        { id: 'thumbs_up', label: 'Thumbs Up', icon: <ShapesIcons.ThumbsUp /> },
                        { id: 'thumbs_down', label: 'Thumbs Down', icon: <ShapesIcons.ThumbsDown /> },
                        { id: 'arrow_up', label: 'Arrow Up', icon: <ShapesIcons.ArrowUp /> },
                        { id: 'arrow_down', label: 'Arrow Down', icon: <ShapesIcons.ArrowDown /> },
                        { id: 'stop', label: 'Stop Sign', icon: <ShapesIcons.Stop /> },
                        { id: 'check', label: 'Check Sign', icon: <ShapesIcons.Check /> },
                        { id: 'buy', label: 'Buy Call', icon: <ShapesIcons.ArrowUp /> },
                        { id: 'sell', label: 'Sell Put', icon: <ShapesIcons.ArrowDown /> },
                        { id: 'arrow', label: 'Arrow', icon: <Play size={10} className="rotate-90" /> },
                    ].map(item => (
                        <div
                            key={item.id}
                            className="flex items-center gap-3 px-3 py-2 text-[11px] text-zinc-300 hover:bg-[#252525] hover:text-white cursor-pointer transition-colors"
                            onClick={() => handleSelect(item.id)}
                        >
                            <span className="w-4 flex justify-center">{item.icon}</span>
                            <span>{item.label}</span>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};
