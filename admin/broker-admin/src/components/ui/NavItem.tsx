import React from 'react';

interface NavItemProps {
    icon: React.ReactNode;
    label: string;
    active: boolean;
    onClick: () => void;
}

export default function NavItem({ icon, label, active, onClick }: NavItemProps) {
    return (
        <button
            onClick={onClick}
            className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${active ? 'bg-emerald-600/20 text-emerald-400' : 'hover:bg-zinc-800 text-zinc-400'
                }`}
        >
            {icon}
            {label}
        </button>
    );
}
