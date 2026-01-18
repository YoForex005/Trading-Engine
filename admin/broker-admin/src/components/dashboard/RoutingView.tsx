import React from 'react';
import { RoutingRule } from '../../types';

export default function RoutingView({ routes }: { routes: RoutingRule[] }) {
    return (
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl overflow-hidden">
            <div className="p-4 border-b border-zinc-800 flex items-center justify-between">
                <h3 className="font-semibold">Order Routing Rules (A-Book vs B-Book)</h3>
                <button className="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 rounded-lg text-sm transition-colors">
                    + Add Rule
                </button>
            </div>
            <table className="w-full">
                <thead className="bg-zinc-800/50">
                    <tr className="text-left text-xs text-zinc-500 uppercase">
                        <th className="p-3">Group</th>
                        <th className="p-3">Symbol</th>
                        <th className="p-3">Volume Range</th>
                        <th className="p-3">Action</th>
                        <th className="p-3">Target LP</th>
                        <th className="p-3">Priority</th>
                    </tr>
                </thead>
                <tbody>
                    {(routes || []).length === 0 ? (
                        <tr>
                            <td colSpan={6} className="p-8 text-center text-zinc-500">
                                No routing rules. All orders will be B-Booked.
                            </td>
                        </tr>
                    ) : (
                        (routes || []).map(rule => (
                            <tr key={rule.id} className="border-t border-zinc-800 hover:bg-zinc-800/30">
                                <td className="p-3">{rule.groupPattern}</td>
                                <td className="p-3">{rule.symbolPattern}</td>
                                <td className="p-3">{rule.minVolume} - {rule.maxVolume}</td>
                                <td className="p-3">
                                    <span className={`px-2 py-0.5 rounded text-xs ${rule.action === 'A_BOOK' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-orange-500/20 text-orange-400'}`}>
                                        {rule.action}
                                    </span>
                                </td>
                                <td className="p-3 text-zinc-400">{rule.targetLp || '-'}</td>
                                <td className="p-3">{rule.priority}</td>
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
    );
}
