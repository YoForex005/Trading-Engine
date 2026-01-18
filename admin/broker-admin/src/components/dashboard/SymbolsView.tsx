import { useState, useEffect } from 'react';
import { Search, Power, Activity } from 'lucide-react';

interface SymbolSpec {
    symbol: string;
    contractSize: number;
    pipSize: number;
    disabled: boolean;
}

export default function SymbolsView() {

    const [symbols, setSymbols] = useState<SymbolSpec[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');

    const fetchSymbols = async () => {
        try {
            const res = await fetch('http://localhost:8080/admin/symbols');
            const data = await res.json();
            // Sort: Enabled first, then alpha
            data.sort((a: SymbolSpec, b: SymbolSpec) => {
                if (a.disabled === b.disabled) return a.symbol.localeCompare(b.symbol);
                return a.disabled ? 1 : -1;
            });
            setSymbols(data);
        } catch (err) {
            console.error('Failed to fetch symbols', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSymbols();
    }, []);

    const toggleSymbol = async (symbol: string, currentStatus: boolean) => {
        try {
            // Optimistic update
            setSymbols(prev => prev.map(s =>
                s.symbol === symbol ? { ...s, disabled: !currentStatus } : s
            ));

            const res = await fetch('http://localhost:8080/admin/symbols/toggle', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ symbol, disabled: !currentStatus })
            });

            if (!res.ok) throw new Error('Failed to toggle');

            // confirm from server
            fetchSymbols();
        } catch (err) {
            console.error('Error toggling symbol', err);
            // Revert on error
            fetchSymbols();
        }
    };

    const filteredSymbols = symbols.filter(s =>
        s.symbol.toLowerCase().includes(searchTerm.toLowerCase())
    );

    if (loading) return <div className="text-gray-400 p-4">Loading market data...</div>;

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-bold text-gray-100 flex items-center gap-2">
                    <Activity className="w-5 h-5 text-blue-400" />
                    Market Symbols
                </h2>
                <div className="relative">
                    <Search className="w-4 h-4 absolute left-3 top-3 text-gray-500" />
                    <input
                        type="text"
                        placeholder="Search symbol..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="bg-gray-800 text-gray-200 pl-10 pr-4 py-2 rounded-lg border border-gray-700 focus:border-blue-500 outline-none w-64"
                    />
                </div>
            </div>

            <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
                <table className="w-full text-left">
                    <thead className="bg-gray-900/50 text-gray-400 text-sm">
                        <tr>
                            <th className="p-4">Symbol</th>
                            <th className="p-4">Contract Size</th>
                            <th className="p-4">Pip Size</th>
                            <th className="p-4">Status</th>
                            <th className="p-4 text-right">Action</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-700">
                        {filteredSymbols.map((sym) => (
                            <tr key={sym.symbol} className="hover:bg-gray-700/30">
                                <td className="p-4 font-medium text-gray-200">{sym.symbol}</td>
                                <td className="p-4 text-gray-400">{sym.contractSize.toLocaleString()}</td>
                                <td className="p-4 text-gray-400">{sym.pipSize}</td>
                                <td className="p-4">
                                    <span className={`px-2 py-1 rounded text-xs font-medium ${!sym.disabled
                                        ? 'bg-green-500/20 text-green-400 border border-green-500/30'
                                        : 'bg-red-500/20 text-red-400 border border-red-500/30'
                                        }`}>
                                        {!sym.disabled ? 'ACTIVE' : 'DISABLED'}
                                    </span>
                                </td>
                                <td className="p-4 text-right">
                                    <button
                                        onClick={() => toggleSymbol(sym.symbol, sym.disabled)}
                                        className={`p-2 rounded-lg transition-colors ${!sym.disabled
                                            ? 'bg-red-500/10 text-red-400 hover:bg-red-500/20'
                                            : 'bg-green-500/10 text-green-400 hover:bg-green-500/20'
                                            }`}
                                        title={!sym.disabled ? "Disable Feed" : "Enable Feed"}
                                    >
                                        <Power className="w-4 h-4" />
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>

                {filteredSymbols.length === 0 && (
                    <div className="p-8 text-center text-gray-500">
                        No symbols found matching "{searchTerm}"
                    </div>
                )}
            </div>

            <div className="text-xs text-gray-500 mt-2">
                * Disabling a symbol stops price updates for all clients. Existing positions remain open but P/L will freeze.
            </div>
        </div>
    );
}
