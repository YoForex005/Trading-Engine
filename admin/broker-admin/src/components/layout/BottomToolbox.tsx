import React, { useState } from 'react';
import Journal from './Journal';

const TABS = [
    'Trade', 'Exposure', 'History', 'News', 'Mailbox',
    'Calendar', 'Company', 'Alerts', 'Signals', 'Articles', 'Code Base', 'Experts', 'Journal'
];

export default function BottomToolbox() {
    const [activeTab, setActiveTab] = useState('Journal');

    return (
        <div className="h-64 w-full bg-[#1E2026] border-t border-[#383A42] flex flex-col select-none group">
            {/* Draggable Resizer Handle (Visual only for now) */}
            <div className="h-1 cursor-ns-resize bg-[#2A2A2A] hover:bg-[#F5C542]/50 w-full" />

            {/* Content Area */}
            <div className="flex-1 bg-[#121316] relative overflow-hidden">
                {activeTab === 'Journal' && <Journal />}
                {activeTab === 'History' && (
                    <div className="flex items-center justify-center h-full text-[#444] font-mono text-xs uppercase tracking-widest">
                        History & Orders View
                    </div>
                )}
                {activeTab !== 'Journal' && activeTab !== 'History' && (
                    <div className="flex items-center justify-center h-full text-[#333] font-mono text-xs uppercase tracking-widest">
                        Empty {activeTab} Container
                    </div>
                )}
            </div>

            {/* Bottom Tab Bar - Strictly Native */}
            <div className="flex items-end bg-[#1E2026] border-t border-[#383A42] overflow-x-auto custom-scrollbar">
                {TABS.map(tab => (
                    <button
                        key={tab}
                        onClick={() => setActiveTab(tab)}
                        className={`
                            px-4 py-1 text-[11px] border-r border-[#383A42] font-sans
                            ${activeTab === tab
                                ? 'bg-[#121316] text-[#F5C542] border-t-2 border-t-[#F5C542] relative -top-[1px]'
                                : 'bg-[#1E2026] text-[#888] border-t border-transparent hover:bg-[#25272E] hover:text-[#CCC]'}
                            focus:outline-none transition-none
                        `}
                    >
                        {tab}
                    </button>
                ))}
            </div>
        </div>
    );
}
