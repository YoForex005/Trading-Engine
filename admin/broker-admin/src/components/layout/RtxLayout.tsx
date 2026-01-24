import React from 'react';
import TopToolbar from './TopToolbar';
import Navigator from './Navigator';
import BottomToolbox from './BottomToolbox';
import MarketWatch from '../dashboard/MarketWatch';

interface RtxLayoutProps {
    children: React.ReactNode;
    onNavigate?: (viewId: string) => void;
}

export default function RtxLayout({ children, onNavigate }: RtxLayoutProps) {
    return (
        <div className="flex flex-col h-screen w-screen bg-charcoal-950 text-xs text-gray-200 overflow-hidden font-sans">
            {/* Top Section: Toolbar */}
            <TopToolbar />

            {/* Middle Section: Navigator + MarketWatch + Main Config */}
            <div className="flex flex-1 overflow-hidden">
                <Navigator onNavigate={onNavigate} />

                {/* Market Watch Panel */}
                <MarketWatch />

                {/* Main Workspace */}
                <main className="flex-1 bg-charcoal-900 relative overflow-auto flex flex-col">
                    {/* Optional Tab Bar for Main Window could go here */}
                    <div className="flex-1 overflow-auto">
                        {children}
                    </div>
                </main>
            </div>

            {/* Bottom Section: Toolbox */}
            <BottomToolbox />

            {/* Status Bar (Very bottom strip) */}
            <div className="h-5 bg-charcoal-950 border-t border-charcoal-border flex items-center px-2 justify-between select-none">
                <div className="flex items-center gap-4">
                    <span>Press F1 for Help</span>
                </div>
                <div className="flex items-center gap-4">
                    <span>0/0 kb</span>
                    <span>12ms</span>
                </div>
            </div>
        </div>
    );
}
