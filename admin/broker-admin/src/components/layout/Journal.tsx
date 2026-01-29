import React, { useEffect, useRef, useState } from 'react';

const LOG_MESSAGES = [
    { source: 'Network', msg: 'Ping to RTX-Live-01: 12ms', type: 'INFO' },
    { source: 'Network', msg: 'Ping to RTX-Live-01: 11ms', type: 'INFO' },
    { source: 'System', msg: 'Memory usage: 45MB', type: 'INFO' },
    { source: 'Trades', msg: 'Order #10023 filled at 2034.50', type: 'INFO' },
    { source: 'Trades', msg: 'Order #10024 placed (Pending)', type: 'INFO' },
    { source: 'Risk', msg: 'Margin level check: OK', type: 'INFO' },
    { source: 'Risk', msg: 'Client 500122 margins updated', type: 'INFO' },
    { source: 'Dealer', msg: 'Spread widening on XAUUSD', type: 'WARN' },
    { source: 'Dealer', msg: 'Quote filtered: 2034.55', type: 'INFO' },
    { source: 'System', msg: 'Garbage collection started', type: 'INFO' },
];

export default function Journal() {
    const [logs, setLogs] = useState<{ id: number; time: string; source: string; message: string; type: string }[]>([
        { id: 1, time: new Date().toLocaleTimeString('en-GB') + '.123', source: 'System', message: 'Journal Started', type: 'INFO' }
    ]);
    const bottomRef = useRef<HTMLDivElement>(null);
    const idCounter = useRef(2);

    useEffect(() => {
        const interval = setInterval(() => {
            const randomLog = LOG_MESSAGES[Math.floor(Math.random() * LOG_MESSAGES.length)];
            const time = new Date();
            const timeStr = `${time.toLocaleTimeString('en-GB')}.${time.getMilliseconds().toString().padStart(3, '0')}`;

            setLogs(prev => {
                const newLogs = [...prev, {
                    id: idCounter.current++,
                    time: timeStr,
                    source: randomLog.source,
                    message: randomLog.msg,
                    type: randomLog.type
                }];
                if (newLogs.length > 200) newLogs.shift(); // Keep buffer size reasonable
                return newLogs;
            });
        }, 800); // Fast updates to feel "Alive"

        return () => clearInterval(interval);
    }, []);

    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: 'auto' }); // Instant scroll, no smooth animation
    }, [logs]);

    return (
        <div className="h-full flex flex-col bg-[#121316] font-mono text-[11px]">
            <div className="flex-1 overflow-auto custom-scrollbar relative">
                <table className="w-full text-left border-collapse table-fixed">
                    <thead className="sticky top-0 bg-[#1E2026] text-[#888] shadow-[0_1px_0_0_#000]">
                        <tr className="h-5">
                            <th className="w-24 px-2 py-0 border-r border-[#383A42] font-normal">Time</th>
                            <th className="w-32 px-2 py-0 border-r border-[#383A42] font-normal">Source</th>
                            <th className="px-2 py-0 font-normal">Message</th>
                        </tr>
                    </thead>
                    <tbody className="bg-[#121316]">
                        {logs.map((log, i) => (
                            <tr key={log.id} className="hover:bg-[#1E2026] leading-4">
                                <td className="px-2 py-0.5 border-r border-[#333] text-[#AAA] whitespace-nowrap">{log.time}</td>
                                <td className="px-2 py-0.5 border-r border-[#333] font-semibold text-[#CCC] whitespace-nowrap">{log.source}</td>
                                <td className={`px-2 py-0.5 whitespace-nowrap ${log.type === 'WARN' ? 'text-[#F5C542]' : 'text-[#CCC]'}`}>
                                    {log.message}
                                </td>
                            </tr>
                        ))}
                        <tr ref={bottomRef as any} />
                    </tbody>
                </table>
                {/* Invisible anchor for scrolling */}
                <div ref={bottomRef} />
            </div>
        </div>
    );
}
