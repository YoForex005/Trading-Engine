import React from 'react';
import { LedgerEntry } from '../../types';

export default function LedgerView({ ledger }: { ledger: LedgerEntry[] }) {
    return (
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl overflow-hidden">
            <div className="p-4 border-b border-zinc-800">
                <h3 className="font-semibold">Transaction Ledger</h3>
            </div>
            <table className="w-full">
                <thead className="bg-zinc-800/50">
                    <tr className="text-left text-xs text-zinc-500 uppercase">
                        <th className="p-3">ID</th>
                        <th className="p-3">Account</th>
                        <th className="p-3">Type</th>
                        <th className="p-3">Amount</th>
                        <th className="p-3">Balance After</th>
                        <th className="p-3">Method</th>
                        <th className="p-3">Reference</th>
                        <th className="p-3">Date</th>
                    </tr>
                </thead>
                <tbody>
                    {(ledger || []).map(entry => (
                        <tr key={entry.id} className="border-t border-zinc-800 hover:bg-zinc-800/30">
                            <td className="p-3 font-mono text-xs">{entry.id}</td>
                            <td className="p-3">#{entry.accountId}</td>
                            <td className="p-3">
                                <span className={`px-2 py-0.5 rounded text-xs ${entry.type === 'DEPOSIT' ? 'bg-emerald-500/20 text-emerald-400' :
                                    entry.type === 'WITHDRAW' ? 'bg-red-500/20 text-red-400' :
                                        entry.type === 'REALIZED_PNL' ? 'bg-blue-500/20 text-blue-400' :
                                            'bg-zinc-700 text-zinc-400'
                                    }`}>
                                    {entry.type}
                                </span>
                            </td>
                            <td className={`p-3 font-mono ${entry.amount >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                                {entry.amount >= 0 ? '+' : ''}{entry.amount.toFixed(2)}
                            </td>
                            <td className="p-3 font-mono">${entry.balanceAfter.toFixed(2)}</td>
                            <td className="p-3 text-xs text-zinc-400">{entry.paymentMethod || '-'}</td>
                            <td className="p-3 text-xs text-zinc-500 max-w-32 truncate" title={entry.paymentRef}>{entry.paymentRef || '-'}</td>
                            <td className="p-3 text-xs text-zinc-500">{new Date(entry.createdAt).toLocaleString()}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}
