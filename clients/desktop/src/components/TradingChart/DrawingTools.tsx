import { MousePointer2, Minus, Square, Trash2 } from 'lucide-react';

export type DrawingType = 'cursor' | 'line' | 'horizontal_line' | 'rectangle';

interface DrawingToolsProps {
    activeTool: DrawingType;
    onToolChange: (tool: DrawingType) => void;
    onClearAll: () => void;
}

export function DrawingTools({ activeTool, onToolChange, onClearAll }: DrawingToolsProps) {
    const tools = [
        { id: 'cursor', icon: <MousePointer2 size={16} />, label: 'Cursor' },
        { id: 'line', icon: <Minus size={16} className="rotate-45" />, label: 'Trendline' },
        { id: 'horizontal_line', icon: <Minus size={16} />, label: 'Horizontal' },
        { id: 'rectangle', icon: <Square size={16} />, label: 'Rectangle' },
    ];

    return (
        <div className="absolute left-4 top-16 flex flex-col gap-1 bg-zinc-900/90 border border-zinc-700/50 p-1 rounded-lg backdrop-blur-sm z-20 shadow-xl">
            {tools.map((tool) => (
                <button
                    key={tool.id}
                    onClick={() => onToolChange(tool.id as DrawingType)}
                    className={`p-2 rounded flex items-center justify-center transition-all ${activeTool === tool.id
                        ? 'bg-emerald-500 text-white shadow-sm'
                        : 'text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800'
                        }`}
                    title={tool.label}
                >
                    {tool.icon}
                </button>
            ))}
            <div className="h-px bg-zinc-700/50 my-1" />
            <button
                onClick={onClearAll}
                className="p-2 rounded flex items-center justify-center text-red-400 hover:bg-red-500/10 hover:text-red-300 transition-colors"
                title="Clear All Drawings"
            >
                <Trash2 size={16} />
            </button>
        </div>
    );
}
