/**
 * File Menu Component - MT5 Style
 * Complete File menu dropdown with exact MT5 behavior
 */

import React, { useState, useEffect } from 'react';
import {
    Save,
    FileText,
    FolderOpen,
    Printer,
    Settings,
    UserPlus,
    LogIn,
    Globe,
    GraduationCap,
    LogOut,
    Image,
    PlusSquare,
    ChevronRight,
    Loader2
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { SaveWorkspaceDialog } from '../dialogs/SaveWorkspaceDialog';
import { PrintSetupDialog } from '../dialogs/PrintSetupDialog';
import { PrintPreviewDialog } from '../dialogs/PrintPreviewDialog';
import { ExitConfirmDialog } from '../dialogs/ExitConfirmDialog';
import { SavePictureDialog } from '../dialogs/SavePictureDialog';
import { useAppStore } from '../../store/useAppStore';
import { chartPrinter, type PrintPreferences, type ChartPrintData } from '../../services/chartPrinter';

// Symbol data type from API
interface SymbolData {
    symbol: string;
    name: string;
    category: string;
}

// Menu item interface for building nested structure
interface NewChartMenuItem {
    label: string;
    symbol?: string;
    children?: NewChartMenuItem[];
}

// Single menu item renderer with recursive submenu support
const NewChartMenuItemRenderer: React.FC<{ item: NewChartMenuItem; depth: number }> = ({ item, depth }) => {
    const hasChildren = item.children && item.children.length > 0;

    const handleClick = () => {
        if (item.symbol) {
            window.dispatchEvent(new CustomEvent('open-chart', { detail: { symbol: item.symbol } }));
        }
    };

    return (
        <div className="relative group/sub px-1 [&:hover>div.submenu]:block">
            <button
                onClick={handleClick}
                className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
            >
                <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                    {/* Empty space for alignment */}
                </div>
                <span className="flex-1">{item.label}</span>
                {hasChildren && <ChevronRight size={12} className="text-zinc-500" />}
            </button>
            {hasChildren && (
                <div className="submenu absolute left-full top-0 -ml-1 hidden min-w-[180px] bg-[#1e1e1e] border border-zinc-700 rounded-md shadow-xl py-1 z-[102]">
                    {item.children!.map((child, idx) => (
                        <NewChartMenuItemRenderer key={idx} item={child} depth={depth + 1} />
                    ))}
                </div>
            )}
        </div>
    );
};

export interface FileMenuProps {
    onSave?: () => void;
    onSaveAsPicture?: () => void;
    onOpenDataFolder?: () => void;
    onPrint?: () => void;
    onPrintPreview?: () => void;
    onPrintSetup?: () => void;
    onExit?: () => void;
    hasUnsavedChanges?: boolean;
}

export const FileMenu: React.FC<FileMenuProps> = ({
    onSave,
    onSaveAsPicture,
    onOpenDataFolder,
    onPrint,
    onPrintPreview,
    onPrintSetup,
    onExit,
    hasUnsavedChanges = false
}) => {
    const { t } = useTranslation();
    const [showSaveDialog, setShowSaveDialog] = useState(false);
    const [showSavePictureDialog, setShowSavePictureDialog] = useState(false);
    const [showPrintSetupDialog, setShowPrintSetupDialog] = useState(false);
    const [showPrintPreviewDialog, setShowPrintPreviewDialog] = useState(false);
    const [showExitDialog, setShowExitDialog] = useState(false);

    // Dynamic symbol loading for New Chart menu
    const [availableSymbols, setAvailableSymbols] = useState<SymbolData[]>([]);
    const [isLoadingSymbols, setIsLoadingSymbols] = useState(false);

    // Fetch symbols on mount
    useEffect(() => {
        const fetchSymbols = async () => {
            try {
                const { default: api } = await import('../../services/api');
                setIsLoadingSymbols(true);
                const symbols = await api.market.getAvailableSymbols();
                if (symbols && symbols.length > 0) {
                    setAvailableSymbols(symbols);
                }
            } catch (error) {
                console.error('Failed to load menu symbols:', error);
            } finally {
                setIsLoadingSymbols(false);
            }
        };

        fetchSymbols();
    }, []);

    // Build nested menu structure from flat symbol list
    const buildNestedMenu = (symbols: SymbolData[]): NewChartMenuItem[] => {
        const root: NewChartMenuItem[] = [];
        const groups: Record<string, SymbolData[]> = {};

        symbols.forEach(sym => {
            // If category is empty, it's a root item
            if (!sym.category || sym.category.trim() === '') {
                root.push({
                    label: sym.name,
                    symbol: sym.symbol
                });
                return;
            }

            const parts = sym.category.split('.');
            const topLevel = parts[0];
            if (!groups[topLevel]) groups[topLevel] = [];

            if (parts.length > 1) {
                groups[topLevel].push({ ...sym, category: parts.slice(1).join('.') });
            } else {
                groups[topLevel].push({ ...sym, category: '' });
            }
        });

        // Add grouped items (folders) after root items
        Object.keys(groups).sort().forEach(key => {
            const groupSymbols = groups[key];
            root.push({
                label: key,
                children: buildNestedMenu(groupSymbols)
            });
        });

        return root;
    };

    const newChartMenuItems = buildNestedMenu(availableSymbols);

    // Get current symbol and timeframe from store
    const selectedSymbol = useAppStore(state => state.selectedSymbol);
    const timeframe = useAppStore(state => state.timeframe);

    // Print state
    const [printPreferences, setPrintPreferences] = useState<PrintPreferences | null>(null);
    const [chartPrintData, setChartPrintData] = useState<ChartPrintData | null>(null);

    const handleSave = () => {
        if (hasUnsavedChanges) {
            setShowSaveDialog(true);
        } else {
            onSave?.();
        }
    };

    const handleSaveAsPicture = () => {
        // Get the active chart canvas from the global window object
        const canvas = (window as any).__activeChartCanvas as HTMLCanvasElement | null;
        const container = (window as any).__activeChartContainer as HTMLElement | null;

        if (!canvas && !container) {
            console.warn('[FileMenu] No active chart canvas or container found for export');
            alert('No chart available to export. Please open a chart first.');
            return;
        }

        setShowSavePictureDialog(true);
        onSaveAsPicture?.();
    };

    const handleOpenDataFolder = () => {
        onOpenDataFolder?.();
        // Open the data folder using electron or browser
        if (window.electron) {
            window.electron.shell.openPath('./data');
        }
    };

    const handlePrint = async () => {
        // Generate chart data for printing
        const data = generateChartPrintData();

        if (onPrint) {
            onPrint();
        } else {
            // Use the chart printer service
            const prefs = chartPrinter.getPreferences();
            await chartPrinter.printChart(data, prefs);
        }
    };

    const handlePrintPreview = () => {
        // Generate chart data for preview
        const data = generateChartPrintData();
        setChartPrintData(data);
        setShowPrintPreviewDialog(true);
    };

    const handlePrintSetup = () => {
        setShowPrintSetupDialog(true);
    };

    const generateChartPrintData = (): ChartPrintData => {
        // Try to capture the chart canvas
        const chartContainer = document.querySelector('.tv-lightweight-charts');
        let chartCanvas: HTMLCanvasElement | undefined;

        if (chartContainer) {
            const canvas = chartContainer.querySelector('canvas');
            if (canvas) {
                // Create a copy of the canvas
                const copiedCanvas = document.createElement('canvas');
                copiedCanvas.width = canvas.width;
                copiedCanvas.height = canvas.height;
                const ctx = copiedCanvas.getContext('2d');
                if (ctx) {
                    ctx.drawImage(canvas, 0, 0);
                    chartCanvas = copiedCanvas;
                }
            }
        }

        return {
            symbol: selectedSymbol || 'BTCUSD',
            timeframe: timeframe || '1m',
            dateRange: {
                from: new Date(Date.now() - 24 * 60 * 60 * 1000),
                to: new Date(),
            },
            chartCanvas,
            indicators: [],
            drawings: [],
        };
    };

    const handleExit = () => {
        if (hasUnsavedChanges) {
            setShowExitDialog(true);
        } else {
            performExit();
        }
    };

    const performExit = () => {
        onExit?.();

        // Clear authentication state
        try {
            useAppStore.getState().clearAuth?.();
        } catch (e) {
            console.log('No auth to clear');
        }

        // Try Electron methods first (if running in Electron)
        if ((window as any).electron) {
            try {
                (window as any).electron.app?.quit?.();
                (window as any).electron.ipcRenderer?.send?.('app-quit');
                return;
            } catch (e) {
                console.log('Electron quit failed, trying browser methods');
            }
        }

        // Try window.close() for popup windows
        try {
            window.close();
        } catch (e) {
            console.log('window.close() not available');
        }

        // For main browser window - show confirmation and redirect to blank
        if (!window.closed) {
            const confirmed = window.confirm('Close the trading terminal?');
            if (confirmed) {
                // Clear all session data
                sessionStorage.clear();
                localStorage.removeItem('authToken');

                // Redirect to about:blank or a logout page
                window.location.href = 'about:blank';
            }
        }
    };



    return (
        <>
            <div className="min-w-[240px] bg-[#1e1e1e] border border-zinc-700 rounded-md shadow-xl py-1">
                {/* New Chart - Dynamic submenu */}
                <div className="relative group/newchart px-1 [&:hover>div.submenu]:block">
                    <button
                        className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                        title="Open a new chart for a symbol"
                    >
                        <div className="w-4 flex items-center justify-center text-emerald-500">
                            <PlusSquare size={14} />
                        </div>
                        <span className="flex-1">{t('menu.file.newChart')}</span>
                        <ChevronRight size={12} className="text-zinc-500" />
                    </button>
                    <div className="submenu absolute left-full top-0 -ml-1 hidden min-w-[180px] bg-[#1e1e1e] border border-zinc-700 rounded-md shadow-xl py-1 z-[101]">
                        {isLoadingSymbols ? (
                            <div className="px-3 py-2 text-zinc-500 text-[12px] flex items-center gap-2">
                                <Loader2 size={14} className="animate-spin" />
                                Loading symbols...
                            </div>
                        ) : newChartMenuItems.length > 0 ? (
                            newChartMenuItems.map((item, idx) => (
                                <NewChartMenuItemRenderer key={idx} item={item} depth={0} />
                            ))
                        ) : (
                            <>
                                {/* Fallback if no API symbols */}
                                {['EURUSD', 'GBPUSD', 'USDJPY', 'USDCHF', 'USDCAD', 'AUDUSD'].map(sym => (
                                    <button
                                        key={sym}
                                        onClick={() => window.dispatchEvent(new CustomEvent('open-chart', { detail: { symbol: sym } }))}
                                        className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default"
                                    >
                                        <div className="w-4"></div>
                                        <span>{sym}</span>
                                    </button>
                                ))}
                            </>
                        )}
                    </div>
                </div>

                {/* Open Deleted */}
                <button
                    disabled
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-600 cursor-not-allowed"
                    title="Restore deleted charts"
                >
                    <div className="w-4 flex items-center justify-center">
                    </div>
                    <span className="flex-1">{t('menu.file.openDeleted')}</span>
                </button>

                {/* Profiles submenu */}
                <div className="relative group/profiles px-1 [&:hover>div.submenu]:block">
                    <button
                        className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default"
                        title="Load profile configurations"
                    >
                        <div className="w-4 flex items-center justify-center">
                        </div>
                        <span className="flex-1">{t('menu.file.profiles')}</span>
                        <ChevronRight size={12} className="text-zinc-500" />
                    </button>
                    <div className="submenu absolute left-full top-0 -ml-1 hidden min-w-[140px] bg-[#1e1e1e] border border-zinc-700 rounded-md shadow-xl py-1 z-[101]">
                        {['Default', 'Euro', 'Market'].map(profile => (
                            <button
                                key={profile}
                                className="w-full flex items-center gap-3 px-3 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default"
                            >
                                {profile}
                            </button>
                        ))}
                    </div>
                </div>

                {/* Close */}
                <button
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                    title="Close current chart"
                >
                    <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                    </div>
                    <span className="flex-1">{t('menu.file.close')}</span>
                    <span className="text-[10px] text-zinc-500 font-mono tracking-tighter">Ctrl+F4</span>
                </button>

                <div className="h-[1px] bg-zinc-700/50 my-1 mx-2"></div>

                {/* Save */}
                <button
                    onClick={handleSave}
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                    title="Save workspace configuration"
                >
                    <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                        <Save size={14} />
                    </div>
                    <span className="flex-1">{t('menu.file.save')}</span>
                    <span className="text-[10px] text-zinc-500 font-mono tracking-tighter">Ctrl+S</span>
                </button>

                {/* Save As Picture */}
                <button
                    onClick={handleSaveAsPicture}
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                    title="Export chart as image"
                >
                    <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                        <Image size={14} />
                    </div>
                    <span className="flex-1">{t('menu.file.saveAsPicture')}</span>
                </button>

                <div className="h-[1px] bg-zinc-700/50 my-1 mx-2"></div>

                {/* Open Data Folder */}
                <button
                    onClick={handleOpenDataFolder}
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                    title="Open application data directory"
                >
                    <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                        <FolderOpen size={14} />
                    </div>
                    <span className="flex-1">{t('menu.file.openDataFolder')}</span>
                    <span className="text-[10px] text-zinc-500 font-mono tracking-tighter">Ctrl+Shift+D</span>
                </button>

                <div className="h-[1px] bg-zinc-700/50 my-1 mx-2"></div>

                {/* Print */}
                <button
                    onClick={handlePrint}
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                    title="Print chart"
                >
                    <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                        <Printer size={14} />
                    </div>
                    <span className="flex-1">{t('menu.file.print')}</span>
                    <span className="text-[10px] text-zinc-500 font-mono tracking-tighter">Ctrl+P</span>
                </button>

                {/* Print Preview */}
                <button
                    onClick={handlePrintPreview}
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                    title="Preview before printing"
                >
                    <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                        <FileText size={14} />
                    </div>
                    <span className="flex-1">{t('menu.file.printPreview')}</span>
                </button>

                {/* Print Setup */}
                <button
                    onClick={handlePrintSetup}
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                    title="Configure print settings"
                >
                    <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                        <Settings size={14} />
                    </div>
                    <span className="flex-1">{t('menu.file.printSetup')}</span>
                </button>

                <div className="h-[1px] bg-zinc-700/50 my-1 mx-2"></div>

                {/* Open an Account */}
                <button
                    disabled
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-600 cursor-not-allowed"
                    title="Register new trading account"
                >
                    <div className="w-4 flex items-center justify-center">
                        <UserPlus size={14} className="text-blue-400/50" />
                    </div>
                    <span className="flex-1">{t('menu.file.openAccount')}</span>
                </button>

                {/* Login to Trade Account */}
                <button
                    disabled
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-600 cursor-not-allowed"
                    title="Connect to trading account"
                >
                    <div className="w-4 flex items-center justify-center">
                        <LogIn size={14} className="text-blue-400/50" />
                    </div>
                    <span className="flex-1">{t('menu.file.loginTrade')}</span>
                </button>

                {/* Login to Web Trader */}
                <button
                    disabled
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-600 cursor-not-allowed"
                    title="Access web trading platform"
                >
                    <div className="w-4 flex items-center justify-center">
                        <Globe size={14} />
                    </div>
                    <span className="flex-1">{t('menu.file.loginWeb')}</span>
                </button>

                {/* Login to MQL5.community */}
                <button
                    disabled
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-600 cursor-not-allowed"
                    title="Connect to MQL5 community"
                >
                    <div className="w-4 flex items-center justify-center">
                        <GraduationCap size={14} className="text-blue-500/50" />
                    </div>
                    <span className="flex-1">{t('menu.file.loginMql5')}</span>
                </button>

                <div className="h-[1px] bg-zinc-700/50 my-1 mx-2"></div>

                {/* Exit */}
                <button
                    onClick={handleExit}
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                    title="Close application"
                >
                    <div className="w-4 flex items-center justify-center text-rose-400 group-hover:text-rose-300">
                        <LogOut size={14} />
                    </div>
                    <span className="flex-1">{t('menu.file.exit')}</span>
                </button>
            </div>

            {/* Dialogs */}
            {showSaveDialog && (
                <SaveWorkspaceDialog
                    onConfirm={() => {
                        setShowSaveDialog(false);
                        onSave?.();
                    }}
                    onCancel={() => setShowSaveDialog(false)}
                />
            )}

            {showPrintSetupDialog && (
                <PrintSetupDialog
                    onConfirm={(preferences) => {
                        setShowPrintSetupDialog(false);
                        setPrintPreferences(preferences);
                        console.log('Print preferences saved:', preferences);
                    }}
                    onCancel={() => setShowPrintSetupDialog(false)}
                    accountId="1"
                />
            )}

            {showPrintPreviewDialog && (
                <PrintPreviewDialog
                    onPrint={handlePrint}
                    onClose={() => setShowPrintPreviewDialog(false)}
                    chartData={chartPrintData || undefined}
                    preferences={printPreferences || undefined}
                />
            )}

            {showSavePictureDialog && (
                <SavePictureDialog
                    isOpen={showSavePictureDialog}
                    onClose={() => setShowSavePictureDialog(false)}
                    chartCanvas={(window as any).__activeChartCanvas || null}
                    chartContainer={(window as any).__activeChartContainer || null}
                    symbol={selectedSymbol || 'EURUSD'}
                    timeframe={timeframe || '5m'}
                />
            )}

            {showExitDialog && (
                <ExitConfirmDialog
                    isOpen={showExitDialog}
                    changes={{
                        hasWorkspaceChanges: hasUnsavedChanges,
                        hasOpenTrades: false,
                        hasDraftOrders: false,
                        openTradesCount: 0,
                        draftOrdersCount: 0,
                        workspaceModified: hasUnsavedChanges,
                    }}
                    onConfirm={(saveWorkspace) => {
                        setShowExitDialog(false);
                        if (saveWorkspace) {
                            onSave?.();
                        }
                        performExit();
                    }}
                    onCancel={() => setShowExitDialog(false)}
                />
            )}
        </>
    );
};
