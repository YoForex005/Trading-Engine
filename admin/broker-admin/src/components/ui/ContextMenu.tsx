import React, { useEffect, useRef, useState } from 'react';
import { ChevronRight, Check } from 'lucide-react';

export interface ContextAction {
    label: string;
    onClick?: () => void;
    separator?: boolean;
    shortcut?: string;
    hasSubmenu?: boolean;
    submenu?: ContextAction[];
    danger?: boolean;
    checked?: boolean; // For columns
}

interface ContextMenuProps {
    x: number;
    y: number;
    onClose: () => void;
    actions: ContextAction[];
    level?: number;
}

export default function ContextMenu({ x, y, onClose, actions, level = 0 }: ContextMenuProps) {
    const menuRef = useRef<HTMLDivElement>(null);
    const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

    useEffect(() => {
        // Only top-level closes on outside click
        if (level === 0) {
            const handleClickOutside = (event: MouseEvent) => {
                // If click is not in ANY context menu (logic tricky for recursion, but simplistic check works for now)
                // We rely on the parent closing everything usually.
                if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                    // Check if target is inside ANY other context menu (submenus are portals or absolute children)
                    // For simply recursion within same DOM tree:
                    onClose();
                }
            };
            document.addEventListener('mousedown', handleClickOutside);
            return () => document.removeEventListener('mousedown', handleClickOutside);
        }
    }, [onClose, level]);

    return (
        <div
            ref={menuRef}
            className={`
                fixed z-[${100 + level}] bg-[#1B1B1B] border border-[#666] 
                shadow-[4px_4px_8px_rgba(0,0,0,0.5)] min-w-[200px] py-1 select-none font-sans text-[11px]
            `}
            style={{
                top: y,
                left: x,
            }}
            onMouseLeave={() => setHoveredIndex(null)}
        >
            {actions.map((action, index) => (
                <div key={index} className="relative group">
                    {action.separator ? (
                        <div className="h-[1px] bg-[#444] my-1 mx-1" />
                    ) : (
                        <div
                            className={`
                                px-3 py-1 cursor-default flex items-center justify-between
                                ${action.danger ? 'text-white' : 'text-white'}
                                ${hoveredIndex === index ? 'bg-[#3399FF] text-white' : ''}
                            `}
                            onMouseEnter={() => setHoveredIndex(index)}
                            onClick={(e) => {
                                e.stopPropagation();
                                if (!action.hasSubmenu && action.onClick) {
                                    action.onClick();
                                    if (level === 0) onClose(); // Only close root triggers full close
                                }
                            }}
                        >
                            <div className="flex items-center gap-2">
                                {action.checked !== undefined && (
                                    <div className="w-3 flex items-center justify-center">
                                        {action.checked && <Check size={10} strokeWidth={4} />}
                                    </div>
                                )}
                                <span className={action.checked !== undefined ? "" : "pl-1"}>{action.label}</span>
                            </div>

                            <div className="flex items-center gap-2">
                                {action.shortcut && <span className="text-[#CCC] text-[9px]">{action.shortcut}</span>}
                                {(action.hasSubmenu || action.submenu) && <ChevronRight size={10} className="text-[#CCC]" />}
                            </div>
                        </div>
                    )}

                    {/* Recursive Submenu Step */}
                    {(action.hasSubmenu || action.submenu) && hoveredIndex === index && action.submenu && (
                        <div className="absolute left-full top-0 -ml-1">
                            <ContextMenu
                                x={0} // Relative positioning via CSS handles this if we make container relative
                                y={0}
                                onClose={onClose}
                                actions={action.submenu}
                                level={level + 1}
                            />
                        </div>
                    )}
                </div>
            ))}
        </div>
    );
}
