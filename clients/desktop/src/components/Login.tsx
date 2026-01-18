import React, { useState } from 'react';
import { Terminal, Lock, Server, ArrowRight, Activity } from 'lucide-react';
import { useAppStore } from '../store/useAppStore';

interface LoginProps {
    onLogin: (accountId: string) => void;
}

export function Login({ onLogin }: LoginProps) {
    const [loading, setLoading] = useState(false);
    const [server, setServer] = useState('localhost:8080');
    const setAuthenticated = useAppStore(state => state.setAuthenticated);
    const setAuthToken = useAppStore(state => state.setAuthToken);

    const handleConnect = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            const form = e.target as HTMLFormElement;
            const username = (form.elements[0] as HTMLInputElement).value;
            const password = (form.elements[1] as HTMLInputElement).value;

            // Assume username IS the account ID for now
            const accountId = username;

            // Note: Dynamic server selection is used. In production, use proper environment handling
            if (server !== 'localhost:8080') {
                console.log('Using server:', server);
            }

            // Make login request with proper error handling
            const res = await fetch(`http://${server}/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });

            if (res.status !== 200) {
                throw new Error('Authentication failed');
            }

            const data = await res.json();

            if (!data.token) {
                throw new Error('No authentication token received');
            }

            // Store token in Zustand store and localStorage for persistence
            setAuthToken(data.token);
            setAuthenticated(true, accountId, data.token);
            localStorage.setItem('rtx_token', data.token);
            localStorage.setItem('rtx_user', JSON.stringify(data.user));

            onLogin(accountId);
        } catch (err: any) {
            console.error('Login error:', err);
            const msg = err.message || "Connection Failed";
            alert(`Login Error: ${msg}\n\nCheck server is running at ${server}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex h-screen w-full items-center justify-center bg-black text-white relative overflow-hidden">
            {/* Background Decor */}
            <div className="absolute top-0 left-0 w-full h-full bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-emerald-900/20 via-black to-black pointer-events-none" />

            <div className="z-10 w-full max-w-md p-8 bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-2xl shadow-2xl animate-in fade-in zoom-in duration-300">
                <div className="flex flex-col items-center mb-8">
                    <div className="w-12 h-12 bg-emerald-500/10 rounded-xl flex items-center justify-center mb-4 border border-emerald-500/20">
                        <Activity className="w-6 h-6 text-emerald-500" />
                    </div>
                    <h1 className="text-2xl font-bold tracking-tight text-white">RTX Terminal</h1>
                    <p className="text-zinc-400 text-sm mt-1">Institutional Trading Environment</p>
                </div>

                <form onSubmit={handleConnect} className="space-y-4">
                    <div className="space-y-2">
                        <label className="text-xs font-medium text-zinc-500 uppercase tracking-wider">Credentials</label>
                        <div className="relative group">
                            <input
                                type="text"
                                placeholder="Account ID (e.g., 1)"
                                className="w-full bg-zinc-950/50 border border-zinc-800 rounded-lg py-2.5 px-3 pl-10 text-sm text-zinc-200 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/50 transition-all placeholder:text-zinc-600"
                                defaultValue="1"
                            />
                            <Terminal className="w-4 h-4 text-zinc-500 absolute left-3 top-3 group-focus-within:text-emerald-500 transition-colors" />
                        </div>
                        <div className="relative group">
                            <input
                                type="password"
                                placeholder="Password"
                                className="w-full bg-zinc-950/50 border border-zinc-800 rounded-lg py-2.5 px-3 pl-10 text-sm text-zinc-200 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/50 transition-all placeholder:text-zinc-600"
                                defaultValue="password"
                            />
                            <Lock className="w-4 h-4 text-zinc-500 absolute left-3 top-3 group-focus-within:text-emerald-500 transition-colors" />
                        </div>
                    </div>

                    <div className="space-y-2 pt-2">
                        <label className="text-xs font-medium text-zinc-500 uppercase tracking-wider flex items-center gap-2">
                            Server Connection
                        </label>
                        <div className="relative group">
                            <input
                                type="text"
                                value={server}
                                onChange={(e) => setServer(e.target.value)}
                                className="w-full bg-zinc-950/30 border border-zinc-800/50 rounded-lg py-2 px-3 pl-10 text-xs text-zinc-400 focus:outline-none focus:border-emerald-500/30 transition-all font-mono"
                            />
                            <Server className="w-3.5 h-3.5 text-zinc-600 absolute left-3 top-2.5" />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={loading}
                        className="w-full mt-6 bg-emerald-600 hover:bg-emerald-500 text-white font-medium py-2.5 rounded-lg transition-all flex items-center justify-center gap-2 shadow-lg shadow-emerald-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {loading ? (
                            <span className="animate-pulse">Connecting...</span>
                        ) : (
                            <>
                                Connect to Gate
                                <ArrowRight className="w-4 h-4" />
                            </>
                        )}
                    </button>
                </form>

                <div className="mt-6 flex items-center justify-between px-2">
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></div>
                        <span className="text-[10px] text-zinc-500">System Operational</span>
                    </div>
                    <span className="text-[10px] text-zinc-600">v1.0.0-alpha</span>
                </div>
            </div>
        </div>
    );
}
