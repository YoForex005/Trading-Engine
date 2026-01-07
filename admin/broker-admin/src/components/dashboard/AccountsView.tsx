import React, { useState } from 'react';
import { Search, Plus, ArrowDownCircle, ArrowUpCircle, Edit2, Lock } from 'lucide-react';
import { Account } from '../../types';

interface AccountsViewProps {
    accounts: Account[];
    onSelect: (a: Account) => void;
    onAction: (a: string) => void;
    selected: Account | null;
    onRefresh: () => void;
}

export default function AccountsView({ accounts, onSelect, onAction, selected, onRefresh }: AccountsViewProps) {
    const [search, setSearch] = useState('');

    const filtered = accounts.filter(a =>
        a.accountNumber?.toLowerCase().includes(search.toLowerCase()) ||
        a.userId?.toLowerCase().includes(search.toLowerCase())
    );

    const handleCreateAccount = () => {
        // Open modal instead of direct fetch
        onAction('create');
        onSelect({ id: 0, accountNumber: 'NEW', balance: 0, leverage: 100 } as any); // Dummy account for modal
    };

    return (
        <div className="space-y-4">
            {/* Toolbar */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <div className="relative">
                        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
                        <input
                            type="text"
                            placeholder="Search accounts..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            className="pl-9 pr-4 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm w-64"
                        />
                    </div>
                </div>
                <button
                    onClick={handleCreateAccount}
                    className="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 rounded-lg text-sm transition-colors"
                >
                    <Plus size={16} />
                    Create Account
                </button>
            </div>

            {/* Accounts Table */}
            <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl overflow-hidden">
                <table className="w-full">
                    <thead className="bg-zinc-800/50">
                        <tr className="text-left text-xs text-zinc-500 uppercase">
                            <th className="p-3">Account</th>
                            <th className="p-3">Balance</th>
                            <th className="p-3">Leverage</th>
                            <th className="p-3">Mode</th>
                            <th className="p-3">Status</th>
                            <th className="p-3">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {(filtered || []).map(account => (
                            <tr
                                key={account.id}
                                className={`border-t border-zinc-800 hover:bg-zinc-800/30 cursor-pointer ${selected?.id === account.id ? 'bg-zinc-800/50' : ''}`}
                                onClick={() => onSelect(account)}
                            >
                                <td className="p-3">
                                    <div>
                                        <div className="font-medium">{account.accountNumber}</div>
                                        <div className="text-xs text-zinc-500">{account.userId}</div>
                                    </div>
                                </td>
                                <td className="p-3">
                                    <span className="font-mono text-emerald-400">${account.balance.toLocaleString()}</span>
                                </td>
                                <td className="p-3">1:{account.leverage}</td>
                                <td className="p-3">
                                    <span className={`px-2 py-0.5 rounded text-xs ${account.marginMode === 'HEDGING' ? 'bg-blue-500/20 text-blue-400' : 'bg-purple-500/20 text-purple-400'}`}>
                                        {account.marginMode}
                                    </span>
                                </td>
                                <td className="p-3">
                                    <span className={`px-2 py-0.5 rounded text-xs ${account.status === 'ACTIVE' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-500/20 text-red-400'}`}>
                                        {account.status}
                                    </span>
                                </td>
                                <td className="p-3">
                                    <div className="flex items-center gap-1">
                                        <button
                                            onClick={(e) => { e.stopPropagation(); onSelect(account); onAction('deposit'); }}
                                            className="p-1.5 hover:bg-emerald-500/20 rounded text-emerald-400" title="Deposit"
                                        >
                                            <ArrowDownCircle size={16} />
                                        </button>
                                        <button
                                            onClick={(e) => { e.stopPropagation(); onSelect(account); onAction('withdraw'); }}
                                            className="p-1.5 hover:bg-red-500/20 rounded text-red-400" title="Withdraw"
                                        >
                                            <ArrowUpCircle size={16} />
                                        </button>
                                        <button
                                            onClick={(e) => { e.stopPropagation(); onSelect(account); onAction('edit'); }}
                                            className="p-1.5 hover:bg-zinc-700 rounded text-zinc-400" title="Edit"
                                        >
                                            <Edit2 size={16} />
                                        </button>
                                        <button
                                            onClick={(e) => { e.stopPropagation(); onSelect(account); onAction('reset-password'); }}
                                            className="p-1.5 hover:bg-yellow-500/20 rounded text-yellow-400" title="Reset Password"
                                        >
                                            <Lock size={16} />
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
