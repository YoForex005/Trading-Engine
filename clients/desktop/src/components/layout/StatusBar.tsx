import { Server } from 'lucide-react';
import { ConnectionStatus } from '../ConnectionStatus';

interface StatusBarProps {
  wsRef?: React.RefObject<WebSocket | null>;
  reconnectCallback?: () => void;
}

export function StatusBar({ wsRef, reconnectCallback }: StatusBarProps) {
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
                {/* Connection Status Indicator (Live) */}
                {wsRef && (
                    <ConnectionStatus wsRef={wsRef} reconnectCallback={reconnectCallback} />
                )}
            </div>
        </div>
    );
}
