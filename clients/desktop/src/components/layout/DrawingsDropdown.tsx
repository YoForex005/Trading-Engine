import React, { useState } from 'react';
import { Trash2, Undo2, MousePointer2, ChevronDown, ListFilter } from 'lucide-react';
import { drawingManager } from '../../services/drawingManager';

export const DrawingsDropdown: React.FC = () => {
    const [isOpen, setIsOpen] = useState(false);

    const handleAction = (action: string) => {
        switch (action) {
            case 'clear_all':
                if (confirm('Are you sure you want to delete all drawing objects?')) {
                    drawingManager.clearAllDrawings();
                }
                break;
            case 'delete_selected':
                drawingManager.deleteSelected();
                break;
            case 'undo':
                drawingManager.undoDelete();
                break;
            case 'unselect_all':
                drawingManager.unselectAll();
                break;
        }
        setIsOpen(false);
    };

    return (
        <div className="relative">
            <button
                className="flex items-center gap-1 px-1.5 py-1.5 rounded text-zinc-400 hover:text-zinc-200 hover:bg-[#333] transition-all"
                onClick={() => setIsOpen(!isOpen)}
                title="Drawing Operations"
            >
                <ListFilter size={15} />
                <ChevronDown size={10} className="text-zinc-600" />
            </button>

            {isOpen && (
                <>
                    <div
                        className="fixed inset-0 z-[90]"
                        onClick={() => setIsOpen(false)}
                    />
                    <div className="absolute top-full right-0 mt-1 w-52 bg-[#1e1e1e] border border-zinc-800 rounded shadow-2xl z-[100] py-1 overflow-hidden animate-in fade-in slide-in-from-top-2 duration-150">
                        <DropdownItem
                            icon={<Trash2 size={13} className="text-red-400" />}
                            label="Delete Selected"
                            onClick={() => handleAction('delete_selected')}
                        />
                        <DropdownItem
                            icon={<Trash2 size={13} />}
                            label="Delete All"
                            onClick={() => handleAction('clear_all')}
                        />
                        <div className="h-[1px] bg-zinc-800 my-1 mx-2" />
                        <DropdownItem
                            icon={<Undo2 size={13} />}
                            label="Undo Last Delete"
                            onClick={() => handleAction('undo')}
                        />
                        <DropdownItem
                            icon={<MousePointer2 size={13} />}
                            label="Unselect All"
                            onClick={() => handleAction('unselect_all')}
                        />
                    </div>
                </>
            )}
        </div>
    );
};

const DropdownItem = ({ icon, label, onClick }: { icon: React.ReactNode, label: string, onClick: () => void }) => (
    <div
        className="flex items-center gap-3 px-3 py-2 text-[11px] text-zinc-300 hover:bg-[#2d2d2d] hover:text-white cursor-pointer transition-colors"
        onClick={onClick}
    >
        <span className="w-4 flex justify-center text-zinc-500">{icon}</span>
        <span>{label}</span>
    </div>
);
