import React, { useState } from 'react';
import {
    TrendingUp, BarChart2, LineChart, Activity, Clock,
    Grid3x3, Volume2, Settings, Image, Printer, ChevronRight,
    Save, FolderOpen, Layers
} from 'lucide-react';
import type { ChartType, Timeframe } from './TradingChart';
import { indicatorManager } from '../services/indicatorManager';
import { useTemplates } from '../store/useTemplateStore';
import { ChartTemplateDialog } from './ChartTemplateDialog';
import type { ChartTemplate } from '../store/useTemplateStore';

interface ChartContextMenuProps {
    visible: boolean;
    x: number;
    y: number;
    currentChartType: ChartType;
    currentTimeframe: Timeframe;
    onChangeChartType: (type: ChartType) => void;
    onChangeTimeframe: (tf: Timeframe) => void;
    onToggleOneClickTrading: () => void;
    onToggleGrid: () => void;
    onToggleVolumes: () => void;
    onOpenProperties: () => void;
    onSaveAsPicture: () => void;
    onPrintChart: () => void;
    onClose: () => void;
    oneClickTrading?: boolean;
    showGrid?: boolean;
    showVolumes?: boolean;
    onLoadTemplate?: (template: ChartTemplate) => void;
    currentIndicators?: any[];
    currentDrawingTools?: any[];
}

