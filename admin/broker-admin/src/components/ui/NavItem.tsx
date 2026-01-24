import React from 'react';

interface NavItemProps {
    icon: React.ReactNode;
    label: string;
    active: boolean;
    onClick: () => void;
    collapsed?: boolean;
}

export default function NavItem({ icon, label, active, onClick, collapsed }: NavItemProps) {
    return (
        <button
            onClick={onClick}
            title={collapsed ? label : undefined}
            className={`
                group relative flex items-center w-full transition-all duration-200
                ${collapsed ? 'justify-center px-0 py-3' : 'px-4 py-3 gap-3'}
                ${active
                    ? 'text-rtx-yellow bg-charcoal-800'
                    : 'text-zinc-500 hover:text-zinc-300 hover:bg-charcoal-800/50'
                }
            `}
        >
            {/* Active Indicator Line */}
            {active && (
                <div className="absolute left-0 top-0 bottom-0 w-1 bg-rtx-yellow shadow-[0_0_10px_rgba(245,200,66,0.3)]"></div>
            )}

            {/* Icon */}
            <div className={`
                transition-transform duration-200 
                ${active ? 'scale-110' : 'group-hover:scale-105'}
            `}>
                {icon}
            </div>

            {/* Label */}
            {!collapsed && (
                <span className="font-medium text-sm tracking-wide transition-opacity duration-200">
                    {label}
                </span>
            )}

            {/* Hover Tooltip (only when collapsed) */}
            {collapsed && (
                <div className="absolute left-full ml-2 px-2 py-1 bg-charcoal-800 border border-charcoal-600 text-xs text-white rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50 shadow-lg">
                    {label}
                </div>
            )}
        </button>
    );
}
