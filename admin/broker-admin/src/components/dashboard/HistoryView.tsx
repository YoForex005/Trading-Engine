import React, { useState } from 'react';
import ContextMenu from '../ui/ContextMenu';

// Mock Data for History
const MOCK_HISTORY: any[] = Array.from({ length: 50 }).map((_, i) => ({
    ticket: (780000 + i).toString(),
    openTime: '2026.01.18 10:00:00',
    type: i % 2 === 0 ? 'buy' : 'sell',
    size: (Math.random() * 5).toFixed(2),
    symbol: i % 3 === 0 ? 'GBPUSD' : 'USDCAD',
    openPrice: (1.2000 + Math.random() * 0.05).toFixed(5),
    closeTime: '2026.01.19 09:00:00',
    closePrice: (1.2000 + Math.random() * 0.05).toFixed(5),
    commission: '-2.50',
    swap: '-1.00',
    profit: (Math.random() * 500 - 100).toFixed(2),
    comment: 'TP Hit',
}));

export default function HistoryView() {
    const [history] = useState(MOCK_HISTORY);
    const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; item: any } | null>(null);

    // Columns Configuration (History Specific)
    const COLUMNS = [
        { key: 'ticket', label: 'Ticket', w: 'w-20', align: 'left' },
        { key: 'openTime', label: 'Open Time', w: 'w-32', align: 'left' },
        { key: 'type', label: 'Type', w: 'w-12', align: 'left' },
        { key: 'size', label: 'Volume', w: 'w-16', align: 'right', mono: true },
        { key: 'symbol', label: 'Symbol', w: 'w-20', align: 'left' },
        { key: 'openPrice', label: 'Open', w: 'w-20', align: 'right', mono: true },
        { key: 'closeTime', label: 'Close Time', w: 'w-32', align: 'left' },
        { key: 'closePrice', label: 'Close', w: 'w-20', align: 'right', mono: true },
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

    const handleRightClick = (e: React.MouseEvent, item: any) => {
        e.preventDefault();
        if (!selectedIds.has(item.ticket)) {
            setSelectedIds(new Set([item.ticket]));
        }
        setContextMenu({ x: e.clientX, y: e.clientY, item });
    };

    const getMenuActions = (item: any) => [
        { label: 'Filter by Symbol', onClick: () => console.log('Filter Symbol') },
        { label: 'Report', onClick: () => console.log('Report') },
        { label: 'Export to Excel', onClick: () => console.log('Excel') },
        { label: 'Copy', onClick: () => console.log('Copy'), shortcut: 'Ctrl+C' },
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
                        {history.map((item, i) => {
                            const isSelected = selectedIds.has(item.ticket);
                            const profitVal = parseFloat(item.profit);

                            let profitColor = 'text-[#F0F0F0]';
                            if (profitVal < 0) profitColor = 'text-rtx-loss';
                            if (profitVal > 0) profitColor = 'text-rtx-profit';

                            let typeColor = item.type === 'buy' ? 'text-[#2ECC71]' : 'text-[#E74C3C]';

                            return (
                                <tr
                                    key={item.ticket}
                                    onClick={(e) => handleRowClick(e, item.ticket)}
                                    onContextMenu={(e) => handleRightClick(e, item)}
                                    className={`
                                        h-5 leading-5 border-b border-[#2A2C33]
                                        ${isSelected ? 'bg-[#2B3D55] text-white' : (i % 2 === 0 ? 'bg-transparent' : 'bg-[#16171A]')}
                                        group
                                    `}
                                >
                                    {COLUMNS.map(col => {
                                        const content = item[col.key];
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
                <span className="w-32">{history.length} Unified Orders</span>
                <span className="w-32">{selectedIds.size} Selected</span>
                <div className="flex-1" />
                <span className="font-mono text-[#AAA]">Server: 12ms</span>
            </div>

            {contextMenu && (
                <ContextMenu
                    x={contextMenu.x}
                    y={contextMenu.y}
                    onClose={() => setContextMenu(null)}
                    actions={getMenuActions(contextMenu.item)}
                />
            )}
        </div>
    );
}
