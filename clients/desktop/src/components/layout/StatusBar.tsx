import { Wifi, Server } from 'lucide-react';

export function StatusBar() {
    return (
        <div className="bg-[#e1e1e1] border-t border-zinc-400 h-6 flex items-center justify-between px-2 text-[11px] font-sans text-black select-none z-50">
            <div className="text-zinc-600 font-medium">
                For Help, press F1
            </div>
            <div className="flex items-center gap-4">
                <div className="flex items-center gap-1.5 border-r border-zinc-400 pr-4">
                    <Server size={12} className="text-zinc-500" />
                    <span className="font-medium text-black">Common</span>
                </div>
                <div className="flex items-center gap-1.5">
                    <div className="flex flex-col gap-[1px]">
                        <div className="w-2.5 h-[2px] bg-green-500"></div>
                        <div className="w-2.5 h-[2px] bg-green-500"></div>
                        <div className="w-2.5 h-[2px] bg-green-500"></div>
                    </div>
                    <span className="font-medium text-black">821/6 Kb</span>
                    <Wifi size={12} className="text-green-600 ml-1" />
                </div>
            </div>
        </div>
    );
}
