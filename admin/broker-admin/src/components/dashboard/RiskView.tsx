import React from 'react';
import { Account } from '../../types';

export default function RiskView({ accounts }: { accounts: Account[] }) {
    const totalBalance = (accounts || []).reduce((sum, a) => sum + a.balance, 0);

    return (
        <div className="space-y-6">
            <div className="grid grid-cols-3 gap-4">
                <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
                    <h4 className="text-sm text-zinc-500 mb-2">Total Accounts</h4>
                    <p className="text-2xl font-bold">{(accounts || []).length}</p>
                </div>
                <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
                    <h4 className="text-sm text-zinc-500 mb-2">Total Balance</h4>
                    <p className="text-2xl font-bold text-emerald-400">${totalBalance.toLocaleString()}</p>
                </div>
                <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
                    <h4 className="text-sm text-zinc-500 mb-2">Execution Mode</h4>
                    <p className="text-2xl font-bold text-orange-400">B-Book</p>
                </div>
            </div>
        </div>
    );
}
