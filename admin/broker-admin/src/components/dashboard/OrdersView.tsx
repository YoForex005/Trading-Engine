import React, { useState } from 'react';
import ContextMenu from '../ui/ContextMenu';

// Mock Data for Orders
const MOCK_ORDERS: any[] = Array.from({ length: 50 }).map((_, i) => ({
    ticket: (880000 + i).toString(),
    time: '2026.01.19 14:30:00',
    type: i % 2 === 0 ? 'buy' : 'sell',
    size: (Math.random() * 5).toFixed(2),
    symbol: i % 3 === 0 ? 'EURUSD' : 'XAUUSD',
    openPrice: (2000 + Math.random() * 50).toFixed(5),
    sl: '0.00000',
    tp: '0.00000',
    price: (2000 + Math.random() * 50).toFixed(5),
    commission: '-2.50',
    swap: i % 5 === 0 ? '-1.20' : '0.00',
    profit: (Math.random() * 500 - 100).toFixed(2),
    comment: 'Expert Advisor',
}));

interface OrdersViewProps {
    onOrderDoubleClick?: (order: any) => void;
}

export default function OrdersView({ onOrderDoubleClick }: OrdersViewProps) {
    const [orders] = useState(MOCK_ORDERS);
    const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; order: any } | null>(null);

    // Columns Configuration (Orders Specific)
    const COLUMNS = [
        { key: 'ticket', label: 'Ticket', w: 'w-20', align: 'left' },
        { key: 'time', label: 'Time', w: 'w-32', align: 'left' },
        { key: 'type', label: 'Type', w: 'w-12', align: 'left' },
        { key: 'size', label: 'Volume', w: 'w-16', align: 'right', mono: true },
        { key: 'symbol', label: 'Symbol', w: 'w-20', align: 'left' },
        { key: 'openPrice', label: 'Price', w: 'w-20', align: 'right', mono: true },
        { key: 'sl', label: 'S / L', w: 'w-20', align: 'right', mono: true },
        { key: 'tp', label: 'T / P', w: 'w-20', align: 'right', mono: true },
        { key: 'price', label: 'Price', w: 'w-20', align: 'right', mono: true },
        { key: 'commission', label: 'Comm', w: 'w-16', align: 'right', mono: true },
        { key: 'swap', label: 'Swap', w: 'w-16', align: 'right', mono: true },
        { key: 'profit', label: 'Profit', w: 'w-20', align: 'right', mono: true, colorLogic: true },
        { key: 'comment', label: 'Comment', w: 'w-32', align: 'left' },
    ];

    const handleRowClick = (e: React.MouseEvent, id: string) => {
        if (e.ctrlKey) {
            const newSet = new Set(selectedIds);
            if (newSet.has(id)) newSet.delete(id);
            else newSet.add(id);
            setSelectedIds(newSet);
        } else {
            setSelectedIds(new Set([id]));
        }
    };

    const handleRightClick = (e: React.MouseEvent, order: any) => {
        e.preventDefault();
        if (!selectedIds.has(order.ticket)) {
            setSelectedIds(new Set([order.ticket]));
        }
        setContextMenu({ x: e.clientX, y: e.clientY, order });
    };

    const getMenuActions = (order: any) => [
        { label: 'Close Order', onClick: () => console.log('Close') },
        { label: 'Modify Order', onClick: () => console.log('Modify') },
        { label: 'Delete Order', onClick: () => console.log('Delete') },
        { label: 'Bulk Operations', onClick: () => console.log('Bulk'), separator: true },
        { label: 'Copy', onClick: () => console.log('Copy'), shortcut: 'Ctrl+C' },
        { label: 'Report', onClick: () => console.log('Report'), separator: true },
    ];

    return (
        <div className="flex flex-col h-full bg-[#121316] select-none cursor-default font-sans text-[11px]">
            <div className="flex-1 overflow-auto custom-scrollbar relative">
                <table className="min-w-max w-full border-collapse table-fixed">
                    <thead className="sticky top-0 z-10 bg-[#1E2026] text-[#A0A0A0] shadow-[0_1px_0_0_#000]">
                        <tr className="h-5">
                            {COLUMNS.map(col => (
                                <th key={col.key} className={`${col.w} px-2 py-0 border-r border-[#383A42] text-${col.align} font-normal truncate`}>
                                    {col.label}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody className="bg-[#121316]">
                        {orders.map((ord, i) => {
                            const isSelected = selectedIds.has(ord.ticket);
                            const profitVal = parseFloat(ord.profit);

                            let profitColor = 'text-[#F0F0F0]';
                            if (profitVal < 0) profitColor = 'text-rtx-loss';
                            if (profitVal > 0) profitColor = 'text-rtx-profit';

                            let typeColor = ord.type === 'buy' ? 'text-[#2ECC71]' : 'text-[#E74C3C]';

                            return (
                                <tr
                                    key={ord.ticket}
                                    onClick={(e) => handleRowClick(e, ord.ticket)}
                                    onDoubleClick={() => onOrderDoubleClick && onOrderDoubleClick(ord)}
                                    onContextMenu={(e) => handleRightClick(e, ord)}
                                    className={`
                                        h-5 leading-5 border-b border-[#2A2C33]
                                        ${isSelected ? 'bg-[#2B3D55] text-white' : (i % 2 === 0 ? 'bg-transparent' : 'bg-[#16171A]')}
                                        group
                                    `}
                                >
                                    {COLUMNS.map(col => {
                                        const content = ord[col.key];
                                        let cellClass = `px-2 border-r border-[#383A42] truncate align-middle text-${col.align} ${col.mono ? 'font-mono' : ''}`;

                                        if (!isSelected) {
                                            if (col.key === 'profit') cellClass += ` ${profitColor}`;
                                            if (col.key === 'type') cellClass += ` ${typeColor}`;
                                        } else {
                                            cellClass += ' text-white';
                                        }

                                        return (
                                            <td key={col.key} className={cellClass}>
                                                {col.key === 'profit' || col.key === 'commission' || col.key === 'swap'
                                                    ? Number(content).toFixed(2) : content}
                                            </td>
                                        );
                                    })}
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>

            <div className="h-5 flex items-center bg-[#1E2026] border-t border-[#383A42] px-2 text-[#888] gap-4 mt-auto text-[10px]">
                <span className="w-32">{orders.length} Orders</span>
                <span className="w-32">{selectedIds.size} Selected</span>
                <div className="flex-1" />
                <span className="font-mono text-[#AAA]">Server: 12ms</span>
            </div>

            {contextMenu && (
                <ContextMenu
                    x={contextMenu.x}
                    y={contextMenu.y}
                    onClose={() => setContextMenu(null)}
                    actions={getMenuActions(contextMenu.order)}
                />
            )}
        </div>
    );
}