export const ChartContextMenu: React.FC<ChartContextMenuProps> = ({
    visible,
    x,
    y,
    currentChartType,
    currentTimeframe,
    onChangeChartType,
    onChangeTimeframe,
    onToggleOneClickTrading,
    onToggleGrid,
    onToggleVolumes,
    onOpenProperties,
    onSaveAsPicture,
    onPrintChart,
    onClose,
    oneClickTrading = false,
    showGrid = true,
    showVolumes = true,
    onLoadTemplate,
    currentIndicators = [],
    currentDrawingTools = []
}) => {
    const templates = useTemplates();
    const [templateDialogOpen, setTemplateDialogOpen] = useState(false);
    const [templateDialogMode, setTemplateDialogMode] = useState<'save' | 'load'>('load');

    if (!visible) return null;

    const timeframes: Timeframe[] = ['M1', 'M5', 'M15', 'M30', 'H1', 'H4', 'D1', 'W1', 'MN'];

    const chartTypes: { value: ChartType; label: string; icon: React.ReactNode }[] = [
        { value: 'candlestick', label: 'Candlestick', icon: <BarChart2 size={14} /> },
        { value: 'bar', label: 'Bar Chart', icon: <BarChart2 size={14} /> },
        { value: 'line', label: 'Line Chart', icon: <LineChart size={14} /> },
        { value: 'area', label: 'Area Chart', icon: <Activity size={14} /> },
        { value: 'heikinAshi', label: 'Heikin-Ashi', icon: <TrendingUp size={14} /> }
    ];

    const indicators = [
        { name: 'Moving Average', type: 'sma' },
        { name: 'Exponential MA', type: 'ema' },
        { name: 'Bollinger Bands', type: 'bb' },
        { name: 'RSI', type: 'rsi' },
        { name: 'MACD', type: 'macd' },
        { name: 'Stochastic', type: 'stochastic' },
        { name: 'CCI', type: 'cci' },
        { name: 'Momentum', type: 'momentum' }
    ];

    const drawingTools = [
        { name: 'Trendline', tool: 'trendline' },
        { name: 'Horizontal Line', tool: 'hline' },
        { name: 'Vertical Line', tool: 'vline' },
        { name: 'Fibonacci Retracement', tool: 'fibonacci' },
        { name: 'Trend Channel', tool: 'channel' },
        { name: 'Text/Annotation', tool: 'text' },
        { name: 'Shapes', tool: 'shapes' }
    ];

    const handleAddIndicator = (type: string) => {
        const id = `${type}-${Date.now()}`;
        indicatorManager.addIndicator({
            id,
            name: type.toUpperCase(),
            type,
            parameters: {},
            visible: true
        });
        onClose();
    };

    const handleSelectDrawingTool = async (tool: string) => {
        try {
            const commandBusModule = await import('../services/commandBus');
            const commandBus = commandBusModule.commandBus;
            if (commandBus && typeof commandBus.subscribe === 'function') {
                // Trigger tool selection - the command bus may not have a publish method
                // Use a custom event instead
                window.dispatchEvent(new CustomEvent('drawing:select-tool', {
                    detail: { tool }
                }));
            }
            onClose();
        } catch (error) {
            console.log('Command bus not available');
        }
    };

    const handleOpenSaveTemplateDialog = () => {
        setTemplateDialogMode('save');
        setTemplateDialogOpen(true);
        onClose();
    };

    const handleOpenLoadTemplateDialog = () => {
        setTemplateDialogMode('load');
        setTemplateDialogOpen(true);
        onClose();
    };

    const handleTemplateLoad = (template: ChartTemplate) => {
        if (onLoadTemplate) {
            onLoadTemplate(template);
        }
        setTemplateDialogOpen(false);
    };

    const currentConfig = {
        chartType: currentChartType,
        timeframe: currentTimeframe,
        showGrid,
        showVolumes,
        indicators: currentIndicators,
        drawingTools: currentDrawingTools,
        colors: {
            background: '#000000',
            grid: '#2b2b2b',
            candle: { up: '#26a69a', down: '#ef5350', border: '#378658' },
            volume: { up: '#26a69a', down: '#ef5350' },
        },
    };

    return (
        <div
            className="fixed z-[1000] bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl py-1 w-56 animate-in fade-in zoom-in-95 duration-100"
            style={{ left: x, top: y }}
            onMouseLeave={onClose}
            onClick={(e) => e.stopPropagation()}
        >
            {/* Timeframe Submenu */}
            <div className="group/submenu relative">
                <ContextMenuItem
                    icon={<Clock size={14} />}
                    label="Timeframe"
                    hasSubmenu
                />
                <div className="absolute left-full top-0 ml-1 hidden group-hover/submenu:block bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl py-1 w-32 animate-in fade-in slide-in-from-left-2 duration-100">
                    {timeframes.map((tf) => (
                        <ContextMenuItem
                            key={tf}
                            label={tf}
                            onClick={() => {
                                onChangeTimeframe(tf);
                                onClose();
                            }}
                            active={currentTimeframe === tf}
                            icon={null}
                        />
                    ))}
                </div>
            </div>

            {/* Chart Type Submenu */}
            <div className="group/submenu relative">
                <ContextMenuItem
                    icon={<BarChart2 size={14} />}
                    label="Chart Type"
                    hasSubmenu
                />
                <div className="absolute left-full top-0 ml-1 hidden group-hover/submenu:block bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl py-1 w-48 animate-in fade-in slide-in-from-left-2 duration-100">
                    {chartTypes.map((ct) => (
                        <ContextMenuItem
                            key={ct.value}
                            icon={ct.icon}
                            label={ct.label}
                            onClick={() => {
                                onChangeChartType(ct.value);
                                onClose();
                            }}
                            active={currentChartType === ct.value}
                        />
                    ))}
                </div>
            </div>

            {/* Indicators Submenu */}
            <div className="group/submenu relative">
                <ContextMenuItem
                    icon={<TrendingUp size={14} />}
                    label="Indicators"
                    hasSubmenu
                />
                <div className="absolute left-full top-0 ml-1 hidden group-hover/submenu:block bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl py-1 w-48 animate-in fade-in slide-in-from-left-2 duration-100 max-h-80 overflow-y-auto">
                    {indicators.map((ind) => (
                        <ContextMenuItem
                            key={ind.type}
                            label={ind.name}
                            onClick={() => handleAddIndicator(ind.type)}
                            icon={null}
                        />
                    ))}
                </div>
            </div>

            {/* Drawing Tools Submenu */}
            <div className="group/submenu relative">
                <ContextMenuItem
                    icon={<Activity size={14} />}
                    label="Drawing Tools"
                    hasSubmenu
                />
                <div className="absolute left-full top-0 ml-1 hidden group-hover/submenu:block bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl py-1 w-48 animate-in fade-in slide-in-from-left-2 duration-100">
                    {drawingTools.map((tool) => (
                        <ContextMenuItem
                            key={tool.tool}
                            label={tool.name}
                            onClick={() => handleSelectDrawingTool(tool.tool)}
                            icon={null}
                        />
                    ))}
                </div>
            </div>

            {/* Templates Submenu */}
            <div className="group/submenu relative">
                <ContextMenuItem
                    icon={<Layers size={14} />}
                    label="Templates"
                    hasSubmenu
                />
                <div className="absolute left-full top-0 ml-1 hidden group-hover/submenu:block bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl py-1 w-56 animate-in fade-in slide-in-from-left-2 duration-100 max-h-96 overflow-y-auto">
                    {/* Quick actions */}
                    <ContextMenuItem
                        icon={<Save size={14} />}
                        label="Save Template..."
                        onClick={handleOpenSaveTemplateDialog}
                    />
                    <ContextMenuItem
                        icon={<FolderOpen size={14} />}
                        label="Manage Templates..."
                        onClick={handleOpenLoadTemplateDialog}
                    />

                    {templates.length > 0 && (
                        <>
                            <div className="h-[1px] bg-zinc-800 my-1 mx-2" />
                            <div className="px-3 py-1 text-[10px] text-zinc-600 uppercase tracking-wide">
                                Quick Load
                            </div>
                            {templates.slice(0, 8).map((template) => (
                                <ContextMenuItem
                                    key={template.id}
                                    label={template.name}
                                    onClick={() => {
                                        handleTemplateLoad(template);
                                        onClose();
                                    }}
                                    icon={null}
                                />
                            ))}
                        </>
                    )}
                </div>
            </div>

            <div className="h-[1px] bg-zinc-800 my-1 mx-2" />

            {/* Toggle Options */}
            <ContextMenuItem
                icon={<Activity size={14} />}
                label="One-Click Trading"
                onClick={() => {
                    onToggleOneClickTrading();
                    onClose();
                }}
                active={oneClickTrading}
            />
            <ContextMenuItem
                icon={<Grid3x3 size={14} />}
                label="Show Grid"
                onClick={() => {
                    onToggleGrid();
                    onClose();
                }}
                active={showGrid}
            />
            <ContextMenuItem
                icon={<Volume2 size={14} />}
                label="Show Volumes"
                onClick={() => {
                    onToggleVolumes();
                    onClose();
                }}
                active={showVolumes}
            />

            <div className="h-[1px] bg-zinc-800 my-1 mx-2" />

            {/* Actions */}
            <ContextMenuItem
                icon={<Settings size={14} />}
                label="Properties"
                onClick={() => {
                    onOpenProperties();
                    onClose();
                }}
            />
            <ContextMenuItem
                icon={<Image size={14} />}
                label="Save As Picture"
                onClick={() => {
                    onSaveAsPicture();
                    onClose();
                }}
            />
            <ContextMenuItem
                icon={<Printer size={14} />}
                label="Print Chart"
                onClick={() => {
                    onPrintChart();
                    onClose();
                }}
            />

            {/* Template Dialog */}
            <ChartTemplateDialog
                isOpen={templateDialogOpen}
                onClose={() => setTemplateDialogOpen(false)}
                mode={templateDialogMode}
                currentConfig={currentConfig}
                onLoadTemplate={handleTemplateLoad}
            />
        </div>
    );
};

interface ContextMenuItemProps {
    icon: React.ReactNode;
    label: string;
    onClick?: () => void;
    hasSubmenu?: boolean;
    active?: boolean;
}

const ContextMenuItem: React.FC<ContextMenuItemProps> = ({
    icon,
    label,
    onClick,
    hasSubmenu,
    active
}) => (
    <div
        className={`flex items-center justify-between px-3 py-2 text-[12px] text-zinc-300 hover:bg-[#2d2d2d] hover:text-white cursor-pointer transition-colors ${
            active ? 'bg-[#2d2d2d] text-white' : ''
        }`}
        onClick={(e) => {
            e.stopPropagation();
            if (onClick) onClick();
        }}
    >
        <div className="flex items-center gap-3">
            {icon && <span className="text-zinc-500">{icon}</span>}
            <span className={!icon ? 'ml-7' : ''}>{label}</span>
        </div>
        {hasSubmenu && (
            <ChevronRight size={14} className="text-zinc-500" />
        )}
        {active && !hasSubmenu && (
            <div className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
        )}
    </div>
);
