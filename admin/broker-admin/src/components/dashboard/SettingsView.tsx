'use client';

import React, { useState } from 'react';
import { AlertTriangle, Loader2, CheckCircle, XCircle, RefreshCw } from 'lucide-react';

interface SettingsViewProps {
    mode: string;
    onModeChange: (m: string) => void;
    lp: string;
    onLPChange: (l: string) => void;
}

export default function SettingsView({ mode, onModeChange, lp, onLPChange }: SettingsViewProps) {
    const [restartStatus, setRestartStatus] = useState<'idle' | 'restarting' | 'success' | 'error'>('idle');
    const [restartError, setRestartError] = useState<string>('');

    const handleRestart = async () => {
        if (!confirm('Are you sure you want to restart the backend? This will disconnect all clients temporarily.')) {
            return;
        }

        setRestartStatus('restarting');
        setRestartError('');

        try {
            // Send restart request
            const response = await fetch('http://localhost:8080/admin/restart', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' }
            });

            if (!response.ok) {
                throw new Error(`Server returned ${response.status}`);
            }

            // Wait for server to shut down
            await new Promise(resolve => setTimeout(resolve, 3000));

            // Poll for server to come back up
            let attempts = 0;
            const maxAttempts = 10;

            while (attempts < maxAttempts) {
                try {
                    const health = await fetch('http://localhost:8080/api/config', {
                        method: 'GET',
                        signal: AbortSignal.timeout(2000)
                    });

                    if (health.ok) {
                        setRestartStatus('success');
                        setTimeout(() => setRestartStatus('idle'), 5000);
                        return;
                    }
                } catch {
                    // Server not ready yet
                }

                await new Promise(resolve => setTimeout(resolve, 1000));
                attempts++;
            }

            // Server didn't come back
            setRestartStatus('error');
            setRestartError('Server did not restart automatically. Please start it manually with: go run ./cmd/server/main.go');
        } catch (e) {
            setRestartStatus('error');
            setRestartError(e instanceof Error ? e.message : 'Failed to restart backend');
        }
    };

    return (
        <div className="space-y-6">
            <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
                <h3 className="font-semibold mb-4">Execution Mode</h3>
                <div className="flex gap-4">
                    <label className={`flex items-center gap-2 p-4 border rounded-lg cursor-pointer transition-colors ${mode === 'BBOOK' ? 'bg-orange-500/10 border-orange-500/30 ring-1 ring-orange-500/50' : 'bg-zinc-800 border-zinc-700'}`}>
                        <input
                            type="radio"
                            name="mode"
                            value="BBOOK"
                            checked={mode === 'BBOOK'}
                            onChange={() => onModeChange('BBOOK')}
                            className="accent-orange-500"
                        />
                        <div>
                            <div className="font-medium text-orange-400">B-Book (Internal)</div>
                            <div className="text-xs text-zinc-500">All orders executed internally</div>
                        </div>
                    </label>
                    <label className={`flex items-center gap-2 p-4 border rounded-lg cursor-pointer transition-colors ${mode === 'ABOOK' ? 'bg-emerald-500/10 border-emerald-500/30 ring-1 ring-emerald-500/50' : 'bg-zinc-800 border-zinc-700'}`}>
                        <input
                            type="radio"
                            name="mode"
                            value="ABOOK"
                            checked={mode === 'ABOOK'}
                            onChange={() => onModeChange('ABOOK')}
                            className="accent-emerald-500"
                        />
                        <div>
                            <div className="font-medium text-emerald-400">A-Book (LP)</div>
                            <div className="text-xs text-zinc-500">Orders routed to liquidity providers</div>
                        </div>
                    </label>
                </div>
            </div>

            <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
                <h3 className="font-semibold mb-4">Price Feed Source</h3>
                <div className="max-w-md">
                    <label className="text-sm text-zinc-400 mb-2 block">Primary Liquidity Provider</label>
                    <select
                        value={lp}
                        onChange={(e) => onLPChange(e.target.value)}
                        className="w-full px-4 py-3 bg-zinc-800 border border-zinc-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    >
                        <option value="OANDA">OANDA (REST API)</option>
                        <option value="FLEXYMARKETS">FlexyMarkets (WebSocket - Crypto)</option>
                        <option value="LMAX_DEMO">LMAX Demo (FIX 4.4)</option>
                        <option value="LMAX_PROD">LMAX Production (FIX 4.4)</option>
                        <option value="CURRENEX">Currenex (FIX 4.4)</option>
                        <option value="BINANCE">Binance (Crypto Only)</option>
                    </select>

                    <p className="mt-2 text-xs text-zinc-500">
                        <AlertTriangle className="inline w-3 h-3 mr-1 text-yellow-500" />
                        Changing this requires a backend restart to re-establish the connection.
                    </p>
                </div>
            </div>

            <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
                <h3 className="font-semibold mb-4">Default Account Settings</h3>
                <div className="grid grid-cols-2 gap-4">
                    <div>
                        <label className="text-sm text-zinc-400 mb-1 block">Default Leverage</label>
                        <select className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg">
                            <option>1:100</option>
                            <option>1:200</option>
                            <option>1:500</option>
                        </select>
                    </div>
                </div>
            </div>

            <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
                <h3 className="font-semibold mb-4 text-red-400">System Actions</h3>
                <div className="flex flex-col gap-4">
                    <div>
                        <p className="text-sm text-zinc-400 mb-3">
                            After changing LP settings, restart the backend to apply changes.
                        </p>

                        <button
                            onClick={handleRestart}
                            disabled={restartStatus === 'restarting'}
                            className={`px-4 py-2 border rounded-lg transition-colors flex items-center gap-2 ${restartStatus === 'restarting'
                                    ? 'bg-yellow-500/20 border-yellow-500/30 text-yellow-400 cursor-wait'
                                    : restartStatus === 'success'
                                        ? 'bg-green-500/20 border-green-500/30 text-green-400'
                                        : restartStatus === 'error'
                                            ? 'bg-red-500/20 border-red-500/30 text-red-400'
                                            : 'bg-red-500/20 border-red-500/30 text-red-400 hover:bg-red-500/30'
                                }`}
                        >
                            {restartStatus === 'restarting' && (
                                <>
                                    <Loader2 className="w-4 h-4 animate-spin" />
                                    Restarting...
                                </>
                            )}
                            {restartStatus === 'success' && (
                                <>
                                    <CheckCircle className="w-4 h-4" />
                                    Restart Complete!
                                </>
                            )}
                            {restartStatus === 'error' && (
                                <>
                                    <XCircle className="w-4 h-4" />
                                    Restart Failed
                                </>
                            )}
                            {restartStatus === 'idle' && (
                                <>
                                    <RefreshCw className="w-4 h-4" />
                                    Restart Backend
                                </>
                            )}
                        </button>

                        {restartStatus === 'error' && restartError && (
                            <div className="mt-3 p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
                                <p className="text-sm text-red-400 font-medium">Error Details:</p>
                                <p className="text-xs text-red-300 mt-1 font-mono">{restartError}</p>
                            </div>
                        )}

                        {restartStatus === 'restarting' && (
                            <p className="text-xs text-yellow-400 mt-2">
                                Waiting for backend to restart... This may take up to 15 seconds.
                            </p>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
