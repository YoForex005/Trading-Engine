import React, { useState, useEffect, useRef } from 'react';
import { X, Minus, Square, ChevronDown, Calendar, Mail, Key, Globe, Shield, Lock } from 'lucide-react';
import ContextMenu, { ContextAction } from '../ui/ContextMenu';
import PositionWindow from './PositionWindow';

interface AccountWindowProps {
    order: any; // The order that triggered this, or null
    accountStr: string;
    onClose: () => void;
}

const TABS = [
    'Overview', 'Exposure', 'Personal', 'Account', 'Limits',
    'Profile', 'Subscriptions', 'Balance', 'Trade', 'History', 'Security'
];

function OverviewTab({ onEdit }: { onEdit: (trade: any) => void }) {
    // Mock trades for overview
    const trades = [
        { ticket: '5034373', symbol: 'xauusd', type: 'sell', vol: '0.02', price: '4637.50', sl: '0.00', tp: '0.00', cur: '4666.80', swap: '0.00', prof: '-58.60' },
        { ticket: '5034374', symbol: 'xauusd', type: 'buy', vol: '0.02', price: '4634.00', sl: '0.00', tp: '0.00', cur: '4662.39', swap: '0.00', prof: '56.78' },
        { ticket: '5034375', symbol: 'xauusd', type: 'buy', vol: '0.02', price: '4634.13', sl: '0.00', tp: '0.00', cur: '4662.39', swap: '0.00', prof: '56.52' },
        { ticket: '5034376', symbol: 'xauusd', type: 'buy', vol: '0.02', price: '4634.26', sl: '0.00', tp: '0.00', cur: '4662.39', swap: '0.00', prof: '56.26' },
        { ticket: '5034377', symbol: 'xauusd', type: 'buy', vol: '0.02', price: '4618.89', sl: '0.00', tp: '0.00', cur: '4662.39', swap: '0.00', prof: '87.00' },
    ];

    const [contextMenu, setContextMenu] = useState<{ x: number, y: number, trade: any } | null>(null);

    const handleContextMenu = (e: React.MouseEvent, trade: any) => {
        e.preventDefault();
        setContextMenu({ x: e.clientX, y: e.clientY, trade });
    };

    const getActions = (trade: any): ContextAction[] => [
        { label: 'New Order', onClick: () => console.log('New Order') },
        { label: 'Close Position', onClick: () => console.log('Close') },
        { label: 'Modify or Cancel', onClick: () => console.log('Modify') },
        { label: 'Activate', onClick: () => console.log('Activate') },
        { separator: true, label: '' },
        { label: 'Edit', onClick: () => onEdit(trade) },
        { label: 'Delete', danger: true, onClick: () => console.log('Delete') },
        { separator: true, label: '' },
        {
            label: 'Volumes',
            hasSubmenu: true,
            submenu: [
                { label: 'Lots', onClick: () => console.log('Lots') },
                { label: 'Amounts', onClick: () => console.log('Amounts') }
            ]
        },
        { label: 'Profit', hasSubmenu: true, submenu: [{ label: 'Points', onClick: () => console.log('Points') }, { label: 'Term Currency', onClick: () => console.log('Term') }] },
        { label: 'Report', hasSubmenu: true, submenu: [{ label: 'Open XML', onClick: () => console.log('XML') }] },
        { separator: true, label: '' },
        { label: 'Show Milliseconds', onClick: () => console.log('Millis') },
        { label: 'Auto Arrange', onClick: () => console.log('Auto') },
        { label: 'Grid', checked: true, onClick: () => console.log('Grid') },
        { separator: true, label: '' },
        { label: 'Columns', onClick: () => console.log('Cols') },
    ];

    return (
        <div className="flex flex-col h-full bg-[#1E1E1E] text-[11px] p-2" onClick={() => setContextMenu(null)}>
            <div className="mb-4 text-[#CCC]">
                <div className="font-bold text-sm mb-1 text-white">Trader, 680962851, ch\m1\contest.ALX , 1 : 100</div>
                <div className="h-2"></div>
                <div>MetaQuotes ID: 2919B90E</div>
                <div className="text-[#3B82F6] cursor-pointer hover:underline">user@example.com</div>
                <div className="text-[#888]">Registered: 2026.01.13 Last access: 2026.01.14 10:54 Last Address: 115.96.122.77</div>
            </div>

            <div className="border border-[#444] bg-[#252526] flex-1 overflow-auto relative">
                <table className="w-full border-collapse">
                    <thead className="bg-[#2D2D30] text-[#A0A0A0] sticky top-0">
                        <tr>
                            {['Symbol', 'Type', 'Volume', 'Price', 'S / L', 'T / P', 'Price', 'Swap', 'Profit'].map(h => (
                                <th key={h} className="border-r border-[#444] px-2 py-1 text-left font-normal">{h}</th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {trades.map((t: any, i) => (
                            <tr
                                key={i}
                                className={`border-b border-[#333] hover:bg-[#2A2D35] cursor-default ${contextMenu?.trade === t ? 'bg-[#3399FF] text-white' : ''}`}
                                onContextMenu={(e) => handleContextMenu(e, t)}
                            >
                                <td className="px-2 border-r border-[#333] flex items-center gap-1">
                                    <div className={`w-2 h-2 border border-[#666] ${t.type === 'buy' ? 'bg-blue-500' : 'bg-red-500'}`}></div>
                                    {t.symbol}
                                </td>
                                <td className="px-2 border-r border-[#333] text-[#CCC]">{t.type}</td>
                                <td className="px-2 border-r border-[#333] text-right font-mono text-[#CCC]">{t.vol}</td>
                                <td className="px-2 border-r border-[#333] text-right font-mono text-[#CCC]">{t.price}</td>
                                <td className="px-2 border-r border-[#333] text-right font-mono text-[#CCC]">{t.sl}</td>
                                <td className="px-2 border-r border-[#333] text-right font-mono text-[#CCC]">{t.tp}</td>
                                <td className="px-2 border-r border-[#333] text-right font-mono text-[#CCC]">{t.cur}</td>
                                <td className="px-2 border-r border-[#333] text-right font-mono text-[#CCC]">{t.swap}</td>
                                <td className="px-2 border-r border-[#333] text-right font-mono text-[#CCC]">{t.prof}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {contextMenu && (
                <ContextMenu
                    x={contextMenu.x}
                    y={contextMenu.y}
                    onClose={() => setContextMenu(null)}
                    actions={getActions(contextMenu.trade)}
                />
            )}


            {/* Summary Footer */}
            <div className="bg-[#333] p-1 font-bold text-[#CCC] flex gap-4 text-[10px] items-center mt-[1px]">
                <span className="text-white">+</span>
                <span>Balance: 50 000.00 USD</span>
                <span>Equity: 50 400.32</span>
                <span>Margin: 555.48</span>
                <span>Free Margin: 49 844.84</span>
                <span>Margin Level: 9 073.29 %</span>
                <div className="flex-1 text-right pr-4">400.32</div>
            </div>
        </div>
    );
}

function ExposureTab() {
    return (
        <div className="flex bg-[#1E1E1E] h-full p-2 gap-4">
            {/* Left Table */}
            <div className="flex-1 border border-[#444] bg-[#252526] flex flex-col">
                <table className="w-full border-collapse text-[11px]">
                    <thead className="bg-[#2D2D30] text-[#A0A0A0]">
                        <tr>
                            <th className="border-r border-[#444] px-2 py-1 text-left font-normal">Asset</th>
                            <th className="border-r border-[#444] px-2 py-1 text-right font-normal">Volume</th>
                            <th className="border-r border-[#444] px-2 py-1 text-right font-normal">Rate</th>
                            <th className="border-r border-[#444] px-2 py-1 text-right font-normal">USD</th>
                            <th className="border-r border-[#444] px-2 py-1 text-center font-normal">Graph</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr className="border-b border-[#333]">
                            <td className="px-2 py-1 border-r border-[#333] flex items-center gap-1 text-[#F1C40F]"><span className="text-[10px]">$</span> USD</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono">4.92959564M</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono">1.00000</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono">49 295.96</td>
                            <td className="px-2 py-1 border-r border-[#333] w-24 align-middle">
                                <div className="h-3 bg-[#3B82F6] w-full relative">
                                    <div className="absolute right-[-14px] top-0 w-3 h-3 bg-[#3B82F6] border border-[#333]"></div>
                                </div>
                            </td>
                        </tr>
                        <tr className="border-b border-[#333]">
                            <td className="px-2 py-1 border-r border-[#333] flex items-center gap-1 text-[#F1C40F]"><span className="text-[10px]">$</span> XAU</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono">12</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono">4629.03000</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono">555.48</td>
                            <td className="px-2 py-1 border-r border-[#333] w-24 align-middle">
                                <div className="h-3 bg-transparent w-full relative">
                                    <div className="absolute right-[-14px] top-0 w-3 h-3 bg-[#E74C3C] border border-[#333]"></div>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
                <div className="bg-[#333] p-1 font-bold text-[#CCC] flex gap-2 text-[10px] items-center mt-auto border-t border-[#444]">
                    <span className="text-white">+</span>
                    <span>Balance: 50 000.00 USD</span>
                    <span>Equity: 50 406.92</span>
                    <span>Margin: 555.48</span>
                    <span>Free Margin: 49 851.44</span>
                    <span>Margin Level: 9 074.48 %</span>
                </div>
            </div>

            {/* Right Pie Chart Sim */}
            <div className="w-[300px] flex flex-col items-center justify-center border border-[#444] bg-[#181818] relative">
                <div className="absolute top-2 right-2 text-[#AAA] text-xs">Long Positions</div>
                {/* CSS Pie Chart */}
                <div className="w-48 h-48 rounded-full relative" style={{ background: 'conic-gradient(#3B82F6 0% 98%, #E74C3C 98% 100%)' }}></div>
                {/* Labels */}
                <div className="absolute left-4 top-1/3 text-[#AAA] text-xs">USD</div>
                <div className="absolute right-4 top-1/2 text-[#AAA] text-xs">XAU</div>
            </div>
        </div>
    );
}

function PersonalTab() {
    return (
        <div className="p-4 grid grid-cols-[100px_1fr_100px_1fr] gap-y-2 gap-x-4 items-center bg-[#1E1E1E] text-[#CCC] text-[11px]">
            <label className="text-right text-[#888]">Name:</label>
            <input type="text" defaultValue="Trader" className="col-span-3 bg-[#252526] border border-[#444] h-6 px-1 text-[#3B82F6]" />

            <label className="text-right text-[#888]">Last name:</label>
            <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />
            <label className="text-right text-[#888]">Middle name:</label>
            <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />

            <label className="text-right text-[#888]">Company:</label>
            <input type="text" className="col-span-3 bg-[#252526] border border-[#444] h-6 px-1" />

            <div className="col-span-4 h-4"></div>

            <label className="text-right text-[#888]">Registered:</label>
            <div className="flex gap-1 items-center">
                <input type="text" defaultValue="2026.01.13 15:45:07" className="w-32 bg-[#252526] border border-[#444] h-6 px-1" disabled />
                <button className="h-6 w-6 border border-[#444] bg-[#333] flex items-center justify-center">📅</button>
            </div>
            <div className="col-span-2"></div>

            <label className="text-right text-[#888]">Language:</label>
            <select className="bg-[#252526] border border-[#444] h-6 px-1"><option>English</option></select>
            <label className="text-right text-[#888]">Status:</label>
            <select className="bg-[#252526] border border-[#444] h-6 px-1"><option></option></select>

            <label className="text-right text-[#888]">ID number:</label>
            <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />
            <label className="text-right text-[#888]">Lead campaign:</label>
            <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />

            <label className="text-right text-[#888]">MetaQuotes ID:</label>
            <div className="flex gap-1">
                <input type="text" defaultValue="2919B90E" className="w-24 bg-[#252526] border border-[#444] h-6 px-1" />
                <button className="h-6 px-2 border border-[#444] bg-[#333]">{">>"}</button>
            </div>
            <label className="text-right text-[#888]">Lead source:</label>
            <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />

            <div className="col-span-4 h-2"></div>

            <label className="text-right text-[#888]">Email:</label>
            <input type="text" defaultValue="user@example.com" className="col-span-3 bg-[#252526] border border-[#444] h-6 px-1" />

            <label className="text-right text-[#888]">Phone:</label>
            <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />
            <div className="col-span-2"></div>

            <div className="col-span-4 h-2"></div>

            <label className="text-right text-[#888]">Country:</label>
            <select className="bg-[#252526] border border-[#444] h-6 px-1"><option></option></select>
            <label className="text-right text-[#888]">State:</label>
            <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />

            <label className="text-right text-[#888]">City:</label>
            <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />
            <label className="text-right text-[#888]">Zip code:</label>
            <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />

            <label className="text-right text-[#888]">Address:</label>
            <input type="text" className="col-span-3 bg-[#252526] border border-[#444] h-6 px-1" />

            <div className="col-span-4 h-2"></div>

            <label className="text-right text-[#888]">Comment:</label>
            <input type="text" className="col-span-3 bg-[#252526] border border-[#444] h-6 px-1" />
        </div>
    );
}

function AccountTab() {
    return (
        <div className="p-4 flex flex-col gap-3 bg-[#1E1E1E] text-[11px] h-full text-[#CCC]">
            {/* Top Config Row */}
            <div className="grid grid-cols-[100px_1fr_100px_1fr] gap-x-4 gap-y-2 items-center">
                <label className="text-right text-[#888]">Group:</label>
                <select className="col-span-3 bg-[#252526] border border-[#444] h-6 px-1 text-[#3B82F6]"><option>ch\m1\contest.ALX</option></select>

                <label className="text-right text-[#888]">Color:</label>
                <div className="relative">
                    <select className="w-full bg-[#252526] border border-[#444] h-6 px-1"><option>None</option></select>
                    <div className="absolute top-[3px] right-[20px] w-3 h-3 bg-black border border-[#666]"></div>
                </div>

                <label className="text-right text-[#888]">Leverage:</label>
                <select className="bg-[#252526] border border-[#444] h-6 px-1"><option>1 : 100</option></select>
            </div>

            <div className="grid grid-cols-[100px_1fr_100px_1fr] gap-x-4 gap-y-2 items-center">
                <label className="text-right text-[#888]">Bank account:</label>
                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />
                <label className="text-right text-[#888]">Agent account:</label>
                <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1" />
            </div>

            {/* Checkboxes */}
            <div className="pl-[116px] flex flex-col gap-1.5 mt-2">
                <label className="flex items-center gap-2"><input type="checkbox" defaultChecked className="bg-[#252526] border-[#444]" /> Enable this account</label>
                <label className="flex items-center gap-2"><input type="checkbox" defaultChecked className="bg-[#252526] border-[#444]" /> Enable password change</label>
                <label className="flex items-center gap-2"><input type="checkbox" defaultChecked className="bg-[#252526] border-[#444]" /> Enable one-time password</label>
                <label className="flex items-center gap-2"><input type="checkbox" className="bg-[#252526] border-[#444]" /> Change password at next login</label>
            </div>

            {/* Trade Accounts Box */}
            <div className="flex-1 flex gap-2 mt-2">
                <label className="w-[100px] text-right text-[#888] pt-1">Trade accounts:</label>
                <div className="flex-1 border border-[#444] bg-[#252526] relative">
                    <div className="bg-[#2D2D30] text-[#A0A0A0] border-b border-[#444] flex text-xs">
                        <div className="px-2 py-1 w-1/2 border-r border-[#444]">Gateway ID</div>
                        <div className="px-2 py-1 w-1/2">Account</div>
                    </div>
                </div>
            </div>
            <div className="pl-[116px] flex gap-2">
                <button className="w-16 h-6 bg-[#333] border border-[#444] text-[#CCC]">Add</button>
                <button className="w-16 h-6 bg-[#333] border border-[#444] text-[#CCC]">Edit</button>
                <button className="w-16 h-6 bg-[#333] border border-[#444] text-[#CCC]">Delete</button>
            </div>
        </div>
    );
}

function LimitsTab() {
    return (
        <div className="p-4 grid grid-cols-2 gap-x-8 gap-y-4 bg-[#1E1E1E] text-[11px] h-full text-[#CCC]">
            <div className="flex flex-col gap-2">
                <label className="flex items-center gap-2"><input type="checkbox" defaultChecked className="bg-[#252526] border-[#444]" /> Show to regular managers</label>
                <label className="flex items-center gap-2"><input type="checkbox" defaultChecked className="bg-[#252526] border-[#444]" /> Include in server reports</label>
                <label className="flex items-center gap-2"><input type="checkbox" defaultChecked className="bg-[#252526] border-[#444]" /> Enable daily reports</label>
            </div>
            <div className="flex flex-col gap-2">
                <label className="flex items-center gap-2"><input type="checkbox" defaultChecked className="bg-[#252526] border-[#444]" /> Enable trading</label>
                <label className="flex items-center gap-2"><input type="checkbox" defaultChecked className="bg-[#252526] border-[#444]" /> Enable algo trading by Expert Advisors</label>
                <label className="flex items-center gap-2"><input type="checkbox" defaultChecked className="bg-[#252526] border-[#444]" /> Enable trailing stops</label>
            </div>

            <div className="flex flex-col gap-2 col-span-2">
                <label className="flex items-center gap-2 grayscale opacity-70"><input type="checkbox" disabled className="bg-[#252526] border-[#444]" /> Enable API connections</label>
                <label className="flex items-center gap-2 grayscale opacity-70"><input type="checkbox" disabled className="bg-[#252526] border-[#444]" /> Enable sponsored VPS hosting</label>
                <label className="flex items-center gap-2 grayscale opacity-70"><input type="checkbox" disabled className="bg-[#252526] border-[#444]" /> Allow access to subscription data via data feeds</label>
            </div>

            <div className="col-span-2 flex items-center gap-2 mt-4">
                <label className="w-40 text-right text-[#888]">Limit total value of positions:</label>
                <select className="bg-[#252526] border border-[#444] h-6 px-1 w-24 text-[#3B82F6]"><option>Default</option></select>
                <span>USD</span>
            </div>

            <div className="col-span-2 flex items-center gap-2">
                <label className="w-40 text-right text-[#888]">Limit number of active orders:</label>
                <select className="bg-[#252526] border border-[#444] h-6 px-1 w-24 text-[#3B82F6]"><option>Default</option></select>
            </div>
        </div>
    );
}

function TradeTab({ order }: { order: any }) {
    const [price, setPrice] = useState(4663.38);
    const canvasRef = useRef<HTMLCanvasElement>(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        let ticks: number[] = [];
        for (let i = 0; i < 100; i++) ticks.push(4660 + Math.random() * 10);

        let animFrameId: number;

        const draw = () => {
            if (!ctx || !canvas) return;
            ctx.fillStyle = '#111';
            ctx.fillRect(0, 0, canvas.width, canvas.height);

            // Grid
            ctx.strokeStyle = '#333';
            ctx.lineWidth = 1;
            ctx.beginPath();
            for (let x = 0; x < canvas.width; x += 40) { ctx.moveTo(x, 0); ctx.lineTo(x, canvas.height); }
            for (let y = 0; y < canvas.height; y += 40) { ctx.moveTo(0, y); ctx.lineTo(canvas.width, y); }
            ctx.stroke();

            // Ticks
            ctx.strokeStyle = '#3B82F6';
            ctx.lineWidth = 2;
            ctx.beginPath();

            // Update logic (simplified for perf in component)
            if (Math.random() > 0.9) {
                ticks.shift();
                ticks.push(4660 + Math.random() * 10);
                setPrice(ticks[ticks.length - 1]);
            }

            const min = 4660;
            const max = 4670;
            const range = max - min;

            ticks.forEach((t, i) => {
                const x = (i / 100) * canvas.width;
                const y = canvas.height - ((t - min) / range) * canvas.height;
                if (i === 0) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            });
            ctx.stroke();

            // Price Line
            const lastY = canvas.height - ((ticks[ticks.length - 1] - min) / range) * canvas.height;
            ctx.strokeStyle = '#3B82F6';
            ctx.setLineDash([2, 2]);
            ctx.beginPath();
            ctx.moveTo(0, lastY);
            ctx.lineTo(canvas.width, lastY);
            ctx.stroke();
            ctx.setLineDash([]);

            animFrameId = requestAnimationFrame(draw);
        };
        draw();
        return () => cancelAnimationFrame(animFrameId);
    }, []);

    return (
        <div className="flex-1 flex p-2 gap-2 overflow-hidden bg-[#1E1E1E]">
            {/* Tick Chart (Left) */}
            <div className="flex-1 bg-black border border-[#333] relative">
                <div className="absolute top-1 left-1 text-[#666] font-bold text-xs">XAUUSD</div>
                <canvas ref={canvasRef} width={450} height={460} className="w-full h-full" />
                <div className="absolute bottom-0 w-full h-4 bg-[#111] border-t border-[#333] flex justify-between px-1 text-[9px] text-[#555]">
                    <span>2026.01.19</span>
                    <span>13:47:00</span>
                </div>
            </div>

            {/* Order Form (Right) */}
            <div className="w-[380px] flex flex-col gap-3 px-2 pt-2">
                <div className="flex items-center gap-2">
                    <label className="w-20 text-right text-[#888]">Symbol:</label>
                    <select className="flex-1 bg-[#252526] border border-[#333] h-6 px-1 text-white">
                        <option>XAUUSD, Gold vs US Dollar</option>
                    </select>
                </div>
                <div className="flex items-center gap-2">
                    <label className="w-20 text-right text-[#888]">Type:</label>
                    <select className="flex-1 bg-[#252526] border border-[#333] h-6 px-1 text-white">
                        <option>Market Order</option>
                    </select>
                </div>

                <div className="h-[1px] bg-[#333] my-1" />

                <div className="flex items-center gap-2">
                    <label className="w-20 text-right text-[#888]">Volume:</label>
                    <input type="number" defaultValue={0.02} className="w-24 bg-[#252526] border border-[#333] h-6 px-1 text-right text-white font-mono" />
                    <span className="text-[#888]">2 XAU</span>
                </div>
                <div className="flex items-center gap-2">
                    <label className="w-20 text-right text-[#888]">Fill Policy:</label>
                    <select className="flex-1 bg-[#252526] border border-[#333] h-6 px-1 text-white">
                        <option>Fill or Kill</option>
                    </select>
                </div>

                <div className="flex items-center gap-2">
                    <label className="w-20 text-right text-[#888]">At Price:</label>
                    <input type="number" defaultValue="4663.71" className="flex-1 bg-[#252526] border border-[#333] h-6 px-1 text-right text-white font-mono" />
                    <button className="px-3 h-6 bg-[#333] border border-[#444] text-[#CCC] hover:bg-[#444]">Update</button>
                    <label className="flex items-center gap-1 text-[#CCC]">
                        <input type="checkbox" defaultChecked className="bg-[#111] border-[#444]" /> Auto
                    </label>
                </div>

                <div className="flex items-center gap-2">
                    <label className="w-20 text-right text-[#888]">Stop Loss:</label>
                    <input type="number" defaultValue="0.00" className="flex-1 bg-[#252526] border border-[#333] h-6 px-1 text-right text-white font-mono" />
                    <label className="w-16 text-right text-[#888]">Take Profit:</label>
                    <input type="number" defaultValue="0.00" className="flex-1 bg-[#252526] border border-[#333] h-6 px-1 text-right text-white font-mono" />
                </div>

                <div className="flex items-center gap-2">
                    <label className="w-20 text-right text-[#888]">Comment:</label>
                    <input type="text" className="flex-1 bg-[#252526] border border-[#333] h-6 px-1 text-white" />
                </div>

                {/* Big Price Display */}
                <div className="flex items-center justify-center gap-4 py-4">
                    <div className="text-3xl font-bold text-[#E74C3C] font-mono">{price.toFixed(2)}</div>
                    <div className="text-2xl text-[#666] font-light">/</div>
                    <div className="text-3xl font-bold text-[#3B82F6] font-mono">{(price + 0.26).toFixed(2)}</div>
                </div>

                {/* Action Buttons */}
                <div className="flex gap-2 mb-2">
                    <button className="flex-1 h-7 bg-gradient-to-b from-[#E74C3C] to-[#C0392B] border border-[#922B21] text-white shadow-lg font-bold">
                        Sell at {price.toFixed(2)}
                    </button>
                    <button className="flex-1 h-7 bg-gradient-to-b from-[#3B82F6] to-[#2980B9] border border-[#1A5276] text-white shadow-lg font-bold">
                        Buy at {(price + 0.26).toFixed(2)}
                    </button>
                </div>

                <div className="w-full h-7 bg-gradient-to-b from-[#F1C40F] to-[#D4AC0D] border border-[#B7950B] text-black font-bold shadow-lg flex items-center justify-center text-[10px] leading-tight hover:brightness-110 cursor-pointer">
                    Close #{order?.ticket || '5028662'} {order?.type || 'buy'} {order?.size || '0.02'} {order?.symbol || 'XAUUSD'} {order?.openPrice || '4634.02'} at {price.toFixed(2)}
                </div>

            </div>
        </div>
    );
}



// ... other tabs ...

function ProfileTab() {
    return (
        <div className="flex flex-col h-full bg-[#1E1E1E] text-[#CCC] text-[11px] items-center relative">
            {/* Header / WebView-like Top */}
            <div className="w-full h-16 bg-[#252526] border-b border-[#333] shadow-md flex flex-col justify-center px-6">
                <span className="text-[#3B82F6] font-normal text-lg">MetaQuotes Support Center</span>
                <span className="text-[#3B82F6] text-xs">https://support.metaquotes.net — Authorization</span>
            </div>

            <div className="flex flex-col items-center justify-center flex-1 w-full max-w-[400px]">
                <div className="mb-8 text-center">
                    <div className="text-[#CCC] mb-1">MetaQuotes Technical Support Center features unique information</div>
                    <div className="text-[#CCC]">and provides direct access to assistance from our support team</div>
                    <div className="text-[#CCC] mt-4 font-bold">Only available to authorized users</div>
                </div>

                <div className="w-full flex flex-col gap-3">
                    <div className="flex items-center border border-[#444] bg-[#252526] h-9">
                        <div className="w-9 h-full flex items-center justify-center text-[#888] border-r border-[#333]">
                            <Mail size={14} />
                        </div>
                        <input type="text" placeholder="Your email" className="flex-1 bg-transparent border-none outline-none h-full text-white px-3 placeholder-[#666]" />
                    </div>
                    <div className="flex items-center border border-[#444] bg-[#252526] h-9">
                        <div className="w-9 h-full flex items-center justify-center text-[#888] border-r border-[#333]">
                            <Key size={14} />
                        </div>
                        <input type="password" placeholder="Support Center password" className="flex-1 bg-transparent border-none outline-none h-full text-white px-3 placeholder-[#666]" />
                    </div>

                    <div className="text-right text-[#3B82F6] cursor-pointer hover:underline text-[10px]">Password recovery</div>

                    <button className="h-10 bg-[#00CC66] hover:bg-[#00B359] text-black font-bold border-none text-[12px] shadow-sm">Enter MetaQuotes Support Center</button>

                    <div className="text-center text-[#888] text-[10px] mt-2">
                        Don't have a Support Center account? <span className="text-[#3B82F6] cursor-pointer hover:underline">Register</span>
                    </div>

                    <div className="flex items-center justify-center gap-2 text-[#3B82F6] mt-2 text-[10px]">
                        <Globe size={12} /> <span className="cursor-pointer hover:underline">Official Support Channel</span>
                    </div>
                </div>
            </div>

            {/* Footer Logos */}
            <div className="w-full grid grid-cols-3 gap-8 text-left text-[10px] text-[#888] px-12 pb-6 border-t border-[#333] pt-4 bg-[#252526]">
                <div className="flex gap-2">
                    <div className="w-8 h-8 bg-[#333] rounded flex items-center justify-center"><Shield size={16} className="text-[#555]" /></div>
                    <div>
                        <div className="text-[#CCC] font-bold mb-0.5">Finteza</div>
                        <div>Use analytics for a</div>
                        <div>brokerage business</div>
                    </div>
                </div>
                <div className="flex gap-2">
                    <div className="w-8 h-8 bg-[#333] rounded flex items-center justify-center"><Lock size={16} className="text-[#555]" /></div>
                    <div>
                        <div className="text-[#CCC] font-bold mb-0.5">App Store</div>
                        <div>Get new MetaTrader 5 features</div>
                    </div>
                </div>
                <div className="flex gap-2">
                    <div className="w-8 h-8 bg-[#333] rounded flex items-center justify-center"><Globe size={16} className="text-[#555]" /></div>
                    <div>
                        <div className="text-[#CCC] font-bold mb-0.5">Webinars</div>
                        <div>Find out how the platform can</div>
                        <div>help you</div>
                    </div>
                </div>
            </div>
        </div>
    );
}

function SubscriptionsTab() {
    return (
        <div className="bg-[#1E1E1E] h-full flex flex-col font-sans text-[11px] text-[#CCC] p-0.5">
            <div className="bg-[#1E1E1E] flex-1 border border-[#444] bg-[#252526]">
                <table className="w-full border-collapse">
                    <thead className="bg-[#2D2D30] text-[#A0A0A0]">
                        <tr>
                            {['ID', 'Subscription', 'Status', 'Subscription time', 'Renewal time', 'Expiration time', 'Price'].map(h => (
                                <th key={h} className="border-r border-[#333] px-2 py-1 text-left font-normal border-b border-[#333]">{h}</th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {/* Empty State */}
                    </tbody>
                </table>
            </div>
        </div>
    );
}

function BalanceTab() {
    return (
        <div className="bg-[#1E1E1E] h-full flex flex-col p-4 gap-4 text-[11px] text-[#CCC]">
            {/* Operation Form - Strict Grid */}
            <div className="grid grid-cols-[80px_300px] gap-y-3 gap-x-2 items-center">
                <label className="text-right text-[#888]">Operation:</label>
                <select className="bg-[#252526] border border-[#444] h-6 px-1 w-full"><option>Balance</option></select>

                <label className="text-right text-[#888]">Amount:</label>
                <div className="flex gap-2">
                    <input type="text" className="bg-[#252526] border border-[#444] h-6 px-1 flex-1 text-right" placeholder="0.00" />
                    <span className="self-center text-[#888]">USD</span>
                </div>

                <label className="text-right text-[#888]">Comment:</label>
                <select className="bg-[#252526] border border-[#444] h-6 px-1 w-full text-[#3B82F6]"><option>... put your comment here...</option></select>
            </div>

            <div className="pl-[88px] flex flex-col gap-3">
                <label className="flex items-center gap-2 select-none">
                    <input type="checkbox" className="bg-[#252526] border-[#444]" />
                    Conduct balance operation without checking the free margin and the current balance on the account
                </label>

                <div className="flex gap-2 w-[300px]">
                    <button className="flex-1 bg-[#333] border border-[#444] h-7 text-[#CCC] hover:bg-[#444] shadow-sm">Deposit</button>
                    <button className="flex-1 bg-[#333] border border-[#444] h-7 text-[#CCC] hover:bg-[#444] shadow-sm">Withdrawal</button>
                </div>
            </div>

            {/* History Table */}
            <div className="flex-1 border border-[#444] mt-2 bg-[#252526] flex flex-col">
                <table className="w-full border-collapse">
                    <thead className="bg-[#2D2D30] text-[#A0A0A0]">
                        <tr>
                            <th className="border-r border-[#333] px-2 py-1 text-left font-normal border-b border-[#333]">Time</th>
                            <th className="border-r border-[#333] px-2 py-1 text-right font-normal border-b border-[#333]">Deal</th>
                            <th className="border-r border-[#333] px-2 py-1 text-right font-normal border-b border-[#333]">Type</th>
                            <th className="border-r border-[#333] px-2 py-1 text-right font-normal border-b border-[#333]">Amount</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr className="border-b border-[#333]">
                            <td className="px-2 py-1 border-r border-[#333] text-[#CCC]">2026.01.13 15:45:08</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono text-[#CCC]">4006818</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right text-[#CCC]">balance</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono text-[#3B82F6]">50 000.00</td>
                        </tr>
                    </tbody>
                </table>
                <div className="bg-[#333] p-1 text-[#CCC] flex gap-2 text-[10px] items-center mt-auto border-t border-[#444]">
                    <span className="text-white font-bold bg-[#888] rounded-sm w-3 h-3 flex items-center justify-center text-[8px]">+</span>
                    <span>Balance: 50 000.00 USD</span>
                    <span className="text-[#888]">|</span>
                    <span>Equity: 50 393.12</span>
                    <span className="text-[#888]">|</span>
                    <span>Margin: 555.48</span>
                    <span className="text-[#888]">|</span>
                    <span>Free: 49 837.64 / 9 072.00 %</span>
                </div>
            </div>

            <div className="flex justify-end gap-1 text-[10px] items-center">
                <select className="bg-[#252526] border border-[#444] h-5 px-1 w-24"><option>Last 3 months</option></select>
                <div className="relative">
                    <input type="text" defaultValue="2025.10.01" className="bg-[#252526] border border-[#444] h-5 w-24 px-1 text-center" />
                    <Calendar size={10} className="absolute right-1 top-1 text-[#888]" />
                </div>
                <div className="relative">
                    <input type="text" defaultValue="2026.01.19" className="bg-[#252526] border border-[#444] h-5 w-24 px-1 text-center" />
                    <Calendar size={10} className="absolute right-1 top-1 text-[#888]" />
                </div>
                <button className="bg-[#333] border border-[#444] h-5 px-3 text-[#CCC] hover:bg-[#444]">Request</button>
            </div>
        </div>
    );
}

function HistoryTab() {
    return (
        <div className="bg-[#1E1E1E] h-full flex flex-col p-2 gap-2 text-[11px] text-[#CCC]">
            <div className="flex-1 border border-[#444] bg-[#252526] flex flex-col overflow-auto">
                <table className="w-full border-collapse min-w-[800px]">
                    <thead className="bg-[#2D2D30] text-[#A0A0A0] sticky top-0 shadow-sm">
                        <tr>
                            {['Time', 'Ticket', 'Type', 'Volume', 'Symbol', 'Price', 'S / L', 'T / P', 'Close Time', 'Close Price', 'Profit'].map(h => (
                                <th key={h} className="border-r border-[#333] px-2 py-1 text-left font-normal border-b border-[#333]">{h}</th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        <tr className="border-b border-[#333]">
                            <td className="px-2 py-1 border-r border-[#333] text-[#CCC]">2026.01.13 15:45:08</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono text-[#CCC]">4006818</td>
                            <td className="px-2 py-1 border-r border-[#333] text-[#CCC]">balance</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono text-[#CCC]"></td>
                            <td className="px-2 py-1 border-r border-[#333] text-[#CCC]"></td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono text-[#CCC]"></td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono text-[#CCC]">{parseFloat("0.00").toFixed(2)}</td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono text-[#CCC]">{parseFloat("0.00").toFixed(2)}</td>
                            <td className="px-2 py-1 border-r border-[#333] text-[#CCC]"></td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono text-[#CCC]"></td>
                            <td className="px-2 py-1 border-r border-[#333] text-right font-mono text-[#3B82F6]">50 000.00</td>
                        </tr>
                    </tbody>
                </table>
            </div>

            {/* Footer Summary */}
            <div className="border border-[#444] bg-[#2D2D30] p-1 flex items-center text-[#CCC] font-bold text-[10px]">
                <div className="flex items-center gap-2 pl-2">
                    <span className="w-1.5 h-1.5 bg-[#AAA] block transform rotate-45"></span>
                    Profit: 0.00
                </div>
                <div className="ml-auto flex gap-4 pr-2">
                    <span>Deposit: 0.00</span>
                    <span>Withdrawal: 0.00</span>
                    <span>Credit: 0.00</span>
                    <span>Balance: 0.00</span>
                </div>
            </div>

            <div className="flex justify-end gap-1 text-[10px] items-center mt-1">
                <select className="bg-[#252526] border border-[#444] h-5 px-1 w-24"><option>Last month</option></select>
                <div className="relative">
                    <input type="text" defaultValue="2026.01.01" className="bg-[#252526] border border-[#444] h-5 w-24 px-1 text-center" />
                    <Calendar size={10} className="absolute right-1 top-1 text-[#888]" />
                </div>
                <div className="relative">
                    <input type="text" defaultValue="2026.01.19" className="bg-[#252526] border border-[#444] h-5 w-24 px-1 text-center" />
                    <Calendar size={10} className="absolute right-1 top-1 text-[#888]" />
                </div>
                <button className="bg-[#333] border border-[#444] h-5 px-3 text-[#CCC] hover:bg-[#444]">Request</button>
            </div>
        </div>
    );
}

function SecurityTab() {
    return (
        <div className="p-4 flex flex-col gap-4 bg-[#1E1E1E] text-[11px] h-full text-[#CCC]">
            <div>
                <div className="mb-1 text-[#888]">Master password is used for full access to the trading account</div>
                <div className="flex gap-2">
                    <input type="password" className="w-[300px] bg-[#252526] border border-[#444] h-6 px-1" />
                    <button className="w-20 bg-[#333] border border-[#444] h-6 text-[#CCC] hover:bg-[#444] shadow-sm">Check</button>
                    <button className="w-20 bg-[#333] border border-[#444] h-6 text-[#CCC] hover:bg-[#444] shadow-sm">Change</button>
                    <button className="w-20 bg-[#333] border border-[#444] h-6 text-[#CCC] hover:bg-[#444] shadow-sm">Generate</button>
                </div>
                <div className="text-[10px] text-[#666] mt-1 ml-1">minimum 8 characters</div>
            </div>

            <div>
                <div className="mb-1 text-[#888]">Investor password is used for limited access to the trading account in read-only mode</div>
                <div className="flex gap-2">
                    <input type="password" className="w-[300px] bg-[#252526] border border-[#444] h-6 px-1" />
                    <button className="w-20 bg-[#333] border border-[#444] h-6 text-[#CCC] hover:bg-[#444] shadow-sm">Check</button>
                    <button className="w-20 bg-[#333] border border-[#444] h-6 text-[#CCC] hover:bg-[#444] shadow-sm">Change</button>
                </div>
                <div className="text-[10px] text-[#666] mt-1 ml-1">minimum 8 characters</div>
            </div>

            <div>
                <div className="mb-1 text-[#888]">API password is used for access to the server using Web API</div>
                <div className="flex gap-2">
                    <input type="password" className="w-[300px] bg-[#252526] border border-[#444] h-6 px-1" />
                    <button className="w-20 bg-[#333] border border-[#444] h-6 text-[#CCC] hover:bg-[#444] shadow-sm">Check</button>
                    <button className="w-20 bg-[#333] border border-[#444] h-6 text-[#CCC] hover:bg-[#444] shadow-sm">Change</button>
                </div>
                <div className="text-[10px] text-[#666] mt-1 ml-1">minimum 8 characters</div>
            </div>

            <div className="mt-2">
                <div className="mb-1 text-[#888]">Phone password allows to identify the account owner when performing trade operations by phone</div>
                <input type="password" className="w-[300px] bg-[#252526] border border-[#444] h-6 px-1" />
                <div className="text-[10px] text-[#666] mt-1 ml-1">to view password set focus to field</div>
            </div>

            <div className="mt-2">
                <div className="mb-1 text-[#888]">Shared secret key in combination with the current timestamp is used to generate one-time password</div>
                <div className="mb-1 text-[#CCC]">OTP secret key</div>
                <input type="password" className="w-[300px] bg-[#252526] border border-[#444] h-6 px-1" />
            </div>
        </div>
    );
}

// --- Main Window Component ---

export default function AccountWindow({ order, accountStr, onClose }: AccountWindowProps) {
    const [activeTab, setActiveTab] = useState('Overview');
    const [editingPosition, setEditingPosition] = useState<any | null>(null);

    const renderContent = () => {
        switch (activeTab) {
            case 'Overview': return <OverviewTab onEdit={setEditingPosition} />;
            case 'Exposure': return <ExposureTab />;
            case 'Personal': return <PersonalTab />;
            case 'Account': return <AccountTab />;
            case 'Limits': return <LimitsTab />;
            case 'Trade': return <TradeTab order={order} />;
            case 'Profile': return <ProfileTab />;
            case 'Subscriptions': return <SubscriptionsTab />;
            case 'Balance': return <BalanceTab />;
            case 'History': return <HistoryTab />;
            case 'Security': return <SecurityTab />;
            default: return <div className="p-4 text-[#888]">Tab content not implemented yet.</div>;
        }
    }

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm select-none font-sans">
            {/* Window Container */}
            <div className="w-[900px] h-[600px] bg-[#1E1E1E] border border-[#444] shadow-2xl flex flex-col text-[#CCC] text-[11px]">

                {/* Title Bar */}
                <div className="h-7 bg-[#2D2D30] flex items-center justify-between px-2 select-none border-b border-[#333]">
                    <div className="flex items-center gap-2">
                        <span className="text-[#2ECC71] font-bold">👤</span>
                        <span className="font-bold text-white">Account: {accountStr}</span>
                    </div>
                    <div className="flex items-center gap-2 text-[#888]">
                        <Minus size={14} className="hover:text-white cursor-pointer" />
                        <Square size={12} className="hover:text-white cursor-pointer" />
                        <X size={14} className="hover:text-red-500 cursor-pointer" onClick={onClose} />
                    </div>
                </div>

                {/* Tab Strip */}
                <div className="h-7 bg-[#1E1E1E] border-b border-[#333] flex items-end px-2 gap-1">
                    {TABS.map(tab => (
                        <div
                            key={tab}
                            onClick={() => setActiveTab(tab)}
                            className={`
                                px-3 py-1 cursor-pointer 
                                ${activeTab === tab
                                    ? 'text-[#3B82F6] border-b-2 border-[#3B82F6] bg-[#252526]'
                                    : 'text-[#888] hover:bg-[#252526] hover:text-[#BBB]'}
                            `}
                        >
                            {tab}
                        </div>
                    ))}
                </div>

                {/* Content Area */}
                <div className="flex-1 overflow-auto bg-[#1E1E1E]">
                    {renderContent()}
                </div>

                {/* Render Detail Window if Editing */}
                {editingPosition && (
                    <PositionWindow
                        position={editingPosition}
                        onClose={() => setEditingPosition(null)}
                    />
                )}
            </div>
        </div>
    );
}
