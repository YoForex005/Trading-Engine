import React from 'react';
import { Flag, ArrowUp, ArrowDown } from 'lucide-react';

const EVENTS = [
    { time: '14:30', currency: 'USD', event: 'Initial Jobless Claims', priority: 3, actual: '220K', forecast: '215K', prev: '210K' },
    { time: '14:30', currency: 'USD', event: 'Philadelphia Fed Manufacturing Index', priority: 2, actual: '-5.0', forecast: '-8.0', prev: '-10.6' },
    { time: '16:00', currency: 'USD', event: 'Existing Home Sales', priority: 2, actual: '', forecast: '3.90M', prev: '3.82M' },
    { time: '17:30', currency: 'USD', event: 'Natural Gas Storage', priority: 1, actual: '', forecast: '', prev: '-14B' },
];

export default function EconomicCalendar() {
    return (
        <div className="flex flex-col h-full bg-charcoal-900 select-none font-sans">
            {/* Table Header */}
            <div className="h-6 bg-charcoal-950 border-b border-charcoal-border flex items-center text-xs text-zinc-500 font-semibold px-0">
                <div className="w-16 pl-2 border-r border-charcoal-border">Time</div>
                <div className="w-16 pl-2 border-r border-charcoal-border">Cur</div>
                <div className="flex-1 pl-2 border-r border-charcoal-border">Event</div>
                <div className="w-16 pl-2 border-r border-charcoal-border center text-center">Imp</div>
                <div className="w-20 pr-2 border-r border-charcoal-border text-right">Actual</div>
                <div className="w-20 pr-2 border-r border-charcoal-border text-right">Forecast</div>
                <div className="w-20 pr-2 border-r border-charcoal-border text-right">Prev</div>
            </div>

            {/* Table Body */}
            <div className="flex-1 overflow-auto custom-scrollbar">
                {EVENTS.map((evt, i) => (
                    <div key={i} className={`flex h-6 items-center text-[11px] text-zinc-300 border-b border-charcoal-border hover:bg-charcoal-800 ${i % 2 === 0 ? 'bg-charcoal-900' : 'bg-[#111214]'}`}>
                        <div className="w-16 pl-2 border-r border-charcoal-border/50 truncate text-zinc-400">{evt.time}</div>
                        <div className="w-16 pl-2 border-r border-charcoal-border/50 truncate font-bold text-zinc-400">{evt.currency}</div>
                        <div className="flex-1 pl-2 border-r border-charcoal-border/50 truncate">{evt.event}</div>
                        <div className="w-16 pl-2 border-r border-charcoal-border/50 flex justify-center">
                            {Array.from({ length: evt.priority }).map((_, idx) => (
                                <div key={idx} className={`w-3 h-1.5 mx-0.5 rounded-none ${evt.priority === 3 ? 'bg-red-500' : evt.priority === 2 ? 'bg-orange-500' : 'bg-zinc-600'}`}></div>
                            ))}
                        </div>
                        <div className="w-20 pr-2 border-r border-charcoal-border/50 text-right font-mono text-rtx-yellow">{evt.actual}</div>
                        <div className="w-20 pr-2 border-r border-charcoal-border/50 text-right font-mono">{evt.forecast}</div>
                        <div className="w-20 pr-2 border-r border-charcoal-border/50 text-right font-mono text-zinc-500">{evt.prev}</div>
                    </div>
                ))}
            </div>
        </div>
    );
}
