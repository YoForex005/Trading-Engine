/**
 * File Menu Component - MT5 Style
 * Complete File menu dropdown with exact MT5 behavior
 */

import React, { useState } from 'react';
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
    Image
} from 'lucide-react';
import { SaveWorkspaceDialog } from '../dialogs/SaveWorkspaceDialog';
import { PrintSetupDialog } from '../dialogs/PrintSetupDialog';
import { PrintPreviewDialog } from '../dialogs/PrintPreviewDialog';
import { ExitConfirmDialog } from '../dialogs/ExitConfirmDialog';
import { SavePictureDialog } from '../dialogs/SavePictureDialog';
import { useAppStore } from '../../store/useAppStore';
import { chartPrinter, type PrintPreferences, type ChartPrintData } from '../../services/chartPrinter';

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
    const [showSaveDialog, setShowSaveDialog] = useState(false);
    const [showSavePictureDialog, setShowSavePictureDialog] = useState(false);
    const [showPrintSetupDialog, setShowPrintSetupDialog] = useState(false);
    const [showPrintPreviewDialog, setShowPrintPreviewDialog] = useState(false);
    const [showExitDialog, setShowExitDialog] = useState(false);

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
            onExit?.();
            window.close();
        }
    };

    return (
        <>
            <div className="min-w-[240px] bg-[#1e1e1e] border border-zinc-700 rounded-md shadow-xl py-1">
                {/* Save */}
                <button
                    onClick={handleSave}
                    className="w-full flex items-center gap-3 px-2 py-1.5 text-[12px] rounded-sm text-left text-zinc-300 hover:bg-[#2a2e39] hover:text-white cursor-default group"
                    title="Save workspace configuration"
                >
                    <div className="w-4 flex items-center justify-center text-zinc-400 group-hover:text-zinc-200">
                        <Save size={14} />
                    </div>
                    <span className="flex-1">Save</span>
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
                    <span className="flex-1">Save As Picture</span>
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
                    <span className="flex-1">Open Data Folder</span>
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
                    <span className="flex-1">Print</span>
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
                    <span className="flex-1">Print Preview</span>
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
                    <span className="flex-1">Print Setup</span>
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
                    <span className="flex-1">Open an Account</span>
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
                    <span className="flex-1">Login to Trade Account</span>
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
                    <span className="flex-1">Login to Web Trader</span>
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
                    <span className="flex-1">Login to MQL5.community</span>
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
                    <span className="flex-1">Exit</span>
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
                    onSave={() => {
                        setShowExitDialog(false);
                        onSave?.();
                        onExit?.();
                        window.close();
                    }}
                    onDontSave={() => {
                        setShowExitDialog(false);
                        onExit?.();
                        window.close();
                    }}
                    onCancel={() => setShowExitDialog(false)}
                />
            )}
        </>
    );
};
