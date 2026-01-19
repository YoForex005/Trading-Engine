import React, { useState, useEffect, useRef } from 'react';
import { MoreVertical } from 'lucide-react';

interface LogEntry {
    id: string;
    time: string;
    source: string;
    message: string;
    type: 'Network' | 'System' | 'Trade' | 'Alert' | 'OpenCL';
}

const MOCK_LOGS: LogEntry[] = [
    { id: '1', time: '14:30:01', source: 'Network', message: 'RTX5 Terminal connected to RTX Cloud (12ms)', type: 'Network' },
    { id: '2', time: '14:30:02', source: 'System', message: 'Synchronization of symbols started', type: 'System' },
    { id: '3', time: '14:30:05', source: 'System', message: 'Charts loaded successfully', type: 'System' },
    { id: '4', time: '14:32:15', source: 'Trade', message: 'Order #3233875 (BUY 0.01 XAUUSD) placed', type: 'Trade' },
    { id: '5', time: '14:45:00', source: 'Alert', message: 'Price Alert: XAUUSD > 2050.00', type: 'Alert' },
    { id: '6', time: '15:00:00', source: 'Network', message: 'Ping updated: 11ms', type: 'Network' },
    { id: '7', time: '15:15:22', source: 'Trade', message: 'Order #3233875 (BUY 0.01 XAUUSD) modified (SL: 2045.00)', type: 'Trade' },
];

export function JournalTab() {
    const bottomRef = useRef<HTMLTableRowElement>(null);
    const [autoScroll, setAutoScroll] = useState(true);

    useEffect(() => {
        if (autoScroll && bottomRef.current) {
            bottomRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [autoScroll]); // Logic would need real data updates to trigger effectively

    const getRowColor = (type: string) => {
        switch (type) {
            case 'Network': return 'text-emerald-400/80';
            case 'System': return 'text-blue-400/80';
            case 'Alert': return 'text-rose-400';
            case 'Trade': return 'text-zinc-300';
            case 'OpenCL': return 'text-purple-400';
            default: return 'text-zinc-300';
        }
    };

    const formatDate = (dateStr: string) => {
        // Assuming just time for now based on prompt "Time" column
        // If dataStr is ISO, parse it. If simple time string, return as is.
        if (dateStr.includes('T') || dateStr.includes('-')) {
            return new Date(dateStr).toLocaleTimeString('en-US', { hour12: false });
        }
        return dateStr;
    };

    return (
        <div className="h-full flex flex-col bg-[#1e1e1e] text-[12px] font-mono">
            <div className="flex-1 overflow-auto scrollbar-thin scrollbar-thumb-zinc-700">
                <table className="w-full text-left border-collapse">
                    <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10 font-bold text-[10px] uppercase tracking-wider border-b border-zinc-600">
                        <tr>
                            <th className="p-1 pl-3 border-r border-zinc-700 w-32">Time</th>
                            <th className="p-1 border-r border-zinc-700 w-40">Source</th>
                            <th className="p-1">Message</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800/30">
                        {MOCK_LOGS.map((log, i) => (
                            <tr key={log.id} className={`${i % 2 === 0 ? 'bg-[#1e1e1e]' : 'bg-[#212121]'} hover:bg-[#2d3436] transition-colors`}>
                                <td className="p-1 pl-3 border-r border-zinc-800/50 text-zinc-500">{formatDate(log.time)}</td>
                                <td className="p-1 border-r border-zinc-800/50 font-bold text-zinc-400">{log.source}</td>
                                <td className={`p-1 pl-2 ${getRowColor(log.type)}`}>{log.message}</td>
                            </tr>
                        ))}
                        {/* Dummy Spacer for AutoScroll */}
                        <tr ref={bottomRef}></tr>
                    </tbody>
                </table>
            </div>
        </div>
    );
}
