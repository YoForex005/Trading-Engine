import React, { useEffect, useState } from 'react';
import { Trash2, Edit2, Layers, Undo2, Lock, Unlock, Hash } from 'lucide-react';
import { drawingManager } from '../services/drawingManager';

interface DrawingContextMenuProps {
    onClose: () => void;
}

export const DrawingContextMenu: React.FC<DrawingContextMenuProps> = ({ onClose }) => {
    const [menuState, setMenuState] = useState<{ x: number; y: number; drawingId: string } | null>(null);

    useEffect(() => {
        const handleContextMenu = (e: any) => {
            setMenuState({
                x: e.detail.x,
                y: e.detail.y,
                drawingId: e.detail.drawingId
            });
        };

        window.addEventListener('drawing:contextmenu' as any, handleContextMenu);
        return () => window.removeEventListener('drawing:contextmenu' as any, handleContextMenu);
    }, []);

    if (!menuState) return null;

    const handleDelete = () => {
        drawingManager.deleteDrawing(menuState.drawingId);
        onClose();
        setMenuState(null);
    };

    const handleToggleLock = () => {
        const drawing = drawingManager.getDrawings().find(d => d.id === menuState.drawingId);
        if (drawing) {
            drawing.locked = !drawing.locked;
            drawingManager.saveToBackend(drawing);
            drawingManager.renderAllDrawings();
        }
        onClose();
        setMenuState(null);
    };

    const handleSetLineStyle = (style: 'solid' | 'dashed' | 'dotted') => {
        const drawing = drawingManager.getDrawings().find(d => d.id === menuState.drawingId);
        if (drawing) {
            drawing.lineStyle = style;
            drawingManager.saveToBackend(drawing);
            drawingManager.renderAllDrawings();
        }
        onClose();
        setMenuState(null);
    };

    const handleProperties = () => {
        const drawing = drawingManager.getDrawings().find(d => d.id === menuState.drawingId);
        if (drawing) {
            const newColor = prompt('Enter color (hex):', drawing.color || '#3b82f6');
            if (newColor) {
                drawing.color = newColor;
                drawingManager.saveToBackend(drawing);
                drawingManager.renderAllDrawings();
            }
        }
        onClose();
        setMenuState(null);
    };

    const handleDeleteAll = () => {
        if (confirm('Delete all drawing objects?')) {
            drawingManager.clearAllDrawings();
        }
        onClose();
        setMenuState(null);
    };

    const handleDeleteSelected = () => {
        drawingManager.deleteSelected();
        onClose();
        setMenuState(null);
    };

    const currentDrawing = drawingManager.getDrawings().find(d => d.id === menuState.drawingId);

    return (
        <div
            className="fixed z-[1000] bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl py-1 w-56 animate-in fade-in zoom-in-95 duration-100"
            style={{ left: menuState.x, top: menuState.y }}
            onMouseLeave={() => setMenuState(null)}
            onClick={(e) => e.stopPropagation()}
        >
            <ContextMenuItem
                icon={<Edit2 size={14} />}
                label="Properties"
                onClick={handleProperties}
            />
            <ContextMenuItem
                icon={<Trash2 size={14} />}
                label="Delete"
                onClick={handleDelete}
            />

            <div className="h-[1px] bg-zinc-800 my-1 mx-2" />

            <ContextMenuItem
                icon={currentDrawing?.locked ? <Unlock size={14} /> : <Lock size={14} />}
                label={currentDrawing?.locked ? "Unlock Object" : "Lock Object"}
                onClick={handleToggleLock}
            />

            <div className="group/submenu relative">
                <ContextMenuItem
                    icon={<Hash size={14} />}
                    label="Line Style"
                    onClick={() => { }}
                />
                <div className="absolute left-full top-0 ml-1 hidden group-hover/submenu:block bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl py-1 w-32 animate-in fade-in slide-in-from-left-2 duration-100">
                    <ContextMenuItem label="Solid" onClick={() => handleSetLineStyle('solid')} icon={null} />
                    <ContextMenuItem label="Dashed" onClick={() => handleSetLineStyle('dashed')} icon={null} />
                    <ContextMenuItem label="Dotted" onClick={() => handleSetLineStyle('dotted')} icon={null} />
                </div>
            </div>

            <div className="h-[1px] bg-zinc-800 my-1 mx-2" />

            <ContextMenuItem
                icon={<Trash2 size={14} className="text-red-400" />}
                label="Delete Selected"
                onClick={handleDeleteSelected}
            />
            <ContextMenuItem
                icon={<Layers size={14} />}
                label="Delete All Shapes"
                onClick={() => drawingManager.deleteByType('shapes')}
            />
            <ContextMenuItem
                icon={<Layers size={14} />}
                label="Delete All Drawings"
                onClick={handleDeleteAll}
            />

            <div className="h-[1px] bg-zinc-800 my-1 mx-2" />

            <ContextMenuItem
                icon={<Undo2 size={14} />}
                label="Undo Delete"
                onClick={() => drawingManager.undoDelete()}
            />
        </div>
    );
};

const ContextMenuItem = ({ icon, label, onClick }: { icon: React.ReactNode, label: string, onClick: () => void }) => (
    <div
        className="flex items-center gap-3 px-3 py-2 text-[12px] text-zinc-300 hover:bg-[#2d2d2d] hover:text-white cursor-pointer transition-colors"
        onClick={(e) => {
            e.stopPropagation();
            onClick();
        }}
    >
        {icon && <span className="text-zinc-500">{icon}</span>}
        <span className={!icon ? 'ml-7' : ''}>{label}</span>
    </div>
);

