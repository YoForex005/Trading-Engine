import React, { useState } from 'react';
import { X, Minus, Square, FileText } from 'lucide-react';

interface PositionWindowProps {
    position: any;
    onClose: () => void;
}

const TABS = ['Details', 'Visualization', 'Ticks', 'Ultency Ticks', 'Ultency Book'];

export default function PositionWindow({ position, onClose }: PositionWindowProps) {
    const [activeTab, setActiveTab] = useState('Details');

    return (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 backdrop-blur-[1px] font-sans select-none text-[11px]">
            {/* Window Frame */}
            <div className="w-[800px] bg-[#1E1E1E] border border-[#444] shadow-2xl flex flex-col text-[#CCC]">

                {/* Title Bar */}
                <div className="h-7 bg-[#2D2D30] flex items-center justify-between px-2 select-none border-b border-[#333]">
                    <div className="flex items-center gap-2">
                        {/* Icon placeholder (Green person or similar) */}
                        <span className="text-[#2ECC71] font-bold">👤</span>
                        <span className="font-bold text-white">
                            Position #{position?.ticket || '5034373'} {position?.type || 'sell'} {position?.vol || '0.02'} {position?.symbol || 'XAUUSD'} {position?.price || '4637.50'}
                        </span>
                    </div>
                    <div className="flex items-center gap-2 text-[#888]">
                        <Minus size={14} className="hover:text-white cursor-pointer" />
                        <Square size={12} className="hover:text-white cursor-pointer" />
                        <X size={14} className="hover:text-red-500 cursor-pointer" onClick={onClose} />
                    </div>
                </div>

                {/* Main Content Area */}
                <div className="flex-1 flex flex-col p-1 bg-[#1E1E1E]">

                    {/* Top Grid (Single Row) */}
                    <div className="border border-[#444] bg-[#252526] mb-2">
                        <table className="w-full border-collapse text-left">
                            <thead className="bg-[#2D2D30] text-[#A0A0A0]">
                                <tr>
                                    {['Ticket', 'Time', 'ID', 'Liquidity provider', 'Type', 'Volume', 'Price', 'Reason', 'Profit'].map(h => (
                                        <th key={h} className="border-r border-[#333] px-2 py-1 font-normal">{h}</th>
                                    ))}
                                </tr>
                            </thead>
                            <tbody>
                                <tr className="text-[#CCC]">
                                    <td className="border-r border-[#333] px-2 py-1 flex items-center gap-1">
                                        <div className={`w-2 h-2 border border-[#666] ${position?.type === 'buy' ? 'bg-blue-500' : 'bg-red-500'}`}></div>
                                        {position?.ticket || '5034373'}
                                    </td>
                                    <td className="border-r border-[#333] px-2 py-1">2026.01.14 11:11:45.240</td>
                                    <td className="border-r border-[#333] px-2 py-1"></td>
                                    <td className="border-r border-[#333] px-2 py-1"></td>
                                    <td className="border-r border-[#333] px-2 py-1 text-blue-400">{position?.type || 'sell'}</td>
                                    <td className="border-r border-[#333] px-2 py-1 font-mono text-right">{parseFloat(position?.vol || '0.02').toFixed(2)}</td>
                                    <td className="border-r border-[#333] px-2 py-1 font-mono text-right">{parseFloat(position?.price || '4637.50').toFixed(2)}</td>
                                    <td className="border-r border-[#333] px-2 py-1 text-right">Client</td>
                                    <td className="border-r border-[#333] px-2 py-1 font-mono text-right text-blue-400">-58.86</td>
                                </tr>
                                {/* Additional rows from screenshot (History/Related) */}
                                <tr className="text-[#888]">
                                    <td className="border-r border-[#333] px-2 py-1 flex items-center gap-1 opacity-70">
                                        <FileText size={10} /> {position?.ticket || '5034373'}
                                    </td>
                                    <td className="border-r border-[#333] px-2 py-1">2026.01.14 11:11:45.240</td>
                                    <td className="border-r border-[#333] px-2 py-1"></td>
                                    <td className="border-r border-[#333] px-2 py-1"></td>
                                    <td className="border-r border-[#333] px-2 py-1">sell</td>
                                    <td className="border-r border-[#333] px-2 py-1 font-mono text-right">0.02 / 0.02</td>
                                    <td className="border-r border-[#333] px-2 py-1 font-mono text-right">market</td>
                                    <td className="border-r border-[#333] px-2 py-1 text-right">Client</td>
                                    <td className="border-r border-[#333] px-2 py-1"></td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    {/* Tabs below grid */}
                    <div className="flex gap-4 px-2 mb-4 text-[#888]">
                        {TABS.map(tab => (
                            <div
                                key={tab}
                                onClick={() => setActiveTab(tab)}
                                className={`cursor-pointer hover:text-[#CCC] ${activeTab === tab ? 'text-[#3B82F6] border-b border-[#3B82F6]' : ''}`}
                            >
                                {tab}
                            </div>
                        ))}
                    </div>

                    {/* Details Form Area */}
                    {activeTab === 'Details' && (
                        <div className="flex gap-4 px-4 pb-4">
                            {/* Left Column Form */}
                            <div className="flex-1 grid grid-cols-[120px_1fr] gap-2 items-center">

                                <label className="text-right text-[#888]">Position: 5034373</label>
                                <div className="flex gap-1">
                                    <select className="bg-[#252526] border border-[#444] h-6 w-24 px-1 text-white"><option>SELL</option></select>
                                    <input type="text" className="bg-[#252526] border border-[#444] h-6 w-20 px-1 text-right text-white font-mono" defaultValue="0.02" />
                                    <select className="flex-1 bg-[#252526] border border-[#444] h-6 px-1 text-white"><option>XAUUSD, Gold vs US Dollar</option></select>
                                    <button className="h-6 w-6 bg-[#333] border border-[#444] text-[#CCC]">...</button>
                                </div>

                                <label className="text-right text-[#888]">Gateway: BUY</label>
                                <div className="flex gap-1">
                                    <select className="bg-[#252526] border border-[#444] h-6 w-24 px-1 text-white"><option>BUY</option></select>
                                    <input type="text" className="bg-[#252526] border border-[#444] h-6 w-20 px-1 text-right text-white font-mono" defaultValue="0.00" />
                                </div>

                                <div className="col-span-2 h-2"></div>

                                <label className="text-right text-[#888]">Reason:</label>
                                <select className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-white"><option>Client</option></select>

                                <label className="text-right text-[#888]">Dealer ID:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-white" defaultValue="0" />

                                <label className="text-right text-[#888]">Expert ID:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-white" />

                                <label className="text-right text-[#888]">External ID:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-white" />

                                <label className="text-right text-[#888]">Comment:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-white" />

                                <label className="text-right text-[#888]">Disabled activations:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-white" />

                                <label className="text-right text-[#888]">Modifications:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-white" />

                            </div>

                            {/* Center/Right Column Data */}
                            <div className="w-[300px] grid grid-cols-[100px_1fr] gap-2 items-center border-l border-[#333] pl-4">

                                <label className="text-right text-[#888]">Gateway price:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-right text-white font-mono" defaultValue="0.00000" />

                                <div className="col-span-2 h-2"></div>

                                <label className="text-right text-[#888]">Open price:</label>
                                <div className="flex gap-1 items-center">
                                    <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 flex-1 text-right text-white font-mono" defaultValue="4637.50000" />
                                    <span className="text-white">📄</span>
                                </div>

                                <label className="text-right text-[#888]">Current price:</label>
                                <div className="flex gap-1 items-center">
                                    <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 flex-1 text-right text-white font-mono" defaultValue="4666.93" />
                                    <span className="text-white">📄</span>
                                </div>

                                <label className="text-right text-[#888]">Stop loss:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-right text-white font-mono" defaultValue="0.00" />

                                <label className="text-right text-[#888]">Take profit:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-right text-white font-mono" defaultValue="0.00" />

                                <label className="text-right text-[#888]">Swap:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-right text-white font-mono" defaultValue="0.00" />

                                <label className="text-right text-[#888]">Profit:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-right text-white font-mono" defaultValue="-58.86" />

                                <label className="text-right text-[#888]">Profit rate:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-right text-white font-mono" defaultValue="1.00000000" />

                                <label className="text-right text-[#888]">Margin rate:</label>
                                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-right text-white font-mono" defaultValue="4637.5000000" />

                            </div>

                            {/* Dates fields specifically positioned */}
                            <div className="absolute left-[380px] top-[265px] flex flex-col gap-2">
                                <div className="flex items-center gap-1">
                                    <label className="text-[#888] text-[10px]">Opened:</label>
                                    <div className="bg-[#252526] border border-[#444] h-6 px-1 w-32 flex items-center justify-between text-[#CCC]">2026.01.14 11:11...</div>
                                    <div className="bg-[#252526] border border-[#444] h-6 w-6 flex items-center justify-center">📅</div>
                                </div>
                                <div className="flex items-center gap-1">
                                    <label className="text-[#888] text-[10px]">Updated:</label>
                                    <div className="bg-[#252526] border border-[#444] h-6 px-1 w-32 flex items-center justify-between text-[#CCC]">2026.01.14 11:11...</div>
                                    <div className="bg-[#252526] border border-[#444] h-6 w-6 flex items-center justify-center">📅</div>
                                </div>
                            </div>

                        </div>
                    )}

                </div>

                {/* Footer Buttons */}
                <div className="h-10 border-t border-[#444] bg-[#2D2D30] flex items-center justify-end px-4 gap-2">
                    <button className="mr-auto px-4 h-6 border border-[#444] bg-[#333] text-[#CCC] hover:bg-[#444]">Report...</button>
                    <button className="px-6 h-6 border border-[#444] bg-[#333] text-[#CCC] hover:bg-[#444]">Update</button>
                    <button className="px-6 h-6 border border-[#444] bg-[#333] text-[#CCC] hover:bg-[#444]" onClick={onClose}>Cancel</button>
                    <button className="px-6 h-6 border border-[#444] bg-[#333] text-[#CCC] hover:bg-[#444]">Help</button>
                </div>

            </div>
        </div>
    );
}
